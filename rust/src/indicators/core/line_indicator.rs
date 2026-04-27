use crate::entities::bar::Bar;
use crate::entities::quote::Quote;
use crate::entities::scalar::Scalar;
use crate::entities::trade::Trade;
use super::indicator::Output;

/// Type aliases for component extraction functions.
pub type BarFunc = fn(&Bar) -> f64;
pub type QuoteFunc = fn(&Quote) -> f64;
pub type TradeFunc = fn(&Trade) -> f64;

/// Provides component extraction and output wrapping for indicators that take
/// a single numeric input and produce a single scalar output.
///
/// Use via composition: store a `LineIndicator` field in a concrete indicator
/// struct, then implement the `Indicator` trait by calling the `update_*`
/// methods with a closure that invokes the indicator's own `update` logic.
pub struct LineIndicator {
    /// Short name of the indicator.
    pub mnemonic: String,
    /// Description of the indicator.
    pub description: String,
    pub bar_func: BarFunc,
    pub quote_func: QuoteFunc,
    pub trade_func: TradeFunc,
}

impl LineIndicator {
    /// Creates a new LineIndicator.
    pub fn new(
        mnemonic: String,
        description: String,
        bar_func: BarFunc,
        quote_func: QuoteFunc,
        trade_func: TradeFunc,
    ) -> Self {
        Self { mnemonic, description, bar_func, quote_func, trade_func }
    }

    /// Updates the indicator given the next scalar sample, using the provided
    /// update function to compute the output value.
    pub fn update_scalar(&self, sample: &Scalar, update_fn: impl FnOnce(f64) -> f64) -> Output {
        let value = update_fn(sample.value);
        vec![Box::new(Scalar::new(sample.time, value))]
    }

    /// Updates the indicator given the next bar sample.
    pub fn update_bar(&self, sample: &Bar, update_fn: impl FnOnce(f64) -> f64) -> Output {
        let scalar = Scalar::new(sample.time, (self.bar_func)(sample));
        self.update_scalar(&scalar, update_fn)
    }

    /// Updates the indicator given the next quote sample.
    pub fn update_quote(&self, sample: &Quote, update_fn: impl FnOnce(f64) -> f64) -> Output {
        let scalar = Scalar::new(sample.time, (self.quote_func)(sample));
        self.update_scalar(&scalar, update_fn)
    }

    /// Updates the indicator given the next trade sample.
    pub fn update_trade(&self, sample: &Trade, update_fn: impl FnOnce(f64) -> f64) -> Output {
        let scalar = Scalar::new(sample.time, (self.trade_func)(sample));
        self.update_scalar(&scalar, update_fn)
    }
}
