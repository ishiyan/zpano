import { } from 'jasmine';

import { DirectionalMovementPlus } from './directional-movement-plus';
import { IndicatorIdentifier } from '../../core/indicator-identifier';
import { Shape } from '../../core/outputs/shape/shape';
import { DirectionalMovementPlusOutput } from './output';
import {
    inputHigh,
    inputLow,
    expectedDmp1,
    expectedDmp14,
} from './testdata';

// TA-Lib test data (252 entries), extracted programmatically from DirectionalMovementPlusTest.cs.
// Expected +DM1 (length=1), 252 entries.
// Expected +DM14 (length=14), 252 entries. First 14 entries are NaN.
// C# has 13 NaN (indices 0-12) plus vestigial value at index 13 (10.28). Per offset bug fix,
// we use 14 NaN (indices 0-13), drop the vestigial value, and start expected values at index 14.
describe('DirectionalMovementPlus', () => {

  describe('constructor', () => {
    it('should create with valid length', () => {
      const dmp = new DirectionalMovementPlus(14);
      expect(dmp.length).toBe(14);
      expect(dmp.isPrimed()).toBe(false);
    });

    it('should throw for length 0', () => {
      expect(() => new DirectionalMovementPlus(0)).toThrow();
    });

    it('should throw for negative length', () => {
      expect(() => new DirectionalMovementPlus(-8)).toThrow();
    });
  });

  describe('isPrimed', () => {
    it('should prime after 2 updates for length=1', () => {
      const dmp = new DirectionalMovementPlus(1);
      expect(dmp.isPrimed()).toBe(false);

      dmp.update(inputHigh[0], inputLow[0]);
      expect(dmp.isPrimed()).toBe(false);

      dmp.update(inputHigh[1], inputLow[1]);
      expect(dmp.isPrimed()).toBe(true);
    });

    it('should not be primed during first length updates for length=14', () => {
      const dmp = new DirectionalMovementPlus(14);
      expect(dmp.isPrimed()).toBe(false);

      for (let i = 0; i < 14; i++) {
        dmp.update(inputHigh[i], inputLow[i]);
        expect(dmp.isPrimed()).toBe(false);
      }

      dmp.update(inputHigh[14], inputLow[14]);
      expect(dmp.isPrimed()).toBe(true);
    });
  });

  describe('update', () => {
    it('should match TA-Lib reference data with length=14', () => {
      const tolerance = 1e-8;
      const dmp = new DirectionalMovementPlus(14);

      for (let i = 0; i < inputHigh.length; i++) {
        const act = dmp.update(inputHigh[i], inputLow[i]);

        if (isNaN(expectedDmp14[i])) {
          expect(isNaN(act)).toBe(true, `[${i}] expected NaN`);
        } else {
          expect(isNaN(act)).toBe(false, `[${i}] got NaN, expected ${expectedDmp14[i]}`);
          expect(Math.abs(act - expectedDmp14[i])).toBeLessThan(tolerance, `[${i}] expected ${expectedDmp14[i]}, got ${act}`);
        }
      }
    });

    it('should match TA-Lib reference data with length=1', () => {
      const tolerance = 1e-8;
      const dmp = new DirectionalMovementPlus(1);

      for (let i = 0; i < inputHigh.length; i++) {
        const act = dmp.update(inputHigh[i], inputLow[i]);

        if (isNaN(expectedDmp1[i])) {
          expect(isNaN(act)).toBe(true, `[${i}] expected NaN`);
        } else {
          expect(isNaN(act)).toBe(false, `[${i}] got NaN, expected ${expectedDmp1[i]}`);
          expect(Math.abs(act - expectedDmp1[i])).toBeLessThan(tolerance, `[${i}] expected ${expectedDmp1[i]}, got ${act}`);
        }
      }
    });
  });

  describe('NaN passthrough', () => {
    it('should return NaN for NaN inputs', () => {
      const dmp = new DirectionalMovementPlus(14);
      expect(isNaN(dmp.update(NaN, 1))).toBe(true);
      expect(isNaN(dmp.update(1, NaN))).toBe(true);
      expect(isNaN(dmp.updateSample(NaN))).toBe(true);
    });
  });

  describe('high < low swap', () => {
    it('should swap high and low when high < low', () => {
      const dmp = new DirectionalMovementPlus(1);
      dmp.update(10, 5);
      // Passing swapped values should produce the same as normal order.
      const normal = new DirectionalMovementPlus(1);
      normal.update(10, 5);
      const v1 = dmp.update(5, 12); // will swap to (12, 5)
      const v2 = normal.update(12, 5);
      expect(v1).toBe(v2);
    });
  });

  describe('metadata', () => {
    it('should return correct metadata', () => {
      const dmp = new DirectionalMovementPlus(14);
      const meta = dmp.metadata();
      expect(meta.identifier).toBe(IndicatorIdentifier.DirectionalMovementPlus);
      expect(meta.mnemonic).toBe('+dm');
      expect(meta.description).toBe('Directional Movement Plus');
      expect(meta.outputs.length).toBe(1);
      expect(meta.outputs[0].kind).toBe(DirectionalMovementPlusOutput.DirectionalMovementPlusValue);
      expect(meta.outputs[0].shape).toBe(Shape.Scalar);
      expect(meta.outputs[0].mnemonic).toBe('+dm');
      expect(meta.outputs[0].description).toBe('Directional Movement Plus');
    });
  });
});
