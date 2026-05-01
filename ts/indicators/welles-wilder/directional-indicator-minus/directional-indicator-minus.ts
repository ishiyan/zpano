import { buildMetadata } from '../../core/build-metadata';
import { Bar } from '../../../entities/bar';
import { Quote } from '../../../entities/quote';
import { Scalar } from '../../../entities/scalar';
import { Trade } from '../../../entities/trade';
import { Indicator } from '../../core/indicator';
import { IndicatorMetadata } from '../../core/indicator-metadata';
import { IndicatorOutput } from '../../core/indicator-output';
import { IndicatorIdentifier } from '../../core/indicator-identifier';
import { AverageTrueRange } from '../average-true-range/average-true-range';
import { DirectionalMovementMinus } from '../directional-movement-minus/directional-movement-minus';

const dimMnemonic = '-di';
const dimDescription = 'Directional Indicator Minus';
const epsilon = 1e-8;

/**
 * Welles Wilder's Directional Indicator Minus (-DI).
 *
 * The directional indicator minus measures the percentage of the average true range
 * that is attributable to downward movement. It is calculated as:
 *
 *   -DI = 100 * -DM(n) / (ATR * length)
 *
 * where -DM(n) is the Wilder-smoothed directional movement minus and ATR is the
 * average true range over the same length.
 *
 * The indicator requires close, high, and low values.
 */
export class DirectionalIndicatorMinus implements Indicator {

  private readonly length_: number;
  private value_ = NaN;
  private readonly averageTrueRange: AverageTrueRange;
  private readonly directionalMovementMinus: DirectionalMovementMinus;

  constructor(length: number) {
    if (length < 1) {
      throw new Error(`invalid length ${length}: must be >= 1`);
    }

    this.length_ = length;
    this.averageTrueRange = new AverageTrueRange(length);
    this.directionalMovementMinus = new DirectionalMovementMinus(length);
  }

  /** Returns the length parameter. */
  public get length(): number {
    return this.length_;
  }

  /** Indicates whether the indicator is primed. */
  public isPrimed(): boolean {
    return this.averageTrueRange.isPrimed() && this.directionalMovementMinus.isPrimed();
  }

  /** Describes the output data of the indicator. */
  public metadata(): IndicatorMetadata {
    return buildMetadata(
      IndicatorIdentifier.DirectionalIndicatorMinus,
      dimMnemonic,
      dimDescription,
      [
        { mnemonic: dimMnemonic, description: dimDescription },
        { mnemonic: '-dm', description: 'Directional Movement Minus' },
        { mnemonic: 'atr', description: 'Average True Range' },
        { mnemonic: 'tr', description: 'True Range' },
      ],
    );
  }

  /** Updates the Directional Indicator Minus given the next bar's close, high, and low values. */
  public update(close: number, high: number, low: number): number {
    if (isNaN(close) || isNaN(high) || isNaN(low)) {
      return NaN;
    }

    const atrValue = this.averageTrueRange.update(close, high, low);
    const dmmValue = this.directionalMovementMinus.update(high, low);

    if (this.averageTrueRange.isPrimed() && this.directionalMovementMinus.isPrimed()) {
      const atrScaled = atrValue * this.length_;

      if (Math.abs(atrScaled) < epsilon) {
        this.value_ = 0;
      } else {
        this.value_ = 100 * dmmValue / atrScaled;
      }

      return this.value_;
    }

    return NaN;
  }

  /** Updates the Directional Indicator Minus using a single sample value as a substitute for close, high, and low. */
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
    const v = (sample.bidPrice + sample.askPrice) / 2;
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
