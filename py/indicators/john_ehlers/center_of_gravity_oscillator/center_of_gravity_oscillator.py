"""Ehler's Center of Gravity oscillator (COG) indicator."""

import math
import sys

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
from .params import CenterOfGravityOscillatorParams


class CenterOfGravityOscillator(Indicator):
    """Ehler's Center of Gravity oscillator (COG).

    CG_i = sum((i+1) * Price_i) / sum(Price_i), where i = 0..L-1

    The indicator has two outputs: the oscillator value and a trigger line
    which is the previous value of the oscillator.
    """

    def __init__(self, length: int,
                 mnemonic: str, description: str,
                 mnemonic_trig: str, description_trig: str,
                 bar_func, quote_func, trade_func) -> None:
        self._length = length
        self._length_min_one = length - 1
        self._mnemonic = mnemonic
        self._description = description
        self._mnemonic_trig = mnemonic_trig
        self._description_trig = description_trig
        self._window: list[float] = [0.0] * length
        self._window_count = 0
        self._denominator_sum = 0.0
        self._value = math.nan
        self._value_previous = math.nan
        self._primed = False
        self._bar_func = bar_func
        self._quote_func = quote_func
        self._trade_func = trade_func

    @staticmethod
    def create(params: CenterOfGravityOscillatorParams) -> 'CenterOfGravityOscillator':
        """Creates a new CenterOfGravityOscillator from parameters."""
        return _new_cog(params)

    def is_primed(self) -> bool:
        return self._primed

    def metadata(self) -> Metadata:
        return build_metadata(
            Identifier.CENTER_OF_GRAVITY_OSCILLATOR,
            self._mnemonic,
            self._description,
            [
                OutputText(self._mnemonic, self._description),
                OutputText(self._mnemonic_trig, self._description_trig),
            ],
        )

    def update(self, sample: float) -> float:
        if math.isnan(sample):
            return math.nan

        if self._primed:
            self._value_previous = self._value
            self._value = self._calculate(sample)
            return self._value

        # Not primed.
        if self._length > self._window_count:
            self._denominator_sum += sample
            self._window[self._window_count] = sample

            if self._length_min_one == self._window_count:
                s = 0.0
                if abs(self._denominator_sum) > sys.float_info.min:
                    for i in range(self._length):
                        s += (1 + i) * self._window[i]
                    s /= self._denominator_sum
                self._value_previous = s
        else:
            self._value = self._calculate(sample)
            self._primed = True
            self._window_count += 1
            return self._value

        self._window_count += 1
        return math.nan

    def _calculate(self, sample: float) -> float:
        self._denominator_sum += sample - self._window[0]

        for i in range(self._length_min_one):
            self._window[i] = self._window[i + 1]
        self._window[self._length_min_one] = sample

        s = 0.0
        if abs(self._denominator_sum) > sys.float_info.min:
            for i in range(self._length):
                s += (1 + i) * self._window[i]
            s /= self._denominator_sum

        return s

    def update_scalar(self, sample: Scalar) -> Output:
        return self._update_entity(sample.time, sample.value)

    def update_bar(self, sample: Bar) -> Output:
        return self._update_entity(sample.time, self._bar_func(sample))

    def update_quote(self, sample: Quote) -> Output:
        return self._update_entity(sample.time, self._quote_func(sample))

    def update_trade(self, sample: Trade) -> Output:
        return self._update_entity(sample.time, self._trade_func(sample))

    def _update_entity(self, time, sample: float) -> Output:
        cog = self.update(sample)
        trig = self._value_previous if not math.isnan(cog) else math.nan
        return [
            Scalar(time=time, value=cog),
            Scalar(time=time, value=trig),
        ]


def _new_cog(p: CenterOfGravityOscillatorParams) -> CenterOfGravityOscillator:
    invalid = "invalid center of gravity oscillator parameters"

    if p.length < 1:
        raise ValueError(f"{invalid}: length should be a positive integer")

    # Default bar component is MedianPrice.
    bc = p.bar_component if p.bar_component is not None else BarComponent.MEDIAN
    qc = p.quote_component if p.quote_component is not None else DEFAULT_QUOTE_COMPONENT
    tc = p.trade_component if p.trade_component is not None else DEFAULT_TRADE_COMPONENT

    bar_func = bar_component_value(bc)
    quote_func = quote_component_value(qc)
    trade_func = trade_component_value(tc)

    component_mnemonic = component_triple_mnemonic(bc, qc, tc)
    mnemonic = f"cog({p.length}{component_mnemonic})"
    mnemonic_trig = f"cogTrig({p.length}{component_mnemonic})"
    desc = f"Center of Gravity oscillator {mnemonic}"
    desc_trig = f"Center of Gravity trigger {mnemonic_trig}"

    return CenterOfGravityOscillator(p.length, mnemonic, desc, mnemonic_trig, desc_trig,
                                      bar_func, quote_func, trade_func)
