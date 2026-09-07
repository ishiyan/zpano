import { ErgodicOscillator } from './ergodic-oscillator';
import { IndicatorIdentifier } from '../../core/indicator-identifier';
import { Scalar } from '../../../entities/scalar';
import {
  testInput,
  expectedErgQ2_R20_S5_U3_L3, expectedSigQ2_R20_S5_U3_L3,
  expectedErgQ2_R32_S5_U1_L5, expectedSigQ2_R32_S5_U1_L5,
  expectedErgQ2_R20_S5_U1_L5, expectedSigQ2_R20_S5_U1_L5,
  expectedErgQ2_R32_S5_U1_L7, expectedSigQ2_R32_S5_U1_L7,
  expectedErgQ2_R25_S13_U1_L5, expectedSigQ2_R25_S13_U1_L5,
  expectedErgQ2_R20_S5_U3_L1, expectedSigQ2_R20_S5_U3_L1,
  expectedErgQ2_R1_S1_U1_L1, expectedSigQ2_R1_S1_U1_L1,
  expectedErgQ2_R20_S5_U3_L9, expectedSigQ2_R20_S5_U3_L9,
  expectedErgQ2_R64_S64_U1_L5, expectedSigQ2_R64_S64_U1_L5,
  expectedErgQ2_R9_S3_U1_L3, expectedSigQ2_R9_S3_U1_L3,
  expectedErgQ3_R20_S5_U3_L3, expectedSigQ3_R20_S5_U3_L3,
  expectedErgQ5_R20_S5_U3_L5, expectedSigQ5_R20_S5_U3_L5,
  expectedErgQ2_R13_S7_U1_L7, expectedSigQ2_R13_S7_U1_L7,
  expectedErgQ2_R20_S5_U1_L3, expectedSigQ2_R20_S5_U1_L3,
} from './testdata';

interface Combo {
  name: string;
  q: number;
  r: number;
  s: number;
  u: number;
  ul: number;
  ergodic: number[];
  signal: number[];
}

const combos: Combo[] = [
  { name: 'Q2_R20_S5_U3_L3', q: 2, r: 20, s: 5, u: 3, ul: 3, ergodic: expectedErgQ2_R20_S5_U3_L3, signal: expectedSigQ2_R20_S5_U3_L3 },
  { name: 'Q2_R32_S5_U1_L5', q: 2, r: 32, s: 5, u: 1, ul: 5, ergodic: expectedErgQ2_R32_S5_U1_L5, signal: expectedSigQ2_R32_S5_U1_L5 },
  { name: 'Q2_R20_S5_U1_L5', q: 2, r: 20, s: 5, u: 1, ul: 5, ergodic: expectedErgQ2_R20_S5_U1_L5, signal: expectedSigQ2_R20_S5_U1_L5 },
  { name: 'Q2_R32_S5_U1_L7', q: 2, r: 32, s: 5, u: 1, ul: 7, ergodic: expectedErgQ2_R32_S5_U1_L7, signal: expectedSigQ2_R32_S5_U1_L7 },
  { name: 'Q2_R25_S13_U1_L5', q: 2, r: 25, s: 13, u: 1, ul: 5, ergodic: expectedErgQ2_R25_S13_U1_L5, signal: expectedSigQ2_R25_S13_U1_L5 },
  { name: 'Q2_R20_S5_U3_L1', q: 2, r: 20, s: 5, u: 3, ul: 1, ergodic: expectedErgQ2_R20_S5_U3_L1, signal: expectedSigQ2_R20_S5_U3_L1 },
  { name: 'Q2_R1_S1_U1_L1', q: 2, r: 1, s: 1, u: 1, ul: 1, ergodic: expectedErgQ2_R1_S1_U1_L1, signal: expectedSigQ2_R1_S1_U1_L1 },
  { name: 'Q2_R20_S5_U3_L9', q: 2, r: 20, s: 5, u: 3, ul: 9, ergodic: expectedErgQ2_R20_S5_U3_L9, signal: expectedSigQ2_R20_S5_U3_L9 },
  { name: 'Q2_R64_S64_U1_L5', q: 2, r: 64, s: 64, u: 1, ul: 5, ergodic: expectedErgQ2_R64_S64_U1_L5, signal: expectedSigQ2_R64_S64_U1_L5 },
  { name: 'Q2_R9_S3_U1_L3', q: 2, r: 9, s: 3, u: 1, ul: 3, ergodic: expectedErgQ2_R9_S3_U1_L3, signal: expectedSigQ2_R9_S3_U1_L3 },
  { name: 'Q3_R20_S5_U3_L3', q: 3, r: 20, s: 5, u: 3, ul: 3, ergodic: expectedErgQ3_R20_S5_U3_L3, signal: expectedSigQ3_R20_S5_U3_L3 },
  { name: 'Q5_R20_S5_U3_L5', q: 5, r: 20, s: 5, u: 3, ul: 5, ergodic: expectedErgQ5_R20_S5_U3_L5, signal: expectedSigQ5_R20_S5_U3_L5 },
  { name: 'Q2_R13_S7_U1_L7', q: 2, r: 13, s: 7, u: 1, ul: 7, ergodic: expectedErgQ2_R13_S7_U1_L7, signal: expectedSigQ2_R13_S7_U1_L7 },
  { name: 'Q2_R20_S5_U1_L3', q: 2, r: 20, s: 5, u: 1, ul: 3, ergodic: expectedErgQ2_R20_S5_U1_L3, signal: expectedSigQ2_R20_S5_U1_L3 },
];

describe('ErgodicOscillator', () => {
  describe('reference data', () => {
    combos.forEach((combo) => {
      it(`matches the reference for ${combo.name}`, () => {
        const ind = new ErgodicOscillator({
          q: combo.q, r: combo.r, s: combo.s, u: combo.u, ul: combo.ul,
        });

        for (let i = 0; i < testInput.length; i++) {
          const [ergodic, signal] = ind.update(testInput[i]);

          if (isNaN(combo.ergodic[i])) {
            expect(ergodic).toBeNaN();
          } else {
            expect(ergodic).toBeCloseTo(combo.ergodic[i], 9);
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

  describe('signal passthrough', () => {
    it('returns signal === ergodic when ul = 1', () => {
      const ind = new ErgodicOscillator({ q: 2, r: 20, s: 5, u: 3, ul: 1 });

      for (let i = 0; i < testInput.length; i++) {
        const [ergodic, signal] = ind.update(testInput[i]);

        if (isNaN(ergodic)) {
          expect(signal).toBeNaN();
        } else {
          expect(signal).toBe(ergodic);
        }
      }
    });
  });

  describe('isPrimed', () => {
    it('primes on the first finite oscillator value (bar q-1)', () => {
      const ind = new ErgodicOscillator({ q: 3, r: 20, s: 5, u: 3, ul: 3 });
      expect(ind.isPrimed()).toBe(false);
      ind.update(testInput[0]);
      expect(ind.isPrimed()).toBe(false);
      ind.update(testInput[1]);
      expect(ind.isPrimed()).toBe(false);
      ind.update(testInput[2]);
      expect(ind.isPrimed()).toBe(true);
    });
  });

  describe('mnemonic', () => {
    it('formats the default mnemonic', () => {
      const ind = new ErgodicOscillator();
      expect(ind.metadata().mnemonic).toBe('ergodic(2,20,5,3,3)');
    });

    it('formats a custom mnemonic', () => {
      const ind = new ErgodicOscillator({ q: 2, r: 25, s: 13, u: 1, ul: 7 });
      expect(ind.metadata().mnemonic).toBe('ergodic(2,25,13,1,7)');
    });
  });

  describe('metadata', () => {
    it('reports the identifier and two outputs', () => {
      const ind = new ErgodicOscillator();
      const meta = ind.metadata();
      expect(meta.identifier).toBe(IndicatorIdentifier.ErgodicOscillator);
      expect(meta.outputs.length).toBe(2);
    });
  });

  describe('updateScalar', () => {
    it('returns [ergodic, signal] in order', () => {
      const ind = new ErgodicOscillator({ q: 2, r: 20, s: 5, u: 3, ul: 3 });
      let out: ReturnType<typeof ind.updateScalar> = [];
      for (let i = 0; i < testInput.length; i++) {
        const s = new Scalar();
        s.time = new Date(0);
        s.value = testInput[i];
        out = ind.updateScalar(s);
      }
      const last = testInput.length - 1;
      expect((out[0] as Scalar).value).toBeCloseTo(expectedErgQ2_R20_S5_U3_L3[last], 9);
      expect((out[1] as Scalar).value).toBeCloseTo(expectedSigQ2_R20_S5_U3_L3[last], 9);
    });
  });

  describe('invalid parameters', () => {
    it('throws when q < 1', () => {
      expect(() => new ErgodicOscillator({ q: 0 })).toThrowError();
    });
    it('throws when r < 1', () => {
      expect(() => new ErgodicOscillator({ r: 0 })).toThrowError();
    });
    it('throws when s < 1', () => {
      expect(() => new ErgodicOscillator({ s: 0 })).toThrowError();
    });
    it('throws when u < 1', () => {
      expect(() => new ErgodicOscillator({ u: 0 })).toThrowError();
    });
    it('throws when ul < 1', () => {
      expect(() => new ErgodicOscillator({ ul: 0 })).toThrowError();
    });
  });
});
