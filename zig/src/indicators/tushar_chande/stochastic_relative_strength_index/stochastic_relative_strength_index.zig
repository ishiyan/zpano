const std = @import("std");
const math = std.math;

const Bar = @import("bar").Bar;
const Quote = @import("quote").Quote;
const Trade = @import("trade").Trade;
const Scalar = @import("scalar").Scalar;
const bar_component = @import("bar_component");
const quote_component = @import("quote_component");
const trade_component = @import("trade_component");

const indicator_mod = @import("../../core/indicator.zig");
const line_indicator_mod = @import("../../core/line_indicator.zig");
const build_metadata_mod = @import("../../core/build_metadata.zig");
const component_triple_mnemonic_mod = @import("../../core/component_triple_mnemonic.zig");
const identifier_mod = @import("../../core/identifier.zig");
const metadata_mod = @import("../../core/metadata.zig");

const rsi_mod = @import("../../welles_wilder/relative_strength_index/relative_strength_index.zig");
const sma_mod = @import("../../common/simple_moving_average/simple_moving_average.zig");
const ema_mod = @import("../../common/exponential_moving_average/exponential_moving_average.zig");

const OutputArray = indicator_mod.OutputArray;
const LineIndicator = line_indicator_mod.LineIndicator;
const Identifier = identifier_mod.Identifier;
const Metadata = metadata_mod.Metadata;

/// Enumerates the outputs of the Stochastic RSI indicator.
pub const StochasticRelativeStrengthIndexOutput = enum(u8) {
    /// The Fast-K line.
    fast_k = 1,
    /// The Fast-D line (smoothed Fast-K).
    fast_d = 2,
};

/// Specifies the moving average type for Fast-D smoothing.
pub const MovingAverageType = enum(u8) {
    sma = 0,
    ema = 1,
};

/// Parameters to create a Stochastic RSI indicator.
pub const StochasticRelativeStrengthIndexParams = struct {
    /// RSI length. Must be >= 2. Default 14.
    length: usize = 14,
    /// Fast-K stochastic lookback. Must be >= 1. Default 5.
    fast_k_length: usize = 5,
    /// Fast-D smoothing length. Must be >= 1. Default 3.
    fast_d_length: usize = 3,
    /// Moving average type for Fast-D.
    moving_average_type: MovingAverageType = .sma,
    /// When true, EMA seeds with average of first period.
    first_is_average: bool = false,
    /// Bar component. `null` = default (Close).
    bar_component: ?bar_component.BarComponent = null,
    /// Quote component. `null` = default (Mid).
    quote_component: ?quote_component.QuoteComponent = null,
    /// Trade component. `null` = default (Price).
    trade_component: ?trade_component.TradeComponent = null,
};

/// A Fast-D smoother: SMA, EMA, or passthrough (length 1).
const Smoother = union(enum) {
    sma: sma_mod.SimpleMovingAverage,
    ema: ema_mod.ExponentialMovingAverage,
    passthrough: void,

    fn update(self: *Smoother, v: f64) f64 {
        return switch (self.*) {
            .sma => |*s| s.update(v),
            .ema => |*e| e.update(v),
            .passthrough => v,
        };
    }

    fn isPrimed(self: *const Smoother) bool {
        return switch (self.*) {
            .sma => |*s| s.isPrimed(),
            .ema => |*e| e.isPrimed(),
            .passthrough => true,
        };
    }

    fn deinit(self: *Smoother) void {
        switch (self.*) {
            .sma => |*s| s.deinit(),
            .ema, .passthrough => {},
        }
    }
};

/// Tushar Chande's Stochastic Relative Strength Index.
///
/// Applies the Stochastic oscillator formula to RSI values instead of price.
/// Oscillates between 0 and 100. First computes RSI, then applies a stochastic
/// calculation over a rolling window of RSI values to produce Fast-K.
/// Fast-D is a moving average of Fast-K.
pub const StochasticRelativeStrengthIndex = struct {
    line: LineIndicator,
    rsi: rsi_mod.RelativeStrengthIndex,
    rsi_buf: []f64,
    rsi_buffer_index: usize,
    rsi_count: usize,
    fast_k_length: usize,
    fast_d_ma: Smoother,
    fast_k: f64,
    fast_d: f64,
    primed: bool,
    allocator: std.mem.Allocator,
    mnemonic_buf: [96]u8,
    mnemonic_len: usize,
    description_buf: [160]u8,
    description_len: usize,

    pub const Error = error{
        InvalidLength,
        InvalidFastKLength,
        InvalidFastDLength,
        OutOfMemory,
        MnemonicTooLong,
        DescriptionTooLong,
    };

    pub fn init(allocator: std.mem.Allocator, params: StochasticRelativeStrengthIndexParams) Error!StochasticRelativeStrengthIndex {
        if (params.length < 2) return error.InvalidLength;
        if (params.fast_k_length < 1) return error.InvalidFastKLength;
        if (params.fast_d_length < 1) return error.InvalidFastDLength;

        const bc = params.bar_component orelse bar_component.default_bar_component;
        const qc = params.quote_component orelse quote_component.default_quote_component;
        const tc = params.trade_component orelse trade_component.default_trade_component;

        // Create internal RSI (no allocator needed).
        var rsi = rsi_mod.RelativeStrengthIndex.init(.{ .length = params.length }) catch return error.InvalidLength;
        rsi.fixSlices();

        // Create Fast-D smoother.
        var fast_d_ma: Smoother = undefined;
        var ma_label: []const u8 = "SMA";

        if (params.fast_d_length < 2) {
            fast_d_ma = .{ .passthrough = {} };
        } else {
            switch (params.moving_average_type) {
                .ema => {
                    ma_label = "EMA";
                    var ema = ema_mod.ExponentialMovingAverage.initLength(.{
                        .length = params.fast_d_length,
                        .first_is_average = params.first_is_average,
                    }) catch return error.InvalidFastDLength;
                    ema.fixSlices();
                    fast_d_ma = .{ .ema = ema };
                },
                .sma => {
                    var sma = sma_mod.SimpleMovingAverage.init(allocator, .{
                        .length = params.fast_d_length,
                    }) catch return error.InvalidFastDLength;
                    sma.fixSlices();
                    fast_d_ma = .{ .sma = sma };
                },
            }
        }

        // Allocate RSI circular buffer.
        const rsi_buf = allocator.alloc(f64, params.fast_k_length) catch return error.OutOfMemory;

        // Build mnemonic.
        var triple_buf: [64]u8 = undefined;
        const triple = component_triple_mnemonic_mod.componentTripleMnemonic(&triple_buf, bc, qc, tc);

        var mnemonic_buf: [96]u8 = undefined;
        const mnemonic_slice = std.fmt.bufPrint(&mnemonic_buf, "stochrsi({d}/{d}/{s}{d}{s})", .{
            params.length, params.fast_k_length, ma_label, params.fast_d_length, triple,
        }) catch return error.MnemonicTooLong;
        const mnemonic_len = mnemonic_slice.len;

        var description_buf: [160]u8 = undefined;
        const desc_slice = std.fmt.bufPrint(&description_buf, "Stochastic Relative Strength Index {s}", .{
            mnemonic_buf[0..mnemonic_len],
        }) catch return error.DescriptionTooLong;
        const description_len = desc_slice.len;

        return StochasticRelativeStrengthIndex{
            .line = LineIndicator.new(mnemonic_buf[0..mnemonic_len], description_buf[0..description_len], params.bar_component, params.quote_component, params.trade_component),
            .rsi = rsi,
            .rsi_buf = rsi_buf,
            .rsi_buffer_index = 0,
            .rsi_count = 0,
            .fast_k_length = params.fast_k_length,
            .fast_d_ma = fast_d_ma,
            .fast_k = math.nan(f64),
            .fast_d = math.nan(f64),
            .primed = false,
            .allocator = allocator,
            .mnemonic_buf = mnemonic_buf,
            .mnemonic_len = mnemonic_len,
            .description_buf = description_buf,
            .description_len = description_len,
        };
    }

    pub fn deinit(self: *StochasticRelativeStrengthIndex) void {
        self.allocator.free(self.rsi_buf);
        self.fast_d_ma.deinit();
    }

    pub fn fixSlices(self: *StochasticRelativeStrengthIndex) void {
        self.line.mnemonic = self.mnemonic_buf[0..self.mnemonic_len];
        self.line.description = self.description_buf[0..self.description_len];
    }

    /// Core update. Returns (fast_k, fast_d) as a packed struct.
    pub fn update(self: *StochasticRelativeStrengthIndex, sample: f64) [2]f64 {
        if (math.isNan(sample)) return .{ math.nan(f64), math.nan(f64) };

        // Feed to internal RSI.
        const rsi_value = self.rsi.update(sample);
        if (math.isNan(rsi_value)) return .{ self.fast_k, self.fast_d };

        // Store in circular buffer.
        self.rsi_buf[self.rsi_buffer_index] = rsi_value;
        self.rsi_buffer_index = (self.rsi_buffer_index + 1) % self.fast_k_length;
        self.rsi_count += 1;

        // Need at least fast_k_length RSI values.
        if (self.rsi_count < self.fast_k_length) return .{ self.fast_k, self.fast_d };

        // Find min/max of RSI values in window.
        var min_rsi = self.rsi_buf[0];
        var max_rsi = self.rsi_buf[0];
        for (self.rsi_buf[1..self.fast_k_length]) |v| {
            if (v < min_rsi) min_rsi = v;
            if (v > max_rsi) max_rsi = v;
        }

        // Calculate Fast-K.
        const diff = max_rsi - min_rsi;
        if (diff > 0) {
            self.fast_k = 100.0 * (rsi_value - min_rsi) / diff;
        } else {
            self.fast_k = 0.0;
        }

        // Feed Fast-K to Fast-D smoother.
        self.fast_d = self.fast_d_ma.update(self.fast_k);

        if (!self.primed and self.fast_d_ma.isPrimed()) {
            self.primed = true;
        }

        return .{ self.fast_k, self.fast_d };
    }

    pub fn isPrimed(self: *const StochasticRelativeStrengthIndex) bool {
        return self.primed;
    }

    pub fn getMetadata(self: *const StochasticRelativeStrengthIndex, out: *Metadata) void {
        const mnemonic = self.mnemonic_buf[0..self.mnemonic_len];
        const description = self.description_buf[0..self.description_len];

        var out0_mnemonic_buf: [128]u8 = undefined;
        const out0_mnemonic = std.fmt.bufPrint(&out0_mnemonic_buf, "{s} fastK", .{mnemonic}) catch mnemonic;
        var out0_desc_buf: [192]u8 = undefined;
        const out0_desc = std.fmt.bufPrint(&out0_desc_buf, "{s} Fast-K", .{description}) catch description;

        var out1_mnemonic_buf: [128]u8 = undefined;
        const out1_mnemonic = std.fmt.bufPrint(&out1_mnemonic_buf, "{s} fastD", .{mnemonic}) catch mnemonic;
        var out1_desc_buf: [192]u8 = undefined;
        const out1_desc = std.fmt.bufPrint(&out1_desc_buf, "{s} Fast-D", .{description}) catch description;

        const texts = [_]build_metadata_mod.OutputText{
            .{ .mnemonic = out0_mnemonic, .description = out0_desc },
            .{ .mnemonic = out1_mnemonic, .description = out1_desc },
        };

        build_metadata_mod.buildMetadata(
            out,
            Identifier.stochastic_relative_strength_index,
            mnemonic,
            description,
            &texts,
        );
    }

    pub fn updateScalar(self: *StochasticRelativeStrengthIndex, sample: *const Scalar) OutputArray {
        const result = self.update(sample.value);
        var output = OutputArray{};
        output.append(.{ .scalar = .{ .time = sample.time, .value = result[0] } });
        output.append(.{ .scalar = .{ .time = sample.time, .value = result[1] } });
        return output;
    }

    pub fn updateBar(self: *StochasticRelativeStrengthIndex, sample: *const Bar) OutputArray {
        return self.updateScalar(&.{ .time = sample.time, .value = self.line.bar_func(sample.*) });
    }

    pub fn updateQuote(self: *StochasticRelativeStrengthIndex, sample: *const Quote) OutputArray {
        return self.updateScalar(&.{ .time = sample.time, .value = self.line.quote_func(sample.*) });
    }

    pub fn updateTrade(self: *StochasticRelativeStrengthIndex, sample: *const Trade) OutputArray {
        return self.updateScalar(&.{ .time = sample.time, .value = self.line.trade_func(sample.*) });
    }

    /// Returns an Indicator interface backed by this instance.
    pub fn indicator(self: *StochasticRelativeStrengthIndex) indicator_mod.Indicator {
        return .{
            .ptr = @ptrCast(self),
            .vtable = &vtable,
        };
    }

    const vtable = indicator_mod.Indicator.GenVTable(StochasticRelativeStrengthIndex);
};

// ── Tests ──────────────────────────────────────────────────────────────────────

const testing = std.testing;

fn almostEqual(a: f64, b: f64, tolerance: f64) bool {
    if (math.isNan(a) and math.isNan(b)) return true;
    if (math.isNan(a) or math.isNan(b)) return false;
    return @abs(a - b) <= tolerance;
}

const test_input = [_]f64{
    91.500000,  94.815000,  94.375000,  95.095000,  93.780000,  94.625000,  92.530000,  92.750000,  90.315000,  92.470000,
    96.125000,  97.250000,  98.500000,  89.875000,  91.000000,  92.815000,  89.155000,  89.345000,  91.625000,  89.875000,
    88.375000,  87.625000,  84.780000,  83.000000,  83.500000,  81.375000,  84.440000,  89.250000,  86.375000,  86.250000,
    85.250000,  87.125000,  85.815000,  88.970000,  88.470000,  86.875000,  86.815000,  84.875000,  84.190000,  83.875000,
    83.375000,  85.500000,  89.190000,  89.440000,  91.095000,  90.750000,  91.440000,  89.000000,  91.000000,  90.500000,
    89.030000,  88.815000,  84.280000,  83.500000,  82.690000,  84.750000,  85.655000,  86.190000,  88.940000,  89.280000,
    88.625000,  88.500000,  91.970000,  91.500000,  93.250000,  93.500000,  93.155000,  91.720000,  90.000000,  89.690000,
    88.875000,  85.190000,  83.375000,  84.875000,  85.940000,  97.250000,  99.875000,  104.940000, 106.000000, 102.500000,
    102.405000, 104.595000, 106.125000, 106.000000, 106.065000, 104.625000, 108.625000, 109.315000, 110.500000, 112.750000,
    123.000000, 119.625000, 118.750000, 119.250000, 117.940000, 116.440000, 115.190000, 111.875000, 110.595000, 118.125000,
    116.000000, 116.000000, 112.000000, 113.750000, 112.940000, 116.000000, 120.500000, 116.620000, 117.000000, 115.250000,
    114.310000, 115.500000, 115.870000, 120.690000, 120.190000, 120.750000, 124.750000, 123.370000, 122.940000, 122.560000,
    123.120000, 122.560000, 124.620000, 129.250000, 131.000000, 132.250000, 131.000000, 132.810000, 134.000000, 137.380000,
    137.810000, 137.880000, 137.250000, 136.310000, 136.250000, 134.630000, 128.250000, 129.000000, 123.870000, 124.810000,
    123.000000, 126.250000, 128.380000, 125.370000, 125.690000, 122.250000, 119.370000, 118.500000, 123.190000, 123.500000,
    122.190000, 119.310000, 123.310000, 121.120000, 123.370000, 127.370000, 128.500000, 123.870000, 122.940000, 121.750000,
    124.440000, 122.000000, 122.370000, 122.940000, 124.000000, 123.190000, 124.560000, 127.250000, 125.870000, 128.860000,
    132.000000, 130.750000, 134.750000, 135.000000, 132.380000, 133.310000, 131.940000, 130.000000, 125.370000, 130.130000,
    127.120000, 125.190000, 122.000000, 125.000000, 123.000000, 123.500000, 120.060000, 121.000000, 117.750000, 119.870000,
    122.000000, 119.190000, 116.370000, 113.500000, 114.250000, 110.000000, 105.060000, 107.000000, 107.870000, 107.000000,
    107.120000, 107.000000, 91.000000,  93.940000,  93.870000,  95.500000,  93.000000,  94.940000,  98.250000,  96.750000,
    94.810000,  94.370000,  91.560000,  90.250000,  93.940000,  93.620000,  97.000000,  95.000000,  95.870000,  94.060000,
    94.620000,  93.750000,  98.000000,  103.940000, 107.870000, 106.060000, 104.500000, 105.000000, 104.190000, 103.060000,
    103.420000, 105.270000, 111.870000, 116.000000, 116.620000, 118.280000, 113.370000, 109.000000, 109.700000, 109.250000,
    107.000000, 109.190000, 110.000000, 109.200000, 110.120000, 108.000000, 108.620000, 109.750000, 109.810000, 109.000000,
    108.750000, 107.870000,
};

// Test case 1: period=14, fastK=14, fastD=1, SMA.
test "StochRSI 14/14/1 SMA" {
    const tolerance = 1e-4;
    var ind = try StochasticRelativeStrengthIndex.init(testing.allocator, .{
        .length = 14,
        .fast_k_length = 14,
        .fast_d_length = 1,
    });
    defer ind.deinit();
    ind.fixSlices();

    // First 27 values should produce NaN FastK.
    for (0..27) |i| {
        const result = ind.update(test_input[i]);
        try testing.expect(math.isNan(result[0]));
    }

    // Index 27: first value.
    const result27 = ind.update(test_input[27]);
    try testing.expect(!math.isNan(result27[0]));
    try testing.expect(almostEqual(result27[0], 94.156709, tolerance));
    try testing.expect(almostEqual(result27[1], 94.156709, tolerance));

    // Feed remaining and check last value.
    for (28..251) |i| {
        _ = ind.update(test_input[i]);
    }
    const result251 = ind.update(test_input[251]);
    try testing.expect(almostEqual(result251[0], 0.0, tolerance));
    try testing.expect(almostEqual(result251[1], 0.0, tolerance));
}

// Test case 2: period=14, fastK=45, fastD=1, SMA.
test "StochRSI 14/45/1 SMA" {
    const tolerance = 1e-4;
    var ind = try StochasticRelativeStrengthIndex.init(testing.allocator, .{
        .length = 14,
        .fast_k_length = 45,
        .fast_d_length = 1,
    });
    defer ind.deinit();
    ind.fixSlices();

    // First 58 values should produce NaN.
    for (0..58) |i| {
        const result = ind.update(test_input[i]);
        try testing.expect(math.isNan(result[0]));
    }

    // Index 58: first value.
    const result58 = ind.update(test_input[58]);
    try testing.expect(!math.isNan(result58[0]));
    try testing.expect(almostEqual(result58[0], 79.729186, tolerance));
    try testing.expect(almostEqual(result58[1], 79.729186, tolerance));

    // Feed remaining and check last value.
    for (59..251) |i| {
        _ = ind.update(test_input[i]);
    }
    const result251 = ind.update(test_input[251]);
    try testing.expect(almostEqual(result251[0], 48.1550743, tolerance));
    try testing.expect(almostEqual(result251[1], 48.1550743, tolerance));
}

// Test case 3: period=11, fastK=13, fastD=16, SMA.
test "StochRSI 11/13/16 SMA" {
    const tolerance = 1e-3;
    var ind = try StochasticRelativeStrengthIndex.init(testing.allocator, .{
        .length = 11,
        .fast_k_length = 13,
        .fast_d_length = 16,
    });
    defer ind.deinit();
    ind.fixSlices();

    // Feed first 38 values.
    for (0..38) |i| {
        _ = ind.update(test_input[i]);
    }

    // Index 38: first primed value.
    const result38 = ind.update(test_input[38]);
    try testing.expect(almostEqual(result38[0], 5.25947, tolerance));
    try testing.expect(almostEqual(result38[1], 57.1711, tolerance));
    try testing.expect(ind.isPrimed());

    // Feed remaining and check last value.
    for (39..251) |i| {
        _ = ind.update(test_input[i]);
    }
    const result251 = ind.update(test_input[251]);
    try testing.expect(almostEqual(result251[0], 0.0, tolerance));
    try testing.expect(almostEqual(result251[1], 15.7303, tolerance));
}

test "StochRSI isPrimed" {
    var ind = try StochasticRelativeStrengthIndex.init(testing.allocator, .{
        .length = 14,
        .fast_k_length = 14,
        .fast_d_length = 1,
    });
    defer ind.deinit();
    ind.fixSlices();

    try testing.expect(!ind.isPrimed());

    for (0..27) |i| {
        _ = ind.update(test_input[i]);
        try testing.expect(!ind.isPrimed());
    }

    _ = ind.update(test_input[27]);
    try testing.expect(ind.isPrimed());
}

test "StochRSI NaN input" {
    var ind = try StochasticRelativeStrengthIndex.init(testing.allocator, .{
        .length = 14,
        .fast_k_length = 14,
        .fast_d_length = 1,
    });
    defer ind.deinit();
    ind.fixSlices();

    const result = ind.update(math.nan(f64));
    try testing.expect(math.isNan(result[0]));
    try testing.expect(math.isNan(result[1]));
}

test "StochRSI metadata" {
    var ind = try StochasticRelativeStrengthIndex.init(testing.allocator, .{
        .length = 14,
        .fast_k_length = 14,
        .fast_d_length = 3,
    });
    defer ind.deinit();
    ind.fixSlices();

    var meta: Metadata = undefined;
    ind.getMetadata(&meta);

    try testing.expectEqual(Identifier.stochastic_relative_strength_index, meta.identifier);
    try testing.expectEqualStrings("stochrsi(14/14/SMA3)", meta.mnemonic);
    try testing.expectEqual(@as(usize, 2), meta.outputs_len);
    try testing.expectEqual(@as(i32, 1), meta.outputs_buf[0].kind);
    try testing.expectEqual(@as(i32, 2), meta.outputs_buf[1].kind);
}

test "StochRSI invalid params" {
    // Length too small.
    try testing.expectError(error.InvalidLength, StochasticRelativeStrengthIndex.init(testing.allocator, .{
        .length = 1,
        .fast_k_length = 14,
        .fast_d_length = 3,
    }));

    // FastK too small.
    try testing.expectError(error.InvalidFastKLength, StochasticRelativeStrengthIndex.init(testing.allocator, .{
        .length = 14,
        .fast_k_length = 0,
        .fast_d_length = 3,
    }));

    // FastD too small.
    try testing.expectError(error.InvalidFastDLength, StochasticRelativeStrengthIndex.init(testing.allocator, .{
        .length = 14,
        .fast_k_length = 14,
        .fast_d_length = 0,
    }));
}
