use super::identifier::Identifier;

/// Contains all info needed to create an indicator.
#[derive(Debug, Clone)]
pub struct Specification {
    /// Identifies the indicator to create.
    pub identifier: Identifier,
    /// Indicator-specific parameters (serialized as JSON or similar).
    pub parameters: Option<String>,
    /// Which outputs to compute (empty = all).
    pub outputs: Vec<i32>,
}

impl Specification {
    pub fn new(identifier: Identifier) -> Self {
        Self {
            identifier,
            parameters: None,
            outputs: Vec::new(),
        }
    }
}
