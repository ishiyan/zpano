"""Chande Momentum Oscillator indicator."""

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
from .params import ChandeMomentumOscillatorParams


class ChandeMomentumOscillator(Indicator):
    """Computes the Chande Momentum Oscillator (CMO).

    CMOi = 100 * (SUi - SDi) / (SUi + SDi),

    where SUi (sum up) is the sum of gains and SDi (sum down)
    is the sum of losses over the chosen length.

    The indicator is not primed during the first l updates.
    """

    def __init__(self, params: ChandeMomentumOscillatorParams) -> None:
        length = params.length
        if length < 1:
            raise ValueError(
                "invalid Chande momentum oscillator parameters: length should be positive")

        bc = params.bar_component if params.bar_component is not None else DEFAULT_BAR_COMPONENT
        qc = params.quote_component if params.quote_component is not None else DEFAULT_QUOTE_COMPONENT
        tc = params.trade_component if params.trade_component is not None else DEFAULT_TRADE_COMPONENT

        bar_func = bar_component_value(bc)
        quote_func = quote_component_value(qc)
        trade_func = trade_component_value(tc)

        mnemonic = f"cmo({length}{component_triple_mnemonic(bc, qc, tc)})"
        description = f"Chande Momentum Oscillator {mnemonic}"

        self._line = LineIndicator(mnemonic, description, bar_func, quote_func, trade_func, self.update)

        self._length: int = length
        self._count: int = 0
        self._ring_buffer: list[float] = [0.0] * length
        self._ring_head: int = 0
        self._previous_sample: float = 0.0
        self._gain_sum: float = 0.0
        self._loss_sum: float = 0.0
        self._primed: bool = False

    def is_primed(self) -> bool:
        return self._primed

    def metadata(self) -> Metadata:
        return build_metadata(
            Identifier.CHANDE_MOMENTUM_OSCILLATOR,
            self._line.mnemonic,
            self._line.description,
            [OutputText(self._line.mnemonic, self._line.description)],
        )

    def update(self, sample: float) -> float:
        if math.isnan(sample):
            return sample

        epsilon = 1e-12

        self._count += 1
        if self._count == 1:
            self._previous_sample = sample
            return math.nan

        # New delta
        delta = sample - self._previous_sample
        self._previous_sample = sample

        if not self._primed:
            # Fill until we have self._length deltas
            self._ring_buffer[self._ring_head] = delta
            self._ring_head = (self._ring_head + 1) % self._length

            if delta > 0:
                self._gain_sum += delta
            elif delta < 0:
                self._loss_sum += -delta

            if self._count <= self._length:
                return math.nan

            # Now we have exactly self._length deltas in the buffer
            self._primed = True
        else:
            # Remove oldest delta and add the new one
            old = self._ring_buffer[self._ring_head]
            if old > 0:
                self._gain_sum -= old
            elif old < 0:
                self._loss_sum -= -old

            self._ring_buffer[self._ring_head] = delta
            self._ring_head = (self._ring_head + 1) % self._length

            if delta > 0:
                self._gain_sum += delta
            elif delta < 0:
                self._loss_sum += -delta

            # Clamp to avoid tiny negative sums from FP noise
            if self._gain_sum < 0:
                self._gain_sum = 0.0
            if self._loss_sum < 0:
                self._loss_sum = 0.0

        den = self._gain_sum + self._loss_sum
        if abs(den) < epsilon:
            return 0.0

        return 100.0 * (self._gain_sum - self._loss_sum) / den

    def update_scalar(self, sample: Scalar) -> Output:
        return self._line.update_scalar(sample)

    def update_bar(self, sample: Bar) -> Output:
        return self._line.update_bar(sample)

    def update_quote(self, sample: Quote) -> Output:
        return self._line.update_quote(sample)

    def update_trade(self, sample: Trade) -> Output:
        return self._line.update_trade(sample)
