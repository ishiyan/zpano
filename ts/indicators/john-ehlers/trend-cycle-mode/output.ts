/** Enumerates outputs of the TrendCycleMode indicator. */
export enum TrendCycleModeOutput {
  /** +1 in trend mode, -1 in cycle mode. */
  Value = 0,
  /** 1 if the trend mode is declared, 0 otherwise. */
  IsTrendMode = 1,
  /** 1 if the cycle mode is declared, 0 otherwise (= 1 − IsTrendMode). */
  IsCycleMode = 2,
  /** The WMA-smoothed instantaneous trend line. */
  InstantaneousTrendLine = 3,
  /** The sine wave value, sin(phase·Deg2Rad). */
  SineWave = 4,
  /** The sine wave lead value, sin((phase+45)·Deg2Rad). */
  SineWaveLead = 5,
  /** The smoothed dominant cycle period. */
  DominantCyclePeriod = 6,
  /** The dominant cycle phase, in degrees. */
  DominantCyclePhase = 7,
}
