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
const testdata = @import("testdata.zig");


fn almostEqual(a: f64, b: f64, tolerance: f64) bool {
    if (math.isNan(a) and math.isNan(b)) return true;
    if (math.isNan(a) or math.isNan(b)) return false;
    return @abs(a - b) <= tolerance;
}

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
        const result = ind.update(testdata.test_input[i]);
        try testing.expect(math.isNan(result[0]));
    }

    // Index 27: first value.
    const result27 = ind.update(testdata.test_input[27]);
    try testing.expect(!math.isNan(result27[0]));
    try testing.expect(almostEqual(result27[0], 94.156709, tolerance));
    try testing.expect(almostEqual(result27[1], 94.156709, tolerance));

    // Feed remaining and check last value.
    for (28..251) |i| {
        _ = ind.update(testdata.test_input[i]);
    }
    const result251 = ind.update(testdata.test_input[251]);
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
        const result = ind.update(testdata.test_input[i]);
        try testing.expect(math.isNan(result[0]));
    }

    // Index 58: first value.
    const result58 = ind.update(testdata.test_input[58]);
    try testing.expect(!math.isNan(result58[0]));
    try testing.expect(almostEqual(result58[0], 79.729186, tolerance));
    try testing.expect(almostEqual(result58[1], 79.729186, tolerance));

    // Feed remaining and check last value.
    for (59..251) |i| {
        _ = ind.update(testdata.test_input[i]);
    }
    const result251 = ind.update(testdata.test_input[251]);
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
        _ = ind.update(testdata.test_input[i]);
    }

    // Index 38: first primed value.
    const result38 = ind.update(testdata.test_input[38]);
    try testing.expect(almostEqual(result38[0], 5.25947, tolerance));
    try testing.expect(almostEqual(result38[1], 57.1711, tolerance));
    try testing.expect(ind.isPrimed());

    // Feed remaining and check last value.
    for (39..251) |i| {
        _ = ind.update(testdata.test_input[i]);
    }
    const result251 = ind.update(testdata.test_input[251]);
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
        _ = ind.update(testdata.test_input[i]);
        try testing.expect(!ind.isPrimed());
    }

    _ = ind.update(testdata.test_input[27]);
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
