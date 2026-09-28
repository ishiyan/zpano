import { DirectionalTrendIndex } from './directional-trend-index';
import { IndicatorIdentifier } from '../../core/indicator-identifier';
import { Bar } from '../../../entities/bar';
import { Quote } from '../../../entities/quote';
import { Scalar } from '../../../entities/scalar';
import { Trade } from '../../../entities/trade';
import {
  testHigh, testLow,
  expectedQ2_R20_S5_U3, expectedQ2_R20_S5_U3_SIG_UL3,
  expectedQ2_R25_S13_U1, expectedQ2_R25_S13_U1_SIG_UL3,
  expectedQ2_R20_S5_U1, expectedQ2_R20_S5_U1_SIG_UL3,
  expectedQ2_R28_S28_U5, expectedQ2_R28_S28_U5_SIG_UL3,
  expectedQ2_R1_S1_U1, expectedQ2_R1_S1_U1_SIG_UL3,
  expectedQ3_R20_S5_U3, expectedQ3_R20_S5_U3_SIG_UL3,
  expectedQ5_R20_S5_U3, expectedQ5_R20_S5_U3_SIG_UL3,
  expectedQ2_R13_S13_U1, expectedQ2_R13_S13_U1_SIG_UL3,
  expectedQ2_R40_S20_U1, expectedQ2_R40_S20_U1_SIG_UL3,
  expectedQ2_R5_S5_U5, expectedQ2_R5_S5_U5_SIG_UL3,
  expectedQ1_R20_S5_U3, expectedQ1_R20_S5_U3_SIG_UL3,
  expectedQ10_R20_S5_U1, expectedQ10_R20_S5_U1_SIG_UL3,
  expectedQ2_R9_S3_U1, expectedQ2_R9_S3_U1_SIG_UL3,
  expectedQ2_R64_S64_U1, expectedQ2_R64_S64_U1_SIG_UL3,
  expectedQ4_R28_S28_U5, expectedQ4_R28_S28_U5_SIG_UL3,
  expectedQ2_R7_S4_U2, expectedQ2_R7_S4_U2_SIG_UL3,
} from './testdata';

interface Combo {
  name: string;
  q: number;
  r: number;
  s: number;
  u: number;
  ul: number;
  dti: number[];
  signal: number[];
}

const combos: Combo[] = [
  { name: 'Q2_R20_S5_U3', q: 2, r: 20, s: 5, u: 3, ul: 3, dti: expectedQ2_R20_S5_U3, signal: expectedQ2_R20_S5_U3_SIG_UL3 },
  { name: 'Q2_R25_S13_U1', q: 2, r: 25, s: 13, u: 1, ul: 3, dti: expectedQ2_R25_S13_U1, signal: expectedQ2_R25_S13_U1_SIG_UL3 },
  { name: 'Q2_R20_S5_U1', q: 2, r: 20, s: 5, u: 1, ul: 3, dti: expectedQ2_R20_S5_U1, signal: expectedQ2_R20_S5_U1_SIG_UL3 },
  { name: 'Q2_R28_S28_U5', q: 2, r: 28, s: 28, u: 5, ul: 3, dti: expectedQ2_R28_S28_U5, signal: expectedQ2_R28_S28_U5_SIG_UL3 },
  { name: 'Q2_R1_S1_U1', q: 2, r: 1, s: 1, u: 1, ul: 3, dti: expectedQ2_R1_S1_U1, signal: expectedQ2_R1_S1_U1_SIG_UL3 },
  { name: 'Q3_R20_S5_U3', q: 3, r: 20, s: 5, u: 3, ul: 3, dti: expectedQ3_R20_S5_U3, signal: expectedQ3_R20_S5_U3_SIG_UL3 },
  { name: 'Q5_R20_S5_U3', q: 5, r: 20, s: 5, u: 3, ul: 3, dti: expectedQ5_R20_S5_U3, signal: expectedQ5_R20_S5_U3_SIG_UL3 },
  { name: 'Q2_R13_S13_U1', q: 2, r: 13, s: 13, u: 1, ul: 3, dti: expectedQ2_R13_S13_U1, signal: expectedQ2_R13_S13_U1_SIG_UL3 },
  { name: 'Q2_R40_S20_U1', q: 2, r: 40, s: 20, u: 1, ul: 3, dti: expectedQ2_R40_S20_U1, signal: expectedQ2_R40_S20_U1_SIG_UL3 },
  { name: 'Q2_R5_S5_U5', q: 2, r: 5, s: 5, u: 5, ul: 3, dti: expectedQ2_R5_S5_U5, signal: expectedQ2_R5_S5_U5_SIG_UL3 },
  { name: 'Q1_R20_S5_U3', q: 1, r: 20, s: 5, u: 3, ul: 3, dti: expectedQ1_R20_S5_U3, signal: expectedQ1_R20_S5_U3_SIG_UL3 },
  { name: 'Q10_R20_S5_U1', q: 10, r: 20, s: 5, u: 1, ul: 3, dti: expectedQ10_R20_S5_U1, signal: expectedQ10_R20_S5_U1_SIG_UL3 },
  { name: 'Q2_R9_S3_U1', q: 2, r: 9, s: 3, u: 1, ul: 3, dti: expectedQ2_R9_S3_U1, signal: expectedQ2_R9_S3_U1_SIG_UL3 },
  { name: 'Q2_R64_S64_U1', q: 2, r: 64, s: 64, u: 1, ul: 3, dti: expectedQ2_R64_S64_U1, signal: expectedQ2_R64_S64_U1_SIG_UL3 },
  { name: 'Q4_R28_S28_U5', q: 4, r: 28, s: 28, u: 5, ul: 3, dti: expectedQ4_R28_S28_U5, signal: expectedQ4_R28_S28_U5_SIG_UL3 },
  { name: 'Q2_R7_S4_U2', q: 2, r: 7, s: 4, u: 2, ul: 3, dti: expectedQ2_R7_S4_U2, signal: expectedQ2_R7_S4_U2_SIG_UL3 },
];

describe('DirectionalTrendIndex', () => {
  describe('reference data', () => {
    combos.forEach((combo) => {
      it(`matches the reference for ${combo.name}`, () => {
        const ind = new DirectionalTrendIndex({
          q: combo.q, r: combo.r, s: combo.s, u: combo.u, ul: combo.ul,
        });

        for (let i = 0; i < testHigh.length; i++) {
          const [dti, signal] = ind.update(testHigh[i], testLow[i]);

          if (isNaN(combo.dti[i])) {
            expect(dti).toBeNaN();
          } else {
            expect(dti).toBeCloseTo(combo.dti[i], 13);
          }

          if (isNaN(combo.signal[i])) {
            expect(signal).toBeNaN();
          } else {
            expect(signal).toBeCloseTo(combo.signal[i], 13);
          }
        }
      });
    });
  });

  describe('passthrough', () => {
    it('reduces to 100*sign(HLM) when all stages are passthroughs', () => {
      const ind = new DirectionalTrendIndex({ q: 2, r: 1, s: 1, u: 1, ul: 1 });

      const [dti0, signal0] = ind.update(10, 9);
      expect(dti0).toBeNaN();
      expect(signal0).toBeNaN();
      expect(ind.update(12, 11)).toEqual([100, 100]);   // HMU=+2, LMD=0
      expect(ind.update(11, 8)).toEqual([-100, -100]);  // HMU=0, LMD=3
      expect(ind.update(10, 9)).toEqual([0, 0]);        // inside bar -> division guard
    });

    it('yields zero on every bar when q = 1', () => {
      const ind = new DirectionalTrendIndex({ q: 1 });

      for (let i = 0; i < testHigh.length; i++) {
        expect(ind.update(testHigh[i], testLow[i])).toEqual([0, 0]);
      }
    });

    it('signal equals dti when ul = 1', () => {
      const ind = new DirectionalTrendIndex({ ul: 1 });

      for (let i = 0; i < testHigh.length; i++) {
        const [dti, signal] = ind.update(testHigh[i], testLow[i]);
        if (isNaN(dti)) {
          expect(signal).toBeNaN();
        } else {
          expect(signal).toBe(dti);
        }
      }
    });
  });

  describe('isPrimed', () => {
    [2, 3, 5, 10].forEach((q) => {
      it(`is NaN and not primed for bars 0..q-2, primed from bar q-1 (q=${q})`, () => {
        const ind = new DirectionalTrendIndex({ q });
        expect(ind.isPrimed()).toBe(false);

        for (let i = 0; i < q - 1; i++) {
          const [dti, signal] = ind.update(testHigh[i], testLow[i]);
          expect(ind.isPrimed()).toBe(false);
          expect(dti).toBeNaN();
          expect(signal).toBeNaN();
        }

        for (let i = q - 1; i < testHigh.length; i++) {
          ind.update(testHigh[i], testLow[i]);
          expect(ind.isPrimed()).toBe(true);
        }
      });
    });

    it('is primed after the first bar when q = 1', () => {
      const ind = new DirectionalTrendIndex({ q: 1 });
      expect(ind.isPrimed()).toBe(false);
      ind.update(testHigh[0], testLow[0]);
      expect(ind.isPrimed()).toBe(true);
    });
  });

  describe('mnemonic', () => {
    it('formats the default mnemonic', () => {
      const ind = new DirectionalTrendIndex();
      expect(ind.metadata().mnemonic).toBe('dti(2,20,5,3,3)');
      expect(ind.metadata().description).toBe('Directional Trend Index dti(2,20,5,3,3)');
    });

    it('formats a custom mnemonic', () => {
      const ind = new DirectionalTrendIndex({ q: 4, r: 28, s: 28, u: 5, ul: 1 });
      expect(ind.metadata().mnemonic).toBe('dti(4,28,28,5,1)');
    });
  });

  describe('metadata', () => {
    it('reports the identifier and two outputs', () => {
      const ind = new DirectionalTrendIndex();
      const meta = ind.metadata();
      expect(meta.identifier).toBe(IndicatorIdentifier.DirectionalTrendIndex);
      expect(meta.outputs.length).toBe(2);
      expect(meta.outputs[0].mnemonic).toBe('dti(2,20,5,3,3) dti');
      expect(meta.outputs[0].description).toBe('Directional Trend Index dti(2,20,5,3,3) DTI');
      expect(meta.outputs[1].mnemonic).toBe('dti(2,20,5,3,3) signal');
      expect(meta.outputs[1].description).toBe('Directional Trend Index dti(2,20,5,3,3) signal');
    });
  });

  describe('updateBar', () => {
    it('returns [dti, signal] in order', () => {
      const ind = new DirectionalTrendIndex();
      let out: ReturnType<typeof ind.updateBar> = [];
      for (let i = 0; i < testHigh.length; i++) {
        const bar = new Bar();
        bar.time = new Date(0);
        bar.open = 0;
        bar.high = testHigh[i];
        bar.low = testLow[i];
        bar.close = 0;
        bar.volume = 0;
        out = ind.updateBar(bar);
      }
      const last = testHigh.length - 1;
      expect(out.length).toBe(2);
      expect((out[0] as Scalar).value).toBeCloseTo(expectedQ2_R20_S5_U3[last], 13);
      expect((out[1] as Scalar).value).toBeCloseTo(expectedQ2_R20_S5_U3_SIG_UL3[last], 13);
    });
  });

  describe('updateScalar', () => {
    it('uses the value as both the high and the low', () => {
      const ind = new DirectionalTrendIndex({ q: 2, r: 1, s: 1, u: 1, ul: 1 });
      const s0 = new Scalar();
      s0.time = new Date(0);
      s0.value = 10;
      let out = ind.updateScalar(s0);
      expect((out[0] as Scalar).value).toBeNaN();
      expect((out[1] as Scalar).value).toBeNaN();

      const s1 = new Scalar();
      s1.time = new Date(0);
      s1.value = 12; // rising value: HLM is the plain one-bar momentum
      out = ind.updateScalar(s1);
      expect((out[0] as Scalar).value).toBe(100);
      expect((out[1] as Scalar).value).toBe(100);
    });
  });

  describe('updateQuote', () => {
    it('uses the mid price as both the high and the low', () => {
      const ind = new DirectionalTrendIndex({ q: 2, r: 1, s: 1, u: 1, ul: 1 });
      const q0 = new Quote();
      q0.time = new Date(0);
      q0.bidPrice = 12;
      q0.askPrice = 14;
      ind.updateQuote(q0);

      const q1 = new Quote();
      q1.time = new Date(0);
      q1.bidPrice = 10;
      q1.askPrice = 12; // mid falls from 13 to 11
      const out = ind.updateQuote(q1);
      expect((out[0] as Scalar).value).toBe(-100);
      expect((out[1] as Scalar).value).toBe(-100);
    });
  });

  describe('updateTrade', () => {
    it('uses the price as both the high and the low', () => {
      const ind = new DirectionalTrendIndex({ q: 2, r: 1, s: 1, u: 1, ul: 1 });
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
      expect(() => new DirectionalTrendIndex({ q: 0 })).toThrowError();
    });
    it('throws when r < 1', () => {
      expect(() => new DirectionalTrendIndex({ r: 0 })).toThrowError();
    });
    it('throws when s < 1', () => {
      expect(() => new DirectionalTrendIndex({ s: 0 })).toThrowError();
    });
    it('throws when u < 1', () => {
      expect(() => new DirectionalTrendIndex({ u: 0 })).toThrowError();
    });
    it('throws when ul < 1', () => {
      expect(() => new DirectionalTrendIndex({ ul: 0 })).toThrowError();
    });
  });
});
