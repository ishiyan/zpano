"""Welles Wilder's Average True Range indicator."""

import math

from .params import AverageTrueRangeParams
from ..true_range.true_range import TrueRange
from ...core.indicator import Indicator
from ...core.identifier import Identifier
from ...core.metadata import Metadata
from ...core.build_metadata import build_metadata, OutputText
from ...core.output import Output
from ....entities.bar import Bar
from ....entities.quote import Quote
from ....entities.trade import Trade
from ....entities.scalar import Scalar


class AverageTrueRange(Indicator):
    """Welles Wilder's Average True Range indicator."""

    def __init__(self, p: AverageTrueRangeParams) -> None:
        length = p.length
        if length < 1:
            raise ValueError(f"invalid length {length}: must be >= 1")

        self._length = length
        self._last_index = length - 1
        self._stage = 0
        self._window_count = 0
        self._window = [0.0] * length if self._last_index > 0 else None
        self._window_sum = 0.0
        self._value = math.nan
        self._primed = False
        self._true_range = TrueRange()

    def is_primed(self) -> bool:
        return self._primed

    def metadata(self) -> Metadata:
        return build_metadata(
            Identifier.AVERAGE_TRUE_RANGE,
            "atr",
            "Average True Range",
            [OutputText("atr", "Average True Range")],
        )

    def update(self, close: float, high: float, low: float) -> float:
        if math.isnan(close) or math.isnan(high) or math.isnan(low):
            return math.nan

        tr_value = self._true_range.update(close, high, low)

        if self._last_index == 0:
            self._value = tr_value
            if self._stage == 0:
                self._stage += 1
            elif self._stage == 1:
                self._stage += 1
                self._primed = True
            return self._value

        if self._stage > 1:
            self._value *= self._last_index
            self._value += tr_value
            self._value /= self._length
            return self._value

        if self._stage == 1:
            self._window_sum += tr_value
            self._window[self._window_count] = tr_value
            self._window_count += 1

            if self._window_count == self._length:
                self._stage += 1
                self._primed = True
                self._value = self._window_sum / self._length

            if self._primed:
                return self._value
            return math.nan

        # stage == 0: first sample used by TR to store close
        self._stage += 1
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
