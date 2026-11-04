import { CompositeCycle } from "./composite-cycle.js";
import { defaultParams } from "./params.js";
import * as data from "./testdata.js";
describe("Lars von Thienen CompositeCycle", () => {
    it("matches prepared case 0", () => {
        const parameters = { ...defaultParams(), ...{ "window": 100, "periods": [20.0, 40.0], "smoothing": 1600.0 } };
        const indicator = new CompositeCycle(parameters);
        const expected: Record<string, number[]> = { "VALUE": data.expectedW100_P20_40_L1600_VALUE, "SLOPE": data.expectedW100_P20_40_L1600_SLOPE, "NEXT_VALUE": data.expectedW100_P20_40_L1600_NEXT_VALUE };
        for (let index = 0; index < data.testInput.length; index++) {
            const result = indicator.update(data.testInput[index]);
            const actual: Record<string, number> = { "VALUE": result.value, "SLOPE": result.slope, "NEXT_VALUE": result.nextValue };
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
        const parameters = { ...defaultParams(), ...{ "window": 100, "periods": [20.0], "smoothing": 0.0 } };
        const indicator = new CompositeCycle(parameters);
        const expected: Record<string, number[]> = { "VALUE": data.expectedW100_P20_L0_VALUE, "SLOPE": data.expectedW100_P20_L0_SLOPE, "NEXT_VALUE": data.expectedW100_P20_L0_NEXT_VALUE };
        for (let index = 0; index < data.testInput.length; index++) {
            const result = indicator.update(data.testInput[index]);
            const actual: Record<string, number> = { "VALUE": result.value, "SLOPE": result.slope, "NEXT_VALUE": result.nextValue };
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
        const parameters = { ...defaultParams(), ...{ "window": 100, "periods": [20.0, 40.0], "smoothing": 0.0 } };
        const indicator = new CompositeCycle(parameters);
        const expected: Record<string, number[]> = { "VALUE": data.expectedW100_P20_40_L0_VALUE, "SLOPE": data.expectedW100_P20_40_L0_SLOPE, "NEXT_VALUE": data.expectedW100_P20_40_L0_NEXT_VALUE };
        for (let index = 0; index < data.testInput.length; index++) {
            const result = indicator.update(data.testInput[index]);
            const actual: Record<string, number> = { "VALUE": result.value, "SLOPE": result.slope, "NEXT_VALUE": result.nextValue };
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
        const parameters = { ...defaultParams(), ...{ "window": 120, "periods": [10.0, 20.0, 40.0], "smoothing": 1600.0 } };
        const indicator = new CompositeCycle(parameters);
        const expected: Record<string, number[]> = { "VALUE": data.expectedW120_P10_20_40_L1600_VALUE, "SLOPE": data.expectedW120_P10_20_40_L1600_SLOPE, "NEXT_VALUE": data.expectedW120_P10_20_40_L1600_NEXT_VALUE };
        for (let index = 0; index < data.testInput.length; index++) {
            const result = indicator.update(data.testInput[index]);
            const actual: Record<string, number> = { "VALUE": result.value, "SLOPE": result.slope, "NEXT_VALUE": result.nextValue };
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
        const parameters = { ...defaultParams(), ...{ "window": 100, "periods": [17.5, 31.0], "smoothing": 1600.0 } };
        const indicator = new CompositeCycle(parameters);
        const expected: Record<string, number[]> = { "VALUE": data.expectedW100_P17_5_31_L1600_VALUE, "SLOPE": data.expectedW100_P17_5_31_L1600_SLOPE, "NEXT_VALUE": data.expectedW100_P17_5_31_L1600_NEXT_VALUE };
        for (let index = 0; index < data.testInput.length; index++) {
            const result = indicator.update(data.testInput[index]);
            const actual: Record<string, number> = { "VALUE": result.value, "SLOPE": result.slope, "NEXT_VALUE": result.nextValue };
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
        const parameters = { ...defaultParams(), ...{ "window": 200, "periods": [20.0, 40.0], "smoothing": 1600.0 } };
        const indicator = new CompositeCycle(parameters);
        const expected: Record<string, number[]> = { "VALUE": data.expectedW200_P20_40_L1600_VALUE, "SLOPE": data.expectedW200_P20_40_L1600_SLOPE, "NEXT_VALUE": data.expectedW200_P20_40_L1600_NEXT_VALUE };
        for (let index = 0; index < data.testInput.length; index++) {
            const result = indicator.update(data.testInput[index]);
            const actual: Record<string, number> = { "VALUE": result.value, "SLOPE": result.slope, "NEXT_VALUE": result.nextValue };
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
        const parameters = { ...defaultParams(), ...{ "window": 100, "periods": [20.0, 40.0], "smoothing": 160.0 } };
        const indicator = new CompositeCycle(parameters);
        const expected: Record<string, number[]> = { "VALUE": data.expectedW100_P20_40_L160_VALUE, "SLOPE": data.expectedW100_P20_40_L160_SLOPE, "NEXT_VALUE": data.expectedW100_P20_40_L160_NEXT_VALUE };
        for (let index = 0; index < data.testInput.length; index++) {
            const result = indicator.update(data.testInput[index]);
            const actual: Record<string, number> = { "VALUE": result.value, "SLOPE": result.slope, "NEXT_VALUE": result.nextValue };
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
        const parameters = { ...defaultParams(), ...{ "window": 253, "periods": [20.0, 40.0], "smoothing": 1600.0 } };
        const indicator = new CompositeCycle(parameters);
        const expected: Record<string, number[]> = { "VALUE": data.expectedW253_P20_40_L1600_VALUE, "SLOPE": data.expectedW253_P20_40_L1600_SLOPE, "NEXT_VALUE": data.expectedW253_P20_40_L1600_NEXT_VALUE };
        for (let index = 0; index < data.testInput.length; index++) {
            const result = indicator.update(data.testInput[index]);
            const actual: Record<string, number> = { "VALUE": result.value, "SLOPE": result.slope, "NEXT_VALUE": result.nextValue };
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
        const indicator = new CompositeCycle();
        indicator.update(NaN);
        expect(indicator.isPrimed()).toBeFalse();
        expect(() => indicator.update(Infinity)).toThrow();
    });
});
