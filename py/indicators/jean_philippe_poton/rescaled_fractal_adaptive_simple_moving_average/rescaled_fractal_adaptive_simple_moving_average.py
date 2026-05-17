"""Rescaled Fractal Adaptive Simple Moving Average (RS-FRASMA) indicator."""

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
from .params import RescaledFractalAdaptiveSimpleMovingAverageParams


class RescaledFractalAdaptiveSimpleMovingAverage(Indicator):
    """Computes the Rescaled Fractal Adaptive Simple Moving Average (RS-FRASMA).

    Uses Rescaled Range (R/S) analysis to estimate the Hurst exponent,
    then adapts the SMA period accordingly.

    The indicator is not primed during the first `period` updates.
    """

    def __init__(self, params: RescaledFractalAdaptiveSimpleMovingAverageParams) -> None:
        period = params.period
        normal_speed = params.normal_speed
        price_scale = params.price_scale

        if period < 4:
            raise ValueError(
                "invalid RS fractal adaptive simple moving average parameters: period should be greater than 3")
        if period & (period - 1) != 0:
            raise ValueError(
                "invalid RS fractal adaptive simple moving average parameters: period must be a power of 2")
        if normal_speed < 1:
            raise ValueError(
                "invalid RS fractal adaptive simple moving average parameters: normal_speed should be greater than 0")

        bc = params.bar_component if params.bar_component is not None else DEFAULT_BAR_COMPONENT
        qc = params.quote_component if params.quote_component is not None else DEFAULT_QUOTE_COMPONENT
        tc = params.trade_component if params.trade_component is not None else DEFAULT_TRADE_COMPONENT

        bar_func = bar_component_value(bc)
        quote_func = quote_component_value(qc)
        trade_func = trade_component_value(tc)

        mnemonic = f"rsfrasma({period},{normal_speed},{price_scale:.1f}{component_triple_mnemonic(bc, qc, tc)})"
        description = f"RS fractal adaptive simple moving average {mnemonic}"

        self._line = LineIndicator(mnemonic, description, bar_func, quote_func, trade_func, self.update)

        self._period: int = period
        self._normal_speed: int = normal_speed
        self._price_scale: float = price_scale
        self._closes: list[float] = []
        self._primed: bool = False

        # Precompute R/S parameters.
        k0 = period // 4
        if k0 < 1:
            self._n_iter = 0
        else:
            self._n_iter = int(math.floor(math.log(k0) / math.log(2))) if k0 >= 2 else 0

        n_iter = self._n_iter
        self._block_sizes: list[int] = [0] * (n_iter + 1)
        self._block_counts: list[int] = [0] * (n_iter + 1)
        for u in range(1, n_iter + 1):
            self._block_sizes[u] = int(2 ** (u + 1))
            self._block_counts[u] = period // self._block_sizes[u]

    def is_primed(self) -> bool:
        return self._primed

    def metadata(self) -> Metadata:
        return build_metadata(
            Identifier.RESCALED_FRACTAL_ADAPTIVE_SIMPLE_MOVING_AVERAGE,
            self._line.mnemonic,
            self._line.description,
            [OutputText(self._line.mnemonic, self._line.description)],
        )

    def update(self, sample: float) -> float:
        if math.isnan(sample):
            return sample

        period = self._period
        price_scale = self._price_scale

        # Accumulate close history.
        self._closes.append(sample)
        n_closes = len(self._closes)

        # Need at least period+1 closes (indices 0..period correspond to MQ4's first valid bar).
        if n_closes <= period:
            return math.nan

        if not self._primed:
            self._primed = True

        # Current position index in closes array.
        pos = n_closes - 1

        # R/S analysis.
        n_iter = self._n_iter
        sumx = 0.0
        sumy = 0.0
        sumx2 = 0.0
        sumxy = 0.0
        valid_scales = 0

        for u in range(1, n_iter + 1):
            block_size = self._block_sizes[u]
            n_blocks_u = self._block_counts[u]
            if n_blocks_u < 1:
                continue

            rs_sum = 0.0
            t = 0
            block_count = 0

            while t <= period - block_size:
                # Block: w[t+j] = price_scale * closes[pos - (t+j)] for j=1..block_size
                mu = 0.0
                for j in range(1, block_size + 1):
                    mu += price_scale * self._closes[pos - (t + j)]
                mu /= block_size

                # Population std.
                sum_sq = 0.0
                for j in range(1, block_size + 1):
                    diff = price_scale * self._closes[pos - (t + j)] - mu
                    sum_sq += diff * diff
                std = math.sqrt(sum_sq / block_size)
                if std <= 0.0:
                    std = 0.1

                # Cumulative deviations and range.
                cum_dev = 0.0
                w_max = 0.0
                w_min = 9999999999.0
                for k in range(1, block_size + 1):
                    cum_dev += price_scale * self._closes[pos - (t + k)] - mu
                    if cum_dev > w_max:
                        w_max = cum_dev
                    if cum_dev < w_min:
                        w_min = cum_dev

                if w_max < 0.0:
                    w_max = 0.0
                if w_min > 0.0:
                    w_min = 0.0

                r_val = w_max - w_min
                rs_sum += r_val / std
                t += block_size
                block_count += 1

            # Average R/S for this scale.
            if block_count > 0:
                rs_avg = rs_sum / block_count
            else:
                rs_avg = 1.0

            if rs_avg <= 0.0:
                rs_avg = 1e-10

            log2_d = math.log(block_size) / math.log(2)
            log2_rs = math.log(rs_avg) / math.log(2)

            sumx += log2_d
            sumy += log2_rs
            sumx2 += log2_d * log2_d
            sumxy += log2_d * log2_rs
            valid_scales += 1

        # Linear regression slope = Hurst exponent.
        if valid_scales < 2:
            h = 0.5
        else:
            h1 = valid_scales * sumxy - sumx * sumy
            h2 = valid_scales * sumx2 - sumx * sumx
            if h2 <= 0.0:
                h2 = 0.1
            h = h1 / h2

        # Guard H.
        if 2.0 * h <= 0.0:
            h = 0.001

        alpha = 1.0 / (2.0 * h)
        spd = max(1, round(self._normal_speed * alpha))

        # Compute SMA with adapted speed.
        sma_start = pos - spd + 1
        if sma_start < 0:
            sma_start = 0
        total = 0.0
        count = pos - sma_start + 1
        for i in range(sma_start, pos + 1):
            total += self._closes[i]

        return total / count

    def update_scalar(self, sample: Scalar) -> Output:
        return self._line.update_scalar(sample)

    def update_bar(self, sample: Bar) -> Output:
        return self._line.update_bar(sample)

    def update_quote(self, sample: Quote) -> Output:
        return self._line.update_quote(sample)

    def update_trade(self, sample: Trade) -> Output:
        return self._line.update_trade(sample)
