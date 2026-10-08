import { KleinKbnAccumulator } from '../../streaming-kbn';

/**
 * Streaming upside/downside capture, number and percentage ratios of an
 * asset against a benchmark, following the PerformanceAnalytics conventions:
 * upside periods have benchmark > 0; downside capture and down-number use
 * benchmark <= 0; down-percentage uses benchmark < 0.
 *
 * `revert()` may remove any previously added pair, so the class works for
 * FIFO rolling windows.
 */
export class Capture {
    private readonly _logretASumUp = new KleinKbnAccumulator();
    private readonly _logretBSumUp = new KleinKbnAccumulator();
    private readonly _logretASumDn = new KleinKbnAccumulator();
    private readonly _logretBSumDn = new KleinKbnAccumulator();
    private readonly _aSumUp = new KleinKbnAccumulator();
    private readonly _bSumUp = new KleinKbnAccumulator();
    private readonly _aSumDn = new KleinKbnAccumulator();
    private readonly _bSumDn = new KleinKbnAccumulator();
    private _aNumUp = 0;
    private _bNumUp = 0;
    private _aNumDn = 0;
    private _bNumDn = 0;
    private _aPercUp = 0;
    private _bPercUp = 0;
    private _aPercDn = 0;
    private _bPercDn = 0;

    reset(): void {
        this._logretASumUp.reset();
        this._logretBSumUp.reset();
        this._logretASumDn.reset();
        this._logretBSumDn.reset();
        this._aSumUp.reset();
        this._bSumUp.reset();
        this._aSumDn.reset();
        this._bSumDn.reset();
        this._aNumUp = 0;
        this._bNumUp = 0;
        this._aNumDn = 0;
        this._bNumDn = 0;
        this._aPercUp = 0;
        this._bPercUp = 0;
        this._aPercDn = 0;
        this._bPercDn = 0;
    }

    private static logret(ret: number): number {
        return ret !== 0 ? Math.log1p(ret) : 0;
    }

    /** Removes a previously added (asset return, benchmark return) pair. */
    revert(retAsset: number, retBenchmark: number): void {
        if (retBenchmark > 0) { // Upside
            // Geometric
            this._logretASumUp.revert(Capture.logret(retAsset));
            this._logretBSumUp.revert(Capture.logret(retBenchmark));
            // Arithmetic
            this._aSumUp.revert(retAsset);
            this._bSumUp.revert(retBenchmark);
            // Number
            this._bNumUp -= 1;
            if (retAsset > 0) {
                this._aNumUp -= 1;
            }
            // Percentage
            this._bPercUp -= 1;
            if (retAsset > retBenchmark) {
                this._aPercUp -= 1;
            }
        } else { // Downside
            // Geometric
            this._logretASumDn.revert(Capture.logret(retAsset));
            this._logretBSumDn.revert(Capture.logret(retBenchmark));
            // Arithmetic
            this._aSumDn.revert(retAsset);
            this._bSumDn.revert(retBenchmark);
            // Number
            this._bNumDn -= 1;
            if (retAsset < 0) {
                this._aNumDn -= 1;
            }
            // Percentage
            if (retBenchmark < 0) {
                this._bPercDn -= 1;
                if (retAsset > retBenchmark) {
                    this._aPercDn -= 1;
                }
            }
        }
    }

    /** Adds an (asset return, benchmark return) pair. */
    update(retAsset: number, retBenchmark: number): void {
        if (retBenchmark > 0) { // Upside
            // Geometric
            this._logretASumUp.update(Capture.logret(retAsset));
            this._logretBSumUp.update(Capture.logret(retBenchmark));
            // Arithmetic
            this._aSumUp.update(retAsset);
            this._bSumUp.update(retBenchmark);
            // Counts
            this._bNumUp += 1;
            if (retAsset > 0) {
                this._aNumUp += 1;
            }
            // Perc
            this._bPercUp += 1;
            if (retAsset > retBenchmark) {
                this._aPercUp += 1;
            }
        } else { // Downside
            // Geometric
            this._logretASumDn.update(Capture.logret(retAsset));
            this._logretBSumDn.update(Capture.logret(retBenchmark));
            // Arithmetic
            this._aSumDn.update(retAsset);
            this._bSumDn.update(retBenchmark);
            // Counts
            this._bNumDn += 1;
            if (retAsset < 0) {
                this._aNumDn += 1;
            }
            // Perc
            if (retBenchmark < 0) {
                this._bPercDn += 1;
                if (retAsset > retBenchmark) {
                    this._aPercDn += 1;
                }
            }
        }
    }

    /** Ratio of compounded asset to compounded benchmark returns in up periods (NaN if undefined). */
    get upsideCaptureRatioGeometric(): number {
        const aCum = Math.expm1(this._logretASumUp.value);
        const bCum = Math.expm1(this._logretBSumUp.value);
        return bCum !== 0 ? aCum / bCum : NaN;
    }

    /** Ratio of summed asset to summed benchmark returns in up periods (NaN if undefined). */
    get upsideCaptureRatioArithmetic(): number {
        const bSum = this._bSumUp.value;
        return bSum !== 0 ? this._aSumUp.value / bSum : NaN;
    }

    /** Ratio of compounded asset to compounded benchmark returns in down periods (NaN if undefined). */
    get downsideCaptureRatioGeometric(): number {
        const aCum = Math.expm1(this._logretASumDn.value);
        const bCum = Math.expm1(this._logretBSumDn.value);
        return bCum !== 0 ? aCum / bCum : NaN;
    }

    /** Ratio of summed asset to summed benchmark returns in down periods (NaN if undefined). */
    get downsideCaptureRatioArithmetic(): number {
        const bSum = this._bSumDn.value;
        return bSum !== 0 ? this._aSumDn.value / bSum : NaN;
    }

    /** Fraction of up periods in which the asset return was positive (NaN if undefined). */
    get upNumberRatio(): number {
        const bNum = this._bNumUp;
        return bNum !== 0 ? this._aNumUp / bNum : NaN;
    }

    /** Fraction of down periods in which the asset return was negative (NaN if undefined). */
    get downNumberRatio(): number {
        const bNum = this._bNumDn;
        return bNum !== 0 ? this._aNumDn / bNum : NaN;
    }

    /** Fraction of up periods in which the asset outperformed the benchmark (NaN if undefined). */
    get upPercentageRatio(): number {
        const bPerc = this._bPercUp;
        return bPerc !== 0 ? this._aPercUp / bPerc : NaN;
    }

    /** Fraction of strictly negative benchmark periods in which the asset outperformed (NaN if undefined). */
    get downPercentageRatio(): number {
        const bPerc = this._bPercDn;
        return bPerc !== 0 ? this._aPercDn / bPerc : NaN;
    }
}
