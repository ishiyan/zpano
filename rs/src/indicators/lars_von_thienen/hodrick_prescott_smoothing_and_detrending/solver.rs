use super::super::harmonic_cycle::model as harmoniccycle;
// Native Go numerical port; Lars von Thienen / WhenToTrade, CC BY 4.0.

pub fn hp_trend(values: Vec<f64>, smoothing: f64) -> Vec<f64> {
    let n: i64 = values.len() as i64;
    if n < 6 {
        panic!("{:?}", "HP requires at least six observations");
    }
    {
        let err = harmoniccycle::validate_real(smoothing, "smoothing", 0.0);
        if err.is_some() {
            panic!("{:?}", err);
        }
    }
    let mut a: Vec<f64> = vec![Default::default(); (n) as usize].clone();
    let mut b: Vec<f64> = vec![Default::default(); (n) as usize].clone();
    let mut c: Vec<f64> = vec![Default::default(); (n) as usize].clone();
    for (__index, value) in values.clone().into_iter().enumerate() {
        let i = __index as i64;
        harmoniccycle::validate_sample(value);
        if (value).is_nan() {
            panic!("{:?}", "HP batch data must be finite");
        }
        a[(i) as usize] = (6.0 * smoothing) + 1.0;
        b[(i) as usize] = (-4 as f64) * smoothing;
        c[(i) as usize] = smoothing;
    }
    a[(0) as usize] = 1.0 + smoothing;
    a[(n - 1) as usize] = 1.0 + smoothing;
    a[(1) as usize] = 1.0 + (5.0 * smoothing);
    a[(n - 2) as usize] = 1.0 + (5.0 * smoothing);
    b[(0) as usize] = (-2 as f64) * smoothing;
    b[(n - 2) as usize] = (-2 as f64) * smoothing;
    b[(n - 1) as usize] = 0.0;
    c[(n - 1) as usize] = 0.0;
    c[(n - 2) as usize] = 0.0;
    let mut h1: f64 = Default::default();
    let mut h2: f64 = Default::default();
    let mut h3: f64 = Default::default();
    let mut h4: f64 = Default::default();
    let mut h5: f64 = Default::default();
    let mut hh2: f64 = Default::default();
    let mut hh3: f64 = Default::default();
    let mut hh5: f64 = Default::default();

    for i in 0..n {
        let z: f64 = harmoniccycle::finite((a[(i) as usize] - (h4 * h1)) - (hh5 * hh2));
        if z <= 0.0 {
            panic!("{:?}", "HP elimination lost positive definiteness");
        }
        let hb: f64 = b[(i) as usize];
        let hh1: f64 = h1;
        h1 = (hb - (h4 * h2)) / z;
        b[(i) as usize] = h1;
        let hc: f64 = c[(i) as usize];
        hh2 = h2;
        h2 = hc / z;
        c[(i) as usize] = h2;
        a[(i) as usize] =
            harmoniccycle::finite(((values[(i) as usize] - (hh3 * hh5)) - (h3 * h4)) / z);
        hh3 = h3;
        h3 = a[(i) as usize];
        h4 = hb - (h5 * hh1);
        hh5 = h5;
        h5 = hc;
    }
    let mut output: Vec<f64> = vec![Default::default(); (n) as usize].clone();
    h2 = 0.0;
    h1 = a[(n - 1) as usize];
    output[(n - 1) as usize] = h1;
    {
        let mut i: i64 = n - 1;
        while i > 0 {
            output[(i - 1) as usize] = harmoniccycle::finite(
                (a[(i - 1) as usize] - (b[(i - 1) as usize] * h1)) - (c[(i - 1) as usize] * h2),
            );
            (h2, h1) = (h1, output[(i - 1) as usize]);
            i -= 1;
        }
    }
    return output.clone();
}
pub fn cycle_data(values: Vec<f64>, smoothing: f64) -> Vec<f64> {
    if smoothing == 0.0 {
        return {
            let mut __values = vec![].clone();
            __values.extend(values.clone());
            __values
        }
        .clone();
    }
    let mut trend: Vec<f64> = hp_trend(values.clone(), smoothing).clone();
    for i in 0..(trend.len() as i64) {
        trend[(i) as usize] = harmoniccycle::finite(values[(i) as usize] - trend[(i) as usize]);
    }
    return trend.clone();
}
