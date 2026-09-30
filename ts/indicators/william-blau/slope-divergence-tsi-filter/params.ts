import { BarComponent } from '../../../entities/bar-component';
import { QuoteComponent } from '../../../entities/quote-component';
import { TradeComponent } from '../../../entities/trade-component';

/**
 * Describes parameters to create an instance of the Slope Divergence TSI Filter indicator.
 *
 * The parameter names `q`, `r`, `s`, `u`, `x` and `y` are the canonical symbols from
 * William Blau's Momentum, Direction, and Divergence (Wiley, 1995), chapter 12 and
 * Appendix B, Figure B-25. They are kept verbatim for fidelity with the book and the
 * test-data naming.
 */
export interface SlopeDivergenceTsiFilterParams {
    /**
     * The TSI momentum look-back period; momentum is `C_k - C_(k-(q-1))`.
     *
     * The look-back distance is `q-1` bars, so `q=2` is the one-bar momentum Blau
     * uses throughout the book. The value should be greater than 0. The default
     * value is 2.
     */
    q?: number;

    /**
     * The period of the 1st (innermost) EMA of the TSI smoothing cascade.
     *
     * The value should be greater than 0. The default value is 32.
     */
    r?: number;

    /**
     * The period of the 2nd EMA of the TSI smoothing cascade.
     *
     * The value should be greater than 0. The default value is 32.
     */
    s?: number;

    /**
     * The period of the 3rd (outermost) EMA of the TSI smoothing cascade.
     *
     * Setting `u=1` switches the 3rd stage off (passthrough), yielding the
     * double-smoothed TSI of the book's raw form (Fig. 12-1). The value should be
     * greater than 0. The default value is 7.
     */
    u?: number;

    /**
     * The period of the 1st EMA of the price reference `DEMA(close, x, y)`.
     *
     * The value should be greater than 0. The default value is 32.
     */
    x?: number;

    /**
     * The period of the 2nd EMA of the price reference `DEMA(close, x, y)`.
     *
     * Setting `y=1` makes the price reference a single EMA. The value should be
     * greater than 0. The default value is 7.
     */
    y?: number;

    /**
     * A component of a bar to use when updating the indicator with a bar sample.
     *
     * If _undefined_, the bar component will have a default value (ClosePrice)
     * and will not be shown in the indicator mnemonic.
     */
    barComponent?: BarComponent;

    /**
     * A component of a quote to use when updating the indicator with a quote sample.
     *
     * If _undefined_, the quote component will have a default value and will not be shown in the indicator mnemonic.
     */
    quoteComponent?: QuoteComponent;

    /**
     * A component of a trade to use when updating the indicator with a trade sample.
     *
     * If _undefined_, the trade component will have a default value and will not be shown in the indicator mnemonic.
     */
    tradeComponent?: TradeComponent;
}

export function defaultParams(): SlopeDivergenceTsiFilterParams {
    return { q: 2, r: 32, s: 32, u: 7, x: 32, y: 7 };
}
