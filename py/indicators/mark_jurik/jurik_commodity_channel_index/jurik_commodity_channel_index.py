"""Jurik commodity channel index indicator."""

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
from ..jurik_moving_average.jurik_moving_average import JurikMovingAverage
from ..jurik_moving_average.params import JurikMovingAverageParams
from .params import JurikCommodityChannelIndexParams


class JurikCommodityChannelIndex(Indicator):
    """Computes the Jurik Commodity Channel Index (JCCX).

    Uses fast JMA(4) and slow JMA(length), normalizes their difference
    by 1.5x MAD of the difference series.
    """

    def __init__(self, params: JurikCommodityChannelIndexParams) -> None:
        length = params.length

        if length < 2:
            raise ValueError(
                "invalid jurik commodity channel index parameters: "
                "length must be >= 2")

        bc = params.bar_component if params.bar_component is not None \
            else DEFAULT_BAR_COMPONENT
        qc = params.quote_component if params.quote_component is not None \
            else DEFAULT_QUOTE_COMPONENT
        tc = params.trade_component if params.trade_component is not None \
            else DEFAULT_TRADE_COMPONENT

        bar_func = bar_component_value(bc)
        quote_func = quote_component_value(qc)
        trade_func = trade_component_value(tc)

        mnemonic = f"jccx({length}{component_triple_mnemonic(bc, qc, tc)})"
        description = f"Jurik commodity channel index {mnemonic}"

        self._line = LineIndicator(
            mnemonic, description, bar_func, quote_func, trade_func, self.update)

        self._fast_jma = JurikMovingAverage(
            JurikMovingAverageParams(length=4, phase=0))
        self._slow_jma = JurikMovingAverage(
            JurikMovingAverageParams(length=length, phase=0))

        self._diff_buffer_size = 3 * length
        self._diff_buffer: list[float] = []
        self._primed = False

    def is_primed(self) -> bool:
        return self._primed

    def metadata(self) -> Metadata:
        return build_metadata(
            Identifier.JURIK_COMMODITY_CHANNEL_INDEX,
            self._line.mnemonic,
            self._line.description,
            [OutputText(
                mnemonic=self._line.mnemonic,
                description=self._line.description)],
        )

    def update(self, sample: float) -> float:
        """Update the indicator with a new sample value."""
        if math.isnan(sample):
            return sample

        fast_val = self._fast_jma.update(sample)
        slow_val = self._slow_jma.update(sample)

        if math.isnan(fast_val) or math.isnan(slow_val):
            return math.nan

        diff = fast_val - slow_val

        self._diff_buffer.append(diff)
        if len(self._diff_buffer) > self._diff_buffer_size:
            self._diff_buffer.pop(0)

        self._primed = True

        # Compute MAD (mean absolute deviation)
        n = len(self._diff_buffer)
        mad = sum(abs(d) for d in self._diff_buffer) / n

        if mad < 0.00001:
            return 0.0

        return diff / (1.5 * mad)

    def update_bar(self, bar: Bar) -> Output:
        return self._line.update_bar(bar)

    def update_quote(self, quote: Quote) -> Output:
        return self._line.update_quote(quote)

    def update_trade(self, trade: Trade) -> Output:
        return self._line.update_trade(trade)

    def update_scalar(self, scalar: Scalar) -> Output:
        return self._line.update_scalar(scalar)
