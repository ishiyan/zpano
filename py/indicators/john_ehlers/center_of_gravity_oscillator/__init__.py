"""CenterOfGravityOscillator indicator package."""

from .center_of_gravity_oscillator import CenterOfGravityOscillator
from .output import CenterOfGravityOscillatorOutput
from .params import (
    CenterOfGravityOscillatorParams,
    Params,
    center_of_gravity_oscillator_mnemonic,
    center_of_gravity_oscillator_trigger_mnemonic,
    default_params,
)

__all__ = [
    "CenterOfGravityOscillator",
    "CenterOfGravityOscillatorOutput",
    "CenterOfGravityOscillatorParams",
    "Params",
    "center_of_gravity_oscillator_mnemonic",
    "center_of_gravity_oscillator_trigger_mnemonic",
    "default_params",
]
