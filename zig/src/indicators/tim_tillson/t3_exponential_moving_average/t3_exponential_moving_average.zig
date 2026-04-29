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

/// Enumerates the outputs of the T3 exponential moving average indicator.
pub const T3ExponentialMovingAverageOutput = enum(u8) {
    value = 1,
};

/// Parameters to create an instance of the T3 EMA based on length.
pub const T3ExponentialMovingAverageLengthParams = struct {
    length: usize,
    volume_factor: f64 = 0.7,
    first_is_average: bool = false,
    bar_component: ?bar_component.BarComponent = null,
    quote_component: ?quote_component.QuoteComponent = null,
    trade_component: ?trade_component.TradeComponent = null,
};

/// Parameters to create an instance of the T3 EMA based on smoothing factor.
pub const T3ExponentialMovingAverageSmoothingFactorParams = struct {
    smoothing_factor: f64,
    volume_factor: f64 = 0.7,
    first_is_average: bool = false,
    bar_component: ?bar_component.BarComponent = null,
    quote_component: ?quote_component.QuoteComponent = null,
    trade_component: ?trade_component.TradeComponent = null,
};

/// Computes the T3 Exponential Moving Average.
///
/// T3 is a six-pole non-linear Kalman filter developed by Tim Tillson.
///
///   c1*EMA6 + c2*EMA5 + c3*EMA4 + c4*EMA3
///
/// Warmup period: 6*length - 6 samples before primed.
pub const T3ExponentialMovingAverage = struct {
    line: LineIndicator,
    smoothing_factor: f64,
    c1: f64,
    c2: f64,
    c3: f64,
    c4: f64,
    sum: f64,
    ema1: f64,
    ema2: f64,
    ema3: f64,
    ema4: f64,
    ema5: f64,
    ema6: f64,
    length: usize,
    length2: usize,
    length3: usize,
    length4: usize,
    length5: usize,
    length6: usize,
    count: usize,
    first_is_average: bool,
    primed: bool,
    mnemonic_buf: [96]u8,
    mnemonic_len: usize,
    description_buf: [160]u8,
    description_len: usize,

    const epsilon: f64 = 0.00000001;

    /// Create T3 EMA from length.
    pub fn initLength(params: T3ExponentialMovingAverageLengthParams) !T3ExponentialMovingAverage {
        if (params.length < 2) {
            return error.InvalidLength;
        }
        if (params.volume_factor < 0.0 or params.volume_factor > 1.0) {
            return error.InvalidVolumeFactor;
        }
        const alpha = 2.0 / @as(f64, @floatFromInt(1 + params.length));
        return initInternal(params.length, alpha, false, params.volume_factor, params.first_is_average, params.bar_component, params.quote_component, params.trade_component);
    }

    /// Create T3 EMA from smoothing factor.
    pub fn initSmoothingFactor(params: T3ExponentialMovingAverageSmoothingFactorParams) !T3ExponentialMovingAverage {
        if (params.volume_factor < 0.0 or params.volume_factor > 1.0) {
            return error.InvalidVolumeFactor;
        }
        var alpha = params.smoothing_factor;
        if (alpha < 0.0 or alpha > 1.0) {
            return error.InvalidSmoothingFactor;
        }
        if (alpha < epsilon) {
            alpha = epsilon;
        }
        const length: usize = @intFromFloat(@round(2.0 / alpha) - 1.0);
        return initInternal(length, alpha, true, params.volume_factor, params.first_is_average, params.bar_component, params.quote_component, params.trade_component);
    }

    fn initInternal(
        length: usize,
        alpha: f64,
        is_alpha_mode: bool,
        v: f64,
        first_is_average: bool,
        bc_opt: ?bar_component.BarComponent,
        qc_opt: ?quote_component.QuoteComponent,
        tc_opt: ?trade_component.TradeComponent,
    ) !T3ExponentialMovingAverage {
        const bc = bc_opt orelse bar_component.default_bar_component;
        const qc = qc_opt orelse quote_component.default_quote_component;
        const tc = tc_opt orelse trade_component.default_trade_component;

        var triple_buf: [64]u8 = undefined;
        const triple = component_triple_mnemonic_mod.componentTripleMnemonic(&triple_buf, bc, qc, tc);

        var mnemonic_buf: [96]u8 = undefined;
        var mnemonic_slice: []u8 = undefined;
        if (is_alpha_mode) {
            mnemonic_slice = std.fmt.bufPrint(&mnemonic_buf, "t3({d}, {d:.8}, {d:.8}{s})", .{ length, alpha, v, triple }) catch
                return error.MnemonicTooLong;
        } else {
            mnemonic_slice = std.fmt.bufPrint(&mnemonic_buf, "t3({d}, {d:.8}{s})", .{ length, v, triple }) catch
                return error.MnemonicTooLong;
        }
        const mnemonic_len = mnemonic_slice.len;

        var description_buf: [160]u8 = undefined;
        const desc_slice = std.fmt.bufPrint(&description_buf, "T3 exponential moving average {s}", .{mnemonic_slice}) catch
            return error.MnemonicTooLong;
        const description_len = desc_slice.len;

        const vv = v * v;
        const c1 = -vv * v;
        const c2 = 3.0 * (vv - c1);
        const c3 = -6.0 * vv - 3.0 * (v - c1);
        const c4 = 1.0 + 3.0 * v - c1 + 3.0 * vv;

        return .{
            .line = LineIndicator.new(
                mnemonic_buf[0..mnemonic_len],
                description_buf[0..description_len],
                bc_opt,
                qc_opt,
                tc_opt,
            ),
            .smoothing_factor = alpha,
            .c1 = c1,
            .c2 = c2,
            .c3 = c3,
            .c4 = c4,
            .sum = 0.0,
            .ema1 = 0.0,
            .ema2 = 0.0,
            .ema3 = 0.0,
            .ema4 = 0.0,
            .ema5 = 0.0,
            .ema6 = 0.0,
            .length = length,
            .length2 = 2 * length - 1,
            .length3 = 3 * length - 2,
            .length4 = 4 * length - 3,
            .length5 = 5 * length - 4,
            .length6 = 6 * length - 5,
            .count = 0,
            .first_is_average = first_is_average,
            .primed = false,
            .mnemonic_buf = mnemonic_buf,
            .mnemonic_len = mnemonic_len,
            .description_buf = description_buf,
            .description_len = description_len,
        };
    }

    pub fn fixSlices(self: *T3ExponentialMovingAverage) void {
        self.line.mnemonic = self.mnemonic_buf[0..self.mnemonic_len];
        self.line.description = self.description_buf[0..self.description_len];
    }

    pub fn update(self: *T3ExponentialMovingAverage, sample: f64) f64 {
        if (math.isNan(sample)) {
            return sample;
        }

        if (self.primed) {
            var v1 = self.ema1;
            var v2 = self.ema2;
            var v3 = self.ema3;
            var v4 = self.ema4;
            var v5 = self.ema5;
            var v6 = self.ema6;
            const sf = self.smoothing_factor;
            v1 += (sample - v1) * sf;
            v2 += (v1 - v2) * sf;
            v3 += (v2 - v3) * sf;
            v4 += (v3 - v4) * sf;
            v5 += (v4 - v5) * sf;
            v6 += (v5 - v6) * sf;
            self.ema1 = v1;
            self.ema2 = v2;
            self.ema3 = v3;
            self.ema4 = v4;
            self.ema5 = v5;
            self.ema6 = v6;
            return self.c1 * v6 + self.c2 * v5 + self.c3 * v4 + self.c4 * v3;
        }

        self.count += 1;
        if (self.first_is_average) {
            if (self.count == 1) {
                self.sum = sample;
            } else if (self.length >= self.count) {
                self.sum += sample;
                if (self.length == self.count) {
                    self.ema1 = self.sum / @as(f64, @floatFromInt(self.length));
                    self.sum = self.ema1;
                }
            } else if (self.length2 >= self.count) {
                self.ema1 += (sample - self.ema1) * self.smoothing_factor;
                self.sum += self.ema1;

                if (self.length2 == self.count) {
                    self.ema2 = self.sum / @as(f64, @floatFromInt(self.length));
                    self.sum = self.ema2;
                }
            } else if (self.length3 >= self.count) {
                self.ema1 += (sample - self.ema1) * self.smoothing_factor;
                self.ema2 += (self.ema1 - self.ema2) * self.smoothing_factor;
                self.sum += self.ema2;

                if (self.length3 == self.count) {
                    self.ema3 = self.sum / @as(f64, @floatFromInt(self.length));
                    self.sum = self.ema3;
                }
            } else if (self.length4 >= self.count) {
                self.ema1 += (sample - self.ema1) * self.smoothing_factor;
                self.ema2 += (self.ema1 - self.ema2) * self.smoothing_factor;
                self.ema3 += (self.ema2 - self.ema3) * self.smoothing_factor;
                self.sum += self.ema3;

                if (self.length4 == self.count) {
                    self.ema4 = self.sum / @as(f64, @floatFromInt(self.length));
                    self.sum = self.ema4;
                }
            } else if (self.length5 >= self.count) {
                self.ema1 += (sample - self.ema1) * self.smoothing_factor;
                self.ema2 += (self.ema1 - self.ema2) * self.smoothing_factor;
                self.ema3 += (self.ema2 - self.ema3) * self.smoothing_factor;
                self.ema4 += (self.ema3 - self.ema4) * self.smoothing_factor;
                self.sum += self.ema4;

                if (self.length5 == self.count) {
                    self.ema5 = self.sum / @as(f64, @floatFromInt(self.length));
                    self.sum = self.ema5;
                }
            } else {
                self.ema1 += (sample - self.ema1) * self.smoothing_factor;
                self.ema2 += (self.ema1 - self.ema2) * self.smoothing_factor;
                self.ema3 += (self.ema2 - self.ema3) * self.smoothing_factor;
                self.ema4 += (self.ema3 - self.ema4) * self.smoothing_factor;
                self.ema5 += (self.ema4 - self.ema5) * self.smoothing_factor;
                self.sum += self.ema5;

                if (self.length6 == self.count) {
                    self.primed = true;
                    self.ema6 = self.sum / @as(f64, @floatFromInt(self.length));
                    return self.c1 * self.ema6 + self.c2 * self.ema5 + self.c3 * self.ema4 + self.c4 * self.ema3;
                }
            }
        } else {
            // Metastock
            if (self.count == 1) {
                self.ema1 = sample;
            } else if (self.length >= self.count) {
                self.ema1 += (sample - self.ema1) * self.smoothing_factor;
                if (self.length == self.count) {
                    self.ema2 = self.ema1;
                }
            } else if (self.length2 >= self.count) {
                self.ema1 += (sample - self.ema1) * self.smoothing_factor;
                self.ema2 += (self.ema1 - self.ema2) * self.smoothing_factor;

                if (self.length2 == self.count) {
                    self.ema3 = self.ema2;
                }
            } else if (self.length3 >= self.count) {
                self.ema1 += (sample - self.ema1) * self.smoothing_factor;
                self.ema2 += (self.ema1 - self.ema2) * self.smoothing_factor;
                self.ema3 += (self.ema2 - self.ema3) * self.smoothing_factor;

                if (self.length3 == self.count) {
                    self.ema4 = self.ema3;
                }
            } else if (self.length4 >= self.count) {
                self.ema1 += (sample - self.ema1) * self.smoothing_factor;
                self.ema2 += (self.ema1 - self.ema2) * self.smoothing_factor;
                self.ema3 += (self.ema2 - self.ema3) * self.smoothing_factor;
                self.ema4 += (self.ema3 - self.ema4) * self.smoothing_factor;

                if (self.length4 == self.count) {
                    self.ema5 = self.ema4;
                }
            } else if (self.length5 >= self.count) {
                self.ema1 += (sample - self.ema1) * self.smoothing_factor;
                self.ema2 += (self.ema1 - self.ema2) * self.smoothing_factor;
                self.ema3 += (self.ema2 - self.ema3) * self.smoothing_factor;
                self.ema4 += (self.ema3 - self.ema4) * self.smoothing_factor;
                self.ema5 += (self.ema4 - self.ema5) * self.smoothing_factor;

                if (self.length5 == self.count) {
                    self.ema6 = self.ema5;
                }
            } else {
                self.ema1 += (sample - self.ema1) * self.smoothing_factor;
                self.ema2 += (self.ema1 - self.ema2) * self.smoothing_factor;
                self.ema3 += (self.ema2 - self.ema3) * self.smoothing_factor;
                self.ema4 += (self.ema3 - self.ema4) * self.smoothing_factor;
                self.ema5 += (self.ema4 - self.ema5) * self.smoothing_factor;
                self.ema6 += (self.ema5 - self.ema6) * self.smoothing_factor;

                if (self.length6 == self.count) {
                    self.primed = true;
                    return self.c1 * self.ema6 + self.c2 * self.ema5 + self.c3 * self.ema4 + self.c4 * self.ema3;
                }
            }
        }

        return math.nan(f64);
    }

    pub fn isPrimed(self: *const T3ExponentialMovingAverage) bool {
        return self.primed;
    }

    pub fn getMetadata(self: *const T3ExponentialMovingAverage, out: *Metadata) void {
        build_metadata_mod.buildMetadata(
            out,
            .t3_exponential_moving_average,
            self.line.mnemonic,
            self.line.description,
            &[_]build_metadata_mod.OutputText{
                .{ .mnemonic = self.line.mnemonic, .description = self.line.description },
            },
        );
    }

    pub fn updateScalar(self: *T3ExponentialMovingAverage, sample: *const Scalar) OutputArray {
        const value = self.update(sample.value);
        return LineIndicator.wrapScalar(sample.time, value);
    }

    pub fn updateBar(self: *T3ExponentialMovingAverage, sample: *const Bar) OutputArray {
        const value = self.update(self.line.extractBar(sample));
        return LineIndicator.wrapScalar(sample.time, value);
    }

    pub fn updateQuote(self: *T3ExponentialMovingAverage, sample: *const Quote) OutputArray {
        const value = self.update(self.line.extractQuote(sample));
        return LineIndicator.wrapScalar(sample.time, value);
    }

    pub fn updateTrade(self: *T3ExponentialMovingAverage, sample: *const Trade) OutputArray {
        const value = self.update(self.line.extractTrade(sample));
        return LineIndicator.wrapScalar(sample.time, value);
    }

    pub fn indicator(self: *T3ExponentialMovingAverage) indicator_mod.Indicator {
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
        const self: *T3ExponentialMovingAverage = @ptrCast(@alignCast(ptr));
        return self.isPrimed();
    }

    fn vtableMetadata(ptr: *anyopaque, out: *Metadata) void {
        const self: *const T3ExponentialMovingAverage = @ptrCast(@alignCast(ptr));
        self.getMetadata(out);
    }

    fn vtableUpdateScalar(ptr: *anyopaque, sample: *const Scalar) OutputArray {
        const self: *T3ExponentialMovingAverage = @ptrCast(@alignCast(ptr));
        return self.updateScalar(sample);
    }

    fn vtableUpdateBar(ptr: *anyopaque, sample: *const Bar) OutputArray {
        const self: *T3ExponentialMovingAverage = @ptrCast(@alignCast(ptr));
        return self.updateBar(sample);
    }

    fn vtableUpdateQuote(ptr: *anyopaque, sample: *const Quote) OutputArray {
        const self: *T3ExponentialMovingAverage = @ptrCast(@alignCast(ptr));
        return self.updateQuote(sample);
    }

    fn vtableUpdateTrade(ptr: *anyopaque, sample: *const Trade) OutputArray {
        const self: *T3ExponentialMovingAverage = @ptrCast(@alignCast(ptr));
        return self.updateTrade(sample);
    }

    pub const InitError = error{
        InvalidLength,
        InvalidSmoothingFactor,
        InvalidVolumeFactor,
        MnemonicTooLong,
    };
};

// ---------------------------------------------------------------------------
// Tests
// ---------------------------------------------------------------------------

const testing = std.testing;

fn testInput() [252]f64 {
    return .{
        91.500000,  94.815000,  94.375000,  95.095000,  93.780000,  94.625000,  92.530000,  92.750000,  90.315000,  92.470000,
        96.125000,  97.250000,  98.500000,  89.875000,  91.000000,  92.815000,  89.155000,  89.345000,  91.625000,  89.875000,
        88.375000,  87.625000,  84.780000,  83.000000,  83.500000,  81.375000,  84.440000,  89.250000,  86.375000,  86.250000,
        85.250000,  87.125000,  85.815000,  88.970000,  88.470000,  86.875000,  86.815000,  84.875000,  84.190000,  83.875000,
        83.375000,  85.500000,  89.190000,  89.440000,  91.095000,  90.750000,  91.440000,  89.000000,  91.000000,  90.500000,
        89.030000,  88.815000,  84.280000,  83.500000,  82.690000,  84.750000,  85.655000,  86.190000,  88.940000,  89.280000,
        88.625000,  88.500000,  91.970000,  91.500000,  93.250000,  93.500000,  93.155000,  91.720000,  90.000000,  89.690000,
        88.875000,  85.190000,  83.375000,  84.875000,  85.940000,  97.250000,  99.875000,  104.940000, 106.000000, 102.500000,
        102.405000, 104.595000, 106.125000, 106.000000, 106.065000, 104.625000, 108.625000, 109.315000, 110.500000, 112.750000,
        123.000000, 119.625000, 118.750000, 119.250000, 117.940000, 116.440000, 115.190000, 111.875000, 110.595000, 118.125000,
        116.000000, 116.000000, 112.000000, 113.750000, 112.940000, 116.000000, 120.500000, 116.620000, 117.000000, 115.250000,
        114.310000, 115.500000, 115.870000, 120.690000, 120.190000, 120.750000, 124.750000, 123.370000, 122.940000, 122.560000,
        123.120000, 122.560000, 124.620000, 129.250000, 131.000000, 132.250000, 131.000000, 132.810000, 134.000000, 137.380000,
        137.810000, 137.880000, 137.250000, 136.310000, 136.250000, 134.630000, 128.250000, 129.000000, 123.870000, 124.810000,
        123.000000, 126.250000, 128.380000, 125.370000, 125.690000, 122.250000, 119.370000, 118.500000, 123.190000, 123.500000,
        122.190000, 119.310000, 123.310000, 121.120000, 123.370000, 127.370000, 128.500000, 123.870000, 122.940000, 121.750000,
        124.440000, 122.000000, 122.370000, 122.940000, 124.000000, 123.190000, 124.560000, 127.250000, 125.870000, 128.860000,
        132.000000, 130.750000, 134.750000, 135.000000, 132.380000, 133.310000, 131.940000, 130.000000, 125.370000, 130.130000,
        127.120000, 125.190000, 122.000000, 125.000000, 123.000000, 123.500000, 120.060000, 121.000000, 117.750000, 119.870000,
        122.000000, 119.190000, 116.370000, 113.500000, 114.250000, 110.000000, 105.060000, 107.000000, 107.870000, 107.000000,
        107.120000, 107.000000, 91.000000,  93.940000,  93.870000,  95.500000,  93.000000,  94.940000,  98.250000,  96.750000,
        94.810000,  94.370000,  91.560000,  90.250000,  93.940000,  93.620000,  97.000000,  95.000000,  95.870000,  94.060000,
        94.620000,  93.750000,  98.000000,  103.940000, 107.870000, 106.060000, 104.500000, 105.000000, 104.190000, 103.060000,
        103.420000, 105.270000, 111.870000, 116.000000, 116.620000, 118.280000, 113.370000, 109.000000, 109.700000, 109.250000,
        107.000000, 109.190000, 110.000000, 109.200000, 110.120000, 108.000000, 108.620000, 109.750000, 109.810000, 109.000000,
        108.750000, 107.870000,
    };
}

// Expected data from test_T3.xls, T3(5, 0.7) — firstIsAverage = true.
fn testExpected() [252]f64 {
    return .{
        math.nan(f64),      math.nan(f64),      math.nan(f64),      math.nan(f64),      math.nan(f64),      math.nan(f64),
        math.nan(f64),      math.nan(f64),      math.nan(f64),      math.nan(f64),      math.nan(f64),      math.nan(f64),
        math.nan(f64),      math.nan(f64),      math.nan(f64),      math.nan(f64),      math.nan(f64),      math.nan(f64),
        math.nan(f64),      math.nan(f64),      math.nan(f64),      math.nan(f64),      math.nan(f64),      math.nan(f64),
        85.72987567371790,  84.36536238957790,  83.48212833499610,  83.64878022075270,  84.20774780003610,  84.81400226243770,
        85.21162459867090,  85.61977953584720,  85.88366564657920,  86.37507196062170,  86.96360606377290,  87.33715584797100,
        87.47581538320310,  87.22704246772630,  86.66138343888210,  85.93988420535540,  85.17149180117680,  84.72447668908230,
        85.01865520943790,  85.83612757497730,  87.03563064369770,  88.27257526381620,  89.41758978061620,  90.07559145334440,
        90.52788384532180,  90.79849305637170,  90.75816999615070,  90.48012595902110,  89.56721885323770,  88.19961294746030,
        86.64012904487460,  85.41932552383290,  84.73734286979860,  84.55421441417930,  85.00890603797630,  85.87022403441470,
        86.77016654750900,  87.51893615115180,  88.47034581811800,  89.45474018834570,  90.52574408904200,  91.57169411417090,
        92.42614882677630,  92.87068016379410,  92.77362350686670,  92.29390466156780,  91.56310318186090,  90.34706085030460,
        88.70815109494770,  87.17399400070270,  86.08896648247290,  86.77721477010180,  89.01690920285690,  92.50578045524320,
        96.44661245628320,  99.68374118082500,  101.92353915902500, 103.51864520203800, 104.78791133759700, 105.75424271690200,
        106.42923203569800, 106.67331521810400, 107.05684930109700, 107.65552066563500, 108.45831824703700, 109.54712933016300,
        111.89434730682000, 114.52574405349600, 116.75228613125600, 118.41726647291900, 119.37242887772400, 119.58121746519900,
        119.15016873725100, 117.99712010891400, 116.35343432693000, 115.50133109382700, 115.19900149595300, 115.20832852971600,
        114.89045855286400, 114.49505935654600, 114.04873581207800, 113.98299653094000, 114.76120631137200, 115.59884277390800,
        116.28205585323000, 116.55693390132200, 116.39983049399400, 116.13322166220300, 115.93071030281200, 116.38166637287000,
        117.22757950324600, 118.23631535248800, 119.65852250290400, 121.06189350073800, 122.17001559831100, 122.88208870818400,
        123.32075513993400, 123.49845436008200, 123.73526507154800, 124.57544794795000, 125.98775939563200, 127.72677347766200,
        129.27852779306300, 130.66084445385100, 131.92850098151000, 133.38763285852100, 134.90304940294100, 136.26189527105800,
        137.25635611457500, 137.76465314041100, 137.88397712608300, 137.57294795595100, 136.27097486999100, 134.48409991908300,
        132.08334163347200, 129.64905881153300, 127.34920616402000, 125.78806210911800, 125.20525121026800, 124.97471519964800,
        124.91187781656600, 124.53236452951700, 123.58691150647000, 122.26362413973900, 121.44605573567300, 121.22904968429300,
        121.26833012683900, 121.06480537490900, 121.11936168993500, 121.16873672766900, 121.42906776612200, 122.30150235050600,
        123.65994185435500, 124.60369529647400, 124.92167475815800, 124.65108838183800, 124.37460034158500, 123.94458586126200,
        123.47997231186300, 123.13057512793100, 123.03814664421900, 123.04147396438600, 123.23225566873300, 123.84790124323500,
        124.53741993908800, 125.48193174896000, 126.88001089241300, 128.29156445167800, 129.94307874395900, 131.63064617976500,
        132.80966556857300, 133.55644670051600, 133.80913491375100, 133.48924486083300, 132.29305889081000, 131.18723345437400,
        130.07504168974000, 128.82765488259600, 127.23236569221000, 125.90878699219100, 124.77446308473800, 123.92185993031100,
        122.95122696884300, 122.05570316734200, 120.96585380865100, 120.08120106718900, 119.74698117219900, 119.52440496847700,
        118.98986994944600, 117.91920934928700, 116.69443222363700, 115.09004689050000, 112.81604242902500, 110.56231508883200,
        108.80117430219600, 107.52097068017500, 106.69052592372500, 106.21332619981500, 104.12646621492700, 101.28478214721100,
        98.45909805366300,  96.29196542448850,  94.59404389669660,  93.56678784171560,  93.49539073445030,  93.90336034189440,
        94.25829840322930,  94.40922959431940,  94.07888457524260,  93.30020230129970,  92.78070057794470,  92.57408795701960,
        92.98810688816150,  93.56080911108470,  94.18829647483600,  94.55572153094430,  94.73726505383050,  94.70162919926040,
        95.03014981975750,  96.34501640344490,  98.68718019257840,  101.19176302440600, 103.18034967184000, 104.57401565486800,
        105.34469118495800, 105.51082025160400, 105.32911720489200, 105.20951060163700, 105.97637243377600, 107.82138781238700,
        110.21769922957300, 112.77811195271300, 114.51075261504700, 114.85539096796300, 114.29440592244600, 113.27318258016400,
        111.89052001075400, 110.70018738987500, 109.94200614625900, 109.47043806636400, 109.29969950462100, 109.08364080071000,
        108.87294256387700, 108.83218146601300, 108.93829305405100, 109.02476604337200, 109.03210341275800, 108.87915000449300,
    };
}

fn createT3Length(length: usize, first_is_average: bool, volume: f64) !T3ExponentialMovingAverage {
    var t3 = try T3ExponentialMovingAverage.initLength(.{
        .length = length,
        .volume_factor = volume,
        .first_is_average = first_is_average,
    });
    t3.fixSlices();
    return t3;
}

fn createT3Alpha(alpha: f64, first_is_average: bool, volume: f64) !T3ExponentialMovingAverage {
    var t3 = try T3ExponentialMovingAverage.initSmoothingFactor(.{
        .smoothing_factor = alpha,
        .volume_factor = volume,
        .first_is_average = first_is_average,
    });
    t3.fixSlices();
    return t3;
}

test "t3 update length 5 firstIsAverage true (t3.xls)" {
    const input = testInput();
    const exp = testExpected();
    var t3 = try createT3Length(5, true, 0.7);
    const lprimed = 6 * 5 - 6;

    for (0..lprimed) |i| {
        try testing.expect(math.isNan(t3.update(input[i])));
    }

    for (lprimed..252) |i| {
        const act = t3.update(input[i]);
        try testing.expect(@abs(exp[i] - act) < 1e-3);
    }

    try testing.expect(math.isNan(t3.update(math.nan(f64))));
}

test "t3 update length 5 firstIsAverage false (Metastock)" {
    const input = testInput();
    var t3 = try createT3Length(5, false, 0.7);
    const lprimed = 6 * 5 - 6;

    for (0..lprimed) |i| {
        try testing.expect(math.isNan(t3.update(input[i])));
    }

    // Spot-check values from TA-Lib tests.
    for (lprimed..252) |i| {
        const act = t3.update(input[i]);
        switch (i) {
            24 => try testing.expect(@abs(85.749 - act) < 1e-3),
            25 => try testing.expect(@abs(84.380 - act) < 1e-3),
            250 => try testing.expect(@abs(109.032 - act) < 1e-3),
            251 => try testing.expect(@abs(108.88 - act) < 1e-3),
            else => {},
        }
    }

    try testing.expect(math.isNan(t3.update(math.nan(f64))));
}

test "t3 isPrimed length 5" {
    const input = testInput();
    const l = 5;
    const lprimed = 6 * l - 6;

    // firstIsAverage = true
    {
        var t3 = try createT3Length(l, true, 0.7);
        try testing.expect(!t3.isPrimed());
        for (0..lprimed) |i| {
            _ = t3.update(input[i]);
            try testing.expect(!t3.isPrimed());
        }
        for (lprimed..252) |i| {
            _ = t3.update(input[i]);
            try testing.expect(t3.isPrimed());
        }
    }

    // firstIsAverage = false
    {
        var t3 = try createT3Length(l, false, 0.7);
        try testing.expect(!t3.isPrimed());
        for (0..lprimed) |i| {
            _ = t3.update(input[i]);
            try testing.expect(!t3.isPrimed());
        }
        for (lprimed..252) |i| {
            _ = t3.update(input[i]);
            try testing.expect(t3.isPrimed());
        }
    }
}

test "t3 metadata length" {
    var t3 = try createT3Length(10, true, 0.3333);
    var m: Metadata = undefined;
    t3.getMetadata(&m);

    try testing.expectEqual(Identifier.t3_exponential_moving_average, m.identifier);
    try testing.expectEqualStrings("t3(10, 0.33330000)", m.mnemonic);
    try testing.expectEqualStrings("T3 exponential moving average t3(10, 0.33330000)", m.description);
}

test "t3 metadata alpha" {
    const alpha: f64 = 2.0 / 11.0;
    var t3 = try createT3Alpha(alpha, false, 0.3333333);
    var m: Metadata = undefined;
    t3.getMetadata(&m);

    try testing.expectEqual(Identifier.t3_exponential_moving_average, m.identifier);
    try testing.expectEqualStrings("t3(10, 0.18181818, 0.33333330)", m.mnemonic);
    try testing.expectEqualStrings("T3 exponential moving average t3(10, 0.18181818, 0.33333330)", m.description);
}

test "t3 update entity" {
    const inp: f64 = 3.0;
    const exp_false: f64 = 1.6675884773662544;
    const exp_true: f64 = 1.6901728395061721;
    const time: i64 = 1617235200;
    const l = 2;
    const lprimed = 6 * l - 6;

    // scalar
    {
        var t3 = try createT3Length(l, false, 0.7);
        for (0..lprimed) |_| {
            _ = t3.update(0.0);
        }
        const out = t3.updateScalar(&.{ .time = time, .value = inp });
        try testing.expectEqual(@as(usize, 1), out.len);
        const s = out.slice()[0].scalar;
        try testing.expectEqual(time, s.time);
        try testing.expect(@abs(exp_false - s.value) < 1e-13);
    }

    // bar
    {
        var t3 = try createT3Length(l, true, 0.7);
        for (0..lprimed) |_| {
            _ = t3.update(0.0);
        }
        const bar = Bar{ .time = time, .open = 0, .high = 0, .low = 0, .close = inp, .volume = 0 };
        const out = t3.updateBar(&bar);
        const s = out.slice()[0].scalar;
        try testing.expect(@abs(exp_true - s.value) < 1e-13);
    }
}

test "t3 init invalid" {
    // length < 2
    try testing.expectError(error.InvalidLength, T3ExponentialMovingAverage.initLength(.{ .length = 1 }));
    try testing.expectError(error.InvalidLength, T3ExponentialMovingAverage.initLength(.{ .length = 0 }));

    // alpha out of range
    try testing.expectError(error.InvalidSmoothingFactor, T3ExponentialMovingAverage.initSmoothingFactor(.{ .smoothing_factor = -1.0 }));
    try testing.expectError(error.InvalidSmoothingFactor, T3ExponentialMovingAverage.initSmoothingFactor(.{ .smoothing_factor = 2.0 }));

    // volume factor out of range
    try testing.expectError(error.InvalidVolumeFactor, T3ExponentialMovingAverage.initLength(.{ .length = 5, .volume_factor = -0.7 }));
    try testing.expectError(error.InvalidVolumeFactor, T3ExponentialMovingAverage.initLength(.{ .length = 5, .volume_factor = 1.7 }));
}

test "t3 mnemonic with bar component" {
    var t3 = try T3ExponentialMovingAverage.initLength(.{
        .length = 10,
        .volume_factor = 0.7,
        .bar_component = .median,
    });
    t3.fixSlices();
    try testing.expectEqualStrings("t3(10, 0.70000000, hl/2)", t3.line.mnemonic);
    try testing.expectEqualStrings("T3 exponential moving average t3(10, 0.70000000, hl/2)", t3.line.description);
}

test "t3 alpha with quote component" {
    var t3 = try T3ExponentialMovingAverage.initSmoothingFactor(.{
        .smoothing_factor = 2.0 / 11.0,
        .volume_factor = 0.7,
        .quote_component = .bid,
    });
    t3.fixSlices();
    try testing.expectEqualStrings("t3(10, 0.18181818, 0.70000000, b)", t3.line.mnemonic);
}
