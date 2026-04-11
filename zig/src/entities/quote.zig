const std = @import("std");
const testing = std.testing;

/// Represents a price quote (bid/ask price and size pair).
pub const Quote = struct {
    time: i64,
    bid_price: f64,
    ask_price: f64,
    bid_size: f64,
    ask_size: f64,

    /// The mid-price: (ask_price + bid_price) / 2.
    pub fn mid(self: Quote) f64 {
        return (self.ask_price + self.bid_price) / 2.0;
    }

    /// The weighted price: (ask*askSize + bid*bidSize) / (askSize + bidSize).
    pub fn weighted(self: Quote) f64 {
        const size = self.ask_size + self.bid_size;
        if (size == 0) return 0;
        return (self.ask_price * self.ask_size + self.bid_price * self.bid_size) / size;
    }

    /// The weighted mid-price (micro-price): (ask*bidSize + bid*askSize) / (askSize + bidSize).
    pub fn weightedMid(self: Quote) f64 {
        const size = self.ask_size + self.bid_size;
        if (size == 0) return 0;
        return (self.ask_price * self.bid_size + self.bid_price * self.ask_size) / size;
    }

    /// The spread in basis points: 20000 * (ask - bid) / (ask + bid).
    pub fn spreadBp(self: Quote) f64 {
        const m = self.ask_price + self.bid_price;
        if (m == 0) return 0;
        return 20000.0 * (self.ask_price - self.bid_price) / m;
    }
};

fn makeQuote(bid: f64, ask: f64, bs: f64, as_: f64) Quote {
    return Quote{ .time = 0, .bid_price = bid, .ask_price = ask, .bid_size = bs, .ask_size = as_ };
}

test "quote mid" {
    const q = makeQuote(3.0, 2.0, 0, 0);
    try testing.expectEqual((q.ask_price + q.bid_price) / 2.0, q.mid());
}

test "quote weighted" {
    const q = makeQuote(3.0, 2.0, 5.0, 4.0);
    const expected = (q.ask_price * q.ask_size + q.bid_price * q.bid_size) / (q.ask_size + q.bid_size);
    try testing.expectEqual(expected, q.weighted());
}

test "quote weighted zero size" {
    const q = makeQuote(3.0, 2.0, 0, 0);
    try testing.expectEqual(@as(f64, 0.0), q.weighted());
}

test "quote weightedMid" {
    const q = makeQuote(3.0, 2.0, 5.0, 4.0);
    const expected = (q.ask_price * q.bid_size + q.bid_price * q.ask_size) / (q.ask_size + q.bid_size);
    try testing.expectEqual(expected, q.weightedMid());
}

test "quote weightedMid zero size" {
    const q = makeQuote(3.0, 2.0, 0, 0);
    try testing.expectEqual(@as(f64, 0.0), q.weightedMid());
}

test "quote spreadBp" {
    const q = makeQuote(3.0, 2.0, 0, 0);
    const expected = 20000.0 * (q.ask_price - q.bid_price) / (q.ask_price + q.bid_price);
    try testing.expectEqual(expected, q.spreadBp());
}

test "quote spreadBp zero mid" {
    const q = makeQuote(0, 0, 0, 0);
    try testing.expectEqual(@as(f64, 0.0), q.spreadBp());
}
