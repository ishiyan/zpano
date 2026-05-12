// ---------------------------------------------------------------------------
// RangeEntity
// ---------------------------------------------------------------------------

/// The entities of range that can be considered when comparing a part of a candlestick
/// to other candlesticks.
pub const RangeEntity = enum(u8) {
    /// Identifies the length of the real body of a candlestick.
    real_body = 0,
    /// Identifies the length of the high-low range of a candlestick.
    high_low = 1,
    /// Identifies the length of the shadows of a candlestick.
    shadows = 2,
};
