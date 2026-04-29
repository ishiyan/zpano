// Root module for the entities library.
// Re-exports all entity types and component sub-modules.

// --- Entity types ---
pub const bar = @import("bar");
pub const quote = @import("quote");
pub const trade = @import("trade");
pub const scalar = @import("scalar");

// --- Component types ---
pub const bar_component = @import("bar_component");
pub const quote_component = @import("quote_component");
pub const trade_component = @import("trade_component");

// --- Convenience type aliases ---
pub const Bar = bar.Bar;
pub const Quote = quote.Quote;
pub const Trade = trade.Trade;
pub const Scalar = scalar.Scalar;
pub const BarComponent = bar_component.BarComponent;
pub const QuoteComponent = quote_component.QuoteComponent;
pub const TradeComponent = trade_component.TradeComponent;
