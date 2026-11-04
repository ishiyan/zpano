import { PhasingOscillator } from "./phasing-oscillator.js";
import { defaultParams } from "./params.js";
import * as data from "./testdata.js";
describe("Lars von Thienen PhasingOscillator", () => {
    it("matches prepared case 0", () => {
        const parameters = { ...defaultParams(), ...{ "genuineThreshold": 49.0, "longMax": 80, "longMin": 31, "phaseCycles": 2.0, "shortMax": 30, "shortMin": 5, "smoothing": 1600.0, "window": 200 } };
        const indicator = new PhasingOscillator(parameters);
        const expected: Record<string, number[]> = { "VALUE": data.expectedW200_S05_S130_L031_L180_HP1600_G49_F2_VALUE, "SHORT_SCORE": data.expectedW200_S05_S130_L031_L180_HP1600_G49_F2_SHORT_SCORE, "LONG_SCORE": data.expectedW200_S05_S130_L031_L180_HP1600_G49_F2_LONG_SCORE, "SHORT_PERIOD": data.expectedW200_S05_S130_L031_L180_HP1600_G49_F2_SHORT_PERIOD, "LONG_PERIOD": data.expectedW200_S05_S130_L031_L180_HP1600_G49_F2_LONG_PERIOD };
        for (let index = 0; index < data.testInput.length; index++) {
            const result = indicator.update(data.testInput[index]);
            const actual: Record<string, number> = { "VALUE": result.value, "SHORT_SCORE": result.shortScore, "LONG_SCORE": result.longScore, "SHORT_PERIOD": result.shortPeriod, "LONG_PERIOD": result.longPeriod };
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
        const parameters = { ...defaultParams(), ...{ "genuineThreshold": 49.0, "longMax": 40, "longMin": 16, "phaseCycles": 0.0, "shortMax": 15, "shortMin": 5, "smoothing": 1600.0, "window": 100 } };
        const indicator = new PhasingOscillator(parameters);
        const expected: Record<string, number[]> = { "VALUE": data.expectedW100_S05_S115_L016_L140_HP1600_G49_F0_VALUE, "SHORT_SCORE": data.expectedW100_S05_S115_L016_L140_HP1600_G49_F0_SHORT_SCORE, "LONG_SCORE": data.expectedW100_S05_S115_L016_L140_HP1600_G49_F0_LONG_SCORE, "SHORT_PERIOD": data.expectedW100_S05_S115_L016_L140_HP1600_G49_F0_SHORT_PERIOD, "LONG_PERIOD": data.expectedW100_S05_S115_L016_L140_HP1600_G49_F0_LONG_PERIOD };
        for (let index = 0; index < data.testInput.length; index++) {
            const result = indicator.update(data.testInput[index]);
            const actual: Record<string, number> = { "VALUE": result.value, "SHORT_SCORE": result.shortScore, "LONG_SCORE": result.longScore, "SHORT_PERIOD": result.shortPeriod, "LONG_PERIOD": result.longPeriod };
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
        const parameters = { ...defaultParams(), ...{ "genuineThreshold": 49.0, "longMax": 40, "longMin": 16, "phaseCycles": 2.0, "shortMax": 15, "shortMin": 5, "smoothing": 0.0, "window": 100 } };
        const indicator = new PhasingOscillator(parameters);
        const expected: Record<string, number[]> = { "VALUE": data.expectedW100_S05_S115_L016_L140_HP0_G49_F2_VALUE, "SHORT_SCORE": data.expectedW100_S05_S115_L016_L140_HP0_G49_F2_SHORT_SCORE, "LONG_SCORE": data.expectedW100_S05_S115_L016_L140_HP0_G49_F2_LONG_SCORE, "SHORT_PERIOD": data.expectedW100_S05_S115_L016_L140_HP0_G49_F2_SHORT_PERIOD, "LONG_PERIOD": data.expectedW100_S05_S115_L016_L140_HP0_G49_F2_LONG_PERIOD };
        for (let index = 0; index < data.testInput.length; index++) {
            const result = indicator.update(data.testInput[index]);
            const actual: Record<string, number> = { "VALUE": result.value, "SHORT_SCORE": result.shortScore, "LONG_SCORE": result.longScore, "SHORT_PERIOD": result.shortPeriod, "LONG_PERIOD": result.longPeriod };
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
        const parameters = { ...defaultParams(), ...{ "genuineThreshold": 100.0, "longMax": 80, "longMin": 31, "phaseCycles": 2.0, "shortMax": 30, "shortMin": 5, "smoothing": 1600.0, "window": 200 } };
        const indicator = new PhasingOscillator(parameters);
        const expected: Record<string, number[]> = { "VALUE": data.expectedW200_S05_S130_L031_L180_HP1600_G100_F2_VALUE, "SHORT_SCORE": data.expectedW200_S05_S130_L031_L180_HP1600_G100_F2_SHORT_SCORE, "LONG_SCORE": data.expectedW200_S05_S130_L031_L180_HP1600_G100_F2_LONG_SCORE, "SHORT_PERIOD": data.expectedW200_S05_S130_L031_L180_HP1600_G100_F2_SHORT_PERIOD, "LONG_PERIOD": data.expectedW200_S05_S130_L031_L180_HP1600_G100_F2_LONG_PERIOD };
        for (let index = 0; index < data.testInput.length; index++) {
            const result = indicator.update(data.testInput[index]);
            const actual: Record<string, number> = { "VALUE": result.value, "SHORT_SCORE": result.shortScore, "LONG_SCORE": result.longScore, "SHORT_PERIOD": result.shortPeriod, "LONG_PERIOD": result.longPeriod };
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
        const indicator = new PhasingOscillator();
        indicator.update(NaN);
        expect(indicator.isPrimed()).toBeFalse();
        expect(() => indicator.update(Infinity)).toThrow();
    });
});
