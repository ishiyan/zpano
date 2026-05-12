/** Upside Gap Two Crows pattern (3-candle bearish). */
import { CandlestickPatternsEngine } from '../core/engine.ts';
import { tProductAll } from '../../fuzzy/index.ts';
import { realBody, isWhite, isBlack, isRealBodyGapUp } from '../core/primitives.ts';

/**
 * Upside Gap Two Crows: a three-candle bearish pattern.
 *
 * Must have:
 * - first candle: long white,
 * - second candle: small black that gaps up from the first,
 * - third candle: black that engulfs the second candle's body and
 *   closes above the first candle's close (gap not filled).
 *
 * The meaning of "long" is specified with `longBody`.
 * The meaning of "short" is specified with `shortBody`.
 *
 * Category A: always bearish (continuous).
 *
 * Returns:
 *     Continuous float in [-100, 0].  Always bearish.
 */
export function upsideGapTwoCrows(cp: CandlestickPatternsEngine): number {
    if (!cp.enough(3, cp.longBody, cp.shortBody)) return 0.0;

    const b1 = cp.bar(3);
    const b2 = cp.bar(2);
    const b3 = cp.bar(1);

    // Crisp gates: colors.
    if (!(isWhite(b1.o, b1.c) && isBlack(b2.o, b2.c) && isBlack(b3.o, b3.c))) return 0.0;
    // Crisp: gap up from first to second.
    if (!isRealBodyGapUp(b1.o, b1.c, b2.o, b2.c)) return 0.0;
    // Crisp: third engulfs second (o3 > o2 and c3 < c2) and closes above c1.
    if (!(b3.o > b2.o && b3.c < b2.c && b3.c > b1.c)) return 0.0;

    // Fuzzy: first candle is long.
    const muLong1 = cp.muGreaterCS(realBody(b1.o, b1.c), cp.longBody, 3);
    // Fuzzy: second candle is short.
    const muShort2 = cp.muLessCS(realBody(b2.o, b2.c), cp.shortBody, 2);

    return -tProductAll(muLong1, muShort2) * 100.0;
}
