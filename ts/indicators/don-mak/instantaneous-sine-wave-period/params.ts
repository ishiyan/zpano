import { BarComponent } from '../../../entities/bar-component';
import { QuoteComponent } from '../../../entities/quote-component';
import { TradeComponent } from '../../../entities/trade-component';

/** Describes parameters to create an instance of the Instantaneous Sine Wave Period indicator. */
export interface InstantaneousSineWavePeriodParams {
    /**
     * The EMA smoothing length applied to input prices before frequency estimation.
     *
     * The value should be >= 0 (0 means no smoothing). The default value is 0.
     */
    smoothing?: number;

    /**
     * The minimum allowed period in bars. Estimates below this are rejected.
     *
     * The value should be > 0. The default value is 4.0.
     */
    minPeriod?: number;

    /**
     * The maximum allowed period in bars. Estimates above this are rejected.
     *
     * The value should be > minPeriod. The default value is 50.0.
     */
    maxPeriod?: number;

    /**
     * The maximum tolerated error for the omega estimate. If both methods exceed this, the output is NaN.
     *
     * The value should be > 0. The default value is 20.0.
     */
    errorThreshold?: number;

    /**
     * The assumed measurement error for each price point (used in error propagation).
     *
     * The value should be > 0. The default value is 0.01.
     */
    dx?: number;

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

export function defaultParams(): InstantaneousSineWavePeriodParams {
    return { smoothing: 0, minPeriod: 4.0, maxPeriod: 50.0, errorThreshold: 20.0, dx: 0.01 };
}
