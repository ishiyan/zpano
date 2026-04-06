import { RoundtripSide } from './side';
import { Roundtrip } from './roundtrip';
import { DayCountConvention, yearFrac } from '../daycounting/index';

// ---- helpers ----

function mean(arr: number[]): number {
    let s = 0;
    for (let i = 0; i < arr.length; i++) s += arr[i];
    return s / arr.length;
}

function stdPop(arr: number[]): number {
    const m = mean(arr);
    let s = 0;
    for (let i = 0; i < arr.length; i++) {
        const d = arr[i] - m;
        s += d * d;
    }
    return Math.sqrt(s / arr.length);
}

function maxConsecutive(arr: boolean[]): number {
    let max = 0;
    let cur = 0;
    for (let i = 0; i < arr.length; i++) {
        if (arr[i]) {
            cur++;
            if (cur > max) max = cur;
        } else {
            cur = 0;
        }
    }
    return max;
}

export class RoundtripPerformance {
    readonly initialBalance: number;
    readonly annualRiskFreeRate: number;
    readonly annualTargetReturn: number;
    readonly dayCountConvention: DayCountConvention;

    roundtrips: Roundtrip[] = [];
    returnsOnInvestments: number[] = [];
    sortinoDownsideReturns: number[] = [];
    returnsOnInvestmentsAnnual: number[] = [];
    sortinoDownsideReturnsAnnual: number[] = [];

    firstTime: Date | null = null;
    lastTime: Date | null = null;
    maxNetPnl: number = 0;
    maxDrawdown: number = 0;
    maxDrawdownPercent: number = 0;

    totalCommission: number = 0;
    grossWinningCommission: number = 0;
    grossLoosingCommission: number = 0;
    netWinningCommission: number = 0;
    netLoosingCommission: number = 0;
    grossWinningLongCommission: number = 0;
    grossLoosingLongCommission: number = 0;
    netWinningLongCommission: number = 0;
    netLoosingLongCommission: number = 0;
    grossWinningShortCommission: number = 0;
    grossLoosingShortCommission: number = 0;
    netWinningShortCommission: number = 0;
    netLoosingShortCommission: number = 0;

    private _netPnl: number = 0;
    private _grossPnl: number = 0;
    private _grossWinningPnl: number = 0;
    private _grossLoosingPnl: number = 0;
    private _netWinningPnl: number = 0;
    private _netLoosingPnl: number = 0;
    private _grossLongPnl: number = 0;
    private _grossShortPnl: number = 0;
    private _netLongPnl: number = 0;
    private _netShortPnl: number = 0;
    private _grossLongWinningPnl: number = 0;
    private _grossLongLoosingPnl: number = 0;
    private _netLongWinningPnl: number = 0;
    private _netLongLoosingPnl: number = 0;
    private _grossShortWinningPnl: number = 0;
    private _grossShortLoosingPnl: number = 0;
    private _netShortWinningPnl: number = 0;
    private _netShortLoosingPnl: number = 0;

    private _totalCount: number = 0;
    private _longCount: number = 0;
    private _shortCount: number = 0;
    private _grossWinningCount: number = 0;
    private _grossLoosingCount: number = 0;
    private _netWinningCount: number = 0;
    private _netLoosingCount: number = 0;
    private _grossLongWinningCount: number = 0;
    private _grossLongLoosingCount: number = 0;
    private _netLongWinningCount: number = 0;
    private _netLongLoosingCount: number = 0;
    private _grossShortWinningCount: number = 0;
    private _grossShortLoosingCount: number = 0;
    private _netShortWinningCount: number = 0;
    private _netShortLoosingCount: number = 0;

    private _durationSec: number = 0;
    private _durationSecLong: number = 0;
    private _durationSecShort: number = 0;
    private _durationSecGrossWinning: number = 0;
    private _durationSecGrossLoosing: number = 0;
    private _durationSecNetWinning: number = 0;
    private _durationSecNetLoosing: number = 0;
    private _durationSecGrossLongWinning: number = 0;
    private _durationSecGrossLongLoosing: number = 0;
    private _durationSecNetLongWinning: number = 0;
    private _durationSecNetLongLoosing: number = 0;
    private _durationSecGrossShortWinning: number = 0;
    private _durationSecGrossShortLoosing: number = 0;
    private _durationSecNetShortWinning: number = 0;
    private _durationSecNetShortLoosing: number = 0;
    private _totalDurationAnnualized: number = 0;

    private _totalMae: number = 0;
    private _totalMfe: number = 0;
    private _totalEff: number = 0;
    private _totalEffEntry: number = 0;
    private _totalEffExit: number = 0;

    private _roiMean: number | null = null;
    private _roiStd: number | null = null;
    private _roiTdd: number | null = null;
    private _roiannMean: number | null = null;
    private _roiannStd: number | null = null;
    private _roiannTdd: number | null = null;

    constructor(
        initialBalance: number = 100000.0,
        annualRiskFreeRate: number = 0.0,
        annualTargetReturn: number = 0.0,
        dayCountConvention: DayCountConvention = DayCountConvention.RAW,
    ) {
        this.initialBalance = initialBalance;
        this.annualRiskFreeRate = annualRiskFreeRate;
        this.annualTargetReturn = annualTargetReturn;
        this.dayCountConvention = dayCountConvention;
    }

    reset(): void {
        this.roundtrips = [];
        this.returnsOnInvestments = [];
        this.sortinoDownsideReturns = [];
        this.returnsOnInvestmentsAnnual = [];
        this.sortinoDownsideReturnsAnnual = [];

        this.firstTime = null;
        this.lastTime = null;
        this.maxNetPnl = 0;
        this.maxDrawdown = 0;
        this.maxDrawdownPercent = 0;

        this.totalCommission = 0;
        this.grossWinningCommission = 0;
        this.grossLoosingCommission = 0;
        this.netWinningCommission = 0;
        this.netLoosingCommission = 0;
        this.grossWinningLongCommission = 0;
        this.grossLoosingLongCommission = 0;
        this.netWinningLongCommission = 0;
        this.netLoosingLongCommission = 0;
        this.grossWinningShortCommission = 0;
        this.grossLoosingShortCommission = 0;
        this.netWinningShortCommission = 0;
        this.netLoosingShortCommission = 0;

        this._netPnl = 0;
        this._grossPnl = 0;
        this._grossWinningPnl = 0;
        this._grossLoosingPnl = 0;
        this._netWinningPnl = 0;
        this._netLoosingPnl = 0;
        this._grossLongPnl = 0;
        this._grossShortPnl = 0;
        this._netLongPnl = 0;
        this._netShortPnl = 0;
        this._grossLongWinningPnl = 0;
        this._grossLongLoosingPnl = 0;
        this._netLongWinningPnl = 0;
        this._netLongLoosingPnl = 0;
        this._grossShortWinningPnl = 0;
        this._grossShortLoosingPnl = 0;
        this._netShortWinningPnl = 0;
        this._netShortLoosingPnl = 0;

        this._totalCount = 0;
        this._longCount = 0;
        this._shortCount = 0;
        this._grossWinningCount = 0;
        this._grossLoosingCount = 0;
        this._netWinningCount = 0;
        this._netLoosingCount = 0;
        this._grossLongWinningCount = 0;
        this._grossLongLoosingCount = 0;
        this._netLongWinningCount = 0;
        this._netLongLoosingCount = 0;
        this._grossShortWinningCount = 0;
        this._grossShortLoosingCount = 0;
        this._netShortWinningCount = 0;
        this._netShortLoosingCount = 0;

        this._durationSec = 0;
        this._durationSecLong = 0;
        this._durationSecShort = 0;
        this._durationSecGrossWinning = 0;
        this._durationSecGrossLoosing = 0;
        this._durationSecNetWinning = 0;
        this._durationSecNetLoosing = 0;
        this._durationSecGrossLongWinning = 0;
        this._durationSecGrossLongLoosing = 0;
        this._durationSecNetLongWinning = 0;
        this._durationSecNetLongLoosing = 0;
        this._durationSecGrossShortWinning = 0;
        this._durationSecGrossShortLoosing = 0;
        this._durationSecNetShortWinning = 0;
        this._durationSecNetShortLoosing = 0;
        this._totalDurationAnnualized = 0;

        this._totalMae = 0;
        this._totalMfe = 0;
        this._totalEff = 0;
        this._totalEffEntry = 0;
        this._totalEffExit = 0;

        this._roiMean = null;
        this._roiStd = null;
        this._roiTdd = null;
        this._roiannMean = null;
        this._roiannStd = null;
        this._roiannTdd = null;
    }

    addRoundtrip(roundtrip: Roundtrip): void {
        this.roundtrips.push(roundtrip);
        this._totalCount += 1;
        const comm = roundtrip.commission;
        this.totalCommission += comm;
        const secs = roundtrip.durationSeconds;
        this._durationSec += secs;
        this._totalMae += roundtrip.maximumAdverseExcursion;
        this._totalMfe += roundtrip.maximumFavorableExcursion;
        this._totalEff += roundtrip.totalEfficiency;
        this._totalEffEntry += roundtrip.entryEfficiency;
        this._totalEffExit += roundtrip.exitEfficiency;

        const netPnl = roundtrip.netPnl;
        this._netPnl += netPnl;
        if (netPnl > 0) {
            this._netWinningCount += 1;
            this._netWinningPnl += netPnl;
            this.netWinningCommission += comm;
            this._durationSecNetWinning += secs;
        } else if (netPnl < 0) {
            this._netLoosingCount += 1;
            this._netLoosingPnl += netPnl;
            this.netLoosingCommission += comm;
            this._durationSecNetLoosing += secs;
        }

        const grossPnl = roundtrip.grossPnl;
        this._grossPnl += grossPnl;
        if (grossPnl > 0) {
            this._grossWinningCount += 1;
            this._grossWinningPnl += grossPnl;
            this.grossWinningCommission += comm;
            this._durationSecGrossWinning += secs;
        } else if (grossPnl < 0) {
            this._grossLoosingCount += 1;
            this._grossLoosingPnl += grossPnl;
            this.grossLoosingCommission += comm;
            this._durationSecGrossLoosing += secs;
        }

        if (roundtrip.side === RoundtripSide.LONG) {
            this._grossLongPnl += grossPnl;
            this._netLongPnl += netPnl;
            this._longCount += 1;
            this._durationSecLong += secs;
            if (grossPnl > 0) {
                this._grossLongWinningCount += 1;
                this._grossLongWinningPnl += grossPnl;
                this.grossWinningLongCommission += comm;
                this._durationSecGrossLongWinning += secs;
            } else if (grossPnl < 0) {
                this._grossLongLoosingCount += 1;
                this._grossLongLoosingPnl += grossPnl;
                this.grossLoosingLongCommission += comm;
                this._durationSecGrossLongLoosing += secs;
            }
            // CRITICAL: net_long_winning/loosing_pnl adds gross_pnl, not net_pnl
            if (netPnl > 0) {
                this._netLongWinningCount += 1;
                this._netLongWinningPnl += grossPnl;
                this.netWinningLongCommission += comm;
                this._durationSecNetLongWinning += secs;
            } else if (netPnl < 0) {
                this._netLongLoosingCount += 1;
                this._netLongLoosingPnl += grossPnl;
                this.netLoosingLongCommission += comm;
                this._durationSecNetLongLoosing += secs;
            }
        } else {
            this._grossShortPnl += grossPnl;
            this._netShortPnl += netPnl;
            this._shortCount += 1;
            this._durationSecShort += secs;
            if (grossPnl > 0) {
                this._grossShortWinningCount += 1;
                this._grossShortWinningPnl += grossPnl;
                this.grossWinningShortCommission += comm;
                this._durationSecGrossShortWinning += secs;
            } else if (grossPnl < 0) {
                this._grossShortLoosingCount += 1;
                this._grossShortLoosingPnl += grossPnl;
                this.grossLoosingShortCommission += comm;
                this._durationSecGrossShortLoosing += secs;
            }
            // CRITICAL: net_short_winning/loosing_pnl adds gross_pnl, not net_pnl
            if (netPnl > 0) {
                this._netShortWinningCount += 1;
                this._netShortWinningPnl += grossPnl;
                this.netWinningShortCommission += comm;
                this._durationSecNetShortWinning += secs;
            } else if (netPnl < 0) {
                this._netShortLoosingCount += 1;
                this._netShortLoosingPnl += grossPnl;
                this.netLoosingShortCommission += comm;
                this._durationSecNetShortLoosing += secs;
            }
        }

        // Update first/last times and duration
        let changed = false;
        if (this.firstTime === null || this.firstTime > roundtrip.entryTime) {
            this.firstTime = roundtrip.entryTime;
            changed = true;
        }
        if (this.lastTime === null || this.lastTime < roundtrip.exitTime) {
            this.lastTime = roundtrip.exitTime;
            changed = true;
        }
        if (changed) {
            this._totalDurationAnnualized = yearFrac(
                this.firstTime, this.lastTime, this.dayCountConvention);
        }

        const roi = netPnl / (roundtrip.quantity * roundtrip.entryPrice);
        this.returnsOnInvestments.push(roi);
        this._roiMean = mean(this.returnsOnInvestments);
        this._roiStd = stdPop(this.returnsOnInvestments);
        const downside = roi - this.annualRiskFreeRate;
        if (downside < 0) {
            this.sortinoDownsideReturns.push(downside);
            let sumSq = 0;
            for (let i = 0; i < this.sortinoDownsideReturns.length; i++) {
                sumSq += this.sortinoDownsideReturns[i] * this.sortinoDownsideReturns[i];
            }
            this._roiTdd = Math.sqrt(sumSq / this.sortinoDownsideReturns.length);
        }

        // Annualized ROI
        const d = yearFrac(roundtrip.entryTime, roundtrip.exitTime,
                           this.dayCountConvention);
        if (d !== 0) {
            const roiann = roi / d;
            this.returnsOnInvestmentsAnnual.push(roiann);
            this._roiannMean = mean(this.returnsOnInvestmentsAnnual);
            this._roiannStd = stdPop(this.returnsOnInvestmentsAnnual);
            const downsideAnn = roiann - this.annualRiskFreeRate;
            if (downsideAnn < 0) {
                this.sortinoDownsideReturnsAnnual.push(downsideAnn);
                let sumSq = 0;
                for (let i = 0; i < this.sortinoDownsideReturnsAnnual.length; i++) {
                    sumSq += this.sortinoDownsideReturnsAnnual[i] * this.sortinoDownsideReturnsAnnual[i];
                }
                this._roiannTdd = Math.sqrt(sumSq / this.sortinoDownsideReturnsAnnual.length);
            }
        }

        // Max drawdown
        if (this.maxNetPnl < this._netPnl) {
            this.maxNetPnl = this._netPnl;
        }
        const dd = this.maxNetPnl - this._netPnl;
        if (this.maxDrawdown < dd) {
            this.maxDrawdown = dd;
            this.maxDrawdownPercent = this.maxDrawdown /
                (this.initialBalance + this.maxNetPnl);
        }
    }

    // ====================== ROI statistics ======================

    get roiMean(): number | null { return this._roiMean; }
    get roiStd(): number | null { return this._roiStd; }
    get roiTdd(): number | null { return this._roiTdd; }
    get roiannMean(): number | null { return this._roiannMean; }
    get roiannStd(): number | null { return this._roiannStd; }
    get roiannTdd(): number | null { return this._roiannTdd; }

    // ====================== risk-adjusted ratios ======================

    get sharpeRatio(): number | null {
        if (this._roiMean === null || this._roiStd === null || this._roiStd === 0) return null;
        return this._roiMean / this._roiStd;
    }

    get sharpeRatioAnnual(): number | null {
        if (this._roiannMean === null || this._roiannStd === null || this._roiannStd === 0) return null;
        return this._roiannMean / this._roiannStd;
    }

    get sortinoRatio(): number | null {
        if (this._roiMean === null || this._roiTdd === null || this._roiTdd === 0) return null;
        return (this._roiMean - this.annualRiskFreeRate) / this._roiTdd;
    }

    get sortinoRatioAnnual(): number | null {
        if (this._roiannMean === null || this._roiannTdd === null || this._roiannTdd === 0) return null;
        return (this._roiannMean - this.annualRiskFreeRate) / this._roiannTdd;
    }

    get calmarRatio(): number | null {
        if (this._roiMean === null || this.maxDrawdownPercent === 0) return null;
        return this._roiMean / this.maxDrawdownPercent;
    }

    get calmarRatioAnnual(): number | null {
        if (this._roiannMean === null || this.maxDrawdownPercent === 0) return null;
        return this._roiannMean / this.maxDrawdownPercent;
    }

    // ====================== rate of return ======================

    get rateOfReturn(): number | null {
        if (this.initialBalance === 0) return null;
        return this._netPnl / this.initialBalance;
    }

    get rateOfReturnAnnual(): number | null {
        if (this._totalDurationAnnualized === 0 || this.initialBalance === 0) return null;
        return (this._netPnl / this.initialBalance) / this._totalDurationAnnualized;
    }

    get recoveryFactor(): number | null {
        const rorann = this.rateOfReturnAnnual;
        if (rorann === null || this.maxDrawdownPercent === 0) return null;
        return rorann / this.maxDrawdownPercent;
    }

    // ====================== profit ratios ======================

    get grossProfitRatio(): number | null {
        return this._grossLoosingPnl !== 0
            ? Math.abs(this._grossWinningPnl / this._grossLoosingPnl) : null;
    }

    get netProfitRatio(): number | null {
        return this._netLoosingPnl !== 0
            ? Math.abs(this._netWinningPnl / this._netLoosingPnl) : null;
    }

    get grossProfitLongRatio(): number | null {
        return this._grossLongLoosingPnl !== 0
            ? Math.abs(this._grossLongWinningPnl / this._grossLongLoosingPnl) : null;
    }

    get netProfitLongRatio(): number | null {
        return this._netLongLoosingPnl !== 0
            ? Math.abs(this._netLongWinningPnl / this._netLongLoosingPnl) : null;
    }

    get grossProfitShortRatio(): number | null {
        return this._grossShortLoosingPnl !== 0
            ? Math.abs(this._grossShortWinningPnl / this._grossShortLoosingPnl) : null;
    }

    get netProfitShortRatio(): number | null {
        return this._netShortLoosingPnl !== 0
            ? Math.abs(this._netShortWinningPnl / this._netShortLoosingPnl) : null;
    }

    // ====================== counts ======================

    get totalCount(): number { return this._totalCount; }
    get longCount(): number { return this._longCount; }
    get shortCount(): number { return this._shortCount; }
    get grossWinningCount(): number { return this._grossWinningCount; }
    get grossLoosingCount(): number { return this._grossLoosingCount; }
    get netWinningCount(): number { return this._netWinningCount; }
    get netLoosingCount(): number { return this._netLoosingCount; }
    get grossLongWinningCount(): number { return this._grossLongWinningCount; }
    get grossLongLoosingCount(): number { return this._grossLongLoosingCount; }
    get netLongWinningCount(): number { return this._netLongWinningCount; }
    get netLongLoosingCount(): number { return this._netLongLoosingCount; }
    get grossShortWinningCount(): number { return this._grossShortWinningCount; }
    get grossShortLoosingCount(): number { return this._grossShortLoosingCount; }
    get netShortWinningCount(): number { return this._netShortWinningCount; }
    get netShortLoosingCount(): number { return this._netShortLoosingCount; }

    // ====================== winning/loosing ratios ======================

    get grossWinningRatio(): number {
        return this._totalCount > 0 ? this._grossWinningCount / this._totalCount : 0.0;
    }
    get grossLoosingRatio(): number {
        return this._totalCount > 0 ? this._grossLoosingCount / this._totalCount : 0.0;
    }
    get netWinningRatio(): number {
        return this._totalCount > 0 ? this._netWinningCount / this._totalCount : 0.0;
    }
    get netLoosingRatio(): number {
        return this._totalCount > 0 ? this._netLoosingCount / this._totalCount : 0.0;
    }
    get grossLongWinningRatio(): number {
        return this._longCount > 0 ? this._grossLongWinningCount / this._longCount : 0.0;
    }
    get grossLongLoosingRatio(): number {
        return this._longCount > 0 ? this._grossLongLoosingCount / this._longCount : 0.0;
    }
    get netLongWinningRatio(): number {
        return this._longCount > 0 ? this._netLongWinningCount / this._longCount : 0.0;
    }
    get netLongLoosingRatio(): number {
        return this._longCount > 0 ? this._netLongLoosingCount / this._longCount : 0.0;
    }
    get grossShortWinningRatio(): number {
        return this._shortCount > 0 ? this._grossShortWinningCount / this._shortCount : 0.0;
    }
    get grossShortLoosingRatio(): number {
        return this._shortCount > 0 ? this._grossShortLoosingCount / this._shortCount : 0.0;
    }
    get netShortWinningRatio(): number {
        return this._shortCount > 0 ? this._netShortWinningCount / this._shortCount : 0.0;
    }
    get netShortLoosingRatio(): number {
        return this._shortCount > 0 ? this._netShortLoosingCount / this._shortCount : 0.0;
    }

    // ====================== PnL totals ======================

    get totalGrossPnl(): number { return this._grossPnl; }
    get totalNetPnl(): number { return this._netPnl; }
    get winningGrossPnl(): number { return this._grossWinningPnl; }
    get loosingGrossPnl(): number { return this._grossLoosingPnl; }
    get winningNetPnl(): number { return this._netWinningPnl; }
    get loosingNetPnl(): number { return this._netLoosingPnl; }
    get winningGrossLongPnl(): number { return this._grossLongWinningPnl; }
    get loosingGrossLongPnl(): number { return this._grossLongLoosingPnl; }
    get winningNetLongPnl(): number { return this._netLongWinningPnl; }
    get loosingNetLongPnl(): number { return this._netLongLoosingPnl; }
    get winningGrossShortPnl(): number { return this._grossShortWinningPnl; }
    get loosingGrossShortPnl(): number { return this._grossShortLoosingPnl; }
    get winningNetShortPnl(): number { return this._netShortWinningPnl; }
    get loosingNetShortPnl(): number { return this._netShortLoosingPnl; }

    // ====================== average PnL ======================

    get averageGrossPnl(): number {
        return this._totalCount > 0 ? this._grossPnl / this._totalCount : 0.0;
    }
    get averageNetPnl(): number {
        return this._totalCount > 0 ? this._netPnl / this._totalCount : 0.0;
    }
    get averageGrossLongPnl(): number {
        return this._longCount > 0 ? this._grossLongPnl / this._longCount : 0.0;
    }
    get averageNetLongPnl(): number {
        return this._longCount > 0 ? this._netLongPnl / this._longCount : 0.0;
    }
    get averageGrossShortPnl(): number {
        return this._shortCount > 0 ? this._grossShortPnl / this._shortCount : 0.0;
    }
    get averageNetShortPnl(): number {
        return this._shortCount > 0 ? this._netShortPnl / this._shortCount : 0.0;
    }
    get averageWinningGrossPnl(): number {
        return this._grossWinningCount > 0 ? this._grossWinningPnl / this._grossWinningCount : 0.0;
    }
    get averageLoosingGrossPnl(): number {
        return this._grossLoosingCount > 0 ? this._grossLoosingPnl / this._grossLoosingCount : 0.0;
    }
    get averageWinningNetPnl(): number {
        return this._netWinningCount > 0 ? this._netWinningPnl / this._netWinningCount : 0.0;
    }
    get averageLoosingNetPnl(): number {
        return this._netLoosingCount > 0 ? this._netLoosingPnl / this._netLoosingCount : 0.0;
    }
    get averageWinningGrossLongPnl(): number {
        return this._grossLongWinningCount > 0 ? this._grossLongWinningPnl / this._grossLongWinningCount : 0.0;
    }
    get averageLoosingGrossLongPnl(): number {
        return this._grossLongLoosingCount > 0 ? this._grossLongLoosingPnl / this._grossLongLoosingCount : 0.0;
    }
    get averageWinningNetLongPnl(): number {
        return this._netLongWinningCount > 0 ? this._netLongWinningPnl / this._netLongWinningCount : 0.0;
    }
    get averageLoosingNetLongPnl(): number {
        return this._netLongLoosingCount > 0 ? this._netLongLoosingPnl / this._netLongLoosingCount : 0.0;
    }
    get averageWinningGrossShortPnl(): number {
        return this._grossShortWinningCount > 0 ? this._grossShortWinningPnl / this._grossShortWinningCount : 0.0;
    }
    get averageLoosingGrossShortPnl(): number {
        return this._grossShortLoosingCount > 0 ? this._grossShortLoosingPnl / this._grossShortLoosingCount : 0.0;
    }
    get averageWinningNetShortPnl(): number {
        return this._netShortWinningCount > 0 ? this._netShortWinningPnl / this._netShortWinningCount : 0.0;
    }
    get averageLoosingNetShortPnl(): number {
        return this._netShortLoosingCount > 0 ? this._netShortLoosingPnl / this._netShortLoosingCount : 0.0;
    }

    // ====================== average win/loss ratio ======================

    get averageGrossWinningLoosingRatio(): number {
        const l = this.averageLoosingGrossPnl;
        return l !== 0 ? this.averageWinningGrossPnl / l : 0.0;
    }
    get averageNetWinningLoosingRatio(): number {
        const l = this.averageLoosingNetPnl;
        return l !== 0 ? this.averageWinningNetPnl / l : 0.0;
    }
    get averageGrossWinningLoosingLongRatio(): number {
        const l = this.averageLoosingGrossLongPnl;
        return l !== 0 ? this.averageWinningGrossLongPnl / l : 0.0;
    }
    get averageNetWinningLoosingLongRatio(): number {
        const l = this.averageLoosingNetLongPnl;
        return l !== 0 ? this.averageWinningNetLongPnl / l : 0.0;
    }
    get averageGrossWinningLoosingShortRatio(): number {
        const l = this.averageLoosingGrossShortPnl;
        return l !== 0 ? this.averageWinningGrossShortPnl / l : 0.0;
    }
    get averageNetWinningLoosingShortRatio(): number {
        const l = this.averageLoosingNetShortPnl;
        return l !== 0 ? this.averageWinningNetShortPnl / l : 0.0;
    }

    // ====================== profit PnL ratio ======================

    get grossProfitPnlRatio(): number {
        return this._grossPnl !== 0 ? this._grossWinningPnl / this._grossPnl : 0.0;
    }
    get netProfitPnlRatio(): number {
        return this._netPnl !== 0 ? this._netWinningPnl / this._netPnl : 0.0;
    }
    get grossProfitPnlLongRatio(): number {
        return this._grossLongPnl !== 0 ? this._grossLongWinningPnl / this._grossLongPnl : 0.0;
    }
    get netProfitPnlLongRatio(): number {
        return this._netLongPnl !== 0 ? this._netLongWinningPnl / this._netLongPnl : 0.0;
    }
    get grossProfitPnlShortRatio(): number {
        return this._grossShortPnl !== 0 ? this._grossShortWinningPnl / this._grossShortPnl : 0.0;
    }
    get netProfitPnlShortRatio(): number {
        return this._netShortPnl !== 0 ? this._netShortWinningPnl / this._netShortPnl : 0.0;
    }

    // ====================== duration (average) ======================

    get averageDurationSeconds(): number {
        return this._totalCount > 0 ? this._durationSec / this._totalCount : 0.0;
    }
    get averageGrossWinningDurationSeconds(): number {
        return this._grossWinningCount > 0 ? this._durationSecGrossWinning / this._grossWinningCount : 0.0;
    }
    get averageGrossLoosingDurationSeconds(): number {
        return this._grossLoosingCount > 0 ? this._durationSecGrossLoosing / this._grossLoosingCount : 0.0;
    }
    get averageNetWinningDurationSeconds(): number {
        return this._netWinningCount > 0 ? this._durationSecNetWinning / this._netWinningCount : 0.0;
    }
    get averageNetLoosingDurationSeconds(): number {
        return this._netLoosingCount > 0 ? this._durationSecNetLoosing / this._netLoosingCount : 0.0;
    }
    get averageLongDurationSeconds(): number {
        return this._longCount > 0 ? this._durationSecLong / this._longCount : 0.0;
    }
    get averageShortDurationSeconds(): number {
        return this._shortCount > 0 ? this._durationSecShort / this._shortCount : 0.0;
    }
    get averageGrossWinningLongDurationSeconds(): number {
        return this._grossLongWinningCount > 0 ? this._durationSecGrossLongWinning / this._grossLongWinningCount : 0.0;
    }
    get averageGrossLoosingLongDurationSeconds(): number {
        return this._grossLongLoosingCount > 0 ? this._durationSecGrossLongLoosing / this._grossLongLoosingCount : 0.0;
    }
    get averageNetWinningLongDurationSeconds(): number {
        return this._netLongWinningCount > 0 ? this._durationSecNetLongWinning / this._netLongWinningCount : 0.0;
    }
    get averageNetLoosingLongDurationSeconds(): number {
        return this._netLongLoosingCount > 0 ? this._durationSecNetLongLoosing / this._netLongLoosingCount : 0.0;
    }
    get averageGrossWinningShortDurationSeconds(): number {
        return this._grossShortWinningCount > 0 ? this._durationSecGrossShortWinning / this._grossShortWinningCount : 0.0;
    }
    get averageGrossLoosingShortDurationSeconds(): number {
        return this._grossShortLoosingCount > 0 ? this._durationSecGrossShortLoosing / this._grossShortLoosingCount : 0.0;
    }
    get averageNetWinningShortDurationSeconds(): number {
        return this._netShortWinningCount > 0 ? this._durationSecNetShortWinning / this._netShortWinningCount : 0.0;
    }
    get averageNetLoosingShortDurationSeconds(): number {
        return this._netShortLoosingCount > 0 ? this._durationSecNetShortLoosing / this._netShortLoosingCount : 0.0;
    }

    // ====================== duration (min/max) ======================

    get minimumDurationSeconds(): number {
        return Math.min(...this.roundtrips.map(r => r.durationSeconds));
    }
    get maximumDurationSeconds(): number {
        return Math.max(...this.roundtrips.map(r => r.durationSeconds));
    }
    get minimumLongDurationSeconds(): number {
        return Math.min(...this.roundtrips.filter(r => r.side === RoundtripSide.LONG).map(r => r.durationSeconds));
    }
    get maximumLongDurationSeconds(): number {
        return Math.max(...this.roundtrips.filter(r => r.side === RoundtripSide.LONG).map(r => r.durationSeconds));
    }
    get minimumShortDurationSeconds(): number {
        return Math.min(...this.roundtrips.filter(r => r.side === RoundtripSide.SHORT).map(r => r.durationSeconds));
    }
    get maximumShortDurationSeconds(): number {
        return Math.max(...this.roundtrips.filter(r => r.side === RoundtripSide.SHORT).map(r => r.durationSeconds));
    }
    get minimumGrossWinningDurationSeconds(): number {
        return Math.min(...this.roundtrips.filter(r => r.grossPnl > 0).map(r => r.durationSeconds));
    }
    get maximumGrossWinningDurationSeconds(): number {
        return Math.max(...this.roundtrips.filter(r => r.grossPnl > 0).map(r => r.durationSeconds));
    }
    get minimumGrossLoosingDurationSeconds(): number {
        return Math.min(...this.roundtrips.filter(r => r.grossPnl < 0).map(r => r.durationSeconds));
    }
    get maximumGrossLoosingDurationSeconds(): number {
        return Math.max(...this.roundtrips.filter(r => r.grossPnl < 0).map(r => r.durationSeconds));
    }
    get minimumNetWinningDurationSeconds(): number {
        return Math.min(...this.roundtrips.filter(r => r.netPnl > 0).map(r => r.durationSeconds));
    }
    get maximumNetWinningDurationSeconds(): number {
        return Math.max(...this.roundtrips.filter(r => r.netPnl > 0).map(r => r.durationSeconds));
    }
    get minimumNetLoosingDurationSeconds(): number {
        return Math.min(...this.roundtrips.filter(r => r.netPnl < 0).map(r => r.durationSeconds));
    }
    get maximumNetLoosingDurationSeconds(): number {
        return Math.max(...this.roundtrips.filter(r => r.netPnl < 0).map(r => r.durationSeconds));
    }
    get minimumGrossWinningLongDurationSeconds(): number {
        return Math.min(...this.roundtrips.filter(r => r.grossPnl > 0 && r.side === RoundtripSide.LONG).map(r => r.durationSeconds));
    }
    get maximumGrossWinningLongDurationSeconds(): number {
        return Math.max(...this.roundtrips.filter(r => r.grossPnl > 0 && r.side === RoundtripSide.LONG).map(r => r.durationSeconds));
    }
    get minimumGrossLoosingLongDurationSeconds(): number {
        return Math.min(...this.roundtrips.filter(r => r.grossPnl < 0 && r.side === RoundtripSide.LONG).map(r => r.durationSeconds));
    }
    get maximumGrossLoosingLongDurationSeconds(): number {
        return Math.max(...this.roundtrips.filter(r => r.grossPnl < 0 && r.side === RoundtripSide.LONG).map(r => r.durationSeconds));
    }
    get minimumNetWinningLongDurationSeconds(): number {
        return Math.min(...this.roundtrips.filter(r => r.netPnl > 0 && r.side === RoundtripSide.LONG).map(r => r.durationSeconds));
    }
    get maximumNetWinningLongDurationSeconds(): number {
        return Math.max(...this.roundtrips.filter(r => r.netPnl > 0 && r.side === RoundtripSide.LONG).map(r => r.durationSeconds));
    }
    get minimumNetLoosingLongDurationSeconds(): number {
        return Math.min(...this.roundtrips.filter(r => r.netPnl < 0 && r.side === RoundtripSide.LONG).map(r => r.durationSeconds));
    }
    get maximumNetLoosingLongDurationSeconds(): number {
        return Math.max(...this.roundtrips.filter(r => r.netPnl < 0 && r.side === RoundtripSide.LONG).map(r => r.durationSeconds));
    }
    get minimumGrossWinningShortDurationSeconds(): number {
        return Math.min(...this.roundtrips.filter(r => r.grossPnl > 0 && r.side === RoundtripSide.SHORT).map(r => r.durationSeconds));
    }
    get maximumGrossWinningShortDurationSeconds(): number {
        return Math.max(...this.roundtrips.filter(r => r.grossPnl > 0 && r.side === RoundtripSide.SHORT).map(r => r.durationSeconds));
    }
    get minimumGrossLoosingShortDurationSeconds(): number {
        return Math.min(...this.roundtrips.filter(r => r.grossPnl < 0 && r.side === RoundtripSide.SHORT).map(r => r.durationSeconds));
    }
    get maximumGrossLoosingShortDurationSeconds(): number {
        return Math.max(...this.roundtrips.filter(r => r.grossPnl < 0 && r.side === RoundtripSide.SHORT).map(r => r.durationSeconds));
    }
    get minimumNetWinningShortDurationSeconds(): number {
        return Math.min(...this.roundtrips.filter(r => r.netPnl > 0 && r.side === RoundtripSide.SHORT).map(r => r.durationSeconds));
    }
    get maximumNetWinningShortDurationSeconds(): number {
        return Math.max(...this.roundtrips.filter(r => r.netPnl > 0 && r.side === RoundtripSide.SHORT).map(r => r.durationSeconds));
    }
    get minimumNetLoosingShortDurationSeconds(): number {
        return Math.min(...this.roundtrips.filter(r => r.netPnl < 0 && r.side === RoundtripSide.SHORT).map(r => r.durationSeconds));
    }
    get maximumNetLoosingShortDurationSeconds(): number {
        return Math.max(...this.roundtrips.filter(r => r.netPnl < 0 && r.side === RoundtripSide.SHORT).map(r => r.durationSeconds));
    }

    // ====================== MAE / MFE / efficiency ======================

    get averageMaximumAdverseExcursion(): number {
        return this._totalCount > 0 ? this._totalMae / this._totalCount : 0.0;
    }
    get averageMaximumFavorableExcursion(): number {
        return this._totalCount > 0 ? this._totalMfe / this._totalCount : 0.0;
    }
    get averageEntryEfficiency(): number {
        return this._totalCount > 0 ? this._totalEffEntry / this._totalCount : 0.0;
    }
    get averageExitEfficiency(): number {
        return this._totalCount > 0 ? this._totalEffExit / this._totalCount : 0.0;
    }
    get averageTotalEfficiency(): number {
        return this._totalCount > 0 ? this._totalEff / this._totalCount : 0.0;
    }

    // ---- filtered MAE/MFE/efficiency averages ----

    private _sumFiltered(extractor: (r: Roundtrip) => number, predicate: (r: Roundtrip) => boolean): number {
        let s = 0;
        for (const r of this.roundtrips) {
            if (predicate(r)) s += extractor(r);
        }
        return s;
    }

    get averageMaximumAdverseExcursionGrossWinning(): number {
        return this._grossWinningCount > 0
            ? this._sumFiltered(r => r.maximumAdverseExcursion, r => r.grossPnl > 0) / this._grossWinningCount : 0.0;
    }
    get averageMaximumAdverseExcursionGrossLoosing(): number {
        return this._grossLoosingCount > 0
            ? this._sumFiltered(r => r.maximumAdverseExcursion, r => r.grossPnl < 0) / this._grossLoosingCount : 0.0;
    }
    get averageMaximumAdverseExcursionNetWinning(): number {
        return this._netWinningCount > 0
            ? this._sumFiltered(r => r.maximumAdverseExcursion, r => r.netPnl > 0) / this._netWinningCount : 0.0;
    }
    get averageMaximumAdverseExcursionNetLoosing(): number {
        return this._netLoosingCount > 0
            ? this._sumFiltered(r => r.maximumAdverseExcursion, r => r.netPnl < 0) / this._netLoosingCount : 0.0;
    }
    get averageMaximumFavorableExcursionGrossWinning(): number {
        return this._grossWinningCount > 0
            ? this._sumFiltered(r => r.maximumFavorableExcursion, r => r.grossPnl > 0) / this._grossWinningCount : 0.0;
    }
    get averageMaximumFavorableExcursionGrossLoosing(): number {
        return this._grossLoosingCount > 0
            ? this._sumFiltered(r => r.maximumFavorableExcursion, r => r.grossPnl < 0) / this._grossLoosingCount : 0.0;
    }
    get averageMaximumFavorableExcursionNetWinning(): number {
        return this._netWinningCount > 0
            ? this._sumFiltered(r => r.maximumFavorableExcursion, r => r.netPnl > 0) / this._netWinningCount : 0.0;
    }
    get averageMaximumFavorableExcursionNetLoosing(): number {
        return this._netLoosingCount > 0
            ? this._sumFiltered(r => r.maximumFavorableExcursion, r => r.netPnl < 0) / this._netLoosingCount : 0.0;
    }
    get averageEntryEfficiencyGrossWinning(): number {
        return this._grossWinningCount > 0
            ? this._sumFiltered(r => r.entryEfficiency, r => r.grossPnl > 0) / this._grossWinningCount : 0.0;
    }
    get averageEntryEfficiencyGrossLoosing(): number {
        return this._grossLoosingCount > 0
            ? this._sumFiltered(r => r.entryEfficiency, r => r.grossPnl < 0) / this._grossLoosingCount : 0.0;
    }
    get averageEntryEfficiencyNetWinning(): number {
        return this._netWinningCount > 0
            ? this._sumFiltered(r => r.entryEfficiency, r => r.netPnl > 0) / this._netWinningCount : 0.0;
    }
    get averageEntryEfficiencyNetLoosing(): number {
        return this._netLoosingCount > 0
            ? this._sumFiltered(r => r.entryEfficiency, r => r.netPnl < 0) / this._netLoosingCount : 0.0;
    }
    get averageExitEfficiencyGrossWinning(): number {
        return this._grossWinningCount > 0
            ? this._sumFiltered(r => r.exitEfficiency, r => r.grossPnl > 0) / this._grossWinningCount : 0.0;
    }
    get averageExitEfficiencyGrossLoosing(): number {
        return this._grossLoosingCount > 0
            ? this._sumFiltered(r => r.exitEfficiency, r => r.grossPnl < 0) / this._grossLoosingCount : 0.0;
    }
    get averageExitEfficiencyNetWinning(): number {
        return this._netWinningCount > 0
            ? this._sumFiltered(r => r.exitEfficiency, r => r.netPnl > 0) / this._netWinningCount : 0.0;
    }
    get averageExitEfficiencyNetLoosing(): number {
        return this._netLoosingCount > 0
            ? this._sumFiltered(r => r.exitEfficiency, r => r.netPnl < 0) / this._netLoosingCount : 0.0;
    }
    get averageTotalEfficiencyGrossWinning(): number {
        return this._grossWinningCount > 0
            ? this._sumFiltered(r => r.totalEfficiency, r => r.grossPnl > 0) / this._grossWinningCount : 0.0;
    }
    get averageTotalEfficiencyGrossLoosing(): number {
        return this._grossLoosingCount > 0
            ? this._sumFiltered(r => r.totalEfficiency, r => r.grossPnl < 0) / this._grossLoosingCount : 0.0;
    }
    get averageTotalEfficiencyNetWinning(): number {
        return this._netWinningCount > 0
            ? this._sumFiltered(r => r.totalEfficiency, r => r.netPnl > 0) / this._netWinningCount : 0.0;
    }
    get averageTotalEfficiencyNetLoosing(): number {
        return this._netLoosingCount > 0
            ? this._sumFiltered(r => r.totalEfficiency, r => r.netPnl < 0) / this._netLoosingCount : 0.0;
    }

    // ====================== consecutive streaks ======================

    get maxConsecutiveGrossWinners(): number {
        return maxConsecutive(this.roundtrips.map(r => r.grossPnl > 0));
    }
    get maxConsecutiveGrossLoosers(): number {
        return maxConsecutive(this.roundtrips.map(r => r.grossPnl < 0));
    }
    get maxConsecutiveNetWinners(): number {
        return maxConsecutive(this.roundtrips.map(r => r.netPnl > 0));
    }
    get maxConsecutiveNetLoosers(): number {
        return maxConsecutive(this.roundtrips.map(r => r.netPnl < 0));
    }
}
