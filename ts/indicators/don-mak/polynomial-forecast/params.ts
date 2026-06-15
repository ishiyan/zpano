import { BarComponent } from '../../../entities/bar-component';
import { QuoteComponent } from '../../../entities/quote-component';
import { TradeComponent } from '../../../entities/trade-component';

/** Describes parameters to create an instance of the Polynomial Forecast indicator. */
export interface PolynomialForecastParams {
    /**
     * The polynomial degree for the local fit (uses degree+1 bars).
     *
     * The value should be >= 2. The default value is 3.
     */
    degree?: number;

    /**
     * The Taylor expansion order: 1 = velocity only (F1V), 2 = velocity + acceleration (F1VA).
     *
     * The value should be 1 or 2. The default value is 1.
     */
    order?: number;

    /**
     * The EMA pre-smoothing period applied to price before fitting (0 = none).
     *
     * The value should be >= 0. The default value is 0.
     */
    smoothing?: number;

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

export function defaultParams(): PolynomialForecastParams {
    return { degree: 3, order: 1, smoothing: 0 };
}
