import { buildMetadata } from '../../core/build-metadata';
import { BarComponent } from '../../../entities/bar-component';
import { componentTripleMnemonic } from '../../core/component-triple-mnemonic';
import { IndicatorMetadata } from '../../core/indicator-metadata';
import { IndicatorIdentifier } from '../../core/indicator-identifier';
import { LineIndicator } from '../../core/line-indicator';
import { RoofingFilterParams } from './params';

/** Function to calculate mnemonic of a __RoofingFilter__ indicator. */
export const roofingFilterMnemonic = (params: RoofingFilterParams): string => {
  const cm = componentTripleMnemonic(
    params.barComponent ?? BarComponent.Median,
    params.quoteComponent,
    params.tradeComponent,
  );
  const poles = params.hasTwoPoleHighpassFilter ? 2 : 1;
  const zm = (params.hasZeroMean && !params.hasTwoPoleHighpassFilter) ? 'zm' : '';

  return `roof${poles}hp${zm}(${params.shortestCyclePeriod}, ${params.longestCyclePeriod}${cm})`;
};

/**
 * RoofingFilter (Ehler's Roofing Filter) is described in Ehler's book
 * "Cycle Analytics for Traders" (2013).
 *
 * The Roofing Filter is comprised of a high-pass filter and a Super Smoother.
 * Given the longest (Λ) and the shortest (λ) cycle periods in bars,
 * the high-pass filter passes cyclic components whose periods are shorter than the longest one,
 * and the Super Smoother filter attenuates cycle periods shorter than the shortest one.
 *
 * Three flavours are available:
 *   - 1-pole high-pass filter (default)
 *   - 1-pole high-pass filter with zero-mean
 *   - 2-pole high-pass filter
 *
 * Reference:
 *
 * Ehlers, John F. (2013). Cycle Analytics for Traders. Wiley.
 */
export class RoofingFilter extends LineIndicator {
  private hpCoeff1: number;
  private hpCoeff2: number;
  private hpCoeff3: number;
  private ssCoeff1: number;
  private ssCoeff2: number;
  private ssCoeff3: number;

  private hasTwoPole: boolean;
  private hasZeroMeanFilter: boolean;

  private count: number;
  private samplePrevious: number;
  private samplePrevious2: number;
  private hpPrevious: number;
  private hpPrevious2: number;
  private ssPrevious: number;
  private ssPrevious2: number;
  private zmPrevious: number;
  private value: number;

  /**
   * Constructs an instance given the parameters.
   */
  public constructor(params: RoofingFilterParams) {
    super();

    const shortest = Math.floor(params.shortestCyclePeriod);
    if (shortest < 2) {
      throw new Error('shortest cycle period should be greater than 1');
    }

    const longest = Math.floor(params.longestCyclePeriod);
    if (longest <= shortest) {
      throw new Error('longest cycle period should be greater than shortest');
    }

    this.hasTwoPole = params.hasTwoPoleHighpassFilter ?? false;
    this.hasZeroMeanFilter = (params.hasZeroMean ?? false) && !this.hasTwoPole;

    this.mnemonic = roofingFilterMnemonic(params);
    this.description = 'Roofing Filter ' + this.mnemonic;
    this.barComponent = params.barComponent;
    this.quoteComponent = params.quoteComponent;
    this.tradeComponent = params.tradeComponent;

    // Calculate high-pass filter coefficients.
    this.hpCoeff3 = 0;
    if (this.hasTwoPole) {
      const angle = (Math.SQRT2 / 2) * 2 * Math.PI / longest;
      const cosAngle = Math.cos(angle);
      const alpha = (Math.sin(angle) + cosAngle - 1) / cosAngle;
      const beta = 1 - alpha / 2;
      this.hpCoeff1 = beta * beta;
      const beta2 = 1 - alpha;
      this.hpCoeff2 = 2 * beta2;
      this.hpCoeff3 = beta2 * beta2;
    } else {
      const angle = 2 * Math.PI / longest;
      const cosAngle = Math.cos(angle);
      const alpha = (Math.sin(angle) + cosAngle - 1) / cosAngle;
      this.hpCoeff1 = 1 - alpha / 2;
      this.hpCoeff2 = 1 - alpha;
    }

    // Calculate super smoother coefficients (uses literal 1.414).
    const beta = 1.414 * Math.PI / shortest;
    const alpha = Math.exp(-beta);
    this.ssCoeff2 = 2 * alpha * Math.cos(beta);
    this.ssCoeff3 = -alpha * alpha;
    this.ssCoeff1 = (1 - this.ssCoeff2 - this.ssCoeff3) / 2;

    this.count = 0;
    this.samplePrevious = 0;
    this.samplePrevious2 = 0;
    this.hpPrevious = 0;
    this.hpPrevious2 = 0;
    this.ssPrevious = 0;
    this.ssPrevious2 = 0;
    this.zmPrevious = 0;
    this.value = Number.NaN;
    this.primed = false;
  }

  /** Describes the output data of the indicator. */
  public metadata(): IndicatorMetadata {
    return buildMetadata(
      IndicatorIdentifier.RoofingFilter,
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

    if (this.hasTwoPole) {
      return this.update2Pole(sample);
    }

    return this.update1Pole(sample);
  }

  private update1Pole(sample: number): number {
    let hp = 0;
    let ss = 0;
    let zm = 0;

    if (this.primed) {
      hp = this.hpCoeff1 * (sample - this.samplePrevious) + this.hpCoeff2 * this.hpPrevious;
      ss = this.ssCoeff1 * (hp + this.hpPrevious) + this.ssCoeff2 * this.ssPrevious + this.ssCoeff3 * this.ssPrevious2;

      if (this.hasZeroMeanFilter) {
        zm = this.hpCoeff1 * (ss - this.ssPrevious) + this.hpCoeff2 * this.zmPrevious;
        this.value = zm;
      } else {
        this.value = ss;
      }
    } else {
      this.count++;

      if (this.count === 1) {
        hp = 0;
        ss = 0;
      } else {
        hp = this.hpCoeff1 * (sample - this.samplePrevious) + this.hpCoeff2 * this.hpPrevious;
        ss = this.ssCoeff1 * (hp + this.hpPrevious) + this.ssCoeff2 * this.ssPrevious + this.ssCoeff3 * this.ssPrevious2;

        if (this.hasZeroMeanFilter) {
          zm = this.hpCoeff1 * (ss - this.ssPrevious) + this.hpCoeff2 * this.zmPrevious;
          if (this.count === 5) {
            this.primed = true;
            this.value = zm;
          }
        } else if (this.count === 4) {
          this.primed = true;
          this.value = ss;
        }
      }
    }

    this.samplePrevious = sample;
    this.hpPrevious = hp;
    this.ssPrevious2 = this.ssPrevious;
    this.ssPrevious = ss;

    if (this.hasZeroMeanFilter) {
      this.zmPrevious = zm;
    }

    return this.value;
  }

  private update2Pole(sample: number): number {
    let hp = 0;
    let ss = 0;

    if (this.primed) {
      hp = this.hpCoeff1 * (sample - 2 * this.samplePrevious + this.samplePrevious2) +
        this.hpCoeff2 * this.hpPrevious - this.hpCoeff3 * this.hpPrevious2;
      ss = this.ssCoeff1 * (hp + this.hpPrevious) + this.ssCoeff2 * this.ssPrevious + this.ssCoeff3 * this.ssPrevious2;
      this.value = ss;
    } else {
      this.count++;

      if (this.count < 4) {
        hp = 0;
        ss = 0;
      } else {
        hp = this.hpCoeff1 * (sample - 2 * this.samplePrevious + this.samplePrevious2) +
          this.hpCoeff2 * this.hpPrevious - this.hpCoeff3 * this.hpPrevious2;
        ss = this.ssCoeff1 * (hp + this.hpPrevious) + this.ssCoeff2 * this.ssPrevious + this.ssCoeff3 * this.ssPrevious2;

        if (this.count === 5) {
          this.primed = true;
          this.value = ss;
        }
      }
    }

    this.samplePrevious2 = this.samplePrevious;
    this.samplePrevious = sample;
    this.hpPrevious2 = this.hpPrevious;
    this.hpPrevious = hp;
    this.ssPrevious2 = this.ssPrevious;
    this.ssPrevious = ss;

    return this.value;
  }
}
