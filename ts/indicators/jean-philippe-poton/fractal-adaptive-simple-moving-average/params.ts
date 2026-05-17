import { BarComponent } from '../../../entities/bar-component';
import { QuoteComponent } from '../../../entities/quote-component';
import { TradeComponent } from '../../../entities/trade-component';

/** Describes parameters to create an instance of the fractal adaptive simple moving average indicator. */
export interface FractalAdaptiveSimpleMovingAverageParams {
    /**
     * Period is the lookback period N for the FDI computation.
     *
     * The value should be greater than 1.
     */
    period: number;

    /**
     * NormalSpeed is the base SMA period before fractal adaptation.
     *
     * The value should be greater than 0.
     */
    normalSpeed: number;

    /**
     * A component of a bar to use when updating the indicator with a bar sample.
     */
    barComponent?: BarComponent;

    /**
     * A component of a quote to use when updating the indicator with a quote sample.
     */
    quoteComponent?: QuoteComponent;

    /**
     * A component of a trade to use when updating the indicator with a trade sample.
     */
    tradeComponent?: TradeComponent;
}

export function defaultParams(): FractalAdaptiveSimpleMovingAverageParams {
    return {
        period: 30,
        normalSpeed: 20,
    };
}
