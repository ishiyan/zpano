//! Internal helper (not part of the public `core` API): a contiguous FIFO
//! buffer used in place of Python's `collections.deque`.
//!
//! Items live in an unmanaged `std.ArrayList(T)`; popping from the front
//! advances a head index and the storage is compacted when the dead prefix
//! is at least as long as the live part, so all operations are amortized
//! O(1) and the live items are always available as one contiguous slice.

const std = @import("std");

pub fn FifoBuffer(comptime T: type) type {
    return struct {
        list: std.ArrayList(T) = .empty,
        head: usize = 0,

        const Self = @This();

        /// Frees the storage.
        pub fn deinit(self: *Self, allocator: std.mem.Allocator) void {
            self.list.deinit(allocator);
            self.head = 0;
        }

        /// Removes all items, retaining capacity.
        pub fn clear(self: *Self) void {
            self.list.clearRetainingCapacity();
            self.head = 0;
        }

        /// Number of live items.
        pub fn len(self: *const Self) usize {
            return self.list.items.len - self.head;
        }

        /// Live items, oldest first. Valid until the next mutating call.
        pub fn slice(self: *const Self) []const T {
            return self.list.items[self.head..];
        }

        /// Pointer to the oldest item (buffer must be non-empty).
        pub fn first(self: *Self) *T {
            return &self.list.items[self.head];
        }

        /// Pointer to the newest item (buffer must be non-empty).
        pub fn last(self: *Self) *T {
            return &self.list.items[self.list.items.len - 1];
        }

        /// Makes room for `n` more appends without reallocation.
        pub fn ensureUnusedCapacity(self: *Self, allocator: std.mem.Allocator, n: usize) std.mem.Allocator.Error!void {
            try self.list.ensureUnusedCapacity(allocator, n);
        }

        /// Appends an item at the back.
        pub fn append(self: *Self, allocator: std.mem.Allocator, x: T) std.mem.Allocator.Error!void {
            try self.list.append(allocator, x);
        }

        /// Removes and returns the oldest item (buffer must be non-empty).
        pub fn popFront(self: *Self) T {
            const x = self.list.items[self.head];
            self.head += 1;
            if (self.head * 2 >= self.list.items.len) self.compact();
            return x;
        }

        fn compact(self: *Self) void {
            const live = self.list.items.len - self.head;
            if (live > 0) std.mem.copyForwards(T, self.list.items[0..live], self.list.items[self.head..]);
            self.list.shrinkRetainingCapacity(live);
            self.head = 0;
        }
    };
}

test "fifo buffer" {
    const a = std.testing.allocator;
    var b: FifoBuffer(f64) = .{};
    defer b.deinit(a);
    for (0..10) |i| try b.append(a, @floatFromInt(i));
    for (0..7) |i| try std.testing.expectEqual(@as(f64, @floatFromInt(i)), b.popFront());
    try std.testing.expectEqual(@as(usize, 3), b.len());
    try std.testing.expectEqualSlices(f64, &.{ 7, 8, 9 }, b.slice());
    try b.append(a, 10);
    try std.testing.expectEqual(@as(f64, 7), b.first().*);
    try std.testing.expectEqual(@as(f64, 10), b.last().*);
    b.clear();
    try std.testing.expectEqual(@as(usize, 0), b.len());
}
