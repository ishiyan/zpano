"""Welles Wilder's Directional Indicator Plus (+DI)."""

import math

from .params import DirectionalIndicatorPlusParams
from ..average_true_range.average_true_range import AverageTrueRange
from ..average_true_range.params import AverageTrueRangeParams
from ..directional_movement_plus.directional_movement_plus import DirectionalMovementPlus
from ..directional_movement_plus.params import DirectionalMovementPlusParams
from ...core.indicator import Indicator
from ...core.identifier import Identifier
from ...core.metadata import Metadata
from ...core.build_metadata import build_metadata, OutputText
from ...core.output import Output
from ....entities.bar import Bar
from ....entities.quote import Quote
from ....entities.trade import Trade
from ....entities.scalar import Scalar

_EPSILON = 1e-8


class DirectionalIndicatorPlus(Indicator):
    """Welles Wilder's Directional Indicator Plus (+DI).

    The directional indicator plus measures the percentage of the average true range
    that is attributable to upward movement. It is calculated as:

        +DI = 100 * +DM(n) / (ATR * length)

    where +DM(n) is the Wilder-smoothed directional movement plus and ATR is the
    average true range over the same length.
    """

    def __init__(self, p: DirectionalIndicatorPlusParams) -> None:
        length = p.length
        if length < 1:
            raise ValueError(f"invalid length {length}: must be >= 1")

        self._length = length
        self._value = math.nan
        self._atr = AverageTrueRange(AverageTrueRangeParams(length=length))
        self._dmp = DirectionalMovementPlus(DirectionalMovementPlusParams(length=length))

    def is_primed(self) -> bool:
        return self._atr.is_primed() and self._dmp.is_primed()

    def metadata(self) -> Metadata:
        return build_metadata(
            Identifier.DIRECTIONAL_INDICATOR_PLUS,
            "+di",
            "Directional Indicator Plus",
            [
                OutputText("+di", "Directional Indicator Plus"),
                OutputText("+dm", "Directional Movement Plus"),
                OutputText("atr", "Average True Range"),
                OutputText("tr", "True Range"),
            ],
        )

    def update(self, close: float, high: float, low: float) -> float:
        """Core update with close, high, low values."""
        if math.isnan(close) or math.isnan(high) or math.isnan(low):
            return math.nan

        atr_value = self._atr.update(close, high, low)
        dmp_value = self._dmp.update(high, low)

        if self._atr.is_primed() and self._dmp.is_primed():
            atr_scaled = atr_value * self._length

            if abs(atr_scaled) < _EPSILON:
                self._value = 0
            else:
                self._value = 100 * dmp_value / atr_scaled

            return self._value

        return math.nan

    def update_sample(self, sample: float) -> float:
        """Updates the indicator given a single sample value."""
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
