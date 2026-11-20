"""Ehlers' January 2007 Transformed Discrete Fourier Transform Spectrum indicator."""

from __future__ import annotations

import datetime
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
    TransformedDiscreteFourierTransformSpectrumParams,
    default_params,
    transformed_discrete_fourier_transform_spectrum_dominant_cycle_mnemonic,
    transformed_discrete_fourier_transform_spectrum_mnemonic,
)

TWO_PI: float = 2.0 * math.pi
LOG_TEN: float = math.log(10.0)
DB_CAP: float = 20.0
CENTROID_DB_LIMIT: float = 3.0


class SpectrumOutput(NamedTuple):
    """Spectrum heatmap column and dominant cycle period returned per observation."""

    heatmap: Heatmap
    dominant_cycle: float


class SpectrumState(NamedTuple):
    """Immutable diagnostic snapshot of committed internal spectral state."""

    high_pass: float
    cleaned_data: float
    maximum_power: float
    dominant_cycle: float
    cosine_parts: tuple[float, ...]
    sine_parts: tuple[float, ...]
    powers: tuple[float, ...]
    decibels: tuple[float, ...]


def _validate_integer(value: object, name: str, minimum: int) -> int:
    if isinstance(value, bool) or not isinstance(value, int) or value < minimum:
        raise ValueError(f"{name} must be an integer >= {minimum}")
    return value


def _validate_real_above(value: object, name: str, lower_bound: float) -> float:
    if isinstance(value, bool) or not isinstance(value, (int, float)):
        raise ValueError(f"{name} must be a finite real number > {lower_bound}")
    try:
        number = float(value)
    except (ValueError, OverflowError) as error:
        raise ValueError(
            f"{name} must be a finite real number > {lower_bound}"
        ) from error
    if not math.isfinite(number) or number <= lower_bound:
        raise ValueError(f"{name} must be a finite real number > {lower_bound}")
    return number


def _validate_sample(value: object) -> float:
    if isinstance(value, bool) or not isinstance(value, (int, float)):
        raise ValueError("input must be a real number; NaN may mark missing data")
    try:
        price = float(value)
    except (ValueError, OverflowError) as error:
        raise ValueError("input cannot be represented as a float") from error
    if math.isinf(price):
        raise ValueError("input must not be infinite")
    return price


class TransformedDiscreteFourierTransformSpectrum(Indicator):
    """Ehlers' January 2007 Transformed Discrete Fourier Transform Spectrum (TDFTS)."""

    def __init__(
        self,
        params: Optional[TransformedDiscreteFourierTransformSpectrumParams] = None,
    ) -> None:
        source = default_params() if params is None else params

        first = _validate_integer(source.min_period, "min_period", 8)
        last = _validate_integer(source.max_period, "max_period", first)
        window = _validate_integer(source.window_length, "window_length", 1)
        hp_period = _validate_real_above(
            source.highpass_period, "highpass_period", 4.0
        )

        omega = TWO_PI / hp_period
        cosine_omega = math.cos(omega)
        if cosine_omega == 0.0:
            raise ValueError("highpass_period produces a singular high-pass pole")
        highpass_alpha = (1.0 - math.sin(omega)) / cosine_omega
        if not math.isfinite(highpass_alpha):
            raise ValueError("highpass_period must produce a finite high-pass pole")

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

        self._parameters = TransformedDiscreteFourierTransformSpectrumParams(
            min_period=first,
            max_period=last,
            window_length=window,
            highpass_period=hp_period,
            bar_component=source.bar_component,
            quote_component=source.quote_component,
            trade_component=source.trade_component,
        )
        self._periods: tuple[int, ...] = tuple(range(first, last + 1))
        self._highpass_alpha: float = highpass_alpha
        self._cosines: tuple[tuple[float, ...], ...] = tuple(
            tuple(math.cos(TWO_PI * lag / period) for lag in range(window))
            for period in self._periods
        )
        self._sines: tuple[tuple[float, ...], ...] = tuple(
            tuple(math.sin(TWO_PI * lag / period) for lag in range(window))
            for period in self._periods
        )
        self._warmup_bars: int = max(5, window - 1)

        self.mnemonic: str = (
            transformed_discrete_fourier_transform_spectrum_mnemonic(self._parameters)
        )
        self.description: str = (
            f"Transformed discrete Fourier transform spectrum {self.mnemonic}"
        )
        self._mnemonic_dc: str = (
            transformed_discrete_fourier_transform_spectrum_dominant_cycle_mnemonic(
                self._parameters
            )
        )
        self._description_dc: str = (
            "Transformed discrete Fourier transform spectrum dominant cycle "
            f"{self._mnemonic_dc}"
        )

        self._bar_func: Callable[[Bar], float] = bar_component_value(bar_comp)
        self._quote_func: Callable[[Quote], float] = quote_component_value(quote_comp)
        self._trade_func: Callable[[Trade], float] = trade_component_value(trade_comp)

        self._samples_seen: int = 0
        self._previous_price: float = 0.0
        self._hp_history: list[float] = []
        self._cleaned: float = 0.0
        self._window: list[float] = []
        self._cosine_parts: list[float] = []
        self._sine_parts: list[float] = []
        self._powers: list[float] = []
        self._decibels: list[float] = []
        self._maximum: float = 0.0
        self._dominant: float = 0.0
        self._output: SpectrumOutput = self._unavailable(None)
        self.reset()

    @staticmethod
    def create(
        params: Optional[TransformedDiscreteFourierTransformSpectrumParams] = None,
    ) -> TransformedDiscreteFourierTransformSpectrum:
        """Construct an instance from validated parameters."""
        return TransformedDiscreteFourierTransformSpectrum(params)

    @staticmethod
    def create_default() -> TransformedDiscreteFourierTransformSpectrum:
        """Construct an instance using Ehlers' January 2007 article defaults."""
        return TransformedDiscreteFourierTransformSpectrum(default_params())

    @property
    def parameters(self) -> TransformedDiscreteFourierTransformSpectrumParams:
        return self._parameters

    @property
    def min_period(self) -> int:
        return self._parameters.min_period

    @property
    def max_period(self) -> int:
        return self._parameters.max_period

    @property
    def highpass_alpha(self) -> float:
        return self._highpass_alpha

    @property
    def warmup_bars(self) -> int:
        return self._warmup_bars

    @property
    def samples_seen(self) -> int:
        return self._samples_seen

    @property
    def state(self) -> SpectrumState:
        return SpectrumState(
            self._hp_history[0],
            self._cleaned,
            self._maximum,
            self._dominant,
            tuple(self._cosine_parts),
            tuple(self._sine_parts),
            tuple(self._powers),
            tuple(self._decibels),
        )

    @property
    def output(self) -> SpectrumOutput:
        return self._output

    def _unavailable(self, time: Any) -> SpectrumOutput:
        return SpectrumOutput(
            Heatmap.empty(
                time,
                float(self._parameters.min_period),
                float(self._parameters.max_period),
                1.0,
            ),
            math.nan,
        )

    def reset(self) -> None:
        """Reset all streaming state while preserving configuration."""
        channel_count = len(self._periods)
        self._samples_seen = 0
        self._previous_price = 0.0
        self._hp_history = [0.0] * 6
        self._cleaned = 0.0
        self._window = [0.0] * self._parameters.window_length
        self._cosine_parts = [0.0] * channel_count
        self._sine_parts = [0.0] * channel_count
        self._powers = [0.0] * channel_count
        self._decibels = [0.0] * channel_count
        self._maximum = 0.0
        self._dominant = 0.0
        self._output = self._unavailable(None)

    def is_primed(self) -> bool:
        """Return whether both the 6-tap filter and Fourier window have warmed up."""
        return self._samples_seen > self._warmup_bars

    def metadata(self) -> Metadata:
        """Describe the indicator and its heatmap and dominant-cycle outputs."""
        return build_metadata(
            Identifier.TRANSFORMED_DISCRETE_FOURIER_TRANSFORM_SPECTRUM,
            self.mnemonic,
            self.description,
            [
                OutputText(mnemonic=self.mnemonic, description=self.description),
                OutputText(
                    mnemonic=self._mnemonic_dc, description=self._description_dc
                ),
            ],
        )

    def update(self, sample: float, time: Any = None) -> SpectrumOutput:
        """Consume one scalar price and return a SpectrumOutput snapshot."""
        price = _validate_sample(sample)
        if math.isnan(price):
            return self._unavailable(time)

        bar_number = self._samples_seen + 1
        if bar_number <= 5:
            high_pass = price
            cleaned = price
        else:
            high_pass = (
                0.5 * (1.0 + self._highpass_alpha) * (price - self._previous_price)
                + self._highpass_alpha * self._hp_history[0]
            )
            cleaned = (
                high_pass
                + 2.0 * self._hp_history[0]
                + 3.0 * self._hp_history[1]
                + 3.0 * self._hp_history[2]
                + 2.0 * self._hp_history[3]
                + self._hp_history[4]
            ) / 12.0

        window_length = self._parameters.window_length
        next_window = [cleaned] + self._window[: window_length - 1]
        cosine_parts: list[float] = []
        sine_parts: list[float] = []
        powers: list[float] = []
        for channel_index in range(len(self._periods)):
            cosine_sum = 0.0
            sine_sum = 0.0
            cosine_row = self._cosines[channel_index]
            sine_row = self._sines[channel_index]
            for lag in range(window_length):
                cosine_sum += next_window[lag] * cosine_row[lag]
                sine_sum += next_window[lag] * sine_row[lag]
            cosine_parts.append(cosine_sum)
            sine_parts.append(sine_sum)
            powers.append(cosine_sum * cosine_sum + sine_sum * sine_sum)

        maximum = max(powers)
        decibels = self._decibels.copy()
        for channel_index, power in enumerate(powers):
            if maximum > 0.0 and power > 0.0:
                decibels[channel_index] = (
                    -10.0 * math.log(0.01 / (1.0 - 0.99 * power / maximum)) / LOG_TEN
                )
            if decibels[channel_index] > DB_CAP:
                decibels[channel_index] = DB_CAP

        numerator = 0.0
        denominator = 0.0
        for period, channel_db in zip(self._periods, decibels):
            if channel_db < CENTROID_DB_LIMIT:
                weight = CENTROID_DB_LIMIT - channel_db
                numerator += period * weight
                denominator += weight
        dominant = numerator / denominator if denominator != 0.0 else self._dominant

        candidates = (
            [high_pass, cleaned, maximum, dominant]
            + cosine_parts
            + sine_parts
            + powers
            + decibels
        )
        if not all(math.isfinite(item) for item in candidates):
            raise ValueError(
                "input causes non-finite filter arithmetic; state was not changed"
            )

        self._previous_price = price
        self._hp_history = [high_pass] + self._hp_history[:5]
        self._cleaned = cleaned
        self._window = next_window
        self._cosine_parts = cosine_parts
        self._sine_parts = sine_parts
        self._powers = powers
        self._decibels = decibels
        self._maximum = maximum
        self._dominant = dominant
        self._samples_seen = bar_number

        if self.is_primed():
            heatmap = Heatmap(
                time,
                float(self._parameters.min_period),
                float(self._parameters.max_period),
                1.0,
                min(decibels),
                max(decibels),
                decibels.copy(),
            )
            result = SpectrumOutput(heatmap, dominant)
        else:
            result = self._unavailable(time)

        self._output = result
        return result

    def _wrap(self, time: datetime.datetime, result: SpectrumOutput) -> Output:
        return [result.heatmap, Scalar(time, result.dominant_cycle)]

    def update_scalar(self, sample: Scalar) -> Output:
        return self._wrap(sample.time, self.update(sample.value, sample.time))

    def update_bar(self, sample: Bar) -> Output:
        return self._wrap(
            sample.time, self.update(self._bar_func(sample), sample.time)
        )

    def update_quote(self, sample: Quote) -> Output:
        return self._wrap(
            sample.time, self.update(self._quote_func(sample), sample.time)
        )

    def update_trade(self, sample: Trade) -> Output:
        return self._wrap(
            sample.time, self.update(self._trade_func(sample), sample.time)
        )
