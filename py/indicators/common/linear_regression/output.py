"""Linear regression output enum."""

from enum import IntEnum


class LinearRegressionOutput(IntEnum):
    """Enumerates the outputs of the linear regression indicator."""

    VALUE = 0
    """The regression value at the last bar of the window: b + m*(period-1)."""

    FORECAST = 1
    """The time series forecast (one bar ahead): b + m*period."""

    INTERCEPT = 2
    """The y-intercept of the regression line: b."""

    SLOPE_RAD = 3
    """The slope of the regression line: m."""

    SLOPE_DEG = 4
    """The slope expressed in degrees: atan(m)*180/pi."""
