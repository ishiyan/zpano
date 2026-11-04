// SPDX-License-Identifier: MPL-2.0
// Source: Detrended Rhythm Oscillator, © WhenToTrade / Lars von Thienen.
const Params = @import("detrended_rhythm_oscillators.zig").DetrendedRhythmOscillatorsParams;
const std = @import("std");
const memory = @import("../harmonic_cycle/memory.zig");
const harmoniccycle = @import("../harmonic_cycle/model.zig");
// Native Go numerical port. Source conventions: description.md.
pub const Snapshot = struct {
    value: f64 = 0,
    direction: f64 = 0,
    cycle_length: f64 = 0,
    pivot_value: f64 = 0,
    pivot_offset: f64 = 0,
};
pub fn emptySnapshot(_: std.mem.Allocator) anyerror!Snapshot {
    const nan = std.math.nan(f64);
    return Snapshot{ .value = nan, .direction = nan, .cycle_length = nan, .pivot_value = nan, .pivot_offset = nan };
}
pub const Engine = struct {
    p: Params = .{},
    fast: f64 = 0,
    slow: f64 = 0,
    window: []f64 = &.{},
    pivots: []Pivot = &.{},
    direction: isize = 0,
    samples_seen: isize = 0,
    last_cycle_length: f64 = 0,
    pub fn reset(self: *Engine, _: std.mem.Allocator) anyerror!void {
        self.fast = std.math.nan(f64);
        self.slow = std.math.nan(f64);
        self.window = &.{};
        self.pivots = &.{};
        self.direction = 0;
        self.samples_seen = 0;
        self.last_cycle_length = std.math.nan(f64);
    }
    pub fn primed(self: *Engine, allocator: std.mem.Allocator) anyerror!bool {
        return (self.samples_seen > (try self.warmup(allocator)));
    }
    pub fn warmup(self: *Engine, allocator: std.mem.Allocator) anyerror!isize {
        return (try rhythmWarmup(allocator, self.p));
    }
    pub fn update(self: *Engine, allocator: std.mem.Allocator, price: f64) anyerror!Snapshot {
        (try harmoniccycle.sample(allocator, price));
        if (std.math.isNan(price)) {
            return (try emptySnapshot(allocator));
        }
        var value = price;
        if (self.p.detrend) {
            self.fast = (try ema(allocator, price, self.fast, self.p.fast_length));
            self.slow = (try ema(allocator, price, self.slow, self.p.slow_length));
            value = std.math.nan(f64);
            if ((self.slow != @as(f64, 0))) {
                value = (try harmoniccycle.finite(allocator, ((@as(f64, 100) * ((self.fast - self.slow))) / self.slow)));
            }
        }
        const index = self.samples_seen;
        self.samples_seen += 1;
        if (std.math.isNan(value)) {
            return (try emptySnapshot(allocator));
        }
        self.window = (try harmoniccycle.push(allocator, self.window, value, self.p.zig_period));
        const temporary1_0 = self.window[@intCast(0)];
        const temporary1_1 = self.window[@intCast(0)];
        var maximum = temporary1_0;
        var minimum = temporary1_1;
        for (self.window) |item| {
            maximum = @max(maximum, item);
            minimum = @min(minimum, item);
        }
        const temporary2_0 = (value == maximum);
        const temporary2_1 = (value == minimum);
        const high = temporary2_0;
        const low = temporary2_1;
        var direction = self.direction;
        if ((high and !low)) {
            direction = 1;
        } else if ((low and !high)) {
            direction = -1;
        }
        if (((direction != 0) and ((high or low)))) {
            if (((direction != self.direction) or (self.pivots.len == 0))) {
                self.pivots = (try memory.concat(allocator, (try memory.dupe(allocator, Pivot, &[_]Pivot{Pivot{ .value = value, .index = index }})), self.pivots));
                if ((@as(isize, @intCast(self.pivots.len)) > (self.p.cycle_memory + 3))) {
                    self.pivots = self.pivots[0..@as(usize, @intCast((self.p.cycle_memory + 3)))];
                }
            } else if ((((direction == 1) and (value > self.pivots[@intCast(0)].value)) or ((direction == -1) and (value < self.pivots[@intCast(0)].value)))) {
                self.pivots[@intCast(0)] = Pivot{ .value = value, .index = index };
            }
        }
        self.direction = direction;
        var cycles = (try memory.make(allocator, f64, @intCast(0)));
        if ((@as(isize, @intCast(self.pivots.len)) >= 5)) {
            {
                var i: isize = 1;
                while ((i < (@as(isize, @intCast(self.pivots.len)) - 2))) : (i += 1) {
                    cycles = (try memory.append(allocator, cycles, @as(f64, @floatFromInt((self.pivots[@intCast(i)].index - self.pivots[@intCast((i + 2))].index)))));
                }
            }
        }
        if ((@as(isize, @intCast(cycles.len)) > self.p.cycle_memory)) {
            cycles = cycles[@as(usize, @intCast((@as(isize, @intCast(cycles.len)) - self.p.cycle_memory)))..];
        }
        self.last_cycle_length = (try median(allocator, cycles));
        if (!(try self.primed(allocator))) {
            return (try emptySnapshot(allocator));
        }
        const temporary3_0 = std.math.nan(f64);
        const temporary3_1 = std.math.nan(f64);
        var pivotValue = temporary3_0;
        var offset = temporary3_1;
        if ((self.pivots.len != 0)) {
            pivotValue = self.pivots[@intCast(0)].value;
            offset = @as(f64, @floatFromInt((self.pivots[@intCast(0)].index - index)));
        }
        return Snapshot{ .value = value, .direction = @as(f64, @floatFromInt(direction)), .cycle_length = self.last_cycle_length, .pivot_value = pivotValue, .pivot_offset = offset };
    }
};
pub const Pivot = struct {
    value: f64 = 0,
    index: isize = 0,
};
pub fn ema(allocator: std.mem.Allocator, value: f64, previous: f64, length: isize) anyerror!f64 {
    if (std.math.isNan(previous)) {
        return value;
    }
    const alpha = (@as(f64, 2) / @as(f64, @floatFromInt((length + 1))));
    return (try harmoniccycle.finite(allocator, ((alpha * value) + (((@as(f64, 1) - alpha)) * previous))));
}
pub fn rhythmWarmup(_: std.mem.Allocator, p: Params) anyerror!isize {
    if (p.detrend) {
        return (p.slow_length - 1);
    }
    return 0;
}
pub fn median(_: std.mem.Allocator, values: []f64) anyerror!f64 {
    if (((values.len == 0) or (@as(isize, @intCast(values.len)) < 1))) {
        return std.math.nan(f64);
    }
    std.mem.sort(f64, values, {}, std.sort.asc(f64));
    const middle = @divTrunc(@as(isize, @intCast(values.len)), 2);
    if ((@rem(@as(isize, @intCast(values.len)), 2) != 0)) {
        return values[@intCast(middle)];
    }
    return (((values[@intCast((middle - 1))] + values[@intCast(middle)])) / @as(f64, 2));
}
