/** In Neck pattern (2-candle bearish continuation). */
import { CandlestickPatternsEngine } from '../core/engine.ts';
import { tProductAll } from '../../fuzzy/index.ts';
import { realBody, isBlack, isWhite } from '../core/primitives.ts';

/**
 * In Neck: a two-candle bearish continuation pattern.
 *
 * Must have:
 * - first candle: long black,
 * - second candle: white, opens below the prior low, closes slightly
 *   into the prior real body (close near the prior close).
 *
 * The meaning of "long" is specified with `cp.longBody`.
 * The meaning of "near" is specified with `cp.near`.
 *
 * Category A: always bearish (continuous).
 *
 * Returns:
 *     Continuous float in [-100, 0].  Always bearish.
 */
export function inNeck(cp: CandlestickPatternsEngine): number {
    if (!cp.enough(2, cp.longBody, cp.near)) return 0.0;

    const b1 = cp.bar(2);
    const b2 = cp.bar(1);

    // Crisp gates: color checks and open below prior low.
    if (!(isBlack(b1.o, b1.c) && isWhite(b2.o, b2.c) && b2.o < b1.l)) return 0.0;

    // Fuzzy conditions.
    const muLong1 = cp.muGreaterCS(realBody(b1.o, b1.c), cp.longBody, 2);
    // Close near prior close: crisp was abs(c2-c1) < nearAvg.
    // Model as muLess(absDiff, nearAvg) — crossover at near boundary.
    const muNearClose = cp.muLessCS(Math.abs(b2.c - b1.c), cp.near, 1);

    return -tProductAll(muLong1, muNearClose) * 100.0;
}
