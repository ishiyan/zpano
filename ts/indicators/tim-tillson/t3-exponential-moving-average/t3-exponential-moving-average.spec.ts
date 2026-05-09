import { } from 'jasmine';

import { T3ExponentialMovingAverage } from './t3-exponential-moving-average';
import { IndicatorIdentifier } from '../../core/indicator-identifier';
import { Shape } from '../../core/outputs/shape/shape';
import { T3ExponentialMovingAverageOutput } from './output';
import { input, expected5sma } from './testdata';

/* eslint-disable max-len */
// Input data is taken from the TA-Lib (http://ta-lib.org/) tests,
//    test_data.c, TA_SREF_close_daily_ref_0_PRIV[252].
//
// Output data is taken from TA-Lib (http://ta-lib.org/) tests,
//    test_ma.c.
//
// /************/
// /*  T3 TEST */
// /************/
// { 1, TA_ANY_MA_TEST, 0, 0, 251, 5, TA_MAType_T3, TA_COMPATIBILITY_DEFAULT, TA_SUCCESS,      0,  85.73, 24,  252-24  }, /* First Value */
// { 0, TA_ANY_MA_TEST, 0, 0, 251, 5, TA_MAType_T3, TA_COMPATIBILITY_DEFAULT, TA_SUCCESS,      1,  84.37, 24,  252-24  },
// { 0, TA_ANY_MA_TEST, 0, 0, 251, 5, TA_MAType_T3, TA_COMPATIBILITY_DEFAULT, TA_SUCCESS, 252-26, 109.03, 24,  252-24  },
// { 0, TA_ANY_MA_TEST, 0, 0, 251, 5, TA_MAType_T3, TA_COMPATIBILITY_DEFAULT, TA_SUCCESS, 252-25, 108.88, 24,  252-24  }, /* Last Value */

/** Taken from TA-Lib (http://ta-lib.org/) tests, test_T3.xls, T3(5,0.7) column.
 * Length is 5, volume factor is 0.7, firstIsAverage = true.
 */
describe('T3ExponentialMovingAverage', () => {

  it('should have correct output enum value', () => {
    expect(T3ExponentialMovingAverageOutput.T3ExponentialMovingAverageValue).toBe(0);
  });

  it('should return expected mnemonic for length-based', () => {
    let t3 = new T3ExponentialMovingAverage({length: 7, volumeFactor: 0.6781, firstIsAverage: true});
    expect(t3.metadata().mnemonic).toBe('t3(7, 0.67810000)');
    t3 = new T3ExponentialMovingAverage({length: 7, volumeFactor: 0.6789, firstIsAverage: false});
    expect(t3.metadata().mnemonic).toBe('t3(7, 0.67890000)');
  });

  it('should return expected mnemonic for smoothing-factor-based', () => {
    // alpha = 0.12345, length = round(2/0.12345) - 1 = round(16.2) - 1 = 15
    const t3 = new T3ExponentialMovingAverage({smoothingFactor: 0.12345, volumeFactor: 0.56789, firstIsAverage: false});
    expect(t3.metadata().mnemonic).toBe('t3(15, 0.12345000, 0.56789000)');
  });

  it('should return expected metadata', () => {
    const t3 = new T3ExponentialMovingAverage({length: 10, volumeFactor: 0.3333, firstIsAverage: true});
    const meta = t3.metadata();

    expect(meta.identifier).toBe(IndicatorIdentifier.T3ExponentialMovingAverage);
    expect(meta.mnemonic).toBe('t3(10, 0.33330000)');
    expect(meta.description).toBe('T3 exponential moving average t3(10, 0.33330000)');
    expect(meta.outputs.length).toBe(1);
    expect(meta.outputs[0].kind).toBe(T3ExponentialMovingAverageOutput.T3ExponentialMovingAverageValue);
    expect(meta.outputs[0].shape).toBe(Shape.Scalar);
    expect(meta.outputs[0].mnemonic).toBe('t3(10, 0.33330000)');
    expect(meta.outputs[0].description).toBe('T3 exponential moving average t3(10, 0.33330000)');
  });

  it('should return expected metadata for smoothing-factor-based', () => {
    // alpha = 2 / (10 + 1) = 2/11 = 0.18181818...
    const alpha = 2 / 11;
    const t3 = new T3ExponentialMovingAverage({smoothingFactor: alpha, volumeFactor: 0.3333333, firstIsAverage: false});
    const meta = t3.metadata();

    expect(meta.identifier).toBe(IndicatorIdentifier.T3ExponentialMovingAverage);
    expect(meta.mnemonic).toBe('t3(10, 0.18181818, 0.33333330)');
    expect(meta.description).toBe('T3 exponential moving average t3(10, 0.18181818, 0.33333330)');
    expect(meta.outputs.length).toBe(1);
    expect(meta.outputs[0].kind).toBe(T3ExponentialMovingAverageOutput.T3ExponentialMovingAverageValue);
    expect(meta.outputs[0].shape).toBe(Shape.Scalar);
    expect(meta.outputs[0].mnemonic).toBe('t3(10, 0.18181818, 0.33333330)');
    expect(meta.outputs[0].description).toBe('T3 exponential moving average t3(10, 0.18181818, 0.33333330)');
  });

  it('should throw if length is less than 2', () => {
    expect(() => { new T3ExponentialMovingAverage({length: 1, volumeFactor: 0.7, firstIsAverage: true}); }).toThrow();
  });

  it('should throw if smoothing factor is less than 0', () => {
    expect(() => { new T3ExponentialMovingAverage({smoothingFactor: -0.1, volumeFactor: 0.7, firstIsAverage: false}); }).toThrow();
  });

  it('should throw if smoothing factor is greater than 1', () => {
    expect(() => { new T3ExponentialMovingAverage({smoothingFactor: 1.1, volumeFactor: 0.7, firstIsAverage: false}); }).toThrow();
  });

  it('should throw if volume factor is less than 0', () => {
    expect(() => { new T3ExponentialMovingAverage({length: 5, volumeFactor: -0.1, firstIsAverage: true}); }).toThrow();
  });

  it('should throw if volume factor is greater than 1', () => {
    expect(() => { new T3ExponentialMovingAverage({length: 5, volumeFactor: 1.1, firstIsAverage: true}); }).toThrow();
  });

  it('should calculate expected output and prime state for length 5, first is NOT SMA', () => {
    const len = 5;
    const lenPrimed = 6*(len - 1);
    const epsilon = 1e-3;
    const t3 = new T3ExponentialMovingAverage({length: len, volumeFactor: 0.7, firstIsAverage: false});

    for (let i = 0; i < lenPrimed; i++) {
      expect(t3.update(input[i])).toBeNaN();
      expect(t3.isPrimed()).toBe(false);
    }

    for (let i = lenPrimed; i < input.length; i++) {
      const act = t3.update(input[i]);
      expect(t3.isPrimed()).toBe(true);

      if (i === 24) {
        expect(act).toBeCloseTo(85.749, epsilon);
      } else if (i === 25) {
        expect(act).toBeCloseTo(84.380, epsilon);
      } else if (i === 250) {
        expect(act).toBeCloseTo(109.03, epsilon);
      } else if (i === 251) {
        expect(act).toBeCloseTo(108.88, epsilon);
      }
    }

    expect(t3.update(Number.NaN)).toBeNaN();
  });

  it('should calculate expected output and prime state for length 5, first is SMA', () => {
    const len = 5;
    const lenPrimed = 6*(len - 1);
    const epsilon = 1e-3;
    const t3 = new T3ExponentialMovingAverage({length: len, volumeFactor: 0.7, firstIsAverage: true});

    for (let i = 0; i < lenPrimed; i++) {
      expect(t3.update(input[i])).toBeNaN();
      expect(t3.isPrimed()).toBe(false);
    }

    for (let i = lenPrimed; i < input.length; i++) {
      const act = t3.update(input[i]);
      expect(t3.isPrimed()).toBe(true);

      if (i === 250) {
        expect(act).toBeCloseTo(109.03, epsilon);
      } else if (i === 251) {
        expect(act).toBeCloseTo(108.88, epsilon);
      }
    }

    expect(t3.update(Number.NaN)).toBeNaN();
  });

  it('should match expected output (Excel) for length 5, first is SMA', () => {
    const eps = 1e-13;
    const len = 5;
    const lenPrimed = 6*(len - 1);
    const t3 = new T3ExponentialMovingAverage({length: len, volumeFactor: 0.7, firstIsAverage: true});

    for (let i = 0; i < lenPrimed; i++) {
      expect(t3.update(input[i])).toBeNaN();
      expect(t3.isPrimed()).toBe(false);
    }

    for (let i = lenPrimed; i < input.length; i++) {
      const act = t3.update(input[i]);
      expect(t3.isPrimed()).toBe(true);
      expect(act).toBeCloseTo(expected5sma[i], eps);
    }

    expect(t3.update(Number.NaN)).toBeNaN();
  });
});
