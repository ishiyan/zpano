"""Ehler's Cyber Cycle (CC) indicator."""

import math

from ...core.indicator import Indicator
from ...core.metadata import Metadata
from ...core.build_metadata import build_metadata, OutputText
from ...core.identifier import Identifier
from ...core.component_triple_mnemonic import component_triple_mnemonic
from ...core.output import Output
from ....entities.bar import Bar
from ....entities.quote import Quote
from ....entities.trade import Trade
from ....entities.scalar import Scalar
from ....entities.bar_component import BarComponent, bar_component_value
from ....entities.quote_component import QuoteComponent, DEFAULT_QUOTE_COMPONENT, quote_component_value
from ....entities.trade_component import TradeComponent, DEFAULT_TRADE_COMPONENT, trade_component_value
from .params import CyberCycleLengthParams, CyberCycleSmoothingFactorParams


_EPSILON = 0.00000001


class CyberCycle(Indicator):
    """Ehler's Cyber Cycle (CC).

    H(z) = ((1-alpha/2)^2 * (1 - 2*z^-1 + z^-2))
            / (1 - 2*(1-alpha)*z^-1 + (1-alpha)^2*z^-2)

    The indicator has two outputs: the cycle value and a signal line which
    is an exponential moving average of the cycle value.
    """

    def __init__(self, length: int, smoothing_factor: float, signal_lag: int,
                 coeff1: float, coeff2: float, coeff3: float,
                 coeff4: float, coeff5: float,
                 mnemonic: str, description: str,
                 mnemonic_signal: str, description_signal: str,
                 bar_func, quote_func, trade_func) -> None:
        self._length = length
        self._smoothing_factor = smoothing_factor
        self._signal_lag = signal_lag
        self._coeff1 = coeff1
        self._coeff2 = coeff2
        self._coeff3 = coeff3
        self._coeff4 = coeff4
        self._coeff5 = coeff5
        self._mnemonic = mnemonic
        self._description = description
        self._mnemonic_signal = mnemonic_signal
        self._description_signal = description_signal
        self._count = 0
        self._previous_sample1 = 0.0
        self._previous_sample2 = 0.0
        self._previous_sample3 = 0.0
        self._smoothed = 0.0
        self._previous_smoothed1 = 0.0
        self._previous_smoothed2 = 0.0
        self._value = math.nan
        self._previous_value1 = 0.0
        self._previous_value2 = 0.0
        self._signal = math.nan
        self._primed = False
        self._bar_func = bar_func
        self._quote_func = quote_func
        self._trade_func = trade_func

    @staticmethod
    def from_length(params: CyberCycleLengthParams) -> 'CyberCycle':
        """Creates a new CyberCycle from length parameters."""
        return _new_cyber_cycle(params.length, math.nan, params.signal_lag,
                                params.bar_component, params.quote_component, params.trade_component)

    @staticmethod
    def from_smoothing_factor(params: CyberCycleSmoothingFactorParams) -> 'CyberCycle':
        """Creates a new CyberCycle from smoothing factor parameters."""
        return _new_cyber_cycle(0, params.smoothing_factor, params.signal_lag,
                                params.bar_component, params.quote_component, params.trade_component)

    def is_primed(self) -> bool:
        return self._primed

    def metadata(self) -> Metadata:
        return build_metadata(
            Identifier.CYBER_CYCLE,
            self._mnemonic,
            self._description,
            [
                OutputText(self._mnemonic, self._description),
                OutputText(self._mnemonic_signal, self._description_signal),
            ],
        )

    def update(self, sample: float) -> float:
        if math.isnan(sample):
            return math.nan

        if self._primed:
            self._previous_smoothed2 = self._previous_smoothed1
            self._previous_smoothed1 = self._smoothed
            self._smoothed = (sample + 2 * self._previous_sample1 + 2 * self._previous_sample2 + self._previous_sample3) / 6

            self._previous_value2 = self._previous_value1
            self._previous_value1 = self._value
            self._value = self._coeff1 * (self._smoothed - 2 * self._previous_smoothed1 + self._previous_smoothed2) + \
                self._coeff2 * self._previous_value1 + self._coeff3 * self._previous_value2

            self._signal = self._coeff4 * self._value + self._coeff5 * self._signal

            self._previous_sample3 = self._previous_sample2
            self._previous_sample2 = self._previous_sample1
            self._previous_sample1 = sample

            return self._value

        self._count += 1

        if self._count == 1:
            self._previous_sample3 = sample
            return math.nan
        elif self._count == 2:
            self._previous_sample2 = sample
            return math.nan
        elif self._count == 3:
            self._signal = self._coeff4 * (sample - 2 * self._previous_sample2 + self._previous_sample3) / 4
            self._previous_sample1 = sample
            return math.nan
        elif self._count == 4:
            self._previous_smoothed2 = (sample + 2 * self._previous_sample1 + 2 * self._previous_sample2 + self._previous_sample3) / 6
            self._signal = self._coeff4 * (sample - 2 * self._previous_sample1 + self._previous_sample2) / 4 + self._coeff5 * self._signal
            self._previous_sample3 = self._previous_sample2
            self._previous_sample2 = self._previous_sample1
            self._previous_sample1 = sample
            return math.nan
        elif self._count == 5:
            self._previous_smoothed1 = (sample + 2 * self._previous_sample1 + 2 * self._previous_sample2 + self._previous_sample3) / 6
            self._signal = self._coeff4 * (sample - 2 * self._previous_sample1 + self._previous_sample2) / 4 + self._coeff5 * self._signal
            self._previous_sample3 = self._previous_sample2
            self._previous_sample2 = self._previous_sample1
            self._previous_sample1 = sample
            return math.nan
        elif self._count == 6:
            self._smoothed = (sample + 2 * self._previous_sample1 + 2 * self._previous_sample2 + self._previous_sample3) / 6
            self._previous_value2 = (sample - 2 * self._previous_sample1 + self._previous_sample2) / 4
            self._signal = self._coeff4 * self._previous_value2 + self._coeff5 * self._signal
            self._previous_sample3 = self._previous_sample2
            self._previous_sample2 = self._previous_sample1
            self._previous_sample1 = sample
            return math.nan
        elif self._count == 7:
            self._previous_smoothed2 = self._previous_smoothed1
            self._previous_smoothed1 = self._smoothed
            self._smoothed = (sample + 2 * self._previous_sample1 + 2 * self._previous_sample2 + self._previous_sample3) / 6
            self._previous_value1 = (sample - 2 * self._previous_sample1 + self._previous_sample2) / 4
            self._signal = self._coeff4 * self._previous_value1 + self._coeff5 * self._signal
            self._previous_sample3 = self._previous_sample2
            self._previous_sample2 = self._previous_sample1
            self._previous_sample1 = sample
            return math.nan
        elif self._count == 8:
            self._previous_smoothed2 = self._previous_smoothed1
            self._previous_smoothed1 = self._smoothed
            self._smoothed = (sample + 2 * self._previous_sample1 + 2 * self._previous_sample2 + self._previous_sample3) / 6

            self._value = self._coeff1 * (self._smoothed - 2 * self._previous_smoothed1 + self._previous_smoothed2) + \
                self._coeff2 * self._previous_value1 + self._coeff3 * self._previous_value2

            self._signal = self._coeff4 * self._value + self._coeff5 * self._signal

            self._previous_sample3 = self._previous_sample2
            self._previous_sample2 = self._previous_sample1
            self._previous_sample1 = sample
            self._primed = True

            return self._value

        return math.nan

    def update_scalar(self, sample: Scalar) -> Output:
        return self._update_entity(sample.time, sample.value)

    def update_bar(self, sample: Bar) -> Output:
        return self._update_entity(sample.time, self._bar_func(sample))

    def update_quote(self, sample: Quote) -> Output:
        return self._update_entity(sample.time, self._quote_func(sample))

    def update_trade(self, sample: Trade) -> Output:
        return self._update_entity(sample.time, self._trade_func(sample))

    def _update_entity(self, time, sample: float) -> Output:
        v = self.update(sample)
        signal = self._signal if not math.isnan(v) else math.nan
        return [
            Scalar(time=time, value=v),
            Scalar(time=time, value=signal),
        ]


def _new_cyber_cycle(length: int, alpha: float, signal_lag: int,
                     bc, qc, tc) -> CyberCycle:
    invalid = "invalid cyber cycle parameters"

    if math.isnan(alpha):
        # Length-based construction.
        if length < 1:
            raise ValueError(f"{invalid}: length should be a positive integer")
        alpha = 2.0 / (1 + length)
    else:
        # Smoothing-factor-based construction.
        if alpha < 0 or alpha > 1:
            raise ValueError(f"{invalid}: smoothing factor should be in range [0, 1]")
        if alpha < _EPSILON:
            length = 2**63
        else:
            length = int(round(2.0 / alpha)) - 1

    if signal_lag < 1:
        raise ValueError(f"{invalid}: signal lag should be a positive integer")

    # Default bar component is MedianPrice.
    bc = bc if bc is not None else BarComponent.MEDIAN
    qc = qc if qc is not None else DEFAULT_QUOTE_COMPONENT
    tc = tc if tc is not None else DEFAULT_TRADE_COMPONENT

    bar_func = bar_component_value(bc)
    quote_func = quote_component_value(qc)
    trade_func = trade_component_value(tc)

    component_mnemonic = component_triple_mnemonic(bc, qc, tc)
    mnemonic = f"cc({length}{component_mnemonic})"
    mnemonic_signal = f"ccSignal({length}{component_mnemonic})"
    desc = f"Cyber Cycle {mnemonic}"
    desc_signal = f"Cyber Cycle signal {mnemonic_signal}"

    # Calculate coefficients.
    x = 1 - alpha / 2
    c1 = x * x

    x = 1 - alpha
    c2 = 2 * x
    c3 = -x * x

    x = 1 / (1 + signal_lag)
    c4 = x
    c5 = 1 - x

    return CyberCycle(length, alpha, signal_lag, c1, c2, c3, c4, c5,
                       mnemonic, desc, mnemonic_signal, desc_signal,
                       bar_func, quote_func, trade_func)
