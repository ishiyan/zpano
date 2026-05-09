import { } from 'jasmine';

import { AverageDirectionalMovementIndex } from './average-directional-movement-index';
import { IndicatorIdentifier } from '../../core/indicator-identifier';
import { Shape } from '../../core/outputs/shape/shape';
import { AverageDirectionalMovementIndexOutput } from './output';
import {
    inputHigh,
    inputLow,
    inputClose,
    expectedAdx14,
} from './testdata';

// TA-Lib test data (252 entries).
// Expected ADX14 (length=14), 252 entries. 27 NaN (indices 0-26), values from index 27 onward.
describe('AverageDirectionalMovementIndex', () => {

  describe('constructor', () => {
    it('should create with valid length', () => {
      const adx = new AverageDirectionalMovementIndex(14);
      expect(adx.length).toBe(14);
      expect(adx.isPrimed()).toBe(false);
    });

    it('should throw for length 0', () => {
      expect(() => new AverageDirectionalMovementIndex(0)).toThrow();
    });

    it('should throw for negative length', () => {
      expect(() => new AverageDirectionalMovementIndex(-8)).toThrow();
    });
  });

  describe('isPrimed', () => {
    it('should prime at index 27 for length=14', () => {
      const adx = new AverageDirectionalMovementIndex(14);
      expect(adx.isPrimed()).toBe(false);

      for (let i = 0; i < 27; i++) {
        adx.update(inputClose[i], inputHigh[i], inputLow[i]);
        expect(adx.isPrimed()).toBe(false);
      }

      adx.update(inputClose[27], inputHigh[27], inputLow[27]);
      expect(adx.isPrimed()).toBe(true);
    });
  });

  describe('update', () => {
    it('should match TA-Lib reference data with length=14', () => {
      const tolerance = 1e-8;
      const adx = new AverageDirectionalMovementIndex(14);

      for (let i = 0; i < inputHigh.length; i++) {
        const act = adx.update(inputClose[i], inputHigh[i], inputLow[i]);

        if (isNaN(expectedAdx14[i])) {
          expect(isNaN(act)).toBe(true, `[${i}] expected NaN`);
        } else {
          expect(isNaN(act)).toBe(false, `[${i}] got NaN, expected ${expectedAdx14[i]}`);
          expect(Math.abs(act - expectedAdx14[i])).toBeLessThan(tolerance, `[${i}] expected ${expectedAdx14[i]}, got ${act}`);
        }
      }
    });
  });

  describe('NaN passthrough', () => {
    it('should return NaN for NaN inputs', () => {
      const adx = new AverageDirectionalMovementIndex(14);
      expect(isNaN(adx.update(NaN, 1, 1))).toBe(true);
      expect(isNaN(adx.update(1, NaN, 1))).toBe(true);
      expect(isNaN(adx.update(1, 1, NaN))).toBe(true);
      expect(isNaN(adx.updateSample(NaN))).toBe(true);
    });
  });

  describe('metadata', () => {
    it('should return correct metadata', () => {
      const adx = new AverageDirectionalMovementIndex(14);
      const meta = adx.metadata();
      expect(meta.identifier).toBe(IndicatorIdentifier.AverageDirectionalMovementIndex);
      expect(meta.mnemonic).toBe('adx');
      expect(meta.description).toBe('Average Directional Movement Index');
      expect(meta.outputs.length).toBe(8);
      expect(meta.outputs[0].kind).toBe(AverageDirectionalMovementIndexOutput.AverageDirectionalMovementIndexValue);
      expect(meta.outputs[0].shape).toBe(Shape.Scalar);
      expect(meta.outputs[0].mnemonic).toBe('adx');
      expect(meta.outputs[0].description).toBe('Average Directional Movement Index');
      expect(meta.outputs[1].kind).toBe(AverageDirectionalMovementIndexOutput.DirectionalMovementIndexValue);
      expect(meta.outputs[2].kind).toBe(AverageDirectionalMovementIndexOutput.DirectionalIndicatorPlusValue);
      expect(meta.outputs[3].kind).toBe(AverageDirectionalMovementIndexOutput.DirectionalIndicatorMinusValue);
      expect(meta.outputs[4].kind).toBe(AverageDirectionalMovementIndexOutput.DirectionalMovementPlusValue);
      expect(meta.outputs[5].kind).toBe(AverageDirectionalMovementIndexOutput.DirectionalMovementMinusValue);
      expect(meta.outputs[6].kind).toBe(AverageDirectionalMovementIndexOutput.AverageTrueRangeValue);
      expect(meta.outputs[7].kind).toBe(AverageDirectionalMovementIndexOutput.TrueRangeValue);
    });
  });
});
