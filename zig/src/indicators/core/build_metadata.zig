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
        const mn = texts[i].mnemonic;
        const desc = texts[i].description;
        const mn_len = @min(mn.len, output_metadata_mod.max_mnemonic_len);
        const desc_len = @min(desc.len, output_metadata_mod.max_description_len);
        // Copy strings into owned storage within the output metadata struct.
        @memcpy(out.outputs_buf[i].mnemonic_storage[0..mn_len], mn[0..mn_len]);
        @memcpy(out.outputs_buf[i].description_storage[0..desc_len], desc[0..desc_len]);
        out.outputs_buf[i].kind = od.kind;
        out.outputs_buf[i].shape = od.shape;
        // Point slices to the owned storage (which is now at its final location in `out`).
        out.outputs_buf[i].mnemonic = out.outputs_buf[i].mnemonic_storage[0..mn_len];
        out.outputs_buf[i].description = out.outputs_buf[i].description_storage[0..desc_len];
    }
}
