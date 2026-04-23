/** Classifies the minimum input data type an indicator consumes. */
export enum InputRequirement {
  /** Consumes a scalar time series (e.g., prices). */
  ScalarInput = 1,

  /** Consumes level-1 quotes. */
  QuoteInput,

  /** Consumes OHLCV bars. */
  BarInput,

  /** Consumes individual trades. */
  TradeInput
}
