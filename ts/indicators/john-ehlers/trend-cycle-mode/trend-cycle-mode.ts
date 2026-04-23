import { buildMetadata } from '../../core/build-metadata';
import { Bar } from '../../../entities/bar';
import { BarComponent, barComponentValue } from '../../../entities/bar-component';
import { Quote } from '../../../entities/quote';
import { DefaultQuoteComponent, quoteComponentValue } from '../../../entities/quote-component';
import { Scalar } from '../../../entities/scalar';
import { Trade } from '../../../entities/trade';
import { DefaultTradeComponent, tradeComponentValue } from '../../../entities/trade-component';
import { componentTripleMnemonic } from '../../core/component-triple-mnemonic';
import { Indicator } from '../../core/indicator';
import { IndicatorMetadata } from '../../core/indicator-metadata';
import { IndicatorOutput } from '../../core/indicator-output';
import { IndicatorIdentifier } from '../../core/indicator-identifier';
import { DominantCycle } from '../dominant-cycle/dominant-cycle';
import { HilbertTransformerCycleEstimatorType } from '../hilbert-transformer/cycle-estimator-type';
import { createEstimator, estimatorMoniker } from '../hilbert-transformer/common';
import { TrendCycleModeParams } from './params';

const DEFAULT_ALPHA_EMA_PERIOD_ADDITIONAL = 0.33;
const DEFAULT_SMOOTHING_LENGTH = 4;
const DEFAULT_ALPHA_EMA_QI = 0.2;
const DEFAULT_ALPHA_EMA_PERIOD = 0.2;
// MBST's DominantCyclePeriod default warm-up is MaxPeriod * 2 = 100.
const DEFAULT_WARM_UP_PERIOD = 100;
const DEFAULT_TREND_LINE_SMOOTHING_LENGTH = 4;
const DEFAULT_CYCLE_PART_MULTIPLIER = 1.0;
const DEFAULT_SEPARATION_PERCENTAGE = 1.5;
const MAX_CYCLE_PART_MULTIPLIER = 10.0;
const MAX_SEPARATION_PERCENTAGE = 100.0;

const DEG2RAD = Math.PI / 180.0;
const LEAD_OFFSET = 45.0;
const FULL_CYCLE = 360.0;
const EPSILON = 1e-308;

/** __Trend versus Cycle Mode__ (Ehlers) classifies the market as in-trend or in-cycle based
 * on the behaviour of the instantaneous dominant cycle period/phase and a WMA-smoothed
 * instantaneous trend line.
 *
 * It exposes eight outputs:
 *
 *	- Value: +1 in trend mode, -1 in cycle mode.
 *	- IsTrendMode: 1 if the trend mode is declared, 0 otherwise.
 *	- IsCycleMode: 1 if the cycle mode is declared, 0 otherwise (= 1 − IsTrendMode).
 *	- InstantaneousTrendLine: the WMA-smoothed instantaneous trend line.
 *	- SineWave: sin(phase·Deg2Rad).
 *	- SineWaveLead: sin((phase+45)·Deg2Rad).
 *	- DominantCyclePeriod: the smoothed dominant cycle period.
 *	- DominantCyclePhase: the dominant cycle phase, in degrees.
 *
 * Reference:
 *
 *	John Ehlers, Rocket Science for Traders, Wiley, 2001, 0471405671, pp 113-118.
 */
export class TrendCycleMode implements Indicator {
  private readonly dc: DominantCycle;
  private readonly cyclePartMultiplier: number;
  private readonly separationFactor: number;
  private readonly coeff0: number;
  private readonly coeff1: number;
  private readonly coeff2: number;
  private readonly coeff3: number;
  private readonly input: number[];
  private readonly inputLength: number;
  private readonly inputLengthMin1: number;

  private trendline = Number.NaN;
  private trendAverage1 = 0;
  private trendAverage2 = 0;
  private trendAverage3 = 0;
  private sinWave = Number.NaN;
  private sinWaveLead = Number.NaN;
  private previousDcPhase = 0;
  private previousSineLeadWaveDifference = 0;
  private samplesInTrend = 0;
  private isTrendMode = true;
  private primed = false;

  private readonly mnemonicValue: string;
  private readonly descriptionValue: string;
  private readonly mnemonicTrend: string;
  private readonly descriptionTrend: string;
  private readonly mnemonicCycle: string;
  private readonly descriptionCycle: string;
  private readonly mnemonicITL: string;
  private readonly descriptionITL: string;
  private readonly mnemonicSine: string;
  private readonly descriptionSine: string;
  private readonly mnemonicSineLead: string;
  private readonly descriptionSineLead: string;
  private readonly mnemonicDCP: string;
  private readonly descriptionDCP: string;
  private readonly mnemonicDCPhase: string;
  private readonly descriptionDCPhase: string;

  private readonly barComponentFunc: (bar: Bar) => number;
  private readonly quoteComponentFunc: (quote: Quote) => number;
  private readonly tradeComponentFunc: (trade: Trade) => number;

  /** Creates an instance using default parameters (α=0.33, trendLineSmoothingLength=4,
   * cyclePartMultiplier=1.0, separationPercentage=1.5, HomodyneDiscriminator cycle estimator
   * with smoothingLength=4, αq=0.2, αp=0.2, warmUpPeriod=100, BarComponent.Median). */
  public static default(): TrendCycleMode {
    return new TrendCycleMode({
      alphaEmaPeriodAdditional: DEFAULT_ALPHA_EMA_PERIOD_ADDITIONAL,
      trendLineSmoothingLength: DEFAULT_TREND_LINE_SMOOTHING_LENGTH,
      cyclePartMultiplier: DEFAULT_CYCLE_PART_MULTIPLIER,
      separationPercentage: DEFAULT_SEPARATION_PERCENTAGE,
      estimatorType: HilbertTransformerCycleEstimatorType.HomodyneDiscriminator,
      estimatorParams: {
        smoothingLength: DEFAULT_SMOOTHING_LENGTH,
        alphaEmaQuadratureInPhase: DEFAULT_ALPHA_EMA_QI,
        alphaEmaPeriod: DEFAULT_ALPHA_EMA_PERIOD,
        warmUpPeriod: DEFAULT_WARM_UP_PERIOD,
      },
    });
  }

  /** Creates an instance based on the given parameters. */
  public static fromParams(params: TrendCycleModeParams): TrendCycleMode {
    return new TrendCycleMode(params);
  }

  private constructor(params: TrendCycleModeParams) {
    const alpha = params.alphaEmaPeriodAdditional;
    if (alpha <= 0 || alpha > 1) {
      throw new Error('invalid trend cycle mode parameters: α for additional smoothing should be in range (0, 1]');
    }

    const tlsl = params.trendLineSmoothingLength ?? DEFAULT_TREND_LINE_SMOOTHING_LENGTH;
    if (tlsl < 2 || tlsl > 4 || !Number.isInteger(tlsl)) {
      throw new Error('invalid trend cycle mode parameters: trend line smoothing length should be 2, 3, or 4');
    }

    const cpm = params.cyclePartMultiplier ?? DEFAULT_CYCLE_PART_MULTIPLIER;
    if (cpm <= 0 || cpm > MAX_CYCLE_PART_MULTIPLIER) {
      throw new Error('invalid trend cycle mode parameters: cycle part multiplier should be in range (0, 10]');
    }

    const sep = params.separationPercentage ?? DEFAULT_SEPARATION_PERCENTAGE;
    if (sep <= 0 || sep > MAX_SEPARATION_PERCENTAGE) {
      throw new Error('invalid trend cycle mode parameters: separation percentage should be in range (0, 100]');
    }

    this.cyclePartMultiplier = cpm;
    this.separationFactor = sep / 100.0;

    // TCM defaults to BarComponent.Median (MBST default; always shown in mnemonic).
    const bc = params.barComponent ?? BarComponent.Median;
    const qc = params.quoteComponent ?? DefaultQuoteComponent;
    const tc = params.tradeComponent ?? DefaultTradeComponent;

    this.barComponentFunc = barComponentValue(bc);
    this.quoteComponentFunc = quoteComponentValue(qc);
    this.tradeComponentFunc = tradeComponentValue(tc);

    // Inner DominantCycle built with explicit components so its mnemonic contains hl/2.
    this.dc = DominantCycle.fromParams({
      alphaEmaPeriodAdditional: alpha,
      estimatorType: params.estimatorType,
      estimatorParams: params.estimatorParams,
      barComponent: bc,
      quoteComponent: qc,
      tradeComponent: tc,
    });

    // Compose estimator moniker (same logic as DominantCycle / SineWave).
    const effectiveType = params.estimatorType ?? HilbertTransformerCycleEstimatorType.HomodyneDiscriminator;
    const htce = createEstimator(params.estimatorType, params.estimatorParams);
    let em = '';
    const isDefaultHd = effectiveType === HilbertTransformerCycleEstimatorType.HomodyneDiscriminator
      && htce.smoothingLength === DEFAULT_SMOOTHING_LENGTH
      && htce.alphaEmaQuadratureInPhase === DEFAULT_ALPHA_EMA_QI
      && htce.alphaEmaPeriod === DEFAULT_ALPHA_EMA_PERIOD;
    if (!isDefaultHd) {
      const moniker = estimatorMoniker(effectiveType, htce);
      if (moniker.length > 0) {
        em = ', ' + moniker;
      }
    }

    const cm = componentTripleMnemonic(bc, qc, tc);
    const a = alpha.toFixed(3);
    const c = cpm.toFixed(3);
    const s = sep.toFixed(3);
    const tail = `${a}, ${tlsl}, ${c}, ${s}%${em}${cm}`;

    this.mnemonicValue = `tcm(${tail})`;
    this.mnemonicTrend = `tcm-trend(${tail})`;
    this.mnemonicCycle = `tcm-cycle(${tail})`;
    this.mnemonicITL = `tcm-itl(${tail})`;
    this.mnemonicSine = `tcm-sine(${tail})`;
    this.mnemonicSineLead = `tcm-sineLead(${tail})`;
    this.mnemonicDCP = `dcp(${a}${em}${cm})`;
    this.mnemonicDCPhase = `dcph(${a}${em}${cm})`;

    this.descriptionValue = 'Trend versus cycle mode ' + this.mnemonicValue;
    this.descriptionTrend = 'Trend versus cycle mode, is-trend flag ' + this.mnemonicTrend;
    this.descriptionCycle = 'Trend versus cycle mode, is-cycle flag ' + this.mnemonicCycle;
    this.descriptionITL = 'Trend versus cycle mode instantaneous trend line ' + this.mnemonicITL;
    this.descriptionSine = 'Trend versus cycle mode sine wave ' + this.mnemonicSine;
    this.descriptionSineLead = 'Trend versus cycle mode sine wave lead ' + this.mnemonicSineLead;
    this.descriptionDCP = 'Dominant cycle period ' + this.mnemonicDCP;
    this.descriptionDCPhase = 'Dominant cycle phase ' + this.mnemonicDCPhase;

    // WMA coefficients.
    let c0 = 0, c1 = 0, c2 = 0, c3 = 0;
    if (tlsl === 2) {
      c0 = 2 / 3; c1 = 1 / 3;
    } else if (tlsl === 3) {
      c0 = 3 / 6; c1 = 2 / 6; c2 = 1 / 6;
    } else { // tlsl === 4
      c0 = 4 / 10; c1 = 3 / 10; c2 = 2 / 10; c3 = 1 / 10;
    }
    this.coeff0 = c0;
    this.coeff1 = c1;
    this.coeff2 = c2;
    this.coeff3 = c3;

    const maxPeriod = this.dc.maxPeriod;
    this.input = new Array<number>(maxPeriod).fill(0);
    this.inputLength = maxPeriod;
    this.inputLengthMin1 = maxPeriod - 1;
  }

  /** Indicates whether the indicator is primed. */
  public isPrimed(): boolean { return this.primed; }

  /** Describes the output data of the indicator. */
  public metadata(): IndicatorMetadata {
    return buildMetadata(
      IndicatorIdentifier.TrendCycleMode,
      this.mnemonicValue,
      this.descriptionValue,
      [
        { mnemonic: this.mnemonicValue, description: this.descriptionValue },
        { mnemonic: this.mnemonicTrend, description: this.descriptionTrend },
        { mnemonic: this.mnemonicCycle, description: this.descriptionCycle },
        { mnemonic: this.mnemonicITL, description: this.descriptionITL },
        { mnemonic: this.mnemonicSine, description: this.descriptionSine },
        { mnemonic: this.mnemonicSineLead, description: this.descriptionSineLead },
        { mnemonic: this.mnemonicDCP, description: this.descriptionDCP },
        { mnemonic: this.mnemonicDCPhase, description: this.descriptionDCPhase },
      ],
    );
  }

  /** Updates the indicator given the next sample value. Returns the 8-tuple
   * (value, isTrendMode, isCycleMode, trendline, sineWave, sineWaveLead, period, phase).
   * isTrendMode / isCycleMode are encoded as 1 / 0 scalars. Returns all-NaN if not primed. */
  public update(sample: number): [number, number, number, number, number, number, number, number] {
    if (Number.isNaN(sample)) {
      return [sample, sample, sample, sample, sample, sample, sample, sample];
    }

    const [, period, phase] = this.dc.update(sample);
    const smoothedPrice = this.dc.smoothedPrice;

    this.pushInput(sample);

    if (this.primed) {
      const smoothedPeriod = period;
      const average = this.calculateTrendAverage(smoothedPeriod);
      this.trendline = this.coeff0 * average
        + this.coeff1 * this.trendAverage1
        + this.coeff2 * this.trendAverage2
        + this.coeff3 * this.trendAverage3;
      this.trendAverage3 = this.trendAverage2;
      this.trendAverage2 = this.trendAverage1;
      this.trendAverage1 = average;

      const diff = this.calculateSineLeadWaveDifference(phase);

      // Condition 1: a cycle mode exists for the half-period of a dominant cycle after
      // the SineWave vs SineWaveLead crossing.
      this.isTrendMode = true;

      if ((diff > 0 && this.previousSineLeadWaveDifference < 0)
        || (diff < 0 && this.previousSineLeadWaveDifference > 0)) {
        this.isTrendMode = false;
        this.samplesInTrend = 0;
      }

      this.previousSineLeadWaveDifference = diff;
      this.samplesInTrend++;

      if (this.samplesInTrend < 0.5 * smoothedPeriod) {
        this.isTrendMode = false;
      }

      // Condition 2: cycle mode if the measured phase rate of change is more than 2/3
      // the phase rate of change of the dominant cycle (360/period) and less than 1.5×.
      const phaseDelta = phase - this.previousDcPhase;
      this.previousDcPhase = phase;

      if (Math.abs(smoothedPeriod) > EPSILON) {
        const dcRate = FULL_CYCLE / smoothedPeriod;
        if (phaseDelta > (2.0 / 3.0) * dcRate && phaseDelta < 1.5 * dcRate) {
          this.isTrendMode = false;
        }
      }

      // Condition 3: if the WMA smoothed price is separated by more than the separation
      // percentage from the instantaneous trend line, force the trend mode.
      if (Math.abs(this.trendline) > EPSILON
        && Math.abs((smoothedPrice - this.trendline) / this.trendline) >= this.separationFactor) {
        this.isTrendMode = true;
      }

      return [
        this.mode(), this.isTrendFloat(), this.isCycleFloat(),
        this.trendline, this.sinWave, this.sinWaveLead, period, phase,
      ];
    }

    if (this.dc.isPrimed()) {
      this.primed = true;
      const smoothedPeriod = period;
      this.trendline = this.calculateTrendAverage(smoothedPeriod);
      this.trendAverage1 = this.trendline;
      this.trendAverage2 = this.trendline;
      this.trendAverage3 = this.trendline;

      this.previousDcPhase = phase;
      this.previousSineLeadWaveDifference = this.calculateSineLeadWaveDifference(phase);

      this.isTrendMode = true;
      this.samplesInTrend++;

      if (this.samplesInTrend < 0.5 * smoothedPeriod) {
        this.isTrendMode = false;
      }

      return [
        this.mode(), this.isTrendFloat(), this.isCycleFloat(),
        this.trendline, this.sinWave, this.sinWaveLead, period, phase,
      ];
    }

    const n = Number.NaN;
    return [n, n, n, n, n, n, n, n];
  }

  /** Updates an indicator given the next scalar sample. */
  public updateScalar(sample: Scalar): IndicatorOutput {
    return this.updateEntity(sample.time, sample.value);
  }

  /** Updates an indicator given the next bar sample. */
  public updateBar(sample: Bar): IndicatorOutput {
    return this.updateEntity(sample.time, this.barComponentFunc(sample));
  }

  /** Updates an indicator given the next quote sample. */
  public updateQuote(sample: Quote): IndicatorOutput {
    return this.updateEntity(sample.time, this.quoteComponentFunc(sample));
  }

  /** Updates an indicator given the next trade sample. */
  public updateTrade(sample: Trade): IndicatorOutput {
    return this.updateEntity(sample.time, this.tradeComponentFunc(sample));
  }

  private updateEntity(time: Date, sample: number): IndicatorOutput {
    const [value, trend, cycle, itl, sine, sineLead, period, phase] = this.update(sample);

    const out: Scalar[] = [];
    for (const v of [value, trend, cycle, itl, sine, sineLead, period, phase]) {
      const s = new Scalar();
      s.time = time;
      s.value = v;
      out.push(s);
    }
    return out;
  }

  private pushInput(value: number): void {
    for (let i = this.inputLengthMin1; i > 0; i--) {
      this.input[i] = this.input[i - 1];
    }
    this.input[0] = value;
  }

  private calculateTrendAverage(smoothedPeriod: number): number {
    let length = Math.floor(smoothedPeriod * this.cyclePartMultiplier + 0.5);
    if (length > this.inputLength) {
      length = this.inputLength;
    } else if (length < 1) {
      length = 1;
    }

    let sum = 0;
    for (let i = 0; i < length; i++) {
      sum += this.input[i];
    }
    return sum / length;
  }

  private calculateSineLeadWaveDifference(phase: number): number {
    const p = phase * DEG2RAD;
    this.sinWave = Math.sin(p);
    this.sinWaveLead = Math.sin(p + LEAD_OFFSET * DEG2RAD);
    return this.sinWave - this.sinWaveLead;
  }

  private mode(): number { return this.isTrendMode ? 1.0 : -1.0; }
  private isTrendFloat(): number { return this.isTrendMode ? 1.0 : 0.0; }
  private isCycleFloat(): number { return this.isTrendMode ? 0.0 : 1.0; }
}
