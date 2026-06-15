/** Enumerates outputs of the Quantum Price Levels indicator. */
export enum QuantumPriceLevelsOutput {

  /** The anharmonic coefficient (lambda) of the quantum potential well. */
  Lambda = 0,

  /** The population standard deviation of the price-return ratios in the window. */
  ReturnStdDev = 1,

  /** The normalized QPR multipliers (1 + scaleFactor*sigma*QPR(n)), one per level. */
  NormalizedMultipliers = 2,

  /** The resistance price levels above the current price (price * NQPR(n)). */
  Resistances = 3,

  /** The support price levels below the current price (price / NQPR(n)). */
  Supports = 4,
}
