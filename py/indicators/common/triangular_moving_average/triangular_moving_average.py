"""Triangular moving average indicator."""

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
from .params import TriangularMovingAverageParams


class TriangularMovingAverage(Indicator):
    """Computes the triangular moving average (TRIMA).

    Equivalent to SMA of SMA, with more weight on the middle of the window.

    The indicator is not primed during the first l-1 updates.
    """

    def __init__(self, params: TriangularMovingAverageParams) -> None:
        length = params.length
        if length < 2:
            raise ValueError(
                "invalid triangular moving average parameters: length should be greater than 1")

        bc = params.bar_component if params.bar_component is not None else DEFAULT_BAR_COMPONENT
        qc = params.quote_component if params.quote_component is not None else DEFAULT_QUOTE_COMPONENT
        tc = params.trade_component if params.trade_component is not None else DEFAULT_TRADE_COMPONENT

        bar_func = bar_component_value(bc)
        quote_func = quote_component_value(qc)
        trade_func = trade_component_value(tc)

        mnemonic = f"trima({length}{component_triple_mnemonic(bc, qc, tc)})"
        description = f"Triangular moving average {mnemonic}"

        self._line = LineIndicator(mnemonic, description, bar_func, quote_func, trade_func, self.update)

        length_half = length >> 1
        l = 1 + length_half
        is_odd = length % 2 == 1

        if is_odd:
            factor = 1.0 / (l * l)
        else:
            factor = 1.0 / (length_half * l)
            length_half -= 1

        self._factor: float = factor
        self._numerator: float = 0.0
        self._numerator_sub: float = 0.0
        self._numerator_add: float = 0.0
        self._window: list[float] = [0.0] * length
        self._window_length: int = length
        self._window_length_half: int = length_half
        self._window_count: int = 0
        self._is_odd: bool = is_odd
        self._primed: bool = False

    def is_primed(self) -> bool:
        return self._primed

    def metadata(self) -> Metadata:
        return build_metadata(
            Identifier.TRIANGULAR_MOVING_AVERAGE,
            self._line.mnemonic,
            self._line.description,
            [OutputText(self._line.mnemonic, self._line.description)],
        )

    def update(self, sample: float) -> float:
        if math.isnan(sample):
            return sample

        temp = sample

        if self._primed:
            self._numerator -= self._numerator_sub
            self._numerator_sub -= self._window[0]

            j = self._window_length - 1
            for i in range(j):
                self._window[i] = self._window[i + 1]

            self._window[j] = temp
            temp = self._window[self._window_length_half]
            self._numerator_sub += temp

            if self._is_odd:
                self._numerator += self._numerator_add
                self._numerator_add -= temp
            else:
                self._numerator_add -= temp
                self._numerator += self._numerator_add

            temp = sample
            self._numerator_add += temp
            self._numerator += temp
        else:
            self._window[self._window_count] = temp
            self._window_count += 1

            if self._window_length > self._window_count:
                return math.nan

            for i in range(self._window_length_half, -1, -1):
                self._numerator_sub += self._window[i]
                self._numerator += self._numerator_sub

            for i in range(self._window_length_half + 1, self._window_length):
                self._numerator_add += self._window[i]
                self._numerator += self._numerator_add

            self._primed = True

        return self._numerator * self._factor

    def update_scalar(self, sample: Scalar) -> Output:
        return self._line.update_scalar(sample)

    def update_bar(self, sample: Bar) -> Output:
        return self._line.update_bar(sample)

    def update_quote(self, sample: Quote) -> Output:
        return self._line.update_quote(sample)

    def update_trade(self, sample: Trade) -> Output:
        return self._line.update_trade(sample)
