/** Enumerates outputs of the Ergodic Oscillator indicator. */
export enum ErgodicOscillatorOutput {

  /** The Ergodic oscillator value (the True Strength Index, range [-100, +100]). */
  ErgodicValue = 0,

  /** The signal-line value: the ul-period EMA of the oscillator. */
  SignalValue = 1,
}
