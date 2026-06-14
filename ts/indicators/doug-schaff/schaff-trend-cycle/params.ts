import { BarComponent } from '../../../entities/bar-component';
import { QuoteComponent } from '../../../entities/quote-component';
import { TradeComponent } from '../../../entities/trade-component';

/** Describes parameters to create an instance of the Schaff Trend Cycle indicator. */
export interface SchaffTrendCycleParams {
    /**
     * The number of periods for the fast EMA of the MACD line.
     *
     * The value should be greater than 0. The default value is 23.
     */
    fast?: number;

    /**
     * The number of periods for the slow EMA of the MACD line.
     * It also sets the warm-up gate (barindex > slow).
     *
     * The value should be greater than 0. The default value is 50.
     */
    slow?: number;

    /**
     * The cycle length — the look-back for both stochastics.
     *
     * The value should be greater than 0. The default value is 10.
     */
    tclen?: number;

    /**
     * The EMA smoothing alpha for both %D stages.
     *
     * The value should be in (0, 1]. The default value is 0.5.
     */
    factor?: number;

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

export function defaultParams(): SchaffTrendCycleParams {
    return { fast: 23, slow: 50, tclen: 10, factor: 0.5 };
}
