"""Instantaneous Sine Wave Period (ISWP) indicator -- Don Mak.

Estimates the dominant cycle period of price data by modeling it locally as a
single sine wave superimposed on a constant level. Two estimation methods (a
4-point method IF4 and a 5-point method IF5) are combined, selecting the one
with the lower estimation error at each bar. When neither method produces a
valid estimate, the outputs are NaN.

Reference:
    Mak, Don K. (2006). Mathematical Techniques in Financial Market Trading.
    World Scientific. Chapter 6 & Appendix 3.

The indicator produces seven outputs:
  - period:       cycle period in bars (T = 2*pi/omega), NaN if invalid;
  - omega:        circular frequency in radians/bar, NaN if invalid;
  - velocity:     wave velocity A*omega*cos(phi), NaN if invalid;
  - acceleration: wave acceleration -A*omega^2*sin(phi), NaN if invalid;
  - amplitude:    sine wave amplitude A, NaN if invalid;
  - phase:        phase angle phi in radians, NaN if invalid;
  - dc_level:     constant level D, NaN if invalid.
"""

import math
from typing import List, Any

from ...core.indicator import Indicator
from ...core.metadata import Metadata
from ...core.build_metadata import build_metadata, OutputText
from ...core.identifier import Identifier
from ...core.component_triple_mnemonic import component_triple_mnemonic
from ....entities.bar import Bar
from ....entities.quote import Quote
from ....entities.trade import Trade
from ....entities.scalar import Scalar
from ....entities.bar_component import BarComponent, DEFAULT_BAR_COMPONENT, bar_component_value
from ....entities.quote_component import QuoteComponent, DEFAULT_QUOTE_COMPONENT, quote_component_value
from ....entities.trade_component import TradeComponent, DEFAULT_TRADE_COMPONENT, trade_component_value
from .params import InstantaneousSineWavePeriodParams


class InstantaneousSineWavePeriod(Indicator):
    """Don Mak's Instantaneous Sine Wave Period (ISWP) indicator."""

    def __init__(self, p: InstantaneousSineWavePeriodParams) -> None:
        smoothing = p.smoothing
        min_period = p.min_period
        max_period = p.max_period
        error_threshold = p.error_threshold
        dx = p.dx

        if smoothing < 0:
            raise ValueError(
                "invalid instantaneous sine wave period parameters: "
                "smoothing must be >= 0")
        if min_period <= 0.0:
            raise ValueError(
                "invalid instantaneous sine wave period parameters: "
                "min_period must be > 0")
        if max_period <= min_period:
            raise ValueError(
                "invalid instantaneous sine wave period parameters: "
                "max_period must be > min_period")
        if error_threshold <= 0.0:
            raise ValueError(
                "invalid instantaneous sine wave period parameters: "
                "error_threshold must be > 0")
        if dx <= 0.0:
            raise ValueError(
                "invalid instantaneous sine wave period parameters: "
                "dx must be > 0")

        bc = p.bar_component if p.bar_component is not None else DEFAULT_BAR_COMPONENT
        qc = p.quote_component if p.quote_component is not None else DEFAULT_QUOTE_COMPONENT
        tc = p.trade_component if p.trade_component is not None else DEFAULT_TRADE_COMPONENT

        self._bar_func = bar_component_value(bc)
        self._quote_func = quote_component_value(qc)
        self._trade_func = trade_component_value(tc)

        self._smoothing = smoothing
        self._min_period = min_period
        self._max_period = max_period
        self._error_threshold = error_threshold
        self._dx = dx

        # EMA state: alpha = 2 / (L + 1), or 1.0 (pass-through) when smoothing == 0.
        self._ema_alpha = 2.0 / (smoothing + 1.0) if smoothing > 0 else 1.0
        self._ema_value = None

        # Ring buffer of the last 5 smoothed prices (index 0 = most recent).
        self._buffer = [0.0] * 5
        self._count = 0

        self._primed = False

        self._mnemonic = (
            f"iswp({smoothing},{min_period:.2f},{max_period:.2f},"
            f"{error_threshold:.2f},{dx:.2f}{component_triple_mnemonic(bc, qc, tc)})")

    def _apply_ema(self, price: float) -> float:
        if self._ema_value is None:
            self._ema_value = price
        else:
            self._ema_value = self._ema_alpha * price + (1.0 - self._ema_alpha) * self._ema_value
        return self._ema_value

    def _push_buffer(self, value: float) -> None:
        for i in range(4, 0, -1):
            self._buffer[i] = self._buffer[i - 1]
        self._buffer[0] = value

    def _calc_omega4(self) -> tuple[float, float]:
        x0 = self._buffer[0]
        xm1 = self._buffer[1]
        xm2 = self._buffer[2]
        xm3 = self._buffer[3]

        den = xm1 - xm2
        if den == 0.0:
            return (math.nan, self._error_threshold)

        ratio = (x0 - xm3) / den

        sqrt_arg = 3.0 - ratio
        if sqrt_arg < 0.0:
            return (math.nan, self._error_threshold)

        arg = 0.5 * math.sqrt(sqrt_arg)
        if arg > 1.0:
            return (math.nan, self._error_threshold)

        omega4 = 2.0 * math.asin(arg)

        dx2 = self._dx * self._dx

        denom1 = 1.0 - 0.25 * sqrt_arg
        if denom1 <= 0.0 or sqrt_arg == 0.0:
            return (omega4, self._error_threshold)

        f1 = 1.0 / (denom1 * sqrt_arg)
        inv_den2 = 1.0 / (den * den)
        q2 = inv_den2 * (dx2 + dx2) + (ratio * ratio) * inv_den2 * (dx2 + dx2)

        product = f1 * q2
        if product < 0.0:
            return (omega4, self._error_threshold)

        return (omega4, 0.5 * math.sqrt(product))

    def _calc_omega5(self) -> tuple[float, float]:
        x0 = self._buffer[0]
        xm1 = self._buffer[1]
        xm3 = self._buffer[3]
        xm4 = self._buffer[4]

        den1 = xm1 - xm3
        if den1 == 0.0:
            return (math.nan, self._error_threshold)

        arg = 0.5 * (x0 - xm4) / den1
        if abs(arg) > 1.0:
            return (math.nan, self._error_threshold)

        omega5 = math.acos(arg)

        dx2 = self._dx * self._dx

        denom = 1.0 - arg * arg
        if denom <= 0.0:
            return (omega5, self._error_threshold)

        f1 = 1.0 / denom
        inv_den1_2 = 1.0 / (den1 * den1)
        numerator_ratio = (x0 - xm4) / (den1 * den1)
        r2 = inv_den1_2 * (dx2 + dx2) + (numerator_ratio * numerator_ratio) * (dx2 + dx2)

        product = f1 * r2
        if product < 0.0:
            return (omega5, self._error_threshold)

        return (omega5, 0.5 * math.sqrt(product))

    def _calc_model_params(self, omega: float) -> tuple[float, float, float, float, float]:
        x0 = self._buffer[0]
        xm1 = self._buffer[1]
        xm2 = self._buffer[2]

        half_w = omega / 2.0
        three_half_w = 1.5 * omega

        sin_hw = math.sin(half_w)
        cos_hw = math.cos(half_w)
        sin_3hw = math.sin(three_half_w)
        cos_3hw = math.cos(three_half_w)

        d0 = sin_hw * sin_hw * cos_hw * sin_3hw - sin_hw * sin_hw * sin_hw * cos_3hw

        nan = math.nan
        if abs(d0) < 1e-15:
            return (nan, nan, nan, nan, nan)

        inv_d0 = 1.0 / d0

        dx0_m1 = x0 - xm1
        dxm1_m2 = xm1 - xm2

        c = inv_d0 * (dx0_m1 * sin_hw * sin_3hw - dxm1_m2 * sin_hw * sin_hw)
        s = inv_d0 * (dxm1_m2 * sin_hw * cos_hw - dx0_m1 * sin_hw * cos_3hw)

        amplitude = 0.5 * math.sqrt(c * c + s * s)
        phi = math.atan2(s, c)
        velocity = amplitude * omega * math.cos(phi)
        acceleration = -amplitude * omega * omega * math.sin(phi)
        dc_level = x0 - s / 2.0

        return (amplitude, phi, velocity, acceleration, dc_level)

    def is_primed(self) -> bool:
        return self._primed

    def metadata(self) -> Metadata:
        desc = f"Instantaneous Sine Wave Period {self._mnemonic}"
        return build_metadata(
            Identifier.INSTANTANEOUS_SINE_WAVE_PERIOD,
            self._mnemonic,
            desc,
            [
                OutputText(f"{self._mnemonic} period", f"{desc} Period"),
                OutputText(f"{self._mnemonic} omega", f"{desc} Omega"),
                OutputText(f"{self._mnemonic} velocity", f"{desc} Velocity"),
                OutputText(f"{self._mnemonic} acceleration", f"{desc} Acceleration"),
                OutputText(f"{self._mnemonic} amplitude", f"{desc} Amplitude"),
                OutputText(f"{self._mnemonic} phase", f"{desc} Phase"),
                OutputText(f"{self._mnemonic} dc_level", f"{desc} DC Level"),
            ],
        )

    def update(self, price: float) -> tuple[float, float, float, float, float, float, float]:
        """Update with a scalar value.

        Returns (period, omega, velocity, acceleration, amplitude, phase, dc_level).
        """
        nan = math.nan

        smoothed = self._apply_ema(price) if self._smoothing > 0 else price

        self._push_buffer(smoothed)
        self._count += 1

        if self._count < 5:
            return nan, nan, nan, nan, nan, nan, nan

        omega4, error4 = self._calc_omega4()
        omega5, error5 = self._calc_omega5()

        if error4 >= self._error_threshold and error5 >= self._error_threshold:
            return nan, nan, nan, nan, nan, nan, nan

        omega = omega5 if error5 < error4 else omega4

        if math.isnan(omega) or omega <= 0.0:
            return nan, nan, nan, nan, nan, nan, nan

        period = (2.0 * math.pi) / omega
        if period < self._min_period or period > self._max_period:
            return nan, nan, nan, nan, nan, nan, nan

        amplitude, phi, velocity, acceleration, dc_level = self._calc_model_params(omega)

        self._primed = True

        return period, omega, velocity, acceleration, amplitude, phi, dc_level

    def update_scalar(self, sample: Scalar) -> List[Any]:
        period, omega, velocity, acceleration, amplitude, phase, dc_level = self.update(sample.value)
        return [
            Scalar(time=sample.time, value=period),
            Scalar(time=sample.time, value=omega),
            Scalar(time=sample.time, value=velocity),
            Scalar(time=sample.time, value=acceleration),
            Scalar(time=sample.time, value=amplitude),
            Scalar(time=sample.time, value=phase),
            Scalar(time=sample.time, value=dc_level),
        ]

    def update_bar(self, sample: Bar) -> List[Any]:
        v = self._bar_func(sample)
        return self.update_scalar(Scalar(time=sample.time, value=v))

    def update_quote(self, sample: Quote) -> List[Any]:
        v = self._quote_func(sample)
        return self.update_scalar(Scalar(time=sample.time, value=v))

    def update_trade(self, sample: Trade) -> List[Any]:
        v = self._trade_func(sample)
        return self.update_scalar(Scalar(time=sample.time, value=v))
