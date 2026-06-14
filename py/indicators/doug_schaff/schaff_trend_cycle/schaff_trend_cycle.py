"""Schaff Trend Cycle (STC) indicator -- Doug Schaff.

STC runs a MACD line through two cascaded stochastics, each followed by an
EMA-style smoothing, producing a cyclical oscillator bounded to [0, 100].

This implementation is byte-for-byte concordant with the ProRealCode
``schaff-trend-cycle2`` reference (F. Malagrida, 2017). See the indicator's
``description.md`` for the full conformance argument and the catalogued
disagreements among public implementations.

The indicator produces three outputs:
  - STC:  the oscillator, range [0, 100], NaN during warm-up (bars 0..slow);
  - MACD: the gated MACD line XMAC (0.0 pre-gate), exposed for stage testing;
  - PF:   the first smoothed %D (0.0 pre-gate), exposed for stage testing.
"""

import math
from collections import deque
from typing import List, Any

from ...core.indicator import Indicator
from ...core.metadata import Metadata
from ...core.build_metadata import build_metadata, OutputText
from ...core.identifier import Identifier
from ...core.component_triple_mnemonic import component_triple_mnemonic
from ....entities.bar import Bar
from ....entities.quote import Quote
from ....entities.trade import Trade
from ....entities.scalar import Scalar
from ....entities.bar_component import BarComponent, DEFAULT_BAR_COMPONENT, bar_component_value
from ....entities.quote_component import QuoteComponent, DEFAULT_QUOTE_COMPONENT, quote_component_value
from ....entities.trade_component import TradeComponent, DEFAULT_TRADE_COMPONENT, trade_component_value
from .params import SchaffTrendCycleParams


class _Ema:
    """Stateful streaming EMA: alpha = 2/(period+1), seeds e_0 = x_0.

    Inlined verbatim from the Blau exponential moving average so the indicator
    is a standalone porting unit. Do NOT change its numerics.
    """

    def __init__(self, period: int) -> None:
        self._alpha = 2.0 / (float(period) + 1.0)
        self._prev = 0.0
        self._primed = False

    def update(self, x: float) -> float:
        if not self._primed:
            self._prev = x
            self._primed = True
            return self._prev
        e = self._alpha * x + (1.0 - self._alpha) * self._prev
        self._prev = e
        return e


class SchaffTrendCycle(Indicator):
    """Doug Schaff's Schaff Trend Cycle (STC) indicator."""

    def __init__(self, p: SchaffTrendCycleParams) -> None:
        fast = p.fast
        slow = p.slow
        tclen = p.tclen
        factor = p.factor

        if fast < 1:
            raise ValueError(
                "invalid schaff trend cycle parameters: "
                "fast should be greater than 0")
        if slow < 1:
            raise ValueError(
                "invalid schaff trend cycle parameters: "
                "slow should be greater than 0")
        if tclen < 1:
            raise ValueError(
                "invalid schaff trend cycle parameters: "
                "tclen should be greater than 0")
        if not (0.0 < factor <= 1.0):
            raise ValueError(
                "invalid schaff trend cycle parameters: "
                "factor should be in (0, 1]")

        bc = p.bar_component if p.bar_component is not None else DEFAULT_BAR_COMPONENT
        qc = p.quote_component if p.quote_component is not None else DEFAULT_QUOTE_COMPONENT
        tc = p.trade_component if p.trade_component is not None else DEFAULT_TRADE_COMPONENT

        self._bar_func = bar_component_value(bc)
        self._quote_func = quote_component_value(qc)
        self._trade_func = trade_component_value(tc)

        self._slow = slow
        self._tclen = tclen
        self._factor = factor

        # Two price EMAs forming the MACD line (run every bar).
        self._ema_fast = _Ema(fast)
        self._ema_slow = _Ema(slow)

        # 0-based bar counter (starts at -1, ++ each update).
        self._bar = -1

        # Rolling windows of the last tclen XMAC and PF values.
        self._macd_win: deque = deque(maxlen=tclen)
        self._pf_win: deque = deque(maxlen=tclen)

        # Carried recursion state (default 0.0).
        self._frac1 = 0.0
        self._frac2 = 0.0
        self._pf = 0.0
        self._pff = 0.0

        self._primed = False

        self._mnemonic = f"stc({fast},{slow},{tclen},{factor:.2f}" \
                         f"{component_triple_mnemonic(bc, qc, tc)})"

    def is_primed(self) -> bool:
        return self._primed

    def metadata(self) -> Metadata:
        desc = f"Schaff Trend Cycle {self._mnemonic}"
        return build_metadata(
            Identifier.SCHAFF_TREND_CYCLE,
            self._mnemonic,
            desc,
            [
                OutputText(f"{self._mnemonic} stc", f"{desc} STC"),
                OutputText(f"{self._mnemonic} macd", f"{desc} MACD"),
                OutputText(f"{self._mnemonic} pf", f"{desc} PF"),
            ],
        )

    def update(self, close: float) -> tuple[float, float, float]:
        """Update with a scalar value. Returns (stc, macd, pf)."""
        self._bar += 1
        k = self._bar

        # Price EMAs always advance (they accumulate over the full history).
        ema_fast = self._ema_fast.update(close)
        ema_slow = self._ema_slow.update(close)

        # GATE: XMAC is only assigned while barindex > slow.
        gate_open = k > self._slow
        macd = (ema_fast - ema_slow) if gate_open else 0.0
        self._macd_win.append(macd)

        if not gate_open:
            self._pf_win.append(self._pf)
            return math.nan, macd, self._pf

        # 1st stochastic of the MACD over tclen (guard on the range).
        ll1 = min(self._macd_win)
        rng1 = max(self._macd_win) - ll1
        if rng1 > 0.0:
            self._frac1 = ((macd - ll1) / rng1) * 100.0

        # 1st smoothing: PF = EMA(Frac1, alpha=factor), seed 0.
        self._pf = self._pf + self._factor * (self._frac1 - self._pf)
        self._pf_win.append(self._pf)

        # 2nd stochastic of PF over tclen.
        ll2 = min(self._pf_win)
        rng2 = max(self._pf_win) - ll2
        if rng2 > 0.0:
            self._frac2 = ((self._pf - ll2) / rng2) * 100.0

        # 2nd smoothing: STC = PFF = EMA(Frac2, alpha=factor), seed 0.
        self._pff = self._pff + self._factor * (self._frac2 - self._pff)
        self._primed = True

        return self._pff, macd, self._pf

    def update_scalar(self, sample: Scalar) -> List[Any]:
        stc, macd, pf = self.update(sample.value)
        return [
            Scalar(time=sample.time, value=stc),
            Scalar(time=sample.time, value=macd),
            Scalar(time=sample.time, value=pf),
        ]

    def update_bar(self, sample: Bar) -> List[Any]:
        v = self._bar_func(sample)
        return self.update_scalar(Scalar(time=sample.time, value=v))

    def update_quote(self, sample: Quote) -> List[Any]:
        v = self._quote_func(sample)
        return self.update_scalar(Scalar(time=sample.time, value=v))

    def update_trade(self, sample: Trade) -> List[Any]:
        v = self._trade_func(sample)
        return self.update_scalar(Scalar(time=sample.time, value=v))
