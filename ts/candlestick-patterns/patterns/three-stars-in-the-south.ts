/** Three Stars In The South pattern (3-candle bullish). */
import { CandlestickPatternsEngine } from '../core/engine.ts';
import { tProductAll } from '../../fuzzy/index.ts';
import { realBody, upperShadow, lowerShadow, isBlack } from '../core/primitives.ts';

/**
 * Three Stars In The South: a three-candle bullish pattern.
 *
 * Must have:
 * - all three candles are black,
 * - first candle: long body with long lower shadow,
 * - second candle: smaller body, opens within or above prior range,
 *   trades lower but its low does not go below the first candle's low,
 * - third candle: small marubozu (very short shadows) engulfed by the
 *   second candle's range.
 *
 * The meaning of "long" is specified with `longBody`.
 * The meaning of "short" is specified with `shortBody`.
 * The meaning of "long" for shadows is specified with `longShadow`.
 * The meaning of "very short" for shadows is specified with
 * `veryShortShadow`.
 *
 * Category A: always bullish (continuous).
 *
 * Returns:
 *     Continuous float in [0, 100].  Always bullish.
 */
export function threeStarsInTheSouth(cp: CandlestickPatternsEngine): number {
    if (!cp.enough(3, cp.longBody, cp.shortBody, cp.longShadow, cp.veryShortShadow)) return 0.0;

    const b1 = cp.bar(3);
    const b2 = cp.bar(2);
    const b3 = cp.bar(1);

    // Crisp gates: all black.
    if (!(isBlack(b1.o, b1.c) && isBlack(b2.o, b2.c) && isBlack(b3.o, b3.c))) return 0.0;

    const rb1 = realBody(b1.o, b1.c);
    const rb2 = realBody(b2.o, b2.c);

    // Crisp: second body smaller than first.
    if (!(rb2 < rb1)) return 0.0;
    // Crisp: second opens within or above prior range, low not below first's low.
    if (!(b2.o <= b1.h && b2.o >= b1.l && b2.l >= b1.l)) return 0.0;
    // Crisp: third engulfed by second's range.
    if (!(b3.h <= b2.h && b3.l >= b2.l)) return 0.0;

    // Fuzzy: first candle long body.
    const muLong1 = cp.muGreaterCS(rb1, cp.longBody, 3);
    // Fuzzy: first candle long lower shadow.
    const muLS1 = cp.muGreaterCS(lowerShadow(b1.o, b1.l, b1.c), cp.longShadow, 3);
    // Fuzzy: third candle short body.
    const muShort3 = cp.muLessCS(realBody(b3.o, b3.c), cp.shortBody, 1);
    // Fuzzy: third candle very short shadows (marubozu).
    const muVSUS3 = cp.muLessCS(upperShadow(b3.o, b3.h, b3.c), cp.veryShortShadow, 1);
    const muVSLS3 = cp.muLessCS(lowerShadow(b3.o, b3.l, b3.c), cp.veryShortShadow, 1);

    return tProductAll(muLong1, muLS1, muShort3, muVSUS3, muVSLS3) * 100.0;
}
