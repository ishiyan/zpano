/** Enumerates outputs of the MACD Index indicator. */
export enum MacdIndexOutput {

  /** The MACD Index line value, in raw price units (unbounded). */
  MACDIValue = 0,

  /** The signal-line value: the ul-period EMA of the index. */
  SignalValue = 1,
}
