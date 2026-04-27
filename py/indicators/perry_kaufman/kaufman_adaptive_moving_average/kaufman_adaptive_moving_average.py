"""Kaufman Adaptive Moving Average indicator."""

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
from ....entities.bar_component import BarComponent, DEFAULT_BAR_COMPONENT, bar_component_value
from ....entities.quote_component import QuoteComponent, DEFAULT_QUOTE_COMPONENT, quote_component_value
from ....entities.trade_component import TradeComponent, DEFAULT_TRADE_COMPONENT, trade_component_value
from .params import (KaufmanAdaptiveMovingAverageLengthParams,
                     KaufmanAdaptiveMovingAverageSmoothingFactorParams)

_EPSILON = 0.00000001


class KaufmanAdaptiveMovingAverage(Indicator):
    """Perry Kaufman's Adaptive Moving Average (KAMA).

    KAMA_i = KAMA_{i-1} + sc * (P_i - KAMA_{i-1})
    where sc = (alpha_slowest + ER * (alpha_fastest - alpha_slowest))^2
    and ER = |P - P_L| / sum(|P_i - P_{i+1}|)
    """

    def __init__(self, efficiency_ratio_length: int,
                 alpha_fastest: float, alpha_slowest: float,
                 mnemonic: str, description: str,
                 bar_func, quote_func, trade_func) -> None:
        self._line = LineIndicator(mnemonic, description, bar_func, quote_func, trade_func, self.update)

        self._efficiency_ratio_length: int = efficiency_ratio_length
        self._window_count: int = 0
        self._window: list[float] = [0.0] * (efficiency_ratio_length + 1)
        self._absolute_delta: list[float] = [0.0] * (efficiency_ratio_length + 1)
        self._absolute_delta_sum: float = 0.0
        self._alpha_fastest: float = alpha_fastest
        self._alpha_slowest: float = alpha_slowest
        self._alpha_diff: float = alpha_fastest - alpha_slowest
        self._value: float = math.nan
        self._efficiency_ratio: float = math.nan
        self._primed: bool = False

    @staticmethod
    def from_length(params: KaufmanAdaptiveMovingAverageLengthParams) -> 'KaufmanAdaptiveMovingAverage':
        """Creates KAMA from length-based parameters."""
        invalid = "invalid Kaufman adaptive moving average parameters"

        if params.efficiency_ratio_length < 2:
            raise ValueError(f"{invalid}: efficiency ratio length should be larger than 1")
        if params.fastest_length < 2:
            raise ValueError(f"{invalid}: fastest smoothing length should be larger than 1")
        if params.slowest_length < 2:
            raise ValueError(f"{invalid}: slowest smoothing length should be larger than 1")

        bc = params.bar_component if params.bar_component is not None else DEFAULT_BAR_COMPONENT
        qc = params.quote_component if params.quote_component is not None else DEFAULT_QUOTE_COMPONENT
        tc = params.trade_component if params.trade_component is not None else DEFAULT_TRADE_COMPONENT

        alpha_fastest = 2.0 / (1 + params.fastest_length)
        alpha_slowest = 2.0 / (1 + params.slowest_length)

        mnemonic = f"kama({params.efficiency_ratio_length}, " \
                   f"{params.fastest_length}, " \
                   f"{params.slowest_length}" \
                   f"{component_triple_mnemonic(bc, qc, tc)})"
        description = f"Kaufman adaptive moving average {mnemonic}"

        bar_func = bar_component_value(bc)
        quote_func = quote_component_value(qc)
        trade_func = trade_component_value(tc)

        return KaufmanAdaptiveMovingAverage(
            params.efficiency_ratio_length,
            alpha_fastest, alpha_slowest,
            mnemonic, description,
            bar_func, quote_func, trade_func)

    @staticmethod
    def from_smoothing_factor(params: KaufmanAdaptiveMovingAverageSmoothingFactorParams) -> 'KaufmanAdaptiveMovingAverage':
        """Creates KAMA from smoothing-factor-based parameters."""
        invalid = "invalid Kaufman adaptive moving average parameters"

        if params.efficiency_ratio_length < 2:
            raise ValueError(f"{invalid}: efficiency ratio length should be larger than 1")
        if params.fastest_smoothing_factor < 0 or params.fastest_smoothing_factor > 1:
            raise ValueError(f"{invalid}: fastest smoothing factor should be in range [0, 1]")
        if params.slowest_smoothing_factor < 0 or params.slowest_smoothing_factor > 1:
            raise ValueError(f"{invalid}: slowest smoothing factor should be in range [0, 1]")

        bc = params.bar_component if params.bar_component is not None else DEFAULT_BAR_COMPONENT
        qc = params.quote_component if params.quote_component is not None else DEFAULT_QUOTE_COMPONENT
        tc = params.trade_component if params.trade_component is not None else DEFAULT_TRADE_COMPONENT

        fastest = params.fastest_smoothing_factor
        slowest = params.slowest_smoothing_factor

        if fastest < _EPSILON:
            fastest = _EPSILON
        if slowest < _EPSILON:
            slowest = _EPSILON

        mnemonic = f"kama({params.efficiency_ratio_length}, " \
                   f"{fastest:.4f}, " \
                   f"{slowest:.4f}" \
                   f"{component_triple_mnemonic(bc, qc, tc)})"
        description = f"Kaufman adaptive moving average {mnemonic}"

        bar_func = bar_component_value(bc)
        quote_func = quote_component_value(qc)
        trade_func = trade_component_value(tc)

        return KaufmanAdaptiveMovingAverage(
            params.efficiency_ratio_length,
            fastest, slowest,
            mnemonic, description,
            bar_func, quote_func, trade_func)

    def is_primed(self) -> bool:
        return self._primed

    def metadata(self) -> Metadata:
        return build_metadata(
            Identifier.KAUFMAN_ADAPTIVE_MOVING_AVERAGE,
            self._line.mnemonic,
            self._line.description,
            [OutputText(self._line.mnemonic, self._line.description)],
        )

    def update(self, sample: float) -> float:
        if math.isnan(sample):
            return sample

        erl = self._efficiency_ratio_length

        if self._primed:
            temp = abs(sample - self._window[erl])
            self._absolute_delta_sum += temp - self._absolute_delta[1]

            for i in range(erl):
                j = i + 1
                self._window[i] = self._window[j]
                self._absolute_delta[i] = self._absolute_delta[j]

            self._window[erl] = sample
            self._absolute_delta[erl] = temp
            delta = abs(sample - self._window[0])

            if self._absolute_delta_sum <= delta or self._absolute_delta_sum < _EPSILON:
                temp = 1.0
            else:
                temp = delta / self._absolute_delta_sum

            self._efficiency_ratio = temp
            temp = self._alpha_slowest + temp * self._alpha_diff
            self._value += (sample - self._value) * temp * temp

            return self._value
        else:
            self._window[self._window_count] = sample
            if 0 < self._window_count:
                temp = abs(sample - self._window[self._window_count - 1])
                self._absolute_delta[self._window_count] = temp
                self._absolute_delta_sum += temp

            if erl == self._window_count:
                self._primed = True
                delta = abs(sample - self._window[0])

                if self._absolute_delta_sum <= delta or self._absolute_delta_sum < _EPSILON:
                    temp = 1.0
                else:
                    temp = delta / self._absolute_delta_sum

                self._efficiency_ratio = temp
                temp = self._alpha_slowest + temp * self._alpha_diff
                self._value = self._window[erl - 1]
                self._value += (sample - self._value) * temp * temp

                return self._value
            else:
                self._window_count += 1

        return math.nan

    def update_scalar(self, sample: Scalar) -> Output:
        return self._line.update_scalar(sample)

    def update_bar(self, sample: Bar) -> Output:
        return self._line.update_bar(sample)

    def update_quote(self, sample: Quote) -> Output:
        return self._line.update_quote(sample)

    def update_trade(self, sample: Trade) -> Output:
        return self._line.update_trade(sample)
