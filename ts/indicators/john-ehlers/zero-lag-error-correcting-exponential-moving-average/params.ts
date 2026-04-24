import { BarComponent } from '../../../entities/bar-component';
import { QuoteComponent } from '../../../entities/quote-component';
import { TradeComponent } from '../../../entities/trade-component';

/** Describes parameters to create an instance of the Zero-lag Error-Correcting Exponential Moving Average indicator. */
export interface ZeroLagErrorCorrectingExponentialMovingAverageParams {
    /**
     * The smoothing factor (alpha) of the EMA.
     *
     * alpha = 2/(length + 1), 0 < alpha <= 1, 1 <= length.
     * The default value is 0.095 (equivalent to length 20).
     */
    smoothingFactor: number;

    /**
     * Defines the range [-g, g] for finding the best gain factor.
     *
     * The value should be positive. The default value is 5.
     */
    gainLimit: number;

    /**
     * Defines the iteration step for finding the best gain factor.
     *
     * The value should be positive. The default value is 0.1.
     */
    gainStep: number;

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

export function defaultParams(): ZeroLagErrorCorrectingExponentialMovingAverageParams {
    return { smoothingFactor: 0.095, gainLimit: 5, gainStep: 0.1 };
}
