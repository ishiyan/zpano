/** Up/Down-Gap Side-By-Side White Lines pattern (3-candle). */
import { CandlestickPatternsEngine } from '../core/engine.ts';
import { tProductAll } from '../../fuzzy/index.ts';
import { realBody, isWhite, isRealBodyGapUp, isRealBodyGapDown } from '../core/primitives.ts';

/**
 * Up/Down-Gap Side-By-Side White Lines: a three-candle pattern.
 *
 * Must have:
 * - first candle: white (for up gap) or black (for down gap),
 * - gap (up or down) between the first and second candle -- both 2nd AND
 *   3rd must gap from the 1st,
 * - second and third candles are both white with similar size and
 *   approximately the same open.
 *
 * Up gap = bullish continuation, down gap = bearish continuation.
 *
 * Category C: both branches evaluated, return stronger signal.
 *
 * Returns:
 *     Continuous float in [-100, +100].
 */
export function upDownGapSideBySideWhiteLines(cp: CandlestickPatternsEngine): number {
    if (!cp.enough(3, cp.near, cp.equal)) return 0.0;

    const b1 = cp.bar(3);
    const b2 = cp.bar(2);
    const b3 = cp.bar(1);

    // Crisp: both 2nd and 3rd must be white.
    if (!(isWhite(b2.o, b2.c) && isWhite(b3.o, b3.c))) return 0.0;

    // Both 2nd and 3rd must gap from 1st in the same direction — crisp.
    const gapUp = isRealBodyGapUp(b1.o, b1.c, b2.o, b2.c) && isRealBodyGapUp(b1.o, b1.c, b3.o, b3.c);
    const gapDown = isRealBodyGapDown(b1.o, b1.c, b2.o, b2.c) && isRealBodyGapDown(b1.o, b1.c, b3.o, b3.c);

    if (!(gapUp || gapDown)) return 0.0;

    const rb2 = realBody(b2.o, b2.c);
    const rb3 = realBody(b3.o, b3.c);

    // Fuzzy: similar size and same open.
    const muNearSize = cp.muLessCS(Math.abs(rb2 - rb3), cp.near, 2);
    const muEqualOpen = cp.muLessCS(Math.abs(b3.o - b2.o), cp.equal, 2);

    const conf = tProductAll(muNearSize, muEqualOpen);

    if (gapUp) return conf * 100.0;
    return -conf * 100.0;
}
