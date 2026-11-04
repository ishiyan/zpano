import * as numerical from './numerical.js';
import { ScanConfig } from '../cycle-scanner/scanner.js';
/** Portable schema keys are deliberately preserved in saved JSON. */
export interface AnnualComponent {
    readonly period: number;
    readonly amplitude: number;
    readonly phase: number;
    readonly strength: number;
    readonly genuine_percent: number;
}
export interface AnnualModel {
    readonly schema: 'lvt-annual-cycle-model';
    readonly version: 1;
    readonly research_reconstruction: true;
    readonly year: number;
    readonly sample_count: number;
    readonly origin: string;
    readonly phase_convention: 'sine radians at ordinal zero';
    readonly components: readonly AnnualComponent[];
    readonly correlation: number;
    readonly subsets_checked: number;
    readonly candidate_count: number;
    readonly config: Record<string, number>;
    readonly dates?: readonly string[];
    readonly start_date?: string;
    readonly end_date?: string;
    readonly coverage?: string;
    readonly fitted_as_of?: string;
}
export type AnnualParams = numerical.Params;
export const defaultParams = numerical.defaultParams;
export const pearson = numerical.pearson;
export const chooseSubset = numerical.chooseSubset;
/** Strict ISO dates, in increasing order, optionally in a single year. */
export function validateDates(dates: readonly string[], year?: number): void {
    let previous = '';
    for (const day of dates) {
        const parsed = new Date(day + 'T00:00:00Z');
        if (!/^\d{4}-\d{2}-\d{2}$/.test(day) || !Number.isFinite(parsed.getTime()) ||
            parsed.toISOString().slice(0, 10) !== day || day <= previous ||
            (year !== undefined && parsed.getUTCFullYear() !== year)) {
            throw new Error('dates must be valid, strictly increasing and in the requested year');
        }
        previous = day;
    }
}
/** Fit numerical data; decodeAnnual enforces calendar completeness. */
export function fitAnnualValues(prices: number[], year: number, params: AnnualParams = defaultParams()): AnnualModel {
    const [fitted, error] = numerical.fitAnnualValues(prices, year, params);
    if (error)
        throw error;
    return {
        schema: 'lvt-annual-cycle-model', version: 1, research_reconstruction: true,
        year, sample_count: fitted.sampleCount,
        origin: 'first-observed-trading-day, zero-based ordinal',
        phase_convention: 'sine radians at ordinal zero',
        components: fitted.components.map(c => ({ period: c.period, amplitude: c.amplitude,
            phase: c.phase, strength: c.strength, genuine_percent: c.genuinePercent })),
        correlation: fitted.correlation, subsets_checked: fitted.subsetsChecked,
        candidate_count: fitted.candidateCount,
        config: { window: prices.length, min_period: params.minPeriod, max_period: params.maxPeriod,
            smoothing: params.smoothing, genuine_threshold: params.genuineThreshold, phase_cycles: 0,
            top_k: params.candidatePool, peak_distance: params.peakDistance, max_segments: 5,
            candidate_pool: params.candidatePool, max_components: params.maxComponents },
    };
}
/** Validate external models before evaluating any saved wave. */
export function validateAnnualModel(value: unknown): asserts value is AnnualModel {
    const m = value as AnnualModel;
    if (!m || m.schema !== 'lvt-annual-cycle-model' || m.version !== 1 ||
        !Number.isInteger(m.year) || m.year < 1 || !Number.isInteger(m.sample_count) ||
        m.sample_count < 6 || m.phase_convention !== 'sine radians at ordinal zero' ||
        !Array.isArray(m.components) || m.components.length < 1 || m.components.length > 7) {
        throw new Error('unsupported annual model schema or fields');
    }
    const periods = new Set<number>();
    for (const c of m.components) {
        if (!Number.isFinite(c.period) || c.period <= 2 || !Number.isFinite(c.amplitude) ||
            c.amplitude < 0 || !Number.isFinite(c.phase) || periods.has(c.period)) {
            throw new Error('invalid or duplicate annual component');
        }
        periods.add(c.period);
    }
    if (m.dates !== undefined) {
        validateDates(m.dates, m.year);
        if (m.dates.length !== m.sample_count)
            throw new Error('dates must match sample_count');
    }
    if (m.fitted_as_of !== undefined) {
        validateDates([m.fitted_as_of]);
        if (m.fitted_as_of < `${m.year + 1}-01-01`)
            throw new Error('fit predates year completion');
    }
}
export interface PriceRecord {
    readonly date: string;
    readonly close: number;
}
/** A past complete year requires an exact calendar or an explicit recorded assertion. */
export function decodeAnnual(records: readonly PriceRecord[], year: number, asOf: string, calendar?: readonly string[], confirmComplete = false, params: AnnualParams = defaultParams()): AnnualModel {
    validateDates(records.map(r => r.date));
    validateDates([asOf]);
    if (asOf < `${year + 1}-01-01`)
        throw new Error('annual decoding requires a completed year');
    const selected = records.filter(r => Number(r.date.slice(0, 4)) === year);
    const dates = selected.map(r => r.date);
    if (dates.length < 6 || dates[0].slice(5, 7) !== '01' || dates[dates.length - 1].slice(5, 7) !== '12') {
        throw new Error('calendar must span January through December');
    }
    if (calendar !== undefined) {
        validateDates(calendar, year);
        if (calendar.length !== dates.length || dates.some((d, i) => d !== calendar[i])) {
            throw new Error('observed dates must match supplied full-year calendar');
        }
    }
    else if (!confirmComplete)
        throw new Error('provide calendar or confirm complete year');
    return { ...fitAnnualValues(selected.map(r => r.close), year, params), dates,
        start_date: dates[0], end_date: dates[dates.length - 1], fitted_as_of: asOf,
        coverage: calendar ? 'exact match to supplied full-year calendar' :
            'caller asserted complete year; session completeness unverified' };
}
/** Frozen annual waves extrapolate mathematically beyond the source year. */
export function annualWave(model: AnnualModel, index: number): number {
    const fitted = new numerical.Model({ components: model.components.map(c => new numerical.Component({
            period: c.period, amplitude: c.amplitude, phase: c.phase
        })) });
    return numerical.annualWave(fitted, index);
}
export class AnnualWaveReader {
    private readonly _model: AnnualModel;
    private _index = 0;
    constructor(model: AnnualModel) {
        validateAnnualModel(model);
        this._model = structuredClone(model);
    }
    reset(): void { this._index = 0; }
    update(): number { return annualWave(this._model, this._index++); }
}
