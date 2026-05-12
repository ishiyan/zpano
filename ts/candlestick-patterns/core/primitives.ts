import { RangeEntity } from './range-entity.ts';

// ---------------------------------------------------------------------------
// Color
// ---------------------------------------------------------------------------

/** Returns true when a candlestick is white (bullish): close >= open. */
export function isWhite(o: number, c: number): boolean {
    return c >= o;
}

/** Returns true when a candlestick is black (bearish): close < open. */
export function isBlack(o: number, c: number): boolean {
    return c < o;
}

// ---------------------------------------------------------------------------
// Real body
// ---------------------------------------------------------------------------

/** Returns the absolute length of the real body. */
export function realBody(o: number, c: number): number {
    if (c >= o) {
        return c - o;
    }
    return o - c;
}

/** Returns the length of the real body of a white candlestick (close - open). */
export function whiteRealBody(o: number, c: number): number {
    return c - o;
}

/** Returns the length of the real body of a black candlestick (open - close). */
export function blackRealBody(o: number, c: number): number {
    return o - c;
}

// ---------------------------------------------------------------------------
// Shadows
// ---------------------------------------------------------------------------

/** Returns the length of the upper shadow. */
export function upperShadow(o: number, h: number, c: number): number {
    if (c >= o) {
        return h - c;
    }
    return h - o;
}

/** Returns the length of the lower shadow. */
export function lowerShadow(o: number, l: number, c: number): number {
    if (c >= o) {
        return o - l;
    }
    return c - l;
}

/** Returns the length of the upper shadow of a white candlestick. */
export function whiteUpperShadow(h: number, c: number): number {
    return h - c;
}

/** Returns the length of the upper shadow of a black candlestick. */
export function blackUpperShadow(o: number, h: number): number {
    return h - o;
}

/** Returns the length of the lower shadow of a white candlestick. */
export function whiteLowerShadow(o: number, l: number): number {
    return o - l;
}

/** Returns the length of the lower shadow of a black candlestick. */
export function blackLowerShadow(l: number, c: number): number {
    return c - l;
}

// ---------------------------------------------------------------------------
// Gap tests
// ---------------------------------------------------------------------------

/** Returns true when max(open1, close1) < min(open2, close2). */
export function isRealBodyGapUp(o1: number, c1: number, o2: number, c2: number): boolean {
    return Math.max(o1, c1) < Math.min(o2, c2);
}

/** Returns true when min(open1, close1) > max(open2, close2). */
export function isRealBodyGapDown(o1: number, c1: number, o2: number, c2: number): boolean {
    return Math.min(o1, c1) > Math.max(o2, c2);
}

/** Returns true when high of first candle < low of second candle. */
export function isHighLowGapUp(h1: number, l2: number): boolean {
    return h1 < l2;
}

/** Returns true when low of first candle > high of second candle. */
export function isHighLowGapDown(l1: number, h2: number): boolean {
    return l1 > h2;
}

// ---------------------------------------------------------------------------
// Enclosure tests
// ---------------------------------------------------------------------------

/** Returns true when the real body of candle 1 completely encloses the real body of candle 2. */
export function isRealBodyEnclosesRealBody(o1: number, c1: number, o2: number, c2: number): boolean {
    let min1: number, max1: number;
    if (c1 > o1) {
        min1 = o1;
        max1 = c1;
    } else {
        min1 = c1;
        max1 = o1;
    }
    let min2: number, max2: number;
    if (c2 > o2) {
        min2 = o2;
        max2 = c2;
    } else {
        min2 = c2;
        max2 = o2;
    }
    return max1 > max2 && min1 < min2;
}

/** Returns true when the real body of candle 1 encloses the open of candle 2. */
export function isRealBodyEnclosesOpen(o1: number, c1: number, o2: number): boolean {
    if (o1 > c1) {
        return o2 < o1 && o2 > c1;
    }
    return o2 > o1 && o2 < c1;
}

/** Returns true when the real body of candle 1 encloses the close of candle 2. */
export function isRealBodyEnclosesClose(o1: number, c1: number, c2: number): boolean {
    if (o1 > c1) {
        return c2 < o1 && c2 > c1;
    }
    return c2 > o1 && c2 < c1;
}

// ---------------------------------------------------------------------------
// Misc comparisons
// ---------------------------------------------------------------------------

/** Returns true when high of candle 1 > close of candle 2. */
export function isHighExceedsClose(h1: number, c2: number): boolean {
    return h1 > c2;
}

/** Returns true when candle 1 opens within the real body of candle 2 (with optional tolerance). */
export function isOpensWithin(o1: number, o2: number, c2: number, tolerance: number): boolean {
    return o1 >= Math.min(o2, c2) - tolerance && o1 <= Math.max(o2, c2) + tolerance;
}

// ---------------------------------------------------------------------------
// Range value for a single candle (used by Criterion)
// ---------------------------------------------------------------------------

/** Computes the range value of a candle for a given RangeEntity type. */
export function candleRangeValue(entity: RangeEntity, o: number, h: number, l: number, c: number): number {
    switch (entity) {
        case RangeEntity.REAL_BODY:
            if (c >= o) {
                return c - o;
            }
            return o - c;
        case RangeEntity.HIGH_LOW:
            return h - l;
        default:
            // SHADOWS: average of upper and lower shadow
            if (c >= o) {
                return (h - c + o - l) / 2.0;
            }
            return (h - o + c - l) / 2.0;
    }
}
