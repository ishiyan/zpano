import math
from typing import List, Tuple

from .params import StochasticRelativeStrengthIndexParams, MovingAverageType
from .output import StochasticRelativeStrengthIndexOutput
from ...core.indicator import Indicator
from ...core.identifier import Identifier
from ...core.metadata import Metadata
from ...core.build_metadata import build_metadata, OutputText
from ...core.component_triple_mnemonic import component_triple_mnemonic
from ...welles_wilder.relative_strength_index.relative_strength_index import RelativeStrengthIndex
from ...welles_wilder.relative_strength_index.params import RelativeStrengthIndexParams
from ...common.simple_moving_average.simple_moving_average import SimpleMovingAverage
from ...common.simple_moving_average.params import SimpleMovingAverageParams
from ...common.exponential_moving_average.exponential_moving_average import ExponentialMovingAverage
from ...common.exponential_moving_average.params import ExponentialMovingAverageLengthParams
from ....entities.bar import Bar
from ....entities.bar_component import BarComponent, bar_component_value, DEFAULT_BAR_COMPONENT
from ....entities.quote import Quote
from ....entities.quote_component import QuoteComponent, quote_component_value, DEFAULT_QUOTE_COMPONENT
from ....entities.trade import Trade
from ....entities.trade_component import TradeComponent, trade_component_value, DEFAULT_TRADE_COMPONENT
from ....entities.scalar import Scalar


class _Passthrough:
    """No-op smoother for Fast-D period of 1."""

    def update(self, v: float) -> float:
        return v

    def is_primed(self) -> bool:
        return True


class StochasticRelativeStrengthIndex(Indicator):
    """Tushar Chande's Stochastic RSI.

    Applies the Stochastic oscillator formula to RSI values instead of price.
    Produces two outputs: Fast-K and Fast-D.
    """

    def __init__(self, params: StochasticRelativeStrengthIndexParams) -> None:
        if params.length < 2:
            raise ValueError("invalid stochastic relative strength index parameters: "
                             "length should be greater than 1")
        if params.fast_k_length < 1:
            raise ValueError("invalid stochastic relative strength index parameters: "
                             "fast K length should be greater than 0")
        if params.fast_d_length < 1:
            raise ValueError("invalid stochastic relative strength index parameters: "
                             "fast D length should be greater than 0")

        bc = params.bar_component if params.bar_component is not None else DEFAULT_BAR_COMPONENT
        qc = params.quote_component if params.quote_component is not None else DEFAULT_QUOTE_COMPONENT
        tc = params.trade_component if params.trade_component is not None else DEFAULT_TRADE_COMPONENT

        self._bar_func = bar_component_value(bc)
        self._quote_func = quote_component_value(qc)
        self._trade_func = trade_component_value(tc)

        # Internal RSI.
        self._rsi = RelativeStrengthIndex(RelativeStrengthIndexParams(length=params.length))

        # Circular buffer for RSI values.
        self._fast_k_length = params.fast_k_length
        self._rsi_buf = [0.0] * params.fast_k_length
        self._rsi_buffer_index = 0
        self._rsi_count = 0

        # Fast-D smoother.
        if params.fast_d_length < 2:
            self._fast_d_ma = _Passthrough()
            ma_label = "SMA"
        elif params.moving_average_type == MovingAverageType.EMA:
            ma_label = "EMA"
            self._fast_d_ma = ExponentialMovingAverage.from_length(
                ExponentialMovingAverageLengthParams(
                    length=params.fast_d_length,
                    first_is_average=params.first_is_average))
        else:
            ma_label = "SMA"
            self._fast_d_ma = SimpleMovingAverage(
                SimpleMovingAverageParams(length=params.fast_d_length))

        self._fast_k = math.nan
        self._fast_d = math.nan
        self._primed = False

        self._mnemonic = f"stochrsi({params.length}/{params.fast_k_length}/" \
            f"{ma_label}{params.fast_d_length}" \
            f"{component_triple_mnemonic(bc, qc, tc)})"
        self._description = f"Stochastic Relative Strength Index {self._mnemonic}"

    def is_primed(self) -> bool:
        return self._primed

    def metadata(self) -> Metadata:
        return build_metadata(
            Identifier.STOCHASTIC_RELATIVE_STRENGTH_INDEX,
            self._mnemonic,
            self._description,
            [
                OutputText(mnemonic=self._mnemonic + " fastK",
                           description=self._description + " Fast-K"),
                OutputText(mnemonic=self._mnemonic + " fastD",
                           description=self._description + " Fast-D"),
            ],
        )

    def update(self, sample: float) -> Tuple[float, float]:
        """Update the indicator and return (Fast-K, Fast-D)."""
        if math.isnan(sample):
            return math.nan, math.nan

        # Feed to internal RSI.
        rsi_value = self._rsi.update(sample)
        if math.isnan(rsi_value):
            return self._fast_k, self._fast_d

        # Store RSI value in circular buffer.
        self._rsi_buf[self._rsi_buffer_index] = rsi_value
        self._rsi_buffer_index = (self._rsi_buffer_index + 1) % self._fast_k_length
        self._rsi_count += 1

        # Need at least fast_k_length RSI values for stochastic calculation.
        if self._rsi_count < self._fast_k_length:
            return self._fast_k, self._fast_d

        # Find min and max of RSI values in the window.
        min_rsi = self._rsi_buf[0]
        max_rsi = self._rsi_buf[0]
        for i in range(1, self._fast_k_length):
            if self._rsi_buf[i] < min_rsi:
                min_rsi = self._rsi_buf[i]
            if self._rsi_buf[i] > max_rsi:
                max_rsi = self._rsi_buf[i]

        # Calculate Fast-K.
        diff = max_rsi - min_rsi
        if diff > 0:
            self._fast_k = 100.0 * (rsi_value - min_rsi) / diff
        else:
            self._fast_k = 0.0

        # Feed Fast-K to Fast-D smoother.
        self._fast_d = self._fast_d_ma.update(self._fast_k)

        if not self._primed and self._fast_d_ma.is_primed():
            self._primed = True

        return self._fast_k, self._fast_d

    def update_scalar(self, scalar: Scalar) -> List:
        fast_k, fast_d = self.update(scalar.value)
        return [
            Scalar(time=scalar.time, value=fast_k),
            Scalar(time=scalar.time, value=fast_d),
        ]

    def update_bar(self, bar: Bar) -> List:
        return self.update_scalar(Scalar(time=bar.time, value=self._bar_func(bar)))

    def update_quote(self, quote: Quote) -> List:
        return self.update_scalar(Scalar(time=quote.time, value=self._quote_func(quote)))

    def update_trade(self, trade: Trade) -> List:
        return self.update_scalar(Scalar(time=trade.time, value=self._trade_func(trade)))
