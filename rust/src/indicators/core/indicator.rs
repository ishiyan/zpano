use crate::entities::bar::Bar;
use crate::entities::quote::Quote;
use crate::entities::scalar::Scalar;
use crate::entities::trade::Trade;

/// Output is a vector of boxed output values (Scalar, Band, Heatmap, Polyline).
pub type Output = Vec<Box<dyn std::any::Any>>;

/// Common indicator functionality.
pub trait Indicator {
    /// Indicates whether the indicator is primed.
    fn is_primed(&self) -> bool;

    /// Returns metadata describing this indicator and its outputs.
    fn metadata(&self) -> super::metadata::Metadata;

    /// Updates the indicator given the next scalar sample.
    fn update_scalar(&mut self, sample: &Scalar) -> Output;

    /// Updates the indicator given the next bar sample.
    fn update_bar(&mut self, sample: &Bar) -> Output;

    /// Updates the indicator given the next quote sample.
    fn update_quote(&mut self, sample: &Quote) -> Output;

    /// Updates the indicator given the next trade sample.
    fn update_trade(&mut self, sample: &Trade) -> Output;
}

/// Updates the indicator given a slice of scalar samples.
pub fn update_scalars(ind: &mut dyn Indicator, samples: &[Scalar]) -> Vec<Output> {
    samples.iter().map(|s| ind.update_scalar(s)).collect()
}

/// Updates the indicator given a slice of bar samples.
pub fn update_bars(ind: &mut dyn Indicator, samples: &[Bar]) -> Vec<Output> {
    samples.iter().map(|s| ind.update_bar(s)).collect()
}

/// Updates the indicator given a slice of quote samples.
pub fn update_quotes(ind: &mut dyn Indicator, samples: &[Quote]) -> Vec<Output> {
    samples.iter().map(|s| ind.update_quote(s)).collect()
}

/// Updates the indicator given a slice of trade samples.
pub fn update_trades(ind: &mut dyn Indicator, samples: &[Trade]) -> Vec<Output> {
    samples.iter().map(|s| ind.update_trade(s)).collect()
}
