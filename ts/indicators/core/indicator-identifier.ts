/** Enumerates indicators. */
export enum IndicatorIdentifier {

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

  /** Identifies the Ehler's __Dominant Cycle__ (DC) indicator, exposing raw period, smoothed period and phase. */
  DominantCycle,

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
  BalanceOfPower,

  /** Identifies the __Rate of Change Ratio__ (ROCR / ROCR100) indicator. */
  RateOfChangeRatio,

  /** Identifies __Pearson's Correlation Coefficient__ (CORREL) indicator. */
  PearsonsCorrelationCoefficient,

  /** Identifies the __Linear Regression__ (LINREG) indicator. */
  LinearRegression,

  /** Identifies Larry Williams' __Ultimate Oscillator__ (ULTOSC) indicator. */
  UltimateOscillator,

  /** Identifies Tushar Chande's __Stochastic Relative Strength Index__ (STOCHRSI) indicator. */
  StochasticRelativeStrengthIndex,

  /** Identifies George Lane's __Stochastic Oscillator__ (STOCH) indicator. */
  Stochastic,

  /** Identifies Tushar Chande's __Aroon__ (AROON) indicator. */
  Aroon,

  /** Identifies Marc Chaikin's __Advance-Decline__ (A/D) indicator. */
  AdvanceDecline,

  /** Identifies Marc Chaikin's __Advance-Decline Oscillator__ (ADOSC) indicator. */
  AdvanceDeclineOscillator,

  /** Identifies Welles Wilder's __Parabolic Stop And Reverse__ (SAR) indicator. */
  ParabolicStopAndReverse,

  /** Identifies Jack Hutson's __Triple Exponential Moving Average Oscillator__ (TRIX) indicator. */
  TripleExponentialMovingAverageOscillator,

  /** Identifies John Bollinger's __Bollinger Bands Trend__ (BBTrend) indicator. */
  BollingerBandsTrend,

  /** Identifies Gerald Appel's __Moving Average Convergence Divergence__ (MACD) indicator. */
  MovingAverageConvergenceDivergence,

  /** Identifies the Ehlers __Sine Wave__ (SW) indicator, exposing sine value, lead sine, band, dominant cycle period and phase. */
  SineWave,

  /** Identifies the Ehlers __Hilbert Transformer Instantaneous Trend Line__ (HTITL) indicator, exposing trend value and dominant cycle period. */
  HilbertTransformerInstantaneousTrendLine,

  /** Identifies the Ehlers __Trend / Cycle Mode__ (TCM) indicator, exposing the trend/cycle value (+1 in trend, −1 in cycle),
   * trend/cycle mode flags, instantaneous trend line, sine wave, lead sine wave, dominant cycle period and phase. */
  TrendCycleMode,

  /** Identifies the Ehlers __Corona Spectrum__ (CSPECT) heatmap indicator, exposing the dB heatmap column,
   * the weighted dominant cycle estimate and its 5-sample median. */
  CoronaSpectrum,

  /** Identifies the Ehlers __Corona Signal To Noise Ratio__ (CSNR) heatmap indicator, exposing the intensity
   * raster heatmap column and the current SNR mapped into the parameter range. */
  CoronaSignalToNoiseRatio,

  /** Identifies the Ehlers __Corona Swing Position__ (CSWING) heatmap indicator, exposing the intensity raster
   * heatmap column and the current swing position mapped into the parameter range. */
  CoronaSwingPosition,

  /** Identifies the Ehlers __Corona Trend Vigor__ (CTV) heatmap indicator, exposing the intensity raster heatmap
   * column and the current trend vigor scaled into the parameter range. */
  CoronaTrendVigor,

  /** Identifies Vladimir Kravchuk's __Adaptive Trend and Cycle Filter__ (ATCF) suite: a bank of five FIR filters
   * (FATL, SATL, RFTL, RSTL, RBCI) plus three composites (FTLM, STLM, PCCI) applied to a single input series. */
  AdaptiveTrendAndCycleFilter,

  /** Identifies the __Maximum Entropy Spectrum__ (MESPECT) heatmap indicator, a Burg maximum-entropy
   * auto-regressive power spectrum over a configurable cycle-period range. */
  MaximumEntropySpectrum,

  /** Identifies the Ehlers __Discrete Fourier Transform Spectrum__ (DFTPS) heatmap indicator, a
   * mean-subtracted DFT power spectrum over a configurable cycle-period range. */
  DiscreteFourierTransformSpectrum,

  /** Identifies the Ehlers __Comb Band-Pass Spectrum__ (CBPS) heatmap indicator, a bank of 2-pole
   * band-pass filters (one per cycle period) fed by a Butterworth highpass + Super Smoother
   * pre-filter cascade, following EasyLanguage listing 10-1. */
  CombBandPassSpectrum,

  /** Identifies the Ehlers __Autocorrelation Indicator__ (ACI) heatmap, a bank of Pearson
   * correlation coefficients between the current filtered series and a lagged copy of itself,
   * following EasyLanguage listing 8-2. */
  AutoCorrelationIndicator,

  /** Identifies the Ehlers __Autocorrelation Periodogram__ (ACP) heatmap, a discrete Fourier
   * transform of the autocorrelation function over a configurable cycle-period range, following
   * EasyLanguage listing 8-3. */
  AutoCorrelationPeriodogram,

  /** Identifies the Jurik __Relative Trend Strength Index__ (RSX) indicator. */
  JurikRelativeTrendStrengthIndex,

  /** Identifies the Jurik __Composite Fractal Behavior Index__ (CFB) indicator. */
  JurikCompositeFractalBehaviorIndex,

  /** Identifies the Jurik __Zero Lag Velocity__ (VEL) indicator. */
  JurikZeroLagVelocity,

  /** Identifies the Jurik __Directional Movement Index__ (DMX) indicator. */
  JurikDirectionalMovementIndex,

  /** Identifies the Jurik __Adaptive Relative Trend Strength Index__ (JARSX) indicator. */
  JurikAdaptiveRelativeTrendStrengthIndex,

  /** Identifies the Jurik __Adaptive Zero Lag Velocity__ (JAVEL) indicator. */
  JurikAdaptiveZeroLagVelocity,

  /** Identifies the Jurik __Commodity Channel Index__ (JCCX) indicator. */
  JurikCommodityChannelIndex,

  /** Identifies the Jurik __Fractal Adaptive Zero Lag Velocity__ (JVELCFB) indicator. */
  JurikFractalAdaptiveZeroLagVelocity,

  /** Identifies the Jurik __Turning Point Oscillator__ (JTPO) indicator. */
  JurikTurningPointOscillator,

  /** Identifies the Jurik __Wavelet Sampler__ (WAV) indicator. */
  JurikWaveletSampler,

  /** Identifies the Arnaud Legoux __Moving Average__ (ALMA) indicator. */
  ArnaudLegouxMovingAverage,
}
