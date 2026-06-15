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

/// Parameters to create an instance of the Instantaneous Sine Wave Period indicator.
pub struct InstantaneousSineWavePeriodParams {
    /// EMA smoothing length applied to input prices before frequency estimation. >= 0. Default 0.
    pub smoothing: i64,
    /// Minimum allowed period in bars. > 0. Default 4.0.
    pub min_period: f64,
    /// Maximum allowed period in bars. > min_period. Default 50.0.
    pub max_period: f64,
    /// Maximum tolerated error for the omega estimate. > 0. Default 20.0.
    pub error_threshold: f64,
    /// Assumed measurement error for each price point. > 0. Default 0.01.
    pub dx: f64,
    /// Bar component to extract. `None` means use default (Close).
    pub bar_component: Option<BarComponent>,
    /// Quote component to extract. `None` means use default (Mid).
    pub quote_component: Option<QuoteComponent>,
    /// Trade component to extract. `None` means use default (Price).
    pub trade_component: Option<TradeComponent>,
}

impl Default for InstantaneousSineWavePeriodParams {
    fn default() -> Self {
        Self {
            smoothing: 0,
            min_period: 4.0,
            max_period: 50.0,
            error_threshold: 20.0,
            dx: 0.01,
            bar_component: None,
            quote_component: None,
            trade_component: None,
        }
    }
}

// ---------------------------------------------------------------------------
// Output
// ---------------------------------------------------------------------------

/// Enumerates the outputs of the Instantaneous Sine Wave Period indicator.
#[derive(Debug, Clone, Copy, PartialEq, Eq)]
#[repr(u8)]
pub enum InstantaneousSineWavePeriodOutput {
    /// The estimated cycle period in bars (may be NaN).
    Period = 1,
    /// The circular frequency in radians/bar (may be NaN).
    Omega = 2,
    /// The wave velocity (may be NaN).
    Velocity = 3,
    /// The wave acceleration (may be NaN).
    Acceleration = 4,
    /// The estimated sine wave amplitude (may be NaN).
    Amplitude = 5,
    /// The phase angle in radians (may be NaN).
    Phase = 6,
    /// The constant level D (may be NaN).
    DcLevel = 7,
}

// ---------------------------------------------------------------------------
// Indicator
// ---------------------------------------------------------------------------

/// Don Mak's Instantaneous Sine Wave Period (ISWP).
///
/// Estimates the dominant cycle period of price data by modeling it locally as a
/// single sine wave superimposed on a constant level, combining a 4-point method
/// (IF4) and a 5-point method (IF5) and selecting the one with the lower estimation
/// error at each bar.
pub struct InstantaneousSineWavePeriod {
    smoothing: i64,
    min_period: f64,
    max_period: f64,
    error_threshold: f64,
    dx: f64,
    ema_alpha: f64,
    ema_value: f64,
    ema_primed: bool,
    buffer: [f64; 5],
    count: usize,
    primed: bool,
    bar_func: fn(&Bar) -> f64,
    quote_func: fn(&Quote) -> f64,
    trade_func: fn(&Trade) -> f64,
    mnemonic: String,
}

impl InstantaneousSineWavePeriod {
    /// Creates a new Instantaneous Sine Wave Period from the given parameters.
    pub fn new(params: &InstantaneousSineWavePeriodParams) -> Result<Self, String> {
        let invalid = "invalid instantaneous sine wave period parameters";

        let smoothing = params.smoothing;
        let min_period = params.min_period;
        let max_period = params.max_period;
        let error_threshold = params.error_threshold;
        let dx = params.dx;

        if smoothing < 0 {
            return Err(format!("{}: smoothing should be >= 0", invalid));
        }
        if min_period <= 0.0 {
            return Err(format!("{}: min_period should be > 0", invalid));
        }
        if max_period <= min_period {
            return Err(format!("{}: max_period should be > min_period", invalid));
        }
        if error_threshold <= 0.0 {
            return Err(format!("{}: error_threshold should be > 0", invalid));
        }
        if dx <= 0.0 {
            return Err(format!("{}: dx should be > 0", invalid));
        }

        let bc = params.bar_component.unwrap_or(DEFAULT_BAR_COMPONENT);
        let qc = params.quote_component.unwrap_or(DEFAULT_QUOTE_COMPONENT);
        let tc = params.trade_component.unwrap_or(DEFAULT_TRADE_COMPONENT);

        let bar_func = bar_component_value(bc);
        let quote_func = quote_component_value(qc);
        let trade_func = trade_component_value(tc);

        let ema_alpha = if smoothing > 0 {
            2.0 / (smoothing as f64 + 1.0)
        } else {
            1.0
        };

        let mnemonic = format!(
            "iswp({},{:.2},{:.2},{:.2},{:.2}{})",
            smoothing,
            min_period,
            max_period,
            error_threshold,
            dx,
            component_triple_mnemonic(bc, qc, tc)
        );

        Ok(Self {
            smoothing,
            min_period,
            max_period,
            error_threshold,
            dx,
            ema_alpha,
            ema_value: 0.0,
            ema_primed: false,
            buffer: [0.0; 5],
            count: 0,
            primed: false,
            bar_func,
            quote_func,
            trade_func,
            mnemonic,
        })
    }

    /// Returns true if the indicator has produced at least one valid period.
    pub fn is_primed(&self) -> bool {
        self.primed
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
            return (f64::NAN, self.error_threshold);
        }

        let ratio = (x0 - xm3) / den;

        let sqrt_arg = 3.0 - ratio;
        if sqrt_arg < 0.0 {
            return (f64::NAN, self.error_threshold);
        }

        let arg = 0.5 * sqrt_arg.sqrt();
        if arg > 1.0 {
            return (f64::NAN, self.error_threshold);
        }

        let omega4 = 2.0 * arg.asin();

        let dx2 = self.dx * self.dx;

        let denom1 = 1.0 - 0.25 * sqrt_arg;
        if denom1 <= 0.0 || sqrt_arg == 0.0 {
            return (omega4, self.error_threshold);
        }

        let f1 = 1.0 / (denom1 * sqrt_arg);
        let inv_den2 = 1.0 / (den * den);
        let q2 = inv_den2 * (dx2 + dx2) + (ratio * ratio) * inv_den2 * (dx2 + dx2);

        let product = f1 * q2;
        if product < 0.0 {
            return (omega4, self.error_threshold);
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
            return (f64::NAN, self.error_threshold);
        }

        let arg = 0.5 * (x0 - xm4) / den1;
        if arg.abs() > 1.0 {
            return (f64::NAN, self.error_threshold);
        }

        let omega5 = arg.acos();

        let dx2 = self.dx * self.dx;

        let denom = 1.0 - arg * arg;
        if denom <= 0.0 {
            return (omega5, self.error_threshold);
        }

        let f1 = 1.0 / denom;
        let inv_den1_sq = 1.0 / (den1 * den1);
        let numerator_ratio = (x0 - xm4) / (den1 * den1);
        let r2 = inv_den1_sq * (dx2 + dx2) + (numerator_ratio * numerator_ratio) * (dx2 + dx2);

        let product = f1 * r2;
        if product < 0.0 {
            return (omega5, self.error_threshold);
        }

        (omega5, 0.5 * product.sqrt())
    }

    /// Returns (amplitude, phase, velocity, acceleration, dc_level).
    fn calc_model_params(&self, omega: f64) -> (f64, f64, f64, f64, f64) {
        let x0 = self.buffer[0];
        let xm1 = self.buffer[1];
        let xm2 = self.buffer[2];

        let half_w = omega / 2.0;
        let three_half_w = 1.5 * omega;

        let sin_hw = half_w.sin();
        let cos_hw = half_w.cos();
        let sin_3hw = three_half_w.sin();
        let cos_3hw = three_half_w.cos();

        let d0 = sin_hw * sin_hw * cos_hw * sin_3hw - sin_hw * sin_hw * sin_hw * cos_3hw;

        if d0.abs() < 1e-15 {
            return (f64::NAN, f64::NAN, f64::NAN, f64::NAN, f64::NAN);
        }

        let inv_d0 = 1.0 / d0;

        let dx0_m1 = x0 - xm1;
        let dxm1_m2 = xm1 - xm2;

        let c = inv_d0 * (dx0_m1 * sin_hw * sin_3hw - dxm1_m2 * sin_hw * sin_hw);
        let s = inv_d0 * (dxm1_m2 * sin_hw * cos_hw - dx0_m1 * sin_hw * cos_3hw);

        let amplitude = 0.5 * (c * c + s * s).sqrt();
        let phase = s.atan2(c);
        let velocity = amplitude * omega * phase.cos();
        let acceleration = -amplitude * omega * omega * phase.sin();
        let dc_level = x0 - s / 2.0;

        (amplitude, phase, velocity, acceleration, dc_level)
    }

    /// Core update returning (period, omega, velocity, acceleration, amplitude, phase, dc_level).
    pub fn update(&mut self, sample: f64) -> (f64, f64, f64, f64, f64, f64, f64) {
        let nan = f64::NAN;
        let invalid = (nan, nan, nan, nan, nan, nan, nan);

        let smoothed = if self.smoothing > 0 {
            self.apply_ema(sample)
        } else {
            sample
        };

        self.push_buffer(smoothed);
        self.count += 1;

        if self.count < 5 {
            return invalid;
        }

        let (omega4, error4) = self.calc_omega4();
        let (omega5, error5) = self.calc_omega5();

        if error4 >= self.error_threshold && error5 >= self.error_threshold {
            return invalid;
        }

        let omega = if error5 < error4 { omega5 } else { omega4 };

        if omega.is_nan() || omega <= 0.0 {
            return invalid;
        }

        let period = (2.0 * std::f64::consts::PI) / omega;
        if period < self.min_period || period > self.max_period {
            return invalid;
        }

        let (amplitude, phase, velocity, acceleration, dc_level) = self.calc_model_params(omega);

        self.primed = true;

        (period, omega, velocity, acceleration, amplitude, phase, dc_level)
    }
}

impl Indicator for InstantaneousSineWavePeriod {
    fn is_primed(&self) -> bool {
        self.primed
    }

    fn metadata(&self) -> Metadata {
        let desc = format!("Instantaneous Sine Wave Period {}", self.mnemonic);
        build_metadata(
            Identifier::InstantaneousSineWavePeriod,
            &self.mnemonic,
            &desc,
            &[
                OutputText { mnemonic: format!("{} period", self.mnemonic), description: format!("{} Period", desc) },
                OutputText { mnemonic: format!("{} omega", self.mnemonic), description: format!("{} Omega", desc) },
                OutputText { mnemonic: format!("{} velocity", self.mnemonic), description: format!("{} Velocity", desc) },
                OutputText { mnemonic: format!("{} acceleration", self.mnemonic), description: format!("{} Acceleration", desc) },
                OutputText { mnemonic: format!("{} amplitude", self.mnemonic), description: format!("{} Amplitude", desc) },
                OutputText { mnemonic: format!("{} phase", self.mnemonic), description: format!("{} Phase", desc) },
                OutputText { mnemonic: format!("{} dcLevel", self.mnemonic), description: format!("{} DC Level", desc) },
            ],
        )
    }

    fn update_scalar(&mut self, sample: &Scalar) -> Output {
        let (period, omega, velocity, acceleration, amplitude, phase, dc_level) = self.update(sample.value);
        vec![
            Box::new(Scalar { time: sample.time, value: period }),
            Box::new(Scalar { time: sample.time, value: omega }),
            Box::new(Scalar { time: sample.time, value: velocity }),
            Box::new(Scalar { time: sample.time, value: acceleration }),
            Box::new(Scalar { time: sample.time, value: amplitude }),
            Box::new(Scalar { time: sample.time, value: phase }),
            Box::new(Scalar { time: sample.time, value: dc_level }),
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
    use super::super::testdata::testdata;

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

    struct Combo {
        smoothing: i64,
        period: Vec<f64>,
        omega: Vec<f64>,
        velocity: Vec<f64>,
        acceleration: Vec<f64>,
    }

    #[test]
    fn test_reference_data_all_combos() {
        let combos = vec![
            Combo { smoothing: 0, period: testdata::expected_s0_period(), omega: testdata::expected_s0_omega(), velocity: testdata::expected_s0_velocity(), acceleration: testdata::expected_s0_acceleration() },
            Combo { smoothing: 3, period: testdata::expected_s3_period(), omega: testdata::expected_s3_omega(), velocity: testdata::expected_s3_velocity(), acceleration: testdata::expected_s3_acceleration() },
            Combo { smoothing: 6, period: testdata::expected_s6_period(), omega: testdata::expected_s6_omega(), velocity: testdata::expected_s6_velocity(), acceleration: testdata::expected_s6_acceleration() },
            Combo { smoothing: 12, period: testdata::expected_s12_period(), omega: testdata::expected_s12_omega(), velocity: testdata::expected_s12_velocity(), acceleration: testdata::expected_s12_acceleration() },
        ];

        let input = testdata::test_input();

        for combo in &combos {
            let mut ind = InstantaneousSineWavePeriod::new(&InstantaneousSineWavePeriodParams {
                smoothing: combo.smoothing,
                ..Default::default()
            })
            .unwrap();

            for i in 0..input.len() {
                let (period, omega, velocity, acceleration, _, _, _) = ind.update(input[i]);
                check("period", i, combo.period[i], period);
                check("omega", i, combo.omega[i], omega);
                check("velocity", i, combo.velocity[i], velocity);
                check("acceleration", i, combo.acceleration[i], acceleration);
            }
        }
    }

    #[test]
    fn test_mnemonic() {
        let ind = InstantaneousSineWavePeriod::new(&InstantaneousSineWavePeriodParams::default()).unwrap();
        assert_eq!(ind.metadata().mnemonic, "iswp(0,4.00,50.00,20.00,0.01)");

        let ind2 = InstantaneousSineWavePeriod::new(&InstantaneousSineWavePeriodParams {
            smoothing: 6,
            ..Default::default()
        })
        .unwrap();
        assert_eq!(ind2.metadata().mnemonic, "iswp(6,4.00,50.00,20.00,0.01)");
    }

    #[test]
    fn test_metadata() {
        let ind = InstantaneousSineWavePeriod::new(&InstantaneousSineWavePeriodParams::default()).unwrap();
        let meta = ind.metadata();
        assert_eq!(meta.identifier, Identifier::InstantaneousSineWavePeriod);
        assert_eq!(meta.outputs.len(), 7);
        assert_eq!(meta.outputs[0].kind, InstantaneousSineWavePeriodOutput::Period as i32);
        assert_eq!(meta.outputs[6].kind, InstantaneousSineWavePeriodOutput::DcLevel as i32);
    }

    #[test]
    fn test_update_scalar_ordering() {
        let input = testdata::test_input();
        let exp_period = testdata::expected_s0_period();

        let mut ind = InstantaneousSineWavePeriod::new(&InstantaneousSineWavePeriodParams::default()).unwrap();
        let mut out: Output = vec![];
        for i in 0..input.len() {
            out = ind.update_scalar(&Scalar { time: 0, value: input[i] });
        }
        assert_eq!(out.len(), 7);
        let last = input.len() - 1;
        let period = out[0].downcast_ref::<Scalar>().unwrap().value;
        check("period", last, exp_period[last], period);
    }

    #[test]
    fn test_invalid_params() {
        assert!(InstantaneousSineWavePeriod::new(&InstantaneousSineWavePeriodParams { smoothing: -1, ..Default::default() }).is_err());
        assert!(InstantaneousSineWavePeriod::new(&InstantaneousSineWavePeriodParams { min_period: 0.0, ..Default::default() }).is_err());
        assert!(InstantaneousSineWavePeriod::new(&InstantaneousSineWavePeriodParams { min_period: 50.0, max_period: 50.0, ..Default::default() }).is_err());
        assert!(InstantaneousSineWavePeriod::new(&InstantaneousSineWavePeriodParams { error_threshold: 0.0, ..Default::default() }).is_err());
        assert!(InstantaneousSineWavePeriod::new(&InstantaneousSineWavePeriodParams { dx: 0.0, ..Default::default() }).is_err());
    }
}
