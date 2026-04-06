const std = @import("std");
const conventions = @import("conventions");
const dc = @import("daycounting");

const DayCountConvention = conventions.DayCountConvention;

const seconds_in_gregorian_year: f64 = 31556952;
const seconds_in_day: f64 = 60 * 60 * 24;

/// DateTime represents a date and time with year, month, day, hour, minute, second.
pub const DateTime = struct {
    year: i32,
    month: u8,
    day: u8,
    hour: u8 = 0,
    minute: u8 = 0,
    second: u8 = 0,

    /// Returns the total seconds from the Julian Day epoch for this DateTime.
    pub fn toTotalSeconds(self: DateTime) f64 {
        const jd = dc.dateToJD(self.year, @intCast(self.month), @intCast(self.day));
        const day_seconds = @as(f64, @floatFromInt(self.hour)) * 3600.0 +
            @as(f64, @floatFromInt(self.minute)) * 60.0 +
            @as(f64, @floatFromInt(self.second));
        return @as(f64, @floatFromInt(jd)) * 86400.0 + day_seconds;
    }

    /// Returns true if self is after other.
    fn isAfter(self: DateTime, other: DateTime) bool {
        return self.toTotalSeconds() > other.toTotalSeconds();
    }

    /// Returns the time-of-day fraction (0..1) representing hour/min/sec as a fraction of 86400.
    fn timeFraction(self: DateTime) f64 {
        return (@as(f64, @floatFromInt(self.hour)) * 3600.0 +
            @as(f64, @floatFromInt(self.minute)) * 60.0 +
            @as(f64, @floatFromInt(self.second))) / 86400.0;
    }
};

/// Calculates the fraction between two dates using a specified day count convention.
/// If day_frac is true, returns fraction in days; if false, returns fraction in years.
/// Returns an error for unknown conventions.
pub fn frac(date_time1: DateTime, date_time2: DateTime, method: DayCountConvention, day_frac: bool) !f64 {
    var dt1 = date_time1;
    var dt2 = date_time2;

    if (dt1.isAfter(dt2)) {
        const tmp = dt1;
        dt1 = dt2;
        dt2 = tmp;
    }

    if (method == .raw) {
        const diff_seconds = dt2.toTotalSeconds() - dt1.toTotalSeconds();
        if (day_frac) {
            return diff_seconds / seconds_in_day;
        }
        return diff_seconds / seconds_in_gregorian_year;
    }

    const y1: i32 = dt1.year;
    const m1: i32 = @intCast(dt1.month);
    const d1: i32 = @intCast(dt1.day);
    const y2: i32 = dt2.year;
    const m2: i32 = @intCast(dt2.month);
    const d2: i32 = @intCast(dt2.day);
    const tm1 = dt1.timeFraction();
    const tm2 = dt2.timeFraction();

    return dc.dispatch(method, y1, m1, d1, y2, m2, d2, tm1, tm2, day_frac) orelse unreachable;
}

/// Calculates the year fraction between two dates using a specified day count convention.
pub fn yearFrac(date_time1: DateTime, date_time2: DateTime, method: DayCountConvention) !f64 {
    return frac(date_time1, date_time2, method, false);
}

/// Calculates the day fraction between two dates using a specified day count convention.
pub fn dayFrac(date_time1: DateTime, date_time2: DateTime, method: DayCountConvention) !f64 {
    return frac(date_time1, date_time2, method, true);
}

// ============== Tests ==============

const epsilon = 1e-14;
const seconds_in_leap_year: f64 = 31622400;
const seconds_in_non_leap_year: f64 = 31536000;

fn almostEqual(a: f64, b: f64, tolerance: f64) bool {
    return @abs(a - b) <= tolerance;
}

test "yearFrac RAW leap year" {
    const dt1 = DateTime{ .year = 2020, .month = 1, .day = 1 };
    const dt2 = DateTime{ .year = 2021, .month = 1, .day = 1 };
    const result = try yearFrac(dt1, dt2, .raw);
    const expected = seconds_in_leap_year / seconds_in_gregorian_year;
    try std.testing.expect(almostEqual(result, expected, 1e-15));
}

test "yearFrac RAW non-leap year" {
    const dt1 = DateTime{ .year = 2021, .month = 1, .day = 1 };
    const dt2 = DateTime{ .year = 2022, .month = 1, .day = 1 };
    const result = try yearFrac(dt1, dt2, .raw);
    const expected = seconds_in_non_leap_year / seconds_in_gregorian_year;
    try std.testing.expect(almostEqual(result, expected, 1e-15));
}

test "yearFrac valid methods" {
    const methods = [_]DayCountConvention{
        .thirty_360_us,
        .thirty_360_us_eom,
        .thirty_360_us_nasd,
        .thirty_360_eu,
        .thirty_360_eu_m2,
        .thirty_360_eu_m3,
        .thirty_360_eu_plus,
        .thirty_365,
        .act_360,
        .act_365_fixed,
        .act_365_nonleap,
        .act_act_excel,
        .act_act_isda,
        .act_act_afb,
    };

    const dt1 = DateTime{ .year = 2020, .month = 1, .day = 1 };
    const dt2 = DateTime{ .year = 2021, .month = 1, .day = 1 };

    for (methods) |method| {
        const result = try yearFrac(dt1, dt2, method);
        const expected = dc.dispatch(method, 2020, 1, 1, 2021, 1, 1, 0, 0, false) orelse unreachable;
        try std.testing.expect(almostEqual(result, expected, 1e-15));
    }
}

test "dayFrac RAW" {
    const dt1 = DateTime{ .year = 2020, .month = 1, .day = 1 };
    const dt2 = DateTime{ .year = 2020, .month = 1, .day = 2 };
    const result = try dayFrac(dt1, dt2, .raw);
    try std.testing.expect(almostEqual(result, 1.0, 1e-10));
}

test "dayFrac valid methods" {
    const dt1 = DateTime{ .year = 2020, .month = 1, .day = 1 };
    const dt2 = DateTime{ .year = 2020, .month = 2, .day = 1 };

    const cases = [_]struct { method: DayCountConvention }{
        .{ .method = .thirty_360_us },
        .{ .method = .thirty_360_eu },
        .{ .method = .act_360 },
        .{ .method = .act_365_fixed },
    };

    for (cases) |tc| {
        const result = try dayFrac(dt1, dt2, tc.method);
        const expected = dc.dispatch(tc.method, 2020, 1, 1, 2020, 2, 1, 0, 0, true) orelse unreachable;
        try std.testing.expect(almostEqual(result, expected, 1e-15));
    }
}

test "frac swapped dates" {
    const dt1 = DateTime{ .year = 2020, .month = 1, .day = 1 };
    const dt2 = DateTime{ .year = 2021, .month = 1, .day = 1 };

    const result1 = try yearFrac(dt1, dt2, .act_365_fixed);
    const result2 = try yearFrac(dt2, dt1, .act_365_fixed);
    try std.testing.expect(almostEqual(result1, result2, epsilon));
}

test "frac with intraday times" {
    const dt1 = DateTime{ .year = 2020, .month = 1, .day = 1, .hour = 9, .minute = 30, .second = 0 };
    const dt2 = DateTime{ .year = 2020, .month = 1, .day = 1, .hour = 15, .minute = 45, .second = 0 };

    const result = try yearFrac(dt1, dt2, .raw);
    try std.testing.expect(result < 1.0);
    try std.testing.expect(result > 0.0);
}

test "dayFrac Eur30360" {
    const dt1 = DateTime{ .year = 2020, .month = 1, .day = 1 };
    const dt2 = DateTime{ .year = 2020, .month = 2, .day = 1 };
    const result = try dayFrac(dt1, dt2, .thirty_360_eu);
    try std.testing.expect(almostEqual(result, 30.0, epsilon));
}

test "yearFrac Eur30360" {
    const dt1 = DateTime{ .year = 2018, .month = 12, .day = 15 };
    const dt2 = DateTime{ .year = 2019, .month = 3, .day = 1 };
    const result = try yearFrac(dt1, dt2, .thirty_360_eu);
    try std.testing.expect(almostEqual(result, 0.21111111, 1e-8));
}

test "actual methods comparison" {
    const dt1 = DateTime{ .year = 2020, .month = 1, .day = 1 };
    const dt2 = DateTime{ .year = 2021, .month = 1, .day = 1 };

    const a360 = try yearFrac(dt1, dt2, .act_360);
    const a365 = try yearFrac(dt1, dt2, .act_365_fixed);
    const aae = try yearFrac(dt1, dt2, .act_act_excel);
    const aai = try yearFrac(dt1, dt2, .act_act_isda);

    // Act/360 should be larger than Act/365
    try std.testing.expect(a360 > a365);

    // ActAct methods should be ~1.0 for full year
    try std.testing.expect(almostEqual(aae, 1.0, 1e-10));
    try std.testing.expect(almostEqual(aai, 1.0, 1e-10));
}

test "Excel compatibility via yearFrac" {
    const cases = [_]struct {
        start: DateTime,
        end: DateTime,
        method: DayCountConvention,
        expected: f64,
        tolerance: f64,
    }{
        .{ .start = .{ .year = 2012, .month = 1, .day = 1 }, .end = .{ .year = 2012, .month = 7, .day = 30 }, .method = .thirty_360_eu, .expected = 0.58055556, .tolerance = 1e-8 },
        .{ .start = .{ .year = 2012, .month = 1, .day = 1 }, .end = .{ .year = 2012, .month = 7, .day = 30 }, .method = .act_act_excel, .expected = 0.576388888888889, .tolerance = 1e-3 },
        .{ .start = .{ .year = 2012, .month = 1, .day = 1 }, .end = .{ .year = 2012, .month = 7, .day = 30 }, .method = .act_365_fixed, .expected = 0.57808219, .tolerance = 1e-8 },
        .{ .start = .{ .year = 2012, .month = 1, .day = 1 }, .end = .{ .year = 2012, .month = 7, .day = 30 }, .method = .act_360, .expected = 211.0 / 360.0, .tolerance = 1e-12 },
    };

    for (cases) |tc| {
        const result = try yearFrac(tc.start, tc.end, tc.method);
        try std.testing.expect(almostEqual(result, tc.expected, tc.tolerance));
    }
}
