"""Zero-lag exponential moving average indicator."""

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
from .params import ZeroLagExponentialMovingAverageParams


_EPSILON = 0.00000001


class ZeroLagExponentialMovingAverage(Indicator):
    """Ehler's Zero-lag Exponential Moving Average (ZEMA).

    ZEMA = alpha*(Price + gainFactor*(Price - Price[momentumLength ago])) + (1 - alpha)*ZEMA[previous]
    """

    def __init__(self, alpha: float, gain_factor: float,
                 momentum_length: int, length: int,
                 mnemonic: str, description: str,
                 bar_func, quote_func, trade_func) -> None:
        self._line = LineIndicator(mnemonic, description, bar_func, quote_func, trade_func, self.update)
        self._alpha: float = alpha
        self._one_min_alpha: float = 1.0 - alpha
        self._gain_factor: float = gain_factor
        self._momentum_length: int = momentum_length
        self._momentum_window: list[float] = [0.0] * (momentum_length + 1)
        self._length: int = length
        self._count: int = 0
        self._value: float = math.nan
        self._primed: bool = False

    @staticmethod
    def create(params: ZeroLagExponentialMovingAverageParams) -> 'ZeroLagExponentialMovingAverage':
        return _new_zema(params)

    def is_primed(self) -> bool:
        return self._primed

    def metadata(self) -> Metadata:
        return build_metadata(
            Identifier.ZERO_LAG_EXPONENTIAL_MOVING_AVERAGE,
            self._line.mnemonic,
            self._line.description,
            [OutputText(self._line.mnemonic, self._line.description)],
        )

    def update(self, sample: float) -> float:
        if math.isnan(sample):
            return sample

        if self._primed:
            # Shift momentum window left by 1.
            self._momentum_window[:-1] = self._momentum_window[1:]
            self._momentum_window[self._momentum_length] = sample
            self._value = self._calculate(sample)
            return self._value

        self._momentum_window[self._count] = sample
        self._count += 1

        if self._count <= self._momentum_length:
            self._value = sample
            return math.nan

        # count == momentum_length + 1: prime the indicator.
        self._value = self._calculate(sample)
        self._primed = True
        return self._value

    def _calculate(self, sample: float) -> float:
        momentum = sample - self._momentum_window[0]
        return self._alpha * (sample + self._gain_factor * momentum) + self._one_min_alpha * self._value

    def update_scalar(self, sample: Scalar) -> Output:
        return self._line.update_scalar(sample)

    def update_bar(self, sample: Bar) -> Output:
        return self._line.update_bar(sample)

    def update_quote(self, sample: Quote) -> Output:
        return self._line.update_quote(sample)

    def update_trade(self, sample: Trade) -> Output:
        return self._line.update_trade(sample)


def _new_zema(p: ZeroLagExponentialMovingAverageParams) -> ZeroLagExponentialMovingAverage:
    invalid = "invalid zero-lag exponential moving average parameters"

    sf = p.smoothing_factor
    if sf <= 0 or sf > 1:
        raise ValueError(f"{invalid}: smoothing factor should be in (0, 1]")

    ml = p.velocity_momentum_length
    if ml < 1:
        raise ValueError(f"{invalid}: velocity momentum length should be positive")

    bc = p.bar_component if p.bar_component is not None else DEFAULT_BAR_COMPONENT
    qc = p.quote_component if p.quote_component is not None else DEFAULT_QUOTE_COMPONENT
    tc = p.trade_component if p.trade_component is not None else DEFAULT_TRADE_COMPONENT

    bar_func = bar_component_value(bc)
    quote_func = quote_component_value(qc)
    trade_func = trade_component_value(tc)

    if sf < _EPSILON:
        length = 2**63  # large sentinel
    else:
        length = int(round(2.0 / sf)) - 1

    gf = p.velocity_gain_factor
    mnemonic = f"zema({sf:.4g}, {gf:.4g}, {ml}{component_triple_mnemonic(bc, qc, tc)})"
    description = f"Zero-lag Exponential Moving Average {mnemonic}"

    return ZeroLagExponentialMovingAverage(sf, gf, ml, length,
                                           mnemonic, description,
                                           bar_func, quote_func, trade_func)
