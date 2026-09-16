/** Enumerates outputs of the Candlestick Momentum Index indicator. */
export enum CandlestickMomentumIndexOutput {

  /** The Candlestick Momentum Index oscillator value (range [-100, +100]). */
  CMIValue = 0,

  /** The signal-line value: the ul-period EMA of the oscillator. */
  SignalValue = 1,
}
