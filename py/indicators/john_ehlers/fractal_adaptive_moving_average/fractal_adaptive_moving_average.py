"""Ehlers' Fractal Adaptive Moving Average (FRAMA) indicator."""

import math
import sys
from typing import Optional

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
from ....entities.bar_component import BarComponent, DEFAULT_BAR_COMPONENT, bar_component_value
from ....entities.quote_component import QuoteComponent, DEFAULT_QUOTE_COMPONENT, quote_component_value
from ....entities.trade_component import TradeComponent, DEFAULT_TRADE_COMPONENT, trade_component_value
from .params import FractalAdaptiveMovingAverageParams


_LOG2E = math.log2(math.e)


class FractalAdaptiveMovingAverage(Indicator):
    """Ehlers' Fractal Adaptive Moving Average (FRAMA).

    An EMA with the smoothing factor adapted based on the estimated fractal
    dimension of the price series.

    Two outputs:
      - Value: the FRAMA value.
      - Fdim: the estimated fractal dimension.
    """

    def __init__(self, length: int, half_length: int,
                 alpha_slowest: float, scaling_factor: float,
                 mnemonic: str, description: str,
                 mnemonic_fdim: str, description_fdim: str,
                 bar_func, quote_func, trade_func) -> None:
        self._length = length
        self._length_min_one = length - 1
        self._half_length = half_length
        self._alpha_slowest = alpha_slowest
        self._scaling_factor = scaling_factor
        self._mnemonic = mnemonic
        self._description = description
        self._mnemonic_fdim = mnemonic_fdim
        self._description_fdim = description_fdim
        self._bar_func = bar_func
        self._quote_func = quote_func
        self._trade_func = trade_func
        self._window_high = [0.0] * length
        self._window_low = [0.0] * length
        self._window_count = 0
        self._fractal_dimension = math.nan
        self._value = math.nan
        self._primed = False

    @staticmethod
    def create(params: FractalAdaptiveMovingAverageParams) -> 'FractalAdaptiveMovingAverage':
        """Creates a new instance from parameters."""
        return _new_frama(params)

    @staticmethod
    def create_default() -> 'FractalAdaptiveMovingAverage':
        """Creates a new instance with default parameters."""
        return _new_frama(FractalAdaptiveMovingAverageParams())

    @property
    def fractal_dimension(self) -> float:
        """Returns the current fractal dimension estimate."""
        return self._fractal_dimension

    def is_primed(self) -> bool:
        return self._primed

    def metadata(self) -> Metadata:
        return build_metadata(
            Identifier.FRACTAL_ADAPTIVE_MOVING_AVERAGE,
            self._mnemonic,
            self._description,
            [
                OutputText(self._mnemonic, self._description),
                OutputText(self._mnemonic_fdim, self._description_fdim),
            ],
        )

    def update(self, sample: float, sample_high: float, sample_low: float) -> float:
        """Updates the FRAMA given the next sample and its high/low values."""
        if math.isnan(sample_high) or math.isnan(sample_low) or math.isnan(sample):
            return math.nan

        if self._primed:
            wh = self._window_high
            wl = self._window_low
            lm1 = self._length_min_one
            for i in range(lm1):
                wh[i] = wh[i + 1]
                wl[i] = wl[i + 1]
            wh[lm1] = sample_high
            wl[lm1] = sample_low

            self._fractal_dimension = self._estimate_fractal_dimension()
            alpha = self._estimate_alpha()
            self._value += (sample - self._value) * alpha
            return self._value
        else:
            wc = self._window_count
            self._window_high[wc] = sample_high
            self._window_low[wc] = sample_low
            wc += 1
            self._window_count = wc

            if wc == self._length_min_one:
                self._value = sample
            elif wc == self._length:
                self._fractal_dimension = self._estimate_fractal_dimension()
                alpha = self._estimate_alpha()
                self._value += (sample - self._value) * alpha
                self._primed = True
                return self._value

        return math.nan

    def update_scalar(self, sample: Scalar) -> Output:
        v = sample.value
        return self._update_entity(sample.time, v, v, v)

    def update_bar(self, sample: Bar) -> Output:
        v = self._bar_func(sample)
        return self._update_entity(sample.time, v, sample.high, sample.low)

    def update_quote(self, sample: Quote) -> Output:
        v = self._quote_func(sample)
        return self._update_entity(sample.time, v, sample.ask_price, sample.bid_price)

    def update_trade(self, sample: Trade) -> Output:
        v = self._trade_func(sample)
        return self._update_entity(sample.time, v, v, v)

    def _update_entity(self, time, sample: float,
                       sample_high: float, sample_low: float) -> Output:
        frama = self.update(sample, sample_high, sample_low)
        fdim = self._fractal_dimension if not math.isnan(frama) else math.nan
        return [
            Scalar(time=time, value=frama),
            Scalar(time=time, value=fdim),
        ]

    def _estimate_fractal_dimension(self) -> float:
        half = self._half_length
        wh = self._window_high
        wl = self._window_low

        min_low_half = sys.float_info.max
        max_high_half = sys.float_info.min

        for i in range(half):
            l = wl[i]
            if min_low_half > l:
                min_low_half = l
            h = wh[i]
            if max_high_half < h:
                max_high_half = h

        range_n1 = max_high_half - min_low_half
        min_low_full = min_low_half
        max_high_full = max_high_half
        min_low_half = sys.float_info.max
        max_high_half = sys.float_info.min

        for j in range(half):
            i = j + half
            l = wl[i]
            if min_low_full > l:
                min_low_full = l
            if min_low_half > l:
                min_low_half = l
            h = wh[i]
            if max_high_full < h:
                max_high_full = h
            if max_high_half < h:
                max_high_half = h

        range_n2 = max_high_half - min_low_half
        range_n3 = max_high_full - min_low_full

        fdim = (math.log((range_n1 + range_n2) / half) -
                math.log(range_n3 / self._length)) * _LOG2E

        return min(max(fdim, 1.0), 2.0)

    def _estimate_alpha(self) -> float:
        alpha = math.exp(self._scaling_factor * (self._fractal_dimension - 1.0))
        return min(max(alpha, self._alpha_slowest), 1.0)


def _new_frama(params: FractalAdaptiveMovingAverageParams) -> FractalAdaptiveMovingAverage:
    invalid = "invalid fractal adaptive moving average parameters"

    if params.length < 2:
        raise ValueError(f"{invalid}: length should be an even integer larger than 1")

    if params.slowest_smoothing_factor < 0.0 or params.slowest_smoothing_factor > 1.0:
        raise ValueError(f"{invalid}: slowest smoothing factor should be in range [0, 1]")

    length = params.length
    if length % 2 != 0:
        length += 1

    bc = params.bar_component if params.bar_component is not None else DEFAULT_BAR_COMPONENT
    qc = params.quote_component if params.quote_component is not None else DEFAULT_QUOTE_COMPONENT
    tc = params.trade_component if params.trade_component is not None else DEFAULT_TRADE_COMPONENT

    comp = component_triple_mnemonic(bc, qc, tc)
    mnemonic = f"frama({length}, {params.slowest_smoothing_factor:.3f}{comp})"
    mnemonic_fdim = f"framaDim({length}, {params.slowest_smoothing_factor:.3f}{comp})"
    descr = "Fractal adaptive moving average "

    bar_func = bar_component_value(bc)
    quote_func = quote_component_value(qc)
    trade_func = trade_component_value(tc)

    return FractalAdaptiveMovingAverage(
        length, length // 2,
        params.slowest_smoothing_factor,
        math.log(params.slowest_smoothing_factor),
        mnemonic, descr + mnemonic,
        mnemonic_fdim, descr + mnemonic_fdim,
        bar_func, quote_func, trade_func,
    )
