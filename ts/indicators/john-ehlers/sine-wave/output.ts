/** Enumerates outputs of the SineWave indicator. */
export enum SineWaveOutput {
  /** The sine wave value, sin(phase·Deg2Rad). */
  Value = 0,
  /** The sine wave lead value, sin((phase+45)·Deg2Rad). */
  Lead = 1,
  /** The band formed by the sine wave (upper) and the lead sine wave (lower). */
  Band = 2,
  /** The smoothed dominant cycle period. */
  DominantCyclePeriod = 3,
  /** The dominant cycle phase, in degrees. */
  DominantCyclePhase = 4,
}
