/// Validate a finite parameter against its lower bound.
pub fn validate_real(value: f64, name: &str, minimum: f64) -> Option<String> {
    if value.is_finite() && value >= minimum {
        None
    } else {
        Some(format!("invalid {name}: finite >= {minimum} required"))
    }
}
pub fn validate_integer(value: i64, name: &str, minimum: i64) -> Option<String> {
    if value >= minimum {
        None
    } else {
        Some(format!("invalid {name}: integer >= {minimum} required"))
    }
}
pub fn validate_sample(value: f64) {
    assert!(!value.is_infinite(), "sample must be finite or NaN");
}
pub fn finite(value: f64) -> f64 {
    assert!(value.is_finite(), "non-finite numerical result");
    value
}
pub fn push(mut values: Vec<f64>, value: f64, capacity: i64) -> Vec<f64> {
    if values.len() == capacity as usize {
        values.remove(0);
    }
    values.push(value);
    values
}
