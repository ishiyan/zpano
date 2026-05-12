// Default criterion definitions matching the Ta-Lib implementation.

use super::criterion::Criterion;
use super::range_entity::RangeEntity;

/// Real body is long when it is longer than the average of the
/// real body of the 10 previous candlesticks.
pub const DEFAULT_LONG_BODY: Criterion = Criterion::new(RangeEntity::RealBody, 10, 1.0);

/// Real body is very long when it is longer than 3 times
/// the average of the real body of the 10 previous candlesticks.
pub const DEFAULT_VERY_LONG_BODY: Criterion = Criterion::new(RangeEntity::RealBody, 10, 3.0);

/// Real body is short when it is shorter than the average of
/// the real body of the 10 previous candlesticks.
pub const DEFAULT_SHORT_BODY: Criterion = Criterion::new(RangeEntity::RealBody, 10, 1.0);

/// Real body is like doji when it is shorter than 10% the
/// average of the high-low range of the 10 previous candlesticks.
pub const DEFAULT_DOJI_BODY: Criterion = Criterion::new(RangeEntity::HighLow, 10, 0.1);

/// Shadow is long when it is longer than the real body.
pub const DEFAULT_LONG_SHADOW: Criterion = Criterion::new(RangeEntity::RealBody, 0, 1.0);

/// Shadow is very long when it is longer than 2 times the real body.
pub const DEFAULT_VERY_LONG_SHADOW: Criterion = Criterion::new(RangeEntity::RealBody, 0, 2.0);

/// Shadow is short when it is shorter than the average of
/// the sum of shadows of the 10 previous candlesticks.
pub const DEFAULT_SHORT_SHADOW: Criterion = Criterion::new(RangeEntity::Shadows, 10, 1.0);

/// Shadow is very short when it is shorter than 10% the
/// average of the high-low range of the 10 previous candlesticks.
pub const DEFAULT_VERY_SHORT_SHADOW: Criterion = Criterion::new(RangeEntity::HighLow, 10, 0.1);

/// When measuring distance between parts of candles or width of gaps,
/// 'near' means <= 20% of the average of the high-low range of the 5 previous candlesticks.
pub const DEFAULT_NEAR: Criterion = Criterion::new(RangeEntity::HighLow, 5, 0.2);

/// When measuring distance between parts of candles or width of gaps,
/// 'far' means >= 60% of the average of the high-low range of the 5 previous candlesticks.
pub const DEFAULT_FAR: Criterion = Criterion::new(RangeEntity::HighLow, 5, 0.6);

/// When measuring distance between parts of candles or width of gaps,
/// 'equal' means <= 5% of the average of the high-low range of the 5 previous candlesticks.
pub const DEFAULT_EQUAL: Criterion = Criterion::new(RangeEntity::HighLow, 5, 0.05);
