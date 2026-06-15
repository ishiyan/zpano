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

/// Enumerates the outputs of the Parabolic Vertex indicator.
pub const ParabolicVertexOutput = enum(u8) {
    /// The bars-to-near-turn value.
    value = 1,
};

/// Parameters to create a Parabolic Vertex indicator.
///
/// The indicator has no numeric parameters; only the input price component is configurable.
pub const ParabolicVertexParams = struct {
    bar_component: ?bar_component.BarComponent = null,
    quote_component: ?quote_component.QuoteComponent = null,
    trade_component: ?trade_component.TradeComponent = null,
};

/// Parabolic Vertex (PVTX) by Don Mak.
///
/// Predicts turning points by fitting a parabola to the 3 most recent price points
/// and computing where the vertex (extremum) occurs relative to the current bar:
///   t_v = -(1.5*x(n) - 2*x(n-1) + 0.5*x(n-2)) / (x(n) - 2*x(n-1) + x(n-2))
/// The output is the number of bars from the current bar to the predicted turning point.
pub const ParabolicVertex = struct {
    line: LineIndicator,

    buffer: [3]f64 = .{ 0.0, 0.0, 0.0 },
    index: usize = 0,
    count: usize = 0,

    primed: bool = false,

    mnemonic_buf: [64]u8,
    mnemonic_len: usize,
    description_buf: [128]u8,
    description_len: usize,

    pub fn init(params: ParabolicVertexParams) ParabolicVertex {
        const bc = params.bar_component orelse bar_component.default_bar_component;
        const qc = params.quote_component orelse quote_component.default_quote_component;
        const tc = params.trade_component orelse trade_component.default_trade_component;

        var triple_buf: [64]u8 = undefined;
        const triple = component_triple_mnemonic_mod.componentTripleMnemonic(&triple_buf, bc, qc, tc);

        var mnemonic_buf: [64]u8 = undefined;
        var mnemonic_len: usize = undefined;
        if (triple.len == 0) {
            const s = std.fmt.bufPrint(&mnemonic_buf, "pvtx", .{}) catch unreachable;
            mnemonic_len = s.len;
        } else {
            // Strip leading ", " from triple.
            const suffix = if (triple.len > 2) triple[2..] else triple;
            const s = std.fmt.bufPrint(&mnemonic_buf, "pvtx({s})", .{suffix}) catch unreachable;
            mnemonic_len = s.len;
        }

        var description_buf: [128]u8 = undefined;
        const desc_slice = std.fmt.bufPrint(&description_buf, "Parabolic vertex {s}", .{mnemonic_buf[0..mnemonic_len]}) catch unreachable;
        const description_len = desc_slice.len;

        return .{
            .line = LineIndicator.new(
                mnemonic_buf[0..mnemonic_len],
                description_buf[0..description_len],
                params.bar_component,
                params.quote_component,
                params.trade_component,
            ),
            .mnemonic_buf = mnemonic_buf,
            .mnemonic_len = mnemonic_len,
            .description_buf = description_buf,
            .description_len = description_len,
        };
    }

    pub fn fixSlices(self: *ParabolicVertex) void {
        self.line.mnemonic = self.mnemonic_buf[0..self.mnemonic_len];
        self.line.description = self.description_buf[0..self.description_len];
    }

    pub fn update(self: *ParabolicVertex, sample: f64) f64 {
        // Store the price in the ring buffer.
        self.buffer[self.index] = sample;
        self.index = (self.index + 1) % 3;
        self.count += 1;

        if (self.count < 3) {
            self.primed = false;
            return math.nan(f64);
        }

        self.primed = true;

        // Extract prices: x[n] (newest), x[n-1], x[n-2] (oldest).
        const idx: i64 = @intCast(self.index);
        const xn = self.buffer[@intCast(@mod(idx - 1, 3))];
        const xn1 = self.buffer[@intCast(@mod(idx - 2, 3))];
        const xn2 = self.buffer[@intCast(@mod(idx - 3, 3))];

        // Denominator = second-order finite difference (proportional to curvature).
        const denom = xn - 2.0 * xn1 + xn2;
        if (denom == 0.0) {
            return math.nan(f64);
        }

        const numer = 1.5 * xn - 2.0 * xn1 + 0.5 * xn2;

        return -numer / denom;
    }

    pub fn isPrimed(self: *const ParabolicVertex) bool {
        return self.primed;
    }

    fn mnemonic(self: *const ParabolicVertex) []const u8 {
        return self.mnemonic_buf[0..self.mnemonic_len];
    }

    fn description(self: *const ParabolicVertex) []const u8 {
        return self.description_buf[0..self.description_len];
    }

    pub fn getMetadata(self: *const ParabolicVertex, out: *Metadata) void {
        const mn = self.mnemonic();
        const desc = self.description();
        build_metadata_mod.buildMetadata(
            out,
            .parabolic_vertex,
            mn,
            desc,
            &[_]build_metadata_mod.OutputText{
                .{ .mnemonic = mn, .description = desc },
            },
        );
    }

    pub fn updateScalar(self: *ParabolicVertex, sample: *const Scalar) OutputArray {
        const value = self.update(sample.value);
        return LineIndicator.wrapScalar(sample.time, value);
    }

    pub fn updateBar(self: *ParabolicVertex, sample: *const Bar) OutputArray {
        const value = self.update(self.line.extractBar(sample));
        return LineIndicator.wrapScalar(sample.time, value);
    }

    pub fn updateQuote(self: *ParabolicVertex, sample: *const Quote) OutputArray {
        const value = self.update(self.line.extractQuote(sample));
        return LineIndicator.wrapScalar(sample.time, value);
    }

    pub fn updateTrade(self: *ParabolicVertex, sample: *const Trade) OutputArray {
        const value = self.update(self.line.extractTrade(sample));
        return LineIndicator.wrapScalar(sample.time, value);
    }

    pub fn indicator(self: *ParabolicVertex) indicator_mod.Indicator {
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
        const self: *ParabolicVertex = @ptrCast(@alignCast(ptr));
        return self.isPrimed();
    }
    fn vtableMetadata(ptr: *anyopaque, out: *Metadata) void {
        const self: *const ParabolicVertex = @ptrCast(@alignCast(ptr));
        self.getMetadata(out);
    }
    fn vtableUpdateScalar(ptr: *anyopaque, sample: *const Scalar) OutputArray {
        const self: *ParabolicVertex = @ptrCast(@alignCast(ptr));
        return self.updateScalar(sample);
    }
    fn vtableUpdateBar(ptr: *anyopaque, sample: *const Bar) OutputArray {
        const self: *ParabolicVertex = @ptrCast(@alignCast(ptr));
        return self.updateBar(sample);
    }
    fn vtableUpdateQuote(ptr: *anyopaque, sample: *const Quote) OutputArray {
        const self: *ParabolicVertex = @ptrCast(@alignCast(ptr));
        return self.updateQuote(sample);
    }
    fn vtableUpdateTrade(ptr: *anyopaque, sample: *const Trade) OutputArray {
        const self: *ParabolicVertex = @ptrCast(@alignCast(ptr));
        return self.updateTrade(sample);
    }
};

// ---------------------------------------------------------------------------
// Tests
// ---------------------------------------------------------------------------

const testing = std.testing;
const testdata = @import("testdata.zig");

const tolerance = 1e-9;

fn checkSeries(name: []const u8, inputs: []const f64, expected: []const f64) !void {
    _ = name;

    var pvtx = ParabolicVertex.init(.{});

    try testing.expectEqual(inputs.len, expected.len);

    for (0..inputs.len) |i| {
        const value = pvtx.update(inputs[i]);
        const exp = expected[i];
        if (math.isNan(exp)) {
            try testing.expect(math.isNan(value));
        } else {
            // Combined absolute + relative tolerance (ill-conditioned near collinear points).
            const delta = tolerance * @max(1.0, @abs(exp));
            try testing.expect(@abs(value - exp) <= delta);
        }
    }
}

test "PVTX reference data" {
    const input_close = testdata.iNPUT_CLOSE();
    const input_ema6 = testdata.iNPUT_EMA6();
    const input_ema20 = testdata.iNPUT_EMA20();
    const expected_raw = testdata.expectedRAW();
    const expected_ema6 = testdata.expectedEMA6();
    const expected_ema20 = testdata.expectedEMA20();
    const test1_input = testdata.tEST1_INPUT_PARABOLA();
    const test1_expected = testdata.tEST1_EXPECTED();

    try checkSeries("RAW", &input_close, &expected_raw);
    try checkSeries("EMA6", &input_ema6, &expected_ema6);
    try checkSeries("EMA20", &input_ema20, &expected_ema20);
    try checkSeries("TEST1", &test1_input, &test1_expected);
}

test "PVTX metadata default" {
    var pvtx = ParabolicVertex.init(.{});
    pvtx.fixSlices();

    var meta: Metadata = undefined;
    pvtx.getMetadata(&meta);

    try testing.expectEqual(Identifier.parabolic_vertex, meta.identifier);
    try testing.expectEqualStrings("pvtx", meta.mnemonic);
    try testing.expectEqual(@as(usize, 1), meta.outputs_len);
    try testing.expectEqual(@as(u8, 1), meta.outputs_buf[0].kind);
}

test "PVTX mnemonic with component" {
    var pvtx = ParabolicVertex.init(.{ .bar_component = .median });
    pvtx.fixSlices();

    var meta: Metadata = undefined;
    pvtx.getMetadata(&meta);

    try testing.expectEqualStrings("pvtx(hl/2)", meta.mnemonic);
}

test "PVTX priming" {
    var pvtx = ParabolicVertex.init(.{});

    try testing.expect(math.isNan(pvtx.update(1.0)));
    try testing.expect(!pvtx.isPrimed());
    try testing.expect(math.isNan(pvtx.update(2.0)));
    try testing.expect(!pvtx.isPrimed());
    // Three collinear points -> zero curvature -> NaN, but primed.
    try testing.expect(math.isNan(pvtx.update(3.0)));
    try testing.expect(pvtx.isPrimed());
}
