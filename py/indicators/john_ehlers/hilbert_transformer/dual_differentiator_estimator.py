"""Dual Differentiator cycle estimator implementation."""
import math

from .cycle_estimator import CycleEstimator
from .cycle_estimator_params import CycleEstimatorParams
from .estimator import (
    DEFAULT_MIN_PERIOD, DEFAULT_MAX_PERIOD, HT_LENGTH, QUADRATURE_INDEX,
    push, ht, correct_amplitude, adjust_period, fill_wma_factors, verify_parameters,
)


class DualDifferentiatorEstimator(CycleEstimator):
    """Implements the Hilbert transformer with Dual Differentiator period estimation.

    John Ehlers, Rocket Science for Traders, Wiley, 2001, 0471405671, pp 70-74.
    """

    def __init__(self, p: CycleEstimatorParams) -> None:
        verify_parameters(p)

        length = p.smoothing_length
        alpha_quad = p.alpha_ema_quadrature_in_phase
        alpha_period = p.alpha_ema_period

        sl_ht1 = length + HT_LENGTH - 1           # 10
        sl_2ht2 = sl_ht1 + HT_LENGTH - 1          # 16
        sl_3ht3 = sl_2ht2 + HT_LENGTH - 1         # 22
        sl_3ht2 = sl_3ht3 + 1                      # 23
        sl_3ht1 = sl_3ht2 + 1                      # 24

        self._smoothing_length = length
        self._min_period = DEFAULT_MIN_PERIOD
        self._max_period = DEFAULT_MAX_PERIOD
        self._alpha_ema_quadrature_in_phase = alpha_quad
        self._alpha_ema_period = alpha_period
        self._warm_up_period = max(p.warm_up_period, sl_3ht1)
        self._sl_ht1 = sl_ht1
        self._sl_2ht2 = sl_2ht2
        self._sl_3ht3 = sl_3ht3
        self._sl_3ht2 = sl_3ht2
        self._sl_3ht1 = sl_3ht1
        self._one_min_alpha_quad = 1.0 - alpha_quad
        self._one_min_alpha_period = 1.0 - alpha_period

        self._raw_values = [0.0] * length
        self._wma_factors = [0.0] * length
        self._wma_smoothed = [0.0] * HT_LENGTH
        self._detrended = [0.0] * HT_LENGTH
        self._in_phase_arr = [0.0] * HT_LENGTH
        self._quadrature_arr = [0.0] * HT_LENGTH
        self._j_in_phase = [0.0] * HT_LENGTH
        self._j_quadrature = [0.0] * HT_LENGTH
        self._count = 0
        self._smoothed_in_phase_prev = 0.0
        self._smoothed_quadrature_prev = 0.0
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
        return self._in_phase_arr[0]

    def quadrature(self) -> float:
        return self._quadrature_arr[0]

    def detrended(self) -> float:
        return self._detrended[0]

    def smoothed(self) -> float:
        return self._wma_smoothed[0]

    def update(self, sample: float) -> None:
        if math.isnan(sample):
            return

        TWO_PI = 2.0 * math.pi

        push(self._raw_values, sample)

        if self._is_primed:
            if not self._is_warmed_up:
                self._count += 1
                if self._warm_up_period < self._count:
                    self._is_warmed_up = True

            push(self._wma_smoothed, self._wma())
            acf = correct_amplitude(self._period)

            push(self._detrended, ht(self._wma_smoothed) * acf)
            push(self._quadrature_arr, ht(self._detrended) * acf)
            push(self._in_phase_arr, self._detrended[QUADRATURE_INDEX])
            push(self._j_in_phase, ht(self._in_phase_arr) * acf)
            push(self._j_quadrature, ht(self._quadrature_arr) * acf)

            # Phasor addition for 3 bar averaging + EMA smoothing.
            si = self._ema_quad(
                self._in_phase_arr[0] - self._j_quadrature[0],
                self._smoothed_in_phase_prev,
            )
            sq = self._ema_quad(
                self._quadrature_arr[0] + self._j_in_phase[0],
                self._smoothed_quadrature_prev,
            )

            # Dual Differential discriminator.
            discriminator = sq * (si - self._smoothed_in_phase_prev) - \
                si * (sq - self._smoothed_quadrature_prev)
            self._smoothed_in_phase_prev = si
            self._smoothed_quadrature_prev = sq

            period_prev = self._period
            period_new = TWO_PI * (si * si + sq * sq) / discriminator

            if not math.isnan(period_new) and not math.isinf(period_new):
                self._period = period_new

            self._period = adjust_period(self._period, period_prev)
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

            push(self._quadrature_arr, ht(self._detrended) * acf)
            push(self._in_phase_arr, self._detrended[QUADRATURE_INDEX])

            if self._sl_3ht3 > self._count:
                return

            push(self._j_in_phase, ht(self._in_phase_arr) * acf)
            push(self._j_quadrature, ht(self._quadrature_arr) * acf)

            if self._sl_3ht3 == self._count:  # count == 22
                self._smoothed_in_phase_prev = \
                    self._in_phase_arr[0] - self._j_quadrature[0]
                self._smoothed_quadrature_prev = \
                    self._quadrature_arr[0] + self._j_in_phase[0]
                return

            # count >= 23
            si = self._ema_quad(
                self._in_phase_arr[0] - self._j_quadrature[0],
                self._smoothed_in_phase_prev,
            )
            sq = self._ema_quad(
                self._quadrature_arr[0] + self._j_in_phase[0],
                self._smoothed_quadrature_prev,
            )

            discriminator = sq * (si - self._smoothed_in_phase_prev) - \
                si * (sq - self._smoothed_quadrature_prev)
            self._smoothed_in_phase_prev = si
            self._smoothed_quadrature_prev = sq

            period_prev = self._period
            period_new = TWO_PI * (si * si + sq * sq) / discriminator

            if not math.isnan(period_new) and not math.isinf(period_new):
                self._period = period_new

            self._period = adjust_period(self._period, period_prev)

            if self._sl_3ht2 < self._count:  # count >= 24
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
