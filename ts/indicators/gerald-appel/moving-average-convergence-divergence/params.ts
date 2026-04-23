import { BarComponent } from '../../../entities/bar-component';
import { QuoteComponent } from '../../../entities/quote-component';
import { TradeComponent } from '../../../entities/trade-component';

/** Specifies the type of moving average to use in the MACD calculation. */
export enum MovingAverageType {
  /** Exponential Moving Average (default for classic MACD). */
  EMA = 0,

  /** Simple Moving Average. */
  SMA = 1,
}

/** Describes parameters to create an instance of the Moving Average Convergence Divergence indicator. */
export interface MovingAverageConvergenceDivergenceParams {
    /**
     * The number of periods for the fast moving average.
     *
     * The value should be greater than 1. The default value is 12.
     */
    fastLength?: number;

    /**
     * The number of periods for the slow moving average.
     *
     * The value should be greater than 1. The default value is 26.
     */
    slowLength?: number;

    /**
     * The number of periods for the signal line moving average.
     *
     * The value should be greater than 0. The default value is 9.
     */
    signalLength?: number;

    /**
     * The type of moving average for the fast and slow lines (EMA or SMA).
     *
     * If _undefined_, the Exponential Moving Average is used.
     */
    movingAverageType?: MovingAverageType;

    /**
     * The type of moving average for the signal line (EMA or SMA).
     *
     * If _undefined_, the Exponential Moving Average is used.
     */
    signalMovingAverageType?: MovingAverageType;

    /**
     * Controls the EMA seeding algorithm.
     * When true (default), the first EMA value is the simple average of the first period values
     * (TA-Lib compatible). When false, the first input value is used directly (Metastock style).
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
