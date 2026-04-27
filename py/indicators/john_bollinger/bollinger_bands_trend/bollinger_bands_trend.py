"""Bollinger Bands Trend indicator."""

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
from ...common.variance.variance import Variance
from ...common.variance.params import VarianceParams
from .params import BollingerBandsTrendParams, MovingAverageType


class _BBLine:
    """Internal Bollinger Band line (MA + Variance)."""

    def __init__(self, length: int, upper_multiplier: float, lower_multiplier: float,
                 is_unbiased: bool, ma_type: MovingAverageType, first_is_average: bool,
                 bc: BarComponent, qc: QuoteComponent, tc: TradeComponent) -> None:
        self._variance = Variance(VarianceParams(
            length=length,
            is_unbiased=is_unbiased,
            bar_component=bc,
            quote_component=qc,
            trade_component=tc,
        ))

        if ma_type == MovingAverageType.EMA:
            self._ma = ExponentialMovingAverage.from_length(
                ExponentialMovingAverageLengthParams(
                    length=length, first_is_average=first_is_average))
        else:
            self._ma = SimpleMovingAverage(
                SimpleMovingAverageParams(length=length))

        self._upper_multiplier = upper_multiplier
        self._lower_multiplier = lower_multiplier

    def update(self, sample: float) -> tuple[float, float, float, bool]:
        """Update and return (lower, middle, upper, primed)."""
        nan = math.nan

        middle = self._ma.update(sample)
        v = self._variance.update(sample)

        primed = self._ma.is_primed() and self._variance.is_primed()

        if math.isnan(middle) or math.isnan(v):
            return nan, nan, nan, primed

        stddev = math.sqrt(v)
        upper = middle + self._upper_multiplier * stddev
        lower = middle - self._lower_multiplier * stddev

        return lower, middle, upper, primed


class BollingerBandsTrend(Indicator):
    """John Bollinger's Bollinger Bands Trend indicator.

    BBTrend measures the difference between the widths of fast and slow Bollinger Bands
    relative to the fast middle band, indicating trend strength and direction.

    The indicator produces a single output:

        bbtrend = (|fastLower - slowLower| - |fastUpper - slowUpper|) / fastMiddle

    Reference:
        Bollinger, John (2002). Bollinger on Bollinger Bands. McGraw-Hill.
    """

    def __init__(self, p: BollingerBandsTrendParams) -> None:
        fast_length = p.fast_length
        slow_length = p.slow_length
        upper_multiplier = p.upper_multiplier
        lower_multiplier = p.lower_multiplier

        if upper_multiplier == 0:
            upper_multiplier = 2.0
        if lower_multiplier == 0:
            lower_multiplier = 2.0

        if fast_length < 2:
            raise ValueError(
                "invalid bollinger bands trend parameters: "
                "fast length should be greater than 1")

        if slow_length < 2:
            raise ValueError(
                "invalid bollinger bands trend parameters: "
                "slow length should be greater than 1")

        if slow_length <= fast_length:
            raise ValueError(
                "invalid bollinger bands trend parameters: "
                "slow length should be greater than fast length")

        is_unbiased = True if p.is_unbiased is None else p.is_unbiased

        bc = p.bar_component if p.bar_component is not None else DEFAULT_BAR_COMPONENT
        qc = p.quote_component if p.quote_component is not None else DEFAULT_QUOTE_COMPONENT
        tc = p.trade_component if p.trade_component is not None else DEFAULT_TRADE_COMPONENT

        self._bar_func = bar_component_value(bc)
        self._quote_func = quote_component_value(qc)
        self._trade_func = trade_component_value(tc)

        self._fast_bb = _BBLine(fast_length, upper_multiplier, lower_multiplier,
                                is_unbiased, p.moving_average_type, p.first_is_average,
                                bc, qc, tc)
        self._slow_bb = _BBLine(slow_length, upper_multiplier, lower_multiplier,
                                is_unbiased, p.moving_average_type, p.first_is_average,
                                bc, qc, tc)

        self._value = math.nan
        self._primed = False

        self._mnemonic = f"bbtrend({fast_length},{slow_length}," \
                         f"{upper_multiplier:.0f},{lower_multiplier:.0f}" \
                         f"{component_triple_mnemonic(bc, qc, tc)})"

    def is_primed(self) -> bool:
        return self._primed

    def metadata(self) -> Metadata:
        desc = f"Bollinger Bands Trend {self._mnemonic}"
        return build_metadata(
            Identifier.BOLLINGER_BANDS_TREND,
            self._mnemonic,
            desc,
            [
                OutputText(self._mnemonic, desc),
            ],
        )

    def update(self, sample: float) -> float:
        """Update with a scalar value and return the BBTrend value."""
        if math.isnan(sample):
            return math.nan

        fast_lower, fast_middle, fast_upper, fast_primed = self._fast_bb.update(sample)
        slow_lower, _, slow_upper, slow_primed = self._slow_bb.update(sample)

        self._primed = fast_primed and slow_primed

        if not self._primed or math.isnan(fast_middle) or \
           math.isnan(fast_lower) or math.isnan(slow_lower):
            self._value = math.nan
            return math.nan

        epsilon = 1e-10

        lower_diff = abs(fast_lower - slow_lower)
        upper_diff = abs(fast_upper - slow_upper)

        if abs(fast_middle) < epsilon:
            self._value = 0.0
            return 0.0

        result = (lower_diff - upper_diff) / fast_middle
        self._value = result

        return result

    def update_scalar(self, sample: Scalar) -> List[Any]:
        v = self.update(sample.value)
        return [Scalar(time=sample.time, value=v)]

    def update_bar(self, sample: Bar) -> List[Any]:
        v = self._bar_func(sample)
        return self.update_scalar(Scalar(time=sample.time, value=v))

    def update_quote(self, sample: Quote) -> List[Any]:
        v = self._quote_func(sample)
        return self.update_scalar(Scalar(time=sample.time, value=v))

    def update_trade(self, sample: Trade) -> List[Any]:
        v = self._trade_func(sample)
        return self.update_scalar(Scalar(time=sample.time, value=v))
