import { Execution, OrderSide } from './execution';
import { RoundtripSide } from './side';
import { Roundtrip } from './roundtrip';

// ---------------------------------------------------------------------------
// Concrete test data
// ---------------------------------------------------------------------------

// Long trade: buy 100 shares at $50, sell at $55
const LONG_ENTRY = new Execution(
    OrderSide.BUY, 50.0, 0.01, 56.0, 48.0,
    new Date(Date.UTC(2024, 0, 1, 9, 30, 0)));  // Jan 1 2024 09:30
const LONG_EXIT = new Execution(
    OrderSide.SELL, 55.0, 0.02, 57.0, 49.0,
    new Date(Date.UTC(2024, 0, 5, 16, 0, 0)));  // Jan 5 2024 16:00
const LONG_QTY = 100.0;

// Short trade: sell 200 shares at $80, cover at $72
const SHORT_ENTRY = new Execution(
    OrderSide.SELL, 80.0, 0.03, 85.0, 72.0,
    new Date(Date.UTC(2024, 1, 1, 10, 0, 0)));  // Feb 1 2024 10:00
const SHORT_EXIT = new Execution(
    OrderSide.BUY, 72.0, 0.02, 83.0, 70.0,
    new Date(Date.UTC(2024, 1, 10, 15, 30, 0))); // Feb 10 2024 15:30
const SHORT_QTY = 200.0;

// ---------------------------------------------------------------------------
// Tests for a LONG round-trip
// ---------------------------------------------------------------------------

describe('Roundtrip Long', () => {
    let rt: Roundtrip;

    beforeEach(() => {
        rt = new Roundtrip(LONG_ENTRY, LONG_EXIT, LONG_QTY);
    });

    it('side', () => {
        expect(rt.side).toEqual(RoundtripSide.LONG);
    });

    it('quantity', () => {
        expect(rt.quantity).toBeCloseTo(100.0, 13);
    });

    it('entryTime', () => {
        expect(rt.entryTime.getTime()).toEqual(new Date(Date.UTC(2024, 0, 1, 9, 30, 0)).getTime());
    });

    it('exitTime', () => {
        expect(rt.exitTime.getTime()).toEqual(new Date(Date.UTC(2024, 0, 5, 16, 0, 0)).getTime());
    });

    it('entryPrice', () => {
        expect(rt.entryPrice).toBeCloseTo(50.0, 13);
    });

    it('exitPrice', () => {
        expect(rt.exitPrice).toBeCloseTo(55.0, 13);
    });

    it('duration', () => {
        // 4 days 6 hours 30 minutes = 369000 seconds
        expect(rt.durationSeconds).toBeCloseTo(369000.0, 13);
    });

    it('highestPrice', () => {
        expect(rt.highestPrice).toBeCloseTo(57.0, 13);
    });

    it('lowestPrice', () => {
        expect(rt.lowestPrice).toBeCloseTo(48.0, 13);
    });

    it('grossPnl', () => {
        // Long: qty * (exit - entry) = 100 * (55 - 50) = 500
        expect(rt.grossPnl).toBeCloseTo(500.0, 13);
    });

    it('commission', () => {
        // (0.01 + 0.02) * 100 = 3.0
        expect(rt.commission).toBeCloseTo(3.0, 13);
    });

    it('netPnl', () => {
        // 500 - 3 = 497
        expect(rt.netPnl).toBeCloseTo(497.0, 13);
    });

    it('maximumAdversePrice', () => {
        // Long: lowest = 48
        expect(rt.maximumAdversePrice).toBeCloseTo(48.0, 13);
    });

    it('maximumFavorablePrice', () => {
        // Long: highest = 57
        expect(rt.maximumFavorablePrice).toBeCloseTo(57.0, 13);
    });

    it('maximumAdverseExcursion', () => {
        // Long MAE: 100 * (1 - 48/50) = 4.0
        expect(rt.maximumAdverseExcursion).toBeCloseTo(4.0, 13);
    });

    it('maximumFavorableExcursion', () => {
        // Long MFE: 100 * (57/55 - 1)
        const expected = 100.0 * (57.0 / 55.0 - 1.0);
        expect(rt.maximumFavorableExcursion).toBeCloseTo(expected, 13);
    });

    it('entryEfficiency', () => {
        // Long: 100 * (highest - entry) / delta = 100 * (57 - 50) / 9
        const expected = 100.0 * (57.0 - 50.0) / 9.0;
        expect(rt.entryEfficiency).toBeCloseTo(expected, 13);
    });

    it('exitEfficiency', () => {
        // Long: 100 * (exit - lowest) / delta = 100 * (55 - 48) / 9
        const expected = 100.0 * (55.0 - 48.0) / 9.0;
        expect(rt.exitEfficiency).toBeCloseTo(expected, 13);
    });

    it('totalEfficiency', () => {
        // Long: 100 * (exit - entry) / delta = 100 * (55 - 50) / 9
        const expected = 100.0 * (55.0 - 50.0) / 9.0;
        expect(rt.totalEfficiency).toBeCloseTo(expected, 13);
    });
});

// ---------------------------------------------------------------------------
// Tests for a SHORT round-trip
// ---------------------------------------------------------------------------

describe('Roundtrip Short', () => {
    let rt: Roundtrip;

    beforeEach(() => {
        rt = new Roundtrip(SHORT_ENTRY, SHORT_EXIT, SHORT_QTY);
    });

    it('side', () => {
        expect(rt.side).toEqual(RoundtripSide.SHORT);
    });

    it('quantity', () => {
        expect(rt.quantity).toBeCloseTo(200.0, 13);
    });

    it('entryTime', () => {
        expect(rt.entryTime.getTime()).toEqual(new Date(Date.UTC(2024, 1, 1, 10, 0, 0)).getTime());
    });

    it('exitTime', () => {
        expect(rt.exitTime.getTime()).toEqual(new Date(Date.UTC(2024, 1, 10, 15, 30, 0)).getTime());
    });

    it('entryPrice', () => {
        expect(rt.entryPrice).toBeCloseTo(80.0, 13);
    });

    it('exitPrice', () => {
        expect(rt.exitPrice).toBeCloseTo(72.0, 13);
    });

    it('duration', () => {
        // Feb 1 10:00 to Feb 10 15:30 = 9 days 5 hours 30 min = 798600 sec
        const expected = (9 * 24 * 3600) + (5 * 3600) + (30 * 60);
        expect(rt.durationSeconds).toBeCloseTo(expected, 13);
    });

    it('highestPrice', () => {
        expect(rt.highestPrice).toBeCloseTo(85.0, 13);
    });

    it('lowestPrice', () => {
        expect(rt.lowestPrice).toBeCloseTo(70.0, 13);
    });

    it('grossPnl', () => {
        // Short: qty * (entry - exit) = 200 * (80 - 72) = 1600
        expect(rt.grossPnl).toBeCloseTo(1600.0, 13);
    });

    it('commission', () => {
        // (0.03 + 0.02) * 200 = 10.0
        expect(rt.commission).toBeCloseTo(10.0, 13);
    });

    it('netPnl', () => {
        // 1600 - 10 = 1590
        expect(rt.netPnl).toBeCloseTo(1590.0, 13);
    });

    it('maximumAdversePrice', () => {
        // Short: highest = 85
        expect(rt.maximumAdversePrice).toBeCloseTo(85.0, 13);
    });

    it('maximumFavorablePrice', () => {
        // Short: lowest = 70
        expect(rt.maximumFavorablePrice).toBeCloseTo(70.0, 13);
    });

    it('maximumAdverseExcursion', () => {
        // Short MAE: 100 * (85/80 - 1) = 6.25
        expect(rt.maximumAdverseExcursion).toBeCloseTo(6.25, 13);
    });

    it('maximumFavorableExcursion', () => {
        // Short MFE: 100 * (1 - 70/72)
        const expected = 100.0 * (1.0 - 70.0 / 72.0);
        expect(rt.maximumFavorableExcursion).toBeCloseTo(expected, 13);
    });

    it('entryEfficiency', () => {
        // Short: 100 * (entry - lowest) / delta = 100 * (80 - 70) / 15
        const expected = 100.0 * (80.0 - 70.0) / 15.0;
        expect(rt.entryEfficiency).toBeCloseTo(expected, 13);
    });

    it('exitEfficiency', () => {
        // Short: 100 * (highest - exit) / delta = 100 * (85 - 72) / 15
        const expected = 100.0 * (85.0 - 72.0) / 15.0;
        expect(rt.exitEfficiency).toBeCloseTo(expected, 13);
    });

    it('totalEfficiency', () => {
        // Short: 100 * (entry - exit) / delta = 100 * (80 - 72) / 15
        const expected = 100.0 * (80.0 - 72.0) / 15.0;
        expect(rt.totalEfficiency).toBeCloseTo(expected, 13);
    });
});

// ---------------------------------------------------------------------------
// Tests for zero-delta edge case (highest == lowest)
// ---------------------------------------------------------------------------

describe('Roundtrip Zero Delta', () => {
    let rt: Roundtrip;

    beforeEach(() => {
        const entry = new Execution(
            OrderSide.BUY, 100.0, 0.0, 100.0, 100.0,
            new Date(Date.UTC(2024, 2, 1, 9, 0, 0)));
        const exit = new Execution(
            OrderSide.SELL, 100.0, 0.0, 100.0, 100.0,
            new Date(Date.UTC(2024, 2, 1, 10, 0, 0)));
        rt = new Roundtrip(entry, exit, 50.0);
    });

    it('entryEfficiency zero', () => {
        expect(rt.entryEfficiency).toBeCloseTo(0.0, 13);
    });

    it('exitEfficiency zero', () => {
        expect(rt.exitEfficiency).toBeCloseTo(0.0, 13);
    });

    it('totalEfficiency zero', () => {
        expect(rt.totalEfficiency).toBeCloseTo(0.0, 13);
    });

    it('grossPnl zero', () => {
        expect(rt.grossPnl).toBeCloseTo(0.0, 13);
    });

    it('netPnl zero', () => {
        expect(rt.netPnl).toBeCloseTo(0.0, 13);
    });
});

// ---------------------------------------------------------------------------
// Immutability tests (readonly fields)
// ---------------------------------------------------------------------------

describe('Roundtrip Immutability', () => {
    let rt: Roundtrip;

    beforeEach(() => {
        rt = new Roundtrip(LONG_ENTRY, LONG_EXIT, LONG_QTY);
    });

    it('has readonly side', () => {
        // TypeScript readonly fields — verify they exist and are correct type
        expect(rt.side).toEqual(RoundtripSide.LONG);
    });

    it('has readonly grossPnl', () => {
        expect(rt.grossPnl).toBeCloseTo(500.0, 13);
    });
});

// ---------------------------------------------------------------------------
// Long losing trade
// ---------------------------------------------------------------------------

describe('Roundtrip Long Loser', () => {
    let rt: Roundtrip;

    beforeEach(() => {
        const entry = new Execution(
            OrderSide.BUY, 60.0, 0.005, 62.0, 53.0,
            new Date(Date.UTC(2024, 3, 1, 9, 30, 0)));
        const exit = new Execution(
            OrderSide.SELL, 54.0, 0.005, 61.0, 52.0,
            new Date(Date.UTC(2024, 3, 3, 16, 0, 0)));
        rt = new Roundtrip(entry, exit, 150.0);
    });

    it('side', () => {
        expect(rt.side).toEqual(RoundtripSide.LONG);
    });

    it('grossPnl negative', () => {
        expect(rt.grossPnl).toBeCloseTo(-900.0, 13);
    });

    it('commission', () => {
        expect(rt.commission).toBeCloseTo(1.5, 13);
    });

    it('netPnl negative', () => {
        expect(rt.netPnl).toBeCloseTo(-901.5, 13);
    });

    it('highestPrice', () => {
        expect(rt.highestPrice).toBeCloseTo(62.0, 13);
    });

    it('lowestPrice', () => {
        expect(rt.lowestPrice).toBeCloseTo(52.0, 13);
    });

    it('mae', () => {
        const expected = 100.0 * (1.0 - 52.0 / 60.0);
        expect(rt.maximumAdverseExcursion).toBeCloseTo(expected, 13);
    });

    it('mfe', () => {
        const expected = 100.0 * (62.0 / 54.0 - 1.0);
        expect(rt.maximumFavorableExcursion).toBeCloseTo(expected, 13);
    });
});

// ---------------------------------------------------------------------------
// Short losing trade
// ---------------------------------------------------------------------------

describe('Roundtrip Short Loser', () => {
    let rt: Roundtrip;

    beforeEach(() => {
        const entry = new Execution(
            OrderSide.SELL, 40.0, 0.01, 42.0, 39.0,
            new Date(Date.UTC(2024, 4, 1, 10, 0, 0)));
        const exit = new Execution(
            OrderSide.BUY, 45.0, 0.01, 46.0, 38.0,
            new Date(Date.UTC(2024, 4, 5, 15, 0, 0)));
        rt = new Roundtrip(entry, exit, 300.0);
    });

    it('side', () => {
        expect(rt.side).toEqual(RoundtripSide.SHORT);
    });

    it('grossPnl negative', () => {
        expect(rt.grossPnl).toBeCloseTo(-1500.0, 13);
    });

    it('commission', () => {
        expect(rt.commission).toBeCloseTo(6.0, 13);
    });

    it('netPnl negative', () => {
        expect(rt.netPnl).toBeCloseTo(-1506.0, 13);
    });

    it('maximumAdversePrice', () => {
        expect(rt.maximumAdversePrice).toBeCloseTo(46.0, 13);
    });

    it('maximumFavorablePrice', () => {
        expect(rt.maximumFavorablePrice).toBeCloseTo(38.0, 13);
    });

    it('mae', () => {
        expect(rt.maximumAdverseExcursion).toBeCloseTo(15.0, 13);
    });

    it('mfe', () => {
        const expected = 100.0 * (1.0 - 38.0 / 45.0);
        expect(rt.maximumFavorableExcursion).toBeCloseTo(expected, 13);
    });
});
