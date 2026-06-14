/** Enumerates outputs of the Schaff Trend Cycle indicator. */
export enum SchaffTrendCycleOutput {

  /** The STC oscillator value (range [0, 100]). */
  STCValue = 0,

  /** The gated MACD line (XMAC) value. */
  MACDValue = 1,

  /** The first smoothed %D stage (PF) value. */
  PFValue = 2,
}
