use super::bar::Bar;

/// Describes a component of the Bar type.
#[derive(Debug, Clone, Copy, PartialEq, Eq)]
#[repr(u8)]
pub enum BarComponent {
    Open = 0,
    High = 1,
    Low = 2,
    Close = 3,
    Volume = 4,
    Median = 5,
    Typical = 6,
    Weighted = 7,
    Average = 8,
}

/// Returns a function that extracts the given component value from a Bar.
pub fn component_value(component: BarComponent) -> fn(&Bar) -> f64 {
    match component {
        BarComponent::Open => |b: &Bar| b.open,
        BarComponent::High => |b: &Bar| b.high,
        BarComponent::Low => |b: &Bar| b.low,
        BarComponent::Close => |b: &Bar| b.close,
        BarComponent::Volume => |b: &Bar| b.volume,
        BarComponent::Median => |b: &Bar| b.median(),
        BarComponent::Typical => |b: &Bar| b.typical(),
        BarComponent::Weighted => |b: &Bar| b.weighted(),
        BarComponent::Average => |b: &Bar| b.average(),
    }
}

/// Returns the mnemonic string for the given bar component.
pub fn component_mnemonic(component: BarComponent) -> &'static str {
    match component {
        BarComponent::Open => "o",
        BarComponent::High => "h",
        BarComponent::Low => "l",
        BarComponent::Close => "c",
        BarComponent::Volume => "v",
        BarComponent::Median => "hl/2",
        BarComponent::Typical => "hlc/3",
        BarComponent::Weighted => "hlcc/4",
        BarComponent::Average => "ohlc/4",
    }
}

#[cfg(test)]
mod tests {
    use super::*;

    fn make_bar(o: f64, h: f64, l: f64, c: f64, v: f64) -> Bar {
        Bar::new(0, o, h, l, c, v)
    }

    #[test]
    fn test_component_value_open() {
        let b = make_bar(2.0, 4.0, 1.0, 3.0, 5.0);
        assert_eq!(component_value(BarComponent::Open)(&b), 2.0);
    }

    #[test]
    fn test_component_value_high() {
        let b = make_bar(2.0, 4.0, 1.0, 3.0, 5.0);
        assert_eq!(component_value(BarComponent::High)(&b), 4.0);
    }

    #[test]
    fn test_component_value_low() {
        let b = make_bar(2.0, 4.0, 1.0, 3.0, 5.0);
        assert_eq!(component_value(BarComponent::Low)(&b), 1.0);
    }

    #[test]
    fn test_component_value_close() {
        let b = make_bar(2.0, 4.0, 1.0, 3.0, 5.0);
        assert_eq!(component_value(BarComponent::Close)(&b), 3.0);
    }

    #[test]
    fn test_component_value_volume() {
        let b = make_bar(2.0, 4.0, 1.0, 3.0, 5.0);
        assert_eq!(component_value(BarComponent::Volume)(&b), 5.0);
    }

    #[test]
    fn test_component_value_median() {
        let b = make_bar(2.0, 4.0, 1.0, 3.0, 5.0);
        assert_eq!(component_value(BarComponent::Median)(&b), (1.0 + 4.0) / 2.0);
    }

    #[test]
    fn test_component_value_typical() {
        let b = make_bar(2.0, 4.0, 1.0, 3.0, 5.0);
        assert_eq!(component_value(BarComponent::Typical)(&b), (1.0 + 4.0 + 3.0) / 3.0);
    }

    #[test]
    fn test_component_value_weighted() {
        let b = make_bar(2.0, 4.0, 1.0, 3.0, 5.0);
        assert_eq!(component_value(BarComponent::Weighted)(&b), (1.0 + 4.0 + 3.0 + 3.0) / 4.0);
    }

    #[test]
    fn test_component_value_average() {
        let b = make_bar(2.0, 4.0, 1.0, 3.0, 5.0);
        assert_eq!(component_value(BarComponent::Average)(&b), (1.0 + 4.0 + 3.0 + 2.0) / 4.0);
    }

    #[test]
    fn test_component_mnemonic() {
        assert_eq!(component_mnemonic(BarComponent::Open), "o");
        assert_eq!(component_mnemonic(BarComponent::High), "h");
        assert_eq!(component_mnemonic(BarComponent::Low), "l");
        assert_eq!(component_mnemonic(BarComponent::Close), "c");
        assert_eq!(component_mnemonic(BarComponent::Volume), "v");
        assert_eq!(component_mnemonic(BarComponent::Median), "hl/2");
        assert_eq!(component_mnemonic(BarComponent::Typical), "hlc/3");
        assert_eq!(component_mnemonic(BarComponent::Weighted), "hlcc/4");
        assert_eq!(component_mnemonic(BarComponent::Average), "ohlc/4");
    }
}
