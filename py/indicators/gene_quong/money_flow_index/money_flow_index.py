"""Money Flow Index indicator."""

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
from ....entities.bar_component import BarComponent, bar_component_value
from ....entities.quote_component import QuoteComponent, DEFAULT_QUOTE_COMPONENT, quote_component_value
from ....entities.trade_component import TradeComponent, DEFAULT_TRADE_COMPONENT, trade_component_value
from .params import MoneyFlowIndexParams


class MoneyFlowIndex(Indicator):
    """Gene Quong's Money Flow Index (MFI).

    MFI = 100 * PositiveMoneyFlow / (PositiveMoneyFlow + NegativeMoneyFlow)

    Default bar component is BarTypicalPrice (not BarClosePrice).
    """

    def __init__(self, params: MoneyFlowIndexParams) -> None:
        length = params.length
        if length < 1:
            raise ValueError(
                "invalid money flow index parameters: length should be greater than 0")

        # Default bar component for MFI is BarTypicalPrice, not BarClosePrice.
        bc = params.bar_component if params.bar_component is not None else BarComponent.TYPICAL
        qc = params.quote_component if params.quote_component is not None else DEFAULT_QUOTE_COMPONENT
        tc = params.trade_component if params.trade_component is not None else DEFAULT_TRADE_COMPONENT

        self._bar_func = bar_component_value(bc)
        quote_func = quote_component_value(qc)
        trade_func = trade_component_value(tc)

        mnemonic = f"mfi({length}{component_triple_mnemonic(bc, qc, tc)})"
        description = f"Money Flow Index {mnemonic}"

        # LineIndicator's update uses volume=1 (scalar-only path).
        self._line = LineIndicator(mnemonic, description, self._bar_func, quote_func, trade_func, self.update)

        self._length: int = length
        self._negative_buffer: list[float] = [0.0] * length
        self._positive_buffer: list[float] = [0.0] * length
        self._negative_sum: float = 0.0
        self._positive_sum: float = 0.0
        self._previous_sample: float = 0.0
        self._buffer_index: int = 0
        self._buffer_low_index: int = 0
        self._buffer_count: int = 0
        self._value: float = math.nan
        self._primed: bool = False

    def is_primed(self) -> bool:
        return self._primed

    def metadata(self) -> Metadata:
        return build_metadata(
            Identifier.MONEY_FLOW_INDEX,
            self._line.mnemonic,
            self._line.description,
            [OutputText(self._line.mnemonic, self._line.description)],
        )

    def update(self, sample: float) -> float:
        """Update with volume = 1 (scalar/quote/trade path)."""
        return self.update_with_volume(sample, 1.0)

    def update_with_volume(self, sample: float, volume: float) -> float:
        """Update with the given sample and volume."""
        if math.isnan(sample) or math.isnan(volume):
            return math.nan

        length_min_one = self._length - 1

        if self._primed:
            self._negative_sum -= self._negative_buffer[self._buffer_low_index]
            self._positive_sum -= self._positive_buffer[self._buffer_low_index]

            amount = sample * volume
            diff = sample - self._previous_sample

            if diff < 0:
                self._negative_buffer[self._buffer_index] = amount
                self._positive_buffer[self._buffer_index] = 0.0
                self._negative_sum += amount
            elif diff > 0:
                self._negative_buffer[self._buffer_index] = 0.0
                self._positive_buffer[self._buffer_index] = amount
                self._positive_sum += amount
            else:
                self._negative_buffer[self._buffer_index] = 0.0
                self._positive_buffer[self._buffer_index] = 0.0

            total = self._positive_sum + self._negative_sum
            if total < 1:
                self._value = 0.0
            else:
                self._value = 100.0 * self._positive_sum / total

            self._buffer_index += 1
            if self._buffer_index > length_min_one:
                self._buffer_index = 0

            self._buffer_low_index += 1
            if self._buffer_low_index > length_min_one:
                self._buffer_low_index = 0

        elif self._buffer_count == 0:
            self._buffer_count += 1
        else:
            amount = sample * volume
            diff = sample - self._previous_sample

            if diff < 0:
                self._negative_buffer[self._buffer_index] = amount
                self._positive_buffer[self._buffer_index] = 0.0
                self._negative_sum += amount
            elif diff > 0:
                self._negative_buffer[self._buffer_index] = 0.0
                self._positive_buffer[self._buffer_index] = amount
                self._positive_sum += amount
            else:
                self._negative_buffer[self._buffer_index] = 0.0
                self._positive_buffer[self._buffer_index] = 0.0

            if self._length == self._buffer_count:
                total = self._positive_sum + self._negative_sum
                if total < 1:
                    self._value = 0.0
                else:
                    self._value = 100.0 * self._positive_sum / total

                self._primed = True

            self._buffer_index += 1
            if self._buffer_index > length_min_one:
                self._buffer_index = 0

            self._buffer_count += 1

        self._previous_sample = sample

        return self._value

    def update_scalar(self, sample: Scalar) -> Output:
        return self._line.update_scalar(sample)

    def update_bar(self, sample: Bar) -> Output:
        """Shadows LineIndicator.update_bar to use bar volume."""
        price = self._bar_func(sample)
        value = self.update_with_volume(price, sample.volume)
        return [Scalar(time=sample.time, value=value)]

    def update_quote(self, sample: Quote) -> Output:
        return self._line.update_quote(sample)

    def update_trade(self, sample: Trade) -> Output:
        return self._line.update_trade(sample)
