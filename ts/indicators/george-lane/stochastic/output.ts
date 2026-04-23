/** Enumerates outputs of the Stochastic Oscillator indicator. */
export enum StochasticOutput {

  /** The Fast-K line (raw stochastic). */
  FastK = 0,

  /** The Slow-K line (smoothed Fast-K, also known as Fast-D). */
  SlowK = 1,

  /** The Slow-D line (smoothed Slow-K). */
  SlowD = 2,
}
