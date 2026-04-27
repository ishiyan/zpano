"""Welles Wilder's Directional Movement Index (DX)."""

import math

from .params import DirectionalMovementIndexParams
from ..directional_indicator_plus.directional_indicator_plus import DirectionalIndicatorPlus
from ..directional_indicator_plus.params import DirectionalIndicatorPlusParams
from ..directional_indicator_minus.directional_indicator_minus import DirectionalIndicatorMinus
from ..directional_indicator_minus.params import DirectionalIndicatorMinusParams
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


class DirectionalMovementIndex(Indicator):
    """Welles Wilder's Directional Movement Index (DX).

    The directional movement index measures the strength of a trend by comparing
    the positive and negative directional indicators. It is calculated as:

        DX = 100 * |+DI - -DI| / (+DI + -DI)
    """

    def __init__(self, p: DirectionalMovementIndexParams) -> None:
        length = p.length
        if length < 1:
            raise ValueError(f"invalid length {length}: must be >= 1")

        self._length = length
        self._value = math.nan
        self._di_plus = DirectionalIndicatorPlus(DirectionalIndicatorPlusParams(length=length))
        self._di_minus = DirectionalIndicatorMinus(DirectionalIndicatorMinusParams(length=length))

    def is_primed(self) -> bool:
        return self._di_plus.is_primed() and self._di_minus.is_primed()

    def metadata(self) -> Metadata:
        return build_metadata(
            Identifier.DIRECTIONAL_MOVEMENT_INDEX,
            "dx",
            "Directional Movement Index",
            [
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

        dip_value = self._di_plus.update(close, high, low)
        dim_value = self._di_minus.update(close, high, low)

        if self._di_plus.is_primed() and self._di_minus.is_primed():
            s = dip_value + dim_value

            if abs(s) < _EPSILON:
                self._value = 0
            else:
                self._value = 100 * abs(dip_value - dim_value) / s

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
