// ---------------------------------------------------------------------------
// Primitives
// ---------------------------------------------------------------------------
//
// Candlestick primitives: color, body, shadow, gap, enclosure, and range functions.
// These are pure functions operating on OHLC values.

const RangeEntity = @import("range_entity.zig").RangeEntity;

// ---------------------------------------------------------------------------
// Color
// ---------------------------------------------------------------------------

/// Returns true when a candlestick is white (bullish): close >= open.
pub fn isWhite(o: f64, c: f64) bool {
    return c >= o;
}

/// Returns true when a candlestick is black (bearish): close < open.
pub fn isBlack(o: f64, c: f64) bool {
    return c < o;
}

// ---------------------------------------------------------------------------
// Real body
// ---------------------------------------------------------------------------

/// Returns the absolute length of the real body.
pub fn realBodyLen(o: f64, c: f64) f64 {
    if (c >= o) return c - o;
    return o - c;
}

// ---------------------------------------------------------------------------
// Shadows
// ---------------------------------------------------------------------------

/// Returns the length of the upper shadow.
pub fn upperShadow(o: f64, h: f64, c: f64) f64 {
    if (c >= o) return h - c;
    return h - o;
}

/// Returns the length of the lower shadow.
pub fn lowerShadow(o: f64, l: f64, c: f64) f64 {
    if (c >= o) return o - l;
    return c - l;
}

// ---------------------------------------------------------------------------
// Gap tests
// ---------------------------------------------------------------------------

/// Returns true when max(open1, close1) < min(open2, close2).
pub fn isRealBodyGapUp(o1: f64, c1: f64, o2: f64, c2: f64) bool {
    return @max(o1, c1) < @min(o2, c2);
}

/// Returns true when min(open1, close1) > max(open2, close2).
pub fn isRealBodyGapDown(o1: f64, c1: f64, o2: f64, c2: f64) bool {
    return @min(o1, c1) > @max(o2, c2);
}

/// Returns true when high of first candle < low of second candle.
pub fn isHighLowGapUp(h1: f64, l2: f64) bool {
    return h1 < l2;
}

/// Returns true when low of first candle > high of second candle.
pub fn isHighLowGapDown(l1: f64, h2: f64) bool {
    return l1 > h2;
}

// ---------------------------------------------------------------------------
// Range value for a single candle (used by Criterion)
// ---------------------------------------------------------------------------

/// Computes the range value of a candle for a given RangeEntity type.
pub fn candleRangeValue(entity: RangeEntity, o: f64, h: f64, l: f64, c: f64) f64 {
    switch (entity) {
        .real_body => {
            if (c >= o) return c - o;
            return o - c;
        },
        .high_low => return h - l,
        .shadows => {
            // SHADOWS: average of upper and lower shadow
            if (c >= o) return (h - c + o - l) / 2.0;
            return (h - o + c - l) / 2.0;
        },
    }
}
