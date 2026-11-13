import { } from 'jasmine';

import { Bar } from '../../../entities/bar';
import { BarComponent, DefaultBarComponent } from '../../../entities/bar-component';
import { Quote } from '../../../entities/quote';
import { DefaultQuoteComponent, QuoteComponent } from '../../../entities/quote-component';
import { Scalar } from '../../../entities/scalar';
import { Trade } from '../../../entities/trade';
import { DefaultTradeComponent, TradeComponent } from '../../../entities/trade-component';
import { componentTripleMnemonic } from '../../core/component-triple-mnemonic';
import { IndicatorIdentifier } from '../../core/indicator-identifier';
import { Heatmap } from '../../core/outputs/heatmap';
import { Shape } from '../../core/outputs/shape/shape';
import { CoronaSignalToNoiseRatio, coronaSignalToNoiseRatioMnemonic } from './corona-signal-to-noise-ratio';
import { CoronaSignalToNoiseRatioOutput } from './output';
import * as testdata from './testdata';

const testTolerance = 1e-10;

function assertClose(actual: number, expected: number, label: string): void {
    if (Number.isNaN(expected)) {
        if (!Number.isNaN(actual)) {
            fail(`${label}: expected NaN, got ${actual}`);
        }
        return;
    }
    if (!Number.isFinite(expected)) {
        if (actual !== expected) {
            fail(`${label}: expected ${expected}, got ${actual}`);
        }
        return;
    }
    if (!Number.isFinite(actual)) {
        fail(`${label}: expected ${expected}, got ${actual}`);
        return;
    }
    const bound = Math.max(testTolerance, Math.abs(expected) * testTolerance);
    if (Math.abs(actual - expected) > bound) {
        fail(`${label}: expected ${expected}, got ${actual} (diff ${Math.abs(actual - expected)} > ${bound})`);
    }
}

describe('CoronaSignalToNoiseRatio', () => {
    const baseTime = new Date(Date.UTC(2020, 0, 1));

    it('should have correct output enum values', () => {
        expect(CoronaSignalToNoiseRatioOutput.Value).toBe(0);
        expect(CoronaSignalToNoiseRatioOutput.SignalToNoiseRatio).toBe(1);
    });

    it('should match all 16 reference cases across 252 samples', () => {
        const cases: Array<{ name: string; hpCutoff: number; minPeriod: number; maxPeriod: number; rasterLength: number }> = [
            { name: 'H30_PMIN6_PMAX30_R50', hpCutoff: 30, minPeriod: 6, maxPeriod: 30, rasterLength: 50 },
            { name: 'H30_PMIN6_PMAX30_R25', hpCutoff: 30, minPeriod: 6, maxPeriod: 30, rasterLength: 25 },
            { name: 'H30_PMIN6_PMAX20_R50', hpCutoff: 30, minPeriod: 6, maxPeriod: 20, rasterLength: 50 },
            { name: 'H30_PMIN6_PMAX20_R25', hpCutoff: 30, minPeriod: 6, maxPeriod: 20, rasterLength: 25 },
            { name: 'H30_PMIN10_PMAX30_R50', hpCutoff: 30, minPeriod: 10, maxPeriod: 30, rasterLength: 50 },
            { name: 'H30_PMIN10_PMAX30_R25', hpCutoff: 30, minPeriod: 10, maxPeriod: 30, rasterLength: 25 },
            { name: 'H30_PMIN10_PMAX48_R50', hpCutoff: 30, minPeriod: 10, maxPeriod: 48, rasterLength: 50 },
            { name: 'H30_PMIN10_PMAX48_R25', hpCutoff: 30, minPeriod: 10, maxPeriod: 48, rasterLength: 25 },
            { name: 'H48_PMIN6_PMAX30_R50', hpCutoff: 48, minPeriod: 6, maxPeriod: 30, rasterLength: 50 },
            { name: 'H48_PMIN6_PMAX30_R25', hpCutoff: 48, minPeriod: 6, maxPeriod: 30, rasterLength: 25 },
            { name: 'H48_PMIN6_PMAX20_R50', hpCutoff: 48, minPeriod: 6, maxPeriod: 20, rasterLength: 50 },
            { name: 'H48_PMIN6_PMAX20_R25', hpCutoff: 48, minPeriod: 6, maxPeriod: 20, rasterLength: 25 },
            { name: 'H48_PMIN10_PMAX30_R50', hpCutoff: 48, minPeriod: 10, maxPeriod: 30, rasterLength: 50 },
            { name: 'H48_PMIN10_PMAX30_R25', hpCutoff: 48, minPeriod: 10, maxPeriod: 30, rasterLength: 25 },
            { name: 'H48_PMIN10_PMAX48_R50', hpCutoff: 48, minPeriod: 10, maxPeriod: 48, rasterLength: 50 },
            { name: 'H48_PMIN10_PMAX48_R25', hpCutoff: 48, minPeriod: 10, maxPeriod: 48, rasterLength: 25 },
        ];

        const lookup = testdata as Record<string, unknown>;
        for (const tc of cases) {
            const indicator = CoronaSignalToNoiseRatio.fromParams({
                rasterLength: tc.rasterLength,
                highPassFilterCutoff: tc.hpCutoff,
                minimalPeriod: tc.minPeriod,
                maximalPeriod: tc.maxPeriod,
            });
            const step = 10.0 / tc.rasterLength;
            const expectedFirst = 1.0 + step;
            const expectedResolution = 1.0 / step;

            const expectedHeatmap = lookup[`expected${tc.name}_HEATMAP`] as number[][];
            const expectedMin = lookup[`expected${tc.name}_HEATMAP_MIN`] as number[];
            const expectedMax = lookup[`expected${tc.name}_HEATMAP_MAX`] as number[];
            const expectedSnr = lookup[`expected${tc.name}_SIGNAL_TO_NOISE_RATIO`] as number[];
            const expectedPhys = lookup[`expected${tc.name}_PHYSICAL_DECIBELS`] as number[];
            const expectedCal = lookup[`expected${tc.name}_CALIBRATED_DECIBELS`] as number[];
            const expectedNorm = lookup[`expected${tc.name}_NORMALIZED_POSITION`] as number[];
            const expectedSignal = lookup[`expected${tc.name}_SIGNAL`] as number[];
            const expectedNoise = lookup[`expected${tc.name}_NOISE`] as number[];
            const expectedMedRange = lookup[`expected${tc.name}_MEDIAN_BAR_RANGE`] as number[];
            const expectedWidth = lookup[`expected${tc.name}_WIDTH`] as number[];
            const expectedDc = lookup[`expected${tc.name}_DOMINANT_CYCLE`] as number[];
            const expectedDcm = lookup[`expected${tc.name}_DOMINANT_CYCLE_MEDIAN`] as number[];
            const expectedHasEst = lookup[`expected${tc.name}_HAS_CYCLE_ESTIMATE`] as number[];
            const expectedFinite = lookup[`expected${tc.name}_RATIO_IS_FINITE`] as number[];

            for (let i = 0; i < testdata.testInput.length; i++) {
                const barTime = new Date(baseTime.getTime() + i * 86_400_000);
                const snap = indicator.update(testdata.testInput[i], testdata.testLow[i], testdata.testHigh[i], barTime);

                assertClose(snap.heatmap.parameterFirst, expectedFirst, `${tc.name}[${i}].first`);
                assertClose(snap.heatmap.parameterLast, 11.0, `${tc.name}[${i}].last`);
                assertClose(snap.heatmap.parameterResolution, expectedResolution, `${tc.name}[${i}].resolution`);

                const rowIsEmpty = Number.isNaN(expectedSnr[i]);
                expect(indicator.isPrimed()).toBe(!rowIsEmpty);
                if (rowIsEmpty) {
                    expect(snap.heatmap.isEmpty()).toBe(true);
                } else {
                    expect(snap.heatmap.values.length).toBe(tc.rasterLength);
                    for (let binIndex = 0; binIndex < tc.rasterLength; binIndex++) {
                        assertClose(snap.heatmap.values[binIndex], expectedHeatmap[i][binIndex], `${tc.name}[${i}].hm[${binIndex}]`);
                    }
                }

                assertClose(snap.heatmap.valueMin, expectedMin[i], `${tc.name}[${i}].min`);
                assertClose(snap.heatmap.valueMax, expectedMax[i], `${tc.name}[${i}].max`);
                assertClose(snap.signalToNoiseRatio, expectedSnr[i], `${tc.name}[${i}].snr`);
                assertClose(snap.physicalDecibels, expectedPhys[i], `${tc.name}[${i}].phys`);
                assertClose(snap.calibratedDecibels, expectedCal[i], `${tc.name}[${i}].cal`);
                assertClose(snap.normalizedPosition, expectedNorm[i], `${tc.name}[${i}].norm`);
                assertClose(snap.signal, expectedSignal[i], `${tc.name}[${i}].signal`);
                assertClose(snap.noise, expectedNoise[i], `${tc.name}[${i}].noise`);
                assertClose(snap.medianBarRange, expectedMedRange[i], `${tc.name}[${i}].medianRange`);
                assertClose(snap.width, expectedWidth[i], `${tc.name}[${i}].width`);
                assertClose(snap.dominantCycle, expectedDc[i], `${tc.name}[${i}].dc`);
                assertClose(snap.dominantCycleMedian, expectedDcm[i], `${tc.name}[${i}].dcm`);
                assertClose(snap.hasCycleEstimate, expectedHasEst[i], `${tc.name}[${i}].hasEst`);
                assertClose(snap.ratioIsFinite, expectedFinite[i], `${tc.name}[${i}].finite`);
            }

            indicator.reset();
            expect(indicator.isPrimed()).toBe(false);
        }
    });

    it('should expose expected metadata, mnemonics, negative minParameterValue, validation, and entity wrappers', () => {
        const indicator = CoronaSignalToNoiseRatio.default();
        const meta = indicator.metadata();
        expect(meta.identifier).toBe(IndicatorIdentifier.CoronaSignalToNoiseRatio);
        expect(meta.mnemonic).toBe('csnr(50, 20, 1, 11, 30)');
        expect(meta.outputs.length).toBe(2);
        expect(meta.outputs[0].shape).toBe(Shape.Heatmap);
        expect(meta.outputs[1].shape).toBe(Shape.Scalar);
        expect(meta.outputs[1].mnemonic).toBe('csnr-snr(30)');

        const negBounds = CoronaSignalToNoiseRatio.fromParams({
            minParameterValue: -5,
            maxParameterValue: 5,
        });
        expect(negBounds.metadata().mnemonic).toBe('csnr(50, 20, -5, 5, 30)');

        const flagged = CoronaSignalToNoiseRatio.fromParams({
            minimalPeriod: 10,
            maximalPeriod: 48,
            barComponent: BarComponent.Median,
        });
        expect(flagged.metadata().mnemonic).toBe('csnr(50, 20, 1, 11, 30, pmin=10, pmax=48, hl/2)');
        expect(flagged.metadata().outputs[1].mnemonic).toBe('csnr-snr(30, pmin=10, pmax=48, hl/2)');

        const combinations: Array<[BarComponent | undefined, QuoteComponent | undefined, TradeComponent | undefined]> = [
            [undefined, undefined, undefined],
            [BarComponent.Open, undefined, undefined],
            [undefined, QuoteComponent.Bid, undefined],
            [undefined, undefined, TradeComponent.Volume],
            [BarComponent.Median, QuoteComponent.Ask, undefined],
            [BarComponent.Median, undefined, TradeComponent.Volume],
            [undefined, QuoteComponent.Ask, TradeComponent.Volume],
            [BarComponent.Open, QuoteComponent.Bid, TradeComponent.Volume],
        ];
        for (const [bc, qc, tc] of combinations) {
            const p = { barComponent: bc, quoteComponent: qc, tradeComponent: tc };
            const instance = CoronaSignalToNoiseRatio.fromParams(p);
            const suffix = componentTripleMnemonic(
                bc ?? DefaultBarComponent,
                qc ?? DefaultQuoteComponent,
                tc ?? DefaultTradeComponent,
            );
            const expected = `csnr(50, 20, 1, 11, 30${suffix})`;
            expect(instance.metadata().mnemonic).toBe(expected);
            expect(coronaSignalToNoiseRatioMnemonic(p)).toBe(expected);
        }

        expect(() => CoronaSignalToNoiseRatio.fromParams({ rasterLength: 1 })).toThrow();
        expect(() => CoronaSignalToNoiseRatio.fromParams({ maxRasterValue: 6 })).toThrow();
        expect(() => CoronaSignalToNoiseRatio.fromParams({ minParameterValue: 11, maxParameterValue: 11 })).toThrow();
        expect(() => CoronaSignalToNoiseRatio.fromParams({ highPassFilterCutoff: 4 })).toThrow();
        expect(() => CoronaSignalToNoiseRatio.fromParams({ minimalPeriod: 5 })).toThrow();
        expect(() => CoronaSignalToNoiseRatio.fromParams({ minimalPeriod: 10, maximalPeriod: 10 })).toThrow();

        const rangeCheck = CoronaSignalToNoiseRatio.default();
        expect(() => rangeCheck.update(100.0, 101.0, 99.0, baseTime)).toThrow();
        expect(() => rangeCheck.updateRange(100.0, -1.0, baseTime)).toThrow();
        expect(() => rangeCheck.update(100.0, 100.0, Number.POSITIVE_INFINITY, baseTime)).toThrow();
        expect(rangeCheck.isPrimed()).toBe(false);
        expect(rangeCheck.update(100.0, 99.0, Number.NaN, baseTime).heatmap.isEmpty()).toBe(true);

        const barInd = CoronaSignalToNoiseRatio.default();
        const scalarInd = CoronaSignalToNoiseRatio.default();
        const quoteInd = CoronaSignalToNoiseRatio.default();
        const tradeInd = CoronaSignalToNoiseRatio.default();
        for (let i = 0; i < 15; i++) {
            const barTime = new Date(baseTime.getTime() + i * 86_400_000);
            const price = testdata.testInput[i];
            const outBar = barInd.updateBar(new Bar({ time: barTime, open: testdata.testOpen[i], high: testdata.testHigh[i], low: testdata.testLow[i], close: price, volume: 1000 })) as unknown[];
            const s = new Scalar();
            s.time = barTime;
            s.value = price;
            const outScalar = scalarInd.updateScalar(s) as unknown[];
            const outQuote = quoteInd.updateQuote(new Quote({ time: barTime, bidPrice: price - 0.5, askPrice: price + 0.5, bidSize: 10, askSize: 10 })) as unknown[];
            const outTrade = tradeInd.updateTrade(new Trade({ time: barTime, price, volume: 10 })) as unknown[];
            expect(outBar.length).toBe(2);
            expect(outBar[0] instanceof Heatmap).toBe(true);
            expect(outScalar.length).toBe(2);
            expect(outQuote.length).toBe(2);
            expect(outTrade.length).toBe(2);
        }
        const finalScalar = new Scalar();
        finalScalar.time = baseTime;
        finalScalar.value = 100.0;
        const finalOut = scalarInd.updateScalar(finalScalar) as [Heatmap, Scalar];
        expect(finalOut[1].value).toBe(11.0);
    });
});
