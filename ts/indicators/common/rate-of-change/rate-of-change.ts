import { componentTripleMnemonic } from '../../core/component-triple-mnemonic';
import { IndicatorMetadata } from '../../core/indicator-metadata';
import { IndicatorType } from '../../core/indicator-type';
import { LineIndicator } from '../../core/line-indicator';
import { OutputType } from '../../core/outputs/output-type';
import { RateOfChangeOutput } from './rate-of-change-output';
import { RateOfChangeParams } from './rate-of-change-params';

/**
 * __Rate of Change__ (__ROC__) is the difference between today's sample and the sample ℓ periods ago
 * scaled by the old sample so as to represent the increase as a fraction.
 *
 * The values are centered at zero and can be positive and negative.
 *
 *     ROCᵢ = 100 (Pᵢ - Pᵢ₋ℓ) / Pᵢ₋ℓ = 100 (Pᵢ/Pᵢ₋ℓ - 1),
 *
 * where ℓ is the length.
 *
 * The indicator is not primed during the first ℓ updates.
 */
export class RateOfChange extends LineIndicator {
  private readonly window: number[];
  private readonly windowLength: number;
  private readonly lastIndex: number;
  private windowCount = 0;

  /** Constructs an instance given the parameters. */
  public constructor(params: RateOfChangeParams) {
    super();
    const length = Math.floor(params.length);
    if (length < 1) {
      throw new Error('length should be positive');
    }

    this.window = new Array<number>(length + 1).fill(0);
    this.windowLength = length + 1;
    this.lastIndex = length;

    const mn = 'roc(' + length.toString() +
      componentTripleMnemonic(params.barComponent, params.quoteComponent, params.tradeComponent) + ')';

    this.mnemonic = mn;
    this.description = 'Rate of Change ' + mn;
    this.barComponent = params.barComponent;
    this.quoteComponent = params.quoteComponent;
    this.tradeComponent = params.tradeComponent;
    this.primed = false;
  }

  /** Describes the output data of the indicator. */
  public metadata(): IndicatorMetadata {
    return {
      type: IndicatorType.RateOfChange,
      mnemonic: this.mnemonic,
      description: this.description,
      outputs: [{
        kind: RateOfChangeOutput.RateOfChangeValue,
        type: OutputType.Scalar,
        mnemonic: this.mnemonic,
        description: this.description,
      }],
    };
  }

  /** Updates the value of the rate of change given the next sample. */
  public update(sample: number): number {
    if (Number.isNaN(sample)) {
      return sample;
    }

    const epsilon = 1e-13;
    const c100 = 100;

    if (this.primed) {
      if (this.lastIndex > 1) {
        for (let i = 0; i < this.lastIndex; i++) {
          this.window[i] = this.window[i + 1];
        }
      }

      this.window[this.lastIndex] = sample;
      const previous = this.window[0];
      if (Math.abs(previous) > epsilon) {
        return (sample / previous - 1) * c100;
      }

      return 0;
    }

    this.window[this.windowCount] = sample;
    this.windowCount++;

    if (this.windowLength === this.windowCount) {
      this.primed = true;
      const previous = this.window[0];
      if (Math.abs(previous) > epsilon) {
        return (sample / previous - 1) * c100;
      }

      return 0;
    }

    return Number.NaN;
  }
}
