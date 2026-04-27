"""Homodyne Discriminator Estimator (TA-Lib unrolled loops) implementation."""
import math

from .cycle_estimator import CycleEstimator
from .cycle_estimator_params import CycleEstimatorParams
from .estimator import (
    DEFAULT_MIN_PERIOD, DEFAULT_MAX_PERIOD,
    verify_parameters,
)


class HomodyneDiscriminatorEstimatorUnrolled(CycleEstimator):
    """Implements the Hilbert transformer with Homodyne Discriminator (TA-Lib unrolled).

    John Ehlers, Rocket Science for Traders, Wiley, 2001, 0471405671, pp 52-77.
    """

    def __init__(self, p: CycleEstimatorParams) -> None:
        verify_parameters(p)

        length = p.smoothing_length
        alpha_quad = p.alpha_ema_quadrature_in_phase
        alpha_period = p.alpha_ema_period

        if length == 4:
            sm = 1.0 / 10.0
        elif length == 3:
            sm = 1.0 / 6.0
        else:
            sm = 1.0 / 3.0

        primed_count = 23

        self._smoothing_length = length
        self._min_period = DEFAULT_MIN_PERIOD
        self._max_period = DEFAULT_MAX_PERIOD
        self._alpha_quad = alpha_quad
        self._alpha_period = alpha_period
        self._warm_up_period = max(p.warm_up_period, primed_count)
        self._one_min_alpha_quad = 1.0 - alpha_quad
        self._one_min_alpha_period = 1.0 - alpha_period
        self._sm = sm
        self._adjusted_period = 0.0
        self._count = 0
        self._idx = 0
        self._i2_prev = 0.0
        self._q2_prev = 0.0
        self._re_val = 0.0
        self._im_val = 0.0
        self._period_val = float(DEFAULT_MIN_PERIOD)
        self._is_primed = False
        self._is_warmed_up = False
        self._smoothed_val = 0.0
        self._detrended_val = 0.0
        self._in_phase_val = 0.0
        self._quadrature_val = 0.0

        # WMA state
        self._wma_sum = 0.0
        self._wma_sub = 0.0
        self._wma_input = [0.0, 0.0, 0.0, 0.0]  # input1..4

        # Detrender odd/even: 3 slots each + previous + previousInput
        self._det_odd = [0.0, 0.0, 0.0]
        self._det_prev_odd = 0.0
        self._det_prev_input_odd = 0.0
        self._det_even = [0.0, 0.0, 0.0]
        self._det_prev_even = 0.0
        self._det_prev_input_even = 0.0

        # Q1 odd/even
        self._q1_odd = [0.0, 0.0, 0.0]
        self._q1_prev_odd = 0.0
        self._q1_prev_input_odd = 0.0
        self._q1_even = [0.0, 0.0, 0.0]
        self._q1_prev_even = 0.0
        self._q1_prev_input_even = 0.0

        # I1 previous
        self._i1_prev1_odd = 0.0
        self._i1_prev2_odd = 0.0
        self._i1_prev1_even = 0.0
        self._i1_prev2_even = 0.0

        # jI odd/even
        self._ji_odd = [0.0, 0.0, 0.0]
        self._ji_prev_odd = 0.0
        self._ji_prev_input_odd = 0.0
        self._ji_even = [0.0, 0.0, 0.0]
        self._ji_prev_even = 0.0
        self._ji_prev_input_even = 0.0

        # jQ odd/even
        self._jq_odd = [0.0, 0.0, 0.0]
        self._jq_prev_odd = 0.0
        self._jq_prev_input_odd = 0.0
        self._jq_even = [0.0, 0.0, 0.0]
        self._jq_prev_even = 0.0
        self._jq_prev_input_even = 0.0

    def smoothing_length(self) -> int:
        return self._smoothing_length

    def smoothed(self) -> float:
        return self._smoothed_val

    def detrended(self) -> float:
        return self._detrended_val

    def quadrature(self) -> float:
        return self._quadrature_val

    def in_phase(self) -> float:
        return self._in_phase_val

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
        return self._alpha_quad

    def alpha_ema_period(self) -> float:
        return self._alpha_period

    def warm_up_period(self) -> int:
        return self._warm_up_period

    def update(self, sample: float) -> None:  # noqa: C901
        if math.isnan(sample):
            return

        A = 0.0962
        B = 0.5769

        self._count += 1
        sl = self._smoothing_length
        value = 0.0
        do_detrend = False

        # WMA smoothing
        if sl >= self._count:
            if 1 == self._count:
                self._wma_sub = sample
                self._wma_input[0] = sample
                self._wma_sum = sample
            elif 2 == self._count:
                self._wma_sub += sample
                self._wma_input[1] = sample
                self._wma_sum += sample * 2
                if 2 == sl:
                    value = self._wma_sum * self._sm
                    do_detrend = True
            elif 3 == self._count:
                self._wma_sub += sample
                self._wma_input[2] = sample
                self._wma_sum += sample * 3
                if 3 == sl:
                    value = self._wma_sum * self._sm
                    do_detrend = True
            else:  # 4
                self._wma_sub += sample
                self._wma_input[3] = sample
                self._wma_sum += sample * 4
                value = self._wma_sum * self._sm
                do_detrend = True

            if not do_detrend:
                return
        else:
            self._wma_sum -= self._wma_sub
            self._wma_sum += sample * sl
            value = self._wma_sum * self._sm
            self._wma_sub += sample
            self._wma_sub -= self._wma_input[0]
            self._wma_input[0] = self._wma_input[1]

            if 4 == sl:
                self._wma_input[1] = self._wma_input[2]
                self._wma_input[2] = self._wma_input[3]
                self._wma_input[3] = sample
            elif 3 == sl:
                self._wma_input[1] = self._wma_input[2]
                self._wma_input[2] = sample
            else:
                self._wma_input[1] = sample

        # Detrender
        self._smoothed_val = value

        if not self._is_warmed_up:
            self._is_warmed_up = self._count > self._warm_up_period
            if not self._is_primed:
                self._is_primed = self._count > 23

        temp = A * self._smoothed_val
        self._adjusted_period = 0.075 * self._period_val + 0.54

        ji = 0.0
        jq = 0.0
        detrender = 0.0

        if 0 == self._count % 2:
            # Even
            idx = self._idx
            if 0 == idx:
                self._idx = 1
                detrender = -self._det_even[0]
                self._det_even[0] = temp
                detrender += temp
                detrender -= self._det_prev_even
                self._det_prev_even = B * self._det_prev_input_even
                self._det_prev_input_even = value
                detrender += self._det_prev_even
                detrender *= self._adjusted_period

                temp2 = A * detrender
                self._quadrature_val = -self._q1_even[0]
                self._q1_even[0] = temp2
                self._quadrature_val += temp2
                self._quadrature_val -= self._q1_prev_even
                self._q1_prev_even = B * self._q1_prev_input_even
                self._q1_prev_input_even = detrender
                self._quadrature_val += self._q1_prev_even
                self._quadrature_val *= self._adjusted_period

                temp3 = A * self._i1_prev2_even
                ji = -self._ji_even[0]
                self._ji_even[0] = temp3
                ji += temp3
                ji -= self._ji_prev_even
                self._ji_prev_even = B * self._ji_prev_input_even
                self._ji_prev_input_even = self._i1_prev2_even
                ji += self._ji_prev_even
                ji *= self._adjusted_period

                temp4 = A * self._quadrature_val
                jq = -self._jq_even[0]
                self._jq_even[0] = temp4
            elif 1 == idx:
                self._idx = 2
                detrender = -self._det_even[1]
                self._det_even[1] = temp
                detrender += temp
                detrender -= self._det_prev_even
                self._det_prev_even = B * self._det_prev_input_even
                self._det_prev_input_even = value
                detrender += self._det_prev_even
                detrender *= self._adjusted_period

                temp2 = A * detrender
                self._quadrature_val = -self._q1_even[1]
                self._q1_even[1] = temp2
                self._quadrature_val += temp2
                self._quadrature_val -= self._q1_prev_even
                self._q1_prev_even = B * self._q1_prev_input_even
                self._q1_prev_input_even = detrender
                self._quadrature_val += self._q1_prev_even
                self._quadrature_val *= self._adjusted_period

                temp3 = A * self._i1_prev2_even
                ji = -self._ji_even[1]
                self._ji_even[1] = temp3
                ji += temp3
                ji -= self._ji_prev_even
                self._ji_prev_even = B * self._ji_prev_input_even
                self._ji_prev_input_even = self._i1_prev2_even
                ji += self._ji_prev_even
                ji *= self._adjusted_period

                temp4 = A * self._quadrature_val
                jq = -self._jq_even[1]
                self._jq_even[1] = temp4
            else:  # 2
                self._idx = 0
                detrender = -self._det_even[2]
                self._det_even[2] = temp
                detrender += temp
                detrender -= self._det_prev_even
                self._det_prev_even = B * self._det_prev_input_even
                self._det_prev_input_even = value
                detrender += self._det_prev_even
                detrender *= self._adjusted_period

                temp2 = A * detrender
                self._quadrature_val = -self._q1_even[2]
                self._q1_even[2] = temp2
                self._quadrature_val += temp2
                self._quadrature_val -= self._q1_prev_even
                self._q1_prev_even = B * self._q1_prev_input_even
                self._q1_prev_input_even = detrender
                self._quadrature_val += self._q1_prev_even
                self._quadrature_val *= self._adjusted_period

                temp3 = A * self._i1_prev2_even
                ji = -self._ji_even[2]
                self._ji_even[2] = temp3
                ji += temp3
                ji -= self._ji_prev_even
                self._ji_prev_even = B * self._ji_prev_input_even
                self._ji_prev_input_even = self._i1_prev2_even
                ji += self._ji_prev_even
                ji *= self._adjusted_period

                temp4 = A * self._quadrature_val
                jq = -self._jq_even[2]
                self._jq_even[2] = temp4

            # jQ continued (even)
            jq += temp4
            jq -= self._jq_prev_even
            self._jq_prev_even = B * self._jq_prev_input_even
            self._jq_prev_input_even = self._quadrature_val
            jq += self._jq_prev_even
            jq *= self._adjusted_period

            self._in_phase_val = self._i1_prev2_even
            self._i1_prev2_odd = self._i1_prev1_odd
            self._i1_prev1_odd = detrender
        else:
            # Odd
            idx = self._idx
            if 0 == idx:
                self._idx = 1
                detrender = -self._det_odd[0]
                self._det_odd[0] = temp
                detrender += temp
                detrender -= self._det_prev_odd
                self._det_prev_odd = B * self._det_prev_input_odd
                self._det_prev_input_odd = value
                detrender += self._det_prev_odd
                detrender *= self._adjusted_period

                temp2 = A * detrender
                self._quadrature_val = -self._q1_odd[0]
                self._q1_odd[0] = temp2
                self._quadrature_val += temp2
                self._quadrature_val -= self._q1_prev_odd
                self._q1_prev_odd = B * self._q1_prev_input_odd
                self._q1_prev_input_odd = detrender
                self._quadrature_val += self._q1_prev_odd
                self._quadrature_val *= self._adjusted_period

                temp3 = A * self._i1_prev2_odd
                ji = -self._ji_odd[0]
                self._ji_odd[0] = temp3
                ji += temp3
                ji -= self._ji_prev_odd
                self._ji_prev_odd = B * self._ji_prev_input_odd
                self._ji_prev_input_odd = self._i1_prev2_odd
                ji += self._ji_prev_odd
                ji *= self._adjusted_period

                temp4 = A * self._quadrature_val
                jq = -self._jq_odd[0]
                self._jq_odd[0] = temp4
            elif 1 == idx:
                self._idx = 2
                detrender = -self._det_odd[1]
                self._det_odd[1] = temp
                detrender += temp
                detrender -= self._det_prev_odd
                self._det_prev_odd = B * self._det_prev_input_odd
                self._det_prev_input_odd = value
                detrender += self._det_prev_odd
                detrender *= self._adjusted_period

                temp2 = A * detrender
                self._quadrature_val = -self._q1_odd[1]
                self._q1_odd[1] = temp2
                self._quadrature_val += temp2
                self._quadrature_val -= self._q1_prev_odd
                self._q1_prev_odd = B * self._q1_prev_input_odd
                self._q1_prev_input_odd = detrender
                self._quadrature_val += self._q1_prev_odd
                self._quadrature_val *= self._adjusted_period

                temp3 = A * self._i1_prev2_odd
                ji = -self._ji_odd[1]
                self._ji_odd[1] = temp3
                ji += temp3
                ji -= self._ji_prev_odd
                self._ji_prev_odd = B * self._ji_prev_input_odd
                self._ji_prev_input_odd = self._i1_prev2_odd
                ji += self._ji_prev_odd
                ji *= self._adjusted_period

                temp4 = A * self._quadrature_val
                jq = -self._jq_odd[1]
                self._jq_odd[1] = temp4
            else:  # 2
                self._idx = 0
                detrender = -self._det_odd[2]
                self._det_odd[2] = temp
                detrender += temp
                detrender -= self._det_prev_odd
                self._det_prev_odd = B * self._det_prev_input_odd
                self._det_prev_input_odd = value
                detrender += self._det_prev_odd
                detrender *= self._adjusted_period

                temp2 = A * detrender
                self._quadrature_val = -self._q1_odd[2]
                self._q1_odd[2] = temp2
                self._quadrature_val += temp2
                self._quadrature_val -= self._q1_prev_odd
                self._q1_prev_odd = B * self._q1_prev_input_odd
                self._q1_prev_input_odd = detrender
                self._quadrature_val += self._q1_prev_odd
                self._quadrature_val *= self._adjusted_period

                temp3 = A * self._i1_prev2_odd
                ji = -self._ji_odd[2]
                self._ji_odd[2] = temp3
                ji += temp3
                ji -= self._ji_prev_odd
                self._ji_prev_odd = B * self._ji_prev_input_odd
                self._ji_prev_input_odd = self._i1_prev2_odd
                ji += self._ji_prev_odd
                ji *= self._adjusted_period

                temp4 = A * self._quadrature_val
                jq = -self._jq_odd[2]
                self._jq_odd[2] = temp4

            # jQ continued (odd)
            jq += temp4
            jq -= self._jq_prev_odd
            self._jq_prev_odd = B * self._jq_prev_input_odd
            self._jq_prev_input_odd = self._quadrature_val
            jq += self._jq_prev_odd
            jq *= self._adjusted_period

            self._in_phase_val = self._i1_prev2_odd
            self._i1_prev2_even = self._i1_prev1_even
            self._i1_prev1_even = detrender

        self._detrended_val = detrender

        # Phasor addition for 3 bar averaging
        i2 = self._in_phase_val - jq
        q2 = self._quadrature_val + ji

        # Smooth I and Q
        i2 = self._alpha_quad * i2 + self._one_min_alpha_quad * self._i2_prev
        q2 = self._alpha_quad * q2 + self._one_min_alpha_quad * self._q2_prev

        # Homodyne discriminator
        self._re_val = self._alpha_quad * (i2 * self._i2_prev + q2 * self._q2_prev) + \
            self._one_min_alpha_quad * self._re_val
        self._im_val = self._alpha_quad * (i2 * self._q2_prev - q2 * self._i2_prev) + \
            self._one_min_alpha_quad * self._im_val
        self._q2_prev = q2
        self._i2_prev = i2
        temp_period = self._period_val

        atan_val = math.atan2(self._im_val, self._re_val)
        if atan_val != 0.0:
            period_new = 2.0 * math.pi / atan_val
            if not math.isnan(period_new) and not math.isinf(period_new):
                self._period_val = period_new

        v = 1.5 * temp_period
        if self._period_val > v:
            self._period_val = v
        else:
            v = 0.67 * temp_period
            if self._period_val < v:
                self._period_val = v

        if self._period_val < DEFAULT_MIN_PERIOD:
            self._period_val = DEFAULT_MIN_PERIOD
        elif self._period_val > DEFAULT_MAX_PERIOD:
            self._period_val = DEFAULT_MAX_PERIOD

        self._period_val = self._alpha_period * self._period_val + \
            self._one_min_alpha_period * temp_period
