import { BarComponent } from '../../../entities/bar-component';
import { QuoteComponent } from '../../../entities/quote-component';
import { TradeComponent } from '../../../entities/trade-component';

/** Describes parameters to create an instance of the Modified Exponential Moving Average indicator. */
export interface ModifiedExponentialMovingAverageParams {
    /**
     * The EMA smoothing period.
     *
     * The value should be >= 2. The default value is 6.
     */
    period?: number;

    /**
     * The polynomial degree for the velocity correction.
     *
     * The value should be >= 2. The default value is 3.
     */
    degree?: number;

    /**
     * The stride for sampling the EMA history (1 = MEMA, >1 = MEMA-D).
     *
     * The value should be >= 1. The default value is 1.
     */
    skip?: number;

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

export function defaultParams(): ModifiedExponentialMovingAverageParams {
    return { period: 6, degree: 3, skip: 1 };
}
