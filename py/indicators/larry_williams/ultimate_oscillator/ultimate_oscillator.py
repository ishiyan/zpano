"""Larry Williams' Ultimate Oscillator."""

import math

from .params import UltimateOscillatorParams
from ...core.indicator import Indicator
from ...core.identifier import Identifier
from ...core.metadata import Metadata
from ...core.build_metadata import build_metadata, OutputText
from ...core.output import Output
from ....entities.bar import Bar
from ....entities.quote import Quote
from ....entities.trade import Trade
from ....entities.scalar import Scalar

_WEIGHT1 = 4.0
_WEIGHT2 = 2.0
_WEIGHT3 = 1.0
_TOTAL_WEIGHT = _WEIGHT1 + _WEIGHT2 + _WEIGHT3  # 7.0


def _sort_three(a: int, b: int, c: int) -> tuple[int, int, int]:
    if a > b:
        a, b = b, a
    if b > c:
        b, c = c, b
    if a > b:
        a, b = b, a
    return a, b, c


class UltimateOscillator(Indicator):
    """Larry Williams' Ultimate Oscillator.

    Combines three time periods weighted 4:2:1 (shortest:medium:longest).
    UO = 100 * (4*avg1 + 2*avg2 + 1*avg3) / 7
    where avg = bpSum / trSum for each period.
    """

    def __init__(self, params: UltimateOscillatorParams) -> None:
        l1 = params.length1 if params.length1 != 0 else 7
        l2 = params.length2 if params.length2 != 0 else 14
        l3 = params.length3 if params.length3 != 0 else 28

        if l1 < 2:
            raise ValueError(f"length1 must be >= 2, got {l1}")
        if l2 < 2:
            raise ValueError(f"length2 must be >= 2, got {l2}")
        if l3 < 2:
            raise ValueError(f"length3 must be >= 2, got {l3}")

        s1, s2, s3 = _sort_three(l1, l2, l3)

        self._p1 = s1
        self._p2 = s2
        self._p3 = s3
        self._previous_close = math.nan

        self._bp_buffer = [0.0] * s3
        self._tr_buffer = [0.0] * s3
        self._buffer_index = 0

        self._bp_sum1 = 0.0
        self._bp_sum2 = 0.0
        self._bp_sum3 = 0.0
        self._tr_sum1 = 0.0
        self._tr_sum2 = 0.0
        self._tr_sum3 = 0.0

        self._count = 0
        self._primed = False

        self._mnemonic_str = f"ultosc({l1}, {l2}, {l3})"

    def is_primed(self) -> bool:
        return self._primed

    def metadata(self) -> Metadata:
        desc = "Ultimate Oscillator " + self._mnemonic_str
        return build_metadata(
            Identifier.ULTIMATE_OSCILLATOR,
            self._mnemonic_str,
            desc,
            [OutputText(self._mnemonic_str, desc)],
        )

    def update(self, close: float, high: float, low: float) -> float:
        """Update with close, high, low. Returns oscillator value."""
        if math.isnan(close) or math.isnan(high) or math.isnan(low):
            return math.nan

        # First bar: just store close.
        if math.isnan(self._previous_close):
            self._previous_close = close
            return math.nan

        # Calculate BP and TR.
        true_low = min(low, self._previous_close)
        bp = close - true_low

        tr = high - low
        d = abs(self._previous_close - high)
        if d > tr:
            tr = d
        d = abs(self._previous_close - low)
        if d > tr:
            tr = d

        self._previous_close = close
        self._count += 1

        p3 = self._p3

        # Remove trailing values BEFORE storing new value.
        if self._count > self._p1:
            old_index = (self._buffer_index - self._p1 + p3) % p3
            self._bp_sum1 -= self._bp_buffer[old_index]
            self._tr_sum1 -= self._tr_buffer[old_index]

        if self._count > self._p2:
            old_index = (self._buffer_index - self._p2 + p3) % p3
            self._bp_sum2 -= self._bp_buffer[old_index]
            self._tr_sum2 -= self._tr_buffer[old_index]

        if self._count > p3:
            old_index = (self._buffer_index - p3 + p3) % p3
            self._bp_sum3 -= self._bp_buffer[old_index]
            self._tr_sum3 -= self._tr_buffer[old_index]

        # Add to running sums.
        self._bp_sum1 += bp
        self._bp_sum2 += bp
        self._bp_sum3 += bp
        self._tr_sum1 += tr
        self._tr_sum2 += tr
        self._tr_sum3 += tr

        # Store in circular buffer.
        self._bp_buffer[self._buffer_index] = bp
        self._tr_buffer[self._buffer_index] = tr
        self._buffer_index = (self._buffer_index + 1) % p3

        if self._count < p3:
            return math.nan

        self._primed = True

        output = 0.0
        if self._tr_sum1 != 0:
            output += _WEIGHT1 * (self._bp_sum1 / self._tr_sum1)
        if self._tr_sum2 != 0:
            output += _WEIGHT2 * (self._bp_sum2 / self._tr_sum2)
        if self._tr_sum3 != 0:
            output += _WEIGHT3 * (self._bp_sum3 / self._tr_sum3)

        return 100.0 * (output / _TOTAL_WEIGHT)

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
