/** Enumerates outputs of the Moving Average Convergence Divergence indicator. */
export enum MovingAverageConvergenceDivergenceOutput {

  /** The MACD line value (fast MA - slow MA). */
  MACDValue = 0,

  /** The signal line value (MA of MACD line). */
  SignalValue = 1,

  /** The histogram value (MACD - signal). */
  HistogramValue = 2,
}
