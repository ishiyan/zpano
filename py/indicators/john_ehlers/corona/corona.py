"""Shared Corona spectral analysis engine and adaptive/raster primitives."""

from __future__ import annotations

import datetime
import math
from typing import NamedTuple, Optional

from ...core.outputs.heatmap import Heatmap
from .params import CoronaParams, default_params

TWO_PI = 2.0 * math.pi
SQRT_THREE_OVER_TWO = math.sqrt(3.0) / 2.0
FIRST_BANK_SAMPLE = 13
FIR_WEIGHTS = (1.0, 2.0, 3.0, 3.0, 2.0, 1.0)


def validate_integer_parameter(name: str, value: int, minimum: int) -> None:
    """Validate an integer parameter against its minimum bound (rejecting booleans)."""
    if isinstance(value, bool) or not isinstance(value, int) or value < minimum:
        raise ValueError(f"{name} must be an integer >= {minimum}")


def validate_finite_parameter(name: str, value: float) -> float:
    """Validate a finite numeric parameter (rejecting booleans)."""
    if isinstance(value, bool) or not isinstance(value, (int, float)) or not math.isfinite(value):
        raise ValueError(f"{name} must be finite")
    return float(value)


def valid_price(price: float) -> bool:
    """NaN skips a bar; other non-finite values are rejected before mutation."""
    if isinstance(price, bool) or not isinstance(price, (int, float)):
        raise ValueError("price must be finite or NaN")
    if math.isnan(price):
        return False
    if not math.isfinite(price):
        raise ValueError("price must be finite or NaN")
    return True


def checked(value: float) -> float:
    """Catch arithmetic overflow before mutating recursive state."""
    if not math.isfinite(value):
        raise ValueError("calculation overflow; use finite, reasonably scaled prices")
    return value


def clip(value: float, lower: float, upper: float) -> float:
    """Clamp value to [lower, upper]."""
    return max(lower, min(upper, value))


def rolling_append(history: list[float], value: float, length: int) -> tuple[float, float]:
    """Append value to bounded window and return (minimum, maximum)."""
    history.append(value)
    if len(history) > length:
        del history[0]
    return min(history), max(history)


class Filter(NamedTuple):
    """Per-bin state of a single bandpass filter channel in the bank."""

    in_phase_1: float = 0.0
    in_phase_2: float = 0.0
    quadrature_1: float = 0.0
    quadrature_2: float = 0.0
    real_1: float = 0.0
    real_2: float = 0.0
    imaginary_1: float = 0.0
    imaginary_2: float = 0.0
    power: float = 0.0
    decibels: float = 0.0


class CoronaSnapshot(NamedTuple):
    """Immutable snapshot produced by Corona.update()."""

    parameters: CoronaParams
    sample_count: int
    accepted: bool
    primed: bool
    price: float
    time: Optional[datetime.datetime]
    high_pass: float
    filtered: float
    bandwidth: float
    dominant_cycle: float
    dominant_cycle_median: float
    maximum_power: float
    normalization_power: float
    selected_bins: int
    has_cycle_estimate: bool
    power: tuple[float, ...]
    relative_power: tuple[float, ...]
    decibels: tuple[float, ...]


class Corona:
    """Streaming channelized receiver for Ehlers Corona indicators."""

    def __init__(self, params: Optional[CoronaParams] = None) -> None:
        source = default_params() if params is None else params
        validate_integer_parameter("high_pass_filter_cutoff", source.high_pass_filter_cutoff, 5)
        validate_integer_parameter("minimal_period", source.minimal_period, 6)
        validate_integer_parameter("maximal_period", source.maximal_period, source.minimal_period + 1)
        lower_threshold = validate_finite_parameter(
            "decibels_lower_threshold", source.decibels_lower_threshold
        )
        upper_threshold = validate_finite_parameter(
            "decibels_upper_threshold", source.decibels_upper_threshold
        )
        if not 0.0 <= lower_threshold < upper_threshold or upper_threshold <= 0.0:
            raise ValueError("require 0 <= decibels_lower_threshold < decibels_upper_threshold")

        angle = TWO_PI / source.high_pass_filter_cutoff
        pole = (1.0 - math.sin(angle)) / math.cos(angle)
        if not math.isfinite(pole) or not 0.0 < pole < 1.0:
            raise ValueError("high-pass coefficient must have a stable positive pole")

        self.parameters = CoronaParams(
            high_pass_filter_cutoff=source.high_pass_filter_cutoff,
            minimal_period=source.minimal_period,
            maximal_period=source.maximal_period,
            decibels_lower_threshold=lower_threshold,
            decibels_upper_threshold=upper_threshold,
        )
        self.minimal_period: int = self.parameters.minimal_period
        self.maximal_period: int = self.parameters.maximal_period
        self.minimal_period_times_two: int = 2 * self.minimal_period
        self.maximal_period_times_two: int = 2 * self.maximal_period
        self.periods: tuple[float, ...] = tuple(
            n / 2.0 for n in range(self.minimal_period_times_two, self.maximal_period_times_two + 1)
        )
        self.filter_bank_length: int = len(self.periods)
        self._betas: tuple[float, ...] = tuple(math.cos(TWO_PI / period) for period in self.periods)
        self._pole: float = pole
        self.reset()

    def reset(self) -> None:
        """Discard streaming history while retaining validated configuration."""
        self.sample_count: int = 0
        self._price_previous: float = 0.0
        self._hp_history: list[float] = [0.0] * 6
        self._filtered_previous: float = 0.0
        self._bank: list[Filter] = [Filter() for _ in self.periods]
        self._dc: float = 0.0
        self._dc_history: list[float] = [0.0] * 5
        self.last_snapshot: Optional[CoronaSnapshot] = None

    def is_primed(self) -> bool:
        """Return True once at least 13 accepted samples have been processed."""
        return self.sample_count >= FIRST_BANK_SAMPLE

    @property
    def dominant_cycle(self) -> float:
        """Current raw dominant cycle centroid."""
        return self._dc

    @property
    def dominant_cycle_median(self) -> float:
        """Current 5-sample median dominant cycle clamped to minimal_period."""
        if self.last_snapshot is not None:
            return self.last_snapshot.dominant_cycle_median
        return float(self.minimal_period)

    @property
    def maximum_power(self) -> float:
        """Current maximum channel power across the filter bank."""
        if self.last_snapshot is not None:
            return self.last_snapshot.maximum_power
        return 0.0

    def update(
        self, price: float, time: Optional[datetime.datetime] = None
    ) -> CoronaSnapshot:
        """Advance the spectral engine by one sample and return an immutable snapshot."""
        if not valid_price(price):
            if self.last_snapshot is not None:
                return self.last_snapshot._replace(accepted=False, price=float(price), time=time)
            zeros = (0.0,) * self.filter_bank_length
            return CoronaSnapshot(
                parameters=self.parameters,
                sample_count=0,
                accepted=False,
                primed=False,
                price=float(price),
                time=time,
                high_pass=0.0,
                filtered=0.0,
                bandwidth=0.5,
                dominant_cycle=0.0,
                dominant_cycle_median=float(self.parameters.minimal_period),
                maximum_power=0.0,
                normalization_power=0.01,
                selected_bins=0,
                has_cycle_estimate=False,
                power=zeros,
                relative_power=zeros,
                decibels=zeros,
            )

        sample = float(price)
        count = self.sample_count + 1
        previous = sample if count == 1 else self._price_previous
        hp = checked(
            0.5 * (1.0 + self._pole) * (sample - previous)
            + self._pole * self._hp_history[-1]
        )
        hp_history = self._hp_history[1:] + [hp]
        filtered = checked(
            sum(item * weight for item, weight in zip(hp_history, FIR_WEIGHTS)) / 12.0
        )
        if count < 7:
            filtered = checked(sample - previous)
        if count == 1:
            filtered = 0.0
        bandwidth = max(0.1, 0.5 - 0.015 * count)

        bank: list[Filter] = []
        for period, beta, old in zip(self.periods, self._betas, self._bank):
            if count < FIRST_BANK_SAMPLE:
                bank.append(old)
                continue
            gamma = 1.0 / math.cos(2.0 * TWO_PI * bandwidth / period)
            alpha = gamma - math.sqrt(gamma * gamma - 1.0)
            in_phase = filtered
            quadrature = checked((period / TWO_PI) * (filtered - self._filtered_previous))
            real = checked(
                0.5 * (1.0 - alpha) * (in_phase - old.in_phase_2)
                + beta * (1.0 + alpha) * old.real_1
                - alpha * old.real_2
            )
            imaginary = checked(
                0.5 * (1.0 - alpha) * (quadrature - old.quadrature_2)
                + beta * (1.0 + alpha) * old.imaginary_1
                - alpha * old.imaginary_2
            )
            power = checked(real * real + imaginary * imaginary)
            bank.append(
                Filter(
                    in_phase_1=in_phase,
                    in_phase_2=old.in_phase_1,
                    quadrature_1=quadrature,
                    quadrature_2=old.quadrature_1,
                    real_1=real,
                    real_2=old.real_1,
                    imaginary_1=imaginary,
                    imaginary_2=old.imaginary_1,
                    power=power,
                    decibels=old.decibels,
                )
            )

        maximum = max(channel.power for channel in bank)
        first_power = bank[0].power if bank[0].power != 0.0 else 0.01
        normalizer = max(first_power, maximum)

        numerator = 0.0
        denominator = 0.0
        selected = 0
        lower_threshold = self.parameters.decibels_lower_threshold
        upper_threshold = self.parameters.decibels_upper_threshold
        for i, (period, channel) in enumerate(zip(self.periods, bank)):
            relative = channel.power / normalizer
            raw = channel.decibels
            if relative > 0.0:
                raw = 10.0 * math.log10((1.0 - 0.99 * relative) / 0.01)
            decibels = min(upper_threshold, 0.33 * raw + 0.67 * channel.decibels)
            bank[i] = channel._replace(decibels=decibels)
            if decibels <= lower_threshold:
                weight = upper_threshold - decibels
                numerator += period * weight
                denominator += weight
                selected += 1

        dc = numerator / denominator if denominator > 0.0 else self._dc
        dc_history = self._dc_history[1:] + [dc]
        median = max(float(self.parameters.minimal_period), sorted(dc_history)[2])
        powers = tuple(channel.power for channel in bank)
        snapshot = CoronaSnapshot(
            parameters=self.parameters,
            sample_count=count,
            accepted=True,
            primed=count >= FIRST_BANK_SAMPLE,
            price=sample,
            time=time,
            high_pass=hp,
            filtered=filtered,
            bandwidth=bandwidth,
            dominant_cycle=dc,
            dominant_cycle_median=median,
            maximum_power=maximum,
            normalization_power=normalizer,
            selected_bins=selected,
            has_cycle_estimate=maximum > 0.0 and selected > 0,
            power=powers,
            relative_power=tuple(power / normalizer for power in powers),
            decibels=tuple(channel.decibels for channel in bank),
        )

        self.sample_count = count
        self._price_previous = sample
        self._hp_history = hp_history
        self._filtered_previous = filtered
        self._bank = bank
        self._dc = dc
        self._dc_history = dc_history
        self.last_snapshot = snapshot
        return snapshot


class AdaptiveCycle(NamedTuple):
    """Output of AdaptiveBandPass.update()."""

    band_pass: float
    quadrature: float
    amplitude: float
    lead60: float


class AdaptiveBandPass:
    """Two-pole bandpass filter tuned each bar to the dominant cycle median period."""

    def __init__(self) -> None:
        self.reset()

    def reset(self) -> None:
        self._prices: list[float] = []
        self._bp1: float = 0.0
        self._bp2: float = 0.0

    def update(self, price: float, period: float) -> AdaptiveCycle:
        if len(self._prices) < 2:
            self._prices.append(price)
            return AdaptiveCycle(0.0, 0.0, 0.0, 0.0)
        omega = TWO_PI / period
        beta = math.cos(omega)
        gamma = 1.0 / math.cos(2.0 * omega * 0.1)
        alpha = gamma - math.sqrt(gamma * gamma - 1.0)
        bp = checked(
            0.5 * (1.0 - alpha) * (price - self._prices[0])
            + beta * (1.0 + alpha) * self._bp1
            - alpha * self._bp2
        )
        quadrature = checked((bp - self._bp1) / omega)
        amplitude = checked(math.sqrt(bp * bp + quadrature * quadrature))
        lead60 = checked(0.5 * bp + SQRT_THREE_OVER_TWO * quadrature)
        self._prices = [self._prices[1], price]
        self._bp2 = self._bp1
        self._bp1 = bp
        return AdaptiveCycle(bp, quadrature, amplitude, lead60)


class Raster:
    """Corona intensity raster for Swing Position, Trend Vigor, and Signal-to-Noise Ratio.

    Logical cells j=1..length sit at display_min + j * step, so the first bin is
    display_min + step and the last bin is display_max. In zpano's Heatmap convention,
    parameter_resolution = 1.0 / step (bins per unit).
    """

    def __init__(self, length: int, lower: float, upper: float, ceiling: float) -> None:
        validate_integer_parameter("raster_length", length, 2)
        self.lower: float = validate_finite_parameter("min_parameter_value", lower)
        self.upper: float = validate_finite_parameter("max_parameter_value", upper)
        self.ceiling: float = validate_finite_parameter("max_raster_value", ceiling)
        if self.upper <= self.lower or self.ceiling <= 0.0:
            raise ValueError("require max_parameter_value > min_parameter_value and max_raster_value > 0")
        self.length: int = length
        self.step: float = (self.upper - self.lower) / float(length)
        self.first: float = self.lower + self.step
        self.resolution: float = 1.0 / self.step
        self.values: list[float] = [0.0] * length

    def reset(self) -> None:
        self.values = [0.0] * self.length

    def empty(self, time: Optional[datetime.datetime]) -> Heatmap:
        return Heatmap.empty(time, self.first, self.upper, self.resolution)

    def column(self, time: Optional[datetime.datetime]) -> Heatmap:
        values = list(self.values)
        return Heatmap(
            time, self.first, self.upper, self.resolution, min(values), max(values), values
        )

    def update(self, position: float, width: float, kind: str, suppress: bool = False) -> None:
        marker = int(math.floor(self.length * position + 0.5))
        exponent, scale, history_scale = {
            "swing": (0.95, 0.5, 0.5),
            "trend": (0.85, 0.8, 0.2),
            "snr": (0.8, 0.5, 1.0),
        }[kind]
        updated: list[float] = []
        for i, old in enumerate(self.values):
            j = i + 1
            if suppress:
                value = self.ceiling
            elif j == marker:
                value = 0.5 * old
            elif width <= 0.0:
                value = self.ceiling
            else:
                distance = abs(self.ceiling * position - self.ceiling * j / self.length) / width
                if kind == "snr" and j > marker and distance <= 1.0:
                    value = self.ceiling
                else:
                    value = scale * (math.pow(distance, exponent) + history_scale * old)
            updated.append(clip(value, 0.0, self.ceiling))
        self.values = updated


def accept_snapshot(
    snapshot: CoronaSnapshot,
    expected_parameters: CoronaParams,
    expected_next_count: int,
) -> bool:
    """Validate snapshot compatibility and monotonic accepted sample ordering."""
    if not isinstance(snapshot, CoronaSnapshot) or snapshot.parameters != expected_parameters:
        raise ValueError("snapshot must use this consumer's Corona parameters")
    if not snapshot.accepted:
        return False
    if snapshot.sample_count != expected_next_count:
        raise ValueError("consume every accepted snapshot exactly once, in order from sample 1")
    return True


def format_number(value: float) -> str:
    """Format a float parameter cleanly for mnemonics (avoiding '-0')."""
    normalized = 0.0 if value == 0.0 else value
    return f"{normalized:g}"
