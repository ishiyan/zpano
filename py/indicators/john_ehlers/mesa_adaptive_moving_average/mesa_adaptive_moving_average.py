"""Ehlers' Mesa Adaptive Moving Average (MAMA) indicator."""

import math
from typing import Optional

from ...core.indicator import Indicator
from ...core.metadata import Metadata
from ...core.build_metadata import build_metadata, OutputText
from ...core.identifier import Identifier
from ...core.component_triple_mnemonic import component_triple_mnemonic
from ...core.output import Output
from ...core.outputs.band import Band
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
from .params import MesaAdaptiveMovingAverageLengthParams, MesaAdaptiveMovingAverageSmoothingFactorParams


_EPSILON = 0.00000001
_RAD2DEG = 180.0 / math.pi


class MesaAdaptiveMovingAverage(Indicator):
    """Ehlers' Mesa Adaptive Moving Average (MAMA).

    Three outputs: MAMA value, FAMA value, and Band (upper=MAMA, lower=FAMA).
    """

    def __init__(self, htce,
                 alpha_fast_limit: float, alpha_slow_limit: float,
                 mnemonic: str, description: str,
                 mnemonic_fama: str, description_fama: str,
                 mnemonic_band: str, description_band: str,
                 bar_func, quote_func, trade_func) -> None:
        self._htce = htce
        self._alpha_fast_limit = alpha_fast_limit
        self._alpha_slow_limit = alpha_slow_limit
        self._mnemonic = mnemonic
        self._description = description
        self._mnemonic_fama = mnemonic_fama
        self._description_fama = description_fama
        self._mnemonic_band = mnemonic_band
        self._description_band = description_band
        self._bar_func = bar_func
        self._quote_func = quote_func
        self._trade_func = trade_func
        self._previous_phase = 0.0
        self._mama = 0.0
        self._fama = 0.0
        self._is_phase_cached = False
        self._primed = False

    @staticmethod
    def from_length(params: MesaAdaptiveMovingAverageLengthParams) -> 'MesaAdaptiveMovingAverage':
        """Creates a new instance from length parameters."""
        return _new_mama(
            params.estimator_type, params.estimator_params,
            params.fast_limit_length, params.slow_limit_length,
            math.nan, math.nan,
            params.bar_component, params.quote_component, params.trade_component)

    @staticmethod
    def from_smoothing_factor(params: MesaAdaptiveMovingAverageSmoothingFactorParams) -> 'MesaAdaptiveMovingAverage':
        """Creates a new instance from smoothing factor parameters."""
        return _new_mama(
            params.estimator_type, params.estimator_params,
            0, 0,
            params.fast_limit_smoothing_factor, params.slow_limit_smoothing_factor,
            params.bar_component, params.quote_component, params.trade_component)

    @staticmethod
    def create_default() -> 'MesaAdaptiveMovingAverage':
        """Creates a new instance with default parameters."""
        return _new_mama(
            CycleEstimatorType.HOMODYNE_DISCRIMINATOR,
            CycleEstimatorParams(smoothing_length=4, alpha_ema_quadrature_in_phase=0.2,
                                 alpha_ema_period=0.2, warm_up_period=0),
            3, 39,
            math.nan, math.nan,
            None, None, None)

    def is_primed(self) -> bool:
        return self._primed

    def metadata(self) -> Metadata:
        return build_metadata(
            Identifier.MESA_ADAPTIVE_MOVING_AVERAGE,
            self._mnemonic,
            self._description,
            [
                OutputText(self._mnemonic, self._description),
                OutputText(self._mnemonic_fama, self._description_fama),
                OutputText(self._mnemonic_band, self._description_band),
            ],
        )

    def update(self, sample: float) -> float:
        """Updates, returning MAMA value. NaN if not primed."""
        if math.isnan(sample):
            return sample

        self._htce.update(sample)

        if self._primed:
            return self._calculate(sample)

        if self._htce.primed():
            if self._is_phase_cached:
                self._primed = True
                return self._calculate(sample)

            self._is_phase_cached = True
            self._previous_phase = self._calculate_phase()
            self._mama = sample
            self._fama = sample

        return math.nan

    def update_scalar(self, sample: Scalar) -> Output:
        return self._update_entity(sample.time, sample.value)

    def update_bar(self, sample: Bar) -> Output:
        return self._update_entity(sample.time, self._bar_func(sample))

    def update_quote(self, sample: Quote) -> Output:
        return self._update_entity(sample.time, self._quote_func(sample))

    def update_trade(self, sample: Trade) -> Output:
        return self._update_entity(sample.time, self._trade_func(sample))

    def _update_entity(self, time, sample: float) -> Output:
        mama = self.update(sample)
        fama = self._fama if not math.isnan(mama) else math.nan
        return [
            Scalar(time=time, value=mama),
            Scalar(time=time, value=fama),
            Band(time=time, upper=mama, lower=fama),
        ]

    def _calculate_phase(self) -> float:
        if self._htce.in_phase() == 0:
            return self._previous_phase

        phase = math.atan(self._htce.quadrature() / self._htce.in_phase()) * _RAD2DEG
        if not math.isnan(phase) and not math.isinf(phase):
            return phase

        return self._previous_phase

    def _calculate_mama(self, sample: float) -> float:
        phase = self._calculate_phase()

        phase_rate_of_change = self._previous_phase - phase
        self._previous_phase = phase

        if phase_rate_of_change < 1:
            phase_rate_of_change = 1

        alpha = min(max(self._alpha_fast_limit / phase_rate_of_change,
                        self._alpha_slow_limit), self._alpha_fast_limit)

        self._mama = alpha * sample + (1.0 - alpha) * self._mama

        return alpha

    def _calculate(self, sample: float) -> float:
        alpha = self._calculate_mama(sample) / 2
        self._fama = alpha * self._mama + (1.0 - alpha) * self._fama
        return self._mama


def _new_mama(
    estimator_type: CycleEstimatorType,
    estimator_params: CycleEstimatorParams,
    fast_limit_length: int, slow_limit_length: int,
    fast_limit_sf: float, slow_limit_sf: float,
    bc: Optional[BarComponent],
    qc: Optional[QuoteComponent],
    tc: Optional[TradeComponent],
) -> MesaAdaptiveMovingAverage:
    invalid = "invalid mesa adaptive moving average parameters"

    estimator = new_cycle_estimator(estimator_type, estimator_params)

    est_moniker = ""
    if estimator_type != CycleEstimatorType.HOMODYNE_DISCRIMINATOR or \
       estimator_params.smoothing_length != 4 or \
       estimator_params.alpha_ema_quadrature_in_phase != 0.2 or \
       estimator_params.alpha_ema_period != 0.2:
        est_moniker = estimator_moniker(estimator_type, estimator)
        if est_moniker:
            est_moniker = ", " + est_moniker

    # MAMA defaults to BarClosePrice (framework default).
    bc_resolved = bc if bc is not None else DEFAULT_BAR_COMPONENT
    qc_resolved = qc if qc is not None else DEFAULT_QUOTE_COMPONENT
    tc_resolved = tc if tc is not None else DEFAULT_TRADE_COMPONENT

    comp = component_triple_mnemonic(bc_resolved, qc_resolved, tc_resolved)

    if math.isnan(fast_limit_sf):
        # Length-based construction.
        if fast_limit_length < 2:
            raise ValueError(f"{invalid}: fast limit length should be larger than 1")
        if slow_limit_length < 2:
            raise ValueError(f"{invalid}: slow limit length should be larger than 1")

        fast_limit_sf = 2.0 / (1 + fast_limit_length)
        slow_limit_sf = 2.0 / (1 + slow_limit_length)

        mn = f"mama({fast_limit_length}, {slow_limit_length}{est_moniker}{comp})"
        mn_fama = f"fama({fast_limit_length}, {slow_limit_length}{est_moniker}{comp})"
        mn_band = f"mama-fama({fast_limit_length}, {slow_limit_length}{est_moniker}{comp})"
    else:
        # Smoothing-factor-based construction.
        if fast_limit_sf < 0 or fast_limit_sf > 1:
            raise ValueError(f"{invalid}: fast limit smoothing factor should be in range [0, 1]")
        if slow_limit_sf < 0 or slow_limit_sf > 1:
            raise ValueError(f"{invalid}: slow limit smoothing factor should be in range [0, 1]")

        if fast_limit_sf < _EPSILON:
            fast_limit_sf = _EPSILON
        if slow_limit_sf < _EPSILON:
            slow_limit_sf = _EPSILON

        mn = f"mama({fast_limit_sf:.3f}, {slow_limit_sf:.3f}{est_moniker}{comp})"
        mn_fama = f"fama({fast_limit_sf:.3f}, {slow_limit_sf:.3f}{est_moniker}{comp})"
        mn_band = f"mama-fama({fast_limit_sf:.3f}, {slow_limit_sf:.3f}{est_moniker}{comp})"

    bar_func = bar_component_value(bc_resolved)
    quote_func = quote_component_value(qc_resolved)
    trade_func = trade_component_value(tc_resolved)

    descr = "Mesa adaptive moving average "

    return MesaAdaptiveMovingAverage(
        estimator, fast_limit_sf, slow_limit_sf,
        mn, descr + mn,
        mn_fama, descr + mn_fama,
        mn_band, descr + mn_band,
        bar_func, quote_func, trade_func,
    )
