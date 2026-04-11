use super::quote::Quote;

/// Describes a component of the Quote type.
#[derive(Debug, Clone, Copy, PartialEq, Eq)]
#[repr(u8)]
pub enum QuoteComponent {
    Bid = 0,
    Ask = 1,
    BidSize = 2,
    AskSize = 3,
    Mid = 4,
    Weighted = 5,
    WeightedMid = 6,
    SpreadBp = 7,
}

/// Returns a function that extracts the given component value from a Quote.
pub fn component_value(component: QuoteComponent) -> fn(&Quote) -> f64 {
    match component {
        QuoteComponent::Bid => |q: &Quote| q.bid_price,
        QuoteComponent::Ask => |q: &Quote| q.ask_price,
        QuoteComponent::BidSize => |q: &Quote| q.bid_size,
        QuoteComponent::AskSize => |q: &Quote| q.ask_size,
        QuoteComponent::Mid => |q: &Quote| q.mid(),
        QuoteComponent::Weighted => |q: &Quote| q.weighted(),
        QuoteComponent::WeightedMid => |q: &Quote| q.weighted_mid(),
        QuoteComponent::SpreadBp => |q: &Quote| q.spread_bp(),
    }
}

/// Returns the mnemonic string for the given quote component.
pub fn component_mnemonic(component: QuoteComponent) -> &'static str {
    match component {
        QuoteComponent::Bid => "b",
        QuoteComponent::Ask => "a",
        QuoteComponent::BidSize => "bs",
        QuoteComponent::AskSize => "as",
        QuoteComponent::Mid => "ba/2",
        QuoteComponent::Weighted => "(bbs+aas)/(bs+as)",
        QuoteComponent::WeightedMid => "(bas+abs)/(bs+as)",
        QuoteComponent::SpreadBp => "spread bp",
    }
}

#[cfg(test)]
mod tests {
    use super::*;

    fn make_quote(bid: f64, ask: f64, bs: f64, ask_s: f64) -> Quote {
        Quote::new(0, bid, ask, bs, ask_s)
    }

    #[test]
    fn test_component_value_bid() {
        let q = make_quote(2.0, 1.0, 4.0, 3.0);
        assert_eq!(component_value(QuoteComponent::Bid)(&q), 2.0);
    }

    #[test]
    fn test_component_value_ask() {
        let q = make_quote(2.0, 1.0, 4.0, 3.0);
        assert_eq!(component_value(QuoteComponent::Ask)(&q), 1.0);
    }

    #[test]
    fn test_component_value_bid_size() {
        let q = make_quote(2.0, 1.0, 4.0, 3.0);
        assert_eq!(component_value(QuoteComponent::BidSize)(&q), 4.0);
    }

    #[test]
    fn test_component_value_ask_size() {
        let q = make_quote(2.0, 1.0, 4.0, 3.0);
        assert_eq!(component_value(QuoteComponent::AskSize)(&q), 3.0);
    }

    #[test]
    fn test_component_value_mid() {
        let q = make_quote(2.0, 1.0, 4.0, 3.0);
        assert_eq!(component_value(QuoteComponent::Mid)(&q), (1.0 + 2.0) / 2.0);
    }

    #[test]
    fn test_component_value_weighted() {
        let q = make_quote(2.0, 1.0, 4.0, 3.0);
        assert_eq!(component_value(QuoteComponent::Weighted)(&q), (1.0 * 3.0 + 2.0 * 4.0) / (3.0 + 4.0));
    }

    #[test]
    fn test_component_value_weighted_mid() {
        let q = make_quote(2.0, 1.0, 4.0, 3.0);
        assert_eq!(component_value(QuoteComponent::WeightedMid)(&q), (1.0 * 4.0 + 2.0 * 3.0) / (3.0 + 4.0));
    }

    #[test]
    fn test_component_value_spread_bp() {
        let q = make_quote(2.0, 1.0, 4.0, 3.0);
        assert_eq!(component_value(QuoteComponent::SpreadBp)(&q), 10000.0 * 2.0 * (1.0 - 2.0) / (1.0 + 2.0));
    }

    #[test]
    fn test_component_mnemonic() {
        assert_eq!(component_mnemonic(QuoteComponent::Bid), "b");
        assert_eq!(component_mnemonic(QuoteComponent::Ask), "a");
        assert_eq!(component_mnemonic(QuoteComponent::BidSize), "bs");
        assert_eq!(component_mnemonic(QuoteComponent::AskSize), "as");
        assert_eq!(component_mnemonic(QuoteComponent::Mid), "ba/2");
        assert_eq!(component_mnemonic(QuoteComponent::Weighted), "(bbs+aas)/(bs+as)");
        assert_eq!(component_mnemonic(QuoteComponent::WeightedMid), "(bas+abs)/(bs+as)");
        assert_eq!(component_mnemonic(QuoteComponent::SpreadBp), "spread bp");
    }
}
