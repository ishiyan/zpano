use super::shape::Shape;

/// Describes a single indicator output.
#[derive(Debug, Clone)]
pub struct OutputMetadata {
    /// Integer representation of the output enumeration of the related indicator.
    pub kind: i32,
    /// The data shape of this indicator output.
    pub shape: Shape,
    /// Short name (mnemonic) of this indicator output.
    pub mnemonic: String,
    /// Description of this indicator output.
    pub description: String,
}
