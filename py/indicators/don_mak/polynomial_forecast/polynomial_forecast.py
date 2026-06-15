"""Polynomial Forecast (POF) indicator -- Don Mak.

One-step-ahead price forecast using a Taylor series expansion built on polynomial
fit derivatives (PFD):

    velocity     = PFD(price, degree, derivative_order=1)
    acceleration = PFD(price, degree, derivative_order=2)

    order=1:  forecast = price + velocity                     (F1V)
    order=2:  forecast = price + velocity + 0.5*acceleration  (F1VA)

Velocity and acceleration are computed from a local polynomial fit of the given
degree to the most recent ``degree + 1`` (optionally EMA-smoothed) prices.

Reference:
    Mak, Don K. (2003). The Science of Financial Market Trading. World Scientific.
    Chapter 10.2.

Single output:
  - value: the 1-bar-ahead price forecast, NaN until ``degree + 1`` prices are
    available.
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
from .params import PolynomialForecastParams


def _compute_coefficients(degree: int, order: int) -> list[float]:
    """Compute FIR coefficients for the ``order``-th derivative of a degree-d
    polynomial fit evaluated at the most recent point (Lagrange basis)."""
    n_points = degree + 1
    coefficients: list[float] = []

    for i in range(n_points):
        denom = 1.0
        for j in range(n_points):
            if j != i:
                denom *= float(j - i)

        others = [j for j in range(n_points) if j != i]

        numerator = 0.0
        if order == 1:
            for ell_idx in range(len(others)):
                term = 1.0
                for m_idx in range(len(others)):
                    if m_idx != ell_idx:
                        term *= float(others[m_idx])
                numerator += term
        elif order == 2:
            for ell_idx in range(len(others)):
                for r_idx in range(ell_idx + 1, len(others)):
                    term = 2.0
                    for m_idx in range(len(others)):
                        if m_idx != ell_idx and m_idx != r_idx:
                            term *= float(others[m_idx])
                    numerator += term
        else:
            raise ValueError("order must be 1 or 2")

        coefficients.append(numerator / denom)

    return coefficients


class PolynomialForecast(Indicator):
    """Don Mak's Polynomial Forecast (POF) indicator."""

    def __init__(self, params: PolynomialForecastParams) -> None:
        degree = params.degree
        order = params.order
        smoothing = params.smoothing

        if degree < 2:
            raise ValueError("invalid polynomial forecast parameters: degree must be >= 2")
        if order < 1 or order > 2:
            raise ValueError("invalid polynomial forecast parameters: order must be 1 or 2")
        if smoothing < 0:
            raise ValueError("invalid polynomial forecast parameters: smoothing must be >= 0")

        bc = params.bar_component if params.bar_component is not None else DEFAULT_BAR_COMPONENT
        qc = params.quote_component if params.quote_component is not None else DEFAULT_QUOTE_COMPONENT
        tc = params.trade_component if params.trade_component is not None else DEFAULT_TRADE_COMPONENT

        bar_func = bar_component_value(bc)
        quote_func = quote_component_value(qc)
        trade_func = trade_component_value(tc)

        mnemonic = f"pof({degree},{order},{smoothing}{component_triple_mnemonic(bc, qc, tc)})"
        description = f"Polynomial forecast {mnemonic}"

        self._line = LineIndicator(mnemonic, description, bar_func, quote_func, trade_func, self.update)

        self._degree = degree
        self._order = order
        self._smoothing = smoothing
        self._n_points = degree + 1

        self._coeff_vel = _compute_coefficients(degree, 1)
        self._coeff_acc = _compute_coefficients(degree, 2) if order == 2 else None

        # EMA state (used only when smoothing > 0).
        self._ema_alpha = 2.0 / (smoothing + 1.0) if smoothing > 0 else 0.0
        self._ema_value = 0.0
        self._ema_initialized = False

        # Ring buffer of smoothed prices (size = degree + 1).
        self._buf = [0.0] * self._n_points
        self._buf_pos = 0
        self._buf_count = 0

        self._primed = False

    def is_primed(self) -> bool:
        return self._primed

    def metadata(self) -> Metadata:
        return build_metadata(
            Identifier.POLYNOMIAL_FORECAST,
            self._line.mnemonic,
            self._line.description,
            [OutputText(self._line.mnemonic, self._line.description)],
        )

    def update(self, sample: float) -> float:
        # Optional EMA pre-smoothing.
        if self._smoothing > 0:
            if not self._ema_initialized:
                self._ema_value = sample
                self._ema_initialized = True
            else:
                self._ema_value = self._ema_alpha * sample + (1.0 - self._ema_alpha) * self._ema_value
            smoothed = self._ema_value
        else:
            smoothed = sample

        # Store the smoothed price in the ring buffer.
        self._buf[self._buf_pos] = smoothed
        self._buf_pos = (self._buf_pos + 1) % self._n_points
        self._buf_count += 1

        if self._buf_count < self._n_points:
            self._primed = False
            return math.nan

        self._primed = True

        # Read buffer most-recent-first.
        velocity = 0.0
        acceleration = 0.0
        for k in range(self._n_points):
            idx = (self._buf_pos - 1 - k) % self._n_points
            value = self._buf[idx]
            velocity += self._coeff_vel[k] * value
            if self._coeff_acc is not None:
                acceleration += self._coeff_acc[k] * value

        forecast = smoothed + velocity
        if self._order == 2:
            forecast += 0.5 * acceleration

        return forecast

    def update_scalar(self, sample: Scalar) -> Output:
        return self._line.update_scalar(sample)

    def update_bar(self, sample: Bar) -> Output:
        return self._line.update_bar(sample)

    def update_quote(self, sample: Quote) -> Output:
        return self._line.update_quote(sample)

    def update_trade(self, sample: Trade) -> Output:
        return self._line.update_trade(sample)
