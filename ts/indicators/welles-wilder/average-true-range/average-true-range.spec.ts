import { } from 'jasmine';

import { AverageTrueRange } from './average-true-range';
import { IndicatorIdentifier } from '../../core/indicator-identifier';
import { Shape } from '../../core/outputs/shape/shape';
import { AverageTrueRangeOutput } from './output';
import {
    inputHigh,
    inputLow,
    inputClose,
    expectedAtr,
} from './testdata';

// TA-Lib test data (252 entries), extracted programmatically from AverageTrueRangeTest.cs.
// Full expected ATR output (252 entries, length=14). First 14 entries are NaN.
// Extracted programmatically from AverageTrueRangeTest.cs.
describe('AverageTrueRange', () => {

  describe('constructor', () => {
    it('should create with valid length', () => {
      const atr = new AverageTrueRange(14);
      expect(atr.length).toBe(14);
      expect(atr.isPrimed()).toBe(false);
    });

    it('should throw for length 0', () => {
      expect(() => new AverageTrueRange(0)).toThrow();
    });

    it('should throw for negative length', () => {
      expect(() => new AverageTrueRange(-8)).toThrow();
    });
  });

  describe('isPrimed', () => {
    it('should not be primed during first length updates', () => {
      const atr = new AverageTrueRange(5);
      expect(atr.isPrimed()).toBe(false);
      for (let i = 0; i < 5; i++) {
        atr.update(inputClose[i], inputHigh[i], inputLow[i]);
        expect(atr.isPrimed()).toBe(false);
      }
      for (let i = 5; i < 10; i++) {
        atr.update(inputClose[i], inputHigh[i], inputLow[i]);
        expect(atr.isPrimed()).toBe(true);
      }
    });
  });

  describe('update', () => {
    it('should match TA-Lib reference data with length=14', () => {
      const tolerance = 1e-12;
      const atr = new AverageTrueRange(14);

      for (let i = 0; i < inputClose.length; i++) {
        const act = atr.update(inputClose[i], inputHigh[i], inputLow[i]);

        if (isNaN(expectedAtr[i])) {
          expect(isNaN(act)).toBe(true, `[${i}] expected NaN`);
        } else {
          expect(isNaN(act)).toBe(false, `[${i}] got NaN, expected ${expectedAtr[i]}`);
          expect(Math.abs(act - expectedAtr[i])).toBeLessThan(tolerance, `[${i}] expected ${expectedAtr[i]}, got ${act}`);
        }
      }
    });

    it('should return TR values with length=1', () => {
      const tolerance = 1e-3;
      const atr = new AverageTrueRange(1);

      const expectedTR = [
        NaN, 3.535, 2.125, 2.69, 3.185, 1.22, 3.0, 3.97, 3.31, 2.435,
        3.78, 3.5, 3.095, 9.685, 4.565, 2.31, 4.5, 1.875, 2.72, 2.5,
        2.845, 1.97, 3.625, 3.22, 2.875, 3.875, 3.19, 5.34, 3.655, 3.155,
        2.75, 2.155, 1.875, 3.44, 2.125, 3.28, 2.315, 3.565, 2.31, 2.03,
        1.94, 5.125, 3.97, 1.47, 3.16, 1.315, 2.22, 2.72, 2.59, 1.655,
        1.5, 2.56, 5.0, 1.935, 1.815, 2.56, 1.875, 2.66, 3.185, 2.185,
        2.5, 1.5, 3.47, 2.28, 4.285, 2.875, 2.03, 2.625, 2.03, 3.125,
        3.0, 3.81, 4.125, 3.375, 3.375, 13.435, 6.81, 5.5, 3.375, 4.25,
        2.78, 3.78, 2.97, 2.25, 2.595, 3.0, 4.125, 2.41, 2.19, 6.47,
        10.25, 5.0, 3.815, 1.815, 2.845, 1.94, 2.095, 4.47, 2.5, 7.72,
      ];

      for (let i = 0; i < expectedTR.length; i++) {
        const act = atr.update(inputClose[i], inputHigh[i], inputLow[i]);

        if (isNaN(expectedTR[i])) {
          expect(isNaN(act)).toBe(true, `[${i}] expected NaN`);
        } else {
          expect(Math.abs(act - expectedTR[i])).toBeLessThan(tolerance, `[${i}] expected ${expectedTR[i]}, got ${act}`);
        }
      }
    });
  });

  describe('NaN passthrough', () => {
    it('should return NaN for NaN inputs', () => {
      const atr = new AverageTrueRange(14);
      expect(isNaN(atr.update(NaN, 1, 1))).toBe(true);
      expect(isNaN(atr.update(1, NaN, 1))).toBe(true);
      expect(isNaN(atr.update(1, 1, NaN))).toBe(true);
      expect(isNaN(atr.updateSample(NaN))).toBe(true);
    });
  });

  describe('metadata', () => {
    it('should return correct metadata', () => {
      const atr = new AverageTrueRange(14);
      const meta = atr.metadata();
      expect(meta.identifier).toBe(IndicatorIdentifier.AverageTrueRange);
      expect(meta.mnemonic).toBe('atr');
      expect(meta.description).toBe('Average True Range');
      expect(meta.outputs.length).toBe(1);
      expect(meta.outputs[0].kind).toBe(AverageTrueRangeOutput.AverageTrueRangeValue);
      expect(meta.outputs[0].shape).toBe(Shape.Scalar);
      expect(meta.outputs[0].mnemonic).toBe('atr');
      expect(meta.outputs[0].description).toBe('Average True Range');
    });
  });
});
