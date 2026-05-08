import { BarComponent } from '../../../entities/bar-component';
import { QuoteComponent } from '../../../entities/quote-component';
import { TradeComponent } from '../../../entities/trade-component';

/** Describes parameters to create an instance of the Jurik Adaptive Relative Trend Strength Index indicator. */
export interface JurikAdaptiveRelativeTrendStrengthIndexParams {
    /**
     * LoLength is the minimum adaptive RSX length.
     * The value should be at least 2.
     */
    loLength: number;

    /**
     * HiLength is the maximum adaptive RSX length.
     * The value should be at least loLength.
     */
    hiLength: number;

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

export function defaultParams(): JurikAdaptiveRelativeTrendStrengthIndexParams {
    return { loLength: 5, hiLength: 30 };
}
