import { BarComponent } from '../../../entities/bar-component';
import { QuoteComponent } from '../../../entities/quote-component';
import { TradeComponent } from '../../../entities/trade-component';

/**
 * Describes parameters to create an instance of the True Strength Index indicator.
 *
 * The parameter names `q`, `r`, `s`, `u` and `ul` are the canonical symbols from
 * William Blau's _Momentum, Direction, and Divergence_ (Wiley, 1995), chapter 2.
 * They are kept verbatim for fidelity with the book, the MQL5 reference, and the
 * test-data naming.
 */
export interface TrueStrengthIndexParams {
    /**
     * The momentum look-back period; momentum is `C_k - C_(k-(q-1))`.
     *
     * The look-back distance is `q-1` bars, so `q=2` is the one-bar momentum
     * Blau uses throughout the book. The value should be greater than 0
     * (`q >= 2` is meaningful). The default value is 2.
     */
    q?: number;

    /**
     * The period of the 1st (innermost) EMA in the smoothing cascade, applied
     * to the momentum.
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
     * book's classic double-smoothed TSI. The value should be greater than 0.
     * The default value is 3.
     */
    u?: number;

    /**
     * The period of the signal-line EMA, applied to the oscillator to produce
     * the second output (Blau's Ergodic signal line).
     *
     * Setting `ul=1` makes the signal a passthrough (signal == tsi every bar).
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

export function defaultParams(): TrueStrengthIndexParams {
    return { q: 2, r: 20, s: 5, u: 3, ul: 3 };
}
