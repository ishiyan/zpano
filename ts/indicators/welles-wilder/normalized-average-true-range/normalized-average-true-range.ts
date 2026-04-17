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
import { AverageTrueRange } from '../average-true-range/average-true-range';
import { NormalizedAverageTrueRangeOutput } from './normalized-average-true-range-output';

const natrMnemonic = 'natr';
const natrDescription = 'Normalized Average True Range';

/**
 * Welles Wilder's Normalized Average True Range (NATR) indicator.
 *
 * NATR is calculated as (ATR / close) * 100, where ATR is the Average True Range.
 * If close == 0, the result is 0 (not division by zero).
 * The indicator is not primed during the first length updates.
 */
export class NormalizedAverageTrueRange implements Indicator {

  private readonly length_: number;
  private value_ = NaN;
  private primed_ = false;
  private readonly averageTrueRange: AverageTrueRange;

  constructor(length: number) {
    if (length < 1) {
      throw new Error(`invalid length ${length}: must be >= 1`);
    }

    this.length_ = length;
    this.averageTrueRange = new AverageTrueRange(length);
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
      kind: NormalizedAverageTrueRangeOutput.NormalizedAverageTrueRangeValue,
      type: OutputType.Scalar,
      mnemonic: natrMnemonic,
      description: natrDescription,
    };

    return {
      type: IndicatorType.NormalizedAverageTrueRange,
      mnemonic: natrMnemonic,
      description: natrDescription,
      outputs: [outputMeta],
    };
  }

  /** Updates the Normalized Average True Range given the next bar's close, high, and low values. */
  public update(close: number, high: number, low: number): number {
    if (isNaN(close) || isNaN(high) || isNaN(low)) {
      return NaN;
    }

    const atrValue = this.averageTrueRange.update(close, high, low);

    if (this.averageTrueRange.isPrimed()) {
      this.primed_ = true;

      if (close === 0) {
        this.value_ = 0;
      } else {
        this.value_ = (atrValue / close) * 100;
      }
    }

    return this.primed_ ? this.value_ : NaN;
  }

  /** Updates the Normalized Average True Range using a single sample value as a substitute for high, low, and close. */
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
