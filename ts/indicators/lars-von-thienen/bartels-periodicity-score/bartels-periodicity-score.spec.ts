import { BartelsPeriodicityScore } from "./bartels-periodicity-score.js";
import { defaultParams } from "./params.js";
import * as data from "./testdata.js";
describe("Lars von Thienen BartelsPeriodicityScore", () => {
    it("matches prepared case 0", () => {
        const parameters = { ...defaultParams(), ...{ "period": 20, "segments": 5, "smoothing": 0.0 } };
        const indicator = new BartelsPeriodicityScore(parameters);
        const expected: Record<string, number[]> = { "SCORE": data.expectedP20_M5_L0_SCORE, "GENUINE_PERCENT": data.expectedP20_M5_L0_GENUINE_PERCENT, "SEGMENTS": data.expectedP20_M5_L0_SEGMENTS };
        for (let index = 0; index < data.testInput.length; index++) {
            const result = indicator.update(data.testInput[index]);
            const actual: Record<string, number> = { "SCORE": result.score, "GENUINE_PERCENT": result.genuinePercent, "SEGMENTS": result.segments };
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
        const parameters = { ...defaultParams(), ...{ "period": 3, "segments": 2, "smoothing": 0.0 } };
        const indicator = new BartelsPeriodicityScore(parameters);
        const expected: Record<string, number[]> = { "SCORE": data.expectedP3_M2_L0_SCORE, "GENUINE_PERCENT": data.expectedP3_M2_L0_GENUINE_PERCENT, "SEGMENTS": data.expectedP3_M2_L0_SEGMENTS };
        for (let index = 0; index < data.testInput.length; index++) {
            const result = indicator.update(data.testInput[index]);
            const actual: Record<string, number> = { "SCORE": result.score, "GENUINE_PERCENT": result.genuinePercent, "SEGMENTS": result.segments };
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
        const parameters = { ...defaultParams(), ...{ "period": 10, "segments": 5, "smoothing": 0.0 } };
        const indicator = new BartelsPeriodicityScore(parameters);
        const expected: Record<string, number[]> = { "SCORE": data.expectedP10_M5_L0_SCORE, "GENUINE_PERCENT": data.expectedP10_M5_L0_GENUINE_PERCENT, "SEGMENTS": data.expectedP10_M5_L0_SEGMENTS };
        for (let index = 0; index < data.testInput.length; index++) {
            const result = indicator.update(data.testInput[index]);
            const actual: Record<string, number> = { "SCORE": result.score, "GENUINE_PERCENT": result.genuinePercent, "SEGMENTS": result.segments };
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
        const parameters = { ...defaultParams(), ...{ "period": 20, "segments": 2, "smoothing": 0.0 } };
        const indicator = new BartelsPeriodicityScore(parameters);
        const expected: Record<string, number[]> = { "SCORE": data.expectedP20_M2_L0_SCORE, "GENUINE_PERCENT": data.expectedP20_M2_L0_GENUINE_PERCENT, "SEGMENTS": data.expectedP20_M2_L0_SEGMENTS };
        for (let index = 0; index < data.testInput.length; index++) {
            const result = indicator.update(data.testInput[index]);
            const actual: Record<string, number> = { "SCORE": result.score, "GENUINE_PERCENT": result.genuinePercent, "SEGMENTS": result.segments };
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
    it("matches prepared case 4", () => {
        const parameters = { ...defaultParams(), ...{ "period": 20, "segments": 5, "smoothing": 1600.0 } };
        const indicator = new BartelsPeriodicityScore(parameters);
        const expected: Record<string, number[]> = { "SCORE": data.expectedP20_M5_L1600_SCORE, "GENUINE_PERCENT": data.expectedP20_M5_L1600_GENUINE_PERCENT, "SEGMENTS": data.expectedP20_M5_L1600_SEGMENTS };
        for (let index = 0; index < data.testInput.length; index++) {
            const result = indicator.update(data.testInput[index]);
            const actual: Record<string, number> = { "SCORE": result.score, "GENUINE_PERCENT": result.genuinePercent, "SEGMENTS": result.segments };
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
    it("matches prepared case 5", () => {
        const parameters = { ...defaultParams(), ...{ "period": 40, "segments": 5, "smoothing": 0.0 } };
        const indicator = new BartelsPeriodicityScore(parameters);
        const expected: Record<string, number[]> = { "SCORE": data.expectedP40_M5_L0_SCORE, "GENUINE_PERCENT": data.expectedP40_M5_L0_GENUINE_PERCENT, "SEGMENTS": data.expectedP40_M5_L0_SEGMENTS };
        for (let index = 0; index < data.testInput.length; index++) {
            const result = indicator.update(data.testInput[index]);
            const actual: Record<string, number> = { "SCORE": result.score, "GENUINE_PERCENT": result.genuinePercent, "SEGMENTS": result.segments };
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
    it("matches prepared case 6", () => {
        const parameters = { ...defaultParams(), ...{ "period": 42, "segments": 6, "smoothing": 0.0 } };
        const indicator = new BartelsPeriodicityScore(parameters);
        const expected: Record<string, number[]> = { "SCORE": data.expectedP42_M6_L0_SCORE, "GENUINE_PERCENT": data.expectedP42_M6_L0_GENUINE_PERCENT, "SEGMENTS": data.expectedP42_M6_L0_SEGMENTS };
        for (let index = 0; index < data.testInput.length; index++) {
            const result = indicator.update(data.testInput[index]);
            const actual: Record<string, number> = { "SCORE": result.score, "GENUINE_PERCENT": result.genuinePercent, "SEGMENTS": result.segments };
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
    it("matches prepared case 7", () => {
        const parameters = { ...defaultParams(), ...{ "period": 60, "segments": 5, "smoothing": 0.0 } };
        const indicator = new BartelsPeriodicityScore(parameters);
        const expected: Record<string, number[]> = { "SCORE": data.expectedP60_M5_L0_SCORE, "GENUINE_PERCENT": data.expectedP60_M5_L0_GENUINE_PERCENT, "SEGMENTS": data.expectedP60_M5_L0_SEGMENTS };
        for (let index = 0; index < data.testInput.length; index++) {
            const result = indicator.update(data.testInput[index]);
            const actual: Record<string, number> = { "SCORE": result.score, "GENUINE_PERCENT": result.genuinePercent, "SEGMENTS": result.segments };
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
        const indicator = new BartelsPeriodicityScore();
        indicator.update(NaN);
        expect(indicator.isPrimed()).toBeFalse();
        expect(() => indicator.update(Infinity)).toThrow();
    });
});
