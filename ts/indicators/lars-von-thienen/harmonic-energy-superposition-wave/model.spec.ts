import * as data from './testdata.js';
import { fitAnnualValues, defaultParams, AnnualWaveReader } from '../annual-cycle-model/model.js';
import { SuperpositionWaveReader, METONIC_LAGS } from '../harmonic-energy-superposition-wave/model.js';
function historicalModels() { return METONIC_LAGS.map((lag, j) => { const p = defaultParams(); p.maxPeriod = 80; return fitAnnualValues(data.testInput.map((_, i) => (1 + 0.1 * j) * data.testInput[(i + 7 * j) % 252]), 2021 - lag, p); }); }
describe("harmonic-energy-superposition-wave prepared models", () => {
    it("case 0", () => {
        const models = historicalModels();
        const reader = new SuperpositionWaveReader(models, 2021, 252, "none", "strict", false);
        for (const expected of data.expectedN0_C0_I0_WAVE)
            expect(reader.update()).toBeCloseTo(expected, 10);
        reader.reset();
        expect(reader.update()).toBeCloseTo(data.expectedN0_C0_I0_WAVE[0], 10);
    });
    it("case 1", () => {
        const models = historicalModels();
        const reader = new SuperpositionWaveReader(models, 2021, 252, "rms", "strict", false);
        for (const expected of data.expectedN1_C0_I0_WAVE)
            expect(reader.update()).toBeCloseTo(expected, 10);
        reader.reset();
        expect(reader.update()).toBeCloseTo(data.expectedN1_C0_I0_WAVE[0], 10);
    });
    it("case 2", () => {
        const models = historicalModels();
        const reader = new SuperpositionWaveReader(models, 2021, 252, "rms", "strict", true);
        for (const expected of data.expectedN1_C0_I1_WAVE)
            expect(reader.update()).toBeCloseTo(expected, 10);
        reader.reset();
        expect(reader.update()).toBeCloseTo(data.expectedN1_C0_I1_WAVE[0], 10);
    });
    it("case 3", () => {
        const models = historicalModels();
        for (let i = 0; i < models.length; i++)
            models[i] = { ...models[i], sample_count: 250 };
        const reader = new SuperpositionWaveReader(models, 2021, 252, "rms", "extend", false);
        for (const expected of data.expectedN1_C1_I0_WAVE)
            expect(reader.update()).toBeCloseTo(expected, 10);
        reader.reset();
        expect(reader.update()).toBeCloseTo(data.expectedN1_C1_I0_WAVE[0], 10);
    });
});
