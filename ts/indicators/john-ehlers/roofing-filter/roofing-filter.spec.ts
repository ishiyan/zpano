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
import { RoofingFilterOutput } from './output';
import { RoofingFilterParams, defaultParams, roofingFilterMnemonic } from './params';
import { RoofingFilter } from './roofing-filter';
import * as testdata from './testdata';

interface ReferenceCase {
    readonly params: RoofingFilterParams;
    readonly expected: number[];
    readonly warmup: number;
}

const referenceCases: ReferenceCase[] = [
    { params: { shortestCyclePeriod: 10, longestCyclePeriod: 48, highpassPoles: 2, zeroMean: false }, expected: testdata.expectedLP10_HP48_PO2_ZM0, warmup: 4 },
    { params: { shortestCyclePeriod: 40, longestCyclePeriod: 80, highpassPoles: 2, zeroMean: false }, expected: testdata.expectedLP40_HP80_PO2_ZM0, warmup: 4 },
    { params: { shortestCyclePeriod: 2, longestCyclePeriod: 3, highpassPoles: 2, zeroMean: false }, expected: testdata.expectedLP2_HP3_PO2_ZM0, warmup: 4 },
    { params: { shortestCyclePeriod: 5, longestCyclePeriod: 24, highpassPoles: 2, zeroMean: false }, expected: testdata.expectedLP5_HP24_PO2_ZM0, warmup: 4 },
    { params: { shortestCyclePeriod: 10.5, longestCyclePeriod: 48, highpassPoles: 2, zeroMean: false }, expected: testdata.expectedLP10_5_HP48_PO2_ZM0, warmup: 4 },
    { params: { shortestCyclePeriod: 10, longestCyclePeriod: 11, highpassPoles: 2, zeroMean: false }, expected: testdata.expectedLP10_HP11_PO2_ZM0, warmup: 4 },
    { params: { shortestCyclePeriod: 20, longestCyclePeriod: 80, highpassPoles: 2, zeroMean: false }, expected: testdata.expectedLP20_HP80_PO2_ZM0, warmup: 4 },
    { params: { shortestCyclePeriod: 40, longestCyclePeriod: 252, highpassPoles: 2, zeroMean: false }, expected: testdata.expectedLP40_HP252_PO2_ZM0, warmup: 4 },
    { params: { shortestCyclePeriod: 10, longestCyclePeriod: 48, highpassPoles: 1, zeroMean: false }, expected: testdata.expectedLP10_HP48_PO1_ZM0, warmup: 3 },
    { params: { shortestCyclePeriod: 2, longestCyclePeriod: 4, highpassPoles: 1, zeroMean: false }, expected: testdata.expectedLP2_HP4_PO1_ZM0, warmup: 3 },
    { params: { shortestCyclePeriod: 10, longestCyclePeriod: 48, highpassPoles: 1, zeroMean: true }, expected: testdata.expectedLP10_HP48_PO1_ZM1, warmup: 4 },
    { params: { shortestCyclePeriod: 2, longestCyclePeriod: 4, highpassPoles: 1, zeroMean: true }, expected: testdata.expectedLP2_HP4_PO1_ZM1, warmup: 4 },
];

describe('RoofingFilter', () => {
    it('matches all twelve 252-bar reference configurations', () => {
        expect(testdata.testInput.length).toBe(252);
        for (const item of referenceCases) {
            const indicator = new RoofingFilter(item.params);
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
            const first = new RoofingFilter(item.params);
            const second = new RoofingFilter(item.params);
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
            const third = new RoofingFilter(item.params);
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

    it('converges to zero on constant and linear-ramp inputs and handles 1-pole period 4', () => {
        const constantIndicator = RoofingFilter.default();
        const constantOutputs: number[] = [];
        for (let index = 0; index < 500; index++) {
            constantOutputs.push(constantIndicator.update(7.0));
        }
        const firstFinite = constantOutputs.findIndex((entry) => Number.isFinite(entry));
        expect(firstFinite).toBe(constantIndicator.warmupBars());
        expect(Math.abs(constantOutputs[constantOutputs.length - 1])).toBeLessThanOrEqual(1e-12);

        const rampIndicator = new RoofingFilter();
        let lastRamp = Number.NaN;
        for (let index = 0; index < 1000; index++) {
            lastRamp = rampIndicator.update(index);
        }
        expect(Math.abs(lastRamp)).toBeLessThanOrEqual(1e-12);

        const boundary = new RoofingFilter({
            shortestCyclePeriod: 2,
            longestCyclePeriod: 4,
            highpassPoles: 1,
            zeroMean: false,
        });
        for (let index = 0; index < 20; index++) {
            const output = boundary.update(testdata.testInput[index]);
            if (index >= boundary.warmupBars()) {
                expect(Number.isFinite(output)).toBeTrue();
            }
        }
    });

    it('rejects invalid parameters and infinite samples without mutating state', () => {
        const invalidParams: RoofingFilterParams[] = [
            { shortestCyclePeriod: 1 },
            { shortestCyclePeriod: 0 },
            { shortestCyclePeriod: -2 },
            { shortestCyclePeriod: Number.NaN },
            { shortestCyclePeriod: Number.POSITIVE_INFINITY },
            { shortestCyclePeriod: 10, longestCyclePeriod: 10 },
            { shortestCyclePeriod: 10, longestCyclePeriod: 5 },
            { shortestCyclePeriod: 10, longestCyclePeriod: Number.NaN },
            { shortestCyclePeriod: 10, longestCyclePeriod: Number.POSITIVE_INFINITY },
            { highpassPoles: 0 },
            { highpassPoles: 3 },
            { highpassPoles: 1.5 },
            { highpassPoles: true as unknown as number },
            { highpassPoles: 2, zeroMean: true },
            { shortestCyclePeriod: true as unknown as number },
            { zeroMean: 1 as unknown as boolean },
        ];
        for (const badParams of invalidParams) {
            expect(() => new RoofingFilter(badParams)).toThrow();
        }

        const indicator = new RoofingFilter();
        expect(() => indicator.update(Number.POSITIVE_INFINITY)).toThrow();
        expect(() => indicator.update(Number.NEGATIVE_INFINITY)).toThrow();
        expect(indicator.isPrimed()).toBeFalse();
    });

    it('produces expected metadata and mnemonics across component combinations', () => {
        const indicator = new RoofingFilter(defaultParams());
        const meta = indicator.metadata();
        expect(meta.identifier).toBe(IndicatorIdentifier.RoofingFilter);
        expect(meta.mnemonic).toBe('roof2hp(10, 48)');
        expect(meta.description).toBe('Roofing Filter roof2hp(10, 48)');
        expect(meta.outputs.length).toBe(1);
        expect(meta.outputs[0].kind).toBe(RoofingFilterOutput.RoofingFilterValue);
        expect(meta.outputs[0].shape).toBe(Shape.Scalar);
        expect(meta.outputs[0].mnemonic).toBe('roof2hp(10, 48)');
        expect(meta.outputs[0].description).toBe('Roofing Filter roof2hp(10, 48)');

        const mnemonicCases: Array<[RoofingFilterParams | undefined, string]> = [
            [undefined, 'roof2hp(10, 48)'],
            [{}, 'roof2hp(10, 48)'],
            [{ shortestCyclePeriod: 10.5, longestCyclePeriod: 48, highpassPoles: 2, zeroMean: false }, 'roof2hp(10.5, 48)'],
            [{ shortestCyclePeriod: 10, longestCyclePeriod: 48, highpassPoles: 1, zeroMean: false }, 'roof1hp(10, 48)'],
            [{ shortestCyclePeriod: 10, longestCyclePeriod: 48, highpassPoles: 1, zeroMean: true }, 'roof1hpzm(10, 48)'],
            [{ barComponent: BarComponent.Median }, 'roof2hp(10, 48, hl/2)'],
            [{ quoteComponent: QuoteComponent.Bid }, 'roof2hp(10, 48, b)'],
            [{ tradeComponent: TradeComponent.Volume }, 'roof2hp(10, 48, v)'],
            [{ barComponent: BarComponent.Open, quoteComponent: QuoteComponent.Ask }, 'roof2hp(10, 48, o, a)'],
            [{ barComponent: BarComponent.Typical, tradeComponent: TradeComponent.Volume }, 'roof2hp(10, 48, hlc/3, v)'],
            [{ quoteComponent: QuoteComponent.Weighted, tradeComponent: TradeComponent.Volume }, 'roof2hp(10, 48, (bbs+aas)/(bs+as), v)'],
        ];
        for (const [candidate, expectedMnemonic] of mnemonicCases) {
            expect(roofingFilterMnemonic(candidate)).toBe(expectedMnemonic);
            expect(new RoofingFilter(candidate).metadata().mnemonic).toBe(expectedMnemonic);
        }
    });

    it('updates entities via default components and instantiates through factory', () => {
        const barIndicator = new RoofingFilter();
        const scalarIndicator = new RoofingFilter();
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

        const quoteIndicator = new RoofingFilter();
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

        const tradeIndicator = new RoofingFilter();
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

        const createdDefault = createIndicator(IndicatorIdentifier.RoofingFilter);
        expect(createdDefault.metadata().mnemonic).toBe('roof2hp(10, 48)');

        const createdCustom = createIndicator(IndicatorIdentifier.RoofingFilter, {
            shortestCyclePeriod: 10.5,
            longestCyclePeriod: 48,
            highpassPoles: 1,
            zeroMean: true,
        });
        expect(createdCustom.metadata().mnemonic).toBe('roof1hpzm(10.5, 48)');
    });
});
