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

/// Enumerates the outputs of the Jurik Moving Average indicator.
pub const JurikMovingAverageOutput = enum(u8) {
    /// The scalar value of the moving average.
    moving_average = 1,
};

/// Parameters for the Jurik Moving Average.
pub const JurikMovingAverageParams = struct {
    /// Length of the moving average. Must be >= 1.
    length: u32 = 14,
    /// Phase parameter in range [-100, 100].
    phase: i32 = 0,
    bar_component: ?bar_component.BarComponent = null,
    quote_component: ?quote_component.QuoteComponent = null,
    trade_component: ?trade_component.TradeComponent = null,
};

/// Jurik Moving Average (JMA), see http://jurikres.com/.
pub const JurikMovingAverage = struct {
    line: LineIndicator,
    primed: bool,

    // Fixed-size arrays (no allocator needed).
    list: [128]f64,
    ring: [128]f64,
    ring2: [11]f64,
    buffer: [62]f64,

    // Integer state.
    s28: i32,
    s30: i32,
    s38: i32,
    s40: i32,
    s48: i32,
    s50: i32,
    s70: i32,
    f0: i32,
    fD8: i32,
    fF0: i32,
    v5: i32,

    // Float state.
    s8: f64,
    s18: f64,
    f10: f64,
    f18: f64,
    f38: f64,
    f50: f64,
    f58: f64,
    f78: f64,
    f88: f64,
    f90: f64,
    f98: f64,
    fA8: f64,
    fB8: f64,
    fC0: f64,
    fC8: f64,
    fF8: f64,
    v1: f64,
    v2: f64,
    v3: f64,

    mnemonic_buf: [96]u8,
    mnemonic_len: usize,

    pub fn init(params: JurikMovingAverageParams) !JurikMovingAverage {
        const length = params.length;
        const phase = params.phase;

        if (length < 1) return error.InvalidLength;
        if (phase < -100 or phase > 100) return error.InvalidPhase;

        const bc = params.bar_component orelse bar_component.default_bar_component;
        const qc = params.quote_component orelse quote_component.default_quote_component;
        const tc = params.trade_component orelse trade_component.default_trade_component;

        var triple_buf: [64]u8 = undefined;
        const triple = component_triple_mnemonic_mod.componentTripleMnemonic(&triple_buf, bc, qc, tc);

        var mnemonic_buf: [96]u8 = undefined;
        const mnemonic = std.fmt.bufPrint(&mnemonic_buf, "jma({d}, {d}{s})", .{
            length, phase, triple,
        }) catch unreachable;
        const mnemonic_len = mnemonic.len;

        const epsilon: f64 = 1e-10;
        const two: f64 = 2.0;

        var half_len: f64 = epsilon;
        if (length > 1) {
            half_len = (@as(f64, @floatFromInt(length)) - 1.0) / two;
        }

        const f10_val: f64 = @as(f64, @floatFromInt(phase)) / 100.0 + 1.5;

        const v1_val = @log(@sqrt(half_len));
        const v2_val = v1_val;
        const v3_val = @max(v2_val / @log(two) + two, 0.0);

        const f98_val = v3_val;
        const f88_val = @max(f98_val - two, 0.5);

        const f78_val = @sqrt(half_len) * f98_val;
        const f90_val = f78_val / (f78_val + 1.0);
        half_len *= 0.9;
        const f50_val = half_len / (half_len + two);

        // Initialize list: first 64 entries = -1000000, rest = +1000000.
        var list: [128]f64 = undefined;
        for (0..64) |i| {
            list[i] = -1000000.0;
        }
        for (64..128) |i| {
            list[i] = 1000000.0;
        }

        return .{
            .line = LineIndicator.new(
                mnemonic_buf[0..mnemonic_len],
                "Jurik moving average ",
                params.bar_component,
                params.quote_component,
                params.trade_component,
            ),
            .primed = false,
            .list = list,
            .ring = [_]f64{0} ** 128,
            .ring2 = [_]f64{0} ** 11,
            .buffer = [_]f64{0} ** 62,
            .s28 = 63,
            .s30 = 64,
            .s38 = 0,
            .s40 = 0,
            .s48 = 0,
            .s50 = 0,
            .s70 = 0,
            .f0 = 1,
            .fD8 = 0,
            .fF0 = 0,
            .v5 = 0,
            .s8 = 0,
            .s18 = 0,
            .f10 = f10_val,
            .f18 = 0,
            .f38 = 0,
            .f50 = f50_val,
            .f58 = 0,
            .f78 = f78_val,
            .f88 = f88_val,
            .f90 = f90_val,
            .f98 = f98_val,
            .fA8 = 0,
            .fB8 = 0,
            .fC0 = 0,
            .fC8 = 0,
            .fF8 = 0,
            .v1 = v1_val,
            .v2 = v2_val,
            .v3 = v3_val,
            .mnemonic_buf = mnemonic_buf,
            .mnemonic_len = mnemonic_len,
        };
    }

    pub fn fixSlices(self: *JurikMovingAverage) void {
        self.line.mnemonic = self.mnemonic_buf[0..self.mnemonic_len];
    }

    pub fn update(self: *JurikMovingAverage, sample: f64) f64 {
        if (math.isNan(sample)) return sample;

        const epsilon: f64 = 1e-10;

        if (self.fF0 < 61) {
            self.fF0 += 1;
            self.buffer[@intCast(self.fF0)] = sample;
        }

        if (self.fF0 <= 30) {
            return math.nan(f64);
        }

        self.primed = true;

        if (self.f0 == 0) {
            self.fD8 = 0;
        } else {
            self.f0 = 0;
            self.v5 = 0;

            var i: i32 = 1;
            while (i < 30) : (i += 1) {
                if (self.buffer[@intCast(i + 1)] != self.buffer[@intCast(i)]) {
                    self.v5 = 1;
                }
            }

            self.fD8 = self.v5 * 30;
            if (self.fD8 == 0) {
                self.f38 = sample;
            } else {
                self.f38 = self.buffer[1];
            }

            self.f18 = self.f38;
            if (self.fD8 > 29) {
                self.fD8 = 29;
            }
        }

        var ii: i32 = self.fD8;
        while (ii >= 0) : (ii -= 1) {
            var f8: f64 = sample;
            if (ii != 0) {
                f8 = self.buffer[@intCast(31 - ii)];
            }

            const f28 = f8 - self.f18;
            const f48 = f8 - self.f38;
            const a28 = @abs(f28);
            const a48 = @abs(f48);
            self.v2 = @max(a28, a48);

            const fA0 = self.v2;
            const v = fA0 + epsilon;

            if (self.s48 <= 1) {
                self.s48 = 127;
            } else {
                self.s48 -= 1;
            }

            if (self.s50 <= 1) {
                self.s50 = 10;
            } else {
                self.s50 -= 1;
            }

            if (self.s70 < 128) {
                self.s70 += 1;
            }

            self.s8 += v - self.ring2[@intCast(self.s50)];
            self.ring2[@intCast(self.s50)] = v;
            var s20 = self.s8 / @as(f64, @floatFromInt(self.s70));

            if (self.s70 > 10) {
                s20 = self.s8 / 10.0;
            }

            var s58: i32 = undefined;
            var s68: i32 = undefined;

            if (self.s70 > 127) {
                const s10 = self.ring[@intCast(self.s48)];
                self.ring[@intCast(self.s48)] = s20;
                s68 = 64;
                s58 = s68;

                while (s68 > 1) {
                    if (self.list[@intCast(s58)] < s10) {
                        s68 = @divTrunc(s68, 2);
                        s58 += s68;
                    } else if (self.list[@intCast(s58)] <= s10) {
                        s68 = 1;
                    } else {
                        s68 = @divTrunc(s68, 2);
                        s58 -= s68;
                    }
                }
            } else {
                self.ring[@intCast(self.s48)] = s20;
                if (self.s28 + self.s30 > 127) {
                    self.s30 -= 1;
                    s58 = self.s30;
                } else {
                    self.s28 += 1;
                    s58 = self.s28;
                }

                self.s38 = @min(self.s28, 96);
                self.s40 = @max(self.s30, 32);
            }

            s68 = 64;
            var s60: i32 = s68;

            while (s68 > 1) {
                if (self.list[@intCast(s60)] >= s20) {
                    if (self.list[@intCast(s60 - 1)] <= s20) {
                        s68 = 1;
                    } else {
                        s68 = @divTrunc(s68, 2);
                        s60 -= s68;
                    }
                } else {
                    s68 = @divTrunc(s68, 2);
                    s60 += s68;
                }

                if (s60 == 127 and s20 > self.list[127]) {
                    s60 = 128;
                }
            }

            if (self.s70 > 127) {
                if (s58 >= s60) {
                    if (self.s38 + 1 > s60 and self.s40 - 1 < s60) {
                        self.s18 += s20;
                    } else if (self.s40 > s60 and self.s40 - 1 < s58) {
                        self.s18 += self.list[@intCast(self.s40 - 1)];
                    }
                } else if (self.s40 >= s60) {
                    if (self.s38 + 1 < s60 and self.s38 + 1 > s58) {
                        self.s18 += self.list[@intCast(self.s38 + 1)];
                    }
                } else if (self.s38 + 2 > s60) {
                    self.s18 += s20;
                } else if (self.s38 + 1 < s60 and self.s38 + 1 > s58) {
                    self.s18 += self.list[@intCast(self.s38 + 1)];
                }

                if (s58 > s60) {
                    if (self.s40 - 1 < s58 and self.s38 + 1 > s58) {
                        self.s18 -= self.list[@intCast(s58)];
                    } else if (self.s38 < s58 and self.s38 + 1 > s60) {
                        self.s18 -= self.list[@intCast(self.s38)];
                    }
                } else {
                    if (self.s38 + 1 > s58 and self.s40 - 1 < s58) {
                        self.s18 -= self.list[@intCast(s58)];
                    } else if (self.s40 > s58 and self.s40 < s60) {
                        self.s18 -= self.list[@intCast(self.s40)];
                    }
                }
            }

            if (s58 <= s60) {
                if (s58 >= s60) {
                    self.list[@intCast(s60)] = s20;
                } else {
                    var k: i32 = s58 + 1;
                    while (k <= s60 - 1) : (k += 1) {
                        self.list[@intCast(k - 1)] = self.list[@intCast(k)];
                    }
                    self.list[@intCast(s60 - 1)] = s20;
                }
            } else {
                var k: i32 = s58 - 1;
                while (k >= s60) : (k -= 1) {
                    self.list[@intCast(k + 1)] = self.list[@intCast(k)];
                }
                self.list[@intCast(s60)] = s20;
            }

            if (self.s70 < 128) {
                self.s18 = 0;
                var k: i32 = self.s40;
                while (k <= self.s38) : (k += 1) {
                    self.s18 += self.list[@intCast(k)];
                }
            }

            const f60 = self.s18 / @as(f64, @floatFromInt(self.s38 - self.s40 + 1));

            if (self.fF8 + 1 > 31) {
                self.fF8 = 31;
            } else {
                self.fF8 += 1;
            }

            if (self.fF8 <= 30) {
                if (f28 > 0) {
                    self.f18 = f8;
                } else {
                    self.f18 = f8 - f28 * self.f90;
                }

                if (f48 < 0) {
                    self.f38 = f8;
                } else {
                    self.f38 = f8 - f48 * self.f90;
                }

                self.fB8 = sample;
                if (self.fF8 != 30) {
                    continue;
                }

                var v4: i32 = 1;
                self.fC0 = sample;

                if (@ceil(self.f78) >= 1.0) {
                    v4 = @intFromFloat(@ceil(self.f78));
                }

                var v2_local: i32 = 1;
                const fE8 = v4;

                if (@floor(self.f78) >= 1.0) {
                    v2_local = @intFromFloat(@floor(self.f78));
                }

                var f68: f64 = 1.0;
                const fE0 = v2_local;

                if (fE8 != fE0) {
                    v4 = fE8 - fE0;
                    f68 = (self.f78 - @as(f64, @floatFromInt(fE0))) / @as(f64, @floatFromInt(v4));
                }

                const v5_local: i32 = @min(fE0, 29);
                const v6_local: i32 = @min(fE8, 29);
                self.fA8 = (sample - self.buffer[@intCast(self.fF0 - v5_local)]) * (1.0 - f68) / @as(f64, @floatFromInt(fE0)) +
                    (sample - self.buffer[@intCast(self.fF0 - v6_local)]) * f68 / @as(f64, @floatFromInt(fE8));
            } else {
                const p = math.pow(f64, fA0 / f60, self.f88);
                self.v1 = @min(self.f98, p);

                if (self.v1 < 1.0) {
                    self.v2 = 1;
                } else {
                    self.v3 = @min(self.f98, p);
                    self.v2 = self.v3;
                }

                self.f58 = self.v2;
                const f70 = math.pow(f64, self.f90, @sqrt(self.f58));

                if (f28 > 0) {
                    self.f18 = f8;
                } else {
                    self.f18 = f8 - f28 * f70;
                }

                if (f48 < 0) {
                    self.f38 = f8;
                } else {
                    self.f38 = f8 - f48 * f70;
                }
            }
        }

        if (self.fF8 > 30) {
            const f30 = math.pow(f64, self.f50, self.f58);
            self.fC0 = (1.0 - f30) * sample + f30 * self.fC0;
            self.fC8 = (sample - self.fC0) * (1.0 - self.f50) + self.f50 * self.fC8;
            const fD0 = self.f10 * self.fC8 + self.fC0;
            const f20 = f30 * -2.0;
            const f40 = f30 * f30;
            const fB0 = f20 + f40 + 1.0;
            self.fA8 = (fD0 - self.fB8) * fB0 + f40 * self.fA8;
            self.fB8 += self.fA8;
        }

        return self.fB8;
    }

    pub fn isPrimed(self: *const JurikMovingAverage) bool {
        return self.primed;
    }

    pub fn getMetadata(self: *const JurikMovingAverage, out: *Metadata) void {
        build_metadata_mod.buildMetadata(
            out,
            .jurik_moving_average,
            self.line.mnemonic,
            self.line.description,
            &[_]build_metadata_mod.OutputText{
                .{ .mnemonic = self.line.mnemonic, .description = self.line.description },
            },
        );
    }

    pub fn updateScalar(self: *JurikMovingAverage, sample: *const Scalar) OutputArray {
        const value = self.update(sample.value);
        return LineIndicator.wrapScalar(sample.time, value);
    }

    pub fn updateBar(self: *JurikMovingAverage, sample: *const Bar) OutputArray {
        const value = self.update(self.line.extractBar(sample));
        return LineIndicator.wrapScalar(sample.time, value);
    }

    pub fn updateQuote(self: *JurikMovingAverage, sample: *const Quote) OutputArray {
        const value = self.update(self.line.extractQuote(sample));
        return LineIndicator.wrapScalar(sample.time, value);
    }

    pub fn updateTrade(self: *JurikMovingAverage, sample: *const Trade) OutputArray {
        const value = self.update(self.line.extractTrade(sample));
        return LineIndicator.wrapScalar(sample.time, value);
    }

    pub fn indicator(self: *JurikMovingAverage) indicator_mod.Indicator {
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
        const self: *JurikMovingAverage = @ptrCast(@alignCast(ptr));
        return self.isPrimed();
    }

    fn vtableMetadata(ptr: *anyopaque, out: *Metadata) void {
        const self: *const JurikMovingAverage = @ptrCast(@alignCast(ptr));
        self.getMetadata(out);
    }

    fn vtableUpdateScalar(ptr: *anyopaque, sample: *const Scalar) OutputArray {
        const self: *JurikMovingAverage = @ptrCast(@alignCast(ptr));
        return self.updateScalar(sample);
    }

    fn vtableUpdateBar(ptr: *anyopaque, sample: *const Bar) OutputArray {
        const self: *JurikMovingAverage = @ptrCast(@alignCast(ptr));
        return self.updateBar(sample);
    }

    fn vtableUpdateQuote(ptr: *anyopaque, sample: *const Quote) OutputArray {
        const self: *JurikMovingAverage = @ptrCast(@alignCast(ptr));
        return self.updateQuote(sample);
    }

    fn vtableUpdateTrade(ptr: *anyopaque, sample: *const Trade) OutputArray {
        const self: *JurikMovingAverage = @ptrCast(@alignCast(ptr));
        return self.updateTrade(sample);
    }

    pub const Error = error{
        InvalidLength,
        InvalidPhase,
    };
};

// ---------------------------------------------------------------------------
// Tests
// ---------------------------------------------------------------------------

const testing = std.testing;
const testdata = @import("testdata.zig");

fn testInput() [252]f64 {
    return testdata.testInput();
}

fn almostEqual(a: f64, b: f64, epsilon: f64) bool {
    return @abs(a - b) <= epsilon;
}

fn createJma(length: u32, phase: i32) JurikMovingAverage {
    var jma = JurikMovingAverage.init(.{
        .length = length,
        .phase = phase,
    }) catch unreachable;
    jma.fixSlices();
    return jma;
}

fn runJmaTest(length: u32, phase: i32, expected: [252]f64) !void {
    var jma = createJma(length, phase);
    const input = testInput();
    const eps = 1e-13;

    for (0..30) |i| {
        const v = jma.update(input[i]);
        try testing.expect(math.isNan(v));
    }

    for (30..252) |i| {
        const v = jma.update(input[i]);
        if (math.isNan(expected[i])) {
            try testing.expect(math.isNan(v));
        } else {
            try testing.expect(!math.isNan(v));
            if (!almostEqual(v, expected[i], eps)) {
                std.debug.print("FAIL [{d}]: expected {d}, got {d}, diff {d}\n", .{ i, expected[i], v, @abs(v - expected[i]) });
                return error.TestUnexpectedResult;
            }
        }
    }

    // NaN passthrough
    try testing.expect(math.isNan(jma.update(math.nan(f64))));
}


test "jurik moving average length 20 phase -30" {
    try runJmaTest(20, -30, testdata.expectedLen20PhaseMin30());
}

test "jurik moving average length 20 phase 30" {
    try runJmaTest(20, 30, testdata.expectedLen20Phase30());
}

test "jurik moving average length 2 phase 1" {
    try runJmaTest(2, 1, testdata.expectedLen2Phase1());
}

test "jurik moving average length 5 phase 1" {
    try runJmaTest(5, 1, testdata.expectedLen5Phase1());
}

test "jurik moving average length 25 phase 1" {
    try runJmaTest(25, 1, testdata.expectedLen25Phase1());
}

test "jurik moving average length 50 phase 1" {
    try runJmaTest(50, 1, testdata.expectedLen50Phase1());
}

test "jurik moving average length 75 phase 1" {
    try runJmaTest(75, 1, testdata.expectedLen75Phase1());
}

test "jurik moving average length 100 phase 1" {
    try runJmaTest(100, 1, testdata.expectedLen100Phase1());
}

test "jurik moving average length 20 phase -100" {
    try runJmaTest(20, -100, testdata.expectedLen20PhaseMin100());
}

test "jurik moving average length 20 phase 0" {
    try runJmaTest(20, 0, testdata.expectedLen20Phase0());
}

test "jurik moving average length 20 phase 100" {
    try runJmaTest(20, 100, testdata.expectedLen20Phase100());
}

test "jurik moving average length 10 phase 1" {
    try runJmaTest(10, 1, testdata.expectedLen10Phase1());
}

test "jurik moving average is primed" {
    var jma = createJma(10, 30);
    const input = testInput();

    try testing.expect(!jma.isPrimed());

    for (0..30) |i| {
        _ = jma.update(input[i]);
        try testing.expect(!jma.isPrimed());
    }

    _ = jma.update(input[30]);
    try testing.expect(jma.isPrimed());
}

test "jurik moving average metadata" {
    var jma = createJma(10, 30);

    var m: Metadata = undefined;
    jma.getMetadata(&m);

    try testing.expectEqual(@import("../../core/identifier.zig").Identifier.jurik_moving_average, m.identifier);
    try testing.expectEqual(@as(usize, 1), m.outputs_len);
    try testing.expectEqual(@as(i32, 1), m.outputs_buf[0].kind);
    try testing.expectEqualStrings("jma(10, 30)", m.mnemonic);
}

test "jurik moving average update scalar" {
    var jma = createJma(10, 30);

    for (0..30) |_| {
        _ = jma.update(3.0);
    }

    const s = Scalar{ .time = 1617235200, .value = 3.0 };
    const out = jma.updateScalar(&s);
    const slice = out.slice();
    try testing.expectEqual(@as(usize, 1), slice.len);
    try testing.expectEqual(@as(i64, 1617235200), slice[0].scalar.time);
}

test "jurik moving average update bar" {
    var jma = createJma(10, 30);

    for (0..30) |_| {
        _ = jma.update(3.0);
    }

    const bar = Bar{ .time = 1617235200, .open = 0, .high = 0, .low = 0, .close = 3.0, .volume = 0 };
    const out = jma.updateBar(&bar);
    const slice = out.slice();
    try testing.expectEqual(@as(usize, 1), slice.len);
}

test "jurik moving average invalid params" {
    // Length < 1
    try testing.expectError(error.InvalidLength, JurikMovingAverage.init(.{
        .length = 0,
    }));

    // Phase out of range
    try testing.expectError(error.InvalidPhase, JurikMovingAverage.init(.{
        .phase = -101,
    }));
    try testing.expectError(error.InvalidPhase, JurikMovingAverage.init(.{
        .phase = 101,
    }));
}
