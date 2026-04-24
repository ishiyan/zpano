import { BarComponent } from '../../../entities/bar-component';
import { QuoteComponent } from '../../../entities/quote-component';
import { TradeComponent } from '../../../entities/trade-component';

/** Describes parameters to create an instance of the indicator. */
export interface StandardDeviationParams {
    /**
     * Length is the length (the number of time periods, ℓ) of the moving window to calculate the standard deviation.
     *
     * The value should be greater than 1.
     */
    length: number;

    /**
     * Unbiased indicates whether the estimate of the standard deviation is the unbiased sample standard deviation
     * or the population standard deviation.
     *
     * When in doubt, use the unbiased sample standard deviation (value is true).
     */
    unbiased: boolean;

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

export function defaultParams(): StandardDeviationParams {
    return {
        length: 20,
        unbiased: true,
    };
}
