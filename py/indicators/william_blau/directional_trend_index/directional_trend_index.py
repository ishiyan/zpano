"""Directional Trend Index (DTI) -- William Blau.

A double-/triple-smoothed High-Low Momentum oscillator bounded to [-100, +100],
paired with an EMA signal line (the Ergodic form, Blau ch.7.3):

    dti_k    = 100 * TEMA(HLM, r, s, u)_k / TEMA(|HLM|, r, s, u)_k
    signal_k = EMA(dti, ul)_k                                  (ul-period EMA)

where the High-Low Momentum is built from how far the high rose and the low
fell relative to q-1 bars ago:

    HMU_k = max(high_k - high_(k-(q-1)), 0)      (upward high movement)
    LMD_k = max(low_(k-(q-1)) - low_k,  0)       (downward low movement)
    HLM_k = HMU_k - LMD_k                         (composite high-low momentum)
    TEMA(x, r, s, u) = EMA(EMA(EMA(x, r), s), u)  (triple EMA cascade)

This is the True Strength Index structure applied to HLM instead of price
momentum. Inputs are the high and low prices only (no close).

It is a TWO-output indicator: each update returns ``(dti, signal)``.

Priming convention -- BOOK / EasyLanguage (Option B), see description.md section 2:
    * HLM is valid from bar q-1 (it needs a high/low from q-1 bars ago), so all
      cascade stages seed at bar q-1 together; both outputs are NaN for bars
      0..q-2 and finite from bar q-1. For q == 1 there is no NaN warm-up.
    * The signal EMA seeds on the first finite oscillator value (bar q-1).
      ul == 1 -> the signal is a passthrough.

Degenerate q == 1: HLM == 0 on every bar, so the denominator is always 0 and
the division guard yields dti == 0.0 for all bars.

Division guard: denominator == 0 -> oscillator 0.0 (matches Blau_DTI.mq5).
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
from .params import DirectionalTrendIndexParams


class _Ema:
    """Stateful streaming EMA: alpha = 2/(period+1), seeds e_0 = x_0.

    Inlined verbatim from the Blau exponential moving average so the indicator
    is a standalone porting unit. Do NOT change its numerics.

    period == 1 -> alpha == 1 -> pure passthrough (output == input).
    """

    def __init__(self, period: int) -> None:
        self._alpha = 2.0 / (float(period) + 1.0)
        self._previous = 0.0
        self._primed = False

    def update(self, x: float) -> float:
        if not self._primed:
            self._previous = x
            self._primed = True
            return self._previous
        e = self._alpha * x + (1.0 - self._alpha) * self._previous
        self._previous = e
        return e


class DirectionalTrendIndex(Indicator):
    """William Blau's Directional Trend Index (DTI) with an EMA signal line."""

    def __init__(self, p: DirectionalTrendIndexParams) -> None:
        q = p.q
        r = p.r
        s = p.s
        u = p.u
        ul = p.ul

        if q < 1:
            raise ValueError(
                "invalid directional trend index parameters: "
                "q should be greater than 0")
        if r < 1:
            raise ValueError(
                "invalid directional trend index parameters: "
                "r should be greater than 0")
        if s < 1:
            raise ValueError(
                "invalid directional trend index parameters: "
                "s should be greater than 0")
        if u < 1:
            raise ValueError(
                "invalid directional trend index parameters: "
                "u should be greater than 0")
        if ul < 1:
            raise ValueError(
                "invalid directional trend index parameters: "
                "ul should be greater than 0")

        # Circular windows of the last q highs and lows. Once q bars have been
        # seen, the oldest value (at the next write position) is the one q-1
        # bars ago.
        self._q = q
        self._highs = [0.0] * q
        self._lows = [0.0] * q
        self._window_count = 0
        self._window_index = 0

        # Two independent 3-stage EMA cascades: one for the signed HLM
        # (numerator), one for the absolute HLM (denominator). Each is wired
        # output -> input.
        self._num_r = _Ema(r)
        self._num_s = _Ema(s)
        self._num_u = _Ema(u)
        self._den_r = _Ema(r)
        self._den_s = _Ema(s)
        self._den_u = _Ema(u)

        # Signal line: a ul-period EMA of the oscillator. Advanced only on
        # finite oscillator values, so it seeds on bar q-1 and shares the
        # oscillator's NaN warm-up region.
        self._signal_ema = _Ema(ul)

        self._primed = False

        self._mnemonic = f"dti({q},{r},{s},{u},{ul})"

    def is_primed(self) -> bool:
        return self._primed

    def metadata(self) -> Metadata:
        desc = f"Directional Trend Index {self._mnemonic}"
        return build_metadata(
            Identifier.DIRECTIONAL_TREND_INDEX,
            self._mnemonic,
            desc,
            [
                OutputText(f"{self._mnemonic} dti", f"{desc} DTI"),
                OutputText(f"{self._mnemonic} signal", f"{desc} signal"),
            ],
        )

    def update(self, high: float, low: float) -> tuple[float, float]:
        """Update with one bar's high and low. Returns (dti, signal)."""
        self._highs[self._window_index] = high
        self._lows[self._window_index] = low
        self._window_index = (self._window_index + 1) % self._q
        if self._window_count < self._q:
            self._window_count += 1

        # HLM needs a high/low from q-1 bars ago. Until then neither output
        # exists -- do NOT advance the EMA cascades.
        if self._window_count < self._q:
            return math.nan, math.nan

        # The oldest value in the full window sits at the next write position:
        # high_(k-(q-1)) and low_(k-(q-1)).
        previous_high = self._highs[self._window_index]
        previous_low = self._lows[self._window_index]

        # Upward high movement and downward low movement, each floored at 0.
        hmu = high - previous_high
        if hmu < 0.0:
            hmu = 0.0
        lmd = previous_low - low
        if lmd < 0.0:
            lmd = 0.0

        # Composite high-low momentum and its magnitude.
        hlm = hmu - lmd
        abs_hlm = abs(hlm)

        # Numerator cascade: TEMA(HLM, r, s, u).
        n = self._num_u.update(self._num_s.update(self._num_r.update(hlm)))
        # Denominator cascade: TEMA(|HLM|, r, s, u).
        d = self._den_u.update(self._den_s.update(self._den_r.update(abs_hlm)))

        # Division guard: denominator 0 -> oscillator 0.0.
        dti = 0.0 if d == 0.0 else 100.0 * n / d

        # Signal line = EMA(dti, ul); seeds on the first finite oscillator.
        signal = self._signal_ema.update(dti)
        self._primed = True
        return dti, signal

    def _update_entity(self, time, high: float, low: float) -> List[Any]:
        dti, signal = self.update(high, low)
        return [
            Scalar(time=time, value=dti),
            Scalar(time=time, value=signal),
        ]

    def update_scalar(self, sample: Scalar) -> List[Any]:
        # A scalar carries a single value, used as both the high and the low.
        v = sample.value
        return self._update_entity(sample.time, v, v)

    def update_bar(self, sample: Bar) -> List[Any]:
        return self._update_entity(sample.time, sample.high, sample.low)

    def update_quote(self, sample: Quote) -> List[Any]:
        # A quote maps the mid price to both the high and the low.
        v = (sample.bid_price + sample.ask_price) / 2
        return self._update_entity(sample.time, v, v)

    def update_trade(self, sample: Trade) -> List[Any]:
        # A trade carries a single price, used as both the high and the low.
        v = sample.price
        return self._update_entity(sample.time, v, v)
