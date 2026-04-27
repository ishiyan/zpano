"""Welles Wilder's Parabolic Stop And Reverse (SAR)."""

import math

from .params import ParabolicStopAndReverseParams
from ...core.line_indicator import LineIndicator
from ...core.identifier import Identifier
from ...core.metadata import Metadata
from ...core.build_metadata import build_metadata, OutputText
from ...core.output import Output
from ....entities.bar import Bar
from ....entities.bar_component import bar_component_value, BarComponent
from ....entities.quote import Quote
from ....entities.quote_component import quote_component_value, DEFAULT_QUOTE_COMPONENT
from ....entities.trade import Trade
from ....entities.trade_component import trade_component_value, DEFAULT_TRADE_COMPONENT
from ....entities.scalar import Scalar


class ParabolicStopAndReverse:
    """Welles Wilder's Parabolic Stop And Reverse (SAR).

    The Parabolic SAR provides potential entry and exit points. Positive output
    values indicate long positions, negative values indicate short positions.
    """

    def __init__(self, p: ParabolicStopAndReverseParams) -> None:
        # Resolve defaults (0 means use default).
        af_init_long = p.acceleration_init_long if p.acceleration_init_long != 0 else 0.02
        af_step_long = p.acceleration_long if p.acceleration_long != 0 else 0.02
        af_max_long = p.acceleration_max_long if p.acceleration_max_long != 0 else 0.20
        af_init_short = p.acceleration_init_short if p.acceleration_init_short != 0 else 0.02
        af_step_short = p.acceleration_short if p.acceleration_short != 0 else 0.02
        af_max_short = p.acceleration_max_short if p.acceleration_max_short != 0 else 0.20

        # Validate.
        if af_init_long < 0 or af_step_long < 0 or af_max_long < 0:
            raise ValueError("long acceleration factors must be non-negative")
        if af_init_short < 0 or af_step_short < 0 or af_max_short < 0:
            raise ValueError("short acceleration factors must be non-negative")
        if p.offset_on_reverse < 0:
            raise ValueError("offset on reverse must be non-negative")

        # Clamp.
        if af_init_long > af_max_long:
            af_init_long = af_max_long
        if af_step_long > af_max_long:
            af_step_long = af_max_long
        if af_init_short > af_max_short:
            af_init_short = af_max_short
        if af_step_short > af_max_short:
            af_step_short = af_max_short

        self._start_value = p.start_value
        self._offset_on_reverse = p.offset_on_reverse
        self._af_init_long = af_init_long
        self._af_step_long = af_step_long
        self._af_max_long = af_max_long
        self._af_init_short = af_init_short
        self._af_step_short = af_step_short
        self._af_max_short = af_max_short

        # State.
        self._count = 0
        self._is_long = False
        self._sar = 0.0
        self._ep = 0.0
        self._af_long = af_init_long
        self._af_short = af_init_short
        self._previous_high = 0.0
        self._previous_low = 0.0
        self._new_high = 0.0
        self._new_low = 0.0
        self._primed = False

        self.mnemonic = "sar()"
        self.description = "Parabolic Stop And Reverse sar()"

    def is_primed(self) -> bool:
        return self._primed

    def metadata(self) -> Metadata:
        return build_metadata(
            Identifier.PARABOLIC_STOP_AND_REVERSE,
            self.mnemonic,
            self.description,
            [
                OutputText(self.mnemonic, self.description),
            ],
        )

    def update(self, sample: float) -> float:
        """Update with a single scalar sample (high = low = sample)."""
        if math.isnan(sample):
            return math.nan
        return self.update_hl(sample, sample)

    def update_hl(self, high: float, low: float) -> float:
        """Update with high and low values."""
        if math.isnan(high) or math.isnan(low):
            return math.nan

        self._count += 1

        # First bar: store high/low, no output yet.
        if self._count == 1:
            self._new_high = high
            self._new_low = low
            return math.nan

        # Second bar: initialize SAR, EP, and direction.
        if self._count == 2:
            previous_high = self._new_high
            previous_low = self._new_low

            if self._start_value == 0:
                # Auto-detect direction.
                minus_dm = previous_low - low
                plus_dm = high - previous_high

                if minus_dm < 0:
                    minus_dm = 0
                if plus_dm < 0:
                    plus_dm = 0

                self._is_long = minus_dm <= plus_dm

                if self._is_long:
                    self._ep = high
                    self._sar = previous_low
                else:
                    self._ep = low
                    self._sar = previous_high
            elif self._start_value > 0:
                self._is_long = True
                self._ep = high
                self._sar = self._start_value
            else:
                self._is_long = False
                self._ep = low
                self._sar = abs(self._start_value)

            self._new_high = high
            self._new_low = low
            self._primed = True

        # Main SAR calculation (bars 2+).
        if self._count >= 2:
            self._previous_low = self._new_low
            self._previous_high = self._new_high
            self._new_low = low
            self._new_high = high

            if self._count == 2:
                self._previous_low = self._new_low
                self._previous_high = self._new_high

            if self._is_long:
                return self._update_long()
            return self._update_short()

        return math.nan

    def _update_long(self) -> float:
        # Switch to short if the low penetrates the SAR value.
        if self._new_low <= self._sar:
            self._is_long = False
            self._sar = self._ep

            if self._sar < self._previous_high:
                self._sar = self._previous_high
            if self._sar < self._new_high:
                self._sar = self._new_high

            if self._offset_on_reverse != 0.0:
                self._sar += self._sar * self._offset_on_reverse

            result = -self._sar

            # Reset short AF and set EP.
            self._af_short = self._af_init_short
            self._ep = self._new_low

            # Calculate the new SAR.
            self._sar = self._sar + self._af_short * (self._ep - self._sar)

            if self._sar < self._previous_high:
                self._sar = self._previous_high
            if self._sar < self._new_high:
                self._sar = self._new_high

            return result

        # No switch — output the current SAR.
        result = self._sar

        # Adjust AF and EP.
        if self._new_high > self._ep:
            self._ep = self._new_high
            self._af_long += self._af_step_long
            if self._af_long > self._af_max_long:
                self._af_long = self._af_max_long

        # Calculate the new SAR.
        self._sar = self._sar + self._af_long * (self._ep - self._sar)

        if self._sar > self._previous_low:
            self._sar = self._previous_low
        if self._sar > self._new_low:
            self._sar = self._new_low

        return result

    def _update_short(self) -> float:
        # Switch to long if the high penetrates the SAR value.
        if self._new_high >= self._sar:
            self._is_long = True
            self._sar = self._ep

            if self._sar > self._previous_low:
                self._sar = self._previous_low
            if self._sar > self._new_low:
                self._sar = self._new_low

            if self._offset_on_reverse != 0.0:
                self._sar -= self._sar * self._offset_on_reverse

            result = self._sar

            # Reset long AF and set EP.
            self._af_long = self._af_init_long
            self._ep = self._new_high

            # Calculate the new SAR.
            self._sar = self._sar + self._af_long * (self._ep - self._sar)

            if self._sar > self._previous_low:
                self._sar = self._previous_low
            if self._sar > self._new_low:
                self._sar = self._new_low

            return result

        # No switch — output the negated SAR.
        result = -self._sar

        # Adjust AF and EP.
        if self._new_low < self._ep:
            self._ep = self._new_low
            self._af_short += self._af_step_short
            if self._af_short > self._af_max_short:
                self._af_short = self._af_max_short

        # Calculate the new SAR.
        self._sar = self._sar + self._af_short * (self._ep - self._sar)

        if self._sar < self._previous_high:
            self._sar = self._previous_high
        if self._sar < self._new_high:
            self._sar = self._new_high

        return result

    def update_scalar(self, sample: Scalar) -> Output:
        return [Scalar(time=sample.time, value=self.update(sample.value))]

    def update_bar(self, sample: Bar) -> Output:
        return [Scalar(time=sample.time, value=self.update_hl(sample.high, sample.low))]

    def update_quote(self, sample: Quote) -> Output:
        v = (sample.bid_price + sample.ask_price) / 2
        return [Scalar(time=sample.time, value=self.update(v))]

    def update_trade(self, sample: Trade) -> Output:
        return [Scalar(time=sample.time, value=self.update(sample.price))]
