"""Balance of Power indicator."""

import math

from ...core.indicator import Indicator
from ...core.line_indicator import LineIndicator
from ...core.metadata import Metadata
from ...core.build_metadata import build_metadata, OutputText
from ...core.identifier import Identifier
from ...core.output import Output
from ....entities.bar import Bar
from ....entities.quote import Quote
from ....entities.trade import Trade
from ....entities.scalar import Scalar
from ....entities.bar_component import DEFAULT_BAR_COMPONENT, bar_component_value
from ....entities.quote_component import DEFAULT_QUOTE_COMPONENT, quote_component_value
from ....entities.trade_component import DEFAULT_TRADE_COMPONENT, trade_component_value
from .params import BalanceOfPowerParams

_EPSILON = 1e-8


class BalanceOfPower(Indicator):
    """Igor Livshin's Balance of Power (BOP).

    BOP = (Close - Open) / (High - Low)

    When the range (High - Low) is less than epsilon, the value is 0.
    """

    def __init__(self, _params: BalanceOfPowerParams) -> None:
        mnemonic = "bop"
        description = "Balance of Power"

        bar_func = bar_component_value(DEFAULT_BAR_COMPONENT)
        quote_func = quote_component_value(DEFAULT_QUOTE_COMPONENT)
        trade_func = trade_component_value(DEFAULT_TRADE_COMPONENT)

        self._line = LineIndicator(mnemonic, description, bar_func, quote_func, trade_func, self.update)

        self._value: float = math.nan

    def is_primed(self) -> bool:
        """Balance of Power is always primed."""
        return True

    def metadata(self) -> Metadata:
        return build_metadata(
            Identifier.BALANCE_OF_POWER,
            self._line.mnemonic,
            self._line.description,
            [OutputText(self._line.mnemonic, self._line.description)],
        )

    def update(self, sample: float) -> float:
        """Scalar update: O=H=L=C so BOP is always 0."""
        if math.isnan(sample):
            return math.nan
        return self.update_ohlc(sample, sample, sample, sample)

    def update_ohlc(self, open_: float, high: float, low: float, close: float) -> float:
        """Update with OHLC values."""
        if math.isnan(open_) or math.isnan(high) or math.isnan(low) or math.isnan(close):
            return math.nan

        r = high - low
        if r < _EPSILON:
            self._value = 0.0
        else:
            self._value = (close - open_) / r

        return self._value

    def update_scalar(self, sample: Scalar) -> Output:
        return self._line.update_scalar(sample)

    def update_bar(self, sample: Bar) -> Output:
        """Shadows LineIndicator.update_bar to extract OHLC from the bar."""
        value = self.update_ohlc(sample.open, sample.high, sample.low, sample.close)
        return [Scalar(time=sample.time, value=value)]

    def update_quote(self, sample: Quote) -> Output:
        return self._line.update_quote(sample)

    def update_trade(self, sample: Trade) -> Output:
        return self._line.update_trade(sample)
