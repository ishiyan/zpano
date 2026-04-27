const identifier_mod = @import("identifier.zig");
const output_metadata_mod = @import("output_metadata.zig");

/// Describes an indicator and its outputs.
pub const Metadata = struct {
    /// Identifies this indicator.
    identifier: identifier_mod.Identifier,
    /// Short name (mnemonic) of this indicator.
    mnemonic: []const u8,
    /// Description of this indicator.
    description: []const u8,
    /// Metadata for individual outputs.
    outputs: []const output_metadata_mod.OutputMetadata,
};
