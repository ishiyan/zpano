/// Classifies how an indicator uses volume information.
#[derive(Debug, Clone, Copy, PartialEq, Eq)]
#[repr(u8)]
pub enum VolumeUsage {
    /// Does not use volume.
    NoVolume = 1,
    /// Consumes per-bar aggregated volume.
    AggregateBarVolume = 2,
    /// Consumes per-trade volume.
    PerTradeVolume = 3,
    /// Consumes quote-side liquidity (bid/ask sizes).
    QuoteLiquidityVolume = 4,
}

impl VolumeUsage {
    pub fn as_str(&self) -> &'static str {
        match self {
            Self::NoVolume => "none",
            Self::AggregateBarVolume => "aggregateBar",
            Self::PerTradeVolume => "perTrade",
            Self::QuoteLiquidityVolume => "quoteLiquidity",
        }
    }
}

impl std::fmt::Display for VolumeUsage {
    fn fmt(&self, f: &mut std::fmt::Formatter<'_>) -> std::fmt::Result {
        f.write_str(self.as_str())
    }
}
