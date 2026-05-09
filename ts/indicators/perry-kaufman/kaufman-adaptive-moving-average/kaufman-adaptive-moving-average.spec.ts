import { } from 'jasmine';

import { KaufmanAdaptiveMovingAverage } from './kaufman-adaptive-moving-average';
import { IndicatorIdentifier } from '../../core/indicator-identifier';
import { Shape } from '../../core/outputs/shape/shape';
import { KaufmanAdaptiveMovingAverageOutput } from './output';
import { input, expected, expectedEr } from './testdata';

/* eslint-disable max-len */
// Input data is taken from the TA-Lib (http://ta-lib.org/) tests,
//    test_data.c, TA_SREF_close_daily_ref_0_PRIV[252].
//
// Expected data is taken from TA-Lib (http://ta-lib.org/) tests, test_KAMA.xsl, KAMA: J5…J256, ER: G5…G256, 252 entries.
// Efficiency ratio length is 10, fastest length is 2, slowest length is 30.

describe('KaufmanAdaptiveMovingAverage', () => {
  const epsilon = 1e-8;

  it('should return expected mnemonic for length-based', () => {
    const kama = new KaufmanAdaptiveMovingAverage({ efficiencyRatioLength: 10, fastestLength: 2, slowestLength: 30 });
    expect(kama.metadata().mnemonic).toBe('kama(10, 2, 30)');
  });

  it('should return expected mnemonic for smoothing-factor-based', () => {
    const kama = new KaufmanAdaptiveMovingAverage({ efficiencyRatioLength: 10, fastestSmoothingFactor: 2 / 3, slowestSmoothingFactor: 2 / 31 });
    expect(kama.metadata().mnemonic).toBe('kama(10, 0.6667, 0.0645)');
  });

  it('should return expected metadata', () => {
    const kama = new KaufmanAdaptiveMovingAverage({ efficiencyRatioLength: 10, fastestLength: 2, slowestLength: 30 });
    const meta = kama.metadata();

    expect(meta.identifier).toBe(IndicatorIdentifier.KaufmanAdaptiveMovingAverage);
    expect(meta.mnemonic).toBe('kama(10, 2, 30)');
    expect(meta.description).toBe('Kaufman adaptive moving average kama(10, 2, 30)');
    expect(meta.outputs.length).toBe(1);
    expect(meta.outputs[0].kind).toBe(KaufmanAdaptiveMovingAverageOutput.KaufmanAdaptiveMovingAverageValue);
    expect(meta.outputs[0].shape).toBe(Shape.Scalar);
    expect(meta.outputs[0].mnemonic).toBe('kama(10, 2, 30)');
    expect(meta.outputs[0].description).toBe('Kaufman adaptive moving average kama(10, 2, 30)');
  });

  it('should throw if efficiency ratio length is less than 2', () => {
    expect(() => { new KaufmanAdaptiveMovingAverage({ efficiencyRatioLength: 1, fastestLength: 2, slowestLength: 30 }); }).toThrow();
  });

  it('should throw if the fastest length is less than 2', () => {
    expect(() => { new KaufmanAdaptiveMovingAverage({ efficiencyRatioLength: 10, fastestLength: 1, slowestLength: 30 }); }).toThrow();
  });

  it('should throw if the slowest length is less than 2', () => {
    expect(() => { new KaufmanAdaptiveMovingAverage({ efficiencyRatioLength: 10, fastestLength: 2, slowestLength: 1 }); }).toThrow();
  });

  it('should throw if the fastest smoothing factor is less or equal to 0', () => {
    expect(() => { new KaufmanAdaptiveMovingAverage({ efficiencyRatioLength: 10, fastestSmoothingFactor: -0.01, slowestSmoothingFactor: 2 / 31 }); }).toThrow();
  });

  it('should throw if the fastest smoothing factor is greater or equal to 1', () => {
    expect(() => { new KaufmanAdaptiveMovingAverage({ efficiencyRatioLength: 10, fastestSmoothingFactor: 1.01, slowestSmoothingFactor: 2 / 31 }); }).toThrow();
  });

  it('should throw if the slowest smoothing factor is less or equal to 0', () => {
    expect(() => { new KaufmanAdaptiveMovingAverage({ efficiencyRatioLength: 10, fastestSmoothingFactor: 2 / 3, slowestSmoothingFactor: -0.01 }); }).toThrow();
  });

  it('should throw if the slowest smoothing factor is greater or equal to 1', () => {
    expect(() => { new KaufmanAdaptiveMovingAverage({ efficiencyRatioLength: 10, fastestSmoothingFactor: 2 / 3, slowestSmoothingFactor: 1.01 }); }).toThrow();
  });

  it('should calculate expected output and prime state', () => {
    const len = 10;
    const kama = new KaufmanAdaptiveMovingAverage({ efficiencyRatioLength: 10, fastestLength: 2, slowestLength: 30 });

    for (let i = 0; i < len; i++) {
      expect(kama.update(input[i])).toBeNaN();
      expect(kama.isPrimed()).toBe(false);
    }

    for (let i = len; i < input.length; i++) {
      const act = kama.update(input[i]);
      expect(kama.isPrimed()).toBe(true);
      expect(act).withContext(`i=${i}: expected ${expected[i]}, actual ${act}`).toBeCloseTo(expected[i], epsilon);
    }

    expect(kama.update(Number.NaN)).toBeNaN();
  });

  it('should calculate expected KAMA and ER outputs', () => {
    const len = 10;
    const kama = new KaufmanAdaptiveMovingAverage({ efficiencyRatioLength: 10, fastestLength: 2, slowestLength: 30 });

    for (let i = 0; i < len; i++) {
      expect(kama.update(input[i])).toBeNaN();
      expect(kama.isPrimed()).toBe(false);
    }

    for (let i = len; i < input.length; i++) {
      const actKama = kama.update(input[i]);
      expect(kama.isPrimed()).toBe(true);
      const actEr = kama.getEfficiencyRatio();
      expect(actKama).withContext(`i=${i}: expected ${expected[i]}, actual ${actKama}`).toBeCloseTo(expected[i], epsilon);
      expect(actEr).withContext(`i=${i}: expected ER ${expectedEr[i]}, actual ${actEr}`).toBeCloseTo(expectedEr[i], epsilon);
    }

    expect(kama.update(Number.NaN)).toBeNaN();
  });

  it('should return NaN for efficiency ratio when not primed', () => {
    const kama = new KaufmanAdaptiveMovingAverage({ efficiencyRatioLength: 10, fastestLength: 2, slowestLength: 30 });
    expect(kama.getEfficiencyRatio()).toBeNaN();
  });
});
