/** Enumerates outputs of the CoronaSpectrum indicator. */
export enum CoronaSpectrumOutput {
  /** The Corona spectrum heatmap column (decibels across the filter bank). */
  Value = 0,
  /** The weighted-center-of-gravity dominant cycle estimate. */
  DominantCycle = 1,
  /** The 5-sample median of DominantCycle. */
  DominantCycleMedian = 2,
}
