"""Frequency response analysis for filters/indicators.

Calculates power, amplitude, and phase spectra of a filter's impulse response
using a direct real FFT.
"""

import math
from dataclasses import dataclass, field
from typing import Protocol

from .metadata import Metadata


class Updater(Protocol):
    """A filter whose frequency response is to be calculated."""

    def metadata(self) -> Metadata: ...
    def update(self, sample: float) -> float: ...


@dataclass
class Component:
    """A single calculated filter frequency response component."""

    data: list[float] = field(default_factory=list)
    min: float = float('-inf')
    max: float = float('inf')


def _new_component(length: int) -> Component:
    return Component(data=[0.0] * length, min=float('-inf'), max=float('inf'))


@dataclass
class FrequencyResponse:
    """Calculated filter frequency response data.

    All lists have the same spectrum length.
    """

    label: str
    normalized_frequency: list[float]
    power_percent: Component
    power_decibel: Component
    amplitude_percent: Component
    amplitude_decibel: Component
    phase_degrees: Component
    phase_degrees_unwrapped: Component


def calculate(signal_length: int, filter: Updater, warmup: int,
              phase_degrees_unwrapping_limit: float = 179.0) -> FrequencyResponse:
    """Calculate frequency response of a given impulse signal length.

    Args:
        signal_length: Must be a power of 2 and >= 4 (e.g. 512, 1024, 2048, 4096).
        filter: The filter/indicator implementing Updater protocol.
        warmup: How many times to update filter with zero before calculations.
        phase_degrees_unwrapping_limit: Phase unwrapping threshold (default 179).

    Returns:
        FrequencyResponse with all spectral components.

    Raises:
        ValueError: If signal_length is not a valid power of 2 >= 4.
    """
    if not _is_valid_signal_length(signal_length):
        raise ValueError(
            f"length should be power of 2 and not less than 4: {signal_length}")

    spectrum_length = signal_length // 2 - 1

    fr = FrequencyResponse(
        label=filter.metadata().mnemonic,
        normalized_frequency=[0.0] * spectrum_length,
        power_percent=_new_component(spectrum_length),
        power_decibel=_new_component(spectrum_length),
        amplitude_percent=_new_component(spectrum_length),
        amplitude_decibel=_new_component(spectrum_length),
        phase_degrees=_new_component(spectrum_length),
        phase_degrees_unwrapped=_new_component(spectrum_length),
    )

    _prepare_frequency_domain(spectrum_length, fr.normalized_frequency)

    signal = _prepare_filtered_signal(signal_length, filter, warmup)
    _direct_real_fft(signal)
    _parse_spectrum(spectrum_length, signal, fr.power_percent, fr.amplitude_percent,
                    fr.phase_degrees, fr.phase_degrees_unwrapped,
                    phase_degrees_unwrapping_limit)
    _to_decibels(spectrum_length, fr.power_percent, fr.power_decibel)
    _to_percents(spectrum_length, fr.power_percent, fr.power_percent)
    _to_decibels(spectrum_length, fr.amplitude_percent, fr.amplitude_decibel)
    _to_percents(spectrum_length, fr.amplitude_percent, fr.amplitude_percent)

    return fr


def _is_valid_signal_length(length: int) -> bool:
    while length > 4:
        if length % 2 != 0:
            return False
        length //= 2
    return length == 4


def _prepare_frequency_domain(spectrum_length: int, freq: list[float]) -> None:
    for i in range(spectrum_length):
        freq[i] = (1 + i) / spectrum_length


def _prepare_filtered_signal(signal_length: int, filter: Updater, warmup: int) -> list[float]:
    for _ in range(warmup):
        filter.update(0.0)

    signal = [0.0] * signal_length
    signal[0] = filter.update(1000.0)

    for i in range(1, signal_length):
        signal[i] = filter.update(0.0)

    return signal


def _parse_spectrum(length: int, signal: list[float],
                    power: Component, amplitude: Component,
                    phase: Component, phase_unwrapped: Component,
                    phase_degrees_unwrapping_limit: float) -> None:
    rad2deg = 180.0 / math.pi

    pmin = float('inf')
    pmax = float('-inf')
    amin = float('inf')
    amax = float('-inf')

    k = 2
    for i in range(length):
        re = signal[k]
        k += 1
        im = signal[k]
        k += 1

        phase.data[i] = -math.atan2(im, re) * rad2deg
        phase_unwrapped.data[i] = 0.0

        pwr = re * re + im * im
        power.data[i] = pwr
        pmin = min(pmin, pwr)
        pmax = max(pmax, pwr)

        amp = math.sqrt(pwr)
        amplitude.data[i] = amp
        amin = min(amin, amp)
        amax = max(amax, amp)

    _unwrap_phase_degrees(length, phase.data, phase_unwrapped, phase_degrees_unwrapping_limit)
    phase.min = -180.0
    phase.max = 180.0
    power.min = pmin
    power.max = pmax
    amplitude.min = amin
    amplitude.max = amax


def _unwrap_phase_degrees(length: int, wrapped: list[float],
                          unwrapped: Component, limit: float) -> None:
    k = 0.0

    min_val = wrapped[0]
    max_val = min_val
    unwrapped.data[0] = min_val

    for i in range(1, length):
        w = wrapped[i]
        increment = wrapped[i] - wrapped[i - 1]

        if increment > limit:
            k -= increment
        elif increment < -limit:
            k += increment

        w += k
        min_val = min(min_val, w)
        max_val = max(max_val, w)
        unwrapped.data[i] = w

    unwrapped.min = min_val
    unwrapped.max = max_val


def _to_decibels(length: int, src: Component, tgt: Component) -> None:
    dbmin = float('inf')
    dbmax = float('-inf')

    base = src.data[0]
    if base < 5e-324:  # math.SmallestNonzeroFloat64 equivalent
        base = src.max

    for i in range(length):
        val = src.data[i] / base
        if val <= 0.0:
            db = float('-inf')
        else:
            db = 20.0 * math.log10(val)
        dbmin = min(dbmin, db)
        dbmax = max(dbmax, db)
        tgt.data[i] = db

    # Snap dbmin to interval floor [-100, -90), [-90, -80), ..., [-10, 0)
    for j in range(10, 0, -1):
        lo = -j * 10.0
        hi = -(j - 1) * 10.0
        if dbmin >= lo and dbmin < hi:
            dbmin = lo
            break

    # Limit minimal decibel values to -100
    if dbmin < -100.0:
        dbmin = -100.0
        for i in range(length):
            if tgt.data[i] < -100.0:
                tgt.data[i] = -100.0

    # Snap dbmax to interval ceiling [0, 5), [5, 10)
    for j in range(2, 0, -1):
        hi = j * 5.0
        lo = (j - 1) * 5.0
        if dbmax >= lo and dbmax < hi:
            dbmax = hi
            break

    # Limit maximal decibel values to 10
    if dbmax > 10.0:
        dbmax = 10.0
        for i in range(length):
            if tgt.data[i] > 10.0:
                tgt.data[i] = 10.0

    tgt.min = dbmin
    tgt.max = dbmax


def _to_percents(length: int, src: Component, tgt: Component) -> None:
    pctmin = 0.0
    pctmax = float('-inf')

    base = src.data[0]
    if base < 5e-324:
        base = src.max

    for i in range(length):
        pct = 100.0 * src.data[i] / base
        pctmax = max(pctmax, pct)
        tgt.data[i] = pct

    # Snap pctmax to interval ceiling [100, 110), [110, 120), ..., [190, 200)
    for j in range(10):
        lo = 100.0 + j * 10.0
        hi = 100.0 + (j + 1) * 10.0
        if pctmax >= lo and pctmax < hi:
            pctmax = hi
            break

    # Limit maximal percentage values to 200
    if pctmax > 200.0:
        pctmax = 200.0
        for i in range(length):
            if tgt.data[i] > 200.0:
                tgt.data[i] = 200.0

    tgt.min = pctmin
    tgt.max = pctmax


def _direct_real_fft(array: list[float]) -> None:
    """Direct real fast Fourier transform (in-place).

    Input: real data. Output: {re, im} pairs.
    Length must be a power of 2.
    """
    half = 0.5
    two_pi = 2.0 * math.pi

    length = len(array)
    ttheta = two_pi / length
    nn = length // 2
    j = 1

    for ii in range(1, nn + 1):
        i = 2 * ii - 1

        if j > i:
            array[j - 1], array[i - 1] = array[i - 1], array[j - 1]
            array[j], array[i] = array[i], array[j]

        m = nn
        while m >= 2 and j > m:
            j -= m
            m //= 2

        j += m

    m_max = 2
    n = length

    while n > m_max:
        istep = 2 * m_max
        theta = two_pi / m_max
        wp_r = math.sin(half * theta)
        wp_r = -2.0 * wp_r * wp_r
        wp_i = math.sin(theta)
        w_r = 1.0
        w_i = 0.0

        for ii in range(1, m_max // 2 + 1):
            m = 2 * ii - 1
            for jj in range(0, (n - m) // istep + 1):
                i = m + jj * istep
                j = i + m_max
                temp_r = w_r * array[j - 1] - w_i * array[j]
                temp_i = w_r * array[j] + w_i * array[j - 1]
                array[j - 1] = array[i - 1] - temp_r
                array[j] = array[i] - temp_i
                array[i - 1] = array[i - 1] + temp_r
                array[i] = array[i] + temp_i

            w_temp = w_r
            w_r = w_r * wp_r - w_i * wp_i + w_r
            w_i = w_i * wp_r + w_temp * wp_i + w_i

        m_max = istep

    twp_r = math.sin(half * ttheta)
    twp_r = -2.0 * twp_r * twp_r
    twp_i = math.sin(ttheta)
    tw_r = 1.0 + twp_r
    tw_i = twp_i
    n = length // 4 + 1

    for i in range(2, n + 1):
        i1 = i + i - 2
        i2 = i1 + 1
        i3 = length + 1 - i2
        i4 = i3 + 1
        w_rs = tw_r
        w_is = tw_i
        h1_r = half * (array[i1] + array[i3])
        h1_i = half * (array[i2] - array[i4])
        h2_r = half * (array[i2] + array[i4])
        h2_i = -half * (array[i1] - array[i3])
        array[i1] = h1_r + w_rs * h2_r - w_is * h2_i
        array[i2] = h1_i + w_rs * h2_i + w_is * h2_r
        array[i3] = h1_r - w_rs * h2_r + w_is * h2_i
        array[i4] = -h1_i + w_rs * h2_i + w_is * h2_r
        tw_temp = tw_r
        tw_r = tw_r * twp_r - tw_i * twp_i + tw_r
        tw_i = tw_i * twp_r + tw_temp * twp_i + tw_i

    tw_r = array[0]
    array[0] = tw_r + array[1]
    array[1] = tw_r - array[1]
