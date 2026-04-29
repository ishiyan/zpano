/// Classifies the semantic role a single indicator output plays in analysis.
#[derive(Debug, Clone, Copy, PartialEq, Eq)]
#[repr(u8)]
pub enum Role {
    /// Trend-following line that smooths price action.
    Smoother = 1,
    /// Upper/lower channel bounds drawn around price.
    Envelope = 2,
    /// Generic overlay drawn on the price pane.
    Overlay = 3,
    /// Variable-length sequence of (offset, value) points.
    Polyline = 4,
    /// Centered, unbounded momentum-style series.
    Oscillator = 5,
    /// Oscillator confined to a fixed range (e.g., 0..100).
    BoundedOscillator = 6,
    /// Dispersion-style measure (standard deviation, ATR, etc.).
    Volatility = 7,
    /// Accumulation/distribution-style volume flow measure.
    VolumeFlow = 8,
    /// Direction-of-movement measure (DI/DM family).
    Directional = 9,
    /// Dominant cycle length output.
    CyclePeriod = 10,
    /// Dominant cycle phase/angle output.
    CyclePhase = 11,
    /// Fractal-dimension-style measure.
    FractalDimension = 12,
    /// Multi-row spectral heat-map column.
    Spectrum = 13,
    /// Derived signal line (e.g., MACD signal).
    Signal = 14,
    /// Bar-style difference series.
    Histogram = 15,
    /// Discrete regime/state indicator.
    RegimeFlag = 16,
    /// Correlation-coefficient-style measure.
    Correlation = 17,
}

impl Role {
    pub fn as_str(&self) -> &'static str {
        match self {
            Self::Smoother => "smoother",
            Self::Envelope => "envelope",
            Self::Overlay => "overlay",
            Self::Polyline => "polyline",
            Self::Oscillator => "oscillator",
            Self::BoundedOscillator => "boundedOscillator",
            Self::Volatility => "volatility",
            Self::VolumeFlow => "volumeFlow",
            Self::Directional => "directional",
            Self::CyclePeriod => "cyclePeriod",
            Self::CyclePhase => "cyclePhase",
            Self::FractalDimension => "fractalDimension",
            Self::Spectrum => "spectrum",
            Self::Signal => "signal",
            Self::Histogram => "histogram",
            Self::RegimeFlag => "regimeFlag",
            Self::Correlation => "correlation",
        }
    }
}

impl std::fmt::Display for Role {
    fn fmt(&self, f: &mut std::fmt::Formatter<'_>) -> std::fmt::Result {
        f.write_str(self.as_str())
    }
}
