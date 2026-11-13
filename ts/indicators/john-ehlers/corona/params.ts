/** Configures a Corona spectral analysis engine.
 *
 * Explicit zero values are preserved (not substituted with defaults);
 * undefined fields fall back to Ehlers' November 2008 TASC defaults.
 */
export interface CoronaParams {
    /** High-pass filter cutoff period (de-trending period), in bars (>= 5). Default: 30. */
    highPassFilterCutoff?: number;

    /** Minimum cycle period (in bars) covered by the bandpass filter bank (>= 6). Default: 6. */
    minimalPeriod?: number;

    /** Maximum cycle period (in bars) covered by the bandpass filter bank (> minimalPeriod). Default: 30. */
    maximalPeriod?: number;

    /** Smoothed dB selection threshold (0 <= lower < upper). Default: 6. */
    decibelsLowerThreshold?: number;

    /** Upper dB clamp and centroid weight ceiling (> 0 and > lower). Default: 20. */
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

export function defaultParams(): CoronaParams {
    return { ...DefaultCoronaParams };
}
