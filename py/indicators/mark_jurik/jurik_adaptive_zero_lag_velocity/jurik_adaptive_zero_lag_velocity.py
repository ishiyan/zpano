"""Jurik adaptive zero lag velocity indicator."""

import math

from ...core.indicator import Indicator
from ...core.line_indicator import LineIndicator
from ...core.metadata import Metadata
from ...core.build_metadata import build_metadata, OutputText
from ...core.identifier import Identifier
from ...core.component_triple_mnemonic import component_triple_mnemonic
from ...core.output import Output
from ....entities.bar import Bar
from ....entities.quote import Quote
from ....entities.trade import Trade
from ....entities.scalar import Scalar
from ....entities.bar_component import BarComponent, DEFAULT_BAR_COMPONENT, bar_component_value
from ....entities.quote_component import QuoteComponent, DEFAULT_QUOTE_COMPONENT, quote_component_value
from ....entities.trade_component import TradeComponent, DEFAULT_TRADE_COMPONENT, trade_component_value
from .params import JurikAdaptiveZeroLagVelocityParams


class JurikAdaptiveZeroLagVelocity(Indicator):
    """Computes the Jurik Adaptive Zero Lag Velocity (JAVEL) indicator.

    Combines adaptive depth selection (volatility regime detection) with
    two-stage velocity computation (WLS slope + adaptive smoother).
    """

    def __init__(self, params: JurikAdaptiveZeroLagVelocityParams) -> None:
        lo_length = params.lo_length
        hi_length = params.hi_length
        sensitivity = params.sensitivity
        period = params.period

        if lo_length < 2:
            raise ValueError(
                "invalid jurik adaptive zero lag velocity parameters: "
                "lo_length should be at least 2")
        if hi_length < lo_length:
            raise ValueError(
                "invalid jurik adaptive zero lag velocity parameters: "
                "hi_length should be at least lo_length")
        if period <= 0:
            raise ValueError(
                "invalid jurik adaptive zero lag velocity parameters: "
                "period should be positive")

        bc = params.bar_component if params.bar_component is not None else DEFAULT_BAR_COMPONENT
        qc = params.quote_component if params.quote_component is not None else DEFAULT_QUOTE_COMPONENT
        tc = params.trade_component if params.trade_component is not None else DEFAULT_TRADE_COMPONENT

        bar_func = bar_component_value(bc)
        quote_func = quote_component_value(qc)
        trade_func = trade_component_value(tc)

        mnemonic = f"javel({lo_length}, {hi_length}, {sensitivity}, {period}" \
                   f"{component_triple_mnemonic(bc, qc, tc)})"
        description = f"Jurik adaptive zero lag velocity {mnemonic}"

        self._line = LineIndicator(mnemonic, description, bar_func, quote_func, trade_func, self.update)
        self._primed = False
        self._lo_length = lo_length
        self._hi_length = hi_length
        self._sensitivity = sensitivity
        self._period = period
        self._eps = 0.001

        # Price history for WLS and adaptive depth.
        self._prices: list[float] = []
        self._bar_count = 0

        # Adaptive depth state: rolling averages of abs(diff).
        self._value1: list[float] = []

        # Stage 2: adaptive smoother state.
        eps2 = 0.0001
        jrc03 = min(500.0, max(eps2, period))
        jrc06 = max(31, math.ceil(2 * period))
        jrc07 = min(30, math.ceil(period))
        ema_factor = 1.0 - math.exp(-math.log(4.0) / (period / 2.0))
        damping = 0.86 - 0.55 / math.sqrt(jrc03)

        self._jrc03 = jrc03
        self._jrc06 = jrc06
        self._jrc07 = jrc07
        self._ema_factor = ema_factor
        self._damping = damping
        self._eps2 = eps2
        self._buffer_size = 1001

        # Stage 2 circular buffer and state.
        self._s2_buffer = [0.0] * 1001
        self._s2_head = 0
        self._s2_length = 0
        self._s2_bar_count = 0
        self._s2_velocity = 0.0
        self._s2_position = 0.0
        self._s2_smoothed_mad = 0.0
        self._s2_initialized = False

    def is_primed(self) -> bool:
        return self._primed

    def metadata(self) -> Metadata:
        return build_metadata(
            Identifier.JURIK_ADAPTIVE_ZERO_LAG_VELOCITY,
            self._line.mnemonic,
            self._line.description,
            [OutputText(mnemonic=self._line.mnemonic, description=self._line.description)],
        )

    def _compute_adaptive_depth(self, bar: int) -> float:
        """Compute adaptive depth for the current bar."""
        long_window = min(bar, 99) + 1
        short_window = min(bar, 9) + 1

        avg1 = sum(self._value1[bar - long_window + 1:bar + 1]) / long_window
        avg2 = sum(self._value1[bar - short_window + 1:bar + 1]) / short_window

        value2 = self._sensitivity * \
            math.log((self._eps + avg1) / (self._eps + avg2))
        value3 = value2 / (1.0 + abs(value2))
        return self._lo_length + \
            (self._hi_length - self._lo_length) * (1.0 + value3) / 2.0

    def _compute_wls_slope(self, bar: int, depth: int) -> float:
        """Compute WLS slope over depth+1 most recent prices."""
        n = depth + 1
        s1 = n * (n + 1) / 2.0
        s2 = s1 * (2 * n + 1) / 3.0
        denom = s1 * s1 * s1 - s2 * s2

        sum_xw = 0.0
        sum_xw2 = 0.0
        for i in range(depth + 1):
            w = float(n - i)
            p = self._prices[bar - i]
            sum_xw += p * w
            sum_xw2 += p * w * w

        return (sum_xw2 * s1 - sum_xw * s2) / denom

    def _stage2_update(self, value: float) -> float:
        """Feed a slope value into the adaptive smoother (stage 2).

        Mirrors the reference jxvel_smooth exactly:
        - Circular buffer with forward indexing
        - Length capped at jrc06
        - MAD seeded for first jrc07+1 bars
        - Response factor with velocity/position dynamics
        """
        self._s2_bar_count += 1

        # Store in circular buffer (same as reference: value_buffer[old_index] = value).
        old_index = self._s2_head % self._buffer_size
        self._s2_buffer[old_index] = value
        self._s2_head += 1

        if self._s2_length < self._jrc06:
            self._s2_length += 1

        length = self._s2_length

        # First bar: initialize position.
        if length < 2:
            if not self._s2_initialized:
                self._s2_position = value
                self._s2_initialized = True
            return self._s2_position

        if not self._s2_initialized:
            self._s2_position = value
            self._s2_initialized = True

        # Linear regression over buffer (forward: k=0 oldest, k=length-1 newest).
        sum_values = 0.0
        sum_weighted = 0.0
        for k in range(length):
            idx = (self._s2_head - length + k) % self._buffer_size
            sum_values += self._s2_buffer[idx]
            sum_weighted += self._s2_buffer[idx] * k

        midpoint = (length - 1) / 2.0
        sum_x_sq = length * (length - 1) * (2 * length - 1) / 6.0
        regression_denom = sum_x_sq - length * midpoint * midpoint

        if abs(regression_denom) < self._eps2:
            regression_slope = 0.0
        else:
            regression_slope = (sum_weighted - midpoint * sum_values) / regression_denom

        intercept = sum_values / length - regression_slope * midpoint

        # Compute MAD from regression residuals.
        sum_abs_dev = 0.0
        for k in range(length):
            idx = (self._s2_head - length + k) % self._buffer_size
            predicted = intercept + regression_slope * k
            sum_abs_dev += abs(self._s2_buffer[idx] - predicted)

        raw_mad = sum_abs_dev / length
        scale = 1.2 * (self._jrc06 / length) ** 0.25
        raw_mad *= scale

        # Smooth MAD with EMA (seed for first jrc07+1 bars).
        if self._s2_bar_count <= self._jrc07 + 1:
            self._s2_smoothed_mad = raw_mad
        else:
            self._s2_smoothed_mad += self._ema_factor * (raw_mad - self._s2_smoothed_mad)

        # Adaptive velocity/position dynamics.
        prediction_error = value - self._s2_position

        if self._s2_smoothed_mad * self._jrc03 < self._eps2:
            response_factor = 1.0
        else:
            response_factor = 1.0 - math.exp(
                -abs(prediction_error) / (self._s2_smoothed_mad * self._jrc03))

        self._s2_velocity = response_factor * prediction_error + \
            self._s2_velocity * self._damping
        self._s2_position += self._s2_velocity

        return self._s2_position

    def update(self, sample: float) -> float:
        """Update the indicator with a new sample value."""
        if math.isnan(sample):
            return sample

        bar = self._bar_count
        self._bar_count += 1

        # Store price.
        self._prices.append(sample)

        # Compute value1 (abs diff).
        if bar == 0:
            self._value1.append(0.0)
        else:
            self._value1.append(abs(sample - self._prices[bar - 1]))

        # Compute adaptive depth.
        adaptive_depth = self._compute_adaptive_depth(bar)
        depth = math.ceil(adaptive_depth)

        # Check if we have enough prices for WLS.
        if bar < depth:
            return math.nan

        # Stage 1: WLS slope.
        slope = self._compute_wls_slope(bar, depth)

        # Stage 2: adaptive smoother.
        result = self._stage2_update(slope)

        if not self._primed:
            self._primed = True

        return result

    def update_bar(self, sample: Bar) -> Output:
        return self._line.update_bar(sample)

    def update_quote(self, sample: Quote) -> Output:
        return self._line.update_quote(sample)

    def update_trade(self, sample: Trade) -> Output:
        return self._line.update_trade(sample)

    def update_scalar(self, sample: Scalar) -> Output:
        return self._line.update_scalar(sample)
