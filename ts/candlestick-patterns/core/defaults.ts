/* Default criterion definitions matching the Ta-Lib implementation. */

import { RangeEntity } from './range-entity.ts';
import { Criterion } from './criterion.ts';

/** Real body is long when it is longer than the average of the real body of the 10 previous candlesticks. */
export const DEFAULT_LONG_BODY = new Criterion(RangeEntity.REAL_BODY, 10, 1.0);

/** Real body is very long when it is longer than 3 times the average of the real body of the 10 previous candlesticks. */
export const DEFAULT_VERY_LONG_BODY = new Criterion(RangeEntity.REAL_BODY, 10, 3.0);

/** Real body is short when it is shorter than the average of the real body of the 10 previous candlesticks. */
export const DEFAULT_SHORT_BODY = new Criterion(RangeEntity.REAL_BODY, 10, 1.0);

/** Real body is like doji when it is shorter than 10% the average of the high-low range of the 10 previous candlesticks. */
export const DEFAULT_DOJI_BODY = new Criterion(RangeEntity.HIGH_LOW, 10, 0.1);

/** Shadow is long when it is longer than the real body. */
export const DEFAULT_LONG_SHADOW = new Criterion(RangeEntity.REAL_BODY, 0, 1.0);

/** Shadow is very long when it is longer than 2 times the real body. */
export const DEFAULT_VERY_LONG_SHADOW = new Criterion(RangeEntity.REAL_BODY, 0, 2.0);

/** Shadow is short when it is shorter than the average of the sum of shadows of the 10 previous candlesticks. */
export const DEFAULT_SHORT_SHADOW = new Criterion(RangeEntity.SHADOWS, 10, 1.0);

/** Shadow is very short when it is shorter than 10% the average of the high-low range of the 10 previous candlesticks. */
export const DEFAULT_VERY_SHORT_SHADOW = new Criterion(RangeEntity.HIGH_LOW, 10, 0.1);

/** 'Near' means <= 20% of the average of the high-low range of the 5 previous candlesticks. */
export const DEFAULT_NEAR = new Criterion(RangeEntity.HIGH_LOW, 5, 0.2);

/** 'Far' means >= 60% of the average of the high-low range of the 5 previous candlesticks. */
export const DEFAULT_FAR = new Criterion(RangeEntity.HIGH_LOW, 5, 0.6);

/** 'Equal' means <= 5% of the average of the high-low range of the 5 previous candlesticks. */
export const DEFAULT_EQUAL = new Criterion(RangeEntity.HIGH_LOW, 5, 0.05);
