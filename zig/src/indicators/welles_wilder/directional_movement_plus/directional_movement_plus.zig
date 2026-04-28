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

/// Enumerates the outputs of the Directional Movement Plus indicator.
pub const DirectionalMovementPlusOutput = enum(u8) {
    /// The scalar value of +DM.
    value = 1,
};

/// Parameters for the Directional Movement Plus indicator.
pub const DirectionalMovementPlusParams = struct {
    /// The smoothing length. Must be >= 1. Default is 14.
    length: usize = 14,
};

/// Welles Wilder's Directional Movement Plus (+DM) indicator.
///
/// UpMove = today's high − yesterday's high
/// DownMove = yesterday's low − today's low
/// if UpMove > DownMove and UpMove > 0, then +DM = UpMove, else +DM = 0
///
/// When length > 1, Wilder's smoothing is applied:
///   +DM(n) = previous +DM(n) − previous +DM(n)/n + today's +DM(1)
pub const DirectionalMovementPlus = struct {
    length: usize,
    no_smoothing: bool,
    count: usize,
    previous_high: f64,
    previous_low: f64,
    value: f64,
    accumulator: f64,
    primed: bool,

    pub const Error = error{InvalidLength};

    pub fn init(params: DirectionalMovementPlusParams) Error!DirectionalMovementPlus {
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

    pub fn deinit(_: *DirectionalMovementPlus) void {}
    pub fn fixSlices(_: *DirectionalMovementPlus) void {}

    /// Core update given high and low values.
    pub fn update(self: *DirectionalMovementPlus, high_in: f64, low_in: f64) f64 {
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
                if (delta_plus > 0 and delta_plus > delta_minus) {
                    self.value = delta_plus;
                } else {
                    self.value = 0;
                }
            } else {
                if (self.count > 0) {
                    const delta_plus = high - self.previous_high;
                    const delta_minus = self.previous_low - low;
                    if (delta_plus > 0 and delta_plus > delta_minus) {
                        self.value = delta_plus;
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
                if (delta_plus > 0 and delta_plus > delta_minus) {
                    self.accumulator += -self.accumulator / n + delta_plus;
                } else {
                    self.accumulator += -self.accumulator / n;
                }
                self.value = self.accumulator;
            } else {
                if (self.count > 0 and self.length >= self.count) {
                    const delta_plus = high - self.previous_high;
                    const delta_minus = self.previous_low - low;
                    if (self.length > self.count) {
                        if (delta_plus > 0 and delta_plus > delta_minus) {
                            self.accumulator += delta_plus;
                        }
                    } else {
                        if (delta_plus > 0 and delta_plus > delta_minus) {
                            self.accumulator += -self.accumulator / n + delta_plus;
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
    pub fn updateSample(self: *DirectionalMovementPlus, sample: f64) f64 {
        return self.update(sample, sample);
    }

    pub fn isPrimed(self: *const DirectionalMovementPlus) bool {
        return self.primed;
    }

    pub fn getMetadata(_: *const DirectionalMovementPlus, out: *Metadata) void {
        const mnemonic = "+dm";
        const description = "Directional Movement Plus";
        build_metadata_mod.buildMetadata(out, Identifier.directional_movement_plus, mnemonic, description, &.{
            .{ .mnemonic = mnemonic, .description = description },
        });
    }

    fn makeOutput(self: *const DirectionalMovementPlus, time: i64) OutputArray {
        var out = OutputArray{};
        out.append(.{ .scalar = Scalar{ .time = time, .value = self.value } });
        return out;
    }

    pub fn updateScalar(self: *DirectionalMovementPlus, sample: *const Scalar) OutputArray {
        _ = self.update(sample.value, sample.value);
        return self.makeOutput(sample.time);
    }

    pub fn updateBar(self: *DirectionalMovementPlus, sample: *const Bar) OutputArray {
        _ = self.update(sample.high, sample.low);
        return self.makeOutput(sample.time);
    }

    pub fn updateQuote(self: *DirectionalMovementPlus, sample: *const Quote) OutputArray {
        const mid = (sample.bid_price + sample.ask_price) / 2.0;
        _ = self.update(mid, mid);
        return self.makeOutput(sample.time);
    }

    pub fn updateTrade(self: *DirectionalMovementPlus, sample: *const Trade) OutputArray {
        _ = self.update(sample.price, sample.price);
        return self.makeOutput(sample.time);
    }

    // --- Indicator interface ---

    pub fn indicator(self: *DirectionalMovementPlus) indicator_mod.Indicator {
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
        const self: *const DirectionalMovementPlus = @ptrCast(@alignCast(ptr));
        return self.isPrimed();
    }

    fn vtableMetadata(ptr: *const anyopaque, out: *Metadata) void {
        const self: *const DirectionalMovementPlus = @ptrCast(@alignCast(ptr));
        self.getMetadata(out);
    }

    fn vtableUpdateScalar(ptr: *anyopaque, sample: *const Scalar) OutputArray {
        const self: *DirectionalMovementPlus = @ptrCast(@alignCast(ptr));
        return self.updateScalar(sample);
    }

    fn vtableUpdateBar(ptr: *anyopaque, sample: *const Bar) OutputArray {
        const self: *DirectionalMovementPlus = @ptrCast(@alignCast(ptr));
        return self.updateBar(sample);
    }

    fn vtableUpdateQuote(ptr: *anyopaque, sample: *const Quote) OutputArray {
        const self: *DirectionalMovementPlus = @ptrCast(@alignCast(ptr));
        return self.updateQuote(sample);
    }

    fn vtableUpdateTrade(ptr: *anyopaque, sample: *const Trade) OutputArray {
        const self: *DirectionalMovementPlus = @ptrCast(@alignCast(ptr));
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

const test_expected_dmp1 = [_]f64{
    math.nan(f64),
    1.6900,
    1.4350,
    0.0000,
    0.0000,
    0.0000,
    0.0000,
    0.0000,
    0.0000,
    0.2800,
    3.5000,
    3.3750,
    0.0000,
    0.0000,
    0.0000,
    1.9350,
    0.0000,
    0.0000,
    1.3150,
    0.2800,
    0.0000,
    0.0000,
    0.0000,
    0.0000,
    0.0000,
    0.0000,
    0.0000,
    4.9650,
    0.0000,
    1.0000,
    0.0000,
    0.0950,
    0.1250,
    1.5950,
    1.0300,
    0.0000,
    0.0000,
    0.0000,
    0.0000,
    0.7200,
    0.0000,
    4.0300,
    0.9700,
    0.5300,
    2.4400,
    0.0000,
    1.5300,
    0.0000,
    0.0000,
    0.5950,
    0.0000,
    0.0000,
    0.0000,
    0.0000,
    0.0000,
    1.4350,
    1.3750,
    1.3150,
    1.4350,
    1.2500,
    0.0000,
    0.0000,
    3.1250,
    1.4050,
    0.0000,
    0.2150,
    0.0000,
    0.0000,
    0.0000,
    0.0000,
    0.0000,
    0.0000,
    0.0000,
    0.0000,
    1.0650,
    13.4350,
    3.9050,
    2.0950,
    2.2500,
    0.0000,
    0.0000,
    1.0000,
    0.6250,
    1.8150,
    0.0000,
    0.7500,
    1.7500,
    2.1900,
    0.0000,
    3.2800,
    8.7800,
    0.0000,
    0.0000,
    0.5000,
    0.0000,
    0.0000,
    0.0000,
    0.0000,
    0.0000,
    5.3150,
    0.0000,
    0.0000,
    0.0000,
    0.0000,
    0.6900,
    0.6900,
    5.6900,
    0.0000,
    1.0000,
    0.0000,
    0.0000,
    0.0000,
    2.3100,
    3.1900,
    0.5000,
    0.0000,
    4.3100,
    2.0000,
    0.0000,
    0.2500,
    1.3100,
    0.0000,
    0.6200,
    6.6900,
    0.0000,
    0.9400,
    1.6300,
    0.0000,
    2.0000,
    1.9400,
    1.2500,
    0.5000,
    0.0000,
    0.0000,
    0.0000,
    0.0000,
    0.0000,
    0.0000,
    0.0000,
    0.0000,
    0.0000,
    1.3800,
    2.1900,
    0.0000,
    0.0000,
    0.0000,
    0.0000,
    0.0000,
    2.1300,
    3.0000,
    0.0000,
    0.0000,
    1.2500,
    0.0600,
    0.6900,
    3.3700,
    0.6900,
    0.0000,
    0.0000,
    0.0000,
    0.5000,
    0.0000,
    0.4400,
    1.8100,
    0.0000,
    0.4300,
    0.3800,
    2.5000,
    0.0000,
    3.0000,
    2.9400,
    1.2500,
    2.5600,
    1.1900,
    0.0000,
    0.0000,
    1.4400,
    0.0000,
    0.0000,
    0.0000,
    0.0000,
    0.0000,
    0.0000,
    0.0000,
    1.8100,
    0.0000,
    0.0000,
    0.0000,
    0.0000,
    0.0000,
    3.2500,
    0.0000,
    0.0000,
    0.0000,
    1.6300,
    0.0000,
    0.0000,
    0.0000,
    1.6200,
    0.1300,
    0.0000,
    4.5600,
    0.0000,
    1.6200,
    0.5000,
    0.8800,
    0.0000,
    0.0000,
    3.6900,
    0.0000,
    0.0000,
    0.0000,
    0.0000,
    0.0000,
    1.0000,
    1.5600,
    1.5600,
    0.4400,
    0.0000,
    0.1200,
    0.0000,
    0.0000,
    3.3800,
    6.8700,
    3.3200,
    1.4300,
    0.0000,
    1.0000,
    0.0000,
    0.0000,
    0.0000,
    1.8700,
    6.5600,
    3.6300,
    2.6900,
    1.8100,
    0.0000,
    0.0000,
    0.0000,
    0.0000,
    0.0000,
    1.7500,
    1.3700,
    0.0000,
    0.0000,
    0.1900,
    0.1300,
    0.0000,
    0.7500,
    0.0000,
    0.0000,
    0.0000,
};

const test_expected_dmp14 = [_]f64{
    math.nan(f64),      math.nan(f64),      math.nan(f64),      math.nan(f64),      math.nan(f64),      math.nan(f64),      math.nan(f64),      math.nan(f64),      math.nan(f64),      math.nan(f64),
    math.nan(f64),      math.nan(f64),      math.nan(f64),      math.nan(f64),      9.545714285714290,  10.798877551020400, 10.027529154519000, 9.311277072053310,  9.961185852620930,  9.529672577433720,
    8.848981679045600,  8.216911559113770,  7.629989304891360,  7.084990068827690,  6.578919349625710,  6.108996538938160,  5.672639643299720,  10.232451097349700, 9.501561733253330,  9.822878752306670,
    9.121244555713330,  8.564727087448090,  8.077960866916090,  9.095963662136360,  9.476251971983770,  8.799376831127780,  8.170849914618660,  7.587217777860180,  7.045273650870170,  7.262039818665150,
    6.743322688760500,  10.291656782420500, 10.526538440819000, 10.304642837903400, 12.008596920910300, 11.150839997988100, 11.884351426703200, 11.035469181938700, 10.247221383228800, 10.110276998712500,
    9.388114355947290,  8.717534759093910,  8.094853704872920,  7.516649868810570,  6.979746306752670,  7.916192999127480,  8.725750642046950,  9.417482739043590,  10.179805400540500, 10.702676443359000,
    9.938199554547660,  9.228328157794250,  11.694161860808900, 12.263864585036900, 11.387874257534200, 10.789454667710400, 10.018779334302500, 9.303152238995170,  8.638641364781230,  8.021595553011140,
    7.448624442081780,  6.916579839075940,  6.422538421999080,  5.963785677570580,  6.602800986315530,  19.566172344435900, 22.073588605547600, 22.591903705151300, 23.228196297640500, 21.569039419237600,
    20.028393746434900, 19.597794193118200, 18.822951750752600, 19.293455197127400, 17.915351254475400, 17.385683307727200, 17.893848785746700, 18.805716729621900, 17.462451248934600, 19.495133302582200,
    26.882623780969100, 24.962436368042800, 23.179405198896900, 22.023733398975700, 20.450609584763100, 18.989851757280000, 17.633433774617200, 16.373902790716000, 15.204338305664800, 19.433314140974500,
    18.045220273762000, 16.756275968493300, 15.559399113600900, 14.448013462629400, 14.106012501013000, 13.788440179512100, 18.493551595261200, 17.172583624171100, 16.945970508158900, 15.735544043290400,
    14.611576611626800, 13.567892567939200, 14.908757384515000, 17.033846142763900, 16.317142846852200, 15.151632643505600, 18.379373168969500, 19.066560799757400, 17.704663599774700, 16.690044771219400,
    16.807898716132300, 15.607334522122800, 15.112524913399800, 20.723058848156900, 19.242840359002900, 18.808351761931200, 19.094898064650400, 17.730976774318300, 18.464478433295500, 19.085587116631600,
    18.972330894015000, 18.117164401585400, 16.823081230043600, 15.621432570754700, 14.505615958558000, 13.469500532946700, 12.507393352021900, 11.614008112591800, 10.784436104549500, 10.014119239938800,
    9.298825008514630,  10.014623222192200, 11.489292992035600, 10.668629206890200, 9.906584263540870,  9.198971101859380,  8.541901737440850,  7.931765899052220,  9.495211191977060,  11.816981821121600,
    10.972911691041400, 10.189132284538500, 10.711337121357200, 10.006241612688800, 9.981510068925310,  12.638545064002100, 12.425791845144800, 11.538235284777300, 10.714075621578900, 9.948784505751860,
    9.738157041055300,  9.042574395265630,  8.836676224175230,  10.015485065305600, 9.300093274926600,  9.065800898146130,  8.798243691135690,  10.669797713197400, 9.907669305111900,  12.199978640461000,
    14.268551594713800, 14.499369337948600, 16.023700099523700, 16.069150092414800, 14.921353657242300, 13.855542681725000, 14.305861061601800, 13.284013842916000, 12.335155711279100, 11.454073160473500,
    10.635925077582500, 9.876216143469460,  9.170772133221640,  8.515716980848670,  9.717451482216620,  9.023347804915440,  8.378822961707190,  7.780335607299540,  7.224597349635280,  6.708554681804190,
    9.479372204532460,  8.802274189923000,  8.173540319214220,  7.589716010698920,  8.677593438506130,  8.057765335755700,  7.482210668916000,  6.947767049707720,  8.071497974728600,  7.624962405105120,
    7.080322233311900,  11.134584930932500, 10.339257435865900, 11.220739047589700, 10.919257687047600, 11.019310709401400, 10.232217087301300, 9.501344438208310,  12.512676978336300, 11.618914337026600,
    10.788991884381800, 10.018349606926000, 9.302753206431250,  8.638270834543300,  9.021251489218780,  9.936876382846010,  10.787099498357000, 10.456592391331500, 9.709692934807830,  9.136143439464420,
    8.483561765216960,  7.877593067701460,  10.694907848579900, 16.800985859395600, 18.920915440867400, 18.999421480805400, 17.642319946462200, 17.382154236000600, 16.140571790572000, 14.987673805531100,
    13.917125676564600, 14.793045271095700, 20.296399180303200, 22.476656381710100, 23.561180925873600, 23.688239431168400, 21.996222328942100, 20.425063591160500, 18.966130477506200, 17.611406871970000,
    16.353449238257900, 16.935345721239500, 17.095678169722400, 15.874558300456500, 14.740661278995300, 13.877756901924200, 13.016488551786800, 12.086739369516300, 11.973400843122300, 11.118157925756400,
    10.324003788202400, 9.586574946187900,
};

test "DirectionalMovementPlus update length=14" {
    const tolerance = 1e-8;
    var dmp = try DirectionalMovementPlus.init(.{ .length = 14 });

    for (0..test_high.len) |i| {
        const act = dmp.update(test_high[i], test_low[i]);
        const exp = test_expected_dmp14[i];

        if (math.isNan(exp)) {
            try testing.expect(math.isNan(act));
        } else {
            try testing.expect(!math.isNan(act));
            try testing.expect(almostEqual(act, exp, tolerance));
        }
    }
}

test "DirectionalMovementPlus update length=1" {
    const tolerance = 1e-8;
    var dmp = try DirectionalMovementPlus.init(.{ .length = 1 });

    for (0..test_high.len) |i| {
        const act = dmp.update(test_high[i], test_low[i]);
        const exp = test_expected_dmp1[i];

        if (math.isNan(exp)) {
            try testing.expect(math.isNan(act));
        } else {
            try testing.expect(!math.isNan(act));
            try testing.expect(almostEqual(act, exp, tolerance));
        }
    }
}

test "DirectionalMovementPlus constructor validation" {
    const result = DirectionalMovementPlus.init(.{ .length = 0 });
    try testing.expectError(error.InvalidLength, result);
}

test "DirectionalMovementPlus isPrimed length=1" {
    var dmp = try DirectionalMovementPlus.init(.{ .length = 1 });

    try testing.expect(!dmp.isPrimed());

    _ = dmp.update(test_high[0], test_low[0]);
    try testing.expect(!dmp.isPrimed());

    _ = dmp.update(test_high[1], test_low[1]);
    try testing.expect(dmp.isPrimed());
}

test "DirectionalMovementPlus isPrimed length=14" {
    var dmp = try DirectionalMovementPlus.init(.{ .length = 14 });

    for (0..14) |i| {
        _ = dmp.update(test_high[i], test_low[i]);
        try testing.expect(!dmp.isPrimed());
    }

    _ = dmp.update(test_high[14], test_low[14]);
    try testing.expect(dmp.isPrimed());
}

test "DirectionalMovementPlus NaN passthrough" {
    var dmp = try DirectionalMovementPlus.init(.{ .length = 14 });

    try testing.expect(math.isNan(dmp.update(math.nan(f64), 1)));
    try testing.expect(math.isNan(dmp.update(1, math.nan(f64))));
    try testing.expect(math.isNan(dmp.update(math.nan(f64), math.nan(f64))));
    try testing.expect(math.isNan(dmp.updateSample(math.nan(f64))));
}

test "DirectionalMovementPlus high/low swap" {
    var dmp1 = try DirectionalMovementPlus.init(.{ .length = 1 });
    var dmp2 = try DirectionalMovementPlus.init(.{ .length = 1 });

    _ = dmp1.update(10, 5);
    _ = dmp2.update(5, 10);

    const v1 = dmp1.update(12, 6);
    const v2 = dmp2.update(6, 12);

    try testing.expectEqual(v1, v2);
}

test "DirectionalMovementPlus metadata" {
    var dmp = try DirectionalMovementPlus.init(.{ .length = 14 });
    var meta: Metadata = undefined;
    dmp.getMetadata(&meta);

    try testing.expectEqual(Identifier.directional_movement_plus, meta.identifier);
    try testing.expectEqualStrings("+dm", meta.mnemonic);
    try testing.expectEqualStrings("Directional Movement Plus", meta.description);
    try testing.expectEqual(@as(usize, 1), meta.outputs_len);
}

test "DirectionalMovementPlus updateBar" {
    var dmp = try DirectionalMovementPlus.init(.{ .length = 14 });

    for (0..14) |i| {
        _ = dmp.update(test_high[i], test_low[i]);
    }

    const bar = Bar{ .time = 42, .open = 0, .high = test_high[14], .low = test_low[14], .close = 0, .volume = 0 };
    const out = dmp.updateBar(&bar);
    try testing.expectEqual(@as(usize, 1), out.len);
}
