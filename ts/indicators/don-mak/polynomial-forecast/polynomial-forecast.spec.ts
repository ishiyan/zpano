import { PolynomialForecast } from './polynomial-forecast';
import { IndicatorIdentifier } from '../../core/indicator-identifier';
import { Scalar } from '../../../entities/scalar';
import * as td from './testdata';

const TOLERANCE = 9;

// [degree, order, smoothing]
const combos: Array<[number, number, number]> = [
  [2, 1, 0], [2, 1, 3], [2, 1, 6], [2, 2, 0], [2, 2, 3], [2, 2, 6],
  [3, 1, 0], [3, 1, 3], [3, 1, 6], [3, 2, 0], [3, 2, 3], [3, 2, 6],
  [4, 1, 0], [4, 1, 3], [4, 1, 6], [4, 2, 0], [4, 2, 3], [4, 2, 6],
  [5, 1, 0], [5, 1, 3], [5, 1, 6], [5, 2, 0], [5, 2, 3], [5, 2, 6],
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

describe('PolynomialForecast', () => {
  describe('reference data', () => {
    combos.forEach(([degree, order, smoothing]) => {
      const name = `EXPECTED_D${degree}_O${order}_S${smoothing}`;
      it(`matches the reference for ${name}`, () => {
        const expected = testData[name];
        const ind = new PolynomialForecast({ degree, order, smoothing });
        const values = td.INPUT_CLOSE.map((c) => ind.update(c));
        expectSeries(values, expected, name);
      });
    });

    it('matches the linear-input series (TEST1 order 1)', () => {
      const ind = new PolynomialForecast({ degree: 3, order: 1, smoothing: 0 });
      const values = td.TEST1_INPUT_LINEAR.map((c) => ind.update(c));
      expectSeries(values, td.TEST1_EXPECTED_D3_O1_S0, 'TEST1_O1');
    });

    it('matches the linear-input series (TEST1 order 2)', () => {
      const ind = new PolynomialForecast({ degree: 3, order: 2, smoothing: 0 });
      const values = td.TEST1_INPUT_LINEAR.map((c) => ind.update(c));
      expectSeries(values, td.TEST1_EXPECTED_D3_O2_S0, 'TEST1_O2');
    });
  });

  describe('mnemonic', () => {
    it('formats the default mnemonic', () => {
      expect(new PolynomialForecast({}).metadata().mnemonic).toBe('pof(3,1,0)');
    });

    it('formats a custom mnemonic', () => {
      expect(new PolynomialForecast({ degree: 5, order: 2, smoothing: 6 }).metadata().mnemonic).toBe('pof(5,2,6)');
    });
  });

  describe('metadata', () => {
    it('reports the identifier and one output', () => {
      const meta = new PolynomialForecast({}).metadata();
      expect(meta.identifier).toBe(IndicatorIdentifier.PolynomialForecast);
      expect(meta.outputs.length).toBe(1);
    });
  });

  describe('updateScalar', () => {
    it('returns a single output', () => {
      const ind = new PolynomialForecast({});
      let out: ReturnType<typeof ind.updateScalar> = [];
      for (const c of td.INPUT_CLOSE) {
        const s = new Scalar();
        s.time = new Date(0);
        s.value = c;
        out = ind.updateScalar(s);
      }
      expect(out.length).toBe(1);
      const last = td.INPUT_CLOSE.length - 1;
      expect((out[0] as Scalar).value).toBeCloseTo(td.EXPECTED_D3_O1_S0[last], TOLERANCE);
    });
  });

  describe('invalid parameters', () => {
    it('throws when degree < 2', () => {
      expect(() => new PolynomialForecast({ degree: 1 })).toThrowError();
    });
    it('throws when order < 1', () => {
      expect(() => new PolynomialForecast({ order: 0 })).toThrowError();
    });
    it('throws when order > 2', () => {
      expect(() => new PolynomialForecast({ order: 3 })).toThrowError();
    });
    it('throws when smoothing < 0', () => {
      expect(() => new PolynomialForecast({ smoothing: -1 })).toThrowError();
    });
  });
});
