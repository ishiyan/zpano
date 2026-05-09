/// Identifies an indicator by enumerating all implemented indicators.
/// Values are 0-based (unlike Go's iota+1).
#[derive(Debug, Clone, Copy, PartialEq, Eq, Hash)]
#[repr(u8)]
pub enum Identifier {
    // ── common ────────────────────────────────────────────────────────────
    /// Identifies the Absolute Price Oscillator (APO) indicator.
    AbsolutePriceOscillator = 0,
    /// Identifies the Exponential Moving Average (EMA) indicator.
    ExponentialMovingAverage = 1,
    /// Identifies the Linear Regression (LINEARREG) indicator.
    LinearRegression = 2,
    /// Identifies the Momentum (MOM) indicator.
    Momentum = 3,
    /// Identifies the Pearson's Correlation Coefficient (CORREL) indicator.
    PearsonsCorrelationCoefficient = 4,
    /// Identifies the Rate of Change (ROC) indicator.
    RateOfChange = 5,
    /// Identifies the Rate of Change Percent (ROCP) indicator.
    RateOfChangePercent = 6,
    /// Identifies the Rate of Change Ratio (ROCR / ROCR100) indicator.
    RateOfChangeRatio = 7,
    /// Identifies the Simple Moving Average (SMA) indicator.
    SimpleMovingAverage = 8,
    /// Identifies the Standard Deviation (STDEV) indicator.
    StandardDeviation = 9,
    /// Identifies the Triangular Moving Average (TRIMA) indicator.
    TriangularMovingAverage = 10,
    /// Identifies the Variance (VAR) indicator.
    Variance = 11,
    /// Identifies the Weighted Moving Average (WMA) indicator.
    WeightedMovingAverage = 12,

    // ── arnaud legoux ──────────────────────────────────────────────────────
    /// Identifies the Arnaud Legoux Moving Average (ALMA) indicator.
    ArnaudLegouxMovingAverage = 13,

    // ── donald lambert ─────────────────────────────────────────────────────
    /// Identifies the Donald Lambert Commodity Channel Index (CCI) indicator.
    CommodityChannelIndex = 14,

    // ── gene quong ─────────────────────────────────────────────────────────
    /// Identifies the Gene Quong Money Flow Index (MFI) indicator.
    MoneyFlowIndex = 15,

    // ── george lane ────────────────────────────────────────────────────────
    /// Identifies the George Lane Stochastic Oscillator (STOCH) indicator.
    Stochastic = 16,

    // ── gerald appel ───────────────────────────────────────────────────────
    /// Identifies Gerald Appel's Moving Average Convergence Divergence (MACD) indicator.
    MovingAverageConvergenceDivergence = 17,
    /// Identifies the Gerald Appel Percentage Price Oscillator (PPO) indicator.
    PercentagePriceOscillator = 18,

    // ── igor livshin ───────────────────────────────────────────────────────
    /// Identifies the Igor Livshin Balance of Power (BOP) indicator.
    BalanceOfPower = 19,

    // ── jack hutson ────────────────────────────────────────────────────────
    /// Identifies Jack Hutson's Triple Exponential Moving Average Oscillator (TRIX) indicator.
    TripleExponentialMovingAverageOscillator = 20,

    // ── john bollinger ─────────────────────────────────────────────────────
    /// Identifies the Bollinger Bands (BB) indicator.
    BollingerBands = 21,
    /// Identifies John Bollinger's Bollinger Bands Trend (BBTrend) indicator.
    BollingerBandsTrend = 22,

    // ── john ehlers ────────────────────────────────────────────────────────
    /// Identifies the Ehlers Autocorrelation Indicator (ACI) heatmap, a bank of Pearson
    /// correlation coefficients between the current filtered series and a lagged copy
    /// of itself, following EasyLanguage listing 8-2.
    AutoCorrelationIndicator = 23,
    /// Identifies the Ehlers Autocorrelation Periodogram (ACP) heatmap, a discrete Fourier
    /// transform of the autocorrelation function over a configurable cycle-period range,
    /// following EasyLanguage listing 8-3.
    AutoCorrelationPeriodogram = 24,
    /// Identifies the Ehlers Center of Gravity Oscillator (COG) indicator.
    CenterOfGravityOscillator = 25,
    /// Identifies the Ehlers Comb Band-Pass Spectrum (CBPS) heatmap indicator, a bank of
    /// 2-pole band-pass filters (one per cycle period) fed by a Butterworth highpass +
    /// Super Smoother pre-filter cascade, following EasyLanguage listing 10-1.
    CombBandPassSpectrum = 26,
    /// Identifies the Ehlers Corona Signal To Noise Ratio (CSNR) heatmap indicator, exposing
    /// the intensity raster heatmap column and the current SNR mapped into the parameter range.
    CoronaSignalToNoiseRatio = 27,
    /// Identifies the Ehlers Corona Spectrum (CSPECT) heatmap indicator, exposing the dB
    /// heatmap column, the weighted dominant cycle estimate and its 5-sample median.
    CoronaSpectrum = 28,
    /// Identifies the Ehlers Corona Swing Position (CSWING) heatmap indicator, exposing the
    /// intensity raster heatmap column and the current swing position mapped into the parameter range.
    CoronaSwingPosition = 29,
    /// Identifies the Ehlers Corona Trend Vigor (CTV) heatmap indicator, exposing the intensity
    /// raster heatmap column and the current trend vigor scaled into the parameter range.
    CoronaTrendVigor = 30,
    /// Identifies the Ehlers Cyber Cycle (CC) indicator.
    CyberCycle = 31,
    /// Identifies the Ehlers Discrete Fourier Transform Spectrum (DFTPS) heatmap indicator,
    /// a mean-subtracted DFT power spectrum over a configurable cycle-period range.
    DiscreteFourierTransformSpectrum = 32,
    /// Identifies the Ehlers Dominant Cycle (DC) indicator, exposing raw period, smoothed period and phase.
    DominantCycle = 33,
    /// Identifies the Ehlers Fractal Adaptive Moving Average (FRAMA) indicator.
    FractalAdaptiveMovingAverage = 34,
    /// Identifies the Ehlers Hilbert Transformer Instantaneous Trend Line (HTITL) indicator,
    /// exposing trend value and dominant cycle period.
    HilbertTransformerInstantaneousTrendLine = 35,
    /// Identifies the Ehlers Instantaneous Trend Line (iTrend) indicator.
    InstantaneousTrendLine = 36,
    /// Identifies the Ehlers MESA Adaptive Moving Average (MAMA) indicator.
    MesaAdaptiveMovingAverage = 37,
    /// Identifies the Ehlers Roofing Filter indicator.
    RoofingFilter = 38,
    /// Identifies the Ehlers Sine Wave (SW) indicator, exposing sine value, lead sine,
    /// band, dominant cycle period and phase.
    SineWave = 39,
    /// Identifies the Ehlers Super Smoother (SS) indicator.
    SuperSmoother = 40,
    /// Identifies the Ehlers Trend / Cycle Mode (TCM) indicator, exposing the trend/cycle
    /// value (+1 in trend, -1 in cycle), trend/cycle mode flags, instantaneous trend line,
    /// sine wave, lead sine wave, dominant cycle period and phase.
    TrendCycleMode = 41,
    /// Identifies the Ehlers Zero-lag Error-Correcting Exponential Moving Average (ZECEMA) indicator.
    ZeroLagErrorCorrectingExponentialMovingAverage = 42,
    /// Identifies the Ehlers Zero-lag Exponential Moving Average (ZEMA) indicator.
    ZeroLagExponentialMovingAverage = 43,

    // ── joseph granville ───────────────────────────────────────────────────
    /// Identifies the Joseph Granville On-Balance Volume (OBV) indicator.
    OnBalanceVolume = 44,

    // ── larry williams ─────────────────────────────────────────────────────
    /// Identifies the Larry Williams Ultimate Oscillator (ULTOSC) indicator.
    UltimateOscillator = 45,
    /// Identifies the Larry Williams Williams %R (WILL%R) indicator.
    WilliamsPercentR = 46,

    // ── manfred durschner ──────────────────────────────────────────────────
    /// Identifies the New Moving Average (NMA) indicator by Durschner.
    NewMovingAverage = 47,

    // ── marc chaikin ───────────────────────────────────────────────────────
    /// Identifies the Marc Chaikin Advance-Decline (A/D) indicator.
    AdvanceDecline = 48,
    /// Identifies the Marc Chaikin Advance-Decline Oscillator (ADOSC) indicator.
    AdvanceDeclineOscillator = 49,

    // ── mark jurik ─────────────────────────────────────────────────────────
    /// Identifies the Jurik Adaptive Relative Trend Strength Index (JARSX) indicator.
    JurikAdaptiveRelativeTrendStrengthIndex = 50,
    /// Identifies the Jurik Adaptive Zero Lag Velocity (JAVEL) indicator.
    JurikAdaptiveZeroLagVelocity = 51,
    /// Identifies the Jurik Commodity Channel Index (JCCX) indicator.
    JurikCommodityChannelIndex = 52,
    /// Identifies the Jurik Composite Fractal Behavior Index (CFB) indicator.
    JurikCompositeFractalBehaviorIndex = 53,
    /// Identifies the Jurik Directional Movement Index (DMX) indicator.
    JurikDirectionalMovementIndex = 54,
    /// Identifies the Jurik Fractal Adaptive Zero Lag Velocity (JVELCFB) indicator.
    JurikFractalAdaptiveZeroLagVelocity = 55,
    /// Identifies the Jurik Moving Average (JMA) indicator.
    JurikMovingAverage = 56,
    /// Identifies the Jurik Relative Trend Strength Index (RSX) indicator.
    JurikRelativeTrendStrengthIndex = 57,
    /// Identifies the Jurik Turning Point Oscillator (JTPO) indicator.
    JurikTurningPointOscillator = 58,
    /// Identifies the Jurik Wavelet Sampler (WAV) indicator.
    JurikWaveletSampler = 59,
    /// Identifies the Jurik Zero Lag Velocity (VEL) indicator.
    JurikZeroLagVelocity = 60,

    // ── patrick mulloy ─────────────────────────────────────────────────────
    /// Identifies the Double Exponential Moving Average (DEMA) indicator.
    DoubleExponentialMovingAverage = 61,
    /// Identifies the Triple Exponential Moving Average (TEMA) indicator.
    TripleExponentialMovingAverage = 62,

    // ── perry kaufman ──────────────────────────────────────────────────────
    /// Identifies the Kaufman Adaptive Moving Average (KAMA) indicator.
    KaufmanAdaptiveMovingAverage = 63,

    // ── tim tillson ────────────────────────────────────────────────────────
    /// Identifies the T2 Exponential Moving Average (T2EMA) indicator.
    T2ExponentialMovingAverage = 64,
    /// Identifies the T3 Exponential Moving Average (T3EMA) indicator.
    T3ExponentialMovingAverage = 65,

    // ── tushar chande ──────────────────────────────────────────────────────
    /// Identifies the Tushar Chande Aroon (AROON) indicator.
    Aroon = 66,
    /// Identifies the Chande Momentum Oscillator (CMO) indicator.
    ChandeMomentumOscillator = 67,
    /// Identifies the Tushar Chande Stochastic RSI (STOCHRSI) indicator.
    StochasticRelativeStrengthIndex = 68,

    // ── vladimir kravchuk ──────────────────────────────────────────────────
    /// Identifies Vladimir Kravchuk's Adaptive Trend and Cycle Filter (ATCF) suite: a bank
    /// of five FIR filters (FATL, SATL, RFTL, RSTL, RBCI) plus three composites (FTLM, STLM, PCCI).
    AdaptiveTrendAndCycleFilter = 69,

    // ── welles wilder ──────────────────────────────────────────────────────
    /// Identifies the Welles Wilder Average Directional Movement Index (ADX) indicator.
    AverageDirectionalMovementIndex = 70,
    /// Identifies the Welles Wilder Average Directional Movement Index Rating (ADXR) indicator.
    AverageDirectionalMovementIndexRating = 71,
    /// Identifies the Welles Wilder Average True Range (ATR) indicator.
    AverageTrueRange = 72,
    /// Identifies the Welles Wilder Directional Indicator Minus (-DI) indicator.
    DirectionalIndicatorMinus = 73,
    /// Identifies the Welles Wilder Directional Indicator Plus (+DI) indicator.
    DirectionalIndicatorPlus = 74,
    /// Identifies the Welles Wilder Directional Movement Index (DX) indicator.
    DirectionalMovementIndex = 75,
    /// Identifies the Welles Wilder Directional Movement Minus (-DM) indicator.
    DirectionalMovementMinus = 76,
    /// Identifies the Welles Wilder Directional Movement Plus (+DM) indicator.
    DirectionalMovementPlus = 77,
    /// Identifies the Welles Wilder Normalized Average True Range (NATR) indicator.
    NormalizedAverageTrueRange = 78,
    /// Identifies the Welles Wilder Parabolic Stop And Reverse (SAR) indicator.
    ParabolicStopAndReverse = 79,
    /// Identifies the Relative Strength Index (RSI) indicator.
    RelativeStrengthIndex = 80,
    /// Identifies the Welles Wilder True Range (TR) indicator.
    TrueRange = 81,

    // ── custom ────────────────────────────────────────────────────────────
    /// Identifies the Goertzel power spectrum (GOERTZEL) indicator.
    GoertzelSpectrum = 82,
    /// Identifies the Maximum Entropy Spectrum (MESPECT) heatmap indicator, a Burg
    /// maximum-entropy auto-regressive power spectrum over a configurable cycle-period range.
    MaximumEntropySpectrum = 83,
}

impl Identifier {
    /// Returns the camelCase string representation matching Go's String().
    pub fn as_str(&self) -> &'static str {
        match self {
            Self::AbsolutePriceOscillator => "absolutePriceOscillator",
            Self::ExponentialMovingAverage => "exponentialMovingAverage",
            Self::LinearRegression => "linearRegression",
            Self::Momentum => "momentum",
            Self::PearsonsCorrelationCoefficient => "pearsonsCorrelationCoefficient",
            Self::RateOfChange => "rateOfChange",
            Self::RateOfChangePercent => "rateOfChangePercent",
            Self::RateOfChangeRatio => "rateOfChangeRatio",
            Self::SimpleMovingAverage => "simpleMovingAverage",
            Self::StandardDeviation => "standardDeviation",
            Self::TriangularMovingAverage => "triangularMovingAverage",
            Self::Variance => "variance",
            Self::WeightedMovingAverage => "weightedMovingAverage",
            Self::ArnaudLegouxMovingAverage => "arnaudLegouxMovingAverage",
            Self::CommodityChannelIndex => "commodityChannelIndex",
            Self::MoneyFlowIndex => "moneyFlowIndex",
            Self::Stochastic => "stochastic",
            Self::MovingAverageConvergenceDivergence => "movingAverageConvergenceDivergence",
            Self::PercentagePriceOscillator => "percentagePriceOscillator",
            Self::BalanceOfPower => "balanceOfPower",
            Self::TripleExponentialMovingAverageOscillator => {
                "tripleExponentialMovingAverageOscillator"
            }
            Self::BollingerBands => "bollingerBands",
            Self::BollingerBandsTrend => "bollingerBandsTrend",
            Self::AutoCorrelationIndicator => "autoCorrelationIndicator",
            Self::AutoCorrelationPeriodogram => "autoCorrelationPeriodogram",
            Self::CenterOfGravityOscillator => "centerOfGravityOscillator",
            Self::CombBandPassSpectrum => "combBandPassSpectrum",
            Self::CoronaSignalToNoiseRatio => "coronaSignalToNoiseRatio",
            Self::CoronaSpectrum => "coronaSpectrum",
            Self::CoronaSwingPosition => "coronaSwingPosition",
            Self::CoronaTrendVigor => "coronaTrendVigor",
            Self::CyberCycle => "cyberCycle",
            Self::DiscreteFourierTransformSpectrum => "discreteFourierTransformSpectrum",
            Self::DominantCycle => "dominantCycle",
            Self::FractalAdaptiveMovingAverage => "fractalAdaptiveMovingAverage",
            Self::HilbertTransformerInstantaneousTrendLine => {
                "hilbertTransformerInstantaneousTrendLine"
            }
            Self::InstantaneousTrendLine => "instantaneousTrendLine",
            Self::MesaAdaptiveMovingAverage => "mesaAdaptiveMovingAverage",
            Self::RoofingFilter => "roofingFilter",
            Self::SineWave => "sineWave",
            Self::SuperSmoother => "superSmoother",
            Self::TrendCycleMode => "trendCycleMode",
            Self::ZeroLagErrorCorrectingExponentialMovingAverage => {
                "zeroLagErrorCorrectingExponentialMovingAverage"
            }
            Self::ZeroLagExponentialMovingAverage => "zeroLagExponentialMovingAverage",
            Self::OnBalanceVolume => "onBalanceVolume",
            Self::UltimateOscillator => "ultimateOscillator",
            Self::WilliamsPercentR => "williamsPercentR",
            Self::NewMovingAverage => "newMovingAverage",
            Self::AdvanceDecline => "advanceDecline",
            Self::AdvanceDeclineOscillator => "advanceDeclineOscillator",
            Self::JurikAdaptiveRelativeTrendStrengthIndex => {
                "jurikAdaptiveRelativeTrendStrengthIndex"
            }
            Self::JurikAdaptiveZeroLagVelocity => "jurikAdaptiveZeroLagVelocity",
            Self::JurikCommodityChannelIndex => "jurikCommodityChannelIndex",
            Self::JurikCompositeFractalBehaviorIndex => "jurikCompositeFractalBehaviorIndex",
            Self::JurikDirectionalMovementIndex => "jurikDirectionalMovementIndex",
            Self::JurikFractalAdaptiveZeroLagVelocity => "jurikFractalAdaptiveZeroLagVelocity",
            Self::JurikMovingAverage => "jurikMovingAverage",
            Self::JurikRelativeTrendStrengthIndex => "jurikRelativeTrendStrengthIndex",
            Self::JurikTurningPointOscillator => "jurikTurningPointOscillator",
            Self::JurikWaveletSampler => "jurikWaveletSampler",
            Self::JurikZeroLagVelocity => "jurikZeroLagVelocity",
            Self::DoubleExponentialMovingAverage => "doubleExponentialMovingAverage",
            Self::TripleExponentialMovingAverage => "tripleExponentialMovingAverage",
            Self::KaufmanAdaptiveMovingAverage => "kaufmanAdaptiveMovingAverage",
            Self::T2ExponentialMovingAverage => "t2ExponentialMovingAverage",
            Self::T3ExponentialMovingAverage => "t3ExponentialMovingAverage",
            Self::Aroon => "aroon",
            Self::ChandeMomentumOscillator => "chandeMomentumOscillator",
            Self::StochasticRelativeStrengthIndex => "stochasticRelativeStrengthIndex",
            Self::AdaptiveTrendAndCycleFilter => "adaptiveTrendAndCycleFilter",
            Self::AverageDirectionalMovementIndex => "averageDirectionalMovementIndex",
            Self::AverageDirectionalMovementIndexRating => "averageDirectionalMovementIndexRating",
            Self::AverageTrueRange => "averageTrueRange",
            Self::DirectionalIndicatorMinus => "directionalIndicatorMinus",
            Self::DirectionalIndicatorPlus => "directionalIndicatorPlus",
            Self::DirectionalMovementIndex => "directionalMovementIndex",
            Self::DirectionalMovementMinus => "directionalMovementMinus",
            Self::DirectionalMovementPlus => "directionalMovementPlus",
            Self::NormalizedAverageTrueRange => "normalizedAverageTrueRange",
            Self::ParabolicStopAndReverse => "parabolicStopAndReverse",
            Self::RelativeStrengthIndex => "relativeStrengthIndex",
            Self::TrueRange => "trueRange",
            Self::GoertzelSpectrum => "goertzelSpectrum",
            Self::MaximumEntropySpectrum => "maximumEntropySpectrum",
        }
    }

    /// Parses a camelCase string into an Identifier.
    pub fn from_str(s: &str) -> Option<Self> {
        match s {
            "absolutePriceOscillator" => Some(Self::AbsolutePriceOscillator),
            "exponentialMovingAverage" => Some(Self::ExponentialMovingAverage),
            "linearRegression" => Some(Self::LinearRegression),
            "momentum" => Some(Self::Momentum),
            "pearsonsCorrelationCoefficient" => Some(Self::PearsonsCorrelationCoefficient),
            "rateOfChange" => Some(Self::RateOfChange),
            "rateOfChangePercent" => Some(Self::RateOfChangePercent),
            "rateOfChangeRatio" => Some(Self::RateOfChangeRatio),
            "simpleMovingAverage" => Some(Self::SimpleMovingAverage),
            "standardDeviation" => Some(Self::StandardDeviation),
            "triangularMovingAverage" => Some(Self::TriangularMovingAverage),
            "variance" => Some(Self::Variance),
            "weightedMovingAverage" => Some(Self::WeightedMovingAverage),
            "arnaudLegouxMovingAverage" => Some(Self::ArnaudLegouxMovingAverage),
            "commodityChannelIndex" => Some(Self::CommodityChannelIndex),
            "moneyFlowIndex" => Some(Self::MoneyFlowIndex),
            "stochastic" => Some(Self::Stochastic),
            "movingAverageConvergenceDivergence" => Some(Self::MovingAverageConvergenceDivergence),
            "percentagePriceOscillator" => Some(Self::PercentagePriceOscillator),
            "balanceOfPower" => Some(Self::BalanceOfPower),
            "tripleExponentialMovingAverageOscillator" => {
                Some(Self::TripleExponentialMovingAverageOscillator)
            }
            "bollingerBands" => Some(Self::BollingerBands),
            "bollingerBandsTrend" => Some(Self::BollingerBandsTrend),
            "autoCorrelationIndicator" => Some(Self::AutoCorrelationIndicator),
            "autoCorrelationPeriodogram" => Some(Self::AutoCorrelationPeriodogram),
            "centerOfGravityOscillator" => Some(Self::CenterOfGravityOscillator),
            "combBandPassSpectrum" => Some(Self::CombBandPassSpectrum),
            "coronaSignalToNoiseRatio" => Some(Self::CoronaSignalToNoiseRatio),
            "coronaSpectrum" => Some(Self::CoronaSpectrum),
            "coronaSwingPosition" => Some(Self::CoronaSwingPosition),
            "coronaTrendVigor" => Some(Self::CoronaTrendVigor),
            "cyberCycle" => Some(Self::CyberCycle),
            "discreteFourierTransformSpectrum" => Some(Self::DiscreteFourierTransformSpectrum),
            "dominantCycle" => Some(Self::DominantCycle),
            "fractalAdaptiveMovingAverage" => Some(Self::FractalAdaptiveMovingAverage),
            "hilbertTransformerInstantaneousTrendLine" => {
                Some(Self::HilbertTransformerInstantaneousTrendLine)
            }
            "instantaneousTrendLine" => Some(Self::InstantaneousTrendLine),
            "mesaAdaptiveMovingAverage" => Some(Self::MesaAdaptiveMovingAverage),
            "roofingFilter" => Some(Self::RoofingFilter),
            "sineWave" => Some(Self::SineWave),
            "superSmoother" => Some(Self::SuperSmoother),
            "trendCycleMode" => Some(Self::TrendCycleMode),
            "zeroLagErrorCorrectingExponentialMovingAverage" => {
                Some(Self::ZeroLagErrorCorrectingExponentialMovingAverage)
            }
            "zeroLagExponentialMovingAverage" => Some(Self::ZeroLagExponentialMovingAverage),
            "onBalanceVolume" => Some(Self::OnBalanceVolume),
            "ultimateOscillator" => Some(Self::UltimateOscillator),
            "williamsPercentR" => Some(Self::WilliamsPercentR),
            "newMovingAverage" => Some(Self::NewMovingAverage),
            "advanceDecline" => Some(Self::AdvanceDecline),
            "advanceDeclineOscillator" => Some(Self::AdvanceDeclineOscillator),
            "jurikAdaptiveRelativeTrendStrengthIndex" => {
                Some(Self::JurikAdaptiveRelativeTrendStrengthIndex)
            }
            "jurikAdaptiveZeroLagVelocity" => Some(Self::JurikAdaptiveZeroLagVelocity),
            "jurikCommodityChannelIndex" => Some(Self::JurikCommodityChannelIndex),
            "jurikCompositeFractalBehaviorIndex" => Some(Self::JurikCompositeFractalBehaviorIndex),
            "jurikDirectionalMovementIndex" => Some(Self::JurikDirectionalMovementIndex),
            "jurikFractalAdaptiveZeroLagVelocity" => {
                Some(Self::JurikFractalAdaptiveZeroLagVelocity)
            }
            "jurikMovingAverage" => Some(Self::JurikMovingAverage),
            "jurikRelativeTrendStrengthIndex" => Some(Self::JurikRelativeTrendStrengthIndex),
            "jurikTurningPointOscillator" => Some(Self::JurikTurningPointOscillator),
            "jurikWaveletSampler" => Some(Self::JurikWaveletSampler),
            "jurikZeroLagVelocity" => Some(Self::JurikZeroLagVelocity),
            "doubleExponentialMovingAverage" => Some(Self::DoubleExponentialMovingAverage),
            "tripleExponentialMovingAverage" => Some(Self::TripleExponentialMovingAverage),
            "kaufmanAdaptiveMovingAverage" => Some(Self::KaufmanAdaptiveMovingAverage),
            "t2ExponentialMovingAverage" => Some(Self::T2ExponentialMovingAverage),
            "t3ExponentialMovingAverage" => Some(Self::T3ExponentialMovingAverage),
            "aroon" => Some(Self::Aroon),
            "chandeMomentumOscillator" => Some(Self::ChandeMomentumOscillator),
            "stochasticRelativeStrengthIndex" => Some(Self::StochasticRelativeStrengthIndex),
            "adaptiveTrendAndCycleFilter" => Some(Self::AdaptiveTrendAndCycleFilter),
            "averageDirectionalMovementIndex" => Some(Self::AverageDirectionalMovementIndex),
            "averageDirectionalMovementIndexRating" => {
                Some(Self::AverageDirectionalMovementIndexRating)
            }
            "averageTrueRange" => Some(Self::AverageTrueRange),
            "directionalIndicatorMinus" => Some(Self::DirectionalIndicatorMinus),
            "directionalIndicatorPlus" => Some(Self::DirectionalIndicatorPlus),
            "directionalMovementIndex" => Some(Self::DirectionalMovementIndex),
            "directionalMovementMinus" => Some(Self::DirectionalMovementMinus),
            "directionalMovementPlus" => Some(Self::DirectionalMovementPlus),
            "normalizedAverageTrueRange" => Some(Self::NormalizedAverageTrueRange),
            "parabolicStopAndReverse" => Some(Self::ParabolicStopAndReverse),
            "relativeStrengthIndex" => Some(Self::RelativeStrengthIndex),
            "trueRange" => Some(Self::TrueRange),
            "goertzelSpectrum" => Some(Self::GoertzelSpectrum),
            "maximumEntropySpectrum" => Some(Self::MaximumEntropySpectrum),
            _ => None,
        }
    }
}

impl std::fmt::Display for Identifier {
    fn fmt(&self, f: &mut std::fmt::Formatter<'_>) -> std::fmt::Result {
        f.write_str(self.as_str())
    }
}
