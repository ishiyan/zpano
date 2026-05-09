import { } from 'jasmine';

import { TripleExponentialMovingAverage } from './triple-exponential-moving-average';
import { IndicatorIdentifier } from '../../core/indicator-identifier';
import { Shape } from '../../core/outputs/shape/shape';
import { TripleExponentialMovingAverageOutput } from './output';
import { input, inputTasc, expectedTasc } from './testdata';

// ng test mb  --code-coverage --include='**/indicators/**/*.spec.ts'
// ng test mb  --code-coverage --include='**/indicators/patrick-mulloy/triple-exponential-moving-average/*.spec.ts'

/* eslint-disable max-len */
// Input data is taken from the TA-Lib (http://ta-lib.org/) tests,
//    test_data.c, TA_SREF_close_daily_ref_0_PRIV[252].
//
// Output data is taken from TA-Lib (http://ta-lib.org/) tests,
//    test_ma.c.
//
//   /*******************************/
//   /*  TEMA TEST - Metastock      */
//   /*******************************/
//   /* No output value. */
//   { 0, TA_ANY_MA_TEST, 0, 1, 1,  14, TA_MAType_TEMA, TA_COMPATIBILITY_METASTOCK, TA_SUCCESS, 0, 0, 0, 0},
//#ifndef TA_FUNC_NO_RANGE_CHECK
//   { 0, TA_ANY_MA_TEST, 0, 0, 251,  0, TA_MAType_TEMA, TA_COMPATIBILITY_METASTOCK, TA_BAD_PARAM, 0, 0, 0, 0 },
//#endif
//
//   /* Test with period 14 */
//   { 1, TA_ANY_MA_TEST, 0, 0, 251, 14, TA_MAType_TEMA, TA_COMPATIBILITY_METASTOCK, TA_SUCCESS,   0,  84.721, 39, 252-39 }, /* First Value */
//   { 0, TA_ANY_MA_TEST, 0, 0, 251, 14, TA_MAType_TEMA, TA_COMPATIBILITY_METASTOCK, TA_SUCCESS,   1,  84.089, 39, 252-39 },
//   { 0, TA_ANY_MA_TEST, 0, 0, 251, 14, TA_MAType_TEMA, TA_COMPATIBILITY_METASTOCK, TA_SUCCESS, 252-40, 108.418, 39, 252-39 }, /* Last Value */

// Input and output data taken from Excel file describing TEMA calculations in
// Technical Analysis of Stocks & Commodities v.12:2 (72-80), Smoothing Data With Less Lag.

/* eslint-disable no-loss-of-precision */

describe('TripleExponentialMovingAverage', () => {

  it('should have correct output enum value', () => {
    expect(TripleExponentialMovingAverageOutput.TripleExponentialMovingAverageValue).toBe(0);
  });

  it('should return expected mnemonic for length-based', () => {
    let tema = new TripleExponentialMovingAverage({length: 7, firstIsAverage: true});
    expect(tema.metadata().mnemonic).toBe('tema(7)');
    tema = new TripleExponentialMovingAverage({length: 7, firstIsAverage: false});
    expect(tema.metadata().mnemonic).toBe('tema(7)');
  });

  it('should return expected mnemonic for smoothing-factor-based', () => {
    // α = 0.12345, length = round(2/0.12345) - 1 = round(16.2) - 1 = 15
    const tema = new TripleExponentialMovingAverage({smoothingFactor: 0.12345});
    expect(tema.metadata().mnemonic).toBe('tema(15, 0.12345000)');
  });

  it('should return expected metadata', () => {
    const tema = new TripleExponentialMovingAverage({length: 14, firstIsAverage: true});
    const meta = tema.metadata();

    expect(meta.identifier).toBe(IndicatorIdentifier.TripleExponentialMovingAverage);
    expect(meta.mnemonic).toBe('tema(14)');
    expect(meta.description).toBe('Triple exponential moving average tema(14)');
    expect(meta.outputs.length).toBe(1);
    expect(meta.outputs[0].kind).toBe(TripleExponentialMovingAverageOutput.TripleExponentialMovingAverageValue);
    expect(meta.outputs[0].shape).toBe(Shape.Scalar);
    expect(meta.outputs[0].mnemonic).toBe('tema(14)');
    expect(meta.outputs[0].description).toBe('Triple exponential moving average tema(14)');
  });

  it('should throw if length is less than 2', () => {
    expect(() => { new TripleExponentialMovingAverage({length: 1, firstIsAverage: true}); }).toThrow();
  });

  it('should throw if smoothing factor is less or equal to 0', () => {
    expect(() => { new TripleExponentialMovingAverage({smoothingFactor: 0}); }).toThrow();
  });

  it('should throw if smoothing factor is greater or equal to 1', () => {
    expect(() => { new TripleExponentialMovingAverage({smoothingFactor: 1}); }).toThrow();
  });

  it('should calculate expected output and prime state for length 14, first is SMA', () => {
    const len = 14;
    const lenPrimed = 3*len - 3;
    const epsilon = 1e-3;
    const tema = new TripleExponentialMovingAverage({length: len, firstIsAverage: true});

    for (let i = 0; i < lenPrimed; i++) {
      expect(tema.update(input[i])).toBeNaN();
      expect(tema.isPrimed()).toBe(false);
    }

    for (let i = lenPrimed; i < input.length; i++) {
      const act = tema.update(input[i]);
      expect(tema.isPrimed()).toBe(true);

      if (i === 39) {
        expect(act).toBeCloseTo(84.8629, epsilon);
      } else if (i === 40) {
        expect(act).toBeCloseTo(84.2246, epsilon);
      } else if (i === 251) {
        expect(act).toBeCloseTo(108.418, epsilon);
      }
    }

    expect(tema.update(Number.NaN)).toBeNaN();
  });

  it('should calculate expected output and prime state for length 14, first is NOT SMA', () => {
    const len = 14;
    const lenPrimed = 3*len - 3;
    const epsilon = 1e-3;
    const tema = new TripleExponentialMovingAverage({length: len, firstIsAverage: false});

    for (let i = 0; i < lenPrimed; i++) {
      expect(tema.update(input[i])).toBeNaN();
      expect(tema.isPrimed()).toBe(false);
    }

    for (let i = lenPrimed; i < input.length; i++) {
      const act = tema.update(input[i]);
      expect(tema.isPrimed()).toBe(true);

      if (i === 39) {
        expect(act).toBeCloseTo(84.721, epsilon);
      } else if (i === 40) {
        expect(act).toBeCloseTo(84.089, epsilon);
      } else if (i === 251) {
        expect(act).toBeCloseTo(108.418, epsilon);
      }
    }

    expect(tema.update(Number.NaN)).toBeNaN();
  });

  it('should match expected output from TASC TEMA calculation Excel, length 26, first is NOT SMA', () => {
    const len = 26;
    const lenPrimed = 3*len - 3;
    const epsilon = 1e-3;
    const firstCheck = 216;
    const tema = new TripleExponentialMovingAverage({length: len, firstIsAverage: false});

    for (let i = 0; i < lenPrimed; i++) {
      expect(tema.update(inputTasc[i])).toBeNaN();
      expect(tema.isPrimed()).toBe(false);
    }

    for (let i = lenPrimed; i < inputTasc.length; i++) {
      const act = tema.update(inputTasc[i]);
      expect(tema.isPrimed()).toBe(true);

      if (i >= firstCheck) {
        expect(act).toBeCloseTo(expectedTasc[i], epsilon);
      }
    }

    expect(tema.update(Number.NaN)).toBeNaN();
  });
});
