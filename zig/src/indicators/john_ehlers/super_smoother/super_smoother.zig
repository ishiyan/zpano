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

/// Enumerates the outputs of the super smoother indicator.
pub const SuperSmootherOutput = enum(u8) {
    value = 1,
};

/// Parameters to create an instance of the Super Smoother.
pub const SuperSmootherParams = struct {
    /// Shortest cycle period in bars. Must be >= 2. Default is 10.
    shortest_cycle_period: i32 = 10,
    bar_component: ?bar_component.BarComponent = null,
    quote_component: ?quote_component.QuoteComponent = null,
    trade_component: ?trade_component.TradeComponent = null,
};

/// Ehler's two-pole Super Smoother (SS).
///
///   β = √2·π / λ
///   α = exp(-β)
///   γ₂ = 2α·cos(β)
///   γ₃ = -α²
///   γ₁ = (1 - γ₂ - γ₃) / 2
///
///   SSᵢ = γ₁·(xᵢ + xᵢ₋₁) + γ₂·SSᵢ₋₁ + γ₃·SSᵢ₋₂
///
/// The indicator is not primed during the first 2 updates.
pub const SuperSmoother = struct {
    line: LineIndicator,
    coeff1: f64,
    coeff2: f64,
    coeff3: f64,
    count: usize,
    sample_previous: f64,
    filter_previous: f64,
    filter_previous2: f64,
    value: f64,
    primed: bool,
    mnemonic_buf: [96]u8,
    mnemonic_len: usize,
    description_buf: [160]u8,
    description_len: usize,

    pub fn init(params: SuperSmootherParams) !SuperSmoother {
        const period = params.shortest_cycle_period;
        if (period < 2) {
            return error.InvalidPeriod;
        }

        // Calculate coefficients.
        const period_f: f64 = @floatFromInt(period);
        const beta = math.sqrt2 * math.pi / period_f;
        const alpha = @exp(-beta);
        const gamma2 = 2.0 * alpha * @cos(beta);
        const gamma3 = -alpha * alpha;
        const gamma1 = (1.0 - gamma2 - gamma3) / 2.0;

        // Build mnemonic.
        var triple_buf: [64]u8 = undefined;
        const triple = component_triple_mnemonic_mod.componentTripleMnemonic(
            &triple_buf,
            params.bar_component orelse bar_component.BarComponent.median,
            params.quote_component orelse quote_component.default_quote_component,
            params.trade_component orelse trade_component.default_trade_component,
        );

        var mnemonic_buf: [96]u8 = undefined;
        const mnemonic_slice = std.fmt.bufPrint(&mnemonic_buf, "ss({d}{s})", .{ period, triple }) catch
            return error.MnemonicTooLong;
        const mnemonic_len = mnemonic_slice.len;

        var description_buf: [160]u8 = undefined;
        const desc_slice = std.fmt.bufPrint(&description_buf, "Super Smoother {s}", .{mnemonic_slice}) catch
            return error.MnemonicTooLong;
        const description_len = desc_slice.len;

        return .{
            .line = LineIndicator.new(
                mnemonic_buf[0..mnemonic_len],
                description_buf[0..description_len],
                params.bar_component orelse bar_component.BarComponent.median,
                params.quote_component orelse quote_component.default_quote_component,
                params.trade_component orelse trade_component.default_trade_component,
            ),
            .coeff1 = gamma1,
            .coeff2 = gamma2,
            .coeff3 = gamma3,
            .count = 0,
            .sample_previous = 0.0,
            .filter_previous = 0.0,
            .filter_previous2 = 0.0,
            .value = math.nan(f64),
            .primed = false,
            .mnemonic_buf = mnemonic_buf,
            .mnemonic_len = mnemonic_len,
            .description_buf = description_buf,
            .description_len = description_len,
        };
    }

    pub fn fixSlices(self: *SuperSmoother) void {
        self.line.mnemonic = self.mnemonic_buf[0..self.mnemonic_len];
        self.line.description = self.description_buf[0..self.description_len];
    }

    /// Update the super smoother given the next sample.
    pub fn update(self: *SuperSmoother, sample: f64) f64 {
        if (math.isNan(sample)) {
            return sample;
        }

        if (self.primed) {
            const filter = self.coeff1 * (sample + self.sample_previous) +
                self.coeff2 * self.filter_previous + self.coeff3 * self.filter_previous2;
            self.value = filter;
            self.sample_previous = sample;
            self.filter_previous2 = self.filter_previous;
            self.filter_previous = filter;
            return self.value;
        }

        self.count += 1;

        if (self.count == 1) {
            self.sample_previous = sample;
            self.filter_previous = sample;
            self.filter_previous2 = sample;
        }

        const filter = self.coeff1 * (sample + self.sample_previous) +
            self.coeff2 * self.filter_previous + self.coeff3 * self.filter_previous2;

        if (self.count == 3) {
            self.primed = true;
            self.value = filter;
        }

        self.sample_previous = sample;
        self.filter_previous2 = self.filter_previous;
        self.filter_previous = filter;

        return self.value;
    }

    pub fn isPrimed(self: *const SuperSmoother) bool {
        return self.primed;
    }

    pub fn getMetadata(self: *const SuperSmoother, out: *Metadata) void {
        build_metadata_mod.buildMetadata(
            out,
            .super_smoother,
            self.line.mnemonic,
            self.line.description,
            &[_]build_metadata_mod.OutputText{
                .{ .mnemonic = self.line.mnemonic, .description = self.line.description },
            },
        );
    }

    pub fn updateScalar(self: *SuperSmoother, sample: *const Scalar) OutputArray {
        const value = self.update(sample.value);
        return LineIndicator.wrapScalar(sample.time, value);
    }

    pub fn updateBar(self: *SuperSmoother, sample: *const Bar) OutputArray {
        const value = self.update(self.line.extractBar(sample));
        return LineIndicator.wrapScalar(sample.time, value);
    }

    pub fn updateQuote(self: *SuperSmoother, sample: *const Quote) OutputArray {
        const value = self.update(self.line.extractQuote(sample));
        return LineIndicator.wrapScalar(sample.time, value);
    }

    pub fn updateTrade(self: *SuperSmoother, sample: *const Trade) OutputArray {
        const value = self.update(self.line.extractTrade(sample));
        return LineIndicator.wrapScalar(sample.time, value);
    }

    /// Returns an Indicator interface backed by this instance.
    pub fn indicator(self: *SuperSmoother) indicator_mod.Indicator {
        return .{
            .ptr = @ptrCast(self),
            .vtable = &vtable,
        };
    }

    const vtable = indicator_mod.Indicator.GenVTable(SuperSmoother);
};

// --- Tests ---
const testing = std.testing;
const testdata = @import("testdata.zig");


fn almostEqual(a: f64, b: f64, tolerance: f64) bool {
    return @abs(a - b) <= tolerance;
}

test "SuperSmoother update" {
    const skip_rows = 60;
    const tolerance = 0.5;

    var ss = try SuperSmoother.init(.{});
    ss.fixSlices();

    for (0..testdata.test_input.len) |i| {
        const act = ss.update(testdata.test_input[i]);

        if (i < 2) {
            try testing.expect(math.isNan(act));
            continue;
        }

        if (i < skip_rows) continue;

        try testing.expect(almostEqual(act, testdata.test_expected[i], tolerance));
    }

    // NaN passthrough.
    try testing.expect(math.isNan(ss.update(math.nan(f64))));
}

test "SuperSmoother isPrimed" {
    var ss = try SuperSmoother.init(.{});
    ss.fixSlices();

    try testing.expect(!ss.isPrimed());

    _ = ss.update(testdata.test_input[0]);
    try testing.expect(!ss.isPrimed());

    _ = ss.update(testdata.test_input[1]);
    try testing.expect(!ss.isPrimed());

    _ = ss.update(testdata.test_input[2]);
    try testing.expect(ss.isPrimed());
}

test "SuperSmoother metadata" {
    var ss = try SuperSmoother.init(.{});
    ss.fixSlices();
    var meta: Metadata = undefined;
    ss.getMetadata(&meta);

    try testing.expectEqual(Identifier.super_smoother, meta.identifier);
    try testing.expectEqualStrings("ss(10, hl/2)", meta.mnemonic);
    try testing.expectEqual(@as(usize, 1), meta.outputs_len);
}

test "SuperSmoother constructor validation" {
    // Valid default.
    _ = try SuperSmoother.init(.{});

    // Period < 2.
    try testing.expectError(error.InvalidPeriod, SuperSmoother.init(.{ .shortest_cycle_period = 1 }));
    try testing.expectError(error.InvalidPeriod, SuperSmoother.init(.{ .shortest_cycle_period = 0 }));
    try testing.expectError(error.InvalidPeriod, SuperSmoother.init(.{ .shortest_cycle_period = -1 }));
}

test "SuperSmoother updateBar" {
    var ss = try SuperSmoother.init(.{});
    ss.fixSlices();

    const bar1 = Bar{ .time = 1000, .open = 91, .high = 100, .low = 100, .close = 91.5, .volume = 1000 };
    const out1 = ss.updateBar(&bar1);
    try testing.expect(math.isNan(out1.slice()[0].scalar.value));

    // Prime.
    const bar2 = Bar{ .time = 2000, .open = 92, .high = 100, .low = 100, .close = 94.815, .volume = 1000 };
    _ = ss.updateBar(&bar2);
    const bar3 = Bar{ .time = 3000, .open = 93, .high = 100, .low = 100, .close = 95, .volume = 1000 };
    const out3 = ss.updateBar(&bar3);
    try testing.expect(!math.isNan(out3.slice()[0].scalar.value));
}

test "SuperSmoother custom bar component mnemonic" {
    var ss = try SuperSmoother.init(.{ .bar_component = bar_component.BarComponent.open });
    ss.fixSlices();
    var meta: Metadata = undefined;
    ss.getMetadata(&meta);
    try testing.expectEqualStrings("ss(10, o)", meta.mnemonic);
}
