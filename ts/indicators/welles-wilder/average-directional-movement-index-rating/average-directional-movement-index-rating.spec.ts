import { } from 'jasmine';

import { AverageDirectionalMovementIndexRating } from './average-directional-movement-index-rating';
import { IndicatorIdentifier } from '../../core/indicator-identifier';
import { Shape } from '../../core/outputs/shape/shape';
import { AverageDirectionalMovementIndexRatingOutput } from './output';
import {
    inputHigh,
    inputLow,
    inputClose,
    expectedAdxr14,
} from './testdata';

// TA-Lib test data (252 entries).
// Expected ADXR14 (length=14), 252 entries. 40 NaN (indices 0-39), values from index 40 onward.
// Computed as ADXR[i] = (ADX[i] + ADX[i-13]) / 2.
describe('AverageDirectionalMovementIndexRating', () => {

  describe('constructor', () => {
    it('should create with valid length', () => {
      const adxr = new AverageDirectionalMovementIndexRating(14);
      expect(adxr.length).toBe(14);
      expect(adxr.isPrimed()).toBe(false);
    });

    it('should throw for length 0', () => {
      expect(() => new AverageDirectionalMovementIndexRating(0)).toThrow();
    });

    it('should throw for negative length', () => {
      expect(() => new AverageDirectionalMovementIndexRating(-8)).toThrow();
    });
  });

  describe('isPrimed', () => {
    it('should prime at index 40 for length=14', () => {
      const adxr = new AverageDirectionalMovementIndexRating(14);
      expect(adxr.isPrimed()).toBe(false);

      for (let i = 0; i < 40; i++) {
        adxr.update(inputClose[i], inputHigh[i], inputLow[i]);
        expect(adxr.isPrimed()).toBe(false);
      }

      adxr.update(inputClose[40], inputHigh[40], inputLow[40]);
      expect(adxr.isPrimed()).toBe(true);
    });
  });

  describe('update', () => {
    it('should match TA-Lib reference data with length=14', () => {
      const tolerance = 1e-8;
      const adxr = new AverageDirectionalMovementIndexRating(14);

      for (let i = 0; i < inputHigh.length; i++) {
        const act = adxr.update(inputClose[i], inputHigh[i], inputLow[i]);

        if (isNaN(expectedAdxr14[i])) {
          expect(isNaN(act)).toBe(true, `[${i}] expected NaN`);
        } else {
          expect(isNaN(act)).toBe(false, `[${i}] got NaN, expected ${expectedAdxr14[i]}`);
          expect(Math.abs(act - expectedAdxr14[i])).toBeLessThan(tolerance, `[${i}] expected ${expectedAdxr14[i]}, got ${act}`);
        }
      }
    });
  });

  describe('NaN passthrough', () => {
    it('should return NaN for NaN inputs', () => {
      const adxr = new AverageDirectionalMovementIndexRating(14);
      expect(isNaN(adxr.update(NaN, 1, 1))).toBe(true);
      expect(isNaN(adxr.update(1, NaN, 1))).toBe(true);
      expect(isNaN(adxr.update(1, 1, NaN))).toBe(true);
      expect(isNaN(adxr.updateSample(NaN))).toBe(true);
    });
  });

  describe('metadata', () => {
    it('should return correct metadata', () => {
      const adxr = new AverageDirectionalMovementIndexRating(14);
      const meta = adxr.metadata();
      expect(meta.identifier).toBe(IndicatorIdentifier.AverageDirectionalMovementIndexRating);
      expect(meta.mnemonic).toBe('adxr');
      expect(meta.description).toBe('Average Directional Movement Index Rating');
      expect(meta.outputs.length).toBe(9);
      expect(meta.outputs[0].kind).toBe(AverageDirectionalMovementIndexRatingOutput.AverageDirectionalMovementIndexRatingValue);
      expect(meta.outputs[0].shape).toBe(Shape.Scalar);
      expect(meta.outputs[0].mnemonic).toBe('adxr');
      expect(meta.outputs[0].description).toBe('Average Directional Movement Index Rating');
      expect(meta.outputs[1].kind).toBe(AverageDirectionalMovementIndexRatingOutput.AverageDirectionalMovementIndexValue);
      expect(meta.outputs[2].kind).toBe(AverageDirectionalMovementIndexRatingOutput.DirectionalMovementIndexValue);
      expect(meta.outputs[3].kind).toBe(AverageDirectionalMovementIndexRatingOutput.DirectionalIndicatorPlusValue);
      expect(meta.outputs[4].kind).toBe(AverageDirectionalMovementIndexRatingOutput.DirectionalIndicatorMinusValue);
      expect(meta.outputs[5].kind).toBe(AverageDirectionalMovementIndexRatingOutput.DirectionalMovementPlusValue);
      expect(meta.outputs[6].kind).toBe(AverageDirectionalMovementIndexRatingOutput.DirectionalMovementMinusValue);
      expect(meta.outputs[7].kind).toBe(AverageDirectionalMovementIndexRatingOutput.AverageTrueRangeValue);
      expect(meta.outputs[8].kind).toBe(AverageDirectionalMovementIndexRatingOutput.TrueRangeValue);
    });
  });
});
