import math

from .params import RelativeStrengthIndexParams
from ...core.line_indicator import LineIndicator
from ...core.identifier import Identifier
from ...core.metadata import Metadata
from ...core.build_metadata import build_metadata, OutputText
from ...core.component_triple_mnemonic import component_triple_mnemonic
from ....entities.bar_component import BarComponent, DEFAULT_BAR_COMPONENT, bar_component_value
from ....entities.quote_component import QuoteComponent, DEFAULT_QUOTE_COMPONENT, quote_component_value
from ....entities.trade_component import TradeComponent, DEFAULT_TRADE_COMPONENT, trade_component_value


class RelativeStrengthIndex(LineIndicator):
    """Welles Wilder's Relative Strength Index (RSI).

    RSI measures the magnitude of recent price changes to evaluate overbought
    or oversold conditions. It oscillates between 0 and 100.
    """

    def __init__(self, p: RelativeStrengthIndexParams) -> None:
        length = p.length
        if length < 2:
            raise ValueError("invalid relative strength index parameters: length should be greater than 1")

        bc = p.bar_component if p.bar_component is not None else DEFAULT_BAR_COMPONENT
        qc = p.quote_component if p.quote_component is not None else DEFAULT_QUOTE_COMPONENT
        tc = p.trade_component if p.trade_component is not None else DEFAULT_TRADE_COMPONENT

        bar_func = bar_component_value(bc)
        quote_func = quote_component_value(qc)
        trade_func = trade_component_value(tc)

        mnemonic = f"rsi({length}{component_triple_mnemonic(bc, qc, tc)})"
        description = f"Relative Strength Index {mnemonic}"

        super().__init__(mnemonic, description, bar_func, quote_func, trade_func, self.update)

        self._length = length
        self._count = -1
        self._previous_sample = 0.0
        self._previous_gain = 0.0
        self._previous_loss = 0.0
        self._value = math.nan
        self._primed = False

    def is_primed(self) -> bool:
        return self._primed

    def metadata(self) -> Metadata:
        return build_metadata(
            Identifier.RELATIVE_STRENGTH_INDEX,
            self.mnemonic,
            self.description,
            [OutputText(self.mnemonic, self.description)],
        )

    def update(self, sample: float) -> float:
        epsilon = 1e-8

        if math.isnan(sample):
            return sample

        self._count += 1

        if self._count == 0:
            self._previous_sample = sample
            return self._value

        temp = sample - self._previous_sample
        self._previous_sample = sample

        if not self._primed:
            # Accumulation phase: count 1..length-1.
            if temp < 0:
                self._previous_loss -= temp
            else:
                self._previous_gain += temp

            if self._count < self._length:
                return self._value

            # Priming: count == length.
            self._previous_gain /= self._length
            self._previous_loss /= self._length
            self._primed = True
        else:
            # Wilder's smoothing.
            self._previous_gain *= (self._length - 1)
            self._previous_loss *= (self._length - 1)

            if temp < 0:
                self._previous_loss -= temp
            else:
                self._previous_gain += temp

            self._previous_gain /= self._length
            self._previous_loss /= self._length

        total = self._previous_gain + self._previous_loss
        if total > epsilon:
            self._value = 100 * self._previous_gain / total
        else:
            self._value = 0

        return self._value
