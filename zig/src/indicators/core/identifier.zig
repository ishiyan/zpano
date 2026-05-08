/// Identifies an indicator by enumerating all implemented indicators.
/// Values are 0-based.
pub const Identifier = enum(u8) {

    // ── common ────────────────────────────────────────────────────────────
    /// Identifies the Absolute Price Oscillator (APO) indicator.
    absolute_price_oscillator = 0,
    /// Identifies the Exponential Moving Average (EMA) indicator.
    exponential_moving_average = 1,
    /// Identifies the Linear Regression (LINEARREG) indicator.
    linear_regression = 2,
    /// Identifies the Momentum (MOM) indicator.
    momentum = 3,
    /// Identifies the Pearson's Correlation Coefficient (CORREL) indicator.
    pearsons_correlation_coefficient = 4,
    /// Identifies the Rate of Change (ROC) indicator.
    rate_of_change = 5,
    /// Identifies the Rate of Change Percent (ROCP) indicator.
    rate_of_change_percent = 6,
    /// Identifies the Rate of Change Ratio (ROCR / ROCR100) indicator.
    rate_of_change_ratio = 7,
    /// Identifies the Simple Moving Average (SMA) indicator.
    simple_moving_average = 8,
    /// Identifies the Standard Deviation (STDEV) indicator.
    standard_deviation = 9,
    /// Identifies the Triangular Moving Average (TRIMA) indicator.
    triangular_moving_average = 10,
    /// Identifies the Variance (VAR) indicator.
    variance = 11,
    /// Identifies the Weighted Moving Average (WMA) indicator.
    weighted_moving_average = 12,

    // ── arnaudlegoux ──────────────────────────────────────────────────────
    /// Identifies the Arnaud Legoux Moving Average (ALMA) indicator.
    arnaud_legoux_moving_average = 13,

    // ── donaldlambert ─────────────────────────────────────────────────────
    /// Identifies the Donald Lambert Commodity Channel Index (CCI) indicator.
    commodity_channel_index = 14,

    // ── genequong ─────────────────────────────────────────────────────────
    /// Identifies the Gene Quong Money Flow Index (MFI) indicator.
    money_flow_index = 15,

    // ── georgelane ────────────────────────────────────────────────────────
    /// Identifies the George Lane Stochastic Oscillator (STOCH) indicator.
    stochastic = 16,

    // ── geraldappel ───────────────────────────────────────────────────────
    /// Identifies Gerald Appel's Moving Average Convergence Divergence (MACD) indicator.
    moving_average_convergence_divergence = 17,
    /// Identifies the Gerald Appel Percentage Price Oscillator (PPO) indicator.
    percentage_price_oscillator = 18,

    // ── igorlivshin ───────────────────────────────────────────────────────
    /// Identifies the Igor Livshin Balance of Power (BOP) indicator.
    balance_of_power = 19,

    // ── jackhutson ────────────────────────────────────────────────────────
    /// Identifies Jack Hutson's Triple Exponential Moving Average Oscillator (TRIX) indicator.
    triple_exponential_moving_average_oscillator = 20,

    // ── johnbollinger ─────────────────────────────────────────────────────
    /// Identifies the Bollinger Bands (BB) indicator.
    bollinger_bands = 21,
    /// Identifies John Bollinger's Bollinger Bands Trend (BBTrend) indicator.
    bollinger_bands_trend = 22,

    // ── johnehlers ────────────────────────────────────────────────────────
    /// Identifies the Ehlers Autocorrelation Indicator (ACI) heatmap, a bank of Pearson
    /// correlation coefficients between the current filtered series and a lagged copy
    /// of itself, following EasyLanguage listing 8-2.
    auto_correlation_indicator = 23,
    /// Identifies the Ehlers Autocorrelation Periodogram (ACP) heatmap, a discrete Fourier
    /// transform of the autocorrelation function over a configurable cycle-period range,
    /// following EasyLanguage listing 8-3.
    auto_correlation_periodogram = 24,
    /// Identifies the Ehlers Center of Gravity Oscillator (COG) indicator.
    center_of_gravity_oscillator = 25,
    /// Identifies the Ehlers Comb Band-Pass Spectrum (CBPS) heatmap indicator, a bank of
    /// 2-pole band-pass filters (one per cycle period) fed by a Butterworth highpass +
    /// Super Smoother pre-filter cascade, following EasyLanguage listing 10-1.
    comb_band_pass_spectrum = 26,
    /// Identifies the Ehlers Corona Signal To Noise Ratio (CSNR) heatmap indicator, exposing
    /// the intensity raster heatmap column and the current SNR mapped into the parameter range.
    corona_signal_to_noise_ratio = 27,
    /// Identifies the Ehlers Corona Spectrum (CSPECT) heatmap indicator, exposing the dB
    /// heatmap column, the weighted dominant cycle estimate and its 5-sample median.
    corona_spectrum = 28,
    /// Identifies the Ehlers Corona Swing Position (CSWING) heatmap indicator, exposing the
    /// intensity raster heatmap column and the current swing position mapped into the parameter range.
    corona_swing_position = 29,
    /// Identifies the Ehlers Corona Trend Vigor (CTV) heatmap indicator, exposing the intensity
    /// raster heatmap column and the current trend vigor scaled into the parameter range.
    corona_trend_vigor = 30,
    /// Identifies the Ehlers Cyber Cycle (CC) indicator.
    cyber_cycle = 31,
    /// Identifies the Ehlers Discrete Fourier Transform Spectrum (DFTPS) heatmap indicator,
    /// a mean-subtracted DFT power spectrum over a configurable cycle-period range.
    discrete_fourier_transform_spectrum = 32,
    /// Identifies the Ehlers Dominant Cycle (DC) indicator, exposing raw period, smoothed period and phase.
    dominant_cycle = 33,
    /// Identifies the Ehlers Fractal Adaptive Moving Average (FRAMA) indicator.
    fractal_adaptive_moving_average = 34,
    /// Identifies the Ehlers Hilbert Transformer Instantaneous Trend Line (HTITL) indicator,
    /// exposing trend value and dominant cycle period.
    hilbert_transformer_instantaneous_trend_line = 35,
    /// Identifies the Ehlers Instantaneous Trend Line (iTrend) indicator.
    instantaneous_trend_line = 36,
    /// Identifies the Ehlers MESA Adaptive Moving Average (MAMA) indicator.
    mesa_adaptive_moving_average = 37,
    /// Identifies the Ehlers Roofing Filter indicator.
    roofing_filter = 38,
    /// Identifies the Ehlers Sine Wave (SW) indicator, exposing sine value, lead sine,
    /// band, dominant cycle period and phase.
    sine_wave = 39,
    /// Identifies the Ehlers Super Smoother (SS) indicator.
    super_smoother = 40,
    /// Identifies the Ehlers Trend / Cycle Mode (TCM) indicator, exposing the trend/cycle
    /// value (+1 in trend, -1 in cycle), trend/cycle mode flags, instantaneous trend line,
    /// sine wave, lead sine wave, dominant cycle period and phase.
    trend_cycle_mode = 41,
    /// Identifies the Ehlers Zero-lag Error-Correcting Exponential Moving Average (ZECEMA) indicator.
    zero_lag_error_correcting_exponential_moving_average = 42,
    /// Identifies the Ehlers Zero-lag Exponential Moving Average (ZEMA) indicator.
    zero_lag_exponential_moving_average = 43,

    // ── josephgranville ───────────────────────────────────────────────────
    /// Identifies the Joseph Granville On-Balance Volume (OBV) indicator.
    on_balance_volume = 44,

    // ── larrywilliams ─────────────────────────────────────────────────────
    /// Identifies the Larry Williams Ultimate Oscillator (ULTOSC) indicator.
    ultimate_oscillator = 45,
    /// Identifies the Larry Williams Williams %R (WILL%R) indicator.
    williams_percent_r = 46,

    // ── manfreddurschner ──────────────────────────────────────────────────
    /// Identifies the New Moving Average (NMA) indicator by Durschner.
    new_moving_average = 47,

    // ── marcchaikin ───────────────────────────────────────────────────────
    /// Identifies the Marc Chaikin Advance-Decline (A/D) indicator.
    advance_decline = 48,
    /// Identifies the Marc Chaikin Advance-Decline Oscillator (ADOSC) indicator.
    advance_decline_oscillator = 49,

    // ── markjurik ─────────────────────────────────────────────────────────
    /// Identifies the Jurik Adaptive Relative Trend Strength Index (JARSX) indicator.
    jurik_adaptive_relative_trend_strength_index = 50,
    /// Identifies the Jurik Adaptive Zero Lag Velocity (JAVEL) indicator.
    jurik_adaptive_zero_lag_velocity = 51,
    /// Identifies the Jurik Commodity Channel Index (JCCX) indicator.
    jurik_commodity_channel_index = 52,
    /// Identifies the Jurik Composite Fractal Behavior Index (CFB) indicator.
    jurik_composite_fractal_behavior_index = 53,
    /// Identifies the Jurik Directional Movement Index (DMX) indicator.
    jurik_directional_movement_index = 54,
    /// Identifies the Jurik Fractal Adaptive Zero Lag Velocity (JVELCFB) indicator.
    jurik_fractal_adaptive_zero_lag_velocity = 55,
    /// Identifies the Jurik Moving Average (JMA) indicator.
    jurik_moving_average = 56,
    /// Identifies the Jurik Relative Trend Strength Index (RSX) indicator.
    jurik_relative_trend_strength_index = 57,
    /// Identifies the Jurik Turning Point Oscillator (JTPO) indicator.
    jurik_turning_point_oscillator = 58,
    /// Identifies the Jurik Wavelet Sampler (WAV) indicator.
    jurik_wavelet_sampler = 59,
    /// Identifies the Jurik Zero Lag Velocity (VEL) indicator.
    jurik_zero_lag_velocity = 60,

    // ── patrickmulloy ─────────────────────────────────────────────────────
    /// Identifies the Double Exponential Moving Average (DEMA) indicator.
    double_exponential_moving_average = 61,
    /// Identifies the Triple Exponential Moving Average (TEMA) indicator.
    triple_exponential_moving_average = 62,

    // ── perrykaufman ──────────────────────────────────────────────────────
    /// Identifies the Kaufman Adaptive Moving Average (KAMA) indicator.
    kaufman_adaptive_moving_average = 63,

    // ── timtillson ────────────────────────────────────────────────────────
    /// Identifies the T2 Exponential Moving Average (T2EMA) indicator.
    t2_exponential_moving_average = 64,
    /// Identifies the T3 Exponential Moving Average (T3EMA) indicator.
    t3_exponential_moving_average = 65,

    // ── tusharchande ──────────────────────────────────────────────────────
    /// Identifies the Tushar Chande Aroon (AROON) indicator.
    aroon = 66,
    /// Identifies the Chande Momentum Oscillator (CMO) indicator.
    chande_momentum_oscillator = 67,
    /// Identifies the Tushar Chande Stochastic RSI (STOCHRSI) indicator.
    stochastic_relative_strength_index = 68,

    // ── vladimirkravchuk ──────────────────────────────────────────────────
    /// Identifies Vladimir Kravchuk's Adaptive Trend and Cycle Filter (ATCF) suite: a bank
    /// of five FIR filters (FATL, SATL, RFTL, RSTL, RBCI) plus three composites (FTLM, STLM, PCCI).
    adaptive_trend_and_cycle_filter = 69,

    // ── welleswilder ──────────────────────────────────────────────────────
    /// Identifies the Welles Wilder Average Directional Movement Index (ADX) indicator.
    average_directional_movement_index = 70,
    /// Identifies the Welles Wilder Average Directional Movement Index Rating (ADXR) indicator.
    average_directional_movement_index_rating = 71,
    /// Identifies the Welles Wilder Average True Range (ATR) indicator.
    average_true_range = 72,
    /// Identifies the Welles Wilder Directional Indicator Minus (-DI) indicator.
    directional_indicator_minus = 73,
    /// Identifies the Welles Wilder Directional Indicator Plus (+DI) indicator.
    directional_indicator_plus = 74,
    /// Identifies the Welles Wilder Directional Movement Index (DX) indicator.
    directional_movement_index = 75,
    /// Identifies the Welles Wilder Directional Movement Minus (-DM) indicator.
    directional_movement_minus = 76,
    /// Identifies the Welles Wilder Directional Movement Plus (+DM) indicator.
    directional_movement_plus = 77,
    /// Identifies the Welles Wilder Normalized Average True Range (NATR) indicator.
    normalized_average_true_range = 78,
    /// Identifies the Welles Wilder Parabolic Stop And Reverse (SAR) indicator.
    parabolic_stop_and_reverse = 79,
    /// Identifies the Relative Strength Index (RSI) indicator.
    relative_strength_index = 80,
    /// Identifies the Welles Wilder True Range (TR) indicator.
    true_range = 81,

    // ── custom ────────────────────────────────────────────────────────────
    /// Identifies the Goertzel power spectrum (GOERTZEL) indicator.
    goertzel_spectrum = 82,
    /// Identifies the Maximum Entropy Spectrum (MESPECT) heatmap indicator, a Burg
    /// maximum-entropy auto-regressive power spectrum over a configurable cycle-period range.
    maximum_entropy_spectrum = 83,

    /// Returns the camelCase string representation matching Go's String().
    pub fn asStr(self: Identifier) []const u8 {
        return switch (self) {

            // ── common ────────────────────────────────────────────────────────────
            .absolute_price_oscillator => "absolutePriceOscillator",
            .exponential_moving_average => "exponentialMovingAverage",
            .linear_regression => "linearRegression",
            .momentum => "momentum",
            .pearsons_correlation_coefficient => "pearsonsCorrelationCoefficient",
            .rate_of_change => "rateOfChange",
            .rate_of_change_percent => "rateOfChangePercent",
            .rate_of_change_ratio => "rateOfChangeRatio",
            .simple_moving_average => "simpleMovingAverage",
            .standard_deviation => "standardDeviation",
            .triangular_moving_average => "triangularMovingAverage",
            .variance => "variance",
            .weighted_moving_average => "weightedMovingAverage",

            // ── arnaudlegoux ──────────────────────────────────────────────────────
            .arnaud_legoux_moving_average => "arnaudLegouxMovingAverage",

            // ── donaldlambert ─────────────────────────────────────────────────────
            .commodity_channel_index => "commodityChannelIndex",

            // ── genequong ─────────────────────────────────────────────────────────
            .money_flow_index => "moneyFlowIndex",

            // ── georgelane ────────────────────────────────────────────────────────
            .stochastic => "stochastic",

            // ── geraldappel ───────────────────────────────────────────────────────
            .moving_average_convergence_divergence => "movingAverageConvergenceDivergence",
            .percentage_price_oscillator => "percentagePriceOscillator",

            // ── igorlivshin ───────────────────────────────────────────────────────
            .balance_of_power => "balanceOfPower",

            // ── jackhutson ────────────────────────────────────────────────────────
            .triple_exponential_moving_average_oscillator => "tripleExponentialMovingAverageOscillator",

            // ── johnbollinger ─────────────────────────────────────────────────────
            .bollinger_bands => "bollingerBands",
            .bollinger_bands_trend => "bollingerBandsTrend",

            // ── johnehlers ────────────────────────────────────────────────────────
            .auto_correlation_indicator => "autoCorrelationIndicator",
            .auto_correlation_periodogram => "autoCorrelationPeriodogram",
            .center_of_gravity_oscillator => "centerOfGravityOscillator",
            .comb_band_pass_spectrum => "combBandPassSpectrum",
            .corona_signal_to_noise_ratio => "coronaSignalToNoiseRatio",
            .corona_spectrum => "coronaSpectrum",
            .corona_swing_position => "coronaSwingPosition",
            .corona_trend_vigor => "coronaTrendVigor",
            .cyber_cycle => "cyberCycle",
            .discrete_fourier_transform_spectrum => "discreteFourierTransformSpectrum",
            .dominant_cycle => "dominantCycle",
            .fractal_adaptive_moving_average => "fractalAdaptiveMovingAverage",
            .hilbert_transformer_instantaneous_trend_line => "hilbertTransformerInstantaneousTrendLine",
            .instantaneous_trend_line => "instantaneousTrendLine",
            .mesa_adaptive_moving_average => "mesaAdaptiveMovingAverage",
            .roofing_filter => "roofingFilter",
            .sine_wave => "sineWave",
            .super_smoother => "superSmoother",
            .trend_cycle_mode => "trendCycleMode",
            .zero_lag_error_correcting_exponential_moving_average => "zeroLagErrorCorrectingExponentialMovingAverage",
            .zero_lag_exponential_moving_average => "zeroLagExponentialMovingAverage",

            // ── josephgranville ───────────────────────────────────────────────────
            .on_balance_volume => "onBalanceVolume",

            // ── larrywilliams ─────────────────────────────────────────────────────
            .ultimate_oscillator => "ultimateOscillator",
            .williams_percent_r => "williamsPercentR",

            // ── manfreddurschner ──────────────────────────────────────────────────
            .new_moving_average => "newMovingAverage",

            // ── marcchaikin ───────────────────────────────────────────────────────
            .advance_decline => "advanceDecline",
            .advance_decline_oscillator => "advanceDeclineOscillator",

            // ── markjurik ─────────────────────────────────────────────────────────
            .jurik_adaptive_relative_trend_strength_index => "jurikAdaptiveRelativeTrendStrengthIndex",
            .jurik_adaptive_zero_lag_velocity => "jurikAdaptiveZeroLagVelocity",
            .jurik_commodity_channel_index => "jurikCommodityChannelIndex",
            .jurik_composite_fractal_behavior_index => "jurikCompositeFractalBehaviorIndex",
            .jurik_directional_movement_index => "jurikDirectionalMovementIndex",
            .jurik_fractal_adaptive_zero_lag_velocity => "jurikFractalAdaptiveZeroLagVelocity",
            .jurik_moving_average => "jurikMovingAverage",
            .jurik_relative_trend_strength_index => "jurikRelativeTrendStrengthIndex",
            .jurik_turning_point_oscillator => "jurikTurningPointOscillator",
            .jurik_wavelet_sampler => "jurikWaveletSampler",
            .jurik_zero_lag_velocity => "jurikZeroLagVelocity",

            // ── patrickmulloy ─────────────────────────────────────────────────────
            .double_exponential_moving_average => "doubleExponentialMovingAverage",
            .triple_exponential_moving_average => "tripleExponentialMovingAverage",

            // ── perrykaufman ──────────────────────────────────────────────────────
            .kaufman_adaptive_moving_average => "kaufmanAdaptiveMovingAverage",

            // ── timtillson ────────────────────────────────────────────────────────
            .t2_exponential_moving_average => "t2ExponentialMovingAverage",
            .t3_exponential_moving_average => "t3ExponentialMovingAverage",

            // ── tusharchande ──────────────────────────────────────────────────────
            .aroon => "aroon",
            .chande_momentum_oscillator => "chandeMomentumOscillator",
            .stochastic_relative_strength_index => "stochasticRelativeStrengthIndex",

            // ── vladimirkravchuk ──────────────────────────────────────────────────
            .adaptive_trend_and_cycle_filter => "adaptiveTrendAndCycleFilter",

            // ── welleswilder ──────────────────────────────────────────────────────
            .average_directional_movement_index => "averageDirectionalMovementIndex",
            .average_directional_movement_index_rating => "averageDirectionalMovementIndexRating",
            .average_true_range => "averageTrueRange",
            .directional_indicator_minus => "directionalIndicatorMinus",
            .directional_indicator_plus => "directionalIndicatorPlus",
            .directional_movement_index => "directionalMovementIndex",
            .directional_movement_minus => "directionalMovementMinus",
            .directional_movement_plus => "directionalMovementPlus",
            .normalized_average_true_range => "normalizedAverageTrueRange",
            .parabolic_stop_and_reverse => "parabolicStopAndReverse",
            .relative_strength_index => "relativeStrengthIndex",
            .true_range => "trueRange",

            // ── custom ────────────────────────────────────────────────────────────
            .goertzel_spectrum => "goertzelSpectrum",
            .maximum_entropy_spectrum => "maximumEntropySpectrum",
        };
    }

    /// Parses a camelCase string into an Identifier.
    pub fn fromStr(s: []const u8) ?Identifier {
        const map = .{

            // ── common ────────────────────────────────────────────────────────────
            .{ "absolutePriceOscillator", Identifier.absolute_price_oscillator },
            .{ "exponentialMovingAverage", Identifier.exponential_moving_average },
            .{ "linearRegression", Identifier.linear_regression },
            .{ "momentum", Identifier.momentum },
            .{ "pearsonsCorrelationCoefficient", Identifier.pearsons_correlation_coefficient },
            .{ "rateOfChange", Identifier.rate_of_change },
            .{ "rateOfChangePercent", Identifier.rate_of_change_percent },
            .{ "rateOfChangeRatio", Identifier.rate_of_change_ratio },
            .{ "simpleMovingAverage", Identifier.simple_moving_average },
            .{ "standardDeviation", Identifier.standard_deviation },
            .{ "triangularMovingAverage", Identifier.triangular_moving_average },
            .{ "variance", Identifier.variance },
            .{ "weightedMovingAverage", Identifier.weighted_moving_average },

            // ── arnaudlegoux ──────────────────────────────────────────────────────
            .{ "arnaudLegouxMovingAverage", Identifier.arnaud_legoux_moving_average },

            // ── donaldlambert ─────────────────────────────────────────────────────
            .{ "commodityChannelIndex", Identifier.commodity_channel_index },

            // ── genequong ─────────────────────────────────────────────────────────
            .{ "moneyFlowIndex", Identifier.money_flow_index },

            // ── georgelane ────────────────────────────────────────────────────────
            .{ "stochastic", Identifier.stochastic },

            // ── geraldappel ───────────────────────────────────────────────────────
            .{ "movingAverageConvergenceDivergence", Identifier.moving_average_convergence_divergence },
            .{ "percentagePriceOscillator", Identifier.percentage_price_oscillator },

            // ── igorlivshin ───────────────────────────────────────────────────────
            .{ "balanceOfPower", Identifier.balance_of_power },

            // ── jackhutson ────────────────────────────────────────────────────────
            .{ "tripleExponentialMovingAverageOscillator", Identifier.triple_exponential_moving_average_oscillator },

            // ── johnbollinger ─────────────────────────────────────────────────────
            .{ "bollingerBands", Identifier.bollinger_bands },
            .{ "bollingerBandsTrend", Identifier.bollinger_bands_trend },

            // ── johnehlers ────────────────────────────────────────────────────────
            .{ "autoCorrelationIndicator", Identifier.auto_correlation_indicator },
            .{ "autoCorrelationPeriodogram", Identifier.auto_correlation_periodogram },
            .{ "centerOfGravityOscillator", Identifier.center_of_gravity_oscillator },
            .{ "combBandPassSpectrum", Identifier.comb_band_pass_spectrum },
            .{ "coronaSignalToNoiseRatio", Identifier.corona_signal_to_noise_ratio },
            .{ "coronaSpectrum", Identifier.corona_spectrum },
            .{ "coronaSwingPosition", Identifier.corona_swing_position },
            .{ "coronaTrendVigor", Identifier.corona_trend_vigor },
            .{ "cyberCycle", Identifier.cyber_cycle },
            .{ "discreteFourierTransformSpectrum", Identifier.discrete_fourier_transform_spectrum },
            .{ "dominantCycle", Identifier.dominant_cycle },
            .{ "fractalAdaptiveMovingAverage", Identifier.fractal_adaptive_moving_average },
            .{ "hilbertTransformerInstantaneousTrendLine", Identifier.hilbert_transformer_instantaneous_trend_line },
            .{ "instantaneousTrendLine", Identifier.instantaneous_trend_line },
            .{ "mesaAdaptiveMovingAverage", Identifier.mesa_adaptive_moving_average },
            .{ "roofingFilter", Identifier.roofing_filter },
            .{ "sineWave", Identifier.sine_wave },
            .{ "superSmoother", Identifier.super_smoother },
            .{ "trendCycleMode", Identifier.trend_cycle_mode },
            .{ "zeroLagErrorCorrectingExponentialMovingAverage", Identifier.zero_lag_error_correcting_exponential_moving_average },
            .{ "zeroLagExponentialMovingAverage", Identifier.zero_lag_exponential_moving_average },

            // ── josephgranville ───────────────────────────────────────────────────
            .{ "onBalanceVolume", Identifier.on_balance_volume },

            // ── larrywilliams ─────────────────────────────────────────────────────
            .{ "ultimateOscillator", Identifier.ultimate_oscillator },
            .{ "williamsPercentR", Identifier.williams_percent_r },

            // ── manfreddurschner ──────────────────────────────────────────────────
            .{ "newMovingAverage", Identifier.new_moving_average },

            // ── marcchaikin ───────────────────────────────────────────────────────
            .{ "advanceDecline", Identifier.advance_decline },
            .{ "advanceDeclineOscillator", Identifier.advance_decline_oscillator },

            // ── markjurik ─────────────────────────────────────────────────────────
            .{ "jurikAdaptiveRelativeTrendStrengthIndex", Identifier.jurik_adaptive_relative_trend_strength_index },
            .{ "jurikAdaptiveZeroLagVelocity", Identifier.jurik_adaptive_zero_lag_velocity },
            .{ "jurikCommodityChannelIndex", Identifier.jurik_commodity_channel_index },
            .{ "jurikCompositeFractalBehaviorIndex", Identifier.jurik_composite_fractal_behavior_index },
            .{ "jurikDirectionalMovementIndex", Identifier.jurik_directional_movement_index },
            .{ "jurikFractalAdaptiveZeroLagVelocity", Identifier.jurik_fractal_adaptive_zero_lag_velocity },
            .{ "jurikMovingAverage", Identifier.jurik_moving_average },
            .{ "jurikRelativeTrendStrengthIndex", Identifier.jurik_relative_trend_strength_index },
            .{ "jurikTurningPointOscillator", Identifier.jurik_turning_point_oscillator },
            .{ "jurikWaveletSampler", Identifier.jurik_wavelet_sampler },
            .{ "jurikZeroLagVelocity", Identifier.jurik_zero_lag_velocity },

            // ── patrickmulloy ─────────────────────────────────────────────────────
            .{ "doubleExponentialMovingAverage", Identifier.double_exponential_moving_average },
            .{ "tripleExponentialMovingAverage", Identifier.triple_exponential_moving_average },

            // ── perrykaufman ──────────────────────────────────────────────────────
            .{ "kaufmanAdaptiveMovingAverage", Identifier.kaufman_adaptive_moving_average },

            // ── timtillson ────────────────────────────────────────────────────────
            .{ "t2ExponentialMovingAverage", Identifier.t2_exponential_moving_average },
            .{ "t3ExponentialMovingAverage", Identifier.t3_exponential_moving_average },

            // ── tusharchande ──────────────────────────────────────────────────────
            .{ "aroon", Identifier.aroon },
            .{ "chandeMomentumOscillator", Identifier.chande_momentum_oscillator },
            .{ "stochasticRelativeStrengthIndex", Identifier.stochastic_relative_strength_index },

            // ── vladimirkravchuk ──────────────────────────────────────────────────
            .{ "adaptiveTrendAndCycleFilter", Identifier.adaptive_trend_and_cycle_filter },

            // ── welleswilder ──────────────────────────────────────────────────────
            .{ "averageDirectionalMovementIndex", Identifier.average_directional_movement_index },
            .{ "averageDirectionalMovementIndexRating", Identifier.average_directional_movement_index_rating },
            .{ "averageTrueRange", Identifier.average_true_range },
            .{ "directionalIndicatorMinus", Identifier.directional_indicator_minus },
            .{ "directionalIndicatorPlus", Identifier.directional_indicator_plus },
            .{ "directionalMovementIndex", Identifier.directional_movement_index },
            .{ "directionalMovementMinus", Identifier.directional_movement_minus },
            .{ "directionalMovementPlus", Identifier.directional_movement_plus },
            .{ "normalizedAverageTrueRange", Identifier.normalized_average_true_range },
            .{ "parabolicStopAndReverse", Identifier.parabolic_stop_and_reverse },
            .{ "relativeStrengthIndex", Identifier.relative_strength_index },
            .{ "trueRange", Identifier.true_range },

            // ── custom ────────────────────────────────────────────────────────────
            .{ "goertzelSpectrum", Identifier.goertzel_spectrum },
            .{ "maximumEntropySpectrum", Identifier.maximum_entropy_spectrum },
        };

        inline for (map) |entry| {
            if (std.mem.eql(u8, s, entry[0])) return entry[1];
        }
        return null;
    }
};

const std = @import("std");
const testing = std.testing;

test "identifier asStr round-trip" {
    // Test a few representative identifiers.
    try testing.expectEqualStrings("simpleMovingAverage", Identifier.simple_moving_average.asStr());
    try testing.expectEqualStrings("bollingerBands", Identifier.bollinger_bands.asStr());
    try testing.expectEqualStrings("autoCorrelationPeriodogram", Identifier.auto_correlation_periodogram.asStr());
}

test "identifier fromStr round-trip" {
    try testing.expectEqual(Identifier.simple_moving_average, Identifier.fromStr("simpleMovingAverage").?);
    try testing.expectEqual(Identifier.bollinger_bands, Identifier.fromStr("bollingerBands").?);
    try testing.expectEqual(@as(?Identifier, null), Identifier.fromStr("nonExistent"));
}

test "identifier fromStr all variants" {
    // Verify every variant round-trips through asStr/fromStr.
    inline for (std.meta.fields(Identifier)) |field| {
        const id: Identifier = @enumFromInt(field.value);
        const str = id.asStr();
        const parsed = Identifier.fromStr(str);
        try testing.expectEqual(id, parsed.?);
    }
}
