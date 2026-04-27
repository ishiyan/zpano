"""Shared helper functions for Hilbert transformer cycle estimators."""
import math

from .cycle_estimator import CycleEstimator
from .cycle_estimator_params import CycleEstimatorParams
from .cycle_estimator_type import CycleEstimatorType

DEFAULT_MIN_PERIOD = 6
DEFAULT_MAX_PERIOD = 50
HT_LENGTH = 7
QUADRATURE_INDEX = HT_LENGTH // 2
ACCUMULATION_LENGTH = 40


def push(array: list[float], value: float) -> None:
    """Shifts all elements to the right and places the new value at index zero."""
    for i in range(len(array) - 1, 0, -1):
        array[i] = array[i - 1]
    array[0] = value


def ht(array: list[float]) -> float:
    """Hilbert transform of a 7-element array."""
    a = 0.0962
    b = 0.5769
    return a * array[0] + b * array[2] - b * array[4] - a * array[6]


def correct_amplitude(previous_period: float) -> float:
    """Computes amplitude correction factor."""
    return 0.54 + 0.075 * previous_period


def adjust_period(period: float, period_previous: float) -> float:
    """Clamps period to [0.67x, 1.5x] of previous and [6, 50] absolute."""
    temp = 1.5 * period_previous
    if period > temp:
        period = temp
    else:
        temp = 0.67 * period_previous
        if period < temp:
            period = temp

    if period < DEFAULT_MIN_PERIOD:
        period = DEFAULT_MIN_PERIOD
    elif period > DEFAULT_MAX_PERIOD:
        period = DEFAULT_MAX_PERIOD

    return period


def fill_wma_factors(length: int, factors: list[float]) -> None:
    """Fills WMA weight factors for the given length (2, 3, or 4)."""
    if length == 4:
        factors[0] = 4.0 / 10.0
        factors[1] = 3.0 / 10.0
        factors[2] = 2.0 / 10.0
        factors[3] = 1.0 / 10.0
    elif length == 3:
        factors[0] = 3.0 / 6.0
        factors[1] = 2.0 / 6.0
        factors[2] = 1.0 / 6.0
    else:  # length == 2
        factors[0] = 2.0 / 3.0
        factors[1] = 1.0 / 3.0


def verify_parameters(p: CycleEstimatorParams) -> None:
    """Validates cycle estimator parameters. Raises ValueError if invalid."""
    invalid = "invalid cycle estimator parameters"

    if p.smoothing_length < 2 or p.smoothing_length > 4:
        raise ValueError(f"{invalid}: SmoothingLength should be in range [2, 4]")

    if p.alpha_ema_quadrature_in_phase <= 0 or p.alpha_ema_quadrature_in_phase >= 1:
        raise ValueError(
            f"{invalid}: AlphaEmaQuadratureInPhase should be in range (0, 1)"
        )

    if p.alpha_ema_period <= 0 or p.alpha_ema_period >= 1:
        raise ValueError(f"{invalid}: AlphaEmaPeriod should be in range (0, 1)")


def estimator_moniker(typ: CycleEstimatorType, estimator: CycleEstimator) -> str:
    """Returns the moniker of the cycle estimator."""
    prefixes = {
        CycleEstimatorType.HOMODYNE_DISCRIMINATOR: "hd",
        CycleEstimatorType.HOMODYNE_DISCRIMINATOR_UNROLLED: "hdu",
        CycleEstimatorType.PHASE_ACCUMULATOR: "pa",
        CycleEstimatorType.DUAL_DIFFERENTIATOR: "dd",
    }

    prefix = prefixes.get(typ)
    if prefix is None:
        return ""

    return f"{prefix}({estimator.smoothing_length()}, " \
           f"{estimator.alpha_ema_quadrature_in_phase():.3f}, " \
           f"{estimator.alpha_ema_period():.3f})"


def new_cycle_estimator(
    typ: CycleEstimatorType, params: CycleEstimatorParams
) -> CycleEstimator:
    """Creates a new cycle estimator based on the specified type and parameters."""
    from .homodyne_discriminator_estimator import HomodyneDiscriminatorEstimator
    from .homodyne_discriminator_estimator_unrolled import \
        HomodyneDiscriminatorEstimatorUnrolled
    from .phase_accumulator_estimator import PhaseAccumulatorEstimator
    from .dual_differentiator_estimator import DualDifferentiatorEstimator

    if typ == CycleEstimatorType.HOMODYNE_DISCRIMINATOR:
        return HomodyneDiscriminatorEstimator(params)
    elif typ == CycleEstimatorType.HOMODYNE_DISCRIMINATOR_UNROLLED:
        return HomodyneDiscriminatorEstimatorUnrolled(params)
    elif typ == CycleEstimatorType.PHASE_ACCUMULATOR:
        return PhaseAccumulatorEstimator(params)
    elif typ == CycleEstimatorType.DUAL_DIFFERENTIATOR:
        return DualDifferentiatorEstimator(params)
    else:
        raise ValueError(f"invalid cycle estimator type: {typ}")
