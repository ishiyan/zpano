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
import { CoronaSpectrum, coronaSpectrumMnemonic } from './corona-spectrum';
import { CoronaSpectrumOutput } from './output';
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

describe('CoronaSpectrum', () => {
    const baseTime = new Date(Date.UTC(2020, 0, 1));

    it('should have correct output enum values', () => {
        expect(CoronaSpectrumOutput.Value).toBe(0);
        expect(CoronaSpectrumOutput.DominantCycle).toBe(1);
        expect(CoronaSpectrumOutput.DominantCycleMedian).toBe(2);
    });

    it('should match all 16 reference cases across 252 samples', () => {
        const cases: Array<{ name: string; hpCutoff: number; minPeriod: number; maxPeriod: number; threshold: number }> = [
            { name: 'H30_PMIN6_PMAX30_DB6', hpCutoff: 30, minPeriod: 6, maxPeriod: 30, threshold: 6 },
            { name: 'H30_PMIN6_PMAX30_DB3', hpCutoff: 30, minPeriod: 6, maxPeriod: 30, threshold: 3 },
            { name: 'H30_PMIN6_PMAX20_DB6', hpCutoff: 30, minPeriod: 6, maxPeriod: 20, threshold: 6 },
            { name: 'H30_PMIN6_PMAX20_DB3', hpCutoff: 30, minPeriod: 6, maxPeriod: 20, threshold: 3 },
            { name: 'H30_PMIN10_PMAX30_DB6', hpCutoff: 30, minPeriod: 10, maxPeriod: 30, threshold: 6 },
            { name: 'H30_PMIN10_PMAX30_DB3', hpCutoff: 30, minPeriod: 10, maxPeriod: 30, threshold: 3 },
            { name: 'H30_PMIN10_PMAX48_DB6', hpCutoff: 30, minPeriod: 10, maxPeriod: 48, threshold: 6 },
            { name: 'H30_PMIN10_PMAX48_DB3', hpCutoff: 30, minPeriod: 10, maxPeriod: 48, threshold: 3 },
            { name: 'H48_PMIN6_PMAX30_DB6', hpCutoff: 48, minPeriod: 6, maxPeriod: 30, threshold: 6 },
            { name: 'H48_PMIN6_PMAX30_DB3', hpCutoff: 48, minPeriod: 6, maxPeriod: 30, threshold: 3 },
            { name: 'H48_PMIN6_PMAX20_DB6', hpCutoff: 48, minPeriod: 6, maxPeriod: 20, threshold: 6 },
            { name: 'H48_PMIN6_PMAX20_DB3', hpCutoff: 48, minPeriod: 6, maxPeriod: 20, threshold: 3 },
            { name: 'H48_PMIN10_PMAX30_DB6', hpCutoff: 48, minPeriod: 10, maxPeriod: 30, threshold: 6 },
            { name: 'H48_PMIN10_PMAX30_DB3', hpCutoff: 48, minPeriod: 10, maxPeriod: 30, threshold: 3 },
            { name: 'H48_PMIN10_PMAX48_DB6', hpCutoff: 48, minPeriod: 10, maxPeriod: 48, threshold: 6 },
            { name: 'H48_PMIN10_PMAX48_DB3', hpCutoff: 48, minPeriod: 10, maxPeriod: 48, threshold: 3 },
        ];

        const lookup = testdata as Record<string, unknown>;
        for (const tc of cases) {
            const indicator = CoronaSpectrum.fromParams({
                minRasterValue: tc.threshold,
                maxRasterValue: 20,
                minParameterValue: tc.minPeriod,
                maxParameterValue: tc.maxPeriod,
                highPassFilterCutoff: tc.hpCutoff,
            });
            const bankLength = 2 * tc.maxPeriod - 2 * tc.minPeriod + 1;
            const expectedHeatmap = lookup[`expected${tc.name}_HEATMAP`] as number[][];
            const expectedMin = lookup[`expected${tc.name}_HEATMAP_MIN`] as number[];
            const expectedMax = lookup[`expected${tc.name}_HEATMAP_MAX`] as number[];
            const expectedDc = lookup[`expected${tc.name}_DOMINANT_CYCLE`] as number[];
            const expectedDcm = lookup[`expected${tc.name}_DOMINANT_CYCLE_MEDIAN`] as number[];
            const expectedMaxPower = lookup[`expected${tc.name}_MAXIMUM_POWER`] as number[];
            const expectedNormPower = lookup[`expected${tc.name}_NORMALIZATION_POWER`] as number[];
            const expectedSelectedBins = lookup[`expected${tc.name}_SELECTED_BINS`] as number[];
            const expectedHasEst = lookup[`expected${tc.name}_HAS_CYCLE_ESTIMATE`] as number[];
            const expectedPower = lookup[`expected${tc.name}_POWER`] as number[][];
            const expectedRelativePower = lookup[`expected${tc.name}_RELATIVE_POWER`] as number[][];

            for (let i = 0; i < testdata.testInput.length; i++) {
                const barTime = new Date(baseTime.getTime() + i * 86_400_000);
                const snap = indicator.update(testdata.testInput[i], barTime);

                assertClose(snap.heatmap.parameterFirst, tc.minPeriod, `${tc.name}[${i}].first`);
                assertClose(snap.heatmap.parameterLast, tc.maxPeriod, `${tc.name}[${i}].last`);
                assertClose(snap.heatmap.parameterResolution, 2.0, `${tc.name}[${i}].resolution`);

                const rowIsEmpty = Number.isNaN(expectedDc[i]);
                expect(indicator.isPrimed()).toBe(!rowIsEmpty);
                expect(snap.power.length).toBe(bankLength);
                expect(snap.relativePower.length).toBe(bankLength);

                if (rowIsEmpty) {
                    expect(snap.heatmap.isEmpty()).toBe(true);
                    expect(snap.power.every((v) => Number.isNaN(v))).toBe(true);
                    expect(snap.relativePower.every((v) => Number.isNaN(v))).toBe(true);
                } else {
                    expect(snap.heatmap.values.length).toBe(bankLength);
                    for (let binIndex = 0; binIndex < bankLength; binIndex++) {
                        assertClose(snap.heatmap.values[binIndex], expectedHeatmap[i][binIndex], `${tc.name}[${i}].hm[${binIndex}]`);
                        assertClose(snap.power[binIndex], expectedPower[i][binIndex], `${tc.name}[${i}].power[${binIndex}]`);
                        assertClose(snap.relativePower[binIndex], expectedRelativePower[i][binIndex], `${tc.name}[${i}].rel[${binIndex}]`);
                    }
                }

                assertClose(snap.heatmap.valueMin, expectedMin[i], `${tc.name}[${i}].min`);
                assertClose(snap.heatmap.valueMax, expectedMax[i], `${tc.name}[${i}].max`);
                assertClose(snap.dominantCycle, expectedDc[i], `${tc.name}[${i}].dc`);
                assertClose(snap.dominantCycleMedian, expectedDcm[i], `${tc.name}[${i}].dcm`);
                assertClose(snap.maximumPower, expectedMaxPower[i], `${tc.name}[${i}].maxPower`);
                assertClose(snap.normalizationPower, expectedNormPower[i], `${tc.name}[${i}].normPower`);
                assertClose(snap.selectedBins, expectedSelectedBins[i], `${tc.name}[${i}].selectedBins`);
                assertClose(snap.hasCycleEstimate, expectedHasEst[i], `${tc.name}[${i}].hasEst`);
            }

            indicator.reset();
            expect(indicator.isPrimed()).toBe(false);
        }
    });

    it('should expose expected metadata, mnemonics, validation, and entity wrappers', () => {
        const indicator = CoronaSpectrum.default();
        const meta = indicator.metadata();
        expect(meta.identifier).toBe(IndicatorIdentifier.CoronaSpectrum);
        expect(meta.mnemonic).toBe('cspect(6, 20, 6, 30, 30)');
        expect(meta.outputs.length).toBe(3);
        expect(meta.outputs[0].shape).toBe(Shape.Heatmap);
        expect(meta.outputs[1].shape).toBe(Shape.Scalar);
        expect(meta.outputs[1].mnemonic).toBe('cspect-dc(30)');
        expect(meta.outputs[2].shape).toBe(Shape.Scalar);
        expect(meta.outputs[2].mnemonic).toBe('cspect-dcm(30)');

        const rounded = CoronaSpectrum.fromParams({
            minRasterValue: 4,
            maxRasterValue: 25,
            minParameterValue: 8.7,
            maxParameterValue: 40.4,
            highPassFilterCutoff: 20,
        });
        expect(rounded.metadata().mnemonic).toBe('cspect(4, 25, 9, 40, 20)');

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
            const instance = CoronaSpectrum.fromParams(p);
            const suffix = componentTripleMnemonic(
                bc ?? DefaultBarComponent,
                qc ?? DefaultQuoteComponent,
                tc ?? DefaultTradeComponent,
            );
            const expected = `cspect(6, 20, 6, 30, 30${suffix})`;
            expect(instance.metadata().mnemonic).toBe(expected);
            expect(coronaSpectrumMnemonic(p)).toBe(expected);
        }

        expect(() => CoronaSpectrum.fromParams({ minRasterValue: -1 })).toThrow();
        expect(() => CoronaSpectrum.fromParams({ minRasterValue: 20, maxRasterValue: 20 })).toThrow();
        expect(() => CoronaSpectrum.fromParams({ minParameterValue: 5 })).toThrow();
        expect(() => CoronaSpectrum.fromParams({ minParameterValue: 20, maxParameterValue: 20 })).toThrow();
        expect(() => CoronaSpectrum.fromParams({ highPassFilterCutoff: 4 })).toThrow();

        const barInd = CoronaSpectrum.default();
        const scalarInd = CoronaSpectrum.default();
        const quoteInd = CoronaSpectrum.default();
        const tradeInd = CoronaSpectrum.default();
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
            expect(outBar.length).toBe(3);
            expect(outBar[0] instanceof Heatmap).toBe(true);
            expect(outScalar.length).toBe(3);
            expect(outQuote.length).toBe(3);
            expect(outTrade.length).toBe(3);
        }
    });
});
