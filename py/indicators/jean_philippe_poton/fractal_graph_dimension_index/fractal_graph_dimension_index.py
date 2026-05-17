"""Fractal Graph Dimension Index indicator."""

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
from .params import FractalGraphDimensionIndexParams


class FractalGraphDimensionIndex(Indicator):
    """Computes the Fractal Graph Dimension Index (FGDI).

    This is Poton's corrected and enhanced version of the Fractal Dimension
    Index (FGDI). It fixes loop boundary and denominator bugs in the original
    and adds standard deviation bands around the estimated dimension.

    The indicator is not primed during the first `period - 1` updates.
    """

    def __init__(self, params: FractalGraphDimensionIndexParams) -> None:
        period = params.period
        if period < 2:
            raise ValueError(
                "invalid fractal graph dimension index parameters: period should be greater than 1")

        bc = params.bar_component if params.bar_component is not None else DEFAULT_BAR_COMPONENT
        qc = params.quote_component if params.quote_component is not None else DEFAULT_QUOTE_COMPONENT
        tc = params.trade_component if params.trade_component is not None else DEFAULT_TRADE_COMPONENT

        self._bar_func = bar_component_value(bc)
        self._quote_func = quote_component_value(qc)
        self._trade_func = trade_component_value(tc)

        mnemonic = f"fgdi({period}{component_triple_mnemonic(bc, qc, tc)})"
        description = f"Fractal graph dimension index {mnemonic}"

        self._mnemonic: str = mnemonic
        self._description: str = description

        self._period: int = period
        self._window: list[float] = [0.0] * period
        self._window_count: int = 0
        self._primed: bool = False

        # Precompute constants.
        n_minus_1 = period - 1
        self._n_minus_1: int = n_minus_1
        self._log_2n1: float = math.log(2.0 * n_minus_1)
        self._ln2: float = math.log(2.0)
        self._inv_n_sq: float = 1.0 / (period * period)

        # Last computed values.
        self._fgdi: float = math.nan
        self._upper: float = math.nan
        self._lower: float = math.nan
        self._stddev: float = math.nan

    def is_primed(self) -> bool:
        return self._primed

    def metadata(self) -> Metadata:
        return build_metadata(
            Identifier.FRACTAL_GRAPH_DIMENSION_INDEX,
            self._mnemonic,
            self._description,
            [
                OutputText(self._mnemonic, self._description),
                OutputText(f"{self._mnemonic} upper", f"{self._description} Upper"),
                OutputText(f"{self._mnemonic} lower", f"{self._description} Lower"),
                OutputText(f"{self._mnemonic} stddev", f"{self._description} Stddev"),
                OutputText(f"{self._mnemonic} band", f"{self._description} Band"),
            ],
        )

    def update(self, sample: float) -> tuple[float, float, float, float]:
        """Update with a scalar value. Returns (fgdi, upper, lower, stddev)."""
        nan = math.nan

        if math.isnan(sample):
            return nan, nan, nan, nan

        period = self._period
        n_minus_1 = self._n_minus_1

        if self._primed:
            # Shift window left by one.
            for i in range(n_minus_1):
                self._window[i] = self._window[i + 1]
            self._window[n_minus_1] = sample
        else:
            self._window[self._window_count] = sample
            self._window_count += 1

            if self._window_count < period:
                return nan, nan, nan, nan

            self._primed = True

        # Find min/max for normalization.
        price_max = self._window[0]
        price_min = self._window[0]
        for k in range(1, period):
            if self._window[k] > price_max:
                price_max = self._window[k]
            if self._window[k] < price_min:
                price_min = self._window[k]

        price_range = price_max - price_min

        if price_range < 1e-10:
            self._fgdi = 1.0
            self._stddev = 0.0
            self._upper = 1.0
            self._lower = 1.0
            return 1.0, 1.0, 1.0, 0.0

        # Normalize and compute path segments.
        prior_norm = (self._window[0] - price_min) / price_range
        length = 0.0
        segments: list[float] = [0.0] * n_minus_1

        for k in range(1, period):
            curr_norm = (self._window[k] - price_min) / price_range
            diff = curr_norm - prior_norm
            seg = math.sqrt(diff * diff + self._inv_n_sq)
            segments[k - 1] = seg
            length += seg
            prior_norm = curr_norm

        # FGDI = 1 + (ln(L) + ln(2)) / ln(2*(N-1))
        fgdi = 1.0 + (math.log(length) + self._ln2) / self._log_2n1

        # Standard deviation of the estimate.
        mean_seg = length / n_minus_1
        sum_sq = 0.0
        for k in range(n_minus_1):
            d = segments[k] - mean_seg
            sum_sq += d * d

        variance = sum_sq / (length * length * self._log_2n1 * self._log_2n1)
        stddev = math.sqrt(variance)

        upper = fgdi + stddev
        lower = fgdi - stddev

        self._fgdi = fgdi
        self._upper = upper
        self._lower = lower
        self._stddev = stddev

        return fgdi, upper, lower, stddev

    def update_scalar(self, sample: Scalar) -> List[Any]:
        fgdi, upper, lower, stddev = self.update(sample.value)
        output: List[Any] = [
            Scalar(time=sample.time, value=fgdi),
            Scalar(time=sample.time, value=upper),
            Scalar(time=sample.time, value=lower),
            Scalar(time=sample.time, value=stddev),
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
