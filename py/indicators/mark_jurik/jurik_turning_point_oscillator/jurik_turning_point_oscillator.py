"""Jurik turning point oscillator indicator."""

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
from .params import JurikTurningPointOscillatorParams


class JurikTurningPointOscillator(Indicator):
    """Computes Spearman rank correlation between price ranks and time positions.

    Output is in [-1, +1].
    """

    def __init__(self, params: JurikTurningPointOscillatorParams) -> None:
        length = params.length

        if length < 2:
            raise ValueError(
                "invalid jurik turning point oscillator parameters: "
                "length should be at least 2")

        bc = params.bar_component if params.bar_component is not None else DEFAULT_BAR_COMPONENT
        qc = params.quote_component if params.quote_component is not None else DEFAULT_QUOTE_COMPONENT
        tc = params.trade_component if params.trade_component is not None else DEFAULT_TRADE_COMPONENT

        bar_func = bar_component_value(bc)
        quote_func = quote_component_value(qc)
        trade_func = trade_component_value(tc)

        mnemonic = f"jtpo({length}{component_triple_mnemonic(bc, qc, tc)})"
        description = f"Jurik turning point oscillator {mnemonic}"

        self._line = LineIndicator(mnemonic, description, bar_func, quote_func, trade_func, self.update)
        self._primed = False
        self._length = length
        self._buffer = [0.0] * length
        self._buf_idx = 0
        self._count = 0
        self._f18 = 12.0 / (length * (length - 1) * (length + 1))
        self._midpoint = (length + 1) / 2.0

    def is_primed(self) -> bool:
        return self._primed

    def metadata(self) -> Metadata:
        return build_metadata(
            Identifier.JURIK_TURNING_POINT_OSCILLATOR,
            self._line.mnemonic,
            self._line.description,
            [OutputText(mnemonic=self._line.mnemonic, description=self._line.description)],
        )

    def update(self, sample: float) -> float:
        """Update the indicator with a new sample value."""
        if math.isnan(sample):
            return sample

        length = self._length

        # Store in circular buffer.
        self._buffer[self._buf_idx] = sample
        self._buf_idx = (self._buf_idx + 1) % length
        self._count += 1

        if self._count < length:
            return math.nan

        # Extract window in chronological order.
        window = [0.0] * length
        for i in range(length):
            window[i] = self._buffer[(self._buf_idx + i) % length]

        # Check if all values are identical.
        all_same = True
        first = window[0]
        for i in range(1, length):
            if window[i] != first:
                all_same = False
                break

        if all_same:
            if not self._primed:
                self._primed = True
            return math.nan

        # Build time positions array (1-based) and sort by price.
        # arr2[i] = original time position of the i-th element.
        indices = list(range(length))
        # Stable sort by price value.
        indices.sort(key=lambda k: window[k])

        arr2 = [0.0] * length
        for i in range(length):
            arr2[i] = float(indices[i] + 1)

        sorted_prices = [window[indices[i]] for i in range(length)]

        # Assign fractional ranks for tied prices.
        arr3 = [0.0] * length
        i = 0
        while i < length:
            j = i
            while j < length - 1 and sorted_prices[j + 1] == sorted_prices[j]:
                j += 1
            avg_rank = (i + 1 + j + 1) / 2.0
            for k in range(i, j + 1):
                arr3[k] = avg_rank
            i = j + 1

        # Compute correlation sum.
        midpoint = self._midpoint
        correlation_sum = 0.0
        for i in range(length):
            correlation_sum += (arr3[i] - midpoint) * (arr2[i] - midpoint)

        if not self._primed:
            self._primed = True

        return self._f18 * correlation_sum

    def update_bar(self, sample: Bar) -> Output:
        return self._line.update_bar(sample)

    def update_quote(self, sample: Quote) -> Output:
        return self._line.update_quote(sample)

    def update_trade(self, sample: Trade) -> Output:
        return self._line.update_trade(sample)

    def update_scalar(self, sample: Scalar) -> Output:
        return self._line.update_scalar(sample)
