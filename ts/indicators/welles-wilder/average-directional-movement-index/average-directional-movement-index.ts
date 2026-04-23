import { buildMetadata } from '../../core/build-metadata';
import { Bar } from '../../../entities/bar';
import { Quote } from '../../../entities/quote';
import { Scalar } from '../../../entities/scalar';
import { Trade } from '../../../entities/trade';
import { Indicator } from '../../core/indicator';
import { IndicatorMetadata } from '../../core/indicator-metadata';
import { IndicatorOutput } from '../../core/indicator-output';
import { IndicatorIdentifier } from '../../core/indicator-identifier';
import { DirectionalMovementIndex } from '../directional-movement-index/directional-movement-index';

const adxMnemonic = 'adx';
const adxDescription = 'Average Directional Movement Index';

/**
 * Welles Wilder's Average Directional Movement Index (ADX).
 *
 * The average directional movement index smooths the directional movement index (DX)
 * using Wilder's smoothing technique. It is calculated as:
 *
 *   Initial ADX = SMA of first `length` DX values
 *   Subsequent ADX = (previousADX * (length-1) + DX) / length
 *
 * The indicator requires close, high, and low values.
 */
export class AverageDirectionalMovementIndex implements Indicator {

  private readonly length_: number;
  private readonly lengthMinusOne: number;
  private count = 0;
  private sum = 0;
  private primed_ = false;
  private value_ = NaN;
  private readonly directionalMovementIndex: DirectionalMovementIndex;

  constructor(length: number) {
    if (length < 1) {
      throw new Error(`invalid length ${length}: must be >= 1`);
    }

    this.length_ = length;
    this.lengthMinusOne = length - 1;
    this.directionalMovementIndex = new DirectionalMovementIndex(length);
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
      IndicatorIdentifier.AverageDirectionalMovementIndex,
      adxMnemonic,
      adxDescription,
      [
        { mnemonic: adxMnemonic, description: adxDescription },
        { mnemonic: 'dx', description: 'Directional Movement Index' },
        { mnemonic: '+di', description: 'Directional Indicator Plus' },
        { mnemonic: '-di', description: 'Directional Indicator Minus' },
        { mnemonic: '+dm', description: 'Directional Movement Plus' },
        { mnemonic: '-dm', description: 'Directional Movement Minus' },
        { mnemonic: 'atr', description: 'Average True Range' },
        { mnemonic: 'tr', description: 'True Range' },
      ],
    );
  }

  /** Updates the Average Directional Movement Index given the next bar's close, high, and low values. */
  public update(close: number, high: number, low: number): number {
    if (isNaN(close) || isNaN(high) || isNaN(low)) {
      return NaN;
    }

    const dxValue = this.directionalMovementIndex.update(close, high, low);

    if (!this.directionalMovementIndex.isPrimed()) {
      return NaN;
    }

    if (this.primed_) {
      this.value_ = (this.value_ * this.lengthMinusOne + dxValue) / this.length_;
      return this.value_;
    }

    this.count++;
    this.sum += dxValue;

    if (this.count === this.length_) {
      this.value_ = this.sum / this.length_;
      this.primed_ = true;
      return this.value_;
    }

    return NaN;
  }

  /** Updates the Average Directional Movement Index using a single sample value as a substitute for close, high, and low. */
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
