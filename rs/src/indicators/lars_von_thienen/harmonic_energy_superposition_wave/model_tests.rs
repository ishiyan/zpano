pub use super::model::*;
#[cfg(test)]
mod tests {
    use super::*;
    use crate::indicators::lars_von_thienen::annual_cycle_model::{
        model::{AnnualModel, fit_annual_values},
        numerical::default_params,
    };
    mod data {
        include!("testdata.rs");
    }
    use data::testdata::*;
    fn historical_models() -> Vec<AnnualModel> {
        METONIC_LAGS
            .iter()
            .enumerate()
            .map(|(j, &lag)| {
                let input = test_input();
                let values = (0..252)
                    .map(|i| (1.0 + 0.1 * j as f64) * input[(i + 7 * j) % 252])
                    .collect::<Vec<_>>();
                let mut p = default_params();
                p.max_period = 80;
                fit_annual_values(&values, 2021 - lag, &p).unwrap()
            })
            .collect()
    }
    #[test]
    fn prepared_case_0() {
        let models = historical_models();
        let mut reader =
            SuperpositionWaveReader::new(&models, 2021, 252, "none", "strict", false).unwrap();
        for expected in expected_n0_c0_i0_wave() {
            let actual = reader.update().unwrap();
            assert!((actual - expected).abs() <= 1e-10_f64.max(expected.abs() * 5e-15));
        }
        reader.reset();
        assert!((reader.update().unwrap() - expected_n0_c0_i0_wave()[0]).abs() < 1e-10);
    }
    #[test]
    fn prepared_case_1() {
        let models = historical_models();
        let mut reader =
            SuperpositionWaveReader::new(&models, 2021, 252, "rms", "strict", false).unwrap();
        for expected in expected_n1_c0_i0_wave() {
            let actual = reader.update().unwrap();
            assert!((actual - expected).abs() <= 1e-10_f64.max(expected.abs() * 5e-15));
        }
        reader.reset();
        assert!((reader.update().unwrap() - expected_n1_c0_i0_wave()[0]).abs() < 1e-10);
    }
    #[test]
    fn prepared_case_2() {
        let models = historical_models();
        let mut reader =
            SuperpositionWaveReader::new(&models, 2021, 252, "rms", "strict", true).unwrap();
        for expected in expected_n1_c0_i1_wave() {
            let actual = reader.update().unwrap();
            assert!((actual - expected).abs() <= 1e-10_f64.max(expected.abs() * 5e-15));
        }
        reader.reset();
        assert!((reader.update().unwrap() - expected_n1_c0_i1_wave()[0]).abs() < 1e-10);
    }
    #[test]
    fn prepared_case_3() {
        let mut models = historical_models();
        for m in &mut models {
            m.fitted.sample_count = 250;
        }
        let mut reader =
            SuperpositionWaveReader::new(&models, 2021, 252, "rms", "extend", false).unwrap();
        for expected in expected_n1_c1_i0_wave() {
            let actual = reader.update().unwrap();
            assert!((actual - expected).abs() <= 1e-10_f64.max(expected.abs() * 5e-15));
        }
        reader.reset();
        assert!((reader.update().unwrap() - expected_n1_c1_i0_wave()[0]).abs() < 1e-10);
    }
}
