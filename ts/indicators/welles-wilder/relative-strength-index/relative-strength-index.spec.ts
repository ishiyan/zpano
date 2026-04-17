import { } from 'jasmine';

import { RelativeStrengthIndex } from './relative-strength-index';
import { IndicatorType } from '../../core/indicator-type';
import { OutputType } from '../../core/outputs/output-type';
import { RelativeStrengthIndexOutput } from './relative-strength-index-output';

// Test data from TA-Lib reference (length=9, 25 entries).
const input = [
  91.15, 90.50, 92.55, 94.70, 95.55, 94.00, 91.30, 91.95, 92.45, 93.80,
  92.50, 94.55, 96.75, 97.80, 98.40, 98.15, 96.70, 98.85, 98.90, 100.50,
  102.60, 104.80, 103.80, 103.10, 102.00,
];

const expected = [
  NaN, NaN, NaN, NaN, NaN, NaN, NaN, NaN, NaN,
  60.6425702811244, 54.2677448337826, 61.4558190165176, 67.6034767388667,
  70.1590191481383, 71.5992400904851, 70.0152589447766, 61.1833361324987,
  67.9312249318593, 68.076417836971, 72.5504646296262, 77.2568847385616,
  81.0801123570899, 74.6619680507228, 70.2808713845906, 63.6754215506388,
];

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

    expect(meta.type).toBe(IndicatorType.RelativeStrengthIndex);
    expect(meta.mnemonic).toBe('rsi(14)');
    expect(meta.description).toBe('Relative Strength Index rsi(14)');
    expect(meta.outputs.length).toBe(1);
    expect(meta.outputs[0].kind).toBe(RelativeStrengthIndexOutput.RelativeStrengthIndexValue);
    expect(meta.outputs[0].type).toBe(OutputType.Scalar);
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
