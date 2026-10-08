import { KleinKbnAccumulator } from '../../streaming-kbn';

/**
 * Streaming cumulative (geometric) returns.
 *
 * Accumulates the sum of log returns, log(1 + r), with compensated
 * summation.  Because only a sum is stored, revert() may remove any
 * previously added return, so the class works for FIFO rolling windows:
 * the caller owns the window and feeds evicted returns to revert().
 */
export class CumulativeReturn {
    private readonly _cumlogretSum = new KleinKbnAccumulator();
    private _count = 0;

    reset(): void {
        this._cumlogretSum.reset();
        this._count = 0;
    }

    /**
     * Removes a previously added return.
     *
     * Throws an Error if there are no returns.
     */
    revert(ret: number): void {
        if (this._count <= 0) {
            throw new Error('Cannot revert from an empty accumulator');
        }
        this._count -= 1;
        this._cumlogretSum.revert(ret !== 0 ? Math.log1p(ret) : 0);
    }

    /** Adds a return, expressed as a decimal (must be > -1). */
    update(ret: number): void {
        this._count += 1;
        this._cumlogretSum.update(ret !== 0 ? Math.log1p(ret) : 0);
    }

    /** The number of returns. */
    get count(): number {
        return this._count;
    }

    /** Cumulative geometric return, prod(1 + r) - 1 (0.0 when empty). */
    get cumulativeGeometricReturn(): number {
        return Math.expm1(this._cumlogretSum.value);
    }

    /**
     * The geometric mean of the returns, prod(1 + r)^(1/n) - 1
     * (NaN when empty).
     */
    get geometricMeanReturn(): number {
        return this._count > 0 ? Math.expm1(this._cumlogretSum.value / this._count) : NaN;
    }

    /**
     * The annualized geometric mean, prod(1 + r)^(periodsPerYear/n) - 1
     * (NaN when empty).
     */
    annualizedGeometricMeanReturn(periodsPerYear: number): number {
        if (this._count === 0) {
            return NaN;
        }
        return Math.expm1(this._cumlogretSum.value * periodsPerYear / this._count);
    }
}
