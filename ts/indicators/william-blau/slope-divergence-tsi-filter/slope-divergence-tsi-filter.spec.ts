import { SlopeDivergenceTsiFilter } from './slope-divergence-tsi-filter';
import { IndicatorIdentifier } from '../../core/indicator-identifier';
import { BarComponent } from '../../../entities/bar-component';
import { QuoteComponent } from '../../../entities/quote-component';
import { TradeComponent } from '../../../entities/trade-component';
import { Bar } from '../../../entities/bar';
import { Quote } from '../../../entities/quote';
import { Trade } from '../../../entities/trade';
import { Scalar } from '../../../entities/scalar';
import {
  testInput,
  expectedR32_S32_U7_X32_Y7,
  expectedR32_S32_U1_X32_Y1,
  expectedR32_S32_U7_X32_Y1,
  expectedR32_S32_U1_X32_Y7,
  expectedR1_S1_U1_X1_Y1,
  expectedR20_S5_U3_X20_Y3,
  expectedR32_S13_U3_X32_Y7,
  expectedR12_S12_U1_X12_Y1,
  expectedR25_S13_U1_X25_Y1,
  expectedR64_S64_U7_X32_Y7,
  expectedR32_S32_U7_X16_Y3,
  expectedR5_S5_U5_X5_Y5,
  expectedR10_S10_U1_X10_Y1,
  expectedR40_S20_U5_X32_Y7,
  expectedR32_S5_U1_X32_Y1,
  expectedR50_S25_U1_X50_Y1,
} from './testdata';

interface Combo {
  name: string;
  r: number;
  s: number;
  u: number;
  x: number;
  y: number;
  expected: number[];
}

// Every expected array uses the book momentum q=2.
const combos: Combo[] = [
  { name: 'R32_S32_U7_X32_Y7', r: 32, s: 32, u: 7, x: 32, y: 7, expected: expectedR32_S32_U7_X32_Y7 },
  { name: 'R32_S32_U1_X32_Y1', r: 32, s: 32, u: 1, x: 32, y: 1, expected: expectedR32_S32_U1_X32_Y1 },
  { name: 'R32_S32_U7_X32_Y1', r: 32, s: 32, u: 7, x: 32, y: 1, expected: expectedR32_S32_U7_X32_Y1 },
  { name: 'R32_S32_U1_X32_Y7', r: 32, s: 32, u: 1, x: 32, y: 7, expected: expectedR32_S32_U1_X32_Y7 },
  { name: 'R1_S1_U1_X1_Y1', r: 1, s: 1, u: 1, x: 1, y: 1, expected: expectedR1_S1_U1_X1_Y1 },
  { name: 'R20_S5_U3_X20_Y3', r: 20, s: 5, u: 3, x: 20, y: 3, expected: expectedR20_S5_U3_X20_Y3 },
  { name: 'R32_S13_U3_X32_Y7', r: 32, s: 13, u: 3, x: 32, y: 7, expected: expectedR32_S13_U3_X32_Y7 },
  { name: 'R12_S12_U1_X12_Y1', r: 12, s: 12, u: 1, x: 12, y: 1, expected: expectedR12_S12_U1_X12_Y1 },
  { name: 'R25_S13_U1_X25_Y1', r: 25, s: 13, u: 1, x: 25, y: 1, expected: expectedR25_S13_U1_X25_Y1 },
  { name: 'R64_S64_U7_X32_Y7', r: 64, s: 64, u: 7, x: 32, y: 7, expected: expectedR64_S64_U7_X32_Y7 },
  { name: 'R32_S32_U7_X16_Y3', r: 32, s: 32, u: 7, x: 16, y: 3, expected: expectedR32_S32_U7_X16_Y3 },
  { name: 'R5_S5_U5_X5_Y5', r: 5, s: 5, u: 5, x: 5, y: 5, expected: expectedR5_S5_U5_X5_Y5 },
  { name: 'R10_S10_U1_X10_Y1', r: 10, s: 10, u: 1, x: 10, y: 1, expected: expectedR10_S10_U1_X10_Y1 },
  { name: 'R40_S20_U5_X32_Y7', r: 40, s: 20, u: 5, x: 32, y: 7, expected: expectedR40_S20_U5_X32_Y7 },
  { name: 'R32_S5_U1_X32_Y1', r: 32, s: 5, u: 1, x: 32, y: 1, expected: expectedR32_S5_U1_X32_Y1 },
  { name: 'R50_S25_U1_X50_Y1', r: 50, s: 25, u: 1, x: 50, y: 1, expected: expectedR50_S25_U1_X50_Y1 },
];

const expectValue = (actual: number, expected: number): void => {
  if (isNaN(expected)) {
    expect(isNaN(actual)).toBe(true);
  } else {
    expect(actual).toBeCloseTo(expected, 13);
  }
};

describe('SlopeDivergenceTsiFilter', () => {
  describe('reference data', () => {
    combos.forEach((combo) => {
      it(`matches the reference for ${combo.name}`, () => {
        const ind = new SlopeDivergenceTsiFilter({
          q: 2, r: combo.r, s: combo.s, u: combo.u, x: combo.x, y: combo.y,
        });

        for (let i = 0; i < testInput.length; i++) {
          expectValue(ind.update(testInput[i]), combo.expected[i]);
        }
      });
    });
  });

  describe('passthrough', () => {
    it('keeps +/-100 only when the TSI and close slopes agree', () => {
      const ind = new SlopeDivergenceTsiFilter({ q: 2, r: 1, s: 1, u: 1, x: 1, y: 1 });

      expect(isNaN(ind.update(10))).toBe(true); // momentum undefined
      expect(ind.update(12)).toBe(0); // first finite TSI, no slope
      expect(ind.update(11)).toBe(-100); // both falling
      expect(ind.update(13)).toBe(100); // both rising
      expect(ind.update(14)).toBe(0); // TSI flat
    });
  });

  describe('warm-up', () => {
    it('returns NaN for bars 0..q-2, then 0 at the first finite bar', () => {
      const q = 5;
      const ind = new SlopeDivergenceTsiFilter({ q });

      for (let i = 0; i < q - 1; i++) {
        expect(isNaN(ind.update(testInput[i]))).toBe(true);
      }

      expect(ind.update(testInput[q - 1])).toBe(0);

      for (let i = q; i < testInput.length; i++) {
        expect(isNaN(ind.update(testInput[i]))).toBe(false);
      }
    });
  });

  describe('bounds', () => {
    it('keeps every finite value within [-100, 100]', () => {
      const ind = new SlopeDivergenceTsiFilter();

      for (let i = 0; i < testInput.length; i++) {
        const value = ind.update(testInput[i]);

        if (!isNaN(value)) {
          expect(value).toBeGreaterThanOrEqual(-100);
          expect(value).toBeLessThanOrEqual(100);
        }
      }
    });
  });

  describe('isPrimed', () => {
    it('primes from the first finite TSI bar', () => {
      const q = 5;
      const ind = new SlopeDivergenceTsiFilter({ q });

      for (let i = 0; i < q - 1; i++) {
        ind.update(testInput[i]);
        expect(ind.isPrimed()).toBe(false);
      }

      for (let i = q - 1; i < testInput.length; i++) {
        ind.update(testInput[i]);
        expect(ind.isPrimed()).toBe(true);
      }
    });
  });

  describe('entity updates', () => {
    const expected = expectedR32_S32_U7_X32_Y7;

    it('updates from scalars', () => {
      const ind = new SlopeDivergenceTsiFilter();

      for (let i = 0; i < testInput.length; i++) {
        const s = new Scalar();
        s.time = new Date(0);
        s.value = testInput[i];
        const out = ind.updateScalar(s);
        expect(out.length).toBe(1);
        expectValue((out[0] as Scalar).value, expected[i]);
      }
    });

    it('updates from bars', () => {
      const ind = new SlopeDivergenceTsiFilter();

      for (let i = 0; i < testInput.length; i++) {
        const b = new Bar();
        b.time = new Date(0);
        b.close = testInput[i];
        expectValue((ind.updateBar(b)[0] as Scalar).value, expected[i]);
      }
    });

    it('updates from quotes', () => {
      const ind = new SlopeDivergenceTsiFilter();

      for (let i = 0; i < testInput.length; i++) {
        const q = new Quote();
        q.time = new Date(0);
        q.bidPrice = testInput[i];
        q.askPrice = testInput[i];
        expectValue((ind.updateQuote(q)[0] as Scalar).value, expected[i]);
      }
    });

    it('updates from trades', () => {
      const ind = new SlopeDivergenceTsiFilter();

      for (let i = 0; i < testInput.length; i++) {
        const t = new Trade();
        t.time = new Date(0);
        t.price = testInput[i];
        expectValue((ind.updateTrade(t)[0] as Scalar).value, expected[i]);
      }
    });
  });

  describe('mnemonic', () => {
    it('formats the default mnemonic', () => {
      const ind = new SlopeDivergenceTsiFilter();
      expect(ind.metadata().mnemonic).toBe('sdtsi(2,32,32,7,32,7)');
      expect(ind.metadata().description).toBe('Slope Divergence TSI Filter sdtsi(2,32,32,7,32,7)');
    });

    it('formats a custom mnemonic', () => {
      const ind = new SlopeDivergenceTsiFilter({ q: 3, r: 20, s: 5, u: 3, x: 20, y: 3 });
      expect(ind.metadata().mnemonic).toBe('sdtsi(3,20,5,3,20,3)');
    });

    it('formats the mnemonic with only the bar component set', () => {
      const ind = new SlopeDivergenceTsiFilter({ barComponent: BarComponent.Median });
      expect(ind.metadata().mnemonic).toBe('sdtsi(2,32,32,7,32,7, hl/2)');
    });

    it('formats the mnemonic with only the quote component set', () => {
      const ind = new SlopeDivergenceTsiFilter({ quoteComponent: QuoteComponent.Bid });
      expect(ind.metadata().mnemonic).toBe('sdtsi(2,32,32,7,32,7, b)');
    });

    it('formats the mnemonic with only the trade component set', () => {
      const ind = new SlopeDivergenceTsiFilter({ tradeComponent: TradeComponent.Volume });
      expect(ind.metadata().mnemonic).toBe('sdtsi(2,32,32,7,32,7, v)');
    });

    it('formats the mnemonic with the bar and quote components set', () => {
      const ind = new SlopeDivergenceTsiFilter({
        barComponent: BarComponent.Open, quoteComponent: QuoteComponent.Bid,
      });
      expect(ind.metadata().mnemonic).toBe('sdtsi(2,32,32,7,32,7, o, b)');
    });

    it('formats the mnemonic with the bar and trade components set', () => {
      const ind = new SlopeDivergenceTsiFilter({
        barComponent: BarComponent.High, tradeComponent: TradeComponent.Volume,
      });
      expect(ind.metadata().mnemonic).toBe('sdtsi(2,32,32,7,32,7, h, v)');
    });

    it('formats the mnemonic with the quote and trade components set', () => {
      const ind = new SlopeDivergenceTsiFilter({
        quoteComponent: QuoteComponent.Ask, tradeComponent: TradeComponent.Volume,
      });
      expect(ind.metadata().mnemonic).toBe('sdtsi(2,32,32,7,32,7, a, v)');
    });
  });

  describe('metadata', () => {
    it('reports the identifier and a single output', () => {
      const ind = new SlopeDivergenceTsiFilter();
      const meta = ind.metadata();
      expect(meta.identifier).toBe(IndicatorIdentifier.SlopeDivergenceTsiFilter);
      expect(meta.mnemonic).toBe('sdtsi(2,32,32,7,32,7)');
      expect(meta.outputs.length).toBe(1);
      expect(meta.outputs[0].mnemonic).toBe('sdtsi(2,32,32,7,32,7)');
      expect(meta.outputs[0].description).toBe('Slope Divergence TSI Filter sdtsi(2,32,32,7,32,7)');
    });
  });

  describe('invalid parameters', () => {
    it('throws when q < 1', () => {
      expect(() => new SlopeDivergenceTsiFilter({ q: 0 })).toThrowError();
    });
    it('throws when r < 1', () => {
      expect(() => new SlopeDivergenceTsiFilter({ r: 0 })).toThrowError();
    });
    it('throws when s < 1', () => {
      expect(() => new SlopeDivergenceTsiFilter({ s: 0 })).toThrowError();
    });
    it('throws when u < 1', () => {
      expect(() => new SlopeDivergenceTsiFilter({ u: 0 })).toThrowError();
    });
    it('throws when x < 1', () => {
      expect(() => new SlopeDivergenceTsiFilter({ x: 0 })).toThrowError();
    });
    it('throws when y < 1', () => {
      expect(() => new SlopeDivergenceTsiFilter({ y: 0 })).toThrowError();
    });
  });
});
