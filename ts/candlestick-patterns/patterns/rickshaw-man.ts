/** Rickshaw Man pattern (1-candle). */
import { CandlestickPatternsEngine } from '../core/engine.ts';
import { tProductAll } from '../../fuzzy/index.ts';
import { realBody, upperShadow, lowerShadow } from '../core/primitives.ts';

/**
 * Rickshaw Man: a one-candle doji pattern.
 *
 * Must have:
 * - doji body (very small real body),
 * - two long shadows,
 * - body near the midpoint of the high-low range.
 *
 * Returns:
 *     Continuous float in [0, 100].  Higher = stronger signal.
 */
export function rickshawMan(cp: CandlestickPatternsEngine): number {
    if (!cp.enough(1, cp.dojiBody, cp.longShadow, cp.near)) return 0.0;

    const b = cp.bar(1);

    const hlRange = b.h - b.l;
    const nearAvg = cp.avgCS(cp.near, 1);
    let nearWidth = 0.0;
    if (nearAvg > 0.0) nearWidth = cp.fuzzRatio * nearAvg;

    const muDoji = cp.muLessCS(realBody(b.o, b.c), cp.dojiBody, 1);
    const muLongUS = cp.muGreaterCS(upperShadow(b.o, b.h, b.c), cp.longShadow, 1);
    const muLongLS = cp.muGreaterCS(lowerShadow(b.o, b.l, b.c), cp.longShadow, 1);
    const midpoint = b.l + hlRange / 2.0;
    const muNearMidLo = cp.muLtRaw(Math.min(b.o, b.c), midpoint + nearAvg, nearWidth);
    const muNearMidHi = cp.muGeRaw(Math.max(b.o, b.c), midpoint - nearAvg, nearWidth);

    return tProductAll(muDoji, muLongUS, muLongLS, muNearMidLo, muNearMidHi) * 100.0;
}
