import { annualWave, validateAnnualModel, AnnualModel, validateDates } from '../annual-cycle-model/model.js';
import { sum } from '../harmonic-cycle/summation.js';
export const METONIC_LAGS = [8, 11, 19, 27, 30, 38] as const;
export interface SuperpositionModel {
    readonly schema: 'lvt-superposition-model';
    readonly version: 1;
    readonly research_reconstruction: true;
    readonly target_year: number;
    readonly sample_count: number;
    readonly source_years: readonly number[];
    readonly lags: readonly number[];
    readonly normalization: 'none' | 'rms';
    readonly calendar_policy: 'strict' | 'extend';
    readonly invert: boolean;
    readonly scales: readonly {
        year: number;
        mean: number;
        scale: number;
        sample_count: number;
    }[];
    readonly annual_models: readonly AnnualModel[];
    readonly dates?: readonly string[];
}
/** Align six source waves at ordinal zero; normalization reads only each source year. */
export function superpositionValues(models: readonly AnnualModel[], targetYear: number, count: number, normalization: 'none' | 'rms' = 'rms', calendarPolicy: 'strict' | 'extend' = 'extend', invert = false): {
    values: number[];
    model: SuperpositionModel;
} {
    if (!Number.isInteger(targetYear) || targetYear < 39 || !Number.isInteger(count) || count < 1 ||
        !['none', 'rms'].includes(normalization) || !['strict', 'extend'].includes(calendarPolicy) ||
        typeof invert !== 'boolean')
        throw new Error('invalid superposition configuration');
    const byYear = new Map<number, AnnualModel>();
    for (const model of models) {
        validateAnnualModel(model);
        if (byYear.has(model.year))
            throw new Error('duplicate source year');
        byYear.set(model.year, model);
    }
    const sourceYears = METONIC_LAGS.map(lag => targetYear - lag);
    if (byYear.size !== 6 || sourceYears.some(year => !byYear.has(year))) {
        throw new Error('require target years minus 8,11,19,27,30,38');
    }
    const annualModels = sourceYears.map(year => byYear.get(year)!);
    const scales = annualModels.map(model => {
        const n = model.sample_count;
        if (calendarPolicy === 'strict' && count !== n)
            throw new Error('strict ordinal alignment requires equal sample counts');
        const base = Array.from({ length: n }, (_, i) => annualWave(model, i));
        const mean = normalization === 'rms' ? sum(base) / n : 0;
        const scale = normalization === 'rms' ? Math.sqrt(sum(base.map(v => (v - mean) * (v - mean))) / n) : 1;
        if (scale === 0)
            throw new Error('zero-energy annual wave cannot be normalized');
        return { year: model.year, mean, scale, sample_count: n };
    });
    const values = Array.from({ length: count }, (_, i) => (invert ? -1 : 1) * sum(annualModels.map((m, j) => (annualWave(m, i) - scales[j].mean) / scales[j].scale)) / 6);
    return { values, model: { schema: 'lvt-superposition-model', version: 1, research_reconstruction: true,
            target_year: targetYear, sample_count: count, source_years: sourceYears, lags: [...METONIC_LAGS],
            normalization, calendar_policy: calendarPolicy, invert, scales, annual_models: annualModels } };
}
export function validateSuperpositionModel(value: unknown): asserts value is SuperpositionModel {
    const m = value as SuperpositionModel;
    if (!m || m.schema !== 'lvt-superposition-model' || m.version !== 1)
        throw new Error('unsupported superposition model');
    if (!Array.isArray(m.lags) || m.lags.length !== 6 || m.lags.some((lag, i) => lag !== METONIC_LAGS[i]) ||
        !Array.isArray(m.source_years) || m.source_years.length !== 6 || m.source_years.some((year, i) => year !== m.target_year - METONIC_LAGS[i]) || typeof m.invert !== 'boolean')
        throw new Error('superposition source years, lags or inversion are invalid');
    superpositionValues(m.annual_models, m.target_year, m.sample_count, m.normalization, m.calendar_policy, m.invert);
    if (m.dates !== undefined) {
        validateDates(m.dates, m.target_year);
        if (m.dates.length !== m.sample_count)
            throw new Error('dates must match sample_count');
    }
}
/** Bounded frozen playback stops at the declared target calendar. */
export class SuperpositionWaveReader {
    private readonly _values: number[];
    private _index = 0;
    constructor(models: readonly AnnualModel[], targetYear: number, count: number, normalization: 'none' | 'rms' = 'rms', calendarPolicy: 'strict' | 'extend' = 'extend', invert = false) {
        this._values = superpositionValues(models, targetYear, count, normalization, calendarPolicy, invert).values;
    }
    reset(): void { this._index = 0; }
    update(): number {
        if (this._index >= this._values.length)
            throw new Error('end of declared target calendar');
        return this._values[this._index++];
    }
}
