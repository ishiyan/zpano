import { HilbertTransformerCycleEstimator } from './cycle-estimator';
import { HilbertTransformerCycleEstimatorParams } from './cycle-estimator-params';
import {
  defaultMinPeriod, defaultMaxPeriod, htLength, quadratureIndex,
  push, correctAmplitude, ht, fillWmaFactors, verifyParameters
} from './common';

const accumulationLength = 40;

function calculateDifferentialPhase(phase: number, phasePrevious: number): number {
  const twoPi = 2 * Math.PI;
  const piOver2 = Math.PI / 2;
  const threePiOver4 = 3 * Math.PI / 4;
  const minDeltaPhase = twoPi / defaultMaxPeriod;
  const maxDeltaPhase = twoPi / defaultMinPeriod;

  // Compute a differential phase.
  let deltaPhase = phasePrevious - phase;

  // Resolve phase wraparound from 1st quadrant to 4th quadrant.
  if (phasePrevious < piOver2 && phase > threePiOver4) {
    deltaPhase += twoPi;
  }

  /*while (deltaPhase < 0) {
    deltaPhase += twoPi;
  }*/

  // Limit deltaPhase to be within [minDeltaPhase, maxDeltaPhase],
  // i.e. within the bounds of [minPeriod, maxPeriod] sample cycles.
  if (deltaPhase < minDeltaPhase) {
    deltaPhase = minDeltaPhase;
  } else if (deltaPhase > maxDeltaPhase) {
    deltaPhase = maxDeltaPhase;
  }

  return deltaPhase;
}

function instantaneousPhase(smoothedInPhase: number, smoothedQuadrature: number, phasePrevious: number): number {
  // Use arctangent to compute the instantaneous phase in radians.
  let phase = Math.atan(Math.abs(smoothedQuadrature / smoothedInPhase));
  if (Number.isNaN(phase) || !Number.isFinite(phase)) {
    return phasePrevious
  }

  // Resolve the ambiguity for quadrants 2, 3, and 4.
  if (smoothedInPhase < 0) {
    if (smoothedQuadrature > 0) {
      phase = Math.PI - phase; // 2nd quadrant.
    } else if (smoothedQuadrature < 0) {
      phase = Math.PI + phase; // 3rd quadrant.
    }
  } else if (smoothedInPhase > 0 && smoothedQuadrature < 0) {
    phase = 2 * Math.PI - phase; // 4th quadrant.
  }

  return phase;
}

function instantaneousPeriod(deltaPhase: number[], periodPrevious: number): number {
  const twoPi = 2 * Math.PI;
  let sumPhase = 0;
  let period = 0;

  for (let i = 0; i < accumulationLength; ++i) {
    sumPhase += deltaPhase[i];
    if (sumPhase >= twoPi) {
      period = i + 1;
      break;
    }
  }

  // Resolve instantaneous period errors.
  if (period === 0) {
    return periodPrevious;
  }

  return period;
}

/** A Hilbert transformer of WMA-smoothed and detrended data followed by the
  * phase accumulation to determine the instant period.
  *
  *  John Ehlers, Rocket Science for Traders, Wiley, 2001, 0471405671, pp 63-66.
  */
export class HilbertTransformerPhaseAccumulator implements HilbertTransformerCycleEstimator {

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
  public get quadrature(): number { return this._quadrature; }

  /** The current InPhase component value. */
  public get inPhase(): number { return this._inPhase; }

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
  private readonly smoothingLengthPlus2HtLengthMin1: number;
  private readonly smoothingLengthPlus2HtLength: number;

  private readonly oneMinAlphaEmaQuadratureInPhase: number;
  private readonly oneMinAlphaEmaPeriod: number;

  private readonly rawValues: Array<number>;
  private readonly wmaFactors: Array<number>;
  private readonly wmaSmoothed: Array<number> = new Array(htLength).fill(0);
  private readonly _detrended: Array<number> = new Array(htLength).fill(0);
  private readonly deltaPhase: Array<number> = new Array(accumulationLength).fill(0);

  private _inPhase: number = 0;
  private _quadrature: number = 0;
  private _count: number = 0;
  private smoothedInPhasePrevious: number = 0;
  private smoothedQuadraturePrevious: number = 0;
  private phasePrevious: number = 0;
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
    this.smoothingLengthPlus2HtLengthMin1 = this.smoothingLengthPlus2HtLengthMin2 + 1
    this.smoothingLengthPlus2HtLength = this.smoothingLengthPlus2HtLengthMin1 + 1;

    this.rawValues = new Array(length).fill(0);
    this.wmaFactors = new Array(length);
    fillWmaFactors(length, this.wmaFactors);

    if (params.warmUpPeriod && params.warmUpPeriod > this.smoothingLengthPlus2HtLength) {
      this.warmUpPeriod = params.warmUpPeriod;
    } else {
      this.warmUpPeriod = this.smoothingLengthPlus2HtLength;
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
      this._quadrature = ht(this._detrended) * amplitudeCorrectionFactor;
      this._inPhase = this._detrended[quadratureIndex];

      // Exponential moving average smoothing.
      const smoothedInPhase = this.emaQuadratureInPhase(this._inPhase, this.smoothedInPhasePrevious);
      const smoothedQuadrature = this.emaQuadratureInPhase(this._quadrature, this.smoothedQuadraturePrevious);
      this.smoothedInPhasePrevious = smoothedInPhase;
      this.smoothedQuadraturePrevious = smoothedQuadrature;

      // Compute an instantaneous phase.
      const phase = instantaneousPhase(smoothedInPhase, smoothedQuadrature, this.phasePrevious);

      // Compute a differential phase.
      push(this.deltaPhase, calculateDifferentialPhase(phase, this.phasePrevious));
      this.phasePrevious = phase;

      // Compute an instantaneous period.
      const periodPrevious = this._period;
      this._period = instantaneousPeriod(this.deltaPhase, periodPrevious);

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

      this._quadrature = ht(this._detrended) * amplitudeCorrectionFactor; // count >= 16
      this._inPhase = this._detrended[quadratureIndex];
      if (this.smoothingLengthPlus2HtLengthMin2 === this._count) { // count == 16
        this.smoothedInPhasePrevious = this._inPhase;
        this.smoothedQuadraturePrevious = this._quadrature;
        return;
      }

      const smoothedInPhase = this.emaQuadratureInPhase(this._inPhase, this.smoothedInPhasePrevious); // count >= 17
      const smoothedQuadrature = this.emaQuadratureInPhase(this._quadrature, this.smoothedQuadraturePrevious);
      this.smoothedInPhasePrevious = smoothedInPhase;
      this.smoothedQuadraturePrevious = smoothedQuadrature;
      const phase = instantaneousPhase(smoothedInPhase, smoothedQuadrature, this.phasePrevious);
      push(this.deltaPhase, calculateDifferentialPhase(phase, this.phasePrevious));
      this.phasePrevious = phase;
      const periodPrevious = this._period;
      this._period = instantaneousPeriod(this.deltaPhase, periodPrevious);

      if (this.smoothingLengthPlus2HtLengthMin1 < this._count) { // count > 17
        this._period = this.emaPeriod(this._period, periodPrevious);
        this.isPrimed = true;
      }
    }
  }
}
