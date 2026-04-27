"""Double exponential moving average indicator."""

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
from .params import DoubleExponentialMovingAverageLengthParams, DoubleExponentialMovingAverageSmoothingFactorParams


_EPSILON = 0.00000001


class DoubleExponentialMovingAverage(Indicator):
    """Computes the Double Exponential Moving Average (DEMA).

    DEMA = 2*EMA1 - EMA2, where EMA2 = EMA(EMA1).
    """

    def __init__(self, length: int, smoothing_factor: float,
                 length2: int, first_is_average: bool,
                 mnemonic: str, description: str,
                 bar_func, quote_func, trade_func) -> None:
        self._line = LineIndicator(mnemonic, description, bar_func, quote_func, trade_func, self.update)
        self._smoothing_factor: float = smoothing_factor
        self._length: int = length
        self._length2: int = length2
        self._first_is_average: bool = first_is_average
        self._sum: float = 0.0
        self._ema1: float = 0.0
        self._ema2: float = 0.0
        self._count: int = 0
        self._primed: bool = False

    @staticmethod
    def from_length(params: DoubleExponentialMovingAverageLengthParams) -> 'DoubleExponentialMovingAverage':
        """Creates a DEMA from length-based parameters."""
        return _new_dema(params.length, math.nan, params.first_is_average,
                         params.bar_component, params.quote_component, params.trade_component)

    @staticmethod
    def from_smoothing_factor(params: DoubleExponentialMovingAverageSmoothingFactorParams) -> 'DoubleExponentialMovingAverage':
        """Creates a DEMA from smoothing-factor-based parameters."""
        return _new_dema(0, params.smoothing_factor, params.first_is_average,
                         params.bar_component, params.quote_component, params.trade_component)

    def is_primed(self) -> bool:
        return self._primed

    def metadata(self) -> Metadata:
        return build_metadata(
            Identifier.DOUBLE_EXPONENTIAL_MOVING_AVERAGE,
            self._line.mnemonic,
            self._line.description,
            [OutputText(self._line.mnemonic, self._line.description)],
        )

    def update(self, sample: float) -> float:
        if math.isnan(sample):
            return sample

        if self._primed:
            sf = self._smoothing_factor
            v1 = self._ema1
            v2 = self._ema2
            v1 += (sample - v1) * sf
            v2 += (v1 - v2) * sf
            self._ema1 = v1
            self._ema2 = v2
            return 2.0 * v1 - v2

        self._count += 1
        if self._first_is_average:
            if self._count == 1:
                self._sum = sample
            elif self._length >= self._count:
                self._sum += sample
                if self._length == self._count:
                    self._ema1 = self._sum / self._length
                    self._sum = self._ema1
            else:
                self._ema1 += (sample - self._ema1) * self._smoothing_factor
                self._sum += self._ema1
                if self._length2 == self._count:
                    self._primed = True
                    self._ema2 = self._sum / self._length
                    return 2.0 * self._ema1 - self._ema2
        else:  # Metastock
            if self._count == 1:
                self._ema1 = sample
            elif self._length >= self._count:
                self._ema1 += (sample - self._ema1) * self._smoothing_factor
                if self._length == self._count:
                    self._ema2 = self._ema1
            else:
                self._ema1 += (sample - self._ema1) * self._smoothing_factor
                self._ema2 += (self._ema1 - self._ema2) * self._smoothing_factor
                if self._length2 == self._count:
                    self._primed = True
                    return 2.0 * self._ema1 - self._ema2

        return math.nan

    def update_scalar(self, sample: Scalar) -> Output:
        return self._line.update_scalar(sample)

    def update_bar(self, sample: Bar) -> Output:
        return self._line.update_bar(sample)

    def update_quote(self, sample: Quote) -> Output:
        return self._line.update_quote(sample)

    def update_trade(self, sample: Trade) -> Output:
        return self._line.update_trade(sample)


def _new_dema(length: int, alpha: float, first_is_average: bool,
              bc_param, qc_param, tc_param) -> DoubleExponentialMovingAverage:
    """Internal constructor for DEMA."""
    invalid = "invalid double exponential moving average parameters"

    bc = bc_param if bc_param is not None else DEFAULT_BAR_COMPONENT
    qc = qc_param if qc_param is not None else DEFAULT_QUOTE_COMPONENT
    tc = tc_param if tc_param is not None else DEFAULT_TRADE_COMPONENT

    if math.isnan(alpha):
        if length < 1:
            raise ValueError(f"{invalid}: length should be positive")
        alpha = 2.0 / (1 + length)
        mnemonic = f"dema({length}{component_triple_mnemonic(bc, qc, tc)})"
    else:
        if alpha < 0.0 or alpha > 1.0:
            raise ValueError(f"{invalid}: smoothing factor should be in range [0, 1]")
        if alpha < _EPSILON:
            alpha = _EPSILON
        length = int(round(2.0 / alpha)) - 1
        mnemonic = f"dema({length}, {alpha:.8f}{component_triple_mnemonic(bc, qc, tc)})"

    bar_func = bar_component_value(bc)
    quote_func = quote_component_value(qc)
    trade_func = trade_component_value(tc)

    length2 = 2 * length - 1
    description = f"Double exponential moving average {mnemonic}"

    return DoubleExponentialMovingAverage(length, alpha, length2, first_is_average,
                                          mnemonic, description,
                                          bar_func, quote_func, trade_func)
