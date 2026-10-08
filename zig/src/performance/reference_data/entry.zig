//! Key/value entry type and lookup for the generated reference data tables.

const std = @import("std");

/// One key/value pair of a generated reference data table.
pub fn Entry(comptime K: type, comptime V: type) type {
    return struct { key: K, value: V };
}

/// Returns the value stored under `key`; panics if the key is absent.
pub fn lookup(comptime K: type, comptime V: type, entries: []const Entry(K, V), key: K) V {
    for (entries) |e| {
        const equal = if (K == []const u8) std.mem.eql(u8, e.key, key) else e.key == key;
        if (equal) return e.value;
    }
    @panic("reference data key not found");
}
