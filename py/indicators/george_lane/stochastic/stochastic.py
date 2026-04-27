import math
from typing import Tuple, List, Any

from .params import StochasticParams, MovingAverageType
from ...core.indicator import Indicator
from ...core.identifier import Identifier
from ...core.metadata import Metadata
from ...core.build_metadata import build_metadata, OutputText
from ....entities.bar import Bar
from ....entities.quote import Quote
from ....entities.trade import Trade
from ....entities.scalar import Scalar
from ...common.simple_moving_average.simple_moving_average import SimpleMovingAverage
from ...common.simple_moving_average.params import SimpleMovingAverageParams
from ...common.exponential_moving_average.exponential_moving_average import ExponentialMovingAverage
from ...common.exponential_moving_average.params import ExponentialMovingAverageLengthParams


class _Passthrough:
    """No-op smoother for period of 1."""

    def update(self, v: float) -> float:
        return v

    def is_primed(self) -> bool:
        return True


def _create_ma(ma_type: MovingAverageType, length: int,
               first_is_average: bool) -> Tuple[Any, str]:
    """Create a moving average smoother. Returns (smoother, label)."""
    if length < 2:
        return _Passthrough(), "SMA"

    if ma_type == MovingAverageType.EMA:
        ema = ExponentialMovingAverage.from_length(
            ExponentialMovingAverageLengthParams(
                length=length, first_is_average=first_is_average))
        return ema, "EMA"
    else:
        sma = SimpleMovingAverage(SimpleMovingAverageParams(length=length))
        return sma, "SMA"


class Stochastic(Indicator):
    """George Lane's Stochastic Oscillator.

    Produces three outputs: Fast-K, Slow-K, and Slow-D.
    Requires bar data (high, low, close).
    """

    def __init__(self, p: StochasticParams) -> None:
        if p.fast_k_length < 1:
            raise ValueError("invalid stochastic parameters: fast K length should be greater than 0")
        if p.slow_k_length < 1:
            raise ValueError("invalid stochastic parameters: slow K length should be greater than 0")
        if p.slow_d_length < 1:
            raise ValueError("invalid stochastic parameters: slow D length should be greater than 0")

        self._fast_k_length = p.fast_k_length
        self._high_buf = [0.0] * p.fast_k_length
        self._low_buf = [0.0] * p.fast_k_length
        self._buffer_index = 0
        self._count = 0

        slow_k_ma, slow_k_label = _create_ma(p.slow_k_ma_type, p.slow_k_length, p.first_is_average)
        slow_d_ma, slow_d_label = _create_ma(p.slow_d_ma_type, p.slow_d_length, p.first_is_average)
        self._slow_k_ma = slow_k_ma
        self._slow_d_ma = slow_d_ma

        self._fast_k = math.nan
        self._slow_k = math.nan
        self._slow_d = math.nan
        self._primed = False

        self._mnemonic = f"stoch({p.fast_k_length}/{slow_k_label}{p.slow_k_length}/{slow_d_label}{p.slow_d_length})"

    def is_primed(self) -> bool:
        return self._primed

    def metadata(self) -> Metadata:
        desc = f"Stochastic Oscillator {self._mnemonic}"
        return build_metadata(
            Identifier.STOCHASTIC,
            self._mnemonic,
            desc,
            [
                OutputText(f"{self._mnemonic} fastK", f"{desc} Fast-K"),
                OutputText(f"{self._mnemonic} slowK", f"{desc} Slow-K"),
                OutputText(f"{self._mnemonic} slowD", f"{desc} Slow-D"),
            ],
        )

    def update(self, close: float, high: float, low: float) -> Tuple[float, float, float]:
        """Update with close, high, low. Returns (fastK, slowK, slowD)."""
        if math.isnan(close) or math.isnan(high) or math.isnan(low):
            return math.nan, math.nan, math.nan

        # Store in circular buffer.
        self._high_buf[self._buffer_index] = high
        self._low_buf[self._buffer_index] = low
        self._buffer_index = (self._buffer_index + 1) % self._fast_k_length
        self._count += 1

        # Need at least fast_k_length bars.
        if self._count < self._fast_k_length:
            return self._fast_k, self._slow_k, self._slow_d

        # Find highest high and lowest low in window.
        hh = self._high_buf[0]
        ll = self._low_buf[0]
        for i in range(1, self._fast_k_length):
            if self._high_buf[i] > hh:
                hh = self._high_buf[i]
            if self._low_buf[i] < ll:
                ll = self._low_buf[i]

        # Calculate Fast-K.
        diff = hh - ll
        if diff > 0:
            self._fast_k = 100 * (close - ll) / diff
        else:
            self._fast_k = 0

        # Feed Fast-K to Slow-K smoother.
        self._slow_k = self._slow_k_ma.update(self._fast_k)

        # Feed Slow-K to Slow-D smoother (only when Slow-K MA is primed).
        if self._slow_k_ma.is_primed():
            self._slow_d = self._slow_d_ma.update(self._slow_k)
            if not self._primed and self._slow_d_ma.is_primed():
                self._primed = True

        return self._fast_k, self._slow_k, self._slow_d

    def update_scalar(self, sample: Scalar) -> List[Any]:
        v = sample.value
        fast_k, slow_k, slow_d = self.update(v, v, v)
        return [
            Scalar(time=sample.time, value=fast_k),
            Scalar(time=sample.time, value=slow_k),
            Scalar(time=sample.time, value=slow_d),
        ]

    def update_bar(self, sample: Bar) -> List[Any]:
        fast_k, slow_k, slow_d = self.update(sample.close, sample.high, sample.low)
        return [
            Scalar(time=sample.time, value=fast_k),
            Scalar(time=sample.time, value=slow_k),
            Scalar(time=sample.time, value=slow_d),
        ]

    def update_quote(self, sample: Quote) -> List[Any]:
        v = (sample.bid_price + sample.ask_price) / 2
        return self.update_scalar(Scalar(time=sample.time, value=v))

    def update_trade(self, sample: Trade) -> List[Any]:
        return self.update_scalar(Scalar(time=sample.time, value=sample.price))
