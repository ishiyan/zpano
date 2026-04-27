"""Internal Goertzel spectrum estimator."""

import math

_TWO_PI = 2.0 * math.pi


class _Estimator:
    """Goertzel spectrum estimator (unexported, used only by GoertzelSpectrum)."""

    __slots__ = (
        'length', 'spectrum_resolution', 'length_spectrum',
        'min_period', 'max_period',
        'is_first_order', 'is_spectral_dilation_compensation',
        'is_automatic_gain_control', 'automatic_gain_control_decay_factor',
        'input_series', 'input_series_minus_mean', 'spectrum', 'period',
        'frequency_sin', 'frequency_cos', 'frequency_cos2',
        'mean', 'spectrum_min', 'spectrum_max', 'previous_spectrum_max',
    )

    def __init__(
        self,
        length: int,
        min_period: float,
        max_period: float,
        spectrum_resolution: int,
        is_first_order: bool,
        is_spectral_dilation_compensation: bool,
        is_automatic_gain_control: bool,
        automatic_gain_control_decay_factor: float,
    ) -> None:
        self.length = length
        self.spectrum_resolution = spectrum_resolution
        self.length_spectrum = int((max_period - min_period) * spectrum_resolution) + 1
        self.min_period = min_period
        self.max_period = max_period
        self.is_first_order = is_first_order
        self.is_spectral_dilation_compensation = is_spectral_dilation_compensation
        self.is_automatic_gain_control = is_automatic_gain_control
        self.automatic_gain_control_decay_factor = automatic_gain_control_decay_factor

        self.input_series = [0.0] * length
        self.input_series_minus_mean = [0.0] * length
        self.spectrum = [0.0] * self.length_spectrum
        self.period = [0.0] * self.length_spectrum

        self.frequency_sin: list[float] = []
        self.frequency_cos: list[float] = []
        self.frequency_cos2: list[float] = []

        self.mean = 0.0
        self.spectrum_min = 0.0
        self.spectrum_max = 0.0
        self.previous_spectrum_max = 0.0

        result = float(spectrum_resolution)

        if is_first_order:
            self.frequency_sin = [0.0] * self.length_spectrum
            self.frequency_cos = [0.0] * self.length_spectrum
            for i in range(self.length_spectrum):
                period = max_period - i / result
                self.period[i] = period
                theta = _TWO_PI / period
                self.frequency_sin[i] = math.sin(theta)
                self.frequency_cos[i] = math.cos(theta)
        else:
            self.frequency_cos2 = [0.0] * self.length_spectrum
            for i in range(self.length_spectrum):
                period = max_period - i / result
                self.period[i] = period
                self.frequency_cos2[i] = 2.0 * math.cos(_TWO_PI / period)

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

        # Seed with first bin.
        spectrum = self._goertzel_estimate(0)
        if self.is_spectral_dilation_compensation:
            spectrum /= self.period[0]

        self.spectrum[0] = spectrum
        self.spectrum_min = spectrum

        if self.is_automatic_gain_control:
            self.spectrum_max = self.automatic_gain_control_decay_factor * self.previous_spectrum_max
            if self.spectrum_max < spectrum:
                self.spectrum_max = spectrum
        else:
            self.spectrum_max = spectrum

        for i in range(1, self.length_spectrum):
            spectrum = self._goertzel_estimate(i)
            if self.is_spectral_dilation_compensation:
                spectrum /= self.period[i]

            self.spectrum[i] = spectrum

            if self.spectrum_max < spectrum:
                self.spectrum_max = spectrum
            elif self.spectrum_min > spectrum:
                self.spectrum_min = spectrum

        self.previous_spectrum_max = self.spectrum_max

    def _goertzel_estimate(self, j: int) -> float:
        if self.is_first_order:
            return self._goertzel_first_order_estimate(j)
        return self._goertzel_second_order_estimate(j)

    def _goertzel_second_order_estimate(self, j: int) -> float:
        cos2 = self.frequency_cos2[j]
        s1 = 0.0
        s2 = 0.0
        data = self.input_series_minus_mean

        for i in range(self.length):
            s0 = data[i] + cos2 * s1 - s2
            s2 = s1
            s1 = s0

        spectrum = s1 * s1 + s2 * s2 - cos2 * s1 * s2
        if spectrum < 0:
            return 0.0
        return spectrum

    def _goertzel_first_order_estimate(self, j: int) -> float:
        cos_theta = self.frequency_cos[j]
        sin_theta = self.frequency_sin[j]
        yre = 0.0
        yim = 0.0
        data = self.input_series_minus_mean

        for i in range(self.length):
            re = data[i] + cos_theta * yre - sin_theta * yim
            im = data[i] + cos_theta * yim + sin_theta * yre
            yre = re
            yim = im

        return yre * yre + yim * yim
