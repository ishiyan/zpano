import { BarComponent } from '../../../entities/bar-component';
import { QuoteComponent } from '../../../entities/quote-component';
import { TradeComponent } from '../../../entities/trade-component';

/** Describes parameters to create an instance of the Zero-lag Exponential Moving Average indicator. */
export interface ZeroLagExponentialMovingAverageParams {
    /**
     * The smoothing factor (alpha) of the EMA.
     *
     * alpha = 2/(length + 1), 0 < alpha <= 1, 1 <= length.
     * The default value is 0.25.
     */
    smoothingFactor: number;

    /**
     * The gain factor used to estimate the velocity.
     *
     * The default value is 0.5.
     */
    velocityGainFactor: number;

    /**
     * The length of the momentum used to estimate the velocity.
     *
     * The value should be positive. The default value is 3.
     */
    velocityMomentumLength: number;

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
