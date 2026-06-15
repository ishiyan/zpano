"""Cubic Vertex (CVTX) indicator -- Don Mak.

Predicts turning points by fitting a cubic polynomial to the 4 most recent price
points and computing where the two vertices (extrema) occur, relative to the
current bar. A cubic has up to two turning points.

Given four consecutive prices x(n), x(n-1), x(n-2), x(n-3) (most recent first),
compute the cubic coefficients (Eq 7.2a-c):

    c = (x(n) - 3*x(n-1) + 3*x(n-2) - x(n-3)) / 6
    d = (2*x(n) - 5*x(n-1) + 4*x(n-2) - x(n-3)) / 2
    e = (11*x(n) - 18*x(n-1) + 9*x(n-2) - 2*x(n-3)) / 6

The vertex locations are the roots of 3c*t^2 + 2d*t + e = 0:

    t = (-d +/- sqrt(d^2 - 3ce)) / (3c)

  - bars_to_near_turn: root with the smaller absolute value (more imminent turn)
  - bars_to_far_turn:  root with the larger absolute value (more distant turn)

This indicator works best on pre-smoothed prices; the caller is responsible for
smoothing the input before feeding it to this indicator.

Reference:
    Mak, Don K. (2003). The Science of Financial Market Trading. World Scientific.
    Chapter 7, Appendix 5.

Two outputs:
  - bars_to_near_turn: bars to the more imminent turning point.
  - bars_to_far_turn:  bars to the more distant turning point.

Both are NaN during the priming period (first 3 bars) or when no real turning point
exists. bars_to_far_turn is additionally NaN in the parabolic-fallback case (c == 0).
"""

import math
from typing import List, Any

from ...core.indicator import Indicator
from ...core.metadata import Metadata
from ...core.build_metadata import build_metadata, OutputText
from ...core.identifier import Identifier
from ...core.component_triple_mnemonic import component_triple_mnemonic
from ....entities.bar import Bar
from ....entities.quote import Quote
from ....entities.trade import Trade
from ....entities.scalar import Scalar
from ....entities.bar_component import BarComponent, DEFAULT_BAR_COMPONENT, bar_component_value
from ....entities.quote_component import QuoteComponent, DEFAULT_QUOTE_COMPONENT, quote_component_value
from ....entities.trade_component import TradeComponent, DEFAULT_TRADE_COMPONENT, trade_component_value
from .params import CubicVertexParams


class CubicVertex(Indicator):
    """Don Mak's Cubic Vertex (CVTX) indicator."""

    def __init__(self, params: CubicVertexParams) -> None:
        bc = params.bar_component if params.bar_component is not None else DEFAULT_BAR_COMPONENT
        qc = params.quote_component if params.quote_component is not None else DEFAULT_QUOTE_COMPONENT
        tc = params.trade_component if params.trade_component is not None else DEFAULT_TRADE_COMPONENT

        self._bar_func = bar_component_value(bc)
        self._quote_func = quote_component_value(qc)
        self._trade_func = trade_component_value(tc)

        suffix = component_triple_mnemonic(bc, qc, tc)
        self._mnemonic = "cvtx" if suffix == "" else f"cvtx({suffix[2:]})"  # strip leading ", "

        # Ring buffer for the 4 most recent prices.
        self._buffer = [0.0, 0.0, 0.0, 0.0]
        self._index = 0
        self._count = 0

        self._primed = False

    def is_primed(self) -> bool:
        return self._primed

    def metadata(self) -> Metadata:
        desc = f"Cubic vertex {self._mnemonic}"
        return build_metadata(
            Identifier.CUBIC_VERTEX,
            self._mnemonic,
            desc,
            [
                OutputText(f"{self._mnemonic} near", f"{desc} near turn"),
                OutputText(f"{self._mnemonic} far", f"{desc} far turn"),
            ],
        )

    def update(self, sample: float) -> tuple[float, float]:
        """Update with a scalar value. Returns (bars_to_near_turn, bars_to_far_turn)."""
        nan = math.nan

        # Store the price in the ring buffer.
        self._buffer[self._index] = sample
        self._index = (self._index + 1) % 4
        self._count += 1

        if self._count < 4:
            self._primed = False
            return nan, nan

        self._primed = True

        # Extract prices: x[n] (newest), x[n-1], x[n-2], x[n-3] (oldest).
        xn = self._buffer[(self._index - 1) % 4]
        xn1 = self._buffer[(self._index - 2) % 4]
        xn2 = self._buffer[(self._index - 3) % 4]
        xn3 = self._buffer[(self._index - 4) % 4]

        # Cubic polynomial coefficients (Eq 7.2a-c).
        c = (xn - 3.0 * xn1 + 3.0 * xn2 - xn3) / 6.0
        d = (2.0 * xn - 5.0 * xn1 + 4.0 * xn2 - xn3) / 2.0
        e = (11.0 * xn - 18.0 * xn1 + 9.0 * xn2 - 2.0 * xn3) / 6.0

        # Case: c == 0 -- cubic term vanishes, reduces to parabola or line.
        if c == 0.0:
            if d == 0.0:
                return nan, nan
            vertex = -e / (2.0 * d)
            return vertex, nan

        # Full cubic: solve quadratic 3c*t^2 + 2d*t + e = 0.
        disc = d * d - 3.0 * c * e

        if disc < 0.0:
            return nan, nan

        if disc == 0.0:
            vertex = -d / (3.0 * c)
            return vertex, vertex

        sqrt_disc = math.sqrt(disc)
        three_c = 3.0 * c

        t_plus = (-d + sqrt_disc) / three_c
        t_minus = (-d - sqrt_disc) / three_c

        if abs(t_plus) <= abs(t_minus):
            return t_plus, t_minus
        return t_minus, t_plus

    def update_scalar(self, sample: Scalar) -> List[Any]:
        near, far = self.update(sample.value)
        return [
            Scalar(time=sample.time, value=near),
            Scalar(time=sample.time, value=far),
        ]

    def update_bar(self, sample: Bar) -> List[Any]:
        v = self._bar_func(sample)
        return self.update_scalar(Scalar(time=sample.time, value=v))

    def update_quote(self, sample: Quote) -> List[Any]:
        v = self._quote_func(sample)
        return self.update_scalar(Scalar(time=sample.time, value=v))

    def update_trade(self, sample: Trade) -> List[Any]:
        v = self._trade_func(sample)
        return self.update_scalar(Scalar(time=sample.time, value=v))
