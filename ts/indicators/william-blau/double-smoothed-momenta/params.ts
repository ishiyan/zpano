import { BarComponent } from '../../../entities/bar-component';
import { QuoteComponent } from '../../../entities/quote-component';
import { TradeComponent } from '../../../entities/trade-component';

/**
 * Describes parameters to create an instance of the Double-Smoothed Momenta indicator.
 *
 * The parameter names `a`, `y` and `z` are the canonical symbols from William Blau's
 * double-smoothed momentum family. They are kept verbatim for fidelity with the
 * catalog definition and the test-data naming.
 */
export interface DoubleSmoothedMomentaParams {
    /**
     * The highest/lowest close look-back.
     *
     * `a=2` gives the one-bar momentum of the RSI family; `a>2` gives a
     * double-smoothed stochastic of the close. The value should be greater
     * than 0. The default value is 2.
     */
    a?: number;

    /**
     * The period of the inner (1st) smoothing EMA of the cascade.
     *
     * Setting `y=1` makes the inner stage a passthrough, so `DM(2,1,z)` is the
     * EMA-form `RSI(z)`. The value should be greater than 0. The default value is 2.
     */
    y?: number;

    /**
     * The period of the outer (2nd) smoothing EMA of the cascade.
     *
     * This is the dominant smoothing (the RSI-style default 14). The value
     * should be greater than 0. The default value is 14.
     */
    z?: number;

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

export function defaultParams(): DoubleSmoothedMomentaParams {
    return { a: 2, y: 2, z: 14 };
}
