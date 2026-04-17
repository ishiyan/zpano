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
import { TrueRangeOutput } from './true-range-output';

const trMnemonic = 'tr';
const trDescription = 'True Range';

/**
 * Welles Wilder's True Range indicator.
 *
 * The True Range is defined as the largest of:
 * - the distance from today's high to today's low
 * - the distance from yesterday's close to today's high
 * - the distance from yesterday's close to today's low
 *
 * The first update stores the close and returns NaN (not primed).
 * The indicator is primed from the second update onward.
 *
 * Unlike most indicators, TrueRange requires bar data (high, low, close)
 * and does not use a single bar component. For scalar, quote, and trade updates,
 * the single value is used as a substitute for high, low, and close.
 */
export class TrueRange implements Indicator {

  private previousClose = NaN;
  private value = NaN;
  private primed_ = false;

  /** Indicates whether the indicator is primed. */
  public isPrimed(): boolean {
    return this.primed_;
  }

  /** Describes the output data of the indicator. */
  public metadata(): IndicatorMetadata {
    const outputMeta: OutputMetadata = {
      kind: TrueRangeOutput.TrueRangeValue,
      type: OutputType.Scalar,
      mnemonic: trMnemonic,
      description: trDescription,
    };

    return {
      type: IndicatorType.TrueRange,
      mnemonic: trMnemonic,
      description: trDescription,
      outputs: [outputMeta],
    };
  }

  /** Updates the True Range given the next bar's close, high, and low values. */
  public update(close: number, high: number, low: number): number {
    if (isNaN(close) || isNaN(high) || isNaN(low)) {
      return NaN;
    }

    if (!this.primed_) {
      if (isNaN(this.previousClose)) {
        this.previousClose = close;
        return NaN;
      }

      this.primed_ = true;
    }

    let greatest = high - low;

    const diffHigh = Math.abs(high - this.previousClose);
    if (greatest < diffHigh) {
      greatest = diffHigh;
    }

    const diffLow = Math.abs(low - this.previousClose);
    if (greatest < diffLow) {
      greatest = diffLow;
    }

    this.value = greatest;
    this.previousClose = close;

    return this.value;
  }

  /** Updates the True Range using a single sample value as a substitute for high, low, and close. */
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
