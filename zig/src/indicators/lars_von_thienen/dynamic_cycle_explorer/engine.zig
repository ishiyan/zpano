const Params = @import("dynamic_cycle_explorer.zig").DynamicCycleExplorerParams;
const std = @import("std");
const memory = @import("../harmonic_cycle/memory.zig");
const harmoniccycle = @import("../harmonic_cycle/model.zig");
const cyclescanner = @import("../cycle_scanner/scanner.zig");
// Native Go numerical port. Source conventions: description.md.
pub const Snapshot = struct {
    value: f64 = 0,
    period: f64 = 0,
    phase: f64 = 0,
    eta_top: f64 = 0,
    eta_bottom: f64 = 0,
    highlighter: f64 = 0,
    highlighter_index: f64 = 0,
    fitted_at_highlighter: f64 = 0,
};
pub fn emptySnapshot(_: std.mem.Allocator) anyerror!Snapshot {
    const nan = std.math.nan(f64);
    return Snapshot{ .value = nan, .period = nan, .phase = nan, .eta_top = nan, .eta_bottom = nan, .highlighter = nan, .highlighter_index = nan, .fitted_at_highlighter = nan };
}
pub const Engine = struct {
    p: Params = .{},
    scanner: ?*cyclescanner.RollingScanner = null,
    pub fn reset(self: *Engine, allocator: std.mem.Allocator) anyerror!void {
        const tuple1 = (try cyclescanner.newRollingScanner(allocator, cyclescanner.ScanConfig{ .window = self.p.window, .min_period = self.p.min_period, .max_period = self.p.max_period, .smoothing = self.p.smoothing, .genuine_threshold = self.p.genuine_threshold, .phase_cycles = self.p.phase_cycles, .top_k = self.p.top_k, .peak_distance = self.p.peak_distance, .max_segments = self.p.max_segments }));
        self.scanner = tuple1;
    }
    pub fn primed(self: *Engine, allocator: std.mem.Allocator) anyerror!bool {
        return (try (self.scanner orelse unreachable).isPrimed(allocator));
    }
    pub fn warmup(self: *Engine, _: std.mem.Allocator) anyerror!isize {
        return (self.p.window - 1);
    }
    pub fn update(self: *Engine, allocator: std.mem.Allocator, price: f64) anyerror!Snapshot {
        (try harmoniccycle.sample(allocator, price));
        if (std.math.isNan(price)) {
            return (try emptySnapshot(allocator));
        }
        const scan = (try (self.scanner orelse unreachable).update(allocator, price));
        const cycle = scan.cycles[@intCast(0)];
        if (std.math.isNan(cycle.period)) {
            return (try emptySnapshot(allocator));
        }
        const longLength = @as(isize, @intFromFloat(cycle.period));
        const shortLength = @max(2, @as(isize, @intFromFloat(@floor(((cycle.period / 2) + 0.5)))));
        const delay = @divTrunc(longLength, 2);
        const center = ((@as(isize, @intCast((self.scanner orelse unreachable).prices.len)) - 1) - delay);
        const highlighter = ((try centeredMean(allocator, (self.scanner orelse unreachable).prices, center, shortLength)) - (try centeredMean(allocator, (self.scanner orelse unreachable).prices, center, longLength)));
        const temporary2_0 = std.math.nan(f64);
        const temporary2_1 = std.math.nan(f64);
        var index = temporary2_0;
        var fitted = temporary2_1;
        if ((!std.math.isNan(highlighter) and !std.math.isInf(highlighter))) {
            index = @as(f64, @floatFromInt((((self.scanner orelse unreachable).samples_seen - 1) - delay)));
            fitted = (try cyclescanner.modeledValue(allocator, cycle, -@as(f64, @floatFromInt(delay))));
        }
        return Snapshot{ .value = cycle.value, .period = cycle.period, .phase = cycle.phase, .eta_top = cycle.eta_top, .eta_bottom = cycle.eta_bottom, .highlighter = highlighter, .highlighter_index = index, .fitted_at_highlighter = fitted };
    }
};
pub fn centeredMean(_: std.mem.Allocator, values: []f64, center: isize, length: isize) anyerror!f64 {
    const half = @divTrunc(length, 2);
    if ((((center - half) < 0) or ((center + half) >= @as(isize, @intCast(values.len))))) {
        return std.math.nan(f64);
    }
    var sum: f64 = 0;

    if ((@rem(length, 2) != 0)) {
        for (values[@as(usize, @intCast((center - half)))..@as(usize, @intCast(((center + half) + 1)))]) |value| {
            sum += value;
        }
        return (sum / @as(f64, @floatFromInt(length)));
    }
    for (values[@as(usize, @intCast(((center - half) + 1)))..@as(usize, @intCast((center + half)))]) |value| {
        sum += value;
    }
    return (((((0.5 * values[@intCast((center - half))]) + sum) + (0.5 * values[@intCast((center + half))]))) / @as(f64, @floatFromInt(length)));
}
