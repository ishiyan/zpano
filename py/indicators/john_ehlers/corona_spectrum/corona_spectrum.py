"""Ehlers' Corona Spectrum heatmap indicator."""

from __future__ import annotations

import datetime
import math
from typing import NamedTuple, Optional

from ....entities.bar import Bar
from ....entities.bar_component import DEFAULT_BAR_COMPONENT, bar_component_value
from ....entities.quote import Quote
from ....entities.quote_component import DEFAULT_QUOTE_COMPONENT, quote_component_value
from ....entities.scalar import Scalar
from ....entities.trade import Trade
from ....entities.trade_component import DEFAULT_TRADE_COMPONENT, trade_component_value
from ...core.build_metadata import OutputText, build_metadata
from ...core.component_triple_mnemonic import component_triple_mnemonic
from ...core.identifier import Identifier
from ...core.indicator import Indicator
from ...core.metadata import Metadata
from ...core.output import Output
from ...core.outputs.heatmap import Heatmap
from ..corona.corona import (
    FIRST_BANK_SAMPLE,
    Corona,
    CoronaSnapshot,
    accept_snapshot,
    format_number,
    validate_finite_parameter,
    validate_integer_parameter,
)
from ..corona.params import CoronaParams
from .params import CoronaSpectrumParams, default_params


class CoronaSpectrumSnapshot(NamedTuple):
    """Full 9-field numerical snapshot for CoronaSpectrum."""

    heatmap: Heatmap
    dominant_cycle: float
    dominant_cycle_median: float
    maximum_power: float
    normalization_power: float
    selected_bins: float
    has_cycle_estimate: float
    power: tuple[float, ...]
    relative_power: tuple[float, ...]


Snapshot = CoronaSpectrumSnapshot
SpectrumOutput = CoronaSpectrumSnapshot


class CoronaSpectrum(Indicator):
    """Ehlers' Corona Spectrum heatmap indicator."""

    def __init__(self, params: Optional[CoronaSpectrumParams] = None) -> None:
        source = default_params() if params is None else params
        min_raster = validate_finite_parameter("min_raster_value", source.min_raster_value)
        max_raster = validate_finite_parameter("max_raster_value", source.max_raster_value)
        min_param_raw = validate_finite_parameter("min_parameter_value", source.min_parameter_value)
        max_param_raw = validate_finite_parameter("max_parameter_value", source.max_parameter_value)
        validate_integer_parameter("high_pass_filter_cutoff", source.high_pass_filter_cutoff, 5)

        min_period = int(math.ceil(min_param_raw))
        max_period = int(math.floor(max_param_raw))
        if min_period < 6:
            raise ValueError("min_parameter_value must ceil to an integer >= 6")
        if max_period <= min_period:
            raise ValueError("max_parameter_value floor must be > min_parameter_value ceil")
        if not 0.0 <= min_raster < max_raster or max_raster <= 0.0:
            raise ValueError("require 0 <= min_raster_value < max_raster_value")

        bc = source.bar_component if source.bar_component is not None else DEFAULT_BAR_COMPONENT
        qc = source.quote_component if source.quote_component is not None else DEFAULT_QUOTE_COMPONENT
        tc = source.trade_component if source.trade_component is not None else DEFAULT_TRADE_COMPONENT

        self._bar_func = bar_component_value(bc)
        self._quote_func = quote_component_value(qc)
        self._trade_func = trade_component_value(tc)

        self._corona = Corona(
            CoronaParams(
                high_pass_filter_cutoff=source.high_pass_filter_cutoff,
                minimal_period=min_period,
                maximal_period=max_period,
                decibels_lower_threshold=min_raster,
                decibels_upper_threshold=max_raster,
            )
        )
        self._corona_parameters = self._corona.parameters
        self._sample_count: int = 0

        self._min_parameter_value: float = float(min_period)
        self._max_parameter_value: float = float(max_period)
        self._parameter_resolution: float = 2.0

        comp = component_triple_mnemonic(bc, qc, tc)
        self.mnemonic: str = (
            f"cspect({format_number(min_raster)}, {format_number(max_raster)}, "
            f"{min_period}, {max_period}, {source.high_pass_filter_cutoff}{comp})"
        )
        self.description: str = "Corona spectrum " + self.mnemonic
        self._mnemonic_dc: str = f"cspect-dc({source.high_pass_filter_cutoff}{comp})"
        self._description_dc: str = "Corona spectrum dominant cycle " + self._mnemonic_dc
        self._mnemonic_dcm: str = f"cspect-dcm({source.high_pass_filter_cutoff}{comp})"
        self._description_dcm: str = "Corona spectrum dominant cycle median " + self._mnemonic_dcm

    def reset(self) -> None:
        """Reset streaming state while preserving configuration."""
        self._corona.reset()
        self._sample_count = 0

    def is_primed(self) -> bool:
        return self._sample_count >= FIRST_BANK_SAMPLE

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

    def _empty(self, time: Optional[datetime.datetime]) -> CoronaSpectrumSnapshot:
        cells = (math.nan,) * self._corona.filter_bank_length
        heatmap = Heatmap.empty(
            time,
            self._min_parameter_value,
            self._max_parameter_value,
            self._parameter_resolution,
        )
        return CoronaSpectrumSnapshot(
            heatmap,
            math.nan,
            math.nan,
            math.nan,
            math.nan,
            math.nan,
            math.nan,
            cells,
            cells,
        )

    def update(
        self, sample: float, time: Optional[datetime.datetime] = None
    ) -> CoronaSpectrumSnapshot:
        """Advance standalone engine with one price sample and return full snapshot."""
        return self.update_from_snapshot(self._corona.update(sample, time))

    def update_from_snapshot(self, snapshot: CoronaSnapshot) -> CoronaSpectrumSnapshot:
        """Consume one shared CoronaSnapshot in accepted-sample order."""
        if not accept_snapshot(snapshot, self._corona_parameters, self._sample_count + 1):
            return self._empty(snapshot.time)
        self._sample_count = snapshot.sample_count
        if not snapshot.primed:
            return self._empty(snapshot.time)
        values = list(snapshot.decibels)
        heatmap = Heatmap(
            snapshot.time,
            self._min_parameter_value,
            self._max_parameter_value,
            self._parameter_resolution,
            min(values),
            max(values),
            values,
        )
        return CoronaSpectrumSnapshot(
            heatmap=heatmap,
            dominant_cycle=snapshot.dominant_cycle,
            dominant_cycle_median=snapshot.dominant_cycle_median,
            maximum_power=snapshot.maximum_power,
            normalization_power=snapshot.normalization_power,
            selected_bins=float(snapshot.selected_bins),
            has_cycle_estimate=float(snapshot.has_cycle_estimate),
            power=snapshot.power,
            relative_power=snapshot.relative_power,
        )

    def _wrap(self, time: datetime.datetime, result: CoronaSpectrumSnapshot) -> Output:
        return [
            result.heatmap,
            Scalar(time, result.dominant_cycle),
            Scalar(time, result.dominant_cycle_median),
        ]

    def update_scalar(self, sample: Scalar) -> Output:
        return self._wrap(sample.time, self.update(sample.value, sample.time))

    def update_bar(self, sample: Bar) -> Output:
        return self._wrap(sample.time, self.update(self._bar_func(sample), sample.time))

    def update_quote(self, sample: Quote) -> Output:
        return self._wrap(sample.time, self.update(self._quote_func(sample), sample.time))

    def update_trade(self, sample: Trade) -> Output:
        return self._wrap(sample.time, self.update(self._trade_func(sample), sample.time))
