use std::f64::consts::{PI, SQRT_2};

use crate::entities::bar::Bar;
use crate::entities::bar_component::{component_value as bar_component_value, BarComponent};
use crate::entities::quote::Quote;
use crate::entities::quote_component::{component_value as quote_component_value, QuoteComponent, DEFAULT_QUOTE_COMPONENT};
use crate::entities::scalar::Scalar;
use crate::entities::trade::Trade;
use crate::entities::trade_component::{component_value as trade_component_value, TradeComponent, DEFAULT_TRADE_COMPONENT};
use crate::indicators::core::build_metadata::{build_metadata, OutputText};
use crate::indicators::core::component_triple_mnemonic::component_triple_mnemonic;
use crate::indicators::core::identifier::Identifier;
use crate::indicators::core::indicator::{Indicator, Output};
use crate::indicators::core::line_indicator::LineIndicator;
use crate::indicators::core::metadata::Metadata;

// ---------------------------------------------------------------------------
// Params
// ---------------------------------------------------------------------------

/// Parameters for the Super Smoother indicator.
pub struct SuperSmootherParams {
    /// The shortest cycle period in bars. Must be >= 2. Default is 10.
    pub shortest_cycle_period: i64,
    /// Bar component to extract. `None` means use default (Median).
    pub bar_component: Option<BarComponent>,
    /// Quote component to extract. `None` means use default (Mid).
    pub quote_component: Option<QuoteComponent>,
    /// Trade component to extract. `None` means use default (Price).
    pub trade_component: Option<TradeComponent>,
}

impl Default for SuperSmootherParams {
    fn default() -> Self {
        Self {
            shortest_cycle_period: 10,
            bar_component: None,
            quote_component: None,
            trade_component: None,
        }
    }
}

// ---------------------------------------------------------------------------
// Output
// ---------------------------------------------------------------------------

/// Enumerates the outputs of the Super Smoother indicator.
#[derive(Debug, Clone, Copy, PartialEq, Eq)]
#[repr(u8)]
pub enum SuperSmootherOutput {
    /// The scalar value of the super smoother.
    Value = 1,
}

// ---------------------------------------------------------------------------
// Indicator
// ---------------------------------------------------------------------------

/// Ehlers' two-pole Super Smoother filter.
///
/// Given the shortest cycle period lambda, the filter attenuates all cycle
/// periods shorter than lambda.
///
///   beta  = sqrt(2) * pi / lambda
///   alpha = exp(-beta)
///   g2    = 2 * alpha * cos(beta)
///   g3    = -alpha^2
///   g1    = (1 - g2 - g3) / 2
///
///   SS_i  = g1*(x_i + x_{i-1}) + g2*SS_{i-1} + g3*SS_{i-2}
///
/// The indicator is not primed during the first 2 updates.
pub struct SuperSmoother {
    line: LineIndicator,
    coeff1: f64,
    coeff2: f64,
    coeff3: f64,
    count: i64,
    sample_previous: f64,
    filter_previous: f64,
    filter_previous2: f64,
    value: f64,
    primed: bool,
}

impl SuperSmoother {
    /// Creates a new Super Smoother from the supplied parameters.
    pub fn new(params: &SuperSmootherParams) -> Result<Self, String> {
        const INVALID: &str = "invalid super smoother parameters";

        let period = params.shortest_cycle_period;
        if period < 2 {
            return Err(format!("{}: shortest cycle period should be greater than 1", INVALID));
        }

        // Default bar component is Median (not Close).
        let bc = params.bar_component.unwrap_or(BarComponent::Median);
        let qc = params.quote_component.unwrap_or(DEFAULT_QUOTE_COMPONENT);
        let tc = params.trade_component.unwrap_or(DEFAULT_TRADE_COMPONENT);

        let bar_func = bar_component_value(bc);
        let quote_func = quote_component_value(qc);
        let trade_func = trade_component_value(tc);

        // Calculate coefficients.
        let beta = SQRT_2 * PI / period as f64;
        let alpha = (-beta).exp();
        let gamma2 = 2.0 * alpha * beta.cos();
        let gamma3 = -alpha * alpha;
        let gamma1 = (1.0 - gamma2 - gamma3) / 2.0;

        let mnemonic = format!("ss({}{})", period, component_triple_mnemonic(bc, qc, tc));
        let description = format!("Super Smoother {}", mnemonic);
        let line = LineIndicator::new(mnemonic, description, bar_func, quote_func, trade_func);

        Ok(Self {
            line,
            coeff1: gamma1,
            coeff2: gamma2,
            coeff3: gamma3,
            count: 0,
            sample_previous: 0.0,
            filter_previous: 0.0,
            filter_previous2: 0.0,
            value: f64::NAN,
            primed: false,
        })
    }

    /// Core update logic. Returns the filter value or NaN if not yet primed.
    pub fn update(&mut self, sample: f64) -> f64 {
        if sample.is_nan() {
            return sample;
        }

        if self.primed {
            let filter = self.coeff1 * (sample + self.sample_previous)
                + self.coeff2 * self.filter_previous
                + self.coeff3 * self.filter_previous2;
            self.value = filter;
            self.sample_previous = sample;
            self.filter_previous2 = self.filter_previous;
            self.filter_previous = filter;
            return self.value;
        }

        self.count += 1;

        if self.count == 1 {
            self.sample_previous = sample;
            self.filter_previous = sample;
            self.filter_previous2 = sample;
        }

        let filter = self.coeff1 * (sample + self.sample_previous)
            + self.coeff2 * self.filter_previous
            + self.coeff3 * self.filter_previous2;

        if self.count == 3 {
            self.primed = true;
            self.value = filter;
        }

        self.sample_previous = sample;
        self.filter_previous2 = self.filter_previous;
        self.filter_previous = filter;

        self.value
    }
}

impl Indicator for SuperSmoother {
    fn is_primed(&self) -> bool {
        self.primed
    }

    fn metadata(&self) -> Metadata {
        build_metadata(
            Identifier::SuperSmoother,
            &self.line.mnemonic,
            &self.line.description,
            &[OutputText {
                mnemonic: self.line.mnemonic.clone(),
                description: self.line.description.clone(),
            }],
        )
    }

    fn update_scalar(&mut self, sample: &Scalar) -> Output {
        let value = self.update(sample.value);
        vec![Box::new(Scalar::new(sample.time, value))]
    }

    fn update_bar(&mut self, sample: &Bar) -> Output {
        let sample_value = (self.line.bar_func)(sample);
        let value = self.update(sample_value);
        vec![Box::new(Scalar::new(sample.time, value))]
    }

    fn update_quote(&mut self, sample: &Quote) -> Output {
        let sample_value = (self.line.quote_func)(sample);
        let value = self.update(sample_value);
        vec![Box::new(Scalar::new(sample.time, value))]
    }

    fn update_trade(&mut self, sample: &Trade) -> Output {
        let sample_value = (self.line.trade_func)(sample);
        let value = self.update(sample_value);
        vec![Box::new(Scalar::new(sample.time, value))]
    }
}

// ---------------------------------------------------------------------------
// Tests
// ---------------------------------------------------------------------------

#[cfg(test)]
mod tests {
    use super::*;
    use crate::entities::bar_component::BarComponent;
    use crate::indicators::core::outputs::shape::Shape;

    #[allow(clippy::excessive_precision)]
    fn test_input() -> Vec<f64> {
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
            1108.5, 1106.75, 1108.0, 1106.5, 1105.25, 1104.25, 1102.0, 1102.5, 1103.25, 1104.0,
            1104.0, 1102.5, 1101.0, 1099.5, 1100.0, 1100.25, 1103.0, 1103.0, 1104.5, 1108.25,
            1110.75, 1107.25, 1107.5, 1105.25, 1103.0, 1101.75, 1101.75, 1096.5, 1099.5, 1093.75,
            1094.5, 1090.75, 1093.25, 1091.75, 1094.75, 1093.75, 1092.25, 1091.5, 1092.5, 1089.75,
            1089.75, 1090.0, 1087.75, 1093.75, 1095.25, 1097.0, 1094.75, 1093.0, 1094.25, 1092.25,
            1093.0, 1095.25, 1096.0, 1094.25, 1092.25, 1094.5, 1092.75, 1094.5, 1098.25, 1097.75,
            1097.5, 1097.5, 1098.25, 1097.75, 1097.5, 1101.5, 1102.75, 1104.5, 1103.25, 1100.0,
            1101.5, 1097.0, 1098.75, 1100.0, 1098.25, 1102.0, 1103.75, 1103.75, 1102.75, 1101.75,
            1100.5, 1096.25, 1100.5, 1104.0, 1105.25, 1104.25, 1102.25, 1104.0, 1103.25, 1104.25,
            1100.0, 1097.75, 1099.5, 1102.0, 1104.75, 1102.75, 1104.5, 1104.25, 1105.75, 1107.5,
            1107.0, 1106.5, 1108.0, 1107.75, 1106.5, 1107.0, 1108.5, 1109.25, 1123.75, 1124.25,
            1123.75, 1123.0, 1121.25, 1120.25, 1118.75, 1119.0, 1119.75, 1109.75, 1109.75, 1115.75,
            1114.75, 1120.5, 1119.25, 1118.25, 1121.5, 1121.25, 1120.0, 1120.5, 1121.25, 1123.0,
            1122.25, 1122.0, 1122.0, 1121.5, 1123.25, 1123.0, 1126.0, 1127.25, 1129.5, 1129.5,
            1129.25, 1130.0, 1130.5, 1130.25, 1128.75, 1128.0, 1132.5, 1129.25, 1122.5, 1119.25,
            1120.25, 1120.0, 1121.75, 1122.0, 1120.75, 1120.25, 1116.0, 1119.75, 1119.25, 1122.75,
            1114.25, 1117.0, 1117.75, 1120.0, 1119.0, 1122.0, 1120.5, 1118.25, 1119.0, 1120.25,
            1119.5, 1120.0, 1119.5, 1113.0, 1112.75, 1114.5, 1116.0, 1116.25, 1114.25, 1114.75,
            1115.25, 1115.75, 1116.25, 1114.75, 1113.5, 1115.75, 1121.25, 1120.25, 1119.25, 1118.25,
            1118.25, 1119.25, 1117.5, 1117.25, 1116.5, 1117.0, 1117.0, 1116.75, 1116.75, 1120.25,
            1119.75, 1120.25, 1117.5, 1117.25, 1115.5, 1110.75, 1112.5, 1111.25, 1112.75, 1117.75,
            1119.75, 1118.75, 1111.0, 1113.0, 1114.0, 1114.25, 1114.0, 1113.75, 1114.25, 1112.75,
            1114.5, 1115.25, 1116.5, 1114.75, 1113.0, 1117.75, 1119.5, 1120.0, 1115.25, 1115.25,
            1115.5, 1114.25, 1114.25, 1112.75, 1110.75, 1104.75, 1099.25, 1102.25, 1095.25, 1095.25,
            1094.25, 1095.75, 1095.5, 1094.75, 1092.25, 1086.25, 1084.75, 1083.75, 1083.25, 1088.75,
            1089.75, 1089.5, 1093.5, 1093.0, 1096.5, 1095.75, 1095.75, 1096.25, 1094.0, 1096.75,
            1099.75, 1102.0, 1103.5, 1102.5, 1107.75, 1109.25, 1111.5, 1111.0, 1113.5, 1115.0,
            1114.25, 1115.0, 1115.0, 1114.5, 1117.75, 1118.5, 1120.5, 1117.5, 1117.0, 1115.75,
            1115.0, 1114.5, 1114.0, 1116.25, 1114.0, 1113.75, 1115.75, 1115.75, 1118.0, 1116.0,
            1095.75, 1097.75, 1102.25, 1098.25, 1100.0, 1098.0, 1097.25, 1093.25, 1093.5, 1091.0,
            1098.0, 1100.5, 1096.25, 1104.75, 1102.25, 1097.5, 1092.75, 1083.75, 1089.0, 1090.5,
            1093.25, 1094.25, 1096.75, 1091.5, 1088.25, 1087.5, 1083.5, 1077.75, 1073.0, 1065.0,
            1073.25, 1068.75, 1066.25, 1057.25, 1045.5, 1046.0, 1039.75, 1014.75, 1014.75, 997.5,
            1026.25, 1020.25, 1034.5, 1030.75, 1035.75, 1045.75, 1039.75, 1041.25, 1046.25, 1057.25,
            1062.75, 1067.0, 1055.75, 1057.75, 1059.75, 1061.75, 1065.5, 1064.25, 1064.0, 1063.25,
            1061.0, 1059.75, 1057.0, 1054.25, 1047.5, 1059.25, 1055.25, 1053.5, 1052.25, 1057.25,
            1058.25, 1055.5, 1056.75, 1057.75, 1062.75, 1059.75, 1061.0, 1059.25, 1078.25, 1076.5,
            1075.5, 1077.75, 1078.75, 1079.75, 1080.25, 1077.25, 1078.75, 1080.5, 1079.75, 1081.25,
            1084.25, 1083.75, 1077.75, 1078.5, 1080.0, 1082.25, 1081.0, 1082.75, 1082.75, 1080.5,
            1082.25, 1083.25, 1085.75, 1086.25, 1084.75, 1085.5, 1083.0, 1085.5, 1086.5, 1091.75,
        ]
    }

    #[allow(clippy::excessive_precision)]
    fn test_expected() -> Vec<f64> {
        vec![
            0.0, 0.0, 0.0, 268.7, 579.33, 828.39, 988.41, 1071.11, 1101.65, 1103.41,
            1093.69, 1082.14, 1072.48, 1066.3, 1062.99, 1060.83, 1059.3, 1058.84, 1059.42, 1060.22,
            1060.97, 1062.17, 1063.76, 1065.45, 1067.81, 1070.57, 1072.95, 1074.95, 1076.71, 1078.14,
            1078.91, 1079.09, 1079.02, 1078.57, 1077.77, 1076.93, 1076.19, 1075.36, 1074.2, 1074.43,
            1076.6, 1079.15, 1081.64, 1083.65, 1084.94, 1085.87, 1086.82, 1087.76, 1088.59, 1089.92,
            1092.04, 1094.23, 1096.12, 1097.53, 1097.72, 1096.95, 1096.43, 1096.67, 1097.2, 1097.31,
            1096.89, 1096.12, 1095.48, 1095.29, 1095.07, 1095.5, 1097.17, 1099.11, 1100.73, 1102.29,
            1104.12, 1106.08, 1107.47, 1107.38, 1106.39, 1104.88, 1103.27, 1102.07, 1100.57, 1099.18,
            1097.99, 1097.03, 1096.72, 1097.2, 1098.11, 1099.12, 1100.2, 1101.16, 1101.58, 1100.94,
            1100.16, 1100.62, 1101.95, 1103.65, 1105.92, 1108.33, 1110.03, 1110.69, 1110.53, 1109.61,
            1108.63, 1107.97, 1107.55, 1107.3, 1106.84, 1106.12, 1105.07, 1103.92, 1103.18, 1102.99,
            1103.17, 1103.26, 1102.92, 1102.1, 1101.17, 1100.52, 1100.54, 1101.16, 1102.08, 1103.54,
            1105.65, 1107.37, 1108.07, 1107.94, 1106.91, 1105.34, 1103.79, 1101.97, 1100.22, 1098.59,
            1096.79, 1094.99, 1093.5, 1092.63, 1092.43, 1092.81, 1093.01, 1092.81, 1092.52, 1092.05,
            1091.27, 1090.6, 1089.89, 1089.81, 1090.97, 1092.75, 1094.27, 1094.8, 1094.72, 1094.31,
            1093.72, 1093.58, 1094.04, 1094.5, 1094.38, 1094.07, 1093.83, 1093.68, 1094.3, 1095.49,
            1096.52, 1097.19, 1097.64, 1097.92, 1097.96, 1098.36, 1099.48, 1100.99, 1102.34, 1102.72,
            1102.37, 1101.44, 1100.15, 1099.43, 1099.05, 1099.17, 1100.16, 1101.47, 1102.46, 1102.82,
            1102.53, 1101.36, 1100.13, 1100.16, 1101.3, 1102.64, 1103.35, 1103.58, 1103.69, 1103.75,
            1103.36, 1102.07, 1100.66, 1100.11, 1100.71, 1101.72, 1102.62, 1103.44, 1104.17, 1105.09,
            1106.02, 1106.58, 1106.98, 1107.38, 1107.47, 1107.33, 1107.38, 1107.78, 1110.15, 1114.63,
            1118.85, 1121.73, 1123.01, 1122.97, 1122.07, 1120.89, 1120.03, 1118.33, 1115.46, 1113.6,
            1113.25, 1114.21, 1116.04, 1117.48, 1118.68, 1119.85, 1120.53, 1120.74, 1120.86, 1121.23,
            1121.73, 1122.04, 1122.16, 1122.1, 1122.15, 1122.41, 1123.05, 1124.22, 1125.75, 1127.33,
            1128.5, 1129.26, 1129.83, 1130.2, 1130.17, 1129.71, 1129.65, 1129.94, 1129.03, 1126.59,
            1123.86, 1121.79, 1120.7, 1120.56, 1120.7, 1120.71, 1120.06, 1119.24, 1118.97, 1119.37,
            1119.32, 1118.36, 1117.72, 1117.75, 1118.2, 1118.97, 1119.86, 1120.11, 1119.83, 1119.67,
            1119.65, 1119.67, 1119.7, 1118.84, 1116.97, 1115.36, 1114.67, 1114.75, 1114.91, 1114.87,
            1114.89, 1115.05, 1115.36, 1115.52, 1115.23, 1114.96, 1115.75, 1117.33, 1118.6, 1119.16,
            1119.16, 1119.05, 1118.84, 1118.38, 1117.81, 1117.31, 1117.02, 1116.87, 1116.77, 1117.17,
            1118.05, 1118.91, 1119.25, 1118.92, 1118.14, 1116.55, 1114.65, 1113.16, 1112.26, 1112.64,
            1114.35, 1116.29, 1116.73, 1115.71, 1114.74, 1114.18, 1113.94, 1113.82, 1113.82, 1113.74,
            1113.68, 1113.95, 1114.55, 1115.07, 1114.98, 1115.04, 1115.98, 1117.32, 1117.94, 1117.52,
            1116.8, 1116.02, 1115.25, 1114.49, 1113.48, 1111.62, 1108.42, 1105.16, 1102.2, 1099.22,
            1096.86, 1095.42, 1094.88, 1094.72, 1094.35, 1092.9, 1090.43, 1087.85, 1085.69, 1084.88,
            1085.65, 1086.98, 1088.67, 1090.52, 1092.35, 1094.06, 1095.19, 1095.86, 1095.95, 1095.84,
            1096.41, 1097.77, 1099.59, 1101.2, 1102.86, 1104.97, 1107.2, 1109.15, 1110.73, 1112.27,
            1113.5, 1114.29, 1114.8, 1114.99, 1115.36, 1116.21, 1117.39, 1118.29, 1118.39, 1117.92,
            1117.09, 1116.15, 1115.28, 1114.89, 1114.78, 1114.51, 1114.46, 1114.77, 1115.42, 1116.09,
            1113.78, 1108.52, 1104.2, 1101.42, 1099.7, 1098.81, 1098.15, 1097.14, 1095.77, 1094.32,
            1093.77, 1094.93, 1096.28, 1097.9, 1099.99, 1100.82, 1099.72, 1096.36, 1092.45, 1090.16,
            1089.65, 1090.48, 1092.09, 1093.27, 1092.89, 1091.47, 1089.37, 1086.3, 1082.27, 1077.25,
            1073.13, 1070.9, 1069.12, 1066.52, 1061.62, 1055.59, 1049.89, 1041.81, 1031.64, 1021.0,
            1014.31, 1013.82, 1017.05, 1022.32, 1027.26, 1032.7, 1037.48, 1040.22, 1042.23, 1045.47,
            1050.48, 1056.19, 1059.85, 1060.57, 1060.4, 1060.42, 1061.24, 1062.5, 1063.43, 1063.86,
            1063.6, 1062.67, 1061.21, 1059.19, 1056.26, 1054.32, 1054.27, 1054.27, 1053.92, 1053.99,
            1054.97, 1055.85, 1056.29, 1056.71, 1057.78, 1059.1, 1059.96, 1060.36, 1062.65, 1067.32,
            1071.43, 1074.44, 1076.64, 1078.21, 1079.31, 1079.62, 1079.33, 1079.29, 1079.49, 1079.82,
            1080.7, 1081.9, 1082.1, 1081.18, 1080.31, 1080.16, 1080.47, 1080.95, 1081.61, 1081.88,
            1081.86, 1082.08, 1082.78, 1083.89, 1084.75, 1085.2, 1085.14, 1084.89, 1085.07, 1086.17,
        ]
    }

    fn create_ss(period: i64) -> SuperSmoother {
        SuperSmoother::new(&SuperSmootherParams {
            shortest_cycle_period: period,
            ..Default::default()
        }).unwrap()
    }

    #[test]
    fn test_update() {
        const SKIP_ROWS: usize = 60;
        const TOLERANCE: f64 = 0.5;

        let input = test_input();
        let expected = test_expected();
        let mut ss = create_ss(10);

        for i in 0..input.len() {
            let act = ss.update(input[i]);

            if i < 2 {
                assert!(act.is_nan(), "[{}] expected NaN, got {}", i, act);
                continue;
            }

            if i < SKIP_ROWS {
                continue;
            }

            assert!(
                (act - expected[i]).abs() <= TOLERANCE,
                "[{}] expected {}, got {}", i, expected[i], act
            );
        }

        // NaN passthrough
        assert!(ss.update(f64::NAN).is_nan());
    }

    #[test]
    fn test_is_primed() {
        let input = test_input();
        let mut ss = create_ss(10);

        assert!(!ss.is_primed());

        for i in 0..2 {
            ss.update(input[i]);
            assert!(!ss.is_primed(), "[{}] should not be primed", i);
        }

        ss.update(input[2]);
        assert!(ss.is_primed(), "[2] should be primed");
    }

    #[test]
    fn test_update_entity() {
        let inp = 100.0_f64;
        let time = 1617235200_i64;
        let mut ss = create_ss(10);

        // Prime
        ss.update(inp);
        ss.update(inp);
        ss.update(inp);

        // Scalar
        let mut ss2 = create_ss(10);
        ss2.update(inp);
        ss2.update(inp);
        ss2.update(inp);
        let out = ss2.update_scalar(&Scalar::new(time, inp));
        assert_eq!(out.len(), 1);
        let s = out[0].downcast_ref::<Scalar>().unwrap();
        assert_eq!(s.time, time);
        assert!(!s.value.is_nan());

        // Bar (default component = Median = (high+low)/2)
        let mut ss2 = create_ss(10);
        ss2.update(inp);
        ss2.update(inp);
        ss2.update(inp);
        let bar = Bar::new(time, 0.0, inp, inp, 0.0, 0.0);
        let out = ss2.update_bar(&bar);
        let s = out[0].downcast_ref::<Scalar>().unwrap();
        assert!(!s.value.is_nan());

        // Quote (default component = Mid)
        let mut ss2 = create_ss(10);
        ss2.update(inp);
        ss2.update(inp);
        ss2.update(inp);
        let quote = Quote::new(time, inp, inp, 0.0, 0.0);
        let out = ss2.update_quote(&quote);
        let s = out[0].downcast_ref::<Scalar>().unwrap();
        assert!(!s.value.is_nan());

        // Trade (default component = Price)
        let mut ss2 = create_ss(10);
        ss2.update(inp);
        ss2.update(inp);
        ss2.update(inp);
        let trade = Trade::new(time, inp, 0.0);
        let out = ss2.update_trade(&trade);
        let s = out[0].downcast_ref::<Scalar>().unwrap();
        assert!(!s.value.is_nan());
    }

    #[test]
    fn test_metadata() {
        let ss = create_ss(10);
        let m = ss.metadata();

        assert_eq!(m.identifier, Identifier::SuperSmoother);
        assert_eq!(m.outputs.len(), 1);
        assert_eq!(m.outputs[0].kind, SuperSmootherOutput::Value as i32);
        assert_eq!(m.outputs[0].shape, Shape::Scalar);
        assert_eq!(m.outputs[0].mnemonic, "ss(10, hl/2)");
        assert_eq!(m.outputs[0].description, "Super Smoother ss(10, hl/2)");
    }

    #[test]
    fn test_new_period_validation() {
        let err_msg = "invalid super smoother parameters: shortest cycle period should be greater than 1";

        let r = SuperSmoother::new(&SuperSmootherParams {
            shortest_cycle_period: 1,
            ..Default::default()
        });
        assert!(r.is_err());
        assert_eq!(r.err().unwrap(), err_msg);

        let r = SuperSmoother::new(&SuperSmootherParams {
            shortest_cycle_period: 0,
            ..Default::default()
        });
        assert!(r.is_err());

        let r = SuperSmoother::new(&SuperSmootherParams {
            shortest_cycle_period: -1,
            ..Default::default()
        });
        assert!(r.is_err());
    }

    #[test]
    fn test_new_all_defaults() {
        let ss = create_ss(10);
        assert_eq!(ss.line.mnemonic, "ss(10, hl/2)");
        assert_eq!(ss.line.description, "Super Smoother ss(10, hl/2)");
    }

    #[test]
    fn test_new_bar_component_open() {
        let ss = SuperSmoother::new(&SuperSmootherParams {
            shortest_cycle_period: 10,
            bar_component: Some(BarComponent::Open),
            ..Default::default()
        }).unwrap();
        assert_eq!(ss.line.mnemonic, "ss(10, o)");
    }
}
