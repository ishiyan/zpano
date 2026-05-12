/** Tristar pattern (3-candle reversal). */
import { CandlestickPatternsEngine } from '../core/engine.ts';
import { tProductAll } from '../../fuzzy/index.ts';
import { realBody, isRealBodyGapUp, isRealBodyGapDown } from '../core/primitives.ts';

/**
 * Tristar: a three-candle reversal pattern with three dojis.
 *
 * Must have:
 * - three consecutive doji candles,
 * - if the second doji gaps up from the first and the third does not
 *   close higher than the second: bearish,
 * - if the second doji gaps down from the first and the third does not
 *   close lower than the second: bullish.
 *
 * Category A: fixed direction per branch (bullish or bearish).
 *
 * Returns:
 *     Continuous float in [-100, +100].
 */
export function tristar(cp: CandlestickPatternsEngine): number {
    if (!cp.enough(3, cp.dojiBody)) return 0.0;

    const b1 = cp.bar(3);
    const b2 = cp.bar(2);
    const b3 = cp.bar(1);

    // Fuzzy: all three must be dojis.
    const muDoji1 = cp.muLessCS(realBody(b1.o, b1.c), cp.dojiBody, 3);
    const muDoji2 = cp.muLessCS(realBody(b2.o, b2.c), cp.dojiBody, 2);
    const muDoji3 = cp.muLessCS(realBody(b3.o, b3.c), cp.dojiBody, 1);

    // Bearish: second gaps up, third is not higher than second — crisp direction checks.
    if (isRealBodyGapUp(b1.o, b1.c, b2.o, b2.c) &&
        Math.max(b3.o, b3.c) < Math.max(b2.o, b2.c)) {
        return -tProductAll(muDoji1, muDoji2, muDoji3) * 100.0;
    }

    // Bullish: second gaps down, third is not lower than second.
    if (isRealBodyGapDown(b1.o, b1.c, b2.o, b2.c) &&
        Math.min(b3.o, b3.c) > Math.min(b2.o, b2.c)) {
        return tProductAll(muDoji1, muDoji2, muDoji3) * 100.0;
    }

    return 0.0;
}
