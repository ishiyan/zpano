import { } from 'jasmine';

import { CyberCycle } from './cyber-cycle';
import { CyberCycleOutput } from './output';
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
    expectedCycle,
    expectedSignal,
    inputHigh,
    inputLow,
} from './testdata';

/* eslint-disable max-len */
// Input data taken from TA-Lib Excel simulation (TALib data), test_iTrend.xsl,
// (high + low)/2 median price, D3...D254, 252 entries.

// Expected cycle (Value) values taken from Excel simulation, test_iTrend.xsl, L3...L254, 252 entries.
// Expected signal values taken from Excel simulation, test_iTrend.xsl, N3...N254, 252 entries.
// Input high price taken from TA-Lib Excel simulation, test_iTrend.xsl, B3...B254, 252 entries.
// Input low price taken from TA-Lib Excel simulation, test_iTrend.xsl, C3...C254, 252 entries.
describe('CyberCycle', () => {
  const time = new Date(2021, 3, 1); // April 1, 2021
  const lprimed = 7; // First 7 values are NaN (primed on sample 8).
  const eps = 1e-8;

  it('should have correct output enum values', () => {
    expect(CyberCycleOutput.Value).toBe(0);
    expect(CyberCycleOutput.Signal).toBe(1);
  });

  it('should return expected mnemonic with default smoothing factor parameters', () => {
    const cc = new CyberCycle({ smoothingFactor: 0.07, signalLag: 9 });
    expect(cc.metadata().mnemonic).toBe('cc(28, hl/2)');
  });

  it('should return expected mnemonic with length-based parameters', () => {
    const cc = new CyberCycle({ length: 10, signalLag: 9 });
    expect(cc.metadata().mnemonic).toBe('cc(10, hl/2)');
  });

  it('should return expected mnemonic with non-default bar component', () => {
    const cc = new CyberCycle({
      smoothingFactor: 0.07,
      signalLag: 9,
      barComponent: BarComponent.Close,
    });
    // Close is the framework default, so it is omitted; but hl/2 is indicator default.
    expect(cc.metadata().mnemonic).toBe('cc(28)');
    expect(cc.metadata().description).toBe('Cyber Cycle cc(28)');
  });

  it('should return expected mnemonic with non-default quote component', () => {
    const cc = new CyberCycle({
      smoothingFactor: 0.07,
      signalLag: 9,
      quoteComponent: QuoteComponent.Bid,
    });
    expect(cc.metadata().mnemonic).toBe('cc(28, hl/2, b)');
  });

  it('should return expected mnemonic with non-default trade component', () => {
    const cc = new CyberCycle({
      smoothingFactor: 0.07,
      signalLag: 9,
      tradeComponent: TradeComponent.Volume,
    });
    expect(cc.metadata().mnemonic).toBe('cc(28, hl/2, v)');
  });

  it('should return expected metadata', () => {
    const cc = new CyberCycle({ smoothingFactor: 0.07, signalLag: 9 });
    const meta = cc.metadata();

    const mn = 'cc(28, hl/2)';
    const mnSignal = 'ccSignal(28, hl/2)';
    const descr = 'Cyber Cycle ';
    const descrSignal = 'Cyber Cycle signal ';

    expect(meta.identifier).toBe(IndicatorIdentifier.CyberCycle);
    expect(meta.mnemonic).toBe(mn);
    expect(meta.description).toBe(descr + mn);
    expect(meta.outputs.length).toBe(2);

    expect(meta.outputs[0].kind).toBe(CyberCycleOutput.Value);
    expect(meta.outputs[0].shape).toBe(Shape.Scalar);
    expect(meta.outputs[0].mnemonic).toBe(mn);
    expect(meta.outputs[0].description).toBe(descr + mn);

    expect(meta.outputs[1].kind).toBe(CyberCycleOutput.Signal);
    expect(meta.outputs[1].shape).toBe(Shape.Scalar);
    expect(meta.outputs[1].mnemonic).toBe(mnSignal);
    expect(meta.outputs[1].description).toBe(descrSignal + mnSignal);
  });

  it('should return expected metadata with length-based and non-default trade component', () => {
    const cc = new CyberCycle({
      length: 3,
      signalLag: 2,
      tradeComponent: TradeComponent.Volume,
    });
    const meta = cc.metadata();
    expect(meta.mnemonic).toBe('cc(3, hl/2, v)');
    expect(meta.description).toBe('Cyber Cycle cc(3, hl/2, v)');
  });

  it('should throw if the length is less than 1', () => {
    expect(() => { new CyberCycle({ length: 0, signalLag: 9 }); }).toThrow();
    expect(() => { new CyberCycle({ length: -1, signalLag: 9 }); }).toThrow();
    expect(() => { new CyberCycle({ length: -8, signalLag: 9 }); }).toThrow();
  });

  it('should not throw for length = 1', () => {
    expect(() => { new CyberCycle({ length: 1, signalLag: 1 }); }).not.toThrow();
  });

  it('should throw if smoothing factor is out of range [0, 1]', () => {
    expect(() => { new CyberCycle({ smoothingFactor: -0.0001, signalLag: 9 }); }).toThrow();
    expect(() => { new CyberCycle({ smoothingFactor: 1.0001, signalLag: 9 }); }).toThrow();
  });

  it('should not throw for smoothing factor boundaries', () => {
    expect(() => { new CyberCycle({ smoothingFactor: 0, signalLag: 9 }); }).not.toThrow();
    expect(() => { new CyberCycle({ smoothingFactor: 1, signalLag: 9 }); }).not.toThrow();
  });

  it('should throw if signal lag is less than 1', () => {
    expect(() => { new CyberCycle({ smoothingFactor: 0.07, signalLag: 0 }); }).toThrow();
    expect(() => { new CyberCycle({ smoothingFactor: 0.07, signalLag: -8 }); }).toThrow();
    expect(() => { new CyberCycle({ length: 10, signalLag: 0 }); }).toThrow();
    expect(() => { new CyberCycle({ length: 10, signalLag: -8 }); }).toThrow();
  });

  it('should calculate expected cycle values from reference implementation', () => {
    const cc = new CyberCycle({ smoothingFactor: 0.07, signalLag: 9 });

    for (let i = 0; i < lprimed; i++) {
      expect(cc.update(input[i])).toBeNaN();
      expect(cc.isPrimed()).toBe(false);
    }

    for (let i = lprimed; i < input.length; i++) {
      const act = cc.update(input[i]);
      expect(cc.isPrimed()).toBe(true);
      expect(Math.abs(act - expectedCycle[i])).withContext(`Cycle [${i}]: expected ${expectedCycle[i]}, actual ${act}`)
        .toBeLessThan(eps);
    }

    expect(cc.update(Number.NaN)).toBeNaN();
  });

  it('should calculate expected signal values from reference implementation via updateScalar', () => {
    const cc = new CyberCycle({ smoothingFactor: 0.07, signalLag: 9 });

    for (let i = 0; i < lprimed; i++) {
      const output = cc.updateScalar(new Scalar({ time, value: input[i] }));
      expect(output.length).toBe(2);
      const s0 = output[0] as Scalar;
      const s1 = output[1] as Scalar;
      expect(Number.isNaN(s0.value)).toBe(true);
      expect(Number.isNaN(s1.value)).toBe(true);
    }

    for (let i = lprimed; i < input.length; i++) {
      const output = cc.updateScalar(new Scalar({ time, value: input[i] }));
      expect(output.length).toBe(2);

      const cycleScalar = output[0] as Scalar;
      const signalScalar = output[1] as Scalar;

      expect(Math.abs(cycleScalar.value - expectedCycle[i]))
        .withContext(`Cycle [${i}]: expected ${expectedCycle[i]}, actual ${cycleScalar.value}`)
        .toBeLessThan(eps);

      expect(Math.abs(signalScalar.value - expectedSignal[i]))
        .withContext(`Signal [${i}]: expected ${expectedSignal[i]}, actual ${signalScalar.value}`)
        .toBeLessThan(eps);
    }
  });

  it('should transition primed state correctly', () => {
    const cc = new CyberCycle({ smoothingFactor: 0.07, signalLag: 9 });

    expect(cc.isPrimed()).toBe(false);

    for (let i = 0; i < lprimed; i++) {
      cc.update(input[i]);
      expect(cc.isPrimed()).toBe(false);
    }

    for (let i = lprimed; i < input.length; i++) {
      cc.update(input[i]);
      expect(cc.isPrimed()).toBe(true);
    }
  });

  it('should produce correct updateEntity output for bar', () => {
    // Default bar component for CyberCycle is BarMedianPrice = (High+Low)/2.
    const cc = new CyberCycle({ smoothingFactor: 0.07, signalLag: 9 });

    for (let i = 0; i < input.length; i++) {
      const bar = new Bar({ time, open: input[i], high: inputHigh[i], low: inputLow[i], close: input[i], volume: 0 });
      const output = cc.updateBar(bar);
      expect(output.length).toBe(2);

      const s0 = output[0] as Scalar;
      const s1 = output[1] as Scalar;

      expect(s0.time).toBe(time);
      expect(s1.time).toBe(time);

      if (Number.isNaN(expectedCycle[i])) {
        expect(Number.isNaN(s0.value)).toBe(true);
        expect(Number.isNaN(s1.value)).toBe(true);
      } else {
        expect(Math.abs(s0.value - expectedCycle[i]))
          .withContext(`Bar Cycle [${i}]: expected ${expectedCycle[i]}, actual ${s0.value}`)
          .toBeLessThan(eps);
        expect(Math.abs(s1.value - expectedSignal[i]))
          .withContext(`Bar Signal [${i}]: expected ${expectedSignal[i]}, actual ${s1.value}`)
          .toBeLessThan(eps);
      }
    }
  });

  it('should produce correct updateEntity output for quote', () => {
    // Use QuoteMidPrice = (Ask+Bid)/2, feeding high/low as ask/bid.
    const cc = new CyberCycle({ smoothingFactor: 0.07, signalLag: 9 });

    for (let i = 0; i < input.length; i++) {
      const quote = new Quote({ time, bidPrice: inputLow[i], askPrice: inputHigh[i], bidSize: 0, askSize: 0 });
      const output = cc.updateQuote(quote);
      expect(output.length).toBe(2);

      const s0 = output[0] as Scalar;
      const s1 = output[1] as Scalar;

      expect(s0.time).toBe(time);
      expect(s1.time).toBe(time);

      if (Number.isNaN(expectedCycle[i])) {
        expect(Number.isNaN(s0.value)).toBe(true);
        expect(Number.isNaN(s1.value)).toBe(true);
      } else {
        expect(Math.abs(s0.value - expectedCycle[i]))
          .withContext(`Quote Cycle [${i}]: expected ${expectedCycle[i]}, actual ${s0.value}`)
          .toBeLessThan(eps);
        expect(Math.abs(s1.value - expectedSignal[i]))
          .withContext(`Quote Signal [${i}]: expected ${expectedSignal[i]}, actual ${s1.value}`)
          .toBeLessThan(eps);
      }
    }
  });

  it('should produce correct updateEntity output for trade', () => {
    const cc = new CyberCycle({ smoothingFactor: 0.07, signalLag: 9 });

    for (let i = 0; i < input.length; i++) {
      const trade = new Trade({ time, price: input[i], volume: 0 });
      const output = cc.updateTrade(trade);
      expect(output.length).toBe(2);

      const s0 = output[0] as Scalar;
      const s1 = output[1] as Scalar;

      expect(s0.time).toBe(time);
      expect(s1.time).toBe(time);

      if (Number.isNaN(expectedCycle[i])) {
        expect(Number.isNaN(s0.value)).toBe(true);
        expect(Number.isNaN(s1.value)).toBe(true);
      } else {
        expect(Math.abs(s0.value - expectedCycle[i]))
          .withContext(`Trade Cycle [${i}]: expected ${expectedCycle[i]}, actual ${s0.value}`)
          .toBeLessThan(eps);
        expect(Math.abs(s1.value - expectedSignal[i]))
          .withContext(`Trade Signal [${i}]: expected ${expectedSignal[i]}, actual ${s1.value}`)
          .toBeLessThan(eps);
      }
    }
  });

  it('should return NaN for signal when cycle is NaN (not primed)', () => {
    const cc = new CyberCycle({ smoothingFactor: 0.07, signalLag: 9 });

    const s = new Scalar({ time, value: 100 });
    const output = cc.updateScalar(s);
    expect(output.length).toBe(2);

    const s0 = output[0] as Scalar;
    const s1 = output[1] as Scalar;
    expect(Number.isNaN(s0.value)).toBe(true);
    expect(Number.isNaN(s1.value)).toBe(true);
  });
});
