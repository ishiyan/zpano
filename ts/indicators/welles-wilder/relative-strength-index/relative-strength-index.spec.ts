import { } from 'jasmine';

import { RelativeStrengthIndex } from './relative-strength-index';
import { IndicatorIdentifier } from '../../core/indicator-identifier';
import { Shape } from '../../core/outputs/shape/shape';
import { RelativeStrengthIndexOutput } from './output';
import { input, expected } from './testdata';

// Test data from TA-Lib reference (length=9, 25 entries).
describe('RelativeStrengthIndex', () => {
  const tolerance = 1e-9;

  it('should throw if length is less than 2', () => {
    expect(() => { new RelativeStrengthIndex({ length: 1 }); }).toThrow();
  });

  it('should calculate expected output (length=9, TA-Lib)', () => {
    const rsi = new RelativeStrengthIndex({ length: 9 });

    for (let i = 0; i < input.length; i++) {
      const act = rsi.update(input[i]);

      if (i < 9) {
        expect(act).toBeNaN();
        expect(rsi.isPrimed()).toBe(false);
        continue;
      }

      expect(rsi.isPrimed()).toBe(true);
      expect(Math.abs(act - expected[i])).toBeLessThan(tolerance);
    }

    expect(rsi.update(Number.NaN)).toBeNaN();
  });

  it('should report correct primed state (length=5)', () => {
    const rsi = new RelativeStrengthIndex({ length: 5 });

    expect(rsi.isPrimed()).toBe(false);

    // 5 updates (values 1-5): NOT primed.
    for (let i = 1; i <= 5; i++) {
      rsi.update(i);
      expect(rsi.isPrimed()).toBe(false);
    }

    // 6th update: primed.
    rsi.update(6);
    expect(rsi.isPrimed()).toBe(true);

    // Further updates remain primed.
    for (let i = 7; i <= 11; i++) {
      rsi.update(i);
      expect(rsi.isPrimed()).toBe(true);
    }
  });

  it('should return correct metadata', () => {
    const rsi = new RelativeStrengthIndex({ length: 14 });
    const meta = rsi.metadata();

    expect(meta.identifier).toBe(IndicatorIdentifier.RelativeStrengthIndex);
    expect(meta.mnemonic).toBe('rsi(14)');
    expect(meta.description).toBe('Relative Strength Index rsi(14)');
    expect(meta.outputs.length).toBe(1);
    expect(meta.outputs[0].kind).toBe(RelativeStrengthIndexOutput.RelativeStrengthIndexValue);
    expect(meta.outputs[0].shape).toBe(Shape.Scalar);
  });

  it('should return RSI values in [0, 100] range', () => {
    const rsi = new RelativeStrengthIndex({ length: 9 });

    for (let i = 0; i < input.length; i++) {
      const act = rsi.update(input[i]);

      if (i >= 9) {
        expect(act).toBeGreaterThanOrEqual(0);
        expect(act).toBeLessThanOrEqual(100);
      }
    }
  });
});
