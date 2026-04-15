import { BarComponent } from '../../../entities/bar-component';
import { QuoteComponent } from '../../../entities/quote-component';
import { TradeComponent } from '../../../entities/trade-component';

/** Describes parameters to create an instance of the indicator based on length. */
export interface T2ExponentialMovingAverageLengthParams {
    /**
     * Length is the length (the number of time periods, ℓ) of the moving window to calculate the average.
     *
     * The value should be greater than 1.
     */
    length: number;

    /**
     * VolumeFactor is the volume factor, ν (0 ≤ ν ≤ 1), of the exponential moving average.
     * The default value is 0.7.
     * When ν=0, T2 is just an EMA, and when ν=1, T2 is DEMA.
     * In between, T2 is a cooler DEMA.
     */
    volumeFactor: number;

    /**
     * FirstIsAverage indicates whether the very first exponential moving average value is
     * a simple average of the first 'period' (the most widely documented approach) or
     * the first input value (used in Metastock).
     */
    firstIsAverage: boolean;

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

/** Describes parameters to create an instance of the indicator based on smoothing factor. */
export interface T2ExponentialMovingAverageSmoothingFactorParams {
    /**
     * SmoothingFactor is the smoothing factor, α in [0,1], of the exponential moving average.
     *
     * The equivalent length ℓ is:
     *
     *     ℓ = 2/α - 1, 0<α≤1, 1≤ℓ.
     */
    smoothingFactor: number;

    /**
     * VolumeFactor is the volume factor, ν (0 ≤ ν ≤ 1), of the exponential moving average.
     * The default value is 0.7.
     * When ν=0, T2 is just an EMA, and when ν=1, T2 is DEMA.
     * In between, T2 is a cooler DEMA.
     */
    volumeFactor: number;

    /**
     * FirstIsAverage indicates whether the very first exponential moving average value is
     * a simple average of the first 'period' (the most widely documented approach) or
     * the first input value (used in Metastock).
     */
    firstIsAverage: boolean;

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
