/// Represents a trade (time and sales) with price and volume.
#[derive(Debug, Clone, Copy)]
pub struct Trade {
    pub time: i64,
    pub price: f64,
    pub volume: f64,
}

impl Trade {
    /// Creates a new Trade.
    pub fn new(time: i64, price: f64, volume: f64) -> Self {
        Self { time, price, volume }
    }
}

#[cfg(test)]
mod tests {
    use super::*;

    #[test]
    fn test_new() {
        let t = Trade::new(100, 1.5, 2.5);
        assert_eq!(t.time, 100);
        assert_eq!(t.price, 1.5);
        assert_eq!(t.volume, 2.5);
    }
}
