"""T3 exponential moving average indicator."""

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
from .params import T3ExponentialMovingAverageLengthParams, T3ExponentialMovingAverageSmoothingFactorParams


_EPSILON = 0.00000001


class T3ExponentialMovingAverage(Indicator):
    """Computes the T3 Exponential Moving Average (T3, T3EMA).

    A six-pole non-linear Kalman filter.
    T3 = c1*ema6 + c2*ema5 + c3*ema4 + c4*ema3.
    """

    def __init__(self, length: int, smoothing_factor: float,
                 c1: float, c2: float, c3: float, c4: float,
                 length2: int, length3: int, length4: int, length5: int, length6: int,
                 first_is_average: bool,
                 mnemonic: str, description: str,
                 bar_func, quote_func, trade_func) -> None:
        self._line = LineIndicator(mnemonic, description, bar_func, quote_func, trade_func, self.update)
        self._smoothing_factor: float = smoothing_factor
        self._c1: float = c1
        self._c2: float = c2
        self._c3: float = c3
        self._c4: float = c4
        self._length: int = length
        self._length2: int = length2
        self._length3: int = length3
        self._length4: int = length4
        self._length5: int = length5
        self._length6: int = length6
        self._first_is_average: bool = first_is_average
        self._sum: float = 0.0
        self._ema1: float = 0.0
        self._ema2: float = 0.0
        self._ema3: float = 0.0
        self._ema4: float = 0.0
        self._ema5: float = 0.0
        self._ema6: float = 0.0
        self._count: int = 0
        self._primed: bool = False

    @staticmethod
    def from_length(params: T3ExponentialMovingAverageLengthParams) -> 'T3ExponentialMovingAverage':
        return _new_t3(params.length, math.nan, params.volume_factor, params.first_is_average,
                       params.bar_component, params.quote_component, params.trade_component)

    @staticmethod
    def from_smoothing_factor(params: T3ExponentialMovingAverageSmoothingFactorParams) -> 'T3ExponentialMovingAverage':
        return _new_t3(0, params.smoothing_factor, params.volume_factor, params.first_is_average,
                       params.bar_component, params.quote_component, params.trade_component)

    def is_primed(self) -> bool:
        return self._primed

    def metadata(self) -> Metadata:
        return build_metadata(
            Identifier.T3_EXPONENTIAL_MOVING_AVERAGE,
            self._line.mnemonic,
            self._line.description,
            [OutputText(self._line.mnemonic, self._line.description)],
        )

    def update(self, sample: float) -> float:
        if math.isnan(sample):
            return sample

        sf = self._smoothing_factor

        if self._primed:
            v1, v2, v3 = self._ema1, self._ema2, self._ema3
            v4, v5, v6 = self._ema4, self._ema5, self._ema6
            v1 += (sample - v1) * sf
            v2 += (v1 - v2) * sf
            v3 += (v2 - v3) * sf
            v4 += (v3 - v4) * sf
            v5 += (v4 - v5) * sf
            v6 += (v5 - v6) * sf
            self._ema1, self._ema2, self._ema3 = v1, v2, v3
            self._ema4, self._ema5, self._ema6 = v4, v5, v6
            return self._c1 * v6 + self._c2 * v5 + self._c3 * v4 + self._c4 * v3

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
            elif self._length4 >= self._count:
                self._ema1 += (sample - self._ema1) * sf
                self._ema2 += (self._ema1 - self._ema2) * sf
                self._ema3 += (self._ema2 - self._ema3) * sf
                self._sum += self._ema3
                if self._length4 == self._count:
                    self._ema4 = self._sum / self._length
                    self._sum = self._ema4
            elif self._length5 >= self._count:
                self._ema1 += (sample - self._ema1) * sf
                self._ema2 += (self._ema1 - self._ema2) * sf
                self._ema3 += (self._ema2 - self._ema3) * sf
                self._ema4 += (self._ema3 - self._ema4) * sf
                self._sum += self._ema4
                if self._length5 == self._count:
                    self._ema5 = self._sum / self._length
                    self._sum = self._ema5
            else:
                self._ema1 += (sample - self._ema1) * sf
                self._ema2 += (self._ema1 - self._ema2) * sf
                self._ema3 += (self._ema2 - self._ema3) * sf
                self._ema4 += (self._ema3 - self._ema4) * sf
                self._ema5 += (self._ema4 - self._ema5) * sf
                self._sum += self._ema5
                if self._length6 == self._count:
                    self._primed = True
                    self._ema6 = self._sum / self._length
                    return self._c1 * self._ema6 + self._c2 * self._ema5 + self._c3 * self._ema4 + self._c4 * self._ema3
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
            elif self._length4 >= self._count:
                self._ema1 += (sample - self._ema1) * sf
                self._ema2 += (self._ema1 - self._ema2) * sf
                self._ema3 += (self._ema2 - self._ema3) * sf
                self._ema4 += (self._ema3 - self._ema4) * sf
                if self._length4 == self._count:
                    self._ema5 = self._ema4
            elif self._length5 >= self._count:
                self._ema1 += (sample - self._ema1) * sf
                self._ema2 += (self._ema1 - self._ema2) * sf
                self._ema3 += (self._ema2 - self._ema3) * sf
                self._ema4 += (self._ema3 - self._ema4) * sf
                self._ema5 += (self._ema4 - self._ema5) * sf
                if self._length5 == self._count:
                    self._ema6 = self._ema5
            else:
                self._ema1 += (sample - self._ema1) * sf
                self._ema2 += (self._ema1 - self._ema2) * sf
                self._ema3 += (self._ema2 - self._ema3) * sf
                self._ema4 += (self._ema3 - self._ema4) * sf
                self._ema5 += (self._ema4 - self._ema5) * sf
                self._ema6 += (self._ema5 - self._ema6) * sf
                if self._length6 == self._count:
                    self._primed = True
                    return self._c1 * self._ema6 + self._c2 * self._ema5 + self._c3 * self._ema4 + self._c4 * self._ema3

        return math.nan

    def update_scalar(self, sample: Scalar) -> Output:
        return self._line.update_scalar(sample)

    def update_bar(self, sample: Bar) -> Output:
        return self._line.update_bar(sample)

    def update_quote(self, sample: Quote) -> Output:
        return self._line.update_quote(sample)

    def update_trade(self, sample: Trade) -> Output:
        return self._line.update_trade(sample)


def _new_t3(length: int, alpha: float, v: float, first_is_average: bool,
            bc_param, qc_param, tc_param) -> T3ExponentialMovingAverage:
    invalid = "invalid t3 exponential moving average parameters"

    if v < 0.0 or v > 1.0:
        raise ValueError(f"{invalid}: volume factor should be in range [0, 1]")

    bc = bc_param if bc_param is not None else DEFAULT_BAR_COMPONENT
    qc = qc_param if qc_param is not None else DEFAULT_QUOTE_COMPONENT
    tc = tc_param if tc_param is not None else DEFAULT_TRADE_COMPONENT

    if math.isnan(alpha):
        if length < 2:
            raise ValueError(f"{invalid}: length should be greater than 1")
        alpha = 2.0 / (1 + length)
        mnemonic = f"t3({length}, {v:.8f}{component_triple_mnemonic(bc, qc, tc)})"
    else:
        if alpha < 0.0 or alpha > 1.0:
            raise ValueError(f"{invalid}: smoothing factor should be in range [0, 1]")
        if alpha < _EPSILON:
            alpha = _EPSILON
        length = int(round(2.0 / alpha)) - 1
        mnemonic = f"t3({length}, {alpha:.8f}, {v:.8f}{component_triple_mnemonic(bc, qc, tc)})"

    bar_func = bar_component_value(bc)
    quote_func = quote_component_value(qc)
    trade_func = trade_component_value(tc)

    vv = v * v
    c1 = -vv * v
    c2 = 3 * (vv - c1)
    c3 = -6 * vv - 3 * (v - c1)
    c4 = 1 + 3 * v - c1 + 3 * vv

    length2 = 2 * length - 1
    length3 = 3 * length - 2
    length4 = 4 * length - 3
    length5 = 5 * length - 4
    length6 = 6 * length - 5
    description = f"T3 exponential moving average {mnemonic}"

    return T3ExponentialMovingAverage(length, alpha, c1, c2, c3, c4,
                                      length2, length3, length4, length5, length6,
                                      first_is_average,
                                      mnemonic, description,
                                      bar_func, quote_func, trade_func)
