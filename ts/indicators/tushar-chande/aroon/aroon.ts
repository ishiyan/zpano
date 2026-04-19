import { Bar } from '../../../entities/bar';
import { Quote } from '../../../entities/quote';
import { Scalar } from '../../../entities/scalar';
import { Trade } from '../../../entities/trade';
import { Indicator } from '../../core/indicator';
import { IndicatorMetadata } from '../../core/indicator-metadata';
import { IndicatorOutput } from '../../core/indicator-output';
import { IndicatorType } from '../../core/indicator-type';
import { OutputType } from '../../core/outputs/output-type';
import { AroonOutput } from './aroon-output';
import { AroonParams } from './aroon-params';

/** Function to calculate mnemonic of an __Aroon__ indicator. */
export const aroonMnemonic = (params: AroonParams): string => `aroon(${params.length})`;

/**
 * Tushar Chande's Aroon indicator.
 *
 * The Aroon indicator measures the number of periods since the highest high
 * and lowest low within a lookback window. It produces three outputs:
 *   - AroonUp: 100 * (Length - periods since highest high) / Length
 *   - AroonDown: 100 * (Length - periods since lowest low) / Length
 *   - AroonOsc: AroonUp - AroonDown
 *
 * The indicator requires bar data (high, low). For scalar, quote, and
 * trade updates, the single value substitutes for both.
 *
 * Reference:
 *
 * Chande, Tushar S. (1995). "The New Technical Trader". John Wiley & Sons.
 */
export class Aroon implements Indicator {

  private readonly length_: number;
  private readonly factor: number;

  private readonly highBuf: Float64Array;
  private readonly lowBuf: Float64Array;
  private bufIdx = 0;
  private count = 0;

  private highestIdx = 0;
  private lowestIdx = 0;

  private up_ = NaN;
  private down_ = NaN;
  private osc_ = NaN;
  private primed_ = false;

  private readonly mnemonic_: string;
  private readonly description_: string;

  constructor(params: AroonParams) {
    const length = Math.floor(params.length);

    if (length < 2) {
      throw new Error('length should be greater than 1');
    }

    this.length_ = length;
    this.factor = 100.0 / length;

    const windowSize = length + 1;
    this.highBuf = new Float64Array(windowSize);
    this.lowBuf = new Float64Array(windowSize);

    this.mnemonic_ = aroonMnemonic(params);
    this.description_ = 'Aroon ' + this.mnemonic_;
  }

  /** Indicates whether the indicator is primed. */
  public isPrimed(): boolean {
    return this.primed_;
  }

  /** Describes the output data of the indicator. */
  public metadata(): IndicatorMetadata {
    return {
      type: IndicatorType.Aroon,
      mnemonic: this.mnemonic_,
      description: this.description_,
      outputs: [
        {
          kind: AroonOutput.Up,
          type: OutputType.Scalar,
          mnemonic: this.mnemonic_ + ' up',
          description: this.description_ + ' Up',
        },
        {
          kind: AroonOutput.Down,
          type: OutputType.Scalar,
          mnemonic: this.mnemonic_ + ' down',
          description: this.description_ + ' Down',
        },
        {
          kind: AroonOutput.Osc,
          type: OutputType.Scalar,
          mnemonic: this.mnemonic_ + ' osc',
          description: this.description_ + ' Oscillator',
        },
      ],
    };
  }

  /** Updates the indicator given the next bar's high and low values. Returns [AroonUp, AroonDown, AroonOsc]. */
  public update(high: number, low: number): [number, number, number] {
    if (isNaN(high) || isNaN(low)) {
      return [NaN, NaN, NaN];
    }

    const windowSize = this.length_ + 1;
    const today = this.count;

    // Store in circular buffer.
    const pos = this.bufIdx;
    this.highBuf[pos] = high;
    this.lowBuf[pos] = low;
    this.bufIdx = (this.bufIdx + 1) % windowSize;
    this.count++;

    // Need at least length+1 bars.
    if (this.count < windowSize) {
      return [this.up_, this.down_, this.osc_];
    }

    const trailingIdx = today - this.length_;

    if (this.count === windowSize) {
      // First time: scan entire window.
      this.highestIdx = trailingIdx;
      this.lowestIdx = trailingIdx;

      for (let i = trailingIdx + 1; i <= today; i++) {
        const bufPos = i % windowSize;

        if (this.highBuf[bufPos] >= this.highBuf[this.highestIdx % windowSize]) {
          this.highestIdx = i;
        }

        if (this.lowBuf[bufPos] <= this.lowBuf[this.lowestIdx % windowSize]) {
          this.lowestIdx = i;
        }
      }
    } else {
      // Subsequent: optimized update.
      if (this.highestIdx < trailingIdx) {
        this.highestIdx = trailingIdx;

        for (let i = trailingIdx + 1; i <= today; i++) {
          const bufPos = i % windowSize;
          if (this.highBuf[bufPos] >= this.highBuf[this.highestIdx % windowSize]) {
            this.highestIdx = i;
          }
        }
      } else if (high >= this.highBuf[this.highestIdx % windowSize]) {
        this.highestIdx = today;
      }

      if (this.lowestIdx < trailingIdx) {
        this.lowestIdx = trailingIdx;

        for (let i = trailingIdx + 1; i <= today; i++) {
          const bufPos = i % windowSize;
          if (this.lowBuf[bufPos] <= this.lowBuf[this.lowestIdx % windowSize]) {
            this.lowestIdx = i;
          }
        }
      } else if (low <= this.lowBuf[this.lowestIdx % windowSize]) {
        this.lowestIdx = today;
      }
    }

    this.up_ = this.factor * (this.length_ - (today - this.highestIdx));
    this.down_ = this.factor * (this.length_ - (today - this.lowestIdx));
    this.osc_ = this.up_ - this.down_;

    if (!this.primed_) {
      this.primed_ = true;
    }

    return [this.up_, this.down_, this.osc_];
  }

  /** Updates the indicator given the next scalar sample. */
  public updateScalar(sample: Scalar): IndicatorOutput {
    const v = sample.value;
    const [up, down, osc] = this.update(v, v);
    const s1 = new Scalar();
    s1.time = sample.time;
    s1.value = up;
    const s2 = new Scalar();
    s2.time = sample.time;
    s2.value = down;
    const s3 = new Scalar();
    s3.time = sample.time;
    s3.value = osc;
    return [s1, s2, s3];
  }

  /** Updates the indicator given the next bar sample. */
  public updateBar(sample: Bar): IndicatorOutput {
    const [up, down, osc] = this.update(sample.high, sample.low);
    const s1 = new Scalar();
    s1.time = sample.time;
    s1.value = up;
    const s2 = new Scalar();
    s2.time = sample.time;
    s2.value = down;
    const s3 = new Scalar();
    s3.time = sample.time;
    s3.value = osc;
    return [s1, s2, s3];
  }

  /** Updates the indicator given the next quote sample. */
  public updateQuote(sample: Quote): IndicatorOutput {
    const v = (sample.bid + sample.ask) / 2;
    const scalar = new Scalar();
    scalar.time = sample.time;
    scalar.value = v;
    return this.updateScalar(scalar);
  }

  /** Updates the indicator given the next trade sample. */
  public updateTrade(sample: Trade): IndicatorOutput {
    const scalar = new Scalar();
    scalar.time = sample.time;
    scalar.value = sample.price;
    return this.updateScalar(scalar);
  }
}
