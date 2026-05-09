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

/// Enumerates the outputs of the cyber cycle indicator.
pub const CyberCycleOutput = enum(u8) {
    value = 1,
    signal = 2,
};

/// Parameters to create a CyberCycle based on length.
pub const LengthParams = struct {
    /// Length (α = 2/(ℓ+1)). Must be >= 1. Default is 28.
    length: i32 = 28,
    /// Signal lag for the EMA signal line. Must be >= 1. Default is 9.
    signal_lag: i32 = 9,
    bar_component: ?bar_component.BarComponent = null,
    quote_component: ?quote_component.QuoteComponent = null,
    trade_component: ?trade_component.TradeComponent = null,
};

/// Parameters to create a CyberCycle based on smoothing factor.
pub const SmoothingFactorParams = struct {
    /// Smoothing factor α in [0, 1]. Default is 0.07.
    smoothing_factor: f64 = 0.07,
    /// Signal lag for the EMA signal line. Must be >= 1. Default is 9.
    signal_lag: i32 = 9,
    bar_component: ?bar_component.BarComponent = null,
    quote_component: ?quote_component.QuoteComponent = null,
    trade_component: ?trade_component.TradeComponent = null,
};

/// Ehler's Cyber Cycle (CC).
///
///   H(z) = ((1-α/2)²(1 - 2z⁻¹ + z⁻²)) / (1 - 2(1-α)z⁻¹ + (1-α)²z⁻²)
///
/// A complementary high-pass filter found by subtracting the Instantaneous
/// Trend Line low-pass filter from unity. Has zero lag and retains relative
/// cycle amplitude.
///
/// Two outputs: cycle value and signal line (EMA of cycle).
/// Primed after 8 samples.
pub const CyberCycle = struct {
    coeff1: f64,
    coeff2: f64,
    coeff3: f64,
    coeff4: f64,
    coeff5: f64,
    count: usize,
    previous_sample1: f64,
    previous_sample2: f64,
    previous_sample3: f64,
    smoothed: f64,
    previous_smoothed1: f64,
    previous_smoothed2: f64,
    value: f64,
    previous_value1: f64,
    previous_value2: f64,
    signal: f64,
    primed: bool,
    bar_func: bar_component.BarFunc,
    quote_func: quote_component.QuoteFunc,
    trade_func: trade_component.TradeFunc,
    mnemonic_buf: [128]u8,
    mnemonic_len: usize,
    description_buf: [192]u8,
    description_len: usize,
    mnemonic_signal_buf: [128]u8,
    mnemonic_signal_len: usize,
    description_signal_buf: [192]u8,
    description_signal_len: usize,

    pub fn initLength(params: LengthParams) !CyberCycle {
        if (params.length < 1) {
            return error.InvalidLength;
        }

        if (params.signal_lag < 1) {
            return error.InvalidSignalLag;
        }

        const alpha: f64 = 2.0 / @as(f64, @floatFromInt(1 + params.length));
        return initCommon(params.length, alpha, params.signal_lag, params.bar_component, params.quote_component, params.trade_component);
    }

    pub fn initSmoothingFactor(params: SmoothingFactorParams) !CyberCycle {
        const alpha = params.smoothing_factor;
        if (alpha < 0.0 or alpha > 1.0) {
            return error.InvalidSmoothingFactor;
        }

        if (params.signal_lag < 1) {
            return error.InvalidSignalLag;
        }

        const epsilon: f64 = 0.00000001;
        const length: i32 = if (alpha < epsilon)
            std.math.maxInt(i32)
        else
            @as(i32, @intFromFloat(@round(2.0 / alpha))) - 1;

        return initCommon(length, alpha, params.signal_lag, params.bar_component, params.quote_component, params.trade_component);
    }

    fn initCommon(
        length: i32,
        alpha: f64,
        signal_lag: i32,
        bc_opt: ?bar_component.BarComponent,
        qc_opt: ?quote_component.QuoteComponent,
        tc_opt: ?trade_component.TradeComponent,
    ) !CyberCycle {
        const bc = bc_opt orelse bar_component.BarComponent.median;
        const qc = qc_opt orelse quote_component.default_quote_component;
        const tc = tc_opt orelse trade_component.default_trade_component;

        // Calculate coefficients.
        // High-pass: c1 = (1 - α/2)², c2 = 2(1-α), c3 = -(1-α)²
        var x = 1.0 - alpha / 2.0;
        const c1 = x * x;

        x = 1.0 - alpha;
        const c2 = 2.0 * x;
        const c3 = -(x * x);

        // Signal EMA: c4 = 1/(1+signalLag), c5 = 1 - c4
        const sl: f64 = @floatFromInt(signal_lag);
        const c4 = 1.0 / (1.0 + sl);
        const c5 = 1.0 - c4;

        // Build mnemonics.
        var triple_buf: [64]u8 = undefined;
        const triple = component_triple_mnemonic_mod.componentTripleMnemonic(
            &triple_buf,
            bc_opt orelse bar_component.BarComponent.median,
            qc_opt orelse quote_component.default_quote_component,
            tc_opt orelse trade_component.default_trade_component,
        );

        var mnemonic_buf: [128]u8 = undefined;
        const mn = std.fmt.bufPrint(&mnemonic_buf, "cc({d}{s})", .{ length, triple }) catch
            return error.MnemonicTooLong;
        const mnemonic_len = mn.len;

        var description_buf: [192]u8 = undefined;
        const desc = std.fmt.bufPrint(&description_buf, "Cyber Cycle {s}", .{mn}) catch
            return error.MnemonicTooLong;
        const description_len = desc.len;

        var mnemonic_signal_buf: [128]u8 = undefined;
        const mn_sig = std.fmt.bufPrint(&mnemonic_signal_buf, "ccSignal({d}{s})", .{ length, triple }) catch
            return error.MnemonicTooLong;
        const mnemonic_signal_len = mn_sig.len;

        var description_signal_buf: [192]u8 = undefined;
        const desc_sig = std.fmt.bufPrint(&description_signal_buf, "Cyber Cycle signal {s}", .{mn_sig}) catch
            return error.MnemonicTooLong;
        const description_signal_len = desc_sig.len;

        return .{
            .coeff1 = c1,
            .coeff2 = c2,
            .coeff3 = c3,
            .coeff4 = c4,
            .coeff5 = c5,
            .count = 0,
            .previous_sample1 = 0.0,
            .previous_sample2 = 0.0,
            .previous_sample3 = 0.0,
            .smoothed = 0.0,
            .previous_smoothed1 = 0.0,
            .previous_smoothed2 = 0.0,
            .value = math.nan(f64),
            .previous_value1 = 0.0,
            .previous_value2 = 0.0,
            .signal = 0.0,
            .primed = false,
            .bar_func = bar_component.componentValue(bc),
            .quote_func = quote_component.componentValue(qc),
            .trade_func = trade_component.componentValue(tc),
            .mnemonic_buf = mnemonic_buf,
            .mnemonic_len = mnemonic_len,
            .description_buf = description_buf,
            .description_len = description_len,
            .mnemonic_signal_buf = mnemonic_signal_buf,
            .mnemonic_signal_len = mnemonic_signal_len,
            .description_signal_buf = description_signal_buf,
            .description_signal_len = description_signal_len,
        };
    }

    pub fn fixSlices(self: *CyberCycle) void {
        _ = self;
    }

    /// Update the cyber cycle given the next sample. Returns the cycle value.
    pub fn update(self: *CyberCycle, sample: f64) f64 {
        if (math.isNan(sample)) {
            return math.nan(f64);
        }

        if (self.primed) {
            self.previous_smoothed2 = self.previous_smoothed1;
            self.previous_smoothed1 = self.smoothed;
            self.smoothed = (sample + 2.0 * self.previous_sample1 + 2.0 * self.previous_sample2 + self.previous_sample3) / 6.0;

            self.previous_value2 = self.previous_value1;
            self.previous_value1 = self.value;
            self.value = self.coeff1 * (self.smoothed - 2.0 * self.previous_smoothed1 + self.previous_smoothed2) +
                self.coeff2 * self.previous_value1 + self.coeff3 * self.previous_value2;

            self.signal = self.coeff4 * self.value + self.coeff5 * self.signal;

            self.previous_sample3 = self.previous_sample2;
            self.previous_sample2 = self.previous_sample1;
            self.previous_sample1 = sample;

            return self.value;
        }

        self.count += 1;

        switch (self.count) {
            1 => {
                self.previous_sample3 = sample;
                return math.nan(f64);
            },
            2 => {
                self.previous_sample2 = sample;
                return math.nan(f64);
            },
            3 => {
                self.signal = self.coeff4 * (sample - 2.0 * self.previous_sample2 + self.previous_sample3) / 4.0;
                self.previous_sample1 = sample;
                return math.nan(f64);
            },
            4 => {
                self.previous_smoothed2 = (sample + 2.0 * self.previous_sample1 + 2.0 * self.previous_sample2 + self.previous_sample3) / 6.0;
                self.signal = self.coeff4 * (sample - 2.0 * self.previous_sample1 + self.previous_sample2) / 4.0 + self.coeff5 * self.signal;

                self.previous_sample3 = self.previous_sample2;
                self.previous_sample2 = self.previous_sample1;
                self.previous_sample1 = sample;
                return math.nan(f64);
            },
            5 => {
                self.previous_smoothed1 = (sample + 2.0 * self.previous_sample1 + 2.0 * self.previous_sample2 + self.previous_sample3) / 6.0;
                self.signal = self.coeff4 * (sample - 2.0 * self.previous_sample1 + self.previous_sample2) / 4.0 + self.coeff5 * self.signal;

                self.previous_sample3 = self.previous_sample2;
                self.previous_sample2 = self.previous_sample1;
                self.previous_sample1 = sample;
                return math.nan(f64);
            },
            6 => {
                self.smoothed = (sample + 2.0 * self.previous_sample1 + 2.0 * self.previous_sample2 + self.previous_sample3) / 6.0;
                self.previous_value2 = (sample - 2.0 * self.previous_sample1 + self.previous_sample2) / 4.0;
                self.signal = self.coeff4 * self.previous_value2 + self.coeff5 * self.signal;

                self.previous_sample3 = self.previous_sample2;
                self.previous_sample2 = self.previous_sample1;
                self.previous_sample1 = sample;
                return math.nan(f64);
            },
            7 => {
                self.previous_smoothed2 = self.previous_smoothed1;
                self.previous_smoothed1 = self.smoothed;
                self.smoothed = (sample + 2.0 * self.previous_sample1 + 2.0 * self.previous_sample2 + self.previous_sample3) / 6.0;
                self.previous_value1 = (sample - 2.0 * self.previous_sample1 + self.previous_sample2) / 4.0;
                self.signal = self.coeff4 * self.previous_value1 + self.coeff5 * self.signal;

                self.previous_sample3 = self.previous_sample2;
                self.previous_sample2 = self.previous_sample1;
                self.previous_sample1 = sample;
                return math.nan(f64);
            },
            8 => {
                self.previous_smoothed2 = self.previous_smoothed1;
                self.previous_smoothed1 = self.smoothed;
                self.smoothed = (sample + 2.0 * self.previous_sample1 + 2.0 * self.previous_sample2 + self.previous_sample3) / 6.0;

                self.value = self.coeff1 * (self.smoothed - 2.0 * self.previous_smoothed1 + self.previous_smoothed2) +
                    self.coeff2 * self.previous_value1 + self.coeff3 * self.previous_value2;

                self.signal = self.coeff4 * self.value + self.coeff5 * self.signal;

                self.previous_sample3 = self.previous_sample2;
                self.previous_sample2 = self.previous_sample1;
                self.previous_sample1 = sample;
                self.primed = true;

                return self.value;
            },
            else => return math.nan(f64),
        }
    }

    pub fn isPrimed(self: *const CyberCycle) bool {
        return self.primed;
    }

    pub fn getMetadata(self: *const CyberCycle, out: *Metadata) void {
        build_metadata_mod.buildMetadata(
            out,
            .cyber_cycle,
            self.mnemonic_buf[0..self.mnemonic_len],
            self.description_buf[0..self.description_len],
            &[_]build_metadata_mod.OutputText{
                .{ .mnemonic = self.mnemonic_buf[0..self.mnemonic_len], .description = self.description_buf[0..self.description_len] },
                .{ .mnemonic = self.mnemonic_signal_buf[0..self.mnemonic_signal_len], .description = self.description_signal_buf[0..self.description_signal_len] },
            },
        );
    }

    pub fn updateScalar(self: *CyberCycle, sample: *const Scalar) OutputArray {
        return self.updateEntity(sample.time, sample.value);
    }

    pub fn updateBar(self: *CyberCycle, sample: *const Bar) OutputArray {
        return self.updateEntity(sample.time, self.bar_func(sample.*));
    }

    pub fn updateQuote(self: *CyberCycle, sample: *const Quote) OutputArray {
        return self.updateEntity(sample.time, self.quote_func(sample.*));
    }

    pub fn updateTrade(self: *CyberCycle, sample: *const Trade) OutputArray {
        return self.updateEntity(sample.time, self.trade_func(sample.*));
    }

    fn updateEntity(self: *CyberCycle, time: i64, sample: f64) OutputArray {
        const v = self.update(sample);
        var sig = self.signal;
        if (math.isNan(v)) {
            sig = math.nan(f64);
        }

        var out = OutputArray{};
        out.append(.{ .scalar = .{ .time = time, .value = v } });
        out.append(.{ .scalar = .{ .time = time, .value = sig } });
        return out;
    }

    /// Returns an Indicator interface backed by this instance.
    pub fn indicator(self: *CyberCycle) indicator_mod.Indicator {
        return .{
            .ptr = @ptrCast(self),
            .vtable = &vtable,
        };
    }

    const vtable = indicator_mod.Indicator.GenVTable(CyberCycle);
};

// --- Tests ---
const testing = std.testing;
const testdata = @import("testdata.zig");


fn almostEqual(a: f64, b: f64, tolerance: f64) bool {
    return @abs(a - b) <= tolerance;
}

// 252-entry input data from test_iTrend.xls (same as ITL).
// Expected cycle (Value) values, 252 entries.
// Expected signal values, 252 entries.
test "CC update cycle value" {
    const tolerance = 1e-8;
    const l_primed = 7; // First 7 values are NaN (primed on sample 8).

    var cc = try CyberCycle.initSmoothingFactor(.{ .smoothing_factor = 0.07, .signal_lag = 9 });
    cc.fixSlices();

    for (0..l_primed) |i| {
        try testing.expect(math.isNan(cc.update(testdata.test_input[i])));
    }

    for (l_primed..testdata.test_input.len) |i| {
        const act = cc.update(testdata.test_input[i]);
        try testing.expect(almostEqual(act, testdata.test_expected_cycle[i], tolerance));
    }

    // NaN passthrough.
    try testing.expect(math.isNan(cc.update(math.nan(f64))));
}

test "CC update signal" {
    const tolerance = 1e-8;
    const l_primed = 7;

    var cc = try CyberCycle.initSmoothingFactor(.{ .smoothing_factor = 0.07, .signal_lag = 9 });
    cc.fixSlices();

    for (0..l_primed) |_i| {
        _ = cc.update(testdata.test_input[_i]);
    }

    for (l_primed..testdata.test_input.len) |i| {
        _ = cc.update(testdata.test_input[i]);
        try testing.expect(almostEqual(cc.signal, testdata.test_expected_signal[i], tolerance));
    }
}

test "CC isPrimed" {
    var cc = try CyberCycle.initSmoothingFactor(.{ .smoothing_factor = 0.07, .signal_lag = 9 });
    cc.fixSlices();

    try testing.expect(!cc.isPrimed());

    const l_primed = 7;
    for (0..l_primed) |i| {
        _ = cc.update(testdata.test_input[i]);
        try testing.expect(!cc.isPrimed());
    }

    // 8th update: primed.
    _ = cc.update(testdata.test_input[l_primed]);
    try testing.expect(cc.isPrimed());
}

test "CC metadata" {
    var cc = try CyberCycle.initSmoothingFactor(.{ .smoothing_factor = 0.07, .signal_lag = 9 });
    cc.fixSlices();
    var meta: Metadata = undefined;
    cc.getMetadata(&meta);

    try testing.expectEqual(Identifier.cyber_cycle, meta.identifier);
    try testing.expectEqualStrings("cc(28, hl/2)", meta.mnemonic);
    try testing.expectEqual(@as(usize, 2), meta.outputs_len);
}

test "CC constructor length" {
    _ = try CyberCycle.initLength(.{ .length = 28, .signal_lag = 14 });
    _ = try CyberCycle.initLength(.{ .length = 1, .signal_lag = 1 });

    try testing.expectError(error.InvalidLength, CyberCycle.initLength(.{ .length = 0, .signal_lag = 1 }));
    try testing.expectError(error.InvalidLength, CyberCycle.initLength(.{ .length = -8, .signal_lag = 1 }));
    try testing.expectError(error.InvalidSignalLag, CyberCycle.initLength(.{ .length = 1, .signal_lag = 0 }));
    try testing.expectError(error.InvalidSignalLag, CyberCycle.initLength(.{ .length = 1, .signal_lag = -8 }));
}

test "CC constructor smoothing factor" {
    _ = try CyberCycle.initSmoothingFactor(.{ .smoothing_factor = 0.07, .signal_lag = 9 });
    _ = try CyberCycle.initSmoothingFactor(.{ .smoothing_factor = 0.0, .signal_lag = 9 });
    _ = try CyberCycle.initSmoothingFactor(.{ .smoothing_factor = 1.0, .signal_lag = 9 });

    try testing.expectError(error.InvalidSmoothingFactor, CyberCycle.initSmoothingFactor(.{ .smoothing_factor = -0.0001, .signal_lag = 8 }));
    try testing.expectError(error.InvalidSmoothingFactor, CyberCycle.initSmoothingFactor(.{ .smoothing_factor = 1.0001, .signal_lag = 8 }));
    try testing.expectError(error.InvalidSignalLag, CyberCycle.initSmoothingFactor(.{ .smoothing_factor = 0.07, .signal_lag = 0 }));
}

test "CC updateScalar" {
    const tolerance = 1e-8;
    var cc = try CyberCycle.initSmoothingFactor(.{ .smoothing_factor = 0.07, .signal_lag = 9 });
    cc.fixSlices();

    for (0..testdata.test_input.len) |i| {
        const s = Scalar{ .time = @intCast(i), .value = testdata.test_input[i] };
        const out = cc.updateScalar(&s);
        const outputs = out.slice();
        try testing.expectEqual(@as(usize, 2), outputs.len);

        if (i < 7) {
            try testing.expect(math.isNan(outputs[0].scalar.value));
            try testing.expect(math.isNan(outputs[1].scalar.value));
        } else {
            try testing.expect(almostEqual(outputs[0].scalar.value, testdata.test_expected_cycle[i], tolerance));
            try testing.expect(almostEqual(outputs[1].scalar.value, testdata.test_expected_signal[i], tolerance));
        }
    }
}
