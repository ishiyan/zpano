import { buildMetadata } from '../../core/build-metadata';
import { Bar } from '../../../entities/bar';
import { Quote } from '../../../entities/quote';
import { Scalar } from '../../../entities/scalar';
import { Trade } from '../../../entities/trade';
import { Indicator } from '../../core/indicator';
import { IndicatorMetadata } from '../../core/indicator-metadata';
import { IndicatorOutput } from '../../core/indicator-output';
import { IndicatorIdentifier } from '../../core/indicator-identifier';
import { AverageDirectionalMovementIndex } from '../average-directional-movement-index/average-directional-movement-index';

const adxrMnemonic = 'adxr';
const adxrDescription = 'Average Directional Movement Index Rating';

/**
 * Welles Wilder's Average Directional Movement Index Rating (ADXR).
 *
 * The average directional movement index rating averages the current ADX value with
 * the ADX value from (length - 1) periods ago. It is calculated as:
 *
 *   ADXR = (ADX[current] + ADX[current - (length - 1)]) / 2
 *
 * The indicator requires close, high, and low values.
 */
export class AverageDirectionalMovementIndexRating implements Indicator {

  private readonly length_: number;
  private readonly bufferSize: number;
  private readonly buffer: number[];
  private bufferIndex = 0;
  private bufferCount = 0;
  private primed_ = false;
  private value_ = NaN;
  private readonly averageDirectionalMovementIndex: AverageDirectionalMovementIndex;

  constructor(length: number) {
    if (length < 1) {
      throw new Error(`invalid length ${length}: must be >= 1`);
    }

    this.length_ = length;
    this.bufferSize = length;
    this.buffer = new Array<number>(length).fill(0);
    this.averageDirectionalMovementIndex = new AverageDirectionalMovementIndex(length);
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
      IndicatorIdentifier.AverageDirectionalMovementIndexRating,
      adxrMnemonic,
      adxrDescription,
      [
        { mnemonic: adxrMnemonic, description: adxrDescription },
        { mnemonic: 'adx', description: 'Average Directional Movement Index' },
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

  /** Updates the Average Directional Movement Index Rating given the next bar's close, high, and low values. */
  public update(close: number, high: number, low: number): number {
    if (isNaN(close) || isNaN(high) || isNaN(low)) {
      return NaN;
    }

    const adxValue = this.averageDirectionalMovementIndex.update(close, high, low);

    if (!this.averageDirectionalMovementIndex.isPrimed()) {
      return NaN;
    }

    // Store ADX value in circular buffer.
    this.buffer[this.bufferIndex] = adxValue;
    this.bufferIndex = (this.bufferIndex + 1) % this.bufferSize;
    this.bufferCount++;

    if (this.bufferCount < this.bufferSize) {
      return NaN;
    }

    // The oldest value in the buffer is at bufferIndex (since we just advanced it).
    const oldADX = this.buffer[this.bufferIndex % this.bufferSize];
    this.value_ = (adxValue + oldADX) / 2;
    this.primed_ = true;

    return this.value_;
  }

  /** Updates the Average Directional Movement Index Rating using a single sample value as a substitute for close, high, and low. */
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
