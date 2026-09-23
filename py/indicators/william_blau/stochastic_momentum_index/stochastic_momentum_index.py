"""Stochastic Momentum Index (SMI) -- William Blau.

A double-/triple-smoothed stochastic oscillator bounded to [-100, +100], paired
with an EMA signal line (the Ergodic form, Blau ch. 3.4):

    smi_k    = 100 * TEMA(sm, r, s, u) / TEMA(hr, r, s, u)
    signal_k = EMA(smi, ul)_k                                  (ul-period EMA)

where, over the last q bars,

    HH_k             = max(high over last q bars)
    LL_k             = min(low  over last q bars)
    sm_k             = close_k - 0.5 * (HH_k + LL_k)   (distance from range midpoint)
    hr_k             = 0.5 * (HH_k - LL_k)             (half of the q-bar range >= 0)
    TEMA(x, r, s, u) = EMA(EMA(EMA(x, r), s), u)       (triple EMA cascade)

Where the ordinary stochastic measures where the close sits inside the recent
high-low range, the SMI measures the close relative to the midpoint of that
range. Because |sm| <= hr on every bar, the ratio is bounded to [-100, +100].
With q == 1 it is Blau's one-day stochastic (sentiment indicator). Inputs are
the high, low and close prices.

It is a TWO-output indicator: each update returns ``(smi, signal)``.

Priming convention -- BOOK / EasyLanguage (Option B), see description.md section 2:
    * sm and hr become valid once q bars of high/low exist, i.e. at bar q-1.
    * All six cascade stages seed at bar q-1 together; both outputs are NaN for
      bars 0..q-2 and finite from bar q-1. For q == 1 there is no NaN warm-up.
    * ul == 1 -> the signal is a passthrough (signal == smi).

Division guard: denominator <= 0 -> oscillator 0.0 (flat HH == LL window).
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
from .params import StochasticMomentumIndexParams


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


class StochasticMomentumIndex(Indicator):
    """William Blau's Stochastic Momentum Index (SMI) with an EMA signal line."""

    def __init__(self, p: StochasticMomentumIndexParams) -> None:
        q = p.q
        r = p.r
        s = p.s
        u = p.u
        ul = p.ul

        if q < 1:
            raise ValueError(
                "invalid stochastic momentum index parameters: "
                "q should be greater than 0")
        if r < 1:
            raise ValueError(
                "invalid stochastic momentum index parameters: "
                "r should be greater than 0")
        if s < 1:
            raise ValueError(
                "invalid stochastic momentum index parameters: "
                "s should be greater than 0")
        if u < 1:
            raise ValueError(
                "invalid stochastic momentum index parameters: "
                "u should be greater than 0")
        if ul < 1:
            raise ValueError(
                "invalid stochastic momentum index parameters: "
                "ul should be greater than 0")

        # Circular windows of the last q highs and lows. Once q bars have been
        # seen they hold exactly the bars [k-(q-1) .. k].
        self._q = q
        self._highs = [0.0] * q
        self._lows = [0.0] * q
        self._window_count = 0
        self._window_index = 0

        # Two independent 3-stage EMA cascades: one for the stochastic momentum
        # (numerator), one for the half-range (denominator). Each is wired
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

        self._mnemonic = f"smi({q},{r},{s},{u},{ul})"

    def is_primed(self) -> bool:
        return self._primed

    def metadata(self) -> Metadata:
        desc = f"Stochastic Momentum Index {self._mnemonic}"
        return build_metadata(
            Identifier.STOCHASTIC_MOMENTUM_INDEX,
            self._mnemonic,
            desc,
            [
                OutputText(f"{self._mnemonic} smi", f"{desc} SMI"),
                OutputText(f"{self._mnemonic} signal", f"{desc} signal"),
            ],
        )

    def update(self, high: float, low: float,
               close: float) -> tuple[float, float]:
        """Update with one bar's high, low and close. Returns (smi, signal)."""
        self._highs[self._window_index] = high
        self._lows[self._window_index] = low
        self._window_index = (self._window_index + 1) % self._q
        if self._window_count < self._q:
            self._window_count += 1

        # Need q bars of high/low before the stochastic is defined. Until then
        # neither output exists -- do NOT advance the EMA cascades.
        if self._window_count < self._q:
            return math.nan, math.nan

        # Rolling extremes over the last q bars.
        hh = max(self._highs)
        ll = min(self._lows)

        # Stochastic momentum (signed) and half-range (non-negative).
        sm = close - 0.5 * (hh + ll)
        hr = 0.5 * (hh - ll)

        # Numerator cascade: TEMA(sm, r, s, u).
        n = self._num_u.update(self._num_s.update(self._num_r.update(sm)))
        # Denominator cascade: TEMA(hr, r, s, u).
        d = self._den_u.update(self._den_s.update(self._den_r.update(hr)))

        # Division guard: flat window so far -> oscillator 0.0.
        smi = 0.0 if d <= 0.0 else 100.0 * n / d

        # Signal line = EMA(smi, ul); seeds on the first finite oscillator.
        signal = self._signal_ema.update(smi)
        self._primed = True
        return smi, signal

    def _update_entity(self, time, high: float, low: float,
                       close: float) -> List[Any]:
        smi, signal = self.update(high, low, close)
        return [
            Scalar(time=time, value=smi),
            Scalar(time=time, value=signal),
        ]

    def update_scalar(self, sample: Scalar) -> List[Any]:
        # A scalar carries a single value, used as the high, the low and the close.
        v = sample.value
        return self._update_entity(sample.time, v, v, v)

    def update_bar(self, sample: Bar) -> List[Any]:
        return self._update_entity(sample.time, sample.high, sample.low,
                                   sample.close)

    def update_quote(self, sample: Quote) -> List[Any]:
        # A quote maps the mid price to the high, the low and the close.
        v = (sample.bid_price + sample.ask_price) / 2
        return self._update_entity(sample.time, v, v, v)

    def update_trade(self, sample: Trade) -> List[Any]:
        # A trade carries a single price, used as the high, the low and the close.
        v = sample.price
        return self._update_entity(sample.time, v, v, v)
