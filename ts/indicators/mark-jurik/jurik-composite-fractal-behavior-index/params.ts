import { BarComponent } from '../../../entities/bar-component';
import { QuoteComponent } from '../../../entities/quote-component';
import { TradeComponent } from '../../../entities/trade-component';

/** Describes parameters to create an instance of the Jurik Composite Fractal Behavior Index indicator. */
export interface JurikCompositeFractalBehaviorIndexParams {
    /**
     * FractalType controls the maximum fractal depth. Valid values are 1–4:
     *   1 = JCFB24 (8 depths: 2,3,4,6,8,12,16,24)
     *   2 = JCFB48 (10 depths: +32,48)
     *   3 = JCFB96 (12 depths: +64,96)
     *   4 = JCFB192 (14 depths: +128,192)
     */
    fractalType: number;

    /**
     * Smooth is the smoothing window for the running averages.
     * The value should be >= 1.
     */
    smooth: number;

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

export function defaultParams(): JurikCompositeFractalBehaviorIndexParams {
    return { fractalType: 1, smooth: 10 };
}
