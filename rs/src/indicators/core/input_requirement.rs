/// Classifies the minimum input data type an indicator consumes.
#[derive(Debug, Clone, Copy, PartialEq, Eq)]
#[repr(u8)]
pub enum InputRequirement {
    /// Consumes a scalar time series.
    ScalarInput = 1,
    /// Consumes level-1 quotes.
    QuoteInput = 2,
    /// Consumes OHLCV bars.
    BarInput = 3,
    /// Consumes individual trades.
    TradeInput = 4,
}

impl InputRequirement {
    pub fn as_str(&self) -> &'static str {
        match self {
            Self::ScalarInput => "scalar",
            Self::QuoteInput => "quote",
            Self::BarInput => "bar",
            Self::TradeInput => "trade",
        }
    }
}

impl std::fmt::Display for InputRequirement {
    fn fmt(&self, f: &mut std::fmt::Formatter<'_>) -> std::fmt::Result {
        f.write_str(self.as_str())
    }
}
