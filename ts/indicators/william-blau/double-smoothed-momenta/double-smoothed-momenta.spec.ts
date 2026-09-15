import { DoubleSmoothedMomenta } from './double-smoothed-momenta';
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
  expectedA2_Y2_Z14,
  expectedA2_Y1_Z14,
  expectedA2_Y1_Z9,
  expectedA2_Y1_Z2,
  expectedA2_Y3_Z9,
  expectedA2_Y5_Z5,
  expectedA2_Y2_Z5,
  expectedA2_Y1_Z1,
  expectedA1_Y1_Z1,
  expectedA5_Y2_Z14,
  expectedA10_Y3_Z5,
  expectedA14_Y2_Z9,
  expectedA20_Y5_Z3,
  expectedA3_Y3_Z3,
  expectedA7_Y4_Z2,
  expectedA32_Y2_Z7,
} from './testdata';

interface Combo {
  name: string;
  a: number;
  y: number;
  z: number;
  expected: number[];
}

const combos: Combo[] = [
  { name: 'A2_Y2_Z14', a: 2, y: 2, z: 14, expected: expectedA2_Y2_Z14 },
  { name: 'A2_Y1_Z14', a: 2, y: 1, z: 14, expected: expectedA2_Y1_Z14 },
  { name: 'A2_Y1_Z9', a: 2, y: 1, z: 9, expected: expectedA2_Y1_Z9 },
  { name: 'A2_Y1_Z2', a: 2, y: 1, z: 2, expected: expectedA2_Y1_Z2 },
  { name: 'A2_Y3_Z9', a: 2, y: 3, z: 9, expected: expectedA2_Y3_Z9 },
  { name: 'A2_Y5_Z5', a: 2, y: 5, z: 5, expected: expectedA2_Y5_Z5 },
  { name: 'A2_Y2_Z5', a: 2, y: 2, z: 5, expected: expectedA2_Y2_Z5 },
  { name: 'A2_Y1_Z1', a: 2, y: 1, z: 1, expected: expectedA2_Y1_Z1 },
  { name: 'A1_Y1_Z1', a: 1, y: 1, z: 1, expected: expectedA1_Y1_Z1 },
  { name: 'A5_Y2_Z14', a: 5, y: 2, z: 14, expected: expectedA5_Y2_Z14 },
  { name: 'A10_Y3_Z5', a: 10, y: 3, z: 5, expected: expectedA10_Y3_Z5 },
  { name: 'A14_Y2_Z9', a: 14, y: 2, z: 9, expected: expectedA14_Y2_Z9 },
  { name: 'A20_Y5_Z3', a: 20, y: 5, z: 3, expected: expectedA20_Y5_Z3 },
  { name: 'A3_Y3_Z3', a: 3, y: 3, z: 3, expected: expectedA3_Y3_Z3 },
  { name: 'A7_Y4_Z2', a: 7, y: 4, z: 2, expected: expectedA7_Y4_Z2 },
  { name: 'A32_Y2_Z7', a: 32, y: 2, z: 7, expected: expectedA32_Y2_Z7 },
];

/** Independently-coded EMA-form RSI: 100 * EMA(up, z) / EMA(up + dn, z). */
const emaFormRsi = (closes: number[], z: number): number[] => {
  const alpha = 2 / (z + 1);
  const result: number[] = [Number.NaN];

  let numerator = 0;
  let denominator = 0;
  let primed = false;

  for (let k = 1; k < closes.length; k++) {
    const diff = closes[k] - closes[k - 1];
    const up = diff > 0 ? diff : 0;
    const dn = diff < 0 ? -diff : 0;

    if (primed) {
      numerator = alpha * up + (1 - alpha) * numerator;
      denominator = alpha * (up + dn) + (1 - alpha) * denominator;
    } else {
      numerator = up;
      denominator = up + dn;
      primed = true;
    }

    result.push(denominator <= 0 ? 0 : 100 * numerator / denominator);
  }

  return result;
};

describe('DoubleSmoothedMomenta', () => {
  describe('reference data', () => {
    combos.forEach((combo) => {
      it(`matches the reference for ${combo.name}`, () => {
        const ind = new DoubleSmoothedMomenta({ a: combo.a, y: combo.y, z: combo.z });

        for (let i = 0; i < testInput.length; i++) {
          const value = ind.update(testInput[i]);

          if (isNaN(combo.expected[i])) {
            expect(isNaN(value)).toBe(true);
          } else {
            expect(value).toBeCloseTo(combo.expected[i], 9);
          }
        }
      });
    });
  });

  describe('warm-up', () => {
    it('returns NaN for bars 0..a-2 and a finite value from bar a-1', () => {
      const a = 5;
      const ind = new DoubleSmoothedMomenta({ a, y: 2, z: 14 });

      for (let i = 0; i < a - 1; i++) {
        expect(isNaN(ind.update(testInput[i]))).toBe(true);
      }

      for (let i = a - 1; i < testInput.length; i++) {
        expect(isNaN(ind.update(testInput[i]))).toBe(false);
      }
    });

    it('has no warm-up region when a = 1', () => {
      const ind = new DoubleSmoothedMomenta({ a: 1, y: 1, z: 1 });

      for (let i = 0; i < testInput.length; i++) {
        expect(isNaN(ind.update(testInput[i]))).toBe(false);
      }
    });
  });

  describe('bounds', () => {
    it('keeps every finite value within [0, 100]', () => {
      const ind = new DoubleSmoothedMomenta();

      for (let i = 0; i < testInput.length; i++) {
        const value = ind.update(testInput[i]);

        if (!isNaN(value)) {
          expect(value).toBeGreaterThanOrEqual(0);
          expect(value).toBeLessThanOrEqual(100);
        }
      }
    });
  });

  describe('degenerate a = 1', () => {
    it('returns identically 0 because the 1-bar close range is 0', () => {
      const ind = new DoubleSmoothedMomenta({ a: 1, y: 1, z: 1 });

      for (let i = 0; i < testInput.length; i++) {
        expect(ind.update(testInput[i])).toBe(0);
      }
    });
  });

  describe('RSI equivalence', () => {
    [1, 2, 9, 14].forEach((z) => {
      it(`matches the EMA-form RSI(${z}) for DM(2,1,${z})`, () => {
        const expected = emaFormRsi(testInput, z);
        const ind = new DoubleSmoothedMomenta({ a: 2, y: 1, z });

        for (let i = 0; i < testInput.length; i++) {
          const value = ind.update(testInput[i]);

          if (isNaN(expected[i])) {
            expect(isNaN(value)).toBe(true);
          } else {
            expect(value).toBeCloseTo(expected[i], 9);
          }
        }
      });
    });
  });

  describe('isPrimed', () => {
    it('primes once a closes have been seen', () => {
      const a = 5;
      const ind = new DoubleSmoothedMomenta({ a, y: 2, z: 14 });

      for (let i = 0; i < a - 1; i++) {
        ind.update(testInput[i]);
        expect(ind.isPrimed()).toBe(false);
      }

      for (let i = a - 1; i < testInput.length; i++) {
        ind.update(testInput[i]);
        expect(ind.isPrimed()).toBe(true);
      }
    });
  });

  describe('entity updates', () => {
    const last = testInput.length - 1;

    it('updates from scalars', () => {
      const ind = new DoubleSmoothedMomenta();
      let out: ReturnType<typeof ind.updateScalar> = [];

      for (let i = 0; i < testInput.length; i++) {
        const s = new Scalar();
        s.time = new Date(0);
        s.value = testInput[i];
        out = ind.updateScalar(s);
      }

      expect((out[0] as Scalar).value).toBeCloseTo(expectedA2_Y2_Z14[last], 9);
    });

    it('updates from bars', () => {
      const ind = new DoubleSmoothedMomenta();
      let out: ReturnType<typeof ind.updateBar> = [];

      for (let i = 0; i < testInput.length; i++) {
        const b = new Bar();
        b.time = new Date(0);
        b.close = testInput[i];
        out = ind.updateBar(b);
      }

      expect((out[0] as Scalar).value).toBeCloseTo(expectedA2_Y2_Z14[last], 9);
    });

    it('updates from quotes', () => {
      const ind = new DoubleSmoothedMomenta();
      let out: ReturnType<typeof ind.updateQuote> = [];

      for (let i = 0; i < testInput.length; i++) {
        const q = new Quote();
        q.time = new Date(0);
        q.bidPrice = testInput[i];
        q.askPrice = testInput[i];
        out = ind.updateQuote(q);
      }

      expect((out[0] as Scalar).value).toBeCloseTo(expectedA2_Y2_Z14[last], 9);
    });

    it('updates from trades', () => {
      const ind = new DoubleSmoothedMomenta();
      let out: ReturnType<typeof ind.updateTrade> = [];

      for (let i = 0; i < testInput.length; i++) {
        const t = new Trade();
        t.time = new Date(0);
        t.price = testInput[i];
        out = ind.updateTrade(t);
      }

      expect((out[0] as Scalar).value).toBeCloseTo(expectedA2_Y2_Z14[last], 9);
    });
  });

  describe('mnemonic', () => {
    it('formats the default mnemonic', () => {
      const ind = new DoubleSmoothedMomenta();
      expect(ind.metadata().mnemonic).toBe('dm(2,2,14)');
      expect(ind.metadata().description).toBe('Double-Smoothed Momenta dm(2,2,14)');
    });

    it('formats a custom mnemonic', () => {
      const ind = new DoubleSmoothedMomenta({ a: 10, y: 3, z: 5 });
      expect(ind.metadata().mnemonic).toBe('dm(10,3,5)');
    });

    it('formats the mnemonic with only the bar component set', () => {
      const ind = new DoubleSmoothedMomenta({ barComponent: BarComponent.Median });
      expect(ind.metadata().mnemonic).toBe('dm(2,2,14, hl/2)');
    });

    it('formats the mnemonic with only the quote component set', () => {
      const ind = new DoubleSmoothedMomenta({ quoteComponent: QuoteComponent.Bid });
      expect(ind.metadata().mnemonic).toBe('dm(2,2,14, b)');
    });

    it('formats the mnemonic with only the trade component set', () => {
      const ind = new DoubleSmoothedMomenta({ tradeComponent: TradeComponent.Volume });
      expect(ind.metadata().mnemonic).toBe('dm(2,2,14, v)');
    });

    it('formats the mnemonic with the bar and quote components set', () => {
      const ind = new DoubleSmoothedMomenta({
        barComponent: BarComponent.Open, quoteComponent: QuoteComponent.Bid,
      });
      expect(ind.metadata().mnemonic).toBe('dm(2,2,14, o, b)');
    });

    it('formats the mnemonic with the bar and trade components set', () => {
      const ind = new DoubleSmoothedMomenta({
        barComponent: BarComponent.High, tradeComponent: TradeComponent.Volume,
      });
      expect(ind.metadata().mnemonic).toBe('dm(2,2,14, h, v)');
    });

    it('formats the mnemonic with the quote and trade components set', () => {
      const ind = new DoubleSmoothedMomenta({
        quoteComponent: QuoteComponent.Ask, tradeComponent: TradeComponent.Volume,
      });
      expect(ind.metadata().mnemonic).toBe('dm(2,2,14, a, v)');
    });
  });

  describe('metadata', () => {
    it('reports the identifier and a single output', () => {
      const ind = new DoubleSmoothedMomenta();
      const meta = ind.metadata();
      expect(meta.identifier).toBe(IndicatorIdentifier.DoubleSmoothedMomenta);
      expect(meta.outputs.length).toBe(1);
      expect(meta.outputs[0].mnemonic).toBe('dm(2,2,14)');
    });
  });

  describe('invalid parameters', () => {
    it('throws when a < 1', () => {
      expect(() => new DoubleSmoothedMomenta({ a: 0 })).toThrowError();
    });
    it('throws when y < 1', () => {
      expect(() => new DoubleSmoothedMomenta({ y: 0 })).toThrowError();
    });
    it('throws when z < 1', () => {
      expect(() => new DoubleSmoothedMomenta({ z: 0 })).toThrowError();
    });
  });
});
