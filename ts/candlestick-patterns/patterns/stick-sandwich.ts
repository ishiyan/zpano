/** Stick Sandwich pattern (3-candle bullish). */
import { CandlestickPatternsEngine } from '../core/engine.ts';
import { isBlack, isWhite } from '../core/primitives.ts';

/**
 * Stick Sandwich: a three-candle bullish pattern.
 *
 * Must have:
 * - first candle: black,
 * - second candle: white, trades above the first candle's close
 *   (low > first close),
 * - third candle: black, close equals the first candle's close.
 *
 * The meaning of "equal" is specified with `equal`.
 *
 * Category A: always bullish (continuous).
 *
 * Returns:
 *     Continuous float in [0, 100].  Always bullish.
 */
export function stickSandwich(cp: CandlestickPatternsEngine): number {
    if (!cp.enough(3, cp.equal)) return 0.0;

    const b1 = cp.bar(3);
    const b2 = cp.bar(2);
    const b3 = cp.bar(1);

    // Crisp gates: colors and gap.
    if (!(isBlack(b1.o, b1.c) && isWhite(b2.o, b2.c) && isBlack(b3.o, b3.c) && b2.l > b1.c)) return 0.0;

    // Fuzzy: third close equals first close (two-sided band).
    return cp.muLessCS(Math.abs(b3.c - b1.c), cp.equal, 3) * 100.0;
}
