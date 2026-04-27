"""Welles Wilder's True Range indicator."""

import math

from .params import TrueRangeParams
from ...core.indicator import Indicator
from ...core.identifier import Identifier
from ...core.metadata import Metadata
from ...core.build_metadata import build_metadata, OutputText
from ...core.output import Output
from ....entities.bar import Bar
from ....entities.quote import Quote
from ....entities.trade import Trade
from ....entities.scalar import Scalar


class TrueRange(Indicator):
    """Welles Wilder's True Range indicator.

    The True Range is defined as the largest of:
      - the distance from today's high to today's low
      - the distance from yesterday's close to today's high
      - the distance from yesterday's close to today's low

    The first update stores the close and returns NaN (not primed).
    The indicator is primed from the second update onward.
    """

    def __init__(self, p: TrueRangeParams | None = None) -> None:
        self._previous_close = math.nan
        self._value = math.nan
        self._primed = False

    def is_primed(self) -> bool:
        return self._primed

    def metadata(self) -> Metadata:
        return build_metadata(
            Identifier.TRUE_RANGE,
            "tr",
            "True Range",
            [OutputText("tr", "True Range")],
        )

    def update(self, close: float, high: float, low: float) -> float:
        """Core update with close, high, low values."""
        if math.isnan(close) or math.isnan(high) or math.isnan(low):
            return math.nan

        if not self._primed:
            if math.isnan(self._previous_close):
                self._previous_close = close
                return math.nan
            self._primed = True

        greatest = high - low
        temp = abs(high - self._previous_close)
        if greatest < temp:
            greatest = temp
        temp = abs(low - self._previous_close)
        if greatest < temp:
            greatest = temp

        self._value = greatest
        self._previous_close = close
        return self._value

    def update_sample(self, sample: float) -> float:
        """Updates the indicator given a single sample value."""
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
