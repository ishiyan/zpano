const std = @import("std");
const annual = @import("../annual_cycle_model/model.zig");
const summation = @import("../harmonic_cycle/summation.zig");
pub const metonic_lags = [_]isize{ 8, 11, 19, 27, 30, 38 };
pub const Scale = struct { year: isize, mean: f64, scale: f64, sample_count: isize };
pub const Model = struct {
    schema: []const u8 = "lvt-superposition-model",
    version: isize = 1,
    research_reconstruction: bool = true,
    target_year: isize,
    sample_count: usize,
    source_years: [6]isize,
    lags: [6]isize = metonic_lags,
    normalization: []const u8,
    calendar_policy: []const u8,
    invert: bool,
    scales: [6]Scale,
    annual_models: [6]annual.Model,
    dates: ?[]const []const u8 = null,
    pub fn validate(self: Model, a: std.mem.Allocator) !void {
        if (!std.mem.eql(u8, self.schema, "lvt-superposition-model") or self.version != 1) return error.InvalidSuperpositionSchema;
        for (metonic_lags, 0..) |lag, i| {
            if (self.lags[i] != lag or self.source_years[i] != self.target_year - lag) return error.InvalidSourceMapping;
        }
        const result = try superpositionValues(a, &self.annual_models, self.target_year, self.sample_count, self.normalization, self.calendar_policy, self.invert);
        a.free(result.values);
        if (self.dates) |d| {
            try annual.validateDates(d, self.target_year);
            if (d.len != self.sample_count) return error.CalendarLengthMismatch;
        }
    }
};
pub const Result = struct { values: []f64, model: Model };
/// Ordinal alignment and source-year-only normalization. Slices belong to allocator.
pub fn superpositionValues(a: std.mem.Allocator, models: []const annual.Model, year: isize, count: usize, normalization: []const u8, policy: []const u8, invert: bool) !Result {
    if (year < 39 or count < 1 or models.len != 6 or (!std.mem.eql(u8, normalization, "none") and !std.mem.eql(u8, normalization, "rms")) or (!std.mem.eql(u8, policy, "strict") and !std.mem.eql(u8, policy, "extend"))) return error.InvalidSuperpositionConfig;
    for (models, 0..) |m, i| {
        try m.validate();
        for (models[0..i]) |other| {
            if (other.year == m.year) return error.DuplicateYear;
        }
    }
    var model = Model{ .target_year = year, .sample_count = count, .source_years = undefined, .normalization = normalization, .calendar_policy = policy, .invert = invert, .scales = undefined, .annual_models = undefined };
    for (metonic_lags, 0..) |lag, i| {
        const source_year = year - lag;
        var found: ?annual.Model = null;
        for (models) |m| {
            if (m.year == source_year) found = m;
        }
        const m = found orelse return error.MissingSourceYear;
        const n: @TypeOf(count) = @intCast(m.sample_count);
        if (std.mem.eql(u8, policy, "strict") and n != count) return error.OrdinalLengthMismatch;
        const base = try a.alloc(f64, n);
        defer a.free(base);
        for (base, 0..) |*v, j| v.* = try annual.annualWave(a, m, @floatFromInt(j));
        var mean: f64 = 0;
        var scale: f64 = 1;
        if (std.mem.eql(u8, normalization, "rms")) {
            mean = (try summation.sum(a, base)) / @as(f64, @floatFromInt(n));
            for (base) |*v| v.* = (v.* - mean) * (v.* - mean);
            scale = @sqrt((try summation.sum(a, base)) / @as(f64, @floatFromInt(n)));
        }
        if (scale == 0) return error.ZeroSourceEnergy;
        model.source_years[i] = source_year;
        model.annual_models[i] = m;
        model.scales[i] = .{ .year = source_year, .mean = mean, .scale = scale, .sample_count = m.sample_count };
    }
    const values = try a.alloc(f64, count);
    for (values, 0..) |*v, i| {
        var terms: [6]f64 = undefined;
        for (model.annual_models, model.scales, 0..) |m, s, j| terms[j] = ((try annual.annualWave(a, m, @floatFromInt(i))) - s.mean) / s.scale;
        v.* = (if (invert) @as(f64, -1) else 1) * (try summation.sum(a, &terms)) / 6;
    }
    return .{ .values = values, .model = model };
}
pub const SuperpositionWaveReader = struct {
    allocator: std.mem.Allocator,
    values: []f64,
    index: usize = 0,
    pub fn init(a: std.mem.Allocator, models: []const annual.Model, year: isize, count: usize, norm: []const u8, policy: []const u8, invert: bool) !SuperpositionWaveReader {
        return .{ .allocator = a, .values = (try superpositionValues(a, models, year, count, norm, policy, invert)).values };
    }
    pub fn deinit(self: *SuperpositionWaveReader) void {
        self.allocator.free(self.values);
    }
    pub fn reset(self: *SuperpositionWaveReader) void {
        self.index = 0;
    }
    pub fn update(self: *SuperpositionWaveReader) ?f64 {
        if (self.index == self.values.len) return null;
        const v = self.values[self.index];
        self.index += 1;
        return v;
    }
};
