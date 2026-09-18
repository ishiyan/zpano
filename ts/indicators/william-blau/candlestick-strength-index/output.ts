/** Enumerates outputs of the Candlestick Strength Index indicator. */
export enum CandlestickStrengthIndexOutput {

  /** The Candlestick Strength Index oscillator value (range [-100, +100]). */
  CSIValue = 0,

  /** The signal-line value: the ul-period EMA of the oscillator. */
  SignalValue = 1,
}
