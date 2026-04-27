"""Ehlers' Comb Band-Pass Spectrum heatmap indicator.

The Comb Band-Pass Spectrum (cbps) displays a power heatmap of cyclic
activity over a configurable cycle-period range. Each cycle bin is
estimated by a dedicated 2-pole band-pass filter tuned to that period,
forming a "comb" filter bank. The close series is pre-conditioned by a
2-pole Butterworth highpass (cutoff = MaxPeriod) followed by a 2-pole
Super Smoother (cutoff = MinPeriod) before it enters the comb. Each bin's
power is the sum of squared band-pass outputs over the last N samples,
optionally compensated for spectral dilation (divide by N) and normalized
by a fast-attack slow-decay automatic gain control.

Reference: John F. Ehlers, "Cycle Analytics for Traders",
Code Listing 10-1 (Comb BandPass Spectrum).
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
    """Internal comb band-pass spectrum estimator (Ehlers listing 10-1)."""

    __slots__ = (
        'min_period', 'max_period', 'length_spectrum',
        'is_spectral_dilation_compensation', 'is_automatic_gain_control',
        'automatic_gain_control_decay_factor',
        'alpha_hp', 'coeff_hp0', 'coeff_hp1', 'coeff_hp2',
        'ss_c1', 'ss_c2', 'ss_c3',
        'periods', 'beta', 'alpha', 'comp',
        'close0', 'close1', 'close2',
        'hp0', 'hp1', 'hp2',
        'filt0', 'filt1', 'filt2',
        'bp', 'spectrum',
        'spectrum_min', 'spectrum_max', 'previous_spectrum_max',
    )

    def __init__(self, min_period: int, max_period: int, bandwidth: float,
                 is_spectral_dilation_compensation: bool,
                 is_automatic_gain_control: bool,
                 automatic_gain_control_decay_factor: float) -> None:
        self.min_period = min_period
        self.max_period = max_period
        length_spectrum = max_period - min_period + 1
        self.length_spectrum = length_spectrum
        self.is_spectral_dilation_compensation = is_spectral_dilation_compensation
        self.is_automatic_gain_control = is_automatic_gain_control
        self.automatic_gain_control_decay_factor = automatic_gain_control_decay_factor

        # Highpass coefficients, cutoff at MaxPeriod.
        omega_hp = 0.707 * TWO_PI / max_period
        alpha_hp = (math.cos(omega_hp) + math.sin(omega_hp) - 1) / math.cos(omega_hp)
        self.alpha_hp = alpha_hp
        self.coeff_hp0 = (1 - alpha_hp / 2) * (1 - alpha_hp / 2)
        self.coeff_hp1 = 2 * (1 - alpha_hp)
        self.coeff_hp2 = (1 - alpha_hp) * (1 - alpha_hp)

        # SuperSmoother coefficients, period = MinPeriod.
        a1 = math.exp(-1.414 * math.pi / min_period)
        b1 = 2 * a1 * math.cos(1.414 * math.pi / min_period)
        self.ss_c2 = b1
        self.ss_c3 = -a1 * a1
        self.ss_c1 = 1 - self.ss_c2 - self.ss_c3

        # Per-bin band-pass coefficients.
        self.periods = [0] * length_spectrum
        self.beta = [0.0] * length_spectrum
        self.alpha = [0.0] * length_spectrum
        self.comp = [0.0] * length_spectrum
        self.bp = [[0.0] * max_period for _ in range(length_spectrum)]
        self.spectrum = [0.0] * length_spectrum

        for i in range(length_spectrum):
            n = min_period + i
            b = math.cos(TWO_PI / n)
            gamma = 1.0 / math.cos(TWO_PI * bandwidth / n)
            a = gamma - math.sqrt(gamma * gamma - 1)
            self.periods[i] = n
            self.beta[i] = b
            self.alpha[i] = a
            self.comp[i] = float(n) if is_spectral_dilation_compensation else 1.0

        # Pre-filter state.
        self.close0 = self.close1 = self.close2 = 0.0
        self.hp0 = self.hp1 = self.hp2 = 0.0
        self.filt0 = self.filt1 = self.filt2 = 0.0

        self.spectrum_min = 0.0
        self.spectrum_max = 0.0
        self.previous_spectrum_max = 0.0

    def update(self, sample: float) -> None:
        # Shift close history.
        self.close2 = self.close1
        self.close1 = self.close0
        self.close0 = sample

        # Shift HP history and compute new HP.
        self.hp2 = self.hp1
        self.hp1 = self.hp0
        self.hp0 = self.coeff_hp0 * (self.close0 - 2 * self.close1 + self.close2) + \
            self.coeff_hp1 * self.hp1 - \
            self.coeff_hp2 * self.hp2

        # Shift Filt history and compute new Filt (SuperSmoother on HP).
        self.filt2 = self.filt1
        self.filt1 = self.filt0
        self.filt0 = self.ss_c1 * (self.hp0 + self.hp1) / 2 + \
            self.ss_c2 * self.filt1 + self.ss_c3 * self.filt2

        diff_filt = self.filt0 - self.filt2

        # AGC seeds the running max with decayed previous max.
        self.spectrum_min = sys.float_info.max
        if self.is_automatic_gain_control:
            self.spectrum_max = self.automatic_gain_control_decay_factor * self.previous_spectrum_max
        else:
            self.spectrum_max = -sys.float_info.max

        for i in range(self.length_spectrum):
            bp_row = self.bp[i]

            # Rightward shift.
            for m in range(self.max_period - 1, 0, -1):
                bp_row[m] = bp_row[m - 1]

            a = self.alpha[i]
            b = self.beta[i]
            bp_row[0] = 0.5 * (1 - a) * diff_filt + b * (1 + a) * bp_row[1] - a * bp_row[2]

            # Power: sum of (BP/comp)^2 over last N samples.
            n = self.periods[i]
            c = self.comp[i]
            pwr = 0.0
            for m in range(n):
                v = bp_row[m] / c
                pwr += v * v

            self.spectrum[i] = pwr

            if self.spectrum_max < pwr:
                self.spectrum_max = pwr
            if self.spectrum_min > pwr:
                self.spectrum_min = pwr

        self.previous_spectrum_max = self.spectrum_max


class CombBandPassSpectrum:
    """Ehlers' Comb Band-Pass Spectrum heatmap indicator."""

    def __init__(self, params: Params) -> None:
        invalid = "invalid comb band-pass spectrum parameters"
        def_min_period = 10
        def_max_period = 48
        def_bandwidth = 0.3
        def_agc_decay = 0.995
        agc_decay_epsilon = 1e-12
        bandwidth_epsilon = 1e-12

        min_period = params.min_period if params.min_period != 0 else def_min_period
        max_period = params.max_period if params.max_period != 0 else def_max_period
        bandwidth = params.bandwidth if params.bandwidth != 0 else def_bandwidth
        agc_decay = params.automatic_gain_control_decay_factor \
            if params.automatic_gain_control_decay_factor != 0 else def_agc_decay

        sdc_on = not params.disable_spectral_dilation_compensation
        agc_on = not params.disable_automatic_gain_control
        floating_norm = not params.fixed_normalization

        if min_period < 2:
            raise ValueError(f"{invalid}: MinPeriod should be >= 2")
        if max_period <= min_period:
            raise ValueError(f"{invalid}: MaxPeriod should be > MinPeriod")
        if bandwidth <= 0 or bandwidth >= 1:
            raise ValueError(f"{invalid}: Bandwidth should be in (0, 1)")
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
        if abs(bandwidth - def_bandwidth) > bandwidth_epsilon:
            flags += f", bw={bandwidth:g}"
        if not sdc_on:
            flags += ", no-sdc"
        if not agc_on:
            flags += ", no-agc"
        if agc_on and abs(agc_decay - def_agc_decay) > agc_decay_epsilon:
            flags += f", agc={agc_decay:g}"
        if not floating_norm:
            flags += ", no-fn"

        self.mnemonic = f"cbps({min_period}, {max_period}{flags}{comp_mn})"
        self.description = "Comb band-pass spectrum " + self.mnemonic

        self._estimator = _Estimator(
            min_period, max_period, bandwidth,
            sdc_on, agc_on, agc_decay,
        )
        self._prime_count = max_period
        self._window_count = 0
        self._primed = False
        self._floating_normalization = floating_norm
        self._min_param = float(min_period)
        self._max_param = float(max_period)
        self._param_res = 1.0

    def is_primed(self) -> bool:
        return self._primed

    def metadata(self) -> Metadata:
        return build_metadata(
            Identifier.COMB_BAND_PASS_SPECTRUM,
            self.mnemonic,
            self.description,
            [OutputText(mnemonic=self.mnemonic, description=self.description)],
        )

    def update(self, sample: float, t: datetime.datetime) -> Heatmap:
        if math.isnan(sample):
            return Heatmap.empty(t, self._min_param, self._max_param, self._param_res)

        self._estimator.update(sample)

        if not self._primed:
            self._window_count += 1
            if self._window_count >= self._prime_count:
                self._primed = True
            else:
                return Heatmap.empty(t, self._min_param, self._max_param, self._param_res)

        ls = self._estimator.length_spectrum

        min_ref = 0.0
        if self._floating_normalization:
            min_ref = self._estimator.spectrum_min

        max_ref = self._estimator.spectrum_max
        spectrum_range = max_ref - min_ref

        values = [0.0] * ls
        value_min = math.inf
        value_max = -math.inf

        # Spectrum is already in axis order (bin 0 = MinPeriod, bin last = MaxPeriod).
        for i in range(ls):
            v = 0.0
            if spectrum_range > 0:
                v = (self._estimator.spectrum[i] - min_ref) / spectrum_range
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
