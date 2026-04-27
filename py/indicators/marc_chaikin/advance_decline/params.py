"""Parameters for the Advance-Decline indicator."""

from dataclasses import dataclass


@dataclass
class AdvanceDeclineParams:
    """Parameters for Advance-Decline indicator. No configurable parameters."""

    pass


def default_params() -> AdvanceDeclineParams:
    """Return default Advance-Decline parameters."""
    return AdvanceDeclineParams()
