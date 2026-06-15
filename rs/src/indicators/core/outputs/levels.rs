/// A single entry of a Levels output, expressed as a value with an optional bar
/// offset and an optional strength.
#[derive(Debug, Clone, Copy)]
pub struct Level {
    /// The value (e.g. a price level or a multiplier) at this entry.
    pub value: f64,
    /// The number of bars back from the Levels' time (0 = current bar).
    pub offset: i32,
    /// Optional significance measure (higher = more significant); NaN when not applicable.
    pub strength: f64,
}

impl Level {
    /// Creates a level with the given value, offset and strength.
    pub fn new(value: f64, offset: i32, strength: f64) -> Self {
        Self { value, offset, strength }
    }

    /// Creates a level with the given value, an offset of 0 and a NaN strength.
    pub fn value_only(value: f64) -> Self {
        Self { value, offset: 0, strength: f64::NAN }
    }
}

/// Holds a time stamp and a variable-length set of levels.
#[derive(Debug, Clone)]
pub struct Levels {
    pub time: i64,
    pub levels: Vec<Level>,
}

impl Levels {
    /// Creates a new Levels with the given time and entries.
    pub fn new(time: i64, levels: Vec<Level>) -> Self {
        Self { time, levels }
    }

    /// Creates a new empty Levels with no entries.
    pub fn empty(time: i64) -> Self {
        Self { time, levels: Vec::new() }
    }

    /// Indicates whether this Levels has no entries.
    pub fn is_empty(&self) -> bool {
        self.levels.is_empty()
    }
}

#[cfg(test)]
mod tests {
    use super::*;

    #[test]
    fn test_new_and_fields() {
        let l = Levels::new(42, vec![Level::new(105.5, 3, 0.8), Level::new(102.0, 1, 0.6)]);
        assert_eq!(l.time, 42);
        assert_eq!(l.levels.len(), 2);
        assert_eq!(l.levels[0].value, 105.5);
        assert_eq!(l.levels[0].offset, 3);
        assert_eq!(l.levels[0].strength, 0.8);
        assert!(!l.is_empty());
    }

    #[test]
    fn test_value_only_has_nan_strength() {
        let lv = Level::value_only(42.0);
        assert_eq!(lv.value, 42.0);
        assert_eq!(lv.offset, 0);
        assert!(lv.strength.is_nan());
    }

    #[test]
    fn test_empty() {
        let l = Levels::empty(7);
        assert_eq!(l.time, 7);
        assert!(l.is_empty());
    }
}
