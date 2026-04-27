"""Phase Accumulator cycle estimator implementation."""
import math

from .cycle_estimator import CycleEstimator
from .cycle_estimator_params import CycleEstimatorParams
from .estimator import (
    DEFAULT_MIN_PERIOD, DEFAULT_MAX_PERIOD, HT_LENGTH, QUADRATURE_INDEX,
    ACCUMULATION_LENGTH,
    push, ht, correct_amplitude, fill_wma_factors, verify_parameters,
)


class PhaseAccumulatorEstimator(CycleEstimator):
    """Implements the Hilbert transformer with Phase Accumulation period estimation.

    John Ehlers, Rocket Science for Traders, Wiley, 2001, 0471405671, pp 63-66.
    """

    def __init__(self, p: CycleEstimatorParams) -> None:
        verify_parameters(p)

        length = p.smoothing_length
        alpha_quad = p.alpha_ema_quadrature_in_phase
        alpha_period = p.alpha_ema_period

        sl_ht1 = length + HT_LENGTH - 1          # 10
        sl_2ht2 = sl_ht1 + HT_LENGTH - 1         # 16
        sl_2ht1 = sl_2ht2 + 1                     # 17
        sl_2ht = sl_2ht1 + 1                      # 18

        self._smoothing_length = length
        self._min_period = DEFAULT_MIN_PERIOD
        self._max_period = DEFAULT_MAX_PERIOD
        self._alpha_ema_quadrature_in_phase = alpha_quad
        self._alpha_ema_period = alpha_period
        self._warm_up_period = max(p.warm_up_period, sl_2ht)
        self._sl_ht1 = sl_ht1
        self._sl_2ht2 = sl_2ht2
        self._sl_2ht1 = sl_2ht1
        self._sl_2ht = sl_2ht
        self._one_min_alpha_quad = 1.0 - alpha_quad
        self._one_min_alpha_period = 1.0 - alpha_period

        self._raw_values = [0.0] * length
        self._wma_factors = [0.0] * length
        self._wma_smoothed = [0.0] * HT_LENGTH
        self._detrended = [0.0] * HT_LENGTH
        self._delta_phase = [0.0] * ACCUMULATION_LENGTH
        self._in_phase = 0.0
        self._quadrature = 0.0
        self._count = 0
        self._smoothed_in_phase_prev = 0.0
        self._smoothed_quadrature_prev = 0.0
        self._phase_prev = 0.0
        self._period = float(DEFAULT_MIN_PERIOD)
        self._is_primed = False
        self._is_warmed_up = False

        fill_wma_factors(length, self._wma_factors)

    def smoothing_length(self) -> int:
        return self._smoothing_length

    def min_period(self) -> int:
        return self._min_period

    def max_period(self) -> int:
        return self._max_period

    def warm_up_period(self) -> int:
        return self._warm_up_period

    def alpha_ema_quadrature_in_phase(self) -> float:
        return self._alpha_ema_quadrature_in_phase

    def alpha_ema_period(self) -> float:
        return self._alpha_ema_period

    def count(self) -> int:
        return self._count

    def primed(self) -> bool:
        return self._is_warmed_up

    def period(self) -> float:
        return self._period

    def in_phase(self) -> float:
        return self._in_phase

    def quadrature(self) -> float:
        return self._quadrature

    def detrended(self) -> float:
        return self._detrended[0]

    def smoothed(self) -> float:
        return self._wma_smoothed[0]

    def update(self, sample: float) -> None:
        if math.isnan(sample):
            return

        push(self._raw_values, sample)

        if self._is_primed:
            if not self._is_warmed_up:
                self._count += 1
                if self._warm_up_period < self._count:
                    self._is_warmed_up = True

            push(self._wma_smoothed, self._wma())
            acf = correct_amplitude(self._period)
            push(self._detrended, ht(self._wma_smoothed) * acf)

            self._quadrature = ht(self._detrended) * acf
            self._in_phase = self._detrended[QUADRATURE_INDEX]

            si = self._ema_quad(self._in_phase, self._smoothed_in_phase_prev)
            sq = self._ema_quad(self._quadrature, self._smoothed_quadrature_prev)
            self._smoothed_in_phase_prev = si
            self._smoothed_quadrature_prev = sq

            phase = _instantaneous_phase(si, sq, self._phase_prev)
            push(self._delta_phase, _calculate_differential_phase(phase, self._phase_prev))
            self._phase_prev = phase

            period_prev = self._period
            self._period = _instantaneous_period(self._delta_phase, period_prev)
            self._period = self._ema_period(self._period, period_prev)
        else:
            self._count += 1
            if self._smoothing_length > self._count:
                return

            push(self._wma_smoothed, self._wma())

            if self._sl_ht1 > self._count:
                return

            acf = correct_amplitude(self._period)
            push(self._detrended, ht(self._wma_smoothed) * acf)

            if self._sl_2ht2 > self._count:
                return

            self._quadrature = ht(self._detrended) * acf
            self._in_phase = self._detrended[QUADRATURE_INDEX]

            if self._sl_2ht2 == self._count:
                self._smoothed_in_phase_prev = self._in_phase
                self._smoothed_quadrature_prev = self._quadrature
                return

            si = self._ema_quad(self._in_phase, self._smoothed_in_phase_prev)
            sq = self._ema_quad(self._quadrature, self._smoothed_quadrature_prev)
            self._smoothed_in_phase_prev = si
            self._smoothed_quadrature_prev = sq

            phase = _instantaneous_phase(si, sq, self._phase_prev)
            push(self._delta_phase, _calculate_differential_phase(phase, self._phase_prev))
            self._phase_prev = phase

            period_prev = self._period
            self._period = _instantaneous_period(self._delta_phase, period_prev)

            if self._sl_2ht1 < self._count:  # count >= 18
                self._period = self._ema_period(self._period, period_prev)
                self._is_primed = True

    def _wma(self) -> float:
        value = 0.0
        for i in range(self._smoothing_length):
            value += self._wma_factors[i] * self._raw_values[i]
        return value

    def _ema_quad(self, value: float, value_prev: float) -> float:
        return self._alpha_ema_quadrature_in_phase * value + \
            self._one_min_alpha_quad * value_prev

    def _ema_period(self, value: float, value_prev: float) -> float:
        return self._alpha_ema_period * value + self._one_min_alpha_period * value_prev


def _calculate_differential_phase(phase: float, phase_prev: float) -> float:
    """Computes differential phase with wraparound resolution."""
    TWO_PI = 2.0 * math.pi
    PI_OVER_2 = math.pi / 2.0
    THREE_PI_OVER_4 = 3.0 * math.pi / 4.0
    MIN_DELTA = TWO_PI / DEFAULT_MAX_PERIOD
    MAX_DELTA = TWO_PI / DEFAULT_MIN_PERIOD

    delta = phase_prev - phase

    # Resolve phase wraparound from 1st quadrant to 4th quadrant.
    if phase_prev < PI_OVER_2 and phase > THREE_PI_OVER_4:
        delta += TWO_PI

    if delta < MIN_DELTA:
        delta = MIN_DELTA
    elif delta > MAX_DELTA:
        delta = MAX_DELTA

    return delta


def _instantaneous_phase(
    smoothed_in_phase: float, smoothed_quadrature: float, phase_prev: float
) -> float:
    """Computes instantaneous phase using arctangent with quadrant resolution."""
    phase = math.atan(abs(smoothed_quadrature / smoothed_in_phase)) \
        if smoothed_in_phase != 0.0 else phase_prev

    if math.isnan(phase) or math.isinf(phase):
        return phase_prev

    if smoothed_in_phase < 0:
        if smoothed_quadrature > 0:
            phase = math.pi - phase       # 2nd quadrant
        elif smoothed_quadrature < 0:
            phase = math.pi + phase       # 3rd quadrant
    elif smoothed_in_phase > 0 and smoothed_quadrature < 0:
        phase = 2.0 * math.pi - phase    # 4th quadrant

    return phase


def _instantaneous_period(delta_phase: list[float], period_prev: float) -> float:
    """Computes instantaneous period by phase accumulation."""
    TWO_PI = 2.0 * math.pi
    sum_phase = 0.0
    period = 0

    for i in range(ACCUMULATION_LENGTH):
        sum_phase += delta_phase[i]
        if sum_phase >= TWO_PI:
            period = i + 1
            break

    if period == 0:
        return period_prev

    return float(period)
