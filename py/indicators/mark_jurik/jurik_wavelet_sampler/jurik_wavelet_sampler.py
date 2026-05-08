"""Jurik wavelet sampler indicator."""

import math
from collections import deque

from ...core.indicator import Indicator
from ...core.line_indicator import LineIndicator
from ...core.metadata import Metadata
from ...core.build_metadata import build_metadata, OutputText
from ...core.identifier import Identifier
from ...core.component_triple_mnemonic import component_triple_mnemonic
from ...core.output import Output
from ....entities.bar import Bar
from ....entities.quote import Quote
from ....entities.trade import Trade
from ....entities.scalar import Scalar
from ....entities.bar_component import BarComponent, DEFAULT_BAR_COMPONENT, bar_component_value
from ....entities.quote_component import QuoteComponent, DEFAULT_QUOTE_COMPONENT, quote_component_value
from ....entities.trade_component import TradeComponent, DEFAULT_TRADE_COMPONENT, trade_component_value
from .params import JurikWaveletSamplerParams


NM_TABLE = [
    (1, 0), (2, 0), (3, 0), (4, 0), (5, 0),
    (7, 2), (10, 2), (14, 4), (19, 4), (26, 8),
    (35, 8), (48, 16), (65, 16), (90, 32), (123, 32),
    (172, 64), (237, 64), (334, 128),
]


class JurikWaveletSampler(Indicator):
    """Computes the Jurik wavelet sampler.

    Produces `index` output columns per bar, each representing a different
    multi-resolution scale. Each column c uses parameters (n, M) from NM_TABLE[c].
    If M == 0, output is a simple lag of n bars.
    If M > 0, output is the mean of (M+1) prices centered at lag n.
    """

    def __init__(self, params: JurikWaveletSamplerParams) -> None:
        index = params.index

        if index < 1 or index > 18:
            raise ValueError(
                "invalid jurik wavelet sampler parameters: index must be in range [1, 18]")

        bc = params.bar_component if params.bar_component is not None else DEFAULT_BAR_COMPONENT
        qc = params.quote_component if params.quote_component is not None else DEFAULT_QUOTE_COMPONENT
        tc = params.trade_component if params.trade_component is not None else DEFAULT_TRADE_COMPONENT

        bar_func = bar_component_value(bc)
        quote_func = quote_component_value(qc)
        trade_func = trade_component_value(tc)

        mnemonic = f"jwav({index}{component_triple_mnemonic(bc, qc, tc)})"
        description = f"Jurik wavelet sampler {mnemonic}"

        self._mnemonic = mnemonic
        self._description = description
        self._bar_func = bar_func
        self._quote_func = quote_func
        self._trade_func = trade_func
        self._index = index

        # Compute buffer size needed: max(n + M//2) across all used columns
        max_lookback = 0
        for c in range(index):
            n, m = NM_TABLE[c]
            lookback = n + m // 2
            if lookback > max_lookback:
                max_lookback = lookback

        self._max_lookback = max_lookback
        self._prices: list[float] = []
        self._bar_count = 0
        self._columns: list[float] = [math.nan] * index
        self._primed = False

    def is_primed(self) -> bool:
        return self._primed

    def metadata(self) -> Metadata:
        return build_metadata(
            Identifier.JURIK_WAVELET_SAMPLER,
            self._mnemonic,
            self._description,
            [OutputText(mnemonic=self._mnemonic, description=self._description)],
        )

    @property
    def columns(self) -> list[float]:
        """Return the current column values after the last update."""
        return list(self._columns)

    def _update(self, sample: float) -> list[float]:
        """Core update logic. Appends price and computes all columns."""
        self._prices.append(sample)
        self._bar_count += 1

        results = []
        all_valid = True
        for c in range(self._index):
            n, m = NM_TABLE[c]
            dead_zone = n + m // 2  # bars needed before first output (0-indexed: need bar_count > dead_zone)

            if self._bar_count <= dead_zone:
                results.append(math.nan)
                all_valid = False
            else:
                if m == 0:
                    # Simple lag: price at (bar_count - 1 - n)
                    val = self._prices[self._bar_count - 1 - n]
                else:
                    # Mean of (M+1) prices centered at lag n
                    half = m // 2
                    center_idx = self._bar_count - 1 - n
                    total = 0.0
                    for k in range(center_idx - half, center_idx + half + 1):
                        total += self._prices[k]
                    val = total / (m + 1)
                results.append(val)

        self._columns = results
        if all_valid:
            self._primed = True
        return results

    def update_scalar(self, scalar: Scalar) -> Output:
        """Updates the indicator given the next scalar sample."""
        values = self._update(scalar.value)
        return [Scalar(scalar.time, values[0])]

    def update_bar(self, bar: Bar) -> Output:
        """Updates the indicator given the next bar sample."""
        return self.update_scalar(Scalar(bar.time, self._bar_func(bar)))

    def update_quote(self, quote: Quote) -> Output:
        """Updates the indicator given the next quote sample."""
        return self.update_scalar(Scalar(quote.time, self._quote_func(quote)))

    def update_trade(self, trade: Trade) -> Output:
        """Updates the indicator given the next trade sample."""
        return self.update_scalar(Scalar(trade.time, self._trade_func(trade)))
