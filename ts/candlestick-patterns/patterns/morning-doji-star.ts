/** Morning Doji Star pattern (3-candle bullish reversal). */
import { CandlestickPatternsEngine } from '../core/engine.ts';
import { tProductAll } from '../../fuzzy/index.ts';
import { realBody, isBlack, isWhite, isRealBodyGapDown } from '../core/primitives.ts';

const MORNING_DOJI_STAR_PENETRATION = 0.3;

/**
 * Morning Doji Star: a three-candle bullish reversal pattern.
 *
 * Must have:
 * - first candle: long black real body,
 * - second candle: doji that gaps down (real body gap down from the first),
 * - third candle: white real body that closes well within the first candle's
 *   real body.
 *
 * The meaning of "long" is specified with `longBody`.
 * The meaning of "doji" is specified with `dojiBody`.
 * The meaning of "short" is specified with `shortBody`.
 *
 * Category A: always bullish (continuous).
 *
 * Returns:
 *     Continuous float in [0, +100].  Always bullish.
 */
export function morningDojiStar(cp: CandlestickPatternsEngine): number {
    if (!cp.enough(3, cp.longBody, cp.dojiBody, cp.shortBody)) return 0.0;

    const b1 = cp.bar(3);
    const b2 = cp.bar(2);
    const b3 = cp.bar(1);

    // Crisp gates: color checks and gap.
    if (!(isBlack(b1.o, b1.c) &&
        isRealBodyGapDown(b1.o, b1.c, b2.o, b2.c) &&
        isWhite(b3.o, b3.c))) return 0.0;

    // Fuzzy conditions.
    const muLong1 = cp.muGreaterCS(realBody(b1.o, b1.c), cp.longBody, 3);
    const muDoji2 = cp.muLessCS(realBody(b2.o, b2.c), cp.dojiBody, 2);

    // b3.c > b1.c + rb1 * penetration
    const rb1 = realBody(b1.o, b1.c);
    const threshold = b1.c + rb1 * MORNING_DOJI_STAR_PENETRATION;
    const width = cp.fuzzRatio * rb1 * MORNING_DOJI_STAR_PENETRATION;
    const muPenetration = cp.muGtRaw(b3.c, threshold, width);

    return tProductAll(muLong1, muDoji2, muPenetration) * 100.0;
}
