"""Pearson's correlation coefficient indicator."""

import math

from ...core.line_indicator import LineIndicator
from ...core.metadata import Metadata
from ...core.build_metadata import build_metadata, OutputText
from ...core.identifier import Identifier
from ...core.component_triple_mnemonic import component_triple_mnemonic
from ...core.output import Output
from ....entities.bar import Bar
from ....entities.scalar import Scalar
from ....entities.bar_component import BarComponent, DEFAULT_BAR_COMPONENT, bar_component_value
from ....entities.quote_component import QuoteComponent, DEFAULT_QUOTE_COMPONENT, quote_component_value
from ....entities.trade_component import TradeComponent, DEFAULT_TRADE_COMPONENT, trade_component_value
from .params import PearsonsCorrelationCoefficientParams


class PearsonsCorrelationCoefficient(LineIndicator):
    """Computes Pearson's Correlation Coefficient (r) over a rolling window.

    Given two input series X and Y, it computes:
        r = (n*sumXY - sumX*sumY) / sqrt((n*sumX2 - sumX^2) * (n*sumY2 - sumY^2))

    For single-input updates, X=Y=sample (degenerate case).
    For bar updates, X=High, Y=Low.
    """

    def __init__(self, params: PearsonsCorrelationCoefficientParams) -> None:
        length = params.length
        if length < 1:
            raise ValueError(
                "invalid pearsons correlation coefficient parameters: length should be positive")

        bc = params.bar_component if params.bar_component is not None else DEFAULT_BAR_COMPONENT
        qc = params.quote_component if params.quote_component is not None else DEFAULT_QUOTE_COMPONENT
        tc = params.trade_component if params.trade_component is not None else DEFAULT_TRADE_COMPONENT

        bar_func = bar_component_value(bc)
        quote_func = quote_component_value(qc)
        trade_func = trade_component_value(tc)

        mnemonic = f"correl({length}{component_triple_mnemonic(bc, qc, tc)})"
        description = f"Pearsons Correlation Coefficient {mnemonic}"

        super().__init__(mnemonic, description, bar_func, quote_func, trade_func, self.update)

        self._bar_func_raw = bar_func
        self._length = length
        self._window_x: list[float] = [0.0] * length
        self._window_y: list[float] = [0.0] * length
        self._count: int = 0
        self._pos: int = 0
        self._sum_x: float = 0.0
        self._sum_y: float = 0.0
        self._sum_x2: float = 0.0
        self._sum_y2: float = 0.0
        self._sum_xy: float = 0.0
        self._primed: bool = False

    def is_primed(self) -> bool:
        return self._primed

    def metadata(self) -> Metadata:
        return build_metadata(
            Identifier.PEARSONS_CORRELATION_COEFFICIENT,
            self.mnemonic,
            self.description,
            [OutputText(self.mnemonic, self.description)],
        )

    def update(self, sample: float) -> float:
        """Single-input update: sets X=Y=sample."""
        return self.update_pair(sample, sample)

    def update_pair(self, x: float, y: float) -> float:
        """Update with an (x, y) pair."""
        if math.isnan(x) or math.isnan(y):
            return math.nan

        n = float(self._length)

        if self._primed:
            old_x = self._window_x[self._pos]
            old_y = self._window_y[self._pos]

            self._sum_x -= old_x
            self._sum_y -= old_y
            self._sum_x2 -= old_x * old_x
            self._sum_y2 -= old_y * old_y
            self._sum_xy -= old_x * old_y

            self._window_x[self._pos] = x
            self._window_y[self._pos] = y
            self._pos = (self._pos + 1) % self._length

            self._sum_x += x
            self._sum_y += y
            self._sum_x2 += x * x
            self._sum_y2 += y * y
            self._sum_xy += x * y

            return self._correlate(n)

        # Accumulating phase.
        self._window_x[self._count] = x
        self._window_y[self._count] = y

        self._sum_x += x
        self._sum_y += y
        self._sum_x2 += x * x
        self._sum_y2 += y * y
        self._sum_xy += x * y

        self._count += 1

        if self._count == self._length:
            self._primed = True
            self._pos = 0
            return self._correlate(n)

        return math.nan

    def _correlate(self, n: float) -> float:
        var_x = self._sum_x2 - (self._sum_x * self._sum_x) / n
        var_y = self._sum_y2 - (self._sum_y * self._sum_y) / n
        temp_real = var_x * var_y

        if temp_real <= 0:
            return 0.0

        return (self._sum_xy - (self._sum_x * self._sum_y) / n) / math.sqrt(temp_real)

    def update_bar(self, sample: Bar) -> Output:
        """Custom bar update: X=High, Y=Low."""
        x = sample.high
        y = sample.low
        value = self.update_pair(x, y)
        return [Scalar(sample.time, value)]
