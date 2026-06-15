"""Moving Mini-Max (MMM) indicator -- Zurab K. Silagadze.

A nonlinear indicator for technical analysis that emphasizes local maximums and minimums
in a price series with inherent smoothing. The algorithm is borrowed from gamma-ray
spectroscopy peak finding and models price exploration as a quantum particle that can
tunnel through small noise barriers but is stopped by genuine trend reversals.

Algorithm (per Silagadze 2011), recomputed each bar over a sliding window of n prices:
  1. Q-values: unnormalized transition weights from symmetric percentage differences with
     the m nearest neighbours on each side (up mini-max uses +exponent, down mini-max -).
  2. transition probabilities P from the Q-values.
  3. mini-max recurrence u_1 = 1, u_i = (P_{i-1,i}/P_{i,i-1}) u_{i-1}, then normalize.
  4. distinct peaks via local-maxima detection with a minimum separation of max(m, 2) bars.

Reference:
    Silagadze, Z. K. (2011). "Moving Mini-Max -- a new indicator for technical analysis."
    IFTA Journal 11, 46-49. arXiv:0802.0984v2.

Six outputs:
  - up: up mini-max value at the most recent bar (scalar).
  - down: down mini-max value at the most recent bar (scalar).
  - resistances: detected resistance levels (levels with price, offset, strength).
  - supports: detected support levels (levels with price, offset, strength).
  - up_distribution: full up mini-max probability distribution (polyline).
  - down_distribution: full down mini-max probability distribution (polyline).

All outputs are NaN/empty during the priming period (first n prices).
"""

import math
from typing import List, Any

from ...core.indicator import Indicator
from ...core.metadata import Metadata
from ...core.build_metadata import build_metadata, OutputText
from ...core.identifier import Identifier
from ...core.component_triple_mnemonic import component_triple_mnemonic
from ...core.outputs.levels import Level, Levels
from ...core.outputs.polyline import Point, Polyline
from ....entities.bar import Bar
from ....entities.quote import Quote
from ....entities.trade import Trade
from ....entities.scalar import Scalar
from ....entities.bar_component import BarComponent, DEFAULT_BAR_COMPONENT, bar_component_value
from ....entities.quote_component import QuoteComponent, DEFAULT_QUOTE_COMPONENT, quote_component_value
from ....entities.trade_component import TradeComponent, DEFAULT_TRADE_COMPONENT, trade_component_value
from .params import MovingMiniMaxParams


def _calc_q_values(window: List[float], n: int, m: int, negate: bool):
    """Compute Q_{i,i+1} and Q_{i,i-1} for each position i = 0..n-1."""
    sign = -1.0 if negate else 1.0

    q_plus = [0.0] * n
    q_minus = [0.0] * n

    for i in range(n):
        s_i = window[i]
        sum_plus = 0.0
        sum_minus = 0.0

        for k in range(1, m + 1):
            idx_plus = i + k
            s_forward = window[n - 1] if idx_plus >= n else window[idx_plus]

            idx_minus = i - k
            s_backward = window[0] if idx_minus < 0 else window[idx_minus]

            denom_plus = s_forward + s_i
            arg_plus = 0.0 if denom_plus == 0.0 else sign * 2.0 * (s_forward - s_i) / denom_plus

            denom_minus = s_backward + s_i
            arg_minus = 0.0 if denom_minus == 0.0 else sign * 2.0 * (s_backward - s_i) / denom_minus

            sum_plus += math.exp(arg_plus)
            sum_minus += math.exp(arg_minus)

        q_plus[i] = sum_plus
        q_minus[i] = sum_minus

    return q_plus, q_minus


def _calc_p_values(q_plus: List[float], q_minus: List[float], n: int):
    """Compute transition probabilities P_{i,i+1} and P_{i,i-1} from Q-values."""
    p_plus = [0.0] * n
    p_minus = [0.0] * n

    for i in range(n):
        denom = q_plus[i] + q_minus[i]
        if denom == 0.0:
            p_plus[i] = 0.5
            p_minus[i] = 0.5
        else:
            p_plus[i] = q_plus[i] / denom
            p_minus[i] = q_minus[i] / denom

    return p_plus, p_minus


def _calc_minimax(p_plus: List[float], p_minus: List[float], n: int) -> List[float]:
    """Compute the normalized mini-max series from transition probabilities."""
    u = [0.0] * n
    u[0] = 1.0

    for i in range(1, n):
        p_prev_to_i = p_plus[i - 1]
        p_i_to_prev = p_minus[i]
        if p_i_to_prev == 0.0:
            u[i] = u[i - 1] * 1e10
        else:
            u[i] = (p_prev_to_i / p_i_to_prev) * u[i - 1]

    total = sum(u)
    if total == 0.0:
        return [1.0 / n] * n
    return [value / total for value in u]


def _find_peaks(values: List[float], num_peaks: int, min_separation: int):
    """Find distinct local peaks, returned as (value, index) sorted by value descending."""
    n = len(values)

    candidates = []
    for i in range(n):
        if i == 0:
            is_peak = values[i] >= values[i + 1] if n > 1 else True
        elif i == n - 1:
            is_peak = values[i] >= values[i - 1]
        else:
            is_peak = values[i] >= values[i - 1] and values[i] >= values[i + 1]
        if is_peak:
            candidates.append((values[i], i))

    candidates.sort(reverse=True)

    selected = []
    for value, index in candidates:
        if len(selected) >= num_peaks:
            break
        too_close = False
        for _, sel_index in selected:
            if abs(index - sel_index) < min_separation:
                too_close = True
                break
        if not too_close:
            selected.append((value, index))

    return selected


class MovingMiniMax(Indicator):
    """Zurab Silagadze's Moving Mini-Max (MMM) indicator."""

    def __init__(self, params: MovingMiniMaxParams) -> None:
        m = params.m
        n = params.n
        num_extrema = params.num_extrema

        if m < 1:
            raise ValueError("invalid moving mini-max parameters: m must be >= 1")
        if n <= 2 * m:
            raise ValueError("invalid moving mini-max parameters: n must be > 2*m")
        if num_extrema < 1:
            raise ValueError("invalid moving mini-max parameters: num_extrema must be >= 1")

        bc = params.bar_component if params.bar_component is not None else DEFAULT_BAR_COMPONENT
        qc = params.quote_component if params.quote_component is not None else DEFAULT_QUOTE_COMPONENT
        tc = params.trade_component if params.trade_component is not None else DEFAULT_TRADE_COMPONENT

        self._bar_func = bar_component_value(bc)
        self._quote_func = quote_component_value(qc)
        self._trade_func = trade_component_value(tc)

        self._m = m
        self._n = n
        self._num_extrema = num_extrema

        # Ring buffer for the price window.
        self._window = [0.0] * n
        self._buf_pos = 0
        self._count = 0

        self._primed = False

        self._mnemonic = \
            f"mmm({m},{n},{num_extrema}{component_triple_mnemonic(bc, qc, tc)})"

    def is_primed(self) -> bool:
        return self._primed

    def metadata(self) -> Metadata:
        desc = f"Moving mini-max {self._mnemonic}"
        return build_metadata(
            Identifier.MOVING_MINI_MAX,
            self._mnemonic,
            desc,
            [
                OutputText(f"{self._mnemonic} up", f"{desc} up value"),
                OutputText(f"{self._mnemonic} down", f"{desc} down value"),
                OutputText(f"{self._mnemonic} resistances", f"{desc} resistances"),
                OutputText(f"{self._mnemonic} supports", f"{desc} supports"),
                OutputText(f"{self._mnemonic} up dist", f"{desc} up distribution"),
                OutputText(f"{self._mnemonic} down dist", f"{desc} down distribution"),
            ],
        )

    def update(self, sample: float):
        """Update with a scalar value.

        Returns (up, down, resistances, supports, up_dist, down_dist) where resistances/
        supports are lists of (price, offset, strength) tuples and up_dist/down_dist are
        lists; on priming they are (nan, nan, [], [], [], []).
        """
        nan = math.nan
        empty = (nan, nan, [], [], [], [])

        # Store the sample in the ring buffer.
        if self._count < self._n:
            self._window[self._count] = sample
            self._count += 1
        else:
            self._window[self._buf_pos] = sample
            self._buf_pos = (self._buf_pos + 1) % self._n

        if self._count < self._n:
            self._primed = False
            return empty

        self._primed = True

        n = self._n
        m = self._m

        # Reconstruct the window in chronological order (oldest -> newest).
        if self._buf_pos == 0:
            window = list(self._window)
        else:
            window = self._window[self._buf_pos:] + self._window[:self._buf_pos]

        q_up_plus, q_up_minus = _calc_q_values(window, n, m, negate=False)
        q_dn_plus, q_dn_minus = _calc_q_values(window, n, m, negate=True)

        p_up_plus, p_up_minus = _calc_p_values(q_up_plus, q_up_minus, n)
        p_dn_plus, p_dn_minus = _calc_p_values(q_dn_plus, q_dn_minus, n)

        up_dist = _calc_minimax(p_up_plus, p_up_minus, n)
        dn_dist = _calc_minimax(p_dn_plus, p_dn_minus, n)

        min_sep = max(m, 2)

        u_peaks = _find_peaks(up_dist, self._num_extrema, min_sep)
        d_peaks = _find_peaks(dn_dist, self._num_extrema, min_sep)

        resistances = [(window[index], (n - 1) - index, strength) for strength, index in u_peaks]
        supports = [(window[index], (n - 1) - index, strength) for strength, index in d_peaks]

        return up_dist[n - 1], dn_dist[n - 1], resistances, supports, up_dist, dn_dist

    def _wrap(self, time, result) -> List[Any]:
        up, down, resistances, supports, up_dist, dn_dist = result
        return [
            Scalar(time=time, value=up),
            Scalar(time=time, value=down),
            Levels(time, [Level(price, offset, strength) for price, offset, strength in resistances]),
            Levels(time, [Level(price, offset, strength) for price, offset, strength in supports]),
            Polyline(time, [Point(i, value) for i, value in enumerate(up_dist)]),
            Polyline(time, [Point(i, value) for i, value in enumerate(dn_dist)]),
        ]

    def update_scalar(self, sample: Scalar) -> List[Any]:
        return self._wrap(sample.time, self.update(sample.value))

    def update_bar(self, sample: Bar) -> List[Any]:
        return self._wrap(sample.time, self.update(self._bar_func(sample)))

    def update_quote(self, sample: Quote) -> List[Any]:
        return self._wrap(sample.time, self.update(self._quote_func(sample)))

    def update_trade(self, sample: Trade) -> List[Any]:
        return self._wrap(sample.time, self.update(self._trade_func(sample)))
