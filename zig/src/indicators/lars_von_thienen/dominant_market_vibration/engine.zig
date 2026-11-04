const Params = @import("dominant_market_vibration.zig").DominantMarketVibrationParams;
const std = @import("std");
const memory = @import("../harmonic_cycle/memory.zig");
const harmoniccycle = @import("../harmonic_cycle/model.zig");
const cyclescanner = @import("../cycle_scanner/scanner.zig");
// Native Go numerical port. Source conventions: description.md.
pub const Snapshot = struct {
    period: f64 = 0,
    amplitude: f64 = 0,
    phase: f64 = 0,
    strength: f64 = 0,
    genuine_percent: f64 = 0,
    segments: f64 = 0,
    fit_amplitude: f64 = 0,
    fit_samples: f64 = 0,
    value: f64 = 0,
    slope: f64 = 0,
    eta_top: f64 = 0,
    eta_bottom: f64 = 0,
};
pub fn emptySnapshot(_: std.mem.Allocator) anyerror!Snapshot {
    const nan = std.math.nan(f64);
    return Snapshot{ .period = nan, .amplitude = nan, .phase = nan, .strength = nan, .genuine_percent = nan, .segments = nan, .fit_amplitude = nan, .fit_samples = nan, .value = nan, .slope = nan, .eta_top = nan, .eta_bottom = nan };
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
        return Snapshot{ .period = cycle.period, .amplitude = cycle.amplitude, .phase = cycle.phase, .strength = cycle.strength, .genuine_percent = cycle.genuine_percent, .segments = cycle.segments, .fit_amplitude = cycle.fit_amplitude, .fit_samples = cycle.fit_samples, .value = cycle.value, .slope = cycle.slope, .eta_top = cycle.eta_top, .eta_bottom = cycle.eta_bottom };
    }
};
