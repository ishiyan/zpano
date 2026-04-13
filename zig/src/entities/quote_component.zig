const std = @import("std");
const testing = std.testing;
const Quote = @import("quote").Quote;

/// Describes a component of the Quote type.
pub const QuoteComponent = enum(u8) {
    bid = 0,
    ask = 1,
    bid_size = 2,
    ask_size = 3,
    mid = 4,
    weighted = 5,
    weighted_mid = 6,
    spread_bp = 7,
};

/// The default quote component used when no explicit component is specified.
pub const default_quote_component: QuoteComponent = .mid;

/// Function type that extracts a component value from a Quote.
pub const QuoteFunc = *const fn (Quote) f64;

/// Returns a function that extracts the given component value from a Quote.
/// For unknown components, returns the mid price.
pub fn componentValue(component: QuoteComponent) QuoteFunc {
    return switch (component) {
        .bid => quoteBid,
        .ask => quoteAsk,
        .bid_size => quoteBidSize,
        .ask_size => quoteAskSize,
        .mid => quoteMid,
        .weighted => quoteWeighted,
        .weighted_mid => quoteWeightedMid,
        .spread_bp => quoteSpreadBp,
    };
}

/// Returns the mnemonic string for the given quote component.
pub fn componentMnemonic(component: QuoteComponent) []const u8 {
    return switch (component) {
        .bid => "b",
        .ask => "a",
        .bid_size => "bs",
        .ask_size => "as",
        .mid => "ba/2",
        .weighted => "(bbs+aas)/(bs+as)",
        .weighted_mid => "(bas+abs)/(bs+as)",
        .spread_bp => "spread bp",
    };
}

fn quoteBid(q: Quote) f64 {
    return q.bid_price;
}
fn quoteAsk(q: Quote) f64 {
    return q.ask_price;
}
fn quoteBidSize(q: Quote) f64 {
    return q.bid_size;
}
fn quoteAskSize(q: Quote) f64 {
    return q.ask_size;
}
fn quoteMid(q: Quote) f64 {
    return q.mid();
}
fn quoteWeighted(q: Quote) f64 {
    return q.weighted();
}
fn quoteWeightedMid(q: Quote) f64 {
    return q.weightedMid();
}
fn quoteSpreadBp(q: Quote) f64 {
    return q.spreadBp();
}

fn makeQuote(bid: f64, ask: f64, bs: f64, as_: f64) Quote {
    return Quote{ .time = 0, .bid_price = bid, .ask_price = ask, .bid_size = bs, .ask_size = as_ };
}

test "quote component value bid" {
    const q = makeQuote(2.0, 1.0, 4.0, 3.0);
    try testing.expectEqual(@as(f64, 2.0), componentValue(.bid)(q));
}

test "quote component value ask" {
    const q = makeQuote(2.0, 1.0, 4.0, 3.0);
    try testing.expectEqual(@as(f64, 1.0), componentValue(.ask)(q));
}

test "quote component value bid_size" {
    const q = makeQuote(2.0, 1.0, 4.0, 3.0);
    try testing.expectEqual(@as(f64, 4.0), componentValue(.bid_size)(q));
}

test "quote component value ask_size" {
    const q = makeQuote(2.0, 1.0, 4.0, 3.0);
    try testing.expectEqual(@as(f64, 3.0), componentValue(.ask_size)(q));
}

test "quote component value mid" {
    const q = makeQuote(2.0, 1.0, 4.0, 3.0);
    try testing.expectEqual((1.0 + 2.0) / 2.0, componentValue(.mid)(q));
}

test "quote component value weighted" {
    const q = makeQuote(2.0, 1.0, 4.0, 3.0);
    try testing.expectEqual((1.0 * 3.0 + 2.0 * 4.0) / (3.0 + 4.0), componentValue(.weighted)(q));
}

test "quote component value weighted_mid" {
    const q = makeQuote(2.0, 1.0, 4.0, 3.0);
    try testing.expectEqual((1.0 * 4.0 + 2.0 * 3.0) / (3.0 + 4.0), componentValue(.weighted_mid)(q));
}

test "quote component value spread_bp" {
    const q = makeQuote(2.0, 1.0, 4.0, 3.0);
    try testing.expectEqual(10000.0 * 2.0 * (1.0 - 2.0) / (1.0 + 2.0), componentValue(.spread_bp)(q));
}

test "quote component mnemonic" {
    try testing.expectEqualStrings("b", componentMnemonic(.bid));
    try testing.expectEqualStrings("a", componentMnemonic(.ask));
    try testing.expectEqualStrings("bs", componentMnemonic(.bid_size));
    try testing.expectEqualStrings("as", componentMnemonic(.ask_size));
    try testing.expectEqualStrings("ba/2", componentMnemonic(.mid));
    try testing.expectEqualStrings("(bbs+aas)/(bs+as)", componentMnemonic(.weighted));
    try testing.expectEqualStrings("(bas+abs)/(bs+as)", componentMnemonic(.weighted_mid));
    try testing.expectEqualStrings("spread bp", componentMnemonic(.spread_bp));
}

test "default quote component" {
    try testing.expectEqual(QuoteComponent.mid, default_quote_component);
}
