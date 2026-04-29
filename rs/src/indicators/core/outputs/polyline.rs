/// A single vertex of a Polyline, expressed as (offset, value).
#[derive(Debug, Clone, Copy)]
pub struct Point {
    /// The number of bars back from the Polyline's time.
    pub offset: i32,
    /// The value (y) at this vertex.
    pub value: f64,
}

/// Holds a time stamp and an ordered, variable-length sequence of points.
#[derive(Debug, Clone)]
pub struct Polyline {
    pub time: i64,
    pub points: Vec<Point>,
}

impl Polyline {
    /// Creates a new polyline with the given time and points.
    pub fn new(time: i64, points: Vec<Point>) -> Self {
        Self { time, points }
    }

    /// Creates a new empty polyline with no points.
    pub fn empty(time: i64) -> Self {
        Self { time, points: Vec::new() }
    }

    /// Indicates whether this polyline has no points.
    pub fn is_empty(&self) -> bool {
        self.points.is_empty()
    }
}
