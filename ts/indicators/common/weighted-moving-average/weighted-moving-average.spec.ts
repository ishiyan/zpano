import { } from 'jasmine';

import { WeightedMovingAverage } from './weighted-moving-average';
import { input } from './testdata';

// ng test mb  --code-coverage --include='**/indicators/**/*.spec.ts'
// ng test mb  --code-coverage --include='**/indicators/*.spec.ts'

/* eslint-disable max-len */
// Input data is taken from the TA-Lib (http://ta-lib.org/) tests,
//    test_data.c, TA_SREF_close_daily_ref_0_PRIV[252].
//
// Output data is taken from TA-Lib (http://ta-lib.org/) tests,
//    test_ma.c.
//
// /*******************************/
// /*   WMA TEST  - CLASSIC       */
// /*******************************/
// #ifndef TA_FUNC_NO_RANGE_CHECK
// /* No output value. */
// { 0, TA_ANY_MA_TEST, 0, 0, 251,  0, TA_MAType_WMA, TA_COMPATIBILITY_DEFAULT, TA_BAD_PARAM, 0, 0, 0, 0 },
// #endif
// /* One value tests. */
// { 0, TA_ANY_MA_TEST, 0, 2,   2,  2, TA_MAType_WMA, TA_COMPATIBILITY_DEFAULT, TA_SUCCESS,   0,  94.52,   2, 1 },
// /* Misc tests: period 2, 30 */
// { 1, TA_ANY_MA_TEST, 0, 0, 251,  2, TA_MAType_WMA, TA_COMPATIBILITY_DEFAULT, TA_SUCCESS,   0,   93.71,  1,  252-1  }, /* First Value */
// { 0, TA_ANY_MA_TEST, 0, 0, 251,  2, TA_MAType_WMA, TA_COMPATIBILITY_DEFAULT, TA_SUCCESS,   1,   94.52,  1,  252-1  },
// { 0, TA_ANY_MA_TEST, 0, 0, 251,  2, TA_MAType_WMA, TA_COMPATIBILITY_DEFAULT, TA_SUCCESS,   2,   94.85,  1,  252-1  },
// { 0, TA_ANY_MA_TEST, 0, 0, 251,  2, TA_MAType_WMA, TA_COMPATIBILITY_DEFAULT, TA_SUCCESS, 250,  108.16,  1,  252-1  }, /* Last Value */
//
// { 1, TA_ANY_MA_TEST, 0, 0, 251, 30, TA_MAType_WMA, TA_COMPATIBILITY_DEFAULT, TA_SUCCESS,   0,  88.567,  29,  252-29 }, /* First Value */
// { 0, TA_ANY_MA_TEST, 0, 0, 251, 30, TA_MAType_WMA, TA_COMPATIBILITY_DEFAULT, TA_SUCCESS,   1,  88.233,  29,  252-29 },
// { 0, TA_ANY_MA_TEST, 0, 0, 251, 30, TA_MAType_WMA, TA_COMPATIBILITY_DEFAULT, TA_SUCCESS,   2,  88.034,  29,  252-29 },
// { 0, TA_ANY_MA_TEST, 0, 0, 251, 30, TA_MAType_WMA, TA_COMPATIBILITY_DEFAULT, TA_SUCCESS,  29,  87.191,  29,  252-29 },
// { 0, TA_ANY_MA_TEST, 0, 0, 251, 30, TA_MAType_WMA, TA_COMPATIBILITY_DEFAULT, TA_SUCCESS, 221, 109.3413, 29,  252-29 },
// { 0, TA_ANY_MA_TEST, 0, 0, 251, 30, TA_MAType_WMA, TA_COMPATIBILITY_DEFAULT, TA_SUCCESS, 222, 109.3466, 29,  252-29 }, /* Last Value */

describe('WeightedMovingAverage', () => {
  const epsilon = 10e-2;

  it('should return expected mnemonic', () => {
    const wma = new WeightedMovingAverage({length: 7});
    expect(wma.metadata().mnemonic).toBe('wma(7)');
  });

  it('should throw if length is less than 2', () => {
    expect(() => { new WeightedMovingAverage({length: 1}); }).toThrow();
  });

  it('should calculate expected output and prime state for length 2', () => {
    const len = 2;
    const wma = new WeightedMovingAverage({length: len});

    for (let i = 0; i < len - 1; i++) {
      expect(wma.update(input[i])).toBeNaN();
      expect(wma.isPrimed()).toBe(false);
    }

    for (let i = len - 1; i < input.length; i++) {
      if (i === 1) {
        expect(wma.update(input[i])).toBeCloseTo(93.71, epsilon);
      } else if (i === 2) {
        expect(wma.update(input[i])).toBeCloseTo(94.52, epsilon);
      } else if (i === 3) {
        expect(wma.update(input[i])).toBeCloseTo(94.855, epsilon);
      } else if (i === 251) {
        expect(wma.update(input[i])).toBeCloseTo(108.16, epsilon);
      } else {
        wma.update(input[i]);
      }

      expect(wma.isPrimed()).toBe(true);
    }

    expect(wma.update(Number.NaN)).toBeNaN();
  });

  it('should calculate expected output and prime state for length 30', () => {
    const len = 30;
    const wma = new WeightedMovingAverage({length: len});

    for (let i = 0; i < len - 1; i++) {
      expect(wma.update(input[i])).toBeNaN();
      expect(wma.isPrimed()).toBe(false);
    }

    for (let i = len - 1; i < input.length; i++) {
      if (i === 29) {
        expect(wma.update(input[i])).toBeCloseTo(88.5677, epsilon);
      } else if (i === 30) {
        expect(wma.update(input[i])).toBeCloseTo(88.2337, epsilon);
      } else if (i === 31) {
        expect(wma.update(input[i])).toBeCloseTo(88.034, epsilon);
      } else if (i === 58) {
        expect(wma.update(input[i])).toBeCloseTo(87.191, epsilon);
      } else if (i === 250) {
        expect(wma.update(input[i])).toBeCloseTo(109.3466, epsilon);
      } else if (i === 251) {
        expect(wma.update(input[i])).toBeCloseTo(109.3413, epsilon);
      } else {
        wma.update(input[i]);
      }

      expect(wma.isPrimed()).toBe(true);
    }

    expect(wma.update(Number.NaN)).toBeNaN();
  });
});
