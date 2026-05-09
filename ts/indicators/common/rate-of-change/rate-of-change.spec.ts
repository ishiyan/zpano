import { } from 'jasmine';

import { RateOfChange } from './rate-of-change';
import { IndicatorIdentifier } from '../../core/indicator-identifier';
import { Shape } from '../../core/outputs/shape/shape';
import { RateOfChangeOutput } from './output';
import { BarComponent } from '../../../entities/bar-component';
import { QuoteComponent } from '../../../entities/quote-component';
import { TradeComponent } from '../../../entities/trade-component';
import { input } from './testdata';

/* eslint-disable max-len */
// Input data is taken from the TA-Lib (http://ta-lib.org/) tests,
//    test_data.c, TA_SREF_close_daily_ref_0_PRIV[252].
//
// Output data, length=14.
// Taken from TA-Lib (http://ta-lib.org/) tests, test_mom.c (ROC TEST section).

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
