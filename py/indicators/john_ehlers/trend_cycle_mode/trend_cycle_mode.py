"""Ehlers' Trend-versus-Cycle Mode indicator."""

import math
from typing import Optional

from ...core.indicator import Indicator
from ...core.metadata import Metadata
from ...core.build_metadata import build_metadata, OutputText
from ...core.identifier import Identifier
from ...core.component_triple_mnemonic import component_triple_mnemonic
from ...core.output import Output
from ....entities.bar import Bar
from ....entities.quote import Quote
from ....entities.trade import Trade
from ....entities.scalar import Scalar
from ....entities.bar_component import BarComponent, DEFAULT_BAR_COMPONENT, bar_component_value
from ....entities.quote_component import QuoteComponent, DEFAULT_QUOTE_COMPONENT, quote_component_value
from ....entities.trade_component import TradeComponent, DEFAULT_TRADE_COMPONENT, trade_component_value
from ..hilbert_transformer.cycle_estimator_type import CycleEstimatorType
from ..hilbert_transformer.cycle_estimator_params import CycleEstimatorParams
from ..hilbert_transformer.estimator import new_cycle_estimator, estimator_moniker
from ..dominant_cycle.dominant_cycle import DominantCycle
from ..dominant_cycle.params import DominantCycleParams
from .params import TrendCycleModeParams

_DEG2RAD = math.pi / 180.0
_EPSILON = 1e-308


class TrendCycleMode(Indicator):
    """Ehlers' Trend-versus-Cycle Mode indicator.

    Wraps a DominantCycle and exposes eight outputs:
      - Value: +1 in trend mode, -1 in cycle mode.
      - IsTrendMode: 1 if trend, 0 otherwise.
      - IsCycleMode: 1 if cycle, 0 otherwise.
      - InstantaneousTrendLine: WMA-smoothed trend line.
      - SineWave: sin(phase * deg2rad).
      - SineWaveLead: sin((phase + 45) * deg2rad).
      - DominantCyclePeriod: smoothed dominant cycle period.
      - DominantCyclePhase: dominant cycle phase in degrees.
    """

    def __init__(self, dc: DominantCycle,
                 cycle_part_multiplier: float,
                 separation_percentage: float,
                 trend_line_smoothing_length: int,
                 coeff0: float, coeff1: float, coeff2: float, coeff3: float,
                 mnemonic: str, description: str,
                 mnemonic_trend: str, description_trend: str,
                 mnemonic_cycle: str, description_cycle: str,
                 mnemonic_itl: str, description_itl: str,
                 mnemonic_sine: str, description_sine: str,
                 mnemonic_sine_lead: str, description_sine_lead: str,
                 mnemonic_dcp: str, description_dcp: str,
                 mnemonic_dc_phase: str, description_dc_phase: str,
                 bar_func, quote_func, trade_func) -> None:
        self._dc = dc
        self._cycle_part_multiplier = cycle_part_multiplier
        self._separation_percentage = separation_percentage
        self._separation_factor = separation_percentage / 100.0
        self._trend_line_smoothing_length = trend_line_smoothing_length
        self._coeff0 = coeff0
        self._coeff1 = coeff1
        self._coeff2 = coeff2
        self._coeff3 = coeff3
        self._mnemonic = mnemonic
        self._description = description
        self._mnemonic_trend = mnemonic_trend
        self._description_trend = description_trend
        self._mnemonic_cycle = mnemonic_cycle
        self._description_cycle = description_cycle
        self._mnemonic_itl = mnemonic_itl
        self._description_itl = description_itl
        self._mnemonic_sine = mnemonic_sine
        self._description_sine = description_sine
        self._mnemonic_sine_lead = mnemonic_sine_lead
        self._description_sine_lead = description_sine_lead
        self._mnemonic_dcp = mnemonic_dcp
        self._description_dcp = description_dcp
        self._mnemonic_dc_phase = mnemonic_dc_phase
        self._description_dc_phase = description_dc_phase
        self._bar_func = bar_func
        self._quote_func = quote_func
        self._trade_func = trade_func
        max_period = dc.max_period()
        self._input = [0.0] * max_period
        self._input_length = max_period
        self._input_length_min1 = max_period - 1
        self._trendline = math.nan
        self._trend_average1 = 0.0
        self._trend_average2 = 0.0
        self._trend_average3 = 0.0
        self._sin_wave = math.nan
        self._sin_wave_lead = math.nan
        self._previous_dc_phase = 0.0
        self._previous_sine_lead_wave_difference = 0.0
        self._samples_in_trend = 0
        self._is_trend_mode = True
        self._primed = False

    @staticmethod
    def create(params: TrendCycleModeParams) -> 'TrendCycleMode':
        """Creates a new instance from parameters."""
        return _new_trend_cycle_mode(
            params.estimator_type, params.estimator_params,
            params.alpha_ema_period_additional,
            params.trend_line_smoothing_length,
            params.cycle_part_multiplier,
            params.separation_percentage,
            params.bar_component, params.quote_component, params.trade_component)

    @staticmethod
    def create_default() -> 'TrendCycleMode':
        """Creates a new instance with default parameters."""
        return _new_trend_cycle_mode(
            CycleEstimatorType.HOMODYNE_DISCRIMINATOR,
            CycleEstimatorParams(smoothing_length=4, alpha_ema_quadrature_in_phase=0.2,
                                 alpha_ema_period=0.2, warm_up_period=100),
            0.33, 4, 1.0, 1.5, None, None, None)

    def is_primed(self) -> bool:
        return self._primed

    def metadata(self) -> Metadata:
        return build_metadata(
            Identifier.TREND_CYCLE_MODE,
            self._mnemonic,
            self._description,
            [
                OutputText(self._mnemonic, self._description),
                OutputText(self._mnemonic_trend, self._description_trend),
                OutputText(self._mnemonic_cycle, self._description_cycle),
                OutputText(self._mnemonic_itl, self._description_itl),
                OutputText(self._mnemonic_sine, self._description_sine),
                OutputText(self._mnemonic_sine_lead, self._description_sine_lead),
                OutputText(self._mnemonic_dcp, self._description_dcp),
                OutputText(self._mnemonic_dc_phase, self._description_dc_phase),
            ],
        )

    def update(self, sample: float) -> tuple[float, float, float, float, float, float, float, float]:
        """Updates, returning (value, isTrend, isCycle, trendline, sine, sineLead, period, phase).

        Returns NaN for all outputs if not primed.
        """
        if math.isnan(sample):
            return (sample, sample, sample, sample, sample, sample, sample, sample)

        # Delegate to inner DominantCycle.
        _, period, phase = self._dc.update(sample)
        smoothed_price = self._dc.smoothed_price()

        self._push_input(sample)

        if self._primed:
            smoothed_period = period
            average = self._calculate_trend_average(smoothed_period)
            self._trendline = self._coeff0 * average + self._coeff1 * self._trend_average1 + \
                self._coeff2 * self._trend_average2 + self._coeff3 * self._trend_average3
            self._trend_average3 = self._trend_average2
            self._trend_average2 = self._trend_average1
            self._trend_average1 = average

            diff = self._calculate_sine_lead_wave_difference(phase)

            # Condition 1: cycle mode after SineWave vs SineWaveLead crossing.
            self._is_trend_mode = True

            if (diff > 0 and self._previous_sine_lead_wave_difference < 0) or \
               (diff < 0 and self._previous_sine_lead_wave_difference > 0):
                self._is_trend_mode = False
                self._samples_in_trend = 0

            self._previous_sine_lead_wave_difference = diff
            self._samples_in_trend += 1

            if self._samples_in_trend < 0.5 * smoothed_period:
                self._is_trend_mode = False

            # Condition 2: cycle mode if phase rate of change is between 2/3 and 1.5 of 360/period.
            phase_delta = phase - self._previous_dc_phase
            self._previous_dc_phase = phase

            if abs(smoothed_period) > _EPSILON:
                dc_rate = 360.0 / smoothed_period
                if phase_delta > (2.0 / 3.0) * dc_rate and phase_delta < 1.5 * dc_rate:
                    self._is_trend_mode = False

            # Condition 3: force trend mode if separation exceeds threshold.
            if abs(self._trendline) > _EPSILON and \
               abs((smoothed_price - self._trendline) / self._trendline) >= self._separation_factor:
                self._is_trend_mode = True

            return (self._mode(), self._is_trend_float(), self._is_cycle_float(),
                    self._trendline, self._sin_wave, self._sin_wave_lead, period, phase)

        if self._dc.is_primed():
            self._primed = True
            smoothed_period = period
            self._trendline = self._calculate_trend_average(smoothed_period)
            self._trend_average1 = self._trendline
            self._trend_average2 = self._trendline
            self._trend_average3 = self._trendline

            self._previous_dc_phase = phase
            self._previous_sine_lead_wave_difference = \
                self._calculate_sine_lead_wave_difference(phase)

            self._is_trend_mode = True
            self._samples_in_trend += 1

            if self._samples_in_trend < 0.5 * smoothed_period:
                self._is_trend_mode = False

            return (self._mode(), self._is_trend_float(), self._is_cycle_float(),
                    self._trendline, self._sin_wave, self._sin_wave_lead, period, phase)

        nan = math.nan
        return (nan, nan, nan, nan, nan, nan, nan, nan)

    def update_scalar(self, sample: Scalar) -> Output:
        return self._update_entity(sample.time, sample.value)

    def update_bar(self, sample: Bar) -> Output:
        return self._update_entity(sample.time, self._bar_func(sample))

    def update_quote(self, sample: Quote) -> Output:
        return self._update_entity(sample.time, self._quote_func(sample))

    def update_trade(self, sample: Trade) -> Output:
        return self._update_entity(sample.time, self._trade_func(sample))

    def _update_entity(self, time, sample: float) -> Output:
        value, trend, cycle, itl, sine, sine_lead, period, phase = self.update(sample)
        return [
            Scalar(time=time, value=value),
            Scalar(time=time, value=trend),
            Scalar(time=time, value=cycle),
            Scalar(time=time, value=itl),
            Scalar(time=time, value=sine),
            Scalar(time=time, value=sine_lead),
            Scalar(time=time, value=period),
            Scalar(time=time, value=phase),
        ]

    def _push_input(self, value: float) -> None:
        inp = self._input
        for i in range(self._input_length_min1, 0, -1):
            inp[i] = inp[i - 1]
        inp[0] = value

    def _calculate_trend_average(self, smoothed_period: float) -> float:
        length = int(math.floor(smoothed_period * self._cycle_part_multiplier + 0.5))
        if length > self._input_length:
            length = self._input_length
        elif length < 1:
            length = 1

        s = 0.0
        for i in range(length):
            s += self._input[i]
        return s / length

    def _calculate_sine_lead_wave_difference(self, phase: float) -> float:
        p = phase * _DEG2RAD
        self._sin_wave = math.sin(p)
        self._sin_wave_lead = math.sin(p + 45.0 * _DEG2RAD)
        return self._sin_wave - self._sin_wave_lead

    def _mode(self) -> float:
        return 1.0 if self._is_trend_mode else -1.0

    def _is_trend_float(self) -> float:
        return 1.0 if self._is_trend_mode else 0.0

    def _is_cycle_float(self) -> float:
        return 0.0 if self._is_trend_mode else 1.0


def _new_trend_cycle_mode(
    estimator_type: CycleEstimatorType,
    estimator_params: CycleEstimatorParams,
    alpha_ema_period_additional: float,
    trend_line_smoothing_length: int,
    cycle_part_multiplier: float,
    separation_percentage: float,
    bc: Optional[BarComponent],
    qc: Optional[QuoteComponent],
    tc: Optional[TradeComponent],
) -> TrendCycleMode:
    invalid = "invalid trend cycle mode parameters"

    if alpha_ema_period_additional <= 0.0 or alpha_ema_period_additional > 1.0:
        raise ValueError(f"{invalid}: \u03b1 for additional smoothing should be in range (0, 1]")

    if trend_line_smoothing_length < 2 or trend_line_smoothing_length > 4:
        raise ValueError(f"{invalid}: trend line smoothing length should be 2, 3, or 4")

    if cycle_part_multiplier <= 0.0 or cycle_part_multiplier > 10.0:
        raise ValueError(f"{invalid}: cycle part multiplier should be in range (0, 10]")

    if separation_percentage <= 0.0 or separation_percentage > 100.0:
        raise ValueError(f"{invalid}: separation percentage should be in range (0, 100]")

    # Default to BarMedianPrice (not framework default).
    bc_resolved = bc if bc is not None else BarComponent.MEDIAN
    qc_resolved = qc if qc is not None else DEFAULT_QUOTE_COMPONENT
    tc_resolved = tc if tc is not None else DEFAULT_TRADE_COMPONENT

    # Build the inner DominantCycle with explicit components.
    dc_params = DominantCycleParams(
        estimator_type=estimator_type,
        estimator_params=estimator_params,
        alpha_ema_period_additional=alpha_ema_period_additional,
        bar_component=bc_resolved,
        quote_component=qc_resolved,
        trade_component=tc_resolved,
    )

    dc = DominantCycle.create(dc_params)

    # Estimator moniker (same logic as HTITL / DominantCycle).
    estimator = new_cycle_estimator(estimator_type, estimator_params)
    est_moniker = ""
    if estimator_type != CycleEstimatorType.HOMODYNE_DISCRIMINATOR or \
       estimator_params.smoothing_length != 4 or \
       estimator_params.alpha_ema_quadrature_in_phase != 0.2 or \
       estimator_params.alpha_ema_period != 0.2:
        est_moniker = estimator_moniker(estimator_type, estimator)
        if est_moniker:
            est_moniker = ", " + est_moniker

    comp = component_triple_mnemonic(bc_resolved, qc_resolved, tc_resolved)

    # Format mnemonics. Note: Go uses %.3f%% which produces e.g. "1.500%".
    fmt_args = (alpha_ema_period_additional, trend_line_smoothing_length,
                cycle_part_multiplier, separation_percentage, est_moniker, comp)

    mn_value = f"tcm({alpha_ema_period_additional:.3f}, {trend_line_smoothing_length}, " \
               f"{cycle_part_multiplier:.3f}, {separation_percentage:.3f}%{est_moniker}{comp})"
    mn_trend = f"tcm-trend({alpha_ema_period_additional:.3f}, {trend_line_smoothing_length}, " \
               f"{cycle_part_multiplier:.3f}, {separation_percentage:.3f}%{est_moniker}{comp})"
    mn_cycle = f"tcm-cycle({alpha_ema_period_additional:.3f}, {trend_line_smoothing_length}, " \
               f"{cycle_part_multiplier:.3f}, {separation_percentage:.3f}%{est_moniker}{comp})"
    mn_itl = f"tcm-itl({alpha_ema_period_additional:.3f}, {trend_line_smoothing_length}, " \
             f"{cycle_part_multiplier:.3f}, {separation_percentage:.3f}%{est_moniker}{comp})"
    mn_sine = f"tcm-sine({alpha_ema_period_additional:.3f}, {trend_line_smoothing_length}, " \
              f"{cycle_part_multiplier:.3f}, {separation_percentage:.3f}%{est_moniker}{comp})"
    mn_sine_lead = f"tcm-sineLead({alpha_ema_period_additional:.3f}, {trend_line_smoothing_length}, " \
                   f"{cycle_part_multiplier:.3f}, {separation_percentage:.3f}%{est_moniker}{comp})"
    mn_dcp = f"dcp({alpha_ema_period_additional:.3f}{est_moniker}{comp})"
    mn_dc_phase = f"dcph({alpha_ema_period_additional:.3f}{est_moniker}{comp})"

    bar_func = bar_component_value(bc_resolved)
    quote_func = quote_component_value(qc_resolved)
    trade_func = trade_component_value(tc_resolved)

    # WMA coefficients.
    if trend_line_smoothing_length == 2:
        c0, c1, c2, c3 = 2.0 / 3.0, 1.0 / 3.0, 0.0, 0.0
    elif trend_line_smoothing_length == 3:
        c0, c1, c2, c3 = 3.0 / 6.0, 2.0 / 6.0, 1.0 / 6.0, 0.0
    else:  # 4
        c0, c1, c2, c3 = 4.0 / 10.0, 3.0 / 10.0, 2.0 / 10.0, 1.0 / 10.0

    return TrendCycleMode(
        dc, cycle_part_multiplier, separation_percentage, trend_line_smoothing_length,
        c0, c1, c2, c3,
        mn_value, f"Trend versus cycle mode {mn_value}",
        mn_trend, f"Trend versus cycle mode, is-trend flag {mn_trend}",
        mn_cycle, f"Trend versus cycle mode, is-cycle flag {mn_cycle}",
        mn_itl, f"Trend versus cycle mode instantaneous trend line {mn_itl}",
        mn_sine, f"Trend versus cycle mode sine wave {mn_sine}",
        mn_sine_lead, f"Trend versus cycle mode sine wave lead {mn_sine_lead}",
        mn_dcp, f"Dominant cycle period {mn_dcp}",
        mn_dc_phase, f"Dominant cycle phase {mn_dc_phase}",
        bar_func, quote_func, trade_func,
    )
