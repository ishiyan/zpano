import { BarComponent } from '../../../entities/bar-component';
import { QuoteComponent } from '../../../entities/quote-component';
import { TradeComponent } from '../../../entities/trade-component';

/** Describes parameters to create an instance of the Polynomial Fit Derivative indicator. */
export interface PolynomialFitDerivativeParams {
    /**
     * The polynomial degree. The number of data points used is degree + 1.
     *
     * The value should be >= 2. The default value is 3 (cubic).
     */
    degree?: number;

    /**
     * The derivative order (1 = velocity, 2 = acceleration).
     *
     * The value should be >= 1 and <= degree. The default value is 1.
     */
    order?: number;

    /**
     * The EMA pre-smoothing length applied before the FIR filter.
     *
     * The value should be >= 0 (0 means no smoothing). The default value is 6.
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

export function defaultParams(): PolynomialFitDerivativeParams {
    return { degree: 3, order: 1, smoothing: 6 };
}
