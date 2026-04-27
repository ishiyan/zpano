"""Balance of Power parameters."""

from dataclasses import dataclass


@dataclass
class BalanceOfPowerParams:
    """Parameters to create an instance of the balance of power indicator.

    BOP has no configurable parameters.
    """

    pass


def default_params() -> BalanceOfPowerParams:
    """Returns default parameters for the balance of power indicator."""
    return BalanceOfPowerParams()
