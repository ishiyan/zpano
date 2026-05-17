import { IndicatorIdentifier } from '../../core/indicator-identifier';
import { RescaledFractalAdaptiveSimpleMovingAverage } from './rescaled-fractal-adaptive-simple-moving-average';
import {
    testInput, expectedP4_S1, expectedP8_S1, expectedP16_S1, expectedP32_S1,
    expectedP64_S1, expectedP128_S1, expectedP32_S100, expectedP32_S10000,
} from './testdata';

describe('RescaledFractalAdaptiveSimpleMovingAverage', () => {
    const check = (index: number, expected: number, actual: number): void => {
        if (Number.isNaN(expected)) {
            expect(Number.isNaN(actual)).withContext(`[${index}] expected NaN, got ${actual}`).toBeTrue();
        } else {
            expect(actual).withContext(`[${index}]`).toBeCloseTo(expected, 13);
        }
    };

    const runTest = (period: number, normalSpeed: number, priceScale: number, expected: number[]): void => {
        const f = new RescaledFractalAdaptiveSimpleMovingAverage({ period, normalSpeed, priceScale });

        for (let i = 0; i < testInput.length; i++) {
            const actual = f.update(testInput[i]);
            check(i, expected[i], actual);
        }
    };

    it('should compute period=4, normal_speed=30, price_scale=1.0', () => { runTest(4, 30, 1.0, expectedP4_S1); });
    it('should compute period=8, normal_speed=30, price_scale=1.0', () => { runTest(8, 30, 1.0, expectedP8_S1); });
    it('should compute period=16, normal_speed=30, price_scale=1.0', () => { runTest(16, 30, 1.0, expectedP16_S1); });
    it('should compute period=32, normal_speed=30, price_scale=1.0', () => { runTest(32, 30, 1.0, expectedP32_S1); });
    it('should compute period=64, normal_speed=30, price_scale=1.0', () => { runTest(64, 30, 1.0, expectedP64_S1); });
    it('should compute period=128, normal_speed=30, price_scale=1.0', () => { runTest(128, 30, 1.0, expectedP128_S1); });
    it('should compute period=32, normal_speed=30, price_scale=100.0', () => { runTest(32, 30, 100.0, expectedP32_S100); });
    it('should compute period=32, normal_speed=30, price_scale=10000.0', () => { runTest(32, 30, 10000.0, expectedP32_S10000); });

    it('should track primed state', () => {
        const f = new RescaledFractalAdaptiveSimpleMovingAverage({ period: 64, normalSpeed: 30, priceScale: 1.0 });

        for (let i = 0; i < 64; i++) {
            f.update(testInput[i]);
            expect(f.isPrimed()).withContext(`index ${i}`).toBeFalse();
        }

        f.update(testInput[64]);
        expect(f.isPrimed()).toBeTrue();
    });

    it('should pass through NaN', () => {
        const f = new RescaledFractalAdaptiveSimpleMovingAverage({ period: 4, normalSpeed: 30, priceScale: 1.0 });
        expect(Number.isNaN(f.update(Number.NaN))).toBeTrue();
    });

    it('should throw for period < 4', () => {
        expect(() => new RescaledFractalAdaptiveSimpleMovingAverage({ period: 2, normalSpeed: 30, priceScale: 1.0 })).toThrowError();
    });

    it('should throw for non-power-of-2 period', () => {
        expect(() => new RescaledFractalAdaptiveSimpleMovingAverage({ period: 6, normalSpeed: 30, priceScale: 1.0 })).toThrowError();
    });

    it('should throw for normalSpeed < 1', () => {
        expect(() => new RescaledFractalAdaptiveSimpleMovingAverage({ period: 4, normalSpeed: 0, priceScale: 1.0 })).toThrowError();
    });

    it('should return correct metadata', () => {
        const f = new RescaledFractalAdaptiveSimpleMovingAverage({ period: 64, normalSpeed: 30, priceScale: 1.0 });
        const meta = f.metadata();
        expect(meta.identifier).toBe(IndicatorIdentifier.RescaledFractalAdaptiveSimpleMovingAverage);
    });
});
