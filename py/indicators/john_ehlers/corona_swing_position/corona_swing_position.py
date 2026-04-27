"""Ehlers' Corona Swing Position heatmap indicator."""

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

_MAX_LEAD_LIST_COUNT = 50
_MAX_POSITION_LIST_COUNT = 20
_LEAD60_COEF_BP = 0.5
_LEAD60_COEF_Q = 0.866
_BP_DELTA = 0.1
_WIDTH_HIGH_THRESHOLD = 0.85
_WIDTH_HIGH_SATURATE = 0.8
_WIDTH_NARROW = 0.01
_WIDTH_SCALE = 0.15
_RASTER_BLEND_EXPONENT = 0.95
_RASTER_BLEND_HALF = 0.5


def _append_rolling(lst: list[float], max_count: int, v: float) -> tuple[float, float]:
    """Append v, drop oldest if at max_count, return (lowest, highest)."""
    if len(lst) >= max_count:
        del lst[0]
    lst.append(v)
    lowest = min(lst)
    highest = max(lst)
    return lowest, highest


class CoronaSwingPosition:
    """Ehlers' Corona Swing Position heatmap indicator."""

    def __init__(self, params: Params) -> None:
        invalid = "invalid corona swing position parameters"

        cfg_raster_len = params.raster_length if params.raster_length != 0 else 50
        cfg_max_raster = params.max_raster_value if params.max_raster_value != 0 else 20.0

        # Both must be zero to trigger defaults (since 0 is valid for either individually).
        if params.min_parameter_value == 0 and params.max_parameter_value == 0:
            cfg_min_param = -5.0
            cfg_max_param = 5.0
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

        self._lead_list: list[float] = []
        self._position_list: list[float] = []
        self._sample_previous = 0.0
        self._sample_previous2 = 0.0
        self._band_pass_previous = 0.0
        self._band_pass_previous2 = 0.0
        self._swing_position = float('nan')
        self._is_started = False

        comp_mn = component_triple_mnemonic(bc, qc, tc)
        self.mnemonic = f"cswing({cfg_raster_len}, {cfg_max_raster:g}, " \
            f"{cfg_min_param:g}, {cfg_max_param:g}, {cfg_hp_cutoff}{comp_mn})"
        self.description = "Corona swing position " + self.mnemonic
        self._mnemonic_sp = f"cswing-sp({cfg_hp_cutoff}{comp_mn})"
        self._description_sp = "Corona swing position scalar " + self._mnemonic_sp

    def is_primed(self) -> bool:
        return self._corona.is_primed()

    def metadata(self) -> Metadata:
        return build_metadata(
            Identifier.CORONA_SWING_POSITION,
            self.mnemonic,
            self.description,
            [
                OutputText(mnemonic=self.mnemonic, description=self.description),
                OutputText(mnemonic=self._mnemonic_sp, description=self._description_sp),
            ],
        )

    def update(self, sample: float, t: datetime.datetime) -> tuple:
        """Returns (heatmap, swing_position)."""
        if math.isnan(sample):
            return (Heatmap.empty(t, self._min_parameter_value,
                                  self._max_parameter_value, self._parameter_resolution),
                    float('nan'))

        primed = self._corona.update(sample)

        if not self._is_started:
            self._sample_previous = sample
            self._is_started = True
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

        # Quadrature = derivative / omega.  Go produces ±Inf on /0; mimic that.
        if omega == 0:
            diff = band_pass - self._band_pass_previous
            if diff == 0:
                quadrature2 = 0.0
            else:
                quadrature2 = math.copysign(math.inf, diff)
        else:
            quadrature2 = (band_pass - self._band_pass_previous) / omega

        self._band_pass_previous2 = self._band_pass_previous
        self._band_pass_previous = band_pass
        self._sample_previous2 = self._sample_previous
        self._sample_previous = sample

        # 60° lead: lead60 = 0.5·BP_previous2 + 0.866·Q
        lead60 = _LEAD60_COEF_BP * self._band_pass_previous2 + _LEAD60_COEF_Q * quadrature2

        lowest, highest = _append_rolling(self._lead_list, _MAX_LEAD_LIST_COUNT, lead60)

        # Normalised lead position in [0, 1].
        position = highest - lowest
        if position > 0:
            position = (lead60 - lowest) / position

        lowest, highest = _append_rolling(self._position_list, _MAX_POSITION_LIST_COUNT, position)
        highest -= lowest

        width = _WIDTH_SCALE * highest
        if highest > _WIDTH_HIGH_THRESHOLD:
            width = _WIDTH_NARROW

        self._swing_position = (self._max_parameter_value - self._min_parameter_value) * \
            position + self._min_parameter_value

        position_scaled_to_raster_length = round(position * self._raster_length)
        position_scaled_to_max_raster_value = position * self._max_raster_value

        for i in range(self._raster_length):
            value = self._raster[i]

            if i == position_scaled_to_raster_length:
                value *= _RASTER_BLEND_HALF
            else:
                argument = position_scaled_to_max_raster_value - self._raster_step * i
                if i > position_scaled_to_raster_length:
                    argument = -argument

                if width > 0:
                    value = _RASTER_BLEND_HALF * \
                        (math.pow(argument / width, _RASTER_BLEND_EXPONENT) + _RASTER_BLEND_HALF * value)

            if value < 0:
                value = 0
            elif value > self._max_raster_value:
                value = self._max_raster_value

            if highest > _WIDTH_HIGH_SATURATE:
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

        return heatmap, self._swing_position

    def update_scalar(self, sample: Scalar) -> list:
        return self._update_entity(sample.time, sample.value)

    def update_bar(self, sample: Bar) -> list:
        return self._update_entity(sample.time, self._bar_func(sample))

    def update_quote(self, sample: Quote) -> list:
        return self._update_entity(sample.time, self._quote_func(sample))

    def update_trade(self, sample: Trade) -> list:
        return self._update_entity(sample.time, self._trade_func(sample))

    def _update_entity(self, t: datetime.datetime, sample: float) -> list:
        heatmap, sp = self.update(sample, t)
        return [heatmap, Scalar(t, sp)]
