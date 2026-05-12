/** Concealing Baby Swallow pattern (4-candle bullish). */
import { CandlestickPatternsEngine } from '../core/engine.ts';
import { tProductAll } from '../../fuzzy/index.ts';
import { lowerShadow, upperShadow, isBlack, isRealBodyGapDown } from '../core/primitives.ts';

/**
 * Concealing Baby Swallow: a four-candle bullish pattern.
 *
 * Must have:
 * - first candle: black marubozu (very short shadows),
 * - second candle: black marubozu (very short shadows),
 * - third candle: black, opens gapping down, upper shadow extends into
 *   the prior real body (upper shadow > very-short avg),
 * - fourth candle: black, completely engulfs the third candle including
 *   shadows (strict > / <).
 *
 * The meaning of "very short" for shadows is specified with
 * veryShortShadow.
 *
 * Category A: always bullish (continuous).
 *
 * Returns:
 *     Continuous float in [0, 100]. Always bullish.
 */
export function concealingBabySwallow(cp: CandlestickPatternsEngine): number {
    if (!cp.enough(4, cp.veryShortShadow)) return 0.0;

    const b1 = cp.bar(4);
    const b2 = cp.bar(3);
    const b3 = cp.bar(2);
    const b4 = cp.bar(1);

    // Crisp gates: all black.
    if (!(isBlack(b1.o, b1.c) && isBlack(b2.o, b2.c) &&
        isBlack(b3.o, b3.c) && isBlack(b4.o, b4.c))) return 0.0;
    // Crisp: gap down and upper shadow extends into prior body.
    if (!(isRealBodyGapDown(b2.o, b2.c, b3.o, b3.c) && b3.h > b2.c)) return 0.0;
    // Crisp: fourth engulfs third including shadows (strict).
    if (!(b4.h > b3.h && b4.l < b3.l)) return 0.0;

    // Fuzzy: first and second are marubozu (very short shadows).
    const muLS1 = cp.muLessCS(lowerShadow(b1.o, b1.l, b1.c), cp.veryShortShadow, 4);
    const muUS1 = cp.muLessCS(upperShadow(b1.o, b1.h, b1.c), cp.veryShortShadow, 4);
    const muLS2 = cp.muLessCS(lowerShadow(b2.o, b2.l, b2.c), cp.veryShortShadow, 3);
    const muUS2 = cp.muLessCS(upperShadow(b2.o, b2.h, b2.c), cp.veryShortShadow, 3);
    // Fuzzy: third candle upper shadow > very-short avg.
    const muUS3Long = cp.muGreaterCS(upperShadow(b3.o, b3.h, b3.c), cp.veryShortShadow, 2);

    return tProductAll(muLS1, muUS1, muLS2, muUS2, muUS3Long) * 100.0;
}
