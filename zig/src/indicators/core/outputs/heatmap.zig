const std = @import("std");

/// Maximum number of values in a heatmap column.
pub const max_heatmap_values = 256;

/// Holds a time stamp (x) and an array of values (z) corresponding to parameter (y) range
/// to paint a heatmap column.
pub const Heatmap = struct {
    time: i64,
    parameter_first: f64,
    parameter_last: f64,
    parameter_resolution: f64,
    value_min: f64,
    value_max: f64,
    values: [max_heatmap_values]f64 = undefined,
    values_len: usize = 0,

    /// Creates a new heatmap from a slice of values.
    pub fn new(
        time: i64,
        parameter_first: f64,
        parameter_last: f64,
        parameter_resolution: f64,
        value_min: f64,
        value_max: f64,
        values: []const f64,
    ) Heatmap {
        var h = Heatmap{
            .time = time,
            .parameter_first = parameter_first,
            .parameter_last = parameter_last,
            .parameter_resolution = parameter_resolution,
            .value_min = value_min,
            .value_max = value_max,
            .values_len = @min(values.len, max_heatmap_values),
        };
        @memcpy(h.values[0..h.values_len], values[0..h.values_len]);
        return h;
    }

    /// Creates a new empty heatmap with NaN min/max and no values.
    pub fn empty(time: i64, parameter_first: f64, parameter_last: f64, parameter_resolution: f64) Heatmap {
        return .{
            .time = time,
            .parameter_first = parameter_first,
            .parameter_last = parameter_last,
            .parameter_resolution = parameter_resolution,
            .value_min = std.math.nan(f64),
            .value_max = std.math.nan(f64),
        };
    }

    /// Indicates whether this heatmap is not initialized.
    pub fn isEmpty(self: Heatmap) bool {
        return self.values_len == 0;
    }

    /// Returns the values as a slice.
    pub fn valuesSlice(self: *const Heatmap) []const f64 {
        return self.values[0..self.values_len];
    }
};
