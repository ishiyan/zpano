/** Classifies whether an indicator adapts its parameters to market conditions. */
export enum Adaptivity {
  /** Fixed parameters. */
  Static = 1,

  /**
   * Adapts parameters to market conditions
   * (e.g., via dominant cycle, efficiency ratio, or fractal dimension).
   */
  Adaptive
}
