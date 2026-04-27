"""Variance indicator."""

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
from .params import VarianceParams


class Variance(Indicator):
    """Computes the variance of samples within a moving window of length l.

    Population variance: sigma^2 = (sum(xi^2) - (sum(xi))^2/l) / l
    Sample variance:     sigma^2 = (sum(xi^2) - (sum(xi))^2/l) / (l-1)

    The indicator is not primed during the first l-1 updates.
    """

    def __init__(self, params: VarianceParams) -> None:
        length = params.length
        if length < 2:
            raise ValueError(
                "invalid variance parameters: length should be greater than 1")

        bc = params.bar_component if params.bar_component is not None else DEFAULT_BAR_COMPONENT
        qc = params.quote_component if params.quote_component is not None else DEFAULT_QUOTE_COMPONENT
        tc = params.trade_component if params.trade_component is not None else DEFAULT_TRADE_COMPONENT

        bar_func = bar_component_value(bc)
        quote_func = quote_component_value(qc)
        trade_func = trade_component_value(tc)

        c = 's' if params.is_unbiased else 'p'
        mnemonic = f"var.{c}({length}{component_triple_mnemonic(bc, qc, tc)})"

        if params.is_unbiased:
            description = f"Unbiased estimation of the sample variance {mnemonic}"
        else:
            description = f"Estimation of the population variance {mnemonic}"

        self._line = LineIndicator(mnemonic, description, bar_func, quote_func, trade_func, self.update)

        self._window: list[float] = [0.0] * length
        self._window_sum: float = 0.0
        self._window_squared_sum: float = 0.0
        self._window_length: int = length
        self._window_count: int = 0
        self._last_index: int = length - 1
        self._primed: bool = False
        self._unbiased: bool = params.is_unbiased

    def is_primed(self) -> bool:
        return self._primed

    def metadata(self) -> Metadata:
        return build_metadata(
            Identifier.VARIANCE,
            self._line.mnemonic,
            self._line.description,
            [OutputText(self._line.mnemonic, self._line.description)],
        )

    def update(self, sample: float) -> float:
        if math.isnan(sample):
            return sample

        temp = sample
        wlen = float(self._window_length)

        if self._primed:
            self._window_sum += temp
            temp *= temp
            self._window_squared_sum += temp
            temp = self._window[0]
            self._window_sum -= temp
            temp *= temp
            self._window_squared_sum -= temp

            if self._unbiased:
                temp = self._window_sum
                temp *= temp
                temp /= wlen
                value = self._window_squared_sum - temp
                value /= float(self._last_index)
            else:
                temp = self._window_sum / wlen
                temp *= temp
                value = self._window_squared_sum / wlen - temp

            for i in range(self._last_index):
                self._window[i] = self._window[i + 1]

            self._window[self._last_index] = sample
        else:
            self._window_sum += temp
            self._window[self._window_count] = temp
            temp *= temp
            self._window_squared_sum += temp

            self._window_count += 1
            if self._window_length == self._window_count:
                self._primed = True
                if self._unbiased:
                    temp = self._window_sum
                    temp *= temp
                    temp /= wlen
                    value = self._window_squared_sum - temp
                    value /= float(self._last_index)
                else:
                    temp = self._window_sum / wlen
                    temp *= temp
                    value = self._window_squared_sum / wlen - temp
            else:
                return math.nan

        return value

    def update_scalar(self, sample: Scalar) -> Output:
        return self._line.update_scalar(sample)

    def update_bar(self, sample: Bar) -> Output:
        return self._line.update_bar(sample)

    def update_quote(self, sample: Quote) -> Output:
        return self._line.update_quote(sample)

    def update_trade(self, sample: Trade) -> Output:
        return self._line.update_trade(sample)
