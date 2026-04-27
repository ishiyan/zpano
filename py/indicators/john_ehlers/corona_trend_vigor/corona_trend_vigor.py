"""Ehlers' Corona Trend Vigor heatmap indicator."""

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

_BP_DELTA = 0.1
_RATIO_NEW_COEF = 0.33
_RATIO_PREVIOUS_COEF = 0.67
_VIGOR_MID_LOW = 0.3
_VIGOR_MID_HIGH = 0.7
_VIGOR_MID = 0.5
_WIDTH_EDGE = 0.01
_RASTER_BLEND_SCALE = 0.8
_RASTER_BLEND_PREVIOUS = 0.2
_RASTER_BLEND_HALF = 0.5
_RASTER_BLEND_EXPONENT = 0.85
_RATIO_LIMIT = 10.0
_VIGOR_SCALE = 0.05


class CoronaTrendVigor:
    """Ehlers' Corona Trend Vigor heatmap indicator."""

    def __init__(self, params: Params) -> None:
        invalid = "invalid corona trend vigor parameters"

        cfg_raster_len = params.raster_length if params.raster_length != 0 else 50
        cfg_max_raster = params.max_raster_value if params.max_raster_value != 0 else 20.0

        if params.min_parameter_value == 0 and params.max_parameter_value == 0:
            cfg_min_param = -10.0
            cfg_max_param = 10.0
        else:
            cfg_min_param = params.min_parameter_value
            cfg_max_param = params.max_parameter_value

        cfg_hp_cutoff = params.high_pass_filter_cutoff if params.high_pass_filter_cutoff != 0 else 30
        cfg_min_period = params.minimal_period if params.minimal_period != 0 else 6
        cfg_max_period = params.maximal_period if params.maximal_period != 0 else 30

        if cfg_raster_len < 2:
            raise ValueError(f"{invalid}: RasterLength should be >= 2")
        if cfg_max_raster <= 0:
            raise ValueError(f"{invalid}: MaxRasterValue should be > 0")
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

        buf_len = self._corona.maximal_period_times_two
        self._sample_buffer = [0.0] * buf_len
        self._sample_count = 0
        self._sample_previous = 0.0
        self._sample_previous2 = 0.0
        self._band_pass_previous = 0.0
        self._band_pass_previous2 = 0.0
        self._ratio_previous = 0.0
        self._trend_vigor = float('nan')

        comp_mn = component_triple_mnemonic(bc, qc, tc)
        self.mnemonic = f"ctv({cfg_raster_len}, {cfg_max_raster:g}, " \
            f"{cfg_min_param:g}, {cfg_max_param:g}, {cfg_hp_cutoff}{comp_mn})"
        self.description = "Corona trend vigor " + self.mnemonic
        self._mnemonic_tv = f"ctv-tv({cfg_hp_cutoff}{comp_mn})"
        self._description_tv = "Corona trend vigor scalar " + self._mnemonic_tv

    def is_primed(self) -> bool:
        return self._corona.is_primed()

    def metadata(self) -> Metadata:
        return build_metadata(
            Identifier.CORONA_TREND_VIGOR,
            self.mnemonic,
            self.description,
            [
                OutputText(mnemonic=self.mnemonic, description=self.description),
                OutputText(mnemonic=self._mnemonic_tv, description=self._description_tv),
            ],
        )

    def update(self, sample: float, t: datetime.datetime) -> tuple:
        """Returns (heatmap, trend_vigor)."""
        if math.isnan(sample):
            return (Heatmap.empty(t, self._min_parameter_value,
                                  self._max_parameter_value, self._parameter_resolution),
                    float('nan'))

        primed = self._corona.update(sample)
        self._sample_count += 1

        buf_last = len(self._sample_buffer) - 1

        if self._sample_count == 1:
            self._sample_previous = sample
            self._sample_buffer[buf_last] = sample
            return (Heatmap.empty(t, self._min_parameter_value,
                                  self._max_parameter_value, self._parameter_resolution),
                    float('nan'))

        # Bandpass filter at the dominant cycle median period.
        dcm = self._corona.dominant_cycle_median
        omega = 2.0 * math.pi / dcm if dcm != 0 else math.inf
        beta2 = math.cos(omega)
        cos_val = math.cos(omega * 2 * _BP_DELTA)
        if cos_val == 0:
            gamma2 = math.copysign(math.inf, 1.0)
        else:
            gamma2 = 1.0 / cos_val
        disc = gamma2 * gamma2 - 1.0
        alpha2 = gamma2 - math.sqrt(disc) if disc >= 0 else 0.0
        band_pass = 0.5 * (1 - alpha2) * (sample - self._sample_previous2) + \
            beta2 * (1 + alpha2) * self._band_pass_previous - \
            alpha2 * self._band_pass_previous2

        # Quadrature = derivative / omega.
        if omega == 0:
            diff = band_pass - self._band_pass_previous
            quadrature2 = math.copysign(math.inf, diff) if diff != 0 else 0.0
        else:
            quadrature2 = (band_pass - self._band_pass_previous) / omega

        self._band_pass_previous2 = self._band_pass_previous
        self._band_pass_previous = band_pass
        self._sample_previous2 = self._sample_previous
        self._sample_previous = sample

        # Left-shift sample buffer and append new sample.
        for i in range(buf_last):
            self._sample_buffer[i] = self._sample_buffer[i + 1]
        self._sample_buffer[buf_last] = sample

        # Cycle amplitude.
        amplitude2 = math.sqrt(band_pass * band_pass + quadrature2 * quadrature2)

        # Trend over cycle period.
        dcm = self._corona.dominant_cycle_median
        if math.isinf(dcm) or math.isnan(dcm):
            cycle_period = len(self._sample_buffer)
        else:
            cycle_period = int(dcm - 1)
            if cycle_period > len(self._sample_buffer):
                cycle_period = len(self._sample_buffer)
        if cycle_period < 1:
            cycle_period = 1

        lookback = cycle_period
        if self._sample_count < lookback:
            lookback = self._sample_count

        trend = sample - self._sample_buffer[len(self._sample_buffer) - lookback]

        ratio = 0.0
        if abs(trend) > 0 and amplitude2 > 0:
            ratio = _RATIO_NEW_COEF * trend / amplitude2 + _RATIO_PREVIOUS_COEF * self._ratio_previous

        if ratio > _RATIO_LIMIT:
            ratio = _RATIO_LIMIT
        elif ratio < -_RATIO_LIMIT:
            ratio = -_RATIO_LIMIT

        self._ratio_previous = ratio

        # vigor in [0, 1]
        vigor = _VIGOR_SCALE * (ratio + _RATIO_LIMIT)

        # Width.
        if vigor >= _VIGOR_MID_LOW and vigor < _VIGOR_MID:
            width = vigor - (_VIGOR_MID_LOW - _WIDTH_EDGE)
        elif vigor >= _VIGOR_MID and vigor <= _VIGOR_MID_HIGH:
            width = (_VIGOR_MID_HIGH + _WIDTH_EDGE) - vigor
        else:
            width = _WIDTH_EDGE

        self._trend_vigor = (self._max_parameter_value - self._min_parameter_value) * \
            vigor + self._min_parameter_value

        vigor_scaled_to_raster_length = round(self._raster_length * vigor)
        vigor_scaled_to_max_raster_value = vigor * self._max_raster_value

        for i in range(self._raster_length):
            value = self._raster[i]

            if i == vigor_scaled_to_raster_length:
                value *= _RASTER_BLEND_HALF
            else:
                argument = vigor_scaled_to_max_raster_value - self._raster_step * i
                if i > vigor_scaled_to_raster_length:
                    argument = -argument

                if width > 0:
                    value = _RASTER_BLEND_SCALE * \
                        (math.pow(argument / width, _RASTER_BLEND_EXPONENT) + _RASTER_BLEND_PREVIOUS * value)

            if value < 0:
                value = 0
            elif value > self._max_raster_value or vigor < _VIGOR_MID_LOW or vigor > _VIGOR_MID_HIGH:
                value = self._max_raster_value

            if math.isnan(value):
                value = 0

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

        return heatmap, self._trend_vigor

    def update_scalar(self, sample: Scalar) -> list:
        return self._update_entity(sample.time, sample.value)

    def update_bar(self, sample: Bar) -> list:
        return self._update_entity(sample.time, self._bar_func(sample))

    def update_quote(self, sample: Quote) -> list:
        return self._update_entity(sample.time, self._quote_func(sample))

    def update_trade(self, sample: Trade) -> list:
        return self._update_entity(sample.time, self._trade_func(sample))

    def _update_entity(self, t: datetime.datetime, sample: float) -> list:
        heatmap, tv = self.update(sample, t)
        return [heatmap, Scalar(t, tv)]
