/** Specifies the type of moving average to use in the ADOSC calculation. */
export enum MovingAverageType {
  /** Simple Moving Average. */
  SMA = 0,

  /** Exponential Moving Average. */
  EMA = 1,
}

/** Describes parameters to create an instance of the Advance-Decline Oscillator indicator. */
export interface AdvanceDeclineOscillatorParams {
    /**
     * The number of periods for the fast moving average.
     *
     * The value should be greater than 1.
     */
    fastLength: number;

    /**
     * The number of periods for the slow moving average.
     *
     * The value should be greater than 1.
     */
    slowLength: number;

    /**
     * The type of moving average (SMA or EMA).
     *
     * If _undefined_, the Exponential Moving Average is used.
     */
    movingAverageType?: MovingAverageType;

    /**
     * Controls the EMA seeding algorithm.
     * When true, the first EMA value is the simple average of the first period values.
     * When false (default), the first input value is used directly (Metastock style).
     * Only relevant when movingAverageType is EMA.
     */
    firstIsAverage?: boolean;
}
