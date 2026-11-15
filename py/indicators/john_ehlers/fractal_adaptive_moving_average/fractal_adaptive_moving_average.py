"""Ehlers' Fractal Adaptive Moving Average (FRAMA) indicator."""

from __future__ import annotations

import math
from typing import Optional

from ....entities.bar import Bar
from ....entities.bar_component import DEFAULT_BAR_COMPONENT, bar_component_value
from ....entities.quote import Quote
from ....entities.quote_component import DEFAULT_QUOTE_COMPONENT, quote_component_value
from ....entities.scalar import Scalar
from ....entities.trade import Trade
from ....entities.trade_component import DEFAULT_TRADE_COMPONENT, trade_component_value
from ...core.build_metadata import OutputText, build_metadata
from ...core.identifier import Identifier
from ...core.indicator import Indicator
from ...core.metadata import Metadata
from ...core.output import Output
from .params import (
    FractalAdaptiveMovingAverageParams,
    default_params,
    fractal_adaptive_moving_average_dimension_mnemonic,
    fractal_adaptive_moving_average_mnemonic,
)

_LOG_TWO = math.log(2.0)


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


def _box_range(highs: list[float], lows: list[float], start: int, end: int) -> float:
    highest = highs[start]
    lowest = lows[start]
    for index in range(start + 1, end):
        if highs[index] > highest:
            highest = highs[index]
        if lows[index] < lowest:
            lowest = lows[index]
    span = highest - lowest
    if not math.isfinite(span):
        raise ValueError("price-range arithmetic overflowed")
    return span


class FractalAdaptiveMovingAverage(Indicator):
    """Ehlers' Fractal Adaptive Moving Average (FRAMA).

    Returns both the adaptive EMA value and the raw guarded 2005 range-based
    fractal-dimension estimate used to adapt it.
    """

    def __init__(
        self, params: Optional[FractalAdaptiveMovingAverageParams] = None
    ) -> None:
        source = default_params() if params is None else params

        if (
            type(source.length) is not int
            or source.length < 2
            or source.length % 2 != 0
        ):
            raise ValueError("length must be an even integer >= 2")

        slowest = _real_number(
            source.slowest_smoothing_factor, "slowest_smoothing_factor"
        )
        if not (0.0 < slowest <= 1.0):
            raise ValueError("slowest_smoothing_factor must satisfy 0 < A <= 1")

        if type(source.coefficient_mode) is not int or source.coefficient_mode not in (
            0,
            1,
        ):
            raise ValueError(
                "coefficient_mode must be 0 (article) or 1 (full precision)"
            )
        if source.coefficient_mode == 0 and slowest != 0.01:
            raise ValueError(
                "literal article mode requires slowest_smoothing_factor=0.01"
            )

        bar_comp = (
            source.bar_component
            if source.bar_component is not None
            else DEFAULT_BAR_COMPONENT
        )
        quote_comp = (
            source.quote_component
            if source.quote_component is not None
            else DEFAULT_QUOTE_COMPONENT
        )
        trade_comp = (
            source.trade_component
            if source.trade_component is not None
            else DEFAULT_TRADE_COMPONENT
        )

        self._length = source.length
        self._half_length = source.length // 2
        self._warmup_count = source.length - 1
        self._slowest_smoothing_factor = slowest
        self._coefficient_mode = source.coefficient_mode
        self._exponent_scale = (
            math.log(slowest) if source.coefficient_mode == 1 else -4.6
        )

        self._bar_func = bar_component_value(bar_comp)
        self._quote_func = quote_component_value(quote_comp)
        self._trade_func = trade_component_value(trade_comp)

        normalized_params = FractalAdaptiveMovingAverageParams(
            length=source.length,
            slowest_smoothing_factor=slowest,
            coefficient_mode=source.coefficient_mode,
            bar_component=source.bar_component,
            quote_component=source.quote_component,
            trade_component=source.trade_component,
        )
        self._mnemonic = fractal_adaptive_moving_average_mnemonic(normalized_params)
        self._mnemonic_fdim = fractal_adaptive_moving_average_dimension_mnemonic(
            normalized_params
        )
        prefix = "Fractal adaptive moving average "
        self._description = prefix + self._mnemonic
        self._description_fdim = prefix + self._mnemonic_fdim

        self._count = 0
        self._highs: list[float] = []
        self._lows: list[float] = []
        self._dimension_state = 0.0
        self._alpha_state = 1.0
        self._filter_state = 0.0
        self._current_value = math.nan
        self._current_dimension = math.nan
        self.reset()

    @staticmethod
    def create(
        params: Optional[FractalAdaptiveMovingAverageParams] = None,
    ) -> FractalAdaptiveMovingAverage:
        """Creates a new instance from parameters."""
        return FractalAdaptiveMovingAverage(params)

    @staticmethod
    def create_default() -> FractalAdaptiveMovingAverage:
        """Creates a new instance with default parameters."""
        return FractalAdaptiveMovingAverage(default_params())

    def reset(self) -> None:
        """Clears streaming histories and priming state while preserving configuration."""
        self._count = 0
        self._highs = []
        self._lows = []
        self._dimension_state = 0.0
        self._alpha_state = 1.0
        self._filter_state = 0.0
        self._current_value = math.nan
        self._current_dimension = math.nan

    def warmup_bars(self) -> int:
        """Returns the number of initial NaN outputs before priming."""
        return self._warmup_count

    def samples_seen(self) -> int:
        """Returns the number of accepted finite observations seen since start or reset."""
        return self._count

    def value(self) -> float:
        """Returns the latest accepted FRAMA value (NaN before priming)."""
        return self._current_value

    def fractal_dimension(self) -> float:
        """Returns the latest accepted raw fractal-dimension estimate (NaN before priming)."""
        return self._current_dimension

    def alpha(self) -> float:
        """Returns the latest diagnostic smoothing factor (NaN before priming)."""
        return self._alpha_state if self.is_primed() else math.nan

    def is_primed(self) -> bool:
        return self._count >= self._length

    def metadata(self) -> Metadata:
        return build_metadata(
            Identifier.FRACTAL_ADAPTIVE_MOVING_AVERAGE,
            self._mnemonic,
            self._description,
            [
                OutputText(self._mnemonic, self._description),
                OutputText(self._mnemonic_fdim, self._description_fdim),
            ],
        )

    def update(
        self,
        price: float,
        high: Optional[float] = None,
        low: Optional[float] = None,
    ) -> float:
        """Consumes one scalar or (price, high, low) sample and returns the FRAMA value."""
        checked_price = _real_number(price, "price", allow_nan=True)
        if (high is None) != (low is None):
            raise ValueError("supply both high and low, or neither")
        if high is None:
            checked_high = checked_price
            checked_low = checked_price
        else:
            checked_high = _real_number(high, "high", allow_nan=True)
            checked_low = _real_number(low, "low", allow_nan=True)

        if (
            math.isnan(checked_price)
            or math.isnan(checked_high)
            or math.isnan(checked_low)
        ):
            return math.nan
        if checked_high < checked_low:
            raise ValueError("high must be >= low")

        length = self._length
        highs = [checked_high] + self._highs[: length - 1]
        lows = [checked_low] + self._lows[: length - 1]
        next_count = self._count + 1
        dimension = self._dimension_state
        alpha = self._alpha_state

        if next_count >= length:
            n1 = _box_range(highs, lows, 0, self._half_length) / self._half_length
            n2 = _box_range(highs, lows, self._half_length, length) / self._half_length
            n3 = _box_range(highs, lows, 0, length) / length
            if n1 > 0.0 and n2 > 0.0 and n3 > 0.0:
                combined = n1 + n2
                if not math.isfinite(combined):
                    raise ValueError("combined box-range arithmetic overflowed")
                dimension = (math.log(combined) - math.log(n3)) / _LOG_TWO
                if not math.isfinite(dimension):
                    raise ValueError("dimension arithmetic overflowed")
            if dimension <= 1.0:
                alpha = 1.0
            else:
                alpha = math.exp(self._exponent_scale * (dimension - 1.0))
                alpha = min(1.0, max(self._slowest_smoothing_factor, alpha))

        if next_count <= length:
            next_value = checked_price
        else:
            next_value = alpha * checked_price + (1.0 - alpha) * self._filter_state
            if not math.isfinite(next_value):
                raise ValueError("FRAMA arithmetic overflowed")

        self._highs = highs
        self._lows = lows
        self._count = next_count
        self._dimension_state = dimension
        self._alpha_state = alpha
        self._filter_state = next_value
        if next_count >= length:
            self._current_value = next_value
            self._current_dimension = dimension
            return next_value
        return math.nan

    def update_scalar(self, sample: Scalar) -> Output:
        extracted = sample.value
        return self._update_entity(sample.time, extracted, extracted, extracted)

    def update_bar(self, sample: Bar) -> Output:
        extracted = self._bar_func(sample)
        return self._update_entity(sample.time, extracted, sample.high, sample.low)

    def update_quote(self, sample: Quote) -> Output:
        extracted = self._quote_func(sample)
        return self._update_entity(
            sample.time, extracted, sample.ask_price, sample.bid_price
        )

    def update_trade(self, sample: Trade) -> Output:
        extracted = self._trade_func(sample)
        return self._update_entity(sample.time, extracted, extracted, extracted)

    def _update_entity(
        self, time, sample: float, sample_high: float, sample_low: float
    ) -> Output:
        frama = self.update(sample, sample_high, sample_low)
        fdim = self._current_dimension if not math.isnan(frama) else math.nan
        return [
            Scalar(time=time, value=frama),
            Scalar(time=time, value=fdim),
        ]
