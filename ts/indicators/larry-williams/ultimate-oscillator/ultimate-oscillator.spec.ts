import { } from 'jasmine';

import { UltimateOscillator } from './ultimate-oscillator';
import { IndicatorIdentifier } from '../../core/indicator-identifier';
import { Shape } from '../../core/outputs/shape/shape';
import { UltimateOscillatorOutput } from './output';
import {
    inputHigh,
    inputLow,
    inputClose,
    expectedUltOsc,
} from './testdata';

// TA-Lib test data (252 entries), same H/L/C as TrueRange and other HLC indicators.
// Expected Excel ULTOSC output (252 entries).
describe('UltimateOscillator', () => {
  const tolerance = 1e-4;

  it('should calculate expected output (TA-Lib, 252 entries, period 7-14-28)', () => {
    const ultosc = new UltimateOscillator();

    for (let i = 0; i < inputClose.length; i++) {
      const act = ultosc.update(inputClose[i], inputHigh[i], inputLow[i]);

      if (i < 28) {
        expect(act).withContext(`index ${i} should be NaN`).toBeNaN();
        continue;
      }

      expect(Math.abs(act - expectedUltOsc[i])).withContext(`index ${i}`).toBeLessThan(tolerance);
    }
  });

  it('should match TA-Lib spot checks', () => {
    const ultosc = new UltimateOscillator();

    let result = NaN;
    for (let i = 0; i < inputClose.length; i++) {
      result = ultosc.update(inputClose[i], inputHigh[i], inputLow[i]);
    }

    // Last value at index 251 (output index 223): expected 40.0854
    expect(Math.abs(result - 40.0854)).toBeLessThan(0.001);
  });

  it('should report correct primed state', () => {
    const ultosc = new UltimateOscillator();

    expect(ultosc.isPrimed()).toBe(false);

    // Feed 28 bars (index 0..27): first sets previousClose, next 27 fill buffer.
    for (let i = 0; i < 28; i++) {
      ultosc.update(inputClose[i], inputHigh[i], inputLow[i]);
      expect(ultosc.isPrimed()).withContext(`index ${i}`).toBe(false);
    }

    // 29th bar (index 28) should prime the indicator.
    ultosc.update(inputClose[28], inputHigh[28], inputLow[28]);
    expect(ultosc.isPrimed()).toBe(true);
  });

  it('should handle NaN passthrough', () => {
    const ultosc = new UltimateOscillator();

    expect(ultosc.update(NaN, 1, 1)).toBeNaN();
    expect(ultosc.update(1, NaN, 1)).toBeNaN();
    expect(ultosc.update(1, 1, NaN)).toBeNaN();
  });

  it('should validate parameters', () => {
    expect(() => new UltimateOscillator({ length1: 1 })).toThrow();
    expect(() => new UltimateOscillator({ length2: 1 })).toThrow();
    expect(() => new UltimateOscillator({ length3: 1 })).toThrow();
    expect(() => new UltimateOscillator({ length1: 5, length2: 10, length3: 20 })).not.toThrow();
  });

  it('should return correct metadata', () => {
    const ultosc = new UltimateOscillator();
    const meta = ultosc.metadata();

    expect(meta.identifier).toBe(IndicatorIdentifier.UltimateOscillator);
    expect(meta.mnemonic).toBe('ultosc(7, 14, 28)');
    expect(meta.outputs.length).toBe(1);
    expect(meta.outputs[0].kind).toBe(UltimateOscillatorOutput.UltimateOscillatorValue);
    expect(meta.outputs[0].shape).toBe(Shape.Scalar);
  });
});
