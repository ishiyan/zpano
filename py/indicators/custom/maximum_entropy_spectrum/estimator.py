"""Internal Burg maximum-entropy spectrum estimator."""

import math
import sys

_TWO_PI = 2.0 * math.pi


class _Estimator:
    """Burg maximum-entropy spectrum estimator (unexported)."""

    __slots__ = (
        'length', 'degree', 'spectrum_resolution', 'length_spectrum',
        'min_period', 'max_period',
        'is_automatic_gain_control', 'automatic_gain_control_decay_factor',
        'input_series', 'input_series_minus_mean', 'coefficients', 'spectrum', 'period',
        'frequency_sin_omega', 'frequency_cos_omega',
        'h', 'g', 'per', 'pef',
        'mean', 'spectrum_min', 'spectrum_max', 'previous_spectrum_max',
    )

    def __init__(
        self,
        length: int,
        degree: int,
        min_period: float,
        max_period: float,
        spectrum_resolution: int,
        is_automatic_gain_control: bool,
        automatic_gain_control_decay_factor: float,
    ) -> None:
        self.length = length
        self.degree = degree
        self.spectrum_resolution = spectrum_resolution
        self.length_spectrum = int((max_period - min_period) * spectrum_resolution) + 1
        self.min_period = min_period
        self.max_period = max_period
        self.is_automatic_gain_control = is_automatic_gain_control
        self.automatic_gain_control_decay_factor = automatic_gain_control_decay_factor

        self.input_series = [0.0] * length
        self.input_series_minus_mean = [0.0] * length
        self.coefficients = [0.0] * degree
        self.spectrum = [0.0] * self.length_spectrum
        self.period = [0.0] * self.length_spectrum

        self.frequency_sin_omega: list[list[float]] = []
        self.frequency_cos_omega: list[list[float]] = []

        self.h = [0.0] * (degree + 1)
        self.g = [0.0] * (degree + 2)
        self.per = [0.0] * (length + 1)
        self.pef = [0.0] * (length + 1)

        self.mean = 0.0
        self.spectrum_min = 0.0
        self.spectrum_max = 0.0
        self.previous_spectrum_max = 0.0

        result = float(spectrum_resolution)

        for i in range(self.length_spectrum):
            p = max_period - i / result
            self.period[i] = p
            theta = _TWO_PI / p

            sin_row = [0.0] * degree
            cos_row = [0.0] * degree
            for j in range(degree):
                omega = -(j + 1) * theta
                sin_row[j] = math.sin(omega)
                cos_row[j] = math.cos(omega)

            self.frequency_sin_omega.append(sin_row)
            self.frequency_cos_omega.append(cos_row)

    def calculate(self) -> None:
        """Compute spectrum from current input_series."""
        # Subtract mean.
        mean = 0.0
        for i in range(self.length):
            mean += self.input_series[i]
        mean /= self.length

        for i in range(self.length):
            self.input_series_minus_mean[i] = self.input_series[i] - mean
        self.mean = mean

        self._burg_estimate(self.input_series_minus_mean)

        # Evaluate spectrum from AR coefficients.
        self.spectrum_min = sys.float_info.max
        if self.is_automatic_gain_control:
            self.spectrum_max = self.automatic_gain_control_decay_factor * self.previous_spectrum_max
        else:
            self.spectrum_max = -sys.float_info.max

        for i in range(self.length_spectrum):
            real = 1.0
            imag = 0.0

            cos_row = self.frequency_cos_omega[i]
            sin_row = self.frequency_sin_omega[i]

            for j in range(self.degree):
                real -= self.coefficients[j] * cos_row[j]
                imag -= self.coefficients[j] * sin_row[j]

            s = 1.0 / (real * real + imag * imag)
            self.spectrum[i] = s

            if self.spectrum_max < s:
                self.spectrum_max = s
            if self.spectrum_min > s:
                self.spectrum_min = s

        self.previous_spectrum_max = self.spectrum_max

    def _burg_estimate(self, series: list[float]) -> None:
        """Burg AR coefficient estimation (Paul Bourke reference)."""
        for i in range(1, self.length + 1):
            self.pef[i] = 0.0
            self.per[i] = 0.0

        for i in range(1, self.degree + 1):
            sn = 0.0
            sd = 0.0
            jj = self.length - i

            for j in range(jj):
                t1 = series[j + i] + self.pef[j]
                t2 = series[j] + self.per[j]
                sn -= 2.0 * t1 * t2
                sd += t1 * t1 + t2 * t2

            t = sn / sd
            self.g[i] = t

            if i != 1:
                for j in range(1, i):
                    self.h[j] = self.g[j] + t * self.g[i - j]
                for j in range(1, i):
                    self.g[j] = self.h[j]
                jj -= 1

            for j in range(jj):
                self.per[j] += t * self.pef[j] + t * series[j + i]
                self.pef[j] = self.pef[j + 1] + t * self.per[j + 1] + t * series[j + 1]

        for i in range(self.degree):
            self.coefficients[i] = -self.g[i + 1]
