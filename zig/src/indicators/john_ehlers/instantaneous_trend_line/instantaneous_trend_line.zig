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

/// Enumerates the outputs of the instantaneous trend line indicator.
pub const InstantaneousTrendLineOutput = enum(u8) {
    value = 1,
    trigger = 2,
};

/// Parameters to create an ITL based on length.
pub const LengthParams = struct {
    /// Length ℓ of the instantaneous trend line (α = 2/(ℓ+1)). Must be >= 1. Default is 28.
    length: i32 = 28,
    bar_component: ?bar_component.BarComponent = null,
    quote_component: ?quote_component.QuoteComponent = null,
    trade_component: ?trade_component.TradeComponent = null,
};

/// Parameters to create an ITL based on smoothing factor.
pub const SmoothingFactorParams = struct {
    /// Smoothing factor α in [0, 1]. Default is 0.07.
    smoothing_factor: f64 = 0.07,
    bar_component: ?bar_component.BarComponent = null,
    quote_component: ?quote_component.QuoteComponent = null,
    trade_component: ?trade_component.TradeComponent = null,
};

/// Ehler's Instantaneous Trend Line (iTrend).
///
///   H(z) = ((α-α²/4) + α²z⁻¹/2 - (α-3α²/4)z⁻²) / (1 - 2(1-α)z⁻¹ + (1-α)²z⁻²)
///
/// Two outputs: trend line value and trigger line.
/// Primed after 5 samples.
pub const InstantaneousTrendLine = struct {
    coeff1: f64,
    coeff2: f64,
    coeff3: f64,
    coeff4: f64,
    coeff5: f64,
    count: usize,
    previous_sample1: f64,
    previous_sample2: f64,
    previous_trend_line1: f64,
    previous_trend_line2: f64,
    trend_line: f64,
    trigger_line: f64,
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

    pub fn initLength(params: LengthParams) !InstantaneousTrendLine {
        if (params.length < 1) {
            return error.InvalidLength;
        }

        const alpha: f64 = 2.0 / @as(f64, @floatFromInt(1 + params.length));
        return initCommon(params.length, alpha, params.bar_component, params.quote_component, params.trade_component);
    }

    pub fn initSmoothingFactor(params: SmoothingFactorParams) !InstantaneousTrendLine {
        const alpha = params.smoothing_factor;
        if (alpha < 0.0 or alpha > 1.0) {
            return error.InvalidSmoothingFactor;
        }

        const epsilon: f64 = 0.00000001;
        const length: i32 = if (alpha < epsilon)
            std.math.maxInt(i32)
        else
            @as(i32, @intFromFloat(@round(2.0 / alpha))) - 1;

        return initCommon(length, alpha, params.bar_component, params.quote_component, params.trade_component);
    }

    fn initCommon(
        length: i32,
        alpha: f64,
        bc_opt: ?bar_component.BarComponent,
        qc_opt: ?quote_component.QuoteComponent,
        tc_opt: ?trade_component.TradeComponent,
    ) !InstantaneousTrendLine {
        const bc = bc_opt orelse bar_component.BarComponent.median;
        const qc = qc_opt orelse quote_component.default_quote_component;
        const tc = tc_opt orelse trade_component.default_trade_component;

        // Calculate coefficients.
        const a2 = alpha * alpha;
        const c1 = alpha - a2 / 4.0;
        const c2 = a2 / 2.0;
        const c3 = -(alpha - 3.0 * a2 / 4.0);
        const x = 1.0 - alpha;
        const c4 = 2.0 * x;
        const c5 = -(x * x);

        // Build mnemonics.
        var triple_buf: [64]u8 = undefined;
        const triple = component_triple_mnemonic_mod.componentTripleMnemonic(
            &triple_buf,
            bc_opt orelse bar_component.BarComponent.median,
            qc_opt orelse quote_component.default_quote_component,
            tc_opt orelse trade_component.default_trade_component,
        );

        var mnemonic_buf: [128]u8 = undefined;
        const mn = std.fmt.bufPrint(&mnemonic_buf, "iTrend({d}{s})", .{ length, triple }) catch
            return error.MnemonicTooLong;
        const mnemonic_len = mn.len;

        var description_buf: [192]u8 = undefined;
        const desc = std.fmt.bufPrint(&description_buf, "Instantaneous Trend Line {s}", .{mn}) catch
            return error.MnemonicTooLong;
        const description_len = desc.len;

        var mnemonic_trig_buf: [128]u8 = undefined;
        const mn_trig = std.fmt.bufPrint(&mnemonic_trig_buf, "iTrendTrigger({d}{s})", .{ length, triple }) catch
            return error.MnemonicTooLong;
        const mnemonic_trig_len = mn_trig.len;

        var description_trig_buf: [192]u8 = undefined;
        const desc_trig = std.fmt.bufPrint(&description_trig_buf, "Instantaneous Trend Line trigger {s}", .{mn_trig}) catch
            return error.MnemonicTooLong;
        const description_trig_len = desc_trig.len;

        return .{
            .coeff1 = c1,
            .coeff2 = c2,
            .coeff3 = c3,
            .coeff4 = c4,
            .coeff5 = c5,
            .count = 0,
            .previous_sample1 = 0.0,
            .previous_sample2 = 0.0,
            .previous_trend_line1 = 0.0,
            .previous_trend_line2 = 0.0,
            .trend_line = math.nan(f64),
            .trigger_line = math.nan(f64),
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

    pub fn fixSlices(self: *InstantaneousTrendLine) void {
        // No slice fields to fix (mnemonics stored as owned buffers + lengths).
        _ = self;
    }

    /// Update the iTrend given the next sample. Returns the trend line value.
    pub fn update(self: *InstantaneousTrendLine, sample: f64) f64 {
        if (math.isNan(sample)) {
            return math.nan(f64);
        }

        if (self.primed) {
            self.trend_line = self.coeff1 * sample + self.coeff2 * self.previous_sample1 +
                self.coeff3 * self.previous_sample2 +
                self.coeff4 * self.previous_trend_line1 + self.coeff5 * self.previous_trend_line2;
            self.trigger_line = 2.0 * self.trend_line - self.previous_trend_line2;

            self.previous_sample2 = self.previous_sample1;
            self.previous_sample1 = sample;
            self.previous_trend_line2 = self.previous_trend_line1;
            self.previous_trend_line1 = self.trend_line;

            return self.trend_line;
        }

        self.count += 1;

        switch (self.count) {
            1 => {
                self.previous_sample2 = sample;
                return math.nan(f64);
            },
            2 => {
                self.previous_sample1 = sample;
                return math.nan(f64);
            },
            3 => {
                self.previous_trend_line2 = (sample + 2.0 * self.previous_sample1 + self.previous_sample2) / 4.0;
                self.previous_sample2 = self.previous_sample1;
                self.previous_sample1 = sample;
                return math.nan(f64);
            },
            4 => {
                self.previous_trend_line1 = (sample + 2.0 * self.previous_sample1 + self.previous_sample2) / 4.0;
                self.previous_sample2 = self.previous_sample1;
                self.previous_sample1 = sample;
                return math.nan(f64);
            },
            5 => {
                self.trend_line = self.coeff1 * sample + self.coeff2 * self.previous_sample1 +
                    self.coeff3 * self.previous_sample2 +
                    self.coeff4 * self.previous_trend_line1 + self.coeff5 * self.previous_trend_line2;
                self.trigger_line = 2.0 * self.trend_line - self.previous_trend_line2;

                self.previous_sample2 = self.previous_sample1;
                self.previous_sample1 = sample;
                self.previous_trend_line2 = self.previous_trend_line1;
                self.previous_trend_line1 = self.trend_line;
                self.primed = true;

                return self.trend_line;
            },
            else => return math.nan(f64),
        }
    }

    pub fn isPrimed(self: *const InstantaneousTrendLine) bool {
        return self.primed;
    }

    pub fn getMetadata(self: *const InstantaneousTrendLine, out: *Metadata) void {
        build_metadata_mod.buildMetadata(
            out,
            .instantaneous_trend_line,
            self.mnemonic_buf[0..self.mnemonic_len],
            self.description_buf[0..self.description_len],
            &[_]build_metadata_mod.OutputText{
                .{ .mnemonic = self.mnemonic_buf[0..self.mnemonic_len], .description = self.description_buf[0..self.description_len] },
                .{ .mnemonic = self.mnemonic_trig_buf[0..self.mnemonic_trig_len], .description = self.description_trig_buf[0..self.description_trig_len] },
            },
        );
    }

    pub fn updateScalar(self: *InstantaneousTrendLine, sample: *const Scalar) OutputArray {
        return self.updateEntity(sample.time, sample.value);
    }

    pub fn updateBar(self: *InstantaneousTrendLine, sample: *const Bar) OutputArray {
        return self.updateEntity(sample.time, self.bar_func(sample.*));
    }

    pub fn updateQuote(self: *InstantaneousTrendLine, sample: *const Quote) OutputArray {
        return self.updateEntity(sample.time, self.quote_func(sample.*));
    }

    pub fn updateTrade(self: *InstantaneousTrendLine, sample: *const Trade) OutputArray {
        return self.updateEntity(sample.time, self.trade_func(sample.*));
    }

    fn updateEntity(self: *InstantaneousTrendLine, time: i64, sample: f64) OutputArray {
        const v = self.update(sample);
        var trig = self.trigger_line;
        if (math.isNan(v)) {
            trig = math.nan(f64);
        }

        var out = OutputArray{};
        out.append(.{ .scalar = .{ .time = time, .value = v } });
        out.append(.{ .scalar = .{ .time = time, .value = trig } });
        return out;
    }

    /// Returns an Indicator interface backed by this instance.
    pub fn indicator(self: *InstantaneousTrendLine) indicator_mod.Indicator {
        return .{
            .ptr = @ptrCast(self),
            .vtable = &vtable,
        };
    }

    const vtable = indicator_mod.Indicator.GenVTable(InstantaneousTrendLine);
};

// --- Tests ---
const testing = std.testing;
const testdata = @import("testdata.zig");


fn almostEqual(a: f64, b: f64, tolerance: f64) bool {
    return @abs(a - b) <= tolerance;
}

// 252-entry input data from test_iTrend.xls.
// Expected trend line values from test_iTrend.xls, 252 entries.
// Expected trigger line values from test_iTrend.xls, 252 entries.
test "ITL update trend line" {
    const tolerance = 1e-8;
    const l_primed = 4;

    var itl = try InstantaneousTrendLine.initSmoothingFactor(.{ .smoothing_factor = 0.07 });
    itl.fixSlices();

    for (0..l_primed) |i| {
        try testing.expect(math.isNan(itl.update(testdata.test_input[i])));
    }

    for (l_primed..testdata.test_input.len) |i| {
        const act = itl.update(testdata.test_input[i]);
        try testing.expect(almostEqual(act, testdata.test_expected_trend[i], tolerance));
    }

    // NaN passthrough.
    try testing.expect(math.isNan(itl.update(math.nan(f64))));
}

test "ITL update trigger line" {
    const tolerance = 1e-8;
    const l_primed = 4;

    var itl = try InstantaneousTrendLine.initSmoothingFactor(.{ .smoothing_factor = 0.07 });
    itl.fixSlices();

    for (0..l_primed) |_i| {
        _ = itl.update(testdata.test_input[_i]);
    }

    for (l_primed..testdata.test_input.len) |i| {
        _ = itl.update(testdata.test_input[i]);
        try testing.expect(almostEqual(itl.trigger_line, testdata.test_expected_trigger[i], tolerance));
    }
}

test "ITL isPrimed" {
    var itl = try InstantaneousTrendLine.initSmoothingFactor(.{ .smoothing_factor = 0.07 });
    itl.fixSlices();

    try testing.expect(!itl.isPrimed());

    // First 4 updates: not primed.
    for (0..4) |i| {
        _ = itl.update(testdata.test_input[i]);
        try testing.expect(!itl.isPrimed());
    }

    // 5th update: primed.
    _ = itl.update(testdata.test_input[4]);
    try testing.expect(itl.isPrimed());
}

test "ITL metadata" {
    var itl = try InstantaneousTrendLine.initSmoothingFactor(.{ .smoothing_factor = 0.07 });
    itl.fixSlices();
    var meta: Metadata = undefined;
    itl.getMetadata(&meta);

    try testing.expectEqual(Identifier.instantaneous_trend_line, meta.identifier);
    try testing.expectEqualStrings("iTrend(28, hl/2)", meta.mnemonic);
    try testing.expectEqual(@as(usize, 2), meta.outputs_len);
}

test "ITL constructor length" {
    // Valid length.
    _ = try InstantaneousTrendLine.initLength(.{ .length = 28 });
    _ = try InstantaneousTrendLine.initLength(.{ .length = 1 });

    // Invalid length.
    try testing.expectError(error.InvalidLength, InstantaneousTrendLine.initLength(.{ .length = 0 }));
    try testing.expectError(error.InvalidLength, InstantaneousTrendLine.initLength(.{ .length = -8 }));
}

test "ITL constructor smoothing factor" {
    // Valid.
    _ = try InstantaneousTrendLine.initSmoothingFactor(.{ .smoothing_factor = 0.07 });
    _ = try InstantaneousTrendLine.initSmoothingFactor(.{ .smoothing_factor = 0.0 });
    _ = try InstantaneousTrendLine.initSmoothingFactor(.{ .smoothing_factor = 1.0 });

    // Invalid.
    try testing.expectError(error.InvalidSmoothingFactor, InstantaneousTrendLine.initSmoothingFactor(.{ .smoothing_factor = -0.0001 }));
    try testing.expectError(error.InvalidSmoothingFactor, InstantaneousTrendLine.initSmoothingFactor(.{ .smoothing_factor = 1.0001 }));
}

test "ITL updateScalar" {
    const tolerance = 1e-8;
    var itl = try InstantaneousTrendLine.initSmoothingFactor(.{ .smoothing_factor = 0.07 });
    itl.fixSlices();

    for (0..testdata.test_input.len) |i| {
        const s = Scalar{ .time = @intCast(i), .value = testdata.test_input[i] };
        const out = itl.updateScalar(&s);
        const outputs = out.slice();
        try testing.expectEqual(@as(usize, 2), outputs.len);

        if (i < 4) {
            try testing.expect(math.isNan(outputs[0].scalar.value));
            try testing.expect(math.isNan(outputs[1].scalar.value));
        } else {
            try testing.expect(almostEqual(outputs[0].scalar.value, testdata.test_expected_trend[i], tolerance));
            try testing.expect(almostEqual(outputs[1].scalar.value, testdata.test_expected_trigger[i], tolerance));
        }
    }
}
