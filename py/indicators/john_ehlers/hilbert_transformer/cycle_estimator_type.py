"""Cycle estimator type enumeration."""
from enum import IntEnum


class CycleEstimatorType(IntEnum):
    """Enumerates types of techniques to estimate an instantaneous period
    using a Hilbert transformer."""

    HOMODYNE_DISCRIMINATOR = 0
    HOMODYNE_DISCRIMINATOR_UNROLLED = 1
    PHASE_ACCUMULATOR = 2
    DUAL_DIFFERENTIATOR = 3
