import { readFileSync, writeFileSync } from 'node:fs';
import * as annual from '../../indicators/lars-von-thienen/annual-cycle-model/model.js';
import * as superposition from '../../indicators/lars-von-thienen/harmonic-energy-superposition-wave/model.js';
import { ScanConfig, ScanOutput, RankedCycle, scanCycles, emptyScan, modeledValue } from '../../indicators/lars-von-thienen/cycle-scanner/scanner.js';
import { phaseState } from '../../indicators/lars-von-thienen/harmonic-cycle/model.js';
import { centeredMean } from '../../indicators/lars-von-thienen/dynamic-cycle-explorer/engine.js';
const workflows = ['complete-cycle-scanner', 'dominant-market-vibration', 'dynamic-cycle-explorer',
    'cycle-phase-score', 'phasing-oscillator', 'annual-cycle-pattern-decoder', 'harmonic-energy-superposition-wave'];
type Row = Record<string, string | number | null>;
const booleanOptions = new Set(['rolling', 'invert', 'confirm-complete-year', 'confirm-complete-years']);
const defaults: Record<string, string | boolean> = { window: '200', 'min-period': '5', 'max-period': '80',
    smoothing: '1600', 'genuine-threshold': '49', 'phase-cycles': '2', 'top-k': '3', 'peak-distance': '2',
    'max-segments': '5', 'candidate-pool': '10', 'max-components': '7', horizon: '0', 'short-min': '5',
    'short-max': '30', 'long-min': '31', 'long-max': '80', normalization: 'rms', 'calendar-policy': 'extend',
    'as-of': new Date().toISOString().slice(0, 10), rolling: false, invert: false,
    'confirm-complete-year': false, 'confirm-complete-years': false };
const valueOptions = new Set([...Object.keys(defaults), 'input', 'calendar', 'model', 'models', 'history',
    'json-out', 'model-out', 'wave-out', 'series-out', 'start', 'end', 'year', 'target-year', 'radians']);
/** Parse quoted CSV fields, including CRLF and escaped double quotes. */
function readCsv(path: string): string[][] {
    const input = readFileSync(path, 'utf8');
    const rows: string[][] = [];
    let row: string[] = [];
    let field = '';
    let quoted = false;
    for (let i = 0; i < input.length; i++) {
        const c = input[i];
        if (c === '"') {
            if (quoted && input[i + 1] === '"') {
                field += '"';
                i++;
            }
            else
                quoted = !quoted;
        }
        else if (!quoted && (c === ',' || c === '\n')) {
            row.push(field.replace(/\r$/, ''));
            field = '';
            if (c === '\n') {
                if (row.some(value => value.length > 0))
                    rows.push(row);
                row = [];
            }
        }
        else
            field += c;
    }
    if (quoted)
        throw new Error('unterminated CSV quote');
    if (field || row.length) {
        row.push(field.replace(/\r$/, ''));
        rows.push(row);
    }
    if (rows.length < 2)
        throw new Error('CSV requires a header and data rows');
    if (rows.some(r => r.length !== rows[0].length))
        throw new Error('inconsistent CSV column counts');
    return rows;
}
function readPrices(path: string): annual.PriceRecord[] {
    const rows = readCsv(path);
    const d = rows[0].indexOf('date');
    const c = rows[0].indexOf('close');
    if (d < 0 || c < 0)
        throw new Error('CSV requires date,close columns');
    const records = rows.slice(1).map(row => ({ date: row[d], close: row[c].trim() ? Number(row[c]) : NaN }));
    annual.validateDates(records.map(r => r.date));
    if (records.some(r => !Number.isFinite(r.close)))
        throw new Error('CSV closes must be finite');
    return records;
}
function readCalendar(path: string): string[] {
    const rows = readCsv(path);
    const i = rows[0].indexOf('date');
    if (i < 0)
        throw new Error('calendar requires date column');
    const dates = rows.slice(1).map(r => r[i]);
    annual.validateDates(dates);
    return dates;
}
function csvText(value: unknown): string {
    const text = value === null || value === undefined || typeof value === 'number' && !Number.isFinite(value)
        ? '' : typeof value === 'number' ? value.toPrecision(15) : String(value);
    return /[",\r\n]/.test(text) ? '"' + text.replace(/"/g, '""') + '"' : text;
}
function saveWave(path: string, dates: readonly string[], values: readonly number[]): void {
    if (dates.length !== values.length)
        throw new Error('wave and calendar lengths differ');
    writeFileSync(path, 'date,value\n' + values.map((v, i) => `${csvText(dates[i])},${csvText(v)}`).join('\n') + '\n');
}
function saveSeries(path: string, series: Row[]): void {
    const keys = ['date', ...Object.keys(series[0]).filter(k => k !== 'date').sort()];
    writeFileSync(path, keys.map(csvText).join(',') + '\n' + series.map(r => keys.map(k => csvText(r[k])).join(',')).join('\n') + '\n');
}
function snake(name: string): string { return name.replace(/[A-Z]/g, c => '_' + c.toLowerCase()); }
function record(value: object): Row { return Object.fromEntries(Object.entries(value).map(([k, v]) => [snake(k), v])) as Row; }
function phaseStatus(phase: number): string {
    if (Number.isNaN(phase))
        return 'unavailable';
    const angle = ((phase % (2 * Math.PI)) + 2 * Math.PI) % (2 * Math.PI);
    return angle < Math.PI / 2 ? 'rising toward top' : angle < Math.PI ? 'falling from top' :
        angle < 3 * Math.PI / 2 ? 'falling toward bottom' : 'rising from bottom';
}
function flattenScan(scan: ScanOutput, config: ScanConfig): Row {
    const output: Row = { accepted_count: scan.acceptedCount };
    for (let i = 0; i < scan.amplitude.length; i++) {
        output['amplitude_p' + (config.minPeriod + i)] = scan.amplitude[i];
        output['genuine_p' + (config.minPeriod + i)] = scan.genuinePercent[i];
    }
    scan.cycles.forEach((c, i) => Object.entries(record(c)).forEach(([k, v]) => { output[`rank${i + 1}_${k}`] = v; }));
    return output;
}
function scannerReport(scan: ScanOutput, config: ScanConfig, dates: string[]): Record<string, unknown> {
    return { schema: 'lvt-cycle-scan', version: 1, research_reconstruction: true,
        start_date: dates[0], end_date: dates[dates.length - 1], sample_count: dates.length, config: record(config),
        phase_convention: 'sine radians at analysis endpoint', validation_input: 'raw prices, most recent complete segments',
        accepted_count: scan.acceptedCount, cycles: scan.cycles.map(record), spectrum: {
            periods: Array.from({ length: config.maxPeriod - config.minPeriod + 1 }, (_, i) => i + config.minPeriod),
            amplitude: scan.amplitude, genuine_percent: scan.genuinePercent
        } };
}
export function run(arguments_: string[]): void {
    if (arguments_.length < 2 || arguments_.includes('--help')) {
        console.log('usage: lvt <workflow> <analyze|fit|build|replay|phase> [options]\n' + workflows.join('\n'));
        return;
    }
    const [slug, command] = arguments_;
    if (!workflows.includes(slug))
        throw new Error('unknown workflow: ' + slug);
    const options = { ...defaults };
    if (slug === 'annual-cycle-pattern-decoder')
        options['max-period'] = '125';
    for (let i = 2; i < arguments_.length; i++) {
        const key = arguments_[i].replace(/^--/, '');
        if (!arguments_[i].startsWith('--') || !valueOptions.has(key))
            throw new Error('unknown option ' + arguments_[i]);
        if (booleanOptions.has(key))
            options[key] = true;
        else if (key === 'models') {
            const paths = arguments_.slice(i + 1, i + 7);
            if (paths.length !== 6 || paths.some(p => p.startsWith('--')))
                throw new Error('--models requires six paths');
            options[key] = paths.join('\u0000');
            i += 6;
        }
        else {
            if (++i >= arguments_.length || arguments_[i].startsWith('--'))
                throw new Error('missing value for --' + key);
            options[key] = arguments_[i];
        }
    }
    const text = (k: string): string => String(options[k] ?? '');
    const required = (k: string): string => { const v = text(k); if (!v)
        throw new Error('missing --' + k); return v; };
    const number = (k: string): number => { const v = Number(required(k)); if (!Number.isFinite(v))
        throw new Error('nonfinite --' + k); return v; };
    const integer = (k: string): number => { const v = number(k); if (!Number.isInteger(v))
        throw new Error('integer required for --' + k); return v; };
    const saveJson = (path: string, value: unknown): void => writeFileSync(path, JSON.stringify(value, null, 2) + '\n');
    const print = (value: unknown): void => console.log(JSON.stringify(value));
    const loadAnnual = (path: string): annual.AnnualModel => {
        const v: unknown = JSON.parse(readFileSync(path, 'utf8'));
        annual.validateAnnualModel(v);
        return v;
    };
    if (command === 'phase' && slug === 'cycle-phase-score') {
        const phase = number('radians');
        print({ value: phaseState(phase), status: phaseStatus(phase) });
        return;
    }
    if (slug === 'annual-cycle-pattern-decoder') {
        let model: annual.AnnualModel;
        if (command === 'fit') {
            const params = annual.defaultParams();
            params.minPeriod = integer('min-period');
            params.maxPeriod = integer('max-period');
            params.smoothing = number('smoothing');
            params.genuineThreshold = number('genuine-threshold');
            params.candidatePool = integer('candidate-pool');
            params.maxComponents = integer('max-components');
            params.peakDistance = integer('peak-distance');
            model = annual.decodeAnnual(readPrices(required('input')), integer('year'), text('as-of'), text('calendar') ? readCalendar(text('calendar')) : undefined, options['confirm-complete-year'] === true, params);
            required('wave-out');
            saveJson(required('model-out'), model);
        }
        else if (command === 'replay')
            model = loadAnnual(required('model'));
        else
            throw new Error('annual workflow supports fit or replay');
        if (!model.dates)
            throw new Error('replay requires model dates');
        saveWave(required('wave-out'), model.dates, model.dates.map((_, i) => annual.annualWave(model, i)));
        print({ year: model.year, correlation: model.correlation, components: model.components.length, sample_count: model.dates.length });
        return;
    }
    if (slug === 'harmonic-energy-superposition-wave') {
        let model: superposition.SuperpositionModel;
        let values: number[];
        if (command === 'build') {
            const dates = readCalendar(required('calendar'));
            const year = integer('target-year');
            annual.validateDates(dates, year);
            if (Boolean(text('models')) === Boolean(text('history')))
                throw new Error('provide --models or --history');
            const history = text('history') ? readPrices(text('history')) : [];
            const models = text('models') ? text('models').split('\u0000').map(loadAnnual) :
                superposition.METONIC_LAGS.map(lag => annual.decodeAnnual(history, year - lag, text('as-of'), undefined, options['confirm-complete-years'] === true));
            const result = superposition.superpositionValues(models, year, dates.length, text('normalization') as 'none' | 'rms', text('calendar-policy') as 'strict' | 'extend', options.invert === true);
            model = { ...result.model, dates };
            values = result.values;
            required('wave-out');
            saveJson(required('model-out'), model);
        }
        else if (command === 'replay') {
            const v: unknown = JSON.parse(readFileSync(required('model'), 'utf8'));
            superposition.validateSuperpositionModel(v);
            model = v;
            values = superposition.superpositionValues(model.annual_models, model.target_year, model.sample_count, model.normalization, model.calendar_policy, model.invert).values;
        }
        else
            throw new Error('superposition workflow supports build or replay');
        if (!model.dates)
            throw new Error('replay requires model dates');
        saveWave(required('wave-out'), model.dates, values);
        print({ target_year: model.target_year, source_years: model.source_years, sample_count: values.length });
        return;
    }
    if (command !== 'analyze')
        throw new Error('streaming workflows support analyze');
    if (text('start'))
        annual.validateDates([text('start')]);
    if (text('end'))
        annual.validateDates([text('end')]);
    if (text('start') && text('end') && text('start') > text('end'))
        throw new Error('start must not exceed end');
    const rows = readPrices(required('input')).filter(r => (!text('start') || r.date >= text('start')) && (!text('end') || r.date <= text('end')));
    if (rows.length < 6)
        throw new Error('selected range needs >=6 prices');
    const horizon = integer('horizon');
    if (horizon < 0)
        throw new Error('horizon must be nonnegative');
    const prices = rows.map(r => r.close);
    const dates = rows.map(r => r.date);
    const rolling = options.rolling === true;
    const config = new ScanConfig({ window: rolling ? integer('window') : prices.length, minPeriod: integer('min-period'),
        maxPeriod: integer('max-period'), smoothing: number('smoothing'), genuineThreshold: number('genuine-threshold'),
        phaseCycles: number('phase-cycles'), topK: integer('top-k'), peakDistance: integer('peak-distance'), maxSegments: integer('max-segments') });
    const invalid = config.validate();
    if (invalid)
        throw invalid;
    const analyze = (values: number[], c = config): ScanOutput => values.length < c.window ? emptyScan(c) : scanCycles(values.slice(-c.window), c);
    const scan = analyze(prices);
    const report = scannerReport(scan, config, dates.slice(-config.window));
    const short = new ScanConfig({ ...config, minPeriod: integer('short-min'), maxPeriod: integer('short-max'), topK: 1 });
    const long = new ScanConfig({ ...config, minPeriod: integer('long-min'), maxPeriod: integer('long-max'), topK: 1 });
    if (slug === 'phasing-oscillator') {
        if (short.maxPeriod >= long.minPeriod)
            throw new Error('phasing ranges must be disjoint');
        for (const c of [short, long]) {
            const e = c.validate();
            if (e)
                throw e;
        }
    }
    const calculate = (values: number[]): Row => {
        const current = analyze(values);
        const c = current.cycles[0];
        if (slug === 'complete-cycle-scanner')
            return flattenScan(current, config);
        if (slug === 'dominant-market-vibration')
            return record(c);
        if (slug === 'cycle-phase-score')
            return { value: phaseState(c.phase), period: c.period, phase: c.phase, status: phaseStatus(c.phase) };
        if (slug === 'phasing-oscillator') {
            const a = analyze(values, short).cycles[0];
            const b = analyze(values, long).cycles[0];
            const x = phaseState(a.phase);
            const y = phaseState(b.phase);
            return { value: x + y, short_score: x, long_score: y, short_period: a.period, long_period: b.period };
        }
        let highlighter = NaN;
        let index = NaN;
        let fitted = NaN;
        if (Number.isFinite(c.period)) {
            const length = Math.trunc(c.period);
            const delay = Math.trunc(length / 2);
            const buffer = values.slice(-config.window);
            const center = buffer.length - 1 - delay;
            highlighter = centeredMean(buffer, center, Math.max(2, Math.floor(c.period / 2 + 0.5))) - centeredMean(buffer, center, length);
            if (Number.isFinite(highlighter)) {
                index = values.length - 1 - delay;
                fitted = modeledValue(c, -delay);
            }
        }
        return { value: c.value, period: c.period, phase: c.phase, eta_top: c.etaTop, eta_bottom: c.etaBottom,
            highlighter, highlighter_index: index, fitted_at_highlighter: fitted };
    };
    report.output = calculate(prices);
    const series: Row[] = rows.map((r, i) => {
        if (rolling || slug === 'phasing-oscillator')
            return { date: r.date, ...calculate(prices.slice(0, i + 1)) };
        const c = scan.cycles[0];
        const row: Row = { date: r.date, fitted: modeledValue(c, i - (prices.length - 1)) };
        if (slug === 'dynamic-cycle-explorer') {
            const length = Math.trunc(c.period);
            row.highlighter = Number.isFinite(c.period) ? centeredMean(prices, i, Math.max(2, Math.floor(length / 2 + 0.5))) - centeredMean(prices, i, length) : NaN;
        }
        return row;
    });
    let sideCycles: RankedCycle[] = [];
    if (slug === 'phasing-oscillator') {
        const a = analyze(prices, short);
        const b = analyze(prices, long);
        sideCycles = [a.cycles[0], b.cycles[0]];
        report.short_scan = scannerReport(a, short, dates.slice(-short.window));
        report.long_scan = scannerReport(b, long, dates.slice(-long.window));
        report.phasing_config = { short: record(short), long: record(long) };
    }
    report.projection = Array.from({ length: horizon }, (_, i) => {
        const h = i + 1;
        const c = scan.cycles[0];
        if (slug === 'phasing-oscillator') {
            const scores = sideCycles.map(c => phaseState(c.phase + 2 * Math.PI * h / c.period));
            return { bars_ahead: h, value: scores[0] + scores[1], short_score: scores[0], long_score: scores[1] };
        }
        return { bars_ahead: h, value: slug === 'cycle-phase-score' ? phaseState(c.phase + 2 * Math.PI * h / c.period) : modeledValue(c, h) };
    });
    if (text('series-out'))
        saveSeries(text('series-out'), series);
    if (text('json-out'))
        saveJson(text('json-out'), report);
    print(report);
}
try {
    run(process.argv.slice(2));
}
catch (error) {
    console.error(error instanceof Error ? error.message : String(error));
    process.exitCode = 1;
}
