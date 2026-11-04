// SPDX-License-Identifier: MPL-2.0
// Source: Detrended Rhythm Oscillator, © WhenToTrade / Lars von Thienen.
import { DetrendedRhythmOscillators } from "./detrended-rhythm-oscillators.js";
import { defaultParams } from "./params.js";
import * as data from "./testdata.js";
describe("Lars von Thienen DetrendedRhythmOscillators", () => {
    it("matches prepared case 0", () => {
        const parameters = { ...defaultParams(), ...{ "fastLength": 3, "slowLength": 50, "zigPeriod": 30, "cycleMemory": 15, "detrend": true } };
        const indicator = new DetrendedRhythmOscillators(parameters);
        const expected: Record<string, number[]> = { "VALUE": data.expectedF3_S50_Z30_M15_D1_VALUE, "DIRECTION": data.expectedF3_S50_Z30_M15_D1_DIRECTION, "CYCLE_LENGTH": data.expectedF3_S50_Z30_M15_D1_CYCLE_LENGTH, "PIVOT_VALUE": data.expectedF3_S50_Z30_M15_D1_PIVOT_VALUE, "PIVOT_OFFSET": data.expectedF3_S50_Z30_M15_D1_PIVOT_OFFSET };
        for (let index = 0; index < data.testInput.length; index++) {
            const result = indicator.update(data.testInput[index]);
            const actual: Record<string, number> = { "VALUE": result.value, "DIRECTION": result.direction, "CYCLE_LENGTH": result.cycleLength, "PIVOT_VALUE": result.pivotValue, "PIVOT_OFFSET": result.pivotOffset };
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
        const parameters = { ...defaultParams(), ...{ "fastLength": 3, "slowLength": 50, "zigPeriod": 5, "cycleMemory": 15, "detrend": true } };
        const indicator = new DetrendedRhythmOscillators(parameters);
        const expected: Record<string, number[]> = { "VALUE": data.expectedF3_S50_Z5_M15_D1_VALUE, "DIRECTION": data.expectedF3_S50_Z5_M15_D1_DIRECTION, "CYCLE_LENGTH": data.expectedF3_S50_Z5_M15_D1_CYCLE_LENGTH, "PIVOT_VALUE": data.expectedF3_S50_Z5_M15_D1_PIVOT_VALUE, "PIVOT_OFFSET": data.expectedF3_S50_Z5_M15_D1_PIVOT_OFFSET };
        for (let index = 0; index < data.testInput.length; index++) {
            const result = indicator.update(data.testInput[index]);
            const actual: Record<string, number> = { "VALUE": result.value, "DIRECTION": result.direction, "CYCLE_LENGTH": result.cycleLength, "PIVOT_VALUE": result.pivotValue, "PIVOT_OFFSET": result.pivotOffset };
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
        const parameters = { ...defaultParams(), ...{ "fastLength": 3, "slowLength": 50, "zigPeriod": 10, "cycleMemory": 15, "detrend": true } };
        const indicator = new DetrendedRhythmOscillators(parameters);
        const expected: Record<string, number[]> = { "VALUE": data.expectedF3_S50_Z10_M15_D1_VALUE, "DIRECTION": data.expectedF3_S50_Z10_M15_D1_DIRECTION, "CYCLE_LENGTH": data.expectedF3_S50_Z10_M15_D1_CYCLE_LENGTH, "PIVOT_VALUE": data.expectedF3_S50_Z10_M15_D1_PIVOT_VALUE, "PIVOT_OFFSET": data.expectedF3_S50_Z10_M15_D1_PIVOT_OFFSET };
        for (let index = 0; index < data.testInput.length; index++) {
            const result = indicator.update(data.testInput[index]);
            const actual: Record<string, number> = { "VALUE": result.value, "DIRECTION": result.direction, "CYCLE_LENGTH": result.cycleLength, "PIVOT_VALUE": result.pivotValue, "PIVOT_OFFSET": result.pivotOffset };
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
        const parameters = { ...defaultParams(), ...{ "fastLength": 2, "slowLength": 20, "zigPeriod": 10, "cycleMemory": 15, "detrend": true } };
        const indicator = new DetrendedRhythmOscillators(parameters);
        const expected: Record<string, number[]> = { "VALUE": data.expectedF2_S20_Z10_M15_D1_VALUE, "DIRECTION": data.expectedF2_S20_Z10_M15_D1_DIRECTION, "CYCLE_LENGTH": data.expectedF2_S20_Z10_M15_D1_CYCLE_LENGTH, "PIVOT_VALUE": data.expectedF2_S20_Z10_M15_D1_PIVOT_VALUE, "PIVOT_OFFSET": data.expectedF2_S20_Z10_M15_D1_PIVOT_OFFSET };
        for (let index = 0; index < data.testInput.length; index++) {
            const result = indicator.update(data.testInput[index]);
            const actual: Record<string, number> = { "VALUE": result.value, "DIRECTION": result.direction, "CYCLE_LENGTH": result.cycleLength, "PIVOT_VALUE": result.pivotValue, "PIVOT_OFFSET": result.pivotOffset };
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
        const parameters = { ...defaultParams(), ...{ "fastLength": 5, "slowLength": 100, "zigPeriod": 30, "cycleMemory": 15, "detrend": true } };
        const indicator = new DetrendedRhythmOscillators(parameters);
        const expected: Record<string, number[]> = { "VALUE": data.expectedF5_S100_Z30_M15_D1_VALUE, "DIRECTION": data.expectedF5_S100_Z30_M15_D1_DIRECTION, "CYCLE_LENGTH": data.expectedF5_S100_Z30_M15_D1_CYCLE_LENGTH, "PIVOT_VALUE": data.expectedF5_S100_Z30_M15_D1_PIVOT_VALUE, "PIVOT_OFFSET": data.expectedF5_S100_Z30_M15_D1_PIVOT_OFFSET };
        for (let index = 0; index < data.testInput.length; index++) {
            const result = indicator.update(data.testInput[index]);
            const actual: Record<string, number> = { "VALUE": result.value, "DIRECTION": result.direction, "CYCLE_LENGTH": result.cycleLength, "PIVOT_VALUE": result.pivotValue, "PIVOT_OFFSET": result.pivotOffset };
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
        const parameters = { ...defaultParams(), ...{ "fastLength": 3, "slowLength": 50, "zigPeriod": 5, "cycleMemory": 3, "detrend": true } };
        const indicator = new DetrendedRhythmOscillators(parameters);
        const expected: Record<string, number[]> = { "VALUE": data.expectedF3_S50_Z5_M3_D1_VALUE, "DIRECTION": data.expectedF3_S50_Z5_M3_D1_DIRECTION, "CYCLE_LENGTH": data.expectedF3_S50_Z5_M3_D1_CYCLE_LENGTH, "PIVOT_VALUE": data.expectedF3_S50_Z5_M3_D1_PIVOT_VALUE, "PIVOT_OFFSET": data.expectedF3_S50_Z5_M3_D1_PIVOT_OFFSET };
        for (let index = 0; index < data.testInput.length; index++) {
            const result = indicator.update(data.testInput[index]);
            const actual: Record<string, number> = { "VALUE": result.value, "DIRECTION": result.direction, "CYCLE_LENGTH": result.cycleLength, "PIVOT_VALUE": result.pivotValue, "PIVOT_OFFSET": result.pivotOffset };
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
        const parameters = { ...defaultParams(), ...{ "fastLength": 3, "slowLength": 50, "zigPeriod": 10, "cycleMemory": 15, "detrend": false } };
        const indicator = new DetrendedRhythmOscillators(parameters);
        const expected: Record<string, number[]> = { "VALUE": data.expectedF3_S50_Z10_M15_D0_VALUE, "DIRECTION": data.expectedF3_S50_Z10_M15_D0_DIRECTION, "CYCLE_LENGTH": data.expectedF3_S50_Z10_M15_D0_CYCLE_LENGTH, "PIVOT_VALUE": data.expectedF3_S50_Z10_M15_D0_PIVOT_VALUE, "PIVOT_OFFSET": data.expectedF3_S50_Z10_M15_D0_PIVOT_OFFSET };
        for (let index = 0; index < data.testInput.length; index++) {
            const result = indicator.update(data.testInput[index]);
            const actual: Record<string, number> = { "VALUE": result.value, "DIRECTION": result.direction, "CYCLE_LENGTH": result.cycleLength, "PIVOT_VALUE": result.pivotValue, "PIVOT_OFFSET": result.pivotOffset };
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
        const parameters = { ...defaultParams(), ...{ "fastLength": 3, "slowLength": 300, "zigPeriod": 30, "cycleMemory": 15, "detrend": true } };
        const indicator = new DetrendedRhythmOscillators(parameters);
        const expected: Record<string, number[]> = { "VALUE": data.expectedF3_S300_Z30_M15_D1_VALUE, "DIRECTION": data.expectedF3_S300_Z30_M15_D1_DIRECTION, "CYCLE_LENGTH": data.expectedF3_S300_Z30_M15_D1_CYCLE_LENGTH, "PIVOT_VALUE": data.expectedF3_S300_Z30_M15_D1_PIVOT_VALUE, "PIVOT_OFFSET": data.expectedF3_S300_Z30_M15_D1_PIVOT_OFFSET };
        for (let index = 0; index < data.testInput.length; index++) {
            const result = indicator.update(data.testInput[index]);
            const actual: Record<string, number> = { "VALUE": result.value, "DIRECTION": result.direction, "CYCLE_LENGTH": result.cycleLength, "PIVOT_VALUE": result.pivotValue, "PIVOT_OFFSET": result.pivotOffset };
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
        const indicator = new DetrendedRhythmOscillators();
        indicator.update(NaN);
        expect(indicator.isPrimed()).toBeFalse();
        expect(() => indicator.update(Infinity)).toThrow();
    });
});
