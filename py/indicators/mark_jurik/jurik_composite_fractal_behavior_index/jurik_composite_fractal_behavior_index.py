"""Jurik composite fractal behavior index indicator."""

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
from .params import JurikCompositeFractalBehaviorIndexParams


# Fractal depths for each type (1-4).
_DEPTH_SETS = [
    [2, 3, 4, 6, 8, 12, 16, 24],                           # Type 1: JCFB24
    [2, 3, 4, 6, 8, 12, 16, 24, 32, 48],                   # Type 2: JCFB48
    [2, 3, 4, 6, 8, 12, 16, 24, 32, 48, 64, 96],           # Type 3: JCFB96
    [2, 3, 4, 6, 8, 12, 16, 24, 32, 48, 64, 96, 128, 192], # Type 4: JCFB192
]

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


class JurikCompositeFractalBehaviorIndex(Indicator):
    """Computes the Jurik CFB indicator.

    CFB measures composite fractal behavior across multiple time depths.
    """

    def __init__(self, params: JurikCompositeFractalBehaviorIndexParams) -> None:
        fractal_type = params.fractal_type
        smooth = params.smooth

        if fractal_type < 1 or fractal_type > 4:
            raise ValueError(
                "invalid jurik composite fractal behavior index parameters: "
                "fractal type should be between 1 and 4")

        if smooth < 1:
            raise ValueError(
                "invalid jurik composite fractal behavior index parameters: "
                "smooth should be at least 1")

        bc = params.bar_component if params.bar_component is not None else DEFAULT_BAR_COMPONENT
        qc = params.quote_component if params.quote_component is not None else DEFAULT_QUOTE_COMPONENT
        tc = params.trade_component if params.trade_component is not None else DEFAULT_TRADE_COMPONENT

        bar_func = bar_component_value(bc)
        quote_func = quote_component_value(qc)
        trade_func = trade_component_value(tc)

        mnemonic = f"jcfb({fractal_type},{smooth}{component_triple_mnemonic(bc, qc, tc)})"
        description = f"Jurik composite fractal behavior index {mnemonic}"

        self._line = LineIndicator(mnemonic, description, bar_func, quote_func, trade_func, self.update)
        self._primed = False
        self._param_fractal = fractal_type
        self._param_smooth = smooth

        depths = _DEPTH_SETS[fractal_type - 1]
        self._num_channels = len(depths)
        self._aux_instances = [_CfbAux(d) for d in depths]
        self._aux_windows = [[0.0] * smooth for _ in range(self._num_channels)]
        self._aux_win_idx = 0
        self._aux_win_len = 0
        self._er23 = [0.0] * self._num_channels
        self._bar = 0
        self._er19 = 20.0

    def is_primed(self) -> bool:
        return self._primed

    def metadata(self) -> Metadata:
        return build_metadata(
            Identifier.JURIK_COMPOSITE_FRACTAL_BEHAVIOR_INDEX,
            self._line.mnemonic,
            self._line.description,
            [OutputText(mnemonic=self._line.mnemonic, description=self._line.description)],
        )

    def update(self, sample: float) -> float:
        """Update the indicator with a new sample value."""
        if math.isnan(sample):
            return sample

        self._bar += 1

        # Feed all aux instances.
        aux_values = [aux.update(sample) for aux in self._aux_instances]

        # Bar 0 in reference outputs 0.0 → NaN for streaming.
        if self._bar == 1:
            return math.nan

        ref_bar = self._bar - 1
        smooth = self._param_smooth
        n = self._num_channels

        if ref_bar <= smooth:
            # Growing window.
            win_pos = self._aux_win_idx
            for i in range(n):
                self._aux_windows[i][win_pos] = aux_values[i]
            self._aux_win_idx = (self._aux_win_idx + 1) % smooth
            self._aux_win_len = ref_bar

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

            # Weighted sum.
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
                self._er19 = 0.0
            else:
                self._er19 = er17 / er18

        if not self._primed:
            if ref_bar > 5:
                self._primed = True

        return self._er19

    def update_bar(self, sample: Bar) -> Output:
        return self._line.update_bar(sample)

    def update_quote(self, sample: Quote) -> Output:
        return self._line.update_quote(sample)

    def update_trade(self, sample: Trade) -> Output:
        return self._line.update_trade(sample)

    def update_scalar(self, sample: Scalar) -> Output:
        return self._line.update_scalar(sample)
