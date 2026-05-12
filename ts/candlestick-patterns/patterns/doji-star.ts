/** Doji Star pattern (2-candle reversal). */
import { CandlestickPatternsEngine } from '../core/engine.ts';
import { tProductAll } from '../../fuzzy/index.ts';
import { realBody, isRealBodyGapUp, isRealBodyGapDown } from '../core/primitives.ts';

/**
 * Doji Star: a two-candle reversal pattern.
 *
 * Must have:
 * - first candle: long real body,
 * - second candle: doji that gaps away from the first candle.
 *
 * - bearish: first candle is long white, doji gaps up,
 * - bullish: first candle is long black, doji gaps down.
 *
 * The meaning of "long" is specified with `longBody`.
 * The meaning of "doji" is specified with `dojiBody`.
 *
 * Category B: direction from 1st candle color (continuous).
 *
 * Returns:
 *   Continuous float in [-100, +100].
 */
export function dojiStar(cp: CandlestickPatternsEngine): number {
    if (!cp.enough(2, cp.longBody, cp.dojiBody)) return 0.0;

    const b1 = cp.bar(2);
    const b2 = cp.bar(1);

    const color1 = b1.c < b1.o ? -1 : 1;

    // Crisp gates: gap direction must match color.
    if (color1 === 1 && !isRealBodyGapUp(b1.o, b1.c, b2.o, b2.c)) return 0.0;
    if (color1 === -1 && !isRealBodyGapDown(b1.o, b1.c, b2.o, b2.c)) return 0.0;

    // Fuzzy conditions.
    const muLong1 = cp.muGreaterCS(realBody(b1.o, b1.c), cp.longBody, 2);
    const muDoji2 = cp.muLessCS(realBody(b2.o, b2.c), cp.dojiBody, 1);

    const confidence = tProductAll(muLong1, muDoji2);
    // Direction: opposite of 1st candle color.
    const direction = color1 === -1 ? 1.0 : -1.0;
    return direction * confidence * 100.0;
}
