import { LinearRegressionKleinKbn } from '../../streaming-kbn';

/**
 * Streaming single-factor model (SFM) regression of the asset excess returns
 * on the benchmark excess returns (both less the risk-free rate).
 *
 * Maintains the full regression plus separate bull (benchmark excess > 0)
 * and bear (benchmark excess < 0) regressions. `revert()` may remove any
 * previously added pair, so the class works for FIFO rolling windows.
 */
export class SFMRegression {
    private readonly _riskFreeRate: number;
    private readonly _full = new LinearRegressionKleinKbn();
    private readonly _bull = new LinearRegressionKleinKbn();
    private readonly _bear = new LinearRegressionKleinKbn();

    constructor(riskFreeRate: number) {
        this._riskFreeRate = riskFreeRate;
    }

    reset(): void {
        this._full.reset();
        this._bull.reset();
        this._bear.reset();
    }

    /** Removes a previously added (asset return, benchmark return) pair. */
    revert(ret: number, benchmark: number): void {
        const x = benchmark - this._riskFreeRate;
        const y = ret - this._riskFreeRate;

        this._full.revert(x, y);

        if (x > 0) {
            this._bull.revert(x, y);
        } else if (x < 0) {
            this._bear.revert(x, y);
        }
    }

    /** Adds an (asset return, benchmark return) pair. */
    update(ret: number, benchmark: number): void {
        const x = benchmark - this._riskFreeRate;
        const y = ret - this._riskFreeRate;

        this._full.update(x, y);

        if (x > 0) {
            this._bull.update(x, y);
        } else if (x < 0) {
            this._bear.update(x, y);
        }
    }

    /** Intercept of the full regression (Jensen's alpha, per period). */
    get alpha(): number {
        return this._full.intercept;
    }

    /** Slope of the full regression. */
    get beta(): number {
        return this._full.slope;
    }

    /** Slope of the bull-market (benchmark excess > 0) regression. */
    get betaBull(): number {
        return this._bull.slope;
    }

    /** Slope of the bear-market (benchmark excess < 0) regression. */
    get betaBear(): number {
        return this._bear.slope;
    }

    /** Coefficient of determination of the full regression (NaN if undefined). */
    get r2(): number {
        const corr = this._full.correlation;
        return !Number.isNaN(corr) ? corr * corr : NaN;
    }
}
