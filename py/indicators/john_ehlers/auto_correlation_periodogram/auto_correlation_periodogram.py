"""Ehlers' Autocorrelation Periodogram indicator with dominant cycle estimation."""

from __future__ import annotations

import datetime
import math
from typing import NamedTuple, Optional

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
    AutoCorrelationPeriodogramParams,
    AutoCorrelationPeriodogramSource,
    auto_correlation_periodogram_dominant_cycle_mnemonic,
    auto_correlation_periodogram_mnemonic,
    default_params,
    resolve_min_period,
)

TWO_PI = 2.0 * math.pi
SQRT_TWO = math.sqrt(2.0)
INVERSE_SQRT_TWO = 1.0 / SQRT_TWO
DFT_LAG_START = 3
POWER_SMOOTHING = 0.2
POWER_RETENTION = 0.8
BOOK_AGC_DECAY = 0.995
ARTICLE_THRESHOLD = 0.25
BOOK_THRESHOLD = 0.5


class PeriodogramOutput(NamedTuple):
    """Autocorrelation periodogram heatmap column and real-valued dominant cycle period."""

    heatmap: Heatmap
    dominant_cycle: float


class PeriodogramState(NamedTuple):
    """Read-only snapshot of the committed internal recursive state."""

    high_pass: float
    filtered: float
    correlations: tuple[float, ...]
    smoothed_power: tuple[float, ...]
    normalization_maximum: float
    internal_dominant_cycle: float


def _validate_integer(value: int, name: str) -> int:
    if isinstance(value, bool) or not isinstance(value, int):
        raise ValueError(f"{name} must be an integer")
    return value


def _validate_source(
    value: AutoCorrelationPeriodogramSource | str,
) -> AutoCorrelationPeriodogramSource:
    if isinstance(value, bool) or not isinstance(value, str):
        raise ValueError('source must be "article" or "book"')
    if value == AutoCorrelationPeriodogramSource.ARTICLE.value:
        return AutoCorrelationPeriodogramSource.ARTICLE
    if value == AutoCorrelationPeriodogramSource.BOOK.value:
        return AutoCorrelationPeriodogramSource.BOOK
    raise ValueError('source must be "article" or "book"')


def _validate_sample(value: float) -> float:
    if isinstance(value, bool) or not isinstance(value, (int, float)):
        raise ValueError("sample must be a real number")
    numeric = float(value)
    if math.isinf(numeric):
        raise ValueError("sample must be finite or NaN (skipped)")
    return numeric


class AutoCorrelationPeriodogram(Indicator):
    """Ehlers' Autocorrelation Periodogram (TASC Sept 2016 / Book Listing 8-3)."""

    def __init__(
        self, params: Optional[AutoCorrelationPeriodogramParams] = None
    ) -> None:
        configuration = default_params() if params is None else params

        resolved_source = _validate_source(configuration.source)
        if configuration.min_period is not None:
            _validate_integer(configuration.min_period, "min_period")
        resolved_min = resolve_min_period(resolved_source, configuration.min_period)
        _validate_integer(resolved_min, "min_period")
        last_period = _validate_integer(configuration.max_period, "max_period")
        averaging = _validate_integer(configuration.averaging_length, "averaging_length")

        if resolved_min < 3:
            raise ValueError("min_period must be at least 3 (strictly above Nyquist)")
        if last_period < 5 or last_period <= resolved_min:
            raise ValueError("max_period must be at least 5 and greater than min_period")
        if averaging != 0 and averaging < 2:
            raise ValueError("averaging_length must be zero or at least 2")
        if not isinstance(configuration.enhance_resolution, bool):
            raise ValueError("enhance_resolution must be a boolean")

        if configuration.bar_component is not None and (
            isinstance(configuration.bar_component, bool)
            or not isinstance(configuration.bar_component, BarComponent)
        ):
            raise ValueError("bar_component must be a BarComponent or None")
        if configuration.quote_component is not None and (
            isinstance(configuration.quote_component, bool)
            or not isinstance(configuration.quote_component, QuoteComponent)
        ):
            raise ValueError("quote_component must be a QuoteComponent or None")
        if configuration.trade_component is not None and (
            isinstance(configuration.trade_component, bool)
            or not isinstance(configuration.trade_component, TradeComponent)
        ):
            raise ValueError("trade_component must be a TradeComponent or None")

        try:
            if resolved_source == AutoCorrelationPeriodogramSource.ARTICLE:
                theta = TWO_PI / last_period
                pole = (1.0 - math.sin(theta)) / math.cos(theta)
                hp0 = 0.5 * (1.0 + pole)
                hp1 = pole
                hp2 = 0.0
            else:
                theta = INVERSE_SQRT_TWO * TWO_PI / last_period
                cosine = math.cos(theta)
                alpha = (cosine + math.sin(theta) - 1.0) / cosine
                pole = 1.0 - alpha
                hp0 = (1.0 - alpha / 2.0) * (1.0 - alpha / 2.0)
                hp1 = 2.0 * pole
                hp2 = pole * pole
            decay = math.exp(-SQRT_TWO * math.pi / resolved_min)
            ss2 = 2.0 * decay * math.cos(SQRT_TWO * math.pi / resolved_min)
            ss3 = -decay * decay
            ss1 = 1.0 - ss2 - ss3
        except (OverflowError, ZeroDivisionError) as error:
            raise ValueError("periods must yield finite stable coefficients") from error
        if not (0.0 < pole < 1.0 and 0.0 < decay < 1.0 and ss1 > 0.0):
            raise ValueError("periods must yield finite stable coefficients")

        bar_comp = (
            configuration.bar_component
            if configuration.bar_component is not None
            else DEFAULT_BAR_COMPONENT
        )
        quote_comp = (
            configuration.quote_component
            if configuration.quote_component is not None
            else DEFAULT_QUOTE_COMPONENT
        )
        trade_comp = (
            configuration.trade_component
            if configuration.trade_component is not None
            else DEFAULT_TRADE_COMPONENT
        )

        self._hp0 = hp0
        self._hp1 = hp1
        self._hp2 = hp2
        self._ss1 = ss1
        self._ss2 = ss2
        self._ss3 = ss3

        self._parameters = AutoCorrelationPeriodogramParams(
            source=resolved_source,
            min_period=resolved_min,
            max_period=last_period,
            averaging_length=averaging,
            enhance_resolution=configuration.enhance_resolution,
            bar_component=configuration.bar_component,
            quote_component=configuration.quote_component,
            trade_component=configuration.trade_component,
        )

        self._history_length: int = max(3, last_period + (averaging or last_period))
        self._cosines: list[tuple[float, ...]] = [
            tuple(
                math.cos(TWO_PI * lag / period)
                for lag in range(DFT_LAG_START, last_period + 1)
            )
            for period in range(resolved_min, last_period + 1)
        ]
        self._sines: list[tuple[float, ...]] = [
            tuple(
                math.sin(TWO_PI * lag / period)
                for lag in range(DFT_LAG_START, last_period + 1)
            )
            for period in range(resolved_min, last_period + 1)
        ]

        self._bar_func = bar_component_value(bar_comp)
        self._quote_func = quote_component_value(quote_comp)
        self._trade_func = trade_component_value(trade_comp)

        self.mnemonic: str = auto_correlation_periodogram_mnemonic(self._parameters)
        self.description: str = "Autocorrelation periodogram " + self.mnemonic
        self._mnemonic_dc: str = auto_correlation_periodogram_dominant_cycle_mnemonic(
            self._parameters
        )
        self._description_dc: str = (
            "Autocorrelation periodogram dominant cycle " + self._mnemonic_dc
        )

        self._samples_seen: int = 0
        self._close1: float = 0.0
        self._close2: float = 0.0
        self._high1: float = 0.0
        self._high2: float = 0.0
        self._filtered: list[float] = []
        self._correlations: list[float] = []
        self._power: list[float] = []
        self._maximum: float = 0.0
        self._dominant: float = 0.0
        self._has_dominant: bool = False
        self._state: PeriodogramState = PeriodogramState(
            0.0, 0.0, (), (), 0.0, 0.0
        )
        self._output: PeriodogramOutput = PeriodogramOutput(
            Heatmap.empty(None, float(resolved_min), float(last_period), 1.0),
            math.nan,
        )
        self.reset()

    @staticmethod
    def create(
        params: Optional[AutoCorrelationPeriodogramParams] = None,
    ) -> "AutoCorrelationPeriodogram":
        """Create a new AutoCorrelationPeriodogram from parameters."""
        return AutoCorrelationPeriodogram(params)

    @staticmethod
    def create_default() -> "AutoCorrelationPeriodogram":
        """Create a new AutoCorrelationPeriodogram with default parameters."""
        return AutoCorrelationPeriodogram()

    @property
    def parameters(self) -> AutoCorrelationPeriodogramParams:
        return self._parameters

    @property
    def min_period(self) -> int:
        assert self._parameters.min_period is not None
        return self._parameters.min_period

    @property
    def max_period(self) -> int:
        return self._parameters.max_period

    @property
    def warmup_bars(self) -> int:
        return self._history_length - 1

    @property
    def samples_seen(self) -> int:
        return self._samples_seen

    @property
    def state(self) -> PeriodogramState:
        return self._state

    @property
    def output(self) -> PeriodogramOutput:
        return self._output

    def reset(self) -> None:
        """Reset all streaming state while preserving configuration."""
        first = self.min_period
        last = self.max_period
        self._samples_seen = 0
        self._close1 = 0.0
        self._close2 = 0.0
        self._high1 = 0.0
        self._high2 = 0.0
        self._filtered = [0.0] * self._history_length
        self._correlations = [0.0] * (last + 1)
        self._power = [0.0] * (last - first + 1)
        self._maximum = 0.0
        self._dominant = 0.0
        self._has_dominant = False
        self._state = PeriodogramState(
            0.0,
            0.0,
            tuple(self._correlations),
            tuple(self._power),
            0.0,
            0.0,
        )
        self._output = PeriodogramOutput(
            Heatmap.empty(None, float(first), float(last), 1.0),
            math.nan,
        )

    def is_primed(self) -> bool:
        """Return whether the filtered Pearson history is fully populated."""
        return self._samples_seen >= self._history_length

    def metadata(self) -> Metadata:
        """Describe the indicator and its heatmap and dominant-cycle outputs."""
        return build_metadata(
            Identifier.AUTO_CORRELATION_PERIODOGRAM,
            self.mnemonic,
            self.description,
            [
                OutputText(mnemonic=self.mnemonic, description=self.description),
                OutputText(
                    mnemonic=self._mnemonic_dc, description=self._description_dc
                ),
            ],
        )

    def update(
        self, sample: float, time: Optional[datetime.datetime] = None
    ) -> PeriodogramOutput:
        """Consume one scalar sample and return PeriodogramOutput."""
        price = _validate_sample(sample)
        first = self.min_period
        last = self.max_period
        averaging = self._parameters.averaging_length
        is_book = self._parameters.source == AutoCorrelationPeriodogramSource.BOOK
        enhance = self._parameters.enhance_resolution

        if math.isnan(price):
            return PeriodogramOutput(
                Heatmap.empty(time, float(first), float(last), 1.0),
                math.nan,
            )

        if not is_book:
            high = self._hp0 * (price - self._close1) + self._hp1 * self._high1
        else:
            high = (
                self._hp0 * (price - 2.0 * self._close1 + self._close2)
                + self._hp1 * self._high1
                - self._hp2 * self._high2
            )
        filtered_value = (
            self._ss1 * (high + self._high1) / 2.0
            + self._ss2 * self._filtered[0]
            + self._ss3 * self._filtered[1]
        )
        if not (math.isfinite(high) and math.isfinite(filtered_value)):
            raise ValueError("filter arithmetic overflowed")
        next_filtered = [filtered_value] + self._filtered[:-1]

        next_correlations = self._correlations[:]
        for lag in range(last + 1):
            window_size = averaging if averaging != 0 else lag
            sx = sy = sxx = syy = sxy = 0.0
            for offset in range(window_size):
                x_sample = next_filtered[offset]
                y_sample = next_filtered[lag + offset]
                sx += x_sample
                sy += y_sample
                sxx += x_sample * x_sample
                sxy += x_sample * y_sample
                syy += y_sample * y_sample
            denominator = (window_size * sxx - sx * sx) * (
                window_size * syy - sy * sy
            )
            if not all(
                math.isfinite(part)
                for part in (sx, sy, sxx, syy, sxy, denominator)
            ):
                raise ValueError("Pearson arithmetic overflowed")
            if denominator > 0.0:
                next_correlations[lag] = (window_size * sxy - sx * sy) / math.sqrt(
                    denominator
                )
            if not math.isfinite(next_correlations[lag]):
                raise ValueError("Pearson correlation overflowed")

        next_power: list[float] = []
        for bin_index in range(len(self._power)):
            cosine_part = 0.0
            sine_part = 0.0
            cosine_row = self._cosines[bin_index]
            sine_row = self._sines[bin_index]
            for offset, lag in enumerate(range(DFT_LAG_START, last + 1)):
                corr_value = next_correlations[lag]
                cosine_part += corr_value * cosine_row[offset]
                sine_part += corr_value * sine_row[offset]
            squared_sum = cosine_part * cosine_part + sine_part * sine_part
            smoothed = (
                POWER_SMOOTHING * squared_sum * squared_sum
                + POWER_RETENTION * self._power[bin_index]
            )
            if not math.isfinite(smoothed):
                raise ValueError("spectral arithmetic overflowed")
            next_power.append(smoothed)

        maximum = BOOK_AGC_DECAY * self._maximum if is_book else 0.0
        for power_value in next_power:
            if power_value > maximum:
                maximum = power_value
        normalized = [
            power_value / maximum if maximum > 0.0 else 0.0
            for power_value in next_power
        ]

        weights = normalized
        if not is_book and enhance:
            weights = [
                norm_value * norm_value * norm_value for norm_value in normalized
            ]
        threshold = BOOK_THRESHOLD if is_book else ARTICLE_THRESHOLD
        peak = max(weights)
        sp = 0.0
        spx = 0.0
        for bin_index, weight in enumerate(weights):
            if weight >= threshold and (is_book or peak >= ARTICLE_THRESHOLD):
                spx += (first + bin_index) * weight
                sp += weight

        dominant = self._dominant
        has_dominant = self._has_dominant
        if sp != 0.0 and (is_book or sp >= ARTICLE_THRESHOLD):
            dominant = spx / sp
            has_dominant = True
        if not math.isfinite(dominant):
            raise ValueError("dominant-cycle arithmetic overflowed")

        values = weights[:]
        if is_book and enhance:
            values = [norm_value * norm_value for norm_value in normalized]

        self._close2, self._close1 = self._close1, price
        self._high2, self._high1 = self._high1, high
        self._filtered = next_filtered
        self._correlations = next_correlations
        self._power = next_power
        self._maximum = maximum
        self._dominant = dominant
        self._has_dominant = has_dominant
        self._samples_seen += 1
        self._state = PeriodogramState(
            high,
            filtered_value,
            tuple(next_correlations),
            tuple(next_power),
            maximum,
            dominant,
        )

        if self.is_primed():
            value_min = min(values)
            value_max = max(values)
            public_dominant = dominant if has_dominant else math.nan
            self._output = PeriodogramOutput(
                Heatmap(
                    time,
                    float(first),
                    float(last),
                    1.0,
                    value_min,
                    value_max,
                    list(values),
                ),
                public_dominant,
            )
            return PeriodogramOutput(
                Heatmap(
                    time,
                    float(first),
                    float(last),
                    1.0,
                    value_min,
                    value_max,
                    values,
                ),
                public_dominant,
            )

        empty_out = PeriodogramOutput(
            Heatmap.empty(time, float(first), float(last), 1.0),
            math.nan,
        )
        self._output = empty_out
        return empty_out

    def _wrap(self, time: datetime.datetime, result: PeriodogramOutput) -> Output:
        return [result.heatmap, Scalar(time, result.dominant_cycle)]

    def update_scalar(self, sample: Scalar) -> Output:
        return self._wrap(sample.time, self.update(sample.value, sample.time))

    def update_bar(self, sample: Bar) -> Output:
        return self._wrap(sample.time, self.update(self._bar_func(sample), sample.time))

    def update_quote(self, sample: Quote) -> Output:
        return self._wrap(
            sample.time, self.update(self._quote_func(sample), sample.time)
        )

    def update_trade(self, sample: Trade) -> Output:
        return self._wrap(
            sample.time, self.update(self._trade_func(sample), sample.time)
        )
