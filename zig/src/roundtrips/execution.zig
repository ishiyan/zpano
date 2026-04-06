/// OrderSide represents the side of an order (buy or sell).
pub const OrderSide = enum(u1) {
    buy = 0,
    sell = 1,

    /// Returns true if this side is a sell order.
    pub fn isSell(self: OrderSide) bool {
        return self == .sell;
    }
};

/// Execution represents a single trade execution with price, commission, and timing data.
pub const Execution = struct {
    side: OrderSide,
    price: f64,
    commission_per_unit: f64,
    unrealized_price_high: f64,
    unrealized_price_low: f64,
    year: i32,
    month: u8,
    day: u8,
    hour: u8 = 0,
    minute: u8 = 0,
    second: u8 = 0,
};
