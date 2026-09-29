/**
 * Describes parameters to create an instance of the Tick Volume Indicator.
 *
 * The parameter names `r`, `s` and `u` are the canonical symbols from
 * William Blau's _Momentum, Direction, and Divergence_ (Wiley, 1995),
 * chapters 4 and 10. They are kept verbatim for fidelity with the book and
 * the test-data naming.
 *
 * The indicator consumes upticks and downticks (derived from the bar range,
 * or from consecutive prices for single-valued samples), so it has no
 * configurable price-component fields.
 */
export interface TickVolumeIndicatorParams {
    /**
     * The period of the 1st (innermost) EMA in the smoothing cascade, applied
     * to the upticks and downticks.
     *
     * The value should be greater than 0. The default value is 12.
     */
    r?: number;

    /**
     * The period of the 2nd EMA in the smoothing cascade, applied to the output
     * of the 1st EMA.
     *
     * The value should be greater than 0. The default value is 12.
     */
    s?: number;

    /**
     * The period of the 3rd (outermost) EMA in the smoothing cascade, applied
     * to the output of the 2nd EMA.
     *
     * The default `u=1` switches the 3rd stage off (passthrough), yielding the
     * book's classic double-smoothed TVI(r, s); chapter 10 uses TVI(32, 32, 5).
     * The value should be greater than 0. The default value is 1.
     */
    u?: number;
}

export function defaultParams(): TickVolumeIndicatorParams {
    return { r: 12, s: 12, u: 1 };
}
