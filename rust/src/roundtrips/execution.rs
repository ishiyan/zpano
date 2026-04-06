use crate::daycounting::fractional::DateTime;

/// Enumerates the sides of an order.
#[derive(Debug, Clone, Copy, PartialEq, Eq)]
#[repr(u8)]
pub enum OrderSide {
    Buy = 0,
    Sell = 1,
}

impl OrderSide {
    /// Returns true if the order side is Sell.
    pub fn is_sell(&self) -> bool {
        *self == OrderSide::Sell
    }
}

/// Represents an order execution (fill).
#[derive(Debug, Clone, Copy)]
pub struct Execution {
    /// The side of the order.
    pub side: OrderSide,
    /// The execution price.
    pub price: f64,
    /// The commission per unit of quantity.
    pub commission_per_unit: f64,
    /// The highest unrealized price during the execution period.
    pub unrealized_price_high: f64,
    /// The lowest unrealized price during the execution period.
    pub unrealized_price_low: f64,
    /// The date and time of the execution.
    pub datetime: DateTime,
}

#[cfg(test)]
mod tests {
    use super::*;

    #[test]
    fn test_order_side_buy_is_not_sell() {
        assert!(!OrderSide::Buy.is_sell());
    }

    #[test]
    fn test_order_side_sell_is_sell() {
        assert!(OrderSide::Sell.is_sell());
    }
}
