"""Absolute Price Oscillator indicator."""

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
from ..simple_moving_average.simple_moving_average import SimpleMovingAverage
from ..simple_moving_average.params import SimpleMovingAverageParams
from ..exponential_moving_average.exponential_moving_average import ExponentialMovingAverage
from ..exponential_moving_average.params import ExponentialMovingAverageLengthParams
from .params import AbsolutePriceOscillatorParams, MovingAverageType


class AbsolutePriceOscillator(Indicator):
    """Computes the Absolute Price Oscillator (APO).

    APO = fast_ma - slow_ma.
    """

    def __init__(self, params: AbsolutePriceOscillatorParams) -> None:
        if params.fast_length < 2:
            raise ValueError(
                "invalid absolute price oscillator parameters: fast length should be greater than 1")
        if params.slow_length < 2:
            raise ValueError(
                "invalid absolute price oscillator parameters: slow length should be greater than 1")

        bc = params.bar_component if params.bar_component is not None else DEFAULT_BAR_COMPONENT
        qc = params.quote_component if params.quote_component is not None else DEFAULT_QUOTE_COMPONENT
        tc = params.trade_component if params.trade_component is not None else DEFAULT_TRADE_COMPONENT

        bar_func = bar_component_value(bc)
        quote_func = quote_component_value(qc)
        trade_func = trade_component_value(tc)

        if params.moving_average_type == MovingAverageType.EMA:
            ma_label = "EMA"
            self._fast_ma = ExponentialMovingAverage.from_length(
                ExponentialMovingAverageLengthParams(
                    length=params.fast_length,
                    first_is_average=params.first_is_average,
                ))
            self._slow_ma = ExponentialMovingAverage.from_length(
                ExponentialMovingAverageLengthParams(
                    length=params.slow_length,
                    first_is_average=params.first_is_average,
                ))
        else:
            ma_label = "SMA"
            self._fast_ma = SimpleMovingAverage(
                SimpleMovingAverageParams(length=params.fast_length))
            self._slow_ma = SimpleMovingAverage(
                SimpleMovingAverageParams(length=params.slow_length))

        mnemonic = f"apo({ma_label}{params.fast_length}/{ma_label}{params.slow_length}" \
                   f"{component_triple_mnemonic(bc, qc, tc)})"
        description = f"Absolute Price Oscillator {mnemonic}"

        self._line = LineIndicator(mnemonic, description, bar_func, quote_func, trade_func, self.update)

        self._value: float = math.nan
        self._primed: bool = False

    def is_primed(self) -> bool:
        return self._primed

    def metadata(self) -> Metadata:
        return build_metadata(
            Identifier.ABSOLUTE_PRICE_OSCILLATOR,
            self._line.mnemonic,
            self._line.description,
            [OutputText(self._line.mnemonic, self._line.description)],
        )

    def update(self, sample: float) -> float:
        if math.isnan(sample):
            return sample

        slow = self._slow_ma.update(sample)
        fast = self._fast_ma.update(sample)
        self._primed = self._slow_ma.is_primed() and self._fast_ma.is_primed()

        if math.isnan(fast) or math.isnan(slow):
            self._value = math.nan
            return self._value

        self._value = fast - slow
        return self._value

    def update_scalar(self, sample: Scalar) -> Output:
        return self._line.update_scalar(sample)

    def update_bar(self, sample: Bar) -> Output:
        return self._line.update_bar(sample)

    def update_quote(self, sample: Quote) -> Output:
        return self._line.update_quote(sample)

    def update_trade(self, sample: Trade) -> Output:
        return self._line.update_trade(sample)
