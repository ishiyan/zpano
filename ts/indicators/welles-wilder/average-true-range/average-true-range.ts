import { Bar } from '../../../entities/bar';
import { Quote } from '../../../entities/quote';
import { Scalar } from '../../../entities/scalar';
import { Trade } from '../../../entities/trade';
import { Indicator } from '../../core/indicator';
import { IndicatorMetadata } from '../../core/indicator-metadata';
import { IndicatorOutput } from '../../core/indicator-output';
import { IndicatorType } from '../../core/indicator-type';
import { OutputMetadata } from '../../core/outputs/output-metadata';
import { OutputType } from '../../core/outputs/output-type';
import { TrueRange } from '../true-range/true-range';
import { AverageTrueRangeOutput } from './average-true-range-output';

const atrMnemonic = 'atr';
const atrDescription = 'Average True Range';

/**
 * Welles Wilder's Average True Range (ATR) indicator.
 *
 * ATR averages True Range (TR) values over the specified length using the Wilder method:
 * - multiply the previous value by (length - 1)
 * - add the current TR value
 * - divide by length
 *
 * The initial ATR value is a simple average of the first length TR values.
 * The indicator is not primed during the first length updates.
 */
export class AverageTrueRange implements Indicator {

  private readonly length_: number;
  private readonly lastIndex: number;
  private stage = 0;
  private windowCount = 0;
  private window: number[];
  private windowSum = 0;
  private value_ = NaN;
  private primed_ = false;
  private readonly trueRange: TrueRange;

  constructor(length: number) {
    if (length < 1) {
      throw new Error(`invalid length ${length}: must be >= 1`);
    }

    this.length_ = length;
    this.lastIndex = length - 1;
    this.window = this.lastIndex > 0 ? new Array<number>(length) : [];
    this.trueRange = new TrueRange();
  }

  /** Returns the length parameter. */
  public get length(): number {
    return this.length_;
  }

  /** Indicates whether the indicator is primed. */
  public isPrimed(): boolean {
    return this.primed_;
  }

  /** Describes the output data of the indicator. */
  public metadata(): IndicatorMetadata {
    const outputMeta: OutputMetadata = {
      kind: AverageTrueRangeOutput.AverageTrueRangeValue,
      type: OutputType.Scalar,
      mnemonic: atrMnemonic,
      description: atrDescription,
    };

    return {
      type: IndicatorType.AverageTrueRange,
      mnemonic: atrMnemonic,
      description: atrDescription,
      outputs: [outputMeta],
    };
  }

  /** Updates the Average True Range given the next bar's close, high, and low values. */
  public update(close: number, high: number, low: number): number {
    if (isNaN(close) || isNaN(high) || isNaN(low)) {
      return NaN;
    }

    const trueRangeValue = this.trueRange.update(close, high, low);

    if (this.lastIndex === 0) {
      this.value_ = trueRangeValue;

      if (this.stage === 0) {
        this.stage++;
      } else if (this.stage === 1) {
        this.stage++;
        this.primed_ = true;
      }

      return this.value_;
    }

    if (this.stage > 1) {
      // Wilder smoothing method.
      this.value_ *= this.lastIndex;
      this.value_ += trueRangeValue;
      this.value_ /= this.length_;
      return this.value_;
    }

    if (this.stage === 1) {
      this.windowSum += trueRangeValue;
      this.window[this.windowCount] = trueRangeValue;
      this.windowCount++;

      if (this.windowCount === this.length_) {
        this.stage++;
        this.primed_ = true;
        this.value_ = this.windowSum / this.length_;
      }

      return this.primed_ ? this.value_ : NaN;
    }

    // The very first sample is used by the True Range.
    this.stage++;
    return NaN;
  }

  /** Updates the Average True Range using a single sample value as a substitute for high, low, and close. */
  public updateSample(sample: number): number {
    return this.update(sample, sample, sample);
  }

  /** Updates the indicator given the next scalar sample. */
  public updateScalar(sample: Scalar): IndicatorOutput {
    const v = sample.value;
    const scalar = new Scalar();
    scalar.time = sample.time;
    scalar.value = this.update(v, v, v);
    return [scalar];
  }

  /** Updates the indicator given the next bar sample. */
  public updateBar(sample: Bar): IndicatorOutput {
    const scalar = new Scalar();
    scalar.time = sample.time;
    scalar.value = this.update(sample.close, sample.high, sample.low);
    return [scalar];
  }

  /** Updates the indicator given the next quote sample. */
  public updateQuote(sample: Quote): IndicatorOutput {
    const v = (sample.bid + sample.ask) / 2;
    const scalar = new Scalar();
    scalar.time = sample.time;
    scalar.value = this.update(v, v, v);
    return [scalar];
  }

  /** Updates the indicator given the next trade sample. */
  public updateTrade(sample: Trade): IndicatorOutput {
    const v = sample.price;
    const scalar = new Scalar();
    scalar.time = sample.time;
    scalar.value = this.update(v, v, v);
    return [scalar];
  }
}
