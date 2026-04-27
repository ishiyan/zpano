"""Cycle estimator parameters."""
from dataclasses import dataclass


@dataclass
class CycleEstimatorParams:
    """Parameters to create an instance of the Hilbert transformer cycle estimator."""

    smoothing_length: int = 4
    """The smoothing length (2, 3, or 4) of the underlying WMA."""

    alpha_ema_quadrature_in_phase: float = 0.2
    """The value of alpha (0 < alpha <= 1) used in EMA to smooth
    the in-phase and quadrature components."""

    alpha_ema_period: float = 0.2
    """The value of alpha (0 < alpha <= 1) used in EMA to smooth
    the instantaneous period."""

    warm_up_period: int = 0
    """The number of updates before the estimator is primed."""


def default_params() -> CycleEstimatorParams:
    """Returns a CycleEstimatorParams with Ehlers defaults."""
    return CycleEstimatorParams()
