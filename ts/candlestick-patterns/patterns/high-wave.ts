/** High Wave pattern (1-candle). */
import { CandlestickPatternsEngine } from '../core/engine.ts';
import { tProductAll } from '../../fuzzy/index.ts';
import { realBody, upperShadow, lowerShadow, isWhite } from '../core/primitives.ts';

/**
 * High Wave: a one-candle pattern.
 *
 * Must have:
 * - short real body,
 * - very long upper shadow,
 * - very long lower shadow.
 *
 * The meaning of "short" is specified with shortBody.
 * The meaning of "very long" (shadow) is specified with veryLongShadow.
 *
 * Category C: color determines sign.
 *
 * Returns:
 *     Continuous number in [-100, +100].
 */
export function highWave(cp: CandlestickPatternsEngine): number {
    if (!cp.enough(1, cp.shortBody, cp.veryLongShadow)) return 0.0;

    const b = cp.bar(1);
    const muShort = cp.muLessCS(realBody(b.o, b.c), cp.shortBody, 1);
    const muLongUS = cp.muGreaterCS(upperShadow(b.o, b.h, b.c), cp.veryLongShadow, 1);
    const muLongLS = cp.muGreaterCS(lowerShadow(b.o, b.l, b.c), cp.veryLongShadow, 1);

    const confidence = tProductAll(muShort, muLongUS, muLongLS);
    if (isWhite(b.o, b.c)) return confidence * 100.0;
    return -confidence * 100.0;
}
