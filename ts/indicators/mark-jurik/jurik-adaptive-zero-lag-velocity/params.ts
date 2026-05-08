import { BarComponent } from '../../../entities/bar-component';
import { QuoteComponent } from '../../../entities/quote-component';
import { TradeComponent } from '../../../entities/trade-component';

/** Describes parameters to create an instance of the Jurik Adaptive Zero Lag Velocity indicator. */
export interface JurikAdaptiveZeroLagVelocityParams {
    /** Minimum adaptive depth. Must be >= 2. */
    loLength: number;

    /** Maximum adaptive depth. Must be >= loLength. */
    hiLength: number;

    /** Controls the volatility regime detection sensitivity. */
    sensitivity: number;

    /** Controls the adaptive smoother period. Must be > 0. */
    period: number;

    /** A component of a bar to use when updating the indicator with a bar sample. */
    barComponent?: BarComponent;

    /** A component of a quote to use when updating the indicator with a quote sample. */
    quoteComponent?: QuoteComponent;

    /** A component of a trade to use when updating the indicator with a trade sample. */
    tradeComponent?: TradeComponent;
}

export function defaultParams(): JurikAdaptiveZeroLagVelocityParams {
    return { loLength: 5, hiLength: 30, sensitivity: 1.0, period: 3.0 };
}
