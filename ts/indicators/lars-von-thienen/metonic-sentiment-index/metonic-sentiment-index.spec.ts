import { MetonicSentimentIndex } from "./metonic-sentiment-index.js";
import { defaultParams } from "./params.js";
import * as data from "./testdata.js";
describe("Lars von Thienen MetonicSentimentIndex", () => {
    it("matches prepared case 0", () => {
        const parameters = { ...defaultParams(), ...{ "length": 13, "delay1": 6939, "delay2": 4016, "delay3": 2922, "mode": 0 } };
        const indicator = new MetonicSentimentIndex(parameters);
        const expected: Record<string, number[]> = { "VALUE": data.expectedL13_D16939_D24016_D32922_M0_VALUE, "DAILY_SENTIMENT": data.expectedL13_D16939_D24016_D32922_M0_DAILY_SENTIMENT, "SMOOTHED_CLOSE": data.expectedL13_D16939_D24016_D32922_M0_SMOOTHED_CLOSE, "SMOOTHED_OPEN": data.expectedL13_D16939_D24016_D32922_M0_SMOOTHED_OPEN };
        for (let index = 0; index < data.testInput.length; index++) {
            const result = indicator.update(data.testInput[index], new Date(Date.UTC(2000, 0, index + 1)), data.testOpen[index]);
            const actual: Record<string, number> = { "VALUE": result.value, "DAILY_SENTIMENT": result.dailySentiment, "SMOOTHED_CLOSE": result.smoothedClose, "SMOOTHED_OPEN": result.smoothedOpen };
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
        const parameters = { ...defaultParams(), ...{ "length": 13, "delay1": 6939, "delay2": 4016, "delay3": 2922, "mode": 1 } };
        const indicator = new MetonicSentimentIndex(parameters);
        const expected: Record<string, number[]> = { "VALUE": data.expectedL13_D16939_D24016_D32922_M1_VALUE, "DAILY_SENTIMENT": data.expectedL13_D16939_D24016_D32922_M1_DAILY_SENTIMENT, "SMOOTHED_CLOSE": data.expectedL13_D16939_D24016_D32922_M1_SMOOTHED_CLOSE, "SMOOTHED_OPEN": data.expectedL13_D16939_D24016_D32922_M1_SMOOTHED_OPEN };
        for (let index = 0; index < data.testInput.length; index++) {
            const result = indicator.update(data.testInput[index], new Date(Date.UTC(2000, 0, index + 1)), data.testOpen[index]);
            const actual: Record<string, number> = { "VALUE": result.value, "DAILY_SENTIMENT": result.dailySentiment, "SMOOTHED_CLOSE": result.smoothedClose, "SMOOTHED_OPEN": result.smoothedOpen };
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
        const parameters = { ...defaultParams(), ...{ "length": 13, "delay1": 19, "delay2": 11, "delay3": 8, "mode": 0 } };
        const indicator = new MetonicSentimentIndex(parameters);
        const expected: Record<string, number[]> = { "VALUE": data.expectedL13_D119_D211_D38_M0_VALUE, "DAILY_SENTIMENT": data.expectedL13_D119_D211_D38_M0_DAILY_SENTIMENT, "SMOOTHED_CLOSE": data.expectedL13_D119_D211_D38_M0_SMOOTHED_CLOSE, "SMOOTHED_OPEN": data.expectedL13_D119_D211_D38_M0_SMOOTHED_OPEN };
        for (let index = 0; index < data.testInput.length; index++) {
            const result = indicator.update(data.testInput[index], new Date(Date.UTC(2000, 0, index + 1)), data.testOpen[index]);
            const actual: Record<string, number> = { "VALUE": result.value, "DAILY_SENTIMENT": result.dailySentiment, "SMOOTHED_CLOSE": result.smoothedClose, "SMOOTHED_OPEN": result.smoothedOpen };
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
        const parameters = { ...defaultParams(), ...{ "length": 13, "delay1": 19, "delay2": 11, "delay3": 8, "mode": 1 } };
        const indicator = new MetonicSentimentIndex(parameters);
        const expected: Record<string, number[]> = { "VALUE": data.expectedL13_D119_D211_D38_M1_VALUE, "DAILY_SENTIMENT": data.expectedL13_D119_D211_D38_M1_DAILY_SENTIMENT, "SMOOTHED_CLOSE": data.expectedL13_D119_D211_D38_M1_SMOOTHED_CLOSE, "SMOOTHED_OPEN": data.expectedL13_D119_D211_D38_M1_SMOOTHED_OPEN };
        for (let index = 0; index < data.testInput.length; index++) {
            const result = indicator.update(data.testInput[index], new Date(Date.UTC(2000, 0, index + 1)), data.testOpen[index]);
            const actual: Record<string, number> = { "VALUE": result.value, "DAILY_SENTIMENT": result.dailySentiment, "SMOOTHED_CLOSE": result.smoothedClose, "SMOOTHED_OPEN": result.smoothedOpen };
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
        const parameters = { ...defaultParams(), ...{ "length": 1, "delay1": 19, "delay2": 11, "delay3": 8, "mode": 0 } };
        const indicator = new MetonicSentimentIndex(parameters);
        const expected: Record<string, number[]> = { "VALUE": data.expectedL1_D119_D211_D38_M0_VALUE, "DAILY_SENTIMENT": data.expectedL1_D119_D211_D38_M0_DAILY_SENTIMENT, "SMOOTHED_CLOSE": data.expectedL1_D119_D211_D38_M0_SMOOTHED_CLOSE, "SMOOTHED_OPEN": data.expectedL1_D119_D211_D38_M0_SMOOTHED_OPEN };
        for (let index = 0; index < data.testInput.length; index++) {
            const result = indicator.update(data.testInput[index], new Date(Date.UTC(2000, 0, index + 1)), data.testOpen[index]);
            const actual: Record<string, number> = { "VALUE": result.value, "DAILY_SENTIMENT": result.dailySentiment, "SMOOTHED_CLOSE": result.smoothedClose, "SMOOTHED_OPEN": result.smoothedOpen };
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
        const parameters = { ...defaultParams(), ...{ "length": 10, "delay1": 19, "delay2": 11, "delay3": 8, "mode": 1 } };
        const indicator = new MetonicSentimentIndex(parameters);
        const expected: Record<string, number[]> = { "VALUE": data.expectedL10_D119_D211_D38_M1_VALUE, "DAILY_SENTIMENT": data.expectedL10_D119_D211_D38_M1_DAILY_SENTIMENT, "SMOOTHED_CLOSE": data.expectedL10_D119_D211_D38_M1_SMOOTHED_CLOSE, "SMOOTHED_OPEN": data.expectedL10_D119_D211_D38_M1_SMOOTHED_OPEN };
        for (let index = 0; index < data.testInput.length; index++) {
            const result = indicator.update(data.testInput[index], new Date(Date.UTC(2000, 0, index + 1)), data.testOpen[index]);
            const actual: Record<string, number> = { "VALUE": result.value, "DAILY_SENTIMENT": result.dailySentiment, "SMOOTHED_CLOSE": result.smoothedClose, "SMOOTHED_OPEN": result.smoothedOpen };
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
        const parameters = { ...defaultParams(), ...{ "length": 20, "delay1": 100, "delay2": 60, "delay3": 40, "mode": 0 } };
        const indicator = new MetonicSentimentIndex(parameters);
        const expected: Record<string, number[]> = { "VALUE": data.expectedL20_D1100_D260_D340_M0_VALUE, "DAILY_SENTIMENT": data.expectedL20_D1100_D260_D340_M0_DAILY_SENTIMENT, "SMOOTHED_CLOSE": data.expectedL20_D1100_D260_D340_M0_SMOOTHED_CLOSE, "SMOOTHED_OPEN": data.expectedL20_D1100_D260_D340_M0_SMOOTHED_OPEN };
        for (let index = 0; index < data.testInput.length; index++) {
            const result = indicator.update(data.testInput[index], new Date(Date.UTC(2000, 0, index + 1)), data.testOpen[index]);
            const actual: Record<string, number> = { "VALUE": result.value, "DAILY_SENTIMENT": result.dailySentiment, "SMOOTHED_CLOSE": result.smoothedClose, "SMOOTHED_OPEN": result.smoothedOpen };
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
        const parameters = { ...defaultParams(), ...{ "length": 13, "delay1": 8, "delay2": 19, "delay3": 11, "mode": 1 } };
        const indicator = new MetonicSentimentIndex(parameters);
        const expected: Record<string, number[]> = { "VALUE": data.expectedL13_D18_D219_D311_M1_VALUE, "DAILY_SENTIMENT": data.expectedL13_D18_D219_D311_M1_DAILY_SENTIMENT, "SMOOTHED_CLOSE": data.expectedL13_D18_D219_D311_M1_SMOOTHED_CLOSE, "SMOOTHED_OPEN": data.expectedL13_D18_D219_D311_M1_SMOOTHED_OPEN };
        for (let index = 0; index < data.testInput.length; index++) {
            const result = indicator.update(data.testInput[index], new Date(Date.UTC(2000, 0, index + 1)), data.testOpen[index]);
            const actual: Record<string, number> = { "VALUE": result.value, "DAILY_SENTIMENT": result.dailySentiment, "SMOOTHED_CLOSE": result.smoothedClose, "SMOOTHED_OPEN": result.smoothedOpen };
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
        const parameters = { ...defaultParams(), ...{ "length": 1, "delay1": 251, "delay2": 100, "delay3": 50, "mode": 0 } };
        const indicator = new MetonicSentimentIndex(parameters);
        const expected: Record<string, number[]> = { "VALUE": data.expectedL1_D1251_D2100_D350_M0_VALUE, "DAILY_SENTIMENT": data.expectedL1_D1251_D2100_D350_M0_DAILY_SENTIMENT, "SMOOTHED_CLOSE": data.expectedL1_D1251_D2100_D350_M0_SMOOTHED_CLOSE, "SMOOTHED_OPEN": data.expectedL1_D1251_D2100_D350_M0_SMOOTHED_OPEN };
        for (let index = 0; index < data.testInput.length; index++) {
            const result = indicator.update(data.testInput[index], new Date(Date.UTC(2000, 0, index + 1)), data.testOpen[index]);
            const actual: Record<string, number> = { "VALUE": result.value, "DAILY_SENTIMENT": result.dailySentiment, "SMOOTHED_CLOSE": result.smoothedClose, "SMOOTHED_OPEN": result.smoothedOpen };
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
        const indicator = new MetonicSentimentIndex();
        indicator.update(NaN);
        expect(indicator.isPrimed()).toBeFalse();
        expect(() => indicator.update(Infinity)).toThrow();
    });
});
