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
const polyline_mod = @import("../../core/outputs/polyline.zig");

const OutputArray = indicator_mod.OutputArray;
const Identifier = identifier_mod.Identifier;
const Metadata = metadata_mod.Metadata;
const Level = levels_mod.Level;
const Levels = levels_mod.Levels;
const Point = polyline_mod.Point;
const Polyline = polyline_mod.Polyline;

/// Maximum window size supported (n <= max_n).
const max_n = 256;
/// Maximum number of extrema supported.
const max_extrema = 64;

/// Enumerates the outputs of the Moving Mini-Max indicator.
pub const MovingMiniMaxOutput = enum(u8) {
    /// The up mini-max value at the most recent bar (emphasizes local maxima).
    up = 1,
    /// The down mini-max value at the most recent bar (emphasizes local minima).
    down = 2,
    /// The detected resistance levels, sorted by strength (strongest first).
    resistances = 3,
    /// The detected support levels, sorted by strength (strongest first).
    supports = 4,
    /// The full up mini-max probability distribution over the window.
    up_distribution = 5,
    /// The full down mini-max probability distribution over the window.
    down_distribution = 6,
};

/// Parameters to create a Moving Mini-Max indicator.
pub const MovingMiniMaxParams = struct {
    m: usize = 5,
    n: usize = 50,
    num_extrema: usize = 3,
    bar_component: ?bar_component.BarComponent = null,
    quote_component: ?quote_component.QuoteComponent = null,
    trade_component: ?trade_component.TradeComponent = null,
};

/// A detected peak as a (strength, index) pair.
const Peak = struct {
    strength: f64,
    index: usize,
};

/// A detected support/resistance level.
const MiniMaxLevel = struct {
    price: f64,
    offset: i32,
    strength: f64,
};

/// Computes Q_{i,i+1} and Q_{i,i-1} for each position i = 0..n-1.
fn calcQValues(window: []const f64, n: usize, m: usize, negate: bool, q_plus: []f64, q_minus: []f64) void {
    const sign: f64 = if (negate) -1.0 else 1.0;

    for (0..n) |i| {
        const si = window[i];
        var sum_plus: f64 = 0.0;
        var sum_minus: f64 = 0.0;

        var k: usize = 1;
        while (k <= m) : (k += 1) {
            const s_forward = if (i + k < n) window[i + k] else window[n - 1];
            const s_backward = if (i >= k) window[i - k] else window[0];

            const denom_plus = s_forward + si;
            const arg_plus = if (denom_plus == 0.0) 0.0 else sign * 2.0 * (s_forward - si) / denom_plus;

            const denom_minus = s_backward + si;
            const arg_minus = if (denom_minus == 0.0) 0.0 else sign * 2.0 * (s_backward - si) / denom_minus;

            sum_plus += @exp(arg_plus);
            sum_minus += @exp(arg_minus);
        }

        q_plus[i] = sum_plus;
        q_minus[i] = sum_minus;
    }
}

/// Computes transition probabilities P_{i,i+1} and P_{i,i-1} from Q-values.
fn calcPValues(q_plus: []const f64, q_minus: []const f64, n: usize, p_plus: []f64, p_minus: []f64) void {
    for (0..n) |i| {
        const denom = q_plus[i] + q_minus[i];
        if (denom == 0.0) {
            p_plus[i] = 0.5;
            p_minus[i] = 0.5;
        } else {
            p_plus[i] = q_plus[i] / denom;
            p_minus[i] = q_minus[i] / denom;
        }
    }
}

/// Computes the normalized mini-max series from transition probabilities.
fn calcMiniMax(p_plus: []const f64, p_minus: []const f64, n: usize, out: []f64) void {
    var u: [max_n]f64 = undefined;
    u[0] = 1.0;

    var i: usize = 1;
    while (i < n) : (i += 1) {
        const p_prev_to_i = p_plus[i - 1];
        const p_i_to_prev = p_minus[i];
        if (p_i_to_prev == 0.0) {
            u[i] = u[i - 1] * 1e10;
        } else {
            u[i] = (p_prev_to_i / p_i_to_prev) * u[i - 1];
        }
    }

    var total: f64 = 0.0;
    for (0..n) |j| total += u[j];

    if (total == 0.0) {
        const uniform = 1.0 / @as(f64, @floatFromInt(n));
        for (0..n) |j| out[j] = uniform;
        return;
    }

    for (0..n) |j| out[j] = u[j] / total;
}

/// Finds distinct local peaks, written to `selected`; returns the count.
fn findPeaks(values: []const f64, num_peaks: usize, min_separation: usize, selected: []Peak) usize {
    const n = values.len;
    var candidates: [max_n]Peak = undefined;
    var cand_len: usize = 0;

    for (0..n) |i| {
        var is_peak: bool = undefined;
        if (i == 0) {
            is_peak = n <= 1 or values[i] >= values[i + 1];
        } else if (i == n - 1) {
            is_peak = values[i] >= values[i - 1];
        } else {
            is_peak = values[i] >= values[i - 1] and values[i] >= values[i + 1];
        }
        if (is_peak) {
            candidates[cand_len] = .{ .strength = values[i], .index = i };
            cand_len += 1;
        }
    }

    // Sort by strength descending; ties break on the larger index first (matches the
    // reference, which sorts (value, index) tuples in reverse).
    std.mem.sort(Peak, candidates[0..cand_len], {}, peakLessThan);

    var sel_len: usize = 0;
    for (candidates[0..cand_len]) |c| {
        if (sel_len >= num_peaks) break;
        var too_close = false;
        for (selected[0..sel_len]) |sel| {
            const diff = if (c.index > sel.index) c.index - sel.index else sel.index - c.index;
            if (diff < min_separation) {
                too_close = true;
                break;
            }
        }
        if (!too_close) {
            selected[sel_len] = c;
            sel_len += 1;
        }
    }

    return sel_len;
}

/// Orders peaks by strength descending, then by index descending.
fn peakLessThan(_: void, a: Peak, b: Peak) bool {
    if (a.strength != b.strength) return a.strength > b.strength;
    return a.index > b.index;
}

/// Zurab Silagadze's Moving Mini-Max (MMM) indicator.
pub const MovingMiniMax = struct {
    bar_func: bar_component.BarFunc,
    quote_func: quote_component.QuoteFunc,
    trade_func: trade_component.TradeFunc,

    m: usize,
    n: usize,
    num_extrema: usize,

    window: []f64,
    buf_pos: usize = 0,
    count: usize = 0,

    primed: bool = false,

    // Scratch result of the last update.
    last_valid: bool = false,
    last_up: f64 = 0.0,
    last_down: f64 = 0.0,
    up_dist_buf: [max_n]f64 = undefined,
    dn_dist_buf: [max_n]f64 = undefined,
    res_buf: [max_extrema]MiniMaxLevel = undefined,
    sup_buf: [max_extrema]MiniMaxLevel = undefined,
    res_len: usize = 0,
    sup_len: usize = 0,

    allocator: std.mem.Allocator,
    mnemonic_buf: [96]u8,
    mnemonic_len: usize,
    description_buf: [160]u8,
    description_len: usize,

    pub const InitError = error{
        InvalidM,
        InvalidN,
        InvalidNumExtrema,
        MnemonicTooLong,
        OutOfMemory,
    };

    pub fn init(allocator: std.mem.Allocator, params: MovingMiniMaxParams) !MovingMiniMax {
        const m = params.m;
        const n = params.n;
        const num_extrema = params.num_extrema;

        if (m < 1) return error.InvalidM;
        if (n <= 2 * m or n > max_n) return error.InvalidN;
        if (num_extrema < 1 or num_extrema > max_extrema) return error.InvalidNumExtrema;

        const bc = params.bar_component orelse bar_component.default_bar_component;
        const qc = params.quote_component orelse quote_component.default_quote_component;
        const tc = params.trade_component orelse trade_component.default_trade_component;

        var triple_buf: [64]u8 = undefined;
        const triple = component_triple_mnemonic_mod.componentTripleMnemonic(&triple_buf, bc, qc, tc);

        var mnemonic_buf: [96]u8 = undefined;
        const mnemonic_slice = std.fmt.bufPrint(&mnemonic_buf, "mmm({d},{d},{d}{s})", .{ m, n, num_extrema, triple }) catch
            return error.MnemonicTooLong;
        const mnemonic_len = mnemonic_slice.len;

        var description_buf: [160]u8 = undefined;
        const desc_slice = std.fmt.bufPrint(&description_buf, "Moving mini-max {s}", .{mnemonic_slice}) catch
            return error.MnemonicTooLong;
        const description_len = desc_slice.len;

        const window = try allocator.alloc(f64, n);
        @memset(window, 0.0);

        return .{
            .bar_func = bar_component.componentValue(bc),
            .quote_func = quote_component.componentValue(qc),
            .trade_func = trade_component.componentValue(tc),
            .m = m,
            .n = n,
            .num_extrema = num_extrema,
            .window = window,
            .allocator = allocator,
            .mnemonic_buf = mnemonic_buf,
            .mnemonic_len = mnemonic_len,
            .description_buf = description_buf,
            .description_len = description_len,
        };
    }

    pub fn deinit(self: *MovingMiniMax) void {
        self.allocator.free(self.window);
    }

    pub fn fixSlices(self: *MovingMiniMax) void {
        _ = self;
        // mnemonic/description are read from the inline buffers via helpers.
    }

    /// Computes the MMM set for the given price; stores it in the scratch fields.
    /// Returns true if a valid set was produced.
    pub fn updateValues(self: *MovingMiniMax, sample: f64) bool {
        self.last_valid = false;

        if (self.count < self.n) {
            self.window[self.count] = sample;
            self.count += 1;
        } else {
            self.window[self.buf_pos] = sample;
            self.buf_pos = (self.buf_pos + 1) % self.n;
        }

        if (self.count < self.n) {
            self.primed = false;
            return false;
        }

        self.primed = true;

        const n = self.n;
        const m = self.m;

        // Reconstruct the window in chronological order (oldest -> newest).
        var window: [max_n]f64 = undefined;
        for (0..n) |i| window[i] = self.window[(self.buf_pos + i) % n];

        var q_up_plus: [max_n]f64 = undefined;
        var q_up_minus: [max_n]f64 = undefined;
        var q_dn_plus: [max_n]f64 = undefined;
        var q_dn_minus: [max_n]f64 = undefined;
        calcQValues(window[0..n], n, m, false, q_up_plus[0..n], q_up_minus[0..n]);
        calcQValues(window[0..n], n, m, true, q_dn_plus[0..n], q_dn_minus[0..n]);

        var p_up_plus: [max_n]f64 = undefined;
        var p_up_minus: [max_n]f64 = undefined;
        var p_dn_plus: [max_n]f64 = undefined;
        var p_dn_minus: [max_n]f64 = undefined;
        calcPValues(q_up_plus[0..n], q_up_minus[0..n], n, p_up_plus[0..n], p_up_minus[0..n]);
        calcPValues(q_dn_plus[0..n], q_dn_minus[0..n], n, p_dn_plus[0..n], p_dn_minus[0..n]);

        calcMiniMax(p_up_plus[0..n], p_up_minus[0..n], n, self.up_dist_buf[0..n]);
        calcMiniMax(p_dn_plus[0..n], p_dn_minus[0..n], n, self.dn_dist_buf[0..n]);

        const min_sep = @max(m, 2);

        var u_peaks: [max_extrema]Peak = undefined;
        var d_peaks: [max_extrema]Peak = undefined;
        const u_count = findPeaks(self.up_dist_buf[0..n], self.num_extrema, min_sep, &u_peaks);
        const d_count = findPeaks(self.dn_dist_buf[0..n], self.num_extrema, min_sep, &d_peaks);

        self.res_len = u_count;
        for (u_peaks[0..u_count], 0..) |pk, i| {
            self.res_buf[i] = .{
                .price = window[pk.index],
                .offset = @intCast((n - 1) - pk.index),
                .strength = pk.strength,
            };
        }

        self.sup_len = d_count;
        for (d_peaks[0..d_count], 0..) |pk, i| {
            self.sup_buf[i] = .{
                .price = window[pk.index],
                .offset = @intCast((n - 1) - pk.index),
                .strength = pk.strength,
            };
        }

        self.last_up = self.up_dist_buf[n - 1];
        self.last_down = self.dn_dist_buf[n - 1];
        self.last_valid = true;
        return true;
    }

    fn levelsOf(time: i64, levels: []const MiniMaxLevel) Levels {
        if (levels.len == 0) return Levels.empty(time);
        var entries: [max_extrema]Level = undefined;
        for (levels, 0..) |lv, i| entries[i] = .{ .value = lv.price, .offset = lv.offset, .strength = lv.strength };
        return Levels.new(time, entries[0..levels.len]);
    }

    fn polylineOf(time: i64, values: []const f64) Polyline {
        if (values.len == 0) return Polyline.empty(time);
        var points: [max_n]Point = undefined;
        for (values, 0..) |v, i| points[i] = .{ .offset = @intCast(i), .value = v };
        return Polyline.new(time, points[0..values.len]);
    }

    fn makeOutput(self: *const MovingMiniMax, time: i64) OutputArray {
        const nan = math.nan(f64);
        var out = OutputArray{};

        if (self.last_valid) {
            out.append(.{ .scalar = .{ .time = time, .value = self.last_up } });
            out.append(.{ .scalar = .{ .time = time, .value = self.last_down } });
            out.append(.{ .levels = levelsOf(time, self.res_buf[0..self.res_len]) });
            out.append(.{ .levels = levelsOf(time, self.sup_buf[0..self.sup_len]) });
            out.append(.{ .polyline = polylineOf(time, self.up_dist_buf[0..self.n]) });
            out.append(.{ .polyline = polylineOf(time, self.dn_dist_buf[0..self.n]) });
        } else {
            out.append(.{ .scalar = .{ .time = time, .value = nan } });
            out.append(.{ .scalar = .{ .time = time, .value = nan } });
            out.append(.{ .levels = Levels.empty(time) });
            out.append(.{ .levels = Levels.empty(time) });
            out.append(.{ .polyline = Polyline.empty(time) });
            out.append(.{ .polyline = Polyline.empty(time) });
        }

        return out;
    }

    pub fn isPrimed(self: *const MovingMiniMax) bool {
        return self.primed;
    }

    fn mnemonic(self: *const MovingMiniMax) []const u8 {
        return self.mnemonic_buf[0..self.mnemonic_len];
    }

    fn description(self: *const MovingMiniMax) []const u8 {
        return self.description_buf[0..self.description_len];
    }

    pub fn getMetadata(self: *const MovingMiniMax, out: *Metadata) void {
        const mn = self.mnemonic();
        const desc = self.description();

        var b0: [128]u8 = undefined;
        var b1: [128]u8 = undefined;
        var b2: [128]u8 = undefined;
        var b3: [128]u8 = undefined;
        var b4: [128]u8 = undefined;
        var b5: [128]u8 = undefined;
        const m0 = std.fmt.bufPrint(&b0, "{s} up", .{mn}) catch mn;
        const m1 = std.fmt.bufPrint(&b1, "{s} down", .{mn}) catch mn;
        const m2 = std.fmt.bufPrint(&b2, "{s} resistances", .{mn}) catch mn;
        const m3 = std.fmt.bufPrint(&b3, "{s} supports", .{mn}) catch mn;
        const m4 = std.fmt.bufPrint(&b4, "{s} up dist", .{mn}) catch mn;
        const m5 = std.fmt.bufPrint(&b5, "{s} down dist", .{mn}) catch mn;

        var d0: [192]u8 = undefined;
        var d1: [192]u8 = undefined;
        var d2: [192]u8 = undefined;
        var d3: [192]u8 = undefined;
        var d4: [192]u8 = undefined;
        var d5: [192]u8 = undefined;
        const e0 = std.fmt.bufPrint(&d0, "{s} up value", .{desc}) catch desc;
        const e1 = std.fmt.bufPrint(&d1, "{s} down value", .{desc}) catch desc;
        const e2 = std.fmt.bufPrint(&d2, "{s} resistances", .{desc}) catch desc;
        const e3 = std.fmt.bufPrint(&d3, "{s} supports", .{desc}) catch desc;
        const e4 = std.fmt.bufPrint(&d4, "{s} up distribution", .{desc}) catch desc;
        const e5 = std.fmt.bufPrint(&d5, "{s} down distribution", .{desc}) catch desc;

        build_metadata_mod.buildMetadata(
            out,
            .moving_mini_max,
            mn,
            desc,
            &[_]build_metadata_mod.OutputText{
                .{ .mnemonic = m0, .description = e0 },
                .{ .mnemonic = m1, .description = e1 },
                .{ .mnemonic = m2, .description = e2 },
                .{ .mnemonic = m3, .description = e3 },
                .{ .mnemonic = m4, .description = e4 },
                .{ .mnemonic = m5, .description = e5 },
            },
        );
    }

    pub fn updateScalar(self: *MovingMiniMax, sample: *const Scalar) OutputArray {
        _ = self.updateValues(sample.value);
        return self.makeOutput(sample.time);
    }

    pub fn updateBar(self: *MovingMiniMax, sample: *const Bar) OutputArray {
        _ = self.updateValues(self.bar_func(sample.*));
        return self.makeOutput(sample.time);
    }

    pub fn updateQuote(self: *MovingMiniMax, sample: *const Quote) OutputArray {
        _ = self.updateValues(self.quote_func(sample.*));
        return self.makeOutput(sample.time);
    }

    pub fn updateTrade(self: *MovingMiniMax, sample: *const Trade) OutputArray {
        _ = self.updateValues(self.trade_func(sample.*));
        return self.makeOutput(sample.time);
    }

    pub fn indicator(self: *MovingMiniMax) indicator_mod.Indicator {
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
        const self: *MovingMiniMax = @ptrCast(@alignCast(ptr));
        return self.isPrimed();
    }
    fn vtableMetadata(ptr: *anyopaque, out: *Metadata) void {
        const self: *const MovingMiniMax = @ptrCast(@alignCast(ptr));
        self.getMetadata(out);
    }
    fn vtableUpdateScalar(ptr: *anyopaque, sample: *const Scalar) OutputArray {
        const self: *MovingMiniMax = @ptrCast(@alignCast(ptr));
        return self.updateScalar(sample);
    }
    fn vtableUpdateBar(ptr: *anyopaque, sample: *const Bar) OutputArray {
        const self: *MovingMiniMax = @ptrCast(@alignCast(ptr));
        return self.updateBar(sample);
    }
    fn vtableUpdateQuote(ptr: *anyopaque, sample: *const Quote) OutputArray {
        const self: *MovingMiniMax = @ptrCast(@alignCast(ptr));
        return self.updateQuote(sample);
    }
    fn vtableUpdateTrade(ptr: *anyopaque, sample: *const Trade) OutputArray {
        const self: *MovingMiniMax = @ptrCast(@alignCast(ptr));
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

fn checkLevels(actual: []const MiniMaxLevel, expected: []const testdata.Extremum) !void {
    try testing.expectEqual(expected.len, actual.len);
    for (0..expected.len) |i| {
        try testing.expect(@abs(actual[i].price - expected[i].price) <= tolerance * @max(1.0, @abs(expected[i].price)));
        try testing.expectEqual(@as(i32, @intCast(expected[i].offset)), actual[i].offset);
        try testing.expect(@abs(actual[i].strength - expected[i].strength) <= tolerance * @max(1.0, @abs(expected[i].strength)));
    }
}

fn checkCombo(
    inputs: []const f64,
    m: usize,
    n: usize,
    num_extrema: usize,
    exp_up: []const f64,
    exp_down: []const f64,
    exp_res: []const testdata.Extremum,
    exp_sup: []const testdata.Extremum,
) !void {
    const allocator = testing.allocator;
    var mmm = try MovingMiniMax.init(allocator, .{ .m = m, .n = n, .num_extrema = num_extrema });
    defer mmm.deinit();
    for (inputs) |p| _ = mmm.updateValues(p);
    try testing.expect(mmm.last_valid);
    try checkSeries(mmm.up_dist_buf[0..mmm.n], exp_up);
    try checkSeries(mmm.dn_dist_buf[0..mmm.n], exp_down);
    try checkLevels(mmm.res_buf[0..mmm.res_len], exp_res);
    try checkLevels(mmm.sup_buf[0..mmm.sup_len], exp_sup);
}

test "MMM m3 combos" {
    const input = testdata.testInput();
    try checkCombo(&input, 3, 50, 1, &testdata.expected_M3_N50_E1_Up(), &testdata.expected_M3_N50_E1_Down(), &testdata.expected_M3_N50_E1_Resistances(), &testdata.expected_M3_N50_E1_Supports());
    try checkCombo(&input, 3, 50, 3, &testdata.expected_M3_N50_E3_Up(), &testdata.expected_M3_N50_E3_Down(), &testdata.expected_M3_N50_E3_Resistances(), &testdata.expected_M3_N50_E3_Supports());
    try checkCombo(&input, 3, 100, 1, &testdata.expected_M3_N100_E1_Up(), &testdata.expected_M3_N100_E1_Down(), &testdata.expected_M3_N100_E1_Resistances(), &testdata.expected_M3_N100_E1_Supports());
    try checkCombo(&input, 3, 100, 3, &testdata.expected_M3_N100_E3_Up(), &testdata.expected_M3_N100_E3_Down(), &testdata.expected_M3_N100_E3_Resistances(), &testdata.expected_M3_N100_E3_Supports());
    try checkCombo(&input, 3, 252, 1, &testdata.expected_M3_N252_E1_Up(), &testdata.expected_M3_N252_E1_Down(), &testdata.expected_M3_N252_E1_Resistances(), &testdata.expected_M3_N252_E1_Supports());
    try checkCombo(&input, 3, 252, 3, &testdata.expected_M3_N252_E3_Up(), &testdata.expected_M3_N252_E3_Down(), &testdata.expected_M3_N252_E3_Resistances(), &testdata.expected_M3_N252_E3_Supports());
}

test "MMM m5 combos" {
    const input = testdata.testInput();
    try checkCombo(&input, 5, 50, 1, &testdata.expected_M5_N50_E1_Up(), &testdata.expected_M5_N50_E1_Down(), &testdata.expected_M5_N50_E1_Resistances(), &testdata.expected_M5_N50_E1_Supports());
    try checkCombo(&input, 5, 50, 3, &testdata.expected_M5_N50_E3_Up(), &testdata.expected_M5_N50_E3_Down(), &testdata.expected_M5_N50_E3_Resistances(), &testdata.expected_M5_N50_E3_Supports());
    try checkCombo(&input, 5, 100, 1, &testdata.expected_M5_N100_E1_Up(), &testdata.expected_M5_N100_E1_Down(), &testdata.expected_M5_N100_E1_Resistances(), &testdata.expected_M5_N100_E1_Supports());
    try checkCombo(&input, 5, 100, 3, &testdata.expected_M5_N100_E3_Up(), &testdata.expected_M5_N100_E3_Down(), &testdata.expected_M5_N100_E3_Resistances(), &testdata.expected_M5_N100_E3_Supports());
    try checkCombo(&input, 5, 252, 1, &testdata.expected_M5_N252_E1_Up(), &testdata.expected_M5_N252_E1_Down(), &testdata.expected_M5_N252_E1_Resistances(), &testdata.expected_M5_N252_E1_Supports());
    try checkCombo(&input, 5, 252, 3, &testdata.expected_M5_N252_E3_Up(), &testdata.expected_M5_N252_E3_Down(), &testdata.expected_M5_N252_E3_Resistances(), &testdata.expected_M5_N252_E3_Supports());
}

test "MMM m10 combos" {
    const input = testdata.testInput();
    try checkCombo(&input, 10, 50, 1, &testdata.expected_M10_N50_E1_Up(), &testdata.expected_M10_N50_E1_Down(), &testdata.expected_M10_N50_E1_Resistances(), &testdata.expected_M10_N50_E1_Supports());
    try checkCombo(&input, 10, 50, 3, &testdata.expected_M10_N50_E3_Up(), &testdata.expected_M10_N50_E3_Down(), &testdata.expected_M10_N50_E3_Resistances(), &testdata.expected_M10_N50_E3_Supports());
    try checkCombo(&input, 10, 100, 1, &testdata.expected_M10_N100_E1_Up(), &testdata.expected_M10_N100_E1_Down(), &testdata.expected_M10_N100_E1_Resistances(), &testdata.expected_M10_N100_E1_Supports());
    try checkCombo(&input, 10, 100, 3, &testdata.expected_M10_N100_E3_Up(), &testdata.expected_M10_N100_E3_Down(), &testdata.expected_M10_N100_E3_Resistances(), &testdata.expected_M10_N100_E3_Supports());
    try checkCombo(&input, 10, 252, 1, &testdata.expected_M10_N252_E1_Up(), &testdata.expected_M10_N252_E1_Down(), &testdata.expected_M10_N252_E1_Resistances(), &testdata.expected_M10_N252_E1_Supports());
    try checkCombo(&input, 10, 252, 3, &testdata.expected_M10_N252_E3_Up(), &testdata.expected_M10_N252_E3_Down(), &testdata.expected_M10_N252_E3_Resistances(), &testdata.expected_M10_N252_E3_Supports());
}

test "MMM m20 combos" {
    const input = testdata.testInput();
    try checkCombo(&input, 20, 50, 1, &testdata.expected_M20_N50_E1_Up(), &testdata.expected_M20_N50_E1_Down(), &testdata.expected_M20_N50_E1_Resistances(), &testdata.expected_M20_N50_E1_Supports());
    try checkCombo(&input, 20, 50, 3, &testdata.expected_M20_N50_E3_Up(), &testdata.expected_M20_N50_E3_Down(), &testdata.expected_M20_N50_E3_Resistances(), &testdata.expected_M20_N50_E3_Supports());
    try checkCombo(&input, 20, 100, 1, &testdata.expected_M20_N100_E1_Up(), &testdata.expected_M20_N100_E1_Down(), &testdata.expected_M20_N100_E1_Resistances(), &testdata.expected_M20_N100_E1_Supports());
    try checkCombo(&input, 20, 100, 3, &testdata.expected_M20_N100_E3_Up(), &testdata.expected_M20_N100_E3_Down(), &testdata.expected_M20_N100_E3_Resistances(), &testdata.expected_M20_N100_E3_Supports());
    try checkCombo(&input, 20, 252, 1, &testdata.expected_M20_N252_E1_Up(), &testdata.expected_M20_N252_E1_Down(), &testdata.expected_M20_N252_E1_Resistances(), &testdata.expected_M20_N252_E1_Supports());
    try checkCombo(&input, 20, 252, 3, &testdata.expected_M20_N252_E3_Up(), &testdata.expected_M20_N252_E3_Down(), &testdata.expected_M20_N252_E3_Resistances(), &testdata.expected_M20_N252_E3_Supports());
}

test "MMM latest scalars equal distribution tails" {
    const allocator = testing.allocator;
    const input = testdata.testInput();
    var mmm = try MovingMiniMax.init(allocator, .{ .m = 3, .n = 50, .num_extrema = 1 });
    defer mmm.deinit();
    for (&input) |p| _ = mmm.updateValues(p);
    try testing.expect(@abs(mmm.last_up - mmm.up_dist_buf[mmm.n - 1]) <= 1e-12);
    try testing.expect(@abs(mmm.last_down - mmm.dn_dist_buf[mmm.n - 1]) <= 1e-12);
}

test "MMM metadata default" {
    const allocator = testing.allocator;
    var mmm = try MovingMiniMax.init(allocator, .{});
    defer mmm.deinit();

    var meta: Metadata = undefined;
    mmm.getMetadata(&meta);
    try testing.expectEqual(Identifier.moving_mini_max, meta.identifier);
    try testing.expectEqualStrings("mmm(5,50,3)", meta.mnemonic);
    try testing.expectEqual(@as(usize, 6), meta.outputs_len);
}

test "MMM invalid params" {
    const allocator = testing.allocator;
    try testing.expectError(error.InvalidM, MovingMiniMax.init(allocator, .{ .m = 0 }));
    try testing.expectError(error.InvalidN, MovingMiniMax.init(allocator, .{ .m = 5, .n = 10 }));
    try testing.expectError(error.InvalidNumExtrema, MovingMiniMax.init(allocator, .{ .num_extrema = 0 }));
}
