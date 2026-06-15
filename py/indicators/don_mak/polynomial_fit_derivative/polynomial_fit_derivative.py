"""Polynomial Fit Derivative (PFD) indicator -- Don Mak.

Fits a polynomial of degree ``degree`` to the most recent ``degree + 1`` price
bars and evaluates its ``order``-th derivative at the current bar. This is a FIR
filter: a dot product of fixed Lagrange-interpolation-derived coefficients with
the last ``degree + 1`` (optionally EMA-smoothed) prices.

degree=2 -> Parabolic, 3 -> Cubic, 4 -> Quartic, 5 -> Quintic, 6 -> Sextic
velocity (order=1) / acceleration (order=2) indicators.

Reference:
    Mak, Don K. (2003). The Science of Financial Market Trading. Ch 6.
    Mak, Don K. (2006). Mathematical Techniques in Financial Market Trading. Ch 8.

Single output:
  - value: the order-th derivative of the polynomial fit at the current bar,
    NaN until ``degree + 1`` smoothed prices are available.
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
from .params import PolynomialFitDerivativeParams


def _compute_coefficients(degree: int, order: int) -> list[float]:
    """Compute the FIR filter coefficients for the order-th derivative of a
    degree-``degree`` polynomial fit, evaluated at the most recent point.

    Uses the Lagrange basis with the elementary-symmetric-polynomial identity:
        c_i = order! * e_{degree-order}(others) / prod_{j != i} (j - i)
    where ``others`` is the set of point positions {0..degree} excluding i.
    """
    n_points = degree + 1

    factorial_order = 1
    for f in range(2, order + 1):
        factorial_order *= f

    coefficients: list[float] = []

    for i in range(n_points):
        denom = 1.0
        for j in range(n_points):
            if j != i:
                denom *= float(j - i)

        others = [j for j in range(n_points) if j != i]
        m = len(others)  # equals degree

        # Elementary symmetric polynomials e[0..m] of the values in ``others``.
        e = [0.0] * (m + 1)
        e[0] = 1.0
        for v in others:
            for k in range(m, 0, -1):
                e[k] += float(v) * e[k - 1]

        numerator = float(factorial_order) * e[m - order]
        coefficients.append(numerator / denom)

    return coefficients


class PolynomialFitDerivative(Indicator):
    """Don Mak's Polynomial Fit Derivative (PFD) indicator."""

    def __init__(self, params: PolynomialFitDerivativeParams) -> None:
        degree = params.degree
        order = params.order
        smoothing = params.smoothing

        if degree < 2:
            raise ValueError(
                "invalid polynomial fit derivative parameters: degree must be >= 2")
        if order < 1 or order > degree:
            raise ValueError(
                "invalid polynomial fit derivative parameters: "
                "order must be >= 1 and <= degree")
        if smoothing < 0:
            raise ValueError(
                "invalid polynomial fit derivative parameters: smoothing must be >= 0")

        bc = params.bar_component if params.bar_component is not None else DEFAULT_BAR_COMPONENT
        qc = params.quote_component if params.quote_component is not None else DEFAULT_QUOTE_COMPONENT
        tc = params.trade_component if params.trade_component is not None else DEFAULT_TRADE_COMPONENT

        bar_func = bar_component_value(bc)
        quote_func = quote_component_value(qc)
        trade_func = trade_component_value(tc)

        mnemonic = f"pfd({degree},{order},{smoothing}{component_triple_mnemonic(bc, qc, tc)})"
        description = f"Polynomial fit derivative {mnemonic}"

        self._line = LineIndicator(mnemonic, description, bar_func, quote_func, trade_func, self.update)

        self._smoothing = smoothing
        self._n_points = degree + 1
        self._coefficients = _compute_coefficients(degree, order)

        # EMA state.
        self._ema_alpha = 2.0 / (smoothing + 1.0) if smoothing > 0 else 0.0
        self._ema_value = 0.0
        self._ema_initialized = False

        # Ring buffer of the last (degree + 1) smoothed prices.
        self._buf = [0.0] * self._n_points
        self._buf_pos = 0
        self._buf_count = 0

        self._primed = False

    def is_primed(self) -> bool:
        return self._primed

    def metadata(self) -> Metadata:
        return build_metadata(
            Identifier.POLYNOMIAL_FIT_DERIVATIVE,
            self._line.mnemonic,
            self._line.description,
            [OutputText(self._line.mnemonic, self._line.description)],
        )

    def update(self, sample: float) -> float:
        # Step 1: optional EMA smoothing.
        if self._smoothing > 0:
            if not self._ema_initialized:
                self._ema_value = sample
                self._ema_initialized = True
            else:
                self._ema_value = self._ema_alpha * sample + (1.0 - self._ema_alpha) * self._ema_value
            smoothed = self._ema_value
        else:
            smoothed = sample

        # Step 2: push into the ring buffer.
        self._buf[self._buf_pos] = smoothed
        self._buf_pos = (self._buf_pos + 1) % self._n_points
        self._buf_count += 1

        # Step 3: not enough data yet.
        if self._buf_count < self._n_points:
            self._primed = False
            return math.nan

        # Step 4: FIR dot product (coefficients[j] multiplies the j-th most recent).
        result = 0.0
        for j in range(self._n_points):
            buf_idx = (self._buf_pos - 1 - j) % self._n_points
            result += self._coefficients[j] * self._buf[buf_idx]

        self._primed = True
        return result

    def update_scalar(self, sample: Scalar) -> Output:
        return self._line.update_scalar(sample)

    def update_bar(self, sample: Bar) -> Output:
        return self._line.update_bar(sample)

    def update_quote(self, sample: Quote) -> Output:
        return self._line.update_quote(sample)

    def update_trade(self, sample: Trade) -> Output:
        return self._line.update_trade(sample)
