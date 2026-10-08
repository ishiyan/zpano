import { RawMomentsKleinKbn } from '../../streaming-kbn';

/**
 * Streaming partial moments.
 *
 * Lower/higher partial moments about a threshold over all returns, plus
 * moments of the excesses strictly below/above the threshold.
 * `revert()` may remove any previously added return, so the class works
 * for FIFO rolling windows.
 */
export class PartialMoments {
    /** Target return or minimum acceptable return (MAR) in same periodicity as returns. */
    threshold: number;
    private _countTotal = 0;

    // This matches scipy's default behavior for kurtosis.
    private readonly _upperExcessKbn = new RawMomentsKleinKbn(1, true, true);
    private readonly _lowerExcessKbn = new RawMomentsKleinKbn(1, true, true);
    private readonly _lpmKbn = new RawMomentsKleinKbn(1, true, true);
    private readonly _hpmKbn = new RawMomentsKleinKbn(1, true, true);

    /**
     * Streaming low/high partial moments.
     *
     * @param threshold Target return or minimum acceptable return (MAR) in same periodicity as returns.
     */
    constructor(threshold: number) {
        this.threshold = threshold;
    }

    reset(): void {
        this._countTotal = 0;
        this._upperExcessKbn.reset();
        this._lowerExcessKbn.reset();
        this._lpmKbn.reset();
        this._hpmKbn.reset();
    }

    /** Removes a previously added return. */
    revert(ret: number): void {
        this._countTotal -= 1;
        // Lower partial moments for the raw returns less target return
        let pm = this.threshold - ret;
        if (pm < 0) {
            this._upperExcessKbn.revert(-pm);
            pm = 0;
        }
        this._lpmKbn.revert(pm);

        // Higher partial moments for the raw returns less required return
        pm = ret - this.threshold;
        if (pm < 0) {
            this._lowerExcessKbn.revert(-pm);
            pm = 0;
        }
        this._hpmKbn.revert(pm);
    }

    /** Adds a return. */
    update(ret: number): void {
        this._countTotal += 1;
        // Lower partial moments for the raw returns less target return
        let pm = this.threshold - ret;
        if (pm < 0) {
            this._upperExcessKbn.update(-pm);
            pm = 0;
        }
        this._lpmKbn.update(pm);

        // Higher partial moments for the raw returns less required return
        pm = ret - this.threshold;
        if (pm < 0) {
            this._lowerExcessKbn.update(-pm);
            pm = 0;
        }
        this._hpmKbn.update(pm);
    }

    /** Mean of max(threshold - r, 0)^1 over all returns. */
    get lowerPartialMoment1(): number {
        return this._lpmKbn.x1;
    }

    /** Mean of max(threshold - r, 0)^2 over all returns. */
    get lowerPartialMoment2(): number {
        return this._lpmKbn.x2;
    }

    /** Mean of max(threshold - r, 0)^3 over all returns. */
    get lowerPartialMoment3(): number {
        return this._lpmKbn.x3;
    }

    /** Mean of max(threshold - r, 0)^4 over all returns. */
    get lowerPartialMoment4(): number {
        return this._lpmKbn.x4;
    }

    /** Mean of max(r - threshold, 0)^1 over all returns. */
    get higherPartialMoment1(): number {
        return this._hpmKbn.x1;
    }

    /** Mean of max(r - threshold, 0)^2 over all returns. */
    get higherPartialMoment2(): number {
        return this._hpmKbn.x2;
    }

    /** Mean of max(r - threshold, 0)^3 over all returns. */
    get higherPartialMoment3(): number {
        return this._hpmKbn.x3;
    }

    /** Mean of max(r - threshold, 0)^4 over all returns. */
    get higherPartialMoment4(): number {
        return this._hpmKbn.x4;
    }

    /** Proportion of returns below threshold (NaN when empty). */
    get downsideFrequency(): number {
        const total = this._countTotal;
        if (total === 0) {
            return NaN;
        }
        return this._lowerExcessKbn.n / total;
    }

    /** Proportion of returns above threshold (NaN when empty). */
    get upsideFrequency(): number {
        const total = this._countTotal;
        if (total === 0) {
            return NaN;
        }
        return this._upperExcessKbn.n / total;
    }

    /** Mean of lower partial moments (also called shortfall). */
    get downsidePotential(): number {
        return this._lpmKbn.mean;
    }

    /** Total number of returns. */
    get totalCount(): number {
        return this._countTotal;
    }

    /** Number of returns above threshold. */
    get upperExcessCount(): number {
        return this._upperExcessKbn.n;
    }

    /** Number of returns below threshold. */
    get lowerExcessCount(): number {
        return this._lowerExcessKbn.n;
    }

    /** Mean of (r - threshold) over returns above threshold. */
    get upperExcessMoment1(): number {
        return this._upperExcessKbn.x1;
    }

    /** Sum of (r - threshold) over returns above threshold. */
    get upperExcessMoment1Sum(): number {
        return this._upperExcessKbn.x1Sum;
    }

    /** Mean of (r - threshold)^2 over returns above threshold. */
    get upperExcessMoment2(): number {
        return this._upperExcessKbn.x2;
    }

    /** Sum of (r - threshold)^2 over returns above threshold. */
    get upperExcessMoment2Sum(): number {
        return this._upperExcessKbn.x2Sum;
    }

    /** Mean of (r - threshold)^3 over returns above threshold. */
    get upperExcessMoment3(): number {
        return this._upperExcessKbn.x3;
    }

    /** Mean of (r - threshold)^4 over returns above threshold. */
    get upperExcessMoment4(): number {
        return this._upperExcessKbn.x4;
    }

    /** Mean of (threshold - r) over returns below threshold. */
    get lowerExcessMoment1(): number {
        return this._lowerExcessKbn.x1;
    }

    /** Mean of (threshold - r)^2 over returns below threshold. */
    get lowerExcessMoment2(): number {
        return this._lowerExcessKbn.x2;
    }

    /** Sum of (threshold - r)^2 over returns below threshold. */
    get lowerExcessMoment2Sum(): number {
        return this._lowerExcessKbn.x2Sum;
    }

    /** Mean of (threshold - r)^3 over returns below threshold. */
    get lowerExcessMoment3(): number {
        return this._lowerExcessKbn.x3;
    }

    /** Mean of (threshold - r)^4 over returns below threshold. */
    get lowerExcessMoment4(): number {
        return this._lowerExcessKbn.x4;
    }
}
