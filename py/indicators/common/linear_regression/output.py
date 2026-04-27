"""Linear regression output enum."""

from enum import IntEnum


class LinearRegressionOutput(IntEnum):
    """Enumerates the outputs of the linear regression indicator."""

    VALUE = 0
    FORECAST = 1
    INTERCEPT = 2
    SLOPE_RAD = 3
    SLOPE_DEG = 4
