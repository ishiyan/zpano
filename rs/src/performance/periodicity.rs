/// Periodicity represents the frequency of performance measurement periods.
#[derive(Debug, Clone, Copy, PartialEq, Eq)]
#[repr(u8)]
pub enum Periodicity {
    Daily = 0,
    Weekly = 1,
    Monthly = 2,
    Quarterly = 3,
    Annual = 4,
}

impl Periodicity {
    /// Returns the number of periods per year for a given periodicity.
    pub fn periods_per_annum(&self) -> i32 {
        match self {
            Self::Daily => 252,
            Self::Weekly => 52,
            Self::Monthly => 12,
            Self::Quarterly => 4,
            Self::Annual => 1,
        }
    }

    /// Returns the number of trading days per period for a given periodicity.
    pub fn days_per_period(&self) -> f64 {
        match self {
            Self::Daily => 1.0,
            Self::Weekly => 252.0 / 52.0,
            Self::Monthly => 252.0 / 12.0,
            Self::Quarterly => 252.0 / 4.0,
            Self::Annual => 252.0,
        }
    }
}

#[cfg(test)]
mod tests {
    use super::*;

    #[test]
    fn test_periods_per_annum() {
        assert_eq!(Periodicity::Daily.periods_per_annum(), 252);
        assert_eq!(Periodicity::Weekly.periods_per_annum(), 52);
        assert_eq!(Periodicity::Monthly.periods_per_annum(), 12);
        assert_eq!(Periodicity::Quarterly.periods_per_annum(), 4);
        assert_eq!(Periodicity::Annual.periods_per_annum(), 1);
    }

    #[test]
    fn test_days_per_period() {
        assert_eq!(Periodicity::Daily.days_per_period(), 1.0);
        assert_eq!(Periodicity::Weekly.days_per_period(), 252.0 / 52.0);
        assert_eq!(Periodicity::Monthly.days_per_period(), 252.0 / 12.0);
        assert_eq!(Periodicity::Quarterly.days_per_period(), 252.0 / 4.0);
        assert_eq!(Periodicity::Annual.days_per_period(), 252.0);
    }
}
