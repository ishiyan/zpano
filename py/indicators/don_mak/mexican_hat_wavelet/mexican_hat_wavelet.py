"""Mexican Hat Wavelet (MHW) indicator -- Don Mak.

A causal bandpass FIR filter derived from the Mexican Hat wavelet (the second
derivative of a Gaussian). It decomposes price data into frequency bands with
zero phase shift at the filter's center frequency.

Three standard bands are provided (HIGH, MID, LOW) plus a CUSTOM band that takes
either a dilation or a center period.

Reference:
    Mak, Don K. (2006). Mathematical Techniques in Financial Market Trading.
    World Scientific. Chapter 5: Causal Wavelet Filters.

Single output:
  - value: the bandpass-filtered price component, NaN until the ring buffer is
    full (requires K + 1 prices).
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
from .params import MexicanHatWaveletParams, Band

# Preset dilation values (a_f) for the three standard bands (Table 5.2).
DILATION_HIGH = 1.483   # omega_0 = 1.3558 rad, period ~ 4.63 bars
DILATION_MID = 4.048    # omega_0 = 0.4670 rad, period ~ 13.45 bars
DILATION_LOW = 15.97    # omega_0 = 0.1156 rad, period ~ 54.35 bars


def _dilation_from_period(period: float) -> float:
    """Compute dilation a_f from a desired center period in bars (Eq 5.11)."""
    omega_0 = 2.0 * math.pi / period
    two_over_a = 1.091 * omega_0 - 0.071 * omega_0 * omega_0
    if two_over_a <= 0.0:
        raise ValueError(
            "invalid mexican hat wavelet parameters: "
            "period is too large for the fitting formula (2/a <= 0)")
    return 2.0 / two_over_a


def _compute_coefficients(a_f: float) -> list[float]:
    """Compute normalized Mexican Hat wavelet FIR coefficients for dilation a_f.

    psi(t) = (1 - 2*t^2) * exp(-t^2); h(n) = psi(n / a_f) for n = 0..K,
    K = 4 * round(a_f), normalized by |H(omega_0)|_f = 0.488 + 0.646*a_f + 0.0001*a_f^2.
    """
    k = 4 * round(a_f)
    if k < 1:
        k = 1

    norm = 0.488 + 0.646 * a_f + 0.0001 * a_f * a_f

    coeffs: list[float] = []
    for n in range(k + 1):
        t = n / a_f
        t2 = t * t
        h_n = (1.0 - 2.0 * t2) * math.exp(-t2)
        coeffs.append(h_n / norm)

    return coeffs


class MexicanHatWavelet(Indicator):
    """Don Mak's Mexican Hat Wavelet (MHW) bandpass filter."""

    def __init__(self, params: MexicanHatWaveletParams) -> None:
        band = Band(params.band)
        dilation = params.dilation
        period = params.period

        # Resolve dilation a_f and the mnemonic suffix.
        if band == Band.HIGH:
            a_f = DILATION_HIGH
            cfg = "high"
        elif band == Band.MID:
            a_f = DILATION_MID
            cfg = "mid"
        elif band == Band.LOW:
            a_f = DILATION_LOW
            cfg = "low"
        elif band == Band.CUSTOM:
            has_dilation = dilation != 0.0
            has_period = period != 0.0
            if has_dilation and has_period:
                raise ValueError(
                    "invalid mexican hat wavelet parameters: "
                    "provide only one of dilation or period, not both")
            if not has_dilation and not has_period:
                raise ValueError(
                    "invalid mexican hat wavelet parameters: "
                    "band=CUSTOM requires either dilation or period")
            if has_period:
                if period <= 2.0:
                    raise ValueError(
                        "invalid mexican hat wavelet parameters: period must be > 2")
                a_f = _dilation_from_period(period)
                cfg = f"p{period:.2f}"
            else:
                if dilation <= 0.0:
                    raise ValueError(
                        "invalid mexican hat wavelet parameters: dilation must be > 0")
                a_f = dilation
                cfg = f"d{dilation:.2f}"
        else:
            raise ValueError("invalid mexican hat wavelet parameters: unknown band")

        bc = params.bar_component if params.bar_component is not None else DEFAULT_BAR_COMPONENT
        qc = params.quote_component if params.quote_component is not None else DEFAULT_QUOTE_COMPONENT
        tc = params.trade_component if params.trade_component is not None else DEFAULT_TRADE_COMPONENT

        bar_func = bar_component_value(bc)
        quote_func = quote_component_value(qc)
        trade_func = trade_component_value(tc)

        mnemonic = f"mhw({cfg}{component_triple_mnemonic(bc, qc, tc)})"
        description = f"Mexican hat wavelet {mnemonic}"

        self._line = LineIndicator(mnemonic, description, bar_func, quote_func, trade_func, self.update)

        self._a_f = a_f
        self._coefficients = _compute_coefficients(a_f)
        self._num_taps = len(self._coefficients)

        # Ring buffer: buffer[0] = most recent price, buffer[k] = x[n-k].
        self._buffer = [0.0] * self._num_taps
        self._count = 0
        self._primed = False

    def is_primed(self) -> bool:
        return self._primed

    def metadata(self) -> Metadata:
        return build_metadata(
            Identifier.MEXICAN_HAT_WAVELET,
            self._line.mnemonic,
            self._line.description,
            [OutputText(self._line.mnemonic, self._line.description)],
        )

    def update(self, sample: float) -> float:
        # Shift buffer right and insert the new price at position 0.
        for i in range(self._num_taps - 1, 0, -1):
            self._buffer[i] = self._buffer[i - 1]
        self._buffer[0] = sample
        self._count += 1

        if self._count < self._num_taps:
            self._primed = False
            return math.nan

        # FIR convolution: y = sum(coeffs[k] * buffer[k]).
        y = 0.0
        for k in range(self._num_taps):
            y += self._coefficients[k] * self._buffer[k]

        self._primed = True
        return y

    def update_scalar(self, sample: Scalar) -> Output:
        return self._line.update_scalar(sample)

    def update_bar(self, sample: Bar) -> Output:
        return self._line.update_bar(sample)

    def update_quote(self, sample: Quote) -> Output:
        return self._line.update_quote(sample)

    def update_trade(self, sample: Trade) -> Output:
        return self._line.update_trade(sample)
