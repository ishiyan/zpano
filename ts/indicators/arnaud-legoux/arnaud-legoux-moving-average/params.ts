import { BarComponent } from '../../../entities/bar-component';
import { QuoteComponent } from '../../../entities/quote-component';
import { TradeComponent } from '../../../entities/trade-component';

/** Describes parameters to create an instance of the ALMA indicator. */
export interface ArnaudLegouxMovingAverageParams {
    /**
     * Window is the number of bars in the lookback window.
     *
     * The value should be greater than 0.
     */
    window: number;

    /**
     * Sigma controls the Gaussian width; larger values produce smoother output.
     *
     * The value should be greater than 0.
     */
    sigma: number;

    /**
     * Offset shifts the Gaussian peak; 0 = centered, 1 = newest bar.
     *
     * The value should be between 0 and 1.
     */
    offset: number;

    /**
     * A component of a bar to use when updating the indicator with a bar sample.
     *
     * If _undefined_, the bar component will have a default value and will not be shown
     * in the indicator mnemonic.
     */
    barComponent?: BarComponent;

    /**
     * A component of a quote to use when updating the indicator with a quote sample.
     *
     * If _undefined_, the quote component will have a default value and will not be shown
     * in the indicator mnemonic.
     */
    quoteComponent?: QuoteComponent;

    /**
     * A component of a trade to use when updating the indicator with a trade sample.
     *
     * If _undefined_, the trade component will have a default value and will not be shown
     * in the indicator mnemonic.
     */
    tradeComponent?: TradeComponent;
}

export function defaultParams(): ArnaudLegouxMovingAverageParams {
    return {
        window: 9,
        sigma: 6.0,
        offset: 0.85,
    };
}
