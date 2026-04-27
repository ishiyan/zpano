"""Ehlers' Corona Spectrum heatmap indicator."""

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


class CoronaSpectrum:
    """Ehlers' Corona Spectrum heatmap indicator.

    Outputs:
    - Value: per-bar heatmap column (decibels across filter bank)
    - DominantCycle: weighted-center-of-gravity dominant cycle estimate
    - DominantCycleMedian: 5-sample median of DominantCycle
    """

    def __init__(self, params: Params) -> None:
        invalid = "invalid corona spectrum parameters"

        cfg_min_raster = params.min_raster_value if params.min_raster_value != 0 else 6.0
        cfg_max_raster = params.max_raster_value if params.max_raster_value != 0 else 20.0
        cfg_min_param = params.min_parameter_value if params.min_parameter_value != 0 else 6.0
        cfg_max_param = params.max_parameter_value if params.max_parameter_value != 0 else 30.0
        cfg_hp_cutoff = params.high_pass_filter_cutoff if params.high_pass_filter_cutoff != 0 else 30

        if cfg_min_raster < 0:
            raise ValueError(f"{invalid}: MinRasterValue should be >= 0")
        if cfg_max_raster <= cfg_min_raster:
            raise ValueError(f"{invalid}: MaxRasterValue should be > MinRasterValue")

        min_param = math.ceil(cfg_min_param)
        max_param = math.floor(cfg_max_param)

        if min_param < 2:
            raise ValueError(f"{invalid}: MinParameterValue should be >= 2")
        if max_param <= min_param:
            raise ValueError(f"{invalid}: MaxParameterValue should be > MinParameterValue")
        if cfg_hp_cutoff < 2:
            raise ValueError(f"{invalid}: HighPassFilterCutoff should be >= 2")

        bc = params.bar_component if params.bar_component is not None else BarComponent.MEDIAN
        qc = params.quote_component if params.quote_component is not None else DEFAULT_QUOTE_COMPONENT
        tc = params.trade_component if params.trade_component is not None else DEFAULT_TRADE_COMPONENT

        self._bar_func = bar_component_value(bc)
        self._quote_func = quote_component_value(qc)
        self._trade_func = trade_component_value(tc)

        self._corona = Corona(CoronaParams(
            high_pass_filter_cutoff=cfg_hp_cutoff,
            minimal_period=int(min_param),
            maximal_period=int(max_param),
            decibels_lower_threshold=cfg_min_raster,
            decibels_upper_threshold=cfg_max_raster,
        ))

        self._min_parameter_value = float(min_param)
        self._max_parameter_value = float(max_param)
        self._parameter_resolution = float(self._corona.filter_bank_length - 1) / \
            (float(max_param) - float(min_param))
        self._min_raster_value = cfg_min_raster
        self._max_raster_value = cfg_max_raster

        comp_mn = component_triple_mnemonic(bc, qc, tc)
        self.mnemonic = f"cspect({cfg_min_raster:g}, {cfg_max_raster:g}, " \
            f"{min_param:g}, {max_param:g}, {cfg_hp_cutoff}{comp_mn})"
        self.description = "Corona spectrum " + self.mnemonic

        self._mnemonic_dc = f"cspect-dc({cfg_hp_cutoff}{comp_mn})"
        self._description_dc = "Corona spectrum dominant cycle " + self._mnemonic_dc
        self._mnemonic_dcm = f"cspect-dcm({cfg_hp_cutoff}{comp_mn})"
        self._description_dcm = "Corona spectrum dominant cycle median " + self._mnemonic_dcm

    def is_primed(self) -> bool:
        return self._corona.is_primed()

    def metadata(self) -> Metadata:
        return build_metadata(
            Identifier.CORONA_SPECTRUM,
            self.mnemonic,
            self.description,
            [
                OutputText(mnemonic=self.mnemonic, description=self.description),
                OutputText(mnemonic=self._mnemonic_dc, description=self._description_dc),
                OutputText(mnemonic=self._mnemonic_dcm, description=self._description_dcm),
            ],
        )

    def update(self, sample: float, t: datetime.datetime) -> tuple:
        """Returns (heatmap, dominant_cycle, dominant_cycle_median)."""
        if math.isnan(sample):
            return (Heatmap.empty(t, self._min_parameter_value,
                                  self._max_parameter_value, self._parameter_resolution),
                    float('nan'), float('nan'))

        primed = self._corona.update(sample)
        if not primed:
            return (Heatmap.empty(t, self._min_parameter_value,
                                  self._max_parameter_value, self._parameter_resolution),
                    float('nan'), float('nan'))

        bank = self._corona.filter_bank
        values = [0.0] * len(bank)
        value_min = math.inf
        value_max = -math.inf

        for i, f in enumerate(bank):
            v = f.decibels
            values[i] = v
            if v < value_min:
                value_min = v
            if v > value_max:
                value_max = v

        heatmap = Heatmap(t, self._min_parameter_value, self._max_parameter_value,
                          self._parameter_resolution, value_min, value_max, values)

        return heatmap, self._corona.dominant_cycle, self._corona.dominant_cycle_median

    def update_scalar(self, sample: Scalar) -> list:
        return self._update_entity(sample.time, sample.value)

    def update_bar(self, sample: Bar) -> list:
        return self._update_entity(sample.time, self._bar_func(sample))

    def update_quote(self, sample: Quote) -> list:
        return self._update_entity(sample.time, self._quote_func(sample))

    def update_trade(self, sample: Trade) -> list:
        return self._update_entity(sample.time, self._trade_func(sample))

    def _update_entity(self, t: datetime.datetime, sample: float) -> list:
        heatmap, dc, dcm = self.update(sample, t)
        return [heatmap, Scalar(t, dc), Scalar(t, dcm)]
