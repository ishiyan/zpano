/// Represents a price quote (bid/ask price and size pair).
#[derive(Debug, Clone, Copy)]
pub struct Quote {
    pub time: i64,
    pub bid_price: f64,
    pub ask_price: f64,
    pub bid_size: f64,
    pub ask_size: f64,
}

impl Quote {
    /// Creates a new Quote.
    pub fn new(time: i64, bid_price: f64, ask_price: f64, bid_size: f64, ask_size: f64) -> Self {
        Self { time, bid_price, ask_price, bid_size, ask_size }
    }

    /// The mid-price: (ask_price + bid_price) / 2.
    pub fn mid(&self) -> f64 {
        (self.ask_price + self.bid_price) / 2.0
    }

    /// The weighted price: (ask*askSize + bid*bidSize) / (askSize + bidSize).
    pub fn weighted(&self) -> f64 {
        let size = self.ask_size + self.bid_size;
        if size == 0.0 { return 0.0; }
        (self.ask_price * self.ask_size + self.bid_price * self.bid_size) / size
    }

    /// The weighted mid-price (micro-price): (ask*bidSize + bid*askSize) / (askSize + bidSize).
    pub fn weighted_mid(&self) -> f64 {
        let size = self.ask_size + self.bid_size;
        if size == 0.0 { return 0.0; }
        (self.ask_price * self.bid_size + self.bid_price * self.ask_size) / size
    }

    /// The spread in basis points: 20000 * (ask - bid) / (ask + bid).
    pub fn spread_bp(&self) -> f64 {
        let mid = self.ask_price + self.bid_price;
        if mid == 0.0 { return 0.0; }
        20000.0 * (self.ask_price - self.bid_price) / mid
    }
}

#[cfg(test)]
mod tests {
    use super::*;

    fn make_quote(bid: f64, ask: f64, bs: f64, ask_s: f64) -> Quote {
        Quote::new(0, bid, ask, bs, ask_s)
    }

    #[test]
    fn test_mid() {
        let q = make_quote(3.0, 2.0, 0.0, 0.0);
        assert_eq!(q.mid(), (q.ask_price + q.bid_price) / 2.0);
    }

    #[test]
    fn test_weighted() {
        let q = make_quote(3.0, 2.0, 5.0, 4.0);
        let expected = (q.ask_price * q.ask_size + q.bid_price * q.bid_size) / (q.ask_size + q.bid_size);
        assert_eq!(q.weighted(), expected);
    }

    #[test]
    fn test_weighted_zero_size() {
        let q = make_quote(3.0, 2.0, 0.0, 0.0);
        assert_eq!(q.weighted(), 0.0);
    }

    #[test]
    fn test_weighted_mid() {
        let q = make_quote(3.0, 2.0, 5.0, 4.0);
        let expected = (q.ask_price * q.bid_size + q.bid_price * q.ask_size) / (q.ask_size + q.bid_size);
        assert_eq!(q.weighted_mid(), expected);
    }

    #[test]
    fn test_weighted_mid_zero_size() {
        let q = make_quote(3.0, 2.0, 0.0, 0.0);
        assert_eq!(q.weighted_mid(), 0.0);
    }

    #[test]
    fn test_spread_bp() {
        let q = make_quote(3.0, 2.0, 0.0, 0.0);
        let expected = 20000.0 * (q.ask_price - q.bid_price) / (q.ask_price + q.bid_price);
        assert_eq!(q.spread_bp(), expected);
    }

    #[test]
    fn test_spread_bp_zero_mid() {
        let q = make_quote(0.0, 0.0, 0.0, 0.0);
        assert_eq!(q.spread_bp(), 0.0);
    }
}
