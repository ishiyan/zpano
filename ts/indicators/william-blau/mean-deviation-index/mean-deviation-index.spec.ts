import { MeanDeviationIndex } from './mean-deviation-index';
import { IndicatorIdentifier } from '../../core/indicator-identifier';
import { Scalar } from '../../../entities/scalar';
import {
  testInput,
  expectedR20_S5_U3, expectedR20_S5_U3_SIG_UL3,
  expectedR20_S5_U1, expectedR20_S5_U1_SIG_UL3,
  expectedR1_S5_U3, expectedR1_S5_U3_SIG_UL3,
  expectedR40_S5_U3, expectedR40_S5_U3_SIG_UL3,
  expectedR10_S5_U3, expectedR10_S5_U3_SIG_UL3,
  expectedR5_S5_U5, expectedR5_S5_U5_SIG_UL3,
  expectedR20_S9_U1, expectedR20_S9_U1_SIG_UL3,
  expectedR26_S12_U9, expectedR26_S12_U9_SIG_UL3,
  expectedR50_S13_U1, expectedR50_S13_U1_SIG_UL3,
  expectedR30_S5_U3, expectedR30_S5_U3_SIG_UL3,
  expectedR3_S3_U3, expectedR3_S3_U3_SIG_UL3,
  expectedR7_S4_U2, expectedR7_S4_U2_SIG_UL3,
  expectedR2_S5_U3, expectedR2_S5_U3_SIG_UL3,
  expectedR20_S1_U1, expectedR20_S1_U1_SIG_UL3,
  expectedR20_S20_U5, expectedR20_S20_U5_SIG_UL3,
  expectedR60_S30_U10, expectedR60_S30_U10_SIG_UL3,
} from './testdata';

// Signal-line EMA period used for every expected signal array.
const UL = 3;

interface Combo {
  name: string;
  r: number;
  s: number;
  u: number;
  mdi: number[];
  signal: number[];
}

const combos: Combo[] = [
  { name: 'R20_S5_U3', r: 20, s: 5, u: 3, mdi: expectedR20_S5_U3, signal: expectedR20_S5_U3_SIG_UL3 },
  { name: 'R20_S5_U1', r: 20, s: 5, u: 1, mdi: expectedR20_S5_U1, signal: expectedR20_S5_U1_SIG_UL3 },
  { name: 'R1_S5_U3', r: 1, s: 5, u: 3, mdi: expectedR1_S5_U3, signal: expectedR1_S5_U3_SIG_UL3 },
  { name: 'R40_S5_U3', r: 40, s: 5, u: 3, mdi: expectedR40_S5_U3, signal: expectedR40_S5_U3_SIG_UL3 },
  { name: 'R10_S5_U3', r: 10, s: 5, u: 3, mdi: expectedR10_S5_U3, signal: expectedR10_S5_U3_SIG_UL3 },
  { name: 'R5_S5_U5', r: 5, s: 5, u: 5, mdi: expectedR5_S5_U5, signal: expectedR5_S5_U5_SIG_UL3 },
  { name: 'R20_S9_U1', r: 20, s: 9, u: 1, mdi: expectedR20_S9_U1, signal: expectedR20_S9_U1_SIG_UL3 },
  { name: 'R26_S12_U9', r: 26, s: 12, u: 9, mdi: expectedR26_S12_U9, signal: expectedR26_S12_U9_SIG_UL3 },
  { name: 'R50_S13_U1', r: 50, s: 13, u: 1, mdi: expectedR50_S13_U1, signal: expectedR50_S13_U1_SIG_UL3 },
  { name: 'R30_S5_U3', r: 30, s: 5, u: 3, mdi: expectedR30_S5_U3, signal: expectedR30_S5_U3_SIG_UL3 },
  { name: 'R3_S3_U3', r: 3, s: 3, u: 3, mdi: expectedR3_S3_U3, signal: expectedR3_S3_U3_SIG_UL3 },
  { name: 'R7_S4_U2', r: 7, s: 4, u: 2, mdi: expectedR7_S4_U2, signal: expectedR7_S4_U2_SIG_UL3 },
  { name: 'R2_S5_U3', r: 2, s: 5, u: 3, mdi: expectedR2_S5_U3, signal: expectedR2_S5_U3_SIG_UL3 },
  { name: 'R20_S1_U1', r: 20, s: 1, u: 1, mdi: expectedR20_S1_U1, signal: expectedR20_S1_U1_SIG_UL3 },
  { name: 'R20_S20_U5', r: 20, s: 20, u: 5, mdi: expectedR20_S20_U5, signal: expectedR20_S20_U5_SIG_UL3 },
  { name: 'R60_S30_U10', r: 60, s: 30, u: 10, mdi: expectedR60_S30_U10, signal: expectedR60_S30_U10_SIG_UL3 },
];

describe('MeanDeviationIndex', () => {
  describe('reference data', () => {
    combos.forEach((combo) => {
      it(`matches the reference for ${combo.name}`, () => {
        const ind = new MeanDeviationIndex({
          r: combo.r, s: combo.s, u: combo.u, ul: UL,
        });

        for (let i = 0; i < testInput.length; i++) {
          const [mdi, signal] = ind.update(testInput[i]);

          expect(mdi).toBeCloseTo(combo.mdi[i], 9);
          expect(signal).toBeCloseTo(combo.signal[i], 9);
        }
      });
    });
  });

  describe('no warm-up', () => {
    it('returns exactly 0 on bar 0 and never NaN', () => {
      const ind = new MeanDeviationIndex();

      const [mdi0, signal0] = ind.update(testInput[0]);
      expect(mdi0).toBe(0);
      expect(signal0).toBe(0);

      for (let i = 1; i < testInput.length; i++) {
        const [mdi, signal] = ind.update(testInput[i]);
        expect(isNaN(mdi)).toBe(false);
        expect(isNaN(signal)).toBe(false);
      }
    });
  });

  describe('degenerate r = 1', () => {
    it('returns identically 0 because the detrend is a passthrough', () => {
      const ind = new MeanDeviationIndex({ r: 1, s: 5, u: 3, ul: 3 });

      for (let i = 0; i < testInput.length; i++) {
        const [mdi, signal] = ind.update(testInput[i]);
        expect(mdi).toBe(0);
        expect(signal).toBe(0);
      }
    });
  });

  describe('passthrough smoothing', () => {
    it('returns price - EMA(price, 2) when s = u = ul = 1', () => {
      const ind = new MeanDeviationIndex({ r: 2, s: 1, u: 1, ul: 1 });

      // Bar 0 seeds the baseline, so the deviation is exactly 0.
      const [mdi0, signal0] = ind.update(10);
      expect(mdi0).toBeCloseTo(0, 12);
      expect(signal0).toBeCloseTo(0, 12);

      // EMA(2) = (2/3)*13 + (1/3)*10 = 12, so the deviation is 13-12 = 1.
      const [mdi1, signal1] = ind.update(13);
      expect(mdi1).toBeCloseTo(1, 12);
      expect(signal1).toBeCloseTo(1, 12);
    });
  });

  describe('signal passthrough', () => {
    it('returns signal === mdi when ul = 1', () => {
      const ind = new MeanDeviationIndex({ r: 20, s: 5, u: 3, ul: 1 });

      for (let i = 0; i < testInput.length; i++) {
        const [mdi, signal] = ind.update(testInput[i]);
        expect(signal).toBe(mdi);
      }
    });
  });

  describe('isPrimed', () => {
    it('primes after the first update (no warm-up region)', () => {
      const ind = new MeanDeviationIndex();
      expect(ind.isPrimed()).toBe(false);
      ind.update(testInput[0]);
      expect(ind.isPrimed()).toBe(true);
    });
  });

  describe('mnemonic', () => {
    it('formats the default mnemonic (ul excluded)', () => {
      const ind = new MeanDeviationIndex();
      expect(ind.metadata().mnemonic).toBe('mdi(20,5,3)');
    });

    it('formats a custom mnemonic', () => {
      const ind = new MeanDeviationIndex({ r: 26, s: 12, u: 9, ul: 7 });
      expect(ind.metadata().mnemonic).toBe('mdi(26,12,9)');
    });
  });

  describe('metadata', () => {
    it('reports the identifier and two outputs', () => {
      const ind = new MeanDeviationIndex();
      const meta = ind.metadata();
      expect(meta.identifier).toBe(IndicatorIdentifier.MeanDeviationIndex);
      expect(meta.outputs.length).toBe(2);
    });
  });

  describe('updateScalar', () => {
    it('returns [mdi, signal] in order', () => {
      const ind = new MeanDeviationIndex({ r: 20, s: 5, u: 3, ul: UL });
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
      expect(() => new MeanDeviationIndex({ r: 0 })).toThrowError();
    });
    it('throws when s < 1', () => {
      expect(() => new MeanDeviationIndex({ s: 0 })).toThrowError();
    });
    it('throws when u < 1', () => {
      expect(() => new MeanDeviationIndex({ u: 0 })).toThrowError();
    });
    it('throws when ul < 1', () => {
      expect(() => new MeanDeviationIndex({ ul: 0 })).toThrowError();
    });
  });
});
