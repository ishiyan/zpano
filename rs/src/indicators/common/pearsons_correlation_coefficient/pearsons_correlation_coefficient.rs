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
use crate::indicators::core::line_indicator::LineIndicator;
use crate::indicators::core::metadata::Metadata;

// ---------------------------------------------------------------------------
// Params
// ---------------------------------------------------------------------------

/// Parameters to create an instance of the Pearson's Correlation Coefficient indicator.
pub struct PearsonsCorrelationCoefficientParams {
    /// The length (number of time periods) of the rolling window.
    /// Must be greater than 0.
    pub length: usize,
    /// Bar component to extract. `None` means use default (Close).
    pub bar_component: Option<BarComponent>,
    /// Quote component to extract. `None` means use default (Mid).
    pub quote_component: Option<QuoteComponent>,
    /// Trade component to extract. `None` means use default (Price).
    pub trade_component: Option<TradeComponent>,
}

impl Default for PearsonsCorrelationCoefficientParams {
    fn default() -> Self {
        Self {
            length: 20,
            bar_component: None,
            quote_component: None,
            trade_component: None,
        }
    }
}

// ---------------------------------------------------------------------------
// Output
// ---------------------------------------------------------------------------

/// Enumerates the outputs of the Pearson's Correlation Coefficient indicator.
#[derive(Debug, Clone, Copy, PartialEq, Eq)]
#[repr(u8)]
pub enum PearsonsCorrelationCoefficientOutput {
    /// The scalar value of the correlation coefficient.
    Value = 1,
}

// ---------------------------------------------------------------------------
// Indicator
// ---------------------------------------------------------------------------

/// Computes Pearson's Correlation Coefficient (r) over a rolling window.
///
/// Given two input series X and Y, it computes:
///
/// r = (n·∑XY − ∑X·∑Y) / √((n·∑X² − (∑X)²) · (n·∑Y² − (∑Y)²))
///
/// The indicator is not primed during the first length−1 updates.
pub struct PearsonsCorrelationCoefficient {
    line: LineIndicator,
    length: usize,
    window_x: Vec<f64>,
    window_y: Vec<f64>,
    count: usize,
    pos: usize,
    sum_x: f64,
    sum_y: f64,
    sum_x2: f64,
    sum_y2: f64,
    sum_xy: f64,
    primed: bool,
}

impl PearsonsCorrelationCoefficient {
    /// Creates a new PearsonsCorrelationCoefficient from the given parameters.
    pub fn new(params: &PearsonsCorrelationCoefficientParams) -> Result<Self, String> {
        if params.length < 1 {
            return Err("invalid pearsons correlation coefficient parameters: length should be positive".to_string());
        }

        let bc = params.bar_component.unwrap_or(DEFAULT_BAR_COMPONENT);
        let qc = params.quote_component.unwrap_or(DEFAULT_QUOTE_COMPONENT);
        let tc = params.trade_component.unwrap_or(DEFAULT_TRADE_COMPONENT);

        let bar_func = bar_component_value(bc);
        let quote_func = quote_component_value(qc);
        let trade_func = trade_component_value(tc);

        let mnemonic = format!("correl({}{})", params.length, component_triple_mnemonic(bc, qc, tc));
        let description = format!("Pearsons Correlation Coefficient {}", mnemonic);

        let line = LineIndicator::new(mnemonic, description, bar_func, quote_func, trade_func);

        Ok(Self {
            line,
            length: params.length,
            window_x: vec![0.0; params.length],
            window_y: vec![0.0; params.length],
            count: 0,
            pos: 0,
            sum_x: 0.0,
            sum_y: 0.0,
            sum_x2: 0.0,
            sum_y2: 0.0,
            sum_xy: 0.0,
            primed: false,
        })
    }

    /// Core update logic for a single scalar (degenerate case: x == y).
    pub fn update(&mut self, sample: f64) -> f64 {
        self.update_pair(sample, sample)
    }

    /// Updates the indicator given an (x, y) pair.
    pub fn update_pair(&mut self, x: f64, y: f64) -> f64 {
        if x.is_nan() || y.is_nan() {
            return f64::NAN;
        }

        let n = self.length as f64;

        if self.primed {
            // Remove the oldest values.
            let old_x = self.window_x[self.pos];
            let old_y = self.window_y[self.pos];

            self.sum_x -= old_x;
            self.sum_y -= old_y;
            self.sum_x2 -= old_x * old_x;
            self.sum_y2 -= old_y * old_y;
            self.sum_xy -= old_x * old_y;

            // Add new values.
            self.window_x[self.pos] = x;
            self.window_y[self.pos] = y;
            self.pos = (self.pos + 1) % self.length;

            self.sum_x += x;
            self.sum_y += y;
            self.sum_x2 += x * x;
            self.sum_y2 += y * y;
            self.sum_xy += x * y;

            return self.correlate(n);
        }

        // Accumulating phase.
        self.window_x[self.count] = x;
        self.window_y[self.count] = y;

        self.sum_x += x;
        self.sum_y += y;
        self.sum_x2 += x * x;
        self.sum_y2 += y * y;
        self.sum_xy += x * y;

        self.count += 1;

        if self.count == self.length {
            self.primed = true;
            self.pos = 0;

            return self.correlate(n);
        }

        f64::NAN
    }

    /// Computes the Pearson correlation from the running sums.
    fn correlate(&self, n: f64) -> f64 {
        let var_x = self.sum_x2 - (self.sum_x * self.sum_x) / n;
        let var_y = self.sum_y2 - (self.sum_y * self.sum_y) / n;
        let temp_real = var_x * var_y;

        if temp_real <= 0.0 {
            return 0.0;
        }

        (self.sum_xy - (self.sum_x * self.sum_y) / n) / temp_real.sqrt()
    }
}

impl Indicator for PearsonsCorrelationCoefficient {
    fn is_primed(&self) -> bool {
        self.primed
    }

    fn metadata(&self) -> Metadata {
        build_metadata(
            Identifier::PearsonsCorrelationCoefficient,
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
        let x = sample.high;
        let y = sample.low;
        let value = self.update_pair(x, y);
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
    use super::super::testdata::testdata;
    use crate::entities::bar_component::BarComponent;
    use crate::indicators::core::outputs::shape::Shape;
    fn create(length: usize) -> PearsonsCorrelationCoefficient {
        PearsonsCorrelationCoefficient::new(&PearsonsCorrelationCoefficientParams {
            length,
            ..Default::default()
        }).unwrap()
    }

    #[test]
    fn test_update_pair_talib_spot_checks() {
        let mut c = create(20);
        let high = testdata::test_high_input();
        let low = testdata::test_low_input();

        for i in 0..19 {
            let act = c.update_pair(high[i], low[i]);
            assert!(act.is_nan(), "[{}] expected NaN", i);
        }

        for i in 19..high.len() {
            let act = c.update_pair(high[i], low[i]);
            match i {
                19 => assert!((0.9401569 - act).abs() < 1e-4, "[{}] expected 0.9401569, got {}", i, act),
                20 => assert!((0.9471812 - act).abs() < 1e-4, "[{}] expected 0.9471812, got {}", i, act),
                251 => assert!((0.8866901 - act).abs() < 1e-4, "[{}] expected 0.8866901, got {}", i, act),
                _ => {}
            }
        }

        assert!(c.update_pair(f64::NAN, 1.0).is_nan());
        assert!(c.update_pair(1.0, f64::NAN).is_nan());
    }

    #[test]
    fn test_update_pair_excel_verification() {
        let mut c = create(20);
        let high = testdata::test_high_input();
        let low = testdata::test_low_input();
        let expected = testdata::test_excel_expected();

        const EPS: f64 = 1e-10;

        for i in 0..19 {
            let act = c.update_pair(high[i], low[i]);
            assert!(act.is_nan(), "[{}] expected NaN", i);
        }

        for i in 19..high.len() {
            let act = c.update_pair(high[i], low[i]);
            assert!(
                (expected[i] - act).abs() < EPS,
                "input {}, expected {:.16}, actual {:.16}", i, expected[i], act
            );
        }
    }

    #[test]
    fn test_update_entity() {
        let time = 1617235200_i64;
        let inp = 3.0_f64;

        // scalar: correl(x,x) with constant value returns 0 (zero variance).
        let mut c = create(2);
        c.update(inp);
        c.update(inp);
        let out = c.update_scalar(&Scalar::new(time, inp));
        assert_eq!(out.len(), 1);
        let s = out[0].downcast_ref::<Scalar>().unwrap();
        assert_eq!(s.time, time);
        assert!((s.value - 0.0).abs() < 1e-10);

        // bar: uses high/low
        let mut c = create(2);
        c.update_pair(10.0, 5.0);
        c.update_pair(20.0, 10.0);
        let bar = Bar::new(time, 0.0, 10.0, 5.0, 0.0, 0.0);
        let out = c.update_bar(&bar);
        let s = out[0].downcast_ref::<Scalar>().unwrap();
        assert_eq!(s.time, time);
        assert!(!s.value.is_nan());

        // quote
        let mut c = create(2);
        c.update(inp);
        c.update(inp);
        let quote = Quote::new(time, inp, inp, 0.0, 0.0);
        let out = c.update_quote(&quote);
        let s = out[0].downcast_ref::<Scalar>().unwrap();
        assert!((s.value - 0.0).abs() < 1e-10);

        // trade
        let mut c = create(2);
        c.update(inp);
        c.update(inp);
        let trade = Trade::new(time, inp, 0.0);
        let out = c.update_trade(&trade);
        let s = out[0].downcast_ref::<Scalar>().unwrap();
        assert!((s.value - 0.0).abs() < 1e-10);
    }

    #[test]
    fn test_is_primed() {
        let high = testdata::test_high_input();
        let low = testdata::test_low_input();

        // length = 1
        let mut c = create(1);
        assert!(!c.is_primed());
        c.update_pair(high[0], low[0]);
        assert!(c.is_primed());

        // length = 2
        let mut c = create(2);
        assert!(!c.is_primed());
        c.update_pair(high[0], low[0]);
        assert!(!c.is_primed());
        c.update_pair(high[1], low[1]);
        assert!(c.is_primed());

        // length = 20
        let mut c = create(20);
        assert!(!c.is_primed());
        for i in 0..19 {
            c.update_pair(high[i], low[i]);
            assert!(!c.is_primed(), "[{}] should not be primed", i);
        }
        c.update_pair(high[19], low[19]);
        assert!(c.is_primed());
    }

    #[test]
    fn test_metadata() {
        let c = create(20);
        let m = c.metadata();
        assert_eq!(m.identifier, Identifier::PearsonsCorrelationCoefficient);
        assert_eq!(m.outputs.len(), 1);
        assert_eq!(m.outputs[0].kind, PearsonsCorrelationCoefficientOutput::Value as i32);
        assert_eq!(m.outputs[0].shape, Shape::Scalar);
        assert_eq!(m.outputs[0].mnemonic, "correl(20)");
        assert_eq!(m.outputs[0].description, "Pearsons Correlation Coefficient correl(20)");
    }

    #[test]
    fn test_new_invalid() {
        // length = 0
        let r = PearsonsCorrelationCoefficient::new(&PearsonsCorrelationCoefficientParams {
            length: 0, ..Default::default()
        });
        assert!(r.is_err());
        assert_eq!(r.err().unwrap(), "invalid pearsons correlation coefficient parameters: length should be positive");
    }

    #[test]
    fn test_mnemonic_components() {
        // all defaults -> no component suffix
        let c = create(20);
        assert_eq!(c.line.mnemonic, "correl(20)");

        // bar component set
        let c = PearsonsCorrelationCoefficient::new(&PearsonsCorrelationCoefficientParams {
            length: 20, bar_component: Some(BarComponent::Median), ..Default::default()
        }).unwrap();
        assert_eq!(c.line.mnemonic, "correl(20, hl/2)");
    }
}
