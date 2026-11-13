import {} from 'jasmine';

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
import { SuperSmootherOutput } from './output';
import { SuperSmootherParams, defaultParams, superSmootherMnemonic } from './params';
import { SuperSmoother } from './super-smoother';
import * as testdata from './testdata';

interface ReferenceCase {
    readonly params: SuperSmootherParams;
    readonly expected: number[];
    readonly warmup: number;
}

const referenceCases: ReferenceCase[] = [
    { params: { shortestCyclePeriod: 2 }, expected: testdata.expectedP2, warmup: 2 },
    { params: { shortestCyclePeriod: 3 }, expected: testdata.expectedP3, warmup: 2 },
    { params: { shortestCyclePeriod: 5 }, expected: testdata.expectedP5, warmup: 2 },
    { params: { shortestCyclePeriod: 8 }, expected: testdata.expectedP8, warmup: 2 },
    { params: { shortestCyclePeriod: 10 }, expected: testdata.expectedP10, warmup: 2 },
    { params: { shortestCyclePeriod: 10.5 }, expected: testdata.expectedP10_5, warmup: 2 },
    { params: { shortestCyclePeriod: 20 }, expected: testdata.expectedP20, warmup: 2 },
    { params: { shortestCyclePeriod: 40 }, expected: testdata.expectedP40, warmup: 2 },
    { params: { shortestCyclePeriod: 80 }, expected: testdata.expectedP80, warmup: 2 },
    { params: { shortestCyclePeriod: 252 }, expected: testdata.expectedP252, warmup: 2 },
];

describe('SuperSmoother', () => {
    it('matches all ten 252-bar reference configurations', () => {
        expect(testdata.testInput.length).toBe(252);
        for (const item of referenceCases) {
            const indicator = new SuperSmoother(item.params);
            expect(indicator.warmupBars()).toBe(item.warmup);
            expect(indicator.isPrimed()).toBeFalse();
            expect(Number.isNaN(indicator.value())).toBeTrue();
            expect(item.expected.length).toBe(252);

            for (let index = 0; index < testdata.testInput.length; index++) {
                const actual = indicator.update(testdata.testInput[index]);
                const target = item.expected[index];
                if (index < item.warmup) {
                    expect(indicator.isPrimed()).toBeFalse();
                    expect(Number.isNaN(actual)).toBeTrue();
                    expect(Number.isNaN(indicator.value())).toBeTrue();
                } else {
                    expect(indicator.isPrimed()).toBeTrue();
                    expect(Number.isNaN(target)).toBeFalse();
                    const bound = Math.max(1e-11, Math.abs(target) * 1e-11);
                    expect(Math.abs(actual - target)).toBeLessThanOrEqual(bound);
                    expect(Math.abs(indicator.value() - actual)).toBeLessThanOrEqual(1e-15);
                }
            }
        }
    });

    it('skips NaN without mutating state and reset reproduces initial stream', () => {
        for (const item of referenceCases) {
            const first = new SuperSmoother(item.params);
            const second = new SuperSmoother(item.params);
            for (const sample of testdata.testInput) {
                expect(Number.isNaN(first.update(Number.NaN))).toBeTrue();
                const left = first.update(sample);
                const right = second.update(sample);
                expect(first.isPrimed()).toBe(second.isPrimed());
                if (Number.isNaN(right)) {
                    expect(Number.isNaN(left)).toBeTrue();
                } else {
                    expect(Math.abs(left - right)).toBeLessThanOrEqual(1e-15);
                }
            }

            first.reset();
            expect(first.isPrimed()).toBeFalse();
            expect(Number.isNaN(first.value())).toBeTrue();
            const third = new SuperSmoother(item.params);
            for (const sample of testdata.testInput) {
                const left = first.update(sample);
                const right = third.update(sample);
                if (Number.isNaN(right)) {
                    expect(Number.isNaN(left)).toBeTrue();
                } else {
                    expect(Math.abs(left - right)).toBeLessThanOrEqual(1e-15);
                }
            }
        }
    });

    it('preserves constant input and suppresses alternating Nyquist input', () => {
        const constantIndicator = SuperSmoother.default();
        const constantOutputs: number[] = [];
        for (let index = 0; index < 500; index++) {
            constantOutputs.push(constantIndicator.update(7.0));
        }
        const firstFinite = constantOutputs.findIndex((entry) => Number.isFinite(entry));
        expect(firstFinite).toBe(constantIndicator.warmupBars());
        expect(Math.abs(constantOutputs[constantOutputs.length - 1] - 7.0)).toBeLessThanOrEqual(1e-12);

        const nyquistIndicator = SuperSmoother.fromParams();
        let lastNyquist = Number.NaN;
        for (let index = 0; index < 500; index++) {
            lastNyquist = nyquistIndicator.update(index % 2 === 0 ? 1.0 : -1.0);
        }
        expect(Math.abs(lastNyquist)).toBeLessThanOrEqual(1e-12);
    });

    it('rejects invalid parameters and infinite samples without mutating state', () => {
        const invalidParams: SuperSmootherParams[] = [
            { shortestCyclePeriod: 1 },
            { shortestCyclePeriod: 1.999 },
            { shortestCyclePeriod: 0 },
            { shortestCyclePeriod: -2 },
            { shortestCyclePeriod: Number.NaN },
            { shortestCyclePeriod: Number.POSITIVE_INFINITY },
            { shortestCyclePeriod: Number.NEGATIVE_INFINITY },
            { shortestCyclePeriod: true as unknown as number },
        ];
        for (const badParams of invalidParams) {
            expect(() => new SuperSmoother(badParams)).toThrow();
        }

        const indicator = new SuperSmoother();
        expect(() => indicator.update(Number.POSITIVE_INFINITY)).toThrow();
        expect(() => indicator.update(Number.NEGATIVE_INFINITY)).toThrow();
        expect(() => indicator.update(true as unknown as number)).toThrow();
        expect(indicator.isPrimed()).toBeFalse();
    });

    it('produces expected metadata and mnemonics across component combinations', () => {
        const indicator = new SuperSmoother(defaultParams());
        const meta = indicator.metadata();
        expect(meta.identifier).toBe(IndicatorIdentifier.SuperSmoother);
        expect(meta.mnemonic).toBe('ss(10)');
        expect(meta.description).toBe('Super Smoother ss(10)');
        expect(meta.outputs.length).toBe(1);
        expect(meta.outputs[0].kind).toBe(SuperSmootherOutput.SuperSmootherValue);
        expect(meta.outputs[0].shape).toBe(Shape.Scalar);
        expect(meta.outputs[0].mnemonic).toBe('ss(10)');
        expect(meta.outputs[0].description).toBe('Super Smoother ss(10)');

        const mnemonicCases: Array<[SuperSmootherParams | undefined, string]> = [
            [undefined, 'ss(10)'],
            [{}, 'ss(10)'],
            [{ shortestCyclePeriod: 10.5 }, 'ss(10.5)'],
            [{ barComponent: BarComponent.Median }, 'ss(10, hl/2)'],
            [{ quoteComponent: QuoteComponent.Bid }, 'ss(10, b)'],
            [{ tradeComponent: TradeComponent.Volume }, 'ss(10, v)'],
            [{ barComponent: BarComponent.Open, quoteComponent: QuoteComponent.Ask }, 'ss(10, o, a)'],
            [{ barComponent: BarComponent.Typical, tradeComponent: TradeComponent.Volume }, 'ss(10, hlc/3, v)'],
            [{ quoteComponent: QuoteComponent.Weighted, tradeComponent: TradeComponent.Volume }, 'ss(10, (bbs+aas)/(bs+as), v)'],
        ];
        for (const [candidate, expectedMnemonic] of mnemonicCases) {
            expect(superSmootherMnemonic(candidate)).toBe(expectedMnemonic);
            expect(new SuperSmoother(candidate).metadata().mnemonic).toBe(expectedMnemonic);
        }
    });

    it('updates entities via default components and instantiates through factory', () => {
        const barIndicator = new SuperSmoother();
        const scalarIndicator = new SuperSmoother();
        const timestamp = new Date(Date.UTC(2024, 0, 2));

        for (let index = 0; index < 10; index++) {
            const bar = new Bar();
            bar.time = timestamp;
            bar.open = testdata.testOpen[index];
            bar.high = testdata.testHigh[index];
            bar.low = testdata.testLow[index];
            bar.close = testdata.testInput[index];
            bar.volume = 1000;

            const scalar = new Scalar();
            scalar.time = timestamp;
            scalar.value = testdata.testInput[index];

            const barOut = barIndicator.updateBar(bar);
            const scalarOut = scalarIndicator.updateScalar(scalar);
            expect(barOut.length).toBe(1);
            const barScalar = barOut[0] as Scalar;
            const refScalar = scalarOut[0] as Scalar;
            expect(barScalar.time).toBe(timestamp);
            if (Number.isNaN(refScalar.value)) {
                expect(Number.isNaN(barScalar.value)).toBeTrue();
            } else {
                expect(Math.abs(barScalar.value - refScalar.value)).toBeLessThanOrEqual(1e-15);
            }
        }

        const quoteIndicator = new SuperSmoother();
        let quoteResult: Scalar | undefined;
        for (let index = 0; index < 5; index++) {
            const quote = new Quote();
            quote.time = timestamp;
            quote.bidPrice = 99;
            quote.askPrice = 101;
            quote.bidSize = 1;
            quote.askSize = 1;
            quoteResult = quoteIndicator.updateQuote(quote)[0] as Scalar;
        }
        expect(quoteResult).toBeDefined();
        expect(Number.isNaN(quoteResult!.value)).toBeFalse();

        const tradeIndicator = new SuperSmoother();
        let tradeResult: Scalar | undefined;
        for (let index = 0; index < 5; index++) {
            const trade = new Trade();
            trade.time = timestamp;
            trade.price = 100;
            trade.volume = 50;
            tradeResult = tradeIndicator.updateTrade(trade)[0] as Scalar;
        }
        expect(tradeResult).toBeDefined();
        expect(Number.isNaN(tradeResult!.value)).toBeFalse();

        const createdDefault = createIndicator(IndicatorIdentifier.SuperSmoother);
        expect(createdDefault.metadata().mnemonic).toBe('ss(10)');

        const createdCustom = createIndicator(IndicatorIdentifier.SuperSmoother, {
            shortestCyclePeriod: 10.5,
        });
        expect(createdCustom.metadata().mnemonic).toBe('ss(10.5)');
    });
});
