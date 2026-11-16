import {} from 'jasmine';

import { Bar } from '../../../entities/bar';
import { BarComponent, DefaultBarComponent } from '../../../entities/bar-component';
import { Quote } from '../../../entities/quote';
import { DefaultQuoteComponent, QuoteComponent } from '../../../entities/quote-component';
import { Scalar } from '../../../entities/scalar';
import { Trade } from '../../../entities/trade';
import { DefaultTradeComponent, TradeComponent } from '../../../entities/trade-component';
import { Adaptivity } from '../../core/adaptivity';
import { componentTripleMnemonic } from '../../core/component-triple-mnemonic';
import { descriptorOf } from '../../core/descriptors';
import { IndicatorIdentifier } from '../../core/indicator-identifier';
import { Heatmap } from '../../core/outputs/heatmap';
import { Shape } from '../../core/outputs/shape/shape';
import { createIndicator } from '../../factory/factory';
import {
    DiscreteFourierTransformSpectrum,
    DiscreteFourierTransformSpectrumParams,
    ResolvedDiscreteFourierTransformSpectrumParams,
    SpectrumOutput,
    SpectrumState,
    defaultParams,
    discreteFourierTransformSpectrumDominantCycleMnemonic,
    discreteFourierTransformSpectrumMnemonic,
} from './discrete-fourier-transform-spectrum';
import { DiscreteFourierTransformSpectrumOutput } from './output';
import * as testdata from './testdata';

const REFERENCE_TOLERANCE = 1e-13;
const ORACLE_TOLERANCE = 2e-11;

function assertClose(
    actual: number,
    expected: number,
    label: string,
    tolerance: number = REFERENCE_TOLERANCE,
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

function assertSameState(actual: SpectrumState, expected: SpectrumState, label: string): void {
    assertClose(actual.highPass, expected.highPass, `${label}.hp`);
    assertClose(actual.filtered, expected.filtered, `${label}.filt`);
    assertClose(actual.normalizationMaximum, expected.normalizationMaximum, `${label}.max`);
    assertClose(actual.internalDominantCycle, expected.internalDominantCycle, `${label}.idc`);
    expect(actual.rawPower.length).toBe(expected.rawPower.length);
    for (let index = 0; index < actual.rawPower.length; index++) {
        assertClose(actual.rawPower[index], expected.rawPower[index], `${label}.raw[${index}]`);
    }
}

function assertSameOutput(actual: SpectrumOutput, expected: SpectrumOutput, label: string): void {
    expect(actual.heatmap.time.getTime()).toBe(expected.heatmap.time.getTime());
    expect(actual.heatmap.parameterFirst).toBe(expected.heatmap.parameterFirst);
    expect(actual.heatmap.parameterLast).toBe(expected.heatmap.parameterLast);
    expect(actual.heatmap.parameterResolution).toBe(expected.heatmap.parameterResolution);
    assertClose(actual.heatmap.valueMin, expected.heatmap.valueMin, `${label}.min`);
    assertClose(actual.heatmap.valueMax, expected.heatmap.valueMax, `${label}.max`);
    expect(actual.heatmap.values.length).toBe(expected.heatmap.values.length);
    for (let index = 0; index < actual.heatmap.values.length; index++) {
        assertClose(actual.heatmap.values[index], expected.heatmap.values[index], `${label}.values[${index}]`);
    }
    assertClose(actual.dominantCycle, expected.dominantCycle, `${label}.dc`);
}

class ListingOracle {
    private readonly prices: number[] = [];
    private readonly hp: number[] = [];
    private readonly filt: number[] = [];
    private maxPower = 0.0;
    private dominant = 0.0;
    private identified = false;

    public constructor(private readonly params: ResolvedDiscreteFourierTransformSpectrumParams) {}

    private previous(series: readonly number[], lag: number): number {
        return series.length > lag ? series[series.length - 1 - lag] : 0.0;
    }

    public update(price: number): { values: number[]; dc: number; hp: number; filt: number } {
        const params = this.params;
        const theta = (Math.SQRT2 * Math.PI) / params.highpassPeriod;
        const alpha = (Math.cos(theta) + Math.sin(theta) - 1.0) / Math.cos(theta);
        const high =
            (1.0 - alpha / 2.0) *
                (1.0 - alpha / 2.0) *
                (price - 2.0 * this.previous(this.prices, 0) + this.previous(this.prices, 1)) +
            2.0 * (1.0 - alpha) * this.previous(this.hp, 0) -
            (1.0 - alpha) * (1.0 - alpha) * this.previous(this.hp, 1);
        const a1 = Math.exp((-Math.SQRT2 * Math.PI) / params.smoothingPeriod);
        const c2 = 2.0 * a1 * Math.cos((Math.SQRT2 * Math.PI) / params.smoothingPeriod);
        const c3 = -a1 * a1;
        const c1 = 1.0 - c2 - c3;
        const filtered =
            (c1 * (high + this.previous(this.hp, 0))) / 2.0 +
            c2 * this.previous(this.filt, 0) +
            c3 * this.previous(this.filt, 1);
        this.prices.push(price);
        this.hp.push(high);
        this.filt.push(filtered);

        const cells = params.maxPeriod - params.minPeriod + 1;
        const powers = new Array<number>(cells);
        for (let index = 0; index < cells; index++) {
            const period = params.minPeriod + index;
            const compensation = params.spectralDilationCompensation ? period : 1.0;
            let cosine = 0.0;
            let sine = 0.0;
            for (let n = 0; n < params.windowLength; n++) {
                const value = this.previous(this.filt, n);
                const angle = (2.0 * Math.PI * n) / period;
                cosine += (value * Math.cos(angle)) / compensation;
                sine += (value * Math.sin(angle)) / compensation;
            }
            powers[index] = cosine * cosine + sine * sine;
        }

        this.maxPower *= 0.995;
        for (const power of powers) {
            if (power > this.maxPower) {
                this.maxPower = power;
            }
        }
        for (let index = 0; index < powers.length; index++) {
            powers[index] = this.maxPower !== 0.0 ? powers[index] / this.maxPower : 0.0;
        }

        let sp = 0.0;
        let spx = 0.0;
        for (let index = 0; index < powers.length; index++) {
            const period = params.minPeriod + index;
            if (powers[index] >= 0.5) {
                spx += period * powers[index];
                sp += powers[index];
            }
        }
        if (sp !== 0.0) {
            this.dominant = spx / sp;
            this.identified = true;
        }
        const values = powers.map((power) =>
            params.enhanceResolution ? power * power * power : power,
        );
        return {
            values,
            dc: this.identified ? this.dominant : Number.NaN,
            hp: high,
            filt: filtered,
        };
    }
}

interface ReferenceCase {
    name: string;
    params: DiscreteFourierTransformSpectrumParams;
    heatmap: number[][];
    min: number[];
    max: number[];
    dc: number[];
}

const REFERENCE_CASES: ReferenceCase[] = [
    {
        name: 'P010_P148_HP48_LP10_SDC1_E0_W48',
        params: { minPeriod: 10, maxPeriod: 48, windowLength: 48, highpassPeriod: 48, smoothingPeriod: 10, spectralDilationCompensation: true, enhanceResolution: false },
        heatmap: testdata.expectedP010_P148_HP48_LP10_SDC1_E0_W48_HEATMAP,
        min: testdata.expectedP010_P148_HP48_LP10_SDC1_E0_W48_MIN,
        max: testdata.expectedP010_P148_HP48_LP10_SDC1_E0_W48_MAX,
        dc: testdata.expectedP010_P148_HP48_LP10_SDC1_E0_W48_DOMINANT_CYCLE,
    },
    {
        name: 'P010_P148_HP48_LP10_SDC0_E0_W48',
        params: { minPeriod: 10, maxPeriod: 48, windowLength: 48, highpassPeriod: 48, smoothingPeriod: 10, spectralDilationCompensation: false, enhanceResolution: false },
        heatmap: testdata.expectedP010_P148_HP48_LP10_SDC0_E0_W48_HEATMAP,
        min: testdata.expectedP010_P148_HP48_LP10_SDC0_E0_W48_MIN,
        max: testdata.expectedP010_P148_HP48_LP10_SDC0_E0_W48_MAX,
        dc: testdata.expectedP010_P148_HP48_LP10_SDC0_E0_W48_DOMINANT_CYCLE,
    },
    {
        name: 'P010_P148_HP48_LP10_SDC1_E1_W48',
        params: { minPeriod: 10, maxPeriod: 48, windowLength: 48, highpassPeriod: 48, smoothingPeriod: 10, spectralDilationCompensation: true, enhanceResolution: true },
        heatmap: testdata.expectedP010_P148_HP48_LP10_SDC1_E1_W48_HEATMAP,
        min: testdata.expectedP010_P148_HP48_LP10_SDC1_E1_W48_MIN,
        max: testdata.expectedP010_P148_HP48_LP10_SDC1_E1_W48_MAX,
        dc: testdata.expectedP010_P148_HP48_LP10_SDC1_E1_W48_DOMINANT_CYCLE,
    },
    {
        name: 'P010_P148_HP48_LP10_SDC0_E1_W48',
        params: { minPeriod: 10, maxPeriod: 48, windowLength: 48, highpassPeriod: 48, smoothingPeriod: 10, spectralDilationCompensation: false, enhanceResolution: true },
        heatmap: testdata.expectedP010_P148_HP48_LP10_SDC0_E1_W48_HEATMAP,
        min: testdata.expectedP010_P148_HP48_LP10_SDC0_E1_W48_MIN,
        max: testdata.expectedP010_P148_HP48_LP10_SDC0_E1_W48_MAX,
        dc: testdata.expectedP010_P148_HP48_LP10_SDC0_E1_W48_DOMINANT_CYCLE,
    },
    {
        name: 'P010_P148_HP48_LP10_SDC1_E0_W24',
        params: { minPeriod: 10, maxPeriod: 48, windowLength: 24, highpassPeriod: 48, smoothingPeriod: 10, spectralDilationCompensation: true, enhanceResolution: false },
        heatmap: testdata.expectedP010_P148_HP48_LP10_SDC1_E0_W24_HEATMAP,
        min: testdata.expectedP010_P148_HP48_LP10_SDC1_E0_W24_MIN,
        max: testdata.expectedP010_P148_HP48_LP10_SDC1_E0_W24_MAX,
        dc: testdata.expectedP010_P148_HP48_LP10_SDC1_E0_W24_DOMINANT_CYCLE,
    },
    {
        name: 'P010_P148_HP48_LP10_SDC1_E0_W96',
        params: { minPeriod: 10, maxPeriod: 48, windowLength: 96, highpassPeriod: 48, smoothingPeriod: 10, spectralDilationCompensation: true, enhanceResolution: false },
        heatmap: testdata.expectedP010_P148_HP48_LP10_SDC1_E0_W96_HEATMAP,
        min: testdata.expectedP010_P148_HP48_LP10_SDC1_E0_W96_MIN,
        max: testdata.expectedP010_P148_HP48_LP10_SDC1_E0_W96_MAX,
        dc: testdata.expectedP010_P148_HP48_LP10_SDC1_E0_W96_DOMINANT_CYCLE,
    },
    {
        name: 'P010_P148_HP48_LP10_SDC1_E0_W252',
        params: { minPeriod: 10, maxPeriod: 48, windowLength: 252, highpassPeriod: 48, smoothingPeriod: 10, spectralDilationCompensation: true, enhanceResolution: false },
        heatmap: testdata.expectedP010_P148_HP48_LP10_SDC1_E0_W252_HEATMAP,
        min: testdata.expectedP010_P148_HP48_LP10_SDC1_E0_W252_MIN,
        max: testdata.expectedP010_P148_HP48_LP10_SDC1_E0_W252_MAX,
        dc: testdata.expectedP010_P148_HP48_LP10_SDC1_E0_W252_DOMINANT_CYCLE,
    },
    {
        name: 'P010_P148_HP48_LP10_5_SDC1_E0_W48',
        params: { minPeriod: 10, maxPeriod: 48, windowLength: 48, highpassPeriod: 48, smoothingPeriod: 10.5, spectralDilationCompensation: true, enhanceResolution: false },
        heatmap: testdata.expectedP010_P148_HP48_LP10_5_SDC1_E0_W48_HEATMAP,
        min: testdata.expectedP010_P148_HP48_LP10_5_SDC1_E0_W48_MIN,
        max: testdata.expectedP010_P148_HP48_LP10_5_SDC1_E0_W48_MAX,
        dc: testdata.expectedP010_P148_HP48_LP10_5_SDC1_E0_W48_DOMINANT_CYCLE,
    },
    {
        name: 'P010_P148_HP80_LP20_SDC1_E0_W48',
        params: { minPeriod: 10, maxPeriod: 48, windowLength: 48, highpassPeriod: 80, smoothingPeriod: 20, spectralDilationCompensation: true, enhanceResolution: false },
        heatmap: testdata.expectedP010_P148_HP80_LP20_SDC1_E0_W48_HEATMAP,
        min: testdata.expectedP010_P148_HP80_LP20_SDC1_E0_W48_MIN,
        max: testdata.expectedP010_P148_HP80_LP20_SDC1_E0_W48_MAX,
        dc: testdata.expectedP010_P148_HP80_LP20_SDC1_E0_W48_DOMINANT_CYCLE,
    },
    {
        name: 'P06_P124_HP24_LP6_SDC1_E0_W24',
        params: { minPeriod: 6, maxPeriod: 24, windowLength: 24, highpassPeriod: 24, smoothingPeriod: 6, spectralDilationCompensation: true, enhanceResolution: false },
        heatmap: testdata.expectedP06_P124_HP24_LP6_SDC1_E0_W24_HEATMAP,
        min: testdata.expectedP06_P124_HP24_LP6_SDC1_E0_W24_MIN,
        max: testdata.expectedP06_P124_HP24_LP6_SDC1_E0_W24_MAX,
        dc: testdata.expectedP06_P124_HP24_LP6_SDC1_E0_W24_DOMINANT_CYCLE,
    },
    {
        name: 'P020_P120_HP48_LP10_SDC1_E0_W48',
        params: { minPeriod: 20, maxPeriod: 20, windowLength: 48, highpassPeriod: 48, smoothingPeriod: 10, spectralDilationCompensation: true, enhanceResolution: false },
        heatmap: testdata.expectedP020_P120_HP48_LP10_SDC1_E0_W48_HEATMAP,
        min: testdata.expectedP020_P120_HP48_LP10_SDC1_E0_W48_MIN,
        max: testdata.expectedP020_P120_HP48_LP10_SDC1_E0_W48_MAX,
        dc: testdata.expectedP020_P120_HP48_LP10_SDC1_E0_W48_DOMINANT_CYCLE,
    },
    {
        name: 'P024_P160_HP60_LP12_SDC1_E0_W60',
        params: { minPeriod: 24, maxPeriod: 60, windowLength: 60, highpassPeriod: 60, smoothingPeriod: 12, spectralDilationCompensation: true, enhanceResolution: false },
        heatmap: testdata.expectedP024_P160_HP60_LP12_SDC1_E0_W60_HEATMAP,
        min: testdata.expectedP024_P160_HP60_LP12_SDC1_E0_W60_MIN,
        max: testdata.expectedP024_P160_HP60_LP12_SDC1_E0_W60_MAX,
        dc: testdata.expectedP024_P160_HP60_LP12_SDC1_E0_W60_DOMINANT_CYCLE,
    },
];

describe('DiscreteFourierTransformSpectrum', () => {
    const baseTime = new Date(Date.UTC(2020, 0, 1));

    it('should have correct output enum values', () => {
        expect(DiscreteFourierTransformSpectrumOutput.Value).toBe(0);
        expect(DiscreteFourierTransformSpectrumOutput.DominantCycle).toBe(1);
    });

    it('should match all 12 reference cases and Listing 9-1 oracle across 252 samples', () => {
        for (const rc of REFERENCE_CASES) {
            const indicator = new DiscreteFourierTransformSpectrum(rc.params);
            const oracle = new ListingOracle(indicator.parameters);
            const cells = indicator.maxPeriod - indicator.minPeriod + 1;

            for (let barIndex = 0; barIndex < testdata.testInput.length; barIndex++) {
                const barTime = new Date(baseTime.getTime() + barIndex * 86400000);
                const price = testdata.testInput[barIndex];
                const result = indicator.update(price, barTime);
                const oracleOut = oracle.update(price);

                assertClose(indicator.state.highPass, oracleOut.hp, `${rc.name}[${barIndex}].hp`, ORACLE_TOLERANCE);
                assertClose(indicator.state.filtered, oracleOut.filt, `${rc.name}[${barIndex}].filt`, ORACLE_TOLERANCE);

                expect(result.heatmap.time.getTime()).toBe(barTime.getTime());
                expect(result.heatmap.parameterFirst).toBe(indicator.minPeriod);
                expect(result.heatmap.parameterLast).toBe(indicator.maxPeriod);
                expect(result.heatmap.parameterResolution).toBe(1.0);

                if (barIndex < indicator.warmupBars) {
                    expect(indicator.isPrimed()).toBeFalse();
                    expect(result.heatmap.isEmpty()).toBeTrue();
                } else {
                    expect(indicator.isPrimed()).toBeTrue();
                    expect(result.heatmap.values.length).toBe(cells);
                    for (let binIndex = 0; binIndex < cells; binIndex++) {
                        assertClose(
                            result.heatmap.values[binIndex],
                            rc.heatmap[barIndex][binIndex],
                            `${rc.name}[${barIndex}].heatmap[${binIndex}]`,
                        );
                        assertClose(
                            result.heatmap.values[binIndex],
                            oracleOut.values[binIndex],
                            `${rc.name}[${barIndex}].oracle[${binIndex}]`,
                            ORACLE_TOLERANCE,
                        );
                    }
                    assertClose(
                        result.dominantCycle,
                        oracleOut.dc,
                        `${rc.name}[${barIndex}].oracle_dc`,
                        ORACLE_TOLERANCE,
                    );
                }

                assertClose(result.heatmap.valueMin, rc.min[barIndex], `${rc.name}[${barIndex}].min`);
                assertClose(result.heatmap.valueMax, rc.max[barIndex], `${rc.name}[${barIndex}].max`);
                assertClose(result.dominantCycle, rc.dc[barIndex], `${rc.name}[${barIndex}].dc`);
            }

            indicator.reset();
            expect(indicator.isPrimed()).toBeFalse();
            expect(indicator.samplesSeen).toBe(0);
        }
    });

    it('should preserve state on missing NaN, infinity, booleans, and arithmetic overflow', () => {
        const indicator = DiscreteFourierTransformSpectrum.default();
        const control = DiscreteFourierTransformSpectrum.default();
        for (let index = 0; index < 120; index++) {
            const sampleTime = new Date(baseTime.getTime() + index * 86400000);
            indicator.update(testdata.testInput[index], sampleTime);
            control.update(testdata.testInput[index], sampleTime);
        }
        const savedState = indicator.state;
        const savedOutput = indicator.output;
        const savedCount = indicator.samplesSeen;
        const missingTime = new Date(Date.UTC(2020, 5, 1));

        const missingResult = indicator.update(Number.NaN, missingTime);
        expect(missingResult.heatmap.isEmpty()).toBeTrue();
        expect(Number.isNaN(missingResult.dominantCycle)).toBeTrue();
        expect(missingResult.heatmap.time.getTime()).toBe(missingTime.getTime());
        assertSameState(indicator.state, savedState, 'after_nan');
        expect(indicator.output).toBe(savedOutput);
        expect(indicator.samplesSeen).toBe(savedCount);

        for (const badValue of [Number.POSITIVE_INFINITY, Number.NEGATIVE_INFINITY, true, 1e308]) {
            expect(() => indicator.update(badValue as never, missingTime)).toThrow();
            assertSameState(indicator.state, savedState, 'after_bad');
            expect(indicator.output).toBe(savedOutput);
            expect(indicator.samplesSeen).toBe(savedCount);
        }

        const nextTime = new Date(baseTime.getTime() + 120 * 86400000);
        assertSameOutput(
            indicator.update(testdata.testInput[120], nextTime),
            control.update(testdata.testInput[120], nextTime),
            'after_rejected',
        );
    });

    it('should isolate owned heatmap snapshots, replay on reset, handle zero input, compensation, enhancement, and sine peak', () => {
        const indicator = DiscreteFourierTransformSpectrum.default();
        let lastOut!: SpectrumOutput;
        for (let index = 0; index < testdata.testInput.length; index++) {
            lastOut = indicator.update(
                testdata.testInput[index],
                new Date(baseTime.getTime() + index * 86400000),
            );
        }
        const snapshot = lastOut.heatmap.values.slice();
        indicator.update(101.0, new Date(baseTime.getTime() + 252 * 86400000));
        expect(lastOut.heatmap.values).toEqual(snapshot);

        lastOut.heatmap.values[0] = -999.0;
        const control = DiscreteFourierTransformSpectrum.default();
        for (let index = 0; index < testdata.testInput.length; index++) {
            control.update(testdata.testInput[index], new Date(baseTime.getTime() + index * 86400000));
        }
        control.update(101.0, new Date(baseTime.getTime() + 252 * 86400000));
        assertSameOutput(
            indicator.update(102.0, new Date(baseTime.getTime() + 253 * 86400000)),
            control.update(102.0, new Date(baseTime.getTime() + 253 * 86400000)),
            'mutated_snapshot',
        );

        indicator.reset();
        const reference = DiscreteFourierTransformSpectrum.default();
        for (let index = 0; index < testdata.testInput.length; index++) {
            const sampleTime = new Date(baseTime.getTime() + index * 86400000);
            assertSameOutput(
                indicator.update(testdata.testInput[index], sampleTime),
                reference.update(testdata.testInput[index], sampleTime),
                `replay[${index}]`,
            );
        }

        const zeroIndicator = DiscreteFourierTransformSpectrum.default();
        let zeroOut = zeroIndicator.output;
        for (let step = 0; step < zeroIndicator.warmupBars + 4; step++) {
            zeroOut = zeroIndicator.update(0.0);
        }
        expect(zeroIndicator.isPrimed()).toBeTrue();
        expect(zeroOut.heatmap.values).toEqual(new Array<number>(39).fill(0.0));
        expect(Number.isNaN(zeroOut.dominantCycle)).toBeTrue();
        expect(zeroIndicator.state.normalizationMaximum).toBe(0.0);

        const compensated = DiscreteFourierTransformSpectrum.default();
        const plain = new DiscreteFourierTransformSpectrum({ spectralDilationCompensation: false });
        const enhanced = new DiscreteFourierTransformSpectrum({ enhanceResolution: true });
        for (let index = 0; index < testdata.testInput.length; index++) {
            const compOut = compensated.update(testdata.testInput[index]);
            plain.update(testdata.testInput[index]);
            const enhOut = enhanced.update(testdata.testInput[index]);
            assertSameState(compensated.state, enhanced.state, `enh_state[${index}]`);
            assertClose(compOut.dominantCycle, enhOut.dominantCycle, `enh_dc[${index}]`);
            if (!compOut.heatmap.isEmpty()) {
                for (let binIndex = 0; binIndex < compOut.heatmap.values.length; binIndex++) {
                    const baseVal = compOut.heatmap.values[binIndex];
                    assertClose(enhOut.heatmap.values[binIndex], baseVal * baseVal * baseVal, `enh_cube[${index}]`);
                }
            }
        }
        for (let binIndex = 0; binIndex < 39; binIndex++) {
            const period = 10 + binIndex;
            assertClose(
                compensated.state.rawPower[binIndex],
                plain.state.rawPower[binIndex] / (period * period),
                `sdc_ratio[${period}]`,
                ORACLE_TOLERANCE,
            );
        }

        const sineIndicator = new DiscreteFourierTransformSpectrum({ windowLength: 120 });
        let sineOut = sineIndicator.output;
        for (let step = 0; step < 1000; step++) {
            sineOut = sineIndicator.update(Math.sin((2.0 * Math.PI * step) / 20.0));
        }
        let peakIndex = 0;
        for (let index = 1; index < sineOut.heatmap.values.length; index++) {
            if (sineOut.heatmap.values[index] > sineOut.heatmap.values[peakIndex]) {
                peakIndex = index;
            }
        }
        const peakPeriod = 10 + peakIndex;
        expect(Math.abs(peakPeriod - 20)).toBeLessThanOrEqual(1);
        expect(sineOut.dominantCycle).toBeGreaterThan(17.0);
        expect(sineOut.dominantCycle).toBeLessThan(23.0);
    });

    it('should build metadata, mnemonics, entity updates, and reject invalid/legacy parameters', () => {
        const desc = descriptorOf(IndicatorIdentifier.DiscreteFourierTransformSpectrum);
        expect(desc).toBeDefined();
        expect(desc!.adaptivity).toBe(Adaptivity.Static);
        expect(desc!.outputs.length).toBe(2);

        const indicator = DiscreteFourierTransformSpectrum.fromParams(defaultParams());
        const meta = indicator.metadata();
        expect(meta.identifier).toBe(IndicatorIdentifier.DiscreteFourierTransformSpectrum);
        expect(meta.mnemonic).toBe('dftps(10, 48, 48, 48, 10)');
        expect(meta.description).toBe('Discrete Fourier transform spectrum dftps(10, 48, 48, 48, 10)');
        expect(meta.outputs.length).toBe(2);
        expect(meta.outputs[DiscreteFourierTransformSpectrumOutput.Value].shape).toBe(Shape.Heatmap);
        expect(meta.outputs[DiscreteFourierTransformSpectrumOutput.Value].mnemonic).toBe('dftps(10, 48, 48, 48, 10)');
        expect(meta.outputs[DiscreteFourierTransformSpectrumOutput.DominantCycle].shape).toBe(Shape.Scalar);
        expect(meta.outputs[DiscreteFourierTransformSpectrumOutput.DominantCycle].mnemonic).toBe('dftps-dc(10, 48, 48, 48, 10)');

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
        for (const [barComp, quoteComp, tradeComp] of combinations) {
            const params: DiscreteFourierTransformSpectrumParams = {
                barComponent: barComp,
                quoteComponent: quoteComp,
                tradeComponent: tradeComp,
            };
            const inst = new DiscreteFourierTransformSpectrum(params);
            const suffix = componentTripleMnemonic(
                barComp ?? DefaultBarComponent,
                quoteComp ?? DefaultQuoteComponent,
                tradeComp ?? DefaultTradeComponent,
            );
            const wantPrimary = `dftps(10, 48, 48, 48, 10${suffix})`;
            const wantDC = `dftps-dc(10, 48, 48, 48, 10${suffix})`;
            expect(inst.metadata().mnemonic).toBe(wantPrimary);
            expect(discreteFourierTransformSpectrumMnemonic(params)).toBe(wantPrimary);
            expect(discreteFourierTransformSpectrumDominantCycleMnemonic(params)).toBe(wantDC);
        }

        const barInd = DiscreteFourierTransformSpectrum.default();
        const scalarInd = DiscreteFourierTransformSpectrum.default();
        const quoteInd = DiscreteFourierTransformSpectrum.default();
        const tradeInd = DiscreteFourierTransformSpectrum.default();
        for (let index = 0; index < 60; index++) {
            const sampleTime = new Date(baseTime.getTime() + index * 86400000);
            const price = testdata.testInput[index];
            const bar = new Bar();
            bar.time = sampleTime;
            bar.open = testdata.testOpen[index];
            bar.high = testdata.testHigh[index];
            bar.low = testdata.testLow[index];
            bar.close = price;
            bar.volume = 1000;
            const scalar = new Scalar();
            scalar.time = sampleTime;
            scalar.value = price;
            const quote = new Quote();
            quote.time = sampleTime;
            quote.bidPrice = price - 0.5;
            quote.askPrice = price + 0.5;
            quote.bidSize = 10;
            quote.askSize = 10;
            const trade = new Trade();
            trade.time = sampleTime;
            trade.price = price;
            trade.volume = 10;

            const outBar = barInd.updateBar(bar);
            const outScalar = scalarInd.updateScalar(scalar);
            const outQuote = quoteInd.updateQuote(quote);
            const outTrade = tradeInd.updateTrade(trade);
            expect(outBar.length).toBe(2);
            expect((outBar[0] as Heatmap).values).toEqual((outScalar[0] as Heatmap).values);
            assertClose((outBar[1] as Scalar).value, (outScalar[1] as Scalar).value, `bar_vs_scalar[${index}]`);
            assertClose((outBar[1] as Scalar).value, (outQuote[1] as Scalar).value, `bar_vs_quote[${index}]`);
            assertClose((outBar[1] as Scalar).value, (outTrade[1] as Scalar).value, `bar_vs_trade[${index}]`);
        }

        const factoryDefault = createIndicator(IndicatorIdentifier.DiscreteFourierTransformSpectrum);
        expect(factoryDefault.metadata().mnemonic).toBe('dftps(10, 48, 48, 48, 10)');
        const factoryCustom = createIndicator(IndicatorIdentifier.DiscreteFourierTransformSpectrum, {
            minPeriod: 8,
            maxPeriod: 32,
            windowLength: 24,
            highpassPeriod: 40,
            smoothingPeriod: 10.5,
            spectralDilationCompensation: false,
            enhanceResolution: true,
            barComponent: BarComponent.Median,
        });
        expect(factoryCustom.metadata().mnemonic).toBe('dftps(8, 32, 24, 40, 10.5, no-sdc, enh, hl/2)');

        for (const badParams of [
            { minPeriod: 2 },
            { minPeriod: 0 },
            { minPeriod: 10.5 },
            { minPeriod: true },
            { minPeriod: 10, maxPeriod: 9 },
            { maxPeriod: true },
            { windowLength: 1 },
            { windowLength: 12.5 },
            { highpassPeriod: 4 },
            { highpassPeriod: Number.NaN },
            { highpassPeriod: Number.POSITIVE_INFINITY },
            { smoothingPeriod: 1 },
            { smoothingPeriod: false },
            { spectralDilationCompensation: 1 },
            { enhanceResolution: 'yes' },
            { length: 48 },
            { spectrumResolution: 1 },
            { disableSpectralDilationCompensation: true },
            { disableAutomaticGainControl: true },
            { automaticGainControlDecayFactor: 0.995 },
            { fixedNormalization: true },
            { unknownKey: 1 },
        ]) {
            expect(() => new DiscreteFourierTransformSpectrum(badParams as never)).toThrow();
            expect(() =>
                createIndicator(IndicatorIdentifier.DiscreteFourierTransformSpectrum, badParams),
            ).toThrow();
        }
    });
});
