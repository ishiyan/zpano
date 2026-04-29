/// Identifies the data shape of an indicator output.
#[derive(Debug, Clone, Copy, PartialEq, Eq)]
#[repr(u8)]
pub enum Shape {
    /// Holds a time stamp and a value.
    Scalar = 1,
    /// Holds a time stamp and two values representing upper and lower lines of a band.
    Band = 2,
    /// Holds a time stamp and an array of values representing a heat-map column.
    Heatmap = 3,
    /// Holds a time stamp and an ordered, variable-length sequence of (offset, value) points.
    Polyline = 4,
}

impl Shape {
    /// Returns the string representation.
    pub fn as_str(&self) -> &'static str {
        match self {
            Shape::Scalar => "scalar",
            Shape::Band => "band",
            Shape::Heatmap => "heatmap",
            Shape::Polyline => "polyline",
        }
    }
}

impl std::fmt::Display for Shape {
    fn fmt(&self, f: &mut std::fmt::Formatter<'_>) -> std::fmt::Result {
        f.write_str(self.as_str())
    }
}
