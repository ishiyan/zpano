import { componentTripleMnemonic } from '../../core/component-triple-mnemonic';
import { IndicatorMetadata } from '../../core/indicator-metadata';
import { IndicatorType } from '../../core/indicator-type';
import { LineIndicator } from '../../core/line-indicator';
import { OutputType } from '../../core/outputs/output-type';
import { VarianceOutput } from './variance-output';
import { VarianceParams } from './variance-params';

/** Function to calculate mnemonic of a __Variance__ indicator. */
export const varianceMnemonic = (params: VarianceParams): string =>
  'var.'.concat(params.unbiased ? 's(' : 'p(', params.length.toString(),
  componentTripleMnemonic(params.barComponent, params.quoteComponent, params.tradeComponent), ')');

/** Variance line indicator.
 *
 * Variance computes the variance of the samples within a moving window of length ℓ:
 *
 *     σ² = (∑xᵢ² - (∑xᵢ)²/ℓ)/ℓ
 *
 * for the estimation of the population variance, or as:
 *
 *     σ² = (∑xᵢ² - (∑xᵢ)²/ℓ)/(ℓ-1)
 *
 * for the unbiased estimation of the sample variance, i={0,…,ℓ-1}.
 */
export class Variance extends LineIndicator {
  private window: Array<number>;
  private windowLength: number;
  private windowSum = 0;
  private windowSquaredSum = 0;
  private windowCount = 0;
  private lastIndex: number;
  private unbiased: boolean;

  /**
   * Constructs an instance given a length in samples.
   * The length should be an integer greater than 1.
   **/
  public constructor(params: VarianceParams){
    super();
    const length = Math.floor(params.length);
    if (length < 2) {
      throw new Error('length should be greater than 1');
    }

    this.unbiased = params.unbiased;
    this.mnemonic = varianceMnemonic(params);
    if (params.unbiased) {
      this.description = 'Unbiased estimation of the sample variance ' + this.mnemonic;
    } else {
      this.description = 'Estimation of the population variance ' + this.mnemonic;
    }

    this.barComponent = params.barComponent;
    this.quoteComponent = params.quoteComponent;
    this.tradeComponent = params.tradeComponent;
    this.window = new Array<number>(length);
    this.windowLength = length;
    this.lastIndex = length - 1;
    this.primed = false;
  }

  /** Describes the output data of the indicator. */
  public metadata(): IndicatorMetadata {
    return {
      type: IndicatorType.Variance,
      mnemonic: this.mnemonic,
      description: this.description,
      outputs: [{
        kind: VarianceOutput.VarianceValue,
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

    let value = 0;
    let temp = sample;

    if (this.primed) {
      this.windowSum += temp;
      temp *= temp;
      this.windowSquaredSum += temp;
      temp = this.window[0];
      this.windowSum -= temp;
      temp *= temp;
      this.windowSquaredSum -= temp;

      if (this.unbiased) {
        temp = this.windowSum;
        temp *= temp;
        temp /= this.windowLength;
        value = this.windowSquaredSum - temp;
        value /= this.lastIndex;
      } else {
        temp = this.windowSum / this.windowLength;
        temp *= temp;
        value = this.windowSquaredSum / this.windowLength - temp;
      }

      for (let i = 0; i < this.lastIndex; i++) {
        this.window[i] = this.window[i+1];
      }

      this.window[this.lastIndex] = sample;
    } else { // Not primed.
      this.windowSum += temp;
      this.window[this.windowCount] = temp;
      temp *= temp;
      this.windowSquaredSum += temp;
      this.windowCount++;

      if (this.windowLength > this.windowCount) {
        return Number.NaN;
      }

      this.primed = true;

      if (this.unbiased) {
        temp = this.windowSum;
        temp *= temp;
        temp /= this.windowLength;
        value = this.windowSquaredSum - temp;
        value /= this.lastIndex;
      } else {
        temp = this.windowSum / this.windowLength;
        temp *= temp;
        value = this.windowSquaredSum / this.windowLength - temp;
      }
    }

    return value;
  }
}
