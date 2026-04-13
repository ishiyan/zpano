const std = @import("std");
const testing = std.testing;
const Trade = @import("trade").Trade;

/// Describes a component of the Trade type.
pub const TradeComponent = enum(u8) {
    price = 0,
    volume = 1,
};

/// The default trade component used when no explicit component is specified.
pub const default_trade_component: TradeComponent = .price;

/// Function type that extracts a component value from a Trade.
pub const TradeFunc = *const fn (Trade) f64;

/// Returns a function that extracts the given component value from a Trade.
/// For unknown components, returns the price.
pub fn componentValue(component: TradeComponent) TradeFunc {
    return switch (component) {
        .price => tradePrice,
        .volume => tradeVolume,
    };
}

/// Returns the mnemonic string for the given trade component.
pub fn componentMnemonic(component: TradeComponent) []const u8 {
    return switch (component) {
        .price => "p",
        .volume => "v",
    };
}

fn tradePrice(t: Trade) f64 {
    return t.price;
}
fn tradeVolume(t: Trade) f64 {
    return t.volume;
}

test "trade component value price" {
    const t = Trade{ .time = 0, .price = 1.0, .volume = 2.0 };
    try testing.expectEqual(@as(f64, 1.0), componentValue(.price)(t));
}

test "trade component value volume" {
    const t = Trade{ .time = 0, .price = 1.0, .volume = 2.0 };
    try testing.expectEqual(@as(f64, 2.0), componentValue(.volume)(t));
}

test "trade component mnemonic" {
    try testing.expectEqualStrings("p", componentMnemonic(.price));
    try testing.expectEqualStrings("v", componentMnemonic(.volume));
}

test "default trade component" {
    try testing.expectEqual(TradeComponent.price, default_trade_component);
}
