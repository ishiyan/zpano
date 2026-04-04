use super::conventions::DayCountConvention;
use super::daycounting::{self, date_to_jd};

const SECONDS_IN_GREGORIAN_YEAR: f64 = 31556952.0;
const SECONDS_IN_DAY: f64 = 86400.0;

/// A simple date-time struct used for day count calculations.
#[derive(Debug, Clone, Copy, PartialEq)]
pub struct DateTime {
    pub year: i32,
    pub month: i32,
    pub day: i32,
    pub hour: i32,
    pub minute: i32,
    pub second: i32,
}

impl DateTime {
    pub fn new(year: i32, month: i32, day: i32, hour: i32, minute: i32, second: i32) -> Self {
        Self { year, month, day, hour, minute, second }
    }

    /// Creates a DateTime with time components set to 0.
    pub fn date(year: i32, month: i32, day: i32) -> Self {
        Self { year, month, day, hour: 0, minute: 0, second: 0 }
    }

    /// Returns the Julian Day number for this date.
    pub fn jd(&self) -> i32 {
        date_to_jd(self.year, self.month, self.day)
    }

    /// Returns the time-of-day as a fraction of 24 hours.
    pub fn time_fraction(&self) -> f64 {
        (self.hour as f64 * 3600.0 + self.minute as f64 * 60.0 + self.second as f64) / 86400.0
    }

    /// Returns total seconds from midnight.
    fn seconds_from_midnight(&self) -> f64 {
        self.hour as f64 * 3600.0 + self.minute as f64 * 60.0 + self.second as f64
    }

    /// Returns true if self is after other (by JD + time).
    fn is_after(&self, other: &DateTime) -> bool {
        let jd1 = self.jd();
        let jd2 = other.jd();
        if jd1 != jd2 {
            return jd1 > jd2;
        }
        self.seconds_from_midnight() > other.seconds_from_midnight()
    }

    /// Returns total difference in seconds between two DateTimes.
    fn diff_seconds(&self, other: &DateTime) -> f64 {
        let jd_diff = (self.jd() - other.jd()) as f64;
        let sec_diff = self.seconds_from_midnight() - other.seconds_from_midnight();
        jd_diff * SECONDS_IN_DAY + sec_diff
    }
}

/// Calculates the fraction between two dates using a specified day count convention.
///
/// If `day_frac` is true, returns fraction in days; if false, returns fraction in years.
pub fn frac(
    date_time1: &DateTime,
    date_time2: &DateTime,
    method: DayCountConvention,
    day_frac: bool,
) -> Result<f64, String> {
    let (dt1, dt2) = if date_time1.is_after(date_time2) {
        (date_time2, date_time1)
    } else {
        (date_time1, date_time2)
    };

    if method == DayCountConvention::Raw {
        let diff_seconds = dt2.diff_seconds(dt1);
        if day_frac {
            return Ok(diff_seconds / SECONDS_IN_DAY);
        }
        return Ok(diff_seconds / SECONDS_IN_GREGORIAN_YEAR);
    }

    let y1 = dt1.year;
    let m1 = dt1.month;
    let d1 = dt1.day;
    let y2 = dt2.year;
    let m2 = dt2.month;
    let d2 = dt2.day;

    let tm1 = dt1.time_fraction();
    let tm2 = dt2.time_fraction();

    match daycounting::dispatch(method, y1, m1, d1, y2, m2, d2, tm1, tm2, day_frac) {
        Some(v) => Ok(v),
        None => Err(format!("unknown day count convention: {:?}", method)),
    }
}

/// Calculates the year fraction between two dates.
pub fn year_frac(
    date_time1: &DateTime,
    date_time2: &DateTime,
    method: DayCountConvention,
) -> Result<f64, String> {
    frac(date_time1, date_time2, method, false)
}

/// Calculates the day fraction between two dates.
pub fn day_frac(
    date_time1: &DateTime,
    date_time2: &DateTime,
    method: DayCountConvention,
) -> Result<f64, String> {
    frac(date_time1, date_time2, method, true)
}

#[cfg(test)]
mod tests {
    use super::*;

    const EPSILON: f64 = 1e-14;
    const SECONDS_IN_LEAP_YEAR: f64 = 31622400.0;
    const SECONDS_IN_NON_LEAP_YEAR: f64 = 31536000.0;

    fn almost_equal(a: f64, b: f64) -> bool {
        (a - b).abs() < EPSILON
    }

    #[test]
    fn test_raw_year_frac() {
        let dt1 = DateTime::date(2024, 1, 1);
        let dt2 = DateTime::date(2025, 1, 1);
        let result = year_frac(&dt1, &dt2, DayCountConvention::Raw).unwrap();
        assert!(almost_equal(result, SECONDS_IN_LEAP_YEAR / SECONDS_IN_GREGORIAN_YEAR));
    }

    #[test]
    fn test_raw_day_frac() {
        let dt1 = DateTime::date(2024, 1, 1);
        let dt2 = DateTime::date(2025, 1, 1);
        let result = day_frac(&dt1, &dt2, DayCountConvention::Raw).unwrap();
        assert!(almost_equal(result, SECONDS_IN_LEAP_YEAR / SECONDS_IN_DAY));
    }

    #[test]
    fn test_raw_non_leap_year_frac() {
        let dt1 = DateTime::date(2023, 1, 1);
        let dt2 = DateTime::date(2024, 1, 1);
        let result = year_frac(&dt1, &dt2, DayCountConvention::Raw).unwrap();
        assert!(almost_equal(result, SECONDS_IN_NON_LEAP_YEAR / SECONDS_IN_GREGORIAN_YEAR));
    }

    #[test]
    fn test_swapped_dates() {
        let dt1 = DateTime::date(2024, 10, 15);
        let dt2 = DateTime::date(2024, 7, 15);
        let result = year_frac(&dt1, &dt2, DayCountConvention::Thirty360Eu).unwrap();
        assert!(almost_equal(result, 90.0 / 360.0));
    }

    #[test]
    fn test_intraday() {
        let dt1 = DateTime::new(2024, 7, 15, 6, 0, 0);
        let dt2 = DateTime::new(2024, 7, 15, 18, 0, 0);
        let result = year_frac(&dt1, &dt2, DayCountConvention::Raw).unwrap();
        let expected = 43200.0 / SECONDS_IN_GREGORIAN_YEAR;
        assert!(almost_equal(result, expected));
    }

    #[test]
    fn test_valid_methods_dispatch() {
        let dt1 = DateTime::date(2024, 7, 15);
        let dt2 = DateTime::date(2024, 10, 15);
        let methods = vec![
            DayCountConvention::Thirty360Us,
            DayCountConvention::Thirty360UsEom,
            DayCountConvention::Thirty360UsNasd,
            DayCountConvention::Thirty360Eu,
            DayCountConvention::Thirty360EuM2,
            DayCountConvention::Thirty360EuM3,
            DayCountConvention::Thirty360EuPlus,
            DayCountConvention::Thirty365,
            DayCountConvention::Act360,
            DayCountConvention::Act365Fixed,
            DayCountConvention::Act365Nonleap,
            DayCountConvention::ActActExcel,
            DayCountConvention::ActActIsda,
            DayCountConvention::ActActAfb,
        ];
        for method in methods {
            let result = year_frac(&dt1, &dt2, method);
            assert!(result.is_ok(), "Expected Ok for {:?}, got {:?}", method, result);
            assert!(result.unwrap() > 0.0, "Expected positive year frac for {:?}", method);
        }
    }

    #[test]
    fn test_act_365_fixed_specific() {
        let dt1 = DateTime::date(2024, 7, 15);
        let dt2 = DateTime::date(2024, 10, 15);
        let result = year_frac(&dt1, &dt2, DayCountConvention::Act365Fixed).unwrap();
        assert!(almost_equal(result, 92.0 / 365.0));
    }

    #[test]
    fn test_act_360_specific() {
        let dt1 = DateTime::date(2024, 7, 15);
        let dt2 = DateTime::date(2024, 10, 15);
        let result = year_frac(&dt1, &dt2, DayCountConvention::Act360).unwrap();
        assert!(almost_equal(result, 92.0 / 360.0));
    }

    #[test]
    fn test_act_act_excel_specific() {
        let dt1 = DateTime::date(2024, 7, 15);
        let dt2 = DateTime::date(2024, 10, 15);
        let result = year_frac(&dt1, &dt2, DayCountConvention::ActActExcel).unwrap();
        assert!(almost_equal(result, 92.0 / 366.0));
    }
}
