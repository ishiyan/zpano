use super::identifier::Identifier;
use super::outputs::metadata::OutputMetadata;

/// Describes an indicator and its outputs.
#[derive(Debug, Clone)]
pub struct Metadata {
    /// Identifies this indicator.
    pub identifier: Identifier,
    /// Short name (mnemonic) of this indicator.
    pub mnemonic: String,
    /// Description of this indicator.
    pub description: String,
    /// Metadata for individual outputs.
    pub outputs: Vec<OutputMetadata>,
}
