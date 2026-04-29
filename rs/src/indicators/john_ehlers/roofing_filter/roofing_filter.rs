use std::any::Any;
use std::f64::consts::PI;

use crate::entities::bar::Bar;
use crate::entities::bar_component::{component_value as bar_component_value, BarComponent};
use crate::entities::quote::Quote;
use crate::entities::quote_component::{
    component_value as quote_component_value, QuoteComponent, DEFAULT_QUOTE_COMPONENT,
};
use crate::entities::scalar::Scalar;
use crate::entities::trade::Trade;
use crate::entities::trade_component::{
    component_value as trade_component_value, TradeComponent, DEFAULT_TRADE_COMPONENT,
};
use crate::indicators::core::build_metadata::{build_metadata, OutputText};
use crate::indicators::core::component_triple_mnemonic::component_triple_mnemonic;
use crate::indicators::core::identifier::Identifier;
use crate::indicators::core::indicator::{Indicator, Output};
use crate::indicators::core::metadata::Metadata;

// ---------------------------------------------------------------------------
// Output
// ---------------------------------------------------------------------------

#[repr(u8)]
#[derive(Debug, Clone, Copy, PartialEq, Eq)]
pub enum RoofingFilterOutput {
    Value = 1,
}

// ---------------------------------------------------------------------------
// Params
// ---------------------------------------------------------------------------

/// Parameters to create a Roofing Filter instance.
pub struct RoofingFilterParams {
    /// Shortest cycle period in bars. Must be > 1. Default 10.
    pub shortest_cycle_period: usize,
    /// Longest cycle period in bars. Must be > shortest. Default 48.
    pub longest_cycle_period: usize,
    /// Use 2-pole high-pass filter instead of 1-pole. Default false.
    pub has_two_pole_highpass_filter: bool,
    /// Apply zero-mean filter (only with 1-pole HP). Default false.
    pub has_zero_mean: bool,
    pub bar_component: Option<BarComponent>,
    pub quote_component: Option<QuoteComponent>,
    pub trade_component: Option<TradeComponent>,
}

impl Default for RoofingFilterParams {
    fn default() -> Self {
        Self {
            shortest_cycle_period: 10,
            longest_cycle_period: 48,
            has_two_pole_highpass_filter: false,
            has_zero_mean: false,
            bar_component: None,
            quote_component: None,
            trade_component: None,
        }
    }
}

// ---------------------------------------------------------------------------
// Indicator
// ---------------------------------------------------------------------------

/// Ehlers' Roofing Filter: high-pass + Super Smoother band-pass filter.
pub struct RoofingFilter {
    mnemonic: String,
    description: String,
    hp_coeff1: f64,
    hp_coeff2: f64,
    hp_coeff3: f64,
    ss_coeff1: f64,
    ss_coeff2: f64,
    ss_coeff3: f64,
    has_two_pole: bool,
    has_zero_mean: bool,
    count: usize,
    sample_previous: f64,
    sample_previous2: f64,
    hp_previous: f64,
    hp_previous2: f64,
    ss_previous: f64,
    ss_previous2: f64,
    zm_previous: f64,
    value: f64,
    primed: bool,
    bar_func: fn(&Bar) -> f64,
    quote_func: fn(&Quote) -> f64,
    trade_func: fn(&Trade) -> f64,
}

impl RoofingFilter {
    pub fn new(p: &RoofingFilterParams) -> Result<Self, String> {
        let shortest = p.shortest_cycle_period;
        if shortest < 2 {
            return Err("invalid roofing filter parameters: shortest cycle period should be greater than 1".to_string());
        }

        let longest = p.longest_cycle_period;
        if longest <= shortest {
            return Err("invalid roofing filter parameters: longest cycle period should be greater than shortest".to_string());
        }

        let bc = p.bar_component.unwrap_or(BarComponent::Median);
        let qc = p.quote_component.unwrap_or(DEFAULT_QUOTE_COMPONENT);
        let tc = p.trade_component.unwrap_or(DEFAULT_TRADE_COMPONENT);

        let bar_func = bar_component_value(bc);
        let quote_func = quote_component_value(qc);
        let trade_func = trade_component_value(tc);

        // High-pass filter coefficients.
        let (hp_coeff1, hp_coeff2, hp_coeff3);

        if p.has_two_pole_highpass_filter {
            let angle = std::f64::consts::SQRT_2 / 2.0 * 2.0 * PI / longest as f64;
            let cos_angle = angle.cos();
            let alpha = (angle.sin() + cos_angle - 1.0) / cos_angle;
            let beta = 1.0 - alpha / 2.0;
            hp_coeff1 = beta * beta;
            let beta2 = 1.0 - alpha;
            hp_coeff2 = 2.0 * beta2;
            hp_coeff3 = beta2 * beta2;
        } else {
            let angle = 2.0 * PI / longest as f64;
            let cos_angle = angle.cos();
            let alpha = (angle.sin() + cos_angle - 1.0) / cos_angle;
            hp_coeff1 = 1.0 - alpha / 2.0;
            hp_coeff2 = 1.0 - alpha;
            hp_coeff3 = 0.0;
        }

        // Super Smoother coefficients. Uses 1.414 (not SQRT_2) to match C# reference.
        let beta = 1.414 * PI / shortest as f64;
        let alpha = (-beta).exp();
        let ss_coeff2 = 2.0 * alpha * beta.cos();
        let ss_coeff3 = -alpha * alpha;
        let ss_coeff1 = (1.0 - ss_coeff2 - ss_coeff3) / 2.0;

        // Mnemonic.
        let poles = if p.has_two_pole_highpass_filter { 2 } else { 1 };
        let zm = if p.has_zero_mean && !p.has_two_pole_highpass_filter { "zm" } else { "" };
        let cm = component_triple_mnemonic(bc, qc, tc);
        let mnemonic = format!("roof{poles}hp{zm}({shortest}, {longest}{cm})");
        let description = format!("Roofing Filter {mnemonic}");

        Ok(Self {
            mnemonic,
            description,
            hp_coeff1,
            hp_coeff2,
            hp_coeff3,
            ss_coeff1,
            ss_coeff2,
            ss_coeff3,
            has_two_pole: p.has_two_pole_highpass_filter,
            has_zero_mean: p.has_zero_mean && !p.has_two_pole_highpass_filter,
            count: 0,
            sample_previous: 0.0,
            sample_previous2: 0.0,
            hp_previous: 0.0,
            hp_previous2: 0.0,
            ss_previous: 0.0,
            ss_previous2: 0.0,
            zm_previous: 0.0,
            value: f64::NAN,
            primed: false,
            bar_func,
            quote_func,
            trade_func,
        })
    }

    pub fn update(&mut self, sample: f64) -> f64 {
        if sample.is_nan() {
            return sample;
        }

        if self.has_two_pole {
            self.update_2pole(sample)
        } else {
            self.update_1pole(sample)
        }
    }

    fn update_1pole(&mut self, sample: f64) -> f64 {
        let hp;
        let ss;
        let mut zm = 0.0;

        if self.primed {
            hp = self.hp_coeff1 * (sample - self.sample_previous) + self.hp_coeff2 * self.hp_previous;
            ss = self.ss_coeff1 * (hp + self.hp_previous) + self.ss_coeff2 * self.ss_previous + self.ss_coeff3 * self.ss_previous2;

            if self.has_zero_mean {
                zm = self.hp_coeff1 * (ss - self.ss_previous) + self.hp_coeff2 * self.zm_previous;
                self.value = zm;
            } else {
                self.value = ss;
            }
        } else {
            self.count += 1;

            if self.count == 1 {
                hp = 0.0;
                ss = 0.0;
            } else {
                hp = self.hp_coeff1 * (sample - self.sample_previous) + self.hp_coeff2 * self.hp_previous;
                ss = self.ss_coeff1 * (hp + self.hp_previous) + self.ss_coeff2 * self.ss_previous + self.ss_coeff3 * self.ss_previous2;

                if self.has_zero_mean {
                    zm = self.hp_coeff1 * (ss - self.ss_previous) + self.hp_coeff2 * self.zm_previous;
                    if self.count == 5 {
                        self.primed = true;
                        self.value = zm;
                    }
                } else if self.count == 4 {
                    self.primed = true;
                    self.value = ss;
                }
            }
        }

        self.sample_previous = sample;
        self.hp_previous = hp;
        self.ss_previous2 = self.ss_previous;
        self.ss_previous = ss;

        if self.has_zero_mean {
            self.zm_previous = zm;
        }

        self.value
    }

    fn update_2pole(&mut self, sample: f64) -> f64 {
        let hp;
        let ss;

        if self.primed {
            hp = self.hp_coeff1 * (sample - 2.0 * self.sample_previous + self.sample_previous2)
                + self.hp_coeff2 * self.hp_previous - self.hp_coeff3 * self.hp_previous2;
            ss = self.ss_coeff1 * (hp + self.hp_previous) + self.ss_coeff2 * self.ss_previous + self.ss_coeff3 * self.ss_previous2;
            self.value = ss;
        } else {
            self.count += 1;

            if self.count < 4 {
                hp = 0.0;
                ss = 0.0;
            } else {
                hp = self.hp_coeff1 * (sample - 2.0 * self.sample_previous + self.sample_previous2)
                    + self.hp_coeff2 * self.hp_previous - self.hp_coeff3 * self.hp_previous2;
                ss = self.ss_coeff1 * (hp + self.hp_previous) + self.ss_coeff2 * self.ss_previous + self.ss_coeff3 * self.ss_previous2;

                if self.count == 5 {
                    self.primed = true;
                    self.value = ss;
                }
            }
        }

        self.sample_previous2 = self.sample_previous;
        self.sample_previous = sample;
        self.hp_previous2 = self.hp_previous;
        self.hp_previous = hp;
        self.ss_previous2 = self.ss_previous;
        self.ss_previous = ss;

        self.value
    }

    fn update_entity(&mut self, time: i64, sample: f64) -> Output {
        let v = self.update(sample);
        vec![Box::new(Scalar::new(time, v)) as Box<dyn Any>]
    }
}

impl Indicator for RoofingFilter {
    fn is_primed(&self) -> bool {
        self.primed
    }

    fn metadata(&self) -> Metadata {
        build_metadata(
            Identifier::RoofingFilter,
            &self.mnemonic,
            &self.description,
            &[OutputText {
                mnemonic: self.mnemonic.clone(),
                description: self.description.clone(),
            }],
        )
    }

    fn update_scalar(&mut self, sample: &Scalar) -> Output {
        self.update_entity(sample.time, sample.value)
    }

    fn update_bar(&mut self, sample: &Bar) -> Output {
        let v = (self.bar_func)(sample);
        self.update_entity(sample.time, v)
    }

    fn update_quote(&mut self, sample: &Quote) -> Output {
        let v = (self.quote_func)(sample);
        self.update_entity(sample.time, v)
    }

    fn update_trade(&mut self, sample: &Trade) -> Output {
        let v = (self.trade_func)(sample);
        self.update_entity(sample.time, v)
    }
}

// ===========================================================================
// Tests
// ===========================================================================

#[cfg(test)]
mod tests {
    use super::*;

    const TOLERANCE: f64 = 0.5;
    const SKIP_ROWS: usize = 30;

    fn input() -> Vec<f64> {
        vec![
            1065.25, 1065.25, 1063.75, 1059.25, 1059.25, 1057.75, 1054.0, 1056.25, 1058.5, 1059.5,
            1064.75, 1063.0, 1062.5, 1065.0, 1061.5, 1058.25, 1058.25, 1061.75, 1062.0, 1061.25,
            1062.5, 1066.5, 1066.5, 1069.25, 1074.75, 1075.0, 1076.0, 1078.0, 1079.25, 1079.75,
            1078.0, 1078.75, 1078.25, 1076.5, 1075.75, 1075.75, 1075.0, 1073.25, 1071.0, 1083.0,
            1082.25, 1084.0, 1085.75, 1085.25, 1085.75, 1087.25, 1089.0, 1089.0, 1090.0, 1095.0,
            1097.25, 1097.25, 1099.0, 1098.25, 1093.75, 1095.0, 1097.25, 1099.25, 1097.5, 1096.0,
            1095.0, 1094.0, 1095.75, 1095.75, 1093.75, 1100.5, 1102.25, 1102.0, 1102.75, 1105.75,
            1108.25, 1109.5, 1107.25, 1102.5, 1104.75, 1099.25, 1102.75, 1099.5, 1096.75, 1098.25,
            1095.25, 1097.0, 1097.75, 1100.5, 1099.5, 1101.75, 1101.75, 1102.75, 1099.75, 1097.0,
            1100.75, 1105.75, 1104.5, 1108.5, 1111.25, 1112.25, 1110.0, 1109.75, 1108.25, 1106.0,
        ]
    }

    fn expected_1pole() -> Vec<f64> {
        vec![
            0.0, 0.0, 0.0, -0.53, -1.62, -2.72, -4.03, -5.09, -5.05, -4.09,
            -2.20, -0.05, 1.29, 2.14, 2.39, 1.46, -0.05, -0.90, -0.80, -0.41,
            0.03, 0.99, 2.30, 3.60, 5.39, 7.33, 8.69, 9.52, 10.00, 10.11,
            9.59, 8.58, 7.46, 6.12, 4.61, 3.26, 2.16, 1.12, -0.11, 0.12,
            2.14, 4.27, 6.08, 7.22, 7.54, 7.48, 7.46, 7.43, 7.29, 7.64,
            8.69, 9.68, 10.26, 10.32, 9.23, 7.38, 5.98, 5.47, 5.30, 4.74,
            3.77, 2.58, 1.66, 1.28, 0.92, 1.21, 2.62, 4.12, 5.14, 5.97,
            6.95, 7.94, 8.26, 7.16, 5.36, 3.27, 1.36, 0.07, -1.34, -2.48,
            -3.29, -3.79, -3.61, -2.72, -1.53, -0.40, 0.67, 1.49, 1.70, 0.89,
            0.04, 0.47, 1.66, 3.05, 4.81, 6.48, 7.28, 7.00, 5.99, 4.62,
        ]
    }

    fn expected_1pole_zm() -> Vec<f64> {
        vec![
            0.0, 0.0, 0.0, -0.50, -1.46, -2.31, -3.26, -3.85, -3.34, -2.02,
            -0.01, 2.01, 3.02, 3.45, 3.26, 1.99, 0.33, -0.52, -0.35, 0.05,
            0.46, 1.31, 2.37, 3.30, 4.57, 5.84, 6.39, 6.38, 6.05, 5.41,
            4.26, 2.79, 1.39, -0.04, -1.45, -2.54, -3.26, -3.83, -4.51, -3.74,
            -1.39, 0.78, 2.39, 3.16, 3.07, 2.64, 2.29, 1.98, 1.61, 1.74,
            2.51, 3.13, 3.29, 2.94, 1.56, -0.37, -1.64, -1.91, -1.84, -2.14,
            -2.79, -3.56, -3.98, -3.85, -3.72, -2.99, -1.29, 0.27, 1.19, 1.83,
            2.53, 3.14, 3.06, 1.65, -0.25, -2.17, -3.70, -4.46, -5.23, -5.66,
            -5.72, -5.49, -4.65, -3.23, -1.72, -0.45, 0.61, 1.30, 1.34, 0.42,
            -0.43, 0.02, 1.14, 2.30, 3.67, 4.79, 4.95, 4.08, 2.63, 0.80,
        ]
    }

    fn expected_2pole() -> Vec<f64> {
        vec![
            0.0, 0.0, 0.0, -0.03, -0.10, -0.17, -0.28, -0.37, -0.38, -0.27,
            0.03, 0.52, 1.13, 1.85, 2.62, 3.37, 4.04, 4.69, 5.35, 6.00,
            6.63, 7.29, 7.99, 8.71, 9.52, 10.42, 11.34, 12.27, 13.19, 14.07,
            14.85, 15.49, 15.99, 16.32, 16.45, 16.40, 16.19, 15.82, 15.27, 14.69,
            14.20, 13.79, 13.45, 13.16, 12.90, 12.66, 12.45, 12.26, 12.07, 11.93,
            11.88, 11.88, 11.91, 11.94, 11.88, 11.69, 11.41, 11.11, 10.76, 10.33,
            9.80, 9.17, 8.47, 7.75, 6.99, 6.26, 5.64, 5.14, 4.71, 4.37,
            4.16, 4.07, 4.02, 3.93, 3.77, 3.50, 3.13, 2.68, 2.13, 1.49,
            0.79, 0.05, -0.67, -1.31, -1.86, -2.31, -2.65, -2.89, -3.06, -3.24,
            -3.40, -3.46, -3.39, -3.21, -2.88, -2.41, -1.89, -1.37, -0.91, -0.51,
        ]
    }

    fn create_1pole() -> RoofingFilter {
        RoofingFilter::new(&RoofingFilterParams {
            shortest_cycle_period: 10,
            longest_cycle_period: 48,
            ..Default::default()
        }).unwrap()
    }

    fn create_1pole_zm() -> RoofingFilter {
        RoofingFilter::new(&RoofingFilterParams {
            shortest_cycle_period: 10,
            longest_cycle_period: 48,
            has_zero_mean: true,
            ..Default::default()
        }).unwrap()
    }

    fn create_2pole() -> RoofingFilter {
        RoofingFilter::new(&RoofingFilterParams {
            shortest_cycle_period: 40,
            longest_cycle_period: 80,
            has_two_pole_highpass_filter: true,
            ..Default::default()
        }).unwrap()
    }

    #[test]
    fn test_update_1pole() {
        let mut rf = create_1pole();
        let inp = input();
        let exp = expected_1pole();

        for i in 0..inp.len() {
            let act = rf.update(inp[i]);
            if i < 3 {
                assert!(act.is_nan(), "[{i}] expected NaN");
                continue;
            }
            if i < SKIP_ROWS { continue; }
            assert!(
                (act - exp[i]).abs() < TOLERANCE,
                "[{i}] expected {}, actual {act}", exp[i]
            );
        }

        assert!(rf.update(f64::NAN).is_nan());
    }

    #[test]
    fn test_update_1pole_zm() {
        let mut rf = create_1pole_zm();
        let inp = input();
        let exp = expected_1pole_zm();

        for i in 0..inp.len() {
            let act = rf.update(inp[i]);
            if i < 4 {
                assert!(act.is_nan(), "[{i}] expected NaN");
                continue;
            }
            if i < SKIP_ROWS { continue; }
            assert!(
                (act - exp[i]).abs() < TOLERANCE,
                "[{i}] expected {}, actual {act}", exp[i]
            );
        }
    }

    #[test]
    fn test_update_2pole() {
        let mut rf = create_2pole();
        let inp = input();
        let exp = expected_2pole();

        for i in 0..inp.len() {
            let act = rf.update(inp[i]);
            if i < 4 {
                assert!(act.is_nan(), "[{i}] expected NaN");
                continue;
            }
            if i < SKIP_ROWS { continue; }
            assert!(
                (act - exp[i]).abs() < TOLERANCE,
                "[{i}] expected {}, actual {act}", exp[i]
            );
        }
    }

    #[test]
    fn test_is_primed_1pole() {
        let mut rf = create_1pole();
        let inp = input();
        assert!(!rf.is_primed());
        for i in 0..3 {
            rf.update(inp[i]);
            assert!(!rf.is_primed());
        }
        rf.update(inp[3]);
        assert!(rf.is_primed());
    }

    #[test]
    fn test_is_primed_1pole_zm() {
        let mut rf = create_1pole_zm();
        let inp = input();
        for i in 0..4 {
            rf.update(inp[i]);
            assert!(!rf.is_primed());
        }
        rf.update(inp[4]);
        assert!(rf.is_primed());
    }

    #[test]
    fn test_is_primed_2pole() {
        let mut rf = create_2pole();
        let inp = input();
        for i in 0..4 {
            rf.update(inp[i]);
            assert!(!rf.is_primed());
        }
        rf.update(inp[4]);
        assert!(rf.is_primed());
    }

    #[test]
    fn test_update_entity() {
        let time: i64 = 1617235200;
        let inp = 100.0;

        let mut rf = create_1pole();
        for _ in 0..4 { rf.update(inp); }

        let s = Scalar::new(time, inp);
        let out = rf.update_scalar(&s);
        assert_eq!(out.len(), 1);
        let s0 = out[0].downcast_ref::<Scalar>().unwrap();
        assert_eq!(s0.time, time);
        assert!(!s0.value.is_nan());
    }

    #[test]
    fn test_metadata() {
        let rf = create_1pole();
        let m = rf.metadata();
        assert_eq!(m.identifier, Identifier::RoofingFilter);
        assert_eq!(m.mnemonic, "roof1hp(10, 48, hl/2)");
        assert_eq!(m.description, "Roofing Filter roof1hp(10, 48, hl/2)");
        assert_eq!(m.outputs.len(), 1);
        assert_eq!(m.outputs[0].kind, RoofingFilterOutput::Value as i32);
        assert_eq!(m.outputs[0].mnemonic, "roof1hp(10, 48, hl/2)");
    }

    #[test]
    fn test_new_errors() {
        // shortest < 2
        assert!(RoofingFilter::new(&RoofingFilterParams {
            shortest_cycle_period: 1,
            longest_cycle_period: 48,
            ..Default::default()
        }).is_err());

        // longest <= shortest
        assert!(RoofingFilter::new(&RoofingFilterParams {
            shortest_cycle_period: 10,
            longest_cycle_period: 10,
            ..Default::default()
        }).is_err());

        // valid
        assert!(RoofingFilter::new(&RoofingFilterParams::default()).is_ok());
    }

    #[test]
    fn test_mnemonics() {
        // 2-pole
        let rf = RoofingFilter::new(&RoofingFilterParams {
            shortest_cycle_period: 10,
            longest_cycle_period: 48,
            has_two_pole_highpass_filter: true,
            ..Default::default()
        }).unwrap();
        assert_eq!(rf.mnemonic, "roof2hp(10, 48, hl/2)");

        // zero-mean
        let rf = RoofingFilter::new(&RoofingFilterParams {
            shortest_cycle_period: 10,
            longest_cycle_period: 48,
            has_zero_mean: true,
            ..Default::default()
        }).unwrap();
        assert_eq!(rf.mnemonic, "roof1hpzm(10, 48, hl/2)");

        // non-default bar component
        let rf = RoofingFilter::new(&RoofingFilterParams {
            shortest_cycle_period: 10,
            longest_cycle_period: 48,
            bar_component: Some(BarComponent::Open),
            ..Default::default()
        }).unwrap();
        assert_eq!(rf.mnemonic, "roof1hp(10, 48, o)");
    }
}
