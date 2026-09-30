import { NonambiguousTrendFilter } from './nonambiguous-trend-filter';
import { NonambiguousTrendFilterBase } from './params';
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
  testHigh,
  testLow,
  testOpen,
  expectedTSI_R32_S13_U3,
  expectedTSI_R20_S5_U3,
  expectedTSI_R40_S20_U5,
  expectedSMI_Q32_R64_S7_U1,
  expectedSMI_Q5_R20_S5_U3,
  expectedSMI_Q13_R25_S2_U1,
  expectedDTI_Q2_R28_S28_U5,
  expectedDTI_Q2_R20_S5_U3,
  expectedDTI_Q4_R14_S14_U3,
  expectedMDI_R20_S5_U3,
  expectedMDI_R40_S5_U3,
  expectedCMI_R20_S5_U3,
  expectedCMI_R10_S5_U3,
  expectedCSI_R32_S32_U1,
  expectedCSI_R20_S5_U3,
  expectedCSI_R1_S1_U1,
  expectedTVI_R32_S32_U5,
  expectedTVI_R12_S12_U1,
  expectedTVI_R25_S13_U1,
} from './testdata';

const B = NonambiguousTrendFilterBase;

interface Combo {
  name: string;
  base: NonambiguousTrendFilterBase;
  q: number;
  r: number;
  s: number;
  u: number;
  expected: number[];
}

// q=0 where the base does not use it.
const combos: Combo[] = [
  { name: 'TSI_R32_S13_U3', base: B.Tsi, q: 2, r: 32, s: 13, u: 3, expected: expectedTSI_R32_S13_U3 },
  { name: 'TSI_R20_S5_U3', base: B.Tsi, q: 2, r: 20, s: 5, u: 3, expected: expectedTSI_R20_S5_U3 },
  { name: 'TSI_R40_S20_U5', base: B.Tsi, q: 2, r: 40, s: 20, u: 5, expected: expectedTSI_R40_S20_U5 },
  { name: 'SMI_Q32_R64_S7_U1', base: B.Smi, q: 32, r: 64, s: 7, u: 1, expected: expectedSMI_Q32_R64_S7_U1 },
  { name: 'SMI_Q5_R20_S5_U3', base: B.Smi, q: 5, r: 20, s: 5, u: 3, expected: expectedSMI_Q5_R20_S5_U3 },
  { name: 'SMI_Q13_R25_S2_U1', base: B.Smi, q: 13, r: 25, s: 2, u: 1, expected: expectedSMI_Q13_R25_S2_U1 },
  { name: 'DTI_Q2_R28_S28_U5', base: B.Dti, q: 2, r: 28, s: 28, u: 5, expected: expectedDTI_Q2_R28_S28_U5 },
  { name: 'DTI_Q2_R20_S5_U3', base: B.Dti, q: 2, r: 20, s: 5, u: 3, expected: expectedDTI_Q2_R20_S5_U3 },
  { name: 'DTI_Q4_R14_S14_U3', base: B.Dti, q: 4, r: 14, s: 14, u: 3, expected: expectedDTI_Q4_R14_S14_U3 },
  { name: 'MDI_R20_S5_U3', base: B.Mdi, q: 0, r: 20, s: 5, u: 3, expected: expectedMDI_R20_S5_U3 },
  { name: 'MDI_R40_S5_U3', base: B.Mdi, q: 0, r: 40, s: 5, u: 3, expected: expectedMDI_R40_S5_U3 },
  { name: 'CMI_R20_S5_U3', base: B.Cmi, q: 0, r: 20, s: 5, u: 3, expected: expectedCMI_R20_S5_U3 },
  { name: 'CMI_R10_S5_U3', base: B.Cmi, q: 0, r: 10, s: 5, u: 3, expected: expectedCMI_R10_S5_U3 },
  { name: 'CSI_R32_S32_U1', base: B.Csi, q: 0, r: 32, s: 32, u: 1, expected: expectedCSI_R32_S32_U1 },
  { name: 'CSI_R20_S5_U3', base: B.Csi, q: 0, r: 20, s: 5, u: 3, expected: expectedCSI_R20_S5_U3 },
  { name: 'CSI_R1_S1_U1', base: B.Csi, q: 0, r: 1, s: 1, u: 1, expected: expectedCSI_R1_S1_U1 },
  { name: 'TVI_R32_S32_U5', base: B.Tvi, q: 0, r: 32, s: 32, u: 5, expected: expectedTVI_R32_S32_U5 },
  { name: 'TVI_R12_S12_U1', base: B.Tvi, q: 0, r: 12, s: 12, u: 1, expected: expectedTVI_R12_S12_U1 },
  { name: 'TVI_R25_S13_U1', base: B.Tvi, q: 0, r: 25, s: 13, u: 1, expected: expectedTVI_R25_S13_U1 },
];

const testTime = new Date(0);

const makeBar = (open: number, high: number, low: number, close: number): Bar => {
  const b = new Bar();
  b.time = testTime;
  b.open = open;
  b.high = high;
  b.low = low;
  b.close = close;
  return b;
};

const testBar = (i: number): Bar => makeBar(testOpen[i], testHigh[i], testLow[i], testInput[i]);

const expectValue = (actual: number, expected: number): void => {
  if (isNaN(expected)) {
    expect(isNaN(actual)).toBe(true);
  } else {
    expect(actual).toBeCloseTo(expected, 13);
  }
};

describe('NonambiguousTrendFilter', () => {
  describe('reference data', () => {
    combos.forEach((combo) => {
      it(`matches the reference for ${combo.name} from bars`, () => {
        const ind = new NonambiguousTrendFilter({ base: combo.base, q: combo.q, r: combo.r, s: combo.s, u: combo.u });

        for (let i = 0; i < testInput.length; i++) {
          const out = ind.updateBar(testBar(i));
          expect(out.length).toBe(1);
          expect((out[0] as Scalar).time).toBe(testTime);
          expectValue((out[0] as Scalar).value, combo.expected[i]);
        }
      });
    });

    combos.filter((c) => c.base === B.Tsi || c.base === B.Mdi).forEach((combo) => {
      it(`matches the reference for ${combo.name} from samples`, () => {
        const ind = new NonambiguousTrendFilter({ base: combo.base, q: combo.q, r: combo.r, s: combo.s, u: combo.u });

        for (let i = 0; i < testInput.length; i++) {
          expectValue(ind.update(testInput[i]), combo.expected[i]);
        }
      });
    });
  });

  describe('rule', () => {
    it('keeps positive-rising and negative-falling values of a passthrough CSI base', () => {
      // A passthrough CSI base is 100*(C-O)/(H-L).
      const ind = new NonambiguousTrendFilter({ base: B.Csi, r: 1, s: 1, u: 1 });
      const value = (close: number): number => (ind.updateBar(makeBar(10, 12, 8, close))[0] as Scalar).value;

      expect(value(11)).toBe(0); // first finite value
      expect(value(12)).toBe(50); // positive and rising
      expect(value(11)).toBe(0); // positive and falling
      expect(value(9)).toBe(-25); // negative and falling
      expect(value(9)).toBe(0); // flat
      expect(value(9.5)).toBe(0); // negative and rising
    });
  });

  describe('warm-up', () => {
    it('propagates the NaN warm-up of the base, then outputs 0 at the first finite bar', () => {
      const q = 13;
      const ind = new NonambiguousTrendFilter({ base: B.Smi, q, r: 25, s: 2, u: 1 });

      for (let i = 0; i < q - 1; i++) {
        expect(isNaN((ind.updateBar(testBar(i))[0] as Scalar).value)).toBe(true);
        expect(ind.isPrimed()).toBe(false);
      }

      expect((ind.updateBar(testBar(q - 1))[0] as Scalar).value).toBe(0);
      expect(ind.isPrimed()).toBe(true);
    });
  });

  describe('entity updates', () => {
    it('routes scalars, bars, quotes and trades to the TSI base', () => {
      const expected = expectedTSI_R32_S13_U3;
      const inds = [0, 1, 2, 3].map(() => new NonambiguousTrendFilter());

      for (let i = 0; i < testInput.length; i++) {
        const s = new Scalar();
        s.time = testTime;
        s.value = testInput[i];

        const q = new Quote();
        q.time = testTime;
        q.bidPrice = testInput[i];
        q.askPrice = testInput[i];

        const t = new Trade();
        t.time = testTime;
        t.price = testInput[i];

        [
          inds[0].updateScalar(s),
          inds[1].updateBar(testBar(i)),
          inds[2].updateQuote(q),
          inds[3].updateTrade(t),
        ].forEach((out) => {
          expect(out.length).toBe(1);
          expect((out[0] as Scalar).time).toBe(testTime);
          expectValue((out[0] as Scalar).value, expected[i]);
        });
      }
    });
  });

  describe('mnemonic', () => {
    it('formats the default mnemonic', () => {
      const ind = new NonambiguousTrendFilter();
      expect(ind.metadata().mnemonic).toBe('ntf.tsi(2,32,13,3)');
      expect(ind.metadata().description).toBe('Nonambiguous Trend Filter ntf.tsi(2,32,13,3)');
    });

    [
      { base: B.Tsi, mnemonic: 'ntf.tsi(2,32,13,3)' },
      { base: B.Smi, mnemonic: 'ntf.smi(32,64,7,1)' },
      { base: B.Dti, mnemonic: 'ntf.dti(2,28,28,5)' },
      { base: B.Tvi, mnemonic: 'ntf.tvi(32,32,5)' },
      { base: B.Mdi, mnemonic: 'ntf.mdi(20,5,3)' },
      { base: B.Cmi, mnemonic: 'ntf.cmi(20,5,3)' },
      { base: B.Csi, mnemonic: 'ntf.csi(32,32,1)' },
    ].forEach((c) => {
      it(`formats the ${c.mnemonic} base mnemonic`, () => {
        expect(new NonambiguousTrendFilter({ base: c.base }).metadata().mnemonic).toBe(c.mnemonic);
      });
    });

    it('formats a custom mnemonic', () => {
      const ind = new NonambiguousTrendFilter({ base: B.Smi, q: 5, r: 20, s: 5, u: 3 });
      expect(ind.metadata().mnemonic).toBe('ntf.smi(5,20,5,3)');
    });

    it('formats the mnemonic with only the bar component set', () => {
      const ind = new NonambiguousTrendFilter({ barComponent: BarComponent.Median });
      expect(ind.metadata().mnemonic).toBe('ntf.tsi(2,32,13,3, hl/2)');
    });

    it('formats the mnemonic with only the quote component set', () => {
      const ind = new NonambiguousTrendFilter({ quoteComponent: QuoteComponent.Bid });
      expect(ind.metadata().mnemonic).toBe('ntf.tsi(2,32,13,3, b)');
    });

    it('formats the mnemonic with only the trade component set', () => {
      const ind = new NonambiguousTrendFilter({ tradeComponent: TradeComponent.Volume });
      expect(ind.metadata().mnemonic).toBe('ntf.tsi(2,32,13,3, v)');
    });

    it('formats the mnemonic with the bar and quote components set', () => {
      const ind = new NonambiguousTrendFilter({ barComponent: BarComponent.Open, quoteComponent: QuoteComponent.Bid });
      expect(ind.metadata().mnemonic).toBe('ntf.tsi(2,32,13,3, o, b)');
    });

    it('formats the mnemonic with the bar and trade components set', () => {
      const ind = new NonambiguousTrendFilter({ barComponent: BarComponent.High, tradeComponent: TradeComponent.Volume });
      expect(ind.metadata().mnemonic).toBe('ntf.tsi(2,32,13,3, h, v)');
    });

    it('formats the mnemonic with the quote and trade components set', () => {
      const ind = new NonambiguousTrendFilter({ quoteComponent: QuoteComponent.Ask, tradeComponent: TradeComponent.Volume });
      expect(ind.metadata().mnemonic).toBe('ntf.tsi(2,32,13,3, a, v)');
    });

    it('formats the MDI base mnemonic with a component', () => {
      const ind = new NonambiguousTrendFilter({ base: B.Mdi, barComponent: BarComponent.Median });
      expect(ind.metadata().mnemonic).toBe('ntf.mdi(20,5,3, hl/2)');
    });

    it('ignores components for the bar bases', () => {
      const ind = new NonambiguousTrendFilter({ base: B.Csi, barComponent: BarComponent.Median });
      expect(ind.metadata().mnemonic).toBe('ntf.csi(32,32,1)');
    });
  });

  describe('metadata', () => {
    it('reports the identifier and a single output', () => {
      const meta = new NonambiguousTrendFilter().metadata();
      expect(meta.identifier).toBe(IndicatorIdentifier.NonambiguousTrendFilter);
      expect(meta.mnemonic).toBe('ntf.tsi(2,32,13,3)');
      expect(meta.outputs.length).toBe(1);
      expect(meta.outputs[0].mnemonic).toBe('ntf.tsi(2,32,13,3)');
      expect(meta.outputs[0].description).toBe('Nonambiguous Trend Filter ntf.tsi(2,32,13,3)');
    });
  });

  describe('invalid parameters', () => {
    it('throws when q < 0', () => {
      expect(() => new NonambiguousTrendFilter({ q: -1 })).toThrowError();
    });
    it('throws when r < 0', () => {
      expect(() => new NonambiguousTrendFilter({ r: -1 })).toThrowError();
    });
    it('throws when s < 0', () => {
      expect(() => new NonambiguousTrendFilter({ s: -1 })).toThrowError();
    });
    it('throws when u < 0', () => {
      expect(() => new NonambiguousTrendFilter({ u: -1 })).toThrowError();
    });
    it('throws on an unknown base', () => {
      expect(() => new NonambiguousTrendFilter({ base: 99 as NonambiguousTrendFilterBase })).toThrowError();
    });
  });
});
