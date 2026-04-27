"""Ehler's two-pole Super Smoother (SS) indicator."""

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
from .params import SuperSmootherParams


class SuperSmoother(Indicator):
    """Ehler's two-pole Super Smoother (SS).

    Given the shortest (lambda) cycle period in bars, the Super Smoother filter
    attenuates cycle periods shorter than this shortest one.

        beta = sqrt(2) * pi / lambda
        alpha = exp(-beta)
        gamma2 = 2 * alpha * cos(beta)
        gamma3 = -alpha^2
        gamma1 = (1 - gamma2 - gamma3) / 2

        SS_i = gamma1 * (x_i + x_{i-1}) + gamma2 * SS_{i-1} + gamma3 * SS_{i-2}
    """

    def __init__(self, coeff1: float, coeff2: float, coeff3: float,
                 mnemonic: str, description: str,
                 bar_func, quote_func, trade_func) -> None:
        self._line = LineIndicator(mnemonic, description, bar_func, quote_func, trade_func, self.update)
        self._coeff1 = coeff1
        self._coeff2 = coeff2
        self._coeff3 = coeff3
        self._count = 0
        self._sample_previous = 0.0
        self._filter_previous = 0.0
        self._filter_previous2 = 0.0
        self._value = math.nan
        self._primed = False

    @staticmethod
    def create(params: SuperSmootherParams) -> 'SuperSmoother':
        """Creates a new SuperSmoother from parameters."""
        return _new_super_smoother(params)

    def is_primed(self) -> bool:
        return self._primed

    def metadata(self) -> Metadata:
        return build_metadata(
            Identifier.SUPER_SMOOTHER,
            self._line.mnemonic,
            self._line.description,
            [OutputText(self._line.mnemonic, self._line.description)],
        )

    def update(self, sample: float) -> float:
        if math.isnan(sample):
            return sample

        if self._primed:
            f = self._coeff1 * (sample + self._sample_previous) + \
                self._coeff2 * self._filter_previous + self._coeff3 * self._filter_previous2
            self._value = f
            self._sample_previous = sample
            self._filter_previous2 = self._filter_previous
            self._filter_previous = f
            return self._value

        self._count += 1

        if self._count == 1:
            self._sample_previous = sample
            self._filter_previous = sample
            self._filter_previous2 = sample

        f = self._coeff1 * (sample + self._sample_previous) + \
            self._coeff2 * self._filter_previous + self._coeff3 * self._filter_previous2

        if self._count == 3:
            self._primed = True
            self._value = f

        self._sample_previous = sample
        self._filter_previous2 = self._filter_previous
        self._filter_previous = f

        return self._value

    def update_scalar(self, sample: Scalar) -> Output:
        return self._line.update_scalar(sample)

    def update_bar(self, sample: Bar) -> Output:
        return self._line.update_bar(sample)

    def update_quote(self, sample: Quote) -> Output:
        return self._line.update_quote(sample)

    def update_trade(self, sample: Trade) -> Output:
        return self._line.update_trade(sample)


def _new_super_smoother(p: SuperSmootherParams) -> SuperSmoother:
    invalid = "invalid super smoother parameters"

    period = p.shortest_cycle_period
    if period < 2:
        raise ValueError(f"{invalid}: shortest cycle period should be greater than 1")

    # Default bar component is MedianPrice, not ClosePrice.
    bc = p.bar_component if p.bar_component is not None else BarComponent.MEDIAN
    qc = p.quote_component if p.quote_component is not None else DEFAULT_QUOTE_COMPONENT
    tc = p.trade_component if p.trade_component is not None else DEFAULT_TRADE_COMPONENT

    bar_func = bar_component_value(bc)
    quote_func = quote_component_value(qc)
    trade_func = trade_component_value(tc)

    # Calculate coefficients.
    beta = math.sqrt(2) * math.pi / period
    alpha = math.exp(-beta)
    gamma2 = 2 * alpha * math.cos(beta)
    gamma3 = -alpha * alpha
    gamma1 = (1 - gamma2 - gamma3) / 2

    mnemonic = f"ss({period}{component_triple_mnemonic(bc, qc, tc)})"
    desc = f"Super Smoother {mnemonic}"

    return SuperSmoother(gamma1, gamma2, gamma3, mnemonic, desc, bar_func, quote_func, trade_func)
