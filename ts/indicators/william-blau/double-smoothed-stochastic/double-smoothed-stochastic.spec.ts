import { DoubleSmoothedStochastic } from './double-smoothed-stochastic';
import { IndicatorIdentifier } from '../../core/indicator-identifier';
import { Bar } from '../../../entities/bar';
import { Quote } from '../../../entities/quote';
import { Scalar } from '../../../entities/scalar';
import { Trade } from '../../../entities/trade';
import {
  testInput, testHigh, testLow,
  expectedDsQ5_R7_S3_G3, expectedSigQ5_R7_S3_G3,
  expectedDsQ2_R3_S15_G3, expectedSigQ2_R3_S15_G3,
  expectedDsQ5_R20_S5_G3, expectedSigQ5_R20_S5_G3,
  expectedDsQ5_R7_S3_G1, expectedSigQ5_R7_S3_G1,
  expectedDsQ2_R3_S15_G1, expectedSigQ2_R3_S15_G1,
  expectedDsQ1_R1_S1_G1, expectedSigQ1_R1_S1_G1,
  expectedDsQ1_R5_S5_G3, expectedSigQ1_R5_S5_G3,
  expectedDsQ8_R5_S3_G3, expectedSigQ8_R5_S3_G3,
  expectedDsQ21_R13_S4_G3, expectedSigQ21_R13_S4_G3,
  expectedDsQ5_R1_S1_G3, expectedSigQ5_R1_S1_G3,
  expectedDsQ3_R10_S10_G5, expectedSigQ3_R10_S10_G5,
  expectedDsQ34_R5_S5_G3, expectedSigQ34_R5_S5_G3,
  expectedDsQ2_R3_S15_G5, expectedSigQ2_R3_S15_G5,
  expectedDsQ10_R7_S3_G3, expectedSigQ10_R7_S3_G3,
} from './testdata';

interface Combo {
  name: string;
  q: number;
  r: number;
  s: number;
  g: number;
  dss: number[];
  signal: number[];
}

const combos: Combo[] = [
  { name: 'Q5_R7_S3_G3', q: 5, r: 7, s: 3, g: 3, dss: expectedDsQ5_R7_S3_G3, signal: expectedSigQ5_R7_S3_G3 },
  { name: 'Q2_R3_S15_G3', q: 2, r: 3, s: 15, g: 3, dss: expectedDsQ2_R3_S15_G3, signal: expectedSigQ2_R3_S15_G3 },
  { name: 'Q5_R20_S5_G3', q: 5, r: 20, s: 5, g: 3, dss: expectedDsQ5_R20_S5_G3, signal: expectedSigQ5_R20_S5_G3 },
  { name: 'Q5_R7_S3_G1', q: 5, r: 7, s: 3, g: 1, dss: expectedDsQ5_R7_S3_G1, signal: expectedSigQ5_R7_S3_G1 },
  { name: 'Q2_R3_S15_G1', q: 2, r: 3, s: 15, g: 1, dss: expectedDsQ2_R3_S15_G1, signal: expectedSigQ2_R3_S15_G1 },
  { name: 'Q1_R1_S1_G1', q: 1, r: 1, s: 1, g: 1, dss: expectedDsQ1_R1_S1_G1, signal: expectedSigQ1_R1_S1_G1 },
  { name: 'Q1_R5_S5_G3', q: 1, r: 5, s: 5, g: 3, dss: expectedDsQ1_R5_S5_G3, signal: expectedSigQ1_R5_S5_G3 },
  { name: 'Q8_R5_S3_G3', q: 8, r: 5, s: 3, g: 3, dss: expectedDsQ8_R5_S3_G3, signal: expectedSigQ8_R5_S3_G3 },
  { name: 'Q21_R13_S4_G3', q: 21, r: 13, s: 4, g: 3, dss: expectedDsQ21_R13_S4_G3, signal: expectedSigQ21_R13_S4_G3 },
  { name: 'Q5_R1_S1_G3', q: 5, r: 1, s: 1, g: 3, dss: expectedDsQ5_R1_S1_G3, signal: expectedSigQ5_R1_S1_G3 },
  { name: 'Q3_R10_S10_G5', q: 3, r: 10, s: 10, g: 5, dss: expectedDsQ3_R10_S10_G5, signal: expectedSigQ3_R10_S10_G5 },
  { name: 'Q34_R5_S5_G3', q: 34, r: 5, s: 5, g: 3, dss: expectedDsQ34_R5_S5_G3, signal: expectedSigQ34_R5_S5_G3 },
  { name: 'Q2_R3_S15_G5', q: 2, r: 3, s: 15, g: 5, dss: expectedDsQ2_R3_S15_G5, signal: expectedSigQ2_R3_S15_G5 },
  { name: 'Q10_R7_S3_G3', q: 10, r: 7, s: 3, g: 3, dss: expectedDsQ10_R7_S3_G3, signal: expectedSigQ10_R7_S3_G3 },
];

describe('DoubleSmoothedStochastic', () => {
  describe('reference data', () => {
    combos.forEach((combo) => {
      it(`matches the reference for ${combo.name}`, () => {
        const ind = new DoubleSmoothedStochastic({
          q: combo.q, r: combo.r, s: combo.s, g: combo.g,
        });

        for (let i = 0; i < testInput.length; i++) {
          const [dss, signal] = ind.update(testHigh[i], testLow[i], testInput[i]);

          if (isNaN(combo.dss[i])) {
            expect(dss).toBeNaN();
          } else {
            expect(dss).toBeCloseTo(combo.dss[i], 10);
          }

          if (isNaN(combo.signal[i])) {
            expect(signal).toBeNaN();
          } else {
            expect(signal).toBeCloseTo(combo.signal[i], 10);
          }
        }
      });
    });
  });

  describe('passthrough', () => {
    it('reduces to 100*(close-low)/(high-low) for the raw one-bar HLC index', () => {
      const ind = new DoubleSmoothedStochastic({ q: 1, r: 1, s: 1, g: 1 });

      expect(ind.update(12, 10, 12)).toEqual([100, 100]); // close at high
      expect(ind.update(12, 10, 10)).toEqual([0, 0]);     // close at low
      expect(ind.update(12, 10, 11)).toEqual([50, 50]);   // exact midpoint
      expect(ind.update(11, 11, 11)).toEqual([0, 0]);     // flat window -> division guard
    });
  });

  describe('signal', () => {
    it('averages an expanding window, then rolls over the last g values', () => {
      const ind = new DoubleSmoothedStochastic({ q: 1, r: 1, s: 1, g: 3 });

      expect(ind.update(12, 10, 12)[1]).toBeCloseTo(100, 10);     // dss 100
      expect(ind.update(12, 10, 10)[1]).toBeCloseTo(50, 10);      // dss 0
      expect(ind.update(12, 10, 11)[1]).toBeCloseTo(50, 10);      // dss 50
      expect(ind.update(12, 10, 11)[1]).toBeCloseTo(100 / 3, 10); // dss 50, 100 dropped
    });

    it('equals dss when g = 1', () => {
      const ind = new DoubleSmoothedStochastic({ q: 5, r: 7, s: 3, g: 1 });

      for (let i = 0; i < testInput.length; i++) {
        const [dss, signal] = ind.update(testHigh[i], testLow[i], testInput[i]);
        if (isNaN(dss)) {
          expect(signal).toBeNaN();
        } else {
          expect(signal).toBe(dss);
        }
      }
    });
  });

  describe('isPrimed', () => {
    it('is NaN and not primed for bars 0..q-2, primed from bar q-1', () => {
      const q = 5;
      const ind = new DoubleSmoothedStochastic({ q });
      expect(ind.isPrimed()).toBe(false);

      for (let i = 0; i < q - 1; i++) {
        const [dss, signal] = ind.update(testHigh[i], testLow[i], testInput[i]);
        expect(ind.isPrimed()).toBe(false);
        expect(dss).toBeNaN();
        expect(signal).toBeNaN();
      }

      for (let i = q - 1; i < testInput.length; i++) {
        ind.update(testHigh[i], testLow[i], testInput[i]);
        expect(ind.isPrimed()).toBe(true);
      }
    });

    it('is primed after the first bar when q = 1', () => {
      const ind = new DoubleSmoothedStochastic({ q: 1 });
      expect(ind.isPrimed()).toBe(false);
      ind.update(testHigh[0], testLow[0], testInput[0]);
      expect(ind.isPrimed()).toBe(true);
    });
  });

  describe('mnemonic', () => {
    it('formats the default mnemonic', () => {
      const ind = new DoubleSmoothedStochastic();
      expect(ind.metadata().mnemonic).toBe('dss(5,7,3,3)');
      expect(ind.metadata().description).toBe('Double Smoothed Stochastic dss(5,7,3,3)');
    });

    it('formats a custom mnemonic', () => {
      const ind = new DoubleSmoothedStochastic({ q: 2, r: 3, s: 15, g: 5 });
      expect(ind.metadata().mnemonic).toBe('dss(2,3,15,5)');
    });
  });

  describe('metadata', () => {
    it('reports the identifier and two outputs', () => {
      const ind = new DoubleSmoothedStochastic();
      const meta = ind.metadata();
      expect(meta.identifier).toBe(IndicatorIdentifier.DoubleSmoothedStochastic);
      expect(meta.outputs.length).toBe(2);
      expect(meta.outputs[0].mnemonic).toBe('dss(5,7,3,3) dss');
      expect(meta.outputs[0].description).toBe('Double Smoothed Stochastic dss(5,7,3,3) DSS');
      expect(meta.outputs[1].mnemonic).toBe('dss(5,7,3,3) signal');
      expect(meta.outputs[1].description).toBe('Double Smoothed Stochastic dss(5,7,3,3) signal');
    });
  });

  describe('updateBar', () => {
    it('returns [dss, signal] in order', () => {
      const ind = new DoubleSmoothedStochastic({ q: 5, r: 7, s: 3, g: 3 });
      let out: ReturnType<typeof ind.updateBar> = [];
      for (let i = 0; i < testInput.length; i++) {
        const bar = new Bar();
        bar.time = new Date(0);
        bar.open = 0;
        bar.high = testHigh[i];
        bar.low = testLow[i];
        bar.close = testInput[i];
        bar.volume = 0;
        out = ind.updateBar(bar);
      }
      const last = testInput.length - 1;
      expect(out.length).toBe(2);
      expect((out[0] as Scalar).value).toBeCloseTo(expectedDsQ5_R7_S3_G3[last], 10);
      expect((out[1] as Scalar).value).toBeCloseTo(expectedSigQ5_R7_S3_G3[last], 10);
    });
  });

  describe('updateScalar', () => {
    it('uses the value as the high, the low and the close', () => {
      const ind = new DoubleSmoothedStochastic({ q: 2, r: 1, s: 1, g: 1 });
      const s0 = new Scalar();
      s0.time = new Date(0);
      s0.value = 10;
      let out = ind.updateScalar(s0);
      expect((out[0] as Scalar).value).toBeNaN();
      expect((out[1] as Scalar).value).toBeNaN();

      const s1 = new Scalar();
      s1.time = new Date(0);
      s1.value = 12; // close at the 2-bar high
      out = ind.updateScalar(s1);
      expect((out[0] as Scalar).value).toBe(100);
      expect((out[1] as Scalar).value).toBe(100);
    });
  });

  describe('updateQuote', () => {
    it('uses the mid price as the high, the low and the close', () => {
      const ind = new DoubleSmoothedStochastic({ q: 2, r: 1, s: 1, g: 1 });
      const q0 = new Quote();
      q0.time = new Date(0);
      q0.bidPrice = 12;
      q0.askPrice = 14;
      ind.updateQuote(q0);

      const q1 = new Quote();
      q1.time = new Date(0);
      q1.bidPrice = 10;
      q1.askPrice = 12; // mid 11 is at the 2-bar low
      const out = ind.updateQuote(q1);
      expect((out[0] as Scalar).value).toBe(0);
      expect((out[1] as Scalar).value).toBe(0);
    });
  });

  describe('updateTrade', () => {
    it('uses the price as the high, the low and the close', () => {
      const ind = new DoubleSmoothedStochastic({ q: 2, r: 1, s: 1, g: 1 });
      const t0 = new Trade();
      t0.time = new Date(0);
      t0.price = 10;
      t0.volume = 1;
      ind.updateTrade(t0);

      const t1 = new Trade();
      t1.time = new Date(0);
      t1.price = 12;
      t1.volume = 1;
      const out = ind.updateTrade(t1);
      expect((out[0] as Scalar).value).toBe(100);
      expect((out[1] as Scalar).value).toBe(100);
    });
  });

  describe('invalid parameters', () => {
    it('throws when q < 1', () => {
      expect(() => new DoubleSmoothedStochastic({ q: 0 })).toThrowError();
    });
    it('throws when r < 1', () => {
      expect(() => new DoubleSmoothedStochastic({ r: 0 })).toThrowError();
    });
    it('throws when s < 1', () => {
      expect(() => new DoubleSmoothedStochastic({ s: 0 })).toThrowError();
    });
    it('throws when g < 1', () => {
      expect(() => new DoubleSmoothedStochastic({ g: 0 })).toThrowError();
    });
  });
});
