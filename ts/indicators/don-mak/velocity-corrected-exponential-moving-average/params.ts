import { BarComponent } from '../../../entities/bar-component';
import { QuoteComponent } from '../../../entities/quote-component';
import { TradeComponent } from '../../../entities/trade-component';

/** Describes parameters to create an instance of the Velocity-Corrected Exponential Moving Average indicator. */
export interface VelocityCorrectedExponentialMovingAverageParams {
    /**
     * The EMA smoothing period.
     *
     * The value should be >= 2. The default value is 6.
     */
    period?: number;

    /**
     * The polynomial degree for the velocity estimation.
     *
     * The value should be >= 2. The default value is 3.
     */
    degree?: number;

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

export function defaultParams(): VelocityCorrectedExponentialMovingAverageParams {
    return { period: 6, degree: 3 };
}
