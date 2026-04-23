import { BarComponent } from '../../../entities/bar-component';
import { QuoteComponent } from '../../../entities/quote-component';
import { TradeComponent } from '../../../entities/trade-component';

/** The default inverse scaling factor (0.015). */
export const DefaultInverseScalingFactor = 0.015;

/** Describes parameters to create an instance of the Commodity Channel Index indicator. */
export interface CommodityChannelIndexParams {
    /**
     * The number of time periods of the commodity channel index.
     *
     * The value should be greater than 1.
     */
    length: number;

    /**
     * The inverse scaling factor to provide more readable value numbers.
     * The default value of 0.015 ensures that approximately 70 to 80 percent
     * of CCI values would fall between -100 and +100.
     *
     * If _undefined_, the default (0.015) is used.
     */
    inverseScalingFactor?: number;

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
