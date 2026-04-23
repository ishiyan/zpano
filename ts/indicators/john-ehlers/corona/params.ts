/** Configures a Corona spectral analysis engine.
 *
 * All fields have zero-value defaults: a zero or _undefined_ value means "use the default".
 * The defaults follow Ehlers' original TASC article (November 2008).
 */
export interface CoronaParams {
  /** High-pass filter cutoff period (de-trending period), in bars. Must be >= 2.
   *
   * The default value is 30.
   */
  highPassFilterCutoff?: number;

  /** Minimum cycle period (in bars) covered by the bandpass filter bank. Must be >= 2.
   *
   * The default value is 6.
   */
  minimalPeriod?: number;

  /** Maximum cycle period (in bars) covered by the bandpass filter bank. Must be > minimalPeriod.
   *
   * The default value is 30.
   */
  maximalPeriod?: number;

  /** Filter bins with smoothed dB value at or below this threshold contribute to the
   * weighted dominant-cycle estimate.
   *
   * The default value is 6.
   */
  decibelsLowerThreshold?: number;

  /** Upper clamp on the smoothed dB value and reference value for the dominant-cycle
   * weighting (weight = upper − dB).
   *
   * The default value is 20.
   */
  decibelsUpperThreshold?: number;
}

/** Default Corona parameters (Ehlers TASC, November 2008). */
export const DefaultCoronaParams: Required<CoronaParams> = {
  highPassFilterCutoff: 30,
  minimalPeriod: 6,
  maximalPeriod: 30,
  decibelsLowerThreshold: 6,
  decibelsUpperThreshold: 20,
};
