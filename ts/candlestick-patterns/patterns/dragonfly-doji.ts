/** Dragonfly Doji pattern (1-candle). */
import { CandlestickPatternsEngine } from '../core/engine.ts';
import { tProductAll } from '../../fuzzy/index.ts';
import { realBody, upperShadow, lowerShadow } from '../core/primitives.ts';

/**
 * Dragonfly Doji: a one-candle pattern.
 *
 * Must have:
 * - doji body (very small real body relative to high-low range),
 * - no or very short upper shadow,
 * - lower shadow is not very short.
 *
 * Returns:
 *   Continuous float in [0, 100].  Higher = stronger signal.
 */
export function dragonflyDoji(cp: CandlestickPatternsEngine): number {
    if (!cp.enough(1, cp.dojiBody, cp.veryShortShadow)) return 0.0;

    const b = cp.bar(1);
    const muDoji = cp.muLessCS(realBody(b.o, b.c), cp.dojiBody, 1);
    const muShortUS = cp.muLessCS(upperShadow(b.o, b.h, b.c), cp.veryShortShadow, 1);
    const muLongLS = cp.muGreaterCS(lowerShadow(b.o, b.l, b.c), cp.veryShortShadow, 1);

    return tProductAll(muDoji, muShortUS, muLongLS) * 100.0;
}
