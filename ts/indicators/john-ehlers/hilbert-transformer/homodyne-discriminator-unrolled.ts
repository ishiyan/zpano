import { HilbertTransformerCycleEstimator } from './cycle-estimator';
import { HilbertTransformerCycleEstimatorParams } from './cycle-estimator-params';
import { defaultMinPeriod, defaultMaxPeriod, verifyParameters } from './common';

// The TA-Lib implementation uses the following lookback value with hardcoded smoothingLength=4.
//
// The fixed lookback is 32 and is establish as follows:
// 12 price bar to be compatible with the implementation of Tradestation found in John Ehlers book,
// 6 price bars for the Detrender,
// 6 price bars for Q1,
// 3 price bars for jI,
// 3 price bars for jQ,
// 1 price bar for Re/Im,
// 1 price bar for the Delta Phase,
// --------
// 32 Total.
//
// The first 9 bars are not used by TA-Lib, they are just skipped for the compatibility with the Tradestation.
// We do not skip them. Thus, we use the fixed lookback value 32 - 9 = 23.
const primedCount = 23

/** A Hilbert transformer of WMA-smoothed and detrended data with the Homodyne Discriminator applied.
  *
  * Copied from the TA-Lib implementation with unrolled loops.
  * 
  *  John Ehlers, Rocket Science for Traders, Wiley, 2001, 0471405671, pp 52-77.
  */
export class HilbertTransformerHomodyneDiscriminatorUnrolled implements HilbertTransformerCycleEstimator {

  /** The underlying linear-Weighted Moving Average (WMA) smoothing length. */
  public readonly smoothingLength: number;

  /** The current WMA-smoothed value used by underlying Hilbert transformer.
   * 
   * The linear-Weighted Moving Average has a window size of __smoothingLength__.
   */
  public get smoothed(): number { return this._smoothed; }

  /** The current de-trended value. */
  public get detrended(): number { return this._detrended; }

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

  private readonly oneMinAlphaEmaQuadratureInPhase: number;
  private readonly oneMinAlphaEmaPeriod: number;

  private _smoothed: number = 0;
  private _detrended: number = 0;
  private _inPhase: number = 0;
  private _quadrature: number = 0;
  private smoothingMultiplier: number;
  private adjustedPeriod: number = 0;
  private _count: number = 0;
  private index: number = 0;
  private i2Previous: number = 0;
  private q2Previous: number = 0;
  private re: number = 0;
  private im: number = 0;
  private _period: number = defaultMinPeriod;
  private isPrimed = false;
  private isWarmedUp = false;

  // WMA smoother private members.
  private wmaSum: number = 0;
  private wmaSub: number = 0;
  private wmaInput1: number = 0;
  private wmaInput2: number = 0;
  private wmaInput3: number = 0;
  private wmaInput4: number = 0;

  // Detrender private members.
  private detrenderOdd0: number = 0;
  private detrenderOdd1: number = 0;
  private detrenderOdd2: number = 0;
  private detrenderPreviousOdd: number = 0;
  private detrenderPreviousInputOdd: number = 0;
  private detrenderEven0: number = 0;
  private detrenderEven1: number = 0;
  private detrenderEven2: number = 0;
  private detrenderPreviousEven: number = 0;
  private detrenderPreviousInputEven: number = 0;

  // Quadrature (Q1) component private members.
  private q1Odd0: number = 0;
  private q1Odd1: number = 0;
  private q1Odd2: number = 0;
  private q1PreviousOdd: number = 0;
  private q1PreviousInputOdd: number = 0;
  private q1Even0: number = 0;
  private q1Even1: number = 0;
  private q1Even2: number = 0;
  private q1PreviousEven: number = 0;
  private q1PreviousInputEven: number = 0;

  // InPhase (I1) private members.
  private i1Previous1Odd: number = 0;
  private i1Previous2Odd: number = 0;
  private i1Previous1Even: number = 0;
  private i1Previous2Even: number = 0;

  // jI private members
  private jiOdd0: number = 0;
  private jiOdd1: number = 0;
  private jiOdd2: number = 0;
  private jiPreviousOdd: number = 0;
  private jiPreviousInputOdd: number = 0;
  private jiEven0: number = 0;
  private jiEven1: number = 0;
  private jiEven2: number = 0;
  private jiPreviousEven: number = 0;
  private jiPreviousInputEven: number = 0;

  // jQ private members.
  private jqOdd0: number = 0;
  private jqOdd1: number = 0;
  private jqOdd2: number = 0;
  private jqPreviousOdd: number = 0;
  private jqPreviousInputOdd: number = 0;
  private jqEven0: number = 0;
  private jqEven1: number = 0;
  private jqEven2: number = 0;
  private jqPreviousEven: number = 0;
  private jqPreviousInputEven: number = 0;

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

    this.smoothingMultiplier = 1 / 3;
    if (length === 4) {
      this.smoothingMultiplier = 1 / 10;
    } else if (length === 3) {
      this.smoothingMultiplier = 1 / 6;
    }

    this.warmUpPeriod = primedCount;
    if (params.warmUpPeriod && params.warmUpPeriod > primedCount) {
      this.warmUpPeriod = params.warmUpPeriod;
    }
  }

  /** Updates the estimator given the next sample value. */
  public update(sample: number): void {
    if (Number.isNaN(sample)) {
      return;
    }

    const a = 0.0962;
    const b = 0.5769;
    const minPreviousPeriodFactor = 0.67;
    const maxPreviousPeriodFactor = 1.5;
    const c0075 = 0.075;
    const c054 = 0.54;

    let value = 0;

    // WMA
    // We need (smoothingLength - 1) bars to accumulate the WMA sub and sum.
    // On (smoothingLength)-th bar we calculate the first WMA smoothed value and begin with detrender.
    do {
      if (this.smoothingLength >= ++this._count) {
        if (1 === this._count) {
          this.wmaSub = sample; this.wmaInput1 = sample; this.wmaSum = sample;
        } else if (2 === this._count) {
          this.wmaSub += sample; this.wmaInput2 = sample; this.wmaSum += sample * 2;
          if (2 === this.smoothingLength) {
            value = this.wmaSum * this.smoothingMultiplier;
            break; // DetrendLabel
          }
        } else if (3 === this._count) {
          this.wmaSub += sample; this.wmaInput3 = sample; this.wmaSum += sample * 3;
          if (3 === this.smoothingLength) {
            value = this.wmaSum * this.smoothingMultiplier;
            break; // DetrendLabel
          }
        } else { //if (4 === count)
          this.wmaSub += sample; this.wmaInput4 = sample; this.wmaSum += sample * 4;
          value = this.wmaSum * this.smoothingMultiplier;
          break; // DetrendLabel
        }
        return;
      }

      this.wmaSum -= this.wmaSub;
      this.wmaSum += sample * this.smoothingLength;
      value = this.wmaSum * this.smoothingMultiplier;
      this.wmaSub += sample; this.wmaSub -= this.wmaInput1;
      this.wmaInput1 = this.wmaInput2;
      if (4 === this.smoothingLength) {
        this.wmaInput2 = this.wmaInput3; this.wmaInput3 = this.wmaInput4; this.wmaInput4 = sample;
      } else if (3 === this.smoothingLength) {
        this.wmaInput2 = this.wmaInput3; this.wmaInput3 = sample;
      } else { //if (2 == smoothingLength)
        this.wmaInput2 = sample;
      }
      // eslint-disable-next-line no-constant-condition
    } while (false);
    // DetrendLabel:
    // Detrender.
    this._smoothed = value;
    if (!this.isWarmedUp) {
      this.isWarmedUp = this._count > this.warmUpPeriod;
      if (!this.isPrimed) {
        this.isPrimed = this._count > primedCount;
      }
    }
    let detrender, ji, jq;
    let temp = a * this._smoothed; this.adjustedPeriod = c0075 * this._period + c054;
    if (0 === this._count % 2) { // Even value count.
      // Explicitely expanded index.
      if (0 === this.index) {
        this.index = 1;
        detrender = -this.detrenderEven0; this.detrenderEven0 = temp; detrender += temp; detrender -= this.detrenderPreviousEven;
        this.detrenderPreviousEven = b * this.detrenderPreviousInputEven; this.detrenderPreviousInputEven = value;
        detrender += this.detrenderPreviousEven; detrender *= this.adjustedPeriod;
        // Quadrature component.
        temp = a * detrender; this._quadrature = -this.q1Even0; this.q1Even0 = temp; this._quadrature += temp; this._quadrature -= this.q1PreviousEven;
        this.q1PreviousEven = b * this.q1PreviousInputEven; this.q1PreviousInputEven = detrender;
        this._quadrature += this.q1PreviousEven; this._quadrature *= this.adjustedPeriod;
        // Advance the phase of the InPhase component by 90°.
        temp = a * this.i1Previous2Even; ji = -this.jiEven0; this.jiEven0 = temp; ji += temp; ji -= this.jiPreviousEven;
        this.jiPreviousEven = b * this.jiPreviousInputEven; this.jiPreviousInputEven = this.i1Previous2Even;
        ji += this.jiPreviousEven; ji *= this.adjustedPeriod;
        // Advance the phase of the Quadrature component by 90°.
        temp = a * this._quadrature; jq = -this.jqEven0; this.jqEven0 = temp;
      } else if (1 === this.index) {
        this.index = 2;
        detrender = -this.detrenderEven1; this.detrenderEven1 = temp; detrender += temp; detrender -= this.detrenderPreviousEven;
        this.detrenderPreviousEven = b * this.detrenderPreviousInputEven; this.detrenderPreviousInputEven = value;
        detrender += this.detrenderPreviousEven; detrender *= this.adjustedPeriod;
        // Quadrature component.
        temp = a * detrender; this._quadrature = -this.q1Even1; this.q1Even1 = temp; this._quadrature += temp; this._quadrature -= this.q1PreviousEven;
        this.q1PreviousEven = b * this.q1PreviousInputEven; this.q1PreviousInputEven = detrender;
        this._quadrature += this.q1PreviousEven; this._quadrature *= this.adjustedPeriod;
        // Advance the phase of the InPhase component by 90°.
        temp = a * this.i1Previous2Even; ji = -this.jiEven1; this.jiEven1 = temp; ji += temp; ji -= this.jiPreviousEven;
        this.jiPreviousEven = b * this.jiPreviousInputEven; this.jiPreviousInputEven = this.i1Previous2Even;
        ji += this.jiPreviousEven; ji *= this.adjustedPeriod;
        // Advance the phase of the Quadrature component by 90°.
        temp = a * this._quadrature; jq = -this.jqEven1; this.jqEven1 = temp;
      } else { //if (2 == index)
        this.index = 0;
        detrender = -this.detrenderEven2; this.detrenderEven2 = temp; detrender += temp; detrender -= this.detrenderPreviousEven;
        this.detrenderPreviousEven = b * this.detrenderPreviousInputEven; this.detrenderPreviousInputEven = value;
        detrender += this.detrenderPreviousEven; detrender *= this.adjustedPeriod;
        // Quadrature component.
        temp = a * detrender; this._quadrature = -this.q1Even2; this.q1Even2 = temp; this._quadrature += temp; this._quadrature -= this.q1PreviousEven;
        this.q1PreviousEven = b * this.q1PreviousInputEven; this.q1PreviousInputEven = detrender;
        this._quadrature += this.q1PreviousEven; this._quadrature *= this.adjustedPeriod;
        // Advance the phase of the InPhase component by 90°.
        temp = a * this.i1Previous2Even; ji = -this.jiEven2; this.jiEven2 = temp; ji += temp; ji -= this.jiPreviousEven;
        this.jiPreviousEven = b * this.jiPreviousInputEven; this.jiPreviousInputEven = this.i1Previous2Even;
        ji += this.jiPreviousEven; ji *= this.adjustedPeriod;
        // Advance the phase of the Quadrature component by 90°.
        temp = a * this._quadrature; jq = -this.jqEven2; this.jqEven2 = temp;
      }
      // Advance the phase of the Quadrature component by 90° (continued).
      jq += temp; jq -= this.jqPreviousEven;
      this.jqPreviousEven = b * this.jqPreviousInputEven; this.jqPreviousInputEven = this._quadrature;
      jq += this.jqPreviousEven; jq *= this.adjustedPeriod;
      // InPhase component.
      this._inPhase = this.i1Previous2Even;
      // The current detrender value will be used by the "odd" logic later.
      this.i1Previous2Odd = this.i1Previous1Odd;
      this.i1Previous1Odd = detrender;
    } else { // Odd value count.
      // Explicitely expanded index.
      if (0 === this.index) {
        this.index = 1;
        detrender = -this.detrenderOdd0; this.detrenderOdd0 = temp; detrender += temp; detrender -= this.detrenderPreviousOdd;
        this.detrenderPreviousOdd = b * this.detrenderPreviousInputOdd; this.detrenderPreviousInputOdd = value;
        detrender += this.detrenderPreviousOdd; detrender *= this.adjustedPeriod;
        // Quadrature component.
        temp = a * detrender; this._quadrature = -this.q1Odd0; this.q1Odd0 = temp; this._quadrature += temp; this._quadrature -= this.q1PreviousOdd;
        this.q1PreviousOdd = b * this.q1PreviousInputOdd; this.q1PreviousInputOdd = detrender;
        this._quadrature += this.q1PreviousOdd; this._quadrature *= this.adjustedPeriod;
        // Advance the phase of the InPhase component by 90°.
        temp = a * this.i1Previous2Odd; ji = -this.jiOdd0; this.jiOdd0 = temp; ji += temp; ji -= this.jiPreviousOdd;
        this.jiPreviousOdd = b * this.jiPreviousInputOdd; this.jiPreviousInputOdd = this.i1Previous2Odd;
        ji += this.jiPreviousOdd; ji *= this.adjustedPeriod;
        // Advance the phase of the Quadrature component by 90°.
        temp = a * this._quadrature; jq = -this.jqOdd0; this.jqOdd0 = temp;
      } else if (1 === this.index) {
        this.index = 2;
        // Quadrature component.
        detrender = -this.detrenderOdd1; this.detrenderOdd1 = temp; detrender += temp; detrender -= this.detrenderPreviousOdd;
        this.detrenderPreviousOdd = b * this.detrenderPreviousInputOdd; this.detrenderPreviousInputOdd = value;
        detrender += this.detrenderPreviousOdd; detrender *= this.adjustedPeriod;
        temp = a * detrender; this._quadrature = -this.q1Odd1; this.q1Odd1 = temp; this._quadrature += temp; this._quadrature -= this.q1PreviousOdd;
        this.q1PreviousOdd = b * this.q1PreviousInputOdd; this.q1PreviousInputOdd = detrender;
        this._quadrature += this.q1PreviousOdd; this._quadrature *= this.adjustedPeriod;
        // Advance the phase of the InPhase component by 90°.
        temp = a * this.i1Previous2Odd; ji = -this.jiOdd1; this.jiOdd1 = temp; ji += temp; ji -= this.jiPreviousOdd;
        this.jiPreviousOdd = b * this.jiPreviousInputOdd; this.jiPreviousInputOdd = this.i1Previous2Odd;
        ji += this.jiPreviousOdd; ji *= this.adjustedPeriod;
        // Advance the phase of the Quadrature component by 90°.
        temp = a * this._quadrature; jq = -this.jqOdd1; this.jqOdd1 = temp;
      } else { //if (2 === index)
        this.index = 0;
        detrender = -this.detrenderOdd2; this.detrenderOdd2 = temp; detrender += temp; detrender -= this.detrenderPreviousOdd;
        this.detrenderPreviousOdd = b * this.detrenderPreviousInputOdd; this.detrenderPreviousInputOdd = value;
        detrender += this.detrenderPreviousOdd; detrender *= this.adjustedPeriod;
        // Quadrature component.
        temp = a * detrender; this._quadrature = -this.q1Odd2; this.q1Odd2 = temp; this._quadrature += temp; this._quadrature -= this.q1PreviousOdd;
        this.q1PreviousOdd = b * this.q1PreviousInputOdd; this.q1PreviousInputOdd = detrender;
        this._quadrature += this.q1PreviousOdd; this._quadrature *= this.adjustedPeriod;
        // Advance the phase of the InPhase component by 90°.
        temp = a * this.i1Previous2Odd; ji = -this.jiOdd2; this.jiOdd2 = temp; ji += temp; ji -= this.jiPreviousOdd;
        this.jiPreviousOdd = b * this.jiPreviousInputOdd; this.jiPreviousInputOdd = this.i1Previous2Odd;
        ji += this.jiPreviousOdd; ji *= this.adjustedPeriod;
        // Advance the phase of the Quadrature component by 90°.
        temp = a * this._quadrature; jq = -this.jqOdd2; this.jqOdd2 = temp;
      }
      // Advance the phase of the Quadrature component by 90° (continued).
      jq += temp; jq -= this.jqPreviousOdd;
      this.jqPreviousOdd = b * this.jqPreviousInputOdd; this.jqPreviousInputOdd = this._quadrature;
      jq += this.jqPreviousOdd; jq *= this.adjustedPeriod;
      // InPhase component.
      this._inPhase = this.i1Previous2Odd;
      // The current detrender value will be used by the "even" logic later.
      this.i1Previous2Even = this.i1Previous1Even;
      this.i1Previous1Even = detrender;
    }

    this._detrended = detrender;
    // Phasor addition for 3 bar averaging.
    let i2 = this._inPhase - jq;
    let q2 = this._quadrature + ji;
    // Smooth the InPhase and the Quadrature components before applying the discriminator.
    i2 = this.alphaEmaQuadratureInPhase * i2 + this.oneMinAlphaEmaQuadratureInPhase * this.i2Previous;
    q2 = this.alphaEmaQuadratureInPhase * q2 + this.oneMinAlphaEmaQuadratureInPhase * this.q2Previous;
    // Homodyne discriminator.
    // Homodyne means we are multiplying the signal by itself.
    // We multiply the signal of the current bar with the complex conjugate of the signal 1 bar ago.
    this.re = this.alphaEmaQuadratureInPhase * (i2 * this.i2Previous + q2 * this.q2Previous) + this.oneMinAlphaEmaQuadratureInPhase * this.re;
    this.im = this.alphaEmaQuadratureInPhase * (i2 * this.q2Previous - q2 * this.i2Previous) + this.oneMinAlphaEmaQuadratureInPhase * this.im;
    this.q2Previous = q2;
    this.i2Previous = i2;
    temp = this._period;
    const periodNew = 2 * Math.PI / Math.atan2(this.im, this.re);
    if (!Number.isNaN(periodNew) && Number.isFinite(periodNew)) {
      this._period = periodNew;
    }
    value = maxPreviousPeriodFactor * temp;
    if (this._period > value) {
      this._period = value;
    } else {
      value = minPreviousPeriodFactor * temp;
      if (this._period < value) {
        this._period = value;
      }
    }
    if (this._period < this.minPeriod) {
      this._period = this.minPeriod;
    } else if (this._period > this.maxPeriod) {
      this._period = this.maxPeriod;
    }
    this._period = this.alphaEmaPeriod * this._period + this.oneMinAlphaEmaPeriod * temp;
  }
}
