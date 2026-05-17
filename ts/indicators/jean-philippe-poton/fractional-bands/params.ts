import { BarComponent } from '../../../entities/bar-component';
import { QuoteComponent } from '../../../entities/quote-component';
import { TradeComponent } from '../../../entities/trade-component';

/** Describes parameters to create an instance of the fractional bands indicator. */
export interface FractionalBandsParams {
    /** Period is the lookback period for the FGDI computation. The value should be greater than 1. */
    period: number;

    /** PriceScale is the multiplier converting price to a working numeric space. The value should be greater than 0. */
    priceScale: number;

    /** A component of a bar to use when updating the indicator with a bar sample. */
    barComponent?: BarComponent;

    /** A component of a quote to use when updating the indicator with a quote sample. */
    quoteComponent?: QuoteComponent;

    /** A component of a trade to use when updating the indicator with a trade sample. */
    tradeComponent?: TradeComponent;
}

export function defaultParams(): FractionalBandsParams {
    return {
        period: 30,
        priceScale: 1.0,
    };
}
