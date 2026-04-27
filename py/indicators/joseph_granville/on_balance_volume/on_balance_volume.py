"""On-Balance Volume indicator."""

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
from .params import OnBalanceVolumeParams


class OnBalanceVolume(Indicator):
    """Joseph Granville's On-Balance Volume (OBV).

    Cumulative volume indicator. If price > previous, add volume; if price < previous,
    subtract volume; if unchanged, keep the same.
    """

    def __init__(self, params: OnBalanceVolumeParams) -> None:
        bc = params.bar_component if params.bar_component is not None else DEFAULT_BAR_COMPONENT
        qc = params.quote_component if params.quote_component is not None else DEFAULT_QUOTE_COMPONENT
        tc = params.trade_component if params.trade_component is not None else DEFAULT_TRADE_COMPONENT

        self._bar_func = bar_component_value(bc)
        quote_func = quote_component_value(qc)
        trade_func = trade_component_value(tc)

        mnemonic = "obv"
        suffix = component_triple_mnemonic(bc, qc, tc)
        if suffix != "":
            mnemonic = f"obv({suffix[2:]})"  # strip leading ", "

        description = "On-Balance Volume OBV"

        self._line = LineIndicator(mnemonic, description, self._bar_func, quote_func, trade_func, self.update)

        self._previous_sample: float = 0.0
        self._value: float = math.nan
        self._primed: bool = False

    def is_primed(self) -> bool:
        return self._primed

    def metadata(self) -> Metadata:
        return build_metadata(
            Identifier.ON_BALANCE_VOLUME,
            self._line.mnemonic,
            self._line.description,
            [OutputText(self._line.mnemonic, self._line.description)],
        )

    def update(self, sample: float) -> float:
        """Update with volume = 1 (scalar/quote/trade path)."""
        return self.update_with_volume(sample, 1.0)

    def update_with_volume(self, sample: float, volume: float) -> float:
        """Update with the given sample and volume."""
        if math.isnan(sample) or math.isnan(volume):
            return math.nan

        if not self._primed:
            self._value = volume
            self._primed = True
        else:
            if sample > self._previous_sample:
                self._value += volume
            elif sample < self._previous_sample:
                self._value -= volume

        self._previous_sample = sample

        return self._value

    def update_scalar(self, sample: Scalar) -> Output:
        return self._line.update_scalar(sample)

    def update_bar(self, sample: Bar) -> Output:
        """Shadows LineIndicator.update_bar to use bar volume."""
        price = self._bar_func(sample)
        value = self.update_with_volume(price, sample.volume)
        return [Scalar(time=sample.time, value=value)]

    def update_quote(self, sample: Quote) -> Output:
        return self._line.update_quote(sample)

    def update_trade(self, sample: Trade) -> Output:
        return self._line.update_trade(sample)
