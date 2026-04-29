use super::conventions::DayCountConvention;

/// Returns true if the given year is a leap year.
pub fn is_leap_year(y: i32) -> bool {
    y % 4 == 0 && (y % 100 != 0 || y % 400 == 0)
}

/// Converts a date to Julian Day number.
pub fn date_to_jd(year: i32, month: i32, day: i32) -> i32 {
    let a = (14 - month) / 12;
    let y = year + 4800 - a;
    let m = month + 12 * a - 3;

    let mut jd = day + (153 * m + 2) / 5 + y * 365;
    jd += y / 4 - y / 100 + y / 400 - 32045;
    jd
}

/// Converts a Julian Day number to a date (year, month, day).
pub fn jd_to_date(jd: i32) -> (i32, i32, i32) {
    let a = jd + 32044;
    let b = (4 * a + 3) / 146097;
    let c = a - (b * 146097) / 4;

    let d = (4 * c + 3) / 1461;
    let e = c - (d * 1461) / 4;
    let m = (5 * e + 2) / 153;
    let m2 = m / 10;

    let day = e + 1 - (153 * m + 2) / 5;
    let month = m + 3 - 12 * m2;
    let year = b * 100 + d - 4800 + m2;

    (year, month, day)
}

/// 30/360 European (30S/360, Eurobond Basis, ICMA). ISO 20022: A011.
pub fn eur_30_360(
    y1: i32, m1: i32, d1: i32, y2: i32, m2: i32, d2: i32,
    df1: f64, df2: f64, frac_days: bool,
) -> f64 {
    let mut diff_days = (360 * (y2 - y1) + 30 * (m2 - m1)) as f64 + df2 - df1;

    let d2_adj = d2.min(30);
    let d1_adj = d1.min(30);

    diff_days += (d2_adj - d1_adj) as f64;

    if frac_days { diff_days } else { diff_days / 360.0 }
}

/// 30E2/360 Eurobond basis model 2. ISO 20022: A012.
pub fn eur_30_360_model2(
    y1: i32, m1: i32, d1: i32, y2: i32, m2: i32, d2: i32,
    df1: f64, df2: f64, frac_days: bool,
) -> f64 {
    let mut diff_days = (360 * (y2 - y1) + 30 * (m2 - m1)) as f64 + df2 - df1;
    let leap1 = is_leap_year(y1);
    let mut d2_adj = d2;

    if leap1 && m2 == 2 && d2 == 28 {
        if d1 == 29 {
            d2_adj = 29;
        } else if d1 >= 30 {
            d2_adj = 30;
        }
    } else if leap1 && m2 == 2 && d2 == 29 {
        if d1 >= 30 {
            d2_adj = 30;
        }
    } else if d2 > 30 {
        d2_adj = 30;
    }

    let d1_adj = d1.min(30);
    diff_days += (d2_adj - d1_adj) as f64;

    if frac_days { diff_days } else { diff_days / 360.0 }
}

/// 30E3/360 Eurobond basis model 3. ISO 20022: A013.
pub fn eur_30_360_model3(
    y1: i32, m1: i32, d1: i32, y2: i32, m2: i32, d2: i32,
    df1: f64, df2: f64, frac_days: bool,
) -> f64 {
    let mut diff_days = (360 * (y2 - y1) + 30 * (m2 - m1)) as f64 + df2 - df1;

    let d2_adj = if m2 == 2 && d2 >= 28 {
        30
    } else if d2 > 30 {
        30
    } else {
        d2
    };

    let d1_adj = if m1 == 2 && d1 >= 28 {
        30
    } else if d1 > 30 {
        30
    } else {
        d1
    };

    diff_days += (d2_adj - d1_adj) as f64;

    if frac_days { diff_days } else { diff_days / 360.0 }
}

/// 30E+/360.
pub fn eur_30_360_plus(
    y1: i32, m1: i32, d1: i32, y2: i32, m2: i32, d2: i32,
    df1: f64, df2: f64, frac_days: bool,
) -> f64 {
    let mut diff_days = (360 * (y2 - y1) + 30 * (m2 - m1)) as f64 + df2 - df1;

    let d2_adj = if d2 == 31 { 32 } else { d2 };
    let d1_adj = if d1 > 30 { 30 } else { d1 };

    diff_days += (d2_adj - d1_adj) as f64;

    if frac_days { diff_days } else { diff_days / 360.0 }
}

/// 30/360 US (ISDA, Bond Basis). ISO 20022: A001.
pub fn us_30_360(
    y1: i32, m1: i32, d1: i32, y2: i32, m2: i32, d2: i32,
    df1: f64, df2: f64, frac_days: bool,
) -> f64 {
    let mut diff_days = (360 * (y2 - y1) + 30 * (m2 - m1)) as f64 + df2 - df1;

    let d2_adj = if d2 == 31 && d1 >= 30 { 30 } else { d2 };
    let d1_adj = d1.min(30);

    diff_days += (d2_adj - d1_adj) as f64;

    if frac_days { diff_days } else { diff_days / 360.0 }
}

/// 30/360 US End-Of-Month.
pub fn us_30_360_eom(
    y1: i32, m1: i32, d1: i32, y2: i32, m2: i32, d2: i32,
    df1: f64, df2: f64, frac_days: bool,
) -> f64 {
    let mut diff_days = (360 * (y2 - y1) + 30 * (m2 - m1)) as f64 + df2 - df1;

    let rule2 = m1 == 2 && d1 >= 28;
    let rule3 = rule2 && m2 == 2 && d2 >= 28;
    let rule4 = d2 == 31 && d1 >= 30;

    let d1_adj = if rule2 { 30 } else if d1 > 30 { 30 } else { d1 };
    let d2_adj = if rule4 || rule3 { 30 } else { d2 };

    diff_days += (d2_adj - d1_adj) as f64;

    if frac_days { diff_days } else { diff_days / 360.0 }
}

/// 30/360 NASD.
pub fn us_30_360_nasd(
    y1: i32, m1: i32, d1: i32, y2: i32, m2: i32, d2: i32,
    df1: f64, df2: f64, frac_days: bool,
) -> f64 {
    let mut diff_days = (360 * (y2 - y1) + 30 * (m2 - m1)) as f64 + df2 - df1;

    let d2_adj = if d2 == 31 {
        if d1 < 30 { 32 } else { 30 }
    } else {
        d2
    };

    let d1_adj = d1.min(30);
    diff_days += (d2_adj - d1_adj) as f64;

    if frac_days { diff_days } else { diff_days / 360.0 }
}

/// 30/365. ISO 20022: A002.
pub fn thirty_365(
    y1: i32, m1: i32, d1: i32, y2: i32, m2: i32, d2: i32,
    df1: f64, df2: f64, frac_days: bool,
) -> f64 {
    let mut diff_days = (360 * (y2 - y1) + 30 * (m2 - m1)) as f64 + df2 - df1;

    let d2_adj = if d2 == 31 && d1 >= 30 { 30 } else { d2 };
    let d1_adj = d1.min(30);

    diff_days += (d2_adj - d1_adj) as f64;

    if frac_days { diff_days } else { diff_days / 365.0 }
}

/// Actual/365 Non-Leap. ISO 20022: A014.
pub fn act_365_nonleap(
    y1: i32, m1: i32, d1: i32, y2: i32, m2: i32, d2: i32,
    df1: f64, df2: f64, frac_days: bool,
) -> f64 {
    let mut diff_days = (date_to_jd(y2, m2, d2) - date_to_jd(y1, m1, d1)) as f64 + df2 - df1;

    let mut leap_years = 0;
    if is_leap_year(y1) && m1 <= 2 {
        leap_years += 1;
    }
    if y1 != y2 && is_leap_year(y2) && m2 >= 3 {
        leap_years += 1;
    }
    if y1 + 1 < y2 {
        for now in (y1 + 1)..y2 {
            if is_leap_year(now) {
                leap_years += 1;
            }
        }
    }

    diff_days -= leap_years as f64;

    if frac_days { diff_days } else { diff_days / 365.0 }
}

/// Actual/365 Fixed. ISO 20022: A005.
pub fn act_365_fixed(
    y1: i32, m1: i32, d1: i32, y2: i32, m2: i32, d2: i32,
    df1: f64, df2: f64, frac_days: bool,
) -> f64 {
    let diff_days = (date_to_jd(y2, m2, d2) - date_to_jd(y1, m1, d1)) as f64 + df2 - df1;

    if frac_days { diff_days } else { diff_days / 365.0 }
}

/// Actual/360. ISO 20022: A004.
pub fn act_360(
    y1: i32, m1: i32, d1: i32, y2: i32, m2: i32, d2: i32,
    df1: f64, df2: f64, frac_days: bool,
) -> f64 {
    let diff_days = (date_to_jd(y2, m2, d2) - date_to_jd(y1, m1, d1)) as f64 + df2 - df1;

    if frac_days { diff_days } else { diff_days / 360.0 }
}

fn feb29_between(y1: i32, m1: i32, d1: i32, y2: i32, m2: i32, d2: i32) -> bool {
    let jd1 = date_to_jd(y1, m1, d1);
    let jd2 = date_to_jd(y2, m2, d2);
    for y in y1..=y2 {
        if is_leap_year(y) {
            let feb29_jd = date_to_jd(y, 2, 29);
            if jd1 <= feb29_jd && feb29_jd <= jd2 {
                return true;
            }
        }
    }
    false
}

fn appears_le_year(y1: i32, m1: i32, d1: i32, y2: i32, m2: i32, d2: i32) -> bool {
    if y1 == y2 {
        return true;
    }
    if y1 + 1 == y2 && (m1 > m2 || (m1 == m2 && d1 >= d2)) {
        return true;
    }
    false
}

/// Excel's Actual/Actual (basis 1) method.
pub fn act_act_excel(
    y1: i32, m1: i32, d1: i32, y2: i32, m2: i32, d2: i32,
    df1: f64, df2: f64, frac_days: bool,
) -> f64 {
    let day_diff = (date_to_jd(y2, m2, d2) - date_to_jd(y1, m1, d1)) as f64;

    if appears_le_year(y1, m1, d1, y2, m2, d2) {
        let year_days = if y1 == y2 && is_leap_year(y1) {
            366.0
        } else if feb29_between(y1, m1, d1, y2, m2, d2) || (m2 == 2 && d2 == 29) {
            366.0
        } else {
            365.0
        };
        if frac_days {
            day_diff + df2 - df1
        } else {
            (day_diff + df2 - df1) / year_days
        }
    } else {
        let jd_start1 = date_to_jd(y1, 1, 1);
        let jd_start2 = date_to_jd(y2 + 1, 1, 1);
        let year_days = (jd_start2 - jd_start1) as f64;
        let avg_year_days = year_days / (y2 - y1 + 1) as f64;
        if frac_days {
            day_diff + df2 - df1
        } else {
            (day_diff + df2 - df1) / avg_year_days
        }
    }
}

/// Actual/Actual ISDA. ISO 20022: A008.
pub fn act_act_isda(
    y1: i32, m1: i32, d1: i32, y2: i32, m2: i32, d2: i32,
    df1: f64, df2: f64, frac_days: bool,
) -> f64 {
    if y1 == y2 {
        let denom = if is_leap_year(y2) { 366.0 } else { 365.0 };
        let diff_days = (date_to_jd(y2, m2, d2) - date_to_jd(y1, m1, d1)) as f64 + df2 - df1;
        if frac_days { diff_days } else { diff_days / denom }
    } else {
        let denom_a = if is_leap_year(y1) { 366.0 } else { 365.0 };
        let diff_a = (date_to_jd(y1, 12, 31) - date_to_jd(y1, m1, d1) + 1) as f64;

        let denom_b = if is_leap_year(y2) { 366.0 } else { 365.0 };
        let diff_b = (date_to_jd(y2, m2, d2) - date_to_jd(y2, 1, 1)) as f64;

        if frac_days {
            let mut diff = diff_a - df1 + diff_b + df2;
            for year in (y1 + 1)..y2 {
                diff += if is_leap_year(year) { 366.0 } else { 365.0 };
            }
            diff
        } else {
            (diff_a - df1) / denom_a + (diff_b + df2) / denom_b + (y2 - y1 - 1) as f64
        }
    }
}

/// Actual/Actual AFB. ISO 20022: A010.
pub fn act_act_afb(
    y1: i32, m1: i32, d1: i32, y2: i32, m2: i32, d2: i32,
    df1: f64, df2: f64, frac_days: bool,
) -> f64 {
    if y1 == y2 {
        let denom = if m1 < 3 && is_leap_year(y1) { 366.0 } else { 365.0 };
        let diff_days = (date_to_jd(y2, m2, d2) - date_to_jd(y1, m1, d1)) as f64 + df2 - df1;
        if frac_days { diff_days } else { diff_days / denom }
    } else {
        let denom_a = if m1 < 3 && is_leap_year(y1) { 366.0 } else { 365.0 };
        let diff_a = (date_to_jd(y1, 12, 31) - date_to_jd(y1, m1, d1) + 1) as f64;

        let denom_b = if m2 >= 3 && is_leap_year(y2) { 366.0 } else { 365.0 };
        let diff_b = (date_to_jd(y2, m2, d2) - date_to_jd(y2, 1, 1)) as f64;

        if frac_days {
            let mut diff = diff_a - df1 + diff_b + df2;
            for year in (y1 + 1)..y2 {
                diff += if is_leap_year(year) { 366.0 } else { 365.0 };
            }
            diff
        } else {
            (diff_a - df1) / denom_a + (diff_b + df2) / denom_b + (y2 - y1 - 1) as f64
        }
    }
}

/// Dispatches to the appropriate day count function based on the convention.
/// Returns `None` for `Raw` convention (handled separately in fractional).
pub fn dispatch(
    convention: DayCountConvention,
    y1: i32, m1: i32, d1: i32, y2: i32, m2: i32, d2: i32,
    df1: f64, df2: f64, frac_days: bool,
) -> Option<f64> {
    match convention {
        DayCountConvention::Raw => None,
        DayCountConvention::Thirty360Us => Some(us_30_360(y1, m1, d1, y2, m2, d2, df1, df2, frac_days)),
        DayCountConvention::Thirty360UsEom => Some(us_30_360_eom(y1, m1, d1, y2, m2, d2, df1, df2, frac_days)),
        DayCountConvention::Thirty360UsNasd => Some(us_30_360_nasd(y1, m1, d1, y2, m2, d2, df1, df2, frac_days)),
        DayCountConvention::Thirty360Eu => Some(eur_30_360(y1, m1, d1, y2, m2, d2, df1, df2, frac_days)),
        DayCountConvention::Thirty360EuM2 => Some(eur_30_360_model2(y1, m1, d1, y2, m2, d2, df1, df2, frac_days)),
        DayCountConvention::Thirty360EuM3 => Some(eur_30_360_model3(y1, m1, d1, y2, m2, d2, df1, df2, frac_days)),
        DayCountConvention::Thirty360EuPlus => Some(eur_30_360_plus(y1, m1, d1, y2, m2, d2, df1, df2, frac_days)),
        DayCountConvention::Thirty365 => Some(thirty_365(y1, m1, d1, y2, m2, d2, df1, df2, frac_days)),
        DayCountConvention::Act360 => Some(act_360(y1, m1, d1, y2, m2, d2, df1, df2, frac_days)),
        DayCountConvention::Act365Fixed => Some(act_365_fixed(y1, m1, d1, y2, m2, d2, df1, df2, frac_days)),
        DayCountConvention::Act365Nonleap => Some(act_365_nonleap(y1, m1, d1, y2, m2, d2, df1, df2, frac_days)),
        DayCountConvention::ActActExcel => Some(act_act_excel(y1, m1, d1, y2, m2, d2, df1, df2, frac_days)),
        DayCountConvention::ActActIsda => Some(act_act_isda(y1, m1, d1, y2, m2, d2, df1, df2, frac_days)),
        DayCountConvention::ActActAfb => Some(act_act_afb(y1, m1, d1, y2, m2, d2, df1, df2, frac_days)),
    }
}

#[cfg(test)]
mod tests {
    use super::*;

    const EPSILON: f64 = 1e-14;

    fn almost_equal(a: f64, b: f64) -> bool {
        (a - b).abs() < EPSILON
    }

    #[test]
    fn test_is_leap_year() {
        let leap_years = vec![
            1904, 1908, 1912, 1916, 1920, 1924, 1928, 1932, 1936, 1940,
            1944, 1948, 1952, 1956, 1960, 1964, 1968, 1972, 1976, 1980,
            1984, 1988, 1992, 1996, 2000, 2004, 2008, 2012, 2016, 2020,
            2024,
        ];
        let non_leap_years = vec![
            1900, 1901, 1902, 1903, 1905, 1906, 1907, 1909, 1910, 1911,
            1913, 1914, 1915, 1917, 1918, 1919, 1921, 1922, 1923, 1925,
            1926, 1927, 1929, 1930, 1931, 1933, 1934, 1935, 1937, 1938,
            1939, 1941, 1942, 1943, 1945, 1946, 1947, 1949, 1950, 1951,
            1953, 1954, 1955, 1957, 1958, 1959, 1961, 1962, 1963, 1965,
            1966, 1967, 1969, 1970, 1971, 1973, 1974, 1975, 1977, 1978,
            1979, 1981, 1982, 1983, 1985, 1986, 1987, 1989, 1990, 1991,
            1993, 1994, 1995, 1997, 1998, 1999, 2001, 2002, 2003, 2005,
            2006, 2007, 2009, 2010, 2011, 2013, 2014, 2015, 2017, 2018,
            2019, 2021, 2022, 2023, 2025, 2026, 2027, 2100,
        ];
        for y in leap_years {
            assert!(is_leap_year(y), "Expected {} to be leap year", y);
        }
        for y in non_leap_years {
            assert!(!is_leap_year(y), "Expected {} to NOT be leap year", y);
        }
    }

    #[test]
    fn test_date_to_jd_and_back() {
        let cases = vec![
            (2000, 1, 1, 2451545),
            (1999, 12, 31, 2451544),
            (2024, 2, 29, 2460370),
            (1900, 1, 1, 2415021),
            (2100, 12, 31, 2488434),
            (1582, 10, 15, 2299161),
            (2024, 7, 15, 2460507),
        ];
        for (y, m, d, expected_jd) in cases {
            let jd = date_to_jd(y, m, d);
            assert_eq!(jd, expected_jd, "DateToJD({}, {}, {}) = {}, want {}", y, m, d, jd, expected_jd);
            let (ry, rm, rd) = jd_to_date(jd);
            assert_eq!((ry, rm, rd), (y, m, d), "JDToDate({}) = ({},{},{}), want ({},{},{})", jd, ry, rm, rd, y, m, d);
        }
    }

    const FD2_360: f64 = 0.2 / 360.0;
    const FD2_365: f64 = 0.2 / 365.0;
    const FD2_366: f64 = 0.2 / 366.0;

    // Eur30360 tests
    #[test]
    fn test_eur_30_360_basic() {
        let result = eur_30_360(2024, 7, 15, 2024, 10, 15, 0.0, 0.0, false);
        assert!(almost_equal(result, 90.0 / 360.0));
    }

    #[test]
    fn test_eur_30_360_time_fractions() {
        let cases = vec![
            (2024, 7, 15, 2024, 10, 15, 0.0, 0.0, 90.0 / 360.0),
            (2024, 7, 15, 2024, 10, 15, 0.5, 0.0, 89.5 / 360.0),
            (2024, 7, 15, 2024, 10, 15, 0.0, 0.5, 90.5 / 360.0),
            (2024, 7, 15, 2024, 10, 15, 0.5, 0.5, 90.0 / 360.0),
            (2024, 7, 15, 2024, 10, 15, 0.2, 0.8, 90.6 / 360.0),
        ];
        for (y1, m1, d1, y2, m2, d2, df1, df2, expected) in cases {
            let result = eur_30_360(y1, m1, d1, y2, m2, d2, df1, df2, false);
            assert!(almost_equal(result, expected),
                "eur_30_360({},{},{},{},{},{},{},{}) = {}, want {}",
                y1, m1, d1, y2, m2, d2, df1, df2, result, expected);
        }
    }

    #[test]
    fn test_eur_30_360_frac_days() {
        let result = eur_30_360(2024, 7, 15, 2024, 10, 15, 0.0, 0.0, true);
        assert!(almost_equal(result, 90.0));
    }

    // Excel compatibility tests for Eur30360 (basis 4) — from Go test data
    #[test]
    fn test_eur_30_360_excel_basis_4() {
        let cases: Vec<(i32,i32,i32,i32,i32,i32,f64,i32)> = vec![
            (1978, 2, 28, 2020, 5, 17, 42.21944444444444, 13),
            (1993, 12, 2, 2022, 4, 18, 28.37777777777780, 13),
            (2018, 12, 15, 2019, 3, 1, 0.211111111111111, 13),
            (2018, 12, 31, 2019, 1, 1, 0.0027777777777778, 13),
            (1994, 6, 30, 1997, 6, 30, 3.0000000000000000, 16),
            (1994, 2, 10, 1994, 6, 30, 0.3888888888888889, 13),
            (2020, 2, 21, 2024, 3, 25, 4.0944444444444440, 13),
            (2020, 2, 29, 2021, 2, 28, 0.9972222222222222, 13),
            (2020, 1, 31, 2021, 2, 28, 1.0777777777777777, 13),
            (2020, 1, 31, 2021, 3, 31, 1.1666666666666667, 13),
            (2020, 1, 31, 2020, 4, 30, 0.2500000000000000, 16),
            (2018, 2, 5, 2023, 5, 14, 5.2750000000000000, 16),
            (2020, 2, 29, 2024, 2, 28, 3.9972222222222222, 13),
            (2010, 3, 31, 2015, 8, 30, 5.4166666666666667, 13),
            (2016, 2, 28, 2016, 10, 30, 0.6722222222222222, 13),
            (2014, 1, 31, 2014, 8, 31, 0.5833333333333333, 13),
            (2014, 2, 28, 2014, 9, 30, 0.5888888888888889, 13),
            (2016, 2, 29, 2016, 6, 15, 0.29444444444444445, 13),
            (2024, 1, 1, 2024, 12, 3, 0.9222222222222223, 13),
            (2024, 1, 1, 2025, 1, 2, 1.0027777777777800, 13),
            (2024, 1, 1, 2024, 2, 29, 0.1611111111111110, 13),
            (2024, 1, 1, 2024, 3, 1, 0.1666666666666670, 13),
            (2023, 1, 1, 2023, 3, 1, 0.1666666666666670, 13),
            (2024, 2, 29, 2025, 2, 28, 0.9972222222222220, 13),
            (2024, 1, 1, 2028, 12, 31, 4.9972222222222200, 13),
            (2024, 3, 1, 2025, 3, 1, 1.0000000000000000, 16),
            (2024, 2, 29, 2025, 3, 1, 1.0055555555555600, 13),
            (2024, 2, 29, 2028, 2, 28, 3.9972222222222200, 13),
            (2024, 2, 29, 2028, 2, 29, 4.0000000000000000, 16),
            (2024, 3, 1, 2028, 3, 1, 4.0000000000000000, 16),
        ];
        for (y1, m1, d1, y2, m2, d2, expected, prec) in cases {
            let result = eur_30_360(y1, m1, d1, y2, m2, d2, 0.0, 0.0, false);
            let tol = 10f64.powi(-prec);
            assert!((result - expected).abs() < tol,
                "eur_30_360({},{},{},{},{},{}) = {:.16}, want {:.16} (prec={})",
                y1, m1, d1, y2, m2, d2, result, expected, prec);
        }
    }

    // Act360 tests
    #[test]
    fn test_act_360_basic() {
        let result = act_360(2024, 7, 15, 2024, 10, 15, 0.0, 0.0, false);
        assert!(almost_equal(result, 92.0 / 360.0));
    }

    #[test]
    fn test_act_360_time_fractions() {
        let cases = vec![
            (2024, 7, 15, 2024, 10, 15, 0.0, 0.0, 92.0 / 360.0),
            (2024, 7, 15, 2024, 10, 15, 0.5, 0.0, 91.5 / 360.0),
            (2024, 7, 15, 2024, 10, 15, 0.0, 0.5, 92.5 / 360.0),
            (2024, 7, 15, 2024, 10, 15, 0.5, 0.5, 92.0 / 360.0),
            (2024, 7, 15, 2024, 10, 15, 0.2, 0.8, 92.6 / 360.0),
        ];
        for (y1, m1, d1, y2, m2, d2, df1, df2, expected) in cases {
            let result = act_360(y1, m1, d1, y2, m2, d2, df1, df2, false);
            assert!(almost_equal(result, expected));
        }
    }

    // Act365Fixed tests
    #[test]
    fn test_act_365_fixed_basic() {
        let result = act_365_fixed(2024, 7, 15, 2024, 10, 15, 0.0, 0.0, false);
        assert!(almost_equal(result, 92.0 / 365.0));
    }

    #[test]
    fn test_act_365_fixed_time_fractions() {
        let cases = vec![
            (2024, 7, 15, 2024, 10, 15, 0.0, 0.0, 92.0 / 365.0),
            (2024, 7, 15, 2024, 10, 15, 0.5, 0.0, 91.5 / 365.0),
            (2024, 7, 15, 2024, 10, 15, 0.0, 0.5, 92.5 / 365.0),
            (2024, 7, 15, 2024, 10, 15, 0.5, 0.5, 92.0 / 365.0),
            (2024, 7, 15, 2024, 10, 15, 0.2, 0.8, 92.6 / 365.0),
        ];
        for (y1, m1, d1, y2, m2, d2, df1, df2, expected) in cases {
            let result = act_365_fixed(y1, m1, d1, y2, m2, d2, df1, df2, false);
            assert!(almost_equal(result, expected));
        }
    }

    // ActActExcel tests
    #[test]
    fn test_act_act_excel_basic() {
        let result = act_act_excel(2024, 7, 15, 2024, 10, 15, 0.0, 0.0, false);
        assert!(almost_equal(result, 92.0 / 366.0));
    }

    #[test]
    fn test_act_act_excel_time_fractions() {
        let cases = vec![
            (2024, 7, 15, 2024, 10, 15, 0.0, 0.0, 92.0 / 366.0),
            (2024, 7, 15, 2024, 10, 15, 0.5, 0.0, 91.5 / 366.0),
            (2024, 7, 15, 2024, 10, 15, 0.0, 0.5, 92.5 / 366.0),
            (2024, 7, 15, 2024, 10, 15, 0.5, 0.5, 92.0 / 366.0),
            (2024, 7, 15, 2024, 10, 15, 0.2, 0.8, 92.6 / 366.0),
        ];
        for (y1, m1, d1, y2, m2, d2, df1, df2, expected) in cases {
            let result = act_act_excel(y1, m1, d1, y2, m2, d2, df1, df2, false);
            assert!(almost_equal(result, expected));
        }
    }

    // ActActIsda tests
    #[test]
    fn test_act_act_isda_basic() {
        let result = act_act_isda(2024, 7, 15, 2024, 10, 15, 0.0, 0.0, false);
        assert!(almost_equal(result, 92.0 / 366.0));
    }

    #[test]
    fn test_act_act_isda_time_fractions() {
        let cases = vec![
            (2024, 7, 15, 2024, 10, 15, 0.0, 0.0, 92.0 / 366.0),
            (2024, 7, 15, 2024, 10, 15, 0.5, 0.0, 91.5 / 366.0),
            (2024, 7, 15, 2024, 10, 15, 0.0, 0.5, 92.5 / 366.0),
            (2024, 7, 15, 2024, 10, 15, 0.5, 0.5, 92.0 / 366.0),
            (2024, 7, 15, 2024, 10, 15, 0.2, 0.8, 92.6 / 366.0),
        ];
        for (y1, m1, d1, y2, m2, d2, df1, df2, expected) in cases {
            let result = act_act_isda(y1, m1, d1, y2, m2, d2, df1, df2, false);
            assert!(almost_equal(result, expected));
        }
    }

    // ActActAfb tests
    #[test]
    fn test_act_act_afb_basic() {
        let result = act_act_afb(2024, 7, 15, 2024, 10, 15, 0.0, 0.0, false);
        assert!(almost_equal(result, 92.0 / 365.0));
    }

    #[test]
    fn test_act_act_afb_time_fractions() {
        let cases = vec![
            (2024, 7, 15, 2024, 10, 15, 0.0, 0.0, 92.0 / 365.0),
            (2024, 7, 15, 2024, 10, 15, 0.5, 0.0, 91.5 / 365.0),
            (2024, 7, 15, 2024, 10, 15, 0.0, 0.5, 92.5 / 365.0),
            (2024, 7, 15, 2024, 10, 15, 0.5, 0.5, 92.0 / 365.0),
            (2024, 7, 15, 2024, 10, 15, 0.2, 0.8, 92.6 / 365.0),
        ];
        for (y1, m1, d1, y2, m2, d2, df1, df2, expected) in cases {
            let result = act_act_afb(y1, m1, d1, y2, m2, d2, df1, df2, false);
            assert!(almost_equal(result, expected));
        }
    }

    // Dispatch test
    #[test]
    fn test_dispatch_raw_returns_none() {
        let result = dispatch(DayCountConvention::Raw, 2024, 7, 15, 2024, 10, 15, 0.0, 0.0, false);
        assert!(result.is_none());
    }

    #[test]
    fn test_dispatch_non_raw() {
        let result = dispatch(DayCountConvention::Thirty360Eu, 2024, 7, 15, 2024, 10, 15, 0.0, 0.0, false);
        assert!(result.is_some());
        assert!(almost_equal(result.unwrap(), 90.0 / 360.0));
    }
}
