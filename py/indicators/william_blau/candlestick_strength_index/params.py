"""Candlestick Strength Index parameters."""

from dataclasses import dataclass


@dataclass
class CandlestickStrengthIndexParams:
    """Parameters to create an instance of the Candlestick Strength Index indicator.

    The parameter names ``r``, ``s``, ``u`` and ``ul`` are the canonical symbols
    from William Blau's *Momentum, Direction, and Divergence* (Wiley, 1995),
    chapter 6. They are kept verbatim for fidelity with the book, the MQL5
    reference, and the test-data naming.

    The indicator consumes the open, high, low and close prices of a bar, so it
    has no configurable price-component fields.
    """

    r: int = 20
    """The period of the 1st (innermost) EMA in the smoothing cascade, applied
    to the candle body and the bar range.

    The value should be greater than 0. The default value is 20.
    """

    s: int = 5
    """The period of the 2nd EMA in the smoothing cascade, applied to the output
    of the 1st EMA.

    The value should be greater than 0. The default value is 5.
    """

    u: int = 3
    """The period of the 3rd (outermost) EMA in the smoothing cascade, applied
    to the output of the 2nd EMA.

    Setting ``u=1`` switches the 3rd stage off (passthrough), yielding the
    book's classic double-smoothed oscillator. The value should be greater
    than 0. The default value is 3.
    """

    ul: int = 3
    """The period of the signal-line EMA, applied to the oscillator to produce
    the second output.

    Setting ``ul=1`` makes the signal a passthrough (signal == csi every bar).
    The value should be greater than 0. The default value is 3.
    """


def default_params() -> CandlestickStrengthIndexParams:
    """Returns default parameters for the Candlestick Strength Index indicator."""
    return CandlestickStrengthIndexParams()
