import { SchaffTrendCycle } from './schaff-trend-cycle';
import { IndicatorIdentifier } from '../../core/indicator-identifier';
import { Scalar } from '../../../entities/scalar';
import {
  testInput,
  expectedStcF23_S50_T10_C50, expectedMacdF23_S50_T10_C50, expectedPfF23_S50_T10_C50,
  expectedStcF12_S26_T10_C50, expectedMacdF12_S26_T10_C50, expectedPfF12_S26_T10_C50,
  expectedStcF5_S10_T5_C50, expectedMacdF5_S10_T5_C50, expectedPfF5_S10_T5_C50,
  expectedStcF3_S7_T3_C50,
  expectedStcF8_S21_T10_C50,
  expectedStcF10_S30_T10_C50,
  expectedStcF15_S40_T14_C50,
  expectedStcF6_S13_T8_C60,
  expectedStcF23_S50_T23_C50,
  expectedStcF23_S50_T5_C50,
  expectedStcF12_S26_T10_C25,
  expectedStcF12_S26_T10_C80,
  expectedStcF12_S26_T10_C100,
  expectedStcF20_S40_T10_C50,
} from './testdata';

interface Combo {
  name: string;
  fast: number;
  slow: number;
  tclen: number;
  factor: number;
  stc: number[];
  macd?: number[];
  pf?: number[];
}

const combos: Combo[] = [
  { name: 'F23_S50_T10_C50', fast: 23, slow: 50, tclen: 10, factor: 0.5, stc: expectedStcF23_S50_T10_C50, macd: expectedMacdF23_S50_T10_C50, pf: expectedPfF23_S50_T10_C50 },
  { name: 'F12_S26_T10_C50', fast: 12, slow: 26, tclen: 10, factor: 0.5, stc: expectedStcF12_S26_T10_C50, macd: expectedMacdF12_S26_T10_C50, pf: expectedPfF12_S26_T10_C50 },
  { name: 'F5_S10_T5_C50', fast: 5, slow: 10, tclen: 5, factor: 0.5, stc: expectedStcF5_S10_T5_C50, macd: expectedMacdF5_S10_T5_C50, pf: expectedPfF5_S10_T5_C50 },
  { name: 'F3_S7_T3_C50', fast: 3, slow: 7, tclen: 3, factor: 0.5, stc: expectedStcF3_S7_T3_C50 },
  { name: 'F8_S21_T10_C50', fast: 8, slow: 21, tclen: 10, factor: 0.5, stc: expectedStcF8_S21_T10_C50 },
  { name: 'F10_S30_T10_C50', fast: 10, slow: 30, tclen: 10, factor: 0.5, stc: expectedStcF10_S30_T10_C50 },
  { name: 'F15_S40_T14_C50', fast: 15, slow: 40, tclen: 14, factor: 0.5, stc: expectedStcF15_S40_T14_C50 },
  { name: 'F6_S13_T8_C60', fast: 6, slow: 13, tclen: 8, factor: 0.6, stc: expectedStcF6_S13_T8_C60 },
  { name: 'F23_S50_T23_C50', fast: 23, slow: 50, tclen: 23, factor: 0.5, stc: expectedStcF23_S50_T23_C50 },
  { name: 'F23_S50_T5_C50', fast: 23, slow: 50, tclen: 5, factor: 0.5, stc: expectedStcF23_S50_T5_C50 },
  { name: 'F12_S26_T10_C25', fast: 12, slow: 26, tclen: 10, factor: 0.25, stc: expectedStcF12_S26_T10_C25 },
  { name: 'F12_S26_T10_C80', fast: 12, slow: 26, tclen: 10, factor: 0.8, stc: expectedStcF12_S26_T10_C80 },
  { name: 'F12_S26_T10_C100', fast: 12, slow: 26, tclen: 10, factor: 1.0, stc: expectedStcF12_S26_T10_C100 },
  { name: 'F20_S40_T10_C50', fast: 20, slow: 40, tclen: 10, factor: 0.5, stc: expectedStcF20_S40_T10_C50 },
];

describe('SchaffTrendCycle', () => {
  describe('reference data', () => {
    combos.forEach((combo) => {
      it(`matches the reference for ${combo.name}`, () => {
        const ind = new SchaffTrendCycle({
          fast: combo.fast, slow: combo.slow, tclen: combo.tclen, factor: combo.factor,
        });

        for (let i = 0; i < testInput.length; i++) {
          const [stc, macd, pf] = ind.update(testInput[i]);

          if (isNaN(combo.stc[i])) {
            expect(stc).toBeNaN();
          } else {
            expect(stc).toBeCloseTo(combo.stc[i], 9);
          }

          if (combo.macd !== undefined) {
            expect(macd).toBeCloseTo(combo.macd[i], 9);
          }
          if (combo.pf !== undefined) {
            expect(pf).toBeCloseTo(combo.pf[i], 9);
          }
        }
      });
    });
  });

  describe('mnemonic', () => {
    it('formats the default mnemonic', () => {
      const ind = new SchaffTrendCycle();
      expect(ind.metadata().mnemonic).toBe('stc(23,50,10,0.50)');
    });

    it('formats a custom mnemonic', () => {
      const ind = new SchaffTrendCycle({ fast: 12, slow: 26, tclen: 10, factor: 0.25 });
      expect(ind.metadata().mnemonic).toBe('stc(12,26,10,0.25)');
    });
  });

  describe('metadata', () => {
    it('reports the identifier and three outputs', () => {
      const ind = new SchaffTrendCycle();
      const meta = ind.metadata();
      expect(meta.identifier).toBe(IndicatorIdentifier.SchaffTrendCycle);
      expect(meta.outputs.length).toBe(3);
    });
  });

  describe('updateScalar', () => {
    it('returns [stc, macd, pf] in order', () => {
      const ind = new SchaffTrendCycle();
      let out: ReturnType<typeof ind.updateScalar> = [];
      for (let i = 0; i < testInput.length; i++) {
        const s = new Scalar();
        s.time = new Date(0);
        s.value = testInput[i];
        out = ind.updateScalar(s);
      }
      const last = testInput.length - 1;
      expect((out[0] as Scalar).value).toBeCloseTo(expectedStcF23_S50_T10_C50[last], 9);
      expect((out[1] as Scalar).value).toBeCloseTo(expectedMacdF23_S50_T10_C50[last], 9);
      expect((out[2] as Scalar).value).toBeCloseTo(expectedPfF23_S50_T10_C50[last], 9);
    });
  });

  describe('invalid parameters', () => {
    it('throws when fast < 1', () => {
      expect(() => new SchaffTrendCycle({ fast: 0 })).toThrowError();
    });
    it('throws when slow < 1', () => {
      expect(() => new SchaffTrendCycle({ slow: 0 })).toThrowError();
    });
    it('throws when tclen < 1', () => {
      expect(() => new SchaffTrendCycle({ tclen: 0 })).toThrowError();
    });
    it('throws when factor <= 0', () => {
      expect(() => new SchaffTrendCycle({ factor: 0 })).toThrowError();
    });
    it('throws when factor > 1', () => {
      expect(() => new SchaffTrendCycle({ factor: 1.5 })).toThrowError();
    });
  });
});
