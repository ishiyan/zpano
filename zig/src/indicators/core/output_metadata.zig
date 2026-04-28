const shape = @import("outputs/shape.zig");

/// Maximum buffer size for output mnemonic strings.
pub const max_mnemonic_len = 160;
/// Maximum buffer size for output description strings.
pub const max_description_len = 256;

/// Describes a single indicator output.
pub const OutputMetadata = struct {
    /// Integer representation of the output enumeration of the related indicator.
    kind: i32,
    /// The data shape of this indicator output.
    shape: shape.Shape,
    /// Short name (mnemonic) of this indicator output.
    mnemonic: []const u8,
    /// Description of this indicator output.
    description: []const u8,

    // Owned storage for mnemonic and description strings.
    mnemonic_storage: [max_mnemonic_len]u8 = undefined,
    description_storage: [max_description_len]u8 = undefined,

    /// Initialize with copies of the given strings. The mnemonic/description slices
    /// will point into the owned storage buffers.
    pub fn initOwned(kind_val: i32, shape_val: shape.Shape, mn: []const u8, desc: []const u8) OutputMetadata {
        var om: OutputMetadata = .{
            .kind = kind_val,
            .shape = shape_val,
            .mnemonic = undefined,
            .description = undefined,
        };
        const mn_len = @min(mn.len, max_mnemonic_len);
        @memcpy(om.mnemonic_storage[0..mn_len], mn[0..mn_len]);
        om.mnemonic = om.mnemonic_storage[0..mn_len];
        const desc_len = @min(desc.len, max_description_len);
        @memcpy(om.description_storage[0..desc_len], desc[0..desc_len]);
        om.description = om.description_storage[0..desc_len];
        return om;
    }

    /// Fix up slice pointers after a struct move/copy.
    pub fn fixSlices(self: *OutputMetadata) void {
        const mn_len = self.mnemonic.len;
        self.mnemonic = self.mnemonic_storage[0..mn_len];
        const desc_len = self.description.len;
        self.description = self.description_storage[0..desc_len];
    }
};
