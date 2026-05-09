import { } from 'jasmine';

import { Aroon } from './aroon';
import { AroonOutput } from './output';
import { IndicatorIdentifier } from '../../core/indicator-identifier';
import { Shape } from '../../core/outputs/shape/shape';
import { highs, lows, expected } from './testdata';

// Standard HL test data (252 entries) from TA-Lib.
// Expected: [up, down, osc] for each of 252 bars. NaN for first 14.
describe('Aroon', () => {

  it('should calculate length=14 full data against all 252 expected values', () => {
    const tolerance = 1e-6;
    const ind = new Aroon({ length: 14 });

    for (let i = 0; i < 252; i++) {
      const [up, down, osc] = ind.update(highs[i], lows[i]);

      if (isNaN(expected[i][0])) {
        expect(up).withContext(`[${i}] AroonUp`).toBeNaN();
        continue;
      }

      expect(Math.abs(up - expected[i][0])).withContext(`[${i}] AroonUp`).toBeLessThan(tolerance);
      expect(Math.abs(down - expected[i][1])).withContext(`[${i}] AroonDown`).toBeLessThan(tolerance);
      expect(Math.abs(osc - expected[i][2])).withContext(`[${i}] AroonOsc`).toBeLessThan(tolerance);
    }
  });

  it('should report correct primed state', () => {
    const ind = new Aroon({ length: 14 });

    expect(ind.isPrimed()).toBe(false);

    for (let i = 0; i < 14; i++) {
      ind.update(highs[i], lows[i]);
      expect(ind.isPrimed()).withContext(`index ${i}`).toBe(false);
    }

    ind.update(highs[14], lows[14]);
    expect(ind.isPrimed()).toBe(true);
  });

  it('should handle NaN passthrough', () => {
    const ind = new Aroon({ length: 14 });

    const [up, down, osc] = ind.update(NaN, 1.0);
    expect(up).toBeNaN();
    expect(down).toBeNaN();
    expect(osc).toBeNaN();
  });

  it('should return correct metadata', () => {
    const ind = new Aroon({ length: 14 });
    const meta = ind.metadata();

    expect(meta.identifier).toBe(IndicatorIdentifier.Aroon);
    expect(meta.mnemonic).toBe('aroon(14)');
    expect(meta.outputs.length).toBe(3);
    expect(meta.outputs[0].kind).toBe(AroonOutput.Up);
    expect(meta.outputs[0].shape).toBe(Shape.Scalar);
    expect(meta.outputs[1].kind).toBe(AroonOutput.Down);
    expect(meta.outputs[2].kind).toBe(AroonOutput.Osc);
  });

  it('should validate parameters', () => {
    expect(() => new Aroon({ length: 1 })).toThrow();
    expect(() => new Aroon({ length: 0 })).toThrow();
    expect(() => new Aroon({ length: -1 })).toThrow();
    expect(() => new Aroon({ length: 2 })).not.toThrow();
    expect(() => new Aroon({ length: 14 })).not.toThrow();
  });
});
