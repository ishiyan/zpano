"""Candlestick primitives: color, body, shadow, gap, and enclosure tests.

These are pure functions operating on OHLCV tuples (or any object with
open, high, low, close attributes).
"""
from __future__ import annotations

from .range_entity import RangeEntity


# ---------------------------------------------------------------------------
# Color
# ---------------------------------------------------------------------------

def is_white(o: float, c: float) -> bool:
    """A candlestick is white (bullish) when close >= open."""
    return c >= o


def is_black(o: float, c: float) -> bool:
    """A candlestick is black (bearish) when close < open."""
    return c < o


# ---------------------------------------------------------------------------
# Real body
# ---------------------------------------------------------------------------

def real_body(o: float, c: float) -> float:
    """Absolute length of the real body."""
    return c - o if c >= o else o - c


def white_real_body(o: float, c: float) -> float:
    """Length of the real body of a white candlestick (close - open)."""
    return c - o


def black_real_body(o: float, c: float) -> float:
    """Length of the real body of a black candlestick (open - close)."""
    return o - c


# ---------------------------------------------------------------------------
# Shadows
# ---------------------------------------------------------------------------

def upper_shadow(o: float, h: float, c: float) -> float:
    """Length of the upper shadow."""
    return h - (c if c >= o else o)


def lower_shadow(o: float, l: float, c: float) -> float:
    """Length of the lower shadow."""
    return (o if c >= o else c) - l


def white_upper_shadow(h: float, c: float) -> float:
    """Length of the upper shadow of a white candlestick."""
    return h - c


def black_upper_shadow(o: float, h: float) -> float:
    """Length of the upper shadow of a black candlestick."""
    return h - o


def white_lower_shadow(o: float, l: float) -> float:
    """Length of the lower shadow of a white candlestick."""
    return o - l


def black_lower_shadow(l: float, c: float) -> float:
    """Length of the lower shadow of a black candlestick."""
    return c - l


# ---------------------------------------------------------------------------
# Gap tests
# ---------------------------------------------------------------------------

def is_real_body_gap_up(o1: float, c1: float, o2: float, c2: float) -> bool:
    """Real body gap up: max(open1, close1) < min(open2, close2)."""
    return max(o1, c1) < min(o2, c2)


def is_real_body_gap_down(o1: float, c1: float, o2: float, c2: float) -> bool:
    """Real body gap down: min(open1, close1) > max(open2, close2)."""
    return min(o1, c1) > max(o2, c2)


def is_high_low_gap_up(h1: float, l2: float) -> bool:
    """High-low gap up: high of first candle < low of second candle."""
    return h1 < l2


def is_high_low_gap_down(l1: float, h2: float) -> bool:
    """High-low gap down: low of first candle > high of second candle."""
    return l1 > h2


# ---------------------------------------------------------------------------
# Enclosure tests
# ---------------------------------------------------------------------------

def is_real_body_encloses_real_body(o1: float, c1: float, o2: float, c2: float) -> bool:
    """The real body of candle 1 completely encloses the real body of candle 2."""
    min1, max1 = (o1, c1) if c1 > o1 else (c1, o1)
    min2, max2 = (o2, c2) if c2 > o2 else (c2, o2)
    return max1 > max2 and min1 < min2


def is_real_body_encloses_open(o1: float, c1: float, o2: float) -> bool:
    """The real body of candle 1 encloses the open of candle 2."""
    if o1 > c1:
        return o2 < o1 and o2 > c1
    return o2 > o1 and o2 < c1


def is_real_body_encloses_close(o1: float, c1: float, c2: float) -> bool:
    """The real body of candle 1 encloses the close of candle 2."""
    if o1 > c1:
        return c2 < o1 and c2 > c1
    return c2 > o1 and c2 < c1


# ---------------------------------------------------------------------------
# Misc comparisons
# ---------------------------------------------------------------------------

def is_high_exceeds_close(h1: float, c2: float) -> bool:
    """High of candle 1 is greater than close of candle 2."""
    return h1 > c2


def is_opens_within(o1: float, o2: float, c2: float, tolerance: float = 0.0) -> bool:
    """Candle 1 opens within the real body of candle 2 (with optional tolerance)."""
    return o1 >= min(o2, c2) - tolerance and o1 <= max(o2, c2) + tolerance


# ---------------------------------------------------------------------------
# Range value for a single candle (used by Criterion)
# ---------------------------------------------------------------------------

def candle_range_value(entity: RangeEntity, o: float, h: float, l: float, c: float) -> float:
    """Compute the range value of a candle for a given RangeEntity type."""
    if entity == RangeEntity.REAL_BODY:
        return c - o if c >= o else o - c
    if entity == RangeEntity.HIGH_LOW:
        return h - l
    # SHADOWS: average of upper and lower shadow
    if c >= o:
        return (h - c + o - l) / 2.0
    return (h - o + c - l) / 2.0
