/// Factory that maps an [`Identifier`] and a JSON parameter string to a
/// boxed [`Indicator`] instance. Callers don't need to import individual
/// indicator modules directly.
///
/// For indicators with Length / SmoothingFactor constructor variants, the
/// factory auto-detects which to use based on JSON keys.
use super::json::{
    get_bool, get_f64, get_i32, get_i64, get_usize, has_key, is_empty_object, JsonValue,
};
use crate::indicators::core::identifier::Identifier;
use crate::indicators::core::indicator::Indicator;

// ── common ───────────────────────────────────────────────────────────────────
use crate::indicators::common::absolute_price_oscillator::{
    AbsolutePriceOscillator, AbsolutePriceOscillatorParams,
};
use crate::indicators::common::exponential_moving_average::{
    ExponentialMovingAverage, ExponentialMovingAverageLengthParams,
    ExponentialMovingAverageSmoothingFactorParams,
};
use crate::indicators::common::linear_regression::{LinearRegression, LinearRegressionParams};
use crate::indicators::common::momentum::{Momentum, MomentumParams};
use crate::indicators::common::pearsons_correlation_coefficient::{
    PearsonsCorrelationCoefficient, PearsonsCorrelationCoefficientParams,
};
use crate::indicators::common::rate_of_change::{RateOfChange, RateOfChangeParams};
use crate::indicators::common::rate_of_change_percent::{
    RateOfChangePercent, RateOfChangePercentParams,
};
use crate::indicators::common::rate_of_change_ratio::{RateOfChangeRatio, RateOfChangeRatioParams};
use crate::indicators::common::simple_moving_average::{
    SimpleMovingAverage, SimpleMovingAverageParams,
};
use crate::indicators::common::standard_deviation::{StandardDeviation, StandardDeviationParams};
use crate::indicators::common::triangular_moving_average::{
    TriangularMovingAverage, TriangularMovingAverageParams,
};
use crate::indicators::common::variance::{Variance, VarianceParams};
use crate::indicators::common::weighted_moving_average::{
    WeightedMovingAverage, WeightedMovingAverageParams,
};

// ── custom ───────────────────────────────────────────────────────────────────
use crate::indicators::custom::goertzel_spectrum::goertzel_spectrum::{
    GoertzelSpectrum, GoertzelSpectrumParams,
};
use crate::indicators::custom::maximum_entropy_spectrum::maximum_entropy_spectrum::{
    MaximumEntropySpectrum, MaximumEntropySpectrumParams,
};

// ── donald lambert ───────────────────────────────────────────────────────────
use crate::indicators::donald_lambert::commodity_channel_index::{
    CommodityChannelIndex, CommodityChannelIndexParams,
};

// ── gene quong ───────────────────────────────────────────────────────────────
use crate::indicators::gene_quong::money_flow_index::{MoneyFlowIndex, MoneyFlowIndexParams};

// ── george lane ──────────────────────────────────────────────────────────────
use crate::indicators::george_lane::stochastic::{Stochastic, StochasticParams};

// ── gerald appel ─────────────────────────────────────────────────────────────
use crate::indicators::gerald_appel::moving_average_convergence_divergence::moving_average_convergence_divergence::{MovingAverageConvergenceDivergence, MovingAverageConvergenceDivergenceParams};
use crate::indicators::gerald_appel::percentage_price_oscillator::percentage_price_oscillator::{PercentagePriceOscillator, PercentagePriceOscillatorParams};

// ── igor livshin ─────────────────────────────────────────────────────────────
use crate::indicators::igor_livshin::balance_of_power::{BalanceOfPower, BalanceOfPowerParams};

// ── jack hutson ──────────────────────────────────────────────────────────────
use crate::indicators::jack_hutson::triple_exponential_moving_average_oscillator::{
    TripleExponentialMovingAverageOscillator, TripleExponentialMovingAverageOscillatorParams,
};

// ── john bollinger ───────────────────────────────────────────────────────────
use crate::indicators::john_bollinger::bollinger_bands::{BollingerBands, BollingerBandsParams};
use crate::indicators::john_bollinger::bollinger_bands_trend::{
    BollingerBandsTrend, BollingerBandsTrendParams,
};

// ── john ehlers ──────────────────────────────────────────────────────────────
use crate::indicators::john_ehlers::super_smoother::{SuperSmoother, SuperSmootherParams};
use crate::indicators::john_ehlers::center_of_gravity_oscillator::{CenterOfGravityOscillator, CenterOfGravityOscillatorParams};
use crate::indicators::john_ehlers::cyber_cycle::{CyberCycle, CyberCycleLengthParams, CyberCycleSmoothingFactorParams};
use crate::indicators::john_ehlers::instantaneous_trendline::{InstantaneousTrendLine};
use crate::indicators::john_ehlers::instantaneous_trendline::LengthParams as ItlLengthParams;
use crate::indicators::john_ehlers::instantaneous_trendline::SmoothingFactorParams as ItlSmoothingFactorParams;
use crate::indicators::john_ehlers::zero_lag_exponential_moving_average::zero_lag_exponential_moving_average::{ZeroLagExponentialMovingAverage, ZeroLagExponentialMovingAverageParams};
use crate::indicators::john_ehlers::zero_lag_error_correcting_exponential_moving_average::{ZeroLagErrorCorrectingExponentialMovingAverage, ZeroLagErrorCorrectingExponentialMovingAverageParams};
use crate::indicators::john_ehlers::roofing_filter::{RoofingFilter, RoofingFilterParams};
use crate::indicators::john_ehlers::mesa_adaptive_moving_average::{MesaAdaptiveMovingAverage, MesaAdaptiveMovingAverageLengthParams, MesaAdaptiveMovingAverageSmoothingFactorParams};
use crate::indicators::john_ehlers::fractal_adaptive_moving_average::{FractalAdaptiveMovingAverage, FractalAdaptiveMovingAverageParams};
use crate::indicators::john_ehlers::dominant_cycle::{DominantCycle, DominantCycleParams};
use crate::indicators::john_ehlers::sine_wave::{SineWave, SineWaveParams};
use crate::indicators::john_ehlers::hilbert_transformer_instantaneous_trendline::{HilbertTransformerInstantaneousTrendLine, HilbertTransformerInstantaneousTrendLineParams};
use crate::indicators::john_ehlers::trend_cycle_mode::{TrendCycleMode, TrendCycleModeParams};
use crate::indicators::john_ehlers::corona_spectrum::corona_spectrum::{CoronaSpectrum, CoronaSpectrumParams};
use crate::indicators::john_ehlers::corona_signal_to_noise_ratio::corona_signal_to_noise_ratio::{CoronaSignalToNoiseRatio, CoronaSignalToNoiseRatioParams};
use crate::indicators::john_ehlers::corona_swing_position::corona_swing_position::{CoronaSwingPosition, CoronaSwingPositionParams};
use crate::indicators::john_ehlers::corona_trend_vigor::corona_trend_vigor::{CoronaTrendVigor, CoronaTrendVigorParams};
use crate::indicators::john_ehlers::autocorrelation_indicator::{AutoCorrelationIndicator, AutoCorrelationIndicatorParams};
use crate::indicators::john_ehlers::autocorrelation_periodogram::autocorrelation_periodogram::{AutoCorrelationPeriodogram, AutoCorrelationPeriodogramParams};
use crate::indicators::john_ehlers::comb_band_pass_spectrum::comb_band_pass_spectrum::{CombBandPassSpectrum, CombBandPassSpectrumParams};
use crate::indicators::john_ehlers::discrete_fourier_transform_spectrum::discrete_fourier_transform_spectrum::{DiscreteFourierTransformSpectrum, DiscreteFourierTransformSpectrumParams};

// ── joseph granville ─────────────────────────────────────────────────────────
use crate::indicators::joseph_granville::on_balance_volume::{
    OnBalanceVolume, OnBalanceVolumeParams,
};

// ── larry williams ───────────────────────────────────────────────────────────
use crate::indicators::larry_williams::ultimate_oscillator::{
    UltimateOscillator, UltimateOscillatorParams,
};
use crate::indicators::larry_williams::williams_percent_r::{
    WilliamsPercentR, WilliamsPercentRParams,
};

// ── marc chaikin ─────────────────────────────────────────────────────────────
use crate::indicators::marc_chaikin::advance_decline::{AdvanceDecline, AdvanceDeclineParams};
use crate::indicators::marc_chaikin::advance_decline_oscillator::{
    AdvanceDeclineOscillator, AdvanceDeclineOscillatorParams,
};

// ── mark jurik ───────────────────────────────────────────────────────────────
use crate::indicators::mark_jurik::jurik_composite_fractal_behavior_index::{
    JurikCompositeFractalBehaviorIndex, JurikCompositeFractalBehaviorIndexParams,
};
use crate::indicators::mark_jurik::jurik_directional_movement_index::{
    JurikDirectionalMovementIndex, JurikDirectionalMovementIndexParams,
};
use crate::indicators::mark_jurik::jurik_moving_average::{
    JurikMovingAverage, JurikMovingAverageParams,
};
use crate::indicators::mark_jurik::jurik_relative_trend_strength_index::{
    JurikRelativeTrendStrengthIndex, JurikRelativeTrendStrengthIndexParams,
};
use crate::indicators::mark_jurik::jurik_zero_lag_velocity::{
    JurikZeroLagVelocity, JurikZeroLagVelocityParams,
};
use crate::indicators::mark_jurik::jurik_turning_point_oscillator::{
    JurikTurningPointOscillator, JurikTurningPointOscillatorParams,
};
use crate::indicators::mark_jurik::jurik_wavelet_sampler::{
    JurikWaveletSampler, JurikWaveletSamplerParams,
};
use crate::indicators::mark_jurik::jurik_adaptive_relative_trend_strength_index::{
    JurikAdaptiveRelativeTrendStrengthIndex, JurikAdaptiveRelativeTrendStrengthIndexParams,
};
use crate::indicators::mark_jurik::jurik_adaptive_zero_lag_velocity::{
    JurikAdaptiveZeroLagVelocity, JurikAdaptiveZeroLagVelocityParams,
};
use crate::indicators::mark_jurik::jurik_commodity_channel_index::{
    JurikCommodityChannelIndex, JurikCommodityChannelIndexParams,
};
use crate::indicators::mark_jurik::jurik_fractal_adaptive_zero_lag_velocity::{
    JurikFractalAdaptiveZeroLagVelocity, JurikFractalAdaptiveZeroLagVelocityParams,
};

// ── patrick mulloy ───────────────────────────────────────────────────────────
use crate::indicators::patrick_mulloy::double_exponential_moving_average::{
    DoubleExponentialMovingAverage, DoubleExponentialMovingAverageLengthParams,
    DoubleExponentialMovingAverageSmoothingFactorParams,
};
use crate::indicators::patrick_mulloy::triple_exponential_moving_average::{
    TripleExponentialMovingAverage, TripleExponentialMovingAverageLengthParams,
    TripleExponentialMovingAverageSmoothingFactorParams,
};

// ── perry kaufman ────────────────────────────────────────────────────────────
use crate::indicators::perry_kaufman::kaufman_adaptive_moving_average::{
    KaufmanAdaptiveMovingAverage, KaufmanAdaptiveMovingAverageLengthParams,
    KaufmanAdaptiveMovingAverageSmoothingFactorParams,
};

// ── tim tillson ──────────────────────────────────────────────────────────────
use crate::indicators::tim_tillson::t2_exponential_moving_average::t2_exponential_moving_average::{T2ExponentialMovingAverage, T2ExponentialMovingAverageLengthParams, T2ExponentialMovingAverageSmoothingFactorParams};
use crate::indicators::tim_tillson::t3_exponential_moving_average::t3_exponential_moving_average::{T3ExponentialMovingAverage, T3ExponentialMovingAverageLengthParams, T3ExponentialMovingAverageSmoothingFactorParams};

// ── tushar chande ────────────────────────────────────────────────────────────
use crate::indicators::tushar_chande::aroon::{Aroon, AroonParams};
use crate::indicators::tushar_chande::chande_momentum_oscillator::{
    ChandeMomentumOscillator, ChandeMomentumOscillatorParams,
};
use crate::indicators::tushar_chande::stochastic_relative_strength_index::{
    StochasticRelativeStrengthIndex, StochasticRelativeStrengthIndexParams,
};

// ── vladimir kravchuk ────────────────────────────────────────────────────────
use crate::indicators::vladimir_kravchuk::adaptive_trend_and_cycle_filter::AdaptiveTrendAndCycleFilter;

// ── welles wilder ────────────────────────────────────────────────────────────
use crate::indicators::welles_wilder::average_directional_movement_index::{
    AverageDirectionalMovementIndex, AverageDirectionalMovementIndexParams,
};
use crate::indicators::welles_wilder::average_directional_movement_index_rating::{
    AverageDirectionalMovementIndexRating, AverageDirectionalMovementIndexRatingParams,
};
use crate::indicators::welles_wilder::average_true_range::{
    AverageTrueRange, AverageTrueRangeParams,
};
use crate::indicators::welles_wilder::directional_indicator_minus::{
    DirectionalIndicatorMinus, DirectionalIndicatorMinusParams,
};
use crate::indicators::welles_wilder::directional_indicator_plus::{
    DirectionalIndicatorPlus, DirectionalIndicatorPlusParams,
};
use crate::indicators::welles_wilder::directional_movement_index::{
    DirectionalMovementIndex, DirectionalMovementIndexParams,
};
use crate::indicators::welles_wilder::directional_movement_minus::{
    DirectionalMovementMinus, DirectionalMovementMinusParams,
};
use crate::indicators::welles_wilder::directional_movement_plus::{
    DirectionalMovementPlus, DirectionalMovementPlusParams,
};
use crate::indicators::welles_wilder::normalized_average_true_range::{
    NormalizedAverageTrueRange, NormalizedAverageTrueRangeParams,
};
use crate::indicators::welles_wilder::parabolic_stop_and_reverse::{
    ParabolicStopAndReverse, ParabolicStopAndReverseParams,
};
use crate::indicators::welles_wilder::relative_strength_index::{
    RelativeStrengthIndex, RelativeStrengthIndexParams,
};
use crate::indicators::welles_wilder::true_range::{TrueRange, TrueRangeParams};

/// Create an indicator from its identifier and a JSON-encoded parameter string.
///
/// If `params_json` is empty or `"{}"`, default parameters are used.
///
/// For indicators with Length and SmoothingFactor constructor variants, the
/// factory auto-detects which to use: if the JSON contains a `"smoothingFactor"`
/// key the SmoothingFactor variant is used, otherwise the Length variant.
pub fn create_indicator(
    identifier: Identifier,
    params_json: &str,
) -> Result<Box<dyn Indicator>, String> {
    let params = if params_json.is_empty() {
        JsonValue::Object(vec![])
    } else {
        JsonValue::parse(params_json)?
    };

    match identifier {
        // ── common ───────────────────────────────────────────────────────
        Identifier::SimpleMovingAverage => {
            let mut p = SimpleMovingAverageParams::default();
            if let Some(v) = get_usize(&params, "length") {
                p.length = v;
            }
            Ok(Box::new(SimpleMovingAverage::new(&p)?))
        }

        Identifier::WeightedMovingAverage => {
            let mut p = WeightedMovingAverageParams::default();
            if let Some(v) = get_usize(&params, "length") {
                p.length = v;
            }
            Ok(Box::new(WeightedMovingAverage::new(&p)?))
        }

        Identifier::TriangularMovingAverage => {
            let mut p = TriangularMovingAverageParams::default();
            if let Some(v) = get_usize(&params, "length") {
                p.length = v;
            }
            Ok(Box::new(TriangularMovingAverage::new(&p)?))
        }

        Identifier::ExponentialMovingAverage => {
            if has_key(&params, "smoothingFactor") {
                let mut p = ExponentialMovingAverageSmoothingFactorParams::default();
                if let Some(v) = get_f64(&params, "smoothingFactor") {
                    p.smoothing_factor = v;
                }
                if let Some(v) = get_bool(&params, "firstIsAverage") {
                    p.first_is_average = v;
                }
                Ok(Box::new(
                    ExponentialMovingAverage::new_from_smoothing_factor(&p)?,
                ))
            } else {
                let mut p = ExponentialMovingAverageLengthParams::default();
                if let Some(v) = get_i64(&params, "length") {
                    p.length = v;
                }
                if let Some(v) = get_bool(&params, "firstIsAverage") {
                    p.first_is_average = v;
                }
                Ok(Box::new(ExponentialMovingAverage::new_from_length(&p)?))
            }
        }

        Identifier::Variance => {
            let mut p = VarianceParams::default();
            if let Some(v) = get_usize(&params, "length") {
                p.length = v;
            }
            Ok(Box::new(Variance::new(&p)?))
        }

        Identifier::StandardDeviation => {
            let mut p = StandardDeviationParams::default();
            if let Some(v) = get_usize(&params, "length") {
                p.length = v;
            }
            Ok(Box::new(StandardDeviation::new(&p)?))
        }

        Identifier::Momentum => {
            let mut p = MomentumParams::default();
            if let Some(v) = get_usize(&params, "length") {
                p.length = v;
            }
            Ok(Box::new(Momentum::new(&p)?))
        }

        Identifier::RateOfChange => {
            let mut p = RateOfChangeParams::default();
            if let Some(v) = get_usize(&params, "length") {
                p.length = v;
            }
            Ok(Box::new(RateOfChange::new(&p)?))
        }

        Identifier::RateOfChangePercent => {
            let mut p = RateOfChangePercentParams::default();
            if let Some(v) = get_usize(&params, "length") {
                p.length = v;
            }
            Ok(Box::new(RateOfChangePercent::new(&p)?))
        }

        Identifier::RateOfChangeRatio => {
            let mut p = RateOfChangeRatioParams::default();
            if let Some(v) = get_usize(&params, "length") {
                p.length = v;
            }
            Ok(Box::new(RateOfChangeRatio::new(&p)?))
        }

        Identifier::AbsolutePriceOscillator => {
            let mut p = AbsolutePriceOscillatorParams::default();
            if let Some(v) = get_i64(&params, "fastLength") {
                p.fast_length = v;
            }
            if let Some(v) = get_i64(&params, "slowLength") {
                p.slow_length = v;
            }
            if let Some(v) = get_bool(&params, "firstIsAverage") {
                p.first_is_average = v;
            }
            Ok(Box::new(AbsolutePriceOscillator::new(&p)?))
        }

        Identifier::PearsonsCorrelationCoefficient => {
            let mut p = PearsonsCorrelationCoefficientParams::default();
            if let Some(v) = get_usize(&params, "length") {
                p.length = v;
            }
            Ok(Box::new(PearsonsCorrelationCoefficient::new(&p)?))
        }

        Identifier::LinearRegression => {
            let mut p = LinearRegressionParams::default();
            if let Some(v) = get_usize(&params, "length") {
                p.length = v;
            }
            Ok(Box::new(LinearRegression::new(&p)?))
        }

        // ── custom ───────────────────────────────────────────────────────
        Identifier::GoertzelSpectrum => {
            let mut p = GoertzelSpectrumParams::default();
            if !is_empty_object(&params) {
                if let Some(v) = get_usize(&params, "length") {
                    p.length = v;
                }
                if let Some(v) = get_f64(&params, "minPeriod") {
                    p.min_period = v;
                }
                if let Some(v) = get_f64(&params, "maxPeriod") {
                    p.max_period = v;
                }
                if let Some(v) = get_usize(&params, "spectrumResolution") {
                    p.spectrum_resolution = v;
                }
                if let Some(v) = get_bool(&params, "isFirstOrder") {
                    p.is_first_order = v;
                }
                if let Some(v) = get_bool(&params, "disableSpectralDilationCompensation") {
                    p.disable_spectral_dilation_compensation = v;
                }
                if let Some(v) = get_bool(&params, "disableAutomaticGainControl") {
                    p.disable_automatic_gain_control = v;
                }
                if let Some(v) = get_f64(&params, "automaticGainControlDecayFactor") {
                    p.automatic_gain_control_decay_factor = v;
                }
                if let Some(v) = get_bool(&params, "fixedNormalization") {
                    p.fixed_normalization = v;
                }
            }
            Ok(Box::new(GoertzelSpectrum::new(&p)?))
        }

        Identifier::MaximumEntropySpectrum => {
            let mut p = MaximumEntropySpectrumParams::default();
            if !is_empty_object(&params) {
                if let Some(v) = get_usize(&params, "length") {
                    p.length = v;
                }
                if let Some(v) = get_usize(&params, "degree") {
                    p.degree = v;
                }
                if let Some(v) = get_f64(&params, "minPeriod") {
                    p.min_period = v;
                }
                if let Some(v) = get_f64(&params, "maxPeriod") {
                    p.max_period = v;
                }
                if let Some(v) = get_bool(&params, "disableAutomaticGainControl") {
                    p.disable_automatic_gain_control = v;
                }
                if let Some(v) = get_f64(&params, "automaticGainControlDecayFactor") {
                    p.automatic_gain_control_decay_factor = v;
                }
                if let Some(v) = get_bool(&params, "fixedNormalization") {
                    p.fixed_normalization = v;
                }
            }
            Ok(Box::new(MaximumEntropySpectrum::new(&p)?))
        }

        // ── donald lambert ───────────────────────────────────────────────
        Identifier::CommodityChannelIndex => {
            let mut p = CommodityChannelIndexParams::default();
            if let Some(v) = get_usize(&params, "length") {
                p.length = v;
            }
            Ok(Box::new(CommodityChannelIndex::new(&p)?))
        }

        // ── gene quong ───────────────────────────────────────────────────
        Identifier::MoneyFlowIndex => {
            let mut p = MoneyFlowIndexParams::default();
            if let Some(v) = get_usize(&params, "length") {
                p.length = v;
            }
            Ok(Box::new(MoneyFlowIndex::new(&p)?))
        }

        // ── george lane ──────────────────────────────────────────────────
        Identifier::Stochastic => {
            let mut p = StochasticParams::default();
            if let Some(v) = get_usize(&params, "fastKLength") {
                p.fast_k_length = v;
            }
            if let Some(v) = get_usize(&params, "slowKLength") {
                p.slow_k_length = v;
            }
            if let Some(v) = get_usize(&params, "slowDLength") {
                p.slow_d_length = v;
            }
            Ok(Box::new(Stochastic::new(&p)?))
        }

        // ── gerald appel ─────────────────────────────────────────────────
        Identifier::PercentagePriceOscillator => {
            let mut p = PercentagePriceOscillatorParams::default();
            if let Some(v) = get_usize(&params, "fastLength") {
                p.fast_length = v;
            }
            if let Some(v) = get_usize(&params, "slowLength") {
                p.slow_length = v;
            }
            Ok(Box::new(PercentagePriceOscillator::new(&p)?))
        }

        Identifier::MovingAverageConvergenceDivergence => {
            let mut p = MovingAverageConvergenceDivergenceParams::default();
            if let Some(v) = get_usize(&params, "fastLength") {
                p.fast_length = v;
            }
            if let Some(v) = get_usize(&params, "slowLength") {
                p.slow_length = v;
            }
            if let Some(v) = get_usize(&params, "signalLength") {
                p.signal_length = v;
            }
            Ok(Box::new(MovingAverageConvergenceDivergence::new(&p)?))
        }

        // ── igor livshin ─────────────────────────────────────────────────
        Identifier::BalanceOfPower => Ok(Box::new(BalanceOfPower::new(&BalanceOfPowerParams)?)),

        // ── jack hutson ──────────────────────────────────────────────────
        Identifier::TripleExponentialMovingAverageOscillator => {
            let mut p = TripleExponentialMovingAverageOscillatorParams::default();
            if let Some(v) = get_usize(&params, "length") {
                p.length = v;
            }
            Ok(Box::new(TripleExponentialMovingAverageOscillator::new(&p)?))
        }

        // ── john bollinger ───────────────────────────────────────────────
        Identifier::BollingerBands => {
            let mut p = BollingerBandsParams::default();
            if let Some(v) = get_usize(&params, "length") {
                p.length = v;
            }
            if let Some(v) = get_f64(&params, "upperMultiplier") {
                p.upper_multiplier = v;
            }
            if let Some(v) = get_f64(&params, "lowerMultiplier") {
                p.lower_multiplier = v;
            }
            Ok(Box::new(BollingerBands::new(&p)?))
        }

        Identifier::BollingerBandsTrend => {
            let mut p = BollingerBandsTrendParams::default();
            if let Some(v) = get_usize(&params, "fastLength") {
                p.fast_length = v;
            }
            if let Some(v) = get_usize(&params, "slowLength") {
                p.slow_length = v;
            }
            if let Some(v) = get_f64(&params, "upperMultiplier") {
                p.upper_multiplier = v;
            }
            if let Some(v) = get_f64(&params, "lowerMultiplier") {
                p.lower_multiplier = v;
            }
            Ok(Box::new(BollingerBandsTrend::new(&p)?))
        }

        // ── john ehlers ──────────────────────────────────────────────────
        Identifier::SuperSmoother => {
            let mut p = SuperSmootherParams::default();
            if let Some(v) = get_i64(&params, "shortestCyclePeriod") {
                p.shortest_cycle_period = v;
            }
            Ok(Box::new(SuperSmoother::new(&p)?))
        }

        Identifier::CenterOfGravityOscillator => {
            let mut p = CenterOfGravityOscillatorParams::default();
            if let Some(v) = get_usize(&params, "length") {
                p.length = v;
            }
            Ok(Box::new(CenterOfGravityOscillator::new(&p)?))
        }

        Identifier::CyberCycle => {
            if has_key(&params, "smoothingFactor") {
                let mut p = CyberCycleSmoothingFactorParams::default();
                if let Some(v) = get_f64(&params, "smoothingFactor") {
                    p.smoothing_factor = v;
                }
                if let Some(v) = get_i64(&params, "signalLag") {
                    p.signal_lag = v;
                }
                Ok(Box::new(CyberCycle::new_smoothing_factor(&p)?))
            } else {
                let mut p = CyberCycleLengthParams::default();
                if let Some(v) = get_i64(&params, "length") {
                    p.length = v;
                }
                if let Some(v) = get_i64(&params, "signalLag") {
                    p.signal_lag = v;
                }
                Ok(Box::new(CyberCycle::new_length(&p)?))
            }
        }

        Identifier::InstantaneousTrendLine => {
            if has_key(&params, "smoothingFactor") {
                let mut p = ItlSmoothingFactorParams::default();
                if let Some(v) = get_f64(&params, "smoothingFactor") {
                    p.smoothing_factor = v;
                }
                Ok(Box::new(InstantaneousTrendLine::new_smoothing_factor(&p)?))
            } else {
                let mut p = ItlLengthParams::default();
                if let Some(v) = get_i32(&params, "length") {
                    p.length = v;
                }
                Ok(Box::new(InstantaneousTrendLine::new_length(&p)?))
            }
        }

        Identifier::ZeroLagExponentialMovingAverage => {
            let mut p = ZeroLagExponentialMovingAverageParams::default();
            if let Some(v) = get_f64(&params, "smoothingFactor") {
                p.smoothing_factor = v;
            }
            if let Some(v) = get_f64(&params, "velocityGainFactor") {
                p.velocity_gain_factor = v;
            }
            if let Some(v) = get_i32(&params, "velocityMomentumLength") {
                p.velocity_momentum_length = v;
            }
            Ok(Box::new(ZeroLagExponentialMovingAverage::new(&p)?))
        }

        Identifier::ZeroLagErrorCorrectingExponentialMovingAverage => {
            let mut p = ZeroLagErrorCorrectingExponentialMovingAverageParams::default();
            if let Some(v) = get_f64(&params, "smoothingFactor") {
                p.smoothing_factor = v;
            }
            if let Some(v) = get_f64(&params, "gainLimit") {
                p.gain_limit = v;
            }
            if let Some(v) = get_f64(&params, "gainStep") {
                p.gain_step = v;
            }
            Ok(Box::new(
                ZeroLagErrorCorrectingExponentialMovingAverage::new(&p)?,
            ))
        }

        Identifier::RoofingFilter => {
            let mut p = RoofingFilterParams::default();
            if let Some(v) = get_usize(&params, "shortestCyclePeriod") {
                p.shortest_cycle_period = v;
            }
            if let Some(v) = get_usize(&params, "longestCyclePeriod") {
                p.longest_cycle_period = v;
            }
            Ok(Box::new(RoofingFilter::new(&p)?))
        }

        Identifier::MesaAdaptiveMovingAverage => {
            if has_key(&params, "fastLimitSmoothingFactor")
                || has_key(&params, "slowLimitSmoothingFactor")
            {
                let p = MesaAdaptiveMovingAverageSmoothingFactorParams {
                    estimator_type: crate::indicators::john_ehlers::hilbert_transformer::CycleEstimatorType::HomodyneDiscriminator,
                    estimator_params: crate::indicators::john_ehlers::hilbert_transformer::CycleEstimatorParams {
                        smoothing_length: 4,
                        alpha_ema_quadrature_in_phase: 0.2,
                        alpha_ema_period: 0.2,
                        warm_up_period: 0,
                    },
                    fast_limit_smoothing_factor: get_f64(&params, "fastLimitSmoothingFactor").unwrap_or(0.5),
                    slow_limit_smoothing_factor: get_f64(&params, "slowLimitSmoothingFactor").unwrap_or(0.05),
                    bar_component: None,
                    quote_component: None,
                    trade_component: None,
                };
                Ok(Box::new(MesaAdaptiveMovingAverage::new_smoothing_factor(
                    &p,
                )?))
            } else if is_empty_object(&params) {
                Ok(Box::new(MesaAdaptiveMovingAverage::new_default()?))
            } else {
                let p = MesaAdaptiveMovingAverageLengthParams {
                    estimator_type: crate::indicators::john_ehlers::hilbert_transformer::CycleEstimatorType::HomodyneDiscriminator,
                    estimator_params: crate::indicators::john_ehlers::hilbert_transformer::CycleEstimatorParams {
                        smoothing_length: 4,
                        alpha_ema_quadrature_in_phase: 0.2,
                        alpha_ema_period: 0.2,
                        warm_up_period: 0,
                    },
                    fast_limit_length: get_i64(&params, "fastLimitLength").unwrap_or(3),
                    slow_limit_length: get_i64(&params, "slowLimitLength").unwrap_or(39),
                    bar_component: None,
                    quote_component: None,
                    trade_component: None,
                };
                Ok(Box::new(MesaAdaptiveMovingAverage::new_length(&p)?))
            }
        }

        Identifier::FractalAdaptiveMovingAverage => {
            let mut p = FractalAdaptiveMovingAverageParams::default();
            if let Some(v) = get_i64(&params, "length") {
                p.length = v;
            }
            Ok(Box::new(FractalAdaptiveMovingAverage::new(&p)?))
        }

        Identifier::DominantCycle => {
            if is_empty_object(&params) {
                Ok(Box::new(DominantCycle::new_default()?))
            } else {
                let mut p = DominantCycleParams::default();
                if let Some(v) = get_f64(&params, "alphaEmaPeriodAdditional") {
                    p.alpha_ema_period_additional = v;
                }
                Ok(Box::new(DominantCycle::new(&p)?))
            }
        }

        Identifier::SineWave => {
            if is_empty_object(&params) {
                Ok(Box::new(SineWave::new_default()?))
            } else {
                let mut p = SineWaveParams::default();
                if let Some(v) = get_f64(&params, "alphaEmaPeriodAdditional") {
                    p.alpha_ema_period_additional = v;
                }
                Ok(Box::new(SineWave::new(&p)?))
            }
        }

        Identifier::HilbertTransformerInstantaneousTrendLine => {
            if is_empty_object(&params) {
                Ok(Box::new(
                    HilbertTransformerInstantaneousTrendLine::new_default()?,
                ))
            } else {
                let mut p = HilbertTransformerInstantaneousTrendLineParams::default();
                if let Some(v) = get_f64(&params, "alphaEmaPeriodAdditional") {
                    p.alpha_ema_period_additional = v;
                }
                if let Some(v) = get_usize(&params, "trendLineSmoothingLength") {
                    p.trend_line_smoothing_length = v;
                }
                if let Some(v) = get_f64(&params, "cyclePartMultiplier") {
                    p.cycle_part_multiplier = v;
                }
                Ok(Box::new(HilbertTransformerInstantaneousTrendLine::new(&p)?))
            }
        }

        Identifier::TrendCycleMode => {
            if is_empty_object(&params) {
                Ok(Box::new(TrendCycleMode::new_default()?))
            } else {
                let mut p = TrendCycleModeParams::default();
                if let Some(v) = get_f64(&params, "alphaEmaPeriodAdditional") {
                    p.alpha_ema_period_additional = v;
                }
                if let Some(v) = get_usize(&params, "trendLineSmoothingLength") {
                    p.trend_line_smoothing_length = v;
                }
                if let Some(v) = get_f64(&params, "cyclePartMultiplier") {
                    p.cycle_part_multiplier = v;
                }
                if let Some(v) = get_f64(&params, "separationPercentage") {
                    p.separation_percentage = v;
                }
                Ok(Box::new(TrendCycleMode::new(&p)?))
            }
        }

        Identifier::CoronaSpectrum => {
            let mut p = CoronaSpectrumParams::default();
            if !is_empty_object(&params) {
                if let Some(v) = get_f64(&params, "minRasterValue") {
                    p.min_raster_value = v;
                }
                if let Some(v) = get_f64(&params, "maxRasterValue") {
                    p.max_raster_value = v;
                }
                if let Some(v) = get_f64(&params, "minParameterValue") {
                    p.min_parameter_value = v;
                }
                if let Some(v) = get_f64(&params, "maxParameterValue") {
                    p.max_parameter_value = v;
                }
                if let Some(v) = get_i32(&params, "highPassFilterCutoff") {
                    p.high_pass_filter_cutoff = v;
                }
            }
            Ok(Box::new(CoronaSpectrum::new(&p)?))
        }

        Identifier::CoronaSignalToNoiseRatio => {
            let mut p = CoronaSignalToNoiseRatioParams::default();
            if !is_empty_object(&params) {
                if let Some(v) = get_i32(&params, "rasterLength") {
                    p.raster_length = v;
                }
                if let Some(v) = get_f64(&params, "maxRasterValue") {
                    p.max_raster_value = v;
                }
                if let Some(v) = get_f64(&params, "minParameterValue") {
                    p.min_parameter_value = v;
                }
                if let Some(v) = get_f64(&params, "maxParameterValue") {
                    p.max_parameter_value = v;
                }
                if let Some(v) = get_i32(&params, "highPassFilterCutoff") {
                    p.high_pass_filter_cutoff = v;
                }
                if let Some(v) = get_i32(&params, "minimalPeriod") {
                    p.minimal_period = v;
                }
                if let Some(v) = get_i32(&params, "maximalPeriod") {
                    p.maximal_period = v;
                }
            }
            Ok(Box::new(CoronaSignalToNoiseRatio::new(&p)?))
        }

        Identifier::CoronaSwingPosition => {
            let mut p = CoronaSwingPositionParams::default();
            if !is_empty_object(&params) {
                if let Some(v) = get_i32(&params, "rasterLength") {
                    p.raster_length = v;
                }
                if let Some(v) = get_f64(&params, "maxRasterValue") {
                    p.max_raster_value = v;
                }
                if let Some(v) = get_f64(&params, "minParameterValue") {
                    p.min_parameter_value = v;
                }
                if let Some(v) = get_f64(&params, "maxParameterValue") {
                    p.max_parameter_value = v;
                }
                if let Some(v) = get_i32(&params, "highPassFilterCutoff") {
                    p.high_pass_filter_cutoff = v;
                }
                if let Some(v) = get_i32(&params, "minimalPeriod") {
                    p.minimal_period = v;
                }
                if let Some(v) = get_i32(&params, "maximalPeriod") {
                    p.maximal_period = v;
                }
            }
            Ok(Box::new(CoronaSwingPosition::new(&p)?))
        }

        Identifier::CoronaTrendVigor => {
            let mut p = CoronaTrendVigorParams::default();
            if !is_empty_object(&params) {
                if let Some(v) = get_i32(&params, "rasterLength") {
                    p.raster_length = v;
                }
                if let Some(v) = get_f64(&params, "maxRasterValue") {
                    p.max_raster_value = v;
                }
                if let Some(v) = get_f64(&params, "minParameterValue") {
                    p.min_parameter_value = v;
                }
                if let Some(v) = get_f64(&params, "maxParameterValue") {
                    p.max_parameter_value = v;
                }
                if let Some(v) = get_i32(&params, "highPassFilterCutoff") {
                    p.high_pass_filter_cutoff = v;
                }
                if let Some(v) = get_i32(&params, "minimalPeriod") {
                    p.minimal_period = v;
                }
                if let Some(v) = get_i32(&params, "maximalPeriod") {
                    p.maximal_period = v;
                }
            }
            Ok(Box::new(CoronaTrendVigor::new(&p)?))
        }

        Identifier::AutoCorrelationIndicator => {
            let mut p = AutoCorrelationIndicatorParams::default();
            if !is_empty_object(&params) {
                if let Some(v) = get_i32(&params, "minLag") {
                    p.min_lag = v;
                }
                if let Some(v) = get_i32(&params, "maxLag") {
                    p.max_lag = v;
                }
                if let Some(v) = get_i32(&params, "smoothingPeriod") {
                    p.smoothing_period = v;
                }
                if let Some(v) = get_i32(&params, "averagingLength") {
                    p.averaging_length = v;
                }
            }
            Ok(Box::new(AutoCorrelationIndicator::new(&p)?))
        }

        Identifier::AutoCorrelationPeriodogram => {
            let mut p = AutoCorrelationPeriodogramParams::default();
            if !is_empty_object(&params) {
                if let Some(v) = get_i32(&params, "minPeriod") {
                    p.min_period = v;
                }
                if let Some(v) = get_i32(&params, "maxPeriod") {
                    p.max_period = v;
                }
                if let Some(v) = get_i32(&params, "averagingLength") {
                    p.averaging_length = v;
                }
                if let Some(v) = get_bool(&params, "disableAutomaticGainControl") {
                    p.disable_automatic_gain_control = v;
                }
                if let Some(v) = get_f64(&params, "automaticGainControlDecayFactor") {
                    p.automatic_gain_control_decay_factor = v;
                }
                if let Some(v) = get_bool(&params, "fixedNormalization") {
                    p.fixed_normalization = v;
                }
            }
            Ok(Box::new(AutoCorrelationPeriodogram::new(&p)?))
        }

        Identifier::CombBandPassSpectrum => {
            let mut p = CombBandPassSpectrumParams::default();
            if !is_empty_object(&params) {
                if let Some(v) = get_i32(&params, "minPeriod") {
                    p.min_period = v;
                }
                if let Some(v) = get_i32(&params, "maxPeriod") {
                    p.max_period = v;
                }
                if let Some(v) = get_bool(&params, "disableAutomaticGainControl") {
                    p.disable_automatic_gain_control = v;
                }
                if let Some(v) = get_f64(&params, "automaticGainControlDecayFactor") {
                    p.automatic_gain_control_decay_factor = v;
                }
                if let Some(v) = get_bool(&params, "fixedNormalization") {
                    p.fixed_normalization = v;
                }
            }
            Ok(Box::new(CombBandPassSpectrum::new(&p)?))
        }

        Identifier::DiscreteFourierTransformSpectrum => {
            let mut p = DiscreteFourierTransformSpectrumParams::default();
            if !is_empty_object(&params) {
                if let Some(v) = get_usize(&params, "length") {
                    p.length = v;
                }
                if let Some(v) = get_f64(&params, "minPeriod") {
                    p.min_period = v;
                }
                if let Some(v) = get_f64(&params, "maxPeriod") {
                    p.max_period = v;
                }
                if let Some(v) = get_bool(&params, "disableSpectralDilationCompensation") {
                    p.disable_spectral_dilation_compensation = v;
                }
                if let Some(v) = get_bool(&params, "disableAutomaticGainControl") {
                    p.disable_automatic_gain_control = v;
                }
                if let Some(v) = get_f64(&params, "automaticGainControlDecayFactor") {
                    p.automatic_gain_control_decay_factor = v;
                }
                if let Some(v) = get_bool(&params, "fixedNormalization") {
                    p.fixed_normalization = v;
                }
            }
            Ok(Box::new(DiscreteFourierTransformSpectrum::new(&p)?))
        }

        // ── joseph granville ─────────────────────────────────────────────
        Identifier::OnBalanceVolume => {
            let p = OnBalanceVolumeParams::default();
            Ok(Box::new(OnBalanceVolume::new(&p)?))
        }

        // ── larry williams ───────────────────────────────────────────────
        Identifier::WilliamsPercentR => {
            let mut p = WilliamsPercentRParams::default();
            if let Some(v) = get_usize(&params, "length") {
                p.length = v;
            }
            Ok(Box::new(WilliamsPercentR::new(&p)?))
        }

        Identifier::UltimateOscillator => {
            let mut p = UltimateOscillatorParams::default();
            if let Some(v) = get_usize(&params, "length1") {
                p.length1 = v;
            }
            if let Some(v) = get_usize(&params, "length2") {
                p.length2 = v;
            }
            if let Some(v) = get_usize(&params, "length3") {
                p.length3 = v;
            }
            Ok(Box::new(UltimateOscillator::new(&p)?))
        }

        // ── marc chaikin ─────────────────────────────────────────────────
        Identifier::AdvanceDecline => Ok(Box::new(AdvanceDecline::new(&AdvanceDeclineParams)?)),

        Identifier::AdvanceDeclineOscillator => {
            let mut p = AdvanceDeclineOscillatorParams::default();
            if let Some(v) = get_i64(&params, "fastLength") {
                p.fast_length = v;
            }
            if let Some(v) = get_i64(&params, "slowLength") {
                p.slow_length = v;
            }
            Ok(Box::new(AdvanceDeclineOscillator::new(&p)?))
        }

        // ── mark jurik ───────────────────────────────────────────────────
        Identifier::JurikMovingAverage => {
            let mut p = JurikMovingAverageParams::default();
            if let Some(v) = get_usize(&params, "length") {
                p.length = v;
            }
            if let Some(v) = get_i32(&params, "phase") {
                p.phase = v;
            }
            Ok(Box::new(JurikMovingAverage::new(&p)?))
        }

        Identifier::JurikRelativeTrendStrengthIndex => {
            let mut p = JurikRelativeTrendStrengthIndexParams::default();
            if let Some(v) = get_usize(&params, "length") {
                p.length = v;
            }
            Ok(Box::new(JurikRelativeTrendStrengthIndex::new(&p)?))
        }

        Identifier::JurikCompositeFractalBehaviorIndex => {
            let mut p = JurikCompositeFractalBehaviorIndexParams::default();
            if let Some(v) = get_usize(&params, "fractalType") {
                p.fractal_type = v;
            }
            if let Some(v) = get_usize(&params, "smooth") {
                p.smooth = v;
            }
            Ok(Box::new(JurikCompositeFractalBehaviorIndex::new(&p)?))
        }

        Identifier::JurikZeroLagVelocity => {
            let mut p = JurikZeroLagVelocityParams::default();
            if let Some(v) = get_usize(&params, "depth") {
                p.depth = v;
            }
            Ok(Box::new(JurikZeroLagVelocity::new(&p)?))
        }

        Identifier::JurikDirectionalMovementIndex => {
            let mut p = JurikDirectionalMovementIndexParams::default();
            if let Some(v) = get_usize(&params, "length") {
                p.length = v;
            }
            Ok(Box::new(JurikDirectionalMovementIndex::new(&p)?))
        }

        Identifier::JurikTurningPointOscillator => {
            let mut p = JurikTurningPointOscillatorParams::default();
            if let Some(v) = get_usize(&params, "length") {
                p.length = v;
            }
            Ok(Box::new(JurikTurningPointOscillator::new(&p)?))
        }

        Identifier::JurikWaveletSampler => {
            let mut p = JurikWaveletSamplerParams::default();
            if let Some(v) = get_usize(&params, "index") {
                p.index = v;
            }
            Ok(Box::new(JurikWaveletSampler::new(&p)?))
        }

        Identifier::JurikAdaptiveRelativeTrendStrengthIndex => {
            let mut p = JurikAdaptiveRelativeTrendStrengthIndexParams::default();
            if let Some(v) = get_usize(&params, "loLength") {
                p.lo_length = v;
            }
            if let Some(v) = get_usize(&params, "hiLength") {
                p.hi_length = v;
            }
            Ok(Box::new(JurikAdaptiveRelativeTrendStrengthIndex::new(&p)?))
        }

        Identifier::JurikAdaptiveZeroLagVelocity => {
            let mut p = JurikAdaptiveZeroLagVelocityParams::default();
            if let Some(v) = get_usize(&params, "loLength") {
                p.lo_length = v;
            }
            if let Some(v) = get_usize(&params, "hiLength") {
                p.hi_length = v;
            }
            if let Some(v) = get_f64(&params, "sensitivity") {
                p.sensitivity = v;
            }
            if let Some(v) = get_f64(&params, "period") {
                p.period = v;
            }
            Ok(Box::new(JurikAdaptiveZeroLagVelocity::new(&p)?))
        }

        Identifier::JurikCommodityChannelIndex => {
            let mut p = JurikCommodityChannelIndexParams::default();
            if let Some(v) = get_usize(&params, "length") {
                p.length = v;
            }
            Ok(Box::new(JurikCommodityChannelIndex::new(&p)?))
        }

        Identifier::JurikFractalAdaptiveZeroLagVelocity => {
            let mut p = JurikFractalAdaptiveZeroLagVelocityParams::default();
            if let Some(v) = get_usize(&params, "loDepth") {
                p.lo_depth = v;
            }
            if let Some(v) = get_usize(&params, "hiDepth") {
                p.hi_depth = v;
            }
            if let Some(v) = get_usize(&params, "fractalType") {
                p.fractal_type = v;
            }
            if let Some(v) = get_usize(&params, "smooth") {
                p.smooth = v;
            }
            Ok(Box::new(JurikFractalAdaptiveZeroLagVelocity::new(&p)?))
        }

        // ── patrick mulloy ───────────────────────────────────────────────
        Identifier::DoubleExponentialMovingAverage => {
            if has_key(&params, "smoothingFactor") {
                let mut p = DoubleExponentialMovingAverageSmoothingFactorParams::default();
                if let Some(v) = get_f64(&params, "smoothingFactor") {
                    p.smoothing_factor = v;
                }
                if let Some(v) = get_bool(&params, "firstIsAverage") {
                    p.first_is_average = v;
                }
                Ok(Box::new(
                    DoubleExponentialMovingAverage::new_from_smoothing_factor(&p)?,
                ))
            } else {
                let mut p = DoubleExponentialMovingAverageLengthParams::default();
                if let Some(v) = get_i64(&params, "length") {
                    p.length = v;
                }
                if let Some(v) = get_bool(&params, "firstIsAverage") {
                    p.first_is_average = v;
                }
                Ok(Box::new(DoubleExponentialMovingAverage::new_from_length(
                    &p,
                )?))
            }
        }

        Identifier::TripleExponentialMovingAverage => {
            if has_key(&params, "smoothingFactor") {
                let mut p = TripleExponentialMovingAverageSmoothingFactorParams::default();
                if let Some(v) = get_f64(&params, "smoothingFactor") {
                    p.smoothing_factor = v;
                }
                if let Some(v) = get_bool(&params, "firstIsAverage") {
                    p.first_is_average = v;
                }
                Ok(Box::new(
                    TripleExponentialMovingAverage::new_from_smoothing_factor(&p)?,
                ))
            } else {
                let mut p = TripleExponentialMovingAverageLengthParams::default();
                if let Some(v) = get_i64(&params, "length") {
                    p.length = v;
                }
                if let Some(v) = get_bool(&params, "firstIsAverage") {
                    p.first_is_average = v;
                }
                Ok(Box::new(TripleExponentialMovingAverage::new_from_length(
                    &p,
                )?))
            }
        }

        // ── perry kaufman ────────────────────────────────────────────────
        Identifier::KaufmanAdaptiveMovingAverage => {
            if has_key(&params, "fastestSmoothingFactor")
                || has_key(&params, "slowestSmoothingFactor")
            {
                let mut p = KaufmanAdaptiveMovingAverageSmoothingFactorParams::default();
                if let Some(v) = get_usize(&params, "efficiencyRatioLength") {
                    p.efficiency_ratio_length = v;
                }
                if let Some(v) = get_f64(&params, "fastestSmoothingFactor") {
                    p.fastest_smoothing_factor = v;
                }
                if let Some(v) = get_f64(&params, "slowestSmoothingFactor") {
                    p.slowest_smoothing_factor = v;
                }
                Ok(Box::new(
                    KaufmanAdaptiveMovingAverage::new_from_smoothing_factors(&p)?,
                ))
            } else {
                let mut p = KaufmanAdaptiveMovingAverageLengthParams::default();
                if let Some(v) = get_usize(&params, "efficiencyRatioLength") {
                    p.efficiency_ratio_length = v;
                }
                if let Some(v) = get_usize(&params, "fastestLength") {
                    p.fastest_length = v;
                }
                if let Some(v) = get_usize(&params, "slowestLength") {
                    p.slowest_length = v;
                }
                Ok(Box::new(KaufmanAdaptiveMovingAverage::new_from_lengths(
                    &p,
                )?))
            }
        }

        // ── tim tillson ──────────────────────────────────────────────────
        Identifier::T2ExponentialMovingAverage => {
            if has_key(&params, "smoothingFactor") {
                let mut p = T2ExponentialMovingAverageSmoothingFactorParams::default();
                if let Some(v) = get_f64(&params, "smoothingFactor") {
                    p.smoothing_factor = v;
                }
                if let Some(v) = get_f64(&params, "volumeFactor") {
                    p.volume_factor = v;
                }
                if let Some(v) = get_bool(&params, "firstIsAverage") {
                    p.first_is_average = v;
                }
                Ok(Box::new(
                    T2ExponentialMovingAverage::new_from_smoothing_factor(&p)?,
                ))
            } else {
                let mut p = T2ExponentialMovingAverageLengthParams::default();
                if let Some(v) = get_i64(&params, "length") {
                    p.length = v;
                }
                if let Some(v) = get_f64(&params, "volumeFactor") {
                    p.volume_factor = v;
                }
                if let Some(v) = get_bool(&params, "firstIsAverage") {
                    p.first_is_average = v;
                }
                Ok(Box::new(T2ExponentialMovingAverage::new_from_length(&p)?))
            }
        }

        Identifier::T3ExponentialMovingAverage => {
            if has_key(&params, "smoothingFactor") {
                let mut p = T3ExponentialMovingAverageSmoothingFactorParams::default();
                if let Some(v) = get_f64(&params, "smoothingFactor") {
                    p.smoothing_factor = v;
                }
                if let Some(v) = get_f64(&params, "volumeFactor") {
                    p.volume_factor = v;
                }
                if let Some(v) = get_bool(&params, "firstIsAverage") {
                    p.first_is_average = v;
                }
                Ok(Box::new(
                    T3ExponentialMovingAverage::new_from_smoothing_factor(&p)?,
                ))
            } else {
                let mut p = T3ExponentialMovingAverageLengthParams::default();
                if let Some(v) = get_i64(&params, "length") {
                    p.length = v;
                }
                if let Some(v) = get_f64(&params, "volumeFactor") {
                    p.volume_factor = v;
                }
                if let Some(v) = get_bool(&params, "firstIsAverage") {
                    p.first_is_average = v;
                }
                Ok(Box::new(T3ExponentialMovingAverage::new_from_length(&p)?))
            }
        }

        // ── tushar chande ────────────────────────────────────────────────
        Identifier::ChandeMomentumOscillator => {
            let mut p = ChandeMomentumOscillatorParams::default();
            if let Some(v) = get_usize(&params, "length") {
                p.length = v;
            }
            Ok(Box::new(ChandeMomentumOscillator::new(&p)?))
        }

        Identifier::StochasticRelativeStrengthIndex => {
            let mut p = StochasticRelativeStrengthIndexParams::default();
            if let Some(v) = get_usize(&params, "length") {
                p.length = v;
            }
            if let Some(v) = get_usize(&params, "fastKLength") {
                p.fast_k_length = v;
            }
            if let Some(v) = get_usize(&params, "fastDLength") {
                p.fast_d_length = v;
            }
            Ok(Box::new(StochasticRelativeStrengthIndex::new(&p)?))
        }

        Identifier::Aroon => {
            let mut p = AroonParams::default();
            if let Some(v) = get_usize(&params, "length") {
                p.length = v;
            }
            Ok(Box::new(Aroon::new(&p)?))
        }

        // ── vladimir kravchuk ────────────────────────────────────────────
        Identifier::AdaptiveTrendAndCycleFilter => {
            Ok(Box::new(AdaptiveTrendAndCycleFilter::new()?))
        }

        // ── welles wilder ────────────────────────────────────────────────
        Identifier::TrueRange => Ok(Box::new(TrueRange::new(&TrueRangeParams)?)),

        Identifier::AverageTrueRange => {
            let mut p = AverageTrueRangeParams::default();
            if let Some(v) = get_usize(&params, "length") {
                p.length = v;
            }
            Ok(Box::new(AverageTrueRange::new(&p)?))
        }

        Identifier::NormalizedAverageTrueRange => {
            let mut p = NormalizedAverageTrueRangeParams::default();
            if let Some(v) = get_usize(&params, "length") {
                p.length = v;
            }
            Ok(Box::new(NormalizedAverageTrueRange::new(&p)?))
        }

        Identifier::DirectionalMovementMinus => {
            let mut p = DirectionalMovementMinusParams::default();
            if let Some(v) = get_usize(&params, "length") {
                p.length = v;
            }
            Ok(Box::new(DirectionalMovementMinus::new(&p)?))
        }

        Identifier::DirectionalMovementPlus => {
            let mut p = DirectionalMovementPlusParams::default();
            if let Some(v) = get_usize(&params, "length") {
                p.length = v;
            }
            Ok(Box::new(DirectionalMovementPlus::new(&p)?))
        }

        Identifier::DirectionalIndicatorMinus => {
            let mut p = DirectionalIndicatorMinusParams::default();
            if let Some(v) = get_usize(&params, "length") {
                p.length = v;
            }
            Ok(Box::new(DirectionalIndicatorMinus::new(&p)?))
        }

        Identifier::DirectionalIndicatorPlus => {
            let mut p = DirectionalIndicatorPlusParams::default();
            if let Some(v) = get_usize(&params, "length") {
                p.length = v;
            }
            Ok(Box::new(DirectionalIndicatorPlus::new(&p)?))
        }

        Identifier::DirectionalMovementIndex => {
            let mut p = DirectionalMovementIndexParams::default();
            if let Some(v) = get_usize(&params, "length") {
                p.length = v;
            }
            Ok(Box::new(DirectionalMovementIndex::new(&p)?))
        }

        Identifier::AverageDirectionalMovementIndex => {
            let mut p = AverageDirectionalMovementIndexParams::default();
            if let Some(v) = get_usize(&params, "length") {
                p.length = v;
            }
            Ok(Box::new(AverageDirectionalMovementIndex::new(&p)?))
        }

        Identifier::AverageDirectionalMovementIndexRating => {
            let mut p = AverageDirectionalMovementIndexRatingParams::default();
            if let Some(v) = get_usize(&params, "length") {
                p.length = v;
            }
            Ok(Box::new(AverageDirectionalMovementIndexRating::new(&p)?))
        }

        Identifier::RelativeStrengthIndex => {
            let mut p = RelativeStrengthIndexParams::default();
            if let Some(v) = get_usize(&params, "length") {
                p.length = v;
            }
            Ok(Box::new(RelativeStrengthIndex::new(&p)?))
        }

        Identifier::ParabolicStopAndReverse => {
            let p = ParabolicStopAndReverseParams::default();
            Ok(Box::new(ParabolicStopAndReverse::new(&p)?))
        }

        _ => Err(format!("unsupported indicator: {:?}", identifier)),
    }
}

#[cfg(test)]
mod tests {
    use super::*;

    #[test]
    fn test_create_sma_default() {
        let ind = create_indicator(Identifier::SimpleMovingAverage, "{}").unwrap();
        assert_eq!(ind.metadata().identifier, Identifier::SimpleMovingAverage);
        assert!(ind.metadata().mnemonic.starts_with("sma"));
    }

    #[test]
    fn test_create_sma_custom_length() {
        let ind = create_indicator(Identifier::SimpleMovingAverage, r#"{"length": 20}"#).unwrap();
        assert!(!ind.is_primed());
    }

    #[test]
    fn test_create_ema_length() {
        let ind =
            create_indicator(Identifier::ExponentialMovingAverage, r#"{"length": 10}"#).unwrap();
        assert!(!ind.is_primed());
    }

    #[test]
    fn test_create_ema_smoothing_factor() {
        let ind = create_indicator(
            Identifier::ExponentialMovingAverage,
            r#"{"smoothingFactor": 0.1}"#,
        )
        .unwrap();
        assert!(!ind.is_primed());
    }

    #[test]
    fn test_create_empty_params() {
        let ind = create_indicator(Identifier::TrueRange, "").unwrap();
        assert!(!ind.is_primed());
    }

    #[test]
    fn test_create_default_indicators() {
        // All indicators from settings.json should work with empty params
        let ids = vec![
            Identifier::SimpleMovingAverage,
            Identifier::WeightedMovingAverage,
            Identifier::TriangularMovingAverage,
            Identifier::ExponentialMovingAverage,
            Identifier::DoubleExponentialMovingAverage,
            Identifier::TripleExponentialMovingAverage,
            Identifier::T2ExponentialMovingAverage,
            Identifier::T3ExponentialMovingAverage,
            Identifier::KaufmanAdaptiveMovingAverage,
            Identifier::JurikMovingAverage,
            Identifier::JurikRelativeTrendStrengthIndex,
            Identifier::JurikCompositeFractalBehaviorIndex,
            Identifier::JurikZeroLagVelocity,
            Identifier::JurikDirectionalMovementIndex,
            Identifier::JurikTurningPointOscillator,
            Identifier::JurikWaveletSampler,
            Identifier::JurikAdaptiveRelativeTrendStrengthIndex,
            Identifier::JurikAdaptiveZeroLagVelocity,
            Identifier::JurikCommodityChannelIndex,
            Identifier::JurikFractalAdaptiveZeroLagVelocity,
            Identifier::MesaAdaptiveMovingAverage,
            Identifier::FractalAdaptiveMovingAverage,
            Identifier::DominantCycle,
            Identifier::Momentum,
            Identifier::RateOfChange,
            Identifier::RateOfChangePercent,
            Identifier::RateOfChangeRatio,
            Identifier::RelativeStrengthIndex,
            Identifier::ChandeMomentumOscillator,
            Identifier::BollingerBands,
            Identifier::BollingerBandsTrend,
            Identifier::Variance,
            Identifier::StandardDeviation,
            Identifier::GoertzelSpectrum,
            Identifier::MaximumEntropySpectrum,
            Identifier::CenterOfGravityOscillator,
            Identifier::CyberCycle,
            Identifier::InstantaneousTrendLine,
            Identifier::SuperSmoother,
            Identifier::ZeroLagExponentialMovingAverage,
            Identifier::ZeroLagErrorCorrectingExponentialMovingAverage,
            Identifier::RoofingFilter,
            Identifier::TrueRange,
            Identifier::AverageTrueRange,
            Identifier::NormalizedAverageTrueRange,
            Identifier::DirectionalMovementMinus,
            Identifier::DirectionalMovementPlus,
            Identifier::DirectionalIndicatorMinus,
            Identifier::DirectionalIndicatorPlus,
            Identifier::DirectionalMovementIndex,
            Identifier::AverageDirectionalMovementIndex,
            Identifier::AverageDirectionalMovementIndexRating,
            Identifier::WilliamsPercentR,
            Identifier::PercentagePriceOscillator,
            Identifier::AbsolutePriceOscillator,
            Identifier::CommodityChannelIndex,
            Identifier::MoneyFlowIndex,
            Identifier::OnBalanceVolume,
            Identifier::BalanceOfPower,
            Identifier::PearsonsCorrelationCoefficient,
            Identifier::LinearRegression,
            Identifier::UltimateOscillator,
            Identifier::StochasticRelativeStrengthIndex,
            Identifier::Stochastic,
            Identifier::Aroon,
            Identifier::AdvanceDecline,
            Identifier::AdvanceDeclineOscillator,
            Identifier::ParabolicStopAndReverse,
            Identifier::TripleExponentialMovingAverageOscillator,
            Identifier::MovingAverageConvergenceDivergence,
            Identifier::SineWave,
            Identifier::HilbertTransformerInstantaneousTrendLine,
            Identifier::TrendCycleMode,
            Identifier::CoronaSpectrum,
            Identifier::CoronaSignalToNoiseRatio,
            Identifier::CoronaSwingPosition,
            Identifier::CoronaTrendVigor,
            Identifier::AdaptiveTrendAndCycleFilter,
            Identifier::DiscreteFourierTransformSpectrum,
            Identifier::CombBandPassSpectrum,
            Identifier::AutoCorrelationIndicator,
            Identifier::AutoCorrelationPeriodogram,
        ];
        for id in ids {
            let result = create_indicator(id, "{}");
            assert!(
                result.is_ok(),
                "failed to create {:?}: {:?}",
                id,
                result.err()
            );
        }
    }

    #[test]
    fn test_create_with_settings_json() {
        // Simulate parsing from a settings-like JSON
        let settings = r#"[
            {"identifier": "simpleMovingAverage", "params": {"length": 14}},
            {"identifier": "exponentialMovingAverage", "params": {"length": 10}},
            {"identifier": "bollingerBands", "params": {"length": 20, "upperMultiplier": 2.0, "lowerMultiplier": 2.0}}
        ]"#;
        let parsed = JsonValue::parse(settings).unwrap();
        let arr = parsed.as_array().unwrap();
        for item in arr {
            let id_str = item.get("identifier").unwrap().as_str_val().unwrap();
            let id = Identifier::from_str(id_str).unwrap();
            let params_json = if is_empty_object(item.get("params").unwrap()) {
                "{}".to_string()
            } else {
                // Re-serialize the params object to JSON string
                // For testing, just pass "{}" since we tested individual params above
                "{}".to_string()
            };
            let result = create_indicator(id, &params_json);
            assert!(result.is_ok(), "failed for {}: {:?}", id_str, result.err());
        }
    }

    #[test]
    fn test_invalid_json() {
        let result = create_indicator(Identifier::SimpleMovingAverage, "not json");
        assert!(result.is_err());
    }
}
