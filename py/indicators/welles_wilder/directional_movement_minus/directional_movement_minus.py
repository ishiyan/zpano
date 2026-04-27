"""Welles Wilder's Directional Movement Minus indicator."""

import math

from .params import DirectionalMovementMinusParams
from ...core.indicator import Indicator
from ...core.identifier import Identifier
from ...core.metadata import Metadata
from ...core.build_metadata import build_metadata, OutputText
from ...core.output import Output
from ....entities.bar import Bar
from ....entities.quote import Quote
from ....entities.trade import Trade
from ....entities.scalar import Scalar


class DirectionalMovementMinus(Indicator):
    """Welles Wilder's Directional Movement Minus indicator.

    The directional movement was developed in 1978 by Welles Wilder
    as an indication of trend strength.

    The calculation of the directional movement (+DM and -DM) is as follows:
      - UpMove = today's high - yesterday's high
      - DownMove = yesterday's low - today's low
      - if DownMove > UpMove and DownMove > 0, then -DM = DownMove, else -DM = 0

    When the length is greater than 1, Wilder's smoothing method is applied:
        Today's -DM(n) = Previous -DM(n) - Previous -DM(n)/n + Today's -DM(1)
    """

    def __init__(self, p: DirectionalMovementMinusParams) -> None:
        length = p.length
        if length < 1:
            raise ValueError(f"invalid length {length}: must be >= 1")

        self._length = length
        self._no_smoothing = length == 1
        self._count = 0
        self._previous_high = 0.0
        self._previous_low = 0.0
        self._value = math.nan
        self._accumulator = 0.0
        self._primed = False

    def is_primed(self) -> bool:
        return self._primed

    def metadata(self) -> Metadata:
        return build_metadata(
            Identifier.DIRECTIONAL_MOVEMENT_MINUS,
            "-dm",
            "Directional Movement Minus",
            [OutputText("-dm", "Directional Movement Minus")],
        )

    def update(self, high: float, low: float) -> float:
        """Core update with high and low values."""
        if math.isnan(high) or math.isnan(low):
            return math.nan

        if high < low:
            high, low = low, high

        if self._no_smoothing:
            if self._primed:
                delta_minus = self._previous_low - low
                delta_plus = high - self._previous_high

                if delta_minus > 0 and delta_plus < delta_minus:
                    self._value = delta_minus
                else:
                    self._value = 0
            else:
                if self._count > 0:
                    delta_minus = self._previous_low - low
                    delta_plus = high - self._previous_high

                    if delta_minus > 0 and delta_plus < delta_minus:
                        self._value = delta_minus
                    else:
                        self._value = 0

                    self._primed = True

                self._count += 1
        else:
            if self._primed:
                delta_minus = self._previous_low - low
                delta_plus = high - self._previous_high

                if delta_minus > 0 and delta_plus < delta_minus:
                    self._accumulator += -self._accumulator / self._length + delta_minus
                else:
                    self._accumulator += -self._accumulator / self._length

                self._value = self._accumulator
            else:
                if self._count > 0 and self._length >= self._count:
                    delta_minus = self._previous_low - low
                    delta_plus = high - self._previous_high

                    if self._length > self._count:
                        if delta_minus > 0 and delta_plus < delta_minus:
                            self._accumulator += delta_minus
                    else:
                        if delta_minus > 0 and delta_plus < delta_minus:
                            self._accumulator += -self._accumulator / self._length + delta_minus
                        else:
                            self._accumulator += -self._accumulator / self._length

                        self._value = self._accumulator
                        self._primed = True

                self._count += 1

        self._previous_low = low
        self._previous_high = high

        return self._value

    def update_sample(self, sample: float) -> float:
        """Updates the indicator given a single sample value."""
        return self.update(sample, sample)

    def update_scalar(self, sample: Scalar) -> Output:
        v = sample.value
        return [Scalar(time=sample.time, value=self.update(v, v))]

    def update_bar(self, sample: Bar) -> Output:
        return [Scalar(time=sample.time, value=self.update(sample.high, sample.low))]

    def update_quote(self, sample: Quote) -> Output:
        v = (sample.bid_price + sample.ask_price) / 2
        return [Scalar(time=sample.time, value=self.update(v, v))]

    def update_trade(self, sample: Trade) -> Output:
        v = sample.price
        return [Scalar(time=sample.time, value=self.update(v, v))]
