const std = @import("std");
const testing = std.testing;
const Bar = @import("bar").Bar;

/// Describes a component of the Bar type.
pub const BarComponent = enum(u8) {
    open = 0,
    high = 1,
    low = 2,
    close = 3,
    volume = 4,
    median = 5,
    typical = 6,
    weighted = 7,
    average = 8,
};

/// Function type that extracts a component value from a Bar.
pub const BarFunc = *const fn (Bar) f64;

/// Returns a function that extracts the given component value from a Bar.
/// For unknown components, returns the close price.
pub fn componentValue(component: BarComponent) BarFunc {
    return switch (component) {
        .open => barOpen,
        .high => barHigh,
        .low => barLow,
        .close => barClose,
        .volume => barVolume,
        .median => barMedian,
        .typical => barTypical,
        .weighted => barWeighted,
        .average => barAverage,
    };
}

/// Returns the mnemonic string for the given bar component.
pub fn componentMnemonic(component: BarComponent) []const u8 {
    return switch (component) {
        .open => "o",
        .high => "h",
        .low => "l",
        .close => "c",
        .volume => "v",
        .median => "hl/2",
        .typical => "hlc/3",
        .weighted => "hlcc/4",
        .average => "ohlc/4",
    };
}

fn barOpen(b: Bar) f64 {
    return b.open;
}
fn barHigh(b: Bar) f64 {
    return b.high;
}
fn barLow(b: Bar) f64 {
    return b.low;
}
fn barClose(b: Bar) f64 {
    return b.close;
}
fn barVolume(b: Bar) f64 {
    return b.volume;
}
fn barMedian(b: Bar) f64 {
    return b.median();
}
fn barTypical(b: Bar) f64 {
    return b.typical();
}
fn barWeighted(b: Bar) f64 {
    return b.weighted();
}
fn barAverage(b: Bar) f64 {
    return b.average();
}

fn makeBar(o: f64, h: f64, l: f64, c: f64, v: f64) Bar {
    return Bar{ .time = 0, .open = o, .high = h, .low = l, .close = c, .volume = v };
}

test "bar component value open" {
    const b = makeBar(2, 4, 1, 3, 5);
    try testing.expectEqual(@as(f64, 2), componentValue(.open)(b));
}

test "bar component value high" {
    const b = makeBar(2, 4, 1, 3, 5);
    try testing.expectEqual(@as(f64, 4), componentValue(.high)(b));
}

test "bar component value low" {
    const b = makeBar(2, 4, 1, 3, 5);
    try testing.expectEqual(@as(f64, 1), componentValue(.low)(b));
}

test "bar component value close" {
    const b = makeBar(2, 4, 1, 3, 5);
    try testing.expectEqual(@as(f64, 3), componentValue(.close)(b));
}

test "bar component value volume" {
    const b = makeBar(2, 4, 1, 3, 5);
    try testing.expectEqual(@as(f64, 5), componentValue(.volume)(b));
}

test "bar component value median" {
    const b = makeBar(2, 4, 1, 3, 5);
    try testing.expectEqual((1.0 + 4.0) / 2.0, componentValue(.median)(b));
}

test "bar component value typical" {
    const b = makeBar(2, 4, 1, 3, 5);
    try testing.expectEqual((1.0 + 4.0 + 3.0) / 3.0, componentValue(.typical)(b));
}

test "bar component value weighted" {
    const b = makeBar(2, 4, 1, 3, 5);
    try testing.expectEqual((1.0 + 4.0 + 3.0 + 3.0) / 4.0, componentValue(.weighted)(b));
}

test "bar component value average" {
    const b = makeBar(2, 4, 1, 3, 5);
    try testing.expectEqual((1.0 + 4.0 + 3.0 + 2.0) / 4.0, componentValue(.average)(b));
}

test "bar component mnemonic" {
    try testing.expectEqualStrings("o", componentMnemonic(.open));
    try testing.expectEqualStrings("h", componentMnemonic(.high));
    try testing.expectEqualStrings("l", componentMnemonic(.low));
    try testing.expectEqualStrings("c", componentMnemonic(.close));
    try testing.expectEqualStrings("v", componentMnemonic(.volume));
    try testing.expectEqualStrings("hl/2", componentMnemonic(.median));
    try testing.expectEqualStrings("hlc/3", componentMnemonic(.typical));
    try testing.expectEqualStrings("hlcc/4", componentMnemonic(.weighted));
    try testing.expectEqualStrings("ohlc/4", componentMnemonic(.average));
}
