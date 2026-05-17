import { } from 'jasmine';

import { FractalBandsHybrideAdaptive } from './fractal-bands-hybride-adaptive';
import {
    testInput,
    expectedFrasmaP10NY05AHP007, expectedUpperP10NY05AHP007, expectedLowerP10NY05AHP007,
    expectedFrasmaP10NY05AHP015, expectedUpperP10NY05AHP015, expectedLowerP10NY05AHP015,
    expectedFrasmaP10NY10AHP007, expectedUpperP10NY10AHP007, expectedLowerP10NY10AHP007,
    expectedFrasmaP10NY10AHP015, expectedUpperP10NY10AHP015, expectedLowerP10NY10AHP015,
    expectedFrasmaP20NY05AHP007, expectedUpperP20NY05AHP007, expectedLowerP20NY05AHP007,
    expectedFrasmaP20NY05AHP015, expectedUpperP20NY05AHP015, expectedLowerP20NY05AHP015,
    expectedFrasmaP20NY10AHP007, expectedUpperP20NY10AHP007, expectedLowerP20NY10AHP007,
    expectedFrasmaP20NY10AHP015, expectedUpperP20NY10AHP015, expectedLowerP20NY10AHP015,
    expectedFrasmaP30NY05AHP007, expectedUpperP30NY05AHP007, expectedLowerP30NY05AHP007,
    expectedFrasmaP30NY05AHP015, expectedUpperP30NY05AHP015, expectedLowerP30NY05AHP015,
    expectedFrasmaP30NY10AHP007, expectedUpperP30NY10AHP007, expectedLowerP30NY10AHP007,
    expectedFrasmaP30NY10AHP015, expectedUpperP30NY10AHP015, expectedLowerP30NY10AHP015,
    expectedFrasmaP50NY05AHP007, expectedUpperP50NY05AHP007, expectedLowerP50NY05AHP007,
    expectedFrasmaP50NY05AHP015, expectedUpperP50NY05AHP015, expectedLowerP50NY05AHP015,
    expectedFrasmaP50NY10AHP007, expectedUpperP50NY10AHP007, expectedLowerP50NY10AHP007,
    expectedFrasmaP50NY10AHP015, expectedUpperP50NY10AHP015, expectedLowerP50NY10AHP015,
} from './testdata';
import { IndicatorIdentifier } from '../../core/indicator-identifier';

describe('FractalBandsHybrideAdaptive', () => {

    function runTest(
        period: number, normalSpeedFallback: number, alpha: number,
        nyquist: number, alphaHP: number,
        expFrasma: number[], expUpper: number[], expLower: number[],
    ) {
        const ind = new FractalBandsHybrideAdaptive({ period, normalSpeedFallback, alpha, nyquist, alphaHP });

        for (let i = 0; i < testInput.length; i++) {
            const [frasma2, upper, lower] = ind.updateAll(testInput[i]);
            if (Number.isNaN(expFrasma[i])) {
                expect(frasma2).toBeNaN();
            } else {
                expect(frasma2).toBeCloseTo(expFrasma[i], 12);
            }
            if (Number.isNaN(expUpper[i])) {
                expect(upper).toBeNaN();
            } else {
                expect(upper).toBeCloseTo(expUpper[i], 12);
            }
            if (Number.isNaN(expLower[i])) {
                expect(lower).toBeNaN();
            } else {
                expect(lower).toBeCloseTo(expLower[i], 12);
            }
        }
    }

    it('should calculate expected output for P10_NY05_AHP007', () => {
        runTest(10, 30, 2.0, 0.5, 0.07, expectedFrasmaP10NY05AHP007, expectedUpperP10NY05AHP007, expectedLowerP10NY05AHP007);
    });

    it('should calculate expected output for P10_NY05_AHP015', () => {
        runTest(10, 30, 2.0, 0.5, 0.15, expectedFrasmaP10NY05AHP015, expectedUpperP10NY05AHP015, expectedLowerP10NY05AHP015);
    });

    it('should calculate expected output for P10_NY10_AHP007', () => {
        runTest(10, 30, 2.0, 1.0, 0.07, expectedFrasmaP10NY10AHP007, expectedUpperP10NY10AHP007, expectedLowerP10NY10AHP007);
    });

    it('should calculate expected output for P10_NY10_AHP015', () => {
        runTest(10, 30, 2.0, 1.0, 0.15, expectedFrasmaP10NY10AHP015, expectedUpperP10NY10AHP015, expectedLowerP10NY10AHP015);
    });

    it('should calculate expected output for P20_NY05_AHP007', () => {
        runTest(20, 30, 2.0, 0.5, 0.07, expectedFrasmaP20NY05AHP007, expectedUpperP20NY05AHP007, expectedLowerP20NY05AHP007);
    });

    it('should calculate expected output for P20_NY05_AHP015', () => {
        runTest(20, 30, 2.0, 0.5, 0.15, expectedFrasmaP20NY05AHP015, expectedUpperP20NY05AHP015, expectedLowerP20NY05AHP015);
    });

    it('should calculate expected output for P20_NY10_AHP007', () => {
        runTest(20, 30, 2.0, 1.0, 0.07, expectedFrasmaP20NY10AHP007, expectedUpperP20NY10AHP007, expectedLowerP20NY10AHP007);
    });

    it('should calculate expected output for P20_NY10_AHP015', () => {
        runTest(20, 30, 2.0, 1.0, 0.15, expectedFrasmaP20NY10AHP015, expectedUpperP20NY10AHP015, expectedLowerP20NY10AHP015);
    });

    it('should calculate expected output for P30_NY05_AHP007', () => {
        runTest(30, 30, 2.0, 0.5, 0.07, expectedFrasmaP30NY05AHP007, expectedUpperP30NY05AHP007, expectedLowerP30NY05AHP007);
    });

    it('should calculate expected output for P30_NY05_AHP015', () => {
        runTest(30, 30, 2.0, 0.5, 0.15, expectedFrasmaP30NY05AHP015, expectedUpperP30NY05AHP015, expectedLowerP30NY05AHP015);
    });

    it('should calculate expected output for P30_NY10_AHP007', () => {
        runTest(30, 30, 2.0, 1.0, 0.07, expectedFrasmaP30NY10AHP007, expectedUpperP30NY10AHP007, expectedLowerP30NY10AHP007);
    });

    it('should calculate expected output for P30_NY10_AHP015', () => {
        runTest(30, 30, 2.0, 1.0, 0.15, expectedFrasmaP30NY10AHP015, expectedUpperP30NY10AHP015, expectedLowerP30NY10AHP015);
    });

    it('should calculate expected output for P50_NY05_AHP007', () => {
        runTest(50, 30, 2.0, 0.5, 0.07, expectedFrasmaP50NY05AHP007, expectedUpperP50NY05AHP007, expectedLowerP50NY05AHP007);
    });

    it('should calculate expected output for P50_NY05_AHP015', () => {
        runTest(50, 30, 2.0, 0.5, 0.15, expectedFrasmaP50NY05AHP015, expectedUpperP50NY05AHP015, expectedLowerP50NY05AHP015);
    });

    it('should calculate expected output for P50_NY10_AHP007', () => {
        runTest(50, 30, 2.0, 1.0, 0.07, expectedFrasmaP50NY10AHP007, expectedUpperP50NY10AHP007, expectedLowerP50NY10AHP007);
    });

    it('should calculate expected output for P50_NY10_AHP015', () => {
        runTest(50, 30, 2.0, 1.0, 0.15, expectedFrasmaP50NY10AHP015, expectedUpperP50NY10AHP015, expectedLowerP50NY10AHP015);
    });

    it('should track primed state', () => {
        const ind = new FractalBandsHybrideAdaptive({ period: 30, normalSpeedFallback: 30, alpha: 2.0, nyquist: 0.5, alphaHP: 0.07 });
        for (let i = 0; i < 30; i++) {
            ind.update(testInput[i]);
            expect(ind.isPrimed()).toBeFalse();
        }
        ind.update(testInput[30]);
        expect(ind.isPrimed()).toBeTrue();
    });

    it('should pass NaN through', () => {
        const ind = new FractalBandsHybrideAdaptive({ period: 5, normalSpeedFallback: 30, alpha: 2.0, nyquist: 0.5, alphaHP: 0.07 });
        const [f, u, l] = ind.updateAll(Number.NaN);
        expect(f).toBeNaN();
        expect(u).toBeNaN();
        expect(l).toBeNaN();
    });

    it('should throw on invalid period', () => {
        expect(() => new FractalBandsHybrideAdaptive({ period: 1, normalSpeedFallback: 30, alpha: 2.0, nyquist: 0.5, alphaHP: 0.07 })).toThrowError();
    });

    it('should throw on invalid normalSpeedFallback', () => {
        expect(() => new FractalBandsHybrideAdaptive({ period: 30, normalSpeedFallback: 0, alpha: 2.0, nyquist: 0.5, alphaHP: 0.07 })).toThrowError();
    });

    it('should throw on invalid alpha', () => {
        expect(() => new FractalBandsHybrideAdaptive({ period: 30, normalSpeedFallback: 30, alpha: 0.0, nyquist: 0.5, alphaHP: 0.07 })).toThrowError();
    });

    it('should throw on invalid nyquist', () => {
        expect(() => new FractalBandsHybrideAdaptive({ period: 30, normalSpeedFallback: 30, alpha: 2.0, nyquist: 0.0, alphaHP: 0.07 })).toThrowError();
    });

    it('should throw on invalid alphaHP', () => {
        expect(() => new FractalBandsHybrideAdaptive({ period: 30, normalSpeedFallback: 30, alpha: 2.0, nyquist: 0.5, alphaHP: 0.0 })).toThrowError();
        expect(() => new FractalBandsHybrideAdaptive({ period: 30, normalSpeedFallback: 30, alpha: 2.0, nyquist: 0.5, alphaHP: 1.0 })).toThrowError();
    });

    it('should have correct metadata', () => {
        const ind = new FractalBandsHybrideAdaptive({ period: 30, normalSpeedFallback: 30, alpha: 2.0, nyquist: 0.5, alphaHP: 0.07 });
        const meta = ind.metadata();
        expect(meta.identifier).toBe(IndicatorIdentifier.FractalBandsHybrideAdaptive);
    });
});
