use super::shared::*;

#[derive(Debug)]
pub struct PhaseAccumulatorEstimator {
    smoothing_length: usize,
    min_period: usize,
    max_period: usize,
    alpha_ema_qi: f64,
    alpha_ema_p: f64,
    warm_up_period: usize,
    sl_plus_ht_min1: usize,
    sl_plus_2ht_min2: usize,
    sl_plus_2ht_min1: usize,
    sl_plus_2ht: usize,
    one_min_alpha_qi: f64,
    one_min_alpha_p: f64,
    raw_values: Vec<f64>,
    wma_factors: Vec<f64>,
    wma_smoothed: Vec<f64>,
    detrended_arr: Vec<f64>,
    delta_phase: Vec<f64>,
    in_phase_val: f64,
    quadrature_val: f64,
    count: usize,
    smoothed_ip_prev: f64,
    smoothed_q_prev: f64,
    phase_prev: f64,
    period_val: f64,
    is_primed: bool,
    is_warmed_up: bool,
}

impl PhaseAccumulatorEstimator {
    pub fn new(p: &CycleEstimatorParams) -> Result<Self, String> {
        verify_parameters(p)?;
        let length = p.smoothing_length;
        let alpha_qi = p.alpha_ema_quadrature_in_phase;
        let alpha_p = p.alpha_ema_period;

        let sl_plus_ht_min1 = length + HT_LENGTH - 1;
        let sl_plus_2ht_min2 = sl_plus_ht_min1 + HT_LENGTH - 1;
        let sl_plus_2ht_min1 = sl_plus_2ht_min2 + 1;
        let sl_plus_2ht = sl_plus_2ht_min1 + 1;

        let mut wma_factors = vec![0.0; length];
        fill_wma_factors(length, &mut wma_factors);

        Ok(Self {
            smoothing_length: length,
            min_period: DEFAULT_MIN_PERIOD,
            max_period: DEFAULT_MAX_PERIOD,
            alpha_ema_qi: alpha_qi,
            alpha_ema_p: alpha_p,
            warm_up_period: p.warm_up_period.max(sl_plus_2ht),
            sl_plus_ht_min1,
            sl_plus_2ht_min2,
            sl_plus_2ht_min1,
            sl_plus_2ht,
            one_min_alpha_qi: 1.0 - alpha_qi,
            one_min_alpha_p: 1.0 - alpha_p,
            raw_values: vec![0.0; length],
            wma_factors,
            wma_smoothed: vec![0.0; HT_LENGTH],
            detrended_arr: vec![0.0; HT_LENGTH],
            delta_phase: vec![0.0; ACCUMULATION_LENGTH],
            in_phase_val: 0.0,
            quadrature_val: 0.0,
            count: 0,
            smoothed_ip_prev: 0.0,
            smoothed_q_prev: 0.0,
            phase_prev: 0.0,
            period_val: DEFAULT_MIN_PERIOD as f64,
            is_primed: false,
            is_warmed_up: false,
        })
    }
}

impl CycleEstimator for PhaseAccumulatorEstimator {
    fn smoothing_length(&self) -> usize { self.smoothing_length }
    fn smoothed(&self) -> f64 { self.wma_smoothed[0] }
    fn detrended(&self) -> f64 { self.detrended_arr[0] }
    fn quadrature(&self) -> f64 { self.quadrature_val }
    fn in_phase(&self) -> f64 { self.in_phase_val }
    fn period(&self) -> f64 { self.period_val }
    fn count(&self) -> usize { self.count }
    fn primed(&self) -> bool { self.is_warmed_up }
    fn min_period(&self) -> usize { self.min_period }
    fn max_period(&self) -> usize { self.max_period }
    fn alpha_ema_quadrature_in_phase(&self) -> f64 { self.alpha_ema_qi }
    fn alpha_ema_period(&self) -> f64 { self.alpha_ema_p }
    fn warm_up_period(&self) -> usize { self.warm_up_period }

    fn update(&mut self, sample: f64) {
        if sample.is_nan() {
            return;
        }

        push(&mut self.raw_values, sample);

        if self.is_primed {
            if !self.is_warmed_up {
                self.count += 1;
                if self.warm_up_period < self.count {
                    self.is_warmed_up = true;
                }
            }

            push(&mut self.wma_smoothed, wma(&self.raw_values, &self.wma_factors, self.smoothing_length));
            let acf = correct_amplitude(self.period_val);
            push(&mut self.detrended_arr, ht(&self.wma_smoothed) * acf);

            self.quadrature_val = ht(&self.detrended_arr) * acf;
            self.in_phase_val = self.detrended_arr[QUADRATURE_INDEX];

            let si = ema(self.alpha_ema_qi, self.one_min_alpha_qi, self.in_phase_val, self.smoothed_ip_prev);
            let sq = ema(self.alpha_ema_qi, self.one_min_alpha_qi, self.quadrature_val, self.smoothed_q_prev);
            self.smoothed_ip_prev = si;
            self.smoothed_q_prev = sq;

            let phase = instantaneous_phase(si, sq, self.phase_prev);
            push(&mut self.delta_phase, calculate_differential_phase(phase, self.phase_prev));
            self.phase_prev = phase;

            let period_previous = self.period_val;
            self.period_val = instantaneous_period(&self.delta_phase, period_previous);
            self.period_val = ema(self.alpha_ema_p, self.one_min_alpha_p, self.period_val, period_previous);
        } else {
            self.count += 1;
            if self.smoothing_length > self.count {
                return;
            }

            push(&mut self.wma_smoothed, wma(&self.raw_values, &self.wma_factors, self.smoothing_length));

            if self.sl_plus_ht_min1 > self.count {
                return;
            }

            let acf = correct_amplitude(self.period_val);
            push(&mut self.detrended_arr, ht(&self.wma_smoothed) * acf);

            if self.sl_plus_2ht_min2 > self.count {
                return;
            }

            self.quadrature_val = ht(&self.detrended_arr) * acf;
            self.in_phase_val = self.detrended_arr[QUADRATURE_INDEX];

            if self.sl_plus_2ht_min2 == self.count {
                self.smoothed_ip_prev = self.in_phase_val;
                self.smoothed_q_prev = self.quadrature_val;
                return;
            }

            let si = ema(self.alpha_ema_qi, self.one_min_alpha_qi, self.in_phase_val, self.smoothed_ip_prev);
            let sq = ema(self.alpha_ema_qi, self.one_min_alpha_qi, self.quadrature_val, self.smoothed_q_prev);
            self.smoothed_ip_prev = si;
            self.smoothed_q_prev = sq;

            let phase = instantaneous_phase(si, sq, self.phase_prev);
            push(&mut self.delta_phase, calculate_differential_phase(phase, self.phase_prev));
            self.phase_prev = phase;

            let period_previous = self.period_val;
            self.period_val = instantaneous_period(&self.delta_phase, period_previous);

            if self.sl_plus_2ht_min1 < self.count {
                self.period_val = ema(self.alpha_ema_p, self.one_min_alpha_p, self.period_val, period_previous);
                self.is_primed = true;
            }
        }
    }
}

#[cfg(test)]
mod tests {
    use super::*;
    use super::super::testdata_phase_accumulator::testdata_phase_accumulator::*;
    use std::f64::consts::PI;

    const EPSILON: f64 = 1e-8;

    fn create_default() -> PhaseAccumulatorEstimator {
        let p = CycleEstimatorParams {
            smoothing_length: 4,
            alpha_ema_quadrature_in_phase: 0.15,
            alpha_ema_period: 0.25,
            warm_up_period: 0,
        };
        PhaseAccumulatorEstimator::new(&p).unwrap()
    }

    fn create_warmup(warm_up: usize) -> PhaseAccumulatorEstimator {
        let p = CycleEstimatorParams {
            smoothing_length: 4,
            alpha_ema_quadrature_in_phase: 0.15,
            alpha_ema_period: 0.25,
            warm_up_period: warm_up,
        };
        PhaseAccumulatorEstimator::new(&p).unwrap()
    }

    fn check(index: usize, expected: f64, actual: f64) {
        if expected.is_nan() {
            return;
        }
        assert!(
            (expected - actual).abs() <= EPSILON,
            "[{}] expected {}, actual {}",
            index, expected, actual
        );
    }

    #[test]
    fn test_smoothed() {
        let mut pae = create_default();
        let inp = input();
        let exp = expected_smoothed();
        let lprimed = 3;

        for i in 0..lprimed {
            pae.update(inp[i]);
            check(i, 0.0, pae.smoothed());
        }
        for i in lprimed..inp.len() {
            pae.update(inp[i]);
            check(i, exp[i], pae.smoothed());
        }
        let previous = pae.smoothed();
        pae.update(f64::NAN);
        check(99999, previous, pae.smoothed());
    }

    #[test]
    fn test_detrended() {
        let mut pae = create_default();
        let inp = input();
        let exp = expected_detrended();
        let lprimed = 9;
        let last = 24;

        for i in 0..lprimed {
            pae.update(inp[i]);
            check(i, 0.0, pae.detrended());
        }
        for i in lprimed..last {
            pae.update(inp[i]);
            check(i, exp[i], pae.detrended());
        }
        let previous = pae.detrended();
        pae.update(f64::NAN);
        check(99999, previous, pae.detrended());
    }

    #[test]
    fn test_quadrature() {
        let mut pae = create_default();
        let inp = input();
        let exp = expected_quadrature();
        let lprimed = 15;
        let last = 24;

        for i in 0..lprimed {
            pae.update(inp[i]);
            check(i, 0.0, pae.quadrature());
        }
        for i in lprimed..last {
            pae.update(inp[i]);
            check(i, exp[i], pae.quadrature());
        }
        let previous = pae.quadrature();
        pae.update(f64::NAN);
        check(99999, previous, pae.quadrature());
    }

    #[test]
    fn test_in_phase() {
        let mut pae = create_default();
        let inp = input();
        let exp = expected_in_phase();
        let lprimed = 15;
        let last = 24;

        for i in 0..lprimed {
            pae.update(inp[i]);
            check(i, 0.0, pae.in_phase());
        }
        for i in lprimed..last {
            pae.update(inp[i]);
            check(i, exp[i], pae.in_phase());
        }
        let previous = pae.in_phase();
        pae.update(f64::NAN);
        check(99999, previous, pae.in_phase());
    }

    #[test]
    fn test_period() {
        let mut pae = create_default();
        let inp = input();
        let exp = expected_period();
        let lprimed = 18;
        let last = 18;
        let not_primed_value = 6.0;

        for i in 0..lprimed {
            pae.update(inp[i]);
            check(i, not_primed_value, pae.period());
        }
        for i in lprimed..last {
            pae.update(inp[i]);
            check(i, exp[i], pae.period());
        }
        let previous = pae.period();
        pae.update(f64::NAN);
        check(99999, previous, pae.period());
    }

    #[test]
    fn test_period_sin() {
        let period = 30.0_f64;
        let omega = 2.0 * PI / period;
        let mut pae = create_default();
        for i in 0..512 {
            pae.update((omega * i as f64).sin());
        }
        assert!((period - pae.period()).abs() <= 1e0,
            "period expected {}, actual {}", period, pae.period());
    }

    #[test]
    fn test_period_sin_min() {
        let period = 3.0_f64;
        let omega = 2.0 * PI / period;
        let mut pae = create_default();
        for i in 0..512 {
            pae.update((omega * i as f64).sin());
        }
        assert!((pae.min_period() as f64 - pae.period()).abs() <= 1e-14,
            "min period expected {}, actual {}", pae.min_period(), pae.period());
    }

    #[test]
    fn test_period_sin_max() {
        let period = 60.0_f64;
        let omega = 2.0 * PI / period;
        let mut pae = create_default();
        for i in 0..512 {
            pae.update((omega * i as f64).sin());
        }
        assert!((pae.max_period() as f64 - pae.period()).abs() <= 12.5e0,
            "max period expected {}, actual {}", pae.max_period(), pae.period());
    }

    #[test]
    fn test_primed() {
        let mut pae = create_default();
        let inp = input();
        let lprimed = 4 + 7 * 2;

        assert!(!pae.primed());
        for i in 0..lprimed {
            pae.update(inp[i]);
            assert!(!pae.primed(), "[{}] should not be primed", i + 1);
        }
        for i in lprimed..inp.len() {
            pae.update(inp[i]);
            assert!(pae.primed(), "[{}] should be primed", i + 1);
        }
    }

    #[test]
    fn test_primed_warmup() {
        let lprimed = 50;
        let mut pae = create_warmup(lprimed);
        let inp = input();

        assert!(!pae.primed());
        for i in 0..lprimed {
            pae.update(inp[i]);
            assert!(!pae.primed(), "[{}] should not be primed", i + 1);
        }
        for i in lprimed..inp.len() {
            pae.update(inp[i]);
            assert!(pae.primed(), "[{}] should be primed", i + 1);
        }
    }

    #[test]
    fn test_validation_errors() {
        let errle = "invalid cycle estimator parameters: SmoothingLength should be in range [2, 4]";
        let erraq = "invalid cycle estimator parameters: AlphaEmaQuadratureInPhase should be in range (0, 1)";
        let errap = "invalid cycle estimator parameters: AlphaEmaPeriod should be in range (0, 1)";

        // Valid default
        let p = CycleEstimatorParams { smoothing_length: 4, alpha_ema_quadrature_in_phase: 0.15, alpha_ema_period: 0.25, warm_up_period: 0 };
        assert!(PhaseAccumulatorEstimator::new(&p).is_ok());

        // Valid with warmup
        let p = CycleEstimatorParams { smoothing_length: 3, alpha_ema_quadrature_in_phase: 0.11, alpha_ema_period: 0.12, warm_up_period: 44 };
        assert!(PhaseAccumulatorEstimator::new(&p).is_ok());

        // sl=0
        let p = CycleEstimatorParams { smoothing_length: 0, alpha_ema_quadrature_in_phase: 0.15, alpha_ema_period: 0.25, warm_up_period: 0 };
        assert_eq!(PhaseAccumulatorEstimator::new(&p).unwrap_err(), errle);

        // sl=1
        let p = CycleEstimatorParams { smoothing_length: 1, alpha_ema_quadrature_in_phase: 0.15, alpha_ema_period: 0.25, warm_up_period: 0 };
        assert_eq!(PhaseAccumulatorEstimator::new(&p).unwrap_err(), errle);

        // sl=5
        let p = CycleEstimatorParams { smoothing_length: 5, alpha_ema_quadrature_in_phase: 0.15, alpha_ema_period: 0.25, warm_up_period: 0 };
        assert_eq!(PhaseAccumulatorEstimator::new(&p).unwrap_err(), errle);

        // alpha_qi = 0
        let p = CycleEstimatorParams { smoothing_length: 4, alpha_ema_quadrature_in_phase: 0.0, alpha_ema_period: 0.25, warm_up_period: 0 };
        assert_eq!(PhaseAccumulatorEstimator::new(&p).unwrap_err(), erraq);

        // alpha_qi < 0
        let p = CycleEstimatorParams { smoothing_length: 4, alpha_ema_quadrature_in_phase: -0.01, alpha_ema_period: 0.25, warm_up_period: 0 };
        assert_eq!(PhaseAccumulatorEstimator::new(&p).unwrap_err(), erraq);

        // alpha_qi = 1
        let p = CycleEstimatorParams { smoothing_length: 4, alpha_ema_quadrature_in_phase: 1.0, alpha_ema_period: 0.25, warm_up_period: 0 };
        assert_eq!(PhaseAccumulatorEstimator::new(&p).unwrap_err(), erraq);

        // alpha_qi > 1
        let p = CycleEstimatorParams { smoothing_length: 4, alpha_ema_quadrature_in_phase: 1.01, alpha_ema_period: 0.25, warm_up_period: 0 };
        assert_eq!(PhaseAccumulatorEstimator::new(&p).unwrap_err(), erraq);

        // alpha_p = 0
        let p = CycleEstimatorParams { smoothing_length: 4, alpha_ema_quadrature_in_phase: 0.15, alpha_ema_period: 0.0, warm_up_period: 0 };
        assert_eq!(PhaseAccumulatorEstimator::new(&p).unwrap_err(), errap);

        // alpha_p < 0
        let p = CycleEstimatorParams { smoothing_length: 4, alpha_ema_quadrature_in_phase: 0.15, alpha_ema_period: -0.01, warm_up_period: 0 };
        assert_eq!(PhaseAccumulatorEstimator::new(&p).unwrap_err(), errap);

        // alpha_p = 1
        let p = CycleEstimatorParams { smoothing_length: 4, alpha_ema_quadrature_in_phase: 0.15, alpha_ema_period: 1.0, warm_up_period: 0 };
        assert_eq!(PhaseAccumulatorEstimator::new(&p).unwrap_err(), errap);

        // alpha_p > 1
        let p = CycleEstimatorParams { smoothing_length: 4, alpha_ema_quadrature_in_phase: 0.15, alpha_ema_period: 1.01, warm_up_period: 0 };
        assert_eq!(PhaseAccumulatorEstimator::new(&p).unwrap_err(), errap);
    }
}
