"""Candlestick pattern recognition engine with streaming (incremental) support.

Usage:
    from py.candlestick_patterns import CandlestickPatterns

    cp = CandlestickPatterns()
    for bar in bars:
        cp.update(bar.open, bar.high, bar.low, bar.close)
        result = cp.abandoned_baby()  # -100, 0, or +100

Each pattern method inspects the most recent N bars (stored in a ring buffer)
and the incrementally maintained running totals for each criterion, giving
O(1) per bar after the warmup period.
"""
from __future__ import annotations

from collections import deque

from .core.criterion import Criterion
from .core.defaults import (
    DEFAULT_LONG_BODY, DEFAULT_VERY_LONG_BODY, DEFAULT_SHORT_BODY, DEFAULT_DOJI_BODY,
    DEFAULT_LONG_SHADOW, DEFAULT_VERY_LONG_SHADOW, DEFAULT_SHORT_SHADOW, DEFAULT_VERY_SHORT_SHADOW,
    DEFAULT_NEAR, DEFAULT_FAR, DEFAULT_EQUAL,
)
from .core.pattern_identifier import PatternIdentifier
from .core.primitives import (
    is_white, is_black, real_body, white_real_body, black_real_body,
    upper_shadow, lower_shadow, white_upper_shadow, black_upper_shadow,
    white_lower_shadow, black_lower_shadow,
    is_real_body_gap_up, is_real_body_gap_down,
    is_high_low_gap_up, is_high_low_gap_down,
    is_real_body_encloses_real_body, is_real_body_encloses_open, is_real_body_encloses_close,
    is_high_exceeds_close, is_opens_within,
)


# Minimum history size: 5-candle patterns + 10 default criterion period + 5 margin.
_MIN_HISTORY = 20


class _CriterionState:
    """Maintains a running total for a single Criterion over a sliding window.

    The window covers the `average_period` bars ending at a configurable offset
    from the current bar. Each pattern decides which offset to use when querying.
    """

    __slots__ = ('criterion', '_ring', '_total')

    def __init__(self, criterion: Criterion, max_shift: int = 5) -> None:
        self.criterion = criterion
        # Ring must hold period + max_shift entries so total_at works for all shifts.
        ring_size = criterion.average_period + max_shift if criterion.average_period > 0 else 0
        self._ring: deque[float] = deque(maxlen=ring_size) if ring_size > 0 else None
        self._total = 0.0

    def push(self, o: float, h: float, l: float, c: float) -> None:
        """Add the contribution of a new bar and evict the oldest if the window is full."""
        if self._ring is None:
            return
        val = self.criterion.candle_contribution(o, h, l, c)
        if len(self._ring) == self._ring.maxlen:
            self._total -= self._ring[0]
        self._ring.append(val)
        self._total += val

    def total_at(self, history: deque, shift: int) -> float:
        """Compute the running total for bars ending at `shift` bars before the current bar.

        For streaming we maintain a single running total for the most recent window.
        To support different shifts per pattern, we recompute from the ring buffer.
        Since the ring is at most 10 elements, this is still O(period) but with tiny constant.
        """
        if self._ring is None or self.criterion.average_period <= 0:
            return 0.0
        period = self.criterion.average_period
        ring = self._ring
        n = len(ring)
        end = n - shift
        start = end - period
        if start < 0 or end <= 0:
            return 0.0
        total = 0.0
        for i in range(start, end):
            total += ring[i]
        return total

    def avg(self, history: deque, shift: int, o: float, h: float, l: float, c: float) -> float:
        """Compute the average criterion value.

        Args:
            history: The bar history ring (unused when period > 0, used only for length check).
            shift: How many bars back from the end of history the window should end.
            o, h, l, c: OHLC of the reference candle (used when average_period == 0).
        """
        return self.criterion.average_value_from_total(
            self.total_at(history, shift), o, h, l, c
        )


class CandlestickPatterns:
    """The candlestick pattern recognition engine.

    Provides streaming bar-by-bar evaluation of 61 Japanese candlestick patterns.
    Call ``update(open, high, low, close)`` for each new bar, then call any pattern
    method to get the result for the current bar.

    Pattern methods return:
        +100 for bullish match, -100 for bearish match, 0 for no match.
        Some patterns return +50/-50 for unconfirmed signals (Hikkake, HikkakeModified).
    """

    def __init__(
        self,
        long_body: Criterion | None = None,
        very_long_body: Criterion | None = None,
        short_body: Criterion | None = None,
        doji_body: Criterion | None = None,
        long_shadow: Criterion | None = None,
        very_long_shadow: Criterion | None = None,
        short_shadow: Criterion | None = None,
        very_short_shadow: Criterion | None = None,
        near: Criterion | None = None,
        far: Criterion | None = None,
        equal: Criterion | None = None,
    ) -> None:
        # Criteria (use copies of defaults so mutations don't leak).
        self._long_body = _CriterionState(
            (long_body or DEFAULT_LONG_BODY).copy())
        self._very_long_body = _CriterionState(
            (very_long_body or DEFAULT_VERY_LONG_BODY).copy())
        self._short_body = _CriterionState(
            (short_body or DEFAULT_SHORT_BODY).copy())
        self._doji_body = _CriterionState(
            (doji_body or DEFAULT_DOJI_BODY).copy())
        self._long_shadow = _CriterionState(
            (long_shadow or DEFAULT_LONG_SHADOW).copy())
        self._very_long_shadow = _CriterionState(
            (very_long_shadow or DEFAULT_VERY_LONG_SHADOW).copy())
        self._short_shadow = _CriterionState(
            (short_shadow or DEFAULT_SHORT_SHADOW).copy())
        self._very_short_shadow = _CriterionState(
            (very_short_shadow or DEFAULT_VERY_SHORT_SHADOW).copy())
        self._near = _CriterionState(
            (near or DEFAULT_NEAR).copy())
        self._far = _CriterionState(
            (far or DEFAULT_FAR).copy())
        self._equal = _CriterionState(
            (equal or DEFAULT_EQUAL).copy())

        self._all_states = [
            self._long_body, self._very_long_body, self._short_body, self._doji_body,
            self._long_shadow, self._very_long_shadow, self._short_shadow, self._very_short_shadow,
            self._near, self._far, self._equal,
        ]

        # Ring buffer of recent bars: each entry is (open, high, low, close).
        # Size is the largest criterion period + 5 candles + 5 margin, floored at _MIN_HISTORY.
        max_period = max(s.criterion.average_period for s in self._all_states)
        history_size = max(max_period + 10, _MIN_HISTORY)
        self._history: deque[tuple[float, float, float, float]] = deque(maxlen=history_size)
        self._count = 0

        # Stateful pattern state: hikkake_modified
        self._hikmod_pattern_result = 0
        self._hikmod_pattern_idx = 0
        self._hikmod_confirmed = False
        self._hikmod_last_signal = 0

    # ------------------------------------------------------------------
    # Streaming interface
    # ------------------------------------------------------------------

    def update(self, o: float, h: float, l: float, c: float) -> None:
        """Feed a new OHLC bar into the engine.

        After calling this, all pattern methods reflect the state including this bar.
        """
        self._history.append((o, h, l, c))
        for state in self._all_states:
            state.push(o, h, l, c)
        self._count += 1

        # Update stateful patterns.
        self._hikmod_confirmed = False
        self._hikmod_last_signal = 0
        from .patterns.hikkake_modified import _hikkake_modified_update
        _hikkake_modified_update(self)

    @property
    def count(self) -> int:
        """Number of bars fed so far."""
        return self._count

    # ------------------------------------------------------------------
    # Helper: get bar at position relative to end.
    # shift=1 means the most recent bar, shift=2 the one before, etc.
    # ------------------------------------------------------------------

    def _bar(self, shift: int) -> tuple[float, float, float, float]:
        """Get OHLC of a bar. shift=1 is most recent, shift=2 is one before, etc."""
        return self._history[-shift]

    def _has(self, n: int) -> bool:
        """Check if we have at least n bars in history."""
        return len(self._history) >= n

    def _enough(self, n_candles: int, *criteria: _CriterionState) -> bool:
        """Check if we have sufficient bars for a pattern requiring n_candles
        plus the maximum average_period of the given criteria."""
        avail = len(self._history) - n_candles
        for cs in criteria:
            if avail < cs.criterion.average_period:
                return False
        return True

    # ------------------------------------------------------------------
    # Criterion average helpers (shift is from the end, 1-based)
    # ------------------------------------------------------------------

    def _avg(self, cs: _CriterionState, shift: int) -> float:
        """Get the criterion average value at a given shift from the most recent bar.
        
        TA-Lib convention: the average uses the `period` bars BEFORE the reference bar
        (excluding the reference bar itself).
        """
        o, h, l, c = self._bar(shift)
        return cs.avg(self._history, shift, o, h, l, c)

    def _avg_ref(self, cs: _CriterionState, shift: int, ref_shift: int) -> float:
        """Get the criterion average with the window ending at `shift` but using
        `ref_shift` bar's OHLC as the reference (for period==0 criteria)."""
        o, h, l, c = self._bar(ref_shift)
        return cs.avg(self._history, shift - 1, o, h, l, c)

    # ------------------------------------------------------------------
    # Pattern methods -- 61 patterns
    # ------------------------------------------------------------------

    # Import pattern methods as mix-ins to keep this file manageable.
    from .patterns.abandoned_baby import abandoned_baby
    from .patterns.advance_block import advance_block
    from .patterns.belt_hold import belt_hold
    from .patterns.breakaway import breakaway
    from .patterns.closing_marubozu import closing_marubozu
    from .patterns.concealing_baby_swallow import concealing_baby_swallow
    from .patterns.counterattack import counterattack
    from .patterns.dark_cloud_cover import dark_cloud_cover
    from .patterns.doji import doji
    from .patterns.doji_star import doji_star
    from .patterns.dragonfly_doji import dragonfly_doji
    from .patterns.engulfing import engulfing
    from .patterns.evening_doji_star import evening_doji_star
    from .patterns.evening_star import evening_star
    from .patterns.gravestone_doji import gravestone_doji
    from .patterns.hammer import hammer
    from .patterns.hanging_man import hanging_man
    from .patterns.harami import harami
    from .patterns.harami_cross import harami_cross
    from .patterns.high_wave import high_wave
    from .patterns.hikkake import hikkake
    from .patterns.hikkake_modified import hikkake_modified
    from .patterns.homing_pigeon import homing_pigeon
    from .patterns.identical_three_crows import identical_three_crows
    from .patterns.in_neck import in_neck
    from .patterns.inverted_hammer import inverted_hammer
    from .patterns.kicking import kicking
    from .patterns.kicking_by_length import kicking_by_length
    from .patterns.ladder_bottom import ladder_bottom
    from .patterns.long_legged_doji import long_legged_doji
    from .patterns.long_line import long_line
    from .patterns.marubozu import marubozu
    from .patterns.matching_low import matching_low
    from .patterns.mat_hold import mat_hold
    from .patterns.morning_doji_star import morning_doji_star
    from .patterns.morning_star import morning_star
    from .patterns.on_neck import on_neck
    from .patterns.piercing import piercing
    from .patterns.rickshaw_man import rickshaw_man
    from .patterns.rising_falling_three_methods import rising_falling_three_methods
    from .patterns.separating_lines import separating_lines
    from .patterns.shooting_star import shooting_star
    from .patterns.short_line import short_line
    from .patterns.spinning_top import spinning_top
    from .patterns.stalled import stalled
    from .patterns.stick_sandwich import stick_sandwich
    from .patterns.takuri import takuri
    from .patterns.tasuki_gap import tasuki_gap
    from .patterns.three_black_crows import three_black_crows
    from .patterns.three_inside import three_inside
    from .patterns.three_line_strike import three_line_strike
    from .patterns.three_outside import three_outside
    from .patterns.three_stars_in_the_south import three_stars_in_the_south
    from .patterns.three_white_soldiers import three_white_soldiers
    from .patterns.thrusting import thrusting
    from .patterns.tristar import tristar
    from .patterns.two_crows import two_crows
    from .patterns.unique_three_river import unique_three_river
    from .patterns.up_down_gap_side_by_side_white_lines import up_down_gap_side_by_side_white_lines
    from .patterns.upside_gap_two_crows import upside_gap_two_crows
    from .patterns.x_side_gap_three_methods import x_side_gap_three_methods

    # ------------------------------------------------------------------
    # Evaluate by PatternIdentifier
    # ------------------------------------------------------------------

    def evaluate(self, pattern_identifier: PatternIdentifier) -> int:
        """Evaluate a single pattern by its identifier.

        Args:
            pattern_identifier: A ``PatternIdentifier`` enum member.

        Returns:
            The pattern result (e.g. -100, 0, +100, +200).
        """
        return _DISPATCH[pattern_identifier](self)


# Dispatch table: PatternIdentifier → unbound method.  Built once at import time.
_DISPATCH: dict[PatternIdentifier, object] = {
    PatternIdentifier(i): getattr(CandlestickPatterns, PatternIdentifier(i).method_name)
    for i in range(len(PatternIdentifier))
}
