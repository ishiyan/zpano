import { BarComponent } from '../../../entities/bar-component';
import { QuoteComponent } from '../../../entities/quote-component';
import { TradeComponent } from '../../../entities/trade-component';

/** Describes parameters to create an instance of the fractal bands hybride adaptive indicator. */
export interface FractalBandsHybrideAdaptiveParams {
    /** Period is the lookback period for the FGDI computation. The value should be greater than 1. */
    period: number;

    /** NormalSpeedFallback is the fallback SMA period when CyclePeriod is unavailable. The value should be greater than 0. */
    normalSpeedFallback: number;

    /** Alpha is the band width multiplier raised to power H. The value should be greater than 0. */
    alpha: number;

    /** Nyquist multiplier applied to the estimated cycle period. The value should be greater than 0. */
    nyquist: number;

    /** High-pass filter alpha for Ehlers CyclePeriod. The value should be between 0 and 1. */
    alphaHP: number;

    /** A component of a bar to use when updating the indicator with a bar sample. */
    barComponent?: BarComponent;

    /** A component of a quote to use when updating the indicator with a quote sample. */
    quoteComponent?: QuoteComponent;

    /** A component of a trade to use when updating the indicator with a trade sample. */
    tradeComponent?: TradeComponent;
}

export function defaultParams(): FractalBandsHybrideAdaptiveParams {
    return {
        period: 30,
        normalSpeedFallback: 30,
        alpha: 2.0,
        nyquist: 0.5,
        alphaHP: 0.07,
    };
}
