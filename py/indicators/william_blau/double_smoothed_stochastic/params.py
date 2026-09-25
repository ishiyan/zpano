"""Double Smoothed Stochastic parameters."""

from dataclasses import dataclass


@dataclass
class DoubleSmoothedStochasticParams:
    """Parameters to create an instance of the Double Smoothed Stochastic indicator.

    The parameter names ``q``, ``r``, ``s`` and ``g`` are the canonical symbols
    from William Blau's *Momentum, Direction, and Divergence* (Wiley, 1995). They
    are kept verbatim for fidelity with the book, the MQL5 reference, and the
    test-data naming.

    The indicator consumes the high, low and close prices of a bar, so it has
    no configurable price-component fields.
    """

    q: int = 5
    """The stochastic look-back period: the number of bars over which the
    highest high and the lowest low are taken.

    Setting ``q=1`` yields the book's one-bar HLC index. The value should be
    greater than 0. The default value is 5.
    """

    r: int = 7
    """The period of the 1st (inner) EMA in the smoothing cascade, applied to
    the raw stochastic and the range.

    The value should be greater than 0. The default value is 7.
    """

    s: int = 3
    """The period of the 2nd (outer) EMA in the smoothing cascade, applied to
    the output of the 1st EMA.

    The value should be greater than 0. The default value is 3.
    """

    g: int = 3
    """The period of the signal-line SMA, applied to the oscillator to produce
    the second output.

    Setting ``g=1`` makes the signal a passthrough (signal == dss every bar).
    The value should be greater than 0. The default value is 3.
    """


def default_params() -> DoubleSmoothedStochasticParams:
    """Returns default parameters for the Double Smoothed Stochastic indicator."""
    return DoubleSmoothedStochasticParams()
