const Bar = @import("bar").Bar;
const Quote = @import("quote").Quote;
const Trade = @import("trade").Trade;
const Scalar = @import("scalar").Scalar;
const bar_component = @import("bar_component");
const quote_component = @import("quote_component");
const trade_component = @import("trade_component");
const indicator_mod = @import("indicator.zig");

const OutputArray = indicator_mod.OutputArray;

/// Type aliases for component extraction functions.
pub const BarFunc = *const fn (Bar) f64;
pub const QuoteFunc = *const fn (Quote) f64;
pub const TradeFunc = *const fn (Trade) f64;

/// Provides component extraction and output wrapping for indicators that take
/// a single numeric input and produce a single scalar output.
///
/// Use via composition: store a `LineIndicator` field in a concrete indicator
/// struct. Call `extractBar`/`extractQuote`/`extractTrade` to get the sample
/// value, run your own update logic, then call `wrapScalar` to produce output.
pub const LineIndicator = struct {
    /// Short name of the indicator.
    mnemonic: []const u8,
    /// Description of the indicator.
    description: []const u8,
    bar_func: BarFunc,
    quote_func: QuoteFunc,
    trade_func: TradeFunc,

    /// Creates a new LineIndicator with component functions resolved from
    /// optional component enums. `null` means use the default component.
    pub fn new(
        mnemonic: []const u8,
        description: []const u8,
        bc: ?bar_component.BarComponent,
        qc: ?quote_component.QuoteComponent,
        tc: ?trade_component.TradeComponent,
    ) LineIndicator {
        return .{
            .mnemonic = mnemonic,
            .description = description,
            .bar_func = bar_component.componentValue(bc orelse bar_component.default_bar_component),
            .quote_func = quote_component.componentValue(qc orelse quote_component.default_quote_component),
            .trade_func = trade_component.componentValue(tc orelse trade_component.default_trade_component),
        };
    }

    /// Extracts the numeric value from a Bar using the stored component function.
    pub fn extractBar(self: *const LineIndicator, bar: *const Bar) f64 {
        return (self.bar_func)(bar.*);
    }

    /// Extracts the numeric value from a Quote using the stored component function.
    pub fn extractQuote(self: *const LineIndicator, quote: *const Quote) f64 {
        return (self.quote_func)(quote.*);
    }

    /// Extracts the numeric value from a Trade using the stored component function.
    pub fn extractTrade(self: *const LineIndicator, trade: *const Trade) f64 {
        return (self.trade_func)(trade.*);
    }

    /// Wraps a computed value into an OutputArray with a single scalar output.
    pub fn wrapScalar(time: i64, value: f64) OutputArray {
        return OutputArray.fromScalar(.{ .time = time, .value = value });
    }
};
