"""Mean Deviation Index (MDI) -- William Blau.

A detrended, double-/triple-smoothed momentum line in raw price units, paired
with an EMA signal line (the Ergodic form, Blau ch. 5):

    md_k     = price_k - EMA(price, r)_k                  (deviation from trend)
    mdi_k    = EMA(EMA(md, s), u)_k                       (the MDI line)
    signal_k = EMA(mdi, ul)_k                             (ul-period EMA)

The price series is detrended by subtracting its own ``r``-period EMA, then the
deviation is smoothed by an ``s``-period EMA and an optional ``u``-period EMA.
Blau notes the MDI approximates the MACD when ``r`` is long and ``s`` is short.

The index is NOT normalized: there is no ``100 * TEMA/TEMA`` ratio and no fixed
range, so the output is in the same price units as the input and may take any
sign or magnitude. Because there is no division there is also no division guard.

It is a TWO-output indicator: each update returns ``(mdi, signal)``.

Priming convention -- BOOK / EasyLanguage (Option B), see description.md §2:
    * Each EMA stage seeds on its first received value.
    * The detrending EMA is defined from bar 0, so ``md_0 = price_0 - price_0 = 0``
      and both smoothing EMAs seed on that 0 -- there is NO NaN warm-up region
      and bar 0 is exactly 0.0.
    * The signal EMA seeds on the bar-0 index value; ul == 1 -> passthrough.

Degenerate case: ``r=1`` makes the detrending EMA a passthrough, so the
deviation is 0 on every bar and the index is identically 0.0.
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
from .params import MeanDeviationIndexParams


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


class MeanDeviationIndex(Indicator):
    """William Blau's Mean Deviation Index (MDI) with an EMA signal line."""

    def __init__(self, p: MeanDeviationIndexParams) -> None:
        r = p.r
        s = p.s
        u = p.u
        ul = p.ul

        if r < 1:
            raise ValueError(
                "invalid mean deviation index parameters: "
                "r should be greater than 0")
        if s < 1:
            raise ValueError(
                "invalid mean deviation index parameters: "
                "s should be greater than 0")
        if u < 1:
            raise ValueError(
                "invalid mean deviation index parameters: "
                "u should be greater than 0")
        if ul < 1:
            raise ValueError(
                "invalid mean deviation index parameters: "
                "ul should be greater than 0")

        bc = p.bar_component if p.bar_component is not None else DEFAULT_BAR_COMPONENT
        qc = p.quote_component if p.quote_component is not None else DEFAULT_QUOTE_COMPONENT
        tc = p.trade_component if p.trade_component is not None else DEFAULT_TRADE_COMPONENT

        self._bar_func = bar_component_value(bc)
        self._quote_func = quote_component_value(qc)
        self._trade_func = trade_component_value(tc)

        # The detrending baseline: a single EMA(r) on the price. The mean
        # deviation is the price minus this trend.
        self._trend = _Ema(r)

        # Two chained EMAs smoothing the deviation: EMA(EMA(md, s), u).
        self._smooth_s = _Ema(s)
        self._smooth_u = _Ema(u)

        # Signal line: a ul-period EMA of the index. The index is finite from
        # bar 0, so this seeds on bar 0 -- no NaN warm-up.
        self._signal_ema = _Ema(ul)

        self._primed = False

        self._mnemonic = f"mdi({r},{s},{u}" \
                         f"{component_triple_mnemonic(bc, qc, tc)})"

    def is_primed(self) -> bool:
        return self._primed

    def metadata(self) -> Metadata:
        desc = f"Mean Deviation Index {self._mnemonic}"
        return build_metadata(
            Identifier.MEAN_DEVIATION_INDEX,
            self._mnemonic,
            desc,
            [
                OutputText(f"{self._mnemonic} mdi", f"{desc} MDI"),
                OutputText(f"{self._mnemonic} signal", f"{desc} signal"),
            ],
        )

    def update(self, price: float) -> tuple[float, float]:
        """Update with a scalar value. Returns (mdi, signal)."""
        # Mean deviation: the price minus its own r-period EMA trend. The
        # baseline EMA seeds on bar 0, so the bar-0 deviation is exactly 0.
        md = price - self._trend.update(price)

        # Smooth the deviation: EMA(EMA(md, s), u). No normalization, no guard.
        mdi = self._smooth_u.update(self._smooth_s.update(md))

        # Signal line = EMA(mdi, ul); seeds here on the bar-0 index value.
        signal = self._signal_ema.update(mdi)
        self._primed = True
        return mdi, signal

    def update_scalar(self, sample: Scalar) -> List[Any]:
        mdi, signal = self.update(sample.value)
        return [
            Scalar(time=sample.time, value=mdi),
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
