"""Ehlers' Corona Trend Vigor indicator."""

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
    AdaptiveBandPass,
    Corona,
    CoronaSnapshot,
    Raster,
    accept_snapshot,
    checked,
    clip,
    format_number,
)
from ..corona.params import CoronaParams
from .params import CoronaTrendVigorParams, default_params, trend_flags


class CoronaTrendVigorSnapshot(NamedTuple):
    """Full 13-field numerical snapshot for CoronaTrendVigor."""

    heatmap: Heatmap
    trend_vigor: float
    normalized_vigor: float
    dominant_cycle: float
    dominant_cycle_median: float
    band_pass: float
    quadrature: float
    cycle_amplitude: float
    trend_displacement: float
    raw_ratio: float
    lookback: float
    width: float
    has_cycle_estimate: float


Snapshot = CoronaTrendVigorSnapshot
TrendOutput = CoronaTrendVigorSnapshot


class CoronaTrendVigor(Indicator):
    """Ehlers' Corona Trend Vigor indicator."""

    def __init__(self, params: Optional[CoronaTrendVigorParams] = None) -> None:
        source = default_params() if params is None else params
        self._raster = Raster(
            source.raster_length,
            source.min_parameter_value,
            source.max_parameter_value,
            source.max_raster_value,
        )
        if (
            isinstance(source.lookback_offset, bool)
            or not isinstance(source.lookback_offset, int)
            or source.lookback_offset not in (0, -1)
        ):
            raise ValueError("lookback_offset must be integer 0 or -1")
        self._lookback_offset: int = source.lookback_offset

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
        flags = trend_flags(source)
        self.mnemonic: str = (
            f"ctv({self._raster.length}, {format_number(self._raster.ceiling)}, "
            f"{format_number(self._raster.lower)}, {format_number(self._raster.upper)}, "
            f"{source.high_pass_filter_cutoff}{flags}{comp})"
        )
        self.description: str = "Corona trend vigor " + self.mnemonic
        self._mnemonic_tv: str = f"ctv-tv({source.high_pass_filter_cutoff}{flags}{comp})"
        self._description_tv: str = "Corona trend vigor scalar " + self._mnemonic_tv

        self.reset()

    def reset(self) -> None:
        """Reset streaming state while preserving configuration."""
        self._corona.reset()
        self._sample_count: int = 0
        self._bandpass.reset()
        self._raster.reset()
        self._prices: list[float] = []
        self._ratio: float = 0.0
        self._ready: bool = False

    def is_primed(self) -> bool:
        return self._ready

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

    def _empty(self, time: Optional[datetime.datetime]) -> CoronaTrendVigorSnapshot:
        return CoronaTrendVigorSnapshot(
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
            math.nan,
        )

    def update(
        self, sample: float, time: Optional[datetime.datetime] = None
    ) -> CoronaTrendVigorSnapshot:
        """Advance standalone engine with one price sample and return full snapshot."""
        return self.update_from_snapshot(self._corona.update(sample, time))

    def update_from_snapshot(self, snapshot: CoronaSnapshot) -> CoronaTrendVigorSnapshot:
        """Consume one shared CoronaSnapshot in accepted-sample order."""
        if not accept_snapshot(snapshot, self._corona_parameters, self._sample_count + 1):
            return self._empty(snapshot.time)
        self._sample_count = snapshot.sample_count

        cycle = self._bandpass.update(snapshot.price, snapshot.dominant_cycle_median)
        self._prices.append(snapshot.price)
        if len(self._prices) > self._corona_parameters.maximal_period + 1:
            del self._prices[0]

        lag = max(1, int(math.floor(snapshot.dominant_cycle_median)) + self._lookback_offset)
        has_history = len(self._prices) > lag
        trend = snapshot.price - self._prices[-1 - lag] if has_history else math.nan
        raw_ratio = (
            trend / cycle.amplitude if has_history and cycle.amplitude > 0.0 else math.nan
        )
        if has_history and trend != 0.0 and cycle.amplitude > 0.0:
            self._ratio = clip(checked(0.33 * raw_ratio + 0.67 * self._ratio), -10.0, 10.0)

        vigor = 0.05 * (self._ratio + 10.0)
        width = min(vigor - 0.3, 0.7 - vigor) if 0.3 <= vigor <= 0.7 else 0.01
        if width == 0.0:
            width = 0.01

        self._raster.update(vigor, width, "trend", suppress=vigor < 0.3 or vigor > 0.7)
        self._ready = snapshot.primed and has_history
        if not self._ready:
            return self._empty(snapshot.time)

        mapped = self._raster.lower + (self._raster.upper - self._raster.lower) * vigor
        return CoronaTrendVigorSnapshot(
            heatmap=self._raster.column(snapshot.time),
            trend_vigor=mapped,
            normalized_vigor=vigor,
            dominant_cycle=snapshot.dominant_cycle,
            dominant_cycle_median=snapshot.dominant_cycle_median,
            band_pass=cycle.band_pass,
            quadrature=cycle.quadrature,
            cycle_amplitude=cycle.amplitude,
            trend_displacement=trend,
            raw_ratio=raw_ratio,
            lookback=float(lag),
            width=width,
            has_cycle_estimate=float(snapshot.has_cycle_estimate),
        )

    def _wrap(self, time: datetime.datetime, result: CoronaTrendVigorSnapshot) -> Output:
        return [result.heatmap, Scalar(time, result.trend_vigor)]

    def update_scalar(self, sample: Scalar) -> Output:
        return self._wrap(sample.time, self.update(sample.value, sample.time))

    def update_bar(self, sample: Bar) -> Output:
        return self._wrap(sample.time, self.update(self._bar_func(sample), sample.time))

    def update_quote(self, sample: Quote) -> Output:
        return self._wrap(sample.time, self.update(self._quote_func(sample), sample.time))

    def update_trade(self, sample: Trade) -> Output:
        return self._wrap(sample.time, self.update(self._trade_func(sample), sample.time))
