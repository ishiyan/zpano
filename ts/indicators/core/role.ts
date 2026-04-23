/** Classifies the semantic role a single indicator output plays in analysis. */
export enum Role {
  /** A trend-following line that smooths price action. */
  Smoother = 1,

  /** Upper/lower channel bounds drawn around price. */
  Envelope,

  /** A generic overlay drawn on the price pane (e.g., SAR dots). */
  Overlay,

  /** A variable-length sequence of (offset, value) points. */
  Polyline,

  /** A centered, unbounded momentum-style series. */
  Oscillator,

  /** An oscillator confined to a fixed range (e.g., 0..100). */
  BoundedOscillator,

  /** A dispersion-style measure (standard deviation, ATR, etc.). */
  Volatility,

  /** An accumulation/distribution-style volume flow measure. */
  VolumeFlow,

  /** A direction-of-movement measure (DI/DM family). */
  Directional,

  /** A dominant cycle length output. */
  CyclePeriod,

  /** A dominant cycle phase/angle output. */
  CyclePhase,

  /** A fractal-dimension-style measure. */
  FractalDimension,

  /** A multi-row spectral heat-map column. */
  Spectrum,

  /** A derived signal line (e.g., MACD signal). */
  Signal,

  /** A bar-style difference series. */
  Histogram,

  /** A discrete regime/state indicator. */
  RegimeFlag,

  /** A correlation-coefficient-style measure. */
  Correlation
}
