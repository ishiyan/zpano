/** Describes a common Hilbert transformer cycle estimator functionality. */
export interface HilbertTransformerCycleEstimator {
  /** The underlying linear-Weighted Moving Average (WMA) smoothing length. */
  readonly smoothingLength: number; // integer

  /** The current WMA-smoothed value used by underlying Hilbert transformer.
   * 
   * The linear-Weighted Moving Average has a window size of __smoothingLength__.
   */
  readonly smoothed: number;

  /** The current de-trended value. */
  readonly detrended: number;

  /** The current Quadrature component value. */
  readonly quadrature: number;

  /** The current InPhase component value. */
  readonly inPhase: number;

  /** The current period value. */
  readonly period: number;

  /** The current count value. */
  readonly count: number; // integer

  /** Indicates whether an estimator is primed. */
  readonly primed: boolean;

  /** The minimal cycle period supported by this Hilbert transformer. */
  readonly minPeriod: number; // integer

  /** The maximual cycle period supported by this Hilbert transformer. */
  readonly maxPeriod: number; // integer

  /** The value of α (0 < α ≤ 1) used in EMA to smooth the in-phase and quadrature components. */
  readonly alphaEmaQuadratureInPhase: number;

  /** The value of α (0 < α ≤ 1) used in EMA to smooth the instantaneous period. */
  readonly alphaEmaPeriod: number;

  /** The number of updates before the estimator is primed (MaxPeriod * 2 = 100). */
  readonly warmUpPeriod: number; // integer

  /** Updates the estimator given the next sample value. */
  update(sample: number): void;
}
