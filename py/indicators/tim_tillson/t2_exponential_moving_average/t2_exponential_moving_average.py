"""T2 exponential moving average indicator."""

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
from .params import T2ExponentialMovingAverageLengthParams, T2ExponentialMovingAverageSmoothingFactorParams


_EPSILON = 0.00000001


class T2ExponentialMovingAverage(Indicator):
    """Computes the T2 Exponential Moving Average (T2, T2EMA).

    A four-pole non-linear Kalman filter.
    T2 = c1*ema4 + c2*ema3 + c3*ema2.
    """

    def __init__(self, length: int, smoothing_factor: float,
                 c1: float, c2: float, c3: float,
                 length2: int, length3: int, length4: int,
                 first_is_average: bool,
                 mnemonic: str, description: str,
                 bar_func, quote_func, trade_func) -> None:
        self._line = LineIndicator(mnemonic, description, bar_func, quote_func, trade_func, self.update)
        self._smoothing_factor: float = smoothing_factor
        self._c1: float = c1
        self._c2: float = c2
        self._c3: float = c3
        self._length: int = length
        self._length2: int = length2
        self._length3: int = length3
        self._length4: int = length4
        self._first_is_average: bool = first_is_average
        self._sum: float = 0.0
        self._ema1: float = 0.0
        self._ema2: float = 0.0
        self._ema3: float = 0.0
        self._ema4: float = 0.0
        self._count: int = 0
        self._primed: bool = False

    @staticmethod
    def from_length(params: T2ExponentialMovingAverageLengthParams) -> 'T2ExponentialMovingAverage':
        return _new_t2(params.length, math.nan, params.volume_factor, params.first_is_average,
                       params.bar_component, params.quote_component, params.trade_component)

    @staticmethod
    def from_smoothing_factor(params: T2ExponentialMovingAverageSmoothingFactorParams) -> 'T2ExponentialMovingAverage':
        return _new_t2(0, params.smoothing_factor, params.volume_factor, params.first_is_average,
                       params.bar_component, params.quote_component, params.trade_component)

    def is_primed(self) -> bool:
        return self._primed

    def metadata(self) -> Metadata:
        return build_metadata(
            Identifier.T2_EXPONENTIAL_MOVING_AVERAGE,
            self._line.mnemonic,
            self._line.description,
            [OutputText(self._line.mnemonic, self._line.description)],
        )

    def update(self, sample: float) -> float:
        if math.isnan(sample):
            return sample

        sf = self._smoothing_factor

        if self._primed:
            v1 = self._ema1
            v2 = self._ema2
            v3 = self._ema3
            v4 = self._ema4
            v1 += (sample - v1) * sf
            v2 += (v1 - v2) * sf
            v3 += (v2 - v3) * sf
            v4 += (v3 - v4) * sf
            self._ema1 = v1
            self._ema2 = v2
            self._ema3 = v3
            self._ema4 = v4
            return self._c1 * v4 + self._c2 * v3 + self._c3 * v2

        self._count += 1
        if self._first_is_average:
            if self._count == 1:
                self._sum = sample
            elif self._length >= self._count:
                self._sum += sample
                if self._length == self._count:
                    self._ema1 = self._sum / self._length
                    self._sum = self._ema1
            elif self._length2 >= self._count:
                self._ema1 += (sample - self._ema1) * sf
                self._sum += self._ema1
                if self._length2 == self._count:
                    self._ema2 = self._sum / self._length
                    self._sum = self._ema2
            elif self._length3 >= self._count:
                self._ema1 += (sample - self._ema1) * sf
                self._ema2 += (self._ema1 - self._ema2) * sf
                self._sum += self._ema2
                if self._length3 == self._count:
                    self._ema3 = self._sum / self._length
                    self._sum = self._ema3
            else:
                self._ema1 += (sample - self._ema1) * sf
                self._ema2 += (self._ema1 - self._ema2) * sf
                self._ema3 += (self._ema2 - self._ema3) * sf
                self._sum += self._ema3
                if self._length4 == self._count:
                    self._primed = True
                    self._ema4 = self._sum / self._length
                    return self._c1 * self._ema4 + self._c2 * self._ema3 + self._c3 * self._ema2
        else:  # Metastock
            if self._count == 1:
                self._ema1 = sample
            elif self._length >= self._count:
                self._ema1 += (sample - self._ema1) * sf
                if self._length == self._count:
                    self._ema2 = self._ema1
            elif self._length2 >= self._count:
                self._ema1 += (sample - self._ema1) * sf
                self._ema2 += (self._ema1 - self._ema2) * sf
                if self._length2 == self._count:
                    self._ema3 = self._ema2
            elif self._length3 >= self._count:
                self._ema1 += (sample - self._ema1) * sf
                self._ema2 += (self._ema1 - self._ema2) * sf
                self._ema3 += (self._ema2 - self._ema3) * sf
                if self._length3 == self._count:
                    self._ema4 = self._ema3
            else:
                self._ema1 += (sample - self._ema1) * sf
                self._ema2 += (self._ema1 - self._ema2) * sf
                self._ema3 += (self._ema2 - self._ema3) * sf
                self._ema4 += (self._ema3 - self._ema4) * sf
                if self._length4 == self._count:
                    self._primed = True
                    return self._c1 * self._ema4 + self._c2 * self._ema3 + self._c3 * self._ema2

        return math.nan

    def update_scalar(self, sample: Scalar) -> Output:
        return self._line.update_scalar(sample)

    def update_bar(self, sample: Bar) -> Output:
        return self._line.update_bar(sample)

    def update_quote(self, sample: Quote) -> Output:
        return self._line.update_quote(sample)

    def update_trade(self, sample: Trade) -> Output:
        return self._line.update_trade(sample)


def _new_t2(length: int, alpha: float, v: float, first_is_average: bool,
            bc_param, qc_param, tc_param) -> T2ExponentialMovingAverage:
    invalid = "invalid t2 exponential moving average parameters"

    if v < 0.0 or v > 1.0:
        raise ValueError(f"{invalid}: volume factor should be in range [0, 1]")

    bc = bc_param if bc_param is not None else DEFAULT_BAR_COMPONENT
    qc = qc_param if qc_param is not None else DEFAULT_QUOTE_COMPONENT
    tc = tc_param if tc_param is not None else DEFAULT_TRADE_COMPONENT

    if math.isnan(alpha):
        if length < 2:
            raise ValueError(f"{invalid}: length should be greater than 1")
        alpha = 2.0 / (1 + length)
        mnemonic = f"t2({length}, {v:.8f}{component_triple_mnemonic(bc, qc, tc)})"
    else:
        if alpha < 0.0 or alpha > 1.0:
            raise ValueError(f"{invalid}: smoothing factor should be in range [0, 1]")
        if alpha < _EPSILON:
            alpha = _EPSILON
        length = int(round(2.0 / alpha)) - 1
        mnemonic = f"t2({length}, {alpha:.8f}, {v:.8f}{component_triple_mnemonic(bc, qc, tc)})"

    bar_func = bar_component_value(bc)
    quote_func = quote_component_value(qc)
    trade_func = trade_component_value(tc)

    v1 = v + 1
    c1 = v * v
    c2 = -2.0 * v * v1
    c3 = v1 * v1

    length2 = 2 * length - 1
    length3 = 3 * length - 2
    length4 = 4 * length - 3
    description = f"T2 exponential moving average {mnemonic}"

    return T2ExponentialMovingAverage(length, alpha, c1, c2, c3,
                                      length2, length3, length4, first_is_average,
                                      mnemonic, description,
                                      bar_func, quote_func, trade_func)
