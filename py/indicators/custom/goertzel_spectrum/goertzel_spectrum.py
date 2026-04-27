"""Goertzel Spectrum heatmap indicator."""

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
from .estimator import _Estimator
from .params import Params


_DEF_LENGTH = 64
_DEF_MIN_PERIOD = 2.0
_DEF_MAX_PERIOD = 64.0
_DEF_SPECTRUM_RESOLUTION = 1
_DEF_AGC_DECAY_FACTOR = 0.991
_AGC_DECAY_EPSILON = 1e-12


def _build_flag_tags(p: Params, sdc_on: bool, agc_on: bool,
                     floating_norm: bool) -> str:
    s = ""
    if p.is_first_order:
        s += ", fo"
    if not sdc_on:
        s += ", no-sdc"
    if not agc_on:
        s += ", no-agc"
    if agc_on and abs(p.automatic_gain_control_decay_factor - _DEF_AGC_DECAY_FACTOR) > _AGC_DECAY_EPSILON:
        s += f", agc={p.automatic_gain_control_decay_factor:g}"
    if not floating_norm:
        s += ", no-fn"
    return s


class GoertzelSpectrum:
    """Goertzel Spectrum heatmap indicator."""

    def __init__(self, params: Params) -> None:
        invalid = "invalid goertzel spectrum parameters"

        cfg_length = params.length if params.length != 0 else _DEF_LENGTH
        cfg_min_period = params.min_period if params.min_period != 0 else _DEF_MIN_PERIOD
        cfg_max_period = params.max_period if params.max_period != 0 else _DEF_MAX_PERIOD
        cfg_resolution = params.spectrum_resolution if params.spectrum_resolution != 0 else _DEF_SPECTRUM_RESOLUTION
        cfg_agc_decay = params.automatic_gain_control_decay_factor \
            if params.automatic_gain_control_decay_factor != 0 else _DEF_AGC_DECAY_FACTOR

        sdc_on = not params.disable_spectral_dilation_compensation
        agc_on = not params.disable_automatic_gain_control
        floating_norm = not params.fixed_normalization

        if cfg_length < 2:
            raise ValueError(f"{invalid}: Length should be >= 2")
        if cfg_min_period < 2:
            raise ValueError(f"{invalid}: MinPeriod should be >= 2")
        if cfg_max_period <= cfg_min_period:
            raise ValueError(f"{invalid}: MaxPeriod should be > MinPeriod")
        if cfg_max_period > 2 * cfg_length:
            raise ValueError(f"{invalid}: MaxPeriod should be <= 2 * Length")
        if cfg_resolution < 1:
            raise ValueError(f"{invalid}: SpectrumResolution should be >= 1")
        if agc_on and (cfg_agc_decay <= 0 or cfg_agc_decay >= 1):
            raise ValueError(
                f"{invalid}: AutomaticGainControlDecayFactor should be in (0, 1)")

        bc = params.bar_component if params.bar_component is not None else BarComponent.MEDIAN
        qc = params.quote_component if params.quote_component is not None else DEFAULT_QUOTE_COMPONENT
        tc = params.trade_component if params.trade_component is not None else DEFAULT_TRADE_COMPONENT

        self._bar_func = bar_component_value(bc)
        self._quote_func = quote_component_value(qc)
        self._trade_func = trade_component_value(tc)

        # Build a temporary params copy with resolved defaults for mnemonic building.
        resolved = Params(
            length=cfg_length,
            min_period=cfg_min_period,
            max_period=cfg_max_period,
            spectrum_resolution=cfg_resolution,
            is_first_order=params.is_first_order,
            disable_spectral_dilation_compensation=params.disable_spectral_dilation_compensation,
            disable_automatic_gain_control=params.disable_automatic_gain_control,
            automatic_gain_control_decay_factor=cfg_agc_decay,
            fixed_normalization=params.fixed_normalization,
        )

        comp_mn = component_triple_mnemonic(bc, qc, tc)
        flags = _build_flag_tags(resolved, sdc_on, agc_on, floating_norm)
        self.mnemonic = f"gspect({cfg_length}, {cfg_min_period:g}, " \
            f"{cfg_max_period:g}, {cfg_resolution}{flags}{comp_mn})"
        self.description = "Goertzel spectrum " + self.mnemonic

        self._estimator = _Estimator(
            cfg_length, cfg_min_period, cfg_max_period, cfg_resolution,
            params.is_first_order, sdc_on, agc_on, cfg_agc_decay,
        )

        self._window_count = 0
        self._last_index = cfg_length - 1
        self._primed = False
        self._floating_normalization = floating_norm
        self._min_parameter_value = cfg_min_period
        self._max_parameter_value = cfg_max_period
        self._parameter_resolution = float(cfg_resolution)

    def is_primed(self) -> bool:
        return self._primed

    def metadata(self) -> Metadata:
        return build_metadata(
            Identifier.GOERTZEL_SPECTRUM,
            self.mnemonic,
            self.description,
            [OutputText(mnemonic=self.mnemonic, description=self.description)],
        )

    def update(self, sample: float, t: datetime.datetime) -> Heatmap:
        """Feed next sample, return heatmap column."""
        if math.isnan(sample):
            return Heatmap.empty(t, self._min_parameter_value,
                                 self._max_parameter_value, self._parameter_resolution)

        window = self._estimator.input_series

        if self._primed:
            window[:self._last_index] = window[1:]
            window[self._last_index] = sample
        else:
            window[self._window_count] = sample
            self._window_count += 1
            if self._window_count == self._estimator.length:
                self._primed = True

        if not self._primed:
            return Heatmap.empty(t, self._min_parameter_value,
                                 self._max_parameter_value, self._parameter_resolution)

        self._estimator.calculate()

        length_spectrum = self._estimator.length_spectrum

        min_ref = self._estimator.spectrum_min if self._floating_normalization else 0.0
        max_ref = self._estimator.spectrum_max
        spectrum_range = max_ref - min_ref

        values = [0.0] * length_spectrum
        value_min = math.inf
        value_max = -math.inf

        for i in range(length_spectrum):
            v = (self._estimator.spectrum[length_spectrum - 1 - i] - min_ref) / spectrum_range
            values[i] = v
            if v < value_min:
                value_min = v
            if v > value_max:
                value_max = v

        return Heatmap(t, self._min_parameter_value, self._max_parameter_value,
                       self._parameter_resolution, value_min, value_max, values)

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
