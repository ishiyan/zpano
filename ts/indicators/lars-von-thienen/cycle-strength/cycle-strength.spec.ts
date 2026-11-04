import { CycleStrength } from "./cycle-strength.js";
import { defaultParams } from "./params.js";
import * as data from "./testdata.js";
describe("Lars von Thienen CycleStrength", () => {
    it("matches prepared case 0", () => {
        const parameters = { ...defaultParams(), ...{ "window": 100, "period": 20.0, "smoothing": 1600.0 } };
        const indicator = new CycleStrength(parameters);
        const expected: Record<string, number[]> = { "STRENGTH": data.expectedW100_P20_L1600_STRENGTH, "AMPLITUDE": data.expectedW100_P20_L1600_AMPLITUDE, "PHASE": data.expectedW100_P20_L1600_PHASE };
        for (let index = 0; index < data.testInput.length; index++) {
            const result = indicator.update(data.testInput[index]);
            const actual: Record<string, number> = { "STRENGTH": result.strength, "AMPLITUDE": result.amplitude, "PHASE": result.phase };
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
        const parameters = { ...defaultParams(), ...{ "window": 100, "period": 20.0, "smoothing": 0.0 } };
        const indicator = new CycleStrength(parameters);
        const expected: Record<string, number[]> = { "STRENGTH": data.expectedW100_P20_L0_STRENGTH, "AMPLITUDE": data.expectedW100_P20_L0_AMPLITUDE, "PHASE": data.expectedW100_P20_L0_PHASE };
        for (let index = 0; index < data.testInput.length; index++) {
            const result = indicator.update(data.testInput[index]);
            const actual: Record<string, number> = { "STRENGTH": result.strength, "AMPLITUDE": result.amplitude, "PHASE": result.phase };
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
        const parameters = { ...defaultParams(), ...{ "window": 60, "period": 10.0, "smoothing": 1600.0 } };
        const indicator = new CycleStrength(parameters);
        const expected: Record<string, number[]> = { "STRENGTH": data.expectedW60_P10_L1600_STRENGTH, "AMPLITUDE": data.expectedW60_P10_L1600_AMPLITUDE, "PHASE": data.expectedW60_P10_L1600_PHASE };
        for (let index = 0; index < data.testInput.length; index++) {
            const result = indicator.update(data.testInput[index]);
            const actual: Record<string, number> = { "STRENGTH": result.strength, "AMPLITUDE": result.amplitude, "PHASE": result.phase };
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
        const parameters = { ...defaultParams(), ...{ "window": 200, "period": 40.0, "smoothing": 1600.0 } };
        const indicator = new CycleStrength(parameters);
        const expected: Record<string, number[]> = { "STRENGTH": data.expectedW200_P40_L1600_STRENGTH, "AMPLITUDE": data.expectedW200_P40_L1600_AMPLITUDE, "PHASE": data.expectedW200_P40_L1600_PHASE };
        for (let index = 0; index < data.testInput.length; index++) {
            const result = indicator.update(data.testInput[index]);
            const actual: Record<string, number> = { "STRENGTH": result.strength, "AMPLITUDE": result.amplitude, "PHASE": result.phase };
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
        const parameters = { ...defaultParams(), ...{ "window": 100, "period": 20.5, "smoothing": 1600.0 } };
        const indicator = new CycleStrength(parameters);
        const expected: Record<string, number[]> = { "STRENGTH": data.expectedW100_P20_5_L1600_STRENGTH, "AMPLITUDE": data.expectedW100_P20_5_L1600_AMPLITUDE, "PHASE": data.expectedW100_P20_5_L1600_PHASE };
        for (let index = 0; index < data.testInput.length; index++) {
            const result = indicator.update(data.testInput[index]);
            const actual: Record<string, number> = { "STRENGTH": result.strength, "AMPLITUDE": result.amplitude, "PHASE": result.phase };
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
        const parameters = { ...defaultParams(), ...{ "window": 100, "period": 20.0, "smoothing": 16000.0 } };
        const indicator = new CycleStrength(parameters);
        const expected: Record<string, number[]> = { "STRENGTH": data.expectedW100_P20_L16000_STRENGTH, "AMPLITUDE": data.expectedW100_P20_L16000_AMPLITUDE, "PHASE": data.expectedW100_P20_L16000_PHASE };
        for (let index = 0; index < data.testInput.length; index++) {
            const result = indicator.update(data.testInput[index]);
            const actual: Record<string, number> = { "STRENGTH": result.strength, "AMPLITUDE": result.amplitude, "PHASE": result.phase };
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
        const parameters = { ...defaultParams(), ...{ "window": 20, "period": 40.0, "smoothing": 1600.0 } };
        const indicator = new CycleStrength(parameters);
        const expected: Record<string, number[]> = { "STRENGTH": data.expectedW20_P40_L1600_STRENGTH, "AMPLITUDE": data.expectedW20_P40_L1600_AMPLITUDE, "PHASE": data.expectedW20_P40_L1600_PHASE };
        for (let index = 0; index < data.testInput.length; index++) {
            const result = indicator.update(data.testInput[index]);
            const actual: Record<string, number> = { "STRENGTH": result.strength, "AMPLITUDE": result.amplitude, "PHASE": result.phase };
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
        const parameters = { ...defaultParams(), ...{ "window": 252, "period": 20.0, "smoothing": 1600.0 } };
        const indicator = new CycleStrength(parameters);
        const expected: Record<string, number[]> = { "STRENGTH": data.expectedW252_P20_L1600_STRENGTH, "AMPLITUDE": data.expectedW252_P20_L1600_AMPLITUDE, "PHASE": data.expectedW252_P20_L1600_PHASE };
        for (let index = 0; index < data.testInput.length; index++) {
            const result = indicator.update(data.testInput[index]);
            const actual: Record<string, number> = { "STRENGTH": result.strength, "AMPLITUDE": result.amplitude, "PHASE": result.phase };
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
        const indicator = new CycleStrength();
        indicator.update(NaN);
        expect(indicator.isPrimed()).toBeFalse();
        expect(() => indicator.update(Infinity)).toThrow();
    });
});
