"""Adaptive Exponential Moving Average (AEMA) indicator -- Don Mak.

An EMA with a time-varying smoothing factor alpha that adapts based on the
instantaneous frequency of the price data. An embedded ISWP (Instantaneous Sine
Wave Period) estimator detects the dominant frequency at each bar.

Reference:
    Mak, D.K. (2006). Mathematical Techniques in Financial Market Trading.
    World Scientific. Chapter 3.6.

The indicator produces three outputs:
  - value: the adaptively smoothed price (never NaN);
  - omega: the instantaneous frequency estimate (may be NaN);
  - alpha: the smoothing factor used for this bar.
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
from .params import AdaptiveExponentialMovingAverageParams


class _InstantaneousSineWavePeriod:
    """Embedded ISWP omega estimator (omega-only reduction).

    Estimates the dominant circular frequency omega of price data by modeling it
    locally as a single sine wave, combining a 4-point and a 5-point method and
    selecting the one with the lower estimation error. Inlined so the indicator
    is a standalone porting unit. Do NOT change its numerics.
    """

    _MIN_PERIOD = 4.0
    _MAX_PERIOD = 50.0
    _ERROR_THRESHOLD = 20.0
    _DX = 0.01

    def __init__(self, smoothing: int) -> None:
        self._smoothing = smoothing
        self._ema_alpha = 2.0 / (smoothing + 1.0) if smoothing > 0 else 1.0
        self._ema_value = None
        self._buffer = [0.0] * 5
        self._count = 0

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
            return (math.nan, self._ERROR_THRESHOLD)

        ratio = (x0 - xm3) / den

        sqrt_arg = 3.0 - ratio
        if sqrt_arg < 0.0:
            return (math.nan, self._ERROR_THRESHOLD)

        sqrt_val = math.sqrt(sqrt_arg)
        arg = 0.5 * sqrt_val
        if arg > 1.0:
            return (math.nan, self._ERROR_THRESHOLD)

        omega4 = 2.0 * math.asin(arg)

        dx = self._DX
        dx2 = dx * dx

        denom1 = 1.0 - 0.25 * sqrt_arg
        if denom1 <= 0.0 or sqrt_arg == 0.0:
            return (omega4, self._ERROR_THRESHOLD)

        f1 = 1.0 / (denom1 * sqrt_arg)
        inv_den2 = 1.0 / (den * den)
        q2 = inv_den2 * (dx2 + dx2) + (ratio * ratio) * inv_den2 * (dx2 + dx2)

        product = f1 * q2
        if product < 0.0:
            return (omega4, self._ERROR_THRESHOLD)

        error4 = 0.5 * math.sqrt(product)
        return (omega4, error4)

    def _calc_omega5(self) -> tuple[float, float]:
        x0 = self._buffer[0]
        xm1 = self._buffer[1]
        xm3 = self._buffer[3]
        xm4 = self._buffer[4]

        den1 = xm1 - xm3
        if den1 == 0.0:
            return (math.nan, self._ERROR_THRESHOLD)

        arg = 0.5 * (x0 - xm4) / den1
        if abs(arg) > 1.0:
            return (math.nan, self._ERROR_THRESHOLD)

        omega5 = math.acos(arg)

        dx = self._DX
        dx2 = dx * dx

        denom = 1.0 - arg * arg
        if denom <= 0.0:
            return (omega5, self._ERROR_THRESHOLD)

        f1 = 1.0 / denom
        inv_den1_2 = 1.0 / (den1 * den1)
        numerator_ratio = (x0 - xm4) / (den1 * den1)
        r2 = inv_den1_2 * (dx2 + dx2) + (numerator_ratio * numerator_ratio) * (dx2 + dx2)

        product = f1 * r2
        if product < 0.0:
            return (omega5, self._ERROR_THRESHOLD)

        error5 = 0.5 * math.sqrt(product)
        return (omega5, error5)

    def update(self, price: float) -> float:
        """Process one price; return the omega estimate (NaN if unavailable)."""
        smoothed = self._apply_ema(price) if self._smoothing > 0 else price

        self._push_buffer(smoothed)
        self._count += 1

        if self._count < 5:
            return math.nan

        omega4, error4 = self._calc_omega4()
        omega5, error5 = self._calc_omega5()

        if error4 >= self._ERROR_THRESHOLD and error5 >= self._ERROR_THRESHOLD:
            return math.nan

        omega = omega5 if error5 < error4 else omega4

        if math.isnan(omega) or omega <= 0.0:
            return math.nan

        period = (2.0 * math.pi) / omega
        if period < self._MIN_PERIOD or period > self._MAX_PERIOD:
            return math.nan

        return omega


class AdaptiveExponentialMovingAverage(Indicator):
    """Don Mak's Adaptive Exponential Moving Average (AEMA)."""

    def __init__(self, p: AdaptiveExponentialMovingAverageParams) -> None:
        alpha_max = p.alpha_max
        alpha_min = p.alpha_min
        omega_0 = p.omega_0
        smoothing = p.smoothing

        if not (0.0 < alpha_min < alpha_max <= 1.0):
            raise ValueError(
                "invalid adaptive exponential moving average parameters: "
                "need 0 < alpha_min < alpha_max <= 1")
        if not (0.0 < omega_0 < math.pi):
            raise ValueError(
                "invalid adaptive exponential moving average parameters: "
                "need 0 < omega_0 < pi")
        if smoothing < 0:
            raise ValueError(
                "invalid adaptive exponential moving average parameters: "
                "smoothing must be >= 0")

        bc = p.bar_component if p.bar_component is not None else DEFAULT_BAR_COMPONENT
        qc = p.quote_component if p.quote_component is not None else DEFAULT_QUOTE_COMPONENT
        tc = p.trade_component if p.trade_component is not None else DEFAULT_TRADE_COMPONENT

        self._bar_func = bar_component_value(bc)
        self._quote_func = quote_component_value(qc)
        self._trade_func = trade_component_value(tc)

        self._alpha_max = alpha_max
        self._alpha_min = alpha_min
        self._omega_0 = omega_0

        # Hyperbolic interpolation constants (Eq 3.15a / 3.15b).
        self._a = (alpha_max - alpha_min) * omega_0 * math.pi / (math.pi - omega_0)
        self._b = alpha_min - self._a / math.pi

        self._iswp = _InstantaneousSineWavePeriod(smoothing)

        self._ema_value = 0.0
        self._initialized = False
        self._primed = False

        self._mnemonic = (
            f"aema({alpha_max:.2f},{alpha_min:.2f},{omega_0:.2f},{smoothing}"
            f"{component_triple_mnemonic(bc, qc, tc)})")

    def _compute_alpha(self, omega: float) -> float:
        if math.isnan(omega):
            return self._alpha_min
        if omega <= self._omega_0:
            return self._alpha_max
        if omega >= math.pi:
            return self._alpha_min

        alpha = self._a / omega + self._b
        if alpha > self._alpha_max:
            return self._alpha_max
        if alpha < self._alpha_min:
            return self._alpha_min
        return alpha

    def is_primed(self) -> bool:
        return self._primed

    def metadata(self) -> Metadata:
        desc = f"Adaptive Exponential Moving Average {self._mnemonic}"
        return build_metadata(
            Identifier.ADAPTIVE_EXPONENTIAL_MOVING_AVERAGE,
            self._mnemonic,
            desc,
            [
                OutputText(f"{self._mnemonic} value", f"{desc} Value"),
                OutputText(f"{self._mnemonic} omega", f"{desc} Omega"),
                OutputText(f"{self._mnemonic} alpha", f"{desc} Alpha"),
            ],
        )

    def update(self, price: float) -> tuple[float, float, float]:
        """Update with a scalar value. Returns (value, omega, alpha)."""
        omega = self._iswp.update(price)
        alpha = self._compute_alpha(omega)

        if not self._initialized:
            self._ema_value = price
            self._initialized = True
        else:
            self._ema_value = alpha * price + (1.0 - alpha) * self._ema_value

        if not math.isnan(omega):
            self._primed = True

        return self._ema_value, omega, alpha

    def update_scalar(self, sample: Scalar) -> List[Any]:
        value, omega, alpha = self.update(sample.value)
        return [
            Scalar(time=sample.time, value=value),
            Scalar(time=sample.time, value=omega),
            Scalar(time=sample.time, value=alpha),
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
