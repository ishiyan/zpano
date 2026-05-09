import { } from 'jasmine';

import { Momentum } from './momentum';
import { IndicatorIdentifier } from '../../core/indicator-identifier';
import { Shape } from '../../core/outputs/shape/shape';
import { MomentumOutput } from './output';
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

describe('Momentum', () => {

  it('should have correct output enum value', () => {
    expect(MomentumOutput.MomentumValue).toBe(0);
  });

  it('should return expected mnemonic for default components', () => {
    const mom = new Momentum({length: 14});
    expect(mom.metadata().mnemonic).toBe('mom(14)');
  });

  it('should return expected mnemonic with non-default bar component', () => {
    const mom = new Momentum({length: 14, barComponent: BarComponent.Median});
    expect(mom.metadata().mnemonic).toBe('mom(14, hl/2)');
    expect(mom.metadata().description).toBe('Momentum mom(14, hl/2)');
  });

  it('should return expected mnemonic with non-default quote component', () => {
    const mom = new Momentum({length: 14, quoteComponent: QuoteComponent.Bid});
    expect(mom.metadata().mnemonic).toBe('mom(14, b)');
    expect(mom.metadata().description).toBe('Momentum mom(14, b)');
  });

  it('should return expected mnemonic with non-default trade component', () => {
    const mom = new Momentum({length: 14, tradeComponent: TradeComponent.Volume});
    expect(mom.metadata().mnemonic).toBe('mom(14, v)');
    expect(mom.metadata().description).toBe('Momentum mom(14, v)');
  });

  it('should return expected metadata', () => {
    const mom = new Momentum({length: 5});
    const meta = mom.metadata();

    expect(meta.identifier).toBe(IndicatorIdentifier.Momentum);
    expect(meta.mnemonic).toBe('mom(5)');
    expect(meta.description).toBe('Momentum mom(5)');
    expect(meta.outputs.length).toBe(1);
    expect(meta.outputs[0].kind).toBe(MomentumOutput.MomentumValue);
    expect(meta.outputs[0].shape).toBe(Shape.Scalar);
    expect(meta.outputs[0].mnemonic).toBe('mom(5)');
    expect(meta.outputs[0].description).toBe('Momentum mom(5)');
  });

  it('should throw if length is less than 1', () => {
    expect(() => { new Momentum({length: 0}); }).toThrow();
    expect(() => { new Momentum({length: -1}); }).toThrow();
  });

  it('should calculate expected output for length = 14', () => {
    const mom = new Momentum({length: 14});
    const eps = 1e-13;

    // Values from index=0 to index=13 are NaN (first 14 updates).
    for (let i = 0; i < 13; i++) {
      expect(mom.update(input[i])).toBeNaN();
    }

    // From index 13 on, the indicator should return values starting at index 14.
    for (let i = 13; i < input.length; i++) {
      const act = mom.update(input[i]);

      switch (i) {
        case 14:
          expect(Math.abs(act - (-0.50))).toBeLessThan(eps);
          break;
        case 15:
          expect(Math.abs(act - (-2.00))).toBeLessThan(eps);
          break;
        case 16:
          expect(Math.abs(act - (-5.22))).toBeLessThan(eps);
          break;
        case 251:
          expect(Math.abs(act - (-1.13))).toBeLessThan(eps);
          break;
      }
    }

    expect(mom.update(Number.NaN)).toBeNaN();
  });

  it('should track primed state correctly for length = 1', () => {
    const mom = new Momentum({length: 1});
    expect(mom.isPrimed()).toBe(false);

    mom.update(input[0]);
    expect(mom.isPrimed()).toBe(false);

    mom.update(input[1]);
    expect(mom.isPrimed()).toBe(true);
  });

  it('should track primed state correctly for length = 2', () => {
    const mom = new Momentum({length: 2});
    expect(mom.isPrimed()).toBe(false);

    for (let i = 0; i < 2; i++) {
      mom.update(input[i]);
      expect(mom.isPrimed()).toBe(false);
    }

    mom.update(input[2]);
    expect(mom.isPrimed()).toBe(true);
  });

  it('should track primed state correctly for length = 5', () => {
    const mom = new Momentum({length: 5});
    expect(mom.isPrimed()).toBe(false);

    for (let i = 0; i < 5; i++) {
      mom.update(input[i]);
      expect(mom.isPrimed()).toBe(false);
    }

    mom.update(input[5]);
    expect(mom.isPrimed()).toBe(true);
  });

  it('should track primed state correctly for length = 10', () => {
    const mom = new Momentum({length: 10});
    expect(mom.isPrimed()).toBe(false);

    for (let i = 0; i < 10; i++) {
      mom.update(input[i]);
      expect(mom.isPrimed()).toBe(false);
    }

    mom.update(input[10]);
    expect(mom.isPrimed()).toBe(true);
  });
});
