const std = @import("std");
const math = std.math;


const entities = @import("entities");
const Bar = entities.Bar;
const Quote = entities.Quote;
const Trade = entities.Trade;
const Scalar = entities.Scalar;
const indicator_mod = @import("../../core/indicator.zig");
const build_metadata_mod = @import("../../core/build_metadata.zig");
const identifier_mod = @import("../../core/identifier.zig");
const metadata_mod = @import("../../core/metadata.zig");

const true_range_mod = @import("../true_range/true_range.zig");
const TrueRange = true_range_mod.TrueRange;

const OutputArray = indicator_mod.OutputArray;
const Identifier = identifier_mod.Identifier;
const Metadata = metadata_mod.Metadata;

/// Enumerates the outputs of the Average True Range indicator.
pub const AverageTrueRangeOutput = enum(u8) {
    /// The scalar value of the Average True Range.
    value = 1,
};

/// Welles Wilder's Average True Range (ATR) indicator.
///
/// ATR averages True Range (TR) values over the specified length using the Wilder method:
///   - multiply the previous value by (length - 1)
///   - add the current TR value
///   - divide by length
///
/// The initial ATR value is a simple average of the first `length` TR values.
/// The indicator is not primed during the first `length` updates.
pub const AverageTrueRange = struct {
    length: i32,
    last_index: i32,
    stage: i32,
    window_count: i32,
    window: ?[]f64,
    window_sum: f64,
    value: f64,
    primed: bool,
    true_range: TrueRange,
    allocator: std.mem.Allocator,

    const mnemonic_str = "atr";
    const description_str = "Average True Range";

    pub const Error = error{
        InvalidLength,
        OutOfMemory,
    };

    pub fn init(allocator: std.mem.Allocator, params: struct { length: i32 = 14 }) Error!AverageTrueRange {
        if (params.length < 1) return Error.InvalidLength;

        const last_index = params.length - 1;
        const window: ?[]f64 = if (last_index > 0)
            allocator.alloc(f64, @intCast(params.length)) catch return Error.OutOfMemory
        else
            null;

        if (window) |w| @memset(w, 0.0);

        return .{
            .length = params.length,
            .last_index = last_index,
            .stage = 0,
            .window_count = 0,
            .window = window,
            .window_sum = 0,
            .value = math.nan(f64),
            .primed = false,
            .true_range = TrueRange.init(),
            .allocator = allocator,
        };
    }

    pub fn deinit(self: *AverageTrueRange) void {
        if (self.window) |w| self.allocator.free(w);
    }

    pub fn fixSlices(_: *AverageTrueRange) void {}

    /// Update given close, high, low values.
    pub fn update(self: *AverageTrueRange, close: f64, high: f64, low: f64) f64 {
        if (math.isNan(close) or math.isNan(high) or math.isNan(low)) return math.nan(f64);

        const tr_value = self.true_range.update(close, high, low);

        if (self.last_index == 0) {
            self.value = tr_value;
            if (self.stage == 0) {
                self.stage += 1;
            } else if (self.stage == 1) {
                self.stage += 1;
                self.primed = true;
            }
            return self.value;
        }

        if (self.stage > 1) {
            // Wilder smoothing method.
            self.value *= @as(f64, @floatFromInt(self.last_index));
            self.value += tr_value;
            self.value /= @as(f64, @floatFromInt(self.length));
            return self.value;
        }

        if (self.stage == 1) {
            self.window_sum += tr_value;
            self.window.?[@intCast(self.window_count)] = tr_value;
            self.window_count += 1;

            if (self.window_count == self.length) {
                self.stage += 1;
                self.primed = true;
                self.value = self.window_sum / @as(f64, @floatFromInt(self.length));
            }

            if (self.primed) return self.value;
            return math.nan(f64);
        }

        // The very first sample is used by the True Range.
        self.stage += 1;
        return math.nan(f64);
    }

    /// Update using a single sample value as substitute for high, low, close.
    pub fn updateSample(self: *AverageTrueRange, sample: f64) f64 {
        return self.update(sample, sample, sample);
    }

    pub fn isPrimed(self: *const AverageTrueRange) bool {
        return self.primed;
    }

    pub fn getMetadata(_: *const AverageTrueRange, out: *Metadata) void {
        build_metadata_mod.buildMetadata(out, Identifier.average_true_range, mnemonic_str, description_str, &.{
            .{ .mnemonic = mnemonic_str, .description = description_str },
        });
    }

    fn makeOutput(self: *const AverageTrueRange, time: i64) OutputArray {
        var out = OutputArray{};
        out.append(.{ .scalar = Scalar{ .time = time, .value = self.value } });
        return out;
    }

    pub fn updateScalar(self: *AverageTrueRange, sample: *const Scalar) OutputArray {
        _ = self.update(sample.value, sample.value, sample.value);
        return self.makeOutput(sample.time);
    }

    pub fn updateBar(self: *AverageTrueRange, sample: *const Bar) OutputArray {
        _ = self.update(sample.close, sample.high, sample.low);
        return self.makeOutput(sample.time);
    }

    pub fn updateQuote(self: *AverageTrueRange, sample: *const Quote) OutputArray {
        const mid = (sample.bid_price + sample.ask_price) / 2.0;
        _ = self.update(mid, mid, mid);
        return self.makeOutput(sample.time);
    }

    pub fn updateTrade(self: *AverageTrueRange, sample: *const Trade) OutputArray {
        _ = self.update(sample.price, sample.price, sample.price);
        return self.makeOutput(sample.time);
    }

    // --- Indicator interface ---

    pub fn indicator(self: *AverageTrueRange) indicator_mod.Indicator {
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
        const self: *const AverageTrueRange = @ptrCast(@alignCast(ptr));
        return self.isPrimed();
    }

    fn vtableMetadata(ptr: *const anyopaque, out: *Metadata) void {
        const self: *const AverageTrueRange = @ptrCast(@alignCast(ptr));
        self.getMetadata(out);
    }

    fn vtableUpdateScalar(ptr: *anyopaque, sample: *const Scalar) OutputArray {
        const self: *AverageTrueRange = @ptrCast(@alignCast(ptr));
        return self.updateScalar(sample);
    }

    fn vtableUpdateBar(ptr: *anyopaque, sample: *const Bar) OutputArray {
        const self: *AverageTrueRange = @ptrCast(@alignCast(ptr));
        return self.updateBar(sample);
    }

    fn vtableUpdateQuote(ptr: *anyopaque, sample: *const Quote) OutputArray {
        const self: *AverageTrueRange = @ptrCast(@alignCast(ptr));
        return self.updateQuote(sample);
    }

    fn vtableUpdateTrade(ptr: *anyopaque, sample: *const Trade) OutputArray {
        const self: *AverageTrueRange = @ptrCast(@alignCast(ptr));
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
    93.25,  94.94,  96.375,  96.19,   96.0,    94.72,  95.0,   93.72,   92.47,   92.75,
    96.25,  99.625, 99.125,  92.75,   91.315,  93.25,  93.405, 90.655,  91.97,   92.25,
    90.345, 88.5,   88.25,   85.5,    84.44,   84.75,  84.44,  89.405,  88.125,  89.125,
    87.155, 87.25,  87.375,  88.97,   90.0,    89.845, 86.97,  85.94,   84.75,   85.47,
    84.47,  88.5,   89.47,   90.0,    92.44,   91.44,  92.97,  91.72,   91.155,  91.75,
    90.0,   88.875, 89.0,    85.25,   83.815,  85.25,  86.625, 87.94,   89.375,  90.625,
    90.75,  88.845, 91.97,   93.375,  93.815,  94.03,  94.03,  91.815,  92.0,    91.94,
    89.75,  88.75,  86.155,  84.875,  85.94,   99.375, 103.28, 105.375, 107.625, 105.25,
    104.5,  105.5,  106.125, 107.94,  106.25,  107.0,  108.75, 110.94,  110.94,  114.22,
    123.0,  121.75, 119.815, 120.315, 119.375, 118.19, 116.69, 115.345, 113.0,   118.315,
    116.87, 116.75, 113.87,  114.62,  115.31,  116.0,  121.69, 119.87,  120.87,  116.75,
    116.5,  116.0,  118.31,  121.5,   122.0,   121.44, 125.75, 127.75,  124.19,  124.44,
    125.75, 124.69, 125.31,  132.0,   131.31,  132.25, 133.88, 133.5,   135.5,   137.44,
    138.69, 139.19, 138.5,   138.13,  137.5,   138.88, 132.13, 129.75,  128.5,   125.44,
    125.12, 126.5,  128.69,  126.62,  126.69,  126.0,  123.12, 121.87,  124.0,   127.0,
    124.44, 122.5,  123.75,  123.81,  124.5,   127.87, 128.56, 129.63,  124.87,  124.37,
    124.87, 123.62, 124.06,  125.87,  125.19,  125.62, 126.0,  128.5,   126.75,  129.75,
    132.69, 133.94, 136.5,   137.69,  135.56,  133.56, 135.0,  132.38,  131.44,  130.88,
    129.63, 127.25, 127.81,  125.0,   126.81,  124.75, 122.81, 122.25,  121.06,  120.0,
    123.25, 122.75, 119.19,  115.06,  116.69,  114.87, 110.87, 107.25,  108.87,  109.0,
    108.5,  113.06, 93.0,    94.62,   95.12,   96.0,   95.56,  95.31,   99.0,    98.81,
    96.81,  95.94,  94.44,   92.94,   93.94,   95.5,   97.06,  97.5,    96.25,   96.37,
    95.0,   94.87,  98.25,   105.12,  108.44,  109.87, 105.0,  106.0,   104.94,  104.5,
    104.44, 106.31, 112.87,  116.5,   119.19,  121.0,  122.12, 111.94,  112.75,  110.19,
    107.94, 109.69, 111.06,  110.44,  110.12,  110.31, 110.44, 110.0,   110.75,  110.5,
    110.5,  109.5,
};

const test_input_low = [_]f64{
    90.75,  91.405, 94.25,   93.5,   92.815,  93.5,   92.0,    89.75,   89.44,  90.625,
    92.75,  96.315, 96.03,   88.815, 86.75,   90.94,  88.905,  88.78,   89.25,  89.75,
    87.5,   86.53,  84.625,  82.28,  81.565,  80.875, 81.25,   84.065,  85.595, 85.97,
    84.405, 85.095, 85.5,    85.53,  87.875,  86.565, 84.655,  83.25,   82.565, 83.44,
    82.53,  85.065, 86.875,  88.53,  89.28,   90.125, 90.75,   89.0,    88.565, 90.095,
    89.0,   86.47,  84.0,    83.315, 82.0,    83.25,  84.75,   85.28,   87.19,  88.44,
    88.25,  87.345, 89.28,   91.095, 89.53,   91.155, 92.0,    90.53,   89.97,  88.815,
    86.75,  85.065, 82.03,   81.5,   82.565,  96.345, 96.47,   101.155, 104.25, 101.75,
    101.72, 101.72, 103.155, 105.69, 103.655, 104.0,  105.53,  108.53,  108.75, 107.75,
    117.0,  118.0,  116.0,   118.5,  116.53,  116.25, 114.595, 110.875, 110.5,  110.72,
    112.62, 114.19, 111.19,  109.44, 111.56,  112.44, 117.5,   116.06,  116.56, 113.31,
    112.56, 114.0,  114.75,  118.87, 119.0,   119.75, 122.62,  123.0,   121.75, 121.56,
    123.12, 122.19, 122.75,  124.37, 128.0,   129.5,  130.81,  130.63,  132.13, 133.88,
    135.38, 135.75, 136.19,  134.5,  135.38,  133.69, 126.06,  126.87,  123.5,  122.62,
    122.75, 123.56, 125.81,  124.62, 124.37,  121.81, 118.19,  118.06,  117.56, 121.0,
    121.12, 118.94, 119.81,  121.0,  122.0,   124.5,  126.56,  123.5,   121.25, 121.06,
    122.31, 121.0,  120.87,  122.06, 122.75,  122.69, 122.87,  125.5,   124.25, 128.0,
    128.38, 130.69, 131.63,  134.38, 132.0,   131.94, 131.94,  129.56,  123.75, 126.0,
    126.25, 124.37, 121.44,  120.44, 121.37,  121.69, 120.0,   119.62,  115.5,  116.75,
    119.06, 119.06, 115.06,  111.06, 113.12,  110.0,  105.0,   104.69,  103.87, 104.69,
    105.44, 107.0,  89.0,    92.5,   92.12,   94.62,  92.81,   94.25,   96.25,  96.37,
    93.69,  93.5,   90.0,    90.19,  90.5,    92.12,  94.12,   94.87,   93.0,   93.87,
    93.0,   92.62,  93.56,   98.37,  104.44,  106.0,  101.81,  104.12,  103.37, 102.12,
    102.25, 103.37, 107.94,  112.5,  115.44,  115.5,  112.25,  107.56,  106.56, 106.87,
    104.5,  105.75, 108.62,  107.75, 108.06,  108.0,  108.19,  108.12,  109.06, 108.75,
    108.56, 106.62,
};

const test_input_close = [_]f64{
    91.5,    94.815,  94.375,  95.095, 93.78,   94.625,  92.53,   92.75,   90.315,  92.47,
    96.125,  97.25,   98.5,    89.875, 91.0,    92.815,  89.155,  89.345,  91.625,  89.875,
    88.375,  87.625,  84.78,   83.0,   83.5,    81.375,  84.44,   89.25,   86.375,  86.25,
    85.25,   87.125,  85.815,  88.97,  88.47,   86.875,  86.815,  84.875,  84.19,   83.875,
    83.375,  85.5,    89.19,   89.44,  91.095,  90.75,   91.44,   89.0,    91.0,    90.5,
    89.03,   88.815,  84.28,   83.5,   82.69,   84.75,   85.655,  86.19,   88.94,   89.28,
    88.625,  88.5,    91.97,   91.5,   93.25,   93.5,    93.155,  91.72,   90.0,    89.69,
    88.875,  85.19,   83.375,  84.875, 85.94,   97.25,   99.875,  104.94,  106.0,   102.5,
    102.405, 104.595, 106.125, 106.0,  106.065, 104.625, 108.625, 109.315, 110.5,   112.75,
    123.0,   119.625, 118.75,  119.25, 117.94,  116.44,  115.19,  111.875, 110.595, 118.125,
    116.0,   116.0,   112.0,   113.75, 112.94,  116.0,   120.5,   116.62,  117.0,   115.25,
    114.31,  115.5,   115.87,  120.69, 120.19,  120.75,  124.75,  123.37,  122.94,  122.56,
    123.12,  122.56,  124.62,  129.25, 131.0,   132.25,  131.0,   132.81,  134.0,   137.38,
    137.81,  137.88,  137.25,  136.31, 136.25,  134.63,  128.25,  129.0,   123.87,  124.81,
    123.0,   126.25,  128.38,  125.37, 125.69,  122.25,  119.37,  118.5,   123.19,  123.5,
    122.19,  119.31,  123.31,  121.12, 123.37,  127.37,  128.5,   123.87,  122.94,  121.75,
    124.44,  122.0,   122.37,  122.94, 124.0,   123.19,  124.56,  127.25,  125.87,  128.86,
    132.0,   130.75,  134.75,  135.0,  132.38,  133.31,  131.94,  130.0,   125.37,  130.13,
    127.12,  125.19,  122.0,   125.0,  123.0,   123.5,   120.06,  121.0,   117.75,  119.87,
    122.0,   119.19,  116.37,  113.5,  114.25,  110.0,   105.06,  107.0,   107.87,  107.0,
    107.12,  107.0,   91.0,    93.94,  93.87,   95.5,    93.0,    94.94,   98.25,   96.75,
    94.81,   94.37,   91.56,   90.25,  93.94,   93.62,   97.0,    95.0,    95.87,   94.06,
    94.62,   93.75,   98.0,    103.94, 107.87,  106.06,  104.5,   105.0,   104.19,  103.06,
    103.42,  105.27,  111.87,  116.0,  116.62,  118.28,  113.37,  109.0,   109.7,   109.25,
    107.0,   109.19,  110.0,   109.2,  110.12,  108.0,   108.62,  109.75,  109.81,  109.0,
    108.75,  107.87,
};

const test_expected_atr = [_]f64{
    math.nan(f64),     math.nan(f64),     math.nan(f64),     math.nan(f64),     math.nan(f64),
    math.nan(f64),     math.nan(f64),     math.nan(f64),     math.nan(f64),     math.nan(f64),
    math.nan(f64),     math.nan(f64),     math.nan(f64),     math.nan(f64),     3.578214285714290,
    3.487627551020410, 3.559939868804670, 3.439587021032900, 3.388187948101980, 3.324745951808980,
    3.290478383822630, 3.196158499263870, 3.226790035030730, 3.226305032528540, 3.201211815919360,
    3.249339543353690, 3.245101004542710, 3.394736647075380, 3.413326886569990, 3.394874966100710,
    3.348812468522080, 3.263540149341940, 3.164358710103230, 3.184047373667280, 3.108401132691050,
    3.120658194641690, 3.063111180738710, 3.098960382114510, 3.042606069106340, 2.970277064170170,
    2.896685845300870, 3.055851142065090, 3.121147489060440, 3.003208382698980, 3.014407783934770,
    2.893021513653720, 2.844948548392740, 2.836023652078970, 2.818450534073330, 2.735346924496660,
    2.647107858461190, 2.640885868571100, 2.809394020816020, 2.746937305043450, 2.680370354683200,
    2.671772472205830, 2.614860152762560, 2.618084427565230, 2.658578397024860, 2.624751368665940,
    2.615840556618370, 2.536137659717060, 2.602842112594410, 2.579781961694810, 2.701583250145180,
    2.713970160849100, 2.665115149359880, 2.662249781548460, 2.617089082866420, 2.653368434090250,
    2.678127831655230, 2.758975843679860, 2.856548997702730, 2.893581212152530, 2.927968268427350,
    3.678470534968250, 3.902151211041950, 4.016283267396100, 3.970477319724950, 3.990443225458880,
    3.903982995068960, 3.895127066849750, 3.829046562074770, 3.716257521926570, 3.636167698931810,
    3.590727149008110, 3.628889495507530, 3.541825960114140, 3.445266962963130, 3.661319322751480,
    4.131939371126370, 4.193943701760200, 4.166876294491610, 3.998885130599360, 3.916464764127970,
    3.775288709547400, 3.655268087436870, 3.713463224048530, 3.626787279473630, 3.919159616654090,
    4.032433929750220, 3.927260077625210, 3.990312929223410, 4.075290577136020, 4.052055535912020,
    4.016908711918300, 4.136415232495570, 4.158099858745880, 4.168949868835460, 4.134739163918640,
    4.120829223638740, 3.969341421950260, 3.940102748953810, 4.060809695457110, 3.985037574353030,
    3.821106319042100, 3.905313010539090, 3.965647795500590, 3.856672952964830, 3.786910599181630,
    3.744274127811510, 3.655397404396400, 3.590726161225230, 3.879245721137710, 3.838585312485020,
    3.760829218736090, 3.711484274540660, 3.651378254930610, 3.631279808149850, 3.626188393282010,
    3.603603508047580, 3.591917543187040, 3.500352004387960, 3.509612575503110, 3.410354534395740,
    3.537472067653190, 3.896938348535100, 3.824299895068310, 3.943992759706290, 3.863707562584410,
    3.757014165256950, 3.738656010595740, 3.677323438410330, 3.683228907095310, 3.585855413731360,
    3.629008598464830, 3.721936555717340, 3.728226801737530, 3.921924887327710, 4.070358823947160,
    4.016761765093790, 3.984135924729950, 4.016697644392090, 3.930504955506940, 3.891183172970730,
    3.934670089187110, 3.796479368530890, 3.963159413635820, 3.938648026947550, 3.893744596451300,
    3.838477125276200, 3.810014473470760, 3.765727725365710, 3.768890030696730, 3.673969314218390,
    3.620828648917080, 3.585769459708710, 3.611071641158090, 3.567423666789660, 3.589750547733250,
    3.641196937180880, 3.613254298810810, 3.765878991752900, 3.733316206627690, 3.720936477582860,
    3.570869586326940, 3.534378901589300, 3.483351837190060, 3.783826705962200, 3.907124798393470,
    3.905187312793940, 3.831959647594370, 4.013248244194770, 4.052301941038000, 4.151423230963860,
    4.073464428752160, 4.032502683841290, 3.932323920709770, 4.048586497801930, 3.991544605101790,
    4.005719990451660, 3.983168562562260, 3.993656522379240, 4.087681056495010, 4.050703838173930,
    4.109224992590080, 4.234994635976500, 4.115352161978180, 4.178541293265460, 4.187931200889350,
    4.107364686540110, 4.246838637501530, 5.229207306251420, 5.114263927233460, 4.963245075288220,
    4.760870427053340, 4.617236825120960, 4.452434194755180, 4.424403180844090, 4.282660096498090,
    4.199612946748220, 4.073926307694780, 4.100074428573720, 4.003640540818460, 3.981237645045710,
    3.938292098971020, 3.902699806187370, 3.811792677173990, 3.771664628804420, 3.680831441032670,
    3.560772052387480, 3.467145477216950, 3.554492228844310, 3.809171355355430, 3.858516258544330,
    3.859336525791160, 3.887241059663220, 3.743866698258710, 3.592876219811650, 3.506242204110820,
    3.412224903817190, 3.378494553544530, 3.680030656862780, 3.747885609944010, 3.748036637805150,
    3.873176877961930, 4.301521386678930, 4.409269859059010, 4.536464869126220, 4.449574521331490,
    4.471033484093530, 4.433102520943990, 4.290738055162280, 4.176399622650690, 4.025228221032780,
    3.902711919530440, 3.798232496706840, 3.661215889799210, 3.520414754813550, 3.393956558041150,
    3.290102518181070, 3.260809481168140,
};

test "AverageTrueRange update length=14" {
    const tolerance = 1e-12;
    var atr = try AverageTrueRange.init(testing.allocator, .{ .length = 14 });
    defer atr.deinit();

    for (0..test_input_close.len) |i| {
        const act = atr.update(test_input_close[i], test_input_high[i], test_input_low[i]);
        const exp = test_expected_atr[i];

        if (math.isNan(exp)) {
            try testing.expect(math.isNan(act));
        } else {
            try testing.expect(!math.isNan(act));
            try testing.expect(almostEqual(act, exp, tolerance));
        }
    }
}

test "AverageTrueRange isPrimed length=5" {
    var atr = try AverageTrueRange.init(testing.allocator, .{ .length = 5 });
    defer atr.deinit();

    try testing.expect(!atr.isPrimed());

    for (0..5) |i| {
        _ = atr.update(test_input_close[i], test_input_high[i], test_input_low[i]);
        try testing.expect(!atr.isPrimed());
    }

    for (5..10) |i| {
        _ = atr.update(test_input_close[i], test_input_high[i], test_input_low[i]);
        try testing.expect(atr.isPrimed());
    }
}

test "AverageTrueRange constructor validation" {
    try testing.expectError(error.InvalidLength, AverageTrueRange.init(testing.allocator, .{ .length = 0 }));
    try testing.expectError(error.InvalidLength, AverageTrueRange.init(testing.allocator, .{ .length = -8 }));

    var atr = try AverageTrueRange.init(testing.allocator, .{ .length = 14 });
    defer atr.deinit();
    try testing.expect(!atr.isPrimed());
}

test "AverageTrueRange metadata" {
    var atr = try AverageTrueRange.init(testing.allocator, .{ .length = 14 });
    defer atr.deinit();
    var meta: Metadata = undefined;
    atr.getMetadata(&meta);

    try testing.expectEqual(Identifier.average_true_range, meta.identifier);
    try testing.expectEqualStrings("atr", meta.mnemonic);
    try testing.expectEqual(@as(usize, 1), meta.outputs_len);
}

test "AverageTrueRange NaN passthrough" {
    var atr = try AverageTrueRange.init(testing.allocator, .{ .length = 14 });
    defer atr.deinit();

    try testing.expect(math.isNan(atr.update(math.nan(f64), 1, 1)));
    try testing.expect(math.isNan(atr.update(1, math.nan(f64), 1)));
    try testing.expect(math.isNan(atr.update(1, 1, math.nan(f64))));
    try testing.expect(math.isNan(atr.updateSample(math.nan(f64))));
}

test "AverageTrueRange updateBar" {
    var atr = try AverageTrueRange.init(testing.allocator, .{ .length = 14 });
    defer atr.deinit();

    // Prime with 14 bars.
    for (0..14) |i| {
        _ = atr.update(test_input_close[i], test_input_high[i], test_input_low[i]);
    }

    const bar = Bar{
        .time = 1000,
        .open = 91,
        .high = test_input_high[14],
        .low = test_input_low[14],
        .close = test_input_close[14],
        .volume = 1000,
    };
    const out = atr.updateBar(&bar);
    try testing.expect(!math.isNan(out.slice()[0].scalar.value));
}
