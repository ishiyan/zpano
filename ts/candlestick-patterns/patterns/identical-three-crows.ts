/** Identical Three Crows pattern (3-candle bearish). */
import { CandlestickPatternsEngine } from '../core/engine.ts';
import { tProductAll } from '../../fuzzy/index.ts';
import { lowerShadow, isBlack } from '../core/primitives.ts';

/**
 * Identical Three Crows: a three-candle bearish pattern.
 *
 * Must have:
 * - three consecutive declining black candles,
 * - each opens very close to the prior candle's close (equal criterion),
 * - very short lower shadows.
 *
 * The meaning of "equal" is specified with `cp.equal`.
 * The meaning of "very short" for shadows is specified with
 * `cp.veryShortShadow`.
 *
 * Category A: always bearish (continuous).
 *
 * Returns:
 *     Continuous float in [-100, 0].  Always bearish.
 */
export function identicalThreeCrows(cp: CandlestickPatternsEngine): number {
    if (!cp.enough(3, cp.equal, cp.veryShortShadow)) return 0.0;

    const b1 = cp.bar(3);
    const b2 = cp.bar(2);
    const b3 = cp.bar(1);

    // Crisp gates: all black, declining closes.
    if (!(isBlack(b1.o, b1.c) && isBlack(b2.o, b2.c) && isBlack(b3.o, b3.c))) return 0.0;
    if (!(b1.c > b2.c && b2.c > b3.c)) return 0.0;

    // Fuzzy conditions.
    const muLS1 = cp.muLessCS(lowerShadow(b1.o, b1.l, b1.c), cp.veryShortShadow, 3);
    const muLS2 = cp.muLessCS(lowerShadow(b2.o, b2.l, b2.c), cp.veryShortShadow, 2);
    const muLS3 = cp.muLessCS(lowerShadow(b3.o, b3.l, b3.c), cp.veryShortShadow, 1);
    // Opens near prior close (equal criterion, two-sided band).
    const muEq2 = cp.muLessCS(Math.abs(b2.o - b1.c), cp.equal, 3);
    const muEq3 = cp.muLessCS(Math.abs(b3.o - b2.c), cp.equal, 2);

    return -tProductAll(muLS1, muLS2, muLS3, muEq2, muEq3) * 100.0;
}
