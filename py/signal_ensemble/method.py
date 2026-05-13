"""Aggregation method enum for signal ensemble blending."""

from enum import IntEnum


class AggregationMethod(IntEnum):
    """Aggregation method for combining multiple signal sources.

    Stateless methods (FIXED, EQUAL) ignore feedback entirely.
    Adaptive methods update weights based on observed outcomes.
    """

    FIXED = 0                   # User-supplied static weights
    EQUAL = 1                   # Uniform 1/n weights
    INVERSE_VARIANCE = 2        # Weight by 1/variance of errors
    EXPONENTIAL_DECAY = 3       # EMA of accuracy
    MULTIPLICATIVE_WEIGHTS = 4  # Hedge algorithm (online learning)
    RANK_BASED = 5              # Weight by rank of rolling accuracy
    BAYESIAN = 6                # Bayesian model averaging
