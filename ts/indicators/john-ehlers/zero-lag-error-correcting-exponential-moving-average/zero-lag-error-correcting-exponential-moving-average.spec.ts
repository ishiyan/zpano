import { } from 'jasmine';

import { ZeroLagErrorCorrectingExponentialMovingAverage } from './zero-lag-error-correcting-exponential-moving-average';
import { ZeroLagErrorCorrectingExponentialMovingAverageParams } from './zero-lag-error-correcting-exponential-moving-average-params';
import { IndicatorType } from '../../core/indicator-type';
import { OutputType } from '../../core/outputs/output-type';
import { ZeroLagErrorCorrectingExponentialMovingAverageOutput } from './zero-lag-error-correcting-exponential-moving-average-output';

const defaultParams: ZeroLagErrorCorrectingExponentialMovingAverageParams = {
  smoothingFactor: 0.095,
  gainLimit: 5,
  gainStep: 0.1,
};

describe('ZeroLagErrorCorrectingExponentialMovingAverage', () => {

  it('should throw if smoothing factor is 0', () => {
    expect(() => { new ZeroLagErrorCorrectingExponentialMovingAverage({ ...defaultParams, smoothingFactor: 0 }); }).toThrow();
  });

  it('should throw if smoothing factor is negative', () => {
    expect(() => { new ZeroLagErrorCorrectingExponentialMovingAverage({ ...defaultParams, smoothingFactor: -0.1 }); }).toThrow();
  });

  it('should throw if smoothing factor is greater than 1', () => {
    expect(() => { new ZeroLagErrorCorrectingExponentialMovingAverage({ ...defaultParams, smoothingFactor: 1.1 }); }).toThrow();
  });

  it('should not throw if smoothing factor is 1', () => {
    expect(() => { new ZeroLagErrorCorrectingExponentialMovingAverage({ ...defaultParams, smoothingFactor: 1 }); }).not.toThrow();
  });

  it('should throw if gain limit is 0', () => {
    expect(() => { new ZeroLagErrorCorrectingExponentialMovingAverage({ ...defaultParams, gainLimit: 0 }); }).toThrow();
  });

  it('should throw if gain limit is negative', () => {
    expect(() => { new ZeroLagErrorCorrectingExponentialMovingAverage({ ...defaultParams, gainLimit: -1 }); }).toThrow();
  });

  it('should throw if gain step is 0', () => {
    expect(() => { new ZeroLagErrorCorrectingExponentialMovingAverage({ ...defaultParams, gainStep: 0 }); }).toThrow();
  });

  it('should throw if gain step is negative', () => {
    expect(() => { new ZeroLagErrorCorrectingExponentialMovingAverage({ ...defaultParams, gainStep: -0.1 }); }).toThrow();
  });

  it('should return expected mnemonic', () => {
    const z = new ZeroLagErrorCorrectingExponentialMovingAverage(defaultParams);
    expect(z.metadata().mnemonic).toBe('zecema(0.095, 5, 0.1)');
  });

  it('should report correct primed state', () => {
    const z = new ZeroLagErrorCorrectingExponentialMovingAverage(defaultParams);

    expect(z.isPrimed()).toBe(false);

    // First 2 updates should not prime.
    z.update(100);
    expect(z.isPrimed()).toBe(false);

    z.update(100);
    expect(z.isPrimed()).toBe(false);

    // 3rd update primes.
    z.update(100);
    expect(z.isPrimed()).toBe(true);
  });

  it('should return NaN before primed', () => {
    const z = new ZeroLagErrorCorrectingExponentialMovingAverage(defaultParams);

    expect(z.update(100)).toBeNaN();
    expect(z.update(100)).toBeNaN();

    // 3rd update should return a number.
    expect(z.update(100)).not.toBeNaN();
  });

  it('should pass through NaN', () => {
    const z = new ZeroLagErrorCorrectingExponentialMovingAverage(defaultParams);
    expect(z.update(Number.NaN)).toBeNaN();
    expect(z.isPrimed()).toBe(false);
  });

  it('should converge to constant for constant input', () => {
    const z = new ZeroLagErrorCorrectingExponentialMovingAverage(defaultParams);
    const val = 42;

    // Prime with constant.
    z.update(val);
    z.update(val);

    const result = z.update(val);
    expect(Math.abs(result - val)).toBeLessThan(1e-6);

    // Further constant updates.
    for (let i = 0; i < 10; i++) {
      expect(Math.abs(z.update(val) - val)).toBeLessThan(1e-6);
    }
  });

  it('should return correct metadata', () => {
    const z = new ZeroLagErrorCorrectingExponentialMovingAverage(defaultParams);
    const meta = z.metadata();

    expect(meta.type).toBe(IndicatorType.ZeroLagErrorCorrectingExponentialMovingAverage);
    expect(meta.mnemonic).toBe('zecema(0.095, 5, 0.1)');
    expect(meta.description).toBe('Zero-lag Error-Correcting Exponential Moving Average zecema(0.095, 5, 0.1)');
    expect(meta.outputs.length).toBe(1);
    expect(meta.outputs[0].kind).toBe(ZeroLagErrorCorrectingExponentialMovingAverageOutput.ZeroLagErrorCorrectingExponentialMovingAverageValue);
    expect(meta.outputs[0].type).toBe(OutputType.Scalar);
  });
});
