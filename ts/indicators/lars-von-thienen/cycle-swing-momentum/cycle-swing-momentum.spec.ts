import { CycleSwingMomentum } from "./cycle-swing-momentum.js";
import { defaultParams } from "./params.js";
import * as data from "./testdata.js";
describe("Lars von Thienen CycleSwingMomentum", () => {
    it("matches prepared case 0", () => {
        const parameters = { ...defaultParams(), ...{ "window": 50, "fastSmoothing": 160.0, "slowSmoothing": 1600.0, "bandMemory": 34, "leveling": 10.0 } };
        const indicator = new CycleSwingMomentum(parameters);
        const expected: Record<string, number[]> = { "VALUE": data.expectedW50_F160_S1600_M34_Q10_VALUE, "LOWER": data.expectedW50_F160_S1600_M34_Q10_LOWER, "UPPER": data.expectedW50_F160_S1600_M34_Q10_UPPER, "SCORE": data.expectedW50_F160_S1600_M34_Q10_SCORE };
        for (let index = 0; index < data.testInput.length; index++) {
            const result = indicator.update(data.testInput[index]);
            const actual: Record<string, number> = { "VALUE": result.value, "LOWER": result.lower, "UPPER": result.upper, "SCORE": result.score };
            for (const [field, values] of Object.entries(expected)) {
                const value = actual[field];
                if (Number.isNaN(values[index]))
                    expect(Number.isNaN(value)).withContext(`bar ${index} ${field}`).toBeTrue();
                else
                    expect(Math.abs(value - values[index])).withContext(`bar ${index} ${field}: ${value} vs ${values[index]}`).toBeLessThanOrEqual(Math.max(1e-13, Math.abs(values[index]) * 5e-15));
            }
        }
        indicator.reset();
        expect(indicator.isPrimed()).toBeFalse();
    });
    it("matches prepared case 1", () => {
        const parameters = { ...defaultParams(), ...{ "window": 6, "fastSmoothing": 16.0, "slowSmoothing": 160.0, "bandMemory": 5, "leveling": 10.0 } };
        const indicator = new CycleSwingMomentum(parameters);
        const expected: Record<string, number[]> = { "VALUE": data.expectedW6_F16_S160_M5_Q10_VALUE, "LOWER": data.expectedW6_F16_S160_M5_Q10_LOWER, "UPPER": data.expectedW6_F16_S160_M5_Q10_UPPER, "SCORE": data.expectedW6_F16_S160_M5_Q10_SCORE };
        for (let index = 0; index < data.testInput.length; index++) {
            const result = indicator.update(data.testInput[index]);
            const actual: Record<string, number> = { "VALUE": result.value, "LOWER": result.lower, "UPPER": result.upper, "SCORE": result.score };
            for (const [field, values] of Object.entries(expected)) {
                const value = actual[field];
                if (Number.isNaN(values[index]))
                    expect(Number.isNaN(value)).withContext(`bar ${index} ${field}`).toBeTrue();
                else
                    expect(Math.abs(value - values[index])).withContext(`bar ${index} ${field}: ${value} vs ${values[index]}`).toBeLessThanOrEqual(Math.max(1e-13, Math.abs(values[index]) * 5e-15));
            }
        }
        indicator.reset();
        expect(indicator.isPrimed()).toBeFalse();
    });
    it("matches prepared case 2", () => {
        const parameters = { ...defaultParams(), ...{ "window": 20, "fastSmoothing": 160.0, "slowSmoothing": 1600.0, "bandMemory": 20, "leveling": 10.0 } };
        const indicator = new CycleSwingMomentum(parameters);
        const expected: Record<string, number[]> = { "VALUE": data.expectedW20_F160_S1600_M20_Q10_VALUE, "LOWER": data.expectedW20_F160_S1600_M20_Q10_LOWER, "UPPER": data.expectedW20_F160_S1600_M20_Q10_UPPER, "SCORE": data.expectedW20_F160_S1600_M20_Q10_SCORE };
        for (let index = 0; index < data.testInput.length; index++) {
            const result = indicator.update(data.testInput[index]);
            const actual: Record<string, number> = { "VALUE": result.value, "LOWER": result.lower, "UPPER": result.upper, "SCORE": result.score };
            for (const [field, values] of Object.entries(expected)) {
                const value = actual[field];
                if (Number.isNaN(values[index]))
                    expect(Number.isNaN(value)).withContext(`bar ${index} ${field}`).toBeTrue();
                else
                    expect(Math.abs(value - values[index])).withContext(`bar ${index} ${field}: ${value} vs ${values[index]}`).toBeLessThanOrEqual(Math.max(1e-13, Math.abs(values[index]) * 5e-15));
            }
        }
        indicator.reset();
        expect(indicator.isPrimed()).toBeFalse();
    });
    it("matches prepared case 3", () => {
        const parameters = { ...defaultParams(), ...{ "window": 100, "fastSmoothing": 160.0, "slowSmoothing": 1600.0, "bandMemory": 34, "leveling": 10.0 } };
        const indicator = new CycleSwingMomentum(parameters);
        const expected: Record<string, number[]> = { "VALUE": data.expectedW100_F160_S1600_M34_Q10_VALUE, "LOWER": data.expectedW100_F160_S1600_M34_Q10_LOWER, "UPPER": data.expectedW100_F160_S1600_M34_Q10_UPPER, "SCORE": data.expectedW100_F160_S1600_M34_Q10_SCORE };
        for (let index = 0; index < data.testInput.length; index++) {
            const result = indicator.update(data.testInput[index]);
            const actual: Record<string, number> = { "VALUE": result.value, "LOWER": result.lower, "UPPER": result.upper, "SCORE": result.score };
            for (const [field, values] of Object.entries(expected)) {
                const value = actual[field];
                if (Number.isNaN(values[index]))
                    expect(Number.isNaN(value)).withContext(`bar ${index} ${field}`).toBeTrue();
                else
                    expect(Math.abs(value - values[index])).withContext(`bar ${index} ${field}: ${value} vs ${values[index]}`).toBeLessThanOrEqual(Math.max(1e-13, Math.abs(values[index]) * 5e-15));
            }
        }
        indicator.reset();
        expect(indicator.isPrimed()).toBeFalse();
    });
    it("matches prepared case 4", () => {
        const parameters = { ...defaultParams(), ...{ "window": 50, "fastSmoothing": 0.0, "slowSmoothing": 1600.0, "bandMemory": 34, "leveling": 10.0 } };
        const indicator = new CycleSwingMomentum(parameters);
        const expected: Record<string, number[]> = { "VALUE": data.expectedW50_F0_S1600_M34_Q10_VALUE, "LOWER": data.expectedW50_F0_S1600_M34_Q10_LOWER, "UPPER": data.expectedW50_F0_S1600_M34_Q10_UPPER, "SCORE": data.expectedW50_F0_S1600_M34_Q10_SCORE };
        for (let index = 0; index < data.testInput.length; index++) {
            const result = indicator.update(data.testInput[index]);
            const actual: Record<string, number> = { "VALUE": result.value, "LOWER": result.lower, "UPPER": result.upper, "SCORE": result.score };
            for (const [field, values] of Object.entries(expected)) {
                const value = actual[field];
                if (Number.isNaN(values[index]))
                    expect(Number.isNaN(value)).withContext(`bar ${index} ${field}`).toBeTrue();
                else
                    expect(Math.abs(value - values[index])).withContext(`bar ${index} ${field}: ${value} vs ${values[index]}`).toBeLessThanOrEqual(Math.max(1e-13, Math.abs(values[index]) * 5e-15));
            }
        }
        indicator.reset();
        expect(indicator.isPrimed()).toBeFalse();
    });
    it("matches prepared case 5", () => {
        const parameters = { ...defaultParams(), ...{ "window": 50, "fastSmoothing": 1600.0, "slowSmoothing": 16000.0, "bandMemory": 34, "leveling": 10.0 } };
        const indicator = new CycleSwingMomentum(parameters);
        const expected: Record<string, number[]> = { "VALUE": data.expectedW50_F1600_S16000_M34_Q10_VALUE, "LOWER": data.expectedW50_F1600_S16000_M34_Q10_LOWER, "UPPER": data.expectedW50_F1600_S16000_M34_Q10_UPPER, "SCORE": data.expectedW50_F1600_S16000_M34_Q10_SCORE };
        for (let index = 0; index < data.testInput.length; index++) {
            const result = indicator.update(data.testInput[index]);
            const actual: Record<string, number> = { "VALUE": result.value, "LOWER": result.lower, "UPPER": result.upper, "SCORE": result.score };
            for (const [field, values] of Object.entries(expected)) {
                const value = actual[field];
                if (Number.isNaN(values[index]))
                    expect(Number.isNaN(value)).withContext(`bar ${index} ${field}`).toBeTrue();
                else
                    expect(Math.abs(value - values[index])).withContext(`bar ${index} ${field}: ${value} vs ${values[index]}`).toBeLessThanOrEqual(Math.max(1e-13, Math.abs(values[index]) * 5e-15));
            }
        }
        indicator.reset();
        expect(indicator.isPrimed()).toBeFalse();
    });
    it("matches prepared case 6", () => {
        const parameters = { ...defaultParams(), ...{ "window": 50, "fastSmoothing": 160.0, "slowSmoothing": 1600.0, "bandMemory": 34, "leveling": 5.0 } };
        const indicator = new CycleSwingMomentum(parameters);
        const expected: Record<string, number[]> = { "VALUE": data.expectedW50_F160_S1600_M34_Q5_VALUE, "LOWER": data.expectedW50_F160_S1600_M34_Q5_LOWER, "UPPER": data.expectedW50_F160_S1600_M34_Q5_UPPER, "SCORE": data.expectedW50_F160_S1600_M34_Q5_SCORE };
        for (let index = 0; index < data.testInput.length; index++) {
            const result = indicator.update(data.testInput[index]);
            const actual: Record<string, number> = { "VALUE": result.value, "LOWER": result.lower, "UPPER": result.upper, "SCORE": result.score };
            for (const [field, values] of Object.entries(expected)) {
                const value = actual[field];
                if (Number.isNaN(values[index]))
                    expect(Number.isNaN(value)).withContext(`bar ${index} ${field}`).toBeTrue();
                else
                    expect(Math.abs(value - values[index])).withContext(`bar ${index} ${field}: ${value} vs ${values[index]}`).toBeLessThanOrEqual(Math.max(1e-13, Math.abs(values[index]) * 5e-15));
            }
        }
        indicator.reset();
        expect(indicator.isPrimed()).toBeFalse();
    });
    it("matches prepared case 7", () => {
        const parameters = { ...defaultParams(), ...{ "window": 252, "fastSmoothing": 160.0, "slowSmoothing": 1600.0, "bandMemory": 34, "leveling": 10.0 } };
        const indicator = new CycleSwingMomentum(parameters);
        const expected: Record<string, number[]> = { "VALUE": data.expectedW252_F160_S1600_M34_Q10_VALUE, "LOWER": data.expectedW252_F160_S1600_M34_Q10_LOWER, "UPPER": data.expectedW252_F160_S1600_M34_Q10_UPPER, "SCORE": data.expectedW252_F160_S1600_M34_Q10_SCORE };
        for (let index = 0; index < data.testInput.length; index++) {
            const result = indicator.update(data.testInput[index]);
            const actual: Record<string, number> = { "VALUE": result.value, "LOWER": result.lower, "UPPER": result.upper, "SCORE": result.score };
            for (const [field, values] of Object.entries(expected)) {
                const value = actual[field];
                if (Number.isNaN(values[index]))
                    expect(Number.isNaN(value)).withContext(`bar ${index} ${field}`).toBeTrue();
                else
                    expect(Math.abs(value - values[index])).withContext(`bar ${index} ${field}: ${value} vs ${values[index]}`).toBeLessThanOrEqual(Math.max(1e-13, Math.abs(values[index]) * 5e-15));
            }
        }
        indicator.reset();
        expect(indicator.isPrimed()).toBeFalse();
    });
    it("skips NaN and rejects invalid observations before priming", () => {
        const indicator = new CycleSwingMomentum();
        indicator.update(NaN);
        expect(indicator.isPrimed()).toBeFalse();
        expect(() => indicator.update(Infinity)).toThrow();
    });
});
