// ---------------------------------------------------------------------------
// Criterion
// ---------------------------------------------------------------------------
//
// Specifies a threshold based on the average value of a candlestick range entity.
//
// The criteria are based on parts of the candlestick and common words indicating length
// (short, long, very long), displacement (near, far), or equality (equal).
//
// For streaming efficiency, the criterion maintains a running total that is updated
// incrementally via add() and remove() rather than rescanning the entire history.

const RangeEntity = @import("range_entity.zig").RangeEntity;
const primitives = @import("primitives.zig");

/// A criterion based on the average value of a certain part of a candlestick multiplied by a factor.
pub const Criterion = struct {
    /// The type of range entity to consider.
    entity: RangeEntity,
    /// Number of previous candlesticks to calculate an average value.
    average_period: usize,
    /// Coefficient to multiply the average value.
    factor: f64,

    /// Computes the criterion threshold from a precomputed running total.
    ///
    /// When average_period > 0, uses the running total.
    /// When average_period == 0, uses the current candle's own range value.
    pub fn averageValueFromTotal(self: Criterion, total: f64, o: f64, h: f64, l: f64, c: f64) f64 {
        if (self.average_period > 0) {
            if (self.entity == .shadows) {
                return self.factor * total / (@as(f64, @floatFromInt(self.average_period)) * 2.0);
            }
            return self.factor * total / @as(f64, @floatFromInt(self.average_period));
        }
        // Period == 0: use the candle's own range value directly.
        return self.factor * primitives.candleRangeValue(self.entity, o, h, l, c);
    }

    /// Computes the contribution of a single candle to the running total.
    ///
    /// For SHADOWS entity, this returns the full (upper + lower) shadow sum
    /// (not yet divided by 2 -- the division happens in averageValueFromTotal).
    pub fn candleContribution(self: Criterion, o: f64, h: f64, l: f64, c: f64) f64 {
        switch (self.entity) {
            .real_body => {
                if (c >= o) return c - o;
                return o - c;
            },
            .high_low => return h - l,
            .shadows => {
                // upper + lower shadow sum (division by 2 deferred to averageValueFromTotal)
                if (c >= o) return h - c + o - l;
                return h - o + c - l;
            },
        }
    }
};
