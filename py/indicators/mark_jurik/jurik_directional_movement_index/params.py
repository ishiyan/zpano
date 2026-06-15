"""Jurik directional movement index parameters."""

from dataclasses import dataclass


@dataclass
class JurikDirectionalMovementIndexParams:
    """Parameters for the Jurik directional movement index indicator."""
    length: int = 14
    """JMA smoothing length (minimum 1)."""


def default_params() -> JurikDirectionalMovementIndexParams:
    """Return default parameters."""
    return JurikDirectionalMovementIndexParams()
