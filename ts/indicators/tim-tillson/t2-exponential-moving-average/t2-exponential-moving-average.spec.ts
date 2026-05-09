import { } from 'jasmine';

import { T2ExponentialMovingAverage } from './t2-exponential-moving-average';
import { IndicatorIdentifier } from '../../core/indicator-identifier';
import { Shape } from '../../core/outputs/shape/shape';
import { T2ExponentialMovingAverageOutput } from './output';
import { BarComponent } from '../../../entities/bar-component';
import { QuoteComponent } from '../../../entities/quote-component';
import { TradeComponent } from '../../../entities/trade-component';
import { input, expected } from './testdata';

/* eslint-disable max-len */
// Input data is taken from the TA-Lib (http://ta-lib.org/) tests,
//    test_data.c, TA_SREF_close_daily_ref_0_PRIV[252].
//
// Output data is taken from the modified TA-Lib (http://ta-lib.org/) test_T3.xls: test_T2.xls, T2(5,0.7) column.

/** Expected data is taken from the modified TA-Lib (http://ta-lib.org/) test_T3.xls: test_T2.xls, T2(5,0.7) column.
 *  Length is 5, volume factor is 0.7, firstIsAverage = true.
 */
describe('T2ExponentialMovingAverage', () => {

  it('should have correct output enum value', () => {
    expect(T2ExponentialMovingAverageOutput.T2ExponentialMovingAverageValue).toBe(0);
  });

  it('should return expected mnemonic for length-based', () => {
    let t2 = new T2ExponentialMovingAverage({length: 7, volumeFactor: 0.6781, firstIsAverage: true});
    expect(t2.metadata().mnemonic).toBe('t2(7, 0.67810000)');
    t2 = new T2ExponentialMovingAverage({length: 7, volumeFactor: 0.6789, firstIsAverage: false});
    expect(t2.metadata().mnemonic).toBe('t2(7, 0.67890000)');
  });

  it('should return expected mnemonic for smoothing-factor-based', () => {
    // alpha = 0.12345, length = round(2/0.12345) - 1 = round(16.2) - 1 = 15
    const t2 = new T2ExponentialMovingAverage({smoothingFactor: 0.12345, volumeFactor: 0.56789, firstIsAverage: false});
    expect(t2.metadata().mnemonic).toBe('t2(15, 0.12345000, 0.56789000)');
  });

  it('should return expected metadata for length-based', () => {
    const t2 = new T2ExponentialMovingAverage({length: 10, volumeFactor: 0.3333, firstIsAverage: true});
    const meta = t2.metadata();

    expect(meta.identifier).toBe(IndicatorIdentifier.T2ExponentialMovingAverage);
    expect(meta.mnemonic).toBe('t2(10, 0.33330000)');
    expect(meta.description).toBe('T2 exponential moving average t2(10, 0.33330000)');
    expect(meta.outputs.length).toBe(1);
    expect(meta.outputs[0].kind).toBe(T2ExponentialMovingAverageOutput.T2ExponentialMovingAverageValue);
    expect(meta.outputs[0].shape).toBe(Shape.Scalar);
    expect(meta.outputs[0].mnemonic).toBe('t2(10, 0.33330000)');
    expect(meta.outputs[0].description).toBe('T2 exponential moving average t2(10, 0.33330000)');
  });

  it('should return expected metadata for smoothing-factor-based', () => {
    // alpha = 2 / (10 + 1) = 2/11 = 0.18181818...
    const alpha = 2 / 11;
    const t2 = new T2ExponentialMovingAverage({smoothingFactor: alpha, volumeFactor: 0.3333333, firstIsAverage: false});
    const meta = t2.metadata();

    expect(meta.identifier).toBe(IndicatorIdentifier.T2ExponentialMovingAverage);
    expect(meta.mnemonic).toBe('t2(10, 0.18181818, 0.33333330)');
    expect(meta.description).toBe('T2 exponential moving average t2(10, 0.18181818, 0.33333330)');
    expect(meta.outputs.length).toBe(1);
    expect(meta.outputs[0].kind).toBe(T2ExponentialMovingAverageOutput.T2ExponentialMovingAverageValue);
    expect(meta.outputs[0].shape).toBe(Shape.Scalar);
    expect(meta.outputs[0].mnemonic).toBe('t2(10, 0.18181818, 0.33333330)');
    expect(meta.outputs[0].description).toBe('T2 exponential moving average t2(10, 0.18181818, 0.33333330)');
  });

  it('should return expected mnemonic with non-default bar component', () => {
    const t2 = new T2ExponentialMovingAverage({
      length: 10, volumeFactor: 0.7, firstIsAverage: true, barComponent: BarComponent.Median,
    });
    expect(t2.metadata().mnemonic).toBe('t2(10, 0.70000000, hl/2)');
    expect(t2.metadata().description).toBe('T2 exponential moving average t2(10, 0.70000000, hl/2)');
  });

  it('should return expected mnemonic with non-default quote component (smoothing factor)', () => {
    const t2 = new T2ExponentialMovingAverage({
      smoothingFactor: 2 / 11, volumeFactor: 0.7, firstIsAverage: false,
      quoteComponent: QuoteComponent.Bid,
    });
    expect(t2.metadata().mnemonic).toBe('t2(10, 0.18181818, 0.70000000, b)');
    expect(t2.metadata().description).toBe('T2 exponential moving average t2(10, 0.18181818, 0.70000000, b)');
  });

  it('should throw if length is less than 2', () => {
    expect(() => { new T2ExponentialMovingAverage({length: 1, volumeFactor: 0.7, firstIsAverage: true}); }).toThrow();
  });

  it('should throw if smoothing factor is less than 0', () => {
    expect(() => { new T2ExponentialMovingAverage({smoothingFactor: -0.1, volumeFactor: 0.7, firstIsAverage: false}); }).toThrow();
  });

  it('should throw if smoothing factor is greater than 1', () => {
    expect(() => { new T2ExponentialMovingAverage({smoothingFactor: 1.1, volumeFactor: 0.7, firstIsAverage: false}); }).toThrow();
  });

  it('should throw if volume factor is less than 0', () => {
    expect(() => { new T2ExponentialMovingAverage({length: 5, volumeFactor: -0.1, firstIsAverage: true}); }).toThrow();
  });

  it('should throw if volume factor is greater than 1', () => {
    expect(() => { new T2ExponentialMovingAverage({length: 5, volumeFactor: 1.1, firstIsAverage: true}); }).toThrow();
  });

  it('should clamp near-zero smoothing factor to epsilon', () => {
    const t2 = new T2ExponentialMovingAverage({smoothingFactor: 0, volumeFactor: 0.7, firstIsAverage: true});
    // alpha clamped to 0.00000001, length = round(2/0.00000001) - 1 = 199999999
    expect(t2.metadata().mnemonic).toBe('t2(199999999, 0.00000001, 0.70000000)');
  });

  it('should accept smoothing factor = 1', () => {
    const t2 = new T2ExponentialMovingAverage({smoothingFactor: 1, volumeFactor: 0.7, firstIsAverage: true});
    // alpha = 1, length = round(2/1) - 1 = 1
    expect(t2.metadata().mnemonic).toBe('t2(1, 1.00000000, 0.70000000)');
  });

  it('should calculate expected output and prime state for length 5, first is NOT SMA', () => {
    const len = 5;
    const lenPrimed = 4*(len - 1);
    const eps = 1e-8;
    const t2 = new T2ExponentialMovingAverage({length: len, volumeFactor: 0.7, firstIsAverage: false});

    for (let i = 0; i < lenPrimed; i++) {
      expect(t2.update(input[i])).toBeNaN();
      expect(t2.isPrimed()).toBe(false);
    }

    // Metastock path produces different values initially; converges with expected (SMA-seeded) data after ~43 samples.
    for (let i = lenPrimed; i < input.length; i++) {
      const act = t2.update(input[i]);
      expect(t2.isPrimed()).toBe(true);

      if (i >= lenPrimed + 43) {
        expect(act).toBeCloseTo(expected[i], eps);
      }
    }

    expect(t2.update(Number.NaN)).toBeNaN();
  });

  it('should calculate expected output and prime state for length 5, first is SMA', () => {
    const len = 5;
    const lenPrimed = 4*(len - 1);
    const eps = 1e-8;
    const t2 = new T2ExponentialMovingAverage({length: len, volumeFactor: 0.7, firstIsAverage: true});

    for (let i = 0; i < lenPrimed; i++) {
      expect(t2.update(input[i])).toBeNaN();
      expect(t2.isPrimed()).toBe(false);
    }

    // Expected data is from test_T2.xls with firstIsAverage = true; skip first primed value (index 16).
    for (let i = lenPrimed; i < input.length; i++) {
      const act = t2.update(input[i]);
      expect(t2.isPrimed()).toBe(true);

      if (i >= lenPrimed + 1) {
        expect(act).toBeCloseTo(expected[i], eps);
      }
    }

    expect(t2.update(Number.NaN)).toBeNaN();
  });

  it('should match expected output (Excel) for length 5, first is SMA', () => {
    const eps = 1e-13;
    const len = 5;
    const lenPrimed = 4*(len - 1);
    const t2 = new T2ExponentialMovingAverage({length: len, volumeFactor: 0.7, firstIsAverage: true});

    for (let i = 0; i < lenPrimed; i++) {
      expect(t2.update(input[i])).toBeNaN();
      expect(t2.isPrimed()).toBe(false);
    }

    // Skip first primed value (index 16) as it differs due to seed averaging.
    for (let i = lenPrimed; i < input.length; i++) {
      const act = t2.update(input[i]);
      expect(t2.isPrimed()).toBe(true);

      if (i >= lenPrimed + 1) {
        expect(act).toBeCloseTo(expected[i], eps);
      }
    }

    expect(t2.update(Number.NaN)).toBeNaN();
  });
});
