const std = @import("std");
const memory = @import("../harmonic_cycle/memory.zig");
const harmoniccycle = @import("../harmonic_cycle/model.zig");
const hodrickprescottsmoothinganddetrending = @import("../hodrick_prescott_smoothing_and_detrending/solver.zig");
const bartelsperiodicityscore = @import("../bartels_periodicity_score/scorer.zig");
// Native Go numerical port. Source conventions: description.md.
pub const ScanConfig = struct {
    window: isize = 0,
    min_period: isize = 0,
    max_period: isize = 0,
    smoothing: f64 = 0,
    genuine_threshold: f64 = 0,
    phase_cycles: f64 = 0,
    top_k: isize = 0,
    peak_distance: isize = 0,
    max_segments: isize = 0,
    pub fn validate(self: *const ScanConfig, allocator: std.mem.Allocator) anyerror!void {
        if (((((((self.window < 6) or (self.min_period < 3)) or (self.max_period < self.min_period)) or (self.top_k < 1)) or (self.peak_distance < 1)) or (self.max_segments < 2))) {
            return error.InvalidParameters;
        }
        try harmoniccycle.real(allocator, self.smoothing, "smoothing", 0);
        try harmoniccycle.real(allocator, self.genuine_threshold, "genuine_threshold", 0);
        try harmoniccycle.real(allocator, self.phase_cycles, "phase_cycles", 0);
        if (((self.genuine_threshold > @as(f64, 100)) or ((self.phase_cycles != @as(f64, 0)) and (self.phase_cycles < @as(f64, 1))))) {
            return error.InvalidParameters;
        }
        return;
    }
};
pub fn defaultConfig(_: std.mem.Allocator) anyerror!ScanConfig {
    return ScanConfig{ .window = 200, .min_period = 5, .max_period = 80, .smoothing = @as(f64, 1600), .genuine_threshold = @as(f64, 49), .phase_cycles = @as(f64, 2), .top_k = 3, .peak_distance = 2, .max_segments = 5 };
}
pub const RankedCycle = struct {
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
pub const ScanOutput = struct {
    amplitude: []f64 = &.{},
    genuine_percent: []f64 = &.{},
    cycles: []RankedCycle = &.{},
    accepted_count: f64 = 0,
};
pub fn emptyCycle(_: std.mem.Allocator) anyerror!RankedCycle {
    const nan = std.math.nan(f64);
    return RankedCycle{ .period = nan, .amplitude = nan, .phase = nan, .strength = nan, .genuine_percent = nan, .segments = nan, .fit_amplitude = nan, .fit_samples = nan, .value = nan, .slope = nan, .eta_top = nan, .eta_bottom = nan };
}
pub fn emptyScan(allocator: std.mem.Allocator, config: ScanConfig) anyerror!ScanOutput {
    const count = ((config.max_period - config.min_period) + 1);
    var output = ScanOutput{ .amplitude = (try memory.make(allocator, f64, @intCast(count))), .genuine_percent = (try memory.make(allocator, f64, @intCast(count))), .cycles = (try memory.make(allocator, RankedCycle, @intCast(config.top_k))), .accepted_count = std.math.nan(f64) };
    {
        var i: isize = 0;
        while (i < count) : (i += 1) {
            output.amplitude[@intCast(i)] = std.math.nan(f64);
            output.genuine_percent[@intCast(i)] = std.math.nan(f64);
        }
    }
    {
        var i: isize = 0;
        while (i < @as(isize, @intCast(output.cycles.len))) : (i += 1) {
            output.cycles[@intCast(i)] = (try emptyCycle(allocator));
        }
    }
    return output;
}
pub fn modeledValue(_: std.mem.Allocator, cycle: RankedCycle, offset: f64) anyerror!f64 {
    return (cycle.fit_amplitude * @sin((cycle.phase + (((@as(f64, 2) * std.math.pi) * offset) / cycle.period))));
}
pub fn modulo(_: std.mem.Allocator, input: f64) anyerror!f64 {
    var value = input;
    value = @rem(value, (@as(f64, 2) * std.math.pi));
    if ((value < @as(f64, 0))) {
        value += (@as(f64, 2) * std.math.pi);
    }
    return value;
}
pub fn scanCycles(allocator: std.mem.Allocator, prices: []f64, config: ScanConfig) anyerror!ScanOutput {
    try config.validate(allocator);
    if ((@as(isize, @intCast(prices.len)) < 6)) {
        return error.InvalidInput;
    }
    for (prices) |value| {
        (try harmoniccycle.sample(allocator, value));
        if (std.math.isNan(value)) {
            return error.InvalidInput;
        }
    }
    const residual = (try hodrickprescottsmoothinganddetrending.cycleData(allocator, prices, config.smoothing));
    var output = (try emptyScan(allocator, config));
    const count = @as(isize, @intCast(output.amplitude.len));
    var models = (try memory.make(allocator, harmoniccycle.CycleModel, @intCast(count)));
    var evidence = (try memory.make(allocator, bartelsperiodicityscore.BartelsOutput, @intCast(count)));
    {
        var i: isize = 0;
        while (i < count) : (i += 1) {
            const period = (config.min_period + i);
            models[@intCast(i)] = (try harmoniccycle.goertzel(allocator, residual, @as(f64, @floatFromInt(period))));
            const segments = @min(@divTrunc(@as(isize, @intCast(prices.len)), period), config.max_segments);
            evidence[@intCast(i)] = (try bartelsperiodicityscore.bartelsScore(allocator, prices[@as(usize, @intCast((@as(isize, @intCast(prices.len)) - (segments * period))))..], period, config.max_segments, 2));
            output.amplitude[@intCast(i)] = models[@intCast(i)].amplitude;
            output.genuine_percent[@intCast(i)] = evidence[@intCast(i)].genuine_percent;
        }
    }
    var candidates = (try memory.make(allocator, isize, @intCast(0)));
    for (models, 0..) |model, ordinal| {
        const i: isize = @intCast(ordinal);
        const temporary1_0 = -std.math.inf(f64);
        const temporary1_1 = -std.math.inf(f64);
        var left = temporary1_0;
        var right = temporary1_1;
        if ((i > 0)) {
            left = output.amplitude[@intCast((i - 1))];
        }
        if (((i + 1) < count)) {
            right = output.amplitude[@intCast((i + 1))];
        }
        const score = evidence[@intCast(i)];
        if ((((((model.amplitude > @as(f64, 0)) and (model.amplitude > left)) and (model.amplitude >= right)) and !std.math.isNan(score.genuine_percent)) and (score.genuine_percent > config.genuine_threshold))) {
            candidates = (try memory.append(allocator, candidates, i));
        }
    }
    {
        var x: isize = 1;
        while ((x < @as(isize, @intCast(candidates.len)))) : (x += 1) {
            {
                var y = x;
                while ((y > 0)) : (y -= 1) {
                    const temporary2_0 = models[@intCast(candidates[@intCast(y)])];
                    const temporary2_1 = models[@intCast(candidates[@intCast((y - 1))])];
                    const left = temporary2_0;
                    const right = temporary2_1;
                    const temporary3_0 = (left.amplitude / left.period);
                    const temporary3_1 = (right.amplitude / right.period);
                    const a = temporary3_0;
                    const b = temporary3_1;
                    if (((a < b) or ((a == b) and (left.period >= right.period)))) {
                        break;
                    }
                    const temporary4_0 = candidates[@intCast((y - 1))];
                    const temporary4_1 = candidates[@intCast(y)];
                    candidates[@intCast(y)] = temporary4_0;
                    candidates[@intCast((y - 1))] = temporary4_1;
                }
            }
        }
    }
    var accepted = (try memory.make(allocator, RankedCycle, @intCast(0)));
    for (candidates) |i| {
        const temporary5_0 = models[@intCast(i)];
        const temporary5_1 = evidence[@intCast(i)];
        const model = temporary5_0;
        const score = temporary5_1;
        var suppressed = false;
        for (accepted) |prior| {
            if ((@abs((model.period - prior.period)) < @as(f64, @floatFromInt(config.peak_distance)))) {
                suppressed = true;
                break;
            }
        }
        if (suppressed) {
            continue;
        }
        var fitN = @as(isize, @intCast(prices.len));
        if ((config.phase_cycles != @as(f64, 0))) {
            fitN = @min(fitN, @max(6, @as(isize, @intFromFloat(@ceil((config.phase_cycles * model.period))))));
        }
        var fit = model;
        if ((fitN != @as(isize, @intCast(prices.len)))) {
            fit = (try harmoniccycle.goertzel(allocator, (try hodrickprescottsmoothinganddetrending.cycleData(allocator, prices[@as(usize, @intCast((@as(isize, @intCast(prices.len)) - fitN)))..], config.smoothing)), model.period));
        }
        const phase = (try modulo(allocator, (fit.phase + (((@as(f64, 2) * std.math.pi) * @as(f64, @floatFromInt((fitN - 1)))) / model.period))));
        const omega = ((@as(f64, 2) * std.math.pi) / model.period);
        const value = (fit.amplitude * @sin(phase));
        const slope = ((fit.amplitude * omega) * @cos(phase));
        const etaTop = ((try modulo(allocator, ((std.math.pi / @as(f64, 2)) - phase))) / omega);
        const etaBottom = ((try modulo(allocator, (((@as(f64, 3) * std.math.pi) / @as(f64, 2)) - phase))) / omega);
        accepted = (try memory.append(allocator, accepted, RankedCycle{ .period = model.period, .amplitude = model.amplitude, .phase = phase, .strength = (model.amplitude / model.period), .genuine_percent = score.genuine_percent, .segments = score.segments, .fit_amplitude = fit.amplitude, .fit_samples = @as(f64, @floatFromInt(fitN)), .value = value, .slope = slope, .eta_top = etaTop, .eta_bottom = etaBottom }));
    }
    memory.copy(output.cycles, accepted);
    output.accepted_count = @as(f64, @floatFromInt(@as(isize, @intCast(accepted.len))));
    return output;
}
pub const RollingScanner = struct {
    config: ScanConfig = .{},
    prices: []f64 = &.{},
    samples_seen: isize = 0,
    pub fn reset(self: *RollingScanner, _: std.mem.Allocator) anyerror!void {
        self.prices = self.prices[0..@as(usize, @intCast(0))];
        self.samples_seen = 0;
    }
    pub fn isPrimed(self: *const RollingScanner, _: std.mem.Allocator) anyerror!bool {
        return (@as(isize, @intCast(self.prices.len)) == self.config.window);
    }
    pub fn update(self: *RollingScanner, allocator: std.mem.Allocator, price: f64) anyerror!ScanOutput {
        (try harmoniccycle.sample(allocator, price));
        if (std.math.isNan(price)) {
            return (try emptyScan(allocator, self.config));
        }
        const prices = (try harmoniccycle.push(allocator, (try memory.concat(allocator, (try memory.make(allocator, f64, @intCast(0))), self.prices)), price, self.config.window));
        var output = (try emptyScan(allocator, self.config));
        if ((@as(isize, @intCast(prices.len)) == self.config.window)) {
            output = (try scanCycles(allocator, prices, self.config));
        }
        self.prices = prices;
        self.samples_seen += 1;
        return output;
    }
};
pub fn newRollingScanner(allocator: std.mem.Allocator, config: ScanConfig) anyerror!?*RollingScanner {
    try config.validate(allocator);
    return (try memory.create(allocator, RollingScanner{ .config = config, .prices = (try memory.make(allocator, f64, @intCast(0))) }));
}
