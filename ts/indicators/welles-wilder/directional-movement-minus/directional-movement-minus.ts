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
import { DirectionalMovementMinusOutput } from './directional-movement-minus-output';

const dmmMnemonic = '-dm';
const dmmDescription = 'Directional Movement Minus';

/**
 * Welles Wilder's Directional Movement Minus indicator.
 *
 * The directional movement was developed in 1978 by Welles Wilder as an indication of trend strength.
 *
 * The calculation of the directional movement (+DM and −DM) is as follows:
 * - UpMove = today's high − yesterday's high
 * - DownMove = yesterday's low − today's low
 * - if DownMove > UpMove and DownMove > 0, then −DM = DownMove, else −DM = 0
 *
 * When the length is greater than 1, Wilder's smoothing method is applied:
 *   Today's −DM(n) = Previous −DM(n) − Previous −DM(n)/n + Today's −DM(1)
 *
 * The indicator is not primed during the first length updates.
 */
export class DirectionalMovementMinus implements Indicator {

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
    const outputMeta: OutputMetadata = {
      kind: DirectionalMovementMinusOutput.DirectionalMovementMinusValue,
      type: OutputType.Scalar,
      mnemonic: dmmMnemonic,
      description: dmmDescription,
    };

    return {
      type: IndicatorType.DirectionalMovementMinus,
      mnemonic: dmmMnemonic,
      description: dmmDescription,
      outputs: [outputMeta],
    };
  }

  /** Updates the Directional Movement Minus given the next bar's high and low values. */
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
        const deltaMinus = this.previousLow - low;
        const deltaPlus = high - this.previousHigh;

        if (deltaMinus > 0 && deltaPlus < deltaMinus) {
          this.value_ = deltaMinus;
        } else {
          this.value_ = 0;
        }
      } else {
        if (this.count > 0) {
          const deltaMinus = this.previousLow - low;
          const deltaPlus = high - this.previousHigh;

          if (deltaMinus > 0 && deltaPlus < deltaMinus) {
            this.value_ = deltaMinus;
          } else {
            this.value_ = 0;
          }

          this.primed_ = true;
        }

        this.count++;
      }
    } else {
      if (this.primed_) {
        const deltaMinus = this.previousLow - low;
        const deltaPlus = high - this.previousHigh;

        if (deltaMinus > 0 && deltaPlus < deltaMinus) {
          this.accumulator += -this.accumulator / this.length_ + deltaMinus;
        } else {
          this.accumulator += -this.accumulator / this.length_;
        }

        this.value_ = this.accumulator;
      } else {
        if (this.count > 0 && this.length_ >= this.count) {
          const deltaMinus = this.previousLow - low;
          const deltaPlus = high - this.previousHigh;

          if (this.length_ > this.count) {
            if (deltaMinus > 0 && deltaPlus < deltaMinus) {
              this.accumulator += deltaMinus;
            }
          } else {
            if (deltaMinus > 0 && deltaPlus < deltaMinus) {
              this.accumulator += -this.accumulator / this.length_ + deltaMinus;
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

  /** Updates the Directional Movement Minus using a single sample value as a substitute for high and low. */
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
