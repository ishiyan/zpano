/// Classifies whether an indicator adapts its parameters to market conditions.
#[derive(Debug, Clone, Copy, PartialEq, Eq)]
#[repr(u8)]
pub enum Adaptivity {
    /// Fixed parameters.
    Static = 1,
    /// Adapts parameters to market conditions.
    Adaptive = 2,
}

impl Adaptivity {
    pub fn as_str(&self) -> &'static str {
        match self {
            Self::Static => "static",
            Self::Adaptive => "adaptive",
        }
    }
}

impl std::fmt::Display for Adaptivity {
    fn fmt(&self, f: &mut std::fmt::Formatter<'_>) -> std::fmt::Result {
        f.write_str(self.as_str())
    }
}
