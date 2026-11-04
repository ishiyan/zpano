import { MetonicRelativeStrengthIndex } from "./metonic-relative-strength-index.js";
import { defaultParams } from "./params.js";
import * as data from "./testdata.js";
describe("Lars von Thienen MetonicRelativeStrengthIndex", () => {
    it("matches prepared case 0", () => {
        const parameters = { ...defaultParams(), ...{ "length": 13, "delay1": 6939, "delay2": 4016, "delay3": 2922 } };
        const indicator = new MetonicRelativeStrengthIndex(parameters);
        const expected: Record<string, number[]> = { "VALUE": data.expectedL13_D16939_D24016_D32922_VALUE, "LINEAR_RSI": data.expectedL13_D16939_D24016_D32922_LINEAR_RSI };
        for (let index = 0; index < data.testInput.length; index++) {
            const result = indicator.update(data.testInput[index], new Date(Date.UTC(2000, 0, index + 1)));
            const actual: Record<string, number> = { "VALUE": result.value, "LINEAR_RSI": result.linearRsi };
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
        const parameters = { ...defaultParams(), ...{ "length": 13, "delay1": 19, "delay2": 11, "delay3": 8 } };
        const indicator = new MetonicRelativeStrengthIndex(parameters);
        const expected: Record<string, number[]> = { "VALUE": data.expectedL13_D119_D211_D38_VALUE, "LINEAR_RSI": data.expectedL13_D119_D211_D38_LINEAR_RSI };
        for (let index = 0; index < data.testInput.length; index++) {
            const result = indicator.update(data.testInput[index], new Date(Date.UTC(2000, 0, index + 1)));
            const actual: Record<string, number> = { "VALUE": result.value, "LINEAR_RSI": result.linearRsi };
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
        const parameters = { ...defaultParams(), ...{ "length": 1, "delay1": 19, "delay2": 11, "delay3": 8 } };
        const indicator = new MetonicRelativeStrengthIndex(parameters);
        const expected: Record<string, number[]> = { "VALUE": data.expectedL1_D119_D211_D38_VALUE, "LINEAR_RSI": data.expectedL1_D119_D211_D38_LINEAR_RSI };
        for (let index = 0; index < data.testInput.length; index++) {
            const result = indicator.update(data.testInput[index], new Date(Date.UTC(2000, 0, index + 1)));
            const actual: Record<string, number> = { "VALUE": result.value, "LINEAR_RSI": result.linearRsi };
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
        const parameters = { ...defaultParams(), ...{ "length": 14, "delay1": 19, "delay2": 11, "delay3": 8 } };
        const indicator = new MetonicRelativeStrengthIndex(parameters);
        const expected: Record<string, number[]> = { "VALUE": data.expectedL14_D119_D211_D38_VALUE, "LINEAR_RSI": data.expectedL14_D119_D211_D38_LINEAR_RSI };
        for (let index = 0; index < data.testInput.length; index++) {
            const result = indicator.update(data.testInput[index], new Date(Date.UTC(2000, 0, index + 1)));
            const actual: Record<string, number> = { "VALUE": result.value, "LINEAR_RSI": result.linearRsi };
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
        const parameters = { ...defaultParams(), ...{ "length": 50, "delay1": 100, "delay2": 60, "delay3": 40 } };
        const indicator = new MetonicRelativeStrengthIndex(parameters);
        const expected: Record<string, number[]> = { "VALUE": data.expectedL50_D1100_D260_D340_VALUE, "LINEAR_RSI": data.expectedL50_D1100_D260_D340_LINEAR_RSI };
        for (let index = 0; index < data.testInput.length; index++) {
            const result = indicator.update(data.testInput[index], new Date(Date.UTC(2000, 0, index + 1)));
            const actual: Record<string, number> = { "VALUE": result.value, "LINEAR_RSI": result.linearRsi };
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
        const parameters = { ...defaultParams(), ...{ "length": 13, "delay1": 8, "delay2": 19, "delay3": 11 } };
        const indicator = new MetonicRelativeStrengthIndex(parameters);
        const expected: Record<string, number[]> = { "VALUE": data.expectedL13_D18_D219_D311_VALUE, "LINEAR_RSI": data.expectedL13_D18_D219_D311_LINEAR_RSI };
        for (let index = 0; index < data.testInput.length; index++) {
            const result = indicator.update(data.testInput[index], new Date(Date.UTC(2000, 0, index + 1)));
            const actual: Record<string, number> = { "VALUE": result.value, "LINEAR_RSI": result.linearRsi };
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
        const parameters = { ...defaultParams(), ...{ "length": 13, "delay1": 238, "delay2": 100, "delay3": 50 } };
        const indicator = new MetonicRelativeStrengthIndex(parameters);
        const expected: Record<string, number[]> = { "VALUE": data.expectedL13_D1238_D2100_D350_VALUE, "LINEAR_RSI": data.expectedL13_D1238_D2100_D350_LINEAR_RSI };
        for (let index = 0; index < data.testInput.length; index++) {
            const result = indicator.update(data.testInput[index], new Date(Date.UTC(2000, 0, index + 1)));
            const actual: Record<string, number> = { "VALUE": result.value, "LINEAR_RSI": result.linearRsi };
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
        const parameters = { ...defaultParams(), ...{ "length": 13, "delay1": 239, "delay2": 100, "delay3": 50 } };
        const indicator = new MetonicRelativeStrengthIndex(parameters);
        const expected: Record<string, number[]> = { "VALUE": data.expectedL13_D1239_D2100_D350_VALUE, "LINEAR_RSI": data.expectedL13_D1239_D2100_D350_LINEAR_RSI };
        for (let index = 0; index < data.testInput.length; index++) {
            const result = indicator.update(data.testInput[index], new Date(Date.UTC(2000, 0, index + 1)));
            const actual: Record<string, number> = { "VALUE": result.value, "LINEAR_RSI": result.linearRsi };
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
        const indicator = new MetonicRelativeStrengthIndex();
        indicator.update(NaN);
        expect(indicator.isPrimed()).toBeFalse();
        expect(() => indicator.update(Infinity)).toThrow();
    });
});
