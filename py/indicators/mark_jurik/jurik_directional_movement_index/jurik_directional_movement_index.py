"""Jurik directional movement index indicator."""

import math
from typing import List, Any

from ...core.indicator import Indicator
from ...core.metadata import Metadata
from ...core.build_metadata import build_metadata, OutputText
from ...core.identifier import Identifier
from ...core.output import Output
from ....entities.bar import Bar
from ....entities.quote import Quote
from ....entities.trade import Trade
from ....entities.scalar import Scalar
from ..jurik_moving_average.jurik_moving_average import JurikMovingAverage
from ..jurik_moving_average.params import JurikMovingAverageParams
from .params import JurikDirectionalMovementIndexParams


class JurikDirectionalMovementIndex(Indicator):
    """Computes the Jurik directional movement index (DMX).

    It produces three output lines:
      - Bipolar: 100*(Plus-Minus)/(Plus+Minus)
      - Plus: JMA(upward) / JMA(TrueRange)
      - Minus: JMA(downward) / JMA(TrueRange)

    The internal JMA instances use phase=-100 (maximum lag, no overshoot).
    """

    def __init__(self, params: JurikDirectionalMovementIndexParams) -> None:
        length = params.length

        if length < 1:
            raise ValueError(
                "invalid jurik directional movement index parameters: "
                "length should be positive")

        jma_params = JurikMovingAverageParams(length=length, phase=-100)

        self._mnemonic = f"dmx({length})"
        self._description = f"Jurik directional movement index {self._mnemonic}"
        self._primed = False
        self._bar = 0
        self._prev_high = math.nan
        self._prev_low = math.nan
        self._prev_close = math.nan
        self._jma_plus = JurikMovingAverage(jma_params)
        self._jma_minus = JurikMovingAverage(JurikMovingAverageParams(length=length, phase=-100))
        self._jma_denom = JurikMovingAverage(JurikMovingAverageParams(length=length, phase=-100))
        self._plus_val = math.nan
        self._minus_val = math.nan
        self._bipolar_val = math.nan

    def is_primed(self) -> bool:
        return self._primed

    def metadata(self) -> Metadata:
        return build_metadata(
            Identifier.JURIK_DIRECTIONAL_MOVEMENT_INDEX,
            self._mnemonic,
            self._description,
            [
                OutputText(mnemonic=self._mnemonic + ":bipolar",
                           description=self._description + " bipolar"),
                OutputText(mnemonic=self._mnemonic + ":plus",
                           description=self._description + " plus"),
                OutputText(mnemonic=self._mnemonic + ":minus",
                           description=self._description + " minus"),
            ],
        )

    def update(self, high: float, low: float, close: float) -> tuple[float, float, float]:
        """Update the indicator given the next high, low, and close values."""
        warmup = 41
        epsilon = 0.00001
        hundred = 100.0

        self._bar += 1

        upward = 0.0
        downward = 0.0
        true_range = 0.0

        if self._bar >= 2:
            v1 = hundred * (high - self._prev_high)
            v2 = hundred * (self._prev_low - low)

            if v1 > v2 and v1 > 0:
                upward = v1
            if v2 > v1 and v2 > 0:
                downward = v2

        if self._bar >= 3:
            m1 = abs(high - low)
            m2 = abs(high - self._prev_close)
            m3 = abs(low - self._prev_close)
            true_range = max(m1, m2, m3)

        self._prev_high = high
        self._prev_low = low
        self._prev_close = close

        # Feed into JMA instances.
        numer_plus = self._jma_plus.update(upward)
        numer_minus = self._jma_minus.update(downward)
        denom = self._jma_denom.update(true_range)

        if self._bar <= warmup:
            self._bipolar_val = math.nan
            self._plus_val = math.nan
            self._minus_val = math.nan
            return math.nan, math.nan, math.nan

        self._primed = True

        # Compute Plus and Minus.
        if denom > epsilon:
            self._plus_val = numer_plus / denom
        else:
            self._plus_val = 0.0

        if denom > epsilon:
            self._minus_val = numer_minus / denom
        else:
            self._minus_val = 0.0

        # Compute Bipolar.
        s = self._plus_val + self._minus_val
        if s > epsilon:
            self._bipolar_val = hundred * (self._plus_val - self._minus_val) / s
        else:
            self._bipolar_val = 0.0

        return self._bipolar_val, self._plus_val, self._minus_val

    def update_bar(self, sample: Bar) -> Output:
        bipolar, plus, minus = self.update(sample.high, sample.low, sample.close)
        return [
            Scalar(sample.time, bipolar),
            Scalar(sample.time, plus),
            Scalar(sample.time, minus),
        ]

    def update_quote(self, sample: Quote) -> Output:
        bipolar, plus, minus = self.update(
            sample.ask_price, sample.bid_price,
            (sample.ask_price + sample.bid_price) / 2)
        return [
            Scalar(sample.time, bipolar),
            Scalar(sample.time, plus),
            Scalar(sample.time, minus),
        ]

    def update_trade(self, sample: Trade) -> Output:
        v = sample.price
        bipolar, plus, minus = self.update(v, v, v)
        return [
            Scalar(sample.time, bipolar),
            Scalar(sample.time, plus),
            Scalar(sample.time, minus),
        ]

    def update_scalar(self, sample: Scalar) -> Output:
        v = sample.value
        bipolar, plus, minus = self.update(v, v, v)
        return [
            Scalar(sample.time, bipolar),
            Scalar(sample.time, plus),
            Scalar(sample.time, minus),
        ]
