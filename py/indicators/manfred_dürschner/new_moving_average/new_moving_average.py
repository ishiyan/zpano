"""New moving average indicator (Dürschner's lag-free moving average)."""

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
from .params import NewMovingAverageParams, MAType


class _StreamingSMA:
    """Streaming simple moving average."""

    def __init__(self, period: int) -> None:
        self._period = period
        self._buffer: list[float] = [0.0] * period
        self._buffer_index: int = 0
        self._buffer_count: int = 0
        self._sum: float = 0.0
        self._primed: bool = False

    def update(self, sample: float) -> float:
        if math.isnan(sample):
            return sample
        period = self._period
        if self._primed:
            self._sum -= self._buffer[self._buffer_index]
        self._buffer[self._buffer_index] = sample
        self._sum += sample
        self._buffer_index = (self._buffer_index + 1) % period
        if not self._primed:
            self._buffer_count += 1
            if self._buffer_count < period:
                return math.nan
            self._primed = True
        return self._sum / period


class _StreamingEMA:
    """Streaming exponential moving average (SMA-seeded)."""

    def __init__(self, period: int) -> None:
        self._period = period
        self._multiplier: float = 2.0 / (period + 1)
        self._count: int = 0
        self._sum: float = 0.0
        self._value: float = math.nan
        self._primed: bool = False

    def update(self, sample: float) -> float:
        if math.isnan(sample):
            return sample
        if not self._primed:
            self._count += 1
            self._sum += sample
            if self._count < self._period:
                return math.nan
            self._value = self._sum / self._period
            self._primed = True
            return self._value
        self._value = (sample - self._value) * self._multiplier + self._value
        return self._value


class _StreamingSMMA:
    """Streaming smoothed moving average (SMA-seeded)."""

    def __init__(self, period: int) -> None:
        self._period = period
        self._count: int = 0
        self._sum: float = 0.0
        self._value: float = math.nan
        self._primed: bool = False

    def update(self, sample: float) -> float:
        if math.isnan(sample):
            return sample
        if not self._primed:
            self._count += 1
            self._sum += sample
            if self._count < self._period:
                return math.nan
            self._value = self._sum / self._period
            self._primed = True
            return self._value
        self._value = (self._value * (self._period - 1) + sample) / self._period
        return self._value


class _StreamingLWMA:
    """Streaming linear weighted moving average."""

    def __init__(self, period: int) -> None:
        self._period = period
        self._buffer: list[float] = [0.0] * period
        self._buffer_index: int = 0
        self._buffer_count: int = 0
        self._weight_sum: float = period * (period + 1) / 2.0
        self._primed: bool = False

    def update(self, sample: float) -> float:
        if math.isnan(sample):
            return sample
        period = self._period
        self._buffer[self._buffer_index] = sample
        self._buffer_index = (self._buffer_index + 1) % period
        if not self._primed:
            self._buffer_count += 1
            if self._buffer_count < period:
                return math.nan
            self._primed = True
        # Compute weighted sum: oldest gets weight 1, newest gets weight period.
        # Oldest is at self._buffer_index (circular buffer).
        result = 0.0
        index = self._buffer_index
        for i in range(period):
            result += (i + 1) * self._buffer[index]
            index = (index + 1) % period
        return result / self._weight_sum


def _create_streaming_ma(ma_type: MAType, period: int):
    """Create a streaming moving average of the given type."""
    if ma_type == MAType.SMA:
        return _StreamingSMA(period)
    elif ma_type == MAType.EMA:
        return _StreamingEMA(period)
    elif ma_type == MAType.SMMA:
        return _StreamingSMMA(period)
    elif ma_type == MAType.LWMA:
        return _StreamingLWMA(period)
    else:
        raise ValueError(f"unknown MA type: {ma_type}")


class NewMovingAverage(Indicator):
    """Computes the New Moving Average (NMA) by Dürschner.

    NMA applies the Nyquist-Shannon sampling theorem to moving average design:
    by cascading two moving averages whose period ratio satisfies the Nyquist
    criterion (lambda = n1/n2 >= 2), the resulting lag can be extrapolated away
    geometrically.

    Formula: NMA = (1 + alpha) * MA1 - alpha * MA2
    where: alpha = lambda * (n1-1) / (n1-lambda), lambda = n1 // n2
    """

    def __init__(self, params: NewMovingAverageParams) -> None:
        primary_period = params.primary_period
        secondary_period = params.secondary_period

        # Enforce Nyquist constraint.
        if primary_period < 4:
            primary_period = 4
        if secondary_period < 2:
            secondary_period = 2
        if primary_period < secondary_period * 2:
            primary_period = secondary_period * 4

        # Compute alpha.
        nyquist_ratio = primary_period // secondary_period
        alpha = nyquist_ratio * (primary_period - 1) / (primary_period - nyquist_ratio)

        bc = params.bar_component if params.bar_component is not None else DEFAULT_BAR_COMPONENT
        qc = params.quote_component if params.quote_component is not None else DEFAULT_QUOTE_COMPONENT
        tc = params.trade_component if params.trade_component is not None else DEFAULT_TRADE_COMPONENT

        bar_func = bar_component_value(bc)
        quote_func = quote_component_value(qc)
        trade_func = trade_component_value(tc)

        mnemonic = f"nma({primary_period}, {secondary_period}, " \
            f"{params.ma_type}{component_triple_mnemonic(bc, qc, tc)})"
        description = f"New moving average {mnemonic}"

        self._line = LineIndicator(mnemonic, description, bar_func, quote_func, trade_func, self.update)

        self._alpha: float = alpha
        self._ma_primary = _create_streaming_ma(params.ma_type, primary_period)
        self._ma_secondary = _create_streaming_ma(params.ma_type, secondary_period)
        self._primed: bool = False

    def is_primed(self) -> bool:
        return self._primed

    def metadata(self) -> Metadata:
        return build_metadata(
            Identifier.NEW_MOVING_AVERAGE,
            self._line.mnemonic,
            self._line.description,
            [OutputText(self._line.mnemonic, self._line.description)],
        )

    def update(self, sample: float) -> float:
        if math.isnan(sample):
            return sample

        # First filter: MA of raw price.
        ma1_value = self._ma_primary.update(sample)
        if math.isnan(ma1_value):
            return math.nan

        # Second filter: MA of MA1 output.
        ma2_value = self._ma_secondary.update(ma1_value)
        if math.isnan(ma2_value):
            return math.nan

        self._primed = True

        # Geometric extrapolation.
        alpha = self._alpha
        return (1.0 + alpha) * ma1_value - alpha * ma2_value

    def update_scalar(self, sample: Scalar) -> Output:
        return self._line.update_scalar(sample)

    def update_bar(self, sample: Bar) -> Output:
        return self._line.update_bar(sample)

    def update_quote(self, sample: Quote) -> Output:
        return self._line.update_quote(sample)

    def update_trade(self, sample: Trade) -> Output:
        return self._line.update_trade(sample)
