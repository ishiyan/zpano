"""Jurik zero lag velocity indicator."""

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
from .params import JurikZeroLagVelocityParams


class _VelAux1:
    """Computes linear regression slope over a window of depth+1 points."""

    def __init__(self, depth: int) -> None:
        self._depth = depth
        size = depth + 1
        self._win = [0.0] * size
        self._idx = 0
        self._bar = 0

        jrc04 = float(size)
        jrc05 = jrc04 * (jrc04 + 1) / 2
        jrc06 = jrc05 * (2 * jrc04 + 1) / 3
        self._jrc04 = jrc04
        self._jrc05 = jrc05
        self._jrc06 = jrc06
        self._jrc07 = jrc05 * jrc05 * jrc05 - jrc06 * jrc06

    def update(self, sample: float) -> float:
        size = self._depth + 1
        self._win[self._idx] = sample
        self._idx = (self._idx + 1) % size
        self._bar += 1

        if self._bar <= self._depth:
            return 0.0

        jrc08 = 0.0
        jrc09 = 0.0
        for j in range(self._depth + 1):
            pos = (self._idx - 1 - j + size * 2) % size
            w = self._jrc04 - float(j)
            jrc08 += self._win[pos] * w
            jrc09 += self._win[pos] * w * w

        return (jrc09 * self._jrc05 - jrc08 * self._jrc06) / self._jrc07


class _VelAux3State:
    """Adaptive smoother for the VEL indicator."""

    def __init__(self) -> None:
        self._length = 30
        self._eps = 0.0001
        self._decay = 3
        self._beta = 0.86 - 0.55 / math.sqrt(3.0)
        self._alpha = 1 - math.exp(-math.log(4) / 3.0 / 2.0)
        self._max_win = 31  # length + 1

        self._src_ring = [0.0] * 100
        self._dev_ring = [0.0] * 100
        self._src_idx = 0
        self._dev_idx = 0

        self._jr08 = 0.0
        self._jr09 = 0.0
        self._jr10 = 0.0
        self._jr11 = 0
        self._jr12 = 0.0
        self._jr13 = 0.0
        self._jr14 = 0.0
        self._jr19 = 0.0
        self._jr20 = 0.0
        self._jr21 = 0.0
        self._jr21a = 0.0
        self._jr21b = 0.0
        self._jr22 = 0.0
        self._jr23 = 0.0

        self._bar = 0
        self._init_done = False
        self._history: list[float] = []

    def feed(self, sample: float, bar_idx: int) -> float:
        if bar_idx < self._length:
            self._history.append(sample)
            return 0.0

        self._bar += 1

        if not self._init_done:
            self._init_done = True

            jr28 = 0.0
            for j in range(1, self._length):
                if self._history[-j] == self._history[-j - 1]:
                    jr28 += 1.0

            if jr28 < float(self._length - 1):
                jr26 = bar_idx - self._length
            else:
                jr26 = bar_idx

            self._jr11 = int(math.trunc(min(1 + float(bar_idx - jr26), float(self._max_win))))

            self._jr21 = self._history[-1]
            jr07 = 3
            self._jr08 = (sample - self._history[-jr07]) / float(jr07)

            for jr15 in range(self._jr11 - 1, 0, -1):
                if self._src_idx <= 0:
                    self._src_idx = 100
                self._src_idx -= 1
                self._src_ring[self._src_idx] = self._history[-jr15]

            self._history = []

        # Push current value to source ring.
        if self._src_idx <= 0:
            self._src_idx = 100
        self._src_idx -= 1
        self._src_ring[self._src_idx] = sample

        if self._jr11 <= self._length:
            # Growing phase.
            if self._bar == 1:
                self._jr21 = sample
            else:
                self._jr21 = math.sqrt(self._alpha) * sample + \
                    (1 - math.sqrt(self._alpha)) * self._jr21a

            if self._bar > 2:
                self._jr08 = (self._jr21 - self._jr21b) / 2
            else:
                self._jr08 = 0.0

            self._jr11 += 1

        elif self._jr11 <= self._max_win:
            # Transition phase.
            self._jr12 = float(self._jr11 * (self._jr11 + 1) * (self._jr11 - 1)) / 12.0
            self._jr13 = float(self._jr11 + 1) / 2.0
            self._jr14 = float(self._jr11 - 1) / 2.0

            self._jr09 = 0.0
            self._jr10 = 0.0

            for jr15 in range(self._jr11 - 1, -1, -1):
                jr24 = (self._src_idx + jr15) % 100
                self._jr09 += self._src_ring[jr24]
                self._jr10 += self._src_ring[jr24] * (self._jr14 - float(jr15))

            jr16 = self._jr10 / self._jr12
            jr17 = (self._jr09 / float(self._jr11)) - (jr16 * self._jr13)

            self._jr19 = 0.0
            for jr15 in range(self._jr11 - 1, -1, -1):
                jr17 += jr16
                jr24 = (self._src_idx + jr15) % 100
                self._jr19 += abs(self._src_ring[jr24] - jr17)

            self._jr20 = (self._jr19 / float(self._jr11)) * \
                pow(float(self._max_win) / float(self._jr11), 0.25)
            self._jr11 += 1

            # Adaptive step.
            self._jr20 = max(self._eps, self._jr20)
            self._jr22 = sample - (self._jr21 + self._jr08 * self._beta)
            self._jr23 = 1 - math.exp(-abs(self._jr22) / self._jr20 / float(self._decay))
            self._jr08 = self._jr23 * self._jr22 + self._jr08 * self._beta
            self._jr21 += self._jr08

        else:
            # Steady state.
            jr24out = (self._src_idx + self._max_win) % 100
            self._jr10 = self._jr10 - self._jr09 + \
                self._src_ring[jr24out] * self._jr13 + sample * self._jr14
            self._jr09 = self._jr09 - self._src_ring[jr24out] + sample

            # Deviation ring update.
            if self._dev_idx <= 0:
                self._dev_idx = self._max_win
            self._dev_idx -= 1
            self._jr19 -= self._dev_ring[self._dev_idx]

            jr16 = self._jr10 / self._jr12
            jr17 = (self._jr09 / float(self._max_win)) + (jr16 * self._jr14)
            self._dev_ring[self._dev_idx] = abs(sample - jr17)
            self._jr19 = max(self._eps, self._jr19 + self._dev_ring[self._dev_idx])
            self._jr20 += ((self._jr19 / float(self._max_win)) - self._jr20) * self._alpha

            # Adaptive step.
            self._jr20 = max(self._eps, self._jr20)
            self._jr22 = sample - (self._jr21 + self._jr08 * self._beta)
            self._jr23 = 1 - math.exp(-abs(self._jr22) / self._jr20 / float(self._decay))
            self._jr08 = self._jr23 * self._jr22 + self._jr08 * self._beta
            self._jr21 += self._jr08

        self._jr21b = self._jr21a
        self._jr21a = self._jr21

        return self._jr21


class JurikZeroLagVelocity(Indicator):
    """Computes the Jurik zero lag velocity (VEL) indicator.

    The indicator is not primed during the first 30 updates.
    """

    def __init__(self, params: JurikZeroLagVelocityParams) -> None:
        depth = params.depth

        if depth < 2:
            raise ValueError(
                "invalid jurik zero lag velocity parameters: depth should be at least 2")

        bc = params.bar_component if params.bar_component is not None else DEFAULT_BAR_COMPONENT
        qc = params.quote_component if params.quote_component is not None else DEFAULT_QUOTE_COMPONENT
        tc = params.trade_component if params.trade_component is not None else DEFAULT_TRADE_COMPONENT

        bar_func = bar_component_value(bc)
        quote_func = quote_component_value(qc)
        trade_func = trade_component_value(tc)

        mnemonic = f"jvel({depth}{component_triple_mnemonic(bc, qc, tc)})"
        description = f"Jurik zero lag velocity {mnemonic}"

        self._line = LineIndicator(mnemonic, description, bar_func, quote_func, trade_func, self.update)
        self._primed = False
        self._aux1 = _VelAux1(depth)
        self._aux3 = _VelAux3State()
        self._bar = 0

    def is_primed(self) -> bool:
        return self._primed

    def metadata(self) -> Metadata:
        return build_metadata(
            Identifier.JURIK_ZERO_LAG_VELOCITY,
            self._line.mnemonic,
            self._line.description,
            [OutputText(mnemonic=self._line.mnemonic, description=self._line.description)],
        )

    def update(self, sample: float) -> float:
        """Update the indicator with a new sample value."""
        if math.isnan(sample):
            return sample

        # Stage 1: compute linear regression slope.
        aux1_val = self._aux1.update(sample)

        # Stage 2: feed into adaptive smoother.
        bar_idx = self._bar
        self._bar += 1

        result = self._aux3.feed(aux1_val, bar_idx)

        # Output is 0 during warmup → NaN.
        if bar_idx < self._aux3._length:
            return math.nan

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
