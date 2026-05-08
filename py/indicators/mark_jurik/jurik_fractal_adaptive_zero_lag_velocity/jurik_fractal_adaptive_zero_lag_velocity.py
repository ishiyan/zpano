"""Jurik fractal adaptive zero lag velocity indicator."""

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
from .params import JurikFractalAdaptiveZeroLagVelocityParams


SCALE_SETS = {
    1: [2, 3, 4, 6, 8, 12, 16, 24],
    2: [2, 3, 4, 6, 8, 12, 16, 24, 32, 48],
    3: [2, 3, 4, 6, 8, 12, 16, 24, 32, 48, 64, 96],
    4: [2, 3, 4, 6, 8, 12, 16, 24, 32, 48, 64, 96, 128, 192],
}

_WEIGHTS_EVEN = [2, 3, 6, 12, 24, 48, 96]
_WEIGHTS_ODD = [4, 8, 16, 32, 64, 128, 256]


class _CfbAux:
    """Streaming state for a single JCFBaux(depth) instance."""

    def __init__(self, depth: int) -> None:
        self._depth = depth
        self._bar = 0
        self._int_a = [0.0] * depth
        self._int_a_idx = 0
        self._src = [0.0] * (depth + 2)
        self._src_idx = 0
        self._jrc04 = 0.0
        self._jrc05 = 0.0
        self._jrc06 = 0.0
        self._prev_sample = 0.0
        self._first_call = True

    def update(self, sample: float) -> float:
        self._bar += 1
        depth = self._depth
        src_size = depth + 2

        self._src[self._src_idx] = sample
        self._src_idx = (self._src_idx + 1) % src_size

        if self._first_call:
            self._first_call = False
            self._prev_sample = sample
            return 0.0

        int_a_val = abs(sample - self._prev_sample)
        self._prev_sample = sample

        old_int_a = self._int_a[self._int_a_idx]
        self._int_a[self._int_a_idx] = int_a_val
        self._int_a_idx = (self._int_a_idx + 1) % depth

        ref_bar = self._bar - 1
        if ref_bar < depth:
            return 0.0

        if ref_bar <= depth * 2:
            # Recompute from scratch.
            self._jrc04 = 0.0
            self._jrc05 = 0.0
            self._jrc06 = 0.0

            cur_int_a_pos = (self._int_a_idx - 1 + depth) % depth
            cur_src_pos = (self._src_idx - 1 + src_size) % src_size

            for j in range(depth):
                int_a_pos = (cur_int_a_pos - j + depth) % depth
                int_a_v = self._int_a[int_a_pos]

                src_pos = (cur_src_pos - j - 1 + src_size * 2) % src_size
                src_v = self._src[src_pos]

                self._jrc04 += int_a_v
                self._jrc05 += float(depth - j) * int_a_v
                self._jrc06 += src_v
        else:
            # Incremental update.
            self._jrc05 = self._jrc05 - self._jrc04 + int_a_val * float(depth)
            self._jrc04 = self._jrc04 - old_int_a + int_a_val

            cur_src_pos = (self._src_idx - 1 + src_size) % src_size
            src_bar_minus1 = (cur_src_pos - 1 + src_size) % src_size
            src_bar_minus_depth_minus1 = (cur_src_pos - depth - 1 + src_size) % src_size

            self._jrc06 = self._jrc06 - self._src[src_bar_minus_depth_minus1] + \
                self._src[src_bar_minus1]

        cur_src_pos = (self._src_idx - 1 + src_size) % src_size
        jrc08 = abs(float(depth) * self._src[cur_src_pos] - self._jrc06)

        if self._jrc05 == 0.0:
            return 0.0

        return jrc08 / self._jrc05


class _Cfb:
    """Composite Fractal Behavior: weighted dominant cycle from multi-scale ERs."""

    def __init__(self, fractal_type: int, smooth: int) -> None:
        scales = SCALE_SETS[fractal_type]
        self._scales = scales
        self._num_channels = len(scales)
        self._aux_instances = [_CfbAux(d) for d in scales]
        self._aux_windows = [[0.0] * smooth for _ in range(self._num_channels)]
        self._aux_win_idx = 0
        self._er23 = [0.0] * self._num_channels
        self._smooth = smooth
        self._bar = 0
        self._cfb_value = 0.0

    def update(self, sample: float) -> float:
        """Return composite CFB value. Always returns a value (may be 0 initially)."""
        self._bar += 1
        ref_bar = self._bar - 1

        # Feed all aux instances.
        aux_values = [aux.update(sample) for aux in self._aux_instances]

        if ref_bar == 0:
            return 0.0

        smooth = self._smooth
        n = self._num_channels

        if ref_bar <= smooth:
            # Growing window.
            win_pos = self._aux_win_idx
            for i in range(n):
                self._aux_windows[i][win_pos] = aux_values[i]
            self._aux_win_idx = (self._aux_win_idx + 1) % smooth

            # Recompute sums from scratch.
            for i in range(n):
                s = 0.0
                for j in range(ref_bar):
                    pos = (self._aux_win_idx - 1 - j + smooth * 2) % smooth
                    s += self._aux_windows[i][pos]
                self._er23[i] = s / float(ref_bar)
        else:
            # Sliding window.
            win_pos = self._aux_win_idx
            for i in range(n):
                old_val = self._aux_windows[i][win_pos]
                self._aux_windows[i][win_pos] = aux_values[i]
                self._er23[i] += (aux_values[i] - old_val) / float(smooth)
            self._aux_win_idx = (self._aux_win_idx + 1) % smooth

        # Compute weighted composite (only when refBar > 5).
        if ref_bar > 5:
            er22 = [0.0] * n

            # Odd-indexed channels (descending).
            er15 = 1.0
            for idx in range(n - 1, 0, -2):
                er22[idx] = er15 * self._er23[idx]
                er15 *= (1 - er22[idx])

            # Even-indexed channels (descending).
            er16 = 1.0
            for idx in range(n - 2, -1, -2):
                er22[idx] = er16 * self._er23[idx]
                er16 *= (1 - er22[idx])

            # Weighted sum using fixed weight arrays (same as JCFB).
            er17 = 0.0
            er18 = 0.0
            for idx in range(n):
                sq = er22[idx] * er22[idx]
                er18 += sq
                if idx % 2 == 0:
                    er17 += sq * _WEIGHTS_EVEN[idx // 2]
                else:
                    er17 += sq * _WEIGHTS_ODD[idx // 2]

            if er18 == 0.0:
                self._cfb_value = 0.0
            else:
                self._cfb_value = er17 / er18

        return self._cfb_value


class _VelSlope:
    """WLS slope computation (Stage 1) — identical to JAVEL."""

    def compute(self, prices: list[float], bar: int, depth: int) -> float | None:
        if bar < depth:
            return None
        n = depth + 1
        s1 = n * (n + 1) / 2.0
        s2 = s1 * (2 * n + 1) / 3.0
        denom = s1 * s1 * s1 - s2 * s2

        sum_xw = 0.0
        sum_xw2 = 0.0
        for i in range(depth + 1):
            w = float(n - i)
            p = prices[bar - i]
            sum_xw += p * w
            sum_xw2 += p * w * w

        return (sum_xw2 * s1 - sum_xw * s2) / denom


class _VelSmooth:
    """Adaptive smoother (Stage 2) — identical to JAVEL with fixed period=3.0."""

    def __init__(self, period: float) -> None:
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

        self._buffer = [0.0] * 1001
        self._idx = 0
        self._length = 0
        self._velocity = 0.0
        self._position = 0.0
        self._smoothed_mad = 0.0
        self._mad_initialized = False
        self._initialized = False

    def update(self, value: float) -> float:
        """Feed a slope value into the adaptive smoother."""
        self._buffer[self._idx] = value
        self._idx = (self._idx + 1) % self._buffer_size
        self._length = min(self._length + 1, self._buffer_size)

        length = self._length

        if not self._initialized:
            self._initialized = True
            self._position = value
            self._velocity = 0.0
            self._smoothed_mad = 0.0
            return self._position

        # Linear regression over capped window.
        n = min(length, self._jrc06)
        sx = 0.0
        sy = 0.0
        sxy = 0.0
        sx2 = 0.0
        for i in range(n):
            idx = (self._idx - 1 - i + self._buffer_size) % self._buffer_size
            x = float(i)
            y = self._buffer[idx]
            sx += x
            sy += y
            sxy += x * y
            sx2 += x * x

        fn = float(n)
        slope = (fn * sxy - sx * sy) / (fn * sx2 - sx * sx) if n > 1 else 0.0
        intercept = (sy - slope * sx) / fn

        # MAD from regression residuals.
        mad = 0.0
        for i in range(n):
            idx = (self._idx - 1 - i + self._buffer_size) % self._buffer_size
            predicted = intercept + slope * float(i)
            mad += abs(self._buffer[idx] - predicted)
        mad /= fn

        # Scale MAD.
        scaled_mad = mad * 1.2 * pow(float(self._jrc06) / fn, 0.25)

        # Smooth MAD with EMA (seed on first non-zero).
        if not self._mad_initialized:
            self._smoothed_mad = scaled_mad
            if scaled_mad > 0:
                self._mad_initialized = True
        else:
            self._smoothed_mad += (scaled_mad - self._smoothed_mad) * self._ema_factor
        smoothed_mad = max(self._eps2, self._smoothed_mad)

        # Adaptive velocity/position dynamics.
        prediction_error = value - self._position
        response_factor = 1.0 - math.exp(-abs(prediction_error) / \
            (smoothed_mad * self._jrc03))
        self._velocity = response_factor * prediction_error + \
            self._velocity * self._damping
        self._position += self._velocity

        return self._position


class JurikFractalAdaptiveZeroLagVelocity(Indicator):
    """Computes the Jurik Fractal Adaptive Zero Lag Velocity (JVELCFB) indicator.

    Combines CFB (Composite Fractal Behavior) cycle detection with VEL
    (two-stage velocity). CFB estimates dominant cycle period, stochastic
    normalization maps it to a depth range, then VEL computes adaptive velocity.
    """

    def __init__(self, params: JurikFractalAdaptiveZeroLagVelocityParams) -> None:
        lo_depth = params.lo_depth
        hi_depth = params.hi_depth
        fractal_type = params.fractal_type
        smooth = params.smooth

        if lo_depth < 2:
            raise ValueError(
                "invalid jurik fractal adaptive zero lag velocity parameters: "
                "lo_depth should be at least 2")
        if hi_depth < lo_depth:
            raise ValueError(
                "invalid jurik fractal adaptive zero lag velocity parameters: "
                "hi_depth should be at least lo_depth")
        if fractal_type < 1 or fractal_type > 4:
            raise ValueError(
                "invalid jurik fractal adaptive zero lag velocity parameters: "
                "fractal_type should be 1-4")
        if smooth < 1:
            raise ValueError(
                "invalid jurik fractal adaptive zero lag velocity parameters: "
                "smooth should be at least 1")

        bc = params.bar_component if params.bar_component is not None else DEFAULT_BAR_COMPONENT
        qc = params.quote_component if params.quote_component is not None else DEFAULT_QUOTE_COMPONENT
        tc = params.trade_component if params.trade_component is not None else DEFAULT_TRADE_COMPONENT

        bar_func = bar_component_value(bc)
        quote_func = quote_component_value(qc)
        trade_func = trade_component_value(tc)

        mnemonic = f"jvelcfb({lo_depth}, {hi_depth}, {fractal_type}, {smooth}" \
                   f"{component_triple_mnemonic(bc, qc, tc)})"
        description = f"Jurik fractal adaptive zero lag velocity {mnemonic}"

        self._line = LineIndicator(mnemonic, description, bar_func, quote_func, trade_func, self.update)
        self._primed = False
        self._lo_depth = lo_depth
        self._hi_depth = hi_depth

        # Price history.
        self._prices: list[float] = []
        self._bar_count = 0

        # CFB.
        self._cfb = _Cfb(fractal_type, smooth)

        # Stochastic normalization state.
        self._cfb_min: float | None = None
        self._cfb_max: float | None = None

        # Stage 1 and Stage 2.
        self._vel_slope = _VelSlope()
        self._vel_smooth = _VelSmooth(3.0)

    def is_primed(self) -> bool:
        return self._primed

    def metadata(self) -> Metadata:
        return build_metadata(
            Identifier.JURIK_FRACTAL_ADAPTIVE_ZERO_LAG_VELOCITY,
            self._line.mnemonic,
            self._line.description,
            [OutputText(mnemonic=self._line.mnemonic, description=self._line.description)],
        )

    def update(self, sample: float) -> float:
        """Update the indicator with a new sample value."""
        if math.isnan(sample):
            return sample

        bar = self._bar_count
        self._bar_count += 1

        self._prices.append(sample)

        # CFB computation.
        cfb = self._cfb.update(sample)

        # Skip first bar (cfb not meaningful).
        if bar == 0:
            return math.nan

        # Stochastic normalization.
        if self._cfb_min is None:
            self._cfb_min = cfb
            self._cfb_max = cfb
        else:
            if cfb < self._cfb_min:
                self._cfb_min = cfb
            if cfb > self._cfb_max:
                self._cfb_max = cfb

        cfb_range = self._cfb_max - self._cfb_min
        if cfb_range != 0.0:
            sr = (cfb - self._cfb_min) / cfb_range
        else:
            sr = 0.5

        depth_f = self._lo_depth + sr * (self._hi_depth - self._lo_depth)
        depth = round(depth_f)

        # Stage 1: WLS slope.
        slope = self._vel_slope.compute(self._prices, bar, depth)
        if slope is None:
            return math.nan

        # Stage 2: adaptive smoother.
        result = self._vel_smooth.update(slope)

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
