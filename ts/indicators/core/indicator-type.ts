/** Enumerates indicator types. */
export enum IndicatorType {

  /** Identifies the __Simple Moving Average__ (SMA) indicator. */
  SimpleMovingAverage,

  /** Identifies the __Weighted Moving Average__ (WMA) indicator. */
  WeightedMovingAverage,

  /** Identifies the __Triangular Moving Average__ (TRIMA) indicator. */
  TriangularMovingAverage,

  /** Identifies the __Exponential Moving Average__ (EMA) indicator. */
  ExponentialMovingAverage,

  /** Identifies the __Double Exponential Moving Average__ (DEMA) indicator. */
  DoubleExponentialMovingAverage,

  /** Identifies the __Triple Exponential Moving Average__ (TEMA) indicator. */
  TripleExponentialMovingAverage,

  /** Identifies the __T2 Exponential Moving Average__ (T2EMA) indicator. */
  T2ExponentialMovingAverage,

  /** Identifies the __T3 Exponential Moving Average__ (T3EMA) indicator. */
  T3ExponentialMovingAverage,

  /** Identifies the __Kaufman Adaptive Moving Average__ (KAMA) indicator. */
  KaufmanAdaptiveMovingAverage,

  /** Identifies the __Jurik Moving Average__ (JMA) indicator. */
  JurikMovingAverage,

  /** Identifies the Ehler's __MESA Adaptive Moving Average__ (MAMA) indicator. */
  MesaAdaptiveMovingAverage,

  /** Identifies the Ehler's __Fractal Adaptive Moving Average__ (FRAMA) indicator. */
  FractalAdaptiveMovingAverage,

  /** Identifies the __Momentum__ (MOM) indicator. */
  Momentum,

  /** Identifies the __Rate of Change__ (ROC) indicator. */
  RateOfChange,

  /** Identifies the __Rate of Change Percent__ (ROCP) indicator. */
  RateOfChangePercent,

  /** Identifies the __Relative Strength Index__ (RSI) indicator. */
  RelativeStrengthIndex,

  /** Identifies the __Chande Momentum Oscillator__ (CMO) indicator. */
  ChandeMomentumOscillator,

  /** Identifies the __Bollinger Bands__ (BB) indicator. */
  BollingerBands,

  /** Identifies the __Variance__ (VAR) indicator. */
  Variance,

  /** Identifies the __Standard Deviation__ (STDEV) indicator. */
  StandardDeviation,

  /** Identifies the __Goertzel power spectrum__ (GOERTZEL) indicator. */
  GoertzelSpectrum,

  /** Identifies the Ehler's __Center of Gravity Oscillator__ (COG) indicator. */
  CenterOfGravityOscillator,

  /** Identifies the Ehler's __Cyber Cycle__ (CC) indicator. */
  CyberCycle,

  /** Identifies the Ehler's __Instantaneous Trend Line__ (iTrend) indicator. */
  InstantaneousTrendLine,

  /** Identifies the Ehler's __Super Smoother__ (SS) indicator. */
  SuperSmoother,

  /** Identifies the Ehler's __Zero-lag Exponential Moving Average__ (ZEMA) indicator. */
  ZeroLagExponentialMovingAverage,

  /** Identifies the Ehler's __Zero-lag Error-Correcting Exponential Moving Average__ (ZECEMA) indicator. */
  ZeroLagErrorCorrectingExponentialMovingAverage,

  /** Identifies the Ehler's __Roofing Filter__ indicator. */
  RoofingFilter,

  /** Identifies Welles Wilder's __True Range__ (TR) indicator. */
  TrueRange,

  /** Identifies Welles Wilder's __Average True Range__ (ATR) indicator. */
  AverageTrueRange,

  /** Identifies Welles Wilder's __Normalized Average True Range__ (NATR) indicator. */
  NormalizedAverageTrueRange,

  /** Identifies Welles Wilder's __Directional Movement Minus__ (-DM) indicator. */
  DirectionalMovementMinus,

  /** Identifies Welles Wilder's __Directional Movement Plus__ (+DM) indicator. */
  DirectionalMovementPlus,

  /** Identifies Welles Wilder's __Directional Indicator Minus__ (-DI) indicator. */
  DirectionalIndicatorMinus,

  /** Identifies Welles Wilder's __Directional Indicator Plus__ (+DI) indicator. */
  DirectionalIndicatorPlus,

  /** Identifies Welles Wilder's __Directional Movement Index__ (DX) indicator. */
  DirectionalMovementIndex,

  /** Identifies Welles Wilder's __Average Directional Movement Index__ (ADX) indicator. */
  AverageDirectionalMovementIndex,

  /** Identifies Welles Wilder's __Average Directional Movement Index Rating__ (ADXR) indicator. */
  AverageDirectionalMovementIndexRating,

  /** Identifies Larry Williams' __Williams %R__ (WILL%R) indicator. */
  WilliamsPercentR,

  /** Identifies Gerald Appel's __Percentage Price Oscillator__ (PPO) indicator. */
  PercentagePriceOscillator,

  /** Identifies the __Absolute Price Oscillator__ (APO) indicator. */
  AbsolutePriceOscillator,

  /** Identifies Donald Lambert's __Commodity Channel Index__ (CCI) indicator. */
  CommodityChannelIndex,

  /** Identifies Gene Quong's __Money Flow Index__ (MFI) indicator. */
  MoneyFlowIndex,

  /** Identifies Joseph Granville's __On-Balance Volume__ (OBV) indicator. */
  OnBalanceVolume,

  /** Identifies Igor Livshin's __Balance of Power__ (BOP) indicator. */
  BalanceOfPower
}
