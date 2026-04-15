import { componentTripleMnemonic } from '../../core/component-triple-mnemonic';
import { IndicatorMetadata } from '../../core/indicator-metadata';
import { IndicatorType } from '../../core/indicator-type';
import { LineIndicator } from '../../core/line-indicator';
import { OutputType } from '../../core/outputs/output-type';
import { ExponentialMovingAverageOutput } from './exponential-moving-average-output';
import { ExponentialMovingAverageLengthParams } from './exponential-moving-average-params';
import { ExponentialMovingAverageSmoothingFactorParams } from './exponential-moving-average-params';

const guardLength = (object: any): object is ExponentialMovingAverageLengthParams => 'length' in object;

/** Function to calculate mnemonic of an __ExponentialMovingAverage__ indicator. */
export const exponentialMovingAverageMnemonic =
  (params: ExponentialMovingAverageLengthParams | ExponentialMovingAverageSmoothingFactorParams): string => {
  if (guardLength(params)) {
    const p = params as ExponentialMovingAverageLengthParams;
    return 'ema('.concat(Math.floor(p.length).toString(),
      componentTripleMnemonic(p.barComponent, p.quoteComponent, p.tradeComponent), ')');
  } else {
    const p = params as ExponentialMovingAverageSmoothingFactorParams;
    const length = Math.round(2 / p.smoothingFactor) - 1;
    return 'ema('.concat(length.toString(), ', ', p.smoothingFactor.toFixed(8),
      componentTripleMnemonic(p.barComponent, p.quoteComponent, p.tradeComponent), ')');
  }
};

/** __Exponential Moving Average__ line indicator computes the exponential, or exponentially weighted, moving average (_EMA_).
 *
 * Given a constant smoothing percentage factor 0 < α ≤ 1, _EMA_ is calculated by applying a constant
 * smoothing factor α to a difference of today's sample and yesterday's _EMA_ value:
 *
 *    EMAᵢ = αPᵢ + (1-α)EMAᵢ₋₁ = EMAᵢ₋₁ + α(Pᵢ - EMAᵢ₋₁), 0 < α ≤ 1.
 *
 * Thus, the weighting for each older sample is given by the geometric progression 1, α, α², α³, …,
 * giving much more importance to recent observations while not discarding older ones: all data
 * previously used are always part of the new _EMA_ value.
 *
 * α may be expressed as a percentage, so a smoothing factor of 10% is equivalent to α = 0.1. A higher α
 * discounts older observations faster. Alternatively, α may be expressed in terms of ℓ time periods (length),
 * where:
 *
 *    α = 2 / (ℓ + 1) and ℓ = 2/α - 1.
 *
 * The indicator is not primed during the first ℓ-1 updates.
 *
 * The 12- and 26-day EMAs are the most popular short-term averages, and they are used to create indicators
 * like MACD and PPO. In general, the 50- and 200-day EMAs are used as signals of long-term trends.
 *
 * The very first EMA value (the seed for subsequent values) is calculated differently.
 * This implementation, when using a length as an input parameter, allows for two algorithms for this seed.
 *
 * ❶ Use a simple average of the first 'period'. This is the most widely documented approach.
 *
 * ❷ Use first sample value as a seed. This is used in Metastock.
 */
export class ExponentialMovingAverage extends LineIndicator {
  private length = 0;
  private sum = 0;
  private count = 0;
  private value = 0;
  private smoothingFactor: number;
  private firstIsAverage = false;

  /**
   * Constructs an instance given a length in samples or a smoothing factor in (0, 1).
   **/
  public constructor(params: ExponentialMovingAverageLengthParams | ExponentialMovingAverageSmoothingFactorParams){
    super();
    let len;
    if (guardLength(params)) {
      const p = params as ExponentialMovingAverageLengthParams;
      len = Math.floor(p.length);
      if (len < 2) {
        throw new Error('length should be greater than 1');
      }

      this.length = len;
      this.smoothingFactor = 2 / (len + 1);
      this.firstIsAverage = p.firstIsAverage ?? false;

    } else {
      const p = params as ExponentialMovingAverageSmoothingFactorParams;
      if (p.smoothingFactor <= 0 || p.smoothingFactor >= 1) {
        throw new Error('smoothing factor should be in range (0, 1)');
      }

      this.smoothingFactor = p.smoothingFactor;
      this.length = Math.round(2 / p.smoothingFactor) - 1;
      this.firstIsAverage = p.firstIsAverage ?? false;
    }

    this.mnemonic = exponentialMovingAverageMnemonic(params);
    this.description = 'Exponential moving average ' + this.mnemonic;
    this.barComponent = params.barComponent;
    this.quoteComponent = params.quoteComponent;
    this.tradeComponent = params.tradeComponent;
    this.primed = false;
  }

  /** Describes the output data of the indicator. */
  public metadata(): IndicatorMetadata {
    return {
      type: IndicatorType.ExponentialMovingAverage,
      mnemonic: this.mnemonic,
      description: this.description,
      outputs: [{
        kind: ExponentialMovingAverageOutput.ExponentialMovingAverageValue,
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
      this.value += (sample - this.value) * this.smoothingFactor;
    } else { // Not primed.
      this.count++;
      if (this.firstIsAverage) {
        this.sum += sample;
        if (this.count < this.length) {
          return Number.NaN;
        }

        this.value = this.sum / this.length;
      } else {
        if (this.count === 1) {
          this.value = sample;
        } else {
          this.value += (sample - this.value) * this.smoothingFactor;
        }

        if (this.count < this.length) {
          return Number.NaN;
        }
      }

      this.primed = true;
    }

    return this.value;
  }
}
