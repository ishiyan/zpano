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
const identifier_mod = @import("../../core/identifier.zig");
const metadata_mod = @import("../../core/metadata.zig");

const OutputArray = indicator_mod.OutputArray;
const LineIndicator = line_indicator_mod.LineIndicator;
const Identifier = identifier_mod.Identifier;
const Metadata = metadata_mod.Metadata;

/// Enumerates the outputs of the RS fractal adaptive simple moving average indicator.
pub const RescaledFractalAdaptiveSimpleMovingAverageOutput = enum(u8) {
    /// The scalar value of the RSFRASMA.
    value = 0,
};

/// Parameters to create an instance of the RS fractal adaptive simple moving average indicator.
pub const RescaledFractalAdaptiveSimpleMovingAverageParams = struct {
    /// The lookback window for R/S analysis. Must be a power of 2, >= 4.
    period: usize,
    /// Base SMA period before fractal adaptation. Must be >= 1.
    normal_speed: usize,
    /// Multiplier applied to prices before R/S calculation. Default is 1.0.
    price_scale: f64 = 1.0,
    /// Bar component to extract. `null` means use default (Close).
    bar_component: ?bar_component.BarComponent = null,
    /// Quote component to extract. `null` means use default (Mid).
    quote_component: ?quote_component.QuoteComponent = null,
    /// Trade component to extract. `null` means use default (Price).
    trade_component: ?trade_component.TradeComponent = null,
};

/// Computes the RS Fractal Adaptive Simple Moving Average (RSFRASMA).
///
/// Uses Rescaled Range (R/S) analysis to estimate the Hurst exponent,
/// then adapts the SMA period accordingly.
/// The indicator is not primed during the first `period` updates.
pub const RescaledFractalAdaptiveSimpleMovingAverage = struct {
    line: LineIndicator,
    closes: std.ArrayList(f64),
    period: usize,
    normal_speed: usize,
    price_scale: f64,
    n_iter: usize,
    block_sizes: []usize,
    block_counts: []usize,
    primed: bool,
    allocator: std.mem.Allocator,
    mnemonic_buf: [80]u8,
    mnemonic_len: usize,
    description_buf: [160]u8,
    description_len: usize,

    pub fn init(allocator: std.mem.Allocator, params: RescaledFractalAdaptiveSimpleMovingAverageParams) !RescaledFractalAdaptiveSimpleMovingAverage {
        if (params.period < 4) {
            return error.InvalidPeriod;
        }
        if (params.period & (params.period - 1) != 0) {
            return error.InvalidPeriodNotPowerOf2;
        }
        if (params.normal_speed < 1) {
            return error.InvalidNormalSpeed;
        }

        const price_scale = if (params.price_scale == 0.0) 1.0 else params.price_scale;

        const bc = params.bar_component orelse bar_component.default_bar_component;
        const qc = params.quote_component orelse quote_component.default_quote_component;
        const tc = params.trade_component orelse trade_component.default_trade_component;

        var triple_buf: [64]u8 = undefined;
        const triple = component_triple_mnemonic_mod.componentTripleMnemonic(&triple_buf, bc, qc, tc);

        var mnemonic_buf: [80]u8 = undefined;
        const mnemonic_slice = std.fmt.bufPrint(&mnemonic_buf, "rsfrasma({d},{d},{d:.1}{s})", .{ params.period, params.normal_speed, price_scale, triple }) catch
            return error.MnemonicTooLong;
        const mnemonic_len = mnemonic_slice.len;

        var description_buf: [160]u8 = undefined;
        const desc_slice = std.fmt.bufPrint(&description_buf, "RS fractal adaptive simple moving average {s}", .{mnemonic_slice}) catch
            return error.MnemonicTooLong;
        const description_len = desc_slice.len;

        // Precompute R/S parameters.
        const k0 = params.period / 4;
        var n_iter: usize = 0;
        if (k0 >= 2) {
            const k0_f: f64 = @floatFromInt(k0);
            n_iter = @intFromFloat(@floor(@log(k0_f) / @log(2.0)));
        }

        const block_sizes = try allocator.alloc(usize, n_iter + 1);
        @memset(block_sizes, 0);
        const block_counts = try allocator.alloc(usize, n_iter + 1);
        @memset(block_counts, 0);

        for (1..n_iter + 1) |u| {
            block_sizes[u] = @as(usize, 1) << @intCast(u + 1);
            block_counts[u] = params.period / block_sizes[u];
        }

        var closes: std.ArrayList(f64) = .empty;
        try closes.ensureTotalCapacity(allocator, 256);

        return .{
            .line = LineIndicator.new(
                mnemonic_buf[0..mnemonic_len],
                description_buf[0..description_len],
                params.bar_component,
                params.quote_component,
                params.trade_component,
            ),
            .closes = closes,
            .period = params.period,
            .normal_speed = params.normal_speed,
            .price_scale = price_scale,
            .n_iter = n_iter,
            .block_sizes = block_sizes,
            .block_counts = block_counts,
            .primed = false,
            .allocator = allocator,
            .mnemonic_buf = mnemonic_buf,
            .mnemonic_len = mnemonic_len,
            .description_buf = description_buf,
            .description_len = description_len,
        };
    }

    pub fn deinit(self: *RescaledFractalAdaptiveSimpleMovingAverage) void {
        self.allocator.free(self.block_sizes);
        self.allocator.free(self.block_counts);
        self.closes.deinit(self.allocator);
    }

    pub fn fixSlices(self: *RescaledFractalAdaptiveSimpleMovingAverage) void {
        self.line.mnemonic = self.mnemonic_buf[0..self.mnemonic_len];
        self.line.description = self.description_buf[0..self.description_len];
    }

    pub fn update(self: *RescaledFractalAdaptiveSimpleMovingAverage, sample: f64) f64 {
        if (math.isNan(sample)) {
            return sample;
        }

        const period = self.period;
        const price_scale = self.price_scale;

        self.closes.append(self.allocator, sample) catch return math.nan(f64);
        const n_closes = self.closes.items.len;

        if (n_closes <= period) {
            return math.nan(f64);
        }

        if (!self.primed) {
            self.primed = true;
        }

        const pos = n_closes - 1;

        // R/S analysis.
        const n_iter = self.n_iter;
        var sumx: f64 = 0.0;
        var sumy: f64 = 0.0;
        var sumx2: f64 = 0.0;
        var sumxy: f64 = 0.0;
        var valid_scales: usize = 0;

        for (1..n_iter + 1) |u| {
            const block_size = self.block_sizes[u];
            const n_blocks_u = self.block_counts[u];

            if (n_blocks_u < 1) continue;

            var rs_sum: f64 = 0.0;
            var t: usize = 0;
            var block_count: usize = 0;

            while (t <= period - block_size) {
                // Block mean.
                var mu: f64 = 0.0;
                for (1..block_size + 1) |j| {
                    mu += price_scale * self.closes.items[pos - (t + j)];
                }
                const bs_f: f64 = @floatFromInt(block_size);
                mu /= bs_f;

                // Population std.
                var sum_sq: f64 = 0.0;
                for (1..block_size + 1) |j| {
                    const diff = price_scale * self.closes.items[pos - (t + j)] - mu;
                    sum_sq += diff * diff;
                }
                var std_val = @sqrt(sum_sq / bs_f);
                if (std_val <= 0.0) std_val = 0.1;

                // Cumulative deviations and range.
                var cum_dev: f64 = 0.0;
                var w_max: f64 = 0.0;
                var w_min: f64 = 9999999999.0;

                for (1..block_size + 1) |k| {
                    cum_dev += price_scale * self.closes.items[pos - (t + k)] - mu;
                    if (cum_dev > w_max) w_max = cum_dev;
                    if (cum_dev < w_min) w_min = cum_dev;
                }

                if (w_max < 0.0) w_max = 0.0;
                if (w_min > 0.0) w_min = 0.0;

                const r_val = w_max - w_min;
                rs_sum += r_val / std_val;
                t += block_size;
                block_count += 1;
            }

            // Average R/S for this scale.
            var rs_avg: f64 = 1.0;
            if (block_count > 0) {
                rs_avg = rs_sum / @as(f64, @floatFromInt(block_count));
            }
            if (rs_avg <= 0.0) rs_avg = 1e-10;

            const log2_d = @log(@as(f64, @floatFromInt(block_size))) / @log(2.0);
            const log2_rs = @log(rs_avg) / @log(2.0);

            sumx += log2_d;
            sumy += log2_rs;
            sumx2 += log2_d * log2_d;
            sumxy += log2_d * log2_rs;
            valid_scales += 1;
        }

        // Linear regression slope = Hurst exponent.
        var h: f64 = 0.5;
        if (valid_scales >= 2) {
            const vs_f: f64 = @floatFromInt(valid_scales);
            const h1 = vs_f * sumxy - sumx * sumy;
            var h2 = vs_f * sumx2 - sumx * sumx;
            if (h2 <= 0.0) h2 = 0.1;
            h = h1 / h2;
        }

        // Guard H.
        if (2.0 * h <= 0.0) h = 0.001;

        const alpha = 1.0 / (2.0 * h);
        const ns_f: f64 = @floatFromInt(self.normal_speed);
        const spd_raw = @round(ns_f * alpha);
        const spd: usize = if (spd_raw < 1.0) 1 else @intFromFloat(spd_raw);

        // Compute SMA with adapted speed.
        const sma_start: usize = if (spd <= pos + 1) pos + 1 - spd else 0;
        const count = pos - sma_start + 1;

        var total: f64 = 0.0;
        for (sma_start..pos + 1) |i| {
            total += self.closes.items[i];
        }

        return total / @as(f64, @floatFromInt(count));
    }

    pub fn isPrimed(self: *const RescaledFractalAdaptiveSimpleMovingAverage) bool {
        return self.primed;
    }

    pub fn getMetadata(self: *const RescaledFractalAdaptiveSimpleMovingAverage, out: *Metadata) void {
        build_metadata_mod.buildMetadata(
            out,
            .rescaled_fractal_adaptive_simple_moving_average,
            self.line.mnemonic,
            self.line.description,
            &[_]build_metadata_mod.OutputText{
                .{ .mnemonic = self.line.mnemonic, .description = self.line.description },
            },
        );
    }

    pub fn updateScalar(self: *RescaledFractalAdaptiveSimpleMovingAverage, sample: *const Scalar) OutputArray {
        const value = self.update(sample.value);
        return LineIndicator.wrapScalar(sample.time, value);
    }

    pub fn updateBar(self: *RescaledFractalAdaptiveSimpleMovingAverage, sample: *const Bar) OutputArray {
        const value = self.update(self.line.extractBar(sample));
        return LineIndicator.wrapScalar(sample.time, value);
    }

    pub fn updateQuote(self: *RescaledFractalAdaptiveSimpleMovingAverage, sample: *const Quote) OutputArray {
        const value = self.update(self.line.extractQuote(sample));
        return LineIndicator.wrapScalar(sample.time, value);
    }

    pub fn updateTrade(self: *RescaledFractalAdaptiveSimpleMovingAverage, sample: *const Trade) OutputArray {
        const value = self.update(self.line.extractTrade(sample));
        return LineIndicator.wrapScalar(sample.time, value);
    }

    pub fn indicator(self: *RescaledFractalAdaptiveSimpleMovingAverage) indicator_mod.Indicator {
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
        const self: *RescaledFractalAdaptiveSimpleMovingAverage = @ptrCast(@alignCast(ptr));
        return self.isPrimed();
    }

    fn vtableMetadata(ptr: *anyopaque, out: *Metadata) void {
        const self: *const RescaledFractalAdaptiveSimpleMovingAverage = @ptrCast(@alignCast(ptr));
        self.getMetadata(out);
    }

    fn vtableUpdateScalar(ptr: *anyopaque, sample: *const Scalar) OutputArray {
        const self: *RescaledFractalAdaptiveSimpleMovingAverage = @ptrCast(@alignCast(ptr));
        return self.updateScalar(sample);
    }

    fn vtableUpdateBar(ptr: *anyopaque, sample: *const Bar) OutputArray {
        const self: *RescaledFractalAdaptiveSimpleMovingAverage = @ptrCast(@alignCast(ptr));
        return self.updateBar(sample);
    }

    fn vtableUpdateQuote(ptr: *anyopaque, sample: *const Quote) OutputArray {
        const self: *RescaledFractalAdaptiveSimpleMovingAverage = @ptrCast(@alignCast(ptr));
        return self.updateQuote(sample);
    }

    fn vtableUpdateTrade(ptr: *anyopaque, sample: *const Trade) OutputArray {
        const self: *RescaledFractalAdaptiveSimpleMovingAverage = @ptrCast(@alignCast(ptr));
        return self.updateTrade(sample);
    }

    pub const InitError = error{
        InvalidPeriod,
        InvalidPeriodNotPowerOf2,
        InvalidNormalSpeed,
        MnemonicTooLong,
        OutOfMemory,
    };
};

// ---------------------------------------------------------------------------
// Tests
// ---------------------------------------------------------------------------

const testing = std.testing;
const testdata = @import("testdata.zig");

fn createRsfrasma(allocator: std.mem.Allocator, period: usize, normal_speed: usize, price_scale: f64) !RescaledFractalAdaptiveSimpleMovingAverage {
    var f = try RescaledFractalAdaptiveSimpleMovingAverage.init(allocator, .{ .period = period, .normal_speed = normal_speed, .price_scale = price_scale });
    f.fixSlices();
    return f;
}

fn runTest(allocator: std.mem.Allocator, period: usize, normal_speed: usize, price_scale: f64, exp: [252]f64) !void {
    const input = testdata.testInput();
    var f = try createRsfrasma(allocator, period, normal_speed, price_scale);
    defer f.deinit();

    for (0..252) |i| {
        const act = f.update(input[i]);
        if (math.isNan(exp[i])) {
            try testing.expect(math.isNan(act));
        } else {
            try testing.expect(@abs(exp[i] - act) < 1e-13);
        }
    }
}

test "rsfrasma update p4 s1" {
    try runTest(testing.allocator, 4, 30, 1.0, testdata.expectedP4_S1());
}

test "rsfrasma update p8 s1" {
    try runTest(testing.allocator, 8, 30, 1.0, testdata.expectedP8_S1());
}

test "rsfrasma update p16 s1" {
    try runTest(testing.allocator, 16, 30, 1.0, testdata.expectedP16_S1());
}

test "rsfrasma update p32 s1" {
    try runTest(testing.allocator, 32, 30, 1.0, testdata.expectedP32_S1());
}

test "rsfrasma update p64 s1" {
    try runTest(testing.allocator, 64, 30, 1.0, testdata.expectedP64_S1());
}

test "rsfrasma update p128 s1" {
    try runTest(testing.allocator, 128, 30, 1.0, testdata.expectedP128_S1());
}

test "rsfrasma update p32 s100" {
    try runTest(testing.allocator, 32, 30, 100.0, testdata.expectedP32_S100());
}

test "rsfrasma update p32 s10000" {
    try runTest(testing.allocator, 32, 30, 10000.0, testdata.expectedP32_S10000());
}

test "rsfrasma is primed" {
    const input = testdata.testInput();
    var f = try createRsfrasma(testing.allocator, 64, 30, 1.0);
    defer f.deinit();

    for (0..64) |i| {
        _ = f.update(input[i]);
        try testing.expect(!f.isPrimed());
    }
    _ = f.update(input[64]);
    try testing.expect(f.isPrimed());
}

test "rsfrasma nan passthrough" {
    var f = try createRsfrasma(testing.allocator, 4, 30, 1.0);
    defer f.deinit();
    try testing.expect(math.isNan(f.update(math.nan(f64))));
}

test "rsfrasma invalid period" {
    const result = RescaledFractalAdaptiveSimpleMovingAverage.init(testing.allocator, .{ .period = 2, .normal_speed = 30 });
    try testing.expectError(error.InvalidPeriod, result);
}

test "rsfrasma invalid period not power of 2" {
    const result = RescaledFractalAdaptiveSimpleMovingAverage.init(testing.allocator, .{ .period = 6, .normal_speed = 30 });
    try testing.expectError(error.InvalidPeriodNotPowerOf2, result);
}

test "rsfrasma invalid normal_speed" {
    const result = RescaledFractalAdaptiveSimpleMovingAverage.init(testing.allocator, .{ .period = 4, .normal_speed = 0 });
    try testing.expectError(error.InvalidNormalSpeed, result);
}
