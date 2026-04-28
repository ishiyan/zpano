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

/// Enumerates the outputs of the super smoother indicator.
pub const SuperSmootherOutput = enum(u8) {
    value = 1,
};

/// Parameters to create an instance of the Super Smoother.
pub const SuperSmootherParams = struct {
    /// Shortest cycle period in bars. Must be >= 2. Default is 10.
    shortest_cycle_period: i32 = 10,
    bar_component: ?bar_component.BarComponent = null,
    quote_component: ?quote_component.QuoteComponent = null,
    trade_component: ?trade_component.TradeComponent = null,
};

/// Ehler's two-pole Super Smoother (SS).
///
///   β = √2·π / λ
///   α = exp(-β)
///   γ₂ = 2α·cos(β)
///   γ₃ = -α²
///   γ₁ = (1 - γ₂ - γ₃) / 2
///
///   SSᵢ = γ₁·(xᵢ + xᵢ₋₁) + γ₂·SSᵢ₋₁ + γ₃·SSᵢ₋₂
///
/// The indicator is not primed during the first 2 updates.
pub const SuperSmoother = struct {
    line: LineIndicator,
    coeff1: f64,
    coeff2: f64,
    coeff3: f64,
    count: usize,
    sample_previous: f64,
    filter_previous: f64,
    filter_previous2: f64,
    value: f64,
    primed: bool,
    mnemonic_buf: [96]u8,
    mnemonic_len: usize,
    description_buf: [160]u8,
    description_len: usize,

    pub fn init(params: SuperSmootherParams) !SuperSmoother {
        const period = params.shortest_cycle_period;
        if (period < 2) {
            return error.InvalidPeriod;
        }

        // Calculate coefficients.
        const period_f: f64 = @floatFromInt(period);
        const beta = math.sqrt2 * math.pi / period_f;
        const alpha = @exp(-beta);
        const gamma2 = 2.0 * alpha * @cos(beta);
        const gamma3 = -alpha * alpha;
        const gamma1 = (1.0 - gamma2 - gamma3) / 2.0;

        // Build mnemonic.
        var triple_buf: [64]u8 = undefined;
        const triple = component_triple_mnemonic_mod.componentTripleMnemonic(
            &triple_buf,
            params.bar_component orelse bar_component.BarComponent.median,
            params.quote_component orelse quote_component.default_quote_component,
            params.trade_component orelse trade_component.default_trade_component,
        );

        var mnemonic_buf: [96]u8 = undefined;
        const mnemonic_slice = std.fmt.bufPrint(&mnemonic_buf, "ss({d}{s})", .{ period, triple }) catch
            return error.MnemonicTooLong;
        const mnemonic_len = mnemonic_slice.len;

        var description_buf: [160]u8 = undefined;
        const desc_slice = std.fmt.bufPrint(&description_buf, "Super Smoother {s}", .{mnemonic_slice}) catch
            return error.MnemonicTooLong;
        const description_len = desc_slice.len;

        return .{
            .line = LineIndicator.new(
                mnemonic_buf[0..mnemonic_len],
                description_buf[0..description_len],
                params.bar_component orelse bar_component.BarComponent.median,
                params.quote_component orelse quote_component.default_quote_component,
                params.trade_component orelse trade_component.default_trade_component,
            ),
            .coeff1 = gamma1,
            .coeff2 = gamma2,
            .coeff3 = gamma3,
            .count = 0,
            .sample_previous = 0.0,
            .filter_previous = 0.0,
            .filter_previous2 = 0.0,
            .value = math.nan(f64),
            .primed = false,
            .mnemonic_buf = mnemonic_buf,
            .mnemonic_len = mnemonic_len,
            .description_buf = description_buf,
            .description_len = description_len,
        };
    }

    pub fn fixSlices(self: *SuperSmoother) void {
        self.line.mnemonic = self.mnemonic_buf[0..self.mnemonic_len];
        self.line.description = self.description_buf[0..self.description_len];
    }

    /// Update the super smoother given the next sample.
    pub fn update(self: *SuperSmoother, sample: f64) f64 {
        if (math.isNan(sample)) {
            return sample;
        }

        if (self.primed) {
            const filter = self.coeff1 * (sample + self.sample_previous) +
                self.coeff2 * self.filter_previous + self.coeff3 * self.filter_previous2;
            self.value = filter;
            self.sample_previous = sample;
            self.filter_previous2 = self.filter_previous;
            self.filter_previous = filter;
            return self.value;
        }

        self.count += 1;

        if (self.count == 1) {
            self.sample_previous = sample;
            self.filter_previous = sample;
            self.filter_previous2 = sample;
        }

        const filter = self.coeff1 * (sample + self.sample_previous) +
            self.coeff2 * self.filter_previous + self.coeff3 * self.filter_previous2;

        if (self.count == 3) {
            self.primed = true;
            self.value = filter;
        }

        self.sample_previous = sample;
        self.filter_previous2 = self.filter_previous;
        self.filter_previous = filter;

        return self.value;
    }

    pub fn isPrimed(self: *const SuperSmoother) bool {
        return self.primed;
    }

    pub fn getMetadata(self: *const SuperSmoother, out: *Metadata) void {
        build_metadata_mod.buildMetadata(
            out,
            .super_smoother,
            self.line.mnemonic,
            self.line.description,
            &[_]build_metadata_mod.OutputText{
                .{ .mnemonic = self.line.mnemonic, .description = self.line.description },
            },
        );
    }

    pub fn updateScalar(self: *SuperSmoother, sample: *const Scalar) OutputArray {
        const value = self.update(sample.value);
        return LineIndicator.wrapScalar(sample.time, value);
    }

    pub fn updateBar(self: *SuperSmoother, sample: *const Bar) OutputArray {
        const value = self.update(self.line.extractBar(sample));
        return LineIndicator.wrapScalar(sample.time, value);
    }

    pub fn updateQuote(self: *SuperSmoother, sample: *const Quote) OutputArray {
        const value = self.update(self.line.extractQuote(sample));
        return LineIndicator.wrapScalar(sample.time, value);
    }

    pub fn updateTrade(self: *SuperSmoother, sample: *const Trade) OutputArray {
        const value = self.update(self.line.extractTrade(sample));
        return LineIndicator.wrapScalar(sample.time, value);
    }

    /// Returns an Indicator interface backed by this instance.
    pub fn indicator(self: *SuperSmoother) indicator_mod.Indicator {
        return .{
            .ptr = @ptrCast(self),
            .vtable = &vtable,
        };
    }

    const vtable = indicator_mod.Indicator.GenVTable(SuperSmoother);
};

// --- Tests ---
const testing = std.testing;

fn almostEqual(a: f64, b: f64, tolerance: f64) bool {
    return @abs(a - b) <= tolerance;
}

const test_input = [_]f64{
    1065.25, 1065.25, 1063.75, 1059.25, 1059.25, 1057.75, 1054,    1056.25, 1058.5,  1059.5,
    1064.75, 1063,    1062.5,  1065,    1061.5,  1058.25, 1058.25, 1061.75, 1062,    1061.25,
    1062.5,  1066.5,  1066.5,  1069.25, 1074.75, 1075,    1076,    1078,    1079.25, 1079.75,
    1078,    1078.75, 1078.25, 1076.5,  1075.75, 1075.75, 1075,    1073.25, 1071,    1083,
    1082.25, 1084,    1085.75, 1085.25, 1085.75, 1087.25, 1089,    1089,    1090,    1095,
    1097.25, 1097.25, 1099,    1098.25, 1093.75, 1095,    1097.25, 1099.25, 1097.5,  1096,
    1095,    1094,    1095.75, 1095.75, 1093.75, 1100.5,  1102.25, 1102,    1102.75, 1105.75,
    1108.25, 1109.5,  1107.25, 1102.5,  1104.75, 1099.25, 1102.75, 1099.5,  1096.75, 1098.25,
    1095.25, 1097,    1097.75, 1100.5,  1099.5,  1101.75, 1101.75, 1102.75, 1099.75, 1097,
    1100.75, 1105.75, 1104.5,  1108.5,  1111.25, 1112.25, 1110,    1109.75, 1108.25, 1106,
    1108.5,  1106.75, 1108,    1106.5,  1105.25, 1104.25, 1102,    1102.5,  1103.25, 1104,
    1104,    1102.5,  1101,    1099.5,  1100,    1100.25, 1103,    1103,    1104.5,  1108.25,
    1110.75, 1107.25, 1107.5,  1105.25, 1103,    1101.75, 1101.75, 1096.5,  1099.5,  1093.75,
    1094.5,  1090.75, 1093.25, 1091.75, 1094.75, 1093.75, 1092.25, 1091.5,  1092.5,  1089.75,
    1089.75, 1090,    1087.75, 1093.75, 1095.25, 1097,    1094.75, 1093,    1094.25, 1092.25,
    1093,    1095.25, 1096,    1094.25, 1092.25, 1094.5,  1092.75, 1094.5,  1098.25, 1097.75,
    1097.5,  1097.5,  1098.25, 1097.75, 1097.5,  1101.5,  1102.75, 1104.5,  1103.25, 1100,
    1101.5,  1097,    1098.75, 1100,    1098.25, 1102,    1103.75, 1103.75, 1102.75, 1101.75,
    1100.5,  1096.25, 1100.5,  1104,    1105.25, 1104.25, 1102.25, 1104,    1103.25, 1104.25,
    1100,    1097.75, 1099.5,  1102,    1104.75, 1102.75, 1104.5,  1104.25, 1105.75, 1107.5,
    1107,    1106.5,  1108,    1107.75, 1106.5,  1107,    1108.5,  1109.25, 1123.75, 1124.25,
    1123.75, 1123,    1121.25, 1120.25, 1118.75, 1119,    1119.75, 1109.75, 1109.75, 1115.75,
    1114.75, 1120.5,  1119.25, 1118.25, 1121.5,  1121.25, 1120,    1120.5,  1121.25, 1123,
    1122.25, 1122,    1122,    1121.5,  1123.25, 1123,    1126,    1127.25, 1129.5,  1129.5,
    1129.25, 1130,    1130.5,  1130.25, 1128.75, 1128,    1132.5,  1129.25, 1122.5,  1119.25,
    1120.25, 1120,    1121.75, 1122,    1120.75, 1120.25, 1116,    1119.75, 1119.25, 1122.75,
    1114.25, 1117,    1117.75, 1120,    1119,    1122,    1120.5,  1118.25, 1119,    1120.25,
    1119.5,  1120,    1119.5,  1113,    1112.75, 1114.5,  1116,    1116.25, 1114.25, 1114.75,
    1115.25, 1115.75, 1116.25, 1114.75, 1113.5,  1115.75, 1121.25, 1120.25, 1119.25, 1118.25,
    1118.25, 1119.25, 1117.5,  1117.25, 1116.5,  1117,    1117,    1116.75, 1116.75, 1120.25,
    1119.75, 1120.25, 1117.5,  1117.25, 1115.5,  1110.75, 1112.5,  1111.25, 1112.75, 1117.75,
    1119.75, 1118.75, 1111,    1113,    1114,    1114.25, 1114,    1113.75, 1114.25, 1112.75,
    1114.5,  1115.25, 1116.5,  1114.75, 1113,    1117.75, 1119.5,  1120,    1115.25, 1115.25,
    1115.5,  1114.25, 1114.25, 1112.75, 1110.75, 1104.75, 1099.25, 1102.25, 1095.25, 1095.25,
    1094.25, 1095.75, 1095.5,  1094.75, 1092.25, 1086.25, 1084.75, 1083.75, 1083.25, 1088.75,
    1089.75, 1089.5,  1093.5,  1093,    1096.5,  1095.75, 1095.75, 1096.25, 1094,    1096.75,
    1099.75, 1102,    1103.5,  1102.5,  1107.75, 1109.25, 1111.5,  1111,    1113.5,  1115,
    1114.25, 1115,    1115,    1114.5,  1117.75, 1118.5,  1120.5,  1117.5,  1117,    1115.75,
    1115,    1114.5,  1114,    1116.25, 1114,    1113.75, 1115.75, 1115.75, 1118,    1116,
    1095.75, 1097.75, 1102.25, 1098.25, 1100,    1098,    1097.25, 1093.25, 1093.5,  1091,
    1098,    1100.5,  1096.25, 1104.75, 1102.25, 1097.5,  1092.75, 1083.75, 1089,    1090.5,
    1093.25, 1094.25, 1096.75, 1091.5,  1088.25, 1087.5,  1083.5,  1077.75, 1073,    1065,
    1073.25, 1068.75, 1066.25, 1057.25, 1045.5,  1046,    1039.75, 1014.75, 1014.75, 997.5,
    1026.25, 1020.25, 1034.5,  1030.75, 1035.75, 1045.75, 1039.75, 1041.25, 1046.25, 1057.25,
    1062.75, 1067,    1055.75, 1057.75, 1059.75, 1061.75, 1065.5,  1064.25, 1064,    1063.25,
    1061,    1059.75, 1057,    1054.25, 1047.5,  1059.25, 1055.25, 1053.5,  1052.25, 1057.25,
    1058.25, 1055.5,  1056.75, 1057.75, 1062.75, 1059.75, 1061,    1059.25, 1078.25, 1076.5,
    1075.5,  1077.75, 1078.75, 1079.75, 1080.25, 1077.25, 1078.75, 1080.5,  1079.75, 1081.25,
    1084.25, 1083.75, 1077.75, 1078.5,  1080,    1082.25, 1081,    1082.75, 1082.75, 1080.5,
    1082.25, 1083.25, 1085.75, 1086.25, 1084.75, 1085.5,  1083,    1085.5,  1086.5,  1091.75,
};

const test_expected = [_]f64{
    0,       0,       0,       268.7,   579.33,  828.39,  988.41,  1071.11, 1101.65, 1103.41,
    1093.69, 1082.14, 1072.48, 1066.3,  1062.99, 1060.83, 1059.3,  1058.84, 1059.42, 1060.22,
    1060.97, 1062.17, 1063.76, 1065.45, 1067.81, 1070.57, 1072.95, 1074.95, 1076.71, 1078.14,
    1078.91, 1079.09, 1079.02, 1078.57, 1077.77, 1076.93, 1076.19, 1075.36, 1074.2,  1074.43,
    1076.6,  1079.15, 1081.64, 1083.65, 1084.94, 1085.87, 1086.82, 1087.76, 1088.59, 1089.92,
    1092.04, 1094.23, 1096.12, 1097.53, 1097.72, 1096.95, 1096.43, 1096.67, 1097.2,  1097.31,
    1096.89, 1096.12, 1095.48, 1095.29, 1095.07, 1095.5,  1097.17, 1099.11, 1100.73, 1102.29,
    1104.12, 1106.08, 1107.47, 1107.38, 1106.39, 1104.88, 1103.27, 1102.07, 1100.57, 1099.18,
    1097.99, 1097.03, 1096.72, 1097.2,  1098.11, 1099.12, 1100.2,  1101.16, 1101.58, 1100.94,
    1100.16, 1100.62, 1101.95, 1103.65, 1105.92, 1108.33, 1110.03, 1110.69, 1110.53, 1109.61,
    1108.63, 1107.97, 1107.55, 1107.3,  1106.84, 1106.12, 1105.07, 1103.92, 1103.18, 1102.99,
    1103.17, 1103.26, 1102.92, 1102.1,  1101.17, 1100.52, 1100.54, 1101.16, 1102.08, 1103.54,
    1105.65, 1107.37, 1108.07, 1107.94, 1106.91, 1105.34, 1103.79, 1101.97, 1100.22, 1098.59,
    1096.79, 1094.99, 1093.5,  1092.63, 1092.43, 1092.81, 1093.01, 1092.81, 1092.52, 1092.05,
    1091.27, 1090.6,  1089.89, 1089.81, 1090.97, 1092.75, 1094.27, 1094.8,  1094.72, 1094.31,
    1093.72, 1093.58, 1094.04, 1094.5,  1094.38, 1094.07, 1093.83, 1093.68, 1094.3,  1095.49,
    1096.52, 1097.19, 1097.64, 1097.92, 1097.96, 1098.36, 1099.48, 1100.99, 1102.34, 1102.72,
    1102.37, 1101.44, 1100.15, 1099.43, 1099.05, 1099.17, 1100.16, 1101.47, 1102.46, 1102.82,
    1102.53, 1101.36, 1100.13, 1100.16, 1101.3,  1102.64, 1103.35, 1103.58, 1103.69, 1103.75,
    1103.36, 1102.07, 1100.66, 1100.11, 1100.71, 1101.72, 1102.62, 1103.44, 1104.17, 1105.09,
    1106.02, 1106.58, 1106.98, 1107.38, 1107.47, 1107.33, 1107.38, 1107.78, 1110.15, 1114.63,
    1118.85, 1121.73, 1123.01, 1122.97, 1122.07, 1120.89, 1120.03, 1118.33, 1115.46, 1113.6,
    1113.25, 1114.21, 1116.04, 1117.48, 1118.68, 1119.85, 1120.53, 1120.74, 1120.86, 1121.23,
    1121.73, 1122.04, 1122.16, 1122.1,  1122.15, 1122.41, 1123.05, 1124.22, 1125.75, 1127.33,
    1128.5,  1129.26, 1129.83, 1130.2,  1130.17, 1129.71, 1129.65, 1129.94, 1129.03, 1126.59,
    1123.86, 1121.79, 1120.7,  1120.56, 1120.7,  1120.71, 1120.06, 1119.24, 1118.97, 1119.37,
    1119.32, 1118.36, 1117.72, 1117.75, 1118.2,  1118.97, 1119.86, 1120.11, 1119.83, 1119.67,
    1119.65, 1119.67, 1119.7,  1118.84, 1116.97, 1115.36, 1114.67, 1114.75, 1114.91, 1114.87,
    1114.89, 1115.05, 1115.36, 1115.52, 1115.23, 1114.96, 1115.75, 1117.33, 1118.6,  1119.16,
    1119.16, 1119.05, 1118.84, 1118.38, 1117.81, 1117.31, 1117.02, 1116.87, 1116.77, 1117.17,
    1118.05, 1118.91, 1119.25, 1118.92, 1118.14, 1116.55, 1114.65, 1113.16, 1112.26, 1112.64,
    1114.35, 1116.29, 1116.73, 1115.71, 1114.74, 1114.18, 1113.94, 1113.82, 1113.82, 1113.74,
    1113.68, 1113.95, 1114.55, 1115.07, 1114.98, 1115.04, 1115.98, 1117.32, 1117.94, 1117.52,
    1116.8,  1116.02, 1115.25, 1114.49, 1113.48, 1111.62, 1108.42, 1105.16, 1102.2,  1099.22,
    1096.86, 1095.42, 1094.88, 1094.72, 1094.35, 1092.9,  1090.43, 1087.85, 1085.69, 1084.88,
    1085.65, 1086.98, 1088.67, 1090.52, 1092.35, 1094.06, 1095.19, 1095.86, 1095.95, 1095.84,
    1096.41, 1097.77, 1099.59, 1101.2,  1102.86, 1104.97, 1107.2,  1109.15, 1110.73, 1112.27,
    1113.5,  1114.29, 1114.8,  1114.99, 1115.36, 1116.21, 1117.39, 1118.29, 1118.39, 1117.92,
    1117.09, 1116.15, 1115.28, 1114.89, 1114.78, 1114.51, 1114.46, 1114.77, 1115.42, 1116.09,
    1113.78, 1108.52, 1104.2,  1101.42, 1099.7,  1098.81, 1098.15, 1097.14, 1095.77, 1094.32,
    1093.77, 1094.93, 1096.28, 1097.9,  1099.99, 1100.82, 1099.72, 1096.36, 1092.45, 1090.16,
    1089.65, 1090.48, 1092.09, 1093.27, 1092.89, 1091.47, 1089.37, 1086.3,  1082.27, 1077.25,
    1073.13, 1070.9,  1069.12, 1066.52, 1061.62, 1055.59, 1049.89, 1041.81, 1031.64, 1021,
    1014.31, 1013.82, 1017.05, 1022.32, 1027.26, 1032.7,  1037.48, 1040.22, 1042.23, 1045.47,
    1050.48, 1056.19, 1059.85, 1060.57, 1060.4,  1060.42, 1061.24, 1062.5,  1063.43, 1063.86,
    1063.6,  1062.67, 1061.21, 1059.19, 1056.26, 1054.32, 1054.27, 1054.27, 1053.92, 1053.99,
    1054.97, 1055.85, 1056.29, 1056.71, 1057.78, 1059.1,  1059.96, 1060.36, 1062.65, 1067.32,
    1071.43, 1074.44, 1076.64, 1078.21, 1079.31, 1079.62, 1079.33, 1079.29, 1079.49, 1079.82,
    1080.7,  1081.9,  1082.1,  1081.18, 1080.31, 1080.16, 1080.47, 1080.95, 1081.61, 1081.88,
    1081.86, 1082.08, 1082.78, 1083.89, 1084.75, 1085.2,  1085.14, 1084.89, 1085.07, 1086.17,
};

test "SuperSmoother update" {
    const skip_rows = 60;
    const tolerance = 0.5;

    var ss = try SuperSmoother.init(.{});
    ss.fixSlices();

    for (0..test_input.len) |i| {
        const act = ss.update(test_input[i]);

        if (i < 2) {
            try testing.expect(math.isNan(act));
            continue;
        }

        if (i < skip_rows) continue;

        try testing.expect(almostEqual(act, test_expected[i], tolerance));
    }

    // NaN passthrough.
    try testing.expect(math.isNan(ss.update(math.nan(f64))));
}

test "SuperSmoother isPrimed" {
    var ss = try SuperSmoother.init(.{});
    ss.fixSlices();

    try testing.expect(!ss.isPrimed());

    _ = ss.update(test_input[0]);
    try testing.expect(!ss.isPrimed());

    _ = ss.update(test_input[1]);
    try testing.expect(!ss.isPrimed());

    _ = ss.update(test_input[2]);
    try testing.expect(ss.isPrimed());
}

test "SuperSmoother metadata" {
    var ss = try SuperSmoother.init(.{});
    ss.fixSlices();
    var meta: Metadata = undefined;
    ss.getMetadata(&meta);

    try testing.expectEqual(Identifier.super_smoother, meta.identifier);
    try testing.expectEqualStrings("ss(10, hl/2)", meta.mnemonic);
    try testing.expectEqual(@as(usize, 1), meta.outputs_len);
}

test "SuperSmoother constructor validation" {
    // Valid default.
    _ = try SuperSmoother.init(.{});

    // Period < 2.
    try testing.expectError(error.InvalidPeriod, SuperSmoother.init(.{ .shortest_cycle_period = 1 }));
    try testing.expectError(error.InvalidPeriod, SuperSmoother.init(.{ .shortest_cycle_period = 0 }));
    try testing.expectError(error.InvalidPeriod, SuperSmoother.init(.{ .shortest_cycle_period = -1 }));
}

test "SuperSmoother updateBar" {
    var ss = try SuperSmoother.init(.{});
    ss.fixSlices();

    const bar1 = Bar{ .time = 1000, .open = 91, .high = 100, .low = 100, .close = 91.5, .volume = 1000 };
    const out1 = ss.updateBar(&bar1);
    try testing.expect(math.isNan(out1.slice()[0].scalar.value));

    // Prime.
    const bar2 = Bar{ .time = 2000, .open = 92, .high = 100, .low = 100, .close = 94.815, .volume = 1000 };
    _ = ss.updateBar(&bar2);
    const bar3 = Bar{ .time = 3000, .open = 93, .high = 100, .low = 100, .close = 95, .volume = 1000 };
    const out3 = ss.updateBar(&bar3);
    try testing.expect(!math.isNan(out3.slice()[0].scalar.value));
}

test "SuperSmoother custom bar component mnemonic" {
    var ss = try SuperSmoother.init(.{ .bar_component = bar_component.BarComponent.open });
    ss.fixSlices();
    var meta: Metadata = undefined;
    ss.getMetadata(&meta);
    try testing.expectEqualStrings("ss(10, o)", meta.mnemonic);
}
