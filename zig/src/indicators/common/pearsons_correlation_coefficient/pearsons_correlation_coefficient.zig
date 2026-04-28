const std = @import("std");
const math = std.math;

const Bar = @import("bar").Bar;
const Quote = @import("quote").Quote;
const Trade = @import("trade").Trade;
const Scalar = @import("scalar").Scalar;
const bar_component = @import("bar_component");
const quote_component = @import("quote_component");
const trade_component = @import("trade_component");

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

/// Enumerates the outputs of the Pearson's Correlation Coefficient indicator.
pub const PearsonsCorrelationCoefficientOutput = enum(u8) {
    /// The scalar value of the Pearson's Correlation Coefficient.
    value = 1,
};

/// Parameters to create an instance of the indicator.
pub const PearsonsCorrelationCoefficientParams = struct {
    /// The length (number of time periods). Must be >= 1.
    length: usize,
    /// Bar component to extract. `null` means use default (Close).
    bar_component: ?bar_component.BarComponent = null,
    /// Quote component to extract. `null` means use default (Mid).
    quote_component: ?quote_component.QuoteComponent = null,
    /// Trade component to extract. `null` means use default (Price).
    trade_component: ?trade_component.TradeComponent = null,
};

/// Computes Pearson's Correlation Coefficient (r) over a rolling window.
///
/// Given two input series X and Y, it computes:
///   r = (n*sumXY - sumX*sumY) / sqrt((n*sumX2 - sumX^2) * (n*sumY2 - sumY^2))
///
/// The indicator is not primed during the first length-1 updates.
/// For single-input updates, both X and Y are set to the same value (always returns 0 when primed).
/// For bar updates, X = High and Y = Low.
pub const PearsonsCorrelationCoefficient = struct {
    line: LineIndicator,
    window_x: []f64,
    window_y: []f64,
    length: usize,
    count: usize,
    pos: usize,
    sum_x: f64,
    sum_y: f64,
    sum_x2: f64,
    sum_y2: f64,
    sum_xy: f64,
    primed: bool,
    allocator: std.mem.Allocator,
    mnemonic_buf: [64]u8,
    mnemonic_len: usize,
    description_buf: [128]u8,
    description_len: usize,

    pub fn init(allocator: std.mem.Allocator, params: PearsonsCorrelationCoefficientParams) !PearsonsCorrelationCoefficient {
        if (params.length < 1) {
            return error.InvalidLength;
        }

        const bc = params.bar_component orelse bar_component.default_bar_component;
        const qc = params.quote_component orelse quote_component.default_quote_component;
        const tc = params.trade_component orelse trade_component.default_trade_component;

        var triple_buf: [64]u8 = undefined;
        const triple = component_triple_mnemonic_mod.componentTripleMnemonic(&triple_buf, bc, qc, tc);

        var mnemonic_buf: [64]u8 = undefined;
        const mnemonic_slice = std.fmt.bufPrint(&mnemonic_buf, "correl({d}{s})", .{ params.length, triple }) catch
            return error.MnemonicTooLong;
        const mnemonic_len = mnemonic_slice.len;

        var description_buf: [128]u8 = undefined;
        const desc_slice = std.fmt.bufPrint(&description_buf, "Pearsons Correlation Coefficient {s}", .{mnemonic_slice}) catch
            return error.MnemonicTooLong;
        const description_len = desc_slice.len;

        const window_x = try allocator.alloc(f64, params.length);
        @memset(window_x, 0.0);
        const window_y = try allocator.alloc(f64, params.length);
        @memset(window_y, 0.0);

        return .{
            .line = LineIndicator.new(
                mnemonic_buf[0..mnemonic_len],
                description_buf[0..description_len],
                params.bar_component,
                params.quote_component,
                params.trade_component,
            ),
            .window_x = window_x,
            .window_y = window_y,
            .length = params.length,
            .count = 0,
            .pos = 0,
            .sum_x = 0.0,
            .sum_y = 0.0,
            .sum_x2 = 0.0,
            .sum_y2 = 0.0,
            .sum_xy = 0.0,
            .primed = false,
            .allocator = allocator,
            .mnemonic_buf = mnemonic_buf,
            .mnemonic_len = mnemonic_len,
            .description_buf = description_buf,
            .description_len = description_len,
        };
    }

    pub fn deinit(self: *PearsonsCorrelationCoefficient) void {
        self.allocator.free(self.window_x);
        self.allocator.free(self.window_y);
    }

    pub fn fixSlices(self: *PearsonsCorrelationCoefficient) void {
        self.line.mnemonic = self.mnemonic_buf[0..self.mnemonic_len];
        self.line.description = self.description_buf[0..self.description_len];
    }

    /// Core update with a single scalar. Calls updatePair(sample, sample).
    pub fn update(self: *PearsonsCorrelationCoefficient, sample: f64) f64 {
        return self.updatePair(sample, sample);
    }

    /// Core update with an (x, y) pair.
    pub fn updatePair(self: *PearsonsCorrelationCoefficient, x: f64, y: f64) f64 {
        if (math.isNan(x) or math.isNan(y)) {
            return math.nan(f64);
        }

        const n: f64 = @floatFromInt(self.length);

        if (self.primed) {
            // Remove oldest values.
            const old_x = self.window_x[self.pos];
            const old_y = self.window_y[self.pos];

            self.sum_x -= old_x;
            self.sum_y -= old_y;
            self.sum_x2 -= old_x * old_x;
            self.sum_y2 -= old_y * old_y;
            self.sum_xy -= old_x * old_y;

            // Add new values.
            self.window_x[self.pos] = x;
            self.window_y[self.pos] = y;
            self.pos = (self.pos + 1) % self.length;

            self.sum_x += x;
            self.sum_y += y;
            self.sum_x2 += x * x;
            self.sum_y2 += y * y;
            self.sum_xy += x * y;

            return self.correlate(n);
        }

        // Accumulating phase.
        self.window_x[self.count] = x;
        self.window_y[self.count] = y;

        self.sum_x += x;
        self.sum_y += y;
        self.sum_x2 += x * x;
        self.sum_y2 += y * y;
        self.sum_xy += x * y;

        self.count += 1;

        if (self.count == self.length) {
            self.primed = true;
            self.pos = 0;
            return self.correlate(n);
        }

        return math.nan(f64);
    }

    /// Computes the Pearson correlation from the running sums.
    fn correlate(self: *const PearsonsCorrelationCoefficient, n: f64) f64 {
        const var_x = self.sum_x2 - (self.sum_x * self.sum_x) / n;
        const var_y = self.sum_y2 - (self.sum_y * self.sum_y) / n;
        const temp_real = var_x * var_y;

        if (temp_real <= 0) {
            return 0;
        }

        return (self.sum_xy - (self.sum_x * self.sum_y) / n) / @sqrt(temp_real);
    }

    pub fn isPrimed(self: *const PearsonsCorrelationCoefficient) bool {
        return self.primed;
    }

    pub fn getMetadata(self: *const PearsonsCorrelationCoefficient, out: *Metadata) void {
        build_metadata_mod.buildMetadata(
            out,
            .pearsons_correlation_coefficient,
            self.line.mnemonic,
            self.line.description,
            &[_]build_metadata_mod.OutputText{
                .{ .mnemonic = self.line.mnemonic, .description = self.line.description },
            },
        );
    }

    pub fn updateScalar(self: *PearsonsCorrelationCoefficient, sample: *const Scalar) OutputArray {
        const value = self.update(sample.value);
        return LineIndicator.wrapScalar(sample.time, value);
    }

    /// UpdateBar overrides to extract High (X) and Low (Y) from the bar.
    pub fn updateBar(self: *PearsonsCorrelationCoefficient, sample: *const Bar) OutputArray {
        const x = sample.high;
        const y = sample.low;
        const value = self.updatePair(x, y);
        return LineIndicator.wrapScalar(sample.time, value);
    }

    pub fn updateQuote(self: *PearsonsCorrelationCoefficient, sample: *const Quote) OutputArray {
        const value = self.update(self.line.extractQuote(sample));
        return LineIndicator.wrapScalar(sample.time, value);
    }

    pub fn updateTrade(self: *PearsonsCorrelationCoefficient, sample: *const Trade) OutputArray {
        const value = self.update(self.line.extractTrade(sample));
        return LineIndicator.wrapScalar(sample.time, value);
    }

    pub fn indicator(self: *PearsonsCorrelationCoefficient) indicator_mod.Indicator {
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
        const self: *PearsonsCorrelationCoefficient = @ptrCast(@alignCast(ptr));
        return self.isPrimed();
    }

    fn vtableMetadata(ptr: *anyopaque, out: *Metadata) void {
        const self: *const PearsonsCorrelationCoefficient = @ptrCast(@alignCast(ptr));
        self.getMetadata(out);
    }

    fn vtableUpdateScalar(ptr: *anyopaque, sample: *const Scalar) OutputArray {
        const self: *PearsonsCorrelationCoefficient = @ptrCast(@alignCast(ptr));
        return self.updateScalar(sample);
    }

    fn vtableUpdateBar(ptr: *anyopaque, sample: *const Bar) OutputArray {
        const self: *PearsonsCorrelationCoefficient = @ptrCast(@alignCast(ptr));
        return self.updateBar(sample);
    }

    fn vtableUpdateQuote(ptr: *anyopaque, sample: *const Quote) OutputArray {
        const self: *PearsonsCorrelationCoefficient = @ptrCast(@alignCast(ptr));
        return self.updateQuote(sample);
    }

    fn vtableUpdateTrade(ptr: *anyopaque, sample: *const Trade) OutputArray {
        const self: *PearsonsCorrelationCoefficient = @ptrCast(@alignCast(ptr));
        return self.updateTrade(sample);
    }

    pub const InitError = error{
        InvalidLength,
        MnemonicTooLong,
        OutOfMemory,
    };
};

// ---------------------------------------------------------------------------
// Tests
// ---------------------------------------------------------------------------

const testing = std.testing;

fn createCorrel(allocator: std.mem.Allocator, length: usize) !PearsonsCorrelationCoefficient {
    var c = try PearsonsCorrelationCoefficient.init(allocator, .{ .length = length });
    c.fixSlices();
    return c;
}

fn almostEqual(a: f64, b: f64, eps: f64) bool {
    return @abs(a - b) < eps;
}

fn testHighInput() [252]f64 {
    return .{
        93.250000,  94.940000,  96.375000,  96.190000,  96.000000,  94.720000,  95.000000,  93.720000,  92.470000,  92.750000,
        96.250000,  99.625000,  99.125000,  92.750000,  91.315000,  93.250000,  93.405000,  90.655000,  91.970000,  92.250000,
        90.345000,  88.500000,  88.250000,  85.500000,  84.440000,  84.750000,  84.440000,  89.405000,  88.125000,  89.125000,
        87.155000,  87.250000,  87.375000,  88.970000,  90.000000,  89.845000,  86.970000,  85.940000,  84.750000,  85.470000,
        84.470000,  88.500000,  89.470000,  90.000000,  92.440000,  91.440000,  92.970000,  91.720000,  91.155000,  91.750000,
        90.000000,  88.875000,  89.000000,  85.250000,  83.815000,  85.250000,  86.625000,  87.940000,  89.375000,  90.625000,
        90.750000,  88.845000,  91.970000,  93.375000,  93.815000,  94.030000,  94.030000,  91.815000,  92.000000,  91.940000,
        89.750000,  88.750000,  86.155000,  84.875000,  85.940000,  99.375000,  103.280000, 105.375000, 107.625000, 105.250000,
        104.500000, 105.500000, 106.125000, 107.940000, 106.250000, 107.000000, 108.750000, 110.940000, 110.940000, 114.220000,
        123.000000, 121.750000, 119.815000, 120.315000, 119.375000, 118.190000, 116.690000, 115.345000, 113.000000, 118.315000,
        116.870000, 116.750000, 113.870000, 114.620000, 115.310000, 116.000000, 121.690000, 119.870000, 120.870000, 116.750000,
        116.500000, 116.000000, 118.310000, 121.500000, 122.000000, 121.440000, 125.750000, 127.750000, 124.190000, 124.440000,
        125.750000, 124.690000, 125.310000, 132.000000, 131.310000, 132.250000, 133.880000, 133.500000, 135.500000, 137.440000,
        138.690000, 139.190000, 138.500000, 138.130000, 137.500000, 138.880000, 132.130000, 129.750000, 128.500000, 125.440000,
        125.120000, 126.500000, 128.690000, 126.620000, 126.690000, 126.000000, 123.120000, 121.870000, 124.000000, 127.000000,
        124.440000, 122.500000, 123.750000, 123.810000, 124.500000, 127.870000, 128.560000, 129.630000, 124.870000, 124.370000,
        124.870000, 123.620000, 124.060000, 125.870000, 125.190000, 125.620000, 126.000000, 128.500000, 126.750000, 129.750000,
        132.690000, 133.940000, 136.500000, 137.690000, 135.560000, 133.560000, 135.000000, 132.380000, 131.440000, 130.880000,
        129.630000, 127.250000, 127.810000, 125.000000, 126.810000, 124.750000, 122.810000, 122.250000, 121.060000, 120.000000,
        123.250000, 122.750000, 119.190000, 115.060000, 116.690000, 114.870000, 110.870000, 107.250000, 108.870000, 109.000000,
        108.500000, 113.060000, 93.000000,  94.620000,  95.120000,  96.000000,  95.560000,  95.310000,  99.000000,  98.810000,
        96.810000,  95.940000,  94.440000,  92.940000,  93.940000,  95.500000,  97.060000,  97.500000,  96.250000,  96.370000,
        95.000000,  94.870000,  98.250000,  105.120000, 108.440000, 109.870000, 105.000000, 106.000000, 104.940000, 104.500000,
        104.440000, 106.310000, 112.870000, 116.500000, 119.190000, 121.000000, 122.120000, 111.940000, 112.750000, 110.190000,
        107.940000, 109.690000, 111.060000, 110.440000, 110.120000, 110.310000, 110.440000, 110.000000, 110.750000, 110.500000,
        110.500000, 109.500000,
    };
}

fn testLowInput() [252]f64 {
    return .{
        90.750000,  91.405000,  94.250000,  93.500000,  92.815000,  93.500000,  92.000000,  89.750000,  89.440000,  90.625000,
        92.750000,  96.315000,  96.030000,  88.815000,  86.750000,  90.940000,  88.905000,  88.780000,  89.250000,  89.750000,
        87.500000,  86.530000,  84.625000,  82.280000,  81.565000,  80.875000,  81.250000,  84.065000,  85.595000,  85.970000,
        84.405000,  85.095000,  85.500000,  85.530000,  87.875000,  86.565000,  84.655000,  83.250000,  82.565000,  83.440000,
        82.530000,  85.065000,  86.875000,  88.530000,  89.280000,  90.125000,  90.750000,  89.000000,  88.565000,  90.095000,
        89.000000,  86.470000,  84.000000,  83.315000,  82.000000,  83.250000,  84.750000,  85.280000,  87.190000,  88.440000,
        88.250000,  87.345000,  89.280000,  91.095000,  89.530000,  91.155000,  92.000000,  90.530000,  89.970000,  88.815000,
        86.750000,  85.065000,  82.030000,  81.500000,  82.565000,  96.345000,  96.470000,  101.155000, 104.250000, 101.750000,
        101.720000, 101.720000, 103.155000, 105.690000, 103.655000, 104.000000, 105.530000, 108.530000, 108.750000, 107.750000,
        117.000000, 118.000000, 116.000000, 118.500000, 116.530000, 116.250000, 114.595000, 110.875000, 110.500000, 110.720000,
        112.620000, 114.190000, 111.190000, 109.440000, 111.560000, 112.440000, 117.500000, 116.060000, 116.560000, 113.310000,
        112.560000, 114.000000, 114.750000, 118.870000, 119.000000, 119.750000, 122.620000, 123.000000, 121.750000, 121.560000,
        123.120000, 122.190000, 122.750000, 124.370000, 128.000000, 129.500000, 130.810000, 130.630000, 132.130000, 133.880000,
        135.380000, 135.750000, 136.190000, 134.500000, 135.380000, 133.690000, 126.060000, 126.870000, 123.500000, 122.620000,
        122.750000, 123.560000, 125.810000, 124.620000, 124.370000, 121.810000, 118.190000, 118.060000, 117.560000, 121.000000,
        121.120000, 118.940000, 119.810000, 121.000000, 122.000000, 124.500000, 126.560000, 123.500000, 121.250000, 121.060000,
        122.310000, 121.000000, 120.870000, 122.060000, 122.750000, 122.690000, 122.870000, 125.500000, 124.250000, 128.000000,
        128.380000, 130.690000, 131.630000, 134.380000, 132.000000, 131.940000, 131.940000, 129.560000, 123.750000, 126.000000,
        126.250000, 124.370000, 121.440000, 120.440000, 121.370000, 121.690000, 120.000000, 119.620000, 115.500000, 116.750000,
        119.060000, 119.060000, 115.060000, 111.060000, 113.120000, 110.000000, 105.000000, 104.690000, 103.870000, 104.690000,
        105.440000, 107.000000, 89.000000,  92.500000,  92.120000,  94.620000,  92.810000,  94.250000,  96.250000,  96.370000,
        93.690000,  93.500000,  90.000000,  90.190000,  90.500000,  92.120000,  94.120000,  94.870000,  93.000000,  93.870000,
        93.000000,  92.620000,  93.560000,  98.370000,  104.440000, 106.000000, 101.810000, 104.120000, 103.370000, 102.120000,
        102.250000, 103.370000, 107.940000, 112.500000, 115.440000, 115.500000, 112.250000, 107.560000, 106.560000, 106.870000,
        104.500000, 105.750000, 108.620000, 107.750000, 108.060000, 108.000000, 108.190000, 108.120000, 109.060000, 108.750000,
        108.560000, 106.620000,
    };
}

fn testExcelExpected() [252]f64 {
    return .{
        math.nan(f64),      math.nan(f64),      math.nan(f64),      math.nan(f64),
        math.nan(f64),      math.nan(f64),      math.nan(f64),      math.nan(f64),
        math.nan(f64),      math.nan(f64),      math.nan(f64),      math.nan(f64),
        math.nan(f64),      math.nan(f64),      math.nan(f64),      math.nan(f64),
        math.nan(f64),      math.nan(f64),      math.nan(f64),      0.9401568590471170,
        0.9471811552404370, 0.9526486511378510, 0.9594395234433260, 0.9684901755431890,
        0.9745456460684540, 0.9823609526746270, 0.9842318414143900, 0.9793319838629490,
        0.9788482705359360, 0.9800972276693750, 0.9785845616077360, 0.9693167394743250,
        0.9470228533431190, 0.9453838549547040, 0.9517592162172850, 0.9433392348270420,
        0.9478926222765860, 0.9475923828106470, 0.9381448555694890, 0.9153603780858040,
        0.9037133253874650, 0.9044272330533300, 0.9144567231385610, 0.9140422416270490,
        0.9240907021037240, 0.9245714706598480, 0.9301864925512240, 0.9685637131517730,
        0.9699713144403260, 0.9722430818693880, 0.9659537966306980, 0.9653083079242570,
        0.9421810188742970, 0.9488679626820190, 0.9540546636443100, 0.9585657829863180,
        0.9580677494166530, 0.9563189205945780, 0.9522216783457150, 0.9499605370495780,
        0.9433582538169970, 0.9448506810540480, 0.9467581283005850, 0.9540541884980060,
        0.9451938775561860, 0.9527747257400080, 0.9545845363155470, 0.9518140900144220,
        0.9526969023077840, 0.9521438024060810, 0.9571587276746930, 0.9534806508367840,
        0.9664678685848590, 0.9676372885038130, 0.9673951425348660, 0.9754944156042740,
        0.9618023912755630, 0.9735983892487120, 0.9814507316373980, 0.9846276725648570,
        0.9860054566213870, 0.9882105937528850, 0.9891883760796960, 0.9901171811095450,
        0.9908918742809730, 0.9913801756613770, 0.9923983921194380, 0.9940283028145840,
        0.9946689515269380, 0.9917313455559860, 0.9912246592297800, 0.9914766005027000,
        0.9905843446629170, 0.9867768283384500, 0.9800208355632590, 0.9775868446129000,
        0.9818158252339090, 0.9805269039354710, 0.9794926025278210, 0.9651664579178690,
        0.9607101819596180, 0.9548369427331770, 0.9472371147678250, 0.9393874519173430,
        0.9262650699399830, 0.9055359430253950, 0.8917513099360840, 0.8845363842595150,
        0.8780023472740510, 0.8768594795208420, 0.8847330319751160, 0.8588444464269830,
        0.8540394922856320, 0.8685071144796010, 0.8867017785665350, 0.9019536780962040,
        0.9351506214260950, 0.9460614408106260, 0.9487795387829030, 0.9755185789962120,
        0.9781420983101290, 0.9796746522636920, 0.9800485956962120, 0.9592735362917000,
        0.9629702197175760, 0.9653515565441640, 0.9722435637049560, 0.9751999286917070,
        0.9793400350341120, 0.9807373681199240, 0.9820819629564370, 0.9819960901599760,
        0.9804984607791450, 0.9802670958185390, 0.9782011441033960, 0.9752990704028820,
        0.9675865981117540, 0.9675082971922110, 0.9636801404836100, 0.9624399587196960,
        0.9623844816290560, 0.9615007729973930, 0.9591567353952830, 0.9776738747439900,
        0.9783423006715560, 0.9795238639226300, 0.9805925440791580, 0.9832188414769910,
        0.9796598649299700, 0.9758993430891440, 0.9735179726086800, 0.9704997683460010,
        0.9656284571769230, 0.9553451160914730, 0.9361819221968540, 0.8782125747224650,
        0.8982719217727120, 0.8477723464095960, 0.8546377589244040, 0.8572871327351020,
        0.8585528871548040, 0.8526022957921970, 0.8292629229925110, 0.8294469606129980,
        0.8190212269209530, 0.8191577970270310, 0.8132742438973340, 0.8118086128565220,
        0.8451960695090540, 0.9147140984114960, 0.9323286922042040, 0.9488173569229080,
        0.9616254945974570, 0.9714396824012870, 0.9749722214786510, 0.9725790982765770,
        0.9766180777342920, 0.9873289956355020, 0.9621423713797460, 0.9561111788043280,
        0.9538925936199330, 0.9496547804857550, 0.9329386834398710, 0.9349122391333710,
        0.9350195404239680, 0.9370998633521060, 0.9414383095728230, 0.9472807230887830,
        0.9572389026417090, 0.9658768229512560, 0.9674312191145570, 0.9673978992813970,
        0.9691937580390360, 0.9698138706945130, 0.9685257570998960, 0.9728666378692480,
        0.9740699834558090, 0.9773946409583190, 0.9859156575982320, 0.9868010214710570,
        0.9860650762821040, 0.9840721831018110, 0.9917556231372550, 0.9922091139248850,
        0.9933724117160440, 0.9926935073704690, 0.9934630349771790, 0.9934195189537240,
        0.9935298892857930, 0.9935049843553920, 0.9928126293583000, 0.9921799215414420,
        0.9899983601098890, 0.9895275103398580, 0.9882614534065030, 0.9855713950443270,
        0.9854298482919770, 0.9852191036619460, 0.9834218730447970, 0.9788631710872740,
        0.9704655850058130, 0.9194300540508450, 0.8681407925170670, 0.8808536680642600,
        0.9440776152383960, 0.9689567658743020, 0.9716435818001360, 0.9756305746107580,
        0.9745748112805210, 0.9757768027062840, 0.9757402824215750, 0.9771541892863710,
        0.9795131678773640, 0.9827854278332160, 0.9859330369551790, 0.9874122533690860,
        0.9784753788149660, 0.9776244828040590, 0.9745149943245450, 0.9710356842260290,
        0.9651684328545020, 0.9548286702701730, 0.9440773679358460, 0.9574334358838210,
        0.9540249522022400, 0.9517837091130880, 0.9466458287565880, 0.9408225232591590,
        0.9304922072407360, 0.9156400478034220, 0.8963662049425160, 0.8866901149929160,
    };
}

test "pearsons correlation coefficient talib spot checks period=20" {
    const high = testHighInput();
    const low = testLowInput();

    var c = try createCorrel(testing.allocator, 20);
    defer c.deinit();

    // First 18 values should be NaN (lookback = 19, primed at index 19).
    for (0..19) |i| {
        const v = c.updatePair(high[i], low[i]);
        if (i < 19) {
            try testing.expect(math.isNan(v));
        }
    }

    // Feed remaining.
    for (19..252) |i| {
        const act = c.updatePair(high[i], low[i]);
        switch (i) {
            19 => try testing.expect(almostEqual(act, 0.9401569, 1e-4)),
            20 => try testing.expect(almostEqual(act, 0.9471812, 1e-4)),
            251 => try testing.expect(almostEqual(act, 0.8866901, 1e-4)),
            else => {},
        }
    }

    // NaN passthrough.
    try testing.expect(math.isNan(c.updatePair(math.nan(f64), 1.0)));
    try testing.expect(math.isNan(c.updatePair(1.0, math.nan(f64))));
}

test "pearsons correlation coefficient excel verification period=20" {
    const high = testHighInput();
    const low = testLowInput();
    const expected = testExcelExpected();
    const eps: f64 = 1e-10;

    var c = try createCorrel(testing.allocator, 20);
    defer c.deinit();

    for (0..252) |i| {
        const act = c.updatePair(high[i], low[i]);
        if (i >= 19) {
            if (math.isNan(expected[i])) {
                try testing.expect(math.isNan(act));
            } else {
                try testing.expect(almostEqual(act, expected[i], eps));
            }
        } else {
            try testing.expect(math.isNan(act));
        }
    }
}

test "pearsons correlation coefficient is primed length=1" {
    const high = testHighInput();
    const low = testLowInput();

    var c = try createCorrel(testing.allocator, 1);
    defer c.deinit();

    try testing.expect(!c.isPrimed());
    _ = c.updatePair(high[0], low[0]);
    try testing.expect(c.isPrimed());
}

test "pearsons correlation coefficient is primed length=2" {
    const high = testHighInput();
    const low = testLowInput();

    var c = try createCorrel(testing.allocator, 2);
    defer c.deinit();

    try testing.expect(!c.isPrimed());
    _ = c.updatePair(high[0], low[0]);
    try testing.expect(!c.isPrimed());
    _ = c.updatePair(high[1], low[1]);
    try testing.expect(c.isPrimed());
}

test "pearsons correlation coefficient is primed length=20" {
    const high = testHighInput();
    const low = testLowInput();

    var c = try createCorrel(testing.allocator, 20);
    defer c.deinit();

    try testing.expect(!c.isPrimed());
    for (0..19) |i| {
        _ = c.updatePair(high[i], low[i]);
        try testing.expect(!c.isPrimed());
    }
    _ = c.updatePair(high[19], low[19]);
    try testing.expect(c.isPrimed());
}

test "pearsons correlation coefficient metadata" {
    var c = try createCorrel(testing.allocator, 20);
    defer c.deinit();
    var m: Metadata = undefined;
    c.getMetadata(&m);

    try testing.expectEqual(Identifier.pearsons_correlation_coefficient, m.identifier);
    try testing.expectEqualStrings("correl(20)", m.mnemonic);
    try testing.expectEqualStrings("Pearsons Correlation Coefficient correl(20)", m.description);
    try testing.expectEqual(@as(usize, 1), m.outputs_len);
    try testing.expectEqualStrings("correl(20)", m.outputs_buf[0].mnemonic);
}

test "pearsons correlation coefficient init invalid length" {
    const r = PearsonsCorrelationCoefficient.init(testing.allocator, .{ .length = 0 });
    try testing.expectError(error.InvalidLength, r);
}

test "pearsons correlation coefficient mnemonic with components" {
    var c = try PearsonsCorrelationCoefficient.init(testing.allocator, .{
        .length = 20,
        .bar_component = .median,
    });
    defer c.deinit();
    c.fixSlices();
    try testing.expectEqualStrings("correl(20, hl/2)", c.line.mnemonic);
}

test "pearsons correlation coefficient update entity bar" {
    var c = try createCorrel(testing.allocator, 2);
    defer c.deinit();

    // Feed one pair via updatePair, then one bar.
    _ = c.updatePair(10, 5);
    _ = c.updatePair(20, 10);
    const bar = Bar{ .time = 1617235200, .open = 0, .high = 30, .low = 15, .close = 0, .volume = 0 };
    const out = c.updateBar(&bar);
    try testing.expectEqual(@as(usize, 1), out.len);
    const s = out.slice()[0].scalar;
    try testing.expectEqual(@as(i64, 1617235200), s.time);
    try testing.expect(!math.isNan(s.value));
}

test "pearsons correlation coefficient update entity scalar" {
    const inp: f64 = 3.0;
    const time: i64 = 1617235200;

    var c = try createCorrel(testing.allocator, 2);
    defer c.deinit();

    _ = c.update(inp);
    _ = c.update(inp);
    const out = c.updateScalar(&.{ .time = time, .value = inp });
    try testing.expectEqual(@as(usize, 1), out.len);
    const s = out.slice()[0].scalar;
    try testing.expectEqual(time, s.time);
    // correl(x,x) with constant value returns 0.
    try testing.expect(almostEqual(s.value, 0.0, 1e-10));
}

test "pearsons correlation coefficient update entity quote" {
    const inp: f64 = 3.0;
    const time: i64 = 1617235200;

    var c = try createCorrel(testing.allocator, 2);
    defer c.deinit();

    _ = c.update(inp);
    _ = c.update(inp);
    const q = Quote{ .time = time, .bid_price = inp, .ask_price = inp, .bid_size = 1, .ask_size = 1 };
    const out = c.updateQuote(&q);
    const s = out.slice()[0].scalar;
    try testing.expect(almostEqual(s.value, 0.0, 1e-10));
}

test "pearsons correlation coefficient update entity trade" {
    const inp: f64 = 3.0;
    const time: i64 = 1617235200;

    var c = try createCorrel(testing.allocator, 2);
    defer c.deinit();

    _ = c.update(inp);
    _ = c.update(inp);
    const t = Trade{ .time = time, .price = inp, .volume = 1 };
    const out = c.updateTrade(&t);
    const s = out.slice()[0].scalar;
    try testing.expect(almostEqual(s.value, 0.0, 1e-10));
}
