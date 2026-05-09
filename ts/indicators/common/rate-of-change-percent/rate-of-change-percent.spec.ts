import { } from 'jasmine';

import { RateOfChangePercent } from './rate-of-change-percent';
import { IndicatorIdentifier } from '../../core/indicator-identifier';
import { Shape } from '../../core/outputs/shape/shape';
import { RateOfChangePercentOutput } from './output';
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
// ROC% = ROC / 100, so the expected values are:
// -0.00546, -0.02109, -0.0553, -0.010367

describe('RateOfChangePercent', () => {

  it('should have correct output enum value', () => {
    expect(RateOfChangePercentOutput.RateOfChangePercentValue).toBe(0);
  });

  it('should return expected mnemonic for default components', () => {
    const rocp = new RateOfChangePercent({length: 14});
    expect(rocp.metadata().mnemonic).toBe('rocp(14)');
  });

  it('should return expected mnemonic with non-default bar component', () => {
    const rocp = new RateOfChangePercent({length: 14, barComponent: BarComponent.Median});
    expect(rocp.metadata().mnemonic).toBe('rocp(14, hl/2)');
    expect(rocp.metadata().description).toBe('Rate of Change Percent rocp(14, hl/2)');
  });

  it('should return expected mnemonic with non-default quote component', () => {
    const rocp = new RateOfChangePercent({length: 14, quoteComponent: QuoteComponent.Bid});
    expect(rocp.metadata().mnemonic).toBe('rocp(14, b)');
    expect(rocp.metadata().description).toBe('Rate of Change Percent rocp(14, b)');
  });

  it('should return expected mnemonic with non-default trade component', () => {
    const rocp = new RateOfChangePercent({length: 14, tradeComponent: TradeComponent.Volume});
    expect(rocp.metadata().mnemonic).toBe('rocp(14, v)');
    expect(rocp.metadata().description).toBe('Rate of Change Percent rocp(14, v)');
  });

  it('should return expected metadata', () => {
    const rocp = new RateOfChangePercent({length: 5});
    const meta = rocp.metadata();

    expect(meta.identifier).toBe(IndicatorIdentifier.RateOfChangePercent);
    expect(meta.mnemonic).toBe('rocp(5)');
    expect(meta.description).toBe('Rate of Change Percent rocp(5)');
    expect(meta.outputs.length).toBe(1);
    expect(meta.outputs[0].kind).toBe(RateOfChangePercentOutput.RateOfChangePercentValue);
    expect(meta.outputs[0].shape).toBe(Shape.Scalar);
    expect(meta.outputs[0].mnemonic).toBe('rocp(5)');
    expect(meta.outputs[0].description).toBe('Rate of Change Percent rocp(5)');
  });

  it('should throw if length is less than 1', () => {
    expect(() => { new RateOfChangePercent({length: 0}); }).toThrow();
    expect(() => { new RateOfChangePercent({length: -1}); }).toThrow();
  });

  it('should calculate expected output for length = 14', () => {
    const rocp = new RateOfChangePercent({length: 14});
    const eps = 1e-4;

    // Values from index=0 to index=13 are NaN (first 14 updates).
    for (let i = 0; i < 13; i++) {
      expect(rocp.update(input[i])).toBeNaN();
    }

    // From index 13 on, the indicator should return values starting at index 14.
    for (let i = 13; i < input.length; i++) {
      const act = rocp.update(input[i]);

      switch (i) {
        case 14:
          expect(Math.abs(act - (-0.00546))).toBeLessThan(eps);
          break;
        case 15:
          expect(Math.abs(act - (-0.02109))).toBeLessThan(eps);
          break;
        case 16:
          expect(Math.abs(act - (-0.0553))).toBeLessThan(eps);
          break;
        case 251:
          expect(Math.abs(act - (-0.010367))).toBeLessThan(eps);
          break;
      }
    }

    expect(rocp.update(Number.NaN)).toBeNaN();
  });

  it('should track primed state correctly for length = 1', () => {
    const rocp = new RateOfChangePercent({length: 1});
    expect(rocp.isPrimed()).toBe(false);

    rocp.update(input[0]);
    expect(rocp.isPrimed()).toBe(false);

    rocp.update(input[1]);
    expect(rocp.isPrimed()).toBe(true);
  });

  it('should track primed state correctly for length = 2', () => {
    const rocp = new RateOfChangePercent({length: 2});
    expect(rocp.isPrimed()).toBe(false);

    for (let i = 0; i < 2; i++) {
      rocp.update(input[i]);
      expect(rocp.isPrimed()).toBe(false);
    }

    rocp.update(input[2]);
    expect(rocp.isPrimed()).toBe(true);
  });

  it('should track primed state correctly for length = 5', () => {
    const rocp = new RateOfChangePercent({length: 5});
    expect(rocp.isPrimed()).toBe(false);

    for (let i = 0; i < 5; i++) {
      rocp.update(input[i]);
      expect(rocp.isPrimed()).toBe(false);
    }

    rocp.update(input[5]);
    expect(rocp.isPrimed()).toBe(true);
  });

  it('should track primed state correctly for length = 10', () => {
    const rocp = new RateOfChangePercent({length: 10});
    expect(rocp.isPrimed()).toBe(false);

    for (let i = 0; i < 10; i++) {
      rocp.update(input[i]);
      expect(rocp.isPrimed()).toBe(false);
    }

    rocp.update(input[10]);
    expect(rocp.isPrimed()).toBe(true);
  });
});
