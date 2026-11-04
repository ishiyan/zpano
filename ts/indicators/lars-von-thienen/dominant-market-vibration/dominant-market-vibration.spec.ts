import { DominantMarketVibration } from "./dominant-market-vibration.js";
import { defaultParams } from "./params.js";
import * as data from "./testdata.js";
describe("Lars von Thienen DominantMarketVibration", () => {
    it("matches prepared case 0", () => {
        const parameters = { ...defaultParams(), ...{ "genuineThreshold": 49.0, "maxPeriod": 80, "maxSegments": 5, "minPeriod": 5, "peakDistance": 2, "phaseCycles": 2.0, "smoothing": 1600.0, "topK": 3, "window": 200 } };
        const indicator = new DominantMarketVibration(parameters);
        const expected: Record<string, number[]> = { "PERIOD": data.expectedW200_P05_P180_L1600_G49_F2_K3_D2_M5_PERIOD, "AMPLITUDE": data.expectedW200_P05_P180_L1600_G49_F2_K3_D2_M5_AMPLITUDE, "PHASE": data.expectedW200_P05_P180_L1600_G49_F2_K3_D2_M5_PHASE, "STRENGTH": data.expectedW200_P05_P180_L1600_G49_F2_K3_D2_M5_STRENGTH, "GENUINE_PERCENT": data.expectedW200_P05_P180_L1600_G49_F2_K3_D2_M5_GENUINE_PERCENT, "SEGMENTS": data.expectedW200_P05_P180_L1600_G49_F2_K3_D2_M5_SEGMENTS, "FIT_AMPLITUDE": data.expectedW200_P05_P180_L1600_G49_F2_K3_D2_M5_FIT_AMPLITUDE, "FIT_SAMPLES": data.expectedW200_P05_P180_L1600_G49_F2_K3_D2_M5_FIT_SAMPLES, "VALUE": data.expectedW200_P05_P180_L1600_G49_F2_K3_D2_M5_VALUE, "SLOPE": data.expectedW200_P05_P180_L1600_G49_F2_K3_D2_M5_SLOPE, "ETA_TOP": data.expectedW200_P05_P180_L1600_G49_F2_K3_D2_M5_ETA_TOP, "ETA_BOTTOM": data.expectedW200_P05_P180_L1600_G49_F2_K3_D2_M5_ETA_BOTTOM };
        for (let index = 0; index < data.testInput.length; index++) {
            const result = indicator.update(data.testInput[index]);
            const actual: Record<string, number> = { "PERIOD": result.period, "AMPLITUDE": result.amplitude, "PHASE": result.phase, "STRENGTH": result.strength, "GENUINE_PERCENT": result.genuinePercent, "SEGMENTS": result.segments, "FIT_AMPLITUDE": result.fitAmplitude, "FIT_SAMPLES": result.fitSamples, "VALUE": result.value, "SLOPE": result.slope, "ETA_TOP": result.etaTop, "ETA_BOTTOM": result.etaBottom };
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
        const indicator = new DominantMarketVibration(parameters);
        const expected: Record<string, number[]> = { "PERIOD": data.expectedW60_P012_P112_L1600_G49_F0_K1_D2_M5_PERIOD, "AMPLITUDE": data.expectedW60_P012_P112_L1600_G49_F0_K1_D2_M5_AMPLITUDE, "PHASE": data.expectedW60_P012_P112_L1600_G49_F0_K1_D2_M5_PHASE, "STRENGTH": data.expectedW60_P012_P112_L1600_G49_F0_K1_D2_M5_STRENGTH, "GENUINE_PERCENT": data.expectedW60_P012_P112_L1600_G49_F0_K1_D2_M5_GENUINE_PERCENT, "SEGMENTS": data.expectedW60_P012_P112_L1600_G49_F0_K1_D2_M5_SEGMENTS, "FIT_AMPLITUDE": data.expectedW60_P012_P112_L1600_G49_F0_K1_D2_M5_FIT_AMPLITUDE, "FIT_SAMPLES": data.expectedW60_P012_P112_L1600_G49_F0_K1_D2_M5_FIT_SAMPLES, "VALUE": data.expectedW60_P012_P112_L1600_G49_F0_K1_D2_M5_VALUE, "SLOPE": data.expectedW60_P012_P112_L1600_G49_F0_K1_D2_M5_SLOPE, "ETA_TOP": data.expectedW60_P012_P112_L1600_G49_F0_K1_D2_M5_ETA_TOP, "ETA_BOTTOM": data.expectedW60_P012_P112_L1600_G49_F0_K1_D2_M5_ETA_BOTTOM };
        for (let index = 0; index < data.testInput.length; index++) {
            const result = indicator.update(data.testInput[index]);
            const actual: Record<string, number> = { "PERIOD": result.period, "AMPLITUDE": result.amplitude, "PHASE": result.phase, "STRENGTH": result.strength, "GENUINE_PERCENT": result.genuinePercent, "SEGMENTS": result.segments, "FIT_AMPLITUDE": result.fitAmplitude, "FIT_SAMPLES": result.fitSamples, "VALUE": result.value, "SLOPE": result.slope, "ETA_TOP": result.etaTop, "ETA_BOTTOM": result.etaBottom };
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
        const indicator = new DominantMarketVibration(parameters);
        const expected: Record<string, number[]> = { "PERIOD": data.expectedW100_P05_P130_L0_G49_F2_K2_D2_M5_PERIOD, "AMPLITUDE": data.expectedW100_P05_P130_L0_G49_F2_K2_D2_M5_AMPLITUDE, "PHASE": data.expectedW100_P05_P130_L0_G49_F2_K2_D2_M5_PHASE, "STRENGTH": data.expectedW100_P05_P130_L0_G49_F2_K2_D2_M5_STRENGTH, "GENUINE_PERCENT": data.expectedW100_P05_P130_L0_G49_F2_K2_D2_M5_GENUINE_PERCENT, "SEGMENTS": data.expectedW100_P05_P130_L0_G49_F2_K2_D2_M5_SEGMENTS, "FIT_AMPLITUDE": data.expectedW100_P05_P130_L0_G49_F2_K2_D2_M5_FIT_AMPLITUDE, "FIT_SAMPLES": data.expectedW100_P05_P130_L0_G49_F2_K2_D2_M5_FIT_SAMPLES, "VALUE": data.expectedW100_P05_P130_L0_G49_F2_K2_D2_M5_VALUE, "SLOPE": data.expectedW100_P05_P130_L0_G49_F2_K2_D2_M5_SLOPE, "ETA_TOP": data.expectedW100_P05_P130_L0_G49_F2_K2_D2_M5_ETA_TOP, "ETA_BOTTOM": data.expectedW100_P05_P130_L0_G49_F2_K2_D2_M5_ETA_BOTTOM };
        for (let index = 0; index < data.testInput.length; index++) {
            const result = indicator.update(data.testInput[index]);
            const actual: Record<string, number> = { "PERIOD": result.period, "AMPLITUDE": result.amplitude, "PHASE": result.phase, "STRENGTH": result.strength, "GENUINE_PERCENT": result.genuinePercent, "SEGMENTS": result.segments, "FIT_AMPLITUDE": result.fitAmplitude, "FIT_SAMPLES": result.fitSamples, "VALUE": result.value, "SLOPE": result.slope, "ETA_TOP": result.etaTop, "ETA_BOTTOM": result.etaBottom };
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
        const indicator = new DominantMarketVibration(parameters);
        const expected: Record<string, number[]> = { "PERIOD": data.expectedW100_P05_P130_L1600_G100_F2_K2_D2_M5_PERIOD, "AMPLITUDE": data.expectedW100_P05_P130_L1600_G100_F2_K2_D2_M5_AMPLITUDE, "PHASE": data.expectedW100_P05_P130_L1600_G100_F2_K2_D2_M5_PHASE, "STRENGTH": data.expectedW100_P05_P130_L1600_G100_F2_K2_D2_M5_STRENGTH, "GENUINE_PERCENT": data.expectedW100_P05_P130_L1600_G100_F2_K2_D2_M5_GENUINE_PERCENT, "SEGMENTS": data.expectedW100_P05_P130_L1600_G100_F2_K2_D2_M5_SEGMENTS, "FIT_AMPLITUDE": data.expectedW100_P05_P130_L1600_G100_F2_K2_D2_M5_FIT_AMPLITUDE, "FIT_SAMPLES": data.expectedW100_P05_P130_L1600_G100_F2_K2_D2_M5_FIT_SAMPLES, "VALUE": data.expectedW100_P05_P130_L1600_G100_F2_K2_D2_M5_VALUE, "SLOPE": data.expectedW100_P05_P130_L1600_G100_F2_K2_D2_M5_SLOPE, "ETA_TOP": data.expectedW100_P05_P130_L1600_G100_F2_K2_D2_M5_ETA_TOP, "ETA_BOTTOM": data.expectedW100_P05_P130_L1600_G100_F2_K2_D2_M5_ETA_BOTTOM };
        for (let index = 0; index < data.testInput.length; index++) {
            const result = indicator.update(data.testInput[index]);
            const actual: Record<string, number> = { "PERIOD": result.period, "AMPLITUDE": result.amplitude, "PHASE": result.phase, "STRENGTH": result.strength, "GENUINE_PERCENT": result.genuinePercent, "SEGMENTS": result.segments, "FIT_AMPLITUDE": result.fitAmplitude, "FIT_SAMPLES": result.fitSamples, "VALUE": result.value, "SLOPE": result.slope, "ETA_TOP": result.etaTop, "ETA_BOTTOM": result.etaBottom };
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
        const indicator = new DominantMarketVibration();
        indicator.update(NaN);
        expect(indicator.isPrimed()).toBeFalse();
        expect(() => indicator.update(Infinity)).toThrow();
    });
});
