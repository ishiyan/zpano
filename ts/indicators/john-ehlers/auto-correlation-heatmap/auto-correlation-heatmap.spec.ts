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
import {
    AutoCorrelationHeatmap,
    AutoCorrelationHeatmapParams,
    autoCorrelationHeatmapMnemonic,
    defaultParams,
} from './auto-correlation-heatmap';
import { AutoCorrelationHeatmapOutput } from './output';
import * as testdata from './testdata';

const FIXTURE_TOLERANCE = 1e-13;
const ORACLE_TOLERANCE = 1e-13;

interface ReferenceCase {
    readonly name: string;
    readonly params: AutoCorrelationHeatmapParams;
    readonly expectedHeatmap: readonly (readonly number[])[];
    readonly expectedMin: readonly number[];
    readonly expectedMax: readonly number[];
}

const REFERENCE_CASES: readonly ReferenceCase[] = [
    {
        name: 'L20_P20_H99',
        params: { length: 20, smoothingPeriod: 20, maxLag: 99 },
        expectedHeatmap: testdata.expectedL20_P20_H99_HEATMAP,
        expectedMin: testdata.expectedL20_P20_H99_MIN,
        expectedMax: testdata.expectedL20_P20_H99_MAX,
    },
    {
        name: 'L2_P20_H99',
        params: { length: 2, smoothingPeriod: 20, maxLag: 99 },
        expectedHeatmap: testdata.expectedL2_P20_H99_HEATMAP,
        expectedMin: testdata.expectedL2_P20_H99_MIN,
        expectedMax: testdata.expectedL2_P20_H99_MAX,
    },
    {
        name: 'L3_P20_H99',
        params: { length: 3, smoothingPeriod: 20, maxLag: 99 },
        expectedHeatmap: testdata.expectedL3_P20_H99_HEATMAP,
        expectedMin: testdata.expectedL3_P20_H99_MIN,
        expectedMax: testdata.expectedL3_P20_H99_MAX,
    },
    {
        name: 'L5_P20_H99',
        params: { length: 5, smoothingPeriod: 20, maxLag: 99 },
        expectedHeatmap: testdata.expectedL5_P20_H99_HEATMAP,
        expectedMin: testdata.expectedL5_P20_H99_MIN,
        expectedMax: testdata.expectedL5_P20_H99_MAX,
    },
    {
        name: 'L40_P20_H99',
        params: { length: 40, smoothingPeriod: 20, maxLag: 99 },
        expectedHeatmap: testdata.expectedL40_P20_H99_HEATMAP,
        expectedMin: testdata.expectedL40_P20_H99_MIN,
        expectedMax: testdata.expectedL40_P20_H99_MAX,
    },
    {
        name: 'L100_P20_H99',
        params: { length: 100, smoothingPeriod: 20, maxLag: 99 },
        expectedHeatmap: testdata.expectedL100_P20_H99_HEATMAP,
        expectedMin: testdata.expectedL100_P20_H99_MIN,
        expectedMax: testdata.expectedL100_P20_H99_MAX,
    },
    {
        name: 'L20_P10_H99',
        params: { length: 20, smoothingPeriod: 10, maxLag: 99 },
        expectedHeatmap: testdata.expectedL20_P10_H99_HEATMAP,
        expectedMin: testdata.expectedL20_P10_H99_MIN,
        expectedMax: testdata.expectedL20_P10_H99_MAX,
    },
    {
        name: 'L20_P10_5_H99',
        params: { length: 20, smoothingPeriod: 10.5, maxLag: 99 },
        expectedHeatmap: testdata.expectedL20_P10_5_H99_HEATMAP,
        expectedMin: testdata.expectedL20_P10_5_H99_MIN,
        expectedMax: testdata.expectedL20_P10_5_H99_MAX,
    },
    {
        name: 'L20_P40_H99',
        params: { length: 20, smoothingPeriod: 40, maxLag: 99 },
        expectedHeatmap: testdata.expectedL20_P40_H99_HEATMAP,
        expectedMin: testdata.expectedL20_P40_H99_MIN,
        expectedMax: testdata.expectedL20_P40_H99_MAX,
    },
    {
        name: 'L20_P20_H32',
        params: { length: 20, smoothingPeriod: 20, maxLag: 32 },
        expectedHeatmap: testdata.expectedL20_P20_H32_HEATMAP,
        expectedMin: testdata.expectedL20_P20_H32_MIN,
        expectedMax: testdata.expectedL20_P20_H32_MAX,
    },
    {
        name: 'L2_P20_H32',
        params: { length: 2, smoothingPeriod: 20, maxLag: 32 },
        expectedHeatmap: testdata.expectedL2_P20_H32_HEATMAP,
        expectedMin: testdata.expectedL2_P20_H32_MIN,
        expectedMax: testdata.expectedL2_P20_H32_MAX,
    },
    {
        name: 'L20_P20_H0',
        params: { length: 20, smoothingPeriod: 20, maxLag: 0 },
        expectedHeatmap: testdata.expectedL20_P20_H0_HEATMAP,
        expectedMin: testdata.expectedL20_P20_H0_MIN,
        expectedMax: testdata.expectedL20_P20_H0_MAX,
    },
];

function batchOracle(
    closes: readonly number[],
    length: number,
    smoothingPeriod: number,
    maxLag: number,
): { filtered: number; bins: number[] | null }[] {
    const filteredSeries = new Array<number>(closes.length).fill(0.0);
    const correlations = new Array<number>(101).fill(0.0);
    const angle = (Math.SQRT2 * Math.PI) / smoothingPeriod;
    const attenuation = Math.exp(-angle);
    const coeffC2 = 2.0 * attenuation * Math.cos(angle);
    const coeffC3 = -attenuation * attenuation;
    const coeffC1 = (1.0 + coeffC2 - coeffC3) / 4.0;
    const primeCount = Math.max(4, length + maxLag);
    const rows: { filtered: number; bins: number[] | null }[] = [];

    for (let barIndex = 0; barIndex < closes.length; barIndex++) {
        const price = closes[barIndex];
        if (barIndex < 3) {
            filteredSeries[barIndex] = price;
        } else {
            filteredSeries[barIndex] =
                (1.0 - coeffC1) * price +
                (2.0 * coeffC1 - coeffC2) * closes[barIndex - 1] -
                (coeffC1 + coeffC3) * closes[barIndex - 2] +
                coeffC2 * filteredSeries[barIndex - 1] +
                coeffC3 * filteredSeries[barIndex - 2];
        }
        for (let lag = 0; lag <= maxLag; lag++) {
            let sumX = 0.0;
            let sumY = 0.0;
            let sumXX = 0.0;
            let sumXY = 0.0;
            let sumYY = 0.0;
            for (let offset = 0; offset < length; offset++) {
                const sampleX = barIndex - offset >= 0 ? filteredSeries[barIndex - offset] : 0.0;
                const sampleY = barIndex - lag - offset >= 0 ? filteredSeries[barIndex - lag - offset] : 0.0;
                sumX += sampleX;
                sumY += sampleY;
                sumXX += sampleX * sampleX;
                sumXY += sampleX * sampleY;
                sumYY += sampleY * sampleY;
            }
            const varX = length * sumXX - sumX * sumX;
            const varY = length * sumYY - sumY * sumY;
            if (varX > 0.0 && varY > 0.0) {
                correlations[lag + 1] = (length * sumXY - sumX * sumY) / Math.sqrt(varX * varY);
            }
        }
        if (barIndex + 1 < primeCount) {
            rows.push({ filtered: filteredSeries[barIndex], bins: null });
        } else {
            rows.push({
                filtered: filteredSeries[barIndex],
                bins: correlations.slice(1, maxLag + 2),
            });
        }
    }
    return rows;
}

function assertClose(actual: number, expected: number, label: string, tolerance: number = FIXTURE_TOLERANCE): void {
    if (Number.isNaN(expected)) {
        expect(Number.isNaN(actual)).withContext(label).toBeTrue();
        return;
    }
    expect(Number.isNaN(actual)).withContext(label).toBeFalse();
    expect(Math.abs(actual - expected)).withContext(`${label}: ${actual} vs ${expected}`).toBeLessThanOrEqual(tolerance);
}

describe('AutoCorrelationHeatmap', () => {
    it('matches all 12 saved reference fixtures and independent batch oracle', () => {
        expect(testdata.testInput.length).toBe(252);

        for (const testCase of REFERENCE_CASES) {
            const indicator = new AutoCorrelationHeatmap(testCase.params);
            const length = testCase.params.length!;
            const smoothingPeriod = testCase.params.smoothingPeriod!;
            const maxLag = testCase.params.maxLag!;
            const warmup = Math.max(4, length + maxLag) - 1;
            expect(indicator.warmupBars).withContext(testCase.name).toBe(warmup);

            const oracle = batchOracle(testdata.testInput, length, smoothingPeriod, maxLag);
            const binCount = maxLag + 1;

            for (let barIndex = 0; barIndex < testdata.testInput.length; barIndex++) {
                const result = indicator.update(testdata.testInput[barIndex]);
                expect(result.parameterFirst).toBe(0);
                expect(result.parameterLast).toBe(maxLag);
                expect(result.parameterResolution).toBe(1);
                expect(indicator.isPrimed()).toBe(barIndex >= warmup);
                assertClose(
                    indicator.state.filtered,
                    oracle[barIndex].filtered,
                    `${testCase.name}[${barIndex}].filtered`,
                    ORACLE_TOLERANCE,
                );

                const expectedRow = testCase.expectedHeatmap[barIndex];
                if (barIndex < warmup) {
                    expect(result.isEmpty()).toBeTrue();
                    expect(Number.isNaN(result.valueMin)).toBeTrue();
                    expect(Number.isNaN(result.valueMax)).toBeTrue();
                    expect(Number.isNaN(testCase.expectedMin[barIndex])).toBeTrue();
                    expect(Number.isNaN(testCase.expectedMax[barIndex])).toBeTrue();
                    continue;
                }

                expect(result.values.length).toBe(binCount);
                assertClose(result.valueMin, testCase.expectedMin[barIndex], `${testCase.name}[${barIndex}].min`);
                assertClose(result.valueMax, testCase.expectedMax[barIndex], `${testCase.name}[${barIndex}].max`);
                const oracleBins = oracle[barIndex].bins!;
                for (let lag = 0; lag < binCount; lag++) {
                    assertClose(result.values[lag], expectedRow[lag], `${testCase.name}[${barIndex}][${lag}]`);
                    assertClose(result.values[lag], oracleBins[lag], `${testCase.name}[${barIndex}][${lag}].oracle`, ORACLE_TOLERANCE);
                }
            }
        }
    });

    it('handles NaN skip, overflow rollback, reset, analytical sine, and degenerate retention', () => {
        const params: AutoCorrelationHeatmapParams = { length: 5, smoothingPeriod: 10, maxLag: 12 };
        const subject = new AutoCorrelationHeatmap(params);
        const clean = new AutoCorrelationHeatmap(params);

        for (let barIndex = 0; barIndex < 50; barIndex++) {
            const before = subject.state;
            const skipped = subject.update(Number.NaN);
            expect(skipped.isEmpty()).toBeTrue();
            expect(subject.state).toEqual(before);

            for (const badSample of [Number.POSITIVE_INFINITY, Number.NEGATIVE_INFINITY, 1e308]) {
                expect(() => subject.update(badSample)).toThrow();
                expect(subject.state).toEqual(before);
            }

            const actual = subject.update(testdata.testInput[barIndex]);
            const expected = clean.update(testdata.testInput[barIndex]);
            expect(subject.isPrimed()).toBe(barIndex + 1 >= 17);
            expect(actual.values).toEqual(expected.values);
        }

        subject.reset();
        clean.reset();
        expect(subject.samplesSeen).toBe(0);
        expect(subject.isPrimed()).toBeFalse();
        for (let barIndex = 0; barIndex < 50; barIndex++) {
            expect(subject.update(testdata.testInput[barIndex]).values).toEqual(
                clean.update(testdata.testInput[barIndex]).values,
            );
        }

        const sineInd = new AutoCorrelationHeatmap({ length: 20, smoothingPeriod: 20, maxLag: 40 });
        let column = Heatmap.newEmptyHeatmap(new Date(0), 0, 40, 1);
        for (let step = 0; step < 1000; step++) {
            column = sineInd.update(Math.sin((2.0 * Math.PI * step) / 20.0));
        }
        for (let lag = 0; lag <= 40; lag++) {
            assertClose(column.values[lag], Math.cos((2.0 * Math.PI * lag) / 20.0), `sine[${lag}]`, 1e-11);
        }

        const flatInd = new AutoCorrelationHeatmap({ length: 3, smoothingPeriod: 20, maxLag: 2 });
        let flatOut = Heatmap.newEmptyHeatmap(new Date(0), 0, 2, 1);
        for (let step = 0; step < 30; step++) {
            flatOut = flatInd.update(0.0);
        }
        expect(flatOut.values).toEqual([0, 0, 0]);
    });

    it('validates mnemonics, metadata, descriptor, entity updates, and factory', () => {
        const mnemonicCases: readonly [AutoCorrelationHeatmapParams | undefined, string][] = [
            [undefined, 'ach(20, 20, 99)'],
            [defaultParams(), 'ach(20, 20, 99)'],
            [{ length: 20, smoothingPeriod: 10.5, maxLag: 0 }, 'ach(20, 10.5, 0)'],
            [{ barComponent: BarComponent.Median }, 'ach(20, 20, 99, hl/2)'],
            [{ quoteComponent: QuoteComponent.Bid }, 'ach(20, 20, 99, b)'],
            [{ tradeComponent: TradeComponent.Volume }, 'ach(20, 20, 99, v)'],
            [
                { barComponent: BarComponent.Typical, quoteComponent: QuoteComponent.Ask },
                'ach(20, 20, 99, hlc/3, a)',
            ],
            [
                {
                    barComponent: BarComponent.Open,
                    quoteComponent: QuoteComponent.Weighted,
                    tradeComponent: TradeComponent.Volume,
                },
                'ach(20, 20, 99, o, (bbs+aas)/(bs+as), v)',
            ],
        ];

        for (const [params, wantMnemonic] of mnemonicCases) {
            expect(autoCorrelationHeatmapMnemonic(params)).toBe(wantMnemonic);
            const instance = new AutoCorrelationHeatmap(params);
            const meta = instance.metadata();
            expect(meta.identifier).toBe(IndicatorIdentifier.AutoCorrelationHeatmap);
            expect(meta.mnemonic).toBe(wantMnemonic);
            expect(meta.description).toBe(`Autocorrelation heatmap ${wantMnemonic}`);
            expect(meta.outputs.length).toBe(1);
            expect(meta.outputs[AutoCorrelationHeatmapOutput.Value].shape).toBe(Shape.Heatmap);
            expect(meta.outputs[AutoCorrelationHeatmapOutput.Value].mnemonic).toBe(wantMnemonic);
        }

        const desc = descriptorOf(IndicatorIdentifier.AutoCorrelationHeatmap);
        expect(desc).toBeDefined();
        expect(desc!.family).toBe('John Ehlers');
        expect(desc!.adaptivity).toBe(Adaptivity.Static);
        expect(desc!.inputRequirement).toBe(InputRequirement.ScalarInput);
        expect(desc!.volumeUsage).toBe(VolumeUsage.NoVolume);
        expect(desc!.outputs.length).toBe(1);
        expect(desc!.outputs[0].shape).toBe(Shape.Heatmap);
        expect(desc!.outputs[0].role).toBe(Role.Correlation);
        expect(desc!.outputs[0].pane).toBe(Pane.Own);

        const baseTime = new Date(Date.UTC(2026, 0, 1));
        const barInd = AutoCorrelationHeatmap.default();
        const scalarInd = AutoCorrelationHeatmap.default();
        const quoteInd = AutoCorrelationHeatmap.default();
        const tradeInd = AutoCorrelationHeatmap.default();
        const openInd = new AutoCorrelationHeatmap({ barComponent: BarComponent.Open });
        const openRef = AutoCorrelationHeatmap.default();

        for (let barIndex = 0; barIndex < testdata.testInput.length; barIndex++) {
            const sampleTime = new Date(baseTime.getTime() + barIndex * 86400000);
            const closePrice = testdata.testInput[barIndex];
            const openPrice = testdata.testOpen[barIndex];

            const bar = new Bar();
            bar.time = sampleTime;
            bar.open = openPrice;
            bar.high = testdata.testHigh[barIndex];
            bar.low = testdata.testLow[barIndex];
            bar.close = closePrice;
            bar.volume = 1000;

            const scalar = new Scalar();
            scalar.time = sampleTime;
            scalar.value = closePrice;

            const quote = new Quote();
            quote.time = sampleTime;
            quote.bidPrice = closePrice;
            quote.askPrice = closePrice;
            quote.bidSize = 10;
            quote.askSize = 10;

            const trade = new Trade();
            trade.time = sampleTime;
            trade.price = closePrice;
            trade.volume = 10;

            const outBar = barInd.updateBar(bar)[0] as Heatmap;
            const outScalar = scalarInd.updateScalar(scalar)[0] as Heatmap;
            const outQuote = quoteInd.updateQuote(quote)[0] as Heatmap;
            const outTrade = tradeInd.updateTrade(trade)[0] as Heatmap;
            expect(outBar.values).toEqual(outScalar.values);
            expect(outQuote.values).toEqual(outScalar.values);
            expect(outTrade.values).toEqual(outScalar.values);

            const openScalar = new Scalar();
            openScalar.time = sampleTime;
            openScalar.value = openPrice;
            const outOpenBar = openInd.updateBar(bar)[0] as Heatmap;
            const outOpenScalar = openRef.updateScalar(openScalar)[0] as Heatmap;
            expect(outOpenBar.values).toEqual(outOpenScalar.values);

            if (!outScalar.isEmpty()) {
                const stateBefore = scalarInd.state;
                outScalar.values[0] = 9999;
                expect(scalarInd.state).toEqual(stateBefore);
            }
        }

        const factoryDefault = createIndicator(IndicatorIdentifier.AutoCorrelationHeatmap);
        expect(factoryDefault.metadata().mnemonic).toBe('ach(20, 20, 99)');
        const factoryCustom = createIndicator(IndicatorIdentifier.AutoCorrelationHeatmap, {
            length: 20,
            smoothingPeriod: 10.5,
            maxLag: 0,
            barComponent: BarComponent.Median,
        });
        expect(factoryCustom.metadata().mnemonic).toBe('ach(20, 10.5, 0, hl/2)');

        for (const badParams of [
            { length: 1 },
            { length: 2.5 },
            { length: null },
            { smoothingPeriod: 0.5 },
            { smoothingPeriod: Number.NaN },
            { smoothingPeriod: Number.POSITIVE_INFINITY },
            { maxLag: -1 },
            { maxLag: 100 },
            { maxLag: 1.5 },
            { unknownKey: 1 },
        ]) {
            expect(() => new AutoCorrelationHeatmap(badParams as never)).toThrow();
            expect(() => createIndicator(IndicatorIdentifier.AutoCorrelationHeatmap, badParams as never)).toThrow();
        }
    });
});
