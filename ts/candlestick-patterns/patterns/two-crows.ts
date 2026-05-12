/** Two Crows pattern (3-candle bearish). */
import { CandlestickPatternsEngine } from '../core/engine.ts';
import { realBody, isWhite, isBlack, isRealBodyGapUp } from '../core/primitives.ts';

/**
 * Two Crows: a three-candle bearish pattern.
 *
 * Must have:
 * - first candle: long white,
 * - second candle: black, gaps up (real body gap up from the first),
 * - third candle: black, opens within the second candle's real body,
 *   closes within the first candle's real body.
 *
 * The meaning of "long" is specified with `longBody`.
 *
 * Category A: always bearish (continuous).
 *
 * Returns:
 *     Continuous float in [-100, 0].  Always bearish.
 */
export function twoCrows(cp: CandlestickPatternsEngine): number {
    if (!cp.enough(3, cp.longBody)) return 0.0;

    const b1 = cp.bar(3);
    const b2 = cp.bar(2);
    const b3 = cp.bar(1);

    // Crisp gates: color checks.
    if (!(isWhite(b1.o, b1.c) && isBlack(b2.o, b2.c) && isBlack(b3.o, b3.c))) return 0.0;
    // Crisp: gap up.
    if (!isRealBodyGapUp(b1.o, b1.c, b2.o, b2.c)) return 0.0;
    // Crisp: third opens within second body (o3 < o2 and o3 > c2).
    if (!(b3.o < b2.o && b3.o > b2.c)) return 0.0;
    // Crisp: third closes within first body (c3 > o1 and c3 < c1).
    if (!(b3.c > b1.o && b3.c < b1.c)) return 0.0;

    // Fuzzy: first candle is long.
    return -cp.muGreaterCS(realBody(b1.o, b1.c), cp.longBody, 3) * 100.0;
}
