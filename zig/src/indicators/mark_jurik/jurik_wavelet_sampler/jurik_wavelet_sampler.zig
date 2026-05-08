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
const metadata_mod = @import("../../core/metadata.zig");

const OutputArray = indicator_mod.OutputArray;
const LineIndicator = line_indicator_mod.LineIndicator;
const Metadata = metadata_mod.Metadata;

/// Enumerates the outputs of the Jurik Wavelet Sampler indicator.
pub const JurikWaveletSamplerOutput = enum(u8) {
    /// The first column value.
    value = 1,
};

/// Parameters for the Jurik Wavelet Sampler.
pub const JurikWaveletSamplerParams = struct {
    /// Index controls the number of output columns (1..18).
    index: u32 = 12,
    bar_component: ?bar_component.BarComponent = null,
    quote_component: ?quote_component.QuoteComponent = null,
    trade_component: ?trade_component.TradeComponent = null,
};

/// (n, M) parameters for each wavelet column.
const NmEntry = struct { n: u32, m: u32 };

const nm_table = [18]NmEntry{
    .{ .n = 1, .m = 0 },
    .{ .n = 2, .m = 0 },
    .{ .n = 3, .m = 0 },
    .{ .n = 4, .m = 0 },
    .{ .n = 5, .m = 0 },
    .{ .n = 7, .m = 2 },
    .{ .n = 10, .m = 2 },
    .{ .n = 14, .m = 4 },
    .{ .n = 19, .m = 4 },
    .{ .n = 26, .m = 8 },
    .{ .n = 35, .m = 8 },
    .{ .n = 48, .m = 16 },
    .{ .n = 65, .m = 16 },
    .{ .n = 90, .m = 32 },
    .{ .n = 123, .m = 32 },
    .{ .n = 172, .m = 64 },
    .{ .n = 237, .m = 64 },
    .{ .n = 334, .m = 128 },
};

const MAX_PRICES = 1024;

/// Jurik Wavelet Sampler (WAV).
/// Produces `index` output columns per bar, each representing a different
/// multi-resolution scale.
pub const JurikWaveletSampler = struct {
    line: LineIndicator,
    primed: bool,
    index: u32,
    max_lookback: u32,
    prices: [MAX_PRICES]f64,
    bar_count: u32,
    columns: [18]f64,

    mnemonic_buf: [96]u8,
    mnemonic_len: usize,

    pub fn init(params: JurikWaveletSamplerParams) !JurikWaveletSampler {
        const index = params.index;

        if (index < 1 or index > 18) return error.InvalidIndex;

        const bc = params.bar_component orelse bar_component.default_bar_component;
        const qc = params.quote_component orelse quote_component.default_quote_component;
        const tc = params.trade_component orelse trade_component.default_trade_component;

        var triple_buf: [64]u8 = undefined;
        const triple = component_triple_mnemonic_mod.componentTripleMnemonic(&triple_buf, bc, qc, tc);

        var mnemonic_buf: [96]u8 = undefined;
        const mnemonic = std.fmt.bufPrint(&mnemonic_buf, "jwav({d}{s})", .{
            index, triple,
        }) catch unreachable;
        const mnemonic_len = mnemonic.len;

        // Compute max lookback.
        var max_lookback: u32 = 0;
        for (0..index) |c| {
            const lb = nm_table[c].n + nm_table[c].m / 2;
            if (lb > max_lookback) max_lookback = lb;
        }

        return .{
            .line = LineIndicator.new(
                mnemonic_buf[0..mnemonic_len],
                "Jurik wavelet sampler ",
                params.bar_component,
                params.quote_component,
                params.trade_component,
            ),
            .primed = false,
            .index = index,
            .max_lookback = max_lookback,
            .prices = [_]f64{0} ** MAX_PRICES,
            .bar_count = 0,
            .columns = [_]f64{0} ** 18,
            .mnemonic_buf = mnemonic_buf,
            .mnemonic_len = mnemonic_len,
        };
    }

    pub fn fixSlices(self: *JurikWaveletSampler) void {
        self.line.mnemonic = self.mnemonic_buf[0..self.mnemonic_len];
    }

    /// Returns a copy of the current column values.
    pub fn getColumns(self: *const JurikWaveletSampler) [18]f64 {
        return self.columns;
    }

    pub fn update(self: *JurikWaveletSampler, sample: f64) f64 {
        if (math.isNan(sample)) return sample;

        self.prices[self.bar_count] = sample;
        self.bar_count += 1;

        var all_valid = true;

        for (0..self.index) |c| {
            const n = nm_table[c].n;
            const m = nm_table[c].m;
            const dead_zone = n + m / 2;

            if (self.bar_count <= dead_zone) {
                self.columns[c] = math.nan(f64);
                all_valid = false;
            } else {
                if (m == 0) {
                    // Simple lag.
                    self.columns[c] = self.prices[self.bar_count - 1 - n];
                } else {
                    // Mean of (M+1) prices centered at lag n.
                    const half = m / 2;
                    const center_idx = self.bar_count - 1 - n;
                    var total: f64 = 0;

                    for (0..m + 1) |k| {
                        total += self.prices[center_idx - half + @as(u32, @intCast(k))];
                    }

                    self.columns[c] = total / @as(f64, @floatFromInt(m + 1));
                }
            }
        }

        if (all_valid) {
            self.primed = true;
        }

        return self.columns[0];
    }

    pub fn isPrimed(self: *const JurikWaveletSampler) bool {
        return self.primed;
    }

    pub fn getMetadata(self: *const JurikWaveletSampler, out: *Metadata) void {
        build_metadata_mod.buildMetadata(
            out,
            .jurik_wavelet_sampler,
            self.line.mnemonic,
            self.line.description,
            &[_]build_metadata_mod.OutputText{
                .{ .mnemonic = self.line.mnemonic, .description = self.line.description },
            },
        );
    }

    pub fn updateScalar(self: *JurikWaveletSampler, sample: *const Scalar) OutputArray {
        const value = self.update(sample.value);
        return LineIndicator.wrapScalar(sample.time, value);
    }

    pub fn updateBar(self: *JurikWaveletSampler, sample: *const Bar) OutputArray {
        const value = self.update(self.line.extractBar(sample));
        return LineIndicator.wrapScalar(sample.time, value);
    }

    pub fn updateQuote(self: *JurikWaveletSampler, sample: *const Quote) OutputArray {
        const value = self.update(self.line.extractQuote(sample));
        return LineIndicator.wrapScalar(sample.time, value);
    }

    pub fn updateTrade(self: *JurikWaveletSampler, sample: *const Trade) OutputArray {
        const value = self.update(self.line.extractTrade(sample));
        return LineIndicator.wrapScalar(sample.time, value);
    }

    pub fn indicator(self: *JurikWaveletSampler) indicator_mod.Indicator {
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
        const self: *JurikWaveletSampler = @ptrCast(@alignCast(ptr));
        return self.isPrimed();
    }

    fn vtableMetadata(ptr: *anyopaque, out: *Metadata) void {
        const self: *const JurikWaveletSampler = @ptrCast(@alignCast(ptr));
        self.getMetadata(out);
    }

    fn vtableUpdateScalar(ptr: *anyopaque, sample: *const Scalar) OutputArray {
        const self: *JurikWaveletSampler = @ptrCast(@alignCast(ptr));
        return self.updateScalar(sample);
    }

    fn vtableUpdateBar(ptr: *anyopaque, sample: *const Bar) OutputArray {
        const self: *JurikWaveletSampler = @ptrCast(@alignCast(ptr));
        return self.updateBar(sample);
    }

    fn vtableUpdateQuote(ptr: *anyopaque, sample: *const Quote) OutputArray {
        const self: *JurikWaveletSampler = @ptrCast(@alignCast(ptr));
        return self.updateQuote(sample);
    }

    fn vtableUpdateTrade(ptr: *anyopaque, sample: *const Trade) OutputArray {
        const self: *JurikWaveletSampler = @ptrCast(@alignCast(ptr));
        return self.updateTrade(sample);
    }

    pub const Error = error{
        InvalidIndex,
    };
};

// ---------------------------------------------------------------------------
// Tests
// ---------------------------------------------------------------------------

const testing = std.testing;
const testdata = @import("testdata.zig");

fn almostEqual(a: f64, b: f64, epsilon: f64) bool {
    return @abs(a - b) <= epsilon;
}

test "jwav index=12 default" {
    var ind = JurikWaveletSampler.init(.{ .index = 12 }) catch unreachable;
    ind.fixSlices();
    const input = testdata.testInput();
    const eps = 1e-13;

    const expected_cols = [12]*const fn () [252]f64{
        &testdata.expectedWAVCol0, &testdata.expectedWAVCol1, &testdata.expectedWAVCol2,
        &testdata.expectedWAVCol3, &testdata.expectedWAVCol4, &testdata.expectedWAVCol5,
        &testdata.expectedWAVCol6, &testdata.expectedWAVCol7, &testdata.expectedWAVCol8,
        &testdata.expectedWAVCol9, &testdata.expectedWAVCol10, &testdata.expectedWAVCol11,
    };

    for (0..252) |i| {
        const result = ind.update(input[i]);
        const exp0 = expected_cols[0]();

        if (math.isNan(exp0[i])) {
            try testing.expect(math.isNan(result));
        } else {
            try testing.expect(!math.isNan(result));
            if (!almostEqual(result, exp0[i], eps)) {
                std.debug.print("FAIL col0 [{d}]: expected {d}, got {d}\n", .{ i, exp0[i], result });
                return error.TestUnexpectedResult;
            }
        }

        // Check all columns.
        const cols = ind.getColumns();
        for (0..12) |c| {
            const exp = expected_cols[c]();
            if (math.isNan(exp[i])) {
                try testing.expect(math.isNan(cols[c]));
            } else {
                if (!almostEqual(cols[c], exp[i], eps)) {
                    std.debug.print("FAIL col{d} [{d}]: expected {d}, got {d}\n", .{ c, i, exp[i], cols[c] });
                    return error.TestUnexpectedResult;
                }
            }
        }
    }
}

test "jwav index=6" {
    var ind = JurikWaveletSampler.init(.{ .index = 6 }) catch unreachable;
    ind.fixSlices();
    const input = testdata.testInput();
    const eps = 1e-13;

    const expected_cols = [6]*const fn () [252]f64{
        &testdata.expectedIndex6Col0, &testdata.expectedIndex6Col1, &testdata.expectedIndex6Col2,
        &testdata.expectedIndex6Col3, &testdata.expectedIndex6Col4, &testdata.expectedIndex6Col5,
    };

    for (0..252) |i| {
        _ = ind.update(input[i]);
        const cols = ind.getColumns();

        for (0..6) |c| {
            const exp = expected_cols[c]();
            if (math.isNan(exp[i])) {
                try testing.expect(math.isNan(cols[c]));
            } else {
                if (!almostEqual(cols[c], exp[i], eps)) {
                    std.debug.print("FAIL idx6 col{d} [{d}]: expected {d}, got {d}\n", .{ c, i, exp[i], cols[c] });
                    return error.TestUnexpectedResult;
                }
            }
        }
    }
}

test "jwav index=16" {
    var ind = JurikWaveletSampler.init(.{ .index = 16 }) catch unreachable;
    ind.fixSlices();
    const input = testdata.testInput();
    const eps = 1e-13;

    const expected_cols = [16]*const fn () [252]f64{
        &testdata.expectedIndex16Col0, &testdata.expectedIndex16Col1, &testdata.expectedIndex16Col2,
        &testdata.expectedIndex16Col3, &testdata.expectedIndex16Col4, &testdata.expectedIndex16Col5,
        &testdata.expectedIndex16Col6, &testdata.expectedIndex16Col7, &testdata.expectedIndex16Col8,
        &testdata.expectedIndex16Col9, &testdata.expectedIndex16Col10, &testdata.expectedIndex16Col11,
        &testdata.expectedIndex16Col12, &testdata.expectedIndex16Col13, &testdata.expectedIndex16Col14,
        &testdata.expectedIndex16Col15,
    };

    for (0..252) |i| {
        _ = ind.update(input[i]);
        const cols = ind.getColumns();

        for (0..16) |c| {
            const exp = expected_cols[c]();
            if (math.isNan(exp[i])) {
                try testing.expect(math.isNan(cols[c]));
            } else {
                if (!almostEqual(cols[c], exp[i], eps)) {
                    std.debug.print("FAIL idx16 col{d} [{d}]: expected {d}, got {d}\n", .{ c, i, exp[i], cols[c] });
                    return error.TestUnexpectedResult;
                }
            }
        }
    }
}
