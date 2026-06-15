"""Polynomial Forecast output enum."""

from enum import IntEnum


class PolynomialForecastOutput(IntEnum):
    """Describes the outputs of the indicator."""

    VALUE = 0
    """The 1-bar-ahead price forecast value."""
