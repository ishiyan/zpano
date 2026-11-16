"""Ehlers' Discrete Fourier Transform Spectrum indicator (2013 Book Listing 9-1)."""

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
    DiscreteFourierTransformSpectrumParams,
    default_params,
    discrete_fourier_transform_spectrum_dominant_cycle_mnemonic,
    discrete_fourier_transform_spectrum_mnemonic,
)

TWO_PI = 2.0 * math.pi
SQRT_TWO = math.sqrt(2.0)
AGC_DECAY = 0.995
CENTROID_THRESHOLD = 0.5


class SpectrumOutput(NamedTuple):
    """Spectrum heatmap and dominant cycle period returned per update."""

    heatmap: Heatmap
    dominant_cycle: float


class SpectrumState(NamedTuple):
    """Internal committed spectral state snapshot."""

    high_pass: float
    filtered: float
    raw_power: tuple[float, ...]
    normalization_maximum: float
    internal_dominant_cycle: float


def _validate_integer(value: int, name: str, minimum: int) -> int:
    if isinstance(value, bool) or not isinstance(value, int) or value < minimum:
        raise ValueError(f"{name} must be an integer >= {minimum}")
    return value


def _validate_real(value: float, name: str, minimum: float) -> float:
    if isinstance(value, bool) or not isinstance(value, (int, float)):
        raise ValueError(f"{name} must be a finite real >= {minimum}")
    numeric = float(value)
    if not math.isfinite(numeric) or numeric < minimum:
        raise ValueError(f"{name} must be a finite real >= {minimum}")
    return numeric


def _validate_boolean(value: bool, name: str) -> bool:
    if not isinstance(value, bool):
        raise ValueError(f"{name} must be a boolean")
    return value


def _validate_sample(value: float) -> float:
    if isinstance(value, bool) or not isinstance(value, (int, float)):
        raise ValueError("price must be a real number")
    numeric = float(value)
    if math.isinf(numeric):
        raise ValueError("price must be finite or NaN (missing)")
    return numeric


class _Roof:
    """Two-pole high-pass followed by two-pole SuperSmoother (Book Equation 3-3)."""

    def __init__(self, highpass_period: float, smoothing_period: float) -> None:
        hp = _validate_real(highpass_period, "highpass_period", 5.0)
        lp = _validate_real(smoothing_period, "smoothing_period", 2.0)
        angle = TWO_PI / (SQRT_TWO * hp)
        cosine = math.cos(angle)
        alpha = (cosine + math.sin(angle) - 1.0) / cosine
        pole = 1.0 - alpha
        a = math.exp(-SQRT_TWO * math.pi / lp)
        self.hp0 = (1.0 - alpha / 2.0) * (1.0 - alpha / 2.0)
        self.hp1 = 2.0 * pole
        self.hp2 = pole * pole
        self.ss2 = 2.0 * a * math.cos(SQRT_TWO * math.pi / lp)
        self.ss3 = -a * a
        self.ss1 = 1.0 - self.ss2 - self.ss3
        if not (0.0 < pole < 1.0 and 0.0 < a < 1.0 and self.ss1 > 0.0):
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


def _normalize_and_centroid(
    periods: tuple[int, ...],
    raw: list[float],
    previous_max: float,
    previous_dc: float,
    has_dc: bool,
    enhance: bool,
) -> tuple[float, float, bool, list[float]]:
    maximum = AGC_DECAY * previous_max
    for power in raw:
        if not math.isfinite(power) or power < 0.0:
            raise ValueError("spectral power arithmetic overflowed")
        if power > maximum:
            maximum = power
    normalized = [power / maximum if maximum > 0.0 else 0.0 for power in raw]
    sp = 0.0
    spx = 0.0
    for period, power in zip(periods, normalized):
        if power >= CENTROID_THRESHOLD:
            spx = spx + period * power
            sp = sp + power
    dominant = previous_dc
    if sp != 0.0:
        dominant = spx / sp
        has_dc = True
    if not math.isfinite(dominant):
        raise ValueError("centroid arithmetic overflowed")
    values = [value * value * value for value in normalized] if enhance else normalized
    return maximum, dominant, has_dc, values


class DiscreteFourierTransformSpectrum(Indicator):
    """Ehlers' Discrete Fourier Transform Spectrum (2013 Book Listing 9-1)."""

    def __init__(
        self, params: Optional[DiscreteFourierTransformSpectrumParams] = None
    ) -> None:
        source = default_params() if params is None else params

        first = _validate_integer(source.min_period, "min_period", 3)
        last = _validate_integer(source.max_period, "max_period", first)
        length = _validate_integer(source.window_length, "window_length", 2)
        hp = _validate_real(source.highpass_period, "highpass_period", 5.0)
        lp = _validate_real(source.smoothing_period, "smoothing_period", 2.0)
        comp = _validate_boolean(
            source.spectral_dilation_compensation, "spectral_dilation_compensation"
        )
        enhance = _validate_boolean(source.enhance_resolution, "enhance_resolution")

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

        self._roof = _Roof(hp, lp)
        self._parameters = DiscreteFourierTransformSpectrumParams(
            min_period=first,
            max_period=last,
            window_length=length,
            highpass_period=hp,
            smoothing_period=lp,
            spectral_dilation_compensation=comp,
            enhance_resolution=enhance,
            bar_component=source.bar_component,
            quote_component=source.quote_component,
            trade_component=source.trade_component,
        )
        self._periods: tuple[int, ...] = tuple(range(first, last + 1))
        self._cosines: list[tuple[float, ...]] = [
            tuple(math.cos(TWO_PI * n / period) for n in range(length))
            for period in self._periods
        ]
        self._sines: list[tuple[float, ...]] = [
            tuple(math.sin(TWO_PI * n / period) for n in range(length))
            for period in self._periods
        ]

        self._bar_func = bar_component_value(bar_comp)
        self._quote_func = quote_component_value(quote_comp)
        self._trade_func = trade_component_value(trade_comp)

        self.mnemonic: str = discrete_fourier_transform_spectrum_mnemonic(
            self._parameters
        )
        self.description: str = "Discrete Fourier transform spectrum " + self.mnemonic
        self._mnemonic_dc: str = (
            discrete_fourier_transform_spectrum_dominant_cycle_mnemonic(self._parameters)
        )
        self._description_dc: str = (
            "Discrete Fourier transform spectrum dominant cycle " + self._mnemonic_dc
        )

        self._count: int = 0
        self._history: list[float] = []
        self._raw: list[float] = []
        self._maximum: float = 0.0
        self._dominant: float = 0.0
        self._has_dominant: bool = False
        self._output: SpectrumOutput = SpectrumOutput(
            Heatmap.empty(None, float(first), float(last), 1.0), math.nan
        )
        self.reset()

    @staticmethod
    def create(
        params: Optional[DiscreteFourierTransformSpectrumParams] = None,
    ) -> "DiscreteFourierTransformSpectrum":
        """Create a new DiscreteFourierTransformSpectrum from parameters."""
        return DiscreteFourierTransformSpectrum(params)

    @staticmethod
    def create_default() -> "DiscreteFourierTransformSpectrum":
        """Create a new DiscreteFourierTransformSpectrum with default parameters."""
        return DiscreteFourierTransformSpectrum()

    @property
    def parameters(self) -> DiscreteFourierTransformSpectrumParams:
        return self._parameters

    @property
    def min_period(self) -> int:
        return self._parameters.min_period

    @property
    def max_period(self) -> int:
        return self._parameters.max_period

    @property
    def warmup_bars(self) -> int:
        return self._parameters.window_length - 1

    @property
    def samples_seen(self) -> int:
        return self._count

    @property
    def state(self) -> SpectrumState:
        return SpectrumState(
            self._roof.h1,
            self._roof.f1,
            tuple(self._raw),
            self._maximum,
            self._dominant,
        )

    @property
    def output(self) -> SpectrumOutput:
        return self._output

    def reset(self) -> None:
        """Reset all streaming state while preserving configuration."""
        self._roof.reset()
        self._count = 0
        self._history = [0.0] * self._parameters.window_length
        self._raw = [0.0] * len(self._periods)
        self._maximum = 0.0
        self._dominant = 0.0
        self._has_dominant = False
        self._output = SpectrumOutput(
            Heatmap.empty(None, float(self.min_period), float(self.max_period), 1.0),
            math.nan,
        )

    def is_primed(self) -> bool:
        """Return whether the window has accumulated window_length valid observations."""
        return self._count >= self._parameters.window_length

    def metadata(self) -> Metadata:
        """Describe the indicator and its heatmap and dominant-cycle outputs."""
        return build_metadata(
            Identifier.DISCRETE_FOURIER_TRANSFORM_SPECTRUM,
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
    ) -> SpectrumOutput:
        """Update the indicator with a scalar sample and return SpectrumOutput."""
        price = _validate_sample(sample)
        params = self._parameters
        if math.isnan(price):
            return SpectrumOutput(
                Heatmap.empty(
                    time, float(params.min_period), float(params.max_period), 1.0
                ),
                math.nan,
            )

        high, filtered = self._roof.candidate(price)
        history = [filtered] + self._history[:-1]
        raw: list[float] = []
        for index, period in enumerate(self._periods):
            compensation = (
                float(period) if params.spectral_dilation_compensation else 1.0
            )
            cosine_part = 0.0
            sine_part = 0.0
            cosine_row = self._cosines[index]
            sine_row = self._sines[index]
            for n in range(params.window_length):
                cosine_part = cosine_part + history[n] * cosine_row[n] / compensation
                sine_part = sine_part + history[n] * sine_row[n] / compensation
            raw.append(cosine_part * cosine_part + sine_part * sine_part)

        maximum, dominant, has_dc, values = _normalize_and_centroid(
            self._periods,
            raw,
            self._maximum,
            self._dominant,
            self._has_dominant,
            params.enhance_resolution,
        )

        self._roof.commit(price, high, filtered)
        self._history = history
        self._raw = raw
        self._maximum = maximum
        self._dominant = dominant
        self._has_dominant = has_dc
        self._count += 1

        if self.is_primed():
            heatmap = Heatmap(
                time,
                float(params.min_period),
                float(params.max_period),
                1.0,
                min(values),
                max(values),
                list(values),
            )
            self._output = SpectrumOutput(
                heatmap, dominant if has_dc else math.nan
            )
        else:
            self._output = SpectrumOutput(
                Heatmap.empty(
                    time, float(params.min_period), float(params.max_period), 1.0
                ),
                math.nan,
            )
        return self._output

    def _wrap(self, time: datetime.datetime, result: SpectrumOutput) -> Output:
        return [result.heatmap, Scalar(time, result.dominant_cycle)]

    def update_scalar(self, sample: Scalar) -> Output:
        return self._wrap(sample.time, self.update(sample.value, sample.time))

    def update_bar(self, sample: Bar) -> Output:
        return self._wrap(sample.time, self.update(self._bar_func(sample), sample.time))

    def update_quote(self, sample: Quote) -> Output:
        return self._wrap(sample.time, self.update(self._quote_func(sample), sample.time))

    def update_trade(self, sample: Trade) -> Output:
        return self._wrap(sample.time, self.update(self._trade_func(sample), sample.time))
