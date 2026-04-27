"""Exponential moving average indicator."""

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
from .params import ExponentialMovingAverageLengthParams, ExponentialMovingAverageSmoothingFactorParams


_EPSILON = 0.00000001


class ExponentialMovingAverage(Indicator):
    """Computes the exponential, or exponentially weighted, moving average (EMA).

    EMAi = EMAi-1 + alpha * (Pi - EMAi-1), 0 < alpha <= 1.

    The indicator is not primed during the first l-1 updates.
    """

    def __init__(self, length: int, smoothing_factor: float,
                 first_is_average: bool, mnemonic: str, description: str,
                 bar_func, quote_func, trade_func) -> None:
        self._line = LineIndicator(mnemonic, description, bar_func, quote_func, trade_func, self.update)
        self._value: float = 0.0
        self._sum: float = 0.0
        self._smoothing_factor: float = smoothing_factor
        self._length: int = length
        self._count: int = 0
        self._first_is_average: bool = first_is_average
        self._primed: bool = False

    @staticmethod
    def from_length(params: ExponentialMovingAverageLengthParams) -> 'ExponentialMovingAverage':
        """Creates an EMA from length-based parameters."""
        return _new_ema(params.length, math.nan, params.first_is_average,
                        params.bar_component, params.quote_component, params.trade_component)

    @staticmethod
    def from_smoothing_factor(params: ExponentialMovingAverageSmoothingFactorParams) -> 'ExponentialMovingAverage':
        """Creates an EMA from smoothing-factor-based parameters."""
        return _new_ema(0, params.smoothing_factor, params.first_is_average,
                        params.bar_component, params.quote_component, params.trade_component)

    def is_primed(self) -> bool:
        return self._primed

    def metadata(self) -> Metadata:
        return build_metadata(
            Identifier.EXPONENTIAL_MOVING_AVERAGE,
            self._line.mnemonic,
            self._line.description,
            [OutputText(self._line.mnemonic, self._line.description)],
        )

    def update(self, sample: float) -> float:
        if math.isnan(sample):
            return sample

        if self._primed:
            self._value += (sample - self._value) * self._smoothing_factor
        else:
            self._count += 1
            if self._first_is_average:
                self._sum += sample
                if self._count < self._length:
                    return math.nan
                self._value = self._sum / self._length
            else:
                if self._count == 1:
                    self._value = sample
                else:
                    self._value += (sample - self._value) * self._smoothing_factor

                if self._count < self._length:
                    return math.nan

            self._primed = True

        return self._value

    def update_scalar(self, sample: Scalar) -> Output:
        return self._line.update_scalar(sample)

    def update_bar(self, sample: Bar) -> Output:
        return self._line.update_bar(sample)

    def update_quote(self, sample: Quote) -> Output:
        return self._line.update_quote(sample)

    def update_trade(self, sample: Trade) -> Output:
        return self._line.update_trade(sample)


def _new_ema(length: int, alpha: float, first_is_average: bool,
             bc_param, qc_param, tc_param) -> ExponentialMovingAverage:
    """Internal constructor for EMA."""
    invalid = "invalid exponential moving average parameters"

    bc = bc_param if bc_param is not None else DEFAULT_BAR_COMPONENT
    qc = qc_param if qc_param is not None else DEFAULT_QUOTE_COMPONENT
    tc = tc_param if tc_param is not None else DEFAULT_TRADE_COMPONENT

    if math.isnan(alpha):
        # Length-based.
        if length < 1:
            raise ValueError(f"{invalid}: length should be positive")
        alpha = 2.0 / (1 + length)
        mnemonic = f"ema({length}{component_triple_mnemonic(bc, qc, tc)})"
    else:
        # Smoothing-factor-based.
        if alpha < 0.0 or alpha > 1.0:
            raise ValueError(f"{invalid}: smoothing factor should be in range [0, 1]")
        if alpha < _EPSILON:
            alpha = _EPSILON
        length = int(round(2.0 / alpha)) - 1
        mnemonic = f"ema({length}, {alpha:.8f}{component_triple_mnemonic(bc, qc, tc)})"

    bar_func = bar_component_value(bc)
    quote_func = quote_component_value(qc)
    trade_func = trade_component_value(tc)

    description = f"Exponential moving average {mnemonic}"

    return ExponentialMovingAverage(length, alpha, first_is_average,
                                    mnemonic, description,
                                    bar_func, quote_func, trade_func)
