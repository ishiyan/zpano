"""Zero-lag error-correcting exponential moving average indicator."""

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
from .params import ZeroLagErrorCorrectingExponentialMovingAverageParams


_EPSILON = 0.00000001


class ZeroLagErrorCorrectingExponentialMovingAverage(Indicator):
    """Ehler's adaptive zero-lag error-correcting exponential moving average (ZECEMA).

    Iterates gain in [-gainLimit, gainLimit] by gainStep to minimize |sample - ec|.
    Primes on the third sample.
    """

    def __init__(self, alpha: float, gain_limit: float, gain_step: float,
                 length: int,
                 mnemonic: str, description: str,
                 bar_func, quote_func, trade_func) -> None:
        self._line = LineIndicator(mnemonic, description, bar_func, quote_func, trade_func, self.update)
        self._alpha: float = alpha
        self._one_min_alpha: float = 1.0 - alpha
        self._gain_limit: float = gain_limit
        self._gain_step: float = gain_step
        self._length: int = length
        self._count: int = 0
        self._value: float = math.nan
        self._ema_value: float = math.nan
        self._primed: bool = False

    @staticmethod
    def create(params: ZeroLagErrorCorrectingExponentialMovingAverageParams) -> 'ZeroLagErrorCorrectingExponentialMovingAverage':
        return _new_zecema(params)

    def is_primed(self) -> bool:
        return self._primed

    def metadata(self) -> Metadata:
        return build_metadata(
            Identifier.ZERO_LAG_ERROR_CORRECTING_EXPONENTIAL_MOVING_AVERAGE,
            self._line.mnemonic,
            self._line.description,
            [OutputText(self._line.mnemonic, self._line.description)],
        )

    def update(self, sample: float) -> float:
        if math.isnan(sample):
            return sample

        if self._primed:
            self._value = self._calculate(sample)
            return self._value

        self._count += 1

        if self._count == 1:
            self._ema_value = sample
            return math.nan

        if self._count == 2:
            self._ema_value = self._calculate_ema(sample)
            self._value = self._ema_value
            return math.nan

        # count == 3: prime the indicator.
        self._value = self._calculate(sample)
        self._primed = True
        return self._value

    def _calculate_ema(self, sample: float) -> float:
        return self._alpha * sample + self._one_min_alpha * self._ema_value

    def _calculate(self, sample: float) -> float:
        self._ema_value = self._calculate_ema(sample)

        least_error = float('inf')
        best_ec = 0.0

        gain = -self._gain_limit
        while gain <= self._gain_limit:
            ec = self._alpha * (self._ema_value + gain * (sample - self._value)) \
                + self._one_min_alpha * self._value
            err = abs(sample - ec)
            if least_error > err:
                least_error = err
                best_ec = ec
            gain += self._gain_step

        return best_ec

    def update_scalar(self, sample: Scalar) -> Output:
        return self._line.update_scalar(sample)

    def update_bar(self, sample: Bar) -> Output:
        return self._line.update_bar(sample)

    def update_quote(self, sample: Quote) -> Output:
        return self._line.update_quote(sample)

    def update_trade(self, sample: Trade) -> Output:
        return self._line.update_trade(sample)


def _new_zecema(p: ZeroLagErrorCorrectingExponentialMovingAverageParams) -> ZeroLagErrorCorrectingExponentialMovingAverage:
    invalid = "invalid zero-lag error-correcting exponential moving average parameters"

    sf = p.smoothing_factor
    if sf <= 0 or sf > 1:
        raise ValueError(f"{invalid}: smoothing factor should be in (0, 1]")

    gl = p.gain_limit
    if gl <= 0:
        raise ValueError(f"{invalid}: gain limit should be positive")

    gs = p.gain_step
    if gs <= 0:
        raise ValueError(f"{invalid}: gain step should be positive")

    bc = p.bar_component if p.bar_component is not None else DEFAULT_BAR_COMPONENT
    qc = p.quote_component if p.quote_component is not None else DEFAULT_QUOTE_COMPONENT
    tc = p.trade_component if p.trade_component is not None else DEFAULT_TRADE_COMPONENT

    bar_func = bar_component_value(bc)
    quote_func = quote_component_value(qc)
    trade_func = trade_component_value(tc)

    if sf < _EPSILON:
        length = 2**63
    else:
        length = int(round(2.0 / sf)) - 1

    mnemonic = f"zecema({sf:.4g}, {gl:.4g}, {gs:.4g}{component_triple_mnemonic(bc, qc, tc)})"
    description = f"Zero-lag Error-Correcting Exponential Moving Average {mnemonic}"

    return ZeroLagErrorCorrectingExponentialMovingAverage(sf, gl, gs, length,
                                                          mnemonic, description,
                                                          bar_func, quote_func, trade_func)
