import { AdxTypeFilter } from './adx-type-filter';
import { AdxTypeFilterSource } from './params';
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
  expectedTSIMTM_Q2_R32_S32,
  expectedTSIMTM_Q2_R20_S5,
  expectedTSIMTM_Q2_R13_S1,
  expectedTSIMTM_Q2_R1_S1,
  expectedTSIMTM_Q5_R32_S32,
  expectedSMIRAW_Q32_R32_S32,
  expectedSMIRAW_Q32_R20_S5,
  expectedSMIRAW_Q5_R32_S32,
  expectedDTINUM_Q2_R32_S32,
  expectedDTINUM_Q2_R28_S28,
  expectedDTINUM_Q5_R32_S32,
  expectedTVI_R32_S32,
  expectedTVI_R12_S12,
  expectedTVI_R1_S1,
  expectedTSINORM_R32_S32,
  expectedTSINORM_R20_S20,
} from './testdata';

interface Combo {
  name: string;
  source: AdxTypeFilterSource;
  q: number;
  r: number;
  s: number;
  expected: number[];
}

const closeCombos: Combo[] = [
  { name: 'TSIMTM_Q2_R32_S32', source: AdxTypeFilterSource.TsiMomentum, q: 2, r: 32, s: 32, expected: expectedTSIMTM_Q2_R32_S32 },
  { name: 'TSIMTM_Q2_R20_S5', source: AdxTypeFilterSource.TsiMomentum, q: 2, r: 20, s: 5, expected: expectedTSIMTM_Q2_R20_S5 },
  { name: 'TSIMTM_Q2_R13_S1', source: AdxTypeFilterSource.TsiMomentum, q: 2, r: 13, s: 1, expected: expectedTSIMTM_Q2_R13_S1 },
  { name: 'TSIMTM_Q2_R1_S1', source: AdxTypeFilterSource.TsiMomentum, q: 2, r: 1, s: 1, expected: expectedTSIMTM_Q2_R1_S1 },
  { name: 'TSIMTM_Q5_R32_S32', source: AdxTypeFilterSource.TsiMomentum, q: 5, r: 32, s: 32, expected: expectedTSIMTM_Q5_R32_S32 },
  { name: 'TSINORM_R32_S32', source: AdxTypeFilterSource.TsiNormalized, q: 2, r: 32, s: 32, expected: expectedTSINORM_R32_S32 },
  { name: 'TSINORM_R20_S20', source: AdxTypeFilterSource.TsiNormalized, q: 2, r: 20, s: 20, expected: expectedTSINORM_R20_S20 },
];

const barCombos: Combo[] = [
  { name: 'SMIRAW_Q32_R32_S32', source: AdxTypeFilterSource.SmiMomentum, q: 32, r: 32, s: 32, expected: expectedSMIRAW_Q32_R32_S32 },
  { name: 'SMIRAW_Q32_R20_S5', source: AdxTypeFilterSource.SmiMomentum, q: 32, r: 20, s: 5, expected: expectedSMIRAW_Q32_R20_S5 },
  { name: 'SMIRAW_Q5_R32_S32', source: AdxTypeFilterSource.SmiMomentum, q: 5, r: 32, s: 32, expected: expectedSMIRAW_Q5_R32_S32 },
  { name: 'DTINUM_Q2_R32_S32', source: AdxTypeFilterSource.DtiMomentum, q: 2, r: 32, s: 32, expected: expectedDTINUM_Q2_R32_S32 },
  { name: 'DTINUM_Q2_R28_S28', source: AdxTypeFilterSource.DtiMomentum, q: 2, r: 28, s: 28, expected: expectedDTINUM_Q2_R28_S28 },
  { name: 'DTINUM_Q5_R32_S32', source: AdxTypeFilterSource.DtiMomentum, q: 5, r: 32, s: 32, expected: expectedDTINUM_Q5_R32_S32 },
  { name: 'TVI_R32_S32', source: AdxTypeFilterSource.TviBalance, q: 0, r: 32, s: 32, expected: expectedTVI_R32_S32 },
  { name: 'TVI_R12_S12', source: AdxTypeFilterSource.TviBalance, q: 0, r: 12, s: 12, expected: expectedTVI_R12_S12 },
  { name: 'TVI_R1_S1', source: AdxTypeFilterSource.TviBalance, q: 0, r: 1, s: 1, expected: expectedTVI_R1_S1 },
];

const testTime = new Date(0);

const testBar = (i: number): Bar => {
  const b = new Bar();
  b.time = testTime;
  b.open = testInput[i];
  b.high = testHigh[i];
  b.low = testLow[i];
  b.close = testInput[i];
  return b;
};

const expectValue = (actual: number, expected: number): void => {
  if (isNaN(expected)) {
    expect(isNaN(actual)).toBe(true);
  } else {
    expect(actual).toBeCloseTo(expected, 13);
  }
};

describe('AdxTypeFilter', () => {
  describe('reference data', () => {
    closeCombos.forEach((combo) => {
      it(`matches the reference for ${combo.name}`, () => {
        const ind = new AdxTypeFilter({ source: combo.source, q: combo.q, r: combo.r, s: combo.s });

        for (let i = 0; i < testInput.length; i++) {
          expectValue(ind.update(testInput[i]), combo.expected[i]);
        }
      });
    });

    barCombos.forEach((combo) => {
      it(`matches the reference for ${combo.name} from bars`, () => {
        const ind = new AdxTypeFilter({ source: combo.source, q: combo.q, r: combo.r, s: combo.s });

        for (let i = 0; i < testInput.length; i++) {
          const out = ind.updateBar(testBar(i));
          expect(out.length).toBe(1);
          expectValue((out[0] as Scalar).value, combo.expected[i]);
        }
      });

      it(`matches the reference for ${combo.name} from high, low and close`, () => {
        const ind = new AdxTypeFilter({ source: combo.source, q: combo.q, r: combo.r, s: combo.s });

        for (let i = 0; i < testInput.length; i++) {
          expectValue(ind.updateHighLowClose(testHigh[i], testLow[i], testInput[i]), combo.expected[i]);
        }
      });
    });
  });

  describe('passthrough', () => {
    it('returns |C - C[1]| for the TSI momentum with r = s = 1', () => {
      const ind = new AdxTypeFilter({ source: AdxTypeFilterSource.TsiMomentum, q: 2, r: 1, s: 1 });
      expect(isNaN(ind.update(10))).toBe(true);
      expect(ind.update(12)).toBe(2);
      expect(ind.update(11)).toBe(1);
    });

    it('returns |C - 0.5*(H + L)| for the SMI momentum with q = r = s = 1', () => {
      const ind = new AdxTypeFilter({ source: AdxTypeFilterSource.SmiMomentum, q: 1, r: 1, s: 1 });
      expect(ind.updateHighLowClose(11, 9, 10.5)).toBe(0.5);
    });

    it('applies the tick rule for the TVI balance with r = s = 1', () => {
      const ind = new AdxTypeFilter({ source: AdxTypeFilterSource.TviBalance, r: 1, s: 1 });
      expect(ind.update(10)).toBe(0);
      expect(ind.update(12)).toBe(2);
      expect(ind.update(9)).toBe(3);
    });
  });

  describe('warm-up', () => {
    it('returns NaN for bars 0..30 with the default SMI look-back', () => {
      const ind = new AdxTypeFilter({ source: AdxTypeFilterSource.SmiMomentum });

      for (let i = 0; i < testInput.length; i++) {
        const value = ind.updateHighLowClose(testHigh[i], testLow[i], testInput[i]);
        expect(isNaN(value)).toBe(i < 31);
      }
    });

    it('has no warm-up for the TVI balance', () => {
      const ind = new AdxTypeFilter({ source: AdxTypeFilterSource.TviBalance });

      for (let i = 0; i < testInput.length; i++) {
        expect(isNaN(ind.updateHighLowClose(testHigh[i], testLow[i], testInput[i]))).toBe(false);
      }
    });
  });

  describe('non-negative', () => {
    [
      AdxTypeFilterSource.TsiMomentum,
      AdxTypeFilterSource.SmiMomentum,
      AdxTypeFilterSource.DtiMomentum,
      AdxTypeFilterSource.TviBalance,
      AdxTypeFilterSource.TsiNormalized,
    ].forEach((source) => {
      it(`keeps every finite value non-negative for source ${source}`, () => {
        const ind = new AdxTypeFilter({ source });

        for (let i = 0; i < testInput.length; i++) {
          const value = (ind.updateBar(testBar(i))[0] as Scalar).value;
          if (!isNaN(value)) {
            expect(value).toBeGreaterThanOrEqual(0);
          }
        }
      });
    });
  });

  describe('isPrimed', () => {
    it('primes from the first finite momentum', () => {
      const q = 5;
      const ind = new AdxTypeFilter({ source: AdxTypeFilterSource.TsiMomentum, q });

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
    const expected = expectedTSIMTM_Q2_R32_S32;

    it('updates the TSI momentum from scalars, bars, quotes and trades', () => {
      const scalarInd = new AdxTypeFilter();
      const barInd = new AdxTypeFilter();
      const quoteInd = new AdxTypeFilter();
      const tradeInd = new AdxTypeFilter();

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
          scalarInd.updateScalar(s),
          barInd.updateBar(testBar(i)),
          quoteInd.updateQuote(q),
          tradeInd.updateTrade(t),
        ].forEach((out) => {
          expect(out.length).toBe(1);
          expect((out[0] as Scalar).time).toBe(testTime);
          expectValue((out[0] as Scalar).value, expected[i]);
        });
      }
    });

    it('uses a single value as the high, the low and the close for the SMI momentum', () => {
      const params = { source: AdxTypeFilterSource.SmiMomentum, q: 5 };
      const reference = new AdxTypeFilter(params);
      const scalarInd = new AdxTypeFilter(params);
      const quoteInd = new AdxTypeFilter(params);
      const tradeInd = new AdxTypeFilter(params);

      for (let i = 0; i < testInput.length; i++) {
        const v = testInput[i];
        const exp = reference.updateHighLowClose(v, v, v);

        const s = new Scalar();
        s.time = testTime;
        s.value = v;

        const q = new Quote();
        q.time = testTime;
        q.bidPrice = v;
        q.askPrice = v;

        const t = new Trade();
        t.time = testTime;
        t.price = v;

        expectValue((scalarInd.updateScalar(s)[0] as Scalar).value, exp);
        expectValue((quoteInd.updateQuote(q)[0] as Scalar).value, exp);
        expectValue((tradeInd.updateTrade(t)[0] as Scalar).value, exp);
      }
    });
  });

  describe('mnemonic', () => {
    it('formats the default mnemonic', () => {
      const ind = new AdxTypeFilter();
      expect(ind.metadata().mnemonic).toBe('atf.tsi(2,32,32)');
      expect(ind.metadata().description).toBe('ADX-Type Filter atf.tsi(2,32,32)');
    });

    [
      { source: AdxTypeFilterSource.TsiMomentum, mnemonic: 'atf.tsi(2,32,32)' },
      { source: AdxTypeFilterSource.SmiMomentum, mnemonic: 'atf.smi(32,32,32)' },
      { source: AdxTypeFilterSource.DtiMomentum, mnemonic: 'atf.dti(2,32,32)' },
      { source: AdxTypeFilterSource.TviBalance, mnemonic: 'atf.tvi(32,32)' },
      { source: AdxTypeFilterSource.TsiNormalized, mnemonic: 'atf.tsin(2,32,32)' },
    ].forEach((c) => {
      it(`formats the ${c.mnemonic} source mnemonic`, () => {
        expect(new AdxTypeFilter({ source: c.source }).metadata().mnemonic).toBe(c.mnemonic);
      });
    });

    it('formats a custom mnemonic', () => {
      const ind = new AdxTypeFilter({ source: AdxTypeFilterSource.SmiMomentum, q: 5, r: 20, s: 5 });
      expect(ind.metadata().mnemonic).toBe('atf.smi(5,20,5)');
    });

    it('formats the mnemonic with only the bar component set', () => {
      const ind = new AdxTypeFilter({ barComponent: BarComponent.Median });
      expect(ind.metadata().mnemonic).toBe('atf.tsi(2,32,32, hl/2)');
    });

    it('formats the mnemonic with only the quote component set', () => {
      const ind = new AdxTypeFilter({ quoteComponent: QuoteComponent.Bid });
      expect(ind.metadata().mnemonic).toBe('atf.tsi(2,32,32, b)');
    });

    it('formats the mnemonic with only the trade component set', () => {
      const ind = new AdxTypeFilter({ tradeComponent: TradeComponent.Volume });
      expect(ind.metadata().mnemonic).toBe('atf.tsi(2,32,32, v)');
    });

    it('formats the mnemonic with the bar and quote components set', () => {
      const ind = new AdxTypeFilter({ barComponent: BarComponent.Open, quoteComponent: QuoteComponent.Bid });
      expect(ind.metadata().mnemonic).toBe('atf.tsi(2,32,32, o, b)');
    });

    it('formats the mnemonic with the bar and trade components set', () => {
      const ind = new AdxTypeFilter({ barComponent: BarComponent.High, tradeComponent: TradeComponent.Volume });
      expect(ind.metadata().mnemonic).toBe('atf.tsi(2,32,32, h, v)');
    });

    it('formats the mnemonic with the quote and trade components set', () => {
      const ind = new AdxTypeFilter({ quoteComponent: QuoteComponent.Ask, tradeComponent: TradeComponent.Volume });
      expect(ind.metadata().mnemonic).toBe('atf.tsi(2,32,32, a, v)');
    });

    it('formats the normalized source mnemonic with a component', () => {
      const ind = new AdxTypeFilter({ source: AdxTypeFilterSource.TsiNormalized, barComponent: BarComponent.Median });
      expect(ind.metadata().mnemonic).toBe('atf.tsin(2,32,32, hl/2)');
    });

    it('ignores components for the bar sources', () => {
      const ind = new AdxTypeFilter({ source: AdxTypeFilterSource.SmiMomentum, barComponent: BarComponent.Median });
      expect(ind.metadata().mnemonic).toBe('atf.smi(32,32,32)');
    });
  });

  describe('metadata', () => {
    it('reports the identifier and a single output', () => {
      const meta = new AdxTypeFilter().metadata();
      expect(meta.identifier).toBe(IndicatorIdentifier.AdxTypeFilter);
      expect(meta.mnemonic).toBe('atf.tsi(2,32,32)');
      expect(meta.outputs.length).toBe(1);
      expect(meta.outputs[0].mnemonic).toBe('atf.tsi(2,32,32)');
      expect(meta.outputs[0].description).toBe('ADX-Type Filter atf.tsi(2,32,32)');
    });
  });

  describe('invalid parameters', () => {
    it('throws when q < 0', () => {
      expect(() => new AdxTypeFilter({ q: -1 })).toThrowError();
    });
    it('throws when r < 1', () => {
      expect(() => new AdxTypeFilter({ r: 0 })).toThrowError();
    });
    it('throws when s < 1', () => {
      expect(() => new AdxTypeFilter({ s: 0 })).toThrowError();
    });
    it('throws on an unknown source', () => {
      expect(() => new AdxTypeFilter({ source: 99 as AdxTypeFilterSource })).toThrowError();
    });
  });
});
