import { } from 'jasmine';

import { ExponentialMovingAverage } from './exponential-moving-average';
import { IndicatorIdentifier } from '../../core/indicator-identifier';
import { Shape } from '../../core/outputs/shape/shape';
import { ExponentialMovingAverageOutput } from './output';
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
// /*   EMA TEST - Classic        */
// /*******************************/
// /* No output value. */
// { 0, TA_ANY_MA_TEST, 0, 1, 1,  14, TA_MAType_EMA, TA_COMPATIBILITY_DEFAULT, TA_SUCCESS, 0, 0, 0, 0},
// #ifndef TA_FUNC_NO_RANGE_CHECK
// { 0, TA_ANY_MA_TEST, 0, 0, 251,  0, TA_MAType_EMA, TA_COMPATIBILITY_DEFAULT, TA_BAD_PARAM, 0, 0, 0, 0 },
// #endif
// /* Misc tests: period 2, 10 */
// { 1, TA_ANY_MA_TEST, 0, 0, 251,  2, TA_MAType_EMA, TA_COMPATIBILITY_DEFAULT, TA_SUCCESS,   0,  93.15, 1, 251 }, /* First Value */
// { 0, TA_ANY_MA_TEST, 0, 0, 251,  2, TA_MAType_EMA, TA_COMPATIBILITY_DEFAULT, TA_SUCCESS,   1,  93.96, 1, 251 },
// { 0, TA_ANY_MA_TEST, 0, 0, 251,  2, TA_MAType_EMA, TA_COMPATIBILITY_DEFAULT, TA_SUCCESS, 250, 108.21, 1, 251 }, /* Last Value */
//
// { 1, TA_ANY_MA_TEST, 0, 0, 251,  10, TA_MAType_EMA, TA_COMPATIBILITY_DEFAULT, TA_SUCCESS,    0,  93.22,  9, 243 }, /* First Value */
// { 0, TA_ANY_MA_TEST, 0, 0, 251,  10, TA_MAType_EMA, TA_COMPATIBILITY_DEFAULT, TA_SUCCESS,    1,  93.75,  9, 243 },
// { 0, TA_ANY_MA_TEST, 0, 0, 251,  10, TA_MAType_EMA, TA_COMPATIBILITY_DEFAULT, TA_SUCCESS,   20,  86.46,  9, 243 },
// { 0, TA_ANY_MA_TEST, 0, 0, 251,  10, TA_MAType_EMA, TA_COMPATIBILITY_DEFAULT, TA_SUCCESS,  242, 108.97,  9, 243 }, /* Last Value */
// /*******************************/
// /*   EMA TEST - Metastock      */
// /*******************************/
// /* No output value. */
// { 0, TA_ANY_MA_TEST, 0, 1, 1,  14, TA_MAType_EMA, TA_COMPATIBILITY_METASTOCK, TA_SUCCESS, 0, 0, 0, 0},
// #ifndef TA_FUNC_NO_RANGE_CHECK
// { 0, TA_ANY_MA_TEST, 0, 0, 251,  0, TA_MAType_EMA, TA_COMPATIBILITY_METASTOCK, TA_BAD_PARAM, 0, 0, 0, 0 },
// #endif
// /* Test with 1 unstable price bar. Test for period 2, 10 */
// { 1, TA_ANY_MA_TEST, 1, 0, 251,  2, TA_MAType_EMA, TA_COMPATIBILITY_METASTOCK, TA_SUCCESS,   0,  94.15, 1+1, 251-1 }, /* First Value */
// { 0, TA_ANY_MA_TEST, 1, 0, 251,  2, TA_MAType_EMA, TA_COMPATIBILITY_METASTOCK, TA_SUCCESS,   1,  94.78, 1+1, 251-1 },
// { 0, TA_ANY_MA_TEST, 1, 0, 251,  2, TA_MAType_EMA, TA_COMPATIBILITY_METASTOCK, TA_SUCCESS, 250-1, 108.21, 1+1, 251-1 }, /* Last Value */
//
// { 1, TA_ANY_MA_TEST, 1, 0, 251,  10, TA_MAType_EMA, TA_COMPATIBILITY_METASTOCK, TA_SUCCESS,    0,  93.24,  9+1, 243-1 }, /* First Value */
// { 0, TA_ANY_MA_TEST, 1, 0, 251,  10, TA_MAType_EMA, TA_COMPATIBILITY_METASTOCK, TA_SUCCESS,    1,  93.97,  9+1, 243-1 },
// { 0, TA_ANY_MA_TEST, 1, 0, 251,  10, TA_MAType_EMA, TA_COMPATIBILITY_METASTOCK, TA_SUCCESS,   20,  86.23,  9+1, 243-1 },
// { 0, TA_ANY_MA_TEST, 1, 0, 251,  10, TA_MAType_EMA, TA_COMPATIBILITY_METASTOCK, TA_SUCCESS, 242-1, 108.97,  9+1, 243-1 }, /* Last Value */
//
// /* Test with 2 unstable price bar. Test for period 2, 10 */
// { 0, TA_ANY_MA_TEST, 2, 0, 251,  2, TA_MAType_EMA, TA_COMPATIBILITY_METASTOCK, TA_SUCCESS,   0,  94.78, 1+2, 251-2 }, /* First Value */
// { 0, TA_ANY_MA_TEST, 2, 0, 251,  2, TA_MAType_EMA, TA_COMPATIBILITY_METASTOCK, TA_SUCCESS,   1,  94.11, 1+2, 251-2 },
// { 0, TA_ANY_MA_TEST, 2, 0, 251,  2, TA_MAType_EMA, TA_COMPATIBILITY_METASTOCK, TA_SUCCESS, 250-2, 108.21, 1+2, 251-2 }, /* Last Value */
//
// { 0, TA_ANY_MA_TEST, 2, 0, 251,  10, TA_MAType_EMA, TA_COMPATIBILITY_METASTOCK, TA_SUCCESS,    0,  93.97,  9+2, 243-2 }, /* First Value */
// { 0, TA_ANY_MA_TEST, 2, 0, 251,  10, TA_MAType_EMA, TA_COMPATIBILITY_METASTOCK, TA_SUCCESS,    1,  94.79,  9+2, 243-2 },
// { 0, TA_ANY_MA_TEST, 2, 0, 251,  10, TA_MAType_EMA, TA_COMPATIBILITY_METASTOCK, TA_SUCCESS,   20,  86.39,  9+2, 243-2 },
// { 0, TA_ANY_MA_TEST, 2, 0, 251,  10, TA_MAType_EMA, TA_COMPATIBILITY_METASTOCK, TA_SUCCESS,  242-2, 108.97,  9+2, 243-2 }, /* Last Value */
//
// /* Last 3 value with 1 unstable, period 10 */
// { 0, TA_ANY_MA_TEST, 1, 249, 251,  10, TA_MAType_EMA, TA_COMPATIBILITY_METASTOCK, TA_SUCCESS,   1, 109.22, 249, 3 },
// { 0, TA_ANY_MA_TEST, 1, 249, 251,  10, TA_MAType_EMA, TA_COMPATIBILITY_METASTOCK, TA_SUCCESS,   2, 108.97, 249, 3 },
//
// /* Last 3 value with 2 unstable, period 10 */
// { 0, TA_ANY_MA_TEST, 2, 249, 251,  10, TA_MAType_EMA, TA_COMPATIBILITY_METASTOCK, TA_SUCCESS,   2, 108.97, 249, 3 },
//
// /* Last 3 value with 3 unstable, period 10 */
// { 0, TA_ANY_MA_TEST, 3, 249, 251,  10, TA_MAType_EMA, TA_COMPATIBILITY_METASTOCK, TA_SUCCESS,   2, 108.97, 249, 3 }

describe('ExponentialMovingAverage', () => {
  const epsilon = 10e-2;

  it('should return expected mnemonic for length-based', () => {
    let ema = new ExponentialMovingAverage({length: 7, firstIsAverage: true});
    expect(ema.metadata().mnemonic).toBe('ema(7)');
    ema = new ExponentialMovingAverage({length: 7, firstIsAverage: false});
    expect(ema.metadata().mnemonic).toBe('ema(7)');
    ema = new ExponentialMovingAverage({length: 7});
    expect(ema.metadata().mnemonic).toBe('ema(7)');
  });

  it('should return expected mnemonic for smoothing-factor-based', () => {
    // α = 2/11 = 0.18181818..., length = round(2/α) - 1 = 10
    const ema = new ExponentialMovingAverage({smoothingFactor: 2 / 11});
    expect(ema.metadata().mnemonic).toBe('ema(10, 0.18181818)');
  });

  it('should return expected metadata', () => {
    const ema = new ExponentialMovingAverage({length: 10, firstIsAverage: true});
    const meta = ema.metadata();

    expect(meta.identifier).toBe(IndicatorIdentifier.ExponentialMovingAverage);
    expect(meta.mnemonic).toBe('ema(10)');
    expect(meta.description).toBe('Exponential moving average ema(10)');
    expect(meta.outputs.length).toBe(1);
    expect(meta.outputs[0].kind).toBe(ExponentialMovingAverageOutput.ExponentialMovingAverageValue);
    expect(meta.outputs[0].shape).toBe(Shape.Scalar);
    expect(meta.outputs[0].mnemonic).toBe('ema(10)');
    expect(meta.outputs[0].description).toBe('Exponential moving average ema(10)');
  });

  it('should throw if length is less than 2', () => {
    expect(() => { new ExponentialMovingAverage({length: 1, firstIsAverage: true}); }).toThrow();
  });

  it('should throw if smoothing factor is less or equal to 0', () => {
    expect(() => { new ExponentialMovingAverage({smoothingFactor: 0}); }).toThrow();
  });

  it('should throw if smoothing factor is greater or equal to 1', () => {
    expect(() => { new ExponentialMovingAverage({smoothingFactor: 1}); }).toThrow();
  });

  it('should calculate expected output and prime state for length 2, first is SMA', () => {
    const len = 2;
    const ema = new ExponentialMovingAverage({length: len, firstIsAverage: true});

    for (let i = 0; i < len - 1; i++) {
      expect(ema.update(input[i])).toBeNaN();
      expect(ema.isPrimed()).toBe(false);
    }

    for (let i = len - 1; i < input.length; i++) {
      const act = ema.update(input[i]);
      expect(ema.isPrimed()).toBe(true);

      if (i === 1) {
        expect(act).toBeCloseTo(93.15, epsilon);
      } else if (i === 2) {
        expect(act).toBeCloseTo(93.96, epsilon);
      } else if (i === 3) {
        expect(act).toBeCloseTo(94.71, epsilon);
      } else if (i === 251) {
        expect(act).toBeCloseTo(108.21, epsilon);
      }
    }

    expect(ema.update(Number.NaN)).toBeNaN();
  });

  it('should calculate expected output and prime state for length 10, first is SMA', () => {
    const len = 10;
    const ema = new ExponentialMovingAverage({length: len, firstIsAverage: true});

    for (let i = 0; i < len - 1; i++) {
      expect(ema.update(input[i])).toBeNaN();
      expect(ema.isPrimed()).toBe(false);
    }

    for (let i = len - 1; i < input.length; i++) {
      const act = ema.update(input[i]);
      expect(ema.isPrimed()).toBe(true);

      if (i === 9) {
        expect(act).toBeCloseTo(93.22, epsilon);
      } else if (i === 10) {
        expect(act).toBeCloseTo(93.75, epsilon);
      } else if (i === 29) {
        expect(act).toBeCloseTo(86.46, epsilon);
      } else if (i === 251) {
        expect(act).toBeCloseTo(108.97, epsilon);
      }
    }

    expect(ema.update(Number.NaN)).toBeNaN();
  });

  it('should calculate expected output and prime state for length 2, first is NOT SMA', () => {
    const len = 2;
    const ema = new ExponentialMovingAverage({length: len, firstIsAverage: false});

    for (let i = 0; i < len - 1; i++) {
      expect(ema.update(input[i])).toBeNaN();
      expect(ema.isPrimed()).toBe(false);
    }

    for (let i = len - 1; i < input.length; i++) {
      const act = ema.update(input[i]);
      expect(ema.isPrimed()).toBe(true);

      if (i === 1) {
        expect(act).toBeCloseTo(93.71, epsilon);
      } else if (i === 2) {
        expect(act).toBeCloseTo(94.15, epsilon);
      } else if (i === 3) {
        expect(act).toBeCloseTo(94.78, epsilon);
      } else if (i === 251) {
        expect(act).toBeCloseTo(108.21, epsilon);
      }
    }

    expect(ema.update(Number.NaN)).toBeNaN();
  });

  it('should calculate expected output and prime state for length 10, first is NOT SMA', () => {
    const len = 10;
    const ema = new ExponentialMovingAverage({length: len, firstIsAverage: false});

    for (let i = 0; i < len - 1; i++) {
      expect(ema.update(input[i])).toBeNaN();
      expect(ema.isPrimed()).toBe(false);
    }

    for (let i = len - 1; i < input.length; i++) {
      const act = ema.update(input[i]);
      expect(ema.isPrimed()).toBe(true);

      if (i === 9) {
        expect(act).toBeCloseTo(92.60, epsilon);
      } else if (i === 10) {
        expect(act).toBeCloseTo(93.24, epsilon);
      } else if (i === 11) {
        expect(act).toBeCloseTo(93.97, epsilon);
      } else if (i === 30) {
        expect(act).toBeCloseTo(86.23, epsilon);
      } else if (i === 251) {
        expect(act).toBeCloseTo(108.97, epsilon);
      }
    }

    expect(ema.update(Number.NaN)).toBeNaN();
  });

  it('should calculate expected output for smoothing-factor-based, length 10', () => {
    // α = 2/(10+1) = 2/11, same as length-based with length=10
    const alpha = 2 / 11;
    const ema = new ExponentialMovingAverage({smoothingFactor: alpha, firstIsAverage: true});

    for (let i = 0; i < 9; i++) {
      expect(ema.update(input[i])).toBeNaN();
      expect(ema.isPrimed()).toBe(false);
    }

    for (let i = 9; i < input.length; i++) {
      const act = ema.update(input[i]);
      expect(ema.isPrimed()).toBe(true);

      if (i === 9) {
        expect(act).toBeCloseTo(93.22, epsilon);
      } else if (i === 10) {
        expect(act).toBeCloseTo(93.75, epsilon);
      } else if (i === 29) {
        expect(act).toBeCloseTo(86.46, epsilon);
      } else if (i === 251) {
        expect(act).toBeCloseTo(108.97, epsilon);
      }
    }

    expect(ema.update(Number.NaN)).toBeNaN();
  });
});
