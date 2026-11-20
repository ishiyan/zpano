"""Ehlers' 2013 Roofed Autocorrelation Heatmap indicator (Book Listing 8-2)."""

from __future__ import annotations

import math
from typing import Any, Callable, NamedTuple, Optional

from ....entities.bar import Bar
from ....entities.bar_component import (
    BarComponent,
    DEFAULT_BAR_COMPONENT,
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
from ...core.outputs.heatmap import Heatmap
from .params import (
    RoofedAutoCorrelationHeatmapParams,
    default_params,
    roofed_auto_correlation_heatmap_mnemonic,
)

TWO_PI: float = 2.0 * math.pi
SQRT_TWO: float = math.sqrt(2.0)


class AutocorrelationState(NamedTuple):
    """Immutable diagnostic snapshot of committed internal roofing and correlation state."""

    high_pass: float
    filtered: float
    correlations: tuple[float, ...]


RoofedAutoCorrelationHeatmapState = AutocorrelationState


def _validate_integer(value: object, name: str, minimum: int) -> int:
    if isinstance(value, bool) or not isinstance(value, int) or value < minimum:
        raise ValueError(f"{name} must be an integer >= {minimum}")
    return value


def _validate_real(value: object, name: str, minimum: float) -> float:
    if isinstance(value, bool) or not isinstance(value, (int, float)):
        raise ValueError(f"{name} must be a finite real >= {minimum}")
    try:
        number = float(value)
    except (ValueError, OverflowError) as error:
        raise ValueError(f"{name} must be a finite real >= {minimum}") from error
    if not math.isfinite(number) or number < minimum:
        raise ValueError(f"{name} must be a finite real >= {minimum}")
    return number


def _validate_sample(value: object) -> float:
    if isinstance(value, bool) or not isinstance(value, (int, float)):
        raise ValueError("price must be a real number")
    try:
        price = float(value)
    except (ValueError, OverflowError) as error:
        raise ValueError("price must be finite or NaN (missing)") from error
    if math.isinf(price):
        raise ValueError("price must be finite or NaN (missing)")
    return price


class _Roof:
    """Two-pole high-pass followed by two-pole SuperSmoother (Book Equation 3-3)."""

    def __init__(self, highpass_period: float, smoothing_period: float) -> None:
        hp = _validate_real(highpass_period, "highpass_period", 5.0)
        lp = _validate_real(smoothing_period, "smoothing_period", 2.0)
        angle = TWO_PI / (SQRT_TWO * hp)
        cosine = math.cos(angle)
        alpha = (cosine + math.sin(angle) - 1.0) / cosine
        pole = 1.0 - alpha
        attenuation = math.exp(-SQRT_TWO * math.pi / lp)
        self.hp0 = (1.0 - alpha / 2.0) * (1.0 - alpha / 2.0)
        self.hp1 = 2.0 * pole
        self.hp2 = pole * pole
        self.ss2 = 2.0 * attenuation * math.cos(SQRT_TWO * math.pi / lp)
        self.ss3 = -attenuation * attenuation
        self.ss1 = 1.0 - self.ss2 - self.ss3
        if (
            not (0.0 < pole < 1.0 and 0.0 < attenuation < 1.0 and self.ss1 > 0.0)
            or not all(
                math.isfinite(coefficient)
                for coefficient in (
                    self.hp0,
                    self.hp1,
                    self.hp2,
                    self.ss1,
                    self.ss2,
                    self.ss3,
                )
            )
        ):
            raise ValueError("periods must produce finite resolvable stable coefficients")
        self.x1 = 0.0
        self.x2 = 0.0
        self.h1 = 0.0
        self.h2 = 0.0
        self.f1 = 0.0
        self.f2 = 0.0

    def reset(self) -> None:
        self.x1 = 0.0
        self.x2 = 0.0
        self.h1 = 0.0
        self.h2 = 0.0
        self.f1 = 0.0
        self.f2 = 0.0

    def candidate(self, price: float) -> tuple[float, float]:
        high = (
            self.hp0 * (price - 2.0 * self.x1 + self.x2)
            + self.hp1 * self.h1
            - self.hp2 * self.h2
        )
        filtered = (
            self.ss1 * (high + self.h1) / 2.0
            + self.ss2 * self.f1
            + self.ss3 * self.f2
        )
        if not (math.isfinite(high) and math.isfinite(filtered)):
            raise ValueError("roofing arithmetic overflowed")
        return high, filtered

    def commit(self, price: float, high: float, filtered: float) -> None:
        self.x2, self.x1 = self.x1, price
        self.h2, self.h1 = self.h1, high
        self.f2, self.f1 = self.f1, filtered


class RoofedAutoCorrelationHeatmap(Indicator):
    """Ehlers' 2013 Roofed Autocorrelation Heatmap (RACH)."""

    def __init__(
        self,
        params: Optional[RoofedAutoCorrelationHeatmapParams] = None,
    ) -> None:
        source = default_params() if params is None else params

        min_lag = _validate_integer(source.min_lag, "min_lag", 0)
        max_lag = _validate_integer(source.max_lag, "max_lag", min_lag)
        averaging_length = _validate_integer(
            source.averaging_length, "averaging_length", 0
        )
        highpass_period = _validate_real(
            source.highpass_period, "highpass_period", 5.0
        )
        smoothing_period = _validate_real(
            source.smoothing_period, "smoothing_period", 2.0
        )

        if source.bar_component is not None and (
            isinstance(source.bar_component, bool)
            or not isinstance(source.bar_component, BarComponent)
        ):
            raise ValueError("bar_component must be a BarComponent or None")
        if source.quote_component is not None and (
            isinstance(source.quote_component, bool)
            or not isinstance(source.quote_component, QuoteComponent)
        ):
            raise ValueError("quote_component must be a QuoteComponent or None")
        if source.trade_component is not None and (
            isinstance(source.trade_component, bool)
            or not isinstance(source.trade_component, TradeComponent)
        ):
            raise ValueError("trade_component must be a TradeComponent or None")

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

        self._roof = _Roof(highpass_period, smoothing_period)
        self._parameters = RoofedAutoCorrelationHeatmapParams(
            min_lag=min_lag,
            max_lag=max_lag,
            averaging_length=averaging_length,
            highpass_period=highpass_period,
            smoothing_period=smoothing_period,
            bar_component=source.bar_component,
            quote_component=source.quote_component,
            trade_component=source.trade_component,
        )
        self._min_lag: int = min_lag
        self._max_lag: int = max_lag
        self._averaging_length: int = averaging_length
        self._highpass_period: float = highpass_period
        self._smoothing_period: float = smoothing_period
        self._history_length: int = max(
            1, max_lag + (averaging_length if averaging_length > 0 else max_lag)
        )

        self.mnemonic: str = roofed_auto_correlation_heatmap_mnemonic(self._parameters)
        self.description: str = f"Roofed autocorrelation heatmap {self.mnemonic}"

        self._bar_func: Callable[[Bar], float] = bar_component_value(bar_comp)
        self._quote_func: Callable[[Quote], float] = quote_component_value(quote_comp)
        self._trade_func: Callable[[Trade], float] = trade_component_value(trade_comp)

        self._count: int = 0
        self._history: list[float] = []
        self._correlations: list[float] = []
        self._output: Heatmap = Heatmap.empty(
            None, float(min_lag), float(max_lag), 1.0
        )
        self.reset()

    @staticmethod
    def create(
        params: Optional[RoofedAutoCorrelationHeatmapParams] = None,
    ) -> RoofedAutoCorrelationHeatmap:
        """Construct an instance from validated parameters."""
        return RoofedAutoCorrelationHeatmap(params)

    @staticmethod
    def create_default() -> RoofedAutoCorrelationHeatmap:
        """Construct an instance using Ehlers' 2013 Listing 8-2 defaults."""
        return RoofedAutoCorrelationHeatmap(default_params())

    @property
    def parameters(self) -> RoofedAutoCorrelationHeatmapParams:
        return self._parameters

    @property
    def min_lag(self) -> int:
        return self._min_lag

    @property
    def max_lag(self) -> int:
        return self._max_lag

    @property
    def averaging_length(self) -> int:
        return self._averaging_length

    @property
    def highpass_period(self) -> float:
        return self._highpass_period

    @property
    def smoothing_period(self) -> float:
        return self._smoothing_period

    @property
    def min_period(self) -> int:
        return self._min_lag

    @property
    def max_period(self) -> int:
        return self._max_lag

    @property
    def warmup_bars(self) -> int:
        """Number of leading empty columns on a finite input stream."""
        return self._history_length - 1

    @property
    def samples_seen(self) -> int:
        return self._count

    @property
    def state(self) -> AutocorrelationState:
        return AutocorrelationState(
            self._roof.h1, self._roof.f1, tuple(self._correlations)
        )

    @property
    def output(self) -> Heatmap:
        if self._output.is_empty():
            return Heatmap.empty(
                self._output.time,
                self._output.parameter_first,
                self._output.parameter_last,
                self._output.parameter_resolution,
            )
        return Heatmap(
            self._output.time,
            self._output.parameter_first,
            self._output.parameter_last,
            self._output.parameter_resolution,
            self._output.value_min,
            self._output.value_max,
            list(self._output.values),
        )

    def reset(self) -> None:
        """Preserve configuration and restore explicit source initial conditions."""
        self._roof.reset()
        self._count = 0
        self._history = [0.0] * self._history_length
        self._correlations = [0.0] * (self._max_lag + 1)
        self._output = Heatmap.empty(
            None, float(self._min_lag), float(self._max_lag), 1.0
        )

    def is_primed(self) -> bool:
        """Return whether sufficient history has accumulated for all lags."""
        return self._count >= self._history_length

    def metadata(self) -> Metadata:
        """Describe the indicator and its roofed correlation heatmap output."""
        return build_metadata(
            Identifier.ROOFED_AUTO_CORRELATION_HEATMAP,
            self.mnemonic,
            self.description,
            [OutputText(mnemonic=self.mnemonic, description=self.description)],
        )

    def update(self, sample: float, time: Any = None) -> Heatmap:
        """Consume one scalar price and return a roofed correlation heatmap snapshot."""
        price = _validate_sample(sample)
        if math.isnan(price):
            return Heatmap.empty(
                time, float(self._min_lag), float(self._max_lag), 1.0
            )

        high, filtered = self._roof.candidate(price)
        next_history = [filtered] + self._history[:-1]
        next_correlations = self._correlations.copy()

        for lag in range(self._max_lag + 1):
            window = self._averaging_length if self._averaging_length > 0 else lag
            sum_x = 0.0
            sum_y = 0.0
            sum_xx = 0.0
            sum_xy = 0.0
            sum_yy = 0.0
            for count in range(window):
                sample_x = next_history[count]
                sample_y = next_history[lag + count]
                sum_x = sum_x + sample_x
                sum_y = sum_y + sample_y
                sum_xx = sum_xx + sample_x * sample_x
                sum_xy = sum_xy + sample_x * sample_y
                sum_yy = sum_yy + sample_y * sample_y
            denominator = (window * sum_xx - sum_x * sum_x) * (
                window * sum_yy - sum_y * sum_y
            )
            if not all(
                math.isfinite(item)
                for item in (sum_x, sum_y, sum_xx, sum_xy, sum_yy, denominator)
            ):
                raise ValueError("Pearson arithmetic overflowed")
            if denominator > 0.0:
                next_correlations[lag] = (window * sum_xy - sum_x * sum_y) / math.sqrt(
                    denominator
                )
            next_correlations[lag] = 0.5 * (next_correlations[lag] + 1.0)
            if not math.isfinite(next_correlations[lag]):
                raise ValueError("correlation arithmetic overflowed")

        self._roof.commit(price, high, filtered)
        self._history = next_history
        self._correlations = next_correlations
        self._count += 1

        if not self.is_primed():
            result = Heatmap.empty(
                time, float(self._min_lag), float(self._max_lag), 1.0
            )
        else:
            displayed = next_correlations[self._min_lag : self._max_lag + 1]
            result = Heatmap(
                time,
                float(self._min_lag),
                float(self._max_lag),
                1.0,
                min(displayed),
                max(displayed),
                list(displayed),
            )
        self._output = (
            Heatmap.empty(
                result.time,
                result.parameter_first,
                result.parameter_last,
                result.parameter_resolution,
            )
            if result.is_empty()
            else Heatmap(
                result.time,
                result.parameter_first,
                result.parameter_last,
                result.parameter_resolution,
                result.value_min,
                result.value_max,
                list(result.values),
            )
        )
        return result

    def update_scalar(self, sample: Scalar) -> Output:
        return [self.update(sample.value, sample.time)]

    def update_bar(self, sample: Bar) -> Output:
        return [self.update(self._bar_func(sample), sample.time)]

    def update_quote(self, sample: Quote) -> Output:
        return [self.update(self._quote_func(sample), sample.time)]

    def update_trade(self, sample: Trade) -> Output:
        return [self.update(self._trade_func(sample), sample.time)]
