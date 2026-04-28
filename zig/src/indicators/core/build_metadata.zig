const identifier_mod = @import("identifier.zig");
const descriptor_mod = @import("descriptor.zig");
const metadata_mod = @import("metadata.zig");
const output_metadata_mod = @import("output_metadata.zig");

/// Per-output text supplied by the indicator implementation.
pub const OutputText = struct {
    mnemonic: []const u8,
    description: []const u8,
};

/// Maximum number of outputs per indicator.
pub const max_outputs = 9;

/// Constructs a `Metadata` by joining the descriptor registry's per-output
/// kind and shape with the supplied per-output mnemonic and description.
/// Uses an out-pointer to avoid returning a large struct by value (which
/// triggers codegen issues in Zig 0.16-dev).
pub fn buildMetadata(
    out: *metadata_mod.Metadata,
    identifier: identifier_mod.Identifier,
    mnemonic: []const u8,
    description: []const u8,
    texts: []const OutputText,
) void {
    const d = descriptor_mod.descriptorOf(identifier) orelse
        @panic("buildMetadata: no descriptor registered for identifier");

    if (texts.len != d.outputs.len) {
        @panic("buildMetadata: output text count mismatch");
    }

    out.identifier = identifier;
    out.mnemonic = mnemonic;
    out.description = description;
    out.outputs_len = d.outputs.len;
    for (d.outputs, 0..) |od, i| {
        out.outputs_buf[i] = .{
            .kind = od.kind,
            .shape = od.shape,
            .mnemonic = texts[i].mnemonic,
            .description = texts[i].description,
        };
    }
}
