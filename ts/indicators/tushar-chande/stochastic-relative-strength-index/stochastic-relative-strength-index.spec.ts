import { } from 'jasmine';

import { StochasticRelativeStrengthIndex } from './stochastic-relative-strength-index';
import { MovingAverageType } from './stochastic-relative-strength-index-params';
import { StochasticRelativeStrengthIndexOutput } from './stochastic-relative-strength-index-output';
import { IndicatorType } from '../../core/indicator-type';
import { OutputType } from '../../core/outputs/output-type';

// Test data from TA-Lib (252 entries).
const input = [
  91.500000, 94.815000, 94.375000, 95.095000, 93.780000, 94.625000, 92.530000, 92.750000, 90.315000, 92.470000, 96.125000,
  97.250000, 98.500000, 89.875000, 91.000000, 92.815000, 89.155000, 89.345000, 91.625000, 89.875000, 88.375000, 87.625000,
  84.780000, 83.000000, 83.500000, 81.375000, 84.440000, 89.250000, 86.375000, 86.250000, 85.250000, 87.125000, 85.815000,
  88.970000, 88.470000, 86.875000, 86.815000, 84.875000, 84.190000, 83.875000, 83.375000, 85.500000, 89.190000, 89.440000,
  91.095000, 90.750000, 91.440000, 89.000000, 91.000000, 90.500000, 89.030000, 88.815000, 84.280000, 83.500000, 82.690000,
  84.750000, 85.655000, 86.190000, 88.940000, 89.280000, 88.625000, 88.500000, 91.970000, 91.500000, 93.250000, 93.500000,
  93.155000, 91.720000, 90.000000, 89.690000, 88.875000, 85.190000, 83.375000, 84.875000, 85.940000, 97.250000, 99.875000,
  104.940000, 106.000000, 102.500000, 102.405000, 104.595000, 106.125000, 106.000000, 106.065000, 104.625000, 108.625000,
  109.315000, 110.500000, 112.750000, 123.000000, 119.625000, 118.750000, 119.250000, 117.940000, 116.440000, 115.190000,
  111.875000, 110.595000, 118.125000, 116.000000, 116.000000, 112.000000, 113.750000, 112.940000, 116.000000, 120.500000,
  116.620000, 117.000000, 115.250000, 114.310000, 115.500000, 115.870000, 120.690000, 120.190000, 120.750000, 124.750000,
  123.370000, 122.940000, 122.560000, 123.120000, 122.560000, 124.620000, 129.250000, 131.000000, 132.250000, 131.000000,
  132.810000, 134.000000, 137.380000, 137.810000, 137.880000, 137.250000, 136.310000, 136.250000, 134.630000, 128.250000,
  129.000000, 123.870000, 124.810000, 123.000000, 126.250000, 128.380000, 125.370000, 125.690000, 122.250000, 119.370000,
  118.500000, 123.190000, 123.500000, 122.190000, 119.310000, 123.310000, 121.120000, 123.370000, 127.370000, 128.500000,
  123.870000, 122.940000, 121.750000, 124.440000, 122.000000, 122.370000, 122.940000, 124.000000, 123.190000, 124.560000,
  127.250000, 125.870000, 128.860000, 132.000000, 130.750000, 134.750000, 135.000000, 132.380000, 133.310000, 131.940000,
  130.000000, 125.370000, 130.130000, 127.120000, 125.190000, 122.000000, 125.000000, 123.000000, 123.500000, 120.060000,
  121.000000, 117.750000, 119.870000, 122.000000, 119.190000, 116.370000, 113.500000, 114.250000, 110.000000, 105.060000,
  107.000000, 107.870000, 107.000000, 107.120000, 107.000000, 91.000000, 93.940000, 93.870000, 95.500000, 93.000000,
  94.940000, 98.250000, 96.750000, 94.810000, 94.370000, 91.560000, 90.250000, 93.940000, 93.620000, 97.000000, 95.000000,
  95.870000, 94.060000, 94.620000, 93.750000, 98.000000, 103.940000, 107.870000, 106.060000, 104.500000, 105.000000,
  104.190000, 103.060000, 103.420000, 105.270000, 111.870000, 116.000000, 116.620000, 118.280000, 113.370000, 109.000000,
  109.700000, 109.250000, 107.000000, 109.190000, 110.000000, 109.200000, 110.120000, 108.000000, 108.620000, 109.750000,
  109.810000, 109.000000, 108.750000, 107.870000,
];

describe('StochasticRelativeStrengthIndex', () => {

  // Test case 1: period=14, fastK=14, fastD=1, SMA.
  // begIdx=27, first: FastK=94.156709, FastD=94.156709.
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
  // begIdx=58, first: FastK=79.729186, FastD=79.729186.
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
  // begIdx=38, first: FastK=5.25947, FastD=57.1711.
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

    expect(meta.type).toBe(IndicatorType.StochasticRelativeStrengthIndex);
    expect(meta.mnemonic).toBe('stochrsi(14/14/SMA3)');
    expect(meta.outputs.length).toBe(2);
    expect(meta.outputs[0].kind).toBe(StochasticRelativeStrengthIndexOutput.FastK);
    expect(meta.outputs[0].type).toBe(OutputType.Scalar);
    expect(meta.outputs[1].kind).toBe(StochasticRelativeStrengthIndexOutput.FastD);
    expect(meta.outputs[1].type).toBe(OutputType.Scalar);
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
