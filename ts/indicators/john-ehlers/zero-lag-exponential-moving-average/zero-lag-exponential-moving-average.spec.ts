import { } from 'jasmine';

import { ZeroLagExponentialMovingAverage } from './zero-lag-exponential-moving-average';
import { ZeroLagExponentialMovingAverageParams } from './zero-lag-exponential-moving-average-params';
import { IndicatorType } from '../../core/indicator-type';
import { OutputType } from '../../core/outputs/output-type';
import { ZeroLagExponentialMovingAverageOutput } from './zero-lag-exponential-moving-average-output';

const defaultParams: ZeroLagExponentialMovingAverageParams = {
  smoothingFactor: 0.25,
  velocityGainFactor: 0.5,
  velocityMomentumLength: 3,
};

describe('ZeroLagExponentialMovingAverage', () => {

  it('should throw if smoothing factor is 0', () => {
    expect(() => { new ZeroLagExponentialMovingAverage({ ...defaultParams, smoothingFactor: 0 }); }).toThrow();
  });

  it('should throw if smoothing factor is negative', () => {
    expect(() => { new ZeroLagExponentialMovingAverage({ ...defaultParams, smoothingFactor: -0.1 }); }).toThrow();
  });

  it('should throw if smoothing factor is greater than 1', () => {
    expect(() => { new ZeroLagExponentialMovingAverage({ ...defaultParams, smoothingFactor: 1.1 }); }).toThrow();
  });

  it('should not throw if smoothing factor is 1', () => {
    expect(() => { new ZeroLagExponentialMovingAverage({ ...defaultParams, smoothingFactor: 1 }); }).not.toThrow();
  });

  it('should throw if momentum length is 0', () => {
    expect(() => { new ZeroLagExponentialMovingAverage({ ...defaultParams, velocityMomentumLength: 0 }); }).toThrow();
  });

  it('should throw if momentum length is negative', () => {
    expect(() => { new ZeroLagExponentialMovingAverage({ ...defaultParams, velocityMomentumLength: -1 }); }).toThrow();
  });

  it('should return expected mnemonic', () => {
    const z = new ZeroLagExponentialMovingAverage(defaultParams);
    expect(z.metadata().mnemonic).toBe('zema(0.25, 0.5, 3)');
  });

  it('should report correct primed state', () => {
    const z = new ZeroLagExponentialMovingAverage(defaultParams);

    expect(z.isPrimed()).toBe(false);

    // First 3 updates (momentumLength=3) should not prime.
    z.update(100);
    expect(z.isPrimed()).toBe(false);

    z.update(100);
    expect(z.isPrimed()).toBe(false);

    z.update(100);
    expect(z.isPrimed()).toBe(false);

    // 4th update primes.
    z.update(100);
    expect(z.isPrimed()).toBe(true);
  });

  it('should return NaN before primed', () => {
    const z = new ZeroLagExponentialMovingAverage(defaultParams);

    expect(z.update(100)).toBeNaN();
    expect(z.update(100)).toBeNaN();
    expect(z.update(100)).toBeNaN();

    // 4th update should return a number.
    expect(z.update(100)).not.toBeNaN();
  });

  it('should pass through NaN', () => {
    const z = new ZeroLagExponentialMovingAverage(defaultParams);
    expect(z.update(Number.NaN)).toBeNaN();
    expect(z.isPrimed()).toBe(false);
  });

  it('should converge to constant for constant input', () => {
    const z = new ZeroLagExponentialMovingAverage(defaultParams);
    const val = 42;

    // Prime with constant.
    for (let i = 0; i < 3; i++) {
      z.update(val);
    }

    const result = z.update(val);
    expect(Math.abs(result - val)).toBeLessThan(1e-10);

    // Further constant updates.
    for (let i = 0; i < 10; i++) {
      expect(Math.abs(z.update(val) - val)).toBeLessThan(1e-10);
    }
  });

  it('should return correct metadata', () => {
    const z = new ZeroLagExponentialMovingAverage(defaultParams);
    const meta = z.metadata();

    expect(meta.type).toBe(IndicatorType.ZeroLagExponentialMovingAverage);
    expect(meta.mnemonic).toBe('zema(0.25, 0.5, 3)');
    expect(meta.description).toBe('Zero-lag Exponential Moving Average zema(0.25, 0.5, 3)');
    expect(meta.outputs.length).toBe(1);
    expect(meta.outputs[0].kind).toBe(ZeroLagExponentialMovingAverageOutput.ZeroLagExponentialMovingAverageValue);
    expect(meta.outputs[0].type).toBe(OutputType.Scalar);
  });
});
