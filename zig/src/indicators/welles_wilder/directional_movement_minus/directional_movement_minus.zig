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

const OutputArray = indicator_mod.OutputArray;
const Identifier = identifier_mod.Identifier;
const Metadata = metadata_mod.Metadata;

/// Enumerates the outputs of the Directional Movement Minus indicator.
pub const DirectionalMovementMinusOutput = enum(u8) {
    /// The scalar value of -DM.
    value = 1,
};

/// Parameters for the Directional Movement Minus indicator.
pub const DirectionalMovementMinusParams = struct {
    /// The smoothing length. Must be >= 1. Default is 14.
    length: usize = 14,
};

/// Welles Wilder's Directional Movement Minus (-DM) indicator.
///
/// UpMove = today's high − yesterday's high
/// DownMove = yesterday's low − today's low
/// if DownMove > UpMove and DownMove > 0, then -DM = DownMove, else -DM = 0
///
/// When length > 1, Wilder's smoothing is applied:
///   -DM(n) = previous -DM(n) − previous -DM(n)/n + today's -DM(1)
pub const DirectionalMovementMinus = struct {
    length: usize,
    no_smoothing: bool,
    count: usize,
    previous_high: f64,
    previous_low: f64,
    value: f64,
    accumulator: f64,
    primed: bool,

    pub const Error = error{InvalidLength};

    pub fn init(params: DirectionalMovementMinusParams) Error!DirectionalMovementMinus {
        if (params.length < 1) return error.InvalidLength;

        return .{
            .length = params.length,
            .no_smoothing = params.length == 1,
            .count = 0,
            .previous_high = 0,
            .previous_low = 0,
            .value = math.nan(f64),
            .accumulator = 0,
            .primed = false,
        };
    }

    pub fn deinit(_: *DirectionalMovementMinus) void {}
    pub fn fixSlices(_: *DirectionalMovementMinus) void {}

    /// Core update given high and low values.
    pub fn update(self: *DirectionalMovementMinus, high_in: f64, low_in: f64) f64 {
        if (math.isNan(high_in) or math.isNan(low_in)) {
            return math.nan(f64);
        }

        var high = high_in;
        var low = low_in;
        if (high < low) {
            const tmp = high;
            high = low;
            low = tmp;
        }

        if (self.no_smoothing) {
            if (self.primed) {
                const delta_plus = high - self.previous_high;
                const delta_minus = self.previous_low - low;
                if (delta_minus > 0 and delta_minus > delta_plus) {
                    self.value = delta_minus;
                } else {
                    self.value = 0;
                }
            } else {
                if (self.count > 0) {
                    const delta_plus = high - self.previous_high;
                    const delta_minus = self.previous_low - low;
                    if (delta_minus > 0 and delta_minus > delta_plus) {
                        self.value = delta_minus;
                    } else {
                        self.value = 0;
                    }
                    self.primed = true;
                }
                self.count += 1;
            }
        } else {
            const n: f64 = @floatFromInt(self.length);
            if (self.primed) {
                const delta_plus = high - self.previous_high;
                const delta_minus = self.previous_low - low;
                if (delta_minus > 0 and delta_minus > delta_plus) {
                    self.accumulator += -self.accumulator / n + delta_minus;
                } else {
                    self.accumulator += -self.accumulator / n;
                }
                self.value = self.accumulator;
            } else {
                if (self.count > 0 and self.length >= self.count) {
                    const delta_plus = high - self.previous_high;
                    const delta_minus = self.previous_low - low;
                    if (self.length > self.count) {
                        if (delta_minus > 0 and delta_minus > delta_plus) {
                            self.accumulator += delta_minus;
                        }
                    } else {
                        if (delta_minus > 0 and delta_minus > delta_plus) {
                            self.accumulator += -self.accumulator / n + delta_minus;
                        } else {
                            self.accumulator += -self.accumulator / n;
                        }
                        self.value = self.accumulator;
                        self.primed = true;
                    }
                }
                self.count += 1;
            }
        }

        self.previous_low = low;
        self.previous_high = high;

        return self.value;
    }

    /// Update using a single sample value as substitute for high and low.
    pub fn updateSample(self: *DirectionalMovementMinus, sample: f64) f64 {
        return self.update(sample, sample);
    }

    pub fn isPrimed(self: *const DirectionalMovementMinus) bool {
        return self.primed;
    }

    pub fn getMetadata(_: *const DirectionalMovementMinus, out: *Metadata) void {
        const mnemonic = "-dm";
        const description = "Directional Movement Minus";
        build_metadata_mod.buildMetadata(out, Identifier.directional_movement_minus, mnemonic, description, &.{
            .{ .mnemonic = mnemonic, .description = description },
        });
    }

    fn makeOutput(self: *const DirectionalMovementMinus, time: i64) OutputArray {
        var out = OutputArray{};
        out.append(.{ .scalar = Scalar{ .time = time, .value = self.value } });
        return out;
    }

    pub fn updateScalar(self: *DirectionalMovementMinus, sample: *const Scalar) OutputArray {
        _ = self.update(sample.value, sample.value);
        return self.makeOutput(sample.time);
    }

    pub fn updateBar(self: *DirectionalMovementMinus, sample: *const Bar) OutputArray {
        _ = self.update(sample.high, sample.low);
        return self.makeOutput(sample.time);
    }

    pub fn updateQuote(self: *DirectionalMovementMinus, sample: *const Quote) OutputArray {
        const mid = (sample.bid_price + sample.ask_price) / 2.0;
        _ = self.update(mid, mid);
        return self.makeOutput(sample.time);
    }

    pub fn updateTrade(self: *DirectionalMovementMinus, sample: *const Trade) OutputArray {
        _ = self.update(sample.price, sample.price);
        return self.makeOutput(sample.time);
    }

    // --- Indicator interface ---

    pub fn indicator(self: *DirectionalMovementMinus) indicator_mod.Indicator {
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
        const self: *const DirectionalMovementMinus = @ptrCast(@alignCast(ptr));
        return self.isPrimed();
    }

    fn vtableMetadata(ptr: *const anyopaque, out: *Metadata) void {
        const self: *const DirectionalMovementMinus = @ptrCast(@alignCast(ptr));
        self.getMetadata(out);
    }

    fn vtableUpdateScalar(ptr: *anyopaque, sample: *const Scalar) OutputArray {
        const self: *DirectionalMovementMinus = @ptrCast(@alignCast(ptr));
        return self.updateScalar(sample);
    }

    fn vtableUpdateBar(ptr: *anyopaque, sample: *const Bar) OutputArray {
        const self: *DirectionalMovementMinus = @ptrCast(@alignCast(ptr));
        return self.updateBar(sample);
    }

    fn vtableUpdateQuote(ptr: *anyopaque, sample: *const Quote) OutputArray {
        const self: *DirectionalMovementMinus = @ptrCast(@alignCast(ptr));
        return self.updateQuote(sample);
    }

    fn vtableUpdateTrade(ptr: *anyopaque, sample: *const Trade) OutputArray {
        const self: *DirectionalMovementMinus = @ptrCast(@alignCast(ptr));
        return self.updateTrade(sample);
    }
};

// ============================================================================
// Tests
// ============================================================================

const testing = std.testing;

fn almostEqual(a: f64, b: f64, tolerance: f64) bool {
    if (math.isNan(a) and math.isNan(b)) return true;
    if (math.isNan(a) or math.isNan(b)) return false;
    return @abs(a - b) <= tolerance;
}

const test_high = [_]f64{
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

const test_low = [_]f64{
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

const test_expected_dmm1 = [_]f64{
    math.nan(f64),
    0.0000,
    0.0000,
    0.7500,
    0.6850,
    0.0000,
    1.5000,
    2.2500,
    0.3100,
    0.0000,
    0.0000,
    0.0000,
    0.2850,
    7.2150,
    2.0650,
    0.0000,
    2.0350,
    0.1250,
    0.0000,
    0.0000,
    2.2500,
    0.9700,
    1.9050,
    2.3450,
    0.7150,
    0.6900,
    0.0000,
    0.0000,
    0.0000,
    0.0000,
    1.5650,
    0.0000,
    0.0000,
    0.0000,
    0.0000,
    1.3100,
    1.9100,
    1.4050,
    0.6850,
    0.0000,
    0.9100,
    0.0000,
    0.0000,
    0.0000,
    0.0000,
    0.0000,
    0.0000,
    1.7500,
    0.4350,
    0.0000,
    1.0950,
    2.5300,
    2.4700,
    0.6850,
    1.3150,
    0.0000,
    0.0000,
    0.0000,
    0.0000,
    0.0000,
    0.1900,
    0.9050,
    0.0000,
    0.0000,
    1.5650,
    0.0000,
    0.0000,
    1.4700,
    0.5600,
    1.1550,
    2.0650,
    1.6850,
    3.0350,
    0.5300,
    0.0000,
    0.0000,
    0.0000,
    0.0000,
    0.0000,
    2.5000,
    0.0300,
    0.0000,
    0.0000,
    0.0000,
    2.0350,
    0.0000,
    0.0000,
    0.0000,
    0.0000,
    0.0000,
    0.0000,
    0.0000,
    2.0000,
    0.0000,
    1.9700,
    0.2800,
    1.6550,
    3.7200,
    0.3750,
    0.0000,
    0.0000,
    0.0000,
    3.0000,
    1.7500,
    0.0000,
    0.0000,
    0.0000,
    1.4400,
    0.0000,
    3.2500,
    0.7500,
    0.0000,
    0.0000,
    0.0000,
    0.0000,
    0.0000,
    0.0000,
    0.0000,
    1.2500,
    0.0000,
    0.0000,
    0.9300,
    0.0000,
    0.0000,
    0.0000,
    0.0000,
    0.0000,
    0.1800,
    0.0000,
    0.0000,
    0.0000,
    0.0000,
    0.0000,
    1.6900,
    0.0000,
    1.6900,
    7.6300,
    0.0000,
    3.3700,
    0.8800,
    0.0000,
    0.0000,
    0.0000,
    1.1900,
    0.2500,
    2.5600,
    3.6200,
    0.1300,
    0.0000,
    0.0000,
    0.0000,
    2.1800,
    0.0000,
    0.0000,
    0.0000,
    0.0000,
    0.0000,
    3.0600,
    2.2500,
    0.1900,
    0.0000,
    1.3100,
    0.0000,
    0.0000,
    0.0000,
    0.0000,
    0.0000,
    0.0000,
    1.2500,
    0.0000,
    0.0000,
    0.0000,
    0.0000,
    0.0000,
    2.3800,
    0.0600,
    0.0000,
    2.3800,
    5.8100,
    0.0000,
    0.0000,
    1.8800,
    2.9300,
    1.0000,
    0.0000,
    0.0000,
    1.6900,
    0.3800,
    4.1200,
    0.0000,
    0.0000,
    0.0000,
    4.0000,
    4.0000,
    0.0000,
    3.1200,
    5.0000,
    0.3100,
    0.0000,
    0.0000,
    0.0000,
    0.0000,
    18.0000,
    0.0000,
    0.0000,
    0.0000,
    1.8100,
    0.0000,
    0.0000,
    0.0000,
    2.6800,
    0.1900,
    3.5000,
    0.0000,
    0.0000,
    0.0000,
    0.0000,
    0.0000,
    1.8700,
    0.0000,
    0.8700,
    0.3800,
    0.0000,
    0.0000,
    0.0000,
    0.0000,
    4.1900,
    0.0000,
    0.7500,
    1.2500,
    0.0000,
    0.0000,
    0.0000,
    0.0000,
    0.0000,
    0.0000,
    3.2500,
    4.6900,
    1.0000,
    0.0000,
    2.3700,
    0.0000,
    0.0000,
    0.8700,
    0.0000,
    0.0000,
    0.0000,
    0.0700,
    0.0000,
    0.3100,
    0.1900,
    1.9400,
};

const test_expected_dmm14 = [_]f64{
    math.nan(f64),      math.nan(f64),      math.nan(f64),      math.nan(f64),      math.nan(f64),      math.nan(f64),      math.nan(f64),      math.nan(f64),      math.nan(f64),      math.nan(f64),
    math.nan(f64),      math.nan(f64),      math.nan(f64),      math.nan(f64),      14.131785714285700, 13.122372448979600, 14.220060131195300, 13.329341550395700, 12.377245725367400, 11.493156744984000,
    12.922216977485200, 12.969201479093400, 13.947829944872400, 15.296556377381500, 14.918945207568600, 14.543306264170800, 13.504498673872900, 12.539891625739100, 11.644185081043500, 10.812457575254600,
    11.605139177022200, 10.776200664377700, 10.006472045493600, 9.291724042244060,  8.628029467798060,  9.321741648669630,  10.565902959478900, 11.216195605230400, 11.100038776285400, 10.307178863693600,
    10.480951802001200, 9.732312387572540,  9.037147217031640,  8.391636701529380,  7.792234079991570,  7.235645931420740,  6.718814079176400,  7.988898787806660,  7.853263160106190,  7.292315791527170,
    7.866436092132370,  9.834547799837210,  11.602080099848800, 11.458360092716800, 11.954905800379900, 11.100983957495600, 10.308056531960200, 9.571766779677320,  8.888069152557510,  8.253207070231970,
    7.853692279501110,  8.197714259536750,  7.612163240998410,  7.068437295212810,  8.128548916983320,  7.547938280055940,  7.008799831480520,  7.978171272089050,  7.968301895511260,  8.554137474403320,
    10.008127654803100, 10.978261393745700, 13.229099865621000, 12.814164160933800, 11.898866720867100, 11.048947669376600, 10.259737121564000, 9.526898755737980,  8.846405987470980,  10.714519845508800,
    9.979196999401000,  9.266397213729500,  8.604511698463110,  7.989903720001460,  9.454196311429920,  8.778896574899210,  8.151832533834980,  7.569558781418200,  7.028876011316900,  6.526813439079980,
    6.060612479145690,  5.627711587778140,  7.225732188651130,  6.709608460890340,  8.200350713683890,  7.894611376992180,  8.985710564349880,  12.063874095467700, 11.577168802934300, 10.750228174153300,
    9.982354733142360,  9.269329395060770,  11.607234438270700, 12.528146264108500, 11.633278673815100, 10.802330197114000, 10.030735183034400, 10.754254098531900, 9.986093091493950,  12.522800727815800,
    12.378314961543300, 11.494149607147300, 10.673138920922500, 9.910771855142320,  9.202859579775010,  8.545512466933940,  7.935118719295800,  7.368324525060390,  8.092015630413220,  7.514014513955130,
    6.977299191529760,  7.408920677849070,  6.879712058002710,  6.388304053859660,  5.931996621441110,  5.508282577052460,  5.114833821548720,  4.929488548580960,  4.577382223682320,  4.250426350562150,
    3.946824468379140,  3.664908434923490,  3.403129261000380,  4.850048599500350,  4.503616556678900,  5.871929659773260,  13.082506112646600, 12.148041390314700, 14.650324148149400, 14.483872423281500,
    13.449310107332900, 12.488645099666200, 11.596599021118600, 11.958270519610200, 11.354108339638000, 13.103100601092400, 15.787164843871600, 14.789510212166400, 13.733116625583100, 12.752179723755800,
    11.841309743487500, 13.175501904667000, 12.234394625762200, 11.360509295350600, 10.549044345682700, 9.795541178133930,  9.095859665410080,  11.506155403595100, 12.934287160481100, 12.200409506161100,
    11.328951684292400, 11.829740849700100, 10.984759360435800, 10.200133691833200, 9.471552713845160,  8.795013234284790,  8.166798003264450,  7.583455288745560,  8.291779910978020,  7.699509917336730,
    7.149544923241250,  6.638863143009730,  6.164658632794750,  5.724325873309410,  7.695445453787310,  7.205770778516790,  6.691072865765590,  8.593139089639470,  13.789343440379500, 12.804390337495300,
    11.889791027674200, 12.920520239983200, 14.927625937127200, 14.861366941618100, 13.799840731502600, 12.814137822109500, 13.588842263387400, 12.998210673145400, 16.189767053635100, 15.033355121232600,
    13.959544041144500, 12.962433752491300, 16.036545627313400, 18.891078082505300, 17.541715362326300, 19.408735693588700, 23.022397429761000, 21.687940470492300, 20.138801865457200, 18.700316017924500,
    17.364579159501300, 16.124252076679800, 32.972519785488400, 30.617339800810700, 28.430386957895600, 26.399645032331600, 26.323956101450800, 24.443673522775800, 22.697696842577500, 21.076432782393400,
    22.250973297936700, 20.851618062369800, 22.862216772200500, 21.229201288471900, 19.712829767866800, 18.304770498733400, 16.997286891681000, 15.783194970846700, 16.525823901500500, 15.345407908536200,
    15.119307343640700, 14.419356819095000, 13.389402760588200, 12.433016849117600, 11.544944217037800, 10.720305344392200, 14.144569248364200, 13.134242873481000, 12.946082668232400, 13.271362477644400,
    12.323408014955500, 11.443164585315800, 10.625795686364700, 9.866810280195770,  9.162038117324650,  8.507606823230030,  11.149920621570700, 15.043497720030000, 14.968962168599300, 13.899750585127900,
    15.276911257618800, 14.185703310646000, 13.172438788457000, 13.101550303567200, 12.165725281883800, 11.296744904606400, 10.489834554277400, 9.810560657543280,  9.109806324861620,  8.769105873085790,
    8.332741167865380,  9.677545370160700,
};

test "DirectionalMovementMinus update length=14" {
    const tolerance = 1e-8;
    var dmm = try DirectionalMovementMinus.init(.{ .length = 14 });

    for (0..test_high.len) |i| {
        const act = dmm.update(test_high[i], test_low[i]);
        const exp = test_expected_dmm14[i];

        if (math.isNan(exp)) {
            try testing.expect(math.isNan(act));
        } else {
            try testing.expect(!math.isNan(act));
            try testing.expect(almostEqual(act, exp, tolerance));
        }
    }
}

test "DirectionalMovementMinus update length=1" {
    const tolerance = 1e-8;
    var dmm = try DirectionalMovementMinus.init(.{ .length = 1 });

    for (0..test_high.len) |i| {
        const act = dmm.update(test_high[i], test_low[i]);
        const exp = test_expected_dmm1[i];

        if (math.isNan(exp)) {
            try testing.expect(math.isNan(act));
        } else {
            try testing.expect(!math.isNan(act));
            try testing.expect(almostEqual(act, exp, tolerance));
        }
    }
}

test "DirectionalMovementMinus constructor validation" {
    const result = DirectionalMovementMinus.init(.{ .length = 0 });
    try testing.expectError(error.InvalidLength, result);
}

test "DirectionalMovementMinus isPrimed length=1" {
    var dmm = try DirectionalMovementMinus.init(.{ .length = 1 });

    try testing.expect(!dmm.isPrimed());

    _ = dmm.update(test_high[0], test_low[0]);
    try testing.expect(!dmm.isPrimed());

    _ = dmm.update(test_high[1], test_low[1]);
    try testing.expect(dmm.isPrimed());
}

test "DirectionalMovementMinus isPrimed length=14" {
    var dmm = try DirectionalMovementMinus.init(.{ .length = 14 });

    for (0..14) |i| {
        _ = dmm.update(test_high[i], test_low[i]);
        try testing.expect(!dmm.isPrimed());
    }

    _ = dmm.update(test_high[14], test_low[14]);
    try testing.expect(dmm.isPrimed());
}

test "DirectionalMovementMinus NaN passthrough" {
    var dmm = try DirectionalMovementMinus.init(.{ .length = 14 });

    try testing.expect(math.isNan(dmm.update(math.nan(f64), 1)));
    try testing.expect(math.isNan(dmm.update(1, math.nan(f64))));
    try testing.expect(math.isNan(dmm.update(math.nan(f64), math.nan(f64))));
    try testing.expect(math.isNan(dmm.updateSample(math.nan(f64))));
}

test "DirectionalMovementMinus high/low swap" {
    var dmm1 = try DirectionalMovementMinus.init(.{ .length = 1 });
    var dmm2 = try DirectionalMovementMinus.init(.{ .length = 1 });

    _ = dmm1.update(10, 5);
    _ = dmm2.update(5, 10);

    const v1 = dmm1.update(12, 6);
    const v2 = dmm2.update(6, 12);

    try testing.expectEqual(v1, v2);
}

test "DirectionalMovementMinus metadata" {
    var dmm = try DirectionalMovementMinus.init(.{ .length = 14 });
    var meta: Metadata = undefined;
    dmm.getMetadata(&meta);

    try testing.expectEqual(Identifier.directional_movement_minus, meta.identifier);
    try testing.expectEqualStrings("-dm", meta.mnemonic);
    try testing.expectEqualStrings("Directional Movement Minus", meta.description);
    try testing.expectEqual(@as(usize, 1), meta.outputs_len);
}

test "DirectionalMovementMinus updateBar" {
    var dmm = try DirectionalMovementMinus.init(.{ .length = 14 });

    for (0..14) |i| {
        _ = dmm.update(test_high[i], test_low[i]);
    }

    const bar = Bar{ .time = 42, .open = 0, .high = test_high[14], .low = test_low[14], .close = 0, .volume = 0 };
    const out = dmm.updateBar(&bar);
    try testing.expectEqual(@as(usize, 1), out.len);
}
