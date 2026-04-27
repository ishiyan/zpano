/// Classifies the semantic role a single indicator output plays in analysis.
pub const Role = enum(u8) {
    /// Trend-following line that smooths price action.
    smoother = 1,
    /// Upper/lower channel bounds drawn around price.
    envelope = 2,
    /// Generic overlay drawn on the price pane.
    overlay = 3,
    /// Variable-length sequence of (offset, value) points.
    polyline = 4,
    /// Centered, unbounded momentum-style series.
    oscillator = 5,
    /// Oscillator confined to a fixed range (e.g., 0..100).
    bounded_oscillator = 6,
    /// Dispersion-style measure (standard deviation, ATR, etc.).
    volatility = 7,
    /// Accumulation/distribution-style volume flow measure.
    volume_flow = 8,
    /// Direction-of-movement measure (DI/DM family).
    directional = 9,
    /// Dominant cycle length output.
    cycle_period = 10,
    /// Dominant cycle phase/angle output.
    cycle_phase = 11,
    /// Fractal-dimension-style measure.
    fractal_dimension = 12,
    /// Multi-row spectral heat-map column.
    spectrum = 13,
    /// Derived signal line (e.g., MACD signal).
    signal = 14,
    /// Bar-style difference series.
    histogram = 15,
    /// Discrete regime/state indicator.
    regime_flag = 16,
    /// Correlation-coefficient-style measure.
    correlation = 17,

    pub fn asStr(self: Role) []const u8 {
        return switch (self) {
            .smoother => "smoother",
            .envelope => "envelope",
            .overlay => "overlay",
            .polyline => "polyline",
            .oscillator => "oscillator",
            .bounded_oscillator => "boundedOscillator",
            .volatility => "volatility",
            .volume_flow => "volumeFlow",
            .directional => "directional",
            .cycle_period => "cyclePeriod",
            .cycle_phase => "cyclePhase",
            .fractal_dimension => "fractalDimension",
            .spectrum => "spectrum",
            .signal => "signal",
            .histogram => "histogram",
            .regime_flag => "regimeFlag",
            .correlation => "correlation",
        };
    }
};
