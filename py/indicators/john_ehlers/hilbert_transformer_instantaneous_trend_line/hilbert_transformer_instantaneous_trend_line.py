"""Ehlers' Hilbert Transformer Instantaneous Trend Line indicator."""

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
from .params import HilbertTransformerInstantaneousTrendLineParams


class HilbertTransformerInstantaneousTrendLine(Indicator):
    """Ehlers' Instantaneous Trend Line built on a Hilbert transformer cycle estimator.

    Two outputs:
      - Value: the instantaneous trend line value (WMA of simple averages over
        windows tracking the smoothed dominant cycle period).
      - DominantCyclePeriod: the additionally EMA-smoothed dominant cycle period.
    """

    def __init__(self, htce,
                 alpha_ema_period_additional: float,
                 cycle_part_multiplier: float,
                 coeff0: float, coeff1: float, coeff2: float, coeff3: float,
                 mnemonic: str, description: str,
                 mnemonic_dcp: str, description_dcp: str,
                 bar_func, quote_func, trade_func) -> None:
        self._htce = htce
        self._alpha = alpha_ema_period_additional
        self._one_min_alpha = 1.0 - alpha_ema_period_additional
        self._cycle_part_multiplier = cycle_part_multiplier
        self._coeff0 = coeff0
        self._coeff1 = coeff1
        self._coeff2 = coeff2
        self._coeff3 = coeff3
        self._mnemonic = mnemonic
        self._description = description
        self._mnemonic_dcp = mnemonic_dcp
        self._description_dcp = description_dcp
        self._bar_func = bar_func
        self._quote_func = quote_func
        self._trade_func = trade_func
        max_period = htce.max_period()
        self._input = [0.0] * max_period
        self._input_length = max_period
        self._input_length_min1 = max_period - 1
        self._smoothed_period = 0.0
        self._value = 0.0
        self._average1 = 0.0
        self._average2 = 0.0
        self._average3 = 0.0
        self._primed = False

    @staticmethod
    def create(params: HilbertTransformerInstantaneousTrendLineParams) -> 'HilbertTransformerInstantaneousTrendLine':
        """Creates a new instance from parameters."""
        return _new_htitl(
            params.estimator_type, params.estimator_params,
            params.alpha_ema_period_additional,
            params.trend_line_smoothing_length,
            params.cycle_part_multiplier,
            params.bar_component, params.quote_component, params.trade_component)

    @staticmethod
    def create_default() -> 'HilbertTransformerInstantaneousTrendLine':
        """Creates a new instance with default parameters."""
        return _new_htitl(
            CycleEstimatorType.HOMODYNE_DISCRIMINATOR,
            CycleEstimatorParams(smoothing_length=4, alpha_ema_quadrature_in_phase=0.2,
                                 alpha_ema_period=0.2, warm_up_period=100),
            0.33, 4, 1.0, None, None, None)

    def is_primed(self) -> bool:
        return self._primed

    def metadata(self) -> Metadata:
        return build_metadata(
            Identifier.HILBERT_TRANSFORMER_INSTANTANEOUS_TREND_LINE,
            self._mnemonic,
            self._description,
            [
                OutputText(self._mnemonic, self._description),
                OutputText(self._mnemonic_dcp, self._description_dcp),
            ],
        )

    def update(self, sample: float) -> tuple[float, float]:
        """Updates, returning (value, period). NaN if not primed."""
        if math.isnan(sample):
            return (sample, sample)

        self._htce.update(sample)
        self._push_input(sample)

        if self._primed:
            self._smoothed_period = self._alpha * self._htce.period() + \
                self._one_min_alpha * self._smoothed_period
            average = self._calculate_average()
            self._value = self._coeff0 * average + self._coeff1 * self._average1 + \
                self._coeff2 * self._average2 + self._coeff3 * self._average3
            self._average3 = self._average2
            self._average2 = self._average1
            self._average1 = average
            return (self._value, self._smoothed_period)

        if self._htce.primed():
            self._primed = True
            self._smoothed_period = self._htce.period()
            average = self._calculate_average()
            self._value = average
            self._average1 = average
            self._average2 = average
            self._average3 = average
            return (self._value, self._smoothed_period)

        nan = math.nan
        return (nan, nan)

    def update_scalar(self, sample: Scalar) -> Output:
        return self._update_entity(sample.time, sample.value)

    def update_bar(self, sample: Bar) -> Output:
        return self._update_entity(sample.time, self._bar_func(sample))

    def update_quote(self, sample: Quote) -> Output:
        return self._update_entity(sample.time, self._quote_func(sample))

    def update_trade(self, sample: Trade) -> Output:
        return self._update_entity(sample.time, self._trade_func(sample))

    def _update_entity(self, time, sample: float) -> Output:
        value, period = self.update(sample)
        return [
            Scalar(time=time, value=value),
            Scalar(time=time, value=period),
        ]

    def _push_input(self, value: float) -> None:
        inp = self._input
        for i in range(self._input_length_min1, 0, -1):
            inp[i] = inp[i - 1]
        inp[0] = value

    def _calculate_average(self) -> float:
        length = int(math.floor(self._smoothed_period * self._cycle_part_multiplier + 0.5))
        if length > self._input_length:
            length = self._input_length
        elif length < 1:
            length = 1

        s = 0.0
        for i in range(length):
            s += self._input[i]
        return s / length


def _new_htitl(
    estimator_type: CycleEstimatorType,
    estimator_params: CycleEstimatorParams,
    alpha_ema_period_additional: float,
    trend_line_smoothing_length: int,
    cycle_part_multiplier: float,
    bc: Optional[BarComponent],
    qc: Optional[QuoteComponent],
    tc: Optional[TradeComponent],
) -> HilbertTransformerInstantaneousTrendLine:
    invalid = "invalid hilbert transformer instantaneous trend line parameters"

    if alpha_ema_period_additional <= 0.0 or alpha_ema_period_additional > 1.0:
        raise ValueError(f"{invalid}: \u03b1 for additional smoothing should be in range (0, 1]")

    if trend_line_smoothing_length < 2 or trend_line_smoothing_length > 4:
        raise ValueError(f"{invalid}: trend line smoothing length should be 2, 3, or 4")

    if cycle_part_multiplier <= 0.0 or cycle_part_multiplier > 10.0:
        raise ValueError(f"{invalid}: cycle part multiplier should be in range (0, 10]")

    # HTITL defaults to BarMedianPrice (not framework default).
    bc_resolved = bc if bc is not None else BarComponent.MEDIAN
    qc_resolved = qc if qc is not None else DEFAULT_QUOTE_COMPONENT
    tc_resolved = tc if tc is not None else DEFAULT_TRADE_COMPONENT

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

    mn = f"htitl({alpha_ema_period_additional:.3f}, {trend_line_smoothing_length}, " \
         f"{cycle_part_multiplier:.3f}{est_moniker}{comp})"
    mn_dcp = f"dcp({alpha_ema_period_additional:.3f}{est_moniker}{comp})"

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

    return HilbertTransformerInstantaneousTrendLine(
        estimator, alpha_ema_period_additional, cycle_part_multiplier,
        c0, c1, c2, c3,
        mn, f"Hilbert transformer instantaneous trend line {mn}",
        mn_dcp, f"Dominant cycle period {mn_dcp}",
        bar_func, quote_func, trade_func,
    )
