import { BarComponent } from '../../../entities/bar-component';
import { QuoteComponent } from '../../../entities/quote-component';
import { TradeComponent } from '../../../entities/trade-component';

/** Describes parameters to create an instance of the Moving Mini-Max indicator. */
export interface MovingMiniMaxParams {
    /**
     * The smoothing window width controlling the quantum tunnelling ability.
     *
     * Larger values produce smoother output, suppressing smaller peaks. The value should be
     * >= 1. The default value is 5.
     */
    m?: number;

    /**
     * The lookback window size: the number of price bars over which the indicator is computed.
     *
     * Priming requires n prices. The value should be > 2*m. The default value is 50.
     */
    n?: number;

    /**
     * The number of distinct support/resistance levels to detect and return.
     *
     * The value should be >= 1. The default value is 3.
     */
    numExtrema?: number;

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

export function defaultParams(): MovingMiniMaxParams {
    return { m: 5, n: 50, numExtrema: 3 };
}
