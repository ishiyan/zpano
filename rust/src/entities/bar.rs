/// Represents an OHLCV price bar.
#[derive(Debug, Clone, Copy)]
pub struct Bar {
    pub time: i64,
    pub open: f64,
    pub high: f64,
    pub low: f64,
    pub close: f64,
    pub volume: f64,
}

impl Bar {
    /// Creates a new Bar.
    pub fn new(time: i64, open: f64, high: f64, low: f64, close: f64, volume: f64) -> Self {
        Self { time, open, high, low, close, volume }
    }

    /// Indicates whether this is a rising bar (open < close).
    pub fn is_rising(&self) -> bool {
        self.open < self.close
    }

    /// Indicates whether this is a falling bar (close < open).
    pub fn is_falling(&self) -> bool {
        self.close < self.open
    }

    /// The median price: (low + high) / 2.
    pub fn median(&self) -> f64 {
        (self.low + self.high) / 2.0
    }

    /// The typical price: (low + high + close) / 3.
    pub fn typical(&self) -> f64 {
        (self.low + self.high + self.close) / 3.0
    }

    /// The weighted price: (low + high + 2*close) / 4.
    pub fn weighted(&self) -> f64 {
        (self.low + self.high + self.close + self.close) / 4.0
    }

    /// The average price: (low + high + open + close) / 4.
    pub fn average(&self) -> f64 {
        (self.low + self.high + self.open + self.close) / 4.0
    }
}

#[cfg(test)]
mod tests {
    use super::*;

    fn make_bar(o: f64, h: f64, l: f64, c: f64, v: f64) -> Bar {
        Bar::new(0, o, h, l, c, v)
    }

    #[test]
    fn test_median() {
        let b = make_bar(0.0, 3.0, 2.0, 0.0, 0.0);
        assert_eq!(b.median(), (b.low + b.high) / 2.0);
    }

    #[test]
    fn test_typical() {
        let b = make_bar(0.0, 4.0, 2.0, 3.0, 0.0);
        assert_eq!(b.typical(), (b.low + b.high + b.close) / 3.0);
    }

    #[test]
    fn test_weighted() {
        let b = make_bar(0.0, 4.0, 2.0, 3.0, 0.0);
        assert_eq!(b.weighted(), (b.low + b.high + b.close + b.close) / 4.0);
    }

    #[test]
    fn test_average() {
        let b = make_bar(3.0, 5.0, 2.0, 4.0, 0.0);
        assert_eq!(b.average(), (b.low + b.high + b.open + b.close) / 4.0);
    }

    #[test]
    fn test_is_rising() {
        assert!(make_bar(2.0, 0.0, 0.0, 3.0, 0.0).is_rising());
        assert!(!make_bar(3.0, 0.0, 0.0, 2.0, 0.0).is_rising());
        assert!(!make_bar(0.0, 0.0, 0.0, 0.0, 0.0).is_rising());
    }

    #[test]
    fn test_is_falling() {
        assert!(!make_bar(2.0, 0.0, 0.0, 3.0, 0.0).is_falling());
        assert!(make_bar(3.0, 0.0, 0.0, 2.0, 0.0).is_falling());
        assert!(!make_bar(0.0, 0.0, 0.0, 0.0, 0.0).is_falling());
    }
}
