"""Default criterion definitions matching the Ta-Lib implementation."""

from .range_entity import RangeEntity
from .criterion import Criterion

# Real body is long when it is longer than the average of the real body
# of the 10 previous candlesticks.
DEFAULT_LONG_BODY = Criterion(RangeEntity.REAL_BODY, 10, 1.0)

# Real body is very long when it is longer than 3 times the average
# of the real body of the 10 previous candlesticks.
DEFAULT_VERY_LONG_BODY = Criterion(RangeEntity.REAL_BODY, 10, 3.0)

# Real body is short when it is shorter than the average of the real body
# of the 10 previous candlesticks.
DEFAULT_SHORT_BODY = Criterion(RangeEntity.REAL_BODY, 10, 1.0)

# Real body is like doji when it is shorter than 10% the average of the
# high-low range of the 10 previous candlesticks.
DEFAULT_DOJI_BODY = Criterion(RangeEntity.HIGH_LOW, 10, 0.1)

# Shadow is long when it is longer than the real body.
DEFAULT_LONG_SHADOW = Criterion(RangeEntity.REAL_BODY, 0, 1.0)

# Shadow is very long when it is longer than 2 times the real body.
DEFAULT_VERY_LONG_SHADOW = Criterion(RangeEntity.REAL_BODY, 0, 2.0)

# Shadow is short when it is shorter than the average of the sum
# of shadows of the 10 previous candlesticks.
DEFAULT_SHORT_SHADOW = Criterion(RangeEntity.SHADOWS, 10, 1.0)

# Shadow is very short when it is shorter than 10% the average of the
# high-low range of the 10 previous candlesticks.
DEFAULT_VERY_SHORT_SHADOW = Criterion(RangeEntity.HIGH_LOW, 10, 0.1)

# When measuring distance between parts of candles or width of gaps,
# 'near' means <= 20% of the average of the high-low range of the 5 previous candlesticks.
DEFAULT_NEAR = Criterion(RangeEntity.HIGH_LOW, 5, 0.2)

# When measuring distance between parts of candles or width of gaps,
# 'far' means >= 60% of the average of the high-low range of the 5 previous candlesticks.
DEFAULT_FAR = Criterion(RangeEntity.HIGH_LOW, 5, 0.6)

# When measuring distance between parts of candles or width of gaps,
# 'equal' means <= 5% of the average of the high-low range of the 5 previous candlesticks.
DEFAULT_EQUAL = Criterion(RangeEntity.HIGH_LOW, 5, 0.05)
