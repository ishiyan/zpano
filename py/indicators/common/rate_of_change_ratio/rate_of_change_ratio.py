"""Rate of Change Ratio indicator."""

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
from .params import RateOfChangeRatioParams


class RateOfChangeRatio(Indicator):
    """Computes the Rate of Change Ratio (ROCR).

    ROCRi = Pi / Pi-l, where l is the length.
    ROCR100i = (Pi / Pi-l) * 100.

    The indicator is not primed during the first l updates.
    """

    def __init__(self, params: RateOfChangeRatioParams) -> None:
        length = params.length
        if length < 1:
            raise ValueError(
                "invalid rate of change ratio parameters: length should be positive")

        bc = params.bar_component if params.bar_component is not None else DEFAULT_BAR_COMPONENT
        qc = params.quote_component if params.quote_component is not None else DEFAULT_QUOTE_COMPONENT
        tc = params.trade_component if params.trade_component is not None else DEFAULT_TRADE_COMPONENT

        bar_func = bar_component_value(bc)
        quote_func = quote_component_value(qc)
        trade_func = trade_component_value(tc)

        self._hundred_scale = params.hundred_scale

        if params.hundred_scale:
            mnemonic = f"rocr100({length}{component_triple_mnemonic(bc, qc, tc)})"
            description = f"Rate of Change Ratio 100 Scale {mnemonic}"
        else:
            mnemonic = f"rocr({length}{component_triple_mnemonic(bc, qc, tc)})"
            description = f"Rate of Change Ratio {mnemonic}"

        self._line = LineIndicator(mnemonic, description, bar_func, quote_func, trade_func, self.update)

        self._window: list[float] = [0.0] * (length + 1)
        self._window_length: int = length + 1
        self._window_count: int = 0
        self._last_index: int = length
        self._primed: bool = False

    def is_primed(self) -> bool:
        return self._primed

    def metadata(self) -> Metadata:
        return build_metadata(
            Identifier.RATE_OF_CHANGE_RATIO,
            self._line.mnemonic,
            self._line.description,
            [OutputText(self._line.mnemonic, self._line.description)],
        )

    def update(self, sample: float) -> float:
        if math.isnan(sample):
            return sample

        epsilon = 1e-13
        scale = 100.0 if self._hundred_scale else 1.0

        if self._primed:
            for i in range(self._last_index):
                self._window[i] = self._window[i + 1]

            self._window[self._last_index] = sample
            previous = self._window[0]
            if abs(previous) > epsilon:
                return (sample / previous) * scale

            return 0.0

        self._window[self._window_count] = sample
        self._window_count += 1

        if self._window_length == self._window_count:
            self._primed = True
            previous = self._window[0]
            if abs(previous) > epsilon:
                return (sample / previous) * scale

            return 0.0

        return math.nan

    def update_scalar(self, sample: Scalar) -> Output:
        return self._line.update_scalar(sample)

    def update_bar(self, sample: Bar) -> Output:
        return self._line.update_bar(sample)

    def update_quote(self, sample: Quote) -> Output:
        return self._line.update_quote(sample)

    def update_trade(self, sample: Trade) -> Output:
        return self._line.update_trade(sample)
