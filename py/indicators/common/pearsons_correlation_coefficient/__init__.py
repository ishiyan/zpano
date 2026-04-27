"""Pearson's correlation coefficient indicator."""

from .pearsons_correlation_coefficient import PearsonsCorrelationCoefficient
from .params import PearsonsCorrelationCoefficientParams, default_params
from .output import PearsonsCorrelationCoefficientOutput

__all__ = [
    'PearsonsCorrelationCoefficient',
    'PearsonsCorrelationCoefficientParams',
    'PearsonsCorrelationCoefficientOutput',
    'default_params',
]
