use std::f64::consts::PI;

use super::shared::*;

#[derive(Debug)]
pub struct HomodyneDiscriminatorEstimatorUnrolled {
    smoothing_length: usize,
    min_period: usize,
    max_period: usize,
    alpha_ema_qi: f64,
    alpha_ema_p: f64,
    warm_up_period: usize,
    one_min_alpha_qi: f64,
    one_min_alpha_p: f64,
    smoothing_multiplier: f64,
    smoothed_val: f64,
    detrended_val: f64,
    in_phase_val: f64,
    quadrature_val: f64,
    adjusted_period: f64,
    count: usize,
    index: usize,
    i2_previous: f64,
    q2_previous: f64,
    re: f64,
    im: f64,
    period_val: f64,
    is_primed: bool,
    is_warmed_up: bool,
    // WMA smoother
    wma_sum: f64,
    wma_sub: f64,
    wma_input1: f64,
    wma_input2: f64,
    wma_input3: f64,
    wma_input4: f64,
    // Detrender odd/even
    detrender_odd: [f64; 3],
    detrender_previous_odd: f64,
    detrender_previous_input_odd: f64,
    detrender_even: [f64; 3],
    detrender_previous_even: f64,
    detrender_previous_input_even: f64,
    // Q1 odd/even
    q1_odd: [f64; 3],
    q1_previous_odd: f64,
    q1_previous_input_odd: f64,
    q1_even: [f64; 3],
    q1_previous_even: f64,
    q1_previous_input_even: f64,
    // I1 odd/even
    i1_previous1_odd: f64,
    i1_previous2_odd: f64,
    i1_previous1_even: f64,
    i1_previous2_even: f64,
    // jI odd/even
    ji_odd: [f64; 3],
    ji_previous_odd: f64,
    ji_previous_input_odd: f64,
    ji_even: [f64; 3],
    ji_previous_even: f64,
    ji_previous_input_even: f64,
    // jQ odd/even
    jq_odd: [f64; 3],
    jq_previous_odd: f64,
    jq_previous_input_odd: f64,
    jq_even: [f64; 3],
    jq_previous_even: f64,
    jq_previous_input_even: f64,
}

impl HomodyneDiscriminatorEstimatorUnrolled {
    pub fn new(p: &CycleEstimatorParams) -> Result<Self, String> {
        verify_parameters(p)?;
        let length = p.smoothing_length;
        let alpha_qi = p.alpha_ema_quadrature_in_phase;
        let alpha_p = p.alpha_ema_period;

        let smoothing_multiplier = match length {
            4 => 1.0 / 10.0,
            3 => 1.0 / 6.0,
            _ => 1.0 / 3.0,
        };

        const PRIMED_COUNT: usize = 23;
        let warm_up = p.warm_up_period.max(PRIMED_COUNT);

        Ok(Self {
            smoothing_length: length,
            min_period: DEFAULT_MIN_PERIOD,
            max_period: DEFAULT_MAX_PERIOD,
            alpha_ema_qi: alpha_qi,
            alpha_ema_p: alpha_p,
            warm_up_period: warm_up,
            one_min_alpha_qi: 1.0 - alpha_qi,
            one_min_alpha_p: 1.0 - alpha_p,
            smoothing_multiplier,
            smoothed_val: 0.0,
            detrended_val: 0.0,
            in_phase_val: 0.0,
            quadrature_val: 0.0,
            adjusted_period: 0.0,
            count: 0,
            index: 0,
            i2_previous: 0.0,
            q2_previous: 0.0,
            re: 0.0,
            im: 0.0,
            period_val: DEFAULT_MIN_PERIOD as f64,
            is_primed: false,
            is_warmed_up: false,
            wma_sum: 0.0,
            wma_sub: 0.0,
            wma_input1: 0.0,
            wma_input2: 0.0,
            wma_input3: 0.0,
            wma_input4: 0.0,
            detrender_odd: [0.0; 3],
            detrender_previous_odd: 0.0,
            detrender_previous_input_odd: 0.0,
            detrender_even: [0.0; 3],
            detrender_previous_even: 0.0,
            detrender_previous_input_even: 0.0,
            q1_odd: [0.0; 3],
            q1_previous_odd: 0.0,
            q1_previous_input_odd: 0.0,
            q1_even: [0.0; 3],
            q1_previous_even: 0.0,
            q1_previous_input_even: 0.0,
            i1_previous1_odd: 0.0,
            i1_previous2_odd: 0.0,
            i1_previous1_even: 0.0,
            i1_previous2_even: 0.0,
            ji_odd: [0.0; 3],
            ji_previous_odd: 0.0,
            ji_previous_input_odd: 0.0,
            ji_even: [0.0; 3],
            ji_previous_even: 0.0,
            ji_previous_input_even: 0.0,
            jq_odd: [0.0; 3],
            jq_previous_odd: 0.0,
            jq_previous_input_odd: 0.0,
            jq_even: [0.0; 3],
            jq_previous_even: 0.0,
            jq_previous_input_even: 0.0,
        })
    }
}

impl CycleEstimator for HomodyneDiscriminatorEstimatorUnrolled {
    fn smoothing_length(&self) -> usize { self.smoothing_length }
    fn smoothed(&self) -> f64 { self.smoothed_val }
    fn detrended(&self) -> f64 { self.detrended_val }
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

    #[allow(clippy::cognitive_complexity)]
    fn update(&mut self, sample: f64) {
        if sample.is_nan() {
            return;
        }

        const A: f64 = 0.0962;
        const B: f64 = 0.5769;

        let mut value: f64;

        // WMA smoothing.
        self.count += 1;

        if self.smoothing_length >= self.count {
            match self.count {
                1 => {
                    self.wma_sub = sample;
                    self.wma_input1 = sample;
                    self.wma_sum = sample;
                    return;
                }
                2 => {
                    self.wma_sub += sample;
                    self.wma_input2 = sample;
                    self.wma_sum += sample * 2.0;
                    if self.smoothing_length == 2 {
                        value = self.wma_sum * self.smoothing_multiplier;
                    } else {
                        return;
                    }
                }
                3 => {
                    self.wma_sub += sample;
                    self.wma_input3 = sample;
                    self.wma_sum += sample * 3.0;
                    if self.smoothing_length == 3 {
                        value = self.wma_sum * self.smoothing_multiplier;
                    } else {
                        return;
                    }
                }
                _ => {
                    // count == 4
                    self.wma_sub += sample;
                    self.wma_input4 = sample;
                    self.wma_sum += sample * 4.0;
                    value = self.wma_sum * self.smoothing_multiplier;
                }
            }
        } else {
            self.wma_sum -= self.wma_sub;
            self.wma_sum += sample * self.smoothing_length as f64;
            value = self.wma_sum * self.smoothing_multiplier;
            self.wma_sub += sample;
            self.wma_sub -= self.wma_input1;
            self.wma_input1 = self.wma_input2;

            match self.smoothing_length {
                4 => {
                    self.wma_input2 = self.wma_input3;
                    self.wma_input3 = self.wma_input4;
                    self.wma_input4 = sample;
                }
                3 => {
                    self.wma_input2 = self.wma_input3;
                    self.wma_input3 = sample;
                }
                _ => {
                    self.wma_input2 = sample;
                }
            }
        }

        // Detrender.
        self.smoothed_val = value;

        if !self.is_warmed_up {
            self.is_warmed_up = self.count > self.warm_up_period;
            if !self.is_primed {
                self.is_primed = self.count > 23;
            }
        }

        let mut detrender: f64;
        let ji: f64;
        let mut jq: f64;

        let temp = A * self.smoothed_val;
        self.adjusted_period = 0.075 * self.period_val + 0.54;

        if self.count % 2 == 0 {
            // Even
            match self.index {
                0 => {
                    self.index = 1;
                    detrender = -self.detrender_even[0];
                    self.detrender_even[0] = temp;
                    detrender += temp;
                    detrender -= self.detrender_previous_even;
                    self.detrender_previous_even = B * self.detrender_previous_input_even;
                    self.detrender_previous_input_even = value;
                    detrender += self.detrender_previous_even;
                    detrender *= self.adjusted_period;

                    let temp2 = A * detrender;
                    self.quadrature_val = -self.q1_even[0];
                    self.q1_even[0] = temp2;
                    self.quadrature_val += temp2;
                    self.quadrature_val -= self.q1_previous_even;
                    self.q1_previous_even = B * self.q1_previous_input_even;
                    self.q1_previous_input_even = detrender;
                    self.quadrature_val += self.q1_previous_even;
                    self.quadrature_val *= self.adjusted_period;

                    let temp3 = A * self.i1_previous2_even;
                    ji = -self.ji_even[0] + temp3 - self.ji_previous_even;
                    self.ji_even[0] = temp3;
                    self.ji_previous_even = B * self.ji_previous_input_even;
                    self.ji_previous_input_even = self.i1_previous2_even;
                    let ji = ji + self.ji_previous_even;
                    let ji = ji * self.adjusted_period;

                    let temp4 = A * self.quadrature_val;
                    jq = -self.jq_even[0];
                    self.jq_even[0] = temp4;

                    jq += temp4;
                    jq -= self.jq_previous_even;
                    self.jq_previous_even = B * self.jq_previous_input_even;
                    self.jq_previous_input_even = self.quadrature_val;
                    jq += self.jq_previous_even;
                    jq *= self.adjusted_period;

                    self.in_phase_val = self.i1_previous2_even;
                    self.i1_previous2_odd = self.i1_previous1_odd;
                    self.i1_previous1_odd = detrender;
                    self.detrended_val = detrender;

                    self.finish_unrolled(ji, jq);
                    return;
                }
                1 => {
                    self.index = 2;
                    detrender = -self.detrender_even[1];
                    self.detrender_even[1] = temp;
                    detrender += temp;
                    detrender -= self.detrender_previous_even;
                    self.detrender_previous_even = B * self.detrender_previous_input_even;
                    self.detrender_previous_input_even = value;
                    detrender += self.detrender_previous_even;
                    detrender *= self.adjusted_period;

                    let temp2 = A * detrender;
                    self.quadrature_val = -self.q1_even[1];
                    self.q1_even[1] = temp2;
                    self.quadrature_val += temp2;
                    self.quadrature_val -= self.q1_previous_even;
                    self.q1_previous_even = B * self.q1_previous_input_even;
                    self.q1_previous_input_even = detrender;
                    self.quadrature_val += self.q1_previous_even;
                    self.quadrature_val *= self.adjusted_period;

                    let temp3 = A * self.i1_previous2_even;
                    ji = -self.ji_even[1] + temp3 - self.ji_previous_even;
                    self.ji_even[1] = temp3;
                    self.ji_previous_even = B * self.ji_previous_input_even;
                    self.ji_previous_input_even = self.i1_previous2_even;
                    let ji = ji + self.ji_previous_even;
                    let ji = ji * self.adjusted_period;

                    let temp4 = A * self.quadrature_val;
                    jq = -self.jq_even[1];
                    self.jq_even[1] = temp4;

                    jq += temp4;
                    jq -= self.jq_previous_even;
                    self.jq_previous_even = B * self.jq_previous_input_even;
                    self.jq_previous_input_even = self.quadrature_val;
                    jq += self.jq_previous_even;
                    jq *= self.adjusted_period;

                    self.in_phase_val = self.i1_previous2_even;
                    self.i1_previous2_odd = self.i1_previous1_odd;
                    self.i1_previous1_odd = detrender;
                    self.detrended_val = detrender;

                    self.finish_unrolled(ji, jq);
                    return;
                }
                _ => {
                    // index == 2
                    self.index = 0;
                    detrender = -self.detrender_even[2];
                    self.detrender_even[2] = temp;
                    detrender += temp;
                    detrender -= self.detrender_previous_even;
                    self.detrender_previous_even = B * self.detrender_previous_input_even;
                    self.detrender_previous_input_even = value;
                    detrender += self.detrender_previous_even;
                    detrender *= self.adjusted_period;

                    let temp2 = A * detrender;
                    self.quadrature_val = -self.q1_even[2];
                    self.q1_even[2] = temp2;
                    self.quadrature_val += temp2;
                    self.quadrature_val -= self.q1_previous_even;
                    self.q1_previous_even = B * self.q1_previous_input_even;
                    self.q1_previous_input_even = detrender;
                    self.quadrature_val += self.q1_previous_even;
                    self.quadrature_val *= self.adjusted_period;

                    let temp3 = A * self.i1_previous2_even;
                    ji = -self.ji_even[2] + temp3 - self.ji_previous_even;
                    self.ji_even[2] = temp3;
                    self.ji_previous_even = B * self.ji_previous_input_even;
                    self.ji_previous_input_even = self.i1_previous2_even;
                    let ji = ji + self.ji_previous_even;
                    let ji = ji * self.adjusted_period;

                    let temp4 = A * self.quadrature_val;
                    jq = -self.jq_even[2];
                    self.jq_even[2] = temp4;

                    jq += temp4;
                    jq -= self.jq_previous_even;
                    self.jq_previous_even = B * self.jq_previous_input_even;
                    self.jq_previous_input_even = self.quadrature_val;
                    jq += self.jq_previous_even;
                    jq *= self.adjusted_period;

                    self.in_phase_val = self.i1_previous2_even;
                    self.i1_previous2_odd = self.i1_previous1_odd;
                    self.i1_previous1_odd = detrender;
                    self.detrended_val = detrender;

                    self.finish_unrolled(ji, jq);
                    return;
                }
            }
        } else {
            // Odd
            match self.index {
                0 => {
                    self.index = 1;
                    detrender = -self.detrender_odd[0];
                    self.detrender_odd[0] = temp;
                    detrender += temp;
                    detrender -= self.detrender_previous_odd;
                    self.detrender_previous_odd = B * self.detrender_previous_input_odd;
                    self.detrender_previous_input_odd = value;
                    detrender += self.detrender_previous_odd;
                    detrender *= self.adjusted_period;

                    let temp2 = A * detrender;
                    self.quadrature_val = -self.q1_odd[0];
                    self.q1_odd[0] = temp2;
                    self.quadrature_val += temp2;
                    self.quadrature_val -= self.q1_previous_odd;
                    self.q1_previous_odd = B * self.q1_previous_input_odd;
                    self.q1_previous_input_odd = detrender;
                    self.quadrature_val += self.q1_previous_odd;
                    self.quadrature_val *= self.adjusted_period;

                    let temp3 = A * self.i1_previous2_odd;
                    ji = -self.ji_odd[0] + temp3 - self.ji_previous_odd;
                    self.ji_odd[0] = temp3;
                    self.ji_previous_odd = B * self.ji_previous_input_odd;
                    self.ji_previous_input_odd = self.i1_previous2_odd;
                    let ji = ji + self.ji_previous_odd;
                    let ji = ji * self.adjusted_period;

                    let temp4 = A * self.quadrature_val;
                    jq = -self.jq_odd[0];
                    self.jq_odd[0] = temp4;

                    jq += temp4;
                    jq -= self.jq_previous_odd;
                    self.jq_previous_odd = B * self.jq_previous_input_odd;
                    self.jq_previous_input_odd = self.quadrature_val;
                    jq += self.jq_previous_odd;
                    jq *= self.adjusted_period;

                    self.in_phase_val = self.i1_previous2_odd;
                    self.i1_previous2_even = self.i1_previous1_even;
                    self.i1_previous1_even = detrender;
                    self.detrended_val = detrender;

                    self.finish_unrolled(ji, jq);
                    return;
                }
                1 => {
                    self.index = 2;
                    detrender = -self.detrender_odd[1];
                    self.detrender_odd[1] = temp;
                    detrender += temp;
                    detrender -= self.detrender_previous_odd;
                    self.detrender_previous_odd = B * self.detrender_previous_input_odd;
                    self.detrender_previous_input_odd = value;
                    detrender += self.detrender_previous_odd;
                    detrender *= self.adjusted_period;

                    let temp2 = A * detrender;
                    self.quadrature_val = -self.q1_odd[1];
                    self.q1_odd[1] = temp2;
                    self.quadrature_val += temp2;
                    self.quadrature_val -= self.q1_previous_odd;
                    self.q1_previous_odd = B * self.q1_previous_input_odd;
                    self.q1_previous_input_odd = detrender;
                    self.quadrature_val += self.q1_previous_odd;
                    self.quadrature_val *= self.adjusted_period;

                    let temp3 = A * self.i1_previous2_odd;
                    ji = -self.ji_odd[1] + temp3 - self.ji_previous_odd;
                    self.ji_odd[1] = temp3;
                    self.ji_previous_odd = B * self.ji_previous_input_odd;
                    self.ji_previous_input_odd = self.i1_previous2_odd;
                    let ji = ji + self.ji_previous_odd;
                    let ji = ji * self.adjusted_period;

                    let temp4 = A * self.quadrature_val;
                    jq = -self.jq_odd[1];
                    self.jq_odd[1] = temp4;

                    jq += temp4;
                    jq -= self.jq_previous_odd;
                    self.jq_previous_odd = B * self.jq_previous_input_odd;
                    self.jq_previous_input_odd = self.quadrature_val;
                    jq += self.jq_previous_odd;
                    jq *= self.adjusted_period;

                    self.in_phase_val = self.i1_previous2_odd;
                    self.i1_previous2_even = self.i1_previous1_even;
                    self.i1_previous1_even = detrender;
                    self.detrended_val = detrender;

                    self.finish_unrolled(ji, jq);
                    return;
                }
                _ => {
                    // index == 2
                    self.index = 0;
                    detrender = -self.detrender_odd[2];
                    self.detrender_odd[2] = temp;
                    detrender += temp;
                    detrender -= self.detrender_previous_odd;
                    self.detrender_previous_odd = B * self.detrender_previous_input_odd;
                    self.detrender_previous_input_odd = value;
                    detrender += self.detrender_previous_odd;
                    detrender *= self.adjusted_period;

                    let temp2 = A * detrender;
                    self.quadrature_val = -self.q1_odd[2];
                    self.q1_odd[2] = temp2;
                    self.quadrature_val += temp2;
                    self.quadrature_val -= self.q1_previous_odd;
                    self.q1_previous_odd = B * self.q1_previous_input_odd;
                    self.q1_previous_input_odd = detrender;
                    self.quadrature_val += self.q1_previous_odd;
                    self.quadrature_val *= self.adjusted_period;

                    let temp3 = A * self.i1_previous2_odd;
                    ji = -self.ji_odd[2] + temp3 - self.ji_previous_odd;
                    self.ji_odd[2] = temp3;
                    self.ji_previous_odd = B * self.ji_previous_input_odd;
                    self.ji_previous_input_odd = self.i1_previous2_odd;
                    let ji = ji + self.ji_previous_odd;
                    let ji = ji * self.adjusted_period;

                    let temp4 = A * self.quadrature_val;
                    jq = -self.jq_odd[2];
                    self.jq_odd[2] = temp4;

                    jq += temp4;
                    jq -= self.jq_previous_odd;
                    self.jq_previous_odd = B * self.jq_previous_input_odd;
                    self.jq_previous_input_odd = self.quadrature_val;
                    jq += self.jq_previous_odd;
                    jq *= self.adjusted_period;

                    self.in_phase_val = self.i1_previous2_odd;
                    self.i1_previous2_even = self.i1_previous1_even;
                    self.i1_previous1_even = detrender;
                    self.detrended_val = detrender;

                    self.finish_unrolled(ji, jq);
                    return;
                }
            }
        }
    }
}

impl HomodyneDiscriminatorEstimatorUnrolled {
    fn finish_unrolled(&mut self, ji: f64, jq: f64) {
        const MIN_PREV_FACTOR: f64 = 0.67;
        const MAX_PREV_FACTOR: f64 = 1.5;

        let i2 = self.in_phase_val - jq;
        let q2 = self.quadrature_val + ji;

        let i2 = self.alpha_ema_qi * i2 + self.one_min_alpha_qi * self.i2_previous;
        let q2 = self.alpha_ema_qi * q2 + self.one_min_alpha_qi * self.q2_previous;

        self.re = self.alpha_ema_qi * (i2 * self.i2_previous + q2 * self.q2_previous)
            + self.one_min_alpha_qi * self.re;
        self.im = self.alpha_ema_qi * (i2 * self.q2_previous - q2 * self.i2_previous)
            + self.one_min_alpha_qi * self.im;
        self.q2_previous = q2;
        self.i2_previous = i2;
        let temp = self.period_val;

        let period_new = 2.0 * PI / self.im.atan2(self.re);
        if !period_new.is_nan() && !period_new.is_infinite() {
            self.period_val = period_new;
        }

        let value = MAX_PREV_FACTOR * temp;
        if self.period_val > value {
            self.period_val = value;
        } else {
            let value = MIN_PREV_FACTOR * temp;
            if self.period_val < value {
                self.period_val = value;
            }
        }

        if self.period_val < DEFAULT_MIN_PERIOD as f64 {
            self.period_val = DEFAULT_MIN_PERIOD as f64;
        } else if self.period_val > DEFAULT_MAX_PERIOD as f64 {
            self.period_val = DEFAULT_MAX_PERIOD as f64;
        }

        self.period_val = self.alpha_ema_p * self.period_val + self.one_min_alpha_p * temp;
    }
}

#[cfg(test)]
mod tests {
    use super::*;
    use super::super::testdata_homodyne_discriminator_unrolled::testdata_homodyne_discriminator_unrolled::*;
    use std::f64::consts::PI;
    const EPSILON: f64 = 1e-8;

    fn create_default() -> HomodyneDiscriminatorEstimatorUnrolled {
        let p = CycleEstimatorParams { smoothing_length: 4, alpha_ema_quadrature_in_phase: 0.2, alpha_ema_period: 0.2, warm_up_period: 0 };
        HomodyneDiscriminatorEstimatorUnrolled::new(&p).unwrap()
    }

    fn create_warmup(warm_up: usize) -> HomodyneDiscriminatorEstimatorUnrolled {
        let p = CycleEstimatorParams { smoothing_length: 4, alpha_ema_quadrature_in_phase: 0.2, alpha_ema_period: 0.2, warm_up_period: warm_up };
        HomodyneDiscriminatorEstimatorUnrolled::new(&p).unwrap()
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
        let mut hdeu = create_default();
        let inp = input();
        let exp = expected_smoothed();
        let lprimed = 3;

        for i in 0..lprimed {
            hdeu.update(inp[i]);
            check(i, 0.0, hdeu.smoothed());
        }
        for i in lprimed..inp.len() {
            hdeu.update(inp[i]);
            check(i, exp[i], hdeu.smoothed());
        }
        let previous = hdeu.smoothed();
        hdeu.update(f64::NAN);
        check(99999, previous, hdeu.smoothed());
    }

    #[test]
    fn test_detrended() {
        let mut hdeu = create_default();
        let inp = input();
        let exp = expected_detrended();
        let lprimed = 3;

        for i in 0..lprimed {
            hdeu.update(inp[i]);
            check(i, 0.0, hdeu.detrended());
        }
        for i in lprimed..inp.len() {
            hdeu.update(inp[i]);
            check(i, exp[i], hdeu.detrended());
        }
        let previous = hdeu.detrended();
        hdeu.update(f64::NAN);
        check(99999, previous, hdeu.detrended());
    }

    #[test]
    fn test_quadrature() {
        let mut hdeu = create_default();
        let inp = input();
        let exp = expected_quadrature();
        let lprimed = 3;

        for i in 0..lprimed {
            hdeu.update(inp[i]);
            check(i, 0.0, hdeu.quadrature());
        }
        for i in lprimed..inp.len() {
            hdeu.update(inp[i]);
            check(i, exp[i], hdeu.quadrature());
        }
        let previous = hdeu.quadrature();
        hdeu.update(f64::NAN);
        check(99999, previous, hdeu.quadrature());
    }

    #[test]
    fn test_in_phase() {
        let mut hdeu = create_default();
        let inp = input();
        let exp = expected_in_phase();
        let lprimed = 3;

        for i in 0..lprimed {
            hdeu.update(inp[i]);
            check(i, 0.0, hdeu.in_phase());
        }
        for i in lprimed..inp.len() {
            hdeu.update(inp[i]);
            check(i, exp[i], hdeu.in_phase());
        }
        let previous = hdeu.in_phase();
        hdeu.update(f64::NAN);
        check(99999, previous, hdeu.in_phase());
    }

    #[test]
    fn test_period() {
        let mut hdeu = create_default();
        let inp = input();
        let exp = expected_period();
        let lprimed = 3;
        let not_primed_value = 6.0;

        for i in 0..lprimed {
            hdeu.update(inp[i]);
            check(i, not_primed_value, hdeu.period());
        }
        for i in lprimed..inp.len() {
            hdeu.update(inp[i]);
            check(i, exp[i], hdeu.period());
        }
        let previous = hdeu.period();
        hdeu.update(f64::NAN);
        check(99999, previous, hdeu.period());
    }

    #[test]
    fn test_period_sin() {
        let period = 30.0_f64;
        let omega = 2.0 * PI / period;
        let mut hdeu = create_default();
        for i in 0..512 {
            hdeu.update((omega * i as f64).sin());
        }
        assert!((period - hdeu.period()).abs() <= 1e-2,
            "period expected {}, actual {}", period, hdeu.period());
    }

    #[test]
    fn test_period_sin_min() {
        let period = 3.0_f64;
        let omega = 2.0 * PI / period;
        let mut hdeu = create_default();
        for i in 0..512 {
            hdeu.update((omega * i as f64).sin());
        }
        assert!((hdeu.min_period() as f64 - hdeu.period()).abs() <= 1e-14,
            "min period expected {}, actual {}", hdeu.min_period(), hdeu.period());
    }

    #[test]
    fn test_period_sin_max() {
        let period = 60.0_f64;
        let omega = 2.0 * PI / period;
        let mut hdeu = create_default();
        for i in 0..512 {
            hdeu.update((omega * i as f64).sin());
        }
        assert!((hdeu.max_period() as f64 - hdeu.period()).abs() <= 1e-14,
            "max period expected {}, actual {}", hdeu.max_period(), hdeu.period());
    }

    #[test]
    fn test_primed() {
        let mut hdeu = create_default();
        let inp = input();
        let lprimed = 2 + 7 * 3;

        assert!(!hdeu.primed());
        for i in 0..lprimed {
            hdeu.update(inp[i]);
            assert!(!hdeu.primed(), "[{}] should not be primed", i + 1);
        }
        for i in lprimed..inp.len() {
            hdeu.update(inp[i]);
            assert!(hdeu.primed(), "[{}] should be primed", i + 1);
        }
    }

    #[test]
    fn test_primed_warmup() {
        let lprimed = 50;
        let mut hdeu = create_warmup(lprimed);
        let inp = input();

        assert!(!hdeu.primed());
        for i in 0..lprimed {
            hdeu.update(inp[i]);
            assert!(!hdeu.primed(), "[{}] should not be primed", i + 1);
        }
        for i in lprimed..inp.len() {
            hdeu.update(inp[i]);
            assert!(hdeu.primed(), "[{}] should be primed", i + 1);
        }
    }

    #[test]
    fn test_validation_errors() {
        let errle = "invalid cycle estimator parameters: SmoothingLength should be in range [2, 4]";
        let erraq = "invalid cycle estimator parameters: AlphaEmaQuadratureInPhase should be in range (0, 1)";
        let errap = "invalid cycle estimator parameters: AlphaEmaPeriod should be in range (0, 1)";

        // Valid default
        let p = CycleEstimatorParams { smoothing_length: 4, alpha_ema_quadrature_in_phase: 0.2, alpha_ema_period: 0.2, warm_up_period: 0 };
        assert!(HomodyneDiscriminatorEstimatorUnrolled::new(&p).is_ok());

        // Valid with warmup
        let p = CycleEstimatorParams { smoothing_length: 3, alpha_ema_quadrature_in_phase: 0.11, alpha_ema_period: 0.12, warm_up_period: 44 };
        assert!(HomodyneDiscriminatorEstimatorUnrolled::new(&p).is_ok());

        // sl=0
        let p = CycleEstimatorParams { smoothing_length: 0, alpha_ema_quadrature_in_phase: 0.2, alpha_ema_period: 0.2, warm_up_period: 0 };
        assert_eq!(HomodyneDiscriminatorEstimatorUnrolled::new(&p).unwrap_err(), errle);

        // sl=1
        let p = CycleEstimatorParams { smoothing_length: 1, alpha_ema_quadrature_in_phase: 0.2, alpha_ema_period: 0.2, warm_up_period: 0 };
        assert_eq!(HomodyneDiscriminatorEstimatorUnrolled::new(&p).unwrap_err(), errle);

        // sl=5
        let p = CycleEstimatorParams { smoothing_length: 5, alpha_ema_quadrature_in_phase: 0.2, alpha_ema_period: 0.2, warm_up_period: 0 };
        assert_eq!(HomodyneDiscriminatorEstimatorUnrolled::new(&p).unwrap_err(), errle);

        // alpha_qi = 0
        let p = CycleEstimatorParams { smoothing_length: 4, alpha_ema_quadrature_in_phase: 0.0, alpha_ema_period: 0.2, warm_up_period: 0 };
        assert_eq!(HomodyneDiscriminatorEstimatorUnrolled::new(&p).unwrap_err(), erraq);

        // alpha_qi < 0
        let p = CycleEstimatorParams { smoothing_length: 4, alpha_ema_quadrature_in_phase: -0.01, alpha_ema_period: 0.2, warm_up_period: 0 };
        assert_eq!(HomodyneDiscriminatorEstimatorUnrolled::new(&p).unwrap_err(), erraq);

        // alpha_qi = 1
        let p = CycleEstimatorParams { smoothing_length: 4, alpha_ema_quadrature_in_phase: 1.0, alpha_ema_period: 0.2, warm_up_period: 0 };
        assert_eq!(HomodyneDiscriminatorEstimatorUnrolled::new(&p).unwrap_err(), erraq);

        // alpha_qi > 1
        let p = CycleEstimatorParams { smoothing_length: 4, alpha_ema_quadrature_in_phase: 1.01, alpha_ema_period: 0.2, warm_up_period: 0 };
        assert_eq!(HomodyneDiscriminatorEstimatorUnrolled::new(&p).unwrap_err(), erraq);

        // alpha_p = 0
        let p = CycleEstimatorParams { smoothing_length: 4, alpha_ema_quadrature_in_phase: 0.2, alpha_ema_period: 0.0, warm_up_period: 0 };
        assert_eq!(HomodyneDiscriminatorEstimatorUnrolled::new(&p).unwrap_err(), errap);

        // alpha_p < 0
        let p = CycleEstimatorParams { smoothing_length: 4, alpha_ema_quadrature_in_phase: 0.2, alpha_ema_period: -0.01, warm_up_period: 0 };
        assert_eq!(HomodyneDiscriminatorEstimatorUnrolled::new(&p).unwrap_err(), errap);

        // alpha_p = 1
        let p = CycleEstimatorParams { smoothing_length: 4, alpha_ema_quadrature_in_phase: 0.2, alpha_ema_period: 1.0, warm_up_period: 0 };
        assert_eq!(HomodyneDiscriminatorEstimatorUnrolled::new(&p).unwrap_err(), errap);

        // alpha_p > 1
        let p = CycleEstimatorParams { smoothing_length: 4, alpha_ema_quadrature_in_phase: 0.2, alpha_ema_period: 1.01, warm_up_period: 0 };
        assert_eq!(HomodyneDiscriminatorEstimatorUnrolled::new(&p).unwrap_err(), errap);
    }
}
