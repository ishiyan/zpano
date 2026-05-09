import { } from 'jasmine';

import { TrueRange } from './true-range';
import { IndicatorIdentifier } from '../../core/indicator-identifier';
import { Shape } from '../../core/outputs/shape/shape';
import { TrueRangeOutput } from './output';
import {
    inputHigh,
    inputLow,
    inputClose,
    expectedTr,
} from './testdata';

// TA-Lib test data (252 entries), extracted programmatically from TrueRangeTest.cs.
describe('TrueRange', () => {
  const tolerance = 1e-3;

  it('should calculate expected output (TA-Lib, 252 entries)', () => {
    const tr = new TrueRange();

    for (let i = 0; i < inputClose.length; i++) {
      const act = tr.update(inputClose[i], inputHigh[i], inputLow[i]);

      if (i === 0) {
        expect(act).toBeNaN();
        continue;
      }

      expect(Math.abs(act - expectedTr[i])).toBeLessThan(tolerance);
    }
  });

  it('should report correct primed state', () => {
    const tr = new TrueRange();

    expect(tr.isPrimed()).toBe(false);

    tr.update(inputClose[0], inputHigh[0], inputLow[0]);
    expect(tr.isPrimed()).toBe(false);

    tr.update(inputClose[1], inputHigh[1], inputLow[1]);
    expect(tr.isPrimed()).toBe(true);

    tr.update(inputClose[2], inputHigh[2], inputLow[2]);
    expect(tr.isPrimed()).toBe(true);
  });

  it('should handle NaN passthrough', () => {
    const tr = new TrueRange();

    expect(tr.update(NaN, 1, 1)).toBeNaN();
    expect(tr.update(1, NaN, 1)).toBeNaN();
    expect(tr.update(1, 1, NaN)).toBeNaN();
  });

  it('should handle scalar updates (H=L=C)', () => {
    const tr = new TrueRange();

    expect(tr.updateSample(100)).toBeNaN();

    const v1 = tr.updateSample(105);
    expect(Math.abs(v1 - 5.0)).toBeLessThan(1e-10);

    const v2 = tr.updateSample(102);
    expect(Math.abs(v2 - 3.0)).toBeLessThan(1e-10);
  });

  it('should return correct metadata', () => {
    const tr = new TrueRange();
    const meta = tr.metadata();

    expect(meta.identifier).toBe(IndicatorIdentifier.TrueRange);
    expect(meta.mnemonic).toBe('tr');
    expect(meta.description).toBe('True Range');
    expect(meta.outputs.length).toBe(1);
    expect(meta.outputs[0].kind).toBe(TrueRangeOutput.TrueRangeValue);
    expect(meta.outputs[0].shape).toBe(Shape.Scalar);
  });
});
