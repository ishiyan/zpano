/** Matching Low pattern (2-candle bullish). */
import { CandlestickPatternsEngine } from '../core/engine.ts';
import { isBlack } from '../core/primitives.ts';

/**
 * Matching Low: a two-candle bullish pattern.
 *
 * Must have:
 * - first candle: black,
 * - second candle: black with close equal to the first candle's close.
 *
 * The meaning of "equal" is specified with `equal`.
 *
 * Category A: always bullish (continuous).
 *
 * Returns:
 *     Continuous float in [0, 100].  Always bullish.
 */
export function matchingLow(cp: CandlestickPatternsEngine): number {
    if (!cp.enough(2, cp.equal)) return 0.0;

    const b1 = cp.bar(2);
    const b2 = cp.bar(1);

    // Crisp gates: both black.
    if (!(isBlack(b1.o, b1.c) && isBlack(b2.o, b2.c))) return 0.0;

    // Fuzzy: close equal to prior close (two-sided band).
    return cp.muLessCS(Math.abs(b2.c - b1.c), cp.equal, 2) * 100.0;
}
