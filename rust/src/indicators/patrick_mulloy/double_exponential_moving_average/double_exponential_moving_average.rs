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

/// Parameters to create a DEMA from length.
pub struct DoubleExponentialMovingAverageLengthParams {
    pub length: i64,
    pub first_is_average: bool,
    pub bar_component: Option<BarComponent>,
    pub quote_component: Option<QuoteComponent>,
    pub trade_component: Option<TradeComponent>,
}

impl Default for DoubleExponentialMovingAverageLengthParams {
    fn default() -> Self {
        Self {
            length: 20,
            first_is_average: true,
            bar_component: None,
            quote_component: None,
            trade_component: None,
        }
    }
}

/// Parameters to create a DEMA from smoothing factor.
pub struct DoubleExponentialMovingAverageSmoothingFactorParams {
    pub smoothing_factor: f64,
    pub first_is_average: bool,
    pub bar_component: Option<BarComponent>,
    pub quote_component: Option<QuoteComponent>,
    pub trade_component: Option<TradeComponent>,
}

impl Default for DoubleExponentialMovingAverageSmoothingFactorParams {
    fn default() -> Self {
        Self {
            smoothing_factor: 0.0952,
            first_is_average: true,
            bar_component: None,
            quote_component: None,
            trade_component: None,
        }
    }
}

// ---------------------------------------------------------------------------
// Output
// ---------------------------------------------------------------------------

#[derive(Debug, Clone, Copy, PartialEq, Eq)]
#[repr(u8)]
pub enum DoubleExponentialMovingAverageOutput {
    Value = 1,
}

// ---------------------------------------------------------------------------
// Indicator
// ---------------------------------------------------------------------------

/// Computes the Double Exponential Moving Average (DEMA).
///
/// DEMA = 2*EMA1 - EMA2, where EMA2 = EMA(EMA1).
pub struct DoubleExponentialMovingAverage {
    line: LineIndicator,
    smoothing_factor: f64,
    sum: f64,
    ema1: f64,
    ema2: f64,
    length: i64,
    length2: i64,
    count: i64,
    first_is_average: bool,
    primed: bool,
}

impl DoubleExponentialMovingAverage {
    pub fn new_from_length(params: &DoubleExponentialMovingAverageLengthParams) -> Result<Self, String> {
        Self::new_internal(params.length, f64::NAN, params.first_is_average,
            params.bar_component, params.quote_component, params.trade_component)
    }

    pub fn new_from_smoothing_factor(params: &DoubleExponentialMovingAverageSmoothingFactorParams) -> Result<Self, String> {
        Self::new_internal(0, params.smoothing_factor, params.first_is_average,
            params.bar_component, params.quote_component, params.trade_component)
    }

    fn new_internal(
        length: i64,
        alpha: f64,
        first_is_average: bool,
        bc_opt: Option<BarComponent>,
        qc_opt: Option<QuoteComponent>,
        tc_opt: Option<TradeComponent>,
    ) -> Result<Self, String> {
        const INVALID: &str = "invalid double exponential moving average parameters";
        const EPSILON: f64 = 0.00000001;

        let bc = bc_opt.unwrap_or(DEFAULT_BAR_COMPONENT);
        let qc = qc_opt.unwrap_or(DEFAULT_QUOTE_COMPONENT);
        let tc = tc_opt.unwrap_or(DEFAULT_TRADE_COMPONENT);

        let (actual_length, actual_alpha, mnemonic);

        if alpha.is_nan() {
            if length < 1 {
                return Err(format!("{}: length should be positive", INVALID));
            }
            actual_alpha = 2.0 / (1 + length) as f64;
            actual_length = length;
            mnemonic = format!("dema({}{})", length, component_triple_mnemonic(bc, qc, tc));
        } else {
            if alpha < 0.0 || alpha > 1.0 {
                return Err(format!("{}: smoothing factor should be in range [0, 1]", INVALID));
            }
            let clamped = if alpha < EPSILON { EPSILON } else { alpha };
            actual_length = (2.0_f64 / clamped).round() as i64 - 1;
            actual_alpha = clamped;
            mnemonic = format!("dema({}, {:.8}{})", actual_length, clamped, component_triple_mnemonic(bc, qc, tc));
        }

        let bar_func = bar_component_value(bc);
        let quote_func = quote_component_value(qc);
        let trade_func = trade_component_value(tc);

        let description = format!("Double exponential moving average {}", mnemonic);
        let line = LineIndicator::new(mnemonic, description, bar_func, quote_func, trade_func);

        Ok(Self {
            line,
            smoothing_factor: actual_alpha,
            sum: 0.0,
            ema1: 0.0,
            ema2: 0.0,
            length: actual_length,
            length2: 2 * actual_length - 1,
            count: 0,
            first_is_average,
            primed: false,
        })
    }

    /// Core update logic.
    pub fn update(&mut self, sample: f64) -> f64 {
        if sample.is_nan() {
            return sample;
        }

        if self.primed {
            let sf = self.smoothing_factor;
            let mut v1 = self.ema1;
            let mut v2 = self.ema2;
            v1 += (sample - v1) * sf;
            v2 += (v1 - v2) * sf;
            self.ema1 = v1;
            self.ema2 = v2;
            return 2.0 * v1 - v2;
        }

        self.count += 1;

        if self.first_is_average {
            if self.count == 1 {
                self.sum = sample;
            } else if self.length >= self.count {
                self.sum += sample;
                if self.length == self.count {
                    self.ema1 = self.sum / self.length as f64;
                    self.sum = self.ema1;
                }
            } else {
                self.ema1 += (sample - self.ema1) * self.smoothing_factor;
                self.sum += self.ema1;

                if self.length2 == self.count {
                    self.primed = true;
                    self.ema2 = self.sum / self.length as f64;
                    return 2.0 * self.ema1 - self.ema2;
                }
            }
        } else {
            // Metastock
            if self.count == 1 {
                self.ema1 = sample;
            } else if self.length >= self.count {
                self.ema1 += (sample - self.ema1) * self.smoothing_factor;
                if self.length == self.count {
                    self.ema2 = self.ema1;
                }
            } else {
                self.ema1 += (sample - self.ema1) * self.smoothing_factor;
                self.ema2 += (self.ema1 - self.ema2) * self.smoothing_factor;

                if self.length2 == self.count {
                    self.primed = true;
                    return 2.0 * self.ema1 - self.ema2;
                }
            }
        }

        f64::NAN
    }
}

impl Indicator for DoubleExponentialMovingAverage {
    fn is_primed(&self) -> bool {
        self.primed
    }

    fn metadata(&self) -> Metadata {
        build_metadata(
            Identifier::DoubleExponentialMovingAverage,
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
    use crate::entities::quote_component::QuoteComponent;
    use crate::entities::trade_component::TradeComponent;
    use crate::indicators::core::outputs::shape::Shape;

    #[allow(clippy::excessive_precision)]
    fn test_input() -> Vec<f64> {
        vec![
            91.500000, 94.815000, 94.375000, 95.095000, 93.780000, 94.625000, 92.530000, 92.750000, 90.315000, 92.470000,
            96.125000, 97.250000, 98.500000, 89.875000, 91.000000, 92.815000, 89.155000, 89.345000, 91.625000, 89.875000,
            88.375000, 87.625000, 84.780000, 83.000000, 83.500000, 81.375000, 84.440000, 89.250000, 86.375000, 86.250000,
            85.250000, 87.125000, 85.815000, 88.970000, 88.470000, 86.875000, 86.815000, 84.875000, 84.190000, 83.875000,
            83.375000, 85.500000, 89.190000, 89.440000, 91.095000, 90.750000, 91.440000, 89.000000, 91.000000, 90.500000,
            89.030000, 88.815000, 84.280000, 83.500000, 82.690000, 84.750000, 85.655000, 86.190000, 88.940000, 89.280000,
            88.625000, 88.500000, 91.970000, 91.500000, 93.250000, 93.500000, 93.155000, 91.720000, 90.000000, 89.690000,
            88.875000, 85.190000, 83.375000, 84.875000, 85.940000, 97.250000, 99.875000, 104.940000, 106.000000, 102.500000,
            102.405000, 104.595000, 106.125000, 106.000000, 106.065000, 104.625000, 108.625000, 109.315000, 110.500000,
            112.750000, 123.000000, 119.625000, 118.750000, 119.250000, 117.940000, 116.440000, 115.190000, 111.875000,
            110.595000, 118.125000, 116.000000, 116.000000, 112.000000, 113.750000, 112.940000, 116.000000, 120.500000,
            116.620000, 117.000000, 115.250000, 114.310000, 115.500000, 115.870000, 120.690000, 120.190000, 120.750000,
            124.750000, 123.370000, 122.940000, 122.560000, 123.120000, 122.560000, 124.620000, 129.250000, 131.000000,
            132.250000, 131.000000, 132.810000, 134.000000, 137.380000, 137.810000, 137.880000, 137.250000, 136.310000,
            136.250000, 134.630000, 128.250000, 129.000000, 123.870000, 124.810000, 123.000000, 126.250000, 128.380000,
            125.370000, 125.690000, 122.250000, 119.370000, 118.500000, 123.190000, 123.500000, 122.190000, 119.310000,
            123.310000, 121.120000, 123.370000, 127.370000, 128.500000, 123.870000, 122.940000, 121.750000, 124.440000,
            122.000000, 122.370000, 122.940000, 124.000000, 123.190000, 124.560000, 127.250000, 125.870000, 128.860000,
            132.000000, 130.750000, 134.750000, 135.000000, 132.380000, 133.310000, 131.940000, 130.000000, 125.370000,
            130.130000, 127.120000, 125.190000, 122.000000, 125.000000, 123.000000, 123.500000, 120.060000, 121.000000,
            117.750000, 119.870000, 122.000000, 119.190000, 116.370000, 113.500000, 114.250000, 110.000000, 105.060000,
            107.000000, 107.870000, 107.000000, 107.120000, 107.000000, 91.000000, 93.940000, 93.870000, 95.500000, 93.000000,
            94.940000, 98.250000, 96.750000, 94.810000, 94.370000, 91.560000, 90.250000, 93.940000, 93.620000, 97.000000,
            95.000000, 95.870000, 94.060000, 94.620000, 93.750000, 98.000000, 103.940000, 107.870000, 106.060000, 104.500000,
            105.000000, 104.190000, 103.060000, 103.420000, 105.270000, 111.870000, 116.000000, 116.620000, 118.280000,
            113.370000, 109.000000, 109.700000, 109.250000, 107.000000, 109.190000, 110.000000, 109.200000, 110.120000,
            108.000000, 108.620000, 109.750000, 109.810000, 109.000000, 108.750000, 107.870000,
        ]
    }

    fn test_tasc_input() -> Vec<f64> {
        vec![
            451.61, 455.20, 453.29, 446.48, 446.17, 440.86, 441.88, 451.61, 438.43, 406.33,
            328.45, 323.30, 326.39, 322.97, 312.49, 316.47, 292.92, 302.57, 326.91, 333.19,
            330.47, 338.47, 340.14, 337.59, 344.66, 345.75, 353.27, 357.12, 363.40, 373.37,
            375.48, 381.58, 372.54, 374.64, 381.83, 373.90, 374.04, 379.23, 379.42, 372.48,
            366.03, 366.66, 376.86, 386.25, 386.92, 391.62, 394.69, 394.33, 394.59, 387.35,
            387.33, 387.71, 378.95, 377.42, 374.43, 376.51, 381.60, 383.91, 384.98, 387.71,
            385.67, 384.59, 388.59, 382.79, 381.02, 373.76, 367.58, 366.38, 373.91, 375.21,
            375.80, 377.34, 381.38, 384.74, 387.09, 391.66, 397.96, 406.35, 402.37, 407.19,
            399.96, 403.99, 405.90, 402.19, 400.94, 406.73, 410.71, 417.68, 423.76, 427.55,
            430.74, 434.83, 442.05, 445.21, 451.63, 453.65, 447.21, 448.36, 435.29, 442.42,
            448.90, 449.29, 452.82, 457.42, 462.48, 461.97, 466.75, 471.34, 471.31, 467.57,
            468.07, 472.92, 483.64, 467.29, 470.67, 452.76, 452.97, 456.19, 456.72, 456.63,
            457.10, 456.22, 443.84, 444.57, 454.82, 458.22, 439.72, 440.88, 421.33, 422.21,
            428.84, 429.01, 419.52, 431.02, 436.76, 442.16, 437.25, 435.54, 430.90, 436.31,
            425.79, 417.98, 428.61, 438.10, 448.31, 453.69, 462.13, 460.87, 467.55, 459.33,
            462.29, 460.53, 468.44, 455.27, 442.59, 417.46, 408.03, 393.49, 367.33, 381.21,
            380.38, 374.42, 362.25, 344.51, 347.36, 327.55, 337.36, 334.36, 336.45, 341.95,
            350.85, 349.04, 359.06, 371.54, 368.83, 373.60, 371.20, 367.24, 361.80, 376.99,
            394.28, 417.69, 436.80, 448.71, 448.95, 456.73, 475.11, 466.29, 464.15, 482.30,
            495.79, 501.62, 501.19, 494.64, 492.10, 493.42, 481.38, 492.67, 506.11, 498.54,
            495.07, 485.82, 475.92, 474.05, 492.71, 497.55, 492.69, 505.67, 508.31, 512.47,
            521.06, 525.68, 516.94, 516.71, 527.19, 524.48, 520.40, 519.05, 538.90, 525.13,
            540.93, 548.08, 531.29, 523.47, 523.90, 536.30, 540.90, 535.76, 565.71, 592.65,
            615.70, 626.85, 624.68, 620.21, 634.95, 636.43, 629.75, 633.47, 615.95, 618.62,
            624.28, 604.67, 590.01, 584.24, 591.81, 572.89, 578.14, 585.76, 574.43, 580.30,
            585.31, 585.43, 569.52, 554.20, 547.84, 563.35, 567.80, 570.52, 565.61, 580.83,
            573.74, 573.18, 563.70, 563.56, 573.44, 583.01, 589.12, 577.20, 571.63, 570.52,
            582.61, 597.30, 605.17, 616.82, 637.16, 642.60, 649.49, 661.60, 655.79, 661.29,
            665.88, 676.95, 677.21, 697.15, 701.64, 696.34, 700.98, 690.54, 663.61, 670.77,
            681.37, 692.78, 682.72, 681.54, 669.85, 665.26, 666.78, 658.41, 661.42, 681.44,
            676.37, 694.29, 700.53, 702.01, 693.19, 689.59, 694.81, 704.49, 705.81, 699.73,
            700.24, 704.70, 718.08, 718.26, 730.96, 734.07,
        ]
    }

    #[allow(clippy::excessive_precision)]
    fn test_tasc_expected() -> Vec<f64> {
        vec![
            448.35153846153900, 444.04591642925000, 440.14325876683900, 435.89797197112700, 432.27216108607300, 428.4823387986880,
            425.44660726945700, 424.32221956733400, 421.62769118311400, 414.83033726696900, 397.84560812203200, 382.1684638859920,
            368.84565318219400, 356.70263630619500, 344.61177746844000, 334.63735513171700, 322.62654116603900, 313.5428856197950,
            309.16728032640400, 306.40916999763100, 303.80344210260600, 302.85705842249500, 302.47933821157700, 301.9983608663500,
            302.78990505807500, 303.85376719743900, 306.06855288785400, 308.77561303182900, 312.25765711668100, 316.9458970680750,
            321.57591953906100, 326.71142072727700, 330.12569035653100, 333.58519070887400, 337.80215498511300, 340.5277984044680,
            343.06768310059700, 346.15558049796200, 349.01153873028100, 350.63684427250100, 351.22996990596900, 351.9095633782500,
            354.02828445734900, 357.30961166805800, 360.37719885601600, 363.82330749394400, 367.36827960468400, 370.5060772052240,
            373.36404591786600, 374.89789264177600, 376.27839007919300, 377.57679520685900, 377.49585696812600, 377.2161727060120,
            376.55094999730900, 376.26572641304400, 376.74867888791700, 377.51889916397200, 378.36695413774200, 379.5197751830860,
            380.26204411670400, 380.77401929783400, 381.80453838016000, 381.89783916607600, 381.73021452208800, 380.5468863182740,
            378.61348759035000, 376.72417599183900, 376.12187633061000, 375.77882448754900, 375.56576179912300, 375.6040783595300,
            376.22280474748100, 377.26090183044400, 378.52719410689100, 380.31178107033600, 382.80280950363100, 386.2185592056820,
            388.68968311097500, 391.57096391806000, 393.09545863514500, 395.01505721352400, 396.98253167529700, 398.1889807999980,
            399.06694346081800, 400.65594056530300, 402.61936913600500, 405.34143412934500, 408.61024741373400, 412.0365400242810,
            415.51424471768900, 419.16284342072000, 423.40690763980300, 427.59861203811200, 432.20484084578100, 436.5487004361930,
            439.44795618689600, 442.14025113812000, 442.61800606262200, 444.00456957917000, 446.10907426987700, 447.9842994202380,
            450.10276104066900, 452.59015125823600, 455.47105424993300, 457.90666488479000, 460.69902426372800, 463.7810386872970,
            466.46056524419900, 468.25088326489400, 469.85338700957900, 471.90939701285600, 475.20696719666600, 475.7478636247790,
            476.64738188687900, 474.83071437832700, 473.18207233394000, 472.11709350392900, 471.19226477584600, 470.3071248387330,
            469.54026481122900, 468.68900476125000, 466.12462403384000, 463.90794851838500, 463.36367692547200, 463.3352208853190,
            460.64352553739400, 458.38651086715700, 453.56640083647600, 449.38259511285200, 446.59240449870100, 444.1271112863010,
            440.57671800386600, 439.05789403467900, 438.53051878518600, 438.83903510058600, 438.42129375178200, 437.8127704074620,
            436.61664122799700, 436.33158320045600, 434.58627785064100, 431.92751110847700, 431.08800595307800, 431.7085234483640,
            433.73215766065500, 436.31460313670400, 439.82803250568000, 442.78284900168400, 446.36834819783800, 448.3858840068980,
            450.59792288842700, 452.30746007553500, 454.94667941935000, 455.40457430577800, 453.98746320398800, 449.1249058769170,
            443.43700745040500, 436.29198063782400, 426.19940220104200, 419.20376553191500, 412.87925527082900, 406.4284240852840,
            398.98606031668800, 389.87305401128500, 382.21779865199500, 372.63726944189300, 365.57562357895500, 358.9385986699680,
            353.41171943430700, 349.36487937940300, 347.12436559088100, 344.96571184576400, 344.56683470162000, 346.0847969047800,
            347.13926492858700, 348.84281150631800, 350.09678320049600, 350.72271622664100, 350.57390590624900, 352.6749455854390,
            357.07627619881700, 364.39354167010100, 373.68560854646800, 383.69745952526600, 392.67196847402800, 401.7848631504070,
            412.52129623701200, 420.81120775166300, 427.86271648369300, 436.69665625267000, 446.44505917530300, 455.9052122479300,
            464.20526001391000, 470.58971654845400, 475.83296273212400, 480.60577070724900, 483.05035493444200, 486.7476714258790,
            491.86738627011100, 495.25492019681100, 497.68217537294300, 498.42920662615400, 497.58813352156200, 496.4813801272450,
            498.07149481592000, 500.09582362409200, 501.12405710915600, 503.81106629945500, 506.50156901377500, 509.4116347481980,
            513.14878775920200, 517.05529733511600, 519.20495014115900, 521.00270487807500, 524.01551076430400, 526.2293264513210,
            527.53566046857800, 528.42388559740100, 531.96578568399100, 533.07598323047000, 536.23825968263300, 539.9951729115380,
            540.86520045120800, 540.44295958768200, 540.04917031737300, 541.39194812915300, 543.17146574211000, 543.9513590824800,
            548.84776223941000, 556.98145349335100, 567.43883513908500, 578.25983402526700, 587.49413439256900, 594.9774007739050,
            603.63437180989800, 611.43929555510500, 617.31490619781300, 622.95122585062700, 625.34175380619100, 627.7206193164830,
            630.51764053276000, 630.08382393199600, 627.48102410866800, 624.22079916214600, 622.28625349928600, 617.7613206415570,
            614.38468371611500, 612.37795228502400, 608.89511421428100, 606.55745406191900, 605.12394595290100, 603.8033446470740,
            600.30058855005100, 594.94307837880300, 589.21899584033200, 586.29714732194300, 584.30106513637300, 582.8872439751560,
            580.90678087274000, 581.29620835902200, 580.61608126092800, 579.91534852930600, 577.92521224483300, 576.1209367622890,
            575.91464550901100, 577.08945838235200, 579.00082986700000, 578.99453886348500, 578.18466172678900, 577.2959886714500,
            578.22170195777100, 581.13509399634300, 584.84432478937700, 589.79790660897800, 597.09481554013700, 604.3466515069590,
            611.75677616160500, 620.04441732005000, 626.54887171613800, 633.07047351406800, 639.47136354102200, 646.6832713773490,
            653.06948442562700, 661.52172642490200, 669.60421468806700, 675.95330591677200, 682.17046475665300, 686.1144724602180,
            685.67825370149000, 686.20461784550800, 688.08343693827800, 691.28514912652000, 692.60313200874900, 693.5105212726140,
            692.55507025042700, 690.95652817927700, 689.66188495954300, 687.23231081390200, 685.42060562284400, 686.5917324172170,
            686.84423902464400, 689.55973926270000, 692.80400358606700, 695.83896817833100, 697.21654979349000, 697.8619301917080,
            699.11454339900500, 701.54579865410000, 703.83467566345300, 704.94047734821500, 705.93269153694400, 707.3883885215910,
            710.53046953848900, 713.29140061860800, 717.49751354398500, 721.62112194478800,
        ]
    }

    fn create_dema_length(length: i64, first_is_average: bool) -> DoubleExponentialMovingAverage {
        DoubleExponentialMovingAverage::new_from_length(&DoubleExponentialMovingAverageLengthParams {
            length,
            first_is_average,
            ..Default::default()
        }).unwrap()
    }

    fn create_dema_alpha(alpha: f64, first_is_average: bool) -> DoubleExponentialMovingAverage {
        DoubleExponentialMovingAverage::new_from_smoothing_factor(&DoubleExponentialMovingAverageSmoothingFactorParams {
            smoothing_factor: alpha,
            first_is_average,
            ..Default::default()
        }).unwrap()
    }

    #[test]
    fn test_update_length_2_first_is_average_true() {
        let mut dema = create_dema_length(2, true);
        let input = test_input();
        let l: usize = 2;
        let lprimed = 2 * l - 2; // 2

        for i in 0..lprimed {
            assert!(dema.update(input[i]).is_nan(), "[{}] expected NaN", i);
        }

        let act = dema.update(input[lprimed]); // index 2
        // Check a few values with tolerance 1e-2
        assert!((94.013 - dema.update(input[3])).abs() < 2.0); // just feed through

        // Feed all remaining
        let mut dema2 = create_dema_length(2, true);
        for i in 0..input.len() {
            let act = dema2.update(input[i]);
            if i == 4 {
                assert!((94.013 - act).abs() < 1e-2, "[4] expected ~94.013, got {}", act);
            }
            if i == 5 {
                assert!((94.539 - act).abs() < 1e-2, "[5] expected ~94.539, got {}", act);
            }
            if i == 251 {
                assert!((107.94 - act).abs() < 1e-2, "[251] expected ~107.94, got {}", act);
            }
        }

        assert!(dema2.update(f64::NAN).is_nan());
    }

    #[test]
    fn test_update_length_14_first_is_average_true() {
        let mut dema = create_dema_length(14, true);
        let input = test_input();
        let lprimed: usize = 2 * 14 - 2; // 26

        for i in 0..lprimed {
            assert!(dema.update(input[i]).is_nan(), "[{}] expected NaN", i);
        }

        for i in lprimed..input.len() {
            let act = dema.update(input[i]);
            match i {
                28 => assert!((84.347 - act).abs() < 1e-2, "[28] got {}", act),
                29 => assert!((84.487 - act).abs() < 1e-2, "[29] got {}", act),
                30 => assert!((84.374 - act).abs() < 1e-2, "[30] got {}", act),
                31 => assert!((84.772 - act).abs() < 1e-2, "[31] got {}", act),
                48 => assert!((89.803 - act).abs() < 1e-2, "[48] got {}", act),
                251 => assert!((109.4676 - act).abs() < 1e-2, "[251] got {}", act),
                _ => {}
            }
        }

        assert!(dema.update(f64::NAN).is_nan());
    }

    #[test]
    fn test_update_length_2_first_is_average_false() {
        let mut dema = create_dema_length(2, false);
        let input = test_input();
        let lprimed: usize = 2 * 2 - 2; // 2

        for i in 0..lprimed {
            assert!(dema.update(input[i]).is_nan(), "[{}] expected NaN", i);
        }

        let mut dema2 = create_dema_length(2, false);
        for i in 0..input.len() {
            let act = dema2.update(input[i]);
            match i {
                4 => assert!((93.977 - act).abs() < 1e-2, "[4] got {}", act),
                5 => assert!((94.522 - act).abs() < 1e-2, "[5] got {}", act),
                251 => assert!((107.94 - act).abs() < 1e-2, "[251] got {}", act),
                _ => {}
            }
        }

        assert!(dema2.update(f64::NAN).is_nan());
    }

    #[test]
    fn test_update_length_14_first_is_average_false() {
        let mut dema = create_dema_length(14, false);
        let input = test_input();
        let lprimed: usize = 2 * 14 - 2;

        for i in 0..lprimed {
            assert!(dema.update(input[i]).is_nan(), "[{}] expected NaN", i);
        }

        for i in lprimed..input.len() {
            let act = dema.update(input[i]);
            match i {
                28 => assert!((84.87 - act).abs() < 1e-2, "[28] got {}", act),
                29 => assert!((84.94 - act).abs() < 1e-2, "[29] got {}", act),
                30 => assert!((84.77 - act).abs() < 1e-2, "[30] got {}", act),
                31 => assert!((85.12 - act).abs() < 1e-2, "[31] got {}", act),
                48 => assert!((89.83 - act).abs() < 1e-2, "[48] got {}", act),
                251 => assert!((109.4676 - act).abs() < 1e-2, "[251] got {}", act),
                _ => {}
            }
        }

        assert!(dema.update(f64::NAN).is_nan());
    }

    #[test]
    fn test_update_length_26_first_is_average_false_tasc() {
        let mut dema = create_dema_length(26, false);
        let input = test_tasc_input();
        let expected = test_tasc_expected();
        let lprimed: usize = 2 * 26 - 2; // 50

        for i in 0..lprimed {
            assert!(dema.update(input[i]).is_nan(), "[{}] expected NaN", i);
        }

        let first_check = 216;
        for i in lprimed..input.len() {
            let act = dema.update(input[i]);
            if i >= first_check {
                assert!((expected[i] - act).abs() < 1e-2, "[{}] expected {}, got {}", i, expected[i], act);
            }
        }

        assert!(dema.update(f64::NAN).is_nan());
    }

    #[test]
    fn test_update_entity() {
        let l: i64 = 2;
        let lprimed = 2 * l - 2;
        let inp = 3.0_f64;
        let exp_false = 2.666666666666667;
        let time = 1617235200_i64;

        // scalar
        let mut dema = create_dema_length(l, false);
        for _ in 0..lprimed { dema.update(0.0); }
        let out = dema.update_scalar(&Scalar::new(time, inp));
        assert_eq!(out.len(), 1);
        let s = out[0].downcast_ref::<Scalar>().unwrap();
        assert_eq!(s.time, time);
        assert!((s.value - exp_false).abs() < 1e-13, "scalar value {} != {}", s.value, exp_false);

        // bar
        let mut dema = create_dema_length(l, false);
        for _ in 0..lprimed { dema.update(0.0); }
        let bar = Bar::new(time, 0.0, 0.0, 0.0, inp, 0.0);
        let out = dema.update_bar(&bar);
        let s = out[0].downcast_ref::<Scalar>().unwrap();
        assert!((s.value - exp_false).abs() < 1e-13);

        // quote
        let mut dema = create_dema_length(l, false);
        for _ in 0..lprimed { dema.update(0.0); }
        let quote = Quote::new(time, inp, inp, 0.0, 0.0);
        let out = dema.update_quote(&quote);
        let s = out[0].downcast_ref::<Scalar>().unwrap();
        assert!((s.value - exp_false).abs() < 1e-13);

        // trade
        let mut dema = create_dema_length(l, false);
        for _ in 0..lprimed { dema.update(0.0); }
        let trade = Trade::new(time, inp, 0.0);
        let out = dema.update_trade(&trade);
        let s = out[0].downcast_ref::<Scalar>().unwrap();
        assert!((s.value - exp_false).abs() < 1e-13);
    }

    #[test]
    fn test_is_primed() {
        let input = test_input();
        let l: usize = 14;
        let lprimed = 2 * l - 2;

        // firstIsAverage = true
        let mut dema = create_dema_length(l as i64, true);
        assert!(!dema.is_primed());
        for i in 0..lprimed {
            dema.update(input[i]);
            assert!(!dema.is_primed(), "[{}] should not be primed", i);
        }
        for i in lprimed..input.len() {
            dema.update(input[i]);
            assert!(dema.is_primed(), "[{}] should be primed", i);
        }

        // firstIsAverage = false
        let mut dema = create_dema_length(l as i64, false);
        assert!(!dema.is_primed());
        for i in 0..lprimed {
            dema.update(input[i]);
            assert!(!dema.is_primed(), "[{}] should not be primed", i);
        }
        for i in lprimed..input.len() {
            dema.update(input[i]);
            assert!(dema.is_primed(), "[{}] should be primed", i);
        }
    }

    #[test]
    fn test_metadata_length() {
        let dema = create_dema_length(10, true);
        let m = dema.metadata();
        assert_eq!(m.identifier, Identifier::DoubleExponentialMovingAverage);
        assert_eq!(m.mnemonic, "dema(10)");
        assert_eq!(m.description, "Double exponential moving average dema(10)");
        assert_eq!(m.outputs.len(), 1);
        assert_eq!(m.outputs[0].kind, DoubleExponentialMovingAverageOutput::Value as i32);
        assert_eq!(m.outputs[0].shape, Shape::Scalar);
        assert_eq!(m.outputs[0].mnemonic, "dema(10)");
        assert_eq!(m.outputs[0].description, "Double exponential moving average dema(10)");
    }

    #[test]
    fn test_metadata_alpha() {
        let alpha = 2.0 / 11.0;
        let dema = create_dema_alpha(alpha, false);
        let m = dema.metadata();
        assert_eq!(m.identifier, Identifier::DoubleExponentialMovingAverage);
        assert_eq!(m.mnemonic, "dema(10, 0.18181818)");
        assert_eq!(m.description, "Double exponential moving average dema(10, 0.18181818)");
    }

    #[test]
    fn test_metadata_length_with_bar_component() {
        let dema = DoubleExponentialMovingAverage::new_from_length(&DoubleExponentialMovingAverageLengthParams {
            length: 10,
            first_is_average: true,
            bar_component: Some(BarComponent::Median),
            ..Default::default()
        }).unwrap();
        let m = dema.metadata();
        assert_eq!(m.mnemonic, "dema(10, hl/2)");
        assert_eq!(m.description, "Double exponential moving average dema(10, hl/2)");
    }

    #[test]
    fn test_metadata_alpha_with_quote_component() {
        let dema = DoubleExponentialMovingAverage::new_from_smoothing_factor(&DoubleExponentialMovingAverageSmoothingFactorParams {
            smoothing_factor: 2.0 / 11.0,
            first_is_average: false,
            quote_component: Some(QuoteComponent::Bid),
            ..Default::default()
        }).unwrap();
        let m = dema.metadata();
        assert_eq!(m.mnemonic, "dema(10, 0.18181818, b)");
        assert_eq!(m.description, "Double exponential moving average dema(10, 0.18181818, b)");
    }

    #[test]
    fn test_new_length_zero() {
        let r = DoubleExponentialMovingAverage::new_from_length(&DoubleExponentialMovingAverageLengthParams {
            length: 0, ..Default::default()
        });
        assert!(r.is_err());
        assert_eq!(r.err().unwrap(), "invalid double exponential moving average parameters: length should be positive");
    }

    #[test]
    fn test_new_length_negative() {
        let r = DoubleExponentialMovingAverage::new_from_length(&DoubleExponentialMovingAverageLengthParams {
            length: -1, ..Default::default()
        });
        assert!(r.is_err());
    }

    #[test]
    fn test_new_alpha_negative() {
        let r = DoubleExponentialMovingAverage::new_from_smoothing_factor(&DoubleExponentialMovingAverageSmoothingFactorParams {
            smoothing_factor: -1.0, ..Default::default()
        });
        assert!(r.is_err());
        assert_eq!(r.err().unwrap(), "invalid double exponential moving average parameters: smoothing factor should be in range [0, 1]");
    }

    #[test]
    fn test_new_alpha_greater_than_1() {
        let r = DoubleExponentialMovingAverage::new_from_smoothing_factor(&DoubleExponentialMovingAverageSmoothingFactorParams {
            smoothing_factor: 2.0, ..Default::default()
        });
        assert!(r.is_err());
    }

    #[test]
    fn test_new_alpha_zero_clamped() {
        let dema = DoubleExponentialMovingAverage::new_from_smoothing_factor(&DoubleExponentialMovingAverageSmoothingFactorParams {
            smoothing_factor: 0.0, ..Default::default()
        }).unwrap();
        assert_eq!(dema.smoothing_factor, 0.00000001);
        assert_eq!(dema.length, 199999999);
    }

    #[test]
    fn test_mnemonic_components() {
        let dema = create_dema_length(10, true);
        assert_eq!(dema.line.mnemonic, "dema(10)");

        let dema = DoubleExponentialMovingAverage::new_from_length(&DoubleExponentialMovingAverageLengthParams {
            length: 10, bar_component: Some(BarComponent::Median), ..Default::default()
        }).unwrap();
        assert_eq!(dema.line.mnemonic, "dema(10, hl/2)");

        let dema = DoubleExponentialMovingAverage::new_from_length(&DoubleExponentialMovingAverageLengthParams {
            length: 10, quote_component: Some(QuoteComponent::Bid), ..Default::default()
        }).unwrap();
        assert_eq!(dema.line.mnemonic, "dema(10, b)");

        let dema = DoubleExponentialMovingAverage::new_from_length(&DoubleExponentialMovingAverageLengthParams {
            length: 10, trade_component: Some(TradeComponent::Volume), ..Default::default()
        }).unwrap();
        assert_eq!(dema.line.mnemonic, "dema(10, v)");
    }
}
