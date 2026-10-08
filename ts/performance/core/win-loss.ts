import { KleinKbnSummator } from '../../streaming-kbn';

/**
 * Streaming winning/loosing return averages and counts.
 */
export class WinLoss {
    private readonly _nonZeroSum = new KleinKbnSummator();
    private readonly _winSum = new KleinKbnSummator();
    private readonly _lossSum = new KleinKbnSummator();

    reset(): void {
        this._nonZeroSum.reset();
        this._winSum.reset();
        this._lossSum.reset();
    }

    revert(ret: number): void {
        if (ret !== 0) {
            this._nonZeroSum.revert(ret);
        }
        if (ret > 0) {
            this._winSum.revert(ret);
        }
        if (ret < 0) {
            this._lossSum.revert(ret);
        }
    }

    update(ret: number): void {
        if (ret !== 0) {
            this._nonZeroSum.update(ret);
        }
        if (ret > 0) {
            this._winSum.update(ret);
        }
        if (ret < 0) {
            this._lossSum.update(ret);
        }
    }

    /** Arithmetic mean (average) of non-zero returns. */
    get nonZeroReturnsMean(): number {
        return this._nonZeroSum.mean;
    }

    /** The number of non-zero returns. */
    get nonZeroReturnsCount(): number {
        return this._nonZeroSum.n;
    }

    /** Sum of winning (positive) returns. */
    get winningReturnsSum(): number {
        return this._winSum.value;
    }

    /** Arithmetic mean (average) of winning (positive) returns. */
    get winningReturnsMean(): number {
        return this._winSum.mean;
    }

    /** The number of winning (positive) returns. */
    get winningReturnsCount(): number {
        return this._winSum.n;
    }

    /** Sum of losing (negative) returns. */
    get losingReturnsSum(): number {
        return this._lossSum.value;
    }

    /** Arithmetic mean (average) of losing (negative) returns. */
    get losingReturnsMean(): number {
        return this._lossSum.mean;
    }

    /** The number of losing (negative) returns. */
    get losingReturnsCount(): number {
        return this._lossSum.n;
    }
}
