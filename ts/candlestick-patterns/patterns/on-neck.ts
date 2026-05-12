/** On Neck pattern (2-candle bearish). */
import { CandlestickPatternsEngine } from '../core/engine.ts';
import { tProductAll } from '../../fuzzy/index.ts';
import { realBody, isBlack, isWhite } from '../core/primitives.ts';

/**
 * On Neck: a two-candle bearish continuation pattern.
 *
 * Must have:
 * - first candle: long black,
 * - second candle: white that opens below the prior low and closes
 *   equal to the prior candle's low.
 *
 * The meaning of "long" is specified with `longBody`.
 * The meaning of "equal" is specified with `equal`.
 *
 * Category A: always bearish (continuous).
 *
 * Returns:
 *     Continuous float in [-100, 0].  Always bearish.
 */
export function onNeck(cp: CandlestickPatternsEngine): number {
    if (!cp.enough(2, cp.longBody, cp.equal)) return 0.0;

    const b1 = cp.bar(2);
    const b2 = cp.bar(1);

    // Crisp gates: color checks and open below prior low.
    if (!(isBlack(b1.o, b1.c) && isWhite(b2.o, b2.c) && b2.o < b1.l)) return 0.0;

    // Fuzzy conditions.
    const muLong1 = cp.muGreaterCS(realBody(b1.o, b1.c), cp.longBody, 2);

    // Close equal to prior low: model as muLess(absDiff, eqAvg) — crossover at eq boundary.
    const muNearLow = cp.muLessCS(Math.abs(b2.c - b1.l), cp.equal, 2);

    return -tProductAll(muLong1, muNearLow) * 100.0;
}
