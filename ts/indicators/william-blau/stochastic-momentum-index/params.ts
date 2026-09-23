/**
 * Describes parameters to create an instance of the Stochastic Momentum Index indicator.
 *
 * The parameter names `q`, `r`, `s`, `u` and `ul` are the canonical symbols from
 * William Blau's _Momentum, Direction, and Divergence_ (Wiley, 1995), chapter 3.
 * They are kept verbatim for fidelity with the book, the MQL5 reference, and the
 * test-data naming.
 *
 * The indicator consumes the high, low and close prices of a bar, so it has no
 * configurable price-component fields.
 */
export interface StochasticMomentumIndexParams {
    /**
     * The stochastic look-back period: the number of bars over which the
     * highest high and the lowest low are taken.
     *
     * Setting `q=1` yields the book's one-day stochastic. The value should be
     * greater than 0. The default value is 5.
     */
    q?: number;

    /**
     * The period of the 1st (innermost) EMA in the smoothing cascade, applied
     * to the stochastic momentum and the half-range.
     *
     * The value should be greater than 0. The default value is 20.
     */
    r?: number;

    /**
     * The period of the 2nd EMA in the smoothing cascade, applied to the output
     * of the 1st EMA.
     *
     * The value should be greater than 0. The default value is 5.
     */
    s?: number;

    /**
     * The period of the 3rd (outermost) EMA in the smoothing cascade, applied
     * to the output of the 2nd EMA.
     *
     * Setting `u=1` switches the 3rd stage off (passthrough), yielding the
     * book's classic double-smoothed oscillator. The value should be greater
     * than 0. The default value is 3.
     */
    u?: number;

    /**
     * The period of the signal-line EMA, applied to the oscillator to produce
     * the second output (Blau's Ergodic signal line).
     *
     * Setting `ul=1` makes the signal a passthrough (signal == smi every bar).
     * The value should be greater than 0. The default value is 3.
     */
    ul?: number;
}

export function defaultParams(): StochasticMomentumIndexParams {
    return { q: 5, r: 20, s: 5, u: 3, ul: 3 };
}
