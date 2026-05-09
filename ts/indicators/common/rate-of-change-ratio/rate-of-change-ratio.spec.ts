import { } from 'jasmine';

import { RateOfChangeRatio } from './rate-of-change-ratio';
import { IndicatorIdentifier } from '../../core/indicator-identifier';
import { Shape } from '../../core/outputs/shape/shape';
import { RateOfChangeRatioOutput } from './output';
import { BarComponent } from '../../../entities/bar-component';
import { QuoteComponent } from '../../../entities/quote-component';
import { TradeComponent } from '../../../entities/trade-component';
import { input } from './testdata';

/* eslint-disable max-len */
// Input data is taken from the TA-Lib (http://ta-lib.org/) tests,
//    test_data.c, TA_SREF_close_daily_ref_0_PRIV[252].
//
// Output data, length=14.
// Taken from TA-Lib (http://ta-lib.org/) tests, test_mom.c.
//
// ROCR TEST (price/previousPrice):
// { 1, TA_ROCR_TEST, 0, 251, 14, TA_SUCCESS,      0, 0.994536,  14,  252-14 },
// { 0, TA_ROCR_TEST, 0, 251, 14, TA_SUCCESS,      1, 0.978906,  14,  252-14 },
// { 0, TA_ROCR_TEST, 0, 251, 14, TA_SUCCESS,      2, 0.944689,  14,  252-14 },
// { 0, TA_ROCR_TEST, 0, 251, 14, TA_SUCCESS, 252-15, 0.989633,  14,  252-14 },
//
// ROCR100 TEST (price/previousPrice)*100:
// { 1, TA_ROCR100_TEST, 0, 251, 14, TA_SUCCESS,      0, 99.4536,  14,  252-14 },
// { 0, TA_ROCR100_TEST, 0, 251, 14, TA_SUCCESS,      1, 97.8906,  14,  252-14 },
// { 0, TA_ROCR100_TEST, 0, 251, 14, TA_SUCCESS,      2, 94.4689,  14,  252-14 },
// { 0, TA_ROCR100_TEST, 0, 251, 14, TA_SUCCESS, 252-15, 98.9633,  14,  252-14 },

describe('RateOfChangeRatio', () => {

  it('should have correct output enum value', () => {
    expect(RateOfChangeRatioOutput.RateOfChangeRatioValue).toBe(0);
  });

  it('should return expected mnemonic for ROCR with default components', () => {
    const rocr = new RateOfChangeRatio({length: 14});
    expect(rocr.metadata().mnemonic).toBe('rocr(14)');
  });

  it('should return expected mnemonic for ROCR100 with default components', () => {
    const rocr100 = new RateOfChangeRatio({length: 14, hundredScale: true});
    expect(rocr100.metadata().mnemonic).toBe('rocr100(14)');
    expect(rocr100.metadata().description).toBe('Rate of Change Ratio 100 Scale rocr100(14)');
  });

  it('should return expected mnemonic with non-default bar component', () => {
    const rocr = new RateOfChangeRatio({length: 14, barComponent: BarComponent.Median});
    expect(rocr.metadata().mnemonic).toBe('rocr(14, hl/2)');
    expect(rocr.metadata().description).toBe('Rate of Change Ratio rocr(14, hl/2)');
  });

  it('should return expected mnemonic with non-default quote component', () => {
    const rocr = new RateOfChangeRatio({length: 14, quoteComponent: QuoteComponent.Bid});
    expect(rocr.metadata().mnemonic).toBe('rocr(14, b)');
    expect(rocr.metadata().description).toBe('Rate of Change Ratio rocr(14, b)');
  });

  it('should return expected mnemonic with non-default trade component', () => {
    const rocr = new RateOfChangeRatio({length: 14, tradeComponent: TradeComponent.Volume});
    expect(rocr.metadata().mnemonic).toBe('rocr(14, v)');
    expect(rocr.metadata().description).toBe('Rate of Change Ratio rocr(14, v)');
  });

  it('should return expected metadata', () => {
    const rocr = new RateOfChangeRatio({length: 5});
    const meta = rocr.metadata();

    expect(meta.identifier).toBe(IndicatorIdentifier.RateOfChangeRatio);
    expect(meta.mnemonic).toBe('rocr(5)');
    expect(meta.description).toBe('Rate of Change Ratio rocr(5)');
    expect(meta.outputs.length).toBe(1);
    expect(meta.outputs[0].kind).toBe(RateOfChangeRatioOutput.RateOfChangeRatioValue);
    expect(meta.outputs[0].shape).toBe(Shape.Scalar);
    expect(meta.outputs[0].mnemonic).toBe('rocr(5)');
    expect(meta.outputs[0].description).toBe('Rate of Change Ratio rocr(5)');
  });

  it('should throw if length is less than 1', () => {
    expect(() => { new RateOfChangeRatio({length: 0}); }).toThrow();
    expect(() => { new RateOfChangeRatio({length: -1}); }).toThrow();
  });

  it('should calculate expected ROCR output for length = 14', () => {
    const rocr = new RateOfChangeRatio({length: 14});
    const eps = 1e-4;

    for (let i = 0; i < 13; i++) {
      expect(rocr.update(input[i])).toBeNaN();
    }

    for (let i = 13; i < input.length; i++) {
      const act = rocr.update(input[i]);

      switch (i) {
        case 14:
          expect(Math.abs(act - 0.994536)).toBeLessThan(eps);
          break;
        case 15:
          expect(Math.abs(act - 0.978906)).toBeLessThan(eps);
          break;
        case 16:
          expect(Math.abs(act - 0.944689)).toBeLessThan(eps);
          break;
        case 251:
          expect(Math.abs(act - 0.989633)).toBeLessThan(eps);
          break;
      }
    }

    expect(rocr.update(Number.NaN)).toBeNaN();
  });

  it('should calculate expected ROCR100 output for length = 14', () => {
    const rocr100 = new RateOfChangeRatio({length: 14, hundredScale: true});
    const eps = 1e-4;

    for (let i = 0; i < 13; i++) {
      expect(rocr100.update(input[i])).toBeNaN();
    }

    for (let i = 13; i < input.length; i++) {
      const act = rocr100.update(input[i]);

      switch (i) {
        case 14:
          expect(Math.abs(act - 99.4536)).toBeLessThan(eps);
          break;
        case 15:
          expect(Math.abs(act - 97.8906)).toBeLessThan(eps);
          break;
        case 16:
          expect(Math.abs(act - 94.4689)).toBeLessThan(eps);
          break;
        case 251:
          expect(Math.abs(act - 98.9633)).toBeLessThan(eps);
          break;
      }
    }

    expect(rocr100.update(Number.NaN)).toBeNaN();
  });

  it('should track primed state correctly for length = 1', () => {
    const rocr = new RateOfChangeRatio({length: 1});
    expect(rocr.isPrimed()).toBe(false);

    rocr.update(input[0]);
    expect(rocr.isPrimed()).toBe(false);

    rocr.update(input[1]);
    expect(rocr.isPrimed()).toBe(true);
  });

  it('should track primed state correctly for length = 2', () => {
    const rocr = new RateOfChangeRatio({length: 2});
    expect(rocr.isPrimed()).toBe(false);

    for (let i = 0; i < 2; i++) {
      rocr.update(input[i]);
      expect(rocr.isPrimed()).toBe(false);
    }

    rocr.update(input[2]);
    expect(rocr.isPrimed()).toBe(true);
  });

  it('should track primed state correctly for length = 5', () => {
    const rocr = new RateOfChangeRatio({length: 5});
    expect(rocr.isPrimed()).toBe(false);

    for (let i = 0; i < 5; i++) {
      rocr.update(input[i]);
      expect(rocr.isPrimed()).toBe(false);
    }

    rocr.update(input[5]);
    expect(rocr.isPrimed()).toBe(true);
  });

  it('should track primed state correctly for length = 10', () => {
    const rocr = new RateOfChangeRatio({length: 10});
    expect(rocr.isPrimed()).toBe(false);

    for (let i = 0; i < 10; i++) {
      rocr.update(input[i]);
      expect(rocr.isPrimed()).toBe(false);
    }

    rocr.update(input[10]);
    expect(rocr.isPrimed()).toBe(true);
  });
});
