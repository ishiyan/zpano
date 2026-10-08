// Code generated from py/performance/reference_data/drawdown_average_length.py. DO NOT EDIT.

const std = @import("std");
const Entry = @import("entry.zig").Entry;

/// `drawdown_average_length.EXPECTED_VALUES`
pub const expected_values: []const f64 = &.{
    0.0, 0.0, 0.0, 1.0,
    2.0, 2.0, 2.0, 2.0,
    1.5, 2.0, 1.6666666666666667, 2.0,
    2.0, 1.75, 2.0, 2.25,
    2.5, 2.75, 3.0, 3.25,
    3.5, 3.75, 4.0, 4.25,
};

/// `drawdown_average_length.EXPECTED_VALUES_ORIGINAL`
pub const expected_values_original: []const f64 = &.{
    0.0, 0.0, 0.0, 2.0,
    2.0, 2.0, 2.0, 2.0,
    2.0, 2.0, 2.0, 2.0,
    2.0, 2.0, 2.25, 2.5,
    2.75, 3.0, 3.25, 3.5,
    3.75, 4.0, 4.25, 4.5,
};
