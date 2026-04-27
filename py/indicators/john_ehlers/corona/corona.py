"""Corona spectral analysis engine.

This is the shared engine consumed by CoronaSpectrum, CoronaSignalToNoiseRatio,
CoronaSwingPosition, and CoronaTrendVigor. It is NOT a registered indicator.
"""

import math
from dataclasses import dataclass, field
from typing import Optional

from .params import CoronaParams

# Constants matching Go implementation.
_HP_BUFFER_SIZE = 6
_FIR_COEF_SUM = 12.0

_DELTA_LOWER_THRESHOLD = 0.1
_DELTA_FACTOR = -0.015
_DELTA_SUMMAND = 0.5

_DC_BUFFER_SIZE = 5
_DC_MEDIAN_INDEX = 2

_DB_SMOOTHING_ALPHA = 0.33
_DB_SMOOTHING_ONE_MINUS = 0.67

_NORMALIZED_AMPLITUDE_FACTOR = 0.99
_DECIBELS_FLOOR = 0.01
_DECIBELS_GAIN = 10.0


@dataclass
class Filter:
    """Per-bin state of a single bandpass filter in the bank."""

    in_phase: float = 0.0
    in_phase_previous: float = 0.0
    quadrature: float = 0.0
    quadrature_previous: float = 0.0
    real: float = 0.0
    real_previous: float = 0.0
    imaginary: float = 0.0
    imaginary_previous: float = 0.0
    amplitude_squared: float = 0.0
    decibels: float = 0.0


class Corona:
    """Shared spectral-analysis engine for Ehlers Corona indicators.

    Call update(sample) once per bar. Read is_primed(), dominant_cycle,
    dominant_cycle_median, maximal_amplitude_squared, and filter_bank.
    """

    def __init__(self, params: Optional[CoronaParams] = None) -> None:
        cfg = CoronaParams() if params is None else CoronaParams(
            high_pass_filter_cutoff=params.high_pass_filter_cutoff,
            minimal_period=params.minimal_period,
            maximal_period=params.maximal_period,
            decibels_lower_threshold=params.decibels_lower_threshold,
            decibels_upper_threshold=params.decibels_upper_threshold,
        )
        _apply_defaults(cfg)
        _verify_parameters(cfg)

        self.minimal_period: int = cfg.minimal_period
        self.maximal_period: int = cfg.maximal_period
        self.minimal_period_times_two: int = cfg.minimal_period * 2
        self.maximal_period_times_two: int = cfg.maximal_period * 2
        self.decibels_lower_threshold: float = cfg.decibels_lower_threshold
        self.decibels_upper_threshold: float = cfg.decibels_upper_threshold

        self.filter_bank_length: int = self.maximal_period_times_two - self.minimal_period_times_two + 1
        self.filter_bank: list[Filter] = [Filter() for _ in range(self.filter_bank_length)]

        # Dominant cycle buffer initialized with MaxFloat64 sentinels.
        self._dc_buffer: list[float] = [math.inf] * _DC_BUFFER_SIZE
        self.dominant_cycle: float = math.inf
        self.dominant_cycle_median: float = math.inf

        # High-pass filter coefficients.
        phi = 2.0 * math.pi / cfg.high_pass_filter_cutoff
        self._alpha: float = (1.0 - math.sin(phi)) / math.cos(phi)
        self._half_one_plus_alpha: float = 0.5 * (1.0 + self._alpha)

        # Pre-calculated beta = cos(4π/n) for each half-period index.
        self._pre_calculated_beta: list[float] = []
        for index in range(self.filter_bank_length):
            n = self.minimal_period_times_two + index
            self._pre_calculated_beta.append(math.cos(4.0 * math.pi / n))

        # HP ring buffer.
        self._hp_buffer: list[float] = [0.0] * _HP_BUFFER_SIZE
        self._sample_previous: float = 0.0
        self._smooth_hp_previous: float = 0.0

        self.maximal_amplitude_squared: float = 0.0
        self._sample_count: int = 0
        self._primed: bool = False

    def is_primed(self) -> bool:
        """Whether the engine has seen enough samples for meaningful output."""
        return self._primed

    def update(self, sample: float) -> bool:
        """Feed the next sample. Returns True once primed."""
        if math.isnan(sample):
            return self._primed

        self._sample_count += 1

        # First sample: store as prior reference, no further processing.
        if self._sample_count == 1:
            self._sample_previous = sample
            return False

        # Step 1: High-pass filter.
        hp = self._alpha * self._hp_buffer[_HP_BUFFER_SIZE - 1] + \
            self._half_one_plus_alpha * (sample - self._sample_previous)
        self._sample_previous = sample

        # Shift buffer left.
        for i in range(_HP_BUFFER_SIZE - 1):
            self._hp_buffer[i] = self._hp_buffer[i + 1]
        self._hp_buffer[_HP_BUFFER_SIZE - 1] = hp

        # Step 2: 6-tap FIR smoothing {1,2,3,3,2,1}/12.
        smooth_hp = (self._hp_buffer[0] +
                     2 * self._hp_buffer[1] +
                     3 * self._hp_buffer[2] +
                     3 * self._hp_buffer[3] +
                     2 * self._hp_buffer[4] +
                     self._hp_buffer[5]) / _FIR_COEF_SUM

        # Step 3: Momentum.
        momentum = smooth_hp - self._smooth_hp_previous
        self._smooth_hp_previous = smooth_hp

        # Step 4: Adaptive delta.
        delta = _DELTA_FACTOR * self._sample_count + _DELTA_SUMMAND
        if delta < _DELTA_LOWER_THRESHOLD:
            delta = _DELTA_LOWER_THRESHOLD

        # Step 5: Filter-bank update.
        self.maximal_amplitude_squared = 0.0
        for index in range(self.filter_bank_length):
            n = self.minimal_period_times_two + index
            nf = float(n)

            gamma = 1.0 / math.cos(8.0 * math.pi * delta / nf)
            a = gamma - math.sqrt(gamma * gamma - 1.0)

            quadrature = momentum * (nf / (4.0 * math.pi))
            in_phase = smooth_hp

            half_one_min_a = 0.5 * (1.0 - a)
            beta = self._pre_calculated_beta[index]
            beta_one_plus_a = beta * (1.0 + a)

            f = self.filter_bank[index]

            real = half_one_min_a * (in_phase - f.in_phase_previous) + \
                beta_one_plus_a * f.real - a * f.real_previous
            imag = half_one_min_a * (quadrature - f.quadrature_previous) + \
                beta_one_plus_a * f.imaginary - a * f.imaginary_previous

            amp_sq = real * real + imag * imag

            f.in_phase_previous = f.in_phase
            f.in_phase = in_phase
            f.quadrature_previous = f.quadrature
            f.quadrature = quadrature
            f.real_previous = f.real
            f.real = real
            f.imaginary_previous = f.imaginary
            f.imaginary = imag
            f.amplitude_squared = amp_sq

            if amp_sq > self.maximal_amplitude_squared:
                self.maximal_amplitude_squared = amp_sq

        # Step 6: dB normalization and dominant-cycle weighted average.
        numerator = 0.0
        denominator = 0.0
        self.dominant_cycle = 0.0
        for index in range(self.filter_bank_length):
            f = self.filter_bank[index]

            decibels = 0.0
            if self.maximal_amplitude_squared > 0:
                normalized = f.amplitude_squared / self.maximal_amplitude_squared
                if normalized > 0:
                    arg = (1.0 - _NORMALIZED_AMPLITUDE_FACTOR * normalized) / _DECIBELS_FLOOR
                    if arg > 0:
                        decibels = _DECIBELS_GAIN * math.log10(arg)

            # EMA smoothing.
            decibels = _DB_SMOOTHING_ALPHA * decibels + _DB_SMOOTHING_ONE_MINUS * f.decibels
            if decibels > self.decibels_upper_threshold:
                decibels = self.decibels_upper_threshold
            f.decibels = decibels

            if decibels <= self.decibels_lower_threshold:
                n = float(self.minimal_period_times_two + index)
                adjusted = self.decibels_upper_threshold - decibels
                numerator += n * adjusted
                denominator += adjusted

        if denominator != 0:
            self.dominant_cycle = 0.5 * numerator / denominator
        if self.dominant_cycle < float(self.minimal_period):
            self.dominant_cycle = float(self.minimal_period)

        # Step 7: 5-sample median of dominant cycle.
        for i in range(_DC_BUFFER_SIZE - 1):
            self._dc_buffer[i] = self._dc_buffer[i + 1]
        self._dc_buffer[_DC_BUFFER_SIZE - 1] = self.dominant_cycle

        sorted_buf = sorted(self._dc_buffer)
        self.dominant_cycle_median = sorted_buf[_DC_MEDIAN_INDEX]
        if self.dominant_cycle_median < float(self.minimal_period):
            self.dominant_cycle_median = float(self.minimal_period)

        if self._sample_count < self.minimal_period_times_two:
            return False
        self._primed = True

        return True


def _apply_defaults(p: CoronaParams) -> None:
    """Fill zero/negative fields with Ehlers defaults."""
    if p.high_pass_filter_cutoff <= 0:
        p.high_pass_filter_cutoff = 30
    if p.minimal_period <= 0:
        p.minimal_period = 6
    if p.maximal_period <= 0:
        p.maximal_period = 30
    if p.decibels_lower_threshold == 0:
        p.decibels_lower_threshold = 6.0
    if p.decibels_upper_threshold == 0:
        p.decibels_upper_threshold = 20.0


def _verify_parameters(p: CoronaParams) -> None:
    """Validate parameters, raise ValueError on invalid config."""
    invalid = "invalid corona parameters"
    if p.high_pass_filter_cutoff < 2:
        raise ValueError(f"{invalid}: HighPassFilterCutoff should be >= 2")
    if p.minimal_period < 2:
        raise ValueError(f"{invalid}: MinimalPeriod should be >= 2")
    if p.maximal_period <= p.minimal_period:
        raise ValueError(f"{invalid}: MaximalPeriod should be > MinimalPeriod")
    if p.decibels_lower_threshold < 0:
        raise ValueError(f"{invalid}: DecibelsLowerThreshold should be >= 0")
    if p.decibels_upper_threshold <= p.decibels_lower_threshold:
        raise ValueError(f"{invalid}: DecibelsUpperThreshold should be > DecibelsLowerThreshold")
