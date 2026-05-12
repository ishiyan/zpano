// ---------------------------------------------------------------------------
// Core types for the candlestick patterns module.
// ---------------------------------------------------------------------------

pub const range_entity = @import("range_entity.zig");
pub const ohlc = @import("ohlc.zig");
pub const primitives = @import("primitives.zig");
pub const criterion = @import("criterion.zig");
pub const defaults = @import("defaults.zig");
pub const criterion_state = @import("criterion_state.zig");
pub const pattern_identifier = @import("pattern_identifier.zig");

// Re-export primary types for convenience.
pub const RangeEntity = range_entity.RangeEntity;
pub const OHLC = ohlc.OHLC;
pub const Criterion = criterion.Criterion;
pub const CriterionState = criterion_state.CriterionState;
pub const PatternIdentifier = pattern_identifier.PatternIdentifier;
pub const pattern_count = pattern_identifier.pattern_count;

// Re-export all primitives.
pub const isWhite = primitives.isWhite;
pub const isBlack = primitives.isBlack;
pub const realBodyLen = primitives.realBodyLen;
pub const upperShadow = primitives.upperShadow;
pub const lowerShadow = primitives.lowerShadow;
pub const isRealBodyGapUp = primitives.isRealBodyGapUp;
pub const isRealBodyGapDown = primitives.isRealBodyGapDown;
pub const isHighLowGapUp = primitives.isHighLowGapUp;
pub const isHighLowGapDown = primitives.isHighLowGapDown;
pub const candleRangeValue = primitives.candleRangeValue;

// Re-export all defaults.
pub const default_long_body = defaults.default_long_body;
pub const default_very_long_body = defaults.default_very_long_body;
pub const default_short_body = defaults.default_short_body;
pub const default_doji_body = defaults.default_doji_body;
pub const default_long_shadow = defaults.default_long_shadow;
pub const default_very_long_shadow = defaults.default_very_long_shadow;
pub const default_short_shadow = defaults.default_short_shadow;
pub const default_very_short_shadow = defaults.default_very_short_shadow;
pub const default_near = defaults.default_near;
pub const default_far = defaults.default_far;
pub const default_equal = defaults.default_equal;
