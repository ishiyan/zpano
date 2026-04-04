import { DayCountConvention, frac } from '../daycounting';
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

    public fractionalPeriods: number[] = [];
    public returns: number[] = [];

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
        dayCountConvention: DayCountConvention = DayCountConvention.RAW
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
    }

    reset(): void {
        this.fractionalPeriods = [];
        this.returns = [];
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
        returnBenchmark: number,
        value: number,
        timeStart: Date,
        timeEnd: Date
    ): void {
        let fractionalPeriod: number;
        
        if (this.periodicity === Periodicity.ANNUAL) {
            fractionalPeriod = frac(timeStart, timeEnd, this.dayCountConvention, false);
        } else {
            fractionalPeriod = frac(timeStart, timeEnd, this.dayCountConvention, true) / this.daysPerPeriod;
        }

        this.fractionalPeriods.push(fractionalPeriod);
        
        if (fractionalPeriod === 0) {
            console.warn('Zero fractional time period, performance not updated');
            return;
        }
        
        this._totalDuration += fractionalPeriod;

        // Normalized returns
        const ret = return_ / fractionalPeriod;
        this.returns.push(ret);
        const l = this.returns.length;
        
        this._returnsMean = mean(this.returns);
        this._returnsStd = l > 1 ? std(this.returns, 1) : null;
        this._returnsAutocorrPenalty = this._autocorrPenalty(this.returns);

        const tmp1 = this.returns.filter(r => r !== 0);
        const len1 = tmp1.length;
        this._avgReturn = len1 > 0 ? mean(tmp1) : null;
        
        const tmp2Wins = this.returns.filter(r => r > 0);
        const len2 = tmp2Wins.length;
        this._winRate = len1 > 0 ? len2 / len1 : null;
        this._avgWin = len2 > 0 ? mean(tmp2Wins) : null;
        
        const tmp2Losses = this.returns.filter(r => r < 0);
        const len2Losses = tmp2Losses.length;
        this._avgLoss = len2Losses > 0 ? mean(tmp2Losses) : null;

        // Excess returns (returns less risk-free rate)
        if (this.riskFreeRate === 0) {
            this._excessMean = this._returnsMean;
            this._excessStd = this._returnsStd;
            this._excessAutocorrPenalty = this._returnsAutocorrPenalty;
        } else {
            const tmp2 = this.returns.map(r => r - this.riskFreeRate);
            this._excessMean = mean(tmp2);
            this._excessStd = l > 1 ? std(tmp2, 1) : null;
            this._excessAutocorrPenalty = this._autocorrPenalty(tmp2);
        }

        // Lower partial moments for the raw returns (less required return)
        let tmp2: number[];
        if (this.requiredReturn === 0) {
            tmp2 = this.returns.map(r => -r);
        } else {
            tmp2 = this.returns.map(r => this.requiredReturn - r);
        }
        // Set the minimum of each to 0
        tmp2 = tmp2.map(val => Math.max(0, val));
        // Calculate the sum of the excess returns to the power of order
        this._requiredLpm1 = tmp2.reduce((sum, val) => sum + val, 0) / l;
        this._requiredLpm2 = tmp2.reduce((sum, val) => sum + Math.pow(val, 2), 0) / l;
        this._requiredLpm3 = tmp2.reduce((sum, val) => sum + Math.pow(val, 3), 0) / l;

        // Higher partial moments for the raw returns (less required return)
        if (this.requiredReturn === 0) {
            tmp2 = [...this.returns];
            this._requiredMean = this._returnsMean;
            this._requiredAutocorrPenalty = this._returnsAutocorrPenalty;
        } else {
            tmp2 = this.returns.map(r => r - this.requiredReturn);
            this._requiredMean = mean(tmp2);
            this._requiredAutocorrPenalty = this._autocorrPenalty(tmp2);
        }
        // Set the minimum of each to 0
        tmp2 = tmp2.map(val => Math.max(0, val));
        // Calculate the sum of the excess returns to the power of order
        this._requiredHpm1 = tmp2.reduce((sum, val) => sum + val, 0) / l;
        this._requiredHpm2 = tmp2.reduce((sum, val) => sum + Math.pow(val, 2), 0) / l;
        this._requiredHpm3 = tmp2.reduce((sum, val) => sum + Math.pow(val, 3), 0) / l;

        // Cumulative returns
        const retlog = Math.log(return_ + 1) / fractionalPeriod;
        this._logretSum += retlog;
        const ret1 = ret + 1;
        
        let cmr: number;
        if (l === 1) {
            cmr = ret1;
            this._cumulativeReturnPlus1 = ret1;
            this._cumulativeReturnGeometricMean = ret;
        } else {
            cmr = Math.exp(this._logretSum);
            this._cumulativeReturnPlus1 = cmr;
            this._cumulativeReturnGeometricMean = Math.pow(cmr, 1 / l) - 1;
        }
        
        if (this._cumulativeReturnPlus1Max < cmr) {
            this._cumulativeReturnPlus1Max = cmr;
        }

        // Drawdowns from peaks to valleys, operates on cumulative returns
        let dd = cmr / this._cumulativeReturnPlus1Max - 1;
        if (this._drawdownsCumulativeMin > dd) {
            this._drawdownsCumulativeMin = dd;
        }
        this._drawdownsCumulative.push(dd);
        
        // Different drawdown calculation used in pain index, ulcer index
        dd = 1;
        for (let j = this._drawdownsPeaksPeak + 1; j < l; j++) {
            dd *= (1 + this.returns[j] * 0.01);
        }
        if (dd > 1) {
            this._drawdownsPeaksPeak = l - 1;
            this._drawdownsPeaks.push(0);
        } else {
            this._drawdownsPeaks.push((dd - 1) * 100);
        }
        
        // Drawdown calculation used in Burke
        if (l > 1) {
            this._drawdownContinuousFinalized = false;
            if (ret < 0) {
                if (!this._drawdownContinuousInside) {
                    this._drawdownContinuousInside = true;
                    this._drawdownContinuousPeak = l - 2;
                }
                this._drawdownContinuous.push(0);
            } else {
                if (this._drawdownContinuousInside) {
                    dd = 1;
                    const j1 = this._drawdownContinuousPeak + 1;
                    for (let j = j1; j < l - 1; j++) {
                        dd = dd * (1 + this.returns[j] * 0.01);
                    }
                    this._drawdownContinuous.push((dd - 1) * 100);
                    this._drawdownContinuousInside = false;
                } else {
                    this._drawdownContinuous.push(0);
                }
            }
        }
    }

    private _autocorrPenalty(returns: number[]): number {
        // Simplified version - returns 1 (no penalty)
        // Full implementation would require correlation calculation
        return 1;
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
                if (this._drawdownContinuousInside) {
                    let dd = 1;
                    const j1 = this._drawdownContinuousPeak + 1;
                    for (let j = j1; j < this.returns.length; j++) {
                        dd = dd * (1 + this.returns[j] * 0.01);
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
        return this.returns.length > 1 ? skewness(this.returns) : null;
    }

    get kurtosis(): number | null {
        return this.returns.length > 1 ? kurtosisExcess(this.returns) : null;
    }

    sharpeRatio(
        ignoreRiskFreeRate: boolean = false,
        autocorrelationPenalty: boolean = false
    ): number | null {
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
        if (this._requiredMean === null || this._requiredLpm1 === null || this._requiredLpm1 === 0) {
            return null;
        }
        return this._requiredMean / this._requiredLpm1 + 1;
    }

    kappaRatio(order: number = 3): number | null {
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
            const tmp = this.returns.map(r => 
                this.requiredReturn === 0 ? -r : this.requiredReturn - r
            ).map(val => Math.max(0, val));
            const lpm = tmp.reduce((sum, val) => sum + Math.pow(val, order), 0) / this.returns.length;
            if (lpm === null || lpm === 0) {
                return null;
            }
            return this._requiredMean / Math.pow(lpm, 1 / order);
        }
    }

    kappa3Ratio(): number | null {
        if (this._requiredMean === null || this._requiredLpm3 === null || this._requiredLpm3 === 0) {
            return null;
        }
        return this._requiredMean / Math.pow(this._requiredLpm3, 1 / 3);
    }

    bernardoLedoitRatio(): number | null {
        const l = this.returns.length;
        if (l < 1) {
            return null;
        }
        const tmp = this.returns.map(r => -r).map(val => Math.max(0, val));
        const lpm1 = tmp.reduce((sum, val) => sum + val, 0) / l;
        if (lpm1 === null || lpm1 === 0) {
            return null;
        }
        const tmp2 = this.returns.map(val => Math.max(0, val));
        const hpm1 = tmp2.reduce((sum, val) => sum + val, 0) / l;
        return hpm1 / lpm1;
    }

    upsidePotentialRatio(full: boolean = true): number | null {
        if (full) {
            if (this._requiredHpm1 === null || this._requiredLpm2 === null || this._requiredLpm2 === 0) {
                return null;
            }
            return this._requiredHpm1 / Math.sqrt(this._requiredLpm2);
        } else {
            const tmpBelow = this.returns.filter(r => r < this.requiredReturn);
            const l = tmpBelow.length;
            if (l < 1) {
                return null;
            }
            const tmp = tmpBelow.map(r => r - this.requiredReturn);
            const lpm2 = tmp.reduce((sum, val) => sum + Math.pow(val, 2), 0) / l;
            if (lpm2 === null || lpm2 === 0) {
                return null;
            }
            const tmpAbove = this.returns.filter(r => r > this.requiredReturn);
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
        return this._cumulativeReturnGeometricMean;
    }

    calmarRatio(): number | null {
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
            burke *= Math.sqrt(this.returns.length);
        }
        return burke;
    }

    painIndex(): number | null {
        const l = this._drawdownsPeaks.length;
        if (l < 1) {
            return null;
        }
        return -this._drawdownsPeaks.reduce((sum, dd) => sum + dd, 0) / l;
    }

    painRatio(): number | null {
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
        const l = this._drawdownsPeaks.length;
        if (l < 1) {
            return null;
        }
        return Math.sqrt(
            this._drawdownsPeaks.reduce((sum, dd) => sum + Math.pow(dd, 2), 0) / l
        );
    }

    martinRatio(): number | null {
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

    get gainToPainRatio(): number | null {
        return this._requiredLpm1 !== null && this._requiredLpm1 !== 0 && this._returnsMean !== null
            ? this._returnsMean / this._requiredLpm1
            : null;
    }

    get riskOfRuin(): number | null {
        const wr = this._winRate;
        if (wr === null) {
            return null;
        }
        return Math.pow((1 - wr) / (1 + wr), this.returns.length);
    }

    get riskReturnRatio(): number | null {
        if (this._returnsMean === null || this._returnsStd === null || this._returnsStd === 0) {
            return null;
        }
        return this._returnsMean / this._returnsStd;
    }
}
