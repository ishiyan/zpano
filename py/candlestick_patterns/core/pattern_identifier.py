"""Enumeration of the 61 candlestick pattern identifiers."""
from __future__ import annotations

from enum import IntEnum


class PatternIdentifier(IntEnum):
    """Identifier for each of the 61 candlestick patterns.

    Values are sequential starting at 0, sorted alphabetically.
    The ``name`` attribute matches the method name on ``CandlestickPatterns``.
    """
    ABANDONED_BABY = 0
    ADVANCE_BLOCK = 1
    BELT_HOLD = 2
    BREAKAWAY = 3
    CLOSING_MARUBOZU = 4
    CONCEALING_BABY_SWALLOW = 5
    COUNTERATTACK = 6
    DARK_CLOUD_COVER = 7
    DOJI = 8
    DOJI_STAR = 9
    DRAGONFLY_DOJI = 10
    ENGULFING = 11
    EVENING_DOJI_STAR = 12
    EVENING_STAR = 13
    GRAVESTONE_DOJI = 14
    HAMMER = 15
    HANGING_MAN = 16
    HARAMI = 17
    HARAMI_CROSS = 18
    HIGH_WAVE = 19
    HIKKAKE = 20
    HIKKAKE_MODIFIED = 21
    HOMING_PIGEON = 22
    IDENTICAL_THREE_CROWS = 23
    IN_NECK = 24
    INVERTED_HAMMER = 25
    KICKING = 26
    KICKING_BY_LENGTH = 27
    LADDER_BOTTOM = 28
    LONG_LEGGED_DOJI = 29
    LONG_LINE = 30
    MARUBOZU = 31
    MATCHING_LOW = 32
    MAT_HOLD = 33
    MORNING_DOJI_STAR = 34
    MORNING_STAR = 35
    ON_NECK = 36
    PIERCING = 37
    RICKSHAW_MAN = 38
    RISING_FALLING_THREE_METHODS = 39
    SEPARATING_LINES = 40
    SHOOTING_STAR = 41
    SHORT_LINE = 42
    SPINNING_TOP = 43
    STALLED = 44
    STICK_SANDWICH = 45
    TAKURI = 46
    TASUKI_GAP = 47
    THREE_BLACK_CROWS = 48
    THREE_INSIDE = 49
    THREE_LINE_STRIKE = 50
    THREE_OUTSIDE = 51
    THREE_STARS_IN_THE_SOUTH = 52
    THREE_WHITE_SOLDIERS = 53
    THRUSTING = 54
    TRISTAR = 55
    TWO_CROWS = 56
    UNIQUE_THREE_RIVER = 57
    UP_DOWN_GAP_SIDE_BY_SIDE_WHITE_LINES = 58
    UPSIDE_GAP_TWO_CROWS = 59
    X_SIDE_GAP_THREE_METHODS = 60

    @property
    def method_name(self) -> str:
        """The snake_case method name on ``CandlestickPatterns``."""
        return self.name.lower()
