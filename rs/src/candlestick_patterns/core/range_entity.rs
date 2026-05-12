/// RangeEntity specifies which part of a candlestick to measure.
#[derive(Debug, Clone, Copy, PartialEq, Eq)]
#[repr(u8)]
pub enum RangeEntity {
    /// The absolute difference between open and close.
    RealBody = 0,
    /// The difference between high and low.
    HighLow = 1,
    /// The average of upper and lower shadows.
    Shadows = 2,
}
