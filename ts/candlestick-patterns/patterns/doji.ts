/** Doji pattern. */
import { CandlestickPatternsEngine } from '../core/engine.ts';
import { realBody } from '../core/primitives.ts';

/**
 * Doji: open quite equal to close.
 *
 * Output is positive but this does not mean it is bullish:
 * doji shows uncertainty and is neither bullish nor bearish when
 * considered alone.
 *
 * The meaning of "doji" is specified with `dojiBody`.
 *
 * Returns:
 *   Continuous float in [0, 100].  Higher = stronger doji signal.
 */
export function doji(cp: CandlestickPatternsEngine): number {
    if (!cp.enough(1, cp.dojiBody)) return 0.0;
    const b = cp.bar(1);
    // Fuzzy: degree to which realBody <= dojiAvg.
    return cp.muLessCS(realBody(b.o, b.c), cp.dojiBody, 1) * 100.0;
}
