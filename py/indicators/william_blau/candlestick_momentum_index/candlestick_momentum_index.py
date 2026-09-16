"""Candlestick Momentum Index (CMI) -- William Blau.

A double-/triple-smoothed intra-bar momentum oscillator bounded to [-100, +100],
paired with an EMA signal line (the Ergodic form, Blau ch.6.4):

    cmi_k    = 100 * TEMA(cmtm, r, s, u) / TEMA(|cmtm|, r, s, u)   (oscillator)
    signal_k = EMA(cmi, ul)_k                                      (ul-period EMA)

where the candle momentum is the signed candle body

    cmtm_k           = close_k - open_k                    (intra-bar momentum)
    TEMA(x, r, s, u) = EMA(EMA(EMA(x, r), s), u)           (triple EMA cascade)

This is the True Strength Index structure applied to the candle body instead of
price momentum. Because it only looks *inside* each bar it is immune to
inter-bar gaps. Inputs are the open and close prices only.

It is a TWO-output indicator: each update returns ``(cmi, signal)``.

Priming convention -- BOOK / EasyLanguage (Option B), see description.md §2:
    * Each EMA stage seeds on its first received value.
    * The candle momentum is defined from bar 0, so there is NO NaN warm-up
      region; all six cascade stages and the signal EMA seed on bar 0 and both
      outputs are finite for every bar.
    * ul == 1 -> the signal is a passthrough (signal == cmi).

Division guard: denominator == 0 -> oscillator 0.0 (matches Blau_CMI.mq5).
"""

import math
from typing import List, Any

from ...core.indicator import Indicator
from ...core.metadata import Metadata
from ...core.build_metadata import build_metadata, OutputText
from ...core.identifier import Identifier
from ....entities.bar import Bar
from ....entities.quote import Quote
from ....entities.trade import Trade
from ....entities.scalar import Scalar
from .params import CandlestickMomentumIndexParams


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


class CandlestickMomentumIndex(Indicator):
    """William Blau's Candlestick Momentum Index (CMI) with an EMA signal line."""

    def __init__(self, p: CandlestickMomentumIndexParams) -> None:
        r = p.r
        s = p.s
        u = p.u
        ul = p.ul

        if r < 1:
            raise ValueError(
                "invalid candlestick momentum index parameters: "
                "r should be greater than 0")
        if s < 1:
            raise ValueError(
                "invalid candlestick momentum index parameters: "
                "s should be greater than 0")
        if u < 1:
            raise ValueError(
                "invalid candlestick momentum index parameters: "
                "u should be greater than 0")
        if ul < 1:
            raise ValueError(
                "invalid candlestick momentum index parameters: "
                "ul should be greater than 0")

        # Two independent 3-stage EMA cascades: one for the signed candle
        # momentum (numerator), one for the absolute candle momentum
        # (denominator). Each is wired output -> input.
        self._num_r = _Ema(r)
        self._num_s = _Ema(s)
        self._num_u = _Ema(u)
        self._den_r = _Ema(r)
        self._den_s = _Ema(s)
        self._den_u = _Ema(u)

        # Signal line: a ul-period EMA of the oscillator. The oscillator is
        # finite from bar 0, so this seeds on bar 0 -- no NaN warm-up.
        self._signal_ema = _Ema(ul)

        self._primed = False

        self._mnemonic = f"cmi({r},{s},{u},{ul})"

    def is_primed(self) -> bool:
        return self._primed

    def metadata(self) -> Metadata:
        desc = f"Candlestick Momentum Index {self._mnemonic}"
        return build_metadata(
            Identifier.CANDLESTICK_MOMENTUM_INDEX,
            self._mnemonic,
            desc,
            [
                OutputText(f"{self._mnemonic} cmi", f"{desc} CMI"),
                OutputText(f"{self._mnemonic} signal", f"{desc} signal"),
            ],
        )

    def update(self, open_: float, close: float) -> tuple[float, float]:
        """Update with one bar's open and close. Returns (cmi, signal)."""
        # Candle momentum: the signed body of the candle.
        cmtm = close - open_
        abs_cmtm = abs(cmtm)

        # Numerator cascade: TEMA(cmtm, r, s, u).
        n = self._num_u.update(self._num_s.update(self._num_r.update(cmtm)))
        # Denominator cascade: TEMA(|cmtm|, r, s, u).
        d = self._den_u.update(self._den_s.update(self._den_r.update(abs_cmtm)))

        # Division guard (Blau_CMI.mq5): denominator 0 -> oscillator 0.0.
        cmi = 0.0 if d == 0.0 else 100.0 * n / d

        # Signal line = EMA(cmi, ul); seeds on bar 0's oscillator value.
        signal = self._signal_ema.update(cmi)
        self._primed = True
        return cmi, signal

    def _update_entity(self, time, open_: float, close: float) -> List[Any]:
        cmi, signal = self.update(open_, close)
        return [
            Scalar(time=time, value=cmi),
            Scalar(time=time, value=signal),
        ]

    def update_scalar(self, sample: Scalar) -> List[Any]:
        # A scalar carries a single value, so open == close and the candle
        # momentum is zero.
        return self._update_entity(sample.time, sample.value, sample.value)

    def update_bar(self, sample: Bar) -> List[Any]:
        return self._update_entity(sample.time, sample.open, sample.close)

    def update_quote(self, sample: Quote) -> List[Any]:
        return self._update_entity(sample.time, sample.bid_price, sample.ask_price)

    def update_trade(self, sample: Trade) -> List[Any]:
        # A trade carries a single price, so open == close and the candle
        # momentum is zero.
        return self._update_entity(sample.time, sample.price, sample.price)
