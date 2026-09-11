import { MacdIndex } from './macd-index';
import { IndicatorIdentifier } from '../../core/indicator-identifier';
import { Scalar } from '../../../entities/scalar';
import {
  testInput,
  expectedR20_S5_U3, expectedR20_S5_U3_SIG_UL3,
  expectedR20_S5_U1, expectedR20_S5_U1_SIG_UL3,
  expectedR26_S12_U3, expectedR26_S12_U3_SIG_UL3,
  expectedR26_S12_U1, expectedR26_S12_U1_SIG_UL3,
  expectedR35_S5_U3, expectedR35_S5_U3_SIG_UL3,
  expectedR10_S3_U5, expectedR10_S3_U5_SIG_UL3,
  expectedR32_S12_U5, expectedR32_S12_U5_SIG_UL3,
  expectedR17_S8_U1, expectedR17_S8_U1_SIG_UL3,
  expectedR20_S10_U3, expectedR20_S10_U3_SIG_UL3,
  expectedR8_S4_U2, expectedR8_S4_U2_SIG_UL3,
  expectedR30_S15_U1, expectedR30_S15_U1_SIG_UL3,
  expectedR3_S2_U3, expectedR3_S2_U3_SIG_UL3,
  expectedR50_S12_U1, expectedR50_S12_U1_SIG_UL3,
  expectedR19_S6_U3, expectedR19_S6_U3_SIG_UL3,
  expectedR20_S5_U5, expectedR20_S5_U5_SIG_UL3,
  expectedR60_S30_U10, expectedR60_S30_U10_SIG_UL3,
} from './testdata';

// Signal-line EMA period used for every expected signal array.
const UL = 3;

interface Combo {
  name: string;
  r: number;
  s: number;
  u: number;
  macdi: number[];
  signal: number[];
}

const combos: Combo[] = [
  { name: 'R20_S5_U3', r: 20, s: 5, u: 3, macdi: expectedR20_S5_U3, signal: expectedR20_S5_U3_SIG_UL3 },
  { name: 'R20_S5_U1', r: 20, s: 5, u: 1, macdi: expectedR20_S5_U1, signal: expectedR20_S5_U1_SIG_UL3 },
  { name: 'R26_S12_U3', r: 26, s: 12, u: 3, macdi: expectedR26_S12_U3, signal: expectedR26_S12_U3_SIG_UL3 },
  { name: 'R26_S12_U1', r: 26, s: 12, u: 1, macdi: expectedR26_S12_U1, signal: expectedR26_S12_U1_SIG_UL3 },
  { name: 'R35_S5_U3', r: 35, s: 5, u: 3, macdi: expectedR35_S5_U3, signal: expectedR35_S5_U3_SIG_UL3 },
  { name: 'R10_S3_U5', r: 10, s: 3, u: 5, macdi: expectedR10_S3_U5, signal: expectedR10_S3_U5_SIG_UL3 },
  { name: 'R32_S12_U5', r: 32, s: 12, u: 5, macdi: expectedR32_S12_U5, signal: expectedR32_S12_U5_SIG_UL3 },
  { name: 'R17_S8_U1', r: 17, s: 8, u: 1, macdi: expectedR17_S8_U1, signal: expectedR17_S8_U1_SIG_UL3 },
  { name: 'R20_S10_U3', r: 20, s: 10, u: 3, macdi: expectedR20_S10_U3, signal: expectedR20_S10_U3_SIG_UL3 },
  { name: 'R8_S4_U2', r: 8, s: 4, u: 2, macdi: expectedR8_S4_U2, signal: expectedR8_S4_U2_SIG_UL3 },
  { name: 'R30_S15_U1', r: 30, s: 15, u: 1, macdi: expectedR30_S15_U1, signal: expectedR30_S15_U1_SIG_UL3 },
  { name: 'R3_S2_U3', r: 3, s: 2, u: 3, macdi: expectedR3_S2_U3, signal: expectedR3_S2_U3_SIG_UL3 },
  { name: 'R50_S12_U1', r: 50, s: 12, u: 1, macdi: expectedR50_S12_U1, signal: expectedR50_S12_U1_SIG_UL3 },
  { name: 'R19_S6_U3', r: 19, s: 6, u: 3, macdi: expectedR19_S6_U3, signal: expectedR19_S6_U3_SIG_UL3 },
  { name: 'R20_S5_U5', r: 20, s: 5, u: 5, macdi: expectedR20_S5_U5, signal: expectedR20_S5_U5_SIG_UL3 },
  { name: 'R60_S30_U10', r: 60, s: 30, u: 10, macdi: expectedR60_S30_U10, signal: expectedR60_S30_U10_SIG_UL3 },
];

describe('MacdIndex', () => {
  describe('reference data', () => {
    combos.forEach((combo) => {
      it(`matches the reference for ${combo.name}`, () => {
        const ind = new MacdIndex({
          r: combo.r, s: combo.s, u: combo.u, ul: UL,
        });

        for (let i = 0; i < testInput.length; i++) {
          const [macdi, signal] = ind.update(testInput[i]);

          expect(macdi).toBeCloseTo(combo.macdi[i], 9);
          expect(signal).toBeCloseTo(combo.signal[i], 9);
        }
      });
    });
  });

  describe('no warm-up', () => {
    it('returns exactly 0 on bar 0 and never NaN', () => {
      const ind = new MacdIndex();

      const [macdi0, signal0] = ind.update(testInput[0]);
      expect(macdi0).toBe(0);
      expect(signal0).toBe(0);

      for (let i = 1; i < testInput.length; i++) {
        const [macdi, signal] = ind.update(testInput[i]);
        expect(isNaN(macdi)).toBe(false);
        expect(isNaN(signal)).toBe(false);
      }
    });
  });

  describe('passthrough smoothing', () => {
    it('returns EMA(c,1) - EMA(c,2) when u = ul = 1', () => {
      const ind = new MacdIndex({ r: 2, s: 1, u: 1, ul: 1 });

      // Bar 0 seeds both price EMAs, so the MACD line is exactly 0.
      const [macdi0, signal0] = ind.update(10);
      expect(macdi0).toBeCloseTo(0, 12);
      expect(signal0).toBeCloseTo(0, 12);

      // EMA(1) is a passthrough -> fast = 12. EMA(2) = (2/3)*12 + (1/3)*10 = 11.3333.
      const [macdi1, signal1] = ind.update(12);
      expect(macdi1).toBeCloseTo(12 - (2 / 3 * 12 + 1 / 3 * 10), 12);
      expect(signal1).toBeCloseTo(macdi1, 12);
    });
  });

  describe('signal passthrough', () => {
    it('returns signal === macdi when ul = 1', () => {
      const ind = new MacdIndex({ r: 20, s: 5, u: 3, ul: 1 });

      for (let i = 0; i < testInput.length; i++) {
        const [macdi, signal] = ind.update(testInput[i]);
        expect(signal).toBe(macdi);
      }
    });
  });

  describe('isPrimed', () => {
    it('primes after the first update (no warm-up region)', () => {
      const ind = new MacdIndex();
      expect(ind.isPrimed()).toBe(false);
      ind.update(testInput[0]);
      expect(ind.isPrimed()).toBe(true);
    });
  });

  describe('mnemonic', () => {
    it('formats the default mnemonic (ul excluded)', () => {
      const ind = new MacdIndex();
      expect(ind.metadata().mnemonic).toBe('macdi(20,5,3)');
    });

    it('formats a custom mnemonic', () => {
      const ind = new MacdIndex({ r: 26, s: 12, u: 9, ul: 7 });
      expect(ind.metadata().mnemonic).toBe('macdi(26,12,9)');
    });
  });

  describe('metadata', () => {
    it('reports the identifier and two outputs', () => {
      const ind = new MacdIndex();
      const meta = ind.metadata();
      expect(meta.identifier).toBe(IndicatorIdentifier.MacdIndex);
      expect(meta.outputs.length).toBe(2);
    });
  });

  describe('updateScalar', () => {
    it('returns [macdi, signal] in order', () => {
      const ind = new MacdIndex({ r: 20, s: 5, u: 3, ul: UL });
      let out: ReturnType<typeof ind.updateScalar> = [];
      for (let i = 0; i < testInput.length; i++) {
        const s = new Scalar();
        s.time = new Date(0);
        s.value = testInput[i];
        out = ind.updateScalar(s);
      }
      const last = testInput.length - 1;
      expect((out[0] as Scalar).value).toBeCloseTo(expectedR20_S5_U3[last], 9);
      expect((out[1] as Scalar).value).toBeCloseTo(expectedR20_S5_U3_SIG_UL3[last], 9);
    });
  });

  describe('invalid parameters', () => {
    it('throws when r < 1', () => {
      expect(() => new MacdIndex({ r: 0 })).toThrowError();
    });
    it('throws when s < 1', () => {
      expect(() => new MacdIndex({ s: 0 })).toThrowError();
    });
    it('throws when u < 1', () => {
      expect(() => new MacdIndex({ u: 0 })).toThrowError();
    });
    it('throws when ul < 1', () => {
      expect(() => new MacdIndex({ ul: 0 })).toThrowError();
    });
    it('throws when s is not less than r', () => {
      expect(() => new MacdIndex({ r: 5, s: 5 })).toThrowError();
    });
    it('throws when s is greater than r', () => {
      expect(() => new MacdIndex({ r: 5, s: 6 })).toThrowError();
    });
  });
});
