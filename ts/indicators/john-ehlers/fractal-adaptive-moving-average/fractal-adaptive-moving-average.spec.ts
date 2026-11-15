import { Bar } from '../../../entities/bar';
import { BarComponent } from '../../../entities/bar-component';
import { Quote } from '../../../entities/quote';
import { QuoteComponent } from '../../../entities/quote-component';
import { Scalar } from '../../../entities/scalar';
import { Trade } from '../../../entities/trade';
import { TradeComponent } from '../../../entities/trade-component';
import { IndicatorIdentifier } from '../../core/indicator-identifier';
import { Shape } from '../../core/outputs/shape/shape';
import { createIndicator } from '../../factory/factory';
import { FractalAdaptiveMovingAverage } from './fractal-adaptive-moving-average';
import { FractalAdaptiveMovingAverageOutput } from './output';
import {
    FractalAdaptiveMovingAverageParams,
    fractalAdaptiveMovingAverageDimensionMnemonic,
    fractalAdaptiveMovingAverageMnemonic,
} from './params';
import * as testdata from './testdata';

interface ReferenceCase {
    params: FractalAdaptiveMovingAverageParams;
    expectedValues: number[];
    expectedDimensions: number[];
}

const REFERENCE_CASES: ReferenceCase[] = [
    { params: { length: 16, slowestSmoothingFactor: 0.01, coefficientMode: 1 }, expectedValues: testdata.expectedL16_A0_01_C1_VALUE, expectedDimensions: testdata.expectedL16_A0_01_C1_FDIM },
    { params: { length: 2, slowestSmoothingFactor: 0.01, coefficientMode: 1 }, expectedValues: testdata.expectedL2_A0_01_C1_VALUE, expectedDimensions: testdata.expectedL2_A0_01_C1_FDIM },
    { params: { length: 4, slowestSmoothingFactor: 0.01, coefficientMode: 1 }, expectedValues: testdata.expectedL4_A0_01_C1_VALUE, expectedDimensions: testdata.expectedL4_A0_01_C1_FDIM },
    { params: { length: 8, slowestSmoothingFactor: 0.01, coefficientMode: 1 }, expectedValues: testdata.expectedL8_A0_01_C1_VALUE, expectedDimensions: testdata.expectedL8_A0_01_C1_FDIM },
    { params: { length: 20, slowestSmoothingFactor: 0.01, coefficientMode: 1 }, expectedValues: testdata.expectedL20_A0_01_C1_VALUE, expectedDimensions: testdata.expectedL20_A0_01_C1_FDIM },
    { params: { length: 30, slowestSmoothingFactor: 0.01, coefficientMode: 1 }, expectedValues: testdata.expectedL30_A0_01_C1_VALUE, expectedDimensions: testdata.expectedL30_A0_01_C1_FDIM },
    { params: { length: 64, slowestSmoothingFactor: 0.01, coefficientMode: 1 }, expectedValues: testdata.expectedL64_A0_01_C1_VALUE, expectedDimensions: testdata.expectedL64_A0_01_C1_FDIM },
    { params: { length: 128, slowestSmoothingFactor: 0.01, coefficientMode: 1 }, expectedValues: testdata.expectedL128_A0_01_C1_VALUE, expectedDimensions: testdata.expectedL128_A0_01_C1_FDIM },
    { params: { length: 250, slowestSmoothingFactor: 0.01, coefficientMode: 1 }, expectedValues: testdata.expectedL250_A0_01_C1_VALUE, expectedDimensions: testdata.expectedL250_A0_01_C1_FDIM },
    { params: { length: 252, slowestSmoothingFactor: 0.01, coefficientMode: 1 }, expectedValues: testdata.expectedL252_A0_01_C1_VALUE, expectedDimensions: testdata.expectedL252_A0_01_C1_FDIM },
    { params: { length: 254, slowestSmoothingFactor: 0.01, coefficientMode: 1 }, expectedValues: testdata.expectedL254_A0_01_C1_VALUE, expectedDimensions: testdata.expectedL254_A0_01_C1_FDIM },
    { params: { length: 16, slowestSmoothingFactor: 0.001, coefficientMode: 1 }, expectedValues: testdata.expectedL16_A0_001_C1_VALUE, expectedDimensions: testdata.expectedL16_A0_001_C1_FDIM },
    { params: { length: 16, slowestSmoothingFactor: 0.1, coefficientMode: 1 }, expectedValues: testdata.expectedL16_A0_1_C1_VALUE, expectedDimensions: testdata.expectedL16_A0_1_C1_FDIM },
    { params: { length: 16, slowestSmoothingFactor: 1.0, coefficientMode: 1 }, expectedValues: testdata.expectedL16_A1_C1_VALUE, expectedDimensions: testdata.expectedL16_A1_C1_FDIM },
    { params: { length: 16, slowestSmoothingFactor: 0.01, coefficientMode: 0 }, expectedValues: testdata.expectedL16_A0_01_C0_VALUE, expectedDimensions: testdata.expectedL16_A0_01_C0_FDIM },
    { params: { length: 30, slowestSmoothingFactor: 0.01, coefficientMode: 0 }, expectedValues: testdata.expectedL30_A0_01_C0_VALUE, expectedDimensions: testdata.expectedL30_A0_01_C0_FDIM },
];

function articleBatch(
    prices: number[],
    length: number,
    slowest = 0.01,
    mode = 0,
    highs?: number[],
    lows?: number[],
): Array<[number, number]> {
    const highSeries = highs ?? prices;
    const lowSeries = lows ?? prices;
    const half = Math.floor(length / 2);
    const scale = mode === 1 ? Math.log(slowest) : -4.6;
    let dimension = 0;
    const filtered = new Array<number>(prices.length).fill(0);
    const outputs: Array<[number, number]> = [];

    for (let index = 0; index < prices.length; index++) {
        const price = prices[index];
        if (index + 1 >= length) {
            let max1 = highSeries[index];
            let min1 = lowSeries[index];
            for (let offset = 1; offset < half; offset++) {
                if (highSeries[index - offset] > max1) max1 = highSeries[index - offset];
                if (lowSeries[index - offset] < min1) min1 = lowSeries[index - offset];
            }
            let max2 = highSeries[index - half];
            let min2 = lowSeries[index - half];
            for (let offset = half + 1; offset < length; offset++) {
                if (highSeries[index - offset] > max2) max2 = highSeries[index - offset];
                if (lowSeries[index - offset] < min2) min2 = lowSeries[index - offset];
            }
            const max3 = Math.max(max1, max2);
            const min3 = Math.min(min1, min2);
            const n1 = (max1 - min1) / half;
            const n2 = (max2 - min2) / half;
            const n3 = (max3 - min3) / length;
            if (n1 > 0 && n2 > 0 && n3 > 0) {
                dimension = (Math.log(n1 + n2) - Math.log(n3)) / Math.log(2);
            }
            const rawAlpha = dimension <= 1 ? 1 : Math.exp(scale * (dimension - 1));
            const alpha = Math.max(slowest, Math.min(1, rawAlpha));
            filtered[index] = index + 1 < length + 1
                ? price
                : alpha * price + (1 - alpha) * filtered[index - 1];
            outputs.push([filtered[index], dimension]);
        } else {
            filtered[index] = price;
            outputs.push([Number.NaN, Number.NaN]);
        }
    }
    return outputs;
}

describe('FractalAdaptiveMovingAverage', () => {
    it('matches all 16 confirmed reference fixtures', () => {
        expect(testdata.testInput.length).toBe(252);
        for (const item of REFERENCE_CASES) {
            const indicator = new FractalAdaptiveMovingAverage(item.params);
            const warmup = (item.params.length ?? 16) - 1;
            expect(indicator.warmupBars()).toBe(warmup);
            expect(indicator.isPrimed()).toBeFalse();
            expect(Number.isNaN(indicator.value())).toBeTrue();
            expect(Number.isNaN(indicator.fractalDimension())).toBeTrue();
            expect(Number.isNaN(indicator.alpha())).toBeTrue();

            for (let index = 0; index < testdata.testInput.length; index++) {
                const actualValue = indicator.update(testdata.testInput[index]);
                const actualDimension = indicator.fractalDimension();
                const targetValue = item.expectedValues[index];
                const targetDimension = item.expectedDimensions[index];

                if (index < warmup) {
                    expect(indicator.isPrimed()).toBeFalse();
                    expect(Number.isNaN(actualValue)).toBeTrue();
                    expect(Number.isNaN(actualDimension)).toBeTrue();
                    expect(Number.isNaN(targetValue)).toBeTrue();
                    expect(Number.isNaN(targetDimension)).toBeTrue();
                } else {
                    expect(indicator.isPrimed()).toBeTrue();
                    const valueBound = Math.max(1e-12, Math.abs(targetValue) * 1e-12);
                    const dimBound = Math.max(1e-12, Math.abs(targetDimension) * 1e-12);
                    expect(Math.abs(actualValue - targetValue)).toBeLessThanOrEqual(valueBound);
                    expect(Math.abs(actualDimension - targetDimension)).toBeLessThanOrEqual(dimBound);
                    expect(Math.abs(indicator.value() - actualValue)).toBeLessThanOrEqual(1e-15);
                }
            }
        }
    });

    it('matches independent batch oracle on midpoint and bar high/low ranges', () => {
        const midpoints = testdata.testHigh.map((high, index) => (high + testdata.testLow[index]) / 2);
        const stream = new FractalAdaptiveMovingAverage({ length: 16, slowestSmoothingFactor: 0.01, coefficientMode: 0 });
        const expected = articleBatch(midpoints, 16, 0.01, 0, testdata.testHigh, testdata.testLow);

        for (let index = 0; index < midpoints.length; index++) {
            const actualValue = stream.update(midpoints[index], testdata.testHigh[index], testdata.testLow[index]);
            const actualDimension = stream.fractalDimension();
            const [targetValue, targetDimension] = expected[index];
            if (Number.isNaN(targetValue)) {
                expect(Number.isNaN(actualValue)).toBeTrue();
                expect(Number.isNaN(actualDimension)).toBeTrue();
            } else {
                expect(Math.abs(actualValue - targetValue)).toBeLessThanOrEqual(1e-12);
                expect(Math.abs(actualDimension - targetDimension)).toBeLessThanOrEqual(1e-12);
            }
        }
    });

    it('preserves startup seed, unclamped ramp/gap dimensions, flat-box guard, and mode endpoints', () => {
        const seeded = FractalAdaptiveMovingAverage.default();
        let seedValue = 0;
        for (let index = 0; index < 20; index++) {
            const price = testdata.testInput[index];
            const actualValue = seeded.update(price);
            if (index === 15) {
                expect(actualValue).toBe(price);
                seedValue = actualValue;
            }
            if (index === 16) {
                const currentAlpha = seeded.alpha();
                expect(actualValue).toBeCloseTo(currentAlpha * price + (1 - currentAlpha) * seedValue, 14);
            }
        }

        const expectedRampDimension = Math.log2(28 / 15);
        for (const offset of [100, 0, -100]) {
            const ramp = FractalAdaptiveMovingAverage.default();
            for (let step = 0; step < 40; step++) {
                const actualValue = ramp.update(offset + step);
                if (step >= 15) {
                    expect(ramp.fractalDimension()).toBeCloseTo(expectedRampDimension, 13);
                    expect(ramp.fractalDimension()).toBeLessThan(1);
                    expect(ramp.alpha()).toBe(1);
                    expect(actualValue).toBe(offset + step);
                }
            }
        }

        const gap = new FractalAdaptiveMovingAverage({ length: 4 });
        let gapValue = Number.NaN;
        for (const price of [0, 1, 100, 101]) {
            gapValue = gap.update(price);
        }
        expect(gap.fractalDimension()).toBeCloseTo(Math.log2(4 / 101), 13);
        expect(gap.fractalDimension()).toBeLessThan(0);
        expect(gap.alpha()).toBe(1);
        expect(gapValue).toBe(101);

        const hold = new FractalAdaptiveMovingAverage({ length: 4 });
        for (const price of [0, 1, 0, 1]) {
            hold.update(price);
        }
        expect(hold.fractalDimension()).toBe(2);
        for (let step = 0; step < 8; step++) {
            hold.update(1);
            expect(hold.fractalDimension()).toBe(2);
        }

        const precise = new FractalAdaptiveMovingAverage({ length: 4, coefficientMode: 1 });
        const article = new FractalAdaptiveMovingAverage({ length: 4, coefficientMode: 0 });
        for (let step = 0; step < 5; step++) {
            precise.update(1, 2, 0);
            article.update(1, 2, 0);
        }
        expect(precise.alpha()).toBeCloseTo(0.01, 15);
        expect(article.alpha()).toBe(Math.exp(-4.6));
        expect(article.alpha()).toBeGreaterThan(precise.alpha());
    });

    it('supports affine invariance, NaN gaps, reset, and atomic validation', () => {
        const original = FractalAdaptiveMovingAverage.default();
        const transformed = FractalAdaptiveMovingAverage.default();
        const withGaps = FractalAdaptiveMovingAverage.default();

        for (const price of testdata.testInput) {
            const leftVal = original.update(price);
            const rightVal = transformed.update(3 * price - 1000);
            const gapVal = withGaps.update(price);
            const seenBefore = withGaps.samplesSeen();
            const savedVal = withGaps.value();

            expect(Number.isNaN(withGaps.update(Number.NaN))).toBeTrue();
            expect(Number.isNaN(withGaps.update(price, Number.NaN, price))).toBeTrue();
            expect(withGaps.samplesSeen()).toBe(seenBefore);
            if (Number.isNaN(savedVal)) {
                expect(Number.isNaN(withGaps.value())).toBeTrue();
            } else {
                expect(withGaps.value()).toBe(savedVal);
            }

            if (!Number.isNaN(leftVal)) {
                expect(leftVal).toBe(gapVal);
                expect(original.fractalDimension()).toBeCloseTo(transformed.fractalDimension(), 11);
                expect(3 * leftVal - 1000).toBeCloseTo(rightVal, 10);
            }
        }

        withGaps.reset();
        expect(withGaps.isPrimed()).toBeFalse();
        expect(withGaps.samplesSeen()).toBe(0);
        const fresh = FractalAdaptiveMovingAverage.default();
        for (const price of testdata.testInput) {
            const a = withGaps.update(price);
            const b = fresh.update(price);
            if (Number.isNaN(b)) {
                expect(Number.isNaN(a)).toBeTrue();
            } else {
                expect(a).toBe(b);
            }
        }

        const atomic = new FractalAdaptiveMovingAverage({ length: 2 });
        atomic.update(0);
        const invalidUpdates: Array<() => void> = [
            () => atomic.update(Number.POSITIVE_INFINITY),
            () => atomic.update(true as unknown as number),
            () => atomic.update(1, 2),
            () => atomic.update(1, 0, 2),
            () => atomic.update(1, Number.POSITIVE_INFINITY, 0),
            () => atomic.update(0, 1e308, -1e308),
        ];
        for (const call of invalidUpdates) {
            expect(call).toThrow();
            expect(atomic.samplesSeen()).toBe(1);
            expect(atomic.isPrimed()).toBeFalse();
        }

        const invalidConfigs: FractalAdaptiveMovingAverageParams[] = [
            { length: 1 },
            { length: 17 },
            { length: 16.5 },
            { length: true as unknown as number },
            { slowestSmoothingFactor: 0 },
            { slowestSmoothingFactor: -0.1 },
            { slowestSmoothingFactor: 1.1 },
            { slowestSmoothingFactor: Number.NaN },
            { slowestSmoothingFactor: Number.POSITIVE_INFINITY },
            { coefficientMode: 2 },
            { coefficientMode: 0, slowestSmoothingFactor: 0.1 },
        ];
        for (const bad of invalidConfigs) {
            expect(() => new FractalAdaptiveMovingAverage(bad)).toThrow();
        }
    });

    it('produces expected metadata, mnemonics, factory instances, and entity outputs', () => {
        expect(FractalAdaptiveMovingAverageOutput.Value).toBe(0);
        expect(FractalAdaptiveMovingAverageOutput.Fdim).toBe(1);

        const defaultInd = FractalAdaptiveMovingAverage.default();
        const meta = defaultInd.metadata();
        expect(meta.identifier).toBe(IndicatorIdentifier.FractalAdaptiveMovingAverage);
        expect(meta.mnemonic).toBe('frama(16, 0.01)');
        expect(meta.description).toBe('Fractal adaptive moving average frama(16, 0.01)');
        expect(meta.outputs.length).toBe(2);
        expect(meta.outputs[0].kind).toBe(FractalAdaptiveMovingAverageOutput.Value);
        expect(meta.outputs[0].shape).toBe(Shape.Scalar);
        expect(meta.outputs[0].mnemonic).toBe('frama(16, 0.01)');
        expect(meta.outputs[1].kind).toBe(FractalAdaptiveMovingAverageOutput.Fdim);
        expect(meta.outputs[1].shape).toBe(Shape.Scalar);
        expect(meta.outputs[1].mnemonic).toBe('framaDim(16, 0.01)');

        const custom: FractalAdaptiveMovingAverageParams = {
            length: 30,
            slowestSmoothingFactor: 0.01,
            coefficientMode: 0,
            barComponent: BarComponent.Median,
            quoteComponent: QuoteComponent.Bid,
            tradeComponent: TradeComponent.Volume,
        };
        expect(fractalAdaptiveMovingAverageMnemonic(custom)).toBe('frama(30, 0.01, mode=0, hl/2, b, v)');
        expect(fractalAdaptiveMovingAverageDimensionMnemonic(custom)).toBe('framaDim(30, 0.01, mode=0, hl/2, b, v)');

        const fromFactory = createIndicator(IndicatorIdentifier.FractalAdaptiveMovingAverage, {
            length: 30,
            slowestSmoothingFactor: 0.01,
            coefficientMode: 0,
        });
        expect(fromFactory.metadata().mnemonic).toBe('frama(30, 0.01, mode=0)');

        const stamp = new Date('2026-01-02T00:00:00Z');
        const entityInd = new FractalAdaptiveMovingAverage({ length: 2 });
        entityInd.updateBar(new Bar({ time: stamp, open: 10, high: 12, low: 9, close: 11, volume: 100 }));
        const barOut = entityInd.updateBar(new Bar({ time: stamp, open: 11, high: 14, low: 10, close: 13, volume: 100 }));
        expect((barOut[0] as Scalar).value).toBe(13);

        const skipped = entityInd.updateScalar(new Scalar({ time: stamp, value: Number.NaN }));
        expect(Number.isNaN((skipped[0] as Scalar).value)).toBeTrue();
        expect(Number.isNaN((skipped[1] as Scalar).value)).toBeTrue();

        const quoteInd = new FractalAdaptiveMovingAverage({ length: 2 });
        quoteInd.updateQuote(new Quote({ time: stamp, bidPrice: 9, askPrice: 11, bidSize: 1, askSize: 1 }));
        const quoteOut = quoteInd.updateQuote(new Quote({ time: stamp, bidPrice: 10, askPrice: 12, bidSize: 1, askSize: 1 }));
        expect((quoteOut[0] as Scalar).value).toBe(11);

        const tradeInd = new FractalAdaptiveMovingAverage({ length: 2 });
        tradeInd.updateTrade(new Trade({ time: stamp, price: 5, volume: 1 }));
        const tradeOut = tradeInd.updateTrade(new Trade({ time: stamp, price: 7, volume: 1 }));
        expect((tradeOut[0] as Scalar).value).toBe(7);
    });
});
