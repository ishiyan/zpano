"""Tushar Chande's Aroon indicator."""

import math

from .params import AroonParams
from ...core.indicator import Indicator
from ...core.identifier import Identifier
from ...core.metadata import Metadata
from ...core.build_metadata import build_metadata, OutputText
from ...core.output import Output
from ....entities.bar import Bar
from ....entities.quote import Quote
from ....entities.trade import Trade
from ....entities.scalar import Scalar


class Aroon(Indicator):
    """Tushar Chande's Aroon indicator.

    Measures the number of periods since the highest high and lowest low
    within a lookback window. Produces three outputs: Up, Down, Osc.

    Up = 100 * (length - periods_since_highest_high) / length
    Down = 100 * (length - periods_since_lowest_low) / length
    Osc = Up - Down
    """

    def __init__(self, params: AroonParams) -> None:
        length = params.length
        if length < 2:
            raise ValueError(f"invalid aroon parameters: length should be greater than 1")

        self._length = length
        self._factor = 100.0 / length
        window_size = length + 1

        self._high_buf = [0.0] * window_size
        self._low_buf = [0.0] * window_size
        self._buffer_index = 0
        self._count = 0

        self._highest_index = 0
        self._lowest_index = 0

        self._up = math.nan
        self._down = math.nan
        self._osc = math.nan
        self._primed = False

        self._mnemonic = f"aroon({length})"

    def is_primed(self) -> bool:
        return self._primed

    def metadata(self) -> Metadata:
        desc = "Aroon " + self._mnemonic
        return build_metadata(
            Identifier.AROON,
            self._mnemonic,
            desc,
            [
                OutputText(self._mnemonic + " up", desc + " Up"),
                OutputText(self._mnemonic + " down", desc + " Down"),
                OutputText(self._mnemonic + " osc", desc + " Oscillator"),
            ],
        )

    def update(self, high: float, low: float) -> tuple[float, float, float]:
        """Update with high and low values. Returns (up, down, osc)."""
        if math.isnan(high) or math.isnan(low):
            return math.nan, math.nan, math.nan

        length = self._length
        window_size = length + 1
        today = self._count

        pos = self._buffer_index
        self._high_buf[pos] = high
        self._low_buf[pos] = low
        self._buffer_index = (self._buffer_index + 1) % window_size
        self._count += 1

        if self._count < window_size:
            return self._up, self._down, self._osc

        trailing_index = today - length

        if self._count == window_size:
            # First time: scan entire window.
            self._highest_index = trailing_index
            self._lowest_index = trailing_index

            for i in range(trailing_index + 1, today + 1):
                buf_pos = i % window_size
                if self._high_buf[buf_pos] >= self._high_buf[self._highest_index % window_size]:
                    self._highest_index = i
                if self._low_buf[buf_pos] <= self._low_buf[self._lowest_index % window_size]:
                    self._lowest_index = i
        else:
            # Subsequent: optimized update.
            if self._highest_index < trailing_index:
                self._highest_index = trailing_index
                for i in range(trailing_index + 1, today + 1):
                    buf_pos = i % window_size
                    if self._high_buf[buf_pos] >= self._high_buf[self._highest_index % window_size]:
                        self._highest_index = i
            elif high >= self._high_buf[self._highest_index % window_size]:
                self._highest_index = today

            if self._lowest_index < trailing_index:
                self._lowest_index = trailing_index
                for i in range(trailing_index + 1, today + 1):
                    buf_pos = i % window_size
                    if self._low_buf[buf_pos] <= self._low_buf[self._lowest_index % window_size]:
                        self._lowest_index = i
            elif low <= self._low_buf[self._lowest_index % window_size]:
                self._lowest_index = today

        self._up = self._factor * (length - (today - self._highest_index))
        self._down = self._factor * (length - (today - self._lowest_index))
        self._osc = self._up - self._down

        if not self._primed:
            self._primed = True

        return self._up, self._down, self._osc

    def update_scalar(self, sample: Scalar) -> Output:
        v = sample.value
        up, down, osc = self.update(v, v)
        return [
            Scalar(time=sample.time, value=up),
            Scalar(time=sample.time, value=down),
            Scalar(time=sample.time, value=osc),
        ]

    def update_bar(self, sample: Bar) -> Output:
        up, down, osc = self.update(sample.high, sample.low)
        return [
            Scalar(time=sample.time, value=up),
            Scalar(time=sample.time, value=down),
            Scalar(time=sample.time, value=osc),
        ]

    def update_quote(self, sample: Quote) -> Output:
        v = (sample.bid_price + sample.ask_price) / 2
        up, down, osc = self.update(v, v)
        return [
            Scalar(time=sample.time, value=up),
            Scalar(time=sample.time, value=down),
            Scalar(time=sample.time, value=osc),
        ]

    def update_trade(self, sample: Trade) -> Output:
        v = sample.price
        up, down, osc = self.update(v, v)
        return [
            Scalar(time=sample.time, value=up),
            Scalar(time=sample.time, value=down),
            Scalar(time=sample.time, value=osc),
        ]
