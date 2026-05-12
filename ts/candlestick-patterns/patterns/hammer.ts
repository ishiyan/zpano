/** Hammer pattern (2-candle bullish). */
import { CandlestickPatternsEngine } from '../core/engine.ts';
import { tProductAll } from '../../fuzzy/index.ts';
import { realBody, upperShadow, lowerShadow } from '../core/primitives.ts';

/**
 * Hammer: a two-candle bullish reversal pattern.
 *
 * Must have:
 * - small real body,
 * - long lower shadow,
 * - no or very short upper shadow,
 * - body is below or near the lows of the previous candle.
 *
 * Returns:
 *     Continuous number in [0, 100]. Higher = stronger hammer signal.
 */
export function hammer(cp: CandlestickPatternsEngine): number {
    if (!cp.enough(2, cp.shortBody, cp.longShadow, cp.veryShortShadow, cp.near)) return 0.0;

    const b1 = cp.bar(2);
    const b2 = cp.bar(1);

    const nearAvg = cp.avgCS(cp.near, 2);
    let nearWidth = cp.fuzzRatio * nearAvg;
    if (nearAvg <= 0.0) nearWidth = 0.0;

    const muShort = cp.muLessCS(realBody(b2.o, b2.c), cp.shortBody, 1);
    const muLongLS = cp.muGreaterCS(lowerShadow(b2.o, b2.l, b2.c), cp.longShadow, 1);
    const muShortUS = cp.muLessCS(upperShadow(b2.o, b2.h, b2.c), cp.veryShortShadow, 1);
    const muNearLow = cp.muLtRaw(Math.min(b2.o, b2.c), b1.l + nearAvg, nearWidth);

    return tProductAll(muShort, muLongLS, muShortUS, muNearLow) * 100.0;
}
