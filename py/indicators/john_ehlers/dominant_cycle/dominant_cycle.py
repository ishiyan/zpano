"""Ehler's Dominant Cycle indicator."""

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
from .params import DominantCycleParams


class DominantCycle(Indicator):
    """Ehler's Dominant Cycle.

    Computes the instantaneous cycle period and phase derived from a
    Hilbert transformer cycle estimator.

    Three outputs:
      - RawPeriod: the raw instantaneous cycle period from the HT estimator.
      - Period: EMA-smoothed dominant cycle period.
      - Phase: dominant cycle phase in degrees.
    """

    def __init__(self, htce, alpha_ema_period_additional: float,
                 mnemonic_raw_period: str, description_raw_period: str,
                 mnemonic_period: str, description_period: str,
                 mnemonic_phase: str, description_phase: str,
                 bar_func, quote_func, trade_func) -> None:
        self._htce = htce
        self._alpha = alpha_ema_period_additional
        self._one_min_alpha = 1.0 - alpha_ema_period_additional
        self._mnemonic_raw_period = mnemonic_raw_period
        self._description_raw_period = description_raw_period
        self._mnemonic_period = mnemonic_period
        self._description_period = description_period
        self._mnemonic_phase = mnemonic_phase
        self._description_phase = description_phase
        self._bar_func = bar_func
        self._quote_func = quote_func
        self._trade_func = trade_func
        max_period = htce.max_period()
        self._smoothed_input = [0.0] * max_period
        self._smoothed_input_len_min1 = max_period - 1
        self._smoothed_period = 0.0
        self._smoothed_phase = 0.0
        self._primed = False

    @staticmethod
    def create(params: DominantCycleParams) -> 'DominantCycle':
        """Creates a new DominantCycle from parameters."""
        return _new_dominant_cycle(
            params.estimator_type, params.estimator_params,
            params.alpha_ema_period_additional,
            params.bar_component, params.quote_component, params.trade_component)

    @staticmethod
    def create_default() -> 'DominantCycle':
        """Creates a new DominantCycle with default parameters."""
        return _new_dominant_cycle(
            CycleEstimatorType.HOMODYNE_DISCRIMINATOR,
            CycleEstimatorParams(smoothing_length=4, alpha_ema_quadrature_in_phase=0.2,
                                 alpha_ema_period=0.2, warm_up_period=100),
            0.33, None, None, None)

    def is_primed(self) -> bool:
        return self._primed

    def smoothed_price(self) -> float:
        """Returns the current WMA-smoothed price value."""
        if not self._primed:
            return math.nan
        return self._htce.smoothed()

    def max_period(self) -> int:
        """Returns the maximum cycle period."""
        return self._htce.max_period()

    def metadata(self) -> Metadata:
        return build_metadata(
            Identifier.DOMINANT_CYCLE,
            self._mnemonic_period,
            self._description_period,
            [
                OutputText(self._mnemonic_raw_period, self._description_raw_period),
                OutputText(self._mnemonic_period, self._description_period),
                OutputText(self._mnemonic_phase, self._description_phase),
            ],
        )

    def update(self, sample: float) -> tuple[float, float, float]:
        """Updates, returning (raw_period, period, phase). NaN if not primed."""
        if math.isnan(sample):
            return (sample, sample, sample)

        self._htce.update(sample)
        self._push_smoothed_input(self._htce.smoothed())

        if self._primed:
            self._smoothed_period = self._alpha * self._htce.period() + \
                self._one_min_alpha * self._smoothed_period
            self._calculate_smoothed_phase()
            return (self._htce.period(), self._smoothed_period, self._smoothed_phase)

        if self._htce.primed():
            self._primed = True
            self._smoothed_period = self._htce.period()
            self._calculate_smoothed_phase()
            return (self._htce.period(), self._smoothed_period, self._smoothed_phase)

        nan = math.nan
        return (nan, nan, nan)

    def update_scalar(self, sample: Scalar) -> Output:
        return self._update_entity(sample.time, sample.value)

    def update_bar(self, sample: Bar) -> Output:
        return self._update_entity(sample.time, self._bar_func(sample))

    def update_quote(self, sample: Quote) -> Output:
        return self._update_entity(sample.time, self._quote_func(sample))

    def update_trade(self, sample: Trade) -> Output:
        return self._update_entity(sample.time, self._trade_func(sample))

    def _update_entity(self, time, sample: float) -> Output:
        raw_period, period, phase = self.update(sample)
        return [
            Scalar(time=time, value=raw_period),
            Scalar(time=time, value=period),
            Scalar(time=time, value=phase),
        ]

    def _push_smoothed_input(self, value: float) -> None:
        si = self._smoothed_input
        for i in range(self._smoothed_input_len_min1, 0, -1):
            si[i] = si[i - 1]
        si[0] = value

    def _calculate_smoothed_phase(self) -> None:
        rad2deg = 180.0 / math.pi
        two_pi = 2.0 * math.pi
        epsilon = 0.01

        length = int(math.floor(self._smoothed_period + 0.5))
        if length > self._smoothed_input_len_min1:
            length = self._smoothed_input_len_min1

        real_part = 0.0
        imag_part = 0.0

        for i in range(length):
            temp = two_pi * i / length
            smoothed = self._smoothed_input[i]
            real_part += smoothed * math.sin(temp)
            imag_part += smoothed * math.cos(temp)

        previous = self._smoothed_phase

        # Use atan (not atan2) to match Go reference.
        if imag_part == 0.0:
            phase = previous
        else:
            phase = math.atan(real_part / imag_part) * rad2deg
            if math.isnan(phase) or math.isinf(phase):
                phase = previous

        if abs(imag_part) <= epsilon:
            if real_part > 0:
                phase += 90.0
            elif real_part < 0:
                phase -= 90.0

        # 90 degree reference shift.
        phase += 90.0
        # Compensate for one bar lag.
        phase += 360.0 / self._smoothed_period
        # Resolve phase ambiguity.
        if imag_part < 0:
            phase += 180.0
        # Cycle wraparound.
        if phase > 360.0:
            phase -= 360.0

        self._smoothed_phase = phase


def _new_dominant_cycle(
    estimator_type: CycleEstimatorType,
    estimator_params: CycleEstimatorParams,
    alpha_ema_period_additional: float,
    bc: Optional[BarComponent],
    qc: Optional[QuoteComponent],
    tc: Optional[TradeComponent],
) -> DominantCycle:
    invalid = "invalid dominant cycle parameters"

    if alpha_ema_period_additional <= 0.0 or alpha_ema_period_additional > 1.0:
        raise ValueError(f"{invalid}: \u03b1 for additional smoothing should be in range (0, 1]")

    estimator = new_cycle_estimator(estimator_type, estimator_params)

    # Build estimator moniker for non-default estimator configs.
    est_moniker = ""
    if estimator_type != CycleEstimatorType.HOMODYNE_DISCRIMINATOR or \
       estimator_params.smoothing_length != 4 or \
       estimator_params.alpha_ema_quadrature_in_phase != 0.2 or \
       estimator_params.alpha_ema_period != 0.2:
        est_moniker = estimator_moniker(estimator_type, estimator)
        if est_moniker:
            est_moniker = ", " + est_moniker

    # Resolve defaults.
    bc_resolved = bc if bc is not None else DEFAULT_BAR_COMPONENT
    qc_resolved = qc if qc is not None else DEFAULT_QUOTE_COMPONENT
    tc_resolved = tc if tc is not None else DEFAULT_TRADE_COMPONENT

    bar_func = bar_component_value(bc_resolved)
    quote_func = quote_component_value(qc_resolved)
    trade_func = trade_component_value(tc_resolved)

    comp = component_triple_mnemonic(bc_resolved, qc_resolved, tc_resolved)

    mn_raw = f"dcp-raw({alpha_ema_period_additional:.3f}{est_moniker}{comp})"
    mn_per = f"dcp({alpha_ema_period_additional:.3f}{est_moniker}{comp})"
    mn_pha = f"dcph({alpha_ema_period_additional:.3f}{est_moniker}{comp})"

    return DominantCycle(
        estimator, alpha_ema_period_additional,
        mn_raw, f"Dominant cycle raw period {mn_raw}",
        mn_per, f"Dominant cycle period {mn_per}",
        mn_pha, f"Dominant cycle phase {mn_pha}",
        bar_func, quote_func, trade_func,
    )
