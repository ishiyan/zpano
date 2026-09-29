import { TickVolumeIndicator } from './tick-volume-indicator';
import { TickVolumeIndicatorOutput } from './output';
import { defaultParams } from './params';
import { IndicatorIdentifier } from '../../core/indicator-identifier';
import { Shape } from '../../core/outputs/shape/shape';
import { Bar } from '../../../entities/bar';
import { Quote } from '../../../entities/quote';
import { Scalar } from '../../../entities/scalar';
import { Trade } from '../../../entities/trade';
import {
  testInput, testOpen, testHigh, testLow,
  testUpticks, testDownticks,
  expectedR12_S12_U1,
  expectedR25_S13_U1,
  expectedR32_S32_U5,
  expectedR1_S1_U1,
  expectedR32_S5_U1,
  expectedR12_S12_U5,
  expectedR20_S5_U3,
  expectedR5_S5_U5,
  expectedR32_S32_U1,
  expectedR10_S10_U1,
  expectedR50_S25_U1,
  expectedR12_S26_U9,
  expectedR3_S3_U3,
  expectedR7_S4_U2,
  expectedR64_S1_U1,
  expectedR12_S12_U3,
} from './testdata';

const DIGITS = 10;

interface Combo {
  name: string;
  r: number;
  s: number;
  u: number;
  expected: number[];
}

const combos: Combo[] = [
  { name: 'R12_S12_U1', r: 12, s: 12, u: 1, expected: expectedR12_S12_U1 },
  { name: 'R25_S13_U1', r: 25, s: 13, u: 1, expected: expectedR25_S13_U1 },
  { name: 'R32_S32_U5', r: 32, s: 32, u: 5, expected: expectedR32_S32_U5 },
  { name: 'R1_S1_U1', r: 1, s: 1, u: 1, expected: expectedR1_S1_U1 },
  { name: 'R32_S5_U1', r: 32, s: 5, u: 1, expected: expectedR32_S5_U1 },
  { name: 'R12_S12_U5', r: 12, s: 12, u: 5, expected: expectedR12_S12_U5 },
  { name: 'R20_S5_U3', r: 20, s: 5, u: 3, expected: expectedR20_S5_U3 },
  { name: 'R5_S5_U5', r: 5, s: 5, u: 5, expected: expectedR5_S5_U5 },
  { name: 'R32_S32_U1', r: 32, s: 32, u: 1, expected: expectedR32_S32_U1 },
  { name: 'R10_S10_U1', r: 10, s: 10, u: 1, expected: expectedR10_S10_U1 },
  { name: 'R50_S25_U1', r: 50, s: 25, u: 1, expected: expectedR50_S25_U1 },
  { name: 'R12_S26_U9', r: 12, s: 26, u: 9, expected: expectedR12_S26_U9 },
  { name: 'R3_S3_U3', r: 3, s: 3, u: 3, expected: expectedR3_S3_U3 },
  { name: 'R7_S4_U2', r: 7, s: 4, u: 2, expected: expectedR7_S4_U2 },
  { name: 'R64_S1_U1', r: 64, s: 1, u: 1, expected: expectedR64_S1_U1 },
  { name: 'R12_S12_U3', r: 12, s: 12, u: 3, expected: expectedR12_S12_U3 },
];

// Tick rule with passthrough stages: first -> 0, up -> +100, down -> -100, flat -> 0.
const tickValues = [10, 12, 11, 11];
const tickExpected = [0, 100, -100, 0];

function checkOutput(out: ReturnType<TickVolumeIndicator['updateScalar']>, time: Date, expected: number): void {
  expect(out.length).toBe(1);
  expect((out[0] as Scalar).time).toBe(time);
  expect((out[0] as Scalar).value).toBeCloseTo(expected, DIGITS);
}

describe('TickVolumeIndicator', () => {
  describe('reference data', () => {
    combos.forEach((combo) => {
      it(`matches the reference for ${combo.name}`, () => {
        const ind = new TickVolumeIndicator({ r: combo.r, s: combo.s, u: combo.u });

        for (let i = 0; i < testUpticks.length; i++) {
          const act = ind.update(testUpticks[i], testDownticks[i]);

          if (isNaN(combo.expected[i])) {
            expect(act).toBeNaN();
          } else {
            expect(act).toBeCloseTo(combo.expected[i], DIGITS);
          }
        }
      });
    });
  });

  describe('passthrough', () => {
    it('reduces to 100*(up-down)/(up+down) when every stage is a passthrough', () => {
      const ind = new TickVolumeIndicator({ r: 1, s: 1, u: 1 });

      expect(ind.update(8, 2)).toBeCloseTo(60, DIGITS);
      expect(ind.update(0, 5)).toBeCloseTo(-100, DIGITS);
      expect(ind.update(0, 0)).toBe(0); // flat market -> division guard
    });
  });

  describe('isPrimed', () => {
    it('is primed after the first update (no NaN warm-up)', () => {
      const ind = new TickVolumeIndicator();
      expect(ind.isPrimed()).toBe(false);
      ind.update(testUpticks[0], testDownticks[0]);
      expect(ind.isPrimed()).toBe(true);
    });

    it('is primed after the first tick-rule update', () => {
      const ind = new TickVolumeIndicator();
      expect(ind.isPrimed()).toBe(false);
      const s = new Scalar();
      s.time = new Date(0);
      s.value = 10;
      ind.updateScalar(s);
      expect(ind.isPrimed()).toBe(true);
    });
  });

  describe('mnemonic', () => {
    it('formats the default mnemonic', () => {
      const ind = new TickVolumeIndicator();
      expect(ind.metadata().mnemonic).toBe('tvi(12,12,1)');
      expect(ind.metadata().description).toBe('Tick Volume Indicator tvi(12,12,1)');
    });

    it('formats a custom mnemonic', () => {
      const ind = new TickVolumeIndicator({ r: 32, s: 32, u: 5 });
      expect(ind.metadata().mnemonic).toBe('tvi(32,32,5)');
    });

    it('uses the default params', () => {
      const ind = new TickVolumeIndicator(defaultParams());
      expect(ind.metadata().mnemonic).toBe('tvi(12,12,1)');
    });
  });

  describe('metadata', () => {
    it('reports the identifier and one output', () => {
      const ind = new TickVolumeIndicator();
      const meta = ind.metadata();
      expect(meta.identifier).toBe(IndicatorIdentifier.TickVolumeIndicator);
      expect(meta.mnemonic).toBe('tvi(12,12,1)');
      expect(meta.description).toBe('Tick Volume Indicator tvi(12,12,1)');
      expect(meta.outputs.length).toBe(1);
      expect(meta.outputs[0].kind).toBe(TickVolumeIndicatorOutput.TickVolumeIndicatorValue);
      expect(meta.outputs[0].shape).toBe(Shape.Scalar);
      expect(meta.outputs[0].mnemonic).toBe('tvi(12,12,1)');
      expect(meta.outputs[0].description).toBe('Tick Volume Indicator tvi(12,12,1)');
    });
  });

  describe('updateBar', () => {
    it('maps up = close - low and down = high - close', () => {
      const ind = new TickVolumeIndicator({ r: 12, s: 12, u: 1 });
      const time = new Date(0);
      for (let i = 0; i < testInput.length; i++) {
        const bar = new Bar();
        bar.time = time;
        bar.open = testOpen[i];
        bar.high = testHigh[i];
        bar.low = testLow[i];
        bar.close = testInput[i];
        bar.volume = 0;
        checkOutput(ind.updateBar(bar), time, expectedR12_S12_U1[i]);
      }
    });
  });

  describe('updateScalar', () => {
    it('applies the tick rule to the value', () => {
      const ind = new TickVolumeIndicator({ r: 1, s: 1, u: 1 });
      const time = new Date(0);
      tickValues.forEach((value, i) => {
        const s = new Scalar();
        s.time = time;
        s.value = value;
        checkOutput(ind.updateScalar(s), time, tickExpected[i]);
      });
    });

    it('matches update of the value changes', () => {
      const ind = new TickVolumeIndicator({ r: 12, s: 12, u: 3 });
      const ref = new TickVolumeIndicator({ r: 12, s: 12, u: 3 });
      const time = new Date(0);
      for (let i = 0; i < testInput.length; i++) {
        const diff = i === 0 ? 0 : testInput[i] - testInput[i - 1];
        const expected = ref.update(Math.max(diff, 0), Math.max(-diff, 0));
        const s = new Scalar();
        s.time = time;
        s.value = testInput[i];
        checkOutput(ind.updateScalar(s), time, expected);
      }
    });
  });

  describe('updateTrade', () => {
    it('applies the tick rule to the price', () => {
      const ind = new TickVolumeIndicator({ r: 1, s: 1, u: 1 });
      const time = new Date(0);
      tickValues.forEach((price, i) => {
        const trade = new Trade();
        trade.time = time;
        trade.price = price;
        trade.volume = 1;
        checkOutput(ind.updateTrade(trade), time, tickExpected[i]);
      });
    });
  });

  describe('updateQuote', () => {
    it('applies the tick rule to the mid price', () => {
      const ind = new TickVolumeIndicator({ r: 1, s: 1, u: 1 });
      const time = new Date(0);
      const quotes: [number, number][] = [[9, 11], [11, 13], [10, 12], [10.5, 11.5]];
      quotes.forEach(([bid, ask], i) => {
        const quote = new Quote();
        quote.time = time;
        quote.bidPrice = bid;
        quote.askPrice = ask;
        quote.bidSize = 1;
        quote.askSize = 1;
        checkOutput(ind.updateQuote(quote), time, tickExpected[i]);
      });
    });
  });

  describe('invalid parameters', () => {
    it('throws when r < 1', () => {
      expect(() => new TickVolumeIndicator({ r: 0 }))
        .toThrowError('invalid tick volume indicator parameters: r should be greater than 0');
    });
    it('throws when s < 1', () => {
      expect(() => new TickVolumeIndicator({ s: 0 }))
        .toThrowError('invalid tick volume indicator parameters: s should be greater than 0');
    });
    it('throws when u < 1', () => {
      expect(() => new TickVolumeIndicator({ u: 0 }))
        .toThrowError('invalid tick volume indicator parameters: u should be greater than 0');
    });
  });
});
