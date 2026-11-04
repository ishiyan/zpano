import { HodrickPrescottSmoothingAndDetrending } from "./hodrick-prescott-smoothing-and-detrending.js";
import { defaultParams } from "./params.js";
import * as data from "./testdata.js";
describe("Lars von Thienen HodrickPrescottSmoothingAndDetrending", () => {
    it("matches prepared case 0", () => {
        const parameters = { ...defaultParams(), ...{ "window": 50, "smoothing": 1600.0 } };
        const indicator = new HodrickPrescottSmoothingAndDetrending(parameters);
        const expected: Record<string, number[]> = { "TREND": data.expectedW50_L1600_TREND, "CYCLE": data.expectedW50_L1600_CYCLE };
        for (let index = 0; index < data.testInput.length; index++) {
            const result = indicator.update(data.testInput[index]);
            const actual: Record<string, number> = { "TREND": result.trend, "CYCLE": result.cycle };
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
        const parameters = { ...defaultParams(), ...{ "window": 6, "smoothing": 1600.0 } };
        const indicator = new HodrickPrescottSmoothingAndDetrending(parameters);
        const expected: Record<string, number[]> = { "TREND": data.expectedW6_L1600_TREND, "CYCLE": data.expectedW6_L1600_CYCLE };
        for (let index = 0; index < data.testInput.length; index++) {
            const result = indicator.update(data.testInput[index]);
            const actual: Record<string, number> = { "TREND": result.trend, "CYCLE": result.cycle };
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
        const parameters = { ...defaultParams(), ...{ "window": 20, "smoothing": 0.0 } };
        const indicator = new HodrickPrescottSmoothingAndDetrending(parameters);
        const expected: Record<string, number[]> = { "TREND": data.expectedW20_L0_TREND, "CYCLE": data.expectedW20_L0_CYCLE };
        for (let index = 0; index < data.testInput.length; index++) {
            const result = indicator.update(data.testInput[index]);
            const actual: Record<string, number> = { "TREND": result.trend, "CYCLE": result.cycle };
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
        const parameters = { ...defaultParams(), ...{ "window": 20, "smoothing": 1.0 } };
        const indicator = new HodrickPrescottSmoothingAndDetrending(parameters);
        const expected: Record<string, number[]> = { "TREND": data.expectedW20_L1_TREND, "CYCLE": data.expectedW20_L1_CYCLE };
        for (let index = 0; index < data.testInput.length; index++) {
            const result = indicator.update(data.testInput[index]);
            const actual: Record<string, number> = { "TREND": result.trend, "CYCLE": result.cycle };
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
        const parameters = { ...defaultParams(), ...{ "window": 50, "smoothing": 160.0 } };
        const indicator = new HodrickPrescottSmoothingAndDetrending(parameters);
        const expected: Record<string, number[]> = { "TREND": data.expectedW50_L160_TREND, "CYCLE": data.expectedW50_L160_CYCLE };
        for (let index = 0; index < data.testInput.length; index++) {
            const result = indicator.update(data.testInput[index]);
            const actual: Record<string, number> = { "TREND": result.trend, "CYCLE": result.cycle };
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
        const parameters = { ...defaultParams(), ...{ "window": 100, "smoothing": 1600.0 } };
        const indicator = new HodrickPrescottSmoothingAndDetrending(parameters);
        const expected: Record<string, number[]> = { "TREND": data.expectedW100_L1600_TREND, "CYCLE": data.expectedW100_L1600_CYCLE };
        for (let index = 0; index < data.testInput.length; index++) {
            const result = indicator.update(data.testInput[index]);
            const actual: Record<string, number> = { "TREND": result.trend, "CYCLE": result.cycle };
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
        const parameters = { ...defaultParams(), ...{ "window": 100, "smoothing": 100000.0 } };
        const indicator = new HodrickPrescottSmoothingAndDetrending(parameters);
        const expected: Record<string, number[]> = { "TREND": data.expectedW100_L100000_TREND, "CYCLE": data.expectedW100_L100000_CYCLE };
        for (let index = 0; index < data.testInput.length; index++) {
            const result = indicator.update(data.testInput[index]);
            const actual: Record<string, number> = { "TREND": result.trend, "CYCLE": result.cycle };
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
        const parameters = { ...defaultParams(), ...{ "window": 252, "smoothing": 1600.0 } };
        const indicator = new HodrickPrescottSmoothingAndDetrending(parameters);
        const expected: Record<string, number[]> = { "TREND": data.expectedW252_L1600_TREND, "CYCLE": data.expectedW252_L1600_CYCLE };
        for (let index = 0; index < data.testInput.length; index++) {
            const result = indicator.update(data.testInput[index]);
            const actual: Record<string, number> = { "TREND": result.trend, "CYCLE": result.cycle };
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
        const parameters = { ...defaultParams(), ...{ "window": 253, "smoothing": 1600.0 } };
        const indicator = new HodrickPrescottSmoothingAndDetrending(parameters);
        const expected: Record<string, number[]> = { "TREND": data.expectedW253_L1600_TREND, "CYCLE": data.expectedW253_L1600_CYCLE };
        for (let index = 0; index < data.testInput.length; index++) {
            const result = indicator.update(data.testInput[index]);
            const actual: Record<string, number> = { "TREND": result.trend, "CYCLE": result.cycle };
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
        const indicator = new HodrickPrescottSmoothingAndDetrending();
        indicator.update(NaN);
        expect(indicator.isPrimed()).toBeFalse();
        expect(() => indicator.update(Infinity)).toThrow();
    });
});
