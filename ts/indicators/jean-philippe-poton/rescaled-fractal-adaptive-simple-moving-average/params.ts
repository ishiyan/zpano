import { BarComponent } from '../../../entities/bar-component';
import { QuoteComponent } from '../../../entities/quote-component';
import { TradeComponent } from '../../../entities/trade-component';

/** Describes parameters to create an instance of the rescaled fractal adaptive simple moving average indicator. */
export interface RescaledFractalAdaptiveSimpleMovingAverageParams {
    /**
     * Period is the lookback window for R/S analysis. Must be a power of 2, >= 4.
     */
    period: number;

    /**
     * NormalSpeed is the base SMA period before fractal adaptation. Must be >= 1.
     */
    normalSpeed: number;

    /**
     * PriceScale is the multiplier applied to prices before R/S calculation. Default is 1.0.
     */
    priceScale: number;

    /** A component of a bar to use when updating the indicator with a bar sample. */
    barComponent?: BarComponent;

    /** A component of a quote to use when updating the indicator with a quote sample. */
    quoteComponent?: QuoteComponent;

    /** A component of a trade to use when updating the indicator with a trade sample. */
    tradeComponent?: TradeComponent;
}

export function defaultParams(): RescaledFractalAdaptiveSimpleMovingAverageParams {
    return {
        period: 64,
        normalSpeed: 30,
        priceScale: 1.0,
    };
}
