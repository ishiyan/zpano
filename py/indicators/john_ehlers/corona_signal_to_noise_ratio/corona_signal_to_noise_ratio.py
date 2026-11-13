"""Ehlers' Corona Signal-to-Noise Ratio indicator."""

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
    Raster,
    accept_snapshot,
    clip,
    format_number,
    valid_price,
)
from ..corona.params import CoronaParams
from .params import CoronaSignalToNoiseRatioParams, default_params, snr_flags


class CoronaSignalToNoiseRatioSnapshot(NamedTuple):
    """Full 13-field numerical snapshot for CoronaSignalToNoiseRatio."""

    heatmap: Heatmap
    signal_to_noise_ratio: float
    physical_decibels: float
    calibrated_decibels: float
    normalized_position: float
    signal: float
    noise: float
    median_bar_range: float
    width: float
    dominant_cycle: float
    dominant_cycle_median: float
    has_cycle_estimate: float
    ratio_is_finite: float


Snapshot = CoronaSignalToNoiseRatioSnapshot
SignalToNoiseOutput = CoronaSignalToNoiseRatioSnapshot


def _validate_range_input(
    high: Optional[float],
    low: Optional[float],
    bar_range: Optional[float],
) -> float:
    """Validate and compute nonnegative bar height before mutating state."""
    if bar_range is not None:
        if high is not None or low is not None:
            raise ValueError("supply bar_range OR both high and low")
        if isinstance(bar_range, bool) or not isinstance(bar_range, (int, float)):
            raise ValueError("bar_range must be numeric")
        value = float(bar_range)
    else:
        if high is None or low is None:
            raise ValueError("SNR requires both high and low, or an explicit bar_range")
        if (
            isinstance(high, bool)
            or isinstance(low, bool)
            or not isinstance(high, (int, float))
            or not isinstance(low, (int, float))
        ):
            raise ValueError("high and low must be numeric")
        if math.isnan(high) or math.isnan(low):
            return math.nan
        if not math.isfinite(high) or not math.isfinite(low):
            raise ValueError("high and low must be finite")
        value = float(high) - float(low)
    if math.isnan(value):
        return math.nan
    if not math.isfinite(value) or value < 0.0:
        raise ValueError("bar_range must be finite and nonnegative; require high >= low")
    return value


class CoronaSignalToNoiseRatio(Indicator):
    """Ehlers' Corona Signal-to-Noise Ratio indicator."""

    def __init__(self, params: Optional[CoronaSignalToNoiseRatioParams] = None) -> None:
        source = default_params() if params is None else params
        self._raster = Raster(
            source.raster_length,
            source.min_parameter_value,
            source.max_parameter_value,
            source.max_raster_value,
        )
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

        bc = source.bar_component if source.bar_component is not None else DEFAULT_BAR_COMPONENT
        qc = source.quote_component if source.quote_component is not None else DEFAULT_QUOTE_COMPONENT
        tc = source.trade_component if source.trade_component is not None else DEFAULT_TRADE_COMPONENT

        self._bar_func = bar_component_value(bc)
        self._quote_func = quote_component_value(qc)
        self._trade_func = trade_component_value(tc)

        comp = component_triple_mnemonic(bc, qc, tc)
        flags = snr_flags(source)
        self.mnemonic: str = (
            f"csnr({self._raster.length}, {format_number(self._raster.ceiling)}, "
            f"{format_number(self._raster.lower)}, {format_number(self._raster.upper)}, "
            f"{source.high_pass_filter_cutoff}{flags}{comp})"
        )
        self.description: str = "Corona signal to noise ratio " + self.mnemonic
        self._mnemonic_snr: str = f"csnr-snr({source.high_pass_filter_cutoff}{flags}{comp})"
        self._description_snr: str = (
            "Corona signal to noise ratio scalar " + self._mnemonic_snr
        )

        self.reset()

    def reset(self) -> None:
        """Reset streaming state while preserving configuration."""
        self._corona.reset()
        self._sample_count: int = 0
        self._raster.reset()
        self._ranges: list[float] = [0.0] * 5
        self._average: float = 0.0
        self._signal: float = 0.0
        self._noise: float = 0.0

    def is_primed(self) -> bool:
        return self._sample_count >= FIRST_BANK_SAMPLE

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

    def _empty(self, time: Optional[datetime.datetime]) -> CoronaSignalToNoiseRatioSnapshot:
        return CoronaSignalToNoiseRatioSnapshot(
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
        self,
        sample: float,
        time: Optional[datetime.datetime] = None,
        *,
        high: Optional[float] = None,
        low: Optional[float] = None,
        bar_range: Optional[float] = None,
    ) -> CoronaSignalToNoiseRatioSnapshot:
        """Advance standalone engine with price + bar range (or high & low)."""
        if not valid_price(sample):
            return self._empty(time)
        range_value = _validate_range_input(high, low, bar_range)
        if math.isnan(range_value):
            return self._empty(time)
        return self.update_from_snapshot(
            self._corona.update(sample, time), bar_range=range_value
        )

    def update_from_snapshot(
        self,
        snapshot: CoronaSnapshot,
        *,
        high: Optional[float] = None,
        low: Optional[float] = None,
        bar_range: Optional[float] = None,
    ) -> CoronaSignalToNoiseRatioSnapshot:
        """Consume one shared CoronaSnapshot in accepted-sample order."""
        if not snapshot.accepted:
            # Validate snapshot parameter compatibility even on skipped bars.
            accept_snapshot(snapshot, self._corona_parameters, self._sample_count + 1)
            return self._empty(snapshot.time)
        range_value = _validate_range_input(high, low, bar_range)
        if math.isnan(range_value):
            raise ValueError("accepted shared snapshot requires a finite bar range")
        accept_snapshot(snapshot, self._corona_parameters, self._sample_count + 1)
        self._sample_count = snapshot.sample_count

        self._ranges = self._ranges[1:] + [range_value]
        median_range = sorted(self._ranges)[2]
        if self._sample_count == 1:
            self._average = snapshot.price
        else:
            self._average = 0.1 * snapshot.price + 0.9 * self._average
            if self._average != 0.0 or snapshot.maximum_power > 0.0:
                self._signal = 0.2 * math.sqrt(snapshot.maximum_power) + 0.9 * self._signal
            if self._average != 0.0 and self._sample_count > 5:
                self._noise = 0.1 * median_range + 0.9 * self._noise

        if self._signal > 0.0 and self._noise > 0.0:
            physical = 20.0 * (math.log10(self._signal) - math.log10(self._noise))
        elif self._signal > 0.0:
            physical = math.inf
        elif self._noise > 0.0:
            physical = -math.inf
        else:
            physical = math.nan

        calibrated = physical + 3.5
        position = 0.0 if math.isnan(calibrated) else clip(calibrated, 0.0, 10.0) / 10.0
        width = max(0.0, 0.2 - 0.4 * position)
        self._raster.update(position, width, "snr", suppress=position > 0.5)

        if not snapshot.primed:
            return self._empty(snapshot.time)

        mapped = self._raster.lower + (self._raster.upper - self._raster.lower) * position
        return CoronaSignalToNoiseRatioSnapshot(
            heatmap=self._raster.column(snapshot.time),
            signal_to_noise_ratio=mapped,
            physical_decibels=physical,
            calibrated_decibels=calibrated,
            normalized_position=position,
            signal=self._signal,
            noise=self._noise,
            median_bar_range=median_range,
            width=width,
            dominant_cycle=snapshot.dominant_cycle,
            dominant_cycle_median=snapshot.dominant_cycle_median,
            has_cycle_estimate=float(snapshot.has_cycle_estimate),
            ratio_is_finite=float(math.isfinite(physical)),
        )

    def _wrap(
        self, time: datetime.datetime, result: CoronaSignalToNoiseRatioSnapshot
    ) -> Output:
        return [result.heatmap, Scalar(time, result.signal_to_noise_ratio)]

    def update_scalar(self, sample: Scalar) -> Output:
        value = sample.value
        return self._wrap(
            sample.time, self.update(value, sample.time, low=value, high=value)
        )

    def update_bar(self, sample: Bar) -> Output:
        return self._wrap(
            sample.time,
            self.update(
                self._bar_func(sample),
                sample.time,
                high=sample.high,
                low=sample.low,
            ),
        )

    def update_quote(self, sample: Quote) -> Output:
        value = self._quote_func(sample)
        return self._wrap(
            sample.time, self.update(value, sample.time, low=value, high=value)
        )

    def update_trade(self, sample: Trade) -> Output:
        value = self._trade_func(sample)
        return self._wrap(
            sample.time, self.update(value, sample.time, low=value, high=value)
        )
