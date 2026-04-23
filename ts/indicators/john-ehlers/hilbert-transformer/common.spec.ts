import { } from 'jasmine';

import { createEstimator } from './common';
import { HilbertTransformerCycleEstimatorType } from './cycle-estimator-type';
import { HilbertTransformerHomodyneDiscriminator } from './homodyne-discriminator';
import { HilbertTransformerHomodyneDiscriminatorUnrolled } from './homodyne-discriminator-unrolled';
import { HilbertTransformerPhaseAccumulator } from './phase-accumulator';
import { HilbertTransformerDualDifferentiator } from './dual-differentiator';

describe('createEstimator', () => {

  it('should default to HomodyneDiscriminator when no type is specified', () => {
    const est = createEstimator();
    expect(est instanceof HilbertTransformerHomodyneDiscriminator).toBeTrue();
  });

  it('should construct a HomodyneDiscriminator with default params', () => {
    const est = createEstimator(HilbertTransformerCycleEstimatorType.HomodyneDiscriminator);
    expect(est instanceof HilbertTransformerHomodyneDiscriminator).toBeTrue();
    expect(est.smoothingLength).toBe(4);
    expect(est.alphaEmaQuadratureInPhase).toBeCloseTo(0.2, 10);
    expect(est.alphaEmaPeriod).toBeCloseTo(0.2, 10);
  });

  it('should construct a HomodyneDiscriminatorUnrolled with default params', () => {
    const est = createEstimator(HilbertTransformerCycleEstimatorType.HomodyneDiscriminatorUnrolled);
    expect(est instanceof HilbertTransformerHomodyneDiscriminatorUnrolled).toBeTrue();
    expect(est.smoothingLength).toBe(4);
    expect(est.alphaEmaQuadratureInPhase).toBeCloseTo(0.2, 10);
    expect(est.alphaEmaPeriod).toBeCloseTo(0.2, 10);
  });

  it('should construct a PhaseAccumulator with default params', () => {
    const est = createEstimator(HilbertTransformerCycleEstimatorType.PhaseAccumulator);
    expect(est instanceof HilbertTransformerPhaseAccumulator).toBeTrue();
    expect(est.smoothingLength).toBe(4);
    expect(est.alphaEmaQuadratureInPhase).toBeCloseTo(0.15, 10);
    expect(est.alphaEmaPeriod).toBeCloseTo(0.25, 10);
  });

  it('should construct a DualDifferentiator with default params', () => {
    const est = createEstimator(HilbertTransformerCycleEstimatorType.DualDifferentiator);
    expect(est instanceof HilbertTransformerDualDifferentiator).toBeTrue();
    expect(est.smoothingLength).toBe(4);
    expect(est.alphaEmaQuadratureInPhase).toBeCloseTo(0.15, 10);
    expect(est.alphaEmaPeriod).toBeCloseTo(0.15, 10);
  });

  it('should honor explicit params', () => {
    const est = createEstimator(
      HilbertTransformerCycleEstimatorType.HomodyneDiscriminator,
      { smoothingLength: 3, alphaEmaQuadratureInPhase: 0.5, alphaEmaPeriod: 0.6 },
    );
    expect(est.smoothingLength).toBe(3);
    expect(est.alphaEmaQuadratureInPhase).toBeCloseTo(0.5, 10);
    expect(est.alphaEmaPeriod).toBeCloseTo(0.6, 10);
  });

  it('should throw for an invalid estimator type', () => {
    expect(() => createEstimator(-1 as HilbertTransformerCycleEstimatorType)).toThrow();
  });
});
