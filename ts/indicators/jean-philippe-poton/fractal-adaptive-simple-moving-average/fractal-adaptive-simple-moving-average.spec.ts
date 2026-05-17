import { } from 'jasmine';

import { FractalAdaptiveSimpleMovingAverage } from './fractal-adaptive-simple-moving-average';
import { testInput, expectedP5, expectedP10, expectedP15, expectedP20, expectedP30, expectedP50, expectedP80, expectedP120 } from './testdata';
import { IndicatorIdentifier } from '../../core/indicator-identifier';

describe('FractalAdaptiveSimpleMovingAverage', () => {

    it('should return expected mnemonic', () => {
        const f = new FractalAdaptiveSimpleMovingAverage({ period: 30, normalSpeed: 20 });
        expect(f.metadata().mnemonic).toBe('frasma(30,20)');
    });

    it('should throw if period is less than 2', () => {
        expect(() => { new FractalAdaptiveSimpleMovingAverage({ period: 1, normalSpeed: 20 }); }).toThrow();
    });

    it('should throw if normalSpeed is less than 1', () => {
        expect(() => { new FractalAdaptiveSimpleMovingAverage({ period: 5, normalSpeed: 0 }); }).toThrow();
    });

    it('should have correct identifier in metadata', () => {
        const f = new FractalAdaptiveSimpleMovingAverage({ period: 30, normalSpeed: 20 });
        expect(f.metadata().identifier).toBe(IndicatorIdentifier.FractalAdaptiveSimpleMovingAverage);
    });

    it('should calculate expected output for period 5', () => {
        const f = new FractalAdaptiveSimpleMovingAverage({ period: 5, normalSpeed: 20 });

        for (let i = 0; i < testInput.length; i++) {
            const result = f.update(testInput[i]);
            if (Number.isNaN(expectedP5[i])) {
                expect(result).toBeNaN();
            } else {
                expect(result).toBeCloseTo(expectedP5[i], 13);
            }
        }
    });

    it('should calculate expected output for period 10', () => {
        const f = new FractalAdaptiveSimpleMovingAverage({ period: 10, normalSpeed: 20 });

        for (let i = 0; i < testInput.length; i++) {
            const result = f.update(testInput[i]);
            if (Number.isNaN(expectedP10[i])) {
                expect(result).toBeNaN();
            } else {
                expect(result).toBeCloseTo(expectedP10[i], 13);
            }
        }
    });

    it('should calculate expected output for period 15', () => {
        const f = new FractalAdaptiveSimpleMovingAverage({ period: 15, normalSpeed: 20 });

        for (let i = 0; i < testInput.length; i++) {
            const result = f.update(testInput[i]);
            if (Number.isNaN(expectedP15[i])) {
                expect(result).toBeNaN();
            } else {
                expect(result).toBeCloseTo(expectedP15[i], 13);
            }
        }
    });

    it('should calculate expected output for period 20', () => {
        const f = new FractalAdaptiveSimpleMovingAverage({ period: 20, normalSpeed: 20 });

        for (let i = 0; i < testInput.length; i++) {
            const result = f.update(testInput[i]);
            if (Number.isNaN(expectedP20[i])) {
                expect(result).toBeNaN();
            } else {
                expect(result).toBeCloseTo(expectedP20[i], 13);
            }
        }
    });

    it('should calculate expected output for period 30', () => {
        const f = new FractalAdaptiveSimpleMovingAverage({ period: 30, normalSpeed: 20 });

        for (let i = 0; i < testInput.length; i++) {
            const result = f.update(testInput[i]);
            if (Number.isNaN(expectedP30[i])) {
                expect(result).toBeNaN();
            } else {
                expect(result).toBeCloseTo(expectedP30[i], 13);
            }
        }
    });

    it('should calculate expected output for period 50', () => {
        const f = new FractalAdaptiveSimpleMovingAverage({ period: 50, normalSpeed: 20 });

        for (let i = 0; i < testInput.length; i++) {
            const result = f.update(testInput[i]);
            if (Number.isNaN(expectedP50[i])) {
                expect(result).toBeNaN();
            } else {
                expect(result).toBeCloseTo(expectedP50[i], 13);
            }
        }
    });

    it('should calculate expected output for period 80', () => {
        const f = new FractalAdaptiveSimpleMovingAverage({ period: 80, normalSpeed: 20 });

        for (let i = 0; i < testInput.length; i++) {
            const result = f.update(testInput[i]);
            if (Number.isNaN(expectedP80[i])) {
                expect(result).toBeNaN();
            } else {
                expect(result).toBeCloseTo(expectedP80[i], 13);
            }
        }
    });

    it('should calculate expected output for period 120', () => {
        const f = new FractalAdaptiveSimpleMovingAverage({ period: 120, normalSpeed: 20 });

        for (let i = 0; i < testInput.length; i++) {
            const result = f.update(testInput[i]);
            if (Number.isNaN(expectedP120[i])) {
                expect(result).toBeNaN();
            } else {
                expect(result).toBeCloseTo(expectedP120[i], 13);
            }
        }
    });

    it('should report primed state correctly', () => {
        const f = new FractalAdaptiveSimpleMovingAverage({ period: 30, normalSpeed: 20 });

        for (let i = 0; i < 29; i++) {
            f.update(testInput[i]);
            expect(f.isPrimed()).toBe(false);
        }

        f.update(testInput[29]);
        expect(f.isPrimed()).toBe(true);
    });

    it('should pass NaN through', () => {
        const f = new FractalAdaptiveSimpleMovingAverage({ period: 5, normalSpeed: 20 });
        expect(f.update(Number.NaN)).toBeNaN();
    });
});
