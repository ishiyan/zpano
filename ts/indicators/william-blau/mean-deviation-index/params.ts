import { BarComponent } from '../../../entities/bar-component';
import { QuoteComponent } from '../../../entities/quote-component';
import { TradeComponent } from '../../../entities/trade-component';

/**
 * Describes parameters to create an instance of the Mean Deviation Index indicator.
 *
 * The parameter names `r`, `s`, `u` and `ul` are the canonical symbols from
 * William Blau's _Momentum, Direction, and Divergence_ (Wiley, 1995), chapter 5.
 * They are kept verbatim for fidelity with the book, the MQL5 reference, and the
 * test-data naming.
 */
export interface MeanDeviationIndexParams {
    /**
     * The period of the baseline (detrending) EMA subtracted from the price.
     *
     * The mean deviation is `md_k = price_k - EMA(price, r)_k`. Setting `r=1`
     * makes the baseline a passthrough, so the deviation is 0 on every bar and
     * the index is identically 0. The value should be greater than 0. The
     * default value is 20.
     */
    r?: number;

    /**
     * The period of the 1st smoothing EMA, applied to the mean deviation.
     *
     * The value should be greater than 0. The default value is 5.
     */
    s?: number;

    /**
     * The period of the 2nd smoothing EMA, applied to the output of the 1st
     * smoothing EMA.
     *
     * Setting `u=1` switches the 2nd stage off (passthrough), yielding the
     * book's pure double-smoothed form `EMA(price - EMA(price, r), s)`. The
     * value should be greater than 0. The default value is 3.
     */
    u?: number;

    /**
     * The period of the signal-line EMA, applied to the index to produce the
     * second output (Blau's Ergodic signal line).
     *
     * Setting `ul=1` makes the signal a passthrough (signal == mdi every bar).
     * The value should be greater than 0. This parameter is not shown in the
     * indicator mnemonic. The default value is 3.
     */
    ul?: number;

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

export function defaultParams(): MeanDeviationIndexParams {
    return { r: 20, s: 5, u: 3, ul: 3 };
}
