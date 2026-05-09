import { } from 'jasmine';

import { StochasticRelativeStrengthIndex } from './stochastic-relative-strength-index';
import { MovingAverageType } from './params';
import { StochasticRelativeStrengthIndexOutput } from './output';
import { IndicatorIdentifier } from '../../core/indicator-identifier';
import { Shape } from '../../core/outputs/shape/shape';
import { input } from './testdata';

// Test data from TA-Lib (252 entries).
describe('StochasticRelativeStrengthIndex', () => {

  // Test case 1: period=14, fastK=14, fastD=1, SMA.
  // begIndex=27, first: FastK=94.156709, FastD=94.156709.
  // last(251): FastK=0.0, FastD=0.0.
  it('should calculate 14/14/1 SMA correctly', () => {
    const tolerance = 1e-4;
    const ind = new StochasticRelativeStrengthIndex({
      length: 14,
      fastKLength: 14,
      fastDLength: 1,
    });

    // First 27 values should produce NaN FastK.
    for (let i = 0; i < 27; i++) {
      const [fastK] = ind.update(input[i]);
      expect(fastK).withContext(`index ${i} should be NaN`).toBeNaN();
    }

    // Index 27: first value.
    let [fastK, fastD] = ind.update(input[27]);
    expect(Math.abs(fastK - 94.156709)).withContext('first FastK').toBeLessThan(tolerance);
    expect(Math.abs(fastD - 94.156709)).withContext('first FastD').toBeLessThan(tolerance);

    // Feed remaining and check last value.
    for (let i = 28; i < 251; i++) {
      ind.update(input[i]);
    }

    [fastK, fastD] = ind.update(input[251]);
    expect(Math.abs(fastK - 0.0)).withContext('last FastK').toBeLessThan(tolerance);
    expect(Math.abs(fastD - 0.0)).withContext('last FastD').toBeLessThan(tolerance);
  });

  // Test case 2: period=14, fastK=45, fastD=1, SMA.
  // begIndex=58, first: FastK=79.729186, FastD=79.729186.
  // last(251): FastK=48.1550743, FastD=48.1550743.
  it('should calculate 14/45/1 SMA correctly', () => {
    const tolerance = 1e-4;
    const ind = new StochasticRelativeStrengthIndex({
      length: 14,
      fastKLength: 45,
      fastDLength: 1,
    });

    // First 58 values should produce NaN FastK.
    for (let i = 0; i < 58; i++) {
      const [fastK] = ind.update(input[i]);
      expect(fastK).withContext(`index ${i} should be NaN`).toBeNaN();
    }

    // Index 58: first value.
    let [fastK, fastD] = ind.update(input[58]);
    expect(Math.abs(fastK - 79.729186)).withContext('first FastK').toBeLessThan(tolerance);
    expect(Math.abs(fastD - 79.729186)).withContext('first FastD').toBeLessThan(tolerance);

    // Feed remaining and check last value.
    for (let i = 59; i < 251; i++) {
      ind.update(input[i]);
    }

    [fastK, fastD] = ind.update(input[251]);
    expect(Math.abs(fastK - 48.1550743)).withContext('last FastK').toBeLessThan(tolerance);
    expect(Math.abs(fastD - 48.1550743)).withContext('last FastD').toBeLessThan(tolerance);
  });

  // Test case 3: period=11, fastK=13, fastD=16, SMA.
  // begIndex=38, first: FastK=5.25947, FastD=57.1711.
  // last(251): FastK=0.0, FastD=15.7303.
  it('should calculate 11/13/16 SMA correctly', () => {
    const tolerance = 1e-3;
    const ind = new StochasticRelativeStrengthIndex({
      length: 11,
      fastKLength: 13,
      fastDLength: 16,
    });

    // Feed first 38 values.
    for (let i = 0; i < 38; i++) {
      ind.update(input[i]);
    }

    // Index 38: first primed value.
    let [fastK, fastD] = ind.update(input[38]);
    expect(Math.abs(fastK - 5.25947)).withContext('first FastK').toBeLessThan(tolerance);
    expect(Math.abs(fastD - 57.1711)).withContext('first FastD').toBeLessThan(tolerance);
    expect(ind.isPrimed()).toBe(true);

    // Feed remaining and check last value.
    for (let i = 39; i < 251; i++) {
      ind.update(input[i]);
    }

    [fastK, fastD] = ind.update(input[251]);
    expect(Math.abs(fastK - 0.0)).withContext('last FastK').toBeLessThan(tolerance);
    expect(Math.abs(fastD - 15.7303)).withContext('last FastD').toBeLessThan(tolerance);
  });

  it('should report correct primed state', () => {
    const ind = new StochasticRelativeStrengthIndex({
      length: 14,
      fastKLength: 14,
      fastDLength: 1,
    });

    expect(ind.isPrimed()).toBe(false);

    for (let i = 0; i < 27; i++) {
      ind.update(input[i]);
      expect(ind.isPrimed()).withContext(`index ${i}`).toBe(false);
    }

    ind.update(input[27]);
    expect(ind.isPrimed()).toBe(true);
  });

  it('should handle NaN passthrough', () => {
    const ind = new StochasticRelativeStrengthIndex({
      length: 14,
      fastKLength: 14,
      fastDLength: 1,
    });

    const [fastK, fastD] = ind.update(NaN);
    expect(fastK).toBeNaN();
    expect(fastD).toBeNaN();
  });

  it('should validate parameters', () => {
    expect(() => new StochasticRelativeStrengthIndex({ length: 1, fastKLength: 14, fastDLength: 3 })).toThrow();
    expect(() => new StochasticRelativeStrengthIndex({ length: 14, fastKLength: 0, fastDLength: 3 })).toThrow();
    expect(() => new StochasticRelativeStrengthIndex({ length: 14, fastKLength: 14, fastDLength: 0 })).toThrow();
    expect(() => new StochasticRelativeStrengthIndex({ length: 14, fastKLength: 14, fastDLength: 3 })).not.toThrow();
  });

  it('should return correct metadata (SMA)', () => {
    const ind = new StochasticRelativeStrengthIndex({
      length: 14,
      fastKLength: 14,
      fastDLength: 3,
    });
    const meta = ind.metadata();

    expect(meta.identifier).toBe(IndicatorIdentifier.StochasticRelativeStrengthIndex);
    expect(meta.mnemonic).toBe('stochrsi(14/14/SMA3)');
    expect(meta.outputs.length).toBe(2);
    expect(meta.outputs[0].kind).toBe(StochasticRelativeStrengthIndexOutput.FastK);
    expect(meta.outputs[0].shape).toBe(Shape.Scalar);
    expect(meta.outputs[1].kind).toBe(StochasticRelativeStrengthIndexOutput.FastD);
    expect(meta.outputs[1].shape).toBe(Shape.Scalar);
  });

  it('should return correct metadata (EMA)', () => {
    const ind = new StochasticRelativeStrengthIndex({
      length: 14,
      fastKLength: 14,
      fastDLength: 3,
      movingAverageType: MovingAverageType.EMA,
    });
    const meta = ind.metadata();

    expect(meta.mnemonic).toBe('stochrsi(14/14/EMA3)');
  });
});
