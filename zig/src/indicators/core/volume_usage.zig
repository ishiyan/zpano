/// Classifies how an indicator uses volume information.
pub const VolumeUsage = enum(u8) {
    /// Does not use volume.
    no_volume = 1,
    /// Consumes per-bar aggregated volume.
    aggregate_bar_volume = 2,
    /// Consumes per-trade volume.
    per_trade_volume = 3,
    /// Consumes quote-side liquidity (bid/ask sizes).
    quote_liquidity_volume = 4,

    pub fn asStr(self: VolumeUsage) []const u8 {
        return switch (self) {
            .no_volume => "none",
            .aggregate_bar_volume => "aggregateBar",
            .per_trade_volume => "perTrade",
            .quote_liquidity_volume => "quoteLiquidity",
        };
    }
};
