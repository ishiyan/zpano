"""Ehlers' Corona Swing Position indicator."""

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
    AdaptiveBandPass,
    Corona,
    CoronaSnapshot,
    Raster,
    accept_snapshot,
    clip,
    format_number,
    rolling_append,
    validate_integer_parameter,
)
from ..corona.params import CoronaParams
from .params import CoronaSwingPositionParams, default_params, swing_flags


class CoronaSwingPositionSnapshot(NamedTuple):
    """Full 12-field numerical snapshot for CoronaSwingPosition."""

    heatmap: Heatmap
    swing_position: float
    normalized_position: float
    dominant_cycle: float
    dominant_cycle_median: float
    band_pass: float
    quadrature: float
    cycle_amplitude: float
    lead60: float
    position_span: float
    width: float
    has_cycle_estimate: float


Snapshot = CoronaSwingPositionSnapshot
SwingOutput = CoronaSwingPositionSnapshot


class CoronaSwingPosition(Indicator):
    """Ehlers' Corona Swing Position indicator."""

    def __init__(self, params: Optional[CoronaSwingPositionParams] = None) -> None:
        source = default_params() if params is None else params
        self._raster = Raster(
            source.raster_length,
            source.min_parameter_value,
            source.max_parameter_value,
            source.max_raster_value,
        )
        validate_integer_parameter("lead_window", source.lead_window, 2)
        validate_integer_parameter("position_window", source.position_window, 2)
        self._lead_window: int = source.lead_window
        self._position_window: int = source.position_window

        self._corona = Corona(
            CoronaParams(
                high_pass_filter_cutoff=source.high_pass_filter_cutoff,
                minimal_period=source.minimal_period,
                maximal_period=source.maximal_period,
                decibels_lower_threshold=6.0,
                decibels_upper_threshold=self._raster.ceiling,
            )
        )
        self._corona_parameters = self._corona.parameters
        self._bandpass = AdaptiveBandPass()

        bc = source.bar_component if source.bar_component is not None else DEFAULT_BAR_COMPONENT
        qc = source.quote_component if source.quote_component is not None else DEFAULT_QUOTE_COMPONENT
        tc = source.trade_component if source.trade_component is not None else DEFAULT_TRADE_COMPONENT

        self._bar_func = bar_component_value(bc)
        self._quote_func = quote_component_value(qc)
        self._trade_func = trade_component_value(tc)

        comp = component_triple_mnemonic(bc, qc, tc)
        flags = swing_flags(source)
        self.mnemonic: str = (
            f"cswing({self._raster.length}, {format_number(self._raster.ceiling)}, "
            f"{format_number(self._raster.lower)}, {format_number(self._raster.upper)}, "
            f"{source.high_pass_filter_cutoff}{flags}{comp})"
        )
        self.description: str = "Corona swing position " + self.mnemonic
        self._mnemonic_sp: str = f"cswing-sp({source.high_pass_filter_cutoff}{flags}{comp})"
        self._description_sp: str = "Corona swing position scalar " + self._mnemonic_sp

        self.reset()

    def reset(self) -> None:
        """Reset streaming state while preserving configuration."""
        self._corona.reset()
        self._sample_count: int = 0
        self._bandpass.reset()
        self._raster.reset()
        self._leads: list[float] = [0.0] * self._lead_window
        self._positions: list[float] = [0.0] * self._position_window

    def is_primed(self) -> bool:
        return self._sample_count >= FIRST_BANK_SAMPLE

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

    def _empty(self, time: Optional[datetime.datetime]) -> CoronaSwingPositionSnapshot:
        return CoronaSwingPositionSnapshot(
            self._raster.empty(time),
            math.nan,
            math.nan,
            math.nan,
            math.nan,
            math.nan,
            math.nan,
            math.nan,
            math.nan,
            math.nan,
            math.nan,
            math.nan,
        )

    def update(
        self, sample: float, time: Optional[datetime.datetime] = None
    ) -> CoronaSwingPositionSnapshot:
        """Advance standalone engine with one price sample and return full snapshot."""
        return self.update_from_snapshot(self._corona.update(sample, time))

    def update_from_snapshot(self, snapshot: CoronaSnapshot) -> CoronaSwingPositionSnapshot:
        """Consume one shared CoronaSnapshot in accepted-sample order."""
        if not accept_snapshot(snapshot, self._corona_parameters, self._sample_count + 1):
            return self._empty(snapshot.time)
        self._sample_count = snapshot.sample_count

        cycle = self._bandpass.update(snapshot.price, snapshot.dominant_cycle_median)
        lowest, highest = rolling_append(self._leads, cycle.lead60, self._lead_window)
        flat = highest == lowest
        position = 0.5 if flat else (cycle.lead60 - lowest) / (highest - lowest)
        position = clip(position, 0.0, 1.0)

        low_position, high_position = rolling_append(
            self._positions, position, self._position_window
        )
        span = high_position - low_position
        width = 0.01 if span > 0.85 else 0.15 * span
        if width == 0.0:
            width = 0.01

        self._raster.update(position, width, "swing", suppress=span > 0.8 or flat)
        if not snapshot.primed:
            return self._empty(snapshot.time)

        mapped = self._raster.lower + (self._raster.upper - self._raster.lower) * position
        return CoronaSwingPositionSnapshot(
            heatmap=self._raster.column(snapshot.time),
            swing_position=mapped,
            normalized_position=position,
            dominant_cycle=snapshot.dominant_cycle,
            dominant_cycle_median=snapshot.dominant_cycle_median,
            band_pass=cycle.band_pass,
            quadrature=cycle.quadrature,
            cycle_amplitude=cycle.amplitude,
            lead60=cycle.lead60,
            position_span=span,
            width=width,
            has_cycle_estimate=float(snapshot.has_cycle_estimate),
        )

    def _wrap(self, time: datetime.datetime, result: CoronaSwingPositionSnapshot) -> Output:
        return [result.heatmap, Scalar(time, result.swing_position)]

    def update_scalar(self, sample: Scalar) -> Output:
        return self._wrap(sample.time, self.update(sample.value, sample.time))

    def update_bar(self, sample: Bar) -> Output:
        return self._wrap(sample.time, self.update(self._bar_func(sample), sample.time))

    def update_quote(self, sample: Quote) -> Output:
        return self._wrap(sample.time, self.update(self._quote_func(sample), sample.time))

    def update_trade(self, sample: Trade) -> Output:
        return self._wrap(sample.time, self.update(self._trade_func(sample), sample.time))
