//! Percentile with NumPy's `method="linear"` definition.

const std = @import("std");
const math = std.math;
const testing = std.testing;

/// Errors returned by `percentile`.
pub const Error = error{InvalidArgument} || std.mem.Allocator.Error;

/// Compute the q-th percentile using NumPy's ``method="linear"``
/// definition.
///
/// `window` is not modified: a sorted copy is made with `allocator`
/// and freed before returning.
///
/// Args:
///     window: Input data.
///     q: Percentile in the range [0, 1].
///
/// Returns the interpolated percentile value, `error.InvalidArgument` if q
/// is outside [0, 1] (or NaN) or `window` is empty (Python ValueError), or
/// `error.OutOfMemory`.
pub fn percentile(allocator: std.mem.Allocator, window: []const f64, q: f64) Error!f64 {
    if (!(0 <= q and q <= 1)) return error.InvalidArgument;
    const values = try allocator.dupe(f64, window);
    defer allocator.free(values);
    return percentileInPlace(values, q);
}

/// Same as `percentile`, but sorts `values` in place instead of copying
/// (no allocation). On return `values` is sorted ascending.
pub fn percentileInPlace(values: []f64, q: f64) error{InvalidArgument}!f64 {
    if (!(0 <= q and q <= 1)) return error.InvalidArgument;
    std.mem.sort(f64, values, {}, std.sort.asc(f64));
    if (values.len == 0) return error.InvalidArgument;
    const n = values.len;

    if (n == 1) return values[0];

    const idx = q * @as(f64, @floatFromInt(n - 1));
    const lo: usize = @intFromFloat(idx);

    if (lo >= n - 1) return values[n - 1];

    const hi = lo + 1;
    const frac = idx - @as(f64, @floatFromInt(lo));

    return values[lo] + frac * (values[hi] - values[lo]);
}

// ── Tests ──────────────────────────────────────────────────────────────────

fn expectAlmostEqual(expected: f64, actual: f64, places: u5) !void {
    if (expected == actual) return;
    const tol = 0.5 * math.pow(f64, 10.0, -@as(f64, @floatFromInt(places)));
    if (!(@abs(expected - actual) <= tol)) {
        std.debug.print("expected {d}, got {d} (places {d})\n", .{ expected, actual, places });
        return error.TestExpectedApproxEq;
    }
}

fn checkWindow(window: []const f64, expected: []const f64) !void {
    const alloc = testing.allocator;
    const q_values = [_]f64{ 0, 0.01, 0.05, 0.1, 0.2, 0.25, 0.3, 0.4, 0.5, 0.6, 0.7, 0.75, 0.8, 0.9, 0.95, 0.99, 1 };

    for (q_values, expected) |q, e| {
        try expectAlmostEqual(e, try percentile(alloc, window, q), 14);
    }

    // Property test: monotonicity
    var last = -math.inf(f64);
    for (0..101) |qi| {
        const q = @as(f64, @floatFromInt(qi)) * 0.01;
        const current = try percentile(alloc, window, q);
        try testing.expect(current >= last);
        last = current;
    }

    // Property test: affine invariance
    // P_q(aX+b) = aP_q(X)+b, a > 0
    const scaled = try alloc.alloc(f64, window.len);
    defer alloc.free(scaled);
    for (window, scaled) |x, *s| s.* = 3 * x + 7;
    for (0..101) |qi| {
        const q = @as(f64, @floatFromInt(qi)) * 0.01;
        const act = 3 * (try percentile(alloc, window, q)) + 7;
        const exp = try percentile(alloc, scaled, q);
        try expectAlmostEqual(exp, act, 13);
    }
}

test "reference dataset" {
    const win = [_]f64{
        -5.453279550656607,  -3.6648332058049427, 5.947309146654682,    3.525093415019491,
        -2.1778089879618197, -3.34372144267231,   1.9661750717437965,   -6.265316287925733,
        3.4551208802924265,  8.836057305398743,   -5.03508570740858,    8.977623036666365,
        3.3447490620074483,  -8.082041288117757,  -1.1632066766437443,  7.729598386550354,
        3.949069997640443,   -3.4705427185977573, 4.67856326660133,     -5.597300889090276,
        -8.368108609155838,  -6.802087978499049,  -3.197996300905894,   -0.6961369259589816,
        -4.671579434184581,  6.315528068496139,   -6.13411221421011,    -7.410618476455994,
        -8.166704969101282,  1.971360273298263,   7.094838087480028,    2.0324248338742628,
        8.63976722271967,    4.495627221840401,   7.211026347865847,    8.586756031506326,
        0.9237201816470613,  8.75345917535514,    -0.10024119842351453, -4.5245363502002505,
        -0.9644258505047869, 3.3007784679906056,
    };
    const exp = [_]f64{
        -8.368108609155838,  -8.28553311673347,  -8.048470147534669,  -6.748410809441717,
        -5.369640782007001,  -4.634818663188499, -3.6065460596427874, -1.7719680634345873,
        0.41173949161177337, 2.793437014344066,  3.821877022854157,   4.632829255411098,
        6.241884284127849,   8.501040267010728,  8.747774577723366,   8.91958108684664,
        8.977623036666365,
    };
    try checkWindow(&win, &exp);
}

test "algorithm" {
    // one element
    try checkWindow(&.{42}, &.{ 42, 42, 42, 42, 42, 42, 42, 42, 42, 42, 42, 42, 42, 42, 42, 42, 42 });

    // two elements
    try checkWindow(&.{ 10, 20 }, &.{ 10.0, 10.1, 10.5, 11.0, 12.0, 12.5, 13, 14, 15, 16, 17, 17.5, 18, 19, 19.5, 19.9, 20 });

    const odd = [_]f64{ 1.0, 1.04, 1.2, 1.4, 1.8, 2.0, 2.2, 2.6, 3.0, 3.4, 3.8, 4.0, 4.2, 4.6, 4.8, 4.96, 5.0 };
    try checkWindow(&.{ 1, 2, 3, 4, 5 }, &odd); // sorted odd number of elements
    try checkWindow(&.{ 5, 2, 1, 4, 3 }, &odd); // unsorted odd number of elements
    try checkWindow(&.{ 5, 4, 3, 2, 1 }, &odd); // reverse sorted odd number of elements

    const even = [_]f64{ 1.0, 1.03, 1.15, 1.3, 1.6, 1.75, 1.9, 2.2, 2.5, 2.8, 3.1, 3.25, 3.4, 3.7, 3.85, 3.97, 4.0 };
    try checkWindow(&.{ 1, 2, 3, 4 }, &even); // sorted even number of elements
    try checkWindow(&.{ 3, 2, 4, 1 }, &even); // unsorted even number of elements
    try checkWindow(&.{ 4, 3, 2, 1 }, &even); // reverse sorted even number of elements

    // duplicate elements
    try checkWindow(&.{ 1, 1, 2, 2, 3 }, &.{ 1.0, 1.0, 1.0, 1.0, 1.0, 1.0, 1.2, 1.6, 2.0, 2.0, 2.0, 2.0, 2.2, 2.6, 2.8, 2.96, 3.0 });

    // more duplicate elements
    try checkWindow(&.{ 1, 2, 2, 2, 5 }, &.{ 1.0, 1.04, 1.2, 1.4, 1.8, 2.0, 2.0, 2.0, 2.0, 2.0, 2.0, 2.0, 2.6, 3.80, 4.4, 4.88, 5.0 });

    // equal elements
    try checkWindow(&.{ 2, 2, 2, 2, 2 }, &.{ 2.0, 2.0, 2.0, 2.0, 2.0, 2.0, 2.0, 2.0, 2.0, 2.0, 2.0, 2.0, 2.0, 2.0, 2.0, 2.0, 2.0 });

    // negative elements
    try checkWindow(&.{ -10, -5, 0, 5, 10 }, &.{ -10.0, -9.8, -9.0, -8.0, -6.0, -5.0, -4.0, -2.0, 0.0, 2.0, 4.0, 5.0, 6.0, 8.0, 9.0, 9.8, 10.0 });

    // floating-point elements
    try checkWindow(&.{ 3.14159, -2.71828, 0.57721, 1.41421, -0.69315 }, &.{
        -2.71828, -2.6372748, -2.313254, -1.908228, -1.098176, -0.69315, -0.439078, 0.069066,
        0.57721,  0.91201,    1.24681,   1.41421,   1.759686,  2.450638, 2.796114,  3.0724948,
        3.14159,
    });

    // Q out of range
    try testing.expectError(error.InvalidArgument, percentile(testing.allocator, &.{42}, -0.01));
    try testing.expectError(error.InvalidArgument, percentile(testing.allocator, &.{42}, 1.01));

    // Empty window (Python: q=42 is rejected first; both are ValueError)
    try testing.expectError(error.InvalidArgument, percentile(testing.allocator, &.{}, 42));
    try testing.expectError(error.InvalidArgument, percentile(testing.allocator, &.{}, 0.5));
}
