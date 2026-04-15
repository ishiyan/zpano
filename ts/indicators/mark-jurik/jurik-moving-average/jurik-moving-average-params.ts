import { BarComponent } from '../../../entities/bar-component';
import { QuoteComponent } from '../../../entities/quote-component';
import { TradeComponent } from '../../../entities/trade-component';

/** Describes parameters to create an instance of the indicator. */
export interface JurikMovingAverageParams {
    /**
     * Length (the number of time periods, ℓ) determines
     * the degree of smoothness and it can be any positive value.
     *
     * Small values make the moving average respond rapidly to price change
     * and larger values produce smoother, flatter curves.
     *
	 * The value should be greater than _1_. Typical values range from _5_ to _20_.
	 *
     * Irrespective from the value, the indicator needs at _30_ first values to be primed.
     */
    length: number;

    /**
     * Phase affects the amount of lag (delay).
     *
     * Lower lag tends to produce larger overshoot during price gaps, so you need
     * to consider the trade-off between lag and overshoot and select a value for
     * phase that balances your trading system's needs.
     *
     * Small values make the moving average respond rapidly to price change
     * and larger values produce smoother, flatter curves.
     * 
	 * The phase values should be in _[-100, 100]_.
	 *
	 * - The value of _-100_ results in maximum lag and no overshoot.
     * - The value of _0_ results in some lag and some overshoot.
     * - The value of _100_ results in minimum lag and maximum overshoot.
     */
    phase: number;

    /**
     * A component of a bar to use when updating the indicator with a bar sample.
     *
     * If _undefined_, the bar component will have a default value and will not be shown in the indicator mnemonic.
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
