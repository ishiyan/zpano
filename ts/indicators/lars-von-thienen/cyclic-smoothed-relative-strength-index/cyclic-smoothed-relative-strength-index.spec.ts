import { CyclicSmoothedRelativeStrengthIndex } from "./cyclic-smoothed-relative-strength-index.js";
import { defaultParams } from "./params.js";
import * as data from "./testdata.js";
describe("Lars von Thienen CyclicSmoothedRelativeStrengthIndex", () => {
    it("matches prepared case 0", () => {
        const parameters = { ...defaultParams(), ...{ "dominantCycle": 20, "vibration": 10, "leveling": 10.0 } };
        const indicator = new CyclicSmoothedRelativeStrengthIndex(parameters);
        const expected: Record<string, number[]> = { "VALUE": data.expectedD20_V10_Q10_VALUE, "LOWER": data.expectedD20_V10_Q10_LOWER, "UPPER": data.expectedD20_V10_Q10_UPPER };
        for (let index = 0; index < data.testInput.length; index++) {
            const result = indicator.update(data.testInput[index]);
            const actual: Record<string, number> = { "VALUE": result.value, "LOWER": result.lower, "UPPER": result.upper };
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
        const parameters = { ...defaultParams(), ...{ "dominantCycle": 10, "vibration": 10, "leveling": 10.0 } };
        const indicator = new CyclicSmoothedRelativeStrengthIndex(parameters);
        const expected: Record<string, number[]> = { "VALUE": data.expectedD10_V10_Q10_VALUE, "LOWER": data.expectedD10_V10_Q10_LOWER, "UPPER": data.expectedD10_V10_Q10_UPPER };
        for (let index = 0; index < data.testInput.length; index++) {
            const result = indicator.update(data.testInput[index]);
            const actual: Record<string, number> = { "VALUE": result.value, "LOWER": result.lower, "UPPER": result.upper };
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
        const parameters = { ...defaultParams(), ...{ "dominantCycle": 21, "vibration": 10, "leveling": 10.0 } };
        const indicator = new CyclicSmoothedRelativeStrengthIndex(parameters);
        const expected: Record<string, number[]> = { "VALUE": data.expectedD21_V10_Q10_VALUE, "LOWER": data.expectedD21_V10_Q10_LOWER, "UPPER": data.expectedD21_V10_Q10_UPPER };
        for (let index = 0; index < data.testInput.length; index++) {
            const result = indicator.update(data.testInput[index]);
            const actual: Record<string, number> = { "VALUE": result.value, "LOWER": result.lower, "UPPER": result.upper };
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
        const parameters = { ...defaultParams(), ...{ "dominantCycle": 40, "vibration": 10, "leveling": 10.0 } };
        const indicator = new CyclicSmoothedRelativeStrengthIndex(parameters);
        const expected: Record<string, number[]> = { "VALUE": data.expectedD40_V10_Q10_VALUE, "LOWER": data.expectedD40_V10_Q10_LOWER, "UPPER": data.expectedD40_V10_Q10_UPPER };
        for (let index = 0; index < data.testInput.length; index++) {
            const result = indicator.update(data.testInput[index]);
            const actual: Record<string, number> = { "VALUE": result.value, "LOWER": result.lower, "UPPER": result.upper };
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
        const parameters = { ...defaultParams(), ...{ "dominantCycle": 100, "vibration": 10, "leveling": 10.0 } };
        const indicator = new CyclicSmoothedRelativeStrengthIndex(parameters);
        const expected: Record<string, number[]> = { "VALUE": data.expectedD100_V10_Q10_VALUE, "LOWER": data.expectedD100_V10_Q10_LOWER, "UPPER": data.expectedD100_V10_Q10_UPPER };
        for (let index = 0; index < data.testInput.length; index++) {
            const result = indicator.update(data.testInput[index]);
            const actual: Record<string, number> = { "VALUE": result.value, "LOWER": result.lower, "UPPER": result.upper };
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
        const parameters = { ...defaultParams(), ...{ "dominantCycle": 20, "vibration": 3, "leveling": 10.0 } };
        const indicator = new CyclicSmoothedRelativeStrengthIndex(parameters);
        const expected: Record<string, number[]> = { "VALUE": data.expectedD20_V3_Q10_VALUE, "LOWER": data.expectedD20_V3_Q10_LOWER, "UPPER": data.expectedD20_V3_Q10_UPPER };
        for (let index = 0; index < data.testInput.length; index++) {
            const result = indicator.update(data.testInput[index]);
            const actual: Record<string, number> = { "VALUE": result.value, "LOWER": result.lower, "UPPER": result.upper };
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
        const parameters = { ...defaultParams(), ...{ "dominantCycle": 20, "vibration": 6, "leveling": 10.0 } };
        const indicator = new CyclicSmoothedRelativeStrengthIndex(parameters);
        const expected: Record<string, number[]> = { "VALUE": data.expectedD20_V6_Q10_VALUE, "LOWER": data.expectedD20_V6_Q10_LOWER, "UPPER": data.expectedD20_V6_Q10_UPPER };
        for (let index = 0; index < data.testInput.length; index++) {
            const result = indicator.update(data.testInput[index]);
            const actual: Record<string, number> = { "VALUE": result.value, "LOWER": result.lower, "UPPER": result.upper };
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
        const parameters = { ...defaultParams(), ...{ "dominantCycle": 20, "vibration": 10, "leveling": 5.0 } };
        const indicator = new CyclicSmoothedRelativeStrengthIndex(parameters);
        const expected: Record<string, number[]> = { "VALUE": data.expectedD20_V10_Q5_VALUE, "LOWER": data.expectedD20_V10_Q5_LOWER, "UPPER": data.expectedD20_V10_Q5_UPPER };
        for (let index = 0; index < data.testInput.length; index++) {
            const result = indicator.update(data.testInput[index]);
            const actual: Record<string, number> = { "VALUE": result.value, "LOWER": result.lower, "UPPER": result.upper };
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
    it("matches prepared case 8", () => {
        const parameters = { ...defaultParams(), ...{ "dominantCycle": 20, "vibration": 10, "leveling": 20.0 } };
        const indicator = new CyclicSmoothedRelativeStrengthIndex(parameters);
        const expected: Record<string, number[]> = { "VALUE": data.expectedD20_V10_Q20_VALUE, "LOWER": data.expectedD20_V10_Q20_LOWER, "UPPER": data.expectedD20_V10_Q20_UPPER };
        for (let index = 0; index < data.testInput.length; index++) {
            const result = indicator.update(data.testInput[index]);
            const actual: Record<string, number> = { "VALUE": result.value, "LOWER": result.lower, "UPPER": result.upper };
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
        const indicator = new CyclicSmoothedRelativeStrengthIndex();
        indicator.update(NaN);
        expect(indicator.isPrimed()).toBeFalse();
        expect(() => indicator.update(Infinity)).toThrow();
    });
});
