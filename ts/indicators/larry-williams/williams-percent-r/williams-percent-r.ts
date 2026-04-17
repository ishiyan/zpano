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
import { WilliamsPercentROutput } from './williams-percent-r-output';

const willrMnemonic = 'willr';
const willrDescription = 'Williams %R';
const defaultLength = 14;
const minLength = 2;

/**
 * Larry Williams' Williams %R momentum indicator.
 *
 * Williams %R reflects the level of the closing price relative to the
 * highest high over a lookback period. The oscillation ranges from 0 to -100;
 * readings from 0 to -20 are considered overbought, readings from -80 to -100
 * are considered oversold.
 *
 * The value is calculated as:
 *   %R = -100 * (HighestHigh - Close) / (HighestHigh - LowestLow)
 *
 * where HighestHigh and LowestLow are computed over the last `length` bars.
 * If HighestHigh equals LowestLow, the value is 0.
 *
 * The indicator requires bar data (high, low, close). For scalar, quote, and
 * trade updates, the single value is used as a substitute for all three.
 */
export class WilliamsPercentR implements Indicator {

  private readonly length: number;
  private readonly lengthMinOne: number;
  private readonly lowCircular: number[];
  private readonly highCircular: number[];
  private circularIndex = 0;
  private circularCount = 0;
  private value = NaN;
  private primed_ = false;

  constructor(length: number = defaultLength) {
    if (length < minLength) {
      length = defaultLength;
    }
    this.length = length;
    this.lengthMinOne = length - 1;
    this.lowCircular = new Array<number>(length);
    this.highCircular = new Array<number>(length);
  }

  /** Indicates whether the indicator is primed. */
  public isPrimed(): boolean {
    return this.primed_;
  }

  /** Describes the output data of the indicator. */
  public metadata(): IndicatorMetadata {
    const outputMeta: OutputMetadata = {
      kind: WilliamsPercentROutput.WilliamsPercentRValue,
      type: OutputType.Scalar,
      mnemonic: willrMnemonic,
      description: willrDescription,
    };

    return {
      type: IndicatorType.WilliamsPercentR,
      mnemonic: willrMnemonic,
      description: willrDescription,
      outputs: [outputMeta],
    };
  }

  /** Updates the Williams %R given the next bar's close, high, and low values. */
  public update(close: number, high: number, low: number): number {
    if (isNaN(close) || isNaN(high) || isNaN(low)) {
      return NaN;
    }

    const index = this.circularIndex;
    this.lowCircular[index] = low;
    this.highCircular[index] = high;

    // Advance circular buffer index.
    this.circularIndex++;
    if (this.circularIndex > this.lengthMinOne) {
      this.circularIndex = 0;
    }

    if (this.length > this.circularCount) {
      if (this.lengthMinOne === this.circularCount) {
        // We have exactly `length` samples; compute for the first time.
        let minLow = this.lowCircular[index];
        let maxHigh = this.highCircular[index];
        let idx = index;

        for (let i = 0; i < this.lengthMinOne; i++) {
          // The value of idx is always positive here.
          idx--;
          const tempLow = this.lowCircular[idx];
          if (minLow > tempLow) {
            minLow = tempLow;
          }
          const tempHigh = this.highCircular[idx];
          if (maxHigh < tempHigh) {
            maxHigh = tempHigh;
          }
        }

        if (Math.abs(maxHigh - minLow) < Number.EPSILON) {
          this.value = 0;
        } else {
          this.value = -100 * (maxHigh - close) / (maxHigh - minLow);
        }

        this.primed_ = true;
      }

      this.circularCount++;
      return this.value;
    }

    // Already primed, compute normally with wrapping.
    let minLow = this.lowCircular[index];
    let maxHigh = this.highCircular[index];
    let idx = index;

    for (let i = 0; i < this.lengthMinOne; i++) {
      if (idx === 0) {
        idx = this.lengthMinOne;
      } else {
        idx--;
      }
      const tempLow = this.lowCircular[idx];
      if (minLow > tempLow) {
        minLow = tempLow;
      }
      const tempHigh = this.highCircular[idx];
      if (maxHigh < tempHigh) {
        maxHigh = tempHigh;
      }
    }

    if (Math.abs(maxHigh - minLow) < Number.EPSILON) {
      this.value = 0;
    } else {
      this.value = -100 * (maxHigh - close) / (maxHigh - minLow);
    }

    return this.value;
  }

  /** Updates the Williams %R using a single sample value as a substitute for high, low, and close. */
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
