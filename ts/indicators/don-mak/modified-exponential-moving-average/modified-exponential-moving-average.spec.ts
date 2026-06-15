import { ModifiedExponentialMovingAverage } from './modified-exponential-moving-average';
import { IndicatorIdentifier } from '../../core/indicator-identifier';
import { Scalar } from '../../../entities/scalar';
import * as td from './testdata';

const TOLERANCE = 9;

// [period, degree, skip]
const combos: Array<[number, number, number]> = [
  [3, 3, 1], [3, 3, 2], [3, 3, 4],
  [3, 4, 1], [3, 4, 2], [3, 4, 4],
  [6, 3, 1], [6, 3, 2], [6, 3, 4],
  [6, 4, 1], [6, 4, 2], [6, 4, 4],
  [12, 3, 1], [12, 3, 2], [12, 3, 4],
  [12, 4, 1], [12, 4, 2], [12, 4, 4],
];

const testData = td as unknown as Record<string, number[]>;

function expectSeries(actual: number[], expected: number[], label: string): void {
  expect(actual.length).toBe(expected.length);
  for (let i = 0; i < expected.length; i++) {
    if (isNaN(expected[i])) {
      expect(actual[i]).withContext(`${label}[${i}]`).toBeNaN();
    } else {
      expect(actual[i]).withContext(`${label}[${i}]`).toBeCloseTo(expected[i], TOLERANCE);
    }
  }
}

describe('ModifiedExponentialMovingAverage', () => {
  describe('reference data', () => {
    combos.forEach(([period, degree, skip]) => {
      const name = `EXPECTED_P${period}_D${degree}_SK${skip}`;
      it(`matches the reference for ${name}`, () => {
        const expected = testData[name];
        const ind = new ModifiedExponentialMovingAverage({ period, degree, skip });
        const values = td.INPUT_CLOSE.map((c) => ind.update(c));
        expectSeries(values, expected, name);
      });
    });

    it('matches the linear-input series (TEST1)', () => {
      const ind = new ModifiedExponentialMovingAverage({ period: 6, degree: 3, skip: 1 });
      const values = td.TEST1_INPUT_LINEAR.map((c) => ind.update(c));
      expectSeries(values, td.TEST1_EXPECTED_P6_D3_SK1, 'TEST1');
    });
  });

  describe('mnemonic', () => {
    it('formats the default mnemonic', () => {
      expect(new ModifiedExponentialMovingAverage({}).metadata().mnemonic).toBe('mema(6,3,1)');
    });

    it('formats a custom mnemonic', () => {
      expect(new ModifiedExponentialMovingAverage({ period: 12, degree: 4, skip: 2 }).metadata().mnemonic).toBe('mema(12,4,2)');
    });
  });

  describe('metadata', () => {
    it('reports the identifier and one output', () => {
      const meta = new ModifiedExponentialMovingAverage({}).metadata();
      expect(meta.identifier).toBe(IndicatorIdentifier.ModifiedExponentialMovingAverage);
      expect(meta.outputs.length).toBe(1);
    });
  });

  describe('updateScalar', () => {
    it('returns a single output', () => {
      const ind = new ModifiedExponentialMovingAverage({});
      let out: ReturnType<typeof ind.updateScalar> = [];
      for (const c of td.INPUT_CLOSE) {
        const s = new Scalar();
        s.time = new Date(0);
        s.value = c;
        out = ind.updateScalar(s);
      }
      expect(out.length).toBe(1);
      const last = td.INPUT_CLOSE.length - 1;
      expect((out[0] as Scalar).value).toBeCloseTo(td.EXPECTED_P6_D3_SK1[last], TOLERANCE);
    });
  });

  describe('invalid parameters', () => {
    it('throws when period < 2', () => {
      expect(() => new ModifiedExponentialMovingAverage({ period: 1 })).toThrowError();
    });
    it('throws when degree < 2', () => {
      expect(() => new ModifiedExponentialMovingAverage({ degree: 1 })).toThrowError();
    });
    it('throws when skip < 1', () => {
      expect(() => new ModifiedExponentialMovingAverage({ skip: 0 })).toThrowError();
    });
  });
});
