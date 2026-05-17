import { BarComponent } from '../../../entities/bar-component';
import { QuoteComponent } from '../../../entities/quote-component';
import { TradeComponent } from '../../../entities/trade-component';

/** Describes parameters to create an instance of the fractal bands indicator. */
export interface FractalBandsParams {
    /** Period is the lookback period for the FGDI computation. The value should be greater than 1. */
    period: number;

    /** NormalSpeed is the base SMA period before fractal adaptation. The value should be greater than 0. */
    normalSpeed: number;

    /** Alpha is the band width multiplier raised to power H. The value should be greater than 0. */
    alpha: number;

    /** A component of a bar to use when updating the indicator with a bar sample. */
    barComponent?: BarComponent;

    /** A component of a quote to use when updating the indicator with a quote sample. */
    quoteComponent?: QuoteComponent;

    /** A component of a trade to use when updating the indicator with a trade sample. */
    tradeComponent?: TradeComponent;
}

export function defaultParams(): FractalBandsParams {
    return {
        period: 30,
        normalSpeed: 20,
        alpha: 2.0,
    };
}
