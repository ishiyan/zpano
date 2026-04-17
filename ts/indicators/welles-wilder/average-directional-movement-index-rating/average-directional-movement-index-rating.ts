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
import { AverageDirectionalMovementIndex } from '../average-directional-movement-index/average-directional-movement-index';
import { AverageDirectionalMovementIndexRatingOutput } from './average-directional-movement-index-rating-output';

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
    const outputs: OutputMetadata[] = [
      {
        kind: AverageDirectionalMovementIndexRatingOutput.AverageDirectionalMovementIndexRatingValue,
        type: OutputType.Scalar,
        mnemonic: adxrMnemonic,
        description: adxrDescription,
      },
      {
        kind: AverageDirectionalMovementIndexRatingOutput.AverageDirectionalMovementIndexValue,
        type: OutputType.Scalar,
        mnemonic: 'adx',
        description: 'Average Directional Movement Index',
      },
      {
        kind: AverageDirectionalMovementIndexRatingOutput.DirectionalMovementIndexValue,
        type: OutputType.Scalar,
        mnemonic: 'dx',
        description: 'Directional Movement Index',
      },
      {
        kind: AverageDirectionalMovementIndexRatingOutput.DirectionalIndicatorPlusValue,
        type: OutputType.Scalar,
        mnemonic: '+di',
        description: 'Directional Indicator Plus',
      },
      {
        kind: AverageDirectionalMovementIndexRatingOutput.DirectionalIndicatorMinusValue,
        type: OutputType.Scalar,
        mnemonic: '-di',
        description: 'Directional Indicator Minus',
      },
      {
        kind: AverageDirectionalMovementIndexRatingOutput.DirectionalMovementPlusValue,
        type: OutputType.Scalar,
        mnemonic: '+dm',
        description: 'Directional Movement Plus',
      },
      {
        kind: AverageDirectionalMovementIndexRatingOutput.DirectionalMovementMinusValue,
        type: OutputType.Scalar,
        mnemonic: '-dm',
        description: 'Directional Movement Minus',
      },
      {
        kind: AverageDirectionalMovementIndexRatingOutput.AverageTrueRangeValue,
        type: OutputType.Scalar,
        mnemonic: 'atr',
        description: 'Average True Range',
      },
      {
        kind: AverageDirectionalMovementIndexRatingOutput.TrueRangeValue,
        type: OutputType.Scalar,
        mnemonic: 'tr',
        description: 'True Range',
      },
    ];

    return {
      type: IndicatorType.AverageDirectionalMovementIndexRating,
      mnemonic: adxrMnemonic,
      description: adxrDescription,
      outputs,
    };
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
