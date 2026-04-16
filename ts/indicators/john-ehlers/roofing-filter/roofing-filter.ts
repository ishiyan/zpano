import { BarComponent } from '../../../entities/bar-component';
import { componentTripleMnemonic } from '../../core/component-triple-mnemonic';
import { IndicatorMetadata } from '../../core/indicator-metadata';
import { IndicatorType } from '../../core/indicator-type';
import { LineIndicator } from '../../core/line-indicator';
import { OutputType } from '../../core/outputs/output-type';
import { RoofingFilterOutput } from './roofing-filter-output';
import { RoofingFilterParams } from './roofing-filter-params';

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
  private samplePrev: number;
  private samplePrev2: number;
  private hpPrev: number;
  private hpPrev2: number;
  private ssPrev: number;
  private ssPrev2: number;
  private zmPrev: number;
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
    this.samplePrev = 0;
    this.samplePrev2 = 0;
    this.hpPrev = 0;
    this.hpPrev2 = 0;
    this.ssPrev = 0;
    this.ssPrev2 = 0;
    this.zmPrev = 0;
    this.value = Number.NaN;
    this.primed = false;
  }

  /** Describes the output data of the indicator. */
  public metadata(): IndicatorMetadata {
    return {
      type: IndicatorType.RoofingFilter,
      mnemonic: this.mnemonic,
      description: this.description,
      outputs: [{
        kind: RoofingFilterOutput.RoofingFilterValue,
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
      hp = this.hpCoeff1 * (sample - this.samplePrev) + this.hpCoeff2 * this.hpPrev;
      ss = this.ssCoeff1 * (hp + this.hpPrev) + this.ssCoeff2 * this.ssPrev + this.ssCoeff3 * this.ssPrev2;

      if (this.hasZeroMeanFilter) {
        zm = this.hpCoeff1 * (ss - this.ssPrev) + this.hpCoeff2 * this.zmPrev;
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
        hp = this.hpCoeff1 * (sample - this.samplePrev) + this.hpCoeff2 * this.hpPrev;
        ss = this.ssCoeff1 * (hp + this.hpPrev) + this.ssCoeff2 * this.ssPrev + this.ssCoeff3 * this.ssPrev2;

        if (this.hasZeroMeanFilter) {
          zm = this.hpCoeff1 * (ss - this.ssPrev) + this.hpCoeff2 * this.zmPrev;
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

    this.samplePrev = sample;
    this.hpPrev = hp;
    this.ssPrev2 = this.ssPrev;
    this.ssPrev = ss;

    if (this.hasZeroMeanFilter) {
      this.zmPrev = zm;
    }

    return this.value;
  }

  private update2Pole(sample: number): number {
    let hp = 0;
    let ss = 0;

    if (this.primed) {
      hp = this.hpCoeff1 * (sample - 2 * this.samplePrev + this.samplePrev2) +
        this.hpCoeff2 * this.hpPrev - this.hpCoeff3 * this.hpPrev2;
      ss = this.ssCoeff1 * (hp + this.hpPrev) + this.ssCoeff2 * this.ssPrev + this.ssCoeff3 * this.ssPrev2;
      this.value = ss;
    } else {
      this.count++;

      if (this.count < 4) {
        hp = 0;
        ss = 0;
      } else {
        hp = this.hpCoeff1 * (sample - 2 * this.samplePrev + this.samplePrev2) +
          this.hpCoeff2 * this.hpPrev - this.hpCoeff3 * this.hpPrev2;
        ss = this.ssCoeff1 * (hp + this.hpPrev) + this.ssCoeff2 * this.ssPrev + this.ssCoeff3 * this.ssPrev2;

        if (this.count === 5) {
          this.primed = true;
          this.value = ss;
        }
      }
    }

    this.samplePrev2 = this.samplePrev;
    this.samplePrev = sample;
    this.hpPrev2 = this.hpPrev;
    this.hpPrev = hp;
    this.ssPrev2 = this.ssPrev;
    this.ssPrev = ss;

    return this.value;
  }
}
