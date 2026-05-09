import { } from 'jasmine';

import { DirectionalMovementIndex } from './directional-movement-index';
import { IndicatorIdentifier } from '../../core/indicator-identifier';
import { Shape } from '../../core/outputs/shape/shape';
import { DirectionalMovementIndexOutput } from './output';
import {
    inputHigh,
    inputLow,
    inputClose,
    expectedDx14,
} from './testdata';

// TA-Lib test data (252 entries), extracted programmatically from DirectionalMovementIndexTest.cs.
// Expected DX14 (length=14), 252 entries. First 14 entries are NaN.
describe('DirectionalMovementIndex', () => {

  describe('constructor', () => {
    it('should create with valid length', () => {
      const dx = new DirectionalMovementIndex(14);
      expect(dx.length).toBe(14);
      expect(dx.isPrimed()).toBe(false);
    });

    it('should throw for length 0', () => {
      expect(() => new DirectionalMovementIndex(0)).toThrow();
    });

    it('should throw for negative length', () => {
      expect(() => new DirectionalMovementIndex(-8)).toThrow();
    });
  });

  describe('isPrimed', () => {
    it('should not be primed during first length updates for length=14', () => {
      const dx = new DirectionalMovementIndex(14);
      expect(dx.isPrimed()).toBe(false);

      for (let i = 0; i < 14; i++) {
        dx.update(inputClose[i], inputHigh[i], inputLow[i]);
        expect(dx.isPrimed()).toBe(false);
      }

      dx.update(inputClose[14], inputHigh[14], inputLow[14]);
      expect(dx.isPrimed()).toBe(true);
    });
  });

  describe('update', () => {
    it('should match TA-Lib reference data with length=14', () => {
      const tolerance = 1e-8;
      const dx = new DirectionalMovementIndex(14);

      for (let i = 0; i < inputHigh.length; i++) {
        const act = dx.update(inputClose[i], inputHigh[i], inputLow[i]);

        if (isNaN(expectedDx14[i])) {
          expect(isNaN(act)).toBe(true, `[${i}] expected NaN`);
        } else {
          expect(isNaN(act)).toBe(false, `[${i}] got NaN, expected ${expectedDx14[i]}`);
          expect(Math.abs(act - expectedDx14[i])).toBeLessThan(tolerance, `[${i}] expected ${expectedDx14[i]}, got ${act}`);
        }
      }
    });
  });

  describe('NaN passthrough', () => {
    it('should return NaN for NaN inputs', () => {
      const dx = new DirectionalMovementIndex(14);
      expect(isNaN(dx.update(NaN, 1, 1))).toBe(true);
      expect(isNaN(dx.update(1, NaN, 1))).toBe(true);
      expect(isNaN(dx.update(1, 1, NaN))).toBe(true);
      expect(isNaN(dx.updateSample(NaN))).toBe(true);
    });
  });

  describe('metadata', () => {
    it('should return correct metadata', () => {
      const dx = new DirectionalMovementIndex(14);
      const meta = dx.metadata();
      expect(meta.identifier).toBe(IndicatorIdentifier.DirectionalMovementIndex);
      expect(meta.mnemonic).toBe('dx');
      expect(meta.description).toBe('Directional Movement Index');
      expect(meta.outputs.length).toBe(7);
      expect(meta.outputs[0].kind).toBe(DirectionalMovementIndexOutput.DirectionalMovementIndexValue);
      expect(meta.outputs[0].shape).toBe(Shape.Scalar);
      expect(meta.outputs[0].mnemonic).toBe('dx');
      expect(meta.outputs[0].description).toBe('Directional Movement Index');
      expect(meta.outputs[1].kind).toBe(DirectionalMovementIndexOutput.DirectionalIndicatorPlusValue);
      expect(meta.outputs[2].kind).toBe(DirectionalMovementIndexOutput.DirectionalIndicatorMinusValue);
      expect(meta.outputs[3].kind).toBe(DirectionalMovementIndexOutput.DirectionalMovementPlusValue);
      expect(meta.outputs[4].kind).toBe(DirectionalMovementIndexOutput.DirectionalMovementMinusValue);
      expect(meta.outputs[5].kind).toBe(DirectionalMovementIndexOutput.AverageTrueRangeValue);
      expect(meta.outputs[6].kind).toBe(DirectionalMovementIndexOutput.TrueRangeValue);
    });
  });
});
