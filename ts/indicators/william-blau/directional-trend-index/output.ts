/** Enumerates outputs of the Directional Trend Index indicator. */
export enum DirectionalTrendIndexOutput {

  /** The Directional Trend Index oscillator value (range [-100, +100]). */
  DTIValue = 0,

  /** The signal-line value: the ul-period EMA of the oscillator. */
  SignalValue = 1,
}
