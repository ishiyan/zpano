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

/// Enumerates the outputs of the Jurik Fractal Adaptive Zero Lag Velocity indicator.
pub const JurikFractalAdaptiveZeroLagVelocityOutput = enum(u8) {
    /// The velocity value.
    value = 1,
};

/// Parameters for the Jurik Fractal Adaptive Zero Lag Velocity.
pub const JurikFractalAdaptiveZeroLagVelocityParams = struct {
    /// Minimum adaptive depth. Must be >= 2.
    lo_depth: u32 = 5,
    /// Maximum adaptive depth. Must be >= lo_depth.
    hi_depth: u32 = 30,
    /// Fractal type (1-4).
    fractal_type: u32 = 1,
    /// Smoothing parameter for CFB. Must be >= 1.
    smooth: u32 = 10,
    bar_component: ?bar_component.BarComponent = null,
    quote_component: ?quote_component.QuoteComponent = null,
    trade_component: ?trade_component.TradeComponent = null,
};

const MAX_BARS = 1024;
const SMOOTH_BUFFER_SIZE = 1001;
const MAX_CFB_CHANNELS = 14;
const MAX_CFB_DEPTH = 192;
const MAX_CFB_SMOOTH = 256;

// Scale sets for different fractal types.
const scale_set_1 = [_]u32{ 2, 3, 4, 6, 8, 12, 16, 24 };
const scale_set_2 = [_]u32{ 2, 3, 4, 6, 8, 12, 16, 24, 32, 48 };
const scale_set_3 = [_]u32{ 2, 3, 4, 6, 8, 12, 16, 24, 32, 48, 64, 96 };
const scale_set_4 = [_]u32{ 2, 3, 4, 6, 8, 12, 16, 24, 32, 48, 64, 96, 128, 192 };

const weights_even = [_]f64{ 2, 3, 6, 12, 24, 48, 96 };
const weights_odd = [_]f64{ 4, 8, 16, 32, 64, 128, 256 };

fn getScaleSet(fractal_type: u32) []const u32 {
    return switch (fractal_type) {
        1 => &scale_set_1,
        2 => &scale_set_2,
        3 => &scale_set_3,
        4 => &scale_set_4,
        else => &scale_set_1,
    };
}

/// Streaming state for a single CFB channel at one depth.
const CfbAux = struct {
    depth: u32,
    bar: u32,
    int_a: [MAX_CFB_DEPTH]f64,
    int_a_idx: u32,
    src: [MAX_CFB_DEPTH + 2]f64,
    src_idx: u32,
    jrc04: f64,
    jrc05: f64,
    jrc06: f64,
    prev_sample: f64,
    first_call: bool,

    fn init(depth: u32) CfbAux {
        return .{
            .depth = depth,
            .bar = 0,
            .int_a = [_]f64{0} ** MAX_CFB_DEPTH,
            .int_a_idx = 0,
            .src = [_]f64{0} ** (MAX_CFB_DEPTH + 2),
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
        const depth = self.depth;
        const src_size = depth + 2;

        self.src[self.src_idx] = sample;
        self.src_idx = (self.src_idx + 1) % src_size;

        if (self.first_call) {
            self.first_call = false;
            self.prev_sample = sample;
            return 0.0;
        }

        const int_a_val = @abs(sample - self.prev_sample);
        self.prev_sample = sample;

        const old_int_a = self.int_a[self.int_a_idx];
        self.int_a[self.int_a_idx] = int_a_val;
        self.int_a_idx = (self.int_a_idx + 1) % depth;

        const ref_bar = self.bar - 1;
        if (ref_bar < depth) {
            return 0.0;
        }

        if (ref_bar <= depth * 2) {
            // Recompute from scratch.
            self.jrc04 = 0.0;
            self.jrc05 = 0.0;
            self.jrc06 = 0.0;

            const cur_int_a_pos = (self.int_a_idx + depth - 1) % depth;
            const cur_src_pos = (self.src_idx + src_size - 1) % src_size;

            for (0..depth) |j| {
                const ji: u32 = @intCast(j);
                const int_a_pos = (cur_int_a_pos + depth - ji) % depth;
                const src_pos = (cur_src_pos + src_size * 2 - ji - 1) % src_size;

                self.jrc04 += self.int_a[int_a_pos];
                self.jrc05 += @as(f64, @floatFromInt(depth - ji)) * self.int_a[int_a_pos];
                self.jrc06 += self.src[src_pos];
            }
        } else {
            // Incremental update.
            self.jrc05 = self.jrc05 - self.jrc04 + int_a_val * @as(f64, @floatFromInt(depth));
            self.jrc04 = self.jrc04 - old_int_a + int_a_val;

            const cur_src_pos = (self.src_idx + src_size - 1) % src_size;
            const src_bar_minus1 = (cur_src_pos + src_size - 1) % src_size;
            const src_bar_minus_depth_minus1 = (cur_src_pos + src_size * 2 - depth - 1) % src_size;

            self.jrc06 = self.jrc06 - self.src[src_bar_minus_depth_minus1] + self.src[src_bar_minus1];
        }

        const cur_src_pos = (self.src_idx + src_size - 1) % src_size;
        const jrc08 = @abs(@as(f64, @floatFromInt(depth)) * self.src[cur_src_pos] - self.jrc06);

        if (self.jrc05 == 0.0) {
            return 0.0;
        }

        return jrc08 / self.jrc05;
    }
};

/// Composite Fractal Behavior weighted dominant cycle.
const Cfb = struct {
    num_channels: u32,
    scales: []const u32,
    auxs: [MAX_CFB_CHANNELS]CfbAux,
    aux_windows: [MAX_CFB_CHANNELS][MAX_CFB_SMOOTH]f64,
    aux_win_idx: u32,
    er23: [MAX_CFB_CHANNELS]f64,
    smooth: u32,
    bar: u32,
    cfb_value: f64,

    fn init(fractal_type: u32, smooth: u32) Cfb {
        const scales = getScaleSet(fractal_type);
        const n: u32 = @intCast(scales.len);

        var result: Cfb = .{
            .num_channels = n,
            .scales = scales,
            .auxs = undefined,
            .aux_windows = [_][MAX_CFB_SMOOTH]f64{[_]f64{0} ** MAX_CFB_SMOOTH} ** MAX_CFB_CHANNELS,
            .aux_win_idx = 0,
            .er23 = [_]f64{0} ** MAX_CFB_CHANNELS,
            .smooth = smooth,
            .bar = 0,
            .cfb_value = 0,
        };

        for (0..n) |i| {
            result.auxs[i] = CfbAux.init(scales[i]);
        }

        return result;
    }

    fn update(self: *Cfb, sample: f64) f64 {
        self.bar += 1;
        const ref_bar = self.bar - 1;
        const n = self.num_channels;
        const smooth = self.smooth;

        var aux_values: [MAX_CFB_CHANNELS]f64 = undefined;
        for (0..n) |i| {
            aux_values[i] = self.auxs[i].update(sample);
        }

        if (ref_bar == 0) {
            return 0.0;
        }

        if (ref_bar <= smooth) {
            const win_pos = self.aux_win_idx;
            for (0..n) |i| {
                self.aux_windows[i][win_pos] = aux_values[i];
            }
            self.aux_win_idx = (self.aux_win_idx + 1) % smooth;

            for (0..n) |i| {
                var s: f64 = 0;
                for (0..ref_bar) |j| {
                    const pos = (self.aux_win_idx + smooth * 2 - 1 - @as(u32, @intCast(j))) % smooth;
                    s += self.aux_windows[i][pos];
                }
                self.er23[i] = s / @as(f64, @floatFromInt(ref_bar));
            }
        } else {
            const win_pos = self.aux_win_idx;
            for (0..n) |i| {
                const old_val = self.aux_windows[i][win_pos];
                self.aux_windows[i][win_pos] = aux_values[i];
                self.er23[i] += (aux_values[i] - old_val) / @as(f64, @floatFromInt(smooth));
            }
            self.aux_win_idx = (self.aux_win_idx + 1) % smooth;
        }

        if (ref_bar > 5) {
            var er22: [MAX_CFB_CHANNELS]f64 = [_]f64{0} ** MAX_CFB_CHANNELS;

            // Odd-indexed channels (descending).
            var er15: f64 = 1.0;
            {
                var idx: i32 = @as(i32, @intCast(n)) - 1;
                while (idx >= 1) : (idx -= 2) {
                    const uidx: usize = @intCast(idx);
                    er22[uidx] = er15 * self.er23[uidx];
                    er15 *= (1.0 - er22[uidx]);
                }
            }

            // Even-indexed channels (descending).
            var er16: f64 = 1.0;
            {
                var idx: i32 = @as(i32, @intCast(n)) - 2;
                while (idx >= 0) : (idx -= 2) {
                    const uidx: usize = @intCast(idx);
                    er22[uidx] = er16 * self.er23[uidx];
                    er16 *= (1.0 - er22[uidx]);
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

            if (er18 == 0.0) {
                self.cfb_value = 0.0;
            } else {
                self.cfb_value = er17 / er18;
            }
        }

        return self.cfb_value;
    }
};

/// Adaptive smoother (Stage 2) with fixed period=3.0.
const VelSmooth = struct {
    jrc03: f64,
    jrc06: u32,
    jrc07: u32,
    ema_factor: f64,
    damping: f64,
    eps2: f64,
    buffer: [SMOOTH_BUFFER_SIZE]f64,
    idx: u32,
    length: u32,
    velocity: f64,
    position: f64,
    smoothed_mad: f64,
    mad_init: bool,
    initialized: bool,

    fn init(period: f64) VelSmooth {
        const eps2: f64 = 0.0001;
        const jrc03 = @min(500.0, @max(eps2, period));
        const jrc06: u32 = @intFromFloat(@max(31.0, @ceil(2.0 * period)));
        const jrc07: u32 = @intFromFloat(@min(30.0, @ceil(period)));
        _ = jrc07;
        const ema_factor = 1.0 - @exp(-@log(4.0) / (period / 2.0));
        const damping = 0.86 - 0.55 / @sqrt(jrc03);

        return .{
            .jrc03 = jrc03,
            .jrc06 = jrc06,
            .jrc07 = 0,
            .ema_factor = ema_factor,
            .damping = damping,
            .eps2 = eps2,
            .buffer = [_]f64{0} ** SMOOTH_BUFFER_SIZE,
            .idx = 0,
            .length = 0,
            .velocity = 0,
            .position = 0,
            .smoothed_mad = 0,
            .mad_init = false,
            .initialized = false,
        };
    }

    fn update(self: *VelSmooth, value: f64) f64 {
        self.buffer[self.idx] = value;
        self.idx = (self.idx + 1) % SMOOTH_BUFFER_SIZE;
        self.length += 1;
        if (self.length > SMOOTH_BUFFER_SIZE) {
            self.length = SMOOTH_BUFFER_SIZE;
        }

        const length = self.length;

        if (!self.initialized) {
            self.initialized = true;
            self.position = value;
            self.velocity = 0.0;
            self.smoothed_mad = 0.0;
            return self.position;
        }

        // Linear regression over capped window.
        var n = length;
        if (n > self.jrc06) {
            n = self.jrc06;
        }

        var sx: f64 = 0;
        var sy: f64 = 0;
        var sxy: f64 = 0;
        var sx2: f64 = 0;

        for (0..n) |i| {
            const buf_idx = (self.idx + SMOOTH_BUFFER_SIZE - 1 - @as(u32, @intCast(i))) % SMOOTH_BUFFER_SIZE;
            const x: f64 = @floatFromInt(i);
            const y = self.buffer[buf_idx];
            sx += x;
            sy += y;
            sxy += x * y;
            sx2 += x * x;
        }

        const fn_val: f64 = @floatFromInt(n);
        var slope: f64 = 0;
        if (n > 1) {
            slope = (fn_val * sxy - sx * sy) / (fn_val * sx2 - sx * sx);
        }

        const intercept = (sy - slope * sx) / fn_val;

        // MAD from regression residuals.
        var mad: f64 = 0;
        for (0..n) |i| {
            const buf_idx = (self.idx + SMOOTH_BUFFER_SIZE - 1 - @as(u32, @intCast(i))) % SMOOTH_BUFFER_SIZE;
            const predicted = intercept + slope * @as(f64, @floatFromInt(i));
            mad += @abs(self.buffer[buf_idx] - predicted);
        }
        mad /= fn_val;

        // Scale MAD.
        const scaled_mad = mad * 1.2 * std.math.pow(f64, @as(f64, @floatFromInt(self.jrc06)) / fn_val, 0.25);

        // Smooth MAD with EMA.
        if (!self.mad_init) {
            self.smoothed_mad = scaled_mad;
            if (scaled_mad > 0) {
                self.mad_init = true;
            }
        } else {
            self.smoothed_mad += (scaled_mad - self.smoothed_mad) * self.ema_factor;
        }

        const smoothed_mad = @max(self.eps2, self.smoothed_mad);

        // Adaptive velocity/position dynamics.
        const prediction_error = value - self.position;
        const response_factor = 1.0 - @exp(-@abs(prediction_error) / (smoothed_mad * self.jrc03));
        self.velocity = response_factor * prediction_error + self.velocity * self.damping;
        self.position += self.velocity;

        return self.position;
    }
};

/// Jurik Fractal Adaptive Zero Lag Velocity (JVELCFB) indicator.
pub const JurikFractalAdaptiveZeroLagVelocity = struct {
    line: LineIndicator,
    primed: bool,
    lo_depth: u32,
    hi_depth: u32,

    prices: [MAX_BARS]f64,
    bar_count: u32,
    cfb_inst: Cfb,
    cfb_min: ?f64,
    cfb_max: ?f64,
    smooth_inst: VelSmooth,

    mnemonic_buf: [128]u8,
    mnemonic_len: usize,

    pub fn init(params: JurikFractalAdaptiveZeroLagVelocityParams) !JurikFractalAdaptiveZeroLagVelocity {
        const lo_depth = params.lo_depth;
        const hi_depth = params.hi_depth;
        const fractal_type = params.fractal_type;
        const smooth = params.smooth;

        if (lo_depth < 2) return error.InvalidLoDepth;
        if (hi_depth < lo_depth) return error.InvalidHiDepth;
        if (fractal_type < 1 or fractal_type > 4) return error.InvalidFractalType;
        if (smooth < 1) return error.InvalidSmooth;

        const bc = params.bar_component orelse bar_component.default_bar_component;
        const qc = params.quote_component orelse quote_component.default_quote_component;
        const tc = params.trade_component orelse trade_component.default_trade_component;

        var triple_buf: [64]u8 = undefined;
        const triple = component_triple_mnemonic_mod.componentTripleMnemonic(&triple_buf, bc, qc, tc);

        var mnemonic_buf: [128]u8 = undefined;
        const mnemonic = std.fmt.bufPrint(&mnemonic_buf, "jvelcfb({d}, {d}, {d}, {d}{s})", .{
            lo_depth, hi_depth, fractal_type, smooth, triple,
        }) catch unreachable;
        const mnemonic_len = mnemonic.len;

        return .{
            .line = LineIndicator.new(
                mnemonic_buf[0..mnemonic_len],
                "Jurik fractal adaptive zero lag velocity ",
                params.bar_component,
                params.quote_component,
                params.trade_component,
            ),
            .primed = false,
            .lo_depth = lo_depth,
            .hi_depth = hi_depth,
            .prices = [_]f64{0} ** MAX_BARS,
            .bar_count = 0,
            .cfb_inst = Cfb.init(fractal_type, smooth),
            .cfb_min = null,
            .cfb_max = null,
            .smooth_inst = VelSmooth.init(3.0),
            .mnemonic_buf = mnemonic_buf,
            .mnemonic_len = mnemonic_len,
        };
    }

    pub fn fixSlices(self: *JurikFractalAdaptiveZeroLagVelocity) void {
        self.line.mnemonic = self.mnemonic_buf[0..self.mnemonic_len];
    }

    pub fn update(self: *JurikFractalAdaptiveZeroLagVelocity, sample: f64) f64 {
        if (math.isNan(sample)) return sample;

        const bar = self.bar_count;
        self.bar_count += 1;

        self.prices[bar] = sample;

        // CFB computation.
        const cfb_val = self.cfb_inst.update(sample);

        if (bar == 0) {
            return math.nan(f64);
        }

        // Stochastic normalization.
        if (self.cfb_min == null) {
            self.cfb_min = cfb_val;
            self.cfb_max = cfb_val;
        } else {
            if (cfb_val < self.cfb_min.?) {
                self.cfb_min = cfb_val;
            }
            if (cfb_val > self.cfb_max.?) {
                self.cfb_max = cfb_val;
            }
        }

        const cfb_range = self.cfb_max.? - self.cfb_min.?;
        var sr: f64 = undefined;
        if (cfb_range != 0.0) {
            sr = (cfb_val - self.cfb_min.?) / cfb_range;
        } else {
            sr = 0.5;
        }

        const depth_f = @as(f64, @floatFromInt(self.lo_depth)) + sr * @as(f64, @floatFromInt(self.hi_depth - self.lo_depth));
        const depth: u32 = @intFromFloat(@round(depth_f));

        // Stage 1: WLS slope.
        if (bar < depth) {
            return math.nan(f64);
        }

        const n: f64 = @floatFromInt(depth + 1);
        const s1 = n * (n + 1.0) / 2.0;
        const s2 = s1 * (2.0 * n + 1.0) / 3.0;
        const denom = s1 * s1 * s1 - s2 * s2;

        var sum_xw: f64 = 0;
        var sum_xw2: f64 = 0;

        for (0..depth + 1) |i| {
            const w = n - @as(f64, @floatFromInt(i));
            const p = self.prices[bar - @as(u32, @intCast(i))];
            sum_xw += p * w;
            sum_xw2 += p * w * w;
        }

        const slope = (sum_xw2 * s1 - sum_xw * s2) / denom;

        // Stage 2: adaptive smoother.
        const result = self.smooth_inst.update(slope);

        if (!self.primed) {
            self.primed = true;
        }

        return result;
    }

    pub fn isPrimed(self: *const JurikFractalAdaptiveZeroLagVelocity) bool {
        return self.primed;
    }

    pub fn getMetadata(self: *const JurikFractalAdaptiveZeroLagVelocity, out: *Metadata) void {
        build_metadata_mod.buildMetadata(
            out,
            .jurik_fractal_adaptive_zero_lag_velocity,
            self.line.mnemonic,
            self.line.description,
            &[_]build_metadata_mod.OutputText{
                .{ .mnemonic = self.line.mnemonic, .description = self.line.description },
            },
        );
    }

    pub fn updateScalar(self: *JurikFractalAdaptiveZeroLagVelocity, sample: *const Scalar) OutputArray {
        const value = self.update(sample.value);
        return LineIndicator.wrapScalar(sample.time, value);
    }

    pub fn updateBar(self: *JurikFractalAdaptiveZeroLagVelocity, sample: *const Bar) OutputArray {
        const value = self.update(self.line.extractBar(sample));
        return LineIndicator.wrapScalar(sample.time, value);
    }

    pub fn updateQuote(self: *JurikFractalAdaptiveZeroLagVelocity, sample: *const Quote) OutputArray {
        const value = self.update(self.line.extractQuote(sample));
        return LineIndicator.wrapScalar(sample.time, value);
    }

    pub fn updateTrade(self: *JurikFractalAdaptiveZeroLagVelocity, sample: *const Trade) OutputArray {
        const value = self.update(self.line.extractTrade(sample));
        return LineIndicator.wrapScalar(sample.time, value);
    }

    pub fn indicator(self: *JurikFractalAdaptiveZeroLagVelocity) indicator_mod.Indicator {
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
        const self: *JurikFractalAdaptiveZeroLagVelocity = @ptrCast(@alignCast(ptr));
        return self.isPrimed();
    }

    fn vtableMetadata(ptr: *anyopaque, out: *Metadata) void {
        const self: *const JurikFractalAdaptiveZeroLagVelocity = @ptrCast(@alignCast(ptr));
        self.getMetadata(out);
    }

    fn vtableUpdateScalar(ptr: *anyopaque, sample: *const Scalar) OutputArray {
        const self: *JurikFractalAdaptiveZeroLagVelocity = @ptrCast(@alignCast(ptr));
        return self.updateScalar(sample);
    }

    fn vtableUpdateBar(ptr: *anyopaque, sample: *const Bar) OutputArray {
        const self: *JurikFractalAdaptiveZeroLagVelocity = @ptrCast(@alignCast(ptr));
        return self.updateBar(sample);
    }

    fn vtableUpdateQuote(ptr: *anyopaque, sample: *const Quote) OutputArray {
        const self: *JurikFractalAdaptiveZeroLagVelocity = @ptrCast(@alignCast(ptr));
        return self.updateQuote(sample);
    }

    fn vtableUpdateTrade(ptr: *anyopaque, sample: *const Trade) OutputArray {
        const self: *JurikFractalAdaptiveZeroLagVelocity = @ptrCast(@alignCast(ptr));
        return self.updateTrade(sample);
    }

    pub const Error = error{
        InvalidLoDepth,
        InvalidHiDepth,
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

fn runJvelcfbTest(lo_depth: u32, hi_depth: u32, fractal_type: u32, smooth: u32, expected: [252]f64) !void {
    var ind = JurikFractalAdaptiveZeroLagVelocity.init(.{
        .lo_depth = lo_depth,
        .hi_depth = hi_depth,
        .fractal_type = fractal_type,
        .smooth = smooth,
    }) catch unreachable;
    ind.fixSlices();
    const input = testdata.testInput();
    const eps = 1e-13;

    for (0..252) |i| {
        const v = ind.update(input[i]);
        if (math.isNan(expected[i])) {
            try testing.expect(math.isNan(v));
        } else {
            try testing.expect(!math.isNan(v));
            if (!almostEqual(v, expected[i], eps)) {
                std.debug.print("FAIL [{d}] lo={d} hi={d} ft={d} sm={d}: expected {d}, got {d}, diff {d}\n", .{ i, lo_depth, hi_depth, fractal_type, smooth, expected[i], v, @abs(v - expected[i]) });
                return error.TestUnexpectedResult;
            }
        }
    }

    try testing.expect(math.isNan(ind.update(math.nan(f64))));
}

test "jvelcfb lo=2 hi=15" {
    try runJvelcfbTest(2, 15, 1, 10, testdata.expectedLo2Hi15());
}
test "jvelcfb lo=2 hi=30" {
    try runJvelcfbTest(2, 30, 1, 10, testdata.expectedLo2Hi30());
}
test "jvelcfb lo=2 hi=60" {
    try runJvelcfbTest(2, 60, 1, 10, testdata.expectedLo2Hi60());
}
test "jvelcfb lo=5 hi=15" {
    try runJvelcfbTest(5, 15, 1, 10, testdata.expectedLo5Hi15());
}
test "jvelcfb lo=5 hi=30" {
    try runJvelcfbTest(5, 30, 1, 10, testdata.expectedLo5Hi30());
}
test "jvelcfb lo=5 hi=60" {
    try runJvelcfbTest(5, 60, 1, 10, testdata.expectedLo5Hi60());
}
test "jvelcfb lo=10 hi=15" {
    try runJvelcfbTest(10, 15, 1, 10, testdata.expectedLo10Hi15());
}
test "jvelcfb lo=10 hi=30" {
    try runJvelcfbTest(10, 30, 1, 10, testdata.expectedLo10Hi30());
}
test "jvelcfb lo=10 hi=60" {
    try runJvelcfbTest(10, 60, 1, 10, testdata.expectedLo10Hi60());
}
test "jvelcfb ftype=2" {
    try runJvelcfbTest(5, 30, 2, 10, testdata.expectedFtype2());
}
test "jvelcfb ftype=3" {
    try runJvelcfbTest(5, 30, 3, 10, testdata.expectedFtype3());
}
test "jvelcfb ftype=4" {
    try runJvelcfbTest(5, 30, 4, 10, testdata.expectedFtype4());
}
test "jvelcfb smooth=5" {
    try runJvelcfbTest(5, 30, 1, 5, testdata.expectedSmooth5());
}
test "jvelcfb smooth=20" {
    try runJvelcfbTest(5, 30, 1, 20, testdata.expectedSmooth20());
}
test "jvelcfb smooth=40" {
    try runJvelcfbTest(5, 30, 1, 40, testdata.expectedSmooth40());
}
