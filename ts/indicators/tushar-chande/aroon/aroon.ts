import { buildMetadata } from '../../core/build-metadata';
import { Bar } from '../../../entities/bar';
import { Quote } from '../../../entities/quote';
import { Scalar } from '../../../entities/scalar';
import { Trade } from '../../../entities/trade';
import { Indicator } from '../../core/indicator';
import { IndicatorMetadata } from '../../core/indicator-metadata';
import { IndicatorOutput } from '../../core/indicator-output';
import { IndicatorIdentifier } from '../../core/indicator-identifier';
import { AroonParams } from './params';

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
  private bufferIndex = 0;
  private count = 0;

  private highestIndex = 0;
  private lowestIndex = 0;

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
    return buildMetadata(
      IndicatorIdentifier.Aroon,
      this.mnemonic_,
      this.description_,
      [
        { mnemonic: this.mnemonic_ + ' up', description: this.description_ + ' Up' },
        { mnemonic: this.mnemonic_ + ' down', description: this.description_ + ' Down' },
        { mnemonic: this.mnemonic_ + ' osc', description: this.description_ + ' Oscillator' },
      ],
    );
  }

  /** Updates the indicator given the next bar's high and low values. Returns [AroonUp, AroonDown, AroonOsc]. */
  public update(high: number, low: number): [number, number, number] {
    if (isNaN(high) || isNaN(low)) {
      return [NaN, NaN, NaN];
    }

    const windowSize = this.length_ + 1;
    const today = this.count;

    // Store in circular buffer.
    const pos = this.bufferIndex;
    this.highBuf[pos] = high;
    this.lowBuf[pos] = low;
    this.bufferIndex = (this.bufferIndex + 1) % windowSize;
    this.count++;

    // Need at least length+1 bars.
    if (this.count < windowSize) {
      return [this.up_, this.down_, this.osc_];
    }

    const trailingIndex = today - this.length_;

    if (this.count === windowSize) {
      // First time: scan entire window.
      this.highestIndex = trailingIndex;
      this.lowestIndex = trailingIndex;

      for (let i = trailingIndex + 1; i <= today; i++) {
        const bufPos = i % windowSize;

        if (this.highBuf[bufPos] >= this.highBuf[this.highestIndex % windowSize]) {
          this.highestIndex = i;
        }

        if (this.lowBuf[bufPos] <= this.lowBuf[this.lowestIndex % windowSize]) {
          this.lowestIndex = i;
        }
      }
    } else {
      // Subsequent: optimized update.
      if (this.highestIndex < trailingIndex) {
        this.highestIndex = trailingIndex;

        for (let i = trailingIndex + 1; i <= today; i++) {
          const bufPos = i % windowSize;
          if (this.highBuf[bufPos] >= this.highBuf[this.highestIndex % windowSize]) {
            this.highestIndex = i;
          }
        }
      } else if (high >= this.highBuf[this.highestIndex % windowSize]) {
        this.highestIndex = today;
      }

      if (this.lowestIndex < trailingIndex) {
        this.lowestIndex = trailingIndex;

        for (let i = trailingIndex + 1; i <= today; i++) {
          const bufPos = i % windowSize;
          if (this.lowBuf[bufPos] <= this.lowBuf[this.lowestIndex % windowSize]) {
            this.lowestIndex = i;
          }
        }
      } else if (low <= this.lowBuf[this.lowestIndex % windowSize]) {
        this.lowestIndex = today;
      }
    }

    this.up_ = this.factor * (this.length_ - (today - this.highestIndex));
    this.down_ = this.factor * (this.length_ - (today - this.lowestIndex));
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
