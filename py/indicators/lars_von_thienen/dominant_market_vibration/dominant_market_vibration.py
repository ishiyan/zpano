# Numerical source: Lars von Thienen / WhenToTrade, 2017.
# https://www.whentotrade.com/ ; CC BY 4.0.
# Research adaptations, startup and reconstruction rules: description.md.
from __future__ import annotations

import math
from collections import deque
from datetime import date, datetime, timedelta
from typing import NamedTuple
from ..cycle_scanner.scanner import RankedCycle
from .engine import Engine
from .params import DominantMarketVibrationParams, default_params
from ...core.indicator import Indicator
from ...core.metadata import Metadata
from ...core.output import Output
from ...core.identifier import Identifier
from ...core.build_metadata import build_metadata, OutputText
from ...core.component_triple_mnemonic import component_triple_mnemonic
from ...core.outputs.heatmap import Heatmap
from ...core.outputs.band import Band
from ...core.outputs.polyline import Polyline, Point
from ....entities.scalar import Scalar
from ....entities.bar import Bar
from ....entities.quote import Quote
from ....entities.trade import Trade
from ....entities.bar_component import DEFAULT_BAR_COMPONENT, bar_component_value
from ....entities.quote_component import DEFAULT_QUOTE_COMPONENT, quote_component_value
from ....entities.trade_component import DEFAULT_TRADE_COMPONENT, trade_component_value


class DominantMarketVibration(Indicator):
    """Lars von Thienen's dominant market vibration research implementation."""

    def __init__(self, params: DominantMarketVibrationParams | None = None) -> None:
        p = default_params() if params is None else params
        self._engine = Engine(window=p.window, min_period=p.min_period, max_period=p.max_period, smoothing=p.smoothing, genuine_threshold=p.genuine_threshold, phase_cycles=p.phase_cycles, top_k=p.top_k, peak_distance=p.peak_distance, max_segments=p.max_segments)
        bc = p.bar_component if p.bar_component is not None else DEFAULT_BAR_COMPONENT
        qc = p.quote_component if p.quote_component is not None else DEFAULT_QUOTE_COMPONENT
        tc = p.trade_component if p.trade_component is not None else DEFAULT_TRADE_COMPONENT
        self._bar_func = bar_component_value(bc)
        self._quote_func = quote_component_value(qc)
        self._trade_func = trade_component_value(tc)
        suffix = component_triple_mnemonic(bc, qc, tc)
        parameters = ",".join(
            str(int(value)) if isinstance(value, bool) else
            "_".join(format(item, ".12g") for item in value) if isinstance(value, (tuple, list)) else
            format(value, ".12g")
            for value in (p.window, p.min_period, p.max_period, p.smoothing, p.genuine_threshold, p.phase_cycles, p.top_k, p.peak_distance, p.max_segments,))
        self._mnemonic = "dmv(" + parameters + suffix + ")"
        self._description = "Dominant market vibration " + self._mnemonic

    def is_primed(self) -> bool:
        """Return current numerical readiness, following the reference contract."""
        return self._engine.is_primed

    def reset(self) -> None:
        """Clear numerical history while preserving configuration."""
        self._engine.reset()

    @property
    def warmup_bars(self) -> int:
        """Number of initial accepted observations before base readiness."""
        return self._engine.warmup_bars

    def metadata(self) -> Metadata:
        """Describe the registered plot outputs."""
        return build_metadata(Identifier.DOMINANT_MARKET_VIBRATION, self._mnemonic,
            self._description, [OutputText(self._mnemonic + " " + field,
            self._description + " " + field) for field in ['period', 'amplitude', 'phase', 'strength', 'genuine_percent', 'value', 'slope', 'eta_top', 'eta_bottom']])

    def update(self, sample: float) -> RankedCycle:
        """Accept one observation and return its numerical snapshot."""
        return self._engine.update(sample)

    def _wrap(self, time: datetime, result: RankedCycle) -> Output:
        return [Scalar(time, result.period), Scalar(time, result.amplitude), Scalar(time, result.phase), Scalar(time, result.strength), Scalar(time, result.genuine_percent), Scalar(time, result.value), Scalar(time, result.slope), Scalar(time, result.eta_top), Scalar(time, result.eta_bottom)]

    def update_bar(self, sample: Bar) -> Output:
        return self._wrap(sample.time, self.update(self._bar_func(sample)))

    def update_scalar(self, sample: Scalar) -> Output:
        return self._wrap(sample.time, self.update(sample.value))

    def update_quote(self, sample: Quote) -> Output:
        return self._wrap(sample.time, self.update(self._quote_func(sample)))

    def update_trade(self, sample: Trade) -> Output:
        return self._wrap(sample.time, self.update(self._trade_func(sample)))
