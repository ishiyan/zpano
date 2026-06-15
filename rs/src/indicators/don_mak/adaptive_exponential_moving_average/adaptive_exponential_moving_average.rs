use crate::entities::bar::Bar;
use crate::entities::bar_component::{component_value as bar_component_value, BarComponent, DEFAULT_BAR_COMPONENT};
use crate::entities::quote::Quote;
use crate::entities::quote_component::{component_value as quote_component_value, QuoteComponent, DEFAULT_QUOTE_COMPONENT};
use crate::entities::scalar::Scalar;
use crate::entities::trade::Trade;
use crate::entities::trade_component::{component_value as trade_component_value, TradeComponent, DEFAULT_TRADE_COMPONENT};
use crate::indicators::core::build_metadata::{build_metadata, OutputText};
use crate::indicators::core::component_triple_mnemonic::component_triple_mnemonic;
use crate::indicators::core::identifier::Identifier;
use crate::indicators::core::indicator::{Indicator, Output};
use crate::indicators::core::metadata::Metadata;

// ---------------------------------------------------------------------------
// Params
// ---------------------------------------------------------------------------

/// Parameters to create an instance of the Adaptive Exponential Moving Average indicator.
pub struct AdaptiveExponentialMovingAverageParams {
    /// Smoothing factor for trending data (low frequency). In (0, 1], > alpha_min. Default 0.5.
    pub alpha_max: f64,
    /// Smoothing factor for noisy data (high frequency). In (0, alpha_max). Default 0.05.
    pub alpha_min: f64,
    /// Crossover frequency in radians/bar. In (0, pi). Default 1.0.
    pub omega0: f64,
    /// Embedded ISWP internal smoothing parameter. >= 0. Default 3.
    pub smoothing: i64,
    /// Bar component to extract. `None` means use default (Close).
    pub bar_component: Option<BarComponent>,
    /// Quote component to extract. `None` means use default (Mid).
    pub quote_component: Option<QuoteComponent>,
    /// Trade component to extract. `None` means use default (Price).
    pub trade_component: Option<TradeComponent>,
}

impl Default for AdaptiveExponentialMovingAverageParams {
    fn default() -> Self {
        Self {
            alpha_max: 0.5,
            alpha_min: 0.05,
            omega0: 1.0,
            smoothing: 3,
            bar_component: None,
            quote_component: None,
            trade_component: None,
        }
    }
}

// ---------------------------------------------------------------------------
// Output
// ---------------------------------------------------------------------------

/// Enumerates the outputs of the Adaptive Exponential Moving Average indicator.
#[derive(Debug, Clone, Copy, PartialEq, Eq)]
#[repr(u8)]
pub enum AdaptiveExponentialMovingAverageOutput {
    /// The adaptively smoothed price value.
    Value = 1,
    /// The instantaneous frequency estimate (may be NaN).
    Omega = 2,
    /// The smoothing factor used for the bar.
    Alpha = 3,
}

// ---------------------------------------------------------------------------
// Embedded ISWP omega estimator
// ---------------------------------------------------------------------------

const ISWP_MIN_PERIOD: f64 = 4.0;
const ISWP_MAX_PERIOD: f64 = 50.0;
const ISWP_ERROR_THRESHOLD: f64 = 20.0;
const ISWP_DX: f64 = 0.01;

/// Embedded Instantaneous Sine Wave Period omega estimator (omega-only reduction).
///
/// Estimates the dominant circular frequency omega of price data by modeling it
/// locally as a single sine wave, combining a 4-point and a 5-point method and
/// selecting the one with the lower estimation error. Inlined so the indicator is a
/// standalone porting unit. Do NOT change its numerics.
struct Iswp {
    smoothing: i64,
    ema_alpha: f64,
    ema_value: f64,
    ema_primed: bool,
    buffer: [f64; 5],
    count: usize,
}

impl Iswp {
    fn new(smoothing: i64) -> Self {
        let ema_alpha = if smoothing > 0 {
            2.0 / (smoothing as f64 + 1.0)
        } else {
            1.0
        };
        Self {
            smoothing,
            ema_alpha,
            ema_value: 0.0,
            ema_primed: false,
            buffer: [0.0; 5],
            count: 0,
        }
    }

    fn apply_ema(&mut self, price: f64) -> f64 {
        if !self.ema_primed {
            self.ema_value = price;
            self.ema_primed = true;
        } else {
            self.ema_value = self.ema_alpha * price + (1.0 - self.ema_alpha) * self.ema_value;
        }
        self.ema_value
    }

    fn push_buffer(&mut self, value: f64) {
        let mut i = 4;
        while i > 0 {
            self.buffer[i] = self.buffer[i - 1];
            i -= 1;
        }
        self.buffer[0] = value;
    }

    fn calc_omega4(&self) -> (f64, f64) {
        let x0 = self.buffer[0];
        let xm1 = self.buffer[1];
        let xm2 = self.buffer[2];
        let xm3 = self.buffer[3];

        let den = xm1 - xm2;
        if den == 0.0 {
            return (f64::NAN, ISWP_ERROR_THRESHOLD);
        }

        let ratio = (x0 - xm3) / den;

        let sqrt_arg = 3.0 - ratio;
        if sqrt_arg < 0.0 {
            return (f64::NAN, ISWP_ERROR_THRESHOLD);
        }

        let arg = 0.5 * sqrt_arg.sqrt();
        if arg > 1.0 {
            return (f64::NAN, ISWP_ERROR_THRESHOLD);
        }

        let omega4 = 2.0 * arg.asin();

        let dx2 = ISWP_DX * ISWP_DX;

        let denom1 = 1.0 - 0.25 * sqrt_arg;
        if denom1 <= 0.0 || sqrt_arg == 0.0 {
            return (omega4, ISWP_ERROR_THRESHOLD);
        }

        let f1 = 1.0 / (denom1 * sqrt_arg);
        let inv_den2 = 1.0 / (den * den);
        let q2 = inv_den2 * (dx2 + dx2) + (ratio * ratio) * inv_den2 * (dx2 + dx2);

        let product = f1 * q2;
        if product < 0.0 {
            return (omega4, ISWP_ERROR_THRESHOLD);
        }

        (omega4, 0.5 * product.sqrt())
    }

    fn calc_omega5(&self) -> (f64, f64) {
        let x0 = self.buffer[0];
        let xm1 = self.buffer[1];
        let xm3 = self.buffer[3];
        let xm4 = self.buffer[4];

        let den1 = xm1 - xm3;
        if den1 == 0.0 {
            return (f64::NAN, ISWP_ERROR_THRESHOLD);
        }

        let arg = 0.5 * (x0 - xm4) / den1;
        if arg.abs() > 1.0 {
            return (f64::NAN, ISWP_ERROR_THRESHOLD);
        }

        let omega5 = arg.acos();

        let dx2 = ISWP_DX * ISWP_DX;

        let denom = 1.0 - arg * arg;
        if denom <= 0.0 {
            return (omega5, ISWP_ERROR_THRESHOLD);
        }

        let f1 = 1.0 / denom;
        let inv_den1_sq = 1.0 / (den1 * den1);
        let numerator_ratio = (x0 - xm4) / (den1 * den1);
        let r2 = inv_den1_sq * (dx2 + dx2) + (numerator_ratio * numerator_ratio) * (dx2 + dx2);

        let product = f1 * r2;
        if product < 0.0 {
            return (omega5, ISWP_ERROR_THRESHOLD);
        }

        (omega5, 0.5 * product.sqrt())
    }

    fn update(&mut self, price: f64) -> f64 {
        let smoothed = if self.smoothing > 0 {
            self.apply_ema(price)
        } else {
            price
        };

        self.push_buffer(smoothed);
        self.count += 1;

        if self.count < 5 {
            return f64::NAN;
        }

        let (omega4, error4) = self.calc_omega4();
        let (omega5, error5) = self.calc_omega5();

        if error4 >= ISWP_ERROR_THRESHOLD && error5 >= ISWP_ERROR_THRESHOLD {
            return f64::NAN;
        }

        let omega = if error5 < error4 { omega5 } else { omega4 };

        if omega.is_nan() || omega <= 0.0 {
            return f64::NAN;
        }

        let period = (2.0 * std::f64::consts::PI) / omega;
        if period < ISWP_MIN_PERIOD || period > ISWP_MAX_PERIOD {
            return f64::NAN;
        }

        omega
    }
}

// ---------------------------------------------------------------------------
// Indicator
// ---------------------------------------------------------------------------

/// Don Mak's Adaptive Exponential Moving Average (AEMA).
///
/// An EMA with a time-varying smoothing factor alpha that adapts based on the
/// instantaneous frequency of the price data, estimated by an embedded ISWP.
///
/// The indicator produces three outputs:
///   - value: the adaptively smoothed price (never NaN);
///   - omega: the instantaneous frequency estimate (may be NaN);
///   - alpha: the smoothing factor used for this bar.
pub struct AdaptiveExponentialMovingAverage {
    alpha_max: f64,
    alpha_min: f64,
    omega0: f64,
    a: f64,
    b: f64,
    iswp: Iswp,
    ema_value: f64,
    initialized: bool,
    primed: bool,
    bar_func: fn(&Bar) -> f64,
    quote_func: fn(&Quote) -> f64,
    trade_func: fn(&Trade) -> f64,
    mnemonic: String,
}

impl AdaptiveExponentialMovingAverage {
    /// Creates a new Adaptive Exponential Moving Average from the given parameters.
    pub fn new(params: &AdaptiveExponentialMovingAverageParams) -> Result<Self, String> {
        let invalid = "invalid adaptive exponential moving average parameters";

        let alpha_max = params.alpha_max;
        let alpha_min = params.alpha_min;
        let omega0 = params.omega0;
        let smoothing = params.smoothing;

        if !(alpha_min > 0.0 && alpha_min < alpha_max && alpha_max <= 1.0) {
            return Err(format!("{}: need 0 < alpha_min < alpha_max <= 1", invalid));
        }
        if !(omega0 > 0.0 && omega0 < std::f64::consts::PI) {
            return Err(format!("{}: need 0 < omega0 < pi", invalid));
        }
        if smoothing < 0 {
            return Err(format!("{}: smoothing should be >= 0", invalid));
        }

        let bc = params.bar_component.unwrap_or(DEFAULT_BAR_COMPONENT);
        let qc = params.quote_component.unwrap_or(DEFAULT_QUOTE_COMPONENT);
        let tc = params.trade_component.unwrap_or(DEFAULT_TRADE_COMPONENT);

        let bar_func = bar_component_value(bc);
        let quote_func = quote_component_value(qc);
        let trade_func = trade_component_value(tc);

        let pi = std::f64::consts::PI;
        let a = (alpha_max - alpha_min) * omega0 * pi / (pi - omega0);
        let b = alpha_min - a / pi;

        let mnemonic = format!(
            "aema({:.2},{:.2},{:.2},{}{})",
            alpha_max,
            alpha_min,
            omega0,
            smoothing,
            component_triple_mnemonic(bc, qc, tc)
        );

        Ok(Self {
            alpha_max,
            alpha_min,
            omega0,
            a,
            b,
            iswp: Iswp::new(smoothing),
            ema_value: 0.0,
            initialized: false,
            primed: false,
            bar_func,
            quote_func,
            trade_func,
            mnemonic,
        })
    }

    /// Returns true if the indicator has produced at least one valid omega estimate.
    pub fn is_primed(&self) -> bool {
        self.primed
    }

    fn compute_alpha(&self, omega: f64) -> f64 {
        if omega.is_nan() {
            return self.alpha_min;
        }
        if omega <= self.omega0 {
            return self.alpha_max;
        }
        if omega >= std::f64::consts::PI {
            return self.alpha_min;
        }

        let alpha = self.a / omega + self.b;
        if alpha > self.alpha_max {
            return self.alpha_max;
        }
        if alpha < self.alpha_min {
            return self.alpha_min;
        }
        alpha
    }

    /// Core update returning (value, omega, alpha).
    pub fn update(&mut self, sample: f64) -> (f64, f64, f64) {
        let omega = self.iswp.update(sample);
        let alpha = self.compute_alpha(omega);

        if !self.initialized {
            self.ema_value = sample;
            self.initialized = true;
        } else {
            self.ema_value = alpha * sample + (1.0 - alpha) * self.ema_value;
        }

        if !omega.is_nan() {
            self.primed = true;
        }

        (self.ema_value, omega, alpha)
    }
}

impl Indicator for AdaptiveExponentialMovingAverage {
    fn is_primed(&self) -> bool {
        self.primed
    }

    fn metadata(&self) -> Metadata {
        let desc = format!("Adaptive Exponential Moving Average {}", self.mnemonic);
        build_metadata(
            Identifier::AdaptiveExponentialMovingAverage,
            &self.mnemonic,
            &desc,
            &[
                OutputText {
                    mnemonic: format!("{} value", self.mnemonic),
                    description: format!("{} Value", desc),
                },
                OutputText {
                    mnemonic: format!("{} omega", self.mnemonic),
                    description: format!("{} Omega", desc),
                },
                OutputText {
                    mnemonic: format!("{} alpha", self.mnemonic),
                    description: format!("{} Alpha", desc),
                },
            ],
        )
    }

    fn update_scalar(&mut self, sample: &Scalar) -> Output {
        let (value, omega, alpha) = self.update(sample.value);
        vec![
            Box::new(Scalar { time: sample.time, value }),
            Box::new(Scalar { time: sample.time, value: omega }),
            Box::new(Scalar { time: sample.time, value: alpha }),
        ]
    }

    fn update_bar(&mut self, sample: &Bar) -> Output {
        let v = (self.bar_func)(sample);
        self.update_scalar(&Scalar { time: sample.time, value: v })
    }

    fn update_quote(&mut self, sample: &Quote) -> Output {
        let v = (self.quote_func)(sample);
        self.update_scalar(&Scalar { time: sample.time, value: v })
    }

    fn update_trade(&mut self, sample: &Trade) -> Output {
        let v = (self.trade_func)(sample);
        self.update_scalar(&Scalar { time: sample.time, value: v })
    }
}

// ===========================================================================
// Tests
// ===========================================================================

#[cfg(test)]
mod tests {
    use super::*;
    use super::super::testdata::test_data;

    const TOLERANCE: f64 = 1e-9;

    fn check(name: &str, i: usize, exp: f64, act: f64) {
        if exp.is_nan() {
            assert!(act.is_nan(), "{}[{}]: expected NaN, got {}", name, i, act);
            return;
        }
        assert!(
            (act - exp).abs() <= TOLERANCE,
            "{}[{}]: expected {}, got {}",
            name,
            i,
            exp,
            act
        );
    }

    struct ValueCombo {
        alpha_max: f64,
        alpha_min: f64,
        omega0: f64,
        smoothing: i64,
        expected: Vec<f64>,
    }

    #[test]
    fn test_value_combos() {
        let combos = vec![
            ValueCombo { alpha_max: 0.5, alpha_min: 0.05, omega0: 1.0, smoothing: 3, expected: test_data::expected_default() },
            ValueCombo { alpha_max: 0.8, alpha_min: 0.02, omega0: 1.0, smoothing: 3, expected: test_data::expected_a0_8_a0_02() },
            ValueCombo { alpha_max: 0.5, alpha_min: 0.05, omega0: 0.5, smoothing: 3, expected: test_data::expected_w0_5() },
            ValueCombo { alpha_max: 0.5, alpha_min: 0.05, omega0: 1.5, smoothing: 3, expected: test_data::expected_w1_5() },
            ValueCombo { alpha_max: 0.5, alpha_min: 0.05, omega0: 1.0, smoothing: 0, expected: test_data::expected_s0() },
            ValueCombo { alpha_max: 0.5, alpha_min: 0.05, omega0: 1.0, smoothing: 6, expected: test_data::expected_s6() },
        ];

        let input = test_data::test_input();

        for combo in &combos {
            let mut ind = AdaptiveExponentialMovingAverage::new(&AdaptiveExponentialMovingAverageParams {
                alpha_max: combo.alpha_max,
                alpha_min: combo.alpha_min,
                omega0: combo.omega0,
                smoothing: combo.smoothing,
                ..Default::default()
            })
            .unwrap();

            for i in 0..input.len() {
                let (value, _, _) = ind.update(input[i]);
                check("value", i, combo.expected[i], value);
            }
        }
    }

    #[test]
    fn test_default_omega_alpha() {
        let input = test_data::test_input();
        let exp_omega = test_data::expected_default_omega();
        let exp_alpha = test_data::expected_default_alpha();

        let mut ind = AdaptiveExponentialMovingAverage::new(&AdaptiveExponentialMovingAverageParams::default()).unwrap();
        for i in 0..input.len() {
            let (_, omega, alpha) = ind.update(input[i]);
            check("omega", i, exp_omega[i], omega);
            check("alpha", i, exp_alpha[i], alpha);
        }
    }

    #[test]
    fn test_sine() {
        let input = test_data::test1_input_sine();
        let exp_value = test_data::test1_expected();
        let exp_omega = test_data::test1_expected_omega();
        let exp_alpha = test_data::test1_expected_alpha();

        let mut ind = AdaptiveExponentialMovingAverage::new(&AdaptiveExponentialMovingAverageParams::default()).unwrap();
        for i in 0..input.len() {
            let (value, omega, alpha) = ind.update(input[i]);
            check("value", i, exp_value[i], value);
            check("omega", i, exp_omega[i], omega);
            check("alpha", i, exp_alpha[i], alpha);
        }
    }

    #[test]
    fn test_mnemonic() {
        let ind = AdaptiveExponentialMovingAverage::new(&AdaptiveExponentialMovingAverageParams::default()).unwrap();
        assert_eq!(ind.metadata().mnemonic, "aema(0.50,0.05,1.00,3)");

        let ind2 = AdaptiveExponentialMovingAverage::new(&AdaptiveExponentialMovingAverageParams {
            alpha_max: 0.8,
            alpha_min: 0.02,
            omega0: 1.5,
            smoothing: 6,
            ..Default::default()
        })
        .unwrap();
        assert_eq!(ind2.metadata().mnemonic, "aema(0.80,0.02,1.50,6)");
    }

    #[test]
    fn test_metadata() {
        let ind = AdaptiveExponentialMovingAverage::new(&AdaptiveExponentialMovingAverageParams::default()).unwrap();
        let meta = ind.metadata();
        assert_eq!(meta.identifier, Identifier::AdaptiveExponentialMovingAverage);
        assert_eq!(meta.outputs.len(), 3);
        assert_eq!(meta.outputs[0].kind, AdaptiveExponentialMovingAverageOutput::Value as i32);
        assert_eq!(meta.outputs[1].kind, AdaptiveExponentialMovingAverageOutput::Omega as i32);
        assert_eq!(meta.outputs[2].kind, AdaptiveExponentialMovingAverageOutput::Alpha as i32);
    }

    #[test]
    fn test_update_scalar_ordering() {
        let input = test_data::test_input();
        let exp_value = test_data::expected_default();
        let exp_alpha = test_data::expected_default_alpha();

        let mut ind = AdaptiveExponentialMovingAverage::new(&AdaptiveExponentialMovingAverageParams::default()).unwrap();
        let mut out: Output = vec![];
        for i in 0..input.len() {
            out = ind.update_scalar(&Scalar { time: 0, value: input[i] });
        }
        let last = input.len() - 1;
        let value = out[0].downcast_ref::<Scalar>().unwrap().value;
        let alpha = out[2].downcast_ref::<Scalar>().unwrap().value;
        check("value", last, exp_value[last], value);
        check("alpha", last, exp_alpha[last], alpha);
    }

    #[test]
    fn test_invalid_params() {
        assert!(AdaptiveExponentialMovingAverage::new(&AdaptiveExponentialMovingAverageParams { alpha_max: 0.05, alpha_min: 0.5, ..Default::default() }).is_err());
        assert!(AdaptiveExponentialMovingAverage::new(&AdaptiveExponentialMovingAverageParams { alpha_max: 1.5, ..Default::default() }).is_err());
        assert!(AdaptiveExponentialMovingAverage::new(&AdaptiveExponentialMovingAverageParams { omega0: 4.0, ..Default::default() }).is_err());
        assert!(AdaptiveExponentialMovingAverage::new(&AdaptiveExponentialMovingAverageParams { smoothing: -1, ..Default::default() }).is_err());
    }
}
