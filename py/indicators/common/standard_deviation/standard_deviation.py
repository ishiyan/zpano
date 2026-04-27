"""Standard deviation indicator."""

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
from ..variance.variance import Variance
from ..variance.params import VarianceParams
from .params import StandardDeviationParams


class StandardDeviation(Indicator):
    """Computes the standard deviation as the square root of variance.

    The indicator is not primed during the first l-1 updates.
    """

    def __init__(self, params: StandardDeviationParams) -> None:
        bc = params.bar_component if params.bar_component is not None else DEFAULT_BAR_COMPONENT
        qc = params.quote_component if params.quote_component is not None else DEFAULT_QUOTE_COMPONENT
        tc = params.trade_component if params.trade_component is not None else DEFAULT_TRADE_COMPONENT

        bar_func = bar_component_value(bc)
        quote_func = quote_component_value(qc)
        trade_func = trade_component_value(tc)

        # Create underlying variance indicator.
        vp = VarianceParams(
            length=params.length,
            is_unbiased=params.is_unbiased,
            bar_component=params.bar_component,
            quote_component=params.quote_component,
            trade_component=params.trade_component,
        )
        self._variance = Variance(vp)

        c = 's' if params.is_unbiased else 'p'
        mnemonic = f"stdev.{c}({params.length}{component_triple_mnemonic(bc, qc, tc)})"

        if params.is_unbiased:
            description = f"Standard deviation based on unbiased estimation of the sample variance {mnemonic}"
        else:
            description = f"Standard deviation based on estimation of the population variance {mnemonic}"

        self._line = LineIndicator(mnemonic, description, bar_func, quote_func, trade_func, self.update)

    def is_primed(self) -> bool:
        return self._variance.is_primed()

    def metadata(self) -> Metadata:
        return build_metadata(
            Identifier.STANDARD_DEVIATION,
            self._line.mnemonic,
            self._line.description,
            [OutputText(self._line.mnemonic, self._line.description)],
        )

    def update(self, sample: float) -> float:
        v = self._variance.update(sample)
        if math.isnan(v):
            return v
        return math.sqrt(v)

    def update_scalar(self, sample: Scalar) -> Output:
        return self._line.update_scalar(sample)

    def update_bar(self, sample: Bar) -> Output:
        return self._line.update_bar(sample)

    def update_quote(self, sample: Quote) -> Output:
        return self._line.update_quote(sample)

    def update_trade(self, sample: Trade) -> Output:
        return self._line.update_trade(sample)
