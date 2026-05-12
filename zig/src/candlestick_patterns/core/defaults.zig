// ---------------------------------------------------------------------------
// Default criterion definitions matching the Ta-Lib implementation.
// ---------------------------------------------------------------------------

const RangeEntity = @import("range_entity.zig").RangeEntity;
const Criterion = @import("criterion.zig").Criterion;

/// Real body is long when it is longer than the average of the real body
/// of the 10 previous candlesticks.
pub const default_long_body = Criterion{ .entity = .real_body, .average_period = 10, .factor = 1.0 };

/// Real body is very long when it is longer than 3 times the average
/// of the real body of the 10 previous candlesticks.
pub const default_very_long_body = Criterion{ .entity = .real_body, .average_period = 10, .factor = 3.0 };

/// Real body is short when it is shorter than the average of the real body
/// of the 10 previous candlesticks.
pub const default_short_body = Criterion{ .entity = .real_body, .average_period = 10, .factor = 1.0 };

/// Real body is like doji when it is shorter than 10% the average of the
/// high-low range of the 10 previous candlesticks.
pub const default_doji_body = Criterion{ .entity = .high_low, .average_period = 10, .factor = 0.1 };

/// Shadow is long when longer than the real body.
pub const default_long_shadow = Criterion{ .entity = .real_body, .average_period = 0, .factor = 1.0 };

/// Shadow is very long when longer than 2x the real body.
pub const default_very_long_shadow = Criterion{ .entity = .real_body, .average_period = 0, .factor = 2.0 };

/// Shadow is short when it is shorter than the average of the sum
/// of shadows of the 10 previous candlesticks.
pub const default_short_shadow = Criterion{ .entity = .shadows, .average_period = 10, .factor = 1.0 };

/// Shadow is very short when it is shorter than 10% the average of the
/// high-low range of the 10 previous candlesticks.
pub const default_very_short_shadow = Criterion{ .entity = .high_low, .average_period = 10, .factor = 0.1 };

/// When measuring distance between parts of candles or width of gaps,
/// 'near' means <= 20% of the average of the high-low range of the 5 previous candlesticks.
pub const default_near = Criterion{ .entity = .high_low, .average_period = 5, .factor = 0.2 };

/// When measuring distance between parts of candles or width of gaps,
/// 'far' means >= 60% of the average of the high-low range of the 5 previous candlesticks.
pub const default_far = Criterion{ .entity = .high_low, .average_period = 5, .factor = 0.6 };

/// When measuring distance between parts of candles or width of gaps,
/// 'equal' means <= 5% of the average of the high-low range of the 5 previous candlesticks.
pub const default_equal = Criterion{ .entity = .high_low, .average_period = 5, .factor = 0.05 };
