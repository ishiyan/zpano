import { } from 'jasmine';

import { InstantaneousTrendLine } from './instantaneous-trend-line';
import { InstantaneousTrendLineOutput } from './output';
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
    expectedTrendLine,
    expectedTrigger,
    inputHigh,
    inputLow,
} from './testdata';

/* eslint-disable max-len */
// Input data taken from TA-Lib Excel simulation (TALib data), test_iTrend.xsl,
// (high + low)/2 median price, D3...D254, 252 entries.

// Expected trend line values taken from Excel simulation, test_iTrend.xsl, F3...F254, 252 entries.
// Expected trigger line values taken from Excel simulation, test_iTrend.xsl, H3...H254, 252 entries.
// Input high price taken from TA-Lib Excel simulation, test_iTrend.xsl, B3...B254, 252 entries.
// Input low price taken from TA-Lib Excel simulation, test_iTrend.xsl, C3...C254, 252 entries.
describe('InstantaneousTrendLine', () => {
  const time = new Date(2021, 3, 1); // April 1, 2021
  const lprimed = 4; // First 4 values are NaN (primed on sample 5).
  const eps = 1e-8;

  it('should have correct output enum values', () => {
    expect(InstantaneousTrendLineOutput.Value).toBe(0);
    expect(InstantaneousTrendLineOutput.Trigger).toBe(1);
  });

  it('should return expected mnemonic with default smoothing factor parameters', () => {
    const itl = new InstantaneousTrendLine({ smoothingFactor: 0.07 });
    expect(itl.metadata().mnemonic).toBe('iTrend(28, hl/2)');
  });

  it('should return expected mnemonic with length-based parameters', () => {
    const itl = new InstantaneousTrendLine({ length: 10 });
    expect(itl.metadata().mnemonic).toBe('iTrend(10, hl/2)');
  });

  it('should return expected mnemonic with non-default bar component', () => {
    const itl = new InstantaneousTrendLine({
      smoothingFactor: 0.07,
      barComponent: BarComponent.Close,
    });
    // Close is the framework default, so it is omitted; but hl/2 is indicator default.
    expect(itl.metadata().mnemonic).toBe('iTrend(28)');
    expect(itl.metadata().description).toBe('Instantaneous Trend Line iTrend(28)');
  });

  it('should return expected mnemonic with non-default quote component', () => {
    const itl = new InstantaneousTrendLine({
      smoothingFactor: 0.07,
      quoteComponent: QuoteComponent.Bid,
    });
    expect(itl.metadata().mnemonic).toBe('iTrend(28, hl/2, b)');
  });

  it('should return expected mnemonic with non-default trade component', () => {
    const itl = new InstantaneousTrendLine({
      smoothingFactor: 0.07,
      tradeComponent: TradeComponent.Volume,
    });
    expect(itl.metadata().mnemonic).toBe('iTrend(28, hl/2, v)');
  });

  it('should return expected metadata', () => {
    const itl = new InstantaneousTrendLine({ smoothingFactor: 0.07 });
    const meta = itl.metadata();

    const mn = 'iTrend(28, hl/2)';
    const mnTrig = 'iTrendTrigger(28, hl/2)';
    const descr = 'Instantaneous Trend Line ';
    const descrTr = 'Instantaneous Trend Line trigger ';

    expect(meta.identifier).toBe(IndicatorIdentifier.InstantaneousTrendLine);
    expect(meta.mnemonic).toBe(mn);
    expect(meta.description).toBe(descr + mn);
    expect(meta.outputs.length).toBe(2);

    expect(meta.outputs[0].kind).toBe(InstantaneousTrendLineOutput.Value);
    expect(meta.outputs[0].shape).toBe(Shape.Scalar);
    expect(meta.outputs[0].mnemonic).toBe(mn);
    expect(meta.outputs[0].description).toBe(descr + mn);

    expect(meta.outputs[1].kind).toBe(InstantaneousTrendLineOutput.Trigger);
    expect(meta.outputs[1].shape).toBe(Shape.Scalar);
    expect(meta.outputs[1].mnemonic).toBe(mnTrig);
    expect(meta.outputs[1].description).toBe(descrTr + mnTrig);
  });

  it('should return expected metadata with length-based and non-default trade component', () => {
    const itl = new InstantaneousTrendLine({
      length: 3,
      tradeComponent: TradeComponent.Volume,
    });
    const meta = itl.metadata();
    expect(meta.mnemonic).toBe('iTrend(3, hl/2, v)');
    expect(meta.description).toBe('Instantaneous Trend Line iTrend(3, hl/2, v)');
  });

  it('should throw if the length is less than 1', () => {
    expect(() => { new InstantaneousTrendLine({ length: 0 }); }).toThrow();
    expect(() => { new InstantaneousTrendLine({ length: -1 }); }).toThrow();
    expect(() => { new InstantaneousTrendLine({ length: -8 }); }).toThrow();
  });

  it('should not throw for length = 1', () => {
    expect(() => { new InstantaneousTrendLine({ length: 1 }); }).not.toThrow();
  });

  it('should throw if smoothing factor is out of range [0, 1]', () => {
    expect(() => { new InstantaneousTrendLine({ smoothingFactor: -0.0001 }); }).toThrow();
    expect(() => { new InstantaneousTrendLine({ smoothingFactor: 1.0001 }); }).toThrow();
  });

  it('should not throw for smoothing factor boundaries', () => {
    expect(() => { new InstantaneousTrendLine({ smoothingFactor: 0 }); }).not.toThrow();
    expect(() => { new InstantaneousTrendLine({ smoothingFactor: 1 }); }).not.toThrow();
  });

  it('should calculate expected trend line values from reference implementation', () => {
    const itl = new InstantaneousTrendLine({ smoothingFactor: 0.07 });

    for (let i = 0; i < lprimed; i++) {
      expect(itl.update(input[i])).toBeNaN();
      expect(itl.isPrimed()).toBe(false);
    }

    for (let i = lprimed; i < input.length; i++) {
      const act = itl.update(input[i]);
      expect(itl.isPrimed()).toBe(true);
      expect(Math.abs(act - expectedTrendLine[i])).withContext(`TrendLine [${i}]: expected ${expectedTrendLine[i]}, actual ${act}`)
        .toBeLessThan(eps);
    }

    expect(itl.update(Number.NaN)).toBeNaN();
  });

  it('should calculate expected trigger line values from reference implementation via updateScalar', () => {
    const itl = new InstantaneousTrendLine({ smoothingFactor: 0.07 });

    for (let i = 0; i < lprimed; i++) {
      const output = itl.updateScalar(new Scalar({ time, value: input[i] }));
      expect(output.length).toBe(2);
      const s0 = output[0] as Scalar;
      const s1 = output[1] as Scalar;
      expect(Number.isNaN(s0.value)).toBe(true);
      expect(Number.isNaN(s1.value)).toBe(true);
    }

    for (let i = lprimed; i < input.length; i++) {
      const output = itl.updateScalar(new Scalar({ time, value: input[i] }));
      expect(output.length).toBe(2);

      const trendScalar = output[0] as Scalar;
      const triggerScalar = output[1] as Scalar;

      expect(Math.abs(trendScalar.value - expectedTrendLine[i]))
        .withContext(`TrendLine [${i}]: expected ${expectedTrendLine[i]}, actual ${trendScalar.value}`)
        .toBeLessThan(eps);

      expect(Math.abs(triggerScalar.value - expectedTrigger[i]))
        .withContext(`Trigger [${i}]: expected ${expectedTrigger[i]}, actual ${triggerScalar.value}`)
        .toBeLessThan(eps);
    }
  });

  it('should transition primed state correctly', () => {
    const itl = new InstantaneousTrendLine({ smoothingFactor: 0.07 });

    expect(itl.isPrimed()).toBe(false);

    for (let i = 0; i < lprimed; i++) {
      itl.update(input[i]);
      expect(itl.isPrimed()).toBe(false);
    }

    for (let i = lprimed; i < input.length; i++) {
      itl.update(input[i]);
      expect(itl.isPrimed()).toBe(true);
    }
  });

  it('should produce correct updateEntity output for bar', () => {
    // Default bar component for ITL is BarMedianPrice = (High+Low)/2.
    const itl = new InstantaneousTrendLine({ smoothingFactor: 0.07 });

    for (let i = 0; i < input.length; i++) {
      const bar = new Bar({ time, open: input[i], high: inputHigh[i], low: inputLow[i], close: input[i], volume: 0 });
      const output = itl.updateBar(bar);
      expect(output.length).toBe(2);

      const s0 = output[0] as Scalar;
      const s1 = output[1] as Scalar;

      expect(s0.time).toBe(time);
      expect(s1.time).toBe(time);

      if (Number.isNaN(expectedTrendLine[i])) {
        expect(Number.isNaN(s0.value)).toBe(true);
        expect(Number.isNaN(s1.value)).toBe(true);
      } else {
        expect(Math.abs(s0.value - expectedTrendLine[i]))
          .withContext(`Bar TrendLine [${i}]: expected ${expectedTrendLine[i]}, actual ${s0.value}`)
          .toBeLessThan(eps);
        expect(Math.abs(s1.value - expectedTrigger[i]))
          .withContext(`Bar Trigger [${i}]: expected ${expectedTrigger[i]}, actual ${s1.value}`)
          .toBeLessThan(eps);
      }
    }
  });

  it('should produce correct updateEntity output for quote', () => {
    // Use QuoteMidPrice = (Ask+Bid)/2, feeding high/low as ask/bid.
    const itl = new InstantaneousTrendLine({ smoothingFactor: 0.07 });

    for (let i = 0; i < input.length; i++) {
      const quote = new Quote({ time, bidPrice: inputLow[i], askPrice: inputHigh[i], bidSize: 0, askSize: 0 });
      const output = itl.updateQuote(quote);
      expect(output.length).toBe(2);

      const s0 = output[0] as Scalar;
      const s1 = output[1] as Scalar;

      expect(s0.time).toBe(time);
      expect(s1.time).toBe(time);

      if (Number.isNaN(expectedTrendLine[i])) {
        expect(Number.isNaN(s0.value)).toBe(true);
        expect(Number.isNaN(s1.value)).toBe(true);
      } else {
        expect(Math.abs(s0.value - expectedTrendLine[i]))
          .withContext(`Quote TrendLine [${i}]: expected ${expectedTrendLine[i]}, actual ${s0.value}`)
          .toBeLessThan(eps);
        expect(Math.abs(s1.value - expectedTrigger[i]))
          .withContext(`Quote Trigger [${i}]: expected ${expectedTrigger[i]}, actual ${s1.value}`)
          .toBeLessThan(eps);
      }
    }
  });

  it('should produce correct updateEntity output for trade', () => {
    const itl = new InstantaneousTrendLine({ smoothingFactor: 0.07 });

    for (let i = 0; i < input.length; i++) {
      const trade = new Trade({ time, price: input[i], volume: 0 });
      const output = itl.updateTrade(trade);
      expect(output.length).toBe(2);

      const s0 = output[0] as Scalar;
      const s1 = output[1] as Scalar;

      expect(s0.time).toBe(time);
      expect(s1.time).toBe(time);

      if (Number.isNaN(expectedTrendLine[i])) {
        expect(Number.isNaN(s0.value)).toBe(true);
        expect(Number.isNaN(s1.value)).toBe(true);
      } else {
        expect(Math.abs(s0.value - expectedTrendLine[i]))
          .withContext(`Trade TrendLine [${i}]: expected ${expectedTrendLine[i]}, actual ${s0.value}`)
          .toBeLessThan(eps);
        expect(Math.abs(s1.value - expectedTrigger[i]))
          .withContext(`Trade Trigger [${i}]: expected ${expectedTrigger[i]}, actual ${s1.value}`)
          .toBeLessThan(eps);
      }
    }
  });

  it('should return NaN for trigger when trend line is NaN (not primed)', () => {
    const itl = new InstantaneousTrendLine({ smoothingFactor: 0.07 });

    const s = new Scalar({ time, value: 100 });
    const output = itl.updateScalar(s);
    expect(output.length).toBe(2);

    const s0 = output[0] as Scalar;
    const s1 = output[1] as Scalar;
    expect(Number.isNaN(s0.value)).toBe(true);
    expect(Number.isNaN(s1.value)).toBe(true);
  });
});
