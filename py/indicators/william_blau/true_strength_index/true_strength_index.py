"""True Strength Index (TSI) -- William Blau.

A double-/triple-smoothed momentum oscillator bounded to [-100, +100], paired
with an EMA signal line (the Ergodic form, Blau ch.1.4):

    tsi_k    = 100 * TEMA(mtm, r, s, u) / TEMA(|mtm|, r, s, u)   (the oscillator)
    signal_k = EMA(tsi, ul)_k                                    (ul-period EMA)

where
    mtm_k             = C_k - C_(k-(q-1))                  (q-period momentum)
    TEMA(x, r, s, u)  = EMA(EMA(EMA(x, r), s), u)          (triple EMA cascade)

It is a TWO-output indicator: each update returns ``(tsi, signal)``.

Priming convention -- BOOK / EasyLanguage (Option B), see description.md §2:
    * Each EMA stage seeds on its first received value.
    * Momentum is valid from bar q-1, so all stages seed at bar q-1 together;
      TSI is NaN for bars 0..q-2 and finite from bar q-1 onward.
    * The signal EMA seeds on the first finite TSI (bar q-1), so the signal is
      ALSO NaN for bars 0..q-2; ul == 1 -> signal is a passthrough.

Division guard: denominator == 0 -> oscillator 0.0 (matches Blau_TSI.mq5).
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
from .params import TrueStrengthIndexParams


class _Ema:
    """Stateful streaming EMA: alpha = 2/(period+1), seeds e_0 = x_0.

    Inlined verbatim from the Blau exponential moving average so the indicator
    is a standalone porting unit. Do NOT change its numerics.

    period == 1 -> alpha == 1 -> pure passthrough (output == input).
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


class TrueStrengthIndex(Indicator):
    """William Blau's True Strength Index (TSI) with an EMA signal line."""

    def __init__(self, p: TrueStrengthIndexParams) -> None:
        q = p.q
        r = p.r
        s = p.s
        u = p.u
        ul = p.ul

        if q < 1:
            raise ValueError(
                "invalid true strength index parameters: "
                "q should be greater than 0")
        if r < 1:
            raise ValueError(
                "invalid true strength index parameters: "
                "r should be greater than 0")
        if s < 1:
            raise ValueError(
                "invalid true strength index parameters: "
                "s should be greater than 0")
        if u < 1:
            raise ValueError(
                "invalid true strength index parameters: "
                "u should be greater than 0")
        if ul < 1:
            raise ValueError(
                "invalid true strength index parameters: "
                "ul should be greater than 0")

        bc = p.bar_component if p.bar_component is not None else DEFAULT_BAR_COMPONENT
        qc = p.quote_component if p.quote_component is not None else DEFAULT_QUOTE_COMPONENT
        tc = p.trade_component if p.trade_component is not None else DEFAULT_TRADE_COMPONENT

        self._bar_func = bar_component_value(bc)
        self._quote_func = quote_component_value(qc)
        self._trade_func = trade_component_value(tc)

        self._q = q

        # Rolling window of recent prices; holding q prices means the leftmost
        # element is exactly C_(k-(q-1)).
        self._history: deque = deque(maxlen=q)

        # Two independent 3-stage EMA cascades: one for signed momentum
        # (numerator), one for absolute momentum (denominator).
        self._num_r = _Ema(r)
        self._num_s = _Ema(s)
        self._num_u = _Ema(u)
        self._den_r = _Ema(r)
        self._den_s = _Ema(s)
        self._den_u = _Ema(u)

        # Signal line: a ul-period EMA of the oscillator, advanced ONLY on finite
        # oscillator values, so it shares the oscillator's NaN warm-up region.
        self._signal_ema = _Ema(ul)

        self._primed = False

        self._mnemonic = f"tsi({q},{r},{s},{u}" \
                         f"{component_triple_mnemonic(bc, qc, tc)})"

    def is_primed(self) -> bool:
        return self._primed

    def metadata(self) -> Metadata:
        desc = f"True Strength Index {self._mnemonic}"
        return build_metadata(
            Identifier.TRUE_STRENGTH_INDEX,
            self._mnemonic,
            desc,
            [
                OutputText(f"{self._mnemonic} tsi", f"{desc} TSI"),
                OutputText(f"{self._mnemonic} signal", f"{desc} signal"),
            ],
        )

    def update(self, price: float) -> tuple[float, float]:
        """Update with a scalar value. Returns (tsi, signal)."""
        self._history.append(price)

        # Momentum needs a price from q-1 bars ago, available only once the
        # window holds q prices. Before then neither output is defined and the
        # signal EMA is NOT advanced.
        if len(self._history) < self._q:
            return math.nan, math.nan

        # mtm_k = C_k - C_(k-(q-1)); the leftmost deque element is C_(k-(q-1)).
        mtm = price - self._history[0]
        abs_mtm = abs(mtm)

        # Numerator cascade: TEMA(mtm, r, s, u).
        n = self._num_u.update(self._num_s.update(self._num_r.update(mtm)))
        # Denominator cascade: TEMA(|mtm|, r, s, u).
        d = self._den_u.update(self._den_s.update(self._den_r.update(abs_mtm)))

        # Division guard (Blau_TSI.mq5): denominator 0 -> oscillator 0.0.
        tsi = 0.0 if d == 0.0 else 100.0 * n / d

        # Signal line = EMA(tsi, ul); seeds here on the first finite oscillator.
        signal = self._signal_ema.update(tsi)
        self._primed = True
        return tsi, signal

    def update_scalar(self, sample: Scalar) -> List[Any]:
        tsi, signal = self.update(sample.value)
        return [
            Scalar(time=sample.time, value=tsi),
            Scalar(time=sample.time, value=signal),
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
