"""Homodyne Discriminator cycle estimator implementation."""
import math

from .cycle_estimator import CycleEstimator
from .cycle_estimator_params import CycleEstimatorParams
from .estimator import (
    DEFAULT_MIN_PERIOD, DEFAULT_MAX_PERIOD, HT_LENGTH, QUADRATURE_INDEX,
    push, ht, correct_amplitude, adjust_period, fill_wma_factors, verify_parameters,
)


class HomodyneDiscriminatorEstimator(CycleEstimator):
    """Implements the Hilbert transformer with the Homodyne Discriminator.

    John Ehlers, Rocket Science for Traders, Wiley, 2001, 0471405671, pp 52-77.
    """

    def __init__(self, p: CycleEstimatorParams) -> None:
        verify_parameters(p)

        length = p.smoothing_length
        alpha_quad = p.alpha_ema_quadrature_in_phase
        alpha_period = p.alpha_ema_period

        sl_ht1 = length + HT_LENGTH - 1
        sl_ht2 = sl_ht1 + HT_LENGTH - 1
        sl_ht3 = sl_ht2 + HT_LENGTH - 1

        self._smoothing_length = length
        self._min_period = DEFAULT_MIN_PERIOD
        self._max_period = DEFAULT_MAX_PERIOD
        self._alpha_ema_quadrature_in_phase = alpha_quad
        self._alpha_ema_period = alpha_period
        self._warm_up_period = max(p.warm_up_period, sl_ht3 + 3)
        self._sl_ht1 = sl_ht1
        self._sl_ht2 = sl_ht2
        self._sl_ht3 = sl_ht3
        self._sl_ht3_p1 = sl_ht3 + 1
        self._sl_ht3_p2 = sl_ht3 + 2
        self._sl_ht3_p3 = sl_ht3 + 3
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

        fill_wma_factors(length, self._wma_factors)

        self._count = 0
        self._smoothed_in_phase_prev = 0.0
        self._smoothed_quadrature_prev = 0.0
        self._re_prev = 0.0
        self._im_prev = 0.0
        self._period_val = float(DEFAULT_MIN_PERIOD)
        self._is_primed = False
        self._is_warmed_up = False

    def smoothing_length(self) -> int:
        return self._smoothing_length

    def smoothed(self) -> float:
        return self._wma_smoothed[0]

    def detrended(self) -> float:
        return self._detrended[0]

    def quadrature(self) -> float:
        return self._quadrature_arr[0]

    def in_phase(self) -> float:
        return self._in_phase_arr[0]

    def period(self) -> float:
        return self._period_val

    def count(self) -> int:
        return self._count

    def primed(self) -> bool:
        return self._is_warmed_up

    def min_period(self) -> int:
        return self._min_period

    def max_period(self) -> int:
        return self._max_period

    def alpha_ema_quadrature_in_phase(self) -> float:
        return self._alpha_ema_quadrature_in_phase

    def alpha_ema_period(self) -> float:
        return self._alpha_ema_period

    def warm_up_period(self) -> int:
        return self._warm_up_period

    def update(self, sample: float) -> None:
        if math.isnan(sample):
            return

        two_pi = 2.0 * math.pi

        push(self._raw_values, sample)

        if self._is_primed:
            if not self._is_warmed_up:
                self._count += 1
                if self._warm_up_period < self._count:
                    self._is_warmed_up = True

            push(self._wma_smoothed, self._wma(self._raw_values))
            acf = correct_amplitude(self._period_val)

            push(self._detrended, ht(self._wma_smoothed) * acf)
            push(self._quadrature_arr, ht(self._detrended) * acf)
            push(self._in_phase_arr, self._detrended[QUADRATURE_INDEX])
            push(self._j_in_phase, ht(self._in_phase_arr) * acf)
            push(self._j_quadrature, ht(self._quadrature_arr) * acf)

            smoothed_ip = self._ema_quad(
                self._in_phase_arr[0] - self._j_quadrature[0],
                self._smoothed_in_phase_prev)
            smoothed_q = self._ema_quad(
                self._quadrature_arr[0] + self._j_in_phase[0],
                self._smoothed_quadrature_prev)

            re = smoothed_ip * self._smoothed_in_phase_prev + \
                smoothed_q * self._smoothed_quadrature_prev
            im = smoothed_ip * self._smoothed_quadrature_prev - \
                smoothed_q * self._smoothed_in_phase_prev
            self._smoothed_in_phase_prev = smoothed_ip
            self._smoothed_quadrature_prev = smoothed_q

            re = self._ema_quad(re, self._re_prev)
            im = self._ema_quad(im, self._im_prev)
            self._re_prev = re
            self._im_prev = im
            period_prev = self._period_val
            atan_val = math.atan2(im, re)
            if atan_val == 0.0:
                period_new = math.inf
            else:
                period_new = two_pi / atan_val

            if not math.isnan(period_new) and not math.isinf(period_new):
                self._period_val = period_new

            self._period_val = adjust_period(self._period_val, period_prev)
            self._period_val = self._ema_period(self._period_val, period_prev)
        else:
            self._count += 1
            if self._smoothing_length > self._count:
                return

            push(self._wma_smoothed, self._wma(self._raw_values))

            if self._sl_ht1 > self._count:
                return

            acf = correct_amplitude(self._period_val)
            push(self._detrended, ht(self._wma_smoothed) * acf)

            if self._sl_ht2 > self._count:
                return

            push(self._quadrature_arr, ht(self._detrended) * acf)
            push(self._in_phase_arr, self._detrended[QUADRATURE_INDEX])

            if self._sl_ht3 > self._count:
                return

            push(self._j_in_phase, ht(self._in_phase_arr) * acf)
            push(self._j_quadrature, ht(self._quadrature_arr) * acf)

            if self._sl_ht3 == self._count:
                self._smoothed_in_phase_prev = \
                    self._in_phase_arr[0] - self._j_quadrature[0]
                self._smoothed_quadrature_prev = \
                    self._quadrature_arr[0] + self._j_in_phase[0]
                return

            smoothed_ip = self._ema_quad(
                self._in_phase_arr[0] - self._j_quadrature[0],
                self._smoothed_in_phase_prev)
            smoothed_q = self._ema_quad(
                self._quadrature_arr[0] + self._j_in_phase[0],
                self._smoothed_quadrature_prev)

            re = smoothed_ip * self._smoothed_in_phase_prev + \
                smoothed_q * self._smoothed_quadrature_prev
            im = smoothed_ip * self._smoothed_quadrature_prev - \
                smoothed_q * self._smoothed_in_phase_prev
            self._smoothed_in_phase_prev = smoothed_ip
            self._smoothed_quadrature_prev = smoothed_q

            if self._sl_ht3_p1 == self._count:
                self._re_prev = re
                self._im_prev = im
                return

            re = self._ema_quad(re, self._re_prev)
            im = self._ema_quad(im, self._im_prev)
            self._re_prev = re
            self._im_prev = im
            period_prev = self._period_val

            atan_val = math.atan2(im, re)
            if atan_val != 0.0:
                period_new = two_pi / atan_val
            else:
                period_new = math.inf
            if not math.isnan(period_new) and not math.isinf(period_new):
                self._period_val = period_new

            self._period_val = adjust_period(self._period_val, period_prev)

            if self._sl_ht3_p2 < self._count:
                self._period_val = self._ema_period(self._period_val, period_prev)
                self._is_primed = True

    def _wma(self, array: list[float]) -> float:
        value = 0.0
        for i in range(self._smoothing_length):
            value += self._wma_factors[i] * array[i]
        return value

    def _ema_quad(self, value: float, value_prev: float) -> float:
        return self._alpha_ema_quadrature_in_phase * value + \
            self._one_min_alpha_quad * value_prev

    def _ema_period(self, value: float, value_prev: float) -> float:
        return self._alpha_ema_period * value + \
            self._one_min_alpha_period * value_prev
