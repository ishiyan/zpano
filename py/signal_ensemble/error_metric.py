"""Error metric enum for signal ensemble weight computation."""

from enum import IntEnum


class ErrorMetric(IntEnum):
    """Error metric used by inverse-variance and rank-based methods.

    ABSOLUTE: |signal_i - outcome|
    SQUARED:  (signal_i - outcome)^2
    """

    ABSOLUTE = 0  # |signal_i - outcome|
    SQUARED = 1   # (signal_i - outcome)^2
