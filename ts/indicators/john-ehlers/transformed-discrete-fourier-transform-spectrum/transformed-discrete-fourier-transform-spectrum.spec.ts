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
import { TransformedDiscreteFourierTransformSpectrumOutput } from './output';
import * as testdata from './testdata';
import {
    SpectrumOutput,
    TransformedDiscreteFourierTransformSpectrum,
    TransformedDiscreteFourierTransformSpectrumParams,
    defaultParams,
    transformedDiscreteFourierTransformSpectrumDominantCycleMnemonic,
    transformedDiscreteFourierTransformSpectrumMnemonic,
} from './transformed-discrete-fourier-transform-spectrum';

const FIXTURE_TOLERANCE = 1e-12;

interface ReferenceCase {
    readonly name: string;
    readonly params: TransformedDiscreteFourierTransformSpectrumParams;
    readonly expectedHeatmap: readonly (readonly number[])[];
    readonly expectedMin: readonly number[];
    readonly expectedMax: readonly number[];
    readonly expectedDC: readonly number[];
}

const REFERENCE_CASES: readonly ReferenceCase[] = [
    {
        name: 'P8_Q50_W50_HP40',
        params: { minPeriod: 8, maxPeriod: 50, windowLength: 50, highpassPeriod: 40 },
        expectedHeatmap: testdata.expectedP8_Q50_W50_HP40_HEATMAP,
        expectedMin: testdata.expectedP8_Q50_W50_HP40_MIN,
        expectedMax: testdata.expectedP8_Q50_W50_HP40_MAX,
        expectedDC: testdata.expectedP8_Q50_W50_HP40_DOMINANT_CYCLE,
    },
    {
        name: 'P8_Q50_W1_HP40',
        params: { minPeriod: 8, maxPeriod: 50, windowLength: 1, highpassPeriod: 40 },
        expectedHeatmap: testdata.expectedP8_Q50_W1_HP40_HEATMAP,
        expectedMin: testdata.expectedP8_Q50_W1_HP40_MIN,
        expectedMax: testdata.expectedP8_Q50_W1_HP40_MAX,
        expectedDC: testdata.expectedP8_Q50_W1_HP40_DOMINANT_CYCLE,
    },
    {
        name: 'P8_Q50_W20_HP40',
        params: { minPeriod: 8, maxPeriod: 50, windowLength: 20, highpassPeriod: 40 },
        expectedHeatmap: testdata.expectedP8_Q50_W20_HP40_HEATMAP,
        expectedMin: testdata.expectedP8_Q50_W20_HP40_MIN,
        expectedMax: testdata.expectedP8_Q50_W20_HP40_MAX,
        expectedDC: testdata.expectedP8_Q50_W20_HP40_DOMINANT_CYCLE,
    },
    {
        name: 'P8_Q50_W48_HP40',
        params: { minPeriod: 8, maxPeriod: 50, windowLength: 48, highpassPeriod: 40 },
        expectedHeatmap: testdata.expectedP8_Q50_W48_HP40_HEATMAP,
        expectedMin: testdata.expectedP8_Q50_W48_HP40_MIN,
        expectedMax: testdata.expectedP8_Q50_W48_HP40_MAX,
        expectedDC: testdata.expectedP8_Q50_W48_HP40_DOMINANT_CYCLE,
    },
    {
        name: 'P8_Q50_W64_HP40',
        params: { minPeriod: 8, maxPeriod: 50, windowLength: 64, highpassPeriod: 40 },
        expectedHeatmap: testdata.expectedP8_Q50_W64_HP40_HEATMAP,
        expectedMin: testdata.expectedP8_Q50_W64_HP40_MIN,
        expectedMax: testdata.expectedP8_Q50_W64_HP40_MAX,
        expectedDC: testdata.expectedP8_Q50_W64_HP40_DOMINANT_CYCLE,
    },
    {
        name: 'P8_Q50_W252_HP40',
        params: { minPeriod: 8, maxPeriod: 50, windowLength: 252, highpassPeriod: 40 },
        expectedHeatmap: testdata.expectedP8_Q50_W252_HP40_HEATMAP,
        expectedMin: testdata.expectedP8_Q50_W252_HP40_MIN,
        expectedMax: testdata.expectedP8_Q50_W252_HP40_MAX,
        expectedDC: testdata.expectedP8_Q50_W252_HP40_DOMINANT_CYCLE,
    },
    {
        name: 'P8_Q50_W50_HP20',
        params: { minPeriod: 8, maxPeriod: 50, windowLength: 50, highpassPeriod: 20 },
        expectedHeatmap: testdata.expectedP8_Q50_W50_HP20_HEATMAP,
        expectedMin: testdata.expectedP8_Q50_W50_HP20_MIN,
        expectedMax: testdata.expectedP8_Q50_W50_HP20_MAX,
        expectedDC: testdata.expectedP8_Q50_W50_HP20_DOMINANT_CYCLE,
    },
    {
        name: 'P8_Q50_W50_HP80',
        params: { minPeriod: 8, maxPeriod: 50, windowLength: 50, highpassPeriod: 80 },
        expectedHeatmap: testdata.expectedP8_Q50_W50_HP80_HEATMAP,
        expectedMin: testdata.expectedP8_Q50_W50_HP80_MIN,
        expectedMax: testdata.expectedP8_Q50_W50_HP80_MAX,
        expectedDC: testdata.expectedP8_Q50_W50_HP80_DOMINANT_CYCLE,
    },
    {
        name: 'P8_Q50_W50_HP40_5',
        params: { minPeriod: 8, maxPeriod: 50, windowLength: 50, highpassPeriod: 40.5 },
        expectedHeatmap: testdata.expectedP8_Q50_W50_HP40_5_HEATMAP,
        expectedMin: testdata.expectedP8_Q50_W50_HP40_5_MIN,
        expectedMax: testdata.expectedP8_Q50_W50_HP40_5_MAX,
        expectedDC: testdata.expectedP8_Q50_W50_HP40_5_DOMINANT_CYCLE,
    },
    {
        name: 'P10_Q30_W50_HP40',
        params: { minPeriod: 10, maxPeriod: 30, windowLength: 50, highpassPeriod: 40 },
        expectedHeatmap: testdata.expectedP10_Q30_W50_HP40_HEATMAP,
        expectedMin: testdata.expectedP10_Q30_W50_HP40_MIN,
        expectedMax: testdata.expectedP10_Q30_W50_HP40_MAX,
        expectedDC: testdata.expectedP10_Q30_W50_HP40_DOMINANT_CYCLE,
    },
    {
        name: 'P8_Q80_W50_HP40',
        params: { minPeriod: 8, maxPeriod: 80, windowLength: 50, highpassPeriod: 40 },
        expectedHeatmap: testdata.expectedP8_Q80_W50_HP40_HEATMAP,
        expectedMin: testdata.expectedP8_Q80_W50_HP40_MIN,
        expectedMax: testdata.expectedP8_Q80_W50_HP40_MAX,
        expectedDC: testdata.expectedP8_Q80_W50_HP40_DOMINANT_CYCLE,
    },
    {
        name: 'P20_Q20_W50_HP40',
        params: { minPeriod: 20, maxPeriod: 20, windowLength: 50, highpassPeriod: 40 },
        expectedHeatmap: testdata.expectedP20_Q20_W50_HP40_HEATMAP,
        expectedMin: testdata.expectedP20_Q20_W50_HP40_MIN,
        expectedMax: testdata.expectedP20_Q20_W50_HP40_MAX,
        expectedDC: testdata.expectedP20_Q20_W50_HP40_DOMINANT_CYCLE,
    },
];

function assertClose(
    actual: number,
    expected: number,
    label: string,
    tolerance: number = FIXTURE_TOLERANCE,
): void {
    if (Number.isNaN(expected)) {
        if (!Number.isNaN(actual)) {
            fail(`${label}: expected NaN, got ${actual}`);
        }
        return;
    }
    if (!Number.isFinite(actual)) {
        fail(`${label}: expected ${expected}, got ${actual}`);
        return;
    }
    const bound = Math.max(tolerance, Math.abs(expected) * tolerance);
    if (Math.abs(actual - expected) > bound) {
        fail(`${label}: expected ${expected}, got ${actual} (diff ${Math.abs(actual - expected)} > ${bound})`);
    }
}

function assertSameOutput(actual: SpectrumOutput, expected: SpectrumOutput, label: string): void {
    expect(actual.heatmap.time.getTime()).toBe(expected.heatmap.time.getTime());
    expect(actual.heatmap.parameterFirst).toBe(expected.heatmap.parameterFirst);
    expect(actual.heatmap.parameterLast).toBe(expected.heatmap.parameterLast);
    expect(actual.heatmap.parameterResolution).toBe(expected.heatmap.parameterResolution);
    assertClose(actual.heatmap.valueMin, expected.heatmap.valueMin, `${label}.min`, 0);
    assertClose(actual.heatmap.valueMax, expected.heatmap.valueMax, `${label}.max`, 0);
    expect(actual.heatmap.values).toEqual(expected.heatmap.values);
    assertClose(actual.dominantCycle, expected.dominantCycle, `${label}.dc`, 0);
}

describe('TransformedDiscreteFourierTransformSpectrum', () => {
    it('matches all 12 pre-generated reference fixture cases', () => {
        const baseTime = new Date(Date.UTC(2026, 0, 1));
        for (const rc of REFERENCE_CASES) {
            const indicator = TransformedDiscreteFourierTransformSpectrum.fromParams(rc.params);
            const minPeriod = rc.params.minPeriod ?? 8;
            const maxPeriod = rc.params.maxPeriod ?? 50;
            const channelCount = maxPeriod - minPeriod + 1;

            for (let barIndex = 0; barIndex < testdata.testInput.length; barIndex++) {
                const sampleTime = new Date(baseTime.getTime() + barIndex * 86400000);
                const result = indicator.update(testdata.testInput[barIndex], sampleTime);
                expect(result.heatmap.time.getTime()).toBe(sampleTime.getTime());
                expect(result.heatmap.parameterFirst).toBe(minPeriod);
                expect(result.heatmap.parameterLast).toBe(maxPeriod);
                expect(result.heatmap.parameterResolution).toBe(1.0);

                const expectedRow = rc.expectedHeatmap[barIndex];
                if (Number.isNaN(rc.expectedDC[barIndex])) {
                    expect(indicator.isPrimed()).toBeFalse();
                    expect(result.heatmap.isEmpty()).toBeTrue();
                    expect(Number.isNaN(result.heatmap.valueMin)).toBeTrue();
                    expect(Number.isNaN(result.heatmap.valueMax)).toBeTrue();
                    expect(Number.isNaN(result.dominantCycle)).toBeTrue();
                    for (const cell of expectedRow) {
                        expect(Number.isNaN(cell)).toBeTrue();
                    }
                } else {
                    expect(indicator.isPrimed()).toBeTrue();
                    expect(result.heatmap.isEmpty()).toBeFalse();
                    expect(result.heatmap.values.length).toBe(channelCount);
                    assertClose(result.heatmap.valueMin, rc.expectedMin[barIndex], `${rc.name}[${barIndex}].min`);
                    assertClose(result.heatmap.valueMax, rc.expectedMax[barIndex], `${rc.name}[${barIndex}].max`);
                    assertClose(result.dominantCycle, rc.expectedDC[barIndex], `${rc.name}[${barIndex}].dc`);
                    for (let channelIndex = 0; channelIndex < channelCount; channelIndex++) {
                        assertClose(
                            result.heatmap.values[channelIndex],
                            expectedRow[channelIndex],
                            `${rc.name}[${barIndex}].cell[${channelIndex}]`,
                        );
                    }
                }
            }
        }
    });

    it('preserves state atomically on NaN, invalid, and overflowing inputs and supports reset', () => {
        const stream = TransformedDiscreteFourierTransformSpectrum.default();
        const control = TransformedDiscreteFourierTransformSpectrum.default();
        const baseTime = new Date(Date.UTC(2026, 0, 1));

        for (let barIndex = 0; barIndex < 80; barIndex++) {
            const sampleTime = new Date(baseTime.getTime() + barIndex * 86400000);
            stream.update(testdata.testInput[barIndex], sampleTime);
            control.update(testdata.testInput[barIndex], sampleTime);
        }

        const beforeState = stream.state;
        const beforeSamples = stream.samplesSeen;
        const beforeOutput = stream.output;

        const missingTime = new Date(Date.UTC(2027, 0, 1));
        const missing = stream.update(Number.NaN, missingTime);
        expect(missing.heatmap.isEmpty()).toBeTrue();
        expect(missing.heatmap.time.getTime()).toBe(missingTime.getTime());
        expect(Number.isNaN(missing.dominantCycle)).toBeTrue();

        for (const invalid of [Number.POSITIVE_INFINITY, Number.NEGATIVE_INFINITY, 1e308, true, null, '1']) {
            expect(() => stream.update(invalid as never, baseTime)).toThrow();
            expect(stream.samplesSeen).toBe(beforeSamples);
            expect(stream.state).toEqual(beforeState);
            expect(stream.output).toBe(beforeOutput);
        }

        const t80 = new Date(baseTime.getTime() + 80 * 86400000);
        assertSameOutput(stream.update(testdata.testInput[80], t80), control.update(testdata.testInput[80], t80), 'bar80');

        const t81 = new Date(baseTime.getTime() + 81 * 86400000);
        const snapshot = stream.update(testdata.testInput[81], t81);
        control.update(testdata.testInput[81], t81);
        snapshot.heatmap.values[0] = -999.0;
        const t82 = new Date(baseTime.getTime() + 82 * 86400000);
        assertSameOutput(stream.update(testdata.testInput[82], t82), control.update(testdata.testInput[82], t82), 'bar82');

        stream.reset();
        const fresh = TransformedDiscreteFourierTransformSpectrum.default();
        for (let barIndex = 0; barIndex < testdata.testInput.length; barIndex++) {
            const sampleTime = new Date(baseTime.getTime() + barIndex * 86400000);
            assertSameOutput(
                stream.update(testdata.testInput[barIndex], sampleTime),
                fresh.update(testdata.testInput[barIndex], sampleTime),
                `reset[${barIndex}]`,
            );
        }
    });

    it('validates parameters, mnemonics, metadata, descriptor, entities, and factory', () => {
        const mnemonicCases: Array<[TransformedDiscreteFourierTransformSpectrumParams, string, string]> = [
            [defaultParams(), 'tdfts(8, 50, 50, 40)', 'tdfts-dc(8, 50, 50, 40)'],
            [
                { minPeriod: 10, maxPeriod: 30, windowLength: 48, highpassPeriod: 40.5, barComponent: BarComponent.Median },
                'tdfts(10, 30, 48, 40.5, hl/2)',
                'tdfts-dc(10, 30, 48, 40.5, hl/2)',
            ],
            [
                { quoteComponent: QuoteComponent.Ask },
                'tdfts(8, 50, 50, 40, a)',
                'tdfts-dc(8, 50, 50, 40, a)',
            ],
            [
                { tradeComponent: TradeComponent.Volume },
                'tdfts(8, 50, 50, 40, v)',
                'tdfts-dc(8, 50, 50, 40, v)',
            ],
            [
                { minPeriod: 10, maxPeriod: 30, windowLength: 48, highpassPeriod: 40.5, barComponent: BarComponent.Median, quoteComponent: QuoteComponent.Ask },
                'tdfts(10, 30, 48, 40.5, hl/2, a)',
                'tdfts-dc(10, 30, 48, 40.5, hl/2, a)',
            ],
            [
                { quoteComponent: QuoteComponent.Bid, tradeComponent: TradeComponent.Volume },
                'tdfts(8, 50, 50, 40, b, v)',
                'tdfts-dc(8, 50, 50, 40, b, v)',
            ],
        ];

        for (const [params, wantMain, wantDC] of mnemonicCases) {
            expect(transformedDiscreteFourierTransformSpectrumMnemonic(params)).toBe(wantMain);
            expect(transformedDiscreteFourierTransformSpectrumDominantCycleMnemonic(params)).toBe(wantDC);
            const inst = new TransformedDiscreteFourierTransformSpectrum(params);
            const meta = inst.metadata();
            expect(meta.identifier).toBe(IndicatorIdentifier.TransformedDiscreteFourierTransformSpectrum);
            expect(meta.outputs.length).toBe(2);
            expect(meta.outputs[TransformedDiscreteFourierTransformSpectrumOutput.Value].shape).toBe(Shape.Heatmap);
            expect(meta.outputs[TransformedDiscreteFourierTransformSpectrumOutput.Value].mnemonic).toBe(wantMain);
            expect(meta.outputs[TransformedDiscreteFourierTransformSpectrumOutput.DominantCycle].shape).toBe(Shape.Scalar);
            expect(meta.outputs[TransformedDiscreteFourierTransformSpectrumOutput.DominantCycle].mnemonic).toBe(wantDC);
        }

        const desc = descriptorOf(IndicatorIdentifier.TransformedDiscreteFourierTransformSpectrum);
        expect(desc).toBeDefined();
        expect(desc!.family).toBe('John Ehlers');
        expect(desc!.adaptivity).toBe(Adaptivity.Static);
        expect(desc!.inputRequirement).toBe(InputRequirement.ScalarInput);
        expect(desc!.volumeUsage).toBe(VolumeUsage.NoVolume);
        expect(desc!.outputs.length).toBe(2);
        expect(desc!.outputs[0].role).toBe(Role.Spectrum);
        expect(desc!.outputs[0].pane).toBe(Pane.Own);
        expect(desc!.outputs[1].role).toBe(Role.CyclePeriod);
        expect(desc!.outputs[1].pane).toBe(Pane.Own);

        const baseTime = new Date(Date.UTC(2026, 0, 1));
        const barInd = TransformedDiscreteFourierTransformSpectrum.default();
        const scalarInd = TransformedDiscreteFourierTransformSpectrum.default();
        const quoteInd = TransformedDiscreteFourierTransformSpectrum.default();
        const tradeInd = TransformedDiscreteFourierTransformSpectrum.default();
        const openInd = new TransformedDiscreteFourierTransformSpectrum({ barComponent: BarComponent.Open });
        const openRef = TransformedDiscreteFourierTransformSpectrum.default();
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
            quote.bidPrice = closePrice - 0.5;
            quote.askPrice = closePrice + 0.5;
            quote.bidSize = 10;
            quote.askSize = 10;
            const trade = new Trade();
            trade.time = sampleTime;
            trade.price = closePrice;
            trade.volume = 10;

            const outBar = barInd.updateBar(bar);
            const outScalar = scalarInd.updateScalar(scalar);
            const outQuote = quoteInd.updateQuote(quote);
            const outTrade = tradeInd.updateTrade(trade);
            expect((outBar[0] as Heatmap).values).toEqual((outScalar[0] as Heatmap).values);
            assertClose((outBar[1] as Scalar).value, (outScalar[1] as Scalar).value, `bar_vs_scalar[${barIndex}]`);
            assertClose((outBar[1] as Scalar).value, (outQuote[1] as Scalar).value, `bar_vs_quote[${barIndex}]`);
            assertClose((outBar[1] as Scalar).value, (outTrade[1] as Scalar).value, `bar_vs_trade[${barIndex}]`);

            const openScalar = new Scalar();
            openScalar.time = sampleTime;
            openScalar.value = openPrice;
            const outOpenBar = openInd.updateBar(bar);
            const outOpenRef = openRef.updateScalar(openScalar);
            expect((outOpenBar[0] as Heatmap).values).toEqual((outOpenRef[0] as Heatmap).values);
            assertClose((outOpenBar[1] as Scalar).value, (outOpenRef[1] as Scalar).value, `open_bar[${barIndex}]`, 0);
        }

        const factoryDefault = createIndicator(IndicatorIdentifier.TransformedDiscreteFourierTransformSpectrum);
        expect(factoryDefault.metadata().mnemonic).toBe('tdfts(8, 50, 50, 40)');
        const factoryCustom = createIndicator(IndicatorIdentifier.TransformedDiscreteFourierTransformSpectrum, {
            minPeriod: 10,
            maxPeriod: 30,
            windowLength: 48,
            highpassPeriod: 40.5,
            barComponent: BarComponent.Median,
        });
        expect(factoryCustom.metadata().mnemonic).toBe('tdfts(10, 30, 48, 40.5, hl/2)');

        for (const badParams of [
            { minPeriod: 7 },
            { minPeriod: 8.5 },
            { minPeriod: null },
            { minPeriod: 20, maxPeriod: 19 },
            { windowLength: 0 },
            { windowLength: 3.5 },
            { highpassPeriod: 4.0 },
            { highpassPeriod: Number.NaN },
            { highpassPeriod: Number.POSITIVE_INFINITY },
            { unknownKey: 1 },
        ]) {
            expect(() => new TransformedDiscreteFourierTransformSpectrum(badParams as never)).toThrow();
            expect(() =>
                createIndicator(IndicatorIdentifier.TransformedDiscreteFourierTransformSpectrum, badParams as never),
            ).toThrow();
        }
    });
});
