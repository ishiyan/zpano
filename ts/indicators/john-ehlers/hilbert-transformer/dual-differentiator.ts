import { HilbertTransformerCycleEstimator } from './cycle-estimator';
import { HilbertTransformerCycleEstimatorParams } from './cycle-estimator-params';
import {
  defaultMinPeriod, defaultMaxPeriod, htLength, quadratureIndex,
  push, correctAmplitude, ht, adjustPeriod, fillWmaFactors, verifyParameters
} from './common';

/** A Hilbert transformer of WMA-smoothed and detrended data with the dual differentiator applied.
  *
  *  John Ehlers, Rocket Science for Traders, Wiley, 2001, 0471405671, pp 70-74.
  */
export class HilbertTransformerDualDifferentiator implements HilbertTransformerCycleEstimator {

  /** The underlying linear-Weighted Moving Average (WMA) smoothing length. */
  public readonly smoothingLength: number;

  /** The current WMA-smoothed value used by underlying Hilbert transformer.
   *
   * The linear-Weighted Moving Average has a window size of __smoothingLength__.
   */
  public get smoothed(): number { return this.wmaSmoothed[0]; }

  /** The current de-trended value. */
  public get detrended(): number { return this._detrended[0]; }

  /** The current Quadrature component value. */
  public get quadrature(): number { return this._quadrature[0]; }

  /** The current InPhase component value. */
  public get inPhase(): number { return this._inPhase[0]; }

  /** The current period value. */
  public get period(): number { return this._period; }

  /** The current count value. */
  public get count(): number { return this._count; }

  /** Indicates whether an estimator is primed. */
  public get primed(): boolean { return this.isWarmedUp; }

  /** The minimal cycle period supported by this Hilbert transformer. */
  public readonly minPeriod: number = defaultMinPeriod;

  /** The maximual cycle period supported by this Hilbert transformer. */
  public readonly maxPeriod: number = defaultMaxPeriod;

  /** The value of α (0 < α ≤ 1) used in EMA to smooth the in-phase and quadrature components. */
  public readonly alphaEmaQuadratureInPhase: number;

  /** The value of α (0 < α ≤ 1) used in EMA to smooth the instantaneous period. */
  public readonly alphaEmaPeriod: number;

  /** The number of updates before the estimator is primed (MaxPeriod * 2 = 100). */
  public readonly warmUpPeriod: number;

  private readonly smoothingLengthPlusHtLengthMin1: number;
  private readonly smoothingLengthPlus2HtLengthMin2: number;
  private readonly smoothingLengthPlus3HtLengthMin3: number;
  private readonly smoothingLengthPlus3HtLengthMin2: number;
  private readonly smoothingLengthPlus3HtLengthMin1: number;
  //private readonly smoothingLengthPlus3HtLength: number;

  private readonly oneMinAlphaEmaQuadratureInPhase: number;
  private readonly oneMinAlphaEmaPeriod: number;

  private readonly rawValues: Array<number>;
  private readonly wmaFactors: Array<number>;
  private readonly wmaSmoothed: Array<number> = new Array(htLength).fill(0);
  private readonly _detrended: Array<number> = new Array(htLength).fill(0);
  private readonly _inPhase: Array<number> = new Array(htLength).fill(0);
  private readonly _quadrature: Array<number> = new Array(htLength).fill(0);
  private readonly jInPhase: Array<number> = new Array(htLength).fill(0);
  private readonly jQuadrature: Array<number> = new Array(htLength).fill(0);

  private _count: number = 0;
  private smoothedInPhasePrevious: number = 0;
  private smoothedQuadraturePrevious: number = 0;
  private _period: number = defaultMinPeriod;
  private isPrimed = false;
  private isWarmedUp = false;

  /**
   * Constructs an instance using given parameters.
   **/
  public constructor(params: HilbertTransformerCycleEstimatorParams) {
    const err = verifyParameters(params);
    if (err) {
      throw new Error(err);
    }

    this.alphaEmaQuadratureInPhase = params.alphaEmaQuadratureInPhase;
    this.oneMinAlphaEmaQuadratureInPhase = 1 - params.alphaEmaQuadratureInPhase;
    this.alphaEmaPeriod = params.alphaEmaPeriod;
    this.oneMinAlphaEmaPeriod = 1 - params.alphaEmaPeriod;

    const length = Math.floor(params.smoothingLength);
    this.smoothingLength = length;
    this.smoothingLengthPlusHtLengthMin1 = length + htLength - 1;
    this.smoothingLengthPlus2HtLengthMin2 = this.smoothingLengthPlusHtLengthMin1 + htLength - 1;
    this.smoothingLengthPlus3HtLengthMin3 = this.smoothingLengthPlus2HtLengthMin2 + htLength - 1;
    this.smoothingLengthPlus3HtLengthMin2 = this.smoothingLengthPlus3HtLengthMin3 + 1;
    this.smoothingLengthPlus3HtLengthMin1 = this.smoothingLengthPlus3HtLengthMin2 + 1;

    this.rawValues = new Array(length).fill(0);
    this.wmaFactors = new Array(length);
    fillWmaFactors(length, this.wmaFactors);

    if (params.warmUpPeriod && params.warmUpPeriod > this.smoothingLengthPlus3HtLengthMin1) {
      this.warmUpPeriod = params.warmUpPeriod;
    } else {
      this.warmUpPeriod = this.smoothingLengthPlus3HtLengthMin1;
    }
  }

  private wma(array: number[]): number {
    let value = 0;
    for (let i = 0; i < this.smoothingLength; ++i) {
      value += this.wmaFactors[i] * array[i];
    }

    return value;
  }

  private emaQuadratureInPhase(value: number, valuePrevious: number): number {
    return this.alphaEmaQuadratureInPhase * value + this.oneMinAlphaEmaQuadratureInPhase * valuePrevious;
  }

  private emaPeriod(value: number, valuePrevious: number): number {
    return this.alphaEmaPeriod * value + this.oneMinAlphaEmaPeriod * valuePrevious;
  }

  /** Updates the estimator given the next sample value. */
  public update(sample: number): void {
    if (Number.isNaN(sample)) {
      return;
    }

    push(this.rawValues, sample);
    if (this.isPrimed) {
      if (!this.isWarmedUp) {
        ++this._count;
        if (this.warmUpPeriod < this._count) {
          this.isWarmedUp = true;
        }
      }

      // The WMA is used to remove some high-frequency components before detrending the signal.
      push(this.wmaSmoothed, this.wma(this.rawValues));

      const amplitudeCorrectionFactor = correctAmplitude(this._period);

      // Since we have an amplitude-corrected Hilbert transformer, and since we want to detrend
      // over its length, we simply use the Hilbert transformer itself as the detrender.
      push(this._detrended, ht(this.wmaSmoothed) * amplitudeCorrectionFactor);

      // Compute both the in-phase and quadrature components of the detrended signal.
      push(this._quadrature, ht(this._detrended) * amplitudeCorrectionFactor);
      push(this._inPhase, this._detrended[quadratureIndex]);

      // Complex averaging: apply the Hilbert Transformer to both the in-phase and quadrature components.
      // This advances the phase of each component by 90°.
      push(this.jInPhase, ht(this._inPhase) * amplitudeCorrectionFactor);
      push(this.jQuadrature, ht(this._quadrature) * amplitudeCorrectionFactor);

      // Phasor addition for 3 bar averaging followed by exponential moving average smoothing.
      const smoothedInPhase = this.emaQuadratureInPhase(
        this._inPhase[0] - this.jQuadrature[0], this.smoothedInPhasePrevious);
      const smoothedQuadrature = this.emaQuadratureInPhase(
        this._quadrature[0] + this.jInPhase[0], this.smoothedQuadraturePrevious);

      // Dual Differential discriminator.
      const discriminator = smoothedQuadrature * (smoothedInPhase - this.smoothedInPhasePrevious) -
        smoothedInPhase * (smoothedQuadrature - this.smoothedQuadraturePrevious);
      this.smoothedInPhasePrevious = smoothedInPhase;
      this.smoothedQuadraturePrevious = smoothedQuadrature;

      const periodPrevious = this._period;
      const periodNew = 2 * Math.PI * (smoothedInPhase * smoothedInPhase +
        smoothedQuadrature * smoothedQuadrature) / discriminator;
      if (!Number.isNaN(periodNew) && Number.isFinite(periodNew)) {
        this._period = periodNew;
      }

      this._period = adjustPeriod(this._period, periodPrevious);

      // Exponential moving average smoothing of the period.
      this._period = this.emaPeriod(this._period, periodPrevious);
    } else { // Not primed.
      // On (smoothingLength)-th sample we calculate the first
      // WMA smoothed value and begin with the detrender.
      ++this._count;
      if (this.smoothingLength > this._count) { // count < 4
        return;
      }

      push(this.wmaSmoothed, this.wma(this.rawValues)); // count >= 4
      if (this.smoothingLengthPlusHtLengthMin1 > this._count) { // count < 10
        return;
      }

      const amplitudeCorrectionFactor = correctAmplitude(this._period); // count >= 10

      push(this._detrended, ht(this.wmaSmoothed) * amplitudeCorrectionFactor);
      if (this.smoothingLengthPlus2HtLengthMin2 > this._count) { // count < 16
        return;
      }

      push(this._quadrature, ht(this._detrended) * amplitudeCorrectionFactor); // count >= 16
      push(this._inPhase, this._detrended[quadratureIndex]);
      if (this.smoothingLengthPlus3HtLengthMin3 > this._count) { // count < 22
        return;
      }

      push(this.jInPhase, ht(this._inPhase) * amplitudeCorrectionFactor); // count >= 22
      push(this.jQuadrature, ht(this._quadrature) * amplitudeCorrectionFactor);

      if (this.smoothingLengthPlus3HtLengthMin3 === this._count) { // count == 22
        this.smoothedInPhasePrevious = this._inPhase[0] - this.jQuadrature[0];
        this.smoothedQuadraturePrevious = this._quadrature[0] + this.jInPhase[0];
        return;
      }

      const smoothedInPhase = this.emaQuadratureInPhase(
        this._inPhase[0] - this.jQuadrature[0], this.smoothedInPhasePrevious); // count >= 23
      const smoothedQuadrature = this.emaQuadratureInPhase(
        this._quadrature[0] + this.jInPhase[0], this.smoothedQuadraturePrevious);

      const discriminator = smoothedQuadrature * (smoothedInPhase - this.smoothedInPhasePrevious) -
        smoothedInPhase * (smoothedQuadrature - this.smoothedQuadraturePrevious);
      this.smoothedInPhasePrevious = smoothedInPhase;
      this.smoothedQuadraturePrevious = smoothedQuadrature;

      const periodPrevious = this._period;
      const periodNew = 2 * Math.PI * (smoothedInPhase * smoothedInPhase +
        smoothedQuadrature * smoothedQuadrature) / discriminator;
      if (!Number.isNaN(periodNew) && Number.isFinite(periodNew)) {
        this._period = periodNew;
      }

      this._period = adjustPeriod(this._period, periodPrevious);
      if (this.smoothingLengthPlus3HtLengthMin2 < this._count) { // count == 24
        this._period = this.emaPeriod(this._period, periodPrevious);
        this.isPrimed = true;
      }
    }
  }
}
