"""Arnaud Legoux moving average indicator."""

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
from .params import ArnaudLegouxMovingAverageParams


class ArnaudLegouxMovingAverage(Indicator):
    """Computes the Arnaud Legoux Moving Average (ALMA).

    ALMA is a Gaussian-weighted moving average that reduces lag while maintaining
    smoothness. It applies a Gaussian bell curve as its kernel, shifted toward
    recent bars via an adjustable offset parameter.

    The indicator is not primed during the first (window - 1) updates.
    """

    def __init__(self, params: ArnaudLegouxMovingAverageParams) -> None:
        window = params.window
        if window < 1:
            raise ValueError(
                "invalid Arnaud Legoux moving average parameters: window should be greater than 0")

        sigma = params.sigma
        if sigma <= 0.0:
            raise ValueError(
                "invalid Arnaud Legoux moving average parameters: sigma should be greater than 0")

        offset = params.offset
        if offset < 0.0 or offset > 1.0:
            raise ValueError(
                "invalid Arnaud Legoux moving average parameters: offset should be between 0 and 1")

        bc = params.bar_component if params.bar_component is not None else DEFAULT_BAR_COMPONENT
        qc = params.quote_component if params.quote_component is not None else DEFAULT_QUOTE_COMPONENT
        tc = params.trade_component if params.trade_component is not None else DEFAULT_TRADE_COMPONENT

        bar_func = bar_component_value(bc)
        quote_func = quote_component_value(qc)
        trade_func = trade_component_value(tc)

        mnemonic = f"alma({window}, {sigma}, {offset}{component_triple_mnemonic(bc, qc, tc)})"
        description = f"Arnaud Legoux moving average {mnemonic}"

        self._line = LineIndicator(mnemonic, description, bar_func, quote_func, trade_func, self.update)

        # Precompute Gaussian weights.
        m = offset * (window - 1)
        s = window / sigma
        weights = [math.exp(-((i - m) ** 2) / (2.0 * s * s)) for i in range(window)]
        norm = sum(weights)
        self._weights: list[float] = [w / norm for w in weights]

        self._window_length: int = window
        self._buffer: list[float] = [0.0] * window
        self._buffer_count: int = 0
        self._buffer_index: int = 0
        self._primed: bool = False

    def is_primed(self) -> bool:
        return self._primed

    def metadata(self) -> Metadata:
        return build_metadata(
            Identifier.ARNAUD_LEGOUX_MOVING_AVERAGE,
            self._line.mnemonic,
            self._line.description,
            [OutputText(self._line.mnemonic, self._line.description)],
        )

    def update(self, sample: float) -> float:
        if math.isnan(sample):
            return sample

        window = self._window_length

        if window == 1:
            self._primed = True
            return sample

        # Fill the circular buffer.
        self._buffer[self._buffer_index] = sample
        self._buffer_index = (self._buffer_index + 1) % window

        if not self._primed:
            self._buffer_count += 1
            if self._buffer_count < window:
                return math.nan
            self._primed = True

        # Compute weighted sum.
        # Weight[0] applies to oldest sample, weight[N-1] to newest.
        # The oldest sample is at self._buffer_index (circular buffer).
        result = 0.0
        index = self._buffer_index
        for i in range(window):
            result += self._weights[i] * self._buffer[index]
            index = (index + 1) % window

        return result

    def update_scalar(self, sample: Scalar) -> Output:
        return self._line.update_scalar(sample)

    def update_bar(self, sample: Bar) -> Output:
        return self._line.update_bar(sample)

    def update_quote(self, sample: Quote) -> Output:
        return self._line.update_quote(sample)

    def update_trade(self, sample: Trade) -> Output:
        return self._line.update_trade(sample)
