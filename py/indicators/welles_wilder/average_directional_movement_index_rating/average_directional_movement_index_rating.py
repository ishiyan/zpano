"""Welles Wilder's Average Directional Movement Index Rating (ADXR)."""

import math

from .params import AverageDirectionalMovementIndexRatingParams
from ..average_directional_movement_index.average_directional_movement_index import AverageDirectionalMovementIndex
from ..average_directional_movement_index.params import AverageDirectionalMovementIndexParams
from ...core.indicator import Indicator
from ...core.identifier import Identifier
from ...core.metadata import Metadata
from ...core.build_metadata import build_metadata, OutputText
from ...core.output import Output
from ....entities.bar import Bar
from ....entities.quote import Quote
from ....entities.trade import Trade
from ....entities.scalar import Scalar


class AverageDirectionalMovementIndexRating(Indicator):
    """Welles Wilder's Average Directional Movement Index Rating (ADXR).

    The average directional movement index rating averages the current ADX value with
    the ADX value from (length - 1) periods ago:

        ADXR = (ADX[current] + ADX[current - (length - 1)]) / 2
    """

    def __init__(self, p: AverageDirectionalMovementIndexRatingParams) -> None:
        length = p.length
        if length < 1:
            raise ValueError(f"invalid length {length}: must be >= 1")

        self._length = length
        self._buffer_size = length
        self._buffer = [0.0] * length
        self._buffer_index = 0
        self._buffer_count = 0
        self._primed = False
        self._value = math.nan
        self._adx = AverageDirectionalMovementIndex(AverageDirectionalMovementIndexParams(length=length))

    def is_primed(self) -> bool:
        return self._primed

    def metadata(self) -> Metadata:
        return build_metadata(
            Identifier.AVERAGE_DIRECTIONAL_MOVEMENT_INDEX_RATING,
            "adxr",
            "Average Directional Movement Index Rating",
            [
                OutputText("adxr", "Average Directional Movement Index Rating"),
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

        adx_value = self._adx.update(close, high, low)

        if not self._adx.is_primed():
            return math.nan

        # Store ADX value in circular buffer.
        self._buffer[self._buffer_index] = adx_value
        self._buffer_index = (self._buffer_index + 1) % self._buffer_size
        self._buffer_count += 1

        if self._buffer_count < self._buffer_size:
            return math.nan

        # The oldest value in the buffer is at buffer_index (since we just advanced it).
        old_adx = self._buffer[self._buffer_index % self._buffer_size]
        self._value = (adx_value + old_adx) / 2
        self._primed = True

        return self._value

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
