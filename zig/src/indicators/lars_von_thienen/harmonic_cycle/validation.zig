const std = @import("std");
const memory = @import("memory.zig");
pub fn real(_: std.mem.Allocator, value: f64, _: []const u8, minimum: f64) !void {
    if (!std.math.isFinite(value) or value < minimum) return error.InvalidParameter;
}
pub fn integer(_: std.mem.Allocator, value: isize, _: []const u8, minimum: isize) !void {
    if (value < minimum) return error.InvalidParameter;
}
pub fn finite(_: std.mem.Allocator, value: f64) !f64 {
    if (!std.math.isFinite(value)) return error.NonFiniteResult;
    return value;
}
pub fn sample(_: std.mem.Allocator, value: f64) !void {
    if (std.math.isInf(value)) return error.InvalidSample;
}
pub fn push(a: std.mem.Allocator, v: []f64, x: f64, capacity: isize) ![]f64 {
    if (v.len == @as(usize, @intCast(capacity))) {
        memory.copy(v, v[1..]);
        v[v.len - 1] = x;
        return v;
    }
    return memory.append(a, v, x);
}
