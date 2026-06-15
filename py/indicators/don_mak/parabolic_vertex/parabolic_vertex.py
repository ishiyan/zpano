"""Parabolic Vertex (PVTX) indicator -- Don Mak.

Predicts turning points by fitting a parabola to the 3 most recent price points
and computing where the vertex (extremum) occurs, relative to the current bar.

Given three consecutive prices x(n), x(n-1), x(n-2) (most recent first) fitted to
the parabola x(t) = d*t^2 + e*t + f at t = 0, -1, -2, the vertex is at:

    t_v = -(1.5*x(n) - 2*x(n-1) + 0.5*x(n-2)) / (x(n) - 2*x(n-1) + x(n-2))

The output is the number of bars from the current bar to the predicted turning
point (positive = future, negative = past, near zero = now).

This indicator works best on pre-smoothed prices; the caller is responsible for
smoothing the input before feeding it to this indicator.

Reference:
    Mak, Don K. (2003). The Science of Financial Market Trading. World Scientific.
    Chapter 7, Appendix 5.

Single output:
  - value: bars from the current bar to the predicted turning point. NaN during the
    priming period (first 2 bars) or when the three points are collinear (zero
    curvature).
"""

import math

from ...core.indicator import Indicator
from ...core.line_indicator import LineIndicator
from ...core.metadata import Metadata
from ...core.build_metadata import build_metadata, OutputText
from ...core.identifier import Identifier
from ...core.output import Output
from ...core.component_triple_mnemonic import component_triple_mnemonic
from ....entities.bar import Bar
from ....entities.quote import Quote
from ....entities.trade import Trade
from ....entities.scalar import Scalar
from ....entities.bar_component import BarComponent, DEFAULT_BAR_COMPONENT, bar_component_value
from ....entities.quote_component import QuoteComponent, DEFAULT_QUOTE_COMPONENT, quote_component_value
from ....entities.trade_component import TradeComponent, DEFAULT_TRADE_COMPONENT, trade_component_value
from .params import ParabolicVertexParams


class ParabolicVertex(Indicator):
    """Don Mak's Parabolic Vertex (PVTX) indicator."""

    def __init__(self, params: ParabolicVertexParams) -> None:
        bc = params.bar_component if params.bar_component is not None else DEFAULT_BAR_COMPONENT
        qc = params.quote_component if params.quote_component is not None else DEFAULT_QUOTE_COMPONENT
        tc = params.trade_component if params.trade_component is not None else DEFAULT_TRADE_COMPONENT

        bar_func = bar_component_value(bc)
        quote_func = quote_component_value(qc)
        trade_func = trade_component_value(tc)

        suffix = component_triple_mnemonic(bc, qc, tc)
        mnemonic = "pvtx" if suffix == "" else f"pvtx({suffix[2:]})"  # strip leading ", "
        description = f"Parabolic vertex {mnemonic}"

        self._line = LineIndicator(mnemonic, description, bar_func, quote_func, trade_func, self.update)

        # Ring buffer for the 3 most recent prices.
        self._buffer = [0.0, 0.0, 0.0]
        self._index = 0
        self._count = 0

        self._primed = False

    def is_primed(self) -> bool:
        return self._primed

    def metadata(self) -> Metadata:
        return build_metadata(
            Identifier.PARABOLIC_VERTEX,
            self._line.mnemonic,
            self._line.description,
            [OutputText(self._line.mnemonic, self._line.description)],
        )

    def update(self, sample: float) -> float:
        # Store the price in the ring buffer.
        self._buffer[self._index] = sample
        self._index = (self._index + 1) % 3
        self._count += 1

        if self._count < 3:
            self._primed = False
            return math.nan

        self._primed = True

        # Extract prices: x[n] (newest), x[n-1], x[n-2] (oldest).
        xn = self._buffer[(self._index - 1) % 3]
        xn1 = self._buffer[(self._index - 2) % 3]
        xn2 = self._buffer[(self._index - 3) % 3]

        # Denominator = second-order finite difference (proportional to curvature).
        denom = xn - 2.0 * xn1 + xn2
        if denom == 0.0:
            return math.nan

        numer = 1.5 * xn - 2.0 * xn1 + 0.5 * xn2

        return -numer / denom

    def update_scalar(self, sample: Scalar) -> Output:
        return self._line.update_scalar(sample)

    def update_bar(self, sample: Bar) -> Output:
        return self._line.update_bar(sample)

    def update_quote(self, sample: Quote) -> Output:
        return self._line.update_quote(sample)

    def update_trade(self, sample: Trade) -> Output:
        return self._line.update_trade(sample)
