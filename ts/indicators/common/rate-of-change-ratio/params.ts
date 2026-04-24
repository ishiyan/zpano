import { BarComponent } from '../../../entities/bar-component';
import { QuoteComponent } from '../../../entities/quote-component';
import { TradeComponent } from '../../../entities/trade-component';

/** Describes parameters to create an instance of the Rate of Change Ratio indicator. */
export interface RateOfChangeRatioParams {
    /**
     * Length is the length (the number of time periods, ℓ) between today's sample and the sample ℓ periods ago.
     *
     * The value should be greater than 0.
     */
    length: number;

    /**
     * Indicates whether to multiply the ratio by 100.
     *
     * If false (default), the result is price/previousPrice (centered at 1).
     * If true, the result is (price/previousPrice)*100 (centered at 100).
     */
    hundredScale?: boolean;

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

export function defaultParams(): RateOfChangeRatioParams {
    return {
        length: 10,
    };
}
