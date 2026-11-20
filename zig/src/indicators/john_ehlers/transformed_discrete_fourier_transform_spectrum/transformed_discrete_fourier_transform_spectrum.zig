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
const descriptor_mod = @import("../../core/descriptor.zig");
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
const log_ten: f64 = math.ln10;
const db_cap: f64 = 20.0;
const centroid_db_limit: f64 = 3.0;

/// Output indices for Ehlers' Transformed Discrete Fourier Transform Spectrum indicator.
pub const TransformedDiscreteFourierTransformSpectrumOutput = enum(u8) {
    value = 1,
    dominant_cycle = 2,
};

/// Parameters to create a `TransformedDiscreteFourierTransformSpectrum`.
///
/// Defaults correspond to Ehlers' January 2007 article settings.
pub const TransformedDiscreteFourierTransformSpectrumParams = struct {
    min_period: i32 = 8,
    max_period: i32 = 50,
    window_length: i32 = 50,
    highpass_period: f64 = 40.0,
    bar_component: ?bar_component.BarComponent = null,
    quote_component: ?quote_component.QuoteComponent = null,
    trade_component: ?trade_component.TradeComponent = null,
};

pub const Params = TransformedDiscreteFourierTransformSpectrumParams;

pub fn defaultParams() TransformedDiscreteFourierTransformSpectrumParams {
    return .{};
}

pub const Error = error{
    InvalidMinPeriod,
    InvalidMaxPeriod,
    SpectrumRangeTooLarge,
    InvalidWindowLength,
    InvalidHighpassPeriod,
    UnstableHighpassCoefficient,
    InvalidSample,
    NumericOverflow,
    MnemonicTooLong,
    OutOfMemory,
};

fn normalizeZero(number: f64) f64 {
    return if (number == 0.0) 0.0 else number;
}

fn formatMnemonicInternal(
    buffer: []u8,
    prefix: []const u8,
    params: TransformedDiscreteFourierTransformSpectrumParams,
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

    return std.fmt.bufPrint(buffer, "{s}({d}, {d}, {d}, {d}{s})", .{
        prefix,
        params.min_period,
        params.max_period,
        params.window_length,
        normalizeZero(params.highpass_period),
        comp_suffix,
    }) catch return error.MnemonicTooLong;
}

/// Builds the primary heatmap mnemonic into `buffer`.
pub fn mnemonic(buffer: []u8, params: TransformedDiscreteFourierTransformSpectrumParams) Error![]const u8 {
    return formatMnemonicInternal(buffer, "tdfts", params);
}

/// Builds the dominant-cycle mnemonic into `buffer`.
pub fn dominantCycleMnemonic(buffer: []u8, params: TransformedDiscreteFourierTransformSpectrumParams) Error![]const u8 {
    return formatMnemonicInternal(buffer, "tdfts-dc", params);
}

/// Transformed dB spectrum heatmap column and dominant cycle period returned per observation.
pub const SpectrumOutput = struct {
    heatmap: Heatmap,
    dominant_cycle: f64,
};

/// Snapshot of committed internal filter, Fourier, dB, and centroid state.
pub const SpectrumState = struct {
    high_pass: f64,
    cleaned_data: f64,
    maximum_power: f64,
    dominant_cycle: f64,
    cosine_parts_storage: [heatmap_mod.max_heatmap_values]f64 = [_]f64{0.0} ** heatmap_mod.max_heatmap_values,
    sine_parts_storage: [heatmap_mod.max_heatmap_values]f64 = [_]f64{0.0} ** heatmap_mod.max_heatmap_values,
    powers_storage: [heatmap_mod.max_heatmap_values]f64 = [_]f64{0.0} ** heatmap_mod.max_heatmap_values,
    decibels_storage: [heatmap_mod.max_heatmap_values]f64 = [_]f64{0.0} ** heatmap_mod.max_heatmap_values,
    bins_len: usize,

    pub fn decibelsSlice(self: *const SpectrumState) []const f64 {
        return self.decibels_storage[0..self.bins_len];
    }

    pub fn eql(self: *const SpectrumState, other: *const SpectrumState) bool {
        if (self.bins_len != other.bins_len) return false;
        if (self.high_pass != other.high_pass) return false;
        if (self.cleaned_data != other.cleaned_data) return false;
        if (self.maximum_power != other.maximum_power) return false;
        if (self.dominant_cycle != other.dominant_cycle) return false;
        if (!std.mem.eql(f64, self.cosine_parts_storage[0..self.bins_len], other.cosine_parts_storage[0..self.bins_len])) return false;
        if (!std.mem.eql(f64, self.sine_parts_storage[0..self.bins_len], other.sine_parts_storage[0..self.bins_len])) return false;
        if (!std.mem.eql(f64, self.powers_storage[0..self.bins_len], other.powers_storage[0..self.bins_len])) return false;
        return std.mem.eql(f64, self.decibelsSlice(), other.decibelsSlice());
    }
};

/// Validated numeric parameters resolved at construction time.
pub const ResolvedParams = struct {
    min_period: usize,
    max_period: usize,
    window_length: usize,
    highpass_period: f64,
};

/// Ehlers' January 2007 Transformed Discrete Fourier Transform Spectrum indicator.
pub const TransformedDiscreteFourierTransformSpectrum = struct {
    allocator: std.mem.Allocator,
    resolved: ResolvedParams,
    bins_len: usize,
    warmup_count: usize,
    highpass_alpha: f64,
    cosines: []f64,
    sines: []f64,
    window: []f64,
    next_window: []f64,
    previous_price: f64 = 0.0,
    hp_history: [6]f64 = [_]f64{0.0} ** 6,
    cleaned: f64 = 0.0,
    cosine_parts: [heatmap_mod.max_heatmap_values]f64 = [_]f64{0.0} ** heatmap_mod.max_heatmap_values,
    sine_parts: [heatmap_mod.max_heatmap_values]f64 = [_]f64{0.0} ** heatmap_mod.max_heatmap_values,
    powers: [heatmap_mod.max_heatmap_values]f64 = [_]f64{0.0} ** heatmap_mod.max_heatmap_values,
    decibels: [heatmap_mod.max_heatmap_values]f64 = [_]f64{0.0} ** heatmap_mod.max_heatmap_values,
    sample_count: usize = 0,
    maximum: f64 = 0.0,
    dominant: f64 = 0.0,
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
        params: TransformedDiscreteFourierTransformSpectrumParams,
    ) Error!TransformedDiscreteFourierTransformSpectrum {
        if (params.min_period < 8) {
            return error.InvalidMinPeriod;
        }
        if (params.max_period < params.min_period) {
            return error.InvalidMaxPeriod;
        }
        const span = @as(i64, params.max_period) - @as(i64, params.min_period) + 1;
        if (span > heatmap_mod.max_heatmap_values) {
            return error.SpectrumRangeTooLarge;
        }
        if (params.window_length < 1) {
            return error.InvalidWindowLength;
        }
        if (!math.isFinite(params.highpass_period) or params.highpass_period <= 4.0) {
            return error.InvalidHighpassPeriod;
        }

        const omega = two_pi / params.highpass_period;
        const cosine_omega = @cos(omega);
        if (cosine_omega == 0.0) {
            return error.UnstableHighpassCoefficient;
        }
        const highpass_alpha = (1.0 - @sin(omega)) / cosine_omega;
        if (!math.isFinite(highpass_alpha)) {
            return error.UnstableHighpassCoefficient;
        }

        const min_period: usize = @intCast(params.min_period);
        const max_period: usize = @intCast(params.max_period);
        const window_length: usize = @intCast(params.window_length);
        const bins_len: usize = @intCast(span);
        const warmup_count: usize = @max(@as(usize, 5), window_length - 1);

        const table_size = math.mul(usize, bins_len, window_length) catch return error.OutOfMemory;
        const cosines = allocator.alloc(f64, table_size) catch return error.OutOfMemory;
        errdefer allocator.free(cosines);
        const sines = allocator.alloc(f64, table_size) catch return error.OutOfMemory;
        errdefer allocator.free(sines);
        const window = allocator.alloc(f64, window_length) catch return error.OutOfMemory;
        errdefer allocator.free(window);
        const next_window = allocator.alloc(f64, window_length) catch return error.OutOfMemory;
        errdefer allocator.free(next_window);

        @memset(window, 0.0);
        @memset(next_window, 0.0);

        for (0..bins_len) |bin_index| {
            const period_f: f64 = @floatFromInt(min_period + bin_index);
            const row_offset = bin_index * window_length;
            for (0..window_length) |lag| {
                const lag_f: f64 = @floatFromInt(lag);
                const angle = two_pi * lag_f / period_f;
                cosines[row_offset + lag] = @cos(angle);
                sines[row_offset + lag] = @sin(angle);
            }
        }

        const bar_comp = params.bar_component orelse bar_component.default_bar_component;
        const quote_comp = params.quote_component orelse quote_component.default_quote_component;
        const trade_comp = params.trade_component orelse trade_component.default_trade_component;

        var mnemonic_storage: [output_metadata_mod.max_mnemonic_len]u8 = undefined;
        const mn = try mnemonic(&mnemonic_storage, params);

        var description_storage: [output_metadata_mod.max_description_len]u8 = undefined;
        const desc = std.fmt.bufPrint(&description_storage, "Transformed discrete Fourier transform spectrum {s}", .{mn}) catch
            return error.MnemonicTooLong;

        var mnemonic_dc_storage: [output_metadata_mod.max_mnemonic_len]u8 = undefined;
        const mn_dc = try dominantCycleMnemonic(&mnemonic_dc_storage, params);

        var description_dc_storage: [output_metadata_mod.max_description_len]u8 = undefined;
        const desc_dc = std.fmt.bufPrint(
            &description_dc_storage,
            "Transformed discrete Fourier transform spectrum dominant cycle {s}",
            .{mn_dc},
        ) catch return error.MnemonicTooLong;

        const resolved = ResolvedParams{
            .min_period = min_period,
            .max_period = max_period,
            .window_length = window_length,
            .highpass_period = params.highpass_period,
        };

        return .{
            .allocator = allocator,
            .resolved = resolved,
            .bins_len = bins_len,
            .warmup_count = warmup_count,
            .highpass_alpha = highpass_alpha,
            .cosines = cosines,
            .sines = sines,
            .window = window,
            .next_window = next_window,
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

    pub fn initDefault(allocator: std.mem.Allocator) Error!TransformedDiscreteFourierTransformSpectrum {
        return init(allocator, defaultParams());
    }

    pub fn deinit(self: *TransformedDiscreteFourierTransformSpectrum) void {
        self.allocator.free(self.cosines);
        self.allocator.free(self.sines);
        self.allocator.free(self.window);
        self.allocator.free(self.next_window);
    }

    pub fn reset(self: *TransformedDiscreteFourierTransformSpectrum) void {
        self.sample_count = 0;
        self.previous_price = 0.0;
        self.hp_history = [_]f64{0.0} ** 6;
        self.cleaned = 0.0;
        @memset(self.window, 0.0);
        @memset(self.next_window, 0.0);
        @memset(&self.cosine_parts, 0.0);
        @memset(&self.sine_parts, 0.0);
        @memset(&self.powers, 0.0);
        @memset(&self.decibels, 0.0);
        self.maximum = 0.0;
        self.dominant = 0.0;
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

    pub fn fixSlices(self: *TransformedDiscreteFourierTransformSpectrum) void {
        _ = self;
    }

    pub fn mnemonicText(self: *const TransformedDiscreteFourierTransformSpectrum) []const u8 {
        return self.mnemonic_storage[0..self.mnemonic_len];
    }

    fn descriptionText(self: *const TransformedDiscreteFourierTransformSpectrum) []const u8 {
        return self.description_storage[0..self.description_len];
    }

    pub fn dominantCycleMnemonicText(self: *const TransformedDiscreteFourierTransformSpectrum) []const u8 {
        return self.mnemonic_dc_storage[0..self.mnemonic_dc_len];
    }

    fn dominantCycleDescriptionText(self: *const TransformedDiscreteFourierTransformSpectrum) []const u8 {
        return self.description_dc_storage[0..self.description_dc_len];
    }

    pub fn parameters(self: *const TransformedDiscreteFourierTransformSpectrum) ResolvedParams {
        return self.resolved;
    }

    pub fn minPeriod(self: *const TransformedDiscreteFourierTransformSpectrum) usize {
        return self.resolved.min_period;
    }

    pub fn maxPeriod(self: *const TransformedDiscreteFourierTransformSpectrum) usize {
        return self.resolved.max_period;
    }

    pub fn warmupBars(self: *const TransformedDiscreteFourierTransformSpectrum) usize {
        return self.warmup_count;
    }

    pub fn samplesSeen(self: *const TransformedDiscreteFourierTransformSpectrum) usize {
        return self.sample_count;
    }

    pub fn state(self: *const TransformedDiscreteFourierTransformSpectrum) SpectrumState {
        var out_state = SpectrumState{
            .high_pass = self.hp_history[0],
            .cleaned_data = self.cleaned,
            .maximum_power = self.maximum,
            .dominant_cycle = self.dominant,
            .bins_len = self.bins_len,
        };
        @memcpy(out_state.cosine_parts_storage[0..self.bins_len], self.cosine_parts[0..self.bins_len]);
        @memcpy(out_state.sine_parts_storage[0..self.bins_len], self.sine_parts[0..self.bins_len]);
        @memcpy(out_state.powers_storage[0..self.bins_len], self.powers[0..self.bins_len]);
        @memcpy(out_state.decibels_storage[0..self.bins_len], self.decibels[0..self.bins_len]);
        return out_state;
    }

    pub fn output(self: *const TransformedDiscreteFourierTransformSpectrum) SpectrumOutput {
        return self.last_output;
    }

    pub fn tryUpdate(self: *TransformedDiscreteFourierTransformSpectrum, sample: f64, time: i64) Error!SpectrumOutput {
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

        const bar_number = self.sample_count + 1;
        var high_pass: f64 = undefined;
        var cleaned: f64 = undefined;
        if (bar_number <= 5) {
            high_pass = sample;
            cleaned = sample;
        } else {
            high_pass = 0.5 * (1.0 + self.highpass_alpha) * (sample - self.previous_price) +
                self.highpass_alpha * self.hp_history[0];
            cleaned = (high_pass +
                2.0 * self.hp_history[0] +
                3.0 * self.hp_history[1] +
                3.0 * self.hp_history[2] +
                2.0 * self.hp_history[3] +
                self.hp_history[4]) / 12.0;
        }
        if (!math.isFinite(high_pass) or !math.isFinite(cleaned)) {
            return error.NumericOverflow;
        }

        self.next_window[0] = cleaned;
        if (self.resolved.window_length > 1) {
            @memcpy(
                self.next_window[1..self.resolved.window_length],
                self.window[0 .. self.resolved.window_length - 1],
            );
        }

        var next_cosines: [heatmap_mod.max_heatmap_values]f64 = undefined;
        var next_sines: [heatmap_mod.max_heatmap_values]f64 = undefined;
        var next_powers: [heatmap_mod.max_heatmap_values]f64 = undefined;
        var maximum: f64 = 0.0;
        for (0..self.bins_len) |bin_index| {
            var cosine_sum: f64 = 0.0;
            var sine_sum: f64 = 0.0;
            const row_offset = bin_index * self.resolved.window_length;
            for (0..self.resolved.window_length) |lag| {
                cosine_sum += self.next_window[lag] * self.cosines[row_offset + lag];
                sine_sum += self.next_window[lag] * self.sines[row_offset + lag];
            }
            const power = cosine_sum * cosine_sum + sine_sum * sine_sum;
            if (!math.isFinite(cosine_sum) or !math.isFinite(sine_sum) or !math.isFinite(power)) {
                return error.NumericOverflow;
            }
            next_cosines[bin_index] = cosine_sum;
            next_sines[bin_index] = sine_sum;
            next_powers[bin_index] = power;
            if (bin_index == 0 or power > maximum) {
                maximum = power;
            }
        }

        var next_decibels: [heatmap_mod.max_heatmap_values]f64 = undefined;
        @memcpy(next_decibels[0..self.bins_len], self.decibels[0..self.bins_len]);
        for (0..self.bins_len) |bin_index| {
            const power = next_powers[bin_index];
            if (maximum > 0.0 and power > 0.0) {
                next_decibels[bin_index] = -10.0 * @log(0.01 / (1.0 - 0.99 * power / maximum)) / log_ten;
            }
            if (next_decibels[bin_index] > db_cap) {
                next_decibels[bin_index] = db_cap;
            }
            if (!math.isFinite(next_decibels[bin_index])) {
                return error.NumericOverflow;
            }
        }

        var numerator: f64 = 0.0;
        var denominator: f64 = 0.0;
        for (0..self.bins_len) |bin_index| {
            const channel_db = next_decibels[bin_index];
            if (channel_db < centroid_db_limit) {
                const weight = centroid_db_limit - channel_db;
                const period_f: f64 = @floatFromInt(self.resolved.min_period + bin_index);
                numerator += period_f * weight;
                denominator += weight;
            }
        }
        const dominant = if (denominator != 0.0) numerator / denominator else self.dominant;
        if (!math.isFinite(maximum) or !math.isFinite(dominant)) {
            return error.NumericOverflow;
        }

        self.previous_price = sample;
        var lag_index: usize = 5;
        while (lag_index >= 1) : (lag_index -= 1) {
            self.hp_history[lag_index] = self.hp_history[lag_index - 1];
        }
        self.hp_history[0] = high_pass;
        self.cleaned = cleaned;
        @memcpy(self.window, self.next_window);
        @memcpy(self.cosine_parts[0..self.bins_len], next_cosines[0..self.bins_len]);
        @memcpy(self.sine_parts[0..self.bins_len], next_sines[0..self.bins_len]);
        @memcpy(self.powers[0..self.bins_len], next_powers[0..self.bins_len]);
        @memcpy(self.decibels[0..self.bins_len], next_decibels[0..self.bins_len]);
        self.maximum = maximum;
        self.dominant = dominant;
        self.sample_count = bar_number;

        const current_output: SpectrumOutput = if (self.sample_count > self.warmup_count) blk: {
            const active_values = next_decibels[0..self.bins_len];
            var value_min = active_values[0];
            var value_max = active_values[0];
            for (active_values[1..]) |item| {
                if (item < value_min) value_min = item;
                if (item > value_max) value_max = item;
            }
            break :blk .{
                .heatmap = Heatmap.new(time, min_f, max_f, 1.0, value_min, value_max, active_values),
                .dominant_cycle = dominant,
            };
        } else .{
            .heatmap = Heatmap.empty(time, min_f, max_f, 1.0),
            .dominant_cycle = math.nan(f64),
        };

        self.last_output = current_output;
        return current_output;
    }

    pub fn update(self: *TransformedDiscreteFourierTransformSpectrum, sample: f64, time: i64) SpectrumOutput {
        return self.tryUpdate(sample, time) catch |err| {
            std.debug.panic("TransformedDiscreteFourierTransformSpectrum.update failed: {s}", .{@errorName(err)});
        };
    }

    pub const updateSample = update;

    fn wrapEntity(time: i64, out_value: SpectrumOutput) OutputArray {
        var arr = OutputArray{};
        arr.append(.{ .heatmap = out_value.heatmap });
        arr.append(.{ .scalar = .{ .time = time, .value = out_value.dominant_cycle } });
        return arr;
    }

    pub fn updateScalar(self: *TransformedDiscreteFourierTransformSpectrum, sample: *const Scalar) OutputArray {
        return wrapEntity(sample.time, self.update(sample.value, sample.time));
    }

    pub fn updateBar(self: *TransformedDiscreteFourierTransformSpectrum, sample: *const Bar) OutputArray {
        return wrapEntity(sample.time, self.update(self.bar_func(sample.*), sample.time));
    }

    pub fn updateQuote(self: *TransformedDiscreteFourierTransformSpectrum, sample: *const Quote) OutputArray {
        return wrapEntity(sample.time, self.update(self.quote_func(sample.*), sample.time));
    }

    pub fn updateTrade(self: *TransformedDiscreteFourierTransformSpectrum, sample: *const Trade) OutputArray {
        return wrapEntity(sample.time, self.update(self.trade_func(sample.*), sample.time));
    }

    pub fn isPrimed(self: *const TransformedDiscreteFourierTransformSpectrum) bool {
        return self.sample_count > self.warmup_count;
    }

    pub fn getMetadata(self: *const TransformedDiscreteFourierTransformSpectrum, out: *Metadata) void {
        const mn = self.mnemonicText();
        const desc = self.descriptionText();
        const texts = [_]build_metadata_mod.OutputText{
            .{ .mnemonic = mn, .description = desc },
            .{
                .mnemonic = self.dominantCycleMnemonicText(),
                .description = self.dominantCycleDescriptionText(),
            },
        };
        build_metadata_mod.buildMetadata(out, .transformed_discrete_fourier_transform_spectrum, mn, desc, &texts);
    }

    pub fn indicator(self: *TransformedDiscreteFourierTransformSpectrum) indicator_mod.Indicator {
        return .{
            .ptr = @ptrCast(self),
            .vtable = &vtable,
        };
    }

    const vtable = indicator_mod.Indicator.GenVTable(TransformedDiscreteFourierTransformSpectrum);
};

// ---------------------------------------------------------------------------
// Tests
// ---------------------------------------------------------------------------
const testing = std.testing;
const testdata = @import("testdata.zig");

const fixture_tolerance: f64 = 1e-12;

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
    try expectCloseTol(actual, expected, fixture_tolerance);
}

fn expectSameOutput(actual: *const SpectrumOutput, expected: *const SpectrumOutput) !void {
    try testing.expectEqual(expected.heatmap.time, actual.heatmap.time);
    try testing.expectEqual(expected.heatmap.parameter_first, actual.heatmap.parameter_first);
    try testing.expectEqual(expected.heatmap.parameter_last, actual.heatmap.parameter_last);
    try testing.expectEqual(expected.heatmap.parameter_resolution, actual.heatmap.parameter_resolution);
    try testing.expectEqual(expected.heatmap.isEmpty(), actual.heatmap.isEmpty());
    try expectCloseTol(actual.heatmap.value_min, expected.heatmap.value_min, 0.0);
    try expectCloseTol(actual.heatmap.value_max, expected.heatmap.value_max, 0.0);
    const lhs = actual.heatmap.valuesSlice();
    const rhs = expected.heatmap.valuesSlice();
    try testing.expectEqual(rhs.len, lhs.len);
    for (0..lhs.len) |bin_index| {
        try expectCloseTol(lhs[bin_index], rhs[bin_index], 0.0);
    }
    try expectCloseTol(actual.dominant_cycle, expected.dominant_cycle, 0.0);
}

fn verifyReferenceCase(
    comptime cols: usize,
    params: Params,
    input_close: *const [252]f64,
    expected_heatmap: *const [252][cols]f64,
    expected_min: *const [252]f64,
    expected_max: *const [252]f64,
    expected_dc: *const [252]f64,
) !void {
    var indicator_instance = try TransformedDiscreteFourierTransformSpectrum.init(testing.allocator, params);
    defer indicator_instance.deinit();

    for (0..252) |bar_index| {
        const timestamp: i64 = @intCast(bar_index + 1);
        const result = indicator_instance.update(input_close[bar_index], timestamp);
        try testing.expectEqual(timestamp, result.heatmap.time);
        try testing.expectEqual(@as(f64, @floatFromInt(params.min_period)), result.heatmap.parameter_first);
        try testing.expectEqual(@as(f64, @floatFromInt(params.max_period)), result.heatmap.parameter_last);
        try testing.expectEqual(@as(f64, 1.0), result.heatmap.parameter_resolution);

        if (math.isNan(expected_dc[bar_index])) {
            try testing.expect(!indicator_instance.isPrimed());
            try testing.expect(result.heatmap.isEmpty());
            try testing.expect(math.isNan(result.heatmap.value_min));
            try testing.expect(math.isNan(result.heatmap.value_max));
            try testing.expect(math.isNan(result.dominant_cycle));
            for (expected_heatmap[bar_index]) |cell| {
                try testing.expect(math.isNan(cell));
            }
        } else {
            try testing.expect(indicator_instance.isPrimed());
            try testing.expect(!result.heatmap.isEmpty());
            const values = result.heatmap.valuesSlice();
            try testing.expectEqual(cols, values.len);
            try expectClose(result.heatmap.value_min, expected_min[bar_index]);
            try expectClose(result.heatmap.value_max, expected_max[bar_index]);
            try expectClose(result.dominant_cycle, expected_dc[bar_index]);
            for (0..cols) |bin_index| {
                try expectClose(values[bin_index], expected_heatmap[bar_index][bin_index]);
            }
        }
    }
}

test "transformed discrete fourier transform spectrum all 12 reference cases" {
    const input_close = testdata.testInput();

    {
        const hm = testdata.expectedP8_Q50_W50_HP40_HEATMAP();
        const mn = testdata.expectedP8_Q50_W50_HP40_MIN();
        const mx = testdata.expectedP8_Q50_W50_HP40_MAX();
        const dc = testdata.expectedP8_Q50_W50_HP40_DOMINANT_CYCLE();
        try verifyReferenceCase(43, .{ .min_period = 8, .max_period = 50, .window_length = 50, .highpass_period = 40.0 }, &input_close, &hm, &mn, &mx, &dc);
    }
    {
        const hm = testdata.expectedP8_Q50_W1_HP40_HEATMAP();
        const mn = testdata.expectedP8_Q50_W1_HP40_MIN();
        const mx = testdata.expectedP8_Q50_W1_HP40_MAX();
        const dc = testdata.expectedP8_Q50_W1_HP40_DOMINANT_CYCLE();
        try verifyReferenceCase(43, .{ .min_period = 8, .max_period = 50, .window_length = 1, .highpass_period = 40.0 }, &input_close, &hm, &mn, &mx, &dc);
    }
    {
        const hm = testdata.expectedP8_Q50_W20_HP40_HEATMAP();
        const mn = testdata.expectedP8_Q50_W20_HP40_MIN();
        const mx = testdata.expectedP8_Q50_W20_HP40_MAX();
        const dc = testdata.expectedP8_Q50_W20_HP40_DOMINANT_CYCLE();
        try verifyReferenceCase(43, .{ .min_period = 8, .max_period = 50, .window_length = 20, .highpass_period = 40.0 }, &input_close, &hm, &mn, &mx, &dc);
    }
    {
        const hm = testdata.expectedP8_Q50_W48_HP40_HEATMAP();
        const mn = testdata.expectedP8_Q50_W48_HP40_MIN();
        const mx = testdata.expectedP8_Q50_W48_HP40_MAX();
        const dc = testdata.expectedP8_Q50_W48_HP40_DOMINANT_CYCLE();
        try verifyReferenceCase(43, .{ .min_period = 8, .max_period = 50, .window_length = 48, .highpass_period = 40.0 }, &input_close, &hm, &mn, &mx, &dc);
    }
    {
        const hm = testdata.expectedP8_Q50_W64_HP40_HEATMAP();
        const mn = testdata.expectedP8_Q50_W64_HP40_MIN();
        const mx = testdata.expectedP8_Q50_W64_HP40_MAX();
        const dc = testdata.expectedP8_Q50_W64_HP40_DOMINANT_CYCLE();
        try verifyReferenceCase(43, .{ .min_period = 8, .max_period = 50, .window_length = 64, .highpass_period = 40.0 }, &input_close, &hm, &mn, &mx, &dc);
    }
    {
        const hm = testdata.expectedP8_Q50_W252_HP40_HEATMAP();
        const mn = testdata.expectedP8_Q50_W252_HP40_MIN();
        const mx = testdata.expectedP8_Q50_W252_HP40_MAX();
        const dc = testdata.expectedP8_Q50_W252_HP40_DOMINANT_CYCLE();
        try verifyReferenceCase(43, .{ .min_period = 8, .max_period = 50, .window_length = 252, .highpass_period = 40.0 }, &input_close, &hm, &mn, &mx, &dc);
    }
    {
        const hm = testdata.expectedP8_Q50_W50_HP20_HEATMAP();
        const mn = testdata.expectedP8_Q50_W50_HP20_MIN();
        const mx = testdata.expectedP8_Q50_W50_HP20_MAX();
        const dc = testdata.expectedP8_Q50_W50_HP20_DOMINANT_CYCLE();
        try verifyReferenceCase(43, .{ .min_period = 8, .max_period = 50, .window_length = 50, .highpass_period = 20.0 }, &input_close, &hm, &mn, &mx, &dc);
    }
    {
        const hm = testdata.expectedP8_Q50_W50_HP80_HEATMAP();
        const mn = testdata.expectedP8_Q50_W50_HP80_MIN();
        const mx = testdata.expectedP8_Q50_W50_HP80_MAX();
        const dc = testdata.expectedP8_Q50_W50_HP80_DOMINANT_CYCLE();
        try verifyReferenceCase(43, .{ .min_period = 8, .max_period = 50, .window_length = 50, .highpass_period = 80.0 }, &input_close, &hm, &mn, &mx, &dc);
    }
    {
        const hm = testdata.expectedP8_Q50_W50_HP40_5_HEATMAP();
        const mn = testdata.expectedP8_Q50_W50_HP40_5_MIN();
        const mx = testdata.expectedP8_Q50_W50_HP40_5_MAX();
        const dc = testdata.expectedP8_Q50_W50_HP40_5_DOMINANT_CYCLE();
        try verifyReferenceCase(43, .{ .min_period = 8, .max_period = 50, .window_length = 50, .highpass_period = 40.5 }, &input_close, &hm, &mn, &mx, &dc);
    }
    {
        const hm = testdata.expectedP10_Q30_W50_HP40_HEATMAP();
        const mn = testdata.expectedP10_Q30_W50_HP40_MIN();
        const mx = testdata.expectedP10_Q30_W50_HP40_MAX();
        const dc = testdata.expectedP10_Q30_W50_HP40_DOMINANT_CYCLE();
        try verifyReferenceCase(21, .{ .min_period = 10, .max_period = 30, .window_length = 50, .highpass_period = 40.0 }, &input_close, &hm, &mn, &mx, &dc);
    }
    {
        const hm = testdata.expectedP8_Q80_W50_HP40_HEATMAP();
        const mn = testdata.expectedP8_Q80_W50_HP40_MIN();
        const mx = testdata.expectedP8_Q80_W50_HP40_MAX();
        const dc = testdata.expectedP8_Q80_W50_HP40_DOMINANT_CYCLE();
        try verifyReferenceCase(73, .{ .min_period = 8, .max_period = 80, .window_length = 50, .highpass_period = 40.0 }, &input_close, &hm, &mn, &mx, &dc);
    }
    {
        const hm = testdata.expectedP20_Q20_W50_HP40_HEATMAP();
        const mn = testdata.expectedP20_Q20_W50_HP40_MIN();
        const mx = testdata.expectedP20_Q20_W50_HP40_MAX();
        const dc = testdata.expectedP20_Q20_W50_HP40_DOMINANT_CYCLE();
        try verifyReferenceCase(1, .{ .min_period = 20, .max_period = 20, .window_length = 50, .highpass_period = 40.0 }, &input_close, &hm, &mn, &mx, &dc);
    }
}

test "transformed discrete fourier transform spectrum missing invalid overflow and reset" {
    const input_close = testdata.testInput();
    var stream = try TransformedDiscreteFourierTransformSpectrum.initDefault(testing.allocator);
    defer stream.deinit();
    var control = try TransformedDiscreteFourierTransformSpectrum.initDefault(testing.allocator);
    defer control.deinit();

    for (0..80) |bar_index| {
        _ = stream.update(input_close[bar_index], @intCast(bar_index));
        _ = control.update(input_close[bar_index], @intCast(bar_index));
    }
    const before_state = stream.state();
    const before_samples = stream.samplesSeen();
    const before_output = stream.output();

    const missing = stream.update(math.nan(f64), 999);
    try testing.expect(missing.heatmap.isEmpty());
    try testing.expectEqual(@as(i64, 999), missing.heatmap.time);
    try testing.expect(math.isNan(missing.dominant_cycle));

    try testing.expectError(error.InvalidSample, stream.tryUpdate(math.inf(f64), 1));
    try testing.expectError(error.InvalidSample, stream.tryUpdate(-math.inf(f64), 1));
    try testing.expectError(error.NumericOverflow, stream.tryUpdate(1e308, 1));
    try testing.expectEqual(before_samples, stream.samplesSeen());
    try testing.expect(stream.state().eql(&before_state));
    const after_output = stream.output();
    try expectSameOutput(&after_output, &before_output);

    const s80 = stream.update(input_close[80], 80);
    const c80 = control.update(input_close[80], 80);
    try expectSameOutput(&s80, &c80);

    stream.reset();
    var fresh = try TransformedDiscreteFourierTransformSpectrum.initDefault(testing.allocator);
    defer fresh.deinit();
    for (0..252) |bar_index| {
        const actual = stream.update(input_close[bar_index], @intCast(bar_index));
        const expected = fresh.update(input_close[bar_index], @intCast(bar_index));
        try expectSameOutput(&actual, &expected);
    }
}

test "transformed discrete fourier transform spectrum validation mnemonics metadata and entities" {
    try testing.expectError(error.InvalidMinPeriod, TransformedDiscreteFourierTransformSpectrum.init(testing.allocator, .{ .min_period = 7 }));
    try testing.expectError(error.InvalidMaxPeriod, TransformedDiscreteFourierTransformSpectrum.init(testing.allocator, .{ .min_period = 20, .max_period = 19 }));
    try testing.expectError(error.SpectrumRangeTooLarge, TransformedDiscreteFourierTransformSpectrum.init(testing.allocator, .{ .min_period = 8, .max_period = 300 }));
    try testing.expectError(error.InvalidWindowLength, TransformedDiscreteFourierTransformSpectrum.init(testing.allocator, .{ .window_length = 0 }));
    try testing.expectError(error.InvalidHighpassPeriod, TransformedDiscreteFourierTransformSpectrum.init(testing.allocator, .{ .highpass_period = 4.0 }));
    try testing.expectError(error.InvalidHighpassPeriod, TransformedDiscreteFourierTransformSpectrum.init(testing.allocator, .{ .highpass_period = math.nan(f64) }));
    try testing.expectError(error.InvalidHighpassPeriod, TransformedDiscreteFourierTransformSpectrum.init(testing.allocator, .{ .highpass_period = math.inf(f64) }));

    const MnemonicCase = struct {
        params: Params,
        want_main: []const u8,
        want_dc: []const u8,
    };
    const mnemonic_cases = [_]MnemonicCase{
        .{ .params = defaultParams(), .want_main = "tdfts(8, 50, 50, 40)", .want_dc = "tdfts-dc(8, 50, 50, 40)" },
        .{
            .params = .{ .min_period = 10, .max_period = 30, .window_length = 48, .highpass_period = 40.5, .bar_component = .median },
            .want_main = "tdfts(10, 30, 48, 40.5, hl/2)",
            .want_dc = "tdfts-dc(10, 30, 48, 40.5, hl/2)",
        },
        .{
            .params = .{ .quote_component = .ask },
            .want_main = "tdfts(8, 50, 50, 40, a)",
            .want_dc = "tdfts-dc(8, 50, 50, 40, a)",
        },
        .{
            .params = .{ .trade_component = .volume },
            .want_main = "tdfts(8, 50, 50, 40, v)",
            .want_dc = "tdfts-dc(8, 50, 50, 40, v)",
        },
        .{
            .params = .{ .min_period = 10, .max_period = 30, .window_length = 48, .highpass_period = 40.5, .bar_component = .median, .quote_component = .ask },
            .want_main = "tdfts(10, 30, 48, 40.5, hl/2, a)",
            .want_dc = "tdfts-dc(10, 30, 48, 40.5, hl/2, a)",
        },
        .{
            .params = .{ .quote_component = .bid, .trade_component = .volume },
            .want_main = "tdfts(8, 50, 50, 40, b, v)",
            .want_dc = "tdfts-dc(8, 50, 50, 40, b, v)",
        },
    };

    for (mnemonic_cases) |case_item| {
        var mn_buf: [64]u8 = undefined;
        var dc_buf: [64]u8 = undefined;
        try testing.expectEqualStrings(case_item.want_main, try mnemonic(&mn_buf, case_item.params));
        try testing.expectEqualStrings(case_item.want_dc, try dominantCycleMnemonic(&dc_buf, case_item.params));

        var inst = try TransformedDiscreteFourierTransformSpectrum.init(testing.allocator, case_item.params);
        defer inst.deinit();
        var md: Metadata = undefined;
        inst.getMetadata(&md);
        try testing.expectEqual(Identifier.transformed_discrete_fourier_transform_spectrum, md.identifier);
        try testing.expectEqual(@as(usize, 2), md.outputs_len);
        try testing.expectEqual(shape_mod.Shape.heatmap, md.outputs_buf[0].shape);
        try testing.expectEqualStrings(case_item.want_main, md.outputs_buf[0].mnemonic);
        try testing.expectEqual(shape_mod.Shape.scalar, md.outputs_buf[1].shape);
        try testing.expectEqualStrings(case_item.want_dc, md.outputs_buf[1].mnemonic);
    }

    const desc = descriptor_mod.descriptorOf(.transformed_discrete_fourier_transform_spectrum) orelse return error.TestUnexpectedResult;
    try testing.expectEqualStrings("John Ehlers", desc.family);
    try testing.expectEqual(@as(usize, 2), desc.outputs.len);

    const input_close = testdata.testInput();
    const input_open = testdata.testOpen();
    const input_high = testdata.testHigh();
    const input_low = testdata.testLow();
    var scalar_ind = try TransformedDiscreteFourierTransformSpectrum.initDefault(testing.allocator);
    defer scalar_ind.deinit();
    var bar_ind = try TransformedDiscreteFourierTransformSpectrum.initDefault(testing.allocator);
    defer bar_ind.deinit();
    var quote_ind = try TransformedDiscreteFourierTransformSpectrum.initDefault(testing.allocator);
    defer quote_ind.deinit();
    var trade_ind = try TransformedDiscreteFourierTransformSpectrum.initDefault(testing.allocator);
    defer trade_ind.deinit();

    for (0..252) |bar_index| {
        const timestamp: i64 = @intCast(bar_index);
        const close_price = input_close[bar_index];
        const sc = Scalar{ .time = timestamp, .value = close_price };
        const br = Bar{
            .time = timestamp,
            .open = input_open[bar_index],
            .high = input_high[bar_index],
            .low = input_low[bar_index],
            .close = close_price,
            .volume = 1000.0,
        };
        const qt = Quote{
            .time = timestamp,
            .bid_price = close_price,
            .ask_price = close_price,
            .bid_size = 10.0,
            .ask_size = 10.0,
        };
        const tr = Trade{
            .time = timestamp,
            .price = close_price,
            .volume = 10.0,
        };
        const out_sc = scalar_ind.indicator().updateScalar(&sc);
        const out_br = bar_ind.indicator().updateBar(&br);
        const out_qt = quote_ind.indicator().updateQuote(&qt);
        const out_tr = trade_ind.indicator().updateTrade(&tr);
        try expectCloseTol(out_sc.slice()[1].scalar.value, out_br.slice()[1].scalar.value, 0.0);
        try expectCloseTol(out_sc.slice()[1].scalar.value, out_qt.slice()[1].scalar.value, 0.0);
        try expectCloseTol(out_sc.slice()[1].scalar.value, out_tr.slice()[1].scalar.value, 0.0);
    }
}
