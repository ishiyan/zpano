"""Candlestick Strength Index (CSI) -- William Blau ("CandleStick Indicator").

A double-/triple-smoothed candle-body-vs-range oscillator bounded to
[-100, +100], paired with an EMA signal line (the Ergodic form, Blau ch. 6.4):

    csi_k    = 100 * TEMA(close - open, r, s, u) / TEMA(high - low, r, s, u)
    signal_k = EMA(csi, ul)_k                                  (ul-period EMA)

where the two intra-bar quantities are

    co_k             = close_k - open_k                  (signed candle body)
    hl_k             = high_k  - low_k                   (the bar's range >= 0)
    TEMA(x, r, s, u) = EMA(EMA(EMA(x, r), s), u)          (triple EMA cascade)

This is Blau's CandleStick Indicator (book, ch. 6; Appendix B, Figure B-15;
MQL5 Blau_CSI.mq5). The numerator is the *signed* candle body (close minus
open): positive for bullish bars, negative for bearish bars. The denominator is
the smoothed bar range. Because every bar has |close - open| <= high - low, the
ratio is bounded to [-100, +100]: +100 when closes pin the high while opens pin
the low (relentless bullish bodies), -100 in the mirror-image bearish case, and
0 when bodies net out. Inputs are the open, high, low and close prices.

It is the range-normalized sibling of the Candlestick Momentum Index (CMI):
both share the signed numerator TEMA(close - open), but the CSI divides by the
smoothed range while the CMI divides by the smoothed absolute body.

It is a TWO-output indicator: each update returns ``(csi, signal)``.

Priming convention -- BOOK / EasyLanguage (Option B), see description.md section 2:
    * Each EMA stage seeds on its first received value.
    * Both intra-bar series are defined from bar 0, so there is NO NaN warm-up
      region; all six cascade stages and the signal EMA seed on bar 0 and both
      outputs are finite for every bar.
    * ul == 1 -> the signal is a passthrough (signal == csi).

Division guard: denominator <= 0 -> oscillator 0.0 (zero-range market).
"""

from typing import List, Any

from ...core.indicator import Indicator
from ...core.metadata import Metadata
from ...core.build_metadata import build_metadata, OutputText
from ...core.identifier import Identifier
from ....entities.bar import Bar
from ....entities.quote import Quote
from ....entities.trade import Trade
from ....entities.scalar import Scalar
from .params import CandlestickStrengthIndexParams


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


class CandlestickStrengthIndex(Indicator):
    """William Blau's Candlestick Strength Index (CSI) with an EMA signal line."""

    def __init__(self, p: CandlestickStrengthIndexParams) -> None:
        r = p.r
        s = p.s
        u = p.u
        ul = p.ul

        if r < 1:
            raise ValueError(
                "invalid candlestick strength index parameters: "
                "r should be greater than 0")
        if s < 1:
            raise ValueError(
                "invalid candlestick strength index parameters: "
                "s should be greater than 0")
        if u < 1:
            raise ValueError(
                "invalid candlestick strength index parameters: "
                "u should be greater than 0")
        if ul < 1:
            raise ValueError(
                "invalid candlestick strength index parameters: "
                "ul should be greater than 0")

        # Two independent 3-stage EMA cascades: one for the signed candle body
        # (numerator), one for the high-low range (denominator). Each is wired
        # output -> input.
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

        self._mnemonic = f"csi({r},{s},{u},{ul})"

    def is_primed(self) -> bool:
        return self._primed

    def metadata(self) -> Metadata:
        desc = f"Candlestick Strength Index {self._mnemonic}"
        return build_metadata(
            Identifier.CANDLESTICK_STRENGTH_INDEX,
            self._mnemonic,
            desc,
            [
                OutputText(f"{self._mnemonic} csi", f"{desc} CSI"),
                OutputText(f"{self._mnemonic} signal", f"{desc} signal"),
            ],
        )

    def update(self, open_: float, high: float, low: float,
               close: float) -> tuple[float, float]:
        """Update with one bar's open, high, low and close. Returns (csi, signal)."""
        # Two intra-bar quantities: signed candle body and (non-negative) range.
        co = close - open_
        hl = high - low

        # Numerator cascade: TEMA(close-open, r, s, u).
        n = self._num_u.update(self._num_s.update(self._num_r.update(co)))
        # Denominator cascade: TEMA(high-low, r, s, u).
        d = self._den_u.update(self._den_s.update(self._den_r.update(hl)))

        # Division guard: zero range so far -> oscillator 0.0.
        csi = 0.0 if d <= 0.0 else 100.0 * n / d

        # Signal line = EMA(csi, ul); seeds on bar 0's oscillator value.
        signal = self._signal_ema.update(csi)
        self._primed = True
        return csi, signal

    def _update_entity(self, time, open_: float, high: float, low: float,
                       close: float) -> List[Any]:
        csi, signal = self.update(open_, high, low, close)
        return [
            Scalar(time=time, value=csi),
            Scalar(time=time, value=signal),
        ]

    def update_scalar(self, sample: Scalar) -> List[Any]:
        # A scalar carries a single value, so the candle body and the bar range
        # are both zero.
        return self._update_entity(sample.time, sample.value, sample.value,
                                   sample.value, sample.value)

    def update_bar(self, sample: Bar) -> List[Any]:
        return self._update_entity(sample.time, sample.open, sample.high,
                                   sample.low, sample.close)

    def update_quote(self, sample: Quote) -> List[Any]:
        # A quote maps the bid to the open and the low, and the ask to the close
        # and the high, so the candle body and the bar range are both the spread.
        return self._update_entity(sample.time, sample.bid_price, sample.ask_price,
                                   sample.bid_price, sample.ask_price)

    def update_trade(self, sample: Trade) -> List[Any]:
        # A trade carries a single price, so the candle body and the bar range
        # are both zero.
        return self._update_entity(sample.time, sample.price, sample.price,
                                   sample.price, sample.price)
