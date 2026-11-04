const std = @import("std");
/// Temporary arrays and cloned state belong to the caller's allocator.
pub fn make(a: std.mem.Allocator, comptime T: type, n: usize) ![]T {
    const v = try a.alloc(T, n);
    for (v) |*x| {
        x.* = std.mem.zeroes(T);
    }
    return v;
}
pub fn dupe(a: std.mem.Allocator, comptime T: type, v: []const T) ![]T {
    return a.dupe(T, v);
}
pub fn create(a: std.mem.Allocator, value: anytype) !*@TypeOf(value) {
    const p = try a.create(@TypeOf(value));
    p.* = value;
    return p;
}
pub fn append(a: std.mem.Allocator, v: anytype, x: std.meta.Elem(@TypeOf(v))) !@TypeOf(v) {
    const r = try a.alloc(std.meta.Elem(@TypeOf(v)), v.len + 1);
    @memcpy(r[0..v.len], v);
    r[v.len] = x;
    return r;
}
pub fn concat(a: std.mem.Allocator, v: anytype, w: @TypeOf(v)) !@TypeOf(v) {
    const r = try a.alloc(std.meta.Elem(@TypeOf(v)), v.len + w.len);
    @memcpy(r[0..v.len], v);
    @memcpy(r[v.len..], w);
    return r;
}
pub fn copy(v: anytype, w: @TypeOf(v)) void {
    const n = @min(v.len, w.len);
    std.mem.copyForwards(std.meta.Elem(@TypeOf(v)), v[0..n], w[0..n]);
}
/// Recursively own every state array before a transactional update.
pub fn clone(a: std.mem.Allocator, value: anytype) !@TypeOf(value) {
    const T = @TypeOf(value);
    switch (@typeInfo(T)) {
        .optional => {
            if (value) |v| return try clone(a, v);
            return null;
        },
        .pointer => |p| {
            if (p.size == .slice) {
                const output = try a.alloc(p.child, value.len);
                for (value, 0..) |v, i| output[i] = try clone(a, v);
                return output;
            }
            if (p.size == .one) return try create(a, try clone(a, value.*));
            return value;
        },
        .@"struct" => |s| {
            var output = value;
            inline for (s.fields) |f| @field(output, f.name) = try clone(a, @field(value, f.name));
            return output;
        },
        .array => {
            var output = value;
            for (value, 0..) |v, i| output[i] = try clone(a, v);
            return output;
        },
        else => return value,
    }
}
