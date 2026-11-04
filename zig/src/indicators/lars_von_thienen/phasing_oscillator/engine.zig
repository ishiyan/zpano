const Params = @import("phasing_oscillator.zig").PhasingOscillatorParams;
const std = @import("std");
const memory = @import("../harmonic_cycle/memory.zig");
const harmoniccycle = @import("../harmonic_cycle/model.zig");
const cyclescanner = @import("../cycle_scanner/scanner.zig");
// Native Go numerical port. Source conventions: description.md.
pub const Snapshot = struct {
    value: f64 = 0,
    short_score: f64 = 0,
    long_score: f64 = 0,
    short_period: f64 = 0,
    long_period: f64 = 0,
};
pub fn emptySnapshot(_: std.mem.Allocator) anyerror!Snapshot {
    const nan = std.math.nan(f64);
    return Snapshot{ .value = nan, .short_score = nan, .long_score = nan, .short_period = nan, .long_period = nan };
}
pub const Engine = struct {
    p: Params = .{},
    short: ?*cyclescanner.RollingScanner = null,
    long: ?*cyclescanner.RollingScanner = null,
    pub fn reset(self: *Engine, allocator: std.mem.Allocator) anyerror!void {
        const tuple1 = (try cyclescanner.newRollingScanner(allocator, cyclescanner.ScanConfig{ .window = self.p.window, .min_period = self.p.short_min, .max_period = self.p.short_max, .smoothing = self.p.smoothing, .genuine_threshold = self.p.genuine_threshold, .phase_cycles = self.p.phase_cycles, .top_k = 1, .peak_distance = 2, .max_segments = 5 }));
        self.short = tuple1;
        const tuple2 = (try cyclescanner.newRollingScanner(allocator, cyclescanner.ScanConfig{ .window = self.p.window, .min_period = self.p.long_min, .max_period = self.p.long_max, .smoothing = self.p.smoothing, .genuine_threshold = self.p.genuine_threshold, .phase_cycles = self.p.phase_cycles, .top_k = 1, .peak_distance = 2, .max_segments = 5 }));
        self.long = tuple2;
    }
    pub fn primed(self: *Engine, allocator: std.mem.Allocator) anyerror!bool {
        return ((try (self.short orelse unreachable).isPrimed(allocator)) and (try (self.long orelse unreachable).isPrimed(allocator)));
    }
    pub fn warmup(self: *Engine, _: std.mem.Allocator) anyerror!isize {
        return (self.p.window - 1);
    }
    pub fn update(self: *Engine, allocator: std.mem.Allocator, price: f64) anyerror!Snapshot {
        (try harmoniccycle.sample(allocator, price));
        if (std.math.isNan(price)) {
            return (try emptySnapshot(allocator));
        }
        const short = (try (self.short orelse unreachable).update(allocator, price)).cycles[@intCast(0)];
        const long = (try (self.long orelse unreachable).update(allocator, price)).cycles[@intCast(0)];
        const temporary3_0 = (try harmoniccycle.phaseState(allocator, short.phase));
        const temporary3_1 = (try harmoniccycle.phaseState(allocator, long.phase));
        const shortScore = temporary3_0;
        const longScore = temporary3_1;
        return Snapshot{ .value = (shortScore + longScore), .short_score = shortScore, .long_score = longScore, .short_period = short.period, .long_period = long.period };
    }
};
