"""Welles Wilder's Average Directional Movement Index (ADX)."""

import math

from .params import AverageDirectionalMovementIndexParams
from ..directional_movement_index.directional_movement_index import DirectionalMovementIndex
from ..directional_movement_index.params import DirectionalMovementIndexParams
from ...core.indicator import Indicator
from ...core.identifier import Identifier
from ...core.metadata import Metadata
from ...core.build_metadata import build_metadata, OutputText
from ...core.output import Output
from ....entities.bar import Bar
from ....entities.quote import Quote
from ....entities.trade import Trade
from ....entities.scalar import Scalar


class AverageDirectionalMovementIndex(Indicator):
    """Welles Wilder's Average Directional Movement Index (ADX).

    The average directional movement index smooths the directional movement index (DX)
    using Wilder's smoothing technique:

        Initial ADX = SMA of first `length` DX values
        Subsequent ADX = (previousADX * (length-1) + DX) / length
    """

    def __init__(self, p: AverageDirectionalMovementIndexParams) -> None:
        length = p.length
        if length < 1:
            raise ValueError(f"invalid length {length}: must be >= 1")

        self._length = length
        self._length_minus_one = float(length - 1)
        self._count = 0
        self._sum = 0.0
        self._primed = False
        self._value = math.nan
        self._dx = DirectionalMovementIndex(DirectionalMovementIndexParams(length=length))

    def is_primed(self) -> bool:
        return self._primed

    def metadata(self) -> Metadata:
        return build_metadata(
            Identifier.AVERAGE_DIRECTIONAL_MOVEMENT_INDEX,
            "adx",
            "Average Directional Movement Index",
            [
                OutputText("adx", "Average Directional Movement Index"),
                OutputText("dx", "Directional Movement Index"),
                OutputText("+di", "Directional Indicator Plus"),
                OutputText("-di", "Directional Indicator Minus"),
                OutputText("+dm", "Directional Movement Plus"),
                OutputText("-dm", "Directional Movement Minus"),
                OutputText("atr", "Average True Range"),
                OutputText("tr", "True Range"),
            ],
        )

    def update(self, close: float, high: float, low: float) -> float:
        if math.isnan(close) or math.isnan(high) or math.isnan(low):
            return math.nan

        dx_value = self._dx.update(close, high, low)

        if not self._dx.is_primed():
            return math.nan

        if self._primed:
            self._value = (self._value * self._length_minus_one + dx_value) / self._length
            return self._value

        self._count += 1
        self._sum += dx_value

        if self._count == self._length:
            self._value = self._sum / self._length
            self._primed = True
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
