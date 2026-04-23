/** Enumerates outputs of the AdaptiveTrendAndCycleFilter indicator. */
export enum AdaptiveTrendAndCycleFilterOutput {
  /** Fast Adaptive Trend Line (39-tap FIR). */
  Fatl = 0,
  /** Slow Adaptive Trend Line (65-tap FIR). */
  Satl = 1,
  /** Reference Fast Trend Line (44-tap FIR). */
  Rftl = 2,
  /** Reference Slow Trend Line (91-tap FIR). */
  Rstl = 3,
  /** Range Bound Channel Index (56-tap FIR). */
  Rbci = 4,
  /** Fast Trend Line Momentum (FATL − RFTL). */
  Ftlm = 5,
  /** Slow Trend Line Momentum (SATL − RSTL). */
  Stlm = 6,
  /** Perfect Commodity Channel Index (sample − FATL). */
  Pcci = 7,
}
