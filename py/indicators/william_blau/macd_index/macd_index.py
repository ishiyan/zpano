"""MACD Index (MACD_I) -- William Blau.

Blau's MACD line is the difference of two EMAs of the close, optionally
smoothed by a third EMA, paired with an EMA signal line (the Ergodic form,
Blau ch. 5):

    macd_k   = EMA(close, s)_k - EMA(close, r)_k          (MACD line; s fast, r slow)
    macdi_k  = EMA(macd, u)_k                             (the MACD_I line)
    signal_k = EMA(macdi, ul)_k                           (ul-period EMA)

with the fast period ``s`` strictly shorter than the slow period ``r``
(``s < r``). Set ``u = 1`` to recover the book's pure two-EMA MACD line. Blau
notes the MACD and the MDI are both double-smoothed momentum indicators with
nearly interchangeable shapes (within a scale factor).

The index is NOT normalized: there is no ``100 * TEMA/TEMA`` ratio and no
fixed range, so the output is in the same price units as the input and may
take any sign or magnitude. Because there is no division there is also no
division guard.

It is a TWO-output indicator: each update returns ``(macdi, signal)``.

Priming convention -- BOOK / EasyLanguage (Option B), see description.md §2:
    * Both price EMAs seed on bar 0 (to close_0), so the MACD line is 0.0 on
      bar 0 and finite on every bar thereafter -- there is NO NaN warm-up
      region.
    * The ``u`` smoothing and the signal EMA likewise seed on bar 0.
"""

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
from .params import MacdIndexParams


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


class MacdIndex(Indicator):
    """William Blau's MACD Index (MACD_I) with an EMA signal line."""

    def __init__(self, p: MacdIndexParams) -> None:
        r = p.r
        s = p.s
        u = p.u
        ul = p.ul

        if r < 1:
            raise ValueError(
                "invalid macd index parameters: "
                "r should be greater than 0")
        if s < 1:
            raise ValueError(
                "invalid macd index parameters: "
                "s should be greater than 0")
        if u < 1:
            raise ValueError(
                "invalid macd index parameters: "
                "u should be greater than 0")
        if ul < 1:
            raise ValueError(
                "invalid macd index parameters: "
                "ul should be greater than 0")
        if s >= r:
            raise ValueError(
                "invalid macd index parameters: "
                "s (fast) should be less than r (slow)")

        bc = p.bar_component if p.bar_component is not None else DEFAULT_BAR_COMPONENT
        qc = p.quote_component if p.quote_component is not None else DEFAULT_QUOTE_COMPONENT
        tc = p.trade_component if p.trade_component is not None else DEFAULT_TRADE_COMPONENT

        self._bar_func = bar_component_value(bc)
        self._quote_func = quote_component_value(qc)
        self._trade_func = trade_component_value(tc)

        # The two price EMAs forming the MACD line = EMA(close, s) - EMA(close, r).
        self._ema_fast = _Ema(s)
        self._ema_slow = _Ema(r)

        # Third smoothing EMA applied to the MACD line: EMA(macd, u).
        self._smooth_u = _Ema(u)

        # Signal line: a ul-period EMA of the index. The index is finite from
        # bar 0, so this seeds on bar 0 -- no NaN warm-up.
        self._signal_ema = _Ema(ul)

        self._primed = False

        self._mnemonic = f"macdi({r},{s},{u}" \
                         f"{component_triple_mnemonic(bc, qc, tc)})"

    def is_primed(self) -> bool:
        return self._primed

    def metadata(self) -> Metadata:
        desc = f"MACD Index {self._mnemonic}"
        return build_metadata(
            Identifier.MACD_INDEX,
            self._mnemonic,
            desc,
            [
                OutputText(f"{self._mnemonic} macdi", f"{desc} MACDI"),
                OutputText(f"{self._mnemonic} signal", f"{desc} signal"),
            ],
        )

    def update(self, price: float) -> tuple[float, float]:
        """Update with a scalar value. Returns (macdi, signal)."""
        # MACD line = fast EMA - slow EMA. Both seed at bar 0 (to close_0), so
        # the line is 0.0 on bar 0 and defined on every bar thereafter.
        macd = self._ema_fast.update(price) - self._ema_slow.update(price)

        # Smooth the MACD line: EMA(macd, u). No normalization, no guard.
        macdi = self._smooth_u.update(macd)

        # Signal line = EMA(macdi, ul); seeds here on the bar-0 index value.
        signal = self._signal_ema.update(macdi)
        self._primed = True
        return macdi, signal

    def update_scalar(self, sample: Scalar) -> List[Any]:
        macdi, signal = self.update(sample.value)
        return [
            Scalar(time=sample.time, value=macdi),
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
