import { } from 'jasmine';

import { DoubleExponentialMovingAverage } from './double-exponential-moving-average';
import { IndicatorIdentifier } from '../../core/indicator-identifier';
import { Shape } from '../../core/outputs/shape/shape';
import { DoubleExponentialMovingAverageOutput } from './output';
import { input, inputTasc, expectedTasc } from './testdata';

/* eslint-disable max-len */
// Input data is taken from the TA-Lib (http://ta-lib.org/) tests,
//    test_data.c, TA_SREF_close_daily_ref_0_PRIV[252].
//
// Output data is taken from TA-Lib (http://ta-lib.org/) tests,
//    test_ma.c.
//
//   /*******************************/
//   /*  DEMA TEST - Metastock      */
//   /*******************************/
//
//   /* No output value. */
//   { 0, TA_ANY_MA_TEST, 0, 1, 1,  14, TA_MAType_DEMA, TA_COMPATIBILITY_METASTOCK, TA_SUCCESS, 0, 0, 0, 0},
//#ifndef TA_FUNC_NO_RANGE_CHECK
//   { 0, TA_ANY_MA_TEST, 0, 0, 251,  0, TA_MAType_DEMA, TA_COMPATIBILITY_METASTOCK, TA_BAD_PARAM, 0, 0, 0, 0 },
//#endif
//
//   /* Test with period 14 */
//   { 0, TA_ANY_MA_TEST, 0, 0, 251, 14, TA_MAType_DEMA, TA_COMPATIBILITY_METASTOCK, TA_SUCCESS,   0,  83.785, 26, 252-26 }, /* First Value */
//   { 0, TA_ANY_MA_TEST, 0, 0, 251, 14, TA_MAType_DEMA, TA_COMPATIBILITY_METASTOCK, TA_SUCCESS,   1,  84.768, 26, 252-26 },
//   { 0, TA_ANY_MA_TEST, 0, 0, 251, 14, TA_MAType_DEMA, TA_COMPATIBILITY_METASTOCK, TA_SUCCESS, 252-27, 109.467, 26, 252-26 }, /* Last Value */
//
//   /* Test with 1 unstable price bar. Test for period 2, 14 */
//   { 1, TA_ANY_MA_TEST, 1, 0, 251,  2, TA_MAType_DEMA, TA_COMPATIBILITY_METASTOCK, TA_SUCCESS,   0,  93.960, 4, 252-4 }, /* First Value */
//   { 0, TA_ANY_MA_TEST, 1, 0, 251,  2, TA_MAType_DEMA, TA_COMPATIBILITY_METASTOCK, TA_SUCCESS,   1,  94.522, 4, 252-4 },
//   { 0, TA_ANY_MA_TEST, 1, 0, 251,  2, TA_MAType_DEMA, TA_COMPATIBILITY_METASTOCK, TA_SUCCESS, 252-5, 107.94, 4, 252-4 }, /* Last Value */
//
//   { 1, TA_ANY_MA_TEST, 1, 0, 251,  14, TA_MAType_DEMA, TA_COMPATIBILITY_METASTOCK, TA_SUCCESS,    0,  84.91,  (13*2)+2, 252-((13*2)+2) }, /* First Value */
//   { 0, TA_ANY_MA_TEST, 1, 0, 251,  14, TA_MAType_DEMA, TA_COMPATIBILITY_METASTOCK, TA_SUCCESS,    1,  84.97,  (13*2)+2, 252-((13*2)+2) },
//   { 0, TA_ANY_MA_TEST, 1, 0, 251,  14, TA_MAType_DEMA, TA_COMPATIBILITY_METASTOCK, TA_SUCCESS,    2,  84.80,  (13*2)+2, 252-((13*2)+2) },
//   { 0, TA_ANY_MA_TEST, 1, 0, 251,  14, TA_MAType_DEMA, TA_COMPATIBILITY_METASTOCK, TA_SUCCESS,    3,  85.14,  (13*2)+2, 252-((13*2)+2) },
//   { 0, TA_ANY_MA_TEST, 1, 0, 251,  14, TA_MAType_DEMA, TA_COMPATIBILITY_METASTOCK, TA_SUCCESS,   20,  89.83,  (13*2)+2, 252-((13*2)+2) },
//   { 0, TA_ANY_MA_TEST, 1, 0, 251,  14, TA_MAType_DEMA, TA_COMPATIBILITY_METASTOCK, TA_SUCCESS, 252-((13*2)+2+1), 109.4676, (13*2)+2, 252-((13*2)+2) }, /* Last Value */

// Input and output data taken from Excel file describing TEMA calculations in
// Technical Analysis of Stocks & Commodities v.12:2 (72-80), Smoothing Data With Less Lag.

/* eslint-disable no-loss-of-precision */

describe('DoubleExponentialMovingAverage', () => {

  it('should have correct output enum value', () => {
    expect(DoubleExponentialMovingAverageOutput.DoubleExponentialMovingAverageValue).toBe(0);
  });

  it('should return expected mnemonic for length-based', () => {
    let dema = new DoubleExponentialMovingAverage({length: 7, firstIsAverage: true});
    expect(dema.metadata().mnemonic).toBe('dema(7)');
    dema = new DoubleExponentialMovingAverage({length: 7, firstIsAverage: false});
    expect(dema.metadata().mnemonic).toBe('dema(7)');
  });

  it('should return expected mnemonic for smoothing-factor-based', () => {
    // a = 0.12345, length = round(2/0.12345) - 1 = round(16.2) - 1 = 15
    const dema = new DoubleExponentialMovingAverage({smoothingFactor: 0.12345});
    expect(dema.metadata().mnemonic).toBe('dema(15, 0.12345000)');
  });

  it('should return expected metadata', () => {
    const dema = new DoubleExponentialMovingAverage({length: 14, firstIsAverage: true});
    const meta = dema.metadata();

    expect(meta.identifier).toBe(IndicatorIdentifier.DoubleExponentialMovingAverage);
    expect(meta.mnemonic).toBe('dema(14)');
    expect(meta.description).toBe('Double exponential moving average dema(14)');
    expect(meta.outputs.length).toBe(1);
    expect(meta.outputs[0].kind).toBe(DoubleExponentialMovingAverageOutput.DoubleExponentialMovingAverageValue);
    expect(meta.outputs[0].shape).toBe(Shape.Scalar);
    expect(meta.outputs[0].mnemonic).toBe('dema(14)');
    expect(meta.outputs[0].description).toBe('Double exponential moving average dema(14)');
  });

  it('should throw if length is less than 2', () => {
    expect(() => { new DoubleExponentialMovingAverage({length: 1, firstIsAverage: true}); }).toThrow();
  });

  it('should throw if smoothing factor is less or equal to 0', () => {
    expect(() => { new DoubleExponentialMovingAverage({smoothingFactor: 0}); }).toThrow();
  });

  it('should throw if smoothing factor is greater or equal to 1', () => {
    expect(() => { new DoubleExponentialMovingAverage({smoothingFactor: 1}); }).toThrow();
  });

  it('should calculate expected output and prime state for length 2, first is SMA', () => {
    const len = 2;
    const lenPrimed = 2*len - 2;
    const epsilon = 10e-2;
    const dema = new DoubleExponentialMovingAverage({length: len, firstIsAverage: true});

    for (let i = 0; i < lenPrimed; i++) {
      expect(dema.update(input[i])).toBeNaN();
      expect(dema.isPrimed()).toBe(false);
    }

    for (let i = lenPrimed; i < input.length; i++) {
      const act = dema.update(input[i]);
      expect(dema.isPrimed()).toBe(true);

      if (i === 4) {
        expect(act).toBeCloseTo(94.013, epsilon);
      } else if (i === 5) {
        expect(act).toBeCloseTo(94.539, epsilon);
      } else if (i === 251) {
        expect(act).toBeCloseTo(107.94, epsilon);
      }
    }

    expect(dema.update(Number.NaN)).toBeNaN();
  });

  it('should calculate expected output and prime state for length 14, first is SMA', () => {
    const len = 14;
    const lenPrimed = 2*len - 2;
    const epsilon = 10e-2;
    const dema = new DoubleExponentialMovingAverage({length: len, firstIsAverage: true});

    for (let i = 0; i < lenPrimed; i++) {
      expect(dema.update(input[i])).toBeNaN();
      expect(dema.isPrimed()).toBe(false);
    }

    for (let i = lenPrimed; i < input.length; i++) {
      const act = dema.update(input[i]);
      expect(dema.isPrimed()).toBe(true);

      if (i === 28) {
        expect(act).toBeCloseTo(84.347, epsilon);
      } else if (i === 29) {
        expect(act).toBeCloseTo(84.487, epsilon);
      } else if (i === 30) {
        expect(act).toBeCloseTo(84.374, epsilon);
      } else if (i === 21) {
        expect(act).toBeCloseTo(84.772, epsilon);
      } else if (i === 48) {
        expect(act).toBeCloseTo(89.803, epsilon);
      } else if (i === 251) {
        expect(act).toBeCloseTo(109.4676, epsilon);
      }
    }

    expect(dema.update(Number.NaN)).toBeNaN();
  });

  it('should calculate expected output and prime state for length 2, first is NOT SMA', () => {
    const len = 2;
    const lenPrimed = 2*len - 2;
    const epsilon = 10e-2;
    const dema = new DoubleExponentialMovingAverage({length: len, firstIsAverage: false});

    for (let i = 0; i < lenPrimed; i++) {
      expect(dema.update(input[i])).toBeNaN();
      expect(dema.isPrimed()).toBe(false);
    }

    for (let i = lenPrimed; i < input.length; i++) {
      const act = dema.update(input[i]);
      expect(dema.isPrimed()).toBe(true);

      if (i === 4) {
        expect(act).toBeCloseTo(93.977, epsilon);
      } else if (i === 5) {
        expect(act).toBeCloseTo(94.522, epsilon);
      } else if (i === 251) {
        expect(act).toBeCloseTo(107.94, epsilon);
      }
    }

    expect(dema.update(Number.NaN)).toBeNaN();
  });

  it('should calculate expected output and prime state for length 14, first is NOT SMA', () => {
    const len = 14;
    const lenPrimed = 2*len - 2;
    const epsilon = 10e-2;
    const dema = new DoubleExponentialMovingAverage({length: len, firstIsAverage: false});

    for (let i = 0; i < lenPrimed; i++) {
      expect(dema.update(input[i])).toBeNaN();
      expect(dema.isPrimed()).toBe(false);
    }

    for (let i = lenPrimed; i < input.length; i++) {
      const act = dema.update(input[i]);
      expect(dema.isPrimed()).toBe(true);

      if (i === 28) {
        expect(act).toBeCloseTo(84.87, epsilon);
      } else if (i === 29) {
        expect(act).toBeCloseTo(84.94, epsilon);
      } else if (i === 30) {
        expect(act).toBeCloseTo(84.77, epsilon);
      } else if (i === 21) {
        expect(act).toBeCloseTo(85.12, epsilon);
      } else if (i === 48) {
        expect(act).toBeCloseTo(89.83, epsilon);
      } else if (i === 251) {
        expect(act).toBeCloseTo(109.4676, epsilon);
      }
    }

    expect(dema.update(Number.NaN)).toBeNaN();
  });

  it('should match expected output from TASC DEMA calculation Excel, length 26, first is NOT SMA', () => {
    const len = 26;
    const lenPrimed = 2*len - 2;
    const epsilon = 1e-3;
    const firstCheck = 216;
    const dema = new DoubleExponentialMovingAverage({length: len, firstIsAverage: false});

    for (let i = 0; i < lenPrimed; i++) {
      expect(dema.update(inputTasc[i])).toBeNaN();
      expect(dema.isPrimed()).toBe(false);
    }

    for (let i = lenPrimed; i < inputTasc.length; i++) {
      const act = dema.update(inputTasc[i]);
      expect(dema.isPrimed()).toBe(true);

      if (i >= firstCheck) {
        expect(act).toBeCloseTo(expectedTasc[i], epsilon);
      }
    }

    expect(dema.update(Number.NaN)).toBeNaN();
  });
});
