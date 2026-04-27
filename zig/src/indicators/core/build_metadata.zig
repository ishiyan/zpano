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
pub fn buildMetadata(
    identifier: identifier_mod.Identifier,
    mnemonic: []const u8,
    description: []const u8,
    texts: []const OutputText,
) metadata_mod.Metadata {
    const d = descriptor_mod.descriptorOf(identifier) orelse
        @panic("buildMetadata: no descriptor registered for identifier");

    if (texts.len != d.outputs.len) {
        @panic("buildMetadata: output text count mismatch");
    }

    var outputs: [max_outputs]output_metadata_mod.OutputMetadata = undefined;
    for (d.outputs, 0..) |od, i| {
        outputs[i] = .{
            .kind = od.kind,
            .shape = od.shape,
            .mnemonic = texts[i].mnemonic,
            .description = texts[i].description,
        };
    }

    return .{
        .identifier = identifier,
        .mnemonic = mnemonic,
        .description = description,
        .outputs = outputs[0..d.outputs.len],
    };
}
