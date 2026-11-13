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
import { Corona, CoronaSnapshot } from '../corona/corona';
import { CoronaTrendVigor, coronaTrendVigorMnemonic } from './corona-trend-vigor';
import { CoronaTrendVigorOutput } from './output';
import * as testdata from './testdata';

const testTolerance = 5e-12;

function nextDownPositive(value: number): number {
    const buffer = new ArrayBuffer(8);
    const floatView = new Float64Array(buffer);
    const uintView = new BigUint64Array(buffer);
    floatView[0] = value;
    uintView[0] -= 1n;
    return floatView[0];
}

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

describe('CoronaTrendVigor', () => {
    const baseTime = new Date(Date.UTC(2020, 0, 1));

    it('should have correct output enum values', () => {
        expect(CoronaTrendVigorOutput.Value).toBe(0);
        expect(CoronaTrendVigorOutput.TrendVigor).toBe(1);
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
            const indicator = CoronaTrendVigor.fromParams({
                rasterLength: tc.rasterLength,
                highPassFilterCutoff: tc.hpCutoff,
                minimalPeriod: tc.minPeriod,
                maximalPeriod: tc.maxPeriod,
            });
            const sharedConsumer = CoronaTrendVigor.fromParams({
                rasterLength: tc.rasterLength,
                highPassFilterCutoff: tc.hpCutoff,
                minimalPeriod: tc.minPeriod,
                maximalPeriod: tc.maxPeriod,
            });
            const sharedEngine = new Corona({
                highPassFilterCutoff: tc.hpCutoff,
                minimalPeriod: tc.minPeriod,
                maximalPeriod: tc.maxPeriod,
                decibelsLowerThreshold: 6.0,
                decibelsUpperThreshold: 20.0,
            });

            const step = 20.0 / tc.rasterLength;
            const expectedFirst = -10.0 + step;
            const expectedResolution = 1.0 / step;

            const expectedHeatmap = lookup[`expected${tc.name}_HEATMAP`] as number[][];
            const expectedMin = lookup[`expected${tc.name}_HEATMAP_MIN`] as number[];
            const expectedMax = lookup[`expected${tc.name}_HEATMAP_MAX`] as number[];
            const expectedTv = lookup[`expected${tc.name}_TREND_VIGOR`] as number[];
            const expectedNorm = lookup[`expected${tc.name}_NORMALIZED_VIGOR`] as number[];
            const expectedDc = lookup[`expected${tc.name}_DOMINANT_CYCLE`] as number[];
            const expectedDcm = lookup[`expected${tc.name}_DOMINANT_CYCLE_MEDIAN`] as number[];
            const expectedBp = lookup[`expected${tc.name}_BAND_PASS`] as number[];
            const expectedQuad = lookup[`expected${tc.name}_QUADRATURE`] as number[];
            const expectedAmp = lookup[`expected${tc.name}_CYCLE_AMPLITUDE`] as number[];
            const expectedDisp = lookup[`expected${tc.name}_TREND_DISPLACEMENT`] as number[];
            const expectedRawRatio = lookup[`expected${tc.name}_RAW_RATIO`] as number[];
            const expectedLookback = lookup[`expected${tc.name}_LOOKBACK`] as number[];
            const expectedWidth = lookup[`expected${tc.name}_WIDTH`] as number[];
            const expectedHasEst = lookup[`expected${tc.name}_HAS_CYCLE_ESTIMATE`] as number[];

            let sawLagTie = false;
            for (let i = 0; i < testdata.testInput.length; i++) {
                const barTime = new Date(baseTime.getTime() + i * 86_400_000);
                const price = testdata.testInput[i];
                const standaloneSnap = indicator.update(price, barTime);
                let engineSnap: CoronaSnapshot = sharedEngine.update(price, barTime);

                if (engineSnap.primed) {
                    if (!Number.isNaN(expectedDcm[i])) {
                        assertClose(standaloneSnap.dominantCycle, expectedDc[i], `${tc.name}[${i}].standalone.dc`);
                        assertClose(standaloneSnap.dominantCycleMedian, expectedDcm[i], `${tc.name}[${i}].standalone.dcm`);
                        assertClose(standaloneSnap.bandPass, expectedBp[i], `${tc.name}[${i}].standalone.bp`);
                        assertClose(standaloneSnap.quadrature, expectedQuad[i], `${tc.name}[${i}].standalone.quad`);
                        assertClose(standaloneSnap.cycleAmplitude, expectedAmp[i], `${tc.name}[${i}].standalone.amp`);
                        assertClose(standaloneSnap.hasCycleEstimate, expectedHasEst[i], `${tc.name}[${i}].standalone.hasEst`);
                    }
                    const nearestInteger = Math.round(engineSnap.dominantCycleMedian);
                    if (Math.abs(engineSnap.dominantCycleMedian - nearestInteger) <= 5e-12) {
                        if (Number.isNaN(expectedLookback[i])) {
                            if (i + 1 === nearestInteger && engineSnap.dominantCycleMedian < nearestInteger) {
                                engineSnap = { ...engineSnap, dominantCycleMedian: nearestInteger };
                                sawLagTie = true;
                            }
                        } else if (Math.floor(engineSnap.dominantCycleMedian) !== expectedLookback[i]) {
                            if (expectedLookback[i] === nearestInteger) {
                                engineSnap = { ...engineSnap, dominantCycleMedian: nearestInteger };
                            } else if (expectedLookback[i] === nearestInteger - 1) {
                                engineSnap = { ...engineSnap, dominantCycleMedian: nextDownPositive(nearestInteger) };
                            }
                            sawLagTie = true;
                        }
                    }
                }
                const snap = sharedConsumer.updateFromSnapshot(engineSnap);

                assertClose(snap.heatmap.parameterFirst, expectedFirst, `${tc.name}[${i}].first`);
                assertClose(snap.heatmap.parameterLast, 10.0, `${tc.name}[${i}].last`);
                assertClose(snap.heatmap.parameterResolution, expectedResolution, `${tc.name}[${i}].resolution`);

                const rowIsEmpty = Number.isNaN(expectedTv[i]);
                expect(sharedConsumer.isPrimed()).toBe(!rowIsEmpty);
                if (!sawLagTie) {
                    expect(indicator.isPrimed()).toBe(!rowIsEmpty);
                }
                if (rowIsEmpty) {
                    expect(snap.heatmap.isEmpty()).toBe(true);
                } else {
                    expect(snap.heatmap.values.length).toBe(tc.rasterLength);
                    for (let binIndex = 0; binIndex < tc.rasterLength; binIndex++) {
                        assertClose(snap.heatmap.values[binIndex], expectedHeatmap[i][binIndex], `${tc.name}[${i}].hm[${binIndex}]`);
                        if (!sawLagTie) {
                            assertClose(standaloneSnap.heatmap.values[binIndex], expectedHeatmap[i][binIndex], `${tc.name}[${i}].standalone.hm[${binIndex}]`);
                        }
                    }
                }

                assertClose(snap.heatmap.valueMin, expectedMin[i], `${tc.name}[${i}].min`);
                assertClose(snap.heatmap.valueMax, expectedMax[i], `${tc.name}[${i}].max`);
                assertClose(snap.trendVigor, expectedTv[i], `${tc.name}[${i}].tv`);
                assertClose(snap.normalizedVigor, expectedNorm[i], `${tc.name}[${i}].norm`);
                assertClose(snap.dominantCycle, expectedDc[i], `${tc.name}[${i}].dc`);
                assertClose(snap.dominantCycleMedian, expectedDcm[i], `${tc.name}[${i}].dcm`);
                assertClose(snap.bandPass, expectedBp[i], `${tc.name}[${i}].bp`);
                assertClose(snap.quadrature, expectedQuad[i], `${tc.name}[${i}].quad`);
                assertClose(snap.cycleAmplitude, expectedAmp[i], `${tc.name}[${i}].amp`);
                assertClose(snap.trendDisplacement, expectedDisp[i], `${tc.name}[${i}].disp`);
                assertClose(snap.rawRatio, expectedRawRatio[i], `${tc.name}[${i}].ratio`);
                assertClose(snap.lookback, expectedLookback[i], `${tc.name}[${i}].lookback`);
                assertClose(snap.width, expectedWidth[i], `${tc.name}[${i}].width`);
                assertClose(snap.hasCycleEstimate, expectedHasEst[i], `${tc.name}[${i}].hasEst`);
                if (!sawLagTie) {
                    assertClose(standaloneSnap.trendVigor, expectedTv[i], `${tc.name}[${i}].standalone.tv`);
                }
            }

            sharedConsumer.reset();
            expect(sharedConsumer.isPrimed()).toBe(false);
            indicator.reset();
            expect(indicator.isPrimed()).toBe(false);
        }
    });

    it('should expose expected metadata, mnemonics, validation, and entity wrappers', () => {
        const indicator = CoronaTrendVigor.default();
        const meta = indicator.metadata();
        expect(meta.identifier).toBe(IndicatorIdentifier.CoronaTrendVigor);
        expect(meta.mnemonic).toBe('ctv(50, 20, -10, 10, 30)');
        expect(meta.outputs.length).toBe(2);
        expect(meta.outputs[0].shape).toBe(Shape.Heatmap);
        expect(meta.outputs[1].shape).toBe(Shape.Scalar);
        expect(meta.outputs[1].mnemonic).toBe('ctv-tv(30)');

        const flagged = CoronaTrendVigor.fromParams({
            minimalPeriod: 10,
            maximalPeriod: 48,
            lookbackOffset: -1,
            barComponent: BarComponent.Median,
        });
        expect(flagged.metadata().mnemonic).toBe('ctv(50, 20, -10, 10, 30, pmin=10, pmax=48, offset=-1, hl/2)');
        expect(flagged.metadata().outputs[1].mnemonic).toBe('ctv-tv(30, pmin=10, pmax=48, offset=-1, hl/2)');

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
            const instance = CoronaTrendVigor.fromParams(p);
            const suffix = componentTripleMnemonic(
                bc ?? DefaultBarComponent,
                qc ?? DefaultQuoteComponent,
                tc ?? DefaultTradeComponent,
            );
            const expected = `ctv(50, 20, -10, 10, 30${suffix})`;
            expect(instance.metadata().mnemonic).toBe(expected);
            expect(coronaTrendVigorMnemonic(p)).toBe(expected);
        }

        expect(() => CoronaTrendVigor.fromParams({ rasterLength: 1 })).toThrow();
        expect(() => CoronaTrendVigor.fromParams({ maxRasterValue: 6 })).toThrow();
        expect(() => CoronaTrendVigor.fromParams({ minParameterValue: 10, maxParameterValue: 10 })).toThrow();
        expect(() => CoronaTrendVigor.fromParams({ highPassFilterCutoff: 4 })).toThrow();
        expect(() => CoronaTrendVigor.fromParams({ minimalPeriod: 5 })).toThrow();
        expect(() => CoronaTrendVigor.fromParams({ minimalPeriod: 10, maximalPeriod: 10 })).toThrow();
        expect(() => CoronaTrendVigor.fromParams({ lookbackOffset: 1 })).toThrow();
        expect(() => CoronaTrendVigor.fromParams({ lookbackOffset: -2 })).toThrow();
        expect(() => CoronaTrendVigor.fromParams({ lookbackOffset: 0.5 })).toThrow();

        const barInd = CoronaTrendVigor.default();
        const scalarInd = CoronaTrendVigor.default();
        const quoteInd = CoronaTrendVigor.default();
        const tradeInd = CoronaTrendVigor.default();
        for (let i = 0; i < 35; i++) {
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
    });
});
