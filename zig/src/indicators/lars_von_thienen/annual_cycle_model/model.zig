const std = @import("std");
const scanner = @import("../cycle_scanner/scanner.zig");
const hp = @import("../hodrick_prescott_smoothing_and_detrending/solver.zig");
const summation = @import("../harmonic_cycle/summation.zig");
const memory = @import("../harmonic_cycle/memory.zig");

pub const Params = struct { min_period: isize = 5, max_period: isize = 125, smoothing: f64 = 1600, genuine_threshold: f64 = 49, candidate_pool: isize = 10, max_components: isize = 7, peak_distance: isize = 2 };
pub const Component = struct { period: f64, amplitude: f64, phase: f64, strength: f64 = 0, genuine_percent: f64 = 0 };
pub const ModelConfig = struct { window: isize = 0, min_period: isize = 5, max_period: isize = 125, smoothing: f64 = 1600, genuine_threshold: f64 = 49, phase_cycles: f64 = 0, top_k: isize = 10, peak_distance: isize = 2, max_segments: isize = 5, candidate_pool: isize = 10, max_components: isize = 7 };
/// Version-one portable schema. Allocated slices belong to the caller's allocator.
pub const Model = struct {
    schema: []const u8 = "lvt-annual-cycle-model",
    version: isize = 1,
    research_reconstruction: bool = true,
    year: isize,
    sample_count: isize,
    origin: []const u8 = "first-observed-trading-day, zero-based ordinal",
    phase_convention: []const u8 = "sine radians at ordinal zero",
    components: []Component,
    correlation: f64 = 0,
    subsets_checked: isize = 0,
    candidate_count: isize = 0,
    config: ModelConfig = .{},
    dates: ?[]const []const u8 = null,
    start_date: ?[]const u8 = null,
    end_date: ?[]const u8 = null,
    coverage: ?[]const u8 = null,
    fitted_as_of: ?[]const u8 = null,
    /// Release owned components and the optional date list; date strings are borrowed.
    pub fn deinit(self: *Model, a: std.mem.Allocator) void {
        a.free(self.components);
        if (self.dates) |dates| a.free(dates);
    }
    pub fn validate(self: Model) !void {
        if (!std.mem.eql(u8, self.schema, "lvt-annual-cycle-model") or self.version != 1 or self.year < 1 or self.sample_count < 6 or !std.mem.eql(u8, self.phase_convention, "sine radians at ordinal zero") or self.components.len < 1 or self.components.len > 7) return error.InvalidAnnualModel;
        for (self.components, 0..) |c, i| {
            if (!std.math.isFinite(c.period) or c.period <= 2 or !std.math.isFinite(c.amplitude) or c.amplitude < 0 or !std.math.isFinite(c.phase)) return error.InvalidAnnualComponent;
            for (self.components[0..i]) |p| {
                if (p.period == c.period) return error.DuplicatePeriod;
            }
        }
        if (self.dates) |dates| {
            try validateDates(dates, self.year);
            if (dates.len != @as(usize, @intCast(self.sample_count))) return error.CalendarLengthMismatch;
        }
        if (self.fitted_as_of) |asof| {
            try validateDates(&.{asof}, null);
            if ((try dateYear(asof)) <= self.year) return error.IncompleteYear;
        }
    }
};
pub const Subset = struct { indices: []usize, correlation: f64, checked: isize };
/// Exhaustive signed Pearson subsets, with deterministic 1e-12 ties.
pub fn chooseSubset(a: std.mem.Allocator, target: []const f64, waves: []const []const f64, periods: []const f64, max_components: isize) !Subset {
    const n = target.len;
    const count = waves.len;
    if (n < 6 or count < 1 or count > 16 or count != periods.len or max_components < 1 or max_components > 7) return error.InvalidSubset;
    const tc = try a.alloc(f64, n);
    defer a.free(tc);
    const energy_terms = try a.alloc(f64, n);
    defer a.free(energy_terms);
    const mean = (try summation.sum(a, target)) / @as(f64, @floatFromInt(n));
    for (target, 0..) |value, i| {
        tc[i] = value - mean;
        energy_terms[i] = tc[i] * tc[i];
    }
    const energy = try summation.sum(a, energy_terms);
    if (energy == 0) return error.ConstantTarget;
    const centered = try a.alloc([]f64, count);
    defer a.free(centered);
    var initialized: usize = 0;
    defer for (centered[0..initialized]) |values| a.free(values);
    for (waves, 0..) |wave, i| {
        if (wave.len != n) return error.WaveLengthMismatch;
        centered[i] = try a.alloc(f64, n);
        initialized += 1;
        const m = (try summation.sum(a, wave)) / @as(f64, @floatFromInt(n));
        for (wave, 0..) |v, j| centered[i][j] = v - m;
    }
    var cov: [16]f64 = undefined;
    var gram: [16][16]f64 = undefined;
    const products = try a.alloc(f64, n);
    defer a.free(products);
    for (centered, 0..) |left, i| {
        for (left, 0..) |v, k| products[k] = tc[k] * v;
        cov[i] = try summation.sum(a, products);
        for (centered, 0..) |right, j| {
            for (left, 0..) |v, k| products[k] = v * right[k];
            gram[i][j] = try summation.sum(a, products);
        }
    }
    var best: [7]usize = undefined;
    var best_len: usize = 0;
    var best_key: [7]f64 = undefined;
    var best_score: f64 = -std.math.inf(f64);
    var checked: isize = 0;
    var k: usize = 1;
    while (k <= @min(@as(usize, @intCast(max_components)), count)) : (k += 1) {
        var mask: usize = 1;
        while (mask < (@as(usize, 1) << @intCast(count))) : (mask += 1) {
            if (@popCount(mask) != k) continue;
            var subset: [7]usize = undefined;
            var length: usize = 0;
            for (0..count) |i| {
                if (mask & (@as(usize, 1) << @intCast(i)) != 0) {
                    subset[length] = i;
                    length += 1;
                }
            }
            checked += 1;
            var numerator: [7]f64 = undefined;
            var denominator: [49]f64 = undefined;
            var key: [7]f64 = undefined;
            var term: usize = 0;
            for (subset[0..k], 0..) |i, j| {
                numerator[j] = cov[i];
                key[j] = periods[i];
                for (subset[0..k]) |other| {
                    denominator[term] = gram[i][other];
                    term += 1;
                }
            }
            const den = try summation.sum(a, denominator[0..term]);
            if (den <= 0) continue;
            const score = @max(-1.0, @min(1.0, (try summation.sum(a, numerator[0..k])) / @sqrt(energy * den)));
            std.mem.sort(f64, key[0..k], {}, std.sort.asc(f64));
            var smaller = best_len > k;
            if (best_len == k) {
                for (key[0..k], best_key[0..k]) |left, right| {
                    if (left != right) {
                        smaller = left < right;
                        break;
                    }
                }
            }
            if (score > best_score + 1e-12 or (@abs(score - best_score) <= 1e-12 and smaller)) {
                best_len = k;
                @memcpy(best[0..k], subset[0..k]);
                @memcpy(best_key[0..k], key[0..k]);
                best_score = score;
            }
        }
    }
    if (best_len == 0 or best_score <= 0) return error.NoPositiveSubset;
    return .{ .indices = try a.dupe(usize, best[0..best_len]), .correlation = best_score, .checked = checked };
}
/// Numerical annual fitting; decodeAnnual additionally enforces dates and coverage.
pub fn fitAnnualValues(a: std.mem.Allocator, prices: []const f64, year: isize, p: Params) !Model {
    var arena = std.heap.ArenaAllocator.init(a);
    defer arena.deinit();
    var model = try fitAnnualValuesInWorkspace(arena.allocator(), prices, year, p);
    model.components = try a.dupe(Component, model.components);
    return model;
}
fn fitAnnualValuesInWorkspace(a: std.mem.Allocator, prices: []const f64, year: isize, p: Params) !Model {
    if (year < 1 or p.candidate_pool < 1 or p.candidate_pool > 16 or p.max_components < 1 or p.max_components > 7) return error.InvalidAnnualBounds;
    const mutable = try a.dupe(f64, prices);
    defer a.free(mutable);
    const config = scanner.ScanConfig{ .window = @intCast(prices.len), .min_period = p.min_period, .max_period = p.max_period, .smoothing = p.smoothing, .genuine_threshold = p.genuine_threshold, .phase_cycles = 0, .top_k = p.candidate_pool, .peak_distance = p.peak_distance, .max_segments = 5 };
    const scan = try scanner.scanCycles(a, mutable, config);
    var count: usize = 0;
    for (scan.cycles) |c| {
        if (std.math.isFinite(c.period)) count += 1;
    }
    if (count == 0) return error.NoAcceptedAnnualCycles;
    const waves = try a.alloc([]const f64, count);
    const periods = try a.alloc(f64, count);
    const n = prices.len;
    for (scan.cycles[0..count], 0..) |c, i| {
        periods[i] = c.period;
        const wave = try a.alloc(f64, n);
        for (wave, 0..) |*v, j| v.* = try scanner.modeledValue(a, c, @as(f64, @floatFromInt(j)) - @as(f64, @floatFromInt(n - 1)));
        waves[i] = wave;
    }
    const selected = try chooseSubset(a, try hp.cycleData(a, mutable, p.smoothing), waves, periods, p.max_components);
    const components = try a.alloc(Component, selected.indices.len);
    for (selected.indices, 0..) |index, i| {
        const c = scan.cycles[index];
        var phase = @rem(c.phase - 2 * std.math.pi * @as(f64, @floatFromInt(n - 1)) / c.period, 2 * std.math.pi);
        if (phase < 0) phase += 2 * std.math.pi;
        components[i] = .{ .period = c.period, .amplitude = c.fit_amplitude, .phase = phase, .strength = c.strength, .genuine_percent = c.genuine_percent };
    }
    std.mem.sort(Component, components, {}, struct {
        fn less(_: void, left: Component, right: Component) bool {
            return left.period < right.period;
        }
    }.less);
    return .{ .year = year, .sample_count = @intCast(n), .components = components, .correlation = selected.correlation, .subsets_checked = selected.checked, .candidate_count = @intCast(count), .config = .{ .window = @intCast(n), .min_period = p.min_period, .max_period = p.max_period, .smoothing = p.smoothing, .genuine_threshold = p.genuine_threshold, .top_k = p.candidate_pool, .peak_distance = p.peak_distance, .candidate_pool = p.candidate_pool, .max_components = p.max_components } };
}
pub fn annualWave(a: std.mem.Allocator, model: Model, index: f64) !f64 {
    var values: [7]f64 = undefined;
    for (model.components, 0..) |c, i| values[i] = c.amplitude * @sin(2 * std.math.pi * index / c.period + c.phase);
    return summation.sum(a, values[0..model.components.len]);
}
pub fn dateYear(day: []const u8) !isize {
    if (day.len != 10) return error.InvalidDate;
    return std.fmt.parseInt(isize, day[0..4], 10);
}
pub fn validateDates(dates: []const []const u8, year: ?isize) !void {
    var previous: []const u8 = "";
    for (dates) |day| {
        if (day.len != 10 or day[4] != '-' or day[7] != '-') return error.InvalidDate;
        for (day, 0..) |c, i| {
            if (i != 4 and i != 7 and !std.ascii.isDigit(c)) return error.InvalidDate;
        }
        const y = try dateYear(day);
        const m = try std.fmt.parseInt(usize, day[5..7], 10);
        const d = try std.fmt.parseInt(usize, day[8..], 10);
        const leap = @rem(y, 4) == 0 and (@rem(y, 100) != 0 or @rem(y, 400) == 0);
        const lengths = [_]usize{ 31, if (leap) 29 else 28, 31, 30, 31, 30, 31, 31, 30, 31, 30, 31 };
        if (y < 1 or m < 1 or m > 12 or d < 1 or d > lengths[m - 1] or std.mem.order(u8, day, previous) != .gt or (year != null and year.? != y)) return error.InvalidDate;
        previous = day;
    }
}
pub const PriceRecord = struct { date: []const u8, close: f64 };
pub fn decodeAnnual(a: std.mem.Allocator, records: []const PriceRecord, year: isize, as_of: []const u8, calendar: ?[]const []const u8, confirm: bool, p: Params) !Model {
    try validateDates(&.{as_of}, null);
    if ((try dateYear(as_of)) <= year) return error.IncompleteYear;
    var dates: std.ArrayList([]const u8) = .empty;
    errdefer dates.deinit(a);
    var prices: std.ArrayList(f64) = .empty;
    defer prices.deinit(a);
    var previous: []const u8 = "";
    for (records) |r| {
        try validateDates(&.{r.date}, null);
        if (std.mem.order(u8, r.date, previous) != .gt) return error.NonIncreasingDates;
        previous = r.date;
        if ((try dateYear(r.date)) == year) {
            try dates.append(a, r.date);
            try prices.append(a, r.close);
        }
    }
    if (dates.items.len < 6 or !std.mem.eql(u8, dates.items[0][5..7], "01") or !std.mem.eql(u8, dates.items[dates.items.len - 1][5..7], "12")) return error.IncompleteAnnualCalendar;
    if (calendar) |expected| {
        try validateDates(expected, year);
        if (expected.len != dates.items.len) return error.CalendarMismatch;
        for (expected, dates.items) |x, y| {
            if (!std.mem.eql(u8, x, y)) return error.CalendarMismatch;
        }
    } else if (!confirm) return error.CompletenessRequired;
    var model = try fitAnnualValues(a, prices.items, year, p);
    model.dates = try dates.toOwnedSlice(a);
    model.start_date = model.dates.?[0];
    model.end_date = model.dates.?[model.dates.?.len - 1];
    model.fitted_as_of = as_of;
    model.coverage = if (calendar != null) "exact match to supplied full-year calendar" else "caller asserted complete year; session completeness unverified";
    return model;
}
pub const AnnualWaveReader = struct {
    allocator: std.mem.Allocator,
    arena: std.heap.ArenaAllocator,
    model: Model,
    index: usize = 0,
    pub fn init(a: std.mem.Allocator, model: Model) !AnnualWaveReader {
        try model.validate();
        var arena = std.heap.ArenaAllocator.init(a);
        errdefer arena.deinit();
        return .{ .allocator = a, .arena = arena, .model = try memory.clone(arena.allocator(), model) };
    }
    pub fn deinit(self: *AnnualWaveReader) void {
        self.arena.deinit();
    }
    pub fn reset(self: *AnnualWaveReader) void {
        self.index = 0;
    }
    pub fn update(self: *AnnualWaveReader) !f64 {
        const value = try annualWave(self.allocator, self.model, @floatFromInt(self.index));
        self.index += 1;
        return value;
    }
};
