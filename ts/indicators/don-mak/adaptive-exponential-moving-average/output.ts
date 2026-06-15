/** Enumerates outputs of the Adaptive Exponential Moving Average indicator. */
export enum AdaptiveExponentialMovingAverageOutput {

  /** The adaptively smoothed price value. */
  Value = 0,

  /** The instantaneous frequency estimate (may be NaN). */
  Omega = 1,

  /** The smoothing factor used for the bar. */
  Alpha = 2,
}
