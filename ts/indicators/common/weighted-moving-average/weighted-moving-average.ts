import { componentTripleMnemonic } from '../../core/component-triple-mnemonic';
import { IndicatorMetadata } from '../../core/indicator-metadata';
import { IndicatorType } from '../../core/indicator-type';
import { LineIndicator } from '../../core/line-indicator';
import { OutputType } from '../../core/outputs/output-type';
import { WeightedMovingAverageOutput } from './weighted-moving-average-output';
import { WeightedMovingAverageParams } from './weighted-moving-average-params';

/** Function to calculate mnemonic of a __WeightedMovingAverage__ indicator. */
export const weightedMovingAverageMnemonic = (params: WeightedMovingAverageParams): string =>
  'wma('.concat(params.length.toString(), componentTripleMnemonic(params.barComponent, params.quoteComponent, params.tradeComponent), ')');

/** Weighted Moving Average line indicator.
 *
 * Computes the weighted moving average (WMA) that has multiplying factors
 * to give arithmetically decreasing weights to the samples in the look back window.
 *
 *    WMAᵢ = (ℓPᵢ + (ℓ-1)Pᵢ₋₁ + ... + Pᵢ₋ℓ) / (ℓ + (ℓ-1) + ... + 2 + 1),
 *
 * where ℓ is the length.
 *
 * The denominator is a triangle number and can be computed as
 *
 *    ½ℓ(ℓ+1).
 *
 * When calculating the WMA across successive values,
 *
 *    WMAᵢ₊₁ - WMAᵢ = ℓPᵢ₊₁ - Pᵢ - ... - Pᵢ₋ℓ₊₁
 *
 * If we denote the sum
 *
 *    Totalᵢ = Pᵢ + ... + Pᵢ₋ℓ₊₁
 *
 * then
 *
 *    Totalᵢ₊₁ = Totalᵢ + Pᵢ₊₁ - Pᵢ₋ℓ₊₁
 *
 *    Numeratorᵢ₊₁ = Numeratorᵢ + ℓPᵢ₊₁ - Totalᵢ
 *
 *    WMAᵢ₊₁ = Numeratorᵢ₊₁ / ½ℓ(ℓ+1)
 *
 * The WMA indicator is not primed during the first ℓ-1 updates.
 */
export class WeightedMovingAverage extends LineIndicator {
  private window: Array<number>;
  private windowLength: number;
  private windowSum = 0;
  private windowSub = 0;
  private divider: number;
  private windowCount = 0;
  private lastIndex: number;

  /**
   * Constructs an instance given a length in samples.
   * The length should be an integer greater than 1.
   **/
  public constructor(params: WeightedMovingAverageParams){
    super();
    const length = Math.floor(params.length);
    if (length < 2) {
      throw new Error('length should be greater than 1');
    }

    this.mnemonic = weightedMovingAverageMnemonic(params);
    this.description = 'Weighted moving average ' + this.mnemonic;
    this.barComponent = params.barComponent;
    this.quoteComponent = params.quoteComponent;
    this.tradeComponent = params.tradeComponent;
    this.window = new Array<number>(length);
    this.windowLength = length;
    this.lastIndex = length - 1;
    this.divider = length * (length +1) / 2;
    this.primed = false;
  }

  /** Describes the output data of the indicator. */
  public metadata(): IndicatorMetadata {
    return {
      type: IndicatorType.WeightedMovingAverage,
      mnemonic: this.mnemonic,
      description: this.description,
      outputs: [{
        kind: WeightedMovingAverageOutput.WeightedMovingAverageValue,
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
      this.windowSum += sample * this.windowLength - this.windowSub;
      this.windowSub += sample - this.window[0];

      for (let i = 0; i < this.lastIndex; i++) {
        this.window[i] = this.window[i+1];
      }

      this.window[this.lastIndex] = sample;
    } else { // Not primed.
      this.window[this.windowCount] = sample;
      this.windowSub += sample;
      ++this.windowCount;
      this.windowSum += sample * this.windowCount;

      if (this.windowLength > this.windowCount) {
        return Number.NaN;
      }

      this.primed = true;
    }

    return this.windowSum / this.divider;
  }
}
