"""Weighted moving average indicator."""

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
from .params import WeightedMovingAverageParams


class WeightedMovingAverage(Indicator):
    """Computes the weighted moving average (WMA).

    WMAi = (l*Pi + (l-1)*Pi-1 + ... + Pi-l) / (l*(l+1)/2), where l is the length.

    The indicator is not primed during the first l-1 updates.
    """

    def __init__(self, params: WeightedMovingAverageParams) -> None:
        length = params.length
        if length < 2:
            raise ValueError(
                "invalid weighted moving average parameters: length should be greater than 1")

        bc = params.bar_component if params.bar_component is not None else DEFAULT_BAR_COMPONENT
        qc = params.quote_component if params.quote_component is not None else DEFAULT_QUOTE_COMPONENT
        tc = params.trade_component if params.trade_component is not None else DEFAULT_TRADE_COMPONENT

        bar_func = bar_component_value(bc)
        quote_func = quote_component_value(qc)
        trade_func = trade_component_value(tc)

        mnemonic = f"wma({length}{component_triple_mnemonic(bc, qc, tc)})"
        description = f"Weighted moving average {mnemonic}"

        self._line = LineIndicator(mnemonic, description, bar_func, quote_func, trade_func, self.update)

        self._window: list[float] = [0.0] * length
        self._window_sum: float = 0.0
        self._window_sub: float = 0.0
        self._divider: float = length * (length + 1) / 2.0
        self._window_length: int = length
        self._window_count: int = 0
        self._last_index: int = length - 1
        self._primed: bool = False

    def is_primed(self) -> bool:
        return self._primed

    def metadata(self) -> Metadata:
        return build_metadata(
            Identifier.WEIGHTED_MOVING_AVERAGE,
            self._line.mnemonic,
            self._line.description,
            [OutputText(self._line.mnemonic, self._line.description)],
        )

    def update(self, sample: float) -> float:
        if math.isnan(sample):
            return sample

        if self._primed:
            self._window_sum -= self._window_sub
            self._window_sum += sample * self._window_length
            self._window_sub -= self._window[0]
            self._window_sub += sample

            for i in range(self._last_index):
                self._window[i] = self._window[i + 1]

            self._window[self._last_index] = sample
        else:
            self._window[self._window_count] = sample
            self._window_sub += sample
            self._window_count += 1
            self._window_sum += sample * self._window_count

            if self._window_length > self._window_count:
                return math.nan

            self._primed = True

        return self._window_sum / self._divider

    def update_scalar(self, sample: Scalar) -> Output:
        return self._line.update_scalar(sample)

    def update_bar(self, sample: Bar) -> Output:
        return self._line.update_bar(sample)

    def update_quote(self, sample: Quote) -> Output:
        return self._line.update_quote(sample)

    def update_trade(self, sample: Trade) -> Output:
        return self._line.update_trade(sample)
