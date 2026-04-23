import { buildMetadata } from '../../core/build-metadata';
import { Bar } from '../../../entities/bar';
import { Quote } from '../../../entities/quote';
import { Scalar } from '../../../entities/scalar';
import { Trade } from '../../../entities/trade';
import { Indicator } from '../../core/indicator';
import { IndicatorMetadata } from '../../core/indicator-metadata';
import { IndicatorOutput } from '../../core/indicator-output';
import { IndicatorIdentifier } from '../../core/indicator-identifier';

const dmpMnemonic = '+dm';
const dmpDescription = 'Directional Movement Plus';

/**
 * Welles Wilder's Directional Movement Plus indicator.
 *
 * The directional movement was developed in 1978 by Welles Wilder as an indication of trend strength.
 *
 * The calculation of the directional movement (+DM and −DM) is as follows:
 * - UpMove = today's high − yesterday's high
 * - DownMove = yesterday's low − today's low
 * - if UpMove > DownMove and UpMove > 0, then +DM = UpMove, else +DM = 0
 *
 * When the length is greater than 1, Wilder's smoothing method is applied:
 *   Today's +DM(n) = Previous +DM(n) − Previous +DM(n)/n + Today's +DM(1)
 *
 * The indicator is not primed during the first length updates.
 */
export class DirectionalMovementPlus implements Indicator {

  private readonly length_: number;
  private readonly noSmoothing: boolean;
  private count = 0;
  private previousHigh = 0;
  private previousLow = 0;
  private value_ = NaN;
  private accumulator = 0;
  private primed_ = false;

  constructor(length: number) {
    if (length < 1) {
      throw new Error(`invalid length ${length}: must be >= 1`);
    }

    this.length_ = length;
    this.noSmoothing = length === 1;
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
    return buildMetadata(
      IndicatorIdentifier.DirectionalMovementPlus,
      dmpMnemonic,
      dmpDescription,
      [
        { mnemonic: dmpMnemonic, description: dmpDescription },
      ],
    );
  }

  /** Updates the Directional Movement Plus given the next bar's high and low values. */
  public update(high: number, low: number): number {
    if (isNaN(high) || isNaN(low)) {
      return NaN;
    }

    if (high < low) {
      const temp = high;
      high = low;
      low = temp;
    }

    if (this.noSmoothing) {
      if (this.primed_) {
        const deltaPlus = high - this.previousHigh;
        const deltaMinus = this.previousLow - low;

        if (deltaPlus > 0 && deltaPlus > deltaMinus) {
          this.value_ = deltaPlus;
        } else {
          this.value_ = 0;
        }
      } else {
        if (this.count > 0) {
          const deltaPlus = high - this.previousHigh;
          const deltaMinus = this.previousLow - low;

          if (deltaPlus > 0 && deltaPlus > deltaMinus) {
            this.value_ = deltaPlus;
          } else {
            this.value_ = 0;
          }

          this.primed_ = true;
        }

        this.count++;
      }
    } else {
      if (this.primed_) {
        const deltaPlus = high - this.previousHigh;
        const deltaMinus = this.previousLow - low;

        if (deltaPlus > 0 && deltaPlus > deltaMinus) {
          this.accumulator += -this.accumulator / this.length_ + deltaPlus;
        } else {
          this.accumulator += -this.accumulator / this.length_;
        }

        this.value_ = this.accumulator;
      } else {
        if (this.count > 0 && this.length_ >= this.count) {
          const deltaPlus = high - this.previousHigh;
          const deltaMinus = this.previousLow - low;

          if (this.length_ > this.count) {
            if (deltaPlus > 0 && deltaPlus > deltaMinus) {
              this.accumulator += deltaPlus;
            }
          } else {
            if (deltaPlus > 0 && deltaPlus > deltaMinus) {
              this.accumulator += -this.accumulator / this.length_ + deltaPlus;
            } else {
              this.accumulator += -this.accumulator / this.length_;
            }

            this.value_ = this.accumulator;
            this.primed_ = true;
          }
        }

        this.count++;
      }
    }

    this.previousLow = low;
    this.previousHigh = high;

    return this.value_;
  }

  /** Updates the Directional Movement Plus using a single sample value as a substitute for high and low. */
  public updateSample(sample: number): number {
    return this.update(sample, sample);
  }

  /** Updates the indicator given the next scalar sample. */
  public updateScalar(sample: Scalar): IndicatorOutput {
    const v = sample.value;
    const scalar = new Scalar();
    scalar.time = sample.time;
    scalar.value = this.update(v, v);
    return [scalar];
  }

  /** Updates the indicator given the next bar sample. */
  public updateBar(sample: Bar): IndicatorOutput {
    const scalar = new Scalar();
    scalar.time = sample.time;
    scalar.value = this.update(sample.high, sample.low);
    return [scalar];
  }

  /** Updates the indicator given the next quote sample. */
  public updateQuote(sample: Quote): IndicatorOutput {
    const v = (sample.bid + sample.ask) / 2;
    const scalar = new Scalar();
    scalar.time = sample.time;
    scalar.value = this.update(v, v);
    return [scalar];
  }

  /** Updates the indicator given the next trade sample. */
  public updateTrade(sample: Trade): IndicatorOutput {
    const v = sample.price;
    const scalar = new Scalar();
    scalar.time = sample.time;
    scalar.value = this.update(v, v);
    return [scalar];
  }
}
