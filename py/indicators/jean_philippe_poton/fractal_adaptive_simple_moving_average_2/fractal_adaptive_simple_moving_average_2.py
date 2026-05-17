"""Fractal Adaptive Simple Moving Average 2 (FRASMA2) indicator."""

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
from .params import FractalAdaptiveSimpleMovingAverage2Params


class FractalAdaptiveSimpleMovingAverage2(Indicator):
    """Computes the Fractal Adaptive Simple Moving Average 2 (FRASMA2).

    Uses the corrected Fractal Graph Dimension Index (FGDI) formula to adaptively modify
    an SMA's period. The corrected formula uses ln(2*(N-1)) denominator and N-1
    path segments. When the market is trending (FGDI near 1.0), the SMA speeds up;
    when erratic (FGDI near 2.0), the SMA slows down.

    The indicator is not primed during the first `period` updates.
    """

    def __init__(self, params: FractalAdaptiveSimpleMovingAverage2Params) -> None:
        period = params.period
        normal_speed = params.normal_speed

        if period < 2:
            raise ValueError(
                "invalid fractal adaptive simple moving average 2 parameters: period should be greater than 1")
        if normal_speed < 1:
            raise ValueError(
                "invalid fractal adaptive simple moving average 2 parameters: normal_speed should be greater than 0")

        bc = params.bar_component if params.bar_component is not None else DEFAULT_BAR_COMPONENT
        qc = params.quote_component if params.quote_component is not None else DEFAULT_QUOTE_COMPONENT
        tc = params.trade_component if params.trade_component is not None else DEFAULT_TRADE_COMPONENT

        bar_func = bar_component_value(bc)
        quote_func = quote_component_value(qc)
        trade_func = trade_component_value(tc)

        mnemonic = f"frasma2({period},{normal_speed}{component_triple_mnemonic(bc, qc, tc)})"
        description = f"Fractal adaptive simple moving average 2 {mnemonic}"

        self._line = LineIndicator(mnemonic, description, bar_func, quote_func, trade_func, self.update)

        self._period: int = period
        self._normal_speed: int = normal_speed
        self._window: list[float] = [0.0] * period
        self._window_count: int = 0
        self._closes: list[float] = []
        self._primed: bool = False

        # Precompute constants for corrected FGDI.
        self._log_2pm1: float = math.log(2.0 * (period - 1))
        self._ln2: float = math.log(2.0)
        self._inv_p_sq: float = 1.0 / (period * period)

    def is_primed(self) -> bool:
        return self._primed

    def metadata(self) -> Metadata:
        return build_metadata(
            Identifier.FRACTAL_ADAPTIVE_SIMPLE_MOVING_AVERAGE_2,
            self._line.mnemonic,
            self._line.description,
            [OutputText(self._line.mnemonic, self._line.description)],
        )

    def update(self, sample: float) -> float:
        if math.isnan(sample):
            return sample

        period = self._period

        # Accumulate close history for SMA computation.
        self._closes.append(sample)

        # Fill the FGDI window. First `period` values are NaN.
        if self._window_count < period:
            self._window[self._window_count] = sample
            self._window_count += 1

            if self._window_count < period:
                return math.nan

            self._primed = True

            return math.nan
        else:
            # Shift window left by one.
            for i in range(period - 1):
                self._window[i] = self._window[i + 1]
            self._window[period - 1] = sample

        # --- Compute FGDI using corrected formula (N-1 segments) ---
        price_max = self._window[0]
        price_min = self._window[0]
        for k in range(1, period):
            if self._window[k] > price_max:
                price_max = self._window[k]
            if self._window[k] < price_min:
                price_min = self._window[k]

        price_range = price_max - price_min

        if price_range <= 0.0:
            return math.nan

        # N-1 segments: iterate from index 1 to period-1.
        prior_norm = (self._window[0] - price_min) / price_range
        length = 0.0

        for k in range(1, period):
            curr_norm = (self._window[k] - price_min) / price_range
            diff = curr_norm - prior_norm
            length += math.sqrt(diff * diff + self._inv_p_sq)
            prior_norm = curr_norm

        if length <= 0.0:
            return math.nan

        fgdi = 1.0 + (math.log(length) + self._ln2) / self._log_2pm1

        # --- Adaptive speed ---
        denom = 2.0 - fgdi
        if abs(denom) < 1e-10:
            return math.nan

        trail_dim = 1.0 / denom
        alpha = trail_dim / 2.0
        speed = max(1, int(round(self._normal_speed * alpha)))

        # --- SMA of length `speed` ending at current position ---
        n_closes = len(self._closes)
        if speed > n_closes:
            return math.nan

        sma_sum = 0.0
        for k in range(n_closes - speed, n_closes):
            sma_sum += self._closes[k]

        return sma_sum / speed

    def update_scalar(self, sample: Scalar) -> Output:
        return self._line.update_scalar(sample)

    def update_bar(self, sample: Bar) -> Output:
        return self._line.update_bar(sample)

    def update_quote(self, sample: Quote) -> Output:
        return self._line.update_quote(sample)

    def update_trade(self, sample: Trade) -> Output:
        return self._line.update_trade(sample)
