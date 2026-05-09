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

/// Enumerates the outputs of the roofing filter indicator.
pub const RoofingFilterOutput = enum(u8) {
    value = 1,
};

/// Parameters to create an instance of the Roofing Filter.
pub const RoofingFilterParams = struct {
    /// Shortest cycle period in bars. Must be >= 2. Default is 10.
    shortest_cycle_period: i32 = 10,
    /// Longest cycle period in bars. Must be > shortest. Default is 48.
    longest_cycle_period: i32 = 48,
    /// Use 2-pole high-pass filter instead of 1-pole.
    has_two_pole_highpass_filter: bool = false,
    /// Apply zero-mean filter after super smoother (1-pole only).
    has_zero_mean: bool = false,
    bar_component: ?bar_component.BarComponent = null,
    quote_component: ?quote_component.QuoteComponent = null,
    trade_component: ?trade_component.TradeComponent = null,
};

/// Ehler's Roofing Filter: high-pass filter + Super Smoother.
///
/// Three flavours:
///   - 1-pole high-pass filter (default)
///   - 1-pole high-pass filter with zero-mean
///   - 2-pole high-pass filter
pub const RoofingFilter = struct {
    line: LineIndicator,
    hp_coeff1: f64,
    hp_coeff2: f64,
    hp_coeff3: f64,
    ss_coeff1: f64,
    ss_coeff2: f64,
    ss_coeff3: f64,

    has_two_pole: bool,
    has_zero_mean: bool,

    count: usize,
    sample_previous: f64,
    sample_previous2: f64,
    hp_previous: f64,
    hp_previous2: f64,
    ss_previous: f64,
    ss_previous2: f64,
    zm_previous: f64,
    value: f64,
    primed: bool,

    mnemonic_buf: [128]u8,
    mnemonic_len: usize,
    description_buf: [192]u8,
    description_len: usize,

    pub fn init(params: RoofingFilterParams) !RoofingFilter {
        const shortest = params.shortest_cycle_period;
        if (shortest < 2) {
            return error.InvalidShortestPeriod;
        }

        const longest = params.longest_cycle_period;
        if (longest <= shortest) {
            return error.InvalidLongestPeriod;
        }

        // Calculate high-pass filter coefficients.
        var hp_coeff1: f64 = 0;
        var hp_coeff2: f64 = 0;
        var hp_coeff3: f64 = 0;

        const longest_f: f64 = @floatFromInt(longest);

        if (params.has_two_pole_highpass_filter) {
            // 2-pole: angle = π√2/Λ
            const angle = math.sqrt2 / 2.0 * 2.0 * math.pi / longest_f;
            const cos_angle = @cos(angle);
            const alpha = (@sin(angle) + cos_angle - 1.0) / cos_angle;
            const beta = 1.0 - alpha / 2.0;
            hp_coeff1 = beta * beta;
            const beta2 = 1.0 - alpha;
            hp_coeff2 = 2.0 * beta2;
            hp_coeff3 = beta2 * beta2;
        } else {
            // 1-pole: angle = 2π/Λ
            const angle = 2.0 * math.pi / longest_f;
            const cos_angle = @cos(angle);
            const alpha = (@sin(angle) + cos_angle - 1.0) / cos_angle;
            hp_coeff1 = 1.0 - alpha / 2.0;
            hp_coeff2 = 1.0 - alpha;
        }

        // Calculate super smoother coefficients.
        // Uses literal 1.414 (not math.sqrt2) to match C# reference.
        const shortest_f: f64 = @floatFromInt(shortest);
        const beta = 1.414 * math.pi / shortest_f;
        const alpha = @exp(-beta);
        const ss_coeff2 = 2.0 * alpha * @cos(beta);
        const ss_coeff3 = -alpha * alpha;
        const ss_coeff1 = (1.0 - ss_coeff2 - ss_coeff3) / 2.0;

        const effective_zero_mean = params.has_zero_mean and !params.has_two_pole_highpass_filter;

        // Build mnemonic.
        var triple_buf: [64]u8 = undefined;
        const triple = component_triple_mnemonic_mod.componentTripleMnemonic(
            &triple_buf,
            params.bar_component orelse bar_component.BarComponent.median,
            params.quote_component orelse quote_component.default_quote_component,
            params.trade_component orelse trade_component.default_trade_component,
        );

        const poles: u8 = if (params.has_two_pole_highpass_filter) 2 else 1;
        const zm_str: []const u8 = if (effective_zero_mean) "zm" else "";

        var mnemonic_buf: [128]u8 = undefined;
        const mnemonic_slice = std.fmt.bufPrint(&mnemonic_buf, "roof{d}hp{s}({d}, {d}{s})", .{
            poles, zm_str, shortest, longest, triple,
        }) catch return error.MnemonicTooLong;
        const mnemonic_len = mnemonic_slice.len;

        var description_buf: [192]u8 = undefined;
        const desc_slice = std.fmt.bufPrint(&description_buf, "Roofing Filter {s}", .{mnemonic_slice}) catch
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
            .hp_coeff1 = hp_coeff1,
            .hp_coeff2 = hp_coeff2,
            .hp_coeff3 = hp_coeff3,
            .ss_coeff1 = ss_coeff1,
            .ss_coeff2 = ss_coeff2,
            .ss_coeff3 = ss_coeff3,
            .has_two_pole = params.has_two_pole_highpass_filter,
            .has_zero_mean = effective_zero_mean,
            .count = 0,
            .sample_previous = 0.0,
            .sample_previous2 = 0.0,
            .hp_previous = 0.0,
            .hp_previous2 = 0.0,
            .ss_previous = 0.0,
            .ss_previous2 = 0.0,
            .zm_previous = 0.0,
            .value = math.nan(f64),
            .primed = false,
            .mnemonic_buf = mnemonic_buf,
            .mnemonic_len = mnemonic_len,
            .description_buf = description_buf,
            .description_len = description_len,
        };
    }

    pub fn fixSlices(self: *RoofingFilter) void {
        self.line.mnemonic = self.mnemonic_buf[0..self.mnemonic_len];
        self.line.description = self.description_buf[0..self.description_len];
    }

    /// Update the roofing filter given the next sample.
    pub fn update(self: *RoofingFilter, sample: f64) f64 {
        if (math.isNan(sample)) {
            return sample;
        }

        if (self.has_two_pole) {
            return self.update2Pole(sample);
        }

        return self.update1Pole(sample);
    }

    fn update1Pole(self: *RoofingFilter, sample: f64) f64 {
        var hp: f64 = 0;
        var ss: f64 = 0;
        var zm: f64 = 0;

        if (self.primed) {
            hp = self.hp_coeff1 * (sample - self.sample_previous) + self.hp_coeff2 * self.hp_previous;
            ss = self.ss_coeff1 * (hp + self.hp_previous) + self.ss_coeff2 * self.ss_previous + self.ss_coeff3 * self.ss_previous2;

            if (self.has_zero_mean) {
                zm = self.hp_coeff1 * (ss - self.ss_previous) + self.hp_coeff2 * self.zm_previous;
                self.value = zm;
            } else {
                self.value = ss;
            }
        } else {
            self.count += 1;

            if (self.count == 1) {
                hp = 0;
                ss = 0;
            } else {
                hp = self.hp_coeff1 * (sample - self.sample_previous) + self.hp_coeff2 * self.hp_previous;
                ss = self.ss_coeff1 * (hp + self.hp_previous) + self.ss_coeff2 * self.ss_previous + self.ss_coeff3 * self.ss_previous2;

                if (self.has_zero_mean) {
                    zm = self.hp_coeff1 * (ss - self.ss_previous) + self.hp_coeff2 * self.zm_previous;
                    if (self.count == 5) {
                        self.primed = true;
                        self.value = zm;
                    }
                } else if (self.count == 4) {
                    self.primed = true;
                    self.value = ss;
                }
            }
        }

        self.sample_previous = sample;
        self.hp_previous = hp;
        self.ss_previous2 = self.ss_previous;
        self.ss_previous = ss;

        if (self.has_zero_mean) {
            self.zm_previous = zm;
        }

        return self.value;
    }

    fn update2Pole(self: *RoofingFilter, sample: f64) f64 {
        var hp: f64 = 0;
        var ss: f64 = 0;

        if (self.primed) {
            hp = self.hp_coeff1 * (sample - 2.0 * self.sample_previous + self.sample_previous2) +
                self.hp_coeff2 * self.hp_previous - self.hp_coeff3 * self.hp_previous2;
            ss = self.ss_coeff1 * (hp + self.hp_previous) + self.ss_coeff2 * self.ss_previous + self.ss_coeff3 * self.ss_previous2;
            self.value = ss;
        } else {
            self.count += 1;

            if (self.count < 4) {
                hp = 0;
                ss = 0;
            } else {
                hp = self.hp_coeff1 * (sample - 2.0 * self.sample_previous + self.sample_previous2) +
                    self.hp_coeff2 * self.hp_previous - self.hp_coeff3 * self.hp_previous2;
                ss = self.ss_coeff1 * (hp + self.hp_previous) + self.ss_coeff2 * self.ss_previous + self.ss_coeff3 * self.ss_previous2;

                if (self.count == 5) {
                    self.primed = true;
                    self.value = ss;
                }
            }
        }

        self.sample_previous2 = self.sample_previous;
        self.sample_previous = sample;
        self.hp_previous2 = self.hp_previous;
        self.hp_previous = hp;
        self.ss_previous2 = self.ss_previous;
        self.ss_previous = ss;

        return self.value;
    }

    pub fn isPrimed(self: *const RoofingFilter) bool {
        return self.primed;
    }

    pub fn getMetadata(self: *const RoofingFilter, out: *Metadata) void {
        build_metadata_mod.buildMetadata(
            out,
            .roofing_filter,
            self.line.mnemonic,
            self.line.description,
            &[_]build_metadata_mod.OutputText{
                .{ .mnemonic = self.line.mnemonic, .description = self.line.description },
            },
        );
    }

    pub fn updateScalar(self: *RoofingFilter, sample: *const Scalar) OutputArray {
        const value = self.update(sample.value);
        return LineIndicator.wrapScalar(sample.time, value);
    }

    pub fn updateBar(self: *RoofingFilter, sample: *const Bar) OutputArray {
        const value = self.update(self.line.extractBar(sample));
        return LineIndicator.wrapScalar(sample.time, value);
    }

    pub fn updateQuote(self: *RoofingFilter, sample: *const Quote) OutputArray {
        const value = self.update(self.line.extractQuote(sample));
        return LineIndicator.wrapScalar(sample.time, value);
    }

    pub fn updateTrade(self: *RoofingFilter, sample: *const Trade) OutputArray {
        const value = self.update(self.line.extractTrade(sample));
        return LineIndicator.wrapScalar(sample.time, value);
    }

    /// Returns an Indicator interface backed by this instance.
    pub fn indicator(self: *RoofingFilter) indicator_mod.Indicator {
        return .{
            .ptr = @ptrCast(self),
            .vtable = &vtable,
        };
    }

    const vtable = indicator_mod.Indicator.GenVTable(RoofingFilter);
};

// --- Tests ---
const testing = std.testing;
const testdata = @import("testdata.zig");


fn almostEqual(a: f64, b: f64, tolerance: f64) bool {
    return @abs(a - b) <= tolerance;
}

// Expected: 1-pole HP, no zero-mean (shortest=10, longest=48).
// Expected: 1-pole HP, zero-mean (shortest=10, longest=48).
// Expected: 2-pole HP (shortest=40, longest=80).
test "RoofingFilter update 1-pole" {
    const skip_rows = 30;
    const tolerance = 0.5;

    var rf = try RoofingFilter.init(.{});
    rf.fixSlices();

    for (0..testdata.test_input.len) |i| {
        const act = rf.update(testdata.test_input[i]);

        if (i < 3) {
            try testing.expect(math.isNan(act));
            continue;
        }

        if (i < skip_rows) continue;

        try testing.expect(almostEqual(act, testdata.test_expected_71[i], tolerance));
    }

    // NaN passthrough.
    try testing.expect(math.isNan(rf.update(math.nan(f64))));
}

test "RoofingFilter update 1-pole zero-mean" {
    const skip_rows = 30;
    const tolerance = 0.5;

    var rf = try RoofingFilter.init(.{ .has_zero_mean = true });
    rf.fixSlices();

    for (0..testdata.test_input.len) |i| {
        const act = rf.update(testdata.test_input[i]);

        if (i < 4) {
            try testing.expect(math.isNan(act));
            continue;
        }

        if (i < skip_rows) continue;

        try testing.expect(almostEqual(act, testdata.test_expected_72[i], tolerance));
    }
}

test "RoofingFilter update 2-pole" {
    const skip_rows = 30;
    const tolerance = 0.5;

    var rf = try RoofingFilter.init(.{
        .shortest_cycle_period = 40,
        .longest_cycle_period = 80,
        .has_two_pole_highpass_filter = true,
    });
    rf.fixSlices();

    for (0..testdata.test_input.len) |i| {
        const act = rf.update(testdata.test_input[i]);

        if (i < 4) {
            try testing.expect(math.isNan(act));
            continue;
        }

        if (i < skip_rows) continue;

        try testing.expect(almostEqual(act, testdata.test_expected_73[i], tolerance));
    }
}

test "RoofingFilter isPrimed 1-pole" {
    var rf = try RoofingFilter.init(.{});
    rf.fixSlices();

    try testing.expect(!rf.isPrimed());

    for (0..3) |i| {
        _ = rf.update(testdata.test_input[i]);
        try testing.expect(!rf.isPrimed());
    }

    _ = rf.update(testdata.test_input[3]);
    try testing.expect(rf.isPrimed());
}

test "RoofingFilter isPrimed 1-pole zero-mean" {
    var rf = try RoofingFilter.init(.{ .has_zero_mean = true });
    rf.fixSlices();

    for (0..4) |i| {
        _ = rf.update(testdata.test_input[i]);
        try testing.expect(!rf.isPrimed());
    }

    _ = rf.update(testdata.test_input[4]);
    try testing.expect(rf.isPrimed());
}

test "RoofingFilter isPrimed 2-pole" {
    var rf = try RoofingFilter.init(.{
        .shortest_cycle_period = 40,
        .longest_cycle_period = 80,
        .has_two_pole_highpass_filter = true,
    });
    rf.fixSlices();

    for (0..4) |i| {
        _ = rf.update(testdata.test_input[i]);
        try testing.expect(!rf.isPrimed());
    }

    _ = rf.update(testdata.test_input[4]);
    try testing.expect(rf.isPrimed());
}

test "RoofingFilter metadata" {
    var rf = try RoofingFilter.init(.{});
    rf.fixSlices();
    var meta: Metadata = undefined;
    rf.getMetadata(&meta);

    try testing.expectEqual(Identifier.roofing_filter, meta.identifier);
    try testing.expectEqualStrings("roof1hp(10, 48, hl/2)", meta.mnemonic);
    try testing.expectEqual(@as(usize, 1), meta.outputs_len);
}

test "RoofingFilter constructor validation" {
    // Valid default.
    _ = try RoofingFilter.init(.{});

    // Shortest < 2.
    try testing.expectError(error.InvalidShortestPeriod, RoofingFilter.init(.{ .shortest_cycle_period = 1 }));

    // Longest <= shortest.
    try testing.expectError(error.InvalidLongestPeriod, RoofingFilter.init(.{ .longest_cycle_period = 10 }));
    try testing.expectError(error.InvalidLongestPeriod, RoofingFilter.init(.{ .longest_cycle_period = 5 }));
}

test "RoofingFilter mnemonic variants" {
    // 1-pole default.
    {
        var rf = try RoofingFilter.init(.{});
        rf.fixSlices();
        var meta: Metadata = undefined;
        rf.getMetadata(&meta);
        try testing.expectEqualStrings("roof1hp(10, 48, hl/2)", meta.mnemonic);
    }

    // 2-pole.
    {
        var rf = try RoofingFilter.init(.{ .has_two_pole_highpass_filter = true });
        rf.fixSlices();
        var meta: Metadata = undefined;
        rf.getMetadata(&meta);
        try testing.expectEqualStrings("roof2hp(10, 48, hl/2)", meta.mnemonic);
    }

    // 1-pole zero-mean.
    {
        var rf = try RoofingFilter.init(.{ .has_zero_mean = true });
        rf.fixSlices();
        var meta: Metadata = undefined;
        rf.getMetadata(&meta);
        try testing.expectEqualStrings("roof1hpzm(10, 48, hl/2)", meta.mnemonic);
    }

    // Custom bar component.
    {
        var rf = try RoofingFilter.init(.{ .bar_component = bar_component.BarComponent.open });
        rf.fixSlices();
        var meta: Metadata = undefined;
        rf.getMetadata(&meta);
        try testing.expectEqualStrings("roof1hp(10, 48, o)", meta.mnemonic);
    }
}

test "RoofingFilter updateBar" {
    var rf = try RoofingFilter.init(.{});
    rf.fixSlices();

    // Prime: need 4 updates for 1-pole.
    const inp: f64 = 100.0;
    for (0..4) |_| {
        _ = rf.update(inp);
    }

    const bar1 = Bar{ .time = 1000, .open = 91, .high = inp, .low = inp, .close = 91.5, .volume = 1000 };
    const out1 = rf.updateBar(&bar1);
    try testing.expect(!math.isNan(out1.slice()[0].scalar.value));
}
