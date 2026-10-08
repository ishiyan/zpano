//! Streaming time-series performance and risk measures
//! (port of Python `performance/__init__.py`).
//!
//! `Measures` calculates ~135 performance and risk measures of a portfolio
//! and its benchmark one return observation at a time: moments, VaR/ES,
//! partial moments, Sharpe/Sortino/Omega/Kappa families, drawdowns and
//! CDaR, single-factor-model and benchmark-relative measures, capture
//! ratios, and miscellaneous ratios. Periodicity is given explicitly by
//! `periods_per_annum`; no timestamps are used.
//!
//! `core` holds the streaming building blocks used by `Measures`.

const rd = @import("reference_data/reference_data.zig");

pub const core = @import("core/core.zig");

pub const measures = @import("measures.zig");
pub const Measures = measures.Measures;
pub const Error = measures.Error;
pub const AddReturnError = measures.AddReturnError;
pub const isNormalFromJb = measures.isNormalFromJb;

pub const periods_per_annum_year = measures.periods_per_annum_year;
pub const periods_per_annum_quarter = measures.periods_per_annum_quarter;
pub const periods_per_annum_month = measures.periods_per_annum_month;
pub const periods_per_annum_week = measures.periods_per_annum_week;
pub const periods_per_annum_day = measures.periods_per_annum_day;
pub const periods_per_annum_minute_us_equities = measures.periods_per_annum_minute_us_equities;
pub const periods_per_annum_minute_crypto = measures.periods_per_annum_minute_crypto;

test {
    _ = core;
}

test {
    inline for (@typeInfo(rd).@"struct".decls) |d| {
        const m = @field(rd, d.name);
        if (@TypeOf(m) == type) {
            inline for (@typeInfo(m).@"struct".decls) |d2| {
                _ = @field(m, d2.name);
            }
        }
    }
}

test {
    _ = @import("measures_test.zig");
}
