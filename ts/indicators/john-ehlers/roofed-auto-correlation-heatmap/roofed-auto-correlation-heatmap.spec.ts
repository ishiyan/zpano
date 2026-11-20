import {} from 'jasmine';

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
import { InputRequirement } from '../../core/input-requirement';
import { Heatmap } from '../../core/outputs/heatmap';
import { Shape } from '../../core/outputs/shape/shape';
import { Pane } from '../../core/pane';
import { Role } from '../../core/role';
import { VolumeUsage } from '../../core/volume-usage';
import { createIndicator } from '../../factory/factory';
import { RoofedAutoCorrelationHeatmapOutput } from './output';
import {
    RoofedAutoCorrelationHeatmap,
    RoofedAutoCorrelationHeatmapParams,
    defaultParams,
    roofedAutoCorrelationHeatmapMnemonic,
} from './roofed-auto-correlation-heatmap';
import * as testdata from './testdata';

const FIXTURE_TOLERANCE = 2e-11;
const A2_FIXTURE_TOLERANCE = 1e-10;
const ORACLE_TOLERANCE = 1e-13;

interface ReferenceCase {
    readonly name: string;
    readonly params: RoofedAutoCorrelationHeatmapParams;
    readonly expectedHeatmap: readonly (readonly number[])[];
    readonly expectedMin: readonly number[];
    readonly expectedMax: readonly number[];
}

const REFERENCE_CASES: readonly ReferenceCase[] = [
    {
        name: 'L03_L148_A0_HP48_LP10',
        params: { minLag: 3, maxLag: 48, averagingLength: 0, highpassPeriod: 48, smoothingPeriod: 10 },
        expectedHeatmap: testdata.expectedL03_L148_A0_HP48_LP10_HEATMAP,
        expectedMin: testdata.expectedL03_L148_A0_HP48_LP10_MIN,
        expectedMax: testdata.expectedL03_L148_A0_HP48_LP10_MAX,
    },
    {
        name: 'L03_L148_A2_HP48_LP10',
        params: { minLag: 3, maxLag: 48, averagingLength: 2, highpassPeriod: 48, smoothingPeriod: 10 },
        expectedHeatmap: testdata.expectedL03_L148_A2_HP48_LP10_HEATMAP,
        expectedMin: testdata.expectedL03_L148_A2_HP48_LP10_MIN,
        expectedMax: testdata.expectedL03_L148_A2_HP48_LP10_MAX,
    },
    {
        name: 'L03_L148_A3_HP48_LP10',
        params: { minLag: 3, maxLag: 48, averagingLength: 3, highpassPeriod: 48, smoothingPeriod: 10 },
        expectedHeatmap: testdata.expectedL03_L148_A3_HP48_LP10_HEATMAP,
        expectedMin: testdata.expectedL03_L148_A3_HP48_LP10_MIN,
        expectedMax: testdata.expectedL03_L148_A3_HP48_LP10_MAX,
    },
    {
        name: 'L03_L148_A20_HP48_LP10',
        params: { minLag: 3, maxLag: 48, averagingLength: 20, highpassPeriod: 48, smoothingPeriod: 10 },
        expectedHeatmap: testdata.expectedL03_L148_A20_HP48_LP10_HEATMAP,
        expectedMin: testdata.expectedL03_L148_A20_HP48_LP10_MIN,
        expectedMax: testdata.expectedL03_L148_A20_HP48_LP10_MAX,
    },
    {
        name: 'L03_L148_A100_HP48_LP10',
        params: { minLag: 3, maxLag: 48, averagingLength: 100, highpassPeriod: 48, smoothingPeriod: 10 },
        expectedHeatmap: testdata.expectedL03_L148_A100_HP48_LP10_HEATMAP,
        expectedMin: testdata.expectedL03_L148_A100_HP48_LP10_MIN,
        expectedMax: testdata.expectedL03_L148_A100_HP48_LP10_MAX,
    },
    {
        name: 'L03_L148_A1_HP48_LP10',
        params: { minLag: 3, maxLag: 48, averagingLength: 1, highpassPeriod: 48, smoothingPeriod: 10 },
        expectedHeatmap: testdata.expectedL03_L148_A1_HP48_LP10_HEATMAP,
        expectedMin: testdata.expectedL03_L148_A1_HP48_LP10_MIN,
        expectedMax: testdata.expectedL03_L148_A1_HP48_LP10_MAX,
    },
    {
        name: 'L03_L148_A0_HP48_LP10_5',
        params: { minLag: 3, maxLag: 48, averagingLength: 0, highpassPeriod: 48, smoothingPeriod: 10.5 },
        expectedHeatmap: testdata.expectedL03_L148_A0_HP48_LP10_5_HEATMAP,
        expectedMin: testdata.expectedL03_L148_A0_HP48_LP10_5_MIN,
        expectedMax: testdata.expectedL03_L148_A0_HP48_LP10_5_MAX,
    },
    {
        name: 'L03_L148_A0_HP48_LP20',
        params: { minLag: 3, maxLag: 48, averagingLength: 0, highpassPeriod: 48, smoothingPeriod: 20 },
        expectedHeatmap: testdata.expectedL03_L148_A0_HP48_LP20_HEATMAP,
        expectedMin: testdata.expectedL03_L148_A0_HP48_LP20_MIN,
        expectedMax: testdata.expectedL03_L148_A0_HP48_LP20_MAX,
    },
    {
        name: 'L03_L148_A0_HP80_LP10',
        params: { minLag: 3, maxLag: 48, averagingLength: 0, highpassPeriod: 80, smoothingPeriod: 10 },
        expectedHeatmap: testdata.expectedL03_L148_A0_HP80_LP10_HEATMAP,
        expectedMin: testdata.expectedL03_L148_A0_HP80_LP10_MIN,
        expectedMax: testdata.expectedL03_L148_A0_HP80_LP10_MAX,
    },
    {
        name: 'L03_L124_A0_HP24_LP10',
        params: { minLag: 3, maxLag: 24, averagingLength: 0, highpassPeriod: 24, smoothingPeriod: 10 },
        expectedHeatmap: testdata.expectedL03_L124_A0_HP24_LP10_HEATMAP,
        expectedMin: testdata.expectedL03_L124_A0_HP24_LP10_MIN,
        expectedMax: testdata.expectedL03_L124_A0_HP24_LP10_MAX,
    },
    {
        name: 'L012_L112_A12_HP48_LP10',
        params: { minLag: 12, maxLag: 12, averagingLength: 12, highpassPeriod: 48, smoothingPeriod: 10 },
        expectedHeatmap: testdata.expectedL012_L112_A12_HP48_LP10_HEATMAP,
        expectedMin: testdata.expectedL012_L112_A12_HP48_LP10_MIN,
        expectedMax: testdata.expectedL012_L112_A12_HP48_LP10_MAX,
    },
    {
        name: 'L00_L10_A3_HP48_LP10',
        params: { minLag: 0, maxLag: 0, averagingLength: 3, highpassPeriod: 48, smoothingPeriod: 10 },
        expectedHeatmap: testdata.expectedL00_L10_A3_HP48_LP10_HEATMAP,
        expectedMin: testdata.expectedL00_L10_A3_HP48_LP10_MIN,
        expectedMax: testdata.expectedL00_L10_A3_HP48_LP10_MAX,
    },
];

interface OracleRow {
    readonly highPass: number;
    readonly filtered: number;
    readonly bins: readonly number[];
}

function batchOracle(closes: readonly number[], params: RoofedAutoCorrelationHeatmapParams): OracleRow[] {
    const minLag = params.minLag ?? 3;
    const maxLag = params.maxLag ?? 48;
    const averagingLength = params.averagingLength ?? 0;
    const highpassPeriod = params.highpassPeriod ?? 48;
    const smoothingPeriod = params.smoothingPeriod ?? 10;

    const angle = (2.0 * Math.PI) / (Math.SQRT2 * highpassPeriod);
    const cosine = Math.cos(angle);
    const alpha = (cosine + Math.sin(angle) - 1.0) / cosine;
    const pole = 1.0 - alpha;
    const attenuation = Math.exp((-Math.SQRT2 * Math.PI) / smoothingPeriod);
    const hp0 = (1.0 - alpha / 2.0) * (1.0 - alpha / 2.0);
    const hp1 = 2.0 * pole;
    const hp2 = pole * pole;
    const ss2 = 2.0 * attenuation * Math.cos((Math.SQRT2 * Math.PI) / smoothingPeriod);
    const ss3 = -attenuation * attenuation;
    const ss1 = 1.0 - ss2 - ss3;

    const historyLength = Math.max(1, maxLag + (averagingLength > 0 ? averagingLength : maxLag));
    const highSeries = new Array<number>(closes.length).fill(0.0);
    const filteredSeries = new Array<number>(closes.length).fill(0.0);
    const correlations = new Array<number>(maxLag + 1).fill(0.0);
    const rows: OracleRow[] = [];

    for (let barIndex = 0; barIndex < closes.length; barIndex++) {
        const price = closes[barIndex];
        const x1 = barIndex >= 1 ? closes[barIndex - 1] : 0.0;
        const x2 = barIndex >= 2 ? closes[barIndex - 2] : 0.0;
        const h1 = barIndex >= 1 ? highSeries[barIndex - 1] : 0.0;
        const h2 = barIndex >= 2 ? highSeries[barIndex - 2] : 0.0;
        const f1 = barIndex >= 1 ? filteredSeries[barIndex - 1] : 0.0;
        const f2 = barIndex >= 2 ? filteredSeries[barIndex - 2] : 0.0;

        const high = hp0 * (price - 2.0 * x1 + x2) + hp1 * h1 - hp2 * h2;
        const filtered = (ss1 * (high + h1)) / 2.0 + ss2 * f1 + ss3 * f2;
        highSeries[barIndex] = high;
        filteredSeries[barIndex] = filtered;

        for (let lag = 0; lag <= maxLag; lag++) {
            const window = averagingLength > 0 ? averagingLength : lag;
            let sumX = 0.0;
            let sumY = 0.0;
            let sumXX = 0.0;
            let sumXY = 0.0;
            let sumYY = 0.0;
            for (let offset = 0; offset < window; offset++) {
                const sampleX = barIndex - offset >= 0 ? filteredSeries[barIndex - offset] : 0.0;
                const sampleY = barIndex - lag - offset >= 0 ? filteredSeries[barIndex - lag - offset] : 0.0;
                sumX += sampleX;
                sumY += sampleY;
                sumXX += sampleX * sampleX;
                sumXY += sampleX * sampleY;
                sumYY += sampleY * sampleY;
            }
            const denominator = (window * sumXX - sumX * sumX) * (window * sumYY - sumY * sumY);
            if (denominator > 0.0) {
                correlations[lag] = (window * sumXY - sumX * sumY) / Math.sqrt(denominator);
            }
            correlations[lag] = 0.5 * (correlations[lag] + 1.0);
        }

        rows.push({
            highPass: high,
            filtered,
            bins: barIndex + 1 >= historyLength ? correlations.slice(minLag, maxLag + 1) : [],
        });
    }
    return rows;
}

function expectClose(actual: number, expected: number, tolerance: number, label: string): void {
    if (Number.isNaN(expected)) {
        expect(Number.isNaN(actual)).withContext(`${label} should be NaN`).toBeTrue();
        return;
    }
    const bound = Math.max(tolerance, Math.abs(expected) * tolerance);
    expect(Math.abs(actual - expected)).withContext(`${label}: ${actual} vs ${expected}`).toBeLessThanOrEqual(bound);
}

describe('RoofedAutoCorrelationHeatmap', () => {
    it('matches all 12 pre-created reference fixtures and Listing 8-2 oracle', () => {
        for (const tc of REFERENCE_CASES) {
            const indicator = new RoofedAutoCorrelationHeatmap(tc.params);
            const oracle = batchOracle(testdata.testInput, tc.params);
            const minLag = tc.params.minLag ?? 3;
            const maxLag = tc.params.maxLag ?? 48;
            const expectedBins = maxLag - minLag + 1;
            const caseTolerance = tc.params.averagingLength === 2 ? A2_FIXTURE_TOLERANCE : FIXTURE_TOLERANCE;

            for (let barIndex = 0; barIndex < testdata.testInput.length; barIndex++) {
                const stamp = new Date(Date.UTC(2026, 0, 1 + barIndex));
                const column = indicator.update(testdata.testInput[barIndex], stamp);
                const state = indicator.state;

                expectClose(state.highPass, oracle[barIndex].highPass, ORACLE_TOLERANCE, `${tc.name}[${barIndex}].highPass`);
                expectClose(state.filtered, oracle[barIndex].filtered, ORACLE_TOLERANCE, `${tc.name}[${barIndex}].filtered`);
                expect(column.parameterFirst).toBe(minLag);
                expect(column.parameterLast).toBe(maxLag);
                expect(column.parameterResolution).toBe(1.0);

                const expectedRow = tc.expectedHeatmap[barIndex];
                const expectedMin = tc.expectedMin[barIndex];
                const expectedMax = tc.expectedMax[barIndex];
                if (Number.isNaN(expectedMin)) {
                    expect(column.isEmpty()).withContext(`${tc.name}[${barIndex}] empty`).toBeTrue();
                    expect(Number.isNaN(column.valueMin)).toBeTrue();
                    expect(Number.isNaN(column.valueMax)).toBeTrue();
                    continue;
                }

                expect(column.values.length).toBe(expectedBins);
                expectClose(column.valueMin, expectedMin, caseTolerance, `${tc.name}[${barIndex}].valueMin`);
                expectClose(column.valueMax, expectedMax, caseTolerance, `${tc.name}[${barIndex}].valueMax`);
                for (let binIndex = 0; binIndex < expectedBins; binIndex++) {
                    expectClose(column.values[binIndex], expectedRow[binIndex], caseTolerance, `${tc.name}[${barIndex}][${binIndex}]`);
                    expectClose(column.values[binIndex], oracle[barIndex].bins[binIndex], ORACLE_TOLERANCE, `${tc.name}[${barIndex}] oracle[${binIndex}]`);
                }
            }
        }
    });

    it('builds mnemonics, metadata, descriptor, and factory instances', () => {
        expect(RoofedAutoCorrelationHeatmapOutput.Value).toBe(0);
        expect(roofedAutoCorrelationHeatmapMnemonic()).toBe('rach(3, 48, 0, 48, 10)');
        expect(roofedAutoCorrelationHeatmapMnemonic(defaultParams())).toBe('rach(3, 48, 0, 48, 10)');
        expect(roofedAutoCorrelationHeatmapMnemonic({ barComponent: BarComponent.Median })).toBe('rach(3, 48, 0, 48, 10, hl/2)');
        expect(roofedAutoCorrelationHeatmapMnemonic({ quoteComponent: QuoteComponent.Weighted })).toBe('rach(3, 48, 0, 48, 10, (bbs+aas)/(bs+as))');
        expect(roofedAutoCorrelationHeatmapMnemonic({ tradeComponent: TradeComponent.Volume })).toBe('rach(3, 48, 0, 48, 10, v)');
        expect(
            roofedAutoCorrelationHeatmapMnemonic({
                minLag: 0,
                maxLag: 0,
                averagingLength: 3,
                highpassPeriod: 24,
                smoothingPeriod: 10.5,
                barComponent: BarComponent.Typical,
                quoteComponent: QuoteComponent.Bid,
            }),
        ).toBe('rach(0, 0, 3, 24, 10.5, hlc/3, b)');
        expect(
            roofedAutoCorrelationHeatmapMnemonic({
                minLag: 12,
                maxLag: 24,
                averagingLength: 12,
                quoteComponent: QuoteComponent.Ask,
                tradeComponent: TradeComponent.Volume,
            }),
        ).toBe('rach(12, 24, 12, 48, 10, a, v)');

        const indicator = RoofedAutoCorrelationHeatmap.default();
        const meta = indicator.metadata();
        expect(meta.identifier).toBe(IndicatorIdentifier.RoofedAutoCorrelationHeatmap);
        expect(meta.mnemonic).toBe('rach(3, 48, 0, 48, 10)');
        expect(meta.description).toBe('Roofed autocorrelation heatmap rach(3, 48, 0, 48, 10)');
        expect(meta.outputs.length).toBe(1);
        expect(meta.outputs[0].shape).toBe(Shape.Heatmap);
        expect(meta.outputs[0].mnemonic).toBe('rach(3, 48, 0, 48, 10)');

        const desc = descriptorOf(IndicatorIdentifier.RoofedAutoCorrelationHeatmap);
        expect(desc).toBeDefined();
        expect(desc!.family).toBe('John Ehlers');
        expect(desc!.adaptivity).toBe(Adaptivity.Static);
        expect(desc!.inputRequirement).toBe(InputRequirement.ScalarInput);
        expect(desc!.volumeUsage).toBe(VolumeUsage.NoVolume);
        expect(desc!.outputs[0].role).toBe(Role.Correlation);
        expect(desc!.outputs[0].pane).toBe(Pane.Own);

        const fromDefaultFactory = createIndicator(IndicatorIdentifier.RoofedAutoCorrelationHeatmap, {});
        expect(fromDefaultFactory.metadata().mnemonic).toBe('rach(3, 48, 0, 48, 10)');

        const fromCustomFactory = createIndicator(IndicatorIdentifier.RoofedAutoCorrelationHeatmap, {
            minLag: 0,
            maxLag: 0,
            averagingLength: 3,
            highpassPeriod: 24,
            smoothingPeriod: 10.5,
            barComponent: BarComponent.Median,
        });
        expect(fromCustomFactory.metadata().mnemonic).toBe('rach(0, 0, 3, 24, 10.5, hl/2)');

        expect(() => createIndicator(IndicatorIdentifier.RoofedAutoCorrelationHeatmap, { unknownKey: 1 })).toThrow();
    });

    it('validates parameters, handles NaN/overflow/reset, zero-variance decay, and entity routing', () => {
        const invalidCases: Record<string, unknown>[] = [
            { minLag: -1 },
            { minLag: 1.5 },
            { minLag: 10, maxLag: 9 },
            { averagingLength: -1 },
            { averagingLength: 2.5 },
            { highpassPeriod: 4.9 },
            { highpassPeriod: Number.NaN },
            { smoothingPeriod: 1.9 },
            { smoothingPeriod: Number.POSITIVE_INFINITY },
            { minLag: null },
            { barComponent: 99 },
        ];
        for (const bad of invalidCases) {
            expect(() => new RoofedAutoCorrelationHeatmap(bad as RoofedAutoCorrelationHeatmapParams)).toThrow();
        }

        const subject = new RoofedAutoCorrelationHeatmap({ minLag: 3, maxLag: 12, averagingLength: 5, highpassPeriod: 24, smoothingPeriod: 10 });
        const clean = new RoofedAutoCorrelationHeatmap({ minLag: 3, maxLag: 12, averagingLength: 5, highpassPeriod: 24, smoothingPeriod: 10 });
        const stamp = new Date(Date.UTC(2026, 2, 1));

        for (let barIndex = 0; barIndex < 40; barIndex++) {
            const before = subject.state;
            const skipped = subject.update(Number.NaN, stamp);
            expect(skipped.isEmpty()).toBeTrue();
            for (const badSample of [Number.POSITIVE_INFINITY, Number.NEGATIVE_INFINITY, 1e308]) {
                expect(() => subject.update(badSample, stamp)).toThrow();
                expect(subject.state).toEqual(before);
            }
            const actual = subject.update(testdata.testInput[barIndex], stamp);
            const expected = clean.update(testdata.testInput[barIndex], stamp);
            expect(actual.values).toEqual(expected.values);
        }

        subject.reset();
        expect(subject.samplesSeen).toBe(0);
        expect(subject.isPrimed()).toBeFalse();

        const flatInd = new RoofedAutoCorrelationHeatmap({ minLag: 0, maxLag: 2, averagingLength: 3 });
        for (let step = 1; step <= 6; step++) {
            const col = flatInd.update(0.0, stamp);
            const expected = 1.0 - Math.pow(0.5, step);
            for (let lag = 0; lag <= 2; lag++) {
                expectClose(flatInd.state.correlations[lag], expected, 1e-15, `flat state ${step}`);
            }
            if (step >= 5) {
                for (let lag = 0; lag <= 2; lag++) {
                    expectClose(col.values[lag], expected, 1e-15, `flat output ${step}`);
                }
            }
        }

        const scalarInd = RoofedAutoCorrelationHeatmap.default();
        const barInd = RoofedAutoCorrelationHeatmap.default();
        const quoteInd = RoofedAutoCorrelationHeatmap.default();
        const tradeInd = RoofedAutoCorrelationHeatmap.default();

        for (let barIndex = 0; barIndex < testdata.testInput.length; barIndex++) {
            const barTime = new Date(Date.UTC(2026, 0, 1 + barIndex));
            const price = testdata.testInput[barIndex];

            const scalar = new Scalar();
            scalar.time = barTime;
            scalar.value = price;

            const bar = new Bar();
            bar.time = barTime;
            bar.open = testdata.testOpen[barIndex];
            bar.high = testdata.testHigh[barIndex];
            bar.low = testdata.testLow[barIndex];
            bar.close = price;
            bar.volume = 1000;

            const quote = new Quote();
            quote.time = barTime;
            quote.bidPrice = price;
            quote.askPrice = price;
            quote.bidSize = 10;
            quote.askSize = 10;

            const trade = new Trade();
            trade.time = barTime;
            trade.price = price;
            trade.volume = 10;

            const outScalar = scalarInd.updateScalar(scalar)[0] as Heatmap;
            const outBar = barInd.updateBar(bar)[0] as Heatmap;
            const outQuote = quoteInd.updateQuote(quote)[0] as Heatmap;
            const outTrade = tradeInd.updateTrade(trade)[0] as Heatmap;

            expect(outScalar.values).toEqual(outBar.values);
            expect(outScalar.values).toEqual(outQuote.values);
            expect(outScalar.values).toEqual(outTrade.values);

            if (!outScalar.isEmpty()) {
                const saved = scalarInd.output.values.slice();
                outScalar.values[0] = 9999;
                expect(scalarInd.output.values).toEqual(saved);
            }
        }
    });
});
