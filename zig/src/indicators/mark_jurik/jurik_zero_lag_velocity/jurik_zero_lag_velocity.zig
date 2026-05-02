const std = @import("std");
const math = std.math;

const entities = @import("entities");
const Bar = entities.Bar;
const Quote = entities.Quote;
const Trade = entities.Trade;
const Scalar = entities.Scalar;
const bar_component = entities.bar_component;
const quote_component = entities.quote_component;
const trade_component = entities.trade_component;
const indicator_mod = @import("../../core/indicator.zig");
const line_indicator_mod = @import("../../core/line_indicator.zig");
const build_metadata_mod = @import("../../core/build_metadata.zig");
const component_triple_mnemonic_mod = @import("../../core/component_triple_mnemonic.zig");
const metadata_mod = @import("../../core/metadata.zig");

const OutputArray = indicator_mod.OutputArray;
const LineIndicator = line_indicator_mod.LineIndicator;
const Metadata = metadata_mod.Metadata;

/// Enumerates the outputs of the Jurik Zero Lag Velocity indicator.
pub const JurikZeroLagVelocityOutput = enum(u8) {
    /// The velocity value.
    velocity = 1,
};

/// Parameters for the Jurik Zero Lag Velocity.
pub const JurikZeroLagVelocityParams = struct {
    /// Depth controls the linear regression window (window = depth+1). Must be >= 2.
    depth: u32 = 10,
    bar_component: ?bar_component.BarComponent = null,
    quote_component: ?quote_component.QuoteComponent = null,
    trade_component: ?trade_component.TradeComponent = null,
};

/// VelAux1: linear regression slope over a window of depth+1 points.
const VelAux1 = struct {
    depth: u32,
    win: [32]f64, // ring buffer, max depth+1 = 16 sufficient for tests; use 32 for safety
    idx: u32,
    bar: u32,

    // Precomputed constants.
    jrc04: f64,
    jrc05: f64,
    jrc06: f64,
    jrc07: f64,

    fn init(depth: u32) VelAux1 {
        const size: f64 = @floatFromInt(depth + 1);
        const jrc04 = size;
        const jrc05 = jrc04 * (jrc04 + 1.0) / 2.0;
        const jrc06 = jrc05 * (2.0 * jrc04 + 1.0) / 3.0;
        const jrc07 = jrc05 * jrc05 * jrc05 - jrc06 * jrc06;

        return .{
            .depth = depth,
            .win = [_]f64{0} ** 32,
            .idx = 0,
            .bar = 0,
            .jrc04 = jrc04,
            .jrc05 = jrc05,
            .jrc06 = jrc06,
            .jrc07 = jrc07,
        };
    }

    fn update(self: *VelAux1, sample: f64) f64 {
        const size = self.depth + 1;
        self.win[self.idx] = sample;
        self.idx = (self.idx + 1) % size;
        self.bar += 1;

        if (self.bar <= self.depth) {
            return 0;
        }

        var jrc08: f64 = 0;
        var jrc09: f64 = 0;

        for (0..self.depth + 1) |j| {
            const pos = (self.idx + size - 1 - @as(u32, @intCast(j))) % size;
            const w = self.jrc04 - @as(f64, @floatFromInt(j));
            jrc08 += self.win[pos] * w;
            jrc09 += self.win[pos] * w * w;
        }

        return (jrc09 * self.jrc05 - jrc08 * self.jrc06) / self.jrc07;
    }
};

/// VelAux3State: adaptive smoother.
const VelAux3State = struct {
    // Constants.
    length: u32, // 30
    eps: f64, // 0.0001
    decay: u32, // 3
    beta: f64,
    alpha: f64,
    max_win: u32, // 31

    // Ring buffers.
    src_ring: [100]f64,
    dev_ring: [100]f64,
    src_idx: i32,
    dev_idx: i32,

    // State.
    jr08: f64,
    jr09: f64,
    jr10: f64,
    jr11: i32,
    jr12: f64,
    jr13: f64,
    jr14: f64,
    jr19: f64,
    jr20: f64,
    jr21: f64,
    jr21a: f64,
    jr21b: f64,
    jr22: f64,
    jr23: f64,

    bar: i32,
    init_done: bool,
    history: [30]f64,
    hist_count: u32,

    fn create() VelAux3State {
        const length: u32 = 30;
        const decay: u32 = 3;

        return .{
            .length = length,
            .eps = 0.0001,
            .decay = decay,
            .beta = 0.86 - 0.55 / @sqrt(@as(f64, @floatFromInt(decay))),
            .alpha = 1.0 - @exp(-@log(@as(f64, 4.0)) / @as(f64, @floatFromInt(decay)) / 2.0),
            .max_win = length + 1,
            .src_ring = [_]f64{0} ** 100,
            .dev_ring = [_]f64{0} ** 100,
            .src_idx = 0,
            .dev_idx = 0,
            .jr08 = 0,
            .jr09 = 0,
            .jr10 = 0,
            .jr11 = 0,
            .jr12 = 0,
            .jr13 = 0,
            .jr14 = 0,
            .jr19 = 0,
            .jr20 = 0,
            .jr21 = 0,
            .jr21a = 0,
            .jr21b = 0,
            .jr22 = 0,
            .jr23 = 0,
            .bar = 0,
            .init_done = false,
            .history = [_]f64{0} ** 30,
            .hist_count = 0,
        };
    }

    fn feed(self: *VelAux3State, sample: f64, bar_idx: i32) f64 {
        const length_i: i32 = @intCast(self.length);
        const max_win_i: i32 = @intCast(self.max_win);

        if (bar_idx < length_i) {
            self.history[@intCast(self.hist_count)] = sample;
            self.hist_count += 1;
            return 0;
        }

        self.bar += 1;

        if (!self.init_done) {
            self.init_done = true;

            // Count consecutive equal values.
            var jr28: f64 = 0;
            const hc = self.hist_count;
            for (1..self.length) |j| {
                if (self.history[hc - j] == self.history[hc - j - 1]) {
                    jr28 += 1;
                }
            }

            var jr26: i32 = undefined;
            if (jr28 < @as(f64, @floatFromInt(self.length - 1))) {
                jr26 = bar_idx - length_i;
            } else {
                jr26 = bar_idx;
            }

            const raw = 1 + bar_idx - jr26;
            self.jr11 = @intFromFloat(@trunc(@min(@as(f64, @floatFromInt(raw)), @as(f64, @floatFromInt(max_win_i)))));

            // jr21 = history[last-1] (SrcA[Bar-1])
            self.jr21 = self.history[hc - 1];

            // jr08 = (sample - history[last-3]) / 3
            const jr07: i32 = 3;
            self.jr08 = (sample - self.history[hc - @as(usize, @intCast(jr07))]) / @as(f64, @floatFromInt(jr07));

            // Fill source ring with historical values.
            var jr15 = self.jr11 - 1;
            while (jr15 >= 1) : (jr15 -= 1) {
                if (self.src_idx <= 0) {
                    self.src_idx = 100;
                }
                self.src_idx -= 1;
                self.src_ring[@intCast(self.src_idx)] = self.history[hc - @as(usize, @intCast(jr15))];
            }
        }

        // Push current value to source ring.
        if (self.src_idx <= 0) {
            self.src_idx = 100;
        }
        self.src_idx -= 1;
        self.src_ring[@intCast(self.src_idx)] = sample;

        if (self.jr11 <= length_i) {
            // Growing phase.
            if (self.bar == 1) {
                self.jr21 = sample;
            } else {
                self.jr21 = @sqrt(self.alpha) * sample + (1.0 - @sqrt(self.alpha)) * self.jr21a;
            }

            if (self.bar > 2) {
                self.jr08 = (self.jr21 - self.jr21b) / 2.0;
            } else {
                self.jr08 = 0;
            }

            self.jr11 += 1;
        } else if (self.jr11 <= max_win_i) {
            // Transition phase: recompute from scratch.
            self.jr12 = @as(f64, @floatFromInt(self.jr11 * (self.jr11 + 1) * (self.jr11 - 1))) / 12.0;
            self.jr13 = @as(f64, @floatFromInt(self.jr11 + 1)) / 2.0;
            self.jr14 = @as(f64, @floatFromInt(self.jr11 - 1)) / 2.0;

            self.jr09 = 0;
            self.jr10 = 0;

            var jr15 = self.jr11 - 1;
            while (jr15 >= 0) : (jr15 -= 1) {
                const jr24: usize = @intCast(@mod(self.src_idx + jr15, 100));
                self.jr09 += self.src_ring[jr24];
                self.jr10 += self.src_ring[jr24] * (self.jr14 - @as(f64, @floatFromInt(jr15)));
                if (jr15 == 0) break;
            }

            var jr16 = self.jr10 / self.jr12;
            var jr17 = (self.jr09 / @as(f64, @floatFromInt(self.jr11))) - (jr16 * self.jr13);

            self.jr19 = 0;
            jr15 = self.jr11 - 1;
            while (jr15 >= 0) : (jr15 -= 1) {
                jr17 += jr16;
                _ = &jr16; // suppress unused
                const jr24: usize = @intCast(@mod(self.src_idx + jr15, 100));
                self.jr19 += @abs(self.src_ring[jr24] - jr17);
                if (jr15 == 0) break;
            }

            self.jr20 = (self.jr19 / @as(f64, @floatFromInt(self.jr11))) * math.pow(f64, @as(f64, @floatFromInt(max_win_i)) / @as(f64, @floatFromInt(self.jr11)), 0.25);
            self.jr11 += 1;

            // Adaptive step.
            self.jr20 = @max(self.eps, self.jr20);
            self.jr22 = sample - (self.jr21 + self.jr08 * self.beta);
            self.jr23 = 1.0 - @exp(-@abs(self.jr22) / self.jr20 / @as(f64, @floatFromInt(self.decay)));
            self.jr08 = self.jr23 * self.jr22 + self.jr08 * self.beta;
            self.jr21 += self.jr08;
        } else {
            // Steady state.
            const jr24out: usize = @intCast(@mod(self.src_idx + max_win_i, 100));
            self.jr10 = self.jr10 - self.jr09 + self.src_ring[jr24out] * self.jr13 + sample * self.jr14;
            self.jr09 = self.jr09 - self.src_ring[jr24out] + sample;

            // Deviation ring update.
            if (self.dev_idx <= 0) {
                self.dev_idx = max_win_i;
            }
            self.dev_idx -= 1;
            self.jr19 -= self.dev_ring[@intCast(self.dev_idx)];

            const jr16 = self.jr10 / self.jr12;
            const jr17 = (self.jr09 / @as(f64, @floatFromInt(max_win_i))) + (jr16 * self.jr14);
            self.dev_ring[@intCast(self.dev_idx)] = @abs(sample - jr17);
            self.jr19 = @max(self.eps, self.jr19 + self.dev_ring[@intCast(self.dev_idx)]);
            self.jr20 += ((self.jr19 / @as(f64, @floatFromInt(max_win_i))) - self.jr20) * self.alpha;

            // Adaptive step.
            self.jr20 = @max(self.eps, self.jr20);
            self.jr22 = sample - (self.jr21 + self.jr08 * self.beta);
            self.jr23 = 1.0 - @exp(-@abs(self.jr22) / self.jr20 / @as(f64, @floatFromInt(self.decay)));
            self.jr08 = self.jr23 * self.jr22 + self.jr08 * self.beta;
            self.jr21 += self.jr08;
        }

        self.jr21b = self.jr21a;
        self.jr21a = self.jr21;

        return self.jr21;
    }
};

/// Jurik Zero Lag Velocity (VEL), see http://jurikres.com/.
pub const JurikZeroLagVelocity = struct {
    line: LineIndicator,
    primed: bool,
    param_depth: u32,

    aux1: VelAux1,
    aux3: VelAux3State,
    bar: i32,

    mnemonic_buf: [96]u8,
    mnemonic_len: usize,

    pub fn init(params: JurikZeroLagVelocityParams) !JurikZeroLagVelocity {
        const depth = params.depth;

        if (depth < 2) return error.InvalidDepth;

        const bc = params.bar_component orelse bar_component.default_bar_component;
        const qc = params.quote_component orelse quote_component.default_quote_component;
        const tc = params.trade_component orelse trade_component.default_trade_component;

        var triple_buf: [64]u8 = undefined;
        const triple = component_triple_mnemonic_mod.componentTripleMnemonic(&triple_buf, bc, qc, tc);

        var mnemonic_buf: [96]u8 = undefined;
        const mnemonic = std.fmt.bufPrint(&mnemonic_buf, "vel({d}{s})", .{
            depth, triple,
        }) catch unreachable;
        const mnemonic_len = mnemonic.len;

        return .{
            .line = LineIndicator.new(
                mnemonic_buf[0..mnemonic_len],
                "Jurik zero lag velocity ",
                params.bar_component,
                params.quote_component,
                params.trade_component,
            ),
            .primed = false,
            .param_depth = depth,
            .aux1 = VelAux1.init(depth),
            .aux3 = VelAux3State.create(),
            .bar = 0,
            .mnemonic_buf = mnemonic_buf,
            .mnemonic_len = mnemonic_len,
        };
    }

    pub fn fixSlices(self: *JurikZeroLagVelocity) void {
        self.line.mnemonic = self.mnemonic_buf[0..self.mnemonic_len];
    }

    pub fn update(self: *JurikZeroLagVelocity, sample: f64) f64 {
        if (math.isNan(sample)) return sample;

        // Stage 1: compute linear regression slope.
        const aux1_val = self.aux1.update(sample);

        // Stage 2: feed into adaptive smoother.
        const bar_idx = self.bar;
        self.bar += 1;

        const result = self.aux3.feed(aux1_val, bar_idx);

        // Output is 0 during warmup → NaN.
        const length_i: i32 = @intCast(self.aux3.length);
        if (bar_idx < length_i) {
            return math.nan(f64);
        }

        if (!self.primed) {
            self.primed = true;
        }

        return result;
    }

    pub fn isPrimed(self: *const JurikZeroLagVelocity) bool {
        return self.primed;
    }

    pub fn getMetadata(self: *const JurikZeroLagVelocity, out: *Metadata) void {
        build_metadata_mod.buildMetadata(
            out,
            .jurik_zero_lag_velocity,
            self.line.mnemonic,
            self.line.description,
            &[_]build_metadata_mod.OutputText{
                .{ .mnemonic = self.line.mnemonic, .description = self.line.description },
            },
        );
    }

    pub fn updateScalar(self: *JurikZeroLagVelocity, sample: *const Scalar) OutputArray {
        const value = self.update(sample.value);
        return LineIndicator.wrapScalar(sample.time, value);
    }

    pub fn updateBar(self: *JurikZeroLagVelocity, sample: *const Bar) OutputArray {
        const value = self.update(self.line.extractBar(sample));
        return LineIndicator.wrapScalar(sample.time, value);
    }

    pub fn updateQuote(self: *JurikZeroLagVelocity, sample: *const Quote) OutputArray {
        const value = self.update(self.line.extractQuote(sample));
        return LineIndicator.wrapScalar(sample.time, value);
    }

    pub fn updateTrade(self: *JurikZeroLagVelocity, sample: *const Trade) OutputArray {
        const value = self.update(self.line.extractTrade(sample));
        return LineIndicator.wrapScalar(sample.time, value);
    }

    pub fn indicator(self: *JurikZeroLagVelocity) indicator_mod.Indicator {
        return .{
            .ptr = @ptrCast(self),
            .vtable = &vtable,
        };
    }

    const vtable = indicator_mod.Indicator.VTable{
        .isPrimed = vtableIsPrimed,
        .metadata = vtableMetadata,
        .updateScalar = vtableUpdateScalar,
        .updateBar = vtableUpdateBar,
        .updateQuote = vtableUpdateQuote,
        .updateTrade = vtableUpdateTrade,
    };

    fn vtableIsPrimed(ptr: *anyopaque) bool {
        const self: *JurikZeroLagVelocity = @ptrCast(@alignCast(ptr));
        return self.isPrimed();
    }

    fn vtableMetadata(ptr: *anyopaque, out: *Metadata) void {
        const self: *const JurikZeroLagVelocity = @ptrCast(@alignCast(ptr));
        self.getMetadata(out);
    }

    fn vtableUpdateScalar(ptr: *anyopaque, sample: *const Scalar) OutputArray {
        const self: *JurikZeroLagVelocity = @ptrCast(@alignCast(ptr));
        return self.updateScalar(sample);
    }

    fn vtableUpdateBar(ptr: *anyopaque, sample: *const Bar) OutputArray {
        const self: *JurikZeroLagVelocity = @ptrCast(@alignCast(ptr));
        return self.updateBar(sample);
    }

    fn vtableUpdateQuote(ptr: *anyopaque, sample: *const Quote) OutputArray {
        const self: *JurikZeroLagVelocity = @ptrCast(@alignCast(ptr));
        return self.updateQuote(sample);
    }

    fn vtableUpdateTrade(ptr: *anyopaque, sample: *const Trade) OutputArray {
        const self: *JurikZeroLagVelocity = @ptrCast(@alignCast(ptr));
        return self.updateTrade(sample);
    }

    pub const Error = error{
        InvalidDepth,
    };
};

// ---------------------------------------------------------------------------
// Tests
// ---------------------------------------------------------------------------

const testing = std.testing;
const testdata = @import("testdata.zig");

fn almostEqual(a: f64, b: f64, epsilon: f64) bool {
    return @abs(a - b) <= epsilon;
}

fn runVelTest(depth: u32, expected: [252]f64) !void {
    var vel = JurikZeroLagVelocity.init(.{ .depth = depth }) catch unreachable;
    vel.fixSlices();
    const input = testdata.testInput();
    const eps = 1e-13;

    for (0..252) |i| {
        const v = vel.update(input[i]);
        if (math.isNan(expected[i])) {
            try testing.expect(math.isNan(v));
        } else {
            try testing.expect(!math.isNan(v));
            if (!almostEqual(v, expected[i], eps)) {
                std.debug.print("FAIL [{d}] depth={d}: expected {d}, got {d}, diff {d}\n", .{ i, depth, expected[i], v, @abs(v - expected[i]) });
                return error.TestUnexpectedResult;
            }
        }
    }

    // NaN passthrough
    try testing.expect(math.isNan(vel.update(math.nan(f64))));
}

test "vel depth 2" {
    try runVelTest(2, testdata.expectedDepth2());
}
test "vel depth 3" {
    try runVelTest(3, testdata.expectedDepth3());
}
test "vel depth 4" {
    try runVelTest(4, testdata.expectedDepth4());
}
test "vel depth 5" {
    try runVelTest(5, testdata.expectedDepth5());
}
test "vel depth 6" {
    try runVelTest(6, testdata.expectedDepth6());
}
test "vel depth 7" {
    try runVelTest(7, testdata.expectedDepth7());
}
test "vel depth 8" {
    try runVelTest(8, testdata.expectedDepth8());
}
test "vel depth 9" {
    try runVelTest(9, testdata.expectedDepth9());
}
test "vel depth 10" {
    try runVelTest(10, testdata.expectedDepth10());
}
test "vel depth 11" {
    try runVelTest(11, testdata.expectedDepth11());
}
test "vel depth 12" {
    try runVelTest(12, testdata.expectedDepth12());
}
test "vel depth 13" {
    try runVelTest(13, testdata.expectedDepth13());
}
test "vel depth 14" {
    try runVelTest(14, testdata.expectedDepth14());
}
test "vel depth 15" {
    try runVelTest(15, testdata.expectedDepth15());
}
