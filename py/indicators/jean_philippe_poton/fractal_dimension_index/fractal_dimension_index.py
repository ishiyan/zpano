"""Fractal Dimension Index indicator."""

import math

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
from .params import FractalDimensionIndexParams


class FractalDimensionIndex(Indicator):
    """Computes the Fractal Dimension Index (FDI).

    Measures the fractal dimension of a price time series using normalized
    path length. Values near 1.5 indicate a random market, near 1.0 a
    trending market, and near 2.0 a highly volatile market.

    The indicator is not primed during the first `period` updates.
    """

    def __init__(self, params: FractalDimensionIndexParams) -> None:
        period = params.period
        if period < 2:
            raise ValueError(
                "invalid fractal dimension parameters: period should be greater than 1")

        bc = params.bar_component if params.bar_component is not None else DEFAULT_BAR_COMPONENT
        qc = params.quote_component if params.quote_component is not None else DEFAULT_QUOTE_COMPONENT
        tc = params.trade_component if params.trade_component is not None else DEFAULT_TRADE_COMPONENT

        bar_func = bar_component_value(bc)
        quote_func = quote_component_value(qc)
        trade_func = trade_component_value(tc)

        mnemonic = f"fdi({period}{component_triple_mnemonic(bc, qc, tc)})"
        description = f"Fractal dimension index {mnemonic}"

        self._line = LineIndicator(mnemonic, description, bar_func, quote_func, trade_func, self.update)

        self._period: int = period
        self._window: list[float] = [0.0] * (period + 1)
        self._window_count: int = 0
        self._primed: bool = False

        # Precompute constants.
        self._log_2n: float = math.log(2.0 * period)
        self._ln2: float = math.log(2.0)
        self._inv_n_sq: float = 1.0 / (period * period)

    def is_primed(self) -> bool:
        return self._primed

    def metadata(self) -> Metadata:
        return build_metadata(
            Identifier.FRACTAL_DIMENSION_INDEX,
            self._line.mnemonic,
            self._line.description,
            [OutputText(self._line.mnemonic, self._line.description)],
        )

    def update(self, sample: float) -> float:
        if math.isnan(sample):
            return sample

        period = self._period

        if self._primed:
            # Shift window left by one.
            for i in range(period):
                self._window[i] = self._window[i + 1]
            self._window[period] = sample
        else:
            self._window[self._window_count] = sample
            self._window_count += 1

            if self._window_count <= period:
                return math.nan

            self._primed = True

        # Find min/max for normalization.
        price_max = self._window[0]
        price_min = self._window[0]
        for k in range(1, period + 1):
            if self._window[k] > price_max:
                price_max = self._window[k]
            if self._window[k] < price_min:
                price_min = self._window[k]

        price_range = price_max - price_min

        if price_range < 1e-10:
            return 1.0

        # Normalize and compute path length.
        prior_norm = (self._window[0] - price_min) / price_range
        length = 0.0

        for k in range(1, period + 1):
            curr_norm = (self._window[k] - price_min) / price_range
            diff = curr_norm - prior_norm
            length += math.sqrt(diff * diff + self._inv_n_sq)
            prior_norm = curr_norm

        # FDI = 1 + (ln(L) + ln(2)) / ln(2N)
        return 1.0 + (math.log(length) + self._ln2) / self._log_2n

    def update_scalar(self, sample: Scalar) -> Output:
        return self._line.update_scalar(sample)

    def update_bar(self, sample: Bar) -> Output:
        return self._line.update_bar(sample)

    def update_quote(self, sample: Quote) -> Output:
        return self._line.update_quote(sample)

    def update_trade(self, sample: Trade) -> Output:
        return self._line.update_trade(sample)
