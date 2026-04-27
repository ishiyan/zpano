"""Welles Wilder's Normalized Average True Range indicator."""

import math

from .params import NormalizedAverageTrueRangeParams
from ..average_true_range.average_true_range import AverageTrueRange
from ..average_true_range.params import AverageTrueRangeParams
from ...core.indicator import Indicator
from ...core.identifier import Identifier
from ...core.metadata import Metadata
from ...core.build_metadata import build_metadata, OutputText
from ...core.output import Output
from ....entities.bar import Bar
from ....entities.quote import Quote
from ....entities.trade import Trade
from ....entities.scalar import Scalar


class NormalizedAverageTrueRange(Indicator):
    """Welles Wilder's Normalized Average True Range indicator.

    NATR = (ATR / close) * 100.
    If close == 0, returns 0 (not division by zero).
    """

    def __init__(self, p: NormalizedAverageTrueRangeParams) -> None:
        length = p.length
        if length < 1:
            raise ValueError(f"invalid length {length}: must be >= 1")

        self._length = length
        self._value = math.nan
        self._primed = False
        self._atr = AverageTrueRange(AverageTrueRangeParams(length=length))

    def is_primed(self) -> bool:
        return self._primed

    def metadata(self) -> Metadata:
        return build_metadata(
            Identifier.NORMALIZED_AVERAGE_TRUE_RANGE,
            "natr",
            "Normalized Average True Range",
            [OutputText("natr", "Normalized Average True Range")],
        )

    def update(self, close: float, high: float, low: float) -> float:
        if math.isnan(close) or math.isnan(high) or math.isnan(low):
            return math.nan

        atr_value = self._atr.update(close, high, low)

        if self._atr.is_primed():
            self._primed = True
            if close == 0:
                self._value = 0
            else:
                self._value = (atr_value / close) * 100

        if self._primed:
            return self._value
        return math.nan

    def update_sample(self, sample: float) -> float:
        return self.update(sample, sample, sample)

    def update_scalar(self, sample: Scalar) -> Output:
        v = sample.value
        return [Scalar(time=sample.time, value=self.update(v, v, v))]

    def update_bar(self, sample: Bar) -> Output:
        return [Scalar(time=sample.time, value=self.update(sample.close, sample.high, sample.low))]

    def update_quote(self, sample: Quote) -> Output:
        v = (sample.bid_price + sample.ask_price) / 2
        return [Scalar(time=sample.time, value=self.update(v, v, v))]

    def update_trade(self, sample: Trade) -> Output:
        v = sample.price
        return [Scalar(time=sample.time, value=self.update(v, v, v))]
