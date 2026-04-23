/** Enumerates outputs of the MesaAdaptiveMovingAverage indicator. */
export enum MesaAdaptiveMovingAverageOutput {
  /** The scalar value of the MAMA (Mesa Adaptive Moving Average). */
  Value = 0,
  /** The scalar value of the FAMA (Following Adaptive Moving Average). */
  Fama = 1,
  /** The band output, with MAMA as the upper line and FAMA as the lower line. */
  Band = 2,
}
