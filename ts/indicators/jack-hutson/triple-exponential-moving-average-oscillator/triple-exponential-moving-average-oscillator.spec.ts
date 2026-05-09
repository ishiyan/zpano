import { } from 'jasmine';

import { TripleExponentialMovingAverageOscillator } from './triple-exponential-moving-average-oscillator';
import { TripleExponentialMovingAverageOscillatorOutput } from './output';
import { IndicatorIdentifier } from '../../core/indicator-identifier';
import { Shape } from '../../core/outputs/shape/shape';
import { input, expected } from './testdata';

// Test data from TA-Lib (252 entries).
// Expected TRIX values for length=5, 252 bars.
describe('TripleExponentialMovingAverageOscillator', () => {

  it('should calculate TRIX with length=5 correctly', () => {
    const tolerance = 1e-10;
    const ind = new TripleExponentialMovingAverageOscillator({ length: 5 });

    for (let i = 0; i < input.length; i++) {
      const result = ind.update(input[i]);

      if (isNaN(expected[i])) {
        expect(isNaN(result)).toBe(true, `expected NaN at index ${i}, got ${result}`);
      } else {
        expect(Math.abs(result - expected[i])).toBeLessThanOrEqual(tolerance,
          `mismatch at index ${i}: expected ${expected[i]}, got ${result}`);
      }
    }
  });

  it('should match TaLib spot checks', () => {
    const ind = new TripleExponentialMovingAverageOscillator({ length: 5 });

    const results: number[] = [];
    for (const v of input) {
      results.push(ind.update(v));
    }

    // begIndex=13, nbElement=239.
    expect(Math.abs(results[13] - 0.2589)).toBeLessThanOrEqual(1e-4);
    expect(Math.abs(results[14] - 0.010495)).toBeLessThanOrEqual(1e-4);
    expect(Math.abs(results[250] - (-0.058))).toBeLessThanOrEqual(1e-3);
    expect(Math.abs(results[251] - (-0.095))).toBeLessThanOrEqual(1e-3);
  });

  it('should be primed at correct index', () => {
    const ind = new TripleExponentialMovingAverageOscillator({ length: 5 });

    // Lookback = 3*(5-1) + 1 = 13.
    for (let i = 0; i < 13; i++) {
      ind.update(input[i]);
      expect(ind.isPrimed()).toBe(false, `should not be primed at index ${i}`);
    }

    ind.update(input[13]);
    expect(ind.isPrimed()).toBe(true);
  });

  it('should return correct metadata', () => {
    const ind = new TripleExponentialMovingAverageOscillator({ length: 30 });
    const meta = ind.metadata();

    expect(meta.identifier).toBe(IndicatorIdentifier.TripleExponentialMovingAverageOscillator);
    expect(meta.mnemonic).toBe('trix(30)');
    expect(meta.description).toBe('Triple exponential moving average oscillator trix(30)');
    expect(meta.outputs.length).toBe(1);
    expect(meta.outputs[0].kind).toBe(TripleExponentialMovingAverageOscillatorOutput.Value);
    expect(meta.outputs[0].shape).toBe(Shape.Scalar);
  });

  it('should throw for invalid params', () => {
    expect(() => new TripleExponentialMovingAverageOscillator({ length: 0 })).toThrow();
  });

  it('should return NaN for NaN input', () => {
    const ind = new TripleExponentialMovingAverageOscillator({ length: 5 });
    expect(isNaN(ind.update(NaN))).toBe(true);
  });
});
