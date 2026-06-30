import { } from 'jasmine';

import { FractalBands } from './fractal-bands';
import {
    testInput,
    expectedFrasma2P10Ns20A2, expectedUpperP10Ns20A2, expectedLowerP10Ns20A2,
    expectedFrasma2P20Ns20A2, expectedUpperP20Ns20A2, expectedLowerP20Ns20A2,
    expectedFrasma2P30Ns20A2, expectedUpperP30Ns20A2, expectedLowerP30Ns20A2,
    expectedFrasma2P50Ns20A2, expectedUpperP50Ns20A2, expectedLowerP50Ns20A2,
    expectedFrasma2P30Ns10A2, expectedUpperP30Ns10A2, expectedLowerP30Ns10A2,
    expectedFrasma2P30Ns40A2, expectedUpperP30Ns40A2, expectedLowerP30Ns40A2,
    expectedFrasma2P30Ns20A1, expectedUpperP30Ns20A1, expectedLowerP30Ns20A1,
    expectedFrasma2P30Ns20A3, expectedUpperP30Ns20A3, expectedLowerP30Ns20A3,
} from './testdata';
import { IndicatorIdentifier } from '../../core/indicator-identifier';

describe('FractalBands', () => {

    function runTest(
        period: number, normalSpeed: number, alpha: number,
        expFrasma2: number[], expUpper: number[], expLower: number[],
    ) {
        const ind = new FractalBands({ period, normalSpeed, alpha });

        for (let i = 0; i < testInput.length; i++) {
            const [frasma2, upper, lower] = ind.updateAll(testInput[i]);
            // Use an absolute tolerance of 1e-13 to match the Go and Rust
            // FractalBands tests (epsilon = 1e-13). Jasmine's toBeCloseTo(x, 13)
            // only allows 5e-14, which is too strict for the price-magnitude band
            // outputs (~170): the canonical reference data is generated with
            // compensated summation while the streaming implementation uses a
            // naive running sum, differing by up to ~3e-14 here — well within the
            // project's 1e-13 indicator tolerance.
            if (Number.isNaN(expFrasma2[i])) {
                expect(frasma2).toBeNaN();
            } else {
                expect(Math.abs(frasma2 - expFrasma2[i])).toBeLessThan(1e-13);
            }
            if (Number.isNaN(expUpper[i])) {
                expect(upper).toBeNaN();
            } else {
                expect(Math.abs(upper - expUpper[i])).toBeLessThan(1e-13);
            }
            if (Number.isNaN(expLower[i])) {
                expect(lower).toBeNaN();
            } else {
                expect(Math.abs(lower - expLower[i])).toBeLessThan(1e-13);
            }
        }
    }

    it('should calculate expected output for P10_NS20_A2', () => {
        runTest(10, 20, 2.0, expectedFrasma2P10Ns20A2, expectedUpperP10Ns20A2, expectedLowerP10Ns20A2);
    });

    it('should calculate expected output for P20_NS20_A2', () => {
        runTest(20, 20, 2.0, expectedFrasma2P20Ns20A2, expectedUpperP20Ns20A2, expectedLowerP20Ns20A2);
    });

    it('should calculate expected output for P30_NS20_A2', () => {
        runTest(30, 20, 2.0, expectedFrasma2P30Ns20A2, expectedUpperP30Ns20A2, expectedLowerP30Ns20A2);
    });

    it('should calculate expected output for P50_NS20_A2', () => {
        runTest(50, 20, 2.0, expectedFrasma2P50Ns20A2, expectedUpperP50Ns20A2, expectedLowerP50Ns20A2);
    });

    it('should calculate expected output for P30_NS10_A2', () => {
        runTest(30, 10, 2.0, expectedFrasma2P30Ns10A2, expectedUpperP30Ns10A2, expectedLowerP30Ns10A2);
    });

    it('should calculate expected output for P30_NS40_A2', () => {
        runTest(30, 40, 2.0, expectedFrasma2P30Ns40A2, expectedUpperP30Ns40A2, expectedLowerP30Ns40A2);
    });

    it('should calculate expected output for P30_NS20_A1', () => {
        runTest(30, 20, 1.0, expectedFrasma2P30Ns20A1, expectedUpperP30Ns20A1, expectedLowerP30Ns20A1);
    });

    it('should calculate expected output for P30_NS20_A3', () => {
        runTest(30, 20, 3.0, expectedFrasma2P30Ns20A3, expectedUpperP30Ns20A3, expectedLowerP30Ns20A3);
    });

    it('should track primed state', () => {
        const ind = new FractalBands({ period: 30, normalSpeed: 20, alpha: 2.0 });
        for (let i = 0; i < 29; i++) {
            ind.update(testInput[i]);
            expect(ind.isPrimed()).toBe(false);
        }
        ind.update(testInput[29]);
        expect(ind.isPrimed()).toBe(true);
    });

    it('should pass through NaN', () => {
        const ind = new FractalBands({ period: 5, normalSpeed: 20, alpha: 2.0 });
        const [frasma2, upper, lower] = ind.updateAll(Number.NaN);
        expect(frasma2).toBeNaN();
        expect(upper).toBeNaN();
        expect(lower).toBeNaN();
    });

    it('should reject invalid period', () => {
        expect(() => new FractalBands({ period: 1, normalSpeed: 20, alpha: 2.0 })).toThrowError();
    });

    it('should reject invalid normalSpeed', () => {
        expect(() => new FractalBands({ period: 30, normalSpeed: 0, alpha: 2.0 })).toThrowError();
    });

    it('should reject invalid alpha', () => {
        expect(() => new FractalBands({ period: 30, normalSpeed: 20, alpha: 0.0 })).toThrowError();
    });

    it('should return correct metadata', () => {
        const ind = new FractalBands({ period: 30, normalSpeed: 20, alpha: 2.0 });
        const meta = ind.metadata();
        expect(meta.identifier).toBe(IndicatorIdentifier.FractalBands);
        expect(meta.mnemonic).toContain('fban(30');
    });
});
