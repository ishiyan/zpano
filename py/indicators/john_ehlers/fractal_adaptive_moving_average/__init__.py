"""FractalAdaptiveMovingAverage indicator package."""

from .fractal_adaptive_moving_average import FractalAdaptiveMovingAverage
from .output import FractalAdaptiveMovingAverageOutput
from .params import (
    FractalAdaptiveMovingAverageParams,
    Params,
    default_params,
    fractal_adaptive_moving_average_dimension_mnemonic,
    fractal_adaptive_moving_average_mnemonic,
)

__all__ = [
    "FractalAdaptiveMovingAverage",
    "FractalAdaptiveMovingAverageOutput",
    "FractalAdaptiveMovingAverageParams",
    "Params",
    "default_params",
    "fractal_adaptive_moving_average_dimension_mnemonic",
    "fractal_adaptive_moving_average_mnemonic",
]
