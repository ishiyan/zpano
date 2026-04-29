use super::identifier::Identifier;
use super::adaptivity::Adaptivity;
use super::input_requirement::InputRequirement;
use super::volume_usage::VolumeUsage;
use super::output_descriptor::OutputDescriptor;

/// Classifies an indicator along multiple taxonomic dimensions.
#[derive(Debug, Clone)]
pub struct Descriptor {
    /// Uniquely identifies the indicator.
    pub identifier: Identifier,
    /// Groups related indicators (e.g., by author or category).
    pub family: &'static str,
    /// Whether the indicator adapts its parameters.
    pub adaptivity: Adaptivity,
    /// The minimum input data type this indicator consumes.
    pub input_requirement: InputRequirement,
    /// How this indicator uses volume information.
    pub volume_usage: VolumeUsage,
    /// Classification of each output.
    pub outputs: &'static [OutputDescriptor],
}

/// Returns the taxonomic descriptor for the given indicator identifier.
pub fn descriptor_of(id: Identifier) -> Option<&'static Descriptor> {
    super::descriptors::DESCRIPTORS.iter().find(|d| d.identifier == id)
}

/// Returns a reference to the full descriptor registry.
pub fn descriptors() -> &'static [Descriptor] {
    super::descriptors::DESCRIPTORS
}
