use std::fmt;
use std::str::FromStr;

/// DayCountConvention represents different day count conventions
/// used in financial calculations.
#[derive(Debug, Clone, Copy, PartialEq, Eq, Hash)]
#[repr(u8)]
pub enum DayCountConvention {
    Raw = 0,
    Thirty360Us = 1,
    Thirty360UsEom = 2,
    Thirty360UsNasd = 3,
    Thirty360Eu = 4,
    Thirty360EuM2 = 5,
    Thirty360EuM3 = 6,
    Thirty360EuPlus = 7,
    Thirty365 = 8,
    Act360 = 9,
    Act365Fixed = 10,
    Act365Nonleap = 11,
    ActActExcel = 12,
    ActActIsda = 13,
    ActActAfb = 14,
}

impl DayCountConvention {
    /// Returns the string key used for convention lookup.
    pub fn as_str(&self) -> &'static str {
        match self {
            Self::Raw => "raw",
            Self::Thirty360Us => "30/360 us",
            Self::Thirty360UsEom => "30/360 us eom",
            Self::Thirty360UsNasd => "30/360 us nasd",
            Self::Thirty360Eu => "30/360 eu",
            Self::Thirty360EuM2 => "30/360 eu2",
            Self::Thirty360EuM3 => "30/360 eu3",
            Self::Thirty360EuPlus => "30/360 eu+",
            Self::Thirty365 => "30/365",
            Self::Act360 => "act/360",
            Self::Act365Fixed => "act/365 fixed",
            Self::Act365Nonleap => "act/365 nonleap",
            Self::ActActExcel => "act/act excel",
            Self::ActActIsda => "act/act isda",
            Self::ActActAfb => "act/act afb",
        }
    }

    /// Returns all valid convention strings.
    pub fn valid_strings() -> &'static [&'static str] {
        &[
            "raw",
            "30/360 us",
            "30u/360",
            "30/360 us eom",
            "30u/360 eom",
            "30/360 us nasd",
            "30u/360 nasd",
            "30/360 eu",
            "30e/360",
            "30/360 eu2",
            "30e2/360",
            "30/360 eu3",
            "30e3/360",
            "30/360 eu+",
            "30e+/360",
            "30/365",
            "act/360",
            "act/365 fixed",
            "act/365 nonleap",
            "act/act excel",
            "act/act isda",
            "act/365 isda",
            "act/act afb",
        ]
    }
}

impl fmt::Display for DayCountConvention {
    fn fmt(&self, f: &mut fmt::Formatter<'_>) -> fmt::Result {
        write!(f, "{}", self.as_str())
    }
}

impl FromStr for DayCountConvention {
    type Err = String;

    fn from_str(s: &str) -> Result<Self, Self::Err> {
        let normalized = s.to_lowercase();
        match normalized.as_str() {
            "raw" => Ok(Self::Raw),
            "30/360 us" | "30u/360" => Ok(Self::Thirty360Us),
            "30/360 us eom" | "30u/360 eom" => Ok(Self::Thirty360UsEom),
            "30/360 us nasd" | "30u/360 nasd" => Ok(Self::Thirty360UsNasd),
            "30/360 eu" | "30e/360" => Ok(Self::Thirty360Eu),
            "30/360 eu2" | "30e2/360" => Ok(Self::Thirty360EuM2),
            "30/360 eu3" | "30e3/360" => Ok(Self::Thirty360EuM3),
            "30/360 eu+" | "30e+/360" => Ok(Self::Thirty360EuPlus),
            "30/365" => Ok(Self::Thirty365),
            "act/360" => Ok(Self::Act360),
            "act/365 fixed" => Ok(Self::Act365Fixed),
            "act/365 nonleap" => Ok(Self::Act365Nonleap),
            "act/act excel" => Ok(Self::ActActExcel),
            "act/act isda" | "act/365 isda" => Ok(Self::ActActIsda),
            "act/act afb" => Ok(Self::ActActAfb),
            _ => Err(format!(
                "day count convention '{}' must be one of: {:?}",
                s,
                Self::valid_strings()
            )),
        }
    }
}

#[cfg(test)]
mod tests {
    use super::*;

    #[test]
    fn test_all_conventions() {
        let cases = vec![
            ("raw", DayCountConvention::Raw),
            ("30/360 us", DayCountConvention::Thirty360Us),
            ("30u/360", DayCountConvention::Thirty360Us),
            ("30/360 us eom", DayCountConvention::Thirty360UsEom),
            ("30u/360 eom", DayCountConvention::Thirty360UsEom),
            ("30/360 us nasd", DayCountConvention::Thirty360UsNasd),
            ("30u/360 nasd", DayCountConvention::Thirty360UsNasd),
            ("30/360 eu", DayCountConvention::Thirty360Eu),
            ("30e/360", DayCountConvention::Thirty360Eu),
            ("30/360 eu2", DayCountConvention::Thirty360EuM2),
            ("30e2/360", DayCountConvention::Thirty360EuM2),
            ("30/360 eu3", DayCountConvention::Thirty360EuM3),
            ("30e3/360", DayCountConvention::Thirty360EuM3),
            ("30/360 eu+", DayCountConvention::Thirty360EuPlus),
            ("30e+/360", DayCountConvention::Thirty360EuPlus),
            ("30/365", DayCountConvention::Thirty365),
            ("act/360", DayCountConvention::Act360),
            ("act/365 fixed", DayCountConvention::Act365Fixed),
            ("act/365 nonleap", DayCountConvention::Act365Nonleap),
            ("act/act excel", DayCountConvention::ActActExcel),
            ("act/act isda", DayCountConvention::ActActIsda),
            ("act/365 isda", DayCountConvention::ActActIsda),
            ("act/act afb", DayCountConvention::ActActAfb),
        ];

        for (s, expected) in cases {
            let result = DayCountConvention::from_str(s).unwrap();
            assert_eq!(result, expected, "Failed for string: {}", s);
        }
    }

    #[test]
    fn test_case_insensitive() {
        assert_eq!(
            DayCountConvention::from_str("RAW").unwrap(),
            DayCountConvention::Raw
        );
        assert_eq!(
            DayCountConvention::from_str("Act/360").unwrap(),
            DayCountConvention::Act360
        );
        assert_eq!(
            DayCountConvention::from_str("ACT/ACT ISDA").unwrap(),
            DayCountConvention::ActActIsda
        );
    }

    #[test]
    fn test_invalid_convention() {
        let result = DayCountConvention::from_str("invalid");
        assert!(result.is_err());
        assert!(result.unwrap_err().contains("day count convention 'invalid'"));
    }
}
