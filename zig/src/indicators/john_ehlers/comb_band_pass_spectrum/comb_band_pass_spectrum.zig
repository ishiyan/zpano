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
const output_metadata_mod = @import("../../core/output_metadata.zig");
const shape_mod = @import("../../core/outputs/shape.zig");
const heatmap_mod = @import("../../core/outputs/heatmap.zig");

const OutputArray = indicator_mod.OutputArray;
const Identifier = identifier_mod.Identifier;
const Metadata = metadata_mod.Metadata;
const Heatmap = heatmap_mod.Heatmap;

const two_pi: f64 = 2.0 * math.pi;
const sqrt_two: f64 = math.sqrt2;
const agc_decay: f64 = 0.995;
const centroid_threshold: f64 = 0.5;
const bank_startup_bars: usize = 12;

/// Output indices for Ehlers' Comb Band-Pass Spectrum indicator.
pub const CombBandPassSpectrumOutput = enum(u8) {
    value = 1,
    dominant_cycle = 2,
};

/// Parameters to create a `CombBandPassSpectrum`.
///
/// Defaults correspond to Ehlers' 2013 Book Listing 10-1 settings.
/// Explicit values (including `0` or `false`) are validated directly without silent substitution.
pub const CombBandPassSpectrumParams = struct {
    min_period: i32 = 10,
    max_period: i32 = 48,
    bandwidth: f64 = 0.3,
    highpass_period: f64 = 48.0,
    smoothing_period: f64 = 10.0,
    spectral_dilation_compensation: bool = true,
    enhance_resolution: bool = false,
    bar_component: ?bar_component.BarComponent = null,
    quote_component: ?quote_component.QuoteComponent = null,
    trade_component: ?trade_component.TradeComponent = null,
};

pub const Params = CombBandPassSpectrumParams;

pub fn defaultParams() CombBandPassSpectrumParams {
    return .{};
}

pub const Error = error{
    InvalidMinPeriod,
    InvalidMaxPeriod,
    SpectrumRangeTooLarge,
    InvalidBandwidth,
    UnstableBandPassPoles,
    InvalidHighpassPeriod,
    InvalidSmoothingPeriod,
    UnstableRoofCoefficients,
    InvalidSample,
    NumericOverflow,
    MnemonicTooLong,
    OutOfMemory,
};

fn normalizeZero(value: f64) f64 {
    return if (value == 0.0) 0.0 else value;
}

fn formatMnemonicInternal(
    buffer: []u8,
    prefix: []const u8,
    include_enhance: bool,
    params: CombBandPassSpectrumParams,
) Error![]const u8 {
    const bar_comp = params.bar_component orelse bar_component.default_bar_component;
    const quote_comp = params.quote_component orelse quote_component.default_quote_component;
    const trade_comp = params.trade_component orelse trade_component.default_trade_component;

    var triple_storage: [64]u8 = undefined;
    const comp_suffix = component_triple_mnemonic_mod.componentTripleMnemonic(
        &triple_storage,
        bar_comp,
        quote_comp,
        trade_comp,
    );
    const sdc_suffix: []const u8 = if (params.spectral_dilation_compensation) "" else ", no-sdc";
    const enh_suffix: []const u8 = if (include_enhance and params.enhance_resolution) ", enh" else "";

    return std.fmt.bufPrint(buffer, "{s}({d}, {d}, {d}, {d}, {d}{s}{s}{s})", .{
        prefix,
        params.min_period,
        params.max_period,
        normalizeZero(params.bandwidth),
        normalizeZero(params.highpass_period),
        normalizeZero(params.smoothing_period),
        sdc_suffix,
        enh_suffix,
        comp_suffix,
    }) catch return error.MnemonicTooLong;
}

/// Builds the primary heatmap mnemonic into `buffer`.
pub fn mnemonic(buffer: []u8, params: CombBandPassSpectrumParams) Error![]const u8 {
    return formatMnemonicInternal(buffer, "cbps", true, params);
}

/// Builds the dominant-cycle mnemonic into `buffer`.
pub fn dominantCycleMnemonic(buffer: []u8, params: CombBandPassSpectrumParams) Error![]const u8 {
    return formatMnemonicInternal(buffer, "cbps-dc", false, params);
}

/// Spectrum heatmap column and dominant cycle period returned per observation.
pub const SpectrumOutput = struct {
    heatmap: Heatmap,
    dominant_cycle: f64,
};

/// Snapshot of committed internal roofing, spectral, and centroid state.
pub const SpectrumState = struct {
    high_pass: f64,
    filtered: f64,
    raw_power_storage: [heatmap_mod.max_heatmap_values]f64 = [_]f64{0.0} ** heatmap_mod.max_heatmap_values,
    bins_len: usize,
    normalization_maximum: f64,
    internal_dominant_cycle: f64,

    pub fn rawPowerSlice(self: *const SpectrumState) []const f64 {
        return self.raw_power_storage[0..self.bins_len];
    }

    pub fn eql(self: *const SpectrumState, other: *const SpectrumState) bool {
        if (self.bins_len != other.bins_len) return false;
        if (self.high_pass != other.high_pass) return false;
        if (self.filtered != other.filtered) return false;
        if (self.normalization_maximum != other.normalization_maximum) return false;
        if (self.internal_dominant_cycle != other.internal_dominant_cycle) return false;
        return std.mem.eql(f64, self.rawPowerSlice(), other.rawPowerSlice());
    }
};

/// Validated numeric parameters resolved at construction time.
pub const ResolvedParams = struct {
    min_period: usize,
    max_period: usize,
    bandwidth: f64,
    highpass_period: f64,
    smoothing_period: f64,
    spectral_dilation_compensation: bool,
    enhance_resolution: bool,
};

const RoofFilter = struct {
    hp0: f64,
    hp1: f64,
    hp2: f64,
    ss1: f64,
    ss2: f64,
    ss3: f64,
    x1: f64 = 0.0,
    x2: f64 = 0.0,
    h1: f64 = 0.0,
    h2: f64 = 0.0,
    f1: f64 = 0.0,
    f2: f64 = 0.0,

    fn init(highpass_period: f64, smoothing_period: f64) Error!RoofFilter {
        if (!math.isFinite(highpass_period) or highpass_period < 5.0) {
            return error.InvalidHighpassPeriod;
        }
        if (!math.isFinite(smoothing_period) or smoothing_period < 2.0) {
            return error.InvalidSmoothingPeriod;
        }

        const angle = two_pi / (sqrt_two * highpass_period);
        const cosine = @cos(angle);
        const alpha = (cosine + @sin(angle) - 1.0) / cosine;
        const pole = 1.0 - alpha;
        const attenuation = @exp(-sqrt_two * math.pi / smoothing_period);
        const hp0 = (1.0 - alpha / 2.0) * (1.0 - alpha / 2.0);
        const hp1 = 2.0 * pole;
        const hp2 = pole * pole;
        const ss2 = 2.0 * attenuation * @cos(sqrt_two * math.pi / smoothing_period);
        const ss3 = -attenuation * attenuation;
        const ss1 = 1.0 - ss2 - ss3;

        if (!(pole > 0.0 and pole < 1.0 and attenuation > 0.0 and attenuation < 1.0 and ss1 > 0.0) or
            !math.isFinite(hp0) or
            !math.isFinite(hp1) or
            !math.isFinite(hp2) or
            !math.isFinite(ss1) or
            !math.isFinite(ss2) or
            !math.isFinite(ss3))
        {
            return error.UnstableRoofCoefficients;
        }

        return .{
            .hp0 = hp0,
            .hp1 = hp1,
            .hp2 = hp2,
            .ss1 = ss1,
            .ss2 = ss2,
            .ss3 = ss3,
        };
    }

    fn reset(self: *RoofFilter) void {
        self.x1 = 0.0;
        self.x2 = 0.0;
        self.h1 = 0.0;
        self.h2 = 0.0;
        self.f1 = 0.0;
        self.f2 = 0.0;
    }

    fn candidate(self: *const RoofFilter, price: f64) Error!struct { high: f64, filtered: f64 } {
        const high = self.hp0 * (price - 2.0 * self.x1 + self.x2) +
            self.hp1 * self.h1 -
            self.hp2 * self.h2;
        const filtered = self.ss1 * (high + self.h1) / 2.0 +
            self.ss2 * self.f1 +
            self.ss3 * self.f2;
        if (!math.isFinite(high) or !math.isFinite(filtered)) {
            return error.NumericOverflow;
        }
        return .{ .high = high, .filtered = filtered };
    }

    fn commit(self: *RoofFilter, price: f64, high: f64, filtered: f64) void {
        self.x2 = self.x1;
        self.x1 = price;
        self.h2 = self.h1;
        self.h1 = high;
        self.f2 = self.f1;
        self.f1 = filtered;
    }
};

/// Ehlers' 2013 Book Listing 10-1 Comb Band-Pass Spectrum indicator.
pub const CombBandPassSpectrum = struct {
    allocator: std.mem.Allocator,
    resolved: ResolvedParams,
    bins_len: usize,
    roof: RoofFilter,
    alpha: [heatmap_mod.max_heatmap_values]f64 = [_]f64{0.0} ** heatmap_mod.max_heatmap_values,
    beta: [heatmap_mod.max_heatmap_values]f64 = [_]f64{0.0} ** heatmap_mod.max_heatmap_values,
    bp_offsets: [heatmap_mod.max_heatmap_values]usize = [_]usize{0} ** heatmap_mod.max_heatmap_values,
    bp: []f64,
    next_bp: []f64,
    raw: [heatmap_mod.max_heatmap_values]f64 = [_]f64{0.0} ** heatmap_mod.max_heatmap_values,
    sample_count: usize = 0,
    maximum: f64 = 0.0,
    dominant: f64 = 0.0,
    has_dominant: bool = false,
    last_output: SpectrumOutput,
    bar_func: bar_component.BarFunc,
    quote_func: quote_component.QuoteFunc,
    trade_func: trade_component.TradeFunc,
    mnemonic_storage: [output_metadata_mod.max_mnemonic_len]u8,
    mnemonic_len: usize,
    description_storage: [output_metadata_mod.max_description_len]u8,
    description_len: usize,
    mnemonic_dc_storage: [output_metadata_mod.max_mnemonic_len]u8,
    mnemonic_dc_len: usize,
    description_dc_storage: [output_metadata_mod.max_description_len]u8,
    description_dc_len: usize,

    pub fn init(
        allocator: std.mem.Allocator,
        params: CombBandPassSpectrumParams,
    ) Error!CombBandPassSpectrum {
        if (params.min_period < 3) {
            return error.InvalidMinPeriod;
        }
        if (params.max_period < params.min_period) {
            return error.InvalidMaxPeriod;
        }
        const span = @as(i64, params.max_period) - @as(i64, params.min_period) + 1;
        if (span > heatmap_mod.max_heatmap_values) {
            return error.SpectrumRangeTooLarge;
        }
        const min_period_f: f64 = @floatFromInt(params.min_period);
        if (!math.isFinite(params.bandwidth) or !(params.bandwidth > 0.0 and params.bandwidth < min_period_f / 4.0)) {
            return error.InvalidBandwidth;
        }

        const roof = try RoofFilter.init(params.highpass_period, params.smoothing_period);

        const min_period: usize = @intCast(params.min_period);
        const max_period: usize = @intCast(params.max_period);
        const bins_len: usize = @intCast(span);

        var alpha_table: [heatmap_mod.max_heatmap_values]f64 = [_]f64{0.0} ** heatmap_mod.max_heatmap_values;
        var beta_table: [heatmap_mod.max_heatmap_values]f64 = [_]f64{0.0} ** heatmap_mod.max_heatmap_values;
        var offsets: [heatmap_mod.max_heatmap_values]usize = [_]usize{0} ** heatmap_mod.max_heatmap_values;
        var total_bp_cells: usize = 0;

        for (0..bins_len) |bin_index| {
            const period = min_period + bin_index;
            const period_f: f64 = @floatFromInt(period);
            const beta_value = @cos(two_pi / period_f);
            const gamma_value = 1.0 / @cos(two_pi * params.bandwidth / period_f);
            const alpha_value = gamma_value - @sqrt(gamma_value * gamma_value - 1.0);
            if (!math.isFinite(alpha_value) or !(alpha_value > 0.0 and alpha_value < 1.0)) {
                return error.UnstableBandPassPoles;
            }
            alpha_table[bin_index] = alpha_value;
            beta_table[bin_index] = beta_value;
            offsets[bin_index] = total_bp_cells;
            total_bp_cells = math.add(usize, total_bp_cells, period) catch return error.OutOfMemory;
        }

        const bp = allocator.alloc(f64, total_bp_cells) catch return error.OutOfMemory;
        errdefer allocator.free(bp);
        const next_bp = allocator.alloc(f64, total_bp_cells) catch return error.OutOfMemory;
        errdefer allocator.free(next_bp);

        @memset(bp, 0.0);
        @memset(next_bp, 0.0);

        const bar_comp = params.bar_component orelse bar_component.default_bar_component;
        const quote_comp = params.quote_component orelse quote_component.default_quote_component;
        const trade_comp = params.trade_component orelse trade_component.default_trade_component;

        var mnemonic_storage: [output_metadata_mod.max_mnemonic_len]u8 = undefined;
        const mn = try mnemonic(&mnemonic_storage, params);

        var description_storage: [output_metadata_mod.max_description_len]u8 = undefined;
        const desc = std.fmt.bufPrint(&description_storage, "Comb band-pass spectrum {s}", .{mn}) catch
            return error.MnemonicTooLong;

        var mnemonic_dc_storage: [output_metadata_mod.max_mnemonic_len]u8 = undefined;
        const mn_dc = try dominantCycleMnemonic(&mnemonic_dc_storage, params);

        var description_dc_storage: [output_metadata_mod.max_description_len]u8 = undefined;
        const desc_dc = std.fmt.bufPrint(
            &description_dc_storage,
            "Comb band-pass spectrum dominant cycle {s}",
            .{mn_dc},
        ) catch return error.MnemonicTooLong;

        const resolved = ResolvedParams{
            .min_period = min_period,
            .max_period = max_period,
            .bandwidth = params.bandwidth,
            .highpass_period = params.highpass_period,
            .smoothing_period = params.smoothing_period,
            .spectral_dilation_compensation = params.spectral_dilation_compensation,
            .enhance_resolution = params.enhance_resolution,
        };

        return .{
            .allocator = allocator,
            .resolved = resolved,
            .bins_len = bins_len,
            .roof = roof,
            .alpha = alpha_table,
            .beta = beta_table,
            .bp_offsets = offsets,
            .bp = bp,
            .next_bp = next_bp,
            .last_output = .{
                .heatmap = Heatmap.empty(0, @floatFromInt(min_period), @floatFromInt(max_period), 1.0),
                .dominant_cycle = math.nan(f64),
            },
            .bar_func = bar_component.componentValue(bar_comp),
            .quote_func = quote_component.componentValue(quote_comp),
            .trade_func = trade_component.componentValue(trade_comp),
            .mnemonic_storage = mnemonic_storage,
            .mnemonic_len = mn.len,
            .description_storage = description_storage,
            .description_len = desc.len,
            .mnemonic_dc_storage = mnemonic_dc_storage,
            .mnemonic_dc_len = mn_dc.len,
            .description_dc_storage = description_dc_storage,
            .description_dc_len = desc_dc.len,
        };
    }

    pub fn initDefault(allocator: std.mem.Allocator) Error!CombBandPassSpectrum {
        return init(allocator, defaultParams());
    }

    pub fn deinit(self: *CombBandPassSpectrum) void {
        self.allocator.free(self.bp);
        self.allocator.free(self.next_bp);
    }

    pub fn reset(self: *CombBandPassSpectrum) void {
        self.roof.reset();
        self.sample_count = 0;
        @memset(self.bp, 0.0);
        @memset(self.next_bp, 0.0);
        @memset(&self.raw, 0.0);
        self.maximum = 0.0;
        self.dominant = 0.0;
        self.has_dominant = false;
        self.last_output = .{
            .heatmap = Heatmap.empty(
                0,
                @floatFromInt(self.resolved.min_period),
                @floatFromInt(self.resolved.max_period),
                1.0,
            ),
            .dominant_cycle = math.nan(f64),
        };
    }

    pub fn fixSlices(self: *CombBandPassSpectrum) void {
        _ = self;
    }

    pub fn mnemonicText(self: *const CombBandPassSpectrum) []const u8 {
        return self.mnemonic_storage[0..self.mnemonic_len];
    }

    fn descriptionText(self: *const CombBandPassSpectrum) []const u8 {
        return self.description_storage[0..self.description_len];
    }

    pub fn dominantCycleMnemonicText(self: *const CombBandPassSpectrum) []const u8 {
        return self.mnemonic_dc_storage[0..self.mnemonic_dc_len];
    }

    fn dominantCycleDescriptionText(self: *const CombBandPassSpectrum) []const u8 {
        return self.description_dc_storage[0..self.description_dc_len];
    }

    pub fn parameters(self: *const CombBandPassSpectrum) ResolvedParams {
        return self.resolved;
    }

    pub fn minPeriod(self: *const CombBandPassSpectrum) usize {
        return self.resolved.min_period;
    }

    pub fn maxPeriod(self: *const CombBandPassSpectrum) usize {
        return self.resolved.max_period;
    }

    pub fn warmupBars(self: *const CombBandPassSpectrum) usize {
        return bank_startup_bars + self.resolved.max_period - 1;
    }

    pub fn samplesSeen(self: *const CombBandPassSpectrum) usize {
        return self.sample_count;
    }

    pub fn bandPassRow(self: *const CombBandPassSpectrum, bin_index: usize) []const f64 {
        const period = self.resolved.min_period + bin_index;
        const offset = self.bp_offsets[bin_index];
        return self.bp[offset .. offset + period];
    }

    pub fn state(self: *const CombBandPassSpectrum) SpectrumState {
        var out_state = SpectrumState{
            .high_pass = self.roof.h1,
            .filtered = self.roof.f1,
            .bins_len = self.bins_len,
            .normalization_maximum = self.maximum,
            .internal_dominant_cycle = self.dominant,
        };
        @memcpy(out_state.raw_power_storage[0..self.bins_len], self.raw[0..self.bins_len]);
        return out_state;
    }

    pub fn output(self: *const CombBandPassSpectrum) SpectrumOutput {
        return self.last_output;
    }

    pub fn tryUpdate(self: *CombBandPassSpectrum, sample: f64, time: i64) Error!SpectrumOutput {
        if (math.isInf(sample)) {
            return error.InvalidSample;
        }
        const min_f: f64 = @floatFromInt(self.resolved.min_period);
        const max_f: f64 = @floatFromInt(self.resolved.max_period);
        if (math.isNan(sample)) {
            return .{
                .heatmap = Heatmap.empty(time, min_f, max_f, 1.0),
                .dominant_cycle = math.nan(f64),
            };
        }

        const roof_pair = try self.roof.candidate(sample);
        const difference = roof_pair.filtered - self.roof.f2;

        var next_raw: [heatmap_mod.max_heatmap_values]f64 = undefined;
        for (0..self.bins_len) |bin_index| {
            const period = self.resolved.min_period + bin_index;
            const offset = self.bp_offsets[bin_index];
            const old_row = self.bp[offset .. offset + period];
            var current: f64 = 0.0;
            if (self.sample_count + 1 > bank_startup_bars) {
                const a = self.alpha[bin_index];
                const b = self.beta[bin_index];
                current = 0.5 * (1.0 - a) * difference + b * (1.0 + a) * old_row[0] - a * old_row[1];
            }
            if (!math.isFinite(current)) {
                return error.NumericOverflow;
            }

            const next_row = self.next_bp[offset .. offset + period];
            next_row[0] = current;
            @memcpy(next_row[1..period], old_row[0 .. period - 1]);

            var power: f64 = 0.0;
            if (self.sample_count + 1 > bank_startup_bars) {
                const compensation: f64 = if (self.resolved.spectral_dilation_compensation)
                    @floatFromInt(period)
                else
                    1.0;
                for (0..period) |m| {
                    const scaled = next_row[m] / compensation;
                    power += scaled * scaled;
                }
            }
            next_raw[bin_index] = power;
        }

        var maximum = agc_decay * self.maximum;
        for (next_raw[0..self.bins_len]) |power| {
            if (!math.isFinite(power) or power < 0.0) {
                return error.NumericOverflow;
            }
            if (power > maximum) {
                maximum = power;
            }
        }

        var normalized: [heatmap_mod.max_heatmap_values]f64 = undefined;
        for (0..self.bins_len) |bin_index| {
            normalized[bin_index] = if (maximum > 0.0) next_raw[bin_index] / maximum else 0.0;
        }

        var sp: f64 = 0.0;
        var spx: f64 = 0.0;
        for (0..self.bins_len) |bin_index| {
            const power = normalized[bin_index];
            if (power >= centroid_threshold) {
                const period_f: f64 = @floatFromInt(self.resolved.min_period + bin_index);
                spx += period_f * power;
                sp += power;
            }
        }

        var dominant = self.dominant;
        var has_dominant = self.has_dominant;
        if (sp != 0.0) {
            dominant = spx / sp;
            has_dominant = true;
        }
        if (!math.isFinite(dominant)) {
            return error.NumericOverflow;
        }

        var display_values: [heatmap_mod.max_heatmap_values]f64 = undefined;
        if (self.resolved.enhance_resolution) {
            for (0..self.bins_len) |bin_index| {
                const item = normalized[bin_index];
                display_values[bin_index] = item * item * item;
            }
        } else {
            @memcpy(display_values[0..self.bins_len], normalized[0..self.bins_len]);
        }

        self.roof.commit(sample, roof_pair.high, roof_pair.filtered);
        @memcpy(self.bp, self.next_bp);
        @memcpy(self.raw[0..self.bins_len], next_raw[0..self.bins_len]);
        self.maximum = maximum;
        self.dominant = dominant;
        self.has_dominant = has_dominant;
        self.sample_count += 1;

        const current_output: SpectrumOutput = if (self.sample_count > self.warmupBars()) blk: {
            const active_values = display_values[0..self.bins_len];
            var value_min = active_values[0];
            var value_max = active_values[0];
            for (active_values[1..]) |item| {
                if (item < value_min) value_min = item;
                if (item > value_max) value_max = item;
            }
            break :blk .{
                .heatmap = Heatmap.new(time, min_f, max_f, 1.0, value_min, value_max, active_values),
                .dominant_cycle = if (has_dominant) dominant else math.nan(f64),
            };
        } else .{
            .heatmap = Heatmap.empty(time, min_f, max_f, 1.0),
            .dominant_cycle = math.nan(f64),
        };

        self.last_output = current_output;
        return current_output;
    }

    pub fn update(self: *CombBandPassSpectrum, sample: f64, time: i64) SpectrumOutput {
        return self.tryUpdate(sample, time) catch |err| {
            std.debug.panic("CombBandPassSpectrum.update failed: {s}", .{@errorName(err)});
        };
    }

    pub const updateSample = update;

    fn wrapEntity(time: i64, out_value: SpectrumOutput) OutputArray {
        var arr = OutputArray{};
        arr.append(.{ .heatmap = out_value.heatmap });
        arr.append(.{ .scalar = .{ .time = time, .value = out_value.dominant_cycle } });
        return arr;
    }

    pub fn updateScalar(self: *CombBandPassSpectrum, sample: *const Scalar) OutputArray {
        return wrapEntity(sample.time, self.update(sample.value, sample.time));
    }

    pub fn updateBar(self: *CombBandPassSpectrum, sample: *const Bar) OutputArray {
        return wrapEntity(sample.time, self.update(self.bar_func(sample.*), sample.time));
    }

    pub fn updateQuote(self: *CombBandPassSpectrum, sample: *const Quote) OutputArray {
        return wrapEntity(sample.time, self.update(self.quote_func(sample.*), sample.time));
    }

    pub fn updateTrade(self: *CombBandPassSpectrum, sample: *const Trade) OutputArray {
        return wrapEntity(sample.time, self.update(self.trade_func(sample.*), sample.time));
    }

    pub fn isPrimed(self: *const CombBandPassSpectrum) bool {
        return self.sample_count > self.warmupBars();
    }

    pub fn getMetadata(self: *const CombBandPassSpectrum, out: *Metadata) void {
        const mn = self.mnemonicText();
        const desc = self.descriptionText();
        const texts = [_]build_metadata_mod.OutputText{
            .{ .mnemonic = mn, .description = desc },
            .{
                .mnemonic = self.dominantCycleMnemonicText(),
                .description = self.dominantCycleDescriptionText(),
            },
        };
        build_metadata_mod.buildMetadata(out, .comb_band_pass_spectrum, mn, desc, &texts);
    }

    pub fn indicator(self: *CombBandPassSpectrum) indicator_mod.Indicator {
        return .{
            .ptr = @ptrCast(self),
            .vtable = &vtable,
        };
    }

    const vtable = indicator_mod.Indicator.GenVTable(CombBandPassSpectrum);
};

// ---------------------------------------------------------------------------
// Tests
// ---------------------------------------------------------------------------
const testing = std.testing;
const testdata = @import("testdata.zig");

const fixture_tolerance: f64 = 1e-11;
const oracle_tolerance: f64 = 2e-11;
const exact_tolerance: f64 = 1e-13;

fn expectCloseTol(actual: f64, expected: f64, tolerance: f64) !void {
    if (math.isNan(expected)) {
        try testing.expect(math.isNan(actual));
        return;
    }
    try testing.expect(math.isFinite(actual));
    const bound = @max(tolerance, @abs(expected) * tolerance);
    try testing.expect(@abs(actual - expected) <= bound);
}

fn expectClose(actual: f64, expected: f64) !void {
    try expectCloseTol(actual, expected, exact_tolerance);
}

fn expectSameOutput(actual: *const SpectrumOutput, expected: *const SpectrumOutput) !void {
    try testing.expectEqual(expected.heatmap.time, actual.heatmap.time);
    try testing.expectEqual(expected.heatmap.parameter_first, actual.heatmap.parameter_first);
    try testing.expectEqual(expected.heatmap.parameter_last, actual.heatmap.parameter_last);
    try testing.expectEqual(expected.heatmap.parameter_resolution, actual.heatmap.parameter_resolution);
    try testing.expectEqual(expected.heatmap.isEmpty(), actual.heatmap.isEmpty());
    try expectClose(actual.heatmap.value_min, expected.heatmap.value_min);
    try expectClose(actual.heatmap.value_max, expected.heatmap.value_max);
    const lhs = actual.heatmap.valuesSlice();
    const rhs = expected.heatmap.valuesSlice();
    try testing.expectEqual(rhs.len, lhs.len);
    for (0..lhs.len) |bin_index| {
        try expectClose(lhs[bin_index], rhs[bin_index]);
    }
    try expectClose(actual.dominant_cycle, expected.dominant_cycle);
}

const ListingOracle = struct {
    allocator: std.mem.Allocator,
    resolved: ResolvedParams,
    stride: usize,
    bp: []f64,
    prices: [252]f64 = [_]f64{0.0} ** 252,
    hp: [252]f64 = [_]f64{0.0} ** 252,
    filt: [252]f64 = [_]f64{0.0} ** 252,
    count: usize = 0,
    max_power: f64 = 0.0,
    dominant: f64 = 0.0,
    identified: bool = false,

    fn init(allocator: std.mem.Allocator, resolved: ResolvedParams) !ListingOracle {
        const bins_len = resolved.max_period - resolved.min_period + 1;
        const stride = resolved.max_period + 1;
        const bp = try allocator.alloc(f64, bins_len * stride);
        @memset(bp, 0.0);
        return .{
            .allocator = allocator,
            .resolved = resolved,
            .stride = stride,
            .bp = bp,
        };
    }

    fn deinit(self: *ListingOracle) void {
        self.allocator.free(self.bp);
    }

    fn lookback(series: []const f64, count: usize, lag: usize) f64 {
        return if (count > lag) series[count - 1 - lag] else 0.0;
    }

    fn update(
        self: *ListingOracle,
        price: f64,
        values_out: []f64,
    ) struct { dominant: f64, hp: f64, filt: f64 } {
        const theta = math.sqrt2 * math.pi / self.resolved.highpass_period;
        const alpha = (@cos(theta) + @sin(theta) - 1.0) / @cos(theta);
        const high_pass = (1.0 - alpha / 2.0) * (1.0 - alpha / 2.0) *
            (price - 2.0 * lookback(&self.prices, self.count, 0) + lookback(&self.prices, self.count, 1)) +
            2.0 * (1.0 - alpha) * lookback(&self.hp, self.count, 0) -
            (1.0 - alpha) * (1.0 - alpha) * lookback(&self.hp, self.count, 1);
        const a1 = @exp(-math.sqrt2 * math.pi / self.resolved.smoothing_period);
        const c2 = 2.0 * a1 * @cos(math.sqrt2 * math.pi / self.resolved.smoothing_period);
        const c3 = -a1 * a1;
        const c1 = 1.0 - c2 - c3;
        const filtered = c1 * (high_pass + lookback(&self.hp, self.count, 0)) / 2.0 +
            c2 * lookback(&self.filt, self.count, 0) +
            c3 * lookback(&self.filt, self.count, 1);

        self.prices[self.count] = price;
        self.hp[self.count] = high_pass;
        self.filt[self.count] = filtered;
        self.count += 1;

        const bins_len = self.resolved.max_period - self.resolved.min_period + 1;
        var powers: [heatmap_mod.max_heatmap_values]f64 = undefined;
        for (0..bins_len) |offset| {
            const period = self.resolved.min_period + offset;
            const period_f: f64 = @floatFromInt(period);
            const comp: f64 = if (self.resolved.spectral_dilation_compensation) period_f else 1.0;
            const row = self.bp[offset * self.stride .. (offset + 1) * self.stride];

            var m: usize = self.resolved.max_period;
            while (m > 1) : (m -= 1) {
                row[m] = row[m - 1];
            }
            row[1] = 0.0;
            if (self.count > 12) {
                const beta = @cos(2.0 * math.pi / period_f);
                const gamma = 1.0 / @cos(2.0 * math.pi * self.resolved.bandwidth / period_f);
                const pole = gamma - @sqrt(gamma * gamma - 1.0);
                row[1] = 0.5 * (1.0 - pole) * (filtered - lookback(&self.filt, self.count, 2)) +
                    beta * (1.0 + pole) * row[2] - pole * row[3];
            }

            var power: f64 = 0.0;
            for (1..period + 1) |sample_slot| {
                power += (row[sample_slot] / comp) * (row[sample_slot] / comp);
            }
            powers[offset] = power;
        }

        self.max_power = 0.995 * self.max_power;
        for (powers[0..bins_len]) |power| {
            if (power > self.max_power) self.max_power = power;
        }
        for (0..bins_len) |offset| {
            powers[offset] = if (self.max_power != 0.0) powers[offset] / self.max_power else 0.0;
        }

        var sp: f64 = 0.0;
        var spx: f64 = 0.0;
        for (0..bins_len) |offset| {
            const power = powers[offset];
            if (power >= 0.5) {
                const period_f: f64 = @floatFromInt(self.resolved.min_period + offset);
                spx += period_f * power;
                sp += power;
            }
        }
        if (sp != 0.0) {
            self.dominant = spx / sp;
            self.identified = true;
        }

        for (0..bins_len) |offset| {
            const item = powers[offset];
            values_out[offset] = if (self.resolved.enhance_resolution) item * item * item else item;
        }

        return .{
            .dominant = if (self.identified) self.dominant else math.nan(f64),
            .hp = high_pass,
            .filt = filtered,
        };
    }
};

fn verifyReferenceCase(
    comptime bins_len: usize,
    params: CombBandPassSpectrumParams,
    expected_heatmap: *const [252][bins_len]f64,
    expected_min: *const [252]f64,
    expected_max: *const [252]f64,
    expected_dc: *const [252]f64,
) !void {
    var instance = try CombBandPassSpectrum.init(testing.allocator, params);
    defer instance.deinit();
    instance.fixSlices();

    var oracle = try ListingOracle.init(testing.allocator, instance.parameters());
    defer oracle.deinit();
    var oracle_values: [bins_len]f64 = undefined;
    const warmup = instance.warmupBars();
    const prices = testdata.testInput();

    for (prices, 0..) |price, bar_index| {
        const current = instance.update(price, @intCast(bar_index));
        const oracle_step = oracle.update(price, &oracle_values);
        const current_state = instance.state();

        try expectCloseTol(current_state.high_pass, oracle_step.hp, oracle_tolerance);
        try expectCloseTol(current_state.filtered, oracle_step.filt, oracle_tolerance);
        try testing.expectEqual(bar_index >= warmup, instance.isPrimed());
        try testing.expectEqual(bar_index < warmup, current.heatmap.isEmpty());
        try testing.expectEqual(@as(i64, @intCast(bar_index)), current.heatmap.time);
        try expectClose(current.heatmap.parameter_first, @floatFromInt(params.min_period));
        try expectClose(current.heatmap.parameter_last, @floatFromInt(params.max_period));
        try expectClose(current.heatmap.parameter_resolution, 1.0);

        if (bar_index < warmup) {
            for (expected_heatmap[bar_index]) |item| {
                try testing.expect(math.isNan(item));
            }
            try testing.expect(math.isNan(current.heatmap.value_min));
            try testing.expect(math.isNan(current.heatmap.value_max));
            try testing.expect(math.isNan(current.dominant_cycle));
            try testing.expect(math.isNan(expected_min[bar_index]));
            try testing.expect(math.isNan(expected_max[bar_index]));
            try testing.expect(math.isNan(expected_dc[bar_index]));
            continue;
        }

        const values = current.heatmap.valuesSlice();
        try testing.expectEqual(bins_len, values.len);
        for (0..bins_len) |bin_index| {
            try expectCloseTol(values[bin_index], expected_heatmap[bar_index][bin_index], fixture_tolerance);
            try expectCloseTol(values[bin_index], oracle_values[bin_index], oracle_tolerance);
        }
        try expectCloseTol(current.heatmap.value_min, expected_min[bar_index], fixture_tolerance);
        try expectCloseTol(current.heatmap.value_max, expected_max[bar_index], fixture_tolerance);
        try expectCloseTol(current.dominant_cycle, expected_dc[bar_index], fixture_tolerance);
        try expectCloseTol(current.dominant_cycle, oracle_step.dominant, oracle_tolerance);
    }
}

test "CombBandPassSpectrum matches all 12 reference cases and Listing 10-1 oracle" {
    {
        const hm = testdata.expectedP010_P148_HP48_LP10_SDC1_E0_B0_3_HEATMAP();
        const mn = testdata.expectedP010_P148_HP48_LP10_SDC1_E0_B0_3_MIN();
        const mx = testdata.expectedP010_P148_HP48_LP10_SDC1_E0_B0_3_MAX();
        const dc = testdata.expectedP010_P148_HP48_LP10_SDC1_E0_B0_3_DOMINANT_CYCLE();
        try verifyReferenceCase(39, .{}, &hm, &mn, &mx, &dc);
    }
    {
        const hm = testdata.expectedP010_P148_HP48_LP10_SDC0_E0_B0_3_HEATMAP();
        const mn = testdata.expectedP010_P148_HP48_LP10_SDC0_E0_B0_3_MIN();
        const mx = testdata.expectedP010_P148_HP48_LP10_SDC0_E0_B0_3_MAX();
        const dc = testdata.expectedP010_P148_HP48_LP10_SDC0_E0_B0_3_DOMINANT_CYCLE();
        try verifyReferenceCase(39, .{ .spectral_dilation_compensation = false }, &hm, &mn, &mx, &dc);
    }
    {
        const hm = testdata.expectedP010_P148_HP48_LP10_SDC1_E1_B0_3_HEATMAP();
        const mn = testdata.expectedP010_P148_HP48_LP10_SDC1_E1_B0_3_MIN();
        const mx = testdata.expectedP010_P148_HP48_LP10_SDC1_E1_B0_3_MAX();
        const dc = testdata.expectedP010_P148_HP48_LP10_SDC1_E1_B0_3_DOMINANT_CYCLE();
        try verifyReferenceCase(39, .{ .enhance_resolution = true }, &hm, &mn, &mx, &dc);
    }
    {
        const hm = testdata.expectedP010_P148_HP48_LP10_SDC0_E1_B0_3_HEATMAP();
        const mn = testdata.expectedP010_P148_HP48_LP10_SDC0_E1_B0_3_MIN();
        const mx = testdata.expectedP010_P148_HP48_LP10_SDC0_E1_B0_3_MAX();
        const dc = testdata.expectedP010_P148_HP48_LP10_SDC0_E1_B0_3_DOMINANT_CYCLE();
        try verifyReferenceCase(39, .{ .spectral_dilation_compensation = false, .enhance_resolution = true }, &hm, &mn, &mx, &dc);
    }
    {
        const hm = testdata.expectedP010_P148_HP48_LP10_SDC1_E0_B0_1_HEATMAP();
        const mn = testdata.expectedP010_P148_HP48_LP10_SDC1_E0_B0_1_MIN();
        const mx = testdata.expectedP010_P148_HP48_LP10_SDC1_E0_B0_1_MAX();
        const dc = testdata.expectedP010_P148_HP48_LP10_SDC1_E0_B0_1_DOMINANT_CYCLE();
        try verifyReferenceCase(39, .{ .bandwidth = 0.1 }, &hm, &mn, &mx, &dc);
    }
    {
        const hm = testdata.expectedP010_P148_HP48_LP10_SDC1_E0_B0_7_HEATMAP();
        const mn = testdata.expectedP010_P148_HP48_LP10_SDC1_E0_B0_7_MIN();
        const mx = testdata.expectedP010_P148_HP48_LP10_SDC1_E0_B0_7_MAX();
        const dc = testdata.expectedP010_P148_HP48_LP10_SDC1_E0_B0_7_DOMINANT_CYCLE();
        try verifyReferenceCase(39, .{ .bandwidth = 0.7 }, &hm, &mn, &mx, &dc);
    }
    {
        const hm = testdata.expectedP010_P148_HP48_LP10_5_SDC1_E0_B0_3_HEATMAP();
        const mn = testdata.expectedP010_P148_HP48_LP10_5_SDC1_E0_B0_3_MIN();
        const mx = testdata.expectedP010_P148_HP48_LP10_5_SDC1_E0_B0_3_MAX();
        const dc = testdata.expectedP010_P148_HP48_LP10_5_SDC1_E0_B0_3_DOMINANT_CYCLE();
        try verifyReferenceCase(39, .{ .smoothing_period = 10.5 }, &hm, &mn, &mx, &dc);
    }
    {
        const hm = testdata.expectedP010_P148_HP80_LP20_SDC1_E0_B0_3_HEATMAP();
        const mn = testdata.expectedP010_P148_HP80_LP20_SDC1_E0_B0_3_MIN();
        const mx = testdata.expectedP010_P148_HP80_LP20_SDC1_E0_B0_3_MAX();
        const dc = testdata.expectedP010_P148_HP80_LP20_SDC1_E0_B0_3_DOMINANT_CYCLE();
        try verifyReferenceCase(39, .{ .highpass_period = 80.0, .smoothing_period = 20.0 }, &hm, &mn, &mx, &dc);
    }
    {
        const hm = testdata.expectedP06_P124_HP24_LP6_SDC1_E0_B0_3_HEATMAP();
        const mn = testdata.expectedP06_P124_HP24_LP6_SDC1_E0_B0_3_MIN();
        const mx = testdata.expectedP06_P124_HP24_LP6_SDC1_E0_B0_3_MAX();
        const dc = testdata.expectedP06_P124_HP24_LP6_SDC1_E0_B0_3_DOMINANT_CYCLE();
        try verifyReferenceCase(19, .{ .min_period = 6, .max_period = 24, .highpass_period = 24.0, .smoothing_period = 6.0 }, &hm, &mn, &mx, &dc);
    }
    {
        const hm = testdata.expectedP020_P120_HP48_LP10_SDC1_E0_B0_3_HEATMAP();
        const mn = testdata.expectedP020_P120_HP48_LP10_SDC1_E0_B0_3_MIN();
        const mx = testdata.expectedP020_P120_HP48_LP10_SDC1_E0_B0_3_MAX();
        const dc = testdata.expectedP020_P120_HP48_LP10_SDC1_E0_B0_3_DOMINANT_CYCLE();
        try verifyReferenceCase(1, .{ .min_period = 20, .max_period = 20 }, &hm, &mn, &mx, &dc);
    }
    {
        const hm = testdata.expectedP024_P160_HP60_LP12_SDC1_E0_B0_3_HEATMAP();
        const mn = testdata.expectedP024_P160_HP60_LP12_SDC1_E0_B0_3_MIN();
        const mx = testdata.expectedP024_P160_HP60_LP12_SDC1_E0_B0_3_MAX();
        const dc = testdata.expectedP024_P160_HP60_LP12_SDC1_E0_B0_3_DOMINANT_CYCLE();
        try verifyReferenceCase(37, .{ .min_period = 24, .max_period = 60, .highpass_period = 60.0, .smoothing_period = 12.0 }, &hm, &mn, &mx, &dc);
    }
    {
        const hm = testdata.expectedP010_P1240_HP240_LP10_SDC1_E0_B0_3_HEATMAP();
        const mn = testdata.expectedP010_P1240_HP240_LP10_SDC1_E0_B0_3_MIN();
        const mx = testdata.expectedP010_P1240_HP240_LP10_SDC1_E0_B0_3_MAX();
        const dc = testdata.expectedP010_P1240_HP240_LP10_SDC1_E0_B0_3_DOMINANT_CYCLE();
        try verifyReferenceCase(231, .{ .min_period = 10, .max_period = 240, .highpass_period = 240.0, .smoothing_period = 10.0 }, &hm, &mn, &mx, &dc);
    }
}

test "CombBandPassSpectrum 12-bar gate, NaN, infinity, overflow, reset, and sine" {
    const prices = testdata.testInput();
    var gated = try CombBandPassSpectrum.initDefault(testing.allocator);
    defer gated.deinit();

    for (0..12) |index| {
        _ = gated.update(prices[index], @intCast(index));
        for (0..39) |bin_index| {
            for (gated.bandPassRow(bin_index)) |cell| {
                try testing.expectEqual(@as(f64, 0.0), cell);
            }
        }
    }
    _ = gated.update(prices[12], 12);
    var any_nonzero = false;
    for (0..39) |bin_index| {
        if (gated.bandPassRow(bin_index)[0] != 0.0) {
            any_nonzero = true;
            break;
        }
    }
    try testing.expect(any_nonzero);

    var instance = try CombBandPassSpectrum.initDefault(testing.allocator);
    defer instance.deinit();
    var control = try CombBandPassSpectrum.initDefault(testing.allocator);
    defer control.deinit();

    for (0..120) |index| {
        _ = instance.update(prices[index], @intCast(index));
        _ = control.update(prices[index], @intCast(index));
    }

    const saved_state = instance.state();
    const saved_output = instance.output();
    const saved_count = instance.samplesSeen();

    const nan_out = instance.update(math.nan(f64), 999);
    try testing.expect(nan_out.heatmap.isEmpty());
    try testing.expectEqual(@as(i64, 999), nan_out.heatmap.time);
    try testing.expect(math.isNan(nan_out.dominant_cycle));
    try testing.expect(instance.state().eql(&saved_state));
    try expectSameOutput(&instance.output(), &saved_output);
    try testing.expectEqual(saved_count, instance.samplesSeen());

    try testing.expectError(error.InvalidSample, instance.tryUpdate(math.inf(f64), 1000));
    try testing.expectError(error.InvalidSample, instance.tryUpdate(-math.inf(f64), 1001));
    try testing.expectError(error.NumericOverflow, instance.tryUpdate(1e308, 1002));
    try testing.expect(instance.state().eql(&saved_state));
    try expectSameOutput(&instance.output(), &saved_output);
    try testing.expectEqual(saved_count, instance.samplesSeen());

    for (120..150) |index| {
        const lhs = instance.update(prices[index], @intCast(index));
        const rhs = control.update(prices[index], @intCast(index));
        try expectSameOutput(&lhs, &rhs);
    }

    instance.reset();
    try testing.expect(!instance.isPrimed());
    try testing.expectEqual(@as(usize, 0), instance.samplesSeen());
    try testing.expect(instance.output().heatmap.isEmpty());
    try testing.expect(math.isNan(instance.output().dominant_cycle));

    var fresh = try CombBandPassSpectrum.initDefault(testing.allocator);
    defer fresh.deinit();
    for (0..80) |index| {
        const lhs = instance.update(prices[index], @intCast(index));
        const rhs = fresh.update(prices[index], @intCast(index));
        try expectSameOutput(&lhs, &rhs);
    }

    var compensated = try CombBandPassSpectrum.initDefault(testing.allocator);
    defer compensated.deinit();
    var uncompensated = try CombBandPassSpectrum.init(testing.allocator, .{ .spectral_dilation_compensation = false });
    defer uncompensated.deinit();
    var enhanced = try CombBandPassSpectrum.init(testing.allocator, .{ .enhance_resolution = true });
    defer enhanced.deinit();

    for (prices, 0..) |price, bar_index| {
        const out_comp = compensated.update(price, @intCast(bar_index));
        _ = uncompensated.update(price, @intCast(bar_index));
        const out_enh = enhanced.update(price, @intCast(bar_index));
        try testing.expect(compensated.state().eql(&enhanced.state()));
        try expectClose(out_comp.dominant_cycle, out_enh.dominant_cycle);
        const comp_raw = compensated.state().rawPowerSlice();
        const uncomp_raw = uncompensated.state().rawPowerSlice();
        for (0..39) |bin_index| {
            const period_f: f64 = @floatFromInt(10 + bin_index);
            try expectClose(comp_raw[bin_index], uncomp_raw[bin_index] / (period_f * period_f));
        }
        if (!out_comp.heatmap.isEmpty()) {
            const plain_vals = out_comp.heatmap.valuesSlice();
            const enh_vals = out_enh.heatmap.valuesSlice();
            for (0..39) |bin_index| {
                const item = plain_vals[bin_index];
                try expectClose(enh_vals[bin_index], item * item * item);
            }
        }
    }

    var sine_ind = try CombBandPassSpectrum.initDefault(testing.allocator);
    defer sine_ind.deinit();
    var sine_out: SpectrumOutput = undefined;
    for (0..1000) |step| {
        const sample = @sin(2.0 * math.pi * @as(f64, @floatFromInt(step)) / 20.0);
        sine_out = sine_ind.update(sample, @intCast(step));
    }
    const sine_vals = sine_out.heatmap.valuesSlice();
    var best_bin: usize = 0;
    for (1..sine_vals.len) |bin_index| {
        if (sine_vals[bin_index] > sine_vals[best_bin]) best_bin = bin_index;
    }
    const peak_period: i32 = @as(i32, @intCast(10 + best_bin));
    try testing.expect(@abs(peak_period - 20) <= 2);
    try testing.expect(sine_out.dominant_cycle > 17.0 and sine_out.dominant_cycle < 23.0);
}

test "CombBandPassSpectrum metadata, mnemonics, entities, and validation" {
    try testing.expectEqual(@as(u8, 1), @intFromEnum(CombBandPassSpectrumOutput.value));
    try testing.expectEqual(@as(u8, 2), @intFromEnum(CombBandPassSpectrumOutput.dominant_cycle));

    var default_ind = try CombBandPassSpectrum.initDefault(testing.allocator);
    defer default_ind.deinit();
    var md: Metadata = undefined;
    default_ind.getMetadata(&md);
    try testing.expectEqual(Identifier.comb_band_pass_spectrum, md.identifier);
    try testing.expectEqualStrings("cbps(10, 48, 0.3, 48, 10)", md.mnemonic);
    try testing.expectEqual(@as(usize, 2), md.outputs_len);
    try testing.expectEqual(shape_mod.Shape.heatmap, md.outputs_buf[0].shape);
    try testing.expectEqualStrings("cbps(10, 48, 0.3, 48, 10)", md.outputs_buf[0].mnemonic);
    try testing.expectEqual(shape_mod.Shape.scalar, md.outputs_buf[1].shape);
    try testing.expectEqualStrings("cbps-dc(10, 48, 0.3, 48, 10)", md.outputs_buf[1].mnemonic);

    const MnemonicCase = struct {
        params: CombBandPassSpectrumParams,
        expected: []const u8,
        expected_dc: []const u8,
    };
    const mnemonic_cases = [_]MnemonicCase{
        .{
            .params = .{},
            .expected = "cbps(10, 48, 0.3, 48, 10)",
            .expected_dc = "cbps-dc(10, 48, 0.3, 48, 10)",
        },
        .{
            .params = .{ .bar_component = .open },
            .expected = "cbps(10, 48, 0.3, 48, 10, o)",
            .expected_dc = "cbps-dc(10, 48, 0.3, 48, 10, o)",
        },
        .{
            .params = .{ .quote_component = .bid },
            .expected = "cbps(10, 48, 0.3, 48, 10, b)",
            .expected_dc = "cbps-dc(10, 48, 0.3, 48, 10, b)",
        },
        .{
            .params = .{ .trade_component = .volume },
            .expected = "cbps(10, 48, 0.3, 48, 10, v)",
            .expected_dc = "cbps-dc(10, 48, 0.3, 48, 10, v)",
        },
        .{
            .params = .{ .bar_component = .high, .quote_component = .ask },
            .expected = "cbps(10, 48, 0.3, 48, 10, h, a)",
            .expected_dc = "cbps-dc(10, 48, 0.3, 48, 10, h, a)",
        },
        .{
            .params = .{ .bar_component = .low, .trade_component = .volume },
            .expected = "cbps(10, 48, 0.3, 48, 10, l, v)",
            .expected_dc = "cbps-dc(10, 48, 0.3, 48, 10, l, v)",
        },
        .{
            .params = .{ .quote_component = .spread_bp, .trade_component = .volume },
            .expected = "cbps(10, 48, 0.3, 48, 10, spread bp, v)",
            .expected_dc = "cbps-dc(10, 48, 0.3, 48, 10, spread bp, v)",
        },
        .{
            .params = .{
                .min_period = 6,
                .max_period = 24,
                .bandwidth = 0.4,
                .highpass_period = 48.0,
                .smoothing_period = 10.5,
                .spectral_dilation_compensation = false,
                .enhance_resolution = true,
                .bar_component = .median,
                .quote_component = .weighted,
                .trade_component = .volume,
            },
            .expected = "cbps(6, 24, 0.4, 48, 10.5, no-sdc, enh, hl/2, (bbs+aas)/(bs+as), v)",
            .expected_dc = "cbps-dc(6, 24, 0.4, 48, 10.5, no-sdc, hl/2, (bbs+aas)/(bs+as), v)",
        },
    };
    for (mnemonic_cases) |case_item| {
        var mn_storage: [output_metadata_mod.max_mnemonic_len]u8 = undefined;
        var mn_dc_storage: [output_metadata_mod.max_mnemonic_len]u8 = undefined;
        const got_mn = try mnemonic(&mn_storage, case_item.params);
        const got_dc = try dominantCycleMnemonic(&mn_dc_storage, case_item.params);
        try testing.expectEqualStrings(case_item.expected, got_mn);
        try testing.expectEqualStrings(case_item.expected_dc, got_dc);
    }

    const prices = testdata.testInput();
    const highs = testdata.testHigh();
    const lows = testdata.testLow();
    const opens = testdata.testOpen();
    var bar_ind = try CombBandPassSpectrum.initDefault(testing.allocator);
    defer bar_ind.deinit();
    var scalar_ind = try CombBandPassSpectrum.initDefault(testing.allocator);
    defer scalar_ind.deinit();
    var quote_ind = try CombBandPassSpectrum.initDefault(testing.allocator);
    defer quote_ind.deinit();
    var trade_ind = try CombBandPassSpectrum.initDefault(testing.allocator);
    defer trade_ind.deinit();

    for (0..70) |index| {
        const ts: i64 = @intCast(index);
        const bar = Bar{
            .time = ts,
            .open = opens[index],
            .high = highs[index],
            .low = lows[index],
            .close = prices[index],
            .volume = 1000.0,
        };
        const scalar = Scalar{ .time = ts, .value = prices[index] };
        const quote = Quote{
            .time = ts,
            .bid_price = prices[index] - 0.5,
            .ask_price = prices[index] + 0.5,
            .bid_size = 10.0,
            .ask_size = 10.0,
        };
        const trade = Trade{
            .time = ts,
            .price = prices[index],
            .volume = 50.0,
        };

        const out_bar = bar_ind.updateBar(&bar);
        const out_scalar = scalar_ind.updateScalar(&scalar);
        const out_quote = quote_ind.updateQuote(&quote);
        const out_trade = trade_ind.updateTrade(&trade);

        try testing.expectEqual(@as(usize, 2), out_bar.len);
        const hm_bar = out_bar.values[0].heatmap;
        const hm_scalar = out_scalar.values[0].heatmap;
        const hm_quote = out_quote.values[0].heatmap;
        const hm_trade = out_trade.values[0].heatmap;

        try testing.expectEqual(hm_scalar.values_len, hm_bar.values_len);
        try testing.expectEqual(hm_quote.values_len, hm_bar.values_len);
        try testing.expectEqual(hm_trade.values_len, hm_bar.values_len);
        for (0..hm_bar.values_len) |bin_index| {
            try expectClose(hm_bar.values[bin_index], hm_scalar.values[bin_index]);
            try expectClose(hm_bar.values[bin_index], hm_quote.values[bin_index]);
            try expectClose(hm_bar.values[bin_index], hm_trade.values[bin_index]);
        }
        try expectClose(out_bar.values[1].scalar.value, out_scalar.values[1].scalar.value);
        try expectClose(out_bar.values[1].scalar.value, out_quote.values[1].scalar.value);
        try expectClose(out_bar.values[1].scalar.value, out_trade.values[1].scalar.value);
    }

    try testing.expectError(error.InvalidMinPeriod, CombBandPassSpectrum.init(testing.allocator, .{ .min_period = 2 }));
    try testing.expectError(error.InvalidMinPeriod, CombBandPassSpectrum.init(testing.allocator, .{ .min_period = 0 }));
    try testing.expectError(error.InvalidMaxPeriod, CombBandPassSpectrum.init(testing.allocator, .{ .min_period = 10, .max_period = 9 }));
    try testing.expectError(error.SpectrumRangeTooLarge, CombBandPassSpectrum.init(testing.allocator, .{ .min_period = 3, .max_period = 300 }));
    try testing.expectError(error.InvalidBandwidth, CombBandPassSpectrum.init(testing.allocator, .{ .bandwidth = 0.0 }));
    try testing.expectError(error.InvalidBandwidth, CombBandPassSpectrum.init(testing.allocator, .{ .bandwidth = -0.1 }));
    try testing.expectError(error.InvalidBandwidth, CombBandPassSpectrum.init(testing.allocator, .{ .bandwidth = 2.5 }));
    try testing.expectError(error.InvalidBandwidth, CombBandPassSpectrum.init(testing.allocator, .{ .min_period = 3, .max_period = 10, .bandwidth = 0.75 }));
    try testing.expectError(error.InvalidBandwidth, CombBandPassSpectrum.init(testing.allocator, .{ .bandwidth = math.nan(f64) }));
    try testing.expectError(error.InvalidBandwidth, CombBandPassSpectrum.init(testing.allocator, .{ .bandwidth = math.inf(f64) }));
    try testing.expectError(error.InvalidHighpassPeriod, CombBandPassSpectrum.init(testing.allocator, .{ .highpass_period = 4.0 }));
    try testing.expectError(error.InvalidSmoothingPeriod, CombBandPassSpectrum.init(testing.allocator, .{ .smoothing_period = 1.0 }));
}
