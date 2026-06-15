"""Quantum Price Levels output enum."""

from enum import IntEnum


class QuantumPriceLevelsOutput(IntEnum):
    """Describes the outputs of the indicator."""

    LAMBDA = 0
    """The anharmonic coefficient (lambda) of the quantum potential well."""

    RETURN_STD_DEV = 1
    """The population standard deviation of the price-return ratios in the window."""

    NORMALIZED_MULTIPLIERS = 2
    """The normalized QPR multipliers (1 + scale_factor*sigma*QPR(n)), one per level."""

    RESISTANCES = 3
    """The resistance price levels above the current price (price * NQPR(n))."""

    SUPPORTS = 4
    """The support price levels below the current price (price / NQPR(n))."""
