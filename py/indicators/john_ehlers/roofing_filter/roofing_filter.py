"""Ehler's Roofing Filter indicator."""

import math

from ...core.indicator import Indicator
from ...core.line_indicator import LineIndicator
from ...core.metadata import Metadata
from ...core.build_metadata import build_metadata, OutputText
from ...core.identifier import Identifier
from ...core.component_triple_mnemonic import component_triple_mnemonic
from ...core.output import Output
from ....entities.bar import Bar
from ....entities.quote import Quote
from ....entities.trade import Trade
from ....entities.scalar import Scalar
from ....entities.bar_component import BarComponent, bar_component_value
from ....entities.quote_component import QuoteComponent, DEFAULT_QUOTE_COMPONENT, quote_component_value
from ....entities.trade_component import TradeComponent, DEFAULT_TRADE_COMPONENT, trade_component_value
from .params import RoofingFilterParams


class RoofingFilter(Indicator):
    """Ehler's Roofing Filter.

    The Roofing Filter is comprised of a high-pass filter and a Super Smoother.
    Given the longest and the shortest cycle periods in bars,
    the high-pass filter passes cyclic components whose periods are shorter than the longest one,
    and the Super Smoother filter attenuates cycle periods shorter than the shortest one.

    Three flavours are available:
      - 1-pole high-pass filter (default)
      - 1-pole high-pass filter with zero-mean
      - 2-pole high-pass filter
    """

    def __init__(self, hp_coeff1: float, hp_coeff2: float, hp_coeff3: float,
                 ss_coeff1: float, ss_coeff2: float, ss_coeff3: float,
                 has_two_pole: bool, has_zero_mean: bool,
                 mnemonic: str, description: str,
                 bar_func, quote_func, trade_func) -> None:
        self._line = LineIndicator(mnemonic, description, bar_func, quote_func, trade_func, self.update)
        self._hp_coeff1 = hp_coeff1
        self._hp_coeff2 = hp_coeff2
        self._hp_coeff3 = hp_coeff3
        self._ss_coeff1 = ss_coeff1
        self._ss_coeff2 = ss_coeff2
        self._ss_coeff3 = ss_coeff3
        self._has_two_pole = has_two_pole
        self._has_zero_mean = has_zero_mean
        self._count = 0
        self._sample_previous = 0.0
        self._sample_previous2 = 0.0
        self._hp_previous = 0.0
        self._hp_previous2 = 0.0
        self._ss_previous = 0.0
        self._ss_previous2 = 0.0
        self._zm_previous = 0.0
        self._value = math.nan
        self._primed = False

    @staticmethod
    def create(params: RoofingFilterParams) -> 'RoofingFilter':
        """Creates a new RoofingFilter from parameters."""
        return _new_roofing_filter(params)

    def is_primed(self) -> bool:
        return self._primed

    def metadata(self) -> Metadata:
        return build_metadata(
            Identifier.ROOFING_FILTER,
            self._line.mnemonic,
            self._line.description,
            [OutputText(self._line.mnemonic, self._line.description)],
        )

    def update(self, sample: float) -> float:
        if math.isnan(sample):
            return sample

        if self._has_two_pole:
            return self._update_2pole(sample)

        return self._update_1pole(sample)

    def _update_1pole(self, sample: float) -> float:
        hp = 0.0
        ss = 0.0
        zm = 0.0

        if self._primed:
            hp = self._hp_coeff1 * (sample - self._sample_previous) + \
                self._hp_coeff2 * self._hp_previous
            ss = self._ss_coeff1 * (hp + self._hp_previous) + \
                self._ss_coeff2 * self._ss_previous + self._ss_coeff3 * self._ss_previous2

            if self._has_zero_mean:
                zm = self._hp_coeff1 * (ss - self._ss_previous) + \
                    self._hp_coeff2 * self._zm_previous
                self._value = zm
            else:
                self._value = ss
        else:
            self._count += 1

            if self._count == 1:
                hp = 0.0
                ss = 0.0
            else:
                hp = self._hp_coeff1 * (sample - self._sample_previous) + \
                    self._hp_coeff2 * self._hp_previous
                ss = self._ss_coeff1 * (hp + self._hp_previous) + \
                    self._ss_coeff2 * self._ss_previous + self._ss_coeff3 * self._ss_previous2

                if self._has_zero_mean:
                    zm = self._hp_coeff1 * (ss - self._ss_previous) + \
                        self._hp_coeff2 * self._zm_previous
                    if self._count == 5:
                        self._primed = True
                        self._value = zm
                elif self._count == 4:
                    self._primed = True
                    self._value = ss

        self._sample_previous = sample
        self._hp_previous = hp
        self._ss_previous2 = self._ss_previous
        self._ss_previous = ss

        if self._has_zero_mean:
            self._zm_previous = zm

        return self._value

    def _update_2pole(self, sample: float) -> float:
        hp = 0.0
        ss = 0.0

        if self._primed:
            hp = self._hp_coeff1 * (sample - 2 * self._sample_previous + self._sample_previous2) + \
                self._hp_coeff2 * self._hp_previous - self._hp_coeff3 * self._hp_previous2
            ss = self._ss_coeff1 * (hp + self._hp_previous) + \
                self._ss_coeff2 * self._ss_previous + self._ss_coeff3 * self._ss_previous2
            self._value = ss
        else:
            self._count += 1

            if self._count < 4:
                hp = 0.0
                ss = 0.0
            else:
                hp = self._hp_coeff1 * (sample - 2 * self._sample_previous + self._sample_previous2) + \
                    self._hp_coeff2 * self._hp_previous - self._hp_coeff3 * self._hp_previous2
                ss = self._ss_coeff1 * (hp + self._hp_previous) + \
                    self._ss_coeff2 * self._ss_previous + self._ss_coeff3 * self._ss_previous2

                if self._count == 5:
                    self._primed = True
                    self._value = ss

        self._sample_previous2 = self._sample_previous
        self._sample_previous = sample
        self._hp_previous2 = self._hp_previous
        self._hp_previous = hp
        self._ss_previous2 = self._ss_previous
        self._ss_previous = ss

        return self._value

    def update_scalar(self, sample: Scalar) -> Output:
        return self._line.update_scalar(sample)

    def update_bar(self, sample: Bar) -> Output:
        return self._line.update_bar(sample)

    def update_quote(self, sample: Quote) -> Output:
        return self._line.update_quote(sample)

    def update_trade(self, sample: Trade) -> Output:
        return self._line.update_trade(sample)


def _new_roofing_filter(p: RoofingFilterParams) -> RoofingFilter:
    invalid = "invalid roofing filter parameters"

    shortest = p.shortest_cycle_period
    if shortest < 2:
        raise ValueError(f"{invalid}: shortest cycle period should be greater than 1")

    longest = p.longest_cycle_period
    if longest <= shortest:
        raise ValueError(f"{invalid}: longest cycle period should be greater than shortest")

    # Default bar component is MedianPrice, not ClosePrice.
    bc = p.bar_component if p.bar_component is not None else BarComponent.MEDIAN
    qc = p.quote_component if p.quote_component is not None else DEFAULT_QUOTE_COMPONENT
    tc = p.trade_component if p.trade_component is not None else DEFAULT_TRADE_COMPONENT

    bar_func = bar_component_value(bc)
    quote_func = quote_component_value(qc)
    trade_func = trade_component_value(tc)

    # Calculate high-pass filter coefficients.
    if p.has_two_pole_highpass_filter:
        # 2-pole high-pass.
        angle = math.sqrt(2) / 2 * 2 * math.pi / longest
        cos_angle = math.cos(angle)
        alpha = (math.sin(angle) + cos_angle - 1) / cos_angle
        beta = 1 - alpha / 2
        hp_coeff1 = beta * beta
        beta2 = 1 - alpha
        hp_coeff2 = 2 * beta2
        hp_coeff3 = beta2 * beta2
    else:
        # 1-pole high-pass.
        angle = 2 * math.pi / longest
        cos_angle = math.cos(angle)
        alpha = (math.sin(angle) + cos_angle - 1) / cos_angle
        hp_coeff1 = 1 - alpha / 2
        hp_coeff2 = 1 - alpha
        hp_coeff3 = 0.0

    # Calculate super smoother coefficients.
    # Uses literal 1.414 (not math.sqrt(2)) to match C# reference.
    beta_ss = 1.414 * math.pi / shortest
    alpha_ss = math.exp(-beta_ss)
    ss_coeff2 = 2 * alpha_ss * math.cos(beta_ss)
    ss_coeff3 = -alpha_ss * alpha_ss
    ss_coeff1 = (1 - ss_coeff2 - ss_coeff3) / 2

    # Build mnemonic.
    poles = 2 if p.has_two_pole_highpass_filter else 1
    zm = "zm" if p.has_zero_mean and not p.has_two_pole_highpass_filter else ""
    mnemonic = f"roof{poles}hp{zm}({shortest}, {longest}{component_triple_mnemonic(bc, qc, tc)})"
    desc = f"Roofing Filter {mnemonic}"

    has_zero_mean = p.has_zero_mean and not p.has_two_pole_highpass_filter

    return RoofingFilter(
        hp_coeff1, hp_coeff2, hp_coeff3,
        ss_coeff1, ss_coeff2, ss_coeff3,
        p.has_two_pole_highpass_filter, has_zero_mean,
        mnemonic, desc, bar_func, quote_func, trade_func,
    )
