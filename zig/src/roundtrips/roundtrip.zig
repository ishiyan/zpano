const std = @import("std");
const exec_mod = @import("execution");
const side_mod = @import("side");
const fractional = @import("fractional");

const Execution = exec_mod.Execution;
const OrderSide = exec_mod.OrderSide;
const RoundtripSide = side_mod.RoundtripSide;
const DateTime = fractional.DateTime;

/// Roundtrip represents a completed round-trip trade with all computed metrics.
pub const Roundtrip = struct {
    side: RoundtripSide,
    quantity: f64,
    entry_year: i32,
    entry_month: u8,
    entry_day: u8,
    entry_hour: u8,
    entry_minute: u8,
    entry_second: u8,
    entry_price: f64,
    exit_year: i32,
    exit_month: u8,
    exit_day: u8,
    exit_hour: u8,
    exit_minute: u8,
    exit_second: u8,
    exit_price: f64,
    duration_seconds: f64,
    highest_price: f64,
    lowest_price: f64,
    commission: f64,
    gross_pnl: f64,
    net_pnl: f64,
    maximum_adverse_price: f64,
    maximum_favorable_price: f64,
    maximum_adverse_excursion: f64,
    maximum_favorable_excursion: f64,
    entry_efficiency: f64,
    exit_efficiency: f64,
    total_efficiency: f64,

    /// Creates a new Roundtrip from entry/exit executions and quantity.
    pub fn init(entry: Execution, exit: Execution, quantity: f64) Roundtrip {
        const rt_side: RoundtripSide = if (entry.side.isSell()) .short else .long;
        const entry_p = entry.price;
        const exit_p = exit.price;

        const pnl = if (rt_side == .short)
            quantity * (entry_p - exit_p)
        else
            quantity * (exit_p - entry_p);

        const comm = (entry.commission_per_unit + exit.commission_per_unit) * quantity;

        const highest_p = @max(entry.unrealized_price_high, exit.unrealized_price_high);
        const lowest_p = @min(entry.unrealized_price_low, exit.unrealized_price_low);
        const delta = highest_p - lowest_p;

        // Compute duration in seconds using DateTime
        const entry_dt = DateTime{
            .year = entry.year,
            .month = entry.month,
            .day = entry.day,
            .hour = entry.hour,
            .minute = entry.minute,
            .second = entry.second,
        };
        const exit_dt = DateTime{
            .year = exit.year,
            .month = exit.month,
            .day = exit.day,
            .hour = exit.hour,
            .minute = exit.minute,
            .second = exit.second,
        };
        const dur_secs = exit_dt.toTotalSeconds() - entry_dt.toTotalSeconds();

        var entry_eff: f64 = 0.0;
        var exit_eff: f64 = 0.0;
        var total_eff: f64 = 0.0;
        var max_adverse_price: f64 = undefined;
        var max_favorable_price: f64 = undefined;
        var mae: f64 = undefined;
        var mfe: f64 = undefined;

        if (rt_side == .long) {
            max_adverse_price = lowest_p;
            max_favorable_price = highest_p;
            mae = 100.0 * (1.0 - lowest_p / entry_p);
            mfe = 100.0 * (highest_p / exit_p - 1.0);
            if (delta != 0.0) {
                entry_eff = 100.0 * (highest_p - entry_p) / delta;
                exit_eff = 100.0 * (exit_p - lowest_p) / delta;
                total_eff = 100.0 * (exit_p - entry_p) / delta;
            }
        } else {
            max_adverse_price = highest_p;
            max_favorable_price = lowest_p;
            mae = 100.0 * (highest_p / entry_p - 1.0);
            mfe = 100.0 * (1.0 - lowest_p / exit_p);
            if (delta != 0.0) {
                entry_eff = 100.0 * (entry_p - lowest_p) / delta;
                exit_eff = 100.0 * (highest_p - exit_p) / delta;
                total_eff = 100.0 * (entry_p - exit_p) / delta;
            }
        }

        return .{
            .side = rt_side,
            .quantity = quantity,
            .entry_year = entry.year,
            .entry_month = entry.month,
            .entry_day = entry.day,
            .entry_hour = entry.hour,
            .entry_minute = entry.minute,
            .entry_second = entry.second,
            .entry_price = entry_p,
            .exit_year = exit.year,
            .exit_month = exit.month,
            .exit_day = exit.day,
            .exit_hour = exit.hour,
            .exit_minute = exit.minute,
            .exit_second = exit.second,
            .exit_price = exit_p,
            .duration_seconds = dur_secs,
            .highest_price = highest_p,
            .lowest_price = lowest_p,
            .commission = comm,
            .gross_pnl = pnl,
            .net_pnl = pnl - comm,
            .maximum_adverse_price = max_adverse_price,
            .maximum_favorable_price = max_favorable_price,
            .maximum_adverse_excursion = mae,
            .maximum_favorable_excursion = mfe,
            .entry_efficiency = entry_eff,
            .exit_efficiency = exit_eff,
            .total_efficiency = total_eff,
        };
    }

    /// Returns entry time as a DateTime.
    pub fn entryTime(self: Roundtrip) DateTime {
        return .{
            .year = self.entry_year,
            .month = self.entry_month,
            .day = self.entry_day,
            .hour = self.entry_hour,
            .minute = self.entry_minute,
            .second = self.entry_second,
        };
    }

    /// Returns exit time as a DateTime.
    pub fn exitTime(self: Roundtrip) DateTime {
        return .{
            .year = self.exit_year,
            .month = self.exit_month,
            .day = self.exit_day,
            .hour = self.exit_hour,
            .minute = self.exit_minute,
            .second = self.exit_second,
        };
    }
};

// ===========================================================================
// Tests
// ===========================================================================

const testing = std.testing;

fn almostEqual(a: f64, b: f64, epsilon: f64) bool {
    return @abs(a - b) < epsilon;
}

// ---------------------------------------------------------------------------
// Long roundtrip
// ---------------------------------------------------------------------------

test "roundtrip long side" {
    const rt = makeLongRt();
    try testing.expectEqual(RoundtripSide.long, rt.side);
}

test "roundtrip long quantity" {
    const rt = makeLongRt();
    try testing.expect(almostEqual(rt.quantity, 100.0, 1e-13));
}

test "roundtrip long entry price" {
    const rt = makeLongRt();
    try testing.expect(almostEqual(rt.entry_price, 50.0, 1e-13));
}

test "roundtrip long exit price" {
    const rt = makeLongRt();
    try testing.expect(almostEqual(rt.exit_price, 55.0, 1e-13));
}

test "roundtrip long duration" {
    const rt = makeLongRt();
    try testing.expect(almostEqual(rt.duration_seconds, 369000.0, 1e-13));
}

test "roundtrip long highest price" {
    const rt = makeLongRt();
    try testing.expect(almostEqual(rt.highest_price, 57.0, 1e-13));
}

test "roundtrip long lowest price" {
    const rt = makeLongRt();
    try testing.expect(almostEqual(rt.lowest_price, 48.0, 1e-13));
}

test "roundtrip long gross pnl" {
    const rt = makeLongRt();
    try testing.expect(almostEqual(rt.gross_pnl, 500.0, 1e-13));
}

test "roundtrip long commission" {
    const rt = makeLongRt();
    try testing.expect(almostEqual(rt.commission, 3.0, 1e-13));
}

test "roundtrip long net pnl" {
    const rt = makeLongRt();
    try testing.expect(almostEqual(rt.net_pnl, 497.0, 1e-13));
}

test "roundtrip long maximum adverse price" {
    const rt = makeLongRt();
    try testing.expect(almostEqual(rt.maximum_adverse_price, 48.0, 1e-13));
}

test "roundtrip long maximum favorable price" {
    const rt = makeLongRt();
    try testing.expect(almostEqual(rt.maximum_favorable_price, 57.0, 1e-13));
}

test "roundtrip long mae" {
    const rt = makeLongRt();
    try testing.expect(almostEqual(rt.maximum_adverse_excursion, 4.0, 1e-13));
}

test "roundtrip long mfe" {
    const rt = makeLongRt();
    const expected = 100.0 * (57.0 / 55.0 - 1.0);
    try testing.expect(almostEqual(rt.maximum_favorable_excursion, expected, 1e-13));
}

test "roundtrip long entry efficiency" {
    const rt = makeLongRt();
    const expected = 100.0 * (57.0 - 50.0) / 9.0;
    try testing.expect(almostEqual(rt.entry_efficiency, expected, 1e-13));
}

test "roundtrip long exit efficiency" {
    const rt = makeLongRt();
    const expected = 100.0 * (55.0 - 48.0) / 9.0;
    try testing.expect(almostEqual(rt.exit_efficiency, expected, 1e-13));
}

test "roundtrip long total efficiency" {
    const rt = makeLongRt();
    const expected = 100.0 * (55.0 - 50.0) / 9.0;
    try testing.expect(almostEqual(rt.total_efficiency, expected, 1e-13));
}

// ---------------------------------------------------------------------------
// Short roundtrip
// ---------------------------------------------------------------------------

test "roundtrip short side" {
    const rt = makeShortRt();
    try testing.expectEqual(RoundtripSide.short, rt.side);
}

test "roundtrip short quantity" {
    const rt = makeShortRt();
    try testing.expect(almostEqual(rt.quantity, 200.0, 1e-13));
}

test "roundtrip short entry price" {
    const rt = makeShortRt();
    try testing.expect(almostEqual(rt.entry_price, 80.0, 1e-13));
}

test "roundtrip short exit price" {
    const rt = makeShortRt();
    try testing.expect(almostEqual(rt.exit_price, 72.0, 1e-13));
}

test "roundtrip short duration" {
    const rt = makeShortRt();
    // Feb 1 10:00 to Feb 10 15:30 = 9 days 5 hours 30 min = 798600 sec
    const expected: f64 = 9.0 * 86400.0 + 5.0 * 3600.0 + 30.0 * 60.0;
    try testing.expect(almostEqual(rt.duration_seconds, expected, 1e-13));
}

test "roundtrip short highest price" {
    const rt = makeShortRt();
    try testing.expect(almostEqual(rt.highest_price, 85.0, 1e-13));
}

test "roundtrip short lowest price" {
    const rt = makeShortRt();
    try testing.expect(almostEqual(rt.lowest_price, 70.0, 1e-13));
}

test "roundtrip short gross pnl" {
    const rt = makeShortRt();
    try testing.expect(almostEqual(rt.gross_pnl, 1600.0, 1e-13));
}

test "roundtrip short commission" {
    const rt = makeShortRt();
    try testing.expect(almostEqual(rt.commission, 10.0, 1e-13));
}

test "roundtrip short net pnl" {
    const rt = makeShortRt();
    try testing.expect(almostEqual(rt.net_pnl, 1590.0, 1e-13));
}

test "roundtrip short maximum adverse price" {
    const rt = makeShortRt();
    try testing.expect(almostEqual(rt.maximum_adverse_price, 85.0, 1e-13));
}

test "roundtrip short maximum favorable price" {
    const rt = makeShortRt();
    try testing.expect(almostEqual(rt.maximum_favorable_price, 70.0, 1e-13));
}

test "roundtrip short mae" {
    const rt = makeShortRt();
    try testing.expect(almostEqual(rt.maximum_adverse_excursion, 6.25, 1e-13));
}

test "roundtrip short mfe" {
    const rt = makeShortRt();
    const expected = 100.0 * (1.0 - 70.0 / 72.0);
    try testing.expect(almostEqual(rt.maximum_favorable_excursion, expected, 1e-13));
}

test "roundtrip short entry efficiency" {
    const rt = makeShortRt();
    const expected = 100.0 * (80.0 - 70.0) / 15.0;
    try testing.expect(almostEqual(rt.entry_efficiency, expected, 1e-13));
}

test "roundtrip short exit efficiency" {
    const rt = makeShortRt();
    const expected = 100.0 * (85.0 - 72.0) / 15.0;
    try testing.expect(almostEqual(rt.exit_efficiency, expected, 1e-13));
}

test "roundtrip short total efficiency" {
    const rt = makeShortRt();
    const expected = 100.0 * (80.0 - 72.0) / 15.0;
    try testing.expect(almostEqual(rt.total_efficiency, expected, 1e-13));
}

// ---------------------------------------------------------------------------
// Zero delta
// ---------------------------------------------------------------------------

test "roundtrip zero delta entry efficiency" {
    const rt = makeZeroDeltaRt();
    try testing.expect(almostEqual(rt.entry_efficiency, 0.0, 1e-13));
}

test "roundtrip zero delta exit efficiency" {
    const rt = makeZeroDeltaRt();
    try testing.expect(almostEqual(rt.exit_efficiency, 0.0, 1e-13));
}

test "roundtrip zero delta total efficiency" {
    const rt = makeZeroDeltaRt();
    try testing.expect(almostEqual(rt.total_efficiency, 0.0, 1e-13));
}

test "roundtrip zero delta gross pnl" {
    const rt = makeZeroDeltaRt();
    try testing.expect(almostEqual(rt.gross_pnl, 0.0, 1e-13));
}

test "roundtrip zero delta net pnl" {
    const rt = makeZeroDeltaRt();
    try testing.expect(almostEqual(rt.net_pnl, 0.0, 1e-13));
}

// ---------------------------------------------------------------------------
// Long loser
// ---------------------------------------------------------------------------

test "roundtrip long loser side" {
    const rt = makeLongLoserRt();
    try testing.expectEqual(RoundtripSide.long, rt.side);
}

test "roundtrip long loser gross pnl" {
    const rt = makeLongLoserRt();
    try testing.expect(almostEqual(rt.gross_pnl, -900.0, 1e-13));
}

test "roundtrip long loser commission" {
    const rt = makeLongLoserRt();
    try testing.expect(almostEqual(rt.commission, 1.5, 1e-13));
}

test "roundtrip long loser net pnl" {
    const rt = makeLongLoserRt();
    try testing.expect(almostEqual(rt.net_pnl, -901.5, 1e-13));
}

test "roundtrip long loser highest" {
    const rt = makeLongLoserRt();
    try testing.expect(almostEqual(rt.highest_price, 62.0, 1e-13));
}

test "roundtrip long loser lowest" {
    const rt = makeLongLoserRt();
    try testing.expect(almostEqual(rt.lowest_price, 52.0, 1e-13));
}

test "roundtrip long loser mae" {
    const rt = makeLongLoserRt();
    const expected = 100.0 * (1.0 - 52.0 / 60.0);
    try testing.expect(almostEqual(rt.maximum_adverse_excursion, expected, 1e-13));
}

test "roundtrip long loser mfe" {
    const rt = makeLongLoserRt();
    const expected = 100.0 * (62.0 / 54.0 - 1.0);
    try testing.expect(almostEqual(rt.maximum_favorable_excursion, expected, 1e-13));
}

// ---------------------------------------------------------------------------
// Short loser
// ---------------------------------------------------------------------------

test "roundtrip short loser side" {
    const rt = makeShortLoserRt();
    try testing.expectEqual(RoundtripSide.short, rt.side);
}

test "roundtrip short loser gross pnl" {
    const rt = makeShortLoserRt();
    try testing.expect(almostEqual(rt.gross_pnl, -1500.0, 1e-13));
}

test "roundtrip short loser commission" {
    const rt = makeShortLoserRt();
    try testing.expect(almostEqual(rt.commission, 6.0, 1e-13));
}

test "roundtrip short loser net pnl" {
    const rt = makeShortLoserRt();
    try testing.expect(almostEqual(rt.net_pnl, -1506.0, 1e-13));
}

test "roundtrip short loser maximum adverse price" {
    const rt = makeShortLoserRt();
    try testing.expect(almostEqual(rt.maximum_adverse_price, 46.0, 1e-13));
}

test "roundtrip short loser maximum favorable price" {
    const rt = makeShortLoserRt();
    try testing.expect(almostEqual(rt.maximum_favorable_price, 38.0, 1e-13));
}

test "roundtrip short loser mae" {
    const rt = makeShortLoserRt();
    try testing.expect(almostEqual(rt.maximum_adverse_excursion, 15.0, 1e-13));
}

test "roundtrip short loser mfe" {
    const rt = makeShortLoserRt();
    const expected = 100.0 * (1.0 - 38.0 / 45.0);
    try testing.expect(almostEqual(rt.maximum_favorable_excursion, expected, 1e-13));
}

// ---------------------------------------------------------------------------
// Test data helpers
// ---------------------------------------------------------------------------

fn makeLongRt() Roundtrip {
    return Roundtrip.init(
        .{ .side = .buy, .price = 50.0, .commission_per_unit = 0.01, .unrealized_price_high = 56.0, .unrealized_price_low = 48.0, .year = 2024, .month = 1, .day = 1, .hour = 9, .minute = 30, .second = 0 },
        .{ .side = .sell, .price = 55.0, .commission_per_unit = 0.02, .unrealized_price_high = 57.0, .unrealized_price_low = 49.0, .year = 2024, .month = 1, .day = 5, .hour = 16, .minute = 0, .second = 0 },
        100.0,
    );
}

fn makeShortRt() Roundtrip {
    return Roundtrip.init(
        .{ .side = .sell, .price = 80.0, .commission_per_unit = 0.03, .unrealized_price_high = 85.0, .unrealized_price_low = 72.0, .year = 2024, .month = 2, .day = 1, .hour = 10, .minute = 0, .second = 0 },
        .{ .side = .buy, .price = 72.0, .commission_per_unit = 0.02, .unrealized_price_high = 83.0, .unrealized_price_low = 70.0, .year = 2024, .month = 2, .day = 10, .hour = 15, .minute = 30, .second = 0 },
        200.0,
    );
}

fn makeZeroDeltaRt() Roundtrip {
    return Roundtrip.init(
        .{ .side = .buy, .price = 100.0, .commission_per_unit = 0.0, .unrealized_price_high = 100.0, .unrealized_price_low = 100.0, .year = 2024, .month = 3, .day = 1, .hour = 9, .minute = 0, .second = 0 },
        .{ .side = .sell, .price = 100.0, .commission_per_unit = 0.0, .unrealized_price_high = 100.0, .unrealized_price_low = 100.0, .year = 2024, .month = 3, .day = 1, .hour = 10, .minute = 0, .second = 0 },
        50.0,
    );
}

fn makeLongLoserRt() Roundtrip {
    return Roundtrip.init(
        .{ .side = .buy, .price = 60.0, .commission_per_unit = 0.005, .unrealized_price_high = 62.0, .unrealized_price_low = 53.0, .year = 2024, .month = 4, .day = 1, .hour = 9, .minute = 30, .second = 0 },
        .{ .side = .sell, .price = 54.0, .commission_per_unit = 0.005, .unrealized_price_high = 61.0, .unrealized_price_low = 52.0, .year = 2024, .month = 4, .day = 3, .hour = 16, .minute = 0, .second = 0 },
        150.0,
    );
}

fn makeShortLoserRt() Roundtrip {
    return Roundtrip.init(
        .{ .side = .sell, .price = 40.0, .commission_per_unit = 0.01, .unrealized_price_high = 42.0, .unrealized_price_low = 39.0, .year = 2024, .month = 5, .day = 1, .hour = 10, .minute = 0, .second = 0 },
        .{ .side = .buy, .price = 45.0, .commission_per_unit = 0.01, .unrealized_price_high = 46.0, .unrealized_price_low = 38.0, .year = 2024, .month = 5, .day = 5, .hour = 15, .minute = 0, .second = 0 },
        300.0,
    );
}
