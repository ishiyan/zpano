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
const sma_mod = @import("../../common/simple_moving_average/simple_moving_average.zig");
const ema_mod = @import("../../common/exponential_moving_average/exponential_moving_average.zig");

const OutputArray = indicator_mod.OutputArray;
const Identifier = identifier_mod.Identifier;
const Metadata = metadata_mod.Metadata;

/// Enumerates the outputs of the MACD indicator.
pub const MacdOutput = enum(u8) {
    /// MACD line (fast MA - slow MA).
    macd = 1,
    /// Signal line (MA of MACD).
    signal = 2,
    /// Histogram (MACD - Signal).
    histogram = 3,
};

/// Specifies the type of moving average.
pub const MovingAverageType = enum(u8) {
    ema = 0,
    sma = 1,
};

/// Parameters to create a MACD indicator.
pub const MacdParams = struct {
    fast_length: usize = 12,
    slow_length: usize = 26,
    signal_length: usize = 9,
    moving_average_type: MovingAverageType = .ema,
    signal_moving_average_type: MovingAverageType = .ema,
    first_is_average: ?bool = null,
    bar_component: ?bar_component.BarComponent = null,
    quote_component: ?quote_component.QuoteComponent = null,
    trade_component: ?trade_component.TradeComponent = null,
};

const MaUnion = union(enum) {
    sma: sma_mod.SimpleMovingAverage,
    ema: ema_mod.ExponentialMovingAverage,

    fn update(self: *MaUnion, sample: f64) f64 {
        return switch (self.*) {
            .sma => |*s| s.update(sample),
            .ema => |*e| e.update(sample),
        };
    }

    fn isPrimed(self: *const MaUnion) bool {
        return switch (self.*) {
            .sma => |*s| s.isPrimed(),
            .ema => |*e| e.isPrimed(),
        };
    }

    fn deinit(self: *MaUnion) void {
        switch (self.*) {
            .sma => |*s| s.deinit(),
            .ema => {},
        }
    }
};

fn newMa(allocator: std.mem.Allocator, ma_type: MovingAverageType, length: usize, first_is_average: bool) !MaUnion {
    switch (ma_type) {
        .sma => {
            var sma = try sma_mod.SimpleMovingAverage.init(allocator, .{ .length = length });
            sma.fixSlices();
            return .{ .sma = sma };
        },
        .ema => {
            var ema = try ema_mod.ExponentialMovingAverage.initLength(.{
                .length = length,
                .first_is_average = first_is_average,
            });
            ema.fixSlices();
            return .{ .ema = ema };
        },
    }
}

fn maLabel(ma_type: MovingAverageType) []const u8 {
    return switch (ma_type) {
        .sma => "SMA",
        .ema => "EMA",
    };
}

/// Moving Average Convergence Divergence (MACD) by Gerald Appel.
///
/// MACD = fast MA - slow MA
/// Signal = MA of MACD
/// Histogram = MACD - Signal
pub const MovingAverageConvergenceDivergence = struct {
    fast_ma: MaUnion,
    slow_ma: MaUnion,
    signal_ma: MaUnion,

    macd_value: f64,
    signal_value: f64,
    histogram_value: f64,
    primed: bool,

    fast_delay: usize,
    fast_count: usize,

    bar_func: bar_component.BarFunc,
    quote_func: quote_component.QuoteFunc,
    trade_func: trade_component.TradeFunc,

    allocator: std.mem.Allocator,
    mnemonic_buf: [128]u8,
    mnemonic_len: usize,
    description_buf: [192]u8,
    description_len: usize,

    pub fn init(allocator: std.mem.Allocator, params: MacdParams) !MovingAverageConvergenceDivergence {
        var fast_length = params.fast_length;
        var slow_length = params.slow_length;
        const signal_length = params.signal_length;

        if (fast_length < 2) return error.InvalidFastLength;
        if (slow_length < 2) return error.InvalidSlowLength;
        if (signal_length < 1) return error.InvalidSignalLength;

        // Auto-swap fast/slow if needed (matches TaLib behavior).
        if (slow_length < fast_length) {
            const tmp = fast_length;
            fast_length = slow_length;
            slow_length = tmp;
        }

        const bc = params.bar_component orelse bar_component.default_bar_component;
        const qc = params.quote_component orelse quote_component.default_quote_component;
        const tc = params.trade_component orelse trade_component.default_trade_component;

        // Default FirstIsAverage to true (TA-Lib compatible).
        const first_is_average = params.first_is_average orelse true;

        var fast_ma = try newMa(allocator, params.moving_average_type, fast_length, first_is_average);
        var slow_ma = try newMa(allocator, params.moving_average_type, slow_length, first_is_average);
        var signal_ma = try newMa(allocator, params.signal_moving_average_type, signal_length, first_is_average);

        // Fix slices after moving into struct fields below.
        _ = &fast_ma;
        _ = &slow_ma;
        _ = &signal_ma;

        var triple_buf: [64]u8 = undefined;
        const triple = component_triple_mnemonic_mod.componentTripleMnemonic(&triple_buf, bc, qc, tc);

        // Build mnemonic: macd(12,26,9) or macd(12,26,9,SMA,EMA) if non-default types.
        var mnemonic_buf: [128]u8 = undefined;
        var mnemonic_slice: []u8 = undefined;

        if (params.moving_average_type != .ema or params.signal_moving_average_type != .ema) {
            mnemonic_slice = std.fmt.bufPrint(&mnemonic_buf, "macd({d},{d},{d},{s},{s}{s})", .{
                fast_length,
                slow_length,
                signal_length,
                maLabel(params.moving_average_type),
                maLabel(params.signal_moving_average_type),
                triple,
            }) catch return error.MnemonicTooLong;
        } else {
            mnemonic_slice = std.fmt.bufPrint(&mnemonic_buf, "macd({d},{d},{d}{s})", .{
                fast_length,
                slow_length,
                signal_length,
                triple,
            }) catch return error.MnemonicTooLong;
        }
        const mnemonic_len = mnemonic_slice.len;

        var description_buf: [192]u8 = undefined;
        const desc_slice = std.fmt.bufPrint(&description_buf, "Moving Average Convergence Divergence {s}", .{mnemonic_slice}) catch
            return error.MnemonicTooLong;
        const description_len = desc_slice.len;

        return .{
            .fast_ma = fast_ma,
            .slow_ma = slow_ma,
            .signal_ma = signal_ma,
            .macd_value = math.nan(f64),
            .signal_value = math.nan(f64),
            .histogram_value = math.nan(f64),
            .primed = false,
            .fast_delay = slow_length - fast_length,
            .fast_count = 0,
            .bar_func = bar_component.componentValue(bc),
            .quote_func = quote_component.componentValue(qc),
            .trade_func = trade_component.componentValue(tc),
            .allocator = allocator,
            .mnemonic_buf = mnemonic_buf,
            .mnemonic_len = mnemonic_len,
            .description_buf = description_buf,
            .description_len = description_len,
        };
    }

    pub fn deinit(self: *MovingAverageConvergenceDivergence) void {
        self.fast_ma.deinit();
        self.slow_ma.deinit();
        self.signal_ma.deinit();
    }

    pub fn fixSlices(self: *MovingAverageConvergenceDivergence) void {
        _ = self;
        // MACD doesn't use LineIndicator, so no slice fixup needed for mnemonic/description.
        // The mnemonic/description are read from the buffers directly.
    }

    /// Returns macd, signal, histogram.
    pub fn updateValues(self: *MovingAverageConvergenceDivergence, sample: f64) struct { macd: f64, signal: f64, histogram: f64 } {
        const nan = math.nan(f64);

        if (math.isNan(sample)) {
            return .{ .macd = nan, .signal = nan, .histogram = nan };
        }

        // Feed slow MA every sample.
        const slow = self.slow_ma.update(sample);

        // Delay fast MA to align SMA seed windows.
        var fast: f64 = nan;
        if (self.fast_count < self.fast_delay) {
            self.fast_count += 1;
        } else {
            fast = self.fast_ma.update(sample);
        }

        if (math.isNan(fast) or math.isNan(slow)) {
            self.macd_value = nan;
            self.signal_value = nan;
            self.histogram_value = nan;
            return .{ .macd = nan, .signal = nan, .histogram = nan };
        }

        const macd = fast - slow;
        self.macd_value = macd;

        const sig = self.signal_ma.update(macd);

        if (math.isNan(sig)) {
            self.signal_value = nan;
            self.histogram_value = nan;
            return .{ .macd = macd, .signal = nan, .histogram = nan };
        }

        self.signal_value = sig;
        const hist = macd - sig;
        self.histogram_value = hist;
        self.primed = self.fast_ma.isPrimed() and self.slow_ma.isPrimed() and self.signal_ma.isPrimed();

        return .{ .macd = macd, .signal = sig, .histogram = hist };
    }

    pub fn isPrimed(self: *const MovingAverageConvergenceDivergence) bool {
        return self.primed;
    }

    fn mnemonic(self: *const MovingAverageConvergenceDivergence) []const u8 {
        return self.mnemonic_buf[0..self.mnemonic_len];
    }

    fn description(self: *const MovingAverageConvergenceDivergence) []const u8 {
        return self.description_buf[0..self.description_len];
    }

    pub fn getMetadata(self: *const MovingAverageConvergenceDivergence, out: *Metadata) void {
        const mn = self.mnemonic();
        const desc = self.description();

        var macd_mn_buf: [160]u8 = undefined;
        const macd_mn = std.fmt.bufPrint(&macd_mn_buf, "{s} macd", .{mn}) catch mn;
        var signal_mn_buf: [160]u8 = undefined;
        const signal_mn = std.fmt.bufPrint(&signal_mn_buf, "{s} signal", .{mn}) catch mn;
        var hist_mn_buf: [160]u8 = undefined;
        const hist_mn = std.fmt.bufPrint(&hist_mn_buf, "{s} histogram", .{mn}) catch mn;

        var macd_desc_buf: [256]u8 = undefined;
        const macd_desc = std.fmt.bufPrint(&macd_desc_buf, "{s} MACD", .{desc}) catch desc;
        var signal_desc_buf: [256]u8 = undefined;
        const signal_desc = std.fmt.bufPrint(&signal_desc_buf, "{s} Signal", .{desc}) catch desc;
        var hist_desc_buf: [256]u8 = undefined;
        const hist_desc = std.fmt.bufPrint(&hist_desc_buf, "{s} Histogram", .{desc}) catch desc;

        build_metadata_mod.buildMetadata(
            out,
            .moving_average_convergence_divergence,
            mn,
            desc,
            &[_]build_metadata_mod.OutputText{
                .{ .mnemonic = macd_mn, .description = macd_desc },
                .{ .mnemonic = signal_mn, .description = signal_desc },
                .{ .mnemonic = hist_mn, .description = hist_desc },
            },
        );
    }

    pub fn updateScalar(self: *MovingAverageConvergenceDivergence, sample: *const Scalar) OutputArray {
        const result = self.updateValues(sample.value);
        return makeOutput(sample.time, result.macd, result.signal, result.histogram);
    }

    pub fn updateBar(self: *MovingAverageConvergenceDivergence, sample: *const Bar) OutputArray {
        const v = self.bar_func(sample.*);
        return self.updateScalar(&Scalar{ .time = sample.time, .value = v });
    }

    pub fn updateQuote(self: *MovingAverageConvergenceDivergence, sample: *const Quote) OutputArray {
        const v = self.quote_func(sample.*);
        return self.updateScalar(&Scalar{ .time = sample.time, .value = v });
    }

    pub fn updateTrade(self: *MovingAverageConvergenceDivergence, sample: *const Trade) OutputArray {
        const v = self.trade_func(sample.*);
        return self.updateScalar(&Scalar{ .time = sample.time, .value = v });
    }

    fn makeOutput(time: i64, macd_v: f64, signal_v: f64, hist_v: f64) OutputArray {
        var out = OutputArray{};
        out.append(.{ .scalar = .{ .time = time, .value = macd_v } });
        out.append(.{ .scalar = .{ .time = time, .value = signal_v } });
        out.append(.{ .scalar = .{ .time = time, .value = hist_v } });
        return out;
    }

    pub fn indicator(self: *MovingAverageConvergenceDivergence) indicator_mod.Indicator {
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
        const self: *MovingAverageConvergenceDivergence = @ptrCast(@alignCast(ptr));
        return self.isPrimed();
    }

    fn vtableMetadata(ptr: *anyopaque, out: *Metadata) void {
        const self: *const MovingAverageConvergenceDivergence = @ptrCast(@alignCast(ptr));
        self.getMetadata(out);
    }

    fn vtableUpdateScalar(ptr: *anyopaque, sample: *const Scalar) OutputArray {
        const self: *MovingAverageConvergenceDivergence = @ptrCast(@alignCast(ptr));
        return self.updateScalar(sample);
    }

    fn vtableUpdateBar(ptr: *anyopaque, sample: *const Bar) OutputArray {
        const self: *MovingAverageConvergenceDivergence = @ptrCast(@alignCast(ptr));
        return self.updateBar(sample);
    }

    fn vtableUpdateQuote(ptr: *anyopaque, sample: *const Quote) OutputArray {
        const self: *MovingAverageConvergenceDivergence = @ptrCast(@alignCast(ptr));
        return self.updateQuote(sample);
    }

    fn vtableUpdateTrade(ptr: *anyopaque, sample: *const Trade) OutputArray {
        const self: *MovingAverageConvergenceDivergence = @ptrCast(@alignCast(ptr));
        return self.updateTrade(sample);
    }

    pub const InitError = error{
        InvalidFastLength,
        InvalidSlowLength,
        InvalidSignalLength,
        MnemonicTooLong,
        OutOfMemory,
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

fn testMacdExpected() [252]f64 {
    const nan = math.nan(f64);
    return .{
        nan,           nan,           nan,           nan,           nan,           nan,
        nan,           nan,           nan,           nan,           nan,           nan,
        nan,           nan,           nan,           nan,           nan,           nan,
        nan,           nan,           nan,           nan,           nan,           nan,
        nan,           -3.3124358974, -3.3275925926, -2.9178425094, -2.7929062048, -2.6731651547,
        -2.6286596008, -2.4142616928, -2.3232745946, -1.9738314311, -1.7174434092, -1.6242340655,
        -1.5374833367, -1.6067528987, -1.6973572233, -1.7741286204, -1.8539452008, -1.7258363085,
        -1.3114394691, -0.9518813817, -0.5273058843, -0.2161734475, 0.0850975323,  0.1255224022,
        0.3153080136,  0.4205211505,  0.3808958744,  0.3283587392,  -0.0783111956, -0.4582570284,
        -0.8153284697, -0.9214633200, -0.9219226664, -0.8690982492, -0.5984341247, -0.3524330685,
        -0.2079315791, -0.1023200809, 0.2583991280,  0.5005762959,  0.8242128087,  1.0883244606,
        1.2553256184,  1.2573883419,  1.1074672467,  0.9526576451,  0.7554973288,  0.2984572253,
        -0.2078101745, -0.4824320954, -0.6071366535, 0.2043015036,  1.0471171366,  2.0995554003,
        2.9847471407,  3.3650569414,  3.6170938857,  3.9480391397,  4.2843861148,  4.4891094017,
        4.6035323826,  4.5258463148,  4.7324928849,  4.8955066605,  5.0619646709,  5.3141814634,
        6.2688905577,  6.6762108959,  6.8494537717,  6.9470149562,  6.8397820956,  6.5581633917,
        6.1630699526,  5.5188449510,  4.8491091810,  4.8698106811,  4.6610177176,  4.4443164759,
        3.9048005005,  3.5772047876,  3.2151602286,  3.1389696456,  3.4024787940,  3.2606414738,
        3.1426704512,  2.8748279640,  2.5572325342,  2.3741907299,  2.2332412738,  2.4818621697,
        2.6084812514,  2.7226303183,  3.1001245409,  3.2504675980,  3.2969133823,  3.2654174768,
        3.2482008275,  3.1530231013,  3.2068522549,  3.5818254977,  3.9743905682,  4.3363786940,
        4.4708551037,  4.6696516628,  4.8671173241,  5.2359907929,  5.4996268064,  5.6490894745,
        5.6515563572,  5.5140981444,  5.3387779134,  5.0113472148,  4.1887579528,  3.5563731128,
        2.6111556670,  1.9158296834,  1.2048383323,  0.8933227343,  0.8089920210,  0.4935877069,
        0.2663781992,  -0.1890865512, -0.7735209107, -1.2919977347, -1.3093570955, -1.2833069412,
        -1.3527741201, -1.6215273330, -1.4945217040, -1.5526851376, -1.4010731824, -0.9472339219,
        -0.4907247130, -0.4968132278, -0.5701097767, -0.7159676452, -0.6074976830, -0.7102349185,
        -0.7531175398, -0.7326624136, -0.6237284838, -0.5958887555, -0.4579984982, -0.1301585702,
        0.0180938928,  0.3725584276,  0.8965114572,  1.1970842130,  1.7380220222,  2.1619701972,
        2.2604826397,  2.3860923011,  2.3480247446,  2.1366837844,  1.5774088580,  1.5009694272,
        1.1838619595,  0.7679645194,  0.1788942799,  -0.0453504296, -0.3800680745, -0.5980939436,
        -1.0365120012, -1.2932040749, -1.7388379893, -1.8990487428, -1.8330138775, -1.9845476567,
        -2.3056120404, -2.7598292978, -3.0244174134, -3.5362808384, -4.2910880136, -4.6788020339,
        -4.8598451890, -5.0157069751, -5.0710892066, -5.0662623030, -6.2810986344, -6.9267840947,
        -7.3593093517, -7.4842871531, -7.6963434117, -7.6200189030, -7.2093370925, -6.9250782712,
        -6.7782079324, -6.6209937589, -6.6465268849, -6.6952889992, -6.3628343709, -6.0553803529,
        -5.4758607622, -5.1189623848, -4.7116036858, -4.4831419850, -4.2083854732, -4.0145633004,
        -3.4779269736, -2.5440047578, -1.4698035373, -0.7558304609, -0.3122808375, 0.0786743471,
        0.3194663083,  0.4143378869,  0.5126636208,  0.7314354792,  1.4209982119,  2.2745188917,
        2.9667699213,  3.6077441477,  3.6771371268,  3.3409961603,  3.0954042787,  2.8318162885,
        2.4135428912,  2.2330319059,  2.1307736479,  1.9625567945,  1.8817879573,  1.6279458431,
        1.4599732939,  1.4018754472,  1.3451677274,  1.2207936709,  1.0894944110,  0.9040092995,
    };
}

fn testSignalExpected() [252]f64 {
    const nan = math.nan(f64);
    return .{
        nan,           nan,           nan,           nan,           nan,           nan,
        nan,           nan,           nan,           nan,           nan,           nan,
        nan,           nan,           nan,           nan,           nan,           nan,
        nan,           nan,           nan,           nan,           nan,           nan,
        nan,           nan,           nan,           nan,           nan,           nan,
        nan,           nan,           nan,           -2.7071077420, -2.5091748755, -2.3321867135,
        -2.1732460381, -2.0599474102, -1.9874293729, -1.9447692224, -1.9266044180, -1.8864507961,
        -1.7714485307, -1.6075351009, -1.3914892576, -1.1564260956, -0.9081213700, -0.7013926156,
        -0.4980524897, -0.3143377617, -0.1752910345, -0.0745610797, -0.0753111029, -0.1519002880,
        -0.2845859243, -0.4119614035, -0.5139536561, -0.5849825747, -0.5876728847, -0.5406249215,
        -0.4740862530, -0.3997330186, -0.2681065893, -0.1143700122, 0.0733465520,  0.2763421337,
        0.4721388306,  0.6291887329,  0.7248444356,  0.7704070775,  0.7674251278,  0.6736315473,
        0.4973432029,  0.3013881433,  0.1196831839,  0.1366068478,  0.3187089056,  0.6748782045,
        1.1368519918,  1.5824929817,  1.9894131625,  2.3811383579,  2.7617879093,  3.1072522078,
        3.4065082427,  3.6303758572,  3.8507992627,  4.0597407423,  4.2601855280,  4.4709847151,
        4.8305658836,  5.1996948861,  5.5296466632,  5.8131203218,  6.0184526766,  6.1263948196,
        6.1337298462,  6.0107528671,  5.7784241299,  5.5967014402,  5.4095646956,  5.2165150517,
        4.9541721415,  4.6787786707,  4.3860549823,  4.1366379149,  3.9898060907,  3.8439731674,
        3.7037126241,  3.5379356921,  3.3417950605,  3.1482741944,  2.9652676103,  2.8685865222,
        2.8165654680,  2.7977784381,  2.8582476586,  2.9366916465,  3.0087359937,  3.0600722903,
        3.0976979977,  3.1087630184,  3.1283808657,  3.2190697921,  3.3701339473,  3.5633828967,
        3.7448773381,  3.9298322030,  4.1172892272,  4.3410295403,  4.5727489936,  4.7880170897,
        4.9607249432,  5.0713995835,  5.1248752494,  5.1021696425,  4.9194873046,  4.6468644662,
        4.2397227064,  3.7749441018,  3.2609229479,  2.7874029052,  2.3917207283,  2.0120941241,
        1.6629509391,  1.2925434410,  0.8793305707,  0.4450649096,  0.0941805086,  -0.1813169814,
        -0.4156084091, -0.6567921939, -0.8243380959, -0.9700075043, -1.0562206399, -1.0344232963,
        -0.9256835796, -0.8399095093, -0.7859495628, -0.7719531792, -0.7390620800, -0.7332966477,
        -0.7372608261, -0.7363411436, -0.7138186116, -0.6902326404, -0.6437858120, -0.5410603636,
        -0.4292295123, -0.2688719243, -0.0357952480, 0.2107806442,  0.5162289198,  0.8453771753,
        1.1283982681,  1.3799370747,  1.5735546087,  1.6861804439,  1.6644261267,  1.6317347868,
        1.5421602213,  1.3873210810,  1.1456357207,  0.9074384907,  0.6499371776,  0.4003309534,
        0.1129623625,  -0.1682709250, -0.4823843379, -0.7657172189, -0.9791765506, -1.1802507718,
        -1.4053230255, -1.6762242800, -1.9458629067, -2.2639464930, -2.6693747971, -3.0712602445,
        -3.4289772334, -3.7463231817, -4.0112763867, -4.2222735700, -4.6340385828, -5.0925876852,
        -5.5459320185, -5.9336030454, -6.2861511187, -6.5529246756, -6.6842071589, -6.7323813814,
        -6.7415466916, -6.7174361051, -6.7032542610, -6.7016612087, -6.6338958411, -6.5181927435,
        -6.3097263472, -6.0715735547, -5.7995795809, -5.5362920618, -5.2707107440, -5.0194812553,
        -4.7111703990, -4.2777372707, -3.7161505240, -3.1240865114, -2.5617253766, -2.0336454319,
        -1.5630230838, -1.1675508897, -0.8315079876, -0.5189192942, -0.1309357930, 0.3501551439,
        0.8734780994,  1.4203313091,  1.8716924726,  2.1655532101,  2.3515234238,  2.4475819968,
        2.4407741757,  2.3992257217,  2.3455353070,  2.2689396045,  2.1915092750,  2.0787965886,
        1.9550319297,  1.8444006332,  1.7445540520,  1.6398019758,  1.5297404628,  1.4045942302,
    };
}

fn testHistogramExpected() [252]f64 {
    const nan = math.nan(f64);
    return .{
        nan,           nan,           nan,           nan,           nan,           nan,
        nan,           nan,           nan,           nan,           nan,           nan,
        nan,           nan,           nan,           nan,           nan,           nan,
        nan,           nan,           nan,           nan,           nan,           nan,
        nan,           nan,           nan,           nan,           nan,           nan,
        nan,           nan,           nan,           0.7332763109,  0.7917314662,  0.7079526480,
        0.6357627014,  0.4531945116,  0.2900721496,  0.1706406020,  0.0726592173,  0.1606144877,
        0.4600090616,  0.6556537193,  0.8641833733,  0.9402526481,  0.9932189023,  0.8269150177,
        0.8133605033,  0.7348589122,  0.5561869088,  0.4029198189,  -0.0030000927, -0.3063567404,
        -0.5307425453, -0.5095019166, -0.4079690103, -0.2841156745, -0.0107612400, 0.1881918530,
        0.2661546738,  0.2974129376,  0.5265057173,  0.6149463082,  0.7508662568,  0.8119823269,
        0.7831867878,  0.6281996090,  0.3826228110,  0.1822505675,  -0.0119277990, -0.3751743219,
        -0.7051533774, -0.7838202387, -0.7268198374, 0.0676946558,  0.7284082310,  1.4246771957,
        1.8478951489,  1.7825639597,  1.6276807232,  1.5669007817,  1.5225982055,  1.3818571939,
        1.1970241398,  0.8954704577,  0.8816936222,  0.8357659183,  0.8017791429,  0.8431967484,
        1.4383246741,  1.4765160098,  1.3198071085,  1.1338946344,  0.8213294190,  0.4317685721,
        0.0293401064,  -0.4919079162, -0.9293149489, -0.7268907590, -0.7485469780, -0.7721985758,
        -1.0493716410, -1.1015738831, -1.1708947537, -0.9976682694, -0.5873272968, -0.5833316935,
        -0.5610421729, -0.6631077281, -0.7845625263, -0.7740834645, -0.7320263365, -0.3867243525,
        -0.2080842166, -0.0751481198, 0.2418768822,  0.3137759515,  0.2881773887,  0.2053451865,
        0.1505028297,  0.0442600829,  0.0784713892,  0.3627557055,  0.6042566209,  0.7729957973,
        0.7259777656,  0.7398194598,  0.7498280968,  0.8949612525,  0.9268778128,  0.8610723847,
        0.6908314140,  0.4426985609,  0.2139026639,  -0.0908224277, -0.7307293518, -1.0904913534,
        -1.6285670394, -1.8591144184, -2.0560846156, -1.8940801708, -1.5827287073, -1.5185064171,
        -1.3965727399, -1.4816299923, -1.6528514814, -1.7370626443, -1.4035376041, -1.1019899598,
        -0.9371657110, -0.9647351391, -0.6701836081, -0.5826776334, -0.3448525425, 0.0871893744,
        0.4349588666,  0.3430962815,  0.2158397860,  0.0559855340,  0.1315643970,  0.0230617292,
        -0.0158567137, 0.0036787300,  0.0900901279,  0.0943438849,  0.1857873138,  0.4109017934,
        0.4473234051,  0.6414303520,  0.9323067052,  0.9863035689,  1.2217931024,  1.3165930219,
        1.1320843716,  1.0061552263,  0.7744701359,  0.4505033406,  -0.0870172687, -0.1307653596,
        -0.3582982618, -0.6193565615, -0.9667414409, -0.9527889202, -1.0300052521, -0.9984248970,
        -1.1494743637, -1.1249331499, -1.2564536514, -1.1333315240, -0.8538373269, -0.8042968849,
        -0.9002890149, -1.0836050178, -1.0785545068, -1.2723343454, -1.6217132165, -1.6075417894,
        -1.4308679556, -1.2693837934, -1.0598128199, -0.8439887331, -1.6470600515, -1.8341964095,
        -1.8133773332, -1.5506841077, -1.4101922930, -1.0670942275, -0.5251299336, -0.1926968898,
        -0.0366612408, 0.0964423462,  0.0567273761,  0.0063722094,  0.2710614703,  0.4628123905,
        0.8338655850,  0.9526111700,  1.0879758952,  1.0531500767,  1.0623252708,  1.0049179549,
        1.2332434254,  1.7337325130,  2.2463469868,  2.3682560505,  2.2494445391,  2.1123197790,
        1.8824893922,  1.5818887766,  1.3441716084,  1.2503547735,  1.5519340049,  1.9243637478,
        2.0932918219,  2.1874128386,  1.8054446542,  1.1754429501,  0.7438808548,  0.3842342917,
        -0.0272312845, -0.1661938158, -0.2147616590, -0.3063828100, -0.3097213178, -0.4508507455,
        -0.4950586358, -0.4425251860, -0.3993863246, -0.4190083049, -0.4402460519, -0.5005849307,
    };
}

test "MACD default params full validation" {
    const allocator = testing.allocator;
    const tolerance = 1e-8;

    const input = testInput();
    const exp_macd = testMacdExpected();
    const exp_signal = testSignalExpected();
    const exp_histogram = testHistogramExpected();

    var ind = try MovingAverageConvergenceDivergence.init(allocator, .{});
    defer ind.deinit();
    ind.fixSlices();

    for (0..252) |i| {
        const result = ind.updateValues(input[i]);

        if (math.isNan(exp_macd[i])) {
            try testing.expect(math.isNan(result.macd));
            try testing.expect(math.isNan(result.signal));
            try testing.expect(math.isNan(result.histogram));
            continue;
        }

        if (!math.isNan(exp_macd[i])) {
            try testing.expect(@abs(result.macd - exp_macd[i]) <= tolerance);
        }

        if (math.isNan(exp_signal[i])) {
            try testing.expect(math.isNan(result.signal));
            try testing.expect(math.isNan(result.histogram));
            continue;
        }

        try testing.expect(@abs(result.signal - exp_signal[i]) <= tolerance);
        try testing.expect(@abs(result.histogram - exp_histogram[i]) <= tolerance);
    }
}

test "MACD TaLib spot check" {
    const allocator = testing.allocator;
    const tolerance = 5e-4;
    const input = testInput();

    var ind = try MovingAverageConvergenceDivergence.init(allocator, .{});
    defer ind.deinit();
    ind.fixSlices();

    var result: @TypeOf(ind.updateValues(0)) = undefined;
    for (0..34) |i| {
        result = ind.updateValues(input[i]);
    }

    try testing.expect(@abs(result.macd - (-1.9738)) < tolerance);
    try testing.expect(@abs(result.signal - (-2.7071)) < tolerance);
    const exp_hist = (-1.9738) - (-2.7071);
    try testing.expect(@abs(result.histogram - exp_hist) < tolerance);
}

test "MACD period inversion" {
    const allocator = testing.allocator;
    const tolerance = 5e-4;
    const input = testInput();

    // fast=26, slow=12 should auto-swap.
    var ind = try MovingAverageConvergenceDivergence.init(allocator, .{
        .fast_length = 26,
        .slow_length = 12,
    });
    defer ind.deinit();
    ind.fixSlices();

    var result: @TypeOf(ind.updateValues(0)) = undefined;
    for (0..34) |i| {
        result = ind.updateValues(input[i]);
    }

    try testing.expect(@abs(result.macd - (-1.9738)) < tolerance);
    try testing.expect(@abs(result.signal - (-2.7071)) < tolerance);
}

test "MACD isPrimed" {
    const allocator = testing.allocator;

    var ind = try MovingAverageConvergenceDivergence.init(allocator, .{
        .fast_length = 3,
        .slow_length = 5,
        .signal_length = 2,
    });
    defer ind.deinit();
    ind.fixSlices();

    try testing.expect(!ind.isPrimed());

    for (0..6) |i| {
        _ = ind.updateValues(@as(f64, @floatFromInt(i + 1)));
        if (i < 5) {
            try testing.expect(!ind.isPrimed());
        }
    }

    try testing.expect(ind.isPrimed());
}

test "MACD NaN passthrough" {
    const allocator = testing.allocator;

    var ind = try MovingAverageConvergenceDivergence.init(allocator, .{});
    defer ind.deinit();
    ind.fixSlices();

    const result = ind.updateValues(math.nan(f64));
    try testing.expect(math.isNan(result.macd));
    try testing.expect(math.isNan(result.signal));
    try testing.expect(math.isNan(result.histogram));
}

test "MACD metadata default" {
    const allocator = testing.allocator;

    var ind = try MovingAverageConvergenceDivergence.init(allocator, .{});
    defer ind.deinit();
    ind.fixSlices();

    var meta: Metadata = undefined;
    ind.getMetadata(&meta);

    try testing.expectEqual(Identifier.moving_average_convergence_divergence, meta.identifier);
    try testing.expectEqualStrings("macd(12,26,9)", meta.mnemonic);
    try testing.expectEqual(@as(usize, 3), meta.outputs_len);
    try testing.expectEqual(@as(u8, 1), meta.outputs_buf[0].kind);
    try testing.expectEqual(@as(u8, 2), meta.outputs_buf[1].kind);
    try testing.expectEqual(@as(u8, 3), meta.outputs_buf[2].kind);
}

test "MACD metadata SMA" {
    const allocator = testing.allocator;

    var ind = try MovingAverageConvergenceDivergence.init(allocator, .{
        .moving_average_type = .sma,
    });
    defer ind.deinit();
    ind.fixSlices();

    var meta: Metadata = undefined;
    ind.getMetadata(&meta);

    try testing.expectEqualStrings("macd(12,26,9,SMA,EMA)", meta.mnemonic);
}

test "MACD invalid params" {
    const allocator = testing.allocator;

    const r1 = MovingAverageConvergenceDivergence.init(allocator, .{ .fast_length = 1 });
    try testing.expect(if (r1) |_| false else |_| true);

    const r2 = MovingAverageConvergenceDivergence.init(allocator, .{ .slow_length = 1 });
    try testing.expect(if (r2) |_| false else |_| true);

    const r3 = MovingAverageConvergenceDivergence.init(allocator, .{ .signal_length = 0 });
    try testing.expect(if (r3) |_| false else |_| true);
}

test "MACD entity update" {
    const allocator = testing.allocator;
    const tolerance = 5e-4;
    const input = testInput();

    var ind = try MovingAverageConvergenceDivergence.init(allocator, .{});
    defer ind.deinit();
    ind.fixSlices();

    // First 24 should have NaN MACD.
    for (0..25) |i| {
        const scalar = Scalar{ .time = 0, .value = input[i] };
        const out = ind.updateScalar(&scalar);
        const items = out.slice();
        const m = items[0].scalar.value;
        try testing.expect(math.isNan(m));
    }

    // Feed through index 33.
    for (25..33) |i| {
        const scalar = Scalar{ .time = 0, .value = input[i] };
        _ = ind.updateScalar(&scalar);
    }

    const scalar = Scalar{ .time = 0, .value = input[33] };
    const out = ind.updateScalar(&scalar);
    const items = out.slice();
    const macd_v = items[0].scalar.value;
    const signal_v = items[1].scalar.value;
    const hist_v = items[2].scalar.value;

    try testing.expect(@abs(macd_v - (-1.9738)) < tolerance);
    try testing.expect(@abs(signal_v - (-2.7071)) < tolerance);
    const exp_hist = (-1.9738) - (-2.7071);
    try testing.expect(@abs(hist_v - exp_hist) < tolerance);
}
