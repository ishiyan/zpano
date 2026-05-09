"""Factory that maps an Identifier and an optional parameter dict to a fully
constructed Indicator instance, so callers don't need to import every indicator
module directly.

For indicators with Length / SmoothingFactor constructor variants the factory
auto-detects which to use: if the params dict contains a ``smoothing_factor``
key (or ``fast_limit_smoothing_factor``/``slow_limit_smoothing_factor`` for
MAMA, ``fastest_smoothing_factor``/``slowest_smoothing_factor`` for KAMA) the
SmoothingFactor variant is used; otherwise the Length variant.

For indicators with ``create`` / ``create_default`` static factories: if params
is ``None`` or an empty dict, ``create_default()`` is called; otherwise
``create(params)`` is used.
"""

from __future__ import annotations

from dataclasses import fields as dataclass_fields
from typing import Any, Optional

from ..core.identifier import Identifier
from ..core.indicator import Indicator


def _is_empty(params: Optional[dict[str, Any]]) -> bool:
    return params is None or len(params) == 0


def _has_key(params: Optional[dict[str, Any]], key: str) -> bool:
    return params is not None and key in params


def _apply(dc_instance: Any, params: Optional[dict[str, Any]]) -> Any:
    """Apply dict values onto a dataclass instance, returning the same instance."""
    if _is_empty(params):
        return dc_instance
    field_names = {f.name for f in dataclass_fields(dc_instance)}
    for k, v in params.items():
        if k in field_names:
            setattr(dc_instance, k, v)
    return dc_instance


def create_indicator(
    identifier: Identifier,
    params: Optional[dict[str, Any]] = None,
) -> Indicator:
    """Create an indicator from its identifier and an optional parameter dict.

    If *params* is ``None`` or empty, default parameters are used.
    """

    # ── common ────────────────────────────────────────────────────────────

    if identifier == Identifier.SIMPLE_MOVING_AVERAGE:
        from ..common.simple_moving_average.params import default_params
        from ..common.simple_moving_average.simple_moving_average import SimpleMovingAverage
        return SimpleMovingAverage(_apply(default_params(), params))

    if identifier == Identifier.WEIGHTED_MOVING_AVERAGE:
        from ..common.weighted_moving_average.params import default_params
        from ..common.weighted_moving_average.weighted_moving_average import WeightedMovingAverage
        return WeightedMovingAverage(_apply(default_params(), params))

    if identifier == Identifier.TRIANGULAR_MOVING_AVERAGE:
        from ..common.triangular_moving_average.params import default_params
        from ..common.triangular_moving_average.triangular_moving_average import TriangularMovingAverage
        return TriangularMovingAverage(_apply(default_params(), params))

    if identifier == Identifier.EXPONENTIAL_MOVING_AVERAGE:
        if _has_key(params, 'smoothing_factor'):
            from ..common.exponential_moving_average.params import ExponentialMovingAverageSmoothingFactorParams
            from ..common.exponential_moving_average.exponential_moving_average import ExponentialMovingAverage
            p = ExponentialMovingAverageSmoothingFactorParams()
            return ExponentialMovingAverage.from_smoothing_factor(_apply(p, params))
        from ..common.exponential_moving_average.params import default_length_params
        from ..common.exponential_moving_average.exponential_moving_average import ExponentialMovingAverage
        return ExponentialMovingAverage.from_length(_apply(default_length_params(), params))

    if identifier == Identifier.VARIANCE:
        from ..common.variance.params import default_params
        from ..common.variance.variance import Variance
        return Variance(_apply(default_params(), params))

    if identifier == Identifier.STANDARD_DEVIATION:
        from ..common.standard_deviation.params import default_params
        from ..common.standard_deviation.standard_deviation import StandardDeviation
        return StandardDeviation(_apply(default_params(), params))

    if identifier == Identifier.MOMENTUM:
        from ..common.momentum.params import default_params
        from ..common.momentum.momentum import Momentum
        return Momentum(_apply(default_params(), params))

    if identifier == Identifier.RATE_OF_CHANGE:
        from ..common.rate_of_change.params import default_params
        from ..common.rate_of_change.rate_of_change import RateOfChange
        return RateOfChange(_apply(default_params(), params))

    if identifier == Identifier.RATE_OF_CHANGE_PERCENT:
        from ..common.rate_of_change_percent.params import default_params
        from ..common.rate_of_change_percent.rate_of_change_percent import RateOfChangePercent
        return RateOfChangePercent(_apply(default_params(), params))

    if identifier == Identifier.RATE_OF_CHANGE_RATIO:
        from ..common.rate_of_change_ratio.params import default_params
        from ..common.rate_of_change_ratio.rate_of_change_ratio import RateOfChangeRatio
        return RateOfChangeRatio(_apply(default_params(), params))

    if identifier == Identifier.ABSOLUTE_PRICE_OSCILLATOR:
        from ..common.absolute_price_oscillator.params import default_params
        from ..common.absolute_price_oscillator.absolute_price_oscillator import AbsolutePriceOscillator
        return AbsolutePriceOscillator(_apply(default_params(), params))

    if identifier == Identifier.PEARSONS_CORRELATION_COEFFICIENT:
        from ..common.pearsons_correlation_coefficient.params import default_params
        from ..common.pearsons_correlation_coefficient.pearsons_correlation_coefficient import PearsonsCorrelationCoefficient
        return PearsonsCorrelationCoefficient(_apply(default_params(), params))

    if identifier == Identifier.LINEAR_REGRESSION:
        from ..common.linear_regression.params import default_params
        from ..common.linear_regression.linear_regression import LinearRegression
        return LinearRegression(_apply(default_params(), params))

    # ── arnaud legoux ─────────────────────────────────────────────────────

    if identifier == Identifier.ARNAUD_LEGOUX_MOVING_AVERAGE:
        from ..arnaud_legoux.arnaud_legoux_moving_average.params import default_params
        from ..arnaud_legoux.arnaud_legoux_moving_average.arnaud_legoux_moving_average import ArnaudLegouxMovingAverage
        return ArnaudLegouxMovingAverage(_apply(default_params(), params))

    # ── donald lambert ────────────────────────────────────────────────────

    if identifier == Identifier.COMMODITY_CHANNEL_INDEX:
        from ..donald_lambert.commodity_channel_index.params import default_params
        from ..donald_lambert.commodity_channel_index.commodity_channel_index import CommodityChannelIndex
        return CommodityChannelIndex(_apply(default_params(), params))

    # ── gene quong ────────────────────────────────────────────────────────

    if identifier == Identifier.MONEY_FLOW_INDEX:
        from ..gene_quong.money_flow_index.params import default_params
        from ..gene_quong.money_flow_index.money_flow_index import MoneyFlowIndex
        return MoneyFlowIndex(_apply(default_params(), params))

    # ── george lane ───────────────────────────────────────────────────────

    if identifier == Identifier.STOCHASTIC:
        from ..george_lane.stochastic.params import default_params
        from ..george_lane.stochastic.stochastic import Stochastic
        return Stochastic(_apply(default_params(), params))

    # ── gerald appel ──────────────────────────────────────────────────────

    if identifier == Identifier.PERCENTAGE_PRICE_OSCILLATOR:
        from ..gerald_appel.percentage_price_oscillator.params import default_params
        from ..gerald_appel.percentage_price_oscillator.percentage_price_oscillator import PercentagePriceOscillator
        return PercentagePriceOscillator(_apply(default_params(), params))

    if identifier == Identifier.MOVING_AVERAGE_CONVERGENCE_DIVERGENCE:
        from ..gerald_appel.moving_average_convergence_divergence.params import default_params
        from ..gerald_appel.moving_average_convergence_divergence.moving_average_convergence_divergence import MovingAverageConvergenceDivergence
        return MovingAverageConvergenceDivergence(_apply(default_params(), params))

    # ── igor livshin ──────────────────────────────────────────────────────

    if identifier == Identifier.BALANCE_OF_POWER:
        from ..igor_livshin.balance_of_power.params import default_params
        from ..igor_livshin.balance_of_power.balance_of_power import BalanceOfPower
        return BalanceOfPower(_apply(default_params(), params))

    # ── jack hutson ───────────────────────────────────────────────────────

    if identifier == Identifier.TRIPLE_EXPONENTIAL_MOVING_AVERAGE_OSCILLATOR:
        from ..jack_hutson.triple_exponential_moving_average_oscillator.params import default_params
        from ..jack_hutson.triple_exponential_moving_average_oscillator.triple_exponential_moving_average_oscillator import TripleExponentialMovingAverageOscillator
        return TripleExponentialMovingAverageOscillator(_apply(default_params(), params))

    # ── john bollinger ────────────────────────────────────────────────────

    if identifier == Identifier.BOLLINGER_BANDS:
        from ..john_bollinger.bollinger_bands.params import default_params
        from ..john_bollinger.bollinger_bands.bollinger_bands import BollingerBands
        return BollingerBands(_apply(default_params(), params))

    if identifier == Identifier.BOLLINGER_BANDS_TREND:
        from ..john_bollinger.bollinger_bands_trend.params import default_params
        from ..john_bollinger.bollinger_bands_trend.bollinger_bands_trend import BollingerBandsTrend
        return BollingerBandsTrend(_apply(default_params(), params))

    # ── john ehlers ───────────────────────────────────────────────────────

    if identifier == Identifier.SUPER_SMOOTHER:
        from ..john_ehlers.super_smoother.params import default_params
        from ..john_ehlers.super_smoother.super_smoother import SuperSmoother
        return SuperSmoother.create(_apply(default_params(), params))

    if identifier == Identifier.CENTER_OF_GRAVITY_OSCILLATOR:
        from ..john_ehlers.center_of_gravity_oscillator.params import default_params
        from ..john_ehlers.center_of_gravity_oscillator.center_of_gravity_oscillator import CenterOfGravityOscillator
        return CenterOfGravityOscillator.create(_apply(default_params(), params))

    if identifier == Identifier.CYBER_CYCLE:
        if _has_key(params, 'smoothing_factor'):
            from ..john_ehlers.cyber_cycle.params import default_smoothing_factor_params
            from ..john_ehlers.cyber_cycle.cyber_cycle import CyberCycle
            return CyberCycle.from_smoothing_factor(_apply(default_smoothing_factor_params(), params))
        from ..john_ehlers.cyber_cycle.params import default_length_params
        from ..john_ehlers.cyber_cycle.cyber_cycle import CyberCycle
        return CyberCycle.from_length(_apply(default_length_params(), params))

    if identifier == Identifier.INSTANTANEOUS_TREND_LINE:
        if _has_key(params, 'smoothing_factor'):
            from ..john_ehlers.instantaneous_trend_line.params import default_smoothing_factor_params
            from ..john_ehlers.instantaneous_trend_line.instantaneous_trend_line import InstantaneousTrendLine
            return InstantaneousTrendLine.from_smoothing_factor(_apply(default_smoothing_factor_params(), params))
        from ..john_ehlers.instantaneous_trend_line.params import default_length_params
        from ..john_ehlers.instantaneous_trend_line.instantaneous_trend_line import InstantaneousTrendLine
        return InstantaneousTrendLine.from_length(_apply(default_length_params(), params))

    if identifier == Identifier.ZERO_LAG_EXPONENTIAL_MOVING_AVERAGE:
        from ..john_ehlers.zero_lag_exponential_moving_average.params import default_params
        from ..john_ehlers.zero_lag_exponential_moving_average.zero_lag_exponential_moving_average import ZeroLagExponentialMovingAverage
        return ZeroLagExponentialMovingAverage.create(_apply(default_params(), params))

    if identifier == Identifier.ZERO_LAG_ERROR_CORRECTING_EXPONENTIAL_MOVING_AVERAGE:
        from ..john_ehlers.zero_lag_error_correcting_exponential_moving_average.params import default_params
        from ..john_ehlers.zero_lag_error_correcting_exponential_moving_average.zero_lag_error_correcting_exponential_moving_average import ZeroLagErrorCorrectingExponentialMovingAverage
        return ZeroLagErrorCorrectingExponentialMovingAverage.create(_apply(default_params(), params))

    if identifier == Identifier.ROOFING_FILTER:
        from ..john_ehlers.roofing_filter.params import default_params
        from ..john_ehlers.roofing_filter.roofing_filter import RoofingFilter
        return RoofingFilter.create(_apply(default_params(), params))

    if identifier == Identifier.MESA_ADAPTIVE_MOVING_AVERAGE:
        if _has_key(params, 'fast_limit_smoothing_factor') or _has_key(params, 'slow_limit_smoothing_factor'):
            from ..john_ehlers.mesa_adaptive_moving_average.params import MesaAdaptiveMovingAverageSmoothingFactorParams
            from ..john_ehlers.mesa_adaptive_moving_average.mesa_adaptive_moving_average import MesaAdaptiveMovingAverage
            p = MesaAdaptiveMovingAverageSmoothingFactorParams()
            return MesaAdaptiveMovingAverage.from_smoothing_factor(_apply(p, params))
        if _is_empty(params):
            from ..john_ehlers.mesa_adaptive_moving_average.mesa_adaptive_moving_average import MesaAdaptiveMovingAverage
            from ..john_ehlers.mesa_adaptive_moving_average.params import MesaAdaptiveMovingAverageLengthParams
            return MesaAdaptiveMovingAverage.from_length(MesaAdaptiveMovingAverageLengthParams())
        from ..john_ehlers.mesa_adaptive_moving_average.params import MesaAdaptiveMovingAverageLengthParams
        from ..john_ehlers.mesa_adaptive_moving_average.mesa_adaptive_moving_average import MesaAdaptiveMovingAverage
        return MesaAdaptiveMovingAverage.from_length(_apply(MesaAdaptiveMovingAverageLengthParams(), params))

    if identifier == Identifier.FRACTAL_ADAPTIVE_MOVING_AVERAGE:
        from ..john_ehlers.fractal_adaptive_moving_average.params import default_params
        from ..john_ehlers.fractal_adaptive_moving_average.fractal_adaptive_moving_average import FractalAdaptiveMovingAverage
        if _is_empty(params):
            return FractalAdaptiveMovingAverage.create_default()
        return FractalAdaptiveMovingAverage.create(_apply(default_params(), params))

    if identifier == Identifier.DOMINANT_CYCLE:
        from ..john_ehlers.dominant_cycle.dominant_cycle import DominantCycle
        if _is_empty(params):
            return DominantCycle.create_default()
        from ..john_ehlers.dominant_cycle.params import default_params
        return DominantCycle.create(_apply(default_params(), params))

    if identifier == Identifier.SINE_WAVE:
        from ..john_ehlers.sine_wave.sine_wave import SineWave
        if _is_empty(params):
            return SineWave.create_default()
        from ..john_ehlers.sine_wave.params import default_params
        return SineWave.create(_apply(default_params(), params))

    if identifier == Identifier.HILBERT_TRANSFORMER_INSTANTANEOUS_TREND_LINE:
        from ..john_ehlers.hilbert_transformer_instantaneous_trend_line.hilbert_transformer_instantaneous_trend_line import HilbertTransformerInstantaneousTrendLine
        if _is_empty(params):
            return HilbertTransformerInstantaneousTrendLine.create_default()
        from ..john_ehlers.hilbert_transformer_instantaneous_trend_line.params import default_params
        return HilbertTransformerInstantaneousTrendLine.create(_apply(default_params(), params))

    if identifier == Identifier.TREND_CYCLE_MODE:
        from ..john_ehlers.trend_cycle_mode.trend_cycle_mode import TrendCycleMode
        if _is_empty(params):
            return TrendCycleMode.create_default()
        from ..john_ehlers.trend_cycle_mode.params import default_params
        return TrendCycleMode.create(_apply(default_params(), params))

    if identifier == Identifier.CORONA_SPECTRUM:
        from ..john_ehlers.corona_spectrum.params import default_params
        from ..john_ehlers.corona_spectrum.corona_spectrum import CoronaSpectrum
        return CoronaSpectrum(_apply(default_params(), params))

    if identifier == Identifier.CORONA_SIGNAL_TO_NOISE_RATIO:
        from ..john_ehlers.corona_signal_to_noise_ratio.params import default_params
        from ..john_ehlers.corona_signal_to_noise_ratio.corona_signal_to_noise_ratio import CoronaSignalToNoiseRatio
        return CoronaSignalToNoiseRatio(_apply(default_params(), params))

    if identifier == Identifier.CORONA_SWING_POSITION:
        from ..john_ehlers.corona_swing_position.params import default_params
        from ..john_ehlers.corona_swing_position.corona_swing_position import CoronaSwingPosition
        return CoronaSwingPosition(_apply(default_params(), params))

    if identifier == Identifier.CORONA_TREND_VIGOR:
        from ..john_ehlers.corona_trend_vigor.params import default_params
        from ..john_ehlers.corona_trend_vigor.corona_trend_vigor import CoronaTrendVigor
        return CoronaTrendVigor(_apply(default_params(), params))

    if identifier == Identifier.AUTO_CORRELATION_INDICATOR:
        from ..john_ehlers.auto_correlation_indicator.params import default_params
        from ..john_ehlers.auto_correlation_indicator.auto_correlation_indicator import AutoCorrelationIndicator
        return AutoCorrelationIndicator(_apply(default_params(), params))

    if identifier == Identifier.AUTO_CORRELATION_PERIODOGRAM:
        from ..john_ehlers.auto_correlation_periodogram.params import default_params
        from ..john_ehlers.auto_correlation_periodogram.auto_correlation_periodogram import AutoCorrelationPeriodogram
        return AutoCorrelationPeriodogram(_apply(default_params(), params))

    if identifier == Identifier.COMB_BAND_PASS_SPECTRUM:
        from ..john_ehlers.comb_band_pass_spectrum.params import default_params
        from ..john_ehlers.comb_band_pass_spectrum.comb_band_pass_spectrum import CombBandPassSpectrum
        return CombBandPassSpectrum(_apply(default_params(), params))

    if identifier == Identifier.DISCRETE_FOURIER_TRANSFORM_SPECTRUM:
        from ..john_ehlers.discrete_fourier_transform_spectrum.params import default_params
        from ..john_ehlers.discrete_fourier_transform_spectrum.discrete_fourier_transform_spectrum import DiscreteFourierTransformSpectrum
        return DiscreteFourierTransformSpectrum(_apply(default_params(), params))

    # ── joseph granville ──────────────────────────────────────────────────

    if identifier == Identifier.ON_BALANCE_VOLUME:
        from ..joseph_granville.on_balance_volume.params import default_params
        from ..joseph_granville.on_balance_volume.on_balance_volume import OnBalanceVolume
        return OnBalanceVolume(_apply(default_params(), params))

    # ── larry williams ────────────────────────────────────────────────────

    if identifier == Identifier.WILLIAMS_PERCENT_R:
        from ..larry_williams.williams_percent_r.williams_percent_r import WilliamsPercentR
        length = 14
        if params and 'length' in params:
            length = params['length']
        return WilliamsPercentR(length)

    if identifier == Identifier.ULTIMATE_OSCILLATOR:
        from ..larry_williams.ultimate_oscillator.params import default_params
        from ..larry_williams.ultimate_oscillator.ultimate_oscillator import UltimateOscillator
        return UltimateOscillator(_apply(default_params(), params))

    # ── manfred durschner ─────────────────────────────────────────────────

    if identifier == Identifier.NEW_MOVING_AVERAGE:
        from ..manfred_dürschner.new_moving_average.params import default_params
        from ..manfred_dürschner.new_moving_average.new_moving_average import NewMovingAverage
        return NewMovingAverage(_apply(default_params(), params))

    # ── marc chaikin ──────────────────────────────────────────────────────

    if identifier == Identifier.ADVANCE_DECLINE:
        from ..marc_chaikin.advance_decline.params import default_params
        from ..marc_chaikin.advance_decline.advance_decline import AdvanceDecline
        return AdvanceDecline(_apply(default_params(), params))

    if identifier == Identifier.ADVANCE_DECLINE_OSCILLATOR:
        from ..marc_chaikin.advance_decline_oscillator.params import default_params
        from ..marc_chaikin.advance_decline_oscillator.advance_decline_oscillator import AdvanceDeclineOscillator
        return AdvanceDeclineOscillator(_apply(default_params(), params))

    # ── mark jurik ────────────────────────────────────────────────────────

    if identifier == Identifier.JURIK_MOVING_AVERAGE:
        from ..mark_jurik.jurik_moving_average.params import default_params
        from ..mark_jurik.jurik_moving_average.jurik_moving_average import JurikMovingAverage
        return JurikMovingAverage(_apply(default_params(), params))

    if identifier == Identifier.JURIK_RELATIVE_TREND_STRENGTH_INDEX:
        from ..mark_jurik.jurik_relative_trend_strength_index.params import default_params
        from ..mark_jurik.jurik_relative_trend_strength_index.jurik_relative_trend_strength_index import JurikRelativeTrendStrengthIndex
        return JurikRelativeTrendStrengthIndex(_apply(default_params(), params))

    if identifier == Identifier.JURIK_COMPOSITE_FRACTAL_BEHAVIOR_INDEX:
        from ..mark_jurik.jurik_composite_fractal_behavior_index.params import default_params
        from ..mark_jurik.jurik_composite_fractal_behavior_index.jurik_composite_fractal_behavior_index import JurikCompositeFractalBehaviorIndex
        return JurikCompositeFractalBehaviorIndex(_apply(default_params(), params))

    if identifier == Identifier.JURIK_ZERO_LAG_VELOCITY:
        from ..mark_jurik.jurik_zero_lag_velocity.params import default_params
        from ..mark_jurik.jurik_zero_lag_velocity.jurik_zero_lag_velocity import JurikZeroLagVelocity
        return JurikZeroLagVelocity(_apply(default_params(), params))

    if identifier == Identifier.JURIK_DIRECTIONAL_MOVEMENT_INDEX:
        from ..mark_jurik.jurik_directional_movement_index.params import default_params
        from ..mark_jurik.jurik_directional_movement_index.jurik_directional_movement_index import JurikDirectionalMovementIndex
        return JurikDirectionalMovementIndex(_apply(default_params(), params))

    if identifier == Identifier.JURIK_ADAPTIVE_RELATIVE_TREND_STRENGTH_INDEX:
        from ..mark_jurik.jurik_adaptive_relative_trend_strength_index.params import default_params
        from ..mark_jurik.jurik_adaptive_relative_trend_strength_index.jurik_adaptive_relative_trend_strength_index import JurikAdaptiveRelativeTrendStrengthIndex
        return JurikAdaptiveRelativeTrendStrengthIndex(_apply(default_params(), params))

    if identifier == Identifier.JURIK_ADAPTIVE_ZERO_LAG_VELOCITY:
        from ..mark_jurik.jurik_adaptive_zero_lag_velocity.params import default_params
        from ..mark_jurik.jurik_adaptive_zero_lag_velocity.jurik_adaptive_zero_lag_velocity import JurikAdaptiveZeroLagVelocity
        return JurikAdaptiveZeroLagVelocity(_apply(default_params(), params))

    if identifier == Identifier.JURIK_COMMODITY_CHANNEL_INDEX:
        from ..mark_jurik.jurik_commodity_channel_index.params import default_params
        from ..mark_jurik.jurik_commodity_channel_index.jurik_commodity_channel_index import JurikCommodityChannelIndex
        return JurikCommodityChannelIndex(_apply(default_params(), params))

    if identifier == Identifier.JURIK_FRACTAL_ADAPTIVE_ZERO_LAG_VELOCITY:
        from ..mark_jurik.jurik_fractal_adaptive_zero_lag_velocity.params import default_params
        from ..mark_jurik.jurik_fractal_adaptive_zero_lag_velocity.jurik_fractal_adaptive_zero_lag_velocity import JurikFractalAdaptiveZeroLagVelocity
        return JurikFractalAdaptiveZeroLagVelocity(_apply(default_params(), params))

    if identifier == Identifier.JURIK_TURNING_POINT_OSCILLATOR:
        from ..mark_jurik.jurik_turning_point_oscillator.params import default_params
        from ..mark_jurik.jurik_turning_point_oscillator.jurik_turning_point_oscillator import JurikTurningPointOscillator
        return JurikTurningPointOscillator(_apply(default_params(), params))

    if identifier == Identifier.JURIK_WAVELET_SAMPLER:
        from ..mark_jurik.jurik_wavelet_sampler.params import default_params
        from ..mark_jurik.jurik_wavelet_sampler.jurik_wavelet_sampler import JurikWaveletSampler
        return JurikWaveletSampler(_apply(default_params(), params))

    # ── patrick mulloy ────────────────────────────────────────────────────

    if identifier == Identifier.DOUBLE_EXPONENTIAL_MOVING_AVERAGE:
        if _has_key(params, 'smoothing_factor'):
            from ..patrick_mulloy.double_exponential_moving_average.params import DoubleExponentialMovingAverageSmoothingFactorParams
            from ..patrick_mulloy.double_exponential_moving_average.double_exponential_moving_average import DoubleExponentialMovingAverage
            p = DoubleExponentialMovingAverageSmoothingFactorParams()
            return DoubleExponentialMovingAverage.from_smoothing_factor(_apply(p, params))
        from ..patrick_mulloy.double_exponential_moving_average.params import default_length_params
        from ..patrick_mulloy.double_exponential_moving_average.double_exponential_moving_average import DoubleExponentialMovingAverage
        return DoubleExponentialMovingAverage.from_length(_apply(default_length_params(), params))

    if identifier == Identifier.TRIPLE_EXPONENTIAL_MOVING_AVERAGE:
        if _has_key(params, 'smoothing_factor'):
            from ..patrick_mulloy.triple_exponential_moving_average.params import TripleExponentialMovingAverageSmoothingFactorParams
            from ..patrick_mulloy.triple_exponential_moving_average.triple_exponential_moving_average import TripleExponentialMovingAverage
            p = TripleExponentialMovingAverageSmoothingFactorParams()
            return TripleExponentialMovingAverage.from_smoothing_factor(_apply(p, params))
        from ..patrick_mulloy.triple_exponential_moving_average.params import default_length_params
        from ..patrick_mulloy.triple_exponential_moving_average.triple_exponential_moving_average import TripleExponentialMovingAverage
        return TripleExponentialMovingAverage.from_length(_apply(default_length_params(), params))

    # ── perry kaufman ─────────────────────────────────────────────────────

    if identifier == Identifier.KAUFMAN_ADAPTIVE_MOVING_AVERAGE:
        if _has_key(params, 'fastest_smoothing_factor') or _has_key(params, 'slowest_smoothing_factor'):
            from ..perry_kaufman.kaufman_adaptive_moving_average.params import default_smoothing_factor_params
            from ..perry_kaufman.kaufman_adaptive_moving_average.kaufman_adaptive_moving_average import KaufmanAdaptiveMovingAverage
            return KaufmanAdaptiveMovingAverage.from_smoothing_factor(_apply(default_smoothing_factor_params(), params))
        from ..perry_kaufman.kaufman_adaptive_moving_average.params import default_length_params
        from ..perry_kaufman.kaufman_adaptive_moving_average.kaufman_adaptive_moving_average import KaufmanAdaptiveMovingAverage
        return KaufmanAdaptiveMovingAverage.from_length(_apply(default_length_params(), params))

    # ── tim tillson ───────────────────────────────────────────────────────

    if identifier == Identifier.T2_EXPONENTIAL_MOVING_AVERAGE:
        if _has_key(params, 'smoothing_factor'):
            from ..tim_tillson.t2_exponential_moving_average.params import default_smoothing_factor_params
            from ..tim_tillson.t2_exponential_moving_average.t2_exponential_moving_average import T2ExponentialMovingAverage
            return T2ExponentialMovingAverage.from_smoothing_factor(_apply(default_smoothing_factor_params(), params))
        from ..tim_tillson.t2_exponential_moving_average.params import default_length_params
        from ..tim_tillson.t2_exponential_moving_average.t2_exponential_moving_average import T2ExponentialMovingAverage
        return T2ExponentialMovingAverage.from_length(_apply(default_length_params(), params))

    if identifier == Identifier.T3_EXPONENTIAL_MOVING_AVERAGE:
        if _has_key(params, 'smoothing_factor'):
            from ..tim_tillson.t3_exponential_moving_average.params import default_smoothing_factor_params
            from ..tim_tillson.t3_exponential_moving_average.t3_exponential_moving_average import T3ExponentialMovingAverage
            return T3ExponentialMovingAverage.from_smoothing_factor(_apply(default_smoothing_factor_params(), params))
        from ..tim_tillson.t3_exponential_moving_average.params import default_length_params
        from ..tim_tillson.t3_exponential_moving_average.t3_exponential_moving_average import T3ExponentialMovingAverage
        return T3ExponentialMovingAverage.from_length(_apply(default_length_params(), params))

    # ── tushar chande ─────────────────────────────────────────────────────

    if identifier == Identifier.CHANDE_MOMENTUM_OSCILLATOR:
        from ..tushar_chande.chande_momentum_oscillator.params import default_params
        from ..tushar_chande.chande_momentum_oscillator.chande_momentum_oscillator import ChandeMomentumOscillator
        return ChandeMomentumOscillator(_apply(default_params(), params))

    if identifier == Identifier.STOCHASTIC_RELATIVE_STRENGTH_INDEX:
        from ..tushar_chande.stochastic_relative_strength_index.params import default_params
        from ..tushar_chande.stochastic_relative_strength_index.stochastic_relative_strength_index import StochasticRelativeStrengthIndex
        return StochasticRelativeStrengthIndex(_apply(default_params(), params))

    if identifier == Identifier.AROON:
        from ..tushar_chande.aroon.params import default_params
        from ..tushar_chande.aroon.aroon import Aroon
        return Aroon(_apply(default_params(), params))

    # ── vladimir kravchuk ─────────────────────────────────────────────────

    if identifier == Identifier.ADAPTIVE_TREND_AND_CYCLE_FILTER:
        from ..vladimir_kravchuk.adaptive_trend_and_cycle_filter.params import default_params
        from ..vladimir_kravchuk.adaptive_trend_and_cycle_filter.adaptive_trend_and_cycle_filter import AdaptiveTrendAndCycleFilter
        if _is_empty(params):
            return AdaptiveTrendAndCycleFilter(_apply(default_params(), None))
        return AdaptiveTrendAndCycleFilter(_apply(default_params(), params))

    # ── welles wilder ─────────────────────────────────────────────────────

    if identifier == Identifier.TRUE_RANGE:
        from ..welles_wilder.true_range.true_range import TrueRange
        return TrueRange()

    if identifier == Identifier.AVERAGE_TRUE_RANGE:
        from ..welles_wilder.average_true_range.params import default_params
        from ..welles_wilder.average_true_range.average_true_range import AverageTrueRange
        return AverageTrueRange(_apply(default_params(), params))

    if identifier == Identifier.NORMALIZED_AVERAGE_TRUE_RANGE:
        from ..welles_wilder.normalized_average_true_range.params import default_params
        from ..welles_wilder.normalized_average_true_range.normalized_average_true_range import NormalizedAverageTrueRange
        return NormalizedAverageTrueRange(_apply(default_params(), params))

    if identifier == Identifier.DIRECTIONAL_MOVEMENT_MINUS:
        from ..welles_wilder.directional_movement_minus.params import default_params
        from ..welles_wilder.directional_movement_minus.directional_movement_minus import DirectionalMovementMinus
        return DirectionalMovementMinus(_apply(default_params(), params))

    if identifier == Identifier.DIRECTIONAL_MOVEMENT_PLUS:
        from ..welles_wilder.directional_movement_plus.params import default_params
        from ..welles_wilder.directional_movement_plus.directional_movement_plus import DirectionalMovementPlus
        return DirectionalMovementPlus(_apply(default_params(), params))

    if identifier == Identifier.DIRECTIONAL_INDICATOR_MINUS:
        from ..welles_wilder.directional_indicator_minus.params import default_params
        from ..welles_wilder.directional_indicator_minus.directional_indicator_minus import DirectionalIndicatorMinus
        return DirectionalIndicatorMinus(_apply(default_params(), params))

    if identifier == Identifier.DIRECTIONAL_INDICATOR_PLUS:
        from ..welles_wilder.directional_indicator_plus.params import default_params
        from ..welles_wilder.directional_indicator_plus.directional_indicator_plus import DirectionalIndicatorPlus
        return DirectionalIndicatorPlus(_apply(default_params(), params))

    if identifier == Identifier.DIRECTIONAL_MOVEMENT_INDEX:
        from ..welles_wilder.directional_movement_index.params import default_params
        from ..welles_wilder.directional_movement_index.directional_movement_index import DirectionalMovementIndex
        return DirectionalMovementIndex(_apply(default_params(), params))

    if identifier == Identifier.AVERAGE_DIRECTIONAL_MOVEMENT_INDEX:
        from ..welles_wilder.average_directional_movement_index.params import default_params
        from ..welles_wilder.average_directional_movement_index.average_directional_movement_index import AverageDirectionalMovementIndex
        return AverageDirectionalMovementIndex(_apply(default_params(), params))

    if identifier == Identifier.AVERAGE_DIRECTIONAL_MOVEMENT_INDEX_RATING:
        from ..welles_wilder.average_directional_movement_index_rating.params import default_params
        from ..welles_wilder.average_directional_movement_index_rating.average_directional_movement_index_rating import AverageDirectionalMovementIndexRating
        return AverageDirectionalMovementIndexRating(_apply(default_params(), params))

    if identifier == Identifier.RELATIVE_STRENGTH_INDEX:
        from ..welles_wilder.relative_strength_index.params import default_params
        from ..welles_wilder.relative_strength_index.relative_strength_index import RelativeStrengthIndex
        return RelativeStrengthIndex(_apply(default_params(), params))

    if identifier == Identifier.PARABOLIC_STOP_AND_REVERSE:
        from ..welles_wilder.parabolic_stop_and_reverse.params import default_params
        from ..welles_wilder.parabolic_stop_and_reverse.parabolic_stop_and_reverse import ParabolicStopAndReverse
        return ParabolicStopAndReverse(_apply(default_params(), params))

    # ── custom ────────────────────────────────────────────────────────────

    if identifier == Identifier.GOERTZEL_SPECTRUM:
        from ..custom.goertzel_spectrum.params import default_params
        from ..custom.goertzel_spectrum.goertzel_spectrum import GoertzelSpectrum
        return GoertzelSpectrum(_apply(default_params(), params))

    if identifier == Identifier.MAXIMUM_ENTROPY_SPECTRUM:
        from ..custom.maximum_entropy_spectrum.params import default_params
        from ..custom.maximum_entropy_spectrum.maximum_entropy_spectrum import MaximumEntropySpectrum
        return MaximumEntropySpectrum(_apply(default_params(), params))

    raise ValueError(f"unsupported indicator: {identifier}")
