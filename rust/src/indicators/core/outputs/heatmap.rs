/// Holds a time stamp (x) and an array of values (z) corresponding to parameter (y) range
/// to paint a heatmap column.
#[derive(Debug, Clone)]
pub struct Heatmap {
    pub time: i64,
    pub parameter_first: f64,
    pub parameter_last: f64,
    pub parameter_resolution: f64,
    pub value_min: f64,
    pub value_max: f64,
    pub values: Vec<f64>,
}

impl Heatmap {
    /// Creates a new heatmap.
    pub fn new(
        time: i64,
        parameter_first: f64,
        parameter_last: f64,
        parameter_resolution: f64,
        value_min: f64,
        value_max: f64,
        values: Vec<f64>,
    ) -> Self {
        Self { time, parameter_first, parameter_last, parameter_resolution, value_min, value_max, values }
    }

    /// Creates a new empty heatmap with NaN min/max and empty values.
    pub fn empty(time: i64, parameter_first: f64, parameter_last: f64, parameter_resolution: f64) -> Self {
        Self {
            time,
            parameter_first,
            parameter_last,
            parameter_resolution,
            value_min: f64::NAN,
            value_max: f64::NAN,
            values: Vec::new(),
        }
    }

    /// Indicates whether this heatmap is not initialized.
    pub fn is_empty(&self) -> bool {
        self.values.is_empty()
    }
}
