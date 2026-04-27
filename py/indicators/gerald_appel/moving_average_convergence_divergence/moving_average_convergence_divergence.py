"""Moving Average Convergence Divergence (MACD) indicator."""

import math
from typing import List, Any

from ...core.indicator import Indicator
from ...core.metadata import Metadata
from ...core.build_metadata import build_metadata, OutputText
from ...core.identifier import Identifier
from ...core.component_triple_mnemonic import component_triple_mnemonic
from ....entities.bar import Bar
from ....entities.quote import Quote
from ....entities.trade import Trade
from ....entities.scalar import Scalar
from ....entities.bar_component import BarComponent, DEFAULT_BAR_COMPONENT, bar_component_value
from ....entities.quote_component import QuoteComponent, DEFAULT_QUOTE_COMPONENT, quote_component_value
from ....entities.trade_component import TradeComponent, DEFAULT_TRADE_COMPONENT, trade_component_value
from ...common.simple_moving_average.simple_moving_average import SimpleMovingAverage
from ...common.simple_moving_average.params import SimpleMovingAverageParams
from ...common.exponential_moving_average.exponential_moving_average import ExponentialMovingAverage
from ...common.exponential_moving_average.params import ExponentialMovingAverageLengthParams
from .params import MovingAverageConvergenceDivergenceParams, MovingAverageType


def _new_ma(ma_type: MovingAverageType, length: int, first_is_average: bool):
    """Create a moving average updater."""
    if ma_type == MovingAverageType.SMA:
        return SimpleMovingAverage(SimpleMovingAverageParams(length=length))
    else:
        return ExponentialMovingAverage.from_length(
            ExponentialMovingAverageLengthParams(
                length=length, first_is_average=first_is_average))


def _ma_label(ma_type: MovingAverageType) -> str:
    if ma_type == MovingAverageType.SMA:
        return "SMA"
    return "EMA"


class MovingAverageConvergenceDivergence(Indicator):
    """Gerald Appel's MACD indicator.

    MACD is calculated by subtracting the slow moving average from the fast moving average.
    A signal line (moving average of MACD) and histogram (MACD minus signal) are also produced.

    The indicator produces three outputs:
      - MACD: fast MA - slow MA
      - Signal: MA of the MACD line
      - Histogram: MACD - Signal
    """

    def __init__(self, p: MovingAverageConvergenceDivergenceParams) -> None:
        fast_length = p.fast_length
        slow_length = p.slow_length
        signal_length = p.signal_length

        if fast_length < 2:
            raise ValueError(
                "invalid moving average convergence divergence parameters: "
                "fast length should be greater than 1")
        if slow_length < 2:
            raise ValueError(
                "invalid moving average convergence divergence parameters: "
                "slow length should be greater than 1")
        if signal_length < 1:
            raise ValueError(
                "invalid moving average convergence divergence parameters: "
                "signal length should be greater than 0")

        # Auto-swap fast/slow if needed (matches TaLib behavior).
        if slow_length < fast_length:
            fast_length, slow_length = slow_length, fast_length

        bc = p.bar_component if p.bar_component is not None else DEFAULT_BAR_COMPONENT
        qc = p.quote_component if p.quote_component is not None else DEFAULT_QUOTE_COMPONENT
        tc = p.trade_component if p.trade_component is not None else DEFAULT_TRADE_COMPONENT

        self._bar_func = bar_component_value(bc)
        self._quote_func = quote_component_value(qc)
        self._trade_func = trade_component_value(tc)

        # Default FirstIsAverage to true (TA-Lib compatible).
        first_is_average = True if p.first_is_average is None else p.first_is_average

        self._fast_ma = _new_ma(p.moving_average_type, fast_length, first_is_average)
        self._slow_ma = _new_ma(p.moving_average_type, slow_length, first_is_average)
        self._signal_ma = _new_ma(p.signal_moving_average_type, signal_length, first_is_average)

        self._fast_delay = slow_length - fast_length
        self._fast_count = 0

        self._macd_value = math.nan
        self._signal_value = math.nan
        self._histogram_value = math.nan
        self._primed = False

        # Build mnemonic.
        suffix = ""
        if p.moving_average_type != MovingAverageType.EMA or \
                p.signal_moving_average_type != MovingAverageType.EMA:
            suffix = f",{_ma_label(p.moving_average_type)},{_ma_label(p.signal_moving_average_type)}"

        self._mnemonic = f"macd({fast_length},{slow_length},{signal_length}" \
                         f"{suffix}{component_triple_mnemonic(bc, qc, tc)})"

    def is_primed(self) -> bool:
        return self._primed

    def metadata(self) -> Metadata:
        desc = f"Moving Average Convergence Divergence {self._mnemonic}"
        return build_metadata(
            Identifier.MOVING_AVERAGE_CONVERGENCE_DIVERGENCE,
            self._mnemonic,
            desc,
            [
                OutputText(f"{self._mnemonic} macd", f"{desc} MACD"),
                OutputText(f"{self._mnemonic} signal", f"{desc} Signal"),
                OutputText(f"{self._mnemonic} histogram", f"{desc} Histogram"),
            ],
        )

    def update(self, sample: float) -> tuple[float, float, float]:
        """Update with a scalar value. Returns (macd, signal, histogram)."""
        nan = math.nan

        if math.isnan(sample):
            return nan, nan, nan

        # Feed the slow MA every sample.
        slow = self._slow_ma.update(sample)

        # Delay the fast MA to align SMA seed windows.
        if self._fast_count < self._fast_delay:
            self._fast_count += 1
            fast = nan
        else:
            fast = self._fast_ma.update(sample)

        if math.isnan(fast) or math.isnan(slow):
            self._macd_value = nan
            self._signal_value = nan
            self._histogram_value = nan
            return nan, nan, nan

        macd = fast - slow
        self._macd_value = macd

        signal = self._signal_ma.update(macd)

        if math.isnan(signal):
            self._signal_value = nan
            self._histogram_value = nan
            return macd, nan, nan

        self._signal_value = signal
        histogram = macd - signal
        self._histogram_value = histogram
        self._primed = self._fast_ma.is_primed() and self._slow_ma.is_primed() \
            and self._signal_ma.is_primed()

        return macd, signal, histogram

    def update_scalar(self, sample: Scalar) -> List[Any]:
        macd, signal, histogram = self.update(sample.value)
        return [
            Scalar(time=sample.time, value=macd),
            Scalar(time=sample.time, value=signal),
            Scalar(time=sample.time, value=histogram),
        ]

    def update_bar(self, sample: Bar) -> List[Any]:
        v = self._bar_func(sample)
        return self.update_scalar(Scalar(time=sample.time, value=v))

    def update_quote(self, sample: Quote) -> List[Any]:
        v = self._quote_func(sample)
        return self.update_scalar(Scalar(time=sample.time, value=v))

    def update_trade(self, sample: Trade) -> List[Any]:
        v = self._trade_func(sample)
        return self.update_scalar(Scalar(time=sample.time, value=v))
