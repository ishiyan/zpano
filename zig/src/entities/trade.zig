const std = @import("std");

/// Represents a trade (time and sales) with price and volume.
pub const Trade = struct {
    time: i64,
    price: f64,
    volume: f64,
};
