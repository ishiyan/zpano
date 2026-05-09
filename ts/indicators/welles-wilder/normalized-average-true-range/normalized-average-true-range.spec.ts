import { } from 'jasmine';

import { NormalizedAverageTrueRange } from './normalized-average-true-range';
import { IndicatorIdentifier } from '../../core/indicator-identifier';
import { Shape } from '../../core/outputs/shape/shape';
import { NormalizedAverageTrueRangeOutput } from './output';
import {
    inputHigh,
    inputLow,
    inputClose,
    expectedNatr,
    expectedNatr1,
} from './testdata';

// TA-Lib test data (252 entries), extracted programmatically from NormalizedAverageTrueRangeTest.cs.
// Full expected NATR output (252 entries, length=14). First 14 entries are NaN.
// Full expected NATR output (252 entries, length=1). First entry is NaN.
describe('NormalizedAverageTrueRange', () => {

  describe('constructor', () => {
    it('should create with valid length', () => {
      const natr = new NormalizedAverageTrueRange(14);
      expect(natr.length).toBe(14);
      expect(natr.isPrimed()).toBe(false);
    });

    it('should throw for length 0', () => {
      expect(() => new NormalizedAverageTrueRange(0)).toThrow();
    });

    it('should throw for negative length', () => {
      expect(() => new NormalizedAverageTrueRange(-8)).toThrow();
    });
  });

  describe('isPrimed', () => {
    it('should not be primed during first length updates', () => {
      const natr = new NormalizedAverageTrueRange(5);
      expect(natr.isPrimed()).toBe(false);
      for (let i = 0; i < 5; i++) {
        natr.update(inputClose[i], inputHigh[i], inputLow[i]);
        expect(natr.isPrimed()).toBe(false);
      }
      for (let i = 5; i < 10; i++) {
        natr.update(inputClose[i], inputHigh[i], inputLow[i]);
        expect(natr.isPrimed()).toBe(true);
      }
    });
  });

  describe('update', () => {
    it('should match TA-Lib reference data with length=14', () => {
      const tolerance = 1e-11;
      const natr = new NormalizedAverageTrueRange(14);

      for (let i = 0; i < inputClose.length; i++) {
        const act = natr.update(inputClose[i], inputHigh[i], inputLow[i]);

        if (isNaN(expectedNatr[i])) {
          expect(isNaN(act)).toBe(true, `[${i}] expected NaN`);
        } else {
          expect(isNaN(act)).toBe(false, `[${i}] got NaN, expected ${expectedNatr[i]}`);
          expect(Math.abs(act - expectedNatr[i])).toBeLessThan(tolerance, `[${i}] expected ${expectedNatr[i]}, got ${act}`);
        }
      }
    });

    it('should match TA-Lib reference data with length=1', () => {
      const tolerance = 1e-11;
      const natr = new NormalizedAverageTrueRange(1);

      for (let i = 0; i < inputClose.length; i++) {
        const act = natr.update(inputClose[i], inputHigh[i], inputLow[i]);

        if (isNaN(expectedNatr1[i])) {
          expect(isNaN(act)).toBe(true, `[${i}] expected NaN`);
        } else {
          expect(isNaN(act)).toBe(false, `[${i}] got NaN, expected ${expectedNatr1[i]}`);
          expect(Math.abs(act - expectedNatr1[i])).toBeLessThan(tolerance, `[${i}] expected ${expectedNatr1[i]}, got ${act}`);
        }
      }
    });

    it('should return 0 when close is 0', () => {
      const natr = new NormalizedAverageTrueRange(14);
      // Prime the indicator.
      for (let i = 0; i < 15; i++) {
        natr.update(inputClose[i], inputHigh[i], inputLow[i]);
      }
      const result = natr.update(0, 3.3, 2.2);
      expect(result).toBe(0);
    });
  });

  describe('NaN passthrough', () => {
    it('should return NaN for NaN inputs', () => {
      const natr = new NormalizedAverageTrueRange(14);
      expect(isNaN(natr.update(NaN, 1, 1))).toBe(true);
      expect(isNaN(natr.update(1, NaN, 1))).toBe(true);
      expect(isNaN(natr.update(1, 1, NaN))).toBe(true);
      expect(isNaN(natr.updateSample(NaN))).toBe(true);
    });
  });

  describe('metadata', () => {
    it('should return correct metadata', () => {
      const natr = new NormalizedAverageTrueRange(14);
      const meta = natr.metadata();
      expect(meta.identifier).toBe(IndicatorIdentifier.NormalizedAverageTrueRange);
      expect(meta.mnemonic).toBe('natr');
      expect(meta.description).toBe('Normalized Average True Range');
      expect(meta.outputs.length).toBe(1);
      expect(meta.outputs[0].kind).toBe(NormalizedAverageTrueRangeOutput.NormalizedAverageTrueRangeValue);
      expect(meta.outputs[0].shape).toBe(Shape.Scalar);
      expect(meta.outputs[0].mnemonic).toBe('natr');
      expect(meta.outputs[0].description).toBe('Normalized Average True Range');
    });
  });
});
