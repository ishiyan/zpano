/** Enumerates indicators. */
export enum IndicatorIdentifier {

    // ── common ────────────────────────────────────────────────────────────

    /** Identifies the __Absolute Price Oscillator__ (APO) indicator. */
    AbsolutePriceOscillator,

    /** Identifies the __Exponential Moving Average__ (EMA) indicator. */
    ExponentialMovingAverage,

    /** Identifies the __Linear Regression__ (LINREG) indicator. */
    LinearRegression,

    /** Identifies the __Momentum__ (MOM) indicator. */
    Momentum,

    /** Identifies __Pearson's Correlation Coefficient__ (CORREL) indicator. */
    PearsonsCorrelationCoefficient,

    /** Identifies the __Rate of Change__ (ROC) indicator. */
    RateOfChange,

    /** Identifies the __Rate of Change Percent__ (ROCP) indicator. */
    RateOfChangePercent,

    /** Identifies the __Rate of Change Ratio__ (ROCR / ROCR100) indicator. */
    RateOfChangeRatio,

    /** Identifies the __Simple Moving Average__ (SMA) indicator. */
    SimpleMovingAverage,

    /** Identifies the __Standard Deviation__ (STDEV) indicator. */
    StandardDeviation,

    /** Identifies the __Triangular Moving Average__ (TRIMA) indicator. */
    TriangularMovingAverage,

    /** Identifies the __Variance__ (VAR) indicator. */
    Variance,

    /** Identifies the __Weighted Moving Average__ (WMA) indicator. */
    WeightedMovingAverage,

    // ── arnaudlegoux ──────────────────────────────────────────────────────

    /** Identifies the Arnaud Legoux __Moving Average__ (ALMA) indicator. */
    ArnaudLegouxMovingAverage,

    // ── donaldlambert ─────────────────────────────────────────────────────

    /** Identifies Donald Lambert's __Commodity Channel Index__ (CCI) indicator. */
    CommodityChannelIndex,

    // ── genequong ─────────────────────────────────────────────────────────

    /** Identifies Gene Quong's __Money Flow Index__ (MFI) indicator. */
    MoneyFlowIndex,

    // ── georgelane ────────────────────────────────────────────────────────

    /** Identifies George Lane's __Stochastic Oscillator__ (STOCH) indicator. */
    Stochastic,

    // ── geraldappel ───────────────────────────────────────────────────────

    /** Identifies Gerald Appel's __Moving Average Convergence Divergence__ (MACD) indicator. */
    MovingAverageConvergenceDivergence,

    /** Identifies Gerald Appel's __Percentage Price Oscillator__ (PPO) indicator. */
    PercentagePriceOscillator,

    // ── igorlivshin ───────────────────────────────────────────────────────

    /** Identifies Igor Livshin's __Balance of Power__ (BOP) indicator. */
    BalanceOfPower,

    // ── jackhutson ────────────────────────────────────────────────────────

    /** Identifies Jack Hutson's __Triple Exponential Moving Average Oscillator__ (TRIX) indicator. */
    TripleExponentialMovingAverageOscillator,

    // ── johnbollinger ─────────────────────────────────────────────────────

    /** Identifies the __Bollinger Bands__ (BB) indicator. */
    BollingerBands,

    /** Identifies John Bollinger's __Bollinger Bands Trend__ (BBTrend) indicator. */
    BollingerBandsTrend,

    // ── johnehlers ────────────────────────────────────────────────────────

    /** Identifies the Ehlers __Autocorrelation Indicator__ (ACI) heatmap, a bank of Pearson
     * correlation coefficients between the current filtered series and a lagged copy of itself,
     * following EasyLanguage listing 8-2. */
    AutoCorrelationIndicator,

    /** Identifies the Ehlers __Autocorrelation Periodogram__ (ACP) heatmap, a discrete Fourier
     * transform of the autocorrelation function over a configurable cycle-period range, following
     * EasyLanguage listing 8-3. */
    AutoCorrelationPeriodogram,

    /** Identifies the Ehler's __Center of Gravity Oscillator__ (COG) indicator. */
    CenterOfGravityOscillator,

    /** Identifies the Ehlers __Comb Band-Pass Spectrum__ (CBPS) heatmap indicator, a bank of 2-pole
     * band-pass filters (one per cycle period) fed by a Butterworth highpass + Super Smoother
     * pre-filter cascade, following EasyLanguage listing 10-1. */
    CombBandPassSpectrum,

    /** Identifies the Ehlers __Corona Signal To Noise Ratio__ (CSNR) heatmap indicator, exposing the intensity
     * raster heatmap column and the current SNR mapped into the parameter range. */
    CoronaSignalToNoiseRatio,

    /** Identifies the Ehlers __Corona Spectrum__ (CSPECT) heatmap indicator, exposing the dB heatmap column,
     * the weighted dominant cycle estimate and its 5-sample median. */
    CoronaSpectrum,

    /** Identifies the Ehlers __Corona Swing Position__ (CSWING) heatmap indicator, exposing the intensity raster
     * heatmap column and the current swing position mapped into the parameter range. */
    CoronaSwingPosition,

    /** Identifies the Ehlers __Corona Trend Vigor__ (CTV) heatmap indicator, exposing the intensity raster heatmap
     * column and the current trend vigor scaled into the parameter range. */
    CoronaTrendVigor,

    /** Identifies the Ehler's __Cyber Cycle__ (CC) indicator. */
    CyberCycle,

    /** Identifies the Ehlers __Discrete Fourier Transform Spectrum__ (DFTPS) heatmap indicator, a
     * mean-subtracted DFT power spectrum over a configurable cycle-period range. */
    DiscreteFourierTransformSpectrum,

    /** Identifies the Ehler's __Dominant Cycle__ (DC) indicator, exposing raw period, smoothed period and phase. */
    DominantCycle,

    /** Identifies the Ehler's __Fractal Adaptive Moving Average__ (FRAMA) indicator. */
    FractalAdaptiveMovingAverage,

    /** Identifies the Ehlers __Hilbert Transformer Instantaneous Trend Line__ (HTITL) indicator, exposing trend value and dominant cycle period. */
    HilbertTransformerInstantaneousTrendLine,

    /** Identifies the Ehler's __Instantaneous Trend Line__ (iTrend) indicator. */
    InstantaneousTrendLine,

    /** Identifies the Ehler's __MESA Adaptive Moving Average__ (MAMA) indicator. */
    MesaAdaptiveMovingAverage,

    /** Identifies the Ehler's __Roofing Filter__ indicator. */
    RoofingFilter,

    /** Identifies the Ehlers __Sine Wave__ (SW) indicator, exposing sine value, lead sine, band, dominant cycle period and phase. */
    SineWave,

    /** Identifies the Ehler's __Super Smoother__ (SS) indicator. */
    SuperSmoother,

    /** Identifies the Ehlers __Trend / Cycle Mode__ (TCM) indicator, exposing the trend/cycle value (+1 in trend, −1 in cycle),
     * trend/cycle mode flags, instantaneous trend line, sine wave, lead sine wave, dominant cycle period and phase. */
    TrendCycleMode,

    /** Identifies the Ehler's __Zero-lag Error-Correcting Exponential Moving Average__ (ZECEMA) indicator. */
    ZeroLagErrorCorrectingExponentialMovingAverage,

    /** Identifies the Ehler's __Zero-lag Exponential Moving Average__ (ZEMA) indicator. */
    ZeroLagExponentialMovingAverage,

    // ── josephgranville ───────────────────────────────────────────────────

    /** Identifies Joseph Granville's __On-Balance Volume__ (OBV) indicator. */
    OnBalanceVolume,

    // ── larrywilliams ─────────────────────────────────────────────────────

    /** Identifies Larry Williams' __Ultimate Oscillator__ (ULTOSC) indicator. */
    UltimateOscillator,

    /** Identifies Larry Williams' __Williams %R__ (WILL%R) indicator. */
    WilliamsPercentR,

    // ── manfreddurschner ──────────────────────────────────────────────────

    /** Identifies the New Moving Average (NMA) indicator by Dürschner. */
    NewMovingAverage,

    // ── marcchaikin ───────────────────────────────────────────────────────

    /** Identifies Marc Chaikin's __Advance-Decline__ (A/D) indicator. */
    AdvanceDecline,

    /** Identifies Marc Chaikin's __Advance-Decline Oscillator__ (ADOSC) indicator. */
    AdvanceDeclineOscillator,

    // ── markjurik ─────────────────────────────────────────────────────────

    /** Identifies the Jurik __Adaptive Relative Trend Strength Index__ (JARSX) indicator. */
    JurikAdaptiveRelativeTrendStrengthIndex,

    /** Identifies the Jurik __Adaptive Zero Lag Velocity__ (JAVEL) indicator. */
    JurikAdaptiveZeroLagVelocity,

    /** Identifies the Jurik __Commodity Channel Index__ (JCCX) indicator. */
    JurikCommodityChannelIndex,

    /** Identifies the Jurik __Composite Fractal Behavior Index__ (CFB) indicator. */
    JurikCompositeFractalBehaviorIndex,

    /** Identifies the Jurik __Directional Movement Index__ (DMX) indicator. */
    JurikDirectionalMovementIndex,

    /** Identifies the Jurik __Fractal Adaptive Zero Lag Velocity__ (JVELCFB) indicator. */
    JurikFractalAdaptiveZeroLagVelocity,

    /** Identifies the Jurik __Moving Average__ (JMA) indicator. */
    JurikMovingAverage,

    /** Identifies the Jurik __Relative Trend Strength Index__ (RSX) indicator. */
    JurikRelativeTrendStrengthIndex,

    /** Identifies the Jurik __Turning Point Oscillator__ (JTPO) indicator. */
    JurikTurningPointOscillator,

    /** Identifies the Jurik __Wavelet Sampler__ (WAV) indicator. */
    JurikWaveletSampler,

    /** Identifies the Jurik __Zero Lag Velocity__ (VEL) indicator. */
    JurikZeroLagVelocity,

    // ── patrickmulloy ─────────────────────────────────────────────────────

    /** Identifies the __Double Exponential Moving Average__ (DEMA) indicator. */
    DoubleExponentialMovingAverage,

    /** Identifies the __Triple Exponential Moving Average__ (TEMA) indicator. */
    TripleExponentialMovingAverage,

    // ── perrykaufman ──────────────────────────────────────────────────────

    /** Identifies the __Kaufman Adaptive Moving Average__ (KAMA) indicator. */
    KaufmanAdaptiveMovingAverage,

    // ── timtillson ────────────────────────────────────────────────────────

    /** Identifies the __T2 Exponential Moving Average__ (T2EMA) indicator. */
    T2ExponentialMovingAverage,

    /** Identifies the __T3 Exponential Moving Average__ (T3EMA) indicator. */
    T3ExponentialMovingAverage,

    // ── tusharchande ──────────────────────────────────────────────────────

    /** Identifies Tushar Chande's __Aroon__ (AROON) indicator. */
    Aroon,

    /** Identifies the __Chande Momentum Oscillator__ (CMO) indicator. */
    ChandeMomentumOscillator,

    /** Identifies Tushar Chande's __Stochastic Relative Strength Index__ (STOCHRSI) indicator. */
    StochasticRelativeStrengthIndex,

    // ── vladimirkravchuk ──────────────────────────────────────────────────

    /** Identifies Vladimir Kravchuk's __Adaptive Trend and Cycle Filter__ (ATCF) suite: a bank of five FIR filters
     * (FATL, SATL, RFTL, RSTL, RBCI) plus three composites (FTLM, STLM, PCCI) applied to a single input series. */
    AdaptiveTrendAndCycleFilter,

    // ── welleswilder ──────────────────────────────────────────────────────

    /** Identifies Welles Wilder's __Average Directional Movement Index__ (ADX) indicator. */
    AverageDirectionalMovementIndex,

    /** Identifies Welles Wilder's __Average Directional Movement Index Rating__ (ADXR) indicator. */
    AverageDirectionalMovementIndexRating,

    /** Identifies Welles Wilder's __Average True Range__ (ATR) indicator. */
    AverageTrueRange,

    /** Identifies Welles Wilder's __Directional Indicator Minus__ (-DI) indicator. */
    DirectionalIndicatorMinus,

    /** Identifies Welles Wilder's __Directional Indicator Plus__ (+DI) indicator. */
    DirectionalIndicatorPlus,

    /** Identifies Welles Wilder's __Directional Movement Index__ (DX) indicator. */
    DirectionalMovementIndex,

    /** Identifies Welles Wilder's __Directional Movement Minus__ (-DM) indicator. */
    DirectionalMovementMinus,

    /** Identifies Welles Wilder's __Directional Movement Plus__ (+DM) indicator. */
    DirectionalMovementPlus,

    /** Identifies Welles Wilder's __Normalized Average True Range__ (NATR) indicator. */
    NormalizedAverageTrueRange,

    /** Identifies Welles Wilder's __Parabolic Stop And Reverse__ (SAR) indicator. */
    ParabolicStopAndReverse,

    /** Identifies the __Relative Strength Index__ (RSI) indicator. */
    RelativeStrengthIndex,

    /** Identifies Welles Wilder's __True Range__ (TR) indicator. */
    TrueRange,

    // ── custom ────────────────────────────────────────────────────────────

    /** Identifies the __Goertzel power spectrum__ (GOERTZEL) indicator. */
    GoertzelSpectrum,

    /** Identifies the __Maximum Entropy Spectrum__ (MESPECT) heatmap indicator, a Burg maximum-entropy
     * auto-regressive power spectrum over a configurable cycle-period range. */
    MaximumEntropySpectrum,
}
