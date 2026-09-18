import { CandlestickStrengthIndex } from './candlestick-strength-index';
import { IndicatorIdentifier } from '../../core/indicator-identifier';
import { Bar } from '../../../entities/bar';
import { Quote } from '../../../entities/quote';
import { Scalar } from '../../../entities/scalar';
import { Trade } from '../../../entities/trade';
import {
  testInput, testOpen, testHigh, testLow,
  expectedR20_S5_U3, expectedR20_S5_U3_SIG_UL3,
  expectedR32_S32_U1, expectedR32_S32_U1_SIG_UL3,
  expectedR1_S1_U1, expectedR1_S1_U1_SIG_UL3,
  expectedR25_S13_U1, expectedR25_S13_U1_SIG_UL3,
  expectedR13_S13_U1, expectedR13_S13_U1_SIG_UL3,
  expectedR5_S5_U5, expectedR5_S5_U5_SIG_UL3,
  expectedR9_S3_U1, expectedR9_S3_U1_SIG_UL3,
  expectedR64_S64_U1, expectedR64_S64_U1_SIG_UL3,
  expectedR32_S32_U3, expectedR32_S32_U3_SIG_UL3,
  expectedR40_S20_U1, expectedR40_S20_U1_SIG_UL3,
  expectedR2_S2_U2, expectedR2_S2_U2_SIG_UL3,
  expectedR7_S4_U2, expectedR7_S4_U2_SIG_UL3,
  expectedR12_S12_U12, expectedR12_S12_U12_SIG_UL3,
  expectedR3_S10_U10, expectedR3_S10_U10_SIG_UL3,
  expectedR50_S1_U1, expectedR50_S1_U1_SIG_UL3,
  expectedR32_S5_U3, expectedR32_S5_U3_SIG_UL3,
} from './testdata';

// Signal-line EMA period used for every expected signal array.
const UL = 3;

interface Combo {
  name: string;
  r: number;
  s: number;
  u: number;
  csi: number[];
  signal: number[];
}

const combos: Combo[] = [
  { name: 'R20_S5_U3', r: 20, s: 5, u: 3, csi: expectedR20_S5_U3, signal: expectedR20_S5_U3_SIG_UL3 },
  { name: 'R32_S32_U1', r: 32, s: 32, u: 1, csi: expectedR32_S32_U1, signal: expectedR32_S32_U1_SIG_UL3 },
  { name: 'R1_S1_U1', r: 1, s: 1, u: 1, csi: expectedR1_S1_U1, signal: expectedR1_S1_U1_SIG_UL3 },
  { name: 'R25_S13_U1', r: 25, s: 13, u: 1, csi: expectedR25_S13_U1, signal: expectedR25_S13_U1_SIG_UL3 },
  { name: 'R13_S13_U1', r: 13, s: 13, u: 1, csi: expectedR13_S13_U1, signal: expectedR13_S13_U1_SIG_UL3 },
  { name: 'R5_S5_U5', r: 5, s: 5, u: 5, csi: expectedR5_S5_U5, signal: expectedR5_S5_U5_SIG_UL3 },
  { name: 'R9_S3_U1', r: 9, s: 3, u: 1, csi: expectedR9_S3_U1, signal: expectedR9_S3_U1_SIG_UL3 },
  { name: 'R64_S64_U1', r: 64, s: 64, u: 1, csi: expectedR64_S64_U1, signal: expectedR64_S64_U1_SIG_UL3 },
  { name: 'R32_S32_U3', r: 32, s: 32, u: 3, csi: expectedR32_S32_U3, signal: expectedR32_S32_U3_SIG_UL3 },
  { name: 'R40_S20_U1', r: 40, s: 20, u: 1, csi: expectedR40_S20_U1, signal: expectedR40_S20_U1_SIG_UL3 },
  { name: 'R2_S2_U2', r: 2, s: 2, u: 2, csi: expectedR2_S2_U2, signal: expectedR2_S2_U2_SIG_UL3 },
  { name: 'R7_S4_U2', r: 7, s: 4, u: 2, csi: expectedR7_S4_U2, signal: expectedR7_S4_U2_SIG_UL3 },
  { name: 'R12_S12_U12', r: 12, s: 12, u: 12, csi: expectedR12_S12_U12, signal: expectedR12_S12_U12_SIG_UL3 },
  { name: 'R3_S10_U10', r: 3, s: 10, u: 10, csi: expectedR3_S10_U10, signal: expectedR3_S10_U10_SIG_UL3 },
  { name: 'R50_S1_U1', r: 50, s: 1, u: 1, csi: expectedR50_S1_U1, signal: expectedR50_S1_U1_SIG_UL3 },
  { name: 'R32_S5_U3', r: 32, s: 5, u: 3, csi: expectedR32_S5_U3, signal: expectedR32_S5_U3_SIG_UL3 },
];

describe('CandlestickStrengthIndex', () => {
  describe('reference data', () => {
    combos.forEach((combo) => {
      it(`matches the reference for ${combo.name}`, () => {
        const ind = new CandlestickStrengthIndex({
          r: combo.r, s: combo.s, u: combo.u, ul: UL,
        });

        for (let i = 0; i < testInput.length; i++) {
          const [csi, signal] = ind.update(testOpen[i], testHigh[i], testLow[i], testInput[i]);

          if (isNaN(combo.csi[i])) {
            expect(csi).toBeNaN();
          } else {
            expect(csi).toBeCloseTo(combo.csi[i], 9);
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
    it('reduces to 100*(close-open)/(high-low) when every stage is a passthrough', () => {
      const ind = new CandlestickStrengthIndex({ r: 1, s: 1, u: 1, ul: 1 });

      expect(ind.update(10, 12, 10, 12)).toEqual([100, 100]);  // body spans the whole range up
      expect(ind.update(12, 12, 10, 10)).toEqual([-100, -100]); // body spans the whole range down
      expect(ind.update(11, 11, 11, 11)[0]).toBe(0);            // zero range -> division guard
    });
  });

  describe('isPrimed', () => {
    it('is primed after the first bar (no NaN warm-up)', () => {
      const ind = new CandlestickStrengthIndex();
      expect(ind.isPrimed()).toBe(false);
      ind.update(testOpen[0], testHigh[0], testLow[0], testInput[0]);
      expect(ind.isPrimed()).toBe(true);
    });
  });

  describe('mnemonic', () => {
    it('formats the default mnemonic', () => {
      const ind = new CandlestickStrengthIndex();
      expect(ind.metadata().mnemonic).toBe('csi(20,5,3,3)');
      expect(ind.metadata().description).toBe('Candlestick Strength Index csi(20,5,3,3)');
    });

    it('formats a custom mnemonic', () => {
      const ind = new CandlestickStrengthIndex({ r: 25, s: 13, u: 1, ul: 7 });
      expect(ind.metadata().mnemonic).toBe('csi(25,13,1,7)');
    });
  });

  describe('metadata', () => {
    it('reports the identifier and two outputs', () => {
      const ind = new CandlestickStrengthIndex();
      const meta = ind.metadata();
      expect(meta.identifier).toBe(IndicatorIdentifier.CandlestickStrengthIndex);
      expect(meta.outputs.length).toBe(2);
    });
  });

  describe('updateBar', () => {
    it('returns [csi, signal] in order', () => {
      const ind = new CandlestickStrengthIndex({ r: 20, s: 5, u: 3, ul: UL });
      let out: ReturnType<typeof ind.updateBar> = [];
      for (let i = 0; i < testInput.length; i++) {
        const bar = new Bar();
        bar.time = new Date(0);
        bar.open = testOpen[i];
        bar.high = testHigh[i];
        bar.low = testLow[i];
        bar.close = testInput[i];
        bar.volume = 0;
        out = ind.updateBar(bar);
      }
      const last = testInput.length - 1;
      expect((out[0] as Scalar).value).toBeCloseTo(expectedR20_S5_U3[last], 9);
      expect((out[1] as Scalar).value).toBeCloseTo(expectedR20_S5_U3_SIG_UL3[last], 9);
    });
  });

  describe('updateQuote', () => {
    it('maps the bid to the open and the low, and the ask to the close and the high', () => {
      const ind = new CandlestickStrengthIndex({ r: 1, s: 1, u: 1, ul: 1 });
      const quote = new Quote();
      quote.time = new Date(0);
      quote.bidPrice = 10;
      quote.askPrice = 12;
      const out = ind.updateQuote(quote);
      expect((out[0] as Scalar).value).toBe(100);
      expect((out[1] as Scalar).value).toBe(100);
    });
  });

  describe('updateScalar', () => {
    it('has a zero candle body and a zero range', () => {
      const ind = new CandlestickStrengthIndex({ r: 1, s: 1, u: 1, ul: 1 });
      const s = new Scalar();
      s.time = new Date(0);
      s.value = 10;
      const out = ind.updateScalar(s);
      expect((out[0] as Scalar).value).toBe(0);
      expect((out[1] as Scalar).value).toBe(0);
    });
  });

  describe('updateTrade', () => {
    it('has a zero candle body and a zero range', () => {
      const ind = new CandlestickStrengthIndex({ r: 1, s: 1, u: 1, ul: 1 });
      const trade = new Trade();
      trade.time = new Date(0);
      trade.price = 10;
      trade.volume = 1;
      const out = ind.updateTrade(trade);
      expect((out[0] as Scalar).value).toBe(0);
      expect((out[1] as Scalar).value).toBe(0);
    });
  });

  describe('invalid parameters', () => {
    it('throws when r < 1', () => {
      expect(() => new CandlestickStrengthIndex({ r: 0 })).toThrowError();
    });
    it('throws when s < 1', () => {
      expect(() => new CandlestickStrengthIndex({ s: 0 })).toThrowError();
    });
    it('throws when u < 1', () => {
      expect(() => new CandlestickStrengthIndex({ u: 0 })).toThrowError();
    });
    it('throws when ul < 1', () => {
      expect(() => new CandlestickStrengthIndex({ ul: 0 })).toThrowError();
    });
  });
});
