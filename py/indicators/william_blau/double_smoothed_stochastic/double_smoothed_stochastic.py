"""Double Smoothed Stochastic (DSS) -- William Blau.

A classic double-smoothed stochastic oscillator bounded to [0, 100], paired
with a short simple-moving-average signal line:

    dss_k    = 100 * EMA(EMA(st, r), s)_k / EMA(EMA(rng, r), s)_k
    signal_k = SMA(dss, g)_k                                   (g-period SMA)

where, over the last q bars,

    HH_k  = max(high over last q bars)
    LL_k  = min(low  over last q bars)
    st_k  = close_k - LL_k            (raw stochastic: close above the low >= 0)
    rng_k = HH_k - LL_k               (q-bar range >= 0)

The raw stochastic and the range are smoothed *separately* with the same
two-stage EMA cascade (r then s), then divided. Because 0 <= st <= rng on every
bar, the ratio is bounded to [0, 100]. With q == 1 it is Blau's one-bar HLC
index. It is exactly the MQL5 Blau_TStochI with its third EMA period u = 1.
Inputs are the high, low and close prices.

It is a TWO-output indicator: each update returns ``(dss, signal)``.

Priming convention -- BOOK / EasyLanguage (Option B), see description.md section 2:
    * st and rng become valid once q bars of high/low exist, i.e. at bar q-1.
    * All four cascade stages seed at bar q-1 together; both outputs are NaN for
      bars 0..q-2 and finite from bar q-1. For q == 1 there is no NaN warm-up.
    * The signal SMA seeds on the first finite oscillator value and returns the
      mean of the oscillator values seen so far (expanding window <= g), then
      the full g-bar rolling mean. g == 1 -> the signal is a passthrough.

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
from .params import DoubleSmoothedStochasticParams


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


class _Sma:
    """Stateful streaming SMA over the last ``period`` inputs.

    Returns the mean of the window's current contents on every update: an
    expanding window while fewer than ``period`` values have arrived, then a
    rolling ``period``-bar window. No NaN warm-up (finite from the first input).

    period == 1 -> pure passthrough (output == input).
    """

    def __init__(self, period: int) -> None:
        self._period = period
        self._window = [0.0] * period
        self._count = 0
        self._index = 0

    def update(self, x: float) -> float:
        self._window[self._index] = x
        self._index = (self._index + 1) % self._period
        if self._count < self._period:
            self._count += 1

        # Naive left-to-right sum from the oldest to the newest value (NOT a
        # compensated sum), so that every port reproduces the same values.
        # Once full, the oldest value sits at the next write position.
        start = self._index if self._count == self._period else 0
        sum = 0.0
        for i in range(self._count):
            sum += self._window[(start + i) % self._period]
        return sum / self._count


class DoubleSmoothedStochastic(Indicator):
    """William Blau's Double Smoothed Stochastic (DSS) with an SMA signal line."""

    def __init__(self, p: DoubleSmoothedStochasticParams) -> None:
        q = p.q
        r = p.r
        s = p.s
        g = p.g

        if q < 1:
            raise ValueError(
                "invalid double smoothed stochastic parameters: "
                "q should be greater than 0")
        if r < 1:
            raise ValueError(
                "invalid double smoothed stochastic parameters: "
                "r should be greater than 0")
        if s < 1:
            raise ValueError(
                "invalid double smoothed stochastic parameters: "
                "s should be greater than 0")
        if g < 1:
            raise ValueError(
                "invalid double smoothed stochastic parameters: "
                "g should be greater than 0")

        # Circular windows of the last q highs and lows. Once q bars have been
        # seen they hold exactly the bars [k-(q-1) .. k].
        self._q = q
        self._highs = [0.0] * q
        self._lows = [0.0] * q
        self._window_count = 0
        self._window_index = 0

        # Two independent 2-stage EMA cascades: one for the raw stochastic
        # (numerator), one for the range (denominator). Each is wired
        # output -> input.
        self._num_r = _Ema(r)
        self._num_s = _Ema(s)
        self._den_r = _Ema(r)
        self._den_s = _Ema(s)

        # Signal line: a g-period SMA of the oscillator. Advanced only on
        # finite oscillator values, so it seeds on bar q-1 and shares the
        # oscillator's NaN warm-up region.
        self._signal_sma = _Sma(g)

        self._primed = False

        self._mnemonic = f"dss({q},{r},{s},{g})"

    def is_primed(self) -> bool:
        return self._primed

    def metadata(self) -> Metadata:
        desc = f"Double Smoothed Stochastic {self._mnemonic}"
        return build_metadata(
            Identifier.DOUBLE_SMOOTHED_STOCHASTIC,
            self._mnemonic,
            desc,
            [
                OutputText(f"{self._mnemonic} dss", f"{desc} DSS"),
                OutputText(f"{self._mnemonic} signal", f"{desc} signal"),
            ],
        )

    def update(self, high: float, low: float,
               close: float) -> tuple[float, float]:
        """Update with one bar's high, low and close. Returns (dss, signal)."""
        self._highs[self._window_index] = high
        self._lows[self._window_index] = low
        self._window_index = (self._window_index + 1) % self._q
        if self._window_count < self._q:
            self._window_count += 1

        # Need q bars of high/low before the stochastic is defined. Until then
        # neither output exists -- do NOT advance the EMA cascades or the SMA.
        if self._window_count < self._q:
            return math.nan, math.nan

        # Rolling extremes over the last q bars.
        hh = max(self._highs)
        ll = min(self._lows)

        # Raw stochastic and range (both non-negative).
        st = close - ll
        rng = hh - ll

        # Numerator cascade: EMA(EMA(st, r), s).
        n = self._num_s.update(self._num_r.update(st))
        # Denominator cascade: EMA(EMA(rng, r), s).
        d = self._den_s.update(self._den_r.update(rng))

        # Division guard: flat window so far -> oscillator 0.0.
        dss = 0.0 if d <= 0.0 else 100.0 * n / d

        # Signal line = SMA(dss, g); seeds on the first finite oscillator.
        signal = self._signal_sma.update(dss)
        self._primed = True
        return dss, signal

    def _update_entity(self, time, high: float, low: float,
                       close: float) -> List[Any]:
        dss, signal = self.update(high, low, close)
        return [
            Scalar(time=time, value=dss),
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
