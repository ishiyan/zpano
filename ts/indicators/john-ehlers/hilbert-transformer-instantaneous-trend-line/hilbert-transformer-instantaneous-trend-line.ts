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
import { HilbertTransformerCycleEstimator } from '../hilbert-transformer/cycle-estimator';
import { HilbertTransformerCycleEstimatorType } from '../hilbert-transformer/cycle-estimator-type';
import { createEstimator, estimatorMoniker } from '../hilbert-transformer/common';
import { HilbertTransformerInstantaneousTrendLineParams } from './params';

const DEFAULT_ALPHA_EMA_PERIOD_ADDITIONAL = 0.33;
const DEFAULT_SMOOTHING_LENGTH = 4;
const DEFAULT_ALPHA_EMA_QI = 0.2;
const DEFAULT_ALPHA_EMA_PERIOD = 0.2;
// MBST's DominantCyclePeriod default warm-up is MaxPeriod * 2 = 100.
const DEFAULT_WARM_UP_PERIOD = 100;
const DEFAULT_TREND_LINE_SMOOTHING_LENGTH = 4;
const DEFAULT_CYCLE_PART_MULTIPLIER = 1.0;
const MAX_CYCLE_PART_MULTIPLIER = 10.0;

/** __Hilbert Transformer Instantaneous Trend Line__ (Ehlers) computes a smoothed trend line
 * derived from averaging over a window whose length tracks the smoothed dominant cycle period
 * produced by a Hilbert transformer cycle estimator.
 *
 * It exposes two outputs:
 *
 *	- Value: the instantaneous trend line value, a length-`trendLineSmoothingLength` WMA of
 *	  simple averages over the last `round(smoothedPeriod·cyclePartMultiplier)` raw input samples.
 *	- DominantCyclePeriod: the additionally EMA-smoothed dominant cycle period.
 *
 * Reference:
 *
 *	John Ehlers, Rocket Science for Traders, Wiley, 2001, 0471405671, pp 107-112.
 */
export class HilbertTransformerInstantaneousTrendLine implements Indicator {
  private readonly htce: HilbertTransformerCycleEstimator;
  private readonly alphaEmaPeriodAdditional: number;
  private readonly oneMinAlphaEmaPeriodAdditional: number;
  private readonly cyclePartMultiplier: number;
  private readonly trendLineSmoothingLength: number;
  private readonly coeff0: number;
  private readonly coeff1: number;
  private readonly coeff2: number;
  private readonly coeff3: number;
  private readonly input: number[];
  private readonly inputLength: number;
  private readonly inputLengthMin1: number;
  private smoothedPeriod = 0;
  private value = Number.NaN;
  private average1 = 0;
  private average2 = 0;
  private average3 = 0;
  private primed = false;

  private readonly mnemonicValue: string;
  private readonly descriptionValue: string;
  private readonly mnemonicDCP: string;
  private readonly descriptionDCP: string;

  private readonly barComponentFunc: (bar: Bar) => number;
  private readonly quoteComponentFunc: (quote: Quote) => number;
  private readonly tradeComponentFunc: (trade: Trade) => number;

  /** Creates an instance using default parameters (α=0.33, trendLineSmoothingLength=4,
   * cyclePartMultiplier=1.0, HomodyneDiscriminator cycle estimator with smoothingLength=4,
   * αq=0.2, αp=0.2, warmUpPeriod=100, BarComponent.Median). */
  public static default(): HilbertTransformerInstantaneousTrendLine {
    return new HilbertTransformerInstantaneousTrendLine({
      alphaEmaPeriodAdditional: DEFAULT_ALPHA_EMA_PERIOD_ADDITIONAL,
      trendLineSmoothingLength: DEFAULT_TREND_LINE_SMOOTHING_LENGTH,
      cyclePartMultiplier: DEFAULT_CYCLE_PART_MULTIPLIER,
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
  public static fromParams(
    params: HilbertTransformerInstantaneousTrendLineParams,
  ): HilbertTransformerInstantaneousTrendLine {
    return new HilbertTransformerInstantaneousTrendLine(params);
  }

  private constructor(params: HilbertTransformerInstantaneousTrendLineParams) {
    const alpha = params.alphaEmaPeriodAdditional;
    if (alpha <= 0 || alpha > 1) {
      throw new Error(
        'invalid hilbert transformer instantaneous trend line parameters: '
          + 'α for additional smoothing should be in range (0, 1]',
      );
    }

    const tlsl = params.trendLineSmoothingLength ?? DEFAULT_TREND_LINE_SMOOTHING_LENGTH;
    if (tlsl < 2 || tlsl > 4 || !Number.isInteger(tlsl)) {
      throw new Error(
        'invalid hilbert transformer instantaneous trend line parameters: '
          + 'trend line smoothing length should be 2, 3, or 4',
      );
    }

    const cpm = params.cyclePartMultiplier ?? DEFAULT_CYCLE_PART_MULTIPLIER;
    if (cpm <= 0 || cpm > MAX_CYCLE_PART_MULTIPLIER) {
      throw new Error(
        'invalid hilbert transformer instantaneous trend line parameters: '
          + 'cycle part multiplier should be in range (0, 10]',
      );
    }

    this.alphaEmaPeriodAdditional = alpha;
    this.oneMinAlphaEmaPeriodAdditional = 1 - alpha;
    this.cyclePartMultiplier = cpm;
    this.trendLineSmoothingLength = tlsl;

    // Default to BarComponent.Median (MBST default; always shown in mnemonic).
    const bc = params.barComponent ?? BarComponent.Median;
    const qc = params.quoteComponent ?? DefaultQuoteComponent;
    const tc = params.tradeComponent ?? DefaultTradeComponent;

    this.barComponentFunc = barComponentValue(bc);
    this.quoteComponentFunc = quoteComponentValue(qc);
    this.tradeComponentFunc = tradeComponentValue(tc);

    this.htce = createEstimator(params.estimatorType, params.estimatorParams);

    const effectiveType = params.estimatorType ?? HilbertTransformerCycleEstimatorType.HomodyneDiscriminator;
    let em = '';
    const isDefaultHd = effectiveType === HilbertTransformerCycleEstimatorType.HomodyneDiscriminator
      && this.htce.smoothingLength === DEFAULT_SMOOTHING_LENGTH
      && this.htce.alphaEmaQuadratureInPhase === DEFAULT_ALPHA_EMA_QI
      && this.htce.alphaEmaPeriod === DEFAULT_ALPHA_EMA_PERIOD;
    if (!isDefaultHd) {
      const moniker = estimatorMoniker(effectiveType, this.htce);
      if (moniker.length > 0) {
        em = ', ' + moniker;
      }
    }

    const cm = componentTripleMnemonic(bc, qc, tc);
    const a = alpha.toFixed(3);
    const c = cpm.toFixed(3);

    this.mnemonicValue = `htitl(${a}, ${tlsl}, ${c}${em}${cm})`;
    this.mnemonicDCP = `dcp(${a}${em}${cm})`;

    this.descriptionValue = 'Hilbert transformer instantaneous trend line ' + this.mnemonicValue;
    this.descriptionDCP = 'Dominant cycle period ' + this.mnemonicDCP;

    // WMA coefficients.
    let c0 = 0, c1 = 0, c2 = 0, c3 = 0;
    if (tlsl === 2) {
      c0 = 2 / 3;
      c1 = 1 / 3;
    } else if (tlsl === 3) {
      c0 = 3 / 6;
      c1 = 2 / 6;
      c2 = 1 / 6;
    } else { // tlsl === 4
      c0 = 4 / 10;
      c1 = 3 / 10;
      c2 = 2 / 10;
      c3 = 1 / 10;
    }
    this.coeff0 = c0;
    this.coeff1 = c1;
    this.coeff2 = c2;
    this.coeff3 = c3;

    const maxPeriod = this.htce.maxPeriod;
    this.input = new Array<number>(maxPeriod).fill(0);
    this.inputLength = maxPeriod;
    this.inputLengthMin1 = maxPeriod - 1;
  }

  /** Indicates whether the indicator is primed. */
  public isPrimed(): boolean { return this.primed; }

  /** Describes the output data of the indicator. */
  public metadata(): IndicatorMetadata {
    return buildMetadata(
      IndicatorIdentifier.HilbertTransformerInstantaneousTrendLine,
      this.mnemonicValue,
      this.descriptionValue,
      [
        { mnemonic: this.mnemonicValue, description: this.descriptionValue },
        { mnemonic: this.mnemonicDCP, description: this.descriptionDCP },
      ],
    );
  }

  /** Updates the indicator given the next sample value. Returns the pair
   * (value, period). Returns (NaN, NaN) if not yet primed. */
  public update(sample: number): [number, number] {
    if (Number.isNaN(sample)) {
      return [sample, sample];
    }

    this.htce.update(sample);
    this.pushInput(sample);

    if (this.primed) {
      this.smoothedPeriod = this.alphaEmaPeriodAdditional * this.htce.period
        + this.oneMinAlphaEmaPeriodAdditional * this.smoothedPeriod;
      const average = this.calculateAverage();
      this.value = this.coeff0 * average
        + this.coeff1 * this.average1
        + this.coeff2 * this.average2
        + this.coeff3 * this.average3;
      this.average3 = this.average2;
      this.average2 = this.average1;
      this.average1 = average;
      return [this.value, this.smoothedPeriod];
    }

    if (this.htce.primed) {
      this.primed = true;
      this.smoothedPeriod = this.htce.period;
      const average = this.calculateAverage();
      this.value = average;
      this.average1 = average;
      this.average2 = average;
      this.average3 = average;
      return [this.value, this.smoothedPeriod];
    }

    return [Number.NaN, Number.NaN];
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
    const [value, period] = this.update(sample);

    const sv = new Scalar();
    sv.time = time;
    sv.value = value;

    const sp = new Scalar();
    sp.time = time;
    sp.value = period;

    return [sv, sp];
  }

  private pushInput(value: number): void {
    for (let i = this.inputLengthMin1; i > 0; i--) {
      this.input[i] = this.input[i - 1];
    }
    this.input[0] = value;
  }

  private calculateAverage(): number {
    // Compute simple average over a window tracking the smoothed dominant cycle period.
    let length = Math.floor(this.smoothedPeriod * this.cyclePartMultiplier + 0.5);
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
}
