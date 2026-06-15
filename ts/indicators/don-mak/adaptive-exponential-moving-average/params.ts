import { BarComponent } from '../../../entities/bar-component';
import { QuoteComponent } from '../../../entities/quote-component';
import { TradeComponent } from '../../../entities/trade-component';

/** Describes parameters to create an instance of the Adaptive Exponential Moving Average indicator. */
export interface AdaptiveExponentialMovingAverageParams {
    /**
     * The smoothing factor for trending data (low frequency).
     *
     * The value should be in (0, 1] and greater than alphaMin. The default value is 0.5.
     */
    alphaMax?: number;

    /**
     * The smoothing factor for noisy data (high frequency).
     *
     * The value should be in (0, alphaMax). The default value is 0.05.
     */
    alphaMin?: number;

    /**
     * The crossover frequency in radians/bar. Below this, alpha = alphaMax.
     *
     * The value should be in (0, pi). The default value is 1.0.
     */
    omega0?: number;

    /**
     * The embedded ISWP internal smoothing parameter.
     *
     * The value should be >= 0. The default value is 3.
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

export function defaultParams(): AdaptiveExponentialMovingAverageParams {
    return { alphaMax: 0.5, alphaMin: 0.05, omega0: 1.0, smoothing: 3 };
}
