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

/// Enumerates the outputs of the Cubic Vertex indicator.
pub const CubicVertexOutput = enum(u8) {
    /// The number of bars to the more imminent turning point.
    bars_to_near_turn = 1,
    /// The number of bars to the more distant turning point.
    bars_to_far_turn = 2,
};

/// Parameters to create a Cubic Vertex indicator.
///
/// The indicator has no numeric parameters; only the input price component is configurable.
pub const CubicVertexParams = struct {
    bar_component: ?bar_component.BarComponent = null,
    quote_component: ?quote_component.QuoteComponent = null,
    trade_component: ?trade_component.TradeComponent = null,
};

/// Cubic Vertex (CVTX) by Don Mak.
///
/// Predicts turning points by fitting a cubic polynomial to the 4 most recent price
/// points and computing where the two vertices (extrema) occur relative to the current
/// bar. Given four consecutive prices x(n), x(n-1), x(n-2), x(n-3) (most recent first),
/// the cubic coefficients are (Eq 7.2a-c):
///   c = (x(n) - 3*x(n-1) + 3*x(n-2) - x(n-3)) / 6
///   d = (2*x(n) - 5*x(n-1) + 4*x(n-2) - x(n-3)) / 2
///   e = (11*x(n) - 18*x(n-1) + 9*x(n-2) - 2*x(n-3)) / 6
/// The vertex locations are the roots of 3c*t^2 + 2d*t + e = 0. near = smaller |t|,
/// far = larger |t|.
pub const CubicVertex = struct {
    bar_func: bar_component.BarFunc,
    quote_func: quote_component.QuoteFunc,
    trade_func: trade_component.TradeFunc,

    buffer: [4]f64 = .{ 0.0, 0.0, 0.0, 0.0 },
    index: usize = 0,
    count: usize = 0,

    primed: bool = false,

    mnemonic_buf: [64]u8,
    mnemonic_len: usize,
    description_buf: [128]u8,
    description_len: usize,

    pub fn init(params: CubicVertexParams) CubicVertex {
        const bc = params.bar_component orelse bar_component.default_bar_component;
        const qc = params.quote_component orelse quote_component.default_quote_component;
        const tc = params.trade_component orelse trade_component.default_trade_component;

        var triple_buf: [64]u8 = undefined;
        const triple = component_triple_mnemonic_mod.componentTripleMnemonic(&triple_buf, bc, qc, tc);

        var mnemonic_buf: [64]u8 = undefined;
        var mnemonic_len: usize = undefined;
        if (triple.len == 0) {
            const s = std.fmt.bufPrint(&mnemonic_buf, "cvtx", .{}) catch unreachable;
            mnemonic_len = s.len;
        } else {
            // Strip leading ", " from triple.
            const suffix = if (triple.len > 2) triple[2..] else triple;
            const s = std.fmt.bufPrint(&mnemonic_buf, "cvtx({s})", .{suffix}) catch unreachable;
            mnemonic_len = s.len;
        }

        var description_buf: [128]u8 = undefined;
        const desc_slice = std.fmt.bufPrint(&description_buf, "Cubic vertex {s}", .{mnemonic_buf[0..mnemonic_len]}) catch unreachable;
        const description_len = desc_slice.len;

        return .{
            .bar_func = bar_component.componentValue(bc),
            .quote_func = quote_component.componentValue(qc),
            .trade_func = trade_component.componentValue(tc),
            .mnemonic_buf = mnemonic_buf,
            .mnemonic_len = mnemonic_len,
            .description_buf = description_buf,
            .description_len = description_len,
        };
    }

    pub fn fixSlices(self: *CubicVertex) void {
        _ = self;
        // CVTX reads mnemonic/description from its buffers directly; no slice fixup needed.
    }

    /// Returns (bars_to_near_turn, bars_to_far_turn).
    pub fn updateValues(self: *CubicVertex, sample: f64) struct { near: f64, far: f64 } {
        const nan = math.nan(f64);

        // Store the price in the ring buffer.
        self.buffer[self.index] = sample;
        self.index = (self.index + 1) % 4;
        self.count += 1;

        if (self.count < 4) {
            self.primed = false;
            return .{ .near = nan, .far = nan };
        }

        self.primed = true;

        // Extract prices: x[n] (newest), x[n-1], x[n-2], x[n-3] (oldest).
        const idx: i64 = @intCast(self.index);
        const xn = self.buffer[@intCast(@mod(idx - 1, 4))];
        const xn1 = self.buffer[@intCast(@mod(idx - 2, 4))];
        const xn2 = self.buffer[@intCast(@mod(idx - 3, 4))];
        const xn3 = self.buffer[@intCast(@mod(idx - 4, 4))];

        // Cubic polynomial coefficients (Eq 7.2a-c).
        const c = (xn - 3.0 * xn1 + 3.0 * xn2 - xn3) / 6.0;
        const d = (2.0 * xn - 5.0 * xn1 + 4.0 * xn2 - xn3) / 2.0;
        const e = (11.0 * xn - 18.0 * xn1 + 9.0 * xn2 - 2.0 * xn3) / 6.0;

        // Case: c == 0 -- cubic term vanishes, reduces to parabola or line.
        if (c == 0.0) {
            if (d == 0.0) {
                return .{ .near = nan, .far = nan };
            }
            const vertex = -e / (2.0 * d);
            return .{ .near = vertex, .far = nan };
        }

        // Full cubic: solve quadratic 3c*t^2 + 2d*t + e = 0.
        const disc = d * d - 3.0 * c * e;

        if (disc < 0.0) {
            return .{ .near = nan, .far = nan };
        }

        if (disc == 0.0) {
            const vertex = -d / (3.0 * c);
            return .{ .near = vertex, .far = vertex };
        }

        const sqrt_disc = @sqrt(disc);
        const three_c = 3.0 * c;

        const t_plus = (-d + sqrt_disc) / three_c;
        const t_minus = (-d - sqrt_disc) / three_c;

        if (@abs(t_plus) <= @abs(t_minus)) {
            return .{ .near = t_plus, .far = t_minus };
        }
        return .{ .near = t_minus, .far = t_plus };
    }

    pub fn isPrimed(self: *const CubicVertex) bool {
        return self.primed;
    }

    fn mnemonic(self: *const CubicVertex) []const u8 {
        return self.mnemonic_buf[0..self.mnemonic_len];
    }

    fn description(self: *const CubicVertex) []const u8 {
        return self.description_buf[0..self.description_len];
    }

    pub fn getMetadata(self: *const CubicVertex, out: *Metadata) void {
        const mn = self.mnemonic();
        const desc = self.description();

        var near_mn_buf: [96]u8 = undefined;
        const near_mn = std.fmt.bufPrint(&near_mn_buf, "{s} near", .{mn}) catch mn;
        var far_mn_buf: [96]u8 = undefined;
        const far_mn = std.fmt.bufPrint(&far_mn_buf, "{s} far", .{mn}) catch mn;

        var near_desc_buf: [160]u8 = undefined;
        const near_desc = std.fmt.bufPrint(&near_desc_buf, "{s} near turn", .{desc}) catch desc;
        var far_desc_buf: [160]u8 = undefined;
        const far_desc = std.fmt.bufPrint(&far_desc_buf, "{s} far turn", .{desc}) catch desc;

        build_metadata_mod.buildMetadata(
            out,
            .cubic_vertex,
            mn,
            desc,
            &[_]build_metadata_mod.OutputText{
                .{ .mnemonic = near_mn, .description = near_desc },
                .{ .mnemonic = far_mn, .description = far_desc },
            },
        );
    }

    fn makeOutput(time: i64, near_v: f64, far_v: f64) OutputArray {
        var out = OutputArray{};
        out.append(.{ .scalar = .{ .time = time, .value = near_v } });
        out.append(.{ .scalar = .{ .time = time, .value = far_v } });
        return out;
    }

    pub fn updateScalar(self: *CubicVertex, sample: *const Scalar) OutputArray {
        const result = self.updateValues(sample.value);
        return makeOutput(sample.time, result.near, result.far);
    }

    pub fn updateBar(self: *CubicVertex, sample: *const Bar) OutputArray {
        const v = self.bar_func(sample.*);
        return self.updateScalar(&Scalar{ .time = sample.time, .value = v });
    }

    pub fn updateQuote(self: *CubicVertex, sample: *const Quote) OutputArray {
        const v = self.quote_func(sample.*);
        return self.updateScalar(&Scalar{ .time = sample.time, .value = v });
    }

    pub fn updateTrade(self: *CubicVertex, sample: *const Trade) OutputArray {
        const v = self.trade_func(sample.*);
        return self.updateScalar(&Scalar{ .time = sample.time, .value = v });
    }

    pub fn indicator(self: *CubicVertex) indicator_mod.Indicator {
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
        const self: *CubicVertex = @ptrCast(@alignCast(ptr));
        return self.isPrimed();
    }
    fn vtableMetadata(ptr: *anyopaque, out: *Metadata) void {
        const self: *const CubicVertex = @ptrCast(@alignCast(ptr));
        self.getMetadata(out);
    }
    fn vtableUpdateScalar(ptr: *anyopaque, sample: *const Scalar) OutputArray {
        const self: *CubicVertex = @ptrCast(@alignCast(ptr));
        return self.updateScalar(sample);
    }
    fn vtableUpdateBar(ptr: *anyopaque, sample: *const Bar) OutputArray {
        const self: *CubicVertex = @ptrCast(@alignCast(ptr));
        return self.updateBar(sample);
    }
    fn vtableUpdateQuote(ptr: *anyopaque, sample: *const Quote) OutputArray {
        const self: *CubicVertex = @ptrCast(@alignCast(ptr));
        return self.updateQuote(sample);
    }
    fn vtableUpdateTrade(ptr: *anyopaque, sample: *const Trade) OutputArray {
        const self: *CubicVertex = @ptrCast(@alignCast(ptr));
        return self.updateTrade(sample);
    }
};

// ---------------------------------------------------------------------------
// Tests
// ---------------------------------------------------------------------------

const testing = std.testing;
const testdata = @import("testdata.zig");

const tolerance = 1e-9;

fn runSeries(inputs: []const f64, near: []f64, far: []f64) void {
    var cvtx = CubicVertex.init(.{});
    for (0..inputs.len) |i| {
        const result = cvtx.updateValues(inputs[i]);
        near[i] = result.near;
        far[i] = result.far;
    }
}

fn checkSeries(actual: []const f64, expected: []const f64) !void {
    try testing.expectEqual(actual.len, expected.len);
    for (0..expected.len) |i| {
        const exp = expected[i];
        if (math.isNan(exp)) {
            try testing.expect(math.isNan(actual[i]));
        } else {
            // Combined absolute + relative tolerance (ill-conditioned near degenerate points).
            const delta = tolerance * @max(1.0, @abs(exp));
            try testing.expect(@abs(actual[i] - exp) <= delta);
        }
    }
}

test "CVTX reference data" {
    var near: [252]f64 = undefined;
    var far: [252]f64 = undefined;

    const input_close = testdata.iNPUT_CLOSE();
    runSeries(&input_close, &near, &far);
    try checkSeries(&near, &testdata.expectedRAW_NEAR());
    try checkSeries(&far, &testdata.expectedRAW_FAR());

    const input_ema6 = testdata.iNPUT_EMA6();
    runSeries(&input_ema6, &near, &far);
    try checkSeries(&near, &testdata.expectedEMA6_NEAR());
    try checkSeries(&far, &testdata.expectedEMA6_FAR());

    const input_ema20 = testdata.iNPUT_EMA20();
    runSeries(&input_ema20, &near, &far);
    try checkSeries(&near, &testdata.expectedEMA20_NEAR());
    try checkSeries(&far, &testdata.expectedEMA20_FAR());
}

test "CVTX TEST1 cubic" {
    var near: [60]f64 = undefined;
    var far: [60]f64 = undefined;

    const input = testdata.tEST1_INPUT_CUBIC();
    runSeries(&input, &near, &far);
    try checkSeries(&near, &testdata.tEST1_EXPECTED_NEAR());
    try checkSeries(&far, &testdata.tEST1_EXPECTED_FAR());
}

test "CVTX metadata default" {
    var cvtx = CubicVertex.init(.{});
    cvtx.fixSlices();

    var meta: Metadata = undefined;
    cvtx.getMetadata(&meta);

    try testing.expectEqual(Identifier.cubic_vertex, meta.identifier);
    try testing.expectEqualStrings("cvtx", meta.mnemonic);
    try testing.expectEqual(@as(usize, 2), meta.outputs_len);
    try testing.expectEqual(@as(u8, 1), meta.outputs_buf[0].kind);
    try testing.expectEqual(@as(u8, 2), meta.outputs_buf[1].kind);
}

test "CVTX mnemonic with component" {
    var cvtx = CubicVertex.init(.{ .bar_component = .median });
    cvtx.fixSlices();

    var meta: Metadata = undefined;
    cvtx.getMetadata(&meta);

    try testing.expectEqualStrings("cvtx(hl/2)", meta.mnemonic);
}

test "CVTX priming" {
    var cvtx = CubicVertex.init(.{});

    for (0..3) |_| {
        const result = cvtx.updateValues(1.0);
        try testing.expect(math.isNan(result.near) and math.isNan(result.far));
        try testing.expect(!cvtx.isPrimed());
    }
    // Four collinear points -> c == 0 and d == 0 -> both NaN, but primed.
    const result = cvtx.updateValues(1.0);
    try testing.expect(math.isNan(result.near) and math.isNan(result.far));
    try testing.expect(cvtx.isPrimed());
}
