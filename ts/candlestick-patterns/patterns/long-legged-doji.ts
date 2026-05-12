/** Long Legged Doji pattern (1-candle). */
import { CandlestickPatternsEngine } from '../core/engine.ts';
import { tProductAll, sMax } from '../../fuzzy/index.ts';
import { realBody, upperShadow, lowerShadow } from '../core/primitives.ts';

/**
 * Long Legged Doji: a one-candle pattern.
 *
 * Must have:
 * - doji body (very small real body),
 * - one or both shadows are long.
 *
 * Returns:
 *     Continuous float in [0, 100].  Higher = stronger signal.
 */
export function longLeggedDoji(cp: CandlestickPatternsEngine): number {
    if (!cp.enough(1, cp.dojiBody, cp.longShadow)) return 0.0;

    const b = cp.bar(1);
    const muDoji = cp.muLessCS(realBody(b.o, b.c), cp.dojiBody, 1);
    const muLongUS = cp.muGreaterCS(upperShadow(b.o, b.h, b.c), cp.longShadow, 1);
    const muLongLS = cp.muGreaterCS(lowerShadow(b.o, b.l, b.c), cp.longShadow, 1);
    const muAnyLong = sMax(muLongUS, muLongLS);

    return tProductAll(muDoji, muAnyLong) * 100.0;
}
