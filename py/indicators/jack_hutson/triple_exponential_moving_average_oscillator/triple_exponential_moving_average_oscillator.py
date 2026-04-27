"""Triple Exponential Moving Average Oscillator (TRIX) indicator."""

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
from ...common.exponential_moving_average.exponential_moving_average import ExponentialMovingAverage
from ...common.exponential_moving_average.params import ExponentialMovingAverageLengthParams
from .params import TripleExponentialMovingAverageOscillatorParams


class TripleExponentialMovingAverageOscillator(Indicator):
    """Jack Hutson's Triple Exponential Moving Average Oscillator (TRIX).

    TRIX = ((EMA3[i] - EMA3[i-1]) / EMA3[i-1]) * 100

    Three chained EMAs (all same length, SMA-seeded), then 1-period ROC.
    """

    def __init__(self, params: TripleExponentialMovingAverageOscillatorParams) -> None:
        length = params.length
        if length < 1:
            raise ValueError(
                "invalid triple exponential moving average oscillator parameters: "
                "length should be positive")

        bc = params.bar_component if params.bar_component is not None else DEFAULT_BAR_COMPONENT
        qc = params.quote_component if params.quote_component is not None else DEFAULT_QUOTE_COMPONENT
        tc = params.trade_component if params.trade_component is not None else DEFAULT_TRADE_COMPONENT

        bar_func = bar_component_value(bc)
        quote_func = quote_component_value(qc)
        trade_func = trade_component_value(tc)

        ema_params = ExponentialMovingAverageLengthParams(
            length=length, first_is_average=True)

        self._ema1 = ExponentialMovingAverage.from_length(ema_params)
        self._ema2 = ExponentialMovingAverage.from_length(ema_params)
        self._ema3 = ExponentialMovingAverage.from_length(ema_params)

        mnemonic = f"trix({length}{component_triple_mnemonic(bc, qc, tc)})"
        description = f"Triple exponential moving average oscillator {mnemonic}"

        self._line = LineIndicator(mnemonic, description, bar_func, quote_func, trade_func, self.update)

        self._previous_ema3: float = math.nan
        self._has_previous_ema: bool = False
        self._primed: bool = False

    def is_primed(self) -> bool:
        return self._primed

    def metadata(self) -> Metadata:
        return build_metadata(
            Identifier.TRIPLE_EXPONENTIAL_MOVING_AVERAGE_OSCILLATOR,
            self._line.mnemonic,
            self._line.description,
            [OutputText(self._line.mnemonic, self._line.description)],
        )

    def update(self, sample: float) -> float:
        if math.isnan(sample):
            return sample

        v1 = self._ema1.update(sample)
        if math.isnan(v1):
            return math.nan

        v2 = self._ema2.update(v1)
        if math.isnan(v2):
            return math.nan

        v3 = self._ema3.update(v2)
        if math.isnan(v3):
            return math.nan

        if not self._has_previous_ema:
            self._previous_ema3 = v3
            self._has_previous_ema = True
            return math.nan

        result = ((v3 - self._previous_ema3) / self._previous_ema3) * 100.0
        self._previous_ema3 = v3

        if not self._primed:
            self._primed = True

        return result

    def update_scalar(self, sample: Scalar) -> Output:
        return self._line.update_scalar(sample)

    def update_bar(self, sample: Bar) -> Output:
        return self._line.update_bar(sample)

    def update_quote(self, sample: Quote) -> Output:
        return self._line.update_quote(sample)

    def update_trade(self, sample: Trade) -> Output:
        return self._line.update_trade(sample)
