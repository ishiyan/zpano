const std = @import("std");
/// Non-overlapping partials retain cancellation precision. Storage belongs to allocator.
pub fn sum(allocator: std.mem.Allocator, values: []const f64) !f64 {
    const partials = try allocator.alloc(f64, values.len);
    defer allocator.free(partials);
    var length: usize = 0;
    for (values) |input| {
        var value = input;
        var i: usize = 0;
        for (partials[0..length]) |term| {
            var y = term;
            if (@abs(value) < @abs(y)) std.mem.swap(f64, &value, &y);
            const high = value + y;
            const low = y - (high - value);
            if (low != 0) {
                partials[i] = low;
                i += 1;
            }
            value = high;
        }
        partials[i] = value;
        length = i + 1;
    }
    var total: f64 = 0;
    while (length > 0) {
        length -= 1;
        total += partials[length];
    }
    return total;
}
