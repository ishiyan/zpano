/**
 * Describes parameters to create an instance of the Double Smoothed Stochastic indicator.
 *
 * The parameter names `q`, `r`, `s` and `g` are the canonical symbols from
 * William Blau's _Momentum, Direction, and Divergence_ (Wiley, 1995). They are
 * kept verbatim for fidelity with the book, the MQL5 reference, and the
 * test-data naming.
 *
 * The indicator consumes the high, low and close prices of a bar, so it has no
 * configurable price-component fields.
 */
export interface DoubleSmoothedStochasticParams {
    /**
     * The stochastic look-back period: the number of bars over which the
     * highest high and the lowest low are taken.
     *
     * Setting `q=1` yields the book's one-bar HLC index. The value should be
     * greater than 0. The default value is 5.
     */
    q?: number;

    /**
     * The period of the 1st (inner) EMA in the smoothing cascade, applied to
     * the raw stochastic and the range.
     *
     * The value should be greater than 0. The default value is 7.
     */
    r?: number;

    /**
     * The period of the 2nd (outer) EMA in the smoothing cascade, applied to
     * the output of the 1st EMA.
     *
     * The value should be greater than 0. The default value is 3.
     */
    s?: number;

    /**
     * The period of the signal-line SMA, applied to the oscillator to produce
     * the second output.
     *
     * Setting `g=1` makes the signal a passthrough (signal == dss every bar).
     * The value should be greater than 0. The default value is 3.
     */
    g?: number;
}

export function defaultParams(): DoubleSmoothedStochasticParams {
    return { q: 5, r: 7, s: 3, g: 3 };
}
