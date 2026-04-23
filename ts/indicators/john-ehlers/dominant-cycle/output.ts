/** Enumerates outputs of the DominantCycle indicator. */
export enum DominantCycleOutput {
  /** The raw instantaneous cycle period produced by the Hilbert transformer estimator. */
  RawPeriod = 0,
  /** The dominant cycle period obtained by additional EMA smoothing of the raw period. */
  Period = 1,
  /** The dominant cycle phase, in degrees. */
  Phase = 2,
}
