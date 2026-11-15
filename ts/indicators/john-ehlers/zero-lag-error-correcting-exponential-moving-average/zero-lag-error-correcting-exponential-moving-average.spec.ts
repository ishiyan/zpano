import { Bar } from '../../../entities/bar';
import { BarComponent } from '../../../entities/bar-component';
import { Quote } from '../../../entities/quote';
import { QuoteComponent } from '../../../entities/quote-component';
import { Scalar } from '../../../entities/scalar';
import { Trade } from '../../../entities/trade';
import { TradeComponent } from '../../../entities/trade-component';
import { Adaptivity } from '../../core/adaptivity';
import { descriptorOf } from '../../core/descriptors';
import { IndicatorIdentifier } from '../../core/indicator-identifier';
import { Shape } from '../../core/outputs/shape/shape';
import { createIndicator } from '../../factory/factory';
import { ZeroLagErrorCorrectingExponentialMovingAverageOutput } from './output';
import {
    equivalentLengthFromAlpha,
    zeroLagErrorCorrectingExponentialMovingAverageEmaMnemonic,
    zeroLagErrorCorrectingExponentialMovingAverageMnemonic,
} from './params';
import * as testdata from './testdata';
import { ZeroLagErrorCorrectingExponentialMovingAverage } from './zero-lag-error-correcting-exponential-moving-average';

interface ReferenceCase {
    length: number;
    gainLimit: number;
    expectedEc: number[];
    expectedEma: number[];
}

const REFERENCE_CASES: ReferenceCase[] = [
    { length: 20, gainLimit: 50, expectedEc: testdata.expectedL20_G50_EC, expectedEma: testdata.expectedL20_G50_EMA },
    { length: 12, gainLimit: 50, expectedEc: testdata.expectedL12_G50_EC, expectedEma: testdata.expectedL12_G50_EMA },
    { length: 32, gainLimit: 22, expectedEc: testdata.expectedL32_G22_EC, expectedEma: testdata.expectedL32_G22_EMA },
    { length: 20, gainLimit: 0, expectedEc: testdata.expectedL20_G0_EC, expectedEma: testdata.expectedL20_G0_EMA },
    { length: 20, gainLimit: 1, expectedEc: testdata.expectedL20_G1_EC, expectedEma: testdata.expectedL20_G1_EMA },
    { length: 20, gainLimit: 10, expectedEc: testdata.expectedL20_G10_EC, expectedEma: testdata.expectedL20_G10_EMA },
    { length: 20, gainLimit: 22, expectedEc: testdata.expectedL20_G22_EC, expectedEma: testdata.expectedL20_G22_EMA },
    { length: 20, gainLimit: 100, expectedEc: testdata.expectedL20_G100_EC, expectedEma: testdata.expectedL20_G100_EMA },
    { length: 1, gainLimit: 50, expectedEc: testdata.expectedL1_G50_EC, expectedEma: testdata.expectedL1_G50_EMA },
    { length: 2, gainLimit: 50, expectedEc: testdata.expectedL2_G50_EC, expectedEma: testdata.expectedL2_G50_EMA },
    { length: 5, gainLimit: 50, expectedEc: testdata.expectedL5_G50_EC, expectedEma: testdata.expectedL5_G50_EMA },
    { length: 10, gainLimit: 50, expectedEc: testdata.expectedL10_G50_EC, expectedEma: testdata.expectedL10_G50_EMA },
    { length: 40, gainLimit: 50, expectedEc: testdata.expectedL40_G50_EC, expectedEma: testdata.expectedL40_G50_EMA },
    { length: 100, gainLimit: 50, expectedEc: testdata.expectedL100_G50_EC, expectedEma: testdata.expectedL100_G50_EMA },
    { length: 252, gainLimit: 50, expectedEc: testdata.expectedL252_G50_EC, expectedEma: testdata.expectedL252_G50_EMA },
    { length: 20, gainLimit: 49, expectedEc: testdata.expectedL20_G49_EC, expectedEma: testdata.expectedL20_G49_EMA },
];

function articleBatch(
    prices: number[],
    length: number,
    gainLimit: number,
): Array<[number, number, number, number]> {
    const alpha = 2 / (length + 1);
    const emaSeries = new Array<number>(prices.length + 1).fill(0);
    const ecSeries = new Array<number>(prices.length + 1).fill(0);
    let bestGain = 0;
    const outputs: Array<[number, number, number, number]> = [];

    for (let step = 1; step <= prices.length; step++) {
        const price = prices[step - 1];
        emaSeries[step] = alpha * price + (1 - alpha) * emaSeries[step - 1];
        let leastError = 1_000_000.0;
        for (let integerGain = -gainLimit; integerGain <= gainLimit; integerGain++) {
            const gain = integerGain / 10.0;
            const trial = alpha * (emaSeries[step] + gain * (price - ecSeries[step - 1])) + (1 - alpha) * ecSeries[step - 1];
            const errorValue = price - trial;
            if (Math.abs(errorValue) < leastError) {
                leastError = Math.abs(errorValue);
                bestGain = gain;
            }
        }
        ecSeries[step] = alpha * (emaSeries[step] + bestGain * (price - ecSeries[step - 1])) + (1 - alpha) * ecSeries[step - 1];
        outputs.push([ecSeries[step], emaSeries[step], bestGain, leastError]);
    }
    return outputs;
}

function expectCloseOrNan(actual: number, expected: number, tolerance = 2e-12): void {
    if (Number.isNaN(expected)) {
        expect(actual).toBeNaN();
        return;
    }
    const scale = Math.max(1, Math.abs(expected));
    expect(Math.abs(actual - expected)).toBeLessThanOrEqual(tolerance * scale);
}

describe('ZeroLagErrorCorrectingExponentialMovingAverage', () => {
    it('matches all 16 reference fixtures and article batch equations', () => {
        for (const item of REFERENCE_CASES) {
            const stream = ZeroLagErrorCorrectingExponentialMovingAverage.fromLength({
                length: item.length,
                gainLimit: item.gainLimit,
            });
            const entityStream = ZeroLagErrorCorrectingExponentialMovingAverage.fromLength({
                length: item.length,
                gainLimit: item.gainLimit,
            });
            const reference = articleBatch(testdata.testInput, item.length, item.gainLimit);

            for (let barIndex = 0; barIndex < testdata.testInput.length; barIndex++) {
                const price = testdata.testInput[barIndex];
                const ec = stream.update(price);
                const ema = stream.ema();
                expectCloseOrNan(ec, item.expectedEc[barIndex]);
                expectCloseOrNan(ema, item.expectedEma[barIndex]);
                expectCloseOrNan(ec, reference[barIndex][0]);
                expectCloseOrNan(ema, reference[barIndex][1]);
                expect(stream.bestGain()).toBe(reference[barIndex][2]);
                expect(stream.leastError()).toBe(reference[barIndex][3]);

                const stamp = new Date(Date.UTC(2020, 0, 1 + barIndex));
                const entityOutput = entityStream.updateScalar(new Scalar({ time: stamp, value: price }));
                expect(entityOutput.length).toBe(2);
                const firstScalar = entityOutput[0] as Scalar;
                const secondScalar = entityOutput[1] as Scalar;
                expect(firstScalar.time.getTime()).toBe(stamp.getTime());
                expect(secondScalar.time.getTime()).toBe(stamp.getTime());
                expect(firstScalar.value).toBe(ec);
                expect(secondScalar.value).toBe(ema);
            }
        }
    });

    it('primes on the first finite observation from explicit zero states and resets cleanly', () => {
        const stream = ZeroLagErrorCorrectingExponentialMovingAverage.default();
        expect(stream.isPrimed()).toBe(false);
        expect(stream.warmupBars()).toBe(0);
        expect(stream.samplesSeen()).toBe(0);
        expect(stream.value()).toBeNaN();
        expect(stream.ema()).toBeNaN();
        expect(stream.bestGain()).toBeNaN();
        expect(stream.leastError()).toBeNaN();
        expect(stream.alpha()).toBe(2 / 21);

        const ecFirst = stream.update(100);
        const alpha = 2 / 21;
        const expectedEma = alpha * 100;
        const expectedEc = alpha * (expectedEma + 5.0 * 100);
        expect(stream.isPrimed()).toBe(true);
        expect(stream.samplesSeen()).toBe(1);
        expect(ecFirst).toBe(expectedEc);
        expect(stream.ema()).toBe(expectedEma);
        expect(stream.bestGain()).toBe(5.0);
        expect(stream.leastError()).toBe(Math.abs(100 - expectedEc));

        stream.reset();
        expect(stream.isPrimed()).toBe(false);
        expect(stream.samplesSeen()).toBe(0);
        expect(stream.value()).toBeNaN();
        expect(stream.ema()).toBeNaN();
        expect(stream.bestGain()).toBeNaN();
        expect(stream.leastError()).toBeNaN();
        expect(stream.update(100)).toBe(expectedEc);
    });

    it('reduces to two cascaded EMAs when gainLimit is zero', () => {
        const stream = new ZeroLagErrorCorrectingExponentialMovingAverage({ length: 20, gainLimit: 0 });
        const alpha = 2 / 21;
        let ema = 0;
        let cascaded = 0;
        for (const price of testdata.testInput) {
            ema = alpha * price + (1 - alpha) * ema;
            cascaded = alpha * ema + (1 - alpha) * cascaded;
            const ec = stream.update(price);
            expect(ec).toBe(cascaded);
            expect(stream.ema()).toBe(ema);
            expect(stream.bestGain()).toBe(0);
        }
    });

    it('preserves strict tie-breaking and literal 1_000_000 leastError sentinel', () => {
        const zeroConstant = new ZeroLagErrorCorrectingExponentialMovingAverage({ length: 20, gainLimit: 50 });
        expect(zeroConstant.update(0)).toBe(0);
        expect(zeroConstant.bestGain()).toBe(-5.0);
        expect(zeroConstant.leastError()).toBe(0);

        const hugeStep = new ZeroLagErrorCorrectingExponentialMovingAverage({ length: 20, gainLimit: 50 });
        const hugePrice = 1e9;
        const ec = hugeStep.update(hugePrice);
        const alpha = 2 / 21;
        expect(hugeStep.bestGain()).toBe(0);
        expect(hugeStep.leastError()).toBe(1_000_000.0);
        expect(ec).toBe(alpha * (alpha * hugePrice));
    });

    it('skips NaN observations without mutating state and rejects invalid or overflowing samples atomically', () => {
        const stream = ZeroLagErrorCorrectingExponentialMovingAverage.default();
        const stamp = new Date(Date.UTC(2020, 0, 1));
        const prePrimedNan = stream.updateScalar(new Scalar({ time: stamp, value: Number.NaN }));
        expect((prePrimedNan[0] as Scalar).value).toBeNaN();
        expect((prePrimedNan[1] as Scalar).value).toBeNaN();
        expect(stream.isPrimed()).toBe(false);

        const acceptedEc = stream.update(testdata.testInput[0]);
        const acceptedEma = stream.ema();
        const acceptedGain = stream.bestGain();
        const acceptedLeastError = stream.leastError();

        expect(stream.update(Number.NaN)).toBeNaN();
        expect(stream.samplesSeen()).toBe(1);
        expect(stream.value()).toBe(acceptedEc);
        expect(stream.ema()).toBe(acceptedEma);
        expect(stream.bestGain()).toBe(acceptedGain);
        expect(stream.leastError()).toBe(acceptedLeastError);

        for (const invalidSample of [
            Number.POSITIVE_INFINITY,
            Number.NEGATIVE_INFINITY,
            1e308,
            true as unknown as number,
            null as unknown as number,
            'invalid' as unknown as number,
        ]) {
            expect(() => stream.update(invalidSample)).toThrow();
            expect(stream.samplesSeen()).toBe(1);
            expect(stream.value()).toBe(acceptedEc);
            expect(stream.ema()).toBe(acceptedEma);
            expect(stream.bestGain()).toBe(acceptedGain);
            expect(stream.leastError()).toBe(acceptedLeastError);
        }
    });

    it('supports smoothingFactor variant, metadata, entity updates, and factory validation', () => {
        const fromLength = ZeroLagErrorCorrectingExponentialMovingAverage.fromLength({ length: 20, gainLimit: 50 });
        const fromAlpha = ZeroLagErrorCorrectingExponentialMovingAverage.fromSmoothingFactor({
            smoothingFactor: 2 / 21,
            gainLimit: 50,
        });
        expect(fromAlpha.length()).toBe(20);
        for (const price of testdata.testInput) {
            expect(fromAlpha.update(price)).toBe(fromLength.update(price));
            expect(fromAlpha.ema()).toBe(fromLength.ema());
        }

        const unitAlpha = ZeroLagErrorCorrectingExponentialMovingAverage.fromSmoothingFactor({
            smoothingFactor: 1.0,
            gainLimit: 50,
        });
        expect(unitAlpha.length()).toBe(1);
        expect(unitAlpha.update(91.5)).toBe(91.5);

        expect(equivalentLengthFromAlpha(1e-9)).toBe(Number.MAX_SAFE_INTEGER);
        expect(equivalentLengthFromAlpha(0.2)).toBe(9);

        const meta = fromLength.metadata();
        const desc = descriptorOf(IndicatorIdentifier.ZeroLagErrorCorrectingExponentialMovingAverage);
        expect(desc).toBeDefined();
        expect(desc!.adaptivity).toBe(Adaptivity.Adaptive);
        expect(meta.identifier).toBe(IndicatorIdentifier.ZeroLagErrorCorrectingExponentialMovingAverage);
        expect(meta.mnemonic).toBe('zecema(20, 50)');
        expect(meta.outputs.length).toBe(2);
        expect(meta.outputs[0].kind).toBe(ZeroLagErrorCorrectingExponentialMovingAverageOutput.Value);
        expect(meta.outputs[0].shape).toBe(Shape.Scalar);
        expect(meta.outputs[0].mnemonic).toBe('zecema(20, 50)');
        expect(meta.outputs[1].kind).toBe(ZeroLagErrorCorrectingExponentialMovingAverageOutput.Ema);
        expect(meta.outputs[1].shape).toBe(Shape.Scalar);
        expect(meta.outputs[1].mnemonic).toBe('zecemaEma(20, 50)');

        const stamp = new Date(Date.UTC(2020, 0, 1));
        for (const entityOutput of [
            ZeroLagErrorCorrectingExponentialMovingAverage.default().updateBar(
                new Bar({ time: stamp, open: 100, high: 101, low: 99, close: 100, volume: 10 }),
            ),
            ZeroLagErrorCorrectingExponentialMovingAverage.default().updateQuote(
                new Quote({ time: stamp, bidPrice: 100, askPrice: 100, bidSize: 1, askSize: 1 }),
            ),
            ZeroLagErrorCorrectingExponentialMovingAverage.default().updateTrade(
                new Trade({ time: stamp, price: 100, volume: 1 }),
            ),
        ]) {
            expect(entityOutput.length).toBe(2);
            expect(Number.isFinite((entityOutput[0] as Scalar).value)).toBe(true);
            expect(Number.isFinite((entityOutput[1] as Scalar).value)).toBe(true);
        }

        expect(
            zeroLagErrorCorrectingExponentialMovingAverageMnemonic({
                length: 32,
                gainLimit: 22,
                barComponent: BarComponent.Open,
                quoteComponent: QuoteComponent.Ask,
            }),
        ).toBe('zecema(32, 22, o, a)');
        expect(
            zeroLagErrorCorrectingExponentialMovingAverageEmaMnemonic({
                smoothingFactor: 0.2,
                gainLimit: 22,
                barComponent: BarComponent.Typical,
                tradeComponent: TradeComponent.Volume,
            }),
        ).toBe('zecemaEma(9, 0.20000000, 22, hlc/3, v)');

        for (const invalidParams of [
            { length: 0 },
            { length: 1.5 },
            { length: 1e20 },
            { gainLimit: -1 },
            { gainLimit: 2.5 },
            { smoothingFactor: 0 },
            { smoothingFactor: -0.1 },
            { smoothingFactor: 1.1 },
            { smoothingFactor: 1e-30 },
            { gainStep: 0.1 },
            { length: 20, smoothingFactor: 0.2 },
        ]) {
            expect(() => new ZeroLagErrorCorrectingExponentialMovingAverage(invalidParams as never)).toThrow();
        }

        const factoryDefault = createIndicator(IndicatorIdentifier.ZeroLagErrorCorrectingExponentialMovingAverage);
        expect(factoryDefault.metadata().mnemonic).toBe('zecema(20, 50)');
        const factoryZeroGain = createIndicator(IndicatorIdentifier.ZeroLagErrorCorrectingExponentialMovingAverage, {
            length: 12,
            gainLimit: 0,
            barComponent: BarComponent.Median,
        });
        expect(factoryZeroGain.metadata().mnemonic).toBe('zecema(12, 0, hl/2)');
        const factoryAlpha = createIndicator(IndicatorIdentifier.ZeroLagErrorCorrectingExponentialMovingAverage, {
            smoothingFactor: 0.2,
            gainLimit: 22,
        });
        expect(factoryAlpha.metadata().mnemonic).toBe('zecema(9, 0.20000000, 22)');

        for (const invalidFactoryParams of [
            { gainStep: 0.1 },
            { length: 20, smoothingFactor: 0.2 },
            { length: 0 },
            { length: 1.5 },
            { gainLimit: -1 },
            { gainLimit: 2.5 },
            { smoothingFactor: 0 },
        ]) {
            expect(() =>
                createIndicator(
                    IndicatorIdentifier.ZeroLagErrorCorrectingExponentialMovingAverage,
                    invalidFactoryParams,
                ),
            ).toThrow();
        }
    });
});
