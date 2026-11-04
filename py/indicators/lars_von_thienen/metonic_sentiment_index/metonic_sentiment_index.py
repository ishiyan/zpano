# Numerical source: Lars von Thienen / WhenToTrade, 2017.
# https://www.whentotrade.com/ ; CC BY 4.0.
# Research adaptations, startup and reconstruction rules: description.md.
from __future__ import annotations

import math
from collections import deque
from datetime import date, datetime, timedelta
from typing import NamedTuple
from ..metonic.calendar import calendar_day
from .engine import Engine, SentimentOutput
from .params import MetonicSentimentIndexParams, default_params
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


class MetonicSentimentIndex(Indicator):
    """Lars von Thienen's metonic sentiment index research implementation."""

    def __init__(self, params: MetonicSentimentIndexParams | None = None) -> None:
        p = default_params() if params is None else params
        self._engine = Engine(length=p.length, delay1=p.delay1, delay2=p.delay2, delay3=p.delay3, mode=p.mode)
        suffix = ""
        parameters = ",".join(
            str(int(value)) if isinstance(value, bool) else
            "_".join(format(item, ".12g") for item in value) if isinstance(value, (tuple, list)) else
            format(value, ".12g")
            for value in (p.length, p.delay1, p.delay2, p.delay3, p.mode,))
        self._mnemonic = "msent(" + parameters + suffix + ")"
        self._description = "Metonic sentiment index " + self._mnemonic

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
        return build_metadata(Identifier.METONIC_SENTIMENT_INDEX, self._mnemonic,
            self._description, [OutputText(self._mnemonic + " " + field,
            self._description + " " + field) for field in ['value', 'daily_sentiment', 'smoothed_close', 'smoothed_open']])

    def update(self, close: float, open_price: float | None = None, time: date | None = None):
        """Accept one observation and return its numerical snapshot."""
        return self._engine.update(close, open_price=open_price, time=time)

    def _wrap(self, time: datetime, result: SentimentOutput) -> Output:
        output = [Scalar(time, value) for value in result]
        return output

    def update_bar(self, sample: Bar) -> Output:
        return self._wrap(sample.time, self.update(sample.close, open_price=sample.open, time=calendar_day(sample.time)))

    def update_scalar(self, sample: Scalar) -> Output:
        return self._wrap(sample.time, self.update(sample.value, time=calendar_day(sample.time)))

    def update_quote(self, sample: Quote) -> Output:
        return self._wrap(sample.time, self.update(sample.mid(), time=calendar_day(sample.time)))

    def update_trade(self, sample: Trade) -> Output:
        return self._wrap(sample.time, self.update(sample.price, time=calendar_day(sample.time)))
