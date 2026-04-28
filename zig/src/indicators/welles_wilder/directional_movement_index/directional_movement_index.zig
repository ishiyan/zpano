const std = @import("std");
const math = std.math;

const Bar = @import("bar").Bar;
const Quote = @import("quote").Quote;
const Trade = @import("trade").Trade;
const Scalar = @import("scalar").Scalar;

const indicator_mod = @import("../../core/indicator.zig");
const build_metadata_mod = @import("../../core/build_metadata.zig");
const identifier_mod = @import("../../core/identifier.zig");
const metadata_mod = @import("../../core/metadata.zig");

const dip_mod = @import("../directional_indicator_plus/directional_indicator_plus.zig");
const DirectionalIndicatorPlus = dip_mod.DirectionalIndicatorPlus;
const dim_mod = @import("../directional_indicator_minus/directional_indicator_minus.zig");
const DirectionalIndicatorMinus = dim_mod.DirectionalIndicatorMinus;

const OutputArray = indicator_mod.OutputArray;
const Identifier = identifier_mod.Identifier;
const Metadata = metadata_mod.Metadata;

const epsilon = 1e-8;

/// Enumerates the outputs of the Directional Movement Index indicator.
pub const DirectionalMovementIndexOutput = enum(u8) {
    /// The scalar value of the directional movement index (DX).
    value = 1,
    /// The scalar value of the directional indicator plus (+DI).
    directional_indicator_plus = 2,
    /// The scalar value of the directional indicator minus (-DI).
    directional_indicator_minus = 3,
    /// The scalar value of the directional movement plus (+DM).
    directional_movement_plus = 4,
    /// The scalar value of the directional movement minus (-DM).
    directional_movement_minus = 5,
    /// The scalar value of the average true range (ATR).
    average_true_range = 6,
    /// The scalar value of the true range (TR).
    true_range = 7,
};

/// Welles Wilder's Directional Movement Index (DX).
///
/// The directional movement index measures the strength of a trend by comparing
/// the positive and negative directional indicators. It is calculated as:
///   DX = 100 * |+DI - -DI| / (+DI + -DI)
///
/// where +DI is the directional indicator plus and -DI is the directional
/// indicator minus, both computed over the same length.
pub const DirectionalMovementIndex = struct {
    length: i32,
    value: f64,
    directional_indicator_plus: DirectionalIndicatorPlus,
    directional_indicator_minus: DirectionalIndicatorMinus,

    const mnemonic_str = "dx";
    const description_str = "Directional Movement Index";

    pub const Error = error{
        InvalidLength,
        OutOfMemory,
    };

    pub fn init(allocator: std.mem.Allocator, params: struct { length: i32 = 14 }) Error!DirectionalMovementIndex {
        if (params.length < 1) return Error.InvalidLength;

        const dip = DirectionalIndicatorPlus.init(allocator, .{ .length = params.length }) catch |e| switch (e) {
            error.InvalidLength => return Error.InvalidLength,
            error.OutOfMemory => return Error.OutOfMemory,
        };

        const dim = DirectionalIndicatorMinus.init(allocator, .{ .length = params.length }) catch |e| switch (e) {
            error.InvalidLength => return Error.InvalidLength,
            error.OutOfMemory => return Error.OutOfMemory,
        };

        return .{
            .length = params.length,
            .value = math.nan(f64),
            .directional_indicator_plus = dip,
            .directional_indicator_minus = dim,
        };
    }

    pub fn deinit(self: *DirectionalMovementIndex) void {
        self.directional_indicator_plus.deinit();
        self.directional_indicator_minus.deinit();
    }

    pub fn fixSlices(_: *DirectionalMovementIndex) void {}

    /// Update given close, high, low values.
    pub fn update(self: *DirectionalMovementIndex, close: f64, high: f64, low: f64) f64 {
        if (math.isNan(close) or math.isNan(high) or math.isNan(low)) return math.nan(f64);

        const dip_value = self.directional_indicator_plus.update(close, high, low);
        const dim_value = self.directional_indicator_minus.update(close, high, low);

        if (self.directional_indicator_plus.isPrimed() and self.directional_indicator_minus.isPrimed()) {
            const sum = dip_value + dim_value;

            if (@abs(sum) < epsilon) {
                self.value = 0;
            } else {
                self.value = 100.0 * @abs(dip_value - dim_value) / sum;
            }

            return self.value;
        }

        return math.nan(f64);
    }

    /// Update using a single sample value as substitute for high, low, close.
    pub fn updateSample(self: *DirectionalMovementIndex, sample: f64) f64 {
        return self.update(sample, sample, sample);
    }

    pub fn isPrimed(self: *const DirectionalMovementIndex) bool {
        return self.directional_indicator_plus.isPrimed() and self.directional_indicator_minus.isPrimed();
    }

    pub fn getMetadata(_: *const DirectionalMovementIndex, out: *Metadata) void {
        build_metadata_mod.buildMetadata(out, Identifier.directional_movement_index, mnemonic_str, description_str, &.{
            .{ .mnemonic = mnemonic_str, .description = description_str },
            .{ .mnemonic = "+di", .description = "Directional Indicator Plus" },
            .{ .mnemonic = "-di", .description = "Directional Indicator Minus" },
            .{ .mnemonic = "+dm", .description = "Directional Movement Plus" },
            .{ .mnemonic = "-dm", .description = "Directional Movement Minus" },
            .{ .mnemonic = "atr", .description = "Average True Range" },
            .{ .mnemonic = "tr", .description = "True Range" },
        });
    }

    fn makeOutput(self: *const DirectionalMovementIndex, time: i64) OutputArray {
        var out = OutputArray{};
        out.append(.{ .scalar = Scalar{ .time = time, .value = self.value } });
        return out;
    }

    pub fn updateScalar(self: *DirectionalMovementIndex, sample: *const Scalar) OutputArray {
        _ = self.update(sample.value, sample.value, sample.value);
        return self.makeOutput(sample.time);
    }

    pub fn updateBar(self: *DirectionalMovementIndex, sample: *const Bar) OutputArray {
        _ = self.update(sample.close, sample.high, sample.low);
        return self.makeOutput(sample.time);
    }

    pub fn updateQuote(self: *DirectionalMovementIndex, sample: *const Quote) OutputArray {
        const mid = (sample.bid_price + sample.ask_price) / 2.0;
        _ = self.update(mid, mid, mid);
        return self.makeOutput(sample.time);
    }

    pub fn updateTrade(self: *DirectionalMovementIndex, sample: *const Trade) OutputArray {
        _ = self.update(sample.price, sample.price, sample.price);
        return self.makeOutput(sample.time);
    }

    // --- Indicator interface ---

    pub fn indicator(self: *DirectionalMovementIndex) indicator_mod.Indicator {
        return indicator_mod.Indicator{
            .ptr = @ptrCast(self),
            .vtable = &.{
                .isPrimed = vtableIsPrimed,
                .metadata = vtableMetadata,
                .updateScalar = vtableUpdateScalar,
                .updateBar = vtableUpdateBar,
                .updateQuote = vtableUpdateQuote,
                .updateTrade = vtableUpdateTrade,
            },
        };
    }

    fn vtableIsPrimed(ptr: *const anyopaque) bool {
        const self: *const DirectionalMovementIndex = @ptrCast(@alignCast(ptr));
        return self.isPrimed();
    }

    fn vtableMetadata(ptr: *const anyopaque, out: *Metadata) void {
        const self: *const DirectionalMovementIndex = @ptrCast(@alignCast(ptr));
        self.getMetadata(out);
    }

    fn vtableUpdateScalar(ptr: *anyopaque, sample: *const Scalar) OutputArray {
        const self: *DirectionalMovementIndex = @ptrCast(@alignCast(ptr));
        return self.updateScalar(sample);
    }

    fn vtableUpdateBar(ptr: *anyopaque, sample: *const Bar) OutputArray {
        const self: *DirectionalMovementIndex = @ptrCast(@alignCast(ptr));
        return self.updateBar(sample);
    }

    fn vtableUpdateQuote(ptr: *anyopaque, sample: *const Quote) OutputArray {
        const self: *DirectionalMovementIndex = @ptrCast(@alignCast(ptr));
        return self.updateQuote(sample);
    }

    fn vtableUpdateTrade(ptr: *anyopaque, sample: *const Trade) OutputArray {
        const self: *DirectionalMovementIndex = @ptrCast(@alignCast(ptr));
        return self.updateTrade(sample);
    }
};

// ============================================================================
// Tests
// ============================================================================

const testing = std.testing;

fn almostEqual(a: f64, b: f64, tolerance: f64) bool {
    return @abs(a - b) <= tolerance;
}

// TA-Lib test data (252 entries).
const test_input_high = [_]f64{
    93.250000,  94.940000,  96.375000,  96.190000,  96.000000,  94.720000,  95.000000,  93.720000,  92.470000,  92.750000,  96.250000,
    99.625000,  99.125000,  92.750000,  91.315000,  93.250000,  93.405000,  90.655000,  91.970000,  92.250000,  90.345000,  88.500000,
    88.250000,  85.500000,  84.440000,  84.750000,  84.440000,  89.405000,  88.125000,  89.125000,  87.155000,  87.250000,  87.375000,
    88.970000,  90.000000,  89.845000,  86.970000,  85.940000,  84.750000,  85.470000,  84.470000,  88.500000,  89.470000,  90.000000,
    92.440000,  91.440000,  92.970000,  91.720000,  91.155000,  91.750000,  90.000000,  88.875000,  89.000000,  85.250000,  83.815000,
    85.250000,  86.625000,  87.940000,  89.375000,  90.625000,  90.750000,  88.845000,  91.970000,  93.375000,  93.815000,  94.030000,
    94.030000,  91.815000,  92.000000,  91.940000,  89.750000,  88.750000,  86.155000,  84.875000,  85.940000,  99.375000,  103.280000,
    105.375000, 107.625000, 105.250000, 104.500000, 105.500000, 106.125000, 107.940000, 106.250000, 107.000000, 108.750000, 110.940000,
    110.940000, 114.220000, 123.000000, 121.750000, 119.815000, 120.315000, 119.375000, 118.190000, 116.690000, 115.345000, 113.000000,
    118.315000, 116.870000, 116.750000, 113.870000, 114.620000, 115.310000, 116.000000, 121.690000, 119.870000, 120.870000, 116.750000,
    116.500000, 116.000000, 118.310000, 121.500000, 122.000000, 121.440000, 125.750000, 127.750000, 124.190000, 124.440000, 125.750000,
    124.690000, 125.310000, 132.000000, 131.310000, 132.250000, 133.880000, 133.500000, 135.500000, 137.440000, 138.690000, 139.190000,
    138.500000, 138.130000, 137.500000, 138.880000, 132.130000, 129.750000, 128.500000, 125.440000, 125.120000, 126.500000, 128.690000,
    126.620000, 126.690000, 126.000000, 123.120000, 121.870000, 124.000000, 127.000000, 124.440000, 122.500000, 123.750000, 123.810000,
    124.500000, 127.870000, 128.560000, 129.630000, 124.870000, 124.370000, 124.870000, 123.620000, 124.060000, 125.870000, 125.190000,
    125.620000, 126.000000, 128.500000, 126.750000, 129.750000, 132.690000, 133.940000, 136.500000, 137.690000, 135.560000, 133.560000,
    135.000000, 132.380000, 131.440000, 130.880000, 129.630000, 127.250000, 127.810000, 125.000000, 126.810000, 124.750000, 122.810000,
    122.250000, 121.060000, 120.000000, 123.250000, 122.750000, 119.190000, 115.060000, 116.690000, 114.870000, 110.870000, 107.250000,
    108.870000, 109.000000, 108.500000, 113.060000, 93.000000,  94.620000,  95.120000,  96.000000,  95.560000,  95.310000,  99.000000,
    98.810000,  96.810000,  95.940000,  94.440000,  92.940000,  93.940000,  95.500000,  97.060000,  97.500000,  96.250000,  96.370000,
    95.000000,  94.870000,  98.250000,  105.120000, 108.440000, 109.870000, 105.000000, 106.000000, 104.940000, 104.500000, 104.440000,
    106.310000, 112.870000, 116.500000, 119.190000, 121.000000, 122.120000, 111.940000, 112.750000, 110.190000, 107.940000, 109.690000,
    111.060000, 110.440000, 110.120000, 110.310000, 110.440000, 110.000000, 110.750000, 110.500000, 110.500000, 109.500000,
};

const test_input_low = [_]f64{
    90.750000,  91.405000,  94.250000,  93.500000,  92.815000,  93.500000,  92.000000,  89.750000,  89.440000,  90.625000,  92.750000,
    96.315000,  96.030000,  88.815000,  86.750000,  90.940000,  88.905000,  88.780000,  89.250000,  89.750000,  87.500000,  86.530000,
    84.625000,  82.280000,  81.565000,  80.875000,  81.250000,  84.065000,  85.595000,  85.970000,  84.405000,  85.095000,  85.500000,
    85.530000,  87.875000,  86.565000,  84.655000,  83.250000,  82.565000,  83.440000,  82.530000,  85.065000,  86.875000,  88.530000,
    89.280000,  90.125000,  90.750000,  89.000000,  88.565000,  90.095000,  89.000000,  86.470000,  84.000000,  83.315000,  82.000000,
    83.250000,  84.750000,  85.280000,  87.190000,  88.440000,  88.250000,  87.345000,  89.280000,  91.095000,  89.530000,  91.155000,
    92.000000,  90.530000,  89.970000,  88.815000,  86.750000,  85.065000,  82.030000,  81.500000,  82.565000,  96.345000,  96.470000,
    101.155000, 104.250000, 101.750000, 101.720000, 101.720000, 103.155000, 105.690000, 103.655000, 104.000000, 105.530000, 108.530000,
    108.750000, 107.750000, 117.000000, 118.000000, 116.000000, 118.500000, 116.530000, 116.250000, 114.595000, 110.875000, 110.500000,
    110.720000, 112.620000, 114.190000, 111.190000, 109.440000, 111.560000, 112.440000, 117.500000, 116.060000, 116.560000, 113.310000,
    112.560000, 114.000000, 114.750000, 118.870000, 119.000000, 119.750000, 122.620000, 123.000000, 121.750000, 121.560000, 123.120000,
    122.190000, 122.750000, 124.370000, 128.000000, 129.500000, 130.810000, 130.630000, 132.130000, 133.880000, 135.380000, 135.750000,
    136.190000, 134.500000, 135.380000, 133.690000, 126.060000, 126.870000, 123.500000, 122.620000, 122.750000, 123.560000, 125.810000,
    124.620000, 124.370000, 121.810000, 118.190000, 118.060000, 117.560000, 121.000000, 121.120000, 118.940000, 119.810000, 121.000000,
    122.000000, 124.500000, 126.560000, 123.500000, 121.250000, 121.060000, 122.310000, 121.000000, 120.870000, 122.060000, 122.750000,
    122.690000, 122.870000, 125.500000, 124.250000, 128.000000, 128.380000, 130.690000, 131.630000, 134.380000, 132.000000, 131.940000,
    131.940000, 129.560000, 123.750000, 126.000000, 126.250000, 124.370000, 121.440000, 120.440000, 121.370000, 121.690000, 120.000000,
    119.620000, 115.500000, 116.750000, 119.060000, 119.060000, 115.060000, 111.060000, 113.120000, 110.000000, 105.000000, 104.690000,
    103.870000, 104.690000, 105.440000, 107.000000, 89.000000,  92.500000,  92.120000,  94.620000,  92.810000,  94.250000,  96.250000,
    96.370000,  93.690000,  93.500000,  90.000000,  90.190000,  90.500000,  92.120000,  94.120000,  94.870000,  93.000000,  93.870000,
    93.000000,  92.620000,  93.560000,  98.370000,  104.440000, 106.000000, 101.810000, 104.120000, 103.370000, 102.120000, 102.250000,
    103.370000, 107.940000, 112.500000, 115.440000, 115.500000, 112.250000, 107.560000, 106.560000, 106.870000, 104.500000, 105.750000,
    108.620000, 107.750000, 108.060000, 108.000000, 108.190000, 108.120000, 109.060000, 108.750000, 108.560000, 106.620000,
};

const test_input_close = [_]f64{
    91.500000,  94.815000,  94.375000,  95.095000,  93.780000,  94.625000,  92.530000,  92.750000,  90.315000,  92.470000,  96.125000,
    97.250000,  98.500000,  89.875000,  91.000000,  92.815000,  89.155000,  89.345000,  91.625000,  89.875000,  88.375000,  87.625000,
    84.780000,  83.000000,  83.500000,  81.375000,  84.440000,  89.250000,  86.375000,  86.250000,  85.250000,  87.125000,  85.815000,
    88.970000,  88.470000,  86.875000,  86.815000,  84.875000,  84.190000,  83.875000,  83.375000,  85.500000,  89.190000,  89.440000,
    91.095000,  90.750000,  91.440000,  89.000000,  91.000000,  90.500000,  89.030000,  88.815000,  84.280000,  83.500000,  82.690000,
    84.750000,  85.655000,  86.190000,  88.940000,  89.280000,  88.625000,  88.500000,  91.970000,  91.500000,  93.250000,  93.500000,
    93.155000,  91.720000,  90.000000,  89.690000,  88.875000,  85.190000,  83.375000,  84.875000,  85.940000,  97.250000,  99.875000,
    104.940000, 106.000000, 102.500000, 102.405000, 104.595000, 106.125000, 106.000000, 106.065000, 104.625000, 108.625000, 109.315000,
    110.500000, 112.750000, 123.000000, 119.625000, 118.750000, 119.250000, 117.940000, 116.440000, 115.190000, 111.875000, 110.595000,
    118.125000, 116.000000, 116.000000, 112.000000, 113.750000, 112.940000, 116.000000, 120.500000, 116.620000, 117.000000, 115.250000,
    114.310000, 115.500000, 115.870000, 120.690000, 120.190000, 120.750000, 124.750000, 123.370000, 122.940000, 122.560000, 123.120000,
    122.560000, 124.620000, 129.250000, 131.000000, 132.250000, 131.000000, 132.810000, 134.000000, 137.380000, 137.810000, 137.880000,
    137.250000, 136.310000, 136.250000, 134.630000, 128.250000, 129.000000, 123.870000, 124.810000, 123.000000, 126.250000, 128.380000,
    125.370000, 125.690000, 122.250000, 119.370000, 118.500000, 123.190000, 123.500000, 122.190000, 119.310000, 123.310000, 121.120000,
    123.370000, 127.370000, 128.500000, 123.870000, 122.940000, 121.750000, 124.440000, 122.000000, 122.370000, 122.940000, 124.000000,
    123.190000, 124.560000, 127.250000, 125.870000, 128.860000, 132.000000, 130.750000, 134.750000, 135.000000, 132.380000, 133.310000,
    131.940000, 130.000000, 125.370000, 130.130000, 127.120000, 125.190000, 122.000000, 125.000000, 123.000000, 123.500000, 120.060000,
    121.000000, 117.750000, 119.870000, 122.000000, 119.190000, 116.370000, 113.500000, 114.250000, 110.000000, 105.060000, 107.000000,
    107.870000, 107.000000, 107.120000, 107.000000, 91.000000,  93.940000,  93.870000,  95.500000,  93.000000,  94.940000,  98.250000,
    96.750000,  94.810000,  94.370000,  91.560000,  90.250000,  93.940000,  93.620000,  97.000000,  95.000000,  95.870000,  94.060000,
    94.620000,  93.750000,  98.000000,  103.940000, 107.870000, 106.060000, 104.500000, 105.000000, 104.190000, 103.060000, 103.420000,
    105.270000, 111.870000, 116.000000, 116.620000, 118.280000, 113.370000, 109.000000, 109.700000, 109.250000, 107.000000, 109.190000,
    110.000000, 109.200000, 110.120000, 108.000000, 108.620000, 109.750000, 109.810000, 109.000000, 108.750000, 107.870000,
};

// Expected DX14 (length=14), 252 entries.
const test_expected_dx14 = [_]f64{
    math.nan(f64),       math.nan(f64),       math.nan(f64),       math.nan(f64),       math.nan(f64),       math.nan(f64),       math.nan(f64),       math.nan(f64),       math.nan(f64),       math.nan(f64),
    math.nan(f64),       math.nan(f64),       math.nan(f64),       math.nan(f64),       19.3689005535696000, 9.7130998503806600,  17.2905063974605000, 17.7471496929783000, 10.8157095287173000, 9.3397712431434300,
    18.7092835938902000, 22.4311553110723000, 29.2793287720686000, 36.6890032745911000, 38.7946711439851000, 40.8395606322553000, 40.8395606322553000, 10.1326444821589000, 10.1326444821589000, 4.7955546119510500,
    11.9842161244257000, 11.4341649237631000, 10.6639295128472000, 1.0646275010482200,  4.6852039226580500,  2.8826301098588800,  12.7826473506586000, 19.2995694634555000, 22.3460750080392000, 17.3322394130484000,
    21.6997767612528000, 2.7933742311496300,  7.6130400469284300,  10.2320150505838000, 21.2938681246593000, 21.2938681246593000, 27.7669805490589000, 16.0140426161702000, 13.2259344626447000, 16.1927664523976000,
    8.8189968692245400,  6.0209576859906500,  17.8059510670395000, 20.7731655050416000, 26.2754206704072000, 16.7469176189106000, 8.3131339697192100,  0.8124809802600350,  6.7744112978396900,  12.9219478024900000,
    11.7160518650264000, 5.9142166280538200,  21.1433227104852000, 26.8743335481008000, 16.7004235940457000, 17.6770841792387000, 17.6770841792387000, 7.6671266877000600,  4.0365012318237000,  3.2127805178293200,
    14.6619668911838000, 22.6974998091634000, 34.6361018048549000, 36.4809712576634000, 28.6248019279665000, 27.8203210414220000, 36.5376935972833000, 40.6771235176730000, 44.8385616205920000, 33.6224376151175000,
    33.4888489787486000, 35.7931280102919000, 37.2562343258967000, 41.4302048052794000, 30.9144859726716000, 32.8948019476628000, 37.4035761720982000, 42.6011017155081000, 42.6011017155082000, 49.8360864090902000,
    63.2057249549376000, 63.2057249549376000, 52.4703204162443000, 53.2974027621747000, 42.7568875300255000, 41.2700834860402000, 32.4868564524372000, 15.1559972936640000, 13.5435600693504000, 28.7676173862113000,
    28.7676173862113000, 28.7676173862113000, 14.5478631637740000, 7.1169032878241300,  9.6068450774441100,  12.1432144526731000, 29.6688098742095000, 22.9826577193214000, 25.8423473229681000, 11.3691843648238000,
    8.2744372797095400,  8.2744372797095400,  16.5570933953484000, 26.4359817169243000, 27.8772828785245000, 27.8772828785245000, 39.6901239591527000, 44.2530244824417000, 37.2631216738646000, 37.9111212262017000,
    41.3307451246211000, 35.6201031533738000, 37.4350861447272000, 52.8736044960360000, 52.8736044960360000, 54.6953538037807000, 57.7456384433200000, 56.4926096764700000, 60.2689878927279000, 63.5719583678035000,
    65.5587266984969000, 66.3493142968400000, 66.3493142968400000, 52.6165345910834000, 52.6165345910834000, 39.2813292371374000, 2.2474209459819200,  2.2474209459819200,  15.1992313086169000, 18.2453861720155000,
    18.2453861720155000, 10.9940558059780000, 0.4648121416401460,  5.6995935294201300,  6.8084521191972100,  17.5056808678286000, 29.7802757134790000, 30.1820385419646000, 18.2445566761904000, 3.8063891636106400,
    3.8063891636106500,  12.7815808967734000, 6.6376506148958300,  6.3382012945742100,  2.7643397508718600,  12.6726974978298000, 15.4724751402168000, 0.1392090666055750,  9.3884365668920600,  10.1657197963870000,
    7.5510819447335400,  13.3534129861646000, 10.8371723485473000, 0.9133958685408690,  0.9133958685408750,  1.5160992206377700,  3.7220402946588300,  16.9084514640943000, 8.8787818518446900,  22.6160019643358000,
    33.2382789735989000, 37.1862037283378000, 44.4333967449730000, 47.4675276003482000, 31.9492964852446000, 31.5733960076196000, 36.2661911599917000, 21.4418885662951000, 5.5663755337796600,  5.5663755337796600,
    5.5663755337796500,  13.3541224730898000, 23.8889480831878000, 27.1447456056352000, 17.3590956483689000, 17.3590956483689000, 23.7167639268673000, 25.1118388910418000, 38.2891867128692000, 38.2891867128693000,
    19.1142448296361000, 19.1142448296361000, 32.4782213721424000, 42.6775799548498000, 33.8076109906405000, 41.3265976095972000, 50.9437351582267000, 51.4748008596843000, 42.7762340673126000, 42.0711737017397000,
    42.0711737017397000, 18.3047689978627000, 52.2566004021269000, 46.3611171619619000, 44.5013657146160000, 41.1030559727158000, 44.0192110127676000, 44.0192110127676000, 28.9261906620021000, 28.9261906620021000,
    34.6912635963940000, 35.0932290292580000, 42.1559963363164000, 42.1559963363164000, 37.2087006471159000, 29.6296251807658000, 22.3513570036971000, 20.2997170136863000, 25.9805477026455000, 25.3630351312637000,
    28.1141481055518000, 29.3392763790860000, 11.1877601802003000, 14.9413990750095000, 24.2106124910075000, 27.8573090025640000, 11.0037527631590000, 13.9200946536369000, 10.9826626051613000, 6.0734956092915200,
    6.0734956092915200,  12.7681578403036000, 31.2739879417899000, 38.9873053291662000, 44.0028311076014000, 47.1509041507628000, 32.7226661743566000, 15.1727774462412000, 11.7788636989788000, 11.7788636989787000,
    3.4034957672375100,  8.8353140274169300,  12.9616235680799000, 9.5699806857578000,  9.5699806857578000,  10.2524849037856000, 10.7488269692743000, 10.3947916371434000, 13.5823477682711000, 11.8118413695872000,
    10.6731513188712000, 0.4722272415190610,
};

test "DirectionalMovementIndex update length=14" {
    const tolerance = 1e-8;
    var dx = try DirectionalMovementIndex.init(testing.allocator, .{ .length = 14 });
    defer dx.deinit();

    for (0..test_input_close.len) |i| {
        const act = dx.update(test_input_close[i], test_input_high[i], test_input_low[i]);
        const exp = test_expected_dx14[i];

        if (math.isNan(exp)) {
            try testing.expect(math.isNan(act));
        } else {
            try testing.expect(!math.isNan(act));
            try testing.expect(almostEqual(act, exp, tolerance));
        }
    }
}

test "DirectionalMovementIndex isPrimed length=14" {
    var dx = try DirectionalMovementIndex.init(testing.allocator, .{ .length = 14 });
    defer dx.deinit();

    for (0..14) |i| {
        _ = dx.update(test_input_close[i], test_input_high[i], test_input_low[i]);
        try testing.expect(!dx.isPrimed());
    }

    _ = dx.update(test_input_close[14], test_input_high[14], test_input_low[14]);
    try testing.expect(dx.isPrimed());
}

test "DirectionalMovementIndex constructor validation" {
    try testing.expectError(error.InvalidLength, DirectionalMovementIndex.init(testing.allocator, .{ .length = 0 }));
    try testing.expectError(error.InvalidLength, DirectionalMovementIndex.init(testing.allocator, .{ .length = -8 }));

    var dx = try DirectionalMovementIndex.init(testing.allocator, .{ .length = 14 });
    defer dx.deinit();
    try testing.expect(!dx.isPrimed());
}

test "DirectionalMovementIndex NaN passthrough" {
    var dx = try DirectionalMovementIndex.init(testing.allocator, .{ .length = 14 });
    defer dx.deinit();

    try testing.expect(math.isNan(dx.update(math.nan(f64), 1, 1)));
    try testing.expect(math.isNan(dx.update(1, math.nan(f64), 1)));
    try testing.expect(math.isNan(dx.update(1, 1, math.nan(f64))));
    try testing.expect(math.isNan(dx.updateSample(math.nan(f64))));
}

test "DirectionalMovementIndex metadata" {
    var dx = try DirectionalMovementIndex.init(testing.allocator, .{ .length = 14 });
    defer dx.deinit();
    var meta: Metadata = undefined;
    dx.getMetadata(&meta);

    try testing.expectEqual(Identifier.directional_movement_index, meta.identifier);
    try testing.expectEqualStrings("dx", meta.mnemonic);
    try testing.expectEqual(@as(usize, 7), meta.outputs_len);
}

test "DirectionalMovementIndex updateBar" {
    var dx = try DirectionalMovementIndex.init(testing.allocator, .{ .length = 14 });
    defer dx.deinit();

    for (0..14) |i| {
        _ = dx.update(test_input_close[i], test_input_high[i], test_input_low[i]);
    }

    const bar = Bar{
        .time = 1000,
        .open = 91,
        .high = test_input_high[14],
        .low = test_input_low[14],
        .close = test_input_close[14],
        .volume = 1000,
    };
    const out = dx.updateBar(&bar);
    try testing.expect(!math.isNan(out.slice()[0].scalar.value));
}
