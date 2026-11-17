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
    CombBandPassSpectrum,
    CombBandPassSpectrumParams,
    ResolvedCombBandPassSpectrumParams,
    SpectrumOutput,
    SpectrumState,
    combBandPassSpectrumDominantCycleMnemonic,
    combBandPassSpectrumMnemonic,
    defaultParams,
} from './comb-band-pass-spectrum';
import { CombBandPassSpectrumOutput } from './output';
import * as testdata from './testdata';

const FIXTURE_TOLERANCE = 1e-11;
const ORACLE_TOLERANCE = 2e-11;
const EXACT_TOLERANCE = 1e-13;

function assertClose(
    actual: number,
    expected: number,
    label: string,
    tolerance: number = EXACT_TOLERANCE,
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
    private readonly bp = new Map<number, number[]>();
    private maxPower = 0.0;
    private dominant = 0.0;
    private identified = false;

    public constructor(private readonly params: ResolvedCombBandPassSpectrumParams) {
        for (let period = params.minPeriod; period <= params.maxPeriod; period++) {
            this.bp.set(period, new Array<number>(params.maxPeriod + 1).fill(0.0));
        }
    }

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
            const row = this.bp.get(period)!;
            for (let m = params.maxPeriod; m > 1; m--) {
                row[m] = row[m - 1];
            }
            row[1] = 0.0;
            if (this.prices.length > 12) {
                const beta = Math.cos((2.0 * Math.PI) / period);
                const gamma = 1.0 / Math.cos((2.0 * Math.PI * params.bandwidth) / period);
                const pole = gamma - Math.sqrt(gamma * gamma - 1.0);
                row[1] =
                    0.5 * (1.0 - pole) * (filtered - this.previous(this.filt, 2)) +
                    beta * (1.0 + pole) * row[2] -
                    pole * row[3];
            }
            let power = 0.0;
            for (let m = 1; m <= period; m++) {
                power += (row[m] / compensation) * (row[m] / compensation);
            }
            powers[index] = power;
        }

        this.maxPower *= 0.995;
        for (const power of powers) {
            if (power > this.maxPower) {
                this.maxPower = power;
            }
        }
        for (let index = 0; index < cells; index++) {
            powers[index] = this.maxPower !== 0.0 ? powers[index] / this.maxPower : 0.0;
        }

        let sp = 0.0;
        let spx = 0.0;
        for (let index = 0; index < cells; index++) {
            const period = params.minPeriod + index;
            const power = powers[index];
            if (power >= 0.5) {
                spx += period * power;
                sp += power;
            }
        }
        if (sp !== 0.0) {
            this.dominant = spx / sp;
            this.identified = true;
        }
        const values = params.enhanceResolution
            ? powers.map((power) => power * power * power)
            : powers.slice();
        return {
            values,
            dc: this.identified ? this.dominant : Number.NaN,
            hp: high,
            filt: filtered,
        };
    }
}

interface ReferenceCase {
    readonly name: string;
    readonly params: CombBandPassSpectrumParams;
    readonly expectedHeatmap: number[][];
    readonly expectedMin: number[];
    readonly expectedMax: number[];
    readonly expectedDC: number[];
}

const REFERENCE_CASES: readonly ReferenceCase[] = [
    {
        name: 'P010_P148_HP48_LP10_SDC1_E0_B0_3',
        params: { minPeriod: 10, maxPeriod: 48, bandwidth: 0.3, highpassPeriod: 48, smoothingPeriod: 10, spectralDilationCompensation: true, enhanceResolution: false },
        expectedHeatmap: testdata.expectedP010_P148_HP48_LP10_SDC1_E0_B0_3_HEATMAP,
        expectedMin: testdata.expectedP010_P148_HP48_LP10_SDC1_E0_B0_3_MIN,
        expectedMax: testdata.expectedP010_P148_HP48_LP10_SDC1_E0_B0_3_MAX,
        expectedDC: testdata.expectedP010_P148_HP48_LP10_SDC1_E0_B0_3_DOMINANT_CYCLE,
    },
    {
        name: 'P010_P148_HP48_LP10_SDC0_E0_B0_3',
        params: { minPeriod: 10, maxPeriod: 48, bandwidth: 0.3, highpassPeriod: 48, smoothingPeriod: 10, spectralDilationCompensation: false, enhanceResolution: false },
        expectedHeatmap: testdata.expectedP010_P148_HP48_LP10_SDC0_E0_B0_3_HEATMAP,
        expectedMin: testdata.expectedP010_P148_HP48_LP10_SDC0_E0_B0_3_MIN,
        expectedMax: testdata.expectedP010_P148_HP48_LP10_SDC0_E0_B0_3_MAX,
        expectedDC: testdata.expectedP010_P148_HP48_LP10_SDC0_E0_B0_3_DOMINANT_CYCLE,
    },
    {
        name: 'P010_P148_HP48_LP10_SDC1_E1_B0_3',
        params: { minPeriod: 10, maxPeriod: 48, bandwidth: 0.3, highpassPeriod: 48, smoothingPeriod: 10, spectralDilationCompensation: true, enhanceResolution: true },
        expectedHeatmap: testdata.expectedP010_P148_HP48_LP10_SDC1_E1_B0_3_HEATMAP,
        expectedMin: testdata.expectedP010_P148_HP48_LP10_SDC1_E1_B0_3_MIN,
        expectedMax: testdata.expectedP010_P148_HP48_LP10_SDC1_E1_B0_3_MAX,
        expectedDC: testdata.expectedP010_P148_HP48_LP10_SDC1_E1_B0_3_DOMINANT_CYCLE,
    },
    {
        name: 'P010_P148_HP48_LP10_SDC0_E1_B0_3',
        params: { minPeriod: 10, maxPeriod: 48, bandwidth: 0.3, highpassPeriod: 48, smoothingPeriod: 10, spectralDilationCompensation: false, enhanceResolution: true },
        expectedHeatmap: testdata.expectedP010_P148_HP48_LP10_SDC0_E1_B0_3_HEATMAP,
        expectedMin: testdata.expectedP010_P148_HP48_LP10_SDC0_E1_B0_3_MIN,
        expectedMax: testdata.expectedP010_P148_HP48_LP10_SDC0_E1_B0_3_MAX,
        expectedDC: testdata.expectedP010_P148_HP48_LP10_SDC0_E1_B0_3_DOMINANT_CYCLE,
    },
    {
        name: 'P010_P148_HP48_LP10_SDC1_E0_B0_1',
        params: { minPeriod: 10, maxPeriod: 48, bandwidth: 0.1, highpassPeriod: 48, smoothingPeriod: 10, spectralDilationCompensation: true, enhanceResolution: false },
        expectedHeatmap: testdata.expectedP010_P148_HP48_LP10_SDC1_E0_B0_1_HEATMAP,
        expectedMin: testdata.expectedP010_P148_HP48_LP10_SDC1_E0_B0_1_MIN,
        expectedMax: testdata.expectedP010_P148_HP48_LP10_SDC1_E0_B0_1_MAX,
        expectedDC: testdata.expectedP010_P148_HP48_LP10_SDC1_E0_B0_1_DOMINANT_CYCLE,
    },
    {
        name: 'P010_P148_HP48_LP10_SDC1_E0_B0_7',
        params: { minPeriod: 10, maxPeriod: 48, bandwidth: 0.7, highpassPeriod: 48, smoothingPeriod: 10, spectralDilationCompensation: true, enhanceResolution: false },
        expectedHeatmap: testdata.expectedP010_P148_HP48_LP10_SDC1_E0_B0_7_HEATMAP,
        expectedMin: testdata.expectedP010_P148_HP48_LP10_SDC1_E0_B0_7_MIN,
        expectedMax: testdata.expectedP010_P148_HP48_LP10_SDC1_E0_B0_7_MAX,
        expectedDC: testdata.expectedP010_P148_HP48_LP10_SDC1_E0_B0_7_DOMINANT_CYCLE,
    },
    {
        name: 'P010_P148_HP48_LP10_5_SDC1_E0_B0_3',
        params: { minPeriod: 10, maxPeriod: 48, bandwidth: 0.3, highpassPeriod: 48, smoothingPeriod: 10.5, spectralDilationCompensation: true, enhanceResolution: false },
        expectedHeatmap: testdata.expectedP010_P148_HP48_LP10_5_SDC1_E0_B0_3_HEATMAP,
        expectedMin: testdata.expectedP010_P148_HP48_LP10_5_SDC1_E0_B0_3_MIN,
        expectedMax: testdata.expectedP010_P148_HP48_LP10_5_SDC1_E0_B0_3_MAX,
        expectedDC: testdata.expectedP010_P148_HP48_LP10_5_SDC1_E0_B0_3_DOMINANT_CYCLE,
    },
    {
        name: 'P010_P148_HP80_LP20_SDC1_E0_B0_3',
        params: { minPeriod: 10, maxPeriod: 48, bandwidth: 0.3, highpassPeriod: 80, smoothingPeriod: 20, spectralDilationCompensation: true, enhanceResolution: false },
        expectedHeatmap: testdata.expectedP010_P148_HP80_LP20_SDC1_E0_B0_3_HEATMAP,
        expectedMin: testdata.expectedP010_P148_HP80_LP20_SDC1_E0_B0_3_MIN,
        expectedMax: testdata.expectedP010_P148_HP80_LP20_SDC1_E0_B0_3_MAX,
        expectedDC: testdata.expectedP010_P148_HP80_LP20_SDC1_E0_B0_3_DOMINANT_CYCLE,
    },
    {
        name: 'P06_P124_HP24_LP6_SDC1_E0_B0_3',
        params: { minPeriod: 6, maxPeriod: 24, bandwidth: 0.3, highpassPeriod: 24, smoothingPeriod: 6, spectralDilationCompensation: true, enhanceResolution: false },
        expectedHeatmap: testdata.expectedP06_P124_HP24_LP6_SDC1_E0_B0_3_HEATMAP,
        expectedMin: testdata.expectedP06_P124_HP24_LP6_SDC1_E0_B0_3_MIN,
        expectedMax: testdata.expectedP06_P124_HP24_LP6_SDC1_E0_B0_3_MAX,
        expectedDC: testdata.expectedP06_P124_HP24_LP6_SDC1_E0_B0_3_DOMINANT_CYCLE,
    },
    {
        name: 'P020_P120_HP48_LP10_SDC1_E0_B0_3',
        params: { minPeriod: 20, maxPeriod: 20, bandwidth: 0.3, highpassPeriod: 48, smoothingPeriod: 10, spectralDilationCompensation: true, enhanceResolution: false },
        expectedHeatmap: testdata.expectedP020_P120_HP48_LP10_SDC1_E0_B0_3_HEATMAP,
        expectedMin: testdata.expectedP020_P120_HP48_LP10_SDC1_E0_B0_3_MIN,
        expectedMax: testdata.expectedP020_P120_HP48_LP10_SDC1_E0_B0_3_MAX,
        expectedDC: testdata.expectedP020_P120_HP48_LP10_SDC1_E0_B0_3_DOMINANT_CYCLE,
    },
    {
        name: 'P024_P160_HP60_LP12_SDC1_E0_B0_3',
        params: { minPeriod: 24, maxPeriod: 60, bandwidth: 0.3, highpassPeriod: 60, smoothingPeriod: 12, spectralDilationCompensation: true, enhanceResolution: false },
        expectedHeatmap: testdata.expectedP024_P160_HP60_LP12_SDC1_E0_B0_3_HEATMAP,
        expectedMin: testdata.expectedP024_P160_HP60_LP12_SDC1_E0_B0_3_MIN,
        expectedMax: testdata.expectedP024_P160_HP60_LP12_SDC1_E0_B0_3_MAX,
        expectedDC: testdata.expectedP024_P160_HP60_LP12_SDC1_E0_B0_3_DOMINANT_CYCLE,
    },
    {
        name: 'P010_P1240_HP240_LP10_SDC1_E0_B0_3',
        params: { minPeriod: 10, maxPeriod: 240, bandwidth: 0.3, highpassPeriod: 240, smoothingPeriod: 10, spectralDilationCompensation: true, enhanceResolution: false },
        expectedHeatmap: testdata.expectedP010_P1240_HP240_LP10_SDC1_E0_B0_3_HEATMAP,
        expectedMin: testdata.expectedP010_P1240_HP240_LP10_SDC1_E0_B0_3_MIN,
        expectedMax: testdata.expectedP010_P1240_HP240_LP10_SDC1_E0_B0_3_MAX,
        expectedDC: testdata.expectedP010_P1240_HP240_LP10_SDC1_E0_B0_3_DOMINANT_CYCLE,
    },
];

describe('CombBandPassSpectrum', () => {
    const baseTime = new Date(Date.UTC(2020, 0, 1));

    it('matches all 12 reference cases and independent Listing 10-1 oracle', () => {
        for (const rc of REFERENCE_CASES) {
            const indicator = new CombBandPassSpectrum(rc.params);
            const oracle = new ListingOracle(indicator.parameters);
            const cells = indicator.maxPeriod - indicator.minPeriod + 1;

            for (let barIndex = 0; barIndex < testdata.testInput.length; barIndex++) {
                const sampleTime = new Date(baseTime.getTime() + barIndex * 86400000);
                const result = indicator.update(testdata.testInput[barIndex], sampleTime);
                const oracleStep = oracle.update(testdata.testInput[barIndex]);

                assertClose(indicator.state.highPass, oracleStep.hp, `${rc.name}[${barIndex}].hp`, ORACLE_TOLERANCE);
                assertClose(indicator.state.filtered, oracleStep.filt, `${rc.name}[${barIndex}].filt`, ORACLE_TOLERANCE);
                expect(result.heatmap.time.getTime()).toBe(sampleTime.getTime());
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
                            rc.expectedHeatmap[barIndex][binIndex],
                            `${rc.name}[${barIndex}][${binIndex}]`,
                            FIXTURE_TOLERANCE,
                        );
                        assertClose(
                            result.heatmap.values[binIndex],
                            oracleStep.values[binIndex],
                            `${rc.name}[${barIndex}][${binIndex}].oracle`,
                            ORACLE_TOLERANCE,
                        );
                    }
                    assertClose(
                        result.dominantCycle,
                        oracleStep.dc,
                        `${rc.name}[${barIndex}].oracle_dc`,
                        ORACLE_TOLERANCE,
                    );
                }

                assertClose(result.heatmap.valueMin, rc.expectedMin[barIndex], `${rc.name}[${barIndex}].min`, FIXTURE_TOLERANCE);
                assertClose(result.heatmap.valueMax, rc.expectedMax[barIndex], `${rc.name}[${barIndex}].max`, FIXTURE_TOLERANCE);
                assertClose(result.dominantCycle, rc.expectedDC[barIndex], `${rc.name}[${barIndex}].dc`, FIXTURE_TOLERANCE);
            }

            indicator.reset();
            expect(indicator.isPrimed()).toBeFalse();
            expect(indicator.samplesSeen).toBe(0);
        }
    });

    it('preserves the 12-bar startup gate before exciting the comb bank', () => {
        const indicator = CombBandPassSpectrum.default();
        for (let index = 0; index < 12; index++) {
            indicator.update(testdata.testInput[index]);
            for (const row of indicator.bandPassHistory) {
                expect(row.every((value) => value === 0.0)).toBeTrue();
            }
        }
        indicator.update(testdata.testInput[12]);
        expect(indicator.bandPassHistory.some((row) => row[0] !== 0.0)).toBeTrue();
    });

    it('preserves state on NaN, rejects infinity/overflow atomically, and owns snapshots', () => {
        const indicator = CombBandPassSpectrum.default();
        const control = CombBandPassSpectrum.default();
        for (let index = 0; index < 120; index++) {
            const sampleTime = new Date(baseTime.getTime() + index * 86400000);
            indicator.update(testdata.testInput[index], sampleTime);
            control.update(testdata.testInput[index], sampleTime);
        }

        const savedState = indicator.state;
        const savedOutput = indicator.output;
        const savedCount = indicator.samplesSeen;

        const nanTime = new Date(Date.UTC(2025, 0, 1));
        const nanOut = indicator.update(Number.NaN, nanTime);
        expect(nanOut.heatmap.isEmpty()).toBeTrue();
        expect(nanOut.heatmap.time.getTime()).toBe(nanTime.getTime());
        expect(Number.isNaN(nanOut.dominantCycle)).toBeTrue();
        assertSameState(indicator.state, savedState, 'after_nan_state');
        assertSameOutput(indicator.output, savedOutput, 'after_nan_output');
        expect(indicator.samplesSeen).toBe(savedCount);

        for (const badSample of [Number.POSITIVE_INFINITY, Number.NEGATIVE_INFINITY, 1e308, true as unknown as number]) {
            expect(() => indicator.update(badSample, nanTime)).toThrow();
            assertSameState(indicator.state, savedState, 'after_bad_state');
            assertSameOutput(indicator.output, savedOutput, 'after_bad_output');
            expect(indicator.samplesSeen).toBe(savedCount);
        }

        for (let index = 120; index < 150; index++) {
            const sampleTime = new Date(baseTime.getTime() + index * 86400000);
            assertSameOutput(
                indicator.update(testdata.testInput[index], sampleTime),
                control.update(testdata.testInput[index], sampleTime),
                `resume[${index}]`,
            );
        }

        const snapshot = indicator.output;
        const secondOriginal = snapshot.heatmap.values[1];
        snapshot.heatmap.values[0] = -999;
        const historyCopy = indicator.bandPassHistory as number[][];
        historyCopy[0][0] = -777;
        expect(indicator.bandPassHistory[0][0]).not.toBe(-777);
        indicator.update(testdata.testInput[150]);
        expect(snapshot.heatmap.values[1]).toBe(secondOriginal);
    });

    it('handles zero input, single-bin, boundary, compensation, enhancement, and sine wave', () => {
        const zeroInd = CombBandPassSpectrum.default();
        let zeroOut!: SpectrumOutput;
        for (let step = 0; step < 70; step++) {
            zeroOut = zeroInd.update(0.0);
        }
        expect(zeroInd.isPrimed()).toBeTrue();
        expect(zeroOut.heatmap.valueMin).toBe(0.0);
        expect(zeroOut.heatmap.valueMax).toBe(0.0);
        expect(Number.isNaN(zeroOut.dominantCycle)).toBeTrue();

        const single = new CombBandPassSpectrum({ minPeriod: 20, maxPeriod: 20 });
        expect(single.warmupBars).toBe(31);
        for (const price of testdata.testInput) {
            const out = single.update(price);
            if (!Number.isNaN(out.dominantCycle)) {
                assertClose(out.dominantCycle, 20.0, 'single_dc');
            }
        }

        const boundary = new CombBandPassSpectrum({
            minPeriod: 10,
            maxPeriod: 240,
            highpassPeriod: 240,
            smoothingPeriod: 10,
        });
        expect(boundary.warmupBars).toBe(251);
        const boundaryOutputs = testdata.testInput.map((price) => boundary.update(price));
        expect(boundaryOutputs.slice(0, 251).every((item) => item.heatmap.isEmpty())).toBeTrue();
        expect(boundaryOutputs[251].heatmap.isEmpty()).toBeFalse();

        const compensated = CombBandPassSpectrum.default();
        const uncompensated = new CombBandPassSpectrum({ spectralDilationCompensation: false });
        const enhanced = new CombBandPassSpectrum({ enhanceResolution: true });
        for (let barIndex = 0; barIndex < testdata.testInput.length; barIndex++) {
            const outComp = compensated.update(testdata.testInput[barIndex]);
            uncompensated.update(testdata.testInput[barIndex]);
            const outEnh = enhanced.update(testdata.testInput[barIndex]);
            assertSameState(compensated.state, enhanced.state, `enh_state[${barIndex}]`);
            assertClose(outComp.dominantCycle, outEnh.dominantCycle, `enh_dc[${barIndex}]`);
            for (let binIndex = 0; binIndex < 39; binIndex++) {
                const period = 10 + binIndex;
                assertClose(
                    compensated.state.rawPower[binIndex],
                    uncompensated.state.rawPower[binIndex] / (period * period),
                    `sdc[${barIndex}][${binIndex}]`,
                );
            }
            if (!outComp.heatmap.isEmpty()) {
                for (let binIndex = 0; binIndex < 39; binIndex++) {
                    const plainValue = outComp.heatmap.values[binIndex];
                    assertClose(
                        outEnh.heatmap.values[binIndex],
                        plainValue * plainValue * plainValue,
                        `cube[${barIndex}][${binIndex}]`,
                    );
                }
            }
        }

        const sineInd = CombBandPassSpectrum.default();
        let sineOut!: SpectrumOutput;
        for (let step = 0; step < 1000; step++) {
            sineOut = sineInd.update(Math.sin((2.0 * Math.PI * step) / 20.0));
        }
        let peakBin = 0;
        for (let binIndex = 1; binIndex < sineOut.heatmap.values.length; binIndex++) {
            if (sineOut.heatmap.values[binIndex] > sineOut.heatmap.values[peakBin]) {
                peakBin = binIndex;
            }
        }
        expect(Math.abs(10 + peakBin - 20)).toBeLessThanOrEqual(2);
        expect(sineOut.dominantCycle).toBeGreaterThan(17.0);
        expect(sineOut.dominantCycle).toBeLessThan(23.0);
    });

    it('verifies metadata, mnemonics, entity wrappers, factory, and parameter rejection', () => {
        expect(CombBandPassSpectrumOutput.Value).toBe(0);
        expect(CombBandPassSpectrumOutput.DominantCycle).toBe(1);
        expect(defaultParams().minPeriod).toBe(10);
        expect(defaultParams().bandwidth).toBe(0.3);

        const desc = descriptorOf(IndicatorIdentifier.CombBandPassSpectrum);
        expect(desc).toBeDefined();
        expect(desc!.adaptivity).toBe(Adaptivity.Static);
        expect(desc!.outputs.length).toBe(2);

        const indicator = CombBandPassSpectrum.default();
        const meta = indicator.metadata();
        expect(meta.identifier).toBe(IndicatorIdentifier.CombBandPassSpectrum);
        expect(meta.mnemonic).toBe('cbps(10, 48, 0.3, 48, 10)');
        expect(meta.outputs.length).toBe(2);
        expect(meta.outputs[0].shape).toBe(Shape.Heatmap);
        expect(meta.outputs[0].mnemonic).toBe('cbps(10, 48, 0.3, 48, 10)');
        expect(meta.outputs[1].shape).toBe(Shape.Scalar);
        expect(meta.outputs[1].mnemonic).toBe('cbps-dc(10, 48, 0.3, 48, 10)');

        const combinations: Array<[BarComponent?, QuoteComponent?, TradeComponent?]> = [
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
            const params: CombBandPassSpectrumParams = {
                barComponent: barComp,
                quoteComponent: quoteComp,
                tradeComponent: tradeComp,
            };
            const inst = new CombBandPassSpectrum(params);
            const suffix = componentTripleMnemonic(
                barComp ?? DefaultBarComponent,
                quoteComp ?? DefaultQuoteComponent,
                tradeComp ?? DefaultTradeComponent,
            );
            const wantPrimary = `cbps(10, 48, 0.3, 48, 10${suffix})`;
            const wantDC = `cbps-dc(10, 48, 0.3, 48, 10${suffix})`;
            expect(inst.metadata().mnemonic).toBe(wantPrimary);
            expect(combBandPassSpectrumMnemonic(params)).toBe(wantPrimary);
            expect(combBandPassSpectrumDominantCycleMnemonic(params)).toBe(wantDC);
        }

        const barInd = CombBandPassSpectrum.default();
        const scalarInd = CombBandPassSpectrum.default();
        const quoteInd = CombBandPassSpectrum.default();
        const tradeInd = CombBandPassSpectrum.default();
        for (let index = 0; index < 70; index++) {
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

        const factoryDefault = createIndicator(IndicatorIdentifier.CombBandPassSpectrum);
        expect(factoryDefault.metadata().mnemonic).toBe('cbps(10, 48, 0.3, 48, 10)');
        const factoryCustom = createIndicator(IndicatorIdentifier.CombBandPassSpectrum, {
            minPeriod: 8,
            maxPeriod: 32,
            bandwidth: 0.4,
            highpassPeriod: 40,
            smoothingPeriod: 10.5,
            spectralDilationCompensation: false,
            enhanceResolution: true,
            barComponent: BarComponent.Median,
        });
        expect(factoryCustom.metadata().mnemonic).toBe('cbps(8, 32, 0.4, 40, 10.5, no-sdc, enh, hl/2)');
        expect(factoryCustom.metadata().outputs[1].mnemonic).toBe('cbps-dc(8, 32, 0.4, 40, 10.5, no-sdc, hl/2)');

        for (const badParams of [
            { minPeriod: 2 },
            { minPeriod: 0 },
            { minPeriod: 10.5 },
            { minPeriod: true },
            { minPeriod: 10, maxPeriod: 9 },
            { maxPeriod: true },
            { bandwidth: 0 },
            { bandwidth: -0.1 },
            { bandwidth: 2.5 },
            { minPeriod: 3, maxPeriod: 10, bandwidth: 0.75 },
            { bandwidth: Number.NaN },
            { bandwidth: Number.POSITIVE_INFINITY },
            { bandwidth: true },
            { highpassPeriod: 4 },
            { highpassPeriod: Number.NaN },
            { highpassPeriod: Number.POSITIVE_INFINITY },
            { smoothingPeriod: 1 },
            { smoothingPeriod: false },
            { spectralDilationCompensation: 1 },
            { enhanceResolution: 'yes' },
            { length: 48 },
            { windowLength: 48 },
            { disableSpectralDilationCompensation: true },
            { disableAutomaticGainControl: true },
            { automaticGainControlDecayFactor: 0.995 },
            { fixedNormalization: true },
            { unknownKey: 1 },
            { minPeriod: null },
            { barComponent: 999 },
        ]) {
            expect(() => new CombBandPassSpectrum(badParams as never)).toThrow();
            expect(() =>
                createIndicator(IndicatorIdentifier.CombBandPassSpectrum, badParams),
            ).toThrow();
        }
    });
});
