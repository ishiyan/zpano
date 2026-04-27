"""Marc Chaikin's Advance-Decline Oscillator (ADOSC)."""

import math

from .params import AdvanceDeclineOscillatorParams, MovingAverageType
from ...common.simple_moving_average.simple_moving_average import SimpleMovingAverage
from ...common.simple_moving_average.params import SimpleMovingAverageParams
from ...common.exponential_moving_average.exponential_moving_average import ExponentialMovingAverage
from ...common.exponential_moving_average.params import ExponentialMovingAverageLengthParams
from ...core.line_indicator import LineIndicator
from ...core.indicator import Indicator
from ...core.identifier import Identifier
from ...core.metadata import Metadata
from ...core.build_metadata import build_metadata, OutputText
from ...core.output import Output
from ....entities.bar import Bar
from ....entities.quote import Quote
from ....entities.trade import Trade
from ....entities.scalar import Scalar
from ....entities.bar_component import DEFAULT_BAR_COMPONENT, bar_component_value
from ....entities.quote_component import DEFAULT_QUOTE_COMPONENT, quote_component_value
from ....entities.trade_component import DEFAULT_TRADE_COMPONENT, trade_component_value


class AdvanceDeclineOscillator(Indicator):
    """Marc Chaikin's Advance-Decline Oscillator (ADOSC).

    ADOSC = FastMA(AD) - SlowMA(AD)

    where AD is the cumulative Advance-Decline line:
    CLV = ((Close - Low) - (High - Close)) / (High - Low)
    AD = AD_prev + CLV * Volume
    """

    def __init__(self, params: AdvanceDeclineOscillatorParams) -> None:
        if params.fast_length < 2:
            raise ValueError("invalid advance-decline oscillator parameters: "
                             "fast length should be greater than 1")
        if params.slow_length < 2:
            raise ValueError("invalid advance-decline oscillator parameters: "
                             "slow length should be greater than 1")

        if params.moving_average_type == MovingAverageType.SMA:
            ma_label = "SMA"
            self._fast_ma = SimpleMovingAverage(
                SimpleMovingAverageParams(length=params.fast_length))
            self._slow_ma = SimpleMovingAverage(
                SimpleMovingAverageParams(length=params.slow_length))
        else:
            ma_label = "EMA"
            self._fast_ma = ExponentialMovingAverage.from_length(
                ExponentialMovingAverageLengthParams(
                    length=params.fast_length,
                    first_is_average=params.first_is_average))
            self._slow_ma = ExponentialMovingAverage.from_length(
                ExponentialMovingAverageLengthParams(
                    length=params.slow_length,
                    first_is_average=params.first_is_average))

        mnemonic = f"adosc({ma_label}{params.fast_length}/{ma_label}{params.slow_length})"
        description = "Chaikin Advance-Decline Oscillator " + mnemonic

        bar_func = bar_component_value(DEFAULT_BAR_COMPONENT)
        quote_func = quote_component_value(DEFAULT_QUOTE_COMPONENT)
        trade_func = trade_component_value(DEFAULT_TRADE_COMPONENT)

        self._line = LineIndicator(mnemonic, description, bar_func, quote_func, trade_func, self.update)
        self._ad = 0.0
        self._value = math.nan
        self._primed = False

    def is_primed(self) -> bool:
        return self._primed

    def metadata(self) -> Metadata:
        return build_metadata(
            Identifier.ADVANCE_DECLINE_OSCILLATOR,
            self._line.mnemonic,
            self._line.description,
            [OutputText(self._line.mnemonic, self._line.description)],
        )

    def update(self, sample: float) -> float:
        """Update with scalar (H=L=C, volume=1, AD unchanged but fed to MAs)."""
        if math.isnan(sample):
            return math.nan
        return self.update_hlcv(sample, sample, sample, 1.0)

    def update_hlcv(self, high: float, low: float, close: float, volume: float) -> float:
        """Update with high, low, close, volume."""
        if math.isnan(high) or math.isnan(low) or math.isnan(close) or math.isnan(volume):
            return math.nan

        temp = high - low
        if temp > 0:
            self._ad += ((close - low) - (high - close)) / temp * volume

        fast = self._fast_ma.update(self._ad)
        slow = self._slow_ma.update(self._ad)
        self._primed = self._fast_ma.is_primed() and self._slow_ma.is_primed()

        if math.isnan(fast) or math.isnan(slow):
            self._value = math.nan
            return self._value

        self._value = fast - slow
        return self._value

    def update_scalar(self, sample: Scalar) -> Output:
        return self._line.update_scalar(sample)

    def update_bar(self, sample: Bar) -> Output:
        """Shadows LineIndicator.update_bar to extract HLCV."""
        value = self.update_hlcv(sample.high, sample.low, sample.close, sample.volume)
        return [Scalar(time=sample.time, value=value)]

    def update_quote(self, sample: Quote) -> Output:
        return self._line.update_quote(sample)

    def update_trade(self, sample: Trade) -> Output:
        return self._line.update_trade(sample)
