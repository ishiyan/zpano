/// Represents a scalar value.
#[derive(Debug, Clone, Copy)]
pub struct Scalar {
    pub time: i64,
    pub value: f64,
}

impl Scalar {
    /// Creates a new Scalar.
    pub fn new(time: i64, value: f64) -> Self {
        Self { time, value }
    }
}

#[cfg(test)]
mod tests {
    use super::*;

    #[test]
    fn test_new() {
        let s = Scalar::new(100, 3.14);
        assert_eq!(s.time, 100);
        assert_eq!(s.value, 3.14);
    }
}
