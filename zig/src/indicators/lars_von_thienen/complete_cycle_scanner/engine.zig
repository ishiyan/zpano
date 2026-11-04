const Params = @import("complete_cycle_scanner.zig").CompleteCycleScannerParams;
const std = @import("std");
const memory = @import("../harmonic_cycle/memory.zig");
const harmoniccycle = @import("../harmonic_cycle/model.zig");
const cyclescanner = @import("../cycle_scanner/scanner.zig");
// Native Go numerical port. Source conventions: description.md.
pub const Snapshot = struct {
    spectrum: []f64 = &.{},
    genuine_percent: []f64 = &.{},
    dominant_period: f64 = 0,
    accepted_count: f64 = 0,
    cycles: []cyclescanner.RankedCycle = &.{},
};
pub fn emptySnapshot(_: std.mem.Allocator) anyerror!Snapshot {
    const nan = std.math.nan(f64);
    return Snapshot{ .spectrum = &.{}, .genuine_percent = &.{}, .dominant_period = nan, .accepted_count = nan, .cycles = &.{} };
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
        return Snapshot{ .spectrum = scan.amplitude, .genuine_percent = scan.genuine_percent, .dominant_period = scan.cycles[@intCast(0)].period, .accepted_count = scan.accepted_count, .cycles = scan.cycles };
    }
};
