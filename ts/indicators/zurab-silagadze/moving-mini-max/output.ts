/** Enumerates outputs of the Moving Mini-Max indicator. */
export enum MovingMiniMaxOutput {

  /** The up mini-max value at the most recent bar (emphasizes local maxima). */
  Up = 0,

  /** The down mini-max value at the most recent bar (emphasizes local minima). */
  Down = 1,

  /** The detected resistance levels, sorted by strength (strongest first). */
  Resistances = 2,

  /** The detected support levels, sorted by strength (strongest first). */
  Supports = 3,

  /** The full up mini-max probability distribution over the window (sums to 1.0). */
  UpDistribution = 4,

  /** The full down mini-max probability distribution over the window (sums to 1.0). */
  DownDistribution = 5,
}
