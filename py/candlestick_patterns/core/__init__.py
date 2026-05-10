"""Core types and primitives for candlestick pattern recognition."""

from .range_entity import RangeEntity
from .criterion import Criterion
from .primitives import (
    is_white, is_black, real_body, white_real_body, black_real_body,
    upper_shadow, lower_shadow, white_upper_shadow, black_upper_shadow,
    white_lower_shadow, black_lower_shadow,
    is_real_body_gap_up, is_real_body_gap_down,
    is_high_low_gap_up, is_high_low_gap_down,
    is_real_body_encloses_real_body, is_real_body_encloses_open, is_real_body_encloses_close,
    is_high_exceeds_close, is_opens_within,
    candle_range_value,
)
from .defaults import (
    DEFAULT_LONG_BODY, DEFAULT_VERY_LONG_BODY, DEFAULT_SHORT_BODY, DEFAULT_DOJI_BODY,
    DEFAULT_LONG_SHADOW, DEFAULT_VERY_LONG_SHADOW, DEFAULT_SHORT_SHADOW, DEFAULT_VERY_SHORT_SHADOW,
    DEFAULT_NEAR, DEFAULT_FAR, DEFAULT_EQUAL,
)
