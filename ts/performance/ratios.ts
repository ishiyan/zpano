import { DayCountConvention, yearFrac, dayFrac } from '../daycounting';
import { Periodicity } from './periodicity';

const SQRT2 = 1.4142135623730950488016887242097;

/**
 * Statistical helper functions
 */

function mean(arr: number[]): number | null {
    if (arr.length === 0) return null;
    return arr.reduce((sum, val) => sum + val, 0) / arr.length;
}

function std(arr: number[], ddof: number = 0): number | null {
    if (arr.length <= ddof) return null;
    const m = mean(arr);
    if (m === null) return null;
    const sumSquares = arr.reduce((sum, val) => sum + Math.pow(val - m, 2), 0);
    return Math.sqrt(sumSquares / (arr.length - ddof));
}

function skewness(arr: number[]): number | null {
    if (arr.length <= 1) return null;
    const m = mean(arr);
    if (m === null) return null;
    const s = std(arr, 0);
    if (s === null || s === 0) return null;
    
    const n = arr.length;
    const sumCubes = arr.reduce((sum, val) => sum + Math.pow((val - m) / s, 3), 0);
    
    // Sample skewness (Fisher-Pearson coefficient)
    return (n / ((n - 1) * (n - 2))) * sumCubes;
}

function kurtosisExcess(arr: number[]): number | null {
    if (arr.length <= 1) return null;
    const m = mean(arr);
    if (m === null) return null;
    const n = arr.length;
    
    // Population excess kurtosis (matches scipy.stats.kurtosis with bias=True, fisher=True)
    // Formula: m4 / m2^2 - 3, where m4 and m2 are population central moments
    const m2 = arr.reduce((sum, val) => sum + Math.pow(val - m, 2), 0) / n;
    if (m2 === 0) return null;
    const m4 = arr.reduce((sum, val) => sum + Math.pow(val - m, 4), 0) / n;
    
    return m4 / (m2 * m2) - 3;
}

/**
 * Various financial ratios to evaluate the performance of a strategy.
 */
export class Ratios {
    public periodicity: Periodicity;
    public periodsPerAnnum: number;
    public daysPerPeriod: number;
    public riskFreeRate: number;
    public requiredReturn: number;
    public dayCountConvention: DayCountConvention;
    public rollingWindow: number | null;
    public minPeriods: number | null;

    public fractionalPeriods: number[] = [];
    public returns: number[] = [];
    private _sampleCount: number = 0;

    private _logretSum: number = 0;
    private _drawdownsCumulative: number[] = [];
    private _drawdownsCumulativeMin: number = Infinity;
    private _drawdownsPeaks: number[] = [];
    private _drawdownsPeaksPeak: number = 0;
    private _drawdownContinuous: number[] = [];
    private _drawdownContinuousFinal: number[] = [];
    private _drawdownContinuousFinalized: boolean = false;
    private _drawdownContinuousPeak: number = 1;
    private _drawdownContinuousInside: boolean = false;
    private _cumulativeReturnPlus1: number = 1;
    private _cumulativeReturnPlus1Max: number = -Infinity;
    private _cumulativeReturnGeometricMean: number | null = null;
    private _returnsMean: number | null = null;
    private _returnsStd: number | null = null;
    private _returnsAutocorrPenalty: number = 1;
    private _excessMean: number | null = null;
    private _excessStd: number | null = null;
    private _excessAutocorrPenalty: number = 1;
    private _requiredMean: number | null = null;
    private _requiredLpm1: number | null = null;
    private _requiredLpm2: number | null = null;
    private _requiredLpm3: number | null = null;
    private _requiredHpm1: number | null = null;
    private _requiredHpm2: number | null = null;
    private _requiredHpm3: number | null = null;
    private _requiredAutocorrPenalty: number = 1;
    private _avgReturn: number | null = null;
    private _avgWin: number | null = null;
    private _avgLoss: number | null = null;
    private _winRate: number | null = null;
    private _totalDuration: number = 0;

    constructor(
        periodicity: Periodicity = Periodicity.DAILY,
        annualRiskFreeRate: number = 0,
        annualTargetReturn: number = 0,
        dayCountConvention: DayCountConvention = DayCountConvention.RAW,
        rollingWindow: number | null = null,
        minPeriods: number | null = null
    ) {
        this.periodicity = periodicity;
        
        const periodsPerAnnum = periodicity === Periodicity.DAILY ? 252
            : periodicity === Periodicity.WEEKLY ? 52
            : periodicity === Periodicity.MONTHLY ? 12
            : periodicity === Periodicity.QUARTERLY ? 4
            : 1;
        this.periodsPerAnnum = periodsPerAnnum;
        
        this.daysPerPeriod = periodicity === Periodicity.DAILY ? 1
            : periodicity === Periodicity.WEEKLY ? 252 / 52
            : periodicity === Periodicity.MONTHLY ? 252 / 12
            : periodicity === Periodicity.QUARTERLY ? 252 / 4
            : 252;

        this.riskFreeRate = annualRiskFreeRate === 0 || periodsPerAnnum === 1
            ? annualRiskFreeRate
            : Math.pow(1 + annualRiskFreeRate, 1 / periodsPerAnnum) - 1;

        this.requiredReturn = annualTargetReturn === 0 || periodsPerAnnum === 1
            ? annualTargetReturn
            : Math.pow(1 + annualTargetReturn, 1 / periodsPerAnnum) - 1;
            
        this.dayCountConvention = dayCountConvention;
        this.rollingWindow = rollingWindow;
        this.minPeriods = (minPeriods !== null && minPeriods > 0) ? minPeriods : null;
    }

    reset(): void {
        this.fractionalPeriods = [];
        this.returns = [];
        this._sampleCount = 0;
        this._logretSum = 0;
        this._drawdownsCumulative = [];
        this._drawdownsCumulativeMin = Infinity;
        this._drawdownsPeaks = [];
        this._drawdownsPeaksPeak = 0;
        this._drawdownContinuous = [];
        this._drawdownContinuousFinal = [];
        this._drawdownContinuousFinalized = false;
        this._drawdownContinuousPeak = 1;
        this._drawdownContinuousInside = false;
        this._cumulativeReturnPlus1 = 1;
        this._cumulativeReturnPlus1Max = -Infinity;
        this._totalDuration = 0;
        this._cumulativeReturnGeometricMean = null;
        this._returnsMean = null;
        this._returnsStd = null;
        this._returnsAutocorrPenalty = 1;
        this._excessMean = null;
        this._excessStd = null;
        this._excessAutocorrPenalty = 1;
        this._requiredMean = null;
        this._requiredLpm1 = null;
        this._requiredLpm2 = null;
        this._requiredLpm3 = null;
        this._requiredHpm1 = null;
        this._requiredHpm2 = null;
        this._requiredHpm3 = null;
        this._requiredAutocorrPenalty = 1;
        this._avgReturn = null;
        this._avgWin = null;
        this._avgLoss = null;
        this._winRate = null;
    }

    addReturn(
        return_: number,
        _returnBenchmark: number,
        _value: number,
        timeStart: Date,
        timeEnd: Date
    ): void {
        let fractionalPeriod: number;
        
        if (this.periodicity === Periodicity.ANNUAL) {
            fractionalPeriod = yearFrac(timeStart, timeEnd, this.dayCountConvention);
        } else {
            fractionalPeriod = dayFrac(timeStart, timeEnd, this.dayCountConvention) / this.daysPerPeriod;
        }

        this.fractionalPeriods.push(fractionalPeriod);
        
        if (fractionalPeriod === 0) {
            console.warn('Zero fractional time period, performance not updated');
            return;
        }
        
        this._totalDuration += fractionalPeriod;
        this._sampleCount += 1;

        // Normalized returns
        const ret = return_ / fractionalPeriod;
        this.returns.push(ret);

        // Window slice
        const allLen = this.returns.length;
        const wStart = (this.rollingWindow !== null && allLen > this.rollingWindow)
            ? allLen - this.rollingWindow : 0;
        const w = this.returns.slice(wStart);
        const l = w.length;
        
        this._returnsMean = mean(w);
        this._returnsStd = l > 1 ? std(w, 1) : null;
        this._returnsAutocorrPenalty = this._autocorrPenalty(w);

        const tmp1 = w.filter(r => r !== 0);
        const len1 = tmp1.length;
        this._avgReturn = len1 > 0 ? mean(tmp1) : null;
        
        const tmp2Wins = w.filter(r => r > 0);
        const len2 = tmp2Wins.length;
        this._winRate = len1 > 0 ? len2 / len1 : null;
        this._avgWin = len2 > 0 ? mean(tmp2Wins) : null;
        
        const tmp2Losses = w.filter(r => r < 0);
        const len2Losses = tmp2Losses.length;
        this._avgLoss = len2Losses > 0 ? mean(tmp2Losses) : null;

        // Excess returns (returns less risk-free rate)
        if (this.riskFreeRate === 0) {
            this._excessMean = this._returnsMean;
            this._excessStd = this._returnsStd;
            this._excessAutocorrPenalty = this._returnsAutocorrPenalty;
        } else {
            const tmp2 = w.map(r => r - this.riskFreeRate);
            this._excessMean = mean(tmp2);
            this._excessStd = l > 1 ? std(tmp2, 1) : null;
            this._excessAutocorrPenalty = this._autocorrPenalty(tmp2);
        }

        // Lower partial moments for the raw returns (less required return)
        let tmp2_lpm: number[];
        if (this.requiredReturn === 0) {
            tmp2_lpm = w.map(r => -r);
        } else {
            tmp2_lpm = w.map(r => this.requiredReturn - r);
        }
        tmp2_lpm = tmp2_lpm.map(val => Math.max(0, val));
        this._requiredLpm1 = tmp2_lpm.reduce((sum, val) => sum + val, 0) / l;
        this._requiredLpm2 = tmp2_lpm.reduce((sum, val) => sum + Math.pow(val, 2), 0) / l;
        this._requiredLpm3 = tmp2_lpm.reduce((sum, val) => sum + Math.pow(val, 3), 0) / l;

        // Higher partial moments for the raw returns (less required return)
        let tmp2_hpm: number[];
        if (this.requiredReturn === 0) {
            tmp2_hpm = [...w];
            this._requiredMean = this._returnsMean;
            this._requiredAutocorrPenalty = this._returnsAutocorrPenalty;
        } else {
            tmp2_hpm = w.map(r => r - this.requiredReturn);
            this._requiredMean = mean(tmp2_hpm);
            this._requiredAutocorrPenalty = this._autocorrPenalty(tmp2_hpm);
        }
        tmp2_hpm = tmp2_hpm.map(val => Math.max(0, val));
        this._requiredHpm1 = tmp2_hpm.reduce((sum, val) => sum + val, 0) / l;
        this._requiredHpm2 = tmp2_hpm.reduce((sum, val) => sum + Math.pow(val, 2), 0) / l;
        this._requiredHpm3 = tmp2_hpm.reduce((sum, val) => sum + Math.pow(val, 3), 0) / l;

        // Cumulative returns — recompute from window
        let logretSumVal = 0;
        for (let j = 0; j < l; j++) {
            const fpJ = this.fractionalPeriods[wStart + j];
            if (fpJ !== 0) {
                logretSumVal += Math.log(w[j] + 1);
            }
        }
        this._logretSum = logretSumVal;
        const cmr = Math.exp(logretSumVal);
        this._cumulativeReturnPlus1 = cmr;
        if (l >= 1) {
            this._cumulativeReturnGeometricMean = Math.pow(cmr, 1 / l) - 1;
        }
        this._cumulativeReturnPlus1Max = -Infinity;
        let cumr = 1;
        for (let j = 0; j < l; j++) {
            cumr *= w[j] + 1;
            if (cumr > this._cumulativeReturnPlus1Max) {
                this._cumulativeReturnPlus1Max = cumr;
            }
        }

        // Drawdowns from peaks to valleys (cumulative returns) — recompute from window
        this._drawdownsCumulative = [];
        this._drawdownsCumulativeMin = Infinity;
        cumr = 1;
        let cumrMax = -Infinity;
        for (let j = 0; j < l; j++) {
            cumr *= w[j] + 1;
            if (cumr > cumrMax) {
                cumrMax = cumr;
            }
            const dd = cumr / cumrMax - 1;
            this._drawdownsCumulative.push(dd);
            if (this._drawdownsCumulativeMin > dd) {
                this._drawdownsCumulativeMin = dd;
            }
        }

        // Drawdown peaks (used in pain index, ulcer index) — recompute from window
        this._drawdownsPeaks = [];
        this._drawdownsPeaksPeak = 0;
        for (let j = 0; j < l; j++) {
            let ddPeak = 1;
            for (let k = this._drawdownsPeaksPeak + 1; k <= j; k++) {
                ddPeak *= (1 + w[k] * 0.01);
            }
            if (ddPeak > 1) {
                this._drawdownsPeaksPeak = j;
                this._drawdownsPeaks.push(0);
            } else {
                this._drawdownsPeaks.push((ddPeak - 1) * 100);
            }
        }

        // Drawdown continuous (used in Burke ratio) — recompute from window
        this._drawdownContinuous = [];
        this._drawdownContinuousFinal = [];
        this._drawdownContinuousFinalized = false;
        this._drawdownContinuousPeak = 1;
        this._drawdownContinuousInside = false;
        for (let j = 1; j < l; j++) {
            if (w[j] < 0) {
                if (!this._drawdownContinuousInside) {
                    this._drawdownContinuousInside = true;
                    this._drawdownContinuousPeak = j - 1;
                }
                this._drawdownContinuous.push(0);
            } else {
                if (this._drawdownContinuousInside) {
                    let ddC = 1;
                    const j1 = this._drawdownContinuousPeak + 1;
                    for (let k = j1; k < j; k++) {
                        ddC *= (1 + w[k] * 0.01);
                    }
                    this._drawdownContinuous.push((ddC - 1) * 100);
                    this._drawdownContinuousInside = false;
                } else {
                    this._drawdownContinuous.push(0);
                }
            }
        }
    }

    private _autocorrPenalty(_returns: number[]): number {
        // Simplified version - returns 1 (no penalty)
        // Full implementation would require correlation calculation
        return 1;
    }

    private _isPrimed(): boolean {
        return this.minPeriods === null || this._sampleCount >= this.minPeriods;
    }

    private _windowReturns(): number[] {
        if (this.rollingWindow !== null && this.returns.length > this.rollingWindow) {
            return this.returns.slice(this.returns.length - this.rollingWindow);
        }
        return this.returns;
    }

    get cumulativeReturn(): number {
        return this._cumulativeReturnPlus1 - 1;
    }

    get drawdownsCumulative(): number[] {
        return this._drawdownsCumulative;
    }

    get minDrawdownsCumulative(): number {
        return this._drawdownsCumulativeMin;
    }

    get worstDrawdownsCumulative(): number {
        return Math.abs(this._drawdownsCumulativeMin);
    }

    drawdownsPeaks(): number[] {
        return this._drawdownsPeaks;
    }

    drawdownsContinuous(peaksOnly: boolean = false, maxPeaks: number | null = null): number[] {
        const finalize = () => {
            if (!this._drawdownContinuousFinalized) {
                const w = this._windowReturns();
                if (this._drawdownContinuousInside) {
                    let dd = 1;
                    const j1 = this._drawdownContinuousPeak + 1;
                    for (let j = j1; j < w.length; j++) {
                        dd = dd * (1 + w[j] * 0.01);
                    }
                    this._drawdownContinuousFinal = [...this._drawdownContinuous, (dd - 1) * 100];
                } else {
                    this._drawdownContinuousFinal = [...this._drawdownContinuous, 0];
                }
                this._drawdownContinuousFinalized = true;
            }
        };
        
        finalize();
        
        if (!peaksOnly) {
            return this._drawdownContinuousFinal;
        }
        
        let drawdowns = this._drawdownContinuousFinal.filter(d => d !== 0);
        if (maxPeaks !== null && drawdowns.length > 0) {
            drawdowns = drawdowns.sort((a, b) => a - b).slice(0, maxPeaks);
        }
        return drawdowns;
    }

    get skew(): number | null {
        if (!this._isPrimed()) return null;
        const w = this._windowReturns();
        return w.length > 1 ? skewness(w) : null;
    }

    get kurtosis(): number | null {
        if (!this._isPrimed()) return null;
        const w = this._windowReturns();
        return w.length > 1 ? kurtosisExcess(w) : null;
    }

    sharpeRatio(
        ignoreRiskFreeRate: boolean = false,
        autocorrelationPenalty: boolean = false
    ): number | null {
        if (!this._isPrimed()) return null;
        if (ignoreRiskFreeRate) {
            if (this._returnsMean === null || this._returnsStd === null || this._returnsStd === 0) {
                return null;
            }
            let denominator = this._returnsStd;
            if (autocorrelationPenalty) {
                denominator *= this._returnsAutocorrPenalty;
            }
            return this._returnsMean / denominator;
        } else {
            if (this._excessMean === null || this._excessStd === null || this._excessStd === 0) {
                return null;
            }
            let denominator = this._excessStd;
            if (autocorrelationPenalty) {
                denominator *= this._excessAutocorrPenalty;
            }
            return this._excessMean / denominator;
        }
    }

    sortinoRatio(
        autocorrelationPenalty: boolean = false,
        divideBySqrt2: boolean = false
    ): number | null {
        if (!this._isPrimed()) return null;
        if (this._requiredMean === null || this._requiredLpm2 === null || this._requiredLpm2 === 0) {
            return null;
        }
        let denominator = Math.sqrt(this._requiredLpm2);
        if (autocorrelationPenalty) {
            denominator *= this._requiredAutocorrPenalty;
        }
        if (divideBySqrt2) {
            denominator *= SQRT2;
        }
        return this._requiredMean / denominator;
    }

    omegaRatio(): number | null {
        if (!this._isPrimed()) return null;
        if (this._requiredMean === null || this._requiredLpm1 === null || this._requiredLpm1 === 0) {
            return null;
        }
        return this._requiredMean / this._requiredLpm1 + 1;
    }

    kappaRatio(order: number = 3): number | null {
        if (!this._isPrimed()) return null;
        if (this._requiredMean === null) {
            return null;
        }
        if (order === 1) {
            if (this._requiredLpm1 === null || this._requiredLpm1 === 0) {
                return null;
            }
            return this._requiredMean / this._requiredLpm1;
        } else if (order === 2) {
            if (this._requiredLpm2 === null || this._requiredLpm2 === 0) {
                return null;
            }
            return this._requiredMean / Math.sqrt(this._requiredLpm2);
        } else if (order === 3) {
            if (this._requiredLpm3 === null || this._requiredLpm3 === 0) {
                return null;
            }
            return this._requiredMean / Math.pow(this._requiredLpm3, 1 / 3);
        } else {
            const w = this._windowReturns();
            const tmp = w.map(r => 
                this.requiredReturn === 0 ? -r : this.requiredReturn - r
            ).map(val => Math.max(0, val));
            const lpm = tmp.reduce((sum, val) => sum + Math.pow(val, order), 0) / w.length;
            if (lpm === null || lpm === 0) {
                return null;
            }
            return this._requiredMean / Math.pow(lpm, 1 / order);
        }
    }

    kappa3Ratio(): number | null {
        if (!this._isPrimed()) return null;
        if (this._requiredMean === null || this._requiredLpm3 === null || this._requiredLpm3 === 0) {
            return null;
        }
        return this._requiredMean / Math.pow(this._requiredLpm3, 1 / 3);
    }

    bernardoLedoitRatio(): number | null {
        if (!this._isPrimed()) return null;
        const w = this._windowReturns();
        const l = w.length;
        if (l < 1) {
            return null;
        }
        const tmp = w.map(r => -r).map(val => Math.max(0, val));
        const lpm1 = tmp.reduce((sum, val) => sum + val, 0) / l;
        if (lpm1 === null || lpm1 === 0) {
            return null;
        }
        const tmp2 = w.map(val => Math.max(0, val));
        const hpm1 = tmp2.reduce((sum, val) => sum + val, 0) / l;
        return hpm1 / lpm1;
    }

    upsidePotentialRatio(full: boolean = true): number | null {
        if (!this._isPrimed()) return null;
        if (full) {
            if (this._requiredHpm1 === null || this._requiredLpm2 === null || this._requiredLpm2 === 0) {
                return null;
            }
            return this._requiredHpm1 / Math.sqrt(this._requiredLpm2);
        } else {
            const w = this._windowReturns();
            const tmpBelow = w.filter(r => r < this.requiredReturn);
            const l = tmpBelow.length;
            if (l < 1) {
                return null;
            }
            const tmp = tmpBelow.map(r => r - this.requiredReturn);
            const lpm2 = tmp.reduce((sum, val) => sum + Math.pow(val, 2), 0) / l;
            if (lpm2 === null || lpm2 === 0) {
                return null;
            }
            const tmpAbove = w.filter(r => r > this.requiredReturn);
            if (tmpAbove.length === 0) {
                return null;
            }
            const tmp2 = tmpAbove.map(r => r - this.requiredReturn);
            const hpm1 = mean(tmp2);
            if (hpm1 === null) {
                return null;
            }
            return hpm1 / Math.sqrt(lpm2);
        }
    }

    compoundGrowthRate(): number | null {
        if (!this._isPrimed()) return null;
        return this._cumulativeReturnGeometricMean;
    }

    calmarRatio(): number | null {
        if (!this._isPrimed()) return null;
        const wdd = this.worstDrawdownsCumulative;
        if (wdd === 0) {
            return null;
        }
        const cagr = this._cumulativeReturnGeometricMean;
        if (cagr === null) {
            return null;
        }
        return cagr / wdd;
    }

    sterlingRatio(annualExcessRate: number = 0): number | null {
        if (!this._isPrimed()) return null;
        const excessRate = annualExcessRate === 0 || this.periodsPerAnnum === 1
            ? annualExcessRate
            : Math.pow(1 + annualExcessRate, 1 / this.periodsPerAnnum) - 1;

        const wdd = this.worstDrawdownsCumulative + excessRate;
        if (wdd === 0) {
            return null;
        }
        const cagr = this._cumulativeReturnGeometricMean;
        if (cagr === null) {
            return null;
        }
        return cagr / wdd;
    }

    burkeRatio(modified: boolean = false): number | null {
        if (!this._isPrimed()) return null;
        const rate = this._cumulativeReturnGeometricMean !== null 
            ? this._cumulativeReturnGeometricMean - this.riskFreeRate 
            : null;
        if (rate === null) {
            return null;
        }
        const drawdowns = this.drawdownsContinuous(true);
        if (drawdowns.length < 1) {
            return null;
        }
        const sqrtSumDrawdownsSquared = Math.sqrt(
            drawdowns.reduce((sum, dd) => sum + Math.pow(dd, 2), 0)
        );
        if (sqrtSumDrawdownsSquared === 0) {
            return null;
        }
        let burke = rate / sqrtSumDrawdownsSquared;
        if (modified) {
            burke *= Math.sqrt(this._windowReturns().length);
        }
        return burke;
    }

    painIndex(): number | null {
        if (!this._isPrimed()) return null;
        const l = this._drawdownsPeaks.length;
        if (l < 1) {
            return null;
        }
        return -this._drawdownsPeaks.reduce((sum, dd) => sum + dd, 0) / l;
    }

    painRatio(): number | null {
        if (!this._isPrimed()) return null;
        const rate = this._cumulativeReturnGeometricMean !== null 
            ? this._cumulativeReturnGeometricMean - this.riskFreeRate 
            : null;
        if (rate === null) {
            return null;
        }
        const l = this._drawdownsPeaks.length;
        if (l < 1) {
            return null;
        }
        const painIndex = -this._drawdownsPeaks.reduce((sum, dd) => sum + dd, 0) / l;
        return painIndex !== 0 ? rate / painIndex : null;
    }

    ulcerIndex(): number | null {
        if (!this._isPrimed()) return null;
        const l = this._drawdownsPeaks.length;
        if (l < 1) {
            return null;
        }
        return Math.sqrt(
            this._drawdownsPeaks.reduce((sum, dd) => sum + Math.pow(dd, 2), 0) / l
        );
    }

    martinRatio(): number | null {
        if (!this._isPrimed()) return null;
        const rate = this._cumulativeReturnGeometricMean !== null 
            ? this._cumulativeReturnGeometricMean - this.riskFreeRate 
            : null;
        if (rate === null) {
            return null;
        }
        const l = this._drawdownsPeaks.length;
        if (l < 1) {
            return null;
        }
        const ulcerIndex = Math.sqrt(
            this._drawdownsPeaks.reduce((sum, dd) => sum + Math.pow(dd, 2), 0) / l
        );
        return ulcerIndex !== 0 ? rate / ulcerIndex : null;
    }

    get logretSum(): number {
        return this._logretSum;
    }

    get requiredHpm2(): number | null {
        return this._requiredHpm2;
    }

    get requiredHpm3(): number | null {
        return this._requiredHpm3;
    }

    get avgReturn(): number | null {
        return this._avgReturn;
    }

    get avgWin(): number | null {
        return this._avgWin;
    }

    get avgLoss(): number | null {
        return this._avgLoss;
    }

    get gainToPainRatio(): number | null {
        if (!this._isPrimed()) return null;
        return this._requiredLpm1 !== null && this._requiredLpm1 !== 0 && this._returnsMean !== null
            ? this._returnsMean / this._requiredLpm1
            : null;
    }

    get riskOfRuin(): number | null {
        if (!this._isPrimed()) return null;
        const wr = this._winRate;
        if (wr === null) {
            return null;
        }
        return Math.pow((1 - wr) / (1 + wr), this._windowReturns().length);
    }

    get riskReturnRatio(): number | null {
        if (!this._isPrimed()) return null;
        if (this._returnsMean === null || this._returnsStd === null || this._returnsStd === 0) {
            return null;
        }
        return this._returnsMean / this._returnsStd;
    }
}
