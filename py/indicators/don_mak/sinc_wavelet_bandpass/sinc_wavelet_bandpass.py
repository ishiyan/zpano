"""Sinc Wavelet Band-Pass (SWB) indicator -- Don Mak.

A causal FIR band-pass filter derived from the sinc wavelet system. It decomposes
price data into frequency bands (HIGH, MID, LOW, FULL). Optionally applies a cubic
velocity kernel to produce a momentum oscillator.

Reference:
    Mak, D.K. (2003). The Science of Financial Market Trading.
    World Scientific. Chapter 9, Appendix 7.

Single output:
  - value: the band-passed price (or velocity of the band-passed price), NaN
    until the indicator is primed.
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
from .params import SincWaveletBandpassParams, Band

# Band parameters: (omega_0, omega_1, num_taps).
_BAND_PARAMS = {
    Band.HIGH: (math.pi / 4, math.pi / 8, 121),
    Band.MID: (math.pi / 8, math.pi / 16, 121),
    Band.LOW: (math.pi / 16, math.pi / 32, 201),
    Band.FULL: (math.pi / 4, math.pi / 32, 201),
}

# Cubic velocity kernel (PFD degree=3, order=1, smoothing=0).
_VELOCITY_KERNEL = [11.0 / 6.0, -3.0, 3.0 / 2.0, -1.0 / 3.0]
_VELOCITY_TAPS = len(_VELOCITY_KERNEL)  # 4


def _compute_coefficients(omega_0: float, omega_1: float, num_taps: int) -> list[float]:
    """Compute sinc band-pass filter coefficients (difference of two sinc functions)."""
    coeffs = [0.0] * num_taps
    coeffs[0] = (omega_0 - omega_1) / math.pi
    for k in range(1, num_taps):
        pi_k = math.pi * k
        coeffs[k] = math.sin(omega_0 * k) / pi_k - math.sin(omega_1 * k) / pi_k
    return coeffs


class SincWaveletBandpass(Indicator):
    """Don Mak's Sinc Wavelet Band-Pass (SWB) filter."""

    def __init__(self, params: SincWaveletBandpassParams) -> None:
        band = Band(params.band)
        velocity = bool(params.velocity)

        if band not in _BAND_PARAMS:
            raise ValueError("invalid sinc wavelet band-pass parameters: unknown band")

        band_names = {Band.HIGH: "high", Band.MID: "mid", Band.LOW: "low", Band.FULL: "full"}

        bc = params.bar_component if params.bar_component is not None else DEFAULT_BAR_COMPONENT
        qc = params.quote_component if params.quote_component is not None else DEFAULT_QUOTE_COMPONENT
        tc = params.trade_component if params.trade_component is not None else DEFAULT_TRADE_COMPONENT

        bar_func = bar_component_value(bc)
        quote_func = quote_component_value(qc)
        trade_func = trade_component_value(tc)

        cfg = band_names[band] + (",v" if velocity else "")
        mnemonic = f"swb({cfg}{component_triple_mnemonic(bc, qc, tc)})"
        description = f"Sinc wavelet band-pass {mnemonic}"

        self._line = LineIndicator(mnemonic, description, bar_func, quote_func, trade_func, self.update)

        self._velocity = velocity

        omega_0, omega_1, num_taps = _BAND_PARAMS[band]
        self._num_taps = num_taps
        self._coefficients = _compute_coefficients(omega_0, omega_1, num_taps)

        # Ring buffer for prices (size = num_taps).
        self._price_buffer = [0.0] * num_taps
        self._price_count = 0
        self._price_index = 0

        # Velocity ring buffer (size = 4), used only when velocity is enabled.
        self._vel_buffer = [0.0] * _VELOCITY_TAPS
        self._vel_count = 0
        self._vel_index = 0

        self._primed = False

    def is_primed(self) -> bool:
        return self._primed

    def metadata(self) -> Metadata:
        return build_metadata(
            Identifier.SINC_WAVELET_BANDPASS,
            self._line.mnemonic,
            self._line.description,
            [OutputText(self._line.mnemonic, self._line.description)],
        )

    def update(self, sample: float) -> float:
        # Store price in the ring buffer.
        self._price_buffer[self._price_index] = sample
        self._price_index = (self._price_index + 1) % self._num_taps
        self._price_count += 1

        if self._price_count < self._num_taps:
            self._primed = False
            return math.nan

        # Band-pass convolution: coefficients[k] multiplies the k-th most recent price.
        bp_value = 0.0
        idx = self._price_index - 1
        for k in range(self._num_taps):
            buf_idx = idx % self._num_taps
            bp_value += self._coefficients[k] * self._price_buffer[buf_idx]
            idx -= 1

        if not self._velocity:
            self._primed = True
            return bp_value

        # Store band-pass output in the velocity ring buffer.
        self._vel_buffer[self._vel_index] = bp_value
        self._vel_index = (self._vel_index + 1) % _VELOCITY_TAPS
        self._vel_count += 1

        if self._vel_count < _VELOCITY_TAPS:
            self._primed = False
            return math.nan

        # Cubic velocity: kernel[k] multiplies the k-th most recent band-pass value.
        vel_value = 0.0
        idx = self._vel_index - 1
        for k in range(_VELOCITY_TAPS):
            buf_idx = idx % _VELOCITY_TAPS
            vel_value += _VELOCITY_KERNEL[k] * self._vel_buffer[buf_idx]
            idx -= 1

        self._primed = True
        return vel_value

    def update_scalar(self, sample: Scalar) -> Output:
        return self._line.update_scalar(sample)

    def update_bar(self, sample: Bar) -> Output:
        return self._line.update_bar(sample)

    def update_quote(self, sample: Quote) -> Output:
        return self._line.update_quote(sample)

    def update_trade(self, sample: Trade) -> Output:
        return self._line.update_trade(sample)
