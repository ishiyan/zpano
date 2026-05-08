"""Static registry of taxonomic descriptors for all implemented indicators.

Output Kind values use 0-based indexing (matching per-indicator Output enums which
start at 0 in Python and TypeScript, unlike Go which starts at iota+1).
"""

from __future__ import annotations

from .identifier import Identifier as Id
from .adaptivity import Adaptivity as A
from .input_requirement import InputRequirement as I
from .volume_usage import VolumeUsage as V
from .output_descriptor import OutputDescriptor
from .outputs.shape import Shape as S
from .role import Role as R
from .pane import Pane as P
from .descriptor import Descriptor


def _o(kind: int, shape: S, role: R, pane: P) -> OutputDescriptor:
    return OutputDescriptor(kind, shape, role, pane)


def _d(identifier: Id, family: str, adaptivity: A, input_requirement: I,
       volume_usage: V, outputs: list[OutputDescriptor]) -> Descriptor:
    return Descriptor(identifier, family, adaptivity, input_requirement, volume_usage, outputs)


_descriptors: dict[Id, Descriptor] = {

    # ── common ────────────────────────────────────────────────────────────

    Id.ABSOLUTE_PRICE_OSCILLATOR: _d(
        Id.ABSOLUTE_PRICE_OSCILLATOR, "Common", A.STATIC, I.SCALAR_INPUT, V.NO_VOLUME,
        [_o(0, S.SCALAR, R.OSCILLATOR, P.OWN)]),
    Id.EXPONENTIAL_MOVING_AVERAGE: _d(
        Id.EXPONENTIAL_MOVING_AVERAGE, "Common", A.STATIC, I.SCALAR_INPUT, V.NO_VOLUME,
        [_o(0, S.SCALAR, R.SMOOTHER, P.PRICE)]),
    Id.LINEAR_REGRESSION: _d(
        Id.LINEAR_REGRESSION, "Common", A.STATIC, I.SCALAR_INPUT, V.NO_VOLUME,
        [_o(0, S.SCALAR, R.SMOOTHER, P.PRICE),
         _o(1, S.SCALAR, R.SMOOTHER, P.PRICE),
         _o(2, S.SCALAR, R.SMOOTHER, P.PRICE),
         _o(3, S.SCALAR, R.OSCILLATOR, P.OWN),
         _o(4, S.SCALAR, R.OSCILLATOR, P.OWN)]),
    Id.MOMENTUM: _d(
        Id.MOMENTUM, "Common", A.STATIC, I.SCALAR_INPUT, V.NO_VOLUME,
        [_o(0, S.SCALAR, R.OSCILLATOR, P.OWN)]),
    Id.PEARSONS_CORRELATION_COEFFICIENT: _d(
        Id.PEARSONS_CORRELATION_COEFFICIENT, "Common", A.STATIC, I.SCALAR_INPUT, V.NO_VOLUME,
        [_o(0, S.SCALAR, R.CORRELATION, P.OWN)]),
    Id.RATE_OF_CHANGE: _d(
        Id.RATE_OF_CHANGE, "Common", A.STATIC, I.SCALAR_INPUT, V.NO_VOLUME,
        [_o(0, S.SCALAR, R.OSCILLATOR, P.OWN)]),
    Id.RATE_OF_CHANGE_PERCENT: _d(
        Id.RATE_OF_CHANGE_PERCENT, "Common", A.STATIC, I.SCALAR_INPUT, V.NO_VOLUME,
        [_o(0, S.SCALAR, R.OSCILLATOR, P.OWN)]),
    Id.RATE_OF_CHANGE_RATIO: _d(
        Id.RATE_OF_CHANGE_RATIO, "Common", A.STATIC, I.SCALAR_INPUT, V.NO_VOLUME,
        [_o(0, S.SCALAR, R.OSCILLATOR, P.OWN)]),
    Id.SIMPLE_MOVING_AVERAGE: _d(
        Id.SIMPLE_MOVING_AVERAGE, "Common", A.STATIC, I.SCALAR_INPUT, V.NO_VOLUME,
        [_o(0, S.SCALAR, R.SMOOTHER, P.PRICE)]),
    Id.STANDARD_DEVIATION: _d(
        Id.STANDARD_DEVIATION, "Common", A.STATIC, I.SCALAR_INPUT, V.NO_VOLUME,
        [_o(0, S.SCALAR, R.VOLATILITY, P.OWN)]),
    Id.TRIANGULAR_MOVING_AVERAGE: _d(
        Id.TRIANGULAR_MOVING_AVERAGE, "Common", A.STATIC, I.SCALAR_INPUT, V.NO_VOLUME,
        [_o(0, S.SCALAR, R.SMOOTHER, P.PRICE)]),
    Id.VARIANCE: _d(
        Id.VARIANCE, "Common", A.STATIC, I.SCALAR_INPUT, V.NO_VOLUME,
        [_o(0, S.SCALAR, R.VOLATILITY, P.OWN)]),
    Id.WEIGHTED_MOVING_AVERAGE: _d(
        Id.WEIGHTED_MOVING_AVERAGE, "Common", A.STATIC, I.SCALAR_INPUT, V.NO_VOLUME,
        [_o(0, S.SCALAR, R.SMOOTHER, P.PRICE)]),

    # ── arnaudlegoux ──────────────────────────────────────────────────────

    Id.ARNAUD_LEGOUX_MOVING_AVERAGE: _d(
        Id.ARNAUD_LEGOUX_MOVING_AVERAGE, "Arnaud Legoux", A.STATIC, I.SCALAR_INPUT, V.NO_VOLUME,
        [_o(0, S.SCALAR, R.SMOOTHER, P.PRICE)]),

    # ── donaldlambert ─────────────────────────────────────────────────────

    Id.COMMODITY_CHANNEL_INDEX: _d(
        Id.COMMODITY_CHANNEL_INDEX, "Donald Lambert", A.STATIC, I.BAR_INPUT, V.NO_VOLUME,
        [_o(0, S.SCALAR, R.BOUNDED_OSCILLATOR, P.OWN)]),

    # ── genequong ─────────────────────────────────────────────────────────

    Id.MONEY_FLOW_INDEX: _d(
        Id.MONEY_FLOW_INDEX, "Gene Quong", A.STATIC, I.BAR_INPUT, V.AGGREGATE_BAR_VOLUME,
        [_o(0, S.SCALAR, R.BOUNDED_OSCILLATOR, P.OWN)]),

    # ── georgelane ────────────────────────────────────────────────────────

    Id.STOCHASTIC: _d(
        Id.STOCHASTIC, "George Lane", A.STATIC, I.BAR_INPUT, V.NO_VOLUME,
        [_o(0, S.SCALAR, R.BOUNDED_OSCILLATOR, P.OWN),
         _o(1, S.SCALAR, R.BOUNDED_OSCILLATOR, P.OWN),
         _o(2, S.SCALAR, R.SIGNAL, P.OWN)]),

    # ── geraldappel ───────────────────────────────────────────────────────

    Id.MOVING_AVERAGE_CONVERGENCE_DIVERGENCE: _d(
        Id.MOVING_AVERAGE_CONVERGENCE_DIVERGENCE, "Gerald Appel", A.STATIC, I.SCALAR_INPUT, V.NO_VOLUME,
        [_o(0, S.SCALAR, R.OSCILLATOR, P.OWN),
         _o(1, S.SCALAR, R.SIGNAL, P.OWN),
         _o(2, S.SCALAR, R.HISTOGRAM, P.OWN)]),
    Id.PERCENTAGE_PRICE_OSCILLATOR: _d(
        Id.PERCENTAGE_PRICE_OSCILLATOR, "Gerald Appel", A.STATIC, I.SCALAR_INPUT, V.NO_VOLUME,
        [_o(0, S.SCALAR, R.OSCILLATOR, P.OWN)]),

    # ── igorlivshin ───────────────────────────────────────────────────────

    Id.BALANCE_OF_POWER: _d(
        Id.BALANCE_OF_POWER, "Igor Livshin", A.STATIC, I.BAR_INPUT, V.NO_VOLUME,
        [_o(0, S.SCALAR, R.BOUNDED_OSCILLATOR, P.OWN)]),

    # ── jackhutson ────────────────────────────────────────────────────────

    Id.TRIPLE_EXPONENTIAL_MOVING_AVERAGE_OSCILLATOR: _d(
        Id.TRIPLE_EXPONENTIAL_MOVING_AVERAGE_OSCILLATOR, "Jack Hutson", A.STATIC, I.SCALAR_INPUT, V.NO_VOLUME,
        [_o(0, S.SCALAR, R.OSCILLATOR, P.OWN)]),

    # ── johnbollinger ─────────────────────────────────────────────────────

    Id.BOLLINGER_BANDS: _d(
        Id.BOLLINGER_BANDS, "John Bollinger", A.STATIC, I.SCALAR_INPUT, V.NO_VOLUME,
        [_o(0, S.SCALAR, R.ENVELOPE, P.PRICE),
         _o(1, S.SCALAR, R.SMOOTHER, P.PRICE),
         _o(2, S.SCALAR, R.ENVELOPE, P.PRICE),
         _o(3, S.SCALAR, R.VOLATILITY, P.OWN),
         _o(4, S.SCALAR, R.BOUNDED_OSCILLATOR, P.OWN),
         _o(5, S.BAND, R.ENVELOPE, P.PRICE)]),
    Id.BOLLINGER_BANDS_TREND: _d(
        Id.BOLLINGER_BANDS_TREND, "John Bollinger", A.STATIC, I.SCALAR_INPUT, V.NO_VOLUME,
        [_o(0, S.SCALAR, R.OSCILLATOR, P.OWN)]),

    # ── johnehlers ────────────────────────────────────────────────────────

    Id.AUTO_CORRELATION_INDICATOR: _d(
        Id.AUTO_CORRELATION_INDICATOR, "John Ehlers", A.STATIC, I.SCALAR_INPUT, V.NO_VOLUME,
        [_o(0, S.HEATMAP, R.CORRELATION, P.OWN)]),
    Id.AUTO_CORRELATION_PERIODOGRAM: _d(
        Id.AUTO_CORRELATION_PERIODOGRAM, "John Ehlers", A.ADAPTIVE, I.SCALAR_INPUT, V.NO_VOLUME,
        [_o(0, S.HEATMAP, R.SPECTRUM, P.OWN)]),
    Id.CENTER_OF_GRAVITY_OSCILLATOR: _d(
        Id.CENTER_OF_GRAVITY_OSCILLATOR, "John Ehlers", A.STATIC, I.SCALAR_INPUT, V.NO_VOLUME,
        [_o(0, S.SCALAR, R.OSCILLATOR, P.OWN),
         _o(1, S.SCALAR, R.SIGNAL, P.OWN)]),
    Id.COMB_BAND_PASS_SPECTRUM: _d(
        Id.COMB_BAND_PASS_SPECTRUM, "John Ehlers", A.STATIC, I.SCALAR_INPUT, V.NO_VOLUME,
        [_o(0, S.HEATMAP, R.SPECTRUM, P.OWN)]),
    Id.CORONA_SIGNAL_TO_NOISE_RATIO: _d(
        Id.CORONA_SIGNAL_TO_NOISE_RATIO, "John Ehlers", A.ADAPTIVE, I.SCALAR_INPUT, V.NO_VOLUME,
        [_o(0, S.HEATMAP, R.SPECTRUM, P.OWN),
         _o(1, S.SCALAR, R.BOUNDED_OSCILLATOR, P.OWN)]),
    Id.CORONA_SPECTRUM: _d(
        Id.CORONA_SPECTRUM, "John Ehlers", A.ADAPTIVE, I.SCALAR_INPUT, V.NO_VOLUME,
        [_o(0, S.HEATMAP, R.SPECTRUM, P.OWN),
         _o(1, S.SCALAR, R.CYCLE_PERIOD, P.OWN),
         _o(2, S.SCALAR, R.CYCLE_PERIOD, P.OWN)]),
    Id.CORONA_SWING_POSITION: _d(
        Id.CORONA_SWING_POSITION, "John Ehlers", A.ADAPTIVE, I.SCALAR_INPUT, V.NO_VOLUME,
        [_o(0, S.HEATMAP, R.SPECTRUM, P.OWN),
         _o(1, S.SCALAR, R.BOUNDED_OSCILLATOR, P.OWN)]),
    Id.CORONA_TREND_VIGOR: _d(
        Id.CORONA_TREND_VIGOR, "John Ehlers", A.ADAPTIVE, I.SCALAR_INPUT, V.NO_VOLUME,
        [_o(0, S.HEATMAP, R.SPECTRUM, P.OWN),
         _o(1, S.SCALAR, R.OSCILLATOR, P.OWN)]),
    Id.CYBER_CYCLE: _d(
        Id.CYBER_CYCLE, "John Ehlers", A.STATIC, I.SCALAR_INPUT, V.NO_VOLUME,
        [_o(0, S.SCALAR, R.OSCILLATOR, P.OWN),
         _o(1, S.SCALAR, R.SIGNAL, P.OWN)]),
    Id.DISCRETE_FOURIER_TRANSFORM_SPECTRUM: _d(
        Id.DISCRETE_FOURIER_TRANSFORM_SPECTRUM, "John Ehlers", A.STATIC, I.SCALAR_INPUT, V.NO_VOLUME,
        [_o(0, S.HEATMAP, R.SPECTRUM, P.OWN)]),
    Id.DOMINANT_CYCLE: _d(
        Id.DOMINANT_CYCLE, "John Ehlers", A.ADAPTIVE, I.SCALAR_INPUT, V.NO_VOLUME,
        [_o(0, S.SCALAR, R.CYCLE_PERIOD, P.OWN),
         _o(1, S.SCALAR, R.CYCLE_PERIOD, P.OWN),
         _o(2, S.SCALAR, R.CYCLE_PHASE, P.OWN)]),
    Id.FRACTAL_ADAPTIVE_MOVING_AVERAGE: _d(
        Id.FRACTAL_ADAPTIVE_MOVING_AVERAGE, "John Ehlers", A.ADAPTIVE, I.SCALAR_INPUT, V.NO_VOLUME,
        [_o(0, S.SCALAR, R.SMOOTHER, P.PRICE),
         _o(1, S.SCALAR, R.FRACTAL_DIMENSION, P.OWN)]),
    Id.HILBERT_TRANSFORMER_INSTANTANEOUS_TREND_LINE: _d(
        Id.HILBERT_TRANSFORMER_INSTANTANEOUS_TREND_LINE, "John Ehlers", A.ADAPTIVE, I.SCALAR_INPUT, V.NO_VOLUME,
        [_o(0, S.SCALAR, R.SMOOTHER, P.PRICE),
         _o(1, S.SCALAR, R.CYCLE_PERIOD, P.OWN)]),
    Id.INSTANTANEOUS_TREND_LINE: _d(
        Id.INSTANTANEOUS_TREND_LINE, "John Ehlers", A.STATIC, I.SCALAR_INPUT, V.NO_VOLUME,
        [_o(0, S.SCALAR, R.SMOOTHER, P.PRICE),
         _o(1, S.SCALAR, R.SIGNAL, P.PRICE)]),
    Id.MESA_ADAPTIVE_MOVING_AVERAGE: _d(
        Id.MESA_ADAPTIVE_MOVING_AVERAGE, "John Ehlers", A.ADAPTIVE, I.SCALAR_INPUT, V.NO_VOLUME,
        [_o(0, S.SCALAR, R.SMOOTHER, P.PRICE),
         _o(1, S.SCALAR, R.SMOOTHER, P.PRICE),
         _o(2, S.BAND, R.ENVELOPE, P.PRICE)]),
    Id.ROOFING_FILTER: _d(
        Id.ROOFING_FILTER, "John Ehlers", A.STATIC, I.SCALAR_INPUT, V.NO_VOLUME,
        [_o(0, S.SCALAR, R.OSCILLATOR, P.OWN)]),
    Id.SINE_WAVE: _d(
        Id.SINE_WAVE, "John Ehlers", A.ADAPTIVE, I.SCALAR_INPUT, V.NO_VOLUME,
        [_o(0, S.SCALAR, R.OSCILLATOR, P.OWN),
         _o(1, S.SCALAR, R.SIGNAL, P.OWN),
         _o(2, S.BAND, R.ENVELOPE, P.OWN),
         _o(3, S.SCALAR, R.CYCLE_PERIOD, P.OWN),
         _o(4, S.SCALAR, R.CYCLE_PHASE, P.OWN)]),
    Id.SUPER_SMOOTHER: _d(
        Id.SUPER_SMOOTHER, "John Ehlers", A.STATIC, I.SCALAR_INPUT, V.NO_VOLUME,
        [_o(0, S.SCALAR, R.SMOOTHER, P.PRICE)]),
    Id.TREND_CYCLE_MODE: _d(
        Id.TREND_CYCLE_MODE, "John Ehlers", A.ADAPTIVE, I.SCALAR_INPUT, V.NO_VOLUME,
        [_o(0, S.SCALAR, R.REGIME_FLAG, P.OWN),
         _o(1, S.SCALAR, R.REGIME_FLAG, P.OWN),
         _o(2, S.SCALAR, R.REGIME_FLAG, P.OWN),
         _o(3, S.SCALAR, R.SMOOTHER, P.PRICE),
         _o(4, S.SCALAR, R.OSCILLATOR, P.OWN),
         _o(5, S.SCALAR, R.SIGNAL, P.OWN),
         _o(6, S.SCALAR, R.CYCLE_PERIOD, P.OWN),
         _o(7, S.SCALAR, R.CYCLE_PHASE, P.OWN)]),
    Id.ZERO_LAG_ERROR_CORRECTING_EXPONENTIAL_MOVING_AVERAGE: _d(
        Id.ZERO_LAG_ERROR_CORRECTING_EXPONENTIAL_MOVING_AVERAGE, "John Ehlers", A.STATIC, I.SCALAR_INPUT, V.NO_VOLUME,
        [_o(0, S.SCALAR, R.SMOOTHER, P.PRICE)]),
    Id.ZERO_LAG_EXPONENTIAL_MOVING_AVERAGE: _d(
        Id.ZERO_LAG_EXPONENTIAL_MOVING_AVERAGE, "John Ehlers", A.STATIC, I.SCALAR_INPUT, V.NO_VOLUME,
        [_o(0, S.SCALAR, R.SMOOTHER, P.PRICE)]),

    # ── josephgranville ───────────────────────────────────────────────────

    Id.ON_BALANCE_VOLUME: _d(
        Id.ON_BALANCE_VOLUME, "Joseph Granville", A.STATIC, I.BAR_INPUT, V.AGGREGATE_BAR_VOLUME,
        [_o(0, S.SCALAR, R.VOLUME_FLOW, P.OWN)]),

    # ── larrywilliams ─────────────────────────────────────────────────────

    Id.ULTIMATE_OSCILLATOR: _d(
        Id.ULTIMATE_OSCILLATOR, "Larry Williams", A.STATIC, I.BAR_INPUT, V.NO_VOLUME,
        [_o(0, S.SCALAR, R.BOUNDED_OSCILLATOR, P.OWN)]),
    Id.WILLIAMS_PERCENT_R: _d(
        Id.WILLIAMS_PERCENT_R, "Larry Williams", A.STATIC, I.BAR_INPUT, V.NO_VOLUME,
        [_o(0, S.SCALAR, R.BOUNDED_OSCILLATOR, P.OWN)]),

    # ── manfreddurschner ──────────────────────────────────────────────────

    Id.NEW_MOVING_AVERAGE: _d(
        Id.NEW_MOVING_AVERAGE, "Manfred Dürschner", A.STATIC, I.SCALAR_INPUT, V.NO_VOLUME,
        [_o(0, S.SCALAR, R.SMOOTHER, P.PRICE)]),

    # ── marcchaikin ───────────────────────────────────────────────────────

    Id.ADVANCE_DECLINE: _d(
        Id.ADVANCE_DECLINE, "Marc Chaikin", A.STATIC, I.BAR_INPUT, V.AGGREGATE_BAR_VOLUME,
        [_o(0, S.SCALAR, R.VOLUME_FLOW, P.OWN)]),
    Id.ADVANCE_DECLINE_OSCILLATOR: _d(
        Id.ADVANCE_DECLINE_OSCILLATOR, "Marc Chaikin", A.STATIC, I.BAR_INPUT, V.AGGREGATE_BAR_VOLUME,
        [_o(0, S.SCALAR, R.VOLUME_FLOW, P.OWN)]),

    # ── markjurik ─────────────────────────────────────────────────────────

    Id.JURIK_ADAPTIVE_RELATIVE_TREND_STRENGTH_INDEX: _d(
        Id.JURIK_ADAPTIVE_RELATIVE_TREND_STRENGTH_INDEX, "Mark Jurik", A.ADAPTIVE, I.SCALAR_INPUT, V.NO_VOLUME,
        [_o(0, S.SCALAR, R.OSCILLATOR, P.OWN)]),
    Id.JURIK_ADAPTIVE_ZERO_LAG_VELOCITY: _d(
        Id.JURIK_ADAPTIVE_ZERO_LAG_VELOCITY, "Mark Jurik", A.ADAPTIVE, I.SCALAR_INPUT, V.NO_VOLUME,
        [_o(0, S.SCALAR, R.OSCILLATOR, P.OWN)]),
    Id.JURIK_COMMODITY_CHANNEL_INDEX: _d(
        Id.JURIK_COMMODITY_CHANNEL_INDEX, "Mark Jurik", A.ADAPTIVE, I.SCALAR_INPUT, V.NO_VOLUME,
        [_o(0, S.SCALAR, R.OSCILLATOR, P.OWN)]),
    Id.JURIK_COMPOSITE_FRACTAL_BEHAVIOR_INDEX: _d(
        Id.JURIK_COMPOSITE_FRACTAL_BEHAVIOR_INDEX, "Mark Jurik", A.STATIC, I.SCALAR_INPUT, V.NO_VOLUME,
        [_o(0, S.SCALAR, R.OSCILLATOR, P.OWN)]),
    Id.JURIK_DIRECTIONAL_MOVEMENT_INDEX: _d(
        Id.JURIK_DIRECTIONAL_MOVEMENT_INDEX, "Mark Jurik", A.ADAPTIVE, I.BAR_INPUT, V.NO_VOLUME,
        [_o(0, S.SCALAR, R.OSCILLATOR, P.OWN),
         _o(1, S.SCALAR, R.OSCILLATOR, P.OWN),
         _o(2, S.SCALAR, R.OSCILLATOR, P.OWN)]),
    Id.JURIK_FRACTAL_ADAPTIVE_ZERO_LAG_VELOCITY: _d(
        Id.JURIK_FRACTAL_ADAPTIVE_ZERO_LAG_VELOCITY, "Mark Jurik", A.ADAPTIVE, I.SCALAR_INPUT, V.NO_VOLUME,
        [_o(0, S.SCALAR, R.OSCILLATOR, P.OWN)]),
    Id.JURIK_MOVING_AVERAGE: _d(
        Id.JURIK_MOVING_AVERAGE, "Mark Jurik", A.ADAPTIVE, I.SCALAR_INPUT, V.NO_VOLUME,
        [_o(0, S.SCALAR, R.SMOOTHER, P.PRICE)]),
    Id.JURIK_RELATIVE_TREND_STRENGTH_INDEX: _d(
        Id.JURIK_RELATIVE_TREND_STRENGTH_INDEX, "Mark Jurik", A.STATIC, I.SCALAR_INPUT, V.NO_VOLUME,
        [_o(0, S.SCALAR, R.OSCILLATOR, P.OWN)]),
    Id.JURIK_TURNING_POINT_OSCILLATOR: _d(
        Id.JURIK_TURNING_POINT_OSCILLATOR, "Mark Jurik", A.STATIC, I.SCALAR_INPUT, V.NO_VOLUME,
        [_o(0, S.SCALAR, R.OSCILLATOR, P.OWN)]),
    Id.JURIK_WAVELET_SAMPLER: _d(
        Id.JURIK_WAVELET_SAMPLER, "Mark Jurik", A.STATIC, I.SCALAR_INPUT, V.NO_VOLUME,
        [_o(0, S.SCALAR, R.SMOOTHER, P.PRICE)]),
    Id.JURIK_ZERO_LAG_VELOCITY: _d(
        Id.JURIK_ZERO_LAG_VELOCITY, "Mark Jurik", A.ADAPTIVE, I.SCALAR_INPUT, V.NO_VOLUME,
        [_o(0, S.SCALAR, R.OSCILLATOR, P.OWN)]),

    # ── patrickmulloy ─────────────────────────────────────────────────────

    Id.DOUBLE_EXPONENTIAL_MOVING_AVERAGE: _d(
        Id.DOUBLE_EXPONENTIAL_MOVING_AVERAGE, "Patrick Mulloy", A.STATIC, I.SCALAR_INPUT, V.NO_VOLUME,
        [_o(0, S.SCALAR, R.SMOOTHER, P.PRICE)]),
    Id.TRIPLE_EXPONENTIAL_MOVING_AVERAGE: _d(
        Id.TRIPLE_EXPONENTIAL_MOVING_AVERAGE, "Patrick Mulloy", A.STATIC, I.SCALAR_INPUT, V.NO_VOLUME,
        [_o(0, S.SCALAR, R.SMOOTHER, P.PRICE)]),

    # ── perrykaufman ──────────────────────────────────────────────────────

    Id.KAUFMAN_ADAPTIVE_MOVING_AVERAGE: _d(
        Id.KAUFMAN_ADAPTIVE_MOVING_AVERAGE, "Perry Kaufman", A.ADAPTIVE, I.SCALAR_INPUT, V.NO_VOLUME,
        [_o(0, S.SCALAR, R.SMOOTHER, P.PRICE)]),

    # ── timtillson ────────────────────────────────────────────────────────

    Id.T2_EXPONENTIAL_MOVING_AVERAGE: _d(
        Id.T2_EXPONENTIAL_MOVING_AVERAGE, "Tim Tillson", A.STATIC, I.SCALAR_INPUT, V.NO_VOLUME,
        [_o(0, S.SCALAR, R.SMOOTHER, P.PRICE)]),
    Id.T3_EXPONENTIAL_MOVING_AVERAGE: _d(
        Id.T3_EXPONENTIAL_MOVING_AVERAGE, "Tim Tillson", A.STATIC, I.SCALAR_INPUT, V.NO_VOLUME,
        [_o(0, S.SCALAR, R.SMOOTHER, P.PRICE)]),

    # ── tusharchande ──────────────────────────────────────────────────────

    Id.AROON: _d(
        Id.AROON, "Tushar Chande", A.STATIC, I.BAR_INPUT, V.NO_VOLUME,
        [_o(0, S.SCALAR, R.BOUNDED_OSCILLATOR, P.OWN),
         _o(1, S.SCALAR, R.BOUNDED_OSCILLATOR, P.OWN),
         _o(2, S.SCALAR, R.OSCILLATOR, P.OWN)]),
    Id.CHANDE_MOMENTUM_OSCILLATOR: _d(
        Id.CHANDE_MOMENTUM_OSCILLATOR, "Tushar Chande", A.STATIC, I.SCALAR_INPUT, V.NO_VOLUME,
        [_o(0, S.SCALAR, R.BOUNDED_OSCILLATOR, P.OWN)]),
    Id.STOCHASTIC_RELATIVE_STRENGTH_INDEX: _d(
        Id.STOCHASTIC_RELATIVE_STRENGTH_INDEX, "Tushar Chande", A.STATIC, I.SCALAR_INPUT, V.NO_VOLUME,
        [_o(0, S.SCALAR, R.BOUNDED_OSCILLATOR, P.OWN),
         _o(1, S.SCALAR, R.SIGNAL, P.OWN)]),

    # ── vladimirkravchuk ──────────────────────────────────────────────────

    Id.ADAPTIVE_TREND_AND_CYCLE_FILTER: _d(
        Id.ADAPTIVE_TREND_AND_CYCLE_FILTER, "Vladimir Kravchuk", A.ADAPTIVE, I.SCALAR_INPUT, V.NO_VOLUME,
        [_o(0, S.SCALAR, R.SMOOTHER, P.PRICE),
         _o(1, S.SCALAR, R.SMOOTHER, P.PRICE),
         _o(2, S.SCALAR, R.SMOOTHER, P.PRICE),
         _o(3, S.SCALAR, R.SMOOTHER, P.PRICE),
         _o(4, S.SCALAR, R.SMOOTHER, P.PRICE),
         _o(5, S.SCALAR, R.OSCILLATOR, P.OWN),
         _o(6, S.SCALAR, R.OSCILLATOR, P.OWN),
         _o(7, S.SCALAR, R.OSCILLATOR, P.OWN)]),

    # ── welleswilder ──────────────────────────────────────────────────────

    Id.AVERAGE_DIRECTIONAL_MOVEMENT_INDEX: _d(
        Id.AVERAGE_DIRECTIONAL_MOVEMENT_INDEX, "Welles Wilder", A.STATIC, I.BAR_INPUT, V.NO_VOLUME,
        [_o(0, S.SCALAR, R.BOUNDED_OSCILLATOR, P.OWN),
         _o(1, S.SCALAR, R.BOUNDED_OSCILLATOR, P.OWN),
         _o(2, S.SCALAR, R.DIRECTIONAL, P.OWN),
         _o(3, S.SCALAR, R.DIRECTIONAL, P.OWN),
         _o(4, S.SCALAR, R.DIRECTIONAL, P.OWN),
         _o(5, S.SCALAR, R.DIRECTIONAL, P.OWN),
         _o(6, S.SCALAR, R.VOLATILITY, P.OWN),
         _o(7, S.SCALAR, R.VOLATILITY, P.OWN)]),
    Id.AVERAGE_DIRECTIONAL_MOVEMENT_INDEX_RATING: _d(
        Id.AVERAGE_DIRECTIONAL_MOVEMENT_INDEX_RATING, "Welles Wilder", A.STATIC, I.BAR_INPUT, V.NO_VOLUME,
        [_o(0, S.SCALAR, R.BOUNDED_OSCILLATOR, P.OWN),
         _o(1, S.SCALAR, R.BOUNDED_OSCILLATOR, P.OWN),
         _o(2, S.SCALAR, R.BOUNDED_OSCILLATOR, P.OWN),
         _o(3, S.SCALAR, R.DIRECTIONAL, P.OWN),
         _o(4, S.SCALAR, R.DIRECTIONAL, P.OWN),
         _o(5, S.SCALAR, R.DIRECTIONAL, P.OWN),
         _o(6, S.SCALAR, R.DIRECTIONAL, P.OWN),
         _o(7, S.SCALAR, R.VOLATILITY, P.OWN),
         _o(8, S.SCALAR, R.VOLATILITY, P.OWN)]),
    Id.AVERAGE_TRUE_RANGE: _d(
        Id.AVERAGE_TRUE_RANGE, "Welles Wilder", A.STATIC, I.BAR_INPUT, V.NO_VOLUME,
        [_o(0, S.SCALAR, R.VOLATILITY, P.OWN)]),
    Id.DIRECTIONAL_INDICATOR_MINUS: _d(
        Id.DIRECTIONAL_INDICATOR_MINUS, "Welles Wilder", A.STATIC, I.BAR_INPUT, V.NO_VOLUME,
        [_o(0, S.SCALAR, R.DIRECTIONAL, P.OWN),
         _o(1, S.SCALAR, R.DIRECTIONAL, P.OWN),
         _o(2, S.SCALAR, R.VOLATILITY, P.OWN),
         _o(3, S.SCALAR, R.VOLATILITY, P.OWN)]),
    Id.DIRECTIONAL_INDICATOR_PLUS: _d(
        Id.DIRECTIONAL_INDICATOR_PLUS, "Welles Wilder", A.STATIC, I.BAR_INPUT, V.NO_VOLUME,
        [_o(0, S.SCALAR, R.DIRECTIONAL, P.OWN),
         _o(1, S.SCALAR, R.DIRECTIONAL, P.OWN),
         _o(2, S.SCALAR, R.VOLATILITY, P.OWN),
         _o(3, S.SCALAR, R.VOLATILITY, P.OWN)]),
    Id.DIRECTIONAL_MOVEMENT_INDEX: _d(
        Id.DIRECTIONAL_MOVEMENT_INDEX, "Welles Wilder", A.STATIC, I.BAR_INPUT, V.NO_VOLUME,
        [_o(0, S.SCALAR, R.BOUNDED_OSCILLATOR, P.OWN),
         _o(1, S.SCALAR, R.DIRECTIONAL, P.OWN),
         _o(2, S.SCALAR, R.DIRECTIONAL, P.OWN),
         _o(3, S.SCALAR, R.DIRECTIONAL, P.OWN),
         _o(4, S.SCALAR, R.DIRECTIONAL, P.OWN),
         _o(5, S.SCALAR, R.VOLATILITY, P.OWN),
         _o(6, S.SCALAR, R.VOLATILITY, P.OWN)]),
    Id.DIRECTIONAL_MOVEMENT_MINUS: _d(
        Id.DIRECTIONAL_MOVEMENT_MINUS, "Welles Wilder", A.STATIC, I.BAR_INPUT, V.NO_VOLUME,
        [_o(0, S.SCALAR, R.DIRECTIONAL, P.OWN)]),
    Id.DIRECTIONAL_MOVEMENT_PLUS: _d(
        Id.DIRECTIONAL_MOVEMENT_PLUS, "Welles Wilder", A.STATIC, I.BAR_INPUT, V.NO_VOLUME,
        [_o(0, S.SCALAR, R.DIRECTIONAL, P.OWN)]),
    Id.NORMALIZED_AVERAGE_TRUE_RANGE: _d(
        Id.NORMALIZED_AVERAGE_TRUE_RANGE, "Welles Wilder", A.STATIC, I.BAR_INPUT, V.NO_VOLUME,
        [_o(0, S.SCALAR, R.VOLATILITY, P.OWN)]),
    Id.PARABOLIC_STOP_AND_REVERSE: _d(
        Id.PARABOLIC_STOP_AND_REVERSE, "Welles Wilder", A.STATIC, I.BAR_INPUT, V.NO_VOLUME,
        [_o(0, S.SCALAR, R.OVERLAY, P.PRICE)]),
    Id.RELATIVE_STRENGTH_INDEX: _d(
        Id.RELATIVE_STRENGTH_INDEX, "Welles Wilder", A.STATIC, I.SCALAR_INPUT, V.NO_VOLUME,
        [_o(0, S.SCALAR, R.BOUNDED_OSCILLATOR, P.OWN)]),
    Id.TRUE_RANGE: _d(
        Id.TRUE_RANGE, "Welles Wilder", A.STATIC, I.BAR_INPUT, V.NO_VOLUME,
        [_o(0, S.SCALAR, R.VOLATILITY, P.OWN)]),

    # ── custom ────────────────────────────────────────────────────────────

    Id.GOERTZEL_SPECTRUM: _d(
        Id.GOERTZEL_SPECTRUM, "Custom", A.STATIC, I.SCALAR_INPUT, V.NO_VOLUME,
        [_o(0, S.HEATMAP, R.SPECTRUM, P.OWN)]),
    Id.MAXIMUM_ENTROPY_SPECTRUM: _d(
        Id.MAXIMUM_ENTROPY_SPECTRUM, "Custom", A.STATIC, I.SCALAR_INPUT, V.NO_VOLUME,
        [_o(0, S.HEATMAP, R.SPECTRUM, P.OWN)]),
}


def descriptor_of(identifier: Id):
    """Returns the taxonomic descriptor for the given indicator identifier, or None."""
    return _descriptors.get(identifier)


def descriptors() -> dict[Id, Descriptor]:
    """Returns a copy of the full descriptor registry."""
    return dict(_descriptors)
