import { BarComponent } from '../../../entities/bar-component';
import { QuoteComponent } from '../../../entities/quote-component';
import { TradeComponent } from '../../../entities/trade-component';

/**
 * Describes parameters to create an instance of the MACD Index indicator.
 *
 * The parameter names `r`, `s`, `u` and `ul` are the canonical symbols from
 * William Blau's _Momentum, Direction, and Divergence_ (Wiley, 1995), chapter 5.
 * They are kept verbatim for fidelity with the book, the MQL5 reference, and the
 * test-data naming.
 */
export interface MacdIndexParams {
    /**
     * The period of the slow EMA in the MACD line (`EMA(close, s) - EMA(close, r)`).
     *
     * The value should be greater than 0 and strictly greater than `s`. The
     * default value is 20.
     */
    r?: number;

    /**
     * The period of the fast EMA in the MACD line.
     *
     * The value should be greater than 0 and strictly less than `r` -- the MACD
     * line is fast minus slow, so the fast period must be shorter. The default
     * value is 5.
     */
    s?: number;

    /**
     * The period of the smoothing EMA applied to the MACD line.
     *
     * Setting `u=1` switches this stage off (passthrough), yielding the book's
     * pure two-EMA MACD line `EMA(close, s) - EMA(close, r)`. The value should
     * be greater than 0. The default value is 3.
     */
    u?: number;

    /**
     * The period of the signal-line EMA, applied to the index to produce the
     * second output (Blau's Ergodic signal line).
     *
     * Setting `ul=1` makes the signal a passthrough (signal == macdi every
     * bar). The value should be greater than 0. This parameter is not shown in
     * the indicator mnemonic. The default value is 3.
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

export function defaultParams(): MacdIndexParams {
    return { r: 20, s: 5, u: 3, ul: 3 };
}
