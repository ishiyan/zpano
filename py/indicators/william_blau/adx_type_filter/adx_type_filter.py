"""ADX-Type Filter (ATF) -- William Blau.

A non-negative trend-strength filter, analogous to Wilder's ADX, built by
rectifying and double-smoothing a bipolar momentum series (book Fig. B-24):

    ATF(Price, r, s) = EMA(|EMA(Price, r)|, s)

The inner EMA(r) smooths the signed momentum, the absolute value discards the
direction and keeps the amplitude, and the outer EMA(s) smooths the amplitude.
A rising ATF signals a strengthening trend, a falling ATF a ranging market.

The bipolar momentum is selected by the source:
    * TSI_MOMENTUM   : C - C[q-1]                               (TSI_ATF)
    * SMI_MOMENTUM   : C - 0.5*(HH(q) + LL(q))                  (SMI_ATF)
    * DTI_MOMENTUM   : max(H - H[q-1], 0) - max(L[q-1] - L, 0)
    * TVI_BALANCE    : upticks - downticks (2C - H - L for a bar)
    * TSI_NORMALIZED : TSI(q, r, 1, 1), which replaces the inner EMA (r = 1).

Priming convention -- BOOK / EasyLanguage (Option B):
    * Each EMA stage seeds on its first finite momentum value.
    * A NaN momentum (the q-bar look-back warm-up) is propagated: the output is
      NaN and the EMAs do not advance.
    * The output is always >= 0.

Reference:

Blau, William (1995). Momentum, Direction, and Divergence, Appendix B Fig. B-24.
Wiley.
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
from ....entities.bar_component import DEFAULT_BAR_COMPONENT, bar_component_value
from ....entities.quote_component import DEFAULT_QUOTE_COMPONENT, quote_component_value
from ....entities.trade_component import DEFAULT_TRADE_COMPONENT, trade_component_value
from ..true_strength_index.true_strength_index import TrueStrengthIndex
from ..true_strength_index.params import TrueStrengthIndexParams
from .params import AdxTypeFilterParams, AdxTypeFilterSource


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


_SOURCE_MNEMONICS = {
    AdxTypeFilterSource.TSI_MOMENTUM: "tsi",
    AdxTypeFilterSource.SMI_MOMENTUM: "smi",
    AdxTypeFilterSource.DTI_MOMENTUM: "dti",
    AdxTypeFilterSource.TVI_BALANCE: "tvi",
    AdxTypeFilterSource.TSI_NORMALIZED: "tsin",
}


class AdxTypeFilter(Indicator):
    """William Blau's ADX-Type Filter (ATF)."""

    def __init__(self, params: AdxTypeFilterParams) -> None:
        try:
            source = AdxTypeFilterSource(params.source)
        except ValueError:
            raise ValueError(
                "invalid adx type filter parameters: "
                f"unknown source {params.source}") from None

        q = params.q
        if q == 0:
            q = 32 if source == AdxTypeFilterSource.SMI_MOMENTUM else 2
        r = params.r
        s = params.s

        if q < 1:
            raise ValueError(
                "invalid adx type filter parameters: "
                "q should be greater than 0")
        if r < 1:
            raise ValueError(
                "invalid adx type filter parameters: "
                "r should be greater than 0")
        if s < 1:
            raise ValueError(
                "invalid adx type filter parameters: "
                "s should be greater than 0")

        # Price components are meaningful only for the single-price TSI sources;
        # the other sources use the bar's high/low/close or the single value of
        # the sample (scalar value, quote mid price, trade price).
        uses_components = source in (AdxTypeFilterSource.TSI_MOMENTUM,
                                     AdxTypeFilterSource.TSI_NORMALIZED)
        bc = DEFAULT_BAR_COMPONENT
        qc = DEFAULT_QUOTE_COMPONENT
        tc = DEFAULT_TRADE_COMPONENT
        if uses_components:
            if params.bar_component is not None:
                bc = params.bar_component
            if params.quote_component is not None:
                qc = params.quote_component
            if params.trade_component is not None:
                tc = params.trade_component

        self._bar_func = bar_component_value(bc)
        self._quote_func = quote_component_value(qc)
        self._trade_func = trade_component_value(tc)

        self._source = source
        self._uses_components = uses_components
        self._q = q

        # Rolling windows of the last q values (for the q-bar look-back).
        self._closes: deque = deque(maxlen=q)
        self._highs: deque = deque(maxlen=q)
        self._lows: deque = deque(maxlen=q)

        # Tick-rule state for single-valued samples of the TVI_BALANCE source.
        self._previous = 0.0
        self._has_previous = False

        # The normalized TSI (TSI_NORMALIZED only) replaces the inner EMA.
        self._tsi = None
        if source == AdxTypeFilterSource.TSI_NORMALIZED:
            self._tsi = TrueStrengthIndex(TrueStrengthIndexParams(q=q, r=r, s=1, u=1, ul=1))
            self._inner = _Ema(1)
        else:
            self._inner = _Ema(r)
        self._outer = _Ema(s)

        self._primed = False

        name = _SOURCE_MNEMONICS[source]
        if source == AdxTypeFilterSource.TVI_BALANCE:
            self._mnemonic = f"atf.{name}({r},{s})"
        else:
            self._mnemonic = f"atf.{name}({q},{r},{s}" \
                             f"{component_triple_mnemonic(bc, qc, tc)})"
        self._description = f"ADX-Type Filter {self._mnemonic}"

    def is_primed(self) -> bool:
        return self._primed

    def metadata(self) -> Metadata:
        return build_metadata(
            Identifier.ADX_TYPE_FILTER,
            self._mnemonic,
            self._description,
            [OutputText(self._mnemonic, self._description)],
        )

    def _filter(self, momentum: float) -> float:
        """Inner smooth -> rectify -> outer smooth; NaN momentum propagates."""
        if math.isnan(momentum):
            # The EMAs do not advance until the first finite momentum.
            return math.nan
        self._primed = True
        return self._outer.update(abs(self._inner.update(momentum)))

    def update(self, sample: float) -> float:
        """Update with a single sample value. Returns the ATF value.

        The SMI and DTI sources use the value as the high, the low and the close;
        the TVI source applies the tick rule (balance = value - previous value).
        """
        source = self._source

        if source == AdxTypeFilterSource.TSI_MOMENTUM:
            self._closes.append(sample)
            if len(self._closes) < self._q:
                return math.nan
            # mtm_k = C_k - C_(k-(q-1)); the leftmost element is C_(k-(q-1)).
            return self._filter(sample - self._closes[0])

        if source == AdxTypeFilterSource.TSI_NORMALIZED:
            tsi, _ = self._tsi.update(sample)
            return self._filter(tsi)

        if source == AdxTypeFilterSource.TVI_BALANCE:
            balance = 0.0
            if self._has_previous:
                balance = sample - self._previous
            self._previous = sample
            self._has_previous = True
            return self._filter(balance)

        return self.update_high_low_close(sample, sample, sample)

    def update_high_low_close(self, high: float, low: float, close: float) -> float:
        """Update with a bar's high, low and close. Returns the ATF value.

        The TSI sources use the close only.
        """
        source = self._source

        if source == AdxTypeFilterSource.SMI_MOMENTUM:
            self._highs.append(high)
            self._lows.append(low)
            if len(self._highs) < self._q:
                return math.nan
            # sm = C - 0.5*(HH(q) + LL(q)).
            hh = max(self._highs)
            ll = min(self._lows)
            return self._filter(close - 0.5 * (hh + ll))

        if source == AdxTypeFilterSource.DTI_MOMENTUM:
            self._highs.append(high)
            self._lows.append(low)
            if len(self._highs) < self._q:
                return math.nan
            # HMU - LMD, the leftmost elements are H_(k-(q-1)) and L_(k-(q-1)).
            hmu = max(high - self._highs[0], 0.0)
            lmd = max(self._lows[0] - low, 0.0)
            return self._filter(hmu - lmd)

        if source == AdxTypeFilterSource.TVI_BALANCE:
            # up - down = (C - L) - (H - C) = 2C - H - L.
            return self._filter(2.0 * close - high - low)

        return self.update(close)

    def update_scalar(self, sample: Scalar) -> List[Any]:
        return [Scalar(time=sample.time, value=self.update(sample.value))]

    def update_bar(self, sample: Bar) -> List[Any]:
        if self._uses_components:
            value = self.update(self._bar_func(sample))
        else:
            value = self.update_high_low_close(sample.high, sample.low, sample.close)
        return [Scalar(time=sample.time, value=value)]

    def update_quote(self, sample: Quote) -> List[Any]:
        return [Scalar(time=sample.time, value=self.update(self._quote_func(sample)))]

    def update_trade(self, sample: Trade) -> List[Any]:
        return [Scalar(time=sample.time, value=self.update(self._trade_func(sample)))]
