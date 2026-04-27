/// Classifies the minimum input data type an indicator consumes.
pub const InputRequirement = enum(u8) {
    /// Consumes a scalar time series.
    scalar_input = 1,
    /// Consumes level-1 quotes.
    quote_input = 2,
    /// Consumes OHLCV bars.
    bar_input = 3,
    /// Consumes individual trades.
    trade_input = 4,

    pub fn asStr(self: InputRequirement) []const u8 {
        return switch (self) {
            .scalar_input => "scalar",
            .quote_input => "quote",
            .bar_input => "bar",
            .trade_input => "trade",
        };
    }
};
