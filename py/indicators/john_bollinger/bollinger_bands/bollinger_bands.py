"""Bollinger Bands indicator."""

import math
from typing import List, Any

from ...core.indicator import Indicator
from ...core.metadata import Metadata
from ...core.build_metadata import build_metadata, OutputText
from ...core.identifier import Identifier
from ...core.component_triple_mnemonic import component_triple_mnemonic
from ...core.outputs.band import Band
from ....entities.bar import Bar
from ....entities.quote import Quote
from ....entities.trade import Trade
from ....entities.scalar import Scalar
from ....entities.bar_component import BarComponent, DEFAULT_BAR_COMPONENT, bar_component_value
from ....entities.quote_component import QuoteComponent, DEFAULT_QUOTE_COMPONENT, quote_component_value
from ....entities.trade_component import TradeComponent, DEFAULT_TRADE_COMPONENT, trade_component_value
from ...common.simple_moving_average.simple_moving_average import SimpleMovingAverage
from ...common.simple_moving_average.params import SimpleMovingAverageParams
from ...common.exponential_moving_average.exponential_moving_average import ExponentialMovingAverage
from ...common.exponential_moving_average.params import ExponentialMovingAverageLengthParams
from ...common.variance.variance import Variance
from ...common.variance.params import VarianceParams
from .params import BollingerBandsParams, MovingAverageType


class BollingerBands(Indicator):
    """John Bollinger's Bollinger Bands indicator.

    Bollinger Bands consist of a middle band (moving average) and upper/lower bands
    placed a specified number of standard deviations above and below the middle band.

    The indicator produces six outputs:
      - LowerValue: middleValue - lowerMultiplier * stddev
      - MiddleValue: moving average of the input
      - UpperValue: middleValue + upperMultiplier * stddev
      - BandWidth: (upperValue - lowerValue) / middleValue
      - PercentBand: (sample - lowerValue) / (upperValue - lowerValue)
      - Band: lower/upper band pair

    Reference:
        Bollinger, John (2002). Bollinger on Bollinger Bands. McGraw-Hill.
    """

    def __init__(self, p: BollingerBandsParams) -> None:
        length = p.length
        if length < 2:
            raise ValueError(
                "invalid bollinger bands parameters: "
                "length should be greater than 1")

        upper_multiplier = p.upper_multiplier
        if upper_multiplier == 0:
            upper_multiplier = 2.0

        lower_multiplier = p.lower_multiplier
        if lower_multiplier == 0:
            lower_multiplier = 2.0

        is_unbiased = True if p.is_unbiased is None else p.is_unbiased

        bc = p.bar_component if p.bar_component is not None else DEFAULT_BAR_COMPONENT
        qc = p.quote_component if p.quote_component is not None else DEFAULT_QUOTE_COMPONENT
        tc = p.trade_component if p.trade_component is not None else DEFAULT_TRADE_COMPONENT

        self._bar_func = bar_component_value(bc)
        self._quote_func = quote_component_value(qc)
        self._trade_func = trade_component_value(tc)

        # Create variance sub-indicator.
        self._variance = Variance(VarianceParams(
            length=length,
            is_unbiased=is_unbiased,
            bar_component=p.bar_component,
            quote_component=p.quote_component,
            trade_component=p.trade_component,
        ))

        # Create moving average sub-indicator.
        if p.moving_average_type == MovingAverageType.EMA:
            self._ma = ExponentialMovingAverage.from_length(
                ExponentialMovingAverageLengthParams(
                    length=length, first_is_average=p.first_is_average))
        else:
            self._ma = SimpleMovingAverage(
                SimpleMovingAverageParams(length=length))

        self._upper_multiplier = upper_multiplier
        self._lower_multiplier = lower_multiplier

        self._middle_value = math.nan
        self._upper_value = math.nan
        self._lower_value = math.nan
        self._band_width = math.nan
        self._percent_band = math.nan
        self._primed = False

        self._mnemonic = f"bb({length},{upper_multiplier:.0f},{lower_multiplier:.0f}" \
                         f"{component_triple_mnemonic(bc, qc, tc)})"

    def is_primed(self) -> bool:
        return self._primed

    def metadata(self) -> Metadata:
        desc = f"Bollinger Bands {self._mnemonic}"
        return build_metadata(
            Identifier.BOLLINGER_BANDS,
            self._mnemonic,
            desc,
            [
                OutputText(f"{self._mnemonic} lower", f"{desc} Lower"),
                OutputText(f"{self._mnemonic} middle", f"{desc} Middle"),
                OutputText(f"{self._mnemonic} upper", f"{desc} Upper"),
                OutputText(f"{self._mnemonic} bandWidth", f"{desc} Band Width"),
                OutputText(f"{self._mnemonic} percentBand", f"{desc} Percent Band"),
                OutputText(f"{self._mnemonic} band", f"{desc} Band"),
            ],
        )

    def update(self, sample: float) -> tuple[float, float, float, float, float]:
        """Update with a scalar value. Returns (lower, middle, upper, bandWidth, percentBand)."""
        nan = math.nan

        if math.isnan(sample):
            return nan, nan, nan, nan, nan

        middle = self._ma.update(sample)
        v = self._variance.update(sample)

        self._primed = self._ma.is_primed() and self._variance.is_primed()

        if math.isnan(middle) or math.isnan(v):
            self._middle_value = nan
            self._upper_value = nan
            self._lower_value = nan
            self._band_width = nan
            self._percent_band = nan
            return nan, nan, nan, nan, nan

        stddev = math.sqrt(v)
        upper = middle + self._upper_multiplier * stddev
        lower = middle - self._lower_multiplier * stddev

        epsilon = 1e-10

        if abs(middle) < epsilon:
            bw = 0.0
        else:
            bw = (upper - lower) / middle

        spread = upper - lower
        if abs(spread) < epsilon:
            pct_b = 0.0
        else:
            pct_b = (sample - lower) / spread

        self._middle_value = middle
        self._upper_value = upper
        self._lower_value = lower
        self._band_width = bw
        self._percent_band = pct_b

        return lower, middle, upper, bw, pct_b

    def update_scalar(self, sample: Scalar) -> List[Any]:
        lower, middle, upper, bw, pct_b = self.update(sample.value)

        output: List[Any] = [
            Scalar(time=sample.time, value=lower),
            Scalar(time=sample.time, value=middle),
            Scalar(time=sample.time, value=upper),
            Scalar(time=sample.time, value=bw),
            Scalar(time=sample.time, value=pct_b),
        ]

        if math.isnan(lower) or math.isnan(upper):
            output.append(Band.empty(sample.time))
        else:
            output.append(Band(sample.time, lower, upper))

        return output

    def update_bar(self, sample: Bar) -> List[Any]:
        v = self._bar_func(sample)
        return self.update_scalar(Scalar(time=sample.time, value=v))

    def update_quote(self, sample: Quote) -> List[Any]:
        v = self._quote_func(sample)
        return self.update_scalar(Scalar(time=sample.time, value=v))

    def update_trade(self, sample: Trade) -> List[Any]:
        v = self._trade_func(sample)
        return self.update_scalar(Scalar(time=sample.time, value=v))
