const std = @import("std");

/// DayCountConvention represents different day count conventions used in financial calculations.
pub const DayCountConvention = enum(u4) {
    /// RAW takes the difference in seconds between two dates and divides
    /// it by the number of seconds in a Gregorian year (31556952).
    raw = 0,

    /// 30/360 US (ISDA) or 30/360 (American Basic Rule). ISO 20022: A001.
    thirty_360_us = 1,

    /// 30/360 US End-Of-Month.
    thirty_360_us_eom = 2,

    /// 30/360 NASD.
    thirty_360_us_nasd = 3,

    /// 30/360 Eurobond Basis or 30/360 ICMA. ISO 20022: A011.
    thirty_360_eu = 4,

    /// 30E2/360 or Eurobond basis model 2. ISO 20022: A012.
    thirty_360_eu_m2 = 5,

    /// 30E3/360 or Eurobond basis model 3. ISO 20022: A013.
    thirty_360_eu_m3 = 6,

    /// 30E+/360.
    thirty_360_eu_plus = 7,

    /// 30/365. ISO 20022: A002.
    thirty_365 = 8,

    /// Actual/360. ISO 20022: A004.
    act_360 = 9,

    /// Actual/365 Fixed. ISO 20022: A005.
    act_365_fixed = 10,

    /// Actual/365 Non-Leap. ISO 20022: A014.
    act_365_nonleap = 11,

    /// Excel-compatible Actual/Actual (basis 1) method.
    act_act_excel = 12,

    /// Actual/Actual ISDA or Actual/365 ISDA. ISO 20022: A008.
    act_act_isda = 13,

    /// Actual/Actual AFB. ISO 20022: A010.
    act_act_afb = 14,
};

pub const FromStringError = error{UnknownConvention};

/// Converts a string representation to a DayCountConvention.
/// The comparison is case-insensitive.
/// Returns an error if the string does not match any known convention.
pub fn fromString(convention: []const u8) FromStringError!DayCountConvention {
    // Convert to lowercase in a stack buffer (max expected length ~20 chars)
    var buf: [64]u8 = undefined;
    if (convention.len > buf.len) return error.UnknownConvention;
    const lower = toLower(convention, &buf);

    const Map = struct {
        key: []const u8,
        val: DayCountConvention,
    };
    const entries = [_]Map{
        .{ .key = "raw", .val = .raw },
        .{ .key = "30/360 us", .val = .thirty_360_us },
        .{ .key = "30u/360", .val = .thirty_360_us },
        .{ .key = "30/360 us eom", .val = .thirty_360_us_eom },
        .{ .key = "30u/360 eom", .val = .thirty_360_us_eom },
        .{ .key = "30/360 us nasd", .val = .thirty_360_us_nasd },
        .{ .key = "30u/360 nasd", .val = .thirty_360_us_nasd },
        .{ .key = "30/360 eu", .val = .thirty_360_eu },
        .{ .key = "30e/360", .val = .thirty_360_eu },
        .{ .key = "30/360 eu2", .val = .thirty_360_eu_m2 },
        .{ .key = "30e2/360", .val = .thirty_360_eu_m2 },
        .{ .key = "30/360 eu3", .val = .thirty_360_eu_m3 },
        .{ .key = "30e3/360", .val = .thirty_360_eu_m3 },
        .{ .key = "30/360 eu+", .val = .thirty_360_eu_plus },
        .{ .key = "30e+/360", .val = .thirty_360_eu_plus },
        .{ .key = "30/365", .val = .thirty_365 },
        .{ .key = "act/360", .val = .act_360 },
        .{ .key = "act/365 fixed", .val = .act_365_fixed },
        .{ .key = "act/365 nonleap", .val = .act_365_nonleap },
        .{ .key = "act/act excel", .val = .act_act_excel },
        .{ .key = "act/act isda", .val = .act_act_isda },
        .{ .key = "act/365 isda", .val = .act_act_isda },
        .{ .key = "act/act afb", .val = .act_act_afb },
    };

    for (entries) |entry| {
        if (std.mem.eql(u8, lower, entry.key)) {
            return entry.val;
        }
    }

    return error.UnknownConvention;
}

fn toLower(s: []const u8, buf: []u8) []const u8 {
    for (s, 0..) |c, i| {
        buf[i] = if (c >= 'A' and c <= 'Z') c + 32 else c;
    }
    return buf[0..s.len];
}

// ============== Tests ==============

test "fromString: valid lowercase conventions" {
    const cases = .{
        .{ "raw", DayCountConvention.raw },
        .{ "30/360 us", DayCountConvention.thirty_360_us },
        .{ "30/360 us eom", DayCountConvention.thirty_360_us_eom },
        .{ "30/360 us nasd", DayCountConvention.thirty_360_us_nasd },
        .{ "30/360 eu", DayCountConvention.thirty_360_eu },
        .{ "30/360 eu2", DayCountConvention.thirty_360_eu_m2 },
        .{ "30/360 eu3", DayCountConvention.thirty_360_eu_m3 },
        .{ "30/360 eu+", DayCountConvention.thirty_360_eu_plus },
        .{ "30/365", DayCountConvention.thirty_365 },
        .{ "act/360", DayCountConvention.act_360 },
        .{ "act/365 fixed", DayCountConvention.act_365_fixed },
        .{ "act/365 nonleap", DayCountConvention.act_365_nonleap },
        .{ "act/act excel", DayCountConvention.act_act_excel },
        .{ "act/act isda", DayCountConvention.act_act_isda },
        .{ "act/act afb", DayCountConvention.act_act_afb },
    };

    inline for (cases) |c| {
        const result = try fromString(c[0]);
        try std.testing.expectEqual(c[1], result);
    }
}

test "fromString: case insensitive" {
    try std.testing.expectEqual(DayCountConvention.act_act_excel, try fromString("Act/Act Excel"));
    try std.testing.expectEqual(DayCountConvention.act_act_afb, try fromString("ACT/ACT AFB"));
    try std.testing.expectEqual(DayCountConvention.act_act_isda, try fromString("act/act ISDA"));
}

test "fromString: invalid convention" {
    const result = fromString("invalid convention");
    try std.testing.expectError(error.UnknownConvention, result);
}
