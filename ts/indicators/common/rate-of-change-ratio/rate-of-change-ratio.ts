import { componentTripleMnemonic } from '../../core/component-triple-mnemonic';
import { IndicatorMetadata } from '../../core/indicator-metadata';
import { IndicatorType } from '../../core/indicator-type';
import { LineIndicator } from '../../core/line-indicator';
import { OutputType } from '../../core/outputs/output-type';
import { RateOfChangeRatioOutput } from './rate-of-change-ratio-output';
import { RateOfChangeRatioParams } from './rate-of-change-ratio-params';

/**
 * __Rate of Change Ratio__ (__ROCR__) is the ratio of today's sample to the sample ℓ periods ago.
 *
 * The values are centered at 1 (or 100 when hundredScale is true) and are always positive.
 *
 *     ROCRᵢ = Pᵢ / Pᵢ₋ℓ,
 *     ROCR100ᵢ = (Pᵢ / Pᵢ₋ℓ) * 100,
 *
 * where ℓ is the length.
 *
 * The indicator is not primed during the first ℓ updates.
 */
export class RateOfChangeRatio extends LineIndicator {
  private readonly window: number[];
  private readonly windowLength: number;
  private readonly lastIndex: number;
  private readonly scale: number;
  private windowCount = 0;

  /** Constructs an instance given the parameters. */
  public constructor(params: RateOfChangeRatioParams) {
    super();
    const length = Math.floor(params.length);
    if (length < 1) {
      throw new Error('length should be positive');
    }

    this.window = new Array<number>(length + 1).fill(0);
    this.windowLength = length + 1;
    this.lastIndex = length;
    this.scale = params.hundredScale ? 100 : 1;

    const prefix = params.hundredScale ? 'rocr100(' : 'rocr(';
    const mn = prefix + length.toString() +
      componentTripleMnemonic(params.barComponent, params.quoteComponent, params.tradeComponent) + ')';

    const descPrefix = params.hundredScale ? 'Rate of Change Ratio 100 Scale ' : 'Rate of Change Ratio ';

    this.mnemonic = mn;
    this.description = descPrefix + mn;
    this.barComponent = params.barComponent;
    this.quoteComponent = params.quoteComponent;
    this.tradeComponent = params.tradeComponent;
    this.primed = false;
  }

  /** Describes the output data of the indicator. */
  public metadata(): IndicatorMetadata {
    return {
      type: IndicatorType.RateOfChangeRatio,
      mnemonic: this.mnemonic,
      description: this.description,
      outputs: [{
        kind: RateOfChangeRatioOutput.RateOfChangeRatioValue,
        type: OutputType.Scalar,
        mnemonic: this.mnemonic,
        description: this.description,
      }],
    };
  }

  /** Updates the value of the rate of change ratio given the next sample. */
  public update(sample: number): number {
    if (Number.isNaN(sample)) {
      return sample;
    }

    const epsilon = 1e-13;

    if (this.primed) {
      if (this.lastIndex > 1) {
        for (let i = 0; i < this.lastIndex; i++) {
          this.window[i] = this.window[i + 1];
        }
      }

      this.window[this.lastIndex] = sample;
      const previous = this.window[0];
      if (Math.abs(previous) > epsilon) {
        return (sample / previous) * this.scale;
      }

      return 0;
    }

    this.window[this.windowCount] = sample;
    this.windowCount++;

    if (this.windowLength === this.windowCount) {
      this.primed = true;
      const previous = this.window[0];
      if (Math.abs(previous) > epsilon) {
        return (sample / previous) * this.scale;
      }

      return 0;
    }

    return Number.NaN;
  }
}
