#[cfg(test)]
pub mod testdata {

    pub fn test_book_input() -> Vec<f64> {
        vec![
            101.0313, 101.0313, 101.1250, 101.9687, 102.7813,
            103.0000, 102.9687, 103.0625, 102.9375, 102.7188,
            102.7500, 102.9063, 102.9687,
        ]
    }

    pub fn test_book_expected() -> Vec<f64> {
        vec![
            f64::NAN, f64::NAN, f64::NAN, f64::NAN, f64::NAN,
            f64::NAN, f64::NAN, f64::NAN, f64::NAN, f64::NAN,
            69.61963786608334, 71.42857142857143, 71.08377992828775,
        ]
    }
}
