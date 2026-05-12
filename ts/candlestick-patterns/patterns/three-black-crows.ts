/** Three Black Crows pattern (4-candle bearish). */
import { CandlestickPatternsEngine } from '../core/engine.ts';
import { tProductAll } from '../../fuzzy/index.ts';
import { lowerShadow, isWhite, isBlack } from '../core/primitives.ts';

/**
 * Three Black Crows: a four-candle bearish reversal pattern.
 *
 * Must have:
 * - preceding candle (oldest) is white,
 * - three consecutive black candles with declining closes,
 * - each opens within the prior black candle's real body,
 * - each has a very short lower shadow,
 * - 1st black closes under the prior white candle's high.
 *
 * Category A: always bearish (continuous).
 *
 * Returns:
 *     Continuous float in [-100, 0].  Always bearish.
 */
export function threeBlackCrows(cp: CandlestickPatternsEngine): number {
    if (!cp.enough(4, cp.veryShortShadow)) return 0.0;

    const b0 = cp.bar(4);  // prior white
    const b1 = cp.bar(3);  // 1st black
    const b2 = cp.bar(2);  // 2nd black
    const b3 = cp.bar(1);  // 3rd black

    // Crisp gates: colors, declining closes, opens within prior body.
    if (!isWhite(b0.o, b0.c)) return 0.0;
    if (!(isBlack(b1.o, b1.c) && isBlack(b2.o, b2.c) && isBlack(b3.o, b3.c))) return 0.0;
    if (!(b1.c > b2.c && b2.c > b3.c)) return 0.0;
    // Opens within prior black body (crisp containment for strict ordering).
    if (!(b2.o < b1.o && b2.o > b1.c && b3.o < b2.o && b3.o > b2.c)) return 0.0;
    // Prior white's high > 1st black's close (crisp).
    if (!(b0.h > b1.c)) return 0.0;

    // Fuzzy: very short lower shadows.
    const muLS1 = cp.muLessCS(lowerShadow(b1.o, b1.l, b1.c), cp.veryShortShadow, 3);
    const muLS2 = cp.muLessCS(lowerShadow(b2.o, b2.l, b2.c), cp.veryShortShadow, 2);
    const muLS3 = cp.muLessCS(lowerShadow(b3.o, b3.l, b3.c), cp.veryShortShadow, 1);

    return -tProductAll(muLS1, muLS2, muLS3) * 100.0;
}
