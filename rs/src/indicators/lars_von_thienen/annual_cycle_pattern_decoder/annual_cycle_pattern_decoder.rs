pub use super::super::annual_cycle_model::model::*;
#[cfg(test)]
mod tests {

    use crate::indicators::lars_von_thienen::annual_cycle_model::{
        model::{AnnualWaveReader, fit_annual_values},
        numerical::default_params,
    };
    mod data {
        include!("testdata.rs");
    }
    use data::testdata::*;
    #[test]
    fn prepared_case_0() {
        let mut p = default_params();
        p.candidate_pool = 10;
        p.genuine_threshold = 49.0;
        p.max_components = 7;
        p.max_period = 125;
        p.min_period = 5;
        p.peak_distance = 2;
        p.smoothing = 1600.0;
        let model = fit_annual_values(&test_input(), 2020, &p).unwrap();
        assert!(
            (model.fitted.correlation - expected_p05_p1125_l1600_g49_c10_k7_d2_correlation()[0])
                .abs()
                < 1e-13
        );
        assert_eq!(
            model.fitted.components.len() as f64,
            expected_p05_p1125_l1600_g49_c10_k7_d2_component_count()[0]
        );
        let mut reader = AnnualWaveReader::new(model).unwrap();
        for expected in expected_p05_p1125_l1600_g49_c10_k7_d2_wave() {
            let actual = reader.update();
            assert!((actual - expected).abs() <= 1e-10_f64.max(expected.abs() * 5e-15));
        }
        reader.reset();
        assert!((reader.update() - expected_p05_p1125_l1600_g49_c10_k7_d2_wave()[0]).abs() < 1e-10);
    }
    #[test]
    fn prepared_case_1() {
        let mut p = default_params();
        p.candidate_pool = 10;
        p.genuine_threshold = 49.0;
        p.max_components = 1;
        p.max_period = 125;
        p.min_period = 5;
        p.peak_distance = 2;
        p.smoothing = 1600.0;
        let model = fit_annual_values(&test_input(), 2020, &p).unwrap();
        assert!(
            (model.fitted.correlation - expected_p05_p1125_l1600_g49_c10_k1_d2_correlation()[0])
                .abs()
                < 1e-13
        );
        assert_eq!(
            model.fitted.components.len() as f64,
            expected_p05_p1125_l1600_g49_c10_k1_d2_component_count()[0]
        );
        let mut reader = AnnualWaveReader::new(model).unwrap();
        for expected in expected_p05_p1125_l1600_g49_c10_k1_d2_wave() {
            let actual = reader.update();
            assert!((actual - expected).abs() <= 1e-10_f64.max(expected.abs() * 5e-15));
        }
        reader.reset();
        assert!((reader.update() - expected_p05_p1125_l1600_g49_c10_k1_d2_wave()[0]).abs() < 1e-10);
    }
    #[test]
    fn prepared_case_2() {
        let mut p = default_params();
        p.candidate_pool = 7;
        p.genuine_threshold = 49.0;
        p.max_components = 3;
        p.max_period = 80;
        p.min_period = 5;
        p.peak_distance = 2;
        p.smoothing = 1600.0;
        let model = fit_annual_values(&test_input(), 2020, &p).unwrap();
        assert!(
            (model.fitted.correlation - expected_p05_p180_l1600_g49_c7_k3_d2_correlation()[0])
                .abs()
                < 1e-13
        );
        assert_eq!(
            model.fitted.components.len() as f64,
            expected_p05_p180_l1600_g49_c7_k3_d2_component_count()[0]
        );
        let mut reader = AnnualWaveReader::new(model).unwrap();
        for expected in expected_p05_p180_l1600_g49_c7_k3_d2_wave() {
            let actual = reader.update();
            assert!((actual - expected).abs() <= 1e-10_f64.max(expected.abs() * 5e-15));
        }
        reader.reset();
        assert!((reader.update() - expected_p05_p180_l1600_g49_c7_k3_d2_wave()[0]).abs() < 1e-10);
    }
    #[test]
    fn prepared_case_3() {
        let mut p = default_params();
        p.candidate_pool = 10;
        p.genuine_threshold = 49.0;
        p.max_components = 7;
        p.max_period = 80;
        p.min_period = 5;
        p.peak_distance = 2;
        p.smoothing = 16000.0;
        let model = fit_annual_values(&test_input(), 2020, &p).unwrap();
        assert!(
            (model.fitted.correlation - expected_p05_p180_l16000_g49_c10_k7_d2_correlation()[0])
                .abs()
                < 1e-13
        );
        assert_eq!(
            model.fitted.components.len() as f64,
            expected_p05_p180_l16000_g49_c10_k7_d2_component_count()[0]
        );
        let mut reader = AnnualWaveReader::new(model).unwrap();
        for expected in expected_p05_p180_l16000_g49_c10_k7_d2_wave() {
            let actual = reader.update();
            assert!((actual - expected).abs() <= 1e-10_f64.max(expected.abs() * 5e-15));
        }
        reader.reset();
        assert!((reader.update() - expected_p05_p180_l16000_g49_c10_k7_d2_wave()[0]).abs() < 1e-10);
    }
}
