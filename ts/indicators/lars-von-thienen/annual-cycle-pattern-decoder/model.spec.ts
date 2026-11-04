import * as data from './testdata.js';
import { fitAnnualValues, defaultParams, AnnualWaveReader } from '../annual-cycle-model/model.js';
import { SuperpositionWaveReader, METONIC_LAGS } from '../harmonic-energy-superposition-wave/model.js';
describe("annual-cycle-pattern-decoder prepared models", () => {
    it("case 0", () => {
        const p = defaultParams();
        p.candidatePool = 10;
        p.genuineThreshold = 49.0;
        p.maxComponents = 7;
        p.maxPeriod = 125;
        p.minPeriod = 5;
        p.peakDistance = 2;
        p.smoothing = 1600.0;
        const model = fitAnnualValues(data.testInput, 2020, p);
        const reader = new AnnualWaveReader(model);
        expect(model.correlation).toBeCloseTo(data.expectedP05_P1125_L1600_G49_C10_K7_D2_CORRELATION[0], 13);
        expect(model.components.length).toBe(data.expectedP05_P1125_L1600_G49_C10_K7_D2_COMPONENT_COUNT[0]);
        for (const expected of data.expectedP05_P1125_L1600_G49_C10_K7_D2_WAVE)
            expect(reader.update()).toBeCloseTo(expected, 10);
        reader.reset();
        expect(reader.update()).toBeCloseTo(data.expectedP05_P1125_L1600_G49_C10_K7_D2_WAVE[0], 10);
    });
    it("case 1", () => {
        const p = defaultParams();
        p.candidatePool = 10;
        p.genuineThreshold = 49.0;
        p.maxComponents = 1;
        p.maxPeriod = 125;
        p.minPeriod = 5;
        p.peakDistance = 2;
        p.smoothing = 1600.0;
        const model = fitAnnualValues(data.testInput, 2020, p);
        const reader = new AnnualWaveReader(model);
        expect(model.correlation).toBeCloseTo(data.expectedP05_P1125_L1600_G49_C10_K1_D2_CORRELATION[0], 13);
        expect(model.components.length).toBe(data.expectedP05_P1125_L1600_G49_C10_K1_D2_COMPONENT_COUNT[0]);
        for (const expected of data.expectedP05_P1125_L1600_G49_C10_K1_D2_WAVE)
            expect(reader.update()).toBeCloseTo(expected, 10);
        reader.reset();
        expect(reader.update()).toBeCloseTo(data.expectedP05_P1125_L1600_G49_C10_K1_D2_WAVE[0], 10);
    });
    it("case 2", () => {
        const p = defaultParams();
        p.candidatePool = 7;
        p.genuineThreshold = 49.0;
        p.maxComponents = 3;
        p.maxPeriod = 80;
        p.minPeriod = 5;
        p.peakDistance = 2;
        p.smoothing = 1600.0;
        const model = fitAnnualValues(data.testInput, 2020, p);
        const reader = new AnnualWaveReader(model);
        expect(model.correlation).toBeCloseTo(data.expectedP05_P180_L1600_G49_C7_K3_D2_CORRELATION[0], 13);
        expect(model.components.length).toBe(data.expectedP05_P180_L1600_G49_C7_K3_D2_COMPONENT_COUNT[0]);
        for (const expected of data.expectedP05_P180_L1600_G49_C7_K3_D2_WAVE)
            expect(reader.update()).toBeCloseTo(expected, 10);
        reader.reset();
        expect(reader.update()).toBeCloseTo(data.expectedP05_P180_L1600_G49_C7_K3_D2_WAVE[0], 10);
    });
    it("case 3", () => {
        const p = defaultParams();
        p.candidatePool = 10;
        p.genuineThreshold = 49.0;
        p.maxComponents = 7;
        p.maxPeriod = 80;
        p.minPeriod = 5;
        p.peakDistance = 2;
        p.smoothing = 16000.0;
        const model = fitAnnualValues(data.testInput, 2020, p);
        const reader = new AnnualWaveReader(model);
        expect(model.correlation).toBeCloseTo(data.expectedP05_P180_L16000_G49_C10_K7_D2_CORRELATION[0], 13);
        expect(model.components.length).toBe(data.expectedP05_P180_L16000_G49_C10_K7_D2_COMPONENT_COUNT[0]);
        for (const expected of data.expectedP05_P180_L16000_G49_C10_K7_D2_WAVE)
            expect(reader.update()).toBeCloseTo(expected, 10);
        reader.reset();
        expect(reader.update()).toBeCloseTo(data.expectedP05_P180_L16000_G49_C10_K7_D2_WAVE[0], 10);
    });
});
