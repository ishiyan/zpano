import { RangeEntity } from './range-entity.ts';
import { candleRangeValue } from './primitives.ts';

/**
 * Criterion specifies a threshold based on the average value of a candlestick range entity.
 *
 * The criteria are based on parts of the candlestick and common words indicating length
 * (short, long, very long), displacement (near, far), or equality (equal).
 *
 * For streaming efficiency, the criterion maintains a running total that is updated
 * incrementally via add() and remove() rather than rescanning the entire history.
 */
export class Criterion {
    /** The type of range entity to consider (RealBody, HighLow, or Shadows). */
    readonly entity: RangeEntity;
    /** The number of previous candlesticks to calculate an average value. */
    readonly averagePeriod: number;
    /** The coefficient to multiply the average value. */
    readonly factor: number;

    constructor(entity: RangeEntity, averagePeriod: number, factor: number) {
        this.entity = entity;
        this.averagePeriod = averagePeriod;
        this.factor = factor;
    }

    /** Creates an independent copy. */
    copy(): Criterion {
        return new Criterion(this.entity, this.averagePeriod, this.factor);
    }

    /**
     * Computes the criterion threshold from a precomputed running total.
     *
     * When averagePeriod > 0, uses the running total.
     * When averagePeriod == 0, uses the current candle's own range value.
     */
    averageValueFromTotal(total: number, o: number, h: number, l: number, c: number): number {
        if (this.averagePeriod > 0) {
            if (this.entity === RangeEntity.SHADOWS) {
                return this.factor * total / (this.averagePeriod * 2.0);
            }
            return this.factor * total / this.averagePeriod;
        }
        // Period == 0: use the candle's own range value directly.
        return this.factor * candleRangeValue(this.entity, o, h, l, c);
    }

    /**
     * Computes the contribution of a single candle to the running total.
     *
     * For Shadows entity, this returns the full (upper + lower) shadow sum
     * (not yet divided by 2 -- the division happens in averageValueFromTotal).
     */
    candleContribution(o: number, h: number, l: number, c: number): number {
        switch (this.entity) {
            case RangeEntity.REAL_BODY:
                if (c >= o) {
                    return c - o;
                }
                return o - c;
            case RangeEntity.HIGH_LOW:
                return h - l;
            default:
                // SHADOWS: upper + lower shadow sum (division by 2 deferred to averageValueFromTotal)
                if (c >= o) {
                    return h - c + o - l;
                }
                return h - o + c - l;
        }
    }
}
