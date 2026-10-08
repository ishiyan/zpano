import { KleinKbnAccumulator } from '../../streaming-kbn';

/**
 * Streaming raw partial moments.
 *
 * Sums (not means) of the lower/higher partial moments about zero, and
 * counts/sums of the strictly negative and strictly positive returns.
 * `revert()` may remove any previously added return.
 */
export class RawPartialMoments {
    private _count = 0;
    private _countPos = 0;
    private _countNeg = 0;
    private readonly _lpmKbn = new KleinKbnAccumulator();
    private readonly _hpmKbn = new KleinKbnAccumulator();
    private readonly _posKbn = new KleinKbnAccumulator();
    private readonly _negKbn = new KleinKbnAccumulator();

    reset(): void {
        this._count = 0;
        this._countPos = 0;
        this._countNeg = 0;
        this._lpmKbn.reset();
        this._hpmKbn.reset();
        this._posKbn.reset();
        this._negKbn.reset();
    }

    /** Removes a previously added return. */
    revert(ret: number): void {
        this._count -= 1;
        // Lower partial moment
        let pm = -ret;
        if (pm < 0) {
            pm = 0;
        }
        this._lpmKbn.revert(pm);

        // Higher partial moment
        pm = ret;
        if (pm < 0) {
            pm = 0;
        }
        this._hpmKbn.revert(pm);

        if (ret > 0) {
            this._countPos -= 1;
            this._posKbn.revert(ret);
        } else if (ret < 0) {
            this._countNeg -= 1;
            this._negKbn.revert(ret);
        }
    }

    /** Adds a return. */
    update(ret: number): void {
        this._count += 1;
        // Lower partial moment
        let pm = -ret;
        if (pm < 0) {
            pm = 0;
        }
        this._lpmKbn.update(pm);

        // Higher partial moment
        pm = ret;
        if (pm < 0) {
            pm = 0;
        }
        this._hpmKbn.update(pm);

        if (ret > 0) {
            this._countPos += 1;
            this._posKbn.update(ret);
        } else if (ret < 0) {
            this._countNeg += 1;
            this._negKbn.update(ret);
        }
    }

    /** Number of returns. */
    get count(): number {
        return this._count;
    }

    /** Sum of max(-r, 0) over all returns. */
    get lowerPartialMoment1(): number {
        return this._lpmKbn.value;
    }

    /** Sum of max(r, 0) over all returns. */
    get higherPartialMoment1(): number {
        return this._hpmKbn.value;
    }

    /** Number of negative returns. */
    get countNegative(): number {
        return this._countNeg;
    }

    /** Sum of negative returns. */
    get sumNegative(): number {
        return this._negKbn.value;
    }

    /** Number of positive returns. */
    get countPositive(): number {
        return this._countPos;
    }

    /** Sum of positive returns. */
    get sumPositive(): number {
        return this._posKbn.value;
    }
}
