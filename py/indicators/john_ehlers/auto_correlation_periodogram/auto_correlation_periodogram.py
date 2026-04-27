"""Ehlers' Autocorrelation Periodogram heatmap.

The Autocorrelation Periodogram (ACP) displays a power heatmap of cyclic
activity by taking a discrete Fourier transform of the autocorrelation
function. The close series is pre-conditioned by a 2-pole Butterworth
highpass (cutoff = MaxPeriod) followed by a 2-pole Super Smoother
(cutoff = MinPeriod). The autocorrelation function is evaluated at lags
0..MaxPeriod using Pearson correlation with a fixed averaging length.
Each period bin's squared-sum Fourier magnitude is exponentially smoothed,
fast-attack / slow-decay AGC normalized, and displayed.

Reference: John F. Ehlers, "Cycle Analytics for Traders", Code Listing 8-3.
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
DFT_LAG_START = 3


class _Estimator:
    """Internal estimator: HP + SS pre-filter, Pearson correlation, DFT, smoothing, AGC."""

    __slots__ = (
        'min_period', 'max_period', 'averaging_length', 'length_spectrum', 'filt_buffer_len',
        'is_spectral_squaring', 'is_smoothing', 'is_automatic_gain_control',
        'automatic_gain_control_decay_factor',
        'coeff_hp0', 'coeff_hp1', 'coeff_hp2', 'ss_c1', 'ss_c2', 'ss_c3',
        'cos_tab', 'sin_tab',
        'close0', 'close1', 'close2', 'hp0', 'hp1', 'hp2',
        'filt', 'corr', 'r_previous', 'spectrum',
        'spectrum_min', 'spectrum_max', 'previous_spectrum_max',
    )

    def __init__(self, min_period: int, max_period: int, averaging_length: int,
                 is_spectral_squaring: bool, is_smoothing: bool,
                 is_automatic_gain_control: bool,
                 automatic_gain_control_decay_factor: float) -> None:
        self.min_period = min_period
        self.max_period = max_period
        self.averaging_length = averaging_length
        self.length_spectrum = max_period - min_period + 1
        self.filt_buffer_len = max_period + averaging_length
        self.is_spectral_squaring = is_spectral_squaring
        self.is_smoothing = is_smoothing
        self.is_automatic_gain_control = is_automatic_gain_control
        self.automatic_gain_control_decay_factor = automatic_gain_control_decay_factor

        corr_len = max_period + 1

        # Highpass coefficients, cutoff at max_period.
        omega_hp = 0.707 * TWO_PI / max_period
        alpha_hp = (math.cos(omega_hp) + math.sin(omega_hp) - 1) / math.cos(omega_hp)
        self.coeff_hp0 = (1 - alpha_hp / 2) * (1 - alpha_hp / 2)
        self.coeff_hp1 = 2 * (1 - alpha_hp)
        self.coeff_hp2 = (1 - alpha_hp) * (1 - alpha_hp)

        # SuperSmoother coefficients, period = min_period.
        a1 = math.exp(-1.414 * math.pi / min_period)
        b1 = 2 * a1 * math.cos(1.414 * math.pi / min_period)
        self.ss_c2 = b1
        self.ss_c3 = -a1 * a1
        self.ss_c1 = 1 - self.ss_c2 - self.ss_c3

        # DFT basis tables.
        ls = self.length_spectrum
        self.cos_tab = [[0.0] * corr_len for _ in range(ls)]
        self.sin_tab = [[0.0] * corr_len for _ in range(ls)]

        for i in range(ls):
            period = min_period + i
            for n in range(DFT_LAG_START, corr_len):
                angle = TWO_PI * n / period
                self.cos_tab[i][n] = math.cos(angle)
                self.sin_tab[i][n] = math.sin(angle)

        # State.
        self.close0 = 0.0
        self.close1 = 0.0
        self.close2 = 0.0
        self.hp0 = 0.0
        self.hp1 = 0.0
        self.hp2 = 0.0
        self.filt = [0.0] * self.filt_buffer_len
        self.corr = [0.0] * corr_len
        self.r_previous = [0.0] * ls
        self.spectrum = [0.0] * ls
        self.spectrum_min = 0.0
        self.spectrum_max = 0.0
        self.previous_spectrum_max = 0.0

    def update(self, sample: float) -> None:
        # Pre-filter cascade.
        self.close2 = self.close1
        self.close1 = self.close0
        self.close0 = sample

        self.hp2 = self.hp1
        self.hp1 = self.hp0
        self.hp0 = self.coeff_hp0 * (self.close0 - 2 * self.close1 + self.close2) + \
            self.coeff_hp1 * self.hp1 - self.coeff_hp2 * self.hp2

        # Shift filt rightward.
        filt = self.filt
        for k in range(self.filt_buffer_len - 1, 0, -1):
            filt[k] = filt[k - 1]

        filt[0] = self.ss_c1 * (self.hp0 + self.hp1) / 2 + \
            self.ss_c2 * filt[1] + self.ss_c3 * filt[2]

        # Pearson correlation per lag [0..max_period], fixed M = averaging_length.
        m = self.averaging_length
        max_period = self.max_period
        corr = self.corr

        for lag in range(max_period + 1):
            sx = sy = sxx = syy = sxy = 0.0
            for c in range(m):
                x = filt[c]
                y = filt[lag + c]
                sx += x
                sy += y
                sxx += x * x
                syy += y * y
                sxy += x * y

            denom = (m * sxx - sx * sx) * (m * syy - sy * sy)

            r = 0.0
            if denom > 0:
                r = (m * sxy - sx * sy) / math.sqrt(denom)

            corr[lag] = r

        # DFT of correlation function, per period bin.
        ls = self.length_spectrum
        cos_tab = self.cos_tab
        sin_tab = self.sin_tab
        spectrum = self.spectrum
        r_previous = self.r_previous

        if self.is_automatic_gain_control:
            self.spectrum_max = self.automatic_gain_control_decay_factor * self.previous_spectrum_max
        else:
            self.spectrum_max = -sys.float_info.max

        self.spectrum_min = sys.float_info.max

        # Pass 1: compute raw R values and track running max for AGC.
        for i in range(ls):
            cos_row = cos_tab[i]
            sin_row = sin_tab[i]

            cos_part = sin_part = 0.0
            for n in range(DFT_LAG_START, max_period + 1):
                cos_part += corr[n] * cos_row[n]
                sin_part += corr[n] * sin_row[n]

            sq_sum = cos_part * cos_part + sin_part * sin_part

            raw = sq_sum
            if self.is_spectral_squaring:
                raw = sq_sum * sq_sum

            if self.is_smoothing:
                r_val = 0.2 * raw + 0.8 * r_previous[i]
            else:
                r_val = raw

            r_previous[i] = r_val
            spectrum[i] = r_val

            if self.spectrum_max < r_val:
                self.spectrum_max = r_val

        self.previous_spectrum_max = self.spectrum_max

        # Pass 2: normalize against running max and track normalized min.
        if self.spectrum_max > 0:
            for i in range(ls):
                v = spectrum[i] / self.spectrum_max
                spectrum[i] = v
                if self.spectrum_min > v:
                    self.spectrum_min = v
        else:
            for i in range(ls):
                spectrum[i] = 0.0
            self.spectrum_min = 0.0


class AutoCorrelationPeriodogram:
    """Ehlers' Autocorrelation Periodogram heatmap."""

    def __init__(self, params: Params) -> None:
        invalid = "invalid autocorrelation periodogram parameters"
        def_min_period = 10
        def_max_period = 48
        def_averaging_len = 3
        def_agc_decay = 0.995
        agc_decay_epsilon = 1e-12

        min_period = params.min_period if params.min_period != 0 else def_min_period
        max_period = params.max_period if params.max_period != 0 else def_max_period
        averaging_length = params.averaging_length if params.averaging_length != 0 else def_averaging_len
        agc_decay = params.automatic_gain_control_decay_factor \
            if params.automatic_gain_control_decay_factor != 0 else def_agc_decay

        squaring_on = not params.disable_spectral_squaring
        smoothing_on = not params.disable_smoothing
        agc_on = not params.disable_automatic_gain_control
        floating_norm = not params.fixed_normalization

        if min_period < 2:
            raise ValueError(f"{invalid}: MinPeriod should be >= 2")
        if max_period <= min_period:
            raise ValueError(f"{invalid}: MaxPeriod should be > MinPeriod")
        if averaging_length < 1:
            raise ValueError(f"{invalid}: AveragingLength should be >= 1")
        if agc_on and (agc_decay <= 0 or agc_decay >= 1):
            raise ValueError(
                f"{invalid}: AutomaticGainControlDecayFactor should be in (0, 1)")

        # Default bar component: BarMedianPrice (Ehlers reference).
        bc = params.bar_component if params.bar_component is not None else BarComponent.MEDIAN
        qc = params.quote_component if params.quote_component is not None else DEFAULT_QUOTE_COMPONENT
        tc = params.trade_component if params.trade_component is not None else DEFAULT_TRADE_COMPONENT

        self._bar_func = bar_component_value(bc)
        self._quote_func = quote_component_value(qc)
        self._trade_func = trade_component_value(tc)

        comp_mn = component_triple_mnemonic(bc, qc, tc)

        # Build flag tags.
        flags = ""
        if averaging_length != def_averaging_len:
            flags += f", average={averaging_length}"
        if not squaring_on:
            flags += ", no-sqr"
        if not smoothing_on:
            flags += ", no-smooth"
        if not agc_on:
            flags += ", no-agc"
        if agc_on and abs(agc_decay - def_agc_decay) > agc_decay_epsilon:
            flags += f", agc={agc_decay:g}"
        if not floating_norm:
            flags += ", no-fn"

        self.mnemonic = f"acp({min_period}, {max_period}{flags}{comp_mn})"
        self.description = "Autocorrelation periodogram " + self.mnemonic

        self._estimator = _Estimator(
            min_period, max_period, averaging_length,
            squaring_on, smoothing_on, agc_on, agc_decay,
        )
        self._prime_count = self._estimator.filt_buffer_len
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
            Identifier.AUTO_CORRELATION_PERIODOGRAM,
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

        length_spectrum = self._estimator.length_spectrum

        min_ref = 0.0
        if self._floating_normalization:
            min_ref = self._estimator.spectrum_min

        max_ref = 1.0
        spectrum_range = max_ref - min_ref

        values = [0.0] * length_spectrum
        value_min = math.inf
        value_max = -math.inf

        for i in range(length_spectrum):
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
