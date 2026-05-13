"""Signal ensemble: weighted blending of multiple signal sources."""

from .method import AggregationMethod
from .error_metric import ErrorMetric
from .aggregator import Aggregator

__all__ = ["Aggregator", "AggregationMethod", "ErrorMetric"]
