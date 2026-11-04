# Numerical source: Lars von Thienen / WhenToTrade, 2017.
# https://www.whentotrade.com/ ; CC BY 4.0.
# Research adaptations, startup and reconstruction rules: description.md.
from __future__ import annotations

import math
from collections import deque
from datetime import date, datetime, timedelta
from typing import NamedTuple
from ..metonic.calendar import calendar_day
from .engine import Engine, MetonicWilliamsOutput
from .params import MetonicWilliamsPercentRParams, default_params
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


class MetonicWilliamsPercentR(Indicator):
    """Lars von Thienen's metonic williams percent r research implementation."""

    def __init__(self, params: MetonicWilliamsPercentRParams | None = None) -> None:
        p = default_params() if params is None else params
        self._engine = Engine(length=p.length, delay1=p.delay1, delay2=p.delay2, delay3=p.delay3)
        suffix = ""
        parameters = ",".join(
            str(int(value)) if isinstance(value, bool) else
            "_".join(format(item, ".12g") for item in value) if isinstance(value, (tuple, list)) else
            format(value, ".12g")
            for value in (p.length, p.delay1, p.delay2, p.delay3,))
        self._mnemonic = "mwillr(" + parameters + suffix + ")"
        self._description = "Metonic williams percent r " + self._mnemonic

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
        return build_metadata(Identifier.METONIC_WILLIAMS_PERCENT_R, self._mnemonic,
            self._description, [OutputText(self._mnemonic + " " + field,
            self._description + " " + field) for field in ['value', 'linear_percent_r']])

    def update(self, close: float, high: float | None = None, low: float | None = None, time: date | None = None):
        """Accept one observation and return its numerical snapshot."""
        return self._engine.update(close, high=high, low=low, time=time)

    def _wrap(self, time: datetime, result: MetonicWilliamsOutput) -> Output:
        output = [Scalar(time, value) for value in result]
        return output

    def update_bar(self, sample: Bar) -> Output:
        return self._wrap(sample.time, self.update(sample.close, high=sample.high, low=sample.low, time=calendar_day(sample.time)))

    def update_scalar(self, sample: Scalar) -> Output:
        raise ValueError("conventional Metonic Williams %R requires a completed HLC bar")

    def update_quote(self, sample: Quote) -> Output:
        raise ValueError("conventional Metonic Williams %R requires a completed HLC bar")

    def update_trade(self, sample: Trade) -> Output:
        raise ValueError("conventional Metonic Williams %R requires a completed HLC bar")
