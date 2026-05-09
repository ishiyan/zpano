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

const OutputArray = indicator_mod.OutputArray;
const Identifier = identifier_mod.Identifier;
const Metadata = metadata_mod.Metadata;

/// Enumerates the outputs of the center of gravity oscillator.
pub const CenterOfGravityOscillatorOutput = enum(u8) {
    value = 1,
    trigger = 2,
};

/// Parameters to create a CenterOfGravityOscillator.
pub const Params = struct {
    /// Length (window size). Must be >= 1. Default is 10.
    length: i32 = 10,
    bar_component: ?bar_component.BarComponent = null,
    quote_component: ?quote_component.QuoteComponent = null,
    trade_component: ?trade_component.TradeComponent = null,
};

/// Ehler's Center of Gravity oscillator (COG).
///
///   CGi = Sigma((i+1) * Price_i) / Sigma(Price_i), where i = 0...l-1.
///
/// The center of gravity in a FIR filter is the position of the average price
/// within the filter window length. It has essentially zero lag and retains
/// relative cycle amplitude.
///
/// Two outputs: oscillator value and trigger (previous oscillator value).
pub const CenterOfGravityOscillator = struct {
    allocator: std.mem.Allocator,
    value: f64,
    value_previous: f64,
    denominator_sum: f64,
    length: usize,
    length_min_one: usize,
    window_count: usize,
    window: []f64,
    primed: bool,
    bar_func: bar_component.BarFunc,
    quote_func: quote_component.QuoteFunc,
    trade_func: trade_component.TradeFunc,
    mnemonic_buf: [128]u8,
    mnemonic_len: usize,
    description_buf: [192]u8,
    description_len: usize,
    mnemonic_trig_buf: [128]u8,
    mnemonic_trig_len: usize,
    description_trig_buf: [192]u8,
    description_trig_len: usize,

    pub fn init(allocator: std.mem.Allocator, params: Params) !CenterOfGravityOscillator {
        if (params.length < 1) {
            return error.InvalidLength;
        }

        const length: usize = @intCast(params.length);

        const window = try allocator.alloc(f64, length);
        @memset(window, 0.0);

        const bc = params.bar_component orelse bar_component.BarComponent.median;
        const qc = params.quote_component orelse quote_component.default_quote_component;
        const tc = params.trade_component orelse trade_component.default_trade_component;

        // Build mnemonics.
        var triple_buf: [64]u8 = undefined;
        const triple = component_triple_mnemonic_mod.componentTripleMnemonic(
            &triple_buf,
            params.bar_component orelse bar_component.BarComponent.median,
            params.quote_component orelse quote_component.default_quote_component,
            params.trade_component orelse trade_component.default_trade_component,
        );

        var mnemonic_buf: [128]u8 = undefined;
        const mn = std.fmt.bufPrint(&mnemonic_buf, "cog({d}{s})", .{ params.length, triple }) catch
            return error.MnemonicTooLong;
        const mnemonic_len = mn.len;

        var description_buf: [192]u8 = undefined;
        const desc = std.fmt.bufPrint(&description_buf, "Center of Gravity oscillator {s}", .{mn}) catch
            return error.MnemonicTooLong;
        const description_len = desc.len;

        var mnemonic_trig_buf: [128]u8 = undefined;
        const mn_trig = std.fmt.bufPrint(&mnemonic_trig_buf, "cogTrig({d}{s})", .{ params.length, triple }) catch
            return error.MnemonicTooLong;
        const mnemonic_trig_len = mn_trig.len;

        var description_trig_buf: [192]u8 = undefined;
        const desc_trig = std.fmt.bufPrint(&description_trig_buf, "Center of Gravity trigger {s}", .{mn_trig}) catch
            return error.MnemonicTooLong;
        const description_trig_len = desc_trig.len;

        return .{
            .allocator = allocator,
            .value = math.nan(f64),
            .value_previous = math.nan(f64),
            .denominator_sum = 0.0,
            .length = length,
            .length_min_one = length - 1,
            .window_count = 0,
            .window = window,
            .primed = false,
            .bar_func = bar_component.componentValue(bc),
            .quote_func = quote_component.componentValue(qc),
            .trade_func = trade_component.componentValue(tc),
            .mnemonic_buf = mnemonic_buf,
            .mnemonic_len = mnemonic_len,
            .description_buf = description_buf,
            .description_len = description_len,
            .mnemonic_trig_buf = mnemonic_trig_buf,
            .mnemonic_trig_len = mnemonic_trig_len,
            .description_trig_buf = description_trig_buf,
            .description_trig_len = description_trig_len,
        };
    }

    pub fn deinit(self: *CenterOfGravityOscillator) void {
        self.allocator.free(self.window);
    }

    pub fn fixSlices(self: *CenterOfGravityOscillator) void {
        _ = self;
    }

    /// Update the center of gravity oscillator given the next sample.
    pub fn update(self: *CenterOfGravityOscillator, sample: f64) f64 {
        if (math.isNan(sample)) {
            return math.nan(f64);
        }

        if (self.primed) {
            self.value_previous = self.value;
            self.value = self.calculate(sample);
            return self.value;
        }

        // Not primed.
        if (self.length > self.window_count) {
            self.denominator_sum += sample;
            self.window[self.window_count] = sample;

            if (self.length_min_one == self.window_count) {
                var sum: f64 = 0.0;
                if (@abs(self.denominator_sum) > math.floatMin(f64)) {
                    for (0..self.length) |i| {
                        sum += @as(f64, @floatFromInt(1 + i)) * self.window[i];
                    }
                    sum /= self.denominator_sum;
                }
                self.value_previous = sum;
            }
        } else {
            self.value = self.calculate(sample);
            self.primed = true;
            self.window_count += 1;
            return self.value;
        }

        self.window_count += 1;
        return math.nan(f64);
    }

    fn calculate(self: *CenterOfGravityOscillator, sample: f64) f64 {
        self.denominator_sum += sample - self.window[0];

        // Shift window left.
        std.mem.copyForwards(f64, self.window[0..self.length_min_one], self.window[1..self.length]);

        self.window[self.length_min_one] = sample;

        var sum: f64 = 0.0;
        if (@abs(self.denominator_sum) > math.floatMin(f64)) {
            for (0..self.length) |i| {
                sum += @as(f64, @floatFromInt(1 + i)) * self.window[i];
            }
            sum /= self.denominator_sum;
        }

        return sum;
    }

    pub fn isPrimed(self: *const CenterOfGravityOscillator) bool {
        return self.primed;
    }

    pub fn getMetadata(self: *const CenterOfGravityOscillator, out: *Metadata) void {
        build_metadata_mod.buildMetadata(
            out,
            .center_of_gravity_oscillator,
            self.mnemonic_buf[0..self.mnemonic_len],
            self.description_buf[0..self.description_len],
            &[_]build_metadata_mod.OutputText{
                .{ .mnemonic = self.mnemonic_buf[0..self.mnemonic_len], .description = self.description_buf[0..self.description_len] },
                .{ .mnemonic = self.mnemonic_trig_buf[0..self.mnemonic_trig_len], .description = self.description_trig_buf[0..self.description_trig_len] },
            },
        );
    }

    pub fn updateScalar(self: *CenterOfGravityOscillator, sample: *const Scalar) OutputArray {
        return self.updateEntity(sample.time, sample.value);
    }

    pub fn updateBar(self: *CenterOfGravityOscillator, sample: *const Bar) OutputArray {
        return self.updateEntity(sample.time, self.bar_func(sample.*));
    }

    pub fn updateQuote(self: *CenterOfGravityOscillator, sample: *const Quote) OutputArray {
        return self.updateEntity(sample.time, self.quote_func(sample.*));
    }

    pub fn updateTrade(self: *CenterOfGravityOscillator, sample: *const Trade) OutputArray {
        return self.updateEntity(sample.time, self.trade_func(sample.*));
    }

    fn updateEntity(self: *CenterOfGravityOscillator, time: i64, sample: f64) OutputArray {
        const cog = self.update(sample);
        var trig = self.value_previous;
        if (math.isNan(cog)) {
            trig = math.nan(f64);
        }

        var out = OutputArray{};
        out.append(.{ .scalar = .{ .time = time, .value = cog } });
        out.append(.{ .scalar = .{ .time = time, .value = trig } });
        return out;
    }

    /// Returns an Indicator interface backed by this instance.
    pub fn indicator(self: *CenterOfGravityOscillator) indicator_mod.Indicator {
        return .{
            .ptr = @ptrCast(self),
            .vtable = &vtable,
        };
    }

    const vtable = indicator_mod.Indicator.GenVTable(CenterOfGravityOscillator);
};

// --- Tests ---
const testing = std.testing;
const testdata = @import("testdata.zig");


fn almostEqual(a: f64, b: f64, tolerance: f64) bool {
    return @abs(a - b) <= tolerance;
}

// 252-entry input data from test_Cog.xls, (high + low)/2 median price.
// Expected COG values, 252 entries.
// Expected trigger values, 252 entries.
// High price data for bar tests, 252 entries.
// Low price data for bar tests, 252 entries.
test "COG update value" {
    const tolerance = 1e-8;
    const l_primed = 10;

    var cog = try CenterOfGravityOscillator.init(testing.allocator, .{ .length = 10 });
    defer cog.deinit();
    cog.fixSlices();

    for (0..l_primed) |i| {
        try testing.expect(math.isNan(cog.update(testdata.test_input[i])));
    }

    for (l_primed..testdata.test_input.len) |i| {
        const act = cog.update(testdata.test_input[i]);
        try testing.expect(almostEqual(act, testdata.test_expected_cog[i], tolerance));
    }

    // NaN passthrough.
    try testing.expect(math.isNan(cog.update(math.nan(f64))));
}

test "COG update trigger" {
    const tolerance = 1e-8;
    const l_primed = 10;

    var cog = try CenterOfGravityOscillator.init(testing.allocator, .{ .length = 10 });
    defer cog.deinit();
    cog.fixSlices();

    for (0..l_primed) |_i| {
        _ = cog.update(testdata.test_input[_i]);
    }

    for (l_primed..testdata.test_input.len) |i| {
        _ = cog.update(testdata.test_input[i]);
        try testing.expect(almostEqual(cog.value_previous, testdata.test_expected_trigger[i], tolerance));
    }
}

test "COG isPrimed" {
    var cog = try CenterOfGravityOscillator.init(testing.allocator, .{ .length = 10 });
    defer cog.deinit();
    cog.fixSlices();

    try testing.expect(!cog.isPrimed());

    const l_primed = 10;
    for (0..l_primed) |i| {
        _ = cog.update(testdata.test_input[i]);
        try testing.expect(!cog.isPrimed());
    }

    // 11th update: primed.
    _ = cog.update(testdata.test_input[l_primed]);
    try testing.expect(cog.isPrimed());
}

test "COG metadata" {
    var cog = try CenterOfGravityOscillator.init(testing.allocator, .{ .length = 10 });
    defer cog.deinit();
    cog.fixSlices();
    var meta: Metadata = undefined;
    cog.getMetadata(&meta);

    try testing.expectEqual(Identifier.center_of_gravity_oscillator, meta.identifier);
    try testing.expectEqualStrings("cog(10, hl/2)", meta.mnemonic);
    try testing.expectEqual(@as(usize, 2), meta.outputs_len);
}

test "COG constructor" {
    var cog = try CenterOfGravityOscillator.init(testing.allocator, .{ .length = 10 });
    defer cog.deinit();

    var cog2 = try CenterOfGravityOscillator.init(testing.allocator, .{ .length = 1 });
    defer cog2.deinit();

    try testing.expectError(error.InvalidLength, CenterOfGravityOscillator.init(testing.allocator, .{ .length = 0 }));
    try testing.expectError(error.InvalidLength, CenterOfGravityOscillator.init(testing.allocator, .{ .length = -8 }));
}

test "COG updateScalar" {
    const tolerance = 1e-8;

    var cog = try CenterOfGravityOscillator.init(testing.allocator, .{ .length = 10 });
    defer cog.deinit();
    cog.fixSlices();

    for (0..testdata.test_input.len) |i| {
        const s = Scalar{ .time = @intCast(i), .value = testdata.test_input[i] };
        const out = cog.updateScalar(&s);
        const outputs = out.slice();
        try testing.expectEqual(@as(usize, 2), outputs.len);

        if (i < 10) {
            try testing.expect(math.isNan(outputs[0].scalar.value));
            try testing.expect(math.isNan(outputs[1].scalar.value));
        } else {
            try testing.expect(almostEqual(outputs[0].scalar.value, testdata.test_expected_cog[i], tolerance));
            try testing.expect(almostEqual(outputs[1].scalar.value, testdata.test_expected_trigger[i], tolerance));
        }
    }
}

test "COG updateBar" {
    const tolerance = 1e-8;

    var cog = try CenterOfGravityOscillator.init(testing.allocator, .{ .length = 10 });
    defer cog.deinit();
    cog.fixSlices();

    for (0..testdata.test_input.len) |i| {
        const b = Bar{ .time = @intCast(i), .high = testdata.test_input_high[i], .low = testdata.test_input_low[i], .open = 0, .close = 0, .volume = 0 };
        const out = cog.updateBar(&b);
        const outputs = out.slice();
        try testing.expectEqual(@as(usize, 2), outputs.len);

        if (i < 10) {
            try testing.expect(math.isNan(outputs[0].scalar.value));
            try testing.expect(math.isNan(outputs[1].scalar.value));
        } else {
            try testing.expect(almostEqual(outputs[0].scalar.value, testdata.test_expected_cog[i], tolerance));
            try testing.expect(almostEqual(outputs[1].scalar.value, testdata.test_expected_trigger[i], tolerance));
        }
    }
}
