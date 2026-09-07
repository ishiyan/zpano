/** Enumerates outputs of the Mean Deviation Index indicator. */
export enum MeanDeviationIndexOutput {

  /** The Mean Deviation Index line value, in raw price units (unbounded). */
  MDIValue = 0,

  /** The signal-line value: the ul-period EMA of the index. */
  SignalValue = 1,
}
