import { BarComponent } from '../../../entities/bar-component';
import { QuoteComponent } from '../../../entities/quote-component';
import { TradeComponent } from '../../../entities/trade-component';

/** Describes parameters to create an instance of the Jurik Fractal Adaptive Zero Lag Velocity indicator. */
export interface JurikFractalAdaptiveZeroLagVelocityParams {
    /** Minimum depth for the velocity computation. Must be >= 2. */
    loDepth: number;

    /** Maximum depth for the velocity computation. Must be >= loDepth. */
    hiDepth: number;

    /** Selects the scale set (1-4). */
    fractalType: number;

    /** Smoothing window for CFB channel averages. Must be >= 1. */
    smooth: number;

    /** A component of a bar to use when updating the indicator with a bar sample. */
    barComponent?: BarComponent;

    /** A component of a quote to use when updating the indicator with a quote sample. */
    quoteComponent?: QuoteComponent;

    /** A component of a trade to use when updating the indicator with a trade sample. */
    tradeComponent?: TradeComponent;
}

export function defaultParams(): JurikFractalAdaptiveZeroLagVelocityParams {
    return { loDepth: 5, hiDepth: 30, fractalType: 1, smooth: 10 };
}
