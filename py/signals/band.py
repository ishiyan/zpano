"""Band signals.

Fuzzy membership for price/value relative to dynamic bands
(e.g., Bollinger Bands, Keltner Channels, Donchian).
"""
from __future__ import annotations

from ..fuzzy import MembershipShape, mu_greater, mu_less, t_product


def mu_above_band(value: float, upper_band: float,
                  width: float = 0.0,
                  shape: MembershipShape = MembershipShape.SIGMOID) -> float:
    """Degree to which *value* is above the upper band.

    When ``width == 0``, this is a crisp check.  With ``width > 0``
    the transition is gradual — useful when the band itself has
    measurement uncertainty.

    A common choice for *width* is a fraction of the band spread:
    ``width = 0.1 * (upper_band - lower_band)``.

    Args:
        value: Current price or indicator value.
        upper_band: Upper band level.
        width: Fuzzy transition width.
        shape: ``MembershipShape.SIGMOID`` or ``MembershipShape.LINEAR``.

    Returns:
        Membership degree ∈ [0, 1].
    """
    return mu_greater(value, upper_band, width, shape)


def mu_below_band(value: float, lower_band: float,
                  width: float = 0.0,
                  shape: MembershipShape = MembershipShape.SIGMOID) -> float:
    """Degree to which *value* is below the lower band.

    Args:
        value: Current price or indicator value.
        lower_band: Lower band level.
        width: Fuzzy transition width.
        shape: ``MembershipShape.SIGMOID`` or ``MembershipShape.LINEAR``.

    Returns:
        Membership degree ∈ [0, 1].
    """
    return mu_less(value, lower_band, width, shape)


def mu_between_bands(value: float, lower_band: float, upper_band: float,
                     shape: MembershipShape = MembershipShape.SIGMOID) -> float:
    """Degree to which *value* is inside the band channel.

    Computed as ``mu_above(value, lower) * mu_below(value, upper)``
    using the band spread as the transition width for both sides.
    This gives a smooth bell that peaks between the bands and falls
    off as the value approaches either edge.

    Args:
        value: Current price or indicator value.
        lower_band: Lower band level.
        upper_band: Upper band level.
        shape: ``MembershipShape.SIGMOID`` or ``MembershipShape.LINEAR``.

    Returns:
        Membership degree ∈ [0, 1].  1.0 when centered, falling
        toward 0 near/outside the bands.
    """
    if upper_band <= lower_band:
        return 0.0
    spread = upper_band - lower_band
    # Width = half the spread — gives a smooth transition at each band edge.
    width = spread * 0.5
    above_lower = mu_greater(value, lower_band, width, shape)
    below_upper = mu_less(value, upper_band, width, shape)
    return t_product(above_lower, below_upper)
