"""Ehler's Instantaneous Trend Line (iTrend) indicator."""

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
from .params import InstantaneousTrendLineLengthParams, InstantaneousTrendLineSmoothingFactorParams


_EPSILON = 0.00000001


class InstantaneousTrendLine(Indicator):
    """Ehler's Instantaneous Trend Line (iTrend).

    H(z) = ((alpha - alpha^2/4) + alpha^2*z^-1/2 - (alpha - 3*alpha^2/4)*z^-2)
            / (1 - 2*(1-alpha)*z^-1 + (1-alpha)^2*z^-2)

    The indicator has two outputs: the trend line value and a trigger line.
    """

    def __init__(self, length: int, smoothing_factor: float,
                 coeff1: float, coeff2: float, coeff3: float,
                 coeff4: float, coeff5: float,
                 mnemonic: str, description: str,
                 mnemonic_trig: str, description_trig: str,
                 bar_func, quote_func, trade_func) -> None:
        self._length = length
        self._smoothing_factor = smoothing_factor
        self._coeff1 = coeff1
        self._coeff2 = coeff2
        self._coeff3 = coeff3
        self._coeff4 = coeff4
        self._coeff5 = coeff5
        self._mnemonic = mnemonic
        self._description = description
        self._mnemonic_trig = mnemonic_trig
        self._description_trig = description_trig
        self._count = 0
        self._previous_sample1 = 0.0
        self._previous_sample2 = 0.0
        self._previous_trend_line1 = 0.0
        self._previous_trend_line2 = 0.0
        self._trend_line = math.nan
        self._trigger_line = math.nan
        self._primed = False
        self._bar_func = bar_func
        self._quote_func = quote_func
        self._trade_func = trade_func

    @staticmethod
    def from_length(params: InstantaneousTrendLineLengthParams) -> 'InstantaneousTrendLine':
        """Creates a new InstantaneousTrendLine from length parameters."""
        return _new_itrend(params.length, math.nan,
                           params.bar_component, params.quote_component, params.trade_component)

    @staticmethod
    def from_smoothing_factor(params: InstantaneousTrendLineSmoothingFactorParams) -> 'InstantaneousTrendLine':
        """Creates a new InstantaneousTrendLine from smoothing factor parameters."""
        return _new_itrend(0, params.smoothing_factor,
                           params.bar_component, params.quote_component, params.trade_component)

    def is_primed(self) -> bool:
        return self._primed

    def metadata(self) -> Metadata:
        return build_metadata(
            Identifier.INSTANTANEOUS_TREND_LINE,
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
            self._trend_line = self._coeff1 * sample + self._coeff2 * self._previous_sample1 + \
                self._coeff3 * self._previous_sample2 + \
                self._coeff4 * self._previous_trend_line1 + self._coeff5 * self._previous_trend_line2
            self._trigger_line = 2 * self._trend_line - self._previous_trend_line2

            self._previous_sample2 = self._previous_sample1
            self._previous_sample1 = sample
            self._previous_trend_line2 = self._previous_trend_line1
            self._previous_trend_line1 = self._trend_line

            return self._trend_line

        self._count += 1

        if self._count == 1:
            self._previous_sample2 = sample
            return math.nan
        elif self._count == 2:
            self._previous_sample1 = sample
            return math.nan
        elif self._count == 3:
            self._previous_trend_line2 = (sample + 2 * self._previous_sample1 + self._previous_sample2) / 4
            self._previous_sample2 = self._previous_sample1
            self._previous_sample1 = sample
            return math.nan
        elif self._count == 4:
            self._previous_trend_line1 = (sample + 2 * self._previous_sample1 + self._previous_sample2) / 4
            self._previous_sample2 = self._previous_sample1
            self._previous_sample1 = sample
            return math.nan
        elif self._count == 5:
            self._trend_line = self._coeff1 * sample + self._coeff2 * self._previous_sample1 + \
                self._coeff3 * self._previous_sample2 + \
                self._coeff4 * self._previous_trend_line1 + self._coeff5 * self._previous_trend_line2
            self._trigger_line = 2 * self._trend_line - self._previous_trend_line2

            self._previous_sample2 = self._previous_sample1
            self._previous_sample1 = sample
            self._previous_trend_line2 = self._previous_trend_line1
            self._previous_trend_line1 = self._trend_line
            self._primed = True

            return self._trend_line

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
        trig = self._trigger_line if not math.isnan(v) else math.nan
        return [
            Scalar(time=time, value=v),
            Scalar(time=time, value=trig),
        ]


def _new_itrend(length: int, alpha: float,
                bc, qc, tc) -> InstantaneousTrendLine:
    invalid = "invalid instantaneous trend line parameters"

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

    # Default bar component is MedianPrice.
    bc = bc if bc is not None else BarComponent.MEDIAN
    qc = qc if qc is not None else DEFAULT_QUOTE_COMPONENT
    tc = tc if tc is not None else DEFAULT_TRADE_COMPONENT

    bar_func = bar_component_value(bc)
    quote_func = quote_component_value(qc)
    trade_func = trade_component_value(tc)

    component_mnemonic = component_triple_mnemonic(bc, qc, tc)
    mnemonic = f"iTrend({length}{component_mnemonic})"
    mnemonic_trig = f"iTrendTrigger({length}{component_mnemonic})"
    desc = f"Instantaneous Trend Line {mnemonic}"
    desc_trig = f"Instantaneous Trend Line trigger {mnemonic_trig}"

    # Calculate coefficients.
    a2 = alpha * alpha
    c1 = alpha - a2 / 4
    c2 = a2 / 2
    c3 = -(alpha - 3 * a2 / 4)
    x = 1 - alpha
    c4 = 2 * x
    c5 = -(x * x)

    return InstantaneousTrendLine(length, alpha, c1, c2, c3, c4, c5,
                                   mnemonic, desc, mnemonic_trig, desc_trig,
                                   bar_func, quote_func, trade_func)
