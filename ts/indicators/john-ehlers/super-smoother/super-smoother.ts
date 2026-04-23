import { buildMetadata } from '../../core/build-metadata';
import { BarComponent } from '../../../entities/bar-component';
import { componentTripleMnemonic } from '../../core/component-triple-mnemonic';
import { IndicatorMetadata } from '../../core/indicator-metadata';
import { IndicatorIdentifier } from '../../core/indicator-identifier';
import { LineIndicator } from '../../core/line-indicator';
import { SuperSmootherParams } from './params';

/** Function to calculate mnemonic of a __SuperSmoother__ indicator. */
export const superSmootherMnemonic = (params: SuperSmootherParams): string => {
  const cm = componentTripleMnemonic(
    params.barComponent ?? BarComponent.Median,
    params.quoteComponent,
    params.tradeComponent,
  );

  return `ss(${params.shortestCyclePeriod}${cm})`;
};

/**
 * SuperSmoother (Ehler's Super Smoother, SS) is described in Ehler's book
 * "Cybernetic Analysis for Stocks and Futures" (2004)
 * and in his presentation "Spectral Dilation" (2013)
 *
 * Given the shortest (λ) cycle period in bars, the Super Smoother filter
 * attenuates cycle periods shorter than this shortest one.
 *
 *	β = √2·π / λ
 *	α = exp(-β)
 *	γ₂ = 2α·cos(β)
 *	γ₃ = -α²
 *	γ₁ = (1 - γ₂ - γ₃) / 2
 *
 *	SSᵢ = γ₁·(xᵢ + xᵢ₋₁) + γ₂·SSᵢ₋₁ + γ₃·SSᵢ₋₂
 *
 * The indicator is not primed during the first 2 updates.
 *
 * Reference:
 * 
 * Ehlers, John F. (2004). Cybernetic Analysis for Stocks and Futures. Wiley. pp 201-205.
 * Ehlers, John F. (2013). Spectral dilation: Presented to the MTA in March 2013. Retrieved from www.mesasoftware.com/seminars/SpectralDilation.pdf
 */
export class SuperSmoother extends LineIndicator {
  private coeff1: number;
  private coeff2: number;
  private coeff3: number;
  private count: number;
  private samplePrevious: number;
  private filterPrevious: number;
  private filterPrevious2: number;
  private value: number;

  /**
   * Constructs an instance given a shortest cycle period in bars.
   * The shortest cycle period should be an integer greater than 1.
   */
  public constructor(params: SuperSmootherParams) {
    super();
    const period = Math.floor(params.shortestCyclePeriod);
    if (period < 2) {
      throw new Error('shortest cycle period should be greater than 1');
    }

    this.mnemonic = superSmootherMnemonic(params);
    this.description = 'Super Smoother ' + this.mnemonic;
    this.barComponent = params.barComponent;
    this.quoteComponent = params.quoteComponent;
    this.tradeComponent = params.tradeComponent;

    // Calculate coefficients.
    const beta = Math.SQRT2 * Math.PI / period;
    const alpha = Math.exp(-beta);
    const gamma2 = 2 * alpha * Math.cos(beta);
    const gamma3 = -alpha * alpha;
    const gamma1 = (1 - gamma2 - gamma3) / 2;

    this.coeff1 = gamma1;
    this.coeff2 = gamma2;
    this.coeff3 = gamma3;
    this.count = 0;
    this.samplePrevious = 0;
    this.filterPrevious = 0;
    this.filterPrevious2 = 0;
    this.value = Number.NaN;
    this.primed = false;
  }

  /** Describes the output data of the indicator. */
  public metadata(): IndicatorMetadata {
    return buildMetadata(
      IndicatorIdentifier.SuperSmoother,
      this.mnemonic,
      this.description,
      [
        { mnemonic: this.mnemonic, description: this.description },
      ],
    );
  }

  /** Updates the value of the indicator given the next sample. */
  public update(sample: number): number {
    if (Number.isNaN(sample)) {
      return sample;
    }

    if (this.primed) {
      const filter = this.coeff1 * (sample + this.samplePrevious) +
        this.coeff2 * this.filterPrevious + this.coeff3 * this.filterPrevious2;
      this.value = filter;
      this.samplePrevious = sample;
      this.filterPrevious2 = this.filterPrevious;
      this.filterPrevious = filter;

      return this.value;
    }

    this.count++;

    if (this.count === 1) {
      this.samplePrevious = sample;
      this.filterPrevious = sample;
      this.filterPrevious2 = sample;
    }

    const filter = this.coeff1 * (sample + this.samplePrevious) +
      this.coeff2 * this.filterPrevious + this.coeff3 * this.filterPrevious2;

    if (this.count === 3) {
      this.primed = true;
      this.value = filter;
    }

    this.samplePrevious = sample;
    this.filterPrevious2 = this.filterPrevious;
    this.filterPrevious = filter;

    return this.value;
  }
}
