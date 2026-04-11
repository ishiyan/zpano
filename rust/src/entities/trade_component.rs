use super::trade::Trade;

/// Describes a component of the Trade type.
#[derive(Debug, Clone, Copy, PartialEq, Eq)]
#[repr(u8)]
pub enum TradeComponent {
    Price = 0,
    Volume = 1,
}

/// Returns a function that extracts the given component value from a Trade.
pub fn component_value(component: TradeComponent) -> fn(&Trade) -> f64 {
    match component {
        TradeComponent::Price => |t: &Trade| t.price,
        TradeComponent::Volume => |t: &Trade| t.volume,
    }
}

/// Returns the mnemonic string for the given trade component.
pub fn component_mnemonic(component: TradeComponent) -> &'static str {
    match component {
        TradeComponent::Price => "p",
        TradeComponent::Volume => "v",
    }
}

#[cfg(test)]
mod tests {
    use super::*;

    #[test]
    fn test_component_value_price() {
        let t = Trade::new(0, 1.0, 2.0);
        assert_eq!(component_value(TradeComponent::Price)(&t), 1.0);
    }

    #[test]
    fn test_component_value_volume() {
        let t = Trade::new(0, 1.0, 2.0);
        assert_eq!(component_value(TradeComponent::Volume)(&t), 2.0);
    }

    #[test]
    fn test_component_mnemonic() {
        assert_eq!(component_mnemonic(TradeComponent::Price), "p");
        assert_eq!(component_mnemonic(TradeComponent::Volume), "v");
    }
}
