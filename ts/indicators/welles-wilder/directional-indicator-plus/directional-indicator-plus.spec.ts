import { } from 'jasmine';

import { DirectionalIndicatorPlus } from './directional-indicator-plus';
import { IndicatorIdentifier } from '../../core/indicator-identifier';
import { Shape } from '../../core/outputs/shape/shape';
import { DirectionalIndicatorPlusOutput } from './output';
import {
    inputHigh,
    inputLow,
    inputClose,
    expectedDi14,
} from './testdata';

// TA-Lib test data (252 entries), extracted programmatically from DirectionalIndicatorPlusTest.cs.
// Expected +DI14 (length=14), 252 entries. First 14 entries are NaN.
describe('DirectionalIndicatorPlus', () => {

  describe('constructor', () => {
    it('should create with valid length', () => {
      const dip = new DirectionalIndicatorPlus(14);
      expect(dip.length).toBe(14);
      expect(dip.isPrimed()).toBe(false);
    });

    it('should throw for length 0', () => {
      expect(() => new DirectionalIndicatorPlus(0)).toThrow();
    });

    it('should throw for negative length', () => {
      expect(() => new DirectionalIndicatorPlus(-8)).toThrow();
    });
  });

  describe('isPrimed', () => {
    it('should not be primed during first length updates for length=14', () => {
      const dip = new DirectionalIndicatorPlus(14);
      expect(dip.isPrimed()).toBe(false);

      for (let i = 0; i < 14; i++) {
        dip.update(inputClose[i], inputHigh[i], inputLow[i]);
        expect(dip.isPrimed()).toBe(false);
      }

      dip.update(inputClose[14], inputHigh[14], inputLow[14]);
      expect(dip.isPrimed()).toBe(true);
    });
  });

  describe('update', () => {
    it('should match TA-Lib reference data with length=14', () => {
      const tolerance = 1e-8;
      const dip = new DirectionalIndicatorPlus(14);

      for (let i = 0; i < inputHigh.length; i++) {
        const act = dip.update(inputClose[i], inputHigh[i], inputLow[i]);

        if (isNaN(expectedDi14[i])) {
          expect(isNaN(act)).toBe(true, `[${i}] expected NaN`);
        } else {
          expect(isNaN(act)).toBe(false, `[${i}] got NaN, expected ${expectedDi14[i]}`);
          expect(Math.abs(act - expectedDi14[i])).toBeLessThan(tolerance, `[${i}] expected ${expectedDi14[i]}, got ${act}`);
        }
      }
    });
  });

  describe('NaN passthrough', () => {
    it('should return NaN for NaN inputs', () => {
      const dip = new DirectionalIndicatorPlus(14);
      expect(isNaN(dip.update(NaN, 1, 1))).toBe(true);
      expect(isNaN(dip.update(1, NaN, 1))).toBe(true);
      expect(isNaN(dip.update(1, 1, NaN))).toBe(true);
      expect(isNaN(dip.updateSample(NaN))).toBe(true);
    });
  });

  describe('metadata', () => {
    it('should return correct metadata', () => {
      const dip = new DirectionalIndicatorPlus(14);
      const meta = dip.metadata();
      expect(meta.identifier).toBe(IndicatorIdentifier.DirectionalIndicatorPlus);
      expect(meta.mnemonic).toBe('+di');
      expect(meta.description).toBe('Directional Indicator Plus');
      expect(meta.outputs.length).toBe(4);
      expect(meta.outputs[0].kind).toBe(DirectionalIndicatorPlusOutput.DirectionalIndicatorPlusValue);
      expect(meta.outputs[0].shape).toBe(Shape.Scalar);
      expect(meta.outputs[0].mnemonic).toBe('+di');
      expect(meta.outputs[0].description).toBe('Directional Indicator Plus');
      expect(meta.outputs[1].kind).toBe(DirectionalIndicatorPlusOutput.DirectionalMovementPlusValue);
      expect(meta.outputs[2].kind).toBe(DirectionalIndicatorPlusOutput.AverageTrueRangeValue);
      expect(meta.outputs[3].kind).toBe(DirectionalIndicatorPlusOutput.TrueRangeValue);
    });
  });
});
