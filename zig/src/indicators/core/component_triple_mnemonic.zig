const bar_component = @import("bar_component");
const quote_component = @import("quote_component");
const trade_component = @import("trade_component");

/// Builds a mnemonic suffix from bar, quote and trade components.
///
/// A component equal to its default is omitted. For example, if the bar
/// component is Median (non-default), the result is ", hl/2". If the bar
/// component is Close (default), it is omitted.
pub fn componentTripleMnemonic(
    buf: []u8,
    bc: bar_component.BarComponent,
    qc: quote_component.QuoteComponent,
    tc: trade_component.TradeComponent,
) []const u8 {
    var pos: usize = 0;

    if (bc != bar_component.default_bar_component) {
        const m = bar_component.componentMnemonic(bc);
        @memcpy(buf[pos .. pos + 2], ", ");
        pos += 2;
        @memcpy(buf[pos .. pos + m.len], m);
        pos += m.len;
    }

    if (qc != quote_component.default_quote_component) {
        const m = quote_component.componentMnemonic(qc);
        @memcpy(buf[pos .. pos + 2], ", ");
        pos += 2;
        @memcpy(buf[pos .. pos + m.len], m);
        pos += m.len;
    }

    if (tc != trade_component.default_trade_component) {
        const m = trade_component.componentMnemonic(tc);
        @memcpy(buf[pos .. pos + 2], ", ");
        pos += 2;
        @memcpy(buf[pos .. pos + m.len], m);
        pos += m.len;
    }

    return buf[0..pos];
}
