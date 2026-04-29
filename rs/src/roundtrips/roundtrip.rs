use crate::daycounting::fractional::DateTime;
use super::execution::{Execution, OrderSide};
use super::side::RoundtripSide;

/// Represents an immutable position round-trip with all computed metrics.
pub struct Roundtrip {
    side: RoundtripSide,
    quantity: f64,
    entry_time: DateTime,
    entry_price: f64,
    exit_time: DateTime,
    exit_price: f64,
    duration_seconds: f64,
    highest_price: f64,
    lowest_price: f64,
    commission: f64,
    gross_pnl: f64,
    net_pnl: f64,
    maximum_adverse_price: f64,
    maximum_favorable_price: f64,
    maximum_adverse_excursion: f64,
    maximum_favorable_excursion: f64,
    entry_efficiency: f64,
    exit_efficiency: f64,
    total_efficiency: f64,
}

impl Roundtrip {
    /// Creates a new Roundtrip from entry and exit executions and a quantity.
    pub fn new(entry: &Execution, exit: &Execution, quantity: f64) -> Self {
        let side = if entry.side == OrderSide::Sell {
            RoundtripSide::Short
        } else {
            RoundtripSide::Long
        };

        let entry_p = entry.price;
        let exit_p = exit.price;

        let pnl = if side == RoundtripSide::Short {
            quantity * (entry_p - exit_p)
        } else {
            quantity * (exit_p - entry_p)
        };

        let commission = (entry.commission_per_unit + exit.commission_per_unit) * quantity;

        let highest_p = if entry.unrealized_price_high > exit.unrealized_price_high {
            entry.unrealized_price_high
        } else {
            exit.unrealized_price_high
        };
        let lowest_p = if entry.unrealized_price_low < exit.unrealized_price_low {
            entry.unrealized_price_low
        } else {
            exit.unrealized_price_low
        };
        let delta = highest_p - lowest_p;

        let mut entry_efficiency = 0.0;
        let mut exit_efficiency = 0.0;
        let mut total_efficiency = 0.0;
        let maximum_adverse_price;
        let maximum_favorable_price;
        let maximum_adverse_excursion;
        let maximum_favorable_excursion;

        if side == RoundtripSide::Long {
            maximum_adverse_price = lowest_p;
            maximum_favorable_price = highest_p;
            maximum_adverse_excursion = 100.0 * (1.0 - lowest_p / entry_p);
            maximum_favorable_excursion = 100.0 * (highest_p / exit_p - 1.0);
            if delta != 0.0 {
                entry_efficiency = 100.0 * (highest_p - entry_p) / delta;
                exit_efficiency = 100.0 * (exit_p - lowest_p) / delta;
                total_efficiency = 100.0 * (exit_p - entry_p) / delta;
            }
        } else {
            maximum_adverse_price = highest_p;
            maximum_favorable_price = lowest_p;
            maximum_adverse_excursion = 100.0 * (highest_p / entry_p - 1.0);
            maximum_favorable_excursion = 100.0 * (1.0 - lowest_p / exit_p);
            if delta != 0.0 {
                entry_efficiency = 100.0 * (entry_p - lowest_p) / delta;
                exit_efficiency = 100.0 * (highest_p - exit_p) / delta;
                total_efficiency = 100.0 * (entry_p - exit_p) / delta;
            }
        }

        let duration_seconds = exit.datetime.diff_seconds(&entry.datetime);

        Self {
            side,
            quantity,
            entry_time: entry.datetime,
            entry_price: entry_p,
            exit_time: exit.datetime,
            exit_price: exit_p,
            duration_seconds,
            highest_price: highest_p,
            lowest_price: lowest_p,
            commission,
            gross_pnl: pnl,
            net_pnl: pnl - commission,
            maximum_adverse_price,
            maximum_favorable_price,
            maximum_adverse_excursion,
            maximum_favorable_excursion,
            entry_efficiency,
            exit_efficiency,
            total_efficiency,
        }
    }

    pub fn side(&self) -> RoundtripSide { self.side }
    pub fn quantity(&self) -> f64 { self.quantity }
    pub fn entry_time(&self) -> &DateTime { &self.entry_time }
    pub fn entry_price(&self) -> f64 { self.entry_price }
    pub fn exit_time(&self) -> &DateTime { &self.exit_time }
    pub fn exit_price(&self) -> f64 { self.exit_price }
    pub fn duration_seconds(&self) -> f64 { self.duration_seconds }
    pub fn highest_price(&self) -> f64 { self.highest_price }
    pub fn lowest_price(&self) -> f64 { self.lowest_price }
    pub fn commission(&self) -> f64 { self.commission }
    pub fn gross_pnl(&self) -> f64 { self.gross_pnl }
    pub fn net_pnl(&self) -> f64 { self.net_pnl }
    pub fn maximum_adverse_price(&self) -> f64 { self.maximum_adverse_price }
    pub fn maximum_favorable_price(&self) -> f64 { self.maximum_favorable_price }
    pub fn maximum_adverse_excursion(&self) -> f64 { self.maximum_adverse_excursion }
    pub fn maximum_favorable_excursion(&self) -> f64 { self.maximum_favorable_excursion }
    pub fn entry_efficiency(&self) -> f64 { self.entry_efficiency }
    pub fn exit_efficiency(&self) -> f64 { self.exit_efficiency }
    pub fn total_efficiency(&self) -> f64 { self.total_efficiency }
}

#[cfg(test)]
mod tests {
    use super::*;

    const EPSILON: f64 = 1e-13;

    fn almost_equal(a: f64, b: f64) -> bool {
        (a - b).abs() < EPSILON
    }

    fn make_long_rt() -> Roundtrip {
        let entry = Execution {
            side: OrderSide::Buy,
            price: 50.0,
            commission_per_unit: 0.01,
            unrealized_price_high: 56.0,
            unrealized_price_low: 48.0,
            datetime: DateTime::new(2024, 1, 1, 9, 30, 0),
        };
        let exit = Execution {
            side: OrderSide::Sell,
            price: 55.0,
            commission_per_unit: 0.02,
            unrealized_price_high: 57.0,
            unrealized_price_low: 49.0,
            datetime: DateTime::new(2024, 1, 5, 16, 0, 0),
        };
        Roundtrip::new(&entry, &exit, 100.0)
    }

    fn make_short_rt() -> Roundtrip {
        let entry = Execution {
            side: OrderSide::Sell,
            price: 80.0,
            commission_per_unit: 0.03,
            unrealized_price_high: 85.0,
            unrealized_price_low: 72.0,
            datetime: DateTime::new(2024, 2, 1, 10, 0, 0),
        };
        let exit = Execution {
            side: OrderSide::Buy,
            price: 72.0,
            commission_per_unit: 0.02,
            unrealized_price_high: 83.0,
            unrealized_price_low: 70.0,
            datetime: DateTime::new(2024, 2, 10, 15, 30, 0),
        };
        Roundtrip::new(&entry, &exit, 200.0)
    }

    fn make_zero_delta_rt() -> Roundtrip {
        let entry = Execution {
            side: OrderSide::Buy,
            price: 100.0,
            commission_per_unit: 0.0,
            unrealized_price_high: 100.0,
            unrealized_price_low: 100.0,
            datetime: DateTime::new(2024, 3, 1, 9, 0, 0),
        };
        let exit = Execution {
            side: OrderSide::Sell,
            price: 100.0,
            commission_per_unit: 0.0,
            unrealized_price_high: 100.0,
            unrealized_price_low: 100.0,
            datetime: DateTime::new(2024, 3, 1, 10, 0, 0),
        };
        Roundtrip::new(&entry, &exit, 50.0)
    }

    fn make_long_loser_rt() -> Roundtrip {
        let entry = Execution {
            side: OrderSide::Buy,
            price: 60.0,
            commission_per_unit: 0.005,
            unrealized_price_high: 62.0,
            unrealized_price_low: 53.0,
            datetime: DateTime::new(2024, 4, 1, 9, 30, 0),
        };
        let exit = Execution {
            side: OrderSide::Sell,
            price: 54.0,
            commission_per_unit: 0.005,
            unrealized_price_high: 61.0,
            unrealized_price_low: 52.0,
            datetime: DateTime::new(2024, 4, 3, 16, 0, 0),
        };
        Roundtrip::new(&entry, &exit, 150.0)
    }

    fn make_short_loser_rt() -> Roundtrip {
        let entry = Execution {
            side: OrderSide::Sell,
            price: 40.0,
            commission_per_unit: 0.01,
            unrealized_price_high: 42.0,
            unrealized_price_low: 39.0,
            datetime: DateTime::new(2024, 5, 1, 10, 0, 0),
        };
        let exit = Execution {
            side: OrderSide::Buy,
            price: 45.0,
            commission_per_unit: 0.01,
            unrealized_price_high: 46.0,
            unrealized_price_low: 38.0,
            datetime: DateTime::new(2024, 5, 5, 15, 0, 0),
        };
        Roundtrip::new(&entry, &exit, 300.0)
    }

    // ---- Long roundtrip ----

    #[test]
    fn test_long_side() {
        let rt = make_long_rt();
        assert_eq!(rt.side(), RoundtripSide::Long);
    }

    #[test]
    fn test_long_quantity() {
        let rt = make_long_rt();
        assert!(almost_equal(rt.quantity(), 100.0));
    }

    #[test]
    fn test_long_entry_price() {
        let rt = make_long_rt();
        assert!(almost_equal(rt.entry_price(), 50.0));
    }

    #[test]
    fn test_long_exit_price() {
        let rt = make_long_rt();
        assert!(almost_equal(rt.exit_price(), 55.0));
    }

    #[test]
    fn test_long_duration() {
        let rt = make_long_rt();
        assert!(almost_equal(rt.duration_seconds(), 369000.0));
    }

    #[test]
    fn test_long_highest_price() {
        let rt = make_long_rt();
        assert!(almost_equal(rt.highest_price(), 57.0));
    }

    #[test]
    fn test_long_lowest_price() {
        let rt = make_long_rt();
        assert!(almost_equal(rt.lowest_price(), 48.0));
    }

    #[test]
    fn test_long_gross_pnl() {
        let rt = make_long_rt();
        assert!(almost_equal(rt.gross_pnl(), 500.0));
    }

    #[test]
    fn test_long_commission() {
        let rt = make_long_rt();
        assert!(almost_equal(rt.commission(), 3.0));
    }

    #[test]
    fn test_long_net_pnl() {
        let rt = make_long_rt();
        assert!(almost_equal(rt.net_pnl(), 497.0));
    }

    #[test]
    fn test_long_maximum_adverse_price() {
        let rt = make_long_rt();
        assert!(almost_equal(rt.maximum_adverse_price(), 48.0));
    }

    #[test]
    fn test_long_maximum_favorable_price() {
        let rt = make_long_rt();
        assert!(almost_equal(rt.maximum_favorable_price(), 57.0));
    }

    #[test]
    fn test_long_mae() {
        let rt = make_long_rt();
        assert!(almost_equal(rt.maximum_adverse_excursion(), 4.0));
    }

    #[test]
    fn test_long_mfe() {
        let rt = make_long_rt();
        let expected = 100.0 * (57.0 / 55.0 - 1.0);
        assert!(almost_equal(rt.maximum_favorable_excursion(), expected));
    }

    #[test]
    fn test_long_entry_efficiency() {
        let rt = make_long_rt();
        let expected = 100.0 * (57.0 - 50.0) / 9.0;
        assert!(almost_equal(rt.entry_efficiency(), expected));
    }

    #[test]
    fn test_long_exit_efficiency() {
        let rt = make_long_rt();
        let expected = 100.0 * (55.0 - 48.0) / 9.0;
        assert!(almost_equal(rt.exit_efficiency(), expected));
    }

    #[test]
    fn test_long_total_efficiency() {
        let rt = make_long_rt();
        let expected = 100.0 * (55.0 - 50.0) / 9.0;
        assert!(almost_equal(rt.total_efficiency(), expected));
    }

    // ---- Short roundtrip ----

    #[test]
    fn test_short_side() {
        let rt = make_short_rt();
        assert_eq!(rt.side(), RoundtripSide::Short);
    }

    #[test]
    fn test_short_quantity() {
        let rt = make_short_rt();
        assert!(almost_equal(rt.quantity(), 200.0));
    }

    #[test]
    fn test_short_entry_price() {
        let rt = make_short_rt();
        assert!(almost_equal(rt.entry_price(), 80.0));
    }

    #[test]
    fn test_short_exit_price() {
        let rt = make_short_rt();
        assert!(almost_equal(rt.exit_price(), 72.0));
    }

    #[test]
    fn test_short_duration() {
        let rt = make_short_rt();
        // Feb 1 10:00 to Feb 10 15:30 = 9 days 5 hours 30 min = 798600 sec
        let expected = 9.0 * 86400.0 + 5.0 * 3600.0 + 30.0 * 60.0;
        assert!(almost_equal(rt.duration_seconds(), expected));
    }

    #[test]
    fn test_short_highest_price() {
        let rt = make_short_rt();
        assert!(almost_equal(rt.highest_price(), 85.0));
    }

    #[test]
    fn test_short_lowest_price() {
        let rt = make_short_rt();
        assert!(almost_equal(rt.lowest_price(), 70.0));
    }

    #[test]
    fn test_short_gross_pnl() {
        let rt = make_short_rt();
        assert!(almost_equal(rt.gross_pnl(), 1600.0));
    }

    #[test]
    fn test_short_commission() {
        let rt = make_short_rt();
        assert!(almost_equal(rt.commission(), 10.0));
    }

    #[test]
    fn test_short_net_pnl() {
        let rt = make_short_rt();
        assert!(almost_equal(rt.net_pnl(), 1590.0));
    }

    #[test]
    fn test_short_maximum_adverse_price() {
        let rt = make_short_rt();
        assert!(almost_equal(rt.maximum_adverse_price(), 85.0));
    }

    #[test]
    fn test_short_maximum_favorable_price() {
        let rt = make_short_rt();
        assert!(almost_equal(rt.maximum_favorable_price(), 70.0));
    }

    #[test]
    fn test_short_mae() {
        let rt = make_short_rt();
        assert!(almost_equal(rt.maximum_adverse_excursion(), 6.25));
    }

    #[test]
    fn test_short_mfe() {
        let rt = make_short_rt();
        let expected = 100.0 * (1.0 - 70.0 / 72.0);
        assert!(almost_equal(rt.maximum_favorable_excursion(), expected));
    }

    #[test]
    fn test_short_entry_efficiency() {
        let rt = make_short_rt();
        let expected = 100.0 * (80.0 - 70.0) / 15.0;
        assert!(almost_equal(rt.entry_efficiency(), expected));
    }

    #[test]
    fn test_short_exit_efficiency() {
        let rt = make_short_rt();
        let expected = 100.0 * (85.0 - 72.0) / 15.0;
        assert!(almost_equal(rt.exit_efficiency(), expected));
    }

    #[test]
    fn test_short_total_efficiency() {
        let rt = make_short_rt();
        let expected = 100.0 * (80.0 - 72.0) / 15.0;
        assert!(almost_equal(rt.total_efficiency(), expected));
    }

    // ---- Zero delta ----

    #[test]
    fn test_zero_delta_entry_efficiency() {
        let rt = make_zero_delta_rt();
        assert!(almost_equal(rt.entry_efficiency(), 0.0));
    }

    #[test]
    fn test_zero_delta_exit_efficiency() {
        let rt = make_zero_delta_rt();
        assert!(almost_equal(rt.exit_efficiency(), 0.0));
    }

    #[test]
    fn test_zero_delta_total_efficiency() {
        let rt = make_zero_delta_rt();
        assert!(almost_equal(rt.total_efficiency(), 0.0));
    }

    #[test]
    fn test_zero_delta_gross_pnl() {
        let rt = make_zero_delta_rt();
        assert!(almost_equal(rt.gross_pnl(), 0.0));
    }

    #[test]
    fn test_zero_delta_net_pnl() {
        let rt = make_zero_delta_rt();
        assert!(almost_equal(rt.net_pnl(), 0.0));
    }

    // ---- Long loser ----

    #[test]
    fn test_long_loser_side() {
        let rt = make_long_loser_rt();
        assert_eq!(rt.side(), RoundtripSide::Long);
    }

    #[test]
    fn test_long_loser_gross_pnl() {
        let rt = make_long_loser_rt();
        assert!(almost_equal(rt.gross_pnl(), -900.0));
    }

    #[test]
    fn test_long_loser_commission() {
        let rt = make_long_loser_rt();
        assert!(almost_equal(rt.commission(), 1.5));
    }

    #[test]
    fn test_long_loser_net_pnl() {
        let rt = make_long_loser_rt();
        assert!(almost_equal(rt.net_pnl(), -901.5));
    }

    #[test]
    fn test_long_loser_highest() {
        let rt = make_long_loser_rt();
        assert!(almost_equal(rt.highest_price(), 62.0));
    }

    #[test]
    fn test_long_loser_lowest() {
        let rt = make_long_loser_rt();
        assert!(almost_equal(rt.lowest_price(), 52.0));
    }

    #[test]
    fn test_long_loser_mae() {
        let rt = make_long_loser_rt();
        let expected = 100.0 * (1.0 - 52.0 / 60.0);
        assert!(almost_equal(rt.maximum_adverse_excursion(), expected));
    }

    #[test]
    fn test_long_loser_mfe() {
        let rt = make_long_loser_rt();
        let expected = 100.0 * (62.0 / 54.0 - 1.0);
        assert!(almost_equal(rt.maximum_favorable_excursion(), expected));
    }

    // ---- Short loser ----

    #[test]
    fn test_short_loser_side() {
        let rt = make_short_loser_rt();
        assert_eq!(rt.side(), RoundtripSide::Short);
    }

    #[test]
    fn test_short_loser_gross_pnl() {
        let rt = make_short_loser_rt();
        assert!(almost_equal(rt.gross_pnl(), -1500.0));
    }

    #[test]
    fn test_short_loser_commission() {
        let rt = make_short_loser_rt();
        assert!(almost_equal(rt.commission(), 6.0));
    }

    #[test]
    fn test_short_loser_net_pnl() {
        let rt = make_short_loser_rt();
        assert!(almost_equal(rt.net_pnl(), -1506.0));
    }

    #[test]
    fn test_short_loser_maximum_adverse_price() {
        let rt = make_short_loser_rt();
        assert!(almost_equal(rt.maximum_adverse_price(), 46.0));
    }

    #[test]
    fn test_short_loser_maximum_favorable_price() {
        let rt = make_short_loser_rt();
        assert!(almost_equal(rt.maximum_favorable_price(), 38.0));
    }

    #[test]
    fn test_short_loser_mae() {
        let rt = make_short_loser_rt();
        assert!(almost_equal(rt.maximum_adverse_excursion(), 15.0));
    }

    #[test]
    fn test_short_loser_mfe() {
        let rt = make_short_loser_rt();
        let expected = 100.0 * (1.0 - 38.0 / 45.0);
        assert!(almost_equal(rt.maximum_favorable_excursion(), expected));
    }
}
