/** Shooting Star pattern (2-candle bearish). */
import { CandlestickPatternsEngine } from '../core/engine.ts';
import { tProductAll } from '../../fuzzy/index.ts';
import { realBody, upperShadow, lowerShadow, isRealBodyGapUp } from '../core/primitives.ts';

/**
 * Shooting Star: a two-candle bearish reversal pattern.
 *
 * Must have:
 * - gap up from the previous candle (real body gap up),
 * - small real body,
 * - long upper shadow,
 * - very short lower shadow.
 *
 * Returns:
 *     Continuous float in [-100, 0].  More negative = stronger signal.
 */
export function shootingStar(cp: CandlestickPatternsEngine): number {
    if (!cp.enough(2, cp.shortBody, cp.longShadow, cp.veryShortShadow)) return 0.0;

    const b1 = cp.bar(2);
    const b2 = cp.bar(1);

    if (!isRealBodyGapUp(b1.o, b1.c, b2.o, b2.c)) return 0.0;

    const muShort = cp.muLessCS(realBody(b2.o, b2.c), cp.shortBody, 1);
    const muLongUS = cp.muGreaterCS(upperShadow(b2.o, b2.h, b2.c), cp.longShadow, 1);
    const muShortLS = cp.muLessCS(lowerShadow(b2.o, b2.l, b2.c), cp.veryShortShadow, 1);

    return -tProductAll(muShort, muLongUS, muShortLS) * 100.0;
}
