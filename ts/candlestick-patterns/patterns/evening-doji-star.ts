/** Evening Doji Star pattern (3-candle bearish reversal). */
import { CandlestickPatternsEngine } from '../core/engine.ts';
import { tProductAll } from '../../fuzzy/index.ts';
import { realBody, isWhite, isBlack, isRealBodyGapUp } from '../core/primitives.ts';

const EVENING_DOJI_STAR_PENETRATION = 0.3;

/**
 * Evening Doji Star: a three-candle bearish reversal pattern.
 *
 * Must have:
 * - first candle: long white real body,
 * - second candle: doji that gaps up (real body gap up from the first),
 * - third candle: black real body that moves well within the first candle's
 *   real body.
 *
 * The meaning of "long" is specified with `longBody`.
 * The meaning of "doji" is specified with `dojiBody`.
 * The meaning of "short" is specified with `shortBody`.
 *
 * Category A: always bearish (continuous).
 *
 * Returns:
 *   Continuous float in [-100, 0].  Always bearish.
 */
export function eveningDojiStar(cp: CandlestickPatternsEngine): number {
    if (!cp.enough(3, cp.longBody, cp.dojiBody, cp.shortBody)) return 0.0;

    const b1 = cp.bar(3);
    const b2 = cp.bar(2);
    const b3 = cp.bar(1);

    // Crisp gates: color checks and gap.
    if (!(isWhite(b1.o, b1.c) &&
        isRealBodyGapUp(b1.o, b1.c, b2.o, b2.c) &&
        isBlack(b3.o, b3.c))) return 0.0;

    // Fuzzy conditions.
    const muLong1 = cp.muGreaterCS(realBody(b1.o, b1.c), cp.longBody, 3);
    const muDoji2 = cp.muLessCS(realBody(b2.o, b2.c), cp.dojiBody, 2);

    // b3.c < b1.c - rb1 * penetration
    const rb1 = realBody(b1.o, b1.c);
    const threshold = b1.c - rb1 * EVENING_DOJI_STAR_PENETRATION;
    const width = cp.fuzzRatio * rb1 * EVENING_DOJI_STAR_PENETRATION;
    const muPenetration = cp.muLtRaw(b3.c, threshold, width);

    return -tProductAll(muLong1, muDoji2, muPenetration) * 100.0;
}
