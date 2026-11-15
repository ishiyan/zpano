"""Ehlers' Center of Gravity Oscillator (COG)."""

from __future__ import annotations

import math
from typing import Any, Optional

from ....entities.bar import Bar
from ....entities.bar_component import (
    DEFAULT_BAR_COMPONENT,
    BarComponent,
    bar_component_value,
)
from ....entities.quote import Quote
from ....entities.quote_component import (
    DEFAULT_QUOTE_COMPONENT,
    QuoteComponent,
    quote_component_value,
)
from ....entities.scalar import Scalar
from ....entities.trade import Trade
from ....entities.trade_component import (
    DEFAULT_TRADE_COMPONENT,
    TradeComponent,
    trade_component_value,
)
from ...core.build_metadata import OutputText, build_metadata
from ...core.identifier import Identifier
from ...core.indicator import Indicator
from ...core.metadata import Metadata
from ...core.output import Output
from .params import (
    CenterOfGravityOscillatorParams,
    center_of_gravity_oscillator_mnemonic,
    center_of_gravity_oscillator_trigger_mnemonic,
    default_params,
)


def _real_number(value: object, name: str, allow_nan: bool = False) -> float:
    if isinstance(value, bool):
        raise ValueError(f"{name} must be a real number")
    try:
        number = float(value)  # type: ignore[arg-type]
    except (TypeError, ValueError, OverflowError) as error:
        raise ValueError(f"{name} must be a real number") from error
    if math.isinf(number) or (math.isnan(number) and not allow_nan):
        suffix = " or NaN" if allow_nan else ""
        raise ValueError(f"{name} must be finite{suffix}")
    return number


def _resolve_component(raw: object, enum_cls: type, default: Any, label: str) -> Any:
    if raw is None:
        return default
    if isinstance(raw, bool) or not isinstance(raw, int):
        raise ValueError(f"invalid {label}")
    try:
        return enum_cls(raw)
    except ValueError as error:
        raise ValueError(f"invalid {label}") from error


class CenterOfGravityOscillator(Indicator):
    """Ehlers' Center of Gravity Oscillator (COG).

    Supports both the May 2002 TASC / Mesa convention (``centered=False``,
    ``-Num / Denom``) and the 2004 book convention (``centered=True``,
    ``-Num / Denom + (length + 1) / 2``), exposing the oscillator value on
    observation ``length`` and its one-observation delayed trigger on
    observation ``length + 1``.
    """

    def __init__(
        self, params: Optional[CenterOfGravityOscillatorParams] = None
    ) -> None:
        source = default_params() if params is None else params
        invalid = "invalid center of gravity oscillator parameters"

        if type(source.length) is not int or source.length < 1:
            raise ValueError(f"{invalid}: length should be a positive integer")
        if type(source.centered) is not bool:
            raise ValueError(f"{invalid}: centered must be a bool")

        bar_comp = _resolve_component(
            source.bar_component, BarComponent, DEFAULT_BAR_COMPONENT, "bar_component"
        )
        quote_comp = _resolve_component(
            source.quote_component,
            QuoteComponent,
            DEFAULT_QUOTE_COMPONENT,
            "quote_component",
        )
        trade_comp = _resolve_component(
            source.trade_component,
            TradeComponent,
            DEFAULT_TRADE_COMPONENT,
            "trade_component",
        )

        self._length = source.length
        self._centered = source.centered
        self._offset = (source.length + 1) / 2.0 if source.centered else 0.0

        self._bar_func = bar_component_value(bar_comp)
        self._quote_func = quote_component_value(quote_comp)
        self._trade_func = trade_component_value(trade_comp)

        normalized_params = CenterOfGravityOscillatorParams(
            length=source.length,
            centered=source.centered,
            bar_component=(
                bar_comp if source.bar_component is not None else None
            ),
            quote_component=(
                quote_comp if source.quote_component is not None else None
            ),
            trade_component=(
                trade_comp if source.trade_component is not None else None
            ),
        )
        self._mnemonic = center_of_gravity_oscillator_mnemonic(normalized_params)
        self._mnemonic_trig = center_of_gravity_oscillator_trigger_mnemonic(
            normalized_params
        )
        self._description = f"Center of Gravity oscillator {self._mnemonic}"
        self._description_trig = (
            f"Center of Gravity trigger {self._mnemonic_trig}"
        )

        self._window: list[float] = [0.0] * self._length
        self._next = 0
        self._count = 0
        self._last_value = 0.0
        self._current_value = math.nan
        self._current_trigger = math.nan

    @staticmethod
    def create(
        params: CenterOfGravityOscillatorParams,
    ) -> "CenterOfGravityOscillator":
        """Creates a new CenterOfGravityOscillator from parameters."""
        return CenterOfGravityOscillator(params)

    @staticmethod
    def create_default() -> "CenterOfGravityOscillator":
        """Creates a new CenterOfGravityOscillator with Ehlers defaults."""
        return CenterOfGravityOscillator(default_params())

    def reset(self) -> None:
        """Restore startup state while preserving configuration."""
        self._window = [0.0] * self._length
        self._next = 0
        self._count = 0
        self._last_value = 0.0
        self._current_value = math.nan
        self._current_trigger = math.nan

    def length(self) -> int:
        """Return the configured observation window length."""
        return self._length

    def centered(self) -> bool:
        """Return whether the zero-centered book baseline is enabled."""
        return self._centered

    def value(self) -> float:
        """Return the latest accepted oscillator value (NaN before priming)."""
        return self._current_value

    def trigger(self) -> float:
        """Return the latest accepted trigger value (NaN until length+1)."""
        return self._current_trigger

    def is_primed(self) -> bool:
        """Return whether at least length finite observations have filled the window."""
        return self._count >= self._length

    def is_trigger_primed(self) -> bool:
        """Return whether a preceding full-window oscillator value exists."""
        return self._count > self._length

    def metadata(self) -> Metadata:
        return build_metadata(
            Identifier.CENTER_OF_GRAVITY_OSCILLATOR,
            self._mnemonic,
            self._description,
            [
                OutputText(self._mnemonic, self._description),
                OutputText(self._mnemonic_trig, self._description_trig),
            ],
        )

    def update(self, sample: float) -> float:
        """Consume one chronological scalar price and return the CG value."""
        price = _real_number(sample, "price", allow_nan=True)
        if math.isnan(price):
            return math.nan

        trigger = self._last_value if self._count >= self._length else math.nan

        self._window[self._next] = price
        self._next = (self._next + 1) % self._length
        if self._count <= self._length:
            self._count += 1
        if self._count < self._length:
            return math.nan

        numerator = 0.0
        denominator = 0.0
        for lag in range(self._length):
            index = (self._next - 1 - lag) % self._length
            item = self._window[index]
            numerator = numerator + (1 + lag) * item
            denominator = denominator + item

        if denominator != 0.0:
            self._last_value = -numerator / denominator + self._offset
        self._current_value = self._last_value
        self._current_trigger = trigger
        return self._current_value

    def update_scalar(self, sample: Scalar) -> Output:
        return self._update_entity(sample.time, sample.value)

    def update_bar(self, sample: Bar) -> Output:
        return self._update_entity(sample.time, self._bar_func(sample))

    def update_quote(self, sample: Quote) -> Output:
        return self._update_entity(sample.time, self._quote_func(sample))

    def update_trade(self, sample: Trade) -> Output:
        return self._update_entity(sample.time, self._trade_func(sample))

    def _update_entity(self, stamp: Any, sample: float) -> Output:
        cog = self.update(sample)
        trig = self._current_trigger if not math.isnan(cog) else math.nan
        return [
            Scalar(time=stamp, value=cog),
            Scalar(time=stamp, value=trig),
        ]
