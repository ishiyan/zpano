"""Ehlers' Autocorrelation Indicator heatmap.

The Autocorrelation Indicator (ACI) displays a heatmap of Pearson correlation
coefficients between the current filtered series and a lagged copy of itself,
across a configurable lag range. The close series is pre-conditioned by a 2-pole
Butterworth highpass (cutoff = MaxLag) followed by a 2-pole Super Smoother
(cutoff = SmoothingPeriod) before the correlation bank is evaluated. Each bin's
value is rescaled from the Pearson [-1, 1] range into [0, 1] via 0.5*(r + 1).

Reference: John F. Ehlers, "Cycle Analytics for Traders", Code Listing 8-2.
"""

import math
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
    """Internal estimator: HP + SuperSmoother pre-filter, then Pearson correlation bank."""

    __slots__ = (
        'min_lag', 'max_lag', 'averaging_length', 'length_spectrum', 'filt_buffer_len',
        'coeff_hp0', 'coeff_hp1', 'coeff_hp2', 'ss_c1', 'ss_c2', 'ss_c3',
        'close0', 'close1', 'close2', 'hp0', 'hp1', 'hp2',
        'filt', 'spectrum',
    )

    def __init__(self, min_lag: int, max_lag: int, smoothing_period: int,
                 averaging_length: int) -> None:
        self.min_lag = min_lag
        self.max_lag = max_lag
        self.averaging_length = averaging_length
        self.length_spectrum = max_lag - min_lag + 1

        m_max = averaging_length if averaging_length > 0 else max_lag
        self.filt_buffer_len = max_lag + m_max

        # Highpass coefficients, cutoff at max_lag.
        omega_hp = 0.707 * TWO_PI / max_lag
        alpha_hp = (math.cos(omega_hp) + math.sin(omega_hp) - 1) / math.cos(omega_hp)
        self.coeff_hp0 = (1 - alpha_hp / 2) * (1 - alpha_hp / 2)
        self.coeff_hp1 = 2 * (1 - alpha_hp)
        self.coeff_hp2 = (1 - alpha_hp) * (1 - alpha_hp)

        # SuperSmoother coefficients, period = smoothing_period.
        a1 = math.exp(-1.414 * math.pi / smoothing_period)
        b1 = 2 * a1 * math.cos(1.414 * math.pi / smoothing_period)
        self.ss_c2 = b1
        self.ss_c3 = -a1 * a1
        self.ss_c1 = 1 - self.ss_c2 - self.ss_c3

        # State.
        self.close0 = 0.0
        self.close1 = 0.0
        self.close2 = 0.0
        self.hp0 = 0.0
        self.hp1 = 0.0
        self.hp2 = 0.0
        self.filt = [0.0] * self.filt_buffer_len
        self.spectrum = [0.0] * self.length_spectrum

    def update(self, sample: float) -> None:
        # Shift close history.
        self.close2 = self.close1
        self.close1 = self.close0
        self.close0 = sample

        # HP.
        self.hp2 = self.hp1
        self.hp1 = self.hp0
        self.hp0 = self.coeff_hp0 * (self.close0 - 2 * self.close1 + self.close2) + \
            self.coeff_hp1 * self.hp1 - self.coeff_hp2 * self.hp2

        # Shift filt rightward.
        filt = self.filt
        for k in range(self.filt_buffer_len - 1, 0, -1):
            filt[k] = filt[k - 1]

        # SuperSmoother.
        filt[0] = self.ss_c1 * (self.hp0 + self.hp1) / 2 + \
            self.ss_c2 * filt[1] + self.ss_c3 * filt[2]

        # Pearson correlation per lag.
        min_lag = self.min_lag
        avg_len = self.averaging_length
        length_spectrum = self.length_spectrum
        spectrum = self.spectrum

        for i in range(length_spectrum):
            lag = min_lag + i
            m = avg_len if avg_len > 0 else lag

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

            spectrum[i] = 0.5 * (r + 1)


class AutoCorrelationIndicator:
    """Ehlers' Autocorrelation Indicator heatmap."""

    def __init__(self, params: Params) -> None:
        invalid = "invalid autocorrelation indicator parameters"
        def_min_lag = 3
        def_max_lag = 48
        def_smoothing = 10
        def_averaging_len = 0

        min_lag = params.min_lag if params.min_lag != 0 else def_min_lag
        max_lag = params.max_lag if params.max_lag != 0 else def_max_lag
        smoothing_period = params.smoothing_period if params.smoothing_period != 0 else def_smoothing
        averaging_length = params.averaging_length

        if min_lag < 1:
            raise ValueError(f"{invalid}: MinLag should be >= 1")
        if max_lag <= min_lag:
            raise ValueError(f"{invalid}: MaxLag should be > MinLag")
        if smoothing_period < 2:
            raise ValueError(f"{invalid}: SmoothingPeriod should be >= 2")
        if averaging_length < 0:
            raise ValueError(f"{invalid}: AveragingLength should be >= 0")

        # Default bar component: BarMedianPrice (Ehlers reference).
        bc = params.bar_component if params.bar_component is not None else BarComponent.MEDIAN
        qc = params.quote_component if params.quote_component is not None else DEFAULT_QUOTE_COMPONENT
        tc = params.trade_component if params.trade_component is not None else DEFAULT_TRADE_COMPONENT

        self._bar_func = bar_component_value(bc)
        self._quote_func = quote_component_value(qc)
        self._trade_func = trade_component_value(tc)

        comp_mn = component_triple_mnemonic(bc, qc, tc)

        flags = ""
        if averaging_length != def_averaging_len:
            flags += f", average={averaging_length}"

        self.mnemonic = f"aci({min_lag}, {max_lag}, {smoothing_period}{flags}{comp_mn})"
        self.description = "Autocorrelation indicator " + self.mnemonic

        self._estimator = _Estimator(min_lag, max_lag, smoothing_period, averaging_length)
        self._prime_count = self._estimator.filt_buffer_len
        self._window_count = 0
        self._primed = False
        self._min_param = float(min_lag)
        self._max_param = float(max_lag)
        self._param_res = 1.0

    def is_primed(self) -> bool:
        return self._primed

    def metadata(self) -> Metadata:
        return build_metadata(
            Identifier.AUTO_CORRELATION_INDICATOR,
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

        spectrum = self._estimator.spectrum
        values = list(spectrum)
        value_min = min(values)
        value_max = max(values)

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
