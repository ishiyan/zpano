"""Tick Volume Indicator (TVI) -- William Blau.

A normalized, double-/triple-smoothed oscillator built from the balance of
upticks vs downticks inside each bar, bounded to [-100, +100] (Blau ch.4, ch.10):

    tvi_k = 100 * (TEMA(up, r, s, u) - TEMA(down, r, s, u))
                / (TEMA(up, r, s, u) + TEMA(down, r, s, u))

where

    TEMA(x, r, s, u) = EMA(EMA(EMA(x, r), s), u)           (triple EMA cascade)

``u == 1`` recovers the book's double-smoothed TVI(r, s), because EMA(., 1) is a
passthrough. Because it is built from intra-bar tick direction rather than from
the close vs a previous close, the TVI is immune to opening gaps.

Inputs are two non-negative series, ``upticks`` and ``downticks``. Genuine tick
counts are fed through ``update(upticks, downticks)``. The entity updates derive
a deterministic proxy:

    * Bar: up = close - low, down = high - close (the intra-bar range split).
    * Scalar / Trade / Quote: a magnitude tick rule against the previous value
      (value, price or mid price): up = max(x - previous, 0),
      down = max(previous - x, 0). The first sample yields (0, 0). On such a
      single-valued series the TVI reduces to a True Strength Index of the
      one-step momentum.

Priming convention -- BOOK / EasyLanguage (Option B), see description.md §3:
    * Each EMA stage seeds on its first received value.
    * There is NO NaN warm-up region; both cascades seed on the first update
      and the output is finite for every update.

Division guard: denominator == 0 (a fully flat market) -> output 0.0.
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
from .params import TickVolumeIndicatorParams


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


class TickVolumeIndicator(Indicator):
    """William Blau's Tick Volume Indicator (TVI)."""

    def __init__(self, p: TickVolumeIndicatorParams) -> None:
        r = p.r
        s = p.s
        u = p.u

        if r < 1:
            raise ValueError(
                "invalid tick volume indicator parameters: "
                "r should be greater than 0")
        if s < 1:
            raise ValueError(
                "invalid tick volume indicator parameters: "
                "s should be greater than 0")
        if u < 1:
            raise ValueError(
                "invalid tick volume indicator parameters: "
                "u should be greater than 0")

        # Two independent 3-stage EMA cascades: one for the upticks, one for
        # the downticks. Each is wired output -> input.
        self._up_r = _Ema(r)
        self._up_s = _Ema(s)
        self._up_u = _Ema(u)
        self._down_r = _Ema(r)
        self._down_s = _Ema(s)
        self._down_u = _Ema(u)

        # Tick-rule state for single-valued samples (scalar, trade, quote).
        self._previous = 0.0
        self._has_previous = False

        self._primed = False

        self._mnemonic = f"tvi({r},{s},{u})"

    def is_primed(self) -> bool:
        return self._primed

    def metadata(self) -> Metadata:
        desc = f"Tick Volume Indicator {self._mnemonic}"
        return build_metadata(
            Identifier.TICK_VOLUME_INDICATOR,
            self._mnemonic,
            desc,
            [OutputText(self._mnemonic, desc)],
        )

    def update(self, upticks: float, downticks: float) -> float:
        """Update with one bar's upticks and downticks. Returns the TVI value."""
        # Upticks cascade: TEMA(up, r, s, u).
        up = self._up_u.update(self._up_s.update(self._up_r.update(upticks)))
        # Downticks cascade: TEMA(down, r, s, u).
        down = self._down_u.update(self._down_s.update(self._down_r.update(downticks)))

        self._primed = True

        # Division guard (Appendix B): fully flat smoothed volume -> 0.0.
        denominator = up + down
        if denominator == 0.0:
            return 0.0
        return 100.0 * (up - down) / denominator

    def _update_tick_rule(self, value: float) -> float:
        """Derive upticks/downticks from the change vs the previous value."""
        upticks = 0.0
        downticks = 0.0
        if self._has_previous:
            diff = value - self._previous
            if diff > 0.0:
                upticks = diff
            elif diff < 0.0:
                downticks = -diff

        self._previous = value
        self._has_previous = True
        return self.update(upticks, downticks)

    def update_scalar(self, sample: Scalar) -> List[Any]:
        # A scalar carries a single value: apply the tick rule to it.
        return [Scalar(time=sample.time, value=self._update_tick_rule(sample.value))]

    def update_bar(self, sample: Bar) -> List[Any]:
        # A bar splits its range: up = close - low, down = high - close.
        value = self.update(sample.close - sample.low, sample.high - sample.close)
        return [Scalar(time=sample.time, value=value)]

    def update_quote(self, sample: Quote) -> List[Any]:
        # A quote applies the tick rule to its mid price.
        return [Scalar(time=sample.time, value=self._update_tick_rule(sample.mid()))]

    def update_trade(self, sample: Trade) -> List[Any]:
        # A trade applies the tick rule to its price.
        return [Scalar(time=sample.time, value=self._update_tick_rule(sample.price))]
