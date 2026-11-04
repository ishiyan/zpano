import { CyclePolarity } from "./cycle-polarity.js";
import { defaultParams } from "./params.js";
import * as data from "./testdata.js";
describe("Lars von Thienen CyclePolarity", () => {
    it("matches prepared case 0", () => {
        const parameters = { ...defaultParams(), ...{ "window": 100, "shortPeriod": 20.0, "longPeriod": 40.0, "smoothing": 1600.0, "tolerance": 0.0 } };
        const indicator = new CyclePolarity(parameters);
        const expected: Record<string, number[]> = { "POLARITY": data.expectedW100_S20_L40_HP1600_T0_POLARITY, "SHORT_VALUE": data.expectedW100_S20_L40_HP1600_T0_SHORT_VALUE, "LONG_VALUE": data.expectedW100_S20_L40_HP1600_T0_LONG_VALUE, "SHORT_SLOPE": data.expectedW100_S20_L40_HP1600_T0_SHORT_SLOPE, "LONG_SLOPE": data.expectedW100_S20_L40_HP1600_T0_LONG_SLOPE };
        for (let index = 0; index < data.testInput.length; index++) {
            const result = indicator.update(data.testInput[index]);
            const actual: Record<string, number> = { "POLARITY": result.polarity, "SHORT_VALUE": result.shortValue, "LONG_VALUE": result.longValue, "SHORT_SLOPE": result.shortSlope, "LONG_SLOPE": result.longSlope };
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
        const parameters = { ...defaultParams(), ...{ "window": 100, "shortPeriod": 20.0, "longPeriod": 40.0, "smoothing": 0.0, "tolerance": 0.0 } };
        const indicator = new CyclePolarity(parameters);
        const expected: Record<string, number[]> = { "POLARITY": data.expectedW100_S20_L40_HP0_T0_POLARITY, "SHORT_VALUE": data.expectedW100_S20_L40_HP0_T0_SHORT_VALUE, "LONG_VALUE": data.expectedW100_S20_L40_HP0_T0_LONG_VALUE, "SHORT_SLOPE": data.expectedW100_S20_L40_HP0_T0_SHORT_SLOPE, "LONG_SLOPE": data.expectedW100_S20_L40_HP0_T0_LONG_SLOPE };
        for (let index = 0; index < data.testInput.length; index++) {
            const result = indicator.update(data.testInput[index]);
            const actual: Record<string, number> = { "POLARITY": result.polarity, "SHORT_VALUE": result.shortValue, "LONG_VALUE": result.longValue, "SHORT_SLOPE": result.shortSlope, "LONG_SLOPE": result.longSlope };
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
        const parameters = { ...defaultParams(), ...{ "window": 60, "shortPeriod": 10.0, "longPeriod": 20.0, "smoothing": 1600.0, "tolerance": 0.0 } };
        const indicator = new CyclePolarity(parameters);
        const expected: Record<string, number[]> = { "POLARITY": data.expectedW60_S10_L20_HP1600_T0_POLARITY, "SHORT_VALUE": data.expectedW60_S10_L20_HP1600_T0_SHORT_VALUE, "LONG_VALUE": data.expectedW60_S10_L20_HP1600_T0_LONG_VALUE, "SHORT_SLOPE": data.expectedW60_S10_L20_HP1600_T0_SHORT_SLOPE, "LONG_SLOPE": data.expectedW60_S10_L20_HP1600_T0_LONG_SLOPE };
        for (let index = 0; index < data.testInput.length; index++) {
            const result = indicator.update(data.testInput[index]);
            const actual: Record<string, number> = { "POLARITY": result.polarity, "SHORT_VALUE": result.shortValue, "LONG_VALUE": result.longValue, "SHORT_SLOPE": result.shortSlope, "LONG_SLOPE": result.longSlope };
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
        const parameters = { ...defaultParams(), ...{ "window": 200, "shortPeriod": 40.0, "longPeriod": 80.0, "smoothing": 1600.0, "tolerance": 0.0 } };
        const indicator = new CyclePolarity(parameters);
        const expected: Record<string, number[]> = { "POLARITY": data.expectedW200_S40_L80_HP1600_T0_POLARITY, "SHORT_VALUE": data.expectedW200_S40_L80_HP1600_T0_SHORT_VALUE, "LONG_VALUE": data.expectedW200_S40_L80_HP1600_T0_LONG_VALUE, "SHORT_SLOPE": data.expectedW200_S40_L80_HP1600_T0_SHORT_SLOPE, "LONG_SLOPE": data.expectedW200_S40_L80_HP1600_T0_LONG_SLOPE };
        for (let index = 0; index < data.testInput.length; index++) {
            const result = indicator.update(data.testInput[index]);
            const actual: Record<string, number> = { "POLARITY": result.polarity, "SHORT_VALUE": result.shortValue, "LONG_VALUE": result.longValue, "SHORT_SLOPE": result.shortSlope, "LONG_SLOPE": result.longSlope };
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
        const parameters = { ...defaultParams(), ...{ "window": 100, "shortPeriod": 17.5, "longPeriod": 31.0, "smoothing": 1600.0, "tolerance": 0.0 } };
        const indicator = new CyclePolarity(parameters);
        const expected: Record<string, number[]> = { "POLARITY": data.expectedW100_S17_5_L31_HP1600_T0_POLARITY, "SHORT_VALUE": data.expectedW100_S17_5_L31_HP1600_T0_SHORT_VALUE, "LONG_VALUE": data.expectedW100_S17_5_L31_HP1600_T0_LONG_VALUE, "SHORT_SLOPE": data.expectedW100_S17_5_L31_HP1600_T0_SHORT_SLOPE, "LONG_SLOPE": data.expectedW100_S17_5_L31_HP1600_T0_LONG_SLOPE };
        for (let index = 0; index < data.testInput.length; index++) {
            const result = indicator.update(data.testInput[index]);
            const actual: Record<string, number> = { "POLARITY": result.polarity, "SHORT_VALUE": result.shortValue, "LONG_VALUE": result.longValue, "SHORT_SLOPE": result.shortSlope, "LONG_SLOPE": result.longSlope };
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
        const parameters = { ...defaultParams(), ...{ "window": 100, "shortPeriod": 20.0, "longPeriod": 40.0, "smoothing": 1600.0, "tolerance": 0.1 } };
        const indicator = new CyclePolarity(parameters);
        const expected: Record<string, number[]> = { "POLARITY": data.expectedW100_S20_L40_HP1600_T0_1_POLARITY, "SHORT_VALUE": data.expectedW100_S20_L40_HP1600_T0_1_SHORT_VALUE, "LONG_VALUE": data.expectedW100_S20_L40_HP1600_T0_1_LONG_VALUE, "SHORT_SLOPE": data.expectedW100_S20_L40_HP1600_T0_1_SHORT_SLOPE, "LONG_SLOPE": data.expectedW100_S20_L40_HP1600_T0_1_LONG_SLOPE };
        for (let index = 0; index < data.testInput.length; index++) {
            const result = indicator.update(data.testInput[index]);
            const actual: Record<string, number> = { "POLARITY": result.polarity, "SHORT_VALUE": result.shortValue, "LONG_VALUE": result.longValue, "SHORT_SLOPE": result.shortSlope, "LONG_SLOPE": result.longSlope };
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
        const parameters = { ...defaultParams(), ...{ "window": 100, "shortPeriod": 20.0, "longPeriod": 40.0, "smoothing": 1600.0, "tolerance": 1000.0 } };
        const indicator = new CyclePolarity(parameters);
        const expected: Record<string, number[]> = { "POLARITY": data.expectedW100_S20_L40_HP1600_T1000_POLARITY, "SHORT_VALUE": data.expectedW100_S20_L40_HP1600_T1000_SHORT_VALUE, "LONG_VALUE": data.expectedW100_S20_L40_HP1600_T1000_LONG_VALUE, "SHORT_SLOPE": data.expectedW100_S20_L40_HP1600_T1000_SHORT_SLOPE, "LONG_SLOPE": data.expectedW100_S20_L40_HP1600_T1000_LONG_SLOPE };
        for (let index = 0; index < data.testInput.length; index++) {
            const result = indicator.update(data.testInput[index]);
            const actual: Record<string, number> = { "POLARITY": result.polarity, "SHORT_VALUE": result.shortValue, "LONG_VALUE": result.longValue, "SHORT_SLOPE": result.shortSlope, "LONG_SLOPE": result.longSlope };
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
        const parameters = { ...defaultParams(), ...{ "window": 253, "shortPeriod": 20.0, "longPeriod": 40.0, "smoothing": 1600.0, "tolerance": 0.0 } };
        const indicator = new CyclePolarity(parameters);
        const expected: Record<string, number[]> = { "POLARITY": data.expectedW253_S20_L40_HP1600_T0_POLARITY, "SHORT_VALUE": data.expectedW253_S20_L40_HP1600_T0_SHORT_VALUE, "LONG_VALUE": data.expectedW253_S20_L40_HP1600_T0_LONG_VALUE, "SHORT_SLOPE": data.expectedW253_S20_L40_HP1600_T0_SHORT_SLOPE, "LONG_SLOPE": data.expectedW253_S20_L40_HP1600_T0_LONG_SLOPE };
        for (let index = 0; index < data.testInput.length; index++) {
            const result = indicator.update(data.testInput[index]);
            const actual: Record<string, number> = { "POLARITY": result.polarity, "SHORT_VALUE": result.shortValue, "LONG_VALUE": result.longValue, "SHORT_SLOPE": result.shortSlope, "LONG_SLOPE": result.longSlope };
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
        const indicator = new CyclePolarity();
        indicator.update(NaN);
        expect(indicator.isPrimed()).toBeFalse();
        expect(() => indicator.update(Infinity)).toThrow();
    });
});
