const identifier_mod = @import("identifier.zig");
const output_metadata_mod = @import("output_metadata.zig");

pub const max_outputs = 9;

/// Describes an indicator and its outputs.
pub const Metadata = struct {
    /// Identifies this indicator.
    identifier: identifier_mod.Identifier,
    /// Short name (mnemonic) of this indicator.
    mnemonic: []const u8,
    /// Description of this indicator.
    description: []const u8,
    /// Metadata for individual outputs (fixed-size buffer, use outputs_len for valid count).
    outputs_buf: [max_outputs]output_metadata_mod.OutputMetadata,
    /// Number of valid entries in outputs_buf.
    outputs_len: usize,
};
