"""Marc Chaikin's Advance-Decline (A/D) Line."""

import math

from .params import AdvanceDeclineParams
from ...core.line_indicator import LineIndicator
from ...core.indicator import Indicator
from ...core.identifier import Identifier
from ...core.metadata import Metadata
from ...core.build_metadata import build_metadata, OutputText
from ...core.output import Output
from ....entities.bar import Bar
from ....entities.quote import Quote
from ....entities.trade import Trade
from ....entities.scalar import Scalar
from ....entities.bar_component import DEFAULT_BAR_COMPONENT, bar_component_value
from ....entities.quote_component import DEFAULT_QUOTE_COMPONENT, quote_component_value
from ....entities.trade_component import DEFAULT_TRADE_COMPONENT, trade_component_value


class AdvanceDecline(Indicator):
    """Marc Chaikin's Advance-Decline (A/D) Line.

    CLV = ((Close - Low) - (High - Close)) / (High - Low)
    AD = AD_prev + CLV * Volume

    When High == Low, AD is unchanged. Always primed after first update.
    """

    def __init__(self, params: AdvanceDeclineParams) -> None:
        mnemonic = "ad"
        description = "Advance-Decline"

        bar_func = bar_component_value(DEFAULT_BAR_COMPONENT)
        quote_func = quote_component_value(DEFAULT_QUOTE_COMPONENT)
        trade_func = trade_component_value(DEFAULT_TRADE_COMPONENT)

        self._line = LineIndicator(mnemonic, description, bar_func, quote_func, trade_func, self.update)
        self._ad = 0.0
        self._value = math.nan
        self._primed = False

    def is_primed(self) -> bool:
        return self._primed

    def metadata(self) -> Metadata:
        return build_metadata(
            Identifier.ADVANCE_DECLINE,
            self._line.mnemonic,
            self._line.description,
            [OutputText(self._line.mnemonic, self._line.description)],
        )

    def update(self, sample: float) -> float:
        """Update with scalar (H=L=C, volume=1, so AD unchanged)."""
        if math.isnan(sample):
            return math.nan
        return self.update_hlcv(sample, sample, sample, 1.0)

    def update_hlcv(self, high: float, low: float, close: float, volume: float) -> float:
        """Update with high, low, close, volume."""
        if math.isnan(high) or math.isnan(low) or math.isnan(close) or math.isnan(volume):
            return math.nan

        temp = high - low
        if temp > 0:
            self._ad += ((close - low) - (high - close)) / temp * volume

        self._value = self._ad
        self._primed = True
        return self._value

    def update_scalar(self, sample: Scalar) -> Output:
        return self._line.update_scalar(sample)

    def update_bar(self, sample: Bar) -> Output:
        """Shadows LineIndicator.update_bar to extract HLCV."""
        value = self.update_hlcv(sample.high, sample.low, sample.close, sample.volume)
        return [Scalar(time=sample.time, value=value)]

    def update_quote(self, sample: Quote) -> Output:
        return self._line.update_quote(sample)

    def update_trade(self, sample: Trade) -> Output:
        return self._line.update_trade(sample)
