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
const tsi_mod = @import("../true_strength_index/true_strength_index.zig");
const smi_mod = @import("../stochastic_momentum_index/stochastic_momentum_index.zig");
const dti_mod = @import("../directional_trend_index/directional_trend_index.zig");
const tvi_mod = @import("../tick_volume_indicator/tick_volume_indicator.zig");
const mdi_mod = @import("../mean_deviation_index/mean_deviation_index.zig");
const cmi_mod = @import("../candlestick_momentum_index/candlestick_momentum_index.zig");
const csi_mod = @import("../candlestick_strength_index/candlestick_strength_index.zig");

const OutputArray = indicator_mod.OutputArray;
const Identifier = identifier_mod.Identifier;
const Metadata = metadata_mod.Metadata;

/// Enumerates the outputs of the Nonambiguous Trend Filter indicator.
pub const NonambiguousTrendFilterOutput = enum(u8) {
    /// The Nonambiguous Trend Filter value (the base oscillator value or 0).
    value = 1,
};

/// Specifies the base oscillator the Nonambiguous Trend Filter is applied to.
pub const NonambiguousTrendFilterBase = enum(u8) {
    /// True Strength Index (named instance TSI_Trade); defaults q=2, r=32, s=13, u=3.
    tsi = 0,
    /// Stochastic Momentum Index (named instance SMI_Trade); defaults q=32, r=64, s=7, u=1.
    smi = 1,
    /// Directional Trend Index (named instance DTI_Trade); defaults q=2, r=28, s=28, u=5.
    dti = 2,
    /// Tick Volume Indicator (named instance TVI_Trade); defaults r=32, s=32, u=5.
    tvi = 3,
    /// Mean Deviation Index (MDI_Trade); defaults r=20, s=5, u=3.
    mdi = 4,
    /// Candlestick Momentum Index (CMI_Trade); defaults r=20, s=5, u=3.
    cmi = 5,
    /// Candlestick Strength Index (CSI_Trade); defaults r=32, s=32, u=1.
    csi = 6,
};

/// Parameters to create a Nonambiguous Trend Filter indicator.
///
/// The filter itself is parameterless; q, r, s and u are the canonical symbols of
/// the base oscillator from William Blau's Momentum, Direction, and Divergence
/// (Wiley, 1995). A zero value selects the book default of the selected base.
pub const NonambiguousTrendFilterParams = struct {
    /// The base oscillator the filter is applied to. Default tsi.
    base: NonambiguousTrendFilterBase = .tsi,
    /// The momentum look-back period of the base (tsi, smi and dti only). Zero selects the base default.
    q: usize = 0,
    /// The period of the 1st EMA of the base smoothing cascade. Zero selects the base default.
    r: usize = 0,
    /// The period of the 2nd EMA of the base smoothing cascade. Zero selects the base default.
    s: usize = 0,
    /// The period of the 3rd EMA of the base smoothing cascade. Zero selects the base default.
    u: usize = 0,
    /// Bar component to extract (tsi and mdi only). `null` means use default (Close).
    bar_component: ?bar_component.BarComponent = null,
    /// Quote component to extract (tsi and mdi only). `null` means use default (Mid).
    quote_component: ?quote_component.QuoteComponent = null,
    /// Trade component to extract (tsi and mdi only). `null` means use default (Price).
    trade_component: ?trade_component.TradeComponent = null,
};

/// The book named-instance defaults of a base.
const BaseDefaults = struct {
    name: []const u8,
    q: usize,
    r: usize,
    s: usize,
    u: usize,
    uses_q: bool,
};

fn baseDefaults(base: NonambiguousTrendFilterBase) BaseDefaults {
    return switch (base) {
        .tsi => .{ .name = "tsi", .q = 2, .r = 32, .s = 13, .u = 3, .uses_q = true },
        .smi => .{ .name = "smi", .q = 32, .r = 64, .s = 7, .u = 1, .uses_q = true },
        .dti => .{ .name = "dti", .q = 2, .r = 28, .s = 28, .u = 5, .uses_q = true },
        .tvi => .{ .name = "tvi", .q = 0, .r = 32, .s = 32, .u = 5, .uses_q = false },
        .mdi => .{ .name = "mdi", .q = 0, .r = 20, .s = 5, .u = 3, .uses_q = false },
        .cmi => .{ .name = "cmi", .q = 0, .r = 20, .s = 5, .u = 3, .uses_q = false },
        .csi => .{ .name = "csi", .q = 0, .r = 32, .s = 32, .u = 1, .uses_q = false },
    };
}

/// The wrapped base indicator.
const BaseIndicator = union(NonambiguousTrendFilterBase) {
    tsi: tsi_mod.TrueStrengthIndex,
    smi: smi_mod.StochasticMomentumIndex,
    dti: dti_mod.DirectionalTrendIndex,
    tvi: tvi_mod.TickVolumeIndicator,
    mdi: mdi_mod.MeanDeviationIndex,
    cmi: cmi_mod.CandlestickMomentumIndex,
    csi: csi_mod.CandlestickStrengthIndex,
};

/// Nonambiguous Trend Filter (_Trade) by William Blau.
///
/// A post-processing transform applied to a normalized, signed base oscillator X
/// (TSI, SMI, DTI, TVI, MDI, CMI, CSI). It keeps X only where its sign and slope
/// agree and zeroes every ambiguous bar (book Ch. 8, Appendix B Figs. B-20..B-23):
///
///   X_Trade[k] = X[k]   if X[k] > 0 and X[k] - X[k-1] > 0   (positive and rising)
///              = X[k]   if X[k] < 0 and X[k] - X[k-1] < 0   (negative and falling)
///              = 0      otherwise                           (ambiguous)
///
/// The nonzero stretches correspond one-to-one with genuine up/down trends;
/// congestion and flat regions are blanked to zero.
///
/// The filter wraps an instance of the base indicator: every sample is routed to
/// the base with its own entity mapping, and the base's primary output is filtered.
///
/// Conventions: a NaN base value (the base's own look-back warm-up) yields NaN and
/// leaves the filter state untouched; the first finite base value has no prior
/// slope, so the output is 0.0; a flat step (delta == 0) is neither rising nor
/// falling, so it is zeroed.
pub const NonambiguousTrendFilter = struct {
    base: BaseIndicator,

    previous: f64,
    primed: bool,

    mnemonic_buf: [128]u8,
    mnemonic_len: usize,
    description_buf: [192]u8,
    description_len: usize,

    pub fn init(allocator: std.mem.Allocator, params: NonambiguousTrendFilterParams) !NonambiguousTrendFilter {
        const d = baseDefaults(params.base);
        const q = if (params.q != 0) params.q else d.q;
        const r = if (params.r != 0) params.r else d.r;
        const s = if (params.s != 0) params.s else d.s;
        const u = if (params.u != 0) params.u else d.u;

        // Price components are meaningful only for the single-price tsi and mdi bases.
        const uses_components = params.base == .tsi or params.base == .mdi;
        const bc: ?bar_component.BarComponent = if (uses_components) params.bar_component else null;
        const qc: ?quote_component.QuoteComponent = if (uses_components) params.quote_component else null;
        const tc: ?trade_component.TradeComponent = if (uses_components) params.trade_component else null;

        var triple_buf: [64]u8 = undefined;
        const triple = component_triple_mnemonic_mod.componentTripleMnemonic(
            &triple_buf,
            bc orelse bar_component.default_bar_component,
            qc orelse quote_component.default_quote_component,
            tc orelse trade_component.default_trade_component,
        );

        var mnemonic_buf: [128]u8 = undefined;
        const mnemonic_slice = (if (d.uses_q)
            std.fmt.bufPrint(&mnemonic_buf, "ntf.{s}({d},{d},{d},{d}{s})", .{ d.name, q, r, s, u, triple })
        else
            std.fmt.bufPrint(&mnemonic_buf, "ntf.{s}({d},{d},{d}{s})", .{ d.name, r, s, u, triple })) catch
            return error.MnemonicTooLong;
        const mnemonic_len = mnemonic_slice.len;

        var description_buf: [192]u8 = undefined;
        const desc_slice = std.fmt.bufPrint(&description_buf, "Nonambiguous Trend Filter {s}", .{mnemonic_slice}) catch
            return error.MnemonicTooLong;
        const description_len = desc_slice.len;

        // The base signal line is unused, so its period is 1.
        const base: BaseIndicator = switch (params.base) {
            .tsi => .{ .tsi = try tsi_mod.TrueStrengthIndex.init(allocator, .{
                .q = q,
                .r = r,
                .s = s,
                .u = u,
                .ul = 1,
                .bar_component = bc,
                .quote_component = qc,
                .trade_component = tc,
            }) },
            .smi => .{ .smi = try smi_mod.StochasticMomentumIndex.init(allocator, .{ .q = q, .r = r, .s = s, .u = u, .ul = 1 }) },
            .dti => .{ .dti = try dti_mod.DirectionalTrendIndex.init(allocator, .{ .q = q, .r = r, .s = s, .u = u, .ul = 1 }) },
            .tvi => .{ .tvi = try tvi_mod.TickVolumeIndicator.init(.{ .r = r, .s = s, .u = u }) },
            .mdi => .{ .mdi = try mdi_mod.MeanDeviationIndex.init(.{
                .r = r,
                .s = s,
                .u = u,
                .ul = 1,
                .bar_component = bc,
                .quote_component = qc,
                .trade_component = tc,
            }) },
            .cmi => .{ .cmi = try cmi_mod.CandlestickMomentumIndex.init(.{ .r = r, .s = s, .u = u, .ul = 1 }) },
            .csi => .{ .csi = try csi_mod.CandlestickStrengthIndex.init(.{ .r = r, .s = s, .u = u, .ul = 1 }) },
        };

        return .{
            .base = base,
            .previous = 0.0,
            .primed = false,
            .mnemonic_buf = mnemonic_buf,
            .mnemonic_len = mnemonic_len,
            .description_buf = description_buf,
            .description_len = description_len,
        };
    }

    pub fn deinit(self: *NonambiguousTrendFilter) void {
        switch (self.base) {
            inline else => |*b| if (@hasDecl(@TypeOf(b.*), "deinit")) b.deinit(),
        }
    }

    pub fn fixSlices(self: *NonambiguousTrendFilter) void {
        // NTF doesn't use LineIndicator; mnemonic/description are read from the
        // owned buffers directly, so only the wrapped base needs a fixup.
        switch (self.base) {
            inline else => |*b| b.fixSlices(),
        }
    }

    /// Keeps x when positive-and-rising or negative-and-falling, else 0.
    fn filter(self: *NonambiguousTrendFilter, x: f64) f64 {
        // The base is still warming up: do not touch the filter state.
        if (math.isNan(x)) return math.nan(f64);

        if (!self.primed) {
            // First finite value: no prior slope, hence ambiguous.
            self.previous = x;
            self.primed = true;
            return 0.0;
        }

        const delta = x - self.previous;
        self.previous = x;

        if (x > 0.0 and delta > 0.0) return x; // positive and rising
        if (x < 0.0 and delta < 0.0) return x; // negative and falling
        return 0.0; // ambiguous / flat / congestion
    }

    /// Filters the primary output of the base.
    fn filterOutput(self: *NonambiguousTrendFilter, output: OutputArray) f64 {
        return self.filter(output.slice()[0].scalar.value);
    }

    /// Updates the indicator given the next single sample value, fed to the base as a scalar.
    pub fn update(self: *NonambiguousTrendFilter, sample: f64) f64 {
        return self.filterOutput(self.baseUpdateScalar(&Scalar{ .time = 0, .value = sample }));
    }

    fn baseUpdateScalar(self: *NonambiguousTrendFilter, sample: *const Scalar) OutputArray {
        return switch (self.base) {
            inline else => |*b| b.updateScalar(sample),
        };
    }

    /// Returns whether the indicator has accumulated enough data.
    pub fn isPrimed(self: *const NonambiguousTrendFilter) bool {
        return self.primed;
    }

    fn mnemonic(self: *const NonambiguousTrendFilter) []const u8 {
        return self.mnemonic_buf[0..self.mnemonic_len];
    }

    fn description(self: *const NonambiguousTrendFilter) []const u8 {
        return self.description_buf[0..self.description_len];
    }

    /// Returns metadata for this indicator.
    pub fn getMetadata(self: *const NonambiguousTrendFilter, out: *Metadata) void {
        const mn = self.mnemonic();
        const desc = self.description();

        build_metadata_mod.buildMetadata(
            out,
            .nonambiguous_trend_filter,
            mn,
            desc,
            &[_]build_metadata_mod.OutputText{
                .{ .mnemonic = mn, .description = desc },
            },
        );
    }

    pub fn updateScalar(self: *NonambiguousTrendFilter, sample: *const Scalar) OutputArray {
        return wrap(sample.time, self.filterOutput(self.baseUpdateScalar(sample)));
    }

    pub fn updateBar(self: *NonambiguousTrendFilter, sample: *const Bar) OutputArray {
        const output = switch (self.base) {
            inline else => |*b| b.updateBar(sample),
        };
        return wrap(sample.time, self.filterOutput(output));
    }

    pub fn updateQuote(self: *NonambiguousTrendFilter, sample: *const Quote) OutputArray {
        const output = switch (self.base) {
            inline else => |*b| b.updateQuote(sample),
        };
        return wrap(sample.time, self.filterOutput(output));
    }

    pub fn updateTrade(self: *NonambiguousTrendFilter, sample: *const Trade) OutputArray {
        const output = switch (self.base) {
            inline else => |*b| b.updateTrade(sample),
        };
        return wrap(sample.time, self.filterOutput(output));
    }

    fn wrap(time: i64, value: f64) OutputArray {
        var out = OutputArray{};
        out.append(.{ .scalar = .{ .time = time, .value = value } });
        return out;
    }

    /// Returns an Indicator interface backed by this instance.
    pub fn indicator(self: *NonambiguousTrendFilter) indicator_mod.Indicator {
        return .{
            .ptr = @ptrCast(self),
            .vtable = &vtable,
        };
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
        const self: *NonambiguousTrendFilter = @ptrCast(@alignCast(ptr));
        return self.isPrimed();
    }

    fn vtableMetadata(ptr: *anyopaque, out: *Metadata) void {
        const self: *const NonambiguousTrendFilter = @ptrCast(@alignCast(ptr));
        self.getMetadata(out);
    }

    fn vtableUpdateScalar(ptr: *anyopaque, sample: *const Scalar) OutputArray {
        const self: *NonambiguousTrendFilter = @ptrCast(@alignCast(ptr));
        return self.updateScalar(sample);
    }

    fn vtableUpdateBar(ptr: *anyopaque, sample: *const Bar) OutputArray {
        const self: *NonambiguousTrendFilter = @ptrCast(@alignCast(ptr));
        return self.updateBar(sample);
    }

    fn vtableUpdateQuote(ptr: *anyopaque, sample: *const Quote) OutputArray {
        const self: *NonambiguousTrendFilter = @ptrCast(@alignCast(ptr));
        return self.updateQuote(sample);
    }

    fn vtableUpdateTrade(ptr: *anyopaque, sample: *const Trade) OutputArray {
        const self: *NonambiguousTrendFilter = @ptrCast(@alignCast(ptr));
        return self.updateTrade(sample);
    }
};

// ---------------------------------------------------------------------------
// Tests
// ---------------------------------------------------------------------------

const testing = std.testing;
const testdata = @import("testdata.zig");

const test_tolerance = 1e-13;

const Combo = struct {
    base: NonambiguousTrendFilterBase,
    q: usize,
    r: usize,
    s: usize,
    u: usize,
    expected: [252]f64,
};

// q=0 where the base does not use it.
const combos = [_]Combo{
    .{ .base = .tsi, .q = 2, .r = 32, .s = 13, .u = 3, .expected = testdata.expectedTSI_R32_S13_U3() },
    .{ .base = .tsi, .q = 2, .r = 20, .s = 5, .u = 3, .expected = testdata.expectedTSI_R20_S5_U3() },
    .{ .base = .tsi, .q = 2, .r = 40, .s = 20, .u = 5, .expected = testdata.expectedTSI_R40_S20_U5() },
    .{ .base = .smi, .q = 32, .r = 64, .s = 7, .u = 1, .expected = testdata.expectedSMI_Q32_R64_S7_U1() },
    .{ .base = .smi, .q = 5, .r = 20, .s = 5, .u = 3, .expected = testdata.expectedSMI_Q5_R20_S5_U3() },
    .{ .base = .smi, .q = 13, .r = 25, .s = 2, .u = 1, .expected = testdata.expectedSMI_Q13_R25_S2_U1() },
    .{ .base = .dti, .q = 2, .r = 28, .s = 28, .u = 5, .expected = testdata.expectedDTI_Q2_R28_S28_U5() },
    .{ .base = .dti, .q = 2, .r = 20, .s = 5, .u = 3, .expected = testdata.expectedDTI_Q2_R20_S5_U3() },
    .{ .base = .dti, .q = 4, .r = 14, .s = 14, .u = 3, .expected = testdata.expectedDTI_Q4_R14_S14_U3() },
    .{ .base = .mdi, .q = 0, .r = 20, .s = 5, .u = 3, .expected = testdata.expectedMDI_R20_S5_U3() },
    .{ .base = .mdi, .q = 0, .r = 40, .s = 5, .u = 3, .expected = testdata.expectedMDI_R40_S5_U3() },
    .{ .base = .cmi, .q = 0, .r = 20, .s = 5, .u = 3, .expected = testdata.expectedCMI_R20_S5_U3() },
    .{ .base = .cmi, .q = 0, .r = 10, .s = 5, .u = 3, .expected = testdata.expectedCMI_R10_S5_U3() },
    .{ .base = .csi, .q = 0, .r = 32, .s = 32, .u = 1, .expected = testdata.expectedCSI_R32_S32_U1() },
    .{ .base = .csi, .q = 0, .r = 20, .s = 5, .u = 3, .expected = testdata.expectedCSI_R20_S5_U3() },
    .{ .base = .csi, .q = 0, .r = 1, .s = 1, .u = 1, .expected = testdata.expectedCSI_R1_S1_U1() },
    .{ .base = .tvi, .q = 0, .r = 32, .s = 32, .u = 5, .expected = testdata.expectedTVI_R32_S32_U5() },
    .{ .base = .tvi, .q = 0, .r = 12, .s = 12, .u = 1, .expected = testdata.expectedTVI_R12_S12_U1() },
    .{ .base = .tvi, .q = 0, .r = 25, .s = 13, .u = 1, .expected = testdata.expectedTVI_R25_S13_U1() },
};

fn checkVal(exp: f64, act: f64, tolerance: f64) !void {
    if (math.isNan(exp)) {
        try testing.expect(math.isNan(act));
        return;
    }

    try testing.expect(@abs(act - exp) <= tolerance);
}

const TestData = struct {
    open: [252]f64,
    high: [252]f64,
    low: [252]f64,
    close: [252]f64,

    fn load() TestData {
        return .{
            .open = testdata.testOpen(),
            .high = testdata.testHigh(),
            .low = testdata.testLow(),
            .close = testdata.testInput(),
        };
    }

    fn bar(self: *const TestData, i: usize) Bar {
        return .{
            .time = @intCast(i),
            .open = self.open[i],
            .high = self.high[i],
            .low = self.low[i],
            .close = self.close[i],
            .volume = 0.0,
        };
    }
};

test "NTF reference data from bars" {
    const allocator = testing.allocator;
    const data = TestData.load();

    for (combos) |combo| {
        var ind = try NonambiguousTrendFilter.init(allocator, .{ .base = combo.base, .q = combo.q, .r = combo.r, .s = combo.s, .u = combo.u });
        defer ind.deinit();

        for (0..252) |i| {
            const out = ind.updateBar(&data.bar(i));
            try testing.expectEqual(@as(usize, 1), out.slice().len);
            try checkVal(combo.expected[i], out.slice()[0].scalar.value, test_tolerance);
        }
    }
}

test "NTF reference data from samples" {
    const allocator = testing.allocator;
    const input = testdata.testInput();

    for (combos) |combo| {
        if (combo.base != .tsi and combo.base != .mdi) continue;

        var ind = try NonambiguousTrendFilter.init(allocator, .{ .base = combo.base, .q = combo.q, .r = combo.r, .s = combo.s, .u = combo.u });
        defer ind.deinit();

        for (0..252) |i| {
            try checkVal(combo.expected[i], ind.update(input[i]), test_tolerance);
        }
    }
}

test "NTF keep/zero rule on a passthrough CSI base" {
    const allocator = testing.allocator;

    // A passthrough CSI base is 100*(C-O)/(H-L).
    var ind = try NonambiguousTrendFilter.init(allocator, .{ .base = .csi, .r = 1, .s = 1, .u = 1 });
    defer ind.deinit();

    const Case = struct { close: f64, expected: f64 };
    const cases = [_]Case{
        .{ .close = 11.0, .expected = 0.0 }, // first finite value
        .{ .close = 12.0, .expected = 50.0 }, // positive and rising
        .{ .close = 11.0, .expected = 0.0 }, // positive and falling
        .{ .close = 9.0, .expected = -25.0 }, // negative and falling
        .{ .close = 9.0, .expected = 0.0 }, // flat
        .{ .close = 9.5, .expected = 0.0 }, // negative and rising
    };

    for (cases) |case| {
        const out = ind.updateBar(&Bar{ .time = 0, .open = 10.0, .high = 12.0, .low = 8.0, .close = case.close, .volume = 0.0 });
        try testing.expectEqual(case.expected, out.slice()[0].scalar.value);
    }
}

test "NTF warm-up region" {
    const allocator = testing.allocator;
    const data = TestData.load();
    const q = 13;

    var ind = try NonambiguousTrendFilter.init(allocator, .{ .base = .smi, .q = q, .r = 25, .s = 2, .u = 1 });
    defer ind.deinit();

    for (0..q - 1) |i| {
        try testing.expect(math.isNan(ind.updateBar(&data.bar(i)).slice()[0].scalar.value));
        try testing.expect(!ind.isPrimed());
    }

    try testing.expectEqual(@as(f64, 0.0), ind.updateBar(&data.bar(q - 1)).slice()[0].scalar.value);
    try testing.expect(ind.isPrimed());
}

test "NTF entity updates" {
    const allocator = testing.allocator;
    const data = TestData.load();
    const expected = testdata.expectedTSI_R32_S13_U3();

    var scalar_ind = try NonambiguousTrendFilter.init(allocator, .{});
    defer scalar_ind.deinit();
    var bar_ind = try NonambiguousTrendFilter.init(allocator, .{});
    defer bar_ind.deinit();
    var quote_ind = try NonambiguousTrendFilter.init(allocator, .{});
    defer quote_ind.deinit();
    var trade_ind = try NonambiguousTrendFilter.init(allocator, .{});
    defer trade_ind.deinit();

    for (0..252) |i| {
        const time: i64 = @intCast(i);
        const v = data.close[i];

        const scalar_out = scalar_ind.updateScalar(&Scalar{ .time = time, .value = v });
        try testing.expectEqual(@as(usize, 1), scalar_out.slice().len);
        try testing.expectEqual(time, scalar_out.slice()[0].scalar.time);
        try checkVal(expected[i], scalar_out.slice()[0].scalar.value, test_tolerance);

        const bar_out = bar_ind.updateBar(&data.bar(i));
        try checkVal(expected[i], bar_out.slice()[0].scalar.value, test_tolerance);

        const quote_out = quote_ind.updateQuote(&Quote{ .time = time, .bid_price = v, .bid_size = 0.0, .ask_price = v, .ask_size = 0.0 });
        try checkVal(expected[i], quote_out.slice()[0].scalar.value, test_tolerance);

        const trade_out = trade_ind.updateTrade(&Trade{ .time = time, .price = v, .volume = 0.0 });
        try checkVal(expected[i], trade_out.slice()[0].scalar.value, test_tolerance);
    }
}

test "NTF metadata default" {
    const allocator = testing.allocator;

    var ind = try NonambiguousTrendFilter.init(allocator, .{});
    defer ind.deinit();

    var meta: Metadata = undefined;
    ind.getMetadata(&meta);

    try testing.expectEqual(Identifier.nonambiguous_trend_filter, meta.identifier);
    try testing.expectEqualStrings("ntf.tsi(2,32,13,3)", meta.mnemonic);
    try testing.expectEqualStrings("Nonambiguous Trend Filter ntf.tsi(2,32,13,3)", meta.description);
    try testing.expectEqual(@as(usize, 1), meta.outputs_len);
    try testing.expectEqual(@as(u8, 1), meta.outputs_buf[0].kind);
}

test "NTF mnemonics" {
    const allocator = testing.allocator;

    const Case = struct {
        params: NonambiguousTrendFilterParams,
        expected: []const u8,
    };

    const cases = [_]Case{
        .{ .params = .{}, .expected = "ntf.tsi(2,32,13,3)" },
        .{ .params = .{ .base = .smi }, .expected = "ntf.smi(32,64,7,1)" },
        .{ .params = .{ .base = .dti }, .expected = "ntf.dti(2,28,28,5)" },
        .{ .params = .{ .base = .tvi }, .expected = "ntf.tvi(32,32,5)" },
        .{ .params = .{ .base = .mdi }, .expected = "ntf.mdi(20,5,3)" },
        .{ .params = .{ .base = .cmi }, .expected = "ntf.cmi(20,5,3)" },
        .{ .params = .{ .base = .csi }, .expected = "ntf.csi(32,32,1)" },
        .{ .params = .{ .base = .smi, .q = 5, .r = 20, .s = 5, .u = 3 }, .expected = "ntf.smi(5,20,5,3)" },
        .{ .params = .{ .bar_component = .median }, .expected = "ntf.tsi(2,32,13,3, hl/2)" },
        .{ .params = .{ .quote_component = .bid }, .expected = "ntf.tsi(2,32,13,3, b)" },
        .{ .params = .{ .trade_component = .volume }, .expected = "ntf.tsi(2,32,13,3, v)" },
        .{ .params = .{ .bar_component = .open, .quote_component = .bid }, .expected = "ntf.tsi(2,32,13,3, o, b)" },
        .{ .params = .{ .bar_component = .high, .trade_component = .volume }, .expected = "ntf.tsi(2,32,13,3, h, v)" },
        .{ .params = .{ .quote_component = .ask, .trade_component = .volume }, .expected = "ntf.tsi(2,32,13,3, a, v)" },
        .{ .params = .{ .base = .mdi, .bar_component = .median }, .expected = "ntf.mdi(20,5,3, hl/2)" },
        .{ .params = .{ .base = .csi, .bar_component = .median }, .expected = "ntf.csi(32,32,1)" },
    };

    for (cases) |case| {
        var ind = try NonambiguousTrendFilter.init(allocator, case.params);
        defer ind.deinit();

        try testing.expectEqualStrings(case.expected, ind.mnemonic());
    }
}
