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

const epsilon = 1e-12;

/// Enumerates the outputs of the Chande Momentum Oscillator.
pub const ChandeMomentumOscillatorOutput = enum(u8) {
    /// The scalar value of the CMO.
    value = 1,
};

/// Parameters to create an instance of the Chande Momentum Oscillator.
pub const ChandeMomentumOscillatorParams = struct {
    /// The length (number of time periods). Must be >= 1.
    length: usize,
    /// Bar component to extract. `null` means use default (Close).
    bar_component: ?bar_component.BarComponent = null,
    /// Quote component to extract. `null` means use default (Mid).
    quote_component: ?quote_component.QuoteComponent = null,
    /// Trade component to extract. `null` means use default (Price).
    trade_component: ?trade_component.TradeComponent = null,
};

/// Chande Momentum Oscillator (CMO).
///
/// CMOi = 100 * (SUi - SDi) / (SUi + SDi)
///
/// where SUi is the sum of gains and SDi is the sum of losses
/// over the chosen length.
///
/// The indicator is not primed during the first length updates.
pub const ChandeMomentumOscillator = struct {
    line: LineIndicator,
    ring_buffer: []f64,
    length: usize,
    ring_head: usize,
    count: usize,
    previous_sample: f64,
    gain_sum: f64,
    loss_sum: f64,
    primed: bool,
    allocator: std.mem.Allocator,
    mnemonic_buf: [64]u8,
    mnemonic_len: usize,
    description_buf: [128]u8,
    description_len: usize,

    pub fn init(allocator: std.mem.Allocator, params: ChandeMomentumOscillatorParams) !ChandeMomentumOscillator {
        if (params.length < 1) {
            return error.InvalidLength;
        }

        const bc = params.bar_component orelse bar_component.default_bar_component;
        const qc = params.quote_component orelse quote_component.default_quote_component;
        const tc = params.trade_component orelse trade_component.default_trade_component;

        var triple_buf: [64]u8 = undefined;
        const triple = component_triple_mnemonic_mod.componentTripleMnemonic(&triple_buf, bc, qc, tc);

        var mnemonic_buf: [64]u8 = undefined;
        const mnemonic_slice = std.fmt.bufPrint(&mnemonic_buf, "cmo({d}{s})", .{ params.length, triple }) catch
            return error.MnemonicTooLong;
        const mnemonic_len = mnemonic_slice.len;

        var description_buf: [128]u8 = undefined;
        const desc_slice = std.fmt.bufPrint(&description_buf, "Chande Momentum Oscillator {s}", .{mnemonic_slice}) catch
            return error.MnemonicTooLong;
        const description_len = desc_slice.len;

        const ring_buffer = try allocator.alloc(f64, params.length);
        @memset(ring_buffer, 0.0);

        return .{
            .line = LineIndicator.new(
                mnemonic_buf[0..mnemonic_len],
                description_buf[0..description_len],
                params.bar_component,
                params.quote_component,
                params.trade_component,
            ),
            .ring_buffer = ring_buffer,
            .length = params.length,
            .ring_head = 0,
            .count = 0,
            .previous_sample = 0.0,
            .gain_sum = 0.0,
            .loss_sum = 0.0,
            .primed = false,
            .allocator = allocator,
            .mnemonic_buf = mnemonic_buf,
            .mnemonic_len = mnemonic_len,
            .description_buf = description_buf,
            .description_len = description_len,
        };
    }

    pub fn deinit(self: *ChandeMomentumOscillator) void {
        self.allocator.free(self.ring_buffer);
    }

    pub fn fixSlices(self: *ChandeMomentumOscillator) void {
        self.line.mnemonic = self.mnemonic_buf[0..self.mnemonic_len];
        self.line.description = self.description_buf[0..self.description_len];
    }

    /// Core update logic.
    pub fn update(self: *ChandeMomentumOscillator, sample: f64) f64 {
        if (math.isNan(sample)) {
            return sample;
        }

        self.count += 1;
        if (self.count == 1) {
            self.previous_sample = sample;
            return math.nan(f64);
        }

        // New delta
        const delta = sample - self.previous_sample;
        self.previous_sample = sample;

        if (!self.primed) {
            // Fill until we have self.length deltas (i.e., self.length+1 samples)
            self.ring_buffer[self.ring_head] = delta;
            self.ring_head = (self.ring_head + 1) % self.length;

            if (delta > 0) {
                self.gain_sum += delta;
            } else if (delta < 0) {
                self.loss_sum += -delta;
            }

            if (self.count <= self.length) {
                return math.nan(f64);
            }

            // Now we have exactly self.length deltas in the buffer
            self.primed = true;
        } else {
            // Remove oldest delta and add the new one
            const old = self.ring_buffer[self.ring_head];
            if (old > 0) {
                self.gain_sum -= old;
            } else if (old < 0) {
                self.loss_sum -= -old;
            }

            self.ring_buffer[self.ring_head] = delta;
            self.ring_head = (self.ring_head + 1) % self.length;

            if (delta > 0) {
                self.gain_sum += delta;
            } else if (delta < 0) {
                self.loss_sum += -delta;
            }

            // Clamp to avoid tiny negative sums from FP noise
            if (self.gain_sum < 0) {
                self.gain_sum = 0;
            }

            if (self.loss_sum < 0) {
                self.loss_sum = 0;
            }
        }

        const den = self.gain_sum + self.loss_sum;
        if (@abs(den) < epsilon) {
            return 0;
        }

        return 100.0 * (self.gain_sum - self.loss_sum) / den;
    }

    pub fn isPrimed(self: *const ChandeMomentumOscillator) bool {
        return self.primed;
    }

    pub fn getMetadata(self: *const ChandeMomentumOscillator, out: *Metadata) void {
        build_metadata_mod.buildMetadata(
            out,
            .chande_momentum_oscillator,
            self.line.mnemonic,
            self.line.description,
            &[_]build_metadata_mod.OutputText{
                .{ .mnemonic = self.line.mnemonic, .description = self.line.description },
            },
        );
    }

    pub fn updateScalar(self: *ChandeMomentumOscillator, sample: *const Scalar) OutputArray {
        const value = self.update(sample.value);
        return LineIndicator.wrapScalar(sample.time, value);
    }

    pub fn updateBar(self: *ChandeMomentumOscillator, sample: *const Bar) OutputArray {
        const value = self.update(self.line.extractBar(sample));
        return LineIndicator.wrapScalar(sample.time, value);
    }

    pub fn updateQuote(self: *ChandeMomentumOscillator, sample: *const Quote) OutputArray {
        const value = self.update(self.line.extractQuote(sample));
        return LineIndicator.wrapScalar(sample.time, value);
    }

    pub fn updateTrade(self: *ChandeMomentumOscillator, sample: *const Trade) OutputArray {
        const value = self.update(self.line.extractTrade(sample));
        return LineIndicator.wrapScalar(sample.time, value);
    }

    pub fn indicator(self: *ChandeMomentumOscillator) indicator_mod.Indicator {
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
        const self: *ChandeMomentumOscillator = @ptrCast(@alignCast(ptr));
        return self.isPrimed();
    }

    fn vtableMetadata(ptr: *anyopaque, out: *Metadata) void {
        const self: *const ChandeMomentumOscillator = @ptrCast(@alignCast(ptr));
        self.getMetadata(out);
    }

    fn vtableUpdateScalar(ptr: *anyopaque, sample: *const Scalar) OutputArray {
        const self: *ChandeMomentumOscillator = @ptrCast(@alignCast(ptr));
        return self.updateScalar(sample);
    }

    fn vtableUpdateBar(ptr: *anyopaque, sample: *const Bar) OutputArray {
        const self: *ChandeMomentumOscillator = @ptrCast(@alignCast(ptr));
        return self.updateBar(sample);
    }

    fn vtableUpdateQuote(ptr: *anyopaque, sample: *const Quote) OutputArray {
        const self: *ChandeMomentumOscillator = @ptrCast(@alignCast(ptr));
        return self.updateQuote(sample);
    }

    fn vtableUpdateTrade(ptr: *anyopaque, sample: *const Trade) OutputArray {
        const self: *ChandeMomentumOscillator = @ptrCast(@alignCast(ptr));
        return self.updateTrade(sample);
    }

    pub const InitError = error{
        InvalidLength,
        MnemonicTooLong,
        OutOfMemory,
    };
};

// ---------------------------------------------------------------------------
// Tests
// ---------------------------------------------------------------------------

const testing = std.testing;

fn testBookLength10Input() [13]f64 {
    return .{
        101.0313, 101.0313, 101.1250, 101.9687, 102.7813,
        103.0000, 102.9687, 103.0625, 102.9375, 102.7188,
        102.7500, 102.9063, 102.9687,
    };
}

fn testBookLength10Output() [13]f64 {
    return .{
        math.nan(f64),     math.nan(f64),     math.nan(f64),     math.nan(f64), math.nan(f64),
        math.nan(f64),     math.nan(f64),     math.nan(f64),     math.nan(f64), math.nan(f64),
        69.61963786608334, 71.42857142857143, 71.08377992828775,
    };
}

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

fn createCmo(allocator: std.mem.Allocator, length: usize) !ChandeMomentumOscillator {
    var cmo = try ChandeMomentumOscillator.init(allocator, .{ .length = length });
    cmo.fixSlices();
    return cmo;
}

test "cmo update length 10 book" {
    const input = testBookLength10Input();
    const output = testBookLength10Output();
    var cmo = try createCmo(testing.allocator, 10);
    defer cmo.deinit();

    // First 10 updates produce NaN.
    for (0..10) |i| {
        const act = cmo.update(input[i]);
        try testing.expect(math.isNan(act));
    }

    // From index 10 onward, primed.
    for (10..13) |i| {
        const act = cmo.update(input[i]);
        try testing.expect(@abs(act - output[i]) < 1e-13);
    }

    // NaN passthrough
    try testing.expect(math.isNan(cmo.update(math.nan(f64))));
}

test "cmo is primed" {
    const input = testInput();

    inline for ([_]usize{ 1, 2, 3, 5, 10 }) |length| {
        var cmo = try createCmo(testing.allocator, length);
        defer cmo.deinit();

        try testing.expect(!cmo.isPrimed());

        for (0..length) |i| {
            _ = cmo.update(input[i]);
            try testing.expect(!cmo.isPrimed());
        }

        for (length..252) |i| {
            _ = cmo.update(input[i]);
            try testing.expect(cmo.isPrimed());
        }
    }
}

test "cmo metadata" {
    var cmo = try createCmo(testing.allocator, 5);
    defer cmo.deinit();
    var m: Metadata = undefined;
    cmo.getMetadata(&m);

    try testing.expectEqual(Identifier.chande_momentum_oscillator, m.identifier);
    try testing.expectEqual(@as(usize, 1), m.outputs_len);
    try testing.expectEqual(@as(i32, 1), m.outputs_buf[0].kind);
    try testing.expectEqualStrings("cmo(5)", m.outputs_buf[0].mnemonic);
    try testing.expectEqualStrings("Chande Momentum Oscillator cmo(5)", m.outputs_buf[0].description);
}

test "cmo update entity" {
    const length: usize = 2;
    const inp: f64 = 3.0;
    const exp: f64 = 100.0; // all gains, no losses
    const time: i64 = 1617235200;

    // scalar
    {
        var cmo = try createCmo(testing.allocator, length);
        defer cmo.deinit();
        _ = cmo.update(0.0);
        _ = cmo.update(0.0);
        const out = cmo.updateScalar(&.{ .time = time, .value = inp });
        try testing.expectEqual(@as(usize, 1), out.len);
        const s = out.slice()[0].scalar;
        try testing.expectEqual(time, s.time);
        try testing.expect(@abs(s.value - exp) < 1e-13);
    }

    // bar
    {
        var cmo = try createCmo(testing.allocator, length);
        defer cmo.deinit();
        _ = cmo.update(0.0);
        _ = cmo.update(0.0);
        const bar = Bar{ .time = time, .open = 0, .high = 0, .low = 0, .close = inp, .volume = 0 };
        const out = cmo.updateBar(&bar);
        const s = out.slice()[0].scalar;
        try testing.expect(@abs(s.value - exp) < 1e-13);
    }

    // quote
    {
        var cmo = try createCmo(testing.allocator, length);
        defer cmo.deinit();
        _ = cmo.update(0.0);
        _ = cmo.update(0.0);
        const quote = Quote{ .time = time, .bid_price = inp, .ask_price = inp, .bid_size = 0, .ask_size = 0 };
        const out = cmo.updateQuote(&quote);
        const s = out.slice()[0].scalar;
        try testing.expect(@abs(s.value - exp) < 1e-13);
    }

    // trade
    {
        var cmo = try createCmo(testing.allocator, length);
        defer cmo.deinit();
        _ = cmo.update(0.0);
        _ = cmo.update(0.0);
        const trade = Trade{ .time = time, .price = inp, .volume = 0 };
        const out = cmo.updateTrade(&trade);
        const s = out.slice()[0].scalar;
        try testing.expect(@abs(s.value - exp) < 1e-13);
    }
}

test "cmo init invalid length" {
    const result = ChandeMomentumOscillator.init(testing.allocator, .{ .length = 0 });
    try testing.expectError(error.InvalidLength, result);
}

test "cmo mnemonic components" {
    // all defaults -> no component suffix
    {
        var cmo = try createCmo(testing.allocator, 5);
        defer cmo.deinit();
        try testing.expectEqualStrings("cmo(5)", cmo.line.mnemonic);
    }

    // bar component set to Median
    {
        var cmo = try ChandeMomentumOscillator.init(testing.allocator, .{
            .length = 5,
            .bar_component = .median,
        });
        defer cmo.deinit();
        cmo.fixSlices();
        try testing.expectEqualStrings("cmo(5, hl/2)", cmo.line.mnemonic);
        try testing.expectEqualStrings("Chande Momentum Oscillator cmo(5, hl/2)", cmo.line.description);
    }

    // bar=high, trade=volume
    {
        var cmo = try ChandeMomentumOscillator.init(testing.allocator, .{
            .length = 5,
            .bar_component = .high,
            .trade_component = .volume,
        });
        defer cmo.deinit();
        cmo.fixSlices();
        try testing.expectEqualStrings("cmo(5, h, v)", cmo.line.mnemonic);
    }
}
