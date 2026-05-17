"""Hurst Difference indicator."""

import math
from typing import List, Any

from ...core.indicator import Indicator
from ...core.metadata import Metadata
from ...core.build_metadata import build_metadata, OutputText
from ...core.identifier import Identifier
from ...core.component_triple_mnemonic import component_triple_mnemonic
from ....entities.bar import Bar
from ....entities.quote import Quote
from ....entities.trade import Trade
from ....entities.scalar import Scalar
from ....entities.bar_component import BarComponent, DEFAULT_BAR_COMPONENT, bar_component_value
from ....entities.quote_component import QuoteComponent, DEFAULT_QUOTE_COMPONENT, quote_component_value
from ....entities.trade_component import TradeComponent, DEFAULT_TRADE_COMPONENT, trade_component_value
from .params import HurstDifferenceParams


class HurstDifference(Indicator):
    """Computes the Hurst Difference (first difference of the corrected FGDI).

    Positive values indicate rising volatility (potential trade entry);
    negative values indicate declining volatility.

    The FGDI is computed using the corrected FGDI formula with (period-1)
    segments and denominator ln(2*(period-1)).

    The indicator is not primed during the first `period` updates.
    The hurst_diff output requires one additional update beyond FGDI priming.
    """

    def __init__(self, params: HurstDifferenceParams) -> None:
        period = params.period
        if period < 2:
            raise ValueError(
                "invalid hurst difference parameters: period should be greater than 1")

        bc = params.bar_component if params.bar_component is not None else DEFAULT_BAR_COMPONENT
        qc = params.quote_component if params.quote_component is not None else DEFAULT_QUOTE_COMPONENT
        tc = params.trade_component if params.trade_component is not None else DEFAULT_TRADE_COMPONENT

        self._bar_func = bar_component_value(bc)
        self._quote_func = quote_component_value(qc)
        self._trade_func = trade_component_value(tc)

        mnemonic = f"hurdif({period}{component_triple_mnemonic(bc, qc, tc)})"
        description = f"Hurst difference {mnemonic}"

        self._mnemonic: str = mnemonic
        self._description: str = description

        self._period: int = period
        self._window: list[float] = [0.0] * (period + 1)
        self._window_count: int = 0
        self._primed: bool = False

        # Precompute constants.
        n_minus_1 = period - 1
        self._n_minus_1: int = n_minus_1
        self._log_2pm1: float = math.log(2.0 * n_minus_1)
        self._ln2: float = math.log(2.0)
        self._inv_n_sq: float = 1.0 / (period * period)

        # Previous FGDI value for differencing.
        self._prev_fgdi: float = math.nan

    def is_primed(self) -> bool:
        return self._primed

    def metadata(self) -> Metadata:
        return build_metadata(
            Identifier.HURST_DIFFERENCE,
            self._mnemonic,
            self._description,
            [
                OutputText(self._mnemonic, self._description),
                OutputText(f"{self._mnemonic} fgdi", f"{self._description} FGDI"),
            ],
        )

    def update(self, sample: float) -> tuple[float, float]:
        """Update with a scalar value. Returns (hurst_diff, fgdi)."""
        nan = math.nan

        if math.isnan(sample):
            return nan, nan

        period = self._period
        n_minus_1 = self._n_minus_1

        if self._primed:
            # Shift window left by one.
            for i in range(period):
                self._window[i] = self._window[i + 1]
            self._window[period] = sample
        else:
            self._window[self._window_count] = sample
            self._window_count += 1

            if self._window_count <= period:
                return nan, nan

            self._primed = True

        # Use the last `period` elements of the window (indices 1..period inclusive).
        # Find min/max for normalization.
        price_max = self._window[1]
        price_min = self._window[1]
        for k in range(2, period + 1):
            if self._window[k] > price_max:
                price_max = self._window[k]
            if self._window[k] < price_min:
                price_min = self._window[k]

        price_range = price_max - price_min

        if price_range <= 0.0:
            fgdi = 0.0
        else:
            # Normalize and compute path length.
            prior_norm = (self._window[1] - price_min) / price_range
            length = 0.0

            for k in range(2, period + 1):
                curr_norm = (self._window[k] - price_min) / price_range
                diff = curr_norm - prior_norm
                length += math.sqrt(diff * diff + self._inv_n_sq)
                prior_norm = curr_norm

            if length > 0.0:
                fgdi = 1.0 + (math.log(length) + self._ln2) / self._log_2pm1
            else:
                fgdi = 0.0

        # First difference.
        if math.isnan(self._prev_fgdi):
            hurst_diff = nan
        else:
            hurst_diff = fgdi - self._prev_fgdi

        self._prev_fgdi = fgdi

        return hurst_diff, fgdi

    def update_scalar(self, sample: Scalar) -> List[Any]:
        hurst_diff, fgdi = self.update(sample.value)
        return [
            Scalar(time=sample.time, value=hurst_diff),
            Scalar(time=sample.time, value=fgdi),
        ]

    def update_bar(self, sample: Bar) -> List[Any]:
        v = self._bar_func(sample)
        return self.update_scalar(Scalar(time=sample.time, value=v))

    def update_quote(self, sample: Quote) -> List[Any]:
        v = self._quote_func(sample)
        return self.update_scalar(Scalar(time=sample.time, value=v))

    def update_trade(self, sample: Trade) -> List[Any]:
        v = self._trade_func(sample)
        return self.update_scalar(Scalar(time=sample.time, value=v))
