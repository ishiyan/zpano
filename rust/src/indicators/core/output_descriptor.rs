use super::outputs::shape::Shape;
use super::pane::Pane;
use super::role::Role;

/// Classifies a single indicator output for charting / discovery.
#[derive(Debug, Clone)]
pub struct OutputDescriptor {
    /// Integer representation of the output enumeration of the related indicator.
    pub kind: i32,
    /// The data shape of this output.
    pub shape: Shape,
    /// The semantic role of this output.
    pub role: Role,
    /// The chart pane on which this output is drawn.
    pub pane: Pane,
}
