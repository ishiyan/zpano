import { TrueStrengthIndex } from './true-strength-index';
import { IndicatorIdentifier } from '../../core/indicator-identifier';
import { Scalar } from '../../../entities/scalar';
import {
  testInput,
  expectedQ2_R20_S5_U3, expectedQ2_R20_S5_U3_SIG_UL3,
  expectedQ2_R25_S13_U1, expectedQ2_R25_S13_U1_SIG_UL3,
  expectedQ2_R20_S5_U1, expectedQ2_R20_S5_U1_SIG_UL3,
  expectedQ2_R32_S5_U1, expectedQ2_R32_S5_U1_SIG_UL3,
  expectedQ2_R13_S13_U1, expectedQ2_R13_S13_U1_SIG_UL3,
  expectedQ2_R20_S40_U1, expectedQ2_R20_S40_U1_SIG_UL3,
  expectedQ2_R40_S20_U1, expectedQ2_R40_S20_U1_SIG_UL3,
  expectedQ2_R64_S64_U1, expectedQ2_R64_S64_U1_SIG_UL3,
  expectedQ2_R100_S5_U1, expectedQ2_R100_S5_U1_SIG_UL3,
  expectedQ2_R1_S1_U1, expectedQ2_R1_S1_U1_SIG_UL3,
  expectedQ2_R1_S5_U3, expectedQ2_R1_S5_U3_SIG_UL3,
  expectedQ2_R20_S1_U1, expectedQ2_R20_S1_U1_SIG_UL3,
  expectedQ2_R5_S5_U5, expectedQ2_R5_S5_U5_SIG_UL3,
  expectedQ3_R20_S5_U3, expectedQ3_R20_S5_U3_SIG_UL3,
  expectedQ5_R20_S5_U3, expectedQ5_R20_S5_U3_SIG_UL3,
  expectedQ10_R20_S5_U1, expectedQ10_R20_S5_U1_SIG_UL3,
  expectedQ2_R9_S3_U1, expectedQ2_R9_S3_U1_SIG_UL3,
  expectedQ2_R7_S4_U2, expectedQ2_R7_S4_U2_SIG_UL3,
} from './testdata';

// Signal-line EMA period used for every expected signal array.
const UL = 3;

interface Combo {
  name: string;
  q: number;
  r: number;
  s: number;
  u: number;
  tsi: number[];
  signal: number[];
}

const combos: Combo[] = [
  { name: 'Q2_R20_S5_U3', q: 2, r: 20, s: 5, u: 3, tsi: expectedQ2_R20_S5_U3, signal: expectedQ2_R20_S5_U3_SIG_UL3 },
  { name: 'Q2_R25_S13_U1', q: 2, r: 25, s: 13, u: 1, tsi: expectedQ2_R25_S13_U1, signal: expectedQ2_R25_S13_U1_SIG_UL3 },
  { name: 'Q2_R20_S5_U1', q: 2, r: 20, s: 5, u: 1, tsi: expectedQ2_R20_S5_U1, signal: expectedQ2_R20_S5_U1_SIG_UL3 },
  { name: 'Q2_R32_S5_U1', q: 2, r: 32, s: 5, u: 1, tsi: expectedQ2_R32_S5_U1, signal: expectedQ2_R32_S5_U1_SIG_UL3 },
  { name: 'Q2_R13_S13_U1', q: 2, r: 13, s: 13, u: 1, tsi: expectedQ2_R13_S13_U1, signal: expectedQ2_R13_S13_U1_SIG_UL3 },
  { name: 'Q2_R20_S40_U1', q: 2, r: 20, s: 40, u: 1, tsi: expectedQ2_R20_S40_U1, signal: expectedQ2_R20_S40_U1_SIG_UL3 },
  { name: 'Q2_R40_S20_U1', q: 2, r: 40, s: 20, u: 1, tsi: expectedQ2_R40_S20_U1, signal: expectedQ2_R40_S20_U1_SIG_UL3 },
  { name: 'Q2_R64_S64_U1', q: 2, r: 64, s: 64, u: 1, tsi: expectedQ2_R64_S64_U1, signal: expectedQ2_R64_S64_U1_SIG_UL3 },
  { name: 'Q2_R100_S5_U1', q: 2, r: 100, s: 5, u: 1, tsi: expectedQ2_R100_S5_U1, signal: expectedQ2_R100_S5_U1_SIG_UL3 },
  { name: 'Q2_R1_S1_U1', q: 2, r: 1, s: 1, u: 1, tsi: expectedQ2_R1_S1_U1, signal: expectedQ2_R1_S1_U1_SIG_UL3 },
  { name: 'Q2_R1_S5_U3', q: 2, r: 1, s: 5, u: 3, tsi: expectedQ2_R1_S5_U3, signal: expectedQ2_R1_S5_U3_SIG_UL3 },
  { name: 'Q2_R20_S1_U1', q: 2, r: 20, s: 1, u: 1, tsi: expectedQ2_R20_S1_U1, signal: expectedQ2_R20_S1_U1_SIG_UL3 },
  { name: 'Q2_R5_S5_U5', q: 2, r: 5, s: 5, u: 5, tsi: expectedQ2_R5_S5_U5, signal: expectedQ2_R5_S5_U5_SIG_UL3 },
  { name: 'Q3_R20_S5_U3', q: 3, r: 20, s: 5, u: 3, tsi: expectedQ3_R20_S5_U3, signal: expectedQ3_R20_S5_U3_SIG_UL3 },
  { name: 'Q5_R20_S5_U3', q: 5, r: 20, s: 5, u: 3, tsi: expectedQ5_R20_S5_U3, signal: expectedQ5_R20_S5_U3_SIG_UL3 },
  { name: 'Q10_R20_S5_U1', q: 10, r: 20, s: 5, u: 1, tsi: expectedQ10_R20_S5_U1, signal: expectedQ10_R20_S5_U1_SIG_UL3 },
  { name: 'Q2_R9_S3_U1', q: 2, r: 9, s: 3, u: 1, tsi: expectedQ2_R9_S3_U1, signal: expectedQ2_R9_S3_U1_SIG_UL3 },
  { name: 'Q2_R7_S4_U2', q: 2, r: 7, s: 4, u: 2, tsi: expectedQ2_R7_S4_U2, signal: expectedQ2_R7_S4_U2_SIG_UL3 },
];

describe('TrueStrengthIndex', () => {
  describe('reference data', () => {
    combos.forEach((combo) => {
      it(`matches the reference for ${combo.name}`, () => {
        const ind = new TrueStrengthIndex({
          q: combo.q, r: combo.r, s: combo.s, u: combo.u, ul: UL,
        });

        for (let i = 0; i < testInput.length; i++) {
          const [tsi, signal] = ind.update(testInput[i]);

          if (isNaN(combo.tsi[i])) {
            expect(tsi).toBeNaN();
          } else {
            expect(tsi).toBeCloseTo(combo.tsi[i], 9);
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

  describe('mnemonic', () => {
    it('formats the default mnemonic (ul excluded)', () => {
      const ind = new TrueStrengthIndex();
      expect(ind.metadata().mnemonic).toBe('tsi(2,20,5,3)');
    });

    it('formats a custom mnemonic', () => {
      const ind = new TrueStrengthIndex({ q: 2, r: 25, s: 13, u: 1, ul: 7 });
      expect(ind.metadata().mnemonic).toBe('tsi(2,25,13,1)');
    });
  });

  describe('metadata', () => {
    it('reports the identifier and two outputs', () => {
      const ind = new TrueStrengthIndex();
      const meta = ind.metadata();
      expect(meta.identifier).toBe(IndicatorIdentifier.TrueStrengthIndex);
      expect(meta.outputs.length).toBe(2);
    });
  });

  describe('updateScalar', () => {
    it('returns [tsi, signal] in order', () => {
      const ind = new TrueStrengthIndex({ q: 2, r: 20, s: 5, u: 3, ul: UL });
      let out: ReturnType<typeof ind.updateScalar> = [];
      for (let i = 0; i < testInput.length; i++) {
        const s = new Scalar();
        s.time = new Date(0);
        s.value = testInput[i];
        out = ind.updateScalar(s);
      }
      const last = testInput.length - 1;
      expect((out[0] as Scalar).value).toBeCloseTo(expectedQ2_R20_S5_U3[last], 9);
      expect((out[1] as Scalar).value).toBeCloseTo(expectedQ2_R20_S5_U3_SIG_UL3[last], 9);
    });
  });

  describe('invalid parameters', () => {
    it('throws when q < 1', () => {
      expect(() => new TrueStrengthIndex({ q: 0 })).toThrowError();
    });
    it('throws when r < 1', () => {
      expect(() => new TrueStrengthIndex({ r: 0 })).toThrowError();
    });
    it('throws when s < 1', () => {
      expect(() => new TrueStrengthIndex({ s: 0 })).toThrowError();
    });
    it('throws when u < 1', () => {
      expect(() => new TrueStrengthIndex({ u: 0 })).toThrowError();
    });
    it('throws when ul < 1', () => {
      expect(() => new TrueStrengthIndex({ ul: 0 })).toThrowError();
    });
  });
});
