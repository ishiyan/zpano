"""Ehlers and Way's Zero-Lag Error-Correcting Exponential Moving Average (ZECEMA)."""

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
    ZeroLagErrorCorrectingExponentialMovingAverageLengthParams,
    ZeroLagErrorCorrectingExponentialMovingAverageParams,
    ZeroLagErrorCorrectingExponentialMovingAverageSmoothingFactorParams,
    default_length_params,
    default_smoothing_factor_params,
    equivalent_length_from_alpha,
    zero_lag_error_correcting_exponential_moving_average_ema_mnemonic,
    zero_lag_error_correcting_exponential_moving_average_mnemonic,
)

_LEAST_ERROR_SENTINEL = 1_000_000.0


def _parse_sample(sample: object) -> float:
    if isinstance(sample, bool):
        raise ValueError("price must be a finite real or NaN")
    try:
        price = float(sample)  # type: ignore[arg-type]
    except (TypeError, ValueError, OverflowError) as error:
        raise ValueError("price must be a finite real or NaN") from error
    if math.isinf(price):
        raise ValueError("price must be a finite real or NaN")
    return price


def _resolve_component(raw: object, enum_cls: type, default: Any, label: str) -> Any:
    if raw is None:
        return default
    if isinstance(raw, bool) or not isinstance(raw, int):
        raise ValueError(f"invalid {label}")
    try:
        return enum_cls(raw)
    except ValueError as error:
        raise ValueError(f"invalid {label}") from error


class ZeroLagErrorCorrectingExponentialMovingAverage(Indicator):
    """Ehlers and Way's Zero-Lag Error-Correcting Exponential Moving Average (ZECEMA).

    Follows the November 2010 TASC EasyLanguage listing ("Zero Lag (Well, Almost)"):
    updates the baseline EMA once per accepted finite sample, evaluates candidate
    gains ``index / 10.0`` for inclusive integers ``-gain_limit .. +gain_limit``
    against the frozen previous-bar EC using strict improvement from the literal
    ``1_000_000.0`` sentinel, and exposes both EC (`value`) and baseline `ema`
    outputs from the first accepted finite observation.
    """

    def __init__(
        self,
        params: Optional[ZeroLagErrorCorrectingExponentialMovingAverageParams] = None,
    ) -> None:
        source = default_length_params() if params is None else params
        invalid = "invalid zero-lag error-correcting exponential moving average parameters"

        if isinstance(
            source, ZeroLagErrorCorrectingExponentialMovingAverageLengthParams
        ):
            if type(source.length) is not int or source.length < 1:
                raise ValueError(f"{invalid}: length must be an integer >= 1")
            try:
                alpha = 2.0 / (source.length + 1)
            except OverflowError as error:
                raise ValueError(
                    f"{invalid}: length must yield representable coefficients"
                ) from error
            complement = 1.0 - alpha
            if not (0.0 < alpha <= 1.0) or not (0.0 <= complement < 1.0):
                raise ValueError(
                    f"{invalid}: length must yield representably stable EMA coefficients"
                )
            length_value = source.length
        elif isinstance(
            source,
            ZeroLagErrorCorrectingExponentialMovingAverageSmoothingFactorParams,
        ):
            if isinstance(source.smoothing_factor, bool) or not isinstance(
                source.smoothing_factor, (int, float)
            ):
                raise ValueError(f"{invalid}: smoothing_factor must be a real number")
            alpha = float(source.smoothing_factor)
            if not math.isfinite(alpha) or not (0.0 < alpha <= 1.0):
                raise ValueError(f"{invalid}: smoothing_factor must be in (0, 1]")
            complement = 1.0 - alpha
            if not (0.0 <= complement < 1.0):
                raise ValueError(
                    f"{invalid}: smoothing_factor must yield representably stable EMA coefficients"
                )
            length_value = equivalent_length_from_alpha(alpha)
        else:
            raise ValueError(f"{invalid}: unsupported params type")

        if type(source.gain_limit) is not int or source.gain_limit < 0:
            raise ValueError(
                f"{invalid}: gain_limit must be a nonnegative integer in tenths"
            )

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

        if isinstance(
            source,
            ZeroLagErrorCorrectingExponentialMovingAverageSmoothingFactorParams,
        ):
            normalized: ZeroLagErrorCorrectingExponentialMovingAverageParams = (
                ZeroLagErrorCorrectingExponentialMovingAverageSmoothingFactorParams(
                    smoothing_factor=alpha,
                    gain_limit=source.gain_limit,
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
            )
        else:
            normalized = ZeroLagErrorCorrectingExponentialMovingAverageLengthParams(
                length=length_value,
                gain_limit=source.gain_limit,
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

        self._length = length_value
        self._gain_limit = source.gain_limit
        self._alpha = alpha
        self._complement = complement

        self._bar_func = bar_component_value(bar_comp)
        self._quote_func = quote_component_value(quote_comp)
        self._trade_func = trade_component_value(trade_comp)

        self._mnemonic = (
            zero_lag_error_correcting_exponential_moving_average_mnemonic(normalized)
        )
        self._mnemonic_ema = (
            zero_lag_error_correcting_exponential_moving_average_ema_mnemonic(
                normalized
            )
        )
        self._description = (
            f"Zero-lag Error-Correcting Exponential Moving Average {self._mnemonic}"
        )
        self._description_ema = (
            "Zero-lag Error-Correcting Exponential Moving Average baseline EMA "
            f"{self._mnemonic_ema}"
        )

        self._count = 0
        self._ema_state = 0.0
        self._ec_state = 0.0
        self._best_gain_state = 0.0
        self._least_error_state = 0.0
        self._current_value = math.nan
        self._current_ema = math.nan

    @staticmethod
    def from_length(
        params: Optional[ZeroLagErrorCorrectingExponentialMovingAverageLengthParams] = None,
    ) -> "ZeroLagErrorCorrectingExponentialMovingAverage":
        """Create an instance from length-based parameters."""
        return ZeroLagErrorCorrectingExponentialMovingAverage(
            default_length_params() if params is None else params
        )

    @staticmethod
    def from_smoothing_factor(
        params: Optional[ZeroLagErrorCorrectingExponentialMovingAverageSmoothingFactorParams] = None,
    ) -> "ZeroLagErrorCorrectingExponentialMovingAverage":
        """Create an instance from smoothing-factor parameters."""
        return ZeroLagErrorCorrectingExponentialMovingAverage(
            default_smoothing_factor_params() if params is None else params
        )

    @staticmethod
    def create(
        params: ZeroLagErrorCorrectingExponentialMovingAverageParams,
    ) -> "ZeroLagErrorCorrectingExponentialMovingAverage":
        """Create an instance from either length or smoothing-factor parameters."""
        return ZeroLagErrorCorrectingExponentialMovingAverage(params)

    @staticmethod
    def create_default() -> "ZeroLagErrorCorrectingExponentialMovingAverage":
        """Create an instance with Ehlers defaults (length=20, gain_limit=50)."""
        return ZeroLagErrorCorrectingExponentialMovingAverage(default_length_params())

    def reset(self) -> None:
        """Restore zero startup state while preserving configuration."""
        self._count = 0
        self._ema_state = 0.0
        self._ec_state = 0.0
        self._best_gain_state = 0.0
        self._least_error_state = 0.0
        self._current_value = math.nan
        self._current_ema = math.nan

    def warmup_bars(self) -> int:
        """Return the warmup requirement (0 for this zero-seeded recursive filter)."""
        return 0

    def length(self) -> int:
        """Return the configured or equivalent EMA length."""
        return self._length

    def gain_limit(self) -> int:
        """Return the configured integer gain bound in tenths."""
        return self._gain_limit

    def alpha(self) -> float:
        """Return the exact baseline EMA smoothing factor."""
        return self._alpha

    def samples_seen(self) -> int:
        """Return the count of accepted finite observations since creation or reset."""
        return self._count

    def is_primed(self) -> bool:
        """Return True once at least one finite observation has been accepted."""
        return self._count > 0

    def value(self) -> float:
        """Return the latest selected EC output (NaN before priming)."""
        return self._current_value

    def ema(self) -> float:
        """Return the latest baseline EMA output (NaN before priming)."""
        return self._current_ema

    def best_gain(self) -> float:
        """Return the latest selected gain in actual gain units (NaN before priming)."""
        return self._best_gain_state if self._count > 0 else math.nan

    def least_error(self) -> float:
        """Return the latest search least-error state (NaN before priming)."""
        return self._least_error_state if self._count > 0 else math.nan

    def metadata(self) -> Metadata:
        """Describe the output data of the indicator."""
        return build_metadata(
            Identifier.ZERO_LAG_ERROR_CORRECTING_EXPONENTIAL_MOVING_AVERAGE,
            self._mnemonic,
            self._description,
            [
                OutputText(self._mnemonic, self._description),
                OutputText(self._mnemonic_ema, self._description_ema),
            ],
        )

    def _step(self, sample: object) -> tuple[float, float]:
        price = _parse_sample(sample)
        if math.isnan(price):
            return (math.nan, math.nan)

        alpha = self._alpha
        complement = self._complement
        previous_ec = self._ec_state

        ema = alpha * price + complement * self._ema_state
        if not math.isfinite(ema):
            raise ValueError("baseline EMA arithmetic overflowed")

        least_error = _LEAST_ERROR_SENTINEL
        best_gain = self._best_gain_state
        for index in range(-self._gain_limit, self._gain_limit + 1):
            gain = index / 10.0
            candidate = (
                alpha * (ema + gain * (price - previous_ec))
                + complement * previous_ec
            )
            error_value = abs(price - candidate)
            if not math.isfinite(candidate) or not math.isfinite(error_value):
                raise ValueError("gain-search arithmetic overflowed")
            if error_value < least_error:
                least_error = error_value
                best_gain = gain

        ec = (
            alpha * (ema + best_gain * (price - previous_ec))
            + complement * previous_ec
        )
        if not math.isfinite(ec):
            raise ValueError("selected EC arithmetic overflowed")

        self._ema_state = ema
        self._ec_state = ec
        self._best_gain_state = best_gain
        self._least_error_state = least_error
        self._count += 1
        self._current_value = ec
        self._current_ema = ema
        return (ec, ema)

    def update(self, sample: float) -> float:
        """Consume one scalar sample and return the selected EC output."""
        ec, _ = self._step(sample)
        return ec

    def _update_entity(self, stamp: Any, sample: float) -> Output:
        ec, ema = self._step(sample)
        return [
            Scalar(time=stamp, value=ec),
            Scalar(time=stamp, value=ema),
        ]

    def update_scalar(self, sample: Scalar) -> Output:
        return self._update_entity(sample.time, sample.value)

    def update_bar(self, sample: Bar) -> Output:
        return self._update_entity(sample.time, self._bar_func(sample))

    def update_quote(self, sample: Quote) -> Output:
        return self._update_entity(sample.time, self._quote_func(sample))

    def update_trade(self, sample: Trade) -> Output:
        return self._update_entity(sample.time, self._trade_func(sample))
