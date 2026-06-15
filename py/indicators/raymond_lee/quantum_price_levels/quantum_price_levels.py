"""Quantum Price Levels (QPL) indicator -- Raymond S. T. Lee.

Computes discrete support/resistance price levels from a quantum-finance analogy:
the market is modelled as a quantum anharmonic oscillator, and the discrete energy
eigenvalues of the system map to price levels above and below the current price.

Algorithm (per Lee 2021), recomputed each bar over a sliding window of returns:
  1. returns r(t) = price(t-1)/price(t)   (inverse ratio, Lee's convention)
  2. population mu, sigma of returns
  3. histogram with bin width dr = 3*sigma/(num_bins/2), centred at r = 1
  4. find the peak (ground state) bin
  5. anharmonic coefficient lambda via the finite-difference method
  6. K0(n) constants (Dasgupta et al. 2007)
  7. solve depressed cubics E(n)^3 - (2n+1)^2 E(n) - lambda (2n+1)^3 K0(n)^3 = 0 (Cardano)
  8. NQPR(n) = 1 + scale_factor * sigma * (QFEL(n)/QFEL(0))
  9. resistances(n) = price * NQPR(n), supports(n) = price / NQPR(n)

Reference:
    Lee, R. S. T. (2021). "Quantum Finance Forecast System with Quantum Anharmonic
    Oscillator Model for Quantum Price Level Modeling." IAJER, 4(02), 1-21.

Five outputs:
  - lambda: anharmonic coefficient (scalar).
  - return_std_dev: population std dev of returns (scalar).
  - normalized_multipliers: NQPR multipliers (levels).
  - resistances: resistance levels above price (levels).
  - supports: support levels below price (levels).

All outputs are NaN/empty during the priming period (first lookback+1 prices) or when
the window is degenerate (zero variance, peak at a histogram edge, etc.).
"""

import math
from typing import List, Any

from ...core.indicator import Indicator
from ...core.metadata import Metadata
from ...core.build_metadata import build_metadata, OutputText
from ...core.identifier import Identifier
from ...core.component_triple_mnemonic import component_triple_mnemonic
from ...core.outputs.levels import Level, Levels
from ....entities.bar import Bar
from ....entities.quote import Quote
from ....entities.trade import Trade
from ....entities.scalar import Scalar
from ....entities.bar_component import BarComponent, DEFAULT_BAR_COMPONENT, bar_component_value
from ....entities.quote_component import QuoteComponent, DEFAULT_QUOTE_COMPONENT, quote_component_value
from ....entities.trade_component import TradeComponent, DEFAULT_TRADE_COMPONENT, trade_component_value
from .params import QuantumPriceLevelsParams


def _cbrt(x: float) -> float:
    """Signed real cube root via pow (matches the reference implementation)."""
    if x >= 0.0:
        return x ** (1.0 / 3.0)
    return -((-x) ** (1.0 / 3.0))


def _compute_k0(n: int) -> float:
    """K0 constant for energy level n (Dasgupta et al. 2007)."""
    numerator = 1.1924 + 33.2383 * n + 56.2169 * n * n
    denominator = 1.0 + 43.6106 * n
    return (numerator / denominator) ** (1.0 / 3.0)


class QuantumPriceLevels(Indicator):
    """Raymond Lee's Quantum Price Levels (QPL) indicator."""

    def __init__(self, params: QuantumPriceLevelsParams) -> None:
        lookback = params.lookback
        num_levels = params.num_levels
        num_bins = params.num_bins
        scale_factor = params.scale_factor

        if lookback < 2:
            raise ValueError("invalid quantum price levels parameters: lookback must be >= 2")
        if num_levels < 1:
            raise ValueError("invalid quantum price levels parameters: num_levels must be >= 1")
        if num_bins < 2:
            raise ValueError("invalid quantum price levels parameters: num_bins must be >= 2")
        if scale_factor <= 0.0:
            raise ValueError("invalid quantum price levels parameters: scale_factor must be > 0")

        bc = params.bar_component if params.bar_component is not None else DEFAULT_BAR_COMPONENT
        qc = params.quote_component if params.quote_component is not None else DEFAULT_QUOTE_COMPONENT
        tc = params.trade_component if params.trade_component is not None else DEFAULT_TRADE_COMPONENT

        self._bar_func = bar_component_value(bc)
        self._quote_func = quote_component_value(qc)
        self._trade_func = trade_component_value(tc)

        self._lookback = lookback
        self._num_levels = num_levels
        self._num_bins = num_bins
        self._scale_factor = scale_factor

        # Pre-compute K0 constants (they never change).
        self._k = [_compute_k0(n) for n in range(num_levels)]

        # Ring buffer for returns.
        self._returns = [0.0] * lookback
        self._buf_pos = 0
        self._count = 0
        self._prev_price = None

        self._primed = False

        self._mnemonic = \
            f"qpl({lookback},{num_levels},{num_bins},{scale_factor:g}" \
            f"{component_triple_mnemonic(bc, qc, tc)})"

    def is_primed(self) -> bool:
        return self._primed

    def metadata(self) -> Metadata:
        desc = f"Quantum price levels {self._mnemonic}"
        return build_metadata(
            Identifier.QUANTUM_PRICE_LEVELS,
            self._mnemonic,
            desc,
            [
                OutputText(f"{self._mnemonic} lambda", f"{desc} anharmonic coefficient"),
                OutputText(f"{self._mnemonic} stddev", f"{desc} return standard deviation"),
                OutputText(f"{self._mnemonic} nqpr", f"{desc} normalized multipliers"),
                OutputText(f"{self._mnemonic} resistances", f"{desc} resistance levels"),
                OutputText(f"{self._mnemonic} supports", f"{desc} support levels"),
            ],
        )

    def update(self, sample: float):
        """Update with a scalar value.

        Returns (lambda, sigma, nqpr, resistances, supports) where nqpr/resistances/
        supports are lists; on priming/degenerate they are (nan, nan, [], [], []).
        """
        nan = math.nan
        empty = (nan, nan, [], [], [])

        # First price: just store it; no return yet.
        if self._prev_price is None:
            self._prev_price = sample
            self._primed = False
            return empty

        # Inverse return ratio (Lee's convention).
        if sample > 0.0:
            new_return = self._prev_price / sample
        else:
            new_return = 1.0
        self._prev_price = sample

        # Store in ring buffer.
        if self._count < self._lookback:
            self._returns[self._count] = new_return
            self._count += 1
        else:
            self._returns[self._buf_pos] = new_return
            self._buf_pos = (self._buf_pos + 1) % self._lookback

        if self._count < self._lookback:
            self._primed = False
            return empty

        self._primed = True

        lookback = self._lookback
        num_bins = self._num_bins
        num_levels = self._num_levels
        scale_factor = self._scale_factor

        # Statistics (population mu, sigma).
        sum_r = 0.0
        for i in range(lookback):
            sum_r += self._returns[i]
        mu = sum_r / lookback

        sum_var = 0.0
        for i in range(lookback):
            diff = self._returns[i] - mu
            sum_var += diff * diff
        sigma = math.sqrt(sum_var / lookback)

        if sigma == 0.0:
            return empty

        # Histogram centred at r = 1.
        half_bins = num_bins // 2
        dr = 3.0 * sigma / half_bins
        left_boundary = 1.0 - half_bins * dr

        q = [0] * num_bins
        total_count = 0
        for i in range(lookback):
            r = self._returns[i]
            bin_index = int((r - left_boundary) / dr)
            if 0 <= bin_index < num_bins:
                q[bin_index] += 1
                total_count += 1

        if total_count == 0:
            return empty

        # Ground state (peak bin).
        max_q = 0.0
        max_qno = 0
        for k in range(num_bins):
            nq = q[k] / total_count
            if nq > max_q:
                max_q = nq
                max_qno = k

        if max_qno == 0 or max_qno == num_bins - 1:
            return empty

        # lambda via FDM.
        phi_plus1 = q[max_qno + 1] / total_count
        phi_minus1 = q[max_qno - 1] / total_count

        r_peak = left_boundary + max_qno * dr
        r0 = r_peak - dr / 2.0
        r_plus1 = r0 + dr
        r_minus1 = r0 - dr

        l_up = (r_minus1 ** 2) * phi_minus1 - (r_plus1 ** 2) * phi_plus1
        l_dw = (r_plus1 ** 4) * phi_plus1 - (r_minus1 ** 4) * phi_minus1

        if l_dw == 0.0:
            return empty

        lambda_ = abs(l_up / l_dw)

        # Energy levels via Cardano.
        qfel = [0.0] * num_levels
        for n in range(num_levels):
            two_n_plus_1 = 2 * n + 1
            p = -(two_n_plus_1 ** 2)
            q_coef = -lambda_ * (two_n_plus_1 ** 3) * (self._k[n] ** 3)
            discriminant = (q_coef * q_coef / 4.0) + (p * p * p / 27.0)
            if discriminant < 0.0:
                return empty
            sqrt_d = math.sqrt(discriminant)
            u = _cbrt(-q_coef / 2.0 + sqrt_d)
            v = _cbrt(-q_coef / 2.0 - sqrt_d)
            qfel[n] = u + v

        if qfel[0] == 0.0:
            return empty

        # NQPR and projection from the current price.
        nqpr = [0.0] * num_levels
        resistances = [0.0] * num_levels
        supports = [0.0] * num_levels
        for n in range(num_levels):
            qpr = qfel[n] / qfel[0]
            nqpr[n] = 1.0 + scale_factor * sigma * qpr
            resistances[n] = sample * nqpr[n]
            supports[n] = sample / nqpr[n]

        return lambda_, sigma, nqpr, resistances, supports

    def _wrap(self, time, result) -> List[Any]:
        lambda_, sigma, nqpr, resistances, supports = result
        return [
            Scalar(time=time, value=lambda_),
            Scalar(time=time, value=sigma),
            Levels(time, [Level(v) for v in nqpr]),
            Levels(time, [Level(v) for v in resistances]),
            Levels(time, [Level(v) for v in supports]),
        ]

    def update_scalar(self, sample: Scalar) -> List[Any]:
        return self._wrap(sample.time, self.update(sample.value))

    def update_bar(self, sample: Bar) -> List[Any]:
        return self._wrap(sample.time, self.update(self._bar_func(sample)))

    def update_quote(self, sample: Quote) -> List[Any]:
        return self._wrap(sample.time, self.update(self._quote_func(sample)))

    def update_trade(self, sample: Trade) -> List[Any]:
        return self._wrap(sample.time, self.update(self._trade_func(sample)))
