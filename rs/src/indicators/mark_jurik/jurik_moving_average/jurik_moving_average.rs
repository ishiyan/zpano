use crate::entities::bar::Bar;
use crate::entities::bar_component::{
    component_value as bar_component_value, BarComponent, DEFAULT_BAR_COMPONENT,
};
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
use crate::indicators::core::line_indicator::{BarFunc, QuoteFunc, TradeFunc};
use crate::indicators::core::metadata::Metadata;

// ---------------------------------------------------------------------------
// Params
// ---------------------------------------------------------------------------

/// Parameters to create an instance of the Jurik Moving Average indicator.
pub struct JurikMovingAverageParams {
    /// Length (number of time periods). Must be >= 1. Default 14.
    pub length: usize,
    /// Phase affects lag. Must be in [-100, 100]. Default 0.
    pub phase: i32,
    pub bar_component: Option<BarComponent>,
    pub quote_component: Option<QuoteComponent>,
    pub trade_component: Option<TradeComponent>,
}

impl Default for JurikMovingAverageParams {
    fn default() -> Self {
        Self {
            length: 14,
            phase: 0,
            bar_component: None,
            quote_component: None,
            trade_component: None,
        }
    }
}

// ---------------------------------------------------------------------------
// Output
// ---------------------------------------------------------------------------

/// Enumerates the outputs of the Jurik Moving Average indicator.
#[derive(Debug, Clone, Copy, PartialEq, Eq)]
#[repr(u8)]
pub enum JurikMovingAverageOutput {
    MovingAverage = 1,
}

// ---------------------------------------------------------------------------
// Indicator
// ---------------------------------------------------------------------------

/// Mark Jurik's Jurik Moving Average (JMA).
#[derive(Debug)]
pub struct JurikMovingAverage {
    primed: bool,
    list: Vec<f64>,
    ring: Vec<f64>,
    ring2: Vec<f64>,
    buffer: Vec<f64>,
    s28: i32,
    s30: i32,
    s38: i32,
    s40: i32,
    s48: i32,
    s50: i32,
    s70: i32,
    f0: i32,
    f_d8: i32,
    f_f0: i32,
    v5: i32,
    s8: f64,
    s18: f64,
    f10: f64,
    f18: f64,
    f38: f64,
    f50: f64,
    f58: f64,
    f78: f64,
    f88: f64,
    f90: f64,
    f98: f64,
    f_a8: f64,
    f_b8: f64,
    f_c0: f64,
    f_c8: f64,
    f_f8: f64,
    v1: f64,
    v2: f64,
    v3: f64,
    bar_func: BarFunc,
    quote_func: QuoteFunc,
    trade_func: TradeFunc,
    mnemonic: String,
    description: String,
}

impl JurikMovingAverage {
    pub fn new(params: &JurikMovingAverageParams) -> Result<Self, String> {
        if params.length < 1 {
            return Err(
                "invalid jurik moving average parameters: length should be positive".to_string(),
            );
        }
        if params.phase < -100 || params.phase > 100 {
            return Err(
                "invalid jurik moving average parameters: phase should be in range [-100, 100]"
                    .to_string(),
            );
        }

        let bc = params.bar_component.unwrap_or(DEFAULT_BAR_COMPONENT);
        let qc = params.quote_component.unwrap_or(DEFAULT_QUOTE_COMPONENT);
        let tc = params.trade_component.unwrap_or(DEFAULT_TRADE_COMPONENT);

        let bar_func = bar_component_value(bc);
        let quote_func = quote_component_value(qc);
        let trade_func = trade_component_value(tc);

        let mnemonic = format!(
            "jma({}, {}{})",
            params.length,
            params.phase,
            component_triple_mnemonic(bc, qc, tc)
        );
        let description = format!("Jurik moving average {}", mnemonic);

        const C128: usize = 128;
        const C11: usize = 11;
        const C62: usize = 62;
        const C_INIT: f64 = 1000000.0;

        let mut list = vec![0.0_f64; C128];
        let ring = vec![0.0_f64; C128];
        let ring2 = vec![0.0_f64; C11];
        let buffer = vec![0.0_f64; C62];

        for i in 0..64 {
            list[i] = -C_INIT;
        }
        for i in 64..C128 {
            list[i] = C_INIT;
        }

        let epsilon = 1e-10_f64;

        let mut f80 = epsilon;
        if params.length > 1 {
            f80 = (params.length as f64 - 1.0) / 2.0;
        }

        let f10 = params.phase as f64 / 100.0 + 1.5;

        let v1 = (f80.sqrt()).ln();
        let v2 = v1;
        let v3 = (v2 / 2.0_f64.ln() + 2.0).max(0.0);

        let f98 = v3;
        let f88 = (f98 - 2.0).max(0.5);

        let f78 = f80.sqrt() * f98;
        let f90 = f78 / (f78 + 1.0);
        f80 *= 0.9;
        let f50 = f80 / (f80 + 2.0);

        Ok(Self {
            primed: false,
            list,
            ring,
            ring2,
            buffer,
            s28: 63,
            s30: 64,
            s38: 0,
            s40: 0,
            s48: 0,
            s50: 0,
            s70: 0,
            f0: 1,
            f_d8: 0,
            f_f0: 0,
            v5: 0,
            s8: 0.0,
            s18: 0.0,
            f10,
            f18: 0.0,
            f38: 0.0,
            f50,
            f58: 0.0,
            f78,
            f88,
            f90,
            f98,
            f_a8: 0.0,
            f_b8: 0.0,
            f_c0: 0.0,
            f_c8: 0.0,
            f_f8: 0.0,
            v1,
            v2,
            v3,
            bar_func,
            quote_func,
            trade_func,
            mnemonic,
            description,
        })
    }

    /// Core update. Returns the JMA value.
    #[allow(clippy::too_many_lines)]
    pub fn update(&mut self, sample: f64) -> f64 {
        if sample.is_nan() {
            return sample;
        }

        const C2: i32 = 2;
        const C10: i32 = 10;
        const C29: i32 = 29;
        const C30: i32 = 30;
        const C31: i32 = 31;
        const C32: i32 = 32;
        const C61: i32 = 61;
        const C64: i32 = 64;
        const C96: i32 = 96;
        const C127: i32 = 127;
        const C128: i32 = 128;
        const EPSILON: f64 = 1e-10;

        if self.f_f0 < C61 {
            self.f_f0 += 1;
            self.buffer[self.f_f0 as usize] = sample;
        }

        if self.f_f0 <= C30 {
            return f64::NAN;
        }

        self.primed = true;

        if self.f0 == 0 {
            self.f_d8 = 0;
        } else {
            self.f0 = 0;
            self.v5 = 0;

            for i in 1..C30 {
                if self.buffer[(i + 1) as usize] != self.buffer[i as usize] {
                    self.v5 = 1;
                }
            }

            self.f_d8 = self.v5 * C30;
            if self.f_d8 == 0 {
                self.f38 = sample;
            } else {
                self.f38 = self.buffer[1];
            }

            self.f18 = self.f38;
            if self.f_d8 > C29 {
                self.f_d8 = C29;
            }
        }

        let mut i = self.f_d8;
        'outer: while i >= 0 {
            let f8 = if i != 0 {
                self.buffer[(C31 - i) as usize]
            } else {
                sample
            };

            let f28 = f8 - self.f18;
            let f48 = f8 - self.f38;
            let a28 = f28.abs();
            let a48 = f48.abs();
            self.v2 = a28.max(a48);

            let f_a0 = self.v2;
            let v = f_a0 + EPSILON;

            if self.s48 <= 1 {
                self.s48 = C127;
            } else {
                self.s48 -= 1;
            }

            if self.s50 <= 1 {
                self.s50 = C10;
            } else {
                self.s50 -= 1;
            }

            if self.s70 < C128 {
                self.s70 += 1;
            }

            self.s8 += v - self.ring2[self.s50 as usize];
            self.ring2[self.s50 as usize] = v;
            let mut s20 = self.s8 / self.s70 as f64;

            if self.s70 > C10 {
                s20 = self.s8 / 10.0;
            }

            let mut s58: i32;
            let mut s68: i32;

            if self.s70 > C127 {
                let s10 = self.ring[self.s48 as usize];
                self.ring[self.s48 as usize] = s20;
                s68 = C64;
                s58 = s68;

                while s68 > 1 {
                    if self.list[s58 as usize] < s10 {
                        s68 /= C2;
                        s58 += s68;
                    } else if self.list[s58 as usize] <= s10 {
                        s68 = 1;
                    } else {
                        s68 /= C2;
                        s58 -= s68;
                    }
                }
            } else {
                self.ring[self.s48 as usize] = s20;
                if self.s28 + self.s30 > C127 {
                    self.s30 -= 1;
                    s58 = self.s30;
                } else {
                    self.s28 += 1;
                    s58 = self.s28;
                }

                self.s38 = self.s28.min(C96);
                self.s40 = self.s30.max(C32);
            }

            s68 = C64;
            let mut s60 = s68;

            while s68 > 1 {
                if self.list[s60 as usize] >= s20 {
                    if self.list[(s60 - 1) as usize] <= s20 {
                        s68 = 1;
                    } else {
                        s68 /= C2;
                        s60 -= s68;
                    }
                } else {
                    s68 /= C2;
                    s60 += s68;
                }

                if s60 == C127 && s20 > self.list[C127 as usize] {
                    s60 = C128;
                }
            }

            if self.s70 > C127 {
                if s58 >= s60 {
                    if self.s38 + 1 > s60 && self.s40 - 1 < s60 {
                        self.s18 += s20;
                    } else if self.s40 > s60 && self.s40 - 1 < s58 {
                        self.s18 += self.list[(self.s40 - 1) as usize];
                    }
                } else if self.s40 >= s60 {
                    if self.s38 + 1 < s60 && self.s38 + 1 > s58 {
                        self.s18 += self.list[(self.s38 + 1) as usize];
                    }
                } else if self.s38 + 2 > s60 {
                    self.s18 += s20;
                } else if self.s38 + 1 < s60 && self.s38 + 1 > s58 {
                    self.s18 += self.list[(self.s38 + 1) as usize];
                }

                if s58 > s60 {
                    if self.s40 - 1 < s58 && self.s38 + 1 > s58 {
                        self.s18 -= self.list[s58 as usize];
                    } else if self.s38 < s58 && self.s38 + 1 > s60 {
                        self.s18 -= self.list[self.s38 as usize];
                    }
                } else {
                    if self.s38 + 1 > s58 && self.s40 - 1 < s58 {
                        self.s18 -= self.list[s58 as usize];
                    } else if self.s40 > s58 && self.s40 < s60 {
                        self.s18 -= self.list[self.s40 as usize];
                    }
                }
            }

            if s58 <= s60 {
                if s58 >= s60 {
                    self.list[s60 as usize] = s20;
                } else {
                    for k in (s58 + 1)..=(s60 - 1) {
                        self.list[(k - 1) as usize] = self.list[k as usize];
                    }
                    self.list[(s60 - 1) as usize] = s20;
                }
            } else {
                let mut k = s58 - 1;
                while k >= s60 {
                    self.list[(k + 1) as usize] = self.list[k as usize];
                    k -= 1;
                }
                self.list[s60 as usize] = s20;
            }

            if self.s70 < C128 {
                self.s18 = 0.0;
                for k in self.s40..=self.s38 {
                    self.s18 += self.list[k as usize];
                }
            }

            let f60 = self.s18 / (self.s38 - self.s40 + 1) as f64;

            if self.f_f8 + 1.0 > C31 as f64 {
                self.f_f8 = C31 as f64;
            } else {
                self.f_f8 += 1.0;
            }

            if self.f_f8 <= C30 as f64 {
                if f28 > 0.0 {
                    self.f18 = f8;
                } else {
                    self.f18 = f8 - f28 * self.f90;
                }

                if f48 < 0.0 {
                    self.f38 = f8;
                } else {
                    self.f38 = f8 - f48 * self.f90;
                }

                self.f_b8 = sample;
                if self.f_f8 != C30 as f64 {
                    i -= 1;
                    continue 'outer;
                }

                let mut v4 = 1_i32;
                self.f_c0 = sample;

                if self.f78.ceil() >= 1.0 {
                    v4 = self.f78.ceil() as i32;
                }

                let mut v2_local = 1_i32;
                let f_e8 = v4;

                if self.f78.floor() >= 1.0 {
                    v2_local = self.f78.floor() as i32;
                }

                let mut f68 = 1.0_f64;
                let f_e0 = v2_local;

                if f_e8 != f_e0 {
                    let diff = f_e8 - f_e0;
                    f68 = (self.f78 - f_e0 as f64) / diff as f64;
                }

                let v5_local = f_e0.min(C29);
                let v6_local = f_e8.min(C29);
                self.f_a8 = (sample - self.buffer[(self.f_f0 - v5_local) as usize]) * (1.0 - f68)
                    / f_e0 as f64
                    + (sample - self.buffer[(self.f_f0 - v6_local) as usize]) * f68 / f_e8 as f64;
            } else {
                let p = (f_a0 / f60).powf(self.f88);
                self.v1 = self.f98.min(p);

                if self.v1 < 1.0 {
                    self.v2 = 1.0;
                } else {
                    self.v3 = self.f98.min(p);
                    self.v2 = self.v3;
                }

                self.f58 = self.v2;
                let f70 = self.f90.powf(self.f58.sqrt());

                if f28 > 0.0 {
                    self.f18 = f8;
                } else {
                    self.f18 = f8 - f28 * f70;
                }

                if f48 < 0.0 {
                    self.f38 = f8;
                } else {
                    self.f38 = f8 - f48 * f70;
                }
            }

            i -= 1;
        }

        if self.f_f8 > C30 as f64 {
            let f30 = self.f50.powf(self.f58);
            self.f_c0 = (1.0 - f30) * sample + f30 * self.f_c0;
            self.f_c8 = (sample - self.f_c0) * (1.0 - self.f50) + self.f50 * self.f_c8;
            let f_d0 = self.f10 * self.f_c8 + self.f_c0;
            let f20 = f30 * -2.0;
            let f40 = f30 * f30;
            let f_b0 = f20 + f40 + 1.0;
            self.f_a8 = (f_d0 - self.f_b8) * f_b0 + f40 * self.f_a8;
            self.f_b8 += self.f_a8;
        }

        self.f_b8
    }
}

impl Indicator for JurikMovingAverage {
    fn is_primed(&self) -> bool {
        self.primed
    }

    fn metadata(&self) -> Metadata {
        build_metadata(
            Identifier::JurikMovingAverage,
            &self.mnemonic,
            &self.description,
            &[OutputText {
                mnemonic: self.mnemonic.clone(),
                description: self.description.clone(),
            }],
        )
    }

    fn update_scalar(&mut self, sample: &Scalar) -> Output {
        let val = self.update(sample.value);
        vec![Box::new(Scalar::new(sample.time, val))]
    }

    fn update_bar(&mut self, bar: &Bar) -> Output {
        let sample = (self.bar_func)(bar);
        let val = self.update(sample);
        vec![Box::new(Scalar::new(bar.time, val))]
    }

    fn update_quote(&mut self, quote: &Quote) -> Output {
        let sample = (self.quote_func)(quote);
        let val = self.update(sample);
        vec![Box::new(Scalar::new(quote.time, val))]
    }

    fn update_trade(&mut self, trade: &Trade) -> Output {
        let sample = (self.trade_func)(trade);
        let val = self.update(sample);
        vec![Box::new(Scalar::new(trade.time, val))]
    }
}

// ===========================================================================
// Tests
// ===========================================================================

#[cfg(test)]
mod tests {
    use super::*;
    use crate::indicators::core::indicator::Indicator;
    use crate::indicators::core::outputs::shape::Shape;

    use super::super::testdata::testdata;

    const TOLERANCE: f64 = 1e-13;

    fn almost_equal(a: f64, b: f64) -> bool {
        (a - b).abs() <= TOLERANCE
    }

    fn create(length: usize, phase: i32) -> JurikMovingAverage {
        JurikMovingAverage::new(&JurikMovingAverageParams {
            length,
            phase,
            ..Default::default()
        })
        .unwrap()
    }

    fn run_test(length: usize, phase: i32, expected: &[f64]) {
        let input = testdata::test_input();
        let mut jma = create(length, phase);
        const LEN_PRIMED: usize = 30;

        for i in 0..input.len() {
            let act = jma.update(input[i]);
            if i < LEN_PRIMED {
                assert!(act.is_nan(), "[{}] expected NaN, got {}", i, act);
            } else {
                assert!(
                    almost_equal(act, expected[i - LEN_PRIMED]),
                    "[{}] expected {}, got {} (diff={})",
                    i,
                    expected[i - LEN_PRIMED],
                    act,
                    (act - expected[i - LEN_PRIMED]).abs()
                );
            }
        }

        // NaN passthrough
        assert!(jma.update(f64::NAN).is_nan());
    }

    #[test]
    fn test_update_l20_pm100() {
        run_test(20, -100, &testdata::expected_l20_pm100());
    }

    #[test]
    fn test_update_l20_p0() {
        run_test(20, 0, &testdata::expected_l20_p0());
    }

    #[test]
    fn test_update_l20_p100() {
        run_test(20, 100, &testdata::expected_l20_p100());
    }

    #[test]
    fn test_update_l2_p1() {
        run_test(2, 1, &testdata::expected_l2_p1());
    }

    #[test]
    fn test_update_l10_p1() {
        run_test(10, 1, &testdata::expected_l10_p1());
    }

    #[test]
    fn test_is_primed() {
        let input = testdata::test_input();
        let mut jma = create(10, 30);
        assert!(!jma.is_primed());

        for i in 0..30 {
            jma.update(input[i]);
            assert!(!jma.is_primed(), "should not be primed at i={}", i);
        }

        for i in 30..input.len() {
            jma.update(input[i]);
            assert!(jma.is_primed(), "should be primed at i={}", i);
        }
    }

    #[test]
    fn test_metadata() {
        let jma = create(10, 30);
        let meta = jma.metadata();

        assert_eq!(meta.identifier, Identifier::JurikMovingAverage);
        assert_eq!(meta.outputs.len(), 1);
        assert_eq!(
            meta.outputs[0].kind,
            JurikMovingAverageOutput::MovingAverage as i32
        );
        assert_eq!(meta.outputs[0].shape, Shape::Scalar);
        assert_eq!(meta.outputs[0].mnemonic, "jma(10, 30)");
        assert_eq!(
            meta.outputs[0].description,
            "Jurik moving average jma(10, 30)"
        );
    }

    #[test]
    fn test_update_scalar() {
        let mut jma = create(10, 11);
        for _ in 0..30 {
            jma.update(3.0);
        }
        let s = Scalar {
            time: 12345,
            value: 3.0,
        };
        let out = jma.update_scalar(&s);
        assert_eq!(out.len(), 1);
    }

    #[test]
    fn test_invalid_params() {
        let r = JurikMovingAverage::new(&JurikMovingAverageParams {
            length: 0,
            phase: 30,
            ..Default::default()
        });
        assert!(r.is_err());
        assert!(r.unwrap_err().contains("length should be positive"));

        let r = JurikMovingAverage::new(&JurikMovingAverageParams {
            length: 10,
            phase: -101,
            ..Default::default()
        });
        assert!(r.is_err());
        assert!(r.unwrap_err().contains("phase should be in range"));

        let r = JurikMovingAverage::new(&JurikMovingAverageParams {
            length: 10,
            phase: 101,
            ..Default::default()
        });
        assert!(r.is_err());
        assert!(r.unwrap_err().contains("phase should be in range"));
    }
}
