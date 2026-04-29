use crate::entities::bar_component::{component_mnemonic as bar_mnemonic, BarComponent, DEFAULT_BAR_COMPONENT};
use crate::entities::quote_component::{component_mnemonic as quote_mnemonic, QuoteComponent, DEFAULT_QUOTE_COMPONENT};
use crate::entities::trade_component::{component_mnemonic as trade_mnemonic, TradeComponent, DEFAULT_TRADE_COMPONENT};

/// Builds a mnemonic suffix from bar, quote and trade components.
///
/// A component equal to its default is omitted. For example, if the bar
/// component is Median (non-default), the result is `", hl/2"`. If the bar
/// component is Close (default), it is omitted.
pub fn component_triple_mnemonic(
    bc: BarComponent,
    qc: QuoteComponent,
    tc: TradeComponent,
) -> String {
    let mut s = String::new();

    if bc != DEFAULT_BAR_COMPONENT {
        s.push_str(", ");
        s.push_str(bar_mnemonic(bc));
    }

    if qc != DEFAULT_QUOTE_COMPONENT {
        s.push_str(", ");
        s.push_str(quote_mnemonic(qc));
    }

    if tc != DEFAULT_TRADE_COMPONENT {
        s.push_str(", ");
        s.push_str(trade_mnemonic(tc));
    }

    s
}
