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
const inverse_sqrt_two: f64 = 1.0 / sqrt_two;
const dft_lag_start: usize = 3;
const power_smoothing: f64 = 0.2;
const power_retention: f64 = 0.8;
const book_agc_decay: f64 = 0.995;
const article_threshold: f64 = 0.25;
const book_threshold: f64 = 0.5;

/// Output indices for Ehlers' Autocorrelation Periodogram indicator.
pub const AutoCorrelationPeriodogramOutput = enum(u8) {
    value = 1,
    dominant_cycle = 2,
};

/// Selects the Ehlers source profile for AutoCorrelationPeriodogram.
pub const AutoCorrelationPeriodogramSource = enum(u8) {
    article = 0,
    book = 1,

    pub fn fromStr(text: []const u8) ?AutoCorrelationPeriodogramSource {
        if (std.mem.eql(u8, text, "article")) return .article;
        if (std.mem.eql(u8, text, "book")) return .book;
        return null;
    }

    pub fn asStr(self: AutoCorrelationPeriodogramSource) []const u8 {
        return switch (self) {
            .article => "article",
            .book => "book",
        };
    }
};

pub const Source = AutoCorrelationPeriodogramSource;

/// Parameters to create an `AutoCorrelationPeriodogram`.
///
/// `min_period = null` selects `8` for `.article` and `10` for `.book`.
/// `averaging_length = 0` selects lag-dependent Pearson windows (`M = lag`).
pub const AutoCorrelationPeriodogramParams = struct {
    source: AutoCorrelationPeriodogramSource = .article,
    min_period: ?i32 = null,
    max_period: i32 = 48,
    averaging_length: i32 = 3,
    enhance_resolution: bool = false,
    bar_component: ?bar_component.BarComponent = null,
    quote_component: ?quote_component.QuoteComponent = null,
    trade_component: ?trade_component.TradeComponent = null,
};

pub const Params = AutoCorrelationPeriodogramParams;

pub fn defaultParams() AutoCorrelationPeriodogramParams {
    return .{};
}

pub const Error = error{
    InvalidMinPeriod,
    InvalidMaxPeriod,
    InvalidAveragingLength,
    SpectrumRangeTooLarge,
    UnstableCoefficients,
    InvalidSample,
    NumericOverflow,
    MnemonicTooLong,
    OutOfMemory,
};

pub fn resolveMinPeriod(source: AutoCorrelationPeriodogramSource, min_period: ?i32) i32 {
    if (min_period) |explicit_min| {
        return explicit_min;
    }
    return switch (source) {
        .article => 8,
        .book => 10,
    };
}

fn formatMnemonicInternal(
    buffer: []u8,
    prefix: []const u8,
    include_book_enhance: bool,
    params: AutoCorrelationPeriodogramParams,
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
    const resolved_min = resolveMinPeriod(params.source, params.min_period);
    const is_book = params.source == .book;
    const source_tag: []const u8 = if (is_book) ", book" else "";
    const enh_tag: []const u8 = if (params.enhance_resolution and (!is_book or include_book_enhance))
        ", enh"
    else
        "";

    return std.fmt.bufPrint(buffer, "{s}({d}, {d}, {d}{s}{s}{s})", .{
        prefix,
        resolved_min,
        params.max_period,
        params.averaging_length,
        source_tag,
        enh_tag,
        comp_suffix,
    }) catch return error.MnemonicTooLong;
}

/// Builds the primary heatmap mnemonic into `buffer`.
pub fn mnemonic(buffer: []u8, params: AutoCorrelationPeriodogramParams) Error![]const u8 {
    return formatMnemonicInternal(buffer, "acp", true, params);
}

/// Builds the dominant-cycle mnemonic into `buffer`.
pub fn dominantCycleMnemonic(buffer: []u8, params: AutoCorrelationPeriodogramParams) Error![]const u8 {
    return formatMnemonicInternal(buffer, "acp-dc", false, params);
}

/// Periodogram heatmap column and dominant cycle period returned per observation.
pub const PeriodogramOutput = struct {
    heatmap: Heatmap,
    dominant_cycle: f64,
};

/// Snapshot of committed internal filter, correlation, spectral, and centroid state.
pub const PeriodogramState = struct {
    high_pass: f64,
    filtered: f64,
    correlations: []const f64,
    smoothed_power_storage: [heatmap_mod.max_heatmap_values]f64 = [_]f64{0.0} ** heatmap_mod.max_heatmap_values,
    bins_len: usize,
    normalization_maximum: f64,
    internal_dominant_cycle: f64,

    pub fn smoothedPowerSlice(self: *const PeriodogramState) []const f64 {
        return self.smoothed_power_storage[0..self.bins_len];
    }
};

/// Validated parameters resolved at construction time.
pub const ResolvedParams = struct {
    source: AutoCorrelationPeriodogramSource,
    min_period: usize,
    max_period: usize,
    averaging_length: usize,
    enhance_resolution: bool,
};

/// Ehlers' Autocorrelation Periodogram indicator (TASC Sept 2016 / Book Listing 8-3).
pub const AutoCorrelationPeriodogram = struct {
    allocator: std.mem.Allocator,
    resolved: ResolvedParams,
    history_length: usize,
    bins_len: usize,
    lag_span: usize,
    hp0: f64,
    hp1: f64,
    hp2: f64,
    ss1: f64,
    ss2: f64,
    ss3: f64,
    cosines: []f64,
    sines: []f64,
    filtered: []f64,
    next_filtered: []f64,
    correlations: []f64,
    next_correlations: []f64,
    power: [heatmap_mod.max_heatmap_values]f64 = [_]f64{0.0} ** heatmap_mod.max_heatmap_values,
    sample_count: usize = 0,
    close1: f64 = 0.0,
    close2: f64 = 0.0,
    high1: f64 = 0.0,
    high2: f64 = 0.0,
    maximum: f64 = 0.0,
    dominant: f64 = 0.0,
    has_dominant: bool = false,
    last_output: PeriodogramOutput,
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
        params: AutoCorrelationPeriodogramParams,
    ) Error!AutoCorrelationPeriodogram {
        const resolved_min_i32 = resolveMinPeriod(params.source, params.min_period);
        if (resolved_min_i32 < 3) {
            return error.InvalidMinPeriod;
        }
        if (params.max_period < 5 or params.max_period <= resolved_min_i32) {
            return error.InvalidMaxPeriod;
        }
        if (params.averaging_length < 0 or params.averaging_length == 1) {
            return error.InvalidAveragingLength;
        }
        const span = @as(i64, params.max_period) - @as(i64, resolved_min_i32) + 1;
        if (span > heatmap_mod.max_heatmap_values) {
            return error.SpectrumRangeTooLarge;
        }

        const min_period: usize = @intCast(resolved_min_i32);
        const max_period: usize = @intCast(params.max_period);
        const averaging_length: usize = @intCast(params.averaging_length);
        const bins_len: usize = @intCast(span);
        const lag_span: usize = max_period - dft_lag_start + 1;

        var pole: f64 = 0.0;
        var hp0: f64 = 0.0;
        var hp1: f64 = 0.0;
        var hp2: f64 = 0.0;
        const max_f: f64 = @floatFromInt(max_period);
        const min_f: f64 = @floatFromInt(min_period);

        if (params.source == .article) {
            const theta = two_pi / max_f;
            pole = (1.0 - @sin(theta)) / @cos(theta);
            hp0 = 0.5 * (1.0 + pole);
            hp1 = pole;
            hp2 = 0.0;
        } else {
            const theta = inverse_sqrt_two * two_pi / max_f;
            const cosine = @cos(theta);
            const alpha = (cosine + @sin(theta) - 1.0) / cosine;
            pole = 1.0 - alpha;
            hp0 = (1.0 - alpha / 2.0) * (1.0 - alpha / 2.0);
            hp1 = 2.0 * pole;
            hp2 = pole * pole;
        }

        const decay = @exp(-sqrt_two * math.pi / min_f);
        const ss2 = 2.0 * decay * @cos(sqrt_two * math.pi / min_f);
        const ss3 = -decay * decay;
        const ss1 = 1.0 - ss2 - ss3;

        if (!(pole > 0.0 and pole < 1.0 and decay > 0.0 and decay < 1.0 and ss1 > 0.0) or
            !math.isFinite(hp0) or
            !math.isFinite(hp1) or
            !math.isFinite(hp2) or
            !math.isFinite(ss1) or
            !math.isFinite(ss2) or
            !math.isFinite(ss3))
        {
            return error.UnstableCoefficients;
        }

        const effective_window: usize = if (averaging_length == 0) max_period else averaging_length;
        const sum_history = math.add(usize, max_period, effective_window) catch return error.OutOfMemory;
        const history_length: usize = @max(@as(usize, 3), sum_history);
        const corr_len = math.add(usize, max_period, 1) catch return error.OutOfMemory;
        const table_size = math.mul(usize, bins_len, lag_span) catch return error.OutOfMemory;

        const cosines = allocator.alloc(f64, table_size) catch return error.OutOfMemory;
        errdefer allocator.free(cosines);
        const sines = allocator.alloc(f64, table_size) catch return error.OutOfMemory;
        errdefer allocator.free(sines);
        const filtered = allocator.alloc(f64, history_length) catch return error.OutOfMemory;
        errdefer allocator.free(filtered);
        const next_filtered = allocator.alloc(f64, history_length) catch return error.OutOfMemory;
        errdefer allocator.free(next_filtered);
        const correlations = allocator.alloc(f64, corr_len) catch return error.OutOfMemory;
        errdefer allocator.free(correlations);
        const next_correlations = allocator.alloc(f64, corr_len) catch return error.OutOfMemory;
        errdefer allocator.free(next_correlations);

        @memset(filtered, 0.0);
        @memset(next_filtered, 0.0);
        @memset(correlations, 0.0);
        @memset(next_correlations, 0.0);

        for (0..bins_len) |bin_index| {
            const period_f: f64 = @floatFromInt(min_period + bin_index);
            const row_offset = bin_index * lag_span;
            for (0..lag_span) |offset| {
                const lag_f: f64 = @floatFromInt(dft_lag_start + offset);
                const angle = two_pi * lag_f / period_f;
                cosines[row_offset + offset] = @cos(angle);
                sines[row_offset + offset] = @sin(angle);
            }
        }

        const bar_comp = params.bar_component orelse bar_component.default_bar_component;
        const quote_comp = params.quote_component orelse quote_component.default_quote_component;
        const trade_comp = params.trade_component orelse trade_component.default_trade_component;

        var mnemonic_storage: [output_metadata_mod.max_mnemonic_len]u8 = undefined;
        const mn = try mnemonic(&mnemonic_storage, params);

        var description_storage: [output_metadata_mod.max_description_len]u8 = undefined;
        const desc = std.fmt.bufPrint(&description_storage, "Autocorrelation periodogram {s}", .{mn}) catch
            return error.MnemonicTooLong;

        var mnemonic_dc_storage: [output_metadata_mod.max_mnemonic_len]u8 = undefined;
        const mn_dc = try dominantCycleMnemonic(&mnemonic_dc_storage, params);

        var description_dc_storage: [output_metadata_mod.max_description_len]u8 = undefined;
        const desc_dc = std.fmt.bufPrint(
            &description_dc_storage,
            "Autocorrelation periodogram dominant cycle {s}",
            .{mn_dc},
        ) catch return error.MnemonicTooLong;

        return .{
            .allocator = allocator,
            .resolved = .{
                .source = params.source,
                .min_period = min_period,
                .max_period = max_period,
                .averaging_length = averaging_length,
                .enhance_resolution = params.enhance_resolution,
            },
            .history_length = history_length,
            .bins_len = bins_len,
            .lag_span = lag_span,
            .hp0 = hp0,
            .hp1 = hp1,
            .hp2 = hp2,
            .ss1 = ss1,
            .ss2 = ss2,
            .ss3 = ss3,
            .cosines = cosines,
            .sines = sines,
            .filtered = filtered,
            .next_filtered = next_filtered,
            .correlations = correlations,
            .next_correlations = next_correlations,
            .last_output = .{
                .heatmap = Heatmap.empty(0, min_f, max_f, 1.0),
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

    pub fn initDefault(allocator: std.mem.Allocator) Error!AutoCorrelationPeriodogram {
        return init(allocator, defaultParams());
    }

    pub fn deinit(self: *AutoCorrelationPeriodogram) void {
        self.allocator.free(self.cosines);
        self.allocator.free(self.sines);
        self.allocator.free(self.filtered);
        self.allocator.free(self.next_filtered);
        self.allocator.free(self.correlations);
        self.allocator.free(self.next_correlations);
    }

    pub fn reset(self: *AutoCorrelationPeriodogram) void {
        self.sample_count = 0;
        self.close1 = 0.0;
        self.close2 = 0.0;
        self.high1 = 0.0;
        self.high2 = 0.0;
        @memset(self.filtered, 0.0);
        @memset(self.next_filtered, 0.0);
        @memset(self.correlations, 0.0);
        @memset(self.next_correlations, 0.0);
        @memset(&self.power, 0.0);
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

    pub fn fixSlices(self: *AutoCorrelationPeriodogram) void {
        _ = self;
    }

    pub fn mnemonicText(self: *const AutoCorrelationPeriodogram) []const u8 {
        return self.mnemonic_storage[0..self.mnemonic_len];
    }

    fn descriptionText(self: *const AutoCorrelationPeriodogram) []const u8 {
        return self.description_storage[0..self.description_len];
    }

    pub fn dominantCycleMnemonicText(self: *const AutoCorrelationPeriodogram) []const u8 {
        return self.mnemonic_dc_storage[0..self.mnemonic_dc_len];
    }

    fn dominantCycleDescriptionText(self: *const AutoCorrelationPeriodogram) []const u8 {
        return self.description_dc_storage[0..self.description_dc_len];
    }

    pub fn parameters(self: *const AutoCorrelationPeriodogram) ResolvedParams {
        return self.resolved;
    }

    pub fn minPeriod(self: *const AutoCorrelationPeriodogram) usize {
        return self.resolved.min_period;
    }

    pub fn maxPeriod(self: *const AutoCorrelationPeriodogram) usize {
        return self.resolved.max_period;
    }

    pub fn warmupBars(self: *const AutoCorrelationPeriodogram) usize {
        return self.history_length - 1;
    }

    pub fn samplesSeen(self: *const AutoCorrelationPeriodogram) usize {
        return self.sample_count;
    }

    pub fn state(self: *const AutoCorrelationPeriodogram) PeriodogramState {
        var out_state = PeriodogramState{
            .high_pass = self.high1,
            .filtered = self.filtered[0],
            .correlations = self.correlations,
            .bins_len = self.bins_len,
            .normalization_maximum = self.maximum,
            .internal_dominant_cycle = self.dominant,
        };
        @memcpy(out_state.smoothed_power_storage[0..self.bins_len], self.power[0..self.bins_len]);
        return out_state;
    }

    pub fn output(self: *const AutoCorrelationPeriodogram) PeriodogramOutput {
        return self.last_output;
    }

    pub fn tryUpdate(self: *AutoCorrelationPeriodogram, sample: f64, time: i64) Error!PeriodogramOutput {
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

        const is_book = self.resolved.source == .book;
        const high = if (!is_book)
            self.hp0 * (sample - self.close1) + self.hp1 * self.high1
        else
            self.hp0 * (sample - 2.0 * self.close1 + self.close2) +
                self.hp1 * self.high1 -
                self.hp2 * self.high2;
        const filtered_value = self.ss1 * (high + self.high1) / 2.0 +
            self.ss2 * self.filtered[0] +
            self.ss3 * self.filtered[1];
        if (!math.isFinite(high) or !math.isFinite(filtered_value)) {
            return error.NumericOverflow;
        }

        self.next_filtered[0] = filtered_value;
        @memcpy(
            self.next_filtered[1..self.history_length],
            self.filtered[0 .. self.history_length - 1],
        );

        @memcpy(self.next_correlations, self.correlations);
        for (0..self.resolved.max_period + 1) |lag| {
            const window_size: usize = if (self.resolved.averaging_length != 0)
                self.resolved.averaging_length
            else
                lag;
            var sx: f64 = 0.0;
            var sy: f64 = 0.0;
            var sxx: f64 = 0.0;
            var syy: f64 = 0.0;
            var sxy: f64 = 0.0;
            for (0..window_size) |offset| {
                const x_sample = self.next_filtered[offset];
                const y_sample = self.next_filtered[lag + offset];
                sx += x_sample;
                sy += y_sample;
                sxx += x_sample * x_sample;
                sxy += x_sample * y_sample;
                syy += y_sample * y_sample;
            }
            const window_f: f64 = @floatFromInt(window_size);
            const denominator = (window_f * sxx - sx * sx) * (window_f * syy - sy * sy);
            if (!math.isFinite(sx) or
                !math.isFinite(sy) or
                !math.isFinite(sxx) or
                !math.isFinite(syy) or
                !math.isFinite(sxy) or
                !math.isFinite(denominator))
            {
                return error.NumericOverflow;
            }
            if (denominator > 0.0) {
                self.next_correlations[lag] = (window_f * sxy - sx * sy) / @sqrt(denominator);
            }
            if (!math.isFinite(self.next_correlations[lag])) {
                return error.NumericOverflow;
            }
        }

        var next_power: [heatmap_mod.max_heatmap_values]f64 = undefined;
        for (0..self.bins_len) |bin_index| {
            var cosine_part: f64 = 0.0;
            var sine_part: f64 = 0.0;
            const row_offset = bin_index * self.lag_span;
            for (0..self.lag_span) |offset| {
                const corr_value = self.next_correlations[dft_lag_start + offset];
                cosine_part += corr_value * self.cosines[row_offset + offset];
                sine_part += corr_value * self.sines[row_offset + offset];
            }
            const squared_sum = cosine_part * cosine_part + sine_part * sine_part;
            const smoothed = power_smoothing * squared_sum * squared_sum +
                power_retention * self.power[bin_index];
            if (!math.isFinite(smoothed)) {
                return error.NumericOverflow;
            }
            next_power[bin_index] = smoothed;
        }

        var maximum: f64 = if (is_book) book_agc_decay * self.maximum else 0.0;
        for (next_power[0..self.bins_len]) |power_value| {
            if (power_value > maximum) {
                maximum = power_value;
            }
        }

        var normalized: [heatmap_mod.max_heatmap_values]f64 = undefined;
        for (0..self.bins_len) |bin_index| {
            normalized[bin_index] = if (maximum > 0.0) next_power[bin_index] / maximum else 0.0;
        }

        var weights: [heatmap_mod.max_heatmap_values]f64 = undefined;
        if (!is_book and self.resolved.enhance_resolution) {
            for (0..self.bins_len) |bin_index| {
                const norm_value = normalized[bin_index];
                weights[bin_index] = norm_value * norm_value * norm_value;
            }
        } else {
            @memcpy(weights[0..self.bins_len], normalized[0..self.bins_len]);
        }

        const threshold = if (is_book) book_threshold else article_threshold;
        var peak = weights[0];
        for (weights[1..self.bins_len]) |weight| {
            if (weight > peak) peak = weight;
        }

        var sp: f64 = 0.0;
        var spx: f64 = 0.0;
        for (0..self.bins_len) |bin_index| {
            const weight = weights[bin_index];
            if (weight >= threshold and (is_book or peak >= article_threshold)) {
                const period_f: f64 = @floatFromInt(self.resolved.min_period + bin_index);
                spx += period_f * weight;
                sp += weight;
            }
        }

        var dominant = self.dominant;
        var has_dominant = self.has_dominant;
        if (sp != 0.0 and (is_book or sp >= article_threshold)) {
            dominant = spx / sp;
            has_dominant = true;
        }
        if (!math.isFinite(dominant)) {
            return error.NumericOverflow;
        }

        var display_values: [heatmap_mod.max_heatmap_values]f64 = undefined;
        if (is_book and self.resolved.enhance_resolution) {
            for (0..self.bins_len) |bin_index| {
                const norm_value = normalized[bin_index];
                display_values[bin_index] = norm_value * norm_value;
            }
        } else {
            @memcpy(display_values[0..self.bins_len], weights[0..self.bins_len]);
        }

        self.close2 = self.close1;
        self.close1 = sample;
        self.high2 = self.high1;
        self.high1 = high;
        @memcpy(self.filtered, self.next_filtered);
        @memcpy(self.correlations, self.next_correlations);
        @memcpy(self.power[0..self.bins_len], next_power[0..self.bins_len]);
        self.maximum = maximum;
        self.dominant = dominant;
        self.has_dominant = has_dominant;
        self.sample_count += 1;

        const current_output: PeriodogramOutput = if (self.sample_count >= self.history_length) blk: {
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

    pub fn update(self: *AutoCorrelationPeriodogram, sample: f64, time: i64) PeriodogramOutput {
        return self.tryUpdate(sample, time) catch |err| {
            std.debug.panic("AutoCorrelationPeriodogram.update failed: {s}", .{@errorName(err)});
        };
    }

    pub const updateSample = update;

    fn wrapEntity(time: i64, out_value: PeriodogramOutput) OutputArray {
        var arr = OutputArray{};
        arr.append(.{ .heatmap = out_value.heatmap });
        arr.append(.{ .scalar = .{ .time = time, .value = out_value.dominant_cycle } });
        return arr;
    }

    pub fn updateScalar(self: *AutoCorrelationPeriodogram, sample: *const Scalar) OutputArray {
        return wrapEntity(sample.time, self.update(sample.value, sample.time));
    }

    pub fn updateBar(self: *AutoCorrelationPeriodogram, sample: *const Bar) OutputArray {
        return wrapEntity(sample.time, self.update(self.bar_func(sample.*), sample.time));
    }

    pub fn updateQuote(self: *AutoCorrelationPeriodogram, sample: *const Quote) OutputArray {
        return wrapEntity(sample.time, self.update(self.quote_func(sample.*), sample.time));
    }

    pub fn updateTrade(self: *AutoCorrelationPeriodogram, sample: *const Trade) OutputArray {
        return wrapEntity(sample.time, self.update(self.trade_func(sample.*), sample.time));
    }

    pub fn isPrimed(self: *const AutoCorrelationPeriodogram) bool {
        return self.sample_count >= self.history_length;
    }

    pub fn getMetadata(self: *const AutoCorrelationPeriodogram, out: *Metadata) void {
        const mn = self.mnemonicText();
        const desc = self.descriptionText();
        const texts = [_]build_metadata_mod.OutputText{
            .{ .mnemonic = mn, .description = desc },
            .{
                .mnemonic = self.dominantCycleMnemonicText(),
                .description = self.dominantCycleDescriptionText(),
            },
        };
        build_metadata_mod.buildMetadata(out, .auto_correlation_periodogram, mn, desc, &texts);
    }

    pub fn indicator(self: *AutoCorrelationPeriodogram) indicator_mod.Indicator {
        return .{
            .ptr = @ptrCast(self),
            .vtable = &vtable,
        };
    }

    const vtable = indicator_mod.Indicator.GenVTable(AutoCorrelationPeriodogram);
};

// ---------------------------------------------------------------------------
// Tests
// ---------------------------------------------------------------------------
const testing = std.testing;
const testdata = @import("testdata.zig");

const fixture_tolerance: f64 = 1e-11;
const oracle_tolerance: f64 = 2e-10;
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

const SourceOracle = struct {
    resolved: ResolvedParams,
    prices: [252]f64 = [_]f64{0.0} ** 252,
    hp: [252]f64 = [_]f64{0.0} ** 252,
    filt: [252]f64 = [_]f64{0.0} ** 252,
    correlations: [65]f64 = [_]f64{0.0} ** 65,
    powers: [65]f64 = [_]f64{0.0} ** 65,
    count: usize = 0,
    maximum: f64 = 0.0,
    dominant: f64 = 0.0,
    has_dominant: bool = false,

    fn degToRad(deg: f64) f64 {
        return deg * math.pi / 180.0;
    }

    fn oldAt(series: []const f64, signed_index: isize) f64 {
        return if (signed_index >= 0) series[@intCast(signed_index)] else 0.0;
    }

    fn step(
        self: *SourceOracle,
        price: f64,
        display_out: []f64,
    ) struct { hp: f64, filt: f64, max: f64, dom: f64, public_dc: f64 } {
        const book = self.resolved.source == .book;
        const step_i: isize = @intCast(self.count);
        const min_f: f64 = @floatFromInt(self.resolved.min_period);
        const max_f: f64 = @floatFromInt(self.resolved.max_period);

        const ss_angle = degToRad(math.sqrt2 * 180.0 / min_f);
        const decay = @exp(-math.sqrt2 * math.pi / min_f);
        const c2 = 2.0 * decay * @cos(ss_angle);
        const c3 = -decay * decay;
        const c1 = 1.0 - c2 - c3;

        const art_angle = degToRad(360.0 / max_f);
        const art_pole = (1.0 - @sin(art_angle)) / @cos(art_angle);
        const book_angle = degToRad((1.0 / math.sqrt2) * 360.0 / max_f);
        const alpha = (@cos(book_angle) + @sin(book_angle) - 1.0) / @cos(book_angle);

        self.prices[self.count] = price;
        const high_value = if (book)
            (1.0 - alpha / 2.0) * (1.0 - alpha / 2.0) *
                (price - 2.0 * oldAt(&self.prices, step_i - 1) + oldAt(&self.prices, step_i - 2)) +
                2.0 * (1.0 - alpha) * oldAt(&self.hp, step_i - 1) -
                (1.0 - alpha) * (1.0 - alpha) * oldAt(&self.hp, step_i - 2)
        else
            0.5 * (1.0 + art_pole) * (price - oldAt(&self.prices, step_i - 1)) +
                art_pole * oldAt(&self.hp, step_i - 1);
        self.hp[self.count] = high_value;

        const filtered_value = c1 * (high_value + oldAt(&self.hp, step_i - 1)) / 2.0 +
            c2 * oldAt(&self.filt, step_i - 1) +
            c3 * oldAt(&self.filt, step_i - 2);
        self.filt[self.count] = filtered_value;

        for (0..self.resolved.max_period + 1) |lag| {
            const window_size = if (self.resolved.averaging_length != 0) self.resolved.averaging_length else lag;
            var sx: f64 = 0.0;
            var sy: f64 = 0.0;
            var sxx: f64 = 0.0;
            var syy: f64 = 0.0;
            var sxy: f64 = 0.0;
            const lag_i: isize = @intCast(lag);
            for (0..window_size) |offset| {
                const off_i: isize = @intCast(offset);
                const x_sample = oldAt(&self.filt, step_i - off_i);
                const y_sample = oldAt(&self.filt, step_i - lag_i - off_i);
                sx += x_sample;
                sy += y_sample;
                sxx += x_sample * x_sample;
                sxy += x_sample * y_sample;
                syy += y_sample * y_sample;
            }
            const window_f: f64 = @floatFromInt(window_size);
            const denominator = (window_f * sxx - sx * sx) * (window_f * syy - sy * sy);
            if (denominator > 0.0) {
                self.correlations[lag] = (window_f * sxy - sx * sy) / @sqrt(denominator);
            }
        }

        const bins_len = self.resolved.max_period - self.resolved.min_period + 1;
        for (0..bins_len) |bin_index| {
            const period_f: f64 = @floatFromInt(self.resolved.min_period + bin_index);
            var cosine_part: f64 = 0.0;
            var sine_part: f64 = 0.0;
            for (3..self.resolved.max_period + 1) |lag| {
                const angle = degToRad(360.0 * @as(f64, @floatFromInt(lag)) / period_f);
                cosine_part += self.correlations[lag] * @cos(angle);
                sine_part += self.correlations[lag] * @sin(angle);
            }
            const squared_sum = cosine_part * cosine_part + sine_part * sine_part;
            self.powers[bin_index] = 0.2 * squared_sum * squared_sum + 0.8 * self.powers[bin_index];
        }

        self.maximum = if (book) 0.995 * self.maximum else 0.0;
        for (self.powers[0..bins_len]) |power_value| {
            if (power_value > self.maximum) self.maximum = power_value;
        }
        var normalized: [65]f64 = undefined;
        var weights: [65]f64 = undefined;
        for (0..bins_len) |bin_index| {
            const norm_value = if (self.maximum > 0.0) self.powers[bin_index] / self.maximum else 0.0;
            normalized[bin_index] = norm_value;
            weights[bin_index] = if (!book and self.resolved.enhance_resolution)
                norm_value * norm_value * norm_value
            else
                norm_value;
        }
        var peak = weights[0];
        for (weights[1..bins_len]) |weight| {
            if (weight > peak) peak = weight;
        }
        var spx: f64 = 0.0;
        var sp: f64 = 0.0;
        for (0..bins_len) |bin_index| {
            const weight = weights[bin_index];
            if ((book and weight >= 0.5) or (!book and peak >= 0.25 and weight >= 0.25)) {
                spx += @as(f64, @floatFromInt(self.resolved.min_period + bin_index)) * weight;
                sp += weight;
            }
        }
        var candidate: f64 = if (sp != 0.0) spx / sp else 0.0;
        if ((book and sp == 0.0) or (!book and sp < 0.25)) {
            candidate = self.dominant;
        } else if (sp != 0.0) {
            self.has_dominant = true;
        }
        self.dominant = candidate;
        for (0..bins_len) |bin_index| {
            display_out[bin_index] = if (book and self.resolved.enhance_resolution)
                normalized[bin_index] * normalized[bin_index]
            else
                weights[bin_index];
        }
        self.count += 1;
        return .{
            .hp = high_value,
            .filt = filtered_value,
            .max = self.maximum,
            .dom = self.dominant,
            .public_dc = if (self.has_dominant) self.dominant else math.nan(f64),
        };
    }
};

fn verifyReferenceCase(
    comptime bins_len: usize,
    params: AutoCorrelationPeriodogramParams,
    expected_heatmap: *const [252][bins_len]f64,
    expected_min: *const [252]f64,
    expected_max: *const [252]f64,
    expected_dc: *const [252]f64,
) !void {
    var instance = try AutoCorrelationPeriodogram.init(testing.allocator, params);
    defer instance.deinit();
    instance.fixSlices();

    var oracle = SourceOracle{ .resolved = instance.parameters() };
    var oracle_display: [bins_len]f64 = undefined;
    const warmup = instance.warmupBars();
    const prices = testdata.testInput();
    const resolved_min_f: f64 = @floatFromInt(instance.minPeriod());
    const resolved_max_f: f64 = @floatFromInt(instance.maxPeriod());

    for (prices, 0..) |price, bar_index| {
        const current = instance.update(price, @intCast(bar_index));
        const oracle_step = oracle.step(price, &oracle_display);
        const current_state = instance.state();

        try expectCloseTol(current_state.high_pass, oracle_step.hp, oracle_tolerance);
        try expectCloseTol(current_state.filtered, oracle_step.filt, oracle_tolerance);
        try expectCloseTol(current_state.normalization_maximum, oracle_step.max, oracle_tolerance);
        try expectCloseTol(current_state.internal_dominant_cycle, oracle_step.dom, oracle_tolerance);
        try testing.expectEqual(bar_index >= warmup, instance.isPrimed());
        try testing.expectEqual(bar_index < warmup, current.heatmap.isEmpty());
        try testing.expectEqual(@as(i64, @intCast(bar_index)), current.heatmap.time);
        try expectCloseTol(current.heatmap.parameter_first, resolved_min_f, exact_tolerance);
        try expectCloseTol(current.heatmap.parameter_last, resolved_max_f, exact_tolerance);
        try expectCloseTol(current.heatmap.parameter_resolution, 1.0, exact_tolerance);

        if (bar_index < warmup) {
            for (expected_heatmap[bar_index]) |cell| {
                try testing.expect(math.isNan(cell));
            }
            try testing.expect(math.isNan(current.heatmap.value_min));
            try testing.expect(math.isNan(current.heatmap.value_max));
            try testing.expect(math.isNan(current.dominant_cycle));
            try testing.expect(math.isNan(expected_min[bar_index]));
            try testing.expect(math.isNan(expected_max[bar_index]));
            try testing.expect(math.isNan(expected_dc[bar_index]));
            continue;
        }

        const active_values = current.heatmap.valuesSlice();
        try testing.expectEqual(bins_len, active_values.len);
        for (0..bins_len) |bin_index| {
            try expectCloseTol(active_values[bin_index], expected_heatmap[bar_index][bin_index], fixture_tolerance);
            try expectCloseTol(active_values[bin_index], oracle_display[bin_index], oracle_tolerance);
        }
        try expectCloseTol(current.heatmap.value_min, expected_min[bar_index], fixture_tolerance);
        try expectCloseTol(current.heatmap.value_max, expected_max[bar_index], fixture_tolerance);
        try expectCloseTol(current.dominant_cycle, expected_dc[bar_index], fixture_tolerance);
        try expectCloseTol(current.dominant_cycle, oracle_step.public_dc, oracle_tolerance);
    }
}

test "AutoCorrelationPeriodogram matches all 64 reference cases and degree-based oracle" {
    {
        const expected_hm = testdata.expectedV0_L8_H24_A0_E0_HEATMAP();
        const expected_min = testdata.expectedV0_L8_H24_A0_E0_MIN();
        const expected_max = testdata.expectedV0_L8_H24_A0_E0_MAX();
        const expected_dc = testdata.expectedV0_L8_H24_A0_E0_DC();
        try verifyReferenceCase(
            17,
            .{ .source = .article, .min_period = 8, .max_period = 24, .averaging_length = 0, .enhance_resolution = false },
            &expected_hm,
            &expected_min,
            &expected_max,
            &expected_dc,
        );
    }
    {
        const expected_hm = testdata.expectedV0_L8_H24_A0_E1_HEATMAP();
        const expected_min = testdata.expectedV0_L8_H24_A0_E1_MIN();
        const expected_max = testdata.expectedV0_L8_H24_A0_E1_MAX();
        const expected_dc = testdata.expectedV0_L8_H24_A0_E1_DC();
        try verifyReferenceCase(
            17,
            .{ .source = .article, .min_period = 8, .max_period = 24, .averaging_length = 0, .enhance_resolution = true },
            &expected_hm,
            &expected_min,
            &expected_max,
            &expected_dc,
        );
    }
    {
        const expected_hm = testdata.expectedV0_L8_H24_A3_E0_HEATMAP();
        const expected_min = testdata.expectedV0_L8_H24_A3_E0_MIN();
        const expected_max = testdata.expectedV0_L8_H24_A3_E0_MAX();
        const expected_dc = testdata.expectedV0_L8_H24_A3_E0_DC();
        try verifyReferenceCase(
            17,
            .{ .source = .article, .min_period = 8, .max_period = 24, .averaging_length = 3, .enhance_resolution = false },
            &expected_hm,
            &expected_min,
            &expected_max,
            &expected_dc,
        );
    }
    {
        const expected_hm = testdata.expectedV0_L8_H24_A3_E1_HEATMAP();
        const expected_min = testdata.expectedV0_L8_H24_A3_E1_MIN();
        const expected_max = testdata.expectedV0_L8_H24_A3_E1_MAX();
        const expected_dc = testdata.expectedV0_L8_H24_A3_E1_DC();
        try verifyReferenceCase(
            17,
            .{ .source = .article, .min_period = 8, .max_period = 24, .averaging_length = 3, .enhance_resolution = true },
            &expected_hm,
            &expected_min,
            &expected_max,
            &expected_dc,
        );
    }
    {
        const expected_hm = testdata.expectedV0_L8_H24_A5_E0_HEATMAP();
        const expected_min = testdata.expectedV0_L8_H24_A5_E0_MIN();
        const expected_max = testdata.expectedV0_L8_H24_A5_E0_MAX();
        const expected_dc = testdata.expectedV0_L8_H24_A5_E0_DC();
        try verifyReferenceCase(
            17,
            .{ .source = .article, .min_period = 8, .max_period = 24, .averaging_length = 5, .enhance_resolution = false },
            &expected_hm,
            &expected_min,
            &expected_max,
            &expected_dc,
        );
    }
    {
        const expected_hm = testdata.expectedV0_L8_H24_A5_E1_HEATMAP();
        const expected_min = testdata.expectedV0_L8_H24_A5_E1_MIN();
        const expected_max = testdata.expectedV0_L8_H24_A5_E1_MAX();
        const expected_dc = testdata.expectedV0_L8_H24_A5_E1_DC();
        try verifyReferenceCase(
            17,
            .{ .source = .article, .min_period = 8, .max_period = 24, .averaging_length = 5, .enhance_resolution = true },
            &expected_hm,
            &expected_min,
            &expected_max,
            &expected_dc,
        );
    }
    {
        const expected_hm = testdata.expectedV0_L8_H24_A10_E0_HEATMAP();
        const expected_min = testdata.expectedV0_L8_H24_A10_E0_MIN();
        const expected_max = testdata.expectedV0_L8_H24_A10_E0_MAX();
        const expected_dc = testdata.expectedV0_L8_H24_A10_E0_DC();
        try verifyReferenceCase(
            17,
            .{ .source = .article, .min_period = 8, .max_period = 24, .averaging_length = 10, .enhance_resolution = false },
            &expected_hm,
            &expected_min,
            &expected_max,
            &expected_dc,
        );
    }
    {
        const expected_hm = testdata.expectedV0_L8_H24_A10_E1_HEATMAP();
        const expected_min = testdata.expectedV0_L8_H24_A10_E1_MIN();
        const expected_max = testdata.expectedV0_L8_H24_A10_E1_MAX();
        const expected_dc = testdata.expectedV0_L8_H24_A10_E1_DC();
        try verifyReferenceCase(
            17,
            .{ .source = .article, .min_period = 8, .max_period = 24, .averaging_length = 10, .enhance_resolution = true },
            &expected_hm,
            &expected_min,
            &expected_max,
            &expected_dc,
        );
    }
    {
        const expected_hm = testdata.expectedV0_L8_H48_A0_E0_HEATMAP();
        const expected_min = testdata.expectedV0_L8_H48_A0_E0_MIN();
        const expected_max = testdata.expectedV0_L8_H48_A0_E0_MAX();
        const expected_dc = testdata.expectedV0_L8_H48_A0_E0_DC();
        try verifyReferenceCase(
            41,
            .{ .source = .article, .min_period = 8, .max_period = 48, .averaging_length = 0, .enhance_resolution = false },
            &expected_hm,
            &expected_min,
            &expected_max,
            &expected_dc,
        );
    }
    {
        const expected_hm = testdata.expectedV0_L8_H48_A0_E1_HEATMAP();
        const expected_min = testdata.expectedV0_L8_H48_A0_E1_MIN();
        const expected_max = testdata.expectedV0_L8_H48_A0_E1_MAX();
        const expected_dc = testdata.expectedV0_L8_H48_A0_E1_DC();
        try verifyReferenceCase(
            41,
            .{ .source = .article, .min_period = 8, .max_period = 48, .averaging_length = 0, .enhance_resolution = true },
            &expected_hm,
            &expected_min,
            &expected_max,
            &expected_dc,
        );
    }
    {
        const expected_hm = testdata.expectedV0_L8_H48_A3_E0_HEATMAP();
        const expected_min = testdata.expectedV0_L8_H48_A3_E0_MIN();
        const expected_max = testdata.expectedV0_L8_H48_A3_E0_MAX();
        const expected_dc = testdata.expectedV0_L8_H48_A3_E0_DC();
        try verifyReferenceCase(
            41,
            .{ .source = .article, .min_period = 8, .max_period = 48, .averaging_length = 3, .enhance_resolution = false },
            &expected_hm,
            &expected_min,
            &expected_max,
            &expected_dc,
        );
    }
    {
        const expected_hm = testdata.expectedV0_L8_H48_A3_E1_HEATMAP();
        const expected_min = testdata.expectedV0_L8_H48_A3_E1_MIN();
        const expected_max = testdata.expectedV0_L8_H48_A3_E1_MAX();
        const expected_dc = testdata.expectedV0_L8_H48_A3_E1_DC();
        try verifyReferenceCase(
            41,
            .{ .source = .article, .min_period = 8, .max_period = 48, .averaging_length = 3, .enhance_resolution = true },
            &expected_hm,
            &expected_min,
            &expected_max,
            &expected_dc,
        );
    }
    {
        const expected_hm = testdata.expectedV0_L8_H48_A5_E0_HEATMAP();
        const expected_min = testdata.expectedV0_L8_H48_A5_E0_MIN();
        const expected_max = testdata.expectedV0_L8_H48_A5_E0_MAX();
        const expected_dc = testdata.expectedV0_L8_H48_A5_E0_DC();
        try verifyReferenceCase(
            41,
            .{ .source = .article, .min_period = 8, .max_period = 48, .averaging_length = 5, .enhance_resolution = false },
            &expected_hm,
            &expected_min,
            &expected_max,
            &expected_dc,
        );
    }
    {
        const expected_hm = testdata.expectedV0_L8_H48_A5_E1_HEATMAP();
        const expected_min = testdata.expectedV0_L8_H48_A5_E1_MIN();
        const expected_max = testdata.expectedV0_L8_H48_A5_E1_MAX();
        const expected_dc = testdata.expectedV0_L8_H48_A5_E1_DC();
        try verifyReferenceCase(
            41,
            .{ .source = .article, .min_period = 8, .max_period = 48, .averaging_length = 5, .enhance_resolution = true },
            &expected_hm,
            &expected_min,
            &expected_max,
            &expected_dc,
        );
    }
    {
        const expected_hm = testdata.expectedV0_L8_H48_A10_E0_HEATMAP();
        const expected_min = testdata.expectedV0_L8_H48_A10_E0_MIN();
        const expected_max = testdata.expectedV0_L8_H48_A10_E0_MAX();
        const expected_dc = testdata.expectedV0_L8_H48_A10_E0_DC();
        try verifyReferenceCase(
            41,
            .{ .source = .article, .min_period = 8, .max_period = 48, .averaging_length = 10, .enhance_resolution = false },
            &expected_hm,
            &expected_min,
            &expected_max,
            &expected_dc,
        );
    }
    {
        const expected_hm = testdata.expectedV0_L8_H48_A10_E1_HEATMAP();
        const expected_min = testdata.expectedV0_L8_H48_A10_E1_MIN();
        const expected_max = testdata.expectedV0_L8_H48_A10_E1_MAX();
        const expected_dc = testdata.expectedV0_L8_H48_A10_E1_DC();
        try verifyReferenceCase(
            41,
            .{ .source = .article, .min_period = 8, .max_period = 48, .averaging_length = 10, .enhance_resolution = true },
            &expected_hm,
            &expected_min,
            &expected_max,
            &expected_dc,
        );
    }
    {
        const expected_hm = testdata.expectedV0_L10_H48_A0_E0_HEATMAP();
        const expected_min = testdata.expectedV0_L10_H48_A0_E0_MIN();
        const expected_max = testdata.expectedV0_L10_H48_A0_E0_MAX();
        const expected_dc = testdata.expectedV0_L10_H48_A0_E0_DC();
        try verifyReferenceCase(
            39,
            .{ .source = .article, .min_period = 10, .max_period = 48, .averaging_length = 0, .enhance_resolution = false },
            &expected_hm,
            &expected_min,
            &expected_max,
            &expected_dc,
        );
    }
    {
        const expected_hm = testdata.expectedV0_L10_H48_A0_E1_HEATMAP();
        const expected_min = testdata.expectedV0_L10_H48_A0_E1_MIN();
        const expected_max = testdata.expectedV0_L10_H48_A0_E1_MAX();
        const expected_dc = testdata.expectedV0_L10_H48_A0_E1_DC();
        try verifyReferenceCase(
            39,
            .{ .source = .article, .min_period = 10, .max_period = 48, .averaging_length = 0, .enhance_resolution = true },
            &expected_hm,
            &expected_min,
            &expected_max,
            &expected_dc,
        );
    }
    {
        const expected_hm = testdata.expectedV0_L10_H48_A3_E0_HEATMAP();
        const expected_min = testdata.expectedV0_L10_H48_A3_E0_MIN();
        const expected_max = testdata.expectedV0_L10_H48_A3_E0_MAX();
        const expected_dc = testdata.expectedV0_L10_H48_A3_E0_DC();
        try verifyReferenceCase(
            39,
            .{ .source = .article, .min_period = 10, .max_period = 48, .averaging_length = 3, .enhance_resolution = false },
            &expected_hm,
            &expected_min,
            &expected_max,
            &expected_dc,
        );
    }
    {
        const expected_hm = testdata.expectedV0_L10_H48_A3_E1_HEATMAP();
        const expected_min = testdata.expectedV0_L10_H48_A3_E1_MIN();
        const expected_max = testdata.expectedV0_L10_H48_A3_E1_MAX();
        const expected_dc = testdata.expectedV0_L10_H48_A3_E1_DC();
        try verifyReferenceCase(
            39,
            .{ .source = .article, .min_period = 10, .max_period = 48, .averaging_length = 3, .enhance_resolution = true },
            &expected_hm,
            &expected_min,
            &expected_max,
            &expected_dc,
        );
    }
    {
        const expected_hm = testdata.expectedV0_L10_H48_A5_E0_HEATMAP();
        const expected_min = testdata.expectedV0_L10_H48_A5_E0_MIN();
        const expected_max = testdata.expectedV0_L10_H48_A5_E0_MAX();
        const expected_dc = testdata.expectedV0_L10_H48_A5_E0_DC();
        try verifyReferenceCase(
            39,
            .{ .source = .article, .min_period = 10, .max_period = 48, .averaging_length = 5, .enhance_resolution = false },
            &expected_hm,
            &expected_min,
            &expected_max,
            &expected_dc,
        );
    }
    {
        const expected_hm = testdata.expectedV0_L10_H48_A5_E1_HEATMAP();
        const expected_min = testdata.expectedV0_L10_H48_A5_E1_MIN();
        const expected_max = testdata.expectedV0_L10_H48_A5_E1_MAX();
        const expected_dc = testdata.expectedV0_L10_H48_A5_E1_DC();
        try verifyReferenceCase(
            39,
            .{ .source = .article, .min_period = 10, .max_period = 48, .averaging_length = 5, .enhance_resolution = true },
            &expected_hm,
            &expected_min,
            &expected_max,
            &expected_dc,
        );
    }
    {
        const expected_hm = testdata.expectedV0_L10_H48_A10_E0_HEATMAP();
        const expected_min = testdata.expectedV0_L10_H48_A10_E0_MIN();
        const expected_max = testdata.expectedV0_L10_H48_A10_E0_MAX();
        const expected_dc = testdata.expectedV0_L10_H48_A10_E0_DC();
        try verifyReferenceCase(
            39,
            .{ .source = .article, .min_period = 10, .max_period = 48, .averaging_length = 10, .enhance_resolution = false },
            &expected_hm,
            &expected_min,
            &expected_max,
            &expected_dc,
        );
    }
    {
        const expected_hm = testdata.expectedV0_L10_H48_A10_E1_HEATMAP();
        const expected_min = testdata.expectedV0_L10_H48_A10_E1_MIN();
        const expected_max = testdata.expectedV0_L10_H48_A10_E1_MAX();
        const expected_dc = testdata.expectedV0_L10_H48_A10_E1_DC();
        try verifyReferenceCase(
            39,
            .{ .source = .article, .min_period = 10, .max_period = 48, .averaging_length = 10, .enhance_resolution = true },
            &expected_hm,
            &expected_min,
            &expected_max,
            &expected_dc,
        );
    }
    {
        const expected_hm = testdata.expectedV0_L16_H64_A0_E0_HEATMAP();
        const expected_min = testdata.expectedV0_L16_H64_A0_E0_MIN();
        const expected_max = testdata.expectedV0_L16_H64_A0_E0_MAX();
        const expected_dc = testdata.expectedV0_L16_H64_A0_E0_DC();
        try verifyReferenceCase(
            49,
            .{ .source = .article, .min_period = 16, .max_period = 64, .averaging_length = 0, .enhance_resolution = false },
            &expected_hm,
            &expected_min,
            &expected_max,
            &expected_dc,
        );
    }
    {
        const expected_hm = testdata.expectedV0_L16_H64_A0_E1_HEATMAP();
        const expected_min = testdata.expectedV0_L16_H64_A0_E1_MIN();
        const expected_max = testdata.expectedV0_L16_H64_A0_E1_MAX();
        const expected_dc = testdata.expectedV0_L16_H64_A0_E1_DC();
        try verifyReferenceCase(
            49,
            .{ .source = .article, .min_period = 16, .max_period = 64, .averaging_length = 0, .enhance_resolution = true },
            &expected_hm,
            &expected_min,
            &expected_max,
            &expected_dc,
        );
    }
    {
        const expected_hm = testdata.expectedV0_L16_H64_A3_E0_HEATMAP();
        const expected_min = testdata.expectedV0_L16_H64_A3_E0_MIN();
        const expected_max = testdata.expectedV0_L16_H64_A3_E0_MAX();
        const expected_dc = testdata.expectedV0_L16_H64_A3_E0_DC();
        try verifyReferenceCase(
            49,
            .{ .source = .article, .min_period = 16, .max_period = 64, .averaging_length = 3, .enhance_resolution = false },
            &expected_hm,
            &expected_min,
            &expected_max,
            &expected_dc,
        );
    }
    {
        const expected_hm = testdata.expectedV0_L16_H64_A3_E1_HEATMAP();
        const expected_min = testdata.expectedV0_L16_H64_A3_E1_MIN();
        const expected_max = testdata.expectedV0_L16_H64_A3_E1_MAX();
        const expected_dc = testdata.expectedV0_L16_H64_A3_E1_DC();
        try verifyReferenceCase(
            49,
            .{ .source = .article, .min_period = 16, .max_period = 64, .averaging_length = 3, .enhance_resolution = true },
            &expected_hm,
            &expected_min,
            &expected_max,
            &expected_dc,
        );
    }
    {
        const expected_hm = testdata.expectedV0_L16_H64_A5_E0_HEATMAP();
        const expected_min = testdata.expectedV0_L16_H64_A5_E0_MIN();
        const expected_max = testdata.expectedV0_L16_H64_A5_E0_MAX();
        const expected_dc = testdata.expectedV0_L16_H64_A5_E0_DC();
        try verifyReferenceCase(
            49,
            .{ .source = .article, .min_period = 16, .max_period = 64, .averaging_length = 5, .enhance_resolution = false },
            &expected_hm,
            &expected_min,
            &expected_max,
            &expected_dc,
        );
    }
    {
        const expected_hm = testdata.expectedV0_L16_H64_A5_E1_HEATMAP();
        const expected_min = testdata.expectedV0_L16_H64_A5_E1_MIN();
        const expected_max = testdata.expectedV0_L16_H64_A5_E1_MAX();
        const expected_dc = testdata.expectedV0_L16_H64_A5_E1_DC();
        try verifyReferenceCase(
            49,
            .{ .source = .article, .min_period = 16, .max_period = 64, .averaging_length = 5, .enhance_resolution = true },
            &expected_hm,
            &expected_min,
            &expected_max,
            &expected_dc,
        );
    }
    {
        const expected_hm = testdata.expectedV0_L16_H64_A10_E0_HEATMAP();
        const expected_min = testdata.expectedV0_L16_H64_A10_E0_MIN();
        const expected_max = testdata.expectedV0_L16_H64_A10_E0_MAX();
        const expected_dc = testdata.expectedV0_L16_H64_A10_E0_DC();
        try verifyReferenceCase(
            49,
            .{ .source = .article, .min_period = 16, .max_period = 64, .averaging_length = 10, .enhance_resolution = false },
            &expected_hm,
            &expected_min,
            &expected_max,
            &expected_dc,
        );
    }
    {
        const expected_hm = testdata.expectedV0_L16_H64_A10_E1_HEATMAP();
        const expected_min = testdata.expectedV0_L16_H64_A10_E1_MIN();
        const expected_max = testdata.expectedV0_L16_H64_A10_E1_MAX();
        const expected_dc = testdata.expectedV0_L16_H64_A10_E1_DC();
        try verifyReferenceCase(
            49,
            .{ .source = .article, .min_period = 16, .max_period = 64, .averaging_length = 10, .enhance_resolution = true },
            &expected_hm,
            &expected_min,
            &expected_max,
            &expected_dc,
        );
    }
    {
        const expected_hm = testdata.expectedV1_L8_H24_A0_E0_HEATMAP();
        const expected_min = testdata.expectedV1_L8_H24_A0_E0_MIN();
        const expected_max = testdata.expectedV1_L8_H24_A0_E0_MAX();
        const expected_dc = testdata.expectedV1_L8_H24_A0_E0_DC();
        try verifyReferenceCase(
            17,
            .{ .source = .book, .min_period = 8, .max_period = 24, .averaging_length = 0, .enhance_resolution = false },
            &expected_hm,
            &expected_min,
            &expected_max,
            &expected_dc,
        );
    }
    {
        const expected_hm = testdata.expectedV1_L8_H24_A0_E1_HEATMAP();
        const expected_min = testdata.expectedV1_L8_H24_A0_E1_MIN();
        const expected_max = testdata.expectedV1_L8_H24_A0_E1_MAX();
        const expected_dc = testdata.expectedV1_L8_H24_A0_E1_DC();
        try verifyReferenceCase(
            17,
            .{ .source = .book, .min_period = 8, .max_period = 24, .averaging_length = 0, .enhance_resolution = true },
            &expected_hm,
            &expected_min,
            &expected_max,
            &expected_dc,
        );
    }
    {
        const expected_hm = testdata.expectedV1_L8_H24_A3_E0_HEATMAP();
        const expected_min = testdata.expectedV1_L8_H24_A3_E0_MIN();
        const expected_max = testdata.expectedV1_L8_H24_A3_E0_MAX();
        const expected_dc = testdata.expectedV1_L8_H24_A3_E0_DC();
        try verifyReferenceCase(
            17,
            .{ .source = .book, .min_period = 8, .max_period = 24, .averaging_length = 3, .enhance_resolution = false },
            &expected_hm,
            &expected_min,
            &expected_max,
            &expected_dc,
        );
    }
    {
        const expected_hm = testdata.expectedV1_L8_H24_A3_E1_HEATMAP();
        const expected_min = testdata.expectedV1_L8_H24_A3_E1_MIN();
        const expected_max = testdata.expectedV1_L8_H24_A3_E1_MAX();
        const expected_dc = testdata.expectedV1_L8_H24_A3_E1_DC();
        try verifyReferenceCase(
            17,
            .{ .source = .book, .min_period = 8, .max_period = 24, .averaging_length = 3, .enhance_resolution = true },
            &expected_hm,
            &expected_min,
            &expected_max,
            &expected_dc,
        );
    }
    {
        const expected_hm = testdata.expectedV1_L8_H24_A5_E0_HEATMAP();
        const expected_min = testdata.expectedV1_L8_H24_A5_E0_MIN();
        const expected_max = testdata.expectedV1_L8_H24_A5_E0_MAX();
        const expected_dc = testdata.expectedV1_L8_H24_A5_E0_DC();
        try verifyReferenceCase(
            17,
            .{ .source = .book, .min_period = 8, .max_period = 24, .averaging_length = 5, .enhance_resolution = false },
            &expected_hm,
            &expected_min,
            &expected_max,
            &expected_dc,
        );
    }
    {
        const expected_hm = testdata.expectedV1_L8_H24_A5_E1_HEATMAP();
        const expected_min = testdata.expectedV1_L8_H24_A5_E1_MIN();
        const expected_max = testdata.expectedV1_L8_H24_A5_E1_MAX();
        const expected_dc = testdata.expectedV1_L8_H24_A5_E1_DC();
        try verifyReferenceCase(
            17,
            .{ .source = .book, .min_period = 8, .max_period = 24, .averaging_length = 5, .enhance_resolution = true },
            &expected_hm,
            &expected_min,
            &expected_max,
            &expected_dc,
        );
    }
    {
        const expected_hm = testdata.expectedV1_L8_H24_A10_E0_HEATMAP();
        const expected_min = testdata.expectedV1_L8_H24_A10_E0_MIN();
        const expected_max = testdata.expectedV1_L8_H24_A10_E0_MAX();
        const expected_dc = testdata.expectedV1_L8_H24_A10_E0_DC();
        try verifyReferenceCase(
            17,
            .{ .source = .book, .min_period = 8, .max_period = 24, .averaging_length = 10, .enhance_resolution = false },
            &expected_hm,
            &expected_min,
            &expected_max,
            &expected_dc,
        );
    }
    {
        const expected_hm = testdata.expectedV1_L8_H24_A10_E1_HEATMAP();
        const expected_min = testdata.expectedV1_L8_H24_A10_E1_MIN();
        const expected_max = testdata.expectedV1_L8_H24_A10_E1_MAX();
        const expected_dc = testdata.expectedV1_L8_H24_A10_E1_DC();
        try verifyReferenceCase(
            17,
            .{ .source = .book, .min_period = 8, .max_period = 24, .averaging_length = 10, .enhance_resolution = true },
            &expected_hm,
            &expected_min,
            &expected_max,
            &expected_dc,
        );
    }
    {
        const expected_hm = testdata.expectedV1_L8_H48_A0_E0_HEATMAP();
        const expected_min = testdata.expectedV1_L8_H48_A0_E0_MIN();
        const expected_max = testdata.expectedV1_L8_H48_A0_E0_MAX();
        const expected_dc = testdata.expectedV1_L8_H48_A0_E0_DC();
        try verifyReferenceCase(
            41,
            .{ .source = .book, .min_period = 8, .max_period = 48, .averaging_length = 0, .enhance_resolution = false },
            &expected_hm,
            &expected_min,
            &expected_max,
            &expected_dc,
        );
    }
    {
        const expected_hm = testdata.expectedV1_L8_H48_A0_E1_HEATMAP();
        const expected_min = testdata.expectedV1_L8_H48_A0_E1_MIN();
        const expected_max = testdata.expectedV1_L8_H48_A0_E1_MAX();
        const expected_dc = testdata.expectedV1_L8_H48_A0_E1_DC();
        try verifyReferenceCase(
            41,
            .{ .source = .book, .min_period = 8, .max_period = 48, .averaging_length = 0, .enhance_resolution = true },
            &expected_hm,
            &expected_min,
            &expected_max,
            &expected_dc,
        );
    }
    {
        const expected_hm = testdata.expectedV1_L8_H48_A3_E0_HEATMAP();
        const expected_min = testdata.expectedV1_L8_H48_A3_E0_MIN();
        const expected_max = testdata.expectedV1_L8_H48_A3_E0_MAX();
        const expected_dc = testdata.expectedV1_L8_H48_A3_E0_DC();
        try verifyReferenceCase(
            41,
            .{ .source = .book, .min_period = 8, .max_period = 48, .averaging_length = 3, .enhance_resolution = false },
            &expected_hm,
            &expected_min,
            &expected_max,
            &expected_dc,
        );
    }
    {
        const expected_hm = testdata.expectedV1_L8_H48_A3_E1_HEATMAP();
        const expected_min = testdata.expectedV1_L8_H48_A3_E1_MIN();
        const expected_max = testdata.expectedV1_L8_H48_A3_E1_MAX();
        const expected_dc = testdata.expectedV1_L8_H48_A3_E1_DC();
        try verifyReferenceCase(
            41,
            .{ .source = .book, .min_period = 8, .max_period = 48, .averaging_length = 3, .enhance_resolution = true },
            &expected_hm,
            &expected_min,
            &expected_max,
            &expected_dc,
        );
    }
    {
        const expected_hm = testdata.expectedV1_L8_H48_A5_E0_HEATMAP();
        const expected_min = testdata.expectedV1_L8_H48_A5_E0_MIN();
        const expected_max = testdata.expectedV1_L8_H48_A5_E0_MAX();
        const expected_dc = testdata.expectedV1_L8_H48_A5_E0_DC();
        try verifyReferenceCase(
            41,
            .{ .source = .book, .min_period = 8, .max_period = 48, .averaging_length = 5, .enhance_resolution = false },
            &expected_hm,
            &expected_min,
            &expected_max,
            &expected_dc,
        );
    }
    {
        const expected_hm = testdata.expectedV1_L8_H48_A5_E1_HEATMAP();
        const expected_min = testdata.expectedV1_L8_H48_A5_E1_MIN();
        const expected_max = testdata.expectedV1_L8_H48_A5_E1_MAX();
        const expected_dc = testdata.expectedV1_L8_H48_A5_E1_DC();
        try verifyReferenceCase(
            41,
            .{ .source = .book, .min_period = 8, .max_period = 48, .averaging_length = 5, .enhance_resolution = true },
            &expected_hm,
            &expected_min,
            &expected_max,
            &expected_dc,
        );
    }
    {
        const expected_hm = testdata.expectedV1_L8_H48_A10_E0_HEATMAP();
        const expected_min = testdata.expectedV1_L8_H48_A10_E0_MIN();
        const expected_max = testdata.expectedV1_L8_H48_A10_E0_MAX();
        const expected_dc = testdata.expectedV1_L8_H48_A10_E0_DC();
        try verifyReferenceCase(
            41,
            .{ .source = .book, .min_period = 8, .max_period = 48, .averaging_length = 10, .enhance_resolution = false },
            &expected_hm,
            &expected_min,
            &expected_max,
            &expected_dc,
        );
    }
    {
        const expected_hm = testdata.expectedV1_L8_H48_A10_E1_HEATMAP();
        const expected_min = testdata.expectedV1_L8_H48_A10_E1_MIN();
        const expected_max = testdata.expectedV1_L8_H48_A10_E1_MAX();
        const expected_dc = testdata.expectedV1_L8_H48_A10_E1_DC();
        try verifyReferenceCase(
            41,
            .{ .source = .book, .min_period = 8, .max_period = 48, .averaging_length = 10, .enhance_resolution = true },
            &expected_hm,
            &expected_min,
            &expected_max,
            &expected_dc,
        );
    }
    {
        const expected_hm = testdata.expectedV1_L10_H48_A0_E0_HEATMAP();
        const expected_min = testdata.expectedV1_L10_H48_A0_E0_MIN();
        const expected_max = testdata.expectedV1_L10_H48_A0_E0_MAX();
        const expected_dc = testdata.expectedV1_L10_H48_A0_E0_DC();
        try verifyReferenceCase(
            39,
            .{ .source = .book, .min_period = 10, .max_period = 48, .averaging_length = 0, .enhance_resolution = false },
            &expected_hm,
            &expected_min,
            &expected_max,
            &expected_dc,
        );
    }
    {
        const expected_hm = testdata.expectedV1_L10_H48_A0_E1_HEATMAP();
        const expected_min = testdata.expectedV1_L10_H48_A0_E1_MIN();
        const expected_max = testdata.expectedV1_L10_H48_A0_E1_MAX();
        const expected_dc = testdata.expectedV1_L10_H48_A0_E1_DC();
        try verifyReferenceCase(
            39,
            .{ .source = .book, .min_period = 10, .max_period = 48, .averaging_length = 0, .enhance_resolution = true },
            &expected_hm,
            &expected_min,
            &expected_max,
            &expected_dc,
        );
    }
    {
        const expected_hm = testdata.expectedV1_L10_H48_A3_E0_HEATMAP();
        const expected_min = testdata.expectedV1_L10_H48_A3_E0_MIN();
        const expected_max = testdata.expectedV1_L10_H48_A3_E0_MAX();
        const expected_dc = testdata.expectedV1_L10_H48_A3_E0_DC();
        try verifyReferenceCase(
            39,
            .{ .source = .book, .min_period = 10, .max_period = 48, .averaging_length = 3, .enhance_resolution = false },
            &expected_hm,
            &expected_min,
            &expected_max,
            &expected_dc,
        );
    }
    {
        const expected_hm = testdata.expectedV1_L10_H48_A3_E1_HEATMAP();
        const expected_min = testdata.expectedV1_L10_H48_A3_E1_MIN();
        const expected_max = testdata.expectedV1_L10_H48_A3_E1_MAX();
        const expected_dc = testdata.expectedV1_L10_H48_A3_E1_DC();
        try verifyReferenceCase(
            39,
            .{ .source = .book, .min_period = 10, .max_period = 48, .averaging_length = 3, .enhance_resolution = true },
            &expected_hm,
            &expected_min,
            &expected_max,
            &expected_dc,
        );
    }
    {
        const expected_hm = testdata.expectedV1_L10_H48_A5_E0_HEATMAP();
        const expected_min = testdata.expectedV1_L10_H48_A5_E0_MIN();
        const expected_max = testdata.expectedV1_L10_H48_A5_E0_MAX();
        const expected_dc = testdata.expectedV1_L10_H48_A5_E0_DC();
        try verifyReferenceCase(
            39,
            .{ .source = .book, .min_period = 10, .max_period = 48, .averaging_length = 5, .enhance_resolution = false },
            &expected_hm,
            &expected_min,
            &expected_max,
            &expected_dc,
        );
    }
    {
        const expected_hm = testdata.expectedV1_L10_H48_A5_E1_HEATMAP();
        const expected_min = testdata.expectedV1_L10_H48_A5_E1_MIN();
        const expected_max = testdata.expectedV1_L10_H48_A5_E1_MAX();
        const expected_dc = testdata.expectedV1_L10_H48_A5_E1_DC();
        try verifyReferenceCase(
            39,
            .{ .source = .book, .min_period = 10, .max_period = 48, .averaging_length = 5, .enhance_resolution = true },
            &expected_hm,
            &expected_min,
            &expected_max,
            &expected_dc,
        );
    }
    {
        const expected_hm = testdata.expectedV1_L10_H48_A10_E0_HEATMAP();
        const expected_min = testdata.expectedV1_L10_H48_A10_E0_MIN();
        const expected_max = testdata.expectedV1_L10_H48_A10_E0_MAX();
        const expected_dc = testdata.expectedV1_L10_H48_A10_E0_DC();
        try verifyReferenceCase(
            39,
            .{ .source = .book, .min_period = 10, .max_period = 48, .averaging_length = 10, .enhance_resolution = false },
            &expected_hm,
            &expected_min,
            &expected_max,
            &expected_dc,
        );
    }
    {
        const expected_hm = testdata.expectedV1_L10_H48_A10_E1_HEATMAP();
        const expected_min = testdata.expectedV1_L10_H48_A10_E1_MIN();
        const expected_max = testdata.expectedV1_L10_H48_A10_E1_MAX();
        const expected_dc = testdata.expectedV1_L10_H48_A10_E1_DC();
        try verifyReferenceCase(
            39,
            .{ .source = .book, .min_period = 10, .max_period = 48, .averaging_length = 10, .enhance_resolution = true },
            &expected_hm,
            &expected_min,
            &expected_max,
            &expected_dc,
        );
    }
    {
        const expected_hm = testdata.expectedV1_L16_H64_A0_E0_HEATMAP();
        const expected_min = testdata.expectedV1_L16_H64_A0_E0_MIN();
        const expected_max = testdata.expectedV1_L16_H64_A0_E0_MAX();
        const expected_dc = testdata.expectedV1_L16_H64_A0_E0_DC();
        try verifyReferenceCase(
            49,
            .{ .source = .book, .min_period = 16, .max_period = 64, .averaging_length = 0, .enhance_resolution = false },
            &expected_hm,
            &expected_min,
            &expected_max,
            &expected_dc,
        );
    }
    {
        const expected_hm = testdata.expectedV1_L16_H64_A0_E1_HEATMAP();
        const expected_min = testdata.expectedV1_L16_H64_A0_E1_MIN();
        const expected_max = testdata.expectedV1_L16_H64_A0_E1_MAX();
        const expected_dc = testdata.expectedV1_L16_H64_A0_E1_DC();
        try verifyReferenceCase(
            49,
            .{ .source = .book, .min_period = 16, .max_period = 64, .averaging_length = 0, .enhance_resolution = true },
            &expected_hm,
            &expected_min,
            &expected_max,
            &expected_dc,
        );
    }
    {
        const expected_hm = testdata.expectedV1_L16_H64_A3_E0_HEATMAP();
        const expected_min = testdata.expectedV1_L16_H64_A3_E0_MIN();
        const expected_max = testdata.expectedV1_L16_H64_A3_E0_MAX();
        const expected_dc = testdata.expectedV1_L16_H64_A3_E0_DC();
        try verifyReferenceCase(
            49,
            .{ .source = .book, .min_period = 16, .max_period = 64, .averaging_length = 3, .enhance_resolution = false },
            &expected_hm,
            &expected_min,
            &expected_max,
            &expected_dc,
        );
    }
    {
        const expected_hm = testdata.expectedV1_L16_H64_A3_E1_HEATMAP();
        const expected_min = testdata.expectedV1_L16_H64_A3_E1_MIN();
        const expected_max = testdata.expectedV1_L16_H64_A3_E1_MAX();
        const expected_dc = testdata.expectedV1_L16_H64_A3_E1_DC();
        try verifyReferenceCase(
            49,
            .{ .source = .book, .min_period = 16, .max_period = 64, .averaging_length = 3, .enhance_resolution = true },
            &expected_hm,
            &expected_min,
            &expected_max,
            &expected_dc,
        );
    }
    {
        const expected_hm = testdata.expectedV1_L16_H64_A5_E0_HEATMAP();
        const expected_min = testdata.expectedV1_L16_H64_A5_E0_MIN();
        const expected_max = testdata.expectedV1_L16_H64_A5_E0_MAX();
        const expected_dc = testdata.expectedV1_L16_H64_A5_E0_DC();
        try verifyReferenceCase(
            49,
            .{ .source = .book, .min_period = 16, .max_period = 64, .averaging_length = 5, .enhance_resolution = false },
            &expected_hm,
            &expected_min,
            &expected_max,
            &expected_dc,
        );
    }
    {
        const expected_hm = testdata.expectedV1_L16_H64_A5_E1_HEATMAP();
        const expected_min = testdata.expectedV1_L16_H64_A5_E1_MIN();
        const expected_max = testdata.expectedV1_L16_H64_A5_E1_MAX();
        const expected_dc = testdata.expectedV1_L16_H64_A5_E1_DC();
        try verifyReferenceCase(
            49,
            .{ .source = .book, .min_period = 16, .max_period = 64, .averaging_length = 5, .enhance_resolution = true },
            &expected_hm,
            &expected_min,
            &expected_max,
            &expected_dc,
        );
    }
    {
        const expected_hm = testdata.expectedV1_L16_H64_A10_E0_HEATMAP();
        const expected_min = testdata.expectedV1_L16_H64_A10_E0_MIN();
        const expected_max = testdata.expectedV1_L16_H64_A10_E0_MAX();
        const expected_dc = testdata.expectedV1_L16_H64_A10_E0_DC();
        try verifyReferenceCase(
            49,
            .{ .source = .book, .min_period = 16, .max_period = 64, .averaging_length = 10, .enhance_resolution = false },
            &expected_hm,
            &expected_min,
            &expected_max,
            &expected_dc,
        );
    }
    {
        const expected_hm = testdata.expectedV1_L16_H64_A10_E1_HEATMAP();
        const expected_min = testdata.expectedV1_L16_H64_A10_E1_MIN();
        const expected_max = testdata.expectedV1_L16_H64_A10_E1_MAX();
        const expected_dc = testdata.expectedV1_L16_H64_A10_E1_DC();
        try verifyReferenceCase(
            49,
            .{ .source = .book, .min_period = 16, .max_period = 64, .averaging_length = 10, .enhance_resolution = true },
            &expected_hm,
            &expected_min,
            &expected_max,
            &expected_dc,
        );
    }
}

test "AutoCorrelationPeriodogram enhancement order, book hold, zero power, sine peak, NaN, overflow, and reset" {
    const prices = testdata.testInput();

    var book_plain = try AutoCorrelationPeriodogram.init(testing.allocator, .{ .source = .book, .enhance_resolution = false });
    defer book_plain.deinit();
    var book_enh = try AutoCorrelationPeriodogram.init(testing.allocator, .{ .source = .book, .enhance_resolution = true });
    defer book_enh.deinit();
    var art_plain = try AutoCorrelationPeriodogram.init(testing.allocator, .{ .source = .article, .enhance_resolution = false });
    defer art_plain.deinit();
    var art_enh = try AutoCorrelationPeriodogram.init(testing.allocator, .{ .source = .article, .enhance_resolution = true });
    defer art_enh.deinit();

    var previous_book_dc: f64 = math.nan(f64);
    var hold_count: usize = 0;
    var last_art_plain_dc: f64 = 0.0;
    var last_art_enh_dc: f64 = 0.0;

    for (prices, 0..) |price, bar_index| {
        const bp = book_plain.update(price, @intCast(bar_index));
        const be = book_enh.update(price, @intCast(bar_index));
        const ap = art_plain.update(price, @intCast(bar_index));
        const ae = art_enh.update(price, @intCast(bar_index));
        if (bar_index >= 50) {
            try expectCloseTol(bp.dominant_cycle, be.dominant_cycle, exact_tolerance);
            const bp_values = bp.heatmap.valuesSlice();
            const be_values = be.heatmap.valuesSlice();
            for (0..bp_values.len) |bin_index| {
                try expectCloseTol(be_values[bin_index], bp_values[bin_index] * bp_values[bin_index], exact_tolerance);
            }
            const ap_values = ap.heatmap.valuesSlice();
            const ae_values = ae.heatmap.valuesSlice();
            try expectCloseTol(ap.heatmap.value_max, 1.0, exact_tolerance);
            for (0..ap_values.len) |bin_index| {
                try expectCloseTol(
                    ae_values[bin_index],
                    ap_values[bin_index] * ap_values[bin_index] * ap_values[bin_index],
                    exact_tolerance,
                );
            }
            if (bar_index >= 51 and bp.heatmap.value_max < 0.5) {
                hold_count += 1;
                try expectCloseTol(bp.dominant_cycle, previous_book_dc, exact_tolerance);
            }
        }
        previous_book_dc = bp.dominant_cycle;
        last_art_plain_dc = ap.dominant_cycle;
        last_art_enh_dc = ae.dominant_cycle;
    }
    try testing.expect(hold_count > 0);
    try testing.expect(@abs(last_art_plain_dc - last_art_enh_dc) > 1e-9);

    var zero_ind = try AutoCorrelationPeriodogram.initDefault(testing.allocator);
    defer zero_ind.deinit();
    var zero_out: PeriodogramOutput = undefined;
    for (0..100) |step_index| {
        zero_out = zero_ind.update(0.0, @intCast(step_index));
    }
    try testing.expect(zero_ind.isPrimed());
    try testing.expect(math.isNan(zero_out.dominant_cycle));
    for (zero_out.heatmap.valuesSlice()) |cell| {
        try testing.expectEqual(@as(f64, 0.0), cell);
    }
    zero_ind.correlations[12] = 0.75;
    _ = zero_ind.update(0.0, 100);
    try testing.expectEqual(@as(f64, 0.75), zero_ind.state().correlations[12]);

    var sine_ind = try AutoCorrelationPeriodogram.initDefault(testing.allocator);
    defer sine_ind.deinit();
    var sine_out: PeriodogramOutput = undefined;
    for (0..1000) |step_index| {
        const sample = 100.0 + @sin(two_pi * @as(f64, @floatFromInt(step_index)) / 20.0);
        sine_out = sine_ind.update(sample, @intCast(step_index));
    }
    var best_bin: usize = 0;
    const sine_values = sine_out.heatmap.valuesSlice();
    for (sine_values, 0..) |cell, bin_index| {
        if (cell > sine_values[best_bin]) best_bin = bin_index;
    }
    try expectCloseTol(sine_out.heatmap.parameter_first + @as(f64, @floatFromInt(best_bin)), 20.0, exact_tolerance);
    try testing.expect(sine_out.dominant_cycle > 18.0 and sine_out.dominant_cycle < 22.0);

    const saved_state = art_plain.state();
    const saved_hp = saved_state.high_pass;
    const saved_filt = saved_state.filtered;
    const saved_count = art_plain.samplesSeen();
    const saved_dc = art_plain.output().dominant_cycle;

    const nan_out = try art_plain.tryUpdate(math.nan(f64), 999);
    try testing.expect(nan_out.heatmap.isEmpty());
    try testing.expect(math.isNan(nan_out.dominant_cycle));
    try testing.expectEqual(saved_count, art_plain.samplesSeen());
    try expectCloseTol(art_plain.output().dominant_cycle, saved_dc, 0.0);

    try testing.expectError(error.InvalidSample, art_plain.tryUpdate(math.inf(f64), 999));
    try testing.expectError(error.InvalidSample, art_plain.tryUpdate(-math.inf(f64), 999));
    try testing.expectError(error.NumericOverflow, art_plain.tryUpdate(1e308, 999));
    try testing.expectEqual(saved_count, art_plain.samplesSeen());
    try expectCloseTol(art_plain.state().high_pass, saved_hp, 0.0);
    try expectCloseTol(art_plain.state().filtered, saved_filt, 0.0);

    art_plain.reset();
    try testing.expectEqual(@as(usize, 0), art_plain.samplesSeen());
    try testing.expect(!art_plain.isPrimed());
    var replay_dc: f64 = 0.0;
    for (prices, 0..) |price, bar_index| {
        replay_dc = art_plain.update(price, @intCast(bar_index)).dominant_cycle;
    }
    try expectCloseTol(replay_dc, saved_dc, 0.0);
}

test "AutoCorrelationPeriodogram metadata, mnemonics, entities, and validation" {
    var default_ind = try AutoCorrelationPeriodogram.initDefault(testing.allocator);
    defer default_ind.deinit();

    var md: Metadata = undefined;
    default_ind.getMetadata(&md);
    try testing.expectEqual(Identifier.auto_correlation_periodogram, md.identifier);
    try testing.expectEqualStrings("acp(8, 48, 3)", md.mnemonic);
    try testing.expectEqual(@as(usize, 2), md.outputs_len);
    try testing.expectEqual(shape_mod.Shape.heatmap, md.outputs_buf[0].shape);
    try testing.expectEqualStrings("acp(8, 48, 3)", md.outputs_buf[0].mnemonic);
    try testing.expectEqual(shape_mod.Shape.scalar, md.outputs_buf[1].shape);
    try testing.expectEqualStrings("acp-dc(8, 48, 3)", md.outputs_buf[1].mnemonic);

    var book_enh = try AutoCorrelationPeriodogram.init(
        testing.allocator,
        .{ .source = .book, .enhance_resolution = true },
    );
    defer book_enh.deinit();
    try testing.expectEqual(@as(usize, 10), book_enh.minPeriod());
    var md_book: Metadata = undefined;
    book_enh.getMetadata(&md_book);
    try testing.expectEqualStrings("acp(10, 48, 3, book, enh)", md_book.mnemonic);
    try testing.expectEqualStrings("acp-dc(10, 48, 3, book)", md_book.outputs_buf[1].mnemonic);

    var art_enh = try AutoCorrelationPeriodogram.init(
        testing.allocator,
        .{ .source = .article, .enhance_resolution = true, .bar_component = .median },
    );
    defer art_enh.deinit();
    var md_art: Metadata = undefined;
    art_enh.getMetadata(&md_art);
    try testing.expectEqualStrings("acp(8, 48, 3, enh, hl/2)", md_art.mnemonic);
    try testing.expectEqualStrings("acp-dc(8, 48, 3, enh, hl/2)", md_art.outputs_buf[1].mnemonic);

    const MnemonicCase = struct {
        params: AutoCorrelationPeriodogramParams,
        expected: []const u8,
        expected_dc: []const u8,
    };
    const mnemonic_cases = [_]MnemonicCase{
        .{ .params = .{}, .expected = "acp(8, 48, 3)", .expected_dc = "acp-dc(8, 48, 3)" },
        .{ .params = .{ .bar_component = .open }, .expected = "acp(8, 48, 3, o)", .expected_dc = "acp-dc(8, 48, 3, o)" },
        .{ .params = .{ .quote_component = .bid }, .expected = "acp(8, 48, 3, b)", .expected_dc = "acp-dc(8, 48, 3, b)" },
        .{ .params = .{ .trade_component = .volume }, .expected = "acp(8, 48, 3, v)", .expected_dc = "acp-dc(8, 48, 3, v)" },
        .{ .params = .{ .bar_component = .high, .quote_component = .ask }, .expected = "acp(8, 48, 3, h, a)", .expected_dc = "acp-dc(8, 48, 3, h, a)" },
        .{ .params = .{ .bar_component = .low, .trade_component = .volume }, .expected = "acp(8, 48, 3, l, v)", .expected_dc = "acp-dc(8, 48, 3, l, v)" },
        .{ .params = .{ .quote_component = .spread_bp, .trade_component = .volume }, .expected = "acp(8, 48, 3, spread bp, v)", .expected_dc = "acp-dc(8, 48, 3, spread bp, v)" },
    };
    var mn_storage: [64]u8 = undefined;
    var mn_dc_storage: [64]u8 = undefined;
    for (mnemonic_cases) |case_item| {
        try testing.expectEqualStrings(case_item.expected, try mnemonic(&mn_storage, case_item.params));
        try testing.expectEqualStrings(case_item.expected_dc, try dominantCycleMnemonic(&mn_dc_storage, case_item.params));
    }

    const prices = testdata.testInput();
    const highs = testdata.testHigh();
    const lows = testdata.testLow();
    const opens = testdata.testOpen();

    var bar_ind = try AutoCorrelationPeriodogram.initDefault(testing.allocator);
    defer bar_ind.deinit();
    var scalar_ind = try AutoCorrelationPeriodogram.initDefault(testing.allocator);
    defer scalar_ind.deinit();
    var quote_ind = try AutoCorrelationPeriodogram.initDefault(testing.allocator);
    defer quote_ind.deinit();
    var trade_ind = try AutoCorrelationPeriodogram.initDefault(testing.allocator);
    defer trade_ind.deinit();

    for (0..60) |step_index| {
        const price = prices[step_index];
        const bar_item = Bar{
            .time = @intCast(step_index),
            .open = opens[step_index],
            .high = highs[step_index],
            .low = lows[step_index],
            .close = price,
            .volume = 1000.0,
        };
        const scalar_item = Scalar{ .time = @intCast(step_index), .value = price };
        const quote_item = Quote{
            .time = @intCast(step_index),
            .bid_price = price - 0.5,
            .ask_price = price + 0.5,
            .bid_size = 10.0,
            .ask_size = 10.0,
        };
        const trade_item = Trade{ .time = @intCast(step_index), .price = price, .volume = 10.0 };

        const out_bar = bar_ind.updateBar(&bar_item);
        const out_scalar = scalar_ind.updateScalar(&scalar_item);
        const out_quote = quote_ind.updateQuote(&quote_item);
        const out_trade = trade_ind.updateTrade(&trade_item);

        try testing.expectEqual(@as(usize, 2), out_bar.len);
        try expectCloseTol(out_bar.values[1].scalar.value, out_scalar.values[1].scalar.value, exact_tolerance);
        try expectCloseTol(out_bar.values[1].scalar.value, out_quote.values[1].scalar.value, exact_tolerance);
        try expectCloseTol(out_bar.values[1].scalar.value, out_trade.values[1].scalar.value, exact_tolerance);
    }

    try testing.expectError(error.InvalidMinPeriod, AutoCorrelationPeriodogram.init(testing.allocator, .{ .min_period = 0 }));
    try testing.expectError(error.InvalidMinPeriod, AutoCorrelationPeriodogram.init(testing.allocator, .{ .min_period = 2 }));
    try testing.expectError(error.InvalidMaxPeriod, AutoCorrelationPeriodogram.init(testing.allocator, .{ .min_period = 3, .max_period = 4 }));
    try testing.expectError(error.InvalidMaxPeriod, AutoCorrelationPeriodogram.init(testing.allocator, .{ .min_period = 8, .max_period = 8 }));
    try testing.expectError(error.InvalidMaxPeriod, AutoCorrelationPeriodogram.init(testing.allocator, .{ .source = .book, .max_period = 10 }));
    try testing.expectError(error.InvalidAveragingLength, AutoCorrelationPeriodogram.init(testing.allocator, .{ .averaging_length = 1 }));
    try testing.expectError(error.InvalidAveragingLength, AutoCorrelationPeriodogram.init(testing.allocator, .{ .averaging_length = -1 }));
    try testing.expectError(error.SpectrumRangeTooLarge, AutoCorrelationPeriodogram.init(testing.allocator, .{ .min_period = 8, .max_period = 300 }));
}
