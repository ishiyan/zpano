import { KleinKbnAccumulator } from '../../streaming-kbn';

/**
 * A single high-water-mark drawdown episode (immutable).
 */
export class DrawdownEpisode {
    /**
     * @param depth Maximum drawdown depth, expressed as a decimal return.
     * @param fromIdx Index of the first underwater observation.
     * @param troughIdx Index of the deepest observation.
     * @param toIdx Recovery observation index if the episode recovered;
     *     otherwise the last observation index currently available.
     * @param recovered True if the high-water mark was recovered at `toIdx`.
     *     False if the series ended while the drawdown was still open.
     */
    constructor(
        readonly depth: number,
        readonly fromIdx: number,
        readonly troughIdx: number,
        readonly toIdx: number,
        readonly recovered: boolean,
    ) {
        Object.freeze(this);
    }
}

/**
 * Streaming high-water-mark drawdown episode tracker.
 *
 * The class consumes drawdown observations produced by
 * `HighWaterMarkDrawdown` and maintains the corresponding
 * drawdown episodes incrementally.
 *
 * Drawdowns are expected as decimals, for example `-0.025`
 * for a 2.5% drawdown and `0.0` for an observation at a
 * high-water mark.
 *
 * Indices are positions in the sequence of observations passed to
 * `update()` since the last `reset()`/`recalculate()`.  For a
 * rolling window, call `recalculate()` with the window's drawdowns
 * whenever they are recomputed, so indices refer to the window.
 *
 * A drawdown episode begins with the first negative drawdown and
 * remains open until a non-negative drawdown is observed.
 *
 * The episode is considered recovered when the drawdown reaches
 * zero or becomes positive.
 */
export class DrawdownEpisodes {
    private readonly _episodes: DrawdownEpisode[] = [];

    // Current open drawdown episode.
    private _currentFrom: number | null = null;
    private _currentTrough: number | null = null;
    private _currentDepth = 0.0;

    // Number of observations processed.
    private _count = 0;

    // Running depth aggregates.
    private readonly _sumDepth = new KleinKbnAccumulator();
    private readonly _sumDepthSquared = new KleinKbnAccumulator();
    // Running episode length/peak-to-trough/recovery aggregates.
    private _sumLength = 0;
    private _sumPeakToTrough = 0;
    private _sumRecovery = 0;

    /** Reset the episode tracker to its initial empty state. */
    reset(): void {
        this._episodes.length = 0;
        this._currentFrom = null;
        this._currentTrough = null;
        this._currentDepth = 0.0;
        this._count = 0;
        this._sumDepth.reset();
        this._sumDepthSquared.reset();
        this._sumLength = 0;
        this._sumPeakToTrough = 0;
        this._sumRecovery = 0;
    }

    /**
     * Add one drawdown observation.
     *
     * @param drawdown Drawdown at the current observation, expressed as a
     *     decimal. Drawdowns must be non-positive, although non-negative
     *     values are accepted and treated as recovery or high-water-mark
     *     observations.
     */
    update(drawdown: number): void {
        const idx = this._count;
        this._count += 1;

        if (drawdown < 0.0) {
            // We are underwater.
            if (this._currentFrom === null) {
                // Start a new drawdown episode.
                this._currentFrom = idx;
                this._currentTrough = idx;
                this._currentDepth = drawdown;
            } else if (drawdown < this._currentDepth) {
                // New trough within the current episode.
                this._currentTrough = idx;
                this._currentDepth = drawdown;
            }
            return;
        }

        // We are at or above the high-water mark.
        if (this._currentFrom !== null) {
            const trough = this._currentTrough as number;
            // Close the current drawdown episode.
            this._episodes.push(new DrawdownEpisode(this._currentDepth,
                this._currentFrom, trough, idx, true));
            this._sumDepth.update(this._currentDepth);
            this._sumDepthSquared.update(this._currentDepth * this._currentDepth);
            this._sumLength += idx - this._currentFrom + 1;
            this._sumPeakToTrough += trough - this._currentFrom + 1;
            this._sumRecovery += idx - trough + 1;
            this._currentFrom = null;
            this._currentTrough = null;
            this._currentDepth = 0.0;
        }
    }

    /**
     * Rebuild all episodes from drawdown observations history.
     *
     * @param drawdowns Drawdown observations.
     */
    recalculate(drawdowns: Iterable<number>): void {
        this.reset();
        for (const dd of drawdowns) {
            this.update(dd);
        }
    }

    /**
     * Drawdown episodes currently known to the tracker (a new array).
     *
     * If the latest drawdown episode is still open, it is included
     * using the last processed observation as `toIdx` and with
     * `recovered=false`.
     */
    get episodes(): DrawdownEpisode[] {
        const episodes = this._episodes.slice();
        if (this._currentFrom !== null) {
            episodes.push(new DrawdownEpisode(this._currentDepth,
                this._currentFrom, this._currentTrough as number, this._count - 1, false));
        }
        return episodes;
    }

    /**
     * Drawdown episode depths currently known to the tracker (a new array),
     * including the depth so far of an open episode.
     */
    get depths(): number[] {
        const depths = this._episodes.map(episode => episode.depth);
        if (this._currentFrom !== null) {
            depths.push(this._currentDepth);
        }
        return depths;
    }

    /** The mean magnitude of the observed discrete episode drawdowns. */
    get averageEpisodeDrawdown(): number {
        let sumDepth = this._sumDepth.value;
        let count = this._episodes.length;
        if (this._currentFrom !== null) {
            sumDepth += this._currentDepth;
            count += 1;
        }
        return count > 0 ? -sumDepth / count : 0.0;
    }

    /**
     * The sum of squared episode depths divided by the number of
     * observations (not the number of episodes), as in
     * PerformanceAnalytics `DrawdownDeviation`.
     */
    get averageEpisodeDrawdownSquared(): number {
        let sumDepthSquared = this._sumDepthSquared.value;
        const count = this._count;
        if (count === 0) {
            return 0.0;
        }
        if (this._currentFrom !== null) {
            sumDepthSquared += this._currentDepth * this._currentDepth;
        }
        return sumDepthSquared / count;
    }

    /** The mean length of the observed discrete drawdown episodes. */
    get averageEpisodeLength(): number {
        let sumLength = this._sumLength;
        let count = this._episodes.length;
        if (this._currentFrom !== null) {
            sumLength += this._count - this._currentFrom;
            count += 1;
        }
        return count > 0 ? sumLength / count : 0.0;
    }

    /** The mean peak-to-trough length of the observed discrete drawdown episodes. */
    get averageEpisodePeakToTrough(): number {
        let sumPeakToTrough = this._sumPeakToTrough;
        let count = this._episodes.length;
        if (this._currentFrom !== null) {
            sumPeakToTrough += (this._currentTrough as number) - this._currentFrom + 1;
            count += 1;
        }
        return count > 0 ? sumPeakToTrough / count : 0.0;
    }

    /** The mean recovery length of the observed discrete drawdown episodes. */
    get averageEpisodeRecovery(): number {
        let sumRecovery = this._sumRecovery;
        let count = this._episodes.length;
        if (this._currentFrom !== null) {
            sumRecovery += this._count - (this._currentTrough as number);
            count += 1;
        }
        return count > 0 ? sumRecovery / count : 0.0;
    }
}
