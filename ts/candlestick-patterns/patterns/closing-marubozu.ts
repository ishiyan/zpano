/** Closing Marubozu pattern (1-candle). */
import { CandlestickPatternsEngine } from '../core/engine.ts';
import { tProductAll } from '../../fuzzy/index.ts';
import { realBody, upperShadow, lowerShadow, isWhite } from '../core/primitives.ts';

/**
 * Closing Marubozu: a one-candle pattern.
 *
 * A long candle with a very short shadow on the closing side:
 * - bullish (white): very short upper shadow,
 * - bearish (black): very short lower shadow.
 *
 * The meaning of "long" is specified with longBody.
 * The meaning of "very short" for shadows is specified with
 * veryShortShadow.
 *
 * Category C: both branches evaluated, return stronger signal.
 *
 * Returns:
 *     Continuous float in [-100, +100].
 */
export function closingMarubozu(cp: CandlestickPatternsEngine): number {
    if (!cp.enough(1, cp.longBody, cp.veryShortShadow)) return 0.0;

    const b = cp.bar(1);
    const muLong = cp.muGreaterCS(realBody(b.o, b.c), cp.longBody, 1);

    // Bullish: white + very short upper shadow.
    let bullSignal = 0.0;
    if (isWhite(b.o, b.c)) {
        const muVS = cp.muLessCS(upperShadow(b.o, b.h, b.c), cp.veryShortShadow, 1);
        bullSignal = tProductAll(muLong, muVS) * 100.0;
    }

    // Bearish: black (not white) + very short lower shadow.
    let bearSignal = 0.0;
    if (!isWhite(b.o, b.c)) {
        const muVS = cp.muLessCS(lowerShadow(b.o, b.l, b.c), cp.veryShortShadow, 1);
        bearSignal = -tProductAll(muLong, muVS) * 100.0;
    }

    return Math.abs(bullSignal) >= Math.abs(bearSignal) ? bullSignal : bearSignal;
}
