const identifier_mod = @import("identifier.zig");
const adaptivity_mod = @import("adaptivity.zig");
const input_requirement_mod = @import("input_requirement.zig");
const volume_usage_mod = @import("volume_usage.zig");
const output_descriptor_mod = @import("output_descriptor.zig");
const descriptors_mod = @import("descriptors.zig");

/// Classifies an indicator along multiple taxonomic dimensions.
pub const Descriptor = struct {
    /// Uniquely identifies the indicator.
    identifier: identifier_mod.Identifier,
    /// Groups related indicators (e.g., by author or category).
    family: []const u8,
    /// Whether the indicator adapts its parameters.
    adaptivity: adaptivity_mod.Adaptivity,
    /// The minimum input data type this indicator consumes.
    input_requirement: input_requirement_mod.InputRequirement,
    /// How this indicator uses volume information.
    volume_usage: volume_usage_mod.VolumeUsage,
    /// Classification of each output.
    outputs: []const output_descriptor_mod.OutputDescriptor,
};

/// Returns the taxonomic descriptor for the given indicator identifier.
pub fn descriptorOf(id: identifier_mod.Identifier) ?*const Descriptor {
    for (&descriptors_mod.descriptors) |*d| {
        if (d.identifier == id) return d;
    }
    return null;
}

/// Returns a reference to the full descriptor registry.
pub fn allDescriptors() []const Descriptor {
    return &descriptors_mod.descriptors;
}
