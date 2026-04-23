/** Enumerates types of techniques to estimate an instantaneous period using a Hilbert transformer. */
export enum HilbertTransformerCycleEstimatorType {

  /** An instantaneous period estimation based on the homodyne discriminator technique. */
  HomodyneDiscriminator,

  /** An instantaneous period estimation based on the homodyne discriminator technique
   * (TA-Lib implementation with unrolled loops).
   */
  HomodyneDiscriminatorUnrolled,

  /** An instantaneous period estimation based on the phase accumulation technique. */
  PhaseAccumulator,

  /** An instantaneous period estimation based on the dual differentiation technique. */
  DualDifferentiator
}
