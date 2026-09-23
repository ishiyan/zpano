/** Enumerates outputs of the Stochastic Momentum Index indicator. */
export enum StochasticMomentumIndexOutput {

  /** The Stochastic Momentum Index oscillator value (range [-100, +100]). */
  SMIValue = 0,

  /** The signal-line value: the ul-period EMA of the oscillator. */
  SignalValue = 1,
}
