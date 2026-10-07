import { KleinKbnAccumulator } from './klein-kbn-accumulator';

/**
 * Counted compensated sum: a KleinKbnAccumulator plus a sample count,
 * which also gives the arithmetic mean.
 *
 * See KleinKbnAccumulator for the summation algorithm.
 *
 * Every call to update() increments the count, including calls with
 * x == 0; every call to revert() decrements it.  Because only sums are
 * stored, revert() may remove any previously added value (not only
 * the most recent one), so the summator works for FIFO rolling windows.
 */
export class KleinKbnSummator {
    private _n = 0;
    private readonly _sum = new KleinKbnAccumulator();

    /** Clears the count and the sum. */
    reset(): void {
        this._n = 0;
        this._sum.reset();
    }

    /**
     * Removes a previously added value x.
     * Removing the final sample clears the sum and its compensation terms.
     *
     * Throws an Error if the summator is empty.
     */
    revert(x: number): void {
        if (this._n <= 0) {
            throw new Error('Cannot revert from an empty summator');
        }
        if (this._n === 1) {
            this.reset();
            return;
        }
        this._n -= 1;
        // Adding zero leaves the accumulator unchanged, so skip it.
        if (x !== 0) {
            this._sum.revert(x);
        }
    }

    /** Adds a value x. */
    update(x: number): void {
        this._n += 1;
        // Adding zero leaves the accumulator unchanged, so skip it.
        if (x !== 0) {
            this._sum.update(x);
        }
    }

    /** The compensated sum of all added values (0.0 when empty). */
    get value(): number {
        return this._sum.value;
    }

    /** The arithmetic mean, sum / n (NaN when empty). */
    get mean(): number {
        const n = this._n;
        if (n <= 0) {
            return NaN;
        }
        return this._sum.value / n;
    }

    /** The number of added values. */
    get n(): number {
        return this._n;
    }
}
