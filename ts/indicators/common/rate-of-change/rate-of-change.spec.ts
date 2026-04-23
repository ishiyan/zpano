import { } from 'jasmine';

import { RateOfChange } from './rate-of-change';
import { IndicatorIdentifier } from '../../core/indicator-identifier';
import { Shape } from '../../core/outputs/shape/shape';
import { RateOfChangeOutput } from './output';
import { BarComponent } from '../../../entities/bar-component';
import { QuoteComponent } from '../../../entities/quote-component';
import { TradeComponent } from '../../../entities/trade-component';

/* eslint-disable max-len */
// Input data is taken from the TA-Lib (http://ta-lib.org/) tests,
//    test_data.c, TA_SREF_close_daily_ref_0_PRIV[252].
//
// Output data, length=14.
// Taken from TA-Lib (http://ta-lib.org/) tests, test_mom.c (ROC TEST section).

const input = [
  91.500000,94.815000,94.375000,95.095000,93.780000,94.625000,92.530000,92.750000,90.315000,92.470000,96.125000,
  97.250000,98.500000,89.875000,91.000000,92.815000,89.155000,89.345000,91.625000,89.875000,88.375000,87.625000,
  84.780000,83.000000,83.500000,81.375000,84.440000,89.250000,86.375000,86.250000,85.250000,87.125000,85.815000,
  88.970000,88.470000,86.875000,86.815000,84.875000,84.190000,83.875000,83.375000,85.500000,89.190000,89.440000,
  91.095000,90.750000,91.440000,89.000000,91.000000,90.500000,89.030000,88.815000,84.280000,83.500000,82.690000,
  84.750000,85.655000,86.190000,88.940000,89.280000,88.625000,88.500000,91.970000,91.500000,93.250000,93.500000,
  93.155000,91.720000,90.000000,89.690000,88.875000,85.190000,83.375000,84.875000,85.940000,97.250000,99.875000,
  104.940000,106.000000,102.500000,102.405000,104.595000,106.125000,106.000000,106.065000,104.625000,108.625000,
  109.315000,110.500000,112.750000,123.000000,119.625000,118.750000,119.250000,117.940000,116.440000,115.190000,
  111.875000,110.595000,118.125000,116.000000,116.000000,112.000000,113.750000,112.940000,116.000000,120.500000,
  116.620000,117.000000,115.250000,114.310000,115.500000,115.870000,120.690000,120.190000,120.750000,124.750000,
  123.370000,122.940000,122.560000,123.120000,122.560000,124.620000,129.250000,131.000000,132.250000,131.000000,
  132.810000,134.000000,137.380000,137.810000,137.880000,137.250000,136.310000,136.250000,134.630000,128.250000,
  129.000000,123.870000,124.810000,123.000000,126.250000,128.380000,125.370000,125.690000,122.250000,119.370000,
  118.500000,123.190000,123.500000,122.190000,119.310000,123.310000,121.120000,123.370000,127.370000,128.500000,
  123.870000,122.940000,121.750000,124.440000,122.000000,122.370000,122.940000,124.000000,123.190000,124.560000,
  127.250000,125.870000,128.860000,132.000000,130.750000,134.750000,135.000000,132.380000,133.310000,131.940000,
  130.000000,125.370000,130.130000,127.120000,125.190000,122.000000,125.000000,123.000000,123.500000,120.060000,
  121.000000,117.750000,119.870000,122.000000,119.190000,116.370000,113.500000,114.250000,110.000000,105.060000,
  107.000000,107.870000,107.000000,107.120000,107.000000,91.000000,93.940000,93.870000,95.500000,93.000000,
  94.940000,98.250000,96.750000,94.810000,94.370000,91.560000,90.250000,93.940000,93.620000,97.000000,95.000000,
  95.870000,94.060000,94.620000,93.750000,98.000000,103.940000,107.870000,106.060000,104.500000,105.000000,
  104.190000,103.060000,103.420000,105.270000,111.870000,116.000000,116.620000,118.280000,113.370000,109.000000,
  109.700000,109.250000,107.000000,109.190000,110.000000,109.200000,110.120000,108.000000,108.620000,109.750000,
  109.810000,109.000000,108.750000,107.870000
];

describe('RateOfChange', () => {

  it('should have correct output enum value', () => {
    expect(RateOfChangeOutput.RateOfChangeValue).toBe(0);
  });

  it('should return expected mnemonic for default components', () => {
    const roc = new RateOfChange({length: 14});
    expect(roc.metadata().mnemonic).toBe('roc(14)');
  });

  it('should return expected mnemonic with non-default bar component', () => {
    const roc = new RateOfChange({length: 14, barComponent: BarComponent.Median});
    expect(roc.metadata().mnemonic).toBe('roc(14, hl/2)');
    expect(roc.metadata().description).toBe('Rate of Change roc(14, hl/2)');
  });

  it('should return expected mnemonic with non-default quote component', () => {
    const roc = new RateOfChange({length: 14, quoteComponent: QuoteComponent.Bid});
    expect(roc.metadata().mnemonic).toBe('roc(14, b)');
    expect(roc.metadata().description).toBe('Rate of Change roc(14, b)');
  });

  it('should return expected mnemonic with non-default trade component', () => {
    const roc = new RateOfChange({length: 14, tradeComponent: TradeComponent.Volume});
    expect(roc.metadata().mnemonic).toBe('roc(14, v)');
    expect(roc.metadata().description).toBe('Rate of Change roc(14, v)');
  });

  it('should return expected metadata', () => {
    const roc = new RateOfChange({length: 5});
    const meta = roc.metadata();

    expect(meta.identifier).toBe(IndicatorIdentifier.RateOfChange);
    expect(meta.mnemonic).toBe('roc(5)');
    expect(meta.description).toBe('Rate of Change roc(5)');
    expect(meta.outputs.length).toBe(1);
    expect(meta.outputs[0].kind).toBe(RateOfChangeOutput.RateOfChangeValue);
    expect(meta.outputs[0].shape).toBe(Shape.Scalar);
    expect(meta.outputs[0].mnemonic).toBe('roc(5)');
    expect(meta.outputs[0].description).toBe('Rate of Change roc(5)');
  });

  it('should throw if length is less than 1', () => {
    expect(() => { new RateOfChange({length: 0}); }).toThrow();
    expect(() => { new RateOfChange({length: -1}); }).toThrow();
  });

  it('should calculate expected output for length = 14', () => {
    const roc = new RateOfChange({length: 14});
    const eps = 1e-2;

    // Values from index=0 to index=13 are NaN (first 14 updates).
    for (let i = 0; i < 13; i++) {
      expect(roc.update(input[i])).toBeNaN();
    }

    // From index 13 on, the indicator should return values starting at index 14.
    for (let i = 13; i < input.length; i++) {
      const act = roc.update(input[i]);

      switch (i) {
        case 14:
          expect(Math.abs(act - (-0.546))).toBeLessThan(eps);
          break;
        case 15:
          expect(Math.abs(act - (-2.109))).toBeLessThan(eps);
          break;
        case 16:
          expect(Math.abs(act - (-5.53))).toBeLessThan(eps);
          break;
        case 251:
          expect(Math.abs(act - (-1.0367))).toBeLessThan(eps);
          break;
      }
    }

    expect(roc.update(Number.NaN)).toBeNaN();
  });

  it('should track primed state correctly for length = 1', () => {
    const roc = new RateOfChange({length: 1});
    expect(roc.isPrimed()).toBe(false);

    roc.update(input[0]);
    expect(roc.isPrimed()).toBe(false);

    roc.update(input[1]);
    expect(roc.isPrimed()).toBe(true);
  });

  it('should track primed state correctly for length = 2', () => {
    const roc = new RateOfChange({length: 2});
    expect(roc.isPrimed()).toBe(false);

    for (let i = 0; i < 2; i++) {
      roc.update(input[i]);
      expect(roc.isPrimed()).toBe(false);
    }

    roc.update(input[2]);
    expect(roc.isPrimed()).toBe(true);
  });

  it('should track primed state correctly for length = 5', () => {
    const roc = new RateOfChange({length: 5});
    expect(roc.isPrimed()).toBe(false);

    for (let i = 0; i < 5; i++) {
      roc.update(input[i]);
      expect(roc.isPrimed()).toBe(false);
    }

    roc.update(input[5]);
    expect(roc.isPrimed()).toBe(true);
  });

  it('should track primed state correctly for length = 10', () => {
    const roc = new RateOfChange({length: 10});
    expect(roc.isPrimed()).toBe(false);

    for (let i = 0; i < 10; i++) {
      roc.update(input[i]);
      expect(roc.isPrimed()).toBe(false);
    }

    roc.update(input[10]);
    expect(roc.isPrimed()).toBe(true);
  });
});
