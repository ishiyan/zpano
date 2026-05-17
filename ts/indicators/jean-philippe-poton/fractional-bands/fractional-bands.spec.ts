import { } from 'jasmine';

import { FractionalBands } from './fractional-bands';
import {
    testInput,
    expectedFrasma2P5S1, expectedUpperP5S1, expectedLowerP5S1,
    expectedFrasma2P10S1, expectedUpperP10S1, expectedLowerP10S1,
    expectedFrasma2P20S1, expectedUpperP20S1, expectedLowerP20S1,
    expectedFrasma2P30S1, expectedUpperP30S1, expectedLowerP30S1,
    expectedFrasma2P50S1, expectedUpperP50S1, expectedLowerP50S1,
    expectedFrasma2P80S1, expectedUpperP80S1, expectedLowerP80S1,
    expectedFrasma2P30S100, expectedUpperP30S100, expectedLowerP30S100,
    expectedFrasma2P30S10000, expectedUpperP30S10000, expectedLowerP30S10000,
} from './testdata';
import { IndicatorIdentifier } from '../../core/indicator-identifier';

describe('FractionalBands', () => {

    function runTest(
        period: number, priceScale: number,
        expFrasma2: number[], expUpper: number[], expLower: number[],
        precision: number = 13,
    ) {
        const ind = new FractionalBands({ period, priceScale });

        for (let i = 0; i < testInput.length; i++) {
            const [frasma2, upper, lower] = ind.updateAll(testInput[i]);
            if (Number.isNaN(expFrasma2[i])) {
                expect(frasma2).toBeNaN();
            } else {
                expect(frasma2).toBeCloseTo(expFrasma2[i], precision);
            }
            if (Number.isNaN(expUpper[i])) {
                expect(upper).toBeNaN();
            } else {
                expect(upper).toBeCloseTo(expUpper[i], precision);
            }
            if (Number.isNaN(expLower[i])) {
                expect(lower).toBeNaN();
            } else {
                expect(lower).toBeCloseTo(expLower[i], precision);
            }
        }
    }

    it('should calculate expected output for P5_S1', () => {
        runTest(5, 1.0, expectedFrasma2P5S1, expectedUpperP5S1, expectedLowerP5S1);
    });

    it('should calculate expected output for P10_S1', () => {
        runTest(10, 1.0, expectedFrasma2P10S1, expectedUpperP10S1, expectedLowerP10S1);
    });

    it('should calculate expected output for P20_S1', () => {
        runTest(20, 1.0, expectedFrasma2P20S1, expectedUpperP20S1, expectedLowerP20S1);
    });

    it('should calculate expected output for P30_S1', () => {
        runTest(30, 1.0, expectedFrasma2P30S1, expectedUpperP30S1, expectedLowerP30S1);
    });

    it('should calculate expected output for P50_S1', () => {
        runTest(50, 1.0, expectedFrasma2P50S1, expectedUpperP50S1, expectedLowerP50S1);
    });

    it('should calculate expected output for P80_S1', () => {
        runTest(80, 1.0, expectedFrasma2P80S1, expectedUpperP80S1, expectedLowerP80S1, 12);
    });

    it('should calculate expected output for P30_S100', () => {
        runTest(30, 100.0, expectedFrasma2P30S100, expectedUpperP30S100, expectedLowerP30S100, 11);
    });

    it('should calculate expected output for P30_S10000', () => {
        runTest(30, 10000.0, expectedFrasma2P30S10000, expectedUpperP30S10000, expectedLowerP30S10000, 11);
    });

    it('should track primed state', () => {
        const ind = new FractionalBands({ period: 30, priceScale: 1.0 });
        for (let i = 0; i < 30; i++) {
            ind.update(testInput[i]);
            expect(ind.isPrimed()).toBe(false);
        }
        ind.update(testInput[30]);
        expect(ind.isPrimed()).toBe(true);
    });

    it('should pass through NaN', () => {
        const ind = new FractionalBands({ period: 5, priceScale: 1.0 });
        const [frasma2, upper, lower] = ind.updateAll(Number.NaN);
        expect(frasma2).toBeNaN();
        expect(upper).toBeNaN();
        expect(lower).toBeNaN();
    });

    it('should reject invalid period', () => {
        expect(() => new FractionalBands({ period: 1, priceScale: 1.0 })).toThrowError();
    });

    it('should reject invalid priceScale', () => {
        expect(() => new FractionalBands({ period: 30, priceScale: 0.0 })).toThrowError();
        expect(() => new FractionalBands({ period: 30, priceScale: -1.0 })).toThrowError();
    });

    it('should return correct metadata', () => {
        const ind = new FractionalBands({ period: 30, priceScale: 1.0 });
        const meta = ind.metadata();
        expect(meta.identifier).toBe(IndicatorIdentifier.FractionalBands);
        expect(meta.mnemonic).toContain('fctban(30');
    });
});
