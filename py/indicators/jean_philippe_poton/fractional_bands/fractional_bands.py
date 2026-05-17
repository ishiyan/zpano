"""Fractional Bands indicator."""

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
from .params import FractionalBandsParams


class FractionalBands(Indicator):
    """Computes the Fractional Bands indicator.

    Fractal-adaptive moving average with FBM-scaled volatility bands.
    Uses fractional Brownian motion power law: band_width = 2 * deviation^(2*H)
    where H is the Hurst exponent derived from the Fractal Dimension Index.

    The indicator is not primed during the first `period` updates.
    """

    def __init__(self, params: FractionalBandsParams) -> None:
        period = params.period
        price_scale = params.price_scale

        if period < 2:
            raise ValueError(
                "invalid fractional bands parameters: period should be greater than 1")
        if price_scale <= 0.0:
            raise ValueError(
                "invalid fractional bands parameters: price_scale should be greater than 0")

        bc = params.bar_component if params.bar_component is not None else DEFAULT_BAR_COMPONENT
        qc = params.quote_component if params.quote_component is not None else DEFAULT_QUOTE_COMPONENT
        tc = params.trade_component if params.trade_component is not None else DEFAULT_TRADE_COMPONENT

        self._bar_func = bar_component_value(bc)
        self._quote_func = quote_component_value(qc)
        self._trade_func = trade_component_value(tc)

        mnemonic = f"fctban({period},{price_scale}{component_triple_mnemonic(bc, qc, tc)})"
        description = f"Fractional bands {mnemonic}"

        self._mnemonic: str = mnemonic
        self._description: str = description

        self._period: int = period
        self._price_scale: float = price_scale

        # FGDI window: period+1 elements.
        self._window_size: int = period + 1
        self._window: list[float] = [0.0] * (period + 1)
        self._window_count: int = 0
        self._closes: list[float] = []
        self._primed: bool = False

        # Precompute constants for FGDI.
        self._log_denom: float = math.log(2.0 * (period - 1))
        self._ln2: float = math.log(2.0)
        self._inv_period_sq: float = 1.0 / (period * period)

        # Last computed values.
        self._frasma2: float = math.nan
        self._upper_band: float = math.nan
        self._lower_band: float = math.nan

    def is_primed(self) -> bool:
        return self._primed

    def metadata(self) -> Metadata:
        return build_metadata(
            Identifier.FRACTIONAL_BANDS,
            self._mnemonic,
            self._description,
            [
                OutputText(self._mnemonic, self._description),
                OutputText(f"{self._mnemonic} upper", f"{self._description} Upper Band"),
                OutputText(f"{self._mnemonic} lower", f"{self._description} Lower Band"),
                OutputText(f"{self._mnemonic} band", f"{self._description} Band"),
            ],
        )

    def update(self, sample: float) -> tuple[float, float, float]:
        """Update with a scalar value. Returns (frasma2, upper_band, lower_band)."""
        nan = math.nan

        if math.isnan(sample):
            return nan, nan, nan

        period = self._period
        window_size = self._window_size
        p = self._price_scale

        # Accumulate close history.
        self._closes.append(sample)

        # Fill the FGDI window (period+1 elements).
        if self._window_count < window_size:
            self._window[self._window_count] = sample
            self._window_count += 1

            if self._window_count < window_size:
                return nan, nan, nan

            self._primed = True
        else:
            # Shift window left by one.
            for i in range(window_size - 1):
                self._window[i] = self._window[i + 1]
            self._window[window_size - 1] = sample

        # FGDI computation over period+1 points.
        price_max = self._window[0]
        price_min = self._window[0]
        for k in range(1, window_size):
            if self._window[k] > price_max:
                price_max = self._window[k]
            if self._window[k] < price_min:
                price_min = self._window[k]

        price_range = price_max - price_min

        if price_range < 1e-10:
            fgdi = 1.0
        else:
            inv_range = 1.0 / price_range
            prev_norm = (self._window[0] - price_min) * inv_range
            length = 0.0
            for i in range(1, period):  # period-1 segments
                cur_norm = (self._window[i] - price_min) * inv_range
                diff = cur_norm - prev_norm
                length += math.sqrt(diff * diff + self._inv_period_sq)
                prev_norm = cur_norm

            if length > 0.0:
                fgdi = 1.0 + (math.log(length) + self._ln2) / self._log_denom
            else:
                fgdi = 1.0

        # Hurst exponent and adaptive speed.
        hurst = 2.0 - fgdi
        if hurst < 0.01:
            hurst = 0.01
        trail_dim = 1.0 / hurst
        beta = trail_dim / 2.0
        speed = max(int(round(period * beta)), 1)

        # FRASMA2: SMA of close over 'speed' bars ending at current position.
        n_closes = len(self._closes)
        if speed > n_closes:
            return nan, nan, nan

        sma_sum = 0.0
        for k in range(n_closes - speed, n_closes):
            sma_sum += self._closes[k]
        frasma2_val = sma_sum / speed

        # Deviation in scaled space over last *period* closes.
        dev_start = n_closes - period
        frasma2_scaled = p * frasma2_val
        sq_sum = 0.0
        for k in range(dev_start, n_closes):
            res = p * self._closes[k] - frasma2_scaled
            sq_sum += res * res
        deviation = math.sqrt(sq_sum / period)

        # FBM band offset: 2 * sigma^(2H).
        two_h = 2.0 * hurst
        band_offset = 2.0 * math.pow(deviation, two_h)
        upper_band = (frasma2_scaled + band_offset) / p
        lower_band = (frasma2_scaled - band_offset) / p

        self._frasma2 = frasma2_val
        self._upper_band = upper_band
        self._lower_band = lower_band

        return frasma2_val, upper_band, lower_band

    def update_scalar(self, sample: Scalar) -> List[Any]:
        frasma2, upper, lower = self.update(sample.value)
        output: List[Any] = [
            Scalar(time=sample.time, value=frasma2),
            Scalar(time=sample.time, value=upper),
            Scalar(time=sample.time, value=lower),
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
