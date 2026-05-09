const std = @import("std");
const math = std.math;

pub fn testPrices() [12]f64 {
    return .{ 1, 2, 8, 4, 9, 6, 7, 13, 9, 10, 3, 12 };
}

pub fn testVolumes() [12]f64 {
    return .{ 100, 90, 200, 150, 500, 100, 300, 150, 100, 300, 200, 100 };
}

pub fn testExpected() [12]f64 {
    return .{ 100, 190, 390, 240, 740, 640, 940, 1090, 990, 1290, 1090, 1190 };
}

