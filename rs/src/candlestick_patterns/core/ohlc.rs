/// A single bar's open, high, low, close values.
#[derive(Debug, Clone, Copy)]
pub struct OHLC {
    pub o: f64,
    pub h: f64,
    pub l: f64,
    pub c: f64,
}
