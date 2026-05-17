import { buildMetadata } from '../../core/build-metadata';
import { componentTripleMnemonic } from '../../core/component-triple-mnemonic';
import { IndicatorMetadata } from '../../core/indicator-metadata';
import { IndicatorIdentifier } from '../../core/indicator-identifier';
import { LineIndicator } from '../../core/line-indicator';
import { FractalDimensionIndexParams } from './params';

/** Function to calculate mnemonic of a __FractalDimensionIndex__ indicator. */
export const fractalDimensionIndexMnemonic = (params: FractalDimensionIndexParams): string =>
  'fdi('.concat(params.period.toString(), componentTripleMnemonic(params.barComponent, params.quoteComponent, params.tradeComponent), ')');

/** Fractal Dimension Index line indicator. */
export class FractalDimensionIndex extends LineIndicator {
  private window: Array<number>;
  private period: number;
  private windowCount: number;
  private log2N: number;
  private ln2: number;
  private invNSq: number;

  /**
   * Constructs an instance given a period.
   * The period should be an integer greater than 1.
   **/
  public constructor(params: FractalDimensionIndexParams){
    super();
    const period = Math.floor(params.period);
    if (period < 2) {
      throw new Error('period should be greater than 1');
    }

    this.mnemonic = fractalDimensionIndexMnemonic(params);
    this.description = 'Fractal dimension index ' + this.mnemonic;
    this.barComponent = params.barComponent;
    this.quoteComponent = params.quoteComponent;
    this.tradeComponent = params.tradeComponent;
    this.window = new Array<number>(period + 1);
    this.period = period;
    this.windowCount = 0;
    this.primed = false;
    this.log2N = Math.log(2.0 * period);
    this.ln2 = Math.log(2.0);
    this.invNSq = 1.0 / (period * period);
  }

  /** Describes the output data of the indicator. */
  public metadata(): IndicatorMetadata {
    return buildMetadata(
      IndicatorIdentifier.FractalDimensionIndex,
      this.mnemonic,
      this.description,
      [{ mnemonic: this.mnemonic, description: this.description }],
    );
  }

  /** Updates the value of the indicator given the next sample. */
  public update(sample: number): number {
    if (Number.isNaN(sample)) {
      return sample;
    }

    const period = this.period;

    if (this.primed) {
      for (let i = 0; i < period; i++) {
        this.window[i] = this.window[i + 1];
      }

      this.window[period] = sample;
    } else {
      this.window[this.windowCount] = sample;
      this.windowCount++;

      if (this.windowCount <= period) {
        return Number.NaN;
      }

      this.primed = true;
    }

    // Find min/max for normalization.
    let priceMax = this.window[0];
    let priceMin = this.window[0];

    for (let k = 1; k <= period; k++) {
      if (this.window[k] > priceMax) { priceMax = this.window[k]; }
      if (this.window[k] < priceMin) { priceMin = this.window[k]; }
    }

    const priceRange = priceMax - priceMin;
    if (priceRange < 1e-10) {
      return 1.0;
    }

    // Normalize and compute path length.
    let priorNorm = (this.window[0] - priceMin) / priceRange;
    let length = 0.0;

    for (let k = 1; k <= period; k++) {
      const currNorm = (this.window[k] - priceMin) / priceRange;
      const diff = currNorm - priorNorm;
      length += Math.sqrt(diff * diff + this.invNSq);
      priorNorm = currNorm;
    }

    return 1.0 + (Math.log(length) + this.ln2) / this.log2N;
  }
}
