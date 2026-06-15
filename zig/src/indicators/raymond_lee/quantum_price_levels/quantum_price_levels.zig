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
const build_metadata_mod = @import("../../core/build_metadata.zig");
const component_triple_mnemonic_mod = @import("../../core/component_triple_mnemonic.zig");
const identifier_mod = @import("../../core/identifier.zig");
const metadata_mod = @import("../../core/metadata.zig");
const levels_mod = @import("../../core/outputs/levels.zig");

const OutputArray = indicator_mod.OutputArray;
const Identifier = identifier_mod.Identifier;
const Metadata = metadata_mod.Metadata;
const Level = levels_mod.Level;
const Levels = levels_mod.Levels;

const max_bins_cap = 1024;
const max_levels_cap = 64;

/// Enumerates the outputs of the Quantum Price Levels indicator.
pub const QuantumPriceLevelsOutput = enum(u8) {
    /// The anharmonic coefficient (lambda) of the quantum potential well.
    lambda = 1,
    /// The population standard deviation of the price-return ratios in the window.
    return_std_dev = 2,
    /// The normalized QPR multipliers (1 + scale_factor*sigma*QPR(n)), one per level.
    normalized_multipliers = 3,
    /// The resistance price levels above the current price (price * NQPR(n)).
    resistances = 4,
    /// The support price levels below the current price (price / NQPR(n)).
    supports = 5,
};

/// Parameters to create a Quantum Price Levels indicator.
pub const QuantumPriceLevelsParams = struct {
    lookback: usize = 2048,
    num_levels: usize = 21,
    num_bins: usize = 100,
    scale_factor: f64 = 0.21,
    bar_component: ?bar_component.BarComponent = null,
    quote_component: ?quote_component.QuoteComponent = null,
    trade_component: ?trade_component.TradeComponent = null,
};

/// Signed real cube root via pow (matches the reference implementation).
fn cbrt(x: f64) f64 {
    if (x >= 0.0) return math.pow(f64, x, 1.0 / 3.0);
    return -math.pow(f64, -x, 1.0 / 3.0);
}

/// K0 constant for energy level n (Dasgupta et al. 2007).
fn computeK0(n: usize) f64 {
    const fn_: f64 = @floatFromInt(n);
    const numerator = 1.1924 + 33.2383 * fn_ + 56.2169 * fn_ * fn_;
    const denominator = 1.0 + 43.6106 * fn_;
    return math.pow(f64, numerator / denominator, 1.0 / 3.0);
}

/// Raymond Lee's Quantum Price Levels (QPL) indicator.
pub const QuantumPriceLevels = struct {
    bar_func: bar_component.BarFunc,
    quote_func: quote_component.QuoteFunc,
    trade_func: trade_component.TradeFunc,

    lookback: usize,
    num_levels: usize,
    num_bins: usize,
    scale_factor: f64,

    k: []f64,
    returns: []f64,
    buf_pos: usize = 0,
    count: usize = 0,
    prev_price: f64 = 0.0,
    have_prev: bool = false,

    primed: bool = false,

    // Scratch result of the last update.
    last_valid: bool = false,
    last_lambda: f64 = 0.0,
    last_sigma: f64 = 0.0,
    last_n: usize = 0,
    nqpr_buf: [max_levels_cap]f64 = undefined,
    res_buf: [max_levels_cap]f64 = undefined,
    sup_buf: [max_levels_cap]f64 = undefined,

    allocator: std.mem.Allocator,
    mnemonic_buf: [96]u8,
    mnemonic_len: usize,
    description_buf: [160]u8,
    description_len: usize,

    pub const InitError = error{
        InvalidLookback,
        InvalidNumLevels,
        InvalidNumBins,
        InvalidScaleFactor,
        MnemonicTooLong,
        OutOfMemory,
    };

    pub fn init(allocator: std.mem.Allocator, params: QuantumPriceLevelsParams) !QuantumPriceLevels {
        const lookback = params.lookback;
        const num_levels = params.num_levels;
        const num_bins = params.num_bins;
        const scale_factor = params.scale_factor;

        if (lookback < 2) return error.InvalidLookback;
        if (num_levels < 1 or num_levels > max_levels_cap) return error.InvalidNumLevels;
        if (num_bins < 2 or num_bins > max_bins_cap) return error.InvalidNumBins;
        if (scale_factor <= 0.0) return error.InvalidScaleFactor;

        const bc = params.bar_component orelse bar_component.default_bar_component;
        const qc = params.quote_component orelse quote_component.default_quote_component;
        const tc = params.trade_component orelse trade_component.default_trade_component;

        var triple_buf: [64]u8 = undefined;
        const triple = component_triple_mnemonic_mod.componentTripleMnemonic(&triple_buf, bc, qc, tc);

        var mnemonic_buf: [96]u8 = undefined;
        const mnemonic_slice = std.fmt.bufPrint(&mnemonic_buf, "qpl({d},{d},{d},{d}{s})", .{ lookback, num_levels, num_bins, scale_factor, triple }) catch
            return error.MnemonicTooLong;
        const mnemonic_len = mnemonic_slice.len;

        var description_buf: [160]u8 = undefined;
        const desc_slice = std.fmt.bufPrint(&description_buf, "Quantum price levels {s}", .{mnemonic_slice}) catch
            return error.MnemonicTooLong;
        const description_len = desc_slice.len;

        const k = try allocator.alloc(f64, num_levels);
        errdefer allocator.free(k);
        for (0..num_levels) |n| k[n] = computeK0(n);

        const returns = try allocator.alloc(f64, lookback);
        @memset(returns, 0.0);

        return .{
            .bar_func = bar_component.componentValue(bc),
            .quote_func = quote_component.componentValue(qc),
            .trade_func = trade_component.componentValue(tc),
            .lookback = lookback,
            .num_levels = num_levels,
            .num_bins = num_bins,
            .scale_factor = scale_factor,
            .k = k,
            .returns = returns,
            .allocator = allocator,
            .mnemonic_buf = mnemonic_buf,
            .mnemonic_len = mnemonic_len,
            .description_buf = description_buf,
            .description_len = description_len,
        };
    }

    pub fn deinit(self: *QuantumPriceLevels) void {
        self.allocator.free(self.k);
        self.allocator.free(self.returns);
    }

    pub fn fixSlices(self: *QuantumPriceLevels) void {
        _ = self;
        // mnemonic/description are read from the inline buffers via helpers.
    }

    /// Computes the QPL set for the given price; stores it in the scratch fields.
    /// Returns true if a valid set was produced.
    pub fn updateValues(self: *QuantumPriceLevels, sample: f64) bool {
        self.last_valid = false;

        if (!self.have_prev) {
            self.prev_price = sample;
            self.have_prev = true;
            self.primed = false;
            return false;
        }

        const new_return: f64 = if (sample > 0.0) self.prev_price / sample else 1.0;
        self.prev_price = sample;

        if (self.count < self.lookback) {
            self.returns[self.count] = new_return;
            self.count += 1;
        } else {
            self.returns[self.buf_pos] = new_return;
            self.buf_pos = (self.buf_pos + 1) % self.lookback;
        }

        if (self.count < self.lookback) {
            self.primed = false;
            return false;
        }

        self.primed = true;

        const lookback = self.lookback;
        const num_bins = self.num_bins;
        const num_levels = self.num_levels;
        const scale_factor = self.scale_factor;

        // Statistics (population mu, sigma).
        var sum_r: f64 = 0.0;
        for (0..lookback) |i| sum_r += self.returns[i];
        const mu = sum_r / @as(f64, @floatFromInt(lookback));

        var sum_var: f64 = 0.0;
        for (0..lookback) |i| {
            const diff = self.returns[i] - mu;
            sum_var += diff * diff;
        }
        const sigma = @sqrt(sum_var / @as(f64, @floatFromInt(lookback)));
        if (sigma == 0.0) return false;

        // Histogram centred at r = 1.
        const half_bins = num_bins / 2;
        const dr = 3.0 * sigma / @as(f64, @floatFromInt(half_bins));
        const left_boundary = 1.0 - @as(f64, @floatFromInt(half_bins)) * dr;

        var q: [max_bins_cap]usize = undefined;
        @memset(q[0..num_bins], 0);
        var total_count: usize = 0;

        for (0..lookback) |i| {
            const r = self.returns[i];
            const idx_f = (r - left_boundary) / dr;
            if (idx_f >= 0.0) {
                const bin_index: usize = @intFromFloat(idx_f);
                if (bin_index < num_bins) {
                    q[bin_index] += 1;
                    total_count += 1;
                }
            }
        }

        if (total_count == 0) return false;
        const total_f: f64 = @floatFromInt(total_count);

        // Ground state (peak bin).
        var max_q: f64 = 0.0;
        var max_qno: usize = 0;
        for (0..num_bins) |kk| {
            const nq = @as(f64, @floatFromInt(q[kk])) / total_f;
            if (nq > max_q) {
                max_q = nq;
                max_qno = kk;
            }
        }

        if (max_qno == 0 or max_qno == num_bins - 1) return false;

        // lambda via FDM.
        const phi_plus1 = @as(f64, @floatFromInt(q[max_qno + 1])) / total_f;
        const phi_minus1 = @as(f64, @floatFromInt(q[max_qno - 1])) / total_f;

        const r_peak = left_boundary + @as(f64, @floatFromInt(max_qno)) * dr;
        const r0 = r_peak - dr / 2.0;
        const r_plus1 = r0 + dr;
        const r_minus1 = r0 - dr;

        const l_up = (r_minus1 * r_minus1) * phi_minus1 - (r_plus1 * r_plus1) * phi_plus1;
        const l_dw = (r_plus1 * r_plus1 * r_plus1 * r_plus1) * phi_plus1 - (r_minus1 * r_minus1 * r_minus1 * r_minus1) * phi_minus1;

        if (l_dw == 0.0) return false;

        const lambda = @abs(l_up / l_dw);

        // Energy levels via Cardano.
        var qfel: [max_levels_cap]f64 = undefined;
        for (0..num_levels) |n| {
            const two_n_plus_1: f64 = @floatFromInt(2 * n + 1);
            const p = -(two_n_plus_1 * two_n_plus_1);
            const q_coef = -lambda * (two_n_plus_1 * two_n_plus_1 * two_n_plus_1) * (self.k[n] * self.k[n] * self.k[n]);
            const discriminant = (q_coef * q_coef) / 4.0 + (p * p * p) / 27.0;
            if (discriminant < 0.0) return false;
            const sqrt_d = @sqrt(discriminant);
            const u = cbrt(-q_coef / 2.0 + sqrt_d);
            const v = cbrt(-q_coef / 2.0 - sqrt_d);
            qfel[n] = u + v;
        }

        if (qfel[0] == 0.0) return false;

        // NQPR and projection from the current price.
        for (0..num_levels) |n| {
            const qpr = qfel[n] / qfel[0];
            self.nqpr_buf[n] = 1.0 + scale_factor * sigma * qpr;
            self.res_buf[n] = sample * self.nqpr_buf[n];
            self.sup_buf[n] = sample / self.nqpr_buf[n];
        }

        self.last_valid = true;
        self.last_lambda = lambda;
        self.last_sigma = sigma;
        self.last_n = num_levels;
        return true;
    }

    fn levelsOf(time: i64, values: []const f64) Levels {
        if (values.len == 0) return Levels.empty(time);
        var entries: [max_levels_cap]Level = undefined;
        for (values, 0..) |v, i| entries[i] = .{ .value = v };
        return Levels.new(time, entries[0..values.len]);
    }

    fn makeOutput(self: *const QuantumPriceLevels, time: i64) OutputArray {
        const nan = math.nan(f64);
        var out = OutputArray{};

        if (self.last_valid) {
            out.append(.{ .scalar = .{ .time = time, .value = self.last_lambda } });
            out.append(.{ .scalar = .{ .time = time, .value = self.last_sigma } });
            out.append(.{ .levels = levelsOf(time, self.nqpr_buf[0..self.last_n]) });
            out.append(.{ .levels = levelsOf(time, self.res_buf[0..self.last_n]) });
            out.append(.{ .levels = levelsOf(time, self.sup_buf[0..self.last_n]) });
        } else {
            out.append(.{ .scalar = .{ .time = time, .value = nan } });
            out.append(.{ .scalar = .{ .time = time, .value = nan } });
            out.append(.{ .levels = Levels.empty(time) });
            out.append(.{ .levels = Levels.empty(time) });
            out.append(.{ .levels = Levels.empty(time) });
        }

        return out;
    }

    pub fn isPrimed(self: *const QuantumPriceLevels) bool {
        return self.primed;
    }

    fn mnemonic(self: *const QuantumPriceLevels) []const u8 {
        return self.mnemonic_buf[0..self.mnemonic_len];
    }

    fn description(self: *const QuantumPriceLevels) []const u8 {
        return self.description_buf[0..self.description_len];
    }

    pub fn getMetadata(self: *const QuantumPriceLevels, out: *Metadata) void {
        const mn = self.mnemonic();
        const desc = self.description();

        var b0: [128]u8 = undefined;
        var b1: [128]u8 = undefined;
        var b2: [128]u8 = undefined;
        var b3: [128]u8 = undefined;
        var b4: [128]u8 = undefined;
        const m0 = std.fmt.bufPrint(&b0, "{s} lambda", .{mn}) catch mn;
        const m1 = std.fmt.bufPrint(&b1, "{s} stddev", .{mn}) catch mn;
        const m2 = std.fmt.bufPrint(&b2, "{s} nqpr", .{mn}) catch mn;
        const m3 = std.fmt.bufPrint(&b3, "{s} resistances", .{mn}) catch mn;
        const m4 = std.fmt.bufPrint(&b4, "{s} supports", .{mn}) catch mn;

        var d0: [192]u8 = undefined;
        var d1: [192]u8 = undefined;
        var d2: [192]u8 = undefined;
        var d3: [192]u8 = undefined;
        var d4: [192]u8 = undefined;
        const e0 = std.fmt.bufPrint(&d0, "{s} anharmonic coefficient", .{desc}) catch desc;
        const e1 = std.fmt.bufPrint(&d1, "{s} return standard deviation", .{desc}) catch desc;
        const e2 = std.fmt.bufPrint(&d2, "{s} normalized multipliers", .{desc}) catch desc;
        const e3 = std.fmt.bufPrint(&d3, "{s} resistance levels", .{desc}) catch desc;
        const e4 = std.fmt.bufPrint(&d4, "{s} support levels", .{desc}) catch desc;

        build_metadata_mod.buildMetadata(
            out,
            .quantum_price_levels,
            mn,
            desc,
            &[_]build_metadata_mod.OutputText{
                .{ .mnemonic = m0, .description = e0 },
                .{ .mnemonic = m1, .description = e1 },
                .{ .mnemonic = m2, .description = e2 },
                .{ .mnemonic = m3, .description = e3 },
                .{ .mnemonic = m4, .description = e4 },
            },
        );
    }

    pub fn updateScalar(self: *QuantumPriceLevels, sample: *const Scalar) OutputArray {
        _ = self.updateValues(sample.value);
        return self.makeOutput(sample.time);
    }

    pub fn updateBar(self: *QuantumPriceLevels, sample: *const Bar) OutputArray {
        _ = self.updateValues(self.bar_func(sample.*));
        return self.makeOutput(sample.time);
    }

    pub fn updateQuote(self: *QuantumPriceLevels, sample: *const Quote) OutputArray {
        _ = self.updateValues(self.quote_func(sample.*));
        return self.makeOutput(sample.time);
    }

    pub fn updateTrade(self: *QuantumPriceLevels, sample: *const Trade) OutputArray {
        _ = self.updateValues(self.trade_func(sample.*));
        return self.makeOutput(sample.time);
    }

    pub fn indicator(self: *QuantumPriceLevels) indicator_mod.Indicator {
        return .{ .ptr = @ptrCast(self), .vtable = &vtable };
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
        const self: *QuantumPriceLevels = @ptrCast(@alignCast(ptr));
        return self.isPrimed();
    }
    fn vtableMetadata(ptr: *anyopaque, out: *Metadata) void {
        const self: *const QuantumPriceLevels = @ptrCast(@alignCast(ptr));
        self.getMetadata(out);
    }
    fn vtableUpdateScalar(ptr: *anyopaque, sample: *const Scalar) OutputArray {
        const self: *QuantumPriceLevels = @ptrCast(@alignCast(ptr));
        return self.updateScalar(sample);
    }
    fn vtableUpdateBar(ptr: *anyopaque, sample: *const Bar) OutputArray {
        const self: *QuantumPriceLevels = @ptrCast(@alignCast(ptr));
        return self.updateBar(sample);
    }
    fn vtableUpdateQuote(ptr: *anyopaque, sample: *const Quote) OutputArray {
        const self: *QuantumPriceLevels = @ptrCast(@alignCast(ptr));
        return self.updateQuote(sample);
    }
    fn vtableUpdateTrade(ptr: *anyopaque, sample: *const Trade) OutputArray {
        const self: *QuantumPriceLevels = @ptrCast(@alignCast(ptr));
        return self.updateTrade(sample);
    }
};

// ---------------------------------------------------------------------------
// Tests
// ---------------------------------------------------------------------------

const testing = std.testing;
const testdata = @import("testdata.zig");

const tolerance = 1e-9;

fn checkSeries(actual: []const f64, expected: []const f64) !void {
    try testing.expectEqual(expected.len, actual.len);
    for (0..expected.len) |i| {
        const delta = tolerance * @max(1.0, @abs(expected[i]));
        try testing.expect(@abs(actual[i] - expected[i]) <= delta);
    }
}

fn runLast(allocator: std.mem.Allocator, inputs: []const f64, lookback_in: usize, num_levels: usize, num_bins: usize, scale_factor: f64) !QuantumPriceLevels {
    const lookback = if (lookback_in == 0) inputs.len - 1 else lookback_in;
    var qpl = try QuantumPriceLevels.init(allocator, .{
        .lookback = lookback, .num_levels = num_levels, .num_bins = num_bins, .scale_factor = scale_factor,
    });
    for (inputs) |p| _ = qpl.updateValues(p);
    return qpl;
}

fn checkCombo(inputs: []const f64, lookback: usize, num_levels: usize, num_bins: usize, scale_factor: f64, exp_nqpr: []const f64, exp_up: []const f64, exp_lo: []const f64) !void {
    const allocator = testing.allocator;
    var qpl = try runLast(allocator, inputs, lookback, num_levels, num_bins, scale_factor);
    defer qpl.deinit();
    try testing.expect(qpl.last_valid);
    try checkSeries(qpl.nqpr_buf[0..qpl.last_n], exp_nqpr);
    try checkSeries(qpl.res_buf[0..qpl.last_n], exp_up);
    try checkSeries(qpl.sup_buf[0..qpl.last_n], exp_lo);
}

test "QPL batch combos" {
    const input = testdata.testInput();
    try checkCombo(&input, 0, 21, 100, 0.21, &testdata.expectedNQPR(), &testdata.expectedUPPER(), &testdata.expectedLOWER());
    try checkCombo(&input, 0, 21, 100, 0.10, &testdata.expectedNQPR_F0_10(), &testdata.expectedUPPER_F0_10(), &testdata.expectedLOWER_F0_10());
    try checkCombo(&input, 0, 21, 100, 0.42, &testdata.expectedNQPR_F0_42(), &testdata.expectedUPPER_F0_42(), &testdata.expectedLOWER_F0_42());
    try checkCombo(&input, 0, 21, 50, 0.21, &testdata.expectedNQPR_B50(), &testdata.expectedUPPER_B50(), &testdata.expectedLOWER_B50());
    try checkCombo(&input, 0, 21, 50, 0.10, &testdata.expectedNQPR_B50_F0_10(), &testdata.expectedUPPER_B50_F0_10(), &testdata.expectedLOWER_B50_F0_10());
    try checkCombo(&input, 0, 21, 50, 0.42, &testdata.expectedNQPR_B50_F0_42(), &testdata.expectedUPPER_B50_F0_42(), &testdata.expectedLOWER_B50_F0_42());
    try checkCombo(&input, 0, 5, 100, 0.21, &testdata.expectedNQPR_L5(), &testdata.expectedUPPER_L5(), &testdata.expectedLOWER_L5());
    try checkCombo(&input, 0, 10, 100, 0.21, &testdata.expectedNQPR_L10(), &testdata.expectedUPPER_L10(), &testdata.expectedLOWER_L10());
    try checkCombo(&input, 0, 10, 50, 0.42, &testdata.expectedNQPR_L10_B50_F0_42(), &testdata.expectedUPPER_L10_B50_F0_42(), &testdata.expectedLOWER_L10_B50_F0_42());
}

test "QPL long 2K" {
    const input = testdata.testInput2K();
    try checkCombo(&input, 0, 21, 100, 0.21, &testdata.expectedNQPR_2K(), &testdata.expectedUPPER_2K(), &testdata.expectedLOWER_2K());
}

test "QPL streaming combos" {
    const input = testdata.testInput();
    try checkCombo(&input, 100, 21, 100, 0.21, &testdata.expectedNQPR_S100(), &testdata.expectedUPPER_S100(), &testdata.expectedLOWER_S100());
    try checkCombo(&input, 150, 21, 50, 0.21, &testdata.expectedNQPR_S150_B50(), &testdata.expectedUPPER_S150_B50(), &testdata.expectedLOWER_S150_B50());
    try checkCombo(&input, 200, 21, 100, 0.42, &testdata.expectedNQPR_S200_F0_42(), &testdata.expectedUPPER_S200_F0_42(), &testdata.expectedLOWER_S200_F0_42());
}

test "QPL reference projection" {
    const allocator = testing.allocator;
    const input = testdata.testInput();
    var qpl = try runLast(allocator, &input, 0, 21, 100, 0.21);
    defer qpl.deinit();
    try testing.expect(qpl.last_valid);

    const nqpr = qpl.nqpr_buf[0..qpl.last_n];
    try checkSeries(nqpr, &testdata.expectedNQPR_R50_0());

    var up: [21]f64 = undefined;
    var lo: [21]f64 = undefined;
    for (nqpr, 0..) |m, i| {
        up[i] = 50.0 * m;
        lo[i] = 50.0 / m;
    }
    try checkSeries(&up, &testdata.expectedUPPER_R50_0());
    try checkSeries(&lo, &testdata.expectedLOWER_R50_0());
}

test "QPL scalars" {
    const allocator = testing.allocator;
    const input = testdata.testInput();
    var qpl = try runLast(allocator, &input, 0, 21, 100, 0.21);
    defer qpl.deinit();
    try testing.expect(@abs(qpl.last_lambda - 9.739608012591481e-01) <= 1e-9);
    try testing.expect(@abs(qpl.last_sigma - 2.662021797593086e-02) <= 1e-9);
}

test "QPL metadata default" {
    const allocator = testing.allocator;
    var qpl = try QuantumPriceLevels.init(allocator, .{});
    defer qpl.deinit();

    var meta: Metadata = undefined;
    qpl.getMetadata(&meta);
    try testing.expectEqual(Identifier.quantum_price_levels, meta.identifier);
    try testing.expectEqualStrings("qpl(2048,21,100,0.21)", meta.mnemonic);
    try testing.expectEqual(@as(usize, 5), meta.outputs_len);
}

test "QPL invalid params" {
    const allocator = testing.allocator;
    try testing.expectError(error.InvalidLookback, QuantumPriceLevels.init(allocator, .{ .lookback = 1 }));
    try testing.expectError(error.InvalidNumLevels, QuantumPriceLevels.init(allocator, .{ .num_levels = 0 }));
    try testing.expectError(error.InvalidNumBins, QuantumPriceLevels.init(allocator, .{ .num_bins = 1 }));
    try testing.expectError(error.InvalidScaleFactor, QuantumPriceLevels.init(allocator, .{ .scale_factor = 0.0 }));
}
