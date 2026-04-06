import { Execution, OrderSide } from './execution';
import { Roundtrip } from './roundtrip';
import { RoundtripPerformance } from './performance';
import { DayCountConvention } from '../daycounting/index';

const BUY = OrderSide.BUY;
const SELL = OrderSide.SELL;

// ---------------------------------------------------------------------------
// Helper to build executions concisely
// ---------------------------------------------------------------------------

function exec(
    side: OrderSide, price: number, comm: number,
    high: number, low: number, dt: Date): Execution {
    return new Execution(side, price, comm, high, low, dt);
}

// ---------------------------------------------------------------------------
// Shared test roundtrips (6 trades, mix of long/short, winning/losing)
// ---------------------------------------------------------------------------

// RT1: Long winner  buy 100 @ $50, sell @ $55
const RT1 = new Roundtrip(
    exec(BUY,  50.0, 0.01,  56.0, 48.0, new Date(Date.UTC(2024, 0, 1,  9, 30))),
    exec(SELL, 55.0, 0.02,  57.0, 49.0, new Date(Date.UTC(2024, 0, 5,  16, 0))),
    100.0);

// RT2: Short winner  sell 200 @ $80, cover @ $72
const RT2 = new Roundtrip(
    exec(SELL, 80.0, 0.03,  85.0, 72.0, new Date(Date.UTC(2024, 1, 1,  10, 0))),
    exec(BUY,  72.0, 0.02,  83.0, 70.0, new Date(Date.UTC(2024, 1, 10, 15, 30))),
    200.0);

// RT3: Long loser  buy 150 @ $60, sell @ $54
const RT3 = new Roundtrip(
    exec(BUY,  60.0, 0.005, 62.0, 53.0, new Date(Date.UTC(2024, 2, 1,  9, 30))),
    exec(SELL, 54.0, 0.005, 61.0, 52.0, new Date(Date.UTC(2024, 2, 3,  16, 0))),
    150.0);

// RT4: Short loser  sell 300 @ $40, cover @ $45
const RT4 = new Roundtrip(
    exec(SELL, 40.0, 0.01,  42.0, 39.0, new Date(Date.UTC(2024, 3, 1,  10, 0))),
    exec(BUY,  45.0, 0.01,  46.0, 38.0, new Date(Date.UTC(2024, 3, 5,  15, 0))),
    300.0);

// RT5: Long winner  buy 50 @ $100, sell @ $110
const RT5 = new Roundtrip(
    exec(BUY,  100.0, 0.02, 112.0, 98.0, new Date(Date.UTC(2024, 4, 1,  9, 0))),
    exec(SELL, 110.0, 0.02, 115.0, 99.0, new Date(Date.UTC(2024, 4, 15, 16, 0))),
    50.0);

// RT6: Short winner  sell 100 @ $90, cover @ $82
const RT6 = new Roundtrip(
    exec(SELL, 90.0, 0.015, 92.0, 84.0, new Date(Date.UTC(2024, 5, 1,  10, 0))),
    exec(BUY,  82.0, 0.015, 93.0, 80.0, new Date(Date.UTC(2024, 5, 20, 15, 0))),
    100.0);

const ALL_RTS = [RT1, RT2, RT3, RT4, RT5, RT6];

// ---------------------------------------------------------------------------
// Initial state
// ---------------------------------------------------------------------------

describe('RoundtripPerformance Init', () => {
    let perf: RoundtripPerformance;

    beforeEach(() => {
        perf = new RoundtripPerformance();
    });

    it('default initial balance', () => {
        expect(perf.initialBalance).toBeCloseTo(100000.0, 13);
    });

    it('default annual risk free rate', () => {
        expect(perf.annualRiskFreeRate).toBeCloseTo(0.0, 13);
    });

    it('total count zero', () => {
        expect(perf.totalCount).toEqual(0);
    });

    it('roi mean null', () => {
        expect(perf.roiMean).toBeNull();
    });

    it('roi std null', () => {
        expect(perf.roiStd).toBeNull();
    });

    it('roi tdd null', () => {
        expect(perf.roiTdd).toBeNull();
    });

    it('sharpe ratio null', () => {
        expect(perf.sharpeRatio).toBeNull();
    });

    it('sortino ratio null', () => {
        expect(perf.sortinoRatio).toBeNull();
    });

    it('calmar ratio null', () => {
        expect(perf.calmarRatio).toBeNull();
    });

    it('empty roundtrips list', () => {
        expect(perf.roundtrips.length).toEqual(0);
    });

    it('total gross pnl zero', () => {
        expect(perf.totalGrossPnl).toBeCloseTo(0.0, 13);
    });

    it('total net pnl zero', () => {
        expect(perf.totalNetPnl).toBeCloseTo(0.0, 13);
    });

    it('max drawdown zero', () => {
        expect(perf.maxDrawdown).toBeCloseTo(0.0, 13);
    });

    it('average net pnl zero', () => {
        expect(perf.averageNetPnl).toBeCloseTo(0.0, 13);
    });
});

// ---------------------------------------------------------------------------
// Reset
// ---------------------------------------------------------------------------

describe('RoundtripPerformance Reset', () => {
    let perf: RoundtripPerformance;

    beforeEach(() => {
        perf = new RoundtripPerformance(100000.0, 0.0, 0.0, DayCountConvention.RAW);
        perf.addRoundtrip(RT1);
        perf.addRoundtrip(RT3);
        perf.reset();
    });

    it('total count zero after reset', () => {
        expect(perf.totalCount).toEqual(0);
    });

    it('total net pnl zero after reset', () => {
        expect(perf.totalNetPnl).toBeCloseTo(0.0, 13);
    });

    it('roi mean null after reset', () => {
        expect(perf.roiMean).toBeNull();
    });

    it('roundtrips list empty after reset', () => {
        expect(perf.roundtrips.length).toEqual(0);
    });

    it('returns on investments empty after reset', () => {
        expect(perf.returnsOnInvestments.length).toEqual(0);
    });

    it('max drawdown zero after reset', () => {
        expect(perf.maxDrawdown).toBeCloseTo(0.0, 13);
    });
});

// ---------------------------------------------------------------------------
// Single long winner
// ---------------------------------------------------------------------------

describe('RoundtripPerformance Single Long Winner', () => {
    let perf: RoundtripPerformance;

    beforeEach(() => {
        perf = new RoundtripPerformance(100000.0, 0.0, 0.0, DayCountConvention.RAW);
        perf.addRoundtrip(RT1);
    });

    // --- counts ---
    it('total count', () => { expect(perf.totalCount).toEqual(1); });
    it('long count', () => { expect(perf.longCount).toEqual(1); });
    it('short count', () => { expect(perf.shortCount).toEqual(0); });
    it('gross winning count', () => { expect(perf.grossWinningCount).toEqual(1); });
    it('gross loosing count', () => { expect(perf.grossLoosingCount).toEqual(0); });
    it('net winning count', () => { expect(perf.netWinningCount).toEqual(1); });
    it('net loosing count', () => { expect(perf.netLoosingCount).toEqual(0); });

    // --- PnL ---
    it('total gross pnl', () => {
        expect(perf.totalGrossPnl).toBeCloseTo(500.0, 13);
    });

    it('total net pnl', () => {
        expect(perf.totalNetPnl).toBeCloseTo(497.0, 13);
    });

    it('total commission', () => {
        expect(perf.totalCommission).toBeCloseTo(3.0, 13);
    });

    // --- ROI ---
    it('roi mean', () => {
        expect(perf.roiMean).toBeCloseTo(0.0994, 13);
    });

    it('roi std zero', () => {
        expect(perf.roiStd).toBeCloseTo(0.0, 13);
    });

    it('roi tdd null', () => {
        expect(perf.roiTdd).toBeNull();
    });

    // --- risk-adjusted ratios ---
    it('sharpe ratio null', () => {
        expect(perf.sharpeRatio).toBeNull();
    });

    it('sortino ratio null', () => {
        expect(perf.sortinoRatio).toBeNull();
    });

    it('calmar ratio null', () => {
        expect(perf.calmarRatio).toBeNull();
    });

    // --- drawdown ---
    it('max drawdown zero', () => {
        expect(perf.maxDrawdown).toBeCloseTo(0.0, 13);
    });

    // --- rate of return ---
    it('rate of return', () => {
        expect(perf.rateOfReturn).toBeCloseTo(0.00497, 13);
    });

    // --- ratios ---
    it('gross winning ratio', () => {
        expect(perf.grossWinningRatio).toBeCloseTo(1.0, 13);
    });

    it('net winning ratio', () => {
        expect(perf.netWinningRatio).toBeCloseTo(1.0, 13);
    });

    // --- profit ratio ---
    it('gross profit ratio null', () => {
        expect(perf.grossProfitRatio).toBeNull();
    });

    it('net profit ratio null', () => {
        expect(perf.netProfitRatio).toBeNull();
    });

    // --- MAE/MFE/efficiency ---
    it('average mae', () => {
        expect(perf.averageMaximumAdverseExcursion).toBeCloseTo(
            RT1.maximumAdverseExcursion, 13);
    });

    it('average mfe', () => {
        expect(perf.averageMaximumFavorableExcursion).toBeCloseTo(
            RT1.maximumFavorableExcursion, 13);
    });

    it('average entry efficiency', () => {
        expect(perf.averageEntryEfficiency).toBeCloseTo(
            RT1.entryEfficiency, 13);
    });

    it('average exit efficiency', () => {
        expect(perf.averageExitEfficiency).toBeCloseTo(
            RT1.exitEfficiency, 13);
    });

    it('average total efficiency', () => {
        expect(perf.averageTotalEfficiency).toBeCloseTo(
            RT1.totalEfficiency, 13);
    });

    // --- duration ---
    it('average duration seconds', () => {
        expect(perf.averageDurationSeconds).toBeCloseTo(369000.0, 13);
    });

    // --- consecutive ---
    it('max consecutive gross winners', () => {
        expect(perf.maxConsecutiveGrossWinners).toEqual(1);
    });

    it('max consecutive gross loosers', () => {
        expect(perf.maxConsecutiveGrossLoosers).toEqual(0);
    });
});

// ---------------------------------------------------------------------------
// Single long loser
// ---------------------------------------------------------------------------

describe('RoundtripPerformance Single Loser', () => {
    let perf: RoundtripPerformance;

    beforeEach(() => {
        perf = new RoundtripPerformance(100000.0, 0.0, 0.0, DayCountConvention.RAW);
        perf.addRoundtrip(RT3);
    });

    it('total net pnl negative', () => {
        expect(perf.totalNetPnl).toBeCloseTo(-901.5, 13);
    });

    it('max drawdown', () => {
        expect(perf.maxDrawdown).toBeCloseTo(901.5, 13);
    });

    it('max drawdown percent', () => {
        expect(perf.maxDrawdownPercent).toBeCloseTo(0.009015, 13);
    });

    it('calmar ratio', () => {
        expect(perf.calmarRatio).toBeCloseTo(-11.11111111111111, 10);
    });

    it('roi mean negative', () => {
        expect(perf.roiMean).toBeCloseTo(-0.10016666666666667, 13);
    });

    it('roi tdd', () => {
        expect(perf.roiTdd).toBeCloseTo(0.10016666666666667, 13);
    });

    it('sortino ratio', () => {
        expect(perf.sortinoRatio).toBeCloseTo(-1.0, 13);
    });

    it('gross loosing count', () => {
        expect(perf.grossLoosingCount).toEqual(1);
    });

    it('net loosing count', () => {
        expect(perf.netLoosingCount).toEqual(1);
    });
});

// ---------------------------------------------------------------------------
// Multiple mixed roundtrips (all 6)
// ---------------------------------------------------------------------------

describe('RoundtripPerformance Multiple Mixed', () => {
    let perf: RoundtripPerformance;

    beforeEach(() => {
        perf = new RoundtripPerformance(100000.0, 0.0, 0.0, DayCountConvention.RAW);
        for (const rt of ALL_RTS) {
            perf.addRoundtrip(rt);
        }
    });

    // ====================== counts ======================

    it('total count', () => { expect(perf.totalCount).toEqual(6); });
    it('long count', () => { expect(perf.longCount).toEqual(3); });
    it('short count', () => { expect(perf.shortCount).toEqual(3); });
    it('gross winning count', () => { expect(perf.grossWinningCount).toEqual(4); });
    it('gross loosing count', () => { expect(perf.grossLoosingCount).toEqual(2); });
    it('net winning count', () => { expect(perf.netWinningCount).toEqual(4); });
    it('net loosing count', () => { expect(perf.netLoosingCount).toEqual(2); });
    it('gross long winning count', () => { expect(perf.grossLongWinningCount).toEqual(2); });
    it('gross long loosing count', () => { expect(perf.grossLongLoosingCount).toEqual(1); });
    it('net long winning count', () => { expect(perf.netLongWinningCount).toEqual(2); });
    it('net long loosing count', () => { expect(perf.netLongLoosingCount).toEqual(1); });
    it('gross short winning count', () => { expect(perf.grossShortWinningCount).toEqual(2); });
    it('gross short loosing count', () => { expect(perf.grossShortLoosingCount).toEqual(1); });
    it('net short winning count', () => { expect(perf.netShortWinningCount).toEqual(2); });
    it('net short loosing count', () => { expect(perf.netShortLoosingCount).toEqual(1); });

    // ====================== PnL totals ======================

    it('total gross pnl', () => {
        expect(perf.totalGrossPnl).toBeCloseTo(1000.0, 13);
    });

    it('total net pnl', () => {
        expect(perf.totalNetPnl).toBeCloseTo(974.5, 13);
    });

    it('winning gross pnl', () => {
        expect(perf.winningGrossPnl).toBeCloseTo(3400.0, 13);
    });

    it('loosing gross pnl', () => {
        expect(perf.loosingGrossPnl).toBeCloseTo(-2400.0, 13);
    });

    it('winning net pnl', () => {
        expect(perf.winningNetPnl).toBeCloseTo(3382.0, 13);
    });

    it('loosing net pnl', () => {
        expect(perf.loosingNetPnl).toBeCloseTo(-2407.5, 13);
    });

    it('winning gross long pnl', () => {
        expect(perf.winningGrossLongPnl).toBeCloseTo(1000.0, 13);
    });

    it('loosing gross long pnl', () => {
        expect(perf.loosingGrossLongPnl).toBeCloseTo(-900.0, 13);
    });

    it('winning gross short pnl', () => {
        expect(perf.winningGrossShortPnl).toBeCloseTo(2400.0, 13);
    });

    it('loosing gross short pnl', () => {
        expect(perf.loosingGrossShortPnl).toBeCloseTo(-1500.0, 13);
    });

    // ====================== commission ======================

    it('total commission', () => {
        expect(perf.totalCommission).toBeCloseTo(25.5, 13);
    });

    it('gross winning commission', () => {
        expect(perf.grossWinningCommission).toBeCloseTo(18.0, 13);
    });

    it('gross loosing commission', () => {
        expect(perf.grossLoosingCommission).toBeCloseTo(7.5, 13);
    });

    it('net winning commission', () => {
        expect(perf.netWinningCommission).toBeCloseTo(18.0, 13);
    });

    it('net loosing commission', () => {
        expect(perf.netLoosingCommission).toBeCloseTo(7.5, 13);
    });

    // ====================== average PnL ======================

    it('average gross pnl', () => {
        expect(perf.averageGrossPnl).toBeCloseTo(1000.0 / 6.0, 13);
    });

    it('average net pnl', () => {
        expect(perf.averageNetPnl).toBeCloseTo(974.5 / 6.0, 13);
    });

    it('average winning gross pnl', () => {
        expect(perf.averageWinningGrossPnl).toBeCloseTo(3400.0 / 4.0, 13);
    });

    it('average loosing gross pnl', () => {
        expect(perf.averageLoosingGrossPnl).toBeCloseTo(-2400.0 / 2.0, 13);
    });

    it('average winning net pnl', () => {
        expect(perf.averageWinningNetPnl).toBeCloseTo(3382.0 / 4.0, 13);
    });

    it('average loosing net pnl', () => {
        expect(perf.averageLoosingNetPnl).toBeCloseTo(-2407.5 / 2.0, 13);
    });

    it('average gross long pnl', () => {
        // (500 - 900 + 500) / 3 = 100/3
        expect(perf.averageGrossLongPnl).toBeCloseTo(100.0 / 3.0, 13);
    });

    it('average gross short pnl', () => {
        // (1600 - 1500 + 800) / 3 = 300
        expect(perf.averageGrossShortPnl).toBeCloseTo(300.0, 13);
    });

    // ====================== win/loss ratios ======================

    it('gross winning ratio', () => {
        expect(perf.grossWinningRatio).toBeCloseTo(4.0 / 6.0, 13);
    });

    it('gross loosing ratio', () => {
        expect(perf.grossLoosingRatio).toBeCloseTo(2.0 / 6.0, 13);
    });

    it('net winning ratio', () => {
        expect(perf.netWinningRatio).toBeCloseTo(4.0 / 6.0, 13);
    });

    it('net loosing ratio', () => {
        expect(perf.netLoosingRatio).toBeCloseTo(2.0 / 6.0, 13);
    });

    it('gross long winning ratio', () => {
        expect(perf.grossLongWinningRatio).toBeCloseTo(2.0 / 3.0, 13);
    });

    it('gross short winning ratio', () => {
        expect(perf.grossShortWinningRatio).toBeCloseTo(2.0 / 3.0, 13);
    });

    // ====================== profit ratios ======================

    it('gross profit ratio', () => {
        expect(perf.grossProfitRatio).toBeCloseTo(1.4166666666666667, 13);
    });

    it('net profit ratio', () => {
        expect(perf.netProfitRatio).toBeCloseTo(1.4047767393561785, 13);
    });

    it('gross profit long ratio', () => {
        expect(perf.grossProfitLongRatio).toBeCloseTo(1.1111111111111112, 13);
    });

    it('gross profit short ratio', () => {
        expect(perf.grossProfitShortRatio).toBeCloseTo(1.6, 13);
    });

    // ====================== profit PnL ratio ======================

    it('gross profit pnl ratio', () => {
        expect(perf.grossProfitPnlRatio).toBeCloseTo(3.4, 13);
    });

    it('net profit pnl ratio', () => {
        expect(perf.netProfitPnlRatio).toBeCloseTo(3382.0 / 974.5, 13);
    });

    // ====================== average win/loss ratio ======================

    it('average gross winning loosing ratio', () => {
        expect(perf.averageGrossWinningLoosingRatio).toBeCloseTo(850.0 / -1200.0, 13);
    });

    it('average net winning loosing ratio', () => {
        expect(perf.averageNetWinningLoosingRatio).toBeCloseTo(845.5 / -1203.75, 13);
    });

    // ====================== ROI statistics ======================

    it('roi mean', () => {
        expect(perf.roiMean).toBeCloseTo(0.026877314814814812, 13);
    });

    it('roi std', () => {
        expect(perf.roiStd).toBeCloseTo(0.0991356544050762, 13);
    });

    it('roi tdd', () => {
        expect(perf.roiTdd).toBeCloseTo(0.11354208715518468, 13);
    });

    it('roiann mean', () => {
        expect(perf.roiannMean).toBeCloseTo(-1.7233887909446202, 12);
    });

    it('roiann std', () => {
        expect(perf.roiannStd).toBeCloseTo(8.73138705463156, 12);
    });

    it('roiann tdd', () => {
        expect(perf.roiannTdd).toBeCloseTo(13.751365296707874, 12);
    });

    // ====================== risk-adjusted ratios ======================

    it('sharpe ratio', () => {
        expect(perf.sharpeRatio).toBeCloseTo(0.27111653194916085, 13);
    });

    it('sharpe ratio annual', () => {
        expect(perf.sharpeRatioAnnual).toBeCloseTo(-0.1973785814512082, 12);
    });

    it('sortino ratio', () => {
        expect(perf.sortinoRatio).toBeCloseTo(0.23671675841293985, 13);
    });

    it('sortino ratio annual', () => {
        expect(perf.sortinoRatioAnnual).toBeCloseTo(-0.1253249225629404, 12);
    });

    it('calmar ratio', () => {
        expect(perf.calmarRatio).toBeCloseTo(1.139698624091381, 12);
    });

    it('calmar ratio annual', () => {
        expect(perf.calmarRatioAnnual).toBeCloseTo(-73.07812731097131, 10);
    });

    // ====================== rate of return ======================

    it('rate of return', () => {
        expect(perf.rateOfReturn).toBeCloseTo(0.009745, 13);
    });

    it('rate of return annual', () => {
        expect(perf.rateOfReturnAnnual).toBeCloseTo(0.020786693247353695, 12);
    });

    it('recovery factor', () => {
        expect(perf.recoveryFactor).toBeCloseTo(0.8814335009522727, 12);
    });

    // ====================== drawdown ======================

    it('max net pnl', () => {
        expect(perf.maxNetPnl).toBeCloseTo(2087.0, 13);
    });

    it('max drawdown', () => {
        expect(perf.maxDrawdown).toBeCloseTo(2407.5, 13);
    });

    it('max drawdown percent', () => {
        expect(perf.maxDrawdownPercent).toBeCloseTo(
            2407.5 / (100000.0 + 2087.0), 13);
    });

    // ====================== duration ======================

    it('average duration seconds', () => {
        expect(perf.averageDurationSeconds).toBeCloseTo(770100.0, 13);
    });

    it('average long duration seconds', () => {
        expect(perf.averageLongDurationSeconds).toBeCloseTo(600000.0, 13);
    });

    it('average short duration seconds', () => {
        expect(perf.averageShortDurationSeconds).toBeCloseTo(940200.0, 13);
    });

    it('average gross winning duration seconds', () => {
        expect(perf.averageGrossWinningDurationSeconds).toBeCloseTo(1015200.0, 13);
    });

    it('average gross loosing duration seconds', () => {
        expect(perf.averageGrossLoosingDurationSeconds).toBeCloseTo(279900.0, 13);
    });

    it('minimum duration seconds', () => {
        expect(perf.minimumDurationSeconds).toBeCloseTo(196200.0, 13);
    });

    it('maximum duration seconds', () => {
        expect(perf.maximumDurationSeconds).toBeCloseTo(1659600.0, 13);
    });

    it('minimum long duration seconds', () => {
        expect(perf.minimumLongDurationSeconds).toBeCloseTo(196200.0, 13);
    });

    it('maximum long duration seconds', () => {
        expect(perf.maximumLongDurationSeconds).toBeCloseTo(1234800.0, 13);
    });

    it('minimum short duration seconds', () => {
        expect(perf.minimumShortDurationSeconds).toBeCloseTo(363600.0, 13);
    });

    it('maximum short duration seconds', () => {
        expect(perf.maximumShortDurationSeconds).toBeCloseTo(1659600.0, 13);
    });

    // ====================== MAE / MFE / efficiency ======================

    it('average mae', () => {
        const expected = ALL_RTS.reduce((s, r) => s + r.maximumAdverseExcursion, 0) / 6.0;
        expect(perf.averageMaximumAdverseExcursion).toBeCloseTo(expected, 13);
    });

    it('average mfe', () => {
        const expected = ALL_RTS.reduce((s, r) => s + r.maximumFavorableExcursion, 0) / 6.0;
        expect(perf.averageMaximumFavorableExcursion).toBeCloseTo(expected, 13);
    });

    it('average entry efficiency', () => {
        const expected = ALL_RTS.reduce((s, r) => s + r.entryEfficiency, 0) / 6.0;
        expect(perf.averageEntryEfficiency).toBeCloseTo(expected, 13);
    });

    it('average exit efficiency', () => {
        const expected = ALL_RTS.reduce((s, r) => s + r.exitEfficiency, 0) / 6.0;
        expect(perf.averageExitEfficiency).toBeCloseTo(expected, 13);
    });

    it('average total efficiency', () => {
        const expected = ALL_RTS.reduce((s, r) => s + r.totalEfficiency, 0) / 6.0;
        expect(perf.averageTotalEfficiency).toBeCloseTo(expected, 13);
    });

    // ====================== consecutive ======================

    it('max consecutive gross winners', () => {
        expect(perf.maxConsecutiveGrossWinners).toEqual(2);
    });

    it('max consecutive gross loosers', () => {
        expect(perf.maxConsecutiveGrossLoosers).toEqual(2);
    });

    it('max consecutive net winners', () => {
        expect(perf.maxConsecutiveNetWinners).toEqual(2);
    });

    it('max consecutive net loosers', () => {
        expect(perf.maxConsecutiveNetLoosers).toEqual(2);
    });

    // ====================== time tracking ======================

    it('first time', () => {
        expect(perf.firstTime!.getTime()).toEqual(new Date(Date.UTC(2024, 0, 1, 9, 30)).getTime());
    });

    it('last time', () => {
        expect(perf.lastTime!.getTime()).toEqual(new Date(Date.UTC(2024, 5, 20, 15, 0)).getTime());
    });
});

// ---------------------------------------------------------------------------
// Edge cases
// ---------------------------------------------------------------------------

describe('RoundtripPerformance Edge Cases', () => {

    it('zero initial balance rate of return null', () => {
        const perf = new RoundtripPerformance(0.0, 0.0, 0.0, DayCountConvention.RAW);
        expect(perf.rateOfReturn).toBeNull();
    });

    it('no roundtrips average gross pnl zero', () => {
        const perf = new RoundtripPerformance();
        expect(perf.averageGrossPnl).toBeCloseTo(0.0, 13);
    });

    it('no roundtrips average net pnl zero', () => {
        const perf = new RoundtripPerformance();
        expect(perf.averageNetPnl).toBeCloseTo(0.0, 13);
    });

    it('no roundtrips gross winning ratio zero', () => {
        const perf = new RoundtripPerformance();
        expect(perf.grossWinningRatio).toBeCloseTo(0.0, 13);
    });

    it('no roundtrips average duration zero', () => {
        const perf = new RoundtripPerformance();
        expect(perf.averageDurationSeconds).toBeCloseTo(0.0, 13);
    });

    it('sharpe null single point', () => {
        const perf = new RoundtripPerformance(100000.0, 0.0, 0.0, DayCountConvention.RAW);
        perf.addRoundtrip(RT1);
        expect(perf.sharpeRatio).toBeNull();
    });

    it('rate of return annual null when zero duration', () => {
        const perf = new RoundtripPerformance(100000.0, 0.0, 0.0, DayCountConvention.RAW);
        expect(perf.rateOfReturnAnnual).toBeNull();
    });

    it('recovery factor null no drawdown', () => {
        const perf = new RoundtripPerformance(100000.0, 0.0, 0.0, DayCountConvention.RAW);
        perf.addRoundtrip(RT1);
        expect(perf.recoveryFactor).toBeNull();
    });
});

// ---------------------------------------------------------------------------
// Incremental update
// ---------------------------------------------------------------------------

describe('RoundtripPerformance Incremental', () => {

    it('roi list length', () => {
        const perf = new RoundtripPerformance(100000.0, 0.0, 0.0, DayCountConvention.RAW);
        for (let i = 0; i < ALL_RTS.length; i++) {
            perf.addRoundtrip(ALL_RTS[i]);
            expect(perf.returnsOnInvestments.length).toEqual(i + 1);
        }
    });

    it('roi values', () => {
        const expectedRois = [
            0.0994,                 // 497 / (100*50)
            0.099375,               // 1590 / (200*80)
            -0.10016666666666667,   // -901.5 / (150*60)
            -0.1255,                // -1506 / (300*40)
            0.0996,                 // 498 / (50*100)
            0.08855555555555556,    // 797 / (100*90)
        ];
        const perf = new RoundtripPerformance(100000.0, 0.0, 0.0, DayCountConvention.RAW);
        for (const rt of ALL_RTS) {
            perf.addRoundtrip(rt);
        }
        for (let i = 0; i < ALL_RTS.length; i++) {
            expect(perf.returnsOnInvestments[i]).toBeCloseTo(expectedRois[i], 13);
        }
    });

    it('sortino downside count', () => {
        const perf = new RoundtripPerformance(100000.0, 0.0, 0.0, DayCountConvention.RAW);
        for (const rt of ALL_RTS) {
            perf.addRoundtrip(rt);
        }
        expect(perf.sortinoDownsideReturns.length).toEqual(2);
    });
});
