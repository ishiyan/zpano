"""Commodity Channel Index indicator."""

import math
import sys

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
from .params import CommodityChannelIndexParams, DEFAULT_INVERSE_SCALING_FACTOR


class CommodityChannelIndex(Indicator):
    """Donald Lambert's Commodity Channel Index (CCI).

    CCI = scalingFactor * (sample - SMA) / meanDeviation

    where scalingFactor = length / inverseScalingFactor (default 0.015).
    """

    def __init__(self, params: CommodityChannelIndexParams) -> None:
        length = params.length
        if length < 2:
            raise ValueError(
                "invalid commodity channel index parameters: length should be greater than 1")

        inverse_factor = params.inverse_scaling_factor
        if inverse_factor == 0:
            inverse_factor = DEFAULT_INVERSE_SCALING_FACTOR

        bc = params.bar_component if params.bar_component is not None else DEFAULT_BAR_COMPONENT
        qc = params.quote_component if params.quote_component is not None else DEFAULT_QUOTE_COMPONENT
        tc = params.trade_component if params.trade_component is not None else DEFAULT_TRADE_COMPONENT

        bar_func = bar_component_value(bc)
        quote_func = quote_component_value(qc)
        trade_func = trade_component_value(tc)

        mnemonic = f"cci({length}{component_triple_mnemonic(bc, qc, tc)})"
        description = f"Commodity Channel Index {mnemonic}"

        self._line = LineIndicator(mnemonic, description, bar_func, quote_func, trade_func, self.update)

        self._length: int = length
        self._scaling_factor: float = float(length) / inverse_factor
        self._window: list[float] = [0.0] * length
        self._window_count: int = 0
        self._window_sum: float = 0.0
        self._value: float = math.nan
        self._primed: bool = False

    def is_primed(self) -> bool:
        return self._primed

    def metadata(self) -> Metadata:
        return build_metadata(
            Identifier.COMMODITY_CHANNEL_INDEX,
            self._line.mnemonic,
            self._line.description,
            [OutputText(self._line.mnemonic, self._line.description)],
        )

    def update(self, sample: float) -> float:
        if math.isnan(sample):
            return sample

        last_index = self._length - 1

        if self._primed:
            self._window_sum += sample - self._window[0]
            self._window[:-1] = self._window[1:]
            self._window[last_index] = sample

            average = self._window_sum / self._length

            temp = 0.0
            for i in range(self._length):
                temp += abs(self._window[i] - average)

            if abs(temp) < sys.float_info.min:
                self._value = 0.0
            else:
                self._value = self._scaling_factor * (sample - average) / temp
        else:
            self._window_sum += sample
            self._window[self._window_count] = sample
            self._window_count += 1

            if self._window_count == self._length:
                self._primed = True

                average = self._window_sum / self._length

                temp = 0.0
                for i in range(self._length):
                    temp += abs(self._window[i] - average)

                if abs(temp) < sys.float_info.min:
                    self._value = 0.0
                else:
                    self._value = self._scaling_factor * (sample - average) / temp

        return self._value

    def update_scalar(self, sample: Scalar) -> Output:
        return self._line.update_scalar(sample)

    def update_bar(self, sample: Bar) -> Output:
        return self._line.update_bar(sample)

    def update_quote(self, sample: Quote) -> Output:
        return self._line.update_quote(sample)

    def update_trade(self, sample: Trade) -> Output:
        return self._line.update_trade(sample)
