"""Polynomial Fit Derivative output enum."""

from enum import IntEnum


class PolynomialFitDerivativeOutput(IntEnum):
    """Describes the outputs of the indicator."""

    VALUE = 0
    """The order-th derivative of the polynomial fit at the current bar."""
