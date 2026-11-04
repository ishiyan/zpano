# Numerical source: Lars von Thienen / WhenToTrade, 2017.
# https://www.whentotrade.com/ ; CC BY 4.0.
# Research adaptations, startup and reconstruction rules: description.md.
from __future__ import annotations

import math
from collections import deque
from datetime import date, datetime, timedelta
from typing import NamedTuple
from .engine import Engine, CycleSwingOutput
from .params import CycleSwingMomentumParams, default_params
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


class CycleSwingMomentum(Indicator):
    """Lars von Thienen's cycle swing momentum research implementation."""

    def __init__(self, params: CycleSwingMomentumParams | None = None) -> None:
        p = default_params() if params is None else params
        self._engine = Engine(window=p.window, fast_smoothing=p.fast_smoothing, slow_smoothing=p.slow_smoothing, band_memory=p.band_memory, leveling=p.leveling)
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
            for value in (p.window, p.fast_smoothing, p.slow_smoothing, p.band_memory, p.leveling,))
        self._mnemonic = "csmo(" + parameters + suffix + ")"
        self._description = "Cycle swing momentum " + self._mnemonic

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
        return build_metadata(Identifier.CYCLE_SWING_MOMENTUM, self._mnemonic,
            self._description, [OutputText(self._mnemonic + " " + field,
            self._description + " " + field) for field in ['value', 'lower', 'upper', 'score', 'band']])

    def update(self, sample: float) -> CycleSwingOutput:
        """Accept one observation and return its numerical snapshot."""
        return self._engine.update(sample)

    def _wrap(self, time: datetime, result: CycleSwingOutput) -> Output:
        output = [Scalar(time, value) for value in result]
        output.append(Band(time, result.lower, result.upper))
        return output

    def update_bar(self, sample: Bar) -> Output:
        return self._wrap(sample.time, self.update(self._bar_func(sample)))

    def update_scalar(self, sample: Scalar) -> Output:
        return self._wrap(sample.time, self.update(sample.value))

    def update_quote(self, sample: Quote) -> Output:
        return self._wrap(sample.time, self.update(self._quote_func(sample)))

    def update_trade(self, sample: Trade) -> Output:
        return self._wrap(sample.time, self.update(self._trade_func(sample)))
