#[cfg(test)]
pub mod testdata {

    pub fn test_input() -> Vec<f64> {
        vec![1.0, 2.0, 8.0, 4.0, 9.0, 6.0, 7.0, 13.0, 9.0, 10.0, 3.0, 12.0]
    }

    /// Population stdev output for length 3 (computed from Go reference implementation).
    pub fn expected_len3_population() -> Vec<f64> {
        vec![
            f64::NAN, f64::NAN,
            3.091206165165235, 2.494438257849293, 2.160246899469286, 2.054804667656327, 1.247219128924651,
            3.091206165165237, 2.494438257849299, 1.699673171197598, 3.091206165165236, 3.858612300930073,
        ]
    }

    /// Sample stdev output for length 3 (computed from Go reference implementation).
    pub fn expected_len3_sample() -> Vec<f64> {
        vec![
            f64::NAN, f64::NAN,
            3.785938897200182, 3.055050463303894, 2.645751311064591, 2.516611478423584, 1.527525231651945,
            3.785938897200182, 3.055050463303895, 2.081665999466135, 3.785938897200182, 4.725815626252608,
        ]
    }
}
