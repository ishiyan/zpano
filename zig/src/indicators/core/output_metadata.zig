const shape = @import("outputs/shape.zig");

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
};
