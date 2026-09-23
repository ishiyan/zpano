import { StochasticMomentumIndex } from './stochastic-momentum-index';
import { IndicatorIdentifier } from '../../core/indicator-identifier';
import { Bar } from '../../../entities/bar';
import { Quote } from '../../../entities/quote';
import { Scalar } from '../../../entities/scalar';
import { Trade } from '../../../entities/trade';
import {
  testInput, testHigh, testLow,
  expectedQ5_R20_S5_U3, expectedQ5_R20_S5_U3_SIG_UL3,
  expectedQ13_R25_S2_U1, expectedQ13_R25_S2_U1_SIG_UL3,
  expectedQ2_R20_S20_U1, expectedQ2_R20_S20_U1_SIG_UL3,
  expectedQ13_R25_S2_U3, expectedQ13_R25_S2_U3_SIG_UL3,
  expectedQ5_R20_S5_U1, expectedQ5_R20_S5_U1_SIG_UL3,
  expectedQ8_R5_S3_U1, expectedQ8_R5_S3_U1_SIG_UL3,
  expectedQ21_R13_S4_U1, expectedQ21_R13_S4_U1_SIG_UL3,
  expectedQ1_R20_S5_U3, expectedQ1_R20_S5_U3_SIG_UL3,
  expectedQ1_R40_S20_U1, expectedQ1_R40_S20_U1_SIG_UL3,
  expectedQ1_R100_S20_U1, expectedQ1_R100_S20_U1_SIG_UL3,
  expectedQ1_R1_S1_U1, expectedQ1_R1_S1_U1_SIG_UL3,
  expectedQ5_R1_S1_U1, expectedQ5_R1_S1_U1_SIG_UL3,
  expectedQ3_R10_S10_U1, expectedQ3_R10_S10_U1_SIG_UL3,
  expectedQ34_R5_S5_U1, expectedQ34_R5_S5_U1_SIG_UL3,
  expectedQ2_R2_S2_U2, expectedQ2_R2_S2_U2_SIG_UL3,
  expectedQ50_R20_S5_U3, expectedQ50_R20_S5_U3_SIG_UL3,
} from './testdata';

// Signal-line EMA period used for every expected signal array.
const UL = 3;

interface Combo {
  name: string;
  q: number;
  r: number;
  s: number;
  u: number;
  smi: number[];
  signal: number[];
}

const combos: Combo[] = [
  { name: 'Q5_R20_S5_U3', q: 5, r: 20, s: 5, u: 3, smi: expectedQ5_R20_S5_U3, signal: expectedQ5_R20_S5_U3_SIG_UL3 },
  { name: 'Q13_R25_S2_U1', q: 13, r: 25, s: 2, u: 1, smi: expectedQ13_R25_S2_U1, signal: expectedQ13_R25_S2_U1_SIG_UL3 },
  { name: 'Q2_R20_S20_U1', q: 2, r: 20, s: 20, u: 1, smi: expectedQ2_R20_S20_U1, signal: expectedQ2_R20_S20_U1_SIG_UL3 },
  { name: 'Q13_R25_S2_U3', q: 13, r: 25, s: 2, u: 3, smi: expectedQ13_R25_S2_U3, signal: expectedQ13_R25_S2_U3_SIG_UL3 },
  { name: 'Q5_R20_S5_U1', q: 5, r: 20, s: 5, u: 1, smi: expectedQ5_R20_S5_U1, signal: expectedQ5_R20_S5_U1_SIG_UL3 },
  { name: 'Q8_R5_S3_U1', q: 8, r: 5, s: 3, u: 1, smi: expectedQ8_R5_S3_U1, signal: expectedQ8_R5_S3_U1_SIG_UL3 },
  { name: 'Q21_R13_S4_U1', q: 21, r: 13, s: 4, u: 1, smi: expectedQ21_R13_S4_U1, signal: expectedQ21_R13_S4_U1_SIG_UL3 },
  { name: 'Q1_R20_S5_U3', q: 1, r: 20, s: 5, u: 3, smi: expectedQ1_R20_S5_U3, signal: expectedQ1_R20_S5_U3_SIG_UL3 },
  { name: 'Q1_R40_S20_U1', q: 1, r: 40, s: 20, u: 1, smi: expectedQ1_R40_S20_U1, signal: expectedQ1_R40_S20_U1_SIG_UL3 },
  { name: 'Q1_R100_S20_U1', q: 1, r: 100, s: 20, u: 1, smi: expectedQ1_R100_S20_U1, signal: expectedQ1_R100_S20_U1_SIG_UL3 },
  { name: 'Q1_R1_S1_U1', q: 1, r: 1, s: 1, u: 1, smi: expectedQ1_R1_S1_U1, signal: expectedQ1_R1_S1_U1_SIG_UL3 },
  { name: 'Q5_R1_S1_U1', q: 5, r: 1, s: 1, u: 1, smi: expectedQ5_R1_S1_U1, signal: expectedQ5_R1_S1_U1_SIG_UL3 },
  { name: 'Q3_R10_S10_U1', q: 3, r: 10, s: 10, u: 1, smi: expectedQ3_R10_S10_U1, signal: expectedQ3_R10_S10_U1_SIG_UL3 },
  { name: 'Q34_R5_S5_U1', q: 34, r: 5, s: 5, u: 1, smi: expectedQ34_R5_S5_U1, signal: expectedQ34_R5_S5_U1_SIG_UL3 },
  { name: 'Q2_R2_S2_U2', q: 2, r: 2, s: 2, u: 2, smi: expectedQ2_R2_S2_U2, signal: expectedQ2_R2_S2_U2_SIG_UL3 },
  { name: 'Q50_R20_S5_U3', q: 50, r: 20, s: 5, u: 3, smi: expectedQ50_R20_S5_U3, signal: expectedQ50_R20_S5_U3_SIG_UL3 },
];

describe('StochasticMomentumIndex', () => {
  describe('reference data', () => {
    combos.forEach((combo) => {
      it(`matches the reference for ${combo.name}`, () => {
        const ind = new StochasticMomentumIndex({
          q: combo.q, r: combo.r, s: combo.s, u: combo.u, ul: UL,
        });

        for (let i = 0; i < testInput.length; i++) {
          const [smi, signal] = ind.update(testHigh[i], testLow[i], testInput[i]);

          if (isNaN(combo.smi[i])) {
            expect(smi).toBeNaN();
          } else {
            expect(smi).toBeCloseTo(combo.smi[i], 9);
          }

          if (isNaN(combo.signal[i])) {
            expect(signal).toBeNaN();
          } else {
            expect(signal).toBeCloseTo(combo.signal[i], 9);
          }
        }
      });
    });
  });

  describe('passthrough', () => {
    it('reduces to 100*(close-mid)/half-range for the raw one-day stochastic', () => {
      const ind = new StochasticMomentumIndex({ q: 1, r: 1, s: 1, u: 1, ul: 1 });

      expect(ind.update(12, 10, 12)).toEqual([100, 100]);   // close at high
      expect(ind.update(12, 10, 10)).toEqual([-100, -100]); // close at low
      expect(ind.update(12, 10, 11)).toEqual([0, 0]);       // exact midpoint
      expect(ind.update(11, 11, 11)).toEqual([0, 0]);       // flat window -> division guard
    });
  });

  describe('isPrimed', () => {
    it('is NaN and not primed for bars 0..q-2, primed from bar q-1', () => {
      const q = 5;
      const ind = new StochasticMomentumIndex({ q });
      expect(ind.isPrimed()).toBe(false);

      for (let i = 0; i < q - 1; i++) {
        const [smi, signal] = ind.update(testHigh[i], testLow[i], testInput[i]);
        expect(ind.isPrimed()).toBe(false);
        expect(smi).toBeNaN();
        expect(signal).toBeNaN();
      }

      for (let i = q - 1; i < testInput.length; i++) {
        ind.update(testHigh[i], testLow[i], testInput[i]);
        expect(ind.isPrimed()).toBe(true);
      }
    });

    it('is primed after the first bar when q = 1', () => {
      const ind = new StochasticMomentumIndex({ q: 1 });
      expect(ind.isPrimed()).toBe(false);
      ind.update(testHigh[0], testLow[0], testInput[0]);
      expect(ind.isPrimed()).toBe(true);
    });
  });

  describe('mnemonic', () => {
    it('formats the default mnemonic', () => {
      const ind = new StochasticMomentumIndex();
      expect(ind.metadata().mnemonic).toBe('smi(5,20,5,3,3)');
      expect(ind.metadata().description).toBe('Stochastic Momentum Index smi(5,20,5,3,3)');
    });

    it('formats a custom mnemonic', () => {
      const ind = new StochasticMomentumIndex({ q: 13, r: 25, s: 2, u: 1, ul: 7 });
      expect(ind.metadata().mnemonic).toBe('smi(13,25,2,1,7)');
    });
  });

  describe('metadata', () => {
    it('reports the identifier and two outputs', () => {
      const ind = new StochasticMomentumIndex();
      const meta = ind.metadata();
      expect(meta.identifier).toBe(IndicatorIdentifier.StochasticMomentumIndex);
      expect(meta.outputs.length).toBe(2);
      expect(meta.outputs[0].mnemonic).toBe('smi(5,20,5,3,3) smi');
      expect(meta.outputs[0].description).toBe('Stochastic Momentum Index smi(5,20,5,3,3) SMI');
      expect(meta.outputs[1].mnemonic).toBe('smi(5,20,5,3,3) signal');
      expect(meta.outputs[1].description).toBe('Stochastic Momentum Index smi(5,20,5,3,3) signal');
    });
  });

  describe('updateBar', () => {
    it('returns [smi, signal] in order', () => {
      const ind = new StochasticMomentumIndex({ q: 5, r: 20, s: 5, u: 3, ul: UL });
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
      expect((out[0] as Scalar).value).toBeCloseTo(expectedQ5_R20_S5_U3[last], 9);
      expect((out[1] as Scalar).value).toBeCloseTo(expectedQ5_R20_S5_U3_SIG_UL3[last], 9);
    });
  });

  describe('updateScalar', () => {
    it('uses the value as the high, the low and the close', () => {
      const ind = new StochasticMomentumIndex({ q: 2, r: 1, s: 1, u: 1, ul: 1 });
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
      const ind = new StochasticMomentumIndex({ q: 2, r: 1, s: 1, u: 1, ul: 1 });
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
      expect((out[0] as Scalar).value).toBe(-100);
      expect((out[1] as Scalar).value).toBe(-100);
    });
  });

  describe('updateTrade', () => {
    it('uses the price as the high, the low and the close', () => {
      const ind = new StochasticMomentumIndex({ q: 2, r: 1, s: 1, u: 1, ul: 1 });
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
      expect(() => new StochasticMomentumIndex({ q: 0 })).toThrowError();
    });
    it('throws when r < 1', () => {
      expect(() => new StochasticMomentumIndex({ r: 0 })).toThrowError();
    });
    it('throws when s < 1', () => {
      expect(() => new StochasticMomentumIndex({ s: 0 })).toThrowError();
    });
    it('throws when u < 1', () => {
      expect(() => new StochasticMomentumIndex({ u: 0 })).toThrowError();
    });
    it('throws when ul < 1', () => {
      expect(() => new StochasticMomentumIndex({ ul: 0 })).toThrowError();
    });
  });
});
