import { DynamicCycleExplorer } from "./dynamic-cycle-explorer.js";
import { defaultParams } from "./params.js";
import * as data from "./testdata.js";
describe("Lars von Thienen DynamicCycleExplorer", () => {
    it("matches prepared case 0", () => {
        const parameters = { ...defaultParams(), ...{ "genuineThreshold": 49.0, "maxPeriod": 80, "maxSegments": 5, "minPeriod": 5, "peakDistance": 2, "phaseCycles": 2.0, "smoothing": 1600.0, "topK": 3, "window": 200 } };
        const indicator = new DynamicCycleExplorer(parameters);
        const expected: Record<string, number[]> = { "VALUE": data.expectedW200_P05_P180_L1600_G49_F2_K3_D2_M5_VALUE, "PERIOD": data.expectedW200_P05_P180_L1600_G49_F2_K3_D2_M5_PERIOD, "PHASE": data.expectedW200_P05_P180_L1600_G49_F2_K3_D2_M5_PHASE, "ETA_TOP": data.expectedW200_P05_P180_L1600_G49_F2_K3_D2_M5_ETA_TOP, "ETA_BOTTOM": data.expectedW200_P05_P180_L1600_G49_F2_K3_D2_M5_ETA_BOTTOM, "HIGHLIGHTER": data.expectedW200_P05_P180_L1600_G49_F2_K3_D2_M5_HIGHLIGHTER, "HIGHLIGHTER_INDEX": data.expectedW200_P05_P180_L1600_G49_F2_K3_D2_M5_HIGHLIGHTER_INDEX, "FITTED_AT_HIGHLIGHTER": data.expectedW200_P05_P180_L1600_G49_F2_K3_D2_M5_FITTED_AT_HIGHLIGHTER };
        for (let index = 0; index < data.testInput.length; index++) {
            const result = indicator.update(data.testInput[index]);
            const actual: Record<string, number> = { "VALUE": result.value, "PERIOD": result.period, "PHASE": result.phase, "ETA_TOP": result.etaTop, "ETA_BOTTOM": result.etaBottom, "HIGHLIGHTER": result.highlighter, "HIGHLIGHTER_INDEX": result.highlighterIndex, "FITTED_AT_HIGHLIGHTER": result.fittedAtHighlighter };
            for (const [field, values] of Object.entries(expected)) {
                const value = actual[field];
                if (Number.isNaN(values[index]))
                    expect(Number.isNaN(value)).withContext(`bar ${index} ${field}`).toBeTrue();
                else
                    expect(Math.abs(value - values[index])).withContext(`bar ${index} ${field}: ${value} vs ${values[index]}`).toBeLessThanOrEqual(Math.max(1e-10, Math.abs(values[index]) * 5e-15));
            }
        }
        indicator.reset();
        expect(indicator.isPrimed()).toBeFalse();
    });
    it("matches prepared case 1", () => {
        const parameters = { ...defaultParams(), ...{ "genuineThreshold": 49.0, "maxPeriod": 12, "maxSegments": 5, "minPeriod": 12, "peakDistance": 2, "phaseCycles": 0.0, "smoothing": 1600.0, "topK": 1, "window": 60 } };
        const indicator = new DynamicCycleExplorer(parameters);
        const expected: Record<string, number[]> = { "VALUE": data.expectedW60_P012_P112_L1600_G49_F0_K1_D2_M5_VALUE, "PERIOD": data.expectedW60_P012_P112_L1600_G49_F0_K1_D2_M5_PERIOD, "PHASE": data.expectedW60_P012_P112_L1600_G49_F0_K1_D2_M5_PHASE, "ETA_TOP": data.expectedW60_P012_P112_L1600_G49_F0_K1_D2_M5_ETA_TOP, "ETA_BOTTOM": data.expectedW60_P012_P112_L1600_G49_F0_K1_D2_M5_ETA_BOTTOM, "HIGHLIGHTER": data.expectedW60_P012_P112_L1600_G49_F0_K1_D2_M5_HIGHLIGHTER, "HIGHLIGHTER_INDEX": data.expectedW60_P012_P112_L1600_G49_F0_K1_D2_M5_HIGHLIGHTER_INDEX, "FITTED_AT_HIGHLIGHTER": data.expectedW60_P012_P112_L1600_G49_F0_K1_D2_M5_FITTED_AT_HIGHLIGHTER };
        for (let index = 0; index < data.testInput.length; index++) {
            const result = indicator.update(data.testInput[index]);
            const actual: Record<string, number> = { "VALUE": result.value, "PERIOD": result.period, "PHASE": result.phase, "ETA_TOP": result.etaTop, "ETA_BOTTOM": result.etaBottom, "HIGHLIGHTER": result.highlighter, "HIGHLIGHTER_INDEX": result.highlighterIndex, "FITTED_AT_HIGHLIGHTER": result.fittedAtHighlighter };
            for (const [field, values] of Object.entries(expected)) {
                const value = actual[field];
                if (Number.isNaN(values[index]))
                    expect(Number.isNaN(value)).withContext(`bar ${index} ${field}`).toBeTrue();
                else
                    expect(Math.abs(value - values[index])).withContext(`bar ${index} ${field}: ${value} vs ${values[index]}`).toBeLessThanOrEqual(Math.max(1e-10, Math.abs(values[index]) * 5e-15));
            }
        }
        indicator.reset();
        expect(indicator.isPrimed()).toBeFalse();
    });
    it("matches prepared case 2", () => {
        const parameters = { ...defaultParams(), ...{ "genuineThreshold": 49.0, "maxPeriod": 30, "maxSegments": 5, "minPeriod": 5, "peakDistance": 2, "phaseCycles": 2.0, "smoothing": 0.0, "topK": 2, "window": 100 } };
        const indicator = new DynamicCycleExplorer(parameters);
        const expected: Record<string, number[]> = { "VALUE": data.expectedW100_P05_P130_L0_G49_F2_K2_D2_M5_VALUE, "PERIOD": data.expectedW100_P05_P130_L0_G49_F2_K2_D2_M5_PERIOD, "PHASE": data.expectedW100_P05_P130_L0_G49_F2_K2_D2_M5_PHASE, "ETA_TOP": data.expectedW100_P05_P130_L0_G49_F2_K2_D2_M5_ETA_TOP, "ETA_BOTTOM": data.expectedW100_P05_P130_L0_G49_F2_K2_D2_M5_ETA_BOTTOM, "HIGHLIGHTER": data.expectedW100_P05_P130_L0_G49_F2_K2_D2_M5_HIGHLIGHTER, "HIGHLIGHTER_INDEX": data.expectedW100_P05_P130_L0_G49_F2_K2_D2_M5_HIGHLIGHTER_INDEX, "FITTED_AT_HIGHLIGHTER": data.expectedW100_P05_P130_L0_G49_F2_K2_D2_M5_FITTED_AT_HIGHLIGHTER };
        for (let index = 0; index < data.testInput.length; index++) {
            const result = indicator.update(data.testInput[index]);
            const actual: Record<string, number> = { "VALUE": result.value, "PERIOD": result.period, "PHASE": result.phase, "ETA_TOP": result.etaTop, "ETA_BOTTOM": result.etaBottom, "HIGHLIGHTER": result.highlighter, "HIGHLIGHTER_INDEX": result.highlighterIndex, "FITTED_AT_HIGHLIGHTER": result.fittedAtHighlighter };
            for (const [field, values] of Object.entries(expected)) {
                const value = actual[field];
                if (Number.isNaN(values[index]))
                    expect(Number.isNaN(value)).withContext(`bar ${index} ${field}`).toBeTrue();
                else
                    expect(Math.abs(value - values[index])).withContext(`bar ${index} ${field}: ${value} vs ${values[index]}`).toBeLessThanOrEqual(Math.max(1e-10, Math.abs(values[index]) * 5e-15));
            }
        }
        indicator.reset();
        expect(indicator.isPrimed()).toBeFalse();
    });
    it("matches prepared case 3", () => {
        const parameters = { ...defaultParams(), ...{ "genuineThreshold": 100.0, "maxPeriod": 30, "maxSegments": 5, "minPeriod": 5, "peakDistance": 2, "phaseCycles": 2.0, "smoothing": 1600.0, "topK": 2, "window": 100 } };
        const indicator = new DynamicCycleExplorer(parameters);
        const expected: Record<string, number[]> = { "VALUE": data.expectedW100_P05_P130_L1600_G100_F2_K2_D2_M5_VALUE, "PERIOD": data.expectedW100_P05_P130_L1600_G100_F2_K2_D2_M5_PERIOD, "PHASE": data.expectedW100_P05_P130_L1600_G100_F2_K2_D2_M5_PHASE, "ETA_TOP": data.expectedW100_P05_P130_L1600_G100_F2_K2_D2_M5_ETA_TOP, "ETA_BOTTOM": data.expectedW100_P05_P130_L1600_G100_F2_K2_D2_M5_ETA_BOTTOM, "HIGHLIGHTER": data.expectedW100_P05_P130_L1600_G100_F2_K2_D2_M5_HIGHLIGHTER, "HIGHLIGHTER_INDEX": data.expectedW100_P05_P130_L1600_G100_F2_K2_D2_M5_HIGHLIGHTER_INDEX, "FITTED_AT_HIGHLIGHTER": data.expectedW100_P05_P130_L1600_G100_F2_K2_D2_M5_FITTED_AT_HIGHLIGHTER };
        for (let index = 0; index < data.testInput.length; index++) {
            const result = indicator.update(data.testInput[index]);
            const actual: Record<string, number> = { "VALUE": result.value, "PERIOD": result.period, "PHASE": result.phase, "ETA_TOP": result.etaTop, "ETA_BOTTOM": result.etaBottom, "HIGHLIGHTER": result.highlighter, "HIGHLIGHTER_INDEX": result.highlighterIndex, "FITTED_AT_HIGHLIGHTER": result.fittedAtHighlighter };
            for (const [field, values] of Object.entries(expected)) {
                const value = actual[field];
                if (Number.isNaN(values[index]))
                    expect(Number.isNaN(value)).withContext(`bar ${index} ${field}`).toBeTrue();
                else
                    expect(Math.abs(value - values[index])).withContext(`bar ${index} ${field}: ${value} vs ${values[index]}`).toBeLessThanOrEqual(Math.max(1e-10, Math.abs(values[index]) * 5e-15));
            }
        }
        indicator.reset();
        expect(indicator.isPrimed()).toBeFalse();
    });
    it("skips NaN and rejects invalid observations before priming", () => {
        const indicator = new DynamicCycleExplorer();
        indicator.update(NaN);
        expect(indicator.isPrimed()).toBeFalse();
        expect(() => indicator.update(Infinity)).toThrow();
    });
});
