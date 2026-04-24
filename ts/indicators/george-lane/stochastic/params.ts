/** Specifies the type of moving average to use for smoothing. */
export enum MovingAverageType {
  /** Simple Moving Average. */
  SMA = 0,

  /** Exponential Moving Average. */
  EMA = 1,
}

/** Describes parameters to create an instance of the Stochastic Oscillator indicator. */
export interface StochasticParams {
    /**
     * The lookback period for the raw %K calculation (highest high / lowest low).
     *
     * The value should be greater than 0.
     */
    fastKLength: number;

    /**
     * The smoothing period for Slow-K (also known as Fast-D).
     *
     * The value should be greater than 0.
     */
    slowKLength: number;

    /**
     * The smoothing period for Slow-D.
     *
     * The value should be greater than 0.
     */
    slowDLength: number;

    /**
     * The type of moving average for Slow-K smoothing (SMA or EMA).
     *
     * If _undefined_, the Simple Moving Average is used.
     */
    slowKMAType?: MovingAverageType;

    /**
     * The type of moving average for Slow-D smoothing (SMA or EMA).
     *
     * If _undefined_, the Simple Moving Average is used.
     */
    slowDMAType?: MovingAverageType;

    /**
     * Controls the EMA seeding algorithm.
     * When true, the first EMA value is the simple average of the first period values.
     * When false (default), the first input value is used directly (Metastock style).
     * Only relevant when an MA type is EMA.
     */
    firstIsAverage?: boolean;
}

export function defaultParams(): StochasticParams {
    return { fastKLength: 5, slowKLength: 3, slowDLength: 3 };
}
