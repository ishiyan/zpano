const std = @import("std");
const math = std.math;


const entities = @import("entities");
const Bar = entities.Bar;
const Quote = entities.Quote;
const Trade = entities.Trade;
const Scalar = entities.Scalar;
const indicator_mod = @import("../../core/indicator.zig");
const build_metadata_mod = @import("../../core/build_metadata.zig");
const identifier_mod = @import("../../core/identifier.zig");
const metadata_mod = @import("../../core/metadata.zig");

const dip_mod = @import("../directional_indicator_plus/directional_indicator_plus.zig");
const DirectionalIndicatorPlus = dip_mod.DirectionalIndicatorPlus;
const dim_mod = @import("../directional_indicator_minus/directional_indicator_minus.zig");
const DirectionalIndicatorMinus = dim_mod.DirectionalIndicatorMinus;

const OutputArray = indicator_mod.OutputArray;
const Identifier = identifier_mod.Identifier;
const Metadata = metadata_mod.Metadata;

const epsilon = 1e-8;

/// Enumerates the outputs of the Directional Movement Index indicator.
pub const DirectionalMovementIndexOutput = enum(u8) {
    /// The scalar value of the directional movement index (DX).
    value = 1,
    /// The scalar value of the directional indicator plus (+DI).
    directional_indicator_plus = 2,
    /// The scalar value of the directional indicator minus (-DI).
    directional_indicator_minus = 3,
    /// The scalar value of the directional movement plus (+DM).
    directional_movement_plus = 4,
    /// The scalar value of the directional movement minus (-DM).
    directional_movement_minus = 5,
    /// The scalar value of the average true range (ATR).
    average_true_range = 6,
    /// The scalar value of the true range (TR).
    true_range = 7,
};

/// Welles Wilder's Directional Movement Index (DX).
///
/// The directional movement index measures the strength of a trend by comparing
/// the positive and negative directional indicators. It is calculated as:
///   DX = 100 * |+DI - -DI| / (+DI + -DI)
///
/// where +DI is the directional indicator plus and -DI is the directional
/// indicator minus, both computed over the same length.
pub const DirectionalMovementIndex = struct {
    length: i32,
    value: f64,
    directional_indicator_plus: DirectionalIndicatorPlus,
    directional_indicator_minus: DirectionalIndicatorMinus,

    const mnemonic_str = "dx";
    const description_str = "Directional Movement Index";

    pub const Error = error{
        InvalidLength,
        OutOfMemory,
    };

    pub fn init(allocator: std.mem.Allocator, params: struct { length: i32 = 14 }) Error!DirectionalMovementIndex {
        if (params.length < 1) return Error.InvalidLength;

        const dip = DirectionalIndicatorPlus.init(allocator, .{ .length = params.length }) catch |e| switch (e) {
            error.InvalidLength => return Error.InvalidLength,
            error.OutOfMemory => return Error.OutOfMemory,
        };

        const dim = DirectionalIndicatorMinus.init(allocator, .{ .length = params.length }) catch |e| switch (e) {
            error.InvalidLength => return Error.InvalidLength,
            error.OutOfMemory => return Error.OutOfMemory,
        };

        return .{
            .length = params.length,
            .value = math.nan(f64),
            .directional_indicator_plus = dip,
            .directional_indicator_minus = dim,
        };
    }

    pub fn deinit(self: *DirectionalMovementIndex) void {
        self.directional_indicator_plus.deinit();
        self.directional_indicator_minus.deinit();
    }

    pub fn fixSlices(_: *DirectionalMovementIndex) void {}

    /// Update given close, high, low values.
    pub fn update(self: *DirectionalMovementIndex, close: f64, high: f64, low: f64) f64 {
        if (math.isNan(close) or math.isNan(high) or math.isNan(low)) return math.nan(f64);

        const dip_value = self.directional_indicator_plus.update(close, high, low);
        const dim_value = self.directional_indicator_minus.update(close, high, low);

        if (self.directional_indicator_plus.isPrimed() and self.directional_indicator_minus.isPrimed()) {
            const sum = dip_value + dim_value;

            if (@abs(sum) < epsilon) {
                self.value = 0;
            } else {
                self.value = 100.0 * @abs(dip_value - dim_value) / sum;
            }

            return self.value;
        }

        return math.nan(f64);
    }

    /// Update using a single sample value as substitute for high, low, close.
    pub fn updateSample(self: *DirectionalMovementIndex, sample: f64) f64 {
        return self.update(sample, sample, sample);
    }

    pub fn isPrimed(self: *const DirectionalMovementIndex) bool {
        return self.directional_indicator_plus.isPrimed() and self.directional_indicator_minus.isPrimed();
    }

    pub fn getMetadata(_: *const DirectionalMovementIndex, out: *Metadata) void {
        build_metadata_mod.buildMetadata(out, Identifier.directional_movement_index, mnemonic_str, description_str, &.{
            .{ .mnemonic = mnemonic_str, .description = description_str },
            .{ .mnemonic = "+di", .description = "Directional Indicator Plus" },
            .{ .mnemonic = "-di", .description = "Directional Indicator Minus" },
            .{ .mnemonic = "+dm", .description = "Directional Movement Plus" },
            .{ .mnemonic = "-dm", .description = "Directional Movement Minus" },
            .{ .mnemonic = "atr", .description = "Average True Range" },
            .{ .mnemonic = "tr", .description = "True Range" },
        });
    }

    fn makeOutput(self: *const DirectionalMovementIndex, time: i64) OutputArray {
        var out = OutputArray{};
        out.append(.{ .scalar = Scalar{ .time = time, .value = self.value } });
        return out;
    }

    pub fn updateScalar(self: *DirectionalMovementIndex, sample: *const Scalar) OutputArray {
        _ = self.update(sample.value, sample.value, sample.value);
        return self.makeOutput(sample.time);
    }

    pub fn updateBar(self: *DirectionalMovementIndex, sample: *const Bar) OutputArray {
        _ = self.update(sample.close, sample.high, sample.low);
        return self.makeOutput(sample.time);
    }

    pub fn updateQuote(self: *DirectionalMovementIndex, sample: *const Quote) OutputArray {
        const mid = (sample.bid_price + sample.ask_price) / 2.0;
        _ = self.update(mid, mid, mid);
        return self.makeOutput(sample.time);
    }

    pub fn updateTrade(self: *DirectionalMovementIndex, sample: *const Trade) OutputArray {
        _ = self.update(sample.price, sample.price, sample.price);
        return self.makeOutput(sample.time);
    }

    // --- Indicator interface ---

    pub fn indicator(self: *DirectionalMovementIndex) indicator_mod.Indicator {
        return indicator_mod.Indicator{
            .ptr = @ptrCast(self),
            .vtable = &.{
                .isPrimed = vtableIsPrimed,
                .metadata = vtableMetadata,
                .updateScalar = vtableUpdateScalar,
                .updateBar = vtableUpdateBar,
                .updateQuote = vtableUpdateQuote,
                .updateTrade = vtableUpdateTrade,
            },
        };
    }

    fn vtableIsPrimed(ptr: *const anyopaque) bool {
        const self: *const DirectionalMovementIndex = @ptrCast(@alignCast(ptr));
        return self.isPrimed();
    }

    fn vtableMetadata(ptr: *const anyopaque, out: *Metadata) void {
        const self: *const DirectionalMovementIndex = @ptrCast(@alignCast(ptr));
        self.getMetadata(out);
    }

    fn vtableUpdateScalar(ptr: *anyopaque, sample: *const Scalar) OutputArray {
        const self: *DirectionalMovementIndex = @ptrCast(@alignCast(ptr));
        return self.updateScalar(sample);
    }

    fn vtableUpdateBar(ptr: *anyopaque, sample: *const Bar) OutputArray {
        const self: *DirectionalMovementIndex = @ptrCast(@alignCast(ptr));
        return self.updateBar(sample);
    }

    fn vtableUpdateQuote(ptr: *anyopaque, sample: *const Quote) OutputArray {
        const self: *DirectionalMovementIndex = @ptrCast(@alignCast(ptr));
        return self.updateQuote(sample);
    }

    fn vtableUpdateTrade(ptr: *anyopaque, sample: *const Trade) OutputArray {
        const self: *DirectionalMovementIndex = @ptrCast(@alignCast(ptr));
        return self.updateTrade(sample);
    }
};

// ============================================================================
// Tests
// ============================================================================

const testing = std.testing;
const testdata = @import("testdata.zig");


fn almostEqual(a: f64, b: f64, tolerance: f64) bool {
    return @abs(a - b) <= tolerance;
}

// TA-Lib test data (252 entries).
// Expected DX14 (length=14), 252 entries.
test "DirectionalMovementIndex update length=14" {
    const tolerance = 1e-8;
    var dx = try DirectionalMovementIndex.init(testing.allocator, .{ .length = 14 });
    defer dx.deinit();

    for (0..testdata.test_input_close.len) |i| {
        const act = dx.update(testdata.test_input_close[i], testdata.test_input_high[i], testdata.test_input_low[i]);
        const exp = testdata.test_expected_dx14[i];

        if (math.isNan(exp)) {
            try testing.expect(math.isNan(act));
        } else {
            try testing.expect(!math.isNan(act));
            try testing.expect(almostEqual(act, exp, tolerance));
        }
    }
}

test "DirectionalMovementIndex isPrimed length=14" {
    var dx = try DirectionalMovementIndex.init(testing.allocator, .{ .length = 14 });
    defer dx.deinit();

    for (0..14) |i| {
        _ = dx.update(testdata.test_input_close[i], testdata.test_input_high[i], testdata.test_input_low[i]);
        try testing.expect(!dx.isPrimed());
    }

    _ = dx.update(testdata.test_input_close[14], testdata.test_input_high[14], testdata.test_input_low[14]);
    try testing.expect(dx.isPrimed());
}

test "DirectionalMovementIndex constructor validation" {
    try testing.expectError(error.InvalidLength, DirectionalMovementIndex.init(testing.allocator, .{ .length = 0 }));
    try testing.expectError(error.InvalidLength, DirectionalMovementIndex.init(testing.allocator, .{ .length = -8 }));

    var dx = try DirectionalMovementIndex.init(testing.allocator, .{ .length = 14 });
    defer dx.deinit();
    try testing.expect(!dx.isPrimed());
}

test "DirectionalMovementIndex NaN passthrough" {
    var dx = try DirectionalMovementIndex.init(testing.allocator, .{ .length = 14 });
    defer dx.deinit();

    try testing.expect(math.isNan(dx.update(math.nan(f64), 1, 1)));
    try testing.expect(math.isNan(dx.update(1, math.nan(f64), 1)));
    try testing.expect(math.isNan(dx.update(1, 1, math.nan(f64))));
    try testing.expect(math.isNan(dx.updateSample(math.nan(f64))));
}

test "DirectionalMovementIndex metadata" {
    var dx = try DirectionalMovementIndex.init(testing.allocator, .{ .length = 14 });
    defer dx.deinit();
    var meta: Metadata = undefined;
    dx.getMetadata(&meta);

    try testing.expectEqual(Identifier.directional_movement_index, meta.identifier);
    try testing.expectEqualStrings("dx", meta.mnemonic);
    try testing.expectEqual(@as(usize, 7), meta.outputs_len);
}

test "DirectionalMovementIndex updateBar" {
    var dx = try DirectionalMovementIndex.init(testing.allocator, .{ .length = 14 });
    defer dx.deinit();

    for (0..14) |i| {
        _ = dx.update(testdata.test_input_close[i], testdata.test_input_high[i], testdata.test_input_low[i]);
    }

    const bar = Bar{
        .time = 1000,
        .open = 91,
        .high = testdata.test_input_high[14],
        .low = testdata.test_input_low[14],
        .close = testdata.test_input_close[14],
        .volume = 1000,
    };
    const out = dx.updateBar(&bar);
    try testing.expect(!math.isNan(out.slice()[0].scalar.value));
}
