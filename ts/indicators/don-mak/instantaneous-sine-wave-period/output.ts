/** Enumerates outputs of the Instantaneous Sine Wave Period indicator. */
export enum InstantaneousSineWavePeriodOutput {

  /** The estimated cycle period in bars (may be NaN). */
  Period = 0,

  /** The circular frequency in radians/bar (may be NaN). */
  Omega = 1,

  /** The wave velocity (may be NaN). */
  Velocity = 2,

  /** The wave acceleration (may be NaN). */
  Acceleration = 3,

  /** The estimated sine wave amplitude (may be NaN). */
  Amplitude = 4,

  /** The phase angle in radians (may be NaN). */
  Phase = 5,

  /** The constant level D (may be NaN). */
  DcLevel = 6,
}
