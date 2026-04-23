import { buildMetadata } from '../../core/build-metadata';
import { Bar } from '../../../entities/bar';
import { BarComponent, barComponentValue } from '../../../entities/bar-component';
import { Quote } from '../../../entities/quote';
import { DefaultQuoteComponent, quoteComponentValue } from '../../../entities/quote-component';
import { Scalar } from '../../../entities/scalar';
import { Trade } from '../../../entities/trade';
import { DefaultTradeComponent, tradeComponentValue } from '../../../entities/trade-component';
import { Indicator } from '../../core/indicator';
import { IndicatorMetadata } from '../../core/indicator-metadata';
import { IndicatorOutput } from '../../core/indicator-output';
import { IndicatorIdentifier } from '../../core/indicator-identifier';
import { componentTripleMnemonic } from '../../core/component-triple-mnemonic';
import { InstantaneousTrendLineLengthParams } from './length-params';
import { InstantaneousTrendLineSmoothingFactorParams } from './smoothing-factor-params';

const guardLength = (object: any): object is InstantaneousTrendLineLengthParams => 'length' in object;

/** Function to calculate mnemonic of an __InstantaneousTrendLine__ indicator. */
export const instantaneousTrendLineMnemonic =
  (params: InstantaneousTrendLineLengthParams | InstantaneousTrendLineSmoothingFactorParams): string => {
  const epsilon = 0.00000001;
  let length: number;
  if (guardLength(params)) {
    length = Math.floor(params.length);
  } else {
    const alpha = params.smoothingFactor;
    if (alpha < epsilon) {
      length = Number.MAX_SAFE_INTEGER;
    } else {
      length = Math.round(2 / alpha) - 1;
    }
  }

  const cm = componentTripleMnemonic(
    params.barComponent ?? BarComponent.Median,
    params.quoteComponent,
    params.tradeComponent,
  );

  return `iTrend(${length}${cm})`;
};

/** __Instantaneous Trend Line__ (Ehler's Instantaneous Trend Line, iTrend) is described
 * in Ehler's book "Cybernetic Analysis for Stocks and Futures" (2004):
 *
 *	H(z) = ((α-α²/4) + α²z⁻¹/2 - (α-3α²/4)z⁻²) / (1 - 2(1-α)z⁻¹ + (1-α)²z⁻²)
 *
 * which is a complementary low-pass filter found by subtracting the CyberCycle
 * high-pass filter from unity.
 *
 * The Instantaneous Trend Line has zero lag and about the same smoothing as an
 * Exponential Moving Average with the same α.
 *
 * The indicator has two outputs: the trend line value and a trigger line.
 *
 * Reference:
 *
 *	Ehlers, John F. (2004). Cybernetic Analysis for Stocks and Futures. Wiley.
 */
export class InstantaneousTrendLine implements Indicator {
  private readonly lengthValue: number;
  private readonly smoothingFactorValue: number;
  private readonly coeff1: number;
  private readonly coeff2: number;
  private readonly coeff3: number;
  private readonly coeff4: number;
  private readonly coeff5: number;
  private count: number = 0;
  private previousSample1: number = 0;
  private previousSample2: number = 0;
  private previousTrendLine1: number = 0;
  private previousTrendLine2: number = 0;
  private trendLineValue: number = Number.NaN;
  private triggerLineValue: number = Number.NaN;
  private primed: boolean = false;
  private readonly mnemonicStr: string;
  private readonly descriptionStr: string;
  private readonly mnemonicTrig: string;
  private readonly descriptionTrig: string;

  private readonly barComponentFunc: (bar: Bar) => number;
  private readonly quoteComponentFunc: (quote: Quote) => number;
  private readonly tradeComponentFunc: (trade: Trade) => number;

  /**
   * Constructs an instance given a length or a smoothing factor.
   */
  public constructor(params: InstantaneousTrendLineLengthParams | InstantaneousTrendLineSmoothingFactorParams) {
    const epsilon = 0.00000001;
    let length: number;
    let alpha: number;

    if (guardLength(params)) {
      length = Math.floor(params.length);
      if (length < 1) {
        throw new Error('length should be a positive integer');
      }

      alpha = 2 / (1 + length);
    } else {
      alpha = params.smoothingFactor;
      if (alpha < 0 || alpha > 1) {
        throw new Error('smoothing factor should be in range [0, 1]');
      }

      if (alpha < epsilon) {
        length = Number.MAX_SAFE_INTEGER;
      } else {
        length = Math.round(2 / alpha) - 1;
      }
    }

    this.lengthValue = length;
    this.smoothingFactorValue = alpha;

    // Calculate coefficients.
    // H(z) = ((α-α²/4) + α²z⁻¹/2 - (α-3α²/4)z⁻²) / (1 - 2(1-α)z⁻¹ + (1-α)²z⁻²)
    const a2 = alpha * alpha;
    this.coeff1 = alpha - a2 / 4;
    this.coeff2 = a2 / 2;
    this.coeff3 = -(alpha - 3 * a2 / 4);

    const x = 1 - alpha;
    this.coeff4 = 2 * x;
    this.coeff5 = -(x * x);

    // Resolve component defaults and create component functions.
    // InstantaneousTrendLine default bar component is Median, not Close.
    const bc = params.barComponent ?? BarComponent.Median;
    const qc = params.quoteComponent ?? DefaultQuoteComponent;
    const tc = params.tradeComponent ?? DefaultTradeComponent;

    this.barComponentFunc = barComponentValue(bc);
    this.quoteComponentFunc = quoteComponentValue(qc);
    this.tradeComponentFunc = tradeComponentValue(tc);

    // Build mnemonics.
    const cm = componentTripleMnemonic(
      params.barComponent ?? BarComponent.Median,
      params.quoteComponent,
      params.tradeComponent,
    );

    this.mnemonicStr = `iTrend(${length}${cm})`;
    this.mnemonicTrig = `iTrendTrigger(${length}${cm})`;

    const descr = 'Instantaneous Trend Line ';
    const descrTr = 'Instantaneous Trend Line trigger ';
    this.descriptionStr = descr + this.mnemonicStr;
    this.descriptionTrig = descrTr + this.mnemonicTrig;
  }

  /** Indicates whether an indicator is primed. */
  public isPrimed(): boolean { return this.primed; }

  /** Describes a requested output data of an indicator. */
  public metadata(): IndicatorMetadata {
    return buildMetadata(
      IndicatorIdentifier.InstantaneousTrendLine,
      this.mnemonicStr,
      this.descriptionStr,
      [
        { mnemonic: this.mnemonicStr, description: this.descriptionStr },
        { mnemonic: this.mnemonicTrig, description: this.descriptionTrig },
      ],
    );
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

  /** Updates the value of the instantaneous trend line given the next sample. */
  public update(sample: number): number {
    if (Number.isNaN(sample)) {
      return Number.NaN;
    }

    if (this.primed) {
      this.trendLineValue = this.coeff1 * sample + this.coeff2 * this.previousSample1
        + this.coeff3 * this.previousSample2
        + this.coeff4 * this.previousTrendLine1 + this.coeff5 * this.previousTrendLine2;
      this.triggerLineValue = 2 * this.trendLineValue - this.previousTrendLine2;

      this.previousSample2 = this.previousSample1;
      this.previousSample1 = sample;
      this.previousTrendLine2 = this.previousTrendLine1;
      this.previousTrendLine1 = this.trendLineValue;

      return this.trendLineValue;
    }

    this.count++;

    switch (this.count) {
      case 1:
        this.previousSample2 = sample;
        return Number.NaN;
      case 2:
        this.previousSample1 = sample;
        return Number.NaN;
      case 3:
        this.previousTrendLine2 = (sample + 2 * this.previousSample1 + this.previousSample2) / 4;

        this.previousSample2 = this.previousSample1;
        this.previousSample1 = sample;
        return Number.NaN;
      case 4:
        this.previousTrendLine1 = (sample + 2 * this.previousSample1 + this.previousSample2) / 4;

        this.previousSample2 = this.previousSample1;
        this.previousSample1 = sample;
        return Number.NaN;
      case 5:
        this.trendLineValue = this.coeff1 * sample + this.coeff2 * this.previousSample1
          + this.coeff3 * this.previousSample2
          + this.coeff4 * this.previousTrendLine1 + this.coeff5 * this.previousTrendLine2;
        this.triggerLineValue = 2 * this.trendLineValue - this.previousTrendLine2;

        this.previousSample2 = this.previousSample1;
        this.previousSample1 = sample;
        this.previousTrendLine2 = this.previousTrendLine1;
        this.previousTrendLine1 = this.trendLineValue;
        this.primed = true;

        return this.trendLineValue;
      default:
        return Number.NaN;
    }
  }

  private updateEntity(time: Date, sample: number): IndicatorOutput {
    const v = this.update(sample);

    let trig = this.triggerLineValue;
    if (Number.isNaN(v)) {
      trig = Number.NaN;
    }

    const scalarValue = new Scalar();
    scalarValue.time = time;
    scalarValue.value = v;

    const scalarTrigger = new Scalar();
    scalarTrigger.time = time;
    scalarTrigger.value = trig;

    return [scalarValue, scalarTrigger];
  }
}
