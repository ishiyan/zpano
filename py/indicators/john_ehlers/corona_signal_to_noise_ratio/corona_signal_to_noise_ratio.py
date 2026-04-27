"""Ehlers' Corona Signal-to-Noise Ratio heatmap indicator."""

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
from ..corona.corona import Corona
from ..corona.params import CoronaParams
from .params import Params

_HL_BUFFER_SIZE = 5
_HL_MEDIAN_INDEX = 2
_AVG_SAMPLE_ALPHA = 0.1
_AVG_SAMPLE_ONE_MINUS = 0.9
_SIGNAL_EMA_ALPHA = 0.2
_SIGNAL_EMA_ONE_MINUS = 0.9  # Intentional: sums to 1.1 per Ehlers
_NOISE_EMA_ALPHA = 0.1
_NOISE_EMA_ONE_MINUS = 0.9
_RATIO_OFFSET_DB = 3.5
_RATIO_UPPER_DB = 10.0
_DB_GAIN = 20.0
_WIDTH_LOW_RATIO_THRESHOLD = 0.5
_WIDTH_BASELINE = 0.2
_WIDTH_SLOPE = 0.4
_RASTER_BLEND_EXPONENT = 0.8
_RASTER_BLEND_HALF = 0.5
_RASTER_NEGATIVE_ARG_CUTOFF = 1.0


class CoronaSignalToNoiseRatio:
    """Ehlers' Corona Signal-to-Noise Ratio heatmap indicator."""

    def __init__(self, params: Params) -> None:
        invalid = "invalid corona signal to noise ratio parameters"

        cfg_raster_len = params.raster_length if params.raster_length != 0 else 50
        cfg_max_raster = params.max_raster_value if params.max_raster_value != 0 else 20.0
        cfg_min_param = params.min_parameter_value if params.min_parameter_value != 0 else 1.0
        cfg_max_param = params.max_parameter_value if params.max_parameter_value != 0 else 11.0
        cfg_hp_cutoff = params.high_pass_filter_cutoff if params.high_pass_filter_cutoff != 0 else 30
        cfg_min_period = params.minimal_period if params.minimal_period != 0 else 6
        cfg_max_period = params.maximal_period if params.maximal_period != 0 else 30

        if cfg_raster_len < 2:
            raise ValueError(f"{invalid}: RasterLength should be >= 2")
        if cfg_max_raster <= 0:
            raise ValueError(f"{invalid}: MaxRasterValue should be > 0")
        if cfg_min_param < 0:
            raise ValueError(f"{invalid}: MinParameterValue should be >= 0")
        if cfg_max_param <= cfg_min_param:
            raise ValueError(f"{invalid}: MaxParameterValue should be > MinParameterValue")
        if cfg_hp_cutoff < 2:
            raise ValueError(f"{invalid}: HighPassFilterCutoff should be >= 2")
        if cfg_min_period < 2:
            raise ValueError(f"{invalid}: MinimalPeriod should be >= 2")
        if cfg_max_period <= cfg_min_period:
            raise ValueError(f"{invalid}: MaximalPeriod should be > MinimalPeriod")

        bc = params.bar_component if params.bar_component is not None else BarComponent.MEDIAN
        qc = params.quote_component if params.quote_component is not None else DEFAULT_QUOTE_COMPONENT
        tc = params.trade_component if params.trade_component is not None else DEFAULT_TRADE_COMPONENT

        self._bar_func = bar_component_value(bc)
        self._quote_func = quote_component_value(qc)
        self._trade_func = trade_component_value(tc)

        self._corona = Corona(CoronaParams(
            high_pass_filter_cutoff=cfg_hp_cutoff,
            minimal_period=cfg_min_period,
            maximal_period=cfg_max_period,
        ))

        self._raster_length = cfg_raster_len
        self._raster_step = cfg_max_raster / cfg_raster_len
        self._max_raster_value = cfg_max_raster
        self._min_parameter_value = cfg_min_param
        self._max_parameter_value = cfg_max_param
        self._parameter_resolution = (cfg_raster_len - 1) / (cfg_max_param - cfg_min_param)
        self._raster = [0.0] * cfg_raster_len

        self._hl_buffer = [0.0] * _HL_BUFFER_SIZE
        self._avg_sample_previous = 0.0
        self._signal_previous = 0.0
        self._noise_previous = 0.0
        self._signal_to_noise_ratio = float('nan')
        self._is_started = False

        comp_mn = component_triple_mnemonic(bc, qc, tc)
        self.mnemonic = f"csnr({cfg_raster_len}, {cfg_max_raster:g}, " \
            f"{cfg_min_param:g}, {cfg_max_param:g}, {cfg_hp_cutoff}{comp_mn})"
        self.description = "Corona signal to noise ratio " + self.mnemonic
        self._mnemonic_snr = f"csnr-snr({cfg_hp_cutoff}{comp_mn})"
        self._description_snr = "Corona signal to noise ratio scalar " + self._mnemonic_snr

    def is_primed(self) -> bool:
        return self._corona.is_primed()

    def metadata(self) -> Metadata:
        return build_metadata(
            Identifier.CORONA_SIGNAL_TO_NOISE_RATIO,
            self.mnemonic,
            self.description,
            [
                OutputText(mnemonic=self.mnemonic, description=self.description),
                OutputText(mnemonic=self._mnemonic_snr, description=self._description_snr),
            ],
        )

    def update(self, sample: float, sample_low: float, sample_high: float,
               t: datetime.datetime) -> tuple:
        """Returns (heatmap, signal_to_noise_ratio)."""
        if math.isnan(sample):
            return (Heatmap.empty(t, self._min_parameter_value,
                                  self._max_parameter_value, self._parameter_resolution),
                    float('nan'))

        primed = self._corona.update(sample)

        if not self._is_started:
            self._avg_sample_previous = sample
            self._hl_buffer[_HL_BUFFER_SIZE - 1] = sample_high - sample_low
            self._is_started = True
            return (Heatmap.empty(t, self._min_parameter_value,
                                  self._max_parameter_value, self._parameter_resolution),
                    float('nan'))

        max_amp_sq = self._corona.maximal_amplitude_squared

        avg_sample = _AVG_SAMPLE_ALPHA * sample + _AVG_SAMPLE_ONE_MINUS * self._avg_sample_previous
        self._avg_sample_previous = avg_sample

        if abs(avg_sample) > 0 or max_amp_sq > 0:
            self._signal_previous = _SIGNAL_EMA_ALPHA * math.sqrt(max_amp_sq) + \
                _SIGNAL_EMA_ONE_MINUS * self._signal_previous

        # Shift H-L ring buffer left; push new value.
        for i in range(_HL_BUFFER_SIZE - 1):
            self._hl_buffer[i] = self._hl_buffer[i + 1]
        self._hl_buffer[_HL_BUFFER_SIZE - 1] = sample_high - sample_low

        ratio = 0.0
        if abs(avg_sample) > 0:
            hl_sorted = sorted(self._hl_buffer)
            self._noise_previous = _NOISE_EMA_ALPHA * hl_sorted[_HL_MEDIAN_INDEX] + \
                _NOISE_EMA_ONE_MINUS * self._noise_previous

            if abs(self._noise_previous) > 0:
                ratio = _DB_GAIN * math.log10(self._signal_previous / self._noise_previous) + \
                    _RATIO_OFFSET_DB
                if ratio < 0:
                    ratio = 0
                elif ratio > _RATIO_UPPER_DB:
                    ratio = _RATIO_UPPER_DB
                ratio /= _RATIO_UPPER_DB  # in [0, 1]

        self._signal_to_noise_ratio = (self._max_parameter_value - self._min_parameter_value) * \
            ratio + self._min_parameter_value

        # Raster update.
        width = 0.0
        if ratio <= _WIDTH_LOW_RATIO_THRESHOLD:
            width = _WIDTH_BASELINE - _WIDTH_SLOPE * ratio

        ratio_scaled_to_raster_length = round(ratio * self._raster_length)
        ratio_scaled_to_max_raster_value = ratio * self._max_raster_value

        for i in range(self._raster_length):
            value = self._raster[i]

            if i == ratio_scaled_to_raster_length:
                value *= 0.5
            elif width == 0:
                pass  # Above high-ratio threshold: handled by ratio>0.5 override below.
            else:
                argument = (ratio_scaled_to_max_raster_value - self._raster_step * i) / width
                if i < ratio_scaled_to_raster_length:
                    value = _RASTER_BLEND_HALF * (math.pow(argument, _RASTER_BLEND_EXPONENT) + value)
                else:
                    argument = -argument
                    if argument > _RASTER_NEGATIVE_ARG_CUTOFF:
                        value = _RASTER_BLEND_HALF * (math.pow(argument, _RASTER_BLEND_EXPONENT) + value)
                    else:
                        value = self._max_raster_value

            if value < 0:
                value = 0
            elif value > self._max_raster_value:
                value = self._max_raster_value

            if ratio > _WIDTH_LOW_RATIO_THRESHOLD:
                value = self._max_raster_value

            self._raster[i] = value

        if not primed:
            return (Heatmap.empty(t, self._min_parameter_value,
                                  self._max_parameter_value, self._parameter_resolution),
                    float('nan'))

        values = list(self._raster)
        value_min = min(values)
        value_max = max(values)

        heatmap = Heatmap(t, self._min_parameter_value, self._max_parameter_value,
                          self._parameter_resolution, value_min, value_max, values)

        return heatmap, self._signal_to_noise_ratio

    def update_scalar(self, sample: Scalar) -> list:
        return self._update_entity(sample.time, sample.value, sample.value, sample.value)

    def update_bar(self, sample: Bar) -> list:
        return self._update_entity(sample.time, self._bar_func(sample), sample.low, sample.high)

    def update_quote(self, sample: Quote) -> list:
        v = self._quote_func(sample)
        return self._update_entity(sample.time, v, v, v)

    def update_trade(self, sample: Trade) -> list:
        v = self._trade_func(sample)
        return self._update_entity(sample.time, v, v, v)

    def _update_entity(self, t: datetime.datetime, sample: float,
                       low: float, high: float) -> list:
        heatmap, snr = self.update(sample, low, high, t)
        return [heatmap, Scalar(t, snr)]
