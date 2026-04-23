/** Enumerates outputs of the Bollinger Bands indicator. */
export enum BollingerBandsOutput {

  /** The lower band value. */
  LowerValue = 0,

  /** The middle band (moving average) value. */
  MiddleValue = 1,

  /** The upper band value. */
  UpperValue = 2,

  /** The band width value. */
  BandWidth = 3,

  /** The percent band (%B) value. */
  PercentBand = 4,

  /** The lower/upper band. */
  Band = 5,
}
