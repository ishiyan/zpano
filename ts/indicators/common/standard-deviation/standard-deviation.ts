import { buildMetadata } from '../../core/build-metadata';
import { componentTripleMnemonic } from '../../core/component-triple-mnemonic';
import { IndicatorMetadata } from '../../core/indicator-metadata';
import { IndicatorIdentifier } from '../../core/indicator-identifier';
import { LineIndicator } from '../../core/line-indicator';
import { Variance } from '../variance/variance';
import { StandardDeviationParams } from './params';

/** Function to calculate mnemonic of a __StandardDeviation__ indicator. */
export const standardDeviationMnemonic = (params: StandardDeviationParams): string =>
  'stdev.'.concat(params.unbiased ? 's(' : 'p(', params.length.toString(),
  componentTripleMnemonic(params.barComponent, params.quoteComponent, params.tradeComponent), ')');

/** Standard deviation line indicator.
 *
 * StandardDeviation computes the standard deviation of the samples within a moving window of length ℓ
 * as a square root of variance:
 *
 *     σ² = (∑xᵢ² - (∑xᵢ)²/ℓ)/ℓ
 *
 * for the estimation of the population variance, or as:
 *
 *     σ² = (∑xᵢ² - (∑xᵢ)²/ℓ)/(ℓ-1)
 *
 * for the unbiased estimation of the sample variance, i={0,…,ℓ-1}.
 */
export class StandardDeviation extends LineIndicator {
  private variance: Variance;

  /**
   * Constructs an instance given a length in samples.
   * The length should be an integer greater than 1.
   **/
  public constructor(params: StandardDeviationParams){
    super();
    const length = Math.floor(params.length);
    if (length < 2) {
      throw new Error('length should be greater than 1');
    }

    this.mnemonic = standardDeviationMnemonic(params);
    if (params.unbiased) {
      this.description = 'Standard deviation based on unbiased estimation of the sample variance ' + this.mnemonic;
    } else {
      this.description = 'Standard deviation based on estimation of the population variance ' + this.mnemonic;
    }

    this.barComponent = params.barComponent;
    this.quoteComponent = params.quoteComponent;
    this.tradeComponent = params.tradeComponent;
    this.variance = new Variance(params);
    this.primed = false;
  }

  /** Describes the output data of the indicator. */
  public metadata(): IndicatorMetadata {
    return buildMetadata(
      IndicatorIdentifier.StandardDeviation,
      this.mnemonic,
      this.description,
      [
        { mnemonic: this.mnemonic, description: this.description },
      ],
    );
  }

  /** Updates the value of the indicator given the next sample. */
  public update(sample: number): number {
    const value = this.variance.update(sample);
    if (Number.isNaN(value)) {
      return value;
    }

    this.primed = this.variance.isPrimed();
    return Math.sqrt(value);
  }
}
