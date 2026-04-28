const std = @import("std");
const math = std.math;
const Indicator = @import("indicator.zig").Indicator;
const OutputArray = @import("indicator.zig").OutputArray;
const Metadata = @import("metadata.zig").Metadata;
const Scalar = @import("scalar").Scalar;

/// A single calculated filter frequency response component data.
pub const Component = struct {
    data: []f64,
    min: f64,
    max: f64,
};

/// Calculated filter frequency response data.
/// All slices have the same spectrum length.
pub const FrequencyResponse = struct {
    /// Mnemonic of the filter used to calculate the frequency response.
    label: []const u8,
    /// Frequency in units of cycles per 2 samples, 1 being the Nyquist frequency.
    normalized_frequency: []f64,
    /// Spectrum power in percentages from a maximum value.
    power_percent: Component,
    /// Spectrum power in decibels.
    power_decibel: Component,
    /// Spectrum amplitude in percentages from a maximum value.
    amplitude_percent: Component,
    /// Spectrum amplitude in decibels.
    amplitude_decibel: Component,
    /// Phase in degrees in range [-180, 180].
    phase_degrees: Component,
    /// Phase in degrees unwrapped.
    phase_degrees_unwrapped: Component,

    /// Label storage buffer (owned copy of mnemonic).
    label_buf: [256]u8 = undefined,

    pub fn deinit(self: *FrequencyResponse, allocator: std.mem.Allocator) void {
        allocator.free(self.normalized_frequency);
        allocator.free(self.power_percent.data);
        allocator.free(self.power_decibel.data);
        allocator.free(self.amplitude_percent.data);
        allocator.free(self.amplitude_decibel.data);
        allocator.free(self.phase_degrees.data);
        allocator.free(self.phase_degrees_unwrapped.data);
    }
};

pub const CalculateError = error{
    InvalidSignalLength,
    OutOfMemory,
};

/// Calculates a frequency response of a given impulse signal length
/// using the filter update function.
///
/// The warm-up counter specifies how many times to update filter with zero value
/// before calculations.
///
/// The phase_degrees_unwrapping_limit controls the phase unwrapping threshold (use 179 as default).
///
/// The impulse signal length should be an integer of a power of 2 and be greater than 4.
/// Realistic values are 512, 1024, 2048, 4096.
pub fn calculate(
    allocator: std.mem.Allocator,
    signal_length: usize,
    filter: Indicator,
    warmup: usize,
    phase_degrees_unwrapping_limit: f64,
) CalculateError!FrequencyResponse {
    if (!isValidSignalLength(signal_length)) {
        return CalculateError.InvalidSignalLength;
    }

    const spectrum_length = signal_length / 2 - 1;

    // Get label from metadata.
    var meta: Metadata = undefined;
    filter.metadata(&meta);

    var fr = FrequencyResponse{
        .label = undefined,
        .normalized_frequency = try allocator.alloc(f64, spectrum_length),
        .power_percent = .{ .data = try allocator.alloc(f64, spectrum_length), .min = math.inf(f64), .max = -math.inf(f64) },
        .power_decibel = .{ .data = try allocator.alloc(f64, spectrum_length), .min = math.inf(f64), .max = -math.inf(f64) },
        .amplitude_percent = .{ .data = try allocator.alloc(f64, spectrum_length), .min = math.inf(f64), .max = -math.inf(f64) },
        .amplitude_decibel = .{ .data = try allocator.alloc(f64, spectrum_length), .min = math.inf(f64), .max = -math.inf(f64) },
        .phase_degrees = .{ .data = try allocator.alloc(f64, spectrum_length), .min = math.inf(f64), .max = -math.inf(f64) },
        .phase_degrees_unwrapped = .{ .data = try allocator.alloc(f64, spectrum_length), .min = math.inf(f64), .max = -math.inf(f64) },
    };

    // Copy label into owned buffer.
    const label_len = @min(meta.mnemonic.len, fr.label_buf.len);
    @memcpy(fr.label_buf[0..label_len], meta.mnemonic[0..label_len]);
    fr.label = fr.label_buf[0..label_len];

    prepareFrequencyDomain(spectrum_length, fr.normalized_frequency);

    const signal = try prepareFilteredSignal(allocator, signal_length, filter, warmup);
    defer allocator.free(signal);

    directRealFastFourierTransform(signal);
    parseSpectrum(
        spectrum_length,
        signal,
        &fr.power_percent,
        &fr.amplitude_percent,
        &fr.phase_degrees,
        &fr.phase_degrees_unwrapped,
        phase_degrees_unwrapping_limit,
    );
    toDecibels(spectrum_length, &fr.power_percent, &fr.power_decibel);
    toPercents(spectrum_length, &fr.power_percent, &fr.power_percent);
    toDecibels(spectrum_length, &fr.amplitude_percent, &fr.amplitude_decibel);
    toPercents(spectrum_length, &fr.amplitude_percent, &fr.amplitude_percent);

    return fr;
}

fn isValidSignalLength(length: usize) bool {
    var l = length;
    while (l > 4) {
        if (l % 2 != 0) return false;
        l /= 2;
    }
    return l == 4;
}

fn prepareFrequencyDomain(spectrum_length: usize, freq: []f64) void {
    for (0..spectrum_length) |i| {
        freq[i] = @as(f64, @floatFromInt(1 + i)) / @as(f64, @floatFromInt(spectrum_length));
    }
}

fn prepareFilteredSignal(
    allocator: std.mem.Allocator,
    signal_length: usize,
    filter: Indicator,
    warmup: usize,
) ![]f64 {
    const zero_scalar = Scalar{ .time = 0, .value = 0 };
    const impulse_scalar = Scalar{ .time = 0, .value = 1000 };

    for (0..warmup) |_| {
        _ = filter.updateScalar(&zero_scalar);
    }

    const signal = try allocator.alloc(f64, signal_length);

    const out0 = filter.updateScalar(&impulse_scalar);
    signal[0] = extractScalarValue(out0);

    for (1..signal_length) |i| {
        const out = filter.updateScalar(&zero_scalar);
        signal[i] = extractScalarValue(out);
    }

    return signal;
}

fn extractScalarValue(out: OutputArray) f64 {
    if (out.len > 0) {
        switch (out.values[0]) {
            .scalar => |s| return s.value,
            else => return 0,
        }
    }
    return 0;
}

fn parseSpectrum(
    length: usize,
    signal: []f64,
    power: *Component,
    amplitude: *Component,
    phase: *Component,
    phase_unwrapped: *Component,
    phase_degrees_unwrapping_limit: f64,
) void {
    const rad2deg: f64 = 180.0 / math.pi;

    var pmin: f64 = math.inf(f64);
    var pmax: f64 = -math.inf(f64);
    var amin: f64 = math.inf(f64);
    var amax: f64 = -math.inf(f64);

    var k: usize = 2;
    for (0..length) |i| {
        const re = signal[k];
        k += 1;
        const im = signal[k];
        k += 1;

        // Wrapped phase -- atan2 returns radians in the [-pi, pi] range.
        phase.data[i] = -math.atan2(im, re) * rad2deg;
        phase_unwrapped.data[i] = 0;

        const pwr = re * re + im * im;
        power.data[i] = pwr;
        pmin = @min(pmin, pwr);
        pmax = @max(pmax, pwr);

        const amp = @sqrt(pwr);
        amplitude.data[i] = amp;
        amin = @min(amin, amp);
        amax = @max(amax, amp);
    }

    unwrapPhaseDegrees(length, phase.data, phase_unwrapped, phase_degrees_unwrapping_limit);
    phase.min = -180;
    phase.max = 180;
    power.min = pmin;
    power.max = pmax;
    amplitude.min = amin;
    amplitude.max = amax;
}

fn unwrapPhaseDegrees(length: usize, wrapped: []f64, unwrapped: *Component, limit: f64) void {
    var k_val: f64 = 0;

    var min_val = wrapped[0];
    var max_val = min_val;
    unwrapped.data[0] = min_val;

    for (1..length) |i| {
        var w = wrapped[i];
        const increment = wrapped[i] - wrapped[i - 1];

        if (increment > limit) {
            k_val -= increment;
        } else if (increment < -limit) {
            k_val += increment;
        }

        w += k_val;
        min_val = @min(min_val, w);
        max_val = @max(max_val, w);
        unwrapped.data[i] = w;
    }

    unwrapped.min = min_val;
    unwrapped.max = max_val;
}

fn toDecibels(length: usize, src: *Component, tgt: *Component) void {
    var dbmin: f64 = math.inf(f64);
    var dbmax: f64 = -math.inf(f64);

    var base = src.data[0];
    if (base < math.floatMin(f64)) {
        base = src.max;
    }

    for (0..length) |i| {
        const db = 20.0 * @log10(src.data[i] / base);
        dbmin = @min(dbmin, db);
        dbmax = @max(dbmax, db);
        tgt.data[i] = db;
    }

    // If dbmin falls into one of [-100, -90), [-90, -80), ..., [-10, 0)
    // intervals, set it to the minimum value of the interval.
    {
        var ii: i32 = 10;
        while (ii > 0) : (ii -= 1) {
            const min_bound = -@as(f64, @floatFromInt(ii)) * 10.0;
            const max_bound = -@as(f64, @floatFromInt(ii - 1)) * 10.0;
            if (dbmin >= min_bound and dbmin < max_bound) {
                dbmin = min_bound;
                break;
            }
        }
    }

    // Limit all minimal decibel values to -100.
    if (dbmin < -100.0) {
        dbmin = -100.0;
        for (0..length) |i| {
            if (tgt.data[i] < -100.0) {
                tgt.data[i] = -100.0;
            }
        }
    }

    // If dbmax falls into one of [0, 5), [5, 10)
    // intervals, set it to the maximum value of the interval.
    {
        var ii: i32 = 2;
        while (ii > 0) : (ii -= 1) {
            const max_bound = @as(f64, @floatFromInt(ii)) * 5.0;
            const min_bound = @as(f64, @floatFromInt(ii - 1)) * 5.0;
            if (dbmax >= min_bound and dbmax < max_bound) {
                dbmax = max_bound;
                break;
            }
        }
    }

    // Limit all maximal decibel values to 10.
    if (dbmax > 10.0) {
        dbmax = 10.0;
        for (0..length) |i| {
            if (tgt.data[i] > 10.0) {
                tgt.data[i] = 10.0;
            }
        }
    }

    tgt.min = dbmin;
    tgt.max = dbmax;
}

fn toPercents(length: usize, src: *Component, tgt: *Component) void {
    var pctmax: f64 = -math.inf(f64);

    var base = src.data[0];
    if (base < math.floatMin(f64)) {
        base = src.max;
    }

    for (0..length) |i| {
        const pct = 100.0 * src.data[i] / base;
        pctmax = @max(pctmax, pct);
        tgt.data[i] = pct;
    }

    // If pctmax falls into one of [100, 110), [110, 120), ..., [190, 200)
    // intervals, set it to the maximum value of the interval.
    for (0..10) |i| {
        const min_bound = 100.0 + @as(f64, @floatFromInt(i)) * 10.0;
        const max_bound = 100.0 + @as(f64, @floatFromInt(i + 1)) * 10.0;
        if (pctmax >= min_bound and pctmax < max_bound) {
            pctmax = max_bound;
            break;
        }
    }

    // Limit all maximal percentage values to 200.
    if (pctmax > 200.0) {
        pctmax = 200.0;
        for (0..length) |i| {
            if (tgt.data[i] > 200.0) {
                tgt.data[i] = 200.0;
            }
        }
    }

    tgt.min = 0;
    tgt.max = pctmax;
}

/// Performs a direct real fast Fourier transform.
///
/// The input parameter is a data array containing real data on input
/// and {re,im} pairs on return.
///
/// The length of the input data slice must be a power of 2.
fn directRealFastFourierTransform(array: []f64) void {
    const length = array.len;
    const two_pi = 2.0 * math.pi;
    const ttheta = two_pi / @as(f64, @floatFromInt(length));
    const nn = length / 2;
    var j: usize = 1;

    for (1..nn + 1) |ii| {
        const i = 2 * ii - 1;

        if (j > i) {
            const temp_r = array[j - 1];
            const temp_i = array[j];
            array[j - 1] = array[i - 1];
            array[j] = array[i];
            array[i - 1] = temp_r;
            array[i] = temp_i;
        }

        var m = nn;
        while (m >= 2 and j > m) {
            j -= m;
            m /= 2;
        }
        j += m;
    }

    var m_max: usize = 2;

    while (length > m_max) {
        const istep = 2 * m_max;
        const theta = two_pi / @as(f64, @floatFromInt(m_max));
        var wp_r = @sin(0.5 * theta);
        wp_r = -2.0 * wp_r * wp_r;
        const wp_i = @sin(theta);
        var w_r: f64 = 1.0;
        var w_i: f64 = 0.0;

        for (1..m_max / 2 + 1) |ii| {
            const m = 2 * ii - 1;
            var jj: usize = 0;
            while (jj <= (length - m) / istep) : (jj += 1) {
                const i = m + jj * istep;
                j = i + m_max;
                const temp_r = w_r * array[j - 1] - w_i * array[j];
                const temp_i = w_r * array[j] + w_i * array[j - 1];
                array[j - 1] = array[i - 1] - temp_r;
                array[j] = array[i] - temp_i;
                array[i - 1] = array[i - 1] + temp_r;
                array[i] = array[i] + temp_i;
            }

            const w_temp = w_r;
            w_r = w_r * wp_r - w_i * wp_i + w_r;
            w_i = w_i * wp_r + w_temp * wp_i + w_i;
        }

        m_max = istep;
    }

    var tw_r = @sin(0.5 * ttheta);
    tw_r = -2.0 * tw_r * tw_r;
    const tw_i = @sin(ttheta);
    var twr: f64 = 1.0 + tw_r;
    var twi: f64 = tw_i;
    const n = length / 4 + 1;

    for (2..n + 1) |i| {
        const idx1 = i + i - 2;
        const idx2 = idx1 + 1;
        const idx3 = length + 1 - idx2;
        const idx4 = idx3 + 1;
        const wrs = twr;
        const wis = twi;
        const h1r = 0.5 * (array[idx1] + array[idx3]);
        const h1i = 0.5 * (array[idx2] - array[idx4]);
        const h2r = 0.5 * (array[idx2] + array[idx4]);
        const h2i = -0.5 * (array[idx1] - array[idx3]);
        array[idx1] = h1r + wrs * h2r - wis * h2i;
        array[idx2] = h1i + wrs * h2i + wis * h2r;
        array[idx3] = h1r - wrs * h2r + wis * h2i;
        array[idx4] = -h1i + wrs * h2i + wis * h2r;
        const tw_temp = twr;
        twr = twr * tw_r - twi * tw_i + twr;
        twi = twi * tw_r + tw_temp * tw_i + twi;
    }

    const saved = array[0];
    array[0] = saved + array[1];
    array[1] = saved - array[1];
}

// --- Tests ---

const testing = std.testing;

/// Identity filter for testing: returns the input sample unchanged.
/// Mirrors Go's testFrequencyResponseIdentytyFilter.
const TestIdentityFilter = struct {
    fn isPrimedFn(_: *anyopaque) bool {
        return true;
    }

    fn metadataFn(_: *anyopaque, out: *Metadata) void {
        out.* = .{
            .identifier = .simple_moving_average,
            .mnemonic = "identity",
            .description = "",
            .outputs_buf = undefined,
            .outputs_len = 0,
        };
    }

    fn updateScalarFn(_: *anyopaque, sample: *const Scalar) OutputArray {
        return OutputArray.fromScalar(.{ .time = sample.time, .value = sample.value });
    }

    fn updateBarFn(_: *anyopaque, _: *const @import("bar").Bar) OutputArray {
        return .{};
    }

    fn updateQuoteFn(_: *anyopaque, _: *const @import("quote").Quote) OutputArray {
        return .{};
    }

    fn updateTradeFn(_: *anyopaque, _: *const @import("trade").Trade) OutputArray {
        return .{};
    }

    const vtable = Indicator.VTable{
        .isPrimed = isPrimedFn,
        .metadata = metadataFn,
        .updateScalar = updateScalarFn,
        .updateBar = updateBarFn,
        .updateQuote = updateQuoteFn,
        .updateTrade = updateTradeFn,
    };

    fn indicator(self: *TestIdentityFilter) Indicator {
        return .{
            .ptr = @ptrCast(self),
            .vtable = &vtable,
        };
    }
};

fn check(exp: f64, act: f64) !void {
    if (@abs(exp - act) > math.floatMin(f64)) {
        std.debug.print("expected {d}, actual {d}\n", .{ exp, act });
        return error.TestExpectedEqual;
    }
}

test "validate signal length" {
    const max_length = 8199;

    for (0..max_length) |i| {
        const exp = switch (i) {
            4, 8, 16, 32, 64, 128, 256, 512, 1024, 2048, 4096, 8192 => true,
            else => false,
        };
        const act = isValidSignalLength(i);

        if (exp != act) {
            std.debug.print("isValidSignalLength({d}): expected {}, actual {}\n", .{ i, exp, act });
            return error.TestExpectedEqual;
        }
    }
}

test "prepare frequency domain" {
    const l: usize = 7;
    const fl: f64 = @floatFromInt(l);

    const expected = [_]f64{ 1 / fl, 2 / fl, 3 / fl, 4 / fl, 5 / fl, 6 / fl, 7 / fl };
    var actual = [_]f64{ 0, 0, 0, 0, 0, 0, 0 };

    prepareFrequencyDomain(l, &actual);

    for (0..l) |i| {
        try check(expected[i], actual[i]);
    }
}

test "prepare filtered signal" {
    const allocator = testing.allocator;
    const length: usize = 7;
    const warmup: usize = 5;

    const expected = [_]f64{ 1000, 0, 0, 0, 0, 0, 0 };
    var filter = TestIdentityFilter{};
    const ind = filter.indicator();

    const actual = try prepareFilteredSignal(allocator, length, ind, warmup);
    defer allocator.free(actual);

    for (0..length) |i| {
        try check(expected[i], actual[i]);
    }
}

test "calculate FFT" {
    var actual = [_]f64{ 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1 };
    const expected = [_]f64{ 16, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0 };

    directRealFastFourierTransform(&actual);

    for (0..expected.len) |i| {
        try check(expected[i], actual[i]);
    }
}

test "calculate" {
    const allocator = testing.allocator;
    const length: usize = 512;
    const warmup: usize = 128;

    var filter = TestIdentityFilter{};
    const ind = filter.indicator();

    var fr = try calculate(allocator, length, ind, warmup, 179);
    defer fr.deinit(allocator);

    try testing.expect(fr.normalized_frequency.len > 0);
}
