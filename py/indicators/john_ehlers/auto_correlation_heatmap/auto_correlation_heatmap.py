"""Ehlers' February 2025 Autocorrelation Heatmap indicator."""

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
    AutoCorrelationHeatmapParams,
    auto_correlation_heatmap_mnemonic,
    default_params,
)

SQRT_TWO: float = math.sqrt(2.0)


class AutocorrelationState(NamedTuple):
    """Immutable diagnostic snapshot of committed internal filter and correlation state."""

    accepted_samples: int
    filtered: float
    correlations: tuple[float, ...]


AutoCorrelationHeatmapState = AutocorrelationState


def _validate_integer_min(value: object, name: str, minimum: int) -> int:
    if isinstance(value, bool) or not isinstance(value, int) or value < minimum:
        raise ValueError(f"{name} must be an integer >= {minimum}")
    return value


def _validate_integer_range(
    value: object, name: str, minimum: int, maximum: int
) -> int:
    if (
        isinstance(value, bool)
        or not isinstance(value, int)
        or value < minimum
        or value > maximum
    ):
        raise ValueError(f"{name} must be an integer in [{minimum}, {maximum}]")
    return value


def _validate_real_at_least(value: object, name: str, minimum: float) -> float:
    if isinstance(value, bool) or not isinstance(value, (int, float)):
        raise ValueError(f"{name} must be a finite real number >= {minimum}")
    try:
        number = float(value)
    except (ValueError, OverflowError) as error:
        raise ValueError(
            f"{name} must be a finite real number >= {minimum}"
        ) from error
    if not math.isfinite(number) or number < minimum:
        raise ValueError(f"{name} must be a finite real number >= {minimum}")
    return number


def _validate_sample(value: object) -> float:
    if isinstance(value, bool) or not isinstance(value, (int, float)):
        raise ValueError("price must be finite or NaN (missing sample)")
    try:
        price = float(value)
    except (ValueError, OverflowError) as error:
        raise ValueError("price must be finite or NaN (missing sample)") from error
    if math.isinf(price):
        raise ValueError("price must be finite or NaN (missing sample)")
    return price


class AutoCorrelationHeatmap(Indicator):
    """Ehlers' February 2025 Autocorrelation Heatmap (ACH)."""

    def __init__(
        self,
        params: Optional[AutoCorrelationHeatmapParams] = None,
    ) -> None:
        source = default_params() if params is None else params

        length = _validate_integer_min(source.length, "length", 2)
        smoothing_period = _validate_real_at_least(
            source.smoothing_period, "smoothing_period", 1.0
        )
        max_lag = _validate_integer_range(source.max_lag, "max_lag", 0, 99)

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

        angle = SQRT_TWO * math.pi / smoothing_period
        attenuation = math.exp(-angle)
        coeff_c2 = 2.0 * attenuation * math.cos(angle)
        coeff_c3 = -attenuation * attenuation
        coeff_c1 = (1.0 + coeff_c2 - coeff_c3) / 4.0
        if not all(
            math.isfinite(coefficient)
            for coefficient in (coeff_c1, coeff_c2, coeff_c3)
        ):
            raise ValueError("smoothing_period produces non-finite filter coefficients")

        self._parameters = AutoCorrelationHeatmapParams(
            length=length,
            smoothing_period=smoothing_period,
            max_lag=max_lag,
            bar_component=source.bar_component,
            quote_component=source.quote_component,
            trade_component=source.trade_component,
        )
        self._length: int = length
        self._smoothing_period: float = smoothing_period
        self._max_lag: int = max_lag
        self._history_length: int = length + max_lag
        self._prime_count: int = max(4, self._history_length)
        self._c1: float = coeff_c1
        self._c2: float = coeff_c2
        self._c3: float = coeff_c3

        self.mnemonic: str = auto_correlation_heatmap_mnemonic(self._parameters)
        self.description: str = f"Autocorrelation heatmap {self.mnemonic}"

        self._bar_func: Callable[[Bar], float] = bar_component_value(bar_comp)
        self._quote_func: Callable[[Quote], float] = quote_component_value(quote_comp)
        self._trade_func: Callable[[Trade], float] = trade_component_value(trade_comp)

        self._count: int = 0
        self._price1: float = 0.0
        self._price2: float = 0.0
        self._filtered1: float = 0.0
        self._filtered2: float = 0.0
        self._history: list[float] = []
        self._correlations: list[float] = []
        self._output: Heatmap = Heatmap.empty(None, 0.0, float(max_lag), 1.0)
        self.reset()

    @staticmethod
    def create(
        params: Optional[AutoCorrelationHeatmapParams] = None,
    ) -> AutoCorrelationHeatmap:
        """Construct an instance from validated parameters."""
        return AutoCorrelationHeatmap(params)

    @staticmethod
    def create_default() -> AutoCorrelationHeatmap:
        """Construct an instance using Ehlers' February 2025 article defaults."""
        return AutoCorrelationHeatmap(default_params())

    @property
    def parameters(self) -> AutoCorrelationHeatmapParams:
        return self._parameters

    @property
    def length(self) -> int:
        return self._length

    @property
    def smoothing_period(self) -> float:
        return self._smoothing_period

    @property
    def max_lag(self) -> int:
        return self._max_lag

    @property
    def warmup_bars(self) -> int:
        """Number of leading empty public columns on a finite input stream."""
        return self._prime_count - 1

    @property
    def samples_seen(self) -> int:
        return self._count

    @property
    def state(self) -> AutocorrelationState:
        return AutocorrelationState(
            self._count, self._filtered1, tuple(self._correlations)
        )

    @property
    def output(self) -> Heatmap:
        return self._output

    def reset(self) -> None:
        """Preserve configuration and restore explicit source initial conditions."""
        self._count = 0
        self._price1 = 0.0
        self._price2 = 0.0
        self._filtered1 = 0.0
        self._filtered2 = 0.0
        self._history = [0.0] * self._history_length
        self._correlations = [0.0] * (self._max_lag + 1)
        self._output = Heatmap.empty(None, 0.0, float(self._max_lag), 1.0)

    def is_primed(self) -> bool:
        """Return whether sufficient history has accumulated for all lags."""
        return self._count >= self._prime_count

    def metadata(self) -> Metadata:
        """Describe the indicator and its correlation heatmap output."""
        return build_metadata(
            Identifier.AUTO_CORRELATION_HEATMAP,
            self.mnemonic,
            self.description,
            [OutputText(mnemonic=self.mnemonic, description=self.description)],
        )

    def update(self, sample: float, time: Any = None) -> Heatmap:
        """Consume one scalar price and return a signed correlation heatmap snapshot."""
        price = _validate_sample(sample)
        if math.isnan(price):
            return Heatmap.empty(time, 0.0, float(self._max_lag), 1.0)

        if self._count < 3:
            filtered = price
        else:
            filtered = (
                (1.0 - self._c1) * price
                + (2.0 * self._c1 - self._c2) * self._price1
                - (self._c1 + self._c3) * self._price2
                + self._c2 * self._filtered1
                + self._c3 * self._filtered2
            )
        if not math.isfinite(filtered):
            raise ValueError("Ultimate Smoother calculation overflowed")

        next_history = [filtered] + self._history[:-1]
        next_correlations = self._correlations.copy()
        length = self._length

        for lag in range(self._max_lag + 1):
            sum_x = 0.0
            sum_y = 0.0
            sum_xx = 0.0
            sum_xy = 0.0
            sum_yy = 0.0
            for offset in range(length):
                sample_x = next_history[offset]
                sample_y = next_history[lag + offset]
                sum_x = sum_x + sample_x
                sum_y = sum_y + sample_y
                sum_xx = sum_xx + sample_x * sample_x
                sum_xy = sum_xy + sample_x * sample_y
                sum_yy = sum_yy + sample_y * sample_y
            var_x = length * sum_xx - sum_x * sum_x
            var_y = length * sum_yy - sum_y * sum_y
            if not all(
                math.isfinite(item)
                for item in (sum_x, sum_y, sum_xx, sum_xy, sum_yy, var_x, var_y)
            ):
                raise ValueError("Pearson sums overflowed; reduce input magnitude")
            if var_x > 0.0 and var_y > 0.0:
                product = var_x * var_y
                numerator = length * sum_xy - sum_x * sum_y
                if (
                    product <= 0.0
                    or not math.isfinite(product)
                    or not math.isfinite(numerator)
                ):
                    raise ValueError(
                        "Pearson denominator or numerator overflowed/underflowed"
                    )
                correlation = numerator / math.sqrt(product)
                if not math.isfinite(correlation):
                    raise ValueError("Pearson correlation is not finite")
                next_correlations[lag] = correlation

        self._price2 = self._price1
        self._price1 = price
        self._filtered2 = self._filtered1
        self._filtered1 = filtered
        self._history = next_history
        self._correlations = next_correlations
        self._count += 1

        if not self.is_primed():
            result = Heatmap.empty(time, 0.0, float(self._max_lag), 1.0)
        else:
            result = Heatmap(
                time,
                0.0,
                float(self._max_lag),
                1.0,
                min(next_correlations),
                max(next_correlations),
                next_correlations.copy(),
            )
        self._output = result
        return result

    def update_scalar(self, sample: Scalar) -> Output:
        return [self.update(sample.value, sample.time)]

    def update_bar(self, sample: Bar) -> Output:
        return [self.update(self._bar_func(sample), sample.time)]

    def update_quote(self, sample: Quote) -> Output:
        return [self.update(self._quote_func(sample), sample.time)]

    def update_trade(self, sample: Trade) -> Output:
        return [self.update(self._trade_func(sample), sample.time)]
