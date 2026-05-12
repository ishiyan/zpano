// Shared test case type for candlestick pattern test data.

pub const TestCase = struct {
    opens: [20]f64,
    highs: [20]f64,
    lows: [20]f64,
    closes: [20]f64,
    expected: i32,
};
