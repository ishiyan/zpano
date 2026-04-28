const Bar = @import("bar").Bar;
const Quote = @import("quote").Quote;
const Trade = @import("trade").Trade;
const Scalar = @import("scalar").Scalar;
const metadata_mod = @import("metadata.zig");
const band_mod = @import("outputs/band.zig");
const heatmap_mod = @import("outputs/heatmap.zig");
const polyline_mod = @import("outputs/polyline.zig");

/// A single output value from an indicator update.
pub const OutputValue = union(enum) {
    scalar: Scalar,
    band: band_mod.Band,
    heatmap: heatmap_mod.Heatmap,
    polyline: polyline_mod.Polyline,
};

/// Maximum number of outputs per indicator update.
pub const max_output_count = 9;

/// Fixed-capacity array of output values returned by indicator updates.
pub const OutputArray = struct {
    values: [max_output_count]OutputValue = undefined,
    len: usize = 0,

    pub fn append(self: *OutputArray, value: OutputValue) void {
        self.values[self.len] = value;
        self.len += 1;
    }

    pub fn slice(self: *const OutputArray) []const OutputValue {
        return self.values[0..self.len];
    }

    /// Creates an OutputArray with a single scalar output.
    pub fn fromScalar(s: Scalar) OutputArray {
        var arr = OutputArray{};
        arr.values[0] = .{ .scalar = s };
        arr.len = 1;
        return arr;
    }
};

/// Common indicator interface, implemented via function pointers.
pub const Indicator = struct {
    ptr: *anyopaque,
    vtable: *const VTable,

    pub const VTable = struct {
        isPrimed: *const fn (ptr: *anyopaque) bool,
        metadata: *const fn (ptr: *anyopaque, out: *metadata_mod.Metadata) void,
        updateScalar: *const fn (ptr: *anyopaque, sample: *const Scalar) OutputArray,
        updateBar: *const fn (ptr: *anyopaque, sample: *const Bar) OutputArray,
        updateQuote: *const fn (ptr: *anyopaque, sample: *const Quote) OutputArray,
        updateTrade: *const fn (ptr: *anyopaque, sample: *const Trade) OutputArray,
    };

    pub fn isPrimed(self: Indicator) bool {
        return self.vtable.isPrimed(self.ptr);
    }

    pub fn metadata(self: Indicator, out: *metadata_mod.Metadata) void {
        self.vtable.metadata(self.ptr, out);
    }

    pub fn updateScalar(self: Indicator, sample: *const Scalar) OutputArray {
        return self.vtable.updateScalar(self.ptr, sample);
    }

    pub fn updateBar(self: Indicator, sample: *const Bar) OutputArray {
        return self.vtable.updateBar(self.ptr, sample);
    }

    pub fn updateQuote(self: Indicator, sample: *const Quote) OutputArray {
        return self.vtable.updateQuote(self.ptr, sample);
    }

    pub fn updateTrade(self: Indicator, sample: *const Trade) OutputArray {
        return self.vtable.updateTrade(self.ptr, sample);
    }
};
