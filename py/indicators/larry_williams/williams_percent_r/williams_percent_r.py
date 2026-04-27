import math
from typing import List

from .output import WilliamsPercentROutput
from ...core.indicator import Indicator
from ...core.identifier import Identifier
from ...core.metadata import Metadata
from ...core.build_metadata import build_metadata, OutputText
from ....entities.bar import Bar
from ....entities.quote import Quote
from ....entities.trade import Trade
from ....entities.scalar import Scalar


_MNEMONIC = "willr"
_DESCRIPTION = "Williams %R"
_DEFAULT_LENGTH = 14
_MIN_LENGTH = 2


class WilliamsPercentR(Indicator):
    """Larry Williams' Williams %R momentum indicator.

    Williams %R reflects the level of the closing price relative to the
    highest high over a lookback period. The oscillation ranges from 0 to -100.

    %R = -100 * (HighestHigh - Close) / (HighestHigh - LowestLow)
    """

    def __init__(self, length: int) -> None:
        if length < _MIN_LENGTH:
            length = _DEFAULT_LENGTH

        self._length = length
        self._length_min_one = length - 1
        self._circular_index = 0
        self._circular_count = 0
        self._low_circular = [0.0] * length
        self._high_circular = [0.0] * length
        self._value = math.nan
        self._primed = False

    def is_primed(self) -> bool:
        return self._primed

    def metadata(self) -> Metadata:
        return build_metadata(
            Identifier.WILLIAMS_PERCENT_R,
            _MNEMONIC,
            _DESCRIPTION,
            [OutputText(mnemonic=_MNEMONIC, description=_DESCRIPTION)],
        )

    def update(self, close: float, high: float, low: float) -> float:
        """Update Williams %R given close, high, low values."""
        if math.isnan(close) or math.isnan(high) or math.isnan(low):
            return math.nan

        index = self._circular_index
        self._low_circular[index] = low
        self._high_circular[index] = high

        # Advance circular buffer index.
        self._circular_index += 1
        if self._circular_index > self._length_min_one:
            self._circular_index = 0

        if self._length > self._circular_count:
            if self._length_min_one == self._circular_count:
                # We have exactly `length` samples; compute for the first time.
                min_low = self._low_circular[index]
                max_high = self._high_circular[index]

                for i in range(self._length_min_one):
                    index -= 1
                    if self._low_circular[index] < min_low:
                        min_low = self._low_circular[index]
                    if self._high_circular[index] > max_high:
                        max_high = self._high_circular[index]

                diff = max_high - min_low
                if abs(diff) < 5e-324:  # math.SmallestNonzeroFloat64 equivalent
                    self._value = 0.0
                else:
                    self._value = -100.0 * (max_high - close) / diff

                self._primed = True

            self._circular_count += 1
            return self._value

        # Already primed, compute normally with wrapping.
        min_low = self._low_circular[index]
        max_high = self._high_circular[index]

        for i in range(self._length_min_one):
            if index == 0:
                index = self._length_min_one
            else:
                index -= 1

            if self._low_circular[index] < min_low:
                min_low = self._low_circular[index]
            if self._high_circular[index] > max_high:
                max_high = self._high_circular[index]

        diff = max_high - min_low
        if abs(diff) < 5e-324:
            self._value = 0.0
        else:
            self._value = -100.0 * (max_high - close) / diff

        return self._value

    def update_sample(self, sample: float) -> float:
        """Update using a single value as substitute for high, low, and close."""
        return self.update(sample, sample, sample)

    def update_bar(self, bar: Bar) -> List:
        v = self.update(bar.close, bar.high, bar.low)
        return [Scalar(time=bar.time, value=v)]

    def update_scalar(self, scalar: Scalar) -> List:
        v = self.update(scalar.value, scalar.value, scalar.value)
        return [Scalar(time=scalar.time, value=v)]

    def update_quote(self, quote: Quote) -> List:
        mid = (quote.bid_price + quote.ask_price) / 2
        v = self.update(mid, mid, mid)
        return [Scalar(time=quote.time, value=v)]

    def update_trade(self, trade: Trade) -> List:
        v = self.update(trade.price, trade.price, trade.price)
        return [Scalar(time=trade.time, value=v)]
