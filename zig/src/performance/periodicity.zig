const std = @import("std");

/// Periodicity represents the frequency of performance measurement periods.
pub const Periodicity = enum {
    /// Daily periodicity (252 trading days per year).
    daily,

    /// Weekly periodicity (52 weeks per year).
    weekly,

    /// Monthly periodicity (12 months per year).
    monthly,

    /// Quarterly periodicity (4 quarters per year).
    quarterly,

    /// Annual periodicity (1 period per year).
    annual,

    /// Returns the number of periods per year for a given periodicity.
    pub fn periodsPerAnnum(self: Periodicity) u16 {
        return switch (self) {
            .daily => 252,
            .weekly => 52,
            .monthly => 12,
            .quarterly => 4,
            .annual => 1,
        };
    }

    /// Returns the number of trading days per period for a given periodicity.
    pub fn daysPerPeriod(self: Periodicity) f64 {
        return switch (self) {
            .daily => 1.0,
            .weekly => 252.0 / 52.0,
            .monthly => 252.0 / 12.0,
            .quarterly => 252.0 / 4.0,
            .annual => 252.0,
        };
    }
};

// ============== Tests ==============

test "periodsPerAnnum" {
    try std.testing.expectEqual(@as(u16, 252), Periodicity.daily.periodsPerAnnum());
    try std.testing.expectEqual(@as(u16, 52), Periodicity.weekly.periodsPerAnnum());
    try std.testing.expectEqual(@as(u16, 12), Periodicity.monthly.periodsPerAnnum());
    try std.testing.expectEqual(@as(u16, 4), Periodicity.quarterly.periodsPerAnnum());
    try std.testing.expectEqual(@as(u16, 1), Periodicity.annual.periodsPerAnnum());
}

test "daysPerPeriod" {
    try std.testing.expectEqual(@as(f64, 1.0), Periodicity.daily.daysPerPeriod());
    try std.testing.expectEqual(@as(f64, 252.0 / 52.0), Periodicity.weekly.daysPerPeriod());
    try std.testing.expectEqual(@as(f64, 252.0 / 12.0), Periodicity.monthly.daysPerPeriod());
    try std.testing.expectEqual(@as(f64, 252.0 / 4.0), Periodicity.quarterly.daysPerPeriod());
    try std.testing.expectEqual(@as(f64, 252.0), Periodicity.annual.daysPerPeriod());
}
