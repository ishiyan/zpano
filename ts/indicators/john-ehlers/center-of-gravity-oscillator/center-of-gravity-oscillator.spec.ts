import { } from 'jasmine';

import { CenterOfGravityOscillator } from './center-of-gravity-oscillator';
import { CenterOfGravityOscillatorOutput } from './output';
import { IndicatorIdentifier } from '../../core/indicator-identifier';
import { Shape } from '../../core/outputs/shape/shape';
import { BarComponent } from '../../../entities/bar-component';
import { QuoteComponent } from '../../../entities/quote-component';
import { TradeComponent } from '../../../entities/trade-component';
import { Scalar } from '../../../entities/scalar';
import { Bar } from '../../../entities/bar';
import { Quote } from '../../../entities/quote';
import { Trade } from '../../../entities/trade';
import {
    input,
    expectedCog,
    expectedTrigger,
    inputHigh,
    inputLow,
} from './testdata';

/* eslint-disable max-len */
// Input data taken from TA-Lib Excel simulation (TALib data), test_Cog.xsl,
// (high + low)/2 median price, D3...D254, 252 entries.

// Expected COG values taken from Excel simulation, test_Cog.xsl, F3...F254, 252 entries.
// Expected trigger values taken from Excel simulation, test_Cog.xsl, H3...H254, 252 entries.
// Input high price taken from TA-Lib Excel simulation, test_Cog.xsl, B3...B254, 252 entries.
// Input low price taken from TA-Lib Excel simulation, test_Cog.xsl, C3...C254, 252 entries.
describe('CenterOfGravityOscillator', () => {
  const time = new Date(2021, 3, 1); // April 1, 2021
  const l = 10;
  const lprimed = 10; // First 10 values are NaN.
  const eps = 1e-8;

  it('should have correct output enum values', () => {
    expect(CenterOfGravityOscillatorOutput.Value).toBe(0);
    expect(CenterOfGravityOscillatorOutput.Trigger).toBe(1);
  });

  it('should return expected mnemonic with default parameters', () => {
    const cog = new CenterOfGravityOscillator({ length: 10 });
    expect(cog.metadata().mnemonic).toBe('cog(10, hl/2)');
  });

  it('should return expected mnemonic with different length', () => {
    const cog = new CenterOfGravityOscillator({ length: 20 });
    expect(cog.metadata().mnemonic).toBe('cog(20, hl/2)');
  });

  it('should return expected mnemonic with non-default bar component', () => {
    const cog = new CenterOfGravityOscillator({
      length: 10,
      barComponent: BarComponent.Close,
    });
    // Close is the framework default, so it is omitted from the mnemonic.
    expect(cog.metadata().mnemonic).toBe('cog(10)');
    expect(cog.metadata().description).toBe('Center of Gravity oscillator cog(10)');
  });

  it('should return expected mnemonic with non-default quote component', () => {
    const cog = new CenterOfGravityOscillator({
      length: 10,
      quoteComponent: QuoteComponent.Bid,
    });
    expect(cog.metadata().mnemonic).toBe('cog(10, hl/2, b)');
  });

  it('should return expected mnemonic with non-default trade component', () => {
    const cog = new CenterOfGravityOscillator({
      length: 10,
      tradeComponent: TradeComponent.Volume,
    });
    expect(cog.metadata().mnemonic).toBe('cog(10, hl/2, v)');
  });

  it('should return expected metadata', () => {
    const cog = new CenterOfGravityOscillator({ length: 10 });
    const meta = cog.metadata();

    const mn = 'cog(10, hl/2)';
    const mnTrig = 'cogTrig(10, hl/2)';
    const descr = 'Center of Gravity oscillator ';
    const descrTrig = 'Center of Gravity trigger ';

    expect(meta.identifier).toBe(IndicatorIdentifier.CenterOfGravityOscillator);
    expect(meta.mnemonic).toBe(mn);
    expect(meta.description).toBe(descr + mn);
    expect(meta.outputs.length).toBe(2);

    expect(meta.outputs[0].kind).toBe(CenterOfGravityOscillatorOutput.Value);
    expect(meta.outputs[0].shape).toBe(Shape.Scalar);
    expect(meta.outputs[0].mnemonic).toBe(mn);
    expect(meta.outputs[0].description).toBe(descr + mn);

    expect(meta.outputs[1].kind).toBe(CenterOfGravityOscillatorOutput.Trigger);
    expect(meta.outputs[1].shape).toBe(Shape.Scalar);
    expect(meta.outputs[1].mnemonic).toBe(mnTrig);
    expect(meta.outputs[1].description).toBe(descrTrig + mnTrig);
  });

  it('should throw if the length is less than 1', () => {
    expect(() => { new CenterOfGravityOscillator({ length: 0 }); }).toThrow();
    expect(() => { new CenterOfGravityOscillator({ length: -1 }); }).toThrow();
    expect(() => { new CenterOfGravityOscillator({ length: -8 }); }).toThrow();
  });

  it('should not throw for length = 1', () => {
    expect(() => { new CenterOfGravityOscillator({ length: 1 }); }).not.toThrow();
  });

  it('should calculate expected COG values from reference implementation', () => {
    const cog = new CenterOfGravityOscillator({ length: l });

    for (let i = 0; i < lprimed; i++) {
      expect(cog.update(input[i])).toBeNaN();
      expect(cog.isPrimed()).toBe(false);
    }

    for (let i = lprimed; i < input.length; i++) {
      const act = cog.update(input[i]);
      expect(cog.isPrimed()).toBe(true);
      expect(Math.abs(act - expectedCog[i])).withContext(`COG [${i}]: expected ${expectedCog[i]}, actual ${act}`)
        .toBeLessThan(eps);
    }

    expect(cog.update(Number.NaN)).toBeNaN();
  });

  it('should calculate expected trigger values from reference implementation via updateScalar', () => {
    const cog = new CenterOfGravityOscillator({ length: l });

    for (let i = 0; i < lprimed; i++) {
      const output = cog.updateScalar(new Scalar({ time, value: input[i] }));
      expect(output.length).toBe(2);
      const s0 = output[0] as Scalar;
      const s1 = output[1] as Scalar;
      expect(Number.isNaN(s0.value)).toBe(true);
      expect(Number.isNaN(s1.value)).toBe(true);
    }

    for (let i = lprimed; i < input.length; i++) {
      const output = cog.updateScalar(new Scalar({ time, value: input[i] }));
      expect(output.length).toBe(2);

      const cogScalar = output[0] as Scalar;
      const trigScalar = output[1] as Scalar;

      expect(Math.abs(cogScalar.value - expectedCog[i]))
        .withContext(`COG [${i}]: expected ${expectedCog[i]}, actual ${cogScalar.value}`)
        .toBeLessThan(eps);

      expect(Math.abs(trigScalar.value - expectedTrigger[i]))
        .withContext(`Trigger [${i}]: expected ${expectedTrigger[i]}, actual ${trigScalar.value}`)
        .toBeLessThan(eps);
    }
  });

  it('should transition primed state correctly', () => {
    const cog = new CenterOfGravityOscillator({ length: l });

    expect(cog.isPrimed()).toBe(false);

    for (let i = 0; i < lprimed; i++) {
      cog.update(input[i]);
      expect(cog.isPrimed()).toBe(false);
    }

    for (let i = lprimed; i < input.length; i++) {
      cog.update(input[i]);
      expect(cog.isPrimed()).toBe(true);
    }
  });

  it('should produce correct updateEntity output for bar', () => {
    // Default bar component for CoG is BarMedianPrice = (High+Low)/2.
    const cog = new CenterOfGravityOscillator({ length: l });

    for (let i = 0; i < input.length; i++) {
      const bar = new Bar({ time, open: input[i], high: inputHigh[i], low: inputLow[i], close: input[i], volume: 0 });
      const output = cog.updateBar(bar);
      expect(output.length).toBe(2);

      const s0 = output[0] as Scalar;
      const s1 = output[1] as Scalar;

      expect(s0.time).toBe(time);
      expect(s1.time).toBe(time);

      if (Number.isNaN(expectedCog[i])) {
        expect(Number.isNaN(s0.value)).toBe(true);
        expect(Number.isNaN(s1.value)).toBe(true);
      } else {
        expect(Math.abs(s0.value - expectedCog[i]))
          .withContext(`Bar COG [${i}]: expected ${expectedCog[i]}, actual ${s0.value}`)
          .toBeLessThan(eps);
        expect(Math.abs(s1.value - expectedTrigger[i]))
          .withContext(`Bar Trigger [${i}]: expected ${expectedTrigger[i]}, actual ${s1.value}`)
          .toBeLessThan(eps);
      }
    }
  });

  it('should produce correct updateEntity output for quote', () => {
    // Use QuoteMidPrice = (Ask+Bid)/2, feeding high/low as ask/bid.
    const cog = new CenterOfGravityOscillator({ length: l });

    for (let i = 0; i < input.length; i++) {
      const quote = new Quote({ time, bidPrice: inputLow[i], askPrice: inputHigh[i], bidSize: 0, askSize: 0 });
      const output = cog.updateQuote(quote);
      expect(output.length).toBe(2);

      const s0 = output[0] as Scalar;
      const s1 = output[1] as Scalar;

      expect(s0.time).toBe(time);
      expect(s1.time).toBe(time);

      if (Number.isNaN(expectedCog[i])) {
        expect(Number.isNaN(s0.value)).toBe(true);
        expect(Number.isNaN(s1.value)).toBe(true);
      } else {
        expect(Math.abs(s0.value - expectedCog[i]))
          .withContext(`Quote COG [${i}]: expected ${expectedCog[i]}, actual ${s0.value}`)
          .toBeLessThan(eps);
        expect(Math.abs(s1.value - expectedTrigger[i]))
          .withContext(`Quote Trigger [${i}]: expected ${expectedTrigger[i]}, actual ${s1.value}`)
          .toBeLessThan(eps);
      }
    }
  });

  it('should produce correct updateEntity output for trade', () => {
    const cog = new CenterOfGravityOscillator({ length: l });

    for (let i = 0; i < input.length; i++) {
      const trade = new Trade({ time, price: input[i], volume: 0 });
      const output = cog.updateTrade(trade);
      expect(output.length).toBe(2);

      const s0 = output[0] as Scalar;
      const s1 = output[1] as Scalar;

      expect(s0.time).toBe(time);
      expect(s1.time).toBe(time);

      if (Number.isNaN(expectedCog[i])) {
        expect(Number.isNaN(s0.value)).toBe(true);
        expect(Number.isNaN(s1.value)).toBe(true);
      } else {
        expect(Math.abs(s0.value - expectedCog[i]))
          .withContext(`Trade COG [${i}]: expected ${expectedCog[i]}, actual ${s0.value}`)
          .toBeLessThan(eps);
        expect(Math.abs(s1.value - expectedTrigger[i]))
          .withContext(`Trade Trigger [${i}]: expected ${expectedTrigger[i]}, actual ${s1.value}`)
          .toBeLessThan(eps);
      }
    }
  });

  it('should return NaN for trigger when cog is NaN (not primed)', () => {
    const cog = new CenterOfGravityOscillator({ length: l });

    const s = new Scalar({ time, value: 100 });
    const output = cog.updateScalar(s);
    expect(output.length).toBe(2);

    const s0 = output[0] as Scalar;
    const s1 = output[1] as Scalar;
    expect(Number.isNaN(s0.value)).toBe(true);
    expect(Number.isNaN(s1.value)).toBe(true);
  });
});
