const std = @import("std");
const conventions = @import("conventions");

/// Returns true if the given year is a leap year.
pub fn isLeapYear(y: i32) bool {
    return @mod(y, 4) == 0 and (@mod(y, 100) != 0 or @mod(y, 400) == 0);
}

/// Converts a date to Julian Day number.
///
/// Algorithm adapted from
/// Press, W. H., Teukolsky, S. A., Vetterling, W. T., & Flannery, B. P. (2007).
/// Numerical Recipes: The Art of Scientific Computing (3rd ed.). Cambridge University Press.
pub fn dateToJD(year: i32, month: i32, day: i32) i32 {
    const a = @divTrunc(14 - month, 12);
    const y = year + 4800 - a;
    const m = month + (12 * a) - 3;
    var jd = day + @divTrunc(153 * m + 2, 5) + y * 365;
    jd += @divTrunc(y, 4) - @divTrunc(y, 100) + @divTrunc(y, 400) - 32045;
    return jd;
}

/// Converts a Julian Day number to a date (year, month, day).
///
/// Algorithm adapted from
/// Press, W. H., Teukolsky, S. A., Vetterling, W. T., & Flannery, B. P. (2007).
/// Numerical Recipes: The Art of Scientific Computing (3rd ed.). Cambridge University Press.
pub fn jdToDate(jd: i32) struct { year: i32, month: i32, day: i32 } {
    const a = jd + 32044;
    const b = @divTrunc(4 * a + 3, 146097);
    const c = a - @divTrunc(b * 146097, 4);
    const d = @divTrunc(4 * c + 3, 1461);
    const e = c - @divTrunc(d * 1461, 4);
    const m = @divTrunc(5 * e + 2, 153);
    const m2 = @divTrunc(m, 10);

    return .{
        .day = e + 1 - @divTrunc(153 * m + 2, 5),
        .month = m + 3 - 12 * m2,
        .year = b * 100 + d - 4800 + m2,
    };
}

/// Returns true if Feb 29 falls between date1 and date2 (inclusive) for any year in range.
/// Uses JD-based comparison instead of time.Time.
fn feb29Between(y1: i32, m1: i32, d1: i32, y2: i32, m2: i32, d2: i32) bool {
    const jd1 = dateToJD(y1, m1, d1);
    const jd2 = dateToJD(y2, m2, d2);
    var y: i32 = y1;
    while (y <= y2) : (y += 1) {
        if (isLeapYear(y)) {
            const leapJD = dateToJD(y, 2, 29);
            if (jd1 <= leapJD and leapJD <= jd2) {
                return true;
            }
        }
    }
    return false;
}

/// Returns true if date1 and date2 "appear" to be 1 year or less apart.
fn appearsLeYear(y1: i32, m1: i32, d1: i32, y2: i32, m2: i32, d2: i32) bool {
    if (y1 == y2) return true;
    if (y1 + 1 == y2 and (m1 > m2 or (m1 == m2 and d1 >= d2))) return true;
    return false;
}

// ---- Day counting functions ----

/// 30/360 European (Eurobond Basis / ICMA). ISO 20022: A011.
pub fn eur30360(y1: i32, m1: i32, d1: i32, y2: i32, m2: i32, d2: i32, df1: f64, df2: f64, frac_days: bool) f64 {
    var diff_days: f64 = @as(f64, @floatFromInt(360 * (y2 - y1) + 30 * (m2 - m1))) + df2 - df1;
    const d2_adj: i32 = @min(d2, 30);
    const d1_adj: i32 = @min(d1, 30);
    diff_days += @as(f64, @floatFromInt(d2_adj - d1_adj));
    if (frac_days) return diff_days;
    return diff_days / 360.0;
}

/// 30E2/360 (Eurobond basis model 2). ISO 20022: A012.
pub fn eur30360Model2(y1: i32, m1: i32, d1: i32, y2: i32, m2: i32, d2: i32, df1: f64, df2: f64, frac_days: bool) f64 {
    var diff_days: f64 = @as(f64, @floatFromInt(360 * (y2 - y1) + 30 * (m2 - m1))) + df2 - df1;
    const leap1 = isLeapYear(y1);
    var d2_adj = d2;

    if (leap1 and m2 == 2 and d2 == 28) {
        if (d1 == 29) {
            d2_adj = 29;
        } else if (d1 >= 30) {
            d2_adj = 30;
        }
    } else if (leap1 and m2 == 2 and d2 == 29) {
        if (d1 >= 30) {
            d2_adj = 30;
        }
    } else if (d2 > 30) {
        d2_adj = 30;
    }

    const d1_adj: i32 = @min(d1, 30);
    diff_days += @as(f64, @floatFromInt(d2_adj - d1_adj));
    if (frac_days) return diff_days;
    return diff_days / 360.0;
}

/// 30E3/360 (Eurobond basis model 3). ISO 20022: A013.
pub fn eur30360Model3(y1: i32, m1: i32, d1: i32, y2: i32, m2: i32, d2: i32, df1: f64, df2: f64, frac_days: bool) f64 {
    var diff_days: f64 = @as(f64, @floatFromInt(360 * (y2 - y1) + 30 * (m2 - m1))) + df2 - df1;

    var d2_adj = d2;
    if (m2 == 2 and d2 >= 28) {
        d2_adj = 30;
    } else if (d2 > 30) {
        d2_adj = 30;
    }

    var d1_adj = d1;
    if (m1 == 2 and d1 >= 28) {
        d1_adj = 30;
    } else if (d1 > 30) {
        d1_adj = 30;
    }

    diff_days += @as(f64, @floatFromInt(d2_adj - d1_adj));
    if (frac_days) return diff_days;
    return diff_days / 360.0;
}

/// 30E+/360.
pub fn eur30360Plus(y1: i32, m1: i32, d1: i32, y2: i32, m2: i32, d2: i32, df1: f64, df2: f64, frac_days: bool) f64 {
    var diff_days: f64 = @as(f64, @floatFromInt(360 * (y2 - y1) + 30 * (m2 - m1))) + df2 - df1;

    var d2_adj = d2;
    if (d2 == 31) {
        d2_adj = 32;
    }

    var d1_adj = d1;
    if (d1 > 30) {
        d1_adj = 30;
    }

    diff_days += @as(f64, @floatFromInt(d2_adj - d1_adj));
    if (frac_days) return diff_days;
    return diff_days / 360.0;
}

/// 30/360 US (ISDA). ISO 20022: A001.
pub fn us30360(y1: i32, m1: i32, d1: i32, y2: i32, m2: i32, d2: i32, df1: f64, df2: f64, frac_days: bool) f64 {
    var diff_days: f64 = @as(f64, @floatFromInt(360 * (y2 - y1) + 30 * (m2 - m1))) + df2 - df1;

    var d2_adj = d2;
    if (d2 == 31 and d1 >= 30) {
        d2_adj = 30;
    }

    const d1_adj: i32 = @min(d1, 30);
    diff_days += @as(f64, @floatFromInt(d2_adj - d1_adj));
    if (frac_days) return diff_days;
    return diff_days / 360.0;
}

/// 30/360 US End-Of-Month.
pub fn us30360Eom(y1: i32, m1: i32, d1: i32, y2: i32, m2: i32, d2: i32, df1: f64, df2: f64, frac_days: bool) f64 {
    var diff_days: f64 = @as(f64, @floatFromInt(360 * (y2 - y1) + 30 * (m2 - m1))) + df2 - df1;

    const rule2 = m1 == 2 and d1 >= 28;
    const rule3 = rule2 and m2 == 2 and d2 >= 28;
    const rule4 = d2 == 31 and d1 >= 30;

    var d1_adj = d1;
    if (rule2) {
        d1_adj = 30;
    } else if (d1 > 30) {
        d1_adj = 30;
    }

    var d2_adj = d2;
    if (rule4 or rule3) {
        d2_adj = 30;
    }

    diff_days += @as(f64, @floatFromInt(d2_adj - d1_adj));
    if (frac_days) return diff_days;
    return diff_days / 360.0;
}

/// 30/360 NASD.
pub fn us30360Nasd(y1: i32, m1: i32, d1: i32, y2: i32, m2: i32, d2: i32, df1: f64, df2: f64, frac_days: bool) f64 {
    var diff_days: f64 = @as(f64, @floatFromInt(360 * (y2 - y1) + 30 * (m2 - m1))) + df2 - df1;

    var d2_adj = d2;
    if (d2 == 31) {
        if (d1 < 30) {
            d2_adj = 32;
        } else {
            d2_adj = 30;
        }
    }

    const d1_adj: i32 = @min(d1, 30);
    diff_days += @as(f64, @floatFromInt(d2_adj - d1_adj));
    if (frac_days) return diff_days;
    return diff_days / 360.0;
}

/// 30/365. ISO 20022: A002.
pub fn thirty365(y1: i32, m1: i32, d1: i32, y2: i32, m2: i32, d2: i32, df1: f64, df2: f64, frac_days: bool) f64 {
    var diff_days: f64 = @as(f64, @floatFromInt(360 * (y2 - y1) + 30 * (m2 - m1))) + df2 - df1;

    var d2_adj = d2;
    if (d2 == 31 and d1 >= 30) {
        d2_adj = 30;
    }

    const d1_adj: i32 = @min(d1, 30);
    diff_days += @as(f64, @floatFromInt(d2_adj - d1_adj));
    if (frac_days) return diff_days;
    return diff_days / 365.0;
}

/// Actual/365 Non-Leap. ISO 20022: A014.
pub fn act365Nonleap(y1: i32, m1: i32, d1: i32, y2: i32, m2: i32, d2: i32, df1: f64, df2: f64, frac_days: bool) f64 {
    var diff_days: f64 = @as(f64, @floatFromInt(dateToJD(y2, m2, d2) - dateToJD(y1, m1, d1))) + df2 - df1;

    var leap_years: i32 = 0;
    if (isLeapYear(y1) and m1 <= 2) {
        leap_years += 1;
    }
    if (y1 != y2 and isLeapYear(y2) and m2 >= 3) {
        leap_years += 1;
    }
    if (y1 + 1 < y2) {
        var now = y1 + 1;
        while (now < y2) : (now += 1) {
            if (isLeapYear(now)) {
                leap_years += 1;
            }
        }
    }

    diff_days -= @as(f64, @floatFromInt(leap_years));
    if (frac_days) return diff_days;
    return diff_days / 365.0;
}

/// Actual/365 Fixed. ISO 20022: A005.
pub fn act365Fixed(y1: i32, m1: i32, d1: i32, y2: i32, m2: i32, d2: i32, df1: f64, df2: f64, frac_days: bool) f64 {
    const diff_days: f64 = @as(f64, @floatFromInt(dateToJD(y2, m2, d2) - dateToJD(y1, m1, d1))) + df2 - df1;
    if (frac_days) return diff_days;
    return diff_days / 365.0;
}

/// Actual/360. ISO 20022: A004.
pub fn act360(y1: i32, m1: i32, d1: i32, y2: i32, m2: i32, d2: i32, df1: f64, df2: f64, frac_days: bool) f64 {
    const diff_days: f64 = @as(f64, @floatFromInt(dateToJD(y2, m2, d2) - dateToJD(y1, m1, d1))) + df2 - df1;
    if (frac_days) return diff_days;
    return diff_days / 360.0;
}

/// Excel-compatible Actual/Actual (basis 1) method.
pub fn actActExcel(y1: i32, m1: i32, d1: i32, y2: i32, m2: i32, d2: i32, df1: f64, df2: f64, frac_days: bool) f64 {
    if (appearsLeYear(y1, m1, d1, y2, m2, d2)) {
        var year_days: f64 = undefined;
        if (y1 == y2 and isLeapYear(y1)) {
            year_days = 366.0;
        } else if (feb29Between(y1, m1, d1, y2, m2, d2) or (m2 == 2 and d2 == 29)) {
            year_days = 366.0;
        } else {
            year_days = 365.0;
        }
        const df: f64 = @as(f64, @floatFromInt(dateToJD(y2, m2, d2) - dateToJD(y1, m1, d1)));
        if (frac_days) return df + df2 - df1;
        return (df + df2 - df1) / year_days;
    } else {
        const jd_start1 = dateToJD(y1, 1, 1);
        const jd_start2 = dateToJD(y2 + 1, 1, 1);
        const year_days: f64 = @as(f64, @floatFromInt(jd_start2 - jd_start1));
        const avg_year_days: f64 = year_days / @as(f64, @floatFromInt(y2 - y1 + 1));
        const df: f64 = @as(f64, @floatFromInt(dateToJD(y2, m2, d2) - dateToJD(y1, m1, d1)));
        if (frac_days) return df + df2 - df1;
        return (df + df2 - df1) / avg_year_days;
    }
}

/// Actual/Actual ISDA. ISO 20022: A008.
pub fn actActIsda(y1: i32, m1: i32, d1: i32, y2: i32, m2: i32, d2: i32, df1: f64, df2: f64, frac_days: bool) f64 {
    if (y1 == y2) {
        var denom: f64 = 365.0;
        if (isLeapYear(y2)) denom = 366.0;
        const diff_days: f64 = @as(f64, @floatFromInt(dateToJD(y2, m2, d2) - dateToJD(y1, m1, d1))) + df2 - df1;
        if (frac_days) return diff_days;
        return diff_days / denom;
    }

    var denom_a: f64 = 365.0;
    if (isLeapYear(y1)) denom_a = 366.0;
    const diff_a: f64 = @as(f64, @floatFromInt(dateToJD(y1, 12, 31) - dateToJD(y1, m1, d1) + 1));

    var denom_b: f64 = 365.0;
    if (isLeapYear(y2)) denom_b = 366.0;
    const diff_b: f64 = @as(f64, @floatFromInt(dateToJD(y2, m2, d2) - dateToJD(y2, 1, 1)));

    if (frac_days) {
        var diff = diff_a - df1 + diff_b + df2;
        var year = y1 + 1;
        while (year < y2) : (year += 1) {
            if (isLeapYear(year)) {
                diff += 366;
            } else {
                diff += 365;
            }
        }
        return diff;
    }

    return (diff_a - df1) / denom_a + (diff_b + df2) / denom_b + @as(f64, @floatFromInt(y2 - y1 - 1));
}

/// Actual/Actual AFB. ISO 20022: A010.
pub fn actActAfb(y1: i32, m1: i32, d1: i32, y2: i32, m2: i32, d2: i32, df1: f64, df2: f64, frac_days: bool) f64 {
    if (y1 == y2) {
        var denom: f64 = 365.0;
        if (m1 < 3 and isLeapYear(y1)) denom = 366.0;
        const diff_days: f64 = @as(f64, @floatFromInt(dateToJD(y2, m2, d2) - dateToJD(y1, m1, d1))) + df2 - df1;
        if (frac_days) return diff_days;
        return diff_days / denom;
    }

    var denom_a: f64 = 365.0;
    if (m1 < 3 and isLeapYear(y1)) denom_a = 366.0;
    const diff_a: f64 = @as(f64, @floatFromInt(dateToJD(y1, 12, 31) - dateToJD(y1, m1, d1) + 1));

    var denom_b: f64 = 365.0;
    if (m2 >= 3 and isLeapYear(y2)) denom_b = 366.0;
    const diff_b: f64 = @as(f64, @floatFromInt(dateToJD(y2, m2, d2) - dateToJD(y2, 1, 1)));

    if (frac_days) {
        var diff = diff_a - df1 + diff_b + df2;
        var year = y1 + 1;
        while (year < y2) : (year += 1) {
            if (isLeapYear(year)) {
                diff += 366;
            } else {
                diff += 365;
            }
        }
        return diff;
    }

    return (diff_a - df1) / denom_a + (diff_b + df2) / denom_b + @as(f64, @floatFromInt(y2 - y1 - 1));
}

/// Dispatches to the appropriate day count function based on the convention.
/// Returns null for RAW convention (handled separately in fractional.zig).
pub fn dispatch(convention: conventions.DayCountConvention, y1: i32, m1: i32, d1: i32, y2: i32, m2: i32, d2: i32, df1: f64, df2: f64, frac_days: bool) ?f64 {
    return switch (convention) {
        .raw => null,
        .thirty_360_us => us30360(y1, m1, d1, y2, m2, d2, df1, df2, frac_days),
        .thirty_360_us_eom => us30360Eom(y1, m1, d1, y2, m2, d2, df1, df2, frac_days),
        .thirty_360_us_nasd => us30360Nasd(y1, m1, d1, y2, m2, d2, df1, df2, frac_days),
        .thirty_360_eu => eur30360(y1, m1, d1, y2, m2, d2, df1, df2, frac_days),
        .thirty_360_eu_m2 => eur30360Model2(y1, m1, d1, y2, m2, d2, df1, df2, frac_days),
        .thirty_360_eu_m3 => eur30360Model3(y1, m1, d1, y2, m2, d2, df1, df2, frac_days),
        .thirty_360_eu_plus => eur30360Plus(y1, m1, d1, y2, m2, d2, df1, df2, frac_days),
        .thirty_365 => thirty365(y1, m1, d1, y2, m2, d2, df1, df2, frac_days),
        .act_360 => act360(y1, m1, d1, y2, m2, d2, df1, df2, frac_days),
        .act_365_fixed => act365Fixed(y1, m1, d1, y2, m2, d2, df1, df2, frac_days),
        .act_365_nonleap => act365Nonleap(y1, m1, d1, y2, m2, d2, df1, df2, frac_days),
        .act_act_excel => actActExcel(y1, m1, d1, y2, m2, d2, df1, df2, frac_days),
        .act_act_isda => actActIsda(y1, m1, d1, y2, m2, d2, df1, df2, frac_days),
        .act_act_afb => actActAfb(y1, m1, d1, y2, m2, d2, df1, df2, frac_days),
    };
}

// ============== Tests ==============

const epsilon = 1e-14;
const fd2_360: f64 = 0.2 / 360.0;
const fd2_365: f64 = 0.2 / 365.0;
const fd2_366: f64 = 0.2 / 366.0;

fn almostEqual(a: f64, b: f64, tolerance: f64) bool {
    return @abs(a - b) <= tolerance;
}

fn pow10(n: i32) f64 {
    return std.math.pow(f64, 10.0, @as(f64, @floatFromInt(-n)));
}

test "isLeapYear" {
    const leap_years = [_]i32{
        1804, 1808, 1812, 1816, 1820, 1824, 1828, 1832, 1836, 1840, 1844,
        1848, 1852, 1856, 1860, 1864, 1868, 1872, 1876, 1880, 1884, 1888,
        1892, 1896, 1904, 1908, 1912, 1916, 1920, 1924, 1928, 1932, 1936,
        1940, 1944, 1948, 1952, 1956, 1960, 1964, 1968, 1972, 1976, 1980,
        1984, 1988, 1992, 1996, 2000, 2004, 2008, 2012, 2016, 2020, 2024,
        2028, 2032, 2036, 2040, 2044, 2048, 2052, 2056, 2060, 2064, 2068,
        2072, 2076, 2080, 2084, 2088, 2092, 2096, 2104, 2108, 2112, 2116,
        2120, 2124, 2128, 2132, 2136, 2140, 2144, 2148, 2152, 2156, 2160,
        2164, 2168, 2172, 2176, 2180, 2184, 2188, 2192, 2196, 2204, 2208,
        2212, 2216, 2220, 2224, 2228, 2232, 2236, 2240, 2244, 2248, 2252,
        2256, 2260, 2264, 2268, 2272, 2276, 2280, 2284, 2288, 2292, 2296,
        2304, 2308, 2312, 2316, 2320, 2324, 2328, 2332, 2336, 2340, 2344,
        2348, 2352, 2356, 2360, 2364, 2368, 2372, 2376, 2380, 2384, 2388,
        2392, 2396, 2400,
    };
    for (leap_years) |year| {
        try std.testing.expect(isLeapYear(year));
    }

    const non_leap_years = [_]i32{ 2017, 2018, 2019, 2021, 2022, 2023, 2025, 2026, 2027, 2029, 2030 };
    for (non_leap_years) |year| {
        try std.testing.expect(!isLeapYear(year));
    }
}

test "julianDayConversion" {
    const cases = [_]struct { jd: i32, year: i32, month: i32, day: i32 }{
        .{ .jd = 0, .year = -4713, .month = 11, .day = 24 },
        .{ .jd = 1, .year = -4713, .month = 11, .day = 25 },
        .{ .jd = 2456700, .year = 2014, .month = 2, .day = 11 },
        .{ .jd = 4168242, .year = 6700, .month = 2, .day = 27 },
        .{ .jd = 4168243, .year = 6700, .month = 2, .day = 28 },
        .{ .jd = 4168244, .year = 6700, .month = 3, .day = 1 },
        .{ .jd = 4168245, .year = 6700, .month = 3, .day = 2 },
    };

    for (cases) |tc| {
        const result = jdToDate(tc.jd);
        try std.testing.expectEqual(tc.year, result.year);
        try std.testing.expectEqual(tc.month, result.month);
        try std.testing.expectEqual(tc.day, result.day);

        const jd = dateToJD(tc.year, tc.month, tc.day);
        try std.testing.expectEqual(tc.jd, jd);
    }
}

test "eur30360: basic" {
    const result = eur30360(2018, 12, 15, 2019, 3, 1, 0, 0, false);
    try std.testing.expect(almostEqual(result, 0.21111111, 1e-8));
}

test "eur30360: time fractions" {
    const cases = [_]struct { y1: i32, m1: i32, d1: i32, y2: i32, m2: i32, d2: i32, df1: f64, df2: f64, expected: f64, tolerance: f64 }{
        .{ .y1 = 2021, .m1 = 1, .d1 = 1, .y2 = 2022, .m2 = 1, .d2 = 1, .df1 = 0.5, .df2 = 0.5, .expected = 1, .tolerance = 1e-16 },
        .{ .y1 = 2020, .m1 = 1, .d1 = 1, .y2 = 2021, .m2 = 1, .d2 = 1, .df1 = 0.5, .df2 = 0.5, .expected = 1, .tolerance = 1e-16 },
        .{ .y1 = 2020, .m1 = 1, .d1 = 1, .y2 = 2021, .m2 = 1, .d2 = 1, .df1 = 0.4, .df2 = 0.6, .expected = 1 + fd2_360, .tolerance = 1e-16 },
        .{ .y1 = 2020, .m1 = 1, .d1 = 1, .y2 = 2021, .m2 = 1, .d2 = 1, .df1 = 0.6, .df2 = 0.4, .expected = 1 - fd2_360, .tolerance = 1e-15 },
        .{ .y1 = 2020, .m1 = 1, .d1 = 1, .y2 = 2020, .m2 = 1, .d2 = 1, .df1 = 0.4, .df2 = 0.6, .expected = fd2_360, .tolerance = 1e-16 },
    };
    for (cases) |tc| {
        const result = eur30360(tc.y1, tc.m1, tc.d1, tc.y2, tc.m2, tc.d2, tc.df1, tc.df2, false);
        try std.testing.expect(almostEqual(result, tc.expected, tc.tolerance));
    }
}

test "eur30360: year fractions" {
    const result = eur30360(2012, 1, 1, 2012, 7, 30, 0, 0, false);
    try std.testing.expect(almostEqual(result, 0.58055556, 1e-8));
}

test "eur30360: Excel basis 4" {
    const cases = [_]struct { y1: i32, m1: i32, d1: i32, y2: i32, m2: i32, d2: i32, expected: f64, precision: i32 }{
        .{ .y1 = 1978, .m1 = 2, .d1 = 28, .y2 = 2020, .m2 = 5, .d2 = 17, .expected = 42.21944444444444, .precision = 13 },
        .{ .y1 = 1993, .m1 = 12, .d1 = 2, .y2 = 2022, .m2 = 4, .d2 = 18, .expected = 28.37777777777780, .precision = 13 },
        .{ .y1 = 2018, .m1 = 12, .d1 = 15, .y2 = 2019, .m2 = 3, .d2 = 1, .expected = 0.211111111111111, .precision = 13 },
        .{ .y1 = 2018, .m1 = 12, .d1 = 31, .y2 = 2019, .m2 = 1, .d2 = 1, .expected = 0.0027777777777778, .precision = 13 },
        .{ .y1 = 1994, .m1 = 6, .d1 = 30, .y2 = 1997, .m2 = 6, .d2 = 30, .expected = 3.0000000000000000, .precision = 16 },
        .{ .y1 = 1994, .m1 = 2, .d1 = 10, .y2 = 1994, .m2 = 6, .d2 = 30, .expected = 0.3888888888888889, .precision = 13 },
        .{ .y1 = 2020, .m1 = 2, .d1 = 21, .y2 = 2024, .m2 = 3, .d2 = 25, .expected = 4.0944444444444440, .precision = 13 },
        .{ .y1 = 2020, .m1 = 2, .d1 = 29, .y2 = 2021, .m2 = 2, .d2 = 28, .expected = 0.9972222222222222, .precision = 13 },
        .{ .y1 = 2020, .m1 = 1, .d1 = 31, .y2 = 2021, .m2 = 2, .d2 = 28, .expected = 1.0777777777777777, .precision = 13 },
        .{ .y1 = 2020, .m1 = 1, .d1 = 31, .y2 = 2021, .m2 = 3, .d2 = 31, .expected = 1.1666666666666667, .precision = 13 },
        .{ .y1 = 2020, .m1 = 1, .d1 = 31, .y2 = 2020, .m2 = 4, .d2 = 30, .expected = 0.2500000000000000, .precision = 16 },
        .{ .y1 = 2018, .m1 = 2, .d1 = 5, .y2 = 2023, .m2 = 5, .d2 = 14, .expected = 5.2750000000000000, .precision = 16 },
        .{ .y1 = 2020, .m1 = 2, .d1 = 29, .y2 = 2024, .m2 = 2, .d2 = 28, .expected = 3.9972222222222222, .precision = 13 },
        .{ .y1 = 2010, .m1 = 3, .d1 = 31, .y2 = 2015, .m2 = 8, .d2 = 30, .expected = 5.4166666666666667, .precision = 13 },
        .{ .y1 = 2016, .m1 = 2, .d1 = 28, .y2 = 2016, .m2 = 10, .d2 = 30, .expected = 0.6722222222222222, .precision = 13 },
        .{ .y1 = 2014, .m1 = 1, .d1 = 31, .y2 = 2014, .m2 = 8, .d2 = 31, .expected = 0.5833333333333333, .precision = 13 },
        .{ .y1 = 2014, .m1 = 2, .d1 = 28, .y2 = 2014, .m2 = 9, .d2 = 30, .expected = 0.5888888888888889, .precision = 13 },
        .{ .y1 = 2016, .m1 = 2, .d1 = 29, .y2 = 2016, .m2 = 6, .d2 = 15, .expected = 0.29444444444444445, .precision = 13 },
        .{ .y1 = 2024, .m1 = 1, .d1 = 1, .y2 = 2024, .m2 = 12, .d2 = 3, .expected = 0.9222222222222223, .precision = 13 },
        .{ .y1 = 2024, .m1 = 1, .d1 = 1, .y2 = 2025, .m2 = 1, .d2 = 2, .expected = 1.0027777777777800, .precision = 13 },
        .{ .y1 = 2024, .m1 = 1, .d1 = 1, .y2 = 2024, .m2 = 2, .d2 = 29, .expected = 0.1611111111111110, .precision = 13 },
        .{ .y1 = 2024, .m1 = 1, .d1 = 1, .y2 = 2024, .m2 = 3, .d2 = 1, .expected = 0.1666666666666670, .precision = 13 },
        .{ .y1 = 2023, .m1 = 1, .d1 = 1, .y2 = 2023, .m2 = 3, .d2 = 1, .expected = 0.1666666666666670, .precision = 13 },
        .{ .y1 = 2024, .m1 = 2, .d1 = 29, .y2 = 2025, .m2 = 2, .d2 = 28, .expected = 0.9972222222222220, .precision = 13 },
        .{ .y1 = 2024, .m1 = 1, .d1 = 1, .y2 = 2028, .m2 = 12, .d2 = 31, .expected = 4.9972222222222200, .precision = 13 },
        .{ .y1 = 2024, .m1 = 3, .d1 = 1, .y2 = 2025, .m2 = 3, .d2 = 1, .expected = 1.0000000000000000, .precision = 16 },
        .{ .y1 = 2024, .m1 = 2, .d1 = 29, .y2 = 2025, .m2 = 3, .d2 = 1, .expected = 1.0055555555555600, .precision = 13 },
        .{ .y1 = 2024, .m1 = 2, .d1 = 29, .y2 = 2028, .m2 = 2, .d2 = 28, .expected = 3.9972222222222200, .precision = 13 },
        .{ .y1 = 2024, .m1 = 2, .d1 = 29, .y2 = 2028, .m2 = 2, .d2 = 29, .expected = 4.0000000000000000, .precision = 16 },
        .{ .y1 = 2024, .m1 = 3, .d1 = 1, .y2 = 2028, .m2 = 3, .d2 = 1, .expected = 4.0000000000000000, .precision = 16 },
    };
    for (cases) |tc| {
        const result = eur30360(tc.y1, tc.m1, tc.d1, tc.y2, tc.m2, tc.d2, 0, 0, false);
        try std.testing.expect(almostEqual(result, tc.expected, pow10(tc.precision)));
    }
}

test "eur30360Model2: basic" {
    const result = eur30360Model2(2018, 12, 15, 2019, 3, 1, 0, 0, false);
    try std.testing.expect(almostEqual(result, 0.21111111, 1e-8));
}

test "eur30360Model2: time fractions" {
    const cases = [_]struct { y1: i32, m1: i32, d1: i32, y2: i32, m2: i32, d2: i32, df1: f64, df2: f64, expected: f64, tolerance: f64 }{
        .{ .y1 = 2021, .m1 = 1, .d1 = 1, .y2 = 2022, .m2 = 1, .d2 = 1, .df1 = 0.5, .df2 = 0.5, .expected = 1, .tolerance = 1e-16 },
        .{ .y1 = 2020, .m1 = 1, .d1 = 1, .y2 = 2021, .m2 = 1, .d2 = 1, .df1 = 0.5, .df2 = 0.5, .expected = 1, .tolerance = 1e-16 },
        .{ .y1 = 2020, .m1 = 1, .d1 = 1, .y2 = 2021, .m2 = 1, .d2 = 1, .df1 = 0.4, .df2 = 0.6, .expected = 1 + fd2_360, .tolerance = 1e-16 },
        .{ .y1 = 2020, .m1 = 1, .d1 = 1, .y2 = 2021, .m2 = 1, .d2 = 1, .df1 = 0.6, .df2 = 0.4, .expected = 1 - fd2_360, .tolerance = 1e-15 },
        .{ .y1 = 2020, .m1 = 1, .d1 = 1, .y2 = 2020, .m2 = 1, .d2 = 1, .df1 = 0.4, .df2 = 0.6, .expected = fd2_360, .tolerance = 1e-16 },
    };
    for (cases) |tc| {
        const result = eur30360Model2(tc.y1, tc.m1, tc.d1, tc.y2, tc.m2, tc.d2, tc.df1, tc.df2, false);
        try std.testing.expect(almostEqual(result, tc.expected, tc.tolerance));
    }
}

test "eur30360Model2: year fractions" {
    const result = eur30360Model2(2012, 1, 1, 2012, 7, 30, 0, 0, false);
    try std.testing.expect(almostEqual(result, 0.58055556, 1e-8));
}

test "eur30360Model2: Excel basis 4" {
    const cases = [_]struct { y1: i32, m1: i32, d1: i32, y2: i32, m2: i32, d2: i32, expected: f64, precision: i32 }{
        .{ .y1 = 1978, .m1 = 2, .d1 = 28, .y2 = 2020, .m2 = 5, .d2 = 17, .expected = 42.21944444444444, .precision = 13 },
        .{ .y1 = 1993, .m1 = 12, .d1 = 2, .y2 = 2022, .m2 = 4, .d2 = 18, .expected = 28.37777777777780, .precision = 13 },
        .{ .y1 = 2018, .m1 = 12, .d1 = 15, .y2 = 2019, .m2 = 3, .d2 = 1, .expected = 0.211111111111111, .precision = 13 },
        .{ .y1 = 2018, .m1 = 12, .d1 = 31, .y2 = 2019, .m2 = 1, .d2 = 1, .expected = 0.0027777777777778, .precision = 13 },
        .{ .y1 = 1994, .m1 = 6, .d1 = 30, .y2 = 1997, .m2 = 6, .d2 = 30, .expected = 3.0000000000000000, .precision = 16 },
        .{ .y1 = 1994, .m1 = 2, .d1 = 10, .y2 = 1994, .m2 = 6, .d2 = 30, .expected = 0.3888888888888889, .precision = 13 },
        .{ .y1 = 2020, .m1 = 2, .d1 = 21, .y2 = 2024, .m2 = 3, .d2 = 25, .expected = 4.0944444444444440, .precision = 13 },
        .{ .y1 = 2020, .m1 = 2, .d1 = 29, .y2 = 2021, .m2 = 2, .d2 = 28, .expected = 0.9972222222222222, .precision = 2 },
        .{ .y1 = 2020, .m1 = 1, .d1 = 31, .y2 = 2021, .m2 = 2, .d2 = 28, .expected = 1.0777777777777800, .precision = 1 },
        .{ .y1 = 2020, .m1 = 1, .d1 = 31, .y2 = 2021, .m2 = 3, .d2 = 31, .expected = 1.1666666666666667, .precision = 13 },
        .{ .y1 = 2020, .m1 = 1, .d1 = 31, .y2 = 2020, .m2 = 4, .d2 = 30, .expected = 0.2500000000000000, .precision = 16 },
        .{ .y1 = 2018, .m1 = 2, .d1 = 5, .y2 = 2023, .m2 = 5, .d2 = 14, .expected = 5.2750000000000000, .precision = 16 },
        .{ .y1 = 2020, .m1 = 2, .d1 = 29, .y2 = 2024, .m2 = 2, .d2 = 28, .expected = 3.9972222222222200, .precision = 2 },
        .{ .y1 = 2010, .m1 = 3, .d1 = 31, .y2 = 2015, .m2 = 8, .d2 = 30, .expected = 5.4166666666666700, .precision = 13 },
        .{ .y1 = 2016, .m1 = 2, .d1 = 28, .y2 = 2016, .m2 = 10, .d2 = 30, .expected = 0.6722222222222220, .precision = 13 },
        .{ .y1 = 2014, .m1 = 1, .d1 = 31, .y2 = 2014, .m2 = 8, .d2 = 31, .expected = 0.5833333333333330, .precision = 13 },
        .{ .y1 = 2014, .m1 = 2, .d1 = 28, .y2 = 2014, .m2 = 9, .d2 = 30, .expected = 0.5888888888888890, .precision = 13 },
        .{ .y1 = 2016, .m1 = 2, .d1 = 29, .y2 = 2016, .m2 = 6, .d2 = 15, .expected = 0.2944444444444440, .precision = 13 },
        .{ .y1 = 2024, .m1 = 1, .d1 = 1, .y2 = 2024, .m2 = 12, .d2 = 3, .expected = 0.9222222222222223, .precision = 13 },
        .{ .y1 = 2024, .m1 = 1, .d1 = 1, .y2 = 2025, .m2 = 1, .d2 = 2, .expected = 1.0027777777777800, .precision = 13 },
        .{ .y1 = 2024, .m1 = 1, .d1 = 1, .y2 = 2024, .m2 = 2, .d2 = 29, .expected = 0.1611111111111110, .precision = 13 },
        .{ .y1 = 2024, .m1 = 1, .d1 = 1, .y2 = 2024, .m2 = 3, .d2 = 1, .expected = 0.1666666666666670, .precision = 13 },
        .{ .y1 = 2023, .m1 = 1, .d1 = 1, .y2 = 2023, .m2 = 3, .d2 = 1, .expected = 0.1666666666666670, .precision = 13 },
        .{ .y1 = 2024, .m1 = 2, .d1 = 29, .y2 = 2025, .m2 = 2, .d2 = 28, .expected = 0.9972222222222220, .precision = 2 },
        .{ .y1 = 2024, .m1 = 1, .d1 = 1, .y2 = 2028, .m2 = 12, .d2 = 31, .expected = 4.9972222222222200, .precision = 13 },
        .{ .y1 = 2024, .m1 = 3, .d1 = 1, .y2 = 2025, .m2 = 3, .d2 = 1, .expected = 1.0000000000000000, .precision = 16 },
        .{ .y1 = 2024, .m1 = 2, .d1 = 29, .y2 = 2025, .m2 = 3, .d2 = 1, .expected = 1.0055555555555600, .precision = 13 },
        .{ .y1 = 2024, .m1 = 2, .d1 = 29, .y2 = 2028, .m2 = 2, .d2 = 28, .expected = 3.9972222222222200, .precision = 2 },
        .{ .y1 = 2024, .m1 = 2, .d1 = 29, .y2 = 2028, .m2 = 2, .d2 = 29, .expected = 4.0000000000000000, .precision = 16 },
        .{ .y1 = 2024, .m1 = 3, .d1 = 1, .y2 = 2028, .m2 = 3, .d2 = 1, .expected = 4.0000000000000000, .precision = 16 },
    };
    for (cases) |tc| {
        const result = eur30360Model2(tc.y1, tc.m1, tc.d1, tc.y2, tc.m2, tc.d2, 0, 0, false);
        try std.testing.expect(almostEqual(result, tc.expected, pow10(tc.precision)));
    }
}

test "eur30360Model3: basic" {
    const result = eur30360Model3(2018, 12, 15, 2019, 3, 1, 0, 0, false);
    try std.testing.expect(almostEqual(result, 0.21111111, 1e-8));
}

test "eur30360Model3: time fractions" {
    const cases = [_]struct { y1: i32, m1: i32, d1: i32, y2: i32, m2: i32, d2: i32, df1: f64, df2: f64, expected: f64, tolerance: f64 }{
        .{ .y1 = 2021, .m1 = 1, .d1 = 1, .y2 = 2022, .m2 = 1, .d2 = 1, .df1 = 0.5, .df2 = 0.5, .expected = 1, .tolerance = 1e-16 },
        .{ .y1 = 2020, .m1 = 1, .d1 = 1, .y2 = 2021, .m2 = 1, .d2 = 1, .df1 = 0.5, .df2 = 0.5, .expected = 1, .tolerance = 1e-16 },
        .{ .y1 = 2020, .m1 = 1, .d1 = 1, .y2 = 2021, .m2 = 1, .d2 = 1, .df1 = 0.4, .df2 = 0.6, .expected = 1 + fd2_360, .tolerance = 1e-16 },
        .{ .y1 = 2020, .m1 = 1, .d1 = 1, .y2 = 2021, .m2 = 1, .d2 = 1, .df1 = 0.6, .df2 = 0.4, .expected = 1 - fd2_360, .tolerance = 1e-15 },
        .{ .y1 = 2020, .m1 = 1, .d1 = 1, .y2 = 2020, .m2 = 1, .d2 = 1, .df1 = 0.4, .df2 = 0.6, .expected = fd2_360, .tolerance = 1e-16 },
    };
    for (cases) |tc| {
        const result = eur30360Model3(tc.y1, tc.m1, tc.d1, tc.y2, tc.m2, tc.d2, tc.df1, tc.df2, false);
        try std.testing.expect(almostEqual(result, tc.expected, tc.tolerance));
    }
}

test "eur30360Model3: year fractions" {
    const result = eur30360Model3(2012, 1, 1, 2012, 7, 30, 0, 0, false);
    try std.testing.expect(almostEqual(result, 0.58055556, 1e-8));
}

test "eur30360Model3: Excel basis 4" {
    const cases = [_]struct { y1: i32, m1: i32, d1: i32, y2: i32, m2: i32, d2: i32, expected: f64, precision: i32 }{
        .{ .y1 = 1978, .m1 = 2, .d1 = 28, .y2 = 2020, .m2 = 5, .d2 = 17, .expected = 42.21944444444444, .precision = 1 },
        .{ .y1 = 1993, .m1 = 12, .d1 = 2, .y2 = 2022, .m2 = 4, .d2 = 18, .expected = 28.37777777777780, .precision = 13 },
        .{ .y1 = 2018, .m1 = 12, .d1 = 15, .y2 = 2019, .m2 = 3, .d2 = 1, .expected = 0.211111111111111, .precision = 13 },
        .{ .y1 = 2018, .m1 = 12, .d1 = 31, .y2 = 2019, .m2 = 1, .d2 = 1, .expected = 0.0027777777777778, .precision = 13 },
        .{ .y1 = 1994, .m1 = 6, .d1 = 30, .y2 = 1997, .m2 = 6, .d2 = 30, .expected = 3.0000000000000000, .precision = 16 },
        .{ .y1 = 1994, .m1 = 2, .d1 = 10, .y2 = 1994, .m2 = 6, .d2 = 30, .expected = 0.3888888888888889, .precision = 13 },
        .{ .y1 = 2020, .m1 = 2, .d1 = 21, .y2 = 2024, .m2 = 3, .d2 = 25, .expected = 4.0944444444444440, .precision = 13 },
        .{ .y1 = 2020, .m1 = 2, .d1 = 29, .y2 = 2021, .m2 = 2, .d2 = 28, .expected = 0.9972222222222222, .precision = 2 },
        .{ .y1 = 2020, .m1 = 1, .d1 = 31, .y2 = 2021, .m2 = 2, .d2 = 28, .expected = 1.0777777777777800, .precision = 1 },
        .{ .y1 = 2020, .m1 = 1, .d1 = 31, .y2 = 2021, .m2 = 3, .d2 = 31, .expected = 1.1666666666666667, .precision = 13 },
        .{ .y1 = 2020, .m1 = 1, .d1 = 31, .y2 = 2020, .m2 = 4, .d2 = 30, .expected = 0.2500000000000000, .precision = 16 },
        .{ .y1 = 2018, .m1 = 2, .d1 = 5, .y2 = 2023, .m2 = 5, .d2 = 14, .expected = 5.2750000000000000, .precision = 16 },
        .{ .y1 = 2020, .m1 = 2, .d1 = 29, .y2 = 2024, .m2 = 2, .d2 = 28, .expected = 3.9972222222222200, .precision = 2 },
        .{ .y1 = 2010, .m1 = 3, .d1 = 31, .y2 = 2015, .m2 = 8, .d2 = 30, .expected = 5.4166666666666700, .precision = 13 },
        .{ .y1 = 2016, .m1 = 2, .d1 = 28, .y2 = 2016, .m2 = 10, .d2 = 30, .expected = 0.6722222222222220, .precision = 1 },
        .{ .y1 = 2014, .m1 = 1, .d1 = 31, .y2 = 2014, .m2 = 8, .d2 = 31, .expected = 0.5833333333333330, .precision = 13 },
        .{ .y1 = 2014, .m1 = 2, .d1 = 28, .y2 = 2014, .m2 = 9, .d2 = 30, .expected = 0.5888888888888890, .precision = 1 },
        .{ .y1 = 2016, .m1 = 2, .d1 = 29, .y2 = 2016, .m2 = 6, .d2 = 15, .expected = 0.2944444444444440, .precision = 2 },
        .{ .y1 = 2024, .m1 = 1, .d1 = 1, .y2 = 2024, .m2 = 12, .d2 = 31, .expected = 0.9972222222222220, .precision = 13 },
        .{ .y1 = 2024, .m1 = 1, .d1 = 1, .y2 = 2025, .m2 = 1, .d2 = 2, .expected = 1.0027777777777800, .precision = 13 },
        .{ .y1 = 2024, .m1 = 1, .d1 = 1, .y2 = 2024, .m2 = 2, .d2 = 29, .expected = 0.1611111111111110, .precision = 2 },
        .{ .y1 = 2024, .m1 = 1, .d1 = 1, .y2 = 2024, .m2 = 3, .d2 = 1, .expected = 0.1666666666666670, .precision = 13 },
        .{ .y1 = 2023, .m1 = 1, .d1 = 1, .y2 = 2023, .m2 = 3, .d2 = 1, .expected = 0.1666666666666670, .precision = 13 },
        .{ .y1 = 2024, .m1 = 2, .d1 = 29, .y2 = 2025, .m2 = 2, .d2 = 28, .expected = 0.9972222222222220, .precision = 2 },
        .{ .y1 = 2024, .m1 = 1, .d1 = 1, .y2 = 2028, .m2 = 12, .d2 = 31, .expected = 4.9972222222222200, .precision = 13 },
        .{ .y1 = 2024, .m1 = 3, .d1 = 1, .y2 = 2025, .m2 = 3, .d2 = 1, .expected = 1.0000000000000000, .precision = 16 },
        .{ .y1 = 2024, .m1 = 2, .d1 = 29, .y2 = 2025, .m2 = 3, .d2 = 1, .expected = 1.0055555555555600, .precision = 2 },
        .{ .y1 = 2024, .m1 = 2, .d1 = 29, .y2 = 2028, .m2 = 2, .d2 = 28, .expected = 3.9972222222222200, .precision = 2 },
        .{ .y1 = 2024, .m1 = 2, .d1 = 29, .y2 = 2028, .m2 = 2, .d2 = 29, .expected = 4.0000000000000000, .precision = 16 },
        .{ .y1 = 2024, .m1 = 3, .d1 = 1, .y2 = 2028, .m2 = 3, .d2 = 1, .expected = 4.0000000000000000, .precision = 16 },
    };
    for (cases) |tc| {
        const result = eur30360Model3(tc.y1, tc.m1, tc.d1, tc.y2, tc.m2, tc.d2, 0, 0, false);
        try std.testing.expect(almostEqual(result, tc.expected, pow10(tc.precision)));
    }
}

test "eur30360Plus: basic" {
    const result = eur30360Plus(2018, 12, 15, 2019, 3, 1, 0, 0, false);
    try std.testing.expect(almostEqual(result, 0.21111111, 1e-8));
}

test "eur30360Plus: time fractions" {
    const cases = [_]struct { y1: i32, m1: i32, d1: i32, y2: i32, m2: i32, d2: i32, df1: f64, df2: f64, expected: f64, tolerance: f64 }{
        .{ .y1 = 2021, .m1 = 1, .d1 = 1, .y2 = 2022, .m2 = 1, .d2 = 1, .df1 = 0.5, .df2 = 0.5, .expected = 1, .tolerance = 1e-16 },
        .{ .y1 = 2020, .m1 = 1, .d1 = 1, .y2 = 2021, .m2 = 1, .d2 = 1, .df1 = 0.5, .df2 = 0.5, .expected = 1, .tolerance = 1e-16 },
        .{ .y1 = 2020, .m1 = 1, .d1 = 1, .y2 = 2021, .m2 = 1, .d2 = 1, .df1 = 0.4, .df2 = 0.6, .expected = 1 + fd2_360, .tolerance = 1e-16 },
        .{ .y1 = 2020, .m1 = 1, .d1 = 1, .y2 = 2021, .m2 = 1, .d2 = 1, .df1 = 0.6, .df2 = 0.4, .expected = 1 - fd2_360, .tolerance = 1e-15 },
        .{ .y1 = 2020, .m1 = 1, .d1 = 1, .y2 = 2020, .m2 = 1, .d2 = 1, .df1 = 0.4, .df2 = 0.6, .expected = fd2_360, .tolerance = 1e-16 },
    };
    for (cases) |tc| {
        const result = eur30360Plus(tc.y1, tc.m1, tc.d1, tc.y2, tc.m2, tc.d2, tc.df1, tc.df2, false);
        try std.testing.expect(almostEqual(result, tc.expected, tc.tolerance));
    }
}

test "eur30360Plus: year fractions" {
    const result = eur30360Plus(2012, 1, 1, 2012, 7, 30, 0, 0, false);
    try std.testing.expect(almostEqual(result, 0.58055556, 1e-8));
}

test "eur30360Plus: Excel basis 4" {
    const cases = [_]struct { y1: i32, m1: i32, d1: i32, y2: i32, m2: i32, d2: i32, expected: f64, precision: i32 }{
        .{ .y1 = 1978, .m1 = 2, .d1 = 28, .y2 = 2020, .m2 = 5, .d2 = 17, .expected = 42.21944444444444, .precision = 13 },
        .{ .y1 = 1993, .m1 = 12, .d1 = 2, .y2 = 2022, .m2 = 4, .d2 = 18, .expected = 28.37777777777780, .precision = 13 },
        .{ .y1 = 2018, .m1 = 12, .d1 = 15, .y2 = 2019, .m2 = 3, .d2 = 1, .expected = 0.211111111111111, .precision = 13 },
        .{ .y1 = 2018, .m1 = 12, .d1 = 31, .y2 = 2019, .m2 = 1, .d2 = 1, .expected = 0.0027777777777778, .precision = 13 },
        .{ .y1 = 1994, .m1 = 6, .d1 = 30, .y2 = 1997, .m2 = 6, .d2 = 30, .expected = 3.0000000000000000, .precision = 16 },
        .{ .y1 = 1994, .m1 = 2, .d1 = 10, .y2 = 1994, .m2 = 6, .d2 = 30, .expected = 0.3888888888888889, .precision = 13 },
        .{ .y1 = 2020, .m1 = 2, .d1 = 21, .y2 = 2024, .m2 = 3, .d2 = 25, .expected = 4.0944444444444440, .precision = 13 },
        .{ .y1 = 2020, .m1 = 2, .d1 = 29, .y2 = 2021, .m2 = 2, .d2 = 28, .expected = 0.9972222222222222, .precision = 13 },
        .{ .y1 = 2020, .m1 = 1, .d1 = 31, .y2 = 2021, .m2 = 2, .d2 = 28, .expected = 1.0777777777777800, .precision = 13 },
        .{ .y1 = 2020, .m1 = 1, .d1 = 31, .y2 = 2021, .m2 = 3, .d2 = 31, .expected = 1.1666666666666667, .precision = 1 },
        .{ .y1 = 2020, .m1 = 1, .d1 = 31, .y2 = 2020, .m2 = 4, .d2 = 30, .expected = 0.2500000000000000, .precision = 16 },
        .{ .y1 = 2018, .m1 = 2, .d1 = 5, .y2 = 2023, .m2 = 5, .d2 = 14, .expected = 5.2750000000000000, .precision = 16 },
        .{ .y1 = 2020, .m1 = 2, .d1 = 29, .y2 = 2024, .m2 = 2, .d2 = 28, .expected = 3.9972222222222200, .precision = 13 },
        .{ .y1 = 2010, .m1 = 3, .d1 = 31, .y2 = 2015, .m2 = 8, .d2 = 30, .expected = 5.4166666666666700, .precision = 13 },
        .{ .y1 = 2016, .m1 = 2, .d1 = 28, .y2 = 2016, .m2 = 10, .d2 = 30, .expected = 0.6722222222222220, .precision = 13 },
        .{ .y1 = 2014, .m1 = 1, .d1 = 31, .y2 = 2014, .m2 = 8, .d2 = 31, .expected = 0.5833333333333330, .precision = 1 },
        .{ .y1 = 2014, .m1 = 2, .d1 = 28, .y2 = 2014, .m2 = 9, .d2 = 30, .expected = 0.5888888888888890, .precision = 13 },
        .{ .y1 = 2016, .m1 = 2, .d1 = 29, .y2 = 2016, .m2 = 6, .d2 = 15, .expected = 0.2944444444444440, .precision = 13 },
        .{ .y1 = 2024, .m1 = 1, .d1 = 1, .y2 = 2024, .m2 = 12, .d2 = 31, .expected = 0.9972222222222220, .precision = 1 },
        .{ .y1 = 2024, .m1 = 1, .d1 = 1, .y2 = 2025, .m2 = 1, .d2 = 2, .expected = 1.0027777777777800, .precision = 13 },
        .{ .y1 = 2024, .m1 = 1, .d1 = 1, .y2 = 2024, .m2 = 2, .d2 = 29, .expected = 0.1611111111111110, .precision = 13 },
        .{ .y1 = 2024, .m1 = 1, .d1 = 1, .y2 = 2024, .m2 = 3, .d2 = 1, .expected = 0.1666666666666670, .precision = 13 },
        .{ .y1 = 2023, .m1 = 1, .d1 = 1, .y2 = 2023, .m2 = 3, .d2 = 1, .expected = 0.1666666666666670, .precision = 13 },
        .{ .y1 = 2024, .m1 = 2, .d1 = 29, .y2 = 2025, .m2 = 2, .d2 = 28, .expected = 0.9972222222222220, .precision = 13 },
        .{ .y1 = 2024, .m1 = 1, .d1 = 1, .y2 = 2028, .m2 = 12, .d2 = 31, .expected = 4.9972222222222200, .precision = 1 },
        .{ .y1 = 2024, .m1 = 3, .d1 = 1, .y2 = 2025, .m2 = 3, .d2 = 1, .expected = 1.0000000000000000, .precision = 16 },
        .{ .y1 = 2024, .m1 = 2, .d1 = 29, .y2 = 2025, .m2 = 3, .d2 = 1, .expected = 1.0055555555555600, .precision = 13 },
        .{ .y1 = 2024, .m1 = 2, .d1 = 29, .y2 = 2028, .m2 = 2, .d2 = 28, .expected = 3.9972222222222200, .precision = 13 },
        .{ .y1 = 2024, .m1 = 2, .d1 = 29, .y2 = 2028, .m2 = 2, .d2 = 29, .expected = 4.0000000000000000, .precision = 16 },
        .{ .y1 = 2024, .m1 = 3, .d1 = 1, .y2 = 2028, .m2 = 3, .d2 = 1, .expected = 4.0000000000000000, .precision = 16 },
    };
    for (cases) |tc| {
        const result = eur30360Plus(tc.y1, tc.m1, tc.d1, tc.y2, tc.m2, tc.d2, 0, 0, false);
        try std.testing.expect(almostEqual(result, tc.expected, pow10(tc.precision)));
    }
}

test "us30360 basic" {
    const result = us30360(2018, 12, 15, 2019, 3, 1, 0, 0, false);
    try std.testing.expect(almostEqual(result, 0.21111111, 1e-8));
}

test "us30360 time fractions" {
    const cases = [_]struct { y1: i32, m1: i32, d1: i32, y2: i32, m2: i32, d2: i32, df1: f64, df2: f64, expected: f64, tolerance: f64 }{
        .{ .y1 = 2021, .m1 = 1, .d1 = 1, .y2 = 2022, .m2 = 1, .d2 = 1, .df1 = 0.5, .df2 = 0.5, .expected = 1, .tolerance = 1e-16 },
        .{ .y1 = 2020, .m1 = 1, .d1 = 1, .y2 = 2021, .m2 = 1, .d2 = 1, .df1 = 0.5, .df2 = 0.5, .expected = 1, .tolerance = 1e-16 },
        .{ .y1 = 2020, .m1 = 1, .d1 = 1, .y2 = 2021, .m2 = 1, .d2 = 1, .df1 = 0.4, .df2 = 0.6, .expected = 1 + fd2_360, .tolerance = 1e-16 },
        .{ .y1 = 2020, .m1 = 1, .d1 = 1, .y2 = 2021, .m2 = 1, .d2 = 1, .df1 = 0.6, .df2 = 0.4, .expected = 1 - fd2_360, .tolerance = 1e-15 },
        .{ .y1 = 2020, .m1 = 1, .d1 = 1, .y2 = 2020, .m2 = 1, .d2 = 1, .df1 = 0.4, .df2 = 0.6, .expected = fd2_360, .tolerance = 1e-16 },
    };
    for (cases) |tc| {
        const result = us30360(tc.y1, tc.m1, tc.d1, tc.y2, tc.m2, tc.d2, tc.df1, tc.df2, false);
        try std.testing.expect(almostEqual(result, tc.expected, tc.tolerance));
    }
}

test "us30360 YEARFRAC basis 0" {
    const result = us30360(2012, 1, 1, 2012, 7, 30, 0, 0, false);
    try std.testing.expect(almostEqual(result, 0.58055556, 1e-8));
}

test "us30360 Excel basis 0" {
    const cases = [_]struct { y1: i32, m1: i32, d1: i32, y2: i32, m2: i32, d2: i32, expected: f64, precision: i32 }{
        .{ .y1 = 1978, .m1 = 2, .d1 = 28, .y2 = 2020, .m2 = 5, .d2 = 17, .expected = 42.2138888888889000, .precision = 1 },
        .{ .y1 = 1993, .m1 = 12, .d1 = 2, .y2 = 2022, .m2 = 4, .d2 = 18, .expected = 28.3777777777778000, .precision = 13 },
        .{ .y1 = 2018, .m1 = 12, .d1 = 15, .y2 = 2019, .m2 = 3, .d2 = 1, .expected = 0.2111111111111110, .precision = 13 },
        .{ .y1 = 2018, .m1 = 12, .d1 = 31, .y2 = 2019, .m2 = 1, .d2 = 1, .expected = 0.0027777777777778, .precision = 13 },
        .{ .y1 = 1994, .m1 = 6, .d1 = 30, .y2 = 1997, .m2 = 6, .d2 = 30, .expected = 3.0000000000000000, .precision = 13 },
        .{ .y1 = 1994, .m1 = 2, .d1 = 10, .y2 = 1994, .m2 = 6, .d2 = 30, .expected = 0.3888888888888890, .precision = 13 },
        .{ .y1 = 2020, .m1 = 2, .d1 = 21, .y2 = 2024, .m2 = 3, .d2 = 25, .expected = 4.0944444444444400, .precision = 13 },
        .{ .y1 = 2020, .m1 = 2, .d1 = 29, .y2 = 2021, .m2 = 2, .d2 = 28, .expected = 1.0000000000000000, .precision = 2 },
        .{ .y1 = 2020, .m1 = 1, .d1 = 31, .y2 = 2021, .m2 = 2, .d2 = 28, .expected = 1.0777777777777800, .precision = 13 },
        .{ .y1 = 2020, .m1 = 1, .d1 = 31, .y2 = 2021, .m2 = 3, .d2 = 31, .expected = 1.1666666666666700, .precision = 13 },
        .{ .y1 = 2020, .m1 = 1, .d1 = 31, .y2 = 2020, .m2 = 4, .d2 = 30, .expected = 0.2500000000000000, .precision = 13 },
        .{ .y1 = 2018, .m1 = 2, .d1 = 5, .y2 = 2023, .m2 = 5, .d2 = 14, .expected = 5.2750000000000000, .precision = 13 },
        .{ .y1 = 2020, .m1 = 2, .d1 = 29, .y2 = 2024, .m2 = 2, .d2 = 28, .expected = 3.9944444444444400, .precision = 2 },
        .{ .y1 = 2010, .m1 = 3, .d1 = 31, .y2 = 2015, .m2 = 8, .d2 = 30, .expected = 5.4166666666666700, .precision = 13 },
        .{ .y1 = 2016, .m1 = 2, .d1 = 28, .y2 = 2016, .m2 = 10, .d2 = 30, .expected = 0.6722222222222220, .precision = 13 },
        .{ .y1 = 2014, .m1 = 1, .d1 = 31, .y2 = 2014, .m2 = 8, .d2 = 31, .expected = 0.5833333333333330, .precision = 13 },
        .{ .y1 = 2014, .m1 = 2, .d1 = 28, .y2 = 2014, .m2 = 9, .d2 = 30, .expected = 0.5833333333333330, .precision = 1 },
        .{ .y1 = 2016, .m1 = 2, .d1 = 29, .y2 = 2016, .m2 = 6, .d2 = 15, .expected = 0.2916666666666670, .precision = 2 },
        .{ .y1 = 2024, .m1 = 1, .d1 = 1, .y2 = 2024, .m2 = 12, .d2 = 31, .expected = 1.0000000000000000, .precision = 16 },
        .{ .y1 = 2024, .m1 = 1, .d1 = 1, .y2 = 2025, .m2 = 1, .d2 = 2, .expected = 1.0027777777777800, .precision = 13 },
        .{ .y1 = 2024, .m1 = 1, .d1 = 1, .y2 = 2024, .m2 = 2, .d2 = 29, .expected = 0.1611111111111110, .precision = 13 },
        .{ .y1 = 2024, .m1 = 1, .d1 = 1, .y2 = 2024, .m2 = 3, .d2 = 1, .expected = 0.1666666666666670, .precision = 13 },
        .{ .y1 = 2023, .m1 = 1, .d1 = 1, .y2 = 2023, .m2 = 3, .d2 = 1, .expected = 0.1666666666666670, .precision = 13 },
        .{ .y1 = 2024, .m1 = 2, .d1 = 29, .y2 = 2025, .m2 = 2, .d2 = 28, .expected = 1.0000000000000000, .precision = 2 },
        .{ .y1 = 2024, .m1 = 1, .d1 = 1, .y2 = 2028, .m2 = 12, .d2 = 31, .expected = 5.0000000000000000, .precision = 13 },
        .{ .y1 = 2024, .m1 = 3, .d1 = 1, .y2 = 2025, .m2 = 3, .d2 = 1, .expected = 1.0000000000000000, .precision = 16 },
        .{ .y1 = 2024, .m1 = 2, .d1 = 29, .y2 = 2025, .m2 = 3, .d2 = 1, .expected = 1.0027777777777800, .precision = 2 },
        .{ .y1 = 2024, .m1 = 2, .d1 = 29, .y2 = 2028, .m2 = 2, .d2 = 28, .expected = 3.9944444444444400, .precision = 2 },
        .{ .y1 = 2024, .m1 = 2, .d1 = 29, .y2 = 2028, .m2 = 2, .d2 = 29, .expected = 4.0000000000000000, .precision = 13 },
        .{ .y1 = 2024, .m1 = 3, .d1 = 1, .y2 = 2028, .m2 = 3, .d2 = 1, .expected = 4.0000000000000000, .precision = 13 },
    };
    for (cases) |tc| {
        const result = us30360(tc.y1, tc.m1, tc.d1, tc.y2, tc.m2, tc.d2, 0, 0, false);
        try std.testing.expect(almostEqual(result, tc.expected, pow10(tc.precision)));
    }
}

test "us30360Eom basic" {
    const result = us30360Eom(2018, 12, 15, 2019, 3, 1, 0, 0, false);
    try std.testing.expect(almostEqual(result, 0.21111111, 1e-8));
}

test "us30360Eom time fractions" {
    const cases = [_]struct { y1: i32, m1: i32, d1: i32, y2: i32, m2: i32, d2: i32, df1: f64, df2: f64, expected: f64, tolerance: f64 }{
        .{ .y1 = 2021, .m1 = 1, .d1 = 1, .y2 = 2022, .m2 = 1, .d2 = 1, .df1 = 0.5, .df2 = 0.5, .expected = 1, .tolerance = 1e-16 },
        .{ .y1 = 2020, .m1 = 1, .d1 = 1, .y2 = 2021, .m2 = 1, .d2 = 1, .df1 = 0.5, .df2 = 0.5, .expected = 1, .tolerance = 1e-16 },
        .{ .y1 = 2020, .m1 = 1, .d1 = 1, .y2 = 2021, .m2 = 1, .d2 = 1, .df1 = 0.4, .df2 = 0.6, .expected = 1 + fd2_360, .tolerance = 1e-16 },
        .{ .y1 = 2020, .m1 = 1, .d1 = 1, .y2 = 2021, .m2 = 1, .d2 = 1, .df1 = 0.6, .df2 = 0.4, .expected = 1 - fd2_360, .tolerance = 1e-15 },
        .{ .y1 = 2020, .m1 = 1, .d1 = 1, .y2 = 2020, .m2 = 1, .d2 = 1, .df1 = 0.4, .df2 = 0.6, .expected = fd2_360, .tolerance = 1e-16 },
    };
    for (cases) |tc| {
        const result = us30360Eom(tc.y1, tc.m1, tc.d1, tc.y2, tc.m2, tc.d2, tc.df1, tc.df2, false);
        try std.testing.expect(almostEqual(result, tc.expected, tc.tolerance));
    }
}

test "us30360Eom YEARFRAC basis 0" {
    const result = us30360Eom(2012, 1, 1, 2012, 7, 30, 0, 0, false);
    try std.testing.expect(almostEqual(result, 0.58055556, 1e-8));
}

test "us30360Eom Excel basis 0" {
    const cases = [_]struct { y1: i32, m1: i32, d1: i32, y2: i32, m2: i32, d2: i32, expected: f64, precision: i32 }{
        .{ .y1 = 1978, .m1 = 2, .d1 = 28, .y2 = 2020, .m2 = 5, .d2 = 17, .expected = 42.2138888888889000, .precision = 13 },
        .{ .y1 = 1993, .m1 = 12, .d1 = 2, .y2 = 2022, .m2 = 4, .d2 = 18, .expected = 28.3777777777778000, .precision = 13 },
        .{ .y1 = 2018, .m1 = 12, .d1 = 15, .y2 = 2019, .m2 = 3, .d2 = 1, .expected = 0.2111111111111110, .precision = 13 },
        .{ .y1 = 2018, .m1 = 12, .d1 = 31, .y2 = 2019, .m2 = 1, .d2 = 1, .expected = 0.0027777777777778, .precision = 13 },
        .{ .y1 = 1994, .m1 = 6, .d1 = 30, .y2 = 1997, .m2 = 6, .d2 = 30, .expected = 3.0000000000000000, .precision = 13 },
        .{ .y1 = 1994, .m1 = 2, .d1 = 10, .y2 = 1994, .m2 = 6, .d2 = 30, .expected = 0.3888888888888890, .precision = 13 },
        .{ .y1 = 2020, .m1 = 2, .d1 = 21, .y2 = 2024, .m2 = 3, .d2 = 25, .expected = 4.0944444444444400, .precision = 13 },
        .{ .y1 = 2020, .m1 = 2, .d1 = 29, .y2 = 2021, .m2 = 2, .d2 = 28, .expected = 1.0000000000000000, .precision = 16 },
        .{ .y1 = 2020, .m1 = 1, .d1 = 31, .y2 = 2021, .m2 = 2, .d2 = 28, .expected = 1.0777777777777800, .precision = 13 },
        .{ .y1 = 2020, .m1 = 1, .d1 = 31, .y2 = 2021, .m2 = 3, .d2 = 31, .expected = 1.1666666666666700, .precision = 13 },
        .{ .y1 = 2020, .m1 = 1, .d1 = 31, .y2 = 2020, .m2 = 4, .d2 = 30, .expected = 0.2500000000000000, .precision = 13 },
        .{ .y1 = 2018, .m1 = 2, .d1 = 5, .y2 = 2023, .m2 = 5, .d2 = 14, .expected = 5.2750000000000000, .precision = 13 },
        .{ .y1 = 2020, .m1 = 2, .d1 = 29, .y2 = 2024, .m2 = 2, .d2 = 28, .expected = 3.9944444444444400, .precision = 1 },
        .{ .y1 = 2010, .m1 = 3, .d1 = 31, .y2 = 2015, .m2 = 8, .d2 = 30, .expected = 5.4166666666666700, .precision = 13 },
        .{ .y1 = 2016, .m1 = 2, .d1 = 28, .y2 = 2016, .m2 = 10, .d2 = 30, .expected = 0.6722222222222220, .precision = 1 },
        .{ .y1 = 2014, .m1 = 1, .d1 = 31, .y2 = 2014, .m2 = 8, .d2 = 31, .expected = 0.5833333333333330, .precision = 13 },
        .{ .y1 = 2014, .m1 = 2, .d1 = 28, .y2 = 2014, .m2 = 9, .d2 = 30, .expected = 0.5833333333333330, .precision = 13 },
        .{ .y1 = 2016, .m1 = 2, .d1 = 29, .y2 = 2016, .m2 = 6, .d2 = 15, .expected = 0.2916666666666670, .precision = 13 },
        .{ .y1 = 2024, .m1 = 1, .d1 = 1, .y2 = 2024, .m2 = 12, .d2 = 31, .expected = 1.0000000000000000, .precision = 16 },
        .{ .y1 = 2024, .m1 = 1, .d1 = 1, .y2 = 2025, .m2 = 1, .d2 = 2, .expected = 1.0027777777777800, .precision = 13 },
        .{ .y1 = 2024, .m1 = 1, .d1 = 1, .y2 = 2024, .m2 = 2, .d2 = 29, .expected = 0.1611111111111110, .precision = 13 },
        .{ .y1 = 2024, .m1 = 1, .d1 = 1, .y2 = 2024, .m2 = 3, .d2 = 1, .expected = 0.1666666666666670, .precision = 13 },
        .{ .y1 = 2023, .m1 = 1, .d1 = 1, .y2 = 2023, .m2 = 3, .d2 = 1, .expected = 0.1666666666666670, .precision = 13 },
        .{ .y1 = 2024, .m1 = 2, .d1 = 29, .y2 = 2025, .m2 = 2, .d2 = 28, .expected = 1.0000000000000000, .precision = 16 },
        .{ .y1 = 2024, .m1 = 1, .d1 = 1, .y2 = 2028, .m2 = 12, .d2 = 31, .expected = 5.0000000000000000, .precision = 13 },
        .{ .y1 = 2024, .m1 = 3, .d1 = 1, .y2 = 2025, .m2 = 3, .d2 = 1, .expected = 1.0000000000000000, .precision = 16 },
        .{ .y1 = 2024, .m1 = 2, .d1 = 29, .y2 = 2025, .m2 = 3, .d2 = 1, .expected = 1.0027777777777800, .precision = 13 },
        .{ .y1 = 2024, .m1 = 2, .d1 = 29, .y2 = 2028, .m2 = 2, .d2 = 28, .expected = 3.9944444444444400, .precision = 1 },
        .{ .y1 = 2024, .m1 = 2, .d1 = 29, .y2 = 2028, .m2 = 2, .d2 = 29, .expected = 4.0000000000000000, .precision = 13 },
        .{ .y1 = 2024, .m1 = 3, .d1 = 1, .y2 = 2028, .m2 = 3, .d2 = 1, .expected = 4.0000000000000000, .precision = 13 },
    };
    for (cases) |tc| {
        const result = us30360Eom(tc.y1, tc.m1, tc.d1, tc.y2, tc.m2, tc.d2, 0, 0, false);
        try std.testing.expect(almostEqual(result, tc.expected, pow10(tc.precision)));
    }
}

test "us30360Nasd basic" {
    const result = us30360Nasd(2018, 12, 15, 2019, 3, 1, 0, 0, false);
    try std.testing.expect(almostEqual(result, 0.21111111, 1e-8));
}

test "us30360Nasd time fractions" {
    const cases = [_]struct { y1: i32, m1: i32, d1: i32, y2: i32, m2: i32, d2: i32, df1: f64, df2: f64, expected: f64, tolerance: f64 }{
        .{ .y1 = 2021, .m1 = 1, .d1 = 1, .y2 = 2022, .m2 = 1, .d2 = 1, .df1 = 0.5, .df2 = 0.5, .expected = 1, .tolerance = 1e-16 },
        .{ .y1 = 2020, .m1 = 1, .d1 = 1, .y2 = 2021, .m2 = 1, .d2 = 1, .df1 = 0.5, .df2 = 0.5, .expected = 1, .tolerance = 1e-16 },
        .{ .y1 = 2020, .m1 = 1, .d1 = 1, .y2 = 2021, .m2 = 1, .d2 = 1, .df1 = 0.4, .df2 = 0.6, .expected = 1 + fd2_360, .tolerance = 1e-16 },
        .{ .y1 = 2020, .m1 = 1, .d1 = 1, .y2 = 2021, .m2 = 1, .d2 = 1, .df1 = 0.6, .df2 = 0.4, .expected = 1 - fd2_360, .tolerance = 1e-15 },
        .{ .y1 = 2020, .m1 = 1, .d1 = 1, .y2 = 2020, .m2 = 1, .d2 = 1, .df1 = 0.4, .df2 = 0.6, .expected = fd2_360, .tolerance = 1e-16 },
    };
    for (cases) |tc| {
        const result = us30360Nasd(tc.y1, tc.m1, tc.d1, tc.y2, tc.m2, tc.d2, tc.df1, tc.df2, false);
        try std.testing.expect(almostEqual(result, tc.expected, tc.tolerance));
    }
}

test "us30360Nasd YEARFRAC basis 0" {
    const result = us30360Nasd(2012, 1, 1, 2012, 7, 30, 0, 0, false);
    try std.testing.expect(almostEqual(result, 0.58055556, 1e-8));
}

test "us30360Nasd Excel basis 0" {
    const cases = [_]struct { y1: i32, m1: i32, d1: i32, y2: i32, m2: i32, d2: i32, expected: f64, precision: i32 }{
        .{ .y1 = 1978, .m1 = 2, .d1 = 28, .y2 = 2020, .m2 = 5, .d2 = 17, .expected = 42.2138888888889000, .precision = 1 },
        .{ .y1 = 1993, .m1 = 12, .d1 = 2, .y2 = 2022, .m2 = 4, .d2 = 18, .expected = 28.3777777777778000, .precision = 13 },
        .{ .y1 = 2018, .m1 = 12, .d1 = 15, .y2 = 2019, .m2 = 3, .d2 = 1, .expected = 0.2111111111111110, .precision = 13 },
        .{ .y1 = 2018, .m1 = 12, .d1 = 31, .y2 = 2019, .m2 = 1, .d2 = 1, .expected = 0.0027777777777778, .precision = 13 },
        .{ .y1 = 1994, .m1 = 6, .d1 = 30, .y2 = 1997, .m2 = 6, .d2 = 30, .expected = 3.0000000000000000, .precision = 13 },
        .{ .y1 = 1994, .m1 = 2, .d1 = 10, .y2 = 1994, .m2 = 6, .d2 = 30, .expected = 0.3888888888888890, .precision = 13 },
        .{ .y1 = 2020, .m1 = 2, .d1 = 21, .y2 = 2024, .m2 = 3, .d2 = 25, .expected = 4.0944444444444400, .precision = 13 },
        .{ .y1 = 2020, .m1 = 2, .d1 = 29, .y2 = 2021, .m2 = 2, .d2 = 28, .expected = 1.0000000000000000, .precision = 2 },
        .{ .y1 = 2020, .m1 = 1, .d1 = 31, .y2 = 2021, .m2 = 2, .d2 = 28, .expected = 1.0777777777777800, .precision = 13 },
        .{ .y1 = 2020, .m1 = 1, .d1 = 31, .y2 = 2021, .m2 = 3, .d2 = 31, .expected = 1.1666666666666700, .precision = 13 },
        .{ .y1 = 2020, .m1 = 1, .d1 = 31, .y2 = 2020, .m2 = 4, .d2 = 30, .expected = 0.2500000000000000, .precision = 13 },
        .{ .y1 = 2018, .m1 = 2, .d1 = 5, .y2 = 2023, .m2 = 5, .d2 = 14, .expected = 5.2750000000000000, .precision = 13 },
        .{ .y1 = 2020, .m1 = 2, .d1 = 29, .y2 = 2024, .m2 = 2, .d2 = 28, .expected = 3.9944444444444400, .precision = 2 },
        .{ .y1 = 2010, .m1 = 3, .d1 = 31, .y2 = 2015, .m2 = 8, .d2 = 30, .expected = 5.4166666666666700, .precision = 13 },
        .{ .y1 = 2016, .m1 = 2, .d1 = 28, .y2 = 2016, .m2 = 10, .d2 = 30, .expected = 0.6722222222222220, .precision = 13 },
        .{ .y1 = 2014, .m1 = 1, .d1 = 31, .y2 = 2014, .m2 = 8, .d2 = 31, .expected = 0.5833333333333330, .precision = 13 },
        .{ .y1 = 2014, .m1 = 2, .d1 = 28, .y2 = 2014, .m2 = 9, .d2 = 30, .expected = 0.5888888888888889, .precision = 13 },
        .{ .y1 = 2016, .m1 = 2, .d1 = 29, .y2 = 2016, .m2 = 6, .d2 = 15, .expected = 0.2916666666666670, .precision = 2 },
        .{ .y1 = 2024, .m1 = 1, .d1 = 1, .y2 = 2024, .m2 = 12, .d2 = 31, .expected = 1.0000000000000000, .precision = 2 },
        .{ .y1 = 2024, .m1 = 1, .d1 = 1, .y2 = 2025, .m2 = 1, .d2 = 2, .expected = 1.0027777777777800, .precision = 13 },
        .{ .y1 = 2024, .m1 = 1, .d1 = 1, .y2 = 2024, .m2 = 2, .d2 = 29, .expected = 0.1611111111111110, .precision = 13 },
        .{ .y1 = 2024, .m1 = 1, .d1 = 1, .y2 = 2024, .m2 = 3, .d2 = 1, .expected = 0.1666666666666670, .precision = 13 },
        .{ .y1 = 2023, .m1 = 1, .d1 = 1, .y2 = 2023, .m2 = 3, .d2 = 1, .expected = 0.1666666666666670, .precision = 13 },
        .{ .y1 = 2024, .m1 = 2, .d1 = 29, .y2 = 2025, .m2 = 2, .d2 = 28, .expected = 1.0000000000000000, .precision = 2 },
        .{ .y1 = 2024, .m1 = 1, .d1 = 1, .y2 = 2028, .m2 = 12, .d2 = 31, .expected = 5.0000000000000000, .precision = 2 },
        .{ .y1 = 2024, .m1 = 3, .d1 = 1, .y2 = 2025, .m2 = 3, .d2 = 1, .expected = 1.0000000000000000, .precision = 16 },
        .{ .y1 = 2024, .m1 = 2, .d1 = 29, .y2 = 2025, .m2 = 3, .d2 = 1, .expected = 1.0027777777777800, .precision = 2 },
        .{ .y1 = 2024, .m1 = 2, .d1 = 29, .y2 = 2028, .m2 = 2, .d2 = 28, .expected = 3.9944444444444400, .precision = 2 },
        .{ .y1 = 2024, .m1 = 2, .d1 = 29, .y2 = 2028, .m2 = 2, .d2 = 29, .expected = 4.0000000000000000, .precision = 13 },
        .{ .y1 = 2024, .m1 = 3, .d1 = 1, .y2 = 2028, .m2 = 3, .d2 = 1, .expected = 4.0000000000000000, .precision = 13 },
    };
    for (cases) |tc| {
        const result = us30360Nasd(tc.y1, tc.m1, tc.d1, tc.y2, tc.m2, tc.d2, 0, 0, false);
        try std.testing.expect(almostEqual(result, tc.expected, pow10(tc.precision)));
    }
}

test "thirty365 basic" {
    const result = thirty365(2018, 12, 15, 2019, 3, 1, 0, 0, false);
    try std.testing.expect(almostEqual(result, 0.20821918, 1e-8));
}

test "thirty365 time fractions" {
    const cases = [_]struct { y1: i32, m1: i32, d1: i32, y2: i32, m2: i32, d2: i32, df1: f64, df2: f64, expected: f64, tolerance: f64 }{
        .{ .y1 = 2021, .m1 = 1, .d1 = 1, .y2 = 2022, .m2 = 1, .d2 = 1, .df1 = 0.5, .df2 = 0.5, .expected = 0.986301369863, .tolerance = 1e-13 },
        .{ .y1 = 2020, .m1 = 1, .d1 = 1, .y2 = 2021, .m2 = 1, .d2 = 1, .df1 = 0.5, .df2 = 0.5, .expected = 0.986301369863, .tolerance = 1e-13 },
        .{ .y1 = 2020, .m1 = 1, .d1 = 1, .y2 = 2021, .m2 = 1, .d2 = 1, .df1 = 0.4, .df2 = 0.6, .expected = 0.986301369863 + fd2_365, .tolerance = 1e-13 },
        .{ .y1 = 2020, .m1 = 1, .d1 = 1, .y2 = 2021, .m2 = 1, .d2 = 1, .df1 = 0.6, .df2 = 0.4, .expected = 0.986301369863 - fd2_365, .tolerance = 1e-13 },
        .{ .y1 = 2020, .m1 = 1, .d1 = 1, .y2 = 2020, .m2 = 1, .d2 = 1, .df1 = 0.4, .df2 = 0.6, .expected = fd2_365, .tolerance = 1e-16 },
    };
    for (cases) |tc| {
        const result = thirty365(tc.y1, tc.m1, tc.d1, tc.y2, tc.m2, tc.d2, tc.df1, tc.df2, false);
        try std.testing.expect(almostEqual(result, tc.expected, tc.tolerance));
    }
}

test "act365Fixed time fractions" {
    const cases = [_]struct { y1: i32, m1: i32, d1: i32, y2: i32, m2: i32, d2: i32, df1: f64, df2: f64, expected: f64, tolerance: f64 }{
        .{ .y1 = 2021, .m1 = 1, .d1 = 1, .y2 = 2022, .m2 = 1, .d2 = 1, .df1 = 0.5, .df2 = 0.5, .expected = 1, .tolerance = 1e-16 },
        .{ .y1 = 2020, .m1 = 1, .d1 = 1, .y2 = 2021, .m2 = 1, .d2 = 1, .df1 = 0.5, .df2 = 0.5, .expected = 1.0027397260274, .tolerance = 1e-13 },
        .{ .y1 = 2020, .m1 = 1, .d1 = 1, .y2 = 2021, .m2 = 1, .d2 = 1, .df1 = 0.4, .df2 = 0.6, .expected = 1.0027397260274 + fd2_365, .tolerance = 1e-13 },
        .{ .y1 = 2020, .m1 = 1, .d1 = 1, .y2 = 2021, .m2 = 1, .d2 = 1, .df1 = 0.6, .df2 = 0.4, .expected = 1.0027397260274 - fd2_365, .tolerance = 1e-13 },
        .{ .y1 = 2020, .m1 = 1, .d1 = 1, .y2 = 2020, .m2 = 1, .d2 = 1, .df1 = 0.4, .df2 = 0.6, .expected = fd2_365, .tolerance = 1e-13 },
    };
    for (cases) |tc| {
        const result = act365Fixed(tc.y1, tc.m1, tc.d1, tc.y2, tc.m2, tc.d2, tc.df1, tc.df2, false);
        try std.testing.expect(almostEqual(result, tc.expected, tc.tolerance));
    }
}

test "act365Fixed year fractions" {
    const result = act365Fixed(2012, 1, 1, 2012, 7, 30, 0, 0, false);
    try std.testing.expect(almostEqual(result, 0.57808219, 1e-8));
}

test "act365Fixed Excel basis 3" {
    const cases = [_]struct { y1: i32, m1: i32, d1: i32, y2: i32, m2: i32, d2: i32, expected: f64, precision: i32 }{
        .{ .y1 = 1978, .m1 = 2, .d1 = 28, .y2 = 2020, .m2 = 5, .d2 = 17, .expected = 42.2438356164384, .precision = 13 },
        .{ .y1 = 1993, .m1 = 12, .d1 = 2, .y2 = 2022, .m2 = 4, .d2 = 18, .expected = 28.3945205479452, .precision = 13 },
        .{ .y1 = 2018, .m1 = 12, .d1 = 15, .y2 = 2019, .m2 = 3, .d2 = 1, .expected = 0.208219178082192, .precision = 13 },
        .{ .y1 = 2018, .m1 = 12, .d1 = 31, .y2 = 2019, .m2 = 1, .d2 = 1, .expected = 0.0027397260273973, .precision = 13 },
        .{ .y1 = 1994, .m1 = 6, .d1 = 30, .y2 = 1997, .m2 = 6, .d2 = 30, .expected = 3.002739726027400, .precision = 13 },
        .{ .y1 = 1994, .m1 = 2, .d1 = 10, .y2 = 1994, .m2 = 6, .d2 = 30, .expected = 0.383561643835616, .precision = 13 },
        .{ .y1 = 2020, .m1 = 2, .d1 = 21, .y2 = 2024, .m2 = 3, .d2 = 25, .expected = 4.093150684931510, .precision = 13 },
        .{ .y1 = 2020, .m1 = 2, .d1 = 29, .y2 = 2021, .m2 = 2, .d2 = 28, .expected = 1.000000000000000, .precision = 16 },
        .{ .y1 = 2020, .m1 = 1, .d1 = 31, .y2 = 2021, .m2 = 2, .d2 = 28, .expected = 1.079452054794520, .precision = 13 },
        .{ .y1 = 2020, .m1 = 1, .d1 = 31, .y2 = 2021, .m2 = 3, .d2 = 31, .expected = 1.164383561643840, .precision = 13 },
        .{ .y1 = 2020, .m1 = 1, .d1 = 31, .y2 = 2020, .m2 = 4, .d2 = 30, .expected = 0.246575342465753, .precision = 13 },
        .{ .y1 = 2018, .m1 = 2, .d1 = 5, .y2 = 2023, .m2 = 5, .d2 = 14, .expected = 5.271232876712330, .precision = 13 },
        .{ .y1 = 2020, .m1 = 2, .d1 = 29, .y2 = 2024, .m2 = 2, .d2 = 28, .expected = 4.000000000000000, .precision = 16 },
        .{ .y1 = 2010, .m1 = 3, .d1 = 31, .y2 = 2015, .m2 = 8, .d2 = 30, .expected = 5.419178082191780, .precision = 13 },
        .{ .y1 = 2016, .m1 = 2, .d1 = 28, .y2 = 2016, .m2 = 10, .d2 = 30, .expected = 0.671232876712329, .precision = 13 },
        .{ .y1 = 2014, .m1 = 1, .d1 = 31, .y2 = 2014, .m2 = 8, .d2 = 31, .expected = 0.580821917808219, .precision = 13 },
        .{ .y1 = 2014, .m1 = 2, .d1 = 28, .y2 = 2014, .m2 = 9, .d2 = 30, .expected = 0.586301369863014, .precision = 13 },
        .{ .y1 = 2016, .m1 = 2, .d1 = 29, .y2 = 2016, .m2 = 6, .d2 = 15, .expected = 0.293150684931507, .precision = 13 },
        .{ .y1 = 2024, .m1 = 1, .d1 = 1, .y2 = 2024, .m2 = 12, .d2 = 31, .expected = 1.000000000000000, .precision = 16 },
        .{ .y1 = 2024, .m1 = 1, .d1 = 1, .y2 = 2025, .m2 = 1, .d2 = 2, .expected = 1.005479452054790, .precision = 13 },
        .{ .y1 = 2024, .m1 = 1, .d1 = 1, .y2 = 2024, .m2 = 2, .d2 = 29, .expected = 0.161643835616438, .precision = 13 },
        .{ .y1 = 2024, .m1 = 1, .d1 = 1, .y2 = 2024, .m2 = 3, .d2 = 1, .expected = 0.164383561643836, .precision = 13 },
        .{ .y1 = 2023, .m1 = 1, .d1 = 1, .y2 = 2023, .m2 = 3, .d2 = 1, .expected = 0.161643835616438, .precision = 13 },
        .{ .y1 = 2024, .m1 = 2, .d1 = 29, .y2 = 2025, .m2 = 2, .d2 = 28, .expected = 1.000000000000000, .precision = 16 },
        .{ .y1 = 2024, .m1 = 1, .d1 = 1, .y2 = 2028, .m2 = 12, .d2 = 31, .expected = 5.002739726027400, .precision = 13 },
        .{ .y1 = 2024, .m1 = 3, .d1 = 1, .y2 = 2025, .m2 = 3, .d2 = 1, .expected = 1.000000000000000, .precision = 16 },
        .{ .y1 = 2024, .m1 = 2, .d1 = 29, .y2 = 2025, .m2 = 3, .d2 = 1, .expected = 1.002739726027400, .precision = 13 },
        .{ .y1 = 2024, .m1 = 2, .d1 = 29, .y2 = 2028, .m2 = 2, .d2 = 28, .expected = 4.000000000000000, .precision = 16 },
        .{ .y1 = 2024, .m1 = 2, .d1 = 29, .y2 = 2028, .m2 = 2, .d2 = 29, .expected = 4.002739726027400, .precision = 13 },
        .{ .y1 = 2024, .m1 = 3, .d1 = 1, .y2 = 2028, .m2 = 3, .d2 = 1, .expected = 4.002739726027400, .precision = 13 },
    };
    for (cases) |tc| {
        const result = act365Fixed(tc.y1, tc.m1, tc.d1, tc.y2, tc.m2, tc.d2, 0, 0, false);
        try std.testing.expect(almostEqual(result, tc.expected, pow10(tc.precision)));
    }
}

test "act360 basic" {
    const result = act360(2018, 12, 15, 2019, 3, 1, 0, 0, false);
    try std.testing.expect(almostEqual(result, 0.2111111111111111, 1e-16));
}

test "act360 time fractions" {
    const cases = [_]struct { y1: i32, m1: i32, d1: i32, y2: i32, m2: i32, d2: i32, df1: f64, df2: f64, expected: f64, tolerance: f64 }{
        .{ .y1 = 2021, .m1 = 1, .d1 = 1, .y2 = 2022, .m2 = 1, .d2 = 1, .df1 = 0.5, .df2 = 0.5, .expected = 1.0138888888889, .tolerance = 1e-13 },
        .{ .y1 = 2020, .m1 = 1, .d1 = 1, .y2 = 2021, .m2 = 1, .d2 = 1, .df1 = 0.5, .df2 = 0.5, .expected = 1.0166666666667, .tolerance = 1e-13 },
        .{ .y1 = 2020, .m1 = 1, .d1 = 1, .y2 = 2021, .m2 = 1, .d2 = 1, .df1 = 0.4, .df2 = 0.6, .expected = 1.0166666666667 + fd2_360, .tolerance = 1e-13 },
        .{ .y1 = 2020, .m1 = 1, .d1 = 1, .y2 = 2021, .m2 = 1, .d2 = 1, .df1 = 0.6, .df2 = 0.4, .expected = 1.0166666666667 - fd2_360, .tolerance = 1e-13 },
        .{ .y1 = 2020, .m1 = 1, .d1 = 1, .y2 = 2020, .m2 = 1, .d2 = 1, .df1 = 0.4, .df2 = 0.6, .expected = fd2_360, .tolerance = 1e-13 },
    };
    for (cases) |tc| {
        const result = act360(tc.y1, tc.m1, tc.d1, tc.y2, tc.m2, tc.d2, tc.df1, tc.df2, false);
        try std.testing.expect(almostEqual(result, tc.expected, tc.tolerance));
    }
}

test "act360 Excel basis 2" {
    const cases = [_]struct { y1: i32, m1: i32, d1: i32, y2: i32, m2: i32, d2: i32, expected: f64, precision: i32 }{
        .{ .y1 = 1978, .m1 = 2, .d1 = 28, .y2 = 2020, .m2 = 5, .d2 = 17, .expected = 42.830555555555600, .precision = 13 },
        .{ .y1 = 1993, .m1 = 12, .d1 = 2, .y2 = 2022, .m2 = 4, .d2 = 18, .expected = 28.788888888888900, .precision = 13 },
        .{ .y1 = 2018, .m1 = 12, .d1 = 15, .y2 = 2019, .m2 = 3, .d2 = 1, .expected = 0.2111111111111110, .precision = 13 },
        .{ .y1 = 2018, .m1 = 12, .d1 = 31, .y2 = 2019, .m2 = 1, .d2 = 1, .expected = 0.0027777777777778, .precision = 13 },
        .{ .y1 = 1994, .m1 = 6, .d1 = 30, .y2 = 1997, .m2 = 6, .d2 = 30, .expected = 3.0444444444444400, .precision = 13 },
        .{ .y1 = 1994, .m1 = 2, .d1 = 10, .y2 = 1994, .m2 = 6, .d2 = 30, .expected = 0.3888888888888890, .precision = 13 },
        .{ .y1 = 2020, .m1 = 2, .d1 = 21, .y2 = 2024, .m2 = 3, .d2 = 25, .expected = 4.1500000000000000, .precision = 13 },
        .{ .y1 = 2020, .m1 = 2, .d1 = 29, .y2 = 2021, .m2 = 2, .d2 = 28, .expected = 1.0138888888888900, .precision = 13 },
        .{ .y1 = 2020, .m1 = 1, .d1 = 31, .y2 = 2021, .m2 = 2, .d2 = 28, .expected = 1.0944444444444400, .precision = 13 },
        .{ .y1 = 2020, .m1 = 1, .d1 = 31, .y2 = 2021, .m2 = 3, .d2 = 31, .expected = 1.1805555555555600, .precision = 13 },
        .{ .y1 = 2020, .m1 = 1, .d1 = 31, .y2 = 2020, .m2 = 4, .d2 = 30, .expected = 0.2500000000000000, .precision = 13 },
        .{ .y1 = 2018, .m1 = 2, .d1 = 5, .y2 = 2023, .m2 = 5, .d2 = 14, .expected = 5.3444444444444400, .precision = 13 },
        .{ .y1 = 2020, .m1 = 2, .d1 = 29, .y2 = 2024, .m2 = 2, .d2 = 28, .expected = 4.0555555555555600, .precision = 13 },
        .{ .y1 = 2010, .m1 = 3, .d1 = 31, .y2 = 2015, .m2 = 8, .d2 = 30, .expected = 5.4944444444444400, .precision = 13 },
        .{ .y1 = 2016, .m1 = 2, .d1 = 28, .y2 = 2016, .m2 = 10, .d2 = 30, .expected = 0.6805555555555560, .precision = 13 },
        .{ .y1 = 2014, .m1 = 1, .d1 = 31, .y2 = 2014, .m2 = 8, .d2 = 31, .expected = 0.5888888888888890, .precision = 13 },
        .{ .y1 = 2014, .m1 = 2, .d1 = 28, .y2 = 2014, .m2 = 9, .d2 = 30, .expected = 0.5944444444444440, .precision = 13 },
        .{ .y1 = 2016, .m1 = 2, .d1 = 29, .y2 = 2016, .m2 = 6, .d2 = 15, .expected = 0.2972222222222220, .precision = 13 },
        .{ .y1 = 2024, .m1 = 1, .d1 = 1, .y2 = 2024, .m2 = 12, .d2 = 31, .expected = 1.0138888888888900, .precision = 13 },
        .{ .y1 = 2024, .m1 = 1, .d1 = 1, .y2 = 2025, .m2 = 1, .d2 = 2, .expected = 1.0194444444444400, .precision = 13 },
        .{ .y1 = 2024, .m1 = 1, .d1 = 1, .y2 = 2024, .m2 = 2, .d2 = 29, .expected = 0.1638888888888890, .precision = 13 },
        .{ .y1 = 2024, .m1 = 1, .d1 = 1, .y2 = 2024, .m2 = 3, .d2 = 1, .expected = 0.1666666666666670, .precision = 13 },
        .{ .y1 = 2023, .m1 = 1, .d1 = 1, .y2 = 2023, .m2 = 3, .d2 = 1, .expected = 0.1638888888888890, .precision = 13 },
        .{ .y1 = 2024, .m1 = 2, .d1 = 29, .y2 = 2025, .m2 = 2, .d2 = 28, .expected = 1.0138888888888900, .precision = 13 },
        .{ .y1 = 2024, .m1 = 1, .d1 = 1, .y2 = 2028, .m2 = 12, .d2 = 31, .expected = 5.0722222222222200, .precision = 13 },
        .{ .y1 = 2024, .m1 = 3, .d1 = 1, .y2 = 2025, .m2 = 3, .d2 = 1, .expected = 1.0138888888888900, .precision = 13 },
        .{ .y1 = 2024, .m1 = 2, .d1 = 29, .y2 = 2025, .m2 = 3, .d2 = 1, .expected = 1.0166666666666700, .precision = 13 },
        .{ .y1 = 2024, .m1 = 2, .d1 = 29, .y2 = 2028, .m2 = 2, .d2 = 28, .expected = 4.0555555555555600, .precision = 13 },
        .{ .y1 = 2024, .m1 = 2, .d1 = 29, .y2 = 2028, .m2 = 2, .d2 = 29, .expected = 4.0583333333333300, .precision = 13 },
        .{ .y1 = 2024, .m1 = 3, .d1 = 1, .y2 = 2028, .m2 = 3, .d2 = 1, .expected = 4.0583333333333300, .precision = 13 },
    };
    for (cases) |tc| {
        const result = act360(tc.y1, tc.m1, tc.d1, tc.y2, tc.m2, tc.d2, 0, 0, false);
        try std.testing.expect(almostEqual(result, tc.expected, pow10(tc.precision)));
    }
}

test "actActExcel time fractions" {
    const cases = [_]struct { y1: i32, m1: i32, d1: i32, y2: i32, m2: i32, d2: i32, df1: f64, df2: f64, expected: f64, tolerance: f64 }{
        .{ .y1 = 2021, .m1 = 1, .d1 = 1, .y2 = 2022, .m2 = 1, .d2 = 1, .df1 = 0.5, .df2 = 0.5, .expected = 1, .tolerance = 1e-16 },
        .{ .y1 = 2020, .m1 = 1, .d1 = 1, .y2 = 2021, .m2 = 1, .d2 = 1, .df1 = 0.5, .df2 = 0.5, .expected = 1, .tolerance = 1e-16 },
        .{ .y1 = 2020, .m1 = 1, .d1 = 1, .y2 = 2021, .m2 = 1, .d2 = 1, .df1 = 0.4, .df2 = 0.6, .expected = 1 + fd2_366, .tolerance = 1e-13 },
        .{ .y1 = 2020, .m1 = 1, .d1 = 1, .y2 = 2021, .m2 = 1, .d2 = 1, .df1 = 0.6, .df2 = 0.4, .expected = 1 - fd2_366, .tolerance = 1e-13 },
        .{ .y1 = 2020, .m1 = 1, .d1 = 1, .y2 = 2020, .m2 = 1, .d2 = 1, .df1 = 0.4, .df2 = 0.6, .expected = fd2_366, .tolerance = 1e-13 },
    };
    for (cases) |tc| {
        const result = actActExcel(tc.y1, tc.m1, tc.d1, tc.y2, tc.m2, tc.d2, tc.df1, tc.df2, false);
        try std.testing.expect(almostEqual(result, tc.expected, tc.tolerance));
    }
}

test "actActExcel year fractions" {
    const result = actActExcel(2012, 1, 1, 2012, 7, 30, 0, 0, false);
    try std.testing.expect(almostEqual(result, 0.57650273, 1e-8));
}

test "actActExcel Excel basis 1" {
    const cases = [_]struct { y1: i32, m1: i32, d1: i32, y2: i32, m2: i32, d2: i32, expected: f64, precision: i32 }{
        .{ .y1 = 1978, .m1 = 2, .d1 = 28, .y2 = 2020, .m2 = 5, .d2 = 17, .expected = 42.21424933146570000, .precision = 13 },
        .{ .y1 = 1993, .m1 = 12, .d1 = 2, .y2 = 2022, .m2 = 4, .d2 = 18, .expected = 28.37638039609380000, .precision = 13 },
        .{ .y1 = 2018, .m1 = 12, .d1 = 15, .y2 = 2019, .m2 = 3, .d2 = 1, .expected = 0.208219178082192000, .precision = 13 },
        .{ .y1 = 2018, .m1 = 12, .d1 = 31, .y2 = 2019, .m2 = 1, .d2 = 1, .expected = 0.002739726027397260, .precision = 13 },
        .{ .y1 = 1994, .m1 = 6, .d1 = 30, .y2 = 1997, .m2 = 6, .d2 = 30, .expected = 3.000684462696780000, .precision = 13 },
        .{ .y1 = 1994, .m1 = 2, .d1 = 10, .y2 = 1994, .m2 = 6, .d2 = 30, .expected = 0.383561643835616000, .precision = 13 },
        .{ .y1 = 2020, .m1 = 2, .d1 = 21, .y2 = 2024, .m2 = 3, .d2 = 25, .expected = 4.088669950738920000, .precision = 13 },
        .{ .y1 = 2020, .m1 = 2, .d1 = 29, .y2 = 2021, .m2 = 2, .d2 = 28, .expected = 0.997267759562842000, .precision = 13 },
        .{ .y1 = 2020, .m1 = 1, .d1 = 31, .y2 = 2021, .m2 = 2, .d2 = 28, .expected = 1.077975376196990000, .precision = 13 },
        .{ .y1 = 2020, .m1 = 1, .d1 = 31, .y2 = 2021, .m2 = 3, .d2 = 31, .expected = 1.162790697674420000, .precision = 13 },
        .{ .y1 = 2020, .m1 = 1, .d1 = 31, .y2 = 2020, .m2 = 4, .d2 = 30, .expected = 0.245901639344262000, .precision = 13 },
        .{ .y1 = 2018, .m1 = 2, .d1 = 5, .y2 = 2023, .m2 = 5, .d2 = 14, .expected = 5.268827019625740000, .precision = 13 },
        .{ .y1 = 2020, .m1 = 2, .d1 = 29, .y2 = 2024, .m2 = 2, .d2 = 28, .expected = 3.995621237000550000, .precision = 13 },
        .{ .y1 = 2010, .m1 = 3, .d1 = 31, .y2 = 2015, .m2 = 8, .d2 = 30, .expected = 5.416704701049750000, .precision = 13 },
        .{ .y1 = 2016, .m1 = 2, .d1 = 28, .y2 = 2016, .m2 = 10, .d2 = 30, .expected = 0.669398907103825000, .precision = 13 },
        .{ .y1 = 2014, .m1 = 1, .d1 = 31, .y2 = 2014, .m2 = 8, .d2 = 31, .expected = 0.580821917808219000, .precision = 13 },
        .{ .y1 = 2014, .m1 = 2, .d1 = 28, .y2 = 2014, .m2 = 9, .d2 = 30, .expected = 0.586301369863014000, .precision = 13 },
        .{ .y1 = 2016, .m1 = 2, .d1 = 29, .y2 = 2016, .m2 = 6, .d2 = 15, .expected = 0.292349726775956000, .precision = 13 },
        .{ .y1 = 2024, .m1 = 1, .d1 = 1, .y2 = 2024, .m2 = 12, .d2 = 31, .expected = 0.997267759562842000, .precision = 13 },
        .{ .y1 = 2024, .m1 = 1, .d1 = 1, .y2 = 2025, .m2 = 1, .d2 = 2, .expected = 1.004103967168260000, .precision = 13 },
        .{ .y1 = 2024, .m1 = 1, .d1 = 1, .y2 = 2024, .m2 = 2, .d2 = 29, .expected = 0.161202185792350000, .precision = 13 },
        .{ .y1 = 2024, .m1 = 1, .d1 = 1, .y2 = 2024, .m2 = 3, .d2 = 1, .expected = 0.163934426229508000, .precision = 13 },
        .{ .y1 = 2023, .m1 = 1, .d1 = 1, .y2 = 2023, .m2 = 3, .d2 = 1, .expected = 0.161643835616438000, .precision = 13 },
        .{ .y1 = 2024, .m1 = 2, .d1 = 29, .y2 = 2025, .m2 = 2, .d2 = 28, .expected = 0.997267759562842000, .precision = 13 },
        .{ .y1 = 2024, .m1 = 1, .d1 = 1, .y2 = 2028, .m2 = 12, .d2 = 31, .expected = 4.997263273125340000, .precision = 13 },
        .{ .y1 = 2024, .m1 = 3, .d1 = 1, .y2 = 2025, .m2 = 3, .d2 = 1, .expected = 1.000000000000000000, .precision = 16 },
        .{ .y1 = 2024, .m1 = 2, .d1 = 29, .y2 = 2025, .m2 = 3, .d2 = 1, .expected = 1.001367989056090000, .precision = 13 },
        .{ .y1 = 2024, .m1 = 2, .d1 = 29, .y2 = 2028, .m2 = 2, .d2 = 28, .expected = 3.995621237000550000, .precision = 12 },
        .{ .y1 = 2024, .m1 = 2, .d1 = 29, .y2 = 2028, .m2 = 2, .d2 = 29, .expected = 3.998357963875210000, .precision = 13 },
        .{ .y1 = 2024, .m1 = 3, .d1 = 1, .y2 = 2028, .m2 = 3, .d2 = 1, .expected = 3.998357963875210000, .precision = 13 },
    };
    for (cases) |tc| {
        const result = actActExcel(tc.y1, tc.m1, tc.d1, tc.y2, tc.m2, tc.d2, 0, 0, false);
        try std.testing.expect(almostEqual(result, tc.expected, pow10(tc.precision)));
    }
}

test "actActIsda basic" {
    const cases = [_]struct { y1: i32, m1: i32, d1: i32, y2: i32, m2: i32, d2: i32, expected: f64, tolerance: f64 }{
        .{ .y1 = 2018, .m1 = 12, .d1 = 15, .y2 = 2019, .m2 = 3, .d2 = 1, .expected = 76.0 / 365.0, .tolerance = 1e-13 },
        .{ .y1 = 2018, .m1 = 12, .d1 = 31, .y2 = 2019, .m2 = 1, .d2 = 1, .expected = 1.0 / 365.0, .tolerance = 1e-13 },
        .{ .y1 = 1994, .m1 = 6, .d1 = 30, .y2 = 1997, .m2 = 6, .d2 = 30, .expected = 3.0, .tolerance = 1e-8 },
        .{ .y1 = 1994, .m1 = 2, .d1 = 10, .y2 = 1994, .m2 = 6, .d2 = 30, .expected = 140.0 / 365.0, .tolerance = 1e-8 },
    };
    for (cases) |tc| {
        const result = actActIsda(tc.y1, tc.m1, tc.d1, tc.y2, tc.m2, tc.d2, 0, 0, false);
        try std.testing.expect(almostEqual(result, tc.expected, tc.tolerance));
    }
}

test "actActIsda time fractions" {
    const cases = [_]struct { y1: i32, m1: i32, d1: i32, y2: i32, m2: i32, d2: i32, df1: f64, df2: f64, expected: f64, tolerance: f64 }{
        .{ .y1 = 2021, .m1 = 1, .d1 = 1, .y2 = 2022, .m2 = 1, .d2 = 1, .df1 = 0.5, .df2 = 0.5, .expected = 1.0, .tolerance = 1e-13 },
        .{ .y1 = 2020, .m1 = 1, .d1 = 1, .y2 = 2021, .m2 = 1, .d2 = 1, .df1 = 0.5, .df2 = 0.5, .expected = 1.0000037427951194, .tolerance = 1e-13 },
        .{ .y1 = 2020, .m1 = 1, .d1 = 1, .y2 = 2021, .m2 = 1, .d2 = 1, .df1 = 0.4, .df2 = 0.6, .expected = 1.000550939441575, .tolerance = 1e-13 },
        .{ .y1 = 2020, .m1 = 1, .d1 = 1, .y2 = 2021, .m2 = 1, .d2 = 1, .df1 = 0.6, .df2 = 0.4, .expected = 0.9994565461486637, .tolerance = 1e-13 },
        .{ .y1 = 2020, .m1 = 1, .d1 = 1, .y2 = 2020, .m2 = 1, .d2 = 1, .df1 = 0.4, .df2 = 0.6, .expected = fd2_366, .tolerance = 1e-13 },
    };
    for (cases) |tc| {
        const result = actActIsda(tc.y1, tc.m1, tc.d1, tc.y2, tc.m2, tc.d2, tc.df1, tc.df2, false);
        try std.testing.expect(almostEqual(result, tc.expected, tc.tolerance));
    }
}

test "actActIsda year fractions" {
    const result = actActIsda(2012, 1, 1, 2012, 7, 30, 0, 0, false);
    try std.testing.expect(almostEqual(result, 0.57650273, 1e-8));
}

test "actActIsda Excel basis 1" {
    const cases = [_]struct { y1: i32, m1: i32, d1: i32, y2: i32, m2: i32, d2: i32, expected: f64, precision: i32 }{
        .{ .y1 = 1978, .m1 = 2, .d1 = 28, .y2 = 2020, .m2 = 5, .d2 = 17, .expected = 42.214249331465700000, .precision = 2 },
        .{ .y1 = 1993, .m1 = 12, .d1 = 2, .y2 = 2022, .m2 = 4, .d2 = 18, .expected = 28.376380396093800000, .precision = 2 },
        .{ .y1 = 2018, .m1 = 12, .d1 = 15, .y2 = 2019, .m2 = 3, .d2 = 1, .expected = 0.208219178082192000, .precision = 2 },
        .{ .y1 = 2018, .m1 = 12, .d1 = 31, .y2 = 2019, .m2 = 1, .d2 = 1, .expected = 0.002739726027397260, .precision = 2 },
        .{ .y1 = 1994, .m1 = 6, .d1 = 30, .y2 = 1997, .m2 = 6, .d2 = 30, .expected = 3.000684462696780000, .precision = 2 },
        .{ .y1 = 1994, .m1 = 2, .d1 = 10, .y2 = 1994, .m2 = 6, .d2 = 30, .expected = 0.383561643835616000, .precision = 13 },
        .{ .y1 = 2020, .m1 = 2, .d1 = 21, .y2 = 2024, .m2 = 3, .d2 = 25, .expected = 4.088669950738920000, .precision = 2 },
        .{ .y1 = 2020, .m1 = 2, .d1 = 29, .y2 = 2021, .m2 = 2, .d2 = 28, .expected = 0.997267759562842000, .precision = 2 },
        .{ .y1 = 2020, .m1 = 1, .d1 = 31, .y2 = 2021, .m2 = 2, .d2 = 28, .expected = 1.077975376196990000, .precision = 2 },
        .{ .y1 = 2020, .m1 = 1, .d1 = 31, .y2 = 2021, .m2 = 3, .d2 = 31, .expected = 1.162790697674420000, .precision = 2 },
        .{ .y1 = 2020, .m1 = 1, .d1 = 31, .y2 = 2020, .m2 = 4, .d2 = 30, .expected = 0.245901639344262000, .precision = 13 },
        .{ .y1 = 2018, .m1 = 2, .d1 = 5, .y2 = 2023, .m2 = 5, .d2 = 14, .expected = 5.268827019625740000, .precision = 2 },
        .{ .y1 = 2020, .m1 = 2, .d1 = 29, .y2 = 2024, .m2 = 2, .d2 = 28, .expected = 3.995621237000550000, .precision = 2 },
        .{ .y1 = 2010, .m1 = 3, .d1 = 31, .y2 = 2015, .m2 = 8, .d2 = 30, .expected = 5.416704701049750000, .precision = 2 },
        .{ .y1 = 2016, .m1 = 2, .d1 = 28, .y2 = 2016, .m2 = 10, .d2 = 30, .expected = 0.669398907103825000, .precision = 13 },
        .{ .y1 = 2014, .m1 = 1, .d1 = 31, .y2 = 2014, .m2 = 8, .d2 = 31, .expected = 0.580821917808219000, .precision = 13 },
        .{ .y1 = 2014, .m1 = 2, .d1 = 28, .y2 = 2014, .m2 = 9, .d2 = 30, .expected = 0.586301369863014000, .precision = 13 },
        .{ .y1 = 2016, .m1 = 2, .d1 = 29, .y2 = 2016, .m2 = 6, .d2 = 15, .expected = 0.292349726775956000, .precision = 13 },
        .{ .y1 = 2024, .m1 = 1, .d1 = 1, .y2 = 2024, .m2 = 12, .d2 = 31, .expected = 0.997267759562842000, .precision = 13 },
        .{ .y1 = 2024, .m1 = 1, .d1 = 1, .y2 = 2025, .m2 = 1, .d2 = 2, .expected = 1.004103967168260000, .precision = 2 },
        .{ .y1 = 2024, .m1 = 1, .d1 = 1, .y2 = 2024, .m2 = 2, .d2 = 29, .expected = 0.161202185792350000, .precision = 13 },
        .{ .y1 = 2024, .m1 = 1, .d1 = 1, .y2 = 2024, .m2 = 3, .d2 = 1, .expected = 0.163934426229508000, .precision = 13 },
        .{ .y1 = 2023, .m1 = 1, .d1 = 1, .y2 = 2023, .m2 = 3, .d2 = 1, .expected = 0.161643835616438000, .precision = 13 },
        .{ .y1 = 2024, .m1 = 2, .d1 = 29, .y2 = 2025, .m2 = 2, .d2 = 28, .expected = 0.997267759562842000, .precision = 2 },
        .{ .y1 = 2024, .m1 = 1, .d1 = 1, .y2 = 2028, .m2 = 12, .d2 = 31, .expected = 4.997263273125340000, .precision = 2 },
        .{ .y1 = 2024, .m1 = 3, .d1 = 1, .y2 = 2025, .m2 = 3, .d2 = 1, .expected = 1.000000000000000000, .precision = 2 },
        .{ .y1 = 2024, .m1 = 2, .d1 = 29, .y2 = 2025, .m2 = 3, .d2 = 1, .expected = 1.001367989056090000, .precision = 2 },
        .{ .y1 = 2024, .m1 = 2, .d1 = 29, .y2 = 2028, .m2 = 2, .d2 = 28, .expected = 3.995621237000550000, .precision = 2 },
        .{ .y1 = 2024, .m1 = 2, .d1 = 29, .y2 = 2028, .m2 = 2, .d2 = 29, .expected = 3.998357963875210000, .precision = 2 },
        .{ .y1 = 2024, .m1 = 3, .d1 = 1, .y2 = 2028, .m2 = 3, .d2 = 1, .expected = 3.998357963875210000, .precision = 2 },
    };
    for (cases) |tc| {
        const result = actActIsda(tc.y1, tc.m1, tc.d1, tc.y2, tc.m2, tc.d2, 0, 0, false);
        try std.testing.expect(almostEqual(result, tc.expected, pow10(tc.precision)));
    }
}

test "actActAfb basic" {
    const cases = [_]struct { y1: i32, m1: i32, d1: i32, y2: i32, m2: i32, d2: i32, expected: f64, tolerance: f64 }{
        .{ .y1 = 2018, .m1 = 12, .d1 = 15, .y2 = 2019, .m2 = 3, .d2 = 1, .expected = 76.0 / 365.0, .tolerance = 1e-13 },
        .{ .y1 = 2018, .m1 = 12, .d1 = 31, .y2 = 2019, .m2 = 1, .d2 = 1, .expected = 1.0 / 365.0, .tolerance = 1e-13 },
        .{ .y1 = 1994, .m1 = 6, .d1 = 30, .y2 = 1997, .m2 = 6, .d2 = 30, .expected = 3.0, .tolerance = 1e-8 },
        .{ .y1 = 1994, .m1 = 2, .d1 = 10, .y2 = 1994, .m2 = 6, .d2 = 30, .expected = 140.0 / 365.0, .tolerance = 1e-8 },
    };
    for (cases) |tc| {
        const result = actActAfb(tc.y1, tc.m1, tc.d1, tc.y2, tc.m2, tc.d2, 0, 0, false);
        try std.testing.expect(almostEqual(result, tc.expected, tc.tolerance));
    }
}

test "actActAfb time fractions" {
    const cases = [_]struct { y1: i32, m1: i32, d1: i32, y2: i32, m2: i32, d2: i32, df1: f64, df2: f64, expected: f64, tolerance: f64 }{
        .{ .y1 = 2021, .m1 = 1, .d1 = 1, .y2 = 2022, .m2 = 1, .d2 = 1, .df1 = 0.5, .df2 = 0.5, .expected = 1.0, .tolerance = 1e-13 },
        .{ .y1 = 2020, .m1 = 1, .d1 = 1, .y2 = 2021, .m2 = 1, .d2 = 1, .df1 = 0.5, .df2 = 0.5, .expected = 1.0000037427951194, .tolerance = 1e-13 },
        .{ .y1 = 2020, .m1 = 1, .d1 = 1, .y2 = 2021, .m2 = 1, .d2 = 1, .df1 = 0.4, .df2 = 0.6, .expected = 1.000550939441575, .tolerance = 1e-13 },
        .{ .y1 = 2020, .m1 = 1, .d1 = 1, .y2 = 2021, .m2 = 1, .d2 = 1, .df1 = 0.6, .df2 = 0.4, .expected = 0.9994565461486637, .tolerance = 1e-13 },
        .{ .y1 = 2020, .m1 = 1, .d1 = 1, .y2 = 2020, .m2 = 1, .d2 = 1, .df1 = 0.4, .df2 = 0.6, .expected = fd2_366, .tolerance = 1e-13 },
    };
    for (cases) |tc| {
        const result = actActAfb(tc.y1, tc.m1, tc.d1, tc.y2, tc.m2, tc.d2, tc.df1, tc.df2, false);
        try std.testing.expect(almostEqual(result, tc.expected, tc.tolerance));
    }
}

test "actActAfb year fractions" {
    const result = actActAfb(2012, 1, 1, 2012, 7, 30, 0, 0, false);
    try std.testing.expect(almostEqual(result, 0.57650273, 1e-8));
}

test "actActAfb Excel basis 1" {
    const cases = [_]struct { y1: i32, m1: i32, d1: i32, y2: i32, m2: i32, d2: i32, expected: f64, precision: i32 }{
        .{ .y1 = 1978, .m1 = 2, .d1 = 28, .y2 = 2020, .m2 = 5, .d2 = 17, .expected = 42.214249331465700000, .precision = 2 },
        .{ .y1 = 1993, .m1 = 12, .d1 = 2, .y2 = 2022, .m2 = 4, .d2 = 18, .expected = 28.376380396093800000, .precision = 2 },
        .{ .y1 = 2018, .m1 = 12, .d1 = 15, .y2 = 2019, .m2 = 3, .d2 = 1, .expected = 0.208219178082192000, .precision = 2 },
        .{ .y1 = 2018, .m1 = 12, .d1 = 31, .y2 = 2019, .m2 = 1, .d2 = 1, .expected = 0.002739726027397260, .precision = 2 },
        .{ .y1 = 1994, .m1 = 6, .d1 = 30, .y2 = 1997, .m2 = 6, .d2 = 30, .expected = 3.000684462696780000, .precision = 2 },
        .{ .y1 = 1994, .m1 = 2, .d1 = 10, .y2 = 1994, .m2 = 6, .d2 = 30, .expected = 0.383561643835616000, .precision = 13 },
        .{ .y1 = 2020, .m1 = 2, .d1 = 21, .y2 = 2024, .m2 = 3, .d2 = 25, .expected = 4.088669950738920000, .precision = 2 },
        .{ .y1 = 2020, .m1 = 2, .d1 = 29, .y2 = 2021, .m2 = 2, .d2 = 28, .expected = 0.997267759562842000, .precision = 2 },
        .{ .y1 = 2020, .m1 = 1, .d1 = 31, .y2 = 2021, .m2 = 2, .d2 = 28, .expected = 1.077975376196990000, .precision = 2 },
        .{ .y1 = 2020, .m1 = 1, .d1 = 31, .y2 = 2021, .m2 = 3, .d2 = 31, .expected = 1.162790697674420000, .precision = 2 },
        .{ .y1 = 2020, .m1 = 1, .d1 = 31, .y2 = 2020, .m2 = 4, .d2 = 30, .expected = 0.245901639344262000, .precision = 13 },
        .{ .y1 = 2018, .m1 = 2, .d1 = 5, .y2 = 2023, .m2 = 5, .d2 = 14, .expected = 5.268827019625740000, .precision = 2 },
        .{ .y1 = 2020, .m1 = 2, .d1 = 29, .y2 = 2024, .m2 = 2, .d2 = 28, .expected = 3.995621237000550000, .precision = 2 },
        .{ .y1 = 2010, .m1 = 3, .d1 = 31, .y2 = 2015, .m2 = 8, .d2 = 30, .expected = 5.416704701049750000, .precision = 2 },
        .{ .y1 = 2016, .m1 = 2, .d1 = 28, .y2 = 2016, .m2 = 10, .d2 = 30, .expected = 0.669398907103825000, .precision = 13 },
        .{ .y1 = 2014, .m1 = 1, .d1 = 31, .y2 = 2014, .m2 = 8, .d2 = 31, .expected = 0.580821917808219000, .precision = 13 },
        .{ .y1 = 2014, .m1 = 2, .d1 = 28, .y2 = 2014, .m2 = 9, .d2 = 30, .expected = 0.586301369863014000, .precision = 13 },
        .{ .y1 = 2016, .m1 = 2, .d1 = 29, .y2 = 2016, .m2 = 6, .d2 = 15, .expected = 0.292349726775956000, .precision = 13 },
        .{ .y1 = 2024, .m1 = 1, .d1 = 1, .y2 = 2024, .m2 = 12, .d2 = 31, .expected = 0.997267759562842000, .precision = 13 },
        .{ .y1 = 2024, .m1 = 1, .d1 = 1, .y2 = 2025, .m2 = 1, .d2 = 2, .expected = 1.004103967168260000, .precision = 2 },
        .{ .y1 = 2024, .m1 = 1, .d1 = 1, .y2 = 2024, .m2 = 2, .d2 = 29, .expected = 0.161202185792350000, .precision = 13 },
        .{ .y1 = 2024, .m1 = 1, .d1 = 1, .y2 = 2024, .m2 = 3, .d2 = 1, .expected = 0.163934426229508000, .precision = 13 },
        .{ .y1 = 2023, .m1 = 1, .d1 = 1, .y2 = 2023, .m2 = 3, .d2 = 1, .expected = 0.161643835616438000, .precision = 13 },
        .{ .y1 = 2024, .m1 = 2, .d1 = 29, .y2 = 2025, .m2 = 2, .d2 = 28, .expected = 0.997267759562842000, .precision = 2 },
        .{ .y1 = 2024, .m1 = 1, .d1 = 1, .y2 = 2028, .m2 = 12, .d2 = 31, .expected = 4.997263273125340000, .precision = 2 },
        .{ .y1 = 2024, .m1 = 3, .d1 = 1, .y2 = 2025, .m2 = 3, .d2 = 1, .expected = 1.000000000000000000, .precision = 2 },
        .{ .y1 = 2024, .m1 = 2, .d1 = 29, .y2 = 2025, .m2 = 3, .d2 = 1, .expected = 1.001367989056090000, .precision = 2 },
        .{ .y1 = 2024, .m1 = 2, .d1 = 29, .y2 = 2028, .m2 = 2, .d2 = 28, .expected = 3.995621237000550000, .precision = 2 },
        .{ .y1 = 2024, .m1 = 2, .d1 = 29, .y2 = 2028, .m2 = 2, .d2 = 29, .expected = 3.998357963875210000, .precision = 2 },
        .{ .y1 = 2024, .m1 = 3, .d1 = 1, .y2 = 2028, .m2 = 3, .d2 = 1, .expected = 3.998357963875210000, .precision = 2 },
    };
    for (cases) |tc| {
        const result = actActAfb(tc.y1, tc.m1, tc.d1, tc.y2, tc.m2, tc.d2, 0, 0, false);
        try std.testing.expect(almostEqual(result, tc.expected, pow10(tc.precision)));
    }
}

test "act365Nonleap basic" {
    const result = act365Nonleap(2018, 12, 15, 2019, 3, 1, 0, 0, false);
    try std.testing.expect(almostEqual(result, 0.20821918, 1e-8));
}

test "act365Nonleap time fractions" {
    const cases = [_]struct { y1: i32, m1: i32, d1: i32, y2: i32, m2: i32, d2: i32, df1: f64, df2: f64, expected: f64, tolerance: f64 }{
        .{ .y1 = 2021, .m1 = 1, .d1 = 1, .y2 = 2022, .m2 = 1, .d2 = 1, .df1 = 0.5, .df2 = 0.5, .expected = 1, .tolerance = 1e-16 },
        .{ .y1 = 2020, .m1 = 1, .d1 = 1, .y2 = 2021, .m2 = 1, .d2 = 1, .df1 = 0.5, .df2 = 0.5, .expected = 1, .tolerance = 1e-16 },
        .{ .y1 = 2020, .m1 = 1, .d1 = 1, .y2 = 2021, .m2 = 1, .d2 = 1, .df1 = 0.4, .df2 = 0.6, .expected = 1 + fd2_365, .tolerance = 1e-13 },
        .{ .y1 = 2020, .m1 = 1, .d1 = 1, .y2 = 2021, .m2 = 1, .d2 = 1, .df1 = 0.6, .df2 = 0.4, .expected = 1 - fd2_365, .tolerance = 1e-13 },
        .{ .y1 = 2020, .m1 = 1, .d1 = 1, .y2 = 2020, .m2 = 1, .d2 = 1, .df1 = 0.4, .df2 = 0.6, .expected = -0.0021917808219, .tolerance = 1e-13 },
    };
    for (cases) |tc| {
        const result = act365Nonleap(tc.y1, tc.m1, tc.d1, tc.y2, tc.m2, tc.d2, tc.df1, tc.df2, false);
        try std.testing.expect(almostEqual(result, tc.expected, tc.tolerance));
    }
}

test "act365Nonleap Excel basis 3" {
    const cases = [_]struct { y1: i32, m1: i32, d1: i32, y2: i32, m2: i32, d2: i32, expected: f64, precision: i32 }{
        .{ .y1 = 1978, .m1 = 2, .d1 = 28, .y2 = 2020, .m2 = 5, .d2 = 17, .expected = 42.2438356164384, .precision = 1 },
        .{ .y1 = 1993, .m1 = 12, .d1 = 2, .y2 = 2022, .m2 = 4, .d2 = 18, .expected = 28.3945205479452, .precision = 1 },
        .{ .y1 = 2018, .m1 = 12, .d1 = 15, .y2 = 2019, .m2 = 3, .d2 = 1, .expected = 0.208219178082192, .precision = 13 },
        .{ .y1 = 2018, .m1 = 12, .d1 = 31, .y2 = 2019, .m2 = 1, .d2 = 1, .expected = 0.0027397260273973, .precision = 13 },
        .{ .y1 = 1994, .m1 = 6, .d1 = 30, .y2 = 1997, .m2 = 6, .d2 = 30, .expected = 3.002739726027400, .precision = 2 },
        .{ .y1 = 1994, .m1 = 2, .d1 = 10, .y2 = 1994, .m2 = 6, .d2 = 30, .expected = 0.383561643835616, .precision = 13 },
        .{ .y1 = 2020, .m1 = 2, .d1 = 21, .y2 = 2024, .m2 = 3, .d2 = 25, .expected = 4.093150684931510, .precision = 2 },
        .{ .y1 = 2020, .m1 = 2, .d1 = 29, .y2 = 2021, .m2 = 2, .d2 = 28, .expected = 1.000000000000000, .precision = 2 },
        .{ .y1 = 2020, .m1 = 1, .d1 = 31, .y2 = 2021, .m2 = 2, .d2 = 28, .expected = 1.079452054794520, .precision = 2 },
        .{ .y1 = 2020, .m1 = 1, .d1 = 31, .y2 = 2021, .m2 = 3, .d2 = 31, .expected = 1.164383561643840, .precision = 2 },
        .{ .y1 = 2020, .m1 = 1, .d1 = 31, .y2 = 2020, .m2 = 4, .d2 = 30, .expected = 0.246575342465753, .precision = 2 },
        .{ .y1 = 2018, .m1 = 2, .d1 = 5, .y2 = 2023, .m2 = 5, .d2 = 14, .expected = 5.271232876712330, .precision = 2 },
        .{ .y1 = 2020, .m1 = 2, .d1 = 29, .y2 = 2024, .m2 = 2, .d2 = 28, .expected = 4.000000000000000, .precision = 2 },
        .{ .y1 = 2010, .m1 = 3, .d1 = 31, .y2 = 2015, .m2 = 8, .d2 = 30, .expected = 5.419178082191780, .precision = 2 },
        .{ .y1 = 2016, .m1 = 2, .d1 = 28, .y2 = 2016, .m2 = 10, .d2 = 30, .expected = 0.671232876712329, .precision = 2 },
        .{ .y1 = 2014, .m1 = 1, .d1 = 31, .y2 = 2014, .m2 = 8, .d2 = 31, .expected = 0.580821917808219, .precision = 13 },
        .{ .y1 = 2014, .m1 = 2, .d1 = 28, .y2 = 2014, .m2 = 9, .d2 = 30, .expected = 0.586301369863014, .precision = 13 },
        .{ .y1 = 2016, .m1 = 2, .d1 = 29, .y2 = 2016, .m2 = 6, .d2 = 15, .expected = 0.293150684931507, .precision = 2 },
        .{ .y1 = 2024, .m1 = 1, .d1 = 1, .y2 = 2024, .m2 = 12, .d2 = 31, .expected = 1.000000000000000, .precision = 2 },
        .{ .y1 = 2024, .m1 = 1, .d1 = 1, .y2 = 2025, .m2 = 1, .d2 = 2, .expected = 1.005479452054790, .precision = 2 },
        .{ .y1 = 2024, .m1 = 1, .d1 = 1, .y2 = 2024, .m2 = 2, .d2 = 29, .expected = 0.161643835616438, .precision = 2 },
        .{ .y1 = 2024, .m1 = 1, .d1 = 1, .y2 = 2024, .m2 = 3, .d2 = 1, .expected = 0.164383561643836, .precision = 2 },
        .{ .y1 = 2023, .m1 = 1, .d1 = 1, .y2 = 2023, .m2 = 3, .d2 = 1, .expected = 0.161643835616438, .precision = 13 },
        .{ .y1 = 2024, .m1 = 2, .d1 = 29, .y2 = 2025, .m2 = 2, .d2 = 28, .expected = 1.000000000000000, .precision = 2 },
        .{ .y1 = 2024, .m1 = 1, .d1 = 1, .y2 = 2028, .m2 = 12, .d2 = 31, .expected = 5.002739726027400, .precision = 2 },
        .{ .y1 = 2024, .m1 = 3, .d1 = 1, .y2 = 2025, .m2 = 3, .d2 = 1, .expected = 1.000000000000000, .precision = 16 },
        .{ .y1 = 2024, .m1 = 2, .d1 = 29, .y2 = 2025, .m2 = 3, .d2 = 1, .expected = 1.002739726027400, .precision = 2 },
        .{ .y1 = 2024, .m1 = 2, .d1 = 29, .y2 = 2028, .m2 = 2, .d2 = 28, .expected = 4.000000000000000, .precision = 2 },
        .{ .y1 = 2024, .m1 = 2, .d1 = 29, .y2 = 2028, .m2 = 2, .d2 = 29, .expected = 4.002739726027400, .precision = 2 },
        .{ .y1 = 2024, .m1 = 3, .d1 = 1, .y2 = 2028, .m2 = 3, .d2 = 1, .expected = 4.002739726027400, .precision = 2 },
    };
    for (cases) |tc| {
        const result = act365Nonleap(tc.y1, tc.m1, tc.d1, tc.y2, tc.m2, tc.d2, 0, 0, false);
        try std.testing.expect(almostEqual(result, tc.expected, pow10(tc.precision)));
    }
}

test "fracDays" {
    const result_days = eur30360(2020, 1, 1, 2020, 2, 1, 0, 0, true);
    const result_years = eur30360(2020, 1, 1, 2020, 2, 1, 0, 0, false);
    try std.testing.expect(almostEqual(result_days, 30.0, epsilon));
    try std.testing.expect(almostEqual(result_years, 30.0 / 360.0, epsilon));
}
