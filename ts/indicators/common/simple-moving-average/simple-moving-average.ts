import { componentTripleMnemonic } from '../../core/component-triple-mnemonic';
import { IndicatorMetadata } from '../../core/indicator-metadata';
import { IndicatorType } from '../../core/indicator-type';
import { LineIndicator } from '../../core/line-indicator';
import { OutputType } from '../../core/outputs/output-type';
import { SimpleMovingAverageOutput } from './simple-moving-average-output';
import { SimpleMovingAverageParams } from './simple-moving-average-params';

/** Function to calculate mnemonic of a __SimpleMovingAverage__ indicator. */
export const simpleMovingAverageMnemonic = (params: SimpleMovingAverageParams): string =>
  'sma('.concat(params.length.toString(), componentTripleMnemonic(params.barComponent, params.quoteComponent, params.tradeComponent), ')');

/** Simple Moving Average line indicator. */
export class SimpleMovingAverage extends LineIndicator {
  private window: Array<number>;
  private windowLength: number;
  private windowSum: number;
  private windowCount: number;
  private lastIndex: number;

  /**
   * Constructs an instance given a length in samples.
   * The length should be an integer greater than 1.
   **/
  public constructor(params: SimpleMovingAverageParams){
    super();
    const length = Math.floor(params.length);
    if (length < 2) {
      throw new Error('length should be greater than 1');
    }

    this.mnemonic = simpleMovingAverageMnemonic(params);
    this.description = 'Simple moving average ' + this.mnemonic;
    this.barComponent = params.barComponent;
    this.quoteComponent = params.quoteComponent;
    this.tradeComponent = params.tradeComponent;
    this.window = new Array<number>(length);
    this.windowLength = length;
    this.windowSum = 0;
    this.windowCount = 0;
    this.lastIndex = length - 1;
    this.primed = false;
  }

  /** Describes the output data of the indicator. */
  public metadata(): IndicatorMetadata {
    return {
      type: IndicatorType.SimpleMovingAverage,
      mnemonic: this.mnemonic,
      description: this.description,
      outputs: [{
        kind: SimpleMovingAverageOutput.SimpleMovingAverageValue,
        type: OutputType.Scalar,
        mnemonic: this.mnemonic,
        description: this.description,
      }],
    };
  }

  /** Updates the value of the indicator given the next sample. */
  public update(sample: number): number {
    if (Number.isNaN(sample)) {
      return sample;
    }

    if (this.primed) {
      this.windowSum += sample - this.window[0];
      for (let i = 0; i < this.lastIndex; i++) {
        this.window[i] = this.window[i+1];
      }

      this.window[this.lastIndex] = sample;
    } else {
      this.windowSum += sample;
      this.window[this.windowCount] = sample;
      this.windowCount++;

      if (this.windowLength > this.windowCount) {
        return Number.NaN;
      }

      this.primed = true;
    }

    return this.windowSum / this.windowLength;
  }
}
