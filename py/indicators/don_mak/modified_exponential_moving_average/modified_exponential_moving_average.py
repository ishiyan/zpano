"""Modified Exponential Moving Average (MEMA / MEMA-D) indicator -- Don Mak.

A reduced-lag EMA that adds the EMA's own polynomial velocity back to its output,
compensating for smoothing delay:

    MEMA(n)   = EMA(n) + PFD(EMA, degree, order=1, stride=1)
    MEMA-D(n) = EMA(n) + PFD(EMA, degree, order=1, stride=skip)

Reference:
    Mak, Don K. (2006). Mathematical Techniques in Financial Market Trading.
    World Scientific. Chapter 4.2.

Single output:
  - value: the velocity-corrected EMA, NaN until ``degree * skip + 1`` EMA values
    are available.
"""

import math

from ...core.indicator import Indicator
from ...core.line_indicator import LineIndicator
from ...core.metadata import Metadata
from ...core.build_metadata import build_metadata, OutputText
from ...core.identifier import Identifier
from ...core.output import Output
from ...core.component_triple_mnemonic import component_triple_mnemonic
from ....entities.bar import Bar
from ....entities.quote import Quote
from ....entities.trade import Trade
from ....entities.scalar import Scalar
from ....entities.bar_component import BarComponent, DEFAULT_BAR_COMPONENT, bar_component_value
from ....entities.quote_component import QuoteComponent, DEFAULT_QUOTE_COMPONENT, quote_component_value
from ....entities.trade_component import TradeComponent, DEFAULT_TRADE_COMPONENT, trade_component_value
from .params import ModifiedExponentialMovingAverageParams


def _compute_velocity_coefficients(degree: int) -> list[float]:
    """Compute FIR coefficients for the first derivative of a degree-d polynomial
    fit evaluated at the most recent point (Lagrange basis, order=1)."""
    n_points = degree + 1
    coefficients: list[float] = []

    for i in range(n_points):
        denom = 1.0
        for j in range(n_points):
            if j != i:
                denom *= float(j - i)

        others = [j for j in range(n_points) if j != i]

        numerator = 0.0
        for ell_idx in range(len(others)):
            term = 1.0
            for m_idx in range(len(others)):
                if m_idx != ell_idx:
                    term *= float(others[m_idx])
            numerator += term

        coefficients.append(numerator / denom)

    return coefficients


class ModifiedExponentialMovingAverage(Indicator):
    """Don Mak's Modified Exponential Moving Average (MEMA / MEMA-D) indicator."""

    def __init__(self, params: ModifiedExponentialMovingAverageParams) -> None:
        period = params.period
        degree = params.degree
        skip = params.skip

        if period < 2:
            raise ValueError("invalid modified exponential moving average parameters: period must be >= 2")
        if degree < 2:
            raise ValueError("invalid modified exponential moving average parameters: degree must be >= 2")
        if skip < 1:
            raise ValueError("invalid modified exponential moving average parameters: skip must be >= 1")

        bc = params.bar_component if params.bar_component is not None else DEFAULT_BAR_COMPONENT
        qc = params.quote_component if params.quote_component is not None else DEFAULT_QUOTE_COMPONENT
        tc = params.trade_component if params.trade_component is not None else DEFAULT_TRADE_COMPONENT

        bar_func = bar_component_value(bc)
        quote_func = quote_component_value(qc)
        trade_func = trade_component_value(tc)

        mnemonic = f"mema({period},{degree},{skip}{component_triple_mnemonic(bc, qc, tc)})"
        description = f"Modified exponential moving average {mnemonic}"

        self._line = LineIndicator(mnemonic, description, bar_func, quote_func, trade_func, self.update)

        self._degree = degree
        self._skip = skip
        self._n_points = degree + 1
        self._coefficients = _compute_velocity_coefficients(degree)

        # EMA state.
        self._ema_alpha = 2.0 / (period + 1.0)
        self._ema_value = 0.0
        self._ema_initialized = False

        # Ring buffer of EMA values (size = degree * skip + 1).
        self._buf_size = degree * skip + 1
        self._buf = [0.0] * self._buf_size
        self._buf_pos = 0
        self._buf_count = 0

        self._primed = False

    def is_primed(self) -> bool:
        return self._primed

    def metadata(self) -> Metadata:
        return build_metadata(
            Identifier.MODIFIED_EXPONENTIAL_MOVING_AVERAGE,
            self._line.mnemonic,
            self._line.description,
            [OutputText(self._line.mnemonic, self._line.description)],
        )

    def update(self, sample: float) -> float:
        # EMA recursion (seed at first sample).
        if not self._ema_initialized:
            self._ema_value = sample
            self._ema_initialized = True
        else:
            self._ema_value = self._ema_alpha * sample + (1.0 - self._ema_alpha) * self._ema_value

        # Store EMA value in the ring buffer.
        self._buf[self._buf_pos] = self._ema_value
        self._buf_pos = (self._buf_pos + 1) % self._buf_size
        self._buf_count += 1

        if self._buf_count < self._buf_size:
            self._primed = False
            return math.nan

        self._primed = True

        # Read EMA values at stride positions and compute the velocity correction.
        velocity = 0.0
        for k in range(self._n_points):
            offset = k * self._skip
            idx = (self._buf_pos - 1 - offset) % self._buf_size
            velocity += self._coefficients[k] * self._buf[idx]

        return self._ema_value + velocity

    def update_scalar(self, sample: Scalar) -> Output:
        return self._line.update_scalar(sample)

    def update_bar(self, sample: Bar) -> Output:
        return self._line.update_bar(sample)

    def update_quote(self, sample: Quote) -> Output:
        return self._line.update_quote(sample)

    def update_trade(self, sample: Trade) -> Output:
        return self._line.update_trade(sample)
