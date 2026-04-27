"""MBST's Discrete Fourier Transform Spectrum heatmap indicator.

The Discrete Fourier Transform Spectrum (psDft) displays a power heatmap of
the cyclic activity over a configurable cycle-period range by evaluating a
discrete Fourier transform on a length-N sliding window with its mean
subtracted. It supports optional spectral dilation compensation (division of
the squared magnitude by the evaluated period), a fast-attack slow-decay
automatic gain control, and either floating or fixed (0-clamped) intensity
normalization.

Reference: MBST Mbs.Trading.Indicators.JohnEhlers.DiscreteFourierTransformSpectrum.
"""

import math
import sys
import datetime

from ....entities.bar import Bar
from ....entities.bar_component import BarComponent, bar_component_value, DEFAULT_BAR_COMPONENT
from ....entities.quote import Quote
from ....entities.quote_component import QuoteComponent, quote_component_value, DEFAULT_QUOTE_COMPONENT
from ....entities.trade import Trade
from ....entities.trade_component import TradeComponent, trade_component_value, DEFAULT_TRADE_COMPONENT
from ....entities.scalar import Scalar
from ...core.identifier import Identifier
from ...core.metadata import Metadata
from ...core.build_metadata import build_metadata, OutputText
from ...core.component_triple_mnemonic import component_triple_mnemonic
from ...core.outputs.heatmap import Heatmap
from .params import Params


TWO_PI = 2.0 * math.pi


class _Estimator:
    """Internal DFT spectrum estimator: sliding window, mean subtraction, DFT, AGC."""

    __slots__ = (
        'length', 'spectrum_resolution', 'length_spectrum', 'max_omega_length',
        'min_period', 'max_period',
        'is_spectral_dilation_compensation', 'is_automatic_gain_control',
        'automatic_gain_control_decay_factor',
        'input_series', 'input_series_minus_mean', 'spectrum', 'period_arr',
        'frequency_sin_omega', 'frequency_cos_omega',
        'mean', 'spectrum_min', 'spectrum_max', 'previous_spectrum_max',
    )

    def __init__(self, length: int, min_period: float, max_period: float,
                 spectrum_resolution: int,
                 is_spectral_dilation_compensation: bool,
                 is_automatic_gain_control: bool,
                 automatic_gain_control_decay_factor: float) -> None:
        self.length = length
        self.spectrum_resolution = spectrum_resolution
        self.length_spectrum = int((max_period - min_period) * spectrum_resolution) + 1
        self.max_omega_length = length
        self.min_period = min_period
        self.max_period = max_period
        self.is_spectral_dilation_compensation = is_spectral_dilation_compensation
        self.is_automatic_gain_control = is_automatic_gain_control
        self.automatic_gain_control_decay_factor = automatic_gain_control_decay_factor

        ls = self.length_spectrum
        mol = self.max_omega_length
        res = float(spectrum_resolution)

        self.input_series = [0.0] * length
        self.input_series_minus_mean = [0.0] * length
        self.spectrum = [0.0] * ls
        self.period_arr = [0.0] * ls

        self.frequency_sin_omega = [[0.0] * mol for _ in range(ls)]
        self.frequency_cos_omega = [[0.0] * mol for _ in range(ls)]

        # Build trig tables; spectrum evaluated from MaxPeriod down to MinPeriod.
        for i in range(ls):
            p = max_period - i / res
            self.period_arr[i] = p
            theta = TWO_PI / p

            sin_row = self.frequency_sin_omega[i]
            cos_row = self.frequency_cos_omega[i]
            for j in range(mol):
                omega = j * theta
                sin_row[j] = math.sin(omega)
                cos_row[j] = math.cos(omega)

        self.mean = 0.0
        self.spectrum_min = 0.0
        self.spectrum_max = 0.0
        self.previous_spectrum_max = 0.0

    def calculate(self) -> None:
        length = self.length
        ls = self.length_spectrum
        mol = self.max_omega_length
        inp = self.input_series
        inp_mm = self.input_series_minus_mean

        # Subtract mean.
        mean = 0.0
        for i in range(length):
            mean += inp[i]
        mean /= length

        for i in range(length):
            inp_mm[i] = inp[i] - mean

        self.mean = mean

        # DFT power spectrum.
        if self.is_automatic_gain_control:
            self.spectrum_max = self.automatic_gain_control_decay_factor * self.previous_spectrum_max
        else:
            self.spectrum_max = -sys.float_info.max

        self.spectrum_min = sys.float_info.max

        spectrum = self.spectrum
        period_arr = self.period_arr
        sin_tab = self.frequency_sin_omega
        cos_tab = self.frequency_cos_omega

        for i in range(ls):
            sin_row = sin_tab[i]
            cos_row = cos_tab[i]

            sum_sin = sum_cos = 0.0
            for j in range(mol):
                sample = inp_mm[j]
                sum_sin += sample * sin_row[j]
                sum_cos += sample * cos_row[j]

            s = sum_sin * sum_sin + sum_cos * sum_cos
            if self.is_spectral_dilation_compensation:
                s /= period_arr[i]

            spectrum[i] = s

            if self.spectrum_max < s:
                self.spectrum_max = s
            if self.spectrum_min > s:
                self.spectrum_min = s

        self.previous_spectrum_max = self.spectrum_max


class DiscreteFourierTransformSpectrum:
    """MBST's Discrete Fourier Transform Spectrum heatmap indicator."""

    def __init__(self, params: Params) -> None:
        invalid = "invalid discrete Fourier transform spectrum parameters"
        def_length = 48
        def_min_period = 10.0
        def_max_period = 48.0
        def_spectrum_res = 1
        def_agc_decay = 0.995
        agc_decay_epsilon = 1e-12

        length = params.length if params.length != 0 else def_length
        min_period = params.min_period if params.min_period != 0 else def_min_period
        max_period = params.max_period if params.max_period != 0 else def_max_period
        spectrum_resolution = params.spectrum_resolution if params.spectrum_resolution != 0 else def_spectrum_res
        agc_decay = params.automatic_gain_control_decay_factor \
            if params.automatic_gain_control_decay_factor != 0 else def_agc_decay

        sdc_on = not params.disable_spectral_dilation_compensation
        agc_on = not params.disable_automatic_gain_control
        floating_norm = not params.fixed_normalization

        if length < 2:
            raise ValueError(f"{invalid}: Length should be >= 2")
        if min_period < 2:
            raise ValueError(f"{invalid}: MinPeriod should be >= 2")
        if max_period <= min_period:
            raise ValueError(f"{invalid}: MaxPeriod should be > MinPeriod")
        if max_period > 2 * length:
            raise ValueError(f"{invalid}: MaxPeriod should be <= 2 * Length")
        if spectrum_resolution < 1:
            raise ValueError(f"{invalid}: SpectrumResolution should be >= 1")
        if agc_on and (agc_decay <= 0 or agc_decay >= 1):
            raise ValueError(
                f"{invalid}: AutomaticGainControlDecayFactor should be in (0, 1)")

        bc = params.bar_component if params.bar_component is not None else BarComponent.MEDIAN
        qc = params.quote_component if params.quote_component is not None else DEFAULT_QUOTE_COMPONENT
        tc = params.trade_component if params.trade_component is not None else DEFAULT_TRADE_COMPONENT

        self._bar_func = bar_component_value(bc)
        self._quote_func = quote_component_value(qc)
        self._trade_func = trade_component_value(tc)

        comp_mn = component_triple_mnemonic(bc, qc, tc)

        # Build flag tags.
        flags = ""
        if not sdc_on:
            flags += ", no-sdc"
        if not agc_on:
            flags += ", no-agc"
        if agc_on and abs(agc_decay - def_agc_decay) > agc_decay_epsilon:
            flags += f", agc={agc_decay:g}"
        if not floating_norm:
            flags += ", no-fn"

        self.mnemonic = f"dftps({length}, {min_period:g}, {max_period:g}, {spectrum_resolution}{flags}{comp_mn})"
        self.description = "Discrete Fourier transform spectrum " + self.mnemonic

        self._estimator = _Estimator(
            length, min_period, max_period, spectrum_resolution,
            sdc_on, agc_on, agc_decay,
        )
        self._last_index = length - 1
        self._window_count = 0
        self._primed = False
        self._floating_normalization = floating_norm
        self._min_param = min_period
        self._max_param = max_period
        self._param_res = float(spectrum_resolution)

    def is_primed(self) -> bool:
        return self._primed

    def metadata(self) -> Metadata:
        return build_metadata(
            Identifier.DISCRETE_FOURIER_TRANSFORM_SPECTRUM,
            self.mnemonic,
            self.description,
            [OutputText(mnemonic=self.mnemonic, description=self.description)],
        )

    def update(self, sample: float, t: datetime.datetime) -> Heatmap:
        if math.isnan(sample):
            return Heatmap.empty(t, self._min_param, self._max_param, self._param_res)

        window = self._estimator.input_series

        if self._primed:
            # Shift left, new sample at end.
            window[:self._last_index] = window[1:]
            window[self._last_index] = sample
        else:
            window[self._window_count] = sample
            self._window_count += 1
            if self._window_count == self._estimator.length:
                self._primed = True

        if not self._primed:
            return Heatmap.empty(t, self._min_param, self._max_param, self._param_res)

        self._estimator.calculate()

        ls = self._estimator.length_spectrum

        min_ref = 0.0
        if self._floating_normalization:
            min_ref = self._estimator.spectrum_min

        max_ref = self._estimator.spectrum_max
        spectrum_range = max_ref - min_ref

        # MBST fills spectrum[0] at MaxPeriod and spectrum[last] at MinPeriod.
        # Heatmap axis runs MinPeriod -> MaxPeriod, so reverse on output.
        values = [0.0] * ls
        value_min = math.inf
        value_max = -math.inf

        for i in range(ls):
            v = (self._estimator.spectrum[ls - 1 - i] - min_ref) / spectrum_range
            values[i] = v
            if v < value_min:
                value_min = v
            if v > value_max:
                value_max = v

        return Heatmap(t, self._min_param, self._max_param, self._param_res,
                       value_min, value_max, values)

    def update_scalar(self, sample: Scalar) -> list:
        return self._update_entity(sample.time, sample.value)

    def update_bar(self, sample: Bar) -> list:
        return self._update_entity(sample.time, self._bar_func(sample))

    def update_quote(self, sample: Quote) -> list:
        return self._update_entity(sample.time, self._quote_func(sample))

    def update_trade(self, sample: Trade) -> list:
        return self._update_entity(sample.time, self._trade_func(sample))

    def _update_entity(self, t: datetime.datetime, sample: float) -> list:
        heatmap = self.update(sample, t)
        return [heatmap]
