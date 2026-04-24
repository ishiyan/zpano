import { BarComponent } from '../../../entities/bar-component';
import { QuoteComponent } from '../../../entities/quote-component';
import { TradeComponent } from '../../../entities/trade-component';

/** Specifies the type of moving average to use in the Bollinger Bands calculation. */
export enum MovingAverageType {
  /** Simple Moving Average. */
  SMA = 0,

  /** Exponential Moving Average. */
  EMA = 1,
}

/** Describes parameters to create an instance of the Bollinger Bands indicator. */
export interface BollingerBandsParams {
    /**
     * The number of periods for the moving average and standard deviation.
     *
     * The value should be greater than 1. The default value is 5.
     */
    length: number;

    /**
     * The number of standard deviations above the middle band.
     *
     * The default value is 2.0.
     */
    upperMultiplier?: number;

    /**
     * The number of standard deviations below the middle band.
     *
     * The default value is 2.0.
     */
    lowerMultiplier?: number;

    /**
     * Indicates whether to use the unbiased sample standard deviation (true)
     * or the population standard deviation (false).
     *
     * If _undefined_, defaults to true (unbiased sample standard deviation).
     */
    isUnbiased?: boolean;

    /**
     * The type of moving average (SMA or EMA).
     *
     * If _undefined_, the Simple Moving Average is used.
     */
    movingAverageType?: MovingAverageType;

    /**
     * Controls the EMA seeding algorithm.
     * When true, the first EMA value is the simple average of the first period values.
     * When false (default), the first input value is used directly (Metastock style).
     * Only relevant when movingAverageType is EMA.
     */
    firstIsAverage?: boolean;

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

export function defaultParams(): BollingerBandsParams {
    return { length: 5, upperMultiplier: 2.0, lowerMultiplier: 2.0 };
}
