# SPDX-License-Identifier: MPL-2.0
# Source: Detrended Rhythm Oscillator, © WhenToTrade / Lars von Thienen.
# Numerical source: Lars von Thienen / WhenToTrade, 2017.
# https://www.whentotrade.com/ ; MPL-2.0.
# Research adaptations, startup and reconstruction rules: description.md.
from __future__ import annotations

import math
from collections import deque
from datetime import date, datetime, timedelta
from typing import NamedTuple
from dataclasses import dataclass
from typing import Optional
from ....entities.bar_component import BarComponent
from ....entities.quote_component import QuoteComponent
from ....entities.trade_component import TradeComponent

@dataclass
class DetrendedRhythmOscillatorsParams:
    """Parameters for detrended-rhythm-oscillators; defaults match the prepared reference."""
    fast_length: int = 3
    """Configured fast length; default 3."""
    slow_length: int = 50
    """Configured slow length; default 50."""
    zig_period: int = 30
    """Configured zig period; default 30."""
    cycle_memory: int = 15
    """Configured cycle memory; default 15."""
    detrend: bool = True
    """Configured detrend; default True."""
    bar_component: Optional[BarComponent] = None
    """The bar component. None uses ClosePrice and omits it from the mnemonic."""
    quote_component: Optional[QuoteComponent] = None
    """The quote component. None uses MidPrice and omits it from the mnemonic."""
    trade_component: Optional[TradeComponent] = None
    """The trade component. None uses Price and omits it from the mnemonic."""


def default_params() -> DetrendedRhythmOscillatorsParams:
    """Return the prepared reference defaults."""
    return DetrendedRhythmOscillatorsParams()
