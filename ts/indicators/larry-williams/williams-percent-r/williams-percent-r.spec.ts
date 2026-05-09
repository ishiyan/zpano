import { } from 'jasmine';

import { WilliamsPercentR } from './williams-percent-r';
import { IndicatorIdentifier } from '../../core/indicator-identifier';
import { Shape } from '../../core/outputs/shape/shape';
import { WilliamsPercentROutput } from './output';
import {
    inputHigh,
    inputLow,
    inputClose,
    expected14,
    expected2,
} from './testdata';

// MBST/TA-Lib test data (252 entries), extracted from WilliamsPercentRTest.cs.
// Expected %R output with period=14 (252 entries). First 13 are NaN.
// Expected %R output with period=2 (252 entries). First 1 is NaN.
describe('WilliamsPercentR', () => {
  const tolerance = 1e-6;

  it('should calculate expected output (period 14, 252 entries)', () => {
    const w = new WilliamsPercentR(14);

    for (let i = 0; i < inputClose.length; i++) {
      const act = w.update(inputClose[i], inputHigh[i], inputLow[i]);

      if (i < 13) {
        expect(act).toBeNaN();
        continue;
      }

      expect(Math.abs(act - expected14[i])).toBeLessThan(tolerance);
    }
  });

  it('should calculate expected output (period 2, 252 entries)', () => {
    const w = new WilliamsPercentR(2);

    for (let i = 0; i < inputClose.length; i++) {
      const act = w.update(inputClose[i], inputHigh[i], inputLow[i]);

      if (i < 1) {
        expect(act).toBeNaN();
        continue;
      }

      expect(Math.abs(act - expected2[i])).toBeLessThan(tolerance);
    }
  });

  it('should report correct primed state', () => {
    const w = new WilliamsPercentR(14);

    expect(w.isPrimed()).toBe(false);

    for (let i = 0; i < 13; i++) {
      w.update(inputClose[i], inputHigh[i], inputLow[i]);
      expect(w.isPrimed()).toBe(false);
    }

    w.update(inputClose[13], inputHigh[13], inputLow[13]);
    expect(w.isPrimed()).toBe(true);

    w.update(inputClose[14], inputHigh[14], inputLow[14]);
    expect(w.isPrimed()).toBe(true);
  });

  it('should handle NaN passthrough', () => {
    const w = new WilliamsPercentR(14);

    expect(w.update(NaN, 1, 1)).toBeNaN();
    expect(w.update(1, NaN, 1)).toBeNaN();
    expect(w.update(1, 1, NaN)).toBeNaN();
  });

  it('should handle scalar updates (H=L=C)', () => {
    const w = new WilliamsPercentR(14);

    for (let i = 0; i < 13; i++) {
      expect(w.updateSample(9.0)).toBeNaN();
    }

    expect(w.updateSample(9.0)).toBe(0);
  });

  it('should return correct metadata', () => {
    const w = new WilliamsPercentR(14);
    const meta = w.metadata();

    expect(meta.identifier).toBe(IndicatorIdentifier.WilliamsPercentR);
    expect(meta.mnemonic).toBe('willr');
    expect(meta.description).toBe('Williams %R');
    expect(meta.outputs.length).toBe(1);
    expect(meta.outputs[0].kind).toBe(WilliamsPercentROutput.WilliamsPercentRValue);
    expect(meta.outputs[0].shape).toBe(Shape.Scalar);
  });
});
