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

/// Enumerates the outputs of the T2 exponential moving average indicator.
pub const T2ExponentialMovingAverageOutput = enum(u8) {
    value = 1,
};

/// Parameters to create an instance of the T2 EMA based on length.
pub const T2ExponentialMovingAverageLengthParams = struct {
    length: usize,
    volume_factor: f64 = 0.7,
    first_is_average: bool = false,
    bar_component: ?bar_component.BarComponent = null,
    quote_component: ?quote_component.QuoteComponent = null,
    trade_component: ?trade_component.TradeComponent = null,
};

/// Parameters to create an instance of the T2 EMA based on smoothing factor.
pub const T2ExponentialMovingAverageSmoothingFactorParams = struct {
    smoothing_factor: f64,
    volume_factor: f64 = 0.7,
    first_is_average: bool = false,
    bar_component: ?bar_component.BarComponent = null,
    quote_component: ?quote_component.QuoteComponent = null,
    trade_component: ?trade_component.TradeComponent = null,
};

/// Computes the T2 Exponential Moving Average.
///
/// T2 is a four-pole non-linear Kalman filter developed by Tim Tillson.
///
///   v^2 * EMA4 - 2v(1+v) * EMA3 + (1+v)^2 * EMA2
///
/// Warmup period: 4*length - 4 samples before primed.
pub const T2ExponentialMovingAverage = struct {
    line: LineIndicator,
    smoothing_factor: f64,
    c1: f64,
    c2: f64,
    c3: f64,
    sum: f64,
    ema1: f64,
    ema2: f64,
    ema3: f64,
    ema4: f64,
    length: usize,
    length2: usize,
    length3: usize,
    length4: usize,
    count: usize,
    first_is_average: bool,
    primed: bool,
    mnemonic_buf: [96]u8,
    mnemonic_len: usize,
    description_buf: [160]u8,
    description_len: usize,

    const epsilon: f64 = 0.00000001;

    /// Create T2 EMA from length.
    pub fn initLength(params: T2ExponentialMovingAverageLengthParams) !T2ExponentialMovingAverage {
        if (params.length < 2) {
            return error.InvalidLength;
        }
        if (params.volume_factor < 0.0 or params.volume_factor > 1.0) {
            return error.InvalidVolumeFactor;
        }
        const alpha = 2.0 / @as(f64, @floatFromInt(1 + params.length));
        return initInternal(params.length, alpha, false, params.volume_factor, params.first_is_average, params.bar_component, params.quote_component, params.trade_component);
    }

    /// Create T2 EMA from smoothing factor.
    pub fn initSmoothingFactor(params: T2ExponentialMovingAverageSmoothingFactorParams) !T2ExponentialMovingAverage {
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
    ) !T2ExponentialMovingAverage {
        const bc = bc_opt orelse bar_component.default_bar_component;
        const qc = qc_opt orelse quote_component.default_quote_component;
        const tc = tc_opt orelse trade_component.default_trade_component;

        var triple_buf: [64]u8 = undefined;
        const triple = component_triple_mnemonic_mod.componentTripleMnemonic(&triple_buf, bc, qc, tc);

        var mnemonic_buf: [96]u8 = undefined;
        var mnemonic_slice: []u8 = undefined;
        if (is_alpha_mode) {
            mnemonic_slice = std.fmt.bufPrint(&mnemonic_buf, "t2({d}, {d:.8}, {d:.8}{s})", .{ length, alpha, v, triple }) catch
                return error.MnemonicTooLong;
        } else {
            mnemonic_slice = std.fmt.bufPrint(&mnemonic_buf, "t2({d}, {d:.8}{s})", .{ length, v, triple }) catch
                return error.MnemonicTooLong;
        }
        const mnemonic_len = mnemonic_slice.len;

        var description_buf: [160]u8 = undefined;
        const desc_slice = std.fmt.bufPrint(&description_buf, "T2 exponential moving average {s}", .{mnemonic_slice}) catch
            return error.MnemonicTooLong;
        const description_len = desc_slice.len;

        const v1 = v + 1.0;
        const c1 = v * v;
        const c2 = -2.0 * v * v1;
        const c3 = v1 * v1;

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
            .sum = 0.0,
            .ema1 = 0.0,
            .ema2 = 0.0,
            .ema3 = 0.0,
            .ema4 = 0.0,
            .length = length,
            .length2 = 2 * length - 1,
            .length3 = 3 * length - 2,
            .length4 = 4 * length - 3,
            .count = 0,
            .first_is_average = first_is_average,
            .primed = false,
            .mnemonic_buf = mnemonic_buf,
            .mnemonic_len = mnemonic_len,
            .description_buf = description_buf,
            .description_len = description_len,
        };
    }

    pub fn fixSlices(self: *T2ExponentialMovingAverage) void {
        self.line.mnemonic = self.mnemonic_buf[0..self.mnemonic_len];
        self.line.description = self.description_buf[0..self.description_len];
    }

    pub fn update(self: *T2ExponentialMovingAverage, sample: f64) f64 {
        if (math.isNan(sample)) {
            return sample;
        }

        if (self.primed) {
            var v1 = self.ema1;
            var v2 = self.ema2;
            var v3 = self.ema3;
            var v4 = self.ema4;
            const sf = self.smoothing_factor;
            v1 += (sample - v1) * sf;
            v2 += (v1 - v2) * sf;
            v3 += (v2 - v3) * sf;
            v4 += (v3 - v4) * sf;
            self.ema1 = v1;
            self.ema2 = v2;
            self.ema3 = v3;
            self.ema4 = v4;
            return self.c1 * v4 + self.c2 * v3 + self.c3 * v2;
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
            } else {
                self.ema1 += (sample - self.ema1) * self.smoothing_factor;
                self.ema2 += (self.ema1 - self.ema2) * self.smoothing_factor;
                self.ema3 += (self.ema2 - self.ema3) * self.smoothing_factor;
                self.sum += self.ema3;

                if (self.length4 == self.count) {
                    self.primed = true;
                    self.ema4 = self.sum / @as(f64, @floatFromInt(self.length));
                    return self.c1 * self.ema4 + self.c2 * self.ema3 + self.c3 * self.ema2;
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
            } else {
                self.ema1 += (sample - self.ema1) * self.smoothing_factor;
                self.ema2 += (self.ema1 - self.ema2) * self.smoothing_factor;
                self.ema3 += (self.ema2 - self.ema3) * self.smoothing_factor;
                self.ema4 += (self.ema3 - self.ema4) * self.smoothing_factor;

                if (self.length4 == self.count) {
                    self.primed = true;
                    return self.c1 * self.ema4 + self.c2 * self.ema3 + self.c3 * self.ema2;
                }
            }
        }

        return math.nan(f64);
    }

    pub fn isPrimed(self: *const T2ExponentialMovingAverage) bool {
        return self.primed;
    }

    pub fn getMetadata(self: *const T2ExponentialMovingAverage, out: *Metadata) void {
        build_metadata_mod.buildMetadata(
            out,
            .t2_exponential_moving_average,
            self.line.mnemonic,
            self.line.description,
            &[_]build_metadata_mod.OutputText{
                .{ .mnemonic = self.line.mnemonic, .description = self.line.description },
            },
        );
    }

    pub fn updateScalar(self: *T2ExponentialMovingAverage, sample: *const Scalar) OutputArray {
        const value = self.update(sample.value);
        return LineIndicator.wrapScalar(sample.time, value);
    }

    pub fn updateBar(self: *T2ExponentialMovingAverage, sample: *const Bar) OutputArray {
        const value = self.update(self.line.extractBar(sample));
        return LineIndicator.wrapScalar(sample.time, value);
    }

    pub fn updateQuote(self: *T2ExponentialMovingAverage, sample: *const Quote) OutputArray {
        const value = self.update(self.line.extractQuote(sample));
        return LineIndicator.wrapScalar(sample.time, value);
    }

    pub fn updateTrade(self: *T2ExponentialMovingAverage, sample: *const Trade) OutputArray {
        const value = self.update(self.line.extractTrade(sample));
        return LineIndicator.wrapScalar(sample.time, value);
    }

    pub fn indicator(self: *T2ExponentialMovingAverage) indicator_mod.Indicator {
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
        const self: *T2ExponentialMovingAverage = @ptrCast(@alignCast(ptr));
        return self.isPrimed();
    }

    fn vtableMetadata(ptr: *anyopaque, out: *Metadata) void {
        const self: *const T2ExponentialMovingAverage = @ptrCast(@alignCast(ptr));
        self.getMetadata(out);
    }

    fn vtableUpdateScalar(ptr: *anyopaque, sample: *const Scalar) OutputArray {
        const self: *T2ExponentialMovingAverage = @ptrCast(@alignCast(ptr));
        return self.updateScalar(sample);
    }

    fn vtableUpdateBar(ptr: *anyopaque, sample: *const Bar) OutputArray {
        const self: *T2ExponentialMovingAverage = @ptrCast(@alignCast(ptr));
        return self.updateBar(sample);
    }

    fn vtableUpdateQuote(ptr: *anyopaque, sample: *const Quote) OutputArray {
        const self: *T2ExponentialMovingAverage = @ptrCast(@alignCast(ptr));
        return self.updateQuote(sample);
    }

    fn vtableUpdateTrade(ptr: *anyopaque, sample: *const Trade) OutputArray {
        const self: *T2ExponentialMovingAverage = @ptrCast(@alignCast(ptr));
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

// Expected data from test_T2.xls, T2(5, 0.7) — firstIsAverage = true.
fn testExpected() [252]f64 {
    return .{
        math.nan(f64),      math.nan(f64),      math.nan(f64),      math.nan(f64),      math.nan(f64),      math.nan(f64),      math.nan(f64),      math.nan(f64),
        math.nan(f64),      math.nan(f64),      math.nan(f64),      math.nan(f64),      math.nan(f64),      math.nan(f64),      math.nan(f64),      math.nan(f64),
        92.00445682439470,  90.91065523008910,  90.64635861170230,  90.30847892058210,  89.67203184711450,  88.90373302682710,  87.58685794329080,  85.95483674752900,
        84.70056892267930,  83.35230597480420,  83.01639757937170,  84.26606438522640,  85.11761359114740,  85.64273951236740,  85.71656813204080,  86.07519575846890,
        86.14040917330100,  86.84731118619480,  87.49707083315330,  87.60487694128920,  87.48806230196600,  86.85037109765170,  86.00510559137550,  85.19270124163530,
        84.44671220986690,  84.39562430058820,  85.47846964605910,  86.77863391237980,  88.27480023486270,  89.43282744282170,  90.38671380457610,  90.47144856898390,
        90.70844123290330,  90.80974940198440,  90.47870378112050,  90.01014971788120,  88.50241981733770,  86.77695943257320,  85.15371443592010,  84.41215046544750,
        84.35591075095750,  84.68800316786790,  85.75412809581360,  86.93319023556810,  87.74938215169110,  88.24027778265170,  89.34170209097540,  90.29228263445890,
        91.39771790246850,  92.36534535927510,  92.98185469287250,  92.98804953756390,  92.35873030252150,  91.56001444442250,  90.66809450828110,  89.04851123746310,
        87.09681376947570,  85.81489615554240,  85.29909582783180,  87.87278112697240,  91.57930166956580,  96.09718145905370,  100.17004554905500, 102.36263587267900,
        103.39235683035700, 104.28507155830600, 105.25701353369700, 105.97310800537900, 106.42554279253100, 106.31201565835700, 106.95673692587200, 107.85537745407500,
        108.90779472362600, 110.30293857497500, 113.94246316874900, 116.72555020318500, 118.38770601306100, 119.39753212263300, 119.61486726510600, 119.11850573798600,
        118.15216910748100, 116.41234112714800, 114.45865964016000, 114.64061162124700, 114.99246753582000, 115.33333018479700, 114.64625300711600, 114.18195589036400,
        113.69963506110000, 114.05025810623800, 115.68652710013500, 116.48211079384600, 116.91736166744200, 116.71412173592500, 116.10716021480600, 115.77116121989300,
        115.68159156874600, 116.84472084196100, 118.05979798379400, 119.16496881132300, 120.94593869721900, 122.23096010279300, 122.95608471296700, 123.22656209567300,
        123.39820574539500, 123.34529134741000, 123.70294192360900, 125.21356151217400, 127.19355195747100, 129.19266295422300, 130.45262728011500, 131.60256664339700,
        132.72482116981200, 134.37601221394000, 135.92461267673900, 137.09424525891600, 137.70388483421700, 137.73781691122000, 137.52589963778700, 136.86260850470000,
        134.64367631336900, 132.54215730025200, 129.69729772115400, 127.44287309781500, 125.45301927369100, 124.78149677056900, 125.23737028614000, 125.28834765521900,
        125.29811473230200, 124.49118133074600, 122.91167204427900, 121.21791414741600, 120.98626501620100, 121.39013089649900, 121.63114611836900, 121.09667781630300,
        121.38604904393200, 121.37683602700200, 121.81714963968400, 123.27680991970300, 125.04065301695500, 125.42231478992600, 125.02871976565900, 124.16184698358900,
        123.95348210659900, 123.42096848684700, 122.98054879392000, 122.80175403486300, 123.00506164290600, 123.10341273782800, 123.47654621169500, 124.50686275946200,
        125.20780013410700, 126.36056179556500, 128.16193338760800, 129.48687497042000, 131.31008039649500, 132.94986405572100, 133.52752016483300, 133.82004244574000,
        133.59810463804000, 132.76202894406400, 130.77137039821000, 129.94273481747400, 128.95593356768700, 127.66988673048200, 125.80391233666600, 124.87150365869300,
        123.99029524198700, 123.45575776078900, 122.35139824579100, 121.53836332656800, 120.24081578533100, 119.58685232736100, 119.81980247526100, 119.65181893967000,
        118.73183613794600, 117.08643352117200, 115.74380263433200, 113.79159129198400, 110.89385419602800, 108.78809510901800, 107.63783653194000, 106.88467024071700,
        106.49446844969200, 106.31788640074800, 102.45936340772700, 98.92286025711640,  96.26320146598560,  94.89830130673550,  93.71214241680890,  93.35493402958790,
        94.19091222655660,  94.93222890927330,  95.05352335355920,  94.85884233631840,  93.94374172767060,  92.68201046821450,  92.46778442700140,  92.58855930337340,
        93.61995692101500,  94.27141015939870,  94.87276841707030,  94.89831480519190,  94.85776894321340,  94.59245060865460,  95.32163415260000,  97.64623243404460,
        100.94595782878400, 103.42636980462600, 104.70917474084900, 105.40458887459900, 105.52523492436600, 105.12339759255700, 104.68875644984300, 104.75482770270600,
        106.53547902116200, 109.48515731469700, 112.33128725513400, 114.90966925830500, 115.64161420440900, 114.51441441748600, 113.13113208320500, 111.81663297883600,
        110.24787839832700, 109.44340073972700, 109.26032936270900, 109.13994050792700, 109.29451387663700, 109.00897104476500, 108.79138478114900, 108.91808951962300,
        109.15266693557400, 109.18435708789600, 109.08448284304800, 108.75304884708700,
    };
}

fn createT2Length(length: usize, first_is_average: bool, volume: f64) !T2ExponentialMovingAverage {
    var t2 = try T2ExponentialMovingAverage.initLength(.{
        .length = length,
        .volume_factor = volume,
        .first_is_average = first_is_average,
    });
    t2.fixSlices();
    return t2;
}

fn createT2Alpha(alpha: f64, first_is_average: bool, volume: f64) !T2ExponentialMovingAverage {
    var t2 = try T2ExponentialMovingAverage.initSmoothingFactor(.{
        .smoothing_factor = alpha,
        .volume_factor = volume,
        .first_is_average = first_is_average,
    });
    t2.fixSlices();
    return t2;
}

test "t2 update length 5 firstIsAverage true (t2.xls)" {
    const input = testInput();
    const exp = testExpected();
    var t2 = try createT2Length(5, true, 0.7);
    const lprimed = 4 * 5 - 4;

    for (0..lprimed) |i| {
        try testing.expect(math.isNan(t2.update(input[i])));
    }

    for (lprimed..252) |i| {
        const act = t2.update(input[i]);
        try testing.expect(@abs(exp[i] - act) < 1e-8);
    }

    try testing.expect(math.isNan(t2.update(math.nan(f64))));
}

test "t2 update length 5 firstIsAverage false (Metastock)" {
    const input = testInput();
    const exp = testExpected();
    var t2 = try createT2Length(5, false, 0.7);
    const lprimed = 4 * 5 - 4;
    const first_check = lprimed + 43;

    for (0..lprimed) |i| {
        try testing.expect(math.isNan(t2.update(input[i])));
    }

    for (lprimed..252) |i| {
        const act = t2.update(input[i]);
        if (i >= first_check) {
            try testing.expect(@abs(exp[i] - act) < 1e-8);
        }
    }

    try testing.expect(math.isNan(t2.update(math.nan(f64))));
}

test "t2 isPrimed length 5" {
    const input = testInput();
    const l = 5;
    const lprimed = 4 * l - 4;

    // firstIsAverage = true
    {
        var t2 = try createT2Length(l, true, 0.7);
        try testing.expect(!t2.isPrimed());
        for (0..lprimed) |i| {
            _ = t2.update(input[i]);
            try testing.expect(!t2.isPrimed());
        }
        for (lprimed..252) |i| {
            _ = t2.update(input[i]);
            try testing.expect(t2.isPrimed());
        }
    }

    // firstIsAverage = false
    {
        var t2 = try createT2Length(l, false, 0.7);
        try testing.expect(!t2.isPrimed());
        for (0..lprimed) |i| {
            _ = t2.update(input[i]);
            try testing.expect(!t2.isPrimed());
        }
        for (lprimed..252) |i| {
            _ = t2.update(input[i]);
            try testing.expect(t2.isPrimed());
        }
    }
}

test "t2 metadata length" {
    var t2 = try createT2Length(10, true, 0.3333);
    var m: Metadata = undefined;
    t2.getMetadata(&m);

    try testing.expectEqual(Identifier.t2_exponential_moving_average, m.identifier);
    try testing.expectEqualStrings("t2(10, 0.33330000)", m.mnemonic);
    try testing.expectEqualStrings("T2 exponential moving average t2(10, 0.33330000)", m.description);
}

test "t2 metadata alpha" {
    const alpha: f64 = 2.0 / 11.0;
    var t2 = try createT2Alpha(alpha, false, 0.3333333);
    var m: Metadata = undefined;
    t2.getMetadata(&m);

    try testing.expectEqual(Identifier.t2_exponential_moving_average, m.identifier);
    try testing.expectEqualStrings("t2(10, 0.18181818, 0.33333330)", m.mnemonic);
    try testing.expectEqualStrings("T2 exponential moving average t2(10, 0.18181818, 0.33333330)", m.description);
}

test "t2 update entity" {
    const inp: f64 = 3.0;
    const exp_false: f64 = 2.0281481481481483;
    const exp_true: f64 = 1.9555555555555555;
    const time: i64 = 1617235200;
    const l = 2;
    const lprimed = 4 * l - 4;

    // scalar
    {
        var t2 = try createT2Length(l, false, 0.7);
        for (0..lprimed) |_| {
            _ = t2.update(0.0);
        }
        const out = t2.updateScalar(&.{ .time = time, .value = inp });
        try testing.expectEqual(@as(usize, 1), out.len);
        const s = out.slice()[0].scalar;
        try testing.expectEqual(time, s.time);
        try testing.expect(@abs(exp_false - s.value) < 1e-13);
    }

    // bar
    {
        var t2 = try createT2Length(l, true, 0.7);
        for (0..lprimed) |_| {
            _ = t2.update(0.0);
        }
        const bar = Bar{ .time = time, .open = 0, .high = 0, .low = 0, .close = inp, .volume = 0 };
        const out = t2.updateBar(&bar);
        const s = out.slice()[0].scalar;
        try testing.expect(@abs(exp_true - s.value) < 1e-13);
    }
}

test "t2 init invalid" {
    // length < 2
    try testing.expectError(error.InvalidLength, T2ExponentialMovingAverage.initLength(.{ .length = 1 }));
    try testing.expectError(error.InvalidLength, T2ExponentialMovingAverage.initLength(.{ .length = 0 }));

    // alpha out of range
    try testing.expectError(error.InvalidSmoothingFactor, T2ExponentialMovingAverage.initSmoothingFactor(.{ .smoothing_factor = -1.0 }));
    try testing.expectError(error.InvalidSmoothingFactor, T2ExponentialMovingAverage.initSmoothingFactor(.{ .smoothing_factor = 2.0 }));

    // volume factor out of range
    try testing.expectError(error.InvalidVolumeFactor, T2ExponentialMovingAverage.initLength(.{ .length = 5, .volume_factor = -0.7 }));
    try testing.expectError(error.InvalidVolumeFactor, T2ExponentialMovingAverage.initLength(.{ .length = 5, .volume_factor = 1.7 }));
}

test "t2 mnemonic with bar component" {
    var t2 = try T2ExponentialMovingAverage.initLength(.{
        .length = 10,
        .volume_factor = 0.7,
        .bar_component = .median,
    });
    t2.fixSlices();
    try testing.expectEqualStrings("t2(10, 0.70000000, hl/2)", t2.line.mnemonic);
    try testing.expectEqualStrings("T2 exponential moving average t2(10, 0.70000000, hl/2)", t2.line.description);
}

test "t2 alpha with quote component" {
    var t2 = try T2ExponentialMovingAverage.initSmoothingFactor(.{
        .smoothing_factor = 2.0 / 11.0,
        .volume_factor = 0.7,
        .quote_component = .bid,
    });
    t2.fixSlices();
    try testing.expectEqualStrings("t2(10, 0.18181818, 0.70000000, b)", t2.line.mnemonic);
}
