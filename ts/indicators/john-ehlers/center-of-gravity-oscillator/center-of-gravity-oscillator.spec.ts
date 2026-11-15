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
import { CenterOfGravityOscillator } from './center-of-gravity-oscillator';
import { CenterOfGravityOscillatorOutput } from './output';
import {
    CenterOfGravityOscillatorParams,
    centerOfGravityOscillatorMnemonic,
    centerOfGravityOscillatorTriggerMnemonic,
} from './params';
import * as testdata from './testdata';

interface ReferenceCase {
    length: number;
    centered: boolean;
    expectedValue: number[];
    expectedTrigger: number[];
}

const REFERENCE_CASES: ReferenceCase[] = [
    { length: 1, centered: false, expectedValue: testdata.expectedL1_C0_CG, expectedTrigger: testdata.expectedL1_C0_TR },
    { length: 1, centered: true, expectedValue: testdata.expectedL1_C1_CG, expectedTrigger: testdata.expectedL1_C1_TR },
    { length: 2, centered: false, expectedValue: testdata.expectedL2_C0_CG, expectedTrigger: testdata.expectedL2_C0_TR },
    { length: 2, centered: true, expectedValue: testdata.expectedL2_C1_CG, expectedTrigger: testdata.expectedL2_C1_TR },
    { length: 3, centered: false, expectedValue: testdata.expectedL3_C0_CG, expectedTrigger: testdata.expectedL3_C0_TR },
    { length: 3, centered: true, expectedValue: testdata.expectedL3_C1_CG, expectedTrigger: testdata.expectedL3_C1_TR },
    { length: 5, centered: false, expectedValue: testdata.expectedL5_C0_CG, expectedTrigger: testdata.expectedL5_C0_TR },
    { length: 5, centered: true, expectedValue: testdata.expectedL5_C1_CG, expectedTrigger: testdata.expectedL5_C1_TR },
    { length: 8, centered: false, expectedValue: testdata.expectedL8_C0_CG, expectedTrigger: testdata.expectedL8_C0_TR },
    { length: 8, centered: true, expectedValue: testdata.expectedL8_C1_CG, expectedTrigger: testdata.expectedL8_C1_TR },
    { length: 10, centered: false, expectedValue: testdata.expectedL10_C0_CG, expectedTrigger: testdata.expectedL10_C0_TR },
    { length: 10, centered: true, expectedValue: testdata.expectedL10_C1_CG, expectedTrigger: testdata.expectedL10_C1_TR },
    { length: 14, centered: false, expectedValue: testdata.expectedL14_C0_CG, expectedTrigger: testdata.expectedL14_C0_TR },
    { length: 14, centered: true, expectedValue: testdata.expectedL14_C1_CG, expectedTrigger: testdata.expectedL14_C1_TR },
    { length: 20, centered: false, expectedValue: testdata.expectedL20_C0_CG, expectedTrigger: testdata.expectedL20_C0_TR },
    { length: 20, centered: true, expectedValue: testdata.expectedL20_C1_CG, expectedTrigger: testdata.expectedL20_C1_TR },
    { length: 32, centered: false, expectedValue: testdata.expectedL32_C0_CG, expectedTrigger: testdata.expectedL32_C0_TR },
    { length: 32, centered: true, expectedValue: testdata.expectedL32_C1_CG, expectedTrigger: testdata.expectedL32_C1_TR },
    { length: 64, centered: false, expectedValue: testdata.expectedL64_C0_CG, expectedTrigger: testdata.expectedL64_C0_TR },
    { length: 64, centered: true, expectedValue: testdata.expectedL64_C1_CG, expectedTrigger: testdata.expectedL64_C1_TR },
    { length: 126, centered: false, expectedValue: testdata.expectedL126_C0_CG, expectedTrigger: testdata.expectedL126_C0_TR },
    { length: 126, centered: true, expectedValue: testdata.expectedL126_C1_CG, expectedTrigger: testdata.expectedL126_C1_TR },
    { length: 252, centered: false, expectedValue: testdata.expectedL252_C0_CG, expectedTrigger: testdata.expectedL252_C0_TR },
    { length: 252, centered: true, expectedValue: testdata.expectedL252_C1_CG, expectedTrigger: testdata.expectedL252_C1_TR },
];

function publishedLoopOutputs(prices: number[], length: number, centered: boolean): Array<[number, number]> {
    let previous = 0;
    const outputs: Array<[number, number]> = [];
    for (let step = 0; step < prices.length; step++) {
        if (step < length - 1) {
            outputs.push([Number.NaN, Number.NaN]);
            continue;
        }
        const trigger = step >= length ? previous : Number.NaN;
        let numerator = 0;
        let denominator = 0;
        for (let lag = 0; lag < length; lag++) {
            numerator += (1 + lag) * prices[step - lag];
            denominator += prices[step - lag];
        }
        if (denominator !== 0) {
            previous = -numerator / denominator;
            if (centered) {
                previous += (length + 1) / 2;
            }
        }
        outputs.push([previous, trigger]);
    }
    return outputs;
}

function expectCloseOrNaN(actual: number, expected: number, tolerance = 1e-13): void {
    if (Number.isNaN(expected)) {
        expect(Number.isNaN(actual)).toBeTrue();
        return;
    }
    expect(Number.isNaN(actual)).toBeFalse();
    const bound = Math.max(tolerance, Math.abs(expected) * tolerance);
    expect(Math.abs(actual - expected)).toBeLessThanOrEqual(bound);
}

describe('CenterOfGravityOscillator', () => {
    it('matches all 24 confirmed reference cases and the published EasyLanguage loop', () => {
        expect(testdata.testInput.length).toBe(252);

        for (const item of REFERENCE_CASES) {
            const params: CenterOfGravityOscillatorParams = { length: item.length, centered: item.centered };
            const indicator = new CenterOfGravityOscillator(params);
            const barIndicator = CenterOfGravityOscillator.fromParams(params);
            const oracle = publishedLoopOutputs(testdata.testInput, item.length, item.centered);

            expect(indicator.length()).toBe(item.length);
            expect(indicator.centered()).toBe(item.centered);
            expect(indicator.isPrimed()).toBeFalse();
            expect(indicator.isTriggerPrimed()).toBeFalse();
            expect(Number.isNaN(indicator.value())).toBeTrue();
            expect(Number.isNaN(indicator.trigger())).toBeTrue();

            for (let index = 0; index < testdata.testInput.length; index++) {
                const price = testdata.testInput[index];
                const actualValue = indicator.update(price);
                const actualTrigger = indicator.trigger();

                expect(indicator.isPrimed()).toBe(index >= item.length - 1);
                expect(indicator.isTriggerPrimed()).toBe(index >= item.length);
                expectCloseOrNaN(actualValue, item.expectedValue[index]);
                expectCloseOrNaN(actualTrigger, item.expectedTrigger[index]);
                expectCloseOrNaN(indicator.value(), actualValue, 0);

                expectCloseOrNaN(actualValue, oracle[index][0], 0);
                expectCloseOrNaN(actualTrigger, oracle[index][1], 0);

                const stamp = new Date(Date.UTC(2020, 0, 1 + index));
                const barOutput = barIndicator.updateBar(
                    new Bar({
                        time: stamp,
                        open: testdata.testOpen[index],
                        high: testdata.testHigh[index],
                        low: testdata.testLow[index],
                        close: price,
                        volume: 1000,
                    }),
                );
                expect(barOutput.length).toBe(2);
                const outValue = barOutput[0] as Scalar;
                const outTrigger = barOutput[1] as Scalar;
                expect(outValue.time).toEqual(stamp);
                expect(outTrigger.time).toEqual(stamp);
                expectCloseOrNaN(outValue.value, actualValue, 0);
                expectCloseOrNaN(outTrigger.value, actualTrigger, 0);
            }
        }
    });

    it('verifies newest-first weights, negation, constant baseline, scale invariance, and crossover', () => {
        const article = new CenterOfGravityOscillator({ length: 3, centered: false });
        const book = new CenterOfGravityOscillator({ length: 3, centered: true });
        let articleValue = Number.NaN;
        let bookValue = Number.NaN;
        for (const price of [10, 20, 30]) {
            articleValue = article.update(price);
            bookValue = book.update(price);
        }
        expectCloseOrNaN(articleValue, -5 / 3, 1e-15);
        expectCloseOrNaN(bookValue, 1 / 3, 1e-15);
        expect(Number.isNaN(article.trigger())).toBeTrue();
        expectCloseOrNaN(article.update(60), -18 / 11, 1e-15);
        expectCloseOrNaN(article.trigger(), -5 / 3, 1e-15);

        for (const length of [1, 2, 3, 10]) {
            for (const centered of [false, true]) {
                const indicator = new CenterOfGravityOscillator({ length, centered });
                const baseline = centered ? 0 : -(length + 1) / 2;
                for (let step = 0; step < length + 2; step++) {
                    expectCloseOrNaN(indicator.update(8), step < length - 1 ? Number.NaN : baseline, 1e-14);
                    expectCloseOrNaN(indicator.trigger(), step < length ? Number.NaN : baseline, 1e-14);
                }
            }
        }

        const base = new CenterOfGravityOscillator({ length: 5, centered: true });
        const scaled = new CenterOfGravityOscillator({ length: 5, centered: true });
        for (let index = 0; index < 25; index++) {
            const price = testdata.testInput[index];
            expectCloseOrNaN(base.update(price), scaled.update(3.75 * price));
            expectCloseOrNaN(base.trigger(), scaled.trigger());
        }

        const rising = new CenterOfGravityOscillator({ length: 4, centered: true });
        const falling = new CenterOfGravityOscillator({ length: 4, centered: true });
        for (let step = 0; step < 4; step++) {
            rising.update(10 + step);
            falling.update(13 - step);
        }
        expect(rising.value()).toBeGreaterThan(0);
        expect(falling.value()).toBeLessThan(0);
        expect(rising.update(1)).toBeLessThan(rising.trigger());
    });

    it('handles zero-denominator hold, initial zero without offset, and subnormal denominators', () => {
        for (const centered of [false, true]) {
            const initial = new CenterOfGravityOscillator({ length: 2, centered });
            initial.update(1);
            expect(initial.update(-1)).toBe(0);
            expect(Number.isNaN(initial.trigger())).toBeTrue();

            const subsequent = new CenterOfGravityOscillator({ length: 2, centered });
            subsequent.update(1);
            const previous = subsequent.update(2);
            expect(subsequent.update(-2)).toBe(previous);
            expect(subsequent.trigger()).toBe(previous);
            const resumed = subsequent.update(4);
            expect(subsequent.trigger()).toBe(previous);
            expect(resumed).toBe(centered ? 1.5 : 0);
        }

        const subnormal = new CenterOfGravityOscillator({ length: 2 });
        subnormal.update(1e-309);
        expectCloseOrNaN(subnormal.update(3e-309), -1.25, 1e-14);
    });

    it('skips NaN without advancing warmup, rejects non-finite samples, and resets cleanly', () => {
        const skipped = new CenterOfGravityOscillator({ length: 3 });
        const uninterrupted = new CenterOfGravityOscillator({ length: 3 });
        const stamp = new Date(Date.UTC(2026, 0, 1));

        for (const price of [1, 2, 3, 4, 5]) {
            expect(Number.isNaN(skipped.update(Number.NaN))).toBeTrue();
            const gap = skipped.updateScalar(new Scalar({ time: stamp, value: Number.NaN }));
            expect(Number.isNaN((gap[0] as Scalar).value)).toBeTrue();
            expect(Number.isNaN((gap[1] as Scalar).value)).toBeTrue();

            const actual = skipped.update(price);
            const reference = uninterrupted.update(price);
            expectCloseOrNaN(actual, reference, 0);
            expectCloseOrNaN(skipped.trigger(), uninterrupted.trigger(), 0);
        }
        expect(skipped.isTriggerPrimed()).toBeTrue();

        const single = new CenterOfGravityOscillator({ length: 1 });
        single.update(4);
        for (const bad of [Number.POSITIVE_INFINITY, Number.NEGATIVE_INFINITY, true as unknown as number, 'bad' as unknown as number]) {
            expect(() => single.update(bad)).toThrow();
        }
        expect(single.update(5)).toBe(-1);
        expect(single.trigger()).toBe(-1);

        const resettable = new CenterOfGravityOscillator({ length: 2, centered: true });
        for (const price of [1, 2, 3]) {
            resettable.update(price);
        }
        resettable.reset();
        expect(resettable.isPrimed()).toBeFalse();
        expect(resettable.isTriggerPrimed()).toBeFalse();
        expect(Number.isNaN(resettable.value())).toBeTrue();
        expect(Number.isNaN(resettable.trigger())).toBeTrue();
        expect(Number.isNaN(resettable.update(1))).toBeTrue();
        expect(resettable.update(-1)).toBe(0);
    });

    it('validates parameters, formats mnemonics/metadata, routes entities, and integrates with factory', () => {
        for (const badLength of [0, -1, 1.5, Number.NaN, true as unknown as number, '10' as unknown as number, null as unknown as number]) {
            expect(() => new CenterOfGravityOscillator({ length: badLength })).toThrow();
        }
        for (const badCentered of [0, 1, null, 'book']) {
            expect(() => new CenterOfGravityOscillator({ centered: badCentered as unknown as boolean })).toThrow();
        }
        expect(() => new CenterOfGravityOscillator({ barComponent: 9999 as BarComponent })).toThrow();

        expect(CenterOfGravityOscillatorOutput.Value).toBe(0);
        expect(CenterOfGravityOscillatorOutput.Trigger).toBe(1);

        const defaultIndicator = CenterOfGravityOscillator.default();
        const meta = defaultIndicator.metadata();
        expect(meta.identifier).toBe(IndicatorIdentifier.CenterOfGravityOscillator);
        expect(meta.mnemonic).toBe('cog(10)');
        expect(meta.description).toBe('Center of Gravity oscillator cog(10)');
        expect(meta.outputs.length).toBe(2);
        expect(meta.outputs[0].kind).toBe(CenterOfGravityOscillatorOutput.Value);
        expect(meta.outputs[0].shape).toBe(Shape.Scalar);
        expect(meta.outputs[0].mnemonic).toBe('cog(10)');
        expect(meta.outputs[1].kind).toBe(CenterOfGravityOscillatorOutput.Trigger);
        expect(meta.outputs[1].shape).toBe(Shape.Scalar);
        expect(meta.outputs[1].mnemonic).toBe('cogTrig(10)');
        expect(meta.outputs[1].description).toBe('Center of Gravity trigger cogTrig(10)');

        const customParams: CenterOfGravityOscillatorParams = {
            length: 14,
            centered: true,
            barComponent: BarComponent.Median,
            quoteComponent: QuoteComponent.Bid,
            tradeComponent: TradeComponent.Volume,
        };
        expect(centerOfGravityOscillatorMnemonic(customParams)).toBe('cog(14, centered, hl/2, b, v)');
        expect(centerOfGravityOscillatorTriggerMnemonic(customParams)).toBe('cogTrig(14, centered, hl/2, b, v)');

        const fromFactory = createIndicator(IndicatorIdentifier.CenterOfGravityOscillator, {
            length: 8,
            centered: true,
            barComponent: BarComponent.Median,
        });
        expect(fromFactory.metadata().mnemonic).toBe('cog(8, centered, hl/2)');
        expect(() => createIndicator(IndicatorIdentifier.CenterOfGravityOscillator, { length: 1.5 })).toThrow();
        expect(() => createIndicator(IndicatorIdentifier.CenterOfGravityOscillator, { centered: 1 })).toThrow();

        const barMedian = new CenterOfGravityOscillator({ length: 4, centered: true, barComponent: BarComponent.Median });
        const quoteMid = new CenterOfGravityOscillator({ length: 4, centered: true });
        const tradePrice = new CenterOfGravityOscillator({ length: 4, centered: true });
        const scalarDirect = new CenterOfGravityOscillator({ length: 4, centered: true });

        for (let step = 0; step < 12; step++) {
            const stamp = new Date(Date.UTC(2026, 0, 1 + step));
            const median = (testdata.testHigh[step] + testdata.testLow[step]) / 2;
            const barOut = barMedian.updateBar(
                new Bar({
                    time: stamp,
                    open: testdata.testOpen[step],
                    high: testdata.testHigh[step],
                    low: testdata.testLow[step],
                    close: testdata.testInput[step],
                    volume: 1000,
                }),
            );
            const quoteOut = quoteMid.updateQuote(
                new Quote({
                    time: stamp,
                    bidPrice: testdata.testLow[step],
                    askPrice: testdata.testHigh[step],
                    bidSize: 1,
                    askSize: 1,
                }),
            );
            const tradeOut = tradePrice.updateTrade(
                new Trade({ time: stamp, price: median, volume: 10 }),
            );
            const scalarOut = scalarDirect.updateScalar(
                new Scalar({ time: stamp, value: median }),
            );

            for (let slot = 0; slot < 2; slot++) {
                const expected = (scalarOut[slot] as Scalar).value;
                expectCloseOrNaN((barOut[slot] as Scalar).value, expected, 1e-15);
                expectCloseOrNaN((quoteOut[slot] as Scalar).value, expected, 1e-15);
                expectCloseOrNaN((tradeOut[slot] as Scalar).value, expected, 1e-15);
            }
        }
    });
});
