/**
 * Defuzzification utilities.
 *
 * Provides alpha-cut conversion from continuous fuzzy output back to crisp
 * discrete values for backward compatibility with TA-Lib-style integer outputs.
 */

/**
 * Convert a continuous fuzzy output to a crisp discrete value.
 *
 * The confidence is abs(value) / scale. If confidence ≥ alpha,
 * the output is rounded to the nearest multiple of scale with the
 * original sign preserved. Otherwise 0 is returned.
 */
export function alphaCut(
    value: number, alpha: number = 0.5, scale: number = 100.0
): number {
    if (scale <= 0.0) return 0;
    const confidence = Math.abs(value) / scale;
    if (confidence < alpha - 1e-10) return 0;
    const sign = value >= 0 ? 1 : -1;
    // Round to nearest multiple of scale.
    const level = Math.max(1, Math.round(confidence));
    return sign * Math.round(level * scale);
}
