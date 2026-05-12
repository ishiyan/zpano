/** Gravestone Doji pattern (1-candle). */
import { CandlestickPatternsEngine } from '../core/engine.ts';
import { tProductAll } from '../../fuzzy/index.ts';
import { realBody, upperShadow, lowerShadow } from '../core/primitives.ts';

/**
 * Gravestone Doji: a one-candle pattern.
 *
 * Must have:
 * - doji body (very small real body relative to high-low range),
 * - no or very short lower shadow,
 * - upper shadow is not very short.
 *
 * Returns:
 *     Continuous number in [0, 100]. Higher = stronger signal.
 */
export function gravestoneDoji(cp: CandlestickPatternsEngine): number {
    if (!cp.enough(1, cp.dojiBody, cp.veryShortShadow)) return 0.0;

    const b = cp.bar(1);
    const muDoji = cp.muLessCS(realBody(b.o, b.c), cp.dojiBody, 1);
    const muShortLS = cp.muLessCS(lowerShadow(b.o, b.l, b.c), cp.veryShortShadow, 1);
    const muLongUS = cp.muGreaterCS(upperShadow(b.o, b.h, b.c), cp.veryShortShadow, 1);

    return tProductAll(muDoji, muShortLS, muLongUS) * 100.0;
}
