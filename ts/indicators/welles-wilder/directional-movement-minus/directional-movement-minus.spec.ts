import { } from 'jasmine';

import { DirectionalMovementMinus } from './directional-movement-minus';
import { IndicatorIdentifier } from '../../core/indicator-identifier';
import { Shape } from '../../core/outputs/shape/shape';
import { DirectionalMovementMinusOutput } from './output';
import {
    inputHigh,
    inputLow,
    expectedDmm1,
    expectedDmm14,
} from './testdata';

// TA-Lib test data (252 entries), extracted programmatically from DirectionalMovementMinusTest.cs.
// Expected -DM1 (length=1), 252 entries.
// Expected -DM14 (length=14), 252 entries. First 14 entries are NaN.
describe('DirectionalMovementMinus', () => {

  describe('constructor', () => {
    it('should create with valid length', () => {
      const dmm = new DirectionalMovementMinus(14);
      expect(dmm.length).toBe(14);
      expect(dmm.isPrimed()).toBe(false);
    });

    it('should throw for length 0', () => {
      expect(() => new DirectionalMovementMinus(0)).toThrow();
    });

    it('should throw for negative length', () => {
      expect(() => new DirectionalMovementMinus(-8)).toThrow();
    });
  });

  describe('isPrimed', () => {
    it('should prime after 2 updates for length=1', () => {
      const dmm = new DirectionalMovementMinus(1);
      expect(dmm.isPrimed()).toBe(false);

      dmm.update(inputHigh[0], inputLow[0]);
      expect(dmm.isPrimed()).toBe(false);

      dmm.update(inputHigh[1], inputLow[1]);
      expect(dmm.isPrimed()).toBe(true);
    });

    it('should not be primed during first length updates for length=14', () => {
      const dmm = new DirectionalMovementMinus(14);
      expect(dmm.isPrimed()).toBe(false);

      for (let i = 0; i < 14; i++) {
        dmm.update(inputHigh[i], inputLow[i]);
        expect(dmm.isPrimed()).toBe(false);
      }

      dmm.update(inputHigh[14], inputLow[14]);
      expect(dmm.isPrimed()).toBe(true);
    });
  });

  describe('update', () => {
    it('should match TA-Lib reference data with length=14', () => {
      const tolerance = 1e-8;
      const dmm = new DirectionalMovementMinus(14);

      for (let i = 0; i < inputHigh.length; i++) {
        const act = dmm.update(inputHigh[i], inputLow[i]);

        if (isNaN(expectedDmm14[i])) {
          expect(isNaN(act)).toBe(true, `[${i}] expected NaN`);
        } else {
          expect(isNaN(act)).toBe(false, `[${i}] got NaN, expected ${expectedDmm14[i]}`);
          expect(Math.abs(act - expectedDmm14[i])).toBeLessThan(tolerance, `[${i}] expected ${expectedDmm14[i]}, got ${act}`);
        }
      }
    });

    it('should match TA-Lib reference data with length=1', () => {
      const tolerance = 1e-8;
      const dmm = new DirectionalMovementMinus(1);

      for (let i = 0; i < inputHigh.length; i++) {
        const act = dmm.update(inputHigh[i], inputLow[i]);

        if (isNaN(expectedDmm1[i])) {
          expect(isNaN(act)).toBe(true, `[${i}] expected NaN`);
        } else {
          expect(isNaN(act)).toBe(false, `[${i}] got NaN, expected ${expectedDmm1[i]}`);
          expect(Math.abs(act - expectedDmm1[i])).toBeLessThan(tolerance, `[${i}] expected ${expectedDmm1[i]}, got ${act}`);
        }
      }
    });
  });

  describe('NaN passthrough', () => {
    it('should return NaN for NaN inputs', () => {
      const dmm = new DirectionalMovementMinus(14);
      expect(isNaN(dmm.update(NaN, 1))).toBe(true);
      expect(isNaN(dmm.update(1, NaN))).toBe(true);
      expect(isNaN(dmm.updateSample(NaN))).toBe(true);
    });
  });

  describe('high < low swap', () => {
    it('should swap high and low when high < low', () => {
      const dmm = new DirectionalMovementMinus(1);
      dmm.update(10, 5);
      // Passing swapped values should produce the same as normal order.
      const normal = new DirectionalMovementMinus(1);
      normal.update(10, 5);
      const v1 = dmm.update(5, 12); // will swap to (12, 5)
      const v2 = normal.update(12, 5);
      expect(v1).toBe(v2);
    });
  });

  describe('metadata', () => {
    it('should return correct metadata', () => {
      const dmm = new DirectionalMovementMinus(14);
      const meta = dmm.metadata();
      expect(meta.identifier).toBe(IndicatorIdentifier.DirectionalMovementMinus);
      expect(meta.mnemonic).toBe('-dm');
      expect(meta.description).toBe('Directional Movement Minus');
      expect(meta.outputs.length).toBe(1);
      expect(meta.outputs[0].kind).toBe(DirectionalMovementMinusOutput.DirectionalMovementMinusValue);
      expect(meta.outputs[0].shape).toBe(Shape.Scalar);
      expect(meta.outputs[0].mnemonic).toBe('-dm');
      expect(meta.outputs[0].description).toBe('Directional Movement Minus');
    });
  });
});
