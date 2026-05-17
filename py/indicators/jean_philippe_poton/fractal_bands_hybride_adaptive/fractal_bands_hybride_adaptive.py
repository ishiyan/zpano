"""Fractal Bands Hybride Adaptive indicator."""

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
from .params import FractalBandsHybrideAdaptiveParams


class FractalBandsHybrideAdaptive(Indicator):
    """Computes the Fractal Bands Hybride Adaptive indicator.

    Hybrid variant of Fractal Bands that replaces fixed normal_speed with
    Ehlers' CyclePeriod indicator output multiplied by a Nyquist factor,
    making the FRASMA2 doubly adaptive to both fractal dimension and
    dominant market cycle.

    The indicator is not primed during the first `period` updates.
    """

    def __init__(self, params: FractalBandsHybrideAdaptiveParams) -> None:
        period = params.period
        normal_speed_fallback = params.normal_speed_fallback
        alpha = params.alpha
        nyquist = params.nyquist
        alpha_hp = params.alpha_hp

        if period < 2:
            raise ValueError(
                "invalid fractal bands hybride adaptive parameters: period should be greater than 1")
        if normal_speed_fallback < 1:
            raise ValueError(
                "invalid fractal bands hybride adaptive parameters: normal_speed_fallback should be greater than 0")
        if alpha <= 0.0:
            raise ValueError(
                "invalid fractal bands hybride adaptive parameters: alpha should be greater than 0")
        if nyquist <= 0.0:
            raise ValueError(
                "invalid fractal bands hybride adaptive parameters: nyquist should be greater than 0")
        if alpha_hp <= 0.0 or alpha_hp >= 1.0:
            raise ValueError(
                "invalid fractal bands hybride adaptive parameters: alpha_hp should be between 0 and 1")

        bc = params.bar_component if params.bar_component is not None else DEFAULT_BAR_COMPONENT
        qc = params.quote_component if params.quote_component is not None else DEFAULT_QUOTE_COMPONENT
        tc = params.trade_component if params.trade_component is not None else DEFAULT_TRADE_COMPONENT

        self._bar_func = bar_component_value(bc)
        self._quote_func = quote_component_value(qc)
        self._trade_func = trade_component_value(tc)

        mnemonic = f"fbanha({period},{normal_speed_fallback},{alpha},{nyquist},{alpha_hp}" \
                   f"{component_triple_mnemonic(bc, qc, tc)})"
        description = f"Fractal bands hybride adaptive {mnemonic}"

        self._mnemonic: str = mnemonic
        self._description: str = description

        self._period: int = period
        self._normal_speed_fallback: int = normal_speed_fallback
        self._alpha: float = alpha
        self._nyquist: float = nyquist
        self._alpha_hp: float = alpha_hp

        # FGDI window: period+1 elements (matching batch: prices[pos-period:pos+1])
        self._window_size: int = period + 1
        self._window: list[float] = [0.0] * (period + 1)
        self._window_count: int = 0
        self._closes: list[float] = []
        self._primed: bool = False

        # Precompute constants for FGDI.
        self._log_denom: float = math.log(2.0 * (period - 1))
        self._ln2: float = math.log(2.0)
        self._inv_period_sq: float = 1.0 / (period * period)

        # Ehlers CyclePeriod state - full arrays matching batch indexing.
        # We store enough history to compute at each step.
        self._smooth_buf: list[float] = []  # all smooth values
        self._cycle_buf: list[float] = []   # all cycle values
        self._q1_buf: list[float] = []      # all q1 values
        self._i1_buf: list[float] = []      # all i1 values
        self._dp_buf: list[float] = []      # all delta_phase values
        self._inst_period_buf: list[float] = []  # all inst_period values

        # Last computed values.
        self._frasma2: float = math.nan
        self._upper_band: float = math.nan
        self._lower_band: float = math.nan

    def is_primed(self) -> bool:
        return self._primed

    def metadata(self) -> Metadata:
        return build_metadata(
            Identifier.FRACTAL_BANDS_HYBRIDE_ADAPTIVE,
            self._mnemonic,
            self._description,
            [
                OutputText(self._mnemonic, self._description),
                OutputText(f"{self._mnemonic} upper", f"{self._description} Upper Band"),
                OutputText(f"{self._mnemonic} lower", f"{self._description} Lower Band"),
                OutputText(f"{self._mnemonic} band", f"{self._description} Band"),
            ],
        )

    def _get_cycle_period(self) -> float:
        """Get the current CyclePeriod estimate using the batch algorithm applied incrementally."""
        # This mirrors _ehlers_cycle_period from the batch version exactly.
        # t is the 0-based index of the current sample in self._closes.
        t = len(self._closes) - 1
        prices = self._closes

        # Extend buffers to index t
        while len(self._smooth_buf) <= t:
            self._smooth_buf.append(0.0)
        while len(self._cycle_buf) <= t:
            self._cycle_buf.append(0.0)
        while len(self._q1_buf) <= t:
            self._q1_buf.append(0.0)
        while len(self._i1_buf) <= t:
            self._i1_buf.append(0.0)
        while len(self._dp_buf) <= t:
            self._dp_buf.append(0.0)
        while len(self._inst_period_buf) <= t:
            self._inst_period_buf.append(6.0)

        if t < 6:
            return math.nan

        # 4-bar weighted smoother
        self._smooth_buf[t] = (prices[t] + 2.0 * prices[t - 1] + \
                               2.0 * prices[t - 2] + prices[t - 3]) / 6.0

        # High-pass filter
        alpha_hp = self._alpha_hp
        hp_coeff = (1.0 - 0.5 * alpha_hp) ** 2
        self._cycle_buf[t] = (
            hp_coeff * (self._smooth_buf[t] - 2.0 * self._smooth_buf[t - 1] + self._smooth_buf[t - 2])
            + 2.0 * (1.0 - alpha_hp) * self._cycle_buf[t - 1]
            - (1.0 - alpha_hp) ** 2 * self._cycle_buf[t - 2]
        )

        # Quadrature component
        self._q1_buf[t] = (
            0.0962 * self._cycle_buf[t]
            + 0.5769 * self._cycle_buf[t - 2]
            - 0.5769 * self._cycle_buf[t - 4]
            - 0.0962 * self._cycle_buf[t - 6]
        ) * (0.5 + 0.08 * self._inst_period_buf[t - 1])

        # In-phase component
        self._i1_buf[t] = self._cycle_buf[t - 3]

        # Smooth I and Q with EMA
        if t > 6:
            self._i1_buf[t] = 0.15 * self._i1_buf[t] + 0.85 * self._i1_buf[t - 1]
            self._q1_buf[t] = 0.15 * self._q1_buf[t] + 0.85 * self._q1_buf[t - 1]

        # Compute delta phase
        if abs(self._i1_buf[t]) > 1e-10:
            dp = math.atan(self._q1_buf[t] / self._i1_buf[t])
        else:
            dp = self._dp_buf[t - 1]

        # Clamp delta phase
        if dp < 0.1:
            dp = 0.1
        if dp > 1.1:
            dp = 1.1
        self._dp_buf[t] = dp

        # Median delta phase over 5 bars
        if t >= 10:
            window5 = [self._dp_buf[t - 4], self._dp_buf[t - 3], self._dp_buf[t - 2],
                       self._dp_buf[t - 1], self._dp_buf[t]]
            window5.sort()
            median_dp = window5[2]
        else:
            median_dp = dp

        # Instantaneous period
        if abs(median_dp) > 1e-10:
            dc = 6.2832 / median_dp + 0.5
        else:
            dc = self._inst_period_buf[t - 1]

        # Clamp and smooth
        if dc < 6.0:
            dc = 6.0
        if dc > 50.0:
            dc = 50.0
        self._inst_period_buf[t] = 0.33 * dc + 0.67 * self._inst_period_buf[t - 1]

        return self._inst_period_buf[t]

    def update(self, sample: float) -> tuple[float, float, float]:
        """Update with a scalar value. Returns (frasma2, upper_band, lower_band)."""
        nan = math.nan

        if math.isnan(sample):
            return nan, nan, nan

        period = self._period
        window_size = self._window_size

        # Accumulate close history.
        self._closes.append(sample)

        # Update Ehlers CyclePeriod.
        cp = self._get_cycle_period()

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

        # FGDI computation over period+1 points (period segments).
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
            length = 0.0
            for i in range(1, window_size):
                norm_cur = (self._window[i] - price_min) / price_range
                norm_prev = (self._window[i - 1] - price_min) / price_range
                diff = norm_cur - norm_prev
                length += math.sqrt(diff * diff + self._inv_period_sq)

            fgdi = 1.0 + (math.log(length) + self._ln2) / self._log_denom

        # Hurst exponent.
        hurst = 2.0 - fgdi
        if hurst < 0.01:
            hurst = 0.01
        trail_dim = 1.0 / hurst
        beta = trail_dim / 2.0

        # Adaptive normal_speed from CyclePeriod.
        if math.isnan(cp) or cp < 1.0:
            ns = float(self._normal_speed_fallback)
        else:
            ns = cp * self._nyquist

        speed = max(int(round(ns * beta)), 1)

        # FRASMA2: SMA of close over 'speed' bars ending at current position.
        n_closes = len(self._closes)
        if speed > n_closes:
            return nan, nan, nan

        sma_window = self._closes[n_closes - speed:]
        frasma2_val = sum(sma_window) / speed

        # Deviation over the last `period` closes.
        sq_sum = 0.0
        dev_start = max(n_closes - period, 0)
        dev_count = n_closes - dev_start
        for k in range(dev_start, n_closes):
            res = self._closes[k] - frasma2_val
            sq_sum += res * res
        deviation = 2.0 * math.sqrt(sq_sum / period)

        # Fractal bands.
        band_mult = deviation * (self._alpha ** hurst)
        upper_band = frasma2_val + band_mult
        lower_band = frasma2_val - band_mult

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
