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
    use crate::entities::bar_component::BarComponent;
    use crate::indicators::core::outputs::shape::Shape;

    fn test_high_input() -> Vec<f64> {
        vec![
            93.250000, 94.940000, 96.375000, 96.190000, 96.000000, 94.720000, 95.000000, 93.720000, 92.470000, 92.750000,
            96.250000, 99.625000, 99.125000, 92.750000, 91.315000, 93.250000, 93.405000, 90.655000, 91.970000, 92.250000,
            90.345000, 88.500000, 88.250000, 85.500000, 84.440000, 84.750000, 84.440000, 89.405000, 88.125000, 89.125000,
            87.155000, 87.250000, 87.375000, 88.970000, 90.000000, 89.845000, 86.970000, 85.940000, 84.750000, 85.470000,
            84.470000, 88.500000, 89.470000, 90.000000, 92.440000, 91.440000, 92.970000, 91.720000, 91.155000, 91.750000,
            90.000000, 88.875000, 89.000000, 85.250000, 83.815000, 85.250000, 86.625000, 87.940000, 89.375000, 90.625000,
            90.750000, 88.845000, 91.970000, 93.375000, 93.815000, 94.030000, 94.030000, 91.815000, 92.000000, 91.940000,
            89.750000, 88.750000, 86.155000, 84.875000, 85.940000, 99.375000, 103.280000, 105.375000, 107.625000, 105.250000,
            104.500000, 105.500000, 106.125000, 107.940000, 106.250000, 107.000000, 108.750000, 110.940000, 110.940000,
            114.220000, 123.000000, 121.750000, 119.815000, 120.315000, 119.375000, 118.190000, 116.690000, 115.345000,
            113.000000, 118.315000, 116.870000, 116.750000, 113.870000, 114.620000, 115.310000, 116.000000, 121.690000,
            119.870000, 120.870000, 116.750000, 116.500000, 116.000000, 118.310000, 121.500000, 122.000000, 121.440000,
            125.750000, 127.750000, 124.190000, 124.440000, 125.750000, 124.690000, 125.310000, 132.000000, 131.310000,
            132.250000, 133.880000, 133.500000, 135.500000, 137.440000, 138.690000, 139.190000, 138.500000, 138.130000,
            137.500000, 138.880000, 132.130000, 129.750000, 128.500000, 125.440000, 125.120000, 126.500000, 128.690000,
            126.620000, 126.690000, 126.000000, 123.120000, 121.870000, 124.000000, 127.000000, 124.440000, 122.500000,
            123.750000, 123.810000, 124.500000, 127.870000, 128.560000, 129.630000, 124.870000, 124.370000, 124.870000,
            123.620000, 124.060000, 125.870000, 125.190000, 125.620000, 126.000000, 128.500000, 126.750000, 129.750000,
            132.690000, 133.940000, 136.500000, 137.690000, 135.560000, 133.560000, 135.000000, 132.380000, 131.440000,
            130.880000, 129.630000, 127.250000, 127.810000, 125.000000, 126.810000, 124.750000, 122.810000, 122.250000,
            121.060000, 120.000000, 123.250000, 122.750000, 119.190000, 115.060000, 116.690000, 114.870000, 110.870000,
            107.250000, 108.870000, 109.000000, 108.500000, 113.060000, 93.000000, 94.620000, 95.120000, 96.000000,
            95.560000, 95.310000, 99.000000, 98.810000, 96.810000, 95.940000, 94.440000, 92.940000, 93.940000, 95.500000,
            97.060000, 97.500000, 96.250000, 96.370000, 95.000000, 94.870000, 98.250000, 105.120000, 108.440000, 109.870000,
            105.000000, 106.000000, 104.940000, 104.500000, 104.440000, 106.310000, 112.870000, 116.500000, 119.190000,
            121.000000, 122.120000, 111.940000, 112.750000, 110.190000, 107.940000, 109.690000, 111.060000, 110.440000,
            110.120000, 110.310000, 110.440000, 110.000000, 110.750000, 110.500000, 110.500000, 109.500000,
        ]
    }

    fn test_low_input() -> Vec<f64> {
        vec![
            90.750000, 91.405000, 94.250000, 93.500000, 92.815000, 93.500000, 92.000000, 89.750000, 89.440000, 90.625000,
            92.750000, 96.315000, 96.030000, 88.815000, 86.750000, 90.940000, 88.905000, 88.780000, 89.250000, 89.750000,
            87.500000, 86.530000, 84.625000, 82.280000, 81.565000, 80.875000, 81.250000, 84.065000, 85.595000, 85.970000,
            84.405000, 85.095000, 85.500000, 85.530000, 87.875000, 86.565000, 84.655000, 83.250000, 82.565000, 83.440000,
            82.530000, 85.065000, 86.875000, 88.530000, 89.280000, 90.125000, 90.750000, 89.000000, 88.565000, 90.095000,
            89.000000, 86.470000, 84.000000, 83.315000, 82.000000, 83.250000, 84.750000, 85.280000, 87.190000, 88.440000,
            88.250000, 87.345000, 89.280000, 91.095000, 89.530000, 91.155000, 92.000000, 90.530000, 89.970000, 88.815000,
            86.750000, 85.065000, 82.030000, 81.500000, 82.565000, 96.345000, 96.470000, 101.155000, 104.250000, 101.750000,
            101.720000, 101.720000, 103.155000, 105.690000, 103.655000, 104.000000, 105.530000, 108.530000, 108.750000,
            107.750000, 117.000000, 118.000000, 116.000000, 118.500000, 116.530000, 116.250000, 114.595000, 110.875000,
            110.500000, 110.720000, 112.620000, 114.190000, 111.190000, 109.440000, 111.560000, 112.440000, 117.500000,
            116.060000, 116.560000, 113.310000, 112.560000, 114.000000, 114.750000, 118.870000, 119.000000, 119.750000,
            122.620000, 123.000000, 121.750000, 121.560000, 123.120000, 122.190000, 122.750000, 124.370000, 128.000000,
            129.500000, 130.810000, 130.630000, 132.130000, 133.880000, 135.380000, 135.750000, 136.190000, 134.500000,
            135.380000, 133.690000, 126.060000, 126.870000, 123.500000, 122.620000, 122.750000, 123.560000, 125.810000,
            124.620000, 124.370000, 121.810000, 118.190000, 118.060000, 117.560000, 121.000000, 121.120000, 118.940000,
            119.810000, 121.000000, 122.000000, 124.500000, 126.560000, 123.500000, 121.250000, 121.060000, 122.310000,
            121.000000, 120.870000, 122.060000, 122.750000, 122.690000, 122.870000, 125.500000, 124.250000, 128.000000,
            128.380000, 130.690000, 131.630000, 134.380000, 132.000000, 131.940000, 131.940000, 129.560000, 123.750000,
            126.000000, 126.250000, 124.370000, 121.440000, 120.440000, 121.370000, 121.690000, 120.000000, 119.620000,
            115.500000, 116.750000, 119.060000, 119.060000, 115.060000, 111.060000, 113.120000, 110.000000, 105.000000,
            104.690000, 103.870000, 104.690000, 105.440000, 107.000000, 89.000000, 92.500000, 92.120000, 94.620000,
            92.810000, 94.250000, 96.250000, 96.370000, 93.690000, 93.500000, 90.000000, 90.190000, 90.500000, 92.120000,
            94.120000, 94.870000, 93.000000, 93.870000, 93.000000, 92.620000, 93.560000, 98.370000, 104.440000, 106.000000,
            101.810000, 104.120000, 103.370000, 102.120000, 102.250000, 103.370000, 107.940000, 112.500000, 115.440000,
            115.500000, 112.250000, 107.560000, 106.560000, 106.870000, 104.500000, 105.750000, 108.620000, 107.750000,
            108.060000, 108.000000, 108.190000, 108.120000, 109.060000, 108.750000, 108.560000, 106.620000,
        ]
    }

    fn test_excel_expected() -> Vec<f64> {
        vec![
            f64::NAN, f64::NAN, f64::NAN, f64::NAN,
            f64::NAN, f64::NAN, f64::NAN, f64::NAN,
            f64::NAN, f64::NAN, f64::NAN, f64::NAN,
            f64::NAN, f64::NAN, f64::NAN, f64::NAN,
            f64::NAN, f64::NAN, f64::NAN,
            0.9401568590471170, 0.9471811552404370, 0.9526486511378510, 0.9594395234433260,
            0.9684901755431890, 0.9745456460684540, 0.9823609526746270, 0.9842318414143900,
            0.9793319838629490, 0.9788482705359360, 0.9800972276693750, 0.9785845616077360,
            0.9693167394743250, 0.9470228533431190, 0.9453838549547040, 0.9517592162172850,
            0.9433392348270420, 0.9478926222765860, 0.9475923828106470, 0.9381448555694890,
            0.9153603780858040, 0.9037133253874650, 0.9044272330533300, 0.9144567231385610,
            0.9140422416270490, 0.9240907021037240, 0.9245714706598480, 0.9301864925512240,
            0.9685637131517730, 0.9699713144403260, 0.9722430818693880, 0.9659537966306980,
            0.9653083079242570, 0.9421810188742970, 0.9488679626820190, 0.9540546636443100,
            0.9585657829863180, 0.9580677494166530, 0.9563189205945780, 0.9522216783457150,
            0.9499605370495780, 0.9433582538169970, 0.9448506810540480, 0.9467581283005850,
            0.9540541884980060, 0.9451938775561860, 0.9527747257400080, 0.9545845363155470,
            0.9518140900144220, 0.9526969023077840, 0.9521438024060810, 0.9571587276746930,
            0.9534806508367840, 0.9664678685848590, 0.9676372885038130, 0.9673951425348660,
            0.9754944156042740, 0.9618023912755630, 0.9735983892487120, 0.9814507316373980,
            0.9846276725648570, 0.9860054566213870, 0.9882105937528850, 0.9891883760796960,
            0.9901171811095450, 0.9908918742809730, 0.9913801756613770, 0.9923983921194380,
            0.9940283028145840, 0.9946689515269380, 0.9917313455559860, 0.9912246592297800,
            0.9914766005027000, 0.9905843446629170, 0.9867768283384500, 0.9800208355632590,
            0.9775868446129000, 0.9818158252339090, 0.9805269039354710, 0.9794926025278210,
            0.9651664579178690, 0.9607101819596180, 0.9548369427331770, 0.9472371147678250,
            0.9393874519173430, 0.9262650699399830, 0.9055359430253950, 0.8917513099360840,
            0.8845363842595150, 0.8780023472740510, 0.8768594795208420, 0.8847330319751160,
            0.8588444464269830, 0.8540394922856320, 0.8685071144796010, 0.8867017785665350,
            0.9019536780962040, 0.9351506214260950, 0.9460614408106260, 0.9487795387829030,
            0.9755185789962120, 0.9781420983101290, 0.9796746522636920, 0.9800485956962120,
            0.9592735362917000, 0.9629702197175760, 0.9653515565441640, 0.9722435637049560,
            0.9751999286917070, 0.9793400350341120, 0.9807373681199240, 0.9820819629564370,
            0.9819960901599760, 0.9804984607791450, 0.9802670958185390, 0.9782011441033960,
            0.9752990704028820, 0.9675865981117540, 0.9675082971922110, 0.9636801404836100,
            0.9624399587196960, 0.9623844816290560, 0.9615007729973930, 0.9591567353952830,
            0.9776738747439900, 0.9783423006715560, 0.9795238639226300, 0.9805925440791580,
            0.9832188414769910, 0.9796598649299700, 0.9758993430891440, 0.9735179726086800,
            0.9704997683460010, 0.9656284571769230, 0.9553451160914730, 0.9361819221968540,
            0.8782125747224650, 0.8982719217727120, 0.8477723464095960, 0.8546377589244040,
            0.8572871327351020, 0.8585528871548040, 0.8526022957921970, 0.8292629229925110,
            0.8294469606129980, 0.8190212269209530, 0.8191577970270310, 0.8132742438973340,
            0.8118086128565220, 0.8451960695090540, 0.9147140984114960, 0.9323286922042040,
            0.9488173569229080, 0.9616254945974570, 0.9714396824012870, 0.9749722214786510,
            0.9725790982765770, 0.9766180777342920, 0.9873289956355020, 0.9621423713797460,
            0.9561111788043280, 0.9538925936199330, 0.9496547804857550, 0.9329386834398710,
            0.9349122391333710, 0.9350195404239680, 0.9370998633521060, 0.9414383095728230,
            0.9472807230887830, 0.9572389026417090, 0.9658768229512560, 0.9674312191145570,
            0.9673978992813970, 0.9691937580390360, 0.9698138706945130, 0.9685257570998960,
            0.9728666378692480, 0.9740699834558090, 0.9773946409583190, 0.9859156575982320,
            0.9868010214710570, 0.9860650762821040, 0.9840721831018110, 0.9917556231372550,
            0.9922091139248850, 0.9933724117160440, 0.9926935073704690, 0.9934630349771790,
            0.9934195189537240, 0.9935298892857930, 0.9935049843553920, 0.9928126293583000,
            0.9921799215414420, 0.9899983601098890, 0.9895275103398580, 0.9882614534065030,
            0.9855713950443270, 0.9854298482919770, 0.9852191036619460, 0.9834218730447970,
            0.9788631710872740, 0.9704655850058130, 0.9194300540508450, 0.8681407925170670,
            0.8808536680642600, 0.9440776152383960, 0.9689567658743020, 0.9716435818001360,
            0.9756305746107580, 0.9745748112805210, 0.9757768027062840, 0.9757402824215750,
            0.9771541892863710, 0.9795131678773640, 0.9827854278332160, 0.9859330369551790,
            0.9874122533690860, 0.9784753788149660, 0.9776244828040590, 0.9745149943245450,
            0.9710356842260290, 0.9651684328545020, 0.9548286702701730, 0.9440773679358460,
            0.9574334358838210, 0.9540249522022400, 0.9517837091130880, 0.9466458287565880,
            0.9408225232591590, 0.9304922072407360, 0.9156400478034220, 0.8963662049425160,
            0.8866901149929160,
        ]
    }

    fn create(length: usize) -> PearsonsCorrelationCoefficient {
        PearsonsCorrelationCoefficient::new(&PearsonsCorrelationCoefficientParams {
            length,
            ..Default::default()
        }).unwrap()
    }

    #[test]
    fn test_update_pair_talib_spot_checks() {
        let mut c = create(20);
        let high = test_high_input();
        let low = test_low_input();

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
        let high = test_high_input();
        let low = test_low_input();
        let expected = test_excel_expected();

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
        let high = test_high_input();
        let low = test_low_input();

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
