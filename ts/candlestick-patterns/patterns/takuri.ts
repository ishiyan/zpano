/** Takuri pattern (1-candle). */
import { CandlestickPatternsEngine } from '../core/engine.ts';
import { tProductAll } from '../../fuzzy/index.ts';
import { realBody, upperShadow, lowerShadow } from '../core/primitives.ts';

/**
 * Takuri (Dragonfly Doji with very long lower shadow): a one-candle pattern.
 *
 * A doji body with a very short upper shadow and a very long lower shadow.
 *
 * Returns:
 *     Continuous float in [0, 100].  Higher = stronger signal.
 */
export function takuri(cp: CandlestickPatternsEngine): number {
    if (!cp.enough(1, cp.dojiBody, cp.veryShortShadow, cp.veryLongShadow)) return 0.0;

    const b = cp.bar(1);

    const muDoji = cp.muLessCS(realBody(b.o, b.c), cp.dojiBody, 1);
    const muShortUS = cp.muLessCS(upperShadow(b.o, b.h, b.c), cp.veryShortShadow, 1);
    const muLongLS = cp.muGreaterCS(lowerShadow(b.o, b.l, b.c), cp.veryLongShadow, 1);

    return tProductAll(muDoji, muShortUS, muLongLS) * 100.0;
}
