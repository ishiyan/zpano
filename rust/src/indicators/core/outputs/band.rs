/// Represents two band values and a time stamp.
#[derive(Debug, Clone, Copy)]
pub struct Band {
    pub time: i64,
    pub lower: f64,
    pub upper: f64,
}

impl Band {
    /// Creates a new band. Values are sorted so lower <= upper.
    pub fn new(time: i64, lower: f64, upper: f64) -> Self {
        if lower < upper {
            Self { time, lower, upper }
        } else {
            Self { time, lower: upper, upper: lower }
        }
    }

    /// Creates a new empty band with NaN values.
    pub fn empty(time: i64) -> Self {
        Self { time, lower: f64::NAN, upper: f64::NAN }
    }

    /// Indicates whether this band is not initialized.
    pub fn is_empty(&self) -> bool {
        self.lower.is_nan() || self.upper.is_nan()
    }
}
