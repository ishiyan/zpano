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

/// Enumerates the outputs of the Jurik Composite Fractal Behavior Index indicator.
pub const JurikCompositeFractalBehaviorIndexOutput = enum(u8) {
    /// The CFB value.
    cfb = 1,
};

/// Parameters for the Jurik Composite Fractal Behavior Index.
pub const JurikCompositeFractalBehaviorIndexParams = struct {
    /// Fractal type (1–4). Controls the maximum fractal depth.
    fractal_type: u32 = 1,
    /// Smoothing window for running averages. Must be >= 1.
    smooth: u32 = 10,
    bar_component: ?bar_component.BarComponent = null,
    quote_component: ?quote_component.QuoteComponent = null,
    trade_component: ?trade_component.TradeComponent = null,
};

// Depth sets for each fractal type (1–4).
const depth_sets = [4][]const u32{
    &[_]u32{ 2, 3, 4, 6, 8, 12, 16, 24 },
    &[_]u32{ 2, 3, 4, 6, 8, 12, 16, 24, 32, 48 },
    &[_]u32{ 2, 3, 4, 6, 8, 12, 16, 24, 32, 48, 64, 96 },
    &[_]u32{ 2, 3, 4, 6, 8, 12, 16, 24, 32, 48, 64, 96, 128, 192 },
};

const weights_even = [7]f64{ 2, 3, 6, 12, 24, 48, 96 };
const weights_odd = [7]f64{ 4, 8, 16, 32, 64, 128, 256 };

/// Streaming state for a single JCFBaux(depth) instance.
const CfbAux = struct {
    depth: u32,
    bar: u32,

    // Ring buffer for IntA values (abs differences). Size = depth.
    int_a: [192]f64,
    int_a_idx: u32,

    // Ring buffer for source values. Size = depth+2.
    src: [194]f64,
    src_idx: u32,

    // Running sums.
    jrc04: f64,
    jrc05: f64,
    jrc06: f64,

    prev_sample: f64,
    first_call: bool,

    fn create(depth: u32) CfbAux {
        return .{
            .depth = depth,
            .bar = 0,
            .int_a = [_]f64{0} ** 192,
            .int_a_idx = 0,
            .src = [_]f64{0} ** 194,
            .src_idx = 0,
            .jrc04 = 0,
            .jrc05 = 0,
            .jrc06 = 0,
            .prev_sample = 0,
            .first_call = true,
        };
    }

    fn update(self: *CfbAux, sample: f64) f64 {
        self.bar += 1;

        const src_size = self.depth + 2;
        self.src[self.src_idx] = sample;
        self.src_idx = (self.src_idx + 1) % src_size;

        if (self.first_call) {
            self.first_call = false;
            self.prev_sample = sample;
            return 0;
        }

        const int_a_val = @abs(sample - self.prev_sample);
        self.prev_sample = sample;

        const old_int_a = self.int_a[self.int_a_idx];
        self.int_a[self.int_a_idx] = int_a_val;
        self.int_a_idx = (self.int_a_idx + 1) % self.depth;

        const ref_bar = self.bar - 1;
        if (ref_bar < self.depth) {
            return 0;
        }

        if (ref_bar <= self.depth * 2) {
            self.jrc04 = 0;
            self.jrc05 = 0;
            self.jrc06 = 0;

            const cur_int_a_pos = (self.int_a_idx + self.depth - 1) % self.depth;
            const cur_src_pos = (self.src_idx + src_size - 1) % src_size;

            for (0..self.depth) |j| {
                const ju: u32 = @intCast(j);
                const int_a_pos = (cur_int_a_pos + self.depth - ju) % self.depth;
                const int_a_v = self.int_a[int_a_pos];

                const src_pos = (cur_src_pos + src_size * 2 - ju - 1) % src_size;
                const src_v = self.src[src_pos];

                self.jrc04 += int_a_v;
                self.jrc05 += @as(f64, @floatFromInt(self.depth - ju)) * int_a_v;
                self.jrc06 += src_v;
            }
        } else {
            self.jrc05 = self.jrc05 - self.jrc04 + int_a_val * @as(f64, @floatFromInt(self.depth));
            self.jrc04 = self.jrc04 - old_int_a + int_a_val;

            const cur_src_pos = (self.src_idx + src_size - 1) % src_size;
            const src_bar_minus_1 = (cur_src_pos + src_size - 1) % src_size;
            const src_bar_minus_depth_minus_1 = (cur_src_pos + src_size - self.depth - 1) % src_size;

            self.jrc06 = self.jrc06 - self.src[src_bar_minus_depth_minus_1] + self.src[src_bar_minus_1];
        }

        const cur_src_pos = (self.src_idx + src_size - 1) % src_size;
        const jrc08 = @abs(@as(f64, @floatFromInt(self.depth)) * self.src[cur_src_pos] - self.jrc06);

        if (self.jrc05 == 0) {
            return 0;
        }

        return jrc08 / self.jrc05;
    }
};

/// Jurik Composite Fractal Behavior Index (CFB), see http://jurikres.com/.
pub const JurikCompositeFractalBehaviorIndex = struct {
    line: LineIndicator,
    primed: bool,
    param_fractal: u32,
    param_smooth: u32,
    num_channels: u32,

    // Aux instances (one per depth channel). Max 14 channels (type 4).
    aux_instances: [14]CfbAux,

    // Sliding window per channel for running average (er23).
    // Each is a ring buffer of size smooth. Max smooth tested = 50.
    aux_windows: [14][64]f64,
    aux_win_idx: u32,

    // Running sums for er23 (the averages).
    er23: [14]f64,

    bar: u32,
    er19: f64,

    mnemonic_buf: [96]u8,
    mnemonic_len: usize,

    pub fn init(params: JurikCompositeFractalBehaviorIndexParams) !JurikCompositeFractalBehaviorIndex {
        const fractal_type = params.fractal_type;
        const smooth = params.smooth;

        if (fractal_type < 1 or fractal_type > 4) return error.InvalidFractalType;
        if (smooth < 1) return error.InvalidSmooth;
        if (smooth > 64) return error.InvalidSmooth;

        const bc = params.bar_component orelse bar_component.default_bar_component;
        const qc = params.quote_component orelse quote_component.default_quote_component;
        const tc = params.trade_component orelse trade_component.default_trade_component;

        var triple_buf: [64]u8 = undefined;
        const triple = component_triple_mnemonic_mod.componentTripleMnemonic(&triple_buf, bc, qc, tc);

        var mnemonic_buf: [96]u8 = undefined;
        const mnemonic = std.fmt.bufPrint(&mnemonic_buf, "jcfb({d},{d}{s})", .{
            fractal_type, smooth, triple,
        }) catch unreachable;
        const mnemonic_len = mnemonic.len;

        const depths = depth_sets[fractal_type - 1];
        const num_ch: u32 = @intCast(depths.len);

        var aux_instances: [14]CfbAux = undefined;
        for (0..num_ch) |i| {
            aux_instances[i] = CfbAux.create(depths[i]);
        }

        var aux_windows: [14][64]f64 = undefined;
        for (0..14) |i| {
            aux_windows[i] = [_]f64{0} ** 64;
        }

        return .{
            .line = LineIndicator.new(
                mnemonic_buf[0..mnemonic_len],
                "Jurik composite fractal behavior index ",
                params.bar_component,
                params.quote_component,
                params.trade_component,
            ),
            .primed = false,
            .param_fractal = fractal_type,
            .param_smooth = smooth,
            .num_channels = num_ch,
            .aux_instances = aux_instances,
            .aux_windows = aux_windows,
            .aux_win_idx = 0,
            .er23 = [_]f64{0} ** 14,
            .bar = 0,
            .er19 = 20,
            .mnemonic_buf = mnemonic_buf,
            .mnemonic_len = mnemonic_len,
        };
    }

    pub fn fixSlices(self: *JurikCompositeFractalBehaviorIndex) void {
        self.line.mnemonic = self.mnemonic_buf[0..self.mnemonic_len];
    }

    pub fn update(self: *JurikCompositeFractalBehaviorIndex, sample: f64) f64 {
        if (math.isNan(sample)) return sample;

        self.bar += 1;

        // Feed all aux instances.
        var aux_values: [14]f64 = undefined;
        for (0..self.num_channels) |i| {
            aux_values[i] = self.aux_instances[i].update(sample);
        }

        // Bar 0 in reference outputs 0.0 → NaN for streaming.
        if (self.bar == 1) {
            return math.nan(f64);
        }

        const ref_bar = self.bar - 1;
        const smooth = self.param_smooth;

        if (ref_bar <= smooth) {
            // Growing window.
            const win_pos = self.aux_win_idx;
            for (0..self.num_channels) |i| {
                self.aux_windows[i][win_pos] = aux_values[i];
            }
            self.aux_win_idx = (self.aux_win_idx + 1) % smooth;

            // Recompute sums from scratch.
            for (0..self.num_channels) |i| {
                var sum: f64 = 0;
                for (0..ref_bar) |j| {
                    const pos = (self.aux_win_idx + smooth * 2 - 1 - @as(u32, @intCast(j))) % smooth;
                    sum += self.aux_windows[i][pos];
                }
                self.er23[i] = sum / @as(f64, @floatFromInt(ref_bar));
            }
        } else {
            // Sliding window.
            const win_pos = self.aux_win_idx;
            for (0..self.num_channels) |i| {
                const old_val = self.aux_windows[i][win_pos];
                self.aux_windows[i][win_pos] = aux_values[i];
                self.er23[i] += (aux_values[i] - old_val) / @as(f64, @floatFromInt(smooth));
            }
            self.aux_win_idx = (self.aux_win_idx + 1) % smooth;
        }

        // Compute weighted composite (only when refBar > 5).
        if (ref_bar > 5) {
            const n = self.num_channels;
            var er22: [14]f64 = [_]f64{0} ** 14;

            // Odd-indexed channels (descending).
            var er15: f64 = 1.0;
            {
                var idx_i: i32 = @as(i32, @intCast(n)) - 1;
                while (idx_i >= 1) : (idx_i -= 2) {
                    const idx: usize = @intCast(idx_i);
                    er22[idx] = er15 * self.er23[idx];
                    er15 *= (1 - er22[idx]);
                }
            }

            // Even-indexed channels (descending).
            var er16: f64 = 1.0;
            {
                var idx_i: i32 = @as(i32, @intCast(n)) - 2;
                while (idx_i >= 0) : (idx_i -= 2) {
                    const idx: usize = @intCast(idx_i);
                    er22[idx] = er16 * self.er23[idx];
                    er16 *= (1 - er22[idx]);
                    if (idx_i == 0) break;
                }
            }

            // Weighted sum.
            var er17: f64 = 0;
            var er18: f64 = 0;

            for (0..n) |idx| {
                const sq = er22[idx] * er22[idx];
                er18 += sq;
                if (idx % 2 == 0) {
                    er17 += sq * weights_even[idx / 2];
                } else {
                    er17 += sq * weights_odd[idx / 2];
                }
            }

            if (er18 == 0) {
                self.er19 = 0;
            } else {
                self.er19 = er17 / er18;
            }
        }

        if (!self.primed) {
            if (ref_bar > 5) {
                self.primed = true;
            }
        }

        return self.er19;
    }

    pub fn isPrimed(self: *const JurikCompositeFractalBehaviorIndex) bool {
        return self.primed;
    }

    pub fn getMetadata(self: *const JurikCompositeFractalBehaviorIndex, out: *Metadata) void {
        build_metadata_mod.buildMetadata(
            out,
            .jurik_composite_fractal_behavior_index,
            self.line.mnemonic,
            self.line.description,
            &[_]build_metadata_mod.OutputText{
                .{ .mnemonic = self.line.mnemonic, .description = self.line.description },
            },
        );
    }

    pub fn updateScalar(self: *JurikCompositeFractalBehaviorIndex, sample: *const Scalar) OutputArray {
        const value = self.update(sample.value);
        return LineIndicator.wrapScalar(sample.time, value);
    }

    pub fn updateBar(self: *JurikCompositeFractalBehaviorIndex, sample: *const Bar) OutputArray {
        const value = self.update(self.line.extractBar(sample));
        return LineIndicator.wrapScalar(sample.time, value);
    }

    pub fn updateQuote(self: *JurikCompositeFractalBehaviorIndex, sample: *const Quote) OutputArray {
        const value = self.update(self.line.extractQuote(sample));
        return LineIndicator.wrapScalar(sample.time, value);
    }

    pub fn updateTrade(self: *JurikCompositeFractalBehaviorIndex, sample: *const Trade) OutputArray {
        const value = self.update(self.line.extractTrade(sample));
        return LineIndicator.wrapScalar(sample.time, value);
    }

    pub fn indicator(self: *JurikCompositeFractalBehaviorIndex) indicator_mod.Indicator {
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
        const self: *JurikCompositeFractalBehaviorIndex = @ptrCast(@alignCast(ptr));
        return self.isPrimed();
    }

    fn vtableMetadata(ptr: *anyopaque, out: *Metadata) void {
        const self: *const JurikCompositeFractalBehaviorIndex = @ptrCast(@alignCast(ptr));
        self.getMetadata(out);
    }

    fn vtableUpdateScalar(ptr: *anyopaque, sample: *const Scalar) OutputArray {
        const self: *JurikCompositeFractalBehaviorIndex = @ptrCast(@alignCast(ptr));
        return self.updateScalar(sample);
    }

    fn vtableUpdateBar(ptr: *anyopaque, sample: *const Bar) OutputArray {
        const self: *JurikCompositeFractalBehaviorIndex = @ptrCast(@alignCast(ptr));
        return self.updateBar(sample);
    }

    fn vtableUpdateQuote(ptr: *anyopaque, sample: *const Quote) OutputArray {
        const self: *JurikCompositeFractalBehaviorIndex = @ptrCast(@alignCast(ptr));
        return self.updateQuote(sample);
    }

    fn vtableUpdateTrade(ptr: *anyopaque, sample: *const Trade) OutputArray {
        const self: *JurikCompositeFractalBehaviorIndex = @ptrCast(@alignCast(ptr));
        return self.updateTrade(sample);
    }

    pub const Error = error{
        InvalidFractalType,
        InvalidSmooth,
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

fn runCfbTest(fractal_type: u32, smooth: u32, expected: [252]f64) !void {
    var cfb = JurikCompositeFractalBehaviorIndex.init(.{
        .fractal_type = fractal_type,
        .smooth = smooth,
    }) catch unreachable;
    cfb.fixSlices();
    const input = testdata.testInput();
    const eps = 1e-13;

    for (0..252) |i| {
        const v = cfb.update(input[i]);

        // Skip last bar: reference aux loop stops at len-2, so last bar's aux values
        // are 0 in reference but computed in streaming.
        if (i == 251) continue;

        if (math.isNan(expected[i])) {
            try testing.expect(math.isNan(v));
        } else {
            try testing.expect(!math.isNan(v));
            if (!almostEqual(v, expected[i], eps)) {
                std.debug.print("FAIL [{d}] type={d} smooth={d}: expected {d}, got {d}, diff {d}\n", .{ i, fractal_type, smooth, expected[i], v, @abs(v - expected[i]) });
                return error.TestUnexpectedResult;
            }
        }
    }

    // NaN passthrough
    try testing.expect(math.isNan(cfb.update(math.nan(f64))));
}

test "jcfb type 1 smooth 2" {
    try runCfbTest(1, 2, testdata.expectedType1Smooth2());
}
test "jcfb type 1 smooth 10" {
    try runCfbTest(1, 10, testdata.expectedType1Smooth10());
}
test "jcfb type 1 smooth 50" {
    try runCfbTest(1, 50, testdata.expectedType1Smooth50());
}
test "jcfb type 2 smooth 2" {
    try runCfbTest(2, 2, testdata.expectedType2Smooth2());
}
test "jcfb type 2 smooth 10" {
    try runCfbTest(2, 10, testdata.expectedType2Smooth10());
}
test "jcfb type 2 smooth 50" {
    try runCfbTest(2, 50, testdata.expectedType2Smooth50());
}
test "jcfb type 3 smooth 2" {
    try runCfbTest(3, 2, testdata.expectedType3Smooth2());
}
test "jcfb type 3 smooth 10" {
    try runCfbTest(3, 10, testdata.expectedType3Smooth10());
}
test "jcfb type 3 smooth 50" {
    try runCfbTest(3, 50, testdata.expectedType3Smooth50());
}
test "jcfb type 4 smooth 2" {
    try runCfbTest(4, 2, testdata.expectedType4Smooth2());
}
test "jcfb type 4 smooth 10" {
    try runCfbTest(4, 10, testdata.expectedType4Smooth10());
}
test "jcfb type 4 smooth 50" {
    try runCfbTest(4, 50, testdata.expectedType4Smooth50());
}
